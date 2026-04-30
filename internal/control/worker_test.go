package control

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
	"github.com/piplabs/rsi-agent-platform/internal/workflowplan"
)

func TestWorkflowActionPhasesQueueAndCompleteTrace(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Collected context and prepared a reply.","confidence":0.91,"decision":"reply_in_thread"}],"reply_draft":"Draft reply","final_answer":"Final reply","confidence":0.91,"context_summary":"Repo and KB context collected.","self_critique":"Follow up if channel policy changes.","proposed_actions":[{"kind":"slack_post","target_ref":"CENG","idempotency_key":"reply-action-1","rationale":"Post the answer back into Slack."}]}`,
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":  "analysis",
							"summary":    "Collected context and prepared a reply.",
							"confidence": 0.91,
							"decision":   "reply_in_thread",
						},
					},
					"reply_draft":     "Draft reply",
					"final_answer":    "Final reply",
					"confidence":      0.91,
					"context_summary": "Repo and KB context collected.",
					"self_critique":   "Follow up if channel policy changes.",
					"proposed_actions": []any{
						map[string]any{
							"kind":            "slack_post",
							"target_ref":      "CENG",
							"idempotency_key": "reply-action-1",
							"rationale":       "Post the answer back into Slack.",
						},
					},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	toolCalls := []string{}
	slackPosts := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
		switch name {
		case "repo.context", "knowledge.context", "sentry.lookup", "kubernetes.inspect", "github.repo_activity", "slack.history", "rsi.workflow_context", "rsi.action_chain", "rsi.runtime_health", "rsi.runtime_deployment_facts":
			toolCalls = append(toolCalls, name)
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

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	contextActions := queuedActionEffectsForPlane(store, "control")
	if len(contextActions) != 0 {
		t.Fatalf("expected no deterministic context actions once live hints are seeded directly into Hermes, got %d", len(contextActions))
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	replyAction := firstQueuedActionEffectByKind(t, store, "control", action.KindSlackPost)
	replyActionID := replyAction.AggregateID
	if receipt, ok := store.GetCommandReceipt(actionCommandID(replyActionID, transition.CommandActionQueue, "")); !ok || receipt.MachineKind != transition.MachineAction {
		t.Fatalf("expected action_queued receipt for reply action %s, got ok=%t receipt=%+v", replyActionID, ok, receipt)
	}
	if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), replyAction); err != nil {
		t.Fatalf("processControlActionEffect(reply) error = %v", err)
	}
	if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), replyAction); err != nil {
		t.Fatalf("processControlActionEffect(reply duplicate) error = %v", err)
	}
	if slackPosts != 1 {
		t.Fatalf("expected 1 slack post, got %d; actions=%#v", slackPosts, store.ListActionIntents())
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected updated trace")
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "completed" {
		t.Fatalf("expected workflow to complete through reducer path, got %s", workflow.Status)
	}
	if len(trace.Reasoning) < 4 {
		t.Fatalf("expected visible reasoning to be recorded, got %d steps", len(trace.Reasoning))
	}
	if len(trace.ToolCalls) != 0 {
		t.Fatalf("expected no deterministic prefetch tool records, got %d", len(trace.ToolCalls))
	}
	if len(trace.SlackActions) != 1 {
		t.Fatalf("expected one slack action record, got %d", len(trace.SlackActions))
	}
	foundSeededEvent := false
	for _, event := range trace.Events {
		if event.EventType == "context.seeded" {
			foundSeededEvent = true
			break
		}
	}
	if !foundSeededEvent {
		t.Fatal("expected trace to record context.seeded for seeded-open runner hints")
	}
	if !reflect.DeepEqual(toolCalls, []string{"slack.history"}) {
		t.Fatalf("expected one bound-thread slack.history prefetch call, got %#v", toolCalls)
	}

	if len(store.ListEvalRuns()) == 0 {
		t.Fatal("expected workflow completion to trigger immediate problem-line evaluation")
	}
	assertWorkflowEffectStatus(t, store, workflowItem.workflowID, transition.EffectInvokeRunner, transition.EffectCompleted)
	assertWorkflowEffectStatus(t, store, workflowItem.workflowID, transition.EffectPostSlackReply, transition.EffectCompleted)
}

func TestWorkflowRunnerUsesHermesExecutorWhenConfigured(t *testing.T) {
	var executorPath string
	var executorTask clients.RunnerTask
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executorPath = r.URL.Path
		var payload struct {
			Task clients.RunnerTask `json:"task"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		executorTask = payload.Task
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake-executor",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Collected context and prepared a reply.","confidence":0.91,"decision":"reply_in_thread"}],"reply_draft":"Draft reply","final_answer":"Final reply","confidence":0.91,"context_summary":"Repo and KB context collected.","self_critique":"Follow up if channel policy changes.","proposed_actions":[{"kind":"slack_post","target_ref":"CENG","idempotency_key":"reply-action-executor-1","rationale":"Post the answer back into Slack."}]}`,
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":  "analysis",
							"summary":    "Collected context and prepared a reply.",
							"confidence": 0.91,
							"decision":   "reply_in_thread",
						},
					},
					"reply_draft":     "Draft reply",
					"final_answer":    "Final reply",
					"confidence":      0.91,
					"context_summary": "Repo and KB context collected.",
					"self_critique":   "Follow up if channel policy changes.",
					"proposed_actions": []any{
						map[string]any{
							"kind":            "slack_post",
							"target_ref":      "CENG",
							"idempotency_key": "reply-action-executor-1",
							"rationale":       "Post the answer back into Slack.",
						},
					},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		HermesExecutorBaseURL:     executor.URL,
		ToolGatewayBaseURL:        "http://tool-gateway.invalid",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
		ProdRunnerTimeout:         930 * time.Second,
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	if executorPath != "/internal/hermes-executions" {
		t.Fatalf("executor path = %q, want /internal/hermes-executions", executorPath)
	}
	if executorTask.OperationID != runnerEffect.ID {
		t.Fatalf("executor operation_id = %q, want %q", executorTask.OperationID, runnerEffect.ID)
	}
	if !strings.HasPrefix(executorTask.ExecutionID, "hexec-") {
		t.Fatalf("executor execution_id = %q, want hexec-*", executorTask.ExecutionID)
	}
	if executorTask.WorkflowID != workflowItem.workflowID {
		t.Fatalf("executor workflow_id = %q, want %q", executorTask.WorkflowID, workflowItem.workflowID)
	}
}

func TestWorkflowRunnerStartsAsyncHermesExecutionAndDefersEffect(t *testing.T) {
	fallbackRunner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected fallback runner call to %s", r.URL.Path)
	}))
	defer fallbackRunner.Close()

	startCalls := 0
	var started clients.HermesExecutionRequest
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		startCalls++
		if err := json.NewDecoder(r.Body).Decode(&started); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: started.Task.ExecutionID,
			OperationID: started.Task.OperationID,
			WorkflowID:  started.Task.WorkflowID,
			TraceID:     started.Task.TraceID,
			Status:      "accepted",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:                     "control-plane",
		DefaultRepo:                     "rsi-agent-platform",
		DefaultKnowledgeBaseURL:         "https://example.test/kb",
		AllowedTargetRepos:              []string{"rsi-agent-platform"},
		RunnerBaseURL:                   fallbackRunner.URL,
		HermesExecutorBaseURL:           executor.URL,
		ToolGatewayBaseURL:              "http://tool-gateway.invalid",
		SandboxNamespace:                "rsi-platform",
		DefaultReasoningVerbosity:       "verbose",
		ProdRunnerTimeout:               930 * time.Second,
		AsyncHermesExecutionEnabled:     true,
		HermesExecutionHeartbeatTimeout: 120 * time.Second,
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	claimed := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)

	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, claimed)

	if startCalls != 1 {
		t.Fatalf("expected one async executor start call, got %d", startCalls)
	}
	if !started.Async {
		t.Fatalf("expected async executor request, got %#v", started)
	}
	if started.Task.OperationID != claimed.ID {
		t.Fatalf("operation_id = %q, want %q", started.Task.OperationID, claimed.ID)
	}
	record, ok := store.GetRunnerExecution(started.Task.ExecutionID)
	if !ok {
		t.Fatalf("expected runner execution %s to be durable", started.Task.ExecutionID)
	}
	if record.Status != "accepted" {
		t.Fatalf("runner execution status = %q, want accepted", record.Status)
	}
	expectedHolder := runnerExecutionHolder(started.Task.ExecutionID)
	if record.Holder != expectedHolder {
		t.Fatalf("runner execution holder = %q, want %q", record.Holder, expectedHolder)
	}
	if got := stringValue(started.Task.ExecutionIntent["runner_execution_holder"]); got != expectedHolder {
		t.Fatalf("runner task execution holder = %q, want %q", got, expectedHolder)
	}
	effect, ok := workflowEffectByPayload(store, workflowItem.workflowID, transition.EffectInvokeRunner, "", "")
	if !ok {
		t.Fatal("expected workflow runner effect")
	}
	if effect.Status != transition.EffectRunning || effect.Holder != "" || effect.NotBefore == nil {
		t.Fatalf("expected async start to defer running effect for polling, got %+v", effect)
	}
}

func TestWorkflowRunnerAsyncImmediateResultTerminalizesRunnerExecution(t *testing.T) {
	startCalls := 0
	var started clients.HermesExecutionRequest
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		startCalls++
		if err := json.NewDecoder(r.Body).Decode(&started); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: started.Task.ExecutionID,
			OperationID: started.Task.OperationID,
			WorkflowID:  started.Task.WorkflowID,
			TraceID:     started.Task.TraceID,
			Status:      "completed",
			Result: &clients.RunnerResponse{
				OK:       true,
				Provider: "fake-executor",
				Message:  "Immediate async result",
				Raw: map[string]any{
					"structured_output": map[string]any{
						"visible_reasoning": []any{
							map[string]any{
								"step_type":  "analysis",
								"summary":    "Executor returned immediately.",
								"confidence": 1.0,
								"decision":   "complete",
							},
						},
						"reply_draft":            "Immediate async result",
						"final_answer":           "Immediate async result",
						"confidence":             1.0,
						"context_summary":        "Executor returned a completed async result.",
						"self_critique":          "",
						"proposed_actions":       []any{},
						"knowledge_drafts":       []any{},
						"outcome_hypotheses":     []any{},
						"produced_artifacts":     []any{},
						"completion_verdict":     "complete",
						"termination_reason":     "normal_completion",
						"reply_delivery":         map[string]any{},
						"artifact_render_briefs": []any{},
					},
				},
			},
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:                     "control-plane",
		DefaultRepo:                     "rsi-agent-platform",
		DefaultKnowledgeBaseURL:         "https://example.test/kb",
		AllowedTargetRepos:              []string{"rsi-agent-platform"},
		RunnerBaseURL:                   executor.URL,
		HermesExecutorBaseURL:           executor.URL,
		ToolGatewayBaseURL:              "http://tool-gateway.invalid",
		SandboxNamespace:                "rsi-platform",
		DefaultReasoningVerbosity:       "verbose",
		ProdRunnerTimeout:               930 * time.Second,
		AsyncHermesExecutionEnabled:     true,
		HermesExecutionHeartbeatTimeout: 120 * time.Second,
	}
	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	claimed := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)

	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, claimed)

	if startCalls != 1 {
		t.Fatalf("expected one async executor start call, got %d", startCalls)
	}
	record, ok := store.GetRunnerExecution(started.Task.ExecutionID)
	if !ok {
		t.Fatalf("expected runner execution %s", started.Task.ExecutionID)
	}
	if record.Status != "completed" || record.CompletedAt == nil {
		t.Fatalf("expected completed runner execution with completed_at, got %+v", record)
	}
	expectedHolder := runnerExecutionHolder(started.Task.ExecutionID)
	if record.Holder != expectedHolder {
		t.Fatalf("runner execution holder = %q, want %q", record.Holder, expectedHolder)
	}
	if len(record.Result) == 0 {
		t.Fatalf("expected completed runner execution to persist result")
	}
	for _, active := range store.ListActiveRunnerExecutions() {
		if active.ExecutionID == started.Task.ExecutionID {
			t.Fatalf("immediate completed execution should not remain active: %+v", active)
		}
	}
	assertWorkflowEffectStatus(t, store, workflowItem.workflowID, transition.EffectInvokeRunner, transition.EffectCompleted)
}

func TestWorkflowRunnerAsyncHeartbeatExpiryFailsClosed(t *testing.T) {
	var expectedExecutionID string
	statusCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/hermes-executions/"+expectedExecutionID:
			statusCalls++
			http.Error(w, "executor unavailable", http.StatusServiceUnavailable)
		case r.Method == http.MethodPost && r.URL.Path == "/internal/hermes-executions":
			t.Fatalf("stale heartbeat with unreachable executor must not launch duplicate execution")
		default:
			http.NotFound(w, r)
		}
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:                     "control-plane",
		DefaultRepo:                     "rsi-agent-platform",
		DefaultKnowledgeBaseURL:         "https://example.test/kb",
		AllowedTargetRepos:              []string{"rsi-agent-platform"},
		RunnerBaseURL:                   executor.URL,
		HermesExecutorBaseURL:           executor.URL,
		ToolGatewayBaseURL:              "http://tool-gateway.invalid",
		SandboxNamespace:                "rsi-platform",
		DefaultReasoningVerbosity:       "verbose",
		ProdRunnerTimeout:               930 * time.Second,
		AsyncHermesExecutionEnabled:     true,
		HermesExecutionHeartbeatTimeout: 120 * time.Second,
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	claimed := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatalf("expected workflow %s", workflowItem.workflowID)
	}
	expectedExecutionID = workflowExecutionID(claimed.ID, time.Now().UTC())
	staleHeartbeat := time.Now().Add(-5 * time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:    expectedExecutionID,
		OperationID:    claimed.ID,
		WorkflowID:     workflowItem.workflowID,
		TraceID:        workflowItem.traceID,
		ConversationID: workflow.ConversationID,
		CaseID:         workflow.CaseID,
		Status:         "running",
		HeartbeatAt:    &staleHeartbeat,
		CreatedAt:      staleHeartbeat,
		UpdatedAt:      staleHeartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, claimed)

	if statusCalls != 1 {
		t.Fatalf("expected one status poll, got %d", statusCalls)
	}
	record, ok := store.GetRunnerExecution(expectedExecutionID)
	if !ok {
		t.Fatalf("expected runner execution %s", expectedExecutionID)
	}
	if record.Status != "failed" || record.CompletedAt == nil {
		t.Fatalf("expected stale execution to fail closed, got %+v", record)
	}
	if record.FailureClass != workflowFailureRunnerExecutorStatusUnavailable {
		t.Fatalf("failure_class = %q, want %q", record.FailureClass, workflowFailureRunnerExecutorStatusUnavailable)
	}
	effect, ok := workflowEffectByPayload(store, workflowItem.workflowID, transition.EffectInvokeRunner, "", "")
	if !ok {
		t.Fatal("expected workflow runner effect")
	}
	if effect.Status != transition.EffectCompleted {
		t.Fatalf("expected heartbeat failure to finalize the claimed effect, got %+v", effect)
	}
}

func TestAsyncHermesExecutionStartTimeoutFailsClosedOnFirstAttempt(t *testing.T) {
	startCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		startCalls++
		http.Error(w, "executor unavailable", http.StatusServiceUnavailable)
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Nanosecond},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-start-timeout"},
		transition.EffectExecution{ID: "eff-start-timeout"},
		"prod",
		workflowContext{
			trace: events.Trace{Summary: events.TraceSummary{TraceID: "trace-start-timeout"}},
			workflow: storepkg.Workflow{
				ID:             "wf-start-timeout",
				ConversationID: "conv-start-timeout",
				CaseID:         "case-start-timeout",
			},
		},
		now,
	)
	if startCalls != 1 {
		t.Fatalf("start calls = %d, want 1", startCalls)
	}
	if wait {
		t.Fatal("expired start attempt should fail closed, not defer")
	}
	if resp.OK {
		t.Fatalf("expected no successful runner response, got %+v", resp)
	}
	var workflowErr *workflowFailureError
	if !errors.As(err, &workflowErr) {
		t.Fatalf("error = %v, want workflowFailureError", err)
	}
	if workflowErr.failure.Class != workflowFailureRunnerTransportTimeout {
		t.Fatalf("failure class = %q, want %q", workflowErr.failure.Class, workflowFailureRunnerTransportTimeout)
	}
	record, ok := store.GetRunnerExecution("hexec-start-timeout")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "failed" || record.CompletedAt == nil {
		t.Fatalf("expected failed runner execution, got %+v", record)
	}
}

func TestAsyncHermesExecutionStartFailureUsesExistingQueuedFreshness(t *testing.T) {
	statusCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/hermes-executions/hexec-existing-queued-stale" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		statusCalls++
		http.Error(w, "executor unavailable", http.StatusServiceUnavailable)
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	startedAt := time.Now().Add(-5 * time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-existing-queued-stale",
		Status:      "queued",
		CreatedAt:   startedAt,
		UpdatedAt:   startedAt,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-existing-queued-stale"},
		transition.EffectExecution{ID: "eff-existing-queued-stale"},
		"prod",
		workflowContext{},
		time.Now().UTC(),
	)
	if statusCalls != 1 {
		t.Fatalf("status calls = %d, want 1", statusCalls)
	}
	if wait || resp.OK {
		t.Fatalf("stale queued start failure should fail closed, wait=%t resp=%+v", wait, resp)
	}
	var workflowErr *workflowFailureError
	if !errors.As(err, &workflowErr) {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v, want workflowFailureError", err)
	}
	if workflowErr.failure.Class != workflowFailureRunnerTransportTimeout {
		t.Fatalf("failure class = %q, want %q", workflowErr.failure.Class, workflowFailureRunnerTransportTimeout)
	}
	record, ok := store.GetRunnerExecution("hexec-existing-queued-stale")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "failed" || record.CompletedAt == nil || record.HeartbeatAt == nil {
		t.Fatalf("expected failed runner execution with terminal heartbeat, got %+v", record)
	}
}

func TestAsyncHermesExecutionFreshQueuedRecoveryErrorDoesNotStartDuplicate(t *testing.T) {
	statusCalls := 0
	startCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/hermes-executions/hexec-existing-queued-fresh":
			statusCalls++
			http.Error(w, "executor unavailable", http.StatusServiceUnavailable)
		case r.Method == http.MethodPost && r.URL.Path == "/internal/hermes-executions":
			startCalls++
			t.Fatalf("fresh queued recovery error must not start duplicate Hermes execution")
		default:
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	startedAt := time.Now().Add(-10 * time.Second).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-existing-queued-fresh",
		Status:      "queued",
		CreatedAt:   startedAt,
		UpdatedAt:   startedAt,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-existing-queued-fresh"},
		transition.EffectExecution{ID: "eff-existing-queued-fresh"},
		"prod",
		workflowContext{},
		time.Now().UTC(),
	)
	if !errors.Is(err, errHermesExecutionStillRunning) || !wait || resp.OK {
		t.Fatalf("fresh queued recovery error result = resp=%+v wait=%t err=%v", resp, wait, err)
	}
	if statusCalls != 1 || startCalls != 0 {
		t.Fatalf("status calls=%d start calls=%d, want 1/0", statusCalls, startCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-existing-queued-fresh")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "queued" || record.CompletedAt != nil {
		t.Fatalf("fresh queued recovery error should leave queued record deferred, got %+v", record)
	}
}

func TestAsyncHermesExecutionQueuedRecoveryPersistsCompletedResult(t *testing.T) {
	statusCalls := 0
	startCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/hermes-executions/hexec-queued-recovered":
			statusCalls++
			_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
				ExecutionID: "hexec-queued-recovered",
				Status:      "completed",
				Result: &clients.RunnerResponse{
					OK:       true,
					Provider: "hermes-executor",
					Message:  "recovered",
					Raw:      map[string]any{"structured_output": map[string]any{"final_answer": "recovered"}},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/internal/hermes-executions":
			startCalls++
			t.Fatalf("queued recovery must not start duplicate Hermes execution")
		default:
			http.NotFound(w, r)
		}
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().Add(-time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-queued-recovered",
		Status:      "queued",
		Holder:      "hermes-executor:hexec-queued-recovered",
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-queued-recovered"},
		transition.EffectExecution{ID: "eff-queued-recovered"},
		"prod",
		workflowContext{},
		time.Now().UTC(),
	)
	if err != nil || wait || !resp.OK || resp.Message != "recovered" {
		t.Fatalf("queued recovery result = resp=%+v wait=%t err=%v", resp, wait, err)
	}
	if statusCalls != 1 || startCalls != 0 {
		t.Fatalf("status calls=%d start calls=%d, want 1/0", statusCalls, startCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-queued-recovered")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "completed" || record.CompletedAt == nil || record.HeartbeatAt == nil {
		t.Fatalf("queued recovery should persist completed record, got %+v", record)
	}
	if stored, ok := runnerResponseFromMap(record.Result); !ok || !stored.OK || stored.Message != "recovered" {
		t.Fatalf("expected persisted recovered response, ok=%t stored=%+v", ok, stored)
	}
}

func TestAsyncHermesExecutionQueuedRecoveryPersistsStillRunningStatus(t *testing.T) {
	statusCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/hermes-executions/hexec-queued-running" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		statusCalls++
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-queued-running",
			Status:      "running",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().Add(-time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-queued-running",
		Status:      "queued",
		Holder:      "hermes-executor:hexec-queued-running",
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-queued-running"},
		transition.EffectExecution{ID: "eff-queued-running"},
		"prod",
		workflowContext{},
		time.Now().UTC(),
	)
	if !errors.Is(err, errHermesExecutionStillRunning) || !wait || resp.OK {
		t.Fatalf("queued running recovery result = resp=%+v wait=%t err=%v", resp, wait, err)
	}
	if statusCalls != 1 {
		t.Fatalf("status calls=%d, want 1", statusCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-queued-running")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "running" || record.HeartbeatAt == nil {
		t.Fatalf("queued recovery should persist running heartbeat, got %+v", record)
	}
}

func TestAsyncHermesExecutionQueuedRecoveryTimesOutInsteadOfDeferringForever(t *testing.T) {
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/hermes-executions/hexec-queued-wedged" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-queued-wedged",
			Status:      "queued",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	startedAt := time.Now().Add(-5 * time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-queued-wedged",
		Status:      "queued",
		Holder:      "hermes-executor:hexec-queued-wedged",
		CreatedAt:   startedAt,
		UpdatedAt:   startedAt,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-queued-wedged"},
		transition.EffectExecution{ID: "eff-queued-wedged"},
		"prod",
		workflowContext{},
		time.Now().UTC(),
	)
	if err != nil || wait || resp.OK {
		t.Fatalf("queued wedged recovery result = resp=%+v wait=%t err=%v", resp, wait, err)
	}
	if got := stringValue(resp.Raw["failure_class"]); got != workflowFailureRunnerExecutorStatusUnavailable {
		t.Fatalf("failure_class = %q, want %q", got, workflowFailureRunnerExecutorStatusUnavailable)
	}
	record, ok := store.GetRunnerExecution("hexec-queued-wedged")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "failed" || record.CompletedAt == nil {
		t.Fatalf("queued wedged recovery should fail closed, got %+v", record)
	}
}

func TestAsyncHermesExecutionQueuedPollTimesOutInsteadOfDeferringForever(t *testing.T) {
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/hermes-executions/hexec-running-queued-wedged" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-running-queued-wedged",
			Status:      "queued",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	heartbeat := time.Now().Add(-5 * time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-running-queued-wedged",
		Status:      "running",
		Holder:      "hermes-executor:hexec-running-queued-wedged",
		HeartbeatAt: &heartbeat,
		CreatedAt:   heartbeat,
		UpdatedAt:   heartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-running-queued-wedged"},
		transition.EffectExecution{ID: "eff-running-queued-wedged"},
		"prod",
		workflowContext{},
		time.Now().UTC(),
	)
	if err != nil || wait || resp.OK {
		t.Fatalf("queued polling timeout result = resp=%+v wait=%t err=%v", resp, wait, err)
	}
	if got := stringValue(resp.Raw["failure_class"]); got != workflowFailureRunnerExecutorStatusUnavailable {
		t.Fatalf("failure_class = %q, want %q", got, workflowFailureRunnerExecutorStatusUnavailable)
	}
	record, ok := store.GetRunnerExecution("hexec-running-queued-wedged")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "failed" || record.CompletedAt == nil {
		t.Fatalf("queued polling timeout should fail closed, got %+v", record)
	}
}

func TestAsyncHermesExecutionStartCASDoesNotOverwriteConcurrentHeartbeat(t *testing.T) {
	store := storepkg.NewMemoryStore()
	startCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		startCalls++
		heartbeat := time.Now().UTC()
		if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
			ExecutionID: "hexec-start-race",
			Status:      "running",
			Holder:      "hermes-executor:hexec-start-race",
			HeartbeatAt: &heartbeat,
			UpdatedAt:   heartbeat,
		}); err != nil {
			t.Fatalf("RecordRunnerExecution(concurrent heartbeat) error = %v", err)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-start-race",
			Status:      "accepted",
		})
	}))
	defer executor.Close()

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-start-race"},
		transition.EffectExecution{ID: "eff-start-race"},
		"prod",
		workflowContext{},
		time.Now().UTC(),
	)
	if !errors.Is(err, errHermesExecutionStillRunning) || !wait || resp.OK {
		t.Fatalf("start race result = resp=%+v wait=%t err=%v", resp, wait, err)
	}
	if startCalls != 1 {
		t.Fatalf("start calls=%d, want 1", startCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-start-race")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "running" {
		t.Fatalf("concurrent heartbeat should not be overwritten by accepted start status: %+v", record)
	}
}

func TestAsyncHermesExecutionStatusTimeoutUsesCreatedAtWithoutHeartbeat(t *testing.T) {
	statusCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/hermes-executions/hexec-no-heartbeat-stale" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		statusCalls++
		http.Error(w, "executor unavailable", http.StatusServiceUnavailable)
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	createdAt := now.Add(-5 * time.Minute)
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-no-heartbeat-stale",
		Status:      "running",
		Holder:      "hermes-executor:old",
		CreatedAt:   createdAt,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	_, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-no-heartbeat-stale"},
		transition.EffectExecution{ID: "eff-no-heartbeat-stale"},
		"prod",
		workflowContext{},
		now,
	)
	if statusCalls != 1 {
		t.Fatalf("status calls = %d, want 1", statusCalls)
	}
	if err != errHermesExecutionStillRunning {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v, want errHermesExecutionStillRunning", err)
	}
	if !wait {
		t.Fatalf("running execution without heartbeat should defer (wait=true), got wait=%t", wait)
	}
	record, ok := store.GetRunnerExecution("hexec-no-heartbeat-stale")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status == "failed" {
		t.Fatalf("execution should not be failed when no heartbeat reference available, got %+v", record)
	}
}

func TestAsyncHermesExecutionPollResultAfterConcurrentCancelIsNonDeliverable(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	heartbeat := now.Add(-10 * time.Second)
	executionID := "hexec-poll-cancel-race"
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: executionID,
		CaseID:      "case-1",
		TraceID:     "trace-old",
		Status:      "running",
		Holder:      runnerExecutionHolder(executionID),
		HeartbeatAt: &heartbeat,
		CreatedAt:   heartbeat,
		UpdatedAt:   heartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	statusCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/hermes-executions/"+executionID {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		statusCalls++
		if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
			ExecutionID:     executionID,
			Status:          "cancel_requested",
			CancelRequested: true,
			FailureClass:    workflowFailureRunnerExecutionCancelled,
			UpdatedAt:       time.Now().UTC(),
		}); err != nil {
			t.Fatalf("RecordRunnerExecution(concurrent cancel) error = %v", err)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: executionID,
			Status:      "completed",
			Result: &clients.RunnerResponse{
				OK:       true,
				Provider: "hermes-executor",
				Message:  "late success",
				Raw:      map[string]any{"structured_output": map[string]any{"final_answer": "late success"}},
			},
		})
	}))
	defer executor.Close()

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: executionID},
		transition.EffectExecution{ID: "eff-poll-cancel-race"},
		"prod",
		workflowContext{},
		now,
	)
	if statusCalls != 1 {
		t.Fatalf("status calls = %d, want 1", statusCalls)
	}
	var failure *workflowFailureError
	if !errors.As(err, &failure) || failure.failure.Class != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v, want cancellation failure", err)
	}
	if wait || resp.OK {
		t.Fatalf("late cancelled result must not be deliverable, resp=%+v wait=%t", resp, wait)
	}
	record, ok := store.GetRunnerExecution(executionID)
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelled" || !record.CancelRequested || record.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("concurrent cancellation should dominate late result, got %+v", record)
	}
}

func TestAsyncHermesExecutionCancelRequestedRunningDispatchesCancel(t *testing.T) {
	cancelCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-cancel-running/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		cancelCalls++
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-cancel-running",
			Status:      "cancelling",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-running",
		Status:          "running",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-cancel-running"},
		transition.EffectExecution{ID: "eff-cancel-running"},
		"prod",
		workflowContext{},
		now,
	)
	if err != nil {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v", err)
	}
	if !wait || resp.OK {
		t.Fatalf("cancel-requested running execution should defer without deliverable response, wait=%t resp=%+v", wait, resp)
	}
	if cancelCalls != 1 {
		t.Fatalf("cancel calls=%d, want 1", cancelCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-cancel-running")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelling" || !record.CancelRequested {
		t.Fatalf("expected cancel dispatch to move running record to cancelling, got %+v", record)
	}
}

func TestAsyncHermesExecutionCancelRequestedDispatchesCancelAndDefers(t *testing.T) {
	cancelCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-cancelled/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		cancelCalls++
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-cancelled",
			Status:      "cancelling",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancelled",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-cancelled"},
		transition.EffectExecution{ID: "eff-cancelled"},
		"prod",
		workflowContext{},
		now,
	)
	if err != nil {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v", err)
	}
	if !wait {
		t.Fatal("cancel-requested execution should defer after dispatching cancellation")
	}
	if resp.OK {
		t.Fatalf("expected no successful runner response while cancellation is in progress, got %+v", resp)
	}
	if cancelCalls != 1 {
		t.Fatalf("cancel calls = %d, want 1", cancelCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-cancelled")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelling" || !record.CancelRequested {
		t.Fatalf("runner execution should be marked cancelling after cancel dispatch, got %+v", record)
	}
	if len(store.ListActiveRunnerExecutions()) != 1 {
		t.Fatalf("cancelling runner execution should remain active until terminal status, got %+v", store.ListActiveRunnerExecutions())
	}
}

func TestAsyncHermesExecutionCancelRetryDoesNotRefreshRunnerHeartbeat(t *testing.T) {
	cancelCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-cancel-error/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		cancelCalls++
		http.Error(w, "temporarily unavailable", http.StatusBadGateway)
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	heartbeat := now.Add(-30 * time.Second)
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-error",
		Status:          "cancel_requested",
		Holder:          "hermes-executor:hexec-cancel-error",
		CancelRequested: true,
		HeartbeatAt:     &heartbeat,
		CreatedAt:       now.Add(-time.Minute),
		UpdatedAt:       heartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-cancel-error"},
		transition.EffectExecution{ID: "eff-cancel-error"},
		"prod",
		workflowContext{},
		now,
	)
	if !errors.Is(err, errHermesExecutionStillRunning) || !wait || resp.OK {
		t.Fatalf("cancel retry result = resp=%+v wait=%t err=%v", resp, wait, err)
	}
	if cancelCalls != 1 {
		t.Fatalf("cancel calls = %d, want 1", cancelCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-cancel-error")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelling" || !record.CancelRequested {
		t.Fatalf("cancel retry should preserve cancellation state, got %+v", record)
	}
	if record.HeartbeatAt == nil || !record.HeartbeatAt.Equal(heartbeat) {
		t.Fatalf("cancel retry must not refresh runner heartbeat, got %+v want %v", record.HeartbeatAt, heartbeat)
	}
	if !record.UpdatedAt.After(heartbeat) {
		t.Fatalf("cancel retry should update audit timestamp without extending heartbeat, got %+v", record)
	}
}

func TestAsyncHermesExecutionCancelRequestedSuccessfulResultIsPersistedButNotDeliverable(t *testing.T) {
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-cancel-success/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-cancel-success",
			Status:      "cancelling",
			Result: &clients.RunnerResponse{
				OK:       true,
				Provider: "hermes-executor",
				Message:  "completed before cancellation landed",
				Raw:      map[string]any{},
			},
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-success",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-cancel-success"},
		transition.EffectExecution{ID: "eff-cancel-success"},
		"prod",
		workflowContext{},
		now,
	)
	if wait {
		t.Fatal("successful cancel race should not defer")
	}
	if resp.OK {
		t.Fatalf("successful cancel race must not return a deliverable response, got %+v", resp)
	}
	var workflowErr *workflowFailureError
	if !errors.As(err, &workflowErr) {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v, want workflowFailureError", err)
	}
	if workflowErr.failure.Class != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("failure class = %q, want %q", workflowErr.failure.Class, workflowFailureRunnerExecutionCancelled)
	}
	record, ok := store.GetRunnerExecution("hexec-cancel-success")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelled" || record.CompletedAt == nil || record.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("successful cancel race should persist non-deliverable cancelled status, got %+v", record)
	}
	if stored, ok := runnerResponseFromMap(record.Result); !ok || !stored.OK || stored.Message != "completed before cancellation landed" {
		t.Fatalf("expected audit result to be preserved, ok=%t stored=%+v", ok, stored)
	}
}

func TestAsyncHermesExecutionCancelRequestedFailedResultUsesCancellationFailureClass(t *testing.T) {
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-cancel-failed-result/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-cancel-failed-result",
			Status:      "failed",
			Result: &clients.RunnerResponse{
				OK:       false,
				Provider: "hermes-executor",
				Message:  "failed before cancellation landed",
				Raw:      map[string]any{"failure_class": "worker_failed"},
			},
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-failed-result",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-cancel-failed-result"},
		transition.EffectExecution{ID: "eff-cancel-failed-result"},
		"prod",
		workflowContext{},
		now,
	)
	if wait || resp.OK {
		t.Fatalf("failed cancel race should not return a deliverable response, wait=%t resp=%+v", wait, resp)
	}
	var workflowErr *workflowFailureError
	if !errors.As(err, &workflowErr) || workflowErr.failure.Class != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v, want cancellation workflow failure", err)
	}
	record, ok := store.GetRunnerExecution("hexec-cancel-failed-result")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelled" || record.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("failed cancel race should persist cancellation failure class, got %+v", record)
	}
	if stored, ok := runnerResponseFromMap(record.Result); !ok || stored.OK || stringValue(stored.Raw["failure_class"]) != "worker_failed" {
		t.Fatalf("expected original audit result to be preserved, ok=%t stored=%+v", ok, stored)
	}
}

func TestAsyncHermesExecutionCancelRequestedCompletedWithoutResultFailsResultUnavailable(t *testing.T) {
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-cancel-completed/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-cancel-completed",
			Status:      "completed",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-completed",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-cancel-completed"},
		transition.EffectExecution{ID: "eff-cancel-completed"},
		"prod",
		workflowContext{},
		now,
	)
	var failure *workflowFailureError
	if !errors.As(err, &failure) || failure.failure.Class != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v, want cancellation failure", err)
	}
	if wait || resp.OK {
		t.Fatalf("completed-without-result cancel race should not be deliverable, wait=%t resp=%+v", wait, resp)
	}
	record, ok := store.GetRunnerExecution("hexec-cancel-completed")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelled" || record.CompletedAt == nil || record.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("completed-without-result cancel race should persist cancelled audit record, got %+v", record)
	}
}

func TestRecoverHermesExecutionResultTreatsActiveStatusesAsStillRunning(t *testing.T) {
	for _, status := range []string{"queued", "starting", "cancel_requested"} {
		t.Run(status, func(t *testing.T) {
			executionID := "hexec-" + status
			executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/internal/hermes-executions/"+executionID {
					t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
				}
				_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
					ExecutionID: executionID,
					Status:      status,
				})
			}))
			defer executor.Close()

			resp, wait, err := recoverHermesExecutionResult(clients.NewRunnerClient(executor.URL), executionID)
			if err != nil {
				t.Fatalf("recoverHermesExecutionResult() error = %v", err)
			}
			if !wait || resp.OK {
				t.Fatalf("recoverHermesExecutionResult() = resp=%+v wait=%t, want still running", resp, wait)
			}
		})
	}
}

func TestAsyncHermesExecutionCancellingDoesNotRedispatchCancel(t *testing.T) {
	cancelCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-cancelling/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		cancelCalls++
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-cancelling",
			Status:      "cancelling",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancelling",
		Status:          "cancelling",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-cancelling"},
		transition.EffectExecution{ID: "eff-cancelling"},
		"prod",
		workflowContext{},
		now,
	)
	if err != nil {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v", err)
	}
	if !wait || resp.OK {
		t.Fatalf("cancelling execution should defer without success, wait=%t resp=%+v", wait, resp)
	}
	if cancelCalls != 1 {
		t.Fatalf("cancel calls = %d, want 1", cancelCalls)
	}
}

func TestHandleClaimedExecutionEffectDefersWhenDraining(t *testing.T) {
	app.StopDrainForTest()
	defer app.StopDrainForTest()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	effect := transition.EffectExecution{
		ID:             "eff-drain-claimed",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-drain",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectRunning,
		Holder:         "worker-1",
		IdempotencyKey: "eff-drain-claimed-key",
		QueueName:      string(queue.WorkflowQueue),
		ScopeKey:       "conv-drain",
		Payload:        map[string]any{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if _, _, err := store.QueueEffectExecution(effect); err != nil {
		t.Fatalf("QueueEffectExecution() error = %v", err)
	}

	app.StartDrain()
	handleClaimedExecutionEffect(config.Config{}, store, nil, nil, effect)

	var updated transition.EffectExecution
	for _, item := range store.ListEffectExecutions() {
		if item.ID == effect.ID {
			updated = item
			break
		}
	}
	if updated.ID == "" {
		t.Fatal("expected effect to remain recorded")
	}
	if updated.Status != transition.EffectRunning || updated.Holder != "" || updated.LastError != "deployment_shutdown" || updated.NotBefore == nil || updated.LeaseExpiresAt == nil {
		t.Fatalf("claimed effect should be deferred for drain, got %+v", updated)
	}
}

func TestAsyncHermesExecutionUsesStoredTerminalRecordWithoutPolling(t *testing.T) {
	statusCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCalls++
		http.Error(w, "unexpected poll", http.StatusInternalServerError)
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	expected := clients.RunnerResponse{
		OK:       false,
		Message:  "stored terminal failure",
		Provider: "hermes-executor",
		Raw: map[string]any{
			"failure_class": workflowFailureRunnerExecutorStatusUnavailable,
		},
	}
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:  "hexec-terminal",
		Status:       "failed",
		Result:       runnerResponseMap(expected),
		FailureClass: workflowFailureRunnerExecutorStatusUnavailable,
		HeartbeatAt:  &now,
		CompletedAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{HermesExecutionHeartbeatTimeout: 120 * time.Second},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-terminal"},
		transition.EffectExecution{ID: "eff-terminal"},
		"prod",
		workflowContext{},
		now,
	)
	if err != nil {
		t.Fatalf("executeOrPollAsyncHermesExecution() error = %v", err)
	}
	if wait {
		t.Fatal("terminal stored record should not defer")
	}
	if statusCalls != 0 {
		t.Fatalf("executor status calls = %d, want 0", statusCalls)
	}
	if resp.Message != expected.Message || resp.OK {
		t.Fatalf("response = %+v, want stored failure %+v", resp, expected)
	}
}

func TestAsyncHermesExecutionCancelledTerminalResultReturnsCancellationFailure(t *testing.T) {
	statusCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCalls++
		http.Error(w, "unexpected poll", http.StatusInternalServerError)
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancelled-terminal",
		Status:          "cancelled",
		Result:          runnerResponseMap(clients.RunnerResponse{OK: true, Message: "late successful result", Provider: "hermes-executor", Raw: map[string]any{}}),
		CancelRequested: true,
		HeartbeatAt:     &now,
		CompletedAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	resp, wait, err := executeOrPollAsyncHermesExecution(
		config.Config{},
		store,
		clients.NewRunnerClient(executor.URL),
		clients.RunnerTask{ExecutionID: "hexec-cancelled-terminal"},
		transition.EffectExecution{ID: "eff-cancelled-terminal"},
		"prod",
		workflowContext{},
		now,
	)
	if wait {
		t.Fatal("cancelled terminal record should not defer")
	}
	if resp.OK {
		t.Fatalf("cancelled terminal record should not return stored success: %+v", resp)
	}
	var workflowErr *workflowFailureError
	if !errors.As(err, &workflowErr) {
		t.Fatalf("expected workflow cancellation failure, got %v", err)
	}
	if workflowErr.failure.Class != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("failure class = %q, want %q", workflowErr.failure.Class, workflowFailureRunnerExecutionCancelled)
	}
	if statusCalls != 0 {
		t.Fatalf("executor status calls = %d, want 0", statusCalls)
	}
}

func TestRunnerResponseFromMapRejectsMalformedTerminalResult(t *testing.T) {
	if resp, ok := runnerResponseFromMap(map[string]any{"status": "completed"}); ok {
		t.Fatalf("malformed result should not decode as runner response: %+v", resp)
	}
	resp, ok := runnerResponseFromMap(map[string]any{
		"ok":       false,
		"message":  "durable failure",
		"provider": "hermes-executor",
		"raw": map[string]any{
			"failure_class": workflowFailureRunnerExecutorResultUnavailable,
		},
	})
	if !ok {
		t.Fatal("expected canonical runner response to decode")
	}
	if resp.OK || resp.Message != "durable failure" || resp.Raw["failure_class"] != workflowFailureRunnerExecutorResultUnavailable {
		t.Fatalf("unexpected decoded runner response: %+v", resp)
	}
}

func TestCancelSupersededHermesExecutionsSendsCancelRequestedOnce(t *testing.T) {
	cancelCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-old/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		cancelCalls++
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-old",
			Status:      "cancelling",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-old",
		CaseID:      "case-1",
		TraceID:     "trace-old",
		Status:      "running",
		HeartbeatAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	client := clients.NewRunnerClient(executor.URL)
	cancelSupersededHermesExecutions(config.Config{}, store, client, "case-1", "trace-current")
	cancelSupersededHermesExecutions(config.Config{}, store, client, "case-1", "trace-current")

	if cancelCalls != 1 {
		t.Fatalf("cancel calls = %d, want 1", cancelCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-old")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelling" || !record.CancelRequested {
		t.Fatalf("expected cancel dispatch to move record to cancelling, got %+v", record)
	}
}

func TestCancelSupersededHermesExecutionsDispatchesCancelRequestedStatus(t *testing.T) {
	cancelCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-superseded/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		cancelCalls++
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-superseded",
			Status:      "cancelling",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-superseded",
		CaseID:          "case-1",
		TraceID:         "trace-old",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	client := clients.NewRunnerClient(executor.URL)
	cancelSupersededHermesExecutions(config.Config{}, store, client, "case-1", "trace-current")

	if cancelCalls != 1 {
		t.Fatalf("cancel calls = %d, want 1 (must dispatch cancel for cancel_requested status)", cancelCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-superseded")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancelling" || !record.CancelRequested {
		t.Fatalf("expected cancel dispatch to move record to cancelling, got %+v", record)
	}
}

func TestCancelSupersededHermesExecutionsRetriesAfterCancelRPCError(t *testing.T) {
	cancelCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-retry-cancel/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		cancelCalls++
		http.Error(w, "temporarily unavailable", http.StatusBadGateway)
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	heartbeat := now.Add(-30 * time.Second)
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-retry-cancel",
		CaseID:          "case-1",
		TraceID:         "trace-old",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &heartbeat,
		CreatedAt:       now.Add(-time.Minute),
		UpdatedAt:       heartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	client := clients.NewRunnerClient(executor.URL)
	cancelSupersededHermesExecutions(config.Config{}, store, client, "case-1", "trace-current")
	cancelSupersededHermesExecutions(config.Config{}, store, client, "case-1", "trace-current")

	if cancelCalls != 2 {
		t.Fatalf("cancel calls = %d, want 2", cancelCalls)
	}
	record, ok := store.GetRunnerExecution("hexec-retry-cancel")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "cancel_requested" || !record.CancelRequested {
		t.Fatalf("cancel RPC errors should preserve retryable cancel_requested state, got %+v", record)
	}
	if record.HeartbeatAt == nil || !record.HeartbeatAt.Equal(heartbeat) {
		t.Fatalf("cancel RPC errors must not refresh heartbeat, got %+v want %v", record.HeartbeatAt, heartbeat)
	}
}

func TestCancelSupersededHermesExecutionsTerminalWithoutResultPersistsFailure(t *testing.T) {
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/hermes-executions/hexec-old-completed/cancel" {
			t.Fatalf("unexpected executor call %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(clients.HermesExecutionStatus{
			ExecutionID: "hexec-old-completed",
			Status:      "completed",
		})
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:  "hexec-old-completed",
		CaseID:       "case-1",
		TraceID:      "trace-old",
		Status:       "running",
		FailureClass: "trace_superseded",
		HeartbeatAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	cancelSupersededHermesExecutions(config.Config{}, store, clients.NewRunnerClient(executor.URL), "case-1", "trace-current")

	record, ok := store.GetRunnerExecution("hexec-old-completed")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "failed" || record.CompletedAt == nil || record.FailureClass != "trace_superseded" {
		t.Fatalf("terminal-without-result superseded execution should persist failure result, got %+v", record)
	}
	resp, ok := runnerResponseFromMap(record.Result)
	if !ok || resp.OK || stringValue(resp.Raw["failure_class"]) != "trace_superseded" {
		t.Fatalf("expected durable result-unavailable response, ok=%t resp=%+v", ok, resp)
	}
	diagnostics := mapValue(resp.Raw["runner_diagnostics"])
	if stringValue(diagnostics["result_failure_class"]) != workflowFailureRunnerExecutorResultUnavailable {
		t.Fatalf("expected result-unavailable diagnostics, got %+v", diagnostics)
	}
}

func TestWorkflowExecutionIDIsStableForEffectRecovery(t *testing.T) {
	started := time.Date(2026, 4, 24, 8, 30, 0, 0, time.UTC)
	first := workflowExecutionID("eff-recover", started)
	second := workflowExecutionID("eff-recover", started.Add(10*time.Minute))
	if first != second {
		t.Fatalf("expected execution id to be stable for operation recovery, got %q and %q", first, second)
	}
	if first == workflowExecutionID("eff-other", started) {
		t.Fatalf("expected operation id to affect execution id")
	}
}

func TestWorkflowRunnerRecoversCompletedHermesExecutorResult(t *testing.T) {
	var expectedExecutionID string
	statusCalls := 0
	executeCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/hermes-executions/"+expectedExecutionID:
			statusCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"execution_id": expectedExecutionID,
				"status":       "completed",
				"result": map[string]any{
					"ok":       true,
					"provider": "fake-executor",
					"message":  "Recovered reply",
					"raw": map[string]any{
						"structured_output": map[string]any{
							"visible_reasoning": []any{
								map[string]any{
									"step_type":  "analysis",
									"summary":    "Recovered the existing Hermes execution result.",
									"confidence": 1.0,
									"decision":   "project_existing_result",
								},
							},
							"reply_draft":     "Recovered reply",
							"final_answer":    "Recovered reply",
							"confidence":      1.0,
							"context_summary": "Recovered from durable Hermes executor status.",
							"self_critique":   "",
							"reply_delivery": map[string]any{
								"status":       "posted",
								"channel_id":   "CENG",
								"thread_ts":    "171000001.000100",
								"body":         "Recovered reply",
								"provider_ref": "171000001.000200",
							},
							"proposed_actions":       []any{},
							"knowledge_drafts":       []any{},
							"outcome_hypotheses":     []any{},
							"produced_artifacts":     []any{},
							"completion_verdict":     "complete",
							"termination_reason":     "normal_completion",
							"artifact_render_briefs": []any{},
						},
						"completion_verdict": "complete",
						"termination_reason": "normal_completion",
					},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/internal/hermes-executions":
			executeCalls++
			t.Fatalf("unexpected duplicate Hermes executor launch")
		default:
			http.NotFound(w, r)
		}
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             executor.URL,
		HermesExecutorBaseURL:     executor.URL,
		ToolGatewayBaseURL:        "http://tool-gateway.invalid",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
		ProdRunnerTimeout:         930 * time.Second,
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	expectedExecutionID = workflowExecutionID(runnerEffect.ID, time.Now().UTC())
	started := time.Now().Add(-2 * time.Minute).UTC()
	runnerEffect.StartedAt = &started
	runnerEffect.UpdatedAt = time.Now().UTC()
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	if statusCalls != 1 {
		t.Fatalf("expected one executor status recovery call, got %d", statusCalls)
	}
	if executeCalls != 0 {
		t.Fatalf("expected no duplicate executor launches, got %d", executeCalls)
	}
	assertWorkflowEffectStatus(t, store, workflowItem.workflowID, transition.EffectInvokeRunner, transition.EffectCompleted)
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace")
	}
	if len(trace.SlackActions) != 1 || trace.SlackActions[0].FinalBody != "Recovered reply" {
		t.Fatalf("expected recovered Slack action projection, got %#v", trace.SlackActions)
	}
}

func TestWorkflowRunnerRecoveryFailsClosedOnUnknownHermesExecutorStatus(t *testing.T) {
	var expectedExecutionID string
	statusCalls := 0
	executeCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/hermes-executions/"+expectedExecutionID:
			statusCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"execution_id": expectedExecutionID,
				"status":       "paused",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/internal/hermes-executions":
			executeCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":       true,
				"provider": "fake-executor",
				"message":  "duplicate run should not launch",
				"raw": map[string]any{
					"structured_output": map[string]any{
						"final_answer": "duplicate run should not launch",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             executor.URL,
		HermesExecutorBaseURL:     executor.URL,
		ToolGatewayBaseURL:        "http://tool-gateway.invalid",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
		ProdRunnerTimeout:         930 * time.Second,
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	expectedExecutionID = workflowExecutionID(runnerEffect.ID, time.Now().UTC())
	started := time.Now().Add(-2 * time.Minute).UTC()
	runnerEffect.StartedAt = &started
	runnerEffect.UpdatedAt = time.Now().UTC()
	err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect)
	if err == nil {
		t.Fatalf("expected unknown Hermes executor status to fail closed")
	}
	var detailed *workflowFailureError
	if !errors.As(err, &detailed) {
		t.Fatalf("expected workflowFailureError, got %T: %v", err, err)
	}
	if detailed.failure.Class != workflowFailureRunnerExecutorStatusUnrecognized {
		t.Fatalf("failure class = %q, want %q", detailed.failure.Class, workflowFailureRunnerExecutorStatusUnrecognized)
	}
	if detailed.failure.RunnerDiagnostics["executor_status"] != "paused" {
		t.Fatalf("expected executor_status diagnostic, got %#v", detailed.failure.RunnerDiagnostics)
	}
	if statusCalls != 1 {
		t.Fatalf("expected one executor status recovery call, got %d", statusCalls)
	}
	if executeCalls != 0 {
		t.Fatalf("expected no duplicate executor launches, got %d", executeCalls)
	}
}

func TestWorkflowRunnerRecoveryDefersStillRunningHermesExecutor(t *testing.T) {
	var expectedExecutionID string
	statusCalls := 0
	executeCalls := 0
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/hermes-executions/"+expectedExecutionID:
			statusCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"execution_id": expectedExecutionID,
				"status":       "running",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/internal/hermes-executions":
			executeCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":       true,
				"provider": "fake-executor",
				"message":  "duplicate run should not launch",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer executor.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             executor.URL,
		HermesExecutorBaseURL:     executor.URL,
		ToolGatewayBaseURL:        "http://tool-gateway.invalid",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
		ProdRunnerTimeout:         930 * time.Second,
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	claimed := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	expectedExecutionID = workflowExecutionID(claimed.ID, time.Now().UTC())
	started := time.Now().Add(-2 * time.Minute).UTC()
	claimed.StartedAt = &started
	claimed.UpdatedAt = time.Now().UTC()

	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, claimed)

	if statusCalls != 1 {
		t.Fatalf("expected one executor status recovery call, got %d", statusCalls)
	}
	if executeCalls != 0 {
		t.Fatalf("expected no duplicate executor launches, got %d", executeCalls)
	}
	effect, ok := workflowEffectByPayload(store, workflowItem.workflowID, transition.EffectInvokeRunner, "", "")
	if !ok {
		t.Fatal("expected workflow runner effect")
	}
	if effect.Status != transition.EffectRunning {
		t.Fatalf("expected effect to remain running while deferred, got %s", effect.Status)
	}
	if effect.Holder != "" {
		t.Fatalf("expected deferred effect holder to be released, got %q", effect.Holder)
	}
	if effect.LeaseExpiresAt == nil || !effect.LeaseExpiresAt.After(time.Now().UTC()) {
		t.Fatalf("expected deferred effect lease expiry in the future, got %#v", effect.LeaseExpiresAt)
	}
	if !strings.Contains(effect.LastError, "still running") {
		t.Fatalf("expected observable still-running reason, got %q", effect.LastError)
	}
}

func TestWorkflowRunnerFailurePreservesProducedArtifacts(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       false,
			"provider": "fake",
			"message":  "direct delivery timed out after artifact render",
			"raw": map[string]any{
				"failure_class": "runner_transport_timeout",
				"structured_output": map[string]any{
					"visible_reasoning":  []any{},
					"reply_draft":        "Artifact was generated but delivery timed out.",
					"final_answer":       "Artifact was generated but delivery timed out.",
					"confidence":         0.8,
					"context_summary":    "Render completed before delivery failed.",
					"self_critique":      "",
					"proposed_actions":   []any{},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
					"produced_artifacts": []any{
						map[string]any{
							"kind":           "diagram",
							"title":          "Architecture",
							"artifact_refs":  []any{"file:///workspace/company/artifacts/diagram.html"},
							"workspace_path": "/workspace/company/artifacts/diagram.html",
							"file_ref":       "file:///workspace/company/artifacts/diagram.html",
							"size_bytes":     42,
						},
					},
				},
			},
		})
	}))
	defer runner.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             runner.URL,
		ToolGatewayBaseURL:        "http://tool-gateway.invalid",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect)

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace")
	}
	if len(trace.Artifacts) == 0 {
		t.Fatalf("expected runner failure to preserve produced artifacts")
	}
	found := false
	for _, artifact := range trace.Artifacts {
		if artifact.URL == "file:///workspace/company/artifacts/diagram.html" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected workspace artifact URL in trace artifacts, got %#v", trace.Artifacts)
	}
}

func TestWorkflowRunnerAttachesBoundSlackThreadContext(t *testing.T) {
	var executorTask clients.RunnerTask
	executor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Task clients.RunnerTask `json:"task"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		executorTask = payload.Task
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake-executor",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Recovered the parent thread request and prepared a reply.","confidence":0.93,"decision":"reply_in_thread"}],"reply_draft":"Draft reply","final_answer":"Final reply","confidence":0.93,"context_summary":"Bound Slack thread was reviewed.","self_critique":"None.","proposed_actions":[{"kind":"slack_post","target_ref":"CENG","idempotency_key":"reply-action-thread-prefetch-1","rationale":"Post the answer back into Slack."}]}`,
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":  "analysis",
							"summary":    "Recovered the parent thread request and prepared a reply.",
							"confidence": 0.93,
							"decision":   "reply_in_thread",
						},
					},
					"reply_draft":     "Draft reply",
					"final_answer":    "Final reply",
					"confidence":      0.93,
					"context_summary": "Bound Slack thread was reviewed.",
					"self_critique":   "None.",
					"proposed_actions": []any{
						map[string]any{
							"kind":            "slack_post",
							"target_ref":      "CENG",
							"idempotency_key": "reply-action-thread-prefetch-1",
							"rationale":       "Post the answer back into Slack.",
						},
					},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer executor.Close()

	toolCalls := []string{}
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		toolCalls = append(toolCalls, name)
		if name != "slack.history" {
			t.Fatalf("unexpected tool invocation %s", name)
		}
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    "slack-history-prefetch-1",
			Approved:      true,
			ApprovalState: "not_required",
			Status:        "completed",
			Available:     true,
			Provider:      "slack",
			ProviderRef:   "171000001.000100",
			Summary:       "Slack thread history loaded from D123 with 2 message(s).",
			Input:         map[string]any{"channel_id": "D123", "thread_ts": "171000001.000100"},
			Output: map[string]any{
				"channel_id": "D123",
				"thread_ts":  "171000001.000100",
				"messages": []any{
					map[string]any{
						"user":      "UALLEN",
						"username":  "Allen",
						"text":      "@Blake @Aiwei where can i see the SoT on the schema for campaigns (aka tasks)?",
						"ts":        "171000001.000100",
						"thread_ts": "171000001.000100",
					},
					map[string]any{
						"user":      "UBLAKE",
						"username":  "Blake",
						"text":      "@RSI pls help",
						"ts":        "171000002.000100",
						"thread_ts": "171000001.000100",
					},
				},
			},
		})
	}))
	defer toolGateway.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             executor.URL,
		HermesExecutorBaseURL:     executor.URL,
		ToolGatewayBaseURL:        toolGateway.URL,
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
		ProdRunnerTimeout:         930 * time.Second,
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	if !strings.Contains(executorTask.Prompt, "Bound Slack thread context is attached in the task evidence") {
		t.Fatalf("expected bound-thread prompt guidance, got %q", executorTask.Prompt)
	}
	if !strings.Contains(executorTask.ContextSummary, "Allen: @Blake @Aiwei where can i see the SoT on the schema for campaigns") {
		t.Fatalf("expected parent question in context summary, got %q", executorTask.ContextSummary)
	}
	found := false
	for _, ref := range executorTask.ContextRefs {
		if ref.ToolName != "slack.history" {
			continue
		}
		if !strings.Contains(ref.Summary, "Allen: @Blake @Aiwei where can i see the SoT on the schema for campaigns") {
			t.Fatalf("expected parent question in prefetched ref, got %#v", ref)
		}
		if ref.ChannelID != executorTask.ChannelID || ref.ThreadTS != executorTask.ThreadTS {
			t.Fatalf("expected bound Slack identifiers, got %#v", ref)
		}
		found = true
	}
	if !found {
		t.Fatalf("expected slack.history context ref, got %#v", executorTask.ContextRefs)
	}
	if !reflect.DeepEqual(toolCalls, []string{"slack.history"}) {
		t.Fatalf("expected one slack.history prefetch call, got %#v", toolCalls)
	}
}

func TestWorkflowRunnerUsesRunnerExecuteWhenHermesExecutorIsUnset(t *testing.T) {
	var runnerPath string
	var runnerTask clients.RunnerTask
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runnerPath = r.URL.Path
		if runnerPath != "/execute" {
			t.Fatalf("unexpected runner path %q", runnerPath)
		}
		var payload struct {
			Task clients.RunnerTask `json:"task"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		runnerTask = payload.Task
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake-runner",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Collected context and prepared a reply.","confidence":0.91,"decision":"reply_in_thread"}],"reply_draft":"Draft reply","final_answer":"Final reply","confidence":0.91,"context_summary":"Repo and KB context collected.","self_critique":"Follow up if channel policy changes.","proposed_actions":[{"kind":"slack_post","target_ref":"CENG","idempotency_key":"reply-action-runner-1","rationale":"Post the answer back into Slack."}]}`,
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":  "analysis",
							"summary":    "Collected context and prepared a reply.",
							"confidence": 0.91,
							"decision":   "reply_in_thread",
						},
					},
					"reply_draft":     "Draft reply",
					"final_answer":    "Final reply",
					"confidence":      0.91,
					"context_summary": "Repo and KB context collected.",
					"self_critique":   "Follow up if channel policy changes.",
					"proposed_actions": []any{
						map[string]any{
							"kind":            "slack_post",
							"target_ref":      "CENG",
							"idempotency_key": "reply-action-runner-1",
							"rationale":       "Post the answer back into Slack.",
						},
					},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             runner.URL,
		ToolGatewayBaseURL:        "http://tool-gateway.invalid",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
		ProdRunnerTimeout:         930 * time.Second,
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	if runnerPath != "/execute" {
		t.Fatalf("runner path = %q, want /execute", runnerPath)
	}
	if runnerTask.OperationID != runnerEffect.ID {
		t.Fatalf("runner operation_id = %q, want %q", runnerTask.OperationID, runnerEffect.ID)
	}
	if !strings.HasPrefix(runnerTask.ExecutionID, "hexec-") {
		t.Fatalf("runner execution_id = %q, want hexec-*", runnerTask.ExecutionID)
	}
	if runnerTask.WorkflowID != workflowItem.workflowID {
		t.Fatalf("runner workflow_id = %q, want %q", runnerTask.WorkflowID, workflowItem.workflowID)
	}
}

func TestWorkflowRunnerSystemMessageOmitsSlackMCPWhenUnavailableInNoneMode(t *testing.T) {
	message := workflowRunnerSystemMessage(false, false, "none", nil)
	if strings.Contains(message, "Use Slack MCP for Slack investigation.") {
		t.Fatalf("expected none-mode system message without Slack MCP to omit Slack MCP guidance, got %q", message)
	}
	if !strings.Contains(message, "Slack posting is blocked by policy for this workflow, so do not send any Slack messages.") {
		t.Fatalf("expected blocked-posting guidance in none-mode message, got %q", message)
	}
}

func TestWorkflowPartialCompletionPostsStandardizedReplyAndPersistsVerdict(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Recovered a partial answer.","confidence":0.72,"decision":"post_partial_reply"}],"reply_draft":"Grounded summary so far.","final_answer":"Grounded summary so far.","confidence":0.72,"context_summary":"Recovered from persisted session evidence.","self_critique":"More reads would improve coverage.","proposed_actions":[{"kind":"slack_post","target_ref":"D123","request_payload":{"body":"Grounded summary so far."},"idempotency_key":"reply-action-partial-1","rationale":"Post the grounded partial answer."}],"knowledge_drafts":[],"outcome_hypotheses":[]}`,
			"raw": map[string]any{
				"completion_verdict":     "partial",
				"termination_reason":     "iteration_budget_exhausted",
				"max_iterations_reached": true,
				"tool_calls": []any{
					map[string]any{
						"id":                     "runner-tool-record-slack-history-1",
						"tool_name":              "slack.history",
						"tool_call_id":           "slack.history:1",
						"request":                map[string]any{"channel_id": "D123", "thread_ts": "171000001.000100"},
						"summary":                "Fetched bound Slack thread history.",
						"raw_artifact_refs":      []any{"artifact://slack/history/1"},
						"approval_state":         "not_required",
						"interpretation_summary": "Fetched bound Slack thread history.",
						"status":                 "completed",
						"created_at":             "2026-04-17T19:35:00Z",
					},
				},
				"runner_diagnostics": map[string]any{
					"completion_verdict":     "partial",
					"termination_reason":     "iteration_budget_exhausted",
					"max_iterations_reached": true,
					"budget_used":            20,
					"budget_max":             20,
				},
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":  "analysis",
							"summary":    "Recovered a partial answer.",
							"confidence": 0.72,
							"decision":   "post_partial_reply",
						},
					},
					"reply_draft":     "Grounded summary so far.",
					"final_answer":    "Grounded summary so far.",
					"confidence":      0.72,
					"context_summary": "Recovered from persisted session evidence.",
					"self_critique":   "More reads would improve coverage.",
					"proposed_actions": []any{
						map[string]any{
							"kind":       "slack_post",
							"target_ref": "D123",
							"request_payload": map[string]any{
								"body": "Grounded summary so far.",
							},
							"idempotency_key": "reply-action-partial-1",
							"rationale":       "Post the grounded partial answer.",
						},
					},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	slackBodies := []string{}
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
		if name == "slack.history" {
			_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
				Name:          name,
				ToolCallID:    "slack-history-prefetch-partial",
				Approved:      true,
				ApprovalState: "not_required",
				Status:        "completed",
				Available:     true,
				Provider:      "slack",
				ProviderRef:   "171000001.000100",
				Summary:       "Slack thread history loaded.",
				Input:         map[string]any{},
				Output:        map[string]any{"messages": []any{}},
			})
			return
		}
		if name != "slack.reply" {
			t.Fatalf("unexpected tool invocation %s", name)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode slack payload: %v", err)
		}
		slackBodies = append(slackBodies, strings.TrimSpace(stringFromMap(payload, "body")))
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    "slack-post-partial-1",
			Approved:      true,
			ApprovalState: "approved",
			Status:        "completed",
			Available:     true,
			Provider:      "slack",
			ProviderRef:   "171000001.000100",
			Summary:       "Slack reply posted.",
			Input:         payload,
			Output:        map[string]any{"posted": true},
		})
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

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	replyAction := firstQueuedActionEffectByKind(t, store, "control", action.KindSlackPost)
	replyIntent, ok := store.GetActionIntent(replyAction.AggregateID)
	if !ok {
		t.Fatalf("expected reply action intent %s", replyAction.AggregateID)
	}
	if got := stringFromMap(replyIntent.RequestPayload, "workflow_reply_command"); got != string(transition.CommandReplyPostedPartial) {
		t.Fatalf("expected partial reply-posted command in action payload, got %q", got)
	}
	if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), replyAction); err != nil {
		t.Fatalf("processControlActionEffect(reply) error = %v", err)
	}

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "completed" {
		t.Fatalf("expected completed workflow, got %s", workflow.Status)
	}
	if workflow.LastVerdict != "partial" {
		t.Fatalf("expected partial workflow verdict, got %q", workflow.LastVerdict)
	}
	if workflow.RunnerDiagnostics["completion_verdict"] != "partial" {
		t.Fatalf("expected partial completion verdict in runner diagnostics, got %#v", workflow.RunnerDiagnostics)
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != events.StatusCompleted {
		t.Fatalf("expected completed trace, got %s", trace.Summary.Status)
	}
	if trace.Summary.LastVerdict != "partial" {
		t.Fatalf("expected partial trace verdict, got %q", trace.Summary.LastVerdict)
	}
	if len(trace.ToolCalls) != 1 {
		t.Fatalf("expected one projected runner tool call, got %d", len(trace.ToolCalls))
	}
	if trace.ToolCalls[0].TraceID != trace.Summary.TraceID {
		t.Fatalf("expected projected tool call trace binding, got %#v", trace.ToolCalls[0])
	}
	if trace.ToolCalls[0].WorkflowID != workflowItem.workflowID {
		t.Fatalf("expected projected tool call workflow binding, got %#v", trace.ToolCalls[0])
	}
	if trace.ToolCalls[0].ToolName != "slack.history" {
		t.Fatalf("expected slack.history tool call, got %#v", trace.ToolCalls[0])
	}
	if trace.ToolCalls[0].ToolCallID != "slack.history:1" {
		t.Fatalf("expected slack.history:1 tool call id, got %#v", trace.ToolCalls[0])
	}
	if len(trace.SlackActions) != 1 {
		t.Fatalf("expected one slack action, got %d", len(trace.SlackActions))
	}
	if !strings.HasPrefix(trace.SlackActions[0].FinalBody, partialCompletionNoticeForTerminationReason("iteration_budget_exhausted")) {
		t.Fatalf("expected partial completion notice in final body, got %q", trace.SlackActions[0].FinalBody)
	}
	if len(slackBodies) != 1 {
		t.Fatalf("expected one posted slack body, got %d", len(slackBodies))
	}
	if !strings.HasPrefix(slackBodies[0], partialCompletionNoticeForTerminationReason("iteration_budget_exhausted")) {
		t.Fatalf("expected partial completion notice in posted body, got %q", slackBodies[0])
	}
}

func TestWorkflowTaskTimeoutPartialCompletionPostsTimeoutNoticeAndPersistsVerdict(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Recovered a timeout partial answer.","confidence":0.68,"decision":"post_partial_reply"}],"reply_draft":"Grounded summary so far.","final_answer":"Grounded summary so far.","confidence":0.68,"context_summary":"Recovered from persisted session evidence after timeout.","self_critique":"More reads would improve coverage.","proposed_actions":[{"kind":"slack_post","target_ref":"D123","request_payload":{"body":"Grounded summary so far."},"idempotency_key":"reply-action-timeout-partial-1","rationale":"Post the grounded partial answer."}],"knowledge_drafts":[],"outcome_hypotheses":[]}`,
			"raw": map[string]any{
				"completion_verdict":     "partial",
				"termination_reason":     "task_timeout",
				"max_iterations_reached": false,
				"timeout_kind":           "task_timeout",
				"runner_diagnostics": map[string]any{
					"completion_verdict":     "partial",
					"termination_reason":     "task_timeout",
					"timeout_kind":           "task_timeout",
					"max_iterations_reached": false,
					"budget_used":            9,
					"budget_max":             20,
				},
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":  "analysis",
							"summary":    "Recovered a timeout partial answer.",
							"confidence": 0.68,
							"decision":   "post_partial_reply",
						},
					},
					"reply_draft":     "Grounded summary so far.",
					"final_answer":    "Grounded summary so far.",
					"confidence":      0.68,
					"context_summary": "Recovered from persisted session evidence after timeout.",
					"self_critique":   "More reads would improve coverage.",
					"proposed_actions": []any{
						map[string]any{
							"kind":       "slack_post",
							"target_ref": "D123",
							"request_payload": map[string]any{
								"body": "Grounded summary so far.",
							},
							"idempotency_key": "reply-action-timeout-partial-1",
							"rationale":       "Post the grounded partial answer.",
						},
					},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	slackBodies := []string{}
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
		if name == "slack.history" {
			_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
				Name:          name,
				ToolCallID:    "slack-history-prefetch-timeout-partial",
				Approved:      true,
				ApprovalState: "not_required",
				Status:        "completed",
				Available:     true,
				Provider:      "slack",
				ProviderRef:   "171000001.000100",
				Summary:       "Slack thread history loaded.",
				Input:         map[string]any{},
				Output:        map[string]any{"messages": []any{}},
			})
			return
		}
		if name != "slack.reply" {
			t.Fatalf("unexpected tool invocation %s", name)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode slack payload: %v", err)
		}
		slackBodies = append(slackBodies, strings.TrimSpace(stringFromMap(payload, "body")))
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    "slack-post-timeout-partial-1",
			Approved:      true,
			ApprovalState: "approved",
			Status:        "completed",
			Available:     true,
			Provider:      "slack",
			ProviderRef:   "171000001.000100",
			Summary:       "Slack reply posted.",
			Input:         payload,
			Output:        map[string]any{"posted": true},
		})
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

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	replyAction := firstQueuedActionEffectByKind(t, store, "control", action.KindSlackPost)
	replyIntent, ok := store.GetActionIntent(replyAction.AggregateID)
	if !ok {
		t.Fatalf("expected reply action intent %s", replyAction.AggregateID)
	}
	if got := stringFromMap(replyIntent.RequestPayload, "workflow_reply_command"); got != string(transition.CommandReplyPostedPartial) {
		t.Fatalf("expected partial reply-posted command in action payload, got %q", got)
	}
	if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), replyAction); err != nil {
		t.Fatalf("processControlActionEffect(reply) error = %v", err)
	}

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "completed" {
		t.Fatalf("expected completed workflow, got %s", workflow.Status)
	}
	if workflow.LastVerdict != "partial" {
		t.Fatalf("expected partial workflow verdict, got %q", workflow.LastVerdict)
	}
	if workflow.RunnerDiagnostics["completion_verdict"] != "partial" {
		t.Fatalf("expected partial completion verdict in runner diagnostics, got %#v", workflow.RunnerDiagnostics)
	}
	if workflow.RunnerDiagnostics["termination_reason"] != "task_timeout" {
		t.Fatalf("expected task_timeout termination reason in runner diagnostics, got %#v", workflow.RunnerDiagnostics)
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != events.StatusCompleted {
		t.Fatalf("expected completed trace, got %s", trace.Summary.Status)
	}
	if trace.Summary.LastVerdict != "partial" {
		t.Fatalf("expected partial trace verdict, got %q", trace.Summary.LastVerdict)
	}
	if len(trace.SlackActions) != 1 {
		t.Fatalf("expected one slack action, got %d", len(trace.SlackActions))
	}
	if !strings.HasPrefix(trace.SlackActions[0].FinalBody, partialCompletionNoticeForTerminationReason("task_timeout")) {
		t.Fatalf("expected timeout partial completion notice in final body, got %q", trace.SlackActions[0].FinalBody)
	}
	if len(slackBodies) != 1 {
		t.Fatalf("expected one posted slack body, got %d", len(slackBodies))
	}
	if !strings.HasPrefix(slackBodies[0], partialCompletionNoticeForTerminationReason("task_timeout")) {
		t.Fatalf("expected timeout partial completion notice in posted body, got %q", slackBodies[0])
	}
}

func TestWorkflowPartialCompletionBlockedByPolicyMovesNeedsHuman(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Recovered a partial answer.","confidence":0.72,"decision":"post_partial_reply"}],"reply_draft":"Grounded summary so far.","final_answer":"Grounded summary so far.","confidence":0.72,"context_summary":"Recovered from persisted session evidence.","self_critique":"More reads would improve coverage.","proposed_actions":[{"kind":"slack_post","target_ref":"C999","request_payload":{"channel_id":"C999","thread_ts":"171000001.000100","body":"Grounded summary so far."},"idempotency_key":"reply-action-partial-blocked","rationale":"Post the grounded partial answer."}],"knowledge_drafts":[],"outcome_hypotheses":[]}`,
			"raw": map[string]any{
				"completion_verdict":     "partial",
				"termination_reason":     "iteration_budget_exhausted",
				"max_iterations_reached": true,
				"runner_diagnostics": map[string]any{
					"completion_verdict":     "partial",
					"termination_reason":     "iteration_budget_exhausted",
					"max_iterations_reached": true,
					"budget_used":            20,
					"budget_max":             20,
				},
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":  "analysis",
							"summary":    "Recovered a partial answer.",
							"confidence": 0.72,
							"decision":   "post_partial_reply",
						},
					},
					"reply_draft":     "Grounded summary so far.",
					"final_answer":    "Grounded summary so far.",
					"confidence":      0.72,
					"context_summary": "Recovered from persisted session evidence.",
					"self_critique":   "More reads would improve coverage.",
					"proposed_actions": []any{
						map[string]any{
							"kind":       "slack_post",
							"target_ref": "C999",
							"request_payload": map[string]any{
								"channel_id": "C999",
								"thread_ts":  "171000001.000100",
								"body":       "Grounded summary so far.",
							},
							"idempotency_key": "reply-action-partial-blocked",
							"rationale":       "Post the grounded partial answer.",
						},
					},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	slackPosts := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
		if name == "slack.history" {
			_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
				Name:          name,
				ToolCallID:    "slack-history-prefetch-blocked-partial",
				Approved:      true,
				ApprovalState: "not_required",
				Status:        "completed",
				Available:     true,
				Provider:      "slack",
				ProviderRef:   "171000001.000100",
				Summary:       "Slack thread history loaded.",
				Input:         map[string]any{},
				Output:        map[string]any{"messages": []any{}},
			})
			return
		}
		if name == "slack.reply" {
			slackPosts++
		}
		t.Fatalf("unexpected tool invocation %s", name)
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

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	replyAction := firstQueuedActionEffectByKind(t, store, "control", action.KindSlackPost)
	if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), replyAction); err != nil {
		t.Fatalf("processControlActionEffect(reply) error = %v", err)
	}

	if slackPosts != 0 {
		t.Fatalf("expected no live slack post for blocked partial completion, got %d", slackPosts)
	}

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "needs_human" {
		t.Fatalf("expected needs_human workflow, got %s", workflow.Status)
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != events.StatusNeedsHuman {
		t.Fatalf("expected needs_human trace, got %s", trace.Summary.Status)
	}
	if len(trace.SlackActions) != 1 {
		t.Fatalf("expected one slack action record, got %d", len(trace.SlackActions))
	}
	if trace.SlackActions[0].SendStatus != "channel_policy_missing" {
		t.Fatalf("expected channel_policy_missing send status, got %q", trace.SlackActions[0].SendStatus)
	}
}

func TestWorkflowNativeMCPReplyDeliveryCompletesWithoutQueuedSlackReply(t *testing.T) {
	var runnerRequest map[string]any
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&runnerRequest); err != nil {
			t.Fatalf("decode runner request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "openai",
			"message":  `{"visible_reasoning":[],"reply_draft":"Final reply from Slack MCP.","final_answer":"Final reply from Slack MCP.","confidence":0.88,"context_summary":"Grounded in Slack MCP evidence.","self_critique":"","proposed_actions":[],"knowledge_drafts":[],"outcome_hypotheses":[],"produced_artifacts":[{"kind":"diagram","artifact_refs":["artifact://structured-diagram"]}]}`,
			"raw": map[string]any{
				"native_mcp_enabled": true,
				"reply_delivery": map[string]any{
					"channel_id":   "D123",
					"thread_ts":    "171000001.000100",
					"body":         "Final reply from Slack MCP.",
					"body_sha1":    "delivery-sha1",
					"body_excerpt": "Final reply from Slack MCP.",
					"tool_call_id": "mcp-send-1",
					"tool_name":    "slack.mcp.send_message",
					"provider_ref": "171000001.000100",
					"send_status":  "posted",
					"artifact_refs": []any{
						"artifact://delivery-proof",
					},
				},
				"tool_calls": []any{
					map[string]any{
						"id":             "runner-tool-record-mcp-send-1",
						"tool_name":      "slack.mcp.send_message",
						"tool_call_id":   "mcp-send-1",
						"request":        map[string]any{"channel_id": "D123", "thread_ts": "171000001.000100"},
						"summary":        "Posted Slack reply through MCP.",
						"status":         "completed",
						"created_at":     "2026-04-18T20:00:00Z",
						"approval_state": "not_required",
					},
				},
				"structured_output": map[string]any{
					"visible_reasoning": []any{},
					"reply_draft":       "Final reply from Slack MCP.",
					"final_answer":      "Final reply from Slack MCP.",
					"confidence":        0.88,
					"context_summary":   "Grounded in Slack MCP evidence.",
					"self_critique":     "",
					"produced_artifacts": []any{
						map[string]any{
							"kind": "diagram",
							"artifact_refs": []any{
								"artifact://structured-diagram",
							},
						},
					},
					"proposed_actions":   []any{},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
				"runner_diagnostics": map[string]any{
					"native_execution_mode":    "openai_responses_mcp",
					"native_mcp_enabled":       true,
					"reply_delivery_attempted": true,
				},
			},
		})
	}))
	defer runner.Close()

	toolGatewayCalls := []string{}
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		toolGatewayCalls = append(toolGatewayCalls, name)
		if name != "slack.history" {
			t.Fatalf("unexpected tool gateway invocation %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    "slack-history-prefetch-direct",
			Approved:      true,
			ApprovalState: "not_required",
			Status:        "completed",
			Available:     true,
			Provider:      "slack",
			ProviderRef:   "171000001.000100",
			Summary:       "Slack thread history loaded.",
			Input:         map[string]any{},
			Output:        map[string]any{"messages": []any{}},
		})
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

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	taskPayload := mapValue(runnerRequest["task"])
	mcpServers, ok := taskPayload["mcp_servers"].([]any)
	if len(mcpServers) != 1 {
		t.Fatalf("expected one MCP server in runner request, got %#v", mcpServers)
	}
	if got := stringFromMap(mapValue(mcpServers[0]), "profile"); got != "slack_mcp_reply" {
		t.Fatalf("expected slack_mcp_reply profile, got %q", got)
	}
	if got := stringFromMap(taskPayload, "reply_delivery_mode"); got != "direct" {
		t.Fatalf("expected direct reply delivery mode, got %q", got)
	}
	if !reflect.DeepEqual(toolGatewayCalls, []string{"slack.history"}) {
		t.Fatalf("expected only slack.history prefetch call, got %#v", toolGatewayCalls)
	}
	if queued := queuedActionEffectsForPlane(store, "control"); len(queued) != 0 {
		t.Fatalf("expected no queued control actions, got %#v", queued)
	}
	if intents := store.ListActionIntents(); len(intents) != 0 {
		t.Fatalf("expected no action intents, got %#v", intents)
	}

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "completed" {
		t.Fatalf("expected completed workflow, got %s", workflow.Status)
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != events.StatusCompleted {
		t.Fatalf("expected completed trace, got %s", trace.Summary.Status)
	}
	if len(trace.SlackActions) != 1 {
		t.Fatalf("expected one Slack action record, got %d", len(trace.SlackActions))
	}
	if trace.SlackActions[0].ID != "mcp-send-1" {
		t.Fatalf("expected reply_delivery tool call id to persist as action id, got %#v", trace.SlackActions[0])
	}
	if trace.SlackActions[0].IdempotencyKey != "delivery-sha1" {
		t.Fatalf("expected reply body digest to persist as idempotency key, got %#v", trace.SlackActions[0])
	}
	if trace.SlackActions[0].FinalBody != "Final reply from Slack MCP." {
		t.Fatalf("expected final body to persist, got %#v", trace.SlackActions[0])
	}
	if trace.SlackActions[0].SendStatus != "posted" {
		t.Fatalf("expected posted send status, got %#v", trace.SlackActions[0])
	}
	if !reflect.DeepEqual(trace.SlackActions[0].ArtifactRefs, []string{"artifact://delivery-proof", "artifact://structured-diagram"}) {
		t.Fatalf("expected merged delivery and structured artifact refs, got %#v", trace.SlackActions[0].ArtifactRefs)
	}
}

func TestWorkflowNativeMCPMissingReplyDeliveryMovesNeedsHuman(t *testing.T) {
	var runnerRequest map[string]any
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&runnerRequest); err != nil {
			t.Fatalf("decode runner request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "openai",
			"message":  `{"visible_reasoning":[],"reply_draft":"Final reply from Slack MCP.","final_answer":"Final reply from Slack MCP.","confidence":0.71,"context_summary":"Grounded answer.","self_critique":"","proposed_actions":[],"knowledge_drafts":[],"outcome_hypotheses":[]}`,
			"raw": map[string]any{
				"native_mcp_enabled": true,
				"structured_output": map[string]any{
					"visible_reasoning":  []any{},
					"reply_draft":        "Final reply from Slack MCP.",
					"final_answer":       "Final reply from Slack MCP.",
					"confidence":         0.71,
					"context_summary":    "Grounded answer.",
					"self_critique":      "",
					"proposed_actions":   []any{},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
				"runner_diagnostics": map[string]any{
					"native_execution_mode": "openai_responses_mcp",
					"native_mcp_enabled":    true,
				},
			},
		})
	}))
	defer runner.Close()

	toolGatewayCalls := []string{}
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		toolGatewayCalls = append(toolGatewayCalls, name)
		if name != "slack.history" {
			t.Fatalf("unexpected tool gateway invocation %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    "slack-history-prefetch-missing-reply-delivery",
			Approved:      true,
			ApprovalState: "not_required",
			Status:        "completed",
			Available:     true,
			Provider:      "slack",
			ProviderRef:   "171000001.000100",
			Summary:       "Slack thread history loaded.",
			Input:         map[string]any{},
			Output:        map[string]any{"messages": []any{}},
		})
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

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	taskPayload := mapValue(runnerRequest["task"])
	mcpServers, ok := taskPayload["mcp_servers"].([]any)
	if len(mcpServers) != 1 {
		t.Fatalf("expected one MCP server in runner request, got %#v", mcpServers)
	}
	if got := stringFromMap(mapValue(mcpServers[0]), "profile"); got != "slack_mcp_reply" {
		t.Fatalf("expected slack_mcp_reply profile, got %q", got)
	}
	if got := stringFromMap(taskPayload, "reply_delivery_mode"); got != "direct" {
		t.Fatalf("expected direct reply delivery mode, got %q", got)
	}
	if !reflect.DeepEqual(toolGatewayCalls, []string{"slack.history"}) {
		t.Fatalf("expected only slack.history prefetch call, got %#v", toolGatewayCalls)
	}
	if queued := queuedActionEffectsForPlane(store, "control"); len(queued) != 0 {
		t.Fatalf("expected no queued control actions, got %#v", queued)
	}

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "needs_human" {
		t.Fatalf("expected needs_human workflow, got %s", workflow.Status)
	}
	if workflow.FailureClass != "missing_reply_delivery" {
		t.Fatalf("expected missing_reply_delivery failure class, got %#v", workflow)
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != events.StatusNeedsHuman {
		t.Fatalf("expected needs_human trace, got %s", trace.Summary.Status)
	}
	if len(trace.SlackActions) != 0 {
		t.Fatalf("expected no Slack actions when reply_delivery is missing, got %#v", trace.SlackActions)
	}
}

func TestWorkflowEmptyReplyDeliveryMovesNeedsHuman(t *testing.T) {
	var runnerRequest map[string]any
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&runnerRequest); err != nil {
			t.Fatalf("decode runner request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "openai",
			"message":  `{"visible_reasoning":[],"reply_draft":"Final reply from Slack MCP.","final_answer":"Final reply from Slack MCP.","confidence":0.71,"context_summary":"Grounded answer.","self_critique":"","proposed_actions":[],"reply_delivery":{},"knowledge_drafts":[],"outcome_hypotheses":[]}`,
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning":  []any{},
					"reply_draft":        "Final reply from Slack MCP.",
					"final_answer":       "Final reply from Slack MCP.",
					"confidence":         0.71,
					"context_summary":    "Grounded answer.",
					"self_critique":      "",
					"proposed_actions":   []any{},
					"reply_delivery":     map[string]any{},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
				"runner_diagnostics": map[string]any{},
			},
		})
	}))
	defer runner.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             runner.URL,
		ToolGatewayBaseURL:        "http://unused.invalid",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	taskPayload := mapValue(runnerRequest["task"])
	if got := stringFromMap(taskPayload, "reply_delivery_mode"); got != "direct" {
		t.Fatalf("expected direct reply delivery mode, got %q", got)
	}

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "needs_human" {
		t.Fatalf("expected needs_human workflow, got %s", workflow.Status)
	}
	if workflow.FailureClass != "missing_reply_delivery" {
		t.Fatalf("expected missing_reply_delivery failure class, got %#v", workflow)
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if len(trace.SlackActions) != 0 {
		t.Fatalf("expected no Slack actions for empty reply_delivery, got %#v", trace.SlackActions)
	}
}

func TestWorkflowFailedReplyDeliveryPersistsSlackAction(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "openai",
			"message":  `{"visible_reasoning":[],"reply_draft":"Final reply.","final_answer":"Final reply.","confidence":0.71,"context_summary":"Grounded answer.","self_critique":"","proposed_actions":[],"reply_delivery":{"channel_id":"D123","thread_ts":"171000002.000100","body":"Final reply.","body_sha1":"delivery-sha1","tool_call_id":"mcp-send-1","send_status":"channel_not_found"},"knowledge_drafts":[],"outcome_hypotheses":[]}`,
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning":  []any{},
					"reply_draft":        "Final reply.",
					"final_answer":       "Final reply.",
					"confidence":         0.71,
					"context_summary":    "Grounded answer.",
					"self_critique":      "",
					"proposed_actions":   []any{},
					"reply_delivery":     map[string]any{"channel_id": "D123", "thread_ts": "171000002.000100", "body": "Final reply.", "body_sha1": "delivery-sha1", "tool_call_id": "mcp-send-1", "send_status": "channel_not_found"},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{
		ServiceName:               "control-plane",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		RunnerBaseURL:             runner.URL,
		ToolGatewayBaseURL:        "http://unused.invalid",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{"prod": clients.NewRunnerClient(cfg.RunnerBaseURL)}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "needs_human" || workflow.FailureClass != "reply_delivery_failed" {
		t.Fatalf("expected reply delivery failure needs_human workflow, got %#v", workflow)
	}
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if len(trace.SlackActions) != 1 || trace.SlackActions[0].SendStatus != "channel_not_found" {
		t.Fatalf("expected failed delivery Slack action to persist, got %#v", trace.SlackActions)
	}
}

func TestWorkflowNativeMCPReplyDeliveryUncertainMovesNeedsHuman(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       false,
			"provider": "openai",
			"message":  "reply delivery is uncertain after a Slack MCP write attempt",
			"raw": map[string]any{
				"failure_class":      "runner_reply_delivery_uncertain",
				"native_mcp_enabled": true,
				"completion_verdict": "complete",
				"termination_reason": "normal_completion",
				"reply_delivery": map[string]any{
					"channel_id":   "D123",
					"thread_ts":    "171000001.000100",
					"body":         "Final reply from Slack MCP.",
					"body_sha1":    "delivery-sha1",
					"body_excerpt": "Final reply from Slack MCP.",
					"tool_call_id": "mcp-send-1",
					"tool_name":    "slack.mcp.send_message",
					"provider_ref": "171000001.000100",
					"send_status":  "posted",
				},
				"tool_calls": []any{
					map[string]any{
						"id":             "runner-tool-record-mcp-send-1",
						"tool_name":      "slack.mcp.send_message",
						"tool_call_id":   "mcp-send-1",
						"request":        map[string]any{"channel_id": "D123", "thread_ts": "171000001.000100"},
						"summary":        "Posted Slack reply through MCP.",
						"status":         "completed",
						"created_at":     "2026-04-18T20:00:00Z",
						"approval_state": "not_required",
					},
				},
				"runner_diagnostics": map[string]any{
					"native_execution_mode":    "openai_responses_mcp",
					"native_mcp_enabled":       true,
					"reply_delivery_attempted": true,
				},
			},
		})
	}))
	defer runner.Close()

	toolGatewayCalls := []string{}
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		toolGatewayCalls = append(toolGatewayCalls, name)
		if name != "slack.history" {
			t.Fatalf("unexpected tool gateway invocation %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    "slack-history-prefetch-uncertain-delivery",
			Approved:      true,
			ApprovalState: "not_required",
			Status:        "completed",
			Available:     true,
			Provider:      "slack",
			ProviderRef:   "171000001.000100",
			Summary:       "Slack thread history loaded.",
			Input:         map[string]any{},
			Output:        map[string]any{"messages": []any{}},
		})
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

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}

	if !reflect.DeepEqual(toolGatewayCalls, []string{"slack.history"}) {
		t.Fatalf("expected only slack.history prefetch call, got %#v", toolGatewayCalls)
	}
	if queued := queuedActionEffectsForPlane(store, "control"); len(queued) != 0 {
		t.Fatalf("expected no queued control actions, got %#v", queued)
	}

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "needs_human" {
		t.Fatalf("expected needs_human workflow, got %s", workflow.Status)
	}
	if workflow.FailureClass != "runner_reply_delivery_uncertain" {
		t.Fatalf("expected runner_reply_delivery_uncertain failure class, got %#v", workflow)
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != events.StatusNeedsHuman {
		t.Fatalf("expected needs_human trace, got %s", trace.Summary.Status)
	}
	if len(trace.SlackActions) != 1 {
		t.Fatalf("expected one persisted uncertain Slack action, got %#v", trace.SlackActions)
	}
	if trace.SlackActions[0].ID != "mcp-send-1" {
		t.Fatalf("expected uncertain reply_delivery action id, got %#v", trace.SlackActions[0])
	}
	if trace.SlackActions[0].IdempotencyKey != "delivery-sha1" {
		t.Fatalf("expected uncertain reply body digest idempotency key, got %#v", trace.SlackActions[0])
	}
	if trace.SlackActions[0].SendStatus != "posted" {
		t.Fatalf("expected persisted send status on uncertain delivery, got %#v", trace.SlackActions[0])
	}
}

func TestPartialCompletionHelpersCoverOutputTokenBudgetExhaustion(t *testing.T) {
	if got := partialCompletionNoticeForTerminationReason("output_token_budget_exhausted"); got != partialCompletionNoticeOutputBudget {
		t.Fatalf("unexpected output-budget notice: %q", got)
	}
	if got := partialCompletionReasoningSummary("output_token_budget_exhausted"); !strings.Contains(got, "response output budget") {
		t.Fatalf("unexpected output-budget reasoning summary: %q", got)
	}
	if got := partialCompletionRunnerDescription("output_token_budget_exhausted", true); !strings.Contains(got, "response output budget") {
		t.Fatalf("unexpected output-budget runner description: %q", got)
	}
}

func TestSupersededTraceDoesNotPostLateSlackReply(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  `{"visible_reasoning":[{"step_type":"analysis","summary":"Collected context and prepared a reply.","confidence":0.91,"decision":"reply_in_thread"}],"reply_draft":"Draft reply","final_answer":"Final reply","confidence":0.91,"proposed_actions":[{"kind":"slack_post","target_ref":"CENG","idempotency_key":"reply-action-2","rationale":"Post the answer back into Slack."}]}`,
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":  "analysis",
							"summary":    "Collected context and prepared a reply.",
							"confidence": 0.91,
							"decision":   "reply_in_thread",
						},
					},
					"reply_draft":  "Draft reply",
					"final_answer": "Final reply",
					"confidence":   0.91,
					"proposed_actions": []any{
						map[string]any{
							"kind":            "slack_post",
							"target_ref":      "CENG",
							"idempotency_key": "reply-action-2",
							"rationale":       "Post the answer back into Slack.",
						},
					},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	slackPosts := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
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
	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	for _, item := range queuedActionEffectsForPlane(store, "control") {
		if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), item); err != nil {
			t.Fatalf("processControlActionEffect(context) error = %v", err)
		}
	}
	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}
	oldReplyAction := firstQueuedActionEffectByKind(t, store, "control", action.KindSlackPost)

	_, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineIngress,
		AggregateID: "slack:slack:CENG:171000099.000100",
		CommandKind: string(transition.CommandIngressRecordEvent),
		CommandID:   "cmd-test-ingress-newer-evidence",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"source":                       string(ingestion.SourceSlack),
			"source_event_id":              "slack-171000099.000100",
			"thread_key":                   "slack:CENG:171000001.000100",
			"dedupe_key":                   "slack:CENG:171000099.000100",
			"severity":                     string(ingestion.SeverityWarning),
			"normalized_problem_statement": "Investigate why staging homepage is failing and propose a fix with newer evidence.",
			"ownership_hint":               "platform",
			"raw_payload_ref":              "memory://slack/CENG/171000099-000100.json",
			"workflow_hint":                "incident",
			"metadata": map[string]any{
				"channel_id": "CENG",
				"user_id":    "U123",
				"thread_ts":  "171000001.000100",
			},
			"created_at": time.Now().UTC(),
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(ingress_record_event) error = %v", err)
	}

	if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), oldReplyAction); err != nil {
		t.Fatalf("processControlActionEffect(old reply) error = %v", err)
	}
	if slackPosts != 0 {
		t.Fatalf("expected superseded reply to not post to Slack, got %d calls", slackPosts)
	}
}

func TestControlActionPersistenceFailureFinalizesTraceAndQueuesEval(t *testing.T) {
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
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
	if err := startWorkflowViaCommand(cfg, baseStore, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	ctx, err := loadWorkflowContext(baseStore, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}
	intent, created, err := ensureActionIntent(baseStore, action.Intent{
		OwnerPlane:     "control",
		ConversationID: ctx.trace.Summary.ConversationID,
		CaseID:         ctx.trace.Summary.CaseID,
		TraceID:        ctx.trace.Summary.TraceID,
		Kind:           action.KindSlackPost,
		PhaseKey:       controlPhaseReplyPost,
		TargetRef:      ctx.ingestion.ChannelID,
		RequestPayload: map[string]any{
			"channel_id":   ctx.ingestion.ChannelID,
			"thread_ts":    ctx.ingestion.ThreadTS,
			"body":         "Draft reply",
			"draft_body":   "Draft reply",
			"final_body":   "Draft reply",
			"resume_queue": string(queue.WorkflowQueue),
		},
		IdempotencyKey: "reply-persistence-failure",
		ApprovalMode:   ctx.workflow.ApprovalMode,
		ApprovalState:  "approved",
		PolicyVerdict:  "allowed",
		Status:         action.StatusQueued,
		RequestedBy:    cfg.ServiceName,
		Rationale:      "Post the reply back to Slack.",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("ensureActionIntent() error = %v", err)
	}
	if !created {
		t.Fatal("expected reply action intent to be created")
	}
	failingIntentID := intent.ID
	failingAction := firstQueuedActionEffectByKind(t, baseStore, "control", action.KindSlackPost)
	store := &failingActionCommandStore{
		Store:        baseStore,
		FailActionID: failingIntentID,
		Err: &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "action_result_pkey",
			TableName:      "action_result",
			Message:        `duplicate key value violates unique constraint "action_result_pkey"`,
		},
	}

	err = processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), failingAction)
	if err == nil {
		t.Fatal("expected persistence failure to bubble up")
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

	trace, ok := baseStore.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	workflow, ok := findWorkflow(baseStore.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "needs_human" {
		t.Fatalf("expected workflow to move to needs_human immediately, got %s", workflow.Status)
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
	trace, _ = baseStore.GetTrace(workflowItem.traceID)
	if trace.Summary.Status != "needs-human" {
		t.Fatalf("expected terminal needs-human trace, got %s", trace.Summary.Status)
	}
	if trace.Summary.EndedAt.IsZero() {
		t.Fatal("expected terminal trace ended_at to be set")
	}
	workflow, ok = findWorkflow(baseStore.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist after reply resume")
	}
	if workflow.Status != "needs_human" {
		t.Fatalf("expected workflow to remain needs_human, got %s", workflow.Status)
	}

	if len(baseStore.ListEvalRuns()) == 0 {
		t.Fatal("expected workflow finalization to trigger immediate problem-line evaluation")
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

func TestReplyPolicyAllowsActiveIngressThreadWithoutChannelPolicy(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineIngress,
		AggregateID: "slack:171000002.000100",
		CommandKind: string(transition.CommandIngressRecordSlack),
		CommandID:   "cmd-reply-policy-active-thread",
		Actor:       "tester",
		OccurredAt:  now,
		Payload: map[string]any{
			"bot_role":   "orchestrator",
			"team_id":    "T123",
			"channel_id": "C123",
			"thread_ts":  "171000002.000100",
			"user_id":    "U123",
			"text":       "Hello <@U0ASDQKU3UL>, can you look at this?",
			"ts":         "171000002.000100",
			"created_at": now,
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(slack ingress) error = %v", err)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), receipt.ResultRef)
	if !ok {
		t.Fatalf("expected ingestion %s", receipt.ResultRef)
	}

	allowed, verdict := replyPolicy(store, "architecture", ingestion.ThreadKey, ingestion.ChannelID)
	if !allowed {
		t.Fatal("expected active ingress thread to allow reply_in_thread")
	}
	if verdict != "thread_allowed" {
		t.Fatalf("expected thread_allowed verdict, got %s", verdict)
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

func TestExecuteSlackPostActionIntentClaimsMatchingReplyEffect(t *testing.T) {
	slackPosts := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
		if name != "slack.reply" {
			t.Fatalf("unexpected tool invocation %s", name)
		}
		slackPosts++
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:          name,
			ToolCallID:    "slack-post-match",
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
	}))
	defer toolGateway.Close()

	baseStore := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, baseStore, "slack:")
	ctx, err := loadWorkflowContext(baseStore, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}

	now := time.Now().UTC()
	if _, err := baseStore.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: "action-reply-match",
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   "cmd-effect-selection-queue",
		OccurredAt:  now,
		Payload: map[string]any{
			"owner_plane":     "control",
			"conversation_id": ctx.trace.Summary.ConversationID,
			"case_id":         ctx.trace.Summary.CaseID,
			"trace_id":        ctx.trace.Summary.TraceID,
			"kind":            string(action.KindSlackPost),
			"phase_key":       controlPhaseReplyPost,
			"target_ref":      ctx.ingestion.ChannelID,
			"request_payload": map[string]any{
				"channel_id": ctx.ingestion.ChannelID,
				"thread_ts":  ctx.ingestion.ThreadTS,
				"body":       "Final reply",
			},
			"idempotency_key": "reply-effect-match",
			"approval_mode":   "not_required",
			"approval_state":  "approved",
			"requested_by":    "control-plane",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_queued) error = %v", err)
	}
	if _, err := baseStore.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: "action-reply-match",
		CommandKind: string(transition.CommandActionStart),
		CommandID:   "cmd-effect-selection-start",
		OccurredAt:  now,
		Payload: map[string]any{
			"operation_id": "reply-effect-match-op",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_started) error = %v", err)
	}
	intent, ok := baseStore.GetActionIntent("action-reply-match")
	if !ok {
		t.Fatalf("expected action intent action-reply-match")
	}

	store := &effectSelectionStore{
		Store: baseStore,
		effects: []transition.EffectExecution{
			{
				ID:          "eff-other",
				MachineKind: transition.MachineWorkflow,
				AggregateID: ctx.workflow.ID,
				EffectKind:  transition.EffectPostSlackReply,
				Status:      transition.EffectQueued,
				Payload: map[string]any{
					"reply_action_id": "action-other",
				},
				CreatedAt: now,
				UpdatedAt: now.Add(time.Minute),
			},
			{
				ID:          "eff-match",
				MachineKind: transition.MachineWorkflow,
				AggregateID: ctx.workflow.ID,
				EffectKind:  transition.EffectPostSlackReply,
				Status:      transition.EffectQueued,
				Payload: map[string]any{
					"reply_action_id": intent.ID,
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	cfg := config.Config{
		ServiceName:        "control-plane",
		ToolGatewayBaseURL: toolGateway.URL,
	}
	if err := executeSlackPostActionIntent(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), ctx, intent); err != nil {
		t.Fatalf("executeSlackPostActionIntent() error = %v", err)
	}

	if slackPosts != 1 {
		t.Fatalf("expected a single Slack post, got %d", slackPosts)
	}
	if len(store.claimed) != 1 || store.claimed[0] != "eff-match" {
		t.Fatalf("expected matching reply effect to be claimed, got %#v", store.claimed)
	}
	if len(store.completed) != 1 || store.completed[0] != "eff-match" {
		t.Fatalf("expected matching reply effect to complete, got %#v", store.completed)
	}
	if len(store.failed) != 0 {
		t.Fatalf("expected no failed reply effects, got %#v", store.failed)
	}
}

func TestClaimNextActionEffectFairClaimOnlyUsesActionQueue(t *testing.T) {
	store := &claimQueueCaptureStore{Store: storepkg.NewMemoryStore()}

	_, ok, err := claimNextActionEffect(
		config.Config{EffectFairClaimEnabled: true, EffectMaxConcurrentPerScope: 1},
		store,
		"control",
		"worker-1",
		time.Minute,
	)
	if err != nil {
		t.Fatalf("claimNextActionEffect() error = %v", err)
	}
	if ok {
		t.Fatal("capture store should not return a claim")
	}
	if len(store.queueNames) != 1 || store.queueNames[0] != "action" {
		t.Fatalf("action fair claim queues = %#v, want only action queue", store.queueNames)
	}
}

func TestFinalizeWorkflowFailureQueuesEvalForFailedTrace(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	cfg := config.Config{ServiceName: "control-plane"}

	if err := finalizeWorkflowFailure(cfg, store, workflowItem, errors.New("runner response missing structured_output")); err != nil {
		t.Fatalf("finalizeWorkflowFailure() error = %v", err)
	}

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "failed" {
		t.Fatalf("expected failed workflow state, got %s", workflow.Status)
	}
	if workflow.LastError != "runner response missing structured_output" {
		t.Fatalf("expected workflow last error to persist, got %q", workflow.LastError)
	}
	if trace.Summary.Status != events.StatusFailed {
		t.Fatalf("expected failed trace, got %s", trace.Summary.Status)
	}

	foundWorkflowFailed := false
	for _, event := range trace.Events {
		switch event.EventType {
		case "workflow.failed":
			foundWorkflowFailed = true
		}
	}
	if !foundWorkflowFailed {
		t.Fatal("expected workflow.failed event to be recorded")
	}
	if len(store.ListEvalRuns()) == 0 {
		t.Fatal("expected failed workflow to trigger immediate problem-line evaluation")
	}
}

func TestHandleClaimedWorkflowRunnerEffectFinalizesStructuredOutputFailure(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  "runner returned prose only",
			"raw": map[string]any{
				"tool_calls": []any{
					map[string]any{
						"id":                     "runner-tool-record-rsi-workflow-context-1",
						"tool_name":              "rsi.workflow_context",
						"tool_call_id":           "rsi.workflow_context:1",
						"request":                map[string]any{"trace_id": "trace-seeded"},
						"summary":                "Fetched workflow context before malformed final output.",
						"raw_artifact_refs":      []any{"artifact://workflow/context/1"},
						"approval_state":         "not_required",
						"interpretation_summary": "Fetched workflow context before malformed final output.",
						"status":                 "completed",
						"created_at":             "2026-04-17T19:40:00Z",
					},
				},
			},
		})
	}))
	defer runner.Close()

	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
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

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	for _, item := range queuedActionEffectsForPlane(store, "control") {
		if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), item); err != nil {
			t.Fatalf("processControlActionEffect(context) error = %v", err)
		}
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect)

	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected failed trace")
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "failed" {
		t.Fatalf("expected workflow to fail, got %s", workflow.Status)
	}
	if workflow.LastError != "runner response missing structured_output" {
		t.Fatalf("expected missing structured_output error, got %q", workflow.LastError)
	}
	if trace.Summary.Status != events.StatusFailed {
		t.Fatalf("expected failed trace, got %s", trace.Summary.Status)
	}
	if len(trace.ToolCalls) != 1 {
		t.Fatalf("expected one projected runner tool call on failure, got %d", len(trace.ToolCalls))
	}
	if trace.ToolCalls[0].TraceID != trace.Summary.TraceID {
		t.Fatalf("expected failure tool call trace binding, got %#v", trace.ToolCalls[0])
	}
	if trace.ToolCalls[0].WorkflowID != workflowItem.workflowID {
		t.Fatalf("expected failure tool call workflow binding, got %#v", trace.ToolCalls[0])
	}
	if trace.ToolCalls[0].ToolName != "rsi.workflow_context" {
		t.Fatalf("expected rsi.workflow_context tool call, got %#v", trace.ToolCalls[0])
	}
	if len(store.ListEvalRuns()) == 0 {
		t.Fatal("expected failed runner response to queue eval")
	}
	assertWorkflowEffectStatus(t, store, workflowItem.workflowID, transition.EffectInvokeRunner, transition.EffectCompleted)
}

func TestHandleClaimedWorkflowRunnerEffectWorkflowCommandPersistenceFailureFailsWorkflow(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  "Final reply",
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":    "analysis",
							"summary":      "Collected context.",
							"alternatives": []any{},
							"confidence":   0.9,
							"decision":     "reply_in_thread",
						},
					},
					"reply_draft":        "Draft reply",
					"final_answer":       "Final reply",
					"confidence":         0.9,
					"context_summary":    "Context collected.",
					"self_critique":      "",
					"proposed_actions":   []any{map[string]any{"kind": "slack_post", "rationale": "Reply in thread.", "target_ref": "D123"}},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
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
	workflowItem := firstQueuedWorkflowItem(t, baseStore, "slack:")
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

	if err := startWorkflowViaCommand(cfg, baseStore, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	for _, item := range queuedActionEffectsForPlane(baseStore, "control") {
		if err := processControlActionEffect(cfg, baseStore, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), item); err != nil {
			t.Fatalf("processControlActionEffect(context) error = %v", err)
		}
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, baseStore, transition.EffectInvokeRunner)
	store := &failingWorkflowCommandStore{
		Store:           baseStore,
		FailWorkflowID:  workflowItem.workflowID,
		FailCommandKind: transition.CommandWorkflowExecutionCompleted,
		Err:             errors.New("workflow command persistence failed"),
	}

	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect)

	workflow, ok := findWorkflow(baseStore.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != string(transition.WorkflowStateFailed) {
		t.Fatalf("expected workflow to fail after local persistence failure, got %s", workflow.Status)
	}
	if workflow.FailureClass != workflowFailureRunnerPostProcessing {
		t.Fatalf("expected failure class %s, got %s", workflowFailureRunnerPostProcessing, workflow.FailureClass)
	}
	trace, ok := baseStore.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != events.StatusFailed {
		t.Fatalf("expected trace to fail after local persistence failure, got %s", trace.Summary.Status)
	}
	assertWorkflowEffectStatus(t, baseStore, workflowItem.workflowID, transition.EffectInvokeRunner, transition.EffectCompleted)
}

func TestHandleClaimedWorkflowRunnerEffectRunnerCompletionInvariantFailureFailsWorkflow(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"provider": "fake",
			"message":  "Final reply",
			"raw": map[string]any{
				"structured_output": map[string]any{
					"visible_reasoning": []any{
						map[string]any{
							"step_type":    "analysis",
							"summary":      "Collected context.",
							"alternatives": []any{},
							"confidence":   0.9,
							"decision":     "reply_in_thread",
						},
					},
					"reply_draft":        "Draft reply",
					"final_answer":       "Final reply",
					"confidence":         0.9,
					"context_summary":    "Context collected.",
					"self_critique":      "",
					"proposed_actions":   []any{map[string]any{"kind": "slack_post", "rationale": "Reply in thread.", "target_ref": "D123"}},
					"knowledge_drafts":   []any{},
					"outcome_hypotheses": []any{},
				},
			},
		})
	}))
	defer runner.Close()

	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
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
	workflowItem := firstQueuedWorkflowItem(t, baseStore, "slack:")
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

	if err := startWorkflowViaCommand(cfg, baseStore, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	for _, item := range queuedActionEffectsForPlane(baseStore, "control") {
		if err := processControlActionEffect(cfg, baseStore, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), item); err != nil {
			t.Fatalf("processControlActionEffect(context) error = %v", err)
		}
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, baseStore, transition.EffectInvokeRunner)
	store := &noopWorkflowCommandStore{
		Store:           baseStore,
		NoopWorkflowID:  workflowItem.workflowID,
		NoopCommandKind: transition.CommandWorkflowExecutionCompleted,
	}

	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect)

	workflow, ok := findWorkflow(baseStore.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != string(transition.WorkflowStateFailed) {
		t.Fatalf("expected workflow to fail after state invariant violation, got %s", workflow.Status)
	}
	if workflow.FailureClass != workflowFailureRunnerStateInvariant {
		t.Fatalf("expected failure class %s, got %s", workflowFailureRunnerStateInvariant, workflow.FailureClass)
	}
	trace, ok := baseStore.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.Status != events.StatusFailed {
		t.Fatalf("expected trace to fail after state invariant violation, got %s", trace.Summary.Status)
	}
	assertWorkflowEffectStatus(t, baseStore, workflowItem.workflowID, transition.EffectInvokeRunner, transition.EffectCompleted)
}

func TestHandleClaimedWorkflowRunnerEffectNonOKSchedulesSuccessorAttempt(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       false,
			"provider": "fake",
			"message":  "provider unavailable",
			"raw": map[string]any{
				"repair_attempted": false,
				"repair_succeeded": false,
			},
		})
	}))
	defer runner.Close()

	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
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

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}
	cfg := config.Config{
		ServiceName:                     "control-plane",
		DefaultRepo:                     "rsi-agent-platform",
		DefaultKnowledgeBaseURL:         "https://example.test/kb",
		AllowedTargetRepos:              []string{"rsi-agent-platform"},
		RunnerBaseURL:                   runner.URL,
		ToolGatewayBaseURL:              toolGateway.URL,
		SandboxNamespace:                "rsi-platform",
		DefaultReasoningVerbosity:       "verbose",
		WorkflowAutoRetryEnabled:        true,
		WorkflowAutoRetryMaxAttempts:    3,
		WorkflowAutoRetryBackoffSeconds: []int{1, 60},
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	for _, item := range queuedActionEffectsForPlane(store, "control") {
		if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), item); err != nil {
			t.Fatalf("processControlActionEffect(context) error = %v", err)
		}
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect)

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected original workflow to exist")
	}
	if workflow.Status != "failed" {
		t.Fatalf("expected original workflow to fail, got %s", workflow.Status)
	}
	if workflow.FailureClass != workflowFailureRunnerNonOK {
		t.Fatalf("expected failure class %s, got %s", workflowFailureRunnerNonOK, workflow.FailureClass)
	}
	if workflow.RetryDecision != "auto_retry" {
		t.Fatalf("expected retry_decision=auto_retry, got %s", workflow.RetryDecision)
	}

	line, ok := store.GetWorkflowLine(ctx.workflow.CaseID)
	if !ok {
		t.Fatalf("expected workflow line for case %s", ctx.workflow.CaseID)
	}
	if line.Status != string(transition.WorkflowLineStateRetryScheduled) {
		t.Fatalf("expected retry_scheduled workflow line, got %s", line.Status)
	}
	if line.AttemptCount != 2 {
		t.Fatalf("expected attempt_count=2 after scheduling successor, got %d", line.AttemptCount)
	}
	if line.CurrentWorkflowID == "" || line.CurrentWorkflowID == workflowItem.workflowID {
		t.Fatalf("expected successor workflow id on line, got %+v", line)
	}
	successor, ok := findWorkflow(store.ListWorkflows(), line.CurrentWorkflowID)
	if !ok {
		t.Fatalf("expected successor workflow %s", line.CurrentWorkflowID)
	}
	if successor.ParentWorkflowID != workflowItem.workflowID {
		t.Fatalf("expected successor parent workflow %s, got %s", workflowItem.workflowID, successor.ParentWorkflowID)
	}
	if successor.AttemptNumber != 2 {
		t.Fatalf("expected successor attempt_number=2, got %d", successor.AttemptNumber)
	}
	replayTrace, ok := store.GetTrace(successor.TraceID)
	if !ok {
		t.Fatalf("expected successor trace %s", successor.TraceID)
	}
	if replayTrace.Summary.SupersedesTraceID != workflowItem.traceID {
		t.Fatalf("expected successor trace to supersede %s, got %s", workflowItem.traceID, replayTrace.Summary.SupersedesTraceID)
	}

	if err := activateDueWorkflowLineRetries(cfg, store, time.Now().UTC().Add(2*time.Second)); err != nil {
		t.Fatalf("activateDueWorkflowLineRetries() error = %v", err)
	}
	line, _ = store.GetWorkflowLine(ctx.workflow.CaseID)
	if line.Status != string(transition.WorkflowLineStateActive) {
		t.Fatalf("expected active workflow line after due activation, got %s", line.Status)
	}
	successor, _ = findWorkflow(store.ListWorkflows(), line.CurrentWorkflowID)
	if successor.Status == string(transition.WorkflowStateQueued) {
		t.Fatalf("expected successor workflow to start, got %+v", successor)
	}
}

func TestHandleClaimedWorkflowRunnerEffectInvalidRequestUsesRunnerDiagnosticsAndSkipsRetry(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       false,
			"provider": "fake",
			"message":  "OpenAI rejected tools[0].name",
			"raw": map[string]any{
				"failure_class": "runner_invalid_request",
				"runner_diagnostics": map[string]any{
					"failure_kind":           "invalid_request",
					"provider_status_code":   400,
					"provider_error_param":   "tools[0].name",
					"provider_error_code":    "invalid_value",
					"provider_error_message": "Invalid 'tools[0].name': string does not match pattern '^[A-Za-z0-9_-]+$'",
					"invalid_tool_names":     []any{"repo.context", "rsi.workflow_context"},
				},
				"repair_attempted": false,
				"repair_succeeded": false,
			},
		})
	}))
	defer runner.Close()

	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		name = strings.TrimSuffix(name, "/internal/hermes-executions")
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

	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}
	cfg := config.Config{
		ServiceName:                     "control-plane",
		DefaultRepo:                     "rsi-agent-platform",
		DefaultKnowledgeBaseURL:         "https://example.test/kb",
		AllowedTargetRepos:              []string{"rsi-agent-platform"},
		RunnerBaseURL:                   runner.URL,
		ToolGatewayBaseURL:              toolGateway.URL,
		SandboxNamespace:                "rsi-platform",
		DefaultReasoningVerbosity:       "verbose",
		WorkflowAutoRetryEnabled:        true,
		WorkflowAutoRetryMaxAttempts:    3,
		WorkflowAutoRetryBackoffSeconds: []int{1, 60},
	}

	if err := startWorkflowViaCommand(cfg, store, workflowItem.workflowID, time.Now().UTC(), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}
	for _, item := range queuedActionEffectsForPlane(store, "control") {
		if err := processControlActionEffect(cfg, store, clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL), item); err != nil {
			t.Fatalf("processControlActionEffect(context) error = %v", err)
		}
	}

	runnerEffect := firstQueuedWorkflowEffectByKind(t, store, transition.EffectInvokeRunner)
	handleClaimedWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect)

	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if workflow.Status != "failed" {
		t.Fatalf("expected workflow to fail, got %s", workflow.Status)
	}
	if workflow.FailureClass != workflowFailureRunnerInvalidRequest {
		t.Fatalf("expected failure class %s, got %s", workflowFailureRunnerInvalidRequest, workflow.FailureClass)
	}
	if workflow.RetryDecision != "needs_human" {
		t.Fatalf("expected retry_decision=needs_human, got %s", workflow.RetryDecision)
	}
	if workflow.RunnerDiagnostics["provider_error_param"] != "tools[0].name" {
		t.Fatalf("expected persisted runner diagnostics, got %#v", workflow.RunnerDiagnostics)
	}

	line, ok := store.GetWorkflowLine(ctx.workflow.CaseID)
	if !ok {
		t.Fatalf("expected workflow line for case %s", ctx.workflow.CaseID)
	}
	if line.Status != string(transition.WorkflowLineStateNeedsHuman) {
		t.Fatalf("expected needs_human workflow line, got %s", line.Status)
	}
	if line.AttemptCount != 1 {
		t.Fatalf("expected attempt_count=1 without successor retry, got %d", line.AttemptCount)
	}
	if line.CurrentWorkflowID != workflowItem.workflowID {
		t.Fatalf("expected current workflow to remain original attempt, got %s", line.CurrentWorkflowID)
	}
}

func TestBuildRunnerTaskDefersToRunnerDefaultTaskBudget(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}
	cfg := config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}

	task := buildRunnerTask(cfg, store, "prod", ctx.trace, ctx.workflow, ctx.ingestion, "Context collected.", nil)
	if task.TimeoutSeconds != 0 {
		t.Fatalf("expected workflow runner task timeout override to be omitted, got %d", task.TimeoutSeconds)
	}
}

func TestBuildRunnerTaskCarriesKubernetesReadScope(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}
	cfg := config.Config{
		Environment:               "stage",
		DefaultRepo:               "depin-backend",
		AllowedTargetRepos:        []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		KubernetesReadNamespaces:  []string{"story"},
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}

	task := buildRunnerTask(cfg, store, "prod", ctx.trace, ctx.workflow, ctx.ingestion, "Context collected.", nil)

	if got := task.KubernetesReadNamespaces; len(got) != 2 || got[0] != "story" || got[1] != "rsi-platform" {
		t.Fatalf("kubernetes read namespaces = %#v, want story+rsi-platform", got)
	}
	if !strings.Contains(task.Prompt, "Kubernetes read scope: story, rsi-platform") {
		t.Fatalf("expected prompt to advertise Kubernetes read scope, got %q", task.Prompt)
	}
	if !strings.Contains(task.ContextSummary, "Runtime deployment targets: depin-backend, depin-ip-registration") {
		t.Fatalf("expected context summary to advertise depin deployment targets, got %q", task.ContextSummary)
	}
	var readScope map[string]any
	var deploymentTargetRef string
	for _, lease := range task.CapabilityLeases {
		if lease.Capability == "read_context" {
			readScope = lease.Scope
			break
		}
	}
	for _, ref := range task.ContextRefs {
		if ref.Kind == "runtime_deployment_targets" {
			deploymentTargetRef = ref.TargetRef
			break
		}
	}
	namespaces, ok := readScope["kubernetes_read_namespaces"].([]string)
	if !ok || len(namespaces) != 2 || namespaces[0] != "story" || namespaces[1] != "rsi-platform" {
		t.Fatalf("expected read_context lease Kubernetes scope, got %#v", readScope)
	}
	if deploymentTargetRef != "depin-backend,depin-ip-registration" {
		t.Fatalf("expected depin runtime deployment target ref, got %q", deploymentTargetRef)
	}
}

func TestWorkflowRetryAtSkipsRetryAfterReplyPostBegins(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}
	now := time.Now().UTC()
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: "action-reply-begun",
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   "cmd-reply-begun",
		OccurredAt:  now,
		Payload: map[string]any{
			"owner_plane":     "control",
			"conversation_id": ctx.trace.Summary.ConversationID,
			"case_id":         ctx.trace.Summary.CaseID,
			"trace_id":        ctx.trace.Summary.TraceID,
			"kind":            string(action.KindSlackPost),
			"phase_key":       controlPhaseReplyPost,
			"target_ref":      ctx.ingestion.ChannelID,
			"request_payload": map[string]any{
				"channel_id": ctx.ingestion.ChannelID,
				"thread_ts":  ctx.ingestion.ThreadTS,
				"body":       "Final reply",
			},
			"idempotency_key": "reply-begun",
			"approval_mode":   "not_required",
			"approval_state":  "approved",
			"requested_by":    "control-plane",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_queued) error = %v", err)
	}

	retryAt, ok := workflowRetryAt(config.Config{
		WorkflowAutoRetryEnabled:        true,
		WorkflowAutoRetryMaxAttempts:    3,
		WorkflowAutoRetryBackoffSeconds: []int{15, 60},
	}, store, ctx.workflow, workflowFailure{
		Class:     workflowFailureRunnerNonOK,
		Summary:   "provider unavailable",
		Retryable: true,
	})
	if ok {
		t.Fatalf("expected reply-post phase to suppress auto retry, got retryAt=%s", retryAt.Format(time.RFC3339))
	}
}

func TestWorkflowRetryAtUsesSecondBackoffForSecondRetry(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected initial workflow to exist")
	}
	now := time.Now().UTC()
	if _, err := submitWorkflowCommand(store, workflow.ID, transition.CommandWorkflowExecutionFailed, "tester", now, map[string]any{
		"last_error":      "provider unavailable",
		"failure_class":   workflowFailureRunnerNonOK,
		"failure_summary": "provider unavailable",
		"retry_decision":  "auto_retry",
		"retry_after":     now.Add(15 * time.Second),
	}); err != nil {
		t.Fatalf("submitWorkflowCommand(workflow_failed) error = %v", err)
	}

	line, ok := store.GetWorkflowLine(workflow.CaseID)
	if !ok {
		t.Fatalf("expected workflow line for case %s", workflow.CaseID)
	}
	successor, ok := findWorkflow(store.ListWorkflows(), line.CurrentWorkflowID)
	if !ok {
		t.Fatalf("expected successor workflow %s", line.CurrentWorkflowID)
	}

	retryAt, ok := workflowRetryAt(config.Config{
		WorkflowAutoRetryEnabled:        true,
		WorkflowAutoRetryMaxAttempts:    3,
		WorkflowAutoRetryBackoffSeconds: []int{15, 60},
	}, store, successor, workflowFailure{
		Class:     workflowFailureRunnerNonOK,
		Summary:   "provider unavailable",
		Retryable: true,
	})
	if !ok {
		t.Fatal("expected second retry to be allowed")
	}
	delay := time.Until(retryAt)
	if delay < 59*time.Second || delay > 61*time.Second {
		t.Fatalf("expected second retry backoff near 60s, got %s", delay)
	}
}

func TestFinalizeWorkflowFailureWithDetailsBudgetExhaustionMovesLineToNeedsHuman(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected initial workflow to exist")
	}
	now := time.Now().UTC()

	if _, err := submitWorkflowCommand(store, workflow.ID, transition.CommandWorkflowExecutionFailed, "tester", now, map[string]any{
		"last_error":      "provider unavailable",
		"failure_class":   workflowFailureRunnerNonOK,
		"failure_summary": "provider unavailable",
		"retry_decision":  "auto_retry",
		"retry_after":     now.Add(15 * time.Second),
	}); err != nil {
		t.Fatalf("submitWorkflowCommand(first workflow_failed) error = %v", err)
	}
	line, ok := store.GetWorkflowLine(workflow.CaseID)
	if !ok {
		t.Fatalf("expected workflow line for case %s", workflow.CaseID)
	}
	secondAttempt, ok := findWorkflow(store.ListWorkflows(), line.CurrentWorkflowID)
	if !ok {
		t.Fatalf("expected second attempt workflow %s", line.CurrentWorkflowID)
	}

	if _, err := submitWorkflowCommand(store, secondAttempt.ID, transition.CommandWorkflowExecutionFailed, "tester", now.Add(time.Second), map[string]any{
		"last_error":      "provider unavailable again",
		"failure_class":   workflowFailureRunnerNonOK,
		"failure_summary": "provider unavailable again",
		"retry_decision":  "auto_retry",
		"retry_after":     now.Add(61 * time.Second),
	}); err != nil {
		t.Fatalf("submitWorkflowCommand(second workflow_failed) error = %v", err)
	}
	line, _ = store.GetWorkflowLine(workflow.CaseID)
	thirdAttempt, ok := findWorkflow(store.ListWorkflows(), line.CurrentWorkflowID)
	if !ok {
		t.Fatalf("expected third attempt workflow %s", line.CurrentWorkflowID)
	}

	cfg := config.Config{
		ServiceName:                     "control-plane",
		WorkflowAutoRetryEnabled:        true,
		WorkflowAutoRetryMaxAttempts:    3,
		WorkflowAutoRetryBackoffSeconds: []int{15, 60},
	}
	if err := finalizeWorkflowFailureWithDetails(cfg, store, workflowLocator{
		traceID:    thirdAttempt.TraceID,
		workflowID: thirdAttempt.ID,
	}, workflowFailure{
		Class:     workflowFailureRunnerNonOK,
		Summary:   "provider unavailable third time",
		Retryable: true,
	}); err != nil {
		t.Fatalf("finalizeWorkflowFailureWithDetails() error = %v", err)
	}

	line, _ = store.GetWorkflowLine(workflow.CaseID)
	if line.Status != string(transition.WorkflowLineStateNeedsHuman) {
		t.Fatalf("expected budget exhaustion to move line to needs_human, got %s", line.Status)
	}
	if line.AttemptCount != 3 {
		t.Fatalf("expected attempt_count=3 after exhausting budget, got %d", line.AttemptCount)
	}
	if line.AutoRetryBudgetRemaining != 0 {
		t.Fatalf("expected retry budget to be exhausted, got %d", line.AutoRetryBudgetRemaining)
	}
	if line.CurrentWorkflowID != thirdAttempt.ID {
		t.Fatalf("expected exhausted line to stay on third attempt %s, got %s", thirdAttempt.ID, line.CurrentWorkflowID)
	}
	finalAttempt, ok := findWorkflow(store.ListWorkflows(), thirdAttempt.ID)
	if !ok {
		t.Fatalf("expected final attempt workflow %s", thirdAttempt.ID)
	}
	if finalAttempt.RetryDecision != "needs_human" {
		t.Fatalf("expected exhausted attempt retry_decision=needs_human, got %s", finalAttempt.RetryDecision)
	}
}

func TestFinalizeWorkflowFailureWithDetailsNonRetryableMovesLineToNeedsHuman(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatal("expected initial workflow to exist")
	}

	cfg := config.Config{
		ServiceName:                     "control-plane",
		WorkflowAutoRetryEnabled:        true,
		WorkflowAutoRetryMaxAttempts:    3,
		WorkflowAutoRetryBackoffSeconds: []int{15, 60},
	}
	if err := finalizeWorkflowFailureWithDetails(cfg, store, workflowItem, workflowFailure{
		Class:     "policy_block",
		Summary:   "operator review required",
		Retryable: false,
	}); err != nil {
		t.Fatalf("finalizeWorkflowFailureWithDetails() error = %v", err)
	}

	line, ok := store.GetWorkflowLine(workflow.CaseID)
	if !ok {
		t.Fatalf("expected workflow line for case %s", workflow.CaseID)
	}
	if line.Status != string(transition.WorkflowLineStateNeedsHuman) {
		t.Fatalf("expected non-retryable failure to move line to needs_human, got %s", line.Status)
	}
	if line.AttemptCount != 1 {
		t.Fatalf("expected non-retryable failure to avoid successor attempts, got attempt_count=%d", line.AttemptCount)
	}
	finalAttempt, ok := findWorkflow(store.ListWorkflows(), workflow.ID)
	if !ok {
		t.Fatalf("expected workflow %s", workflow.ID)
	}
	if finalAttempt.RetryDecision != "needs_human" {
		t.Fatalf("expected retry_decision=needs_human, got %s", finalAttempt.RetryDecision)
	}
}

func TestToolPlanForRepoProgressQuestionUsesGitHubActivity(t *testing.T) {
	plan := workflowplan.ToolPlan("question", "Hello RSI, can you give me a quick rundown of how depin-backend api progressed in the last week", "depin-backend", "C123", "171000001.000100")
	if !containsString(plan, "github.repo_activity") {
		t.Fatalf("expected github.repo_activity in tool plan, got %#v", plan)
	}
	if !containsString(plan, "slack.history") {
		t.Fatalf("expected slack.history in tool plan, got %#v", plan)
	}
}

func TestToolInputForIntentUsesMentionedRepoAndTimeWindow(t *testing.T) {
	cfg := config.Config{
		DefaultRepo:             "rsi-agent-platform",
		AllowedTargetRepos:      []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL: "https://example.test/kb",
		SandboxNamespace:        "rsi-platform",
	}
	trace := events.TraceSummary{
		TraceID:        "trace-123",
		WorkflowID:     "workflow-123",
		ConversationID: "conv-123",
		CaseID:         "case-123",
	}
	input := toolInputForIntent(cfg, trace, storepkg.Workflow{AssignedBot: "arch", Kind: "architecture"}, slackpkg.Ingestion{
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
	if got := input["trace_id"]; got != trace.TraceID {
		t.Fatalf("expected trace binding, got %#v", got)
	}
	if got := input["workflow_id"]; got != trace.WorkflowID {
		t.Fatalf("expected workflow binding, got %#v", got)
	}
	if got := input["conversation_id"]; got != trace.ConversationID {
		t.Fatalf("expected conversation binding, got %#v", got)
	}
	if got := input["case_id"]; got != trace.CaseID {
		t.Fatalf("expected case binding, got %#v", got)
	}
}

func TestBuildRunnerTaskUsesConfiguredTimeoutAndSlackBinding(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatalf("expected trace %s", workflowItem.traceID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatalf("expected workflow %s", workflowItem.workflowID)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), workflowItem.ingestionID)
	if !ok {
		t.Fatalf("expected ingestion %s", workflowItem.ingestionID)
	}
	task := buildRunnerTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		DefaultReasoningVerbosity: "verbose",
		ProdRunnerTaskTimeout:     300 * time.Second,
	}, store, "prod", trace, workflow, ingestion, "context", nil)

	if task.TimeoutSeconds != 0 {
		t.Fatalf("task timeout = %d, want 0", task.TimeoutSeconds)
	}
	if task.ChannelID != ingestion.ChannelID {
		t.Fatalf("channel id = %q, want %q", task.ChannelID, ingestion.ChannelID)
	}
	if task.ThreadTS != ingestion.ThreadTS {
		t.Fatalf("thread ts = %q, want %q", task.ThreadTS, ingestion.ThreadTS)
	}
	if task.ContractVersion != clients.RunnerExecutionContractVersion {
		t.Fatalf("contract_version = %q, want %q", task.ContractVersion, clients.RunnerExecutionContractVersion)
	}
	if !runnerTaskHasCapability(task, "artifact_write") || !runnerTaskHasCapability(task, "slack_send") || !runnerTaskHasCapability(task, "memory_write") {
		t.Fatalf("expected runner-first workflow capabilities, got %#v", task.CapabilityLeases)
	}
	if task.DeliveryPolicy == nil || !task.DeliveryPolicy.DirectSendAllowed || !task.DeliveryPolicy.UploadAllowed {
		t.Fatalf("expected direct delivery policy, got %#v", task.DeliveryPolicy)
	}
	if task.WorkspacePolicy == nil || task.WorkspacePolicy.ComputerRoot != "/workspace/company" {
		t.Fatalf("expected company workspace policy, got %#v", task.WorkspacePolicy)
	}
}

func TestBuildRunnerTaskAllowsTopLevelDirectMessageDelivery(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatalf("expected trace %s", workflowItem.traceID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatalf("expected workflow %s", workflowItem.workflowID)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), workflowItem.ingestionID)
	if !ok {
		t.Fatalf("expected ingestion %s", workflowItem.ingestionID)
	}
	ingestion.ChannelID = "D123"
	ingestion.ThreadTS = ""

	task := buildRunnerTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		DefaultReasoningVerbosity: "verbose",
	}, store, "prod", trace, workflow, ingestion, "context", nil)

	if task.ThreadTS != "" {
		t.Fatalf("top-level DM task should not invent thread ts, got %q", task.ThreadTS)
	}
	if task.DeliveryPolicy == nil {
		t.Fatal("expected delivery policy")
	}
	if task.DeliveryPolicy.BoundChannelID != "D123" || task.DeliveryPolicy.BoundThreadTS != "" {
		t.Fatalf("expected channel-only DM delivery policy, got %#v", task.DeliveryPolicy)
	}
	if task.DeliveryPolicy.TargetSurface != "direct_message" {
		t.Fatalf("expected direct_message target surface, got %#v", task.DeliveryPolicy)
	}
}

func TestBuildRunnerTaskBoundsNativeMCPToolSurface(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatalf("expected trace %s", workflowItem.traceID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatalf("expected workflow %s", workflowItem.workflowID)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), workflowItem.ingestionID)
	if !ok {
		t.Fatalf("expected ingestion %s", workflowItem.ingestionID)
	}
	task := buildRunnerTask(config.Config{
		Environment:                   "stage",
		DefaultRepo:                   "rsi-agent-platform",
		AllowedTargetRepos:            []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL:       "https://example.test/kb",
		SandboxNamespace:              "rsi-platform",
		DefaultReasoningVerbosity:     "verbose",
		ProdRunnerArtifactTaskTimeout: 30 * time.Minute,
	}, store, "prod", trace, workflow, ingestion, "context", nil)

	if len(task.MCPServers) != 1 {
		t.Fatalf("expected one Slack MCP server, got %#v", task.MCPServers)
	}
	if task.MCPServers[0].Profile != "slack_mcp_reply" {
		t.Fatalf("expected reply-capable Slack MCP profile, got %#v", task.MCPServers)
	}
	liveHints := workflowplan.BuildLiveHints(workflowplan.RuntimeConfig{
		DefaultRepo:      "rsi-agent-platform",
		AllowedRepos:     []string{"depin-backend", "rsi-agent-platform"},
		KnowledgeBaseURL: "https://example.test/kb",
		SandboxNamespace: "rsi-platform",
	}, workflowplan.RequestContext{
		Trace:          trace.Summary,
		WorkflowID:     workflow.ID,
		ConversationID: workflow.ConversationID,
		CaseID:         workflow.CaseID,
		WorkflowKind:   workflow.Kind,
		AssignedBot:    workflow.AssignedBot,
		Question:       ingestion.Text,
		ChannelID:      ingestion.ChannelID,
		ThreadTS:       ingestion.ThreadTS,
		EntityRefs:     append([]slackpkg.EntityRef(nil), ingestion.EntityRefs...),
	}, time.Now().UTC())
	expectedTools := workflowRunnerAllowedTools(liveHints, true, true)
	if len(expectedTools) == 0 {
		t.Fatalf("expected non-empty bounded tool surface from live hints")
	}
	if !reflect.DeepEqual(task.AllowedTools, expectedTools) {
		t.Fatalf("expected allowed tools %#v, got %#v", expectedTools, task.AllowedTools)
	}
	for _, expected := range []string{"cloudflare.inspect", "kubernetes.events", "repo.read_file", "repo.search", "rsi.trace_context", "slack.upload_file"} {
		if !containsString(task.AllowedTools, expected) {
			t.Fatalf("expected %s in bounded tool surface, got %#v", expected, task.AllowedTools)
		}
	}
	for _, forbidden := range []string{"slack.history", "slack.search", "slack.reply"} {
		if containsString(task.AllowedTools, forbidden) {
			t.Fatalf("expected %s to be absent from bounded tool surface, got %#v", forbidden, task.AllowedTools)
		}
	}
}

func TestProcessWorkflowRunnerEffectKeepsFeatureRequestQuestionOnAgenticWorkflowPath(t *testing.T) {
	classifierCalls := 0
	workflowCalls := 0
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Task clients.RunnerTask `json:"task"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode runner payload: %v", err)
		}
		switch payload.Task.TaskType {
		case "general":
			classifierCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":       true,
				"provider": "fake",
				"message":  `{"workflow_kind":"architecture","rationale":"This is a read-heavy progress and alignment question, not a build request.","confidence":0.93}`,
				"raw": map[string]any{
					"structured_output": map[string]any{
						"workflow_kind": "architecture",
						"rationale":     "This is a read-heavy progress and alignment question, not a build request.",
						"confidence":    0.93,
					},
				},
			})
		case "workflow":
			workflowCalls++
			if payload.Task.Intent != "question" {
				t.Fatalf("expected workflow intent question after classification, got %#v", payload.Task)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":       true,
				"provider": "fake",
				"message":  `{"final_answer":"Here is the rundown.","confidence":0.88}`,
				"raw": map[string]any{
					"native_mcp_enabled": true,
					"structured_output": map[string]any{
						"final_answer":    "Here is the rundown.",
						"reply_draft":     "Here is the rundown.",
						"confidence":      0.88,
						"context_summary": "Grounded in repo and Slack evidence.",
					},
					"reply_delivery": map[string]any{
						"channel_id":  payload.Task.ChannelID,
						"thread_ts":   payload.Task.ThreadTS,
						"body":        "Here is the rundown.",
						"send_status": "posted",
					},
				},
			})
		default:
			t.Fatalf("unexpected task type %#v", payload.Task)
		}
	}))
	defer runner.Close()

	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineIngress,
		AggregateID: "slack:171000555.000100",
		CommandKind: string(transition.CommandIngressRecordSlack),
		CommandID:   "cmd-test-feature-request-reroute",
		Actor:       "tester",
		OccurredAt:  now,
		Payload: map[string]any{
			"bot_role":   "orchestrator",
			"team_id":    "T123",
			"channel_id": "D123",
			"thread_ts":  "171000555.000100",
			"user_id":    "U123",
			"text":       "I need a quick rundown of how depin-backend progressed this week and whether it is aligned with numo.",
			"ts":         "171000555.000100",
			"created_at": now,
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(slack ingress) error = %v", err)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), receipt.ResultRef)
	if !ok {
		t.Fatalf("expected ingestion %s", receipt.ResultRef)
	}
	var trace events.Trace
	for _, item := range store.ListTraces() {
		if item.IngestionID != ingestion.ID {
			continue
		}
		loaded, loadedOK := store.GetTrace(item.TraceID)
		if !loadedOK {
			t.Fatalf("expected trace %s", item.TraceID)
		}
		trace = loaded
		break
	}
	if trace.Summary.TraceID == "" {
		t.Fatal("expected trace for ingested feature-request candidate")
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), trace.Summary.WorkflowID)
	if !ok {
		t.Fatalf("expected workflow %s", trace.Summary.WorkflowID)
	}
	if workflow.Kind != "feature-request" {
		t.Fatalf("expected heuristic feature-request workflow, got %#v", workflow)
	}

	cfg := config.Config{
		ServiceName:               "control-plane",
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		AllowedTargetRepos:        []string{"depin-backend", "rsi-agent-platform"},
		RunnerBaseURL:             runner.URL,
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}
	if err := startWorkflowViaCommand(cfg, store, workflow.ID, now.Add(time.Second), queue.WorkflowQueue); err != nil {
		t.Fatalf("startWorkflowViaCommand() error = %v", err)
	}

	var runnerEffect transition.EffectExecution
	for _, effect := range store.ListEffectExecutionsByAggregate(transition.MachineWorkflow, workflow.ID) {
		if effect.EffectKind != transition.EffectInvokeRunner || effect.Status != transition.EffectQueued {
			continue
		}
		claimed, claimedOK, claimErr := store.ClaimEffectExecution(effect.ID, "tester", 30*time.Second)
		if claimErr != nil {
			t.Fatalf("ClaimEffectExecution(%s) error = %v", effect.ID, claimErr)
		}
		if !claimedOK {
			t.Fatalf("expected to claim runner effect %s", effect.ID)
		}
		runnerEffect = claimed
		break
	}
	if runnerEffect.ID == "" {
		t.Fatal("expected queued workflow runner effect")
	}

	if err := processWorkflowRunnerEffect(cfg, store, map[string]*clients.RunnerClient{
		"prod": clients.NewRunnerClient(cfg.RunnerBaseURL),
	}, runnerEffect); err != nil {
		t.Fatalf("processWorkflowRunnerEffect() error = %v", err)
	}
	if classifierCalls != 1 {
		t.Fatalf("expected exactly one classifier call, got %d", classifierCalls)
	}
	if workflowCalls != 1 {
		t.Fatalf("expected exactly one workflow runner call, got %d", workflowCalls)
	}
	assertWorkflowEffectStatus(t, store, workflow.ID, transition.EffectInvokeRunner, transition.EffectCompleted)

	questionRuns := store.ListQuestionRuns()
	if len(questionRuns) != 0 {
		t.Fatalf("expected no question runs after agentic workflow execution, got %#v", questionRuns)
	}
}

func TestBuildRunnerTaskRequestsDiagramArtifact(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatalf("expected trace %s", workflowItem.traceID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatalf("expected workflow %s", workflowItem.workflowID)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), workflowItem.ingestionID)
	if !ok {
		t.Fatalf("expected ingestion %s", workflowItem.ingestionID)
	}
	ingestion.Text = "@RSI can you draw an architecture diagram of depin-backend? Use /architecture-diagram skill"
	task := buildRunnerTask(config.Config{
		Environment:                   "stage",
		DefaultRepo:                   "rsi-agent-platform",
		AllowedTargetRepos:            []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL:       "https://example.test/kb",
		SandboxNamespace:              "rsi-platform",
		DefaultReasoningVerbosity:     "verbose",
		ProdRunnerArtifactTaskTimeout: 30 * time.Minute,
	}, store, "prod", trace, workflow, ingestion, "context", nil)
	if !task.ArtifactOptional {
		t.Fatalf("expected requested artifact to be optional, got %#v", task)
	}
	if len(task.RequestedArtifacts) != 1 || task.RequestedArtifacts[0].Kind != "diagram" {
		t.Fatalf("expected one requested diagram artifact, got %#v", task.RequestedArtifacts)
	}
	if !containsString(task.ExpectedOutputs, "produced_artifacts") {
		t.Fatalf("expected produced_artifacts in expected outputs, got %#v", task.ExpectedOutputs)
	}
	if task.TimeoutSeconds != 1800 {
		t.Fatalf("expected artifact timeout override, got %d", task.TimeoutSeconds)
	}
}

func TestBuildRunnerTaskUsesPromptEnvelopeForArtifactDetection(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatalf("expected trace %s", workflowItem.traceID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatalf("expected workflow %s", workflowItem.workflowID)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), workflowItem.ingestionID)
	if !ok {
		t.Fatalf("expected ingestion %s", workflowItem.ingestionID)
	}
	ingestion.Text = "Summarize the architecture work."
	ingestion.Prompt = slackpkg.SlackPromptEnvelope{
		RawText:      "@RSI can you draw an architecture diagram of depin-backend? Use /architecture-diagram skill",
		RenderedText: "@RSI can you draw an architecture diagram of depin-backend? Use /architecture-diagram skill",
	}

	task := buildRunnerTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}, store, "prod", trace, workflow, ingestion, "context", nil)
	if len(task.RequestedArtifacts) != 1 || task.RequestedArtifacts[0].Kind != "diagram" {
		t.Fatalf("expected one requested diagram artifact from prompt envelope, got %#v", task.RequestedArtifacts)
	}
	if len(task.RequestedSkills) != 1 || task.RequestedSkills[0] != "architecture-diagram" {
		t.Fatalf("expected requested architecture skill from prompt envelope, got %#v", task.RequestedSkills)
	}
}

func TestBuildRunnerTaskUsesBoundThreadContextForArtifactDetection(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatalf("expected trace %s", workflowItem.traceID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatalf("expected workflow %s", workflowItem.workflowID)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), workflowItem.ingestionID)
	if !ok {
		t.Fatalf("expected ingestion %s", workflowItem.ingestionID)
	}
	ingestion.Text = "@RSI can you help with the above"
	contextRefs := []clients.RunnerContextRef{{
		ToolName:  "slack.history",
		Source:    "prefetched_slack_thread",
		ChannelID: ingestion.ChannelID,
		ThreadTS:  ingestion.ThreadTS,
		Summary:   "Bound Slack thread history: Blake: @RSI can you draw an architecture diagram of depin-backend? Use /architecture-diagram skill | Blake: @RSI can you help with the above",
	}}

	task := buildRunnerTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
	}, store, "prod", trace, workflow, ingestion, "context", contextRefs)

	if len(task.RequestedArtifacts) != 1 || task.RequestedArtifacts[0].Kind != "diagram" {
		t.Fatalf("expected bound thread context to request a diagram artifact, got %#v", task.RequestedArtifacts)
	}
	if !containsString(task.RequestedSkills, "architecture-diagram") {
		t.Fatalf("expected bound thread context to request architecture-diagram, got %#v", task.RequestedSkills)
	}
	if !strings.Contains(task.ExecutionIntent["user_request"].(string), "help with the above") {
		t.Fatalf("expected user request to remain the triggering message, got %#v", task.ExecutionIntent)
	}
	if !strings.Contains(task.Prompt, "help with the above") {
		t.Fatalf("expected prompt to include triggering request, got %q", task.Prompt)
	}
}

func TestTraceArtifactsFromProducedArtifactsUsesUniqueStableIDs(t *testing.T) {
	items := []runnerutil.ProducedArtifact{
		{Kind: "diagram", ArtifactRefs: []string{"artifact://diagram-1"}},
		{Kind: "rendered_output", ArtifactRefs: []string{"artifact://render-1"}},
	}

	artifacts := traceArtifactsFromProducedArtifacts("trace-123", items)
	if len(artifacts) != 2 {
		t.Fatalf("expected two trace artifacts, got %#v", artifacts)
	}
	if artifacts[0].ID == artifacts[1].ID {
		t.Fatalf("expected unique artifact IDs, got %#v", artifacts)
	}

	again := traceArtifactsFromProducedArtifacts("trace-123", items)
	if len(again) != 2 {
		t.Fatalf("expected two trace artifacts on second run, got %#v", again)
	}
	if artifacts[0].ID != again[0].ID || artifacts[1].ID != again[1].ID {
		t.Fatalf("expected stable artifact IDs across repeated projections, got first=%#v second=%#v", artifacts, again)
	}
}

func TestTraceArtifactsFromProducedArtifactsKeepsEachRefURL(t *testing.T) {
	items := []runnerutil.ProducedArtifact{
		{
			Kind:         "diagram",
			ArtifactRefs: []string{"hermes-file://workspace/company/diagram.html", "slack://file/F123"},
			FileRef:      "file:///workspace/company/diagram.html",
			SizeBytes:    42,
		},
	}

	artifacts := traceArtifactsFromProducedArtifacts("trace-123", items)
	if len(artifacts) != 3 {
		t.Fatalf("expected one artifact per ref, got %#v", artifacts)
	}
	urlSet := map[string]bool{}
	for _, artifact := range artifacts {
		urlSet[artifact.URL] = true
		if artifact.SizeBytes != 42 {
			t.Fatalf("expected size to project, got %#v", artifacts)
		}
	}
	for _, expected := range []string{"hermes-file://workspace/company/diagram.html", "slack://file/F123", "file:///workspace/company/diagram.html"} {
		if !urlSet[expected] {
			t.Fatalf("missing artifact URL %q in %#v", expected, artifacts)
		}
	}
}

func TestTraceArtifactsFromExecutionLedgerIncludesArtifactManifestPaths(t *testing.T) {
	items := []events.ExecutionLedgerEvent{{
		Kind: "artifact.manifest",
		Payload: map[string]any{
			"kind":          "diagram",
			"rendered_path": "file:///tmp/diagram.png",
			"preview_path":  "slack-file://F123",
			"source_path":   "file:///tmp/diagram.html",
		},
	}}

	artifacts := traceArtifactsFromExecutionLedger("trace-123", items)
	if len(artifacts) != 3 {
		t.Fatalf("expected manifest paths to become trace artifacts, got %#v", artifacts)
	}
	urlSet := map[string]bool{}
	for _, artifact := range artifacts {
		urlSet[artifact.URL] = true
	}
	for _, expected := range []string{"file:///tmp/diagram.png", "slack-file://F123", "file:///tmp/diagram.html"} {
		if !urlSet[expected] {
			t.Fatalf("missing manifest artifact URL %q in %#v", expected, artifacts)
		}
	}
}

func TestToolCallRecordsFromExecutionLedgerProjectsTerminalCallsOnly(t *testing.T) {
	createdAt := time.Date(2026, 4, 25, 21, 8, 0, 0, time.UTC)
	items := []events.ExecutionLedgerEvent{
		{
			ID:     "ledger-start",
			Kind:   "tool.call.started",
			Status: "running",
			Payload: map[string]any{
				"tool_name":    "kubernetes_inspect",
				"tool_call_id": "call-k8s-1",
				"args":         map[string]any{"namespace": "story"},
			},
			RecordedAt: createdAt,
		},
		{
			ID:     "ledger-progress-running",
			Kind:   "tool.call.progress",
			Status: "running",
			Payload: map[string]any{
				"tool_name":      "kubernetes_inspect",
				"progress_event": "tool.started",
			},
			RecordedAt: createdAt,
		},
		{
			ID:     "ledger-progress-completed",
			Kind:   "tool.call.progress",
			Status: "completed",
			Payload: map[string]any{
				"tool_name":      "kubernetes_inspect",
				"progress_event": "tool.completed",
			},
			RecordedAt: createdAt,
		},
		{
			ID:     "ledger-completed",
			Kind:   "tool.call.completed",
			Status: "completed",
			Payload: map[string]any{
				"tool_name":    "kubernetes_inspect",
				"tool_call_id": "call-k8s-1",
				"args":         map[string]any{"namespace": "story"},
				"result": `{
					"status": "failed",
					"summary": "Kubernetes inspection failed: pods is forbidden",
					"approval_state": "not_required",
					"raw_artifact_refs": ["artifact://k8s/failure"]
				}`,
			},
			RecordedAt: createdAt.Add(time.Second),
		},
	}

	records := toolCallRecordsFromExecutionLedger(items, createdAt)
	if len(records) != 1 {
		t.Fatalf("expected one terminal projected tool call, got %#v", records)
	}
	record := records[0]
	if record.ToolName != "kubernetes_inspect" || record.ToolCallID != "call-k8s-1" {
		t.Fatalf("unexpected projected tool identity: %#v", record)
	}
	if record.Status != "failed" {
		t.Fatalf("expected failed status from tool result payload, got %#v", record)
	}
	if record.Summary != "Kubernetes inspection failed: pods is forbidden" {
		t.Fatalf("expected result summary to project, got %#v", record)
	}
	if record.ApprovalState != "not_required" || !reflect.DeepEqual(record.RawArtifactRefs, []string{"artifact://k8s/failure"}) {
		t.Fatalf("expected result metadata to project, got %#v", record)
	}
}

func TestToolCallRecordsFromExecutionEnvelopeProjectsTerminalCallsOnly(t *testing.T) {
	raw := map[string]any{
		"execution_envelope": map[string]any{
			"contract_version": "execution-envelope/v1",
			"execution_id":     "hexec-1",
			"ledger_events": []any{
				map[string]any{
					"event_id": "ledger-start",
					"kind":     "tool.call.started",
					"status":   "running",
					"payload": map[string]any{
						"tool_name":    "repo_search",
						"tool_call_id": "call-repo-1",
					},
				},
				map[string]any{
					"event_id": "ledger-progress",
					"kind":     "tool.call.progress",
					"status":   "completed",
					"payload": map[string]any{
						"tool_name":      "repo_search",
						"progress_event": "tool.completed",
					},
				},
				map[string]any{
					"event_id": "ledger-completed",
					"kind":     "tool.call.completed",
					"status":   "completed",
					"payload": map[string]any{
						"tool_name":    "repo_search",
						"tool_call_id": "call-repo-1",
						"result": map[string]any{
							"status":  "ok",
							"summary": "Found 4 matches.",
						},
					},
				},
			},
		},
	}

	records := toolCallRecordsFromRunnerRaw(raw)
	if len(records) != 1 {
		t.Fatalf("expected one terminal envelope tool call, got %#v", records)
	}
	if records[0].ToolCallID != "call-repo-1" || records[0].Status != "completed" || records[0].Summary != "Found 4 matches." {
		t.Fatalf("unexpected envelope tool projection: %#v", records[0])
	}
}

func TestWorkflowReplyDeliveryRecordsFailedExecutionEnvelopeDeliveryAttempt(t *testing.T) {
	record, ok := workflowReplyDelivery(map[string]any{
		"reply_delivery": map[string]any{
			"body":       "legacy top-level reply",
			"channel_id": "C-legacy",
			"thread_ts":  "T-legacy",
		},
		"structured_output": map[string]any{
			"reply_delivery": map[string]any{
				"body":       "legacy structured reply",
				"channel_id": "C-structured",
				"thread_ts":  "T-structured",
			},
		},
		"execution_envelope": map[string]any{
			"deliveries": []any{
				map[string]any{
					"body":         "envelope reply",
					"channel_id":   "C-envelope",
					"thread_ts":    "T-envelope",
					"tool_call_id": "tool-call-envelope",
					"send_status":  "failed",
				},
			},
		},
	}, "C-fallback", "T-fallback")
	if !ok || record.SendStatus != "failed" {
		t.Fatalf("expected failed envelope delivery attempt to be recorded, got ok=%t record=%#v", ok, record)
	}
}

func TestWorkflowReplyDeliveryFromExecutionLedgerRecordsFailedStatus(t *testing.T) {
	createdAt := time.Now().UTC()
	record, ok := workflowReplyDeliveryFromExecutionLedger([]events.ExecutionLedgerEvent{
		{
			ID:          "ledger-failed",
			ExecutionID: "hexec-1",
			Kind:        "slack.message.sent",
			Status:      "failed",
			Seq:         1,
			Payload: map[string]any{
				"body":        "failed reply",
				"channel_id":  "C123",
				"thread_ts":   "171000001.000100",
				"send_status": "failed",
			},
			RecordedAt: createdAt,
		},
	}, "C-fallback", "T-fallback", createdAt)
	if !ok || record.SendStatus != "failed" {
		t.Fatalf("expected failed ledger delivery attempt to be recorded, got ok=%t record=%#v", ok, record)
	}
	record, ok = workflowReplyDeliveryFromExecutionLedger([]events.ExecutionLedgerEvent{
		{
			ID:             "ledger-sent",
			ExecutionID:    "hexec-1",
			Kind:           "slack.message.sent",
			Status:         "posted",
			Seq:            2,
			IdempotencyKey: "idem-1",
			Payload: map[string]any{
				"body":        "posted reply",
				"channel_id":  "C123",
				"thread_ts":   "171000001.000100",
				"send_status": "posted",
			},
			RecordedAt: createdAt,
		},
	}, "C-fallback", "T-fallback", createdAt)
	if !ok {
		t.Fatal("expected successful ledger delivery")
	}
	if record.FinalBody != "posted reply" || record.IdempotencyKey != "idem-1" {
		t.Fatalf("unexpected ledger delivery projection: %#v", record)
	}
}

func TestWorkflowReplyDeliveryProjectionPrefersLedgerDeliveryAttempt(t *testing.T) {
	createdAt := time.Now().UTC()
	raw := map[string]any{
		"reply_delivery": map[string]any{
			"body":        "raw posted reply",
			"channel_id":  "C-raw",
			"thread_ts":   "T-raw",
			"send_status": "posted",
		},
	}
	record, ok := workflowReplyDeliveryProjection(raw, []events.ExecutionLedgerEvent{
		{
			ID:          "ledger-observed",
			ExecutionID: "hexec-1",
			Kind:        "slack.message.sent",
			Status:      "observed",
			Seq:         1,
			Payload: map[string]any{
				"body":        "ledger reply",
				"channel_id":  "C-ledger",
				"thread_ts":   "T-ledger",
				"send_status": "observed",
			},
			RecordedAt: createdAt,
		},
	}, true, "C-fallback", "T-fallback", createdAt)
	if !ok {
		t.Fatal("expected ledger delivery attempt projection")
	}
	if record.FinalBody != "ledger reply" || record.ChannelID != "C-ledger" || record.SendStatus != "observed" {
		t.Fatalf("unexpected ledger delivery attempt: %#v", record)
	}
}

func TestWorkflowReplyDeliveryProjectionPrefersSuccessfulLedgerDelivery(t *testing.T) {
	createdAt := time.Now().UTC()
	raw := map[string]any{
		"reply_delivery": map[string]any{
			"body":        "raw posted reply",
			"channel_id":  "C-raw",
			"thread_ts":   "T-raw",
			"send_status": "posted",
		},
	}
	record, ok := workflowReplyDeliveryProjection(raw, []events.ExecutionLedgerEvent{
		{
			ID:          "ledger-sent",
			ExecutionID: "hexec-1",
			Kind:        "slack.message.sent",
			Status:      "posted",
			Seq:         1,
			Payload: map[string]any{
				"body":        "ledger posted reply",
				"channel_id":  "C-ledger",
				"thread_ts":   "T-ledger",
				"send_status": "posted",
			},
			RecordedAt: createdAt,
		},
		{
			ID:          "ledger-failed-retry",
			ExecutionID: "hexec-1",
			Kind:        "slack.message.sent",
			Status:      "failed",
			Seq:         2,
			Payload: map[string]any{
				"body":        "failed retry",
				"channel_id":  "C-ledger",
				"thread_ts":   "T-ledger",
				"send_status": "failed",
			},
			RecordedAt: createdAt.Add(time.Second),
		},
	}, true, "C-fallback", "T-fallback", createdAt)
	if !ok {
		t.Fatal("expected successful ledger delivery")
	}
	if record.FinalBody != "ledger posted reply" || record.ChannelID != "C-ledger" {
		t.Fatalf("expected ledger delivery to win, got %#v", record)
	}
}

func TestBuildRunnerTaskIncludesNotionMCPWhenEnabled(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	trace, ok := store.GetTrace(workflowItem.traceID)
	if !ok {
		t.Fatalf("expected trace %s", workflowItem.traceID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowItem.workflowID)
	if !ok {
		t.Fatalf("expected workflow %s", workflowItem.workflowID)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), workflowItem.ingestionID)
	if !ok {
		t.Fatalf("expected ingestion %s", workflowItem.ingestionID)
	}
	task := buildRunnerTask(config.Config{
		Environment:                  "stage",
		DefaultRepo:                  "rsi-agent-platform",
		AllowedTargetRepos:           []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL:      "https://example.test/kb",
		SandboxNamespace:             "rsi-platform",
		DefaultReasoningVerbosity:    "verbose",
		NotionMCPEnabled:             true,
		NotionMCPServerURL:           "https://mcp.notion.com/mcp",
		NotionMCPHeaderEnvVars:       map[string]string{"CF-Access-Client-Id": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID"},
		NotionMCPAuthorizationEnvVar: "RSI_NOTION_MCP_AUTHORIZATION",
	}, store, "prod", trace, workflow, ingestion, "context", nil)

	if len(task.MCPServers) != 2 {
		t.Fatalf("expected Slack and Notion MCP servers, got %#v", task.MCPServers)
	}
	if task.MCPServers[0].Profile != "slack_mcp_reply" {
		t.Fatalf("expected first MCP server to remain Slack reply, got %#v", task.MCPServers)
	}
	if task.MCPServers[1].ServerLabel != "notion" {
		t.Fatalf("expected notion MCP server, got %#v", task.MCPServers[1])
	}
	if task.MCPServers[1].AuthorizationEnvVar != "RSI_NOTION_MCP_AUTHORIZATION" {
		t.Fatalf("unexpected notion auth env var %#v", task.MCPServers[1])
	}
	if !reflect.DeepEqual(task.MCPServers[1].HeaderEnvVars, map[string]string{"CF-Access-Client-Id": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID"}) {
		t.Fatalf("unexpected notion header env vars %#v", task.MCPServers[1].HeaderEnvVars)
	}
	if !reflect.DeepEqual(task.MCPServers[1].AllowedTools, map[string]any{"read_only": true}) {
		t.Fatalf("expected read-only notion tool surface, got %#v", task.MCPServers[1].AllowedTools)
	}
	if !strings.Contains(task.SystemMessage, "Use Notion MCP for Notion workspace search and page fetches when relevant.") {
		t.Fatalf("expected notion MCP instruction in system prompt, got %q", task.SystemMessage)
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

func firstQueuedWorkflowItem(t *testing.T, store storepkg.Store, threadPrefix string) workflowLocator {
	t.Helper()
	for _, trace := range store.ListTraces() {
		if trace.Status != events.StatusQueued || !strings.HasPrefix(trace.ThreadKey, threadPrefix) {
			continue
		}
		if _, ok := findWorkflow(store.ListWorkflows(), trace.WorkflowID); !ok {
			continue
		}
		return workflowLocator{
			traceID:     trace.TraceID,
			workflowID:  trace.WorkflowID,
			ingestionID: trace.IngestionID,
		}
	}
	t.Fatalf("expected queued workflow item with thread prefix %s", threadPrefix)
	return workflowLocator{}
}

func queuedActionEffectsForPlane(store storepkg.Store, ownerPlane string) []transition.EffectExecution {
	out := make([]transition.EffectExecution, 0)
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAction || effect.EffectKind != transition.EffectInvokeAction || effect.Status != transition.EffectQueued {
			continue
		}
		if stringFromMap(effect.Payload, "owner_plane") != ownerPlane {
			continue
		}
		claimed, ok, err := store.ClaimEffectExecution(effect.ID, "tester", 30*time.Second)
		if err == nil && ok {
			out = append(out, claimed)
			continue
		}
		out = append(out, effect)
	}
	return out
}

func firstQueuedActionEffectByKind(t *testing.T, store storepkg.Store, ownerPlane string, kind action.Kind) transition.EffectExecution {
	t.Helper()
	for _, effect := range queuedActionEffectsForPlane(store, ownerPlane) {
		if stringFromMap(effect.Payload, "kind") == string(kind) {
			return effect
		}
	}
	t.Fatalf("expected queued action effect owner_plane=%s kind=%s", ownerPlane, kind)
	return transition.EffectExecution{}
}

func firstQueuedWorkflowEffectByKind(t *testing.T, store storepkg.Store, kind transition.EffectKind) transition.EffectExecution {
	t.Helper()
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind == transition.MachineWorkflow && effect.EffectKind == kind && effect.Status == transition.EffectQueued {
			claimed, ok, err := store.ClaimEffectExecution(effect.ID, "tester", 30*time.Second)
			if err != nil {
				t.Fatalf("ClaimEffectExecution(%s) error = %v", effect.ID, err)
			}
			if ok {
				return claimed
			}
			return effect
		}
	}
	t.Fatalf("expected queued workflow effect kind=%s", kind)
	return transition.EffectExecution{}
}

func assertWorkflowEffectStatus(t *testing.T, store storepkg.Store, workflowID string, kind transition.EffectKind, want transition.EffectStatus) {
	t.Helper()
	for _, effect := range store.ListEffectExecutionsByAggregate(transition.MachineWorkflow, workflowID) {
		if effect.EffectKind != kind {
			continue
		}
		if effect.Status != want {
			t.Fatalf("expected workflow effect %s to be %s, got %+v", kind, want, effect)
		}
		return
	}
	t.Fatalf("expected workflow effect %s for workflow %s", kind, workflowID)
}

type failingActionCommandStore struct {
	storepkg.Store
	FailActionID string
	Err          error
}

type failingWorkflowCommandStore struct {
	storepkg.Store
	FailWorkflowID  string
	FailCommandKind transition.WorkflowCommandKind
	Err             error
}

type noopWorkflowCommandStore struct {
	storepkg.Store
	NoopWorkflowID  string
	NoopCommandKind transition.WorkflowCommandKind
}

type effectSelectionStore struct {
	storepkg.Store
	effects    []transition.EffectExecution
	claimed    []string
	completed  []string
	failed     []string
	resultRefs map[string]string
}

type claimQueueCaptureStore struct {
	storepkg.Store
	queueNames []string
}

func (s *claimQueueCaptureStore) ClaimNextEffectExecutionForKinds(holder string, lease time.Duration, queueNames []string, maxPerScope int, selectors []storepkg.EffectClaimSelector) (transition.EffectExecution, bool, error) {
	s.queueNames = append([]string{}, queueNames...)
	return transition.EffectExecution{}, false, nil
}

func (s *effectSelectionStore) ListEffectExecutionsByAggregate(machineKind transition.MachineKind, aggregateID string) []transition.EffectExecution {
	if machineKind != transition.MachineWorkflow {
		return s.Store.ListEffectExecutionsByAggregate(machineKind, aggregateID)
	}
	out := make([]transition.EffectExecution, 0, len(s.effects))
	for _, effect := range s.effects {
		if effect.MachineKind == machineKind && effect.AggregateID == aggregateID {
			out = append(out, effect)
		}
	}
	return out
}

func (s *effectSelectionStore) ClaimEffectExecution(effectID string, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	for idx := range s.effects {
		if s.effects[idx].ID != effectID {
			continue
		}
		s.claimed = append(s.claimed, effectID)
		s.effects[idx].Status = transition.EffectRunning
		return s.effects[idx], true, nil
	}
	return s.Store.ClaimEffectExecution(effectID, holder, lease)
}

func (s *effectSelectionStore) CompleteEffectExecution(effectID string, holder string, resultRef string) (transition.EffectExecution, error) {
	for idx := range s.effects {
		if s.effects[idx].ID != effectID {
			continue
		}
		s.completed = append(s.completed, effectID)
		if s.resultRefs == nil {
			s.resultRefs = map[string]string{}
		}
		s.resultRefs[effectID] = resultRef
		s.effects[idx].Status = transition.EffectCompleted
		s.effects[idx].ResultRef = resultRef
		return s.effects[idx], nil
	}
	return s.Store.CompleteEffectExecution(effectID, holder, resultRef)
}

func (s *effectSelectionStore) FailEffectExecution(effectID string, holder string, lastError string) (transition.EffectExecution, error) {
	for idx := range s.effects {
		if s.effects[idx].ID != effectID {
			continue
		}
		s.failed = append(s.failed, effectID)
		s.effects[idx].Status = transition.EffectFailed
		s.effects[idx].LastError = lastError
		return s.effects[idx], nil
	}
	return s.Store.FailEffectExecution(effectID, holder, lastError)
}

func (s *failingActionCommandStore) SubmitCommand(command transition.CommandEnvelope) (transition.CommandReceipt, error) {
	if command.MachineKind == transition.MachineAction && command.AggregateID == s.FailActionID {
		switch transition.ActionExecutionCommandKind(command.CommandKind) {
		case transition.CommandActionSucceed, transition.CommandActionBlock, transition.CommandActionFail:
			if boolValue, ok := command.Payload["record_result"].(bool); !ok || boolValue {
				return transition.CommandReceipt{}, s.Err
			}
		}
	}
	return s.Store.SubmitCommand(command)
}

func (s *failingWorkflowCommandStore) SubmitCommand(command transition.CommandEnvelope) (transition.CommandReceipt, error) {
	if command.MachineKind == transition.MachineWorkflow &&
		command.AggregateID == s.FailWorkflowID &&
		transition.WorkflowCommandKind(command.CommandKind) == s.FailCommandKind {
		return transition.CommandReceipt{}, s.Err
	}
	return s.Store.SubmitCommand(command)
}

func (s *noopWorkflowCommandStore) SubmitCommand(command transition.CommandEnvelope) (transition.CommandReceipt, error) {
	if command.MachineKind == transition.MachineWorkflow &&
		command.AggregateID == s.NoopWorkflowID &&
		transition.WorkflowCommandKind(command.CommandKind) == s.NoopCommandKind {
		now := command.OccurredAt
		if now.IsZero() {
			now = time.Now().UTC()
		}
		return transition.CommandReceipt{
			CommandID:        command.CommandID,
			MachineKind:      command.MachineKind,
			AggregateID:      command.AggregateID,
			CommandKind:      command.CommandKind,
			CausationID:      command.CausationID,
			Actor:            command.Actor,
			DecisionKind:     transition.DecisionAdvance,
			Reason:           "noop command for invariant test",
			AggregateVersion: 0,
			CreatedAt:        now,
			UpdatedAt:        now,
		}, nil
	}
	return s.Store.SubmitCommand(command)
}

func runnerTaskHasCapability(task clients.RunnerTask, capability string) bool {
	for _, lease := range task.CapabilityLeases {
		if lease.Capability == capability && lease.Granted {
			return true
		}
	}
	return false
}
