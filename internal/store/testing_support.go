package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
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

func SeedChangeAttemptForTesting(target any, item improvement.ChangeAttempt) (improvement.ChangeAttempt, error) {
	switch store := target.(type) {
	case *MemoryStore:
		store.mu.Lock()
		defer store.mu.Unlock()
		return store.upsertChangeAttemptLocked(item)
	case *PostgresStore:
		return store.upsertChangeAttemptDirect(item)
	default:
		return improvement.ChangeAttempt{}, fmt.Errorf("unsupported test store %T", target)
	}
}

func CreateDerivedTraceForTesting(target any, req DerivedTraceRequest) (events.Trace, Workflow, error) {
	switch store := target.(type) {
	case *MemoryStore:
		store.mu.Lock()
		defer store.mu.Unlock()
		return store.createDerivedTraceLocked(req)
	case *PostgresStore:
		return store.createDerivedTraceDirect(req)
	default:
		return events.Trace{}, Workflow{}, fmt.Errorf("unsupported test store %T", target)
	}
}

func SeedAttemptWorkspaceForTesting(target any, item improvement.AttemptWorkspace) (improvement.AttemptWorkspace, error) {
	switch store := target.(type) {
	case *MemoryStore:
		store.mu.Lock()
		defer store.mu.Unlock()
		return store.upsertAttemptWorkspaceLocked(item)
	case *PostgresStore:
		return store.upsertAttemptWorkspaceDirect(item)
	default:
		return improvement.AttemptWorkspace{}, fmt.Errorf("unsupported test store %T", target)
	}
}

func SeedRepoChangeJobForTesting(target any, job improvement.RepoChangeJob) (improvement.RepoChangeJob, error) {
	switch store := target.(type) {
	case *MemoryStore:
		store.mu.Lock()
		defer store.mu.Unlock()
		return store.upsertRepoChangeJobLocked(job)
	case *PostgresStore:
		return store.upsertRepoChangeJobDirect(job)
	default:
		return improvement.RepoChangeJob{}, fmt.Errorf("unsupported test store %T", target)
	}
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

	commands := []transition.ProposalLineCommandKind{
		transition.CommandProposalMarkRepoChangeQueued,
		transition.CommandProposalMarkRepoChangeRunning,
		transition.CommandProposalMarkValidationPending,
	}
	for idx, kind := range commands {
		receipt, err := target.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposalID,
			CommandKind: string(kind),
			CommandID:   fmt.Sprintf("cmd-proposal-status:%s:%02d", proposalID, idx+1),
			Actor:       "tester",
			OccurredAt:  now.Add(time.Duration(idx) * time.Millisecond),
		})
		if err != nil {
			return review.Proposal{}, improvement.ChangeAttempt{}, err
		}
		if receipt.DecisionKind == transition.DecisionReject {
			return review.Proposal{}, improvement.ChangeAttempt{}, fmt.Errorf("command %s rejected: %s", kind, receipt.Reason)
		}
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
	attempt.State = improvement.AttemptStateSandboxFailed
	attempt.FailureClass = "sandbox_failure"
	attempt.FailureSummary = "validation failed"
	attempt.RetryDecision = "retry"
	attempt.UpdatedAt = now.Add(3 * time.Millisecond)
	recordedAttempt, err := SeedChangeAttemptForTesting(target, attempt)
	if err != nil {
		return review.Proposal{}, improvement.ChangeAttempt{}, err
	}

	receipt, err := target.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposalID,
		CommandKind: string(transition.CommandProposalMarkFailedValidation),
		CommandID:   fmt.Sprintf("cmd-proposal-status:%s:04", proposalID),
		Actor:       "tester",
		OccurredAt:  now.Add(4 * time.Millisecond),
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
