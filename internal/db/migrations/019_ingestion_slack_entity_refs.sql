alter table if exists ingestion
  add column if not exists entity_refs jsonb not null default '[]'::jsonb;
