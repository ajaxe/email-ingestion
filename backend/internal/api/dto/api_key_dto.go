package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateAPIKeyRequest struct {
	Name       string `json:"name"`
	KeyPrefix  string `json:"keyPrefix"`
	ExpireDays int64  `json:"expireDays"`
}

type CreateApiKeyResponse struct {
	APIKey string `json:"apiKey"`
}

type APIKeyResponse struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"applicationId"`
	Name          string    `json:"name"`
	KeyPrefix     string    `json:"keyPrefix"`
	CreatedAt     time.Time `json:"createdAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
	LastUsedAt    time.Time `json:"lastUsedAt"`
}

