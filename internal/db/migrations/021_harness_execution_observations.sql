create table if not exists harness_execution_observation (
  id text primary key,
  execution_id text not null,
  operation_id text not null default '',
  trace_id text not null default '',
  workflow_id text not null default '',
  hermes_session_id text not null default '',
  role text not null default '',
  phase text not null,
  event_type text not null,
  status text not null default '',
  seq integer not null,
  payload jsonb not null default '{}'::jsonb,
  recorded_at timestamptz not null default now(),
  unique (execution_id, seq)
);

create index if not exists harness_execution_observation_trace_idx
  on harness_execution_observation (trace_id, recorded_at desc, seq asc);
create index if not exists harness_execution_observation_operation_idx
  on harness_execution_observation (operation_id, recorded_at desc, seq asc);
create index if not exists harness_execution_observation_session_idx
  on harness_execution_observation (hermes_session_id, recorded_at desc, seq asc);
