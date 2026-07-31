package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const payloadKey = "payload"

type StreamService struct {
	client *redis.Client
	// long polling duration used when reading the stream
	block time.Duration
}

func (s *StreamService) Publish(ctx context.Context, stream string, values any) error {
	m := make(map[string]any)
	m[payloadKey] = values
	return s.client.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: m}).Err()
}

func (s *StreamService) CreateGroup(ctx context.Context, streamName, groupName string) error {
	// MKSTREAM creates the stream automatically if it doesn't exist
	err := s.client.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && !isGroupExistsErr(err) {
		return err
	}
	return nil
}

func (s *StreamService) Consume(ctx context.Context, streamName, groupName, consumerID string) (*StreamData, error) {
	// ">" means fetch new messages that haven't been delivered to any consumer in this group
	streams, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerID,
		Streams:  []string{streamName, ">"},
		Count:    1,
		Block:    s.block, // Wait up to 2 seconds for new messages
	}).Result()

	if errors.Is(err, redis.Nil) || len(streams) == 0 {
		return nil, nil // No new messages
	} else if err != nil {
		// slog.ErrorContext(ctx, "failed to read redis stream", slog.String("consumerID", consumerID), slog.Any("error", err))
		return nil, err
	}
	m := streams[0].Messages[0]
	d := &StreamData{
		MessageID:  m.ID,
		Data:       m.Values[payloadKey],
		StreamName: streamName,
		GroupName:  groupName,
		ConsumerID: consumerID,
	}
	return d, nil
}

func (s *StreamService) MarkCompleted(ctx context.Context, data *StreamData) error {
	if data == nil {
		return nil
	}
	r := s.client.XAck(ctx, data.StreamName, data.GroupName, data.MessageID)
	return r.Err()
}

func (s *StreamService) CheckPending(ctx context.Context, streamName, groupName, consumerID string) ([]*StreamData, error) {
	messages, _, err := s.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   streamName,
		Group:    groupName,
		Consumer: consumerID,
		MinIdle:  30 * time.Second,
		Start:    "0-0",
		Count:    10,
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	data := make([]*StreamData, len(messages))
	for _, m := range messages {
		d := &StreamData{
			MessageID:  m.ID,
			Data:       m.Values[payloadKey],
			StreamName: streamName,
			GroupName:  groupName,
			ConsumerID: consumerID,
		}
		data = append(data, d)
	}

	return data, nil
}

func isGroupExistsErr(err error) bool {
	return err != nil && (err.Error() == "BUSYGROUP Consumer Group name already exists")
}
