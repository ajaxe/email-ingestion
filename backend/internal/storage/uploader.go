package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ajaxe/email-ingestion/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3StorageService struct {
	config *config.StorageConfig
	mu     sync.Mutex
	client *transfermanager.Client
}

func NewStorageService(config *config.StorageConfig) *S3StorageService {
	return &S3StorageService{
		config: config,
	}
}

// UploadRawEmail uploads the raw email content to S3 and returns the S3 object key.
func (s *S3StorageService) UploadRawEmail(ctx context.Context, messageID string, content io.Reader) (string, error) {
	txmanager, err := s.transferManager(ctx)
	if err != nil {
		return "", err
	}
	key := ingestionS3Key(s.config.IngestionPrefix, messageID)
	_, err = txmanager.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(s.config.S3Bucket),
		Key:    aws.String(key),
		Body:   content,
	})
	if err != nil {
		return "", err
	}
	return key, nil
}

func (s *S3StorageService) transferManager(ctx context.Context) (*transfermanager.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return cached client if already initialized
	if s.client != nil {
		return s.client, nil
	}

	// Initialize S3 client
	cfg, err := awscfg.LoadDefaultConfig(ctx, awscfg.WithRegion(s.config.AwsRegion), awscfg.WithBaseEndpoint(""))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if s.config.S3BaseEndpoint != "" {
			o.BaseEndpoint = aws.String(s.config.S3BaseEndpoint)
			o.UsePathStyle = s.config.UsePathStyle // Use path-style addressing for S3
		}
	})
	c := transfermanager.New(s3client, func(o *transfermanager.Options) {
		o.PartSizeBytes = 5 * 1024 * 1024 // 5 MiB
	})

	// Cache client on success
	s.client = c
	return s.client, nil
}

func ingestionS3Key(ingestionPrefix, messageID string) string {
	if strings.HasSuffix(ingestionPrefix, "/") {
		return ingestionPrefix + messageID
	}
	return fmt.Sprintf("%s/%s", ingestionPrefix, messageID)
}
