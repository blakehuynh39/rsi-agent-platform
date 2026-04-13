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
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
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

func signGitHubWebhook(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
