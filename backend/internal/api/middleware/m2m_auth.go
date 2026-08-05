package middleware

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// M2MAuth middleware checks for a valid machine-to-machine (M2M) authentication token in the request headers.
// checks authorization hear token against the api_keys table in the database.
func M2MAuth(authz *service.AuthorizationService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// M2M auth check from api_keys table using Authorization header bearer token.
			// The token is expected to be in the format "Bearer <token>"
			ctx := c.Request().Context()
			authHeader := c.Request().Header.Get("Authorization")
			p := strings.Split(strings.TrimSpace(authHeader), " ")
			if len(p) != 2 || p[1] == "" {
				return apperror.Unauthorized("invalid authorization header")
			}

			ok, appID, err := authz.ValidateApiKey(ctx, p[1])
			if err != nil {
				slog.ErrorContext(ctx, "failed to validate api key", "error", err)
				return apperror.Unauthorized("unauthorized")
			}

			if !ok {
				slog.ErrorContext(ctx, "invalid api key or it is expired")
				return apperror.Unauthorized("invalid authorization header")
			}

			ctx = context.WithValue(ctx, model.ApplicationIDContextKey, appID)
			n := c.Request().WithContext(ctx)
			c.SetRequest(n)

			return next(c)
		}
	}
}
