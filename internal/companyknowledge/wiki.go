package companyknowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type WikiPublishResult struct {
	Source  store.CompanyWikiSourceRevisionResult `json:"source"`
	Page    store.CompanyWikiPagePublishResult    `json:"page"`
	Audit   store.CompanyWikiAuditRecord          `json:"audit"`
	Skipped bool                                  `json:"skipped"`
	Reason  string                                `json:"reason,omitempty"`
}

type WikiManifestReconcileResult struct {
	OK       bool                             `json:"ok"`
	Checked  int                              `json:"checked"`
	Warnings []WikiManifestReconcileWarning   `json:"warnings,omitempty"`
	Repaired []WikiManifestReconcileWarning   `json:"repaired,omitempty"`
	Manifest []store.CompanyWikiManifestEntry `json:"manifest,omitempty"`
}

type WikiManifestReconcileWarning struct {
	Path           string `json:"path"`
	WikiPageID     string `json:"wiki_page_id"`
	WikiRevisionID string `json:"wiki_revision_id"`
	ExpectedSHA256 string `json:"expected_sha256"`
	ActualSHA256   string `json:"actual_sha256,omitempty"`
	Reason         string `json:"reason"`
}

var manifestFileMutex sync.Mutex

const (
	CompanyWikiCompilerVersion        = "compiler.v1"
	CompanyWikiSchemaVersion          = "schema.v1"
	CompanyWikiRendererVersion        = "renderer.v1"
	CompanyWikiModelPolicyVersion     = "model_policy.v1"
	CompanyWikiCompileMaxAttemptCount = 5
)

type WikiMarkdownRead struct {
	OK      bool   `json:"ok"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WikiLogEntry struct {
	Action           string
	Title            string
	Slug             string
	Status           string
	Actor            string
	Reason           string
	WikiRevisionID   string
	SourceDocumentID string
	SourceRevisionID string
	Summary          string
}

func RecordAndPublishWikiSource(ctx context.Context, cfg config.Config, repo any, input store.CompanyWikiSourceRevisionInput) (WikiPublishResult, error) {
	_ = ctx
	recorded, err := RecordWikiSourceRevision(ctx, cfg, repo, input)
	if err != nil || recorded.Skipped {
		return recorded, err
	}
	return PublishWikiSourceDocument(ctx, cfg, repo, recorded.Source)
}

func RecordEnqueueAndMaybePublishWikiSource(ctx context.Context, cfg config.Config, repo any, input store.CompanyWikiSourceRevisionInput) (WikiPublishResult, error) {
	recorded, err := RecordWikiSourceRevision(ctx, cfg, repo, input)
	if err != nil || recorded.Skipped {
		return recorded, err
	}
	if recorded.Source.Changed {
		if _, _, err := EnqueueWikiCompileItemForSource(ctx, cfg, repo, recorded.Source); err != nil {
			return WikiPublishResult{}, err
		}
	}
	if strings.EqualFold(strings.TrimSpace(cfg.CompanyWikiSourcePageMode), "off") {
		return recorded, nil
	}
	return PublishWikiSourceDocument(ctx, cfg, repo, recorded.Source)
}

func RecordWikiSourceRevision(ctx context.Context, cfg config.Config, repo any, input store.CompanyWikiSourceRevisionInput) (WikiPublishResult, error) {
	_ = ctx
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return WikiPublishResult{Skipped: true, Reason: "store_not_company_wiki_capable"}, nil
	}
	source, err := wikiStore.UpsertCompanyWikiSourceRevision(input)
	if err != nil {
		return WikiPublishResult{}, err
	}
	if !source.Changed {
		return WikiPublishResult{Source: source, Skipped: true, Reason: "revision_already_exists"}, nil
	}
	return WikiPublishResult{Source: source}, nil
}

func EnqueueWikiCompileItemForSource(ctx context.Context, cfg config.Config, repo any, source store.CompanyWikiSourceRevisionResult) (store.CompanyWikiCompileItem, bool, error) {
	_ = ctx
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok || strings.TrimSpace(source.Revision.ID) == "" {
		return store.CompanyWikiCompileItem{}, false, nil
	}
	chunks := source.Chunks
	if len(chunks) == 0 {
		var err error
		evidence, found, err := wikiStore.GetCompanyWikiSourceEvidence(source.Revision.ID)
		if err != nil {
			return store.CompanyWikiCompileItem{}, false, err
		}
		if !found {
			return store.CompanyWikiCompileItem{}, false, nil
		}
		chunks = evidence.Chunks
	}
	return wikiStore.EnqueueCompanyWikiCompileItem(store.CompanyWikiCompileItemInput{
		SourceRevisionID:   source.Revision.ID,
		CompilerVersion:    CompanyWikiCompilerVersion,
		SchemaVersion:      CompanyWikiSchemaVersion,
		RendererVersion:    CompanyWikiRendererVersion,
		ModelPolicyVersion: CompanyWikiModelPolicyVersion,
		InputHash:          SourceCentricCompileInputHash(source.Revision, chunks),
		Status:             store.CompanyWikiCompileStatusPending,
	})
}

func BackfillCompanyWikiCompileItems(ctx context.Context, cfg config.Config, wikiStore store.CompanyWikiStore, limit int) (int, error) {
	if wikiStore == nil {
		return 0, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	revisionIDs, err := wikiStore.ListCompanyWikiSourceRevisionIDsWithoutCompileItem(
		CompanyWikiCompilerVersion,
		CompanyWikiSchemaVersion,
		CompanyWikiRendererVersion,
		CompanyWikiModelPolicyVersion,
		limit,
	)
	if err != nil {
		return 0, err
	}
	enqueued := 0
	for _, revisionID := range revisionIDs {
		select {
		case <-ctx.Done():
			return enqueued, ctx.Err()
		default:
		}
		evidence, found, err := wikiStore.GetCompanyWikiSourceEvidence(revisionID)
		if err != nil {
			return enqueued, err
		}
		if !found {
			continue
		}
		_, inserted, err := EnqueueWikiCompileItemForSource(ctx, cfg, wikiStore, store.CompanyWikiSourceRevisionResult{
			Document: evidence.Document,
			Revision: evidence.Revision,
			Chunks:   evidence.Chunks,
			Inserted: true,
			Changed:  true,
		})
		if err != nil {
			return enqueued, err
		}
		if inserted {
			enqueued++
		}
	}
	return enqueued, nil
}

func SourceCentricCompileInputHash(revision store.CompanyWikiSourceRevision, chunks []store.CompanyWikiSourceChunk) string {
	parts := []string{
		revision.ID,
		revision.ContentSHA256,
		CompanyWikiCompilerVersion,
		CompanyWikiSchemaVersion,
		CompanyWikiRendererVersion,
		CompanyWikiModelPolicyVersion,
	}
	chunksCopy := make([]store.CompanyWikiSourceChunk, len(chunks))
	copy(chunksCopy, chunks)
	sort.SliceStable(chunksCopy, func(i, j int) bool { return chunksCopy[i].ID < chunksCopy[j].ID })
	for _, chunk := range chunksCopy {
		parts = append(parts, chunk.ID, chunk.ContentSHA256)
	}
	return store.CompanyWikiSHA256(strings.Join(parts, "\x00"))
}

func PublishWikiSourceDocument(ctx context.Context, cfg config.Config, repo any, source store.CompanyWikiSourceRevisionResult) (WikiPublishResult, error) {
	_ = ctx
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return WikiPublishResult{Source: source, Skipped: true, Reason: "store_not_company_wiki_capable"}, nil
	}
	if strings.TrimSpace(cfg.CompanyWikiRoot) == "" {
		return WikiPublishResult{Source: source, Skipped: true, Reason: "company_wiki_root_not_configured"}, nil
	}
	if strings.EqualFold(strings.TrimSpace(cfg.CompanyWikiSourcePageMode), "off") {
		return WikiPublishResult{Source: source, Skipped: true, Reason: "source_page_mode_off"}, nil
	}
	if strings.TrimSpace(source.Document.ID) == "" {
		return WikiPublishResult{Source: source, Skipped: true, Reason: "source_document_not_recorded"}, nil
	}
	chunks, err := wikiStore.ListCompanyWikiSourceChunks(source.Document.ID)
	if err != nil {
		return WikiPublishResult{}, err
	}
	if len(chunks) == 0 {
		return WikiPublishResult{Source: source, Skipped: true, Reason: "no_chunks"}, nil
	}
	body, citations := BuildCompiledWikiMarkdown(source.Document, chunks)
	slug := WikiSlugForSource(source.Document)
	relativePath := EvidenceWikiPathForSource(source.Document, slug)
	audit, err := wikiStore.BeginCompanyWikiAudit(store.CompanyWikiAuditInput{
		Mode:           store.CompanyWikiAuditModeCompiler,
		Actor:          "company_wiki_compiler",
		Reason:         "source_revision_compiled",
		IdempotencyKey: source.Revision.ID + ":" + store.CompanyWikiSHA256(body),
		Slug:           slug,
		Title:          source.Document.Title,
		ProposedPath:   relativePath,
		Metadata: map[string]any{
			"source_document_id": source.Document.ID,
			"source_revision_id": source.Revision.ID,
		},
	})
	if err != nil {
		return WikiPublishResult{}, err
	}
	sha, err := PublishMarkdownFile(cfg.CompanyWikiRoot, relativePath, body)
	if err != nil {
		failed, _ := wikiStore.FailCompanyWikiAudit(audit.ID, err.Error(), map[string]any{"stage": "publish_markdown"})
		return WikiPublishResult{Source: source, Audit: failed}, err
	}
	sourceRevisionIDs := uniqueRevisionIDsFromChunks(chunks)
	page, err := wikiStore.PublishCompanyWikiPage(store.CompanyWikiPagePublishInput{
		AuditID:           audit.ID,
		Slug:              slug,
		Title:             source.Document.Title,
		Body:              body,
		Path:              relativePath,
		SHA256:            sha,
		CompilerRunID:     "compiler_" + time.Now().UTC().Format("20060102T150405Z"),
		SourceRevisionIDs: sourceRevisionIDs,
		Citations:         citations,
		Metadata: map[string]any{
			"type":               "evidence",
			"source_type":        source.Document.SourceType,
			"source_key":         source.Document.SourceKey,
			"source_session_key": source.Document.SourceSessionKey,
		},
		PublishedAt: time.Now().UTC(),
	})
	if err != nil {
		failed, _ := wikiStore.FailCompanyWikiAudit(audit.ID, err.Error(), map[string]any{"stage": "record_revision"})
		return WikiPublishResult{Source: source, Audit: failed}, err
	}
	completed, err := wikiStore.CompleteCompanyWikiAudit(audit.ID, page.Revision.ID, relativePath, map[string]any{
		"sha256":           sha,
		"wiki_revision_id": page.Revision.ID,
	})
	if err != nil {
		return WikiPublishResult{}, err
	}
	if err := WriteManifestFile(cfg.CompanyWikiRoot, page.Page.Slug, page.Revision.ID, relativePath, sha, page.Revision.CompilerRunID, page.Revision.PublishedAt); err != nil {
		return WikiPublishResult{}, err
	}
	if err := WriteIndexFile(cfg.CompanyWikiRoot, wikiStore); err != nil {
		return WikiPublishResult{}, err
	}
	if err := AppendLogEntry(cfg.CompanyWikiRoot, WikiLogEntry{
		Action:           "ingest",
		Title:            source.Document.Title,
		Slug:             page.Page.Slug,
		Status:           "published",
		WikiRevisionID:   page.Revision.ID,
		SourceDocumentID: source.Document.ID,
		SourceRevisionID: source.Revision.ID,
		Summary:          wikiOneLineSummary(page.Revision.Body),
	}); err != nil {
		return WikiPublishResult{}, err
	}
	return WikiPublishResult{Source: source, Page: page, Audit: completed}, nil
}

func BuildCompiledWikiMarkdown(document store.CompanyWikiSourceDocument, chunks []store.CompanyWikiSourceChunk) (string, []store.CompanyWikiCitationInput) {
	sort.SliceStable(chunks, func(i, j int) bool {
		if chunks[i].RevisionID == chunks[j].RevisionID {
			return chunks[i].ChunkIndex < chunks[j].ChunkIndex
		}
		if chunks[i].CreatedAt.Equal(chunks[j].CreatedAt) {
			return chunks[i].ID < chunks[j].ID
		}
		return chunks[i].CreatedAt.Before(chunks[j].CreatedAt)
	})
	title := strings.TrimSpace(document.Title)
	if title == "" {
		title = document.SourceKey
	}
	citations := make([]store.CompanyWikiCitationInput, 0, len(chunks))
	sourceRevisionIDs := uniqueRevisionIDsFromChunks(chunks)
	var b strings.Builder
	b.WriteString("---\n")
	writeYAMLScalar(&b, "title", title)
	writeYAMLScalar(&b, "type", "evidence")
	writeYAMLScalar(&b, "wiki_page_source", document.SourceType)
	writeYAMLScalar(&b, "source_document_id", document.ID)
	writeYAMLScalar(&b, "source_key", document.SourceKey)
	writeYAMLScalar(&b, "source_session_key", document.SourceSessionKey)
	b.WriteString("source_revision_ids:\n")
	for _, revisionID := range sourceRevisionIDs {
		b.WriteString("  - ")
		b.WriteString(yamlQuote(revisionID))
		b.WriteString("\n")
	}
	b.WriteString("conflicts: []\n")
	b.WriteString("---\n\n")
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")
	if strings.TrimSpace(document.URL) != "" {
		b.WriteString("Source: ")
		b.WriteString(document.URL)
		b.WriteString("\n\n")
	}
	b.WriteString("## Compiled Evidence\n\n")
	for _, chunk := range chunks {
		citation := store.CompanyWikiCitationInput{
			ClaimKey:         "source:" + document.ID,
			SourceDocumentID: chunk.DocumentID,
			SourceRevisionID: chunk.RevisionID,
			ChunkID:          chunk.ID,
			NativeLocator:    chunk.NativeLocator,
			Quote:            truncateForCitation(chunk.Content, 280),
		}
		citations = append(citations, citation)
		b.WriteString("### Citation ")
		b.WriteString(fmt.Sprintf("%d", len(citations)))
		b.WriteString("\n\n")
		b.WriteString("- `source_document_id`: `")
		b.WriteString(citation.SourceDocumentID)
		b.WriteString("`\n")
		b.WriteString("- `source_revision_id`: `")
		b.WriteString(citation.SourceRevisionID)
		b.WriteString("`\n")
		b.WriteString("- `chunk_id`: `")
		b.WriteString(citation.ChunkID)
		b.WriteString("`\n")
		if citation.NativeLocator != "" {
			b.WriteString("- `native_locator`: `")
			b.WriteString(citation.NativeLocator)
			b.WriteString("`\n")
		}
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(chunk.Content))
		b.WriteString("\n\n")
	}
	return b.String(), citations
}

func PublishMarkdownFile(root string, relativePath string, body string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("company wiki root is required")
	}
	relativePath = cleanRelativeWikiPath(relativePath)
	if relativePath == "" {
		return "", errors.New("wiki relative path is required")
	}
	target := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	stageDir := filepath.Join(root, ".staging")
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return "", err
	}
	tmp := filepath.Join(stageDir, filepath.Base(relativePath)+fmt.Sprintf(".%d.tmp", time.Now().UnixNano()))
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		return "", err
	}
	if err := fsyncFile(tmp); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	_ = fsyncDir(filepath.Dir(target))
	return store.CompanyWikiSHA256(body), nil
}

func WriteManifestFile(root string, slug string, revisionID string, relativePath string, sha string, compilerRunID string, generatedAt time.Time) error {
	if strings.TrimSpace(root) == "" {
		return nil
	}
	return withManifestFileLock(root, func() error {
		manifestPath := filepath.Join(root, "manifest.json")
		existing := map[string]any{}
		if raw, err := os.ReadFile(manifestPath); err == nil {
			_ = json.Unmarshal(raw, &existing)
		}
		pages, _ := existing["pages"].(map[string]any)
		if pages == nil {
			pages = map[string]any{}
		}
		pages[slug] = map[string]any{
			"path":             relativePath,
			"wiki_revision_id": revisionID,
			"sha256":           sha,
			"compiler_run_id":  compilerRunID,
			"generated_at":     generatedAt.Format(time.RFC3339),
		}
		existing["schema_version"] = 1
		existing["pages"] = pages
		raw, err := json.MarshalIndent(existing, "", "  ")
		if err != nil {
			return err
		}
		raw = append(raw, '\n')
		if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
			return err
		}
		tmp := manifestPath + fmt.Sprintf(".%d.tmp", time.Now().UnixNano())
		if err := os.WriteFile(tmp, raw, 0o644); err != nil {
			return err
		}
		if err := fsyncFile(tmp); err != nil {
			_ = os.Remove(tmp)
			return err
		}
		if err := os.Rename(tmp, manifestPath); err != nil {
			_ = os.Remove(tmp)
			return err
		}
		return fsyncDir(filepath.Dir(manifestPath))
	})
}

func ReconcileWikiManifest(ctx context.Context, cfg config.Config, repo any, repair bool) (WikiManifestReconcileResult, error) {
	_ = ctx
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return WikiManifestReconcileResult{}, errors.New("configured store does not support company wiki")
	}
	root := strings.TrimSpace(cfg.CompanyWikiRoot)
	if root == "" {
		return WikiManifestReconcileResult{}, errors.New("company wiki root is required")
	}
	entries, err := wikiStore.ListCompanyWikiManifestEntries()
	if err != nil {
		return WikiManifestReconcileResult{}, err
	}
	if repair {
		if err := WriteSchemaFile(root); err != nil {
			return WikiManifestReconcileResult{}, err
		}
		if err := WriteIndexFile(root, wikiStore); err != nil {
			return WikiManifestReconcileResult{}, err
		}
	}
	result := WikiManifestReconcileResult{OK: true, Checked: len(entries), Manifest: entries}
	for _, entry := range entries {
		warning := reconcileManifestEntry(root, entry)
		if warning.Reason == "" {
			_ = wikiStore.UpdateCompanyWikiManifestRepair(entry.Path, store.CompanyWikiManifestRepairOK, "")
			continue
		}
		result.OK = false
		_ = wikiStore.UpdateCompanyWikiManifestRepair(entry.Path, store.CompanyWikiManifestRepairNeeded, warning.Reason)
		_ = AppendLogEntry(root, WikiLogEntry{Action: "repair_needed", WikiRevisionID: entry.WikiRevisionID, Summary: warning.Reason})
		if repair {
			page, found, err := wikiStore.GetCompanyWikiPage(entry.WikiPageID)
			if err == nil && found && page.Revision.ID == entry.WikiRevisionID {
				if sha, publishErr := PublishMarkdownFile(root, entry.Path, page.Revision.Body); publishErr == nil && sha == entry.SHA256 {
					_ = WriteManifestFile(root, page.Page.Slug, page.Revision.ID, page.Revision.Path, sha, page.Revision.CompilerRunID, page.Revision.PublishedAt)
					_ = wikiStore.UpdateCompanyWikiManifestRepair(entry.Path, store.CompanyWikiManifestRepairOK, "")
					_ = AppendLogEntry(root, WikiLogEntry{Action: "repair_completed", Title: page.Page.Title, Slug: page.Page.Slug, WikiRevisionID: page.Revision.ID, Summary: warning.Reason})
					result.Repaired = append(result.Repaired, warning)
					continue
				} else if publishErr != nil {
					warning.Reason += ": " + publishErr.Error()
				}
			}
			if err != nil {
				warning.Reason += ": " + err.Error()
			}
			_ = wikiStore.UpdateCompanyWikiManifestRepair(entry.Path, store.CompanyWikiManifestRepairFailed, warning.Reason)
			_ = AppendLogEntry(root, WikiLogEntry{Action: "repair_failed", WikiRevisionID: entry.WikiRevisionID, Summary: warning.Reason})
		}
		result.Warnings = append(result.Warnings, warning)
	}
	if len(result.Warnings) == 0 {
		result.OK = true
	}
	return result, nil
}

func WriteIndexFile(root string, wikiStore store.CompanyWikiStore) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	body, err := BuildIndexMarkdown(wikiStore)
	if err != nil {
		return err
	}
	_, err = PublishMarkdownFile(root, "index.md", body)
	return err
}

func BuildIndexMarkdown(wikiStore store.CompanyWikiStore) (string, error) {
	entries, err := wikiStore.ListCompanyWikiManifestEntries()
	if err != nil {
		return "", err
	}
	type indexItem struct {
		category string
		title    string
		slug     string
		path     string
		summary  string
		meta     string
	}
	items := []indexItem{}
	for _, entry := range entries {
		page, found, err := wikiStore.GetCompanyWikiPage(entry.WikiPageID)
		if err != nil {
			return "", err
		}
		if !found {
			continue
		}
		category := strings.TrimSpace(stringFromMap(page.Revision.Metadata, "type"))
		if category == "evidence" {
			category = strings.TrimSpace(stringFromMap(page.Revision.Metadata, "source_type"))
		}
		if category == "" {
			category = "manual"
		}
		items = append(items, indexItem{
			category: category,
			title:    firstNonEmpty(page.Page.Title, page.Revision.Title, page.Page.Slug),
			slug:     page.Page.Slug,
			path:     page.Revision.Path,
			summary:  wikiOneLineSummary(page.Revision.Body),
			meta:     fmt.Sprintf("updated=%s; source revisions=%d; wiki_revision_id=%s", page.Revision.PublishedAt.Format(time.RFC3339), len(page.Revision.SourceRevisionIDs), page.Revision.ID),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if left, right := wikiIndexCategoryRank(items[i].category, items[i].path), wikiIndexCategoryRank(items[j].category, items[j].path); left != right {
			return left < right
		}
		if items[i].category == items[j].category {
			return items[i].slug < items[j].slug
		}
		return items[i].category < items[j].category
	})
	var b strings.Builder
	b.WriteString("# Company Wiki Index\n\n")
	b.WriteString("This content-oriented catalog is regenerated on every company wiki publish. Read it first, then drill into relevant pages.\n\n")
	current := ""
	for _, item := range items {
		if item.category != current {
			current = item.category
			b.WriteString("## ")
			b.WriteString(current)
			b.WriteString("\n\n")
		}
		b.WriteString("- [")
		b.WriteString(escapeMarkdownLinkText(item.title))
		b.WriteString("](")
		b.WriteString(filepath.ToSlash(item.path))
		b.WriteString(") — ")
		b.WriteString(item.summary)
		b.WriteString(" `")
		b.WriteString(item.meta)
		b.WriteString("`\n")
	}
	if len(items) == 0 {
		b.WriteString("_No published pages yet._\n")
	}
	return b.String(), nil
}

func wikiIndexCategoryRank(category string, path string) int {
	if strings.HasPrefix(filepath.ToSlash(path), "sources/") {
		return 100
	}
	category = strings.TrimSpace(category)
	if strings.EqualFold(category, "manual") {
		return 50
	}
	switch store.NormalizeCompanyWikiSlug(strings.ReplaceAll(category, "_", "-")) {
	case "project", "projects":
		return 0
	case "system", "systems":
		return 1
	case "decision", "decisions":
		return 2
	case "runbook", "runbooks":
		return 3
	case "policy", "policies":
		return 4
	case "concept", "concepts":
		return 5
	case "person", "people":
		return 6
	case "open-question", "open-questions":
		return 7
	}
	return 60
}

func AppendLogEntry(root string, entry WikiLogEntry) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	action := strings.ToLower(strings.TrimSpace(entry.Action))
	if action == "" {
		action = "event"
	}
	title := strings.TrimSpace(entry.Title)
	if title == "" {
		title = strings.TrimSpace(entry.Slug)
	}
	if title == "" {
		title = "company wiki"
	}
	logPath := filepath.Join(root, "log.md")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	stat, statErr := file.Stat()
	var b strings.Builder
	if statErr == nil && stat.Size() == 0 {
		b.WriteString("# Company Wiki Log\n\n")
		b.WriteString("Append-only timeline. Entries intentionally start with `## [` so Unix tools can parse them, e.g. `grep '^## \\\\[' log.md | tail -5`.\n\n")
	}
	now := time.Now().UTC()
	b.WriteString("## [")
	b.WriteString(now.Format(time.RFC3339))
	b.WriteString("] ")
	b.WriteString(action)
	b.WriteString(" | ")
	b.WriteString(strings.ReplaceAll(title, "\n", " "))
	b.WriteString("\n\n")
	writeLogField(&b, "status", entry.Status)
	writeLogField(&b, "slug", entry.Slug)
	writeLogField(&b, "wiki_revision_id", entry.WikiRevisionID)
	writeLogField(&b, "source_document_id", entry.SourceDocumentID)
	writeLogField(&b, "source_revision_id", entry.SourceRevisionID)
	writeLogField(&b, "actor", entry.Actor)
	writeLogField(&b, "reason", entry.Reason)
	writeLogField(&b, "summary", entry.Summary)
	b.WriteString("\n")
	if _, err := file.WriteString(b.String()); err != nil {
		return err
	}
	return file.Sync()
}

func ReadIndexFile(root string) (WikiMarkdownRead, error) {
	return readWikiMarkdownFile(root, "index.md")
}

func ReadLogFile(root string, limit int) (WikiMarkdownRead, error) {
	read, err := readWikiMarkdownFile(root, "log.md")
	if err != nil || limit <= 0 {
		return read, err
	}
	read.Content = tailWikiLogEntries(read.Content, limit)
	return read, nil
}

func reconcileManifestEntry(root string, entry store.CompanyWikiManifestEntry) WikiManifestReconcileWarning {
	warning := WikiManifestReconcileWarning{
		Path:           entry.Path,
		WikiPageID:     entry.WikiPageID,
		WikiRevisionID: entry.WikiRevisionID,
		ExpectedSHA256: entry.SHA256,
	}
	raw, err := os.ReadFile(filepath.Join(root, cleanRelativeWikiPath(entry.Path)))
	if err != nil {
		warning.Reason = "published_file_missing_or_unreadable"
		return warning
	}
	actual := store.CompanyWikiSHA256(string(raw))
	if actual != entry.SHA256 {
		warning.ActualSHA256 = actual
		warning.Reason = "published_file_sha256_mismatch"
		return warning
	}
	return WikiManifestReconcileWarning{}
}

func WikiSlugForSource(document store.CompanyWikiSourceDocument) string {
	prefix := strings.TrimSpace(document.SourceType)
	title := strings.TrimSpace(document.Title)
	if title == "" {
		title = document.SourceKey
	}
	suffix := document.ID
	if len(document.ID) > 8 {
		suffix = document.ID[len(document.ID)-8:]
	}
	return store.NormalizeCompanyWikiSlug(filepath.ToSlash(filepath.Join(prefix, title+"-"+suffix)))
}

func EvidenceWikiPathForSource(document store.CompanyWikiSourceDocument, slug string) string {
	sourceType := strings.TrimSpace(document.SourceType)
	root := "sources"
	evidenceSlug := store.NormalizeCompanyWikiSlug(slug)
	sourcePrefix := store.NormalizeCompanyWikiSlug(sourceType)
	if sourcePrefix != "" && strings.HasPrefix(evidenceSlug, sourcePrefix+"/") {
		evidenceSlug = strings.TrimPrefix(evidenceSlug, sourcePrefix+"/")
	}
	switch sourceType {
	case SlackMessageSourceType:
		root = filepath.ToSlash(filepath.Join("sources", "slack"))
	case NotionDocumentSourceType:
		root = filepath.ToSlash(filepath.Join("sources", "notion"))
	default:
		root = filepath.ToSlash(filepath.Join("sources", store.NormalizeCompanyWikiSlug(sourceType)))
	}
	return filepath.ToSlash(filepath.Join(root, evidenceSlug+".md"))
}

func cleanRelativeWikiPath(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	path = strings.TrimPrefix(path, "/")
	path = filepath.Clean(path)
	for strings.HasPrefix(path, "../") {
		path = strings.TrimPrefix(path, "../")
	}
	if path == "." || path == ".." {
		return ""
	}
	return filepath.ToSlash(path)
}

func withManifestFileLock(root string, fn func() error) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return fn()
	}
	manifestFileMutex.Lock()
	defer manifestFileMutex.Unlock()

	lockDir := filepath.Join(root, ".locks")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return err
	}
	lockFile, err := os.OpenFile(filepath.Join(lockDir, "manifest.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer lockFile.Close()
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }()
	return fn()
}

func writeYAMLScalar(b *strings.Builder, key string, value string) {
	b.WriteString(key)
	b.WriteString(": ")
	b.WriteString(yamlQuote(value))
	b.WriteString("\n")
}

func yamlQuote(value string) string {
	raw, _ := json.Marshal(strings.TrimSpace(value))
	return string(raw)
}

func uniqueRevisionIDsFromChunks(chunks []store.CompanyWikiSourceChunk) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, chunk := range chunks {
		if chunk.RevisionID == "" {
			continue
		}
		if _, ok := seen[chunk.RevisionID]; ok {
			continue
		}
		seen[chunk.RevisionID] = struct{}{}
		out = append(out, chunk.RevisionID)
	}
	sort.Strings(out)
	return out
}

func truncateForCitation(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}

func wikiOneLineSummary(body string) string {
	body = stripYAMLFrontmatter(body)
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "- `") {
			continue
		}
		return truncateForCitation(line, 180)
	}
	return "No summary available."
}

func stripYAMLFrontmatter(body string) string {
	body = strings.TrimSpace(body)
	if !strings.HasPrefix(body, "---\n") && !strings.HasPrefix(body, "---\r\n") {
		return body
	}
	lines := strings.SplitN(body, "\n", 2)
	if len(lines) < 2 {
		return body
	}
	remaining := lines[1]
	closingIdx := strings.Index(remaining, "\n---\n")
	skipLen := 5
	if closingIdx == -1 {
		closingIdx = strings.Index(remaining, "\n---\r\n")
		skipLen = 6
	}
	if closingIdx == -1 && strings.HasSuffix(remaining, "\n---") {
		return ""
	}
	if closingIdx == -1 {
		return body
	}
	return strings.TrimSpace(remaining[closingIdx+skipLen:])
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if value, ok := values[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func escapeMarkdownLinkText(value string) string {
	value = strings.ReplaceAll(value, "[", "\\[")
	value = strings.ReplaceAll(value, "]", "\\]")
	return strings.ReplaceAll(value, "\n", " ")
}

func writeLogField(b *strings.Builder, key string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	b.WriteString("- ")
	b.WriteString(key)
	b.WriteString(": ")
	b.WriteString(value)
	b.WriteString("\n")
}

func readWikiMarkdownFile(root string, relativePath string) (WikiMarkdownRead, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return WikiMarkdownRead{}, errors.New("company wiki root is required")
	}
	relativePath = cleanRelativeWikiPath(relativePath)
	raw, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		return WikiMarkdownRead{}, err
	}
	return WikiMarkdownRead{OK: true, Path: relativePath, Content: string(raw)}, nil
}

func tailWikiLogEntries(content string, limit int) string {
	if limit <= 0 {
		return content
	}
	parts := strings.Split(content, "\n## [")
	if len(parts) <= 1 {
		return content
	}
	header := parts[0]
	entries := parts[1:]
	if len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	for _, entry := range entries {
		b.WriteString("\n## [")
		b.WriteString(entry)
	}
	return b.String()
}

func fsyncFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return file.Sync()
}

func fsyncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer dir.Close()
	return dir.Sync()
}
