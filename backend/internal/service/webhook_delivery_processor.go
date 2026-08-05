package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/webhook"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type WebhookStorage interface {
	DownloadObject(ctx context.Context, key string) (io.ReadCloser, error)
}

type WebhookDeliveryProcessor struct {
	queries *public.Queries
	storage WebhookStorage
	cfg     *config.WebhookConfig
}

func NewWebhookDeliveryProcessor(queries *public.Queries, storage WebhookStorage, cfg *config.WebhookConfig) *WebhookDeliveryProcessor {
	return &WebhookDeliveryProcessor{
		queries: queries,
		storage: storage,
		cfg:     cfg,
	}
}

func (w *WebhookDeliveryProcessor) Process(ctx context.Context, payload *model.WebhookDeliveryPayload) error {
	appID, err := uuid.Parse(payload.ApplicationID)
	if err != nil {
		return fmt.Errorf("invalid application id: %w", err)
	}
	emailID, err := uuid.Parse(payload.IngestedEmailID)
	if err != nil {
		return fmt.Errorf("invalid ingested email id: %w", err)
	}

	// 1. Get Application Details
	app, err := w.queries.GetApplicationByID(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get application %s: %w", appID, err)
	}

	if app.WebhookUrl == "" {
		slog.WarnContext(ctx, "application has no webhook URL configured, skipping delivery", "application_id", appID)
		return nil
	}

	// 2. Get Job Details (To check retry count)
	job, err := w.queries.GetWebhookJobByIDs(ctx, public.GetWebhookJobByIDsParams{
		ApplicationID:   appID,
		IngestedEmailID: emailID,
	})
	if err != nil {
		return fmt.Errorf("failed to get webhook job: %w", err)
	}

	// 3. Get Ingested Email Metadata
	email, err := w.queries.GetIngestedEmailByID(ctx, emailID)
	if err != nil {
		return fmt.Errorf("failed to get ingested email: %w", err)
	}

	// 4. Formulate the payload containing parsed JSON content metadata.
	contentKey := fmt.Sprintf("%s/contents.json", email.S3KeyPrefix)
	stream, err := w.storage.DownloadObject(ctx, contentKey)
	if err != nil {
		return model.NewRetryableError(fmt.Errorf("failed to download parsed content: %w", err))
	}
	defer stream.Close()

	bodyBytes, err := io.ReadAll(stream)
	if err != nil {
		return model.NewRetryableError(fmt.Errorf("failed to read content stream: %w", err))
	}

	// Wrap inside a webhook event payload
	eventPayload := map[string]interface{}{
		"event":             "email.received",
		"ingested_email_id": email.ID.String(),
		"reference_token":   email.ReferenceToken,
		"from_address":      email.FromAddress,
		"subject":           email.Subject,
		"message_id":        email.MessageID,
		"content":           json.RawMessage(bodyBytes), // Embed the contents.json directly
	}

	reqBody, _ := json.Marshal(eventPayload)

	// 5. Build Request & HMAC-SHA256 signature generator
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, app.WebhookUrl, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to build webhook request: %w", err)
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(app.WebhookSecret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(reqBody)
	signature := hex.EncodeToString(mac.Sum(nil))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gateway-Signature", fmt.Sprintf("t=%s,v1=%s", timestamp, signature))

	// 6. Execute Request
	client := webhook.NewSSRFProtectedClient(w.cfg, app.IsTrusted)
	start := time.Now()
	resp, err := client.Do(req)
	durationMs := int32(time.Since(start).Milliseconds())

	status := public.WebhookStatusSUCCESS
	statusCode := int32(0)
	respBody := ""
	var deliverErr error

	if err != nil {
		deliverErr = fmt.Errorf("http request failed: %w", err)
		status = public.WebhookStatusFAILED
		respBody = err.Error()
	} else {
		defer resp.Body.Close()
		statusCode = int32(resp.StatusCode)
		if statusCode >= 200 && statusCode < 300 {
			// Success!
		} else {
			status = public.WebhookStatusFAILED
			deliverErr = fmt.Errorf("non-success http status code: %d", statusCode)
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			respBody = string(b)
		}
	}

	// 7. Handle retries and write audit attempts
	isRetry := job.RetryCount > 0
	_ = w.queries.LogWebhookAttempt(ctx, public.LogWebhookAttemptParams{
		WebhookDeliveryJobID: job.ID,
		AttemptNumber:        job.RetryCount + 1,
		HttpStatusCode:       statusCode,
		ResponseBody:         respBody,
		IsRetry:              isRetry,
		DurationMs:           durationMs,
	})

	if deliverErr != nil {
		// in case of webhook error update status of the job as DEAD if retry exhausted
		// otherwise, update as PENDING and set the next try time based on exponential backoff with full jitter.
		if job.RetryCount < app.MaxRetries {
			status = public.WebhookStatusPENDING
			// Exponential Backoff with Full Jitter
			// base = 2s, cap = 60s
			base := float64(2)
			backoff := math.Pow(2, float64(job.RetryCount)) * base
			if backoff > 60 {
				backoff = 60
			}
			jitter := rand.Float64() * backoff
			nextDelivery := time.Now().Add(time.Duration(jitter) * time.Second)

			_ = w.queries.UpdateWebhookJobStatus(ctx, public.UpdateWebhookJobStatusParams{
				ID:             job.ID,
				Status:         status,
				RetryCount:     job.RetryCount + 1,
				NextDeliveryAt: nextDelivery,
			})
			// Return nil so the stream consumer ACKs the message.
			// The cron job will re-enqueue it when NextDeliveryAt is reached.
			return nil
		} else {
			status = public.WebhookStatusDEAD
			_ = w.queries.UpdateWebhookJobStatus(ctx, public.UpdateWebhookJobStatusParams{
				ID:             job.ID,
				Status:         status,
				RetryCount:     job.RetryCount + 1,
				NextDeliveryAt: time.Now(),
			})
			return nil // Acknowledged as dead
		}
	}

	_ = w.queries.UpdateWebhookJobStatus(ctx, public.UpdateWebhookJobStatusParams{
		ID:             job.ID,
		Status:         status,
		RetryCount:     job.RetryCount,
		NextDeliveryAt: time.Now(),
	})

	return nil
}
