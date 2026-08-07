package service

import (
	"context"
	"encoding/json"
	"io"

	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type EmailStorageDownload interface {
	DownloadObject(ctx context.Context, key string) (io.ReadCloser, error)
}

func NewEmailService(queries *public.Queries, storage EmailStorageDownload) *EmailService {
	return &EmailService{
		queries: queries,
		storage: storage,
	}
}

type EmailService struct {
	queries *public.Queries
	storage EmailStorageDownload
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
		IngestedEmail:       data,
		EmailStorageContent: contents,
	}
	return r, nil
}
