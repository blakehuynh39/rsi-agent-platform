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
	createCalls  int
	messageCalls int
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

func (f *fakeControlHonchoDocuments) CreateMessages(workspaceID string, sessionID string, messages []clients.HonchoMessageCreate) ([]clients.HonchoMessage, error) {
	f.messageCalls++
	out := make([]clients.HonchoMessage, 0, len(messages))
	for i, msg := range messages {
		out = append(out, clients.HonchoMessage{ID: "msg_1", Content: msg.Content})
		if i > 0 {
			out[i].ID = "msg_2"
		}
	}
	return out, nil
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
	runner, err := newNotionMirrorRunner(cfg, fakeNotionAPI{}, state, mirror, "page-abc")
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

type fakeNotionGraphAPI struct {
	pages             map[string]clients.NotionPage
	databases         map[string]clients.NotionDatabase
	children          map[string][]clients.NotionBlock
	databaseRows      map[string][]clients.NotionPage
	pageTypeMismatch  map[string]bool
	retrievePageCalls map[string]int
	listBlockCalls    map[string]int
}

func (f *fakeNotionGraphAPI) RetrievePage(ctx context.Context, pageID string) (clients.NotionPage, error) {
	pageID = normalizeNotionID(pageID)
	if f.retrievePageCalls == nil {
		f.retrievePageCalls = map[string]int{}
	}
	f.retrievePageCalls[pageID]++
	if f.pageTypeMismatch[pageID] {
		return clients.NotionPage{}, notionEndpointTypeMismatchError("database", "page")
	}
	if page, ok := f.pages[pageID]; ok {
		return page, nil
	}
	return clients.NotionPage{}, clients.NotionAPIError{StatusCode: 404, Body: "not found"}
}

func (f *fakeNotionGraphAPI) RetrieveDatabase(ctx context.Context, databaseID string) (clients.NotionDatabase, error) {
	databaseID = normalizeNotionID(databaseID)
	if database, ok := f.databases[databaseID]; ok {
		return database, nil
	}
	return clients.NotionDatabase{}, clients.NotionAPIError{StatusCode: 404, Body: "not found"}
}

func (f *fakeNotionGraphAPI) ListBlockChildren(ctx context.Context, blockID string, cursor string, pageSize int) (clients.NotionListResponse[clients.NotionBlock], error) {
	blockID = normalizeNotionID(blockID)
	if f.listBlockCalls == nil {
		f.listBlockCalls = map[string]int{}
	}
	f.listBlockCalls[blockID]++
	return clients.NotionListResponse[clients.NotionBlock]{Results: f.children[blockID]}, nil
}

func (f *fakeNotionGraphAPI) QueryDatabase(ctx context.Context, databaseID string, cursor string, pageSize int) (clients.NotionListResponse[clients.NotionPage], error) {
	databaseID = normalizeNotionID(databaseID)
	return clients.NotionListResponse[clients.NotionPage]{Results: f.databaseRows[databaseID]}, nil
}

func notionEndpointTypeMismatchError(actual string, requested string) clients.NotionAPIError {
	return clients.NotionAPIError{
		StatusCode: 400,
		Body:       `{"object":"error","status":400,"code":"validation_error","message":"Provided ID abc is a ` + actual + `, not a ` + requested + `. Use the retrieve ` + actual + ` API instead."}`,
	}
}

func notionTestPage(id string, title string, archived bool) clients.NotionPage {
	return clients.NotionPage{
		Object:         "page",
		ID:             normalizeNotionID(id),
		URL:            "https://notion.so/" + normalizeNotionID(id),
		LastEditedTime: "2026-05-02T10:00:00.000Z",
		CreatedTime:    "2026-05-01T10:00:00.000Z",
		Archived:       archived,
		Properties: map[string]any{
			"title": map[string]any{
				"type":  "title",
				"title": []any{map[string]any{"plain_text": title}},
			},
		},
	}
}

func TestNotionMirrorCheckpointDoesNotBypassStoreAuthorityOrChildDiscovery(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	childID := "childpage"
	rootInput := companyknowledge.NotionDocumentInput{
		WorkspaceID:    "notion",
		PageID:         rootID,
		Title:          "Root",
		URL:            "https://notion.so/rootpage",
		LastEditedTime: "2026-05-02T10:00:00.000Z",
	}
	checkpoint := notionMirrorCheckpoint{
		RootID:           rootID,
		WorkspaceID:      "notion",
		CompletedPages:   map[string]string{companyknowledge.NotionDocumentSourceKey("notion", rootID): companyknowledge.NotionDocumentSourceRevision(rootInput)},
		CompletedObjects: map[string]notionMirrorCheckpointObject{},
	}
	if err := writeNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, checkpoint); err != nil {
		t.Fatalf("write checkpoint error = %v", err)
	}
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rootID:  notionTestPage(rootID, "Root", false),
			childID: notionTestPage(childID, "Child", false),
		},
		databases: map[string]clients.NotionDatabase{},
		children: map[string][]clients.NotionBlock{
			rootID: {
				{ID: childID, Type: "child_page", Raw: map[string]any{"child_page": map[string]any{"title": "Child"}}},
			},
			childID: {
				{ID: "childtext", Type: "paragraph", Raw: map[string]any{"paragraph": map[string]any{"rich_text": []any{map[string]any{"plain_text": "Child body"}}}}},
			},
		},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), rootID); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	if honcho.createCalls != 2 {
		t.Fatalf("CreateConclusions calls = %d, want root and child despite checkpoint-only match", honcho.createCalls)
	}
	if api.listBlockCalls[rootID] == 0 || api.listBlockCalls[childID] == 0 {
		t.Fatalf("expected child discovery scans, got calls=%v", api.listBlockCalls)
	}
}

func TestNotionMirrorPageRootUsesInitialRetrieve(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	api := &fakeNotionGraphAPI{
		pages:     map[string]clients.NotionPage{rootID: notionTestPage(rootID, "Root", false)},
		databases: map[string]clients.NotionDatabase{},
		children:  map[string][]clients.NotionBlock{rootID: {}},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), rootID); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	if got := api.retrievePageCalls[rootID]; got != 1 {
		t.Fatalf("RetrievePage(%q) calls = %d, want 1", rootID, got)
	}
}

func TestNotionMirrorDatabaseRootWritesDatabaseObjectAndRows(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	databaseID := "databaseabc"
	rowID := "rowpageabc"
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rowID: notionTestPage(rowID, "Row Page", false),
		},
		databases: map[string]clients.NotionDatabase{
			databaseID: {
				Object:         "database",
				ID:             databaseID,
				URL:            "https://notion.so/databaseabc",
				Title:          []clients.NotionText{{PlainText: "Numo Product Backlog"}},
				LastEditedTime: "2026-05-02T10:00:00.000Z",
				Properties: map[string]any{
					"Status": map[string]any{
						"type": "status",
						"status": map[string]any{
							"options": []any{map[string]any{"name": "Todo"}, map[string]any{"name": "Done"}},
						},
					},
				},
			},
		},
		children: map[string][]clients.NotionBlock{
			rowID: {
				{ID: "rowtext", Type: "paragraph", Raw: map[string]any{"paragraph": map[string]any{"rich_text": []any{map[string]any{"plain_text": "Row body"}}}}},
			},
		},
		databaseRows:     map[string][]clients.NotionPage{databaseID: {notionTestPage(rowID, "Row Page", false)}},
		pageTypeMismatch: map[string]bool{databaseID: true},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, databaseID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), databaseID); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	databaseKey := companyknowledge.NotionObjectSourceKey("notion", companyknowledge.NotionObjectKindDatabase, databaseID)
	databaseRecord, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, databaseKey)
	if err != nil || !found {
		t.Fatalf("database source record found=%t err=%v", found, err)
	}
	if databaseRecord.SourceSessionKey != "notion:notion:database:"+databaseID {
		t.Fatalf("database session key = %q", databaseRecord.SourceSessionKey)
	}
	rowRecord, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, companyknowledge.NotionDocumentSourceKey("notion", rowID))
	if err != nil || !found || rowRecord.Status != store.SourceMirrorStatusComplete {
		t.Fatalf("row source record found=%t err=%v record=%+v", found, err, rowRecord)
	}
}

func TestNotionMirrorArchivedPageMarksSourceRecordStale(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	pageID := "archivedpage"
	api := &fakeNotionGraphAPI{
		pages:     map[string]clients.NotionPage{pageID: notionTestPage(pageID, "Archived", true)},
		databases: map[string]clients.NotionDatabase{},
		children:  map[string][]clients.NotionBlock{},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, pageID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), pageID); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	record, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, companyknowledge.NotionDocumentSourceKey("notion", pageID))
	if err != nil || !found {
		t.Fatalf("stale record found=%t err=%v", found, err)
	}
	if record.Status != store.SourceMirrorStatusStale {
		t.Fatalf("record status = %s, want stale", record.Status)
	}
}

func TestNotionMirrorRootNotFound404ReturnsError(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "nonexistent"
	api := &fakeNotionGraphAPI{
		pages:     map[string]clients.NotionPage{},
		databases: map[string]clients.NotionDatabase{},
		children:  map[string][]clients.NotionBlock{},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	err = runner.mirrorRoot(context.Background(), rootID)
	if err == nil {
		t.Fatalf("mirrorRoot() should return error for non-existent root, got nil")
	}
	expectedMsg := "notion allowlist root nonexistent is neither a visible page nor a visible database"
	if err.Error() != expectedMsg {
		t.Fatalf("mirrorRoot() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestNotionMirrorChildDatabase404RecordsCrawlMiss(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	childDBID := "missingdb"
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rootID: notionTestPage(rootID, "Root", false),
		},
		databases: map[string]clients.NotionDatabase{},
		children: map[string][]clients.NotionBlock{
			rootID: {
				{ID: childDBID, Type: "child_database", Raw: map[string]any{"child_database": map[string]any{"title": "Missing DB"}}},
			},
		},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	err = runner.mirrorRoot(context.Background(), rootID)
	if err != nil {
		t.Fatalf("mirrorRoot() should succeed despite child 404, got error = %v", err)
	}
	crawlMissKey := companyknowledge.NotionCrawlMissSourceKey("notion", rootID, childDBID)
	record, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionCrawlMissSourceType, crawlMissKey)
	if err != nil || !found {
		t.Fatalf("crawl miss record found=%t err=%v", found, err)
	}
	if record.Status != store.SourceMirrorStatusStale {
		t.Fatalf("crawl miss record status = %s, want stale", record.Status)
	}
}
