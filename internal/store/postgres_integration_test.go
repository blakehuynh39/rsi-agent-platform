package store

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func TestPostgresEvaluateTraceCreatesCandidateAndPromotableProposal(t *testing.T) {
	postgresURL, cleanup := openTempPostgresURL(t)
	defer cleanup()

	db, err := platformdb.OpenPostgres(postgresURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if _, err := platformdb.ApplyMigrations(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	store, err := NewPostgresStore(config.Config{
		StoreBackend:       "postgres",
		PostgresURL:        postgresURL,
		DefaultProposalCap: 2,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer store.db.Close()

	trace, run, judgments, proposal := seedPromotableFailureProposal(t, store)
	if proposal.ID == "" {
		t.Fatal("expected proposal after promotion")
	}
	if run.TraceID != trace.Summary.TraceID {
		t.Fatalf("expected eval run for trace %s, got %+v", trace.Summary.TraceID, run)
	}
	if len(judgments) == 0 {
		t.Fatalf("expected eval judgments for trace %s", trace.Summary.TraceID)
	}

	candidates := store.ListCandidates()
	if len(candidates) == 0 {
		t.Fatal("expected candidate to be created")
	}
	if candidates[0].FailureMode != "action_result_primary_key_collision" {
		t.Fatalf("unexpected candidate failure mode: %+v", candidates[0])
	}
}

func TestPostgresRetryProposalRepoChangePersistsSandboxRequeue(t *testing.T) {
	postgresURL, cleanup := openTempPostgresURL(t)
	defer cleanup()

	db, err := platformdb.OpenPostgres(postgresURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if _, err := platformdb.ApplyMigrations(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	store, err := NewPostgresStore(config.Config{
		StoreBackend:       "postgres",
		PostgresURL:        postgresURL,
		DefaultProposalCap: 2,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer store.db.Close()

	_, _, _, proposal := seedPromotableFailureProposal(t, store)
	approved, err := ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	if strings.TrimSpace(approved.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt after approval, got %+v", approved)
	}
	now := time.Now().UTC()
	if _, err := SeedRepoChangeJobForTesting(store, improvement.RepoChangeJob{
		ID:               "job-postgres-retry-1",
		ProposalID:       proposal.ID,
		AttemptID:        approved.CurrentAttemptID,
		ConversationID:   approved.ConversationID,
		CaseID:           approved.CaseID,
		OriginTraceID:    firstNonEmpty(approved.OriginTraceID, approved.TraceID),
		CandidateKey:     approved.CandidateKey,
		Status:           string(review.ProposalFailedValidation),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       "codex/postgres-retry",
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		t.Fatalf("upsert repo change job: %v", err)
	}
	submitProposalCommandForTest(t, store, proposal.ID, transition.CommandProposalMarkFailedValidation, "cmd-postgres-integration-proposal-failed-validation", nil)

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposal.ID,
		CommandKind: string(transition.CommandProposalRetryAttempt),
		CommandID:   "cmd-postgres-integration-retry",
		Actor:       "alice",
		OccurredAt:  now.Add(time.Second),
		Payload: map[string]any{
			"reviewer_id": "alice",
			"scope":       string(review.FeedbackScopeLine),
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(proposal_retry_attempt) error = %v", err)
	}
	if receipt.DecisionKind == transition.DecisionReject {
		t.Fatalf("expected retry command accepted, got %+v", receipt)
	}

	reloaded, err := NewPostgresStore(config.Config{
		StoreBackend:       "postgres",
		PostgresURL:        postgresURL,
		DefaultProposalCap: 2,
	})
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer reloaded.db.Close()

	reloadedProposal, ok := findProposalByIDInPostgresIntegration(reloaded.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after retry", proposal.ID)
	}
	if strings.TrimSpace(reloadedProposal.CurrentAttemptID) == "" || reloadedProposal.CurrentAttemptID == approved.CurrentAttemptID {
		t.Fatalf("expected retry to materialize a new current attempt, got %+v", reloadedProposal)
	}
	foundEffect := false
	for _, effect := range reloaded.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != reloadedProposal.CurrentAttemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			foundEffect = true
		default:
			t.Fatalf("unexpected retry bootstrap effect %s", effect.EffectKind)
		}
	}
	if !foundEffect {
		t.Fatalf("expected queued retry bootstrap effect for attempt %s", reloadedProposal.CurrentAttemptID)
	}
}

func findProposalByIDInPostgresIntegration(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}

func TestPostgresRunProposalPromoterNormalizesBlankTargetFields(t *testing.T) {
	postgresURL, cleanup := openTempPostgresURL(t)
	defer cleanup()

	db, err := platformdb.OpenPostgres(postgresURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if _, err := platformdb.ApplyMigrations(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	store, err := NewPostgresStore(config.Config{
		StoreBackend:       "postgres",
		PostgresURL:        postgresURL,
		DefaultProposalCap: 2,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer store.db.Close()

	_, _, _, proposal := seedPromotableFailureProposal(t, store)
	if proposal.ID == "" {
		t.Fatal("expected promoted proposal")
	}

	err = store.withLoadedStoreTx(func(tx *sql.Tx, loaded *MemoryStore) error {
		for key, candidate := range loaded.candidates {
			candidate.TargetLayer = ""
			candidate.TargetKind = ""
			candidate.TargetRef = ""
			loaded.candidates[key] = candidate
		}
		for key, current := range loaded.proposals {
			current.TargetLayer = ""
			current.TargetKind = ""
			current.TargetRef = ""
			loaded.proposals[key] = current
		}
		return replaceProposalPromoterScope(tx, loaded)
	})
	if err != nil {
		t.Fatalf("replaceProposalPromoterScope() error = %v", err)
	}

	for _, candidate := range store.ListCandidates() {
		if strings.TrimSpace(candidate.TargetRef) == "" {
			t.Fatalf("expected candidate target_ref to be normalized, got %+v", candidate)
		}
	}
	for _, current := range store.ListProposals() {
		if strings.TrimSpace(current.TargetRef) == "" {
			t.Fatalf("expected proposal target_ref to be normalized, got %+v", current)
		}
	}
}

func openTempPostgresURL(t *testing.T) (string, func()) {
	t.Helper()
	baseURL := strings.TrimSpace(os.Getenv("RSI_TEST_POSTGRES_URL"))
	if baseURL == "" {
		t.Skip("RSI_TEST_POSTGRES_URL not set")
	}
	admin, err := platformdb.OpenPostgres(baseURL)
	if err != nil {
		t.Fatalf("open admin postgres: %v", err)
	}
	dbName := fmt.Sprintf("rsi_store_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(`create database ` + dbName); err != nil {
		_ = admin.Close()
		t.Fatalf("create database %s: %v", dbName, err)
	}
	testURL, err := withStoreDatabase(baseURL, dbName)
	if err != nil {
		_, _ = admin.Exec(`drop database if exists ` + dbName + ` with (force)`)
		_ = admin.Close()
		t.Fatalf("build database URL: %v", err)
	}
	return testURL, func() {
		_, _ = admin.Exec(`drop database if exists ` + dbName + ` with (force)`)
		_ = admin.Close()
	}
}

func withStoreDatabase(rawURL string, dbName string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.Path = "/" + dbName
	return parsed.String(), nil
}

func seedPromotableFailureProposal(t *testing.T, store *PostgresStore) (events.Trace, evals.Run, []evals.Judgment, review.Proposal) {
	t.Helper()

	receipt := submitIngressCommandForTest(t, store, "slack:171000001.000100", transition.CommandIngressRecordSlack, "cmd-postgres-seed-slack-ingress", "tester", time.Now().UTC(), map[string]any{
		"bot_role":   string(slackpkg.BotArch),
		"team_id":    "T123",
		"channel_id": "D123",
		"thread_ts":  "171000001.000100",
		"user_id":    "U123",
		"text":       "RSI is failing because Postgres persistence wedged the recursive loop.",
		"ts":         "171000001.000100",
		"created_at": time.Now().UTC(),
	})
	if receipt.ResultRef == "" {
		t.Fatal("expected ingestion to be created")
	}

	traces := store.ListTraces()
	if len(traces) == 0 {
		t.Fatal("expected workflow trace after ingress")
	}
	traceID := traces[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}

	description := `subsystem=shared-store failure_mode=action_result_primary_key_collision provider=github action_intent_id=action-002 effect_execution_id=eff-003 kind=tool_read sqlstate=23505 constraint=action_result_pkey table=action_result error="duplicate key value violates unique constraint \"action_result_pkey\""`
	projectedAt := time.Now().UTC()
	if receipt := submitProblemLineCommandForTest(t, store, traceID, transition.CommandProblemLineProjectTrace, "cmd-postgres-seed-project-trace", "integration", projectedAt, map[string]any{
		"trace_id":        traceID,
		"trace_status":    string(events.StatusNeedsHuman),
		"workflow_status": "needs-human",
		"workflow_error":  description,
		"trace_events": []events.TraceEvent{
			{
				TraceID:        trace.Summary.TraceID,
				IngestionID:    trace.Summary.IngestionID,
				WorkflowID:     trace.Summary.WorkflowID,
				ConversationID: trace.Summary.ConversationID,
				CaseID:         trace.Summary.CaseID,
				TriggerEventID: trace.Summary.TriggerEventID,
				Plane:          "control",
				Service:        "control-plane",
				Actor:          "action-worker",
				EventType:      "action.persistence_failed",
				Status:         events.StatusNeedsHuman,
				StartedAt:      projectedAt,
				Description:    description,
			},
		},
	}); receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	evalReceipt := submitProblemLineCommandForTest(t, store, traceID, transition.CommandProblemLineEvaluateTrace, "cmd-postgres-seed-evaluate-trace", "integration", time.Now().UTC(), map[string]any{
		"trigger": "integration",
	})
	run, judgments, ok := findEvalRunForReceipt(store, evalReceipt)
	if !ok {
		t.Fatalf("expected eval run %s", evalReceipt.ResultRef)
	}

	promoteReceipt := submitProblemLineCommandForTest(t, store, "integration-test", transition.CommandProblemLinePromote, "cmd-postgres-seed-promote", "integration-test", time.Now().UTC(), map[string]any{
		"requested_by": "integration-test",
	})
	result, err := loadPromotionResultForReceipt(store, promoteReceipt)
	if err != nil {
		t.Fatalf("loadPromotionResultForReceipt() error = %v", err)
	}
	if result.Promoted == 0 {
		t.Fatalf("expected promoted proposal, got %+v", result)
	}
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected proposal after promotion")
	}
	return trace, run, judgments, proposals[0]
}
