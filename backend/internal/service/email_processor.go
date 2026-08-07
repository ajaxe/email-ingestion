package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/internal/worker"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
)

// ==========================================
// 1. Consumer-Defined Interfaces
// ==========================================

// EmailStorage abstracts S3 interactions (Satisfied by storage.S3StorageService)
type EmailStorage interface {
	DownloadObject(ctx context.Context, key string) (io.ReadCloser, error)
	UploadObject(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	DeleteObject(ctx context.Context, key string) error
}

// EventPublisher abstracts Redis streams (Satisfied by redis.StreamService)
type EventPublisher interface {
	Publish(ctx context.Context, stream string, payload interface{}) error
}

// EmailRepository defines the exact queries required by the EmailProcessor.
// To satisfy the Transactional Outbox pattern, the `CreateIngestedEmailAndJobTx`
// method should be implemented in your DB layer to execute both inserts in a single transaction.
// (You can also satisfy this using sqlc's generated Querier interface if you adjust the Tx method).
type EmailRepository interface {
	GetAssignedEmailByLocalPart(ctx context.Context, localPart string) (public.AssignedEmail, error)
	UpdateSpooledEmailStatus(ctx context.Context, arg public.UpdateSpooledEmailStatusParams) error

	// Transactional Outbox Method
	CreateIngestedEmailAndJobTx(ctx context.Context, emailParams public.CreateIngestedEmailParams) (public.IngestedEmail, error)
}

// ==========================================
// 2. Service Definition
// ==========================================

type EmailProcessor struct {
	repo              EmailRepository
	storage           EmailStorage
	publisher         EventPublisher
	webhookStreamName string
}

func NewEmailProcessor(repo EmailRepository, storage EmailStorage, publisher EventPublisher, webhookStreamName string) *EmailProcessor {
	return &EmailProcessor{
		repo:              repo,
		storage:           storage,
		publisher:         publisher,
		webhookStreamName: webhookStreamName,
	}
}

// ==========================================
// 3. Orchestration Method (Clean SRP)
// ==========================================

func (e *EmailProcessor) Process(ctx context.Context, payload *worker.IngestEmailPayload) error {
	spoolID, err := uuid.Parse(payload.SpoolID)
	if err != nil {
		return fmt.Errorf("invalid spool ID: %w", err)
	}

	slog.InfoContext(ctx, "begin processing email", "spool_id", spoolID, "upload_key", payload.UploadKey)

	// Centralized error handling & state transition
	var processErr error
	defer func() {
		e.finalizeSpoolStatus(ctx, spoolID, processErr)
	}()

	// 1. Download & Parse
	env, processErr := e.fetchAndParseEmail(ctx, payload.UploadKey)
	if processErr != nil {
		return processErr
	}

	// 2. Resolve Recipient
	assignedEmail, refToken, processErr := e.resolveRecipient(ctx, env)
	if processErr != nil {
		return processErr
	}

	slog.InfoContext(ctx, "found assigned email address", "spool_id", spoolID, "application_id", assignedEmail.ApplicationID)

	msgID, found := e.emailMessageID(env)
	if !found {
		slog.Info("falling back to spoolID as Message-ID", "spool_id", spoolID)
		msgID = spoolID.String()
	}

	// 3. Extract and Upload Assets to S3
	basePath := util.ProcessedEmailS3KeyPrefix(assignedEmail.ApplicationID.String(), msgID)
	processErr = e.uploadParsedAssets(ctx, env, basePath)
	if processErr != nil {
		return apperror.NewRetryableError(fmt.Errorf("failed to upload parsed assets: %w", processErr))
	}

	// 4. Transactional Outbox (Atomic DB Insert)
	ingestedEmail, processErr := e.repo.CreateIngestedEmailAndJobTx(ctx, public.CreateIngestedEmailParams{
		ApplicationID:   assignedEmail.ApplicationID,
		AssignedEmailID: assignedEmail.ID,
		ReferenceToken:  refToken,
		FromAddress:     env.GetHeader("From"),
		Subject:         env.GetHeader("Subject"),
		MessageID:       msgID,
		S3KeyPrefix:     basePath,
	})
	if processErr != nil {
		return apperror.NewRetryableError(fmt.Errorf("db transaction failed: %w", processErr))
	}

	j, err := util.JSON(&worker.WebhookDeliveryPayload{
		ApplicationID:   assignedEmail.ApplicationID.String(),
		IngestedEmailID: ingestedEmail.ID.String(),
	})
	if err != nil {
		return apperror.NewRetryableError(fmt.Errorf("failed to serialize webhook payload: %w", err))
	}
	// 5. Notify downstream workers via Message Broker
	processErr = e.publisher.Publish(ctx, e.webhookStreamName, j)
	if processErr != nil {
		return apperror.NewRetryableError(fmt.Errorf("failed to publish webhook delivery job: %w", processErr))
	}

	// 6. Cleanup raw email from S3 (Spool DB status is updated in defer block)
	_ = e.storage.DeleteObject(ctx, payload.UploadKey)
	slog.InfoContext(ctx, "email processed successfully", "spool_id", spoolID, "ingested_email_id", ingestedEmail.ID)

	return nil
}

// ==========================================
// 4. Helper Methods
// ==========================================

func (e *EmailProcessor) finalizeSpoolStatus(ctx context.Context, spoolID uuid.UUID, processErr error) {
	status := public.SpoolStatusSUCCESS
	errMsg := ""

	if processErr != nil {
		status = public.SpoolStatusFAILED
		// If it's a retryable error, mark as PENDING to retry, else it's a terminal FAILED
		if _, ok := processErr.(*apperror.RetryableError); ok {
			status = public.SpoolStatusPENDING
		}
		errMsg = processErr.Error()
	}

	_ = e.repo.UpdateSpooledEmailStatus(ctx, public.UpdateSpooledEmailStatusParams{
		ID:               spoolID,
		Status:           status,
		LastErrorMessage: errMsg,
	})
}

func (e *EmailProcessor) fetchAndParseEmail(ctx context.Context, uploadKey string) (*enmime.Envelope, error) {
	stream, err := e.storage.DownloadObject(ctx, uploadKey)
	if err != nil {
		return nil, apperror.NewRetryableError(fmt.Errorf("failed to download spool object: %w", err))
	}
	defer stream.Close()

	env, err := enmime.ReadEnvelope(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to parse raw email: %w", err)
	}
	return env, nil
}

func (e *EmailProcessor) resolveRecipient(ctx context.Context, env *enmime.Envelope) (public.AssignedEmail, string, error) {
	toAddress := env.GetHeader("Delivered-To")
	if toAddress == "" {
		toAddress = env.GetHeader("To")
	}

	addr, err := mail.ParseAddress(toAddress)
	if err != nil {
		return public.AssignedEmail{}, "", fmt.Errorf("invalid to address: %w", err)
	}

	localPart := strings.Split(addr.Address, "@")[0]
	refToken := ""
	if idx := strings.Index(localPart, "+"); idx != -1 {
		refToken = localPart[idx+1:]
		localPart = localPart[:idx]
	}

	assignedEmail, err := e.repo.GetAssignedEmailByLocalPart(ctx, localPart)
	if err != nil {
		return public.AssignedEmail{}, "", fmt.Errorf("failed to get assigned email: %w", err)
	}

	if assignedEmail.ApplicationID == uuid.Nil {
		return public.AssignedEmail{}, "", fmt.Errorf("unassigned email address")
	}

	return assignedEmail, refToken, nil
}

func (e *EmailProcessor) uploadParsedAssets(ctx context.Context, env *enmime.Envelope, basePath string) error {
	headerMap := make(map[string]string)
	for _, key := range env.GetHeaderKeys() {
		headerMap[key] = env.GetHeader(key)
	}

	var err error
	var attachments []storage.EmailStorageAttachment
	atchPrefix := util.AttachmentStorageKeyPrefix(basePath)

	for _, att := range env.Attachments {
		attKey := fmt.Sprintf("%s/%s", atchPrefix, att.FileName)
		_, err = e.storage.UploadObject(ctx, attKey, bytes.NewReader(att.Content), att.ContentType)
		if err != nil {
			return fmt.Errorf("failed to upload attachment %s: %w", att.FileName, err)
		}
		slog.DebugContext(ctx, "uploaded attachment", "attachment_key", attKey, "size", len(att.Content), "content_type", att.ContentType, "filename", att.FileName, "content-disposition", att.Disposition)
		attachments = append(attachments, storage.EmailStorageAttachment{
			AttachmentID: attKey,
			FileName:     att.FileName,
			ContentType:  att.ContentType,
			Size:         int64(len(att.Content)),
			IsInline:     false,
		})
	}

	for _, in := range env.Inlines {
		attKey := fmt.Sprintf("%s/%s", atchPrefix, in.FileName)
		_, err = e.storage.UploadObject(ctx, attKey, bytes.NewReader(in.Content), in.ContentType)
		if err != nil {
			return fmt.Errorf("failed to upload inline attachment %s: %w", in.FileName, err)
		}
		slog.DebugContext(ctx, "uploaded inline attachment", "attachment_key", attKey, "size", len(in.Content), "content_type", in.ContentType, "filename", in.FileName, "content-disposition", in.Disposition)
		attachments = append(attachments, storage.EmailStorageAttachment{
			AttachmentID: attKey,
			FileName:     in.FileName,
			ContentType:  in.ContentType,
			Size:         int64(len(in.Content)),
			IsInline:     true,
		})
	}

	contentBody := storage.EmailStorageContent{
		Text:        env.Text,
		HTML:        env.HTML,
		Headers:     headerMap,
		Attachments: attachments,
	}

	contentJSON, _ := json.Marshal(contentBody)

	contentKey := util.ContentJSONStorageKeyPrefix(basePath)
	_, err = e.storage.UploadObject(ctx, contentKey, bytes.NewReader(contentJSON), "application/json")
	if err != nil {
		return fmt.Errorf("failed to upload contents.json: %w", err)
	}

	return nil
}

// emailMessageID extracts Message-ID email header from the envelop and transforms into a url safe hash.
func (e *EmailProcessor) emailMessageID(env *enmime.Envelope) (v string, found bool) {
	msgID := strings.TrimSpace(env.GetHeader("Message-Id"))
	msgID = strings.Trim(msgID, "<>")
	if msgID == "" {
		return "", false
	} else {
		hash := sha256.Sum256([]byte(msgID))
		msgID = hex.EncodeToString(hash[:])
	}
	return msgID, true
}
