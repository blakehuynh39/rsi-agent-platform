package reviewui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var dist embed.FS

func NewHandler(apiBaseURL string) http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "embedded ui unavailable", http.StatusInternalServerError)
		})
	}

	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			http.NotFound(w, r)
			return
		}
		if strings.Contains(r.URL.Path, ".") {
			fileServer.ServeHTTP(w, r)
			return
		}
		req := r.Clone(r.Context())
		req.URL.Path = "/"
		fileServer.ServeHTTP(w, req)
	})
}

