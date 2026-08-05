package storage

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/ajaxe/email-ingestion/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3StorageService struct {
	config    *config.StorageConfig
	mu        sync.Mutex
	txmanager *transfermanager.Client
	s3client  *s3.Client
}

func NewStorageService(config *config.StorageConfig) *S3StorageService {
	return &S3StorageService{
		config: config,
	}
}

func (s *S3StorageService) Config() *config.StorageConfig {
	return s.config
}

// UploadRawEmail uploads the raw email content to S3 and returns the S3 object key.
func (s *S3StorageService) UploadObject(ctx context.Context, key string, content io.Reader, contentType string) (string, error) {
	txmanager, _, err := s.transferManager(ctx)
	if err != nil {
		return "", err
	}

	_, err = txmanager.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(s.config.S3Bucket),
		Key:         aws.String(key),
		Body:        content,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return key, nil
}

func (s *S3StorageService) DownloadObject(ctx context.Context, key string) (io.ReadCloser, error) {
	_, s3client, err := s.transferManager(ctx)
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return nil, fmt.Errorf("empty object key")
	}

	data, err := s3client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.S3Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	return data.Body, nil
}

func (s *S3StorageService) DeleteObject(ctx context.Context, key string) error {
	_, s3client, err := s.transferManager(ctx)
	if err != nil {
		return err
	}

	_, err = s3client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.S3Bucket),
		Key:    aws.String(key),
	})

	return err
}

func (s *S3StorageService) transferManager(ctx context.Context) (*transfermanager.Client, *s3.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return cached client if already initialized
	if s.txmanager != nil {
		return s.txmanager, s.s3client, nil
	}

	// Initialize S3 client
	cfg, err := awscfg.LoadDefaultConfig(ctx, awscfg.WithRegion(s.config.AwsRegion))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load AWS config: %w", err)
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
	s.txmanager = c
	s.s3client = s3client
	return s.txmanager, s.s3client, nil
}
