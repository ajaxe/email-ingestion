package dto

import "time"

type RegisterWebhookRequest struct {
	WebhookURL string `json:"webhookUrl"`
}

type RegisterWebhookResponse struct {
	Message       string `json:"message"`
	WebhookSecret string `json:"webhookSecret"`
}

type WebhookJobSummary struct {
	ID             string    `json:"id"`
	ApplicationID  string    `json:"application_id"`
	Status         string    `json:"status"`
	HTTPStatusCode *int32    `json:"http_status_code,omitempty"`
	DurationMS     *int32    `json:"duration_ms,omitempty"`
	AttemptNumber  *int32    `json:"attempt_number,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
