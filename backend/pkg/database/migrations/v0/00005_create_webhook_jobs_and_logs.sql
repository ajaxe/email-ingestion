-- +goose Up
create table if not exists public.webhook_delivery_jobs (
  id                uuid                        primary key default uuid_generate_v4(),
  application_id    uuid                        not null references public.applications(id) on delete cascade,
  ingested_email_id uuid                        not null references public.ingested_emails(id) on delete cascade,
  status            public.webhook_status       not null default 'PENDING',
  retry_count       int                         not null default 0,
  next_delivery_at  timestamp with time zone    not null default current_timestamp,
  created_at        timestamp with time zone    not null default current_timestamp
);

create index if not exists idx_webhook_jobs_app_id_lookup 
  on public.webhook_delivery_jobs(application_id, status);

create index if not exists idx_webhook_jobs_scheduled 
  on public.webhook_delivery_jobs(status, next_delivery_at)
  where status in ('PENDING', 'PROCESSING');

create table if not exists public.webhook_logs (
  id                      uuid                        primary key default uuid_generate_v4(),
  webhook_delivery_job_id uuid                        not null references public.webhook_delivery_jobs(id) on delete cascade,
  attempt_number          int                         not null,
  http_status_code        int                         not null default 0,
  response_body           text                        not null default '',
  is_retry                boolean                     not null,
  duration_ms             int                         not null,
  executed_at             timestamp with time zone    not null default current_timestamp
);

-- +goose Down
drop table if exists public.webhook_logs cascade;
drop index if exists idx_webhook_jobs_scheduled;
drop index if exists idx_webhook_jobs_app_id_lookup;
drop table if exists public.webhook_delivery_jobs cascade;
