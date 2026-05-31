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
