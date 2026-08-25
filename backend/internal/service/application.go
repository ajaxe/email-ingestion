package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ListEmailsFilter struct {
	Limit     int32
	Offset    int32
	LocalPart string
	Search    string
}

type ApplicationRepository interface {
	GetApplicationByID(ctx context.Context, appID uuid.UUID) (public.Application, error)
	CreateAssignedEmail(ctx context.Context, args public.CreateAssignedEmailParams) (public.AssignedEmail, error)
	ListAssignedEmailsByApplication(ctx context.Context, appID uuid.UUID) ([]public.AssignedEmail, error)
	ListIngestedEmailsByApplication(ctx context.Context, args public.ListIngestedEmailsByApplicationParams) ([]public.ListIngestedEmailsByApplicationRow, error)
	CountIngestedEmailsByApplication(ctx context.Context, args public.CountIngestedEmailsByApplicationParams) (int64, error)
	UpdateAssignedEmailStatus(ctx context.Context, args public.UpdateAssignedEmailStatusParams) error
	GetWebhookDeliveryStatsByApplication(ctx context.Context, applicationID uuid.UUID) (public.GetWebhookDeliveryStatsByApplicationRow, error)

	// CreateApplication creates an application for a user, and maps it to the user via user_application_access table.
	CreateApplication(ctx context.Context, userID uuid.UUID, appName string) (public.Application, error)
}

type ApplicationService struct {
	repo ApplicationRepository
}

func NewApplicationService(repo ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, userID uuid.UUID, name string) (public.Application, error) {
	return s.repo.CreateApplication(ctx, userID, name)
}

func (s *ApplicationService) ApplicationByID(ctx context.Context, appID uuid.UUID) (public.Application, error) {
	return s.repo.GetApplicationByID(ctx, appID)
}

func (s *ApplicationService) CreateAddress(ctx context.Context, appID uuid.UUID, description string) (public.AssignedEmail, error) {
	// Generate a 10-character unique email prefix
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return public.AssignedEmail{}, err
	}
	localPart := hex.EncodeToString(bytes) // 10 hex characters

	var desc pgtype.Text
	if description != "" {
		desc = pgtype.Text{String: description, Valid: true}
	}

	return s.repo.CreateAssignedEmail(ctx, public.CreateAssignedEmailParams{
		ApplicationID: appID,
		LocalPart:     localPart,
		Description:   desc,
	})
}

func (s *ApplicationService) ListEmails(ctx context.Context, appID uuid.UUID, filter ListEmailsFilter) ([]public.ListIngestedEmailsByApplicationRow, int64, error) {
	var localPartText pgtype.Text
	if filter.LocalPart != "" && filter.LocalPart != "ALL" {
		localPartText = pgtype.Text{String: filter.LocalPart, Valid: true}
	}

	var searchText pgtype.Text
	if filter.Search != "" {
		searchText = pgtype.Text{String: filter.Search, Valid: true}
	}

	l, err := s.repo.ListIngestedEmailsByApplication(ctx, public.ListIngestedEmailsByApplicationParams{
		ApplicationID: appID,
		Limit:         filter.Limit,
		Offset:        filter.Offset,
		LocalPart:     localPartText,
		Search:        searchText,
	})
	if err != nil {
		return nil, 0, err
	}

	if l == nil {
		l = []public.ListIngestedEmailsByApplicationRow{}
	}

	total, err := s.repo.CountIngestedEmailsByApplication(ctx, public.CountIngestedEmailsByApplicationParams{
		ApplicationID: appID,
		LocalPart:     localPartText,
		Search:        searchText,
	})
	if err != nil {
		return nil, 0, err
	}

	return l, total, nil
}

func (s *ApplicationService) ListAddresses(ctx context.Context, appID uuid.UUID) ([]public.AssignedEmail, error) {
	return s.repo.ListAssignedEmailsByApplication(ctx, appID)
}

func (s *ApplicationService) ToggleAddressStatus(ctx context.Context, appID uuid.UUID, addressID uuid.UUID, isActive bool) error {
	return s.repo.UpdateAssignedEmailStatus(ctx, public.UpdateAssignedEmailStatusParams{
		ID:            addressID,
		IsActive:      isActive,
		ApplicationID: appID,
	})
}

func (s *ApplicationService) GetApplicationStats(ctx context.Context, appID uuid.UUID) (stats ApplicationStats, err error) {

	defer func() {
		if err != nil {
			stats = ApplicationStats{}
			err = apperror.Internal("failed to get application statistics", err)
			return
		}
	}()

	emailCount, err := s.repo.CountIngestedEmailsByApplication(ctx, public.CountIngestedEmailsByApplicationParams{
		ApplicationID: appID,
	})
	if err != nil {
		return
	}
	addr, err := s.repo.ListAssignedEmailsByApplication(ctx, appID)
	if err != nil {
		return
	}
	activeAddr := 0
	for _, a := range addr {
		if a.IsActive {
			activeAddr++
		}
	}

	webhookStats, err := s.repo.GetWebhookDeliveryStatsByApplication(ctx, appID)
	if err != nil {
		return
	}

	var r float32
	if webhookStats.Total != 0 {
		r = float32(webhookStats.Success) / float32(webhookStats.Total)
	}

	stats = ApplicationStats{
		TotalEmails:         emailCount,
		TotalAddresses:      int64(len(addr)),
		ActiveAddresses:     int64(activeAddr),
		WebhookSuccessRate:  r,
		FailWebhookJobCount: webhookStats.Failures,
	}

	return
}
