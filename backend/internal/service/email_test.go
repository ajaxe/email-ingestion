package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/google/uuid"
)

type mockStorageClient struct {
	contents map[string][]byte
}

func (m *mockStorageClient) DownloadObject(ctx context.Context, key string) (io.ReadCloser, error) {
	data, ok := m.contents[key]
	if !ok {
		return nil, io.EOF
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *mockStorageClient) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	return "https://s3.localstack.site/test-bucket/" + key + "?signature=test", nil
}

func (m *mockStorageClient) PresignTTL() time.Duration {
	return 15 * time.Minute
}

func TestGetAttachmentURL(t *testing.T) {
	appID := uuid.New()
	s3Prefix := "apps/" + appID.String() + "/emails/msg-123"

	mockStorage := &mockStorageClient{
		contents: make(map[string][]byte),
	}

	contentBody := storage.EmailStorageContent{
		Text: "Hello",
		HTML: "<p>Hello</p>",
		Attachments: []storage.EmailStorageAttachment{
			{
				AttachmentID: s3Prefix + "/attachments/report.pdf",
				FileName:     "report.pdf",
				ContentType:  "application/pdf",
				Size:         1024,
				IsInline:     false,
			},
		},
	}

	contentJSON, _ := json.Marshal(contentBody)
	mockStorage.contents[s3Prefix+"/contents.json"] = contentJSON

	ctx := context.Background()
	ttl := 15 * time.Minute
	presigned, err := mockStorage.GeneratePresignedURL(ctx, s3Prefix+"/attachments/report.pdf", ttl)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(presigned, "report.pdf") {
		t.Errorf("expected presigned URL to contain filename, got %s", presigned)
	}
}
