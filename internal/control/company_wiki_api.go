package control

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
	"go.yaml.in/yaml/v3"
)

type companyWikiSearchResponse struct {
	OK      bool                            `json:"ok"`
	Query   string                          `json:"query"`
	Results []store.CompanyWikiSearchResult `json:"results"`
}

type companyWikiEditProposeRequest struct {
	Actor          string                           `json:"actor"`
	Reason         string                           `json:"reason"`
	IdempotencyKey string                           `json:"idempotency_key"`
	Slug           string                           `json:"slug"`
	Title          string                           `json:"title"`
	Body           string                           `json:"body,omitempty"`
	Citations      []store.CompanyWikiCitationInput `json:"citations,omitempty"`
	Metadata       map[string]any                   `json:"metadata,omitempty"`
}

type companyWikiEditApplyRequest struct {
	Actor          string                           `json:"actor"`
	Reason         string                           `json:"reason"`
	IdempotencyKey string                           `json:"idempotency_key"`
	Slug           string                           `json:"slug"`
	Title          string                           `json:"title"`
	Body           string                           `json:"body"`
	Citations      []store.CompanyWikiCitationInput `json:"citations"`
	Metadata       map[string]any                   `json:"metadata,omitempty"`
}

type companyWikiEditResponse struct {
	OK    bool                                `json:"ok"`
	Audit store.CompanyWikiAuditRecord        `json:"audit"`
	Page  *store.CompanyWikiPagePublishResult `json:"page,omitempty"`
}

var companyWikiSensitiveBodyPatterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{name: "private key", re: regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{name: "slack token", re: regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{8,}`)},
	{name: "secret assignment", re: regexp.MustCompile(`(?i)\b(?:api[_-]?key|token|secret|password|private[_-]?key)\s*[:=]\s*["']?[^\s"']{8,}`)},
	{name: "prompt fragment", re: regexp.MustCompile(`(?i)\b(?:system|developer|assistant)\s+prompt\b|</?(?:system|developer)>`)},
}

func companyWikiSearch(_ context.Context, repo any, query string, limit int) (companyWikiSearchResponse, int, error) {
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return companyWikiSearchResponse{}, http.StatusInternalServerError, errors.New("configured store does not support company wiki")
	}
	results, err := wikiStore.SearchCompanyWikiPages(query, limit)
	if err != nil {
		return companyWikiSearchResponse{}, http.StatusInternalServerError, err
	}
	return companyWikiSearchResponse{OK: true, Query: strings.TrimSpace(query), Results: results}, http.StatusOK, nil
}

func companyWikiPageGet(_ context.Context, repo any, ref string) (store.CompanyWikiPageRead, int, error) {
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return store.CompanyWikiPageRead{}, http.StatusInternalServerError, errors.New("configured store does not support company wiki")
	}
	if unescaped, err := url.PathUnescape(ref); err == nil {
		ref = unescaped
	}
	page, found, err := wikiStore.GetCompanyWikiPage(ref)
	if err != nil {
		return store.CompanyWikiPageRead{}, http.StatusInternalServerError, err
	}
	if !found {
		return store.CompanyWikiPageRead{}, http.StatusNotFound, errors.New("wiki page not found")
	}
	return page, http.StatusOK, nil
}

func companyWikiIndexGet(_ context.Context, cfg config.Config, repo any) (companyknowledge.WikiMarkdownRead, int, error) {
	read, err := companyknowledge.ReadIndexFile(cfg.CompanyWikiRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && companyWikiRootExists(cfg.CompanyWikiRoot) {
			empty, fallbackErr := emptyCompanyWikiMarkdownRead(repo, "index.md")
			if fallbackErr != nil {
				return companyknowledge.WikiMarkdownRead{}, http.StatusInternalServerError, fallbackErr
			}
			if empty.OK {
				return empty, http.StatusOK, nil
			}
		}
		return companyknowledge.WikiMarkdownRead{}, http.StatusNotFound, err
	}
	return read, http.StatusOK, nil
}

func companyWikiLogGet(_ context.Context, cfg config.Config, repo any, limit int) (companyknowledge.WikiMarkdownRead, int, error) {
	read, err := companyknowledge.ReadLogFile(cfg.CompanyWikiRoot, limit)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && companyWikiRootExists(cfg.CompanyWikiRoot) {
			empty, fallbackErr := emptyCompanyWikiMarkdownRead(repo, "log.md")
			if fallbackErr != nil {
				return companyknowledge.WikiMarkdownRead{}, http.StatusInternalServerError, fallbackErr
			}
			if empty.OK {
				return empty, http.StatusOK, nil
			}
		}
		return companyknowledge.WikiMarkdownRead{}, http.StatusNotFound, err
	}
	return read, http.StatusOK, nil
}

func companyWikiRootExists(root string) bool {
	info, err := os.Stat(strings.TrimSpace(root))
	return err == nil && info.IsDir()
}

func emptyCompanyWikiMarkdownRead(repo any, relativePath string) (companyknowledge.WikiMarkdownRead, error) {
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return companyknowledge.WikiMarkdownRead{}, errors.New("configured store does not support company wiki")
	}
	entries, err := wikiStore.ListCompanyWikiManifestEntries()
	if err != nil {
		return companyknowledge.WikiMarkdownRead{}, err
	}
	if len(entries) > 0 {
		return companyknowledge.WikiMarkdownRead{}, nil
	}
	switch relativePath {
	case "index.md":
		return companyknowledge.WikiMarkdownRead{
			OK:      true,
			Path:    "index.md",
			Content: "# Company Wiki Index\n\n_No published pages yet._\n",
		}, nil
	case "log.md":
		return companyknowledge.WikiMarkdownRead{
			OK:      true,
			Path:    "log.md",
			Content: "# Company Wiki Log\n\n_No wiki log entries yet._\n",
		}, nil
	default:
		return companyknowledge.WikiMarkdownRead{}, nil
	}
}

func companyWikiManifestReconcile(ctx context.Context, cfg config.Config, repo any, repair bool) (companyknowledge.WikiManifestReconcileResult, int, error) {
	result, err := companyknowledge.ReconcileWikiManifest(ctx, cfg, repo, repair)
	if err != nil {
		return companyknowledge.WikiManifestReconcileResult{}, http.StatusInternalServerError, err
	}
	if !result.OK {
		return result, http.StatusConflict, nil
	}
	return result, http.StatusOK, nil
}

func parseBoolQuery(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func companyWikiEditPropose(_ context.Context, cfg config.Config, repo any, req companyWikiEditProposeRequest) (companyWikiEditResponse, int, error) {
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return companyWikiEditResponse{}, http.StatusInternalServerError, errors.New("configured store does not support company wiki")
	}
	if strings.TrimSpace(req.Actor) == "" || strings.TrimSpace(req.Reason) == "" || strings.TrimSpace(req.IdempotencyKey) == "" {
		return companyWikiEditResponse{}, http.StatusBadRequest, errors.New("actor, reason, and idempotency_key are required")
	}
	audit, err := wikiStore.BeginCompanyWikiAudit(store.CompanyWikiAuditInput{
		Mode:           store.CompanyWikiAuditModePropose,
		Actor:          req.Actor,
		Reason:         req.Reason,
		IdempotencyKey: req.IdempotencyKey,
		Slug:           req.Slug,
		Title:          req.Title,
		Metadata: map[string]any{
			"body_sha256": store.CompanyWikiSHA256(req.Body),
			"citations":   len(req.Citations),
			"details":     req.Metadata,
		},
	})
	if err != nil {
		return companyWikiEditResponse{}, http.StatusInternalServerError, err
	}
	if err := companyknowledge.AppendLogEntry(cfg.CompanyWikiRoot, companyknowledge.WikiLogEntry{
		Action:  "edit_propose",
		Title:   req.Title,
		Slug:    req.Slug,
		Status:  "proposed",
		Actor:   req.Actor,
		Reason:  req.Reason,
		Summary: "Recorded audited wiki edit proposal.",
	}); err != nil {
		return companyWikiEditResponse{}, http.StatusInternalServerError, err
	}
	return companyWikiEditResponse{OK: true, Audit: audit}, http.StatusAccepted, nil
}

func companyWikiEditApply(_ context.Context, cfg config.Config, repo any, req companyWikiEditApplyRequest) (companyWikiEditResponse, int, error) {
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return companyWikiEditResponse{}, http.StatusInternalServerError, errors.New("configured store does not support company wiki")
	}
	if strings.TrimSpace(req.Actor) == "" || strings.TrimSpace(req.Reason) == "" || strings.TrimSpace(req.IdempotencyKey) == "" {
		return companyWikiEditResponse{}, http.StatusBadRequest, errors.New("actor, reason, and idempotency_key are required")
	}
	if strings.TrimSpace(req.Body) == "" {
		return companyWikiEditResponse{}, http.StatusBadRequest, errors.New("body is required")
	}
	if err := store.ValidateCompanyWikiCitationInputs(req.Citations); err != nil {
		return companyWikiEditResponse{}, http.StatusBadRequest, err
	}
	if err := validateCompanyWikiEditApplyBody(wikiStore, req.Body, req.Citations); err != nil {
		return companyWikiEditResponse{}, http.StatusBadRequest, err
	}
	slug := store.NormalizeCompanyWikiSlug(req.Slug)
	relativePath := "pages/" + slug + ".md"
	audit, err := wikiStore.BeginCompanyWikiAudit(store.CompanyWikiAuditInput{
		Mode:           store.CompanyWikiAuditModeApply,
		Actor:          req.Actor,
		Reason:         req.Reason,
		IdempotencyKey: req.IdempotencyKey,
		Slug:           slug,
		Title:          req.Title,
		ProposedPath:   relativePath,
		Metadata:       req.Metadata,
	})
	if err != nil {
		return companyWikiEditResponse{}, http.StatusInternalServerError, err
	}
	sha, err := companyknowledge.PublishMarkdownFile(cfg.CompanyWikiRoot, relativePath, req.Body)
	if err != nil {
		failed, _ := wikiStore.FailCompanyWikiAudit(audit.ID, err.Error(), map[string]any{"stage": "publish_markdown"})
		return companyWikiEditResponse{OK: false, Audit: failed}, http.StatusInternalServerError, err
	}
	page, err := wikiStore.PublishCompanyWikiPage(store.CompanyWikiPagePublishInput{
		AuditID:     audit.ID,
		Slug:        slug,
		Title:       req.Title,
		Body:        req.Body,
		Path:        relativePath,
		SHA256:      sha,
		Citations:   req.Citations,
		Metadata:    req.Metadata,
		PublishedAt: time.Now().UTC(),
	})
	if err != nil {
		failed, _ := wikiStore.FailCompanyWikiAudit(audit.ID, err.Error(), map[string]any{"stage": "record_revision"})
		return companyWikiEditResponse{OK: false, Audit: failed}, http.StatusInternalServerError, err
	}
	completed, err := wikiStore.CompleteCompanyWikiAudit(audit.ID, page.Revision.ID, relativePath, map[string]any{"sha256": sha})
	if err != nil {
		return companyWikiEditResponse{}, http.StatusInternalServerError, err
	}
	if err := companyknowledge.WriteManifestFile(cfg.CompanyWikiRoot, page.Page.Slug, page.Revision.ID, relativePath, sha, page.Revision.CompilerRunID, page.Revision.PublishedAt); err != nil {
		return companyWikiEditResponse{}, http.StatusInternalServerError, err
	}
	if err := companyknowledge.WriteIndexFile(cfg.CompanyWikiRoot, wikiStore); err != nil {
		return companyWikiEditResponse{}, http.StatusInternalServerError, err
	}
	if err := companyknowledge.AppendLogEntry(cfg.CompanyWikiRoot, companyknowledge.WikiLogEntry{
		Action:         "edit_apply",
		Title:          req.Title,
		Slug:           page.Page.Slug,
		Status:         "published",
		Actor:          req.Actor,
		Reason:         req.Reason,
		WikiRevisionID: page.Revision.ID,
		Summary:        "Applied audited wiki edit.",
	}); err != nil {
		return companyWikiEditResponse{}, http.StatusInternalServerError, err
	}
	return companyWikiEditResponse{OK: true, Audit: completed, Page: &page}, http.StatusCreated, nil
}

func validateCompanyWikiEditApplyBody(wikiStore store.CompanyWikiStore, body string, citations []store.CompanyWikiCitationInput) error {
	frontmatter, markdownBody, err := parseCompanyWikiFrontmatter(body)
	if err != nil {
		return err
	}
	if strings.TrimSpace(markdownBody) == "" {
		return errors.New("body after YAML frontmatter is required")
	}
	if strings.TrimSpace(frontmatterString(frontmatter["title"])) == "" {
		return errors.New("frontmatter.title is required")
	}
	sourceRevisionIDs := frontmatterStringSlice(frontmatter["source_revision_ids"])
	if len(sourceRevisionIDs) == 0 {
		return errors.New("frontmatter.source_revision_ids is required")
	}
	sourceRevisionSet := map[string]struct{}{}
	for _, revisionID := range sourceRevisionIDs {
		sourceRevisionSet[revisionID] = struct{}{}
	}
	for _, citation := range citations {
		revisionID := strings.TrimSpace(citation.SourceRevisionID)
		if _, ok := sourceRevisionSet[revisionID]; !ok {
			return fmt.Errorf("frontmatter.source_revision_ids must include citation source_revision_id %q", revisionID)
		}
		if err := validateCompanyWikiCitationReference(wikiStore, citation); err != nil {
			return err
		}
	}
	return validateCompanyWikiBodyPrivacy(body)
}

func parseCompanyWikiFrontmatter(body string) (map[string]any, string, error) {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, "", errors.New("body must start with YAML frontmatter")
	}
	closeIndex := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closeIndex = i
			break
		}
	}
	if closeIndex < 0 {
		return nil, "", errors.New("YAML frontmatter must have a closing delimiter")
	}
	rawFrontmatter := strings.Join(lines[1:closeIndex], "\n")
	frontmatter := map[string]any{}
	if err := yaml.Unmarshal([]byte(rawFrontmatter), &frontmatter); err != nil {
		return nil, "", fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	if len(frontmatter) == 0 {
		return nil, "", errors.New("YAML frontmatter must not be empty")
	}
	return frontmatter, strings.Join(lines[closeIndex+1:], "\n"), nil
}

func frontmatterString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func frontmatterStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return uniqueNonEmpty(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, frontmatterString(item))
		}
		return uniqueNonEmpty(out)
	case string:
		return uniqueNonEmpty([]string{typed})
	default:
		return nil
	}
}

func validateCompanyWikiCitationReference(wikiStore store.CompanyWikiStore, citation store.CompanyWikiCitationInput) error {
	chunks, err := wikiStore.ListCompanyWikiSourceChunks(citation.SourceDocumentID)
	if err != nil {
		return err
	}
	for _, chunk := range chunks {
		if chunk.ID == strings.TrimSpace(citation.ChunkID) && chunk.RevisionID == strings.TrimSpace(citation.SourceRevisionID) {
			return nil
		}
	}
	return fmt.Errorf("citation chunk %q for source_revision_id %q was not found", strings.TrimSpace(citation.ChunkID), strings.TrimSpace(citation.SourceRevisionID))
}

func validateCompanyWikiBodyPrivacy(body string) error {
	for _, pattern := range companyWikiSensitiveBodyPatterns {
		if pattern.re.MatchString(body) {
			return fmt.Errorf("body failed privacy validation: %s detected", pattern.name)
		}
	}
	return nil
}
