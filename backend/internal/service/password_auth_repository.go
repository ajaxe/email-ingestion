package service

import "github.com/ajaxe/email-ingestion/pkg/database/public"

type AppPasswordAuthRepository struct {
	*public.Queries
	*AuthorizationService
}

func NewAppPasswordAuthRepository(q *public.Queries, authService *AuthorizationService) *AppPasswordAuthRepository {
	return &AppPasswordAuthRepository{
		Queries:              q,
		AuthorizationService: authService,
	}
}
