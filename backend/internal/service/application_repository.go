package service

import (
	"context"
	"fmt"

	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxApplicationRepository struct {
	*public.Queries
	pool *pgxpool.Pool
}

func NewPgxApplicationRepository(pool *pgxpool.Pool) *PgxApplicationRepository {
	return &PgxApplicationRepository{
		Queries: public.New(pool),
		pool:    pool,
	}
}

func (a *PgxApplicationRepository) CreateApplication(ctx context.Context, userID uuid.UUID, appName string) (public.Application, error) {
	if appName == "" {
		return public.Application{}, fmt.Errorf("empty application name")
	}
	if userID == uuid.Nil {
		return public.Application{}, fmt.Errorf("invalid user id")
	}
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return public.Application{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := a.Queries.WithTx(tx)

	org, err := qtx.CreatePersonalOrganization(ctx, public.CreatePersonalOrganizationParams{
		Name:        fmt.Sprintf("%s Org", appName),
		OwnerUserID: userID,
	})
	if err != nil {
		return public.Application{}, fmt.Errorf("failed to create personal organization: %w", err)
	}

	app, err := qtx.InsertApplication(ctx, public.InsertApplicationParams{
		Name: appName,
		OrganizationID: pgtype.UUID{
			Bytes: org.ID,
			Valid: true,
		},
	})
	if err != nil {
		return public.Application{}, fmt.Errorf("failed to insert application: %w", err)
	}

	err = qtx.InsertUserApplication(ctx, public.InsertUserApplicationParams{
		ApplicationID: app.ID,
		UserID:        userID,
	})
	if err != nil {
		return public.Application{}, fmt.Errorf("failed to insert user access application: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return public.Application{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return app, nil
}
