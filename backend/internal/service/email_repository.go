package service

import (
	"context"
	"fmt"

	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxEmailRepository implements EmailRepository using pgx and sqlc
type PgxEmailRepository struct {
	*public.Queries
	pool *pgxpool.Pool
}

func NewPgxEmailRepository(pool *pgxpool.Pool) *PgxEmailRepository {
	return &PgxEmailRepository{
		Queries: public.New(pool),
		pool:    pool,
	}
}

// CreateIngestedEmailAndJobTx executes the creation of the ingested email and the enqueueing
// of the webhook job in a single atomic database transaction, fulfilling the Outbox Pattern.
func (r *PgxEmailRepository) CreateIngestedEmailAndJobTx(ctx context.Context, emailParams public.CreateIngestedEmailParams) (public.IngestedEmail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return public.IngestedEmail{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.Queries.WithTx(tx)

	ingestedEmail, err := qtx.CreateIngestedEmail(ctx, emailParams)
	if err != nil {
		return public.IngestedEmail{}, fmt.Errorf("failed to insert ingested email: %w", err)
	}

	_, err = qtx.EnqueueWebhookJob(ctx, public.EnqueueWebhookJobParams{
		ApplicationID:   emailParams.ApplicationID,
		IngestedEmailID: ingestedEmail.ID,
	})
	if err != nil {
		return public.IngestedEmail{}, fmt.Errorf("failed to enqueue webhook job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return public.IngestedEmail{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ingestedEmail, nil
}
