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
