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
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
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

	toolCalls := 0
	slackPosts := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/tools/")
		name = strings.TrimSuffix(name, "/execute")
		switch name {
		case "repo.context", "knowledge.context", "sentry.lookup", "kubernetes.inspect", "github.repo_activity", "slack.history", "rsi.workflow_context", "rsi.action_chain", "rsi.runtime_health", "rsi.runtime_deployment_facts":
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
	if toolCalls != 0 {
		t.Fatalf("expected no control-plane tool prefetch calls, got %d", toolCalls)
	}

	if len(store.ListEvalRuns()) == 0 {
		t.Fatal("expected workflow completion to trigger immediate problem-line evaluation")
	}
	assertWorkflowEffectStatus(t, store, workflowItem.workflowID, transition.EffectInvokeRunner, transition.EffectCompleted)
	assertWorkflowEffectStatus(t, store, workflowItem.workflowID, transition.EffectPostSlackReply, transition.EffectCompleted)
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
			"message":  `{"visible_reasoning":[],"reply_draft":"Final reply from Slack MCP.","final_answer":"Final reply from Slack MCP.","confidence":0.88,"context_summary":"Grounded in Slack MCP evidence.","self_critique":"","proposed_actions":[],"knowledge_drafts":[],"outcome_hypotheses":[]}`,
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
					"visible_reasoning":  []any{},
					"reply_draft":        "Final reply from Slack MCP.",
					"final_answer":       "Final reply from Slack MCP.",
					"confidence":         0.88,
					"context_summary":    "Grounded in Slack MCP evidence.",
					"self_critique":      "",
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

	toolGatewayCalls := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toolGatewayCalls++
		t.Fatalf("unexpected tool gateway invocation %s", r.URL.Path)
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
	if toolGatewayCalls != 0 {
		t.Fatalf("expected no tool gateway calls, got %d", toolGatewayCalls)
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

	toolGatewayCalls := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toolGatewayCalls++
		t.Fatalf("unexpected tool gateway invocation %s", r.URL.Path)
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
	if toolGatewayCalls != 0 {
		t.Fatalf("expected no tool gateway calls, got %d", toolGatewayCalls)
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

	toolGatewayCalls := 0
	toolGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toolGatewayCalls++
		t.Fatalf("unexpected tool gateway invocation %s", r.URL.Path)
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

	if toolGatewayCalls != 0 {
		t.Fatalf("expected no tool gateway calls, got %d", toolGatewayCalls)
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
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"depin-backend", "rsi-agent-platform"},
		DefaultKnowledgeBaseURL:   "https://example.test/kb",
		SandboxNamespace:          "rsi-platform",
		DefaultReasoningVerbosity: "verbose",
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
	expectedTools := workflowRunnerAllowedTools(liveHints, true)
	if len(expectedTools) == 0 {
		t.Fatalf("expected non-empty bounded tool surface from live hints")
	}
	if !reflect.DeepEqual(task.AllowedTools, expectedTools) {
		t.Fatalf("expected allowed tools %#v, got %#v", expectedTools, task.AllowedTools)
	}
	for _, forbidden := range []string{"slack.history", "slack.search", "slack.reply", "cloudflare.inspect"} {
		if containsString(task.AllowedTools, forbidden) {
			t.Fatalf("expected %s to be absent from bounded tool surface, got %#v", forbidden, task.AllowedTools)
		}
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
