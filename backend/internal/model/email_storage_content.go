package model

type EmailStorageContent struct {
	Text        string                   `json:"text"`
	HTML        string                   `json:"html"`
	Headers     map[string]string        `json:"headers"`
	Attachments []EmailStorageAttachment `json:"attachments"`
}

type EmailStorageAttachment struct {
	AttachmentID string `json:"attachment_id"`
	FileName     string `json:"file_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	IsInline     bool   `json:"is_inline"`
}
