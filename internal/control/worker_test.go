package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestWorkflowActionPhasesQueueAndCompleteTrace(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Collected context and prepared a reply.","confidence":0.91,"decision":"reply_in_thread"}],"reply_draft":"Draft reply","final_answer":"Final reply","confidence":0.91,"context_summary":"Repo and KB context collected.","self_critique":"Follow up if channel policy changes.","proposed_actions":[{"kind":"slack_post","target_ref":"CENG","idempotency_key":"reply-action-1","rationale":"Post the answer back into Slack."}]}`,
			"raw":      map[string]any{},
		})
	}))
	defer runner.Close()

	toolCalls := 0
	slackPosts := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		switch name {
		case "repo.context", "knowledge.context", "sentry.lookup", "kubernetes.inspect":
			toolCalls++
			_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
				Name:          name,
				ToolCallID:    name + "-call",
				Approved:      true,
				ApprovalState: "not_required",
				Status:        "completed",
				Available:     true,
				Summary:       "Context gathered.",
				Input:         map[string]any{},
				Output:        map[string]any{"ok": true},
			})
		case "slack.reply":
			slackPosts++
			_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
				Name:          name,
				ToolCallID:    "slack-post-1",
				Approved:      true,
				ApprovalState: "approved",
				Status:        "completed",
				Available:     true,
				Provider:      "slack",
				ProviderRef:   "171000001.000100",
				Summary:       "Slack reply posted.",
				Input:         map[string]any{},
				Output:        map[string]any{"posted": true},
			})
		default:
			t.Fatalf("unexpected tool invocation %s", name)
		}
	}))
	defer toolGateway.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
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

	if err := processWorkflowItem(cfg, store, clients.NewRunnerClient(cfg.RunnerBaseURL), workflowItem); err != nil {
		t.Fatalf("processWorkflowItem(run_workflow) error = %v", err)
	}
	_, _ = store.CompleteWorkItem(workflowItem.ID)

	contextActions := queuedWorkItemsForQueue(store, queue.ControlActionQueue)
	if len(contextActions) != 2 {
		t.Fatalf("expected 2 control actions for context collection, got %d", len(contextActions))
	}
	for _, item := range contextActions {
		if err := processControlActionItem(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), item); err != nil {
			t.Fatalf("processControlActionItem(context) error = %v", err)
		}
		_, _ = store.CompleteWorkItem(item.ID)
	}
	if toolCalls != 2 {
		t.Fatalf("expected 2 context tool calls, got %d", toolCalls)
	}

	resumeContext := firstQueuedWorkItem(t, store, queue.WorkflowQueue, controlWorkResumeAfterContext)
	if err := processWorkflowItem(cfg, store, clients.NewRunnerClient(cfg.RunnerBaseURL), resumeContext); err != nil {
		t.Fatalf("processWorkflowItem(resume_after_context) error = %v", err)
	}
	_, _ = store.CompleteWorkItem(resumeContext.ID)
	if countQueuedItems(store, queue.WorkflowQueue, controlWorkResumeAfterContext) != 0 {
		t.Fatal("expected resume_after_context to be deduped to a single queued item")
	}

	replyAction := firstQueuedWorkItem(t, store, queue.ControlActionQueue, "execute_action")
	if err := processControlActionItem(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), replyAction); err != nil {
		t.Fatalf("processControlActionItem(reply) error = %v", err)
	}
	if err := processControlActionItem(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), replyAction); err != nil {
		t.Fatalf("processControlActionItem(reply duplicate) error = %v", err)
	}
	_, _ = store.CompleteWorkItem(replyAction.ID)
	if slackPosts != 1 {
		t.Fatalf("expected 1 slack post, got %d; actions=%#v work_items=%#v", slackPosts, store.ListActionIntents(), store.ListWorkItems())
	}

	resumeReply := firstQueuedWorkItem(t, store, queue.WorkflowQueue, controlWorkResumeAfterReply)
	if err := processWorkflowItem(cfg, store, clients.NewRunnerClient(cfg.RunnerBaseURL), resumeReply); err != nil {
		t.Fatalf("processWorkflowItem(resume_after_reply) error = %v", err)
	}

	trace, ok := store.GetTrace(workflowItem.TraceID)
	if !ok {
		t.Fatal("expected updated trace")
	}
	if len(trace.Reasoning) < 4 {
		t.Fatalf("expected visible reasoning to be recorded, got %d steps", len(trace.Reasoning))
	}
	if len(trace.ToolCalls) != 2 {
		t.Fatalf("expected tool call records to be persisted, got %d", len(trace.ToolCalls))
	}
	if len(trace.SlackActions) != 1 {
		t.Fatalf("expected one slack action record, got %d", len(trace.SlackActions))
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

func TestSupersededTraceDoesNotPostLateSlackReply(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Collected context and prepared a reply.","confidence":0.91,"decision":"reply_in_thread"}],"reply_draft":"Draft reply","final_answer":"Final reply","confidence":0.91,"proposed_actions":[{"kind":"slack_post","target_ref":"CENG","idempotency_key":"reply-action-2","rationale":"Post the answer back into Slack."}]}`,
			"raw":      map[string]any{},
		})
	}))
	defer runner.Close()

	slackPosts := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		if name == "slack.reply" {
			slackPosts++
		}
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    name + "-call",
			Approved:      true,
			ApprovalState: "approved",
			Status:        "completed",
			Available:     true,
			Summary:       "ok",
			Input:         map[string]any{},
			Output:        map[string]any{"posted": true},
		})
	}))
	defer toolGateway.Close()

	store := storepkg.NewMemoryStore()
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

	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	if err := processWorkflowItem(cfg, store, clients.NewRunnerClient(cfg.RunnerBaseURL), workflowItem); err != nil {
		t.Fatalf("processWorkflowItem(run_workflow) error = %v", err)
	}
	_, _ = store.CompleteWorkItem(workflowItem.ID)
	for _, item := range queuedWorkItemsForQueue(store, queue.ControlActionQueue) {
		if err := processControlActionItem(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), item); err != nil {
			t.Fatalf("processControlActionItem(context) error = %v", err)
		}
		_, _ = store.CompleteWorkItem(item.ID)
	}
	resumeContext := firstQueuedWorkItem(t, store, queue.WorkflowQueue, controlWorkResumeAfterContext)
	if err := processWorkflowItem(cfg, store, clients.NewRunnerClient(cfg.RunnerBaseURL), resumeContext); err != nil {
		t.Fatalf("processWorkflowItem(resume_after_context) error = %v", err)
	}
	_, _ = store.CompleteWorkItem(resumeContext.ID)
	oldReplyAction := firstQueuedWorkItem(t, store, queue.ControlActionQueue, "execute_action")

	_, err := store.CreateEvent(ingestion.EventEnvelope{
		Source:                     ingestion.SourceSlack,
		SourceEventID:              "slack-171000099.000100",
		ThreadKey:                  "slack:CENG:171000001.000100",
		DedupeKey:                  "slack:CENG:171000099.000100",
		Severity:                   ingestion.SeverityWarning,
		NormalizedProblemStatement: "Investigate why staging homepage is failing and propose a fix with newer evidence.",
		OwnershipHint:              "platform",
		RawPayloadRef:              "memory://slack/CENG/171000099-000100.json",
		WorkflowHint:               "incident",
		Metadata: map[string]any{
			"channel_id": "CENG",
			"user_id":    "U123",
			"thread_ts":  "171000001.000100",
		},
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	if err := processControlActionItem(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), oldReplyAction); err != nil {
		t.Fatalf("processControlActionItem(old reply) error = %v", err)
	}
	if slackPosts != 0 {
		t.Fatalf("expected superseded reply to not post to Slack, got %d calls", slackPosts)
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

func firstQueuedWorkItem(t *testing.T, store storepkg.Store, queueName queue.QueueName, kind string) queue.WorkItem {
	t.Helper()
	for _, item := range store.ListWorkItems() {
		if item.Queue == queueName && item.Status == queue.WorkQueued && item.Kind == kind {
			return item
		}
	}
	t.Fatalf("expected queued work item queue=%s kind=%s", queueName, kind)
	return queue.WorkItem{}
}

func firstQueuedWorkflowItem(t *testing.T, store storepkg.Store, threadPrefix string) queue.WorkItem {
	t.Helper()
	for _, item := range store.ListWorkItems() {
		if item.Queue == queue.WorkflowQueue && item.Status == queue.WorkQueued && item.Kind == "run_workflow" && strings.HasPrefix(item.ThreadKey, threadPrefix) {
			return item
		}
	}
	t.Fatalf("expected queued workflow item with thread prefix %s", threadPrefix)
	return queue.WorkItem{}
}

func queuedWorkItemsForQueue(store storepkg.Store, queueName queue.QueueName) []queue.WorkItem {
	out := make([]queue.WorkItem, 0)
	for _, item := range store.ListWorkItems() {
		if item.Queue == queueName && item.Status == queue.WorkQueued {
			out = append(out, item)
		}
	}
	return out
}

func countQueuedItems(store storepkg.Store, queueName queue.QueueName, kind string) int {
	count := 0
	for _, item := range store.ListWorkItems() {
		if item.Queue == queueName && item.Status == queue.WorkQueued && item.Kind == kind {
			count++
		}
	}
	return count
}
