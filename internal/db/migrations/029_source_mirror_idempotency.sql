create table if not exists source_mirror_record (
  source_type text not null,
  source_key text not null,
  workspace text not null default '',
  environment text not null default '',
  source_session_key text not null default '',
  honcho_workspace text not null default '',
  honcho_session_id text not null default '',
  honcho_message_id text not null default '',
  source_revision text not null default '',
  status text not null default 'pending' check (status in ('pending', 'complete', 'failed')),
  metadata jsonb not null default '{}'::jsonb,
  last_error text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  primary key (source_type, source_key)
);

create index if not exists source_mirror_record_session_idx
  on source_mirror_record (source_type, source_session_key, updated_at desc);

create index if not exists source_mirror_record_status_idx
  on source_mirror_record (source_type, status, updated_at desc);

create index if not exists source_mirror_record_honcho_idx
  on source_mirror_record (honcho_workspace, honcho_session_id, updated_at desc);
