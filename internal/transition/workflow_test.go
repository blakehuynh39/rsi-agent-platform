package transition

import (
	"testing"
	"time"
)

func TestReduceWorkflowRejectsIllegalCombination(t *testing.T) {
	decision := ReduceWorkflow(WorkflowSnapshot{
		State: WorkflowStateQueued,
	}, CommandEnvelope{
		MachineKind: MachineWorkflow,
		CommandKind: string(CommandReplyPosted),
		CommandID:   "cmd-illegal",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionReject {
		t.Fatalf("expected reject, got %+v", decision)
	}
}

func TestReduceWorkflowRunnerCompletedQueuesReply(t *testing.T) {
	decision := ReduceWorkflow(WorkflowSnapshot{
		State: WorkflowStateReasoning,
	}, CommandEnvelope{
		MachineKind: MachineWorkflow,
		CommandKind: string(CommandRunnerCompleted),
		CommandID:   "cmd-runner",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != WorkflowStateReplyPending {
		t.Fatalf("expected reply_pending, got %s", decision.NextState)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectPostSlackReply {
		t.Fatalf("expected post_slack_reply effect, got %+v", decision.Effects)
	}
}

func TestReduceWorkflowReplyPostedCompletesAndQueuesFollowOnEvalCommand(t *testing.T) {
	decision := ReduceWorkflow(WorkflowSnapshot{
		State:   WorkflowStateReplyPending,
		TraceID: "trace-1",
	}, CommandEnvelope{
		MachineKind: MachineWorkflow,
		CommandKind: string(CommandReplyPosted),
		CommandID:   "cmd-reply",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != WorkflowStateCompleted {
		t.Fatalf("expected completed, got %s", decision.NextState)
	}
	if len(decision.Commands) != 1 {
		t.Fatalf("expected one follow-on command, got %+v", decision.Commands)
	}
	if decision.Commands[0].MachineKind != MachineProblemLine || decision.Commands[0].AggregateID != "trace-1" || decision.Commands[0].CommandKind != string(CommandProblemLineEvaluateTrace) {
		t.Fatalf("expected follow-on problem-line evaluation command, got %+v", decision.Commands[0])
	}
}

func TestWorkflowTransitionTableExplicitForKnownStates(t *testing.T) {
	states := []WorkflowStateKind{
		WorkflowStateQueued,
		WorkflowStateCollectingContext,
		WorkflowStateWaitingOnActions,
		WorkflowStateReasoning,
		WorkflowStateReplyPending,
		WorkflowStateNeedsHuman,
		WorkflowStateCompleted,
		WorkflowStateFailed,
		WorkflowStateSuperseded,
	}
	commands := []WorkflowCommandKind{
		CommandWorkflowStarted,
		CommandContextActionsQueued,
		CommandContextSkipped,
		CommandContextCompleted,
		CommandRunnerCompleted,
		CommandRunnerCompletedNoReply,
		CommandReplyPosted,
		CommandWorkflowBlocked,
		CommandWorkflowFailed,
		CommandWorkflowSuperseded,
	}
	for _, state := range states {
		for _, command := range commands {
			decision := ReduceWorkflow(WorkflowSnapshot{State: state}, CommandEnvelope{
				MachineKind: MachineWorkflow,
				CommandKind: string(command),
				CommandID:   "cmd-coverage",
				OccurredAt:  time.Now().UTC(),
			})
			if decision.DecisionKind != DecisionAdvance && decision.DecisionKind != DecisionReject {
				t.Fatalf("state=%s command=%s returned unsupported decision kind %+v", state, command, decision)
			}
		}
	}
}
