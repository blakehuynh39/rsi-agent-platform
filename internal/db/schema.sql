create table if not exists thread_policy (
  thread_key text primary key,
  state text not null,
  owner_bot text not null,
  muted boolean not null default false,
  close_reason text,
  last_policy_version text not null,
  updated_at timestamptz not null default now()
);

create table if not exists channel_policy (
  channel_id text primary key,
  proactive_enabled boolean not null default false,
  auto_post_allowed boolean not null default false,
  allowed_workflow_kinds text[] not null default '{}',
  updated_at timestamptz not null default now()
);

create table if not exists ownership_registry (
  domain text primary key,
  owner_team text not null,
  escalation_slack text not null
);

create table if not exists capability_registry (
  name text primary key,
  kind text not null,
  allowed_bots text[] not null default '{}',
  approval_needed boolean not null default false
);

create table if not exists workflow_templates (
  name text primary key,
  kind text not null,
  description text not null,
  steps text[] not null default '{}'
);

create table if not exists experiment_registry (
  name text primary key,
  candidate text not null,
  baseline text not null,
  state text not null,
  reviewed_by text
);

create table if not exists ingestion (
  id text primary key,
  thread_key text not null,
  workflow_hint text not null,
  source text not null,
  channel_id text not null,
  user_id text not null,
  text text not null,
  created_at timestamptz not null default now()
);

create table if not exists workflow (
  id text primary key,
  thread_key text not null,
  kind text not null,
  assigned_bot text not null,
  status text not null,
  created_at timestamptz not null default now()
);

create table if not exists assignment (
  id text primary key,
  thread_key text not null,
  assigned_bot text not null,
  confidence double precision not null,
  rationale text not null,
  created_at timestamptz not null default now()
);

create table if not exists run (
  id bigserial primary key,
  trace_id text not null,
  workflow_id text not null,
  status text not null,
  created_at timestamptz not null default now()
);

create table if not exists trace_event (
  id bigserial primary key,
  trace_id text not null,
  ingestion_id text not null,
  workflow_id text not null,
  parent_event_id text,
  plane text not null,
  service text not null,
  actor text not null,
  event_type text not null,
  status text not null,
  started_at timestamptz not null,
  ended_at timestamptz,
  payload_ref text,
  artifact_ref text,
  cost_tokens integer default 0,
  latency_ms bigint default 0
);

create table if not exists human_rating (
  id bigserial primary key,
  trace_id text not null,
  score integer not null,
  verdict text not null,
  labels text[] not null default '{}',
  notes text,
  reviewer_id text not null,
  created_at timestamptz not null default now()
);

create table if not exists improvement_note (
  id bigserial primary key,
  trace_id text not null,
  category text not null,
  note text not null,
  suggested_owner text,
  created_by text not null,
  created_at timestamptz not null default now()
);

create table if not exists proposal (
  id text primary key,
  trace_id text not null,
  title text not null,
  category text not null,
  summary text not null,
  status text not null,
  reviewer text,
  created_at timestamptz not null default now()
);

create table if not exists proposal_review (
  id bigserial primary key,
  proposal_id text not null,
  decision text not null,
  rationale text not null,
  reviewer_id text not null,
  created_at timestamptz not null default now()
);

create table if not exists sandbox_session (
  id text primary key,
  trace_id text not null,
  pod_name text not null,
  namespace text not null,
  status text not null,
  created_at timestamptz not null default now(),
  expires_at timestamptz
);

