package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
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

func (a *ApiKeyService) ListAPIKeys(ctx context.Context, appID uuid.UUID) ([]dto.APIKeyResponse, error) {
	rows, err := a.queries.ListApiKeysByApplication(ctx, appID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.APIKeyResponse, 0, len(rows))
	for _, row := range rows {
		res = append(res, dto.APIKeyResponse{
			ID:            row.ID,
			ApplicationID: row.ApplicationID,
			Name:          row.Name,
			KeyPrefix:     row.KeyPrefix,
			CreatedAt:     row.CreatedAt,
			ExpiresAt:     row.ExpiresAt,
			LastUsedAt:    row.LastUsedAt,
		})
	}
	return res, nil
}

func (a *ApiKeyService) CreateAPIKey(ctx context.Context, appID uuid.UUID, data dto.CreateAPIKeyRequest) (string, error) {
	b := make([]byte, 24)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	randomPart := hex.EncodeToString(b)

	prefix := data.KeyPrefix
	if prefix == "" {
		prefix = "eg_live_"
	}

	name := data.Name
	if name == "" {
		name = "Ingestion API Key"
	}

	expireDays := data.ExpireDays
	if expireDays <= 0 {
		expireDays = 365
	}

	fullKey := prefix + randomPart
	hashApiKey := sha256.Sum256([]byte(fullKey))

	err = a.queries.CreateApiKey(ctx, public.CreateApiKeyParams{
		ApplicationID: appID,
		Name:          name,
		KeyPrefix:     prefix,
		KeyHash:       hex.EncodeToString(hashApiKey[:]),
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(time.Duration(expireDays) * 24 * time.Hour),
	})
	if err != nil {
		return "", err
	}
	return fullKey, nil
}

func (a *ApiKeyService) RevokeAPIKey(ctx context.Context, appID uuid.UUID, keyID uuid.UUID) error {
	return a.queries.DeleteApiKey(ctx, public.DeleteApiKeyParams{
		ID:            keyID,
		ApplicationID: appID,
	})
}
