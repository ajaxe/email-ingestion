package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/worker"
)

type JobsHandler struct {
	processor  *service.JobsProcessor
	publisher  service.EventPublisher
	streamName string
}

func NewJobsHandler(processor *service.JobsProcessor, publisher service.EventPublisher, streamName string) *JobsHandler {
	return &JobsHandler{
		processor:  processor,
		publisher:  publisher,
		streamName: streamName,
	}
}

func getBackoffDuration(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 1 * time.Second
	case 2:
		return 3 * time.Second
	default:
		return 5 * time.Second
	}
}

func (h *JobsHandler) Handle(ctx context.Context, msg *redis.StreamData) error {
	j, ok := msg.Data.(string)
	if !ok {
		slog.ErrorContext(ctx, "failed to parse jobs stream message data as string")
		return nil
	}

	p := &worker.JobPayload{}
	err := json.Unmarshal([]byte(j), p)
	if err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal job payload", "error", err, "raw_data", j)
		return nil
	}

	err = h.processor.Process(ctx, p)
	if err != nil {
		p.RetryCount++
		if p.RetryCount < 3 && h.publisher != nil && h.streamName != "" {
			backoff := getBackoffDuration(p.RetryCount)
			slog.WarnContext(ctx, "job processing failed, re-enqueueing with backoff delay",
				"job_type", p.Type,
				"application_id", p.ApplicationID,
				"attempt", p.RetryCount,
				"backoff_ms", backoff.Milliseconds(),
				"max_retries", 3,
				"error", err,
			)
			payloadStr, marshalErr := json.Marshal(p)
			if marshalErr == nil {
				time.Sleep(backoff)
				_ = h.publisher.Publish(ctx, h.streamName, string(payloadStr))
				return nil
			}
		}

		slog.ErrorContext(ctx, "job processing failed permanently after 3 retries",
			"job_type", p.Type,
			"application_id", p.ApplicationID,
			"retry_count", p.RetryCount,
			"error", err,
		)
		// Return nil to ACK and drop message from stream after reaching max retries
		return nil
	}

	return nil
}
