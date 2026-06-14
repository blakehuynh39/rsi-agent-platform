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

func TestNativeKanbanCreateTicketRequiresProjectSelectionAsFailedNonMutation(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")
	resp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:        "kanban",
		Operation:      "create_ticket",
		IdempotencyKey: "kanban-missing-project",
		Reason:         "user asked RSI to create a ticket",
		Arguments: map[string]any{
			"title":      "Create the Kanban board",
			"channel_id": "C123",
			"message_ts": "171000001.000200",
		},
	})
	if err == nil || status != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d err=%v resp=%#v, want 422 needs_project_selection", status, err, resp)
	}
	if resp.OK || resp.Action.State != storepkg.ExternalToolActionStateFailed {
		t.Fatalf("missing project should be failed/non-mutating, got %#v", resp)
	}
	output, ok := resp.Output.(map[string]any)
	if !ok {
		t.Fatalf("output type = %T, want map", resp.Output)
	}
	if reason, _ := output["reason"].(string); reason != "needs_project_selection" {
		t.Fatalf("output reason=%q, want needs_project_selection", reason)
	}
	if len(state.ListKanbanTickets("")) != 0 {
		t.Fatalf("missing project should not create tickets")
	}
}

func TestNativeKanbanProjectToolsCreateAndListProjects(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")

	createResp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:        "kanban",
		Operation:      "create_project",
		IdempotencyKey: "kanban-create-project-numo",
		Reason:         "Blake asked RSI to create the Numo Kanban project",
		Arguments: map[string]any{
			"slug":        "numo",
			"name":        "Numo",
			"description": "Numo project board",
		},
	})
	if err != nil || status != http.StatusOK || !createResp.OK {
		t.Fatalf("create project status=%d err=%v resp=%#v", status, err, createResp)
	}
	if got := len(state.ListKanbanProjects()); got != 1 {
		t.Fatalf("expected one Kanban project, got %d", got)
	}
	project, ok := state.GetKanbanProject("numo")
	if !ok {
		t.Fatalf("created project not found by slug")
	}
	if _, ok := state.GetKanbanDefaultBoard(project.ID); !ok {
		t.Fatalf("create_project should create the default board")
	}

	listResp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:   "kanban",
		Operation: "list_projects",
		Arguments: map[string]any{
			"include_routes": true,
		},
	})
	if err != nil || status != http.StatusOK || !listResp.OK {
		t.Fatalf("list projects status=%d err=%v resp=%#v", status, err, listResp)
	}
	output, ok := listResp.Output.(map[string]any)
	if !ok {
		t.Fatalf("list output type=%T", listResp.Output)
	}
	projects, ok := output["projects"].([]storepkg.KanbanProject)
	if !ok || len(projects) != 1 {
		t.Fatalf("projects=%#v", output["projects"])
	}
	routes, ok := output["routes"].([]storepkg.KanbanProjectSlackRoute)
	if !ok || len(routes) != 0 {
		t.Fatalf("routes=%#v", output["routes"])
	}
}

func TestNativeKanbanSetProjectSlackRouteTool(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	project, err := state.CreateKanbanProject(storepkg.KanbanProjectCreateInput{Slug: "trace", Name: "Trace"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")
	resp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:        "kanban",
		Operation:      "set_project_slack_route",
		IdempotencyKey: "kanban-bind-trace-thread",
		Reason:         "Bind the Slack thread to the Trace Kanban project",
		Arguments: map[string]any{
			"project_slug": "trace",
			"team_id":      "T123",
			"channel_id":   "C123",
			"thread_ts":    "171000001.000100",
		},
	})
	if err != nil || status != http.StatusOK || !resp.OK {
		t.Fatalf("set route status=%d err=%v resp=%#v", status, err, resp)
	}
	resolved, ok := state.ResolveKanbanSlackProject("T123", "C123", "171000001.000100")
	if !ok || resolved.ID != project.ID {
		t.Fatalf("expected route to resolve project %s, got %#v ok=%v", project.ID, resolved, ok)
	}
}

func TestNativeKanbanListProjectRoutesUsesSlackContextAsRouteFilter(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	project, err := state.CreateKanbanProject(storepkg.KanbanProjectCreateInput{Slug: "trace", Name: "Trace"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := state.SetKanbanSlackProjectRoute(storepkg.KanbanProjectSlackRouteInput{
		ProjectID: project.ID,
		TeamID:    "T123",
		ChannelID: "C123",
		Actor:     storepkg.KanbanActor{Type: "test", ID: "tester"},
	}, time.Now().UTC()); err != nil {
		t.Fatalf("set slack route: %v", err)
	}
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")
	resp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:   "kanban",
		Operation: "list_project_routes",
		Arguments: map[string]any{
			"team_id":    "T123",
			"channel_id": "C123",
		},
	})
	if err != nil || status != http.StatusOK || !resp.OK {
		t.Fatalf("list routes status=%d err=%v resp=%#v", status, err, resp)
	}
	output, ok := resp.Output.(map[string]any)
	if !ok {
		t.Fatalf("list routes output type=%T", resp.Output)
	}
	routes, ok := output["routes"].([]storepkg.KanbanProjectSlackRoute)
	if !ok || len(routes) != 1 || routes[0].ProjectID != project.ID {
		t.Fatalf("routes=%#v", output["routes"])
	}
}

func TestNativeKanbanCreateTicketWithProject(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	project, err := state.CreateKanbanProject(storepkg.KanbanProjectCreateInput{Slug: "platform", Name: "Platform"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")
	resp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:        "kanban",
		Operation:      "create_ticket",
		IdempotencyKey: "kanban-create-project",
		Reason:         "track user request",
		Arguments: map[string]any{
			"project_id": project.ID,
			"title":      "Ship Kanban",
			"message_ts": "171000001.000200",
		},
	})
	if err != nil || status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d err=%v resp=%#v", status, err, resp)
	}
	if len(state.ListKanbanTickets(project.ID)) != 1 {
		t.Fatalf("expected one project ticket, got %d", len(state.ListKanbanTickets(project.ID)))
	}
}

func TestNativeKanbanCreateTicketUsesSlackChannelDefault(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	project, err := state.CreateKanbanProject(storepkg.KanbanProjectCreateInput{Slug: "platform", Name: "Platform"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := state.SetKanbanSlackProjectRoute(storepkg.KanbanProjectSlackRouteInput{
		ProjectID: project.ID,
		TeamID:    "T123",
		ChannelID: "C123",
		Actor:     storepkg.KanbanActor{Type: "test", ID: "tester"},
	}, time.Now().UTC()); err != nil {
		t.Fatalf("set slack route: %v", err)
	}
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")
	resp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:        "kanban",
		Operation:      "create_ticket",
		IdempotencyKey: "kanban-create-from-slack-route",
		Reason:         "track slack request",
		Arguments: map[string]any{
			"team_id":    "T123",
			"channel_id": "C123",
			"title":      "Route-created ticket",
			"message_ts": "171000001.000200",
		},
	})
	if err != nil || status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d err=%v resp=%#v", status, err, resp)
	}
	if len(state.ListKanbanTickets(project.ID)) != 1 {
		t.Fatalf("expected one project ticket, got %d", len(state.ListKanbanTickets(project.ID)))
	}
	sourceRefs := 0
	snapshot, ok := state.GetKanbanBoardSnapshot(project.ID)
	if ok {
		sourceRefs = len(snapshot.SourceRefs)
	}
	if sourceRefs != 1 {
		t.Fatalf("expected one Slack source ref, got %d", sourceRefs)
	}
}

func TestNativeKanbanCreateTicketUsesUniqueTeamScopedSlackDefaultWithoutTeamArg(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	project, err := state.CreateKanbanProject(storepkg.KanbanProjectCreateInput{Slug: "platform", Name: "Platform"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := state.SetKanbanSlackProjectRoute(storepkg.KanbanProjectSlackRouteInput{
		ProjectID: project.ID,
		TeamID:    "T123",
		ChannelID: "C123",
		Actor:     storepkg.KanbanActor{Type: "test", ID: "tester"},
	}, time.Now().UTC()); err != nil {
		t.Fatalf("set slack route: %v", err)
	}
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")
	claims.SlackChannelID = "C123"
	claims.SlackThreadTS = "171000001.000100"
	resp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:        "kanban",
		Operation:      "create_ticket",
		IdempotencyKey: "kanban-create-unique-team-route",
		Reason:         "track slack request",
		Arguments: map[string]any{
			"title":      "Route-created ticket",
			"message_ts": "171000001.000200",
		},
	})
	if err != nil || status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d err=%v resp=%#v", status, err, resp)
	}
	if len(state.ListKanbanTickets(project.ID)) != 1 {
		t.Fatalf("expected one project ticket, got %d", len(state.ListKanbanTickets(project.ID)))
	}
}

func TestNativeKanbanCreateTicketRejectsAmbiguousTeamScopedSlackDefaultWithoutTeamArg(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	first, err := state.CreateKanbanProject(storepkg.KanbanProjectCreateInput{Slug: "first", Name: "First"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create first project: %v", err)
	}
	second, err := state.CreateKanbanProject(storepkg.KanbanProjectCreateInput{Slug: "second", Name: "Second"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}
	for _, route := range []struct {
		projectID string
		teamID    string
	}{
		{first.ID, "T123"},
		{second.ID, "T999"},
	} {
		if _, err := state.SetKanbanSlackProjectRoute(storepkg.KanbanProjectSlackRouteInput{
			ProjectID: route.projectID,
			TeamID:    route.teamID,
			ChannelID: "C123",
			Actor:     storepkg.KanbanActor{Type: "test", ID: "tester"},
		}, time.Now().UTC()); err != nil {
			t.Fatalf("set slack route: %v", err)
		}
	}
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")
	claims.SlackChannelID = "C123"
	resp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
		Surface:        "kanban",
		Operation:      "create_ticket",
		IdempotencyKey: "kanban-create-ambiguous-team-route",
		Reason:         "track slack request",
		Arguments: map[string]any{
			"title":      "Ambiguous route ticket",
			"message_ts": "171000001.000200",
		},
	})
	if err == nil || status != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d err=%v resp=%#v, want 422", status, err, resp)
	}
	if resp.OK || resp.Action.State != storepkg.ExternalToolActionStateFailed {
		t.Fatalf("ambiguous route should be failed/non-mutating, got %#v", resp)
	}
	if got := len(state.ListKanbanTickets("")); got != 0 {
		t.Fatalf("ambiguous route should not create tickets, got %d", got)
	}
}

func TestNativeKanbanCreateTicketWithoutExactSlackMessageDoesNotCollapseThread(t *testing.T) {
	cfg := nativeToolsTestConfig()
	state := storepkg.NewMemoryStore()
	project, err := state.CreateKanbanProject(storepkg.KanbanProjectCreateInput{Slug: "platform", Name: "Platform"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := state.SetKanbanSlackProjectRoute(storepkg.KanbanProjectSlackRouteInput{
		ProjectID: project.ID,
		ChannelID: "C123",
		Actor:     storepkg.KanbanActor{Type: "test", ID: "tester"},
	}, time.Now().UTC()); err != nil {
		t.Fatalf("set slack route: %v", err)
	}
	claims := nativeToolsValidClaims(time.Now().UTC(), "kanban")
	claims.SlackThreadTS = "171000001.000100"
	for i, title := range []string{"First ticket", "Second ticket"} {
		resp, status, err := handleNativeToolAction(context.Background(), cfg, state, claims, nativeToolActionRequest{
			Surface:        "kanban",
			Operation:      "create_ticket",
			IdempotencyKey: "kanban-thread-create-" + title,
			Reason:         "track slack request",
			Arguments: map[string]any{
				"channel_id": "C123",
				"title":      title,
			},
		})
		if err != nil || status != http.StatusOK || !resp.OK {
			t.Fatalf("call %d status=%d err=%v resp=%#v", i, status, err, resp)
		}
	}
	snapshot, ok := state.GetKanbanBoardSnapshot(project.ID)
	if !ok {
		t.Fatalf("missing board snapshot")
	}
	if len(snapshot.Tickets) != 2 {
		t.Fatalf("expected two distinct tickets, got %d", len(snapshot.Tickets))
	}
	if len(snapshot.SourceRefs) != 0 {
		t.Fatalf("without exact message_ts, source refs should not be created; got %d", len(snapshot.SourceRefs))
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

func TestNativeSlackWritesAreBoundToExecutionThreadScope(t *testing.T) {
	cfg := nativeToolsTestConfig()
	router := NewRouter(cfg, storepkg.NewMemoryStore())
	now := time.Now().UTC()
	claims := nativeToolsValidClaims(now, "slack")
	claims.SlackThreadTS = "171000001.000100"
	token := nativeToolsTestToken(t, cfg, claims)

	allowed := nativeToolsPost(t, router, token, []byte(`{"surface":"slack","operation":"message_post","idempotency_key":"bound-ok","reason":"reply","arguments":{"channel_id":"C123","thread_ts":"171000001.000100","text":"hello"}}`))
	if allowed.Code != http.StatusFailedDependency {
		t.Fatalf("allowed status = %d, want missing Slack token dependency after policy pass; body=%s", allowed.Code, allowed.Body.String())
	}

	allowedReaction := nativeToolsPost(t, router, token, []byte(`{"surface":"slack","operation":"reaction_add","idempotency_key":"bound-reaction-ok","reason":"ack","arguments":{"channel_id":"C123","timestamp":"171000001.000100","name":"white_check_mark"}}`))
	if allowedReaction.Code != http.StatusFailedDependency {
		t.Fatalf("allowed reaction status = %d, want missing Slack token dependency after policy pass; body=%s", allowedReaction.Code, allowedReaction.Body.String())
	}

	wrongChannel := nativeToolsPost(t, router, token, []byte(`{"surface":"slack","operation":"message_post","idempotency_key":"bound-channel","reason":"reply","arguments":{"channel_id":"C999","thread_ts":"171000001.000100","text":"hello"}}`))
	if wrongChannel.Code != http.StatusForbidden || !strings.Contains(wrongChannel.Body.String(), "outside bound Slack delivery scope") {
		t.Fatalf("wrong channel response = %d %s", wrongChannel.Code, wrongChannel.Body.String())
	}

	wrongReactionTimestamp := nativeToolsPost(t, router, token, []byte(`{"surface":"slack","operation":"reaction_add","idempotency_key":"bound-reaction-wrong","reason":"ack","arguments":{"channel_id":"C123","timestamp":"171000002.000200","name":"white_check_mark"}}`))
	if wrongReactionTimestamp.Code != http.StatusForbidden || !strings.Contains(wrongReactionTimestamp.Body.String(), "outside bound Slack delivery scope") {
		t.Fatalf("wrong reaction timestamp response = %d %s", wrongReactionTimestamp.Code, wrongReactionTimestamp.Body.String())
	}

	wrongThread := nativeToolsPost(t, router, token, []byte(`{"surface":"slack","operation":"file_upload","idempotency_key":"bound-thread","reason":"upload","arguments":{"channel_id":"C123","thread_ts":"171000002.000200","content":"hello","filename":"hello.txt"}}`))
	if wrongThread.Code != http.StatusForbidden || !strings.Contains(wrongThread.Body.String(), "outside bound Slack delivery scope") {
		t.Fatalf("wrong thread response = %d %s", wrongThread.Code, wrongThread.Body.String())
	}

	noThread := nativeToolsPost(t, router, token, []byte(`{"surface":"slack","operation":"report_post","idempotency_key":"bound-no-thread","reason":"report","arguments":{"channel_id":"C123","report_schema_version":1,"summary":"hello"}}`))
	if noThread.Code != http.StatusForbidden || !strings.Contains(noThread.Body.String(), "requires thread_ts") {
		t.Fatalf("missing thread response = %d %s", noThread.Code, noThread.Body.String())
	}

	unsupportedWrite := nativeToolsPost(t, router, token, []byte(`{"surface":"slack","operation":"channel_create","idempotency_key":"bound-channel-create","reason":"create","arguments":{"name":"nope"}}`))
	if unsupportedWrite.Code != http.StatusForbidden || !strings.Contains(unsupportedWrite.Body.String(), "not available to workflow execution tokens") {
		t.Fatalf("unsupported write response = %d %s", unsupportedWrite.Code, unsupportedWrite.Body.String())
	}
}

func TestNativeSlackWritesStillHonorEmergencyDenylist(t *testing.T) {
	cfg := nativeToolsTestConfig()
	router := NewRouter(cfg, storepkg.NewMemoryStore())
	claims := nativeToolsValidClaims(time.Now().UTC(), "slack")
	claims.SlackChannelID = "CDENY"
	token := nativeToolsTestToken(t, cfg, claims)

	rec := nativeToolsPost(t, router, token, []byte(`{"surface":"slack","operation":"message_post","idempotency_key":"deny","reason":"reply","arguments":{"channel_id":"CDENY","text":"hello"}}`))
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "denied by policy") {
		t.Fatalf("denylist response = %d %s", rec.Code, rec.Body.String())
	}
}

func TestNativeSentryIssuesListUsesCLIWithServerSideToken(t *testing.T) {
	cfg := nativeToolsTestConfig()
	cfg.SentryAuthToken = "sntrys-secret"
	cfg.SentryOrganization = "story-protocol"
	oldRunner := sentryCommandRunner
	defer func() { sentryCommandRunner = oldRunner }()
	var observedArgs []string
	var observedEnv []string
	sentryCommandRunner = func(ctx context.Context, args []string, env []string) ([]byte, []byte, error) {
		observedArgs = append([]string{}, args...)
		observedEnv = append([]string{}, env...)
		return []byte(`[{"shortId":"DEPIN-1","title":"boom"}]`), nil, nil
	}

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "sentry"), nativeToolActionRequest{
		Surface:   "sentry",
		Operation: "issues_list",
		Arguments: map[string]any{
			"project_ref": "depin-backend",
			"query":       "is:unresolved",
			"limit":       10,
			"sort":        "freq",
		},
	})
	if err != nil {
		t.Fatalf("native sentry action returned error: %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d response=%#v", status, resp)
	}
	if resp.Action.TargetRef != "depin-backend" {
		t.Fatalf("target ref = %q, want project_ref fallback", resp.Action.TargetRef)
	}
	wantArgs := []string{"issue", "list", "story-protocol/depin-backend", "--query", "is:unresolved", "--limit", "10", "--sort", "freq", "--json"}
	if strings.Join(observedArgs, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %#v, want %#v", observedArgs, wantArgs)
	}
	if !envContains(observedEnv, "SENTRY_AUTH_TOKEN=sntrys-secret") || !envContains(observedEnv, "SENTRY_FORCE_ENV_TOKEN=1") {
		t.Fatalf("sentry env missing required token controls: %#v", observedEnv)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if strings.Contains(string(data), "sntrys-secret") {
		t.Fatalf("native sentry response leaked token: %s", data)
	}
}

func TestNativeSentryRejectsModelSuppliedOrganization(t *testing.T) {
	cfg := nativeToolsTestConfig()
	cfg.SentryAuthToken = "sntrys-secret"
	cfg.SentryOrganization = "story-protocol"
	oldRunner := sentryCommandRunner
	defer func() { sentryCommandRunner = oldRunner }()
	sentryCommandRunner = func(ctx context.Context, args []string, env []string) ([]byte, []byte, error) {
		t.Fatalf("sentry CLI should not run for rejected cross-org request: %#v", args)
		return nil, nil, nil
	}

	for _, tc := range []struct {
		name      string
		operation string
		args      map[string]any
	}{
		{
			name:      "project_ref org prefix",
			operation: "issues_list",
			args:      map[string]any{"project_ref": "other-org/depin-backend"},
		},
		{
			name:      "project list org argument",
			operation: "projects_list",
			args:      map[string]any{"org": "other-org"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "sentry"), nativeToolActionRequest{
				Surface:   "sentry",
				Operation: tc.operation,
				Arguments: tc.args,
			})
			if err == nil || status != http.StatusBadRequest || resp.OK {
				t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
			}
			if !strings.Contains(err.Error(), "configured organization story-protocol") {
				t.Fatalf("error = %q, want configured org rejection", err.Error())
			}
		})
	}
}

func TestNativeSentryRequiresServerSideToken(t *testing.T) {
	cfg := nativeToolsTestConfig()
	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "sentry"), nativeToolActionRequest{
		Surface:   "sentry",
		Operation: "projects_list",
	})
	if err == nil || status != http.StatusFailedDependency || resp.OK {
		t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
	}
}

func TestNativeAWSReadUsesAllowlistedRunnerAndRedactsOutput(t *testing.T) {
	cfg := nativeToolsTestConfig()
	oldRunner := awsNativeRunner
	defer func() { awsNativeRunner = oldRunner }()
	var observed awsReadRequest
	awsNativeRunner = func(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error) {
		observed = req
		return map[string]any{
			"Events": []any{
				map[string]any{"Message": "DB instance restarted"},
			},
			"SecretAccessKey": "should-not-leak",
			"AccessKeyId":     "camel-case-credential",
			"NextToken":       "pagination-token-is-not-secret",
		}, nil
	}

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "aws"), nativeToolActionRequest{
		Surface:   "aws",
		Operation: "read",
		Arguments: map[string]any{
			"account":   "stage",
			"region":    "us-east-1",
			"service":   "rds",
			"operation": "describe-events",
			"params": map[string]any{
				"source_identifier": "depin-backend",
				"max_records":       20,
			},
		},
	})
	if err != nil {
		t.Fatalf("native aws read returned error: %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d response=%#v", status, resp)
	}
	if observed.Account != "stage" || observed.Region != "us-east-1" || observed.Service != "rds" || observed.Operation != "describe-events" {
		t.Fatalf("observed request = %#v", observed)
	}
	if resp.Action.TargetRef != "aws:stage:us-east-1:rds:describe-events" {
		t.Fatalf("target ref = %q", resp.Action.TargetRef)
	}
	if resp.Action.SourceRef != "aws:stage:us-east-1:rds:describe-events" {
		t.Fatalf("source ref = %q", resp.Action.SourceRef)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if strings.Contains(string(data), "should-not-leak") {
		t.Fatalf("native aws response leaked credential-shaped output: %s", data)
	}
	if strings.Contains(string(data), "camel-case-credential") {
		t.Fatalf("native aws response leaked camelCase credential-shaped output: %s", data)
	}
	if !strings.Contains(string(data), "pagination-token-is-not-secret") {
		t.Fatalf("native aws response redacted non-secret pagination token: %s", data)
	}
	if !strings.Contains(string(data), "[REDACTED]") {
		t.Fatalf("native aws response did not redact credential-shaped output: %s", data)
	}
}

func TestNativeAWSReadRedactsEmbeddedJSONStringOutput(t *testing.T) {
	cfg := nativeToolsTestConfig()
	oldRunner := awsNativeRunner
	defer func() { awsNativeRunner = oldRunner }()
	awsNativeRunner = func(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error) {
		return map[string]any{
			"Events": []any{
				map[string]any{
					"EventName":       "AssumeRole",
					"CloudTrailEvent": `{"responseElements":{"credentials":{"accessKeyId":"embedded-access-key","sessionToken":"embedded-session-token"}},"requestID":"safe-request"}`,
				},
			},
		}, nil
	}

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "aws"), nativeToolActionRequest{
		Surface:   "aws",
		Operation: "read",
		Arguments: map[string]any{
			"account":   "stage",
			"region":    "us-east-1",
			"service":   "cloudtrail",
			"operation": "lookup-events",
		},
	})
	if err != nil {
		t.Fatalf("native aws read returned error: %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d response=%#v", status, resp)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	for _, leaked := range []string{"embedded-access-key", "embedded-session-token"} {
		if strings.Contains(string(data), leaked) {
			t.Fatalf("native aws response leaked embedded JSON credential %q: %s", leaked, data)
		}
	}
	if !strings.Contains(string(data), "safe-request") {
		t.Fatalf("native aws response dropped non-secret embedded JSON fields: %s", data)
	}
}

func TestNativeAWSReadKeepsSuccessfulOutputAtTimeoutBoundary(t *testing.T) {
	cfg := nativeToolsTestConfig()
	oldRunner := awsNativeRunner
	oldTimeout := awsNativeReadTimeout
	defer func() {
		awsNativeRunner = oldRunner
		awsNativeReadTimeout = oldTimeout
	}()
	awsNativeReadTimeout = time.Nanosecond
	awsNativeRunner = func(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error) {
		<-ctx.Done()
		return map[string]any{"DBInstances": []any{map[string]any{"DBInstanceIdentifier": "stage-db"}}}, nil
	}

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "aws"), nativeToolActionRequest{
		Surface:   "aws",
		Operation: "read",
		Arguments: map[string]any{
			"account":   "stage",
			"region":    "us-east-1",
			"service":   "rds",
			"operation": "describe-db-instances",
		},
	})
	if err != nil {
		t.Fatalf("native aws read returned error: %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d response=%#v", status, resp)
	}
}

func TestNativeAWSReadBlocksSecretBearingServices(t *testing.T) {
	cfg := nativeToolsTestConfig()
	oldRunner := awsNativeRunner
	defer func() { awsNativeRunner = oldRunner }()
	awsNativeRunner = func(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error) {
		t.Fatalf("aws runner should not run for blocked secret service: %#v", req)
		return nil, nil
	}

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "aws"), nativeToolActionRequest{
		Surface:   "aws",
		Operation: "read",
		Arguments: map[string]any{
			"account":   "prod",
			"service":   "secretsmanager",
			"operation": "get-secret-value",
			"params":    map[string]any{"secret_id": "prod/depin/database"},
		},
	})
	if err != nil || status != http.StatusForbidden || resp.OK {
		t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
	}
	if !strings.Contains(resp.Error, "blocked") {
		t.Fatalf("error = %q, want blocked", resp.Error)
	}
}

func TestNativeAWSReadBlocksCamelCaseCredentialParams(t *testing.T) {
	cfg := nativeToolsTestConfig()
	oldRunner := awsNativeRunner
	defer func() { awsNativeRunner = oldRunner }()
	awsNativeRunner = func(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error) {
		t.Fatalf("aws runner should not run for blocked credential param: %#v", req)
		return nil, nil
	}

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "aws"), nativeToolActionRequest{
		Surface:   "aws",
		Operation: "read",
		Arguments: map[string]any{
			"account":   "stage",
			"service":   "cloudwatch",
			"operation": "list-metrics",
			"params": map[string]any{
				"AccessKeyId": "should-not-be-accepted",
				"NextToken":   "pagination-token-is-ok",
			},
		},
	})
	if err != nil || status != http.StatusForbidden || resp.OK {
		t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
	}
	if !strings.Contains(resp.Error, "secret-safety") {
		t.Fatalf("error = %q, want secret-safety rejection", resp.Error)
	}
}

type fakeAWSReadInput struct {
	MaxRecords int `json:"MaxRecords"`
}

type fakeAWSReadOutput struct {
	Items []string
}

type fakeAWSReadClient struct {
	observed fakeAWSReadInput
}

func (c *fakeAWSReadClient) DescribeDBParameters(ctx context.Context, input *fakeAWSReadInput, optFns ...func(*struct{})) (*fakeAWSReadOutput, error) {
	c.observed = *input
	return &fakeAWSReadOutput{Items: []string{"ok"}}, nil
}

func TestNativeAWSReadAllowsGenericSafeReadOperations(t *testing.T) {
	if err, status := validateAWSReadRequest(awsReadRequest{
		Account:   "stage",
		Region:    "us-east-1",
		Service:   "rds",
		Operation: "describe-db-parameters",
		Params:    map[string]any{"db_parameter_group_name": "default.postgres16"},
	}); err != nil || status != http.StatusOK {
		t.Fatalf("validateAWSReadRequest returned status=%d err=%v", status, err)
	}

	client := &fakeAWSReadClient{}
	out, err := callAWSReadOperation(context.Background(), client, "describe-db-parameters", map[string]any{"max_records": 3})
	if err != nil {
		t.Fatalf("callAWSReadOperation returned error: %v", err)
	}
	if client.observed.MaxRecords != 3 {
		t.Fatalf("MaxRecords = %d, want 3", client.observed.MaxRecords)
	}
	if got, ok := out.(*fakeAWSReadOutput); !ok || len(got.Items) != 1 || got.Items[0] != "ok" {
		t.Fatalf("output = %#v", out)
	}
}

func TestNativeAWSReadBlocksRawCloudWatchLogContent(t *testing.T) {
	cfg := nativeToolsTestConfig()
	oldRunner := awsNativeRunner
	defer func() { awsNativeRunner = oldRunner }()
	awsNativeRunner = func(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error) {
		t.Fatalf("aws runner should not run for raw log-content read: %#v", req)
		return nil, nil
	}

	for _, operation := range []string{"filter-log-events", "get-log-record"} {
		t.Run(operation, func(t *testing.T) {
			resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "aws"), nativeToolActionRequest{
				Surface:   "aws",
				Operation: "read",
				Arguments: map[string]any{
					"account":   "prod",
					"service":   "logs",
					"operation": operation,
					"params": map[string]any{
						"log_group_name": "/aws/eks/use1-prod/application",
						"filter_pattern": "password secret token",
					},
				},
			})
			if err != nil || status != http.StatusForbidden || resp.OK {
				t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
			}
			if !strings.Contains(resp.Error, "raw log") {
				t.Fatalf("error = %q, want raw log rejection", resp.Error)
			}
		})
	}
}

func TestNativeAWSReadLimitsSTSReadsToCallerIdentity(t *testing.T) {
	if err, status := validateAWSReadRequest(awsReadRequest{
		Account:   "stage",
		Region:    "us-east-1",
		Service:   "sts",
		Operation: "get-caller-identity",
	}); err != nil || status != http.StatusOK {
		t.Fatalf("get-caller-identity status=%d err=%v, want allowed", status, err)
	}

	for _, operation := range []string{"get-session-token", "get-federation-token"} {
		t.Run(operation, func(t *testing.T) {
			err, status := validateAWSReadRequest(awsReadRequest{
				Account:   "stage",
				Region:    "us-east-1",
				Service:   "sts",
				Operation: operation,
			})
			if err == nil || status != http.StatusForbidden {
				t.Fatalf("%s status=%d err=%v, want forbidden", operation, status, err)
			}
			if !strings.Contains(err.Error(), "caller identity") {
				t.Fatalf("error = %q, want caller identity restriction", err.Error())
			}
		})
	}
}

func TestNativeAWSReadBlocksSerialConsoleOutput(t *testing.T) {
	err, status := validateAWSReadRequest(awsReadRequest{
		Account:   "prod",
		Region:    "us-east-1",
		Service:   "ec2",
		Operation: "get-serial-console-output",
		Params:    map[string]any{"instance_id": "i-123"},
	})
	if err == nil || status != http.StatusForbidden {
		t.Fatalf("status=%d err=%v, want forbidden", status, err)
	}
	if !strings.Contains(err.Error(), "console") {
		t.Fatalf("error = %q, want console rejection", err.Error())
	}
}

func TestNativeAWSReadBlocksBootstrapUserDataReads(t *testing.T) {
	cases := []struct {
		service   string
		operation string
	}{
		{service: "ec2", operation: "describe-instance-attribute"},
		{service: "ec2", operation: "describe-launch-template-versions"},
		{service: "ec2", operation: "get-launch-template-data"},
		{service: "autoscaling", operation: "describe-launch-configurations"},
	}
	for _, tc := range cases {
		t.Run(tc.service+"_"+tc.operation, func(t *testing.T) {
			err, status := validateAWSReadRequest(awsReadRequest{
				Account:   "prod",
				Region:    "us-east-1",
				Service:   tc.service,
				Operation: tc.operation,
			})
			if err == nil || status != http.StatusForbidden {
				t.Fatalf("%s.%s status=%d err=%v, want forbidden", tc.service, tc.operation, status, err)
			}
			if !strings.Contains(err.Error(), "credential-bearing") {
				t.Fatalf("error = %q, want credential-bearing rejection", err.Error())
			}
		})
	}
}

func TestAWSConfigForReadAllowsStageAmbientButRequiresProdRole(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
	t.Setenv("AWS_SESSION_TOKEN", "test-session-token")
	cfg := nativeToolsTestConfig()
	cfg.AWSReadStageRoleARN = ""
	cfg.AWSReadProdRoleARN = ""

	stageCfg, err := awsConfigForRead(context.Background(), cfg, awsReadRequest{
		Account: "stage",
		Region:  "us-east-1",
	})
	if err != nil {
		t.Fatalf("stage ambient config returned error: %v", err)
	}
	if stageCfg.Region != "us-east-1" {
		t.Fatalf("stage region = %q, want us-east-1", stageCfg.Region)
	}

	_, err = awsConfigForRead(context.Background(), cfg, awsReadRequest{
		Account: "prod",
		Region:  "us-east-1",
	})
	if err == nil || !strings.Contains(err.Error(), "no role ARN configured") {
		t.Fatalf("prod config err = %v, want missing role error", err)
	}
}

func TestNativeAWSReadBlocksMutations(t *testing.T) {
	cfg := nativeToolsTestConfig()
	oldRunner := awsNativeRunner
	defer func() { awsNativeRunner = oldRunner }()
	awsNativeRunner = func(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error) {
		t.Fatalf("aws runner should not run for mutation: %#v", req)
		return nil, nil
	}

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "aws"), nativeToolActionRequest{
		Surface:   "aws",
		Operation: "read",
		Arguments: map[string]any{
			"account":   "stage",
			"service":   "rds",
			"operation": "modify-db-instance",
			"params":    map[string]any{"db_instance_identifier": "depin-backend"},
		},
	})
	if err != nil || status != http.StatusBadRequest || resp.OK {
		t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
	}
	if !strings.Contains(resp.Error, "not a read-only") {
		t.Fatalf("error = %q, want read-only rejection", resp.Error)
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

func TestNativeNotionPageArchiveUsesInTrashPayload(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/pages/page-archive" {
			t.Fatalf("unexpected Notion request %s %s", r.Method, r.URL.String())
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"page","id":"page-archive","in_trash":true}`))
	}))
	defer server.Close()

	cfg := nativeToolsTestConfig()
	cfg.NotionToken = "ntn-test"
	cfg.NotionAPIBaseURL = server.URL
	router := NewRouter(cfg, storepkg.NewMemoryStore())
	token := nativeToolsTestToken(t, cfg, nativeToolsValidClaims(time.Now().UTC(), "notion"))
	body := []byte(`{"surface":"notion","operation":"page_archive","idempotency_key":"archive-page","reason":"test archive payload","confirm_destroy":true,"arguments":{"page_id":"page-archive"}}`)

	rec := nativeToolsPost(t, router, token, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if gotPayload["in_trash"] != true {
		t.Fatalf("expected in_trash=true payload, got %#v", gotPayload)
	}
	if _, ok := gotPayload["archived"]; ok {
		t.Fatalf("page_archive must not send deprecated archived field, got %#v", gotPayload)
	}
}

func TestNativeNotionBlocksChildrenReturnsReadableText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/blocks/page-1/children" {
			t.Fatalf("unexpected Notion request %s %s", r.Method, r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"object":"list",
			"has_more":false,
			"next_cursor":null,
			"results":[{
				"object":"block",
				"id":"block-1",
				"type":"paragraph",
				"has_children":false,
				"paragraph":{"rich_text":[{"type":"text","plain_text":"Launch checklist","text":{"content":"Launch checklist"}}]}
			}]
		}`))
	}))
	defer server.Close()

	cfg := nativeToolsTestConfig()
	cfg.NotionToken = "ntn-test"
	cfg.NotionAPIBaseURL = server.URL
	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "notion"), nativeToolActionRequest{
		Surface:   "notion",
		Operation: "blocks_children",
		Arguments: map[string]any{"block_id": "page-1"},
	})
	if err != nil || status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
	}
	output, ok := resp.Output.(map[string]any)
	if !ok {
		t.Fatalf("output type = %T, want map", resp.Output)
	}
	results, ok := output["results"].([]map[string]any)
	if !ok || len(results) != 1 {
		t.Fatalf("results = %#v", output["results"])
	}
	if results[0]["plain_text"] != "Launch checklist" || results[0]["markdown"] != "Launch checklist" {
		t.Fatalf("expected readable block text, got %#v", results[0])
	}
	if _, ok := results[0]["type_payload"].(map[string]any); !ok {
		t.Fatalf("expected typed Notion block payload, got %#v", results[0])
	}
}

func TestNativeKnowledgeDocumentGetResolvesRawNotionIDViaMirror(t *testing.T) {
	const notionID = "ca4efc315c2143689d1543b9de66a111"
	state := storepkg.NewMemoryStore()
	if _, err := state.MarkSourceMirrorRecordStale(storepkg.SourceMirrorRecord{
		SourceType:       companyknowledge.NotionDocumentSourceType,
		SourceKey:        companyknowledge.NotionDocumentSourceKey("notion", notionID),
		Workspace:        "notion",
		Environment:      "stage",
		SourceSessionKey: companyknowledge.NotionDocumentSessionKey("notion", notionID),
		HonchoWorkspace:  "rsi_company_knowledge",
		HonchoSessionID:  "notion_" + notionID,
		HonchoObjectType: "document",
		HonchoObjectID:   "doc_notion_1",
		SourceRevision:   "rev-1",
	}, "seed", nil); err != nil {
		t.Fatalf("seed notion source mirror record: %v", err)
	}
	var gotFilters map[string]any
	honcho := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/workspaces/rsi_company_knowledge/conclusions/list" {
			t.Fatalf("unexpected Honcho request %s %s", r.Method, r.URL.String())
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode honcho payload: %v", err)
		}
		gotFilters, _ = payload["filters"].(map[string]any)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":"doc_notion_1","content":"mirrored body","observer_id":"notion_mirror","observed_id":"story_company"}],"page":1,"size":1,"pages":1,"total":1}`))
	}))
	defer honcho.Close()

	cfg := nativeToolsTestConfig()
	cfg.HonchoBaseURL = honcho.URL
	resp, status, err := handleNativeToolAction(context.Background(), cfg, state, nativeToolsValidClaims(time.Now().UTC(), "knowledge"), nativeToolActionRequest{
		Surface:   "knowledge",
		Operation: "document_get",
		Arguments: map[string]any{"document_id": notionID},
	})
	if err != nil || status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
	}
	if !strings.Contains(mustJSONForTest(t, gotFilters), "doc_notion_1") {
		t.Fatalf("expected Honcho lookup by resolved document id, got %#v", gotFilters)
	}
	output, ok := resp.Output.(map[string]any)
	if !ok {
		t.Fatalf("output type = %T, want map", resp.Output)
	}
	lookup, ok := output["lookup"].(map[string]any)
	if !ok || lookup["resolved_document_id"] != "doc_notion_1" {
		t.Fatalf("expected resolved lookup metadata, got %#v", output["lookup"])
	}
}

func TestNativeKnowledgeDocumentGetRejectsUnresolvedRawNotionID(t *testing.T) {
	cfg := nativeToolsTestConfig()
	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "knowledge"), nativeToolActionRequest{
		Surface:   "knowledge",
		Operation: "document_get",
		Arguments: map[string]any{"document_id": "6a075c2acabb4dcf91dd83667d414aac"},
	})
	if err == nil || status != http.StatusBadRequest || resp.OK {
		t.Fatalf("status=%d err=%v response=%#v", status, err, resp)
	}
	if !strings.Contains(err.Error(), "raw Notion id") {
		t.Fatalf("expected raw Notion id repair hint, got %v", err)
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
		ServiceName:                   "control-plane",
		Environment:                   "stage",
		NativeToolsEnabled:            true,
		NativeToolsClientToken:        "native-secret",
		SlackIngressAllowedChannelIDs: []string{"C123"},
		SlackMirrorChannelDenylist:    []string{"CDENY"},
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
		SlackChannelID: "C123",
		SlackScope:     "bound_thread",
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

func mustJSONForTest(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(data)
}

func envContains(env []string, value string) bool {
	for _, item := range env {
		if item == value {
			return true
		}
	}
	return false
}
