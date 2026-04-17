package improvementplane

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func seedRouterActionIntent(t *testing.T, store *storepkg.MemoryStore, intent action.Intent, prefix string, resultPayload map[string]any) action.Intent {
	t.Helper()
	now := intent.CreatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if intent.ID == "" {
		intent.ID = prefix + "-action"
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: intent.ID,
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   prefix + "-queue",
		OccurredAt:  now,
		Payload: map[string]any{
			"conversation_id": intent.ConversationID,
			"case_id":         intent.CaseID,
			"trace_id":        intent.TraceID,
			"kind":            string(intent.Kind),
			"target_ref":      intent.TargetRef,
			"request_payload": intent.RequestPayload,
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_queued) error = %v", err)
	}
	if resultPayload != nil {
		startedAt := now
		completedAt := now.Add(time.Second)
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionStart),
			CommandID:   prefix + "-start",
			OccurredAt:  startedAt,
			Payload:     map[string]any{"operation_id": prefix + "-op"},
		}); err != nil {
			t.Fatalf("SubmitCommand(action_started) error = %v", err)
		}
		payload := map[string]any{
			"operation_id": prefix + "-op",
			"started_at":   startedAt,
			"completed_at": completedAt,
		}
		for key, value := range resultPayload {
			payload[key] = value
		}
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionSucceed),
			CommandID:   prefix + "-succeed",
			OccurredAt:  completedAt,
			Payload:     payload,
		}); err != nil {
			t.Fatalf("SubmitCommand(action_succeeded) error = %v", err)
		}
	}
	created, ok := store.GetActionIntent(intent.ID)
	if !ok {
		t.Fatalf("expected action intent %s", intent.ID)
	}
	return created
}

func seedRouterKnowledgeEntry(t *testing.T, store *storepkg.MemoryStore, entryID string, entry knowledge.Entry, links []knowledge.EvidenceLink, commandID string) knowledge.Entry {
	t.Helper()
	occurredAt := entry.UpdatedAt
	if occurredAt.IsZero() {
		occurredAt = entry.CreatedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	createdAt := entry.CreatedAt
	if createdAt.IsZero() {
		createdAt = occurredAt
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineKnowledge,
		AggregateID: entryID,
		CommandKind: string(transition.CommandKnowledgeRecordDraft),
		CommandID:   commandID,
		Actor:       "tester",
		OccurredAt:  occurredAt,
		Payload: map[string]any{
			"tier":           string(entry.Tier),
			"kind":           string(entry.Kind),
			"scope_type":     string(entry.ScopeType),
			"scope_id":       entry.ScopeID,
			"title":          entry.Title,
			"summary":        entry.Summary,
			"body":           entry.Body,
			"status":         string(entry.Status),
			"confidence":     entry.Confidence,
			"source_type":    string(entry.SourceType),
			"created_at":     createdAt,
			"updated_at":     occurredAt,
			"evidence_links": links,
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(knowledge_record_draft) error = %v", err)
	}
	created, ok := store.GetKnowledgeEntry(entryID)
	if !ok {
		t.Fatalf("expected knowledge entry %s", entryID)
	}
	return created
}

func seedRouterRepoChangeJobViaCommand(t *testing.T, store *storepkg.MemoryStore, proposal review.Proposal, commandPrefix string, jobID string, branchName string) {
	t.Helper()
	if strings.TrimSpace(proposal.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt for proposal %s", proposal.ID)
	}
	occurredAt := time.Now().UTC()
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposal.ID,
		CommandKind: string(transition.CommandProposalMarkRepoChangeQueued),
		CommandID:   commandPrefix + "-proposal-queued",
		Actor:       "tester",
		OccurredAt:  occurredAt,
	}); err != nil {
		t.Fatalf("SubmitCommand(proposal_mark_repo_change_queued) error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAttempt,
		AggregateID: proposal.CurrentAttemptID,
		CommandKind: string(transition.CommandWorkspaceReady),
		CommandID:   commandPrefix + "-workspace-ready",
		Actor:       "tester",
		OccurredAt:  occurredAt.Add(time.Millisecond),
		Payload: map[string]any{
			"workspace_id":       "workspace-" + proposal.CurrentAttemptID,
			"job_id":             jobID,
			"repo":               "rsi-agent-platform",
			"base_ref":           "main",
			"branch_name":        branchName,
			"sandbox_namespace":  "rsi-platform",
			"sandbox_job_name":   "workspace-job-" + proposal.CurrentAttemptID,
			"sandbox_pod_name":   "workspace-pod-" + proposal.CurrentAttemptID,
			"allowed_path_globs": []string{"internal/**"},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workspace_ready) error = %v", err)
	}
}

func TestRouterConversationCaseAndTraceEndpoints(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceSummaries := store.ListTraces()
	if len(traceSummaries) == 0 {
		t.Fatal("expected seeded trace")
	}
	trace, ok := store.GetTrace(traceSummaries[0].TraceID)
	if !ok {
		t.Fatal("expected seeded trace detail")
	}
	now := time.Now().UTC()
	seedRouterActionIntent(t, store, action.Intent{
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        trace.Summary.TraceID,
		Kind:           action.KindSlackPost,
		CreatedAt:      now,
	}, "router-conversation", map[string]any{
		"executor":     "tool-gateway",
		"provider":     "slack",
		"provider_ref": "slack-provider-ref",
	})
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: trace.Summary.TraceID,
		CommandKind: string(transition.CommandProblemLineRecordOutcome),
		CommandID:   "router-conversation-outcome",
		Actor:       "tester",
		OccurredAt:  now,
		Payload: map[string]any{
			"source":          "operator",
			"recorded_by":     "tester",
			"conversation_id": trace.Summary.ConversationID,
			"case_id":         trace.Summary.CaseID,
			"trace_id":        trace.Summary.TraceID,
			"outcome_type":    string(outcome.TypeAnswerQuality),
			"verdict":         string(outcome.VerdictPositive),
			"score":           1,
			"summary":         "Trace resolved the question well.",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(problem_line_record_outcome) error = %v", err)
	}
	entry := seedRouterKnowledgeEntry(t, store, "knowledge-trace-learning", knowledge.Entry{
		Tier:       knowledge.TierWorking,
		Kind:       knowledge.KindFact,
		ScopeType:  knowledge.ScopeCase,
		ScopeID:    trace.Summary.CaseID,
		Title:      "Trace learning",
		Summary:    "Structured action/outcome evidence is present.",
		Status:     knowledge.StatusDraft,
		Confidence: 0.8,
		SourceType: knowledge.SourceAgent,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, []knowledge.EvidenceLink{
		{
			EvidenceType: "trace",
			EvidenceID:   trace.Summary.TraceID,
			EvidenceRef:  events.EvidenceRef{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
		},
	}, "router-conversation-knowledge")
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	listReq := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("conversation list status = %d, want %d", listRec.Code, http.StatusOK)
	}

	var listPayload struct {
		Conversations []map[string]any `json:"conversations"`
	}
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode conversation list: %v", err)
	}
	if len(listPayload.Conversations) == 0 {
		t.Fatal("expected at least one conversation in summary list")
	}
	if _, ok := listPayload.Conversations[0]["conversation_id"]; !ok {
		t.Fatal("expected conversation list item to include conversation_id")
	}
	if _, ok := listPayload.Conversations[0]["trace_attempts"]; ok {
		t.Fatal("conversation list should not include transcript or trace detail payloads")
	}

	conversationID, _ := listPayload.Conversations[0]["conversation_id"].(string)
	if conversationID == "" {
		t.Fatal("expected non-empty conversation_id from conversation list")
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/conversations/"+conversationID, nil)
	detailRec := httptest.NewRecorder()
	router.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("conversation detail status = %d, want %d", detailRec.Code, http.StatusOK)
	}

	var detailPayload map[string]any
	if err := json.NewDecoder(detailRec.Body).Decode(&detailPayload); err != nil {
		t.Fatalf("decode conversation detail: %v", err)
	}
	if _, ok := detailPayload["conversation"].(map[string]any); !ok {
		t.Fatal("expected conversation detail payload to include conversation object")
	}
	traceAttempts, ok := detailPayload["trace_attempts"].([]any)
	if !ok || len(traceAttempts) == 0 {
		t.Fatal("expected trace_attempts to be a non-empty JSON array")
	}
	if _, ok := detailPayload["transcript"].([]any); !ok {
		t.Fatal("expected transcript to be a JSON array")
	}
	if _, ok := detailPayload["linked_proposals"].([]any); !ok {
		t.Fatal("expected linked_proposals to be a JSON array")
	}

	caseReq := httptest.NewRequest(http.MethodGet, "/api/cases", nil)
	caseRec := httptest.NewRecorder()
	router.ServeHTTP(caseRec, caseReq)
	if caseRec.Code != http.StatusOK {
		t.Fatalf("case list status = %d, want %d", caseRec.Code, http.StatusOK)
	}

	var casePayload struct {
		Cases []map[string]any `json:"cases"`
	}
	if err := json.NewDecoder(caseRec.Body).Decode(&casePayload); err != nil {
		t.Fatalf("decode case list: %v", err)
	}
	if len(casePayload.Cases) == 0 {
		t.Fatal("expected at least one case")
	}

	traceSummary, _ := traceAttempts[0].(map[string]any)
	traceID, _ := traceSummary["trace_id"].(string)
	if traceID == "" {
		t.Fatal("expected trace_id in trace attempts")
	}
	traceReq := httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID, nil)
	traceRec := httptest.NewRecorder()
	router.ServeHTTP(traceRec, traceReq)
	if traceRec.Code != http.StatusOK {
		t.Fatalf("trace detail status = %d, want %d", traceRec.Code, http.StatusOK)
	}

	var tracePayload map[string]any
	if err := json.NewDecoder(traceRec.Body).Decode(&tracePayload); err != nil {
		t.Fatalf("decode trace detail: %v", err)
	}
	if _, ok := tracePayload["trace"].(map[string]any); !ok {
		t.Fatal("expected trace detail payload to include trace object")
	}
	if _, ok := tracePayload["transcript_slice"].([]any); !ok {
		t.Fatal("expected transcript_slice to be a JSON array")
	}
	if _, ok := tracePayload["feedback_records"].([]any); !ok {
		t.Fatal("expected feedback_records to be a JSON array")
	}
	if _, ok := tracePayload["linked_proposals"].([]any); !ok {
		t.Fatal("expected linked_proposals to be a JSON array")
	}
	if _, ok := tracePayload["judgments_by_eval_run"].(map[string]any); !ok {
		t.Fatal("expected judgments_by_eval_run to be a JSON object")
	}
	if items, ok := tracePayload["action_intents"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected action_intents to be a non-empty JSON array")
	}
	if items, ok := tracePayload["action_results"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected action_results to be a non-empty JSON array")
	}
	if items, ok := tracePayload["outcomes"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected outcomes to be a non-empty JSON array")
	}
	if items, ok := tracePayload["knowledge_entries"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected knowledge_entries to be a non-empty JSON array")
	}

	actionReq := httptest.NewRequest(http.MethodGet, "/api/actions?trace="+traceID, nil)
	actionRec := httptest.NewRecorder()
	router.ServeHTTP(actionRec, actionReq)
	if actionRec.Code != http.StatusOK {
		t.Fatalf("action list status = %d, want %d", actionRec.Code, http.StatusOK)
	}

	outcomeReq := httptest.NewRequest(http.MethodGet, "/api/outcomes?trace="+traceID, nil)
	outcomeRec := httptest.NewRecorder()
	router.ServeHTTP(outcomeRec, outcomeReq)
	if outcomeRec.Code != http.StatusOK {
		t.Fatalf("outcome list status = %d, want %d", outcomeRec.Code, http.StatusOK)
	}

	knowledgeReq := httptest.NewRequest(http.MethodGet, "/api/knowledge", nil)
	knowledgeRec := httptest.NewRecorder()
	router.ServeHTTP(knowledgeRec, knowledgeReq)
	if knowledgeRec.Code != http.StatusOK {
		t.Fatalf("knowledge list status = %d, want %d", knowledgeRec.Code, http.StatusOK)
	}

	knowledgeDetailReq := httptest.NewRequest(http.MethodGet, "/api/knowledge/"+entry.ID, nil)
	knowledgeDetailRec := httptest.NewRecorder()
	router.ServeHTTP(knowledgeDetailRec, knowledgeDetailReq)
	if knowledgeDetailRec.Code != http.StatusOK {
		t.Fatalf("knowledge detail status = %d, want %d", knowledgeDetailRec.Code, http.StatusOK)
	}
}

func TestRouterFeedbackAndReplayRoutesSubmitProblemLineCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatal("expected seeded trace")
	}
	now := time.Now().UTC()
	intent := seedRouterActionIntent(t, store, action.Intent{
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        traceID,
		Kind:           action.KindToolRead,
		CreatedAt:      now,
	}, "router-feedback", map[string]any{
		"executor":     "tool-gateway",
		"provider":     "github",
		"provider_ref": "repo-activity-ref",
	})
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	feedbackReq := httptest.NewRequest(http.MethodPost, "/api/feedback", strings.NewReader(`{
		"target_type":"action_intent",
		"target_id":"`+intent.ID+`",
		"verdict":"useful",
		"reviewer_id":"alice"
	}`))
	feedbackRec := httptest.NewRecorder()
	router.ServeHTTP(feedbackRec, feedbackReq)
	if feedbackRec.Code != http.StatusCreated {
		t.Fatalf("feedback status = %d, want %d", feedbackRec.Code, http.StatusCreated)
	}
	foundFeedbackEvent := false
	for _, item := range store.ListDomainEvents() {
		if item.EventKind == "problem_line_feedback_recorded" {
			foundFeedbackEvent = true
			break
		}
	}
	if !foundFeedbackEvent {
		t.Fatal("expected feedback route to record a problem-line domain event")
	}

	replayReq := httptest.NewRequest(http.MethodPost, "/api/traces/"+traceID+"/replay", strings.NewReader(`{"requested_by":"alice"}`))
	replayRec := httptest.NewRecorder()
	router.ServeHTTP(replayRec, replayReq)
	if replayRec.Code != http.StatusAccepted {
		t.Fatalf("replay status = %d, want %d", replayRec.Code, http.StatusAccepted)
	}
	var replayReceipt transition.CommandReceipt
	if err := json.NewDecoder(replayRec.Body).Decode(&replayReceipt); err != nil {
		t.Fatalf("decode replay receipt: %v", err)
	}
	if replayReceipt.MachineKind != transition.MachineProblemLine || replayReceipt.CommandKind != string(transition.CommandProblemLineScheduleReplay) {
		t.Fatalf("unexpected replay receipt %+v", replayReceipt)
	}
	foundReplayEvent := false
	foundEvalReceipt := false
	foundEvalEffect := false
	for _, item := range store.ListDomainEvents() {
		if item.EventKind == "problem_line_replay_scheduled" {
			foundReplayEvent = true
		}
	}
	if _, ok := store.GetCommandReceipt(replayReceipt.CommandID + ":evaluate"); ok {
		foundEvalReceipt = true
	}
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind == transition.MachineProblemLine && effect.AggregateID == traceID && effect.EffectKind == transition.EffectInvokeRunner && effect.Status == transition.EffectQueued {
			foundEvalEffect = true
			break
		}
	}
	if !foundReplayEvent {
		t.Fatal("expected replay route to record a problem-line domain event")
	}
	if !foundEvalReceipt {
		t.Fatal("expected replay route to emit a follow-on eval command receipt")
	}
	if !foundEvalEffect {
		t.Fatal("expected replay route to queue a problem-line eval runner effect")
	}
}

func TestRouterProposalDetailAndRuntimeEndpoints(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":                  "ok",
			"role":                    "eval",
			"backend":                 "hermes-aiagent",
			"provider":                "openai",
			"model":                   "openai/gpt-5.4",
			"provider_model":          "gpt-5.4",
			"api_mode":                "codex_responses",
			"reasoning_effort":        "xhigh",
			"available":               true,
			"hermes_available":        true,
			"openai_configured":       true,
			"honcho_configured":       true,
			"honcho_available":        true,
			"honcho_base_url":         "http://use1-stage-rsi-agent-platform-honcho-api:8000",
			"honcho_workspace":        "rsi-stage",
			"honcho_environment":      "stage",
			"honcho_recall_mode":      "hybrid",
			"honcho_write_frequency":  "async",
			"honcho_session_strategy": "global",
			"honcho_ai_peer":          "rsi:stage:eval",
		})
	}))
	defer runner.Close()
	honcho := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":               "ok",
			"namespace":            "rsi-stage",
			"db_schema":            "honcho",
			"cache_enabled":        true,
			"cache_url_configured": true,
			"deriver": map[string]any{
				"provider":         "openai",
				"model":            "gpt-5.4",
				"reasoning_effort": "xhigh",
			},
			"summary": map[string]any{
				"provider":         "openai",
				"model":            "gpt-5.4",
				"reasoning_effort": "xhigh",
			},
			"dialectic_levels": map[string]any{
				"minimal": map[string]any{
					"provider":               "openai",
					"model":                  "gpt-5.4",
					"reasoning_effort":       "xhigh",
					"thinking_budget_tokens": 0,
				},
			},
		})
	}))
	defer honcho.Close()

	store := storepkg.NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposal")
	}
	proposal := proposals[0]

	approved, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for repo change.",
		ReviewerID: "operator",
	})
	if err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	if strings.TrimSpace(approved.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt after approval, got %+v", approved)
	}
	seedRouterRepoChangeJobViaCommand(t, store, approved, "proposal-list", "job-proposal-list-1", "codex/proposal-test")
	if _, err := store.RecordPRAttempt(improvement.PRAttempt{
		ProposalID:       proposal.ID,
		AttemptID:        approved.CurrentAttemptID,
		Repo:             "rsi-agent-platform",
		BranchName:       "codex/proposal-test",
		PRURL:            "https://github.com/piplabs/rsi-agent-platform/pull/42",
		Status:           string(review.ProposalPROpen),
		ValidationStatus: "pending",
	}); err != nil {
		t.Fatalf("record pr attempt: %v", err)
	}

	cfg := config.Config{
		PublicBaseURL:          "http://example.test",
		RunnerBaseURL:          runner.URL,
		ProdRunnerBaseURL:      runner.URL,
		ProactiveRunnerBaseURL: runner.URL,
		EvalRunnerBaseURL:      runner.URL,
		ProposalRunnerBaseURL:  runner.URL,
		HonchoRuntimeBaseURL:   honcho.URL,
	}
	router := NewRouter(cfg, store)

	proposalListReq := httptest.NewRequest(http.MethodGet, "/api/proposals", nil)
	proposalListRec := httptest.NewRecorder()
	router.ServeHTTP(proposalListRec, proposalListReq)
	if proposalListRec.Code != http.StatusOK {
		t.Fatalf("proposal list status = %d, want %d", proposalListRec.Code, http.StatusOK)
	}

	var proposalListPayload struct {
		Proposals []map[string]any `json:"proposals"`
	}
	if err := json.NewDecoder(proposalListRec.Body).Decode(&proposalListPayload); err != nil {
		t.Fatalf("decode proposal list: %v", err)
	}
	if len(proposalListPayload.Proposals) == 0 {
		t.Fatal("expected at least one proposal summary")
	}
	firstProposal := proposalListPayload.Proposals[0]
	if got, _ := firstProposal["repo_change_status"].(string); got == "" {
		t.Fatal("expected compact repo_change_status on proposal summary")
	}
	if got, _ := firstProposal["pr_status"].(string); got == "" {
		t.Fatal("expected compact pr_status on proposal summary")
	}
	if got, _ := firstProposal["pr_url"].(string); got == "" {
		t.Fatal("expected compact pr_url on proposal summary")
	}
	if _, ok := firstProposal["repo_change_jobs"]; ok {
		t.Fatal("proposal list should not hydrate repo_change_jobs history")
	}
	if _, ok := firstProposal["pr_attempts"]; ok {
		t.Fatal("proposal list should not hydrate pr_attempts history")
	}

	proposalReq := httptest.NewRequest(http.MethodGet, "/api/proposals/"+proposal.ID, nil)
	proposalRec := httptest.NewRecorder()
	router.ServeHTTP(proposalRec, proposalReq)
	if proposalRec.Code != http.StatusOK {
		t.Fatalf("proposal detail status = %d, want %d", proposalRec.Code, http.StatusOK)
	}

	var proposalPayload map[string]any
	if err := json.NewDecoder(proposalRec.Body).Decode(&proposalPayload); err != nil {
		t.Fatalf("decode proposal detail: %v", err)
	}
	if items, ok := proposalPayload["repo_change_jobs"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected repo_change_jobs array with at least one item")
	}
	if items, ok := proposalPayload["pr_attempts"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected pr_attempts array with at least one item")
	}
	if items, ok := proposalPayload["attempts"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected attempts array with at least one item")
	}
	if items, ok := proposalPayload["linked_trace_summaries"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected linked_trace_summaries array with at least one item")
	}
	if _, ok := proposalPayload["action_intents"].([]any); !ok {
		t.Fatal("expected action_intents array in proposal detail")
	}
	if _, ok := proposalPayload["outcomes"].([]any); !ok {
		t.Fatal("expected outcomes array in proposal detail")
	}
	if _, ok := proposalPayload["knowledge_entries"].([]any); !ok {
		t.Fatal("expected knowledge_entries array in proposal detail")
	}
	attempts := proposalPayload["attempts"].([]any)
	attempt, _ := attempts[0].(map[string]any)
	attemptID, _ := attempt["id"].(string)
	if attemptID == "" {
		t.Fatal("expected attempt id in proposal detail")
	}
	attemptReq := httptest.NewRequest(http.MethodGet, "/api/proposals/"+proposal.ID+"/attempts/"+attemptID, nil)
	attemptRec := httptest.NewRecorder()
	router.ServeHTTP(attemptRec, attemptReq)
	if attemptRec.Code != http.StatusOK {
		t.Fatalf("attempt detail status = %d, want %d", attemptRec.Code, http.StatusOK)
	}

	runtimeReq := httptest.NewRequest(http.MethodGet, "/api/runtime", nil)
	runtimeRec := httptest.NewRecorder()
	router.ServeHTTP(runtimeRec, runtimeReq)
	if runtimeRec.Code != http.StatusOK {
		t.Fatalf("runtime status = %d, want %d", runtimeRec.Code, http.StatusOK)
	}

	var runtimePayload struct {
		Roles  []map[string]any `json:"roles"`
		Honcho map[string]any   `json:"honcho"`
	}
	if err := json.NewDecoder(runtimeRec.Body).Decode(&runtimePayload); err != nil {
		t.Fatalf("decode runtime payload: %v", err)
	}
	if len(runtimePayload.Roles) != 4 {
		t.Fatalf("expected 4 runtime roles, got %d", len(runtimePayload.Roles))
	}
	if got, _ := runtimePayload.Honcho["status"].(string); got != "ok" {
		t.Fatalf("expected honcho runtime ok, got %q", got)
	}
	if got, _ := runtimePayload.Honcho["namespace"].(string); got != "rsi-stage" {
		t.Fatalf("expected honcho namespace rsi-stage, got %q", got)
	}
	deriver, ok := runtimePayload.Honcho["deriver"].(map[string]any)
	if !ok {
		t.Fatal("expected honcho deriver payload")
	}
	if got, _ := deriver["model"].(string); got != "gpt-5.4" {
		t.Fatalf("expected honcho deriver model gpt-5.4, got %q", got)
	}
	if got, _ := deriver["reasoning_effort"].(string); got != "xhigh" {
		t.Fatalf("expected honcho deriver reasoning xhigh, got %q", got)
	}
	for _, role := range runtimePayload.Roles {
		if got, _ := role["model"].(string); got != "openai/gpt-5.4" {
			t.Fatalf("expected model openai/gpt-5.4, got %q", got)
		}
		if got, _ := role["reasoning_effort"].(string); got != "xhigh" {
			t.Fatalf("expected reasoning effort xhigh, got %q", got)
		}
		if got, _ := role["api_mode"].(string); got != "codex_responses" {
			t.Fatalf("expected api_mode codex_responses, got %q", got)
		}
		if got, _ := role["honcho_workspace"].(string); got != "rsi-stage" {
			t.Fatalf("expected honcho_workspace rsi-stage, got %q", got)
		}
		if got, _ := role["honcho_recall_mode"].(string); got != "hybrid" {
			t.Fatalf("expected honcho_recall_mode hybrid, got %q", got)
		}
	}
}

func TestRouterProposalRetryEndpoint(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]

	approved, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "operator",
	})
	if err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	if strings.TrimSpace(approved.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt after approval, got %+v", approved)
	}
	if _, _, err := storepkg.AdvanceProposalToFailedValidationForTesting(store, proposal.ID, time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceProposalToFailedValidationForTesting() error = %v", err)
	}

	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	req := httptest.NewRequest(http.MethodPost, "/api/proposals/"+proposal.ID+"/commands", strings.NewReader(`{"command_kind":"proposal_retry_attempt","actor":"ui-operator"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("retry proposal status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	var receipt transition.CommandReceipt
	if err := json.NewDecoder(rec.Body).Decode(&receipt); err != nil {
		t.Fatalf("decode retry response: %v", err)
	}
	if receipt.ResultRef != proposal.ID {
		t.Fatalf("expected retry command result ref %s, got %s", proposal.ID, receipt.ResultRef)
	}
	updated, ok := findProposalByID(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after retry", proposal.ID)
	}
	if strings.TrimSpace(updated.CurrentAttemptID) == "" {
		t.Fatalf("expected retry to materialize a new current attempt, got %+v", updated)
	}
	foundRetryEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != updated.CurrentAttemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			foundRetryEffect = true
		}
	}
	if !foundRetryEffect {
		t.Fatalf("expected queued retry bootstrap effect for attempt %s", updated.CurrentAttemptID)
	}
}

func TestRouterProposalStopEndpoint(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}

	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	req := httptest.NewRequest(http.MethodPost, "/api/proposals/"+proposal.ID+"/commands", strings.NewReader(`{"command_kind":"proposal_stop_line","actor":"ui-operator","payload":{"rationale":"Stop this remediation line."}}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("stop proposal status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	var receipt transition.CommandReceipt
	if err := json.NewDecoder(rec.Body).Decode(&receipt); err != nil {
		t.Fatalf("decode stop receipt: %v", err)
	}
	updated, ok := findProposalByID(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatal("expected proposal to exist")
	}
	if updated.Status != review.ProposalCanceled {
		t.Fatalf("expected canceled status, got %q", updated.Status)
	}
}

func TestRouterCommandEndpointsEvaluateTraceAndUpdateSettings(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	evalReq := httptest.NewRequest(http.MethodPost, "/api/problem-lines/"+traceID+"/commands", strings.NewReader(`{"command_kind":"problem_line_evaluate_trace","actor":"ui-operator","payload":{"trigger":"manual"}}`))
	evalReq.Header.Set("Content-Type", "application/json")
	evalRec := httptest.NewRecorder()
	router.ServeHTTP(evalRec, evalReq)
	if evalRec.Code != http.StatusAccepted {
		t.Fatalf("evaluate command status = %d, want %d", evalRec.Code, http.StatusAccepted)
	}

	settingsReq := httptest.NewRequest(http.MethodPost, "/api/settings/commands", strings.NewReader(`{"command_kind":"settings_update","actor":"ui-operator","payload":{"active_proposal_cap":1}}`))
	settingsReq.Header.Set("Content-Type", "application/json")
	settingsRec := httptest.NewRecorder()
	router.ServeHTTP(settingsRec, settingsReq)
	if settingsRec.Code != http.StatusAccepted {
		t.Fatalf("settings command status = %d, want %d", settingsRec.Code, http.StatusAccepted)
	}
	if got := store.GetSettings().ActiveProposalCap; got != 1 {
		t.Fatalf("expected proposal cap 1, got %d", got)
	}
	if len(store.ListEvalRuns()) == 0 {
		t.Fatal("expected eval run to be created")
	}
}

func TestLegacyMutationRoutesAbsent(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	proposalID := store.ListProposals()[0].ID
	knowledgeEntry := seedRouterKnowledgeEntry(t, store, "knowledge-legacy-route-coverage", knowledge.Entry{
		Tier:       knowledge.TierWorking,
		Kind:       knowledge.KindFact,
		ScopeType:  knowledge.ScopeGlobal,
		Title:      "Legacy route coverage",
		Status:     knowledge.StatusDraft,
		SourceType: knowledge.SourceAgent,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}, nil, "router-legacy-route-knowledge")
	knowledgeID := knowledgeEntry.ID
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	for _, path := range []string{
		"/api/traces/" + traceID + "/evaluate",
		"/api/proposals/promote",
		"/api/proposals/" + proposalID + "/decision",
		"/api/proposals/" + proposalID + "/retry",
		"/api/proposals/" + proposalID + "/stop",
		"/api/knowledge/" + knowledgeID + "/review",
		"/api/settings",
	} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected %s to be absent, got %d", path, rec.Code)
		}
	}
}

func findProposalByID(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}
