package control

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestNativeToolsAuthFailsClosedAndRejectsStaticToken(t *testing.T) {
	store := storepkg.NewMemoryStore()
	reqBody := []byte(`{"surface":"slack","operation":"channels_list"}`)

	missingTokenRouter := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage", NativeToolsEnabled: true}, store)
	missingRec := httptest.NewRecorder()
	missingReq := httptest.NewRequest(http.MethodPost, "/internal/native-tools/actions", bytes.NewReader(reqBody))
	missingTokenRouter.ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusUnauthorized {
		t.Fatalf("missing client token status = %d, want %d", missingRec.Code, http.StatusUnauthorized)
	}

	cfg := nativeToolsTestConfig()
	router := NewRouter(cfg, store)
	staticRec := httptest.NewRecorder()
	staticReq := httptest.NewRequest(http.MethodPost, "/internal/native-tools/actions", bytes.NewReader(reqBody))
	staticReq.Header.Set("Authorization", "Bearer "+cfg.NativeToolsClientToken)
	router.ServeHTTP(staticRec, staticReq)
	if staticRec.Code != http.StatusUnauthorized {
		t.Fatalf("static token status = %d, want %d", staticRec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(staticRec.Body.String(), "static native tools client token") {
		t.Fatalf("expected static token rejection, got %s", staticRec.Body.String())
	}
}

func TestNativeToolsRejectsExpiredTamperedAndWrongAudienceTokens(t *testing.T) {
	cfg := nativeToolsTestConfig()
	router := NewRouter(cfg, storepkg.NewMemoryStore())
	body := []byte(`{"surface":"slack","operation":"channels_list"}`)
	now := time.Now().UTC()

	cases := []struct {
		name  string
		token string
	}{
		{
			name: "expired",
			token: nativeToolsTestToken(t, cfg, nativeToolClaims{
				Audience: nativeToolsAudience, IssuedAt: now.Add(-20 * time.Minute).Unix(), ExpiresAt: now.Add(-10 * time.Minute).Unix(),
				ExecutionID: "exec-1", OperationID: "op-1", TraceID: "trace-1", WorkflowID: "wf-1", ConversationID: "conv-1", Actor: "user-1", Surfaces: []string{"slack"},
			}),
		},
		{
			name: "wrong-audience",
			token: nativeToolsTestToken(t, cfg, nativeToolClaims{
				Audience: "other", IssuedAt: now.Unix(), ExpiresAt: now.Add(time.Hour).Unix(),
				ExecutionID: "exec-1", OperationID: "op-1", TraceID: "trace-1", WorkflowID: "wf-1", ConversationID: "conv-1", Actor: "user-1", Surfaces: []string{"slack"},
			}),
		},
		{
			name:  "tampered",
			token: nativeToolsTestToken(t, cfg, nativeToolsValidClaims(now, "slack")) + "x",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/internal/native-tools/actions", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+tc.token)
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}
		})
	}
}

func TestNativeToolsIdempotentReplayConflictAndFailureAudit(t *testing.T) {
	cfg := nativeToolsTestConfig()
	store := storepkg.NewMemoryStore()
	router := NewRouter(cfg, store)
	token := nativeToolsTestToken(t, cfg, nativeToolsValidClaims(time.Now().UTC(), "slack"))
	body := []byte(`{"surface":"slack","operation":"message_post","idempotency_key":"idem-1","reason":"reply to user","arguments":{"channel_id":"C123","text":"hello"}}`)

	first := nativeToolsPost(t, router, token, body)
	if first.Code != http.StatusFailedDependency {
		t.Fatalf("first status = %d, want %d body=%s", first.Code, http.StatusFailedDependency, first.Body.String())
	}
	var firstPayload nativeToolActionResponse
	if err := json.Unmarshal(first.Body.Bytes(), &firstPayload); err != nil {
		t.Fatalf("decode first response: %v", err)
	}
	if firstPayload.Action.State != storepkg.ExternalToolActionStateFailed {
		t.Fatalf("first action state = %s, want failed", firstPayload.Action.State)
	}
	if firstPayload.Action.ErrorMessage == "" {
		t.Fatalf("expected failure error recorded in action: %#v", firstPayload.Action)
	}
	if len(store.ListExternalToolActions()) != 1 {
		t.Fatalf("expected one action record, got %d", len(store.ListExternalToolActions()))
	}

	replay := nativeToolsPost(t, router, token, body)
	if replay.Code != http.StatusOK {
		t.Fatalf("replay status = %d, want %d body=%s", replay.Code, http.StatusOK, replay.Body.String())
	}
	var replayPayload nativeToolActionResponse
	if err := json.Unmarshal(replay.Body.Bytes(), &replayPayload); err != nil {
		t.Fatalf("decode replay response: %v", err)
	}
	if !replayPayload.Replayed || replayPayload.Action.ID != firstPayload.Action.ID {
		t.Fatalf("expected replay of same action, got %#v", replayPayload)
	}

	conflictBody := []byte(`{"surface":"slack","operation":"message_post","idempotency_key":"idem-1","reason":"reply to user","arguments":{"channel_id":"C123","text":"different"}}`)
	conflict := nativeToolsPost(t, router, token, conflictBody)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status = %d, want %d body=%s", conflict.Code, http.StatusConflict, conflict.Body.String())
	}
	if len(store.ListExternalToolActions()) != 1 {
		t.Fatalf("conflicting replay should not create a second action, got %d", len(store.ListExternalToolActions()))
	}
}

func TestNativeToolReplayReturnsPersistedResultPayload(t *testing.T) {
	cfg := nativeToolsTestConfig()
	store := storepkg.NewMemoryStore()
	claims := nativeToolsValidClaims(time.Now().UTC(), "slack")
	input := nativeToolActionRequest{
		Surface:        "slack",
		Operation:      "message_post",
		IdempotencyKey: "idem-result",
		Reason:         "reply to user",
		Arguments:      map[string]any{"channel_id": "C123", "text": "hello"},
	}

	first, _, _ := handleNativeToolAction(context.Background(), cfg, store, claims, input)
	if first.Action.ID == "" {
		t.Fatalf("expected first call to create action: %#v", first)
	}
	_, err := store.UpdateExternalToolActionResult(first.Action.ID, storepkg.ExternalToolActionResultUpdate{
		State:           storepkg.ExternalToolActionStateSucceeded,
		ResponseSummary: "posted Slack message",
		SourceRef:       "slack:C123:123.456",
		ResultPayload:   map[string]any{"channel_id": "C123", "ts": "123.456"},
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("seed result payload: %v", err)
	}

	replay, status, err := handleNativeToolAction(context.Background(), cfg, store, claims, input)
	if err != nil {
		t.Fatalf("replay returned error: %v", err)
	}
	if status != http.StatusOK || !replay.OK || !replay.Replayed {
		t.Fatalf("unexpected replay response status=%d payload=%#v", status, replay)
	}
	output := mapValue(replay.Output)
	if output["ts"] != "123.456" {
		t.Fatalf("expected persisted result payload on replay, got %#v", replay.Output)
	}
}

func TestNativeToolSlackReportCanResumeSucceededPartialUpload(t *testing.T) {
	action := storepkg.ExternalToolAction{
		State: storepkg.ExternalToolActionStateSucceeded,
		ResultPayload: map[string]any{
			"render_manifest": map[string]any{
				"main_message": map[string]any{
					"status":     "posted",
					"channel_id": "C123",
					"ts":         "123.456",
				},
				"uploads": []any{
					map[string]any{"id": "table-0", "status": "failed"},
				},
			},
		},
	}
	input := nativeToolActionRequest{Surface: "slack", Operation: "report_post"}
	if !nativeToolActionCanResume(input, action) {
		t.Fatalf("expected succeeded Slack report with failed uploads to resume")
	}
}

func TestNativeToolSlackReportResumeWithNilArgumentsDoesNotPanic(t *testing.T) {
	cfg := nativeToolsTestConfig()
	store := storepkg.NewMemoryStore()
	claims := nativeToolsValidClaims(time.Now().UTC(), "slack")
	input := nativeToolActionRequest{
		Surface:        "slack",
		Operation:      "report_post",
		IdempotencyKey: "report-nil-args",
		Reason:         "resume report",
	}
	requestHash, err := nativeToolRequestHash(input, false)
	if err != nil {
		t.Fatalf("hash request: %v", err)
	}
	now := time.Now().UTC()
	actionRecord, _, err := store.UpsertExternalToolAction(storepkg.ExternalToolActionCreateInput{
		Surface:        input.Surface,
		Operation:      input.Operation,
		IdempotencyKey: input.IdempotencyKey,
		RequestHash:    requestHash,
		Actor:          claims.Actor,
		Reason:         input.Reason,
		ExecutionID:    claims.ExecutionID,
		OperationID:    claims.OperationID,
		TraceID:        claims.TraceID,
		WorkflowID:     claims.WorkflowID,
		ConversationID: claims.ConversationID,
	}, now)
	if err != nil {
		t.Fatalf("seed external action: %v", err)
	}
	_, err = store.UpdateExternalToolActionResult(actionRecord.ID, storepkg.ExternalToolActionResultUpdate{
		State: storepkg.ExternalToolActionStateFailed,
		ResultPayload: map[string]any{
			"render_manifest": map[string]any{
				"main_message": map[string]any{
					"status":     "posted",
					"channel_id": "C123",
					"ts":         "123.456",
				},
				"uploads": []any{
					map[string]any{"id": "table-0", "status": "failed"},
				},
			},
		},
	}, now)
	if err != nil {
		t.Fatalf("seed external action result: %v", err)
	}

	resp, status, err := handleNativeToolAction(context.Background(), cfg, store, claims, input)
	if err != nil {
		t.Fatalf("resume returned unexpected transport error: %v", err)
	}
	if status == 0 {
		t.Fatalf("expected HTTP status for nil-argument resume, got response=%#v", resp)
	}
	stored, ok := store.GetExternalToolAction(actionRecord.ID)
	if !ok {
		t.Fatalf("expected stored external action %s", actionRecord.ID)
	}
	if !slackReportResultHasPostedMain(stored.ResultPayload) {
		t.Fatalf("expected failed resume to preserve posted main message manifest, got %#v", stored.ResultPayload)
	}
}

func TestNativeNotionWriteResolvesMirrorRootFromSourceMirror(t *testing.T) {
	cfg := nativeToolsTestConfig()
	cfg.NotionMirrorEnabled = true
	state := storepkg.NewMemoryStore()
	_, err := state.MarkSourceMirrorRecordStale(storepkg.SourceMirrorRecord{
		SourceType:       companyknowledge.NotionDocumentSourceType,
		SourceKey:        companyknowledge.NotionDocumentSourceKey("notion", "child-page"),
		Workspace:        "notion",
		Environment:      "stage",
		SourceSessionKey: companyknowledge.NotionDocumentSessionKey("notion", "child-page"),
		HonchoWorkspace:  "rsi_company_knowledge",
		HonchoSessionID:  "notion_child_page",
		SourceRevision:   "rev-1",
		Metadata: map[string]any{
			"notion_page_id": "child-page",
			"notion_root_id": "root-page",
		},
	}, "seed", nil)
	if err != nil {
		t.Fatalf("seed notion source mirror record: %v", err)
	}
	router := NewRouter(cfg, state)
	token := nativeToolsTestToken(t, cfg, nativeToolsValidClaims(time.Now().UTC(), "notion"))
	body := []byte(`{"surface":"notion","operation":"page_update","idempotency_key":"notion-root","reason":"test root resolution","arguments":{"page_id":"child-page","properties":{}}}`)

	rec := nativeToolsPost(t, router, token, body)
	if rec.Code != http.StatusFailedDependency {
		t.Fatalf("status = %d, want missing-token dependency after root resolution; body=%s", rec.Code, rec.Body.String())
	}
	var payload nativeToolActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if strings.Contains(payload.Action.ErrorMessage, "mirror_root_id") {
		t.Fatalf("expected source mirror root resolution, got validation error: %#v", payload.Action)
	}
}

func TestNativeKnowledgeMessagesReadRefusesUnboundedChannelRead(t *testing.T) {
	cfg := nativeToolsTestConfig()
	cfg.SlackMirrorChannelDiscovery = "explicit"
	cfg.SlackMirrorChannelAllowlist = []string{"C123"}
	router := NewRouter(cfg, storepkg.NewMemoryStore())
	token := nativeToolsTestToken(t, cfg, nativeToolsValidClaims(time.Now().UTC(), "knowledge"))
	body := []byte(`{"surface":"knowledge","operation":"messages_read","arguments":{"channel_id":"C123"}}`)

	rec := nativeToolsPost(t, router, token, body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "oldest_ts or latest_ts") {
		t.Fatalf("expected bounded read error, got %s", rec.Body.String())
	}
}

func nativeToolsTestConfig() config.Config {
	return config.Config{
		ServiceName:            "control-plane",
		Environment:            "stage",
		NativeToolsEnabled:     true,
		NativeToolsClientToken: "native-secret",
		AllowedSlackChannelIDs: []string{"C123"},
	}
}

func nativeToolsValidClaims(now time.Time, surfaces ...string) nativeToolClaims {
	return nativeToolClaims{
		Audience:       nativeToolsAudience,
		IssuedAt:       now.Unix(),
		ExpiresAt:      now.Add(time.Hour).Unix(),
		ExecutionID:    "exec-1",
		OperationID:    "op-1",
		TraceID:        "trace-1",
		WorkflowID:     "wf-1",
		ConversationID: "conv-1",
		Actor:          "user-1",
		Surfaces:       surfaces,
	}
}

func nativeToolsTestToken(t *testing.T, cfg config.Config, claims nativeToolClaims) string {
	t.Helper()
	token, err := mintNativeToolsExecutionToken(cfg.NativeToolsClientToken, claims)
	if err != nil {
		t.Fatalf("mint token: %v", err)
	}
	return token
}

func nativeToolsPost(t *testing.T, router http.Handler, token string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/native-tools/actions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(rec, req)
	return rec
}
