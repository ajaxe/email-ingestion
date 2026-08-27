package service

import (
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
)

type EmailContent struct {
	public.GetIngestedEmailByIDRow
	IsDeleted bool `json:"isDeleted"`
	storage.EmailStorageContent
}
