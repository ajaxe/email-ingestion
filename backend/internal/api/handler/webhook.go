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

		secret, err := svc.RegisterWebhook(ctx, appID, req.WebhookURL)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.RegisterWebhookResponse{
			Message:       "Webhook registered and verified successfully",
			WebhookSecret: secret,
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

		limit := int32(50)
		if limitStr := c.QueryParam("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = int32(l)
			}
		}

		offset := int32(0)
		if offsetStr := c.QueryParam("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = int32(o)
			}
		}

		status := c.QueryParam("status")

		jobs, err := svc.ListJobs(ctx, appID, limit, offset, status)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, jobs)
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

		if err := svc.RedeliverJob(ctx, appID, jobID); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.RedeliverWebhookJobResponse{
			Message: "Webhook delivery job re-queued successfully",
			JobID:   jobID.String(),
			Status:  "PENDING",
		})
	}
}

