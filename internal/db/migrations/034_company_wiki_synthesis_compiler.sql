alter table if exists company_wiki_manifest
  add column if not exists repair_status text not null default 'ok'
    check (repair_status in ('ok', 'repair_needed', 'repair_failed', 'not_generated'));

alter table if exists company_wiki_manifest
  add column if not exists last_repair_error text not null default '';

alter table if exists company_wiki_manifest
  add column if not exists last_checked_at timestamptz;

alter table if exists company_wiki_manifest
  add column if not exists last_repaired_at timestamptz;

create table if not exists company_wiki_compile_item (
  id text primary key,
  source_revision_id text not null references company_source_revision(id) on delete cascade,
  compiler_version text not null,
  schema_version text not null,
  renderer_version text not null,
  model_policy_version text not null,
  input_hash text not null default '',
  status text not null default 'pending'
    check (status in ('pending', 'claimed', 'completed', 'failed', 'skipped')),
  lease_holder text not null default '',
  lease_expires_at timestamptz,
  attempt_count integer not null default 0,
  last_attempt_id text not null default '',
  last_error text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (source_revision_id, compiler_version, schema_version, renderer_version, model_policy_version)
);

create index if not exists company_wiki_compile_item_status_idx
  on company_wiki_compile_item (status, lease_expires_at, updated_at);

create table if not exists company_wiki_compile_attempt (
  id text primary key,
  compile_item_id text not null references company_wiki_compile_item(id) on delete cascade,
  compiler_run_id text not null default '',
  status text not null default 'claimed'
    check (status in ('pending', 'claimed', 'completed', 'failed', 'skipped')),
  model text not null default '',
  context_hash text not null default '',
  output_hash text not null default '',
  request_metadata_hash text not null default '',
  response_metadata_hash text not null default '',
  duration_millis bigint not null default 0,
  validation_errors jsonb not null default '[]'::jsonb,
  last_error text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  completed_at timestamptz
);

create index if not exists company_wiki_compile_attempt_item_idx
  on company_wiki_compile_attempt (compile_item_id, created_at desc);

create table if not exists company_wiki_compile_item_target (
  id text primary key,
  compile_item_id text not null references company_wiki_compile_item(id) on delete cascade,
  target_slug text not null,
  target_path text not null default '',
  target_type text not null default '',
  status text not null default 'pending'
    check (status in ('pending', 'published', 'failed', 'skipped', 'superseded')),
  wiki_revision_id text not null default '',
  idempotency_key text not null default '',
  body_hash text not null default '',
  last_error text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (compile_item_id, target_slug)
);

create index if not exists company_wiki_compile_item_target_item_idx
  on company_wiki_compile_item_target (compile_item_id, status, target_slug);

create table if not exists company_wiki_conflict_citation (
  conflict_id text not null references company_wiki_conflict(id) on delete cascade,
  citation_id text not null references company_wiki_citation(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (conflict_id, citation_id)
);

create index if not exists company_wiki_page_metadata_gin_idx
  on company_wiki_page using gin (metadata);

create index if not exists company_wiki_revision_metadata_gin_idx
  on company_wiki_revision using gin (metadata);
