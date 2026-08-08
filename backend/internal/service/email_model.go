package service

import (
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
)

type EmailContent struct {
	public.GetIngestedEmailByIDRow
	storage.EmailStorageContent
}
