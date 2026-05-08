package control

import (
	"bytes"
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
		WorkflowID:       "wf",
		HermesSessionID:  "session:a",
		HermesToolCallID: "b",
		ArgsHash:         "sha256:args",
	}, "sha256:abc")
	right := dbReadIdempotencyKey(dbReadQueryRequest{
		WorkflowID:       "wf",
		HermesSessionID:  "session",
		HermesToolCallID: "a:b",
		ArgsHash:         "sha256:args",
	}, "sha256:abc")
	if left == right {
		t.Fatalf("expected distinct idempotency keys for distinct field boundaries")
	}
}

func TestDBReadArgsHashMatchesPythonCanonicalJSON(t *testing.T) {
	got := dbReadArgsHash(dbReadQueryRequest{
		Target:  " depin-prod ",
		SQL:     " SELECT '<>&é🚀' AS value ",
		Purpose: " query ",
	})
	const want = "sha256:05f0100863a898794c62013ed117749c243dd3e6745a2ce5f1f37ef2aa63b49a"
	if got != want {
		t.Fatalf("dbReadArgsHash() = %s, want %s", got, want)
	}
}

func TestDBReadExecutionTokenVerifiesAndRejectsExpiredOrTamperedTokens(t *testing.T) {
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	auth := mintDBReadExecutionToken(t, "secret", map[string]any{
		"version":               "v1",
		"db_read_query_allowed": true,
		"execution_id":          "hexec-1",
		"conversation_id":       "conv-1",
		"workflow_id":           "wf-1",
		"trace_id":              "trace-1",
		"channel_id":            "C123",
		"thread_ts":             "171000001.000100",
		"requester":             "user:U123",
		"iat":                   now.Add(-time.Minute).Unix(),
		"exp":                   now.Add(time.Hour).Unix(),
	})
	ctx, ok := verifyDBReadExecutionToken("secret", auth, now)
	if !ok {
		t.Fatal("expected execution token to verify")
	}
	if !ctx.Scoped || !ctx.DBReadQueryAllowed || ctx.Requester != "user:U123" || ctx.WorkflowID != "wf-1" || ctx.ThreadTS != "171000001.000100" {
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
		"thread_ts":"0.000000",
		"hermes_session_id":"session-1",
		"hermes_tool_call_id":"call_db_1",
		"canonical_tool_name":"db_read.query",
		"transport_tool_name":"db_read_query",
		"args_hash":"sha256:args"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/db-read/query", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mintDBReadExecutionToken(t, cfg.DBReadClientToken, map[string]any{
		"version":               "v1",
		"db_read_query_allowed": true,
		"execution_id":          "hexec-1",
		"conversation_id":       "conv-real",
		"workflow_id":           "wf-real",
		"trace_id":              "trace-real",
		"channel_id":            "C123",
		"thread_ts":             "171000001.000100",
		"requester":             "user:U123",
		"iat":                   now.Add(-time.Minute).Unix(),
		"exp":                   now.Add(time.Hour).Unix(),
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

func TestDBReadScopedQueryCreatesExternalToolPause(t *testing.T) {
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
		"sql":"SELECT COUNT(*) FROM scripts",
		"purpose":"query",
		"hermes_session_id":"session-1",
		"hermes_tool_call_id":"call_db_1",
		"canonical_tool_name":"db_read.query",
		"transport_tool_name":"db_read_query",
		"args_hash":"sha256:args",
		"requester":"forged",
		"workflow_id":"wf-forged"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/db-read/query", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mintDBReadExecutionToken(t, cfg.DBReadClientToken, map[string]any{
		"version":               "v1",
		"db_read_query_allowed": true,
		"operation_id":          "op-real",
		"execution_id":          "hexec-real",
		"conversation_id":       "conv-real",
		"workflow_id":           "wf-real",
		"trace_id":              "trace-real",
		"channel_id":            "C123",
		"thread_ts":             "171000001.000100",
		"requester":             "user:U123",
		"iat":                   now.Add(-time.Minute).Unix(),
		"exp":                   now.Add(time.Hour).Unix(),
	}))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("query status = %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["delivery_mode"] != "external_tool_resume" {
		t.Fatalf("delivery_mode = %#v, want external_tool_resume", payload["delivery_mode"])
	}
	pauses := store.ListExternalToolPauses()
	if len(pauses) != 1 {
		t.Fatalf("expected one external pause, got %#v", pauses)
	}
	pause := pauses[0]
	if pause.WorkflowID != "wf-real" || pause.TraceID != "trace-real" || pause.OperationID != "op-real" || pause.ExecutionID != "hexec-real" {
		t.Fatalf("pause did not use authoritative token scope: %#v", pause)
	}
	if pause.HermesSessionID != "session-1" || pause.ToolCallID != "call_db_1" || pause.TransportToolName != "db_read_query" || pause.CanonicalToolName != "db_read.query" {
		t.Fatalf("pause did not store Hermes tool identity: %#v", pause)
	}
	requests := store.ListDBReadRequests()
	if len(requests) != 1 {
		t.Fatalf("expected one DB read request, got %#v", requests)
	}
	if pause.DBReadRequestID != requests[0].ID || pause.SQLSHA256 != requests[0].SQLSHA256 {
		t.Fatalf("pause/request linkage mismatch: pause=%#v request=%#v", pause, requests[0])
	}
}

func TestDBReadQueryRejectsSQLTooLongForExactApproval(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		DBReadClientToken: "secret",
		DBReadTargetsJSON: `{"targets":[{"id":"depin-prod","placement":"prod","allowed_schemas":["public"],"allowed_tables":["scripts"],"caps":{"max_rows":10,"max_bytes":4096,"timeout_seconds":5}}]}`,
	}
	router := chi.NewRouter()
	registerDBReadRoutes(router, cfg, store)
	now := time.Now().UTC()
	body, err := json.Marshal(map[string]string{
		"target": "depin-prod",
		"sql":    "SELECT " + strings.Repeat("1", 2200),
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/internal/db-read/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mintDBReadExecutionToken(t, cfg.DBReadClientToken, map[string]any{
		"version":               "v1",
		"db_read_query_allowed": true,
		"conversation_id":       "conv-real",
		"workflow_id":           "wf-real",
		"trace_id":              "trace-real",
		"channel_id":            "C123",
		"thread_ts":             "171000001.000100",
		"requester":             "user:U123",
		"iat":                   now.Add(-time.Minute).Unix(),
		"exp":                   now.Add(time.Hour).Unix(),
	}))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("query status = %d body=%s", resp.Code, resp.Body.String())
	}
	if len(store.ListDBReadRequests()) != 0 {
		t.Fatalf("long SQL should be rejected before request creation: %#v", store.ListDBReadRequests())
	}
}

func TestDBReadQueryRejectsLegacyNonHermesSubmission(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		DBReadClientToken: "secret",
		DBReadTargetsJSON: `{"targets":[{"id":"depin-prod","placement":"prod","allowed_schemas":["public"],"allowed_tables":["scripts"],"caps":{"max_rows":10,"max_bytes":4096,"timeout_seconds":5}}]}`,
	}
	router := chi.NewRouter()
	registerDBReadRoutes(router, cfg, store)
	now := time.Now().UTC()
	req := httptest.NewRequest(http.MethodPost, "/internal/db-read/query", strings.NewReader(`{"target":"depin-prod","sql":"SELECT 1","purpose":"query"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mintDBReadExecutionToken(t, cfg.DBReadClientToken, map[string]any{
		"version":               "v1",
		"db_read_query_allowed": true,
		"conversation_id":       "conv-real",
		"workflow_id":           "wf-real",
		"trace_id":              "trace-real",
		"channel_id":            "C123",
		"thread_ts":             "171000001.000100",
		"requester":             "user:U123",
		"iat":                   now.Add(-time.Minute).Unix(),
		"exp":                   now.Add(time.Hour).Unix(),
	}))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("query status = %d body=%s", resp.Code, resp.Body.String())
	}
	if requests := store.ListDBReadRequests(); len(requests) != 0 {
		t.Fatalf("legacy query should not create requests: %#v", requests)
	}
}

func TestDBReadQueryRejectsExecutionTokenWithoutQueryGrant(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		DBReadClientToken: "secret",
		DBReadTargetsJSON: `{"targets":[{"id":"depin-prod","placement":"prod","allowed_schemas":["public"],"allowed_tables":["scripts"],"caps":{"max_rows":10,"max_bytes":4096,"timeout_seconds":5}}]}`,
	}
	router := chi.NewRouter()
	registerDBReadRoutes(router, cfg, store)
	now := time.Now().UTC()
	req := httptest.NewRequest(http.MethodPost, "/internal/db-read/query", strings.NewReader(`{
		"target":"depin-prod",
		"sql":"SELECT 1",
		"purpose":"query",
		"hermes_session_id":"session-1",
		"hermes_tool_call_id":"call_db_1",
		"canonical_tool_name":"db_read.query",
		"transport_tool_name":"db_read_query",
		"args_hash":"sha256:args"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mintDBReadExecutionToken(t, cfg.DBReadClientToken, map[string]any{
		"version":         "v1",
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
	if resp.Code != http.StatusForbidden {
		t.Fatalf("query status = %d body=%s", resp.Code, resp.Body.String())
	}
	if requests := store.ListDBReadRequests(); len(requests) != 0 {
		t.Fatalf("ungifted token query should not create requests: %#v", requests)
	}
}

func TestDBReadQueryRejectsWrongToolIdentity(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		DBReadClientToken: "secret",
		DBReadTargetsJSON: `{"targets":[{"id":"depin-prod","placement":"prod","allowed_schemas":["public"],"allowed_tables":["scripts"],"caps":{"max_rows":10,"max_bytes":4096,"timeout_seconds":5}}]}`,
	}
	router := chi.NewRouter()
	registerDBReadRoutes(router, cfg, store)
	now := time.Now().UTC()
	req := httptest.NewRequest(http.MethodPost, "/internal/db-read/query", strings.NewReader(`{
		"target":"depin-prod",
		"sql":"SELECT 1",
		"purpose":"query",
		"hermes_session_id":"session-1",
		"hermes_tool_call_id":"call_db_1",
		"canonical_tool_name":"terminal.exec",
		"transport_tool_name":"terminal_exec",
		"args_hash":"sha256:args"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mintDBReadExecutionToken(t, cfg.DBReadClientToken, map[string]any{
		"version":               "v1",
		"db_read_query_allowed": true,
		"conversation_id":       "conv-real",
		"workflow_id":           "wf-real",
		"trace_id":              "trace-real",
		"channel_id":            "C123",
		"thread_ts":             "171000001.000100",
		"requester":             "user:U123",
		"iat":                   now.Add(-time.Minute).Unix(),
		"exp":                   now.Add(time.Hour).Unix(),
	}))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("query status = %d body=%s", resp.Code, resp.Body.String())
	}
	if requests := store.ListDBReadRequests(); len(requests) != 0 {
		t.Fatalf("wrong tool query should not create requests: %#v", requests)
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
