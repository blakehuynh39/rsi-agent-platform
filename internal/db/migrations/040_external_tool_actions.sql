create table if not exists external_tool_action (
  id text primary key,
  surface text not null,
  operation text not null,
  target_ref text,
  idempotency_key text not null,
  request_hash text not null,
  state text not null,
  actor text not null,
  reason text,
  destructive boolean not null default false,
  execution_id text,
  operation_id text,
  trace_id text,
  workflow_id text,
  conversation_id text,
  response_summary text,
  error_message text,
  source_ref text,
  wiki_audit_id text,
  mirror_effect jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  completed_at timestamptz
);

create unique index if not exists external_tool_action_idempotency_idx
  on external_tool_action (surface, operation, idempotency_key);

create index if not exists external_tool_action_execution_idx
  on external_tool_action (execution_id, created_at desc)
  where execution_id is not null and execution_id <> '';

create index if not exists external_tool_action_trace_idx
  on external_tool_action (trace_id, created_at desc)
  where trace_id is not null and trace_id <> '';

create index if not exists external_tool_action_state_idx
  on external_tool_action (surface, state, updated_at desc);
