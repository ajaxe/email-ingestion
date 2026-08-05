package model

type CreateAPIKeyRequest struct {
	Name       string `json:"name"`
	KeyPrefix  string `json:"keyPrefix"`
	ExpireDays int64  `json:"expireDays"`
}

type CreateApiKeyResponse struct {
	APIKey string `json:"apiKey"`
}

type CreateAddressRequest struct {
	Description string `json:"description"`
}

type RegisterWebhookRequest struct {
	WebhookURL string `json:"webhookUrl"`
}

type RegisterWebhookResponse struct {
	Message       string `json:"message"`
	WebhookSecret string `json:"webhookSecret"`
}
