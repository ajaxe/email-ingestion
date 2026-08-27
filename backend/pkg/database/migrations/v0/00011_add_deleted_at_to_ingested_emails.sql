-- +goose Up
ALTER TABLE public.ingested_emails
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_ingested_emails_app_deleted 
  ON public.ingested_emails(application_id, deleted_at);

-- +goose Down
DROP INDEX IF EXISTS public.idx_ingested_emails_app_deleted;
ALTER TABLE public.ingested_emails
  DROP COLUMN IF EXISTS deleted_at;
