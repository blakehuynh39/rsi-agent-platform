alter table if exists workflow
  add column if not exists conversation_id text;
alter table if exists workflow
  add column if not exists case_id text;
alter table if exists workflow
  add column if not exists attempt_number integer not null default 0;
alter table if exists workflow
  add column if not exists parent_workflow_id text;
alter table if exists workflow
  add column if not exists failure_class text;
alter table if exists workflow
  add column if not exists failure_summary text;
alter table if exists workflow
  add column if not exists retry_decision text;
alter table if exists workflow
  add column if not exists retry_after timestamptz;
alter table if exists workflow
  add column if not exists repair_attempted boolean not null default false;
alter table if exists workflow
  add column if not exists repair_succeeded boolean not null default false;
alter table if exists workflow
  add column if not exists version bigint not null default 0;

create table if not exists workflow_line (
  case_id text primary key,
  conversation_id text not null,
  status text not null,
  current_workflow_id text,
  latest_workflow_id text,
  attempt_count integer not null default 0,
  auto_retry_budget_remaining integer not null default 0,
  last_failure_class text,
  next_retry_action text,
  retry_after timestamptz,
  line_stop_reason text,
  version bigint not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  completed_at timestamptz
);

create index if not exists workflow_line_conversation_idx
  on workflow_line (conversation_id, updated_at desc);
