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
  conversation_id text,
  case_id text,
  thread_key text not null,
  kind text not null,
  intent text,
  assigned_bot text not null,
  approval_mode text,
  response_mode text,
  status text not null,
  last_error text,
  attempt_number integer not null default 0,
  parent_workflow_id text,
  failure_class text,
  failure_summary text,
  retry_decision text,
  retry_after timestamptz,
  runner_diagnostics jsonb not null default '{}'::jsonb,
  repair_attempted boolean not null default false,
  repair_succeeded boolean not null default false,
  version bigint not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  completed_at timestamptz
);

create table if not exists workflow_line (
  case_id text primary key,
  conversation_id text not null,
  status text not null,
  current_workflow_id text,
  latest_workflow_id text,
  attempt_count integer not null default 0,
  auto_retry_budget_remaining integer not null default 0,
  last_failure_class text,
  next_retry_action text,
  retry_after timestamptz,
  line_stop_reason text,
  version bigint not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  completed_at timestamptz
);

create index if not exists workflow_line_conversation_idx on workflow_line (conversation_id, updated_at desc);

create table if not exists runtime_diagnosis (
  id text primary key,
  candidate_key text not null,
  repo text not null,
  conversation_id text,
  case_id text,
  latest_trace_id text,
  status text not null,
  subsystem text,
  failure_mode text,
  summary text,
  evidence_refs jsonb not null default '[]'::jsonb,
  missing_evidence jsonb not null default '[]'::jsonb,
  recommended_fix text,
  target_surface text,
  validation_plan text,
  session_scope_kind text,
  session_scope_id text,
  last_result jsonb not null default '{}'::jsonb,
  last_error text,
  last_attempted_at timestamptz,
  promoted_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists runtime_diagnosis_candidate_idx on runtime_diagnosis (candidate_key, updated_at desc);
create index if not exists runtime_diagnosis_trace_idx on runtime_diagnosis (latest_trace_id, updated_at desc);

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

create table if not exists conversation (
  id text primary key,
  source text not null,
  external_key text not null unique,
  external_conversation text not null,
  title text not null default '',
  status text not null default 'active',
  participant_ids jsonb not null default '[]'::jsonb,
  active_case_id text,
  latest_event_id text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists conversation_updated_idx on conversation (updated_at desc);

create table if not exists conversation_entry (
  id text primary key,
  conversation_id text not null,
  event_id text,
  trace_id text,
  source text not null,
  source_event_id text not null,
  entry_type text not null,
  actor_id text,
  actor_type text,
  body text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists conversation_entry_conv_idx on conversation_entry (conversation_id, created_at asc);

create table if not exists case_record (
  id text primary key,
  conversation_id text not null,
  kind text not null,
  intent text not null,
  title text not null default '',
  summary text not null default '',
  status text not null default 'active',
  approval_mode text,
  response_mode text,
  assigned_bot text not null default '',
  opened_by_event_id text,
  closed_by_event_id text,
  latest_trace_id text,
  resolution_state text not null default 'unresolved',
  resolved_at timestamptz,
  latest_outcome_id text,
  outcome_score double precision not null default 0,
  superseded_by_case_id text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  closed_at timestamptz
);

create index if not exists case_record_conversation_idx on case_record (conversation_id, updated_at desc);

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
  conversation_id text,
  case_id text,
  origin_trace_id text,
  evidence_trace_ids jsonb not null default '[]'::jsonb,
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
  conversation_id text,
  case_id text,
  origin_trace_id text,
  evidence_trace_ids jsonb not null default '[]'::jsonb,
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
  conversation_id text,
  case_id text,
  origin_trace_id text,
  evidence_trace_ids jsonb not null default '[]'::jsonb,
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
  conversation_id text,
  case_id text,
  origin_trace_id text,
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
  conversation_id text,
  case_id text,
  origin_trace_id text,
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
  conversation_id text,
  case_id text,
  baseline_score double precision not null default 0,
  candidate_score double precision not null default 0,
  improved boolean not null default false,
  created_at timestamptz not null default now()
);

create table if not exists action_intent (
  id text primary key,
  owner_plane text not null,
  conversation_id text,
  case_id text,
  trace_id text,
  proposal_id text,
  kind text not null,
  phase_key text,
  target_ref text,
  request_payload jsonb not null default '{}'::jsonb,
  idempotency_key text,
  approval_mode text,
  approval_state text,
  policy_verdict text,
  status text not null,
  superseded_by_action_id text,
  requested_by text,
  rationale text,
  evidence_refs jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

alter table if exists action_intent add column if not exists phase_key text;

create index if not exists action_intent_scope_idx on action_intent (conversation_id, case_id, trace_id, proposal_id, created_at desc);

create table if not exists action_result (
  id text primary key,
  action_intent_id text not null,
  attempt_number integer not null default 1,
  executor text not null,
  provider text,
  provider_ref text,
  request_artifact_id text,
  response_artifact_id text,
  status text not null,
  error_code text,
  error_message text,
  started_at timestamptz not null default now(),
  completed_at timestamptz not null default now()
);

create index if not exists action_result_intent_idx on action_result (action_intent_id, attempt_number asc);

create table if not exists outcome_record (
  id text primary key,
  source text not null,
  source_event_id text,
  conversation_id text,
  case_id text,
  trace_id text,
  proposal_id text,
  outcome_type text not null,
  verdict text not null,
  score double precision not null default 0,
  summary text,
  details text,
  external_ref text,
  recorded_by text,
  recorded_at timestamptz not null default now()
);

create index if not exists outcome_record_scope_idx on outcome_record (conversation_id, case_id, trace_id, proposal_id, recorded_at desc);

create table if not exists knowledge_entry (
  id text primary key,
  tier text not null,
  kind text not null,
  scope_type text not null,
  scope_id text,
  title text not null,
  summary text,
  body text,
  structured_facts jsonb not null default '{}'::jsonb,
  status text not null,
  confidence double precision not null default 0,
  fresh_until timestamptz,
  source_type text not null,
  supersedes_entry_id text,
  contradicted_by_entry_id text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists knowledge_entry_scope_idx on knowledge_entry (scope_type, scope_id, status, updated_at desc);

create table if not exists knowledge_evidence_link (
  id bigserial primary key,
  knowledge_entry_id text not null,
  evidence_type text not null,
  evidence_id text not null,
  relevance_summary text,
  evidence_ref jsonb not null default '{}'::jsonb
);

create index if not exists knowledge_evidence_link_entry_idx on knowledge_evidence_link (knowledge_entry_id);

create table if not exists knowledge_review (
  id text primary key,
  knowledge_entry_id text not null,
  decision text not null,
  reviewer_id text not null,
  rationale text,
  created_at timestamptz not null default now()
);

create index if not exists knowledge_review_entry_idx on knowledge_review (knowledge_entry_id, created_at desc);

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
  conversation_id text,
  case_id text,
  trigger_event_id text,
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
  conversation_id text,
  case_id text,
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
  conversation_id text,
  case_id text,
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
  conversation_id text,
  case_id text,
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

create table if not exists feedback_record (
  id text primary key,
  conversation_id text,
  case_id text,
  trace_id text,
  target_type text not null,
  target_id text not null,
  score integer not null default 0,
  verdict text,
  labels jsonb not null default '[]'::jsonb,
  notes text,
  reviewer_id text not null,
  created_at timestamptz not null default now()
);

create index if not exists feedback_record_trace_idx on feedback_record (trace_id, created_at asc);

alter table if exists ingestion add column if not exists event_id text;
alter table if exists ingestion add column if not exists conversation_id text;
alter table if exists ingestion add column if not exists case_id text;
alter table if exists ingestion add column if not exists thread_ts text;
alter table if exists ingestion add column if not exists intent text;
alter table if exists ingestion add column if not exists bot_role text;

alter table if exists workflow add column if not exists ingestion_id text;
alter table if exists workflow add column if not exists trace_id text;
alter table if exists workflow add column if not exists conversation_id text;
alter table if exists workflow add column if not exists case_id text;
alter table if exists workflow add column if not exists intent text;
alter table if exists workflow add column if not exists approval_mode text;
alter table if exists workflow add column if not exists response_mode text;
alter table if exists workflow add column if not exists last_error text;
alter table if exists workflow add column if not exists updated_at timestamptz not null default now();
alter table if exists workflow add column if not exists completed_at timestamptz;

alter table if exists assignment add column if not exists conversation_id text;
alter table if exists assignment add column if not exists case_id text;

alter table if exists trace_summary add column if not exists conversation_id text;
alter table if exists trace_summary add column if not exists case_id text;
alter table if exists trace_summary add column if not exists trigger_event_id text;
alter table if exists trace_summary add column if not exists supersedes_trace_id text;
alter table if exists trace_summary add column if not exists reasoning_step_count integer not null default 0;
alter table if exists trace_summary add column if not exists tool_call_count integer not null default 0;
alter table if exists trace_summary add column if not exists slack_action_count integer not null default 0;

alter table if exists trace_event add column if not exists conversation_id text;
alter table if exists trace_event add column if not exists case_id text;
alter table if exists trace_event add column if not exists trigger_event_id text;
alter table if exists trace_event add column if not exists description text not null default '';

alter table if exists reasoning_step add column if not exists conversation_id text;
alter table if exists reasoning_step add column if not exists case_id text;

alter table if exists tool_call_record add column if not exists conversation_id text;
alter table if exists tool_call_record add column if not exists case_id text;

alter table if exists slack_action_record add column if not exists conversation_id text;
alter table if exists slack_action_record add column if not exists case_id text;

alter table if exists work_item add column if not exists conversation_id text;
alter table if exists work_item add column if not exists case_id text;
alter table if exists work_item add column if not exists trigger_event_id text;

alter table if exists proposal add column if not exists conversation_id text;
alter table if exists proposal add column if not exists case_id text;
alter table if exists proposal add column if not exists origin_trace_id text;
alter table if exists proposal add column if not exists evidence_trace_ids jsonb not null default '[]'::jsonb;

alter table if exists proposal_review add column if not exists failure_classes jsonb not null default '[]'::jsonb;
alter table if exists proposal_memory add column if not exists conversation_id text;
alter table if exists proposal_memory add column if not exists case_id text;
alter table if exists proposal_memory add column if not exists origin_trace_id text;
alter table if exists proposal_memory add column if not exists evidence_trace_ids jsonb not null default '[]'::jsonb;
alter table if exists proposal_memory add column if not exists failure_classes jsonb not null default '[]'::jsonb;

alter table if exists improvement_candidate add column if not exists conversation_id text;
alter table if exists improvement_candidate add column if not exists case_id text;
alter table if exists improvement_candidate add column if not exists origin_trace_id text;
alter table if exists improvement_candidate add column if not exists evidence_trace_ids jsonb not null default '[]'::jsonb;

alter table if exists repo_change_job add column if not exists conversation_id text;
alter table if exists repo_change_job add column if not exists case_id text;
alter table if exists repo_change_job add column if not exists origin_trace_id text;

alter table if exists pr_attempt add column if not exists conversation_id text;
alter table if exists pr_attempt add column if not exists case_id text;
alter table if exists pr_attempt add column if not exists origin_trace_id text;

alter table if exists post_merge_replay add column if not exists conversation_id text;
alter table if exists post_merge_replay add column if not exists case_id text;


alter table if exists proposal_review add column if not exists idempotency_key text;

update proposal_review
set idempotency_key = proposal_id || ':' || decision
where coalesce(idempotency_key, '') = '';

with canonical_review as (
  select
    min(id) as canonical_id,
    proposal_id,
    decision,
    rationale,
    reviewer_id,
    coalesce(failure_class, '') as failure_class_key,
    failure_classes,
    created_at
  from proposal_review
  group by proposal_id, decision, rationale, reviewer_id, coalesce(failure_class, ''), failure_classes, created_at
),
duplicate_review as (
  select pr.id
  from proposal_review pr
  join canonical_review cr
    on cr.proposal_id = pr.proposal_id
   and cr.decision = pr.decision
   and cr.rationale = pr.rationale
   and cr.reviewer_id = pr.reviewer_id
   and cr.failure_class_key = coalesce(pr.failure_class, '')
   and cr.failure_classes = pr.failure_classes
   and cr.created_at = pr.created_at
  where pr.id <> cr.canonical_id
)
delete from proposal_review pr
using duplicate_review dr
where pr.id = dr.id;

create unique index if not exists proposal_review_proposal_idempotency_idx on proposal_review (proposal_id, idempotency_key);

alter table if exists proposal_memory add column if not exists review_id bigint;

with canonical_review as (
  select
    min(id) as canonical_id,
    proposal_id,
    decision,
    rationale,
    created_at
  from proposal_review
  group by proposal_id, decision, rationale, created_at
)
update proposal_memory pm
set review_id = cr.canonical_id
from canonical_review cr
where pm.review_id is null
  and pm.proposal_id = cr.proposal_id
  and pm.disposition = cr.decision
  and pm.review_rationale = cr.rationale
  and pm.created_at = cr.created_at;

with canonical_memory as (
  select
    min(id) as canonical_id,
    proposal_id,
    coalesce(review_id, 0) as review_id_key,
    candidate_key,
    coalesce(conversation_id, '') as conversation_key,
    coalesce(case_id, '') as case_key,
    coalesce(origin_trace_id, '') as origin_trace_key,
    evidence_trace_ids,
    hypothesis,
    diff_summary,
    review_rationale,
    disposition,
    coalesce(disposition_reason, '') as disposition_reason_key,
    coalesce(failure_class, '') as failure_class_key,
    failure_classes,
    source_eval_ids,
    linked_artifact_ids,
    linked_proposal_ids,
    created_at
  from proposal_memory
  group by
    proposal_id,
    coalesce(review_id, 0),
    candidate_key,
    coalesce(conversation_id, ''),
    coalesce(case_id, ''),
    coalesce(origin_trace_id, ''),
    evidence_trace_ids,
    hypothesis,
    diff_summary,
    review_rationale,
    disposition,
    coalesce(disposition_reason, ''),
    coalesce(failure_class, ''),
    failure_classes,
    source_eval_ids,
    linked_artifact_ids,
    linked_proposal_ids,
    created_at
),
duplicate_memory as (
  select pm.id
  from proposal_memory pm
  join canonical_memory cm
    on cm.proposal_id = pm.proposal_id
   and cm.review_id_key = coalesce(pm.review_id, 0)
   and cm.candidate_key = pm.candidate_key
   and cm.conversation_key = coalesce(pm.conversation_id, '')
   and cm.case_key = coalesce(pm.case_id, '')
   and cm.origin_trace_key = coalesce(pm.origin_trace_id, '')
   and cm.evidence_trace_ids = pm.evidence_trace_ids
   and cm.hypothesis = pm.hypothesis
   and cm.diff_summary = pm.diff_summary
   and cm.review_rationale = pm.review_rationale
   and cm.disposition = pm.disposition
   and cm.disposition_reason_key = coalesce(pm.disposition_reason, '')
   and cm.failure_class_key = coalesce(pm.failure_class, '')
   and cm.failure_classes = pm.failure_classes
   and cm.source_eval_ids = pm.source_eval_ids
   and cm.linked_artifact_ids = pm.linked_artifact_ids
   and cm.linked_proposal_ids = pm.linked_proposal_ids
   and cm.created_at = pm.created_at
  where pm.id <> cm.canonical_id
)
delete from proposal_memory pm
using duplicate_memory dm
where pm.id = dm.id;

create unique index if not exists proposal_memory_review_idx on proposal_memory (review_id) where review_id is not null;

alter table if exists repo_change_job add column if not exists sandbox_namespace text;
alter table if exists repo_change_job add column if not exists sandbox_job_name text;
alter table if exists repo_change_job add column if not exists sandbox_pod_name text;
alter table if exists repo_change_job add column if not exists validation_error text;
alter table if exists repo_change_job add column if not exists validation_ref text;
alter table if exists repo_change_job add column if not exists log_artifact_id text;
alter table if exists repo_change_job add column if not exists updated_at timestamptz not null default now();

update repo_change_job
set updated_at = created_at
where updated_at < created_at;

update proposal
set active_slot_consuming = case
  when status in ('pending_review', 'approved', 'repo_change_queued', 'repo_change_running', 'validation_pending', 'pr_open') then true
  else false
end;


alter table if exists improvement_candidate add column if not exists target_layer text not null default 'repo_change';
alter table if exists improvement_candidate add column if not exists target_kind text not null default 'repo';
alter table if exists improvement_candidate add column if not exists target_ref text not null default '';

alter table if exists proposal add column if not exists target_layer text not null default 'repo_change';
alter table if exists proposal add column if not exists target_kind text not null default 'repo';
alter table if exists proposal add column if not exists target_ref text not null default '';

create table if not exists harness_profile (
  id text primary key,
  role text not null,
  name text not null,
  description text not null default '',
  model text not null default '',
  reasoning_effort text not null default '',
  prompt_fragments jsonb not null default '[]'::jsonb,
  few_shot_snippets jsonb not null default '[]'::jsonb,
  tool_preference_order jsonb not null default '[]'::jsonb,
  retrieval_bias text not null default '',
  reasoning_verbosity text not null default '',
  memory_read_enabled boolean not null default true,
  memory_write_enabled boolean not null default true,
  repo_ref text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists harness_profile_role_idx on harness_profile (role);

create table if not exists harness_overlay (
  id text primary key,
  profile_id text not null references harness_profile(id),
  role text not null,
  version text not null,
  status text not null,
  target_kind text not null default '',
  target_ref text not null default '',
  proposal_id text,
  prompt_fragments jsonb not null default '[]'::jsonb,
  few_shot_snippets jsonb not null default '[]'::jsonb,
  tool_preference_order jsonb not null default '[]'::jsonb,
  retrieval_bias text not null default '',
  reasoning_verbosity text not null default '',
  memory_read_enabled boolean,
  memory_write_enabled boolean,
  created_by text not null default '',
  approved_by text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  activated_at timestamptz
);

create unique index if not exists harness_overlay_role_version_idx on harness_overlay (role, version);
create index if not exists harness_overlay_role_status_idx on harness_overlay (role, status, updated_at desc);

create table if not exists harness_experiment (
  id text primary key,
  profile_id text not null references harness_profile(id),
  overlay_id text references harness_overlay(id),
  proposal_id text,
  role text not null,
  status text not null,
  summary text not null default '',
  metrics jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists harness_experiment_role_idx on harness_experiment (role, updated_at desc);

create table if not exists harness_session_binding (
  role text not null,
  scope_kind text not null,
  scope_id text not null,
  parent_scope_kind text not null default '',
  parent_scope_id text not null default '',
  hermes_session_id text not null,
  parent_session_id text not null default '',
  memory_backend text not null default '',
  assistant_peer_id text not null default '',
  user_peer_id text not null default '',
  harness_profile_id text not null default '',
  effective_overlay_id text not null default '',
  effective_overlay_version text not null default '',
  last_used_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  primary key (role, scope_kind, scope_id)
);

create index if not exists harness_session_binding_last_used_idx on harness_session_binding (last_used_at desc);

create table if not exists harness_execution (
  id text primary key,
  trace_id text,
  proposal_id text,
  role text not null,
  session_scope_kind text not null,
  session_scope_id text not null,
  hermes_session_id text not null,
  parent_session_id text not null default '',
  harness_profile_id text not null default '',
  effective_overlay_id text not null default '',
  effective_overlay_version text not null default '',
  memory_backend text not null default '',
  memory_reads jsonb not null default '[]'::jsonb,
  memory_writes jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists harness_execution_trace_idx on harness_execution (trace_id, created_at desc);
create index if not exists harness_execution_proposal_idx on harness_execution (proposal_id, created_at desc);
create index if not exists harness_execution_role_scope_idx on harness_execution (role, session_scope_kind, session_scope_id, created_at desc);

insert into harness_profile (
  id,
  role,
  name,
  description,
  model,
  reasoning_effort,
  prompt_fragments,
  few_shot_snippets,
  tool_preference_order,
  retrieval_bias,
  reasoning_verbosity,
  memory_read_enabled,
  memory_write_enabled,
  repo_ref,
  created_at,
  updated_at
)
values
  (
    'harness-profile-prod',
    'prod',
    'Production Operator',
    'Live conversation and incident workflow agent with durable memory and explicit evidence-first reasoning.',
    'openai/gpt-5.4',
    'xhigh',
    '["Ground answers in explicit evidence. Prefer concrete repo, Slack, and tool context over generic advice."]'::jsonb,
    '[]'::jsonb,
    '["repo.context","knowledge.context","github.repo_activity","sentry.lookup","kubernetes.logs"]'::jsonb,
    'canonical_then_working_then_session',
    'verbose',
    true,
    true,
    'main',
    now(),
    now()
  ),
  (
    'harness-profile-proactive',
    'proactive',
    'Proactive Thread Agent',
    'Monitors and joins conversations when evidence justifies intervention.',
    'openai/gpt-5.4',
    'xhigh',
    '["Intervene only when the evidence supports a useful reply or workflow launch."]'::jsonb,
    '[]'::jsonb,
    '["knowledge.context","repo.context","github.repo_activity"]'::jsonb,
    'canonical_then_session',
    'verbose',
    true,
    true,
    'main',
    now(),
    now()
  ),
  (
    'harness-profile-eval',
    'eval',
    'Eval Analyst',
    'Summarizes failures, compares traces, and improves recurring eval lines without hiding uncertainty.',
    'openai/gpt-5.4',
    'xhigh',
    '["Focus on observable evidence, failure patterns, and novelty relative to prior rejected proposals."]'::jsonb,
    '[]'::jsonb,
    '["knowledge.context"]'::jsonb,
    'canonical_then_session',
    'verbose',
    true,
    true,
    'main',
    now(),
    now()
  ),
  (
    'harness-profile-proposal',
    'proposal',
    'Proposal Materializer',
    'Turns approved candidate lines into governed repo-change or overlay-ready reasoning with prior memory context.',
    'openai/gpt-5.4',
    'xhigh',
    '["Respect proposal memory, review rationale, and rollback expectations before materializing work."]'::jsonb,
    '[]'::jsonb,
    '["knowledge.context","repo.context"]'::jsonb,
    'canonical_then_working_then_session',
    'verbose',
    true,
    true,
    'main',
    now(),
    now()
  )
on conflict (id) do update set
  role = excluded.role,
  name = excluded.name,
  description = excluded.description,
  model = excluded.model,
  reasoning_effort = excluded.reasoning_effort,
  prompt_fragments = excluded.prompt_fragments,
  few_shot_snippets = excluded.few_shot_snippets,
  tool_preference_order = excluded.tool_preference_order,
  retrieval_bias = excluded.retrieval_bias,
  reasoning_verbosity = excluded.reasoning_verbosity,
  memory_read_enabled = excluded.memory_read_enabled,
  memory_write_enabled = excluded.memory_write_enabled,
  repo_ref = excluded.repo_ref,
  updated_at = excluded.updated_at;


CREATE EXTENSION IF NOT EXISTS vector;

CREATE SCHEMA IF NOT EXISTS honcho;
CREATE TABLE honcho.active_queue_sessions (
    last_updated timestamp with time zone DEFAULT now() NOT NULL,
    id text NOT NULL,
    work_unit_key text NOT NULL
);
CREATE TABLE honcho.alembic_version (
    version_num character varying(32) NOT NULL
);
CREATE TABLE honcho.collections (
    id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    workspace_name text NOT NULL,
    internal_metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    observer text NOT NULL,
    observed text NOT NULL,
    CONSTRAINT ck_collections_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_collections_id_length CHECK ((length(id) = 21)),
    CONSTRAINT ck_collections_public_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_collections_public_id_length CHECK ((length(id) = 21))
);
CREATE TABLE honcho.documents (
    id text NOT NULL,
    internal_metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    content text NOT NULL,
    embedding public.vector(1536),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    workspace_name text NOT NULL,
    session_name text,
    observer text NOT NULL,
    observed text NOT NULL,
    level text DEFAULT 'explicit'::text NOT NULL,
    times_derived integer DEFAULT 1 NOT NULL,
    source_ids jsonb,
    deleted_at timestamp with time zone,
    sync_state text DEFAULT 'pending'::text NOT NULL,
    last_sync_at timestamp with time zone,
    sync_attempts integer DEFAULT 0 NOT NULL,
    CONSTRAINT ck_documents_content_length CHECK ((length(content) <= 65535)),
    CONSTRAINT ck_documents_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_documents_id_length CHECK ((length(id) = 21)),
    CONSTRAINT ck_documents_public_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_documents_public_id_length CHECK ((length(id) = 21))
);
CREATE TABLE honcho.message_embeddings (
    id bigint NOT NULL,
    content text NOT NULL,
    embedding public.vector(1536),
    message_id text NOT NULL,
    workspace_name text NOT NULL,
    session_name text NOT NULL,
    peer_name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    sync_state text DEFAULT 'pending'::text NOT NULL,
    last_sync_at timestamp with time zone,
    sync_attempts integer DEFAULT 0 NOT NULL
);
ALTER TABLE honcho.message_embeddings ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME honcho.message_embeddings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE honcho.messages (
    id bigint NOT NULL,
    public_id text NOT NULL,
    content text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    peer_name text NOT NULL,
    workspace_name text NOT NULL,
    session_name text NOT NULL,
    token_count integer DEFAULT 0 NOT NULL,
    internal_metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    seq_in_session bigint NOT NULL,
    CONSTRAINT ck_messages_content_length CHECK ((length(content) <= 65535)),
    CONSTRAINT ck_messages_public_id_format CHECK ((public_id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_messages_public_id_length CHECK ((length(public_id) = 21))
);
ALTER TABLE honcho.messages ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME honcho.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE honcho.peers (
    id text NOT NULL,
    name text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    configuration jsonb DEFAULT '{}'::jsonb NOT NULL,
    internal_metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    workspace_name text NOT NULL,
    CONSTRAINT ck_peers_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_peers_id_length CHECK ((length(id) = 21)),
    CONSTRAINT ck_users_name_length CHECK ((length(name) <= 512)),
    CONSTRAINT ck_users_public_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_users_public_id_length CHECK ((length(id) = 21))
);
CREATE TABLE honcho.queue (
    id bigint NOT NULL,
    session_id text,
    payload jsonb NOT NULL,
    processed boolean DEFAULT false NOT NULL,
    task_type text NOT NULL,
    work_unit_key text NOT NULL,
    error text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    workspace_name text,
    message_id bigint
);
ALTER TABLE honcho.queue ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME honcho.queue_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE honcho.session_peers (
    workspace_name text NOT NULL,
    session_name text NOT NULL,
    peer_name text NOT NULL,
    configuration jsonb DEFAULT '{}'::jsonb NOT NULL,
    internal_metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    joined_at timestamp with time zone DEFAULT now() NOT NULL,
    left_at timestamp with time zone
);
CREATE TABLE honcho.sessions (
    id text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    configuration jsonb DEFAULT '{}'::jsonb NOT NULL,
    internal_metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    workspace_name text NOT NULL,
    name text NOT NULL,
    CONSTRAINT ck_sessions_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_sessions_id_length CHECK ((length(id) = 21)),
    CONSTRAINT ck_sessions_name_length CHECK ((length(name) <= 512)),
    CONSTRAINT ck_sessions_public_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_sessions_public_id_length CHECK ((length(id) = 21))
);
CREATE TABLE honcho.webhook_endpoints (
    id text NOT NULL,
    workspace_name text NOT NULL,
    url text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_webhook_endpoints_webhook_endpoint_url_length CHECK ((length(url) <= 2048))
);
CREATE TABLE honcho.workspaces (
    id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    configuration jsonb DEFAULT '{}'::jsonb NOT NULL,
    internal_metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT ck_apps_name_length CHECK ((length(name) <= 512)),
    CONSTRAINT ck_apps_public_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_apps_public_id_length CHECK ((length(id) = 21)),
    CONSTRAINT ck_workspaces_id_format CHECK ((id ~ '^[A-Za-z0-9_-]+$'::text)),
    CONSTRAINT ck_workspaces_id_length CHECK ((length(id) = 21))
);
ALTER TABLE ONLY honcho.alembic_version
    ADD CONSTRAINT alembic_version_pkc PRIMARY KEY (version_num);
ALTER TABLE ONLY honcho.active_queue_sessions
    ADD CONSTRAINT pk_active_queue_sessions PRIMARY KEY (id);
ALTER TABLE ONLY honcho.collections
    ADD CONSTRAINT pk_collections PRIMARY KEY (id);
ALTER TABLE ONLY honcho.documents
    ADD CONSTRAINT pk_documents PRIMARY KEY (id);
ALTER TABLE ONLY honcho.message_embeddings
    ADD CONSTRAINT pk_message_embeddings PRIMARY KEY (id);
ALTER TABLE ONLY honcho.messages
    ADD CONSTRAINT pk_messages PRIMARY KEY (id);
ALTER TABLE ONLY honcho.peers
    ADD CONSTRAINT pk_peers PRIMARY KEY (id);
ALTER TABLE ONLY honcho.queue
    ADD CONSTRAINT pk_queue PRIMARY KEY (id);
ALTER TABLE ONLY honcho.session_peers
    ADD CONSTRAINT pk_session_peers PRIMARY KEY (workspace_name, session_name, peer_name);
ALTER TABLE ONLY honcho.sessions
    ADD CONSTRAINT pk_sessions PRIMARY KEY (id);
ALTER TABLE ONLY honcho.webhook_endpoints
    ADD CONSTRAINT pk_webhook_endpoints PRIMARY KEY (id);
ALTER TABLE ONLY honcho.workspaces
    ADD CONSTRAINT pk_workspaces PRIMARY KEY (id);
ALTER TABLE ONLY honcho.active_queue_sessions
    ADD CONSTRAINT uq_active_queue_sessions_work_unit_key UNIQUE (work_unit_key);
ALTER TABLE ONLY honcho.collections
    ADD CONSTRAINT uq_collections_observer_observed_workspace_name UNIQUE (observer, observed, workspace_name);
ALTER TABLE ONLY honcho.messages
    ADD CONSTRAINT uq_messages_public_id UNIQUE (public_id);
ALTER TABLE ONLY honcho.messages
    ADD CONSTRAINT uq_messages_workspace_name_session_name_seq_in_session UNIQUE (workspace_name, session_name, seq_in_session);
ALTER TABLE ONLY honcho.peers
    ADD CONSTRAINT uq_peers_name_workspace_name UNIQUE (name, workspace_name);
ALTER TABLE ONLY honcho.sessions
    ADD CONSTRAINT uq_sessions_name_workspace_name UNIQUE (name, workspace_name);
ALTER TABLE ONLY honcho.workspaces
    ADD CONSTRAINT uq_workspaces_name UNIQUE (name);
CREATE INDEX ix_collections_created_at ON honcho.collections USING btree (created_at);
CREATE INDEX ix_collections_observed ON honcho.collections USING btree (observed);
CREATE INDEX ix_collections_observer ON honcho.collections USING btree (observer);
CREATE INDEX ix_collections_workspace_name ON honcho.collections USING btree (workspace_name);
CREATE INDEX ix_documents_created_at ON honcho.documents USING btree (created_at);
CREATE INDEX ix_documents_deleted_at ON honcho.documents USING btree (deleted_at) WHERE (deleted_at IS NOT NULL);
CREATE INDEX ix_documents_embedding_hnsw ON honcho.documents USING hnsw (embedding public.vector_cosine_ops) WITH (m='16', ef_construction='64');
CREATE INDEX ix_documents_observed ON honcho.documents USING btree (observed);
CREATE INDEX ix_documents_observer ON honcho.documents USING btree (observer);
CREATE INDEX ix_documents_session_name ON honcho.documents USING btree (session_name);
CREATE INDEX ix_documents_source_ids_gin ON honcho.documents USING gin (source_ids);
CREATE INDEX ix_documents_sync_state ON honcho.documents USING btree (sync_state);
CREATE INDEX ix_documents_sync_state_last_sync_at ON honcho.documents USING btree (sync_state, last_sync_at);
CREATE INDEX ix_documents_workspace_name ON honcho.documents USING btree (workspace_name);
CREATE INDEX ix_message_embeddings_created_at ON honcho.message_embeddings USING btree (created_at);
CREATE INDEX ix_message_embeddings_embedding_hnsw ON honcho.message_embeddings USING hnsw (embedding public.vector_cosine_ops) WITH (m='16', ef_construction='64');
CREATE INDEX ix_message_embeddings_message_id ON honcho.message_embeddings USING btree (message_id);
CREATE INDEX ix_message_embeddings_peer_name ON honcho.message_embeddings USING btree (peer_name);
CREATE INDEX ix_message_embeddings_session_name ON honcho.message_embeddings USING btree (session_name);
CREATE INDEX ix_message_embeddings_sync_state ON honcho.message_embeddings USING btree (sync_state);
CREATE INDEX ix_message_embeddings_sync_state_last_sync_at ON honcho.message_embeddings USING btree (sync_state, last_sync_at);
CREATE INDEX ix_message_embeddings_workspace_name ON honcho.message_embeddings USING btree (workspace_name);
CREATE INDEX ix_messages_content_gin ON honcho.messages USING gin (to_tsvector('english'::regconfig, content));
CREATE INDEX ix_messages_created_at ON honcho.messages USING btree (created_at);
CREATE INDEX ix_messages_peer_name ON honcho.messages USING btree (peer_name);
CREATE INDEX ix_messages_session_lookup ON honcho.messages USING btree (session_name, id) INCLUDE (id, created_at);
CREATE INDEX ix_messages_workspace_name ON honcho.messages USING btree (workspace_name);
CREATE INDEX ix_peers_created_at ON honcho.peers USING btree (created_at);
CREATE INDEX ix_peers_workspace_name ON honcho.peers USING btree (workspace_name);
CREATE INDEX ix_queue_created_at ON honcho.queue USING btree (created_at);
CREATE INDEX ix_queue_message_id_not_null ON honcho.queue USING btree (message_id) WHERE (message_id IS NOT NULL);
CREATE INDEX ix_queue_processed ON honcho.queue USING btree (processed);
CREATE INDEX ix_queue_session_id ON honcho.queue USING btree (session_id);
CREATE INDEX ix_queue_work_unit_key_processed_id ON honcho.queue USING btree (work_unit_key, processed, id);
CREATE INDEX ix_queue_workspace_name ON honcho.queue USING btree (workspace_name);
CREATE INDEX ix_sessions_created_at ON honcho.sessions USING btree (created_at);
CREATE INDEX ix_sessions_workspace_name ON honcho.sessions USING btree (workspace_name);
CREATE INDEX ix_webhook_endpoints_workspace_name ON honcho.webhook_endpoints USING btree (workspace_name);
CREATE INDEX ix_workspaces_created_at ON honcho.workspaces USING btree (created_at);
CREATE UNIQUE INDEX uq_queue_dream_pending_work_unit_key ON honcho.queue USING btree (work_unit_key) WHERE ((task_type = 'dream'::text) AND (processed = false));
CREATE UNIQUE INDEX uq_queue_reconciler_pending_work_unit_key ON honcho.queue USING btree (work_unit_key) WHERE ((task_type = 'reconciler'::text) AND (processed = false));
ALTER TABLE ONLY honcho.collections
    ADD CONSTRAINT fk_collections_observed_workspace_name_peers FOREIGN KEY (observed, workspace_name) REFERENCES honcho.peers(name, workspace_name);
ALTER TABLE ONLY honcho.collections
    ADD CONSTRAINT fk_collections_observer_workspace_name_peers FOREIGN KEY (observer, workspace_name) REFERENCES honcho.peers(name, workspace_name);
ALTER TABLE ONLY honcho.collections
    ADD CONSTRAINT fk_collections_workspace_name_workspaces FOREIGN KEY (workspace_name) REFERENCES honcho.workspaces(name);
ALTER TABLE ONLY honcho.documents
    ADD CONSTRAINT fk_documents_observed_workspace_name_peers FOREIGN KEY (observed, workspace_name) REFERENCES honcho.peers(name, workspace_name);
ALTER TABLE ONLY honcho.documents
    ADD CONSTRAINT fk_documents_observer_observed_workspace_name_collections FOREIGN KEY (observer, observed, workspace_name) REFERENCES honcho.collections(observer, observed, workspace_name);
ALTER TABLE ONLY honcho.documents
    ADD CONSTRAINT fk_documents_observer_workspace_name_peers FOREIGN KEY (observer, workspace_name) REFERENCES honcho.peers(name, workspace_name);
ALTER TABLE ONLY honcho.documents
    ADD CONSTRAINT fk_documents_session_workspace FOREIGN KEY (session_name, workspace_name) REFERENCES honcho.sessions(name, workspace_name);
ALTER TABLE ONLY honcho.documents
    ADD CONSTRAINT fk_documents_workspace_name_workspaces FOREIGN KEY (workspace_name) REFERENCES honcho.workspaces(name);
ALTER TABLE ONLY honcho.message_embeddings
    ADD CONSTRAINT fk_message_embeddings_message_id_messages FOREIGN KEY (message_id) REFERENCES honcho.messages(public_id) ON DELETE CASCADE;
ALTER TABLE ONLY honcho.message_embeddings
    ADD CONSTRAINT fk_message_embeddings_peer_name_workspace_name_peers FOREIGN KEY (peer_name, workspace_name) REFERENCES honcho.peers(name, workspace_name);
ALTER TABLE ONLY honcho.message_embeddings
    ADD CONSTRAINT fk_message_embeddings_session_name_workspace_name_sessions FOREIGN KEY (session_name, workspace_name) REFERENCES honcho.sessions(name, workspace_name);
ALTER TABLE ONLY honcho.message_embeddings
    ADD CONSTRAINT fk_message_embeddings_workspace_name_workspaces FOREIGN KEY (workspace_name) REFERENCES honcho.workspaces(name);
ALTER TABLE ONLY honcho.messages
    ADD CONSTRAINT fk_messages_peer_name_peers FOREIGN KEY (peer_name, workspace_name) REFERENCES honcho.peers(name, workspace_name);
ALTER TABLE ONLY honcho.messages
    ADD CONSTRAINT fk_messages_session_name_sessions FOREIGN KEY (session_name, workspace_name) REFERENCES honcho.sessions(name, workspace_name);
ALTER TABLE ONLY honcho.peers
    ADD CONSTRAINT fk_peers_workspace_name_workspaces FOREIGN KEY (workspace_name) REFERENCES honcho.workspaces(name);
ALTER TABLE ONLY honcho.queue
    ADD CONSTRAINT fk_queue_message_id FOREIGN KEY (message_id) REFERENCES honcho.messages(id);
ALTER TABLE ONLY honcho.queue
    ADD CONSTRAINT fk_queue_session_id FOREIGN KEY (session_id) REFERENCES honcho.sessions(id);
ALTER TABLE ONLY honcho.queue
    ADD CONSTRAINT fk_queue_workspace_name FOREIGN KEY (workspace_name) REFERENCES honcho.workspaces(name);
ALTER TABLE ONLY honcho.session_peers
    ADD CONSTRAINT fk_session_peers_peer_name_workspace_name_peers FOREIGN KEY (peer_name, workspace_name) REFERENCES honcho.peers(name, workspace_name);
ALTER TABLE ONLY honcho.session_peers
    ADD CONSTRAINT fk_session_peers_session_name_workspace_name_sessions FOREIGN KEY (session_name, workspace_name) REFERENCES honcho.sessions(name, workspace_name);
ALTER TABLE ONLY honcho.session_peers
    ADD CONSTRAINT fk_session_peers_workspace_name FOREIGN KEY (workspace_name) REFERENCES honcho.workspaces(name);
ALTER TABLE ONLY honcho.sessions
    ADD CONSTRAINT fk_sessions_workspace_name_workspaces FOREIGN KEY (workspace_name) REFERENCES honcho.workspaces(name);
ALTER TABLE ONLY honcho.webhook_endpoints
    ADD CONSTRAINT fk_webhook_endpoints_workspace_name_workspaces FOREIGN KEY (workspace_name) REFERENCES honcho.workspaces(name);


alter table if exists improvement_candidate add column if not exists line_status text not null default 'queued_for_promotion';
alter table if exists improvement_candidate add column if not exists retryable_failure_class text not null default '';
alter table if exists improvement_candidate add column if not exists last_attempt_id text not null default '';
alter table if exists improvement_candidate add column if not exists attempt_count integer not null default 0;
alter table if exists improvement_candidate add column if not exists auto_retry_budget_remaining integer not null default 3;
alter table if exists improvement_candidate add column if not exists current_target_layer text not null default 'repo_change';

update improvement_candidate
set line_status = case
  when status = 'promoted' then 'active_line'
  when status = 'needs_evidence' then 'needs_evidence'
  when status = 'dormant' then 'dormant'
  else 'queued_for_promotion'
end
where coalesce(line_status, '') = ''
   or line_status = 'queued_for_promotion';

update improvement_candidate
set current_target_layer = coalesce(nullif(target_layer, ''), 'repo_change')
where coalesce(current_target_layer, '') = '';

alter table if exists proposal add column if not exists current_attempt_id text not null default '';
alter table if exists proposal add column if not exists attempt_count integer not null default 0;
alter table if exists proposal add column if not exists auto_retry_budget_remaining integer not null default 3;
alter table if exists proposal add column if not exists last_failure_class text not null default '';
alter table if exists proposal add column if not exists next_retry_action text not null default '';
alter table if exists proposal add column if not exists line_stopped_by text not null default '';
alter table if exists proposal add column if not exists line_stop_reason text not null default '';
alter table if exists proposal add column if not exists line_stopped_at timestamptz;

alter table if exists repo_change_job add column if not exists attempt_id text not null default '';
alter table if exists pr_attempt add column if not exists attempt_id text not null default '';
alter table if exists pr_attempt add column if not exists head_sha text not null default '';
alter table if exists action_intent add column if not exists attempt_id text not null default '';
alter table if exists action_result add column if not exists attempt_id text not null default '';
alter table if exists outcome_record add column if not exists attempt_id text not null default '';
alter table if exists harness_experiment add column if not exists attempt_id text not null default '';

create table if not exists change_attempt (
  id text primary key,
  proposal_id text not null,
  candidate_key text not null,
  attempt_number integer not null,
  target_layer text not null,
  target_kind text not null default '',
  target_ref text not null default '',
  trigger text not null,
  state text not null,
  attempt_trace_id text not null default '',
  parent_attempt_id text not null default '',
  branch_name text not null default '',
  pr_url text not null default '',
  head_sha text not null default '',
  failure_class text not null default '',
  failure_summary text not null default '',
  retry_decision text not null default '',
  retry_after timestamptz,
  material_hypothesis_change boolean not null default false,
  diff_summary text not null default '',
  changed_files jsonb not null default '[]'::jsonb,
  validation_summary text not null default '',
  change_plan text not null default '',
  repo_patch text not null default '',
  validation_plan text not null default '',
  retry_assessment text not null default '',
  hypothesis_delta text not null default '',
  overlay_payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index if not exists change_attempt_proposal_attempt_number_idx on change_attempt (proposal_id, attempt_number);
create index if not exists change_attempt_proposal_created_idx on change_attempt (proposal_id, created_at desc);
create index if not exists change_attempt_candidate_created_idx on change_attempt (candidate_key, created_at desc);


create table if not exists operation_execution (
  id text primary key,
  scope_kind text not null,
  scope_id text not null,
  operation_kind text not null,
  operation_key text not null,
  status text not null,
  queue text not null default '',
  requested_by text not null default '',
  holder text not null default '',
  trace_id text not null default '',
  proposal_id text not null default '',
  attempt_id text not null default '',
  payload_hash text not null default '',
  result_ref text not null default '',
  last_error text not null default '',
  retry_count integer not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  started_at timestamptz,
  completed_at timestamptz
);

create unique index if not exists operation_execution_scope_idx
  on operation_execution (scope_kind, scope_id, operation_kind, operation_key);
create index if not exists operation_execution_scope_status_idx
  on operation_execution (scope_kind, scope_id, status, updated_at desc);
create index if not exists operation_execution_proposal_idx
  on operation_execution (proposal_id, updated_at desc);
create index if not exists operation_execution_attempt_idx
  on operation_execution (attempt_id, updated_at desc);

alter table if exists work_item add column if not exists operation_id text not null default '';
alter table if exists action_intent add column if not exists operation_id text not null default '';
alter table if exists action_result add column if not exists operation_id text not null default '';
alter table if exists harness_execution add column if not exists operation_id text not null default '';
alter table if exists outcome_record add column if not exists operation_id text not null default '';

create unique index if not exists work_item_operation_idx
  on work_item (operation_id)
  where operation_id <> '';
create unique index if not exists repo_change_job_attempt_idx
  on repo_change_job (attempt_id)
  where attempt_id <> '';
create unique index if not exists pr_attempt_attempt_idx
  on pr_attempt (attempt_id)
  where attempt_id <> '';
create unique index if not exists action_result_intent_operation_idx
  on action_result (action_intent_id, operation_id)
  where operation_id <> '';
create unique index if not exists harness_execution_operation_idx
  on harness_execution (operation_id)
  where operation_id <> '';
create unique index if not exists outcome_record_operation_idx
  on outcome_record (operation_id)
  where operation_id <> '';


alter table if exists proposal
  add column if not exists recommended_intervention_kind text not null default 'repo_change',
  add column if not exists recommended_intervention_rationale text not null default '',
  add column if not exists target_surface text not null default '',
  add column if not exists touched_files jsonb not null default '[]'::jsonb,
  add column if not exists validation_plan text not null default '',
  add column if not exists material_risk_summary text not null default '',
  add column if not exists recommended_disposition text not null default '';

alter table if exists proposal_review
  add column if not exists scope text not null default 'line';

update proposal
set recommended_intervention_kind = case
      when coalesce(nullif(target_layer, ''), 'repo_change') = 'harness_overlay' then 'harness_overlay'
      else 'repo_change'
    end,
    recommended_intervention_rationale = case
      when coalesce(nullif(summary, ''), '') <> '' then summary
      else 'Intervention recommendation inferred from candidate evidence.'
    end,
    target_surface = case
      when coalesce(nullif(proposed_scope, ''), '') <> '' then proposed_scope
      when coalesce(nullif(target_kind, ''), '') <> '' and coalesce(nullif(target_ref, ''), '') <> '' then target_kind || ':' || target_ref
      when coalesce(nullif(target_ref, ''), '') <> '' then target_ref
      when coalesce(nullif(target_kind, ''), '') <> '' then target_kind
      else 'unspecified_target_surface'
    end,
    validation_plan = case
      when coalesce(nullif(target_layer, ''), 'repo_change') = 'harness_overlay' then 'Generate a bounded runtime overlay, validate behavior in the target role, and activate only if the change remains inside the approved scope.'
      else 'Generate a bounded repo change, validate it in sandbox, and only then open a draft PR.'
    end,
    material_risk_summary = case
      when coalesce(nullif(risk_tier, ''), '') <> '' and coalesce(nullif(target_layer, ''), 'repo_change') = 'harness_overlay' then risk_tier || ' risk intervention on runtime harness surface.'
      when coalesce(nullif(risk_tier, ''), '') <> '' then risk_tier || ' risk repo-change intervention.'
      when coalesce(nullif(target_layer, ''), 'repo_change') = 'harness_overlay' then 'medium risk intervention on runtime harness surface.'
      else 'medium risk repo-change intervention.'
    end,
    recommended_disposition = 'approve_intervention'
where coalesce(recommended_intervention_kind, '') = ''
   or coalesce(target_surface, '') = ''
   or coalesce(validation_plan, '') = ''
   or coalesce(material_risk_summary, '') = ''
   or coalesce(recommended_disposition, '') = '';


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


alter table if exists proposal add column if not exists version bigint not null default 0;
alter table if exists change_attempt add column if not exists version bigint not null default 0;

create table if not exists domain_event (
  id text primary key,
  machine_kind text not null,
  aggregate_id text not null,
  aggregate_version bigint not null,
  event_kind text not null,
  command_id text not null default '',
  causation_id text not null default '',
  payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists domain_event_aggregate_idx
  on domain_event (machine_kind, aggregate_id, aggregate_version desc);
create index if not exists domain_event_kind_idx
  on domain_event (event_kind, created_at desc);

create table if not exists effect_execution (
  id text primary key,
  machine_kind text not null,
  aggregate_id text not null,
  attempt_id text not null default '',
  effect_kind text not null,
  status text not null,
  idempotency_key text not null,
  payload jsonb not null default '{}'::jsonb,
  result_ref text not null default '',
  last_error text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  started_at timestamptz,
  completed_at timestamptz
);

create unique index if not exists effect_execution_idempotency_idx
  on effect_execution (idempotency_key);
create index if not exists effect_execution_attempt_idx
  on effect_execution (attempt_id, updated_at desc);


alter table if exists workflow
  add column if not exists version bigint not null default 0;

alter table if exists effect_execution
  add column if not exists holder text not null default '';
alter table if exists effect_execution
  add column if not exists retry_count integer not null default 0;
alter table if exists effect_execution
  add column if not exists lease_expires_at timestamptz;

create index if not exists effect_execution_aggregate_idx
  on effect_execution (machine_kind, aggregate_id, updated_at desc);
create index if not exists effect_execution_status_idx
  on effect_execution (status, updated_at desc);

create table if not exists command_receipt (
  command_id text primary key,
  machine_kind text not null,
  aggregate_id text not null,
  command_kind text not null,
  causation_id text not null default '',
  actor text not null default '',
  decision_kind text not null,
  reason text not null default '',
  aggregate_version bigint not null default 0,
  result_ref text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists command_receipt_machine_idx
  on command_receipt (machine_kind, aggregate_id, updated_at desc);


drop index if exists work_item_operation_idx;
drop index if exists work_item_queue_status_idx;
drop table if exists operation_execution;
drop table if exists work_item;
