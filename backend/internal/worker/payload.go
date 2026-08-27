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

const JobTypePurgeEmailStorage = "PURGE_EMAIL_STORAGE"

type JobPayload struct {
	Type          string   `json:"type"`
	ApplicationID string   `json:"application_id,omitempty"`
	S3KeyPrefixes []string `json:"s3_key_prefixes,omitempty"`
	RetryCount    int      `json:"retry_count"`
}
