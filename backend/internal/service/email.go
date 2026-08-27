package service

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/internal/worker"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type EmailStorageClient interface {
	DownloadObject(ctx context.Context, key string) (io.ReadCloser, error)
	GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	PresignTTL() time.Duration
}

func NewEmailService(queries *public.Queries, storage EmailStorageClient, publisher EventPublisher, jobsStreamName string) *EmailService {
	return &EmailService{
		queries:        queries,
		storage:        storage,
		publisher:      publisher,
		jobsStreamName: jobsStreamName,
	}
}

type EmailService struct {
	queries        *public.Queries
	storage        EmailStorageClient
	publisher      EventPublisher
	jobsStreamName string
}

// GetEmailByID fetches an email by its ID and application ID.
// If the email is soft-deleted, it returns the metadata with IsDeleted: true without downloading content from object storage.
func (e *EmailService) GetEmailByID(ctx context.Context, appID uuid.UUID, emailID uuid.UUID) (*EmailContent, error) {
	data, err := e.queries.GetIngestedEmailByID(ctx, public.GetIngestedEmailByIDParams{
		ID:            emailID,
		ApplicationID: appID,
	})

	if err != nil {
		return nil, err
	}

	if data.DeletedAt != nil {
		return &EmailContent{
			GetIngestedEmailByIDRow: data,
			IsDeleted:               true,
		}, nil
	}

	contentsKey := util.ContentJSONStorageKeyPrefix(data.S3KeyPrefix)
	c, err := e.storage.DownloadObject(ctx, contentsKey)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	contents := storage.EmailStorageContent{}
	err = json.NewDecoder(c).Decode(&contents)
	if err != nil {
		return nil, err
	}
	r := &EmailContent{
		GetIngestedEmailByIDRow: data,
		IsDeleted:               false,
		EmailStorageContent:     contents,
	}
	return r, nil
}

// SoftDeleteEmail marks an email as soft deleted in Postgres and enqueues an async storage purge job.
func (e *EmailService) SoftDeleteEmail(ctx context.Context, appID uuid.UUID, emailID uuid.UUID) error {
	deletedEmail, err := e.queries.SoftDeleteIngestedEmail(ctx, public.SoftDeleteIngestedEmailParams{
		ID:            emailID,
		ApplicationID: appID,
	})
	if err != nil {
		return apperror.NotFound("Email not found or already deleted", err)
	}

	if e.publisher != nil && e.jobsStreamName != "" && deletedEmail.S3KeyPrefix != "" {
		payload := &worker.JobPayload{
			Type:          worker.JobTypePurgeEmailStorage,
			ApplicationID: appID.String(),
			S3KeyPrefixes: []string{deletedEmail.S3KeyPrefix},
		}
		payloadStr, err := util.JSON(payload)
		if err == nil {
			_ = e.publisher.Publish(ctx, e.jobsStreamName, payloadStr)
		}
	}

	return nil
}

// BulkSoftDeleteEmails marks multiple emails as soft deleted in Postgres and enqueues a bulk async storage purge job.
func (e *EmailService) BulkSoftDeleteEmails(ctx context.Context, appID uuid.UUID, emailIDs []uuid.UUID) (int64, error) {
	if len(emailIDs) == 0 {
		return 0, nil
	}

	deletedRows, err := e.queries.SoftDeleteIngestedEmailsBulk(ctx, public.SoftDeleteIngestedEmailsBulkParams{
		Column1:       emailIDs,
		ApplicationID: appID,
	})
	if err != nil {
		return 0, apperror.Internal("Failed to bulk delete emails", err)
	}

	var prefixes []string
	for _, row := range deletedRows {
		if row.S3KeyPrefix != "" {
			prefixes = append(prefixes, row.S3KeyPrefix)
		}
	}

	if len(prefixes) > 0 && e.publisher != nil && e.jobsStreamName != "" {
		payload := &worker.JobPayload{
			Type:          worker.JobTypePurgeEmailStorage,
			ApplicationID: appID.String(),
			S3KeyPrefixes: prefixes,
		}
		payloadStr, err := util.JSON(payload)
		if err == nil {
			_ = e.publisher.Publish(ctx, e.jobsStreamName, payloadStr)
		}
	}

	return int64(len(deletedRows)), nil
}

// GetEmailWebhookHistory fetches all webhook delivery jobs and logs for a specific ingested email.
func (e *EmailService) GetEmailWebhookHistory(ctx context.Context, appID uuid.UUID, emailID uuid.UUID) ([]public.GetWebhookJobsByEmailIDRow, error) {
	return e.queries.GetWebhookJobsByEmailID(ctx, public.GetWebhookJobsByEmailIDParams{
		ApplicationID:   appID,
		IngestedEmailID: emailID,
	})
}

// GetAttachmentURL verifies tenant email and attachment ownership and generates a short-lived S3 presigned URL.
func (e *EmailService) GetAttachmentURL(ctx context.Context, appID uuid.UUID, emailID uuid.UUID, attachmentID string) (*dto.AttachmentURLResponse, error) {
	data, err := e.queries.GetIngestedEmailByID(ctx, public.GetIngestedEmailByIDParams{
		ID:            emailID,
		ApplicationID: appID,
	})
	if err != nil {
		return nil, apperror.NotFound("Email not found", err)
	}

	if data.DeletedAt != nil {
		return nil, apperror.NotFound("Email content has been soft-deleted")
	}

	contentsKey := util.ContentJSONStorageKeyPrefix(data.S3KeyPrefix)
	c, err := e.storage.DownloadObject(ctx, contentsKey)
	if err != nil {
		return nil, apperror.NotFound("Email contents not found", err)
	}
	defer c.Close()

	var contents storage.EmailStorageContent
	if err := json.NewDecoder(c).Decode(&contents); err != nil {
		return nil, apperror.Internal("Failed to decode email contents", err)
	}

	var targetKey string
	for _, att := range contents.Attachments {
		if att.AttachmentID == attachmentID || att.FileName == attachmentID || strings.HasSuffix(att.AttachmentID, "/"+attachmentID) {
			targetKey = att.AttachmentID
			break
		}
	}

	if targetKey == "" {
		return nil, apperror.NotFound("Attachment not found")
	}

	if !strings.HasPrefix(targetKey, data.S3KeyPrefix) {
		return nil, apperror.Forbidden("Attachment access unauthorized")
	}

	ttl := e.storage.PresignTTL()
	presignedURL, err := e.storage.GeneratePresignedURL(ctx, targetKey, ttl)
	if err != nil {
		return nil, apperror.Internal("Failed to generate presigned URL", err)
	}

	return &dto.AttachmentURLResponse{
		AttachmentID: attachmentID,
		DownloadURL:  presignedURL,
		ExpiresAt:    time.Now().Add(ttl),
	}, nil
}
