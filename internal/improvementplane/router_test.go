package improvementplane

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func seedRouterActionIntent(t *testing.T, store *storepkg.MemoryStore, intent action.Intent, prefix string, resultPayload map[string]any) action.Intent {
	t.Helper()
	now := intent.CreatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if intent.ID == "" {
		intent.ID = prefix + "-action"
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: intent.ID,
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   prefix + "-queue",
		OccurredAt:  now,
		Payload: map[string]any{
			"conversation_id": intent.ConversationID,
			"case_id":         intent.CaseID,
			"trace_id":        intent.TraceID,
			"kind":            string(intent.Kind),
			"target_ref":      intent.TargetRef,
			"request_payload": intent.RequestPayload,
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_queued) error = %v", err)
	}
	if resultPayload != nil {
		startedAt := now
		completedAt := now.Add(time.Second)
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionStart),
			CommandID:   prefix + "-start",
			OccurredAt:  startedAt,
			Payload:     map[string]any{"operation_id": prefix + "-op"},
		}); err != nil {
			t.Fatalf("SubmitCommand(action_started) error = %v", err)
		}
		payload := map[string]any{
			"operation_id": prefix + "-op",
			"started_at":   startedAt,
			"completed_at": completedAt,
		}
		for key, value := range resultPayload {
			payload[key] = value
		}
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionSucceed),
			CommandID:   prefix + "-succeed",
			OccurredAt:  completedAt,
			Payload:     payload,
		}); err != nil {
			t.Fatalf("SubmitCommand(action_succeeded) error = %v", err)
		}
	}
	created, ok := store.GetActionIntent(intent.ID)
	if !ok {
		t.Fatalf("expected action intent %s", intent.ID)
	}
	return created
}

func seedRouterKnowledgeEntry(t *testing.T, store *storepkg.MemoryStore, entryID string, entry knowledge.Entry, links []knowledge.EvidenceLink, commandID string) knowledge.Entry {
	t.Helper()
	occurredAt := entry.UpdatedAt
	if occurredAt.IsZero() {
		occurredAt = entry.CreatedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	createdAt := entry.CreatedAt
	if createdAt.IsZero() {
		createdAt = occurredAt
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineKnowledge,
		AggregateID: entryID,
		CommandKind: string(transition.CommandKnowledgeRecordDraft),
		CommandID:   commandID,
		Actor:       "tester",
		OccurredAt:  occurredAt,
		Payload: map[string]any{
			"tier":           string(entry.Tier),
			"kind":           string(entry.Kind),
			"scope_type":     string(entry.ScopeType),
			"scope_id":       entry.ScopeID,
			"title":          entry.Title,
			"summary":        entry.Summary,
			"body":           entry.Body,
			"status":         string(entry.Status),
			"confidence":     entry.Confidence,
			"source_type":    string(entry.SourceType),
			"created_at":     createdAt,
			"updated_at":     occurredAt,
			"evidence_links": links,
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(knowledge_record_draft) error = %v", err)
	}
	created, ok := store.GetKnowledgeEntry(entryID)
	if !ok {
		t.Fatalf("expected knowledge entry %s", entryID)
	}
	return created
}

func seedRouterRepoChangeJobViaCommand(t *testing.T, store *storepkg.MemoryStore, proposal review.Proposal, commandPrefix string, jobID string, branchName string) {
	t.Helper()
	if strings.TrimSpace(proposal.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt for proposal %s", proposal.ID)
	}
	occurredAt := time.Now().UTC()
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposal.ID,
		CommandKind: string(transition.CommandProposalMarkRepoChangeQueued),
		CommandID:   commandPrefix + "-proposal-queued",
		Actor:       "tester",
		OccurredAt:  occurredAt,
	}); err != nil {
		t.Fatalf("SubmitCommand(proposal_mark_repo_change_queued) error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAttempt,
		AggregateID: proposal.CurrentAttemptID,
		CommandKind: string(transition.CommandWorkspaceReady),
		CommandID:   commandPrefix + "-workspace-ready",
		Actor:       "tester",
		OccurredAt:  occurredAt.Add(time.Millisecond),
		Payload: map[string]any{
			"workspace_id":       "workspace-" + proposal.CurrentAttemptID,
			"job_id":             jobID,
			"repo":               "rsi-agent-platform",
			"base_ref":           "main",
			"branch_name":        branchName,
			"sandbox_namespace":  "rsi-platform",
			"sandbox_job_name":   "workspace-job-" + proposal.CurrentAttemptID,
			"sandbox_pod_name":   "workspace-pod-" + proposal.CurrentAttemptID,
			"allowed_path_globs": []string{"internal/**"},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workspace_ready) error = %v", err)
	}
}

func TestRouterConversationCaseAndTraceEndpoints(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceSummaries := store.ListTraces()
	if len(traceSummaries) == 0 {
		t.Fatal("expected seeded trace")
	}
	trace, ok := store.GetTrace(traceSummaries[0].TraceID)
	if !ok {
		t.Fatal("expected seeded trace detail")
	}
	now := time.Now().UTC()
	seedRouterActionIntent(t, store, action.Intent{
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        trace.Summary.TraceID,
		Kind:           action.KindSlackPost,
		CreatedAt:      now,
	}, "router-conversation", map[string]any{
		"executor":     "native-hermes-required",
		"provider":     "slack",
		"provider_ref": "slack-provider-ref",
	})
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: trace.Summary.TraceID,
		CommandKind: string(transition.CommandProblemLineRecordOutcome),
		CommandID:   "router-conversation-outcome",
		Actor:       "tester",
		OccurredAt:  now,
		Payload: map[string]any{
			"source":          "operator",
			"recorded_by":     "tester",
			"conversation_id": trace.Summary.ConversationID,
			"case_id":         trace.Summary.CaseID,
			"trace_id":        trace.Summary.TraceID,
			"outcome_type":    string(outcome.TypeAnswerQuality),
			"verdict":         string(outcome.VerdictPositive),
			"score":           1,
			"summary":         "Trace resolved the question well.",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(problem_line_record_outcome) error = %v", err)
	}
	entry := seedRouterKnowledgeEntry(t, store, "knowledge-trace-learning", knowledge.Entry{
		Tier:       knowledge.TierWorking,
		Kind:       knowledge.KindFact,
		ScopeType:  knowledge.ScopeCase,
		ScopeID:    trace.Summary.CaseID,
		Title:      "Trace learning",
		Summary:    "Structured action/outcome evidence is present.",
		Status:     knowledge.StatusDraft,
		Confidence: 0.8,
		SourceType: knowledge.SourceAgent,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, []knowledge.EvidenceLink{
		{
			EvidenceType: "trace",
			EvidenceID:   trace.Summary.TraceID,
			EvidenceRef:  events.EvidenceRef{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
		},
	}, "router-conversation-knowledge")
	if _, err := store.RecordHarnessExecutionObservation(harness.ExecutionObservation{
		ExecutionID: "hexec-router-trace",
		OperationID: "op-router-trace",
		TraceID:     trace.Summary.TraceID,
		WorkflowID:  "workflow-router-trace",
		Role:        "prod",
		Phase:       "investigate",
		EventType:   "executor.subprocess.output",
		Status:      "streaming",
		Seq:         1,
		RecordedAt:  now,
		Payload: map[string]any{
			"engine":         "hermes_aiagent_subprocess",
			"workspace_root": "/workspace/router-trace",
			"stream":         "stderr",
			"chunk_text":     "executor output",
			"chunk_index":    0,
		},
	}); err != nil {
		t.Fatalf("RecordHarnessExecutionObservation() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	listReq := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("conversation list status = %d, want %d", listRec.Code, http.StatusOK)
	}

	var listPayload struct {
		Conversations []map[string]any `json:"conversations"`
	}
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode conversation list: %v", err)
	}
	if len(listPayload.Conversations) == 0 {
		t.Fatal("expected at least one conversation in summary list")
	}
	if _, ok := listPayload.Conversations[0]["conversation_id"]; !ok {
		t.Fatal("expected conversation list item to include conversation_id")
	}
	if _, ok := listPayload.Conversations[0]["trace_attempts"]; ok {
		t.Fatal("conversation list should not include transcript or trace detail payloads")
	}

	conversationID, _ := listPayload.Conversations[0]["conversation_id"].(string)
	if conversationID == "" {
		t.Fatal("expected non-empty conversation_id from conversation list")
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/conversations/"+conversationID, nil)
	detailRec := httptest.NewRecorder()
	router.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("conversation detail status = %d, want %d", detailRec.Code, http.StatusOK)
	}

	var detailPayload map[string]any
	if err := json.NewDecoder(detailRec.Body).Decode(&detailPayload); err != nil {
		t.Fatalf("decode conversation detail: %v", err)
	}
	if _, ok := detailPayload["conversation"].(map[string]any); !ok {
		t.Fatal("expected conversation detail payload to include conversation object")
	}
	if _, ok := detailPayload["workflow_line"].(map[string]any); !ok {
		t.Fatal("expected conversation detail payload to include workflow_line")
	}
	workflowAttempts, ok := detailPayload["workflow_attempts"].([]any)
	if !ok || len(workflowAttempts) == 0 {
		t.Fatal("expected workflow_attempts to be a non-empty JSON array")
	}
	traceAttempts, ok := detailPayload["trace_attempts"].([]any)
	if !ok || len(traceAttempts) == 0 {
		t.Fatal("expected trace_attempts to be a non-empty JSON array")
	}
	if _, ok := detailPayload["transcript"].([]any); !ok {
		t.Fatal("expected transcript to be a JSON array")
	}
	if _, ok := detailPayload["linked_proposals"].([]any); !ok {
		t.Fatal("expected linked_proposals to be a JSON array")
	}

	pagedReq := httptest.NewRequest(http.MethodGet, "/api/conversations/"+conversationID+"?include=transcript&transcript_limit=1", nil)
	pagedRec := httptest.NewRecorder()
	router.ServeHTTP(pagedRec, pagedReq)
	if pagedRec.Code != http.StatusOK {
		t.Fatalf("paged conversation detail status = %d, want %d", pagedRec.Code, http.StatusOK)
	}
	var pagedPayload map[string]any
	if err := json.NewDecoder(pagedRec.Body).Decode(&pagedPayload); err != nil {
		t.Fatalf("decode paged conversation detail: %v", err)
	}
	transcript, ok := pagedPayload["transcript"].([]any)
	if !ok || len(transcript) > 1 {
		t.Fatalf("expected bounded transcript payload, got %#v", pagedPayload["transcript"])
	}
	if traces, ok := pagedPayload["trace_attempts"].([]any); !ok || len(traces) != 0 {
		t.Fatalf("expected trace attempts to be omitted by include filter, got %#v", pagedPayload["trace_attempts"])
	}

	caseReq := httptest.NewRequest(http.MethodGet, "/api/cases", nil)
	caseRec := httptest.NewRecorder()
	router.ServeHTTP(caseRec, caseReq)
	if caseRec.Code != http.StatusOK {
		t.Fatalf("case list status = %d, want %d", caseRec.Code, http.StatusOK)
	}

	var casePayload struct {
		Cases []map[string]any `json:"cases"`
	}
	if err := json.NewDecoder(caseRec.Body).Decode(&casePayload); err != nil {
		t.Fatalf("decode case list: %v", err)
	}
	if len(casePayload.Cases) == 0 {
		t.Fatal("expected at least one case")
	}
	caseSummary := casePayload.Cases[0]
	caseID, _ := caseSummary["case_id"].(string)
	if caseID == "" {
		t.Fatal("expected non-empty case_id from case list")
	}

	caseDetailReq := httptest.NewRequest(http.MethodGet, "/api/cases/"+caseID, nil)
	caseDetailRec := httptest.NewRecorder()
	router.ServeHTTP(caseDetailRec, caseDetailReq)
	if caseDetailRec.Code != http.StatusOK {
		t.Fatalf("case detail status = %d, want %d", caseDetailRec.Code, http.StatusOK)
	}
	var caseDetailPayload map[string]any
	if err := json.NewDecoder(caseDetailRec.Body).Decode(&caseDetailPayload); err != nil {
		t.Fatalf("decode case detail: %v", err)
	}
	if _, ok := caseDetailPayload["workflow_line"].(map[string]any); !ok {
		t.Fatal("expected case detail payload to include workflow_line")
	}
	if items, ok := caseDetailPayload["workflow_attempts"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected case detail payload to include workflow_attempts")
	}

	traceSummary, _ := traceAttempts[0].(map[string]any)
	traceID, _ := traceSummary["trace_id"].(string)
	if traceID == "" {
		t.Fatal("expected trace_id in trace attempts")
	}
	if runtimeSummary, ok := traceSummary["runtime_summary"].(map[string]any); !ok || runtimeSummary["runtime_source"] != "hermes-executor" {
		t.Fatal("expected trace attempt summary to include executor runtime summary")
	}
	traceReq := httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID, nil)
	traceRec := httptest.NewRecorder()
	router.ServeHTTP(traceRec, traceReq)
	if traceRec.Code != http.StatusOK {
		t.Fatalf("trace detail status = %d, want %d", traceRec.Code, http.StatusOK)
	}

	var tracePayload map[string]any
	if err := json.NewDecoder(traceRec.Body).Decode(&tracePayload); err != nil {
		t.Fatalf("decode trace detail: %v", err)
	}
	if _, ok := tracePayload["trace"].(map[string]any); !ok {
		t.Fatal("expected trace detail payload to include trace object")
	}
	if _, ok := tracePayload["workflow_line"].(map[string]any); !ok {
		t.Fatal("expected trace detail payload to include workflow_line")
	}
	if items, ok := tracePayload["workflow_attempts"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected trace detail payload to include workflow_attempts")
	}
	if _, ok := tracePayload["transcript_slice"].([]any); !ok {
		t.Fatal("expected transcript_slice to be a JSON array")
	}
	if _, ok := tracePayload["feedback_records"].([]any); !ok {
		t.Fatal("expected feedback_records to be a JSON array")
	}
	if _, ok := tracePayload["linked_proposals"].([]any); !ok {
		t.Fatal("expected linked_proposals to be a JSON array")
	}
	if _, ok := tracePayload["judgments_by_eval_run"].(map[string]any); !ok {
		t.Fatal("expected judgments_by_eval_run to be a JSON object")
	}
	if items, ok := tracePayload["action_intents"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected action_intents to be a non-empty JSON array")
	}
	if items, ok := tracePayload["action_results"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected action_results to be a non-empty JSON array")
	}
	if items, ok := tracePayload["outcomes"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected outcomes to be a non-empty JSON array")
	}
	if items, ok := tracePayload["knowledge_entries"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected knowledge_entries to be a non-empty JSON array")
	}
	if items, ok := tracePayload["harness_execution_observations"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected harness_execution_observations to be a non-empty JSON array")
	}
	if runtimeSummary, ok := tracePayload["runtime_summary"].(map[string]any); !ok || runtimeSummary["runtime_source"] != "hermes-executor" {
		t.Fatal("expected trace detail payload to include executor runtime summary")
	}

	workflowAttemptSummary, _ := workflowAttempts[0].(map[string]any)
	workflowID, _ := workflowAttemptSummary["workflow_id"].(string)
	if workflowID == "" {
		t.Fatal("expected workflow_id in workflow_attempts")
	}
	workflowAttemptReq := httptest.NewRequest(http.MethodGet, "/api/workflow-attempts/"+workflowID, nil)
	workflowAttemptRec := httptest.NewRecorder()
	router.ServeHTTP(workflowAttemptRec, workflowAttemptReq)
	if workflowAttemptRec.Code != http.StatusOK {
		t.Fatalf("workflow attempt detail status = %d, want %d", workflowAttemptRec.Code, http.StatusOK)
	}
	var workflowAttemptPayload map[string]any
	if err := json.NewDecoder(workflowAttemptRec.Body).Decode(&workflowAttemptPayload); err != nil {
		t.Fatalf("decode workflow attempt detail: %v", err)
	}
	if _, ok := workflowAttemptPayload["workflow_attempt"].(map[string]any); !ok {
		t.Fatal("expected workflow attempt detail payload to include workflow_attempt")
	}
	if _, ok := workflowAttemptPayload["workflow_line"].(map[string]any); !ok {
		t.Fatal("expected workflow attempt detail payload to include workflow_line")
	}
	if items, ok := workflowAttemptPayload["workflow_attempts"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected workflow attempt detail payload to include workflow_attempts")
	}

	actionReq := httptest.NewRequest(http.MethodGet, "/api/actions?trace="+traceID, nil)
	actionRec := httptest.NewRecorder()
	router.ServeHTTP(actionRec, actionReq)
	if actionRec.Code != http.StatusOK {
		t.Fatalf("action list status = %d, want %d", actionRec.Code, http.StatusOK)
	}

	outcomeReq := httptest.NewRequest(http.MethodGet, "/api/outcomes?trace="+traceID, nil)
	outcomeRec := httptest.NewRecorder()
	router.ServeHTTP(outcomeRec, outcomeReq)
	if outcomeRec.Code != http.StatusOK {
		t.Fatalf("outcome list status = %d, want %d", outcomeRec.Code, http.StatusOK)
	}

	knowledgeReq := httptest.NewRequest(http.MethodGet, "/api/knowledge", nil)
	knowledgeRec := httptest.NewRecorder()
	router.ServeHTTP(knowledgeRec, knowledgeReq)
	if knowledgeRec.Code != http.StatusOK {
		t.Fatalf("knowledge list status = %d, want %d", knowledgeRec.Code, http.StatusOK)
	}

	knowledgeDetailReq := httptest.NewRequest(http.MethodGet, "/api/knowledge/"+entry.ID, nil)
	knowledgeDetailRec := httptest.NewRecorder()
	router.ServeHTTP(knowledgeDetailRec, knowledgeDetailReq)
	if knowledgeDetailRec.Code != http.StatusOK {
		t.Fatalf("knowledge detail status = %d, want %d", knowledgeDetailRec.Code, http.StatusOK)
	}
}

func TestTraceStreamBackfillsLedgerEventsAndFiltersScope(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ExecutionID: "hexec-main",
			TraceID:     traceID,
			WorkflowID:  "workflow-main",
			PhaseID:     "investigate",
			Kind:        "model.output.delta",
			Status:      "streaming",
			Seq:         1,
			Payload:     map[string]any{"role": "prod", "delta": "hello"},
			RecordedAt:  now,
		},
		{
			ExecutionID: "hexec-eval",
			TraceID:     traceID,
			WorkflowID:  "workflow-eval",
			PhaseID:     "eval",
			Kind:        "model.output.delta",
			Status:      "streaming",
			Seq:         1,
			Payload:     map[string]any{"role": "eval", "delta": "eval"},
			RecordedAt:  now.Add(time.Millisecond),
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID+"/stream?scope=main", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("stream status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: ledger") || !strings.Contains(body, "hello") {
		t.Fatalf("expected backfilled ledger event, got %q", body)
	}
	if strings.Contains(body, "eval") {
		t.Fatalf("expected scope filter to hide eval event, got %q", body)
	}
}

func TestTraceActivityRouteProjectsCleanScopedActivity(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ExecutionID: "hexec-main",
			TraceID:     traceID,
			WorkflowID:  "workflow-main",
			PhaseID:     "investigate",
			Kind:        "tool.call.completed",
			Status:      "completed",
			Seq:         1,
			Payload: map[string]any{
				"tool_name": "terminal",
				"args":      map[string]any{"command": "echo sensitive"},
				"result":    `{"status":"ok","exit_code":0,"output":"done"}`,
			},
			RecordedAt: now,
		},
		{
			ExecutionID: "hexec-eval",
			TraceID:     traceID,
			WorkflowID:  "workflow-eval",
			PhaseID:     "eval",
			Kind:        "tool.call.completed",
			Status:      "completed",
			Seq:         1,
			Payload: map[string]any{
				"role":      "eval",
				"tool_name": "terminal",
				"result":    `{"status":"ok","exit_code":0}`,
			},
			RecordedAt: now.Add(time.Millisecond),
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	req := httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID+"/activity?scope=main&mode=clean", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("activity status = %d body=%s", rec.Code, rec.Body.String())
	}
	var snapshot TraceActivitySnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode activity snapshot: %v", err)
	}
	if snapshot.Metrics.LedgerEventCount != 1 {
		t.Fatalf("ledger event count = %d, want 1", snapshot.Metrics.LedgerEventCount)
	}
	if len(snapshot.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(snapshot.Items))
	}
	if got := snapshot.Items[0].ToolName; got != "terminal" {
		t.Fatalf("tool name = %q, want terminal", got)
	}
	if _, ok := snapshot.Items[0].Details["args_excerpt"]; ok {
		t.Fatalf("clean activity unexpectedly included args excerpt: %#v", snapshot.Items[0].Details)
	}
}

func TestCompanyWikiDashboardRoutesReadIndexAndFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "pages", "runbooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.md"), []byte("# Company Wiki Index\n\n- [Deploy](pages/runbooks/deploy.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pages", "runbooks", "deploy.md"), []byte("# Deploy\n\n- citation: [Slack source](https://storyprotocol.slack.com/archives/C123/p1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(config.Config{CompanyWikiRoot: root}, storepkg.NewMemoryStore())

	for _, tc := range []struct {
		path string
		want string
	}{
		{path: "/api/company-wiki/index", want: "Company Wiki Index"},
		{path: "/api/company-wiki/file?path=pages%2Frunbooks%2Fdeploy.md", want: "Slack source"},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", tc.path, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), tc.want) {
			t.Fatalf("%s missing %q in %s", tc.path, tc.want, rec.Body.String())
		}
	}
}

func TestCompanyWikiDashboardFileFallsBackToStoreWhenFileMissing(t *testing.T) {
	root := t.TempDir()
	store := storepkg.NewMemoryStore()
	source, err := store.UpsertCompanyWikiSourceRevision(storepkg.CompanyWikiSourceRevisionInput{
		SourceType:        "notion_document",
		DocumentSourceKey: "notion:integrate-poseidon-processing-pipeline-with-numo",
		SourceRevision:    "1",
		Title:             "Integrate Poseidon Processing Pipeline with Numo",
		Content:           "Project to integrate the Poseidon Processing Pipeline with Numo.",
		NativeLocator:     "https://www.notion.so/Integrate-Poseidon-Processing-Pipeline-with-Numo",
	})
	if err != nil {
		t.Fatalf("UpsertCompanyWikiSourceRevision() error = %v", err)
	}
	body := "# Integrate Poseidon Processing Pipeline with Numo\n\nProject to integrate the Poseidon Processing Pipeline with Numo.\n"
	if _, err := store.PublishCompanyWikiPage(storepkg.CompanyWikiPagePublishInput{
		Slug:        "projects/integrate-poseidon-processing-pipeline-with-numo",
		Title:       "Integrate Poseidon Processing Pipeline with Numo",
		Body:        body,
		Path:        "pages/projects/integrate-poseidon-processing-pipeline-with-numo.md",
		SHA256:      storepkg.CompanyWikiSHA256(body),
		PublishedAt: time.Now().UTC(),
		Citations: []storepkg.CompanyWikiCitationInput{{
			ClaimKey:         "project-summary",
			SourceDocumentID: source.Document.ID,
			SourceRevisionID: source.Revision.ID,
			ChunkID:          source.Chunks[0].ID,
			NativeLocator:    source.Chunks[0].NativeLocator,
		}},
	}); err != nil {
		t.Fatalf("PublishCompanyWikiPage() error = %v", err)
	}
	router := NewRouter(config.Config{CompanyWikiRoot: root}, store)
	req := httptest.NewRequest(http.MethodGet, "/api/company-wiki/file?path=pages%2Fprojects%2Fintegrate-poseidon-processing-pipeline-with-numo.md", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("file fallback status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Poseidon Processing Pipeline with Numo") {
		t.Fatalf("fallback body missing page content: %s", rec.Body.String())
	}
}

func TestKanbanDashboardActorFromCloudflareAccessIdentity(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/kanban/projects", nil)
	req.Header.Set("Cf-Access-Authenticated-User-Email", "blake.huynh@piplabs.xyz")
	actor, ok := kanbanDashboardActorFromHeaders(req, "dashboard")
	if !ok {
		t.Fatal("Cloudflare Access piplabs email was not resolved")
	}
	if actor.ID != "blake.huynh@piplabs.xyz" || actor.Display != "Blake Huynh" || actor.Surface != "dashboard" {
		t.Fatalf("actor = %+v, want piplabs email display attribution", actor)
	}

	jwt := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." +
		base64.RawURLEncoding.EncodeToString([]byte(`{"email":"ada.lovelace@piplabs.xyz","name":"Ada Lovelace"}`)) + "."
	req = httptest.NewRequest(http.MethodGet, "/api/kanban/projects", nil)
	req.Header.Set("Cf-Access-Jwt-Assertion", jwt)
	actor, ok = kanbanDashboardActorFromHeaders(req, "dashboard")
	if !ok {
		t.Fatal("Cloudflare Access JWT piplabs email was not resolved")
	}
	if actor.ID != "ada.lovelace@piplabs.xyz" || actor.Display != "Ada Lovelace" {
		t.Fatalf("JWT actor = %+v, want Ada Lovelace", actor)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kanban/projects", nil)
	req.Header.Set("Cf-Access-Authenticated-User-Email", "mallory@example.com")
	if actor, ok := kanbanDashboardActorFromHeaders(req, "dashboard"); ok {
		t.Fatalf("non-company Access email resolved as actor: %+v", actor)
	}
}

func TestHermesCompatibilityEndpointsExposePlatformState(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{
		ServiceName:              "improvement-plane",
		PublicBaseURL:            "http://example.test",
		DefaultRepo:              "depin-backend",
		ProposalPromoterInterval: 15 * time.Minute,
	}, store)

	cases := []struct {
		path string
		want string
	}{
		{path: "/api/status", want: "gateway_platforms"},
		{path: "/api/skills", want: "repo.answer_question"},
		{path: "/api/tools/toolsets", want: "rsi-tool-gateway"},
		{path: "/api/cron/jobs", want: "proposal-promoter"},
		{path: "/api/dashboard/themes", want: "midnight"},
		{path: "/api/dashboard/plugins", want: "[]"},
		{path: "/api/sessions", want: "sessions"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", tc.path, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), tc.want) {
			t.Fatalf("%s missing %q in %s", tc.path, tc.want, rec.Body.String())
		}
	}
}

func TestKanbanRoutesCreateProjectTicketAndKeepHermesPluginDisabled(t *testing.T) {
	state := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "improvement-plane"}, state)

	projectReq := httptest.NewRequest(http.MethodPost, "/api/kanban/projects", strings.NewReader(`{"name":"Project Data Audit","slug":"project-data-audit"}`))
	projectRec := httptest.NewRecorder()
	router.ServeHTTP(projectRec, projectReq)
	if projectRec.Code != http.StatusCreated {
		t.Fatalf("create project status=%d body=%s", projectRec.Code, projectRec.Body.String())
	}
	var projectPayload struct {
		Project storepkg.KanbanProject `json:"project"`
	}
	if err := json.Unmarshal(projectRec.Body.Bytes(), &projectPayload); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	routeReq := httptest.NewRequest(http.MethodPost, "/api/kanban/projects/"+projectPayload.Project.ID+"/slack-routes", strings.NewReader(`{"team_id":"T123","channel_id":"C123"}`))
	routeRec := httptest.NewRecorder()
	router.ServeHTTP(routeRec, routeReq)
	if routeRec.Code != http.StatusCreated {
		t.Fatalf("create slack route status=%d body=%s", routeRec.Code, routeRec.Body.String())
	}

	ticketReq := httptest.NewRequest(http.MethodPost, "/api/kanban/tickets", strings.NewReader(`{"project_id":"`+projectPayload.Project.ID+`","title":"Audit Slack ticket creation","description":"Track the V1 workflow."}`))
	ticketRec := httptest.NewRecorder()
	router.ServeHTTP(ticketRec, ticketReq)
	if ticketRec.Code != http.StatusCreated {
		t.Fatalf("create ticket status=%d body=%s", ticketRec.Code, ticketRec.Body.String())
	}
	var ticketPayload struct {
		Ticket storepkg.KanbanTicket `json:"ticket"`
	}
	if err := json.Unmarshal(ticketRec.Body.Bytes(), &ticketPayload); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}

	commentReq := httptest.NewRequest(http.MethodPost, "/api/kanban/tickets/"+ticketPayload.Ticket.ID+"/comments", strings.NewReader(`{"body":"Dashboard follow-up"}`))
	commentReq.Header.Set("Cf-Access-Authenticated-User-Email", "blake.huynh@piplabs.xyz")
	commentRec := httptest.NewRecorder()
	router.ServeHTTP(commentRec, commentReq)
	if commentRec.Code != http.StatusCreated {
		t.Fatalf("create comment status=%d body=%s", commentRec.Code, commentRec.Body.String())
	}
	var commentPayload struct {
		Comment storepkg.KanbanTicketComment `json:"comment"`
	}
	if err := json.Unmarshal(commentRec.Body.Bytes(), &commentPayload); err != nil {
		t.Fatalf("decode comment: %v", err)
	}
	if commentPayload.Comment.ActorID != "blake.huynh@piplabs.xyz" {
		t.Fatalf("comment actor_id = %q, want piplabs email", commentPayload.Comment.ActorID)
	}
	if commentPayload.Comment.ActorDisplay != "Blake Huynh" {
		t.Fatalf("comment actor_display = %q, want Blake Huynh", commentPayload.Comment.ActorDisplay)
	}
	if commentPayload.Comment.SourceSurface != "dashboard" {
		t.Fatalf("comment source_surface = %q, want dashboard", commentPayload.Comment.SourceSurface)
	}

	statusReq := httptest.NewRequest(http.MethodPost, "/api/kanban/tickets", strings.NewReader(`{"project_id":"`+projectPayload.Project.ID+`","title":"Bypass triage","status":"todo"}`))
	statusRec := httptest.NewRecorder()
	router.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusBadRequest {
		t.Fatalf("create ticket with status = %d, want 400 body=%s", statusRec.Code, statusRec.Body.String())
	}
	if !strings.Contains(statusRec.Body.String(), "start in triage") {
		t.Fatalf("create ticket with status error missing triage guidance: %s", statusRec.Body.String())
	}

	boardReq := httptest.NewRequest(http.MethodGet, "/api/kanban/projects/"+projectPayload.Project.ID+"/board", nil)
	boardRec := httptest.NewRecorder()
	router.ServeHTTP(boardRec, boardReq)
	if boardRec.Code != http.StatusOK || !strings.Contains(boardRec.Body.String(), "Audit Slack ticket creation") {
		t.Fatalf("board response status=%d body=%s", boardRec.Code, boardRec.Body.String())
	}
	latestEventID := state.LatestKanbanEventID(projectPayload.Project.ID)
	if got := kanbanStreamStartCursor(state, projectPayload.Project.ID, "", false); got != latestEventID {
		t.Fatalf("empty stream cursor = %q, want latest %q", got, latestEventID)
	}
	if got := kanbanStreamStartCursor(state, projectPayload.Project.ID, "missing-event", true); got != "" {
		t.Fatalf("explicit stale stream cursor = %q, want backfill from start", got)
	}

	pluginReq := httptest.NewRequest(http.MethodGet, "/api/plugins/kanban/board", nil)
	pluginRec := httptest.NewRecorder()
	router.ServeHTTP(pluginRec, pluginReq)
	if pluginRec.Code != http.StatusNotFound {
		t.Fatalf("Hermes kanban plugin route status=%d, want 404", pluginRec.Code)
	}
}

func TestHermesCompatibilitySkillsIncludeExportedHermesCatalog(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	req := httptest.NewRequest(http.MethodGet, "/api/skills", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("skills status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"repo.answer_question", "rsi-platform-investigation", "depin-prod-admin-read"} {
		if !strings.Contains(body, want) {
			t.Fatalf("skills response missing %q in %s", want, body)
		}
	}
}

func TestHermesCompatibilitySkillDetailIncludesMarkdown(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	req := httptest.NewRequest(http.MethodGet, "/api/skills/rsi-platform-investigation", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("skill detail status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("skill detail response is not JSON: %v", err)
	}
	if got := payload["name"]; got != "rsi-platform-investigation" {
		t.Fatalf("name = %v, want rsi-platform-investigation", got)
	}
	content, _ := payload["content"].(string)
	if !strings.Contains(content, "# RSI Platform Investigation") {
		t.Fatalf("skill detail content missing markdown body: %s", content)
	}
	if got := payload["path"]; got != "story-company/rsi-platform-investigation/SKILL.md" {
		t.Fatalf("path = %v, want story-company/rsi-platform-investigation/SKILL.md", got)
	}
}

func TestHermesCompatibilitySessionsExposeGroupingMetadata(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	trace := store.ListTraces()[0]
	conversationID := trace.ConversationID
	if conversationID == "" {
		conversationID = store.ListConversations()[0].ID
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=200", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("sessions status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Sessions []struct {
			ID              string `json:"id"`
			Type            string `json:"type"`
			ConversationID  string `json:"conversation_id"`
			TraceID         string `json:"trace_id"`
			ParentSessionID string `json:"parent_session_id"`
			ThreadKey       string `json:"thread_key"`
			WorkflowKind    string `json:"workflow_kind"`
			Status          string `json:"status"`
			TraceCount      int    `json:"trace_count"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}

	var conversationSession, traceSession *struct {
		ID              string `json:"id"`
		Type            string `json:"type"`
		ConversationID  string `json:"conversation_id"`
		TraceID         string `json:"trace_id"`
		ParentSessionID string `json:"parent_session_id"`
		ThreadKey       string `json:"thread_key"`
		WorkflowKind    string `json:"workflow_kind"`
		Status          string `json:"status"`
		TraceCount      int    `json:"trace_count"`
	}
	for i := range payload.Sessions {
		session := &payload.Sessions[i]
		if session.ID == conversationID {
			conversationSession = session
		}
		if session.ID == trace.TraceID {
			traceSession = session
		}
	}
	if conversationSession == nil {
		t.Fatalf("missing conversation session %s in %+v", conversationID, payload.Sessions)
	}
	if traceSession == nil {
		t.Fatalf("missing trace session %s in %+v", trace.TraceID, payload.Sessions)
	}
	if conversationSession.Type != "conversation" || conversationSession.ConversationID != conversationID {
		t.Fatalf("conversation metadata = %+v, want conversation_id %s", *conversationSession, conversationID)
	}
	if conversationSession.TraceCount == 0 {
		t.Fatalf("conversation trace_count = 0, want related trace count")
	}
	if traceSession.Type != "trace" ||
		traceSession.TraceID != trace.TraceID ||
		traceSession.ConversationID != conversationID ||
		traceSession.ParentSessionID != conversationID {
		t.Fatalf("trace grouping metadata = %+v, want parent %s", *traceSession, conversationID)
	}
	if traceSession.ThreadKey == "" || traceSession.WorkflowKind == "" || traceSession.Status == "" {
		t.Fatalf("trace operational metadata incomplete: %+v", *traceSession)
	}
}

type pagedHermesSessionStore struct {
	*storepkg.MemoryStore
	page       storepkg.HermesSessionListPage
	calls      int
	lastLimit  int
	lastOffset int
}

func (s *pagedHermesSessionStore) ListHermesSessionsPage(limit int, offset int) storepkg.HermesSessionListPage {
	s.calls++
	s.lastLimit = limit
	s.lastOffset = offset
	return s.page
}

func TestHermesCompatibilitySessionsUsesPagedLoader(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	store := &pagedHermesSessionStore{
		MemoryStore: storepkg.NewMemoryStore(),
		page: storepkg.HermesSessionListPage{
			Total: 123,
			Items: []storepkg.HermesSessionListItem{{
				ID:             "conv-hot",
				Type:           "conversation",
				RawTitle:       "Hot conversation",
				StartedAt:      now.Add(-time.Hour),
				LastActive:     now,
				ConversationID: "conv-hot",
				Status:         "open",
				MessageCount:   2,
				TraceCount:     3,
				ProposalCount:  1,
			}},
		},
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=7&offset=14", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("sessions status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.calls != 1 || store.lastLimit != 7 || store.lastOffset != 14 {
		t.Fatalf("paged loader calls=%d limit=%d offset=%d", store.calls, store.lastLimit, store.lastOffset)
	}
	var payload struct {
		Total    int `json:"total"`
		Limit    int `json:"limit"`
		Offset   int `json:"offset"`
		Sessions []struct {
			ID            string `json:"id"`
			Type          string `json:"type"`
			Title         string `json:"title"`
			MessageCount  int    `json:"message_count"`
			TraceCount    int    `json:"trace_count"`
			ProposalCount int    `json:"proposal_count"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	if payload.Total != 123 || payload.Limit != 7 || payload.Offset != 14 {
		t.Fatalf("pagination metadata = total:%d limit:%d offset:%d", payload.Total, payload.Limit, payload.Offset)
	}
	if len(payload.Sessions) != 1 || payload.Sessions[0].ID != "conv-hot" || payload.Sessions[0].Title != "Hot conversation" {
		t.Fatalf("sessions payload = %+v", payload.Sessions)
	}
	if payload.Sessions[0].MessageCount != 2 || payload.Sessions[0].TraceCount != 3 || payload.Sessions[0].ProposalCount != 1 {
		t.Fatalf("session counts = %+v", payload.Sessions[0])
	}
}

type titleBatchStore struct {
	*storepkg.MemoryStore
	steps          []events.ReasoningStep
	getTraceCalls  int
	reasoningCalls int
	eventIDCalls   int
}

func (s *titleBatchStore) GetTrace(traceID string) (events.Trace, bool) {
	s.getTraceCalls++
	return s.MemoryStore.GetTrace(traceID)
}

func (s *titleBatchStore) ListEventsByIDs(ids []string) []ingestion.EventEnvelope {
	s.eventIDCalls++
	return nil
}

func (s *titleBatchStore) ListReasoningStepsByTraceIDsAndType(traceIDs []string, stepType string) []events.ReasoningStep {
	s.reasoningCalls++
	return s.steps
}

func TestHermesCompatibilityTitleCacheUsesBatchedSessionTitles(t *testing.T) {
	base := storepkg.NewMemoryStore()
	summary := base.ListTraces()[0]
	store := &titleBatchStore{
		MemoryStore: base,
		steps: []events.ReasoningStep{{
			ID:        "reason-session-title-batched",
			TraceID:   summary.TraceID,
			StepType:  "session_title",
			Summary:   "Batched session title",
			CreatedAt: time.Now().UTC(),
		}},
	}

	cache := buildTitleCache(store, []events.TraceSummary{summary})
	got := traceTitleForDisplayWithCache(store, summary, cache)

	if got.Title != "Batched session title" || !got.IsSummary {
		t.Fatalf("title = %+v, want batched summary title", got)
	}
	if store.reasoningCalls != 1 || store.eventIDCalls != 1 {
		t.Fatalf("batch calls reasoning=%d events=%d", store.reasoningCalls, store.eventIDCalls)
	}
	if store.getTraceCalls != 0 {
		t.Fatalf("GetTrace calls = %d, want no per-trace fallback", store.getTraceCalls)
	}
}

func TestHermesCompatibilitySessionTitlesKeepThreadRootStableAcrossFollowUps(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	rootTS := "171010000.000100"
	threadKey := "slack:thread:CENG:" + rootTS
	rootText := "@RSI can you review these PRs and flag risks?"
	followUpText := "@RSI also think through the follow-up UX in this thread"

	submitSlack := func(commandID string, ts string, threadTS string, text string) string {
		t.Helper()
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineIngress,
			AggregateID: "slack:" + ts,
			CommandKind: string(transition.CommandIngressRecordSlack),
			CommandID:   commandID,
			Actor:       "tester",
			OccurredAt:  time.Now().UTC(),
			Payload: map[string]any{
				"team_id":    "TENG",
				"channel_id": "CENG",
				"user_id":    "U123",
				"thread_ts":  threadTS,
				"ts":         ts,
				"text":       text,
			},
		}); err != nil {
			t.Fatalf("SubmitCommand(slack ingress) error = %v", err)
		}
		for _, ingestion := range store.ListIngestions() {
			if ingestion.ThreadKey == threadKey && ingestion.ThreadTS == threadTS && ingestion.Text == text {
				return ingestion.EventID
			}
		}
		t.Fatalf("missing ingestion for %q", text)
		return ""
	}

	traceForEvent := func(eventID string) events.Trace {
		t.Helper()
		for _, summary := range store.ListTraces() {
			if summary.TriggerEventID == eventID {
				trace, ok := store.GetTrace(summary.TraceID)
				if !ok {
					t.Fatalf("missing trace %s", summary.TraceID)
				}
				return trace
			}
		}
		t.Fatalf("missing trace for event %s", eventID)
		return events.Trace{}
	}

	projectTitle := func(trace events.Trace, title string) {
		t.Helper()
		now := time.Now().UTC()
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineProblemLine,
			AggregateID: trace.Summary.TraceID,
			CommandKind: string(transition.CommandProblemLineProjectTrace),
			CommandID:   "cmd-project-title-" + trace.Summary.TraceID,
			Actor:       "tester",
			OccurredAt:  now,
			Payload: map[string]any{
				"trace_id": trace.Summary.TraceID,
				"reasoning_steps": []events.ReasoningStep{{
					ID:             "reason-session-title-" + trace.Summary.TraceID,
					TraceID:        trace.Summary.TraceID,
					WorkflowID:     trace.Summary.WorkflowID,
					ConversationID: trace.Summary.ConversationID,
					CaseID:         trace.Summary.CaseID,
					StepType:       "session_title",
					Summary:        title,
					CreatedAt:      now,
				}},
			},
		}); err != nil {
			t.Fatalf("SubmitCommand(project title) error = %v", err)
		}
	}

	rootEventID := submitSlack("cmd-slack-root-title", rootTS, rootTS, rootText)
	rootTrace := traceForEvent(rootEventID)
	projectTitle(rootTrace, "Review PR risks")
	followUpEventID := submitSlack("cmd-slack-follow-up-title", "171010005.000200", rootTS, followUpText)
	followUpTrace := traceForEvent(followUpEventID)
	projectTitle(followUpTrace, "Plan follow-up thread UX")

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=200", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("sessions status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Sessions []struct {
			ID              string `json:"id"`
			Type            string `json:"type"`
			Title           string `json:"title"`
			OriginalTitle   string `json:"original_title"`
			TitleIsSummary  bool   `json:"title_is_summary"`
			ParentSessionID string `json:"parent_session_id"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}

	var conversationSession, rootTraceSession, followUpTraceSession *struct {
		ID              string `json:"id"`
		Type            string `json:"type"`
		Title           string `json:"title"`
		OriginalTitle   string `json:"original_title"`
		TitleIsSummary  bool   `json:"title_is_summary"`
		ParentSessionID string `json:"parent_session_id"`
	}
	for i := range payload.Sessions {
		session := &payload.Sessions[i]
		switch session.ID {
		case rootTrace.Summary.ConversationID:
			conversationSession = session
		case rootTrace.Summary.TraceID:
			rootTraceSession = session
		case followUpTrace.Summary.TraceID:
			followUpTraceSession = session
		}
	}
	if conversationSession == nil || rootTraceSession == nil || followUpTraceSession == nil {
		t.Fatalf("missing expected sessions in %+v", payload.Sessions)
	}
	if conversationSession.Title != "Review PR risks" || conversationSession.OriginalTitle != rootText || !conversationSession.TitleIsSummary {
		t.Fatalf("conversation title metadata = %+v", *conversationSession)
	}
	if rootTraceSession.Title != "Review PR risks" || followUpTraceSession.Title != "Plan follow-up thread UX" {
		t.Fatalf("trace titles root=%+v follow_up=%+v", *rootTraceSession, *followUpTraceSession)
	}
	if followUpTraceSession.ParentSessionID != conversationSession.ID {
		t.Fatalf("follow-up trace parent = %q, want %q", followUpTraceSession.ParentSessionID, conversationSession.ID)
	}
}

func TestHermesCompatibilityLiveSessionsFollowRunnerExecutions(t *testing.T) {
	cfg := config.Config{
		PublicBaseURL:                   "http://example.test",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}
	store := storepkg.NewMemoryStore()
	router := NewRouter(cfg, store)
	trace := store.ListTraces()[0]
	traceID := trace.TraceID
	conversationID := trace.ConversationID
	if conversationID == "" {
		conversationID = store.ListConversations()[0].ID
	}

	assertActive := func(t *testing.T, want bool) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=200", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("sessions status = %d body=%s", rec.Code, rec.Body.String())
		}
		var payload struct {
			Sessions []struct {
				ID       string `json:"id"`
				IsActive bool   `json:"is_active"`
			} `json:"sessions"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode sessions: %v", err)
		}
		seen := map[string]bool{}
		for _, session := range payload.Sessions {
			if session.ID == traceID || session.ID == conversationID {
				seen[session.ID] = true
				if session.IsActive != want {
					t.Fatalf("session %s is_active=%v, want %v", session.ID, session.IsActive, want)
				}
			}
		}
		for _, id := range []string{traceID, conversationID} {
			if !seen[id] {
				t.Fatalf("missing session %s in %+v", id, payload.Sessions)
			}
		}
	}

	assertActive(t, false)

	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:    "hexec-live-ui",
		TraceID:        traceID,
		ConversationID: conversationID,
		Status:         "running",
		HeartbeatAt:    &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(live) error = %v", err)
	}
	assertActive(t, true)

	staleStore := storepkg.NewMemoryStore()
	staleRouter := NewRouter(cfg, staleStore)
	staleTrace := staleStore.ListTraces()[0]
	staleConversationID := staleTrace.ConversationID
	if staleConversationID == "" {
		staleConversationID = staleStore.ListConversations()[0].ID
	}
	stale := now.Add(-2 * time.Minute)
	if _, err := staleStore.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:    "hexec-stale-ui",
		TraceID:        staleTrace.TraceID,
		ConversationID: staleConversationID,
		Status:         "running",
		HeartbeatAt:    &stale,
		CreatedAt:      stale,
		UpdatedAt:      stale,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(stale) error = %v", err)
	}
	router = staleRouter
	traceID = staleTrace.TraceID
	conversationID = staleConversationID
	assertActive(t, false)
}

func TestHermesCompatibilityLiveSessionsFollowFreshLedgerActivity(t *testing.T) {
	cfg := config.Config{
		PublicBaseURL:                   "http://example.test",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}
	store := storepkg.NewMemoryStore()
	trace := store.ListTraces()[0]
	traceID := trace.TraceID
	conversationID := trace.ConversationID
	if conversationID == "" {
		conversationID = store.ListConversations()[0].ID
	}
	now := time.Now().UTC().Add(time.Second)
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-fresh-terminal-output",
			ExecutionID: "hexec-ledger-live-ui",
			TraceID:     traceID,
			WorkflowID:  trace.WorkflowID,
			Kind:        "terminal.output",
			Status:      "streaming",
			Seq:         1,
			Payload:     map[string]any{"tool_name": "terminal", "chunk_text": "still streaming"},
			RecordedAt:  now,
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents(fresh) error = %v", err)
	}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=200", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("sessions status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Sessions []struct {
			ID         string `json:"id"`
			IsActive   bool   `json:"is_active"`
			LastActive int64  `json:"last_active"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	for _, id := range []string{traceID, conversationID} {
		var found *struct {
			ID         string `json:"id"`
			IsActive   bool   `json:"is_active"`
			LastActive int64  `json:"last_active"`
		}
		for i := range payload.Sessions {
			if payload.Sessions[i].ID == id {
				found = &payload.Sessions[i]
				break
			}
		}
		if found == nil {
			t.Fatalf("missing session %s in %+v", id, payload.Sessions)
		}
		if !found.IsActive {
			t.Fatalf("session %s is_active=false, want true from fresh ledger activity", id)
		}
		if found.LastActive != now.Unix() {
			t.Fatalf("session %s last_active=%d, want %d", id, found.LastActive, now.Unix())
		}
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	statusRec := httptest.NewRecorder()
	router.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("status code = %d body=%s", statusRec.Code, statusRec.Body.String())
	}
	var statusPayload struct {
		ActiveSessions int `json:"active_sessions"`
	}
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusPayload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if statusPayload.ActiveSessions != 2 {
		t.Fatalf("active_sessions=%d, want trace plus parent conversation", statusPayload.ActiveSessions)
	}
}

func TestHermesCompatibilityShowsExternalToolWaitAsLiveSession(t *testing.T) {
	cfg := config.Config{
		PublicBaseURL:                   "http://example.test",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}
	store := storepkg.NewMemoryStore()
	workflow := store.ListWorkflows()[0]
	traceID := workflow.TraceID
	conversationID := workflow.ConversationID
	now := time.Now().UTC()
	commands := []transition.CommandEnvelope{
		{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandWorkflowStarted),
			CommandID:   "cmd-test-ext-wait-started",
			OccurredAt:  now,
		},
		{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandContextSkipped),
			CommandID:   "cmd-test-ext-wait-context-skipped",
			OccurredAt:  now.Add(time.Second),
		},
		{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandWorkflowWaitingExternalTool),
			CommandID:   "cmd-test-ext-wait-waiting",
			OccurredAt:  now.Add(2 * time.Second),
			Payload: map[string]any{
				"external_tool_pause_id": "extpause-test",
			},
		},
	}
	for _, command := range commands {
		if _, err := store.SubmitCommand(command); err != nil {
			t.Fatalf("SubmitCommand(%s) error = %v", command.CommandKind, err)
		}
	}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=200", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("sessions status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Sessions []struct {
			ID       string `json:"id"`
			IsActive bool   `json:"is_active"`
			Status   string `json:"status"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	for _, want := range []struct {
		id            string
		expectedTrace bool
	}{
		{id: traceID, expectedTrace: true},
		{id: conversationID},
	} {
		found := false
		for _, session := range payload.Sessions {
			if session.ID != want.id {
				continue
			}
			found = true
			if !session.IsActive {
				t.Fatalf("session %s is_active=false, want true while external tool is pending", want.id)
			}
			if want.expectedTrace && session.Status != string(transition.WorkflowStateWaitingExternalTool) {
				t.Fatalf("trace status = %q, want %q", session.Status, transition.WorkflowStateWaitingExternalTool)
			}
		}
		if !found {
			t.Fatalf("missing session %s in %+v", want.id, payload.Sessions)
		}
	}
}

type batchedHermesLiveLedgerStore struct {
	*storepkg.MemoryStore
	events       []storepkg.HermesLiveLedgerEvent
	batchCalls   int
	byTraceCalls int
}

func (s *batchedHermesLiveLedgerStore) ListRecentHermesLiveLedgerEvents(since time.Time) []storepkg.HermesLiveLedgerEvent {
	s.batchCalls++
	out := []storepkg.HermesLiveLedgerEvent{}
	for _, item := range s.events {
		if item.RecordedAt.Before(since) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func (s *batchedHermesLiveLedgerStore) ListExecutionLedgerEventsByTrace(traceID string) []events.ExecutionLedgerEvent {
	s.byTraceCalls++
	return s.MemoryStore.ListExecutionLedgerEventsByTrace(traceID)
}

func TestHermesCompatibilityLiveSessionsUseBatchedLedgerReader(t *testing.T) {
	cfg := config.Config{
		PublicBaseURL:                   "http://example.test",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}
	base := storepkg.NewMemoryStore()
	trace := base.ListTraces()[0]
	conversationID := trace.ConversationID
	if conversationID == "" {
		conversationID = base.ListConversations()[0].ID
	}
	now := time.Now().UTC()
	store := &batchedHermesLiveLedgerStore{
		MemoryStore: base,
		events: []storepkg.HermesLiveLedgerEvent{{
			TraceID:        trace.TraceID,
			ConversationID: conversationID,
			Kind:           "tool.progress",
			Status:         "running",
			RecordedAt:     now,
		}},
	}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		ActiveSessions int `json:"active_sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if payload.ActiveSessions != 2 {
		t.Fatalf("active_sessions=%d, want trace plus parent conversation", payload.ActiveSessions)
	}
	if store.batchCalls == 0 {
		t.Fatal("batched ledger reader was not called")
	}
	if store.byTraceCalls != 0 {
		t.Fatalf("per-trace ledger calls=%d, want 0 when batched reader is available", store.byTraceCalls)
	}
}

func TestHermesCompatibilityLiveSessionsTerminalSameTimestampWins(t *testing.T) {
	cfg := config.Config{
		PublicBaseURL:                   "http://example.test",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}
	store := storepkg.NewMemoryStore()
	trace := store.ListTraces()[0]
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-same-time-live",
			ExecutionID: "hexec-same-time",
			TraceID:     trace.TraceID,
			WorkflowID:  trace.WorkflowID,
			Kind:        "tool.progress",
			Status:      "running",
			Seq:         1,
			RecordedAt:  now,
		},
		{
			ID:          "xled-same-time-terminal",
			ExecutionID: "hexec-same-time",
			TraceID:     trace.TraceID,
			WorkflowID:  trace.WorkflowID,
			Kind:        "tool.complete",
			Status:      "completed",
			Seq:         2,
			RecordedAt:  now,
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		ActiveSessions int `json:"active_sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if payload.ActiveSessions != 0 {
		t.Fatalf("active_sessions=%d, want terminal event to override same-timestamp live event", payload.ActiveSessions)
	}
}

func TestHermesCompatibilityBatchedLiveSessionsTerminalSameTimestampWins(t *testing.T) {
	cfg := config.Config{
		PublicBaseURL:                   "http://example.test",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}
	base := storepkg.NewMemoryStore()
	trace := base.ListTraces()[0]
	conversationID := trace.ConversationID
	if conversationID == "" {
		conversationID = base.ListConversations()[0].ID
	}
	now := time.Now().UTC()
	store := &batchedHermesLiveLedgerStore{
		MemoryStore: base,
		events: []storepkg.HermesLiveLedgerEvent{
			{
				TraceID:        trace.TraceID,
				ConversationID: conversationID,
				Kind:           "tool.progress",
				Status:         "running",
				RecordedAt:     now,
			},
			{
				TraceID:        trace.TraceID,
				ConversationID: conversationID,
				Kind:           "tool.complete",
				Status:         "completed",
				RecordedAt:     now,
			},
		},
	}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		ActiveSessions int `json:"active_sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if payload.ActiveSessions != 0 {
		t.Fatalf("active_sessions=%d, want terminal event to override same-timestamp live event", payload.ActiveSessions)
	}
	if store.batchCalls == 0 {
		t.Fatal("batched ledger reader was not called")
	}
	if store.byTraceCalls != 0 {
		t.Fatalf("per-trace ledger calls=%d, want 0 for status with batched reader", store.byTraceCalls)
	}
}

func TestHermesCompatibilitySessionDetailUsesOlderLedgerActivityWithBatchedReader(t *testing.T) {
	cfg := config.Config{
		PublicBaseURL:                   "http://example.test",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}
	base := storepkg.NewMemoryStore()
	trace := base.ListTraces()[0]
	baseline := trace.EndedAt
	if baseline.IsZero() {
		baseline = trace.StartedAt
	}
	ledgerAt := baseline.Add(10 * time.Minute)
	if err := base.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-older-display-activity",
			ExecutionID: "hexec-older-display-activity",
			TraceID:     trace.TraceID,
			WorkflowID:  trace.WorkflowID,
			Kind:        "tool.complete",
			Status:      "completed",
			Seq:         1,
			RecordedAt:  ledgerAt,
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	store := &batchedHermesLiveLedgerStore{MemoryStore: base}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+trace.TraceID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("session status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		LastActive int64 `json:"last_active"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if payload.LastActive != ledgerAt.Unix() {
		t.Fatalf("last_active=%d, want older ledger activity %d", payload.LastActive, ledgerAt.Unix())
	}
	if store.batchCalls == 0 {
		t.Fatal("batched ledger reader was not called")
	}
	if store.byTraceCalls == 0 {
		t.Fatal("per-trace ledger activity was not loaded for session detail")
	}
}

func TestHermesCompatibilityStatusCapsRunnerRuntimeProbes(t *testing.T) {
	newSlowRunner := func() *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
			}
		}))
	}
	prod := newSlowRunner()
	defer prod.Close()
	proactive := newSlowRunner()
	defer proactive.Close()
	eval := newSlowRunner()
	defer eval.Close()
	proposal := newSlowRunner()
	defer proposal.Close()

	cfg := config.Config{
		PublicBaseURL:          "http://example.test",
		ProdRunnerBaseURL:      prod.URL,
		ProactiveRunnerBaseURL: proactive.URL,
		EvalRunnerBaseURL:      eval.URL,
		ProposalRunnerBaseURL:  proposal.URL,
	}
	router := NewRouter(cfg, storepkg.NewMemoryStore())

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	start := time.Now()
	router.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d body=%s", rec.Code, rec.Body.String())
	}
	if elapsed > 1200*time.Millisecond {
		t.Fatalf("/api/status took %s with slow runner probes, want capped parallel probes", elapsed)
	}
	var payload struct {
		GatewayPlatforms map[string]map[string]string `json:"gateway_platforms"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if len(payload.GatewayPlatforms) != 4 {
		t.Fatalf("gateway platforms = %d, want 4", len(payload.GatewayPlatforms))
	}
}

func TestHermesCompatibilityStaleLedgerActivityDoesNotStayLive(t *testing.T) {
	cfg := config.Config{
		PublicBaseURL:                   "http://example.test",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}
	store := storepkg.NewMemoryStore()
	trace := store.ListTraces()[0]
	traceID := trace.TraceID
	conversationID := trace.ConversationID
	if conversationID == "" {
		conversationID = store.ListConversations()[0].ID
	}
	stale := time.Now().UTC().Add(-2 * time.Minute)
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-stale-terminal-output",
			ExecutionID: "hexec-ledger-stale-ui",
			TraceID:     traceID,
			WorkflowID:  trace.WorkflowID,
			Kind:        "terminal.output",
			Status:      "streaming",
			Seq:         1,
			Payload:     map[string]any{"tool_name": "terminal", "chunk_text": "old output"},
			RecordedAt:  stale,
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents(stale) error = %v", err)
	}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=200", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("sessions status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Sessions []struct {
			ID       string `json:"id"`
			IsActive bool   `json:"is_active"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	for _, id := range []string{traceID, conversationID} {
		for _, session := range payload.Sessions {
			if session.ID == id && session.IsActive {
				t.Fatalf("session %s is_active=true from stale ledger activity", id)
			}
		}
	}
}

func TestHermesCompatibilitySessionMessagesMapConversationsAndTraces(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	conversationID := store.ListConversations()[0].ID
	traceID := store.ListTraces()[0].TraceID

	for _, sessionID := range []string{conversationID, traceID} {
		req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sessionID+"/messages", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("messages(%s) status = %d body=%s", sessionID, rec.Code, rec.Body.String())
		}
		var payload struct {
			SessionID string `json:"session_id"`
			Messages  []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode messages(%s): %v", sessionID, err)
		}
		if payload.SessionID != sessionID {
			t.Fatalf("session_id = %q want %q", payload.SessionID, sessionID)
		}
		if len(payload.Messages) == 0 {
			t.Fatalf("expected messages for %s", sessionID)
		}
	}
}

func TestHermesCompatibilityProjectsLedgerToolCallsIntoSessions(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-tool-generation",
			ExecutionID: "hexec-tool-ui",
			TraceID:     traceID,
			WorkflowID:  "workflow-tool-ui",
			Kind:        "tool.generation.started",
			Status:      "running",
			Seq:         1,
			Payload:     map[string]any{"tool_name": "delegate_task"},
			RecordedAt:  now,
		},
		{
			ID:          "xled-tool-start",
			ExecutionID: "hexec-tool-ui",
			TraceID:     traceID,
			WorkflowID:  "workflow-tool-ui",
			Kind:        "tool.call.started",
			Status:      "running",
			Seq:         2,
			Payload: map[string]any{
				"tool_call_id": "call-repo-search",
				"tool_name":    "repo.search",
				"args":         map[string]any{"query": "numo"},
			},
			RecordedAt: now.Add(time.Millisecond),
		},
		{
			ID:          "xled-tool-complete",
			ExecutionID: "hexec-tool-ui",
			TraceID:     traceID,
			WorkflowID:  "workflow-tool-ui",
			Kind:        "tool.call.completed",
			Status:      "completed",
			Seq:         3,
			Payload: map[string]any{
				"tool_call_id": "call-repo-search",
				"tool_name":    "repo.search",
				"result":       "found numo docs",
			},
			RecordedAt: now.Add(2 * time.Millisecond),
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+traceID, nil)
	sessionRec := httptest.NewRecorder()
	router.ServeHTTP(sessionRec, sessionReq)
	if sessionRec.Code != http.StatusOK {
		t.Fatalf("session status = %d body=%s", sessionRec.Code, sessionRec.Body.String())
	}
	var sessionPayload struct {
		ToolCallCount int `json:"tool_call_count"`
	}
	if err := json.Unmarshal(sessionRec.Body.Bytes(), &sessionPayload); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if sessionPayload.ToolCallCount < 2 {
		t.Fatalf("tool_call_count = %d, want projected ledger calls", sessionPayload.ToolCallCount)
	}

	messagesReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+traceID+"/messages", nil)
	messagesRec := httptest.NewRecorder()
	router.ServeHTTP(messagesRec, messagesReq)
	if messagesRec.Code != http.StatusOK {
		t.Fatalf("messages status = %d body=%s", messagesRec.Code, messagesRec.Body.String())
	}
	var messagesPayload struct {
		Messages []struct {
			Role       string `json:"role"`
			Content    string `json:"content"`
			ToolName   string `json:"tool_name"`
			ToolCallID string `json:"tool_call_id"`
			ToolCalls  []struct {
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(messagesRec.Body.Bytes(), &messagesPayload); err != nil {
		t.Fatalf("decode messages: %v", err)
	}
	seenGeneration := false
	seenCompleted := false
	for _, msg := range messagesPayload.Messages {
		if msg.Role != "tool" || len(msg.ToolCalls) == 0 {
			continue
		}
		if msg.ToolName == "delegate_task" {
			seenGeneration = true
		}
		if msg.ToolName == "repo.search" {
			seenCompleted = true
			if msg.ToolCallID != "call-repo-search" {
				t.Fatalf("repo.search tool_call_id = %q, want call-repo-search", msg.ToolCallID)
			}
			if !strings.Contains(msg.Content, "found numo docs") {
				t.Fatalf("repo.search content = %q, want completion result", msg.Content)
			}
			if got := msg.ToolCalls[0].Function.Arguments; !strings.Contains(got, "numo") {
				t.Fatalf("repo.search arguments = %q, want started args", got)
			}
		}
	}
	if !seenGeneration || !seenCompleted {
		t.Fatalf("projected tools generation=%v completed=%v messages=%+v", seenGeneration, seenCompleted, messagesPayload.Messages)
	}
}

func TestHermesCompatibilitySessionStreamUsesHermesEventsAndResume(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-hermes-1",
			ExecutionID: "hexec-main",
			TraceID:     traceID,
			WorkflowID:  "workflow-main",
			Kind:        "tool.call",
			Status:      "running",
			Seq:         1,
			Payload:     map[string]any{"tool_name": "repo.search", "summary": "searching"},
			RecordedAt:  now,
		},
		{
			ID:          "xled-hermes-2",
			ExecutionID: "hexec-main",
			TraceID:     traceID,
			WorkflowID:  "workflow-main",
			Kind:        "tool.call",
			Status:      "complete",
			Seq:         2,
			Payload:     map[string]any{"tool_name": "repo.search", "summary": "done"},
			RecordedAt:  now.Add(time.Millisecond),
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+traceID+"/stream", nil).WithContext(ctx)
	req.Header.Set("Last-Event-ID", "xled-hermes-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("stream status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "xled-hermes-1") {
		t.Fatalf("expected resume to skip first event, got %q", body)
	}
	if !strings.Contains(body, "event: tool.complete") || !strings.Contains(body, "xled-hermes-2") {
		t.Fatalf("expected Hermes tool.complete event, got %q", body)
	}
	if !strings.Contains(body, `"name":"repo.search"`) || !strings.Contains(body, `"tool_id":"xled-hermes-2"`) {
		t.Fatalf("expected Hermes-compatible tool aliases, got %q", body)
	}
}

func TestHermesCompatibilitySessionStreamHonorsLimit(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-hermes-limit-1",
			ExecutionID: "hexec-hermes-limit",
			TraceID:     traceID,
			WorkflowID:  "workflow-hermes-limit",
			Kind:        "tool.call",
			Status:      "running",
			Seq:         1,
			Payload:     map[string]any{"tool_name": "repo.search", "summary": "oldest"},
			RecordedAt:  now,
		},
		{
			ID:          "xled-hermes-limit-2",
			ExecutionID: "hexec-hermes-limit",
			TraceID:     traceID,
			WorkflowID:  "workflow-hermes-limit",
			Kind:        "tool.call",
			Status:      "running",
			Seq:         2,
			Payload:     map[string]any{"tool_name": "repo.search", "summary": "middle"},
			RecordedAt:  now.Add(time.Millisecond),
		},
		{
			ID:          "xled-hermes-limit-3",
			ExecutionID: "hexec-hermes-limit",
			TraceID:     traceID,
			WorkflowID:  "workflow-hermes-limit",
			Kind:        "tool.call",
			Status:      "complete",
			Seq:         3,
			Payload:     map[string]any{"tool_name": "repo.search", "summary": "latest"},
			RecordedAt:  now.Add(2 * time.Millisecond),
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+traceID+"/stream?limit=2", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("stream status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "oldest") {
		t.Fatalf("expected limited Hermes stream to omit oldest event, got %q", body)
	}
	if !strings.Contains(body, "middle") || !strings.Contains(body, "latest") {
		t.Fatalf("expected limited Hermes stream to include latest events, got %q", body)
	}
}

func TestTraceStreamBackfillsAllEventsWhenResumeIDIsMissing(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ExecutionID: "hexec-main",
			TraceID:     traceID,
			WorkflowID:  "workflow-main",
			PhaseID:     "investigate",
			Kind:        "model.output.delta",
			Status:      "streaming",
			Seq:         1,
			Payload:     map[string]any{"role": "prod", "delta": "first"},
			RecordedAt:  now,
		},
		{
			ExecutionID: "hexec-main",
			TraceID:     traceID,
			WorkflowID:  "workflow-main",
			PhaseID:     "investigate",
			Kind:        "tool.call.completed",
			Status:      "completed",
			Seq:         2,
			Payload:     map[string]any{"role": "prod", "summary": "second"},
			RecordedAt:  now.Add(time.Millisecond),
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID+"/stream?scope=main", nil).WithContext(ctx)
	req.Header.Set("Last-Event-ID", "xled-stale")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("stream status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "first") || !strings.Contains(body, "second") {
		t.Fatalf("expected stale resume id to backfill all events, got %q", body)
	}
}

func TestTraceStreamBackfillHonorsLimit(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-limit-1",
			ExecutionID: "hexec-limit",
			TraceID:     traceID,
			WorkflowID:  "workflow-limit",
			Kind:        "model.output.delta",
			Status:      "streaming",
			Seq:         1,
			Payload:     map[string]any{"role": "prod", "delta": "oldest"},
			RecordedAt:  now,
		},
		{
			ID:          "xled-limit-2",
			ExecutionID: "hexec-limit",
			TraceID:     traceID,
			WorkflowID:  "workflow-limit",
			Kind:        "model.output.delta",
			Status:      "streaming",
			Seq:         2,
			Payload:     map[string]any{"role": "prod", "delta": "middle"},
			RecordedAt:  now.Add(time.Millisecond),
		},
		{
			ID:          "xled-limit-3",
			ExecutionID: "hexec-limit",
			TraceID:     traceID,
			WorkflowID:  "workflow-limit",
			Kind:        "model.output.delta",
			Status:      "streaming",
			Seq:         3,
			Payload:     map[string]any{"role": "prod", "delta": "latest"},
			RecordedAt:  now.Add(2 * time.Millisecond),
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID+"/stream?scope=main&limit=2", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("stream status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "oldest") {
		t.Fatalf("expected initial stream backfill to omit oldest event, got %q", body)
	}
	if !strings.Contains(body, "middle") || !strings.Contains(body, "latest") {
		t.Fatalf("expected initial stream backfill to include latest limited events, got %q", body)
	}
}

func TestTraceLedgerEndpointPaginatesOlderEvents(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ID:          "xled-page-1",
			ExecutionID: "hexec-page",
			TraceID:     traceID,
			WorkflowID:  "workflow-page",
			Kind:        "model.output.delta",
			Status:      "streaming",
			Seq:         1,
			Payload:     map[string]any{"role": "prod", "delta": "oldest"},
			RecordedAt:  now,
		},
		{
			ID:          "xled-page-2",
			ExecutionID: "hexec-page",
			TraceID:     traceID,
			WorkflowID:  "workflow-page",
			Kind:        "model.output.delta",
			Status:      "streaming",
			Seq:         2,
			Payload:     map[string]any{"role": "prod", "delta": "middle"},
			RecordedAt:  now.Add(time.Millisecond),
		},
		{
			ID:          "xled-page-3",
			ExecutionID: "hexec-page",
			TraceID:     traceID,
			WorkflowID:  "workflow-page",
			Kind:        "tool.call.completed",
			Status:      "completed",
			Seq:         3,
			Payload:     map[string]any{"role": "prod", "summary": "latest"},
			RecordedAt:  now.Add(2 * time.Millisecond),
		},
	}); err != nil {
		t.Fatalf("RecordExecutionLedgerEvents() error = %v", err)
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	req := httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID+"/ledger?scope=main&limit=2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ledger status = %d body=%s", rec.Code, rec.Body.String())
	}
	var latest struct {
		Events []events.ExecutionLedgerEvent `json:"events"`
		Paging struct {
			HasMore    bool   `json:"has_more"`
			NextBefore string `json:"next_before"`
		} `json:"paging"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &latest); err != nil {
		t.Fatalf("decode latest page: %v", err)
	}
	if len(latest.Events) != 2 || latest.Events[0].ID != "xled-page-2" || latest.Events[1].ID != "xled-page-3" {
		t.Fatalf("expected latest two chronological events, got %#v", latest.Events)
	}
	if !latest.Paging.HasMore || latest.Paging.NextBefore != "xled-page-2" {
		t.Fatalf("expected older page cursor, got %#v", latest.Paging)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/traces/"+traceID+"/ledger?scope=main&limit=2&before="+latest.Paging.NextBefore, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("older ledger status = %d body=%s", rec.Code, rec.Body.String())
	}
	var older struct {
		Events []events.ExecutionLedgerEvent `json:"events"`
		Paging struct {
			HasMore bool `json:"has_more"`
		} `json:"paging"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &older); err != nil {
		t.Fatalf("decode older page: %v", err)
	}
	if len(older.Events) != 1 || older.Events[0].ID != "xled-page-1" {
		t.Fatalf("expected only older event, got %#v", older.Events)
	}
	if older.Paging.HasMore {
		t.Fatalf("expected no more older pages")
	}
}

func TestRouterConversationDetailEnrichesPlainSlackTranscriptEntities(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineIngress,
		AggregateID: "slack:171000111.000100",
		CommandKind: string(transition.CommandIngressRecordSlack),
		CommandID:   "cmd-router-transcript-plain-text",
		Actor:       "tester",
		OccurredAt:  now,
		Payload: map[string]any{
			"bot_role":   "orchestrator",
			"team_id":    "T123",
			"channel_id": "C123",
			"thread_ts":  "171000111.000100",
			"user_id":    "U123",
			"text":       "Hello @U0ASDQKU3UL, please review #C0AKH5SNGKH and #C0AL7EKNHDF.",
			"ts":         "171000111.000100",
			"entity_refs": []map[string]any{
				{"kind": "user", "id": "U0ASDQKU3UL", "source": "plain_text"},
				{"kind": "channel", "id": "C0AKH5SNGKH", "source": "plain_text"},
				{"kind": "channel", "id": "C0AL7EKNHDF", "source": "plain_text"},
			},
			"created_at": now,
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(slack ingress) error = %v", err)
	}

	ingestionID := strings.TrimSpace(receipt.ResultRef)
	if ingestionID == "" {
		t.Fatal("expected ingestion result ref")
	}
	conversationID := ""
	for _, item := range store.ListIngestions() {
		if item.ID == ingestionID {
			conversationID = item.ConversationID
			break
		}
	}
	if conversationID == "" {
		t.Fatalf("expected conversation for ingestion %s", ingestionID)
	}

	router := newRouterWithTranscriptResolver(config.Config{PublicBaseURL: "http://example.test"}, store, stubSlackTranscriptResolver{
		userNames: map[string]string{
			"U0ASDQKU3UL": "blake",
		},
		channelNames: map[string]string{
			"C0AKH5SNGKH": "depin-backend",
			"C0AL7EKNHDF": "numo-project",
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/conversations/"+conversationID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("conversation detail status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode conversation detail: %v", err)
	}
	transcript, ok := payload["transcript"].([]any)
	if !ok || len(transcript) != 1 {
		t.Fatalf("expected one transcript entry, got %#v", payload["transcript"])
	}
	entry, ok := transcript[0].(map[string]any)
	if !ok {
		t.Fatalf("expected transcript entry object, got %#v", transcript[0])
	}
	metadata, ok := entry["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected transcript metadata object, got %#v", entry["metadata"])
	}
	userNames, ok := metadata[slackUserNamesMetadataKey].(map[string]any)
	if !ok {
		t.Fatalf("expected resolved slack_user_names metadata, got %#v", metadata[slackUserNamesMetadataKey])
	}
	gotUserName, _ := userNames["U0ASDQKU3UL"].(string)
	if got := strings.TrimSpace(gotUserName); got != "blake" {
		t.Fatalf("expected resolved user label, got %q in %#v", got, userNames)
	}
	channelNames, ok := metadata[slackChannelNamesMetadataKey].(map[string]any)
	if !ok {
		t.Fatalf("expected resolved slack_channel_names metadata, got %#v", metadata[slackChannelNamesMetadataKey])
	}
	gotPrimaryChannel, _ := channelNames["C0AKH5SNGKH"].(string)
	if got := strings.TrimSpace(gotPrimaryChannel); got != "depin-backend" {
		t.Fatalf("expected depin-backend channel label, got %q in %#v", got, channelNames)
	}
	gotSecondaryChannel, _ := channelNames["C0AL7EKNHDF"].(string)
	if got := strings.TrimSpace(gotSecondaryChannel); got != "numo-project" {
		t.Fatalf("expected numo-project channel label, got %q in %#v", got, channelNames)
	}
}

func TestRouterFeedbackAndReplayRoutesSubmitProblemLineCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatal("expected seeded trace")
	}
	now := time.Now().UTC()
	intent := seedRouterActionIntent(t, store, action.Intent{
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        traceID,
		Kind:           action.KindToolRead,
		CreatedAt:      now,
	}, "router-feedback", map[string]any{
		"executor":     "native-hermes-required",
		"provider":     "github",
		"provider_ref": "repo-activity-ref",
	})
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	feedbackReq := httptest.NewRequest(http.MethodPost, "/api/feedback", strings.NewReader(`{
		"target_type":"action_intent",
		"target_id":"`+intent.ID+`",
		"verdict":"useful",
		"reviewer_id":"alice"
	}`))
	feedbackRec := httptest.NewRecorder()
	router.ServeHTTP(feedbackRec, feedbackReq)
	if feedbackRec.Code != http.StatusCreated {
		t.Fatalf("feedback status = %d, want %d", feedbackRec.Code, http.StatusCreated)
	}
	foundFeedbackEvent := false
	for _, item := range store.ListDomainEvents() {
		if item.EventKind == "problem_line_feedback_recorded" {
			foundFeedbackEvent = true
			break
		}
	}
	if !foundFeedbackEvent {
		t.Fatal("expected feedback route to record a problem-line domain event")
	}

	badReplayReq := httptest.NewRequest(http.MethodPost, "/api/traces/"+traceID+"/replay", strings.NewReader(`{"requested_by":`))
	badReplayRec := httptest.NewRecorder()
	router.ServeHTTP(badReplayRec, badReplayReq)
	if badReplayRec.Code != http.StatusBadRequest {
		t.Fatalf("malformed replay status = %d, want %d", badReplayRec.Code, http.StatusBadRequest)
	}

	replayReq := httptest.NewRequest(http.MethodPost, "/api/traces/"+traceID+"/replay", strings.NewReader(`{"requested_by":"alice"}`))
	replayRec := httptest.NewRecorder()
	router.ServeHTTP(replayRec, replayReq)
	if replayRec.Code != http.StatusAccepted {
		t.Fatalf("replay status = %d, want %d", replayRec.Code, http.StatusAccepted)
	}
	var replayReceipt transition.CommandReceipt
	if err := json.NewDecoder(replayRec.Body).Decode(&replayReceipt); err != nil {
		t.Fatalf("decode replay receipt: %v", err)
	}
	if replayReceipt.MachineKind != transition.MachineWorkflowLine || replayReceipt.CommandKind != string(transition.CommandWorkflowLineScheduleRetry) {
		t.Fatalf("unexpected replay receipt %+v", replayReceipt)
	}
	foundReplayEvent := false
	for _, item := range store.ListDomainEvents() {
		if item.EventKind == "workflow_line_retry_scheduled" {
			foundReplayEvent = true
		}
	}
	if !foundReplayEvent {
		t.Fatal("expected replay route to record a workflow-line domain event")
	}
	trace, ok = store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected original trace %s", traceID)
	}
	line, ok := store.GetWorkflowLine(trace.Summary.CaseID)
	if !ok {
		t.Fatalf("expected workflow line for case %s", trace.Summary.CaseID)
	}
	if line.CurrentWorkflowID == "" || line.CurrentWorkflowID == trace.Summary.WorkflowID {
		t.Fatalf("expected replay route to create a successor workflow attempt, got %+v", line)
	}
	if _, ok := store.GetCommandReceipt(replayReceipt.CommandID + ":activate"); !ok {
		t.Fatal("expected replay route to activate the scheduled workflow retry")
	}
	startCommandID := "cmd-workflow:" + line.CurrentWorkflowID + ":" + string(transition.CommandWorkflowStarted)
	if _, ok := store.GetCommandReceipt(startCommandID); !ok {
		t.Fatalf("expected replay route to start successor workflow %s", line.CurrentWorkflowID)
	}
	successor, ok := findWorkflowView(store.ListWorkflows(), line.CurrentWorkflowID)
	if !ok {
		t.Fatalf("expected successor workflow %s", line.CurrentWorkflowID)
	}
	replayTrace, ok := store.GetTrace(successor.TraceID)
	if !ok {
		t.Fatalf("expected successor trace %s", successor.TraceID)
	}
	if replayTrace.Summary.SupersedesTraceID != traceID {
		t.Fatalf("expected successor trace to supersede %s, got %s", traceID, replayTrace.Summary.SupersedesTraceID)
	}
}

func TestRouterResetAppDataRouteClearsStore(t *testing.T) {
	store := storepkg.NewMemoryStore()
	if len(store.ListConversations()) == 0 {
		t.Fatal("expected seeded conversations before reset")
	}
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	resetReq := httptest.NewRequest(http.MethodPost, "/api/app-data/reset", strings.NewReader(`{}`))
	resetRec := httptest.NewRecorder()
	router.ServeHTTP(resetRec, resetReq)
	if resetRec.Code != http.StatusOK {
		t.Fatalf("reset status = %d, want %d", resetRec.Code, http.StatusOK)
	}

	var payload storepkg.AppDataResetResult
	if err := json.NewDecoder(resetRec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode reset payload: %v", err)
	}
	if payload.Backend != "memory" {
		t.Fatalf("reset backend = %q, want memory", payload.Backend)
	}
	if len(store.ListConversations()) != 0 {
		t.Fatal("expected conversations to be cleared")
	}
	if len(store.ListTraces()) != 0 {
		t.Fatal("expected traces to be cleared")
	}
	if len(store.ListHarnessProfiles()) != 0 {
		t.Fatal("expected harness profiles to be cleared")
	}
	if settings := store.GetSettings(); settings.ActiveProposalCap != 2 {
		t.Fatalf("expected default proposal cap after reset, got %+v", settings)
	}
}

func TestRouterProposalDetailAndRuntimeEndpoints(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":                  "ok",
			"role":                    "eval",
			"backend":                 "hermes-aiagent",
			"provider":                "openai",
			"model":                   "openai/gpt-5.4",
			"provider_model":          "gpt-5.4",
			"api_mode":                "codex_responses",
			"reasoning_effort":        "xhigh",
			"available":               true,
			"hermes_available":        true,
			"openai_configured":       true,
			"honcho_configured":       true,
			"honcho_available":        true,
			"honcho_base_url":         "http://use1-stage-rsi-agent-platform-honcho-api:8000",
			"honcho_workspace":        "rsi-stage",
			"honcho_environment":      "stage",
			"honcho_recall_mode":      "hybrid",
			"honcho_write_frequency":  "async",
			"honcho_session_strategy": "global",
			"honcho_ai_peer":          "rsi:stage:eval",
		})
	}))
	defer runner.Close()
	honcho := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":               "ok",
			"namespace":            "rsi-stage",
			"db_schema":            "honcho",
			"cache_enabled":        true,
			"cache_url_configured": true,
			"deriver": map[string]any{
				"provider":         "openai",
				"model":            "gpt-5.4",
				"reasoning_effort": "xhigh",
			},
			"summary": map[string]any{
				"provider":         "openai",
				"model":            "gpt-5.4",
				"reasoning_effort": "xhigh",
			},
			"dialectic_levels": map[string]any{
				"minimal": map[string]any{
					"provider":               "openai",
					"model":                  "gpt-5.4",
					"reasoning_effort":       "xhigh",
					"thinking_budget_tokens": 0,
				},
			},
		})
	}))
	defer honcho.Close()

	store := storepkg.NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposal")
	}
	proposal := proposals[0]

	approved, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for repo change.",
		ReviewerID: "operator",
	})
	if err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	if strings.TrimSpace(approved.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt after approval, got %+v", approved)
	}
	seedRouterRepoChangeJobViaCommand(t, store, approved, "proposal-list", "job-proposal-list-1", "codex/proposal-test")
	if _, err := store.RecordPRAttempt(improvement.PRAttempt{
		ProposalID:       proposal.ID,
		AttemptID:        approved.CurrentAttemptID,
		Repo:             "rsi-agent-platform",
		BranchName:       "codex/proposal-test",
		PRURL:            "https://github.com/piplabs/rsi-agent-platform/pull/42",
		Status:           string(review.ProposalPROpen),
		ValidationStatus: "pending",
	}); err != nil {
		t.Fatalf("record pr attempt: %v", err)
	}

	cfg := config.Config{
		PublicBaseURL:          "http://example.test",
		RunnerBaseURL:          runner.URL,
		ProdRunnerBaseURL:      runner.URL,
		ProactiveRunnerBaseURL: runner.URL,
		EvalRunnerBaseURL:      runner.URL,
		ProposalRunnerBaseURL:  runner.URL,
		HonchoRuntimeBaseURL:   honcho.URL,
	}
	router := NewRouter(cfg, store)

	proposalListReq := httptest.NewRequest(http.MethodGet, "/api/proposals", nil)
	proposalListRec := httptest.NewRecorder()
	router.ServeHTTP(proposalListRec, proposalListReq)
	if proposalListRec.Code != http.StatusOK {
		t.Fatalf("proposal list status = %d, want %d", proposalListRec.Code, http.StatusOK)
	}

	var proposalListPayload struct {
		Proposals []map[string]any `json:"proposals"`
	}
	if err := json.NewDecoder(proposalListRec.Body).Decode(&proposalListPayload); err != nil {
		t.Fatalf("decode proposal list: %v", err)
	}
	if len(proposalListPayload.Proposals) == 0 {
		t.Fatal("expected at least one proposal summary")
	}
	firstProposal := proposalListPayload.Proposals[0]
	if got, _ := firstProposal["repo_change_status"].(string); got == "" {
		t.Fatal("expected compact repo_change_status on proposal summary")
	}
	if got, _ := firstProposal["pr_status"].(string); got == "" {
		t.Fatal("expected compact pr_status on proposal summary")
	}
	if got, _ := firstProposal["pr_url"].(string); got == "" {
		t.Fatal("expected compact pr_url on proposal summary")
	}
	if _, ok := firstProposal["repo_change_jobs"]; ok {
		t.Fatal("proposal list should not hydrate repo_change_jobs history")
	}
	if _, ok := firstProposal["pr_attempts"]; ok {
		t.Fatal("proposal list should not hydrate pr_attempts history")
	}

	proposalReq := httptest.NewRequest(http.MethodGet, "/api/proposals/"+proposal.ID, nil)
	proposalRec := httptest.NewRecorder()
	router.ServeHTTP(proposalRec, proposalReq)
	if proposalRec.Code != http.StatusOK {
		t.Fatalf("proposal detail status = %d, want %d", proposalRec.Code, http.StatusOK)
	}

	var proposalPayload map[string]any
	if err := json.NewDecoder(proposalRec.Body).Decode(&proposalPayload); err != nil {
		t.Fatalf("decode proposal detail: %v", err)
	}
	if items, ok := proposalPayload["workspace_sessions"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected workspace_sessions array with at least one item")
	}
	if items, ok := proposalPayload["validation_runs"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected validation_runs array with at least one item")
	}
	if items, ok := proposalPayload["pr_attempts"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected pr_attempts array with at least one item")
	}
	if items, ok := proposalPayload["attempts"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected attempts array with at least one item")
	}
	if _, ok := proposalPayload["current_phase"].(map[string]any); !ok {
		t.Fatal("expected current_phase object in proposal detail")
	}
	if items, ok := proposalPayload["linked_trace_summaries"].([]any); !ok || len(items) == 0 {
		t.Fatal("expected linked_trace_summaries array with at least one item")
	}
	if _, ok := proposalPayload["action_intents"].([]any); !ok {
		t.Fatal("expected action_intents array in proposal detail")
	}
	if _, ok := proposalPayload["outcomes"].([]any); !ok {
		t.Fatal("expected outcomes array in proposal detail")
	}
	if _, ok := proposalPayload["knowledge_entries"].([]any); !ok {
		t.Fatal("expected knowledge_entries array in proposal detail")
	}
	attempts := proposalPayload["attempts"].([]any)
	attempt, _ := attempts[0].(map[string]any)
	attemptID, _ := attempt["id"].(string)
	if attemptID == "" {
		t.Fatal("expected attempt id in proposal detail")
	}
	attemptReq := httptest.NewRequest(http.MethodGet, "/api/proposals/"+proposal.ID+"/attempts/"+attemptID, nil)
	attemptRec := httptest.NewRecorder()
	router.ServeHTTP(attemptRec, attemptReq)
	if attemptRec.Code != http.StatusOK {
		t.Fatalf("attempt detail status = %d, want %d", attemptRec.Code, http.StatusOK)
	}

	runtimeReq := httptest.NewRequest(http.MethodGet, "/api/runtime", nil)
	runtimeRec := httptest.NewRecorder()
	router.ServeHTTP(runtimeRec, runtimeReq)
	if runtimeRec.Code != http.StatusOK {
		t.Fatalf("runtime status = %d, want %d", runtimeRec.Code, http.StatusOK)
	}

	var runtimePayload struct {
		Roles  []map[string]any `json:"roles"`
		Honcho map[string]any   `json:"honcho"`
	}
	if err := json.NewDecoder(runtimeRec.Body).Decode(&runtimePayload); err != nil {
		t.Fatalf("decode runtime payload: %v", err)
	}
	if len(runtimePayload.Roles) != 4 {
		t.Fatalf("expected 4 runtime roles, got %d", len(runtimePayload.Roles))
	}
	if got, _ := runtimePayload.Honcho["status"].(string); got != "ok" {
		t.Fatalf("expected honcho runtime ok, got %q", got)
	}
	if got, _ := runtimePayload.Honcho["namespace"].(string); got != "rsi-stage" {
		t.Fatalf("expected honcho namespace rsi-stage, got %q", got)
	}
	deriver, ok := runtimePayload.Honcho["deriver"].(map[string]any)
	if !ok {
		t.Fatal("expected honcho deriver payload")
	}
	if got, _ := deriver["model"].(string); got != "gpt-5.4" {
		t.Fatalf("expected honcho deriver model gpt-5.4, got %q", got)
	}
	if got, _ := deriver["reasoning_effort"].(string); got != "xhigh" {
		t.Fatalf("expected honcho deriver reasoning xhigh, got %q", got)
	}
	for _, role := range runtimePayload.Roles {
		if got, _ := role["model"].(string); got != "openai/gpt-5.4" {
			t.Fatalf("expected model openai/gpt-5.4, got %q", got)
		}
		if got, _ := role["reasoning_effort"].(string); got != "xhigh" {
			t.Fatalf("expected reasoning effort xhigh, got %q", got)
		}
		if got, _ := role["api_mode"].(string); got != "codex_responses" {
			t.Fatalf("expected api_mode codex_responses, got %q", got)
		}
		if got, _ := role["honcho_workspace"].(string); got != "rsi-stage" {
			t.Fatalf("expected honcho_workspace rsi-stage, got %q", got)
		}
		if got, _ := role["honcho_recall_mode"].(string); got != "hybrid" {
			t.Fatalf("expected honcho_recall_mode hybrid, got %q", got)
		}
	}
}

func TestRouterProposalRetryEndpoint(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]

	approved, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "operator",
	})
	if err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	if strings.TrimSpace(approved.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt after approval, got %+v", approved)
	}
	if _, _, err := storepkg.AdvanceProposalToFailedValidationForTesting(store, proposal.ID, time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceProposalToFailedValidationForTesting() error = %v", err)
	}

	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	req := httptest.NewRequest(http.MethodPost, "/api/proposals/"+proposal.ID+"/commands", strings.NewReader(`{"command_kind":"proposal_retry_attempt","actor":"ui-operator"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("retry proposal status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	var receipt transition.CommandReceipt
	if err := json.NewDecoder(rec.Body).Decode(&receipt); err != nil {
		t.Fatalf("decode retry response: %v", err)
	}
	if receipt.ResultRef != proposal.ID {
		t.Fatalf("expected retry command result ref %s, got %s", proposal.ID, receipt.ResultRef)
	}
	updated, ok := findProposalByID(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after retry", proposal.ID)
	}
	if strings.TrimSpace(updated.CurrentAttemptID) == "" {
		t.Fatalf("expected retry to materialize a new current attempt, got %+v", updated)
	}
	foundRetryEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != updated.CurrentAttemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			foundRetryEffect = true
		}
	}
	if !foundRetryEffect {
		t.Fatalf("expected queued retry bootstrap effect for attempt %s", updated.CurrentAttemptID)
	}
}

func TestRouterProposalStopEndpoint(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}

	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)
	req := httptest.NewRequest(http.MethodPost, "/api/proposals/"+proposal.ID+"/commands", strings.NewReader(`{"command_kind":"proposal_stop_line","actor":"ui-operator","payload":{"rationale":"Stop this remediation line."}}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("stop proposal status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	var receipt transition.CommandReceipt
	if err := json.NewDecoder(rec.Body).Decode(&receipt); err != nil {
		t.Fatalf("decode stop receipt: %v", err)
	}
	updated, ok := findProposalByID(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatal("expected proposal to exist")
	}
	if updated.Status != review.ProposalCanceled {
		t.Fatalf("expected canceled status, got %q", updated.Status)
	}
}

func TestRouterCommandEndpointsEvaluateTraceAndUpdateSettings(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	evalReq := httptest.NewRequest(http.MethodPost, "/api/problem-lines/"+traceID+"/commands", strings.NewReader(`{"command_kind":"problem_line_evaluate_trace","actor":"ui-operator","payload":{"trigger":"manual"}}`))
	evalReq.Header.Set("Content-Type", "application/json")
	evalRec := httptest.NewRecorder()
	router.ServeHTTP(evalRec, evalReq)
	if evalRec.Code != http.StatusAccepted {
		t.Fatalf("evaluate command status = %d, want %d", evalRec.Code, http.StatusAccepted)
	}

	settingsReq := httptest.NewRequest(http.MethodPost, "/api/settings/commands", strings.NewReader(`{"command_kind":"settings_update","actor":"ui-operator","payload":{"active_proposal_cap":1}}`))
	settingsReq.Header.Set("Content-Type", "application/json")
	settingsRec := httptest.NewRecorder()
	router.ServeHTTP(settingsRec, settingsReq)
	if settingsRec.Code != http.StatusAccepted {
		t.Fatalf("settings command status = %d, want %d", settingsRec.Code, http.StatusAccepted)
	}
	if got := store.GetSettings().ActiveProposalCap; got != 1 {
		t.Fatalf("expected proposal cap 1, got %d", got)
	}
	if len(store.ListEvalRuns()) == 0 {
		t.Fatal("expected eval run to be created")
	}
}

func TestLegacyMutationRoutesAbsent(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	proposalID := store.ListProposals()[0].ID
	knowledgeEntry := seedRouterKnowledgeEntry(t, store, "knowledge-legacy-route-coverage", knowledge.Entry{
		Tier:       knowledge.TierWorking,
		Kind:       knowledge.KindFact,
		ScopeType:  knowledge.ScopeGlobal,
		Title:      "Legacy route coverage",
		Status:     knowledge.StatusDraft,
		SourceType: knowledge.SourceAgent,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}, nil, "router-legacy-route-knowledge")
	knowledgeID := knowledgeEntry.ID
	router := NewRouter(config.Config{PublicBaseURL: "http://example.test"}, store)

	for _, path := range []string{
		"/api/traces/" + traceID + "/evaluate",
		"/api/proposals/promote",
		"/api/proposals/" + proposalID + "/decision",
		"/api/proposals/" + proposalID + "/retry",
		"/api/proposals/" + proposalID + "/stop",
		"/api/knowledge/" + knowledgeID + "/review",
		"/api/settings",
	} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected %s to be absent, got %d", path, rec.Code)
		}
	}
}

func findProposalByID(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}
