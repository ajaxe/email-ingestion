package model

import "github.com/ajaxe/email-ingestion/pkg/database/public"

type EmailContent struct {
	public.IngestedEmail
	EmailStorageContent
}
