package middleware

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/labstack/echo/v4"
)

type UserAccessVerifier interface {
	VerifyToken(ctx context.Context, token string) (*dto.UserProfile, error)
	PermittedApplications(ctx context.Context, userID string) ([]public.Application, error)
}

func AppAuth(verifier UserAccessVerifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			authHeader := c.Request().Header.Get("Authorization")
			p := strings.Split(strings.TrimSpace(authHeader), " ")
			if len(p) != 2 || p[1] == "" {
				return apperror.Unauthorized("invalid authorization header")
			}
			profile, err := verifier.VerifyToken(ctx, p[1])

			if err != nil {
				slog.ErrorContext(ctx, "failed to validate api key", "error", err)
				return apperror.Forbidden("invalid token", err)
			}

			apps, err := verifier.PermittedApplications(ctx, profile.UserID)

			if err != nil {
				slog.ErrorContext(ctx, "failed to fetch permitted applications", "error", err)
				return apperror.Forbidden("invalid token", err)
			}

			ctx = WithUserAccess(ctx, &dto.UserAccessResult{
				UserProfile:  profile,
				Applications: apps,
			})
			n := c.Request().WithContext(ctx)
			c.SetRequest(n)

			return next(c)
		}
	}
}
