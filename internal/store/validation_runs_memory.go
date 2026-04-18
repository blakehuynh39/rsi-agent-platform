package store

import (
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
)

func (s *MemoryStore) ListValidationRuns() []improvement.ValidationRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]improvement.ValidationRun, 0, len(s.validationRuns))
	for _, item := range s.validationRuns {
		out = append(out, normalizeValidationRun(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) RecordValidationRun(run improvement.ValidationRun) (improvement.ValidationRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recordValidationRunLocked(run)
}

func (s *MemoryStore) recordValidationRunLocked(run improvement.ValidationRun) (improvement.ValidationRun, error) {
	run = normalizeValidationRun(run)
	now := time.Now().UTC()
	if run.ID == "" {
		run.ID = nextID("validation", len(s.validationRuns)+1)
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	if run.UpdatedAt.IsZero() || run.UpdatedAt.Before(run.CreatedAt) {
		run.UpdatedAt = run.CreatedAt
	}
	if run.ProposalID != "" {
		if proposal, ok := s.proposals[run.ProposalID]; ok {
			run.ConversationID = firstNonEmpty(run.ConversationID, proposal.ConversationID)
			run.CaseID = firstNonEmpty(run.CaseID, proposal.CaseID)
			run.OriginTraceID = firstNonEmpty(run.OriginTraceID, proposal.OriginTraceID)
		}
	}
	s.validationRuns[run.ID] = run
	return run, nil
}

func normalizeValidationRun(item improvement.ValidationRun) improvement.ValidationRun {
	item.ID = strings.TrimSpace(item.ID)
	item.ProposalID = strings.TrimSpace(item.ProposalID)
	item.AttemptID = strings.TrimSpace(item.AttemptID)
	item.ConversationID = strings.TrimSpace(item.ConversationID)
	item.CaseID = strings.TrimSpace(item.CaseID)
	item.OriginTraceID = strings.TrimSpace(item.OriginTraceID)
	item.WorkspaceID = strings.TrimSpace(item.WorkspaceID)
	item.OperationID = strings.TrimSpace(item.OperationID)
	item.Repo = strings.TrimSpace(item.Repo)
	item.BranchName = strings.TrimSpace(item.BranchName)
	item.Command = strings.TrimSpace(item.Command)
	if item.Status == "" {
		item.Status = improvement.ValidationRunRequested
	}
	item.SandboxNamespace = strings.TrimSpace(item.SandboxNamespace)
	item.SandboxJobName = strings.TrimSpace(item.SandboxJobName)
	item.SandboxPodName = strings.TrimSpace(item.SandboxPodName)
	item.ValidationRef = strings.TrimSpace(item.ValidationRef)
	item.ErrorMessage = strings.TrimSpace(item.ErrorMessage)
	item.LogArtifactID = strings.TrimSpace(item.LogArtifactID)
	return item
}
