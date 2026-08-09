package handler

import (
	"net/http"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func HandleListAPIKeys(apiKeyService *service.ApiKeyService) echo.HandlerFunc {
	return func(c echo.Context) error {
		a := c.Param("app_id")
		appID, err := uuid.Parse(a)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		ctx := c.Request().Context()
		keys, err := apiKeyService.ListAPIKeys(ctx, appID)
		if err != nil {
			return apperror.Internal("failed to list api keys", err)
		}
		return c.JSON(http.StatusOK, keys)
	}
}

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

func HandleRevokeAPIKey(apiKeyService *service.ApiKeyService) echo.HandlerFunc {
	return func(c echo.Context) error {
		a := c.Param("app_id")
		appID, err := uuid.Parse(a)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		k := c.Param("key_id")
		keyID, err := uuid.Parse(k)
		if err != nil {
			return apperror.Validation("invalid key ID")
		}

		ctx := c.Request().Context()
		if err := apiKeyService.RevokeAPIKey(ctx, appID, keyID); err != nil {
			return apperror.Internal("failed to revoke api key", err)
		}
		return c.NoContent(http.StatusNoContent)
	}
}
