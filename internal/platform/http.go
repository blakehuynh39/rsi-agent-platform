package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/reviewui"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
)

func NewWorkflowAPIRouter(cfg config.Config, store Store) http.Handler {
	r := newBaseRouter(cfg)
	r.Get("/api/ingestions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"service":    cfg.ServiceName,
			"ingestions": store.ListIngestions(),
			"workflows":  store.ListWorkflows(),
			"assignments": store.ListAssignments(),
		})
	})
	r.Post("/api/ingestions", func(w http.ResponseWriter, r *http.Request) {
		var envelope slack.SlackEnvelope
		if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if envelope.CreatedAt.IsZero() {
			envelope.CreatedAt = time.Now().UTC()
		}
		ingestion := store.CreateIngestion(envelope)
		writeJSON(w, http.StatusCreated, ingestion)
	})
	return r
}

func NewControlPlaneRouter(cfg config.Config, store Store) http.Handler {
	r := newBaseRouter(cfg)
	r.Get("/api/thread-policies", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"thread_policies": store.ListThreadPolicies(),
			"channel_policies": store.ListChannelPolicies(),
			"ownership": store.ListOwnershipRecords(),
			"capabilities": store.ListCapabilities(),
			"templates": store.ListTemplates(),
			"experiments": store.ListExperiments(),
		})
	})
	r.Post("/api/thread-policies/{threadKey}/mute", func(w http.ResponseWriter, r *http.Request) {
		threadKey := chi.URLParam(r, "threadKey")
		item, err := store.SetThreadState(threadKey, policy.ThreadStateMuted, "")
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})
	r.Post("/api/thread-policies/{threadKey}/resume", func(w http.ResponseWriter, r *http.Request) {
		threadKey := chi.URLParam(r, "threadKey")
		item, err := store.SetThreadState(threadKey, policy.ThreadStateActive, "")
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})
	return r
}

func NewToolGatewayRouter(cfg config.Config, store Store) http.Handler {
	r := newBaseRouter(cfg)
	r.Get("/api/tools", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"service":      cfg.ServiceName,
			"capabilities": store.ListCapabilities(),
		})
	})
	r.Post("/api/tools/{toolName}/execute", func(w http.ResponseWriter, r *http.Request) {
		toolName := chi.URLParam(r, "toolName")
		input := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&input)
		writeJSON(w, http.StatusOK, store.ExecuteTool(toolName, input))
	})
	return r
}

func NewImprovementPlaneRouter(cfg config.Config, store Store) http.Handler {
	r := newBaseRouter(cfg)
	r.Get("/api/traces", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"traces": store.ListTraces()})
	})
	r.Get("/api/traces/{traceID}", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		trace, ok := store.GetTrace(traceID)
		if !ok {
			writeError(w, http.StatusNotFound, errors.New("trace not found"))
			return
		}
		writeJSON(w, http.StatusOK, trace)
	})
	r.Get("/api/traces/{traceID}/artifacts", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		trace, ok := store.GetTrace(traceID)
		if !ok {
			writeError(w, http.StatusNotFound, errors.New("trace not found"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"artifacts": trace.Artifacts})
	})
	r.Post("/api/traces/{traceID}/ratings", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		var body review.HumanRating
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		rating, err := store.AddRating(traceID, body)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusCreated, rating)
	})
	r.Post("/api/traces/{traceID}/notes", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		var body review.ImprovementNote
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		note, err := store.AddImprovementNote(traceID, body)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusCreated, note)
	})
	r.Post("/api/traces/{traceID}/replay", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		var payload struct {
			RequestedBy string `json:"requested_by"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		item, err := store.ScheduleReplay(traceID, payload.RequestedBy)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusAccepted, item)
	})
	r.Get("/api/proposals", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"proposals": store.ListProposals()})
	})
	r.Post("/api/proposals/{proposalID}/decision", func(w http.ResponseWriter, r *http.Request) {
		proposalID := chi.URLParam(r, "proposalID")
		var body review.ProposalReview
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := store.ReviewProposal(proposalID, body)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	ui := reviewui.NewHandler(cfg.PublicBaseURL)
	r.Handle("/*", ui)
	return r
}

func newBaseRouter(cfg config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": cfg.ServiceName})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "service": cfg.ServiceName})
	})
	r.Get("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"service": cfg.ServiceName,
			"env": cfg.Environment,
			"queues": map[string]string{
				"workflow": cfg.WorkflowQueueURL,
				"proactive": cfg.ProactiveQueueURL,
				"eval": cfg.EvalQueueURL,
				"proposal": cfg.ProposalQueueURL,
				"sandbox": cfg.SandboxQueueURL,
			},
			"default_repo": cfg.DefaultRepo,
		})
	})
	return r
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
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

