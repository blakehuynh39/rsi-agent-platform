package companyknowledge

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type fakeHonchoDocuments struct {
	ensureWorkspaceCalls int
	ensureSessionCalls   int
	messageCalls         int
	createCalls          int
	createErr            error
	lastMessages         []clients.HonchoMessageCreate
	lastConclusions      []clients.HonchoConclusionCreate
}

func (f *fakeHonchoDocuments) EnsureWorkspace(id string, metadata map[string]any) (clients.HonchoWorkspace, error) {
	f.ensureWorkspaceCalls++
	return clients.HonchoWorkspace{ID: id, Metadata: metadata}, nil
}

func (f *fakeHonchoDocuments) EnsureSession(workspaceID string, sessionID string, metadata map[string]any) (clients.HonchoSession, error) {
	f.ensureSessionCalls++
	return clients.HonchoSession{ID: sessionID, WorkspaceID: workspaceID, Metadata: metadata}, nil
}

func (f *fakeHonchoDocuments) CreateMessages(workspaceID string, sessionID string, messages []clients.HonchoMessageCreate) ([]clients.HonchoMessage, error) {
	f.messageCalls++
	f.lastMessages = append([]clients.HonchoMessageCreate(nil), messages...)
	out := make([]clients.HonchoMessage, 0, len(messages))
	for i := range messages {
		out = append(out, clients.HonchoMessage{ID: fmt.Sprintf("msg_notion_chunk_%03d", i), WorkspaceID: workspaceID, SessionID: sessionID})
	}
	return out, nil
}

func (f *fakeHonchoDocuments) CreateConclusions(workspaceID string, conclusions []clients.HonchoConclusionCreate) ([]clients.HonchoConclusion, error) {
	f.createCalls++
	f.lastConclusions = append([]clients.HonchoConclusionCreate(nil), conclusions...)
	if f.createErr != nil {
		return nil, f.createErr
	}
	return []clients.HonchoConclusion{
		{
			ID:         "doc_notion_1",
			Content:    conclusions[0].Content,
			ObserverID: conclusions[0].ObserverID,
			ObservedID: conclusions[0].ObservedID,
			SessionID:  conclusions[0].SessionID,
		},
	}, nil
}

func TestNotionMirrorIngestDocumentChunksLargePagesBeforeSummaryConclusion(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeHonchoDocuments{}
	mirror := NewNotionMirror(state, honcho, NotionMirrorOptions{
		Environment:     "stage",
		HonchoWorkspace: "rsi_company_knowledge",
	})
	large := strings.Repeat("Long Notion runbook paragraph with deployment notes.\n", 900)
	result, err := mirror.IngestDocument(nil, NotionDocumentInput{
		WorkspaceID:    "notion",
		PageID:         "page_large",
		RootID:         "root_abc",
		Title:          "Large Runbook",
		LastEditedTime: "2026-05-02T10:00:00.000Z",
		Content:        large,
	})
	if err != nil {
		t.Fatalf("IngestDocument() error = %v", err)
	}
	if result.Skipped {
		t.Fatalf("large page should mirror successfully through chunks: %+v", result)
	}
	if honcho.messageCalls != 1 || len(honcho.lastMessages) < 2 {
		t.Fatalf("expected large page to be split into multiple Honcho messages, calls=%d messages=%d", honcho.messageCalls, len(honcho.lastMessages))
	}
	if honcho.createCalls != 1 || len(honcho.lastConclusions) != 1 {
		t.Fatalf("expected one compact summary conclusion, calls=%d conclusions=%d", honcho.createCalls, len(honcho.lastConclusions))
	}
	if len(honcho.lastConclusions[0].Content) > 12000 {
		t.Fatalf("summary conclusion should stay compact, length=%d", len(honcho.lastConclusions[0].Content))
	}
}

func TestNotionDocumentChunkMessagesReturnsEmptyForEmptyContentAndTitle(t *testing.T) {
	messages := NotionDocumentChunkMessages(NotionDocumentInput{
		WorkspaceID: "notion",
		PageID:      "page_empty",
		RootID:      "root_abc",
	}, map[string]any{"source": "notion"})
	if len(messages) != 0 {
		t.Fatalf("empty content/title should not create an empty Honcho message batch: %+v", messages)
	}
}

func TestNotionDocumentChunkMessagesBoundsUTF8Bytes(t *testing.T) {
	messages := NotionDocumentChunkMessages(NotionDocumentInput{
		WorkspaceID: "notion",
		PageID:      "page_emoji",
		RootID:      "root_abc",
		Content:     strings.Repeat("🧠", 7000),
	}, map[string]any{"source": "notion"})
	if len(messages) < 2 {
		t.Fatalf("expected multi-byte content to be split by byte budget, got %d messages", len(messages))
	}
	for _, message := range messages {
		if len(message.Content) > notionHonchoMessageMaxBytes {
			t.Fatalf("message content bytes = %d, want <= %d", len(message.Content), notionHonchoMessageMaxBytes)
		}
		if !utf8.ValidString(message.Content) {
			t.Fatalf("message content is not valid UTF-8")
		}
	}
}

func TestNotionMirrorIngestDocumentTreatsHonchoValidationAsObjectFailure(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeHonchoDocuments{
		createErr: &clients.HTTPStatusError{Service: "honcho", StatusCode: 422, Body: "content too large"},
	}
	mirror := NewNotionMirror(state, honcho, NotionMirrorOptions{
		Environment:     "stage",
		HonchoWorkspace: "rsi_company_knowledge",
	})

	result, err := mirror.IngestDocument(nil, NotionDocumentInput{
		WorkspaceID: "notion",
		PageID:      "page_oversized",
		RootID:      "root_abc",
		Title:       "Oversized Runbook",
		Content:     "too large for honcho",
	})
	if err != nil {
		t.Fatalf("IngestDocument() should not fail the mirror run for a per-document 422: %v", err)
	}
	if result.Status != store.SourceMirrorStatusFailed || !result.Skipped || result.SkipReason != "honcho_validation_failed" {
		t.Fatalf("unexpected validation failure result: %+v", result)
	}
	record, found, err := state.GetSourceMirrorRecord(NotionDocumentSourceType, NotionDocumentSourceKey("notion", "page_oversized"))
	if err != nil {
		t.Fatalf("GetSourceMirrorRecord() error = %v", err)
	}
	if !found || record.Status != store.SourceMirrorStatusFailed || !strings.Contains(record.LastError, "content too large") {
		t.Fatalf("unexpected failed source mirror record: found=%t record=%+v", found, record)
	}
}

func TestNotionMirrorIngestDocumentIsIdempotent(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeHonchoDocuments{}
	mirror := NewNotionMirror(state, honcho, NotionMirrorOptions{
		Environment:     "stage",
		HonchoWorkspace: "rsi_company_knowledge",
	})
	input := NotionDocumentInput{
		WorkspaceID:    "notion",
		PageID:         "page_abc",
		RootID:         "root_abc",
		Title:          "Runbook",
		URL:            "https://notion.so/page_abc",
		LastEditedTime: "2026-05-02T10:00:00.000Z",
		Content:        "Steps to debug prod.",
	}

	first, err := mirror.IngestDocument(nil, input)
	if err != nil {
		t.Fatalf("first IngestDocument() error = %v", err)
	}
	if first.Skipped || first.HonchoDocumentID != "doc_notion_1" {
		t.Fatalf("unexpected first result: %+v", first)
	}
	second, err := mirror.IngestDocument(nil, input)
	if err != nil {
		t.Fatalf("second IngestDocument() error = %v", err)
	}
	if !second.Skipped || second.SkipReason != "already_complete" || second.HonchoDocumentID != "doc_notion_1" {
		t.Fatalf("unexpected second result: %+v", second)
	}
	if honcho.createCalls != 1 {
		t.Fatalf("CreateConclusions calls = %d, want 1", honcho.createCalls)
	}
	record, found, err := state.GetSourceMirrorRecord(NotionDocumentSourceType, NotionDocumentSourceKey("notion", "page_abc"))
	if err != nil {
		t.Fatalf("GetSourceMirrorRecord() error = %v", err)
	}
	if !found || record.HonchoObjectType != "document" || record.HonchoObjectID != "doc_notion_1" {
		t.Fatalf("unexpected source mirror record: found=%t record=%+v", found, record)
	}
}

func TestNotionDatabaseSchemaHashIsDeterministic(t *testing.T) {
	left := map[string]any{
		"Status": map[string]any{
			"type": "status",
			"status": map[string]any{
				"options": []any{map[string]any{"name": "Done"}, map[string]any{"name": "Todo"}},
			},
		},
		"Owner": map[string]any{"type": "people", "people": map[string]any{}},
	}
	right := map[string]any{
		"Owner": map[string]any{"type": "people", "people": map[string]any{}},
		"Status": map[string]any{
			"status": map[string]any{
				"options": []any{map[string]any{"name": "Todo"}, map[string]any{"name": "Done"}},
			},
			"type": "status",
		},
	}
	leftSummary, leftHash := NotionDatabaseSchemaSummary(left)
	rightSummary, rightHash := NotionDatabaseSchemaSummary(right)
	if leftHash == "" || leftHash != rightHash {
		t.Fatalf("schema hashes differ: %q vs %q", leftHash, rightHash)
	}
	if leftSummary != rightSummary {
		t.Fatalf("schema summaries differ:\nleft=%s\nright=%s", leftSummary, rightSummary)
	}
}
