package control

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func TestRuntimeObservationEndpointRecordsHarnessAndLedgerEvents(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)
	body := bytes.NewBufferString(`{
		"execution_id":"hexec-live",
		"operation_id":"op-live",
		"trace_id":"trace-live",
		"workflow_id":"wf-live",
		"hermes_session_id":"session-live",
		"role":"prod",
		"phase":"main",
		"event_type":"model.reasoning.delta",
		"status":"streaming",
		"seq":7,
		"payload":{"delta":"hello","idempotency_key":"obs-live-7"},
		"recorded_at":"2026-05-01T04:11:15Z"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/runtime/observations", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		responseBody, _ := io.ReadAll(rec.Body)
		t.Fatalf("expected status 200, got %d: %s", rec.Code, string(responseBody))
	}
	observations := store.ListHarnessExecutionObservations()
	if len(observations) != 1 {
		t.Fatalf("expected one harness observation, got %d", len(observations))
	}
	observation := observations[0]
	if observation.ExecutionID != "hexec-live" || observation.TraceID != "trace-live" || observation.Seq != 7 {
		t.Fatalf("unexpected observation: %+v", observation)
	}
	ledger := store.ListExecutionLedgerEventsByTrace("trace-live")
	if len(ledger) != 1 {
		t.Fatalf("expected one ledger event, got %d", len(ledger))
	}
	event := ledger[0]
	if event.Kind != "model.reasoning.delta" || event.PhaseID != "main" || event.Status != "streaming" {
		t.Fatalf("unexpected ledger event: %+v", event)
	}
	if event.Payload["observation_id"] == "" || event.Payload["role"] != "prod" || event.Payload["delta"] != "hello" {
		t.Fatalf("unexpected ledger payload: %#v", event.Payload)
	}
}

func TestRuntimeObservationEndpointRejectsInvalidRecordedAt(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)
	body := bytes.NewBufferString(`{
		"execution_id":"hexec-live",
		"phase":"main",
		"event_type":"model.reasoning.delta",
		"seq":7,
		"recorded_at":"not-a-timestamp"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/runtime/observations", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		responseBody, _ := io.ReadAll(rec.Body)
		t.Fatalf("expected status 400, got %d: %s", rec.Code, string(responseBody))
	}
	if observations := store.ListHarnessExecutionObservations(); len(observations) != 0 {
		t.Fatalf("expected invalid observation to be rejected, got %#v", observations)
	}
}

func TestSourceMirrorMessageWriteIsIdempotent(t *testing.T) {
	honchoMessagesCreated := 0
	honcho := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces":
			_, _ = w.Write([]byte(`{"id":"rsi_company_knowledge","metadata":{}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/sessions":
			_, _ = w.Write([]byte(`{"id":"slack_T123_C123_1710000000_000000","workspace_id":"rsi_company_knowledge","metadata":{}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/sessions/slack_T123_C123_1710000000_000000/messages":
			honchoMessagesCreated++
			_, _ = w.Write([]byte(`[{"id":"msg_analysis_1","content":"extracted text","peer_id":"rsi_attachment_analyzer"}]`))
		default:
			t.Fatalf("unexpected honcho request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer honcho.Close()

	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{
		ServiceName:   "control-plane",
		Environment:   "stage",
		HonchoBaseURL: honcho.URL,
	}, store)
	body := `{
		"record":{
			"source_type":"slack_attachment_analysis",
			"source_key":"slack_attachment_analysis:T123:C123:1710000000.000000:F123:text",
			"workspace":"T123",
			"environment":"stage",
			"source_session_key":"slack:T123:C123:1710000000.000000",
			"honcho_workspace":"rsi_company_knowledge",
			"honcho_session_id":"slack_T123_C123_1710000000_000000",
			"source_revision":"text:sha256:abc123",
			"metadata":{"source":"slack_attachment_analysis"}
		},
		"message":{
			"content":"extracted text",
			"peer_id":"rsi_attachment_analyzer",
			"metadata":{"extraction_status":"extracted"}
		}
	}`

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/internal/source-mirror/messages", bytes.NewBufferString(body)))
	if first.Code != http.StatusCreated {
		responseBody, _ := io.ReadAll(first.Body)
		t.Fatalf("first write status = %d: %s", first.Code, string(responseBody))
	}
	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/internal/source-mirror/messages", bytes.NewBufferString(body)))
	if second.Code != http.StatusOK {
		responseBody, _ := io.ReadAll(second.Body)
		t.Fatalf("second write status = %d: %s", second.Code, string(responseBody))
	}
	if honchoMessagesCreated != 1 {
		t.Fatalf("expected one Honcho message create, got %d", honchoMessagesCreated)
	}
	var out sourceMirrorMessageWriteResponse
	if err := json.NewDecoder(second.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.ShouldWrite || out.Reason != "already_complete" || out.HonchoMessageID != "msg_analysis_1" {
		t.Fatalf("unexpected idempotent response: %+v", out)
	}
}

func TestSourceMirrorDocumentWriteIsIdempotent(t *testing.T) {
	honchoDocumentsCreated := 0
	honcho := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces":
			_, _ = w.Write([]byte(`{"id":"rsi_company_knowledge","metadata":{}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/sessions":
			_, _ = w.Write([]byte(`{"id":"notion_T123_page_abc","workspace_id":"rsi_company_knowledge","metadata":{}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/conclusions":
			honchoDocumentsCreated++
			_, _ = w.Write([]byte(`[{"id":"doc_notion_1","content":"Runbook content","observer_id":"notion_mirror","observed_id":"story_company","session_id":"notion_T123_page_abc","created_at":"2026-05-02T10:00:00Z"}]`))
		default:
			t.Fatalf("unexpected honcho request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer honcho.Close()

	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{
		ServiceName:   "control-plane",
		Environment:   "stage",
		HonchoBaseURL: honcho.URL,
	}, store)
	body := `{
		"record":{
			"source_type":"notion_document",
			"source_key":"notion_document:T123:page_abc",
			"workspace":"T123",
			"environment":"stage",
			"source_session_key":"notion:T123:page_abc",
			"honcho_workspace":"rsi_company_knowledge",
			"honcho_session_id":"notion_T123_page_abc",
			"source_revision":"last_edited_time:2026-05-02T10:00:00Z",
			"metadata":{"source":"notion","source_url":"https://notion.so/page_abc"}
		},
		"document":{
			"content":"Runbook content",
			"observer_id":"notion_mirror",
			"observed_id":"story_company",
			"metadata":{"source_page_id":"page_abc"}
		}
	}`

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/internal/source-mirror/documents", bytes.NewBufferString(body)))
	if first.Code != http.StatusCreated {
		responseBody, _ := io.ReadAll(first.Body)
		t.Fatalf("first write status = %d: %s", first.Code, string(responseBody))
	}
	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/internal/source-mirror/documents", bytes.NewBufferString(body)))
	if second.Code != http.StatusOK {
		responseBody, _ := io.ReadAll(second.Body)
		t.Fatalf("second write status = %d: %s", second.Code, string(responseBody))
	}
	if honchoDocumentsCreated != 1 {
		t.Fatalf("expected one Honcho document create, got %d", honchoDocumentsCreated)
	}
	var out sourceMirrorDocumentWriteResponse
	if err := json.NewDecoder(second.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.ShouldWrite || out.Reason != "already_complete" || out.HonchoDocumentID != "doc_notion_1" {
		t.Fatalf("unexpected idempotent response: %+v", out)
	}
	if out.Record.HonchoObjectType != "document" || out.Record.HonchoObjectID != "doc_notion_1" {
		t.Fatalf("source mirror record did not track document object: %+v", out.Record)
	}
}

func TestSourceMirrorDocumentRevisionCreatesNewHonchoDocument(t *testing.T) {
	honchoDocumentsCreated := 0
	honcho := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces":
			_, _ = w.Write([]byte(`{"id":"rsi_company_knowledge","metadata":{}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/sessions":
			_, _ = w.Write([]byte(`{"id":"notion_T123_page_abc","workspace_id":"rsi_company_knowledge","metadata":{}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/conclusions":
			honchoDocumentsCreated++
			_, _ = w.Write([]byte(fmt.Sprintf(`[{"id":"doc_notion_%d","content":"Runbook content","observer_id":"notion_mirror","observed_id":"story_company","session_id":"notion_T123_page_abc","created_at":"2026-05-02T10:00:00Z"}]`, honchoDocumentsCreated)))
		default:
			t.Fatalf("unexpected honcho request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer honcho.Close()

	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{
		ServiceName:   "control-plane",
		Environment:   "stage",
		HonchoBaseURL: honcho.URL,
	}, store)
	body := func(revision string) string {
		return `{
			"record":{
				"source_type":"notion_document",
				"source_key":"notion_document:T123:page_abc",
				"workspace":"T123",
				"environment":"stage",
				"source_session_key":"notion:T123:page_abc",
				"honcho_workspace":"rsi_company_knowledge",
				"honcho_session_id":"notion_T123_page_abc",
				"source_revision":"` + revision + `",
				"metadata":{"source":"notion"}
			},
			"document":{
				"content":"Runbook content",
				"observer_id":"notion_mirror",
				"observed_id":"story_company"
			}
		}`
	}

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/internal/source-mirror/documents", bytes.NewBufferString(body("rev1"))))
	if first.Code != http.StatusCreated {
		responseBody, _ := io.ReadAll(first.Body)
		t.Fatalf("first write status = %d: %s", first.Code, string(responseBody))
	}
	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/internal/source-mirror/documents", bytes.NewBufferString(body("rev2"))))
	if second.Code != http.StatusCreated {
		responseBody, _ := io.ReadAll(second.Body)
		t.Fatalf("second write status = %d: %s", second.Code, string(responseBody))
	}
	if honchoDocumentsCreated != 2 {
		t.Fatalf("expected two Honcho document creates after revision change, got %d", honchoDocumentsCreated)
	}
	var out sourceMirrorDocumentWriteResponse
	if err := json.NewDecoder(second.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Reason != "revision_changed" || out.HonchoDocumentID != "doc_notion_2" {
		t.Fatalf("unexpected revision response: %+v", out)
	}
}

func TestSourceMirrorStatusEndpointFailsLoudlyForMissingRequiredSourceType(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/source-mirror/status?source_type=slack_message", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		responseBody, _ := io.ReadAll(rec.Body)
		t.Fatalf("expected status 503, got %d: %s", rec.Code, string(responseBody))
	}
	var out sourceMirrorStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.OK || len(out.Issues) == 0 {
		t.Fatalf("expected failing source mirror status, got %+v", out)
	}
}

func TestSourceMirrorStatusEndpointReportsCompletedRecords(t *testing.T) {
	store := storepkg.NewMemoryStore()
	record := storepkg.SourceMirrorRecord{
		SourceType:       "slack_message",
		SourceKey:        "slack:T123:C123:1710000000.000000",
		Workspace:        "T123",
		Environment:      "stage",
		SourceSessionKey: "slack:T123:C123:channel",
		HonchoWorkspace:  "rsi_company_knowledge",
		HonchoSessionID:  "slack_T123_C123_channel",
		SourceRevision:   "rev1",
		Status:           storepkg.SourceMirrorStatusPending,
	}
	if _, err := store.ClaimSourceMirrorRecord(record, time.Minute); err != nil {
		t.Fatalf("claim source mirror record: %v", err)
	}
	if _, err := store.CompleteSourceMirrorRecord(record.SourceType, record.SourceKey, "msg_1", nil); err != nil {
		t.Fatalf("complete source mirror record: %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/source-mirror/status?source_type=slack_message&max_age_seconds=3600", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		responseBody, _ := io.ReadAll(rec.Body)
		t.Fatalf("expected status 200, got %d: %s", rec.Code, string(responseBody))
	}
	var out sourceMirrorStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !out.OK || out.SourceTypes["slack_message"].LatestComplete == nil {
		t.Fatalf("expected successful source mirror status, got %+v", out)
	}
}

func TestSourceMirrorStatusEndpointReportsHistoricalStaleRecordsWithoutFailingType(t *testing.T) {
	store := storepkg.NewMemoryStore()
	complete := storepkg.SourceMirrorRecord{
		SourceType:       companyknowledge.NotionDocumentSourceType,
		SourceKey:        companyknowledge.NotionDocumentSourceKey("notion", "active_page"),
		Workspace:        "notion",
		Environment:      "stage",
		SourceSessionKey: companyknowledge.NotionDocumentSessionKey("notion", "active_page"),
		HonchoWorkspace:  "rsi_company_knowledge",
		HonchoSessionID:  "notion_active_page",
		SourceRevision:   "rev1",
	}
	if _, err := store.ClaimSourceMirrorRecord(complete, time.Minute); err != nil {
		t.Fatalf("claim complete record: %v", err)
	}
	if _, err := store.CompleteSourceMirrorObject(complete.SourceType, complete.SourceKey, "document", "doc_active", nil); err != nil {
		t.Fatalf("complete record: %v", err)
	}
	stale := storepkg.SourceMirrorRecord{
		SourceType:       companyknowledge.NotionDocumentSourceType,
		SourceKey:        companyknowledge.NotionDocumentSourceKey("notion", "stale_page"),
		Workspace:        "notion",
		Environment:      "stage",
		SourceSessionKey: companyknowledge.NotionDocumentSessionKey("notion", "stale_page"),
		HonchoWorkspace:  "rsi_company_knowledge",
		HonchoSessionID:  "notion_stale_page",
		SourceRevision:   "stale:archived",
	}
	if _, err := store.MarkSourceMirrorRecordStale(stale, "archived", nil); err != nil {
		t.Fatalf("mark stale record: %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/source-mirror/status?source_type=notion_document", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		responseBody, _ := io.ReadAll(rec.Body)
		t.Fatalf("expected status 200, got %d: %s", rec.Code, string(responseBody))
	}
	var out sourceMirrorStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	status := out.SourceTypes[companyknowledge.NotionDocumentSourceType]
	if !out.OK || status.LatestComplete == nil || status.LatestStale == nil || status.Counts[storepkg.SourceMirrorStatusStale] != 1 {
		t.Fatalf("expected complete plus reportable stale record, got %+v", out)
	}
}

func TestSourceMirrorHealthPerformsSyntheticHonchoWrites(t *testing.T) {
	honchoMessagesCreated := 0
	honchoDocumentsCreated := 0
	honcho := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces":
			_, _ = w.Write([]byte(`{"id":"rsi_company_knowledge","metadata":{}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/sessions":
			_, _ = w.Write([]byte(`{"id":"source_mirror_health_stage","workspace_id":"rsi_company_knowledge","metadata":{}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/sessions/source_mirror_health_stage/messages":
			honchoMessagesCreated++
			_, _ = w.Write([]byte(`[{"id":"msg_health_1","content":"health","peer_id":"source_mirror_health"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/workspaces/rsi_company_knowledge/conclusions":
			honchoDocumentsCreated++
			_, _ = w.Write([]byte(`[{"id":"doc_health_1","content":"health","observer_id":"source_mirror_health","observed_id":"rsi_company_knowledge","session_id":"source_mirror_health_stage"}]`))
		default:
			t.Fatalf("unexpected honcho request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer honcho.Close()
	tmp := t.TempDir()
	store := storepkg.NewMemoryStore()
	report, err := CheckSourceMirrorHealth(context.Background(), config.Config{
		ServiceName:                "control-plane",
		Environment:                "stage",
		HonchoBaseURL:              honcho.URL,
		HonchoWorkspaceID:          "rsi_company_knowledge",
		SourceMirrorCheckpointRoot: tmp,
	}, store)
	if err != nil {
		t.Fatalf("CheckSourceMirrorHealth() error = %v", err)
	}
	if !report.OK || honchoMessagesCreated != 1 || honchoDocumentsCreated != 1 {
		t.Fatalf("unexpected source mirror health report=%+v messages=%d documents=%d", report, honchoMessagesCreated, honchoDocumentsCreated)
	}
	if _, err := os.Stat(filepath.Join(tmp, "health", "write-check.json")); err != nil {
		t.Fatalf("expected checkpoint root write-check: %v", err)
	}
}

func prepareProposalAttemptForWebhookTest(t *testing.T, store *storepkg.MemoryStore, proposal review.Proposal, mode string) (review.Proposal, improvement.ChangeAttempt) {
	t.Helper()

	proposal, ok := findProposalByID(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s", proposal.ID)
	}
	if proposal.CurrentAttemptID == "" {
		t.Fatalf("expected proposal %s to materialize a current attempt", proposal.ID)
	}
	attempt, ok := store.GetChangeAttempt(proposal.CurrentAttemptID)
	if !ok {
		t.Fatalf("expected attempt %s", proposal.CurrentAttemptID)
	}

	now := time.Now().UTC()
	commands := []transition.CommandEnvelope{
		{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposal.ID,
			CommandKind: string(transition.CommandProposalMarkRepoChangeQueued),
			CommandID:   "cmd-webhook-setup-queued:" + proposal.ID,
			Actor:       "tester",
			OccurredAt:  now,
		},
		{
			MachineKind: transition.MachineAttempt,
			AggregateID: attempt.ID,
			CommandKind: string(transition.CommandWorkspaceReady),
			CommandID:   "cmd-webhook-setup-workspace-ready:" + attempt.ID,
			Actor:       "tester",
			OccurredAt:  now.Add(time.Millisecond),
			Payload: map[string]any{
				"workspace_id":        "workspace-" + attempt.ID,
				"repo":                "rsi-agent-platform",
				"base_ref":            "main",
				"branch_name":         attempt.BranchName,
				"workspace_namespace": "rsi-platform",
				"workspace_job_name":  "workspace-job-" + attempt.ID,
			},
		},
		{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposal.ID,
			CommandKind: string(transition.CommandProposalMarkRepoChangeRunning),
			CommandID:   "cmd-webhook-setup-running:" + proposal.ID,
			Actor:       "tester",
			OccurredAt:  now.Add(2 * time.Millisecond),
		},
		{
			MachineKind: transition.MachineAttempt,
			AggregateID: attempt.ID,
			CommandKind: string(transition.CommandImplementationCompleted),
			CommandID:   "cmd-webhook-setup-implemented:" + attempt.ID,
			Actor:       "tester",
			OccurredAt:  now.Add(3 * time.Millisecond),
			Payload: map[string]any{
				"change_plan":     "Implement the approved remediation.",
				"validation_plan": "Run the governed validation flow.",
				"diff_summary":    "formal webhook setup",
				"changed_files":   []string{"internal/store/commands.go"},
			},
		},
		{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposal.ID,
			CommandKind: string(transition.CommandProposalMarkValidationPending),
			CommandID:   "cmd-webhook-setup-validation-pending:" + proposal.ID,
			Actor:       "tester",
			OccurredAt:  now.Add(4 * time.Millisecond),
		},
	}
	if mode == "pr_open" {
		commands = append(commands,
			transition.CommandEnvelope{
				MachineKind: transition.MachineAttempt,
				AggregateID: attempt.ID,
				CommandKind: string(transition.CommandAttemptPROpened),
				CommandID:   "cmd-webhook-setup-pr-opened:" + attempt.ID,
				Actor:       "tester",
				OccurredAt:  now.Add(5 * time.Millisecond),
				Payload: map[string]any{
					"pr_url":   "https://github.com/piplabs/rsi-agent-platform/pull/seed",
					"head_sha": "seedsha",
				},
			},
			transition.CommandEnvelope{
				MachineKind: transition.MachineProposalLine,
				AggregateID: proposal.ID,
				CommandKind: string(transition.CommandProposalMarkPROpen),
				CommandID:   "cmd-webhook-setup-proposal-pr-open:" + proposal.ID,
				Actor:       "tester",
				OccurredAt:  now.Add(6 * time.Millisecond),
			},
		)
	}

	for _, command := range commands {
		receipt, err := store.SubmitCommand(command)
		if err != nil {
			t.Fatalf("SubmitCommand(%s) error = %v", command.CommandKind, err)
		}
		if receipt.DecisionKind == transition.DecisionReject {
			t.Fatalf("SubmitCommand(%s) rejected: %s", command.CommandKind, receipt.Reason)
		}
	}

	proposal, ok = findProposalByID(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after setup", proposal.ID)
	}
	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s after setup", attempt.ID)
	}
	return proposal, attempt
}

func TestGitHubWebhookCreatesLinkedOutcome(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposalID := store.ListProposals()[0].ID
	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)

	payload := map[string]any{
		"action": "closed",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   42,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/42",
			"state":    "closed",
			"merged":   true,
			"title":    "RSI proposal " + proposalID + " for rsi-agent-platform",
			"head": map[string]any{
				"ref": "codex/" + proposalID,
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-1")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	outcomes := store.ListOutcomes()
	found := false
	for _, item := range outcomes {
		if item.Source == string(ingestion.SourceGitHub) && item.ProposalID == proposalID {
			found = true
			if item.Verdict != "positive" {
				t.Fatalf("expected positive verdict, got %s", item.Verdict)
			}
			if item.ExternalRef != "https://github.com/piplabs/rsi-agent-platform/pull/42" {
				t.Fatalf("unexpected external ref %s", item.ExternalRef)
			}
		}
	}
	if !found {
		t.Fatal("expected linked github outcome to be recorded")
	}
}

func TestEventAndIngestionRoutesSubmitIngressCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	eventReq := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader([]byte(`{
		"source":"system",
		"source_event_id":"manual-1",
		"dedupe_key":"manual-1",
		"severity":"warning",
		"normalized_problem_statement":"Investigate the latest deploy issue.",
		"workflow_hint":"question"
	}`)))
	eventRec := httptest.NewRecorder()
	router.ServeHTTP(eventRec, eventReq)
	if eventRec.Code != http.StatusCreated {
		t.Fatalf("event status = %d, want %d", eventRec.Code, http.StatusCreated)
	}

	ingestionReq := httptest.NewRequest(http.MethodPost, "/api/ingestions", bytes.NewReader([]byte(`{
		"bot_role":"orchestrator",
		"team_id":"T1",
		"channel_id":"C1",
		"thread_ts":"171000001.000100",
		"user_id":"U1",
		"text":"Can RSI summarize the latest deploy issue?",
		"ts":"171000001.000100"
	}`)))
	ingestionRec := httptest.NewRecorder()
	router.ServeHTTP(ingestionRec, ingestionReq)
	if ingestionRec.Code != http.StatusCreated {
		t.Fatalf("ingestion status = %d, want %d", ingestionRec.Code, http.StatusCreated)
	}

	foundEvent := false
	foundSlack := false
	for _, item := range store.ListDomainEvents() {
		switch item.EventKind {
		case "ingress_event_recorded":
			foundEvent = true
		case "ingress_slack_recorded":
			foundSlack = true
		}
	}
	if !foundEvent || !foundSlack {
		t.Fatalf("expected both ingress domain events, foundEvent=%t foundSlack=%t", foundEvent, foundSlack)
	}
}

func TestInternalActiveExecutionsBlocksDeploymentWhenPolicyRequiresIt(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-active",
		OperationID: "eff-active",
		WorkflowID:  "wf-active",
		TraceID:     "trace-active",
		CaseID:      "case-active",
		Status:      "running",
		HeartbeatAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{
		ServiceName:                     "control-plane",
		Environment:                     "development",
		DeploymentActiveExecutionPolicy: "block",
	}, store)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/internal/executions/active", nil))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := int(payload["active_execution_count"].(float64)); got != 1 {
		t.Fatalf("active_execution_count = %d, want 1", got)
	}
	if payload["deployment_policy"] != "block" {
		t.Fatalf("deployment_policy = %#v, want block", payload["deployment_policy"])
	}
}

func TestInternalActiveExecutionsReconcilesStaleCancellationBeforeDeployGate(t *testing.T) {
	store := storepkg.NewMemoryStore()
	stale := time.Now().Add(-5 * time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-stale-cancel",
		OperationID:     "eff-stale-cancel",
		WorkflowID:      "wf-stale-cancel",
		TraceID:         "trace-stale-cancel",
		CaseID:          "case-stale-cancel",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &stale,
		CreatedAt:       stale,
		UpdatedAt:       stale,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{
		ServiceName:                     "control-plane",
		Environment:                     "stage",
		DeploymentActiveExecutionPolicy: "block",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}, store)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/internal/executions/active", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := int(payload["active_execution_count"].(float64)); got != 0 {
		t.Fatalf("active_execution_count = %d, want 0", got)
	}
	record, ok := store.GetRunnerExecution("hexec-stale-cancel")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "cancelled" || record.CompletedAt == nil || record.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("stale cancellation should terminalize as cancelled before deploy gate, got %+v", record)
	}
}

func TestInternalActiveExecutionsReconcilesStaleCancellationWithoutHeartbeat(t *testing.T) {
	store := storepkg.NewMemoryStore()
	stale := time.Now().Add(-5 * time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-stale-cancel-no-heartbeat",
		WorkflowID:      "wf-stale-cancel-no-heartbeat",
		TraceID:         "trace-stale-cancel-no-heartbeat",
		CaseID:          "case-stale-cancel-no-heartbeat",
		Status:          "cancel_requested",
		CancelRequested: true,
		CreatedAt:       stale,
		UpdatedAt:       stale,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{
		ServiceName:                     "control-plane",
		Environment:                     "stage",
		DeploymentActiveExecutionPolicy: "block",
		HermesExecutionHeartbeatTimeout: time.Minute,
	}, store)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/internal/executions/active", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	record, ok := store.GetRunnerExecution("hexec-stale-cancel-no-heartbeat")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "cancelled" || record.CompletedAt == nil || record.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("stale cancellation without heartbeat should terminalize as cancelled before deploy gate, got %+v", record)
	}
}

func TestRunnerExecutionHeartbeatRejectsTerminalAndRequiresHolder(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-terminal",
		Status:      "completed",
		Result:      map[string]any{"ok": true},
		Holder:      "worker-1",
		HeartbeatAt: &now,
		CompletedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "development"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-terminal/heartbeat", bytes.NewReader([]byte(`{"holder":"worker-1","status":"running"}`)))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("terminal heartbeat status = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	record, ok := store.GetRunnerExecution("hexec-terminal")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "completed" || record.CompletedAt == nil || record.Result["ok"] != true {
		t.Fatalf("terminal record mutated: %+v", record)
	}
}

func TestRunnerExecutionHeartbeatMissingHolderIsBadRequest(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-missing-holder",
		Status:      "running",
		Holder:      "worker-1",
		HeartbeatAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "development"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-missing-holder/heartbeat", bytes.NewReader([]byte(`{"status":"running"}`)))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing heartbeat holder status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRunnerExecutionHeartbeatAcceptsSameHolderWithoutPriorHeartbeat(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-no-heartbeat",
		Status:      "running",
		Holder:      "worker-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "development"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-no-heartbeat/heartbeat", bytes.NewReader([]byte(`{"holder":"worker-1","status":"running"}`)))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("heartbeat status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	record, ok := store.GetRunnerExecution("hexec-no-heartbeat")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.HeartbeatAt == nil || record.Holder != "worker-1" || record.Status != "running" {
		t.Fatalf("heartbeat did not update same-holder record without prior heartbeat: %+v", record)
	}
}

func TestRunnerExecutionCompleteRejectsHolderMismatch(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-running",
		Status:      "running",
		Holder:      "worker-1",
		HeartbeatAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "development"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-running/complete", bytes.NewReader([]byte(`{"holder":"worker-2","status":"completed"}`)))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("holder mismatch status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	record, _ := store.GetRunnerExecution("hexec-running")
	if record.Status != "running" || record.CompletedAt != nil {
		t.Fatalf("holder mismatch mutated record: %+v", record)
	}
}

func TestRunnerExecutionCompleteMissingHolderIsBadRequest(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-complete-missing-holder",
		Status:      "running",
		Holder:      "worker-1",
		HeartbeatAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "development"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-complete-missing-holder/complete", bytes.NewReader([]byte(`{"status":"completed","result":{"ok":true}}`)))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing complete holder status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRunnerExecutionCompleteAcceptsSameHolderWithoutPriorHeartbeat(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-complete-no-heartbeat",
		Status:      "running",
		Holder:      "worker-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "development"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-complete-no-heartbeat/complete", bytes.NewReader([]byte(`{"holder":"worker-1","status":"completed","result":{"ok":true,"message":"done","provider":"hermes-executor","raw":{}}}`)))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("complete status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	record, ok := store.GetRunnerExecution("hexec-complete-no-heartbeat")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "completed" || record.CompletedAt == nil || record.Result["ok"] != true {
		t.Fatalf("complete did not terminalize same-holder record without prior heartbeat: %+v", record)
	}
}

func TestRunnerExecutionCompleteAfterCancellationNormalizesFailedStatusToCancelled(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-complete-failed",
		Status:          "cancel_requested",
		Holder:          "worker-1",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "development"}, store)

	body := []byte(`{
		"holder":"worker-1",
		"status":"failed",
		"failure_class":"worker_failed",
		"result":{"ok":false,"message":"worker failed after cancellation","provider":"hermes-executor","raw":{"failure_class":"worker_failed"}}
	}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-cancel-complete-failed/complete", bytes.NewReader(body))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("complete status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	record, ok := store.GetRunnerExecution("hexec-cancel-complete-failed")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "cancelled" || !record.CancelRequested || record.CompletedAt == nil || record.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("cancelled completion should dominate failed worker status, got %+v", record)
	}
	if record.Result["ok"] != false {
		t.Fatalf("audit result should be preserved, got %+v", record.Result)
	}
}

func TestValidateRunnerExecutionHolderMarksStaleTakeover(t *testing.T) {
	staleHeartbeat := time.Now().Add(-5 * time.Minute).UTC()
	cas, err := validateRunnerExecutionHolder(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		storepkg.RunnerExecution{
			ExecutionID: "hexec-stale",
			Status:      "running",
			Holder:      "hermes-executor:old",
			HeartbeatAt: &staleHeartbeat,
			CreatedAt:   staleHeartbeat,
			UpdatedAt:   staleHeartbeat,
		},
		"hermes-executor:new",
	)
	if err != nil {
		t.Fatalf("validateRunnerExecutionHolder() error = %v", err)
	}
	if cas.ExpectedHolder != "hermes-executor:old" {
		t.Fatalf("cas = %+v, want stale takeover from old holder", cas)
	}
}

func TestValidateRunnerExecutionHolderUsesCreatedAtWithoutHeartbeat(t *testing.T) {
	now := time.Now().UTC()
	createdAt := now.Add(-5 * time.Minute)
	_, err := validateRunnerExecutionHolder(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		storepkg.RunnerExecution{
			ExecutionID: "hexec-stale-created",
			Status:      "running",
			Holder:      "hermes-executor:old",
			CreatedAt:   createdAt,
			UpdatedAt:   now,
		},
		"hermes-executor:new",
	)
	if err == nil {
		t.Fatalf("validateRunnerExecutionHolder() expected holder mismatch error when no heartbeat reference available")
	}
}

func TestValidateRunnerExecutionHolderRejectsStartedAtOnlyTakeover(t *testing.T) {
	now := time.Now().UTC()
	startedAt := now.Add(-5 * time.Minute)
	_, err := validateRunnerExecutionHolder(
		config.Config{HermesExecutionHeartbeatTimeout: time.Minute},
		storepkg.RunnerExecution{
			ExecutionID: "hexec-stale-started",
			Status:      "running",
			Holder:      "hermes-executor:old",
			StartedAt:   &startedAt,
			CreatedAt:   startedAt,
			UpdatedAt:   now,
		},
		"hermes-executor:new",
	)
	if err == nil {
		t.Fatalf("validateRunnerExecutionHolder() expected holder mismatch when active execution has no heartbeat")
	}
}

func TestRunnerExecutionHeartbeatRejectsCancelRequestedWithoutMutation(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	heartbeat := now.Add(-30 * time.Second)
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-requested",
		Status:          "cancel_requested",
		Holder:          "worker-1",
		CancelRequested: true,
		HeartbeatAt:     &heartbeat,
		CreatedAt:       now.Add(-time.Minute),
		UpdatedAt:       heartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "development"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-cancel-requested/heartbeat", bytes.NewReader([]byte(`{"holder":"worker-1","status":"running"}`)))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("heartbeat status = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	record, ok := store.GetRunnerExecution("hexec-cancel-requested")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Holder != "worker-1" || record.Status != "cancel_requested" || record.HeartbeatAt == nil || record.HeartbeatAt.Before(heartbeat) {
		t.Fatalf("cancel-requested heartbeat should keep status but refresh heartbeat_at, got %+v", record)
	}
}

func TestRunnerExecutionHeartbeatRefreshesCancelRequested(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	heartbeat := now.Add(-30 * time.Second)
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-refresh",
		Status:          "cancel_requested",
		Holder:          "worker-1",
		CancelRequested: true,
		HeartbeatAt:     &heartbeat,
		CreatedAt:       now.Add(-time.Minute),
		UpdatedAt:       heartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/runner-executions/hexec-cancel-refresh/heartbeat", bytes.NewReader([]byte(`{"holder":"worker-1","status":"cancel_requested"}`)))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("heartbeat status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	record, ok := store.GetRunnerExecution("hexec-cancel-refresh")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "cancel_requested" || record.HeartbeatAt == nil || !record.HeartbeatAt.After(heartbeat) {
		t.Fatalf("cancel-requested heartbeat should refresh without changing lifecycle state, got %+v", record)
	}
}

func TestGitHubWebhookRejectsInvalidSignature(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-2")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized response, got %d", rec.Code)
	}
}

func TestGitHubWebhookIgnoresUnknownProposalWithoutWorkflow(t *testing.T) {
	store := storepkg.NewMemoryStore()
	initialOutcomes := len(store.ListOutcomes())
	initialTraces := len(store.ListTraces())
	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)

	payload := map[string]any{
		"action": "opened",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   43,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/43",
			"state":    "open",
			"merged":   false,
			"title":    "RSI proposal proposal-does-not-exist for rsi-agent-platform",
			"head": map[string]any{
				"ref": "codex/proposal-does-not-exist",
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-3")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	if len(store.ListOutcomes()) != initialOutcomes {
		t.Fatalf("expected no new linked outcomes, got %d -> %d", initialOutcomes, len(store.ListOutcomes()))
	}
	if len(store.ListTraces()) != initialTraces {
		t.Fatalf("expected no new workflow traces, got %d -> %d", initialTraces, len(store.ListTraces()))
	}
}

func TestThreadPolicyCommandsRouteMutatesViaFormalCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	req := httptest.NewRequest(http.MethodPost, "/api/thread-policies/slack:CENG:171000001.000100/commands", bytes.NewReader([]byte(`{"command_kind":"thread_mute","actor":"ui-operator"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	item, ok := findThreadPolicy(store.ListThreadPolicies(), "slack:CENG:171000001.000100")
	if !ok {
		t.Fatal("expected thread policy to exist")
	}
	if item.State != policy.ThreadStateMuted || !item.Muted {
		t.Fatalf("expected muted thread policy, got %+v", item)
	}
}

func TestLegacyThreadPolicyMutationRoutesAbsent(t *testing.T) {
	store := storepkg.NewMemoryStore()
	router := NewRouter(config.Config{ServiceName: "control-plane", Environment: "stage"}, store)

	for _, path := range []string{
		"/api/thread-policies/slack:CENG:171000001.000100/mute",
		"/api/thread-policies/slack:CENG:171000001.000100/resume",
	} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected %s to be absent, got %d", path, rec.Code)
		}
	}
}

func TestGitHubWebhookClosedPRQueuesRetryForCurrentAttempt(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for recursive remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	proposal, attempt := prepareProposalAttemptForWebhookTest(t, store, proposal, "pr_open")

	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)
	payload := map[string]any{
		"action": "closed",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   99,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/99",
			"state":    "closed",
			"merged":   false,
			"title":    "RSI proposal " + proposal.ID + " attempt " + attempt.ID + " for rsi-agent-platform",
			"head": map[string]any{
				"ref": attempt.BranchName,
				"sha": "deadbeef",
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-retry")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	updatedAttempt, ok := store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatal("expected updated attempt")
	}
	if updatedAttempt.FailureClass != "closed_unmerged" {
		t.Fatalf("expected closed_unmerged failure class, got %q", updatedAttempt.FailureClass)
	}
	if updatedAttempt.RetryDecision != "auto_retry" {
		t.Fatalf("expected auto_retry decision, got %q", updatedAttempt.RetryDecision)
	}
	if proposal, ok = findProposalByID(store.ListProposals(), proposal.ID); !ok {
		t.Fatal("expected proposal after webhook")
	}
	if proposal.Status != review.ProposalApproved {
		t.Fatalf("expected proposal to remain approved for retry, got %s", proposal.Status)
	}
	if proposal.CurrentAttemptID == "" || proposal.CurrentAttemptID == attempt.ID {
		t.Fatalf("expected retry to materialize a successor attempt, got %q", proposal.CurrentAttemptID)
	}
	foundRetry := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != proposal.CurrentAttemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			foundRetry = true
		default:
			t.Fatalf("unexpected retry bootstrap effect %s", effect.EffectKind)
		}
	}
	if !foundRetry {
		t.Fatal("expected queued retry effect for the successor attempt")
	}
}

func TestGitHubWebhookOpenedPRRoutesProposalThroughCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for recursive remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	proposal, attempt := prepareProposalAttemptForWebhookTest(t, store, proposal, "validation_pending")

	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)
	payload := map[string]any{
		"action": "opened",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   77,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/77",
			"state":    "open",
			"merged":   false,
			"title":    "RSI proposal " + proposal.ID + " attempt " + attempt.ID + " for rsi-agent-platform",
			"head": map[string]any{
				"ref": attempt.BranchName,
				"sha": "abc123",
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-open")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	updatedAttempt, ok := store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatal("expected updated attempt")
	}
	if updatedAttempt.State != improvement.AttemptStateCIObserving {
		t.Fatalf("expected attempt state ci_observing, got %s", updatedAttempt.State)
	}
	if updatedAttempt.PRURL != "https://github.com/piplabs/rsi-agent-platform/pull/77" {
		t.Fatalf("expected pr url to persist, got %q", updatedAttempt.PRURL)
	}
	if proposal, ok = findProposalByID(store.ListProposals(), proposal.ID); !ok {
		t.Fatal("expected proposal after webhook")
	}
	if proposal.Status != review.ProposalPROpen {
		t.Fatalf("expected proposal to move to pr_open, got %s", proposal.Status)
	}
	receipt, ok := store.GetCommandReceipt("cmd-github-proposal:" + proposal.ID + ":" + string(transition.CommandProposalMarkPROpen) + ":delivery-open")
	if !ok {
		t.Fatal("expected proposal command receipt for pr_open outcome")
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected pr_open command to advance, got %s", receipt.DecisionKind)
	}
}

func TestGitHubWebhookMergedPRRoutesProposalThroughCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	if _, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Approved for recursive remediation.",
		ReviewerID: "operator",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	proposal, attempt := prepareProposalAttemptForWebhookTest(t, store, proposal, "pr_open")

	cfg := config.Config{
		ServiceName:         "control-plane",
		Environment:         "stage",
		GitHubWebhookSecret: "secret",
	}
	router := NewRouter(cfg, store)
	payload := map[string]any{
		"action": "closed",
		"repository": map[string]any{
			"full_name": "piplabs/rsi-agent-platform",
			"name":      "rsi-agent-platform",
		},
		"sender": map[string]any{
			"login": "blake",
		},
		"pull_request": map[string]any{
			"number":   78,
			"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/78",
			"state":    "closed",
			"merged":   true,
			"title":    "RSI proposal " + proposal.ID + " attempt " + attempt.ID + " for rsi-agent-platform",
			"head": map[string]any{
				"ref": attempt.BranchName,
				"sha": "def456",
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-merged")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhook(body, cfg.GitHubWebhookSecret))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", rec.Code)
	}
	updatedAttempt, ok := store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatal("expected updated attempt")
	}
	if updatedAttempt.State != improvement.AttemptStateMerged {
		t.Fatalf("expected attempt state merged, got %s", updatedAttempt.State)
	}
	if proposal, ok = findProposalByID(store.ListProposals(), proposal.ID); !ok {
		t.Fatal("expected proposal after webhook")
	}
	if proposal.Status != review.ProposalMerged {
		t.Fatalf("expected proposal to move to merged, got %s", proposal.Status)
	}
	receipt, ok := store.GetCommandReceipt("cmd-github-proposal:" + proposal.ID + ":" + string(transition.CommandProposalMarkMerged) + ":delivery-merged")
	if !ok {
		t.Fatal("expected proposal command receipt for merged outcome")
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected merged command to advance, got %s", receipt.DecisionKind)
	}
}

func findProposalByID(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}

func findThreadPolicy(items []policy.ThreadPolicy, threadKey string) (policy.ThreadPolicy, bool) {
	for _, item := range items {
		if item.ThreadKey == threadKey {
			return item, true
		}
	}
	return policy.ThreadPolicy{}, false
}

func signGitHubWebhook(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
