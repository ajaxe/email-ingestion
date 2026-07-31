package worker

import (
	"context"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
)

type EmailProcessor struct {
	queries        *public.Queries
	storageService *storage.S3StorageService
}

func (e *EmailProcessor) Process(ctx context.Context, payload *model.IngestEmailPayload) error {
	// Implement the logic to process the email here.

	return nil
}
