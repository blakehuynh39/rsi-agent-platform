package transition

import (
	"testing"
	"time"
)

func TestReduceAttemptRejectsIllegalCombination(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus: ProposalApproved,
		AttemptState:   AttemptStatePatchPlan,
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandValidationCompleted),
		CommandID:   "cmd-1",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionReject {
		t.Fatalf("expected reject, got %+v", decision)
	}
}

func TestReduceAttemptWorkspaceReadyAdvances(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus: ProposalRepoChangeQueued,
		AttemptState:   AttemptStatePatchPlan,
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandWorkspaceReady),
		CommandID:   "cmd-2",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextPhase != AttemptPhaseImplementing {
		t.Fatalf("expected implementing phase, got %s", decision.NextPhase)
	}
	if len(decision.AllowedProposalNext) == 0 || decision.AllowedProposalNext[0] != ProposalRepoChangeRunning {
		t.Fatalf("expected proposal repo_change_running to be allowed, got %+v", decision.AllowedProposalNext)
	}
}

func TestReduceAttemptValidationCompletedAdvancesToPROpen(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus: ProposalRepoChangeRunning,
		AttemptState:   AttemptStatePatchGenerated,
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandValidationCompleted),
		CommandID:   "cmd-3",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextPhase != AttemptPhasePROpen {
		t.Fatalf("expected pr_open phase, got %s", decision.NextPhase)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectOpenDraftPR {
		t.Fatalf("expected open_draft_pr effect, got %+v", decision.Effects)
	}
}
