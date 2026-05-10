package transition

import (
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
)

func TestReduceActionExecutionRejectsIllegalCombination(t *testing.T) {
	decision := ReduceActionExecution(ActionExecutionSnapshot{
		State: action.StatusQueued,
	}, CommandEnvelope{
		MachineKind: MachineAction,
		CommandKind: string(CommandActionSucceed),
		CommandID:   "cmd-illegal",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionReject {
		t.Fatalf("expected reject, got %+v", decision)
	}
}

func TestReduceActionExecutionQueueAdvancesFromEmptyState(t *testing.T) {
	for _, kind := range []action.Kind{action.KindToolRead, action.KindSlackPost, action.KindSlackReport} {
		decision := ReduceActionExecution(ActionExecutionSnapshot{}, CommandEnvelope{
			MachineKind: MachineAction,
			CommandKind: string(CommandActionQueue),
			CommandID:   "cmd-queue",
			OccurredAt:  time.Now().UTC(),
			Payload: map[string]any{
				"kind": string(kind),
			},
		})
		if decision.DecisionKind != DecisionAdvance {
			t.Fatalf("kind %s: expected advance, got %+v", kind, decision)
		}
		if decision.NextState != action.StatusQueued {
			t.Fatalf("kind %s: expected queued, got %s", kind, decision.NextState)
		}
		if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectInvokeAction {
			t.Fatalf("kind %s: expected invoke_action effect, got %+v", kind, decision.Effects)
		}
	}
}

func TestReduceActionExecutionStartAdvancesToExecuting(t *testing.T) {
	decision := ReduceActionExecution(ActionExecutionSnapshot{
		State: action.StatusQueued,
	}, CommandEnvelope{
		MachineKind: MachineAction,
		CommandKind: string(CommandActionStart),
		CommandID:   "cmd-start",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != action.StatusExecuting {
		t.Fatalf("expected executing, got %s", decision.NextState)
	}
}

func TestReduceActionExecutionSuccessAdvancesToSucceeded(t *testing.T) {
	decision := ReduceActionExecution(ActionExecutionSnapshot{
		State: action.StatusExecuting,
	}, CommandEnvelope{
		MachineKind: MachineAction,
		CommandKind: string(CommandActionSucceed),
		CommandID:   "cmd-succeed",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != action.StatusSucceeded {
		t.Fatalf("expected succeeded, got %s", decision.NextState)
	}
}

func TestActionExecutionTransitionTableExplicitForKnownStates(t *testing.T) {
	states := []action.Status{
		"",
		action.StatusDrafted,
		action.StatusApproved,
		action.StatusQueued,
		action.StatusExecuting,
		action.StatusSucceeded,
		action.StatusBlocked,
		action.StatusFailed,
		action.StatusCanceled,
		action.StatusSuperseded,
	}
	commands := []ActionExecutionCommandKind{
		CommandActionQueue,
		CommandActionStart,
		CommandActionSucceed,
		CommandActionBlock,
		CommandActionFail,
		CommandActionCancel,
	}
	for _, state := range states {
		for _, command := range commands {
			decision := ReduceActionExecution(ActionExecutionSnapshot{State: state}, CommandEnvelope{
				MachineKind: MachineAction,
				CommandKind: string(command),
				CommandID:   "cmd-coverage",
				OccurredAt:  time.Now().UTC(),
			})
			if decision.DecisionKind != DecisionAdvance && decision.DecisionKind != DecisionReject && decision.DecisionKind != DecisionNoop {
				t.Fatalf("state=%s command=%s returned unsupported decision kind %+v", state, command, decision)
			}
		}
	}
}
