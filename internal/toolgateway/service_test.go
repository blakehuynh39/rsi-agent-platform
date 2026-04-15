package toolgateway

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestGitHubCreatePRUsesExternalAPI(t *testing.T) {
	privateKey := testGitHubAppPrivateKey(t)
	var (
		seenAppAuth string
		seenPRAuth  string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/installations/456/access_tokens":
			seenAppAuth = r.Header.Get("Authorization")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token":      "installation-token",
				"expires_at": "2026-04-14T00:00:00Z",
			})
		case "/repos/piplabs/rsi-agent-platform/pulls":
			seenPRAuth = r.Header.Get("Authorization")
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if draft, ok := body["draft"].(bool); !ok || !draft {
				t.Fatalf("expected draft PR payload, got %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/123",
				"number":   123,
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(config.Config{
		GitHubAppID:             "123",
		GitHubAppInstallationID: "456",
		GitHubAppPrivateKey:     privateKey,
		GitHubOwner:             "piplabs",
		GitHubAPIBaseURL:        server.URL,
		DefaultRepo:             "rsi-agent-platform",
	}, storepkg.NewMemoryStore())

	result := service.Execute("github.create_pr", map[string]interface{}{
		"repo":        "rsi-agent-platform",
		"branch_name": "codex/test-branch",
		"base_ref":    "main",
		"title":       "Test PR",
		"body":        "Draft body",
	})

	if seenAppAuth == "" || seenPRAuth != "Bearer installation-token" {
		t.Fatalf("unexpected app/pr auth headers app=%q pr=%q", seenAppAuth, seenPRAuth)
	}
	if result.Output["pr_url"] != "https://github.com/piplabs/rsi-agent-platform/pull/123" {
		t.Fatalf("unexpected pr url %#v", result.Output)
	}
}

func TestGitHubCreatePRRequiresAppConfig(t *testing.T) {
	service := NewService(config.Config{
		GitHubOwner:      "piplabs",
		GitHubAPIBaseURL: "https://api.github.com",
		DefaultRepo:      "rsi-agent-platform",
	}, storepkg.NewMemoryStore())

	result := service.Execute("github.create_pr", map[string]interface{}{
		"repo":        "rsi-agent-platform",
		"branch_name": "codex/test-branch",
		"base_ref":    "main",
		"title":       "Test PR",
		"body":        "Draft body",
	})

	if result.Status != "blocked" {
		t.Fatalf("expected blocked status, got %s", result.Status)
	}
	if result.Output["error"] == nil {
		t.Fatalf("unexpected output %#v", result.Output)
	}
}

func TestUnknownToolIsRejectedWithoutStoreFallback(t *testing.T) {
	service := NewService(config.Config{
		DefaultRepo: "rsi-agent-platform",
	}, storepkg.NewMemoryStore())

	result := service.Execute("github.legacy_fallback", map[string]interface{}{
		"proposal_id": "proposal-123",
	})

	if result.Available {
		t.Fatalf("expected unknown tool to be unavailable, got %+v", result)
	}
	if result.Status != "blocked" {
		t.Fatalf("expected unavailable status, got %+v", result)
	}
	if got := result.Output["error"]; got != "unknown_tool" {
		t.Fatalf("expected unknown_tool error, got %#v", result.Output)
	}
}

func TestGitHubRepoActivityUsesExternalAPI(t *testing.T) {
	privateKey := testGitHubAppPrivateKey(t)
	var seenAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/installations/456/access_tokens":
			seenAuth = r.Header.Get("Authorization")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token":      "installation-token",
				"expires_at": "2026-04-14T00:00:00Z",
			})
		case "/repos/piplabs/depin-backend/commits":
			if r.Header.Get("Authorization") != "Bearer installation-token" {
				t.Fatalf("unexpected commit auth %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"sha":      "abc123",
					"html_url": "https://github.com/piplabs/depin-backend/commit/abc123",
					"commit": map[string]any{
						"message": "Add progress endpoint",
						"author": map[string]any{
							"name": "blake",
							"date": "2026-04-10T12:00:00Z",
						},
					},
				},
			})
		case "/repos/piplabs/depin-backend/pulls":
			if r.Header.Get("Authorization") != "Bearer installation-token" {
				t.Fatalf("unexpected pull auth %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"number":     42,
					"title":      "Improve API throughput",
					"state":      "closed",
					"html_url":   "https://github.com/piplabs/depin-backend/pull/42",
					"created_at": "2026-04-09T12:00:00Z",
					"merged_at":  "2026-04-10T18:00:00Z",
					"user": map[string]any{
						"login": "blake",
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(config.Config{
		GitHubAppID:             "123",
		GitHubAppInstallationID: "456",
		GitHubAppPrivateKey:     privateKey,
		GitHubOwner:             "piplabs",
		GitHubAPIBaseURL:        server.URL,
		DefaultRepo:             "depin-backend",
	}, storepkg.NewMemoryStore())

	result := service.Execute("github.repo_activity", map[string]interface{}{
		"repo":  "depin-backend",
		"since": "2026-04-05T00:00:00Z",
		"until": "2026-04-12T00:00:00Z",
	})

	if seenAuth == "" {
		t.Fatalf("unexpected auth header %q", seenAuth)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	if len(result.Output["commits"].([]map[string]interface{})) != 1 {
		t.Fatalf("expected one commit in output, got %#v", result.Output["commits"])
	}
	if len(result.Output["merged_pull_requests"].([]map[string]interface{})) != 1 {
		t.Fatalf("expected one merged PR in output, got %#v", result.Output["merged_pull_requests"])
	}
}

func TestRepoContextUsesGitHubSearchAndCodeSnippets(t *testing.T) {
	privateKey := testGitHubAppPrivateKey(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/installations/456/access_tokens":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token":      "installation-token",
				"expires_at": "2026-04-14T00:00:00Z",
			})
		case "/repos/piplabs/rsi-agent-platform":
			if r.Header.Get("Authorization") != "Bearer installation-token" {
				t.Fatalf("unexpected repo auth %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"default_branch": "main",
				"html_url":       "https://github.com/piplabs/rsi-agent-platform",
				"description":    "Recursive self-improvement platform",
			})
		case "/search/code":
			if r.Header.Get("Authorization") != "Bearer installation-token" {
				t.Fatalf("unexpected search auth %q", r.Header.Get("Authorization"))
			}
			query := r.URL.Query().Get("q")
			if !strings.Contains(query, "action_result_pkey") {
				t.Fatalf("expected search query to include action_result_pkey, got %q", query)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"name":     "postgres.go",
						"path":     "internal/store/postgres.go",
						"html_url": "https://github.com/piplabs/rsi-agent-platform/blob/main/internal/store/postgres.go",
					},
				},
			})
		case "/repos/piplabs/rsi-agent-platform/contents/internal/store/postgres.go":
			if r.Header.Get("Authorization") != "Bearer installation-token" {
				t.Fatalf("unexpected contents auth %q", r.Header.Get("Authorization"))
			}
			content := "func persistActionResults() {\n  // action_result_pkey collision surfaced here\n  insertActionResult()\n}\n"
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"encoding": "base64",
				"content":  base64.StdEncoding.EncodeToString([]byte(content)),
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(config.Config{
		GitHubAppID:             "123",
		GitHubAppInstallationID: "456",
		GitHubAppPrivateKey:     privateKey,
		GitHubOwner:             "piplabs",
		GitHubAPIBaseURL:        server.URL,
		DefaultRepo:             "rsi-agent-platform",
	}, storepkg.NewMemoryStore())

	result := service.Execute("repo.context", map[string]interface{}{
		"repo":     "rsi-agent-platform",
		"question": "Fix action_result_pkey collision in the shared store",
	})

	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	if got := result.Output["default_branch"]; got != "main" {
		t.Fatalf("default_branch = %#v, want main", got)
	}
	matches, ok := result.Output["matches"].([]map[string]interface{})
	if !ok || len(matches) != 1 {
		t.Fatalf("expected one repo match, got %#v", result.Output["matches"])
	}
	if got := matches[0]["path"]; got != "internal/store/postgres.go" {
		t.Fatalf("unexpected match path %#v", got)
	}
	if snippet := stringValue(matches[0]["snippet"]); !strings.Contains(snippet, "action_result_pkey collision") {
		t.Fatalf("expected snippet to include action_result_pkey collision, got %q", snippet)
	}
}

func TestSlackReplyWithoutTokenIsBlocked(t *testing.T) {
	service := NewService(config.Config{}, storepkg.NewMemoryStore())

	result := service.Execute("slack.reply", map[string]interface{}{
		"channel_id": "D123",
		"thread_ts":  "171000001.000100",
		"body":       "Hello",
	})

	if result.Status != "blocked" {
		t.Fatalf("expected blocked status, got %s", result.Status)
	}
	if result.Available {
		t.Fatal("expected unavailable slack provider when token is missing")
	}
}

func TestRSITraceContextReturnsTraceEvidence(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traces := store.ListTraces()
	if len(traces) == 0 {
		t.Fatal("expected seeded traces")
	}
	service := NewService(config.Config{}, store)

	result := service.Execute("rsi.trace_context", map[string]interface{}{
		"trace_id": traces[0].TraceID,
	})

	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	traceSummary, ok := result.Output["trace"].(events.TraceSummary)
	if !ok {
		t.Fatalf("expected trace summary in output, got %#v", result.Output["trace"])
	}
	if traceSummary.TraceID != traces[0].TraceID {
		t.Fatalf("unexpected trace id %#v", traceSummary)
	}
}

func TestToolExecutionPersistsTraceToolCallEvidence(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traces := store.ListTraces()
	if len(traces) == 0 {
		t.Fatal("expected seeded traces")
	}
	service := NewService(config.Config{}, store)

	result := service.Execute("rsi.trace_context", map[string]interface{}{
		"trace_id": traces[0].TraceID,
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	trace, ok := store.GetTrace(traces[0].TraceID)
	if !ok {
		t.Fatalf("expected trace %s", traces[0].TraceID)
	}
	found := false
	for _, call := range trace.ToolCalls {
		if call.ToolCallID == result.ToolCallID && call.ToolName == "rsi.trace_context" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected trace tool call evidence to be persisted")
	}
}

func TestRSICandidateContextReturnsCandidateAndMemory(t *testing.T) {
	store := storepkg.NewMemoryStore()
	candidates := store.ListCandidates()
	if len(candidates) == 0 {
		t.Fatal("expected seeded candidates")
	}
	service := NewService(config.Config{}, store)

	result := service.Execute("rsi.candidate_context", map[string]interface{}{
		"candidate_key": candidates[0].CandidateKey,
	})

	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	candidate, ok := result.Output["candidate"].(improvement.Candidate)
	if !ok {
		t.Fatalf("expected candidate payload, got %#v", result.Output["candidate"])
	}
	if candidate.CandidateKey != candidates[0].CandidateKey {
		t.Fatalf("unexpected candidate key %#v", result.Output)
	}
}

func testGitHubAppPrivateKey(t *testing.T) string {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}))
}
