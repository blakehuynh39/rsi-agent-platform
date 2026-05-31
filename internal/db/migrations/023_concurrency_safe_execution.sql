alter table if exists effect_execution
  add column if not exists queue_name text not null default 'workflow';
alter table if exists effect_execution
  add column if not exists scope_key text not null default '';
alter table if exists effect_execution
  add column if not exists task_class text not null default 'simple';
alter table if exists effect_execution
  add column if not exists priority integer not null default 0;
alter table if exists effect_execution
  add column if not exists not_before timestamptz;

update effect_execution
set queue_name = case
      when machine_kind = 'action' then 'action'
      else coalesce(nullif(payload->>'resume_queue', ''), nullif(queue_name, ''), 'workflow')
    end,
    scope_key = coalesce(nullif(payload->>'conversation_id', ''), nullif(payload->>'case_id', ''), nullif(scope_key, ''), aggregate_id),
    task_class = case
      when (payload->>'requested_artifact_count') ~ '^[0-9]+$' and (payload->>'requested_artifact_count')::integer > 0 then 'artifact'
      when lower(coalesce(payload->>'task_class', '')) = 'artifact' then 'artifact'
      when lower(coalesce(payload->>'task_class', '')) = 'improvement' then 'improvement'
      when machine_kind in ('attempt','problem_line','runtime_diagnosis') then 'improvement'
      else coalesce(nullif(payload->>'task_class', ''), nullif(task_class, ''), 'simple')
    end,
    priority = case
      when (payload->>'requested_artifact_count') ~ '^[0-9]+$' and (payload->>'requested_artifact_count')::integer > 0 then 50
      when lower(coalesce(payload->>'task_class', task_class, '')) = 'artifact' then 50
      when lower(coalesce(payload->>'task_class', task_class, '')) = 'improvement' then 10
      when machine_kind in ('attempt','problem_line','runtime_diagnosis') then 10
      else greatest(priority, 100)
    end
where coalesce(scope_key, '') = ''
   or priority = 0
   or coalesce(task_class, '') = ''
   or (machine_kind = 'action' and queue_name <> 'action');

create index if not exists effect_execution_claim_idx
  on effect_execution (queue_name, status, priority desc, created_at asc);
create index if not exists effect_execution_scope_idx
  on effect_execution (scope_key, status, lease_expires_at);

create table if not exists runner_execution (
  execution_id text primary key,
  operation_id text not null default '',
  workflow_id text not null default '',
  trace_id text not null default '',
  conversation_id text not null default '',
  case_id text not null default '',
  role text not null default '',
  status text not null,
  task jsonb not null default '{}'::jsonb,
  result jsonb not null default '{}'::jsonb,
  failure_class text not null default '',
  holder text not null default '',
  retry_count integer not null default 0,
  cancel_requested boolean not null default false,
  heartbeat_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  started_at timestamptz,
  completed_at timestamptz
);

create index if not exists runner_execution_active_idx
  on runner_execution (status, updated_at desc)
  where status in ('queued','accepted','starting','running','finalizing','cancelling','cancel_requested');
create index if not exists runner_execution_case_idx
  on runner_execution (case_id, trace_id, status);
create index if not exists runner_execution_operation_idx
  on runner_execution (operation_id, updated_at desc);
