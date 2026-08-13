-- +goose Up
create extension if not exists "uuid-ossp";

create type public.spool_status as enum (
  'PENDING',
  'PROCESSING',
  'SUCCESS',
  'FAILED',
  'DEAD'
);

create type public.webhook_status as enum (
  'PENDING',
  'PROCESSING',
  'SUCCESS',
  'FAILED',
  'DEAD'
);

-- +goose Down
drop type if exists public.webhook_status;
drop type if exists public.spool_status;
drop extension if exists "uuid-ossp";
