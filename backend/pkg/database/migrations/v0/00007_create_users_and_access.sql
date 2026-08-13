-- +goose Up
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
  on public.users(idp_user_sub, email);

create table if not exists public.user_application_access (
  id             uuid                        primary key default uuid_generate_v4(),
  application_id uuid                        not null references public.applications(id) on delete cascade,
  user_id        uuid                        not null references public.users(id) on delete cascade,
  created_at     timestamp with time zone    not null default current_timestamp,
  updated_at     timestamp with time zone    not null default current_timestamp,

  unique (application_id, user_id)
);

create table if not exists public.api_keys (
  id             uuid                        primary key default uuid_generate_v4(),
  application_id uuid                        not null references public.applications(id) on delete cascade,
  name           varchar(200)                not null,
  key_prefix     text                        not null,
  key_hash       text                        not null unique,
  created_at     timestamp with time zone    not null default current_timestamp,
  expires_at     timestamp with time zone    not null default current_timestamp + interval '365 days',
  last_used_at   timestamp with time zone    not null default current_timestamp
);

create index if not exists idx_api_keys_hash on public.api_keys(key_hash);
create index if not exists idx_api_keys_app_id on public.api_keys(application_id);

-- +goose Down
drop index if exists idx_api_keys_app_id;
drop index if exists idx_api_keys_hash;
drop table if exists public.api_keys cascade;
drop table if exists public.user_application_access cascade;
drop index if exists idx_user_lookup;
drop table if exists public.users cascade;
