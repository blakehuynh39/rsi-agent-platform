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
	RootID            string            `json:"root_id"`
	WorkspaceID       string            `json:"workspace_id"`
	CompletedPages    map[string]string `json:"completed_pages,omitempty"`
	LastPageID        string            `json:"last_page_id,omitempty"`
	LastProgressAt    time.Time         `json:"last_progress_at,omitempty"`
	LastCompletedAt   time.Time         `json:"last_completed_at,omitempty"`
	LastPageCount     int               `json:"last_page_count"`
	LastDatabaseCount int               `json:"last_database_count"`
}

type notionMirrorRunner struct {
	cfg        config.Config
	api        notionAPI
	mirror     *companyknowledge.NotionMirror
	workspace  string
	seenPages  map[string]struct{}
	seenDBs    map[string]struct{}
	maxBlocks  int
	pageCount  int
	dbCount    int
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
	api := clients.NewNotionClientWithOptions(cfg.NotionAPIBaseURL, cfg.NotionToken, cfg.NotionAPIVersion)
	mirror := companyknowledge.NewNotionMirror(mirrorStore, clients.NewHonchoClientWithAPIKey(cfg.HonchoBaseURL, cfg.HonchoAPIKey), companyknowledge.NotionMirrorOptions{
		Environment:     cfg.Environment,
		HonchoWorkspace: cfg.HonchoWorkspaceID,
	})
	for _, rootID := range roots {
		runner, err := newNotionMirrorRunner(cfg, api, mirror, rootID)
		if err != nil {
			return err
		}
		if err := runner.mirrorRoot(ctx, rootID); err != nil {
			return err
		}
		runner.checkpoint.LastCompletedAt = time.Now().UTC()
		runner.checkpoint.LastPageCount = runner.pageCount
		runner.checkpoint.LastDatabaseCount = runner.dbCount
		if err := writeNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, runner.checkpoint); err != nil {
			return err
		}
		log.Printf("notion mirror root=%s complete pages=%d databases=%d", rootID, runner.pageCount, runner.dbCount)
	}
	return nil
}

func newNotionMirrorRunner(cfg config.Config, api notionAPI, mirror *companyknowledge.NotionMirror, rootID string) (*notionMirrorRunner, error) {
	checkpoint, err := readNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, rootID)
	if err != nil {
		return nil, err
	}
	if checkpoint.CompletedPages == nil {
		checkpoint.CompletedPages = map[string]string{}
	}
	workspace := "notion"
	return &notionMirrorRunner{
		cfg:        cfg,
		api:        api,
		mirror:     mirror,
		workspace:  workspace,
		seenPages:  map[string]struct{}{},
		seenDBs:    map[string]struct{}{},
		maxBlocks:  1000,
		checkpoint: checkpoint,
	}, nil
}

func (r *notionMirrorRunner) mirrorRoot(ctx context.Context, rootID string) error {
	rootID = normalizeNotionID(rootID)
	r.checkpoint.RootID = rootID
	r.checkpoint.WorkspaceID = r.workspace
	log.Printf("notion mirror root=%s starting", rootID)
	_, pageErr := r.api.RetrievePage(ctx, rootID)
	if pageErr == nil {
		if err := r.mirrorPage(ctx, rootID, rootID, nil); err != nil {
			return fmt.Errorf("mirror notion page root=%s: %w", rootID, err)
		}
		return nil
	} else if !isNotionNotFound(pageErr) {
		return fmt.Errorf("retrieve notion page root=%s: %w", rootID, pageErr)
	}
	if err := r.mirrorDatabase(ctx, rootID, rootID, nil); err == nil {
		return nil
	} else if !isNotionNotFound(err) {
		return fmt.Errorf("mirror notion database root=%s: %w", rootID, err)
	}
	return fmt.Errorf("notion allowlist root %s is neither a visible page nor a visible database", rootID)
}

func (r *notionMirrorRunner) mirrorPage(ctx context.Context, pageID string, rootID string, hierarchy []string) error {
	pageID = normalizeNotionID(pageID)
	if pageID == "" {
		return nil
	}
	if _, seen := r.seenPages[pageID]; seen {
		return nil
	}
	r.seenPages[pageID] = struct{}{}
	page, err := r.api.RetrievePage(ctx, pageID)
	if err != nil {
		return err
	}
	if page.Archived || page.InTrash {
		return nil
	}
	title := notionPageTitle(page)
	if title == "" {
		title = pageID
	}
	currentHierarchy := append(append([]string{}, hierarchy...), title)
	seenBlocks := 0
	body, linkedPages, linkedDatabases, err := r.extractPageBody(ctx, pageID, 0, &seenBlocks)
	if err != nil {
		return fmt.Errorf("extract notion page body page=%s: %w", pageID, err)
	}
	input := companyknowledge.NotionDocumentInput{
		WorkspaceID:    r.workspace,
		PageID:         pageID,
		RootID:         normalizeNotionID(rootID),
		ParentID:       notionParentID(page.Parent),
		DatabaseID:     notionParentDatabaseID(page.Parent),
		Title:          title,
		URL:            strings.TrimSpace(page.URL),
		LastEditedTime: strings.TrimSpace(page.LastEditedTime),
		CreatedTime:    strings.TrimSpace(page.CreatedTime),
		Content:        body,
		Hierarchy:      currentHierarchy,
		Raw:            map[string]any{"object": page.Object},
	}
	revision := companyknowledge.NotionDocumentSourceRevision(input)
	sourceKey := companyknowledge.NotionDocumentSourceKey(input.WorkspaceID, input.PageID)
	if r.checkpoint.CompletedPages[sourceKey] == revision {
		log.Printf("notion mirror page=%s skipped unchanged", pageID)
	} else {
		result, err := r.mirror.IngestDocument(ctx, input)
		if err != nil {
			return fmt.Errorf("mirror notion page=%s into honcho: %w", pageID, err)
		}
		r.checkpoint.CompletedPages[sourceKey] = revision
		r.checkpoint.LastPageID = pageID
		r.checkpoint.LastProgressAt = time.Now().UTC()
		r.pageCount++
		if err := writeNotionMirrorCheckpoint(r.cfg.SourceMirrorCheckpointRoot, r.checkpoint); err != nil {
			return err
		}
		log.Printf("notion mirror page=%s status=%s skipped=%t reason=%s", pageID, result.Status, result.Skipped, result.SkipReason)
	}
	for _, childDatabaseID := range linkedDatabases {
		if err := r.mirrorDatabase(ctx, childDatabaseID, rootID, currentHierarchy); err != nil {
			return err
		}
	}
	for _, childPageID := range linkedPages {
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
		return err
	}
	if database.Archived || database.InTrash {
		return nil
	}
	r.dbCount++
	title := strings.TrimSpace(richTextPlainText(database.Title))
	if title == "" {
		title = databaseID
	}
	currentHierarchy := append(append([]string{}, hierarchy...), title)
	cursor := ""
	for {
		page, err := r.api.QueryDatabase(ctx, databaseID, cursor, 100)
		if err != nil {
			return err
		}
		for _, result := range page.Results {
			if strings.TrimSpace(result.ID) == "" || result.Archived || result.InTrash {
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

func (r *notionMirrorRunner) extractPageBody(ctx context.Context, pageID string, depth int, seenBlocks *int) (string, []string, []string, error) {
	var lines []string
	var linkedPages []string
	var linkedDatabases []string
	cursor := ""
	for {
		page, err := r.api.ListBlockChildren(ctx, pageID, cursor, 100)
		if err != nil {
			return "", nil, nil, err
		}
		for _, block := range page.Results {
			if block.Archived || block.InTrash {
				continue
			}
			*seenBlocks++
			if *seenBlocks > r.maxBlocks {
				lines = append(lines, "[Mirror truncated: Notion page exceeded block extraction limit.]")
				return strings.Join(lines, "\n"), linkedPages, linkedDatabases, nil
			}
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
				if block.HasChildren && depth < 4 {
					childText, childPages, childDatabases, err := r.extractPageBody(ctx, block.ID, depth+1, seenBlocks)
					if err != nil {
						return "", nil, nil, err
					}
					if childText != "" {
						lines = append(lines, childText)
					}
					linkedPages = append(linkedPages, childPages...)
					linkedDatabases = append(linkedDatabases, childDatabases...)
				}
			}
		}
		cursor = strings.TrimSpace(page.NextCursor)
		if !page.HasMore || cursor == "" {
			break
		}
	}
	return strings.Join(lines, "\n"), uniqueNonEmpty(linkedPages), uniqueNonEmpty(linkedDatabases), nil
}

func readNotionMirrorCheckpoint(root string, rootID string) (notionMirrorCheckpoint, error) {
	path := notionMirrorCheckpointPath(root, rootID)
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return notionMirrorCheckpoint{RootID: normalizeNotionID(rootID), CompletedPages: map[string]string{}}, nil
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
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func notionMirrorCheckpointPath(root string, rootID string) string {
	return filepath.Join(strings.TrimSpace(root), "notion", sanitizePathPart(normalizeNotionID(rootID))+".json")
}

func normalizeNotionID(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "-", "")
}

func isNotionNotFound(err error) bool {
	var apiErr clients.NotionAPIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
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
