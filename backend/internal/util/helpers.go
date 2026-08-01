package util

import (
	"fmt"
	"strings"
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
func StreamName(appName, env string) string {
	return fmt.Sprintf("%s:%s:stream", appName, env)
}
