package dto

import (
	"time"

	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type CreateAddressRequest struct {
	Description string `json:"description"`
}

type ApplicationModelResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	WebhookURL string    `json:"webhookUrl"`
	MaxRetries int       `json:"maxRetries"`
	IsTrusted  bool      `json:"isTrusted"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type ToggleAddressStatusRequest struct {
	Status string `json:"status"`
}

type AttachmentURLResponse struct {
	AttachmentID string    `json:"attachmentId"`
	DownloadURL  string    `json:"downloadUrl"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type AppCreateRequest struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
}

type PaginationMeta struct {
	CurrentPage int64 `json:"currentPage"`
	Limit       int64 `json:"limit"`
	TotalCount  int64 `json:"totalCount"`
	TotalPages  int64 `json:"totalPages"`
}

type PaginatedEmailsResponse struct {
	Emails     []public.ListIngestedEmailsByApplicationRow `json:"emails"`
	Pagination PaginationMeta                              `json:"pagination"`
}

type PaginatedWebhookJobsResponse struct {
	Jobs       []public.ListWebhookJobsByApplicationRow `json:"jobs"`
	Pagination PaginationMeta                           `json:"pagination"`
}

type BulkDeleteEmailsRequest struct {
	EmailIDs []uuid.UUID `json:"emailIds"`
}

type BulkDeleteEmailsResponse struct {
	DeletedCount int64 `json:"deletedCount"`
}
