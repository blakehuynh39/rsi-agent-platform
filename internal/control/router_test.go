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
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
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
	initialWorkItems := len(store.ListWorkItems())
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
	if len(store.ListWorkItems()) != initialWorkItems {
		t.Fatalf("expected no new workflow work items, got %d -> %d", initialWorkItems, len(store.ListWorkItems()))
	}
}

func TestGitHubWebhookClosedPRQueuesRetryForCurrentAttempt(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for recursive remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	attempt, err := store.UpsertChangeAttempt(improvement.ChangeAttempt{
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
	foundRetry := false
	for _, item := range store.ListWorkItems() {
		if item.Queue == "proposal" && item.Kind == "approved_proposal" && item.ProposalID == proposal.ID {
			foundRetry = true
			break
		}
	}
	if !foundRetry {
		t.Fatal("expected approved_proposal retry work item to be queued")
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

func signGitHubWebhook(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
