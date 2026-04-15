package store

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

const defaultProposalRetryBudget = 3

func (s *MemoryStore) ListChangeAttempts() []improvement.ChangeAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]improvement.ChangeAttempt, 0, len(s.changeAttempts))
	for _, item := range s.changeAttempts {
		out = append(out, normalizeChangeAttempt(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) GetChangeAttempt(attemptID string) (improvement.ChangeAttempt, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.changeAttempts[attemptID]
	if !ok {
		return improvement.ChangeAttempt{}, false
	}
	return normalizeChangeAttempt(item), true
}

func (s *MemoryStore) UpsertChangeAttempt(item improvement.ChangeAttempt) (improvement.ChangeAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	item = normalizeChangeAttempt(item)
	s.changeAttempts[item.ID] = item
	if proposal, ok := s.proposals[item.ProposalID]; ok {
		proposal.CurrentAttemptID = item.ID
		if item.AttemptNumber > proposal.AttemptCount {
			proposal.AttemptCount = item.AttemptNumber
		}
		proposal.AutoRetryBudgetRemaining = maxInt(0, defaultProposalRetryBudget-item.AttemptNumber)
		proposal.LastFailureClass = item.FailureClass
		s.proposals[proposal.ID] = proposal
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

func (s *MemoryStore) StopProposalLine(proposalID string, requestedBy string, rationale string) (review.Proposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return review.Proposal{}, errors.New("proposal not found")
	}
	now := time.Now().UTC()
	proposal.Status = review.ProposalCanceled
	proposal.ActiveSlotConsuming = false
	proposal.NextRetryAction = ""
	proposal.LineStoppedBy = requestedBy
	proposal.LineStopReason = rationale
	proposal.LineStoppedAt = &now
	s.proposals[proposal.ID] = proposal
	if candidate, ok := s.candidates[proposal.CandidateKey]; ok {
		candidate.LineStatus = improvement.LineClosed
		candidate.AutoRetryBudgetRemaining = 0
		candidate.UpdatedAt = now
		s.candidates[candidate.CandidateKey] = candidate
	}
	if proposal.CurrentAttemptID != "" {
		if attempt, ok := s.changeAttempts[proposal.CurrentAttemptID]; ok && !isTerminalAttemptState(attempt.State) {
			attempt.State = improvement.AttemptStateAbandoned
			attempt.FailureClass = firstNonEmpty(attempt.FailureClass, "stopped_by_operator")
			attempt.FailureSummary = firstNonEmpty(attempt.FailureSummary, rationale)
			attempt.RetryDecision = "stop_line"
			attempt.UpdatedAt = now
			s.changeAttempts[attempt.ID] = normalizeChangeAttempt(attempt)
		}
	}
	s.proposalMemory = append(s.proposalMemory, review.ProposalMemory{
		ID:                nextID("memory", len(s.proposalMemory)+1),
		ProposalID:        proposal.ID,
		CandidateKey:      proposal.CandidateKey,
		ConversationID:    proposal.ConversationID,
		CaseID:            proposal.CaseID,
		OriginTraceID:     proposal.OriginTraceID,
		EvidenceTraceIDs:  append([]string(nil), proposal.EvidenceTraceIDs...),
		Hypothesis:        proposal.Summary,
		DiffSummary:       proposal.ProposedScope,
		ReviewRationale:   firstNonEmpty(rationale, "Line stopped by operator."),
		Disposition:       review.ProposalCanceled,
		DispositionReason: firstNonEmpty(rationale, "Line stopped by operator."),
		FailureClass:      "stopped_by_operator",
		LinkedProposalIDs: append([]string(nil), proposal.PriorSimilarProposalIDs...),
		CreatedAt:         now,
	})
	return proposal, nil
}

func (s *MemoryStore) CreateDerivedTrace(req DerivedTraceRequest) (events.Trace, Workflow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := req.CreatedAt.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	var source events.Trace
	if strings.TrimSpace(req.SourceTraceID) != "" {
		source = s.traces[req.SourceTraceID]
	}
	traceID := nextID("trace", len(s.traces)+1)
	workflowID := nextID("wf", len(s.workflows)+1)
	ingestionID := firstNonEmpty(req.IngestionID, source.Summary.IngestionID, fmt.Sprintf("derived:%s", traceID))
	conversationID := firstNonEmpty(req.ConversationID, source.Summary.ConversationID)
	caseID := firstNonEmpty(req.CaseID, source.Summary.CaseID)
	threadKey := firstNonEmpty(req.ThreadKey, source.Summary.ThreadKey, fmt.Sprintf("proposal:%s", req.ProposalID))
	workflowKind := firstNonEmpty(req.WorkflowKind, source.Summary.WorkflowKind, "proposal_attempt")
	triggerEventID := firstNonEmpty(req.TriggerEventID, source.Summary.TriggerEventID)

	workflow := Workflow{
		ID:             workflowID,
		IngestionID:    ingestionID,
		TraceID:        traceID,
		ConversationID: conversationID,
		CaseID:         caseID,
		ThreadKey:      threadKey,
		Kind:           workflowKind,
		Intent:         workflowKind,
		AssignedBot:    "proposal",
		ApprovalMode:   "human_review",
		ResponseMode:   "analysis",
		Status:         "running",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.upsertWorkflowLocked(workflow)
	trace := events.Trace{
		Summary: events.TraceSummary{
			TraceID:           traceID,
			IngestionID:       ingestionID,
			WorkflowID:        workflowID,
			ConversationID:    conversationID,
			CaseID:            caseID,
			TriggerEventID:    triggerEventID,
			SupersedesTraceID: req.SourceTraceID,
			ThreadKey:         threadKey,
			WorkflowKind:      workflowKind,
			Status:            events.StatusQueued,
			StartedAt:         now,
			EndedAt:           now,
		},
		Events: []events.TraceEvent{
			{
				TraceID:        traceID,
				IngestionID:    ingestionID,
				WorkflowID:     workflowID,
				ConversationID: conversationID,
				CaseID:         caseID,
				TriggerEventID: triggerEventID,
				Plane:          "improvement",
				Service:        "improvement-plane",
				Actor:          "attempt-supervisor",
				EventType:      "change_attempt.queued",
				Status:         events.StatusQueued,
				StartedAt:      now,
				Description:    firstNonEmpty(req.Description, fmt.Sprintf("Queued remediation attempt %s for proposal %s.", req.AttemptID, req.ProposalID)),
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:         nextID("reason", len(s.traces)+1),
				TraceID:    traceID,
				WorkflowID: workflowID,
				StepType:   "attempt_bootstrap",
				Summary:    firstNonEmpty(req.Description, fmt.Sprintf("Start bounded remediation attempt %s under proposal %s.", req.AttemptID, req.ProposalID)),
				Confidence: 0.9,
				Decision:   req.AttemptID,
				CreatedAt:  now,
			},
		},
	}
	recomputeTraceSummary(&trace)
	s.traces[traceID] = trace
	return trace, workflow, nil
}

func normalizeChangeAttempt(item improvement.ChangeAttempt) improvement.ChangeAttempt {
	if item.ChangedFiles == nil {
		item.ChangedFiles = []string{}
	}
	if item.OverlayPayload == nil {
		item.OverlayPayload = map[string]any{}
	}
	return item
}

func isTerminalAttemptState(state improvement.ChangeAttemptState) bool {
	switch state {
	case improvement.AttemptStateSandboxFailed,
		improvement.AttemptStateCIFailed,
		improvement.AttemptStateClosedUnmerged,
		improvement.AttemptStateOverlayActive,
		improvement.AttemptStateMerged,
		improvement.AttemptStateNeedsReview,
		improvement.AttemptStateAbandoned,
		improvement.AttemptStateSuperseded:
		return true
	default:
		return false
	}
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
