alter table if exists pr_attempt add column if not exists operation_id text;
alter table if exists pr_attempt add column if not exists generation integer;

alter table if exists attempt_workspace add column if not exists operation_id text;
alter table if exists attempt_workspace add column if not exists generation integer;
alter table if exists attempt_workspace add column if not exists last_error text not null default '';
alter table if exists attempt_workspace add column if not exists repairable boolean not null default false;

create table if not exists validation_run (
  id text primary key,
  proposal_id text not null,
  attempt_id text not null default '',
  conversation_id text,
  case_id text,
  origin_trace_id text,
  workspace_id text,
  operation_id text,
  generation integer,
  repo text,
  branch_name text,
  command text,
  status text not null,
  sandbox_namespace text,
  sandbox_job_name text,
  sandbox_pod_name text,
  validation_ref text,
  error_message text,
  log_artifact_id text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists validation_run_attempt_idx
  on validation_run (attempt_id, created_at desc);
create index if not exists validation_run_proposal_idx
  on validation_run (proposal_id, created_at desc);
create index if not exists validation_run_operation_idx
  on validation_run (operation_id)
  where operation_id is not null and operation_id <> '';

update proposal
set active_slot_consuming = case
  when status in ('pending_review', 'approved', 'in_progress', 'needs_review', 'repo_change_queued', 'repo_change_running', 'validation_pending', 'pr_open') then true
  else false
end;
