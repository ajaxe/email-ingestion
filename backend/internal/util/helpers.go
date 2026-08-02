package util

import (
	"fmt"
	"strings"

	"github.com/ajaxe/email-ingestion/pkg/config"
)

func IngestionS3Key(ingestionPrefix, messageID string) string {
	if strings.HasSuffix(ingestionPrefix, "/") {
		return ingestionPrefix + messageID
	}
	return fmt.Sprintf("%s/%s", ingestionPrefix, messageID)
}

func ProcessedEmailS3KeyPrefix(applicationID, messageID string) string {
	return fmt.Sprintf("apps/%s/emails/%s", applicationID, messageID)
}
func ProcessedAttachmentS3KeyPrefix(emailBasePath string) string {
	return fmt.Sprintf("%s/attachments", emailBasePath)
}
func EmailStreamName(env string) string {
	return fmt.Sprintf("%s:%s:email:stream", config.AppName, env)
}

func WebhookStreamName(env string) string {
	return fmt.Sprintf("%s:%s:webhook:stream", config.AppName, env)
}
