package store

import (
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const defaultWorkflowAutoRetryMaxAttempts = 3

func (s *MemoryStore) ListWorkflowLines() []WorkflowLine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]WorkflowLine, 0, len(s.workflowLines))
	for _, item := range s.workflowLines {
		out = append(out, normalizeWorkflowLine(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].CaseID > out[j].CaseID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *MemoryStore) GetWorkflowLine(caseID string) (WorkflowLine, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.workflowLines[strings.TrimSpace(caseID)]
	if !ok {
		return WorkflowLine{}, false
	}
	return normalizeWorkflowLine(item), true
}

func normalizeWorkflowLine(item WorkflowLine) WorkflowLine {
	if item.AttemptCount < 0 {
		item.AttemptCount = 0
	}
	if item.AutoRetryBudgetRemaining < 0 {
		item.AutoRetryBudgetRemaining = 0
	}
	return item
}

func workflowLineStatusFromState(state transition.WorkflowLineStateKind) string {
	return strings.TrimSpace(string(state))
}

func workflowLineStateFromStatus(status string) transition.WorkflowLineStateKind {
	switch strings.TrimSpace(status) {
	case "", string(transition.WorkflowLineStateActive):
		return transition.WorkflowLineStateActive
	case string(transition.WorkflowLineStateRetryScheduled):
		return transition.WorkflowLineStateRetryScheduled
	case "needs-human", string(transition.WorkflowLineStateNeedsHuman):
		return transition.WorkflowLineStateNeedsHuman
	case string(transition.WorkflowLineStateCompleted):
		return transition.WorkflowLineStateCompleted
	case string(transition.WorkflowLineStateSuperseded):
		return transition.WorkflowLineStateSuperseded
	default:
		return transition.WorkflowLineStateKind(strings.TrimSpace(status))
	}
}

func workflowLineRetryBudgetRemaining(attemptCount int) int {
	if attemptCount < 0 {
		attemptCount = 0
	}
	remaining := defaultWorkflowAutoRetryMaxAttempts - attemptCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *MemoryStore) ensureWorkflowLineLocked(caseRecord conversation.Case, createdAt time.Time) WorkflowLine {
	caseID := strings.TrimSpace(caseRecord.ID)
	if existing, ok := s.workflowLines[caseID]; ok {
		existing.ConversationID = firstNonEmpty(caseRecord.ConversationID, existing.ConversationID)
		if existing.Status == "" {
			existing.Status = workflowLineStatusFromState(transition.WorkflowLineStateActive)
		}
		if existing.Version == 0 {
			existing.Version = 1
		}
		if existing.CreatedAt.IsZero() {
			existing.CreatedAt = createdAt
		}
		existing.UpdatedAt = createdAt
		s.workflowLines[caseID] = existing
		return existing
	}
	item := WorkflowLine{
		CaseID:                   caseID,
		ConversationID:           caseRecord.ConversationID,
		Status:                   workflowLineStatusFromState(transition.WorkflowLineStateActive),
		AutoRetryBudgetRemaining: workflowLineRetryBudgetRemaining(0),
		Version:                  1,
		CreatedAt:                createdAt,
		UpdatedAt:                createdAt,
	}
	s.workflowLines[caseID] = item
	return item
}

func (s *MemoryStore) upsertWorkflowLineLocked(item WorkflowLine) {
	item = normalizeWorkflowLine(item)
	item.CaseID = strings.TrimSpace(item.CaseID)
	if item.CaseID == "" {
		return
	}
	for existingID, existing := range s.workflowLines {
		if existingID != item.CaseID {
			continue
		}
		if item.Version <= existing.Version {
			item.Version = existing.Version + 1
		}
		s.workflowLines[item.CaseID] = item
		return
	}
	if item.Version == 0 {
		item.Version = 1
	}
	s.workflowLines[item.CaseID] = item
}

func findWorkflowLineByCaseID(items []WorkflowLine, caseID string) (WorkflowLine, bool) {
	caseID = strings.TrimSpace(caseID)
	for _, item := range items {
		if strings.TrimSpace(item.CaseID) == caseID {
			return item, true
		}
	}
	return WorkflowLine{}, false
}
