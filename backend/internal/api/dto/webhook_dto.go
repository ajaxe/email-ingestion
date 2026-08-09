package dto

import "time"

type RegisterWebhookRequest struct {
	WebhookURL string `json:"webhookUrl"`
	MaxRetries int    `json:"maxRetries"`
}

type UpdateWebhookRequest struct {
	RegisterWebhookRequest
	VerifyOnly bool `json:"verifyOnly"`
}

type RegisterWebhookResponse struct {
	Message       string `json:"message"`
	WebhookSecret string `json:"webhookSecret"`
}

type WebhookJobSummary struct {
	ID             string    `json:"id"`
	ApplicationID  string    `json:"applicationId"`
	Status         string    `json:"status"`
	HTTPStatusCode *int32    `json:"httpStatusCode,omitempty"`
	DurationMS     *int32    `json:"durationMs,omitempty"`
	AttemptNumber  *int32    `json:"attemptNumber,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type RedeliverWebhookJobResponse struct {
	Message string `json:"message"`
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
}
