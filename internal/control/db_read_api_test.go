package control

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestDBReadIdempotencyKeyEscapesFieldBoundaries(t *testing.T) {
	left := dbReadIdempotencyKey(dbReadQueryRequest{
		ConversationID: "a:b",
		ThreadTS:       "",
		Target:         "depin-stage",
		Requester:      "hermes",
		Purpose:        "query",
	}, "sha256:abc")
	right := dbReadIdempotencyKey(dbReadQueryRequest{
		ConversationID: "a",
		ThreadTS:       "b",
		Target:         "depin-stage",
		Requester:      "hermes",
		Purpose:        "query",
	}, "sha256:abc")
	if left == right {
		t.Fatalf("expected distinct idempotency keys for distinct field boundaries")
	}
}

func TestDBReadExecutionTokenVerifiesAndRejectsExpiredOrTamperedTokens(t *testing.T) {
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	auth := mintDBReadExecutionToken(t, "secret", map[string]any{
		"version":         "v1",
		"execution_id":    "hexec-1",
		"conversation_id": "conv-1",
		"workflow_id":     "wf-1",
		"trace_id":        "trace-1",
		"channel_id":      "C123",
		"thread_ts":       "171000001.000100",
		"requester":       "user:U123",
		"iat":             now.Add(-time.Minute).Unix(),
		"exp":             now.Add(time.Hour).Unix(),
	})
	ctx, ok := verifyDBReadExecutionToken("secret", auth, now)
	if !ok {
		t.Fatal("expected execution token to verify")
	}
	if !ctx.Scoped || ctx.Requester != "user:U123" || ctx.WorkflowID != "wf-1" || ctx.ThreadTS != "171000001.000100" {
		t.Fatalf("unexpected auth context: %#v", ctx)
	}
	if _, ok := verifyDBReadExecutionToken("secret", auth+"tampered", now); ok {
		t.Fatal("expected tampered execution token to fail")
	}
	expired := mintDBReadExecutionToken(t, "secret", map[string]any{
		"version": "v1",
		"iat":     now.Add(-2 * time.Hour).Unix(),
		"exp":     now.Add(-time.Minute).Unix(),
	})
	if _, ok := verifyDBReadExecutionToken("secret", expired, now); ok {
		t.Fatal("expected expired execution token to fail")
	}
}

func TestDBReadScopedQueryOverridesModelSuppliedScope(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		DBReadClientToken: "secret",
		DBReadTargetsJSON: `{"targets":[{"id":"depin-prod","placement":"prod","allowed_schemas":["public"],"allowed_tables":["scripts"],"caps":{"max_rows":10,"max_bytes":4096,"timeout_seconds":5}}]}`,
	}
	router := chi.NewRouter()
	registerDBReadRoutes(router, cfg, store)
	now := time.Now().UTC()
	body := strings.NewReader(`{
		"target":"depin-prod",
		"sql":"SELECT 1",
		"purpose":"query",
		"requester":"attacker",
		"conversation_id":"conv-forged",
		"workflow_id":"wf-forged",
		"trace_id":"trace-forged",
		"channel_id":"C-FORGED",
		"thread_ts":"0.000000"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/db-read/query", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mintDBReadExecutionToken(t, cfg.DBReadClientToken, map[string]any{
		"version":         "v1",
		"execution_id":    "hexec-1",
		"conversation_id": "conv-real",
		"workflow_id":     "wf-real",
		"trace_id":        "trace-real",
		"channel_id":      "C123",
		"thread_ts":       "171000001.000100",
		"requester":       "user:U123",
		"iat":             now.Add(-time.Minute).Unix(),
		"exp":             now.Add(time.Hour).Unix(),
	}))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("query status = %d body=%s", resp.Code, resp.Body.String())
	}
	requests := store.ListDBReadRequests()
	if len(requests) != 1 {
		t.Fatalf("expected one DB read request, got %#v", requests)
	}
	request := requests[0]
	if request.Requester != "user:U123" ||
		request.ConversationID != "conv-real" ||
		request.WorkflowID != "wf-real" ||
		request.TraceID != "trace-real" ||
		request.ChannelID != "C123" ||
		request.ThreadTS != "171000001.000100" {
		t.Fatalf("scoped token claims were not authoritative: %#v", request)
	}
}

func TestDBReadStaticClientTokenIsRejected(t *testing.T) {
	cfg := config.Config{DBReadClientToken: "secret"}
	req := httptest.NewRequest(http.MethodGet, "/internal/db-read/sources", nil)
	req.Header.Set("Authorization", "Bearer secret")
	if _, ok := authenticateDBReadClient(cfg, req); ok {
		t.Fatal("expected static DB-read client token to be rejected")
	}
}

func TestLatestDBReadResponseOwnerMatchesScopedRequestAfterStart(t *testing.T) {
	store := storepkg.NewMemoryStore()
	started := time.Now().UTC()
	stale, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey:    "stale",
		Target:            "depin-prod",
		Purpose:           "query",
		SQL:               "select 1",
		SQLSHA256:         "sha256:stale",
		Requester:         "hermes",
		ConversationID:    "conv-1",
		ChannelID:         "C123",
		ThreadTS:          "171000001.000100",
		ExecutionScopeKey: "thread:C123:171000001.000100",
		ExpiresAt:         started.Add(time.Hour),
	}, started.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendDBReadValidationAttempt(storepkg.NewDBReadValidationAttempt(stale, storepkg.DBReadValidationStatusSucceeded, "target_prepare", "", nil, started.Add(-time.Minute))); err != nil {
		t.Fatal(err)
	}
	current, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey:    "current",
		Target:            "depin-prod",
		Purpose:           "query",
		SQL:               "select 2",
		SQLSHA256:         "sha256:current",
		Requester:         "hermes",
		ConversationID:    "conv-1",
		WorkflowID:        "wf-1",
		TraceID:           "trace-1",
		ChannelID:         "C123",
		ThreadTS:          "171000001.000100",
		ExecutionScopeKey: "workflow:wf-1",
		ExpiresAt:         started.Add(time.Hour),
	}, started.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendDBReadValidationAttempt(storepkg.NewDBReadValidationAttempt(current, storepkg.DBReadValidationStatusSucceeded, "target_prepare", "", nil, started.Add(time.Second))); err != nil {
		t.Fatal(err)
	}

	found, ok := latestDBReadResponseOwner(store, dbReadResponseScope{
		ChannelID: "C123",
		ThreadTS:  "171000001.000100",
		NotBefore: started,
	})
	if !ok {
		t.Fatal("expected scoped DB read response owner")
	}
	if found.ID != current.ID {
		t.Fatalf("expected current request %s, got %s", current.ID, found.ID)
	}
}

func TestLatestDBReadResponseOwnerIgnoresWorkflowMatchBeforeStart(t *testing.T) {
	store := storepkg.NewMemoryStore()
	started := time.Now().UTC()
	stale, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey:    "stale-workflow",
		Target:            "depin-prod",
		Purpose:           "query",
		SQL:               "select 1",
		SQLSHA256:         "sha256:stale",
		Requester:         "hermes",
		WorkflowID:        "wf-1",
		TraceID:           "trace-1",
		ExecutionScopeKey: "workflow:wf-1",
		ExpiresAt:         started.Add(time.Hour),
	}, started.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendDBReadValidationAttempt(storepkg.NewDBReadValidationAttempt(stale, storepkg.DBReadValidationStatusSucceeded, "target_prepare", "", nil, started.Add(-time.Minute))); err != nil {
		t.Fatal(err)
	}

	if found, ok := latestDBReadResponseOwner(store, dbReadResponseScope{
		WorkflowID: "wf-1",
		TraceID:    "trace-1",
		NotBefore:  started,
	}); ok {
		t.Fatalf("stale workflow/trace DB read should not own this execution's Slack response, got %s", found.ID)
	}
}

func TestLatestDBReadResponseOwnerIgnoresValidationFailure(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	request, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey:    "bad",
		Target:            "depin-prod",
		Purpose:           "query",
		SQL:               "select * from missing",
		SQLSHA256:         "sha256:bad",
		Requester:         "hermes",
		WorkflowID:        "wf-1",
		ExecutionScopeKey: "workflow:wf-1",
		ExpiresAt:         now.Add(time.Hour),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendDBReadValidationAttempt(storepkg.NewDBReadValidationAttempt(request, storepkg.DBReadValidationStatusFailed, "offline_parse", "bad sql", nil, now)); err != nil {
		t.Fatal(err)
	}

	if found, ok := latestDBReadResponseOwner(store, dbReadResponseScope{WorkflowID: "wf-1"}); ok {
		t.Fatalf("validation failures should not own Slack response, got %s", found.ID)
	}
}

func mintDBReadExecutionToken(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	rawClaims, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	encodedClaims := base64.RawURLEncoding.EncodeToString(rawClaims)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encodedClaims))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return "Bearer v1." + encodedClaims + "." + signature
}
