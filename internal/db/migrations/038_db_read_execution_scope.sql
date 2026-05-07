alter table db_read_request
  add column if not exists execution_scope_key text;

update db_read_request
set execution_scope_key = case
  when nullif(trim(workflow_id), '') is not null then 'workflow:' || trim(workflow_id)
  when nullif(trim(trace_id), '') is not null then 'trace:' || trim(trace_id)
  when nullif(trim(channel_id), '') is not null and nullif(trim(thread_ts), '') is not null then 'thread:' || trim(channel_id) || ':' || trim(thread_ts)
  else null
end
where execution_scope_key is null;

create unique index if not exists db_read_request_active_scope_target_idx
  on db_read_request (target, execution_scope_key)
  where execution_scope_key is not null
    and execution_scope_key <> ''
    and state in ('validating', 'pending_approval', 'approved', 'executing');
