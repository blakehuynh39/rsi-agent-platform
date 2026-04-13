package toolgateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestGitHubCreatePRUsesExternalAPI(t *testing.T) {
	var seenAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/repos/piplabs/rsi-agent-platform/pulls" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
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
	}))
	defer server.Close()

	service := NewService(config.Config{
		GitHubToken:      "test-token",
		GitHubOwner:      "piplabs",
		GitHubAPIBaseURL: server.URL,
		DefaultRepo:      "rsi-agent-platform",
	}, storepkg.NewMemoryStore())

	result := service.Execute("github.create_pr", map[string]interface{}{
		"repo":        "rsi-agent-platform",
		"branch_name": "codex/test-branch",
		"base_ref":    "main",
		"title":       "Test PR",
		"body":        "Draft body",
	})

	if seenAuth != "Bearer test-token" {
		t.Fatalf("unexpected auth header %q", seenAuth)
	}
	if result.Output["pr_url"] != "https://github.com/piplabs/rsi-agent-platform/pull/123" {
		t.Fatalf("unexpected pr url %#v", result.Output)
	}
}

func TestGitHubRepoActivityUsesExternalAPI(t *testing.T) {
	var seenAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/repos/piplabs/depin-backend/commits":
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
		GitHubToken:      "test-token",
		GitHubOwner:      "piplabs",
		GitHubAPIBaseURL: server.URL,
		DefaultRepo:      "depin-backend",
	}, storepkg.NewMemoryStore())

	result := service.Execute("github.repo_activity", map[string]interface{}{
		"repo":  "depin-backend",
		"since": "2026-04-05T00:00:00Z",
		"until": "2026-04-12T00:00:00Z",
	})

	if seenAuth != "Bearer test-token" {
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
