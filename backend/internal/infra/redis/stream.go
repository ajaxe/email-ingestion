package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type StreamService struct {
	client *redis.Client
}

func (s *StreamService) Publish(ctx context.Context, stream string, values any) error {
	return s.client.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: map[string]interface{}{
		"payload": values,
	}}).Err()
}
