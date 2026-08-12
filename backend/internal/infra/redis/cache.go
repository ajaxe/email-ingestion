package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client         *redis.Client
	cacheKeyPrefix string
}

func (c *CacheService) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	k := fmt.Sprintf("%s%s", c.cacheKeyPrefix, key)
	e := c.client.Set(ctx, k, val, ttl).Err()
	if e != nil {
		slog.ErrorContext(ctx, "failed to set redis key", slog.String("key", k), slog.Any("error", e))
	}
	return e
}

func (c *CacheService) Get(ctx context.Context, key string) (v string, found bool, err error) {
	k := fmt.Sprintf("%s%s", c.cacheKeyPrefix, key)
	r := c.client.Get(ctx, k)
	if r.Err() == redis.Nil {
		return "", false, nil
	} else if r.Err() != nil {
		e := r.Err()
		slog.ErrorContext(ctx, "failed to get redis key", slog.String("key", k), slog.Any("error", e))
		return "", false, e
	}
	return r.Val(), true, nil
}

func (c *CacheService) GetValue(ctx context.Context, key string, v any) (found bool, err error) {
	s, found, err := c.Get(ctx, key)
	if found {
		err = json.Unmarshal([]byte(s), v)
	}
	return found, err
}
