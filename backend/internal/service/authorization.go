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

func (a *AuthorizationService) ProvisionUser(ctx context.Context, userData *UserProvisionData) (*public.User, error) {
	if userData == nil {
		return nil, fmt.Errorf("user data is nil")
	}
	if userData.Email == "" {
		return nil, fmt.Errorf("user data - email is nil")
	}
	if userData.Subject == "" {
		return nil, fmt.Errorf("user data - subject is nil")
	}

	key := userData.CacheKey()

	u := &public.User{}
	found, err := a.cache.GetValue(ctx, key, u)

	if err != nil {
		slog.WarnContext(ctx, "failed to get cached provisioned user", "error", err)
	}

	if found {
		slog.InfoContext(ctx, "found cached provisioned user", "userID", u.ID.String())
		return u, nil
	}

	// try provisioning using user Subject ID first, if status already "active" then no update is needed.
	user, ok, err := a.provisionBySubject(ctx, userData)

	if err != nil {
		return nil, err
	}

	if !ok {
		// try provisioning using user Email
		slog.InfoContext(ctx, "user provision - could not locate user by subject")
		user, ok, err = a.provisionByEmail(ctx, userData)
		if err != nil {
			return nil, err
		}
	}

	if !ok {
		return nil, fmt.Errorf("failed to provision user by subject or email, missing partial user data")
	}

	if err := a.ensurePersonalOrg(ctx, user); err != nil {
		slog.WarnContext(ctx, "failed to ensure personal organization during provision", "userID", user.ID, "error", err)
	}

	json, err := util.JSON(user)
	if err != nil {
		return nil, err
	}

	err = a.cache.Set(ctx, key, json, 5*time.Minute)

	if err != nil {
		slog.WarnContext(ctx, "failed to cache provisioned user", "error", err)
	}

	return user, nil
}

func (a *AuthorizationService) ensurePersonalOrg(ctx context.Context, user *public.User) error {
	orgName := user.Email
	if orgName == "" {
		orgName = fmt.Sprintf("User %s Org", user.ID.String()[:8])
	}
	_, err := a.queries.CreatePersonalOrganization(ctx, public.CreatePersonalOrganizationParams{
		Name:        orgName,
		OwnerUserID: user.ID,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to ensure personal organization for user", "userID", user.ID, "error", err)
		return err
	}
	return nil
}

func (a *AuthorizationService) provisionBySubject(ctx context.Context, userData *UserProvisionData) (*public.User, bool, error) {
	u, err := a.queries.GetUserBySubject(ctx, userData.Subject)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if u.Status == "active" {
		return &u, true, nil
	}

	args := public.UpdateUserParams{
		ID:          u.ID,
		Email:       userData.Email,
		IdpUserSub:  userData.Subject,
		Status:      "active",
		CreatedAt:   u.CreatedAt,
		ActivatedAt: time.Now(),
		LastLoginAt: time.Now(),
	}
	err = a.queries.UpdateUser(ctx, args)
	if err != nil {
		return nil, false, err
	}
	slog.InfoContext(ctx, "provisioned user by subject", "subject", userData.Subject)

	u.Email = args.Email
	u.Status = args.Status
	u.ActivatedAt = args.ActivatedAt
	u.LastLoginAt = args.LastLoginAt

	return &u, true, nil
}

func (a *AuthorizationService) provisionByEmail(ctx context.Context, userData *UserProvisionData) (*public.User, bool, error) {
	u, err := a.queries.GetUserByEmail(ctx, userData.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	args := public.UpdateUserParams{
		ID:          u.ID,
		Email:       userData.Email,
		IdpUserSub:  userData.Subject,
		Status:      "active",
		CreatedAt:   u.CreatedAt,
		ActivatedAt: time.Now(),
		LastLoginAt: time.Now(),
	}
	err = a.queries.UpdateUser(ctx, args)
	if err != nil {
		return nil, false, err
	}
	slog.InfoContext(ctx, "provisioned user by email", "subject", userData.Subject, "userID", u.ID)

	u.IdpUserSub = args.IdpUserSub
	u.Status = args.Status
	u.ActivatedAt = args.ActivatedAt
	u.LastLoginAt = args.LastLoginAt

	return &u, true, nil
}

func authCacheKey(apiKey string) string {
	return fmt.Sprintf("authz:api-key:%s", apiKey)
}
