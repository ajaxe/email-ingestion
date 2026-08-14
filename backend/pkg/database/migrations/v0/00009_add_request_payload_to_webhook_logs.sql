-- +goose Up
alter table public.webhook_logs
  add column if not exists request_payload text not null default '';

-- +goose Down
alter table public.webhook_logs
  drop column if exists request_payload;
