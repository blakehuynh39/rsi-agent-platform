package control

import (
	"context"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type fakeNotionAPI struct{}

func (f fakeNotionAPI) RetrievePage(ctx context.Context, pageID string) (clients.NotionPage, error) {
	return clients.NotionPage{
		Object:         "page",
		ID:             pageID,
		URL:            "https://notion.so/" + pageID,
		LastEditedTime: "2026-05-02T10:00:00.000Z",
		CreatedTime:    "2026-05-01T10:00:00.000Z",
		Properties: map[string]any{
			"title": map[string]any{
				"type": "title",
				"title": []any{
					map[string]any{"plain_text": "Deploy Runbook"},
				},
			},
		},
	}, nil
}

func (f fakeNotionAPI) RetrieveDatabase(ctx context.Context, databaseID string) (clients.NotionDatabase, error) {
	return clients.NotionDatabase{}, clients.NotionAPIError{StatusCode: 404, Body: "not found"}
}

func (f fakeNotionAPI) ListBlockChildren(ctx context.Context, blockID string, cursor string, pageSize int) (clients.NotionListResponse[clients.NotionBlock], error) {
	return clients.NotionListResponse[clients.NotionBlock]{
		Results: []clients.NotionBlock{
			{
				ID:   "block-1",
				Type: "paragraph",
				Raw: map[string]any{
					"paragraph": map[string]any{
						"rich_text": []any{map[string]any{"plain_text": "Roll forward after validation."}},
					},
				},
			},
		},
	}, nil
}

func (f fakeNotionAPI) QueryDatabase(ctx context.Context, databaseID string, cursor string, pageSize int) (clients.NotionListResponse[clients.NotionPage], error) {
	return clients.NotionListResponse[clients.NotionPage]{}, nil
}

type fakeControlHonchoDocuments struct {
	createCalls int
}

func (f *fakeControlHonchoDocuments) EnsureWorkspace(id string, metadata map[string]any) (clients.HonchoWorkspace, error) {
	return clients.HonchoWorkspace{ID: id, Metadata: metadata}, nil
}

func (f *fakeControlHonchoDocuments) EnsureSession(workspaceID string, sessionID string, metadata map[string]any) (clients.HonchoSession, error) {
	return clients.HonchoSession{ID: sessionID, WorkspaceID: workspaceID, Metadata: metadata}, nil
}

func (f *fakeControlHonchoDocuments) CreateConclusions(workspaceID string, conclusions []clients.HonchoConclusionCreate) ([]clients.HonchoConclusion, error) {
	f.createCalls++
	return []clients.HonchoConclusion{{ID: "doc_1", Content: conclusions[0].Content}}, nil
}

func TestNotionMirrorRunnerMirrorsPageAndWritesCheckpoint(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{
		Environment:     "stage",
		HonchoWorkspace: "rsi_company_knowledge",
	})
	runner, err := newNotionMirrorRunner(cfg, fakeNotionAPI{}, mirror, "page-abc")
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), "page-abc"); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	if honcho.createCalls != 1 {
		t.Fatalf("CreateConclusions calls = %d, want 1", honcho.createCalls)
	}
	record, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, companyknowledge.NotionDocumentSourceKey("notion", "pageabc"))
	if err != nil {
		t.Fatalf("GetSourceMirrorRecord() error = %v", err)
	}
	if !found || record.HonchoObjectType != "document" || record.HonchoObjectID != "doc_1" {
		t.Fatalf("unexpected source mirror record found=%t record=%+v", found, record)
	}
	checkpoint, err := readNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, "pageabc")
	if err != nil {
		t.Fatalf("readNotionMirrorCheckpoint() error = %v", err)
	}
	if checkpoint.CompletedPages[companyknowledge.NotionDocumentSourceKey("notion", "pageabc")] == "" {
		t.Fatalf("checkpoint did not record mirrored page: %+v", checkpoint)
	}
}
