package service

type ApplicationStats struct {
	TotalEmails         int64   `json:"totalEmails"`
	TotalAddresses      int64   `json:"totalAddresses"`
	ActiveAddresses     int64   `json:"activeAddresses"`
	WebhookSuccessRate  float32 `json:"webhookSuccessRate"`
	FailWebhookJobCount int64   `json:"failWebhookJobCount"`
}
