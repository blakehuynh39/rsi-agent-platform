package control

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
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

type dbReadAuthContext struct {
	Scoped         bool
	Requester      string
	ConversationID string
	WorkflowID     string
	TraceID        string
	ChannelID      string
	ThreadTS       string
}

func (ctx dbReadAuthContext) apply(input *dbReadQueryRequest) {
	if !ctx.Scoped || input == nil {
		return
	}
	input.Requester = ctx.Requester
	input.ConversationID = ctx.ConversationID
	input.WorkflowID = ctx.WorkflowID
	input.TraceID = ctx.TraceID
	input.ChannelID = ctx.ChannelID
	input.ThreadTS = ctx.ThreadTS
}

func registerDBReadRoutes(r chi.Router, cfg config.Config, store storepkg.Store) {
	r.Get("/internal/db-read/sources", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := authenticateDBReadClient(cfg, r); !ok {
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
		if _, ok := authenticateDBReadClient(cfg, r); !ok {
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
		if _, ok := authenticateDBReadClient(cfg, r); !ok {
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
		authCtx, ok := authenticateDBReadClient(cfg, r)
		if !ok {
			app.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized db read client"))
			return
		}
		var input dbReadQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		authCtx.apply(&input)
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
				"message": "A DB read request already exists for this execution scope and target. Use db_read.status on the existing request instead of creating a second approval.",
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
		if _, ok := authenticateDBReadClient(cfg, r); !ok {
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

func authenticateDBReadClient(cfg config.Config, r *http.Request) (dbReadAuthContext, bool) {
	token := strings.TrimSpace(cfg.DBReadClientToken)
	if token == "" {
		return dbReadAuthContext{}, true
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	return verifyDBReadExecutionToken(token, auth, time.Now().UTC())
}

func verifyDBReadExecutionToken(secret string, auth string, now time.Time) (dbReadAuthContext, bool) {
	auth = strings.TrimSpace(auth)
	if !strings.HasPrefix(auth, "Bearer ") {
		return dbReadAuthContext{}, false
	}
	rawToken := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 || parts[0] != "v1" {
		return dbReadAuthContext{}, false
	}
	expectedMAC := hmac.New(sha256.New, []byte(secret))
	_, _ = expectedMAC.Write([]byte(parts[1]))
	expected := base64.RawURLEncoding.EncodeToString(expectedMAC.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expected)) != 1 {
		return dbReadAuthContext{}, false
	}
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return dbReadAuthContext{}, false
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return dbReadAuthContext{}, false
	}
	if strings.TrimSpace(stringValueFromMap(claims, "version")) != "v1" {
		return dbReadAuthContext{}, false
	}
	expiresAt := int64(floatValueFromMap(claims, "exp"))
	issuedAt := int64(floatValueFromMap(claims, "iat"))
	if expiresAt <= 0 || now.Unix() > expiresAt {
		return dbReadAuthContext{}, false
	}
	if issuedAt > 0 && issuedAt > now.Add(5*time.Minute).Unix() {
		return dbReadAuthContext{}, false
	}
	return dbReadAuthContext{
		Scoped:         true,
		Requester:      firstNonEmpty(strings.TrimSpace(stringValueFromMap(claims, "requester")), "hermes"),
		ConversationID: strings.TrimSpace(stringValueFromMap(claims, "conversation_id")),
		WorkflowID:     strings.TrimSpace(stringValueFromMap(claims, "workflow_id")),
		TraceID:        strings.TrimSpace(stringValueFromMap(claims, "trace_id")),
		ChannelID:      strings.TrimSpace(stringValueFromMap(claims, "channel_id")),
		ThreadTS:       strings.TrimSpace(stringValueFromMap(claims, "thread_ts")),
	}, true
}

func floatValueFromMap(values map[string]any, key string) float64 {
	switch value := values[key].(type) {
	case float64:
		return value
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		parsed, _ := value.Float64()
		return parsed
	case string:
		var parsed float64
		_, _ = fmt.Sscanf(value, "%f", &parsed)
		return parsed
	default:
		return 0
	}
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
