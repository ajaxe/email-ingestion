package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ajaxe/email-ingestion/internal/ingest"
)

// EdgeAuth ensures the request has the correct edge token
func EdgeAuth(expectedToken string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get(ingest.HeaderEdgeToken)
			if token == "" || token != expectedToken {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or missing edge auth token")
			}
			return next(c)
		}
	}
}
