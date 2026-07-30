package redis

import (
	"log/slog"
	"os"
	"time"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/redis/go-redis/v9"
)

type Manager struct {
	client *redis.Client
	Cache  *CacheService
	Stream *StreamService
}

func NewManager(cfg *config.AppConfig) *Manager {
	o, err := redis.ParseURL(cfg.Database.Redis)
	if err != nil {
		slog.Error("failed to initialize redis client", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(o)

	return &Manager{
		client: rdb,
		Cache:  &CacheService{client: rdb},
		Stream: &StreamService{
			client: rdb,
			block:  20 * time.Second,
		},
	}
}

func (m *Manager) Close() error {
	return m.client.Close()
}
