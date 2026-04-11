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
	r.Get("/api/traces", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{"traces": store.ListTraces()})
	})
	r.Get("/api/evals", func(w http.ResponseWriter, r *http.Request) {
		runs := store.ListEvalRuns()
		judgments := map[string]interface{}{}
		for _, run := range runs {
			judgments[run.ID] = store.ListEvalJudgments(run.ID)
		}
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"suites":     store.ListEvalSuites(),
			"eval_runs":  runs,
			"judgments":  judgments,
			"candidates": store.ListCandidates(),
			"work_items": store.ListWorkItems(),
			"settings":   store.GetSettings(),
		})
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
		trace, ok := store.GetTrace(traceID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("trace not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, trace)
	})
	r.Get("/api/traces/{traceID}/artifacts", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		trace, ok := store.GetTrace(traceID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("trace not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{"artifacts": trace.Artifacts})
	})
	r.Post("/api/traces/{traceID}/ratings", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		var body review.HumanRating
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		rating, err := store.AddRating(traceID, body)
		if err != nil {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, rating)
	})
	r.Post("/api/traces/{traceID}/notes", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		var body review.ImprovementNote
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		note, err := store.AddImprovementNote(traceID, body)
		if err != nil {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, note)
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
			"proposals":          store.ListProposals(),
			"proposal_slots":     store.GetProposalSlots(),
			"candidates":         store.ListCandidates(),
			"proposal_memory":    store.ListProposalMemories(),
			"repo_change_jobs":   store.ListRepoChangeJobs(),
			"pr_attempts":        store.ListPRAttempts(),
			"post_merge_replays": store.ListPostMergeReplays(),
			"work_items":         store.ListWorkItems(),
			"settings":           store.GetSettings(),
		})
	})
	r.Get("/api/work-items", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"work_items": store.ListWorkItems(),
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
	r.Get("/api/candidates", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"candidates":      store.ListCandidates(),
			"proposal_memory": store.ListProposalMemories(),
			"proposal_slots":  store.GetProposalSlots(),
		})
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
