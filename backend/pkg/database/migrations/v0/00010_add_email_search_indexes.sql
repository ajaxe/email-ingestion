-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_ingested_emails_from_trgm 
  ON public.ingested_emails USING gin (from_address gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_ingested_emails_subject_trgm 
  ON public.ingested_emails USING gin (subject gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_ingested_emails_app_assigned
  ON public.ingested_emails(application_id, assigned_email_id);

-- +goose Down
DROP INDEX IF EXISTS idx_ingested_emails_app_assigned;
DROP INDEX IF EXISTS idx_ingested_emails_subject_trgm;
DROP INDEX IF EXISTS idx_ingested_emails_from_trgm;
