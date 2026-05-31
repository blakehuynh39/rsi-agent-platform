package store

import (
	"errors"
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func (s *MemoryStore) applyWorkflowLineCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	line, ok := s.workflowLines[strings.TrimSpace(command.AggregateID)]
	if !ok {
		return commandApplyResult{}, errors.New("workflow line not found")
	}
	decision := transition.ReduceWorkflowLine(transition.WorkflowLineSnapshot{
		State: workflowLineStateFromStatus(line.Status),
	}, command)
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		var (
			err            error
			createdTraceID string
		)
		switch transition.WorkflowLineCommandKind(command.CommandKind) {
		case transition.CommandWorkflowLineScheduleRetry:
			line, createdTraceID, err = s.scheduleWorkflowLineRetryLocked(line, command, decision.NextState)
		default:
			line, err = s.setWorkflowLineStateLocked(line.CaseID, decision.NextState, command)
		}
		if err != nil {
			return commandApplyResult{}, err
		}
		result := commandApplyResult{
			receipt: buildCommandReceipt(command, decision.TransitionDecision, line.UpdatedAt, line.Version, line.CaseID),
			bundle:  buildCommandBundle(command, decision.TransitionDecision, line.Version),
			traceID: createdTraceID,
		}
		return result, nil
	case transition.DecisionReject, transition.DecisionNoop:
		return commandApplyResult{
			receipt: buildCommandReceipt(command, decision.TransitionDecision, line.UpdatedAt, line.Version, line.CaseID),
		}, nil
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported workflow line decision kind %s", decision.DecisionKind)
	}
}

func (s *MemoryStore) setWorkflowLineStateLocked(caseID string, state transition.WorkflowLineStateKind, command transition.CommandEnvelope) (WorkflowLine, error) {
	line, ok := s.workflowLines[strings.TrimSpace(caseID)]
	if !ok {
		return WorkflowLine{}, errors.New("workflow line not found")
	}
	line.Status = workflowLineStatusFromState(state)
	line.LastFailureClass = firstNonEmpty(stringFromCommand(command, "failure_class"), line.LastFailureClass)
	line.NextRetryAction = stringFromCommand(command, "next_retry_action")
	line.LineStopReason = firstNonEmpty(stringFromCommand(command, "line_stop_reason"), line.LineStopReason)
	if retryAfter := optionalTimeFromCommand(command, "retry_after"); retryAfter != nil {
		line.RetryAfter = retryAfter
	} else if state != transition.WorkflowLineStateRetryScheduled {
		line.RetryAfter = nil
	}
	line.UpdatedAt = command.OccurredAt
	line.Version++
	switch state {
	case transition.WorkflowLineStateCompleted, transition.WorkflowLineStateNeedsHuman, transition.WorkflowLineStateSuperseded:
		completedAt := command.OccurredAt
		line.CompletedAt = &completedAt
	default:
		line.CompletedAt = nil
	}
	if state == transition.WorkflowLineStateCompleted {
		line.NextRetryAction = ""
		line.LineStopReason = ""
		line.RetryAfter = nil
	}
	s.workflowLines[line.CaseID] = line
	return line, nil
}

func (s *MemoryStore) scheduleWorkflowLineRetryLocked(line WorkflowLine, command transition.CommandEnvelope, state transition.WorkflowLineStateKind) (WorkflowLine, string, error) {
	sourceWorkflowID := firstNonEmpty(stringFromCommand(command, "source_workflow_id"), line.CurrentWorkflowID, line.LatestWorkflowID)
	sourceTraceID := firstNonEmpty(stringFromCommand(command, "source_trace_id"))
	if sourceWorkflowID == "" && sourceTraceID == "" {
		return WorkflowLine{}, "", errors.New("workflow line retry requires a source workflow or source trace")
	}

	var (
		parentWorkflow Workflow
		sourceTrace    events.Trace
	)
	if sourceWorkflowID != "" {
		item, ok := findWorkflowByID(s.workflows, sourceWorkflowID)
		if !ok {
			return WorkflowLine{}, "", fmt.Errorf("source workflow %s not found", sourceWorkflowID)
		}
		parentWorkflow = item
	}
	if sourceTraceID == "" {
		sourceTraceID = parentWorkflow.TraceID
	}
	if sourceTraceID != "" {
		item, ok := s.traces[sourceTraceID]
		if !ok {
			return WorkflowLine{}, "", fmt.Errorf("source trace %s not found", sourceTraceID)
		}
		sourceTrace = item
	}
	if parentWorkflow.ID == "" && strings.TrimSpace(sourceTrace.Summary.WorkflowID) != "" {
		item, ok := findWorkflowByID(s.workflows, sourceTrace.Summary.WorkflowID)
		if !ok {
			return WorkflowLine{}, "", fmt.Errorf("parent workflow %s not found", sourceTrace.Summary.WorkflowID)
		}
		parentWorkflow = item
	}
	caseRecord, ok := s.cases[line.CaseID]
	if !ok {
		return WorkflowLine{}, "", fmt.Errorf("case %s not found for workflow line", line.CaseID)
	}
	conv, ok := s.conversations[caseRecord.ConversationID]
	if !ok {
		return WorkflowLine{}, "", fmt.Errorf("conversation %s not found for workflow line", caseRecord.ConversationID)
	}

	createdAt := command.OccurredAt
	traceID := nextID("trace", len(s.traces)+1)
	workflowID := nextID("wf", len(s.workflows)+1)
	triggerEventID := firstNonEmpty(stringFromCommand(command, "trigger_event_id"), sourceTrace.Summary.TriggerEventID)
	ingestionID := firstNonEmpty(stringFromCommand(command, "ingestion_id"), sourceTrace.Summary.IngestionID, parentWorkflow.IngestionID)
	retryAfter := optionalTimeFromCommand(command, "retry_after")
	supersedesTraceID := s.supersedeInFlightTracesLocked(caseRecord.ID, traceID, triggerEventID, createdAt)
	if supersedesTraceID == "" {
		supersedesTraceID = sourceTraceID
	}

	workflow := Workflow{
		ID:               workflowID,
		IngestionID:      ingestionID,
		TraceID:          traceID,
		ConversationID:   caseRecord.ConversationID,
		CaseID:           caseRecord.ID,
		ThreadKey:        firstNonEmpty(parentWorkflow.ThreadKey, conv.ExternalKey),
		Kind:             firstNonEmpty(parentWorkflow.Kind, caseRecord.Kind),
		Intent:           firstNonEmpty(parentWorkflow.Intent, caseRecord.Intent),
		AssignedBot:      firstNonEmpty(parentWorkflow.AssignedBot, caseRecord.AssignedBot),
		ApprovalMode:     firstNonEmpty(parentWorkflow.ApprovalMode, caseRecord.ApprovalMode),
		ResponseMode:     firstNonEmpty(parentWorkflow.ResponseMode, caseRecord.ResponseMode),
		Status:           string(transition.WorkflowStateQueued),
		AttemptNumber:    line.AttemptCount + 1,
		ParentWorkflowID: parentWorkflow.ID,
		FailureClass:     stringFromCommand(command, "failure_class"),
		FailureSummary:   stringFromCommand(command, "failure_summary"),
		RetryDecision:    stringFromCommand(command, "retry_decision"),
		RetryAfter:       retryAfter,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
		Version:          1,
	}
	s.upsertWorkflowLocked(workflow)

	trace := events.Trace{
		Summary: events.TraceSummary{
			TraceID:           traceID,
			IngestionID:       ingestionID,
			WorkflowID:        workflowID,
			ConversationID:    caseRecord.ConversationID,
			CaseID:            caseRecord.ID,
			TriggerEventID:    triggerEventID,
			SupersedesTraceID: supersedesTraceID,
			ThreadKey:         firstNonEmpty(sourceTrace.Summary.ThreadKey, conv.ExternalKey),
			WorkflowKind:      firstNonEmpty(sourceTrace.Summary.WorkflowKind, parentWorkflow.Kind, caseRecord.Kind),
			Status:            events.StatusQueued,
			StartedAt:         createdAt,
			EndedAt:           createdAt,
		},
		Events: []events.TraceEvent{{
			TraceID:        traceID,
			IngestionID:    ingestionID,
			WorkflowID:     workflowID,
			ConversationID: caseRecord.ConversationID,
			CaseID:         caseRecord.ID,
			TriggerEventID: triggerEventID,
			Plane:          "control",
			Service:        firstNonEmpty(command.Actor, "control-plane"),
			Actor:          "workflow-line",
			EventType:      "workflow.retry_queued",
			Status:         events.StatusQueued,
			StartedAt:      createdAt,
			Description:    firstNonEmpty(stringFromCommand(command, "trace_description"), fmt.Sprintf("Queued workflow attempt %d from line %s.", workflow.AttemptNumber, line.CaseID)),
		}},
		Reasoning: []events.ReasoningStep{{
			ID:         nextID("reason", len(s.traces)+1),
			TraceID:    traceID,
			WorkflowID: workflowID,
			StepType:   "workflow_retry_plan",
			Summary:    firstNonEmpty(stringFromCommand(command, "failure_summary"), "Retry the live workflow attempt with a fresh successor trace."),
			Confidence: 0.82,
			Decision:   firstNonEmpty(stringFromCommand(command, "retry_decision"), "auto_retry"),
			CreatedAt:  createdAt,
		}},
	}
	recomputeTraceSummary(&trace)
	s.traces[traceID] = trace

	line.Status = workflowLineStatusFromState(state)
	line.CurrentWorkflowID = workflowID
	line.LatestWorkflowID = workflowID
	line.AttemptCount++
	line.AutoRetryBudgetRemaining = workflowLineRetryBudgetRemaining(line.AttemptCount)
	line.LastFailureClass = firstNonEmpty(stringFromCommand(command, "failure_class"), line.LastFailureClass)
	line.NextRetryAction = firstNonEmpty(stringFromCommand(command, "next_retry_action"), "activate_retry")
	line.RetryAfter = retryAfter
	line.LineStopReason = ""
	line.UpdatedAt = createdAt
	line.CompletedAt = nil
	line.Version++
	s.workflowLines[line.CaseID] = line
	return line, traceID, nil
}
