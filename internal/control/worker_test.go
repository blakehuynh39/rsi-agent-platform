package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
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

func TestControlActionPersistenceFailureFinalizesTraceAndQueuesEval(t *testing.T) {
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
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
	}))
	defer toolGateway.Close()

	baseStore := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		ToolGatewayBaseURL:        toolGateway.URL,
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}

	workflowItem := firstQueuedWorkflowItem(t, baseStore, "slack:")
	if err := processWorkflowItem(cfg, baseStore, nil, workflowItem); err != nil {
		t.Fatalf("processWorkflowItem(run_workflow) error = %v", err)
	}
	_, _ = baseStore.CompleteWorkItem(workflowItem.ID)

	contextActions := queuedWorkItemsForQueue(baseStore, queue.ControlActionQueue)
	if len(contextActions) == 0 {
		t.Fatal("expected queued control actions")
	}
	failingAction := contextActions[0]
	failingIntentID := stringFromMap(failingAction.Payload, "action_intent_id")
	store := &failingRecordActionResultStore{
		Store:        baseStore,
		FailActionID: failingIntentID,
		Err: &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "action_result_pkey",
			TableName:      "action_result",
			Message:        `duplicate key value violates unique constraint "action_result_pkey"`,
		},
	}

	err := processControlActionItem(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), failingAction)
	if err == nil {
		t.Fatal("expected persistence failure to bubble up")
	}
	if _, failErr := store.FailWorkItem(failingAction.ID, err.Error()); failErr != nil {
		t.Fatalf("FailWorkItem() error = %v", failErr)
	}

	intent, ok := baseStore.GetActionIntent(failingIntentID)
	if !ok {
		t.Fatal("expected action intent to exist")
	}
	if intent.Status != action.StatusFailed {
		t.Fatalf("expected failed action intent, got %s", intent.Status)
	}
	if intent.PolicyVerdict != "action_result_primary_key_collision" {
		t.Fatalf("expected runtime failure mode on intent, got %s", intent.PolicyVerdict)
	}

	trace, ok := baseStore.GetTrace(workflowItem.TraceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != "needs-human" {
		t.Fatalf("expected trace to move to needs-human immediately, got %s", trace.Summary.Status)
	}
	foundFailureEvent := false
	for _, event := range trace.Events {
		if event.EventType != "action.persistence_failed" {
			continue
		}
		foundFailureEvent = true
		if !strings.Contains(event.Description, "sqlstate=23505") {
			t.Fatalf("expected SQLSTATE in failure event, got %q", event.Description)
		}
		if !strings.Contains(event.Description, "constraint=action_result_pkey") {
			t.Fatalf("expected constraint in failure event, got %q", event.Description)
		}
	}
	if !foundFailureEvent {
		t.Fatal("expected explicit action.persistence_failed trace event")
	}
	if len(contextActions) > 1 {
		remainingAction := contextActions[1]
		remainingIntentID := stringFromMap(remainingAction.Payload, "action_intent_id")
		if err := processControlActionItem(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), remainingAction); err != nil {
			t.Fatalf("processControlActionItem(remaining) error = %v", err)
		}
		remainingIntent, ok := baseStore.GetActionIntent(remainingIntentID)
		if !ok {
			t.Fatal("expected remaining action intent to exist")
		}
		if remainingIntent.Status != action.StatusCanceled {
			t.Fatalf("expected remaining action to be canceled after terminal failure, got %s", remainingIntent.Status)
		}
		if countQueuedItems(baseStore, queue.WorkflowQueue, controlWorkResumeAfterContext) != 0 {
			t.Fatal("expected no resume_after_context work after terminal persistence failure")
		}
	}

	resumeReply := firstQueuedWorkItem(t, baseStore, queue.WorkflowQueue, controlWorkResumeAfterReply)
	if err := processWorkflowItem(cfg, store, nil, resumeReply); err != nil {
		t.Fatalf("processWorkflowItem(resume_after_reply) error = %v", err)
	}

	trace, _ = baseStore.GetTrace(workflowItem.TraceID)
	if trace.Summary.Status != "needs-human" {
		t.Fatalf("expected terminal needs-human trace, got %s", trace.Summary.Status)
	}
	if trace.Summary.EndedAt.IsZero() {
		t.Fatal("expected terminal trace ended_at to be set")
	}

	foundEval := false
	for _, item := range baseStore.ListWorkItems() {
		if item.Queue == queue.EvalQueue && item.TraceID == workflowItem.TraceID {
			foundEval = true
			break
		}
	}
	if !foundEval {
		t.Fatal("expected eval work item after persistence failure finalization")
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

func TestToolPlanForRepoProgressQuestionUsesGitHubActivity(t *testing.T) {
	plan := toolPlanForIntent("question", "Hello RSI, can you give me a quick rundown of how depin-backend api progressed in the last week", "depin-backend")
	if !containsString(plan, "github.repo_activity") {
		t.Fatalf("expected github.repo_activity in tool plan, got %#v", plan)
	}
}

func TestToolInputForIntentUsesMentionedRepoAndTimeWindow(t *testing.T) {
	cfg := config.Config{
		DefaultRepo:             "rsi-agent-platform",
		AllowedTargetRepos:      []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL: "https://example.test/kb",
		SandboxNamespace:        "rsi-platform",
	}
	input := toolInputForIntent(cfg, storepkg.Workflow{AssignedBot: "arch", Kind: "architecture"}, slackpkg.Ingestion{
		Text:      "Hello RSI, can you give me a quick rundown of how depin-backend api progressed in the last week",
		ChannelID: "D123",
		ThreadTS:  "171000001.000100",
	})

	if got := input["repo"]; got != "depin-backend" {
		t.Fatalf("expected mentioned repo, got %#v", got)
	}
	since, ok := input["since"].(string)
	if !ok || since == "" {
		t.Fatalf("expected non-empty since value, got %#v", input["since"])
	}
	until, ok := input["until"].(string)
	if !ok || until == "" {
		t.Fatalf("expected non-empty until value, got %#v", input["until"])
	}
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		t.Fatalf("parse since: %v", err)
	}
	untilTime, err := time.Parse(time.RFC3339, until)
	if err != nil {
		t.Fatalf("parse until: %v", err)
	}
	if !sinceTime.Before(untilTime) {
		t.Fatalf("expected since before until, got since=%s until=%s", since, until)
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
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

type failingRecordActionResultStore struct {
	storepkg.Store
	FailActionID string
	Err          error
}

func (s *failingRecordActionResultStore) RecordActionResult(result action.Result) (action.Result, error) {
	if result.ActionIntentID == s.FailActionID {
		return action.Result{}, s.Err
	}
	return s.Store.RecordActionResult(result)
}
