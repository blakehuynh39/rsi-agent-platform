package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/operation"
)

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

func shouldRequeueOperationForMissingWorkItem(op operation.Execution, reopenTerminal bool) bool {
	switch op.Status {
	case operation.StatusFailed:
		return true
	case operation.StatusRunning, operation.StatusCompleted:
		return reopenTerminal
	default:
		return false
	}
}
