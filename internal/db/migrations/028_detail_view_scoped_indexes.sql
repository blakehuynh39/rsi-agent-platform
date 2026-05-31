create index if not exists trace_summary_case_started_idx
  on trace_summary (case_id, started_at desc);

create index if not exists workflow_conversation_created_idx
  on workflow (conversation_id, created_at desc);
create index if not exists workflow_case_created_idx
  on workflow (case_id, created_at desc);

create index if not exists case_record_conversation_updated_idx
  on case_record (conversation_id, updated_at desc);

create index if not exists eval_run_trace_created_idx
  on eval_run (trace_id, created_at desc);
create index if not exists eval_judgment_run_created_idx
  on eval_judgment (eval_run_id, created_at asc);

create index if not exists proposal_conversation_created_idx
  on proposal (conversation_id, created_at desc);
create index if not exists proposal_case_created_idx
  on proposal (case_id, created_at desc);
create index if not exists proposal_trace_created_idx
  on proposal (trace_id, created_at desc);
create index if not exists proposal_origin_trace_created_idx
  on proposal (origin_trace_id, created_at desc);
create index if not exists proposal_evidence_trace_ids_gin_idx
  on proposal using gin (evidence_trace_ids);

create index if not exists action_intent_trace_created_idx
  on action_intent (trace_id, created_at desc);
create index if not exists action_intent_case_created_idx
  on action_intent (case_id, created_at desc);
create index if not exists action_intent_proposal_created_idx
  on action_intent (proposal_id, created_at desc);

create index if not exists outcome_record_trace_recorded_idx
  on outcome_record (trace_id, recorded_at desc);
create index if not exists outcome_record_case_recorded_idx
  on outcome_record (case_id, recorded_at desc);
create index if not exists outcome_record_proposal_recorded_idx
  on outcome_record (proposal_id, recorded_at desc);

create index if not exists feedback_record_trace_created_idx
  on feedback_record (trace_id, created_at asc);

create index if not exists improvement_candidate_latest_trace_updated_idx
  on improvement_candidate (latest_trace_id, updated_at desc);
create index if not exists improvement_candidate_evidence_trace_ids_gin_idx
  on improvement_candidate using gin (evidence_trace_ids);
