package control

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func TestGitHubWebhookCreatesLinkedOutcome(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposalID := store.ListProposals()[0].ID
	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)

	payload := map[string]any{
		"action": "closed",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   42,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/42",
			"state":    "closed",
			"merged":   true,
			"title":    "RSI proposal " + proposalID + " for rsi-agent-platform",
			"head": map[string]any{
				"ref": "codex/" + proposalID,
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-1")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	outcomes := store.ListOutcomes()
	found := false
	for _, item := range outcomes {
		if item.Source == string(ingestion.SourceGitHub) && item.ProposalID == proposalID {
			found = true
			if item.Verdict != "positive" {
				t.Fatalf("expected positive verdict, got %s", item.Verdict)
			}
			if item.ExternalRef != "https://github.com/piplabs/rsi-agent-platform/pull/42" {
				t.Fatalf("unexpected external ref %s", item.ExternalRef)
			}
		}
	}
	if !found {
		t.Fatal("expected linked github outcome to be recorded")
	}
}

func TestEventAndIngestionRoutesSubmitIngressCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	eventReq := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader([]byte(`{
		"source":"system",
		"source_event_id":"manual-1",
		"dedupe_key":"manual-1",
		"severity":"warning",
		"normalized_problem_statement":"Investigate the latest deploy issue.",
		"workflow_hint":"question"
	}`)))
	eventRec := httptest.NewRecorder()
	router.ServeHTTP(eventRec, eventReq)
	if eventRec.Code != http.StatusCreated {
		t.Fatalf("event status = %d, want %d", eventRec.Code, http.StatusCreated)
	}

	ingestionReq := httptest.NewRequest(http.MethodPost, "/api/ingestions", bytes.NewReader([]byte(`{
		"bot_role":"orchestrator",
		"team_id":"T1",
		"channel_id":"C1",
		"thread_ts":"171000001.000100",
		"user_id":"U1",
		"text":"Can RSI summarize the latest deploy issue?",
		"ts":"171000001.000100"
	}`)))
	ingestionRec := httptest.NewRecorder()
	router.ServeHTTP(ingestionRec, ingestionReq)
	if ingestionRec.Code != http.StatusCreated {
		t.Fatalf("ingestion status = %d, want %d", ingestionRec.Code, http.StatusCreated)
	}

	foundEvent := false
	foundSlack := false
	for _, item := range store.ListDomainEvents() {
		switch item.EventKind {
		case "ingress_event_recorded":
			foundEvent = true
		case "ingress_slack_recorded":
			foundSlack = true
		}
	}
	if !foundEvent || !foundSlack {
		t.Fatalf("expected both ingress domain events, foundEvent=%t foundSlack=%t", foundEvent, foundSlack)
	}
}

func TestGitHubWebhookRejectsInvalidSignature(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-2")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized response, got %d", rec.Code)
	}
}

func TestGitHubWebhookIgnoresUnknownProposalWithoutWorkflow(t *testing.T) {
	store := storepkg.NewMemoryStore()
	initialOutcomes := len(store.ListOutcomes())
	initialTraces := len(store.ListTraces())
	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)

	payload := map[string]any{
		"action": "opened",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   43,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/43",
			"state":    "open",
			"merged":   false,
			"title":    "RSI proposal proposal-does-not-exist for rsi-agent-platform",
			"head": map[string]any{
				"ref": "codex/proposal-does-not-exist",
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-3")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	if len(store.ListOutcomes()) != initialOutcomes {
		t.Fatalf("expected no new linked outcomes, got %d -> %d", initialOutcomes, len(store.ListOutcomes()))
	}
	if len(store.ListTraces()) != initialTraces {
		t.Fatalf("expected no new workflow traces, got %d -> %d", initialTraces, len(store.ListTraces()))
	}
}

func TestThreadPolicyCommandsRouteMutatesViaFormalCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	req := httptest.NewRequest(http.MethodPost, "/api/thread-policies/slack:CENG:171000001.000100/commands", bytes.NewReader([]byte(`{"command_kind":"thread_mute","actor":"ui-operator"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	item, ok := findThreadPolicy(store.ListThreadPolicies(), "slack:CENG:171000001.000100")
	if !ok {
		t.Fatal("expected thread policy to exist")
	}
	if item.State != policy.ThreadStateMuted || !item.Muted {
		t.Fatalf("expected muted thread policy, got %+v", item)
	}
}

func TestLegacyThreadPolicyMutationRoutesAbsent(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	for _, path := range []string{
		"/api/thread-policies/slack:CENG:171000001.000100/mute",
		"/api/thread-policies/slack:CENG:171000001.000100/resume",
	} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected %s to be absent, got %d", path, rec.Code)
		}
	}
}

func TestGitHubWebhookClosedPRQueuesRetryForCurrentAttempt(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for recursive remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	attempt, err := storepkg.SeedChangeAttemptForTesting(store, improvement.ChangeAttempt{
		ProposalID:     proposal.ID,
		CandidateKey:   proposal.CandidateKey,
		AttemptNumber:  1,
		TargetLayer:    proposal.TargetLayer,
		TargetKind:     proposal.TargetKind,
		TargetRef:      proposal.TargetRef,
		Trigger:        improvement.AttemptTriggerProposalApproved,
		State:          improvement.AttemptStatePROpen,
		AttemptTraceID: proposal.TraceID,
		BranchName:     "codex/" + proposal.ID + "/attempt-01",
	})
	if err != nil {
		t.Fatalf("upsert change attempt: %v", err)
	}

	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)
	payload := map[string]any{
		"action": "closed",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   99,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/99",
			"state":    "closed",
			"merged":   false,
			"title":    "RSI proposal " + proposal.ID + " attempt " + attempt.ID + " for rsi-agent-platform",
			"head": map[string]any{
				"ref": attempt.BranchName,
				"sha": "deadbeef",
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-retry")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	updatedAttempt, ok := store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatal("expected updated attempt")
	}
	if updatedAttempt.FailureClass != "closed_unmerged" {
		t.Fatalf("expected closed_unmerged failure class, got %q", updatedAttempt.FailureClass)
	}
	if updatedAttempt.RetryDecision != "auto_retry" {
		t.Fatalf("expected auto_retry decision, got %q", updatedAttempt.RetryDecision)
	}
	if proposal, ok = findProposalByID(store.ListProposals(), proposal.ID); !ok {
		t.Fatal("expected proposal after webhook")
	}
	if proposal.Status != review.ProposalApproved {
		t.Fatalf("expected proposal to remain approved for retry, got %s", proposal.Status)
	}
	if proposal.CurrentAttemptID == "" || proposal.CurrentAttemptID == attempt.ID {
		t.Fatalf("expected retry to materialize a successor attempt, got %q", proposal.CurrentAttemptID)
	}
	foundRetry := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != proposal.CurrentAttemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			foundRetry = true
		default:
			t.Fatalf("unexpected retry bootstrap effect %s", effect.EffectKind)
		}
	}
	if !foundRetry {
		t.Fatal("expected queued retry effect for the successor attempt")
	}
}

func TestGitHubWebhookOpenedPRRoutesProposalThroughCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for recursive remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	attempt, err := storepkg.SeedChangeAttemptForTesting(store, improvement.ChangeAttempt{
		ProposalID:     proposal.ID,
		CandidateKey:   proposal.CandidateKey,
		AttemptNumber:  1,
		TargetLayer:    proposal.TargetLayer,
		TargetKind:     proposal.TargetKind,
		TargetRef:      proposal.TargetRef,
		Trigger:        improvement.AttemptTriggerProposalApproved,
		State:          improvement.AttemptStateValidationRunning,
		AttemptTraceID: proposal.TraceID,
		BranchName:     "codex/" + proposal.ID + "/attempt-01",
	})
	if err != nil {
		t.Fatalf("upsert change attempt: %v", err)
	}

	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)
	payload := map[string]any{
		"action": "opened",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   77,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/77",
			"state":    "open",
			"merged":   false,
			"title":    "RSI proposal " + proposal.ID + " attempt " + attempt.ID + " for rsi-agent-platform",
			"head": map[string]any{
				"ref": attempt.BranchName,
				"sha": "abc123",
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-open")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	updatedAttempt, ok := store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatal("expected updated attempt")
	}
	if updatedAttempt.State != improvement.AttemptStateCIObserving {
		t.Fatalf("expected attempt state ci_observing, got %s", updatedAttempt.State)
	}
	if updatedAttempt.PRURL != "https://github.com/piplabs/rsi-agent-platform/pull/77" {
		t.Fatalf("expected pr url to persist, got %q", updatedAttempt.PRURL)
	}
	if proposal, ok = findProposalByID(store.ListProposals(), proposal.ID); !ok {
		t.Fatal("expected proposal after webhook")
	}
	if proposal.Status != review.ProposalPROpen {
		t.Fatalf("expected proposal to move to pr_open, got %s", proposal.Status)
	}
	receipt, ok := store.GetCommandReceipt("cmd-github-proposal:" + proposal.ID + ":" + string(transition.CommandProposalMarkPROpen) + ":delivery-open")
	if !ok {
		t.Fatal("expected proposal command receipt for pr_open outcome")
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected pr_open command to advance, got %s", receipt.DecisionKind)
	}
}

func TestGitHubWebhookMergedPRRoutesProposalThroughCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for recursive remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	attempt, err := storepkg.SeedChangeAttemptForTesting(store, improvement.ChangeAttempt{
		ProposalID:     proposal.ID,
		CandidateKey:   proposal.CandidateKey,
		AttemptNumber:  1,
		TargetLayer:    proposal.TargetLayer,
		TargetKind:     proposal.TargetKind,
		TargetRef:      proposal.TargetRef,
		Trigger:        improvement.AttemptTriggerProposalApproved,
		State:          improvement.AttemptStatePROpen,
		AttemptTraceID: proposal.TraceID,
		BranchName:     "codex/" + proposal.ID + "/attempt-01",
	})
	if err != nil {
		t.Fatalf("upsert change attempt: %v", err)
	}

	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)
	payload := map[string]any{
		"action": "closed",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   78,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/78",
			"state":    "closed",
			"merged":   true,
			"title":    "RSI proposal " + proposal.ID + " attempt " + attempt.ID + " for rsi-agent-platform",
			"head": map[string]any{
				"ref": attempt.BranchName,
				"sha": "def456",
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-merged")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	updatedAttempt, ok := store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatal("expected updated attempt")
	}
	if updatedAttempt.State != improvement.AttemptStateMerged {
		t.Fatalf("expected attempt state merged, got %s", updatedAttempt.State)
	}
	if proposal, ok = findProposalByID(store.ListProposals(), proposal.ID); !ok {
		t.Fatal("expected proposal after webhook")
	}
	if proposal.Status != review.ProposalMerged {
		t.Fatalf("expected proposal to move to merged, got %s", proposal.Status)
	}
	receipt, ok := store.GetCommandReceipt("cmd-github-proposal:" + proposal.ID + ":" + string(transition.CommandProposalMarkMerged) + ":delivery-merged")
	if !ok {
		t.Fatal("expected proposal command receipt for merged outcome")
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected merged command to advance, got %s", receipt.DecisionKind)
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

func findThreadPolicy(items []policy.ThreadPolicy, threadKey string) (policy.ThreadPolicy, bool) {
	for _, item := range items {
		if item.ThreadKey == threadKey {
			return item, true
		}
	}
	return policy.ThreadPolicy{}, false
}

func signGitHubWebhook(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
