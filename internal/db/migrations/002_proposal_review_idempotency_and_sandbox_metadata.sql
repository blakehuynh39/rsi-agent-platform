alter table if exists proposal_review add column if not exists idempotency_key text;

update proposal_review
set idempotency_key = proposal_id || ':' || decision
where coalesce(idempotency_key, '') = '';

with canonical_review as (
  select
    min(id) as canonical_id,
    proposal_id,
    decision,
    rationale,
    reviewer_id,
    coalesce(failure_class, '') as failure_class_key,
    failure_classes,
    created_at
  from proposal_review
  group by proposal_id, decision, rationale, reviewer_id, coalesce(failure_class, ''), failure_classes, created_at
),
duplicate_review as (
  select pr.id
  from proposal_review pr
  join canonical_review cr
    on cr.proposal_id = pr.proposal_id
   and cr.decision = pr.decision
   and cr.rationale = pr.rationale
   and cr.reviewer_id = pr.reviewer_id
   and cr.failure_class_key = coalesce(pr.failure_class, '')
   and cr.failure_classes = pr.failure_classes
   and cr.created_at = pr.created_at
  where pr.id <> cr.canonical_id
)
delete from proposal_review pr
using duplicate_review dr
where pr.id = dr.id;

create unique index if not exists proposal_review_proposal_idempotency_idx on proposal_review (proposal_id, idempotency_key);

alter table if exists proposal_memory add column if not exists review_id bigint;

with canonical_review as (
  select
    min(id) as canonical_id,
    proposal_id,
    decision,
    rationale,
    created_at
  from proposal_review
  group by proposal_id, decision, rationale, created_at
)
update proposal_memory pm
set review_id = cr.canonical_id
from canonical_review cr
where pm.review_id is null
  and pm.proposal_id = cr.proposal_id
  and pm.disposition = cr.decision
  and pm.review_rationale = cr.rationale
  and pm.created_at = cr.created_at;

with canonical_memory as (
  select
    min(id) as canonical_id,
    proposal_id,
    coalesce(review_id, 0) as review_id_key,
    candidate_key,
    coalesce(conversation_id, '') as conversation_key,
    coalesce(case_id, '') as case_key,
    coalesce(origin_trace_id, '') as origin_trace_key,
    evidence_trace_ids,
    hypothesis,
    diff_summary,
    review_rationale,
    disposition,
    coalesce(disposition_reason, '') as disposition_reason_key,
    coalesce(failure_class, '') as failure_class_key,
    failure_classes,
    source_eval_ids,
    linked_artifact_ids,
    linked_proposal_ids,
    created_at
  from proposal_memory
  group by
    proposal_id,
    coalesce(review_id, 0),
    candidate_key,
    coalesce(conversation_id, ''),
    coalesce(case_id, ''),
    coalesce(origin_trace_id, ''),
    evidence_trace_ids,
    hypothesis,
    diff_summary,
    review_rationale,
    disposition,
    coalesce(disposition_reason, ''),
    coalesce(failure_class, ''),
    failure_classes,
    source_eval_ids,
    linked_artifact_ids,
    linked_proposal_ids,
    created_at
),
duplicate_memory as (
  select pm.id
  from proposal_memory pm
  join canonical_memory cm
    on cm.proposal_id = pm.proposal_id
   and cm.review_id_key = coalesce(pm.review_id, 0)
   and cm.candidate_key = pm.candidate_key
   and cm.conversation_key = coalesce(pm.conversation_id, '')
   and cm.case_key = coalesce(pm.case_id, '')
   and cm.origin_trace_key = coalesce(pm.origin_trace_id, '')
   and cm.evidence_trace_ids = pm.evidence_trace_ids
   and cm.hypothesis = pm.hypothesis
   and cm.diff_summary = pm.diff_summary
   and cm.review_rationale = pm.review_rationale
   and cm.disposition = pm.disposition
   and cm.disposition_reason_key = coalesce(pm.disposition_reason, '')
   and cm.failure_class_key = coalesce(pm.failure_class, '')
   and cm.failure_classes = pm.failure_classes
   and cm.source_eval_ids = pm.source_eval_ids
   and cm.linked_artifact_ids = pm.linked_artifact_ids
   and cm.linked_proposal_ids = pm.linked_proposal_ids
   and cm.created_at = pm.created_at
  where pm.id <> cm.canonical_id
)
delete from proposal_memory pm
using duplicate_memory dm
where pm.id = dm.id;

create unique index if not exists proposal_memory_review_idx on proposal_memory (review_id) where review_id is not null;

alter table if exists repo_change_job add column if not exists sandbox_namespace text;
alter table if exists repo_change_job add column if not exists sandbox_job_name text;
alter table if exists repo_change_job add column if not exists sandbox_pod_name text;
alter table if exists repo_change_job add column if not exists validation_error text;
alter table if exists repo_change_job add column if not exists validation_ref text;
alter table if exists repo_change_job add column if not exists log_artifact_id text;
alter table if exists repo_change_job add column if not exists updated_at timestamptz not null default now();

update repo_change_job
set updated_at = created_at
where updated_at < created_at;

update proposal
set active_slot_consuming = case
  when status in ('pending_review', 'approved', 'repo_change_queued', 'repo_change_running', 'validation_pending', 'pr_open') then true
  else false
end;
