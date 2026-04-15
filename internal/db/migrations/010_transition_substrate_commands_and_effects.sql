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
