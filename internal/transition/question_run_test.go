package transition

import (
	"testing"
	"time"
)

func TestReduceQuestionRunQueuedStartsCompilation(t *testing.T) {
	decision := ReduceQuestionRun(QuestionRunSnapshot{State: QuestionRunStateQueued}, CommandEnvelope{
		MachineKind: MachineQuestionRun,
		CommandKind: string(CommandQuestionRunStarted),
		CommandID:   "cmd-question-run-start",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != QuestionRunStateCompilingSpec {
		t.Fatalf("expected compiling_spec, got %s", decision.NextState)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectCompileInvestigationSpec {
		t.Fatalf("expected compile_investigation_spec effect, got %+v", decision.Effects)
	}
}

func TestReduceQuestionRunSpecBuiltQueuesGatherEvidence(t *testing.T) {
	decision := ReduceQuestionRun(QuestionRunSnapshot{State: QuestionRunStateCompilingSpec}, CommandEnvelope{
		MachineKind: MachineQuestionRun,
		CommandKind: string(CommandInvestigationSpecBuilt),
		CommandID:   "cmd-question-run-spec",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != QuestionRunStateGatheringEvidence {
		t.Fatalf("expected gathering_evidence, got %s", decision.NextState)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectGatherEvidence {
		t.Fatalf("expected gather_evidence effect, got %+v", decision.Effects)
	}
}

func TestReduceQuestionRunGatheredEvidenceQueuesReduce(t *testing.T) {
	decision := ReduceQuestionRun(QuestionRunSnapshot{State: QuestionRunStateGatheringEvidence}, CommandEnvelope{
		MachineKind: MachineQuestionRun,
		CommandKind: string(CommandEvidenceGatheredPartial),
		CommandID:   "cmd-question-run-gather",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != QuestionRunStateReducing {
		t.Fatalf("expected reducing, got %s", decision.NextState)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectReduceReply {
		t.Fatalf("expected reduce_reply effect, got %+v", decision.Effects)
	}
}

func TestReduceQuestionRunReplyReducedPartialCompletes(t *testing.T) {
	decision := ReduceQuestionRun(QuestionRunSnapshot{State: QuestionRunStateReducing}, CommandEnvelope{
		MachineKind: MachineQuestionRun,
		CommandKind: string(CommandReplyReducedPartial),
		CommandID:   "cmd-question-run-partial",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != QuestionRunStateCompleted {
		t.Fatalf("expected completed, got %s", decision.NextState)
	}
	if len(decision.Events) != 1 || decision.Events[0].Kind != "question_run_reply_reduced_partial" {
		t.Fatalf("expected partial completion event, got %+v", decision.Events)
	}
}
