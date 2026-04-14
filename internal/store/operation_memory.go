package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
)

func (s *MemoryStore) ListOperations() []operation.Execution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]operation.Execution, 0, len(s.operations))
	for _, item := range s.operations {
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

func (s *MemoryStore) GetOperation(operationID string) (operation.Execution, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.operations[strings.TrimSpace(operationID)]
	return item, ok
}

func (s *MemoryStore) ListOperationsByScope(scopeKind operation.ScopeKind, scopeID string) []operation.Execution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []operation.Execution{}
	for _, item := range s.operations {
		if item.ScopeKind == scopeKind && item.ScopeID == strings.TrimSpace(scopeID) {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *MemoryStore) GetOrCreateOperation(item operation.Execution) (operation.Execution, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getOrCreateOperationLocked(item)
}

func (s *MemoryStore) getOrCreateOperationLocked(item operation.Execution) (operation.Execution, bool, error) {
	normalized, err := normalizeOperationExecution(item)
	if err != nil {
		return operation.Execution{}, false, err
	}
	if existing, ok := findOperationLocked(s.operations, normalized.ScopeKind, normalized.ScopeID, normalized.OperationKind, normalized.OperationKey); ok {
		return existing, false, nil
	}
	s.operations[normalized.ID] = normalized
	return normalized, true, nil
}

func (s *MemoryStore) ClaimOperation(operationID string, holder string) (operation.Execution, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.operations[strings.TrimSpace(operationID)]
	if !ok {
		return operation.Execution{}, false, errors.New("operation not found")
	}
	switch item.Status {
	case operation.StatusQueued, operation.StatusFailed:
	default:
		return item, false, nil
	}
	now := time.Now().UTC()
	item.Status = operation.StatusRunning
	item.Holder = strings.TrimSpace(holder)
	item.UpdatedAt = now
	if item.StartedAt == nil {
		item.StartedAt = &now
	}
	s.operations[item.ID] = item
	return item, true, nil
}

func (s *MemoryStore) CompleteOperation(operationID string, resultRef string) (operation.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.operations[strings.TrimSpace(operationID)]
	if !ok {
		return operation.Execution{}, errors.New("operation not found")
	}
	now := time.Now().UTC()
	item.Status = operation.StatusCompleted
	item.ResultRef = strings.TrimSpace(resultRef)
	item.LastError = ""
	item.UpdatedAt = now
	item.CompletedAt = &now
	s.operations[item.ID] = item
	return item, nil
}

func (s *MemoryStore) FailOperation(operationID string, lastError string) (operation.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.operations[strings.TrimSpace(operationID)]
	if !ok {
		return operation.Execution{}, errors.New("operation not found")
	}
	now := time.Now().UTC()
	item.Status = operation.StatusFailed
	item.LastError = strings.TrimSpace(lastError)
	item.RetryCount++
	item.UpdatedAt = now
	item.CompletedAt = &now
	s.operations[item.ID] = item
	return item, nil
}

func normalizeOperationExecution(item operation.Execution) (operation.Execution, error) {
	now := time.Now().UTC()
	item.ScopeID = strings.TrimSpace(item.ScopeID)
	item.OperationKind = strings.TrimSpace(item.OperationKind)
	item.OperationKey = strings.TrimSpace(item.OperationKey)
	if item.ScopeKind == "" || item.ScopeID == "" || item.OperationKind == "" || item.OperationKey == "" {
		return operation.Execution{}, fmt.Errorf("operation scope_kind, scope_id, operation_kind, and operation_key are required")
	}
	if item.ID == "" {
		item.ID = operationExecutionID(item.ScopeKind, item.ScopeID, item.OperationKind, item.OperationKey)
	}
	if item.Status == "" {
		item.Status = operation.StatusQueued
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = item.CreatedAt
	}
	item.RequestedBy = strings.TrimSpace(item.RequestedBy)
	item.Holder = strings.TrimSpace(item.Holder)
	item.TraceID = strings.TrimSpace(item.TraceID)
	item.ProposalID = strings.TrimSpace(item.ProposalID)
	item.AttemptID = strings.TrimSpace(item.AttemptID)
	item.PayloadHash = strings.TrimSpace(item.PayloadHash)
	item.ResultRef = strings.TrimSpace(item.ResultRef)
	item.LastError = strings.TrimSpace(item.LastError)
	return item, nil
}

func operationExecutionID(scopeKind operation.ScopeKind, scopeID string, operationKind string, operationKey string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s", scopeKind, strings.TrimSpace(scopeID), strings.TrimSpace(operationKind), strings.TrimSpace(operationKey))))
	return "op-" + hex.EncodeToString(sum[:16])
}

func findOperationLocked(items map[string]operation.Execution, scopeKind operation.ScopeKind, scopeID string, operationKind string, operationKey string) (operation.Execution, bool) {
	for _, item := range items {
		if item.ScopeKind == scopeKind &&
			item.ScopeID == scopeID &&
			item.OperationKind == operationKind &&
			item.OperationKey == operationKey {
			return item, true
		}
	}
	return operation.Execution{}, false
}

func (s *MemoryStore) ensureOperationWorkItemLocked(op operation.Execution, item queue.WorkItem) (operation.Execution, queue.WorkItem, bool, error) {
	createdOp, _, err := s.getOrCreateOperationLocked(op)
	if err != nil {
		return operation.Execution{}, queue.WorkItem{}, false, err
	}
	now := time.Now().UTC()
	if createdOp.Status == operation.StatusFailed && item.Status == queue.WorkQueued {
		createdOp.Status = operation.StatusQueued
		createdOp.Holder = ""
		createdOp.LastError = ""
		createdOp.UpdatedAt = now
		createdOp.CompletedAt = nil
		if op.PayloadHash != "" {
			createdOp.PayloadHash = op.PayloadHash
		}
		s.operations[createdOp.ID] = createdOp
	}
	item.OperationID = createdOp.ID
	for _, existing := range s.workItems {
		if existing.OperationID == createdOp.ID {
			if item.Status == queue.WorkQueued && (existing.Status == queue.WorkFailed || existing.Status == queue.WorkCanceled) {
				existing.Status = queue.WorkQueued
				existing.Payload = cloneMetadata(item.Payload)
				if existing.Payload == nil {
					existing.Payload = map[string]interface{}{}
				}
				existing.LastError = ""
				existing.LeaseOwner = ""
				existing.LeaseExpiresAt = nil
				existing.CompletedAt = nil
				existing.UpdatedAt = now
				s.workItems[existing.ID] = existing
			}
			return createdOp, existing, false, nil
		}
	}
	createdItem, err := s.enqueueWorkItemLocked(item)
	if err != nil {
		return operation.Execution{}, queue.WorkItem{}, false, err
	}
	return createdOp, createdItem, true, nil
}
