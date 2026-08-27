package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/worker"
)

type mockStorageDeleter struct {
	deletedPrefixes []string
	shouldFail      bool
}

func (m *mockStorageDeleter) DeletePrefix(ctx context.Context, prefix string) error {
	if m.shouldFail {
		return fmt.Errorf("mock storage error")
	}
	m.deletedPrefixes = append(m.deletedPrefixes, prefix)
	return nil
}

func TestJobsProcessor_PurgeEmailStorage(t *testing.T) {
	mockStore := &mockStorageDeleter{}
	processor := service.NewJobsProcessor(mockStore)

	payload := &worker.JobPayload{
		Type:          worker.JobTypePurgeEmailStorage,
		ApplicationID: "test-app-id",
		S3KeyPrefixes: []string{"apps/app1/emails/msg1", "apps/app1/emails/msg2"},
	}

	err := processor.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mockStore.deletedPrefixes) != 2 {
		t.Errorf("expected 2 deleted prefixes, got %d", len(mockStore.deletedPrefixes))
	}
}

func TestJobsProcessor_PurgeEmailStorage_Error(t *testing.T) {
	mockStore := &mockStorageDeleter{shouldFail: true}
	processor := service.NewJobsProcessor(mockStore)

	payload := &worker.JobPayload{
		Type:          worker.JobTypePurgeEmailStorage,
		ApplicationID: "test-app-id",
		S3KeyPrefixes: []string{"apps/app1/emails/msg1"},
	}

	err := processor.Process(context.Background(), payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
