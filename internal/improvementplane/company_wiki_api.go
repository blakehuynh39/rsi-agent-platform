package improvementplane

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type companyWikiSearchResponse struct {
	OK      bool                               `json:"ok"`
	Query   string                             `json:"query"`
	Results []storepkg.CompanyWikiSearchResult `json:"results"`
}

func registerCompanyWikiRoutes(r chi.Router, cfg config.Config, repo storepkg.Repository) {
	r.Get("/api/company-wiki/search", func(w http.ResponseWriter, r *http.Request) {
		wikiStore, ok := repo.(storepkg.CompanyWikiStore)
		if !ok {
			app.WriteError(w, http.StatusInternalServerError, errors.New("configured store does not support company wiki"))
			return
		}
		limit := parsePositiveIntQuery(r.URL.Query().Get("limit"), 10)
		results, err := wikiStore.SearchCompanyWikiPages(r.URL.Query().Get("query"), limit)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		writeCompanyWikiCachedJSON(w, http.StatusOK, 15, companyWikiSearchResponse{
			OK:      true,
			Query:   strings.TrimSpace(r.URL.Query().Get("query")),
			Results: results,
		})
	})
	r.Get("/api/company-wiki/index", func(w http.ResponseWriter, r *http.Request) {
		read, err := companyknowledge.ReadIndexFile(cfg.CompanyWikiRoot)
		if err == nil {
			writeCompanyWikiCachedJSON(w, http.StatusOK, 30, read)
			return
		}
		if !errors.Is(err, os.ErrNotExist) {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		body, fallbackErr := companyWikiDashboardIndexFallback(repo)
		if fallbackErr != nil {
			app.WriteError(w, http.StatusInternalServerError, fallbackErr)
			return
		}
		writeCompanyWikiCachedJSON(w, http.StatusOK, 30, companyknowledge.WikiMarkdownRead{OK: true, Path: "index.md", Content: body})
	})
	r.Get("/api/company-wiki/log", func(w http.ResponseWriter, r *http.Request) {
		limit := parsePositiveIntQuery(r.URL.Query().Get("limit"), 0)
		read, err := companyknowledge.ReadLogFile(cfg.CompanyWikiRoot, limit)
		if err == nil {
			writeCompanyWikiCachedJSON(w, http.StatusOK, 30, read)
			return
		}
		if errors.Is(err, os.ErrNotExist) && companyWikiDashboardManifestEmpty(repo) {
			writeCompanyWikiCachedJSON(w, http.StatusOK, 30, companyknowledge.WikiMarkdownRead{OK: true, Path: "log.md", Content: "# Company Wiki Log\n\n_No wiki log entries yet._\n"})
			return
		}
		app.WriteError(w, http.StatusNotFound, err)
	})
	r.Get("/api/company-wiki/file", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		if decoded, err := url.PathUnescape(path); err == nil {
			path = decoded
		}
		read, err := companyknowledge.ReadWikiMarkdownFile(cfg.CompanyWikiRoot, path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fallback, found, fallbackErr := companyWikiDashboardFileFallback(repo, path)
				if fallbackErr != nil {
					app.WriteError(w, http.StatusInternalServerError, fallbackErr)
					return
				}
				if found {
					writeCompanyWikiCachedJSON(w, http.StatusOK, 30, fallback)
					return
				}
			}
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		writeCompanyWikiCachedJSON(w, http.StatusOK, 30, read)
	})
}

func writeCompanyWikiCachedJSON(w http.ResponseWriter, status int, maxAgeSeconds int, payload any) {
	if maxAgeSeconds > 0 {
		w.Header().Set(
			"Cache-Control",
			fmt.Sprintf("private, max-age=%d, stale-while-revalidate=%d", maxAgeSeconds, maxAgeSeconds*4),
		)
	}
	app.WriteJSON(w, status, payload)
}

func companyWikiDashboardIndexFallback(repo storepkg.Repository) (string, error) {
	wikiStore, ok := repo.(storepkg.CompanyWikiStore)
	if !ok {
		return "", errors.New("configured store does not support company wiki")
	}
	return companyknowledge.BuildIndexMarkdown(wikiStore)
}

func companyWikiDashboardFileFallback(repo storepkg.Repository, rawPath string) (companyknowledge.WikiMarkdownRead, bool, error) {
	wikiStore, ok := repo.(storepkg.CompanyWikiStore)
	if !ok {
		return companyknowledge.WikiMarkdownRead{}, false, errors.New("configured store does not support company wiki")
	}
	cleanPath := companyWikiDashboardCleanPath(rawPath)
	if cleanPath == "" {
		return companyknowledge.WikiMarkdownRead{}, false, nil
	}
	entries, err := wikiStore.ListCompanyWikiManifestEntries()
	if err != nil {
		return companyknowledge.WikiMarkdownRead{}, false, err
	}
	for _, entry := range entries {
		if companyWikiDashboardCleanPath(entry.Path) != cleanPath {
			continue
		}
		page, found, err := wikiStore.GetCompanyWikiPage(entry.WikiPageID)
		if err != nil {
			return companyknowledge.WikiMarkdownRead{}, false, err
		}
		if found {
			return companyWikiDashboardMarkdownReadFromPage(page, cleanPath), true, nil
		}
	}
	for _, ref := range companyWikiDashboardRefsForPath(cleanPath) {
		page, found, err := wikiStore.GetCompanyWikiPage(ref)
		if err != nil {
			return companyknowledge.WikiMarkdownRead{}, false, err
		}
		if found {
			return companyWikiDashboardMarkdownReadFromPage(page, cleanPath), true, nil
		}
	}
	return companyknowledge.WikiMarkdownRead{}, false, nil
}

func companyWikiDashboardMarkdownReadFromPage(page storepkg.CompanyWikiPageRead, requestedPath string) companyknowledge.WikiMarkdownRead {
	path := strings.TrimSpace(page.Revision.Path)
	if path == "" {
		path = requestedPath
	}
	return companyknowledge.WikiMarkdownRead{
		OK:      true,
		Path:    path,
		Content: page.Revision.Body,
	}
}

func companyWikiDashboardRefsForPath(cleanPath string) []string {
	base := strings.TrimSuffix(cleanPath, ".md")
	refs := []string{cleanPath}
	if base != cleanPath {
		refs = append(refs, base)
	}
	if slug, ok := strings.CutPrefix(base, "pages/"); ok {
		refs = append(refs, slug)
	}
	return dedupeCompanyWikiDashboardRefs(refs)
}

func dedupeCompanyWikiDashboardRefs(refs []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	return out
}

func companyWikiDashboardCleanPath(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	value = strings.TrimPrefix(value, "/")
	value = path.Clean(value)
	for strings.HasPrefix(value, "../") {
		value = strings.TrimPrefix(value, "../")
	}
	if value == "." || value == ".." {
		return ""
	}
	return value
}

func companyWikiDashboardManifestEmpty(repo storepkg.Repository) bool {
	wikiStore, ok := repo.(storepkg.CompanyWikiStore)
	if !ok {
		return false
	}
	entries, err := wikiStore.ListCompanyWikiManifestEntries()
	return err == nil && len(entries) == 0
}

func parsePositiveIntQuery(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
