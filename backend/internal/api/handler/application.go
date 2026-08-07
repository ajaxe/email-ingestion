package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/internal/api/middleware"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func CanAccessApplication(ctx context.Context, appID uuid.UUID) error {
	ua, ok := middleware.UserAccessFromContext(ctx)
	if !ok {
		slog.InfoContext(ctx, "cannot access user application from context")
		return apperror.Forbidden("user cannot access applications")
	}
	if !ua.CanAccessApplication(appID) {
		slog.InfoContext(ctx, "application not present user application list, from context")
		return apperror.Forbidden("user cannot access applications")
	}
	return nil
}

func HandleGetApplicationByID(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		ctx := c.Request().Context()
		ua, ok := middleware.UserAccessFromContext(ctx)

		if !ok {
			return apperror.Forbidden("user cannot access applications")
		}

		app := ua.ApplicationByID(appID)
		if app != nil {
			return apperror.NotFound("application not found")
		}

		return c.JSON(http.StatusOK, app)
	}
}

func HandleGetApplications(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		ua, ok := middleware.UserAccessFromContext(ctx)
		if !ok {
			return apperror.Forbidden("user cannot access applications")
		}

		d := []*dto.ApplicationModelResponse{}

		for _, a := range ua.Applications {
			d = append(d, &dto.ApplicationModelResponse{
				ID:         a.ID,
				Name:       a.Name,
				WebhookURL: a.WebhookUrl,
				MaxRetries: int(a.MaxRetries),
				IsTrusted:  a.IsTrusted,
				CreatedAt:  a.CreatedAt,
				UpdatedAt:  a.UpdatedAt,
			})
		}

		return c.JSON(http.StatusOK, d)
	}
}

func HandleCreateAddress(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		var req dto.CreateAddressRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		assignedEmail, err := svc.CreateAddress(ctx, appID, req.Description)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, assignedEmail)
	}
}

func HandleListEmails(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		limitStr := c.QueryParam("limit")
		offsetStr := c.QueryParam("offset")

		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		offset := 0
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		emails, err := svc.ListEmails(ctx, appID, int32(limit), int32(offset))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, emails)
	}
}

func HandleListAddresses(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		addresses, err := svc.ListAddresses(ctx, appID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, addresses)
	}
}

func HandleToggleAddressStatus(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		addressIDStr := c.Param("address_id")
		addressID, err := uuid.Parse(addressIDStr)
		if err != nil {
			return apperror.Validation("invalid address ID")
		}

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		var req dto.ToggleAddressStatusRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		isActive := strings.ToUpper(req.Status) == "ACTIVE" || req.Status == "true"

		if err := svc.ToggleAddressStatus(ctx, appID, addressID, isActive); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Address status updated successfully"})
	}
}

func HandleGetAttachmentURL(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		emailIDStr := c.Param("email_id")
		_, err = uuid.Parse(emailIDStr)
		if err != nil {
			return apperror.Validation("invalid email ID")
		}

		attachmentIDStr := c.Param("attachment_id")
		if attachmentIDStr == "" {
			return apperror.Validation("invalid attachment ID")
		}

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		// TODO: Generate AWS STS pre-signed S3 download URL (Phase 6.2)
		// 1. Validate JWT / User context and resolve request client to application identity (appID).
		// 2. Query Postgres to fetch the application's unique aws_iam_role_arn.
		// 3. Call AWS STS AssumeRole using Go AWS SDK.
		// 4. Instantiate a scoped S3 client using returned transient STS credentials.
		// 5. Generate a short-lived S3 Presigned URL for the requested attachment object key.

		return c.JSON(http.StatusOK, dto.AttachmentURLResponse{
			AttachmentID: attachmentIDStr,
			DownloadURL:  "https://placeholder-storage.s3.amazonaws.com/attachments/" + attachmentIDStr,
			ExpiresAt:    time.Now().Add(15 * time.Minute),
		})
	}
}

