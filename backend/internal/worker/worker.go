package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

const groupName = "worker_group"

type EmailWorker struct {
	cfg          *config.AppConfig
	processor    *EmailProcessor
	redisManager *redis.Manager
	streamName   string
	consumerID   string
}

func (w *EmailWorker) Start(ctx context.Context) {
	err := w.initGroup(ctx)
	if err != nil {
		slog.Error("failed to create stream group", "error", err)
		return
	}

	go w.autoClaimLoop(ctx)

	for {
		select {
		case <-ctx.Done(): // cancellation received from host
			return
		default:
			// read one message at a time for processing
			data, err := w.redisManager.Stream.Consume(ctx, w.streamName, groupName, w.consumerID)
			if err != nil {
				slog.WarnContext(ctx, "failed to read data from stream", "consumerID", w.consumerID, "error", err)
				return
			}
			if data == nil {
				slog.InfoContext(ctx, "no data for processing", "consumerID", w.consumerID)
			}
			w.processMessage(ctx, data)
			if err != nil {
				slog.ErrorContext(ctx, "failed to mark message as completed", "consumerID", w.consumerID, "error", err)
			}
		}
	}
}

func (w *EmailWorker) initGroup(ctx context.Context) error {
	err := w.redisManager.Stream.CreateGroup(ctx, w.streamName, groupName)
	return err
}

func (w *EmailWorker) autoClaimLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // cancellation received from host
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
				err = w.redisManager.Stream.MarkCompleted(ctx, msg)
				if err != nil {
					slog.ErrorContext(ctx, "failed to mark message as completed", "consumerID", w.consumerID, "error", err)
				}
			}
		}
	}
}

func (w *EmailWorker) processMessage(ctx context.Context, msg *redis.StreamData) error {
	j, ok := msg.Data.(string)
	if !ok {
		slog.ErrorContext(ctx, "failed to parse message data", "consumerID", w.consumerID)
	}
	p, err := model.ParseIngestEmail(j)
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse message data", "consumerID", w.consumerID, "error", err)
	}

	err = w.processor.Process(ctx, p)

	if err != nil {
		slog.ErrorContext(ctx, "failed to process message", "consumerID", w.consumerID, "error", err)
	} else {
		err = w.redisManager.Stream.MarkCompleted(ctx, msg)

		if err != nil {
			slog.ErrorContext(ctx, "failed to mark message as complete", "consumerID", w.consumerID, "error", err)
		} else {
			slog.InfoContext(ctx, "message processed successfully")
		}
	}
	return err
}

func New(cfg *config.AppConfig, queries *public.Queries, redisManager *redis.Manager, storageService *storage.S3StorageService) *EmailWorker {
	s := service.StreamName(config.AppName, cfg.Environment)
	p := &EmailProcessor{
		queries:        queries,
		storageService: storageService,
	}
	return &EmailWorker{
		cfg:          cfg,
		processor:    p,
		redisManager: redisManager,
		streamName:   s,
		consumerID:   uuid.New().String(),
	}
}
