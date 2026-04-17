package store

import (
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func (s *MemoryStore) applyWorkflowCommandMetadataLocked(workflowID string, command transition.CommandEnvelope) (Workflow, error) {
	for i := range s.workflows {
		if s.workflows[i].ID != workflowID {
			continue
		}
		if verdict := transition.WorkflowCommandLastVerdict(transition.WorkflowCommandKind(command.CommandKind)); verdict != "" {
			s.workflows[i].LastVerdict = verdict
		}
		if failureClass := strings.TrimSpace(stringFromCommand(command, "failure_class")); failureClass != "" {
			s.workflows[i].FailureClass = failureClass
		}
		if failureSummary := strings.TrimSpace(stringFromCommand(command, "failure_summary")); failureSummary != "" {
			s.workflows[i].FailureSummary = failureSummary
		} else if transition.WorkflowCommandKind(command.CommandKind) == transition.CommandWorkflowFailed || transition.WorkflowCommandKind(command.CommandKind) == transition.CommandWorkflowBlocked {
			s.workflows[i].FailureSummary = firstNonEmpty(s.workflows[i].FailureSummary, workflowLastErrorForCommand(command))
		}
		if retryDecision := strings.TrimSpace(stringFromCommand(command, "retry_decision")); retryDecision != "" {
			s.workflows[i].RetryDecision = retryDecision
		}
		if _, ok := command.Payload["retry_after"]; ok {
			s.workflows[i].RetryAfter = optionalTimeFromCommand(command, "retry_after")
		}
		if _, ok := command.Payload["runner_diagnostics"]; ok {
			s.workflows[i].RunnerDiagnostics = anyMapFromCommand(command, "runner_diagnostics")
		}
		if _, ok := command.Payload["repair_attempted"]; ok {
			s.workflows[i].RepairAttempted = boolFromCommand(command, "repair_attempted", s.workflows[i].RepairAttempted)
		}
		if _, ok := command.Payload["repair_succeeded"]; ok {
			s.workflows[i].RepairSucceeded = boolFromCommand(command, "repair_succeeded", s.workflows[i].RepairSucceeded)
		}
		s.workflows[i].UpdatedAt = command.OccurredAt
		return s.workflows[i], nil
	}
	return Workflow{}, fmt.Errorf("workflow %s not found", workflowID)
}

func (s *MemoryStore) appendWorkflowLineFollowOnCommandLocked(bundle *transitionPersistBundle, parent transition.CommandEnvelope, workflow Workflow) {
	caseID := strings.TrimSpace(workflow.CaseID)
	if caseID == "" || bundle == nil {
		return
	}
	switch transition.WorkflowCommandKind(parent.CommandKind) {
	case transition.CommandRunnerCompletedNoReply, transition.CommandRunnerCompletedPartialNoReply, transition.CommandReplyPosted, transition.CommandReplyPostedPartial:
		appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflowLine,
			AggregateID: caseID,
			CommandKind: string(transition.CommandWorkflowLineMarkCompleted),
			CommandID:   fmt.Sprintf("%s:workflow-line:completed", parent.CommandID),
			Actor:       parent.Actor,
			OccurredAt:  parent.OccurredAt,
			Payload: map[string]any{
				"source_workflow_id": workflow.ID,
				"source_trace_id":    workflow.TraceID,
				"line_stop_reason":   string(parent.CommandKind),
			},
		}, "workflow attempt completed and the workflow line can close")
	case transition.CommandWorkflowBlocked:
		appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflowLine,
			AggregateID: caseID,
			CommandKind: string(transition.CommandWorkflowLineNeedsHuman),
			CommandID:   fmt.Sprintf("%s:workflow-line:needs-human", parent.CommandID),
			Actor:       parent.Actor,
			OccurredAt:  parent.OccurredAt,
			Payload: map[string]any{
				"source_workflow_id": workflow.ID,
				"source_trace_id":    workflow.TraceID,
				"failure_class":      workflow.FailureClass,
				"line_stop_reason":   firstNonEmpty(workflow.FailureClass, workflow.LastError, "needs_human"),
			},
		}, "workflow attempt needs human intervention and the workflow line must stop")
	case transition.CommandWorkflowFailed:
		if strings.EqualFold(strings.TrimSpace(workflow.RetryDecision), "auto_retry") {
			payload := map[string]any{
				"source_workflow_id": workflow.ID,
				"source_trace_id":    workflow.TraceID,
				"failure_class":      workflow.FailureClass,
				"failure_summary":    firstNonEmpty(workflow.FailureSummary, workflow.LastError),
				"retry_decision":     workflow.RetryDecision,
				"next_retry_action":  "activate_retry",
			}
			if workflow.RetryAfter != nil {
				payload["retry_after"] = *workflow.RetryAfter
			}
			appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
				MachineKind: transition.MachineWorkflowLine,
				AggregateID: caseID,
				CommandKind: string(transition.CommandWorkflowLineScheduleRetry),
				CommandID:   fmt.Sprintf("%s:workflow-line:schedule-retry", parent.CommandID),
				Actor:       parent.Actor,
				OccurredAt:  parent.OccurredAt,
				Payload:     payload,
			}, "workflow attempt failed before reply posting and the workflow line scheduled a successor attempt")
			return
		}
		appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflowLine,
			AggregateID: caseID,
			CommandKind: string(transition.CommandWorkflowLineNeedsHuman),
			CommandID:   fmt.Sprintf("%s:workflow-line:failed-needs-human", parent.CommandID),
			Actor:       parent.Actor,
			OccurredAt:  parent.OccurredAt,
			Payload: map[string]any{
				"source_workflow_id": workflow.ID,
				"source_trace_id":    workflow.TraceID,
				"failure_class":      workflow.FailureClass,
				"line_stop_reason":   firstNonEmpty(workflow.FailureClass, workflow.LastError, "failed"),
			},
		}, "workflow attempt failed terminally and the workflow line moved to needs human")
	}
}
