package store

import (
	"testing"
	"time"
)

func TestMemoryStoreMarkSourceMirrorRecordStaleAdoptsIncomingHonchoFieldsWhenExistingBlank(t *testing.T) {
	state := NewMemoryStore()
	_, err := state.ClaimSourceMirrorRecord(sourceMirrorRecordForTest("rev-1"), time.Minute)
	if err != nil {
		t.Fatalf("ClaimSourceMirrorRecord() error = %v", err)
	}

	stale := sourceMirrorRecordForTest("rev-2")
	stale.HonchoMessageID = " message-123 "
	stale.HonchoObjectType = " message "
	stale.HonchoObjectID = " message-123 "

	record, err := state.MarkSourceMirrorRecordStale(stale, "not visible", nil)
	if err != nil {
		t.Fatalf("MarkSourceMirrorRecordStale() error = %v", err)
	}
	if record.HonchoMessageID != "message-123" {
		t.Fatalf("honcho message id = %q, want %q", record.HonchoMessageID, "message-123")
	}
	if record.HonchoObjectType != "message" {
		t.Fatalf("honcho object type = %q, want %q", record.HonchoObjectType, "message")
	}
	if record.HonchoObjectID != "message-123" {
		t.Fatalf("honcho object id = %q, want %q", record.HonchoObjectID, "message-123")
	}
}

func TestMemoryStoreMarkSourceMirrorRecordStalePreservesExistingNonEmptyHonchoFields(t *testing.T) {
	state := NewMemoryStore()
	base := sourceMirrorRecordForTest("rev-1")
	_, err := state.ClaimSourceMirrorRecord(base, time.Minute)
	if err != nil {
		t.Fatalf("ClaimSourceMirrorRecord() error = %v", err)
	}
	_, err = state.CompleteSourceMirrorObject(base.SourceType, base.SourceKey, "message", "message-123", nil)
	if err != nil {
		t.Fatalf("CompleteSourceMirrorObject() error = %v", err)
	}

	stale := sourceMirrorRecordForTest("rev-2")
	stale.HonchoMessageID = "message-456"
	stale.HonchoObjectType = "page"
	stale.HonchoObjectID = "page-456"

	record, err := state.MarkSourceMirrorRecordStale(stale, "changed", nil)
	if err != nil {
		t.Fatalf("MarkSourceMirrorRecordStale() error = %v", err)
	}
	if record.HonchoMessageID != "message-123" {
		t.Fatalf("honcho message id = %q, want %q", record.HonchoMessageID, "message-123")
	}
	if record.HonchoObjectType != "message" {
		t.Fatalf("honcho object type = %q, want %q", record.HonchoObjectType, "message")
	}
	if record.HonchoObjectID != "message-123" {
		t.Fatalf("honcho object id = %q, want %q", record.HonchoObjectID, "message-123")
	}
}

func sourceMirrorRecordForTest(revision string) SourceMirrorRecord {
	return SourceMirrorRecord{
		SourceType:       "notion_page",
		SourceKey:        "page-abc",
		Workspace:        "workspace-1",
		Environment:      "test",
		SourceSessionKey: "source-session-1",
		HonchoWorkspace:  "honcho-workspace-1",
		HonchoSessionID:  "honcho-session-1",
		SourceRevision:   revision,
		Metadata:         map[string]any{"revision": revision},
	}
}
