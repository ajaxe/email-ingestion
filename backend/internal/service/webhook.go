package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/internal/webhook"
	"github.com/ajaxe/email-ingestion/internal/worker"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/crypto"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type WebhookService struct {
	queries           *public.Queries
	cfg               *config.WebhookConfig
	publisher         EventPublisher
	webhookStreamName string
}

func NewWebhookService(queries *public.Queries, cfg *config.WebhookConfig, publisher EventPublisher, webhookStreamName string) *WebhookService {
	return &WebhookService{
		queries:           queries,
		cfg:               cfg,
		publisher:         publisher,
		webhookStreamName: webhookStreamName,
	}
}

func (s *WebhookService) RegisterWebhook(ctx context.Context, appID uuid.UUID, webhookURL string) (string, error) {
	app, err := s.queries.GetApplicationByID(ctx, appID)
	if err != nil {
		return "", apperror.NotFound("application not found", err)
	}

	// Perform SSRF Challenge Handshake
	client := webhook.NewSSRFProtectedClient(s.cfg, app.IsTrusted)
	if err := webhook.PerformChallengeHandshake(ctx, client, webhookURL); err != nil {
		return "", apperror.UnprocessableEntity("webhook verification failed", err.Error(), err)
	}

	// Generate a new webhook secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", apperror.Internal("failed to generate webhook secret", err)
	}
	webhookSecret := hex.EncodeToString(secretBytes)

	// Encrypt the webhook secret before saving it to the database
	encryptedSecret, err := crypto.Encrypt(webhookSecret, s.cfg.EncryptionKey)
	if err != nil {
		return "", apperror.Internal("failed to encrypt webhook secret", err)
	}

	// Save the webhook config to the application
	err = s.queries.UpdateApplicationWebhook(ctx, public.UpdateApplicationWebhookParams{
		ID:            appID,
		WebhookUrl:    webhookURL,
		WebhookSecret: encryptedSecret,
	})
	if err != nil {
		return "", apperror.Internal("failed to save webhook configuration", err)
	}
	return webhookSecret, nil
}

func (s *WebhookService) ListJobs(ctx context.Context, appID uuid.UUID, limit, offset int32, status string) ([]public.ListWebhookJobsByApplicationRow, error) {
	jobs, err := s.queries.ListWebhookJobsByApplication(ctx, public.ListWebhookJobsByApplicationParams{
		ApplicationID: appID,
		Limit:         limit,
		Offset:        offset,
		Status:        pgtype.Text{String: status, Valid: status != ""},
	})

	if err != nil {
		return nil, apperror.Internal("failed to list webhook jobs", err)
	}
	return jobs, nil
}

func (s *WebhookService) RedeliverJob(ctx context.Context, appID uuid.UUID, jobID uuid.UUID) (*public.WebhookDeliveryJob, error) {
	// Validate job ownership & existence
	oldJob, err := s.queries.GetWebhookJobByIDAndAppID(ctx, public.GetWebhookJobByIDAndAppIDParams{
		ID:            jobID,
		ApplicationID: appID,
	})
	if err != nil {
		return nil, apperror.NotFound("webhook job not found", err)
	}

	// Block redelivery if the job is currently processing
	if oldJob.Status == public.WebhookStatusPROCESSING {
		return nil, apperror.Conflict("webhook job is currently processing")
	}

	// Cancel previous job instance if PENDING or FAILED so cron poller ignores it
	_ = s.queries.CancelWebhookJob(ctx, public.CancelWebhookJobParams{
		ID:            jobID,
		ApplicationID: appID,
	})

	// Create a NEW job instance for redelivery (Approach B: Immutable Job History)
	newJob, err := s.queries.EnqueueWebhookJob(ctx, public.EnqueueWebhookJobParams{
		ApplicationID:   appID,
		IngestedEmailID: oldJob.IngestedEmailID,
	})
	if err != nil {
		return nil, apperror.Internal("failed to enqueue new webhook job for redelivery", err)
	}

	// Notify Redis outbox for immediate reprocessing
	if s.publisher != nil && s.webhookStreamName != "" {
		// Mark as PROCESSING immediately so background cron poller doesn't double-pick it
		_ = s.queries.UpdateWebhookJobStatus(ctx, public.UpdateWebhookJobStatusParams{
			ID:             newJob.ID,
			Status:         public.WebhookStatusPROCESSING,
			RetryCount:     0,
			NextDeliveryAt: newJob.NextDeliveryAt,
		})

		payload, err := util.JSON(&worker.WebhookDeliveryPayload{
			ApplicationID:   newJob.ApplicationID.String(),
			IngestedEmailID: newJob.IngestedEmailID.String(),
			JobID:           newJob.ID.String(),
		})
		if err != nil {
			return nil, apperror.Internal("failed to serialize webhook payload", fmt.Errorf("%w", err))
		}

		if err := s.publisher.Publish(ctx, s.webhookStreamName, payload); err != nil {
			// Revert to PENDING if stream publishing failed
			_ = s.queries.UpdateWebhookJobStatus(ctx, public.UpdateWebhookJobStatusParams{
				ID:             newJob.ID,
				Status:         public.WebhookStatusPENDING,
				RetryCount:     0,
				NextDeliveryAt: newJob.NextDeliveryAt,
			})
			return nil, apperror.Internal("failed to publish webhook delivery job for redelivery", err)
		}
	}

	return &newJob, nil
}
