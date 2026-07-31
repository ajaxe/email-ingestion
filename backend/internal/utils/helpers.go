package utils

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
