package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

func HandleCreateApplication(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		req := dto.AppCreateRequest{}
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		d, ok := middleware.UserAccessFromContext(ctx)
		if !ok {
			return apperror.Internal("failed to get user from context")
		}
		uid, err := uuid.Parse(d.UserProfile.UserID)
		if err != nil {
			return apperror.Internal("failed to get user from context", err)
		}

		app, err := svc.CreateApplication(ctx, uid, req.Name)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, dto.ApplicationModelResponse{
			ID:         app.ID,
			Name:       app.Name,
			WebhookURL: app.WebhookUrl,
			MaxRetries: int(app.MaxRetries),
			IsTrusted:  app.IsTrusted,
			CreatedAt:  app.CreatedAt,
			UpdatedAt:  app.UpdatedAt,
		})
	}
}

func HandleGetApplicationByID(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		ctx := c.Request().Context()

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		ua, ok := middleware.UserAccessFromContext(ctx)

		if !ok {
			return apperror.Forbidden("user cannot access applications")
		}

		app := ua.ApplicationByID(appID)
		if app == nil {
			return apperror.NotFound("application not found")
		}

		return c.JSON(http.StatusOK, app)
	}
}

func HandleGetApplicationStats(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		ctx := c.Request().Context()

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		stats, err := svc.GetApplicationStats(ctx, appID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, stats)
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
		pageStr := c.QueryParam("page")

		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		offset := 0
		page := 1
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
			offset = (page - 1) * limit
		}

		search := c.QueryParam("search")
		localPart := c.QueryParam("localPart")

		includeDeletedStr := c.QueryParam("includeDeleted")
		includeDeleted := strings.EqualFold(includeDeletedStr, "true") || includeDeletedStr == "1"

		emails, totalCount, err := svc.ListEmails(ctx, appID, service.ListEmailsFilter{
			Limit:          int32(limit),
			Offset:         int32(offset),
			LocalPart:      localPart,
			Search:         search,
			IncludeDeleted: includeDeleted,
		})
		if err != nil {
			return err
		}

		totalPages := int64(0)
		if limit > 0 {
			totalPages = (totalCount + int64(limit) - 1) / int64(limit)
		}

		return c.JSON(http.StatusOK, dto.PaginatedEmailsResponse{
			Emails: emails,
			Pagination: dto.PaginationMeta{
				CurrentPage: int64(page),
				Limit:       int64(limit),
				TotalCount:  totalCount,
				TotalPages:  totalPages,
			},
		})
	}
}

func HandleGetEmailByID(svc *service.EmailService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		emailIDStr := c.Param("email_id")

		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		emailID, err := uuid.Parse(emailIDStr)
		if err != nil {
			return apperror.Validation("invalid email ID")
		}

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		email, err := svc.GetEmailByID(ctx, appID, emailID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, email)
	}
}

func HandleDeleteEmail(svc *service.EmailService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		emailIDStr := c.Param("email_id")

		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		emailID, err := uuid.Parse(emailIDStr)
		if err != nil {
			return apperror.Validation("invalid email ID")
		}

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		if err := svc.SoftDeleteEmail(ctx, appID, emailID); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Email deleted successfully"})
	}
}

func HandleBulkDeleteEmails(svc *service.EmailService) echo.HandlerFunc {
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

		var req dto.BulkDeleteEmailsRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request payload", err)
		}

		deletedCount, err := svc.BulkSoftDeleteEmails(ctx, appID, req.EmailIDs)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.BulkDeleteEmailsResponse{
			DeletedCount: deletedCount,
		})
	}
}

func HandleGetEmailWebhookHistory(svc *service.EmailService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		emailIDStr := c.Param("email_id")

		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		emailID, err := uuid.Parse(emailIDStr)
		if err != nil {
			return apperror.Validation("invalid email ID")
		}

		if err = CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		history, err := svc.GetEmailWebhookHistory(ctx, appID, emailID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, history)
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

func HandleGetAttachmentURL(svc *service.EmailService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		emailIDStr := c.Param("email_id")
		emailID, err := uuid.Parse(emailIDStr)
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

		res, err := svc.GetAttachmentURL(ctx, appID, emailID, attachmentIDStr)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	}
}
