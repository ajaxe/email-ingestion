package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/worker"
	"github.com/ajaxe/email-ingestion/internal/worker/handler"
)

type mockFailingStorageDeleter struct{}

func (m *mockFailingStorageDeleter) DeletePrefix(ctx context.Context, prefix string) error {
	return fmt.Errorf("simulated S3 storage failure")
}

type mockEventPublisher struct {
	publishedStream string
	publishedData   string
}

func (m *mockEventPublisher) Publish(ctx context.Context, stream string, payload interface{}) error {
	m.publishedStream = stream
	if str, ok := payload.(string); ok {
		m.publishedData = str
	}
	return nil
}

func TestJobsHandler_BackoffAndRetry(t *testing.T) {
	mockStore := &mockFailingStorageDeleter{}
	processor := service.NewJobsProcessor(mockStore)
	publisher := &mockEventPublisher{}

	h := handler.NewJobsHandler(processor, publisher, "jobs_stream")

	payload := worker.JobPayload{
		Type:          worker.JobTypePurgeEmailStorage,
		ApplicationID: "app-123",
		S3KeyPrefixes: []string{"apps/app-123/emails/msg1"},
		RetryCount:    0,
	}

	payloadBytes, _ := json.Marshal(payload)
	msg := &redis.StreamData{
		Data: string(payloadBytes),
	}

	start := time.Now()
	err := h.Handle(context.Background(), msg)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected nil error (job handled with retry), got %v", err)
	}

	// Should take at least 1s for the backoff delay
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected backoff delay of ~1s, elapsed was %v", elapsed)
	}

	if publisher.publishedStream != "jobs_stream" {
		t.Errorf("expected re-publish to jobs_stream, got %s", publisher.publishedStream)
	}

	var retriedPayload worker.JobPayload
	if err := json.Unmarshal([]byte(publisher.publishedData), &retriedPayload); err != nil {
		t.Fatalf("failed to unmarshal retried payload: %v", err)
	}

	if retriedPayload.RetryCount != 1 {
		t.Errorf("expected incremented RetryCount to be 1, got %d", retriedPayload.RetryCount)
	}
}

func TestJobsHandler_MaxRetriesReached(t *testing.T) {
	mockStore := &mockFailingStorageDeleter{}
	processor := service.NewJobsProcessor(mockStore)
	publisher := &mockEventPublisher{}

	h := handler.NewJobsHandler(processor, publisher, "jobs_stream")

	payload := worker.JobPayload{
		Type:          worker.JobTypePurgeEmailStorage,
		ApplicationID: "app-123",
		S3KeyPrefixes: []string{"apps/app-123/emails/msg1"},
		RetryCount:    2, // Next failure makes it 3 (max retries reached)
	}

	payloadBytes, _ := json.Marshal(payload)
	msg := &redis.StreamData{
		Data: string(payloadBytes),
	}

	err := h.Handle(context.Background(), msg)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Max retries reached, should not re-publish
	if publisher.publishedData != "" {
		t.Errorf("expected no re-publish after reaching max retries, got %s", publisher.publishedData)
	}
}
