package handler

import (
	"log/slog"
	"net/http"

	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/labstack/echo/v4"
)

// HandleIngest processes the streaming MIME payload from the edge proxy.
func HandleIngest(emailService *service.EmailIngestionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		err := emailService.Process(ctx, c.Request().Body)

		if err != nil {
			slog.ErrorContext(ctx, "failed to process email ingestion", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "failed",
				"message": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": "email received",
		})
	}
}

func HandleIngestEmailLookup(paramName string, emailService *service.EmailIngestionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		email := c.Param(paramName)

		isValid, err := emailService.LookupAssignedEmail(ctx, email)
		if err != nil {
			slog.ErrorContext(ctx, "failed to lookup email address", "email", email, "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "failed",
				"message": err.Error(),
			})
		}

		slog.InfoContext(ctx, "email address lookup successful", "email", email, "is_valid", isValid)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"email":    email,
			"is_valid": isValid,
		})
	}
}
