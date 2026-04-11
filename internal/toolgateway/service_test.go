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
