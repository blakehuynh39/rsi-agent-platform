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

func TestReduceQuestionRunAlignmentRequiredQueuesLedgerRefresh(t *testing.T) {
	decision := ReduceQuestionRun(QuestionRunSnapshot{State: QuestionRunStateCompilingSpec}, CommandEnvelope{
		MachineKind: MachineQuestionRun,
		CommandKind: string(CommandInvestigationSpecBuilt),
		CommandID:   "cmd-question-run-spec",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"alignment_required": true,
		},
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != QuestionRunStateRefreshingAlignmentLedger {
		t.Fatalf("expected refreshing_alignment_ledger, got %s", decision.NextState)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectRefreshAlignmentLedger {
		t.Fatalf("expected refresh_alignment_ledger effect, got %+v", decision.Effects)
	}
}

func TestReduceQuestionRunSeedEvidenceWithOpenGapsQueuesExpansion(t *testing.T) {
	decision := ReduceQuestionRun(QuestionRunSnapshot{State: QuestionRunStateCollectingSeedEvidence}, CommandEnvelope{
		MachineKind: MachineQuestionRun,
		CommandKind: string(CommandSeedEvidenceCollected),
		CommandID:   "cmd-question-run-seed",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"should_expand": true,
		},
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != QuestionRunStateExpandingEvidence {
		t.Fatalf("expected expanding_evidence, got %s", decision.NextState)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectExpandEvidence {
		t.Fatalf("expected expand_evidence effect, got %+v", decision.Effects)
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
