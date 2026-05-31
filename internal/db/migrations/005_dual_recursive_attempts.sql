alter table if exists improvement_candidate add column if not exists line_status text not null default 'queued_for_promotion';
alter table if exists improvement_candidate add column if not exists retryable_failure_class text not null default '';
alter table if exists improvement_candidate add column if not exists last_attempt_id text not null default '';
alter table if exists improvement_candidate add column if not exists attempt_count integer not null default 0;
alter table if exists improvement_candidate add column if not exists auto_retry_budget_remaining integer not null default 3;
alter table if exists improvement_candidate add column if not exists current_target_layer text not null default 'repo_change';

update improvement_candidate
set line_status = case
  when status = 'promoted' then 'active_line'
  when status = 'needs_evidence' then 'needs_evidence'
  when status = 'dormant' then 'dormant'
  else 'queued_for_promotion'
end
where coalesce(line_status, '') = ''
   or line_status = 'queued_for_promotion';

update improvement_candidate
set current_target_layer = coalesce(nullif(target_layer, ''), 'repo_change')
where coalesce(current_target_layer, '') = '';

alter table if exists proposal add column if not exists current_attempt_id text not null default '';
alter table if exists proposal add column if not exists attempt_count integer not null default 0;
alter table if exists proposal add column if not exists auto_retry_budget_remaining integer not null default 3;
alter table if exists proposal add column if not exists last_failure_class text not null default '';
alter table if exists proposal add column if not exists next_retry_action text not null default '';
alter table if exists proposal add column if not exists line_stopped_by text not null default '';
alter table if exists proposal add column if not exists line_stop_reason text not null default '';
alter table if exists proposal add column if not exists line_stopped_at timestamptz;

alter table if exists repo_change_job add column if not exists attempt_id text not null default '';
alter table if exists pr_attempt add column if not exists attempt_id text not null default '';
alter table if exists pr_attempt add column if not exists head_sha text not null default '';
alter table if exists action_intent add column if not exists attempt_id text not null default '';
alter table if exists action_result add column if not exists attempt_id text not null default '';
alter table if exists outcome_record add column if not exists attempt_id text not null default '';
alter table if exists harness_experiment add column if not exists attempt_id text not null default '';

create table if not exists change_attempt (
  id text primary key,
  proposal_id text not null,
  candidate_key text not null,
  attempt_number integer not null,
  target_layer text not null,
  target_kind text not null default '',
  target_ref text not null default '',
  trigger text not null,
  state text not null,
  attempt_trace_id text not null default '',
  parent_attempt_id text not null default '',
  branch_name text not null default '',
  pr_url text not null default '',
  head_sha text not null default '',
  failure_class text not null default '',
  failure_summary text not null default '',
  retry_decision text not null default '',
  retry_after timestamptz,
  material_hypothesis_change boolean not null default false,
  diff_summary text not null default '',
  changed_files jsonb not null default '[]'::jsonb,
  validation_summary text not null default '',
  change_plan text not null default '',
  repo_patch text not null default '',
  validation_plan text not null default '',
  retry_assessment text not null default '',
  hypothesis_delta text not null default '',
  overlay_payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index if not exists change_attempt_proposal_attempt_number_idx on change_attempt (proposal_id, attempt_number);
create index if not exists change_attempt_proposal_created_idx on change_attempt (proposal_id, created_at desc);
create index if not exists change_attempt_candidate_created_idx on change_attempt (candidate_key, created_at desc);
