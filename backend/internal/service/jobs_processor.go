package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ajaxe/email-ingestion/internal/worker"
)

type StoragePrefixDeleter interface {
	DeletePrefix(ctx context.Context, prefix string) error
}

type JobsProcessor struct {
	storage StoragePrefixDeleter
}

func NewJobsProcessor(storage StoragePrefixDeleter) *JobsProcessor {
	return &JobsProcessor{
		storage: storage,
	}
}

func (j *JobsProcessor) Process(ctx context.Context, payload *worker.JobPayload) error {
	if payload == nil {
		return fmt.Errorf("nil job payload")
	}

	switch payload.Type {
	case worker.JobTypePurgeEmailStorage:
		return j.processPurgeEmailStorage(ctx, payload)
	default:
		slog.WarnContext(ctx, "unsupported job payload type", "type", payload.Type, "application_id", payload.ApplicationID)
		return fmt.Errorf("unsupported job payload type: %s", payload.Type)
	}
}

func (j *JobsProcessor) processPurgeEmailStorage(ctx context.Context, payload *worker.JobPayload) error {
	slog.InfoContext(ctx, "begin PURGE_EMAIL_STORAGE job",
		"application_id", payload.ApplicationID,
		"prefixes_count", len(payload.S3KeyPrefixes),
		"retry_count", payload.RetryCount,
	)

	var lastErr error
	successCount := 0

	for _, prefix := range payload.S3KeyPrefixes {
		if prefix == "" {
			continue
		}
		slog.InfoContext(ctx, "purging s3 key prefix", "application_id", payload.ApplicationID, "prefix", prefix)
		err := j.storage.DeletePrefix(ctx, prefix)
		if err != nil {
			slog.ErrorContext(ctx, "failed to purge s3 key prefix", "application_id", payload.ApplicationID, "prefix", prefix, "error", err)
			lastErr = err
		} else {
			successCount++
		}
	}

	if lastErr != nil {
		return fmt.Errorf("failed purging s3 prefixes (%d/%d succeeded): %w", successCount, len(payload.S3KeyPrefixes), lastErr)
	}

	slog.InfoContext(ctx, "successfully completed PURGE_EMAIL_STORAGE job",
		"application_id", payload.ApplicationID,
		"purged_count", successCount,
	)

	return nil
}
