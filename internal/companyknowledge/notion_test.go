package companyknowledge

import (
	"strings"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type fakeHonchoDocuments struct {
	ensureWorkspaceCalls int
	ensureSessionCalls   int
	createCalls          int
}

func (f *fakeHonchoDocuments) EnsureWorkspace(id string, metadata map[string]any) (clients.HonchoWorkspace, error) {
	f.ensureWorkspaceCalls++
	return clients.HonchoWorkspace{ID: id, Metadata: metadata}, nil
}

func (f *fakeHonchoDocuments) EnsureSession(workspaceID string, sessionID string, metadata map[string]any) (clients.HonchoSession, error) {
	f.ensureSessionCalls++
	return clients.HonchoSession{ID: sessionID, WorkspaceID: workspaceID, Metadata: metadata}, nil
}

func (f *fakeHonchoDocuments) CreateConclusions(workspaceID string, conclusions []clients.HonchoConclusionCreate) ([]clients.HonchoConclusion, error) {
	f.createCalls++
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

func TestNotionDocumentConclusionContentCarriesProvenance(t *testing.T) {
	content := NotionDocumentConclusionContent(NotionDocumentInput{
		PageID:         "page_abc",
		Title:          "Deploy Runbook",
		URL:            "https://notion.so/page_abc",
		LastEditedTime: "2026-05-02T10:00:00.000Z",
		Content:        "Roll forward after validation.",
		Hierarchy:      []string{"Engineering", "Runbooks", "Deploy Runbook"},
	})
	for _, expected := range []string{
		"# Deploy Runbook",
		"URL: https://notion.so/page_abc",
		"Notion page id: page_abc",
		"Last edited: 2026-05-02T10:00:00.000Z",
		"Hierarchy: Engineering > Runbooks > Deploy Runbook",
		"Roll forward after validation.",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected %q in content:\n%s", expected, content)
		}
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
