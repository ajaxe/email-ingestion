package middleware

import (
	"context"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/google/uuid"
)

type contextKey string

const (
	ApplicationIDContextKey contextKey = "application_id"
	UserAccessContextKey    contextKey = "user_access"
)

// WithApplicationID injects application_id into context.
func WithApplicationID(ctx context.Context, appID uuid.UUID) context.Context {
	return context.WithValue(ctx, ApplicationIDContextKey, appID)
}

// ApplicationIDFromContext retrieves application_id from context.
func ApplicationIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val, ok := ctx.Value(ApplicationIDContextKey).(uuid.UUID)
	return val, ok
}

// WithUserAccess injects UserAccessResult into context.
func WithUserAccess(ctx context.Context, ua *dto.UserAccessResult) context.Context {
	return context.WithValue(ctx, UserAccessContextKey, ua)
}

// UserAccessFromContext retrieves UserAccessResult from context.
func UserAccessFromContext(ctx context.Context) (*dto.UserAccessResult, bool) {
	val, ok := ctx.Value(UserAccessContextKey).(*dto.UserAccessResult)
	return val, ok
}
