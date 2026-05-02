package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
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

var ErrRunnerExecutionHolderRequired = errors.New("runner execution holder is required")

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
		envelope.Prompt = slack.CanonicalizePromptEnvelope(envelope, slack.NewEntityResolver(cfg.SlackBotToken))
		receipt, err := submitIngressSlackCommand(
			r.Context(),
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
			"workflow_count":                     len(store.ListWorkflows()),
			"assignment_count":                   len(store.ListAssignments()),
			"thread_count":                       len(store.ListThreadPolicies()),
			"active_execution_count":             len(store.ListActiveRunnerExecutions()),
			"effect_scheduler_mode":              app.EffectSchedulerModeName(cfg.EffectFairClaimEnabled),
			"async_hermes_execution_enabled":     cfg.AsyncHermesExecutionEnabled,
			"deployment_active_execution_policy": cfg.DeploymentActiveExecutionPolicy,
		})
	})
	r.Post("/internal/runtime/observations", func(w http.ResponseWriter, r *http.Request) {
		var payload runtimeObservationRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		status, out := recordRuntimeObservation(store, payload)
		app.WriteJSON(w, status, out)
	})
	// These runner lifecycle endpoints are intentionally unauthenticated here.
	// They are cluster-internal control-plane hooks protected by Kubernetes/network
	// boundaries; adding app-layer auth would require extra rollout wiring for
	// drain/deploy-gate paths that must keep working during incident handling.
	lifecycle := newRunnerExecutionLifecycle(cfg, store)
	r.Get("/internal/executions/active", func(w http.ResponseWriter, r *http.Request) {
		active := lifecycle.reconcileStaleActiveExecutions()
		status := http.StatusOK
		if strings.EqualFold(cfg.DeploymentActiveExecutionPolicy, "block") && len(active) > 0 {
			status = http.StatusConflict
		}
		app.WriteJSON(w, status, map[string]any{
			"active_execution_count": len(active),
			"executions":             active,
			"deployment_policy":      cfg.DeploymentActiveExecutionPolicy,
		})
	})
	r.Post("/internal/runner-executions/{executionID}/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		executionID := chi.URLParam(r, "executionID")
		payload := map[string]any{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		item, statusCode, err := lifecycle.heartbeat(executionID, payload)
		if err != nil {
			app.WriteError(w, statusCode, err)
			return
		}
		app.WriteJSON(w, statusCode, item)
	})
	r.Post("/internal/runner-executions/{executionID}/complete", func(w http.ResponseWriter, r *http.Request) {
		executionID := chi.URLParam(r, "executionID")
		payload := map[string]any{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		item, statusCode, err := lifecycle.complete(executionID, payload)
		if err != nil {
			app.WriteError(w, statusCode, err)
			return
		}
		app.WriteJSON(w, statusCode, item)
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

type runnerExecutionHolderCAS struct {
	ExpectedHolder string
}

func validateRunnerExecutionHolder(cfg config.Config, existing storepkg.RunnerExecution, requestHolder string) (runnerExecutionHolderCAS, error) {
	requestHolder = strings.TrimSpace(requestHolder)
	if requestHolder == "" {
		return runnerExecutionHolderCAS{}, ErrRunnerExecutionHolderRequired
	}
	existingHolder := strings.TrimSpace(existing.Holder)
	if requestHolder != existingHolder {
		if cfg.HermesExecutionHeartbeatTimeout > 0 {
			now := time.Now().UTC()
			referenceTime := runnerExecutionHeartbeatReferenceTime(existing)
			if !referenceTime.IsZero() && now.Sub(referenceTime) > cfg.HermesExecutionHeartbeatTimeout {
				expectedHolder := existingHolder
				if expectedHolder == "" {
					expectedHolder = storepkg.HolderCASExpectEmpty()
				}
				return runnerExecutionHolderCAS{ExpectedHolder: expectedHolder}, nil
			}
		}
		return runnerExecutionHolderCAS{}, fmt.Errorf("holder mismatch")
	}
	return runnerExecutionHolderCAS{ExpectedHolder: existingHolder}, nil
}

func runnerExecutionHeartbeatStatusAllowed(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "accepted", "starting", "running", "cancelling", "cancel_requested":
		return true
	default:
		return false
	}
}

func runnerExecutionCompleteStatusAllowed(status string) bool {
	return storepkg.RunnerExecutionStatusTerminal(status)
}
