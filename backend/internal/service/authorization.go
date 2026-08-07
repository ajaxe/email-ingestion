package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type AuthorizationService struct {
	cache         *redis.CacheService
	apiKeyService *ApiKeyService
	queries       *public.Queries
}

func NewAuthorizationService(queries *public.Queries, cache *redis.CacheService, apiKeyService *ApiKeyService) *AuthorizationService {
	return &AuthorizationService{
		cache:         cache,
		apiKeyService: apiKeyService,
		queries:       queries,
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

func (a *AuthorizationService) ProvisionUser(ctx context.Context, userData *model.UserProvisionData) error {
	if userData == nil {
		return fmt.Errorf("user data is nil")
	}
	if userData.Email == "" {
		return fmt.Errorf("user data - email is nil")
	}
	if userData.Subject == "" {
		return fmt.Errorf("user data - subject is nil")
	}

	// try provisioning using user Subject ID first, if status already "active" then no update is needed.
	ok, err := a.provisionBySubject(ctx, userData)

	if err != nil {
		return err
	}

	if !ok {
		// try provisioning using user Email
		slog.InfoContext(ctx, "user provision - could not locate user by subject")
		ok, err = a.provisionByEmail(ctx, userData)
		if err != nil {
			return err
		}
	}

	if !ok {
		return fmt.Errorf("failed to provision user by subject or email, missing partial user data")
	}

	return nil
}

func (a *AuthorizationService) provisionBySubject(ctx context.Context, userData *model.UserProvisionData) (ok bool, err error) {
	u, err := a.queries.GetUserBySubject(ctx, userData.Subject)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	if u.Status == "active" {
		return true, nil
	}

	err = a.queries.UpdateUser(ctx, public.UpdateUserParams{
		ID:          u.ID,
		Email:       userData.Email,
		IdpUserSub:  userData.Subject,
		Status:      "active",
		CreatedAt:   u.CreatedAt,
		ActivatedAt: time.Now(),
		LastLoginAt: time.Now(),
	})
	if err != nil {
		return false, err
	}
	slog.InfoContext(ctx, "provisioned user by subject", "subject", userData.Subject)

	return true, nil
}

func (a *AuthorizationService) provisionByEmail(ctx context.Context, userData *model.UserProvisionData) (ok bool, err error) {
	u, err := a.queries.GetUserByEmail(ctx, userData.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	err = a.queries.UpdateUser(ctx, public.UpdateUserParams{
		ID:          u.ID,
		Email:       userData.Email,
		IdpUserSub:  userData.Subject,
		Status:      "active",
		CreatedAt:   u.CreatedAt,
		ActivatedAt: time.Now(),
		LastLoginAt: time.Now(),
	})
	if err != nil {
		return false, err
	}
	slog.InfoContext(ctx, "provisioned user by email", "subject", userData.Subject, "userID", u.ID)

	return true, nil
}

func authCacheKey(apiKey string) string {
	return fmt.Sprintf("authz:api-key:%s", apiKey)
}
