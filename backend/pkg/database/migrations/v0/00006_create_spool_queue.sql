-- +goose Up
create table if not exists public.inbound_spool_queue (
  id                  uuid                        primary key default uuid_generate_v4(),
  s3_object_key       varchar(1024)               not null,
  status              public.spool_status         not null default 'PENDING',
  attempt_count       int                         not null default 0,
  last_error_message  varchar(1024)               not null default '',
  created_at          timestamp with time zone    not null default current_timestamp,
  updated_at          timestamp with time zone    not null default current_timestamp
);

-- +goose Down
drop table if exists public.inbound_spool_queue cascade;
