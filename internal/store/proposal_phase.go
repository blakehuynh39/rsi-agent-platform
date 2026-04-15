package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type ProposalAttemptPhaseAdvance struct {
	ProposalID    string
	WorkItemID    string
	OperationID   string
	ResultRef     string
	Proposal      *review.Proposal
	Attempt       *improvement.ChangeAttempt
	Workspace     *improvement.AttemptWorkspace
	RepoChangeJob *improvement.RepoChangeJob
	TraceID       string
	TraceUpdate   *TraceUpdate
	NextOperation *operation.Execution
	NextWorkItem  *queue.WorkItem
}

type ProposalAttemptPhaseDefer struct {
	ProposalID    string
	WorkItemID    string
	OperationID   string
	LastError     string
	AvailableAt   time.Time
	Payload       map[string]any
	Proposal      *review.Proposal
	Attempt       *improvement.ChangeAttempt
	Workspace     *improvement.AttemptWorkspace
	RepoChangeJob *improvement.RepoChangeJob
	TraceID       string
	TraceUpdate   *TraceUpdate
}

type ProposalAttemptPhaseFailure struct {
	ProposalID    string
	WorkItemID    string
	OperationID   string
	LastError     string
	Proposal      *review.Proposal
	Attempt       *improvement.ChangeAttempt
	Workspace     *improvement.AttemptWorkspace
	RepoChangeJob *improvement.RepoChangeJob
	TraceID       string
	TraceUpdate   *TraceUpdate
	NextOperation *operation.Execution
	NextWorkItem  *queue.WorkItem
}

type proposalAttemptPhaseMutationResult struct {
	ProposalID       string
	CandidateKey     string
	AttemptID        string
	WorkspaceID      string
	TraceID          string
	CurrentWorkItem  string
	CurrentOperation string
	NextWorkItem     string
	NextOperation    string
	RepoJobTouched   bool
	Transition       transitionPersistBundle
}

func (s *MemoryStore) AdvanceProposalAttemptPhase(req ProposalAttemptPhaseAdvance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.advanceProposalAttemptPhaseLocked(req)
	return err
}

func (s *MemoryStore) DeferProposalAttemptPhase(req ProposalAttemptPhaseDefer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.deferProposalAttemptPhaseLocked(req)
	return err
}

func (s *MemoryStore) FailProposalAttemptPhase(req ProposalAttemptPhaseFailure) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.failProposalAttemptPhaseLocked(req)
	return err
}

func (s *MemoryStore) ReconcileProposalAttemptPhase(proposalID string, requestedBy string) (queue.WorkItem, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reconcileProposalAttemptPhaseLocked(proposalID, requestedBy)
}

func isProposalAttemptPhaseKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "line_activate", "attempt_plan", "workspace_open", "implement_attempt", "workspace_validate":
		return true
	default:
		return false
	}
}

func isProposalAttemptPhaseOperation(item queue.WorkItem, op operation.Execution) bool {
	if item.Queue != queue.ProposalQueue {
		return false
	}
	return isProposalAttemptPhaseKind(firstNonEmpty(strings.TrimSpace(op.OperationKind), strings.TrimSpace(item.Kind)))
}

func (s *MemoryStore) currentProposalPhaseWorkLocked(workItemID string, operationID string) (queue.WorkItem, operation.Execution, error) {
	item, ok := s.workItems[strings.TrimSpace(workItemID)]
	if !ok {
		return queue.WorkItem{}, operation.Execution{}, fmt.Errorf("proposal phase work item %s not found", workItemID)
	}
	opID := firstNonEmpty(strings.TrimSpace(operationID), strings.TrimSpace(item.OperationID))
	if opID == "" {
		return queue.WorkItem{}, operation.Execution{}, fmt.Errorf("proposal phase work item %s missing operation_id", item.ID)
	}
	op, ok := s.operations[opID]
	if !ok {
		return queue.WorkItem{}, operation.Execution{}, fmt.Errorf("proposal phase operation %s not found", opID)
	}
	if item.OperationID != "" && strings.TrimSpace(item.OperationID) != op.ID {
		return queue.WorkItem{}, operation.Execution{}, fmt.Errorf("proposal phase work item %s linked to %s, got %s", item.ID, item.OperationID, op.ID)
	}
	if !isProposalAttemptPhaseOperation(item, op) {
		return queue.WorkItem{}, operation.Execution{}, fmt.Errorf("work item %s is not an operation-backed proposal phase item", item.ID)
	}
	return item, op, nil
}

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

func (s *MemoryStore) applyProposalPhaseMutationsLocked(proposal *review.Proposal, attempt *improvement.ChangeAttempt, workspace *improvement.AttemptWorkspace, job *improvement.RepoChangeJob) (proposalID string, attemptID string, workspaceID string, repoJobTouched bool, err error) {
	if attempt != nil {
		var updated improvement.ChangeAttempt
		updated, err = s.upsertChangeAttemptLocked(*attempt)
		if err != nil {
			return "", "", "", false, err
		}
		*attempt = updated
		attemptID = updated.ID
		proposalID = firstNonEmpty(proposalID, updated.ProposalID)
	}
	if workspace != nil {
		var updated improvement.AttemptWorkspace
		updated, err = s.upsertAttemptWorkspaceLocked(*workspace)
		if err != nil {
			return proposalID, attemptID, "", false, err
		}
		*workspace = updated
		workspaceID = updated.ID
		proposalID = firstNonEmpty(proposalID, updated.ProposalID)
	}
	if job != nil {
		var updated improvement.RepoChangeJob
		updated, err = s.upsertRepoChangeJobLocked(*job)
		if err != nil {
			return proposalID, attemptID, workspaceID, false, err
		}
		*job = updated
		proposalID = firstNonEmpty(proposalID, updated.ProposalID)
		repoJobTouched = true
	}
	if proposal != nil {
		updated := s.upsertProposalLocked(*proposal)
		*proposal = updated
		proposalID = firstNonEmpty(proposalID, updated.ID)
	}
	return proposalID, attemptID, workspaceID, repoJobTouched, nil
}

func cloneAnyPayload(payload map[string]any) map[string]any {
	if payload == nil {
		return nil
	}
	return cloneMetadata(payload)
}

func (s *MemoryStore) advanceProposalAttemptPhaseLocked(req ProposalAttemptPhaseAdvance) (proposalAttemptPhaseMutationResult, error) {
	currentItem, currentOp, err := s.currentProposalPhaseWorkLocked(req.WorkItemID, req.OperationID)
	if err != nil {
		return proposalAttemptPhaseMutationResult{}, err
	}
	currentProposal, currentAttempt, err := proposalTransitionSnapshots(firstNonEmpty(strings.TrimSpace(req.ProposalID), currentItem.ProposalID, currentOp.ProposalID), firstNonEmpty(currentOp.AttemptID), s.proposals, s.changeAttempts)
	if err != nil && strings.TrimSpace(currentOp.AttemptID) != "" {
		return proposalAttemptPhaseMutationResult{}, err
	}
	if (req.NextOperation == nil) != (req.NextWorkItem == nil) {
		return proposalAttemptPhaseMutationResult{}, fmt.Errorf("proposal phase advance must provide both next operation and next work item")
	}
	result := proposalAttemptPhaseMutationResult{
		CurrentWorkItem:  currentItem.ID,
		CurrentOperation: currentOp.ID,
	}
	proposalID, attemptID, workspaceID, repoJobTouched, err := s.applyProposalPhaseMutationsLocked(req.Proposal, req.Attempt, req.Workspace, req.RepoChangeJob)
	if err != nil {
		return result, err
	}
	result.ProposalID = firstNonEmpty(proposalID, strings.TrimSpace(req.ProposalID), currentItem.ProposalID, currentOp.ProposalID)
	if req.Attempt != nil {
		result.CandidateKey = strings.TrimSpace(req.Attempt.CandidateKey)
	}
	result.AttemptID = firstNonEmpty(attemptID, currentOp.AttemptID)
	result.WorkspaceID = workspaceID
	result.RepoJobTouched = repoJobTouched
	if req.TraceUpdate != nil && strings.TrimSpace(req.TraceID) != "" {
		if _, err := s.applyTraceUpdateLocked(req.TraceID, *req.TraceUpdate); err != nil {
			return result, err
		}
		result.TraceID = strings.TrimSpace(req.TraceID)
	}
	if req.NextOperation != nil && req.NextWorkItem != nil {
		nextOp, nextItem, _, err := s.ensureOperationWorkItemLocked(*req.NextOperation, *req.NextWorkItem)
		if err != nil {
			return result, err
		}
		result.NextOperation = nextOp.ID
		result.NextWorkItem = nextItem.ID
	}
	now := time.Now().UTC()
	currentItem.Status = queue.WorkCompleted
	currentItem.LastError = ""
	currentItem.LeaseOwner = ""
	currentItem.LeaseExpiresAt = nil
	currentItem.CompletedAt = &now
	currentItem.UpdatedAt = now
	s.workItems[currentItem.ID] = currentItem
	currentOp.Status = operation.StatusCompleted
	currentOp.Holder = ""
	currentOp.ResultRef = firstNonEmpty(strings.TrimSpace(req.ResultRef), currentItem.ID)
	currentOp.LastError = ""
	currentOp.UpdatedAt = now
	currentOp.CompletedAt = &now
	s.operations[currentOp.ID] = currentOp
	updatedProposal, updatedAttempt, err := proposalTransitionSnapshots(firstNonEmpty(result.ProposalID, currentItem.ProposalID, currentOp.ProposalID), firstNonEmpty(result.AttemptID, currentOp.AttemptID), s.proposals, s.changeAttempts)
	if err != nil && strings.TrimSpace(result.AttemptID) != "" {
		return result, err
	}
	commandKind, err := proposalPhaseAdvanceCommand(currentOp.OperationKind, firstNonEmpty(operationKind(req.NextOperation), workItemKind(req.NextWorkItem)))
	if err != nil {
		return result, err
	}
	decision := transition.ReduceAttempt(transition.AttemptSnapshot{
		ProposalStatus:       proposalStatusOr(currentProposal),
		AttemptState:         attemptStateOr(currentAttempt),
		CurrentOperationKind: strings.TrimSpace(currentOp.OperationKind),
	}, transition.CommandEnvelope{
		MachineKind:     transition.MachineAttempt,
		AggregateID:     firstNonEmpty(result.AttemptID, currentOp.AttemptID),
		CommandKind:     string(commandKind),
		CommandID:       currentOp.ID,
		CausationID:     currentItem.ID,
		Actor:           firstNonEmpty(currentItem.RequestedBy, currentOp.RequestedBy),
		OccurredAt:      now,
		ExpectedVersion: aggregateVersionOr(currentAttempt),
	})
	var nextItem *queue.WorkItem
	if result.NextWorkItem != "" {
		if item, ok := s.workItems[result.NextWorkItem]; ok {
			copy := item
			nextItem = &copy
		}
	}
	var nextOp *operation.Execution
	if result.NextOperation != "" {
		if item, ok := s.operations[result.NextOperation]; ok {
			copy := item
			nextOp = &copy
		}
	}
	result.Transition, err = buildTransitionBundle(now, transition.CommandEnvelope{
		MachineKind:     transition.MachineAttempt,
		AggregateID:     firstNonEmpty(result.AttemptID, currentOp.AttemptID),
		CommandKind:     string(commandKind),
		CommandID:       currentOp.ID,
		CausationID:     currentItem.ID,
		Actor:           firstNonEmpty(currentItem.RequestedBy, currentOp.RequestedBy),
		OccurredAt:      now,
		ExpectedVersion: aggregateVersionOr(currentAttempt),
	}, decision, updatedProposal, updatedAttempt, currentItem, currentOp, nextItem, nextOp, "")
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *MemoryStore) deferProposalAttemptPhaseLocked(req ProposalAttemptPhaseDefer) (proposalAttemptPhaseMutationResult, error) {
	currentItem, currentOp, err := s.currentProposalPhaseWorkLocked(req.WorkItemID, req.OperationID)
	if err != nil {
		return proposalAttemptPhaseMutationResult{}, err
	}
	currentProposal, currentAttempt, err := proposalTransitionSnapshots(firstNonEmpty(strings.TrimSpace(req.ProposalID), currentItem.ProposalID, currentOp.ProposalID), firstNonEmpty(currentOp.AttemptID), s.proposals, s.changeAttempts)
	if err != nil && strings.TrimSpace(currentOp.AttemptID) != "" {
		return proposalAttemptPhaseMutationResult{}, err
	}
	result := proposalAttemptPhaseMutationResult{
		CurrentWorkItem:  currentItem.ID,
		CurrentOperation: currentOp.ID,
	}
	proposalID, attemptID, workspaceID, repoJobTouched, err := s.applyProposalPhaseMutationsLocked(req.Proposal, req.Attempt, req.Workspace, req.RepoChangeJob)
	if err != nil {
		return result, err
	}
	result.ProposalID = firstNonEmpty(proposalID, strings.TrimSpace(req.ProposalID), currentItem.ProposalID, currentOp.ProposalID)
	if req.Attempt != nil {
		result.CandidateKey = strings.TrimSpace(req.Attempt.CandidateKey)
	}
	result.AttemptID = firstNonEmpty(attemptID, currentOp.AttemptID)
	result.WorkspaceID = workspaceID
	result.RepoJobTouched = repoJobTouched
	if req.TraceUpdate != nil && strings.TrimSpace(req.TraceID) != "" {
		if _, err := s.applyTraceUpdateLocked(req.TraceID, *req.TraceUpdate); err != nil {
			return result, err
		}
		result.TraceID = strings.TrimSpace(req.TraceID)
	}
	now := time.Now().UTC()
	payload := cloneAnyPayload(req.Payload)
	if payload == nil {
		payload = cloneAnyPayload(currentItem.Payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	if req.AvailableAt.IsZero() {
		delete(payload, "retry_after_unix")
	} else {
		payload["retry_after_unix"] = req.AvailableAt.Unix()
	}
	currentItem.Payload = payload
	currentItem.Status = queue.WorkQueued
	currentItem.LeaseOwner = ""
	currentItem.LeaseExpiresAt = nil
	currentItem.LastError = strings.TrimSpace(req.LastError)
	currentItem.CompletedAt = nil
	currentItem.UpdatedAt = now
	s.workItems[currentItem.ID] = currentItem
	currentOp.Status = operation.StatusQueued
	currentOp.Holder = ""
	currentOp.LastError = strings.TrimSpace(req.LastError)
	currentOp.UpdatedAt = now
	currentOp.CompletedAt = nil
	s.operations[currentOp.ID] = currentOp
	updatedProposal, updatedAttempt, err := proposalTransitionSnapshots(firstNonEmpty(result.ProposalID, currentItem.ProposalID, currentOp.ProposalID), firstNonEmpty(result.AttemptID, currentOp.AttemptID), s.proposals, s.changeAttempts)
	if err != nil && strings.TrimSpace(result.AttemptID) != "" {
		return result, err
	}
	commandKind, err := proposalPhaseDeferCommand(currentOp.OperationKind)
	if err != nil {
		return result, err
	}
	command := transition.CommandEnvelope{
		MachineKind:     transition.MachineAttempt,
		AggregateID:     firstNonEmpty(result.AttemptID, currentOp.AttemptID),
		CommandKind:     string(commandKind),
		CommandID:       currentOp.ID,
		CausationID:     currentItem.ID,
		Actor:           firstNonEmpty(currentItem.RequestedBy, currentOp.RequestedBy),
		OccurredAt:      now,
		ExpectedVersion: aggregateVersionOr(currentAttempt),
	}
	decision := transition.ReduceAttempt(transition.AttemptSnapshot{
		ProposalStatus:       proposalStatusOr(currentProposal),
		AttemptState:         attemptStateOr(currentAttempt),
		CurrentOperationKind: strings.TrimSpace(currentOp.OperationKind),
	}, command)
	result.Transition, err = buildTransitionBundle(now, command, decision, updatedProposal, updatedAttempt, currentItem, currentOp, nil, nil, strings.TrimSpace(req.LastError))
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *MemoryStore) failProposalAttemptPhaseLocked(req ProposalAttemptPhaseFailure) (proposalAttemptPhaseMutationResult, error) {
	currentItem, currentOp, err := s.currentProposalPhaseWorkLocked(req.WorkItemID, req.OperationID)
	if err != nil {
		return proposalAttemptPhaseMutationResult{}, err
	}
	currentProposal, currentAttempt, err := proposalTransitionSnapshots(firstNonEmpty(strings.TrimSpace(req.ProposalID), currentItem.ProposalID, currentOp.ProposalID), firstNonEmpty(currentOp.AttemptID), s.proposals, s.changeAttempts)
	if err != nil && strings.TrimSpace(currentOp.AttemptID) != "" {
		return proposalAttemptPhaseMutationResult{}, err
	}
	if (req.NextOperation == nil) != (req.NextWorkItem == nil) {
		return proposalAttemptPhaseMutationResult{}, fmt.Errorf("proposal phase failure must provide both next operation and next work item")
	}
	result := proposalAttemptPhaseMutationResult{
		CurrentWorkItem:  currentItem.ID,
		CurrentOperation: currentOp.ID,
	}
	proposalID, attemptID, workspaceID, repoJobTouched, err := s.applyProposalPhaseMutationsLocked(req.Proposal, req.Attempt, req.Workspace, req.RepoChangeJob)
	if err != nil {
		return result, err
	}
	result.ProposalID = firstNonEmpty(proposalID, strings.TrimSpace(req.ProposalID), currentItem.ProposalID, currentOp.ProposalID)
	if req.Attempt != nil {
		result.CandidateKey = strings.TrimSpace(req.Attempt.CandidateKey)
	}
	result.AttemptID = firstNonEmpty(attemptID, currentOp.AttemptID)
	result.WorkspaceID = workspaceID
	result.RepoJobTouched = repoJobTouched
	if req.TraceUpdate != nil && strings.TrimSpace(req.TraceID) != "" {
		if _, err := s.applyTraceUpdateLocked(req.TraceID, *req.TraceUpdate); err != nil {
			return result, err
		}
		result.TraceID = strings.TrimSpace(req.TraceID)
	}
	if req.NextOperation != nil && req.NextWorkItem != nil {
		nextOp, nextItem, _, err := s.ensureOperationWorkItemLocked(*req.NextOperation, *req.NextWorkItem)
		if err != nil {
			return result, err
		}
		result.NextOperation = nextOp.ID
		result.NextWorkItem = nextItem.ID
	}
	now := time.Now().UTC()
	currentItem.Status = queue.WorkFailed
	currentItem.LeaseOwner = ""
	currentItem.LeaseExpiresAt = nil
	currentItem.LastError = strings.TrimSpace(req.LastError)
	currentItem.CompletedAt = &now
	currentItem.UpdatedAt = now
	s.workItems[currentItem.ID] = currentItem
	currentOp.Status = operation.StatusFailed
	currentOp.Holder = ""
	currentOp.LastError = strings.TrimSpace(req.LastError)
	currentOp.RetryCount++
	currentOp.UpdatedAt = now
	currentOp.CompletedAt = &now
	s.operations[currentOp.ID] = currentOp
	updatedProposal, updatedAttempt, err := proposalTransitionSnapshots(firstNonEmpty(result.ProposalID, currentItem.ProposalID, currentOp.ProposalID), firstNonEmpty(result.AttemptID, currentOp.AttemptID), s.proposals, s.changeAttempts)
	if err != nil && strings.TrimSpace(result.AttemptID) != "" {
		return result, err
	}
	commandKind, err := proposalPhaseFailureCommand(currentOp.OperationKind, updatedProposal, updatedAttempt)
	if err != nil {
		return result, err
	}
	command := transition.CommandEnvelope{
		MachineKind:     transition.MachineAttempt,
		AggregateID:     firstNonEmpty(result.AttemptID, currentOp.AttemptID),
		CommandKind:     string(commandKind),
		CommandID:       currentOp.ID,
		CausationID:     currentItem.ID,
		Actor:           firstNonEmpty(currentItem.RequestedBy, currentOp.RequestedBy),
		OccurredAt:      now,
		ExpectedVersion: aggregateVersionOr(currentAttempt),
	}
	decision := transition.ReduceAttempt(transition.AttemptSnapshot{
		ProposalStatus:       proposalStatusOr(currentProposal),
		AttemptState:         attemptStateOr(currentAttempt),
		CurrentOperationKind: strings.TrimSpace(currentOp.OperationKind),
	}, command)
	var nextItem *queue.WorkItem
	if result.NextWorkItem != "" {
		if item, ok := s.workItems[result.NextWorkItem]; ok {
			copy := item
			nextItem = &copy
		}
	}
	var nextOp *operation.Execution
	if result.NextOperation != "" {
		if item, ok := s.operations[result.NextOperation]; ok {
			copy := item
			nextOp = &copy
		}
	}
	result.Transition, err = buildTransitionBundle(now, command, decision, updatedProposal, updatedAttempt, currentItem, currentOp, nextItem, nextOp, strings.TrimSpace(req.LastError))
	if err != nil {
		return result, err
	}
	return result, nil
}

func proposalStatusOr(item *review.Proposal) review.ProposalStatus {
	if item == nil {
		return ""
	}
	return item.Status
}

func attemptStateOr(item *improvement.ChangeAttempt) improvement.ChangeAttemptState {
	if item == nil {
		return ""
	}
	return item.State
}

func aggregateVersionOr(item *improvement.ChangeAttempt) int64 {
	if item == nil {
		return 0
	}
	return item.Version
}

func operationKind(item *operation.Execution) string {
	if item == nil {
		return ""
	}
	return strings.TrimSpace(item.OperationKind)
}

func workItemKind(item *queue.WorkItem) string {
	if item == nil {
		return ""
	}
	return strings.TrimSpace(item.Kind)
}

func activePhaseOperationLocked(ops []operation.Execution) (operation.Execution, bool) {
	var best operation.Execution
	found := false
	for _, op := range ops {
		if !isProposalAttemptPhaseKind(op.OperationKind) && strings.TrimSpace(op.OperationKind) != "pr_open" {
			continue
		}
		if op.Status != operation.StatusQueued && op.Status != operation.StatusRunning {
			continue
		}
		if !found || op.UpdatedAt.After(best.UpdatedAt) {
			best = op
			found = true
		}
	}
	return best, found
}

func activeRepoChangeResumeOperation(ops []operation.Execution) (operation.Execution, bool) {
	var best operation.Execution
	found := false
	for _, op := range ops {
		kind := strings.TrimSpace(op.OperationKind)
		if kind != "sandbox_launch" && kind != "pr_open" {
			continue
		}
		if op.Status != operation.StatusQueued && op.Status != operation.StatusRunning {
			continue
		}
		if !found || op.UpdatedAt.After(best.UpdatedAt) {
			best = op
			found = true
		}
	}
	return best, found
}

func queuedOrLeasedWorkItemByOperation(items []queue.WorkItem, operationID string) (queue.WorkItem, bool) {
	for _, item := range items {
		if strings.TrimSpace(item.OperationID) != strings.TrimSpace(operationID) {
			continue
		}
		if item.Status == queue.WorkQueued || item.Status == queue.WorkLeased {
			return item, true
		}
	}
	return queue.WorkItem{}, false
}

func hasAttemptOperationLocked(ops []operation.Execution, kind string) bool {
	for _, op := range ops {
		if strings.TrimSpace(op.OperationKind) == strings.TrimSpace(kind) {
			return true
		}
	}
	return false
}

func findWorkItemByOperationLocked(items map[string]queue.WorkItem, operationID string) (queue.WorkItem, bool) {
	for _, item := range items {
		if strings.TrimSpace(item.OperationID) == strings.TrimSpace(operationID) {
			return item, true
		}
	}
	return queue.WorkItem{}, false
}

func proposalPhaseWorkItemPayload(kind string, proposal review.Proposal, attempt improvement.ChangeAttempt, workspace *improvement.AttemptWorkspace, job *improvement.RepoChangeJob) map[string]any {
	payload := map[string]any{
		"attempt_id": attempt.ID,
	}
	if workspace != nil && workspace.ID != "" {
		payload["workspace_id"] = workspace.ID
	}
	switch strings.TrimSpace(kind) {
	case "workspace_validate":
		repo := proposalRepo(proposal)
		branchName := attempt.BranchName
		baseRef := "main"
		if workspace != nil {
			repo = firstNonEmpty(workspace.Repo, repo)
			branchName = firstNonEmpty(workspace.BranchName, branchName)
			baseRef = firstNonEmpty(workspace.BaseRef, baseRef)
		}
		payload["validation_command"] = "make test"
		payload["branch_name"] = branchName
		payload["base_ref"] = baseRef
		payload["title"] = fmt.Sprintf("RSI proposal %s attempt %s for %s", proposal.ID, attempt.ID, repo)
		payload["body"] = fmt.Sprintf("Automated draft PR for proposal %s attempt %s after workspace validation.", proposal.ID, attempt.ID)
	}
	if strings.TrimSpace(kind) == "pr_open" && job != nil {
		payload["job_id"] = job.ID
		payload["job_name"] = job.SandboxJobName
		payload["namespace"] = job.SandboxNamespace
		payload["repo"] = job.Repo
		payload["branch_name"] = job.BranchName
		payload["base_ref"] = firstNonEmpty(job.BaseRef, "main")
		payload["title"] = fmt.Sprintf("RSI proposal %s attempt %s for %s", proposal.ID, attempt.ID, job.Repo)
		payload["body"] = fmt.Sprintf("Automated draft PR for proposal %s attempt %s after workspace validation.", proposal.ID, attempt.ID)
	}
	return payload
}

func proposalPhaseWorkItemTemplate(kind string, proposal review.Proposal, attempt improvement.ChangeAttempt, requestedBy string, payload map[string]any) queue.WorkItem {
	now := time.Now().UTC()
	item := queue.WorkItem{
		Queue:          queue.ProposalQueue,
		Kind:           kind,
		Status:         queue.WorkQueued,
		TraceID:        attempt.AttemptTraceID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		TriggerEventID: proposal.OriginTraceID,
		ProposalID:     proposal.ID,
		RequestedBy:    requestedBy,
		ApprovalMode:   "human_review",
		CreatedAt:      now,
		UpdatedAt:      now,
		Payload:        payload,
	}
	if item.Payload == nil {
		item.Payload = map[string]any{}
	}
	return item
}

func proposalPhaseOperationTemplate(kind string, proposal review.Proposal, attempt improvement.ChangeAttempt, requestedBy string) operation.Execution {
	now := time.Now().UTC()
	scopeKind := operation.ScopeAttempt
	scopeID := attempt.ID
	operationKey := kind
	if strings.TrimSpace(kind) == "line_activate" {
		scopeKind = operation.ScopeProposal
		scopeID = proposal.ID
		operationKey = fmt.Sprintf("attempt-%02d", attempt.AttemptNumber)
	}
	return operation.Execution{
		ScopeKind:     scopeKind,
		ScopeID:       scopeID,
		OperationKind: kind,
		OperationKey:  operationKey,
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   requestedBy,
		TraceID:       attempt.AttemptTraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func prOpenOperationTemplate(proposal review.Proposal, attempt improvement.ChangeAttempt, requestedBy string) operation.Execution {
	now := time.Now().UTC()
	return operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: "pr_open",
		OperationKey:  "pr_open",
		Status:        operation.StatusQueued,
		Queue:         queue.ImprovementActionQueue,
		RequestedBy:   requestedBy,
		TraceID:       attempt.AttemptTraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func prOpenWorkItemTemplate(proposal review.Proposal, attempt improvement.ChangeAttempt, job improvement.RepoChangeJob, requestedBy string) queue.WorkItem {
	now := time.Now().UTC()
	return queue.WorkItem{
		Queue:          queue.ImprovementActionQueue,
		Kind:           "draft_pr_open",
		Status:         queue.WorkQueued,
		TraceID:        attempt.AttemptTraceID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		TriggerEventID: proposal.OriginTraceID,
		ProposalID:     proposal.ID,
		RepoScope:      job.Repo,
		RequestedBy:    requestedBy,
		ApprovalMode:   "approved",
		CreatedAt:      now,
		UpdatedAt:      now,
		Payload:        proposalPhaseWorkItemPayload("pr_open", proposal, attempt, nil, &job),
	}
}

func (s *MemoryStore) reconcileProposalAttemptPhaseLocked(proposalID string, requestedBy string) (queue.WorkItem, bool, error) {
	proposal, ok := s.proposals[strings.TrimSpace(proposalID)]
	if !ok {
		return queue.WorkItem{}, false, errors.New("proposal not found")
	}
	if proposal.Status != review.ProposalApproved &&
		proposal.Status != review.ProposalRepoChangeQueued &&
		proposal.Status != review.ProposalRepoChangeRunning &&
		proposal.Status != review.ProposalValidationPending {
		return queue.WorkItem{}, false, nil
	}
	if !review.ProposalExecutableIntervention(proposal.RecommendedInterventionKind) {
		return queue.WorkItem{}, false, nil
	}
	if strings.TrimSpace(proposal.CurrentAttemptID) == "" {
		nextAttemptNumber := maxInt(1, proposal.AttemptCount+1)
		attempt := improvement.ChangeAttempt{
			ID:             fmt.Sprintf("attempt-%s-%02d", strings.ReplaceAll(strings.TrimSpace(proposal.ID), "/", "-"), nextAttemptNumber),
			ProposalID:     proposal.ID,
			CandidateKey:   proposal.CandidateKey,
			AttemptNumber:  nextAttemptNumber,
			TargetLayer:    proposal.TargetLayer,
			TargetKind:     proposal.TargetKind,
			TargetRef:      proposal.TargetRef,
			AttemptTraceID: firstNonEmpty(proposal.TraceID, proposal.OriginTraceID),
		}
		op := proposalPhaseOperationTemplate("line_activate", proposal, attempt, requestedBy)
		item := queue.WorkItem{
			Queue:          queue.ProposalQueue,
			Kind:           "approved_proposal",
			Status:         queue.WorkQueued,
			TraceID:        proposal.TraceID,
			ConversationID: proposal.ConversationID,
			CaseID:         proposal.CaseID,
			TriggerEventID: proposal.OriginTraceID,
			ProposalID:     proposal.ID,
			RequestedBy:    requestedBy,
			ApprovalMode:   "human_review",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			Payload: map[string]any{
				"trigger":        string(improvement.AttemptTriggerOperatorRetry),
				"candidate_key":  proposal.CandidateKey,
				"risk_tier":      proposal.RiskTier,
				"attempt_number": nextAttemptNumber,
			},
		}
		_, created, _, err := s.ensureQueuedOperationWorkItemLocked(op, item)
		return created, err == nil, err
	}
	attempt, ok := s.changeAttempts[proposal.CurrentAttemptID]
	if !ok || isTerminalAttemptState(attempt.State) {
		return queue.WorkItem{}, false, nil
	}
	ops := listOperationsByScopeLocked(s.operations, operation.ScopeAttempt, attempt.ID)
	if active, ok := activePhaseOperationLocked(ops); ok {
		if item, ok := findWorkItemByOperationLocked(s.workItems, active.ID); ok && (item.Status == queue.WorkQueued || item.Status == queue.WorkLeased) {
			return item, false, nil
		}
		var item queue.WorkItem
		switch strings.TrimSpace(active.OperationKind) {
		case "pr_open":
			job, ok := repoChangeJobByAttemptLocked(s.repoChangeJobs, proposal.ID, attempt.ID)
			if !ok {
				return queue.WorkItem{}, false, nil
			}
			item = prOpenWorkItemTemplate(proposal, attempt, job, requestedBy)
		default:
			workspace, workspaceOK := findAttemptWorkspaceByAttemptLocked(s.attemptWorkspaces, attempt.ID)
			item = proposalPhaseWorkItemTemplate(active.OperationKind, proposal, attempt, requestedBy, proposalPhaseWorkItemPayload(active.OperationKind, proposal, attempt, workspacePtr(workspace, workspaceOK), nil))
		}
		_, created, _, err := s.ensureQueuedOperationWorkItemLocked(active, item)
		return created, err == nil, err
	}
	workspace, workspaceOK := findAttemptWorkspaceByAttemptLocked(s.attemptWorkspaces, attempt.ID)
	job, jobOK := repoChangeJobByAttemptLocked(s.repoChangeJobs, proposal.ID, attempt.ID)
	if proposal.Status == review.ProposalValidationPending && jobOK && !hasAttemptOperationLocked(ops, "pr_open") {
		op := prOpenOperationTemplate(proposal, attempt, requestedBy)
		item := prOpenWorkItemTemplate(proposal, attempt, job, requestedBy)
		_, created, _, err := s.ensureQueuedOperationWorkItemLocked(op, item)
		return created, err == nil, err
	}
	if hasImplementArtifacts(attempt) && !hasAttemptOperationLocked(ops, "workspace_validate") {
		op := proposalPhaseOperationTemplate("workspace_validate", proposal, attempt, requestedBy)
		item := proposalPhaseWorkItemTemplate("workspace_validate", proposal, attempt, requestedBy, proposalPhaseWorkItemPayload("workspace_validate", proposal, attempt, workspacePtr(workspace, workspaceOK), workspaceJobPtr(job, jobOK)))
		_, created, _, err := s.ensureQueuedOperationWorkItemLocked(op, item)
		return created, err == nil, err
	}
	if workspaceOK && workspaceReady(workspace) && !hasAttemptOperationLocked(ops, "implement_attempt") {
		op := proposalPhaseOperationTemplate("implement_attempt", proposal, attempt, requestedBy)
		item := proposalPhaseWorkItemTemplate("implement_attempt", proposal, attempt, requestedBy, proposalPhaseWorkItemPayload("implement_attempt", proposal, attempt, &workspace, workspaceJobPtr(job, jobOK)))
		_, created, _, err := s.ensureQueuedOperationWorkItemLocked(op, item)
		return created, err == nil, err
	}
	if (workspaceOK && !workspaceReady(workspace)) || attempt.State == improvement.AttemptStatePatchPlan {
		op := proposalPhaseOperationTemplate("workspace_open", proposal, attempt, requestedBy)
		item := proposalPhaseWorkItemTemplate("workspace_open", proposal, attempt, requestedBy, proposalPhaseWorkItemPayload("workspace_open", proposal, attempt, workspacePtr(workspace, workspaceOK), workspaceJobPtr(job, jobOK)))
		_, created, _, err := s.ensureQueuedOperationWorkItemLocked(op, item)
		return created, err == nil, err
	}
	return queue.WorkItem{}, false, nil
}

func workspaceReady(item improvement.AttemptWorkspace) bool {
	return item.Status == improvement.WorkspaceReady || strings.TrimSpace(item.PodName) != ""
}

func hasImplementArtifacts(attempt improvement.ChangeAttempt) bool {
	return strings.TrimSpace(attempt.RepoPatch) != "" ||
		len(attempt.ChangedFiles) > 0 ||
		strings.TrimSpace(attempt.DiffSummary) != "" ||
		attempt.State == improvement.AttemptStatePatchGenerated ||
		attempt.State == improvement.AttemptStateValidationRunning
}

func findAttemptWorkspaceByAttemptLocked(items map[string]improvement.AttemptWorkspace, attemptID string) (improvement.AttemptWorkspace, bool) {
	for _, item := range items {
		if strings.TrimSpace(item.AttemptID) == strings.TrimSpace(attemptID) {
			return item, true
		}
	}
	return improvement.AttemptWorkspace{}, false
}

func repoChangeJobByAttemptLocked(items map[string]improvement.RepoChangeJob, proposalID string, attemptID string) (improvement.RepoChangeJob, bool) {
	for _, item := range items {
		if strings.TrimSpace(item.ProposalID) == strings.TrimSpace(proposalID) &&
			strings.TrimSpace(item.AttemptID) == strings.TrimSpace(attemptID) {
			return item, true
		}
	}
	return improvement.RepoChangeJob{}, false
}

func workspacePtr(item improvement.AttemptWorkspace, ok bool) *improvement.AttemptWorkspace {
	if !ok {
		return nil
	}
	copy := item
	return &copy
}

func workspaceJobPtr(item improvement.RepoChangeJob, ok bool) *improvement.RepoChangeJob {
	if !ok {
		return nil
	}
	copy := item
	return &copy
}
