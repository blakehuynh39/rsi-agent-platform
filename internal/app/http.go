package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

func NewBaseRouter(cfg config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": cfg.ServiceName})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ready", "service": cfg.ServiceName})
	})
	r.Get("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"service": cfg.ServiceName,
			"env":     cfg.Environment,
			"queues": map[string]string{
				"workflow":  cfg.WorkflowQueueURL,
				"proactive": cfg.ProactiveQueueURL,
				"eval":      cfg.EvalQueueURL,
				"proposal":  cfg.ProposalQueueURL,
				"sandbox":   cfg.SandboxQueueURL,
			},
			"default_repo": cfg.DefaultRepo,
		})
	})
	return r
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func ListenAndServe(cfg config.Config, handler http.Handler) error {
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	return http.ListenAndServe(addr, handler)
}

func SanitizedTracePath(traceID string) string {
	return path.Clean("/" + traceID)
}
