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
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type EmailStorageClient interface {
	DownloadObject(ctx context.Context, key string) (io.ReadCloser, error)
	GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	PresignTTL() time.Duration
}

func NewEmailService(queries *public.Queries, storage EmailStorageClient) *EmailService {
	return &EmailService{
		queries: queries,
		storage: storage,
	}
}

type EmailService struct {
	queries *public.Queries
	storage EmailStorageClient
}

// GetEmailByID fetches an email by its ID and application ID, retrieves the associated content from storage, and returns a combined EmailContent struct.
func (e *EmailService) GetEmailByID(ctx context.Context, appID uuid.UUID, emailID uuid.UUID) (*EmailContent, error) {
	data, err := e.queries.GetIngestedEmailByID(ctx, public.GetIngestedEmailByIDParams{
		ID:            emailID,
		ApplicationID: appID,
	})

	if err != nil {
		return nil, err
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
		EmailStorageContent:     contents,
	}
	return r, nil
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
