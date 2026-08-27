package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ajaxe/email-ingestion/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func (s *S3StorageService) DeletePrefix(ctx context.Context, prefix string) error {
	_, s3client, err := s.transferManager(ctx)
	if err != nil {
		return err
	}

	if prefix == "" {
		return fmt.Errorf("empty prefix for bulk delete")
	}

	paginator := s3.NewListObjectsV2Paginator(s3client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.S3Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects for prefix %s: %w", prefix, err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		var objectsToDelete []types.ObjectIdentifier
		for _, obj := range page.Contents {
			if obj.Key != nil {
				objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{Key: obj.Key})
			}
		}

		if len(objectsToDelete) > 0 {
			_, err = s3client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(s.config.S3Bucket),
				Delete: &types.Delete{
					Objects: objectsToDelete,
					Quiet:   aws.Bool(true),
				},
			})
			if err != nil {
				return fmt.Errorf("failed to delete objects for prefix %s: %w", prefix, err)
			}
		}
	}

	return nil
}

func (s *S3StorageService) PresignTTL() time.Duration {
	return s.config.PresignedURLTTL()
}

func (s *S3StorageService) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	_, s3client, err := s.transferManager(ctx)
	if err != nil {
		return "", err
	}

	if len(key) == 0 {
		return "", fmt.Errorf("empty object key")
	}

	if expiration <= 0 {
		expiration = s.config.PresignedURLTTL()
	}

	presignClient := s3.NewPresignClient(s3client)
	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.S3Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
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
