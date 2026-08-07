package dto

import (
	"time"

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

