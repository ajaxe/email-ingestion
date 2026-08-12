package service

import "github.com/ajaxe/email-ingestion/pkg/database/public"

type PgxOIDCAuthRepository struct {
	*public.Queries
	*AuthorizationService
}

func NewPgxOIDCAuthRepository(q *public.Queries, authService *AuthorizationService) *PgxOIDCAuthRepository {
	return &PgxOIDCAuthRepository{
		Queries:              q,
		AuthorizationService: authService,
	}
}
