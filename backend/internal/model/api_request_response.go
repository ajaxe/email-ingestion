package model

import (
	"time"

	"github.com/ajaxe/email-ingestion/pkg/database/public"
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

type CreateAddressRequest struct {
	Description string `json:"description"`
}

type RegisterWebhookRequest struct {
	WebhookURL string `json:"webhookUrl"`
}

type RegisterWebhookResponse struct {
	Message       string `json:"message"`
	WebhookSecret string `json:"webhookSecret"`
}

type LoginAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserProfile struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Subject  string `json:"subject"`
	Email    string `json:"email"`
}

type UserAccessResult struct {
	UserProfile  *UserProfile
	Applications []public.Application
}

type ApplicationModelResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	WebhookURL string    `json:"webhookUrl"`
	MaxRetries int       `json:"maxRetries"`
	IsTrusted  bool      `json:"isTrusted"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
