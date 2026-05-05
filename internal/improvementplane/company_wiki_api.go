package improvementplane

import (
	"errors"
	"net/http"
	"net/url"
	"os"
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
		app.WriteJSON(w, http.StatusOK, companyWikiSearchResponse{
			OK:      true,
			Query:   strings.TrimSpace(r.URL.Query().Get("query")),
			Results: results,
		})
	})
	r.Get("/api/company-wiki/index", func(w http.ResponseWriter, r *http.Request) {
		read, err := companyknowledge.ReadIndexFile(cfg.CompanyWikiRoot)
		if err == nil {
			app.WriteJSON(w, http.StatusOK, read)
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
		app.WriteJSON(w, http.StatusOK, companyknowledge.WikiMarkdownRead{OK: true, Path: "index.md", Content: body})
	})
	r.Get("/api/company-wiki/log", func(w http.ResponseWriter, r *http.Request) {
		limit := parsePositiveIntQuery(r.URL.Query().Get("limit"), 0)
		read, err := companyknowledge.ReadLogFile(cfg.CompanyWikiRoot, limit)
		if err == nil {
			app.WriteJSON(w, http.StatusOK, read)
			return
		}
		if errors.Is(err, os.ErrNotExist) && companyWikiDashboardManifestEmpty(repo) {
			app.WriteJSON(w, http.StatusOK, companyknowledge.WikiMarkdownRead{OK: true, Path: "log.md", Content: "# Company Wiki Log\n\n_No wiki log entries yet._\n"})
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
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteJSON(w, http.StatusOK, read)
	})
}

func companyWikiDashboardIndexFallback(repo storepkg.Repository) (string, error) {
	wikiStore, ok := repo.(storepkg.CompanyWikiStore)
	if !ok {
		return "", errors.New("configured store does not support company wiki")
	}
	entries, err := wikiStore.ListCompanyWikiManifestEntries()
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "# Company Wiki Index\n\n_No published pages yet._\n", nil
	}
	return companyknowledge.BuildIndexMarkdown(wikiStore)
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
