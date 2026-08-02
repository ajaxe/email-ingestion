package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/service"
)

type EmailIngestionHandler struct {
	processor *service.EmailProcessor
}

func NewEmailIngestionHandler(processor *service.EmailProcessor) *EmailIngestionHandler {
	return &EmailIngestionHandler{
		processor: processor,
	}
}

func (h *EmailIngestionHandler) Handle(ctx context.Context, msg *redis.StreamData) error {
	j, ok := msg.Data.(string)
	if !ok {
		slog.ErrorContext(ctx, "failed to parse message data as string")
		// Not retryable, so return nil to ack
		return nil
	}

	p, err := model.ParseIngestEmail(j)
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse ingest email payload", "error", err)
		// Not retryable
		return nil
	}

	err = h.processor.Process(ctx, p)
	if err != nil {
		slog.ErrorContext(ctx, "failed to process email message", "spoolID", p.SpoolID, "error", err)
		var retryErr *model.RetryableError
		if errors.As(err, &retryErr) {
			// Return error so consumer does not ack it (it will be auto-claimed and retried)
			return err
		}
		// Not a retryable error, return nil to ack and move on
		return nil
	}

	return nil
}
