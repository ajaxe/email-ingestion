package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/util"
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

	slog.InfoContext(ctx, "starting main data processing loop")
	logCtr := -1

	for {
		select {
		case <-ctx.Done(): // cancellation received from host
			return
		default:
			// read one message at a time for processing
			data, err := w.redisManager.Stream.Consume(ctx, w.streamName, groupName, w.consumerID)
			if err != nil {
				slog.WarnContext(ctx, "failed to read data from stream", "consumerID", w.consumerID, "error", err)
				continue
			}

			if data == nil {
				if logCtr == -1 || logCtr%10 == 0 {
					logCtr = 0
					slog.InfoContext(ctx, "no data for processing", "consumerID", w.consumerID)
				}
				logCtr++
				continue
			}
			w.processMessage(ctx, data)
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

	slog.InfoContext(ctx, "starting auto-claim loop")
	logCtr := -1

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

			if len(data) == 0 {
				if logCtr == -1 || logCtr%10 == 0 {
					logCtr = 0
					slog.InfoContext(ctx, "no auto-claimed data for processing", "consumerID", w.consumerID)
				}
				logCtr++
				continue
			}

			for _, msg := range data {
				w.processMessage(ctx, msg)
			}
		}
	}
}

func (w *EmailWorker) processMessage(ctx context.Context, msg *redis.StreamData) error {
	j, ok := msg.Data.(string)
	if !ok {
		slog.ErrorContext(ctx, "failed to parse message data", "consumerID", w.consumerID)
		return fmt.Errorf("failed to parse message data")
	}
	p, err := model.ParseIngestEmail(j)
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse message data", "consumerID", w.consumerID, "error", err)
		return err
	}

	err = w.processor.Process(ctx, p)

	if err != nil {
		slog.ErrorContext(ctx, "failed to process message", "spoolID", p.SpoolID, "consumerID", w.consumerID, "error", err)
		var retryErr *RetryableError
		if errors.As(err, &retryErr) {
			// Do not ack, return so it can be retried
			return err
		}
	}

	ackErr := w.redisManager.Stream.MarkCompleted(ctx, msg)
	if ackErr != nil {
		slog.ErrorContext(ctx, "failed to mark message as complete", "spoolID", p.SpoolID, "consumerID", w.consumerID, "error", ackErr)
	} else {
		slog.InfoContext(ctx, "message processed and acknowledged", "spoolID", p.SpoolID, "consumerID", w.consumerID)
	}

	return err
}

func New(cfg *config.AppConfig, queries *public.Queries, redisManager *redis.Manager, storageService *storage.S3StorageService) *EmailWorker {
	s := util.StreamName(config.AppName, cfg.Environment)
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
