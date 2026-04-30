alter table if exists runner_execution
  add column if not exists executor_instance_id text not null default '';

alter table if exists runner_execution
  add column if not exists executor_base_url text not null default '';

create index if not exists runner_execution_executor_idx
  on runner_execution (executor_instance_id, status, updated_at desc)
  where status in ('queued','accepted','starting','running','cancelling','cancel_requested');
