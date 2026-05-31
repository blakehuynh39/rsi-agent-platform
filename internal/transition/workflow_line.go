package transition

type WorkflowLineStateKind string

const (
	WorkflowLineStateActive         WorkflowLineStateKind = "active"
	WorkflowLineStateRetryScheduled WorkflowLineStateKind = "retry_scheduled"
	WorkflowLineStateNeedsHuman     WorkflowLineStateKind = "needs_human"
	WorkflowLineStateCompleted      WorkflowLineStateKind = "completed"
	WorkflowLineStateSuperseded     WorkflowLineStateKind = "superseded"
)

type WorkflowLineCommandKind string

const (
	CommandWorkflowLineScheduleRetry WorkflowLineCommandKind = "workflow_line_schedule_retry"
	CommandWorkflowLineActivateRetry WorkflowLineCommandKind = "workflow_line_activate_retry"
	CommandWorkflowLineMarkCompleted WorkflowLineCommandKind = "workflow_line_mark_completed"
	CommandWorkflowLineNeedsHuman    WorkflowLineCommandKind = "workflow_line_needs_human"
	CommandWorkflowLineSupersede     WorkflowLineCommandKind = "workflow_line_supersede"
)

type WorkflowLineSnapshot struct {
	State WorkflowLineStateKind `json:"state"`
}

type WorkflowLineDecision struct {
	TransitionDecision
	NextState WorkflowLineStateKind `json:"next_state,omitempty"`
}

func ReduceWorkflowLine(snapshot WorkflowLineSnapshot, command CommandEnvelope) WorkflowLineDecision {
	current := snapshot.State
	if current == "" {
		current = WorkflowLineStateActive
	}
	switch WorkflowLineCommandKind(command.CommandKind) {
	case CommandWorkflowLineScheduleRetry:
		switch current {
		case WorkflowLineStateSuperseded:
			return workflowLineReject(current, "workflow line already superseded")
		default:
			return workflowLineAdvance(WorkflowLineStateRetryScheduled, "workflow line scheduled a successor attempt", "workflow_line_retry_scheduled")
		}
	case CommandWorkflowLineActivateRetry:
		switch current {
		case WorkflowLineStateRetryScheduled:
			return workflowLineAdvance(WorkflowLineStateActive, "workflow line activated the scheduled successor attempt", "workflow_line_retry_activated")
		case WorkflowLineStateSuperseded:
			return workflowLineReject(current, "workflow line already superseded")
		default:
			return workflowLineNoop(current, "workflow line is not waiting on a scheduled retry")
		}
	case CommandWorkflowLineMarkCompleted:
		switch current {
		case WorkflowLineStateSuperseded:
			return workflowLineReject(current, "workflow line already superseded")
		default:
			return workflowLineAdvance(WorkflowLineStateCompleted, "workflow line completed successfully", "workflow_line_completed")
		}
	case CommandWorkflowLineNeedsHuman:
		switch current {
		case WorkflowLineStateSuperseded:
			return workflowLineReject(current, "workflow line already superseded")
		default:
			return workflowLineAdvance(WorkflowLineStateNeedsHuman, "workflow line requires human intervention", "workflow_line_needs_human")
		}
	case CommandWorkflowLineSupersede:
		if current == WorkflowLineStateSuperseded {
			return workflowLineNoop(current, "workflow line already superseded")
		}
		return workflowLineAdvance(WorkflowLineStateSuperseded, "workflow line superseded by a successor line or case transition", "workflow_line_superseded")
	default:
		return workflowLineReject(current, "unsupported workflow line command for current state")
	}
}

func workflowLineAdvance(next WorkflowLineStateKind, reason string, eventKind string) WorkflowLineDecision {
	return WorkflowLineDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: eventKind,
			}},
		},
		NextState: next,
	}
}

func workflowLineNoop(state WorkflowLineStateKind, reason string) WorkflowLineDecision {
	return WorkflowLineDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionNoop,
			Reason:       reason,
		},
		NextState: state,
	}
}

func workflowLineReject(state WorkflowLineStateKind, reason string) WorkflowLineDecision {
	return WorkflowLineDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       reason,
		},
		NextState: state,
	}
}
