package control

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func NewRouter(cfg config.Config, store storepkg.Repository) http.Handler {
	r := app.NewBaseRouter(cfg)

	r.Get("/api/events", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"service":     cfg.ServiceName,
			"events":      store.ListEvents(),
			"ingestions":  store.ListIngestions(),
			"workflows":   store.ListWorkflows(),
			"assignments": store.ListAssignments(),
			"work_items":  store.ListWorkItems(),
		})
	})
	r.Post("/api/events", func(w http.ResponseWriter, r *http.Request) {
		var event ingestion.EventEnvelope
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		created, err := store.CreateEvent(event)
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, created)
	})
	r.Get("/api/ingestions", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"service":     cfg.ServiceName,
			"events":      store.ListEvents(),
			"ingestions":  store.ListIngestions(),
			"workflows":   store.ListWorkflows(),
			"assignments": store.ListAssignments(),
			"work_items":  store.ListWorkItems(),
		})
	})
	r.Post("/api/ingestions", func(w http.ResponseWriter, r *http.Request) {
		var envelope slack.SlackEnvelope
		if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		if envelope.CreatedAt.IsZero() {
			envelope.CreatedAt = time.Now().UTC()
		}
		ingestion := store.CreateIngestion(envelope)
		app.WriteJSON(w, http.StatusCreated, ingestion)
	})
	r.Post("/webhooks/github", func(w http.ResponseWriter, r *http.Request) {
		handleGitHubWebhook(cfg, store, w, r)
	})
	r.Get("/api/thread-policies", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"thread_policies":  store.ListThreadPolicies(),
			"channel_policies": store.ListChannelPolicies(),
			"ownership":        store.ListOwnershipRecords(),
			"capabilities":     store.ListCapabilities(),
			"templates":        store.ListTemplates(),
			"experiments":      store.ListExperiments(),
		})
	})
	r.Post("/api/thread-policies/{threadKey}/mute", func(w http.ResponseWriter, r *http.Request) {
		threadKey := chi.URLParam(r, "threadKey")
		item, err := store.SetThreadState(threadKey, policy.ThreadStateMuted, "")
		if err != nil {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteJSON(w, http.StatusOK, item)
	})
	r.Post("/api/thread-policies/{threadKey}/resume", func(w http.ResponseWriter, r *http.Request) {
		threadKey := chi.URLParam(r, "threadKey")
		item, err := store.SetThreadState(threadKey, policy.ThreadStateActive, "")
		if err != nil {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteJSON(w, http.StatusOK, item)
	})
	r.Get("/api/orchestration/status", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"workflow_count":   len(store.ListWorkflows()),
			"assignment_count": len(store.ListAssignments()),
			"thread_count":     len(store.ListThreadPolicies()),
			"work_item_count":  len(store.ListWorkItems()),
		})
	})
	r.Post("/api/sandbox/jobs/preview", func(w http.ResponseWriter, r *http.Request) {
		var req sandbox.JobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		if req.Repo == "" {
			req.Repo = cfg.DefaultRepo
		}
		app.WriteJSON(w, http.StatusOK, sandbox.BuildJob(cfg, req))
	})

	return r
}
