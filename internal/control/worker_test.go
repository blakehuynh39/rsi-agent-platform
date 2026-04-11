package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestProcessWorkflowItemRecordsReasoningAndQueuesEval(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Collected context and prepared a reply.","confidence":0.91,"decision":"reply_in_thread"}],"reply_draft":"Draft reply","final_answer":"Final reply","confidence":0.91,"context_summary":"Repo and KB context collected.","self_critique":"Follow up if channel policy changes."}`,
			"raw":      map[string]any{},
		})
	}))
	defer runner.Close()

	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    "tool-call-1",
			Approved:      true,
			ApprovalState: "not_required",
			Summary:       "Tool call completed.",
			Input:         map[string]any{},
			Output:        map[string]any{"posted": true},
		})
	}))
	defer toolGateway.Close()

	store := storepkg.NewMemoryStore()
	var workflowItem queue.WorkItem
	for _, item := range store.ListWorkItems() {
		if item.Queue == queue.WorkflowQueue {
			workflowItem = item
			break
		}
	}
	if workflowItem.ID == "" {
		t.Fatal("expected queued workflow work item")
	}

	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             runner.URL,
		ToolGatewayBaseURL:        toolGateway.URL,
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}

	if err := processWorkflowItem(cfg, store, clients.NewRunnerClient(cfg.RunnerBaseURL), clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), workflowItem); err != nil {
		t.Fatalf("processWorkflowItem() error = %v", err)
	}

	trace, ok := store.GetTrace(workflowItem.TraceID)
	if !ok {
		t.Fatal("expected updated trace")
	}
	if len(trace.Reasoning) < 3 {
		t.Fatalf("expected visible reasoning to be recorded, got %d steps", len(trace.Reasoning))
	}
	if len(trace.SlackActions) == 0 {
		t.Fatal("expected slack action record to be recorded")
	}

	foundEval := false
	for _, item := range store.ListWorkItems() {
		if item.Queue == queue.EvalQueue && item.TraceID == workflowItem.TraceID {
			foundEval = true
			break
		}
	}
	if !foundEval {
		t.Fatal("expected eval work item to be queued")
	}
}

func TestReplyPolicyAllowsDirectMessagesWithoutChannelPolicy(t *testing.T) {
	store := storepkg.NewMemoryStore()

	allowed, verdict := replyPolicy(store, "architecture", "slack:D123:171000002.000100", "D123")
	if !allowed {
		t.Fatal("expected direct messages to be allowed")
	}
	if verdict != "direct_message" {
		t.Fatalf("expected direct_message verdict, got %s", verdict)
	}
}

func TestReplyPolicyStillBlocksUnknownChannels(t *testing.T) {
	store := storepkg.NewMemoryStore()

	allowed, verdict := replyPolicy(store, "architecture", "slack:C999:171000002.000100", "C999")
	if allowed {
		t.Fatal("expected unknown channels to remain blocked")
	}
	if verdict != "channel_policy_missing" {
		t.Fatalf("expected channel_policy_missing verdict, got %s", verdict)
	}
}
