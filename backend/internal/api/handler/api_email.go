package handler

import (
	"net/http"

	"github.com/ajaxe/email-ingestion/internal/api/middleware"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// HandleAPIEmailByID handles the GET request to retrieve an ingested email by its ID for a specific application.
// Returns the contents of the ingested email wihtout any attachments.
func HandleAPIEmailByID(svc *service.EmailService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		e := c.Param("email_id")
		emailID, err := uuid.Parse(e)
		if err != nil {
			return apperror.Validation("Invalid email ID", err)
		}

		appID, ok := middleware.ApplicationIDFromContext(ctx)
		if !ok {
			return apperror.Unauthorized("Missing application context")
		}

		r, err := svc.GetEmailByID(ctx, appID, emailID)
		if err != nil {
			return apperror.NotFound("Email not found", err)
		}
		return c.JSON(http.StatusOK, r)
	}
}

// HandleAPIGetAttachmentURL handles the GET request to retrieve a presigned download URL for an email attachment.
func HandleAPIGetAttachmentURL(svc *service.EmailService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		e := c.Param("email_id")
		emailID, err := uuid.Parse(e)
		if err != nil {
			return apperror.Validation("Invalid email ID", err)
		}

		attachmentID := c.Param("attachment_id")
		if attachmentID == "" {
			return apperror.Validation("Invalid attachment ID")
		}

		appID, ok := middleware.ApplicationIDFromContext(ctx)
		if !ok {
			return apperror.Unauthorized("Missing application context")
		}

		r, err := svc.GetAttachmentURL(ctx, appID, emailID, attachmentID)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, r)
	}
}
