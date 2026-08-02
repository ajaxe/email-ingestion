package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/ajaxe/email-ingestion/internal/webhook"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/crypto"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RegisterWebhookRequest struct {
	WebhookURL string `json:"webhook_url"`
}

type RegisterWebhookResponse struct {
	Message       string `json:"message"`
	WebhookSecret string `json:"webhook_secret"`
}

func HandleRegisterWebhook(queries *public.Queries, cfg *config.WebhookConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		var req RegisterWebhookRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		if req.WebhookURL == "" {
			return apperror.Validation("webhook_url is required")
		}

		app, err := queries.GetApplicationByID(c.Request().Context(), appID)
		if err != nil {
			return apperror.NotFound("application not found", err)
		}

		// Perform SSRF Challenge Handshake
		client := webhook.NewSSRFProtectedClient(cfg, app.IsTrusted)
		if err := webhook.PerformChallengeHandshake(c.Request().Context(), client, req.WebhookURL); err != nil {
			return apperror.UnprocessableEntity("webhook verification failed", err.Error(), err)
		}

		// Generate a new webhook secret
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			return apperror.Internal("failed to generate webhook secret", err)
		}
		webhookSecret := hex.EncodeToString(secretBytes)

		// Encrypt the webhook secret before saving it to the database
		encryptedSecret, err := crypto.Encrypt(webhookSecret, cfg.EncryptionKey)
		if err != nil {
			return apperror.Internal("failed to encrypt webhook secret", err)
		}

		// Save the webhook config to the application
		err = queries.UpdateApplicationWebhook(c.Request().Context(), public.UpdateApplicationWebhookParams{
			ID:            appID,
			WebhookUrl:    req.WebhookURL,
			WebhookSecret: encryptedSecret,
		})
		if err != nil {
			return apperror.Internal("failed to save webhook configuration", err)
		}

		return c.JSON(http.StatusOK, RegisterWebhookResponse{
			Message:       "Webhook registered and verified successfully",
			WebhookSecret: webhookSecret,
		})
	}
}
