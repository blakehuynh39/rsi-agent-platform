package improvementplane

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/reviewui"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func NewRouter(cfg config.Config, store storepkg.Repository) http.Handler {
	r := app.NewBaseRouter(cfg)
	r.Get("/api/conversations", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{"conversations": buildConversationList(store)})
	})
	r.Get("/api/conversations/{conversationID}", func(w http.ResponseWriter, r *http.Request) {
		conversationID := chi.URLParam(r, "conversationID")
		payload, ok := buildConversationDetail(store, conversationID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("conversation not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, payload)
	})
	r.Get("/api/cases", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{"cases": buildCaseList(store)})
	})
	r.Get("/api/cases/{caseID}", func(w http.ResponseWriter, r *http.Request) {
		caseID := chi.URLParam(r, "caseID")
		payload, ok := buildCaseDetail(store, caseID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("case not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, payload)
	})
	r.Post("/api/feedback", func(w http.ResponseWriter, r *http.Request) {
		var body review.FeedbackRecord
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		traceID, err := resolveFeedbackTargetTraceID(store, body)
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		now := time.Now().UTC()
		receipt, err := submitProblemLineCommand(
			store,
			traceID,
			transition.CommandProblemLineRecordFeedback,
			firstNonEmptyString(body.ReviewerID, "ui-operator"),
			now,
			fmt.Sprintf("cmd-problem-line:feedback:%s:%d", traceID, now.UnixNano()),
			map[string]any{
				"trace_id":    traceID,
				"target_type": string(body.TargetType),
				"target_id":   strings.TrimSpace(body.TargetID),
				"score":       body.Score,
				"verdict":     strings.TrimSpace(body.Verdict),
				"labels":      append([]string(nil), body.Labels...),
				"notes":       body.Notes,
				"reviewer_id": firstNonEmptyString(body.ReviewerID, "ui-operator"),
			},
		)
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		item, err := loadFeedbackFromReceipt(store, receipt)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, item)
	})
	r.Get("/api/actions", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"action_intents": sliceOrEmpty(listActionIntents(store, actionFilters{
				ConversationID: r.URL.Query().Get("conversation"),
				CaseID:         r.URL.Query().Get("case"),
				TraceID:        r.URL.Query().Get("trace"),
				ProposalID:     r.URL.Query().Get("proposal"),
			})),
		})
	})
	r.Get("/api/actions/{actionID}", func(w http.ResponseWriter, r *http.Request) {
		actionID := chi.URLParam(r, "actionID")
		payload, ok := buildActionDetail(store, actionID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("action not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, payload)
	})
	r.Get("/api/outcomes", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"outcomes": sliceOrEmpty(listOutcomes(
				store,
				r.URL.Query().Get("conversation"),
				r.URL.Query().Get("case"),
				r.URL.Query().Get("trace"),
				r.URL.Query().Get("proposal"),
			)),
		})
	})
	r.Post("/api/problem-lines/{aggregateID}/commands", func(w http.ResponseWriter, r *http.Request) {
		aggregateID := chi.URLParam(r, "aggregateID")
		receipt, ok := app.SubmitMachineCommand(w, r, store, transition.MachineProblemLine, aggregateID, "ui-operator")
		if !ok {
			return
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
	})
	r.Get("/api/knowledge", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"knowledge_entries": sliceOrEmpty(listKnowledgeEntries(store, knowledgeFilters{
				Tier:      r.URL.Query().Get("tier"),
				Status:    r.URL.Query().Get("status"),
				ScopeType: r.URL.Query().Get("scope_type"),
				ScopeID:   r.URL.Query().Get("scope_id"),
			})),
		})
	})
	r.Get("/api/knowledge/{knowledgeID}", func(w http.ResponseWriter, r *http.Request) {
		knowledgeID := chi.URLParam(r, "knowledgeID")
		payload, ok := buildKnowledgeDetail(store, knowledgeID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("knowledge entry not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, payload)
	})
	r.Post("/api/knowledge/{knowledgeID}/commands", func(w http.ResponseWriter, r *http.Request) {
		knowledgeID := chi.URLParam(r, "knowledgeID")
		receipt, ok := app.SubmitMachineCommand(w, r, store, transition.MachineKnowledge, knowledgeID, "ui-operator")
		if !ok {
			return
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
	})
	r.Get("/api/traces/{traceID}", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		payload, ok := buildTraceDetail(store, traceID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("trace not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, payload)
	})
	r.Post("/api/traces/{traceID}/replay", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		var payload struct {
			RequestedBy string `json:"requested_by"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if _, ok := store.GetTrace(traceID); !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("trace not found"))
			return
		}
		trace, _ := store.GetTrace(traceID)
		now := time.Now().UTC()
		requestedBy := firstNonEmptyString(payload.RequestedBy, "ui-operator")
		receipt, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflowLine,
			AggregateID: trace.Summary.CaseID,
			CommandKind: string(transition.CommandWorkflowLineScheduleRetry),
			CommandID:   fmt.Sprintf("cmd-workflow-line:replay:%s:%d", traceID, now.UnixNano()),
			Actor:       requestedBy,
			OccurredAt:  now,
			Payload: map[string]any{
				"requested_by":       requestedBy,
				"source_workflow_id": trace.Summary.WorkflowID,
				"source_trace_id":    traceID,
				"retry_decision":     "manual_replay",
				"retry_after":        now,
				"next_retry_action":  "activate_retry",
				"trace_description":  fmt.Sprintf("Queued manual replay from trace %s.", traceID),
			},
		})
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		line, ok := store.GetWorkflowLine(trace.Summary.CaseID)
		if ok && line.CurrentWorkflowID != "" {
			if _, err := store.SubmitCommand(transition.CommandEnvelope{
				MachineKind: transition.MachineWorkflowLine,
				AggregateID: trace.Summary.CaseID,
				CommandKind: string(transition.CommandWorkflowLineActivateRetry),
				CommandID:   fmt.Sprintf("%s:activate", receipt.CommandID),
				Actor:       requestedBy,
				OccurredAt:  now,
			}); err != nil {
				app.WriteError(w, http.StatusBadRequest, err)
				return
			}
			if _, err := store.SubmitCommand(transition.CommandEnvelope{
				MachineKind: transition.MachineWorkflow,
				AggregateID: line.CurrentWorkflowID,
				CommandKind: string(transition.CommandWorkflowStarted),
				CommandID:   fmt.Sprintf("cmd-workflow:%s:%s", line.CurrentWorkflowID, transition.CommandWorkflowStarted),
				Actor:       requestedBy,
				OccurredAt:  now,
				Payload: map[string]any{
					"default_repo":         cfg.DefaultRepo,
					"allowed_target_repos": append([]string(nil), cfg.AllowedTargetRepos...),
					"knowledge_base_url":   cfg.DefaultKnowledgeBaseURL,
					"sandbox_namespace":    cfg.SandboxNamespace,
					"resume_queue":         string(queue.WorkflowQueue),
				},
			}); err != nil {
				app.WriteError(w, http.StatusBadRequest, err)
				return
			}
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
	})
	r.Get("/api/workflow-attempts/{workflowID}", func(w http.ResponseWriter, r *http.Request) {
		workflowID := chi.URLParam(r, "workflowID")
		payload, ok := buildWorkflowAttemptDetail(store, workflowID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("workflow attempt not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, payload)
	})
	r.Get("/api/proposals", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"proposals":         buildProposalSummaries(store),
			"proposal_slots":    normalizeProposalSlots(store.GetProposalSlots()),
			"candidates":        sliceOrEmpty(store.ListCandidates()),
			"runtime_diagnoses": sliceOrEmpty(store.ListRuntimeDiagnoses()),
			"settings":          store.GetSettings(),
		})
	})
	r.Get("/api/proposals/{proposalID}", func(w http.ResponseWriter, r *http.Request) {
		proposalID := chi.URLParam(r, "proposalID")
		payload, ok := buildProposalDetail(store, proposalID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("proposal not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, payload)
	})
	r.Get("/api/proposals/{proposalID}/attempts/{attemptID}", func(w http.ResponseWriter, r *http.Request) {
		proposalID := chi.URLParam(r, "proposalID")
		attemptID := chi.URLParam(r, "attemptID")
		payload, ok := buildAttemptDetail(store, proposalID, attemptID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("attempt not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, payload)
	})
	r.Get("/api/runtime", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"roles":  buildRuntimeStatus(cfg, store),
			"honcho": buildHonchoRuntimeStatus(cfg),
		})
	})
	r.Get("/api/harness", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, buildHarnessOverview(cfg, store))
	})
	r.Get("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, store.GetSettings())
	})
	r.Post("/api/settings/commands", func(w http.ResponseWriter, r *http.Request) {
		receipt, ok := app.SubmitMachineCommand(w, r, store, transition.MachineSettings, "settings", "ui-operator")
		if !ok {
			return
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
	})
	r.Post("/api/proposals/{proposalID}/commands", func(w http.ResponseWriter, r *http.Request) {
		proposalID := chi.URLParam(r, "proposalID")
		receipt, ok := app.SubmitMachineCommand(w, r, store, transition.MachineProposalLine, proposalID, "ui-operator")
		if !ok {
			return
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
	})
	r.Post("/api/attempts/{attemptID}/commands", func(w http.ResponseWriter, r *http.Request) {
		attemptID := chi.URLParam(r, "attemptID")
		receipt, ok := app.SubmitMachineCommand(w, r, store, transition.MachineAttempt, attemptID, "ui-operator")
		if !ok {
			return
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
	})
	r.Post("/api/harness/{overlayID}/commands", func(w http.ResponseWriter, r *http.Request) {
		overlayID := chi.URLParam(r, "overlayID")
		receipt, ok := app.SubmitMachineCommand(w, r, store, transition.MachineHarness, overlayID, "ui-operator")
		if !ok {
			return
		}
		app.WriteJSON(w, http.StatusAccepted, receipt)
	})
	r.Get("/api/cron/status", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"proposal_slots": store.GetProposalSlots(),
			"settings":       store.GetSettings(),
			"service":        cfg.ServiceName,
		})
	})

	ui := reviewui.NewHandler(cfg.PublicBaseURL)
	r.Handle("/*", ui)
	return r
}

func sliceOrEmpty[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

func normalizeProposalSlots(state storepkg.ProposalSlotState) storepkg.ProposalSlotState {
	state.ActiveProposalIDs = sliceOrEmpty(state.ActiveProposalIDs)
	state.StaleProposalIDs = sliceOrEmpty(state.StaleProposalIDs)
	return state
}

func RunCron(cfg config.Config, store storepkg.Repository, once bool) {
	run := func() {
		queuedDiagnoses, err := queueRuntimeDiagnoses(cfg, store, cfg.ServiceName, time.Now().UTC())
		if err != nil {
			log.Printf("improvement-plane runtime diagnosis queue error: %v", err)
			return
		}
		receipt, err := submitProblemLineCommand(
			store,
			"problem-lines",
			transition.CommandProblemLinePromote,
			cfg.ServiceName,
			time.Now().UTC(),
			fmt.Sprintf("cmd-problem-line:promote:%s:%d", cfg.ServiceName, time.Now().UTC().UnixNano()),
			map[string]any{
				"requested_by": cfg.ServiceName,
			},
		)
		if err != nil {
			log.Printf("improvement-plane cron error: %v", err)
			return
		}
		result, err := loadPromotionResultFromReceipt(store, receipt)
		if err != nil {
			log.Printf("improvement-plane cron result error: %v", err)
			return
		}
		log.Printf("improvement-plane cron queued_diagnoses=%d promoted=%d blocked_by_cap=%t stale=%v ids=%v", queuedDiagnoses, result.Promoted, result.BlockedByCap, result.StaleProposalIDs, result.PromotedIDs)
	}
	if once {
		run()
		return
	}
	ticker := time.NewTicker(cfg.ProposalPromoterInterval)
	defer ticker.Stop()
	run()
	for range ticker.C {
		run()
	}
}
