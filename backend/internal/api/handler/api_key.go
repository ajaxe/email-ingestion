package handler

import (
	"net/http"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func HandleCreateAPIKey(apiKeyService *service.ApiKeyService) echo.HandlerFunc {
	return func(c echo.Context) error {
		a := c.Param("app_id")
		appID, err := uuid.Parse(a)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		var req dto.CreateAPIKeyRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		ctx := c.Request().Context()
		apiKey, err := apiKeyService.CreateAPIKey(ctx, appID, req)
		if err != nil {
			return apperror.Internal("failed to create api key", err)
		}
		return c.JSON(http.StatusCreated, dto.CreateApiKeyResponse{
			APIKey: apiKey,
		})

	}
}
