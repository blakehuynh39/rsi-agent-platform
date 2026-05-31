package transition

import (
	"testing"
	"time"
)

func TestReduceProblemLineFeedbackAdvancesWithTrace(t *testing.T) {
	decision := ReduceProblemLine(ProblemLineSnapshot{
		State:       ProblemLineStateObserving,
		TraceExists: true,
	}, CommandEnvelope{
		MachineKind: MachineProblemLine,
		CommandKind: string(CommandProblemLineRecordFeedback),
		CommandID:   "cmd-problem-feedback",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if len(decision.Events) != 1 || decision.Events[0].Kind != "problem_line_feedback_recorded" {
		t.Fatalf("unexpected events %+v", decision.Events)
	}
}

func TestReduceProblemLineRejectsFeedbackWithoutTrace(t *testing.T) {
	decision := ReduceProblemLine(ProblemLineSnapshot{
		State:       ProblemLineStateObserving,
		TraceExists: false,
	}, CommandEnvelope{
		MachineKind: MachineProblemLine,
		CommandKind: string(CommandProblemLineRecordFeedback),
		CommandID:   "cmd-problem-feedback-missing-trace",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionReject {
		t.Fatalf("expected reject, got %+v", decision)
	}
}

func TestReduceProblemLineReplayScheduleAdvances(t *testing.T) {
	decision := ReduceProblemLine(ProblemLineSnapshot{
		State:       ProblemLineStateObserving,
		TraceExists: true,
	}, CommandEnvelope{
		MachineKind: MachineProblemLine,
		CommandKind: string(CommandProblemLineScheduleReplay),
		CommandID:   "cmd-problem-replay",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != ProblemLineStateObserving {
		t.Fatalf("expected observing state, got %s", decision.NextState)
	}
	if len(decision.Events) != 1 || decision.Events[0].Kind != "problem_line_replay_scheduled" {
		t.Fatalf("unexpected events %+v", decision.Events)
	}
	if len(decision.Commands) != 1 || decision.Commands[0].MachineKind != MachineProblemLine || decision.Commands[0].CommandKind != string(CommandProblemLineEvaluateTrace) {
		t.Fatalf("unexpected follow-on commands %+v", decision.Commands)
	}
}

func TestReduceProblemLineEvaluateTraceQueuesRunnerEffect(t *testing.T) {
	decision := ReduceProblemLine(ProblemLineSnapshot{
		State:       ProblemLineStateObserving,
		TraceExists: true,
	}, CommandEnvelope{
		MachineKind: MachineProblemLine,
		CommandKind: string(CommandProblemLineEvaluateTrace),
		CommandID:   "cmd-problem-evaluate",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectInvokeRunner || decision.Effects[0].Status != EffectQueued {
		t.Fatalf("unexpected effects %+v", decision.Effects)
	}
}
