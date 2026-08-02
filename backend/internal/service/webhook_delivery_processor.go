package service

import (
	"context"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
)

type WebhookDeliveryProcessor struct {
	queries *public.Queries
}

func NewWebhookDeliveryProcessor(queries *public.Queries) *WebhookDeliveryProcessor {
	return &WebhookDeliveryProcessor{
		queries: queries,
	}
}

func (w *WebhookDeliveryProcessor) Process(ctx context.Context, payload *model.WebhookDeliveryPayload) error {
	// implement webhook delivery processing logic here
	return nil
}
