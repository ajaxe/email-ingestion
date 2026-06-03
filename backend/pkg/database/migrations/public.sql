-- enable uuid extension
create extension if not exists "uuid-ossp";

-- webhook delivery transaction outbox enum
create type public.webhook_status as enum (
  'PENDING',
  'PROCESSING',
  'SUCCESS',
  'FAILED',
  'DEAD'
);

-- ==========================================
-- applications (tenants)
-- ==========================================
create table if not exists public.applications (
  id               uuid                        primary key default uuid_generate_v4(),
  name             varchar(255)                not null,
  api_key_hash     varchar(64)                 not null unique, -- sha256 hash of api key
  webhook_url      varchar(2048)               not null,
  webhook_secret   varchar(128)                not null,        -- key used to sign hmac payloads
  aws_iam_role_arn varchar(2048)               not null,        -- dedicated iam role mapped at registration
  max_retries      int                         not null default 5,
  created_at       timestamp with time zone    not null default current_timestamp,
  updated_at       timestamp with time zone    not null default current_timestamp
);

-- ==========================================
-- assigned email addresses
-- ==========================================
create table if not exists public.assigned_emails (
  id             uuid                        primary key default uuid_generate_v4(),
  application_id uuid                        not null,
  local_part     varchar(10)                 not null unique, -- exactly 10 characters
  description    varchar(500),
  is_active      boolean                     not null default true,
  created_at     timestamp with time zone    not null default current_timestamp,

  -- foreign keys & constraints
  foreign key (application_id) references public.applications(id) on delete cascade,
  constraint chk_local_part_len check (char_length(local_part) = 10)
);

create index idx_assigned_emails_lookup 
  on public.assigned_emails(local_part)
  where is_active = true;

-- ==========================================
-- ingested emails metadata
-- ==========================================
create table if not exists public.ingested_emails (
  id                uuid                        primary key default uuid_generate_v4(),
  application_id    uuid                        not null,
  assigned_email_id uuid                        not null,
  reference_token   varchar(53)                 not null default '', -- extracted from local-part + addressing
  from_address      varchar(512)                not null,
  subject           varchar(998)                not null,        -- rfc 2822 max subject length
  message_id        varchar(255)                not null,        -- external message-id header
  s3_key_prefix     varchar(1024)               not null,        -- s3 base path of contents & attachments
  received_at       timestamp with time zone    not null default current_timestamp, -- fixed typo 'imestamp'

  -- foreign keys & constraints
  foreign key (application_id)    references public.applications(id)    on delete cascade,
  foreign key (assigned_email_id) references public.assigned_emails(id) on delete restrict    
);

create index idx_ingested_emails_app_search 
  on public.ingested_emails(application_id, received_at desc);

-- ==========================================
-- webhook delivery jobs
-- ==========================================
create table if not exists public.webhook_delivery_jobs (
  id                uuid                        primary key default uuid_generate_v4(),
  application_id    uuid                        not null,
  ingested_email_id uuid                        not null,
  status            public.webhook_status       not null default 'PENDING',
  retry_count       int                         not null default 0,
  next_delivery_at  timestamp with time zone    not null default current_timestamp,
  created_at        timestamp with time zone    not null default current_timestamp,

  -- foreign keys & constraints
  foreign key (application_id)    references public.applications(id)    on delete cascade,
  foreign key (ingested_email_id) references public.ingested_emails(id) on delete cascade
);

create index idx_webhook_jobs_scheduled 
  on public.webhook_delivery_jobs(status, next_delivery_at)
  where status in ('PENDING', 'PROCESSING');

-- ==========================================
-- webhook invocation history (logs)
-- ==========================================
create table if not exists public.webhook_logs (
  id                      uuid                        primary key default uuid_generate_v4(),
  webhook_delivery_job_id uuid                        not null,
  attempt_number          int                         not null,
  http_status_code        int,
  response_body           text,
  is_retry                boolean                     not null,
  duration_ms             int                         not null,
  executed_at             timestamp with time zone    not null default current_timestamp,

  -- foreign keys & constraints
  foreign key (webhook_delivery_job_id) references public.webhook_delivery_jobs(id) on delete cascade
);