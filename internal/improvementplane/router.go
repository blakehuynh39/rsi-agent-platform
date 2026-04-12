package improvementplane

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/reviewui"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
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
		item, err := store.AddFeedback(body)
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, item)
	})
	r.Post("/api/traces/{traceID}/evaluate", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		run, judgments, err := store.EvaluateTrace(traceID, "manual")
		if err != nil {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, map[string]interface{}{
			"eval_run":  run,
			"judgments": judgments,
		})
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
		item, err := store.ScheduleReplay(traceID, payload.RequestedBy)
		if err != nil {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteJSON(w, http.StatusAccepted, item)
	})
	r.Get("/api/proposals", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"proposals":          normalizeProposals(store.ListProposals()),
			"proposal_slots":     normalizeProposalSlots(store.GetProposalSlots()),
			"candidates":         sliceOrEmpty(store.ListCandidates()),
			"proposal_memory":    sliceOrEmpty(store.ListProposalMemories()),
			"repo_change_jobs":   sliceOrEmpty(store.ListRepoChangeJobs()),
			"pr_attempts":        sliceOrEmpty(store.ListPRAttempts()),
			"post_merge_replays": sliceOrEmpty(store.ListPostMergeReplays()),
			"work_items":         sliceOrEmpty(store.ListWorkItems()),
			"settings":           store.GetSettings(),
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
	r.Get("/api/runtime", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"roles": buildRuntimeStatus(cfg),
		})
	})
	r.Get("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, store.GetSettings())
	})
	r.Post("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ActiveProposalCap int `json:"active_proposal_cap"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		item, err := store.UpdateSettings(improvement.Settings{ActiveProposalCap: body.ActiveProposalCap})
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusOK, item)
	})
	r.Post("/api/proposals/promote", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			RequestedBy string `json:"requested_by"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		result, err := store.RunProposalPromoter(payload.RequestedBy)
		if err != nil {
			app.WriteError(w, http.StatusConflict, err)
			return
		}
		app.WriteJSON(w, http.StatusAccepted, result)
	})
	r.Post("/api/proposals/{proposalID}/decision", func(w http.ResponseWriter, r *http.Request) {
		proposalID := chi.URLParam(r, "proposalID")
		var body review.ProposalReview
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		item, err := store.ReviewProposal(proposalID, body)
		if err != nil {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteJSON(w, http.StatusOK, item)
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
		result, err := store.RunProposalPromoter(cfg.ServiceName)
		if err != nil {
			log.Printf("improvement-plane cron error: %v", err)
			return
		}
		log.Printf("improvement-plane cron promoted=%d blocked_by_cap=%t stale=%v ids=%v", result.Promoted, result.BlockedByCap, result.StaleProposalIDs, result.PromotedIDs)
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
