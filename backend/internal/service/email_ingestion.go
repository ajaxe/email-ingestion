package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/emersion/go-msgauth/dkim"
	"github.com/google/uuid"
)

func NewEmailIngestion(redisManager *redis.Manager, queries *public.Queries, storageService *storage.S3StorageService, env string) *EmailIngestionService {
	p := util.EmailStreamName(env)
	return &EmailIngestionService{
		redisManager:   redisManager,
		queries:        queries,
		storageService: storageService,
		streamName:     p,
	}
}

type EmailIngestionService struct {
	redisManager   *redis.Manager
	queries        *public.Queries
	storageService *storage.S3StorageService
	streamName     string
}

// LookupAssignedEmail looks up the database for the assigned email. It caches the result whether it finds the email or not.
func (e *EmailIngestionService) LookupAssignedEmail(ctx context.Context, to string) (bool, error) {
	parts := strings.Split(to, "@")
	if len(parts) != 2 {
		return false, fmt.Errorf("Invalid email")
	}
	localPart := parts[0]

	baseLocalPart := localPart
	if plusIdx := strings.Index(localPart, "+"); plusIdx != -1 {
		baseLocalPart = localPart[:plusIdx]
	}

	cacheKey := fmt.Sprintf("assigned_email:%s:found", baseLocalPart)
	// Check cache
	var isValid bool
	if val, found, _ := e.redisManager.Cache.Get(ctx, cacheKey); found {
		isValid = "true" == val
	} else {
		// Cache miss, check db
		ctx := context.Background()
		_, err := e.queries.GetAssignedEmailByLocalPart(ctx, baseLocalPart) // Only check the first 10 characters of the local part
		if err != nil {
			// Not found or db error
			isValid = false
			e.redisManager.Cache.Set(ctx, cacheKey, "false", 5*time.Minute)
		} else {
			isValid = true
			e.redisManager.Cache.Set(ctx, cacheKey, "true", 1*time.Hour)
		}
	}

	if !isValid {
		return false, fmt.Errorf("User Unknown")
	}

	return true, nil
}

// Process handles the data read operation for the SMTP server. It processes the incoming email and stores it in the database.
// It publishes a queue message for background processing.
func (e *EmailIngestionService) Process(ctx context.Context, data io.Reader) error {
	dataCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	pr, pw := io.Pipe()
	tee := io.TeeReader(data, pw)

	var verifications []*dkim.Verification
	var dkimErr error
	done := make(chan struct{})

	go func() {
		defer close(done)
		verifications, dkimErr = dkim.Verify(pr)
		// dkim.Verify might not consume the entire reader. We must drain it
		// so that the TeeReader (S3 upload) is not blocked.
		io.Copy(io.Discard, pr)
	}()

	identifier := uuid.New()
	id := identifier.String()

	key := util.IngestionS3Key(e.storageService.Config().IngestionPrefix, id)
	uploadKey, err := e.storageService.UploadObject(dataCtx, key, tee, "")

	// Close the pipe writer to signal EOF (or error) to the reader goroutine
	pw.CloseWithError(err)
	<-done // Wait for DKIM verification to finish

	if dkimErr != nil {
		slog.Warn("DKIM verification encountered an error", "error", dkimErr, "identifier", id)
	} else {
		for _, v := range verifications {
			if v.Err != nil {
				slog.Info("DKIM signature invalid", "domain", v.Domain, "error", v.Err, "identifier", id)
			} else {
				slog.Info("DKIM signature valid", "domain", v.Domain, "identifier", id)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to upload raw email: %v", err)
	}

	_, err = e.queries.CreateInboundSpooledEmail(dataCtx, public.CreateInboundSpooledEmailParams{
		ID:               identifier,
		S3ObjectKey:      uploadKey,
		Status:           "PENDING",
		AttemptCount:     0,
		LastErrorMessage: "",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	})

	if err != nil {
		return fmt.Errorf("failed to log inbound email to database")
	}

	p := &model.IngestEmailPayload{
		SpoolID:   id,
		UploadKey: uploadKey,
	}

	j, err := p.JSON()
	if err != nil {
		return fmt.Errorf("failed to serialize payload for redis stream: id-%s: %v", id, err)
	}

	err = e.redisManager.Stream.Publish(ctx, e.streamName, j)
	if err != nil {
		return fmt.Errorf("failed to enqueue email for processing: id-%s: %v", id, err)
	}

	return nil
}
