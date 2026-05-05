package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type notionMirrorCheckpoint struct {
	RootID            string                                  `json:"root_id"`
	WorkspaceID       string                                  `json:"workspace_id"`
	CompletedPages    map[string]string                       `json:"completed_pages,omitempty"`
	CompletedObjects  map[string]notionMirrorCheckpointObject `json:"completed_objects,omitempty"`
	DataSources       map[string]notionMirrorDataSourceState  `json:"data_sources,omitempty"`
	DirtyObjects      map[string]notionMirrorDirtyObject      `json:"dirty_objects,omitempty"`
	LastPageID        string                                  `json:"last_page_id,omitempty"`
	LastProgressAt    time.Time                               `json:"last_progress_at,omitempty"`
	LastCompletedAt   time.Time                               `json:"last_completed_at,omitempty"`
	LastPageCount     int                                     `json:"last_page_count"`
	LastDatabaseCount int                                     `json:"last_database_count"`
	TraversalStatus   string                                  `json:"traversal_status,omitempty"`
}

type notionMirrorCheckpointObject struct {
	SourceRevision          string    `json:"source_revision"`
	ChildPageIDs            []string  `json:"child_page_ids,omitempty"`
	ChildDatabaseIDs        []string  `json:"child_database_ids,omitempty"`
	ChildDataSourceIDs      []string  `json:"child_data_source_ids,omitempty"`
	BlockPaginationComplete bool      `json:"block_pagination_complete"`
	Truncated               bool      `json:"truncated"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type notionMirrorDataSourceState struct {
	DataSourceID            string    `json:"data_source_id"`
	DatabaseID              string    `json:"database_id,omitempty"`
	RowHighWatermark        string    `json:"row_high_watermark,omitempty"`
	LastDeltaCompletedAt    time.Time `json:"last_delta_completed_at,omitempty"`
	LastFullScanCompletedAt time.Time `json:"last_full_scan_completed_at,omitempty"`
	ObservedPageIDs         []string  `json:"observed_page_ids,omitempty"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type notionMirrorDirtyObject struct {
	ObjectKind     string    `json:"object_kind"`
	ObjectID       string    `json:"object_id"`
	EventType      string    `json:"event_type,omitempty"`
	EventTimestamp string    `json:"event_timestamp,omitempty"`
	RecordedAt     time.Time `json:"recorded_at"`
}

type notionPageExtraction struct {
	Body                    string
	LinkedPages             []string
	LinkedDatabases         []string
	LinkedDataSources       []string
	OutboundReferences      []companyknowledge.NotionOutboundReference
	BlockPaginationComplete bool
	Truncated               bool
}

type notionMirrorRunner struct {
	cfg             config.Config
	api             notionAPI
	mirror          *companyknowledge.NotionMirror
	store           store.SourceMirrorWriteStore
	workspace       string
	seenPages       map[string]struct{}
	seenDBs         map[string]struct{}
	seenDataSources map[string]struct{}
	maxBlocks       int
	maxDepth        int
	maxDBs          int
	pageCount       int
	dbCount         int
	truncated       bool
	checkpoint      notionMirrorCheckpoint
}

type notionAPI interface {
	RetrievePage(ctx context.Context, pageID string) (clients.NotionPage, error)
	RetrieveDatabase(ctx context.Context, databaseID string) (clients.NotionDatabase, error)
	RetrieveDataSource(ctx context.Context, dataSourceID string) (clients.NotionDataSource, error)
	RetrievePageMarkdown(ctx context.Context, pageID string, includeTranscript bool) (clients.NotionPageMarkdown, error)
	ListBlockChildren(ctx context.Context, blockID string, cursor string, pageSize int) (clients.NotionListResponse[clients.NotionBlock], error)
	QueryDataSource(ctx context.Context, dataSourceID string, opts clients.NotionDataSourceQueryOptions) (clients.NotionListResponse[clients.NotionPage], error)
}

func RunNotionMirror(ctx context.Context, cfg config.Config, mirrorStore store.SourceMirrorWriteStore) error {
	if !cfg.NotionMirrorEnabled {
		return errors.New("notion mirror is disabled")
	}
	if mirrorStore == nil {
		return errors.New("configured store does not support source mirror idempotency")
	}
	roots := uniqueNonEmpty(cfg.NotionMirrorAllowlist)
	if len(roots) == 0 {
		return errors.New("RSI_NOTION_MIRROR_ALLOWLIST is empty")
	}
	sort.Strings(roots)
	api := clients.NewNotionClientWithConfig(clients.NotionClientOptions{
		BaseURL:           cfg.NotionAPIBaseURL,
		Token:             cfg.NotionToken,
		Version:           cfg.NotionAPIVersion,
		RequestsPerSecond: cfg.NotionMirrorRequestsPerSecond,
		MaxRetries:        cfg.NotionMirrorMaxRetries,
		RetryBaseDelay:    cfg.NotionMirrorRetryBaseDelay,
	})
	mirror := companyknowledge.NewNotionMirror(mirrorStore, clients.NewHonchoClientWithAPIKey(cfg.HonchoBaseURL, cfg.HonchoAPIKey), companyknowledge.NotionMirrorOptions{
		Environment:     cfg.Environment,
		HonchoWorkspace: cfg.HonchoWorkspaceID,
	})
	for _, rootID := range roots {
		runner, err := newNotionMirrorRunner(cfg, api, mirrorStore, mirror, rootID)
		if err != nil {
			return err
		}
		if err := runner.mirrorRoot(ctx, rootID); err != nil {
			return err
		}
		runner.checkpoint.LastCompletedAt = time.Now().UTC()
		runner.checkpoint.LastPageCount = runner.pageCount
		runner.checkpoint.LastDatabaseCount = runner.dbCount
		if runner.truncated {
			runner.checkpoint.TraversalStatus = companyknowledge.NotionTraversalTruncated
		} else {
			runner.checkpoint.TraversalStatus = companyknowledge.NotionTraversalComplete
		}
		if err := writeNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, runner.checkpoint); err != nil {
			return err
		}
		log.Printf("notion mirror root=%s complete pages=%d databases=%d traversal=%s", rootID, runner.pageCount, runner.dbCount, runner.checkpoint.TraversalStatus)
		if runner.truncated {
			return fmt.Errorf("notion mirror root=%s produced a truncated traversal; refusing clean success", rootID)
		}
	}
	return nil
}

func newNotionMirrorRunner(cfg config.Config, api notionAPI, mirrorStore store.SourceMirrorWriteStore, mirror *companyknowledge.NotionMirror, rootID string) (*notionMirrorRunner, error) {
	checkpoint, err := readNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, rootID)
	if err != nil {
		return nil, err
	}
	if checkpoint.CompletedPages == nil {
		checkpoint.CompletedPages = map[string]string{}
	}
	if checkpoint.CompletedObjects == nil {
		checkpoint.CompletedObjects = map[string]notionMirrorCheckpointObject{}
	}
	if checkpoint.DataSources == nil {
		checkpoint.DataSources = map[string]notionMirrorDataSourceState{}
	}
	if checkpoint.DirtyObjects == nil {
		checkpoint.DirtyObjects = map[string]notionMirrorDirtyObject{}
	}
	workspace := "notion"
	maxBlocks := cfg.NotionMirrorMaxBlocksPerPage
	if maxBlocks <= 0 {
		maxBlocks = 1000
	}
	maxDepth := cfg.NotionMirrorMaxDepth
	if maxDepth <= 0 {
		maxDepth = 4
	}
	maxDBs := cfg.NotionMirrorMaxDatabasesPerRoot
	if maxDBs <= 0 {
		maxDBs = 50
	}
	return &notionMirrorRunner{
		cfg:             cfg,
		api:             api,
		mirror:          mirror,
		store:           mirrorStore,
		workspace:       workspace,
		seenPages:       map[string]struct{}{},
		seenDBs:         map[string]struct{}{},
		seenDataSources: map[string]struct{}{},
		maxBlocks:       maxBlocks,
		maxDepth:        maxDepth,
		maxDBs:          maxDBs,
		checkpoint:      checkpoint,
	}, nil
}

func (r *notionMirrorRunner) mirrorRoot(ctx context.Context, rootID string) error {
	rootID = normalizeNotionID(rootID)
	r.checkpoint.RootID = rootID
	r.checkpoint.WorkspaceID = r.workspace
	log.Printf("notion mirror root=%s starting", rootID)
	if err := r.mirrorDirtyObjects(ctx, rootID); err != nil {
		return err
	}
	page, pageErr := r.api.RetrievePage(ctx, rootID)
	if pageErr == nil {
		if err := r.mirrorLoadedPage(ctx, rootID, rootID, nil, page); err != nil {
			return fmt.Errorf("mirror notion page root=%s: %w", rootID, err)
		}
		return nil
	} else if !isNotionNotFound(pageErr) && !isNotionPageEndpointTypeMismatch(pageErr) {
		return fmt.Errorf("retrieve notion page root=%s: %w", rootID, pageErr)
	}
	if err := r.mirrorDatabase(ctx, rootID, rootID, nil); err == nil {
		return nil
	} else if !isNotionNotFound(err) && !isNotionDatabaseEndpointTypeMismatch(err) {
		return fmt.Errorf("mirror notion database root=%s: %w", rootID, err)
	}
	dataSource, dataSourceErr := r.api.RetrieveDataSource(ctx, rootID)
	if dataSourceErr == nil {
		if err := r.mirrorLoadedDataSource(ctx, rootID, "", rootID, nil, dataSource); err != nil {
			return fmt.Errorf("mirror notion data source root=%s: %w", rootID, err)
		}
		return nil
	} else if !isNotionNotFound(dataSourceErr) && !isNotionDataSourceEndpointTypeMismatch(dataSourceErr) {
		return fmt.Errorf("retrieve notion data source root=%s: %w", rootID, dataSourceErr)
	}
	return fmt.Errorf("notion allowlist root %s is neither a visible page, database, nor data source", rootID)
}

func (r *notionMirrorRunner) mirrorDirtyObjects(ctx context.Context, rootID string) error {
	if len(r.checkpoint.DirtyObjects) == 0 {
		return nil
	}
	keys := make([]string, 0, len(r.checkpoint.DirtyObjects))
	for key := range r.checkpoint.DirtyObjects {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		dirty := r.checkpoint.DirtyObjects[key]
		objectID := normalizeNotionID(dirty.ObjectID)
		if objectID == "" {
			delete(r.checkpoint.DirtyObjects, key)
			continue
		}
		switch strings.TrimSpace(dirty.ObjectKind) {
		case companyknowledge.NotionObjectKindPage, "":
			if err := r.mirrorPage(ctx, objectID, rootID, nil); err != nil {
				return fmt.Errorf("mirror dirty notion page=%s: %w", objectID, err)
			}
		case companyknowledge.NotionObjectKindDatabase:
			if err := r.mirrorDatabase(ctx, objectID, rootID, nil); err != nil {
				return fmt.Errorf("mirror dirty notion database=%s: %w", objectID, err)
			}
		case companyknowledge.NotionObjectKindDataSource:
			if err := r.mirrorDataSource(ctx, objectID, "", rootID, nil); err != nil {
				return fmt.Errorf("mirror dirty notion data source=%s: %w", objectID, err)
			}
		default:
			log.Printf("notion mirror root=%s skipping dirty object=%s unsupported kind=%s", rootID, objectID, dirty.ObjectKind)
		}
		delete(r.checkpoint.DirtyObjects, key)
		r.checkpoint.LastProgressAt = time.Now().UTC()
		if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
			return err
		}
	}
	return nil
}

func (r *notionMirrorRunner) mirrorPage(ctx context.Context, pageID string, rootID string, hierarchy []string) error {
	pageID = normalizeNotionID(pageID)
	if pageID == "" {
		return nil
	}
	if !r.claimPage(pageID) {
		return nil
	}
	page, err := r.api.RetrievePage(ctx, pageID)
	if err != nil {
		if isNotionNotFound(err) {
			return r.recordCrawlMiss(ctx, rootID, pageID, "page", "notion page was not reachable")
		}
		if isNotionPageEndpointTypeMismatch(err) {
			dbErr := r.mirrorDatabase(ctx, pageID, rootID, hierarchy)
			if dbErr == nil {
				return nil
			}
			if !isNotionNotFound(dbErr) && !isNotionDatabaseEndpointTypeMismatch(dbErr) {
				return dbErr
			}
			dataSourceErr := r.mirrorDataSource(ctx, pageID, "", rootID, hierarchy)
			if dataSourceErr == nil {
				return nil
			}
			if !isNotionNotFound(dataSourceErr) && !isNotionDataSourceEndpointTypeMismatch(dataSourceErr) {
				return dataSourceErr
			}
			return r.recordCrawlMiss(ctx, rootID, pageID, "page", "notion object type mismatch and fallback endpoints failed")
		}
		return err
	}
	return r.mirrorPageData(ctx, pageID, rootID, hierarchy, page)
}

func (r *notionMirrorRunner) mirrorLoadedPage(ctx context.Context, pageID string, rootID string, hierarchy []string, page clients.NotionPage) error {
	pageID = normalizeNotionID(pageID)
	if pageID == "" {
		return nil
	}
	if !r.claimPage(pageID) {
		return nil
	}
	return r.mirrorPageData(ctx, pageID, rootID, hierarchy, page)
}

func shouldPublishNotionWikiSource(result companyknowledge.NotionMirrorResult) bool {
	return !result.Skipped
}

func (r *notionMirrorRunner) claimPage(pageID string) bool {
	if _, seen := r.seenPages[pageID]; seen {
		return false
	}
	r.seenPages[pageID] = struct{}{}
	return true
}

func (r *notionMirrorRunner) mirrorPageData(ctx context.Context, pageID string, rootID string, hierarchy []string, page clients.NotionPage) error {
	if notionObjectInTrash(page.Archived, page.InTrash) {
		return r.markNotionObjectStale(pageID, companyknowledge.NotionObjectKindPage, rootID, "notion page is in trash", map[string]any{
			"archived": page.Archived,
			"in_trash": page.InTrash,
		})
	}
	title := notionPageTitle(page)
	if title == "" {
		title = pageID
	}
	currentHierarchy := append(append([]string{}, hierarchy...), title)
	input := notionPageDocumentInput(r.cfg, r.workspace, rootID, currentHierarchy, pageID, title, page, "", false, nil)
	revision := companyknowledge.NotionDocumentSourceRevision(input)
	sourceKey := companyknowledge.NotionObjectSourceKey(input.WorkspaceID, input.ObjectKind, input.ObjectID)
	skipped, cachedChildren, err := r.shouldSkipNotionPage(sourceKey, revision)
	if err != nil {
		return err
	}
	if skipped {
		r.checkpoint.CompletedPages[sourceKey] = revision
		cachedChildren.UpdatedAt = time.Now().UTC()
		r.checkpoint.CompletedObjects[sourceKey] = cachedChildren
		r.checkpoint.LastPageID = pageID
		r.checkpoint.LastProgressAt = time.Now().UTC()
		r.pageCount++
		if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
			return err
		}
		log.Printf("notion mirror page=%s skipped=true reason=already_complete", pageID)
		return r.traverseCachedNotionChildren(ctx, rootID, currentHierarchy, cachedChildren)
	}
	extraction, err := r.extractPageMarkdown(ctx, pageID)
	if err != nil {
		return fmt.Errorf("extract notion markdown page=%s: %w", pageID, err)
	}
	body := extraction.Body
	truncated := extraction.Truncated
	outboundReferences := capOutboundReferences(extraction.OutboundReferences, 200)
	if len(extraction.OutboundReferences) > 200 {
		truncated = true
	}
	if truncated {
		r.truncated = true
	}
	input = notionPageDocumentInput(r.cfg, r.workspace, rootID, currentHierarchy, pageID, title, page, body, truncated, outboundReferences)
	revision = companyknowledge.NotionDocumentSourceRevision(input)
	sourceKey = companyknowledge.NotionObjectSourceKey(input.WorkspaceID, input.ObjectKind, input.ObjectID)
	result, err := r.mirror.IngestDocument(ctx, input)
	if err != nil {
		return fmt.Errorf("mirror notion page=%s into honcho: %w", pageID, err)
	}
	if shouldPublishNotionWikiSource(result) {
		if _, err := companyknowledge.RecordEnqueueAndMaybePublishWikiSource(ctx, r.cfg, r.store, companyknowledge.NotionWikiSourceRevisionInput(input)); err != nil {
			return fmt.Errorf("publish notion wiki source page=%s: %w", pageID, err)
		}
	}
	childPages := uniqueNonEmpty(extraction.LinkedPages)
	childDatabases := uniqueNonEmpty(extraction.LinkedDatabases)
	childDataSources := uniqueNonEmpty(extraction.LinkedDataSources)
	r.checkpoint.CompletedPages[sourceKey] = revision
	r.checkpoint.CompletedObjects[sourceKey] = notionMirrorCheckpointObject{
		SourceRevision:          revision,
		ChildPageIDs:            childPages,
		ChildDatabaseIDs:        childDatabases,
		ChildDataSourceIDs:      childDataSources,
		BlockPaginationComplete: extraction.BlockPaginationComplete,
		Truncated:               truncated,
		UpdatedAt:               time.Now().UTC(),
	}
	r.checkpoint.LastPageID = pageID
	r.checkpoint.LastProgressAt = time.Now().UTC()
	r.pageCount++
	if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
		return err
	}
	log.Printf("notion mirror page=%s status=%s skipped=%t reason=%s", pageID, result.Status, result.Skipped, result.SkipReason)
	for _, childDataSourceID := range childDataSources {
		if err := r.mirrorDataSource(ctx, childDataSourceID, "", rootID, currentHierarchy); err != nil {
			return err
		}
	}
	for _, childDatabaseID := range childDatabases {
		if err := r.mirrorDatabase(ctx, childDatabaseID, rootID, currentHierarchy); err != nil {
			return err
		}
	}
	for _, childPageID := range childPages {
		if err := r.mirrorPage(ctx, childPageID, rootID, currentHierarchy); err != nil {
			return err
		}
	}
	return nil
}

func notionPageDocumentInput(cfg config.Config, workspace string, rootID string, hierarchy []string, pageID string, title string, page clients.NotionPage, body string, truncated bool, outboundReferences []companyknowledge.NotionOutboundReference) companyknowledge.NotionDocumentInput {
	return companyknowledge.NotionDocumentInput{
		WorkspaceID:        workspace,
		ObjectKind:         companyknowledge.NotionObjectKindPage,
		ObjectID:           pageID,
		PageID:             pageID,
		RootID:             normalizeNotionID(rootID),
		ParentID:           notionParentID(page.Parent),
		DatabaseID:         notionParentDatabaseID(page.Parent),
		Title:              title,
		URL:                strings.TrimSpace(page.URL),
		LastEditedTime:     strings.TrimSpace(page.LastEditedTime),
		CreatedTime:        strings.TrimSpace(page.CreatedTime),
		Content:            body,
		TraversalStatus:    traversalStatus(truncated),
		Truncated:          truncated,
		Allowlisted:        notionRootAllowedByConfig(cfg, rootID),
		Hierarchy:          hierarchy,
		OutboundReferences: outboundReferences,
		Raw:                map[string]any{"object": page.Object},
	}
}

func (r *notionMirrorRunner) shouldSkipNotionPage(sourceKey string, revision string) (bool, notionMirrorCheckpointObject, error) {
	cached, ok := r.checkpoint.CompletedObjects[sourceKey]
	if !ok || cached.SourceRevision != revision || !cached.BlockPaginationComplete || cached.Truncated {
		return false, notionMirrorCheckpointObject{}, nil
	}
	record, found, err := r.store.GetSourceMirrorRecord(companyknowledge.NotionDocumentSourceType, sourceKey)
	if err != nil {
		return false, notionMirrorCheckpointObject{}, err
	}
	if !found || record.Status != store.SourceMirrorStatusComplete || record.SourceRevision != revision {
		return false, notionMirrorCheckpointObject{}, nil
	}
	if strings.TrimSpace(record.HonchoObjectID) == "" && strings.TrimSpace(record.HonchoMessageID) == "" {
		return false, notionMirrorCheckpointObject{}, nil
	}
	return true, cached, nil
}

func (r *notionMirrorRunner) traverseCachedNotionChildren(ctx context.Context, rootID string, hierarchy []string, cached notionMirrorCheckpointObject) error {
	for _, dataSourceID := range cached.ChildDataSourceIDs {
		if err := r.mirrorDataSource(ctx, dataSourceID, "", rootID, hierarchy); err != nil {
			return err
		}
	}
	for _, databaseID := range cached.ChildDatabaseIDs {
		if err := r.mirrorDatabase(ctx, databaseID, rootID, hierarchy); err != nil {
			return err
		}
	}
	for _, pageID := range cached.ChildPageIDs {
		if err := r.mirrorPage(ctx, pageID, rootID, hierarchy); err != nil {
			return err
		}
	}
	return nil
}

func (r *notionMirrorRunner) mirrorDatabase(ctx context.Context, databaseID string, rootID string, hierarchy []string) error {
	databaseID = normalizeNotionID(databaseID)
	if databaseID == "" {
		return nil
	}
	if _, seen := r.seenDBs[databaseID]; seen {
		return nil
	}
	r.seenDBs[databaseID] = struct{}{}
	database, err := r.api.RetrieveDatabase(ctx, databaseID)
	if err != nil {
		if isNotionNotFound(err) {
			if dataSource, dataSourceErr := r.api.RetrieveDataSource(ctx, databaseID); dataSourceErr == nil {
				return r.mirrorLoadedDataSource(ctx, databaseID, "", rootID, hierarchy, dataSource)
			} else if !isNotionNotFound(dataSourceErr) && !isNotionDataSourceEndpointTypeMismatch(dataSourceErr) {
				return dataSourceErr
			}
			isRoot := databaseID == normalizeNotionID(rootID) && hierarchy == nil
			if isRoot {
				return err
			}
			return r.recordCrawlMiss(ctx, rootID, databaseID, "database", "notion database was not reachable")
		}
		if isNotionDatabaseEndpointTypeMismatch(err) {
			if dataSource, dataSourceErr := r.api.RetrieveDataSource(ctx, databaseID); dataSourceErr == nil {
				return r.mirrorLoadedDataSource(ctx, databaseID, "", rootID, hierarchy, dataSource)
			} else if !isNotionNotFound(dataSourceErr) && !isNotionDataSourceEndpointTypeMismatch(dataSourceErr) {
				return dataSourceErr
			}
			return r.mirrorPage(ctx, databaseID, rootID, hierarchy)
		}
		return err
	}
	if notionObjectInTrash(database.Archived, database.InTrash) {
		return r.markNotionObjectStale(databaseID, companyknowledge.NotionObjectKindDatabase, rootID, "notion database is in trash", map[string]any{
			"archived": database.Archived,
			"in_trash": database.InTrash,
		})
	}
	r.dbCount++
	if r.dbCount > r.maxDBs {
		r.truncated = true
		return fmt.Errorf("notion mirror root=%s exceeded max databases per root (%d)", rootID, r.maxDBs)
	}
	title := strings.TrimSpace(richTextPlainText(database.Title))
	if title == "" {
		title = databaseID
	}
	currentHierarchy := append(append([]string{}, hierarchy...), title)
	schemaSummary, schemaHash := companyknowledge.NotionDatabaseSchemaSummary(database.Properties)
	databaseInput := companyknowledge.NotionDocumentInput{
		WorkspaceID:     r.workspace,
		ObjectKind:      companyknowledge.NotionObjectKindDatabase,
		ObjectID:        databaseID,
		DatabaseID:      databaseID,
		RootID:          normalizeNotionID(rootID),
		ParentID:        notionParentID(database.Parent),
		Title:           title,
		URL:             strings.TrimSpace(database.URL),
		LastEditedTime:  strings.TrimSpace(database.LastEditedTime),
		CreatedTime:     strings.TrimSpace(database.CreatedTime),
		Content:         "Data sources in this database are mirrored as separate Notion data source documents.",
		SchemaSummary:   schemaSummary,
		SchemaHash:      schemaHash,
		TraversalStatus: companyknowledge.NotionTraversalComplete,
		Allowlisted:     notionRootAllowedByConfig(r.cfg, rootID),
		Hierarchy:       currentHierarchy,
		Raw:             map[string]any{"object": database.Object},
	}
	databaseRevision := companyknowledge.NotionDocumentSourceRevision(databaseInput)
	databaseSourceKey := companyknowledge.NotionObjectSourceKey(databaseInput.WorkspaceID, databaseInput.ObjectKind, databaseInput.ObjectID)
	databaseResult, err := r.mirror.IngestDocument(ctx, databaseInput)
	if err != nil {
		return fmt.Errorf("mirror notion database=%s into honcho: %w", databaseID, err)
	}
	if shouldPublishNotionWikiSource(databaseResult) {
		if _, err := companyknowledge.RecordEnqueueAndMaybePublishWikiSource(ctx, r.cfg, r.store, companyknowledge.NotionWikiSourceRevisionInput(databaseInput)); err != nil {
			return fmt.Errorf("publish notion wiki source database=%s: %w", databaseID, err)
		}
	}
	r.checkpoint.CompletedObjects[databaseSourceKey] = notionMirrorCheckpointObject{
		SourceRevision:          databaseRevision,
		ChildDataSourceIDs:      uniqueNonEmpty(notionDatabaseDataSourceIDs(database)),
		BlockPaginationComplete: true,
		UpdatedAt:               time.Now().UTC(),
	}
	r.checkpoint.LastProgressAt = time.Now().UTC()
	if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
		return err
	}
	log.Printf("notion mirror database=%s status=%s skipped=%t reason=%s", databaseID, databaseResult.Status, databaseResult.Skipped, databaseResult.SkipReason)
	for _, dataSourceID := range notionDatabaseDataSourceIDs(database) {
		if err := r.mirrorDataSource(ctx, dataSourceID, databaseID, rootID, currentHierarchy); err != nil {
			return err
		}
	}
	return nil
}

func (r *notionMirrorRunner) mirrorDataSource(ctx context.Context, dataSourceID string, databaseID string, rootID string, hierarchy []string) error {
	dataSourceID = normalizeNotionID(dataSourceID)
	if dataSourceID == "" {
		return nil
	}
	if _, seen := r.seenDataSources[dataSourceID]; seen {
		return nil
	}
	r.seenDataSources[dataSourceID] = struct{}{}
	dataSource, err := r.api.RetrieveDataSource(ctx, dataSourceID)
	if err != nil {
		if isNotionNotFound(err) {
			return r.recordCrawlMiss(ctx, rootID, dataSourceID, "data_source", "notion data source was not reachable")
		}
		if isNotionDataSourceEndpointTypeMismatch(err) {
			return r.mirrorDatabase(ctx, dataSourceID, rootID, hierarchy)
		}
		return err
	}
	return r.mirrorClaimedDataSource(ctx, dataSourceID, databaseID, rootID, hierarchy, dataSource)
}

func (r *notionMirrorRunner) mirrorLoadedDataSource(ctx context.Context, dataSourceID string, databaseID string, rootID string, hierarchy []string, dataSource clients.NotionDataSource) error {
	dataSourceID = normalizeNotionID(dataSourceID)
	if dataSourceID == "" {
		return nil
	}
	if _, seen := r.seenDataSources[dataSourceID]; seen {
		return nil
	}
	r.seenDataSources[dataSourceID] = struct{}{}
	return r.mirrorClaimedDataSource(ctx, dataSourceID, databaseID, rootID, hierarchy, dataSource)
}

func (r *notionMirrorRunner) mirrorClaimedDataSource(ctx context.Context, dataSourceID string, databaseID string, rootID string, hierarchy []string, dataSource clients.NotionDataSource) error {
	if notionObjectInTrash(dataSource.Archived, dataSource.InTrash) {
		return r.markNotionObjectStale(dataSourceID, companyknowledge.NotionObjectKindDataSource, rootID, "notion data source is in trash", map[string]any{
			"archived":    dataSource.Archived,
			"in_trash":    dataSource.InTrash,
			"database_id": normalizeNotionID(databaseID),
		})
	}
	r.dbCount++
	if r.dbCount > r.maxDBs {
		r.truncated = true
		return fmt.Errorf("notion mirror root=%s exceeded max databases per root (%d)", rootID, r.maxDBs)
	}
	title := notionDataSourceTitle(dataSource)
	if title == "" {
		title = dataSourceID
	}
	currentHierarchy := append(append([]string{}, hierarchy...), title)
	schemaSummary, schemaHash := companyknowledge.NotionDatabaseSchemaSummary(dataSource.Properties)
	dataSourceInput := companyknowledge.NotionDocumentInput{
		WorkspaceID:     r.workspace,
		ObjectKind:      companyknowledge.NotionObjectKindDataSource,
		ObjectID:        dataSourceID,
		DataSourceID:    dataSourceID,
		DatabaseID:      normalizeNotionID(firstNonEmpty(databaseID, notionDataSourceDatabaseID(dataSource))),
		RootID:          normalizeNotionID(rootID),
		ParentID:        notionParentID(firstNonNilMap(dataSource.Parent, dataSource.DatabaseParent)),
		Title:           title,
		URL:             strings.TrimSpace(dataSource.URL),
		LastEditedTime:  strings.TrimSpace(dataSource.LastEditedTime),
		CreatedTime:     strings.TrimSpace(dataSource.CreatedTime),
		Content:         "Row pages in this Notion data source are mirrored as separate Notion documents.",
		SchemaSummary:   schemaSummary,
		SchemaHash:      schemaHash,
		TraversalStatus: companyknowledge.NotionTraversalComplete,
		Allowlisted:     notionRootAllowedByConfig(r.cfg, rootID),
		Hierarchy:       currentHierarchy,
		Raw:             map[string]any{"object": dataSource.Object},
	}
	dataSourceRevision := companyknowledge.NotionDocumentSourceRevision(dataSourceInput)
	dataSourceKey := companyknowledge.NotionObjectSourceKey(dataSourceInput.WorkspaceID, dataSourceInput.ObjectKind, dataSourceInput.ObjectID)
	dataSourceResult, err := r.mirror.IngestDocument(ctx, dataSourceInput)
	if err != nil {
		return fmt.Errorf("mirror notion data source=%s into honcho: %w", dataSourceID, err)
	}
	if shouldPublishNotionWikiSource(dataSourceResult) {
		if _, err := companyknowledge.RecordEnqueueAndMaybePublishWikiSource(ctx, r.cfg, r.store, companyknowledge.NotionWikiSourceRevisionInput(dataSourceInput)); err != nil {
			return fmt.Errorf("publish notion wiki source data source=%s: %w", dataSourceID, err)
		}
	}
	r.checkpoint.CompletedObjects[dataSourceKey] = notionMirrorCheckpointObject{
		SourceRevision:          dataSourceRevision,
		BlockPaginationComplete: true,
		UpdatedAt:               time.Now().UTC(),
	}
	if r.checkpoint.DataSources == nil {
		r.checkpoint.DataSources = map[string]notionMirrorDataSourceState{}
	}
	state := r.checkpoint.DataSources[dataSourceID]
	state.DataSourceID = dataSourceID
	state.DatabaseID = strings.TrimSpace(dataSourceInput.DatabaseID)
	state.UpdatedAt = time.Now().UTC()
	r.checkpoint.DataSources[dataSourceID] = state
	r.checkpoint.LastProgressAt = time.Now().UTC()
	if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
		return err
	}
	log.Printf("notion mirror data_source=%s status=%s skipped=%t reason=%s", dataSourceID, dataSourceResult.Status, dataSourceResult.Skipped, dataSourceResult.SkipReason)
	return r.mirrorDataSourceRows(ctx, dataSourceID, rootID, currentHierarchy)
}

func (r *notionMirrorRunner) mirrorDataSourceRows(ctx context.Context, dataSourceID string, rootID string, hierarchy []string) error {
	state := r.checkpoint.DataSources[dataSourceID]
	mode := "full"
	filterAfter := ""
	if r.shouldDeltaScanDataSource(state) {
		mode = "delta"
		filterAfter = notionTimestampWithLookback(state.RowHighWatermark, r.cfg.NotionMirrorDeltaLookback)
	}
	cursor := ""
	observed := map[string]struct{}{}
	previousObserved := stringSet(state.ObservedPageIDs)
	for {
		page, err := r.api.QueryDataSource(ctx, dataSourceID, clients.NotionDataSourceQueryOptions{
			Cursor:                  cursor,
			PageSize:                100,
			LastEditedTimeOnOrAfter: filterAfter,
			SortTimestamp:           "last_edited_time",
			SortDirection:           "ascending",
			FilterProperties:        []string{"title"},
		})
		if err != nil {
			return err
		}
		for _, result := range page.Results {
			pageID := normalizeNotionID(result.ID)
			if pageID == "" {
				continue
			}
			observed[pageID] = struct{}{}
			if err := r.mirrorLoadedPage(ctx, pageID, rootID, hierarchy, result); err != nil {
				return err
			}
			state.RowHighWatermark = maxNotionTimestamp(state.RowHighWatermark, result.LastEditedTime)
			state.UpdatedAt = time.Now().UTC()
			r.checkpoint.DataSources[dataSourceID] = state
			r.checkpoint.LastProgressAt = state.UpdatedAt
			if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
				return err
			}
		}
		cursor = strings.TrimSpace(page.NextCursor)
		if !page.HasMore || cursor == "" {
			break
		}
	}
	now := time.Now().UTC()
	state.UpdatedAt = now
	if mode == "delta" {
		state.LastDeltaCompletedAt = now
	} else {
		state.LastFullScanCompletedAt = now
		state.ObservedPageIDs = sortedSetKeys(observed)
		for pageID := range previousObserved {
			if _, stillObserved := observed[pageID]; !stillObserved {
				if err := r.markNotionObjectStale(pageID, companyknowledge.NotionObjectKindPage, rootID, "notion data source row was not observed during full scan", map[string]any{
					"data_source_id": dataSourceID,
				}); err != nil {
					return err
				}
			}
		}
	}
	r.checkpoint.DataSources[dataSourceID] = state
	r.checkpoint.LastProgressAt = now
	if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
		return err
	}
	log.Printf("notion mirror data_source=%s mode=%s rows=%d high_watermark=%s", dataSourceID, mode, len(observed), state.RowHighWatermark)
	return nil
}

func (r *notionMirrorRunner) shouldDeltaScanDataSource(state notionMirrorDataSourceState) bool {
	if !r.cfg.NotionMirrorDeltaEnabled {
		return false
	}
	if strings.TrimSpace(state.RowHighWatermark) == "" || state.LastFullScanCompletedAt.IsZero() {
		return false
	}
	interval := r.cfg.NotionMirrorFullScanInterval
	if interval <= 0 {
		return true
	}
	return time.Since(state.LastFullScanCompletedAt) < interval
}

func notionRootAllowedByConfig(cfg config.Config, rootID string) bool {
	rootID = normalizeNotionID(rootID)
	for _, item := range cfg.NotionMirrorAllowlist {
		if normalizeNotionID(item) == rootID {
			return true
		}
	}
	return false
}

func (r *notionMirrorRunner) extractPageMarkdown(ctx context.Context, pageID string) (notionPageExtraction, error) {
	markdown, err := r.api.RetrievePageMarkdown(ctx, pageID, false)
	if err != nil {
		return notionPageExtraction{}, err
	}
	bodyParts := []string{strings.TrimSpace(markdown.Markdown)}
	refs := notionMarkdownReferences(pageID, markdown.Markdown)
	truncated := false
	unknownIDs := uniqueNonEmpty(markdown.UnknownBlockIDs)
	if markdown.Truncated && len(unknownIDs) > 0 {
		if len(unknownIDs) > 100 {
			unknownIDs = unknownIDs[:100]
			truncated = true
		}
		for _, unknownID := range unknownIDs {
			child, err := r.api.RetrievePageMarkdown(ctx, unknownID, false)
			if err != nil {
				refs = append(refs, companyknowledge.NotionOutboundReference{
					ReferenceKind: "unknown_block",
					TargetID:      normalizeNotionID(unknownID),
					Reason:        "notion markdown reported unknown block that could not be fetched",
				})
				continue
			}
			if strings.TrimSpace(child.Markdown) != "" {
				bodyParts = append(bodyParts, strings.TrimSpace(child.Markdown))
			}
			refs = append(refs, notionMarkdownReferences(unknownID, child.Markdown)...)
			if child.Truncated && len(child.UnknownBlockIDs) > 0 {
				truncated = true
				refs = append(refs, companyknowledge.NotionOutboundReference{
					ReferenceKind: "unknown_block",
					SourceBlockID: normalizeNotionID(unknownID),
					Reason:        "nested notion markdown remained truncated",
				})
			}
		}
	} else if markdown.Truncated {
		truncated = true
	}
	body := strings.TrimSpace(strings.Join(nonEmptyStrings(bodyParts...), "\n\n"))
	linkedPages, linkedDatabases, linkedDataSources := notionMarkdownLinkedObjects(body)
	return notionPageExtraction{
		Body:                    body,
		LinkedPages:             linkedPages,
		LinkedDatabases:         linkedDatabases,
		LinkedDataSources:       linkedDataSources,
		OutboundReferences:      refs,
		BlockPaginationComplete: !truncated,
		Truncated:               truncated,
	}, nil
}

func (r *notionMirrorRunner) extractPageBody(ctx context.Context, pageID string, depth int, seenBlocks *int) (notionPageExtraction, error) {
	var lines []string
	var linkedPages []string
	var linkedDatabases []string
	var references []companyknowledge.NotionOutboundReference
	paginationComplete := true
	truncated := false
	cursor := ""
	for {
		page, err := r.api.ListBlockChildren(ctx, pageID, cursor, 100)
		if err != nil {
			return notionPageExtraction{}, err
		}
		for _, block := range page.Results {
			if block.Archived || block.InTrash {
				continue
			}
			*seenBlocks++
			if *seenBlocks > r.maxBlocks {
				lines = append(lines, "[Mirror truncated: Notion page exceeded block extraction limit.]")
				paginationComplete = false
				truncated = true
				return notionPageExtraction{
					Body:                    strings.Join(lines, "\n"),
					LinkedPages:             uniqueNonEmpty(linkedPages),
					LinkedDatabases:         uniqueNonEmpty(linkedDatabases),
					OutboundReferences:      references,
					BlockPaginationComplete: paginationComplete,
					Truncated:               truncated,
				}, nil
			}
			references = append(references, notionBlockReferences(block)...)
			line := notionBlockMarkdown(block, depth)
			if line != "" {
				lines = append(lines, line)
			}
			switch block.Type {
			case "child_page":
				if block.ID != "" {
					linkedPages = append(linkedPages, normalizeNotionID(block.ID))
				}
			case "child_database":
				if block.ID != "" {
					linkedDatabases = append(linkedDatabases, normalizeNotionID(block.ID))
				}
			default:
				if block.HasChildren && depth < r.maxDepth {
					childExtraction, err := r.extractPageBody(ctx, block.ID, depth+1, seenBlocks)
					if err != nil {
						return notionPageExtraction{}, err
					}
					if childExtraction.Body != "" {
						lines = append(lines, childExtraction.Body)
					}
					linkedPages = append(linkedPages, childExtraction.LinkedPages...)
					linkedDatabases = append(linkedDatabases, childExtraction.LinkedDatabases...)
					references = append(references, childExtraction.OutboundReferences...)
					if !childExtraction.BlockPaginationComplete {
						paginationComplete = false
					}
					if childExtraction.Truncated {
						truncated = true
					}
				} else if block.HasChildren {
					paginationComplete = false
					truncated = true
					r.truncated = true
				}
			}
		}
		cursor = strings.TrimSpace(page.NextCursor)
		if !page.HasMore || cursor == "" {
			break
		}
	}
	return notionPageExtraction{
		Body:                    strings.Join(lines, "\n"),
		LinkedPages:             uniqueNonEmpty(linkedPages),
		LinkedDatabases:         uniqueNonEmpty(linkedDatabases),
		OutboundReferences:      references,
		BlockPaginationComplete: paginationComplete,
		Truncated:               truncated,
	}, nil
}

func readNotionMirrorCheckpoint(root string, rootID string) (notionMirrorCheckpoint, error) {
	path := notionMirrorCheckpointPath(root, rootID)
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return notionMirrorCheckpoint{
			RootID:           normalizeNotionID(rootID),
			CompletedPages:   map[string]string{},
			CompletedObjects: map[string]notionMirrorCheckpointObject{},
			DataSources:      map[string]notionMirrorDataSourceState{},
			DirtyObjects:     map[string]notionMirrorDirtyObject{},
		}, nil
	}
	if err != nil {
		return notionMirrorCheckpoint{}, err
	}
	var checkpoint notionMirrorCheckpoint
	if err := json.Unmarshal(raw, &checkpoint); err != nil {
		return notionMirrorCheckpoint{}, fmt.Errorf("decode notion mirror checkpoint %s: %w", path, err)
	}
	if checkpoint.CompletedPages == nil {
		checkpoint.CompletedPages = map[string]string{}
	}
	if checkpoint.CompletedObjects == nil {
		checkpoint.CompletedObjects = map[string]notionMirrorCheckpointObject{}
	}
	if checkpoint.DataSources == nil {
		checkpoint.DataSources = map[string]notionMirrorDataSourceState{}
	}
	if checkpoint.DirtyObjects == nil {
		checkpoint.DirtyObjects = map[string]notionMirrorDirtyObject{}
	}
	for sourceKey, revision := range checkpoint.CompletedPages {
		if strings.TrimSpace(sourceKey) == "" || strings.TrimSpace(revision) == "" {
			continue
		}
		if _, ok := checkpoint.CompletedObjects[sourceKey]; !ok {
			checkpoint.CompletedObjects[sourceKey] = notionMirrorCheckpointObject{
				SourceRevision:          revision,
				BlockPaginationComplete: false,
				Truncated:               false,
				UpdatedAt:               checkpoint.LastProgressAt,
			}
		}
	}
	return checkpoint, nil
}

func writeNotionMirrorCheckpoint(root string, checkpoint notionMirrorCheckpoint) error {
	path := notionMirrorCheckpointPath(root, checkpoint.RootID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	if dir, err := os.Open(filepath.Dir(path)); err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

func notionMirrorCheckpointPath(root string, rootID string) string {
	return filepath.Join(strings.TrimSpace(root), "notion", sanitizePathPart(normalizeNotionID(rootID))+".json")
}

func (r *notionMirrorRunner) markNotionObjectStale(objectID string, objectKind string, rootID string, reason string, metadata map[string]any) error {
	objectID = normalizeNotionID(objectID)
	if objectID == "" {
		return nil
	}
	record := store.SourceMirrorRecord{
		SourceType:       companyknowledge.NotionDocumentSourceType,
		SourceKey:        companyknowledge.NotionObjectSourceKey(r.workspace, objectKind, objectID),
		Workspace:        r.workspace,
		Environment:      strings.TrimSpace(r.cfg.Environment),
		SourceSessionKey: companyknowledge.NotionObjectSessionKey(r.workspace, objectKind, objectID),
		HonchoWorkspace:  r.mirror.HonchoWorkspace(),
		HonchoSessionID:  companyknowledge.HonchoCompatibleName("notion", companyknowledge.NotionObjectSessionKey(r.workspace, objectKind, objectID)),
		SourceRevision:   "stale:" + strings.TrimSpace(reason),
		Status:           store.SourceMirrorStatusPending,
		Metadata: mergeStringAnyMaps(map[string]any{
			"source":       "notion",
			"object_kind":  objectKind,
			"object_id":    objectID,
			"root_id":      normalizeNotionID(rootID),
			"stale_reason": strings.TrimSpace(reason),
		}, metadata),
	}
	_, err := r.store.MarkSourceMirrorRecordStale(record, reason, map[string]any{"stale_observed_at": time.Now().UTC().Format(time.RFC3339)})
	return err
}

func (r *notionMirrorRunner) recordCrawlMiss(ctx context.Context, rootID string, targetID string, targetKind string, reason string) error {
	_ = ctx
	targetID = normalizeNotionID(targetID)
	rootID = normalizeNotionID(rootID)
	if targetID == "" || rootID == "" {
		return nil
	}
	sourceKey := companyknowledge.NotionCrawlMissSourceKey(r.workspace, rootID, targetID)
	record := store.SourceMirrorRecord{
		SourceType:       companyknowledge.NotionCrawlMissSourceType,
		SourceKey:        sourceKey,
		Workspace:        r.workspace,
		Environment:      strings.TrimSpace(r.cfg.Environment),
		SourceSessionKey: "notion:" + r.workspace + ":crawl_miss:" + rootID,
		HonchoWorkspace:  r.mirror.HonchoWorkspace(),
		HonchoSessionID:  companyknowledge.HonchoCompatibleName("notion", "notion:"+r.workspace+":crawl_miss:"+rootID),
		SourceRevision:   "miss:" + targetID + ":" + strings.TrimSpace(reason),
		Status:           store.SourceMirrorStatusPending,
		Metadata: map[string]any{
			"source":             "notion",
			"source_key":         sourceKey,
			"root_id":            rootID,
			"target_id":          targetID,
			"target_object_kind": targetKind,
			"reason":             strings.TrimSpace(reason),
		},
	}
	_, err := r.store.MarkSourceMirrorRecordStale(record, reason, map[string]any{"miss_observed_at": time.Now().UTC().Format(time.RFC3339)})
	return err
}

func traversalStatus(truncated bool) string {
	if truncated {
		return companyknowledge.NotionTraversalTruncated
	}
	return companyknowledge.NotionTraversalComplete
}

func capOutboundReferences(refs []companyknowledge.NotionOutboundReference, max int) []companyknowledge.NotionOutboundReference {
	if max <= 0 || len(refs) <= max {
		return refs
	}
	out := append([]companyknowledge.NotionOutboundReference(nil), refs[:max]...)
	out = append(out, companyknowledge.NotionOutboundReference{
		ReferenceKind: "truncation",
		Reason:        fmt.Sprintf("outbound references capped at %d", max),
	})
	return out
}

func mergeStringAnyMaps(base map[string]any, overlay map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	for key, value := range overlay {
		out[key] = value
	}
	return out
}

var notionIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{32}$`)

func normalizeNotionID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return normalizeNotionURLID(parsed)
	}
	if id := normalizeStrictNotionID(value); id != "" {
		return id
	}
	compact := strings.ReplaceAll(value, "-", "")
	compact = strings.ReplaceAll(compact, "_", "")
	return strings.TrimSpace(compact)
}

func normalizeNotionURLID(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if id := normalizeStrictNotionID(segments[i]); id != "" {
			return id
		}
	}
	for _, values := range parsed.Query() {
		for _, raw := range values {
			if id := normalizeStrictNotionID(raw); id != "" {
				return id
			}
		}
	}
	return ""
}

func normalizeStrictNotionID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return normalizeNotionURLID(parsed)
	}
	lastDashIndex := strings.LastIndex(value, "-")
	if lastDashIndex >= 0 {
		candidate := strings.TrimSpace(value[lastDashIndex+1:])
		if id := compactNotionID(candidate); id != "" {
			return id
		}
	}
	return compactNotionID(value)
}

func compactNotionID(value string) string {
	compact := strings.ReplaceAll(value, "-", "")
	compact = strings.ReplaceAll(compact, "_", "")
	compact = strings.TrimSpace(compact)
	if notionIDPattern.MatchString(compact) {
		return strings.ToLower(compact)
	}
	return ""
}

func isNotionNotFound(err error) bool {
	var apiErr clients.NotionAPIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

func isNotionPageEndpointTypeMismatch(err error) bool {
	return isNotionRequestedTypeMismatch(err, "page")
}

func isNotionDatabaseEndpointTypeMismatch(err error) bool {
	return isNotionRequestedTypeMismatch(err, "database")
}

func isNotionDataSourceEndpointTypeMismatch(err error) bool {
	return isNotionRequestedTypeMismatch(err, "data source") || isNotionRequestedTypeMismatch(err, "data_source")
}

func isNotionRequestedTypeMismatch(err error, requested string) bool {
	var apiErr clients.NotionAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
		return false
	}
	body := strings.ToLower(apiErr.Body)
	requested = strings.ToLower(strings.TrimSpace(requested))
	return strings.Contains(body, "validation_error") && strings.Contains(body, "not a "+requested)
}

func notionObjectInTrash(archived bool, inTrash bool) bool {
	return archived || inTrash
}

func notionDatabaseDataSourceIDs(database clients.NotionDatabase) []string {
	ids := make([]string, 0, len(database.DataSources))
	for _, ref := range database.DataSources {
		if id := normalizeNotionID(ref.ID); id != "" {
			ids = append(ids, id)
		}
	}
	return uniqueNonEmpty(ids)
}

func notionDataSourceTitle(dataSource clients.NotionDataSource) string {
	if value := strings.TrimSpace(dataSource.Name); value != "" {
		return value
	}
	if value := strings.TrimSpace(richTextPlainText(dataSource.Title)); value != "" {
		return value
	}
	return ""
}

func notionDataSourceDatabaseID(dataSource clients.NotionDataSource) string {
	for _, parent := range []map[string]any{dataSource.DatabaseParent, dataSource.Parent} {
		if id := notionParentDatabaseID(parent); id != "" {
			return id
		}
	}
	return ""
}

func firstNonNilMap(values ...map[string]any) map[string]any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func notionPageTitle(page clients.NotionPage) string {
	for _, raw := range page.Properties {
		property, ok := raw.(map[string]any)
		if !ok || fmt.Sprint(property["type"]) != "title" {
			continue
		}
		return richTextPlainTextFromAny(property["title"])
	}
	return ""
}

func notionParentID(parent map[string]any) string {
	for _, key := range []string{"page_id", "block_id", "database_id", "workspace"} {
		if value := strings.TrimSpace(fmt.Sprint(parent[key])); value != "" && value != "<nil>" {
			return normalizeNotionID(value)
		}
	}
	return ""
}

func notionParentDatabaseID(parent map[string]any) string {
	if value := strings.TrimSpace(fmt.Sprint(parent["database_id"])); value != "" && value != "<nil>" {
		return normalizeNotionID(value)
	}
	return ""
}

var (
	notionMarkdownPageTagRE       = regexp.MustCompile(`(?i)<page\b[^>]*\burl="([^"]+)"`)
	notionMarkdownDatabaseTagRE   = regexp.MustCompile(`(?i)<database\b[^>]*\burl="([^"]+)"`)
	notionMarkdownDataSourceTagRE = regexp.MustCompile(`(?i)<data_source\b[^>]*\burl="([^"]+)"`)
)

func notionMarkdownLinkedObjects(markdown string) ([]string, []string, []string) {
	pages := notionMarkdownIDs(markdown, notionMarkdownPageTagRE)
	databases := notionMarkdownIDs(markdown, notionMarkdownDatabaseTagRE)
	dataSources := notionMarkdownIDs(markdown, notionMarkdownDataSourceTagRE)
	return pages, databases, dataSources
}

func notionMarkdownIDs(markdown string, re *regexp.Regexp) []string {
	matches := re.FindAllStringSubmatch(markdown, -1)
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		if id := normalizeNotionID(match[1]); id != "" {
			ids = append(ids, id)
		}
	}
	return uniqueNonEmpty(ids)
}

func notionMarkdownReferences(sourceID string, markdown string) []companyknowledge.NotionOutboundReference {
	pages, databases, dataSources := notionMarkdownLinkedObjects(markdown)
	refs := make([]companyknowledge.NotionOutboundReference, 0, len(pages)+len(databases)+len(dataSources))
	for _, id := range pages {
		refs = append(refs, companyknowledge.NotionOutboundReference{
			ReferenceKind:    "page",
			SourceBlockID:    normalizeNotionID(sourceID),
			TargetID:         id,
			TargetObjectKind: companyknowledge.NotionObjectKindPage,
			Traversed:        true,
			Reason:           "notion markdown page reference traversed",
		})
	}
	for _, id := range databases {
		refs = append(refs, companyknowledge.NotionOutboundReference{
			ReferenceKind:    "database",
			SourceBlockID:    normalizeNotionID(sourceID),
			TargetID:         id,
			TargetObjectKind: companyknowledge.NotionObjectKindDatabase,
			Traversed:        true,
			Reason:           "notion markdown database reference traversed",
		})
	}
	for _, id := range dataSources {
		refs = append(refs, companyknowledge.NotionOutboundReference{
			ReferenceKind:    "data_source",
			SourceBlockID:    normalizeNotionID(sourceID),
			TargetID:         id,
			TargetObjectKind: companyknowledge.NotionObjectKindDataSource,
			Traversed:        true,
			Reason:           "notion markdown data source reference traversed",
		})
	}
	return refs
}

func notionTimestampWithLookback(raw string, lookback time.Duration) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || lookback <= 0 {
		return raw
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return raw
	}
	return parsed.Add(-lookback).UTC().Format(time.RFC3339Nano)
}

func maxNotionTimestamp(left string, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if right == "" {
		return left
	}
	if left == "" {
		return right
	}
	leftTime, leftErr := time.Parse(time.RFC3339Nano, left)
	rightTime, rightErr := time.Parse(time.RFC3339Nano, right)
	if leftErr != nil || rightErr != nil {
		if right > left {
			return right
		}
		return left
	}
	if rightTime.After(leftTime) {
		return right
	}
	return left
}

func stringSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		if value = normalizeNotionID(value); value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}

func sortedSetKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func notionBlockMarkdown(block clients.NotionBlock, depth int) string {
	payload, ok := block.Raw[block.Type].(map[string]any)
	if !ok {
		return ""
	}
	text := richTextPlainTextFromAny(payload["rich_text"])
	indent := strings.Repeat("  ", depth)
	switch block.Type {
	case "paragraph":
		return indent + text
	case "heading_1":
		return "# " + text
	case "heading_2":
		return "## " + text
	case "heading_3":
		return "### " + text
	case "bulleted_list_item":
		return indent + "- " + text
	case "numbered_list_item":
		return indent + "1. " + text
	case "to_do":
		checked := " "
		if value, _ := payload["checked"].(bool); value {
			checked = "x"
		}
		return fmt.Sprintf("%s- [%s] %s", indent, checked, text)
	case "quote":
		return indent + "> " + text
	case "callout":
		return indent + "Callout: " + text
	case "code":
		language := strings.TrimSpace(fmt.Sprint(payload["language"]))
		if language == "<nil>" {
			language = ""
		}
		return fmt.Sprintf("%s```%s\n%s\n%s```", indent, language, text, indent)
	case "toggle":
		return indent + "Toggle: " + text
	case "child_page":
		title := strings.TrimSpace(fmt.Sprint(payload["title"]))
		if title == "" || title == "<nil>" {
			title = block.ID
		}
		return indent + "Child page: " + title
	case "child_database":
		title := strings.TrimSpace(fmt.Sprint(payload["title"]))
		if title == "" || title == "<nil>" {
			title = block.ID
		}
		return indent + "Child database: " + title
	default:
		if text != "" {
			return indent + text
		}
		return ""
	}
}

func notionBlockReferences(block clients.NotionBlock) []companyknowledge.NotionOutboundReference {
	payload, _ := block.Raw[block.Type].(map[string]any)
	refs := []companyknowledge.NotionOutboundReference{}
	for _, ref := range notionRichTextReferences(block.ID, payload["rich_text"]) {
		refs = append(refs, ref)
	}
	for _, key := range []string{"url", "link"} {
		if value := strings.TrimSpace(fmt.Sprint(payload[key])); value != "" && value != "<nil>" {
			refs = append(refs, companyknowledge.NotionOutboundReference{
				ReferenceKind: "url",
				SourceBlockID: strings.TrimSpace(block.ID),
				TargetURL:     value,
				Traversed:     false,
				Reason:        "notion outbound URL recorded but not traversed",
			})
		}
	}
	switch block.Type {
	case "embed", "bookmark", "link_preview", "file", "pdf", "image", "video":
		if len(refs) == 0 {
			refs = append(refs, companyknowledge.NotionOutboundReference{
				ReferenceKind: "unsupported_embed",
				SourceBlockID: strings.TrimSpace(block.ID),
				Traversed:     false,
				Reason:        "notion block type recorded but not extracted in this mirror tranche",
			})
		}
	}
	return refs
}

func notionRichTextReferences(blockID string, value any) []companyknowledge.NotionOutboundReference {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	refs := []companyknowledge.NotionOutboundReference{}
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if href := strings.TrimSpace(fmt.Sprint(object["href"])); href != "" && href != "<nil>" {
			refs = append(refs, companyknowledge.NotionOutboundReference{
				ReferenceKind: "href",
				SourceBlockID: strings.TrimSpace(blockID),
				TargetURL:     href,
				Traversed:     false,
				Reason:        "notion rich-text href recorded but not traversed",
			})
		}
		mention, ok := object["mention"].(map[string]any)
		if !ok {
			continue
		}
		mentionType := strings.TrimSpace(fmt.Sprint(mention["type"]))
		targetID := ""
		if payload, ok := mention[mentionType].(map[string]any); ok {
			for _, key := range []string{"id", "page_id", "database_id", "user_id"} {
				if value := strings.TrimSpace(fmt.Sprint(payload[key])); value != "" && value != "<nil>" {
					targetID = normalizeNotionID(value)
					break
				}
			}
		}
		if targetID != "" {
			refs = append(refs, companyknowledge.NotionOutboundReference{
				ReferenceKind:    "mention",
				SourceBlockID:    strings.TrimSpace(blockID),
				TargetID:         targetID,
				TargetObjectKind: mentionType,
				Traversed:        false,
				Reason:           "notion mention recorded as provenance",
			})
		}
	}
	return refs
}

func richTextPlainText(items []clients.NotionText) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.PlainText) != "" {
			parts = append(parts, item.PlainText)
		}
	}
	return strings.TrimSpace(strings.Join(parts, ""))
}

func richTextPlainTextFromAny(value any) string {
	items, ok := value.([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if text := strings.TrimSpace(fmt.Sprint(object["plain_text"])); text != "" && text != "<nil>" {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, ""))
}
