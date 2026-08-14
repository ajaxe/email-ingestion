-- drop all tables 
drop table if exists api_keys;
drop table if exists public.webhook_logs;
drop table if exists public.webhook_delivery_jobs;
drop type if exists public.webhook_status;
drop table if exists public.ingested_emails;
drop table if exists public.assigned_emails;
drop table if exists public.user_application_access;
drop table if exists public.applications;
drop table if exists public.organizations;
drop table if exists public.inbound_spool_queue;
drop type if exists public.spool_status;
drop table if exists public.users;