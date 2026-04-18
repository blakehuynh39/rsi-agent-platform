package store

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func (s *MemoryStore) ListDomainEvents() []transition.DomainEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]transition.DomainEvent(nil), s.domainEvents...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) ListEffectExecutions() []transition.EffectExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return listEffectExecutionsLocked(s.effectExecutions, "", "")
}

func (s *MemoryStore) ListEffectExecutionsByAggregate(machineKind transition.MachineKind, aggregateID string) []transition.EffectExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return listEffectExecutionsLocked(s.effectExecutions, machineKind, aggregateID)
}

func listEffectExecutionsLocked(items map[string]transition.EffectExecution, machineKind transition.MachineKind, aggregateID string) []transition.EffectExecution {
	out := []transition.EffectExecution{}
	aggregateID = strings.TrimSpace(aggregateID)
	for _, item := range items {
		if machineKind != "" && item.MachineKind != machineKind {
			continue
		}
		if aggregateID != "" && strings.TrimSpace(item.AggregateID) != aggregateID {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *MemoryStore) GetCommandReceipt(commandID string) (transition.CommandReceipt, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.commandReceipts[strings.TrimSpace(commandID)]
	return item, ok
}

func (s *MemoryStore) RecordCommandReceipt(item transition.CommandReceipt) (transition.CommandReceipt, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recordCommandReceiptLocked(item)
}

func (s *MemoryStore) QueueEffectExecution(effect transition.EffectExecution) (transition.EffectExecution, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	effect.ID = strings.TrimSpace(effect.ID)
	effect.AggregateID = strings.TrimSpace(effect.AggregateID)
	effect.AttemptID = strings.TrimSpace(effect.AttemptID)
	effect.IdempotencyKey = strings.TrimSpace(effect.IdempotencyKey)
	if effect.ID == "" {
		return transition.EffectExecution{}, false, errors.New("effect execution id is required")
	}
	if effect.MachineKind == "" {
		return transition.EffectExecution{}, false, errors.New("machine kind is required")
	}
	if effect.AggregateID == "" {
		return transition.EffectExecution{}, false, errors.New("aggregate id is required")
	}
	if effect.EffectKind == "" {
		return transition.EffectExecution{}, false, errors.New("effect kind is required")
	}
	if effect.IdempotencyKey == "" {
		return transition.EffectExecution{}, false, errors.New("idempotency key is required")
	}
	for _, existing := range s.effectExecutions {
		if existing.IdempotencyKey == effect.IdempotencyKey {
			return existing, false, nil
		}
	}
	now := time.Now().UTC()
	if effect.Status == "" {
		effect.Status = transition.EffectQueued
	}
	if effect.Payload == nil {
		effect.Payload = map[string]any{}
	}
	if effect.CreatedAt.IsZero() {
		effect.CreatedAt = now
	}
	if effect.UpdatedAt.IsZero() || effect.UpdatedAt.Before(effect.CreatedAt) {
		effect.UpdatedAt = effect.CreatedAt
	}
	s.effectExecutions[effect.ID] = effect
	return effect, true, nil
}

func (s *MemoryStore) recordCommandReceiptLocked(item transition.CommandReceipt) (transition.CommandReceipt, bool, error) {
	item.CommandID = strings.TrimSpace(item.CommandID)
	if item.CommandID == "" {
		return transition.CommandReceipt{}, false, errors.New("command_id is required")
	}
	if existing, ok := s.commandReceipts[item.CommandID]; ok {
		return existing, false, nil
	}
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = item.CreatedAt
	}
	s.commandReceipts[item.CommandID] = item
	return item, true, nil
}

func (s *MemoryStore) ClaimEffectExecution(effectID string, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.effectExecutions[strings.TrimSpace(effectID)]
	if !ok {
		return transition.EffectExecution{}, false, errors.New("effect execution not found")
	}
	now := time.Now().UTC()
	switch item.Status {
	case transition.EffectQueued, transition.EffectFailed:
	case transition.EffectRunning:
		if item.LeaseExpiresAt == nil || item.LeaseExpiresAt.After(now) {
			return item, false, nil
		}
	default:
		return item, false, nil
	}
	item.Status = transition.EffectRunning
	item.Holder = strings.TrimSpace(holder)
	item.UpdatedAt = now
	if item.StartedAt == nil {
		item.StartedAt = &now
	}
	if lease > 0 {
		expires := now.Add(lease)
		item.LeaseExpiresAt = &expires
	} else {
		item.LeaseExpiresAt = nil
	}
	item.CompletedAt = nil
	s.effectExecutions[item.ID] = item
	return item, true, nil
}

func (s *MemoryStore) CompleteEffectExecution(effectID string, holder string, resultRef string) (transition.EffectExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.effectExecutions[strings.TrimSpace(effectID)]
	if !ok {
		return transition.EffectExecution{}, errors.New("effect execution not found")
	}
	switch item.Status {
	case transition.EffectCompleted, transition.EffectCanceled, transition.EffectSuperseded:
		return item, nil
	}
	if !effectExecutionLeaseHeldBy(item, holder, time.Now().UTC()) {
		return item, nil
	}
	now := time.Now().UTC()
	item.Status = transition.EffectCompleted
	item.ResultRef = strings.TrimSpace(resultRef)
	item.LastError = ""
	item.Holder = ""
	item.UpdatedAt = now
	item.LeaseExpiresAt = nil
	item.CompletedAt = &now
	s.effectExecutions[item.ID] = item
	return item, nil
}

func (s *MemoryStore) FailEffectExecution(effectID string, holder string, lastError string) (transition.EffectExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.effectExecutions[strings.TrimSpace(effectID)]
	if !ok {
		return transition.EffectExecution{}, errors.New("effect execution not found")
	}
	switch item.Status {
	case transition.EffectCompleted, transition.EffectCanceled, transition.EffectSuperseded:
		return item, nil
	}
	if !effectExecutionLeaseHeldBy(item, holder, time.Now().UTC()) {
		return item, nil
	}
	now := time.Now().UTC()
	item.Status = transition.EffectFailed
	item.LastError = strings.TrimSpace(lastError)
	item.Holder = ""
	item.RetryCount++
	item.UpdatedAt = now
	item.LeaseExpiresAt = nil
	item.CompletedAt = &now
	s.effectExecutions[item.ID] = item
	return item, nil
}

func effectExecutionLeaseHeldBy(item transition.EffectExecution, holder string, now time.Time) bool {
	if item.Status != transition.EffectRunning {
		return false
	}
	if strings.TrimSpace(item.Holder) == "" || strings.TrimSpace(item.Holder) != strings.TrimSpace(holder) {
		return false
	}
	return item.LeaseExpiresAt == nil || item.LeaseExpiresAt.After(now)
}

func (s *MemoryStore) appendTransitionBundleLocked(bundle transitionPersistBundle) {
	if len(bundle.Events) > 0 {
		s.domainEvents = append(s.domainEvents, bundle.Events...)
	}
	for _, item := range bundle.Effects {
		s.effectExecutions[item.ID] = item
	}
}
