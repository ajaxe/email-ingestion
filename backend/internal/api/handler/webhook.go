package handler

import (
	"net/http"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func HandleRegisterWebhook(svc *service.WebhookService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		var req dto.RegisterWebhookRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		if req.WebhookURL == "" {
			return apperror.Validation("webhook_url is required")
		}

		secret, err := svc.RegisterWebhook(c.Request().Context(), appID, req.WebhookURL)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.RegisterWebhookResponse{
			Message:       "Webhook registered and verified successfully",
			WebhookSecret: secret,
		})
	}
}
