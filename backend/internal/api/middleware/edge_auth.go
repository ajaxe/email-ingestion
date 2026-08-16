package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ajaxe/email-ingestion/internal/ingest"
	"github.com/labstack/echo/v4"
)

// EdgeAuth ensures the request has the correct edge token
func EdgeAuth(expectedToken string) echo.MiddlewareFunc {
	expected := strings.TrimSpace(expectedToken)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := strings.TrimSpace(c.Request().Header.Get(ingest.HeaderEdgeToken))
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "empty edge auth token")
			}
			if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
				p := ""
				if len(token) > 10 {
					p = token[:3] + "..." + token[len(token)-3:]
				} else {
					p = token[:3] + "..."
				}
				slog.WarnContext(c.Request().Context(), "edge auth token do not match",
					"token_preview", p,
					"token_len", len(token),
					"expected_len", len(expected),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, "edge auth token do not match")
			}
			return next(c)
		}
	}
}
