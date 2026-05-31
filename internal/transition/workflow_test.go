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
		State: WorkflowStateExecuting,
	}, CommandEnvelope{
		MachineKind: MachineWorkflow,
		CommandKind: string(CommandWorkflowExecutionCompleted),
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

func TestReduceWorkflowContextSkippedIgnoresLegacyReadHeavyExecutionStrategy(t *testing.T) {
	decision := ReduceWorkflow(WorkflowSnapshot{
		State: WorkflowStateCollectingContext,
	}, CommandEnvelope{
		MachineKind: MachineWorkflow,
		CommandKind: string(CommandContextSkipped),
		CommandID:   "cmd-context-skipped-read-heavy",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"execution_strategy": "read_heavy_slack_qna",
		},
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != WorkflowStateExecuting {
		t.Fatalf("expected executing, got %s", decision.NextState)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectInvokeRunner {
		t.Fatalf("expected invoke_runner effect even with legacy read-heavy execution_strategy, got %+v", decision.Effects)
	}
}

func TestReduceWorkflowRunnerCompletedPartialQueuesReply(t *testing.T) {
	decision := ReduceWorkflow(WorkflowSnapshot{
		State: WorkflowStateExecuting,
	}, CommandEnvelope{
		MachineKind: MachineWorkflow,
		CommandKind: string(CommandWorkflowExecutionCompletedPartial),
		CommandID:   "cmd-runner-partial",
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

func TestReduceWorkflowReplyPostedPartialCompletesAndQueuesFollowOnEvalCommand(t *testing.T) {
	decision := ReduceWorkflow(WorkflowSnapshot{
		State:   WorkflowStateReplyPending,
		TraceID: "trace-1",
	}, CommandEnvelope{
		MachineKind: MachineWorkflow,
		CommandKind: string(CommandReplyPostedPartial),
		CommandID:   "cmd-reply-partial",
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

func TestWorkflowTransitionHelpersReturnPartialCommands(t *testing.T) {
	if got := WorkflowExecutionCompletionCommand("partial", true); got != CommandWorkflowExecutionCompletedPartial {
		t.Fatalf("expected partial reply runner command, got %s", got)
	}
	if got := WorkflowExecutionCompletionCommand("partial", false); got != CommandWorkflowExecutionCompletedPartialNoReply {
		t.Fatalf("expected partial no-reply runner command, got %s", got)
	}
	if got := WorkflowReplyPostedCommand("partial"); got != CommandReplyPostedPartial {
		t.Fatalf("expected partial reply-posted command, got %s", got)
	}
	if got := WorkflowCommandLastVerdict(CommandReplyPostedPartial); got != "partial" {
		t.Fatalf("expected partial last verdict, got %q", got)
	}
}

func TestWorkflowTransitionTableExplicitForKnownStates(t *testing.T) {
	states := []WorkflowStateKind{
		WorkflowStateQueued,
		WorkflowStateCollectingContext,
		WorkflowStateWaitingOnActions,
		WorkflowStateExecuting,
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
		CommandWorkflowExecutionCompleted,
		CommandWorkflowExecutionCompletedPartial,
		CommandWorkflowExecutionCompletedNoReply,
		CommandWorkflowExecutionCompletedPartialNoReply,
		CommandReplyPosted,
		CommandReplyPostedPartial,
		CommandWorkflowExecutionNeedsHuman,
		CommandWorkflowExecutionFailed,
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
