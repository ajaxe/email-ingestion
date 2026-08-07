package worker

type IngestEmailPayload struct {
	SpoolID   string `json:"spool_id"`
	UploadKey string `json:"upload_key"`
}

type WebhookDeliveryPayload struct {
	ApplicationID   string
	IngestedEmailID string
}
