package control

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/dbread"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type dbReadQueryRequest struct {
	Target         string `json:"target"`
	SQL            string `json:"sql"`
	Purpose        string `json:"purpose,omitempty"`
	Requester      string `json:"requester,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	WorkflowID     string `json:"workflow_id,omitempty"`
	TraceID        string `json:"trace_id,omitempty"`
	ChannelID      string `json:"channel_id,omitempty"`
	ThreadTS       string `json:"thread_ts,omitempty"`
}

func registerDBReadRoutes(r chi.Router, cfg config.Config, store storepkg.Store) {
	r.Get("/internal/db-read/sources", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeDBReadClient(cfg, r) {
			app.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized db read client"))
			return
		}
		registry, err := dbread.LoadRegistry(cfg.DBReadTargetsJSON)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		app.WriteJSON(w, http.StatusOK, map[string]any{"targets": registry.PublicSources()})
	})

	r.Get("/internal/db-read/schema", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeDBReadClient(cfg, r) {
			app.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized db read client"))
			return
		}
		targetID := strings.TrimSpace(r.URL.Query().Get("target"))
		registry, err := dbread.LoadRegistry(cfg.DBReadTargetsJSON)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		target, ok := registry.Target(targetID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, fmt.Errorf("unknown db read target %q", targetID))
			return
		}
		app.WriteJSON(w, http.StatusOK, target.SchemaView())
	})

	r.Post("/internal/db-read/validate", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeDBReadClient(cfg, r) {
			app.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized db read client"))
			return
		}
		var input dbReadQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		registry, err := dbread.LoadRegistry(cfg.DBReadTargetsJSON)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		target, ok := registry.Target(input.Target)
		if !ok {
			app.WriteError(w, http.StatusNotFound, fmt.Errorf("unknown db read target %q", input.Target))
			return
		}
		result := dbread.ValidateSQLSafety(input.SQL)
		if result.OK && target.DSN != "" {
			result = dbread.ValidateAgainstTarget(r.Context(), target, input.SQL)
		}
		app.WriteJSON(w, http.StatusOK, result)
	})

	r.Post("/internal/db-read/query", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeDBReadClient(cfg, r) {
			app.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized db read client"))
			return
		}
		var input dbReadQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		registry, err := dbread.LoadRegistry(cfg.DBReadTargetsJSON)
		if err != nil {
			app.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		target, ok := registry.Target(input.Target)
		if !ok {
			app.WriteError(w, http.StatusNotFound, fmt.Errorf("unknown db read target %q", input.Target))
			return
		}
		validation := dbread.ValidateSQLSafety(input.SQL)
		now := time.Now().UTC()
		request, created, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
			IdempotencyKey:    dbReadIdempotencyKey(input, validation.SQLSHA256),
			Target:            target.ID,
			Purpose:           firstNonEmpty(input.Purpose, "query"),
			SQL:               input.SQL,
			SQLSHA256:         validation.SQLSHA256,
			ExecutionScopeKey: dbReadExecutionScopeKey(input),
			Requester:         firstNonEmpty(input.Requester, "hermes"),
			ConversationID:    input.ConversationID,
			WorkflowID:        input.WorkflowID,
			TraceID:           input.TraceID,
			ChannelID:         input.ChannelID,
			ThreadTS:          input.ThreadTS,
			ExpiresAt:         now.Add(target.TTL()),
			Caps:              target.Caps,
			Redaction:         target.Redaction,
		}, now)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if strings.Contains(err.Error(), "db read") && (strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "is required")) {
				statusCode = http.StatusBadRequest
			}
			app.WriteError(w, statusCode, err)
			return
		}
		if !created && request.SQLSHA256 != validation.SQLSHA256 {
			app.WriteJSON(w, http.StatusConflict, map[string]any{
				"status":  "blocked_by_existing_db_read_request",
				"message": "A DB read request already exists for this execution scope and target. Use rsi-db status on the existing request instead of creating a second approval.",
				"request": request,
			})
			return
		}
		var attempt storepkg.DBReadValidationAttempt
		if created && !validation.OK {
			attempt, err = store.AppendDBReadValidationAttempt(storepkg.NewDBReadValidationAttempt(request, storepkg.DBReadValidationStatusFailed, "offline_parse", validation.Message, map[string]any{
				"error_code": validation.ErrorCode,
				"preview":    validation.Preview,
			}, now))
			if err != nil {
				app.WriteError(w, http.StatusBadRequest, err)
				return
			}
		}
		request, _ = store.GetDBReadRequest(request.ID)
		statusText := string(request.State)
		app.WriteJSON(w, http.StatusAccepted, map[string]any{
			"request":            request,
			"validation":         validation,
			"validation_attempt": attempt,
			"status":             statusText,
		})
	})

	r.Get("/internal/db-read/requests/{requestID}", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeDBReadClient(cfg, r) {
			app.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized db read client"))
			return
		}
		requestID := chi.URLParam(r, "requestID")
		request, ok := store.GetDBReadRequest(requestID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, fmt.Errorf("db read request not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"request":             request,
			"validation_attempts": store.ListDBReadValidationAttempts(request.ID),
			"execution_results":   store.ListDBReadExecutionResults(request.ID),
		})
	})
}

func authorizeDBReadClient(cfg config.Config, r *http.Request) bool {
	token := strings.TrimSpace(cfg.DBReadClientToken)
	if token == "" {
		return true
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	expected := "Bearer " + token
	return subtle.ConstantTimeCompare([]byte(auth), []byte(expected)) == 1
}

func dbReadIdempotencyKey(input dbReadQueryRequest, hash string) string {
	parts := []string{
		strings.TrimSpace(input.ConversationID),
		strings.TrimSpace(input.ThreadTS),
		strings.TrimSpace(input.Target),
		strings.TrimSpace(hash),
		strings.TrimSpace(firstNonEmpty(input.Requester, "hermes")),
		strings.TrimSpace(firstNonEmpty(input.Purpose, "query")),
	}
	raw, _ := json.Marshal(parts)
	sum := sha256.Sum256(raw)
	return "dbread:sha256:" + hex.EncodeToString(sum[:])
}

func dbReadExecutionScopeKey(input dbReadQueryRequest) string {
	if value := strings.TrimSpace(input.WorkflowID); value != "" {
		return "workflow:" + value
	}
	if value := strings.TrimSpace(input.TraceID); value != "" {
		return "trace:" + value
	}
	channelID := strings.TrimSpace(input.ChannelID)
	threadTS := strings.TrimSpace(input.ThreadTS)
	if channelID != "" && threadTS != "" {
		return "thread:" + channelID + ":" + threadTS
	}
	return ""
}
