package store

import (
	"database/sql"
	"errors"
	"sort"
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
		record.HonchoObjectType = ""
		record.HonchoObjectID = ""
		record.LastError = ""
		record.CreatedAt = now
		record.UpdatedAt = now
		record.Metadata = cloneAnyMap(record.Metadata)
		m.sourceMirrorRecords[key] = record
		return SourceMirrorClaimResult{Record: record, ShouldWrite: true, Reason: "new"}, nil
	}
	if existing.Status == SourceMirrorStatusComplete && existing.SourceRevision == record.SourceRevision && sourceMirrorRecordHasHonchoObject(existing) {
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
	record.HonchoObjectType = ""
	record.HonchoObjectID = ""
	record.LastError = ""
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = now
	record.Metadata = cloneAnyMap(record.Metadata)
	m.sourceMirrorRecords[key] = record
	return SourceMirrorClaimResult{Record: record, ShouldWrite: true, Reason: reason}, nil
}

func (m *MemoryStore) CompleteSourceMirrorRecord(sourceType string, sourceKey string, honchoMessageID string, metadata map[string]any) (SourceMirrorRecord, error) {
	return m.CompleteSourceMirrorObject(sourceType, sourceKey, "message", honchoMessageID, metadata)
}

func (m *MemoryStore) CompleteSourceMirrorObject(sourceType string, sourceKey string, honchoObjectType string, honchoObjectID string, metadata map[string]any) (SourceMirrorRecord, error) {
	honchoObjectType = strings.TrimSpace(honchoObjectType)
	honchoObjectID = strings.TrimSpace(honchoObjectID)
	if honchoObjectType == "" {
		return SourceMirrorRecord{}, errors.New("honcho object type is required")
	}
	if honchoObjectID == "" {
		return SourceMirrorRecord{}, errors.New("honcho object id is required")
	}
	honchoMessageID := ""
	if honchoObjectType == "message" {
		honchoMessageID = honchoObjectID
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
	existing.HonchoObjectType = honchoObjectType
	existing.HonchoObjectID = honchoObjectID
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

func (m *MemoryStore) ListSourceMirrorRecords(sourceTypes []string, limit int) ([]SourceMirrorRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	wanted := sourceMirrorTypeSet(sourceTypes)
	records := make([]SourceMirrorRecord, 0, len(m.sourceMirrorRecords))
	for _, record := range m.sourceMirrorRecords {
		if len(wanted) > 0 {
			if _, ok := wanted[strings.TrimSpace(record.SourceType)]; !ok {
				continue
			}
		}
		records = append(records, cloneSourceMirrorRecord(record))
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].UpdatedAt.After(records[j].UpdatedAt)
	})
	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}
	return records, nil
}

func sourceMirrorMemoryKey(sourceType string, sourceKey string) string {
	return strings.TrimSpace(sourceType) + "\x00" + strings.TrimSpace(sourceKey)
}

func sourceMirrorTypeSet(sourceTypes []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, sourceType := range sourceTypes {
		sourceType = strings.TrimSpace(sourceType)
		if sourceType != "" {
			out[sourceType] = struct{}{}
		}
	}
	return out
}

func cloneSourceMirrorRecord(record SourceMirrorRecord) SourceMirrorRecord {
	record.Metadata = cloneAnyMap(record.Metadata)
	return record
}

func sourceMirrorRecordHasHonchoObject(record SourceMirrorRecord) bool {
	return strings.TrimSpace(record.HonchoObjectID) != "" || strings.TrimSpace(record.HonchoMessageID) != ""
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
