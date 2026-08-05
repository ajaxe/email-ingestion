package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type ApiKeyService struct {
	queries *public.Queries
}

func NewApiKeyService(queries *public.Queries) *ApiKeyService {
	return &ApiKeyService{
		queries: queries,
	}
}

func (a *ApiKeyService) GetApiKey(ctx context.Context, apiKey string) (*public.ApiKey, error) {
	hashBytes := sha256.Sum256([]byte(apiKey))
	k, err := a.queries.GetApiKeyByKeyHash(ctx, hex.EncodeToString(hashBytes[:]))
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (a *ApiKeyService) CreateAPIKey(ctx context.Context, appID uuid.UUID, data model.CreateAPIKeyRequest) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	apiKey := hex.EncodeToString(b)

	hashApiKey := sha256.Sum256([]byte(apiKey))
	err = a.queries.CreateApiKey(ctx, public.CreateApiKeyParams{
		ApplicationID: appID,
		Name:          data.Name,
		KeyPrefix:     data.KeyPrefix,
		KeyHash:       hex.EncodeToString(hashApiKey[:]),
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(time.Duration(data.ExpireDays) * 24 * time.Hour),
	})
	if err != nil {
		return "", err
	}
	return apiKey, nil
}
