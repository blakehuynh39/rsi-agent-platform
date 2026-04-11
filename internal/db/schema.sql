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
  allowed_workflow_kinds jsonb not null default '[]'::jsonb,
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
  allowed_bots jsonb not null default '[]'::jsonb,
  approval_needed boolean not null default false
);

create table if not exists workflow_templates (
  name text primary key,
  kind text not null,
  description text not null,
  steps jsonb not null default '[]'::jsonb
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
  event_id text,
  thread_key text not null,
  thread_ts text,
  workflow_hint text not null,
  intent text,
  bot_role text,
  source text not null,
  channel_id text not null,
  user_id text not null,
  text text not null,
  created_at timestamptz not null default now()
);

create table if not exists workflow (
  id text primary key,
  ingestion_id text,
  trace_id text,
  thread_key text not null,
  kind text not null,
  intent text,
  assigned_bot text not null,
  approval_mode text,
  response_mode text,
  status text not null,
  last_error text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  completed_at timestamptz
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

create table if not exists trace_summary (
  trace_id text primary key,
  ingestion_id text not null,
  workflow_id text not null,
  thread_key text not null,
  workflow_kind text not null,
  status text not null,
  last_verdict text,
  started_at timestamptz not null,
  ended_at timestamptz not null,
  event_count integer not null default 0,
  artifact_count integer not null default 0,
  reasoning_step_count integer not null default 0,
  tool_call_count integer not null default 0,
  slack_action_count integer not null default 0
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

create table if not exists artifact (
  id text primary key,
  trace_id text not null,
  kind text not null,
  content_type text not null,
  url text not null,
  size_bytes bigint not null default 0,
  source text not null
);

create table if not exists human_rating (
  id bigserial primary key,
  trace_id text not null,
  score integer not null,
  verdict text not null,
  labels jsonb not null default '[]'::jsonb,
  notes text,
  reviewer_id text not null,
  created_at timestamptz not null default now()
);

create table if not exists event_envelope (
  id text primary key,
  source text not null,
  source_event_id text not null,
  thread_key text,
  incident_key text,
  dedupe_key text not null,
  severity text not null,
  normalized_problem_statement text not null,
  ownership_hint text,
  raw_payload_ref text,
  workflow_hint text,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create unique index if not exists event_envelope_dedupe_idx on event_envelope (source, dedupe_key);

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
  candidate_key text not null default '',
  source_eval_ids jsonb not null default '[]'::jsonb,
  risk_tier text not null default '',
  proposed_scope text not null default '',
  evidence_artifact_ids jsonb not null default '[]'::jsonb,
  active_slot_consuming boolean not null default false,
  review_deadline timestamptz,
  prior_similar_proposal_ids jsonb not null default '[]'::jsonb,
  new_evidence_since_last_rejection boolean not null default false,
  created_at timestamptz not null default now()
);

create index if not exists proposal_active_idx on proposal (active_slot_consuming, status, created_at desc);

create table if not exists proposal_review (
  id bigserial primary key,
  proposal_id text not null,
  decision text not null,
  rationale text not null,
  reviewer_id text not null,
  failure_class text,
  failure_classes jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now()
);

create table if not exists proposal_memory (
  id text primary key,
  proposal_id text not null,
  candidate_key text not null,
  hypothesis text not null,
  diff_summary text not null,
  review_rationale text not null,
  disposition text not null,
  disposition_reason text,
  failure_class text,
  failure_classes jsonb not null default '[]'::jsonb,
  source_eval_ids jsonb not null default '[]'::jsonb,
  linked_artifact_ids jsonb not null default '[]'::jsonb,
  linked_proposal_ids jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists proposal_memory_candidate_idx on proposal_memory (candidate_key, created_at desc);

create table if not exists eval_suite (
  name text primary key,
  description text not null,
  event_kinds jsonb not null default '[]'::jsonb,
  layers jsonb not null default '[]'::jsonb
);

create table if not exists eval_run (
  id text primary key,
  trace_id text not null,
  event_id text,
  suite_name text not null,
  status text not null,
  trigger text not null,
  overall_score double precision not null default 0,
  overall_verdict text not null default '',
  created_at timestamptz not null default now(),
  completed_at timestamptz
);

create index if not exists eval_run_trace_idx on eval_run (trace_id, created_at desc);

create table if not exists eval_judgment (
  id text primary key,
  eval_run_id text not null,
  layer text not null,
  category text not null,
  score double precision not null,
  passed boolean not null default false,
  rationale text not null default '',
  created_at timestamptz not null default now()
);

create table if not exists improvement_candidate (
  id text primary key,
  candidate_key text not null unique,
  subsystem text not null,
  failure_mode text not null,
  intervention_type text not null,
  status text not null,
  severity text not null,
  recurrence_count integer not null default 0,
  expected_impact double precision not null default 0,
  novelty_score double precision not null default 0,
  confidence_score double precision not null default 0,
  freshness_score double precision not null default 0,
  priority_score double precision not null default 0,
  risk_tier text not null default 'medium',
  hypothesis text not null default '',
  proposed_scope text not null default '',
  latest_trace_id text,
  source_eval_ids jsonb not null default '[]'::jsonb,
  evidence_artifact_ids jsonb not null default '[]'::jsonb,
  prior_similar_proposal_ids jsonb not null default '[]'::jsonb,
  new_evidence_since_last_rejection boolean not null default false,
  last_evaluated_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists improvement_candidate_priority_idx on improvement_candidate (status, priority_score desc, updated_at desc);

create table if not exists repo_change_job (
  id text primary key,
  proposal_id text not null,
  candidate_key text not null,
  status text not null,
  repo text not null,
  base_ref text not null,
  branch_name text not null,
  allowed_path_globs jsonb not null default '[]'::jsonb,
  context_summary text not null default '',
  created_at timestamptz not null default now()
);

create table if not exists pr_attempt (
  id text primary key,
  proposal_id text not null,
  repo text not null,
  branch_name text not null,
  pr_url text,
  status text not null,
  validation_status text not null,
  created_at timestamptz not null default now()
);

create table if not exists post_merge_replay (
  id text primary key,
  proposal_id text not null,
  trace_id text not null,
  baseline_score double precision not null default 0,
  candidate_score double precision not null default 0,
  improved boolean not null default false,
  created_at timestamptz not null default now()
);

create table if not exists cron_lease (
  name text primary key,
  holder text not null,
  expires_at timestamptz not null
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

create table if not exists improvement_settings (
  key text primary key,
  active_proposal_cap integer not null default 2,
  updated_at timestamptz not null default now()
);

create table if not exists work_item (
  id text primary key,
  queue text not null,
  kind text not null,
  status text not null,
  trace_id text,
  workflow_id text,
  ingestion_id text,
  proposal_id text,
  thread_key text,
  intent text,
  repo_scope text,
  requested_by text,
  approval_mode text,
  response_mode text,
  payload jsonb not null default '{}'::jsonb,
  attempts integer not null default 0,
  lease_owner text,
  lease_expires_at timestamptz,
  last_error text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  completed_at timestamptz
);

create index if not exists work_item_queue_status_idx on work_item (queue, status, created_at asc);

create table if not exists reasoning_step (
  id text primary key,
  trace_id text not null,
  workflow_id text,
  step_type text not null,
  summary text not null,
  evidence_refs jsonb not null default '[]'::jsonb,
  alternatives jsonb not null default '[]'::jsonb,
  confidence double precision not null default 0,
  decision text,
  created_at timestamptz not null default now()
);

create index if not exists reasoning_step_trace_idx on reasoning_step (trace_id, created_at asc);

create table if not exists tool_call_record (
  id text primary key,
  trace_id text not null,
  workflow_id text,
  tool_name text not null,
  tool_call_id text not null,
  request jsonb not null default '{}'::jsonb,
  summary text,
  raw_artifact_refs jsonb not null default '[]'::jsonb,
  approval_state text,
  interpretation_summary text,
  status text,
  created_at timestamptz not null default now()
);

create index if not exists tool_call_record_trace_idx on tool_call_record (trace_id, created_at asc);

create table if not exists slack_action_record (
  id text primary key,
  trace_id text not null,
  workflow_id text,
  channel_id text,
  thread_ts text,
  idempotency_key text not null,
  draft_body text,
  final_body text,
  policy_verdict text,
  send_status text,
  artifact_refs jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists slack_action_record_trace_idx on slack_action_record (trace_id, created_at asc);

alter table if exists ingestion add column if not exists event_id text;
alter table if exists ingestion add column if not exists thread_ts text;
alter table if exists ingestion add column if not exists intent text;
alter table if exists ingestion add column if not exists bot_role text;

alter table if exists workflow add column if not exists ingestion_id text;
alter table if exists workflow add column if not exists trace_id text;
alter table if exists workflow add column if not exists intent text;
alter table if exists workflow add column if not exists approval_mode text;
alter table if exists workflow add column if not exists response_mode text;
alter table if exists workflow add column if not exists last_error text;
alter table if exists workflow add column if not exists updated_at timestamptz not null default now();
alter table if exists workflow add column if not exists completed_at timestamptz;

alter table if exists trace_summary add column if not exists reasoning_step_count integer not null default 0;
alter table if exists trace_summary add column if not exists tool_call_count integer not null default 0;
alter table if exists trace_summary add column if not exists slack_action_count integer not null default 0;

alter table if exists proposal_review add column if not exists failure_classes jsonb not null default '[]'::jsonb;
alter table if exists proposal_memory add column if not exists failure_classes jsonb not null default '[]'::jsonb;
