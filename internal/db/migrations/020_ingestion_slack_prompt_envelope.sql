alter table if exists ingestion
  add column if not exists prompt_envelope jsonb not null default '{}'::jsonb;
