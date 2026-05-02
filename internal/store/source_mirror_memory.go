package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

func (m *MemoryStore) ClaimSourceMirrorRecord(record SourceMirrorRecord, lease time.Duration) (SourceMirrorClaimResult, error) {
	if err := validateSourceMirrorRecord(record); err != nil {
		return SourceMirrorClaimResult{}, err
	}
	if lease <= 0 {
		lease = 5 * time.Minute
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sourceMirrorRecords == nil {
		m.sourceMirrorRecords = map[string]SourceMirrorRecord{}
	}
	key := sourceMirrorMemoryKey(record.SourceType, record.SourceKey)
	now := time.Now().UTC()
	existing, found := m.sourceMirrorRecords[key]
	if !found {
		record.Status = SourceMirrorStatusPending
		record.HonchoMessageID = ""
		record.LastError = ""
		record.CreatedAt = now
		record.UpdatedAt = now
		record.Metadata = cloneAnyMap(record.Metadata)
		m.sourceMirrorRecords[key] = record
		return SourceMirrorClaimResult{Record: record, ShouldWrite: true, Reason: "new"}, nil
	}
	if existing.Status == SourceMirrorStatusComplete && existing.SourceRevision == record.SourceRevision && strings.TrimSpace(existing.HonchoMessageID) != "" {
		return SourceMirrorClaimResult{Record: cloneSourceMirrorRecord(existing), ShouldWrite: false, Reason: "already_complete"}, nil
	}
	if existing.Status == SourceMirrorStatusPending && now.Sub(existing.UpdatedAt) < lease {
		return SourceMirrorClaimResult{Record: cloneSourceMirrorRecord(existing), ShouldWrite: false, Reason: "leased"}, nil
	}
	reason := "retry"
	if existing.SourceRevision != record.SourceRevision {
		reason = "revision_changed"
	}
	record.Status = SourceMirrorStatusPending
	record.HonchoMessageID = ""
	record.LastError = ""
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = now
	record.Metadata = cloneAnyMap(record.Metadata)
	m.sourceMirrorRecords[key] = record
	return SourceMirrorClaimResult{Record: record, ShouldWrite: true, Reason: reason}, nil
}

func (m *MemoryStore) CompleteSourceMirrorRecord(sourceType string, sourceKey string, honchoMessageID string, metadata map[string]any) (SourceMirrorRecord, error) {
	honchoMessageID = strings.TrimSpace(honchoMessageID)
	if honchoMessageID == "" {
		return SourceMirrorRecord{}, errors.New("honcho message id is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sourceMirrorMemoryKey(sourceType, sourceKey)
	existing, found := m.sourceMirrorRecords[key]
	if !found {
		return SourceMirrorRecord{}, sql.ErrNoRows
	}
	existing.Status = SourceMirrorStatusComplete
	existing.HonchoMessageID = honchoMessageID
	existing.LastError = ""
	existing.UpdatedAt = time.Now().UTC()
	existing.Metadata = mergeAnyMaps(existing.Metadata, metadata)
	m.sourceMirrorRecords[key] = existing
	return cloneSourceMirrorRecord(existing), nil
}

func (m *MemoryStore) FailSourceMirrorRecord(sourceType string, sourceKey string, lastError string, metadata map[string]any) (SourceMirrorRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sourceMirrorMemoryKey(sourceType, sourceKey)
	existing, found := m.sourceMirrorRecords[key]
	if !found {
		return SourceMirrorRecord{}, sql.ErrNoRows
	}
	existing.Status = SourceMirrorStatusFailed
	existing.LastError = strings.TrimSpace(lastError)
	existing.UpdatedAt = time.Now().UTC()
	existing.Metadata = mergeAnyMaps(existing.Metadata, metadata)
	m.sourceMirrorRecords[key] = existing
	return cloneSourceMirrorRecord(existing), nil
}

func (m *MemoryStore) GetSourceMirrorRecord(sourceType string, sourceKey string) (SourceMirrorRecord, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.sourceMirrorRecords == nil {
		return SourceMirrorRecord{}, false, nil
	}
	record, found := m.sourceMirrorRecords[sourceMirrorMemoryKey(sourceType, sourceKey)]
	if !found {
		return SourceMirrorRecord{}, false, nil
	}
	return cloneSourceMirrorRecord(record), true, nil
}

func sourceMirrorMemoryKey(sourceType string, sourceKey string) string {
	return strings.TrimSpace(sourceType) + "\x00" + strings.TrimSpace(sourceKey)
}

func cloneSourceMirrorRecord(record SourceMirrorRecord) SourceMirrorRecord {
	record.Metadata = cloneAnyMap(record.Metadata)
	return record
}

func mergeAnyMaps(base map[string]any, overlay map[string]any) map[string]any {
	out := cloneAnyMap(base)
	for key, value := range overlay {
		out[key] = value
	}
	return out
}

func cloneAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
