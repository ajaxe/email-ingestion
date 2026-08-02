package model

type IngestEmailPayload struct {
	SpoolID   string `json:"spool_id"`
	UploadKey string `json:"upload_key"`
}
