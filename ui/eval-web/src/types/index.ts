export type NullableList<T> = T[] | null;
export type JsonPrimitive = string | number | boolean | null;
export type JsonValue = JsonPrimitive | JsonObject | JsonValue[];
export type JsonObject = {
  [key: string]: JsonValue;
};
export type CreatedAt = {
  created_at: string;
};

export type OptionalCreatedAt = {
  created_at?: string;
};

export type Timestamped = CreatedAt & {
  updated_at: string;
};

export type TabKey = "conversations" | "cases" | "proposals" | "knowledge" | "harness";
export type ProposalSegment = "active" | "candidates" | "history";
export type KnowledgeSegment = "working" | "review" | "canonical" | "stale";
export type TraceInspectorTab =
  | "overview"
  | "timeline"
  | "reasoning"
  | "tools"
  | "actions"
  | "slack"
  | "outcomes"
  | "evals"
  | "feedback"
  | "proposals";

export type TraceEvalSummary = CreatedAt & {
  run_id: string;
  verdict: string;
  score: number;
  suite_name: string;
};

export type TraceAttemptSummary = {
  trace_id: string;
  conversation_id: string;
  case_id: string;
  trigger_event_id?: string;
  supersedes_trace_id?: string;
  workflow_kind: string;
  status: string;
  thread_key: string;
  started_at: string;
  event_count: number;
  reasoning_count: number;
  tool_call_count: number;
  slack_action_count: number;
  latest_eval?: TraceEvalSummary;
};

export type CaseSummary = {
  case_id: string;
  conversation_id: string;
  kind: string;
  intent: string;
  title: string;
  summary: string;
  status: string;
  assigned_bot: string;
  latest_trace_id?: string;
  latest_trace_verdict?: string;
  recurrence: number;
  linked_proposal_ids: NullableList<string>;
  updated_at: string;
};

export type ConversationListItem = {
  conversation_id: string;
  source: string;
  external_key: string;
  title: string;
  status: string;
  active_case?: CaseSummary;
  latest_message_at: string;
  latest_trace_verdict?: string;
  open_trace_count: number;
  proposal_count: number;
};

export type ConversationEntry = CreatedAt & {
  id: string;
  event_id?: string;
  trace_id?: string;
  source: string;
  source_event_id: string;
  entry_type: string;
  actor_id?: string;
  actor_type?: string;
  body: string;
};

export type ConversationDetailResponse = {
  conversation: {
    id: string;
    source: string;
    external_key: string;
    title: string;
    status: string;
    active_case_id?: string;
  };
  active_case?: CaseSummary;
  workflow_line?: WorkflowLineSummary;
  workflow_attempts?: NullableList<WorkflowAttemptSummary>;
  cases: NullableList<CaseSummary>;
  transcript: NullableList<ConversationEntry>;
  trace_attempts: NullableList<TraceAttemptSummary>;
  action_intents: NullableList<ActionIntent>;
  action_results: NullableList<ActionResult>;
  outcomes: NullableList<OutcomeRecord>;
  knowledge_entries: NullableList<KnowledgeEntry>;
  linked_proposals: NullableList<Proposal>;
};

export type CaseDetailResponse = {
  case: CaseSummary;
  conversation: ConversationListItem;
  workflow_line?: WorkflowLineSummary;
  workflow_attempts?: NullableList<WorkflowAttemptSummary>;
  trace_attempts: NullableList<TraceAttemptSummary>;
  latest_eval_runs: NullableList<EvalRun>;
  action_intents: NullableList<ActionIntent>;
  action_results: NullableList<ActionResult>;
  outcomes: NullableList<OutcomeRecord>;
  knowledge_entries: NullableList<KnowledgeEntry>;
  linked_proposals: NullableList<Proposal>;
};

export type TraceEvent = {
  trace_id: string;
  event_type: string;
  plane: string;
  service: string;
  actor: string;
  status: string;
  description?: string;
  started_at: string;
  ended_at?: string;
};

export type EvidenceRef = {
  kind: string;
  ref: string;
  summary?: string;
};

export type ReasoningStep = {
  id: string;
  step_type: string;
  summary: string;
  evidence_refs?: NullableList<EvidenceRef>;
  alternatives?: NullableList<string>;
  confidence?: number;
  decision?: string;
  created_at: string;
};

export type ToolCallRecord = {
  id: string;
  tool_name: string;
  tool_call_id: string;
  summary?: string;
  approval_state?: string;
  interpretation_summary?: string;
  status?: string;
  created_at: string;
};

export type SlackActionRecord = {
  id: string;
  channel_id?: string;
  thread_ts?: string;
  draft_body?: string;
  final_body?: string;
  policy_verdict?: string;
  send_status?: string;
  created_at: string;
};

export type Artifact = {
  id: string;
  kind: string;
  url: string;
  source: string;
};

export type TraceDetailResponse = {
  trace: {
    summary: {
      trace_id: string;
      conversation_id: string;
      case_id: string;
      trigger_event_id?: string;
      workflow_kind: string;
      status: string;
      thread_key: string;
      started_at: string;
      event_count: number;
      artifact_count: number;
      reasoning_step_count: number;
      tool_call_count: number;
      slack_action_count: number;
      last_verdict?: string;
    };
    events: NullableList<TraceEvent>;
    artifacts: NullableList<Artifact>;
    reasoning: NullableList<ReasoningStep>;
    tool_calls: NullableList<ToolCallRecord>;
    slack_actions: NullableList<SlackActionRecord>;
  };
  conversation: ConversationListItem;
  case?: CaseSummary;
  workflow_line?: WorkflowLineSummary;
  workflow_attempts?: NullableList<WorkflowAttemptSummary>;
  transcript_slice: NullableList<ConversationEntry>;
  linked_eval_runs: NullableList<EvalRun>;
  judgments_by_eval_run: Record<string, NullableList<EvalJudgment>>;
  action_intents: NullableList<ActionIntent>;
  action_results: NullableList<ActionResult>;
  outcomes: NullableList<OutcomeRecord>;
  knowledge_entries: NullableList<KnowledgeEntry>;
  feedback_records: NullableList<FeedbackRecord>;
  linked_proposals: NullableList<Proposal>;
  harness_executions: NullableList<HarnessExecution>;
};

export type WorkflowLineSummary = {
  case_id: string;
  conversation_id: string;
  status: string;
  current_workflow_id?: string;
  latest_workflow_id?: string;
  attempt_count: number;
  auto_retry_budget_remaining: number;
  last_failure_class?: string;
  next_retry_action?: string;
  retry_after?: string;
  line_stop_reason?: string;
  updated_at: string;
};

export type RunnerDiagnostics = Record<string, unknown>;

export type WorkflowAttemptSummary = {
  workflow_id: string;
  trace_id?: string;
  conversation_id?: string;
  case_id?: string;
  workflow_kind: string;
  status: string;
  trace_status?: string;
  attempt_number: number;
  parent_workflow_id?: string;
  supersedes_trace_id?: string;
  failure_class?: string;
  failure_summary?: string;
  retry_decision?: string;
  retry_after?: string;
  runner_diagnostics?: RunnerDiagnostics;
  repair_attempted?: boolean;
  repair_succeeded?: boolean;
  created_at: string;
  updated_at: string;
  completed_at?: string;
};

export type EvalRun = CreatedAt & {
  id: string;
  trace_id: string;
  suite_name: string;
  trigger: string;
  overall_score: number;
  overall_verdict: string;
};

export type EvalJudgment = {
  id: string;
  layer: string;
  category: string;
  score: number;
  passed: boolean;
  rationale: string;
};

export type FeedbackRecord = CreatedAt & {
  id: string;
  conversation_id?: string;
  case_id?: string;
  trace_id?: string;
  target_type: string;
  target_id: string;
  score?: number;
  verdict?: string;
  labels?: NullableList<string>;
  notes?: string;
  reviewer_id: string;
};

export type ActionIntent = Timestamped & {
  id: string;
  owner_plane: string;
  conversation_id?: string;
  case_id?: string;
  trace_id?: string;
  proposal_id?: string;
  attempt_id?: string;
  kind: string;
  phase_key?: string;
  target_ref?: string;
  request_payload?: JsonObject;
  idempotency_key?: string;
  approval_mode?: string;
  approval_state?: string;
  policy_verdict?: string;
  status: string;
  superseded_by_action_id?: string;
  requested_by?: string;
  rationale?: string;
  evidence_refs?: NullableList<EvidenceRef>;
};

export type ActionResult = {
  id: string;
  action_intent_id: string;
  operation_id?: string;
  attempt_id?: string;
  attempt_number: number;
  executor: string;
  provider?: string;
  provider_ref?: string;
  request_artifact_id?: string;
  response_artifact_id?: string;
  status: string;
  error_code?: string;
  error_message?: string;
  started_at: string;
  completed_at: string;
};

export type EffectExecution = {
  id: string;
  machine_kind: string;
  aggregate_id: string;
  attempt_id?: string;
  effect_kind: string;
  status: string;
  holder?: string;
  idempotency_key: string;
  payload?: JsonObject;
  result_ref?: string;
  last_error?: string;
  retry_count: number;
  created_at: string;
  updated_at: string;
  started_at?: string;
  lease_expires_at?: string;
  completed_at?: string;
};

export type OutcomeRecord = CreatedAt & {
  id: string;
  source: string;
  source_event_id?: string;
  conversation_id?: string;
  case_id?: string;
  trace_id?: string;
  proposal_id?: string;
  attempt_id?: string;
  outcome_type: string;
  verdict: string;
  score?: number;
  summary?: string;
  details?: string;
  external_ref?: string;
  recorded_by?: string;
  recorded_at: string;
};

export type KnowledgeEntry = Timestamped & {
  id: string;
  tier: string;
  kind: string;
  scope_type: string;
  scope_id?: string;
  title: string;
  summary?: string;
  body?: string;
  structured_facts?: JsonObject;
  status: string;
  confidence?: number;
  fresh_until?: string;
  source_type: string;
  supersedes_entry_id?: string;
  contradicted_by_entry_id?: string;
};

export type KnowledgeEvidenceLink = {
  knowledge_entry_id: string;
  evidence_type: string;
  evidence_id: string;
  relevance_summary?: string;
  evidence_ref: EvidenceRef;
};

export type KnowledgeReview = {
  id: string;
  knowledge_entry_id: string;
  decision: string;
  reviewer_id: string;
  rationale?: string;
  created_at: string;
};

export type Candidate = {
  id: string;
  candidate_key: string;
  subsystem: string;
  failure_mode: string;
  intervention_type: string;
  target_layer: string;
  target_kind?: string;
  target_ref?: string;
  status: string;
  severity: string;
  recurrence_count: number;
  priority_score: number;
  confidence_score: number;
  latest_trace_id?: string;
  new_evidence_since_last_rejection: boolean;
  prior_similar_proposal_ids?: NullableList<string>;
};

export type Proposal = CreatedAt & {
  id: string;
  trace_id: string;
  conversation_id?: string;
  case_id?: string;
  origin_trace_id?: string;
  evidence_trace_ids?: NullableList<string>;
  title: string;
  category: string;
  summary: string;
  status: string;
  reviewer?: string;
  candidate_key: string;
  target_layer: string;
  target_kind?: string;
  target_ref?: string;
  source_eval_ids?: NullableList<string>;
  risk_tier?: string;
  proposed_scope?: string;
  evidence_artifact_ids?: NullableList<string>;
  active_slot_consuming: boolean;
  review_deadline?: string;
  prior_similar_proposal_ids?: NullableList<string>;
  new_evidence_since_last_rejection: boolean;
  current_attempt_id?: string;
  attempt_count: number;
  auto_retry_budget_remaining: number;
  last_failure_class?: string;
  next_retry_action?: string;
  line_stopped_by?: string;
  line_stop_reason?: string;
  line_stopped_at?: string;
  recommended_intervention_kind?: string;
  recommended_intervention_rationale?: string;
  target_surface?: string;
  touched_files?: NullableList<string>;
  validation_plan?: string;
  material_risk_summary?: string;
  recommended_disposition?: string;
};

export type ProposalListItem = Proposal & {
  repo_change_status?: string;
  pr_status?: string;
  pr_url?: string;
};

export type ProposalReview = CreatedAt & {
  proposal_id: string;
  decision: string;
  scope?: string;
  rationale: string;
  reviewer_id: string;
  failure_class?: string;
  failure_classes?: NullableList<string>;
};

export type ProposalMemory = CreatedAt & {
  id: string;
  review_id?: number;
  proposal_id: string;
  candidate_key: string;
  conversation_id?: string;
  case_id?: string;
  origin_trace_id?: string;
  evidence_trace_ids?: NullableList<string>;
  hypothesis: string;
  diff_summary: string;
  review_rationale: string;
  disposition: string;
  disposition_reason?: string;
  failure_class?: string;
  failure_classes?: NullableList<string>;
};

export type RepoChangeJob = {
  id: string;
  proposal_id: string;
  attempt_id?: string;
  status: string;
  repo: string;
  branch_name: string;
  context_summary: string;
  sandbox_namespace?: string;
  sandbox_job_name?: string;
  sandbox_pod_name?: string;
  validation_error?: string;
  validation_ref?: string;
  log_artifact_id?: string;
  created_at?: string;
  updated_at?: string;
};

export type PRAttempt = CreatedAt & {
  id: string;
  proposal_id: string;
  attempt_id?: string;
  repo?: string;
  branch_name?: string;
  pr_url?: string;
  head_sha?: string;
  status: string;
  validation_status: string;
};

export type AttemptWorkspace = Timestamped & {
  id: string;
  attempt_id: string;
  proposal_id: string;
  repo: string;
  base_ref?: string;
  branch_name: string;
  namespace?: string;
  job_name?: string;
  pod_name?: string;
  status: string;
  allowed_path_globs?: NullableList<string>;
  head_sha?: string;
  diff_summary?: string;
  expires_at?: string;
};

export type ChangeAttempt = Timestamped & {
  id: string;
  proposal_id: string;
  candidate_key: string;
  attempt_number: number;
  target_layer: string;
  target_kind?: string;
  target_ref?: string;
  trigger: string;
  state: string;
  attempt_trace_id?: string;
  parent_attempt_id?: string;
  branch_name?: string;
  pr_url?: string;
  head_sha?: string;
  failure_class?: string;
  failure_summary?: string;
  retry_decision?: string;
  retry_after?: string;
  material_hypothesis_change?: boolean;
  diff_summary?: string;
  changed_files?: NullableList<string>;
  validation_summary?: string;
  change_plan?: string;
  repo_patch?: string;
  validation_plan?: string;
  retry_assessment?: string;
  hypothesis_delta?: string;
  overlay_payload?: JsonObject;
};

export type PostMergeReplay = CreatedAt & {
  id: string;
  proposal_id: string;
  trace_id: string;
  baseline_score: number;
  candidate_score: number;
  improved: boolean;
};

export type ProposalSlots = {
  cap: number;
  active: number;
  available: number;
  active_proposal_ids: NullableList<string>;
  stale_proposal_ids: NullableList<string>;
};

export type ImprovementSettings = {
  active_proposal_cap: number;
  updated_at: string;
};

export type ProposalResponse = {
  proposals: NullableList<ProposalListItem>;
  proposal_slots: ProposalSlots;
  candidates: NullableList<Candidate>;
  settings: ImprovementSettings;
};

export type ProposalDetailResponse = {
  proposal: Proposal;
  current_phase?: ProposalCurrentPhaseSummary;
  attempts: NullableList<ChangeAttempt>;
  attempt_workspaces: NullableList<AttemptWorkspace>;
  effects: NullableList<EffectExecution>;
  reviews: NullableList<ProposalReview>;
  related_proposal_memory: NullableList<ProposalMemory>;
  repo_change_jobs: NullableList<RepoChangeJob>;
  pr_attempts: NullableList<PRAttempt>;
  post_merge_replays: NullableList<PostMergeReplay>;
  linked_trace_summaries: NullableList<TraceAttemptSummary>;
  linked_eval_runs: NullableList<EvalRun>;
  action_intents: NullableList<ActionIntent>;
  action_results: NullableList<ActionResult>;
  outcomes: NullableList<OutcomeRecord>;
  knowledge_entries: NullableList<KnowledgeEntry>;
  harness_executions: NullableList<HarnessExecution>;
};

export type ProposalCurrentPhaseSummary = {
  attempt_id?: string;
  effect_id?: string;
  effect_kind?: string;
  effect_status?: string;
  reconciliation_needed: boolean;
};

export type HarnessProfile = Timestamped & {
  id: string;
  role: string;
  name: string;
  description?: string;
  model?: string;
  reasoning_effort?: string;
  prompt_fragments?: NullableList<string>;
  few_shot_snippets?: NullableList<string>;
  tool_preference_order?: NullableList<string>;
  retrieval_bias?: string;
  reasoning_verbosity?: string;
  memory_read_enabled: boolean;
  memory_write_enabled: boolean;
  repo_ref?: string;
};

export type HarnessOverlay = Timestamped & {
  id: string;
  profile_id: string;
  role: string;
  version: string;
  status: string;
  target_kind?: string;
  target_ref?: string;
  proposal_id?: string;
  prompt_fragments?: NullableList<string>;
  few_shot_snippets?: NullableList<string>;
  tool_preference_order?: NullableList<string>;
  retrieval_bias?: string;
  reasoning_verbosity?: string;
  memory_read_enabled?: boolean;
  memory_write_enabled?: boolean;
  created_by?: string;
  approved_by?: string;
  activated_at?: string;
};

export type HarnessExperiment = Timestamped & {
  id: string;
  profile_id: string;
  overlay_id?: string;
  proposal_id?: string;
  role: string;
  status: string;
  summary?: string;
  metrics?: JsonObject;
};

export type HarnessSessionBinding = Timestamped & {
  role: string;
  scope_kind: string;
  scope_id: string;
  parent_scope_kind?: string;
  parent_scope_id?: string;
  hermes_session_id: string;
  parent_session_id?: string;
  memory_backend: string;
  assistant_peer_id?: string;
  user_peer_id?: string;
  harness_profile_id?: string;
  effective_overlay_id?: string;
  effective_overlay_version?: string;
  last_used_at: string;
};

export type HarnessMemoryArtifact = OptionalCreatedAt & {
  kind: string;
  summary: string;
  ref?: string;
  source?: string;
};

export type HarnessExecution = CreatedAt & {
  id: string;
  trace_id?: string;
  proposal_id?: string;
  role: string;
  session_scope_kind: string;
  session_scope_id: string;
  hermes_session_id: string;
  parent_session_id?: string;
  harness_profile_id?: string;
  effective_overlay_id?: string;
  effective_overlay_version?: string;
  memory_backend?: string;
  memory_reads?: NullableList<HarnessMemoryArtifact>;
  memory_writes?: NullableList<HarnessMemoryArtifact>;
};

export type KnowledgeListResponse = {
  knowledge_entries: NullableList<KnowledgeEntry>;
};

export type KnowledgeDetailResponse = {
  knowledge_entry: KnowledgeEntry;
  evidence_links: NullableList<KnowledgeEvidenceLink>;
  reviews: NullableList<KnowledgeReview>;
};

export type RuntimeRole = {
  role: string;
  reported_role?: string;
  base_url: string;
  timeout_seconds: number;
  status: string;
  backend: string;
  provider: string;
  model: string;
  provider_model?: string;
  api_mode?: string;
  reasoning_effort: string;
  available: boolean;
  healthy: boolean;
  openai_configured: boolean;
  hermes_available: boolean;
  persistence_enabled: boolean;
  hermes_home?: string;
  session_db_path?: string;
  memory_backend?: string;
  honcho_configured: boolean;
  honcho_available: boolean;
  harness_profile_id?: string;
  active_overlay_version?: string;
  error?: string;
};

export type RuntimeResponse = {
  roles: NullableList<RuntimeRole>;
};

export type AppDataResetResponse = {
  backend: string;
  reset_at: string;
  truncated_tables: NullableList<string>;
  preserved_tables: NullableList<string>;
};

export type HarnessResponse = {
  profiles: NullableList<HarnessProfile>;
  overlays: NullableList<HarnessOverlay>;
  experiments: NullableList<HarnessExperiment>;
  session_bindings: NullableList<HarnessSessionBinding>;
  executions: NullableList<HarnessExecution>;
  roles: NullableList<RuntimeRole>;
};

export type ViewState = {
  tab: TabKey;
  conversation?: string;
  case?: string;
  trace?: string;
  proposal?: string;
  knowledge?: string;
  role?: string;
};
