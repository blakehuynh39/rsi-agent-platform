package store

import (
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/harness"
)

func (s *MemoryStore) ListHarnessProfiles() []harness.Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]harness.Profile, 0, len(s.harnessProfiles))
	for _, item := range s.harnessProfiles {
		out = append(out, normalizeHarnessProfile(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Role == out[j].Role {
			return out[i].ID < out[j].ID
		}
		return out[i].Role < out[j].Role
	})
	return out
}

func (s *MemoryStore) GetHarnessProfile(profileID string) (harness.Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.harnessProfiles[strings.TrimSpace(profileID)]
	if !ok {
		return harness.Profile{}, false
	}
	return normalizeHarnessProfile(item), true
}

func (s *MemoryStore) ListHarnessOverlays() []harness.Overlay {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]harness.Overlay, 0, len(s.harnessOverlays))
	for _, item := range s.harnessOverlays {
		out = append(out, normalizeHarnessOverlay(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *MemoryStore) GetActiveHarnessOverlay(role string) (harness.Overlay, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getActiveHarnessOverlayLocked(role)
}

func (s *MemoryStore) getActiveHarnessOverlayLocked(role string) (harness.Overlay, bool) {
	role = strings.TrimSpace(role)
	var selected harness.Overlay
	ok := false
	for _, item := range s.harnessOverlays {
		if item.Role != role || item.Status != harness.OverlayStatusActive {
			continue
		}
		if !ok || item.UpdatedAt.After(selected.UpdatedAt) {
			selected = item
			ok = true
		}
	}
	if !ok {
		return harness.Overlay{}, false
	}
	return normalizeHarnessOverlay(selected), true
}

func (s *MemoryStore) UpsertHarnessOverlay(item harness.Overlay) (harness.Overlay, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	item = normalizeHarnessOverlay(item)
	if item.ID == "" {
		item.ID = nextUUID("overlay")
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = now
	}
	if item.Status == harness.OverlayStatusActive {
		for id, existing := range s.harnessOverlays {
			if existing.Role == item.Role && existing.Status == harness.OverlayStatusActive && id != item.ID {
				existing.Status = harness.OverlayStatusSuperseded
				existing.UpdatedAt = now
				s.harnessOverlays[id] = normalizeHarnessOverlay(existing)
			}
		}
		if item.ActivatedAt == nil {
			item.ActivatedAt = ptrTimeValue(now)
		}
	}
	s.harnessOverlays[item.ID] = item
	return item, nil
}

func (s *MemoryStore) ListHarnessExperiments() []harness.Experiment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]harness.Experiment, 0, len(s.harnessExperiments))
	for _, item := range s.harnessExperiments {
		out = append(out, normalizeHarnessExperiment(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *MemoryStore) RecordHarnessExperiment(item harness.Experiment) (harness.Experiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	item = normalizeHarnessExperiment(item)
	if item.ID == "" {
		item.ID = nextUUID("hexp")
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = now
	}
	s.harnessExperiments[item.ID] = item
	return item, nil
}

func (s *MemoryStore) ListHarnessSessionBindings() []harness.SessionBinding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]harness.SessionBinding, 0, len(s.harnessSessionBindings))
	for _, item := range s.harnessSessionBindings {
		out = append(out, normalizeHarnessSessionBinding(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].LastUsedAt.Equal(out[j].LastUsedAt) {
			if out[i].Role == out[j].Role {
				if out[i].ScopeKind == out[j].ScopeKind {
					return out[i].ScopeID < out[j].ScopeID
				}
				return out[i].ScopeKind < out[j].ScopeKind
			}
			return out[i].Role < out[j].Role
		}
		return out[i].LastUsedAt.After(out[j].LastUsedAt)
	})
	return out
}

func (s *MemoryStore) UpsertHarnessSessionBinding(item harness.SessionBinding) (harness.SessionBinding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	item = normalizeHarnessSessionBinding(item)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.LastUsedAt.IsZero() {
		item.LastUsedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = now
	}
	key := harnessSessionBindingKey(item.Role, item.ScopeKind, item.ScopeID)
	if existing, ok := s.harnessSessionBindings[key]; ok {
		item.CreatedAt = existing.CreatedAt
	}
	s.harnessSessionBindings[key] = item
	return item, nil
}

func (s *MemoryStore) ListHarnessExecutions() []harness.Execution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]harness.Execution(nil), s.harnessExecutions...)
	for i := range out {
		out[i] = normalizeHarnessExecution(out[i])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) RecordHarnessExecution(item harness.Execution) (harness.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	item = normalizeHarnessExecution(item)
	if item.ID == "" {
		item.ID = nextUUID("hexec")
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	for i, existing := range s.harnessExecutions {
		if existing.ID == item.ID {
			s.harnessExecutions[i] = item
			return item, nil
		}
	}
	s.harnessExecutions = append(s.harnessExecutions, item)
	return item, nil
}

func harnessSessionBindingKey(role, scopeKind, scopeID string) string {
	return strings.TrimSpace(role) + "|" + strings.TrimSpace(scopeKind) + "|" + strings.TrimSpace(scopeID)
}

func normalizeHarnessProfile(item harness.Profile) harness.Profile {
	item.PromptFragments = append([]string(nil), item.PromptFragments...)
	if item.PromptFragments == nil {
		item.PromptFragments = []string{}
	}
	item.FewShotSnippets = append([]string(nil), item.FewShotSnippets...)
	if item.FewShotSnippets == nil {
		item.FewShotSnippets = []string{}
	}
	item.ToolPreferenceOrder = append([]string(nil), item.ToolPreferenceOrder...)
	if item.ToolPreferenceOrder == nil {
		item.ToolPreferenceOrder = []string{}
	}
	return item
}

func normalizeHarnessOverlay(item harness.Overlay) harness.Overlay {
	item.PromptFragments = append([]string(nil), item.PromptFragments...)
	if item.PromptFragments == nil {
		item.PromptFragments = []string{}
	}
	item.FewShotSnippets = append([]string(nil), item.FewShotSnippets...)
	if item.FewShotSnippets == nil {
		item.FewShotSnippets = []string{}
	}
	item.ToolPreferenceOrder = append([]string(nil), item.ToolPreferenceOrder...)
	if item.ToolPreferenceOrder == nil {
		item.ToolPreferenceOrder = []string{}
	}
	return item
}

func normalizeHarnessExperiment(item harness.Experiment) harness.Experiment {
	if item.Metrics == nil {
		item.Metrics = map[string]any{}
	}
	return item
}

func normalizeHarnessSessionBinding(item harness.SessionBinding) harness.SessionBinding {
	return item
}

func normalizeHarnessExecution(item harness.Execution) harness.Execution {
	item.MemoryReads = memoryArtifactsOrEmpty(item.MemoryReads)
	item.MemoryWrites = memoryArtifactsOrEmpty(item.MemoryWrites)
	return item
}

func memoryArtifactsOrEmpty(items []harness.MemoryArtifact) []harness.MemoryArtifact {
	if items == nil {
		return []harness.MemoryArtifact{}
	}
	return items
}

func ptrTimeValue(value time.Time) *time.Time {
	return &value
}
