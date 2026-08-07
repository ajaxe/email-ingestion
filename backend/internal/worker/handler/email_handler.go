package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/worker"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
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

	p := &worker.IngestEmailPayload{}
	err := json.Unmarshal([]byte(j), p)
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse ingest email payload", "error", err)
		// Not retryable
		return nil
	}

	err = h.processor.Process(ctx, p)
	if err != nil {
		slog.ErrorContext(ctx, "failed to process email message", "spoolID", p.SpoolID, "error", err)
		var retryErr *apperror.RetryableError
		if errors.As(err, &retryErr) {
			// Return error so consumer does not ack it (it will be auto-claimed and retried)
			return err
		}
		// Not a retryable error, return nil to ack and move on
		return nil
	}

	return nil
}
