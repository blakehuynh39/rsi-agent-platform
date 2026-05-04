package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	BlockPaginationComplete bool      `json:"block_pagination_complete"`
	Truncated               bool      `json:"truncated"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type notionPageExtraction struct {
	Body                    string
	LinkedPages             []string
	LinkedDatabases         []string
	OutboundReferences      []companyknowledge.NotionOutboundReference
	BlockPaginationComplete bool
	Truncated               bool
}

type notionMirrorRunner struct {
	cfg        config.Config
	api        notionAPI
	mirror     *companyknowledge.NotionMirror
	store      store.SourceMirrorWriteStore
	workspace  string
	seenPages  map[string]struct{}
	seenDBs    map[string]struct{}
	maxBlocks  int
	maxDepth   int
	maxDBs     int
	pageCount  int
	dbCount    int
	truncated  bool
	checkpoint notionMirrorCheckpoint
}

type notionAPI interface {
	RetrievePage(ctx context.Context, pageID string) (clients.NotionPage, error)
	RetrieveDatabase(ctx context.Context, databaseID string) (clients.NotionDatabase, error)
	ListBlockChildren(ctx context.Context, blockID string, cursor string, pageSize int) (clients.NotionListResponse[clients.NotionBlock], error)
	QueryDatabase(ctx context.Context, databaseID string, cursor string, pageSize int) (clients.NotionListResponse[clients.NotionPage], error)
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
		cfg:        cfg,
		api:        api,
		mirror:     mirror,
		store:      mirrorStore,
		workspace:  workspace,
		seenPages:  map[string]struct{}{},
		seenDBs:    map[string]struct{}{},
		maxBlocks:  maxBlocks,
		maxDepth:   maxDepth,
		maxDBs:     maxDBs,
		checkpoint: checkpoint,
	}, nil
}

func (r *notionMirrorRunner) mirrorRoot(ctx context.Context, rootID string) error {
	rootID = normalizeNotionID(rootID)
	r.checkpoint.RootID = rootID
	r.checkpoint.WorkspaceID = r.workspace
	log.Printf("notion mirror root=%s starting", rootID)
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
	return fmt.Errorf("notion allowlist root %s is neither a visible page nor a visible database", rootID)
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
	if page.Archived || page.InTrash {
		return r.markNotionObjectStale(pageID, companyknowledge.NotionObjectKindPage, rootID, "notion page is archived or in trash", map[string]any{
			"archived": page.Archived,
			"in_trash": page.InTrash,
		})
	}
	title := notionPageTitle(page)
	if title == "" {
		title = pageID
	}
	currentHierarchy := append(append([]string{}, hierarchy...), title)
	seenBlocks := 0
	extraction, err := r.extractPageBody(ctx, pageID, 0, &seenBlocks)
	if err != nil {
		return fmt.Errorf("extract notion page body page=%s: %w", pageID, err)
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
	input := companyknowledge.NotionDocumentInput{
		WorkspaceID:        r.workspace,
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
		Hierarchy:          currentHierarchy,
		OutboundReferences: outboundReferences,
		Raw:                map[string]any{"object": page.Object},
	}
	revision := companyknowledge.NotionDocumentSourceRevision(input)
	sourceKey := companyknowledge.NotionObjectSourceKey(input.WorkspaceID, input.ObjectKind, input.ObjectID)
	result, err := r.mirror.IngestDocument(ctx, input)
	if err != nil {
		return fmt.Errorf("mirror notion page=%s into honcho: %w", pageID, err)
	}
	if shouldPublishNotionWikiSource(result) {
		if _, err := companyknowledge.RecordAndPublishWikiSource(ctx, r.cfg, r.store, companyknowledge.NotionWikiSourceRevisionInput(input)); err != nil {
			return fmt.Errorf("publish notion wiki source page=%s: %w", pageID, err)
		}
	}
	r.checkpoint.CompletedPages[sourceKey] = revision
	r.checkpoint.CompletedObjects[sourceKey] = notionMirrorCheckpointObject{
		SourceRevision:          revision,
		ChildPageIDs:            uniqueNonEmpty(extraction.LinkedPages),
		ChildDatabaseIDs:        uniqueNonEmpty(extraction.LinkedDatabases),
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
	for _, childDatabaseID := range extraction.LinkedDatabases {
		if err := r.mirrorDatabase(ctx, childDatabaseID, rootID, currentHierarchy); err != nil {
			return err
		}
	}
	for _, childPageID := range extraction.LinkedPages {
		if err := r.mirrorPage(ctx, childPageID, rootID, currentHierarchy); err != nil {
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
			isRoot := databaseID == normalizeNotionID(rootID) && hierarchy == nil
			if isRoot {
				return err
			}
			return r.recordCrawlMiss(ctx, rootID, databaseID, "database", "notion database was not reachable")
		}
		return err
	}
	if database.Archived || database.InTrash {
		return r.markNotionObjectStale(databaseID, companyknowledge.NotionObjectKindDatabase, rootID, "notion database is archived or in trash", map[string]any{
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
		Content:         "Row pages in this database are mirrored as separate Notion documents.",
		SchemaSummary:   schemaSummary,
		SchemaHash:      schemaHash,
		TraversalStatus: companyknowledge.NotionTraversalComplete,
		Hierarchy:       currentHierarchy,
		Raw:             map[string]any{"object": database.Object},
	}
	databaseRevision := companyknowledge.NotionDocumentSourceRevision(databaseInput)
	databaseSourceKey := companyknowledge.NotionObjectSourceKey(databaseInput.WorkspaceID, databaseInput.ObjectKind, databaseInput.ObjectID)
	databaseResult, err := r.mirror.IngestDocument(ctx, databaseInput)
	if err != nil {
		return fmt.Errorf("mirror notion database=%s into honcho: %w", databaseID, err)
	}
	r.checkpoint.CompletedObjects[databaseSourceKey] = notionMirrorCheckpointObject{
		SourceRevision:          databaseRevision,
		BlockPaginationComplete: true,
		UpdatedAt:               time.Now().UTC(),
	}
	r.checkpoint.LastProgressAt = time.Now().UTC()
	if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
		return err
	}
	log.Printf("notion mirror database=%s status=%s skipped=%t reason=%s", databaseID, databaseResult.Status, databaseResult.Skipped, databaseResult.SkipReason)
	cursor := ""
	for {
		page, err := r.api.QueryDatabase(ctx, databaseID, cursor, 100)
		if err != nil {
			return err
		}
		for _, result := range page.Results {
			if strings.TrimSpace(result.ID) == "" {
				continue
			}
			if err := r.mirrorPage(ctx, result.ID, rootID, currentHierarchy); err != nil {
				return err
			}
		}
		cursor = strings.TrimSpace(page.NextCursor)
		if !page.HasMore || cursor == "" {
			return nil
		}
	}
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

func normalizeNotionID(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "-", "")
}

func isNotionNotFound(err error) bool {
	var apiErr clients.NotionAPIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

func isNotionPageEndpointTypeMismatch(err error) bool {
	return isNotionTypeMismatch(err, "database", "page")
}

func isNotionDatabaseEndpointTypeMismatch(err error) bool {
	return isNotionTypeMismatch(err, "page", "database")
}

func isNotionTypeMismatch(err error, actual string, requested string) bool {
	var apiErr clients.NotionAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
		return false
	}
	body := strings.ToLower(apiErr.Body)
	return strings.Contains(body, "validation_error") &&
		strings.Contains(body, "is a "+actual+", not a "+requested)
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
