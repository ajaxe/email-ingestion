package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/internal/api/handler"
	"github.com/ajaxe/email-ingestion/internal/api/middleware"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type mockAppRepo struct {
	emails       []public.ListIngestedEmailsByApplicationRow
	totalCount   int64
	lastListArg  public.ListIngestedEmailsByApplicationParams
	lastCountArg public.CountIngestedEmailsByApplicationParams
}

func (m *mockAppRepo) GetApplicationByID(ctx context.Context, appID uuid.UUID) (public.Application, error) {
	return public.Application{ID: appID}, nil
}

func (m *mockAppRepo) CreateAssignedEmail(ctx context.Context, args public.CreateAssignedEmailParams) (public.AssignedEmail, error) {
	return public.AssignedEmail{}, nil
}

func (m *mockAppRepo) ListAssignedEmailsByApplication(ctx context.Context, appID uuid.UUID) ([]public.AssignedEmail, error) {
	return nil, nil
}

func (m *mockAppRepo) ListIngestedEmailsByApplication(ctx context.Context, args public.ListIngestedEmailsByApplicationParams) ([]public.ListIngestedEmailsByApplicationRow, error) {
	m.lastListArg = args
	return m.emails, nil
}

func (m *mockAppRepo) CountIngestedEmailsByApplication(ctx context.Context, args public.CountIngestedEmailsByApplicationParams) (int64, error) {
	m.lastCountArg = args
	return m.totalCount, nil
}

func (m *mockAppRepo) UpdateAssignedEmailStatus(ctx context.Context, args public.UpdateAssignedEmailStatusParams) error {
	return nil
}

func (m *mockAppRepo) GetWebhookDeliveryStatsByApplication(ctx context.Context, applicationID uuid.UUID) (public.GetWebhookDeliveryStatsByApplicationRow, error) {
	return public.GetWebhookDeliveryStatsByApplicationRow{}, nil
}

func (m *mockAppRepo) CreateApplication(ctx context.Context, userID uuid.UUID, appName string) (public.Application, error) {
	return public.Application{}, nil
}

func TestHandleListEmails_Pagination(t *testing.T) {
	e := echo.New()
	appID := uuid.New()

	mockRepo := &mockAppRepo{
		emails: []public.ListIngestedEmailsByApplicationRow{
			{ID: uuid.New(), Subject: "Test Email 1"},
			{ID: uuid.New(), Subject: "Test Email 2"},
		},
		totalCount: 45,
	}

	appSvc := service.NewApplicationService(mockRepo)
	h := handler.HandleListEmails(appSvc)

	req := httptest.NewRequest(http.MethodGet, "/applications/"+appID.String()+"/emails?limit=10&page=2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("app_id")
	c.SetParamValues(appID.String())

	userAccess := &dto.UserAccessResult{
		Applications: []public.Application{
			{ID: appID},
		},
	}
	ctx := middleware.WithUserAccess(req.Context(), userAccess)
	c.SetRequest(req.WithContext(ctx))

	if err := h(c); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp dto.PaginatedEmailsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}

	if len(resp.Emails) != 2 {
		t.Errorf("expected 2 emails in payload, got %d", len(resp.Emails))
	}

	if resp.Pagination.CurrentPage != 2 {
		t.Errorf("expected currentPage 2, got %d", resp.Pagination.CurrentPage)
	}

	if resp.Pagination.Limit != 10 {
		t.Errorf("expected limit 10, got %d", resp.Pagination.Limit)
	}

	if resp.Pagination.TotalCount != 45 {
		t.Errorf("expected totalCount 45, got %d", resp.Pagination.TotalCount)
	}

	if resp.Pagination.TotalPages != 5 {
		t.Errorf("expected totalPages 5, got %d", resp.Pagination.TotalPages)
	}
}

func TestHandleListEmails_SearchFilter(t *testing.T) {
	e := echo.New()
	appID := uuid.New()

	mockRepo := &mockAppRepo{
		emails: []public.ListIngestedEmailsByApplicationRow{
			{ID: uuid.New(), Subject: "Invoice #100", FromAddress: "john@example.com"},
		},
		totalCount: 1,
	}

	appSvc := service.NewApplicationService(mockRepo)
	h := handler.HandleListEmails(appSvc)

	req := httptest.NewRequest(http.MethodGet, "/applications/"+appID.String()+"/emails?search=invoice&local_part=sales", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("app_id")
	c.SetParamValues(appID.String())

	userAccess := &dto.UserAccessResult{
		Applications: []public.Application{
			{ID: appID},
		},
	}
	ctx := middleware.WithUserAccess(req.Context(), userAccess)
	c.SetRequest(req.WithContext(ctx))

	if err := h(c); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if mockRepo.lastListArg.Search.String != "invoice" || !mockRepo.lastListArg.Search.Valid {
		t.Errorf("expected Search param 'invoice', got %v", mockRepo.lastListArg.Search)
	}

	if mockRepo.lastListArg.LocalPart.String != "sales" || !mockRepo.lastListArg.LocalPart.Valid {
		t.Errorf("expected LocalPart param 'sales', got %v", mockRepo.lastListArg.LocalPart)
	}

	if mockRepo.lastCountArg.Search.String != "invoice" || !mockRepo.lastCountArg.Search.Valid {
		t.Errorf("expected Count Search param 'invoice', got %v", mockRepo.lastCountArg.Search)
	}

	if mockRepo.lastCountArg.LocalPart.String != "sales" || !mockRepo.lastCountArg.LocalPart.Valid {
		t.Errorf("expected Count LocalPart param 'sales', got %v", mockRepo.lastCountArg.LocalPart)
	}
}
