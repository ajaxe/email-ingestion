package service

import "github.com/ajaxe/email-ingestion/pkg/database/public"

type PgxPasswordAuthRepository struct {
	*public.Queries
	*AuthorizationService
}

func NewAppPasswordAuthRepository(q *public.Queries, authService *AuthorizationService) *PgxPasswordAuthRepository {
	return &PgxPasswordAuthRepository{
		Queries:              q,
		AuthorizationService: authService,
	}
}
