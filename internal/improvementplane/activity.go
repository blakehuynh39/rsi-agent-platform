package improvementplane

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	traceActivityDefaultLimit        = 250
	traceActivityMaxLimit            = 500
	traceActivityQuietThreshold      = 5 * time.Second
	traceActivityNativeDedupeWindow  = 5 * time.Second
	traceActivityLargeTraceWarnCount = 10000
)

type TraceActivitySnapshot struct {
	TraceID     string               `json:"trace_id"`
	Scope       string               `json:"scope"`
	Mode        string               `json:"mode"`
	GeneratedAt time.Time            `json:"generated_at"`
	Items       []TraceActivityItem  `json:"items"`
	Paging      TraceActivityPaging  `json:"paging"`
	Metrics     TraceActivityMetrics `json:"metrics"`
}

type TraceActivityPaging struct {
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type TraceActivityMetrics struct {
	LedgerEventCount int   `json:"ledger_event_count"`
	ItemCount        int   `json:"item_count"`
	ProjectionMs     int64 `json:"projection_ms"`
	OversizedTrace   bool  `json:"oversized_trace,omitempty"`
}

type TraceActivityItem struct {
	ID              string         `json:"id"`
	Revision        string         `json:"revision"`
	Kind            string         `json:"kind"`
	Status          string         `json:"status"`
	Title           string         `json:"title"`
	Summary         string         `json:"summary,omitempty"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty"`
	DurationMS      int64          `json:"duration_ms,omitempty"`
	ToolName        string         `json:"tool_name,omitempty"`
	ToolCallID      string         `json:"tool_call_id,omitempty"`
	SourceLedgerIDs []string       `json:"source_ledger_ids"`
	RawEventIDs     []string       `json:"raw_event_ids"`
	Details         map[string]any `json:"details,omitempty"`
}

type traceActivityProjector struct {
	scope string
	mode  string
	now   time.Time
}

type traceActivityToolGroup struct {
	key      string
	kind     string
	status   string
	name     string
	callID   string
	started  time.Time
	finished time.Time
	events   []events.ExecutionLedgerEvent
	args     any
	result   map[string]any
	summary  string
	error    string
}

func buildTraceActivitySnapshot(store storepkg.Repository, traceID string, scope string, mode string, limit int, cursor string, now time.Time) (TraceActivitySnapshot, bool) {
	snapshot, _, ok := buildTraceActivitySnapshotWithHighWater(store, traceID, scope, mode, limit, cursor, now)
	return snapshot, ok
}

func buildTraceActivitySnapshotWithHighWater(store storepkg.Repository, traceID string, scope string, mode string, limit int, cursor string, now time.Time) (TraceActivitySnapshot, string, bool) {
	if !traceExistsForStream(store, traceID) {
		return TraceActivitySnapshot{}, "none", false
	}
	normalizedScope := normalizeTraceActivityScope(scope)
	normalizedMode := normalizeTraceActivityMode(mode)
	start := time.Now()
	eventsForTrace := store.ListExecutionLedgerEventsByTrace(traceID)
	highWater := traceActivityHighWaterID(eventsForTrace)
	projector := traceActivityProjector{scope: normalizedScope, mode: normalizedMode, now: now.UTC()}
	items, scopedCount := projector.Project(eventsForTrace)
	metrics := TraceActivityMetrics{
		LedgerEventCount: scopedCount,
		ItemCount:        len(items),
		ProjectionMs:     time.Since(start).Milliseconds(),
		OversizedTrace:   scopedCount > traceActivityLargeTraceWarnCount,
	}
	if metrics.OversizedTrace {
		log.Printf("improvement-plane activity projection oversized trace=%s scope=%s ledger_events=%d items=%d projection_ms=%d", traceID, normalizedScope, scopedCount, len(items), metrics.ProjectionMs)
	}
	page, paging := pageTraceActivityItems(items, limit, cursor)
	return TraceActivitySnapshot{
		TraceID:     traceID,
		Scope:       normalizedScope,
		Mode:        normalizedMode,
		GeneratedAt: now.UTC(),
		Items:       page,
		Paging:      paging,
		Metrics:     metrics,
	}, highWater, true
}

func (p traceActivityProjector) Project(raw []events.ExecutionLedgerEvent) ([]TraceActivityItem, int) {
	scope := normalizeTraceActivityScope(p.scope)
	mode := normalizeTraceActivityMode(p.mode)
	items := []events.ExecutionLedgerEvent{}
	for _, item := range raw {
		if !storepkg.LedgerEventMatchesScope(item, scope) {
			continue
		}
		if item.Payload == nil {
			item.Payload = map[string]any{}
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].RecordedAt.Equal(items[j].RecordedAt) {
			if items[i].ExecutionID == items[j].ExecutionID {
				if items[i].Seq == items[j].Seq {
					return items[i].ID < items[j].ID
				}
				return items[i].Seq < items[j].Seq
			}
			return items[i].ExecutionID < items[j].ExecutionID
		}
		return items[i].RecordedAt.Before(items[j].RecordedAt)
	})

	toolGroups, toolOrder, generationEvents := p.projectToolGroups(items)
	out := make([]TraceActivityItem, 0, len(toolOrder)+2)
	if todo := p.projectTodo(items); todo != nil {
		out = append(out, *todo)
	}
	for _, key := range toolOrder {
		group := toolGroups[key]
		if group == nil || len(group.events) == 0 || group.name == "todo" {
			continue
		}
		out = append(out, p.toolGroupItem(group, mode))
	}
	for _, gen := range generationEvents {
		if !p.generationMatchedTool(gen, toolGroups) {
			if item := p.syntheticGenerationItem(gen); item != nil {
				out = append(out, *item)
			}
		}
	}
	if final := p.projectFinalResponse(items, mode); final != nil {
		out = append(out, *final)
	}
	sort.SliceStable(out, func(i, j int) bool {
		left := activityItemSortTime(out[i])
		right := activityItemSortTime(out[j])
		if left.Equal(right) {
			return out[i].ID < out[j].ID
		}
		return left.Before(right)
	})
	return out, len(items)
}

func (p traceActivityProjector) projectToolGroups(items []events.ExecutionLedgerEvent) (map[string]*traceActivityToolGroup, []string, []events.ExecutionLedgerEvent) {
	groups := map[string]*traceActivityToolGroup{}
	order := []string{}
	generationEvents := []events.ExecutionLedgerEvent{}
	lastByName := map[string]string{}
	for _, item := range items {
		name := canonicalTraceActivityToolName(traceActivityToolName(item))
		if item.Kind == "tool.generation.started" && name != "" {
			generationEvents = append(generationEvents, item)
			continue
		}
		if traceActivityLooksLikeFinalResponse(item) && !strings.Contains(strings.ToLower(item.Kind), "tool.") {
			continue
		}
		if name == "" || !traceActivityLooksLikeTool(item) {
			continue
		}
		key := p.toolGroupKey(item, name, lastByName, groups)
		if key == "" {
			continue
		}
		group := groups[key]
		if group == nil {
			group = &traceActivityToolGroup{
				key:     key,
				kind:    "tool",
				name:    name,
				callID:  traceActivityToolCallID(item),
				started: item.RecordedAt,
			}
			groups[key] = group
			order = append(order, key)
		}
		group.events = append(group.events, item)
		group.name = firstNonEmptyString(group.name, name)
		group.callID = firstNonEmptyString(group.callID, traceActivityToolCallID(item))
		if group.started.IsZero() || item.RecordedAt.Before(group.started) {
			group.started = item.RecordedAt
		}
		if item.RecordedAt.After(group.finished) {
			group.finished = item.RecordedAt
		}
		if args := traceActivityToolArgs(item); args != nil && group.args == nil {
			group.args = args
		}
		if result := traceActivityResultPayload(item.Payload); len(result) > 0 {
			group.result = result
			if summary := strings.TrimSpace(stringValue(result["summary"])); summary != "" {
				group.summary = summary
			}
			if errText := strings.TrimSpace(stringValue(result["error"])); errText != "" {
				group.error = errText
			}
		}
		if summary := traceActivityPayloadSummary(item.Payload); summary != "" {
			group.summary = summary
		}
		if resultError := traceActivityPlainResultError(item.Payload); resultError != "" {
			group.error = resultError
		}
		if errText := strings.TrimSpace(stringValue(item.Payload["error"])); errText != "" {
			group.error = errText
		}
		group.status = traceActivityInferStatus(group.status, item.Status, group.result, item.Payload)
		lastByName[name] = key
	}
	p.attachGenerationEvents(generationEvents, groups)
	p.dedupeNativeEvents(groups)
	return groups, order, generationEvents
}

func (p traceActivityProjector) toolGroupKey(item events.ExecutionLedgerEvent, name string, lastByName map[string]string, groups map[string]*traceActivityToolGroup) string {
	if requestID := traceActivityDBReadRequestID(item); requestID != "" {
		return "dbread:" + requestID
	}
	if sqlHash := traceActivityDBReadSQLHash(item); sqlHash != "" && strings.HasPrefix(name, "db_read.") {
		return "dbread-sql:" + sqlHash
	}
	if callID := traceActivityToolCallID(item); callID != "" {
		return "toolcall:" + callID
	}
	if actionID := traceActivityNativeActionID(item); actionID != "" {
		return "native:" + actionID
	}
	if strings.Contains(strings.ToLower(item.Kind), "native_tool") {
		if key := p.findNearbyToolGroup(item, name, lastByName, groups); key != "" {
			return key
		}
	}
	if key := lastByName[name]; key != "" && !strings.Contains(strings.ToLower(item.Kind), "started") {
		return key
	}
	return fmt.Sprintf("event:%s", firstNonEmptyString(item.ID, fmt.Sprintf("%s:%d", name, item.Seq)))
}

func (p traceActivityProjector) findNearbyToolGroup(item events.ExecutionLedgerEvent, name string, lastByName map[string]string, groups map[string]*traceActivityToolGroup) string {
	key := lastByName[name]
	if key == "" {
		return ""
	}
	group := groups[key]
	if group == nil {
		return ""
	}
	anchor := group.finished
	if anchor.IsZero() {
		anchor = group.started
	}
	if anchor.IsZero() {
		return ""
	}
	delta := item.RecordedAt.Sub(anchor)
	if delta < 0 {
		delta = -delta
	}
	if delta > traceActivityNativeDedupeWindow {
		return ""
	}
	return key
}

func (p traceActivityProjector) attachGenerationEvents(generationEvents []events.ExecutionLedgerEvent, groups map[string]*traceActivityToolGroup) {
	for _, gen := range generationEvents {
		name := canonicalTraceActivityToolName(traceActivityToolName(gen))
		if name == "" {
			continue
		}
		var best *traceActivityToolGroup
		bestDelta := time.Duration(1<<63 - 1)
		for _, group := range groups {
			if group.name != name || group.started.IsZero() || group.started.Before(gen.RecordedAt) {
				continue
			}
			delta := group.started.Sub(gen.RecordedAt)
			if delta <= time.Minute && delta < bestDelta {
				best = group
				bestDelta = delta
			}
		}
		if best != nil {
			best.events = append([]events.ExecutionLedgerEvent{gen}, best.events...)
			if gen.RecordedAt.Before(best.started) {
				best.started = gen.RecordedAt
			}
		}
	}
}

func (p traceActivityProjector) dedupeNativeEvents(groups map[string]*traceActivityToolGroup) {
	for _, nativeGroup := range groups {
		if nativeGroup == nil || !traceActivityGroupHasNativeOnly(nativeGroup) {
			continue
		}
		for _, target := range groups {
			if target == nil || target == nativeGroup || target.name != nativeGroup.name {
				continue
			}
			if traceActivityNativeGroupsMatch(nativeGroup, target) {
				target.events = append(target.events, nativeGroup.events...)
				if target.summary == "" {
					target.summary = nativeGroup.summary
				}
				if target.error == "" {
					target.error = nativeGroup.error
				}
				nativeGroup.events = nil
				break
			}
		}
	}
}

func traceActivityGroupHasNativeOnly(group *traceActivityToolGroup) bool {
	if len(group.events) == 0 {
		return false
	}
	for _, item := range group.events {
		if !strings.Contains(strings.ToLower(item.Kind), "native_tool") {
			return false
		}
	}
	return true
}

func traceActivityNativeGroupsMatch(nativeGroup *traceActivityToolGroup, target *traceActivityToolGroup) bool {
	nativeAction := ""
	for _, item := range nativeGroup.events {
		nativeAction = firstNonEmptyString(nativeAction, traceActivityNativeActionID(item))
	}
	targetAction := ""
	for _, item := range target.events {
		targetAction = firstNonEmptyString(targetAction, traceActivityNativeActionID(item))
	}
	if nativeAction != "" && targetAction != "" && nativeAction == targetAction {
		return true
	}
	if nativeGroup.finished.IsZero() || target.finished.IsZero() {
		return false
	}
	delta := target.finished.Sub(nativeGroup.finished)
	if delta < 0 {
		delta = -delta
	}
	return delta <= traceActivityNativeDedupeWindow
}

func (p traceActivityProjector) toolGroupItem(group *traceActivityToolGroup, mode string) TraceActivityItem {
	sourceIDs := traceActivityLedgerIDs(group.events)
	status := traceActivityNormalizeStatus(group.status)
	if status == "" {
		status = "running"
	}
	title, summary, details := p.summarizeToolGroup(group, mode)
	item := TraceActivityItem{
		ID:              fmt.Sprintf("activity:%s:tool:%s", p.scope, traceActivitySafeID(group.key)),
		Kind:            "tool",
		Status:          status,
		Title:           title,
		Summary:         summary,
		ToolName:        group.name,
		ToolCallID:      group.callID,
		SourceLedgerIDs: sourceIDs,
		RawEventIDs:     sourceIDs,
		Details:         details,
	}
	if !group.started.IsZero() {
		started := group.started.UTC()
		item.StartedAt = &started
	}
	if !group.finished.IsZero() && status != "running" {
		completed := group.finished.UTC()
		item.CompletedAt = &completed
	}
	if item.StartedAt != nil && item.CompletedAt != nil {
		item.DurationMS = item.CompletedAt.Sub(*item.StartedAt).Milliseconds()
	}
	item.Revision = traceActivityRevision(item)
	return item
}

func (p traceActivityProjector) summarizeToolGroup(group *traceActivityToolGroup, mode string) (string, string, map[string]any) {
	details := map[string]any{
		"type":    traceActivityDetailType(group.name),
		"version": 1,
	}
	if group.callID != "" {
		details["tool_call_id"] = group.callID
	}
	if actionID := traceActivityGroupNativeActionID(group); actionID != "" {
		details["native_action_id"] = actionID
	}
	if mode == "detailed" {
		if group.args != nil {
			details["args_excerpt"] = traceActivityBoundedJSON(group.args, 3000)
		}
		if len(group.result) > 0 {
			details["result_excerpt"] = traceActivityBoundedJSON(group.result, 5000)
		}
	}
	if group.error != "" {
		details["error_excerpt"] = traceActivityTruncate(group.error, 500)
	}
	switch {
	case strings.HasPrefix(group.name, "rsi_slack."):
		return traceActivitySlackSummary(group, details)
	case strings.HasPrefix(group.name, "rsi_notion."):
		return traceActivityNotionSummary(group, details)
	case strings.HasPrefix(group.name, "rsi_knowledge."):
		return traceActivityKnowledgeSummary(group, details)
	case strings.HasPrefix(group.name, "rsi_sentry."):
		return traceActivitySentrySummary(group, details)
	case strings.HasPrefix(group.name, "rsi_observability."):
		return traceActivityObservabilitySummary(group, details)
	case strings.HasPrefix(group.name, "db_read."):
		return traceActivityDBReadSummary(group, details)
	case group.name == "terminal":
		return traceActivityTerminalSummary(group, details)
	case group.name == "delegate_task":
		return "Delegated work", firstNonEmptyString(group.summary, "Delegated a subtask."), details
	case strings.Contains(group.name, "read_file") || strings.Contains(group.name, "search"):
		return traceActivityFileSummary(group, details)
	case strings.HasPrefix(group.name, "skill_"):
		return "Loaded skill context", firstNonEmptyString(group.summary, "Read Hermes skill instructions."), details
	default:
		return group.name, traceActivityGenericToolSummary(group), details
	}
}

func traceActivitySlackSummary(group *traceActivityToolGroup, details map[string]any) (string, string, map[string]any) {
	op := strings.TrimPrefix(group.name, "rsi_slack.")
	title := "Slack " + strings.ReplaceAll(op, "_", " ")
	if delivery, ok := group.result["reply_delivery"].(map[string]any); ok {
		if ref := strings.TrimSpace(stringValue(delivery["provider_ref"])); ref != "" {
			details["provider_ref"] = ref
		}
		if status := strings.TrimSpace(stringValue(delivery["send_status"])); status != "" {
			details["send_status"] = status
		}
	}
	return title, firstNonEmptyString(group.summary, traceActivityOperationSummary("Slack", op)), details
}

func traceActivityNotionSummary(group *traceActivityToolGroup, details map[string]any) (string, string, map[string]any) {
	op := strings.TrimPrefix(group.name, "rsi_notion.")
	return "Notion " + strings.ReplaceAll(op, "_", " "), firstNonEmptyString(group.summary, traceActivityOperationSummary("Notion", op)), details
}

func traceActivityKnowledgeSummary(group *traceActivityToolGroup, details map[string]any) (string, string, map[string]any) {
	op := strings.TrimPrefix(group.name, "rsi_knowledge.")
	count := traceActivityCountFromResult(group.result, "results", "documents", "pages", "messages")
	if count >= 0 {
		details["result_count"] = count
	}
	return "Knowledge " + strings.ReplaceAll(op, "_", " "), firstNonEmptyString(group.summary, traceActivityCountSummary("Knowledge", op, count)), details
}

func traceActivitySentrySummary(group *traceActivityToolGroup, details map[string]any) (string, string, map[string]any) {
	op := strings.TrimPrefix(group.name, "rsi_sentry.")
	count := traceActivityCountFromResult(group.result, "issues", "events", "projects", "releases")
	if count >= 0 {
		details["result_count"] = count
	}
	return "Sentry " + strings.ReplaceAll(op, "_", " "), firstNonEmptyString(group.summary, traceActivityCountSummary("Sentry", op, count)), details
}

func traceActivityObservabilitySummary(group *traceActivityToolGroup, details map[string]any) (string, string, map[string]any) {
	op := strings.TrimPrefix(group.name, "rsi_observability.")
	count := traceActivityCountFromResult(group.result, "log_lines", "data", "datasources", "dashboards", "alerts", "rules")
	if count >= 0 {
		details["result_count"] = count
	}
	if expr := traceActivityStringFromArgs(group.args, "expr"); expr != "" {
		details["query"] = traceActivityTruncate(expr, 500)
	}
	return "Observability " + strings.ReplaceAll(op, "_", " "), firstNonEmptyString(group.summary, traceActivityCountSummary("Observability", op, count)), details
}

func traceActivityDBReadSummary(group *traceActivityToolGroup, details map[string]any) (string, string, map[string]any) {
	requestID := ""
	target := ""
	sqlHash := ""
	rows := -1
	truncated := false
	for _, item := range group.events {
		requestID = firstNonEmptyString(requestID, traceActivityDBReadRequestID(item))
		target = firstNonEmptyString(target, traceActivityPayloadNestedString(item.Payload, "target"))
		sqlHash = firstNonEmptyString(sqlHash, traceActivityDBReadSQLHash(item))
		if rows < 0 {
			rows = traceActivityPayloadInt(item.Payload, "row_count", "rows")
		}
		truncated = truncated || traceActivityPayloadBool(item.Payload, "truncated")
	}
	if requestID != "" {
		details["request_id"] = requestID
	}
	if target != "" {
		details["target"] = target
	}
	if sqlHash != "" {
		details["sql_sha256"] = sqlHash
	}
	if rows >= 0 {
		details["rows"] = rows
	}
	if truncated {
		details["truncated"] = true
	}
	parts := []string{}
	if target != "" {
		parts = append(parts, target)
	}
	if rows >= 0 {
		parts = append(parts, fmt.Sprintf("%d row(s)", rows))
	}
	if truncated {
		parts = append(parts, "truncated")
	}
	return "DB read", firstNonEmptyString(group.summary, strings.Join(parts, " · "), "DB read request updated."), details
}

func traceActivityTerminalSummary(group *traceActivityToolGroup, details map[string]any) (string, string, map[string]any) {
	command := traceActivityStringFromArgs(group.args, "command", "cmd")
	if command != "" {
		details["command"] = traceActivityTruncate(command, 1000)
	}
	exitCode := traceActivityIntFromResult(group.result, "exit_code")
	if exitCode >= 0 {
		details["exit_code"] = exitCode
	}
	output := firstNonEmptyString(stringValue(group.result["output"]), stringValue(group.result["stdout"]))
	if output != "" && details["result_excerpt"] == nil {
		details["output_excerpt"] = traceActivityTruncate(output, 800)
	}
	summary := firstNonEmptyString(group.summary, traceActivityTruncate(command, 120))
	if exitCode >= 0 {
		summary = fmt.Sprintf("exit %d", exitCode)
		if command != "" {
			summary += " · " + traceActivityTruncate(command, 120)
		}
	}
	return "Terminal", summary, details
}

func traceActivityFileSummary(group *traceActivityToolGroup, details map[string]any) (string, string, map[string]any) {
	path := firstNonEmptyString(traceActivityStringFromArgs(group.args, "path", "file", "pattern"), traceActivityStringFromResult(group.result, "path"))
	if path != "" {
		details["path"] = path
	}
	return strings.ReplaceAll(group.name, "_", " "), firstNonEmptyString(group.summary, path, "File/search tool completed."), details
}

func traceActivityGenericToolSummary(group *traceActivityToolGroup) string {
	if group.summary != "" {
		return traceActivityTruncate(group.summary, 240)
	}
	if group.error != "" {
		return traceActivityTruncate(group.error, 240)
	}
	argsSize := len(traceActivityBoundedJSON(group.args, 100000))
	resultSize := len(traceActivityBoundedJSON(group.result, 100000))
	parts := []string{}
	if argsSize > 0 {
		parts = append(parts, fmt.Sprintf("args %dB", argsSize))
	}
	if resultSize > 0 {
		parts = append(parts, fmt.Sprintf("result %dB", resultSize))
	}
	return firstNonEmptyString(strings.Join(parts, " · "), "Tool call completed.")
}

func (p traceActivityProjector) projectTodo(items []events.ExecutionLedgerEvent) *TraceActivityItem {
	var latest *events.ExecutionLedgerEvent
	var todos []map[string]any
	sourceIDs := []string{}
	for i := range items {
		item := items[i]
		name := canonicalTraceActivityToolName(traceActivityToolName(item))
		if name != "todo" || !strings.Contains(strings.ToLower(item.Kind), "completed") {
			continue
		}
		result := traceActivityResultPayload(item.Payload)
		if parsed := traceActivityTodosFromPayload(result); len(parsed) > 0 {
			todos = parsed
			latest = &items[i]
			sourceIDs = append(sourceIDs, item.ID)
			continue
		}
		if parsed := traceActivityTodosFromPayload(item.Payload); len(parsed) > 0 {
			todos = parsed
			latest = &items[i]
			sourceIDs = append(sourceIDs, item.ID)
		}
	}
	if latest == nil || len(todos) == 0 {
		return nil
	}
	statusCounts := map[string]int{}
	current := ""
	for _, todo := range todos {
		status := strings.TrimSpace(strings.ToLower(stringValue(todo["status"])))
		if status == "" {
			status = "pending"
		}
		statusCounts[status]++
		if status == "in_progress" && current == "" {
			current = strings.TrimSpace(stringValue(todo["content"]))
		}
	}
	status := "completed"
	if statusCounts["in_progress"] > 0 {
		status = "running"
	} else if statusCounts["pending"] > 0 {
		status = "pending"
	} else if statusCounts["cancelled"] > 0 && statusCounts["completed"] == 0 {
		status = "cancelled"
	}
	summary := fmt.Sprintf("%d/%d complete", statusCounts["completed"], len(todos))
	if current != "" {
		summary += " · " + traceActivityTruncate(current, 100)
	}
	detailsTodos := []map[string]any{}
	for _, todo := range todos {
		detailsTodos = append(detailsTodos, map[string]any{
			"id":      stringValue(todo["id"]),
			"content": stringValue(todo["content"]),
			"status":  firstNonEmptyString(strings.ToLower(stringValue(todo["status"])), "pending"),
		})
	}
	completed := latest.RecordedAt.UTC()
	item := TraceActivityItem{
		ID:              "activity:" + p.scope + ":todo",
		Kind:            "todo",
		Status:          status,
		Title:           "Todo list",
		Summary:         summary,
		CompletedAt:     &completed,
		SourceLedgerIDs: traceActivityUniqueStrings(sourceIDs),
		RawEventIDs:     traceActivityUniqueStrings(sourceIDs),
		Details: map[string]any{
			"type":    "todo",
			"version": 1,
			"todos":   detailsTodos,
			"summary": map[string]any{
				"total":       len(todos),
				"pending":     statusCounts["pending"],
				"in_progress": statusCounts["in_progress"],
				"completed":   statusCounts["completed"],
				"cancelled":   statusCounts["cancelled"],
			},
		},
	}
	item.Revision = traceActivityRevision(item)
	return &item
}

func traceActivityTodosFromPayload(payload map[string]any) []map[string]any {
	if len(payload) == 0 {
		return nil
	}
	raw, ok := payload["todos"]
	if !ok {
		raw = payload["todo"]
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := []map[string]any{}
	for _, item := range items {
		if typed, ok := item.(map[string]any); ok {
			out = append(out, typed)
		}
	}
	return out
}

func (p traceActivityProjector) projectFinalResponse(items []events.ExecutionLedgerEvent, mode string) *TraceActivityItem {
	var delivery *events.ExecutionLedgerEvent
	for i := range items {
		if traceActivityLooksLikeFinalResponse(items[i]) {
			delivery = &items[i]
		}
	}
	if delivery == nil {
		return nil
	}
	body := firstNonEmptyString(
		stringValue(delivery.Payload["body_excerpt"]),
		stringValue(delivery.Payload["body"]),
		stringValue(delivery.Payload["summary"]),
		stringValue(delivery.Payload["final_response"]),
		stringValue(delivery.Payload["final_answer"]),
		stringValue(delivery.Payload["reply_draft"]),
	)
	status := traceActivityNormalizeStatus(firstNonEmptyString(stringValue(delivery.Payload["send_status"]), delivery.Status))
	if status == "" {
		status = "completed"
	}
	completed := delivery.RecordedAt.UTC()
	details := map[string]any{
		"type":    "final_response",
		"version": 1,
	}
	for _, key := range []string{"delivery_id", "provider_ref", "message_link", "channel_id", "thread_ts", "tool_name", "transport_tool_name", "native_action_id", "send_status", "status_code"} {
		if value := stringValue(delivery.Payload[key]); value != "" {
			details[key] = value
		}
	}
	if refs := traceActivityStringSlice(delivery.Payload["artifact_refs"]); len(refs) > 0 {
		details["artifact_refs"] = refs
	}
	if mode == "detailed" {
		if full := stringValue(delivery.Payload["body"]); full != "" {
			details["body_excerpt"] = traceActivityTruncate(full, 3000)
		}
	}
	item := TraceActivityItem{
		ID:              "activity:" + p.scope + ":final_response",
		Kind:            "final_response",
		Status:          status,
		Title:           "Final response",
		Summary:         traceActivityTruncate(body, 500),
		CompletedAt:     &completed,
		SourceLedgerIDs: []string{delivery.ID},
		RawEventIDs:     []string{delivery.ID},
		Details:         details,
	}
	item.Revision = traceActivityRevision(item)
	return &item
}

func traceActivityLooksLikeFinalResponse(item events.ExecutionLedgerEvent) bool {
	kind := strings.ToLower(strings.TrimSpace(item.Kind))
	switch {
	case kind == "reply_delivery", kind == "model.reply_delivery", kind == "slack.message.sent":
		return true
	case strings.HasPrefix(kind, "slack.direct_delivery."):
		return true
	case kind == "final_response" || kind == "model.final_response":
		return true
	}
	if firstNonEmptyString(
		stringValue(item.Payload["final_response"]),
		stringValue(item.Payload["final_answer"]),
		stringValue(item.Payload["reply_draft"]),
	) != "" {
		return true
	}
	if firstNonEmptyString(stringValue(item.Payload["send_status"]), stringValue(item.Payload["delivery_id"])) == "" {
		return false
	}
	return firstNonEmptyString(
		stringValue(item.Payload["body_excerpt"]),
		stringValue(item.Payload["body"]),
		stringValue(item.Payload["provider_ref"]),
		stringValue(item.Payload["message_link"]),
	) != ""
}

func (p traceActivityProjector) generationMatchedTool(gen events.ExecutionLedgerEvent, groups map[string]*traceActivityToolGroup) bool {
	name := canonicalTraceActivityToolName(traceActivityToolName(gen))
	for _, group := range groups {
		if group == nil || group.name != name {
			continue
		}
		for _, item := range group.events {
			if item.ID == gen.ID {
				return true
			}
		}
	}
	return false
}

func (p traceActivityProjector) syntheticGenerationItem(gen events.ExecutionLedgerEvent) *TraceActivityItem {
	name := canonicalTraceActivityToolName(traceActivityToolName(gen))
	if name == "" {
		return nil
	}
	if p.now.Sub(gen.RecordedAt) < traceActivityQuietThreshold {
		return nil
	}
	started := gen.RecordedAt.UTC()
	details := map[string]any{
		"type":    "tool_generation",
		"version": 1,
	}
	quietSeconds := int(p.now.Sub(gen.RecordedAt).Seconds())
	if quietSeconds > 0 {
		// Bound the clock-derived detail so live streams do not re-render a
		// synthetic quiet row on every one-second projection poll.
		details["quiet_seconds"] = (quietSeconds / 10) * 10
	}
	item := TraceActivityItem{
		ID:              fmt.Sprintf("activity:%s:synthetic:tool_generation:%s:%d:%s", p.scope, traceActivitySafeID(gen.ExecutionID), gen.Seq, traceActivitySafeID(name)),
		Kind:            "tool_generation",
		Status:          "running",
		Title:           "Preparing " + name,
		Summary:         "Generating tool call arguments.",
		StartedAt:       &started,
		ToolName:        name,
		SourceLedgerIDs: []string{gen.ID},
		RawEventIDs:     []string{gen.ID},
		Details:         details,
	}
	item.Revision = traceActivityRevision(item)
	return &item
}

func traceActivityLooksLikeTool(item events.ExecutionLedgerEvent) bool {
	kind := strings.ToLower(item.Kind)
	return strings.Contains(kind, "tool") || traceActivityToolName(item) != ""
}

func traceActivityToolName(item events.ExecutionLedgerEvent) string {
	return firstNonEmptyString(
		stringValue(item.Payload["tool_name"]),
		stringValue(item.Payload["name"]),
		stringValue(item.Payload["tool"]),
		stringValue(item.Payload["transport_tool_name"]),
	)
}

func canonicalTraceActivityToolName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	explicit := map[string]string{
		"rsi_slack_channels_list":              "rsi_slack.channels_list",
		"rsi_slack_channel_info":               "rsi_slack.channel_info",
		"rsi_slack_conversation_read":          "rsi_slack.conversation_read",
		"rsi_slack_user_lookup":                "rsi_slack.user_lookup",
		"rsi_slack_message_post":               "rsi_slack.message_post",
		"rsi_slack_report_post":                "rsi_slack.report_post",
		"rsi_slack_message_update":             "rsi_slack.message_update",
		"rsi_slack_message_delete":             "rsi_slack.message_delete",
		"rsi_slack_reaction_add":               "rsi_slack.reaction_add",
		"rsi_slack_reaction_remove":            "rsi_slack.reaction_remove",
		"rsi_slack_file_upload":                "rsi_slack.file_upload",
		"rsi_slack_channel_create":             "rsi_slack.channel_create",
		"rsi_slack_channel_rename":             "rsi_slack.channel_rename",
		"rsi_slack_channel_archive":            "rsi_slack.channel_archive",
		"rsi_slack_channel_invite":             "rsi_slack.channel_invite",
		"db_read_sources":                      "db_read.sources",
		"db_read_schema":                       "db_read.schema",
		"db_read_validate":                     "db_read.validate",
		"db_read_query":                        "db_read.query",
		"db_read_status":                       "db_read.status",
		"rsi_notion_blocks_children":           "rsi_notion.blocks_children",
		"rsi_notion_data_source_get":           "rsi_notion.data_source_get",
		"rsi_notion_data_source_query":         "rsi_notion.data_source_query",
		"rsi_knowledge_document_get":           "rsi_knowledge.document_get",
		"rsi_knowledge_conversation_get":       "rsi_knowledge.conversation_get",
		"rsi_knowledge_messages_read":          "rsi_knowledge.messages_read",
		"rsi_knowledge_wiki_search":            "rsi_knowledge.wiki_search",
		"rsi_knowledge_wiki_page_get":          "rsi_knowledge.wiki_page_get",
		"rsi_knowledge_wiki_index_get":         "rsi_knowledge.wiki_index_get",
		"rsi_knowledge_wiki_log_get":           "rsi_knowledge.wiki_log_get",
		"rsi_knowledge_source_status":          "rsi_knowledge.source_status",
		"rsi_knowledge_wiki_edit_propose":      "rsi_knowledge.wiki_edit_propose",
		"rsi_knowledge_wiki_edit_apply":        "rsi_knowledge.wiki_edit_apply",
		"rsi_sentry_projects_list":             "rsi_sentry.projects_list",
		"rsi_sentry_issues_list":               "rsi_sentry.issues_list",
		"rsi_sentry_issue_view":                "rsi_sentry.issue_view",
		"rsi_sentry_issue_events":              "rsi_sentry.issue_events",
		"rsi_sentry_releases_list":             "rsi_sentry.releases_list",
		"rsi_observability_datasources":        "rsi_observability.datasources",
		"rsi_observability_metrics_query":      "rsi_observability.metrics_query",
		"rsi_observability_logs_query":         "rsi_observability.logs_query",
		"rsi_observability_dashboards_search":  "rsi_observability.dashboards_search",
		"rsi_observability_dashboard_get":      "rsi_observability.dashboard_get",
		"rsi_observability_alert_rules_search": "rsi_observability.alert_rules_search",
		"rsi_observability_alert_rule_get":     "rsi_observability.alert_rule_get",
		"rsi_observability_active_alerts":      "rsi_observability.active_alerts",
	}
	if mapped, ok := explicit[name]; ok {
		return mapped
	}
	prefixes := []string{"rsi_notion", "rsi_knowledge", "rsi_sentry", "rsi_observability", "rsi_slack", "db_read"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix+"_") {
			return prefix + "." + strings.TrimPrefix(name, prefix+"_")
		}
	}
	return name
}

func traceActivityToolCallID(item events.ExecutionLedgerEvent) string {
	return firstNonEmptyString(
		stringValue(item.Payload["tool_call_id"]),
		stringValue(item.Payload["tool_id"]),
		stringValue(item.Payload["call_id"]),
	)
}

func traceActivityToolArgs(item events.ExecutionLedgerEvent) any {
	for _, key := range []string{"args", "arguments", "request", "request_payload", "input"} {
		if value, ok := item.Payload[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func traceActivityResultPayload(payload map[string]any) map[string]any {
	if len(payload) == 0 {
		return nil
	}
	switch result := payload["result"].(type) {
	case map[string]any:
		return result
	case string:
		var parsed map[string]any
		if err := json.Unmarshal([]byte(result), &parsed); err == nil {
			return parsed
		}
	}
	if output, ok := payload["output"].(map[string]any); ok {
		return output
	}
	return nil
}

func traceActivityInferStatus(current string, eventStatus string, result map[string]any, payload map[string]any) string {
	for _, candidate := range []string{
		stringValue(result["status"]),
		stringValue(payload["status"]),
		eventStatus,
	} {
		normalized := traceActivityNormalizeStatus(candidate)
		if normalized == "failed" || normalized == "cancelled" {
			return normalized
		}
		if normalized != "" {
			current = normalized
		}
	}
	if ok, hasOK := traceActivityBoolValue(result["ok"]); hasOK && !ok {
		return "failed"
	}
	if errText := firstNonEmptyString(stringValue(result["error"]), stringValue(payload["error"])); errText != "" {
		return "failed"
	}
	if traceActivityPlainResultError(payload) != "" {
		return "failed"
	}
	return current
}

func traceActivityPlainResultError(payload map[string]any) string {
	result := strings.TrimSpace(stringValue(payload["result"]))
	if strings.HasPrefix(strings.ToLower(result), "error:") {
		return traceActivityTruncate(result, 500)
	}
	return ""
}

func traceActivityNormalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "ok", "success", "succeeded", "complete", "completed", "posted", "delivered":
		return "completed"
	case "failed", "failure", "error", "blocked":
		return "failed"
	case "running", "streaming", "started", "in_progress", "pending":
		return "running"
	case "cancelled", "canceled", "expired", "denied":
		return "cancelled"
	default:
		return strings.TrimSpace(status)
	}
}

func traceActivityPayloadSummary(payload map[string]any) string {
	for _, key := range []string{"summary", "message", "response_summary", "preview"} {
		if value := strings.TrimSpace(stringValue(payload[key])); value != "" {
			return traceActivityTruncate(value, 500)
		}
	}
	return ""
}

func traceActivityDBReadRequestID(item events.ExecutionLedgerEvent) string {
	for _, payload := range []map[string]any{item.Payload, traceActivityResultPayload(item.Payload)} {
		if len(payload) == 0 {
			continue
		}
		if value := firstNonEmptyString(
			stringValue(payload["request_id"]),
			stringValue(payload["request_ref"]),
			stringValue(payload["db_read_request_id"]),
		); value != "" {
			return value
		}
	}
	return ""
}

func traceActivityDBReadSQLHash(item events.ExecutionLedgerEvent) string {
	for _, payload := range []map[string]any{item.Payload, traceActivityResultPayload(item.Payload)} {
		if len(payload) == 0 {
			continue
		}
		if value := firstNonEmptyString(
			stringValue(payload["sql_sha256"]),
			stringValue(payload["sql_hash"]),
			stringValue(payload["hash"]),
		); value != "" {
			return value
		}
	}
	return ""
}

func traceActivityNativeActionID(item events.ExecutionLedgerEvent) string {
	if value := firstNonEmptyString(
		stringValue(item.Payload["native_action_id"]),
		stringValue(item.Payload["action_id"]),
	); value != "" {
		return value
	}
	if action, ok := item.Payload["action"].(map[string]any); ok {
		if value := firstNonEmptyString(stringValue(action["id"]), stringValue(action["action_id"])); value != "" {
			return value
		}
	}
	if result := traceActivityResultPayload(item.Payload); len(result) > 0 {
		if value := firstNonEmptyString(stringValue(result["native_action_id"]), stringValue(result["action_id"])); value != "" {
			return value
		}
		if action, ok := result["action"].(map[string]any); ok {
			return firstNonEmptyString(stringValue(action["id"]), stringValue(action["action_id"]))
		}
		if output, ok := result["output"].(map[string]any); ok {
			if action, ok := output["action"].(map[string]any); ok {
				return firstNonEmptyString(stringValue(action["id"]), stringValue(action["action_id"]))
			}
		}
	}
	return ""
}

func traceActivityGroupNativeActionID(group *traceActivityToolGroup) string {
	for _, item := range group.events {
		if id := traceActivityNativeActionID(item); id != "" {
			return id
		}
	}
	return ""
}

func traceActivityLedgerIDs(items []events.ExecutionLedgerEvent) []string {
	out := []string{}
	for _, item := range items {
		if item.ID != "" {
			out = append(out, item.ID)
		}
	}
	return traceActivityUniqueStrings(out)
}

func traceActivityRevision(item TraceActivityItem) string {
	payload := map[string]any{
		"id":      item.ID,
		"kind":    item.Kind,
		"status":  item.Status,
		"title":   item.Title,
		"summary": item.Summary,
		"source":  item.SourceLedgerIDs,
		"details": item.Details,
	}
	raw, _ := json.Marshal(payload)
	sum := sha1.Sum(raw)
	return hex.EncodeToString(sum[:])[:16]
}

func pageTraceActivityItems(items []TraceActivityItem, limit int, cursor string) ([]TraceActivityItem, TraceActivityPaging) {
	if limit <= 0 {
		limit = traceActivityDefaultLimit
	}
	if limit > traceActivityMaxLimit {
		limit = traceActivityMaxLimit
	}
	end := len(items)
	if cursor != "" {
		for index, item := range items {
			if item.ID == cursor {
				end = index
				break
			}
		}
	}
	start := end - limit
	if start < 0 {
		start = 0
	}
	page := append([]TraceActivityItem(nil), items[start:end]...)
	paging := TraceActivityPaging{Limit: limit, HasMore: start > 0}
	if paging.HasMore && len(page) > 0 {
		paging.NextCursor = page[0].ID
	}
	return page, paging
}

func traceActivityHighWaterID(items []events.ExecutionLedgerEvent) string {
	if len(items) == 0 {
		return "none"
	}
	latest := items[len(items)-1]
	return firstNonEmptyString(latest.ID, fmt.Sprintf("%s:%d", latest.ExecutionID, latest.Seq))
}

func normalizeTraceActivityScope(scope string) string {
	scope = strings.TrimSpace(strings.ToLower(scope))
	if scope == "" {
		return "main"
	}
	return scope
}

func normalizeTraceActivityMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "detailed" {
		return "detailed"
	}
	return "clean"
}

func traceActivitySafeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(":", "_", "/", "_", " ", "_", "\n", "_", "\r", "_")
	return replacer.Replace(value)
}

func activityItemSortTime(item TraceActivityItem) time.Time {
	if item.StartedAt != nil {
		return *item.StartedAt
	}
	if item.CompletedAt != nil {
		return *item.CompletedAt
	}
	return time.Time{}
}

func traceActivityDetailType(toolName string) string {
	switch {
	case strings.HasPrefix(toolName, "rsi_slack."):
		return "slack"
	case strings.HasPrefix(toolName, "rsi_notion."):
		return "notion"
	case strings.HasPrefix(toolName, "rsi_knowledge."):
		return "knowledge"
	case strings.HasPrefix(toolName, "rsi_sentry."):
		return "sentry"
	case strings.HasPrefix(toolName, "rsi_observability."):
		return "observability"
	case strings.HasPrefix(toolName, "db_read."):
		return "db_read"
	case toolName == "terminal":
		return "terminal"
	case toolName == "todo":
		return "todo"
	default:
		return "tool"
	}
}

func traceActivityOperationSummary(surface string, op string) string {
	return fmt.Sprintf("%s %s completed.", surface, strings.ReplaceAll(op, "_", " "))
}

func traceActivityCountSummary(surface string, op string, count int) string {
	if count >= 0 {
		return fmt.Sprintf("Returned %d %s result(s).", count, surface)
	}
	return traceActivityOperationSummary(surface, op)
}

func traceActivityCountFromResult(result map[string]any, keys ...string) int {
	for _, key := range keys {
		value, ok := result[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case []any:
			return len(typed)
		case []map[string]any:
			return len(typed)
		case float64:
			return int(typed)
		case int:
			return typed
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
				return parsed
			}
		}
	}
	if output, ok := result["output"].(map[string]any); ok {
		return traceActivityCountFromResult(output, keys...)
	}
	return -1
}

func traceActivityStringFromArgs(args any, keys ...string) string {
	values, ok := args.(map[string]any)
	if !ok {
		return ""
	}
	for _, key := range keys {
		if value := strings.TrimSpace(stringValue(values[key])); value != "" {
			return value
		}
	}
	return ""
}

func traceActivityStringFromResult(result map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(stringValue(result[key])); value != "" {
			return value
		}
	}
	return ""
}

func traceActivityIntFromResult(result map[string]any, keys ...string) int {
	for _, key := range keys {
		switch typed := result[key].(type) {
		case int:
			return typed
		case float64:
			return int(typed)
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
				return parsed
			}
		}
	}
	return -1
}

func traceActivityPayloadInt(payload map[string]any, keys ...string) int {
	for _, key := range keys {
		if value := traceActivityIntFromResult(payload, key); value >= 0 {
			return value
		}
	}
	return -1
}

func traceActivityPayloadBool(payload map[string]any, key string) bool {
	value, _ := traceActivityBoolValue(payload[key])
	return value
}

func traceActivityBoolValue(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "yes", "1":
			return true, true
		case "false", "no", "0":
			return false, true
		}
	}
	return false, false
}

func traceActivityPayloadNestedString(payload map[string]any, key string) string {
	if value := strings.TrimSpace(stringValue(payload[key])); value != "" {
		return value
	}
	if result := traceActivityResultPayload(payload); len(result) > 0 {
		return strings.TrimSpace(stringValue(result[key]))
	}
	return ""
}

func traceActivityStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := []string{}
		for _, item := range typed {
			if value := strings.TrimSpace(stringValue(item)); value != "" {
				out = append(out, value)
			}
		}
		return out
	default:
		return nil
	}
}

func traceActivityUniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func traceActivityBoundedJSON(value any, limit int) string {
	if value == nil {
		return ""
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return traceActivityTruncate(fmt.Sprintf("%v", value), limit)
	}
	return traceActivityTruncate(string(raw), limit)
}

func traceActivityTruncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if limit <= 0 || len(runes) <= limit {
		return value
	}
	if limit <= 1 {
		return "…"
	}
	return string(runes[:limit-1]) + "…"
}
