package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/events"
)

func (s *MemoryStore) ReconcileWorkflowTrace(workflowID string) (events.Trace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reconcileWorkflowTraceLocked(workflowID)
}

func (p *PostgresStore) ReconcileWorkflowTrace(workflowID string) (trace events.Trace, repaired bool, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		store, loadErr := loadStore(tx)
		if loadErr != nil {
			return loadErr
		}
		trace, repaired, err = store.reconcileWorkflowTraceLocked(workflowID)
		if err != nil {
			return err
		}
		if !repaired {
			return nil
		}
		return persistStore(tx, store)
	})
	return
}

func (s *MemoryStore) reconcileWorkflowTraceLocked(workflowID string) (events.Trace, bool, error) {
	workflowID = strings.TrimSpace(workflowID)
	if workflowID == "" {
		return events.Trace{}, false, fmt.Errorf("workflow id is required")
	}
	workflow, ok := findWorkflowRecord(s.workflows, workflowID)
	if !ok {
		return events.Trace{}, false, fmt.Errorf("workflow %s not found", workflowID)
	}
	trace, ok := s.traces[workflow.TraceID]
	if !ok {
		return events.Trace{}, false, fmt.Errorf("trace %s not found", workflow.TraceID)
	}

	status, eventType, description, ok := reconciledTraceProjection(workflow)
	if !ok {
		return trace, false, nil
	}
	needsStatus := trace.Summary.Status != status
	needsEvent := !traceHasEventType(trace, eventType)
	needsVerdict := strings.TrimSpace(trace.Summary.LastVerdict) != strings.TrimSpace(workflow.LastVerdict)
	repaired := false

	if needsStatus || needsEvent || needsVerdict {
		occurredAt := workflow.UpdatedAt
		if workflow.CompletedAt != nil && !workflow.CompletedAt.IsZero() {
			occurredAt = *workflow.CompletedAt
		}
		if occurredAt.IsZero() {
			occurredAt = time.Now().UTC()
		}
		update := TraceUpdate{
			Status: ptrStatus(status),
		}
		if verdict := strings.TrimSpace(workflow.LastVerdict); needsVerdict || verdict != "" {
			update.LastVerdict = &verdict
		}
		if needsEvent {
			update.Events = []events.TraceEvent{{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     "control-plane",
				Actor:       "worker",
				EventType:   eventType,
				Status:      status,
				StartedAt:   occurredAt,
				Description: description,
			}}
		}
		updated, err := s.applyTraceUpdateLocked(trace.Summary.TraceID, update)
		if err != nil {
			return events.Trace{}, false, err
		}
		trace = updated
		repaired = true
	}

	reconciledTrace, actionRepaired, err := s.reconcileProjectedActionTraceLocked(trace, workflow)
	if err != nil {
		return events.Trace{}, false, err
	}
	if actionRepaired {
		trace = reconciledTrace
	}
	return trace, repaired || actionRepaired, nil
}

func reconciledTraceProjection(workflow Workflow) (events.Status, string, string, bool) {
	switch strings.TrimSpace(workflow.Status) {
	case "failed":
		return events.StatusFailed, "workflow.failed", firstNonEmpty(workflow.FailureSummary, workflow.LastError, "workflow failed"), true
	case "completed":
		if strings.TrimSpace(workflow.LastVerdict) == "partial" {
			return events.StatusCompleted, "workflow.completed", "Workflow completed with a partial answer.", true
		}
		return events.StatusCompleted, "workflow.completed", "Workflow completed.", true
	case "needs_human":
		return events.StatusNeedsHuman, "workflow.blocked", firstNonEmpty(workflow.FailureSummary, workflow.LastError, "workflow needs human intervention"), true
	default:
		return "", "", "", false
	}
}

func findWorkflowRecord(items []Workflow, workflowID string) (Workflow, bool) {
	for _, item := range items {
		if item.ID == workflowID {
			return item, true
		}
	}
	return Workflow{}, false
}

func (s *MemoryStore) reconcileProjectedActionTraceLocked(trace events.Trace, workflow Workflow) (events.Trace, bool, error) {
	update := TraceUpdate{}
	for _, intent := range s.actionIntents {
		if strings.TrimSpace(intent.TraceID) != trace.Summary.TraceID {
			continue
		}
		if !isTerminalReconciledActionStatus(intent.Status) {
			continue
		}
		result, ok := latestActionResultLocked(s.actionResults, intent.ID)
		startedAt := intent.CreatedAt
		completedAt := intent.UpdatedAt
		providerRef := ""
		errorMessage := ""
		rawArtifacts := []string{}
		sendStatus := ""
		if ok {
			if !result.StartedAt.IsZero() {
				startedAt = result.StartedAt
			}
			if !result.CompletedAt.IsZero() {
				completedAt = result.CompletedAt
			}
			providerRef = result.ProviderRef
			errorMessage = result.ErrorMessage
		}
		switch intent.Kind {
		case action.KindToolRead:
			toolCallID := firstNonEmpty(providerRef, intent.OperationID, intent.ID)
			if traceHasToolCallRecord(trace, toolCallID, intent.TargetRef) {
				continue
			}
			summary := firstNonEmpty(errorMessage, intent.PolicyVerdict, intent.Rationale, string(intent.Status))
			eventStatus, eventType := actionTraceEventStatus(intent.Status, "tool.completed", "tool.blocked", "tool.failed")
			update.Events = append(update.Events, events.TraceEvent{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     "control-plane",
				Actor:       workflow.AssignedBot,
				EventType:   eventType,
				Status:      eventStatus,
				StartedAt:   startedAt,
				EndedAt:     ptrTime(completedAt),
				Description: summary,
			})
			update.ToolCalls = append(update.ToolCalls, events.ToolCallRecord{
				ID:                    fmt.Sprintf("reconciled-tool-record-%s", intent.ID),
				TraceID:               trace.Summary.TraceID,
				WorkflowID:            trace.Summary.WorkflowID,
				ConversationID:        intent.ConversationID,
				CaseID:                intent.CaseID,
				ToolName:              intent.TargetRef,
				ToolCallID:            toolCallID,
				Request:               intent.RequestPayload,
				Summary:               summary,
				RawArtifactRefs:       rawArtifacts,
				ApprovalState:         intent.ApprovalState,
				InterpretationSummary: summary,
				Status:                string(intent.Status),
				CreatedAt:             completedAt,
			})
		case action.KindSlackPost:
			if traceHasSlackActionRecord(trace, intent.IdempotencyKey) {
				continue
			}
			summary := firstNonEmpty(errorMessage, intent.PolicyVerdict, intent.Rationale, string(intent.Status))
			eventStatus, eventType := actionTraceEventStatus(intent.Status, "slack.reply.posted", "slack.reply.blocked", "slack.reply.failed")
			switch intent.Status {
			case action.StatusSucceeded:
				sendStatus = "posted"
			case action.StatusBlocked:
				sendStatus = firstNonEmpty(intent.PolicyVerdict, "blocked")
			default:
				sendStatus = "failed"
			}
			update.Events = append(update.Events, events.TraceEvent{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "edge",
				Service:     "control-plane",
				Actor:       workflow.AssignedBot,
				EventType:   eventType,
				Status:      eventStatus,
				StartedAt:   startedAt,
				EndedAt:     ptrTime(completedAt),
				Description: summary,
			})
			update.SlackActions = append(update.SlackActions, events.SlackActionRecord{
				ID:             fmt.Sprintf("reconciled-slack-action-%s", intent.ID),
				TraceID:        trace.Summary.TraceID,
				WorkflowID:     trace.Summary.WorkflowID,
				ConversationID: intent.ConversationID,
				CaseID:         intent.CaseID,
				ChannelID:      firstNonEmpty(stringFromPayload(intent.RequestPayload, "channel_id"), intent.TargetRef),
				ThreadTS:       stringFromPayload(intent.RequestPayload, "thread_ts"),
				IdempotencyKey: intent.IdempotencyKey,
				DraftBody:      firstNonEmpty(stringFromPayload(intent.RequestPayload, "draft_body"), stringFromPayload(intent.RequestPayload, "body")),
				FinalBody:      firstNonEmpty(stringFromPayload(intent.RequestPayload, "final_body"), stringFromPayload(intent.RequestPayload, "body")),
				PolicyVerdict:  firstNonEmpty(intent.PolicyVerdict, sendStatus),
				SendStatus:     sendStatus,
				CreatedAt:      completedAt,
			})
		}
	}
	if len(update.Events) == 0 && len(update.ToolCalls) == 0 && len(update.SlackActions) == 0 {
		return trace, false, nil
	}
	updated, err := s.applyTraceUpdateLocked(trace.Summary.TraceID, update)
	if err != nil {
		return events.Trace{}, false, err
	}
	return updated, true, nil
}

func latestActionResultLocked(items map[string][]action.Result, actionID string) (action.Result, bool) {
	results := items[strings.TrimSpace(actionID)]
	if len(results) == 0 {
		return action.Result{}, false
	}
	return results[len(results)-1], true
}

func traceHasToolCallRecord(trace events.Trace, toolCallID string, toolName string) bool {
	for _, item := range trace.ToolCalls {
		if toolCallID != "" && strings.TrimSpace(item.ToolCallID) == strings.TrimSpace(toolCallID) {
			return true
		}
		if toolName != "" && strings.TrimSpace(item.ToolName) == strings.TrimSpace(toolName) && strings.TrimSpace(item.Status) != "" {
			return true
		}
	}
	return false
}

func traceHasSlackActionRecord(trace events.Trace, idempotencyKey string) bool {
	for _, item := range trace.SlackActions {
		if idempotencyKey != "" && strings.TrimSpace(item.IdempotencyKey) == strings.TrimSpace(idempotencyKey) {
			return true
		}
	}
	return false
}

func actionTraceEventStatus(status action.Status, completedType string, blockedType string, failedType string) (events.Status, string) {
	switch status {
	case action.StatusBlocked:
		return events.StatusNeedsHuman, blockedType
	case action.StatusFailed:
		return events.StatusNeedsHuman, failedType
	default:
		return events.StatusCompleted, completedType
	}
}

func isTerminalReconciledActionStatus(status action.Status) bool {
	switch status {
	case action.StatusSucceeded, action.StatusBlocked, action.StatusFailed, action.StatusCanceled, action.StatusSuperseded:
		return true
	default:
		return false
	}
}
