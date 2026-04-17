package improvementplane

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type traceEvalSummary struct {
	RunID     string    `json:"run_id"`
	Verdict   string    `json:"verdict"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"created_at"`
	SuiteName string    `json:"suite_name"`
}

type traceAttemptSummary struct {
	TraceID           string            `json:"trace_id"`
	ConversationID    string            `json:"conversation_id"`
	CaseID            string            `json:"case_id"`
	TriggerEventID    string            `json:"trigger_event_id"`
	SupersedesTraceID string            `json:"supersedes_trace_id,omitempty"`
	WorkflowKind      string            `json:"workflow_kind"`
	Status            events.Status     `json:"status"`
	ThreadKey         string            `json:"thread_key"`
	StartedAt         time.Time         `json:"started_at"`
	EventCount        int               `json:"event_count"`
	ReasoningCount    int               `json:"reasoning_count"`
	ToolCallCount     int               `json:"tool_call_count"`
	SlackActionCount  int               `json:"slack_action_count"`
	LatestEval        *traceEvalSummary `json:"latest_eval,omitempty"`
}

type workflowLineSummary struct {
	CaseID                   string     `json:"case_id"`
	ConversationID           string     `json:"conversation_id"`
	Status                   string     `json:"status"`
	CurrentWorkflowID        string     `json:"current_workflow_id,omitempty"`
	LatestWorkflowID         string     `json:"latest_workflow_id,omitempty"`
	AttemptCount             int        `json:"attempt_count"`
	AutoRetryBudgetRemaining int        `json:"auto_retry_budget_remaining"`
	LastFailureClass         string     `json:"last_failure_class,omitempty"`
	NextRetryAction          string     `json:"next_retry_action,omitempty"`
	RetryAfter               *time.Time `json:"retry_after,omitempty"`
	LineStopReason           string     `json:"line_stop_reason,omitempty"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type workflowAttemptSummary struct {
	WorkflowID        string         `json:"workflow_id"`
	TraceID           string         `json:"trace_id,omitempty"`
	ConversationID    string         `json:"conversation_id,omitempty"`
	CaseID            string         `json:"case_id,omitempty"`
	WorkflowKind      string         `json:"workflow_kind"`
	Status            string         `json:"status"`
	TraceStatus       string         `json:"trace_status,omitempty"`
	AttemptNumber     int            `json:"attempt_number"`
	ParentWorkflowID  string         `json:"parent_workflow_id,omitempty"`
	SupersedesTraceID string         `json:"supersedes_trace_id,omitempty"`
	FailureClass      string         `json:"failure_class,omitempty"`
	FailureSummary    string         `json:"failure_summary,omitempty"`
	RetryDecision     string         `json:"retry_decision,omitempty"`
	RetryAfter        *time.Time     `json:"retry_after,omitempty"`
	RunnerDiagnostics map[string]any `json:"runner_diagnostics,omitempty"`
	RepairAttempted   bool           `json:"repair_attempted,omitempty"`
	RepairSucceeded   bool           `json:"repair_succeeded,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty"`
}

type caseSummary struct {
	CaseID             string                       `json:"case_id"`
	ConversationID     string                       `json:"conversation_id"`
	Kind               string                       `json:"kind"`
	Intent             string                       `json:"intent"`
	Title              string                       `json:"title"`
	Summary            string                       `json:"summary"`
	Status             conversation.CaseStatus      `json:"status"`
	ResolutionState    conversation.ResolutionState `json:"resolution_state"`
	AssignedBot        string                       `json:"assigned_bot"`
	LatestTraceID      string                       `json:"latest_trace_id,omitempty"`
	LatestTraceVerdict string                       `json:"latest_trace_verdict,omitempty"`
	LatestOutcomeID    string                       `json:"latest_outcome_id,omitempty"`
	OutcomeScore       float64                      `json:"outcome_score,omitempty"`
	Recurrence         int                          `json:"recurrence"`
	LinkedProposalIDs  []string                     `json:"linked_proposal_ids"`
	UpdatedAt          time.Time                    `json:"updated_at"`
}

type conversationListItem struct {
	ConversationID     string       `json:"conversation_id"`
	Source             string       `json:"source"`
	ExternalKey        string       `json:"external_key"`
	Title              string       `json:"title"`
	Status             string       `json:"status"`
	ActiveCase         *caseSummary `json:"active_case,omitempty"`
	LatestMessageAt    time.Time    `json:"latest_message_at"`
	LatestTraceVerdict string       `json:"latest_trace_verdict,omitempty"`
	OpenTraceCount     int          `json:"open_trace_count"`
	ProposalCount      int          `json:"proposal_count"`
}

type conversationDetailResponse struct {
	Conversation     conversation.Conversation `json:"conversation"`
	ActiveCase       *caseSummary              `json:"active_case,omitempty"`
	WorkflowLine     *workflowLineSummary      `json:"workflow_line,omitempty"`
	WorkflowAttempts []workflowAttemptSummary  `json:"workflow_attempts"`
	Cases            []caseSummary             `json:"cases"`
	Transcript       []conversation.Entry      `json:"transcript"`
	TraceAttempts    []traceAttemptSummary     `json:"trace_attempts"`
	ActionIntents    []action.Intent           `json:"action_intents"`
	ActionResults    []action.Result           `json:"action_results"`
	Outcomes         []outcome.Record          `json:"outcomes"`
	KnowledgeEntries []knowledge.Entry         `json:"knowledge_entries"`
	LinkedProposals  []review.Proposal         `json:"linked_proposals"`
}

type caseDetailResponse struct {
	Case             caseSummary              `json:"case"`
	Conversation     conversationListItem     `json:"conversation"`
	WorkflowLine     *workflowLineSummary     `json:"workflow_line,omitempty"`
	WorkflowAttempts []workflowAttemptSummary `json:"workflow_attempts"`
	TraceAttempts    []traceAttemptSummary    `json:"trace_attempts"`
	LatestEvalRuns   []evals.Run              `json:"latest_eval_runs"`
	ActionIntents    []action.Intent          `json:"action_intents"`
	ActionResults    []action.Result          `json:"action_results"`
	Outcomes         []outcome.Record         `json:"outcomes"`
	KnowledgeEntries []knowledge.Entry        `json:"knowledge_entries"`
	LinkedProposals  []review.Proposal        `json:"linked_proposals"`
}

type traceDetailResponse struct {
	Trace              events.Trace                `json:"trace"`
	Conversation       conversationListItem        `json:"conversation"`
	Case               *caseSummary                `json:"case,omitempty"`
	WorkflowLine       *workflowLineSummary        `json:"workflow_line,omitempty"`
	WorkflowAttempts   []workflowAttemptSummary    `json:"workflow_attempts"`
	TranscriptSlice    []conversation.Entry        `json:"transcript_slice"`
	LinkedEvalRuns     []evals.Run                 `json:"linked_eval_runs"`
	JudgmentsByEvalRun map[string][]evals.Judgment `json:"judgments_by_eval_run"`
	ActionIntents      []action.Intent             `json:"action_intents"`
	ActionResults      []action.Result             `json:"action_results"`
	Outcomes           []outcome.Record            `json:"outcomes"`
	KnowledgeEntries   []knowledge.Entry           `json:"knowledge_entries"`
	FeedbackRecords    []review.FeedbackRecord     `json:"feedback_records"`
	LinkedProposals    []review.Proposal           `json:"linked_proposals"`
	HarnessExecutions  []harness.Execution         `json:"harness_executions"`
}

type proposalDetailResponse struct {
	Proposal              review.Proposal                `json:"proposal"`
	CurrentPhase          *proposalCurrentPhaseSummary   `json:"current_phase,omitempty"`
	Attempts              []improvement.ChangeAttempt    `json:"attempts"`
	AttemptWorkspaces     []improvement.AttemptWorkspace `json:"attempt_workspaces"`
	Effects               []transition.EffectExecution   `json:"effects"`
	Reviews               []review.ProposalReview        `json:"reviews"`
	RelatedProposalMemory []review.ProposalMemory        `json:"related_proposal_memory"`
	RepoChangeJobs        []improvement.RepoChangeJob    `json:"repo_change_jobs"`
	PRAttempts            []improvement.PRAttempt        `json:"pr_attempts"`
	PostMergeReplays      []improvement.PostMergeReplay  `json:"post_merge_replays"`
	LinkedTraceSummaries  []traceAttemptSummary          `json:"linked_trace_summaries"`
	LinkedEvalRuns        []evals.Run                    `json:"linked_eval_runs"`
	ActionIntents         []action.Intent                `json:"action_intents"`
	ActionResults         []action.Result                `json:"action_results"`
	Outcomes              []outcome.Record               `json:"outcomes"`
	KnowledgeEntries      []knowledge.Entry              `json:"knowledge_entries"`
	HarnessExecutions     []harness.Execution            `json:"harness_executions"`
}

type proposalCurrentPhaseSummary struct {
	AttemptID            string                  `json:"attempt_id,omitempty"`
	EffectID             string                  `json:"effect_id,omitempty"`
	EffectKind           transition.EffectKind   `json:"effect_kind,omitempty"`
	EffectStatus         transition.EffectStatus `json:"effect_status,omitempty"`
	ReconciliationNeeded bool                    `json:"reconciliation_needed"`
}

type attemptDetailResponse struct {
	Attempt           improvement.ChangeAttempt     `json:"attempt"`
	Trace             *events.Trace                 `json:"trace,omitempty"`
	Workspace         *improvement.AttemptWorkspace `json:"workspace,omitempty"`
	Effects           []transition.EffectExecution  `json:"effects"`
	ActionIntents     []action.Intent               `json:"action_intents"`
	ActionResults     []action.Result               `json:"action_results"`
	Outcomes          []outcome.Record              `json:"outcomes"`
	RepoChangeJobs    []improvement.RepoChangeJob   `json:"repo_change_jobs"`
	PRAttempts        []improvement.PRAttempt       `json:"pr_attempts"`
	HarnessExecutions []harness.Execution           `json:"harness_executions"`
}

type workflowAttemptDetailResponse struct {
	WorkflowLine     *workflowLineSummary     `json:"workflow_line,omitempty"`
	WorkflowAttempt  storepkg.Workflow        `json:"workflow_attempt"`
	Trace            *events.Trace            `json:"trace,omitempty"`
	WorkflowAttempts []workflowAttemptSummary `json:"workflow_attempts"`
}

type proposalListItem struct {
	ID                               string                          `json:"id"`
	TraceID                          string                          `json:"trace_id"`
	ConversationID                   string                          `json:"conversation_id,omitempty"`
	CaseID                           string                          `json:"case_id,omitempty"`
	OriginTraceID                    string                          `json:"origin_trace_id,omitempty"`
	EvidenceTraceIDs                 []string                        `json:"evidence_trace_ids"`
	Title                            string                          `json:"title"`
	Category                         string                          `json:"category"`
	Summary                          string                          `json:"summary"`
	Status                           review.ProposalStatus           `json:"status"`
	Reviewer                         string                          `json:"reviewer,omitempty"`
	CandidateKey                     string                          `json:"candidate_key"`
	TargetLayer                      harness.TargetLayer             `json:"target_layer"`
	TargetKind                       string                          `json:"target_kind,omitempty"`
	TargetRef                        string                          `json:"target_ref,omitempty"`
	SourceEvalIDs                    []string                        `json:"source_eval_ids"`
	RiskTier                         string                          `json:"risk_tier,omitempty"`
	ProposedScope                    string                          `json:"proposed_scope,omitempty"`
	EvidenceArtifactIDs              []string                        `json:"evidence_artifact_ids"`
	ActiveSlotConsuming              bool                            `json:"active_slot_consuming"`
	ReviewDeadline                   time.Time                       `json:"review_deadline,omitempty"`
	PriorSimilarProposalIDs          []string                        `json:"prior_similar_proposal_ids"`
	NewEvidenceSinceLastRejection    bool                            `json:"new_evidence_since_last_rejection"`
	RecommendedInterventionKind      review.ProposalInterventionKind `json:"recommended_intervention_kind,omitempty"`
	RecommendedInterventionRationale string                          `json:"recommended_intervention_rationale,omitempty"`
	TargetSurface                    string                          `json:"target_surface,omitempty"`
	TouchedFiles                     []string                        `json:"touched_files"`
	ValidationPlan                   string                          `json:"validation_plan,omitempty"`
	MaterialRiskSummary              string                          `json:"material_risk_summary,omitempty"`
	RecommendedDisposition           string                          `json:"recommended_disposition,omitempty"`
	CreatedAt                        time.Time                       `json:"created_at"`
	RepoChangeStatus                 string                          `json:"repo_change_status,omitempty"`
	PRStatus                         string                          `json:"pr_status,omitempty"`
	PRURL                            string                          `json:"pr_url,omitempty"`
}

type runtimeRoleStatus struct {
	Role                     string   `json:"role"`
	ReportedRole             string   `json:"reported_role,omitempty"`
	BaseURL                  string   `json:"base_url"`
	TimeoutSeconds           int      `json:"timeout_seconds"`
	TaskTimeoutSeconds       int      `json:"task_timeout_seconds"`
	InactivityTimeoutSeconds int      `json:"inactivity_timeout_seconds"`
	MaxIterations            int      `json:"max_iterations"`
	Status                   string   `json:"status"`
	Backend                  string   `json:"backend"`
	Provider                 string   `json:"provider"`
	Model                    string   `json:"model"`
	ProviderModel            string   `json:"provider_model,omitempty"`
	APIMode                  string   `json:"api_mode,omitempty"`
	ReasoningEffort          string   `json:"reasoning_effort"`
	HermesVersion            string   `json:"hermes_version,omitempty"`
	HermesPin                string   `json:"hermes_pin,omitempty"`
	ToolPolicyMode           string   `json:"tool_policy_mode,omitempty"`
	ToolAllowlistEffective   []string `json:"tool_allowlist_effective,omitempty"`
	BlockedToolNames         []string `json:"blocked_tool_names,omitempty"`
	Available                bool     `json:"available"`
	Healthy                  bool     `json:"healthy"`
	OpenAIConfigured         bool     `json:"openai_configured"`
	HermesAvailable          bool     `json:"hermes_available"`
	PersistenceEnabled       bool     `json:"persistence_enabled"`
	SessionContinuityStatus  string   `json:"session_continuity_status,omitempty"`
	HermesHome               string   `json:"hermes_home,omitempty"`
	SessionDBPath            string   `json:"session_db_path,omitempty"`
	ContextEngineMode        string   `json:"context_engine_mode,omitempty"`
	ContextEngineStatus      string   `json:"context_engine_status,omitempty"`
	LifecycleHookStatus      string   `json:"lifecycle_hook_status,omitempty"`
	MemoryBackend            string   `json:"memory_backend,omitempty"`
	HonchoConfigured         bool     `json:"honcho_configured"`
	HonchoAvailable          bool     `json:"honcho_available"`
	HonchoBaseURL            string   `json:"honcho_base_url,omitempty"`
	HonchoWorkspace          string   `json:"honcho_workspace,omitempty"`
	HonchoEnvironment        string   `json:"honcho_environment,omitempty"`
	HonchoRecallMode         string   `json:"honcho_recall_mode,omitempty"`
	HonchoWriteFrequency     string   `json:"honcho_write_frequency,omitempty"`
	HonchoSessionStrategy    string   `json:"honcho_session_strategy,omitempty"`
	HonchoAIPeer             string   `json:"honcho_ai_peer,omitempty"`
	HarnessProfileID         string   `json:"harness_profile_id,omitempty"`
	ActiveOverlayVersion     string   `json:"active_overlay_version,omitempty"`
	Error                    string   `json:"error,omitempty"`
}

type honchoRuntimeStatus struct {
	BaseURL            string                            `json:"base_url"`
	Status             string                            `json:"status"`
	Namespace          string                            `json:"namespace,omitempty"`
	DBSchema           string                            `json:"db_schema,omitempty"`
	CacheEnabled       bool                              `json:"cache_enabled"`
	CacheURLConfigured bool                              `json:"cache_url_configured"`
	Deriver            harness.RuntimeComponent          `json:"deriver"`
	Summary            harness.RuntimeComponent          `json:"summary"`
	DialecticLevels    map[string]harness.DialecticLevel `json:"dialectic_levels"`
	Error              string                            `json:"error,omitempty"`
}

type harnessOverviewResponse struct {
	Profiles        []harness.Profile        `json:"profiles"`
	Overlays        []harness.Overlay        `json:"overlays"`
	Experiments     []harness.Experiment     `json:"experiments"`
	SessionBindings []harness.SessionBinding `json:"session_bindings"`
	Executions      []harness.Execution      `json:"executions"`
	Roles           []runtimeRoleStatus      `json:"roles"`
	Honcho          honchoRuntimeStatus      `json:"honcho"`
}

func buildConversationList(store storepkg.Repository) []conversationListItem {
	conversations := store.ListConversations()
	traces := store.ListTraces()
	proposals := normalizeProposals(store.ListProposals())
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	traceSummaries := buildTraceSummaries(traces, latestEvalByTrace)
	tracesByConversation := map[string][]traceAttemptSummary{}
	for _, item := range traceSummaries {
		tracesByConversation[item.ConversationID] = append(tracesByConversation[item.ConversationID], item)
	}
	proposalsByConversation := map[string]int{}
	for _, proposal := range proposals {
		if proposal.ConversationID != "" {
			proposalsByConversation[proposal.ConversationID]++
		}
	}
	caseIndex := buildCaseSummaryIndex(store, traces, proposals)
	out := make([]conversationListItem, 0, len(conversations))
	for _, item := range conversations {
		tracesForConversation := tracesByConversation[item.ID]
		activeCase := caseIndex[item.ActiveCaseID]
		latestMessageAt := item.UpdatedAt
		latestTraceVerdict := ""
		openTraceCount := 0
		for _, trace := range tracesForConversation {
			if trace.StartedAt.After(latestMessageAt) {
				latestMessageAt = trace.StartedAt
			}
			if latestTraceVerdict == "" && trace.LatestEval != nil {
				latestTraceVerdict = trace.LatestEval.Verdict
			}
			if isOpenTraceStatus(trace.Status) {
				openTraceCount++
			}
		}
		out = append(out, conversationListItem{
			ConversationID:     item.ID,
			Source:             string(item.Source),
			ExternalKey:        item.ExternalKey,
			Title:              firstNonEmptyString(item.Title, item.ExternalKey),
			Status:             string(item.Status),
			ActiveCase:         activeCase,
			LatestMessageAt:    latestMessageAt,
			LatestTraceVerdict: latestTraceVerdict,
			OpenTraceCount:     openTraceCount,
			ProposalCount:      proposalsByConversation[item.ID],
		})
	}
	return out
}

func buildConversationDetail(store storepkg.Repository, conversationID string) (conversationDetailResponse, bool) {
	item, ok := store.GetConversation(conversationID)
	if !ok {
		return conversationDetailResponse{}, false
	}
	proposals := normalizeProposals(store.ListProposals())
	traces := traceSummariesForConversation(store.ListTraces(), conversationID)
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	traceSummaries := buildTraceSummaries(traces, latestEvalByTrace)
	workflowAttempts := workflowAttemptsForConversation(store.ListWorkflows(), store.ListTraces(), conversationID)
	caseIndex := buildCaseSummaryIndex(store, store.ListTraces(), proposals)
	cases := casesForConversation(store.ListCases(), conversationID, caseIndex)
	var workflowLine *workflowLineSummary
	if active := caseIndex[item.ActiveCaseID]; active != nil {
		workflowLine = workflowLineForCase(store.ListWorkflowLines(), active.CaseID)
	}
	return conversationDetailResponse{
		Conversation:     item,
		ActiveCase:       caseIndex[item.ActiveCaseID],
		WorkflowLine:     workflowLine,
		WorkflowAttempts: workflowAttempts,
		Cases:            cases,
		Transcript:       sliceOrEmpty(store.ListConversationEntries(conversationID)),
		TraceAttempts:    traceSummaries,
		ActionIntents:    sliceOrEmpty(listActionIntents(store, actionFilters{ConversationID: conversationID})),
		ActionResults:    sliceOrEmpty(flattenActionResults(store, listActionIntents(store, actionFilters{ConversationID: conversationID}))),
		Outcomes:         sliceOrEmpty(listOutcomes(store, conversationID, "", "", "")),
		KnowledgeEntries: sliceOrEmpty(relatedKnowledgeEntries(store, conversationID, "", "", "")),
		LinkedProposals:  filterProposalsByConversation(proposals, conversationID),
	}, true
}

func buildCaseList(store storepkg.Repository) []caseSummary {
	return caseSummaries(store, store.ListCases())
}

func buildCaseDetail(store storepkg.Repository, caseID string) (caseDetailResponse, bool) {
	caseIndex := buildCaseSummaryIndex(store, store.ListTraces(), normalizeProposals(store.ListProposals()))
	caseItem, ok := caseIndex[caseID]
	if !ok {
		return caseDetailResponse{}, false
	}
	conversationSummary, ok := findConversationSummary(buildConversationList(store), caseItem.ConversationID)
	if !ok {
		return caseDetailResponse{}, false
	}
	traces := traceSummariesForCase(store.ListTraces(), caseID)
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	traceSummaries := buildTraceSummaries(traces, latestEvalByTrace)
	workflowAttempts := workflowAttemptsForCase(store.ListWorkflows(), store.ListTraces(), caseID)
	return caseDetailResponse{
		Case:             *caseItem,
		Conversation:     conversationSummary,
		WorkflowLine:     workflowLineForCase(store.ListWorkflowLines(), caseID),
		WorkflowAttempts: workflowAttempts,
		TraceAttempts:    traceSummaries,
		LatestEvalRuns:   latestEvalRunsForTraceSet(store.ListEvalRuns(), traceSummaries),
		ActionIntents:    sliceOrEmpty(listActionIntents(store, actionFilters{CaseID: caseID})),
		ActionResults:    sliceOrEmpty(flattenActionResults(store, listActionIntents(store, actionFilters{CaseID: caseID}))),
		Outcomes:         sliceOrEmpty(listOutcomes(store, "", caseID, "", "")),
		KnowledgeEntries: sliceOrEmpty(relatedKnowledgeEntries(store, conversationSummary.ConversationID, caseID, "", "")),
		LinkedProposals:  filterProposalsByCase(normalizeProposals(store.ListProposals()), caseID),
	}, true
}

func buildTraceDetail(store storepkg.Repository, traceID string) (traceDetailResponse, bool) {
	trace, ok := store.GetTrace(traceID)
	if !ok {
		return traceDetailResponse{}, false
	}
	trace = normalizeTrace(trace)
	runs := filterEvalRunsForTrace(store.ListEvalRuns(), traceID)
	judgments := map[string][]evals.Judgment{}
	for _, run := range runs {
		judgments[run.ID] = sliceOrEmpty(store.ListEvalJudgments(run.ID))
	}
	conversations := buildConversationList(store)
	conversationSummary, _ := findConversationSummary(conversations, trace.Summary.ConversationID)
	caseIndex := buildCaseSummaryIndex(store, store.ListTraces(), normalizeProposals(store.ListProposals()))
	actionIntents := listActionIntents(store, actionFilters{TraceID: traceID})
	outcomes := listOutcomes(store, trace.Summary.ConversationID, trace.Summary.CaseID, traceID, "")
	extraEvidence := make([]string, 0, len(trace.Reasoning)+len(trace.ToolCalls)+len(outcomes))
	for _, step := range trace.Reasoning {
		extraEvidence = append(extraEvidence, step.ID)
	}
	for _, call := range trace.ToolCalls {
		extraEvidence = append(extraEvidence, call.ID)
	}
	for _, item := range outcomes {
		extraEvidence = append(extraEvidence, item.ID)
	}
	return traceDetailResponse{
		Trace:              trace,
		Conversation:       conversationSummary,
		Case:               caseIndex[trace.Summary.CaseID],
		WorkflowLine:       workflowLineForCase(store.ListWorkflowLines(), trace.Summary.CaseID),
		WorkflowAttempts:   workflowAttemptsForCase(store.ListWorkflows(), store.ListTraces(), trace.Summary.CaseID),
		TranscriptSlice:    transcriptSlice(store.ListConversationEntries(trace.Summary.ConversationID), trace.Summary.TriggerEventID),
		LinkedEvalRuns:     runs,
		JudgmentsByEvalRun: judgments,
		ActionIntents:      sliceOrEmpty(actionIntents),
		ActionResults:      sliceOrEmpty(flattenActionResults(store, actionIntents)),
		Outcomes:           sliceOrEmpty(outcomes),
		KnowledgeEntries:   sliceOrEmpty(relatedKnowledgeEntries(store, trace.Summary.ConversationID, trace.Summary.CaseID, traceID, "", extraEvidence...)),
		FeedbackRecords:    sliceOrEmpty(store.ListFeedback(traceID)),
		LinkedProposals:    filterProposalsForTrace(normalizeProposals(store.ListProposals()), traceID),
		HarnessExecutions:  sliceOrEmpty(filterHarnessExecutions(store.ListHarnessExecutions(), traceID, "")),
	}, true
}

func buildProposalDetail(store storepkg.Repository, proposalID string) (proposalDetailResponse, bool) {
	proposal, ok := findProposalView(store.ListProposals(), proposalID)
	if !ok {
		return proposalDetailResponse{}, false
	}
	attempts := attemptsForProposal(store, proposal.ID)
	traceSummaries := linkTraceSummaries(store.ListTraces(), proposal, store.ListPostMergeReplays())
	actionIntents := listActionIntents(store, actionFilters{ProposalID: proposal.ID})
	outcomes := listOutcomes(store, proposal.ConversationID, proposal.CaseID, "", proposal.ID)
	workspaces := filterAttemptWorkspacesByProposal(store.ListAttemptWorkspaces(), proposal.ID)
	extraEvidence := append([]string{}, proposal.EvidenceTraceIDs...)
	extraEvidence = appendUniqueStrings(extraEvidence, proposal.OriginTraceID, proposal.TraceID)
	for _, item := range outcomes {
		extraEvidence = append(extraEvidence, item.ID)
	}
	effects := sliceOrEmpty(proposalEffects(store, proposal.ID, attempts))
	return proposalDetailResponse{
		Proposal:              normalizeProposal(proposal),
		CurrentPhase:          buildProposalCurrentPhase(proposal, attempts, effects),
		Attempts:              sliceOrEmpty(attempts),
		AttemptWorkspaces:     sliceOrEmpty(workspaces),
		Effects:               effects,
		Reviews:               sliceOrEmpty(proposal.Reviews),
		RelatedProposalMemory: sliceOrEmpty(filterProposalMemory(store.ListProposalMemories(), proposal.CandidateKey)),
		RepoChangeJobs:        sliceOrEmpty(filterRepoChangeJobs(store.ListRepoChangeJobs(), proposal.ID)),
		PRAttempts:            sliceOrEmpty(filterPRAttempts(store.ListPRAttempts(), proposal.ID)),
		PostMergeReplays:      sliceOrEmpty(filterPostMergeReplays(store.ListPostMergeReplays(), proposal.ID)),
		LinkedTraceSummaries:  traceSummaries,
		LinkedEvalRuns:        sliceOrEmpty(filterEvalRunsForProposal(store.ListEvalRuns(), proposal)),
		ActionIntents:         sliceOrEmpty(actionIntents),
		ActionResults:         sliceOrEmpty(flattenActionResults(store, actionIntents)),
		Outcomes:              sliceOrEmpty(outcomes),
		KnowledgeEntries:      sliceOrEmpty(relatedKnowledgeEntries(store, proposal.ConversationID, proposal.CaseID, proposal.OriginTraceID, proposal.ID, extraEvidence...)),
		HarnessExecutions:     sliceOrEmpty(filterHarnessExecutions(store.ListHarnessExecutions(), proposal.OriginTraceID, proposal.ID)),
	}, true
}

func buildProposalCurrentPhase(proposal review.Proposal, attempts []improvement.ChangeAttempt, effects []transition.EffectExecution) *proposalCurrentPhaseSummary {
	currentAttemptID := strings.TrimSpace(proposal.CurrentAttemptID)
	if currentAttemptID == "" {
		return nil
	}
	attempt, ok := findAttemptByID(attempts, currentAttemptID)
	if !ok {
		return &proposalCurrentPhaseSummary{
			AttemptID:            currentAttemptID,
			ReconciliationNeeded: true,
		}
	}
	if effect, ok := activeAttemptEffectView(effects, attempt.ID); ok {
		return &proposalCurrentPhaseSummary{
			AttemptID:    attempt.ID,
			EffectID:     effect.ID,
			EffectKind:   effect.EffectKind,
			EffectStatus: effect.Status,
		}
	}
	return &proposalCurrentPhaseSummary{
		AttemptID:            attempt.ID,
		ReconciliationNeeded: !isAttemptTerminal(attempt.State),
	}
}

func findAttemptByID(items []improvement.ChangeAttempt, attemptID string) (improvement.ChangeAttempt, bool) {
	attemptID = strings.TrimSpace(attemptID)
	for _, item := range items {
		if strings.TrimSpace(item.ID) == attemptID {
			return item, true
		}
	}
	return improvement.ChangeAttempt{}, false
}

func activeAttemptEffectView(items []transition.EffectExecution, attemptID string) (transition.EffectExecution, bool) {
	var best transition.EffectExecution
	found := false
	attemptID = strings.TrimSpace(attemptID)
	for _, item := range items {
		if item.MachineKind != transition.MachineAttempt {
			continue
		}
		if strings.TrimSpace(item.AttemptID) != attemptID && strings.TrimSpace(item.AggregateID) != attemptID {
			continue
		}
		if item.Status != transition.EffectQueued && item.Status != transition.EffectRunning {
			continue
		}
		if !found || item.UpdatedAt.After(best.UpdatedAt) {
			best = item
			found = true
		}
	}
	return best, found
}

func buildAttemptDetail(store storepkg.Repository, proposalID string, attemptID string) (attemptDetailResponse, bool) {
	attempt, ok := store.GetChangeAttempt(attemptID)
	if !ok || attempt.ProposalID != proposalID {
		return attemptDetailResponse{}, false
	}
	var trace *events.Trace
	var workspace *improvement.AttemptWorkspace
	if strings.TrimSpace(attempt.AttemptTraceID) != "" {
		if item, ok := store.GetTrace(attempt.AttemptTraceID); ok {
			normalized := normalizeTrace(item)
			trace = &normalized
		}
	}
	if item, ok := findAttemptWorkspaceByAttempt(store.ListAttemptWorkspaces(), attempt.ID); ok {
		normalized := item
		workspace = &normalized
	}
	actionIntents := filterActionIntentsByAttempt(listActionIntents(store, actionFilters{ProposalID: proposalID}), attempt.ID)
	return attemptDetailResponse{
		Attempt:           attempt,
		Trace:             trace,
		Workspace:         workspace,
		Effects:           sliceOrEmpty(store.ListEffectExecutionsByAggregate(transition.MachineAttempt, attempt.ID)),
		ActionIntents:     sliceOrEmpty(actionIntents),
		ActionResults:     sliceOrEmpty(flattenActionResults(store, actionIntents)),
		Outcomes:          sliceOrEmpty(filterOutcomesByAttempt(listOutcomes(store, "", "", "", proposalID), attempt.ID)),
		RepoChangeJobs:    sliceOrEmpty(filterRepoChangeJobsByAttempt(store.ListRepoChangeJobs(), proposalID, attempt.ID)),
		PRAttempts:        sliceOrEmpty(filterPRAttemptsByAttempt(store.ListPRAttempts(), proposalID, attempt.ID)),
		HarnessExecutions: sliceOrEmpty(filterHarnessExecutions(store.ListHarnessExecutions(), attempt.AttemptTraceID, proposalID)),
	}, true
}

func buildWorkflowAttemptDetail(store storepkg.Repository, workflowID string) (workflowAttemptDetailResponse, bool) {
	workflow, ok := findWorkflowView(store.ListWorkflows(), workflowID)
	if !ok {
		return workflowAttemptDetailResponse{}, false
	}
	var trace *events.Trace
	if strings.TrimSpace(workflow.TraceID) != "" {
		if item, ok := store.GetTrace(workflow.TraceID); ok {
			normalized := normalizeTrace(item)
			trace = &normalized
		}
	}
	return workflowAttemptDetailResponse{
		WorkflowLine:     workflowLineForCase(store.ListWorkflowLines(), workflow.CaseID),
		WorkflowAttempt:  workflow,
		Trace:            trace,
		WorkflowAttempts: workflowAttemptsForCase(store.ListWorkflows(), store.ListTraces(), workflow.CaseID),
	}, true
}

func proposalEffects(store storepkg.Repository, proposalID string, attempts []improvement.ChangeAttempt) []transition.EffectExecution {
	items := sliceOrEmpty(store.ListEffectExecutionsByAggregate(transition.MachineProposalLine, proposalID))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		seen[item.ID] = struct{}{}
	}
	for _, attempt := range attempts {
		for _, item := range store.ListEffectExecutionsByAggregate(transition.MachineAttempt, attempt.ID) {
			if _, ok := seen[item.ID]; ok {
				continue
			}
			items = append(items, item)
			seen[item.ID] = struct{}{}
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items
}

func workflowLineForCase(items []storepkg.WorkflowLine, caseID string) *workflowLineSummary {
	item, ok := findWorkflowLineSummaryByCaseID(items, caseID)
	if !ok {
		return nil
	}
	return &workflowLineSummary{
		CaseID:                   item.CaseID,
		ConversationID:           item.ConversationID,
		Status:                   item.Status,
		CurrentWorkflowID:        item.CurrentWorkflowID,
		LatestWorkflowID:         item.LatestWorkflowID,
		AttemptCount:             item.AttemptCount,
		AutoRetryBudgetRemaining: item.AutoRetryBudgetRemaining,
		LastFailureClass:         item.LastFailureClass,
		NextRetryAction:          item.NextRetryAction,
		RetryAfter:               item.RetryAfter,
		LineStopReason:           item.LineStopReason,
		UpdatedAt:                item.UpdatedAt,
	}
}

func workflowAttemptsForConversation(workflows []storepkg.Workflow, traces []events.TraceSummary, conversationID string) []workflowAttemptSummary {
	out := make([]workflowAttemptSummary, 0)
	for _, workflow := range workflows {
		if workflow.ConversationID != conversationID {
			continue
		}
		out = append(out, workflowAttemptSummaryView(workflow, traces))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].AttemptNumber == out[j].AttemptNumber {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].AttemptNumber < out[j].AttemptNumber
	})
	return out
}

func workflowAttemptsForCase(workflows []storepkg.Workflow, traces []events.TraceSummary, caseID string) []workflowAttemptSummary {
	out := make([]workflowAttemptSummary, 0)
	for _, workflow := range workflows {
		if workflow.CaseID != caseID {
			continue
		}
		out = append(out, workflowAttemptSummaryView(workflow, traces))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].AttemptNumber == out[j].AttemptNumber {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].AttemptNumber < out[j].AttemptNumber
	})
	return out
}

func workflowAttemptSummaryView(workflow storepkg.Workflow, traces []events.TraceSummary) workflowAttemptSummary {
	summary := workflowAttemptSummary{
		WorkflowID:        workflow.ID,
		TraceID:           workflow.TraceID,
		ConversationID:    workflow.ConversationID,
		CaseID:            workflow.CaseID,
		WorkflowKind:      workflow.Kind,
		Status:            workflow.Status,
		AttemptNumber:     workflow.AttemptNumber,
		ParentWorkflowID:  workflow.ParentWorkflowID,
		FailureClass:      workflow.FailureClass,
		FailureSummary:    workflow.FailureSummary,
		RetryDecision:     workflow.RetryDecision,
		RetryAfter:        workflow.RetryAfter,
		RunnerDiagnostics: workflow.RunnerDiagnostics,
		RepairAttempted:   workflow.RepairAttempted,
		RepairSucceeded:   workflow.RepairSucceeded,
		CreatedAt:         workflow.CreatedAt,
		UpdatedAt:         workflow.UpdatedAt,
		CompletedAt:       workflow.CompletedAt,
	}
	for _, trace := range traces {
		if trace.WorkflowID != workflow.ID {
			continue
		}
		summary.TraceStatus = string(trace.Status)
		summary.SupersedesTraceID = trace.SupersedesTraceID
		if !trace.StartedAt.IsZero() {
			summary.CreatedAt = trace.StartedAt
		}
		break
	}
	return summary
}

func findWorkflowView(items []storepkg.Workflow, workflowID string) (storepkg.Workflow, bool) {
	workflowID = strings.TrimSpace(workflowID)
	for _, item := range items {
		if strings.TrimSpace(item.ID) == workflowID {
			return item, true
		}
	}
	return storepkg.Workflow{}, false
}

func findWorkflowLineSummaryByCaseID(items []storepkg.WorkflowLine, caseID string) (storepkg.WorkflowLine, bool) {
	caseID = strings.TrimSpace(caseID)
	for _, item := range items {
		if strings.TrimSpace(item.CaseID) == caseID {
			return item, true
		}
	}
	return storepkg.WorkflowLine{}, false
}

func buildProposalSummaries(store storepkg.Repository) []proposalListItem {
	proposals := normalizeProposals(store.ListProposals())
	latestJobs := latestRepoChangeJobByProposal(store.ListRepoChangeJobs())
	latestPRAttempts := latestPRAttemptByProposal(store.ListPRAttempts())
	out := make([]proposalListItem, 0, len(proposals))
	for _, item := range proposals {
		summary := proposalListItem{
			ID:                               item.ID,
			TraceID:                          item.TraceID,
			ConversationID:                   item.ConversationID,
			CaseID:                           item.CaseID,
			OriginTraceID:                    item.OriginTraceID,
			EvidenceTraceIDs:                 sliceOrEmpty(item.EvidenceTraceIDs),
			Title:                            item.Title,
			Category:                         item.Category,
			Summary:                          item.Summary,
			Status:                           item.Status,
			Reviewer:                         item.Reviewer,
			CandidateKey:                     item.CandidateKey,
			TargetLayer:                      item.TargetLayer,
			TargetKind:                       item.TargetKind,
			TargetRef:                        item.TargetRef,
			SourceEvalIDs:                    sliceOrEmpty(item.SourceEvalIDs),
			RiskTier:                         item.RiskTier,
			ProposedScope:                    item.ProposedScope,
			EvidenceArtifactIDs:              sliceOrEmpty(item.EvidenceArtifactIDs),
			ActiveSlotConsuming:              item.ActiveSlotConsuming,
			ReviewDeadline:                   item.ReviewDeadline,
			PriorSimilarProposalIDs:          sliceOrEmpty(item.PriorSimilarProposalIDs),
			NewEvidenceSinceLastRejection:    item.NewEvidenceSinceLastRejection,
			RecommendedInterventionKind:      item.RecommendedInterventionKind,
			RecommendedInterventionRationale: item.RecommendedInterventionRationale,
			TargetSurface:                    item.TargetSurface,
			TouchedFiles:                     sliceOrEmpty(item.TouchedFiles),
			ValidationPlan:                   item.ValidationPlan,
			MaterialRiskSummary:              item.MaterialRiskSummary,
			RecommendedDisposition:           item.RecommendedDisposition,
			CreatedAt:                        item.CreatedAt,
		}
		if job, ok := latestJobs[item.ID]; ok {
			summary.RepoChangeStatus = job.Status
		}
		if attempt, ok := latestPRAttempts[item.ID]; ok {
			summary.PRStatus = attempt.Status
			summary.PRURL = attempt.PRURL
		}
		out = append(out, summary)
	}
	return out
}

func buildRuntimeStatus(cfg config.Config, store storepkg.Repository) []runtimeRoleStatus {
	roleURLs := cfg.RunnerURLs()
	cache := map[string]harness.RuntimeResponse{}
	cacheErrs := map[string]error{}
	roles := []string{"prod", "proactive", "eval", "proposal"}
	out := make([]runtimeRoleStatus, 0, len(roles))
	for _, role := range roles {
		baseURL := roleURLs[role]
		if _, ok := cache[baseURL]; !ok && cacheErrs[baseURL] == nil {
			resp, err := clients.NewRunnerClientWithTimeout(baseURL, cfg.RunnerTimeoutForRole(role)).Runtime()
			if err != nil {
				cacheErrs[baseURL] = err
			} else {
				cache[baseURL] = resp
			}
		}
		item := runtimeRoleStatus{
			Role:            role,
			BaseURL:         baseURL,
			TimeoutSeconds:  int(cfg.RunnerTimeoutForRole(role).Seconds()),
			Status:          "unreachable",
			Model:           "openai/gpt-5.4",
			ReasoningEffort: "xhigh",
		}
		if profile, ok := store.GetHarnessProfile(harness.DefaultProfileID(role)); ok {
			item.HarnessProfileID = profile.ID
		}
		if overlay, ok := store.GetActiveHarnessOverlay(role); ok {
			item.ActiveOverlayVersion = overlay.Version
		}
		if err := cacheErrs[baseURL]; err != nil {
			item.Error = err.Error()
			out = append(out, item)
			continue
		}
		resp := cache[baseURL]
		item.ReportedRole = resp.Role
		item.Status = resp.Status
		item.Backend = resp.Backend
		item.Provider = resp.Provider
		item.Model = firstNonEmptyString(strings.TrimSpace(resp.Model), item.Model)
		item.ProviderModel = resp.ProviderModel
		item.APIMode = resp.APIMode
		item.ReasoningEffort = firstNonEmptyString(strings.TrimSpace(resp.ReasoningEffort), item.ReasoningEffort)
		item.HermesVersion = resp.HermesVersion
		item.HermesPin = resp.HermesPin
		item.MaxIterations = resp.MaxIterations
		item.TaskTimeoutSeconds = resp.TaskTimeoutSeconds
		item.InactivityTimeoutSeconds = resp.InactivityTimeoutSeconds
		if resp.TransportTimeoutSeconds > 0 {
			item.TimeoutSeconds = resp.TransportTimeoutSeconds
		}
		item.ToolPolicyMode = resp.ToolPolicyMode
		item.ToolAllowlistEffective = sliceOrEmpty(resp.ToolAllowlistEffective)
		item.BlockedToolNames = sliceOrEmpty(resp.BlockedToolNames)
		item.Available = resp.Available
		item.Healthy = resp.Available && strings.EqualFold(resp.Status, "ok")
		item.OpenAIConfigured = resp.OpenAIConfigured
		item.HermesAvailable = resp.HermesAvailable
		item.PersistenceEnabled = resp.PersistenceEnabled
		item.SessionContinuityStatus = resp.SessionContinuityStatus
		item.HermesHome = resp.HermesHome
		item.SessionDBPath = resp.SessionDBPath
		item.ContextEngineMode = resp.ContextEngineMode
		item.ContextEngineStatus = resp.ContextEngineStatus
		item.LifecycleHookStatus = resp.LifecycleHookStatus
		item.MemoryBackend = resp.MemoryBackend
		item.HonchoConfigured = resp.HonchoConfigured
		item.HonchoAvailable = resp.HonchoAvailable
		item.HonchoBaseURL = resp.HonchoBaseURL
		item.HonchoWorkspace = resp.HonchoWorkspace
		item.HonchoEnvironment = resp.HonchoEnvironment
		item.HonchoRecallMode = resp.HonchoRecallMode
		item.HonchoWriteFrequency = resp.HonchoWriteFrequency
		item.HonchoSessionStrategy = resp.HonchoSessionStrategy
		item.HonchoAIPeer = resp.HonchoAIPeer
		out = append(out, item)
	}
	return out
}

func buildHonchoRuntimeStatus(cfg config.Config) honchoRuntimeStatus {
	item := honchoRuntimeStatus{
		BaseURL:         cfg.HonchoRuntimeBaseURL,
		Status:          "unreachable",
		DialecticLevels: map[string]harness.DialecticLevel{},
	}
	if strings.TrimSpace(cfg.HonchoRuntimeBaseURL) == "" {
		item.Status = "disabled"
		item.Error = "RSI_HONCHO_RUNTIME_BASE_URL is not configured"
		return item
	}
	resp, err := clients.NewHonchoClient(cfg.HonchoRuntimeBaseURL).Runtime()
	if err != nil {
		item.Error = err.Error()
		return item
	}
	item.Status = firstNonEmptyString(strings.TrimSpace(resp.Status), "ok")
	item.Namespace = resp.Namespace
	item.DBSchema = resp.DBSchema
	item.CacheEnabled = resp.CacheEnabled
	item.CacheURLConfigured = resp.CacheURLConfigured
	item.Deriver = resp.Deriver
	item.Summary = resp.Summary
	item.DialecticLevels = resp.DialecticLevels
	return item
}

func buildHarnessOverview(cfg config.Config, store storepkg.Repository) harnessOverviewResponse {
	return harnessOverviewResponse{
		Profiles:        sliceOrEmpty(store.ListHarnessProfiles()),
		Overlays:        sliceOrEmpty(store.ListHarnessOverlays()),
		Experiments:     sliceOrEmpty(store.ListHarnessExperiments()),
		SessionBindings: sliceOrEmpty(store.ListHarnessSessionBindings()),
		Executions:      sliceOrEmpty(store.ListHarnessExecutions()),
		Roles:           buildRuntimeStatus(cfg, store),
		Honcho:          buildHonchoRuntimeStatus(cfg),
	}
}

func buildCaseSummaryIndex(store storepkg.Repository, traces []events.TraceSummary, proposals []review.Proposal) map[string]*caseSummary {
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	proposalIDsByCase := map[string][]string{}
	for _, proposal := range proposals {
		if proposal.CaseID == "" {
			continue
		}
		proposalIDsByCase[proposal.CaseID] = appendUniqueStrings(proposalIDsByCase[proposal.CaseID], proposal.ID)
	}
	traceVerdictByCase := map[string]string{}
	recurrenceByCase := map[string]int{}
	for _, trace := range traces {
		recurrenceByCase[trace.CaseID]++
		if trace.CaseID == "" {
			continue
		}
		if current, ok := traceVerdictByCase[trace.CaseID]; ok && current != "" {
			continue
		}
		if run, ok := latestEvalByTrace[trace.TraceID]; ok {
			traceVerdictByCase[trace.CaseID] = run.OverallVerdict
		}
	}
	out := map[string]*caseSummary{}
	for _, item := range store.ListCases() {
		summary := &caseSummary{
			CaseID:             item.ID,
			ConversationID:     item.ConversationID,
			Kind:               item.Kind,
			Intent:             item.Intent,
			Title:              item.Title,
			Summary:            item.Summary,
			Status:             item.Status,
			ResolutionState:    item.ResolutionState,
			AssignedBot:        item.AssignedBot,
			LatestTraceID:      item.LatestTraceID,
			LatestTraceVerdict: traceVerdictByCase[item.ID],
			LatestOutcomeID:    item.LatestOutcomeID,
			OutcomeScore:       item.OutcomeScore,
			Recurrence:         recurrenceByCase[item.ID],
			LinkedProposalIDs:  sliceOrEmpty(proposalIDsByCase[item.ID]),
			UpdatedAt:          item.UpdatedAt,
		}
		out[item.ID] = summary
	}
	return out
}

func caseSummaries(store storepkg.Repository, cases []conversation.Case) []caseSummary {
	index := buildCaseSummaryIndex(store, store.ListTraces(), normalizeProposals(store.ListProposals()))
	out := make([]caseSummary, 0, len(cases))
	for _, item := range cases {
		if summary, ok := index[item.ID]; ok {
			out = append(out, *summary)
		}
	}
	return out
}

func casesForConversation(cases []conversation.Case, conversationID string, index map[string]*caseSummary) []caseSummary {
	out := make([]caseSummary, 0)
	for _, item := range cases {
		if item.ConversationID != conversationID {
			continue
		}
		if summary, ok := index[item.ID]; ok {
			out = append(out, *summary)
		}
	}
	return out
}

func buildTraceSummaries(traces []events.TraceSummary, latestEvalByTrace map[string]evals.Run) []traceAttemptSummary {
	out := make([]traceAttemptSummary, 0, len(traces))
	for _, trace := range traces {
		item := traceAttemptSummary{
			TraceID:           trace.TraceID,
			ConversationID:    trace.ConversationID,
			CaseID:            trace.CaseID,
			TriggerEventID:    trace.TriggerEventID,
			SupersedesTraceID: trace.SupersedesTraceID,
			WorkflowKind:      trace.WorkflowKind,
			Status:            trace.Status,
			ThreadKey:         trace.ThreadKey,
			StartedAt:         trace.StartedAt,
			EventCount:        trace.EventCount,
			ReasoningCount:    trace.ReasoningStepCount,
			ToolCallCount:     trace.ToolCallCount,
			SlackActionCount:  trace.SlackActionCount,
		}
		if run, ok := latestEvalByTrace[trace.TraceID]; ok {
			item.LatestEval = &traceEvalSummary{
				RunID:     run.ID,
				Verdict:   run.OverallVerdict,
				Score:     run.OverallScore,
				CreatedAt: run.CreatedAt,
				SuiteName: run.SuiteName,
			}
		}
		out = append(out, item)
	}
	return out
}

func traceSummariesForConversation(traces []events.TraceSummary, conversationID string) []events.TraceSummary {
	out := make([]events.TraceSummary, 0)
	for _, item := range traces {
		if item.ConversationID == conversationID {
			out = append(out, item)
		}
	}
	return out
}

func traceSummariesForCase(traces []events.TraceSummary, caseID string) []events.TraceSummary {
	out := make([]events.TraceSummary, 0)
	for _, item := range traces {
		if item.CaseID == caseID {
			out = append(out, item)
		}
	}
	return out
}

func filterProposalsForTrace(items []review.Proposal, traceID string) []review.Proposal {
	out := make([]review.Proposal, 0)
	for _, item := range items {
		if item.TraceID == traceID || item.OriginTraceID == traceID || slicesContain(item.EvidenceTraceIDs, traceID) {
			out = append(out, item)
		}
	}
	return out
}

func filterProposalsByConversation(items []review.Proposal, conversationID string) []review.Proposal {
	out := make([]review.Proposal, 0)
	for _, item := range items {
		if item.ConversationID == conversationID {
			out = append(out, item)
		}
	}
	return out
}

func filterProposalsByCase(items []review.Proposal, caseID string) []review.Proposal {
	out := make([]review.Proposal, 0)
	for _, item := range items {
		if item.CaseID == caseID {
			out = append(out, item)
		}
	}
	return out
}

func filterEvalRunsForTrace(items []evals.Run, traceID string) []evals.Run {
	out := make([]evals.Run, 0)
	for _, item := range items {
		if item.TraceID == traceID {
			out = append(out, item)
		}
	}
	return out
}

func filterEvalRunsForProposal(items []evals.Run, proposal review.Proposal) []evals.Run {
	byID := map[string]struct{}{}
	for _, sourceEvalID := range proposal.SourceEvalIDs {
		byID[sourceEvalID] = struct{}{}
	}
	out := make([]evals.Run, 0)
	for _, item := range items {
		if item.TraceID == proposal.TraceID || item.TraceID == proposal.OriginTraceID || slicesContain(proposal.EvidenceTraceIDs, item.TraceID) {
			out = append(out, item)
			continue
		}
		if _, ok := byID[item.ID]; ok {
			out = append(out, item)
		}
	}
	return out
}

func latestEvalRunByTrace(items []evals.Run) map[string]evals.Run {
	out := map[string]evals.Run{}
	for _, item := range items {
		current, ok := out[item.TraceID]
		if !ok || item.CreatedAt.After(current.CreatedAt) {
			out[item.TraceID] = item
		}
	}
	return out
}

func latestEvalRunsForTraceSet(items []evals.Run, traces []traceAttemptSummary) []evals.Run {
	traceSet := map[string]struct{}{}
	for _, item := range traces {
		traceSet[item.TraceID] = struct{}{}
	}
	latest := map[string]evals.Run{}
	for _, item := range items {
		if _, ok := traceSet[item.TraceID]; !ok {
			continue
		}
		current, ok := latest[item.TraceID]
		if !ok || item.CreatedAt.After(current.CreatedAt) {
			latest[item.TraceID] = item
		}
	}
	out := make([]evals.Run, 0, len(latest))
	for _, item := range latest {
		out = append(out, item)
	}
	return out
}

func filterRepoChangeJobs(items []improvement.RepoChangeJob, proposalID string) []improvement.RepoChangeJob {
	out := make([]improvement.RepoChangeJob, 0)
	for _, item := range items {
		if item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	return out
}

func filterPRAttempts(items []improvement.PRAttempt, proposalID string) []improvement.PRAttempt {
	out := make([]improvement.PRAttempt, 0)
	for _, item := range items {
		if item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	return out
}

func filterRepoChangeJobsByAttempt(items []improvement.RepoChangeJob, proposalID string, attemptID string) []improvement.RepoChangeJob {
	out := make([]improvement.RepoChangeJob, 0)
	for _, item := range items {
		if item.ProposalID == proposalID && item.AttemptID == attemptID {
			out = append(out, item)
		}
	}
	return out
}

func filterPRAttemptsByAttempt(items []improvement.PRAttempt, proposalID string, attemptID string) []improvement.PRAttempt {
	out := make([]improvement.PRAttempt, 0)
	for _, item := range items {
		if item.ProposalID == proposalID && item.AttemptID == attemptID {
			out = append(out, item)
		}
	}
	return out
}

func filterAttemptWorkspacesByProposal(items []improvement.AttemptWorkspace, proposalID string) []improvement.AttemptWorkspace {
	out := make([]improvement.AttemptWorkspace, 0)
	for _, item := range items {
		if item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	return out
}

func findAttemptWorkspaceByAttempt(items []improvement.AttemptWorkspace, attemptID string) (improvement.AttemptWorkspace, bool) {
	for _, item := range items {
		if item.AttemptID == attemptID {
			return item, true
		}
	}
	return improvement.AttemptWorkspace{}, false
}

func filterPostMergeReplays(items []improvement.PostMergeReplay, proposalID string) []improvement.PostMergeReplay {
	out := make([]improvement.PostMergeReplay, 0)
	for _, item := range items {
		if item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	return out
}

func latestRepoChangeJobByProposal(items []improvement.RepoChangeJob) map[string]improvement.RepoChangeJob {
	out := map[string]improvement.RepoChangeJob{}
	for _, item := range items {
		current, ok := out[item.ProposalID]
		if !ok || itemUpdatedAt(item).After(itemUpdatedAt(current)) {
			out[item.ProposalID] = item
		}
	}
	return out
}

func latestPRAttemptByProposal(items []improvement.PRAttempt) map[string]improvement.PRAttempt {
	out := map[string]improvement.PRAttempt{}
	for _, item := range items {
		current, ok := out[item.ProposalID]
		if !ok || item.CreatedAt.After(current.CreatedAt) {
			out[item.ProposalID] = item
		}
	}
	return out
}

func linkTraceSummaries(items []events.TraceSummary, proposal review.Proposal, replays []improvement.PostMergeReplay) []traceAttemptSummary {
	index := map[string]events.TraceSummary{}
	for _, item := range items {
		index[item.TraceID] = item
	}
	out := make([]events.TraceSummary, 0)
	seen := map[string]struct{}{}
	for _, traceID := range proposalLinkedTraceIDs(proposal, replays) {
		if _, ok := seen[traceID]; ok {
			continue
		}
		seen[traceID] = struct{}{}
		if trace, ok := index[traceID]; ok {
			out = append(out, trace)
		}
	}
	return buildTraceSummaries(out, map[string]evals.Run{})
}

func proposalLinkedTraceIDs(proposal review.Proposal, replays []improvement.PostMergeReplay) []string {
	out := []string{firstNonEmptyString(proposal.OriginTraceID, proposal.TraceID)}
	out = append(out, proposal.EvidenceTraceIDs...)
	for _, replay := range replays {
		if replay.ProposalID == proposal.ID && strings.TrimSpace(replay.TraceID) != "" {
			out = append(out, replay.TraceID)
		}
	}
	return out
}

func normalizeProposals(items []review.Proposal) []review.Proposal {
	out := make([]review.Proposal, 0, len(items))
	for _, item := range items {
		out = append(out, normalizeProposal(item))
	}
	return out
}

func normalizeProposal(item review.Proposal) review.Proposal {
	item.SourceEvalIDs = sliceOrEmpty(item.SourceEvalIDs)
	item.EvidenceArtifactIDs = sliceOrEmpty(item.EvidenceArtifactIDs)
	item.PriorSimilarProposalIDs = sliceOrEmpty(item.PriorSimilarProposalIDs)
	item.EvidenceTraceIDs = sliceOrEmpty(item.EvidenceTraceIDs)
	item.Reviews = sliceOrEmpty(item.Reviews)
	return item
}

func filterHarnessExecutions(items []harness.Execution, traceID string, proposalID string) []harness.Execution {
	out := make([]harness.Execution, 0)
	for _, item := range items {
		if traceID != "" && item.TraceID == traceID {
			out = append(out, item)
			continue
		}
		if proposalID != "" && item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	return out
}

func filterActionIntentsByAttempt(items []action.Intent, attemptID string) []action.Intent {
	out := make([]action.Intent, 0)
	for _, item := range items {
		if strings.TrimSpace(item.AttemptID) == strings.TrimSpace(attemptID) {
			out = append(out, item)
		}
	}
	return out
}

func filterOutcomesByAttempt(items []outcome.Record, attemptID string) []outcome.Record {
	out := make([]outcome.Record, 0)
	for _, item := range items {
		if strings.TrimSpace(item.AttemptID) == strings.TrimSpace(attemptID) {
			out = append(out, item)
		}
	}
	return out
}

func itemUpdatedAt(item improvement.RepoChangeJob) time.Time {
	if !item.UpdatedAt.IsZero() {
		return item.UpdatedAt
	}
	if !item.CreatedAt.IsZero() {
		return item.CreatedAt
	}
	return time.Time{}
}

func normalizeTrace(trace events.Trace) events.Trace {
	trace.Events = sliceOrEmpty(trace.Events)
	trace.Artifacts = sliceOrEmpty(trace.Artifacts)
	trace.Reasoning = sliceOrEmpty(trace.Reasoning)
	trace.ToolCalls = sliceOrEmpty(trace.ToolCalls)
	trace.SlackActions = sliceOrEmpty(trace.SlackActions)
	return trace
}

func transcriptSlice(entries []conversation.Entry, triggerEventID string) []conversation.Entry {
	if len(entries) == 0 {
		return []conversation.Entry{}
	}
	if triggerEventID == "" {
		if len(entries) <= 12 {
			return entries
		}
		return entries[len(entries)-12:]
	}
	position := -1
	for i, item := range entries {
		if item.EventID == triggerEventID {
			position = i
		}
	}
	if position == -1 {
		if len(entries) <= 12 {
			return entries
		}
		return entries[len(entries)-12:]
	}
	start := position - 6
	if start < 0 {
		start = 0
	}
	end := position + 1
	if end < len(entries) {
		end = minInt(end+5, len(entries))
	}
	return entries[start:end]
}

func findConversationSummary(items []conversationListItem, conversationID string) (conversationListItem, bool) {
	for _, item := range items {
		if item.ConversationID == conversationID {
			return item, true
		}
	}
	return conversationListItem{}, false
}

func findProposalView(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}

func isOpenTraceStatus(status events.Status) bool {
	switch status {
	case events.StatusQueued, events.StatusRunning, events.StatusNeedsHuman, events.StatusInReview, events.StatusReplayed:
		return true
	default:
		return false
	}
}

func runtimeSummary(status runtimeRoleStatus) string {
	if status.Error != "" {
		return fmt.Sprintf("%s unavailable: %s", status.Role, status.Error)
	}
	return fmt.Sprintf("%s -> %s %s effort=%s", status.Role, status.Backend, status.Model, status.ReasoningEffort)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func slicesContain(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func appendUniqueStrings(existing []string, values ...string) []string {
	seen := map[string]struct{}{}
	out := append([]string(nil), existing...)
	for _, item := range existing {
		seen[item] = struct{}{}
	}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
