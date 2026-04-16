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

func TestReduceAttemptValidationStartedAdvancesWithinValidating(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus: ProposalRepoChangeQueued,
		AttemptState:   AttemptStatePatchGenerated,
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandValidationStarted),
		CommandID:   "cmd-validation-started",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"sandbox_namespace": "rsi-platform",
			"sandbox_job_name":  "sandbox-job-1",
		},
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextPhase != AttemptPhaseValidating {
		t.Fatalf("expected validating phase, got %s", decision.NextPhase)
	}
	if len(decision.Effects) != 1 || decision.Effects[0].Kind != EffectObserveWorkspaceValidation {
		t.Fatalf("expected observe_workspace_validation effect, got %+v", decision.Effects)
	}
	if len(decision.AllowedAttemptNext) == 0 || decision.AllowedAttemptNext[0] != AttemptStatePatchGenerated {
		t.Fatalf("expected patch_generated attempt state to be allowed, got %+v", decision.AllowedAttemptNext)
	}
}

func TestReduceAttemptRunnerProgressCommandsStayInImplementingPhase(t *testing.T) {
	commands := []AttemptPhaseCommandKind{
		CommandWorkspaceMetadataSynced,
		CommandWorkspaceToolValidationStarted,
		CommandWorkspaceToolValidationCompleted,
		CommandWorkspaceToolValidationFailed,
		CommandAttemptRunnerStarted,
		CommandAttemptRunnerCompleted,
	}
	for _, kind := range commands {
		decision := ReduceAttempt(AttemptSnapshot{
			ProposalStatus:       ProposalRepoChangeRunning,
			AttemptState:         AttemptStatePatchPlan,
			CurrentOperationKind: "implement_attempt",
		}, CommandEnvelope{
			MachineKind: MachineAttempt,
			CommandKind: string(kind),
			CommandID:   "cmd-runner-progress-" + string(kind),
			OccurredAt:  time.Now().UTC(),
		})
		if decision.DecisionKind != DecisionAdvance {
			t.Fatalf("expected advance for %s, got %+v", kind, decision)
		}
		if decision.NextPhase != AttemptPhaseImplementing {
			t.Fatalf("expected implementing phase for %s, got %s", kind, decision.NextPhase)
		}
		if len(decision.AllowedAttemptNext) != 1 || decision.AllowedAttemptNext[0] != AttemptStatePatchPlan {
			t.Fatalf("expected attempt state to remain patch_plan for %s, got %+v", kind, decision.AllowedAttemptNext)
		}
	}
}

func TestReduceAttemptWorkspaceMetadataSyncAllowedDuringValidating(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus:       ProposalRepoChangeRunning,
		AttemptState:         AttemptStateValidationRunning,
		CurrentOperationKind: "workspace_validate",
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandWorkspaceMetadataSynced),
		CommandID:   "cmd-workspace-metadata-sync",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextPhase != AttemptPhaseValidating {
		t.Fatalf("expected validating phase, got %s", decision.NextPhase)
	}
	if len(decision.AllowedAttemptNext) != 1 || decision.AllowedAttemptNext[0] != AttemptStateValidationRunning {
		t.Fatalf("expected validation_running attempt state, got %+v", decision.AllowedAttemptNext)
	}
}

func TestReduceAttemptPROpenFailedReviewAdvancesToTerminal(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus:       ProposalValidationPending,
		AttemptState:         AttemptStateValidationRunning,
		CurrentOperationKind: "pr_open",
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandPROpenFailedReview),
		CommandID:   "cmd-pr-open-failed-review",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextPhase != AttemptPhaseTerminal {
		t.Fatalf("expected terminal phase, got %s", decision.NextPhase)
	}
	if len(decision.AllowedAttemptNext) != 1 || decision.AllowedAttemptNext[0] != AttemptStateNeedsReview {
		t.Fatalf("expected needs_review attempt state, got %+v", decision.AllowedAttemptNext)
	}
	if len(decision.AllowedProposalNext) != 1 || decision.AllowedProposalNext[0] != ProposalPendingReview {
		t.Fatalf("expected pending_review proposal state, got %+v", decision.AllowedProposalNext)
	}
}

func TestReduceAttemptPROpenFailedRetryableAdvancesToRetryDecision(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus:       ProposalValidationPending,
		AttemptState:         AttemptStateValidationRunning,
		CurrentOperationKind: "pr_open",
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandPROpenFailedRetryable),
		CommandID:   "cmd-pr-open-failed-retryable",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextPhase != AttemptPhaseRetryDeciding {
		t.Fatalf("expected retry_deciding phase, got %s", decision.NextPhase)
	}
	if len(decision.AllowedAttemptNext) != 1 || decision.AllowedAttemptNext[0] != AttemptStateNeedsReview {
		t.Fatalf("expected needs_review attempt state, got %+v", decision.AllowedAttemptNext)
	}
}

func TestReduceAttemptOverlayActivatedAdvancesToTerminal(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus:       ProposalApproved,
		AttemptState:         AttemptStateOverlayPlan,
		CurrentOperationKind: "implement_attempt",
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandOverlayActivated),
		CommandID:   "cmd-overlay-activated",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextPhase != AttemptPhaseTerminal {
		t.Fatalf("expected terminal phase, got %s", decision.NextPhase)
	}
	if len(decision.AllowedAttemptNext) != 1 || decision.AllowedAttemptNext[0] != AttemptStateOverlayActive {
		t.Fatalf("expected overlay_active attempt state, got %+v", decision.AllowedAttemptNext)
	}
}

func TestReduceAttemptWorkspaceFailedRetryableAdvancesToRetryDecision(t *testing.T) {
	decision := ReduceAttempt(AttemptSnapshot{
		ProposalStatus:       ProposalRepoChangeQueued,
		AttemptState:         AttemptStatePatchPlan,
		CurrentOperationKind: "workspace_open",
	}, CommandEnvelope{
		MachineKind: MachineAttempt,
		CommandKind: string(CommandWorkspaceFailedRetryable),
		CommandID:   "cmd-workspace-failed-retryable",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextPhase != AttemptPhaseRetryDeciding {
		t.Fatalf("expected retry_deciding phase, got %s", decision.NextPhase)
	}
	if len(decision.AllowedAttemptNext) == 0 || decision.AllowedAttemptNext[0] != AttemptStateSandboxFailed {
		t.Fatalf("expected sandbox_failed attempt state to be allowed, got %+v", decision.AllowedAttemptNext)
	}
}
