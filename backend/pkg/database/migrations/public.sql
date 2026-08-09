-- enable uuid extension
create extension if not exists "uuid-ossp";

-- webhook delivery transaction outbox enum
/*do $$
begin*/
  create type public.spool_status as enum (
      'PENDING',
      'PROCESSING',
      'SUCCESS',
      'FAILED',
      'DEAD'
    );
/*exception
  when duplicate_object then null;
end $$;*/

/*do $$
begin*/
  create type public.webhook_status as enum (
    'PENDING',
    'PROCESSING',
    'SUCCESS',
    'FAILED',
    'DEAD'
  );
/*exception
  when duplicate_object then null;
end $$;*/

-- ==========================================
-- applications (tenants)
-- ==========================================
create table if not exists public.applications (
  id               uuid                        primary key default uuid_generate_v4(),
  name             varchar(255)                not null,
  webhook_url      varchar(2048)               not null,
  webhook_secret   varchar(255)                not null,        -- key used to sign hmac payloads
  aws_iam_role_arn varchar(2048)               not null,        -- dedicated iam role mapped at registration
  max_retries      int                         not null default 5,
  is_trusted       boolean                     not null default false, -- skips SSRF webhook checks
  created_at       timestamp with time zone    not null default current_timestamp,
  updated_at       timestamp with time zone    not null default current_timestamp
);

-- ==========================================
-- assigned email addresses
-- ==========================================
create table if not exists public.assigned_emails (
  id             uuid                        primary key default uuid_generate_v4(),
  application_id uuid                        not null,
  local_part     varchar(64)                 not null unique, -- exactly 64 characters, app creates 10 character unique email prefix
  description    varchar(500),
  is_active      boolean                     not null default true,
  created_at     timestamp with time zone    not null default current_timestamp,

  -- foreign keys & constraints
  foreign key (application_id) references public.applications(id) on delete cascade
);

create index if not exists idx_assigned_emails_lookup 
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

create index if not exists idx_ingested_emails_app_search 
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

create index if not exists idx_webhook_jobs_app_id_lookup 
  on public.webhook_delivery_jobs(application_id, status);

create index if not exists idx_webhook_jobs_scheduled 
  on public.webhook_delivery_jobs(status, next_delivery_at)
  where status in ('PENDING', 'PROCESSING');

-- ==========================================
-- webhook invocation history (logs)
-- ==========================================
create table if not exists public.webhook_logs (
  id                      uuid                        primary key default uuid_generate_v4(),
  webhook_delivery_job_id uuid                        not null,
  attempt_number          int                         not null,
  http_status_code        int                         not null default 0,
  response_body           text                        not null default '',
  is_retry                boolean                     not null,
  duration_ms             int                         not null,
  executed_at             timestamp with time zone    not null default current_timestamp,

  -- foreign keys & constraints
  foreign key (webhook_delivery_job_id) references public.webhook_delivery_jobs(id) on delete cascade
);

create table if not exists public.inbound_spool_queue (
  id                  uuid                        primary key default uuid_generate_v4(),
  s3_object_key       varchar(1024)               not null,
  status              public.spool_status         not null default 'PENDING',
  attempt_count       int                         not null default 0,
  last_error_message  varchar(1024)               not null default '',
  created_at          timestamp with time zone    not null default current_timestamp,
  updated_at          timestamp with time zone    not null default current_timestamp
);

create table if not exists public.users (
  id             uuid                        primary key default uuid_generate_v4(),
  email          text                        not null default '',
  idp_user_sub   text                        not null default '',
  status         text                        not null default 'pending',
  created_at     timestamp with time zone    not null default current_timestamp,
  activated_at   timestamp with time zone    not null default current_timestamp,
  last_login_at  timestamp with time zone    not null default current_timestamp
);

create index if not exists idx_user_lookup 
  on public.user_application_access(idp_user_sub, email);

create table if not exists public.user_application_access (
  id             uuid                        primary key default uuid_generate_v4(),
  application_id uuid                        not null references public.applications(id) on delete cascade,
  user_id        uuid                        not null references public.users(id) on delete cascade,
  created_at     timestamp with time zone    not null default current_timestamp,
  updated_at     timestamp with time zone    not null default current_timestamp,

  -- foreign keys & constraints
  unique (application_id, user_id)
);

create table if not exists public.api_keys (
  id             uuid                        primary key default uuid_generate_v4(),
  application_id uuid                        not null references public.applications(id) on delete cascade,
  name           varchar(200)                not null,
  key_prefix     text                        not null, -- first 8 characters of api key   
  key_hash       text                        not null unique, -- sha256 hash of api key
  created_at     timestamp with time zone    not null default current_timestamp,
  expires_at     timestamp with time zone    not null default current_timestamp + interval '365 days',
  last_used_at   timestamp with time zone    not null default current_timestamp
);

create index if not exists idx_api_keys_hash on api_keys(key_hash);
create index if not exists idx_api_keys_app_id on api_keys(application_id);
