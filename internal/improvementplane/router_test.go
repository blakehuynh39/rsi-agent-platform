package improvementplane

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestRouterTraceEndpointsReturnSummaryAndDetailShapes(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	listReq := httptest.NewRequest(http.MethodGet, "/api/traces", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("trace list status = %d, want %d", listRec.Code, http.StatusOK)
	}

	var listPayload struct {
		Traces []map[string]any `json:"traces"`
	}
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode trace list: %v", err)
	}
	if len(listPayload.Traces) == 0 {
		t.Fatal("expected at least one trace in summary list")
	}
	if _, ok := listPayload.Traces[0]["trace_id"]; !ok {
		t.Fatal("expected trace list item to include trace_id")
	}
	if _, ok := listPayload.Traces[0]["events"]; ok {
		t.Fatal("trace list should not include full event detail payloads")
	}

	traceID, _ := listPayload.Traces[0]["trace_id"].(string)
	if traceID == "" {
		t.Fatal("expected non-empty trace_id from trace list")
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID, nil)
	detailRec := httptest.NewRecorder()
	router.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("trace detail status = %d, want %d", detailRec.Code, http.StatusOK)
	}

	var detailPayload map[string]any
	if err := json.NewDecoder(detailRec.Body).Decode(&detailPayload); err != nil {
		t.Fatalf("decode trace detail: %v", err)
	}
	if _, ok := detailPayload["trace"].(map[string]any); !ok {
		t.Fatal("expected trace detail payload to include trace object")
	}
	if _, ok := detailPayload["linked_eval_runs"].([]any); !ok {
		t.Fatal("expected linked_eval_runs to be a JSON array")
	}
	if _, ok := detailPayload["linked_proposals"].([]any); !ok {
		t.Fatal("expected linked_proposals to be a JSON array")
	}
	if _, ok := detailPayload["ratings"].([]any); !ok {
		t.Fatal("expected ratings to be a JSON array")
	}
	if _, ok := detailPayload["improvement_notes"].([]any); !ok {
		t.Fatal("expected improvement_notes to be a JSON array")
	}
	if _, ok := detailPayload["judgments_by_eval_run"].(map[string]any); !ok {
		t.Fatal("expected judgments_by_eval_run to be a JSON object")
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
