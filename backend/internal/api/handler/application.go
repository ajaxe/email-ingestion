package handler

import (
	"net/http"
	"strconv"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func HandleGetApplicationByID(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		app, err := svc.ApplicationByID(c.Request().Context(), appID)
		if err != nil {
			// Check if err is pgx.ErrNoRows in reality, but service might just return it
			return apperror.NotFound("application not found", err)
		}

		return c.JSON(http.StatusOK, app)
	}
}

func HandleGetApplications(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		ua, ok := ctx.Value(model.UserAccessContextKey).(*model.UserAccessResult)
		if !ok {
			return apperror.Forbidden("user cannot access applications")
		}

		d := []*model.ApplicationModelResponse{}

		for _, a := range ua.Applications {
			d = append(d, &model.ApplicationModelResponse{
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
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
		}

		var req model.CreateAddressRequest
		if err := c.Bind(&req); err != nil {
			return apperror.Validation("invalid request body", err)
		}

		assignedEmail, err := svc.CreateAddress(c.Request().Context(), appID, req.Description)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, assignedEmail)
	}
}

func HandleListEmails(svc *service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			return apperror.Validation("invalid application ID")
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

		emails, err := svc.ListEmails(c.Request().Context(), appID, int32(limit), int32(offset))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, emails)
	}
}
