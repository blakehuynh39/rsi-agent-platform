create table if not exists db_read_request (
  id text primary key,
  idempotency_key text not null unique,
  target text not null,
  purpose text not null default 'query',
  sql_text text not null,
  sql_sha256 text not null,
  execution_scope_key text,
  requester text not null,
  conversation_id text,
  workflow_id text,
  trace_id text,
  channel_id text,
  thread_ts text,
  state text not null,
  current_validation_attempt_id text,
  approved_by_slack_user_id text,
  approved_at timestamptz,
  expires_at timestamptz not null,
  caps jsonb not null default '{}'::jsonb,
  redaction jsonb not null default '{}'::jsonb,
  slack_message_channel_id text,
  slack_message_ts text,
  lease_holder text,
  lease_token text,
  lease_generation integer not null default 0,
  lease_expires_at timestamptz,
  result_artifact_ref text,
  result_sample jsonb not null default '[]'::jsonb,
  row_count integer not null default 0,
  truncated boolean not null default false,
  error_message text,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create index if not exists db_read_request_state_idx
  on db_read_request (state, expires_at, created_at);

create index if not exists db_read_request_thread_idx
  on db_read_request (channel_id, thread_ts, created_at desc);

create unique index if not exists db_read_request_active_scope_target_idx
  on db_read_request (target, execution_scope_key)
  where execution_scope_key is not null
    and execution_scope_key <> ''
    and state in ('validating', 'pending_approval', 'approved', 'executing');

create table if not exists db_read_validation_attempt (
  id text primary key,
  request_id text not null references db_read_request(id) on delete cascade,
  target text not null,
  sql_sha256 text not null,
  status text not null,
  stage text not null,
  error_code text,
  error_message text,
  details jsonb not null default '{}'::jsonb,
  validated_at timestamptz,
  created_at timestamptz not null
);

create index if not exists db_read_validation_attempt_request_idx
  on db_read_validation_attempt (request_id, created_at asc);

create table if not exists db_read_execution_result (
  id text primary key,
  request_id text not null references db_read_request(id) on delete cascade,
  lease_token text,
  status text not null,
  row_count integer not null default 0,
  truncated boolean not null default false,
  sample jsonb not null default '[]'::jsonb,
  artifact_ref text,
  error_code text,
  error_message text,
  created_at timestamptz not null
);

create index if not exists db_read_execution_result_request_idx
  on db_read_execution_result (request_id, created_at asc);
