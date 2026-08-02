package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/ajaxe/email-ingestion/internal/webhook"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/crypto"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type WebhookService struct {
	queries *public.Queries
	cfg     *config.WebhookConfig
}

func NewWebhookService(queries *public.Queries, cfg *config.WebhookConfig) *WebhookService {
	return &WebhookService{
		queries: queries,
		cfg:     cfg,
	}
}

func (s *WebhookService) RegisterWebhook(ctx context.Context, appID uuid.UUID, webhookURL string) (string, error) {
	app, err := s.queries.GetApplicationByID(ctx, appID)
	if err != nil {
		return "", apperror.NotFound("application not found", err)
	}

	// Perform SSRF Challenge Handshake
	client := webhook.NewSSRFProtectedClient(s.cfg, app.IsTrusted)
	if err := webhook.PerformChallengeHandshake(ctx, client, webhookURL); err != nil {
		return "", apperror.UnprocessableEntity("webhook verification failed", err.Error(), err)
	}

	// Generate a new webhook secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", apperror.Internal("failed to generate webhook secret", err)
	}
	webhookSecret := hex.EncodeToString(secretBytes)

	// Encrypt the webhook secret before saving it to the database
	encryptedSecret, err := crypto.Encrypt(webhookSecret, s.cfg.EncryptionKey)
	if err != nil {
		return "", apperror.Internal("failed to encrypt webhook secret", err)
	}

	// Save the webhook config to the application
	err = s.queries.UpdateApplicationWebhook(ctx, public.UpdateApplicationWebhookParams{
		ID:            appID,
		WebhookUrl:    webhookURL,
		WebhookSecret: encryptedSecret,
	})
	if err != nil {
		return "", apperror.Internal("failed to save webhook configuration", err)
	}

	return webhookSecret, nil
}
