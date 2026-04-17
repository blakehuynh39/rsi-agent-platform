package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

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
	if !needsStatus && !needsEvent {
		return trace, false, nil
	}

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
	return updated, true, nil
}

func reconciledTraceProjection(workflow Workflow) (events.Status, string, string, bool) {
	switch strings.TrimSpace(workflow.Status) {
	case "failed":
		return events.StatusFailed, "workflow.failed", firstNonEmpty(workflow.FailureSummary, workflow.LastError, "workflow failed"), true
	case "completed":
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
