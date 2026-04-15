package store

import (
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
)

func (s *MemoryStore) ListAttemptWorkspaces() []improvement.AttemptWorkspace {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]improvement.AttemptWorkspace, 0, len(s.attemptWorkspaces))
	for _, item := range s.attemptWorkspaces {
		out = append(out, normalizeAttemptWorkspace(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) GetAttemptWorkspace(workspaceID string) (improvement.AttemptWorkspace, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.attemptWorkspaces[strings.TrimSpace(workspaceID)]
	if !ok {
		return improvement.AttemptWorkspace{}, false
	}
	return normalizeAttemptWorkspace(item), true
}

func (s *MemoryStore) GetAttemptWorkspaceByAttempt(attemptID string) (improvement.AttemptWorkspace, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	attemptID = strings.TrimSpace(attemptID)
	for _, item := range s.attemptWorkspaces {
		if strings.TrimSpace(item.AttemptID) == attemptID {
			return normalizeAttemptWorkspace(item), true
		}
	}
	return improvement.AttemptWorkspace{}, false
}

func (s *MemoryStore) UpsertAttemptWorkspace(item improvement.AttemptWorkspace) (improvement.AttemptWorkspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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

func normalizeAttemptWorkspace(item improvement.AttemptWorkspace) improvement.AttemptWorkspace {
	item.ID = strings.TrimSpace(item.ID)
	item.AttemptID = strings.TrimSpace(item.AttemptID)
	item.ProposalID = strings.TrimSpace(item.ProposalID)
	item.Repo = strings.TrimSpace(item.Repo)
	item.BaseRef = firstNonEmpty(strings.TrimSpace(item.BaseRef), "main")
	item.BranchName = strings.TrimSpace(item.BranchName)
	item.Namespace = strings.TrimSpace(item.Namespace)
	item.JobName = strings.TrimSpace(item.JobName)
	item.PodName = strings.TrimSpace(item.PodName)
	if item.Status == "" {
		item.Status = improvement.WorkspaceQueued
	}
	if item.AllowedPathGlobs == nil {
		item.AllowedPathGlobs = []string{}
	}
	item.HeadSHA = strings.TrimSpace(item.HeadSHA)
	item.DiffSummary = strings.TrimSpace(item.DiffSummary)
	return item
}
