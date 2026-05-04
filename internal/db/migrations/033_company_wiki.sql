create table if not exists company_source_document (
  id text primary key,
  source_type text not null,
  source_key text not null,
  source_session_key text not null default '',
  workspace text not null default '',
  environment text not null default '',
  title text not null default '',
  url text not null default '',
  status text not null default 'active' check (status in ('active', 'tombstoned')),
  current_revision_id text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (source_type, source_key)
);

create index if not exists company_source_document_session_idx
  on company_source_document (source_type, source_session_key, updated_at desc);

create table if not exists company_source_revision (
  id text primary key,
  document_id text not null references company_source_document(id) on delete cascade,
  source_revision text not null,
  content_sha256 text not null,
  title text not null default '',
  url text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  observed_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  unique (document_id, source_revision)
);

create index if not exists company_source_revision_document_idx
  on company_source_revision (document_id, created_at desc);

create table if not exists company_source_chunk (
  id text primary key,
  document_id text not null references company_source_document(id) on delete cascade,
  revision_id text not null references company_source_revision(id) on delete cascade,
  chunk_index integer not null,
  chunk_kind text not null default 'text',
  content text not null,
  content_sha256 text not null,
  native_locator text not null default '',
  token_estimate integer not null default 0,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  unique (revision_id, chunk_index)
);

create index if not exists company_source_chunk_document_idx
  on company_source_chunk (document_id, created_at asc, chunk_index asc);

create table if not exists company_source_event (
  id text primary key,
  document_id text not null references company_source_document(id) on delete cascade,
  source_type text not null,
  source_key text not null,
  event_kind text not null,
  native_locator text not null default '',
  payload jsonb not null default '{}'::jsonb,
  observed_at timestamptz not null default now(),
  created_at timestamptz not null default now()
);

create index if not exists company_source_event_document_idx
  on company_source_event (document_id, observed_at desc);

create table if not exists company_source_tombstone (
  id text primary key,
  document_id text not null references company_source_document(id) on delete cascade,
  source_type text not null,
  source_key text not null,
  reason text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create table if not exists company_wiki_compiler_run (
  id text primary key,
  status text not null default 'running' check (status in ('running', 'completed', 'failed')),
  source_document_ids jsonb not null default '[]'::jsonb,
  metadata jsonb not null default '{}'::jsonb,
  last_error text not null default '',
  started_at timestamptz not null default now(),
  completed_at timestamptz
);

create table if not exists company_wiki_page (
  id text primary key,
  slug text not null unique,
  title text not null default '',
  status text not null default 'published' check (status in ('draft', 'published')),
  current_revision_id text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists company_wiki_revision (
  id text primary key,
  page_id text not null references company_wiki_page(id) on delete cascade,
  revision_number integer not null,
  compiler_run_id text not null default '',
  title text not null default '',
  body text not null,
  body_sha256 text not null,
  path text not null,
  source_revision_ids jsonb not null default '[]'::jsonb,
  metadata jsonb not null default '{}'::jsonb,
  published_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  unique (page_id, revision_number),
  unique (page_id, body_sha256)
);

create index if not exists company_wiki_revision_page_idx
  on company_wiki_revision (page_id, revision_number desc);

create table if not exists company_wiki_claim (
  id text primary key,
  wiki_revision_id text not null references company_wiki_revision(id) on delete cascade,
  claim_key text not null,
  claim_text text not null default '',
  confidence numeric not null default 1.0,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists company_wiki_claim_key_idx
  on company_wiki_claim (claim_key, created_at desc);

create table if not exists company_wiki_citation (
  id text primary key,
  wiki_revision_id text not null references company_wiki_revision(id) on delete cascade,
  claim_key text not null default '',
  source_document_id text not null,
  source_revision_id text not null,
  chunk_id text not null,
  native_locator text not null default '',
  quote text not null default '',
  created_at timestamptz not null default now()
);

create index if not exists company_wiki_citation_revision_idx
  on company_wiki_citation (wiki_revision_id);

create table if not exists company_wiki_conflict (
  id text primary key,
  wiki_revision_id text not null references company_wiki_revision(id) on delete cascade,
  claim_key text not null,
  summary text not null default '',
  citation_ids jsonb not null default '[]'::jsonb,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create table if not exists company_wiki_manifest (
  path text primary key,
  wiki_page_id text not null references company_wiki_page(id) on delete cascade,
  wiki_revision_id text not null references company_wiki_revision(id) on delete cascade,
  sha256 text not null,
  compiler_run_id text not null default '',
  generated_at timestamptz not null default now()
);

create table if not exists company_wiki_write_audit (
  id text primary key,
  mode text not null check (mode in ('compiler', 'propose', 'apply')),
  status text not null check (status in ('intent', 'published', 'failed')),
  actor text not null default '',
  reason text not null default '',
  idempotency_key text not null default '',
  page_id text not null default '',
  wiki_revision_id text not null default '',
  slug text not null default '',
  title text not null default '',
  proposed_path text not null default '',
  published_path text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  last_error text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index if not exists company_wiki_write_audit_idempotency_idx
  on company_wiki_write_audit (mode, idempotency_key)
  where idempotency_key <> '';
