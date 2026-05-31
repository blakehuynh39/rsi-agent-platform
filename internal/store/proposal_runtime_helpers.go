package store

import (
	"errors"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func (s *MemoryStore) upsertProposalLocked(item review.Proposal) review.Proposal {
	now := time.Now().UTC()
	existing, ok := s.proposals[item.ID]
	if ok {
		if item.CreatedAt.IsZero() {
			item.CreatedAt = existing.CreatedAt
		}
		if len(item.Reviews) == 0 && len(existing.Reviews) > 0 {
			item.Reviews = append([]review.ProposalReview(nil), existing.Reviews...)
		}
		if item.Version <= existing.Version {
			item.Version = existing.Version + 1
		}
	} else if item.CreatedAt.IsZero() {
		item.CreatedAt = now
		if item.Version == 0 {
			item.Version = 1
		}
	} else if item.Version == 0 {
		item.Version = 1
	}
	item.ActiveSlotConsuming = review.ConsumesActiveProposalSlot(item.Status)
	item = normalizeProposalTargetFields(item)
	s.proposals[item.ID] = item
	return item
}

func (s *MemoryStore) upsertChangeAttemptLocked(item improvement.ChangeAttempt) (improvement.ChangeAttempt, error) {
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextID("attempt", len(s.changeAttempts)+1)
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	if existing, ok := s.changeAttempts[item.ID]; ok {
		if item.Version <= existing.Version {
			item.Version = existing.Version + 1
		}
	} else if item.Version == 0 {
		item.Version = 1
	}
	item = normalizeChangeAttempt(item)
	s.changeAttempts[item.ID] = item
	if proposal, ok := s.proposals[item.ProposalID]; ok {
		proposal.CurrentAttemptID = item.ID
		if item.AttemptNumber > proposal.AttemptCount {
			proposal.AttemptCount = item.AttemptNumber
		}
		proposal.AutoRetryBudgetRemaining = maxInt(0, defaultProposalRetryBudget-item.AttemptNumber)
		proposal.LastFailureClass = item.FailureClass
		proposal.NextRetryAction = item.RetryDecision
		if proposal.Version == 0 {
			proposal.Version = 1
		} else {
			proposal.Version++
		}
		s.proposals[proposal.ID] = normalizeProposalTargetFields(proposal)
	}
	if candidate, ok := s.candidates[item.CandidateKey]; ok {
		candidate.LastAttemptID = item.ID
		if item.AttemptNumber > candidate.AttemptCount {
			candidate.AttemptCount = item.AttemptNumber
		}
		candidate.RetryableFailureClass = firstNonEmpty(item.FailureClass, candidate.RetryableFailureClass)
		if strings.TrimSpace(string(item.TargetLayer)) != "" {
			candidate.CurrentTargetLayer = item.TargetLayer
		}
		candidate.AutoRetryBudgetRemaining = maxInt(0, defaultProposalRetryBudget-item.AttemptNumber)
		if candidate.LineStatus == "" {
			candidate.LineStatus = improvement.LineActive
		}
		candidate.UpdatedAt = item.UpdatedAt
		s.candidates[candidate.CandidateKey] = candidate
	}
	return item, nil
}

func (s *MemoryStore) upsertAttemptWorkspaceLocked(item improvement.AttemptWorkspace) (improvement.AttemptWorkspace, error) {
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextID("ws", len(s.attemptWorkspaces)+1)
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	item = normalizeAttemptWorkspace(item)
	s.attemptWorkspaces[item.ID] = item
	return item, nil
}

func (s *MemoryStore) upsertRepoChangeJobLocked(job improvement.RepoChangeJob) (improvement.RepoChangeJob, error) {
	now := time.Now().UTC()
	if job.ID == "" {
		job.ID = nextID("job", len(s.repoChangeJobs)+1)
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = now
	}
	s.repoChangeJobs[job.ID] = job
	return job, nil
}

func (s *MemoryStore) updateProposalStatusLocked(proposalID string, status review.ProposalStatus) (review.Proposal, error) {
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return review.Proposal{}, errors.New("proposal not found")
	}
	proposal.Status = status
	proposal.ActiveSlotConsuming = review.ConsumesActiveProposalSlot(status)
	if proposal.Version == 0 {
		proposal.Version = 1
	} else {
		proposal.Version++
	}
	s.proposals[proposalID] = normalizeProposalTargetFields(proposal)
	return s.proposals[proposalID], nil
}

func transitionProposalStatusOr(item *review.Proposal) transition.ProposalStatus {
	if item == nil {
		return ""
	}
	return transition.ProposalStatus(item.Status)
}

func transitionAttemptStateOr(item *improvement.ChangeAttempt) transition.AttemptState {
	if item == nil {
		return ""
	}
	return transition.AttemptState(item.State)
}
