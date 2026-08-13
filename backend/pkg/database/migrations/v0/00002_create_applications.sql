-- +goose Up
create table if not exists public.applications (
  id               uuid                        primary key default uuid_generate_v4(),
  name             varchar(255)                not null,
  webhook_url      varchar(2048)               not null,
  webhook_secret   varchar(255)                not null,
  aws_iam_role_arn varchar(2048)               not null,
  max_retries      int                         not null default 5,
  is_trusted       boolean                     not null default false,
  created_at       timestamp with time zone    not null default current_timestamp,
  updated_at       timestamp with time zone    not null default current_timestamp
);

-- +goose Down
drop table if exists public.applications cascade;
