package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type AuthorizationService struct {
	cache         *redis.CacheService
	apiKeyService *ApiKeyService
}

func NewAuthorizationService(cache *redis.CacheService, apiKeyService *ApiKeyService) *AuthorizationService {
	return &AuthorizationService{
		cache:         cache,
		apiKeyService: apiKeyService,
	}
}

func (a *AuthorizationService) ValidateApiKey(ctx context.Context, apiKey string) (ok bool, appID uuid.UUID, err error) {
	k := authCacheKey(apiKey)
	j, found, err := a.cache.Get(ctx, k)

	if err != nil {
		return false, uuid.Nil, err
	}

	key := &public.ApiKey{}
	if found {
		err = json.Unmarshal([]byte(j), key)
		if err != nil {
			return false, uuid.Nil, err
		}
	} else {
		// fetch and set the cache
		key, err = a.apiKeyService.GetApiKey(ctx, apiKey)
		if err != nil {
			return false, uuid.Nil, err
		}
		b, err := util.JSON(key)
		if err != nil {
			return false, uuid.Nil, err
		}
		ttl := time.Until(key.ExpiresAt)
		if ttl > 30*time.Minute {
			ttl = 30 * time.Minute
		}
		err = a.cache.Set(ctx, k, b, ttl)
		if err != nil {
			slog.WarnContext(ctx, "failed to set auth api key in cache", "key", k, "error", err)
		}
	}
	valid := key.ExpiresAt.After(time.Now())
	return valid, key.ApplicationID, nil
}
func authCacheKey(apiKey string) string {
	return fmt.Sprintf("authz:api-key:%s", apiKey)
}
