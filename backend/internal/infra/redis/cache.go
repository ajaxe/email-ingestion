package redis

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
}

func (c *CacheService) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	e := c.client.Set(ctx, key, val, ttl).Err()
	if e != nil {
		slog.ErrorContext(ctx, "failed to set redis key", slog.String("key", key), slog.Any("error", e))
	}
	return e
}

func (c *CacheService) Get(ctx context.Context, key string) (v string, found bool, err error) {
	r := c.client.Get(ctx, key)
	if r.Err() == redis.Nil {
		return "", false, nil
	} else if r.Err() != nil {
		e := r.Err()
		slog.ErrorContext(ctx, "failed to get redis key", slog.String("key", key), slog.Any("error", e))
		return "", false, e
	}
	return r.Val(), true, nil
}
