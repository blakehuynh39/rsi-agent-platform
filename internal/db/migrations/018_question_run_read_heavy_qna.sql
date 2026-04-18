create table if not exists question_run (
  id text primary key,
  workflow_id text not null,
  trace_id text,
  conversation_id text,
  case_id text,
  ingestion_id text,
  role text,
  strategy text,
  status text not null,
  investigation_spec jsonb not null default '{}'::jsonb,
  evidence_ledger jsonb not null default '{}'::jsonb,
  result jsonb not null default '{}'::jsonb,
  failure_class text,
  failure_summary text,
  last_error text,
  runner_diagnostics jsonb not null default '{}'::jsonb,
  version bigint not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  completed_at timestamptz
);

create index if not exists question_run_workflow_idx
  on question_run (workflow_id, updated_at desc);

create index if not exists question_run_trace_idx
  on question_run (trace_id, updated_at desc);
