drop index if exists db_read_request_active_scope_target_idx;

create table if not exists external_tool_pause (
  id text primary key,
  idempotency_key text not null unique,
  conversation_id text,
  workflow_id text not null,
  trace_id text,
  operation_id text,
  execution_id text,
  hermes_session_id text not null,
  canonical_tool_name text,
  transport_tool_name text not null,
  tool_call_id text not null,
  args_hash text,
  db_read_request_id text references db_read_request(id) on delete set null,
  sql_sha256 text,
  approval_status text not null,
  tool_outcome text not null,
  resume_status text not null,
  approval_ref text,
  result_ref text,
  expires_at timestamptz,
  pending_assistant_message jsonb not null default '{}'::jsonb,
  transcript_snapshot jsonb not null default '[]'::jsonb,
  resume_payload jsonb not null default '{}'::jsonb,
  error_message text,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create index if not exists external_tool_pause_workflow_idx
  on external_tool_pause (workflow_id, resume_status, tool_outcome, updated_at desc);

create index if not exists external_tool_pause_db_read_idx
  on external_tool_pause (db_read_request_id, created_at desc)
  where db_read_request_id is not null and db_read_request_id <> '';

create index if not exists external_tool_pause_resume_ready_idx
  on external_tool_pause (resume_status, tool_outcome, updated_at)
  where resume_status in ('not_ready', 'failed');
