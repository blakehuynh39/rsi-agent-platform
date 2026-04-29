package toolgateway

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func NewRouter(cfg config.Config, store storepkg.Repository) http.Handler {
	r := app.NewBaseRouter(cfg)
	service := NewService(cfg, store)
	r.Get("/api/tools", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"service":      cfg.ServiceName,
			"capabilities": store.ListCapabilities(),
		})
	})
	r.Post("/api/tools/{toolName}/execute", func(w http.ResponseWriter, r *http.Request) {
		toolName := chi.URLParam(r, "toolName")
		input := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&input)
		app.WriteJSON(w, http.StatusOK, service.Execute(toolName, input))
	})
	r.Post("/api/github/installation-token", func(w http.ResponseWriter, r *http.Request) {
		input := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&input)
		status, out := service.GitHubInstallationToken(input)
		app.WriteJSON(w, status, out)
	})
	r.Post("/api/runtime/observations", func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		status, out := service.RecordRuntimeObservation(payload)
		app.WriteJSON(w, status, out)
	})
	return r
}
