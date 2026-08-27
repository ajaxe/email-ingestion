package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/google/uuid"
)

// MessageHandler is an interface for handling specific payloads from a stream.
type MessageHandler interface {
	// Handle processes the raw stream message.
	// Returning an error indicates a transient failure and prevents the message from being acknowledged,
	// allowing it to be auto-claimed and retried later.
	// Returning nil indicates the message was successfully processed (or permanently failed) and should be acknowledged.
	Handle(ctx context.Context, msg *redis.StreamData) error
}

// StreamConsumer continuously consumes messages from a specific Redis stream and passes them to a handler.
type StreamConsumer struct {
	redisManager *redis.Manager
	streamName   string
	groupName    string
	consumerID   string
	handler      MessageHandler
}

func NewStreamConsumer(redisManager *redis.Manager, streamName string, groupName string, handler MessageHandler) *StreamConsumer {
	return &StreamConsumer{
		redisManager: redisManager,
		streamName:   streamName,
		groupName:    groupName,
		consumerID:   uuid.New().String(),
		handler:      handler,
	}
}

func (w *StreamConsumer) Start(ctx context.Context) error {
	err := w.initGroup(ctx)
	if err != nil {
		slog.Error("failed to create stream group", "stream", w.streamName, "group", w.groupName, "error", err)
		return err
	}

	go w.autoClaimLoop(ctx)

	slog.InfoContext(ctx, "starting main data processing loop", "stream", w.streamName, "group", w.groupName)
	logCtr := -1

	for {
		select {
		case <-ctx.Done(): // cancellation received from host
			return ctx.Err()
		default:
			// read one message at a time for processing
			data, err := w.redisManager.Stream.Consume(ctx, w.streamName, w.groupName, w.consumerID)
			if err != nil {
				slog.WarnContext(ctx, "failed to read data from stream", "consumerID", w.consumerID, "stream", w.streamName, "error", err)
				continue
			}

			if data == nil {
				if logCtr == -1 || logCtr%10 == 0 {
					logCtr = 0
					slog.InfoContext(ctx, "no data for processing", "consumerID", w.consumerID, "stream", w.streamName)
				}
				logCtr++
				continue
			}
			w.processMessage(ctx, data)
		}
	}
}

func (w *StreamConsumer) initGroup(ctx context.Context) error {
	err := w.redisManager.Stream.CreateGroup(ctx, w.streamName, w.groupName)
	return err
}

func (w *StreamConsumer) autoClaimLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	slog.InfoContext(ctx, "starting auto-claim loop", "stream", w.streamName)
	logCtr := -1

	for {
		select {
		case <-ctx.Done(): // cancellation received from host
			return
		case <-ticker.C:
			// Claim messages pending for more than 30 seconds
			data, err := w.redisManager.Stream.CheckPending(ctx, w.streamName, w.groupName, w.consumerID)

			if err != nil {
				slog.WarnContext(ctx, "autoclaim error", "consumerID", w.consumerID, "stream", w.streamName, "error", err)
				continue
			}

			if len(data) == 0 {
				if logCtr == -1 || logCtr%10 == 0 {
					logCtr = 0
					slog.InfoContext(ctx, "no auto-claimed data for processing", "consumerID", w.consumerID, "stream", w.streamName)
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

func (w *StreamConsumer) processMessage(ctx context.Context, msg *redis.StreamData) {
	err := w.handler.Handle(ctx, msg)

	if err != nil {
		slog.ErrorContext(ctx, "handler returned transient error, skipping ack", "consumerID", w.consumerID, "stream", w.streamName, "error", err)
		return
	}

	ackErr := w.redisManager.Stream.MarkCompleted(ctx, msg)
	if ackErr != nil {
		slog.ErrorContext(ctx, "failed to mark message as complete", "consumerID", w.consumerID, "stream", w.streamName, "error", ackErr)
	} else {
		slog.InfoContext(ctx, "message processed and acknowledged", "consumerID", w.consumerID, "stream", w.streamName)
	}
}
