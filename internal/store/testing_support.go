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
