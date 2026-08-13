-- +goose Up
create table if not exists public.assigned_emails (
  id             uuid                        primary key default uuid_generate_v4(),
  application_id uuid                        not null references public.applications(id) on delete cascade,
  local_part     varchar(64)                 not null unique,
  description    varchar(500),
  is_active      boolean                     not null default true,
  created_at     timestamp with time zone    not null default current_timestamp
);

create index if not exists idx_assigned_emails_lookup 
  on public.assigned_emails(local_part)
  where is_active = true;

-- +goose Down
drop index if exists idx_assigned_emails_lookup;
drop table if exists public.assigned_emails cascade;
