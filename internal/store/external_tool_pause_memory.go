package store

import (
	"errors"
	"sort"
	"strings"
	"time"
)

func (s *MemoryStore) ListExternalToolPauses() []ExternalToolPause {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ExternalToolPause, 0, len(s.externalToolPauses))
	for _, item := range s.externalToolPauses {
		out = append(out, cloneExternalToolPause(item))
	}
	sortExternalToolPauses(out)
	return out
}

func (s *MemoryStore) GetExternalToolPause(id string) (ExternalToolPause, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.externalToolPauses[strings.TrimSpace(id)]
	return cloneExternalToolPause(item), ok
}

func (s *MemoryStore) GetExternalToolPauseByDBReadRequestID(requestID string) (ExternalToolPause, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return ExternalToolPause{}, false
	}
	var found ExternalToolPause
	ok := false
	for _, item := range s.externalToolPauses {
		if strings.TrimSpace(item.DBReadRequestID) != requestID {
			continue
		}
		if !ok || item.CreatedAt.After(found.CreatedAt) {
			found = item
			ok = true
		}
	}
	return cloneExternalToolPause(found), ok
}

func (s *MemoryStore) UpsertExternalToolPause(input ExternalToolPauseCreateInput, now time.Time) (ExternalToolPause, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.TrimSpace(input.IdempotencyKey)
	if id, ok := s.externalToolPauseByIdempotencyKey[key]; ok {
		item, ok := s.externalToolPauses[id]
		if ok {
			return cloneExternalToolPause(item), false, nil
		}
	}
	item, err := NewExternalToolPause(input, now)
	if err != nil {
		return ExternalToolPause{}, false, err
	}
	s.externalToolPauses[item.ID] = item
	s.externalToolPauseByIdempotencyKey[item.IdempotencyKey] = item.ID
	return cloneExternalToolPause(item), true, nil
}

func (s *MemoryStore) UpdateExternalToolPause(id string, mutate func(*ExternalToolPause) error) (ExternalToolPause, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.externalToolPauses[strings.TrimSpace(id)]
	if !ok {
		return ExternalToolPause{}, errors.New("external tool pause not found")
	}
	if mutate != nil {
		if err := mutate(&item); err != nil {
			return ExternalToolPause{}, err
		}
	}
	item.UpdatedAt = time.Now().UTC()
	s.externalToolPauses[item.ID] = item
	return cloneExternalToolPause(item), nil
}

func sortExternalToolPauses(items []ExternalToolPause) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}
