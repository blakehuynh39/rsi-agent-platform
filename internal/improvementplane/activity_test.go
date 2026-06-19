package improvementplane

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestTraceActivityProjectorCollapsesTodoCompletions(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("todo-1", 1, "tool.call.completed", "completed", now, map[string]any{
			"tool_name": "todo",
			"result":    `{"todos":[{"id":"1","content":"Read Sentry issue","status":"completed"},{"id":"2","content":"Check Loki logs","status":"in_progress"},{"id":"3","content":"Post answer","status":"pending"}]}`,
		}),
	})
	var todo *TraceActivityItem
	for i := range items {
		if items[i].Kind == "todo" {
			todo = &items[i]
		}
	}
	if todo == nil {
		t.Fatal("missing todo activity item")
	}
	if todo.Status != "running" {
		t.Fatalf("todo status=%q, want running", todo.Status)
	}
	if got := todo.Summary; got != "1/3 complete · Check Loki logs" {
		t.Fatalf("todo summary=%q", got)
	}
	details, ok := todo.Details["todos"].([]map[string]any)
	if !ok || len(details) != 3 {
		t.Fatalf("todo details=%#v, want 3 todos", todo.Details["todos"])
	}
}

func TestTraceActivityProjectorDedupeNativeLifecycleWithToolCompletion(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "detailed", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("native-1", 1, "native_tool.completed", "completed", now, map[string]any{
			"tool_name":           "rsi_slack.report_post",
			"transport_tool_name": "rsi_slack_report_post",
			"ok":                  true,
			"action": map[string]any{
				"id":               "extact-1",
				"response_summary": "Posted Slack report.",
			},
		}),
		ledgerEvent("tool-1", 2, "tool.call.completed", "completed", now.Add(500*time.Millisecond), map[string]any{
			"tool_name":           "rsi_slack_report_post",
			"tool_call_id":        "call-1",
			"transport_tool_name": "rsi_slack_report_post",
			"result":              `{"status":"ok","summary":"Posted Slack report.","output":{"action":{"id":"extact-1"}}}`,
		}),
	})
	count := 0
	for _, item := range items {
		if item.ToolName == "rsi_slack.report_post" {
			count++
			if item.ToolCallID != "call-1" {
				t.Fatalf("tool_call_id=%q, want call-1", item.ToolCallID)
			}
			if got := item.Details["native_action_id"]; got != "extact-1" {
				t.Fatalf("native_action_id=%#v, want extact-1", got)
			}
		}
	}
	if count != 1 {
		t.Fatalf("slack report activity count=%d, want 1", count)
	}
}

func TestTraceActivityProjectorProjectsKnowledgeTargets(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("wiki-1", 1, "tool.call.completed", "failed", now, map[string]any{
			"tool_name": "wiki_page_get",
			"args": map[string]any{
				"page_ref": "architecture/project-data-audit",
				"slug":     "project-data-audit",
			},
			"result": `Error: wiki_page_get failed: company wiki API returned HTTP 404`,
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want 1", len(items))
	}
	item := items[0]
	if item.ToolName != "rsi_knowledge.wiki_page_get" {
		t.Fatalf("tool_name=%q, want canonical wiki page tool", item.ToolName)
	}
	if got := item.Details["page_ref"]; got != "architecture/project-data-audit" {
		t.Fatalf("page_ref=%#v, want requested wiki page", got)
	}
	if got := item.Details["slug"]; got != "project-data-audit" {
		t.Fatalf("slug=%#v, want requested wiki slug", got)
	}
	if !strings.Contains(item.Summary, "architecture/project-data-audit") {
		t.Fatalf("summary=%q, want wiki page target", item.Summary)
	}
}

func TestTraceActivityProjectorProjectsSlackThreadTargets(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("slack-1", 1, "tool.call.completed", "completed", now, map[string]any{
			"tool_name": "rsi_slack_conversation_read",
			"args": map[string]any{
				"channel_id": "CENG",
				"thread_ts":  "171000001.000100",
				"limit":      25,
			},
			"result": `{"status":"ok","messages":[{"text":"hello"}]}`,
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want 1", len(items))
	}
	item := items[0]
	if item.ToolName != "rsi_slack.conversation_read" {
		t.Fatalf("tool_name=%q, want canonical Slack conversation read", item.ToolName)
	}
	if got := item.Details["channel_id"]; got != "CENG" {
		t.Fatalf("channel_id=%#v, want Slack channel", got)
	}
	if got := item.Details["thread_ts"]; got != "171000001.000100" {
		t.Fatalf("thread_ts=%#v, want Slack thread", got)
	}
	if !strings.Contains(item.Summary, "CENG") || !strings.Contains(item.Summary, "171000001.000100") {
		t.Fatalf("summary=%q, want Slack thread target", item.Summary)
	}
}

func TestTraceActivityProjectorProjectsKanbanProjectTools(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("kanban-project", 1, "tool.call.completed", "completed", now, map[string]any{
			"tool_name":    "rsi_kanban_create_project",
			"tool_call_id": "call-kanban-project",
			"args": map[string]any{
				"name":       "Numo",
				"slug":       "numo",
				"channel_id": "C123",
				"thread_ts":  "171000001.000100",
			},
			"result": `{
				"status":"ok",
				"summary":"created Kanban project numo",
				"output":{
					"project":{"id":"kproj_1","slug":"numo","name":"Numo"},
					"route":{"project_id":"kproj_1","channel_id":"C123","thread_ts":"171000001.000100"}
				}
			}`,
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want 1", len(items))
	}
	item := items[0]
	if item.ToolName != "rsi_kanban.create_project" {
		t.Fatalf("tool_name=%q, want canonical Kanban create project", item.ToolName)
	}
	if got := item.Details["project_slug"]; got != "numo" {
		t.Fatalf("project_slug=%#v, want numo", got)
	}
	if got := item.Details["channel_id"]; got != "C123" {
		t.Fatalf("channel_id=%#v, want C123", got)
	}
	if !strings.Contains(item.Summary, "created Kanban project numo") {
		t.Fatalf("summary=%q, want project creation summary", item.Summary)
	}
}

func TestTraceActivityProjectorHumanizesSlackUserIDQuery(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("knowledge-user", 1, "tool.call.completed", "completed", now, map[string]any{
			"tool_name": "rsi_knowledge_search",
			"args": map[string]any{
				"query": "U0772SH7BRA",
			},
			"result": `{"status":"ok","summary":"U0772SH7BRA","results":[]}`,
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want 1", len(items))
	}
	if got := items[0].Details["user_id"]; got != "U0772SH7BRA" {
		t.Fatalf("user_id=%#v, want Slack user ID detail", got)
	}
	if got := items[0].Summary; got != "Slack user ID U0772SH7BRA" {
		t.Fatalf("summary=%q, want humanized Slack ID", got)
	}
}

func TestTraceActivityProjectorProjectsNotionPageMetadata(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("notion-page", 1, "tool.call.completed", "completed", now, map[string]any{
			"tool_name": "rsi_notion_page_get",
			"args": map[string]any{
				"page_id": "page-abc",
			},
			"result": `{
				"status":"ok",
				"summary":"loaded Notion page",
				"output":{
					"object":"page",
					"id":"page-abc",
					"url":"https://notion.so/page-abc",
					"properties":{
						"Name":{"title":[{"plain_text":"Project Data Audit"}]},
						"Description":{"rich_text":[{"plain_text":"Architecture context for the audit portal."}]}
					}
				}
			}`,
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want 1", len(items))
	}
	item := items[0]
	if got := item.Details["title"]; got != "Project Data Audit" {
		t.Fatalf("title=%#v, want Notion title", got)
	}
	if got := item.Details["description"]; got != "Architecture context for the audit portal." {
		t.Fatalf("description=%#v, want Notion description", got)
	}
	if got := item.Details["url"]; got != "https://notion.so/page-abc" {
		t.Fatalf("url=%#v, want Notion URL", got)
	}
	if !strings.Contains(item.Summary, "Project Data Audit") {
		t.Fatalf("summary=%q, want Notion title", item.Summary)
	}
}

func TestTraceActivityNotionTitleIgnoresDescriptionOnlyProperties(t *testing.T) {
	payload := map[string]any{
		"output": map[string]any{
			"object": "page",
			"properties": map[string]any{
				"Description": map[string]any{
					"rich_text": []any{
						map[string]any{"plain_text": "Architecture context for the audit portal."},
					},
				},
			},
		},
	}
	if got := traceActivityNotionTitleFromValue(payload); got != "" {
		t.Fatalf("title=%#v, want empty title for description-only properties", got)
	}
}

func TestTraceActivityStringFromResultSkipsStructuredValues(t *testing.T) {
	result := map[string]any{
		"output": map[string]any{
			"page": map[string]any{
				"title": "Deployment Guide",
			},
			"count": 2,
		},
	}
	if got := traceActivityStringFromResult(result, "page"); got != "" {
		t.Fatalf("structured page value=%q, want empty string", got)
	}
	if got := traceActivityStringFromResult(result, "count"); got != "2" {
		t.Fatalf("numeric scalar value=%q, want 2", got)
	}
}

func TestTraceActivityProjectorGroupsDBReadByRequestAcrossExecutions(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	first := ledgerEvent("db-1", 1, "tool.call.completed", "completed", now, map[string]any{
		"tool_name":    "db_read_query",
		"tool_call_id": "call-db-1",
		"result":       `{"status":"pending","request_id":"dbread-1","target":"depin-prod","sql_sha256":"sha256:abc"}`,
	})
	first.ExecutionID = "hexec-a"
	second := ledgerEvent("db-2", 1, "tool.call.completed", "completed", now.Add(30*time.Minute), map[string]any{
		"tool_name":    "db_read_query",
		"tool_call_id": "call-db-2",
		"result":       `{"status":"ok","request_id":"dbread-1","target":"depin-prod","sql_sha256":"sha256:abc","rows":12,"truncated":false}`,
	})
	second.ExecutionID = "hexec-b"
	items, _ := projector.Project([]events.ExecutionLedgerEvent{first, second})
	count := 0
	for _, item := range items {
		if item.ToolName == "db_read.query" {
			count++
			if got := item.Details["request_id"]; got != "dbread-1" {
				t.Fatalf("request_id=%#v, want dbread-1", got)
			}
		}
	}
	if count != 1 {
		t.Fatalf("db read activity count=%d, want 1", count)
	}
}

func TestTraceActivityProjectorInfersPlainTextToolErrors(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("tool-error", 1, "tool.call.completed", "completed", now, map[string]any{
			"tool_name": "terminal",
			"result":    "Error: command timed out",
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want 1", len(items))
	}
	if items[0].Status != "failed" {
		t.Fatalf("status=%q, want failed", items[0].Status)
	}
	if got := items[0].Details["error_excerpt"]; got != "Error: command timed out" {
		t.Fatalf("error_excerpt=%#v, want plain error", got)
	}
}

func TestTraceActivityProjectorProjectsSlackMessageSentAsFinalResponse(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "detailed", now: now}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("delivery-1", 1, "slack.message.sent", "posted", now, map[string]any{
			"delivery_id":  "delivery-abc",
			"send_status":  "posted",
			"channel_id":   "C123",
			"thread_ts":    "171000001.000100",
			"body":         "Final report posted.",
			"provider_ref": "slack:C123:171000001.000200",
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want only final response row: %#v", len(items), items)
	}
	var final *TraceActivityItem
	for i := range items {
		if items[i].Kind == "final_response" {
			final = &items[i]
		}
	}
	if final == nil {
		t.Fatal("missing final response activity item")
	}
	if final.Status != "completed" {
		t.Fatalf("final status=%q, want completed", final.Status)
	}
	if final.Summary != "Final report posted." {
		t.Fatalf("final summary=%q, want final body", final.Summary)
	}
	if got := final.Details["provider_ref"]; got != "slack:C123:171000001.000200" {
		t.Fatalf("provider_ref=%#v, want Slack provider ref", got)
	}
}

func TestTraceActivitySnapshotEncodesEmptyItemsAsArray(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	snapshot, ok := buildTraceActivitySnapshot(store, traceID, "main", "clean", 250, "", time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("buildTraceActivitySnapshot returned !ok")
	}
	if snapshot.Items == nil {
		t.Fatal("snapshot items is nil, want empty slice")
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	if !strings.Contains(string(raw), `"items":[]`) {
		t.Fatalf("snapshot JSON encoded items as non-array: %s", raw)
	}
}

func TestTraceActivityProjectorFiltersScopeBeforeProjection(t *testing.T) {
	now := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: now}
	items, scopedCount := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("main-1", 1, "tool.call.completed", "completed", now, map[string]any{
			"tool_name": "terminal",
			"result":    `{"status":"ok","exit_code":0}`,
		}),
		ledgerEvent("eval-1", 2, "tool.call.completed", "completed", now.Add(time.Second), map[string]any{
			"role":      "eval",
			"tool_name": "terminal",
			"result":    `{"status":"ok","exit_code":0}`,
		}),
	})
	if scopedCount != 1 {
		t.Fatalf("scoped count=%d, want 1", scopedCount)
	}
	if len(items) != 1 || items[0].RawEventIDs[0] != "main-1" {
		t.Fatalf("items=%#v, want only main event", items)
	}
}

func TestTraceActivityProjectorSyntheticQuietRowHasStableID(t *testing.T) {
	start := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	event := ledgerEvent("gen-1", 7, "tool.generation.started", "running", start, map[string]any{
		"tool_name": "rsi_slack_report_post",
	})
	firstProjector := traceActivityProjector{scope: "main", mode: "clean", now: start.Add(6 * time.Second)}
	secondProjector := traceActivityProjector{scope: "main", mode: "clean", now: start.Add(16 * time.Second)}
	first, _ := firstProjector.Project([]events.ExecutionLedgerEvent{event})
	second, _ := secondProjector.Project([]events.ExecutionLedgerEvent{event})
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("quiet rows len first=%d second=%d, want 1/1", len(first), len(second))
	}
	if first[0].ID != second[0].ID {
		t.Fatalf("quiet row id changed: %s vs %s", first[0].ID, second[0].ID)
	}
	if first[0].Revision == second[0].Revision {
		t.Fatalf("quiet row revision did not change with injected clock")
	}
}

func TestTraceActivityProjectorClosesSyntheticGenerationWhenPhaseFails(t *testing.T) {
	start := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: start.Add(30 * time.Second)}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("gen-1", 7, "tool.generation.started", "running", start, map[string]any{
			"tool_name": "delegate_task",
		}),
		ledgerEvent("phase-1", 8, "phase.completed", "failed", start.Add(20*time.Second), map[string]any{
			"termination_reason": "execution_error",
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want 1: %#v", len(items), items)
	}
	if items[0].Status != "failed" {
		t.Fatalf("synthetic generation status=%q, want failed", items[0].Status)
	}
	if items[0].CompletedAt == nil || !items[0].CompletedAt.Equal(start.Add(20*time.Second)) {
		t.Fatalf("completed_at=%v, want phase completion", items[0].CompletedAt)
	}
	if got := items[0].Details["terminal_phase_reason"]; got != "execution_error" {
		t.Fatalf("terminal reason=%#v, want execution_error", got)
	}
}

func TestTraceActivityProjectorClosesStartedToolWhenPhaseFails(t *testing.T) {
	start := time.Date(2026, 5, 14, 21, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: start.Add(time.Minute)}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("tool-1", 10, "tool.call.started", "running", start, map[string]any{
			"tool_name":    "terminal",
			"tool_call_id": "call-terminal",
		}),
		ledgerEvent("phase-1", 11, "phase.completed", "failed", start.Add(15*time.Second), map[string]any{
			"failure_class": "runner_non_ok",
		}),
	})
	if len(items) != 1 {
		t.Fatalf("items len=%d, want 1: %#v", len(items), items)
	}
	if items[0].Status != "failed" {
		t.Fatalf("tool status=%q, want failed", items[0].Status)
	}
	if items[0].DurationMS != 15000 {
		t.Fatalf("duration_ms=%d, want 15000", items[0].DurationMS)
	}
	if !containsString(items[0].SourceLedgerIDs, "phase-1") {
		t.Fatalf("source ledger IDs=%#v, want terminal phase event", items[0].SourceLedgerIDs)
	}
}

func TestTraceActivityProjectorDoesNotReopenCompletedToolWithDetachedProgressStart(t *testing.T) {
	start := time.Date(2026, 6, 19, 18, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: start.Add(time.Minute)}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("tool-previous-start", 1, "tool.call.started", "running", start, map[string]any{
			"tool_name":    "terminal",
			"tool_call_id": "call-previous",
			"args": map[string]any{
				"command": "ls /workspace/company",
			},
		}),
		ledgerEvent("tool-previous-done", 2, "tool.call.completed", "completed", start.Add(time.Second), map[string]any{
			"tool_name":    "terminal",
			"tool_call_id": "call-previous",
			"args": map[string]any{
				"command": "ls /workspace/company",
			},
			"result": `{"output":"ok","exit_code":0,"error":null}`,
		}),
		ledgerEvent("detached-next-start", 3, "tool.call.progress", "running", start.Add(2*time.Second), map[string]any{
			"tool_name":      "terminal",
			"progress_event": "tool.started",
			"args": map[string]any{
				"command": "gh pr create",
			},
		}),
		ledgerEvent("tool-next-start", 4, "tool.call.started", "running", start.Add(3*time.Second), map[string]any{
			"tool_name":    "terminal",
			"tool_call_id": "call-next",
			"args": map[string]any{
				"command": "gh pr create",
			},
		}),
		ledgerEvent("phase-1", 5, "phase.completed", "failed", start.Add(5*time.Second), map[string]any{
			"termination_reason": "execution_error",
		}),
	})

	previous := traceActivityItemByToolCallID(t, items, "call-previous")
	if previous.Status != "completed" {
		t.Fatalf("previous tool status=%q, want completed: %#v", previous.Status, previous)
	}
	if containsString(previous.SourceLedgerIDs, "detached-next-start") {
		t.Fatalf("previous source ledger IDs=%#v, should not include next detached start", previous.SourceLedgerIDs)
	}
	if containsString(previous.SourceLedgerIDs, "phase-1") {
		t.Fatalf("previous source ledger IDs=%#v, should not include terminal phase", previous.SourceLedgerIDs)
	}
	if _, ok := previous.Details["terminal_phase_event_id"]; ok {
		t.Fatalf("previous details=%#v, should not include terminal phase details", previous.Details)
	}

	next := traceActivityItemByToolCallID(t, items, "call-next")
	if next.Status != "failed" {
		t.Fatalf("next tool status=%q, want failed: %#v", next.Status, next)
	}
	if !containsString(next.SourceLedgerIDs, "detached-next-start") {
		t.Fatalf("next source ledger IDs=%#v, want detached start", next.SourceLedgerIDs)
	}
	if !containsString(next.SourceLedgerIDs, "phase-1") {
		t.Fatalf("next source ledger IDs=%#v, want terminal phase", next.SourceLedgerIDs)
	}
}

func TestTraceActivityInferStatusDoesNotDowngradeTerminalStatusToRunning(t *testing.T) {
	start := time.Date(2026, 6, 19, 18, 0, 0, 0, time.UTC)
	projector := traceActivityProjector{scope: "main", mode: "clean", now: start.Add(time.Minute)}
	items, _ := projector.Project([]events.ExecutionLedgerEvent{
		ledgerEvent("tool-start", 1, "tool.call.started", "running", start, map[string]any{
			"tool_name":    "terminal",
			"tool_call_id": "call-terminal",
		}),
		ledgerEvent("tool-done", 2, "tool.call.completed", "completed", start.Add(time.Second), map[string]any{
			"tool_name":    "terminal",
			"tool_call_id": "call-terminal",
			"result":       `{"output":"ok","exit_code":0,"error":null}`,
		}),
		ledgerEvent("late-running-progress", 3, "tool.call.progress", "running", start.Add(2*time.Second), map[string]any{
			"tool_name":    "terminal",
			"tool_call_id": "call-terminal",
		}),
		ledgerEvent("phase-1", 4, "phase.completed", "failed", start.Add(5*time.Second), map[string]any{
			"termination_reason": "execution_error",
		}),
	})

	item := traceActivityItemByToolCallID(t, items, "call-terminal")
	if item.Status != "completed" {
		t.Fatalf("tool status=%q, want completed: %#v", item.Status, item)
	}
	if containsString(item.SourceLedgerIDs, "phase-1") {
		t.Fatalf("source ledger IDs=%#v, should not include terminal phase", item.SourceLedgerIDs)
	}
}

func TestTraceActivityTruncateKeepsUTF8Valid(t *testing.T) {
	got := traceActivityTruncate("hello 世界 🚀", 9)
	if !utf8.ValidString(got) {
		t.Fatalf("truncated string is invalid UTF-8: %q", got)
	}
	if got != "hello 世界…" {
		t.Fatalf("truncated string=%q, want %q", got, "hello 世界…")
	}
	if one := traceActivityTruncate("🚀x", 1); one != "…" || !utf8.ValidString(one) {
		t.Fatalf("single-rune limit result=%q valid=%v, want ellipsis", one, utf8.ValidString(one))
	}
}

func traceActivityItemByToolCallID(t *testing.T, items []TraceActivityItem, callID string) TraceActivityItem {
	t.Helper()
	for _, item := range items {
		if item.ToolCallID == callID {
			return item
		}
	}
	t.Fatalf("missing activity item for tool_call_id=%s: %#v", callID, items)
	return TraceActivityItem{}
}

func ledgerEvent(id string, seq int, kind string, status string, at time.Time, payload map[string]any) events.ExecutionLedgerEvent {
	return events.ExecutionLedgerEvent{
		ID:          id,
		ExecutionID: "hexec-test",
		TraceID:     "trace-test",
		WorkflowID:  "wf-test",
		PhaseID:     "main",
		Kind:        kind,
		Status:      status,
		Seq:         seq,
		Payload:     payload,
		RecordedAt:  at,
	}
}
