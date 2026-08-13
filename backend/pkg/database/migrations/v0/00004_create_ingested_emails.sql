-- +goose Up
create table if not exists public.ingested_emails (
  id                uuid                        primary key default uuid_generate_v4(),
  application_id    uuid                        not null references public.applications(id) on delete cascade,
  assigned_email_id uuid                        not null references public.assigned_emails(id) on delete restrict,
  reference_token   varchar(53)                 not null default '',
  from_address      varchar(512)                not null,
  subject           varchar(998)                not null,
  message_id        varchar(255)                not null,
  s3_key_prefix     varchar(1024)               not null,
  received_at       timestamp with time zone    not null default current_timestamp
);

create index if not exists idx_ingested_emails_app_search 
  on public.ingested_emails(application_id, received_at desc);

-- +goose Down
drop index if exists idx_ingested_emails_app_search;
drop table if exists public.ingested_emails cascade;
