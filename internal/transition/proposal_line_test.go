package transition

import (
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/review"
)

func TestReduceProposalLineMarkRepoChangeRunningAdvances(t *testing.T) {
	decision := ReduceProposalLine(ProposalLineSnapshot{
		State:            review.ProposalRepoChangeQueued,
		InterventionKind: review.InterventionRepoChange,
	}, CommandEnvelope{
		MachineKind: MachineProposalLine,
		CommandKind: string(CommandProposalMarkRepoChangeRunning),
		CommandID:   "cmd-proposal-running",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != review.ProposalRepoChangeRunning {
		t.Fatalf("expected repo_change_running next state, got %s", decision.NextState)
	}
}

func TestReduceProposalLineApproveQueuesInternalResumeCommand(t *testing.T) {
	decision := ReduceProposalLine(ProposalLineSnapshot{
		State:            review.ProposalPendingReview,
		InterventionKind: review.InterventionRepoChange,
	}, CommandEnvelope{
		MachineKind: MachineProposalLine,
		AggregateID: "proposal-1",
		CommandKind: string(CommandProposalApproveIntervention),
		CommandID:   "cmd-proposal-approve",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != review.ProposalApproved {
		t.Fatalf("expected approved next state, got %s", decision.NextState)
	}
	if len(decision.Commands) != 1 || decision.Commands[0].CommandKind != string(CommandProposalResumeExecution) {
		t.Fatalf("expected internal resume command, got %+v", decision.Commands)
	}
}

func TestReduceProposalLineRetryRequestQueuesInternalResumeCommand(t *testing.T) {
	decision := ReduceProposalLine(ProposalLineSnapshot{
		State:            review.ProposalFailedValidation,
		InterventionKind: review.InterventionRepoChange,
	}, CommandEnvelope{
		MachineKind: MachineProposalLine,
		AggregateID: "proposal-1",
		CommandKind: string(CommandProposalRetryAttempt),
		CommandID:   "cmd-proposal-retry",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != review.ProposalRepoChangeQueued {
		t.Fatalf("expected repo_change_queued next state, got %s", decision.NextState)
	}
	if len(decision.Commands) != 1 || decision.Commands[0].CommandKind != string(CommandProposalResumeExecution) {
		t.Fatalf("expected internal resume command, got %+v", decision.Commands)
	}
}

func TestReduceProposalLineMarkValidationPendingAdvances(t *testing.T) {
	decision := ReduceProposalLine(ProposalLineSnapshot{
		State:            review.ProposalRepoChangeRunning,
		InterventionKind: review.InterventionRepoChange,
	}, CommandEnvelope{
		MachineKind: MachineProposalLine,
		CommandKind: string(CommandProposalMarkValidationPending),
		CommandID:   "cmd-proposal-validation-pending",
		OccurredAt:  time.Now().UTC(),
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != review.ProposalValidationPending {
		t.Fatalf("expected validation_pending next state, got %s", decision.NextState)
	}
}
