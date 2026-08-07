package dto

type CreateAPIKeyRequest struct {
	Name       string `json:"name"`
	KeyPrefix  string `json:"keyPrefix"`
	ExpireDays int64  `json:"expireDays"`
}

type CreateApiKeyResponse struct {
	APIKey string `json:"apiKey"`
}
