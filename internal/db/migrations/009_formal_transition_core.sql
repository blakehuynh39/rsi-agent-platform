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
