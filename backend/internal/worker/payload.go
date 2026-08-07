package worker

type IngestEmailPayload struct {
	SpoolID   string `json:"spool_id"`
	UploadKey string `json:"upload_key"`
}

type WebhookDeliveryPayload struct {
	ApplicationID   string `json:"application_id"`
	IngestedEmailID string `json:"ingested_email_id"`
	JobID           string `json:"job_id,omitempty"`
}
