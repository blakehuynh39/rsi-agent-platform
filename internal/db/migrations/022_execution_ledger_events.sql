create table if not exists execution_ledger_event (
  id text primary key,
  execution_id text not null,
  operation_id text not null default '',
  trace_id text not null default '',
  workflow_id text not null default '',
  phase_id text not null default '',
  kind text not null,
  status text not null default '',
  seq integer not null,
  idempotency_key text not null default '',
  payload jsonb not null default '{}'::jsonb,
  recorded_at timestamptz not null default now(),
  unique (execution_id, seq)
);

create index if not exists execution_ledger_event_execution_idx
  on execution_ledger_event (execution_id, seq asc);
create index if not exists execution_ledger_event_trace_idx
  on execution_ledger_event (trace_id, recorded_at desc, seq asc);
create index if not exists execution_ledger_event_workflow_idx
  on execution_ledger_event (workflow_id, recorded_at desc, seq asc);
create index if not exists execution_ledger_event_kind_idx
  on execution_ledger_event (kind, recorded_at desc);
create index if not exists execution_ledger_event_idempotency_idx
  on execution_ledger_event (idempotency_key)
  where idempotency_key <> '';
