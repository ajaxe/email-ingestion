package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

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

func (s *ApplicationService) GetApplication(ctx context.Context, appID uuid.UUID) (public.Application, error) {
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

func (s *ApplicationService) ListEmails(ctx context.Context, appID uuid.UUID, limit, offset int32) ([]public.IngestedEmail, error) {
	return s.queries.ListIngestedEmailsByApplication(ctx, public.ListIngestedEmailsByApplicationParams{
		ApplicationID: appID,
		Limit:         limit,
		Offset:        offset,
	})
}
