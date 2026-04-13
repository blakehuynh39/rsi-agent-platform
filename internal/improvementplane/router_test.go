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
)

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
	intent, err := store.UpsertActionIntent(action.Intent{
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        trace.Summary.TraceID,
		Kind:           action.KindSlackPost,
		Status:         action.StatusSucceeded,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		t.Fatalf("upsert action intent: %v", err)
	}
	if _, err := store.RecordActionResult(action.Result{
		ActionIntentID: intent.ID,
		Executor:       "tool-gateway",
		Provider:       "slack",
		Status:         action.StatusSucceeded,
		StartedAt:      now,
		CompletedAt:    now,
	}); err != nil {
		t.Fatalf("record action result: %v", err)
	}
	if _, err := store.RecordOutcome(outcome.Record{
		Source:         "operator",
		RecordedBy:     "tester",
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        trace.Summary.TraceID,
		OutcomeType:    outcome.TypeAnswerQuality,
		Verdict:        outcome.VerdictPositive,
		Score:          1,
		Summary:        "Trace resolved the question well.",
		RecordedAt:     now,
	}); err != nil {
		t.Fatalf("record outcome: %v", err)
	}
	entry, err := store.UpsertKnowledgeEntry(knowledge.Entry{
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
	})
	if err != nil {
		t.Fatalf("upsert knowledge entry: %v", err)
	}
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

func TestRouterProposalDetailAndRuntimeEndpoints(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":            "ok",
			"role":              "eval",
			"backend":           "hermes-aiagent",
			"provider":          "openai",
			"model":             "openai/gpt-5.4",
			"provider_model":    "gpt-5.4",
			"api_mode":          "codex_responses",
			"reasoning_effort":  "xhigh",
			"available":         true,
			"hermes_available":  true,
			"openai_configured": true,
		})
	}))
	defer runner.Close()

	store := storepkg.NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposal")
	}
	proposal := proposals[0]

	if _, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for repo change.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	if _, err := store.MaterializeApprovedProposal(proposal.ID, "operator"); err != nil {
		t.Fatalf("materialize approved proposal: %v", err)
	}
	if _, err := store.RecordPRAttempt(improvement.PRAttempt{
		ProposalID:       proposal.ID,
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
	}
	router := NewRouter(cfg, store)

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

	runtimeReq := httptest.NewRequest(http.MethodGet, "/api/runtime", nil)
	runtimeRec := httptest.NewRecorder()
	router.ServeHTTP(runtimeRec, runtimeReq)
	if runtimeRec.Code != http.StatusOK {
		t.Fatalf("runtime status = %d, want %d", runtimeRec.Code, http.StatusOK)
	}

	var runtimePayload struct {
		Roles []map[string]any `json:"roles"`
	}
	if err := json.NewDecoder(runtimeRec.Body).Decode(&runtimePayload); err != nil {
		t.Fatalf("decode runtime payload: %v", err)
	}
	if len(runtimePayload.Roles) != 4 {
		t.Fatalf("expected 4 runtime roles, got %d", len(runtimePayload.Roles))
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
	}
}

func TestRouterProposalRetryEndpoint(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]

	if _, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	job, err := store.MaterializeApprovedProposal(proposal.ID, "operator")
	if err != nil {
		t.Fatalf("materialize approved proposal: %v", err)
	}
	if _, err := store.UpdateRepoChangeJobStatus(job.ID, string(review.ProposalFailedValidation)); err != nil {
		t.Fatalf("update repo change job status: %v", err)
	}
	if _, err := store.UpdateProposalStatus(proposal.ID, review.ProposalFailedValidation); err != nil {
		t.Fatalf("update proposal status: %v", err)
	}

	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	req := httptest.NewRequest(http.MethodPost, "/api/proposals/"+proposal.ID+"/retry", strings.NewReader(`{"requested_by":"ui-operator"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("retry proposal status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	var item map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
		t.Fatalf("decode retry response: %v", err)
	}
	if got, _ := item["queue"].(string); got != "sandbox" {
		t.Fatalf("expected sandbox retry queue, got %q", got)
	}
}
