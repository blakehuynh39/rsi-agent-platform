create table if not exists attempt_workspace (
  id text primary key,
  attempt_id text not null unique,
  proposal_id text not null,
  repo text not null,
  base_ref text not null default 'main',
  branch_name text not null,
  namespace text,
  job_name text,
  pod_name text,
  status text not null default 'queued',
  allowed_path_globs jsonb not null default '[]'::jsonb,
  head_sha text,
  diff_summary text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  expires_at timestamptz
);

create index if not exists attempt_workspace_proposal_idx on attempt_workspace (proposal_id, created_at desc);
