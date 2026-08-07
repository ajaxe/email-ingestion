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
	"github.com/jackc/pgx/v5/pgtype"
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

func (s *WebhookService) ListJobs(ctx context.Context, appID uuid.UUID, limit, offset int32, status string) ([]public.ListWebhookJobsByApplicationRow, error) {
	jobs, err := s.queries.ListWebhookJobsByApplication(ctx, public.ListWebhookJobsByApplicationParams{
		ApplicationID: appID,
		Limit:         limit,
		Offset:        offset,
		Status:        pgtype.Text{String: status, Valid: true},
	})

	if err != nil {
		return nil, apperror.Internal("failed to list webhook jobs", err)
	}
	return jobs, nil
}

func (s *WebhookService) RedeliverJob(ctx context.Context, appID uuid.UUID, jobID uuid.UUID) error {
	// Validate job ownership
	_, err := s.queries.GetWebhookJobByIDAndAppID(ctx, public.GetWebhookJobByIDAndAppIDParams{
		ID:            jobID,
		ApplicationID: appID,
	})
	if err != nil {
		return apperror.NotFound("webhook job not found", err)
	}

	// Reset job status
	_, err = s.queries.ResetWebhookJobForRedelivery(ctx, public.ResetWebhookJobForRedeliveryParams{
		ID:            jobID,
		ApplicationID: appID,
	})
	if err != nil {
		return apperror.Internal("failed to reset webhook job", err)
	}

	// Notify Redis outbox for immediate reprocessing
	return nil
}
