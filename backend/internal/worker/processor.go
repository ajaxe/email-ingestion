package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

const groupName = "worker_group"

type EmailProcessor struct {
	cfg            *config.AppConfig
	queries        *public.Queries
	storageService *storage.S3StorageService
	wg             sync.WaitGroup
	redisManager   *redis.Manager
	streamName     string
	consumerID     string
}

func (w *EmailProcessor) Start(ctx context.Context) {
	err := w.initGroup(ctx)
	if err != nil {
		slog.Error("failed to create stream group", "error", err)
		return
	}

	go w.autoClaimLoop(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			data, err := w.redisManager.Stream.Consume(ctx, w.streamName, groupName, w.consumerID)
			if err != nil {
				slog.WarnContext(ctx, "failed to read data from stream", "consumerID", w.consumerID, "error", err)
				return
			}
			w.processMessage(ctx, data)
		}
	}
}

func (w *EmailProcessor) initGroup(ctx context.Context) error {
	err := w.redisManager.Stream.CreateGroup(ctx, w.streamName, groupName)
	return err
}

func (w *EmailProcessor) autoClaimLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Claim messages pending for more than 30 seconds
			data, err := w.redisManager.Stream.CheckPending(ctx, w.streamName, groupName, w.consumerID)

			if err != nil {
				slog.WarnContext(ctx, "autoclaim error", "consumerID", w.consumerID, "error", err)
				continue
			}

			for _, msg := range data {
				w.processMessage(ctx, msg)
			}
		}
	}
}

func (w *EmailProcessor) processMessage(ctx context.Context, msg *redis.StreamData) {
	// TODO: Implement message processing logic here

}

func NewProcessor(cfg *config.AppConfig, queries *public.Queries, redisManager *redis.Manager, storageService *storage.S3StorageService) *EmailProcessor {
	p := service.StreamName(config.AppName, cfg.Environment)
	return &EmailProcessor{
		cfg:            cfg,
		queries:        queries,
		storageService: storageService,
		redisManager:   redisManager,
		streamName:     p,
		consumerID:     uuid.New().String(),
	}
}
