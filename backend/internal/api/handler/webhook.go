package handler

import (
	"net/http"
	"strconv"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func HandleRegisterWebhook(svc *service.WebhookService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		ctx := c.Request().Context()
		if err := CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		var req dto.RegisterWebhookRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		if req.WebhookURL == "" {
			return apperror.Validation("webhook_url is required")
		}

		secret, err := svc.RegisterWebhook(ctx, appID, req.WebhookURL, req.MaxRetries)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.RegisterWebhookResponse{
			Message:       "Webhook registered and verified successfully",
			WebhookSecret: secret,
		})
	}
}

func HandlePutUpdateWebhook(svc *service.WebhookService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		ctx := c.Request().Context()
		if err := CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		var req dto.UpdateWebhookRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		if req.WebhookURL == "" {
			return apperror.Validation("webhookUrl is required")
		}

		if req.VerifyOnly {
			if err := svc.Verify(ctx, appID, req.WebhookURL); err != nil {
				return err
			}
			return c.JSON(http.StatusOK, map[string]string{
				"message": "Webhook configuration verified successfully",
			})
		}
		if err := svc.UpdateWebhook(ctx, appID, req.WebhookURL, req.MaxRetries); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Webhook configuration updated successfully",
		})
	}
}

func HandleListWebhookJobs(svc *service.WebhookService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		ctx := c.Request().Context()
		if err := CanAccessApplication(ctx, appID); err != nil {
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

		status := c.QueryParam("status")

		jobs, totalCount, err := svc.ListJobs(ctx, appID, int32(limit), int32(offset), status)
		if err != nil {
			return err
		}

		totalPages := int64(0)
		if limit > 0 {
			totalPages = (totalCount + int64(limit) - 1) / int64(limit)
		}

		return c.JSON(http.StatusOK, dto.PaginatedWebhookJobsResponse{
			Jobs: jobs,
			Pagination: dto.PaginationMeta{
				CurrentPage: int64(page),
				Limit:       int64(limit),
				TotalCount:  totalCount,
				TotalPages:  totalPages,
			},
		})
	}
}

func HandleRedeliverWebhookJob(svc *service.WebhookService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		jobIDStr := c.Param("job_id")
		jobID, err := uuid.Parse(jobIDStr)
		if err != nil {
			return apperror.Validation("invalid job ID")
		}

		ctx := c.Request().Context()
		if err := CanAccessApplication(ctx, appID); err != nil {
			return err
		}

		newJob, err := svc.RedeliverJob(ctx, appID, jobID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.RedeliverWebhookJobResponse{
			Message: "Webhook delivery job re-queued successfully",
			JobID:   newJob.ID.String(),
			Status:  string(newJob.Status),
		})
	}
}
