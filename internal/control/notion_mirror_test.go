package control

import (
	"context"
	"strings"
	"testing"
	"time"

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

func (f fakeNotionAPI) RetrieveDataSource(ctx context.Context, dataSourceID string) (clients.NotionDataSource, error) {
	return clients.NotionDataSource{}, clients.NotionAPIError{StatusCode: 404, Body: "not found"}
}

func (f fakeNotionAPI) RetrievePageMarkdown(ctx context.Context, pageID string, includeTranscript bool) (clients.NotionPageMarkdown, error) {
	return clients.NotionPageMarkdown{
		Object:   "markdown",
		ID:       pageID,
		Markdown: "Roll forward after validation.",
	}, nil
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

func (f fakeNotionAPI) QueryDataSource(ctx context.Context, dataSourceID string, opts clients.NotionDataSourceQueryOptions) (clients.NotionListResponse[clients.NotionPage], error) {
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
	pages                     map[string]clients.NotionPage
	databases                 map[string]clients.NotionDatabase
	dataSources               map[string]clients.NotionDataSource
	markdown                  map[string]clients.NotionPageMarkdown
	markdownErrors            map[string]error
	children                  map[string][]clients.NotionBlock
	databaseRows              map[string][]clients.NotionPage
	dataSourceRows            map[string][]clients.NotionPage
	pageTypeMismatch          map[string]bool
	dbTypeMismatch            map[string]bool
	dbNoAccessibleDataSources map[string]bool
	dataSourceTypeMismatch    map[string]bool
	retrievePageCalls         map[string]int
	retrieveDataSourceCalls   map[string]int
	retrieveMarkdownCalls     map[string]int
	queryDataSourceCalls      map[string]int
	queryDataSourceOptions    map[string]clients.NotionDataSourceQueryOptions
	listBlockCalls            map[string]int
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
	if f.dbTypeMismatch[databaseID] {
		return clients.NotionDatabase{}, notionEndpointTypeMismatchError("page", "database")
	}
	if f.dbNoAccessibleDataSources[databaseID] {
		return clients.NotionDatabase{}, notionDatabaseNoAccessibleDataSourcesError(databaseID)
	}
	if database, ok := f.databases[databaseID]; ok {
		return database, nil
	}
	return clients.NotionDatabase{}, clients.NotionAPIError{StatusCode: 404, Body: "not found"}
}

func (f *fakeNotionGraphAPI) RetrieveDataSource(ctx context.Context, dataSourceID string) (clients.NotionDataSource, error) {
	dataSourceID = normalizeNotionID(dataSourceID)
	if f.retrieveDataSourceCalls == nil {
		f.retrieveDataSourceCalls = map[string]int{}
	}
	f.retrieveDataSourceCalls[dataSourceID]++
	if f.dataSourceTypeMismatch[dataSourceID] {
		return clients.NotionDataSource{}, notionEndpointTypeMismatchError("database", "data source")
	}
	if dataSource, ok := f.dataSources[dataSourceID]; ok {
		return dataSource, nil
	}
	return clients.NotionDataSource{}, clients.NotionAPIError{StatusCode: 404, Body: "not found"}
}

func (f *fakeNotionGraphAPI) RetrievePageMarkdown(ctx context.Context, pageID string, includeTranscript bool) (clients.NotionPageMarkdown, error) {
	pageID = normalizeNotionID(pageID)
	if f.retrieveMarkdownCalls == nil {
		f.retrieveMarkdownCalls = map[string]int{}
	}
	f.retrieveMarkdownCalls[pageID]++
	if err, ok := f.markdownErrors[pageID]; ok {
		return clients.NotionPageMarkdown{}, err
	}
	if markdown, ok := f.markdown[pageID]; ok {
		return markdown, nil
	}
	if blocks := f.children[pageID]; len(blocks) > 0 {
		var parts []string
		for _, block := range blocks {
			switch block.Type {
			case "child_page":
				parts = append(parts, `<page url="`+notionTestReferenceURL(block.ID)+`">Child</page>`)
			case "child_database":
				parts = append(parts, `<database url="`+notionTestReferenceURL(block.ID)+`">Child</database>`)
			default:
				if line := notionBlockMarkdown(block, 0); line != "" {
					parts = append(parts, line)
				}
			}
		}
		return clients.NotionPageMarkdown{Object: "markdown", ID: pageID, Markdown: strings.Join(parts, "\n")}, nil
	}
	return clients.NotionPageMarkdown{Object: "markdown", ID: pageID, Markdown: "Body for " + pageID}, nil
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

func (f *fakeNotionGraphAPI) QueryDataSource(ctx context.Context, dataSourceID string, opts clients.NotionDataSourceQueryOptions) (clients.NotionListResponse[clients.NotionPage], error) {
	dataSourceID = normalizeNotionID(dataSourceID)
	if f.queryDataSourceCalls == nil {
		f.queryDataSourceCalls = map[string]int{}
	}
	if f.queryDataSourceOptions == nil {
		f.queryDataSourceOptions = map[string]clients.NotionDataSourceQueryOptions{}
	}
	f.queryDataSourceCalls[dataSourceID]++
	f.queryDataSourceOptions[dataSourceID] = opts
	rows := f.dataSourceRows[dataSourceID]
	if len(rows) == 0 {
		rows = f.databaseRows[dataSourceID]
	}
	return clients.NotionListResponse[clients.NotionPage]{Results: rows}, nil
}

func notionTestReferenceURL(id string) string {
	normalized := normalizeNotionID(id)
	if normalized == "" {
		normalized = strings.TrimSpace(id)
	}
	if compactNotionID(normalized) != "" {
		return "https://notion.so/" + normalized
	}
	return normalized
}

func notionEndpointTypeMismatchError(actual string, requested string) clients.NotionAPIError {
	return clients.NotionAPIError{
		StatusCode: 400,
		Body:       `{"object":"error","status":400,"code":"validation_error","message":"Provided ID abc is a ` + actual + `, not a ` + requested + `. Use the retrieve ` + actual + ` API instead."}`,
	}
}

func notionDatabaseNoAccessibleDataSourcesError(databaseID string) clients.NotionAPIError {
	return clients.NotionAPIError{
		StatusCode: 400,
		Body:       `{"object":"error","status":400,"code":"validation_error","message":"Database with ID ` + databaseID + ` does not contain any data sources accessible by this API bot."}`,
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
		markdown: map[string]clients.NotionPageMarkdown{
			rootID:  {Object: "markdown", ID: rootID, Markdown: `<page url="` + notionTestReferenceURL(childID) + `">Child</page>`},
			childID: {Object: "markdown", ID: childID, Markdown: "Child body"},
		},
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
	if api.retrieveMarkdownCalls[rootID] == 0 || api.retrieveMarkdownCalls[childID] == 0 {
		t.Fatalf("expected markdown fetches for root and child, got calls=%v", api.retrieveMarkdownCalls)
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
		CompanyWikiRoot:            t.TempDir(),
	}
	databaseID := "databaseabc"
	dataSourceID := "datasourceabc"
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
				DataSources:    []clients.NotionDataSourceReference{{ID: dataSourceID, Name: "Backlog data"}},
				Properties: map[string]any{
					"Priority": map[string]any{"type": "select"},
				},
			},
		},
		dataSources: map[string]clients.NotionDataSource{
			dataSourceID: {
				Object:         "data_source",
				ID:             dataSourceID,
				Name:           "Backlog data",
				URL:            "https://notion.so/datasourceabc",
				LastEditedTime: "2026-05-02T10:00:00.000Z",
				Parent:         map[string]any{"database_id": databaseID},
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
		markdown:         map[string]clients.NotionPageMarkdown{rowID: {Object: "markdown", ID: rowID, Markdown: "Row body"}},
		dataSourceRows:   map[string][]clients.NotionPage{dataSourceID: {notionTestPage(rowID, "Row Page", false)}},
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
	if !strings.Contains(databaseRecord.SourceRevision, "schema_hash:") {
		t.Fatalf("database source revision did not include schema hash: %q", databaseRecord.SourceRevision)
	}
	dataSourceKey := companyknowledge.NotionObjectSourceKey("notion", companyknowledge.NotionObjectKindDataSource, dataSourceID)
	dataSourceRecord, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, dataSourceKey)
	if err != nil || !found || dataSourceRecord.Status != store.SourceMirrorStatusComplete {
		t.Fatalf("data source record found=%t err=%v record=%+v", found, err, dataSourceRecord)
	}
	rowRecord, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, companyknowledge.NotionDocumentSourceKey("notion", rowID))
	if err != nil || !found || rowRecord.Status != store.SourceMirrorStatusComplete {
		t.Fatalf("row source record found=%t err=%v record=%+v", found, err, rowRecord)
	}
	pages, err := state.SearchCompanyWikiPages("Product Backlog", 10)
	if err != nil {
		t.Fatalf("SearchCompanyWikiPages() error = %v", err)
	}
	foundDatabaseWikiPage := false
	for _, page := range pages {
		if page.Title == "Numo Product Backlog" {
			foundDatabaseWikiPage = true
		}
	}
	if !foundDatabaseWikiPage {
		t.Fatalf("database root was not published into company wiki pages: %+v", pages)
	}
}

func TestNotionMirrorDatabaseFallbackDataSourceTypeMismatchDoesNotAbort(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	databaseID := "databaseabc"
	api := &fakeNotionGraphAPI{
		databases: map[string]clients.NotionDatabase{
			databaseID: {
				Object:         "database",
				ID:             databaseID,
				URL:            "https://notion.so/databaseabc",
				Title:          []clients.NotionText{{PlainText: "Legacy Database"}},
				LastEditedTime: "2026-05-02T10:00:00.000Z",
				DataSources:    []clients.NotionDataSourceReference{{ID: databaseID, Name: "Legacy fallback"}},
			},
		},
		pageTypeMismatch:       map[string]bool{databaseID: true},
		dataSourceTypeMismatch: map[string]bool{databaseID: true},
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
	if err != nil || !found || databaseRecord.Status != store.SourceMirrorStatusComplete {
		t.Fatalf("database source record found=%t err=%v record=%+v", found, err, databaseRecord)
	}
}

func TestNotionMirrorDatabaseWithoutDataSourcesDoesNotRecordFalseDataSourceMiss(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	databaseID := "databaseabc"
	api := &fakeNotionGraphAPI{
		databases: map[string]clients.NotionDatabase{
			databaseID: {
				Object:         "database",
				ID:             databaseID,
				URL:            "https://notion.so/databaseabc",
				Title:          []clients.NotionText{{PlainText: "Empty Database"}},
				LastEditedTime: "2026-05-02T10:00:00.000Z",
			},
		},
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
	if got := api.retrieveDataSourceCalls[databaseID]; got != 0 {
		t.Fatalf("RetrieveDataSource(%q) calls = %d, want 0", databaseID, got)
	}
	missKey := companyknowledge.NotionCrawlMissSourceKey("notion", databaseID, databaseID)
	if missRecord, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionCrawlMissSourceType, missKey); err != nil || found {
		t.Fatalf("false data source miss found=%t err=%v record=%+v", found, err, missRecord)
	}
}

func TestNotionMirrorDataSourceMissMarksSeen(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	dataSourceID := "missingdatasource"
	api := &fakeNotionGraphAPI{}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorDataSource(context.Background(), dataSourceID, "", rootID, nil); err != nil {
		t.Fatalf("first mirrorDataSource() error = %v", err)
	}
	if err := runner.mirrorDataSource(context.Background(), dataSourceID, "", rootID, nil); err != nil {
		t.Fatalf("second mirrorDataSource() error = %v", err)
	}
	if got := api.retrieveDataSourceCalls[dataSourceID]; got != 1 {
		t.Fatalf("RetrieveDataSource(%q) calls = %d, want 1", dataSourceID, got)
	}
	missKey := companyknowledge.NotionCrawlMissSourceKey("notion", rootID, dataSourceID)
	if missRecord, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionCrawlMissSourceType, missKey); err != nil || !found || missRecord.Status != store.SourceMirrorStatusStale {
		t.Fatalf("crawl miss found=%t err=%v record=%+v", found, err, missRecord)
	}
}

func TestNotionMirrorChildPageTypeMismatchFallsBackToDatabase(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
		CompanyWikiRoot:            t.TempDir(),
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
				Title:          []clients.NotionText{{PlainText: "Engineering Home"}},
				LastEditedTime: "2026-05-02T10:00:00.000Z",
			},
		},
		children: map[string][]clients.NotionBlock{
			rowID: {
				{ID: databaseID, Type: "child_page", Raw: map[string]any{"child_page": map[string]any{"title": "Engineering Home"}}},
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
	if _, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, databaseKey); err != nil || !found {
		t.Fatalf("database source record found=%t err=%v", found, err)
	}
}

func TestNotionMirrorChildPageTypeMismatchFallsBackToDataSourceAfterDatabase404(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	dataSourceID := "datasourceabc"
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rootID: notionTestPage(rootID, "Root", false),
		},
		dataSources: map[string]clients.NotionDataSource{
			dataSourceID: {
				Object:         "data_source",
				ID:             dataSourceID,
				Name:           "Backlog data",
				URL:            "https://notion.so/datasourceabc",
				LastEditedTime: "2026-05-02T10:00:00.000Z",
			},
		},
		markdown: map[string]clients.NotionPageMarkdown{
			rootID: {Object: "markdown", ID: rootID, Markdown: `<page url="` + notionTestReferenceURL(dataSourceID) + `">Linked data source</page>`},
		},
		pageTypeMismatch: map[string]bool{dataSourceID: true},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), rootID); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	dataSourceKey := companyknowledge.NotionObjectSourceKey("notion", companyknowledge.NotionObjectKindDataSource, dataSourceID)
	dataSourceRecord, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, dataSourceKey)
	if err != nil || !found || dataSourceRecord.Status != store.SourceMirrorStatusComplete {
		t.Fatalf("data source record found=%t err=%v record=%+v", found, err, dataSourceRecord)
	}
}

func TestNotionMirrorChildDatabaseTypeMismatchFallsBackToPage(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	childID := "childpage"
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rootID:  notionTestPage(rootID, "Root", false),
			childID: notionTestPage(childID, "Child", false),
		},
		databases: map[string]clients.NotionDatabase{},
		children: map[string][]clients.NotionBlock{
			rootID: {
				{ID: childID, Type: "child_database", Raw: map[string]any{"child_database": map[string]any{"title": "Child"}}},
			},
			childID: {
				{ID: "childtext", Type: "paragraph", Raw: map[string]any{"paragraph": map[string]any{"rich_text": []any{map[string]any{"plain_text": "Child body"}}}}},
			},
		},
		dbTypeMismatch: map[string]bool{childID: true},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), rootID); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	childRecord, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, companyknowledge.NotionDocumentSourceKey("notion", childID))
	if err != nil || !found || childRecord.Status != store.SourceMirrorStatusComplete {
		t.Fatalf("child source record found=%t err=%v record=%+v", found, err, childRecord)
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
	if record.Metadata["archived"] != true {
		t.Fatalf("record archived metadata = %v, want true", record.Metadata["archived"])
	}
	if record.Metadata["in_trash"] != false {
		t.Fatalf("record in_trash metadata = %v, want false", record.Metadata["in_trash"])
	}
	if record.Metadata["stale_reason"] != "notion page is archived" {
		t.Fatalf("record stale_reason = %v", record.Metadata["stale_reason"])
	}
}

func TestNotionMirrorArchivedDataSourceMarksSourceRecordStale(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	databaseID := "databaseabc"
	dataSourceID := "datasourceabc"
	api := &fakeNotionGraphAPI{
		dataSources: map[string]clients.NotionDataSource{
			dataSourceID: {
				Object:         "data_source",
				ID:             dataSourceID,
				Name:           "Archived data source",
				LastEditedTime: "2026-05-02T10:00:00.000Z",
				Archived:       true,
				InTrash:        true,
			},
		},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorDataSource(context.Background(), dataSourceID, databaseID, rootID, nil); err != nil {
		t.Fatalf("mirrorDataSource() error = %v", err)
	}
	record, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, companyknowledge.NotionObjectSourceKey("notion", companyknowledge.NotionObjectKindDataSource, dataSourceID))
	if err != nil || !found {
		t.Fatalf("stale data source record found=%t err=%v", found, err)
	}
	if record.Status != store.SourceMirrorStatusStale {
		t.Fatalf("record status = %s, want stale", record.Status)
	}
	if record.Metadata["archived"] != true || record.Metadata["in_trash"] != true {
		t.Fatalf("record stale metadata = %+v, want archived and in_trash true", record.Metadata)
	}
	if record.Metadata["database_id"] != databaseID {
		t.Fatalf("record database_id = %v, want %s", record.Metadata["database_id"], databaseID)
	}
	if record.Metadata["stale_reason"] != "notion data source is archived and in trash" {
		t.Fatalf("record stale_reason = %v", record.Metadata["stale_reason"])
	}
}

func TestNotionMirrorPageTypeMismatchFallbackExhaustionRecordsCrawlMiss(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	objectID := "unknownobject"
	api := &fakeNotionGraphAPI{
		pageTypeMismatch:       map[string]bool{objectID: true},
		dataSourceTypeMismatch: map[string]bool{objectID: true},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, objectID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	if err := runner.mirrorPage(context.Background(), objectID, objectID, nil); err != nil {
		t.Fatalf("mirrorPage() should record a crawl miss instead of returning type mismatch, got error = %v", err)
	}
	crawlMissKey := companyknowledge.NotionCrawlMissSourceKey("notion", objectID, objectID)
	record, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionCrawlMissSourceType, crawlMissKey)
	if err != nil || !found {
		t.Fatalf("crawl miss record found=%t err=%v", found, err)
	}
	if record.Status != store.SourceMirrorStatusStale {
		t.Fatalf("crawl miss status = %s, want stale", record.Status)
	}
	if record.Metadata["target_object_kind"] != "page" {
		t.Fatalf("crawl miss target_object_kind = %v", record.Metadata["target_object_kind"])
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
	expectedMsg := "notion allowlist root nonexistent is neither a visible page, database, nor data source"
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
		markdown: map[string]clients.NotionPageMarkdown{
			rootID: {Object: "markdown", ID: rootID, Markdown: `<database url="` + notionTestReferenceURL(childDBID) + `">Missing DB</database>`},
		},
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

func TestNotionMirrorChildDatabaseNoAccessibleDataSourcesRecordsCrawlMiss(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	childDBID := "inaccessibledb"
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rootID: notionTestPage(rootID, "Root", false),
		},
		markdown: map[string]clients.NotionPageMarkdown{
			rootID: {Object: "markdown", ID: rootID, Markdown: `<database url="` + notionTestReferenceURL(childDBID) + `">Inaccessible DB</database>`},
		},
		dbNoAccessibleDataSources: map[string]bool{childDBID: true},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("newNotionMirrorRunner() error = %v", err)
	}
	err = runner.mirrorRoot(context.Background(), rootID)
	if err != nil {
		t.Fatalf("mirrorRoot() should succeed despite child database with inaccessible data sources, got error = %v", err)
	}
	crawlMissKey := companyknowledge.NotionCrawlMissSourceKey("notion", rootID, childDBID)
	record, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionCrawlMissSourceType, crawlMissKey)
	if err != nil || !found {
		t.Fatalf("crawl miss record found=%t err=%v", found, err)
	}
	if record.Status != store.SourceMirrorStatusStale {
		t.Fatalf("crawl miss record status = %s, want stale", record.Status)
	}
	if record.Metadata["reason"] != "notion database has no data sources accessible by this API bot" {
		t.Fatalf("crawl miss reason = %v", record.Metadata["reason"])
	}
}

func TestNotionMirrorUnchangedPageSkipsMarkdownFetchUsingStoreAuthority(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rootID: notionTestPage(rootID, "Root", false),
		},
		markdown: map[string]clients.NotionPageMarkdown{
			rootID: {Object: "markdown", ID: rootID, Markdown: "Root body"},
		},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	first, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("new first runner: %v", err)
	}
	if err := first.mirrorRoot(context.Background(), rootID); err != nil {
		t.Fatalf("first mirrorRoot() error = %v", err)
	}
	api.retrieveMarkdownCalls = map[string]int{}
	second, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("new second runner: %v", err)
	}
	if err := second.mirrorRoot(context.Background(), rootID); err != nil {
		t.Fatalf("second mirrorRoot() error = %v", err)
	}
	if got := api.retrieveMarkdownCalls[rootID]; got != 0 {
		t.Fatalf("unchanged page markdown fetches = %d, want 0", got)
	}
	if honcho.createCalls != 1 {
		t.Fatalf("CreateConclusions calls = %d, want first write only", honcho.createCalls)
	}
}

func TestNotionMirrorDataSourceDeltaUsesHighWatermarkLookback(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                  "stage",
		SourceMirrorCheckpointRoot:   t.TempDir(),
		HonchoWorkspaceID:            "rsi_company_knowledge",
		NotionMirrorDeltaEnabled:     true,
		NotionMirrorDeltaLookback:    10 * time.Minute,
		NotionMirrorFullScanInterval: 24 * time.Hour,
	}
	dataSourceID := "datasourceabc"
	rowID := "rowpageabc"
	checkpoint := notionMirrorCheckpoint{
		RootID:           dataSourceID,
		WorkspaceID:      "notion",
		CompletedPages:   map[string]string{},
		CompletedObjects: map[string]notionMirrorCheckpointObject{},
		DataSources: map[string]notionMirrorDataSourceState{
			dataSourceID: {
				DataSourceID:            dataSourceID,
				RowHighWatermark:        "2026-05-02T10:00:00Z",
				LastFullScanCompletedAt: time.Now().UTC(),
			},
		},
	}
	if err := writeNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, checkpoint); err != nil {
		t.Fatalf("write checkpoint: %v", err)
	}
	api := &fakeNotionGraphAPI{
		dataSources: map[string]clients.NotionDataSource{
			dataSourceID: {Object: "data_source", ID: dataSourceID, Name: "Engineering", LastEditedTime: "2026-05-02T10:00:00Z"},
		},
		dataSourceRows: map[string][]clients.NotionPage{
			dataSourceID: {notionTestPage(rowID, "Row", false)},
		},
		markdown: map[string]clients.NotionPageMarkdown{
			rowID: {Object: "markdown", ID: rowID, Markdown: "Row body"},
		},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, dataSourceID)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), dataSourceID); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	opts := api.queryDataSourceOptions[dataSourceID]
	if opts.LastEditedTimeOnOrAfter != "2026-05-02T09:50:00Z" {
		t.Fatalf("delta filter = %q, want lookback-adjusted high watermark", opts.LastEditedTimeOnOrAfter)
	}
	if opts.SortTimestamp != "last_edited_time" || opts.SortDirection != "ascending" {
		t.Fatalf("sort opts = %#v", opts)
	}
	if len(opts.FilterProperties) != 1 || opts.FilterProperties[0] != "title" {
		t.Fatalf("filter properties = %#v", opts.FilterProperties)
	}
}

func TestNotionMirrorMarkdownUnknownBlocksDoNotFailRoot(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rootID: notionTestPage(rootID, "Root", false),
		},
		markdown: map[string]clients.NotionPageMarkdown{
			rootID: {Object: "markdown", ID: rootID, Markdown: "Root body", Truncated: true, UnknownBlockIDs: []string{"missingblock"}},
		},
		markdownErrors: map[string]error{
			"missingblock": clients.NotionAPIError{StatusCode: 404, Body: "not found"},
		},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), rootID); err != nil {
		t.Fatalf("mirrorRoot() should not fail on inaccessible unknown block: %v", err)
	}
	record, found, err := state.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, companyknowledge.NotionDocumentSourceKey("notion", rootID))
	if err != nil || !found || record.Status != store.SourceMirrorStatusComplete {
		t.Fatalf("source record found=%t err=%v record=%+v", found, err, record)
	}
}

func TestNotionMirrorDirtyObjectCheckpointAcceleratesPageCrawl(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeControlHonchoDocuments{}
	cfg := config.Config{
		Environment:                "stage",
		SourceMirrorCheckpointRoot: t.TempDir(),
		HonchoWorkspaceID:          "rsi_company_knowledge",
	}
	rootID := "rootpage"
	childID := "childpage"
	if _, status, err := recordNotionMirrorDirtyObject(cfg, notionMirrorDirtyObjectRequest{
		RootID:     rootID,
		ObjectID:   childID,
		ObjectKind: companyknowledge.NotionObjectKindPage,
		EventType:  "page.content_updated",
	}); err != nil || status != 202 {
		t.Fatalf("record dirty status=%d err=%v", status, err)
	}
	api := &fakeNotionGraphAPI{
		pages: map[string]clients.NotionPage{
			rootID:  notionTestPage(rootID, "Root", false),
			childID: notionTestPage(childID, "Child", false),
		},
		markdown: map[string]clients.NotionPageMarkdown{
			rootID:  {Object: "markdown", ID: rootID, Markdown: "Root body"},
			childID: {Object: "markdown", ID: childID, Markdown: "Child body"},
		},
	}
	mirror := companyknowledge.NewNotionMirror(state, honcho, companyknowledge.NotionMirrorOptions{Environment: "stage", HonchoWorkspace: "rsi_company_knowledge"})
	runner, err := newNotionMirrorRunner(cfg, api, state, mirror, rootID)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	if err := runner.mirrorRoot(context.Background(), rootID); err != nil {
		t.Fatalf("mirrorRoot() error = %v", err)
	}
	if api.retrievePageCalls[childID] == 0 {
		t.Fatalf("dirty child page was not crawled before normal root traversal")
	}
	checkpoint, err := readNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, rootID)
	if err != nil {
		t.Fatalf("read checkpoint: %v", err)
	}
	if len(checkpoint.DirtyObjects) != 0 {
		t.Fatalf("dirty queue not drained: %+v", checkpoint.DirtyObjects)
	}
}

func TestNormalizeNotionIDAllowsRawIDsButStrictlyExtractsURLs(t *testing.T) {
	notionID := "121051299a5480dd989ddd05c1b3a694"
	hyphenated := "12105129-9a54-80dd-989d-dd05c1b3a694"
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "raw fake id stays usable for tests",
			input: "rootpage",
			want:  "rootpage",
		},
		{
			name:  "hyphenated notion id compacts",
			input: hyphenated,
			want:  notionID,
		},
		{
			name:  "slug ending in hex does not shift id",
			input: "https://www.notion.so/storyprotocol/Code-" + notionID,
			want:  notionID,
		},
		{
			name:  "url without notion id returns empty",
			input: "https://www.notion.so/storyprotocol/workspace-without-id",
			want:  "",
		},
		{
			name:  "query notion id is accepted",
			input: "https://www.notion.so/storyprotocol/wiki?p=" + hyphenated,
			want:  notionID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeNotionID(tt.input); got != tt.want {
				t.Fatalf("normalizeNotionID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
