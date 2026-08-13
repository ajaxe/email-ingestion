-- +goose Up
create table if not exists public.organizations (
  id            uuid                        primary key default uuid_generate_v4(),
  name          varchar(255)                not null,
  owner_user_id uuid                        not null references public.users(id) on delete cascade,
  is_personal   boolean                     not null default true,
  created_at    timestamp with time zone    not null default current_timestamp
);

create unique index if not exists idx_organizations_personal_owner
  on public.organizations(owner_user_id)
  where is_personal = true;

alter table public.applications
  add column if not exists organization_id uuid references public.organizations(id) on delete restrict;

create index if not exists idx_applications_organization_id 
  on public.applications(organization_id);

-- +goose Down
drop index if exists idx_applications_organization_id;
alter table public.applications drop column if exists organization_id;
drop index if exists idx_organizations_personal_owner;
drop table if exists public.organizations cascade;
