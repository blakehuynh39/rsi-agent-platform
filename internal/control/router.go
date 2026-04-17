package control

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
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
		})
	})
	r.Post("/api/events", func(w http.ResponseWriter, r *http.Request) {
		var event ingestion.EventEnvelope
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		now := time.Now().UTC()
		receipt, err := submitIngressEventCommand(
			cfg,
			store,
			event,
			"ui-operator",
			now,
			"cmd-ingress:event:"+ingressAggregateID(string(event.Source), firstNonEmpty(event.DedupeKey, event.SourceEventID, event.ID)),
		)
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		created, err := loadIngressEventFromReceipt(store, receipt)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
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
		receipt, err := submitIngressSlackCommand(
			cfg,
			store,
			envelope,
			"ui-operator",
			envelope.CreatedAt,
			"cmd-ingress:slack:"+ingressAggregateID("slack", firstNonEmpty(envelope.TS, envelope.ThreadTS, envelope.ChannelID)),
		)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		ingestion, err := loadSlackIngestionFromReceipt(store, receipt)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, ingestion)
	})
	r.Post("/webhooks/github", func(w http.ResponseWriter, r *http.Request) {
		handleGitHubWebhook(cfg, store, w, r)
	})
	r.Post("/api/workflows/{workflowID}/commands", func(w http.ResponseWriter, r *http.Request) {
		workflowID := chi.URLParam(r, "workflowID")
		receipt, ok := app.SubmitMachineCommand(w, r, store, transition.MachineWorkflow, workflowID, "ui-operator")
		if !ok {
			return
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
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
	r.Post("/api/thread-policies/{threadKey}/commands", func(w http.ResponseWriter, r *http.Request) {
		threadKey := chi.URLParam(r, "threadKey")
		receipt, ok := app.SubmitMachineCommand(w, r, store, transition.MachineThreadPolicy, threadKey, "ui-operator")
		if !ok {
			return
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
	})
	r.Get("/api/orchestration/status", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"workflow_count":   len(store.ListWorkflows()),
			"assignment_count": len(store.ListAssignments()),
			"thread_count":     len(store.ListThreadPolicies()),
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
