package apperror

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPErrorBody is the standard JSON error envelope returned to API clients.
type HTTPErrorBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

// codeToStatus maps application error codes to HTTP status codes.
var codeToStatus = map[Code]int{
	CodeValidation:   http.StatusBadRequest,          // 400
	CodeUnauthorized: http.StatusUnauthorized,        // 401
	CodeForbidden:    http.StatusForbidden,           // 403
	CodeNotFound:     http.StatusNotFound,            // 404
	CodeConflict:     http.StatusConflict,            // 409
	CodeInternal:     http.StatusInternalServerError, // 500
}

// NewEchoHTTPErrorHandler returns a centralized error handler for Echo.
// It maps AppError codes to HTTP status codes and falls back to Echo's
// default behavior for echo.HTTPError (e.g., from middleware).
//
// Register with: e.HTTPErrorHandler = apperrors.NewEchoHTTPErrorHandler()
func NewEchoHTTPErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		// 1. Handle AppError from services/domain layer
		var appErr *AppError
		if errors.As(err, &appErr) {
			status := codeToStatus[appErr.Code]
			if status == 0 {
				status = http.StatusInternalServerError
			}

			// Log wrapped root cause if present (never sent to client)
			if appErr.Err != nil {
				slog.ErrorContext(c.Request().Context(),
					"application error",
					slog.String("code", string(appErr.Code)),
					slog.String("message", appErr.Message),
					slog.Any("cause", appErr.Err),
				)
			}

			_ = c.JSON(status, HTTPErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			})
			return
		}

		// 2. Handle echo.HTTPError from Echo middleware (authz, bind, etc.)
		var he *echo.HTTPError
		if errors.As(err, &he) {
			msg := http.StatusText(he.Code)
			if m, ok := he.Message.(string); ok {
				msg = m
			}
			_ = c.JSON(he.Code, HTTPErrorBody{
				Code:    CodeInternal,
				Message: msg,
			})
			return
		}

		// 3. Unknown error — log and return generic 500
		slog.ErrorContext(c.Request().Context(),
			"unhandled error",
			slog.Any("error", err),
		)
		_ = c.JSON(http.StatusInternalServerError, HTTPErrorBody{
			Code:    CodeInternal,
			Message: "internal server error",
		})
	}
}
