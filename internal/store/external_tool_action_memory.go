package store

import (
	"errors"
	"time"
)

func (s *MemoryStore) ListExternalToolActions() []ExternalToolAction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ExternalToolAction, 0, len(s.externalToolActions))
	for _, item := range s.externalToolActions {
		out = append(out, cloneExternalToolAction(item))
	}
	SortExternalToolActions(out)
	return out
}

func (s *MemoryStore) GetExternalToolAction(actionID string) (ExternalToolAction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.externalToolActions[actionID]
	return cloneExternalToolAction(item), ok
}

func (s *MemoryStore) GetExternalToolActionByIdempotency(surface string, operation string, idempotencyKey string) (ExternalToolAction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.externalToolActionByIdempotency[ExternalToolActionIdempotencyKey(surface, operation, idempotencyKey)]
	if !ok {
		return ExternalToolAction{}, false
	}
	item, ok := s.externalToolActions[id]
	return cloneExternalToolAction(item), ok
}

func (s *MemoryStore) UpsertExternalToolAction(input ExternalToolActionCreateInput, now time.Time) (ExternalToolAction, ExternalToolActionUpsertStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := ExternalToolActionIdempotencyKey(input.Surface, input.Operation, input.IdempotencyKey)
	if id, ok := s.externalToolActionByIdempotency[key]; ok {
		item, ok := s.externalToolActions[id]
		if ok {
			status := ExternalToolActionUpsertReplay
			if item.RequestHash != input.RequestHash {
				status = ExternalToolActionUpsertConflict
			}
			return cloneExternalToolAction(item), status, nil
		}
	}
	item, err := NewExternalToolAction(input, now)
	if err != nil {
		return ExternalToolAction{}, "", err
	}
	s.externalToolActions[item.ID] = item
	s.externalToolActionByIdempotency[key] = item.ID
	return cloneExternalToolAction(item), ExternalToolActionUpsertCreated, nil
}

func (s *MemoryStore) UpdateExternalToolActionResult(actionID string, update ExternalToolActionResultUpdate, now time.Time) (ExternalToolAction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.externalToolActions[actionID]
	if !ok {
		return ExternalToolAction{}, errors.New("external tool action not found")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if update.State != "" {
		item.State = update.State
	}
	item.ResponseSummary = update.ResponseSummary
	item.ErrorMessage = update.ErrorMessage
	item.SourceRef = update.SourceRef
	item.WikiAuditID = update.WikiAuditID
	item.ResultPayload = CloneJSONMap(update.ResultPayload)
	item.MirrorEffect = CloneJSONMap(update.MirrorEffect)
	item.UpdatedAt = now
	if item.State == ExternalToolActionStateSucceeded || item.State == ExternalToolActionStateFailed {
		completed := now
		item.CompletedAt = &completed
	}
	s.externalToolActions[item.ID] = item
	return cloneExternalToolAction(item), nil
}
