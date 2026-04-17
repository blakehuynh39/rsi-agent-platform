package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func ReviewProposalForTesting(store interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
	ListProposals() []review.Proposal
}, proposalID string, decision review.ProposalReview) (review.Proposal, error) {
	commandKind, err := proposalCommandKindForDecision(decision.Decision)
	if err != nil {
		return review.Proposal{}, err
	}
	if decision.Scope == "" {
		decision.Scope = review.FeedbackScopeLine
	}
	decision.IdempotencyKey = firstNonEmpty(strings.TrimSpace(decision.IdempotencyKey), proposalDecisionIdempotencyKey(proposalID, decision.Decision, decision.Scope))
	commandID := fmt.Sprintf("cmd-proposal-review:%s", decision.IdempotencyKey)
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposalID,
		CommandKind: string(commandKind),
		CommandID:   commandID,
		Actor:       firstNonEmpty(decision.ReviewerID, "system"),
		OccurredAt:  firstNonZeroTime(&decision.CreatedAt, time.Now().UTC()),
		Payload: map[string]any{
			"idempotency_key": decision.IdempotencyKey,
			"rationale":       decision.Rationale,
			"reviewer_id":     decision.ReviewerID,
			"failure_class":   decision.FailureClass,
			"failure_classes": append([]string(nil), decision.FailureClasses...),
			"scope":           string(decision.Scope),
		},
	}); err != nil {
		return review.Proposal{}, err
	}
	for _, proposal := range store.ListProposals() {
		if proposal.ID == proposalID {
			return proposal, nil
		}
	}
	return review.Proposal{}, errors.New("proposal not found")
}

func AdvanceProposalToFailedValidationForTesting(target interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
	ListProposals() []review.Proposal
	GetChangeAttempt(string) (improvement.ChangeAttempt, bool)
}, proposalID string, occurredAt time.Time) (review.Proposal, improvement.ChangeAttempt, error) {
	now := occurredAt.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	findProposal := func() (review.Proposal, bool) {
		for _, proposal := range target.ListProposals() {
			if proposal.ID == proposalID {
				return proposal, true
			}
		}
		return review.Proposal{}, false
	}

	proposal, ok := findProposal()
	if !ok {
		return review.Proposal{}, improvement.ChangeAttempt{}, errors.New("proposal not found")
	}

	receipt, err := target.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposalID,
		CommandKind: string(transition.CommandProposalMarkRepoChangeQueued),
		CommandID:   fmt.Sprintf("cmd-proposal-status:%s:%02d", proposalID, 1),
		Actor:       "tester",
		OccurredAt:  now,
	})
	if err != nil {
		return review.Proposal{}, improvement.ChangeAttempt{}, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return review.Proposal{}, improvement.ChangeAttempt{}, fmt.Errorf("command %s rejected: %s", transition.CommandProposalMarkRepoChangeQueued, receipt.Reason)
	}

	proposal, ok = findProposal()
	if !ok {
		return review.Proposal{}, improvement.ChangeAttempt{}, errors.New("proposal not found after validation setup")
	}
	attemptID := strings.TrimSpace(proposal.CurrentAttemptID)
	if attemptID == "" {
		return review.Proposal{}, improvement.ChangeAttempt{}, errors.New("proposal current attempt not found")
	}
	attempt, ok := target.GetChangeAttempt(attemptID)
	if !ok {
		return review.Proposal{}, improvement.ChangeAttempt{}, fmt.Errorf("change attempt %s not found", attemptID)
	}
	for idx, command := range []transition.CommandEnvelope{
		{
			MachineKind: transition.MachineAttempt,
			AggregateID: attempt.ID,
			CommandKind: string(transition.CommandWorkspaceReady),
			CommandID:   fmt.Sprintf("cmd-attempt-status:%s:%02d", attempt.ID, 1),
			Actor:       "tester",
			OccurredAt:  now.Add(3 * time.Millisecond),
			Payload: map[string]any{
				"workspace_id":        "workspace-" + attempt.ID,
				"repo":                "rsi-agent-platform",
				"base_ref":            "main",
				"branch_name":         attempt.BranchName,
				"workspace_namespace": "rsi-platform",
				"workspace_job_name":  "workspace-job-" + attempt.ID,
			},
		},
		{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposalID,
			CommandKind: string(transition.CommandProposalMarkRepoChangeRunning),
			CommandID:   fmt.Sprintf("cmd-proposal-status:%s:%02d", proposalID, 2),
			Actor:       "tester",
			OccurredAt:  now.Add(4 * time.Millisecond),
		},
		{
			MachineKind: transition.MachineAttempt,
			AggregateID: attempt.ID,
			CommandKind: string(transition.CommandImplementationCompleted),
			CommandID:   fmt.Sprintf("cmd-attempt-status:%s:%02d", attempt.ID, 2),
			Actor:       "tester",
			OccurredAt:  now.Add(5 * time.Millisecond),
			Payload: map[string]any{
				"change_plan":     "Implement the approved remediation.",
				"validation_plan": "Run governed validation.",
				"diff_summary":    "formal failed-validation setup",
				"changed_files":   []string{"internal/store/commands.go"},
			},
		},
		{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposalID,
			CommandKind: string(transition.CommandProposalMarkValidationPending),
			CommandID:   fmt.Sprintf("cmd-proposal-status:%s:%02d", proposalID, 3),
			Actor:       "tester",
			OccurredAt:  now.Add(6 * time.Millisecond),
		},
		{
			MachineKind: transition.MachineAttempt,
			AggregateID: attempt.ID,
			CommandKind: string(transition.CommandValidationFailedRetryable),
			CommandID:   fmt.Sprintf("cmd-attempt-status:%s:%02d", attempt.ID, 3),
			Actor:       "tester",
			OccurredAt:  now.Add(7 * time.Millisecond),
			Payload: map[string]any{
				"failure_class":   "sandbox_failure",
				"failure_summary": "validation failed",
				"retry_decision":  "retry",
			},
		},
	} {
		receipt, err := target.SubmitCommand(command)
		if err != nil {
			return review.Proposal{}, improvement.ChangeAttempt{}, err
		}
		if receipt.DecisionKind == transition.DecisionReject {
			return review.Proposal{}, improvement.ChangeAttempt{}, fmt.Errorf("command %s rejected at step %d: %s", command.CommandKind, idx+1, receipt.Reason)
		}
	}
	recordedAttempt, ok := target.GetChangeAttempt(attempt.ID)
	if !ok {
		return review.Proposal{}, improvement.ChangeAttempt{}, fmt.Errorf("change attempt %s not found after validation failure", attempt.ID)
	}

	receipt, err = target.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposalID,
		CommandKind: string(transition.CommandProposalMarkFailedValidation),
		CommandID:   fmt.Sprintf("cmd-proposal-status:%s:04", proposalID),
		Actor:       "tester",
		OccurredAt:  now.Add(8 * time.Millisecond),
		Payload: map[string]any{
			"failure_class":   recordedAttempt.FailureClass,
			"failure_summary": recordedAttempt.FailureSummary,
			"retry_decision":  recordedAttempt.RetryDecision,
		},
	})
	if err != nil {
		return review.Proposal{}, improvement.ChangeAttempt{}, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return review.Proposal{}, improvement.ChangeAttempt{}, fmt.Errorf("command %s rejected: %s", transition.CommandProposalMarkFailedValidation, receipt.Reason)
	}

	proposal, ok = findProposal()
	if !ok {
		return review.Proposal{}, improvement.ChangeAttempt{}, errors.New("proposal not found after failed validation")
	}
	return proposal, recordedAttempt, nil
}
