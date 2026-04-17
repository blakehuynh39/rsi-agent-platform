package transition

import (
	"testing"
	"time"
)

func TestReduceWorkflowLineScheduleRetryAdvances(t *testing.T) {
	decision := ReduceWorkflowLine(WorkflowLineSnapshot{
		State: WorkflowLineStateActive,
	}, CommandEnvelope{
		MachineKind: MachineWorkflowLine,
		CommandKind: string(CommandWorkflowLineScheduleRetry),
		CommandID:   "cmd-line-retry",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != WorkflowLineStateRetryScheduled {
		t.Fatalf("expected retry_scheduled, got %s", decision.NextState)
	}
}

func TestReduceWorkflowLineActivateRetryNoopsOutsideRetryScheduled(t *testing.T) {
	decision := ReduceWorkflowLine(WorkflowLineSnapshot{
		State: WorkflowLineStateActive,
	}, CommandEnvelope{
		MachineKind: MachineWorkflowLine,
		CommandKind: string(CommandWorkflowLineActivateRetry),
		CommandID:   "cmd-line-activate",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionNoop {
		t.Fatalf("expected noop, got %+v", decision)
	}
	if decision.NextState != WorkflowLineStateActive {
		t.Fatalf("expected active state, got %s", decision.NextState)
	}
}

func TestReduceWorkflowLineNeedsHumanAdvances(t *testing.T) {
	decision := ReduceWorkflowLine(WorkflowLineSnapshot{
		State: WorkflowLineStateRetryScheduled,
	}, CommandEnvelope{
		MachineKind: MachineWorkflowLine,
		CommandKind: string(CommandWorkflowLineNeedsHuman),
		CommandID:   "cmd-line-needs-human",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != WorkflowLineStateNeedsHuman {
		t.Fatalf("expected needs_human, got %s", decision.NextState)
	}
}
