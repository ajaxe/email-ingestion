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

type ApplicationService struct {
	queries *public.Queries
}

func NewApplicationService(queries *public.Queries) *ApplicationService {
	return &ApplicationService{queries: queries}
}

func (s *ApplicationService) ApplicationByID(ctx context.Context, appID uuid.UUID) (public.Application, error) {
	return s.queries.GetApplicationByID(ctx, appID)
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

	return s.queries.CreateAssignedEmail(ctx, public.CreateAssignedEmailParams{
		ApplicationID: appID,
		LocalPart:     localPart,
		Description:   desc,
	})
}

func (s *ApplicationService) ListEmails(ctx context.Context, appID uuid.UUID, limit, offset int32) ([]public.ListIngestedEmailsByApplicationRow, error) {
	return s.queries.ListIngestedEmailsByApplication(ctx, public.ListIngestedEmailsByApplicationParams{
		ApplicationID: appID,
		Limit:         limit,
		Offset:        offset,
	})
}

func (s *ApplicationService) ListAddresses(ctx context.Context, appID uuid.UUID) ([]public.AssignedEmail, error) {
	return s.queries.ListAssignedEmailsByApplication(ctx, appID)
}

func (s *ApplicationService) ToggleAddressStatus(ctx context.Context, appID uuid.UUID, addressID uuid.UUID, isActive bool) error {
	return s.queries.UpdateAssignedEmailStatus(ctx, public.UpdateAssignedEmailStatusParams{
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

	emailCount, err := s.queries.CountIngestedEmailsByApplication(ctx, appID)
	if err != nil {
		return
	}
	addr, err := s.queries.ListAssignedEmailsByApplication(ctx, appID)
	if err != nil {
		return
	}
	activeAddr := 0
	for _, a := range addr {
		if a.IsActive {
			activeAddr++
		}
	}

	webhookStats, err := s.queries.GetWebhookDeliveryStatsByApplication(ctx, appID)
	if err != nil {
		return
	}

	stats = ApplicationStats{
		TotalEmails:         emailCount,
		TotalAddresses:      int64(len(addr)),
		ActiveAddresses:     int64(activeAddr),
		WebhookSuccessRate:  float32(webhookStats.Success) / float32(webhookStats.Total),
		FailWebhookJobCount: webhookStats.Failures,
	}

	return
}
