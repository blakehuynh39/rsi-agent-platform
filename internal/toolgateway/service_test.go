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
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
	slackapi "github.com/slack-go/slack"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
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

func TestUnknownToolIsRejectedWithoutFallback(t *testing.T) {
	service := NewService(config.Config{
		DefaultRepo: "rsi-agent-platform",
	}, storepkg.NewMemoryStore())

	result := service.Execute("github.unknown_tool", map[string]interface{}{
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

func TestRepoContextWithoutGitHubAuthIsBlocked(t *testing.T) {
	service := NewService(config.Config{
		DefaultRepo: "rsi-agent-platform",
	}, storepkg.NewMemoryStore())

	result := service.Execute("repo.context", map[string]interface{}{
		"repo":     "rsi-agent-platform",
		"question": "Fix action_result_pkey collision in the shared store",
	})

	if result.Status != "blocked" {
		t.Fatalf("expected blocked status, got %s %#v", result.Status, result.Output)
	}
	if result.Available {
		t.Fatalf("expected repo context to be unavailable, got %+v", result)
	}
	if result.ApprovalState != "provider_unavailable" {
		t.Fatalf("expected provider_unavailable approval state, got %#v", result)
	}
}

func TestRSIRuntimeConfigReturnsSanitizedConfig(t *testing.T) {
	service := NewService(config.Config{
		Environment:            "stage",
		DefaultRepo:            "rsi-agent-platform",
		AllowedTargetRepos:     []string{"rsi-agent-platform", "depin-backend"},
		DefaultProposalCap:     2,
		ToolGatewayBaseURL:     "http://tool-gateway.internal:8080",
		HonchoRuntimeBaseURL:   "http://honcho.internal:8000",
		ProdRunnerBaseURL:      "http://runner-prod.internal:8090",
		EvalRunnerBaseURL:      "http://runner-eval.internal:8090",
		ProposalRunnerBaseURL:  "http://runner-proposal.internal:8090",
		ProdRunnerTimeout:      60 * time.Second,
		EvalRunnerTimeout:      330 * time.Second,
		ProposalRunnerTimeout:  450 * time.Second,
		ProactiveRunnerTimeout: 60 * time.Second,
	}, storepkg.NewMemoryStore())

	result := service.Execute("rsi.runtime_config", map[string]interface{}{})

	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s", result.Status)
	}
	if got := result.Output["default_repo"]; got != "rsi-agent-platform" {
		t.Fatalf("unexpected default repo %#v", got)
	}
	runnerURLs, ok := result.Output["runner_urls"].(map[string]string)
	if !ok {
		t.Fatalf("expected runner_urls map[string]string, got %#v", result.Output["runner_urls"])
	}
	if runnerURLs["proposal"] != "http://runner-proposal.internal:8090" {
		t.Fatalf("unexpected proposal runner url %#v", runnerURLs)
	}
	taskTimeouts, ok := result.Output["runner_task_timeouts_seconds"].(map[string]int)
	if !ok {
		t.Fatalf("expected runner_task_timeouts_seconds map[string]int, got %#v", result.Output["runner_task_timeouts_seconds"])
	}
	if taskTimeouts["proposal"] != 420 {
		t.Fatalf("unexpected proposal task timeout %#v", taskTimeouts)
	}
}

func TestRSIActionChainReturnsIntentsResultsAndOutcomes(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: "intent-1",
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   "cmd-toolgateway-action-queue",
		OccurredAt:  now,
		Payload: map[string]any{
			"owner_plane":     "improvement",
			"trace_id":        "trace-1",
			"proposal_id":     "proposal-1",
			"attempt_id":      "attempt-1",
			"kind":            string(action.KindDraftPROpen),
			"idempotency_key": "attempt-1:pr-open",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_queued) error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: "intent-1",
		CommandKind: string(transition.CommandActionStart),
		CommandID:   "cmd-toolgateway-action-start",
		OccurredAt:  now,
		Payload: map[string]any{
			"operation_id": "op-toolgateway-action",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_started) error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: "intent-1",
		CommandKind: string(transition.CommandActionSucceed),
		CommandID:   "cmd-toolgateway-action-succeed",
		OccurredAt:  now.Add(time.Second),
		Payload: map[string]any{
			"operation_id": "op-toolgateway-action",
			"attempt_id":   "attempt-1",
			"executor":     "tool-gateway",
			"started_at":   now,
			"completed_at": now.Add(time.Second),
			"provider_ref": "result-1",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_succeeded) error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: "trace-1",
		CommandKind: string(transition.CommandProblemLineRecordOutcome),
		CommandID:   "cmd-toolgateway-outcome",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"outcome_id":   "outcome-1",
			"trace_id":     "trace-1",
			"proposal_id":  "proposal-1",
			"attempt_id":   "attempt-1",
			"outcome_type": string(outcome.TypeProposalEffectiveness),
			"verdict":      string(outcome.VerdictPositive),
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(problem_line_record_outcome) error = %v", err)
	}

	service := NewService(config.Config{DefaultRepo: "rsi-agent-platform"}, store)
	result := service.Execute("rsi.action_chain", map[string]interface{}{
		"trace_id":    "trace-1",
		"proposal_id": "proposal-1",
		"attempt_id":  "attempt-1",
	})

	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s", result.Status)
	}
	if len(result.Output["action_intents"].([]interface{})) != 1 {
		t.Fatalf("expected one action intent, got %#v", result.Output["action_intents"])
	}
	if len(result.Output["action_results"].([]interface{})) != 1 {
		t.Fatalf("expected one action result, got %#v", result.Output["action_results"])
	}
	if len(result.Output["outcomes"].([]interface{})) != 1 {
		t.Fatalf("expected one outcome, got %#v", result.Output["outcomes"])
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

func TestSlackHistoryUsesChannelHistoryForProgressQuestion(t *testing.T) {
	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("channel"); got != "C123" {
			t.Fatalf("expected channel C123, got %q", got)
		}
		if got := r.Form.Get("limit"); got != "25" {
			t.Fatalf("expected default limit 25, got %q", got)
		}
		switch r.URL.Path {
		case "/conversations.history":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":       true,
				"has_more": false,
				"messages": []map[string]any{
					{"type": "message", "user": "U123", "text": "Raised the control-plane budget and added Slack reads.", "ts": "171000001.000100"},
				},
				"response_metadata": map[string]any{"next_cursor": ""},
			})
		default:
			t.Fatalf("unexpected slack path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(config.Config{
		SlackBotToken:          "xoxb-test",
		AllowedSlackChannelIDs: []string{"C123"},
	}, storepkg.NewMemoryStore())
	service.slackClient = slackapi.New("xoxb-test", slackapi.OptionAPIURL(server.URL+"/"))

	result := service.Execute("slack.history", map[string]interface{}{
		"channel_id": "C123",
		"thread_ts":  "171000001.000100",
		"question":   "How did depin-backend progress in the last week?",
	})

	if seenPath != "/conversations.history" {
		t.Fatalf("expected conversations.history, got %s", seenPath)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	if got := result.Output["scope"]; got != "channel" {
		t.Fatalf("expected channel scope, got %#v", got)
	}
	messages, ok := result.Output["messages"].([]map[string]interface{})
	if !ok || len(messages) != 1 {
		t.Fatalf("expected one slack message, got %#v", result.Output["messages"])
	}
}

func TestSlackHistoryUsesThreadRepliesForConversationQuestion(t *testing.T) {
	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("ts"); got != "171000001.000100" {
			t.Fatalf("expected thread ts 171000001.000100, got %q", got)
		}
		switch r.URL.Path {
		case "/conversations.replies":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":       true,
				"has_more": false,
				"messages": []map[string]any{
					{"type": "message", "user": "U123", "text": "We already wired the Slack app into this thread.", "ts": "171000001.000100", "thread_ts": "171000001.000100"},
				},
				"response_metadata": map[string]any{"next_cursor": ""},
			})
		default:
			t.Fatalf("unexpected slack path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(config.Config{
		SlackBotToken:          "xoxb-test",
		AllowedSlackChannelIDs: []string{"C123"},
	}, storepkg.NewMemoryStore())
	service.slackClient = slackapi.New("xoxb-test", slackapi.OptionAPIURL(server.URL+"/"))

	result := service.Execute("slack.history", map[string]interface{}{
		"channel_id": "C123",
		"thread_ts":  "171000001.000100",
		"question":   "What did we say in the latest convo?",
	})

	if seenPath != "/conversations.replies" {
		t.Fatalf("expected conversations.replies, got %s", seenPath)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	if got := result.Output["scope"]; got != "thread" {
		t.Fatalf("expected thread scope, got %#v", got)
	}
}

func TestSlackHistoryAllowsMentionedChannelWhenMentionsOnlyConfigured(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineIngress,
		AggregateID: "slack:171000001.000100",
		CommandKind: string(transition.CommandIngressRecordSlack),
		CommandID:   "cmd-toolgateway-slack-mentions-only",
		Actor:       "tester",
		OccurredAt:  now,
		Payload: map[string]any{
			"bot_role":   "orchestrator",
			"team_id":    "T123",
			"channel_id": "C123",
			"thread_ts":  "171000001.000100",
			"user_id":    "U123",
			"text":       "What did we decide in <#C999> for numo?",
			"ts":         "171000001.000100",
			"created_at": now,
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(slack ingress) error = %v", err)
	}
	ingestionID := receipt.ResultRef
	traceID := ""
	for _, workflow := range store.ListWorkflows() {
		if workflow.IngestionID == ingestionID {
			traceID = workflow.TraceID
			break
		}
	}
	if traceID == "" {
		t.Fatal("expected trace for ingested slack conversation")
	}

	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("channel"); got != "C999" {
			t.Fatalf("expected channel C999, got %q", got)
		}
		switch r.URL.Path {
		case "/conversations.history":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":       true,
				"has_more": false,
				"messages": []map[string]any{
					{"type": "message", "user": "U999", "text": "Numo discussed here.", "ts": "171000001.000200"},
				},
				"response_metadata": map[string]any{"next_cursor": ""},
			})
		default:
			t.Fatalf("unexpected slack path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(config.Config{
		SlackBotToken:          "xoxb-test",
		AllowedSlackChannelIDs: []string{"MENTIONS_ONLY"},
	}, store)
	service.slackClient = slackapi.New("xoxb-test", slackapi.OptionAPIURL(server.URL+"/"))

	result := service.Execute("slack.history", map[string]interface{}{
		"trace_id":   traceID,
		"channel_id": "C999",
		"scope":      "channel",
	})

	if seenPath != "/conversations.history" {
		t.Fatalf("expected conversations.history, got %s", seenPath)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
}

func TestSlackSearchReturnsFilteredMessages(t *testing.T) {
	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("query"); got != "control plane 5 minute timeout" {
			t.Fatalf("expected query %q, got %q", "control plane 5 minute timeout", got)
		}
		switch r.URL.Path {
		case "/search.messages":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"messages": map[string]any{
					"total": 3,
					"matches": []map[string]any{
						{
							"type":      "message",
							"channel":   map[string]any{"id": "C123", "name": "rsi-platform"},
							"user":      "U123",
							"username":  "blake",
							"ts":        "1775952000.000100",
							"text":      "We bumped the control plane timeout to 5 minutes.",
							"permalink": "https://example.test/C123/p1775952000000100",
						},
						{
							"type":      "message",
							"channel":   map[string]any{"id": "C999", "name": "other"},
							"user":      "U999",
							"username":  "other",
							"ts":        "1775952001.000100",
							"text":      "Wrong channel.",
							"permalink": "https://example.test/C999/p1775952001000100",
						},
						{
							"type":      "message",
							"channel":   map[string]any{"id": "C123", "name": "rsi-platform"},
							"user":      "U123",
							"username":  "blake",
							"ts":        "1775606400.000100",
							"text":      "Outside the requested window.",
							"permalink": "https://example.test/C123/p1775606400000100",
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected slack path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(config.Config{
		SlackBotToken:          "xoxb-test",
		AllowedSlackChannelIDs: []string{"C123", "C999"},
	}, storepkg.NewMemoryStore())
	service.slackClient = slackapi.New("xoxb-test", slackapi.OptionAPIURL(server.URL+"/"))

	result := service.Execute("slack.search", map[string]interface{}{
		"query":       "control plane 5 minute timeout",
		"channel_ids": []interface{}{"C123"},
		"since":       "2026-04-10T00:00:00Z",
		"until":       "2026-04-17T00:00:00Z",
		"limit":       1,
	})

	if seenPath != "/search.messages" {
		t.Fatalf("expected search.messages, got %s", seenPath)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	if got := result.Output["search_total"]; got != 3 {
		t.Fatalf("expected search_total 3, got %#v", got)
	}
	channelIDs, ok := result.Output["channel_ids"].([]string)
	if !ok || len(channelIDs) != 1 || channelIDs[0] != "C123" {
		t.Fatalf("expected output channel_ids [C123], got %#v", result.Output["channel_ids"])
	}
	messages, ok := result.Output["messages"].([]map[string]interface{})
	if !ok || len(messages) != 1 {
		t.Fatalf("expected one filtered slack message, got %#v", result.Output["messages"])
	}
	if got := messages[0]["channel_id"]; got != "C123" {
		t.Fatalf("expected message in channel C123, got %#v", got)
	}
}

func TestRSIRuntimeDeploymentFactsReturnsDeploymentSummary(t *testing.T) {
	service := NewService(config.Config{
		Environment:                      "stage",
		SandboxNamespace:                 "rsi-platform",
		ToolGatewayBaseURL:               "http://tool-gateway.internal:8080",
		HonchoRuntimeBaseURL:             "http://honcho.internal:8000",
		PublicBaseURL:                    "https://staging-rsi-platform.storyprotocol.net",
		ProdRunnerBaseURL:                "http://runner-prod.internal:8090",
		ProactiveRunnerBaseURL:           "http://runner-proactive.internal:8090",
		EvalRunnerBaseURL:                "http://runner-eval.internal:8090",
		ProposalRunnerBaseURL:            "http://runner-proposal.internal:8090",
		ProdRunnerTimeout:                330 * time.Second,
		ProactiveRunnerTimeout:           60 * time.Second,
		EvalRunnerTimeout:                330 * time.Second,
		ProposalRunnerTimeout:            450 * time.Second,
		ProdRunnerTaskTimeout:            300 * time.Second,
		ProactiveRunnerTaskTimeout:       300 * time.Second,
		EvalRunnerTaskTimeout:            300 * time.Second,
		ProposalRunnerTaskTimeout:        420 * time.Second,
		SlackAppIdentity:                 "rsi-stage",
		SlackSocketModeEnabled:           true,
		SlackBotToken:                    "xoxb-test",
		AllowedSlackChannelIDs:           []string{"C123"},
		HermesNativeGovernedToolsEnabled: true,
	}, storepkg.NewMemoryStore())
	service.kubeClient = kubefake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "use1-stage-rsi-agent-platform-control-plane",
			Namespace:         "rsi-platform",
			CreationTimestamp: metav1.NewTime(time.Date(2026, time.April, 17, 0, 0, 0, 0, time.UTC)),
			Labels:            map[string]string{"app.kubernetes.io/name": "control-plane"},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "control-plane", Image: "ghcr.io/piplabs/control-plane:abc123"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 3,
			Replicas:           1,
			ReadyReplicas:      1,
			UpdatedReplicas:    1,
			AvailableReplicas:  1,
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:               appsv1.DeploymentProgressing,
					Status:             corev1.ConditionTrue,
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet is available",
					LastUpdateTime:     metav1.NewTime(time.Date(2026, time.April, 17, 0, 1, 0, 0, time.UTC)),
					LastTransitionTime: metav1.NewTime(time.Date(2026, time.April, 17, 0, 1, 0, 0, time.UTC)),
				},
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
					Reason: "MinimumReplicasAvailable",
				},
			},
		},
	})

	result := service.Execute("rsi.runtime_deployment_facts", map[string]interface{}{
		"service": "control plane",
	})

	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	if got := result.Output["namespace"]; got != "rsi-platform" {
		t.Fatalf("expected namespace rsi-platform, got %#v", got)
	}
	targets, ok := result.Output["service_targets"].([]string)
	if !ok || len(targets) != 1 || targets[0] != "control-plane" {
		t.Fatalf("expected normalized service_targets [control-plane], got %#v", result.Output["service_targets"])
	}
	deployments, ok := result.Output["deployments"].([]map[string]interface{})
	if !ok || len(deployments) != 1 {
		t.Fatalf("expected one deployment summary, got %#v", result.Output["deployments"])
	}
	if got := deployments[0]["name"]; got != "use1-stage-rsi-agent-platform-control-plane" {
		t.Fatalf("unexpected deployment name %#v", got)
	}
	if got := deployments[0]["available_status"]; got != "True" {
		t.Fatalf("expected available_status=True, got %#v", got)
	}
	images, ok := deployments[0]["images"].([]string)
	if !ok || len(images) != 1 || images[0] != "ghcr.io/piplabs/control-plane:abc123" {
		t.Fatalf("unexpected images %#v", deployments[0]["images"])
	}
}

func TestRSITraceContextReturnsTraceEvidence(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflow := store.ListWorkflows()[0]
	trace, ok := store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflow.ID,
		CommandKind: string(transition.CommandWorkflowFailed),
		CommandID:   "cmd-toolgateway-trace-context-workflow-failed",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"last_error":         "OpenAI rejected tools[0].name",
			"failure_class":      "runner_invalid_request",
			"runner_diagnostics": map[string]any{"provider_error_param": "tools[0].name"},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workflow_failed) error = %v", err)
	}
	service := NewService(config.Config{}, store)

	result := service.Execute("rsi.trace_context", map[string]interface{}{
		"trace_id": trace.Summary.TraceID,
	})

	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	traceSummary, ok := result.Output["trace"].(events.TraceSummary)
	if !ok {
		t.Fatalf("expected trace summary in output, got %#v", result.Output["trace"])
	}
	if traceSummary.TraceID != trace.Summary.TraceID {
		t.Fatalf("unexpected trace id %#v", traceSummary)
	}
	if _, ok := result.Output["workflow_line"]; !ok {
		t.Fatalf("expected workflow_line in output, got %#v", result.Output)
	}
	workflowAttempts, ok := result.Output["workflow_attempts"].([]interface{})
	if !ok || len(workflowAttempts) == 0 {
		t.Fatalf("expected workflow_attempts in output, got %#v", result.Output["workflow_attempts"])
	}
	firstAttempt, ok := workflowAttempts[0].(storepkg.Workflow)
	if !ok {
		t.Fatalf("expected workflow attempt payload, got %#v", workflowAttempts[0])
	}
	if firstAttempt.RunnerDiagnostics["provider_error_param"] != "tools[0].name" {
		t.Fatalf("expected runner diagnostics on workflow attempt, got %#v", firstAttempt.RunnerDiagnostics)
	}
	if _, ok := result.Output["harness_executions"].([]interface{}); !ok {
		t.Fatalf("expected harness_executions in output, got %#v", result.Output["harness_executions"])
	}
}

func TestToolExecutionDoesNotMutateTraceEvidence(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traces := store.ListTraces()
	if len(traces) == 0 {
		t.Fatal("expected seeded traces")
	}
	traceBefore, ok := store.GetTrace(traces[0].TraceID)
	if !ok {
		t.Fatalf("expected trace %s", traces[0].TraceID)
	}
	service := NewService(config.Config{}, store)

	result := service.Execute("rsi.trace_context", map[string]interface{}{
		"trace_id": traces[0].TraceID,
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s %#v", result.Status, result.Output)
	}
	traceAfter, ok := store.GetTrace(traces[0].TraceID)
	if !ok {
		t.Fatalf("expected trace %s", traces[0].TraceID)
	}
	if len(traceAfter.ToolCalls) != len(traceBefore.ToolCalls) {
		t.Fatalf("expected tool execution to leave trace evidence unchanged, before=%d after=%d", len(traceBefore.ToolCalls), len(traceAfter.ToolCalls))
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
