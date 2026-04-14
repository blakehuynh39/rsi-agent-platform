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
	"github.com/piplabs/rsi-agent-platform/internal/review"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
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
	if _, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "alice",
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	job, err := store.MaterializeApprovedProposal(proposal.ID, "alice")
	if err != nil {
		t.Fatalf("materialize proposal: %v", err)
	}
	if _, err := store.UpdateRepoChangeJobStatus(job.ID, string(review.ProposalFailedValidation)); err != nil {
		t.Fatalf("update job status: %v", err)
	}
	if _, err := store.UpdateProposalStatus(proposal.ID, review.ProposalFailedValidation); err != nil {
		t.Fatalf("update proposal status: %v", err)
	}

	item, err := store.RetryProposalRepoChange(proposal.ID, "alice")
	if err != nil {
		t.Fatalf("RetryProposalRepoChange() error = %v", err)
	}
	if item.Queue != "sandbox" {
		t.Fatalf("expected sandbox work item, got %+v", item)
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

	jobs := reloaded.ListRepoChangeJobs()
	if len(jobs) == 0 || jobs[0].Status != string(review.ProposalRepoChangeQueued) {
		t.Fatalf("expected repo change job persisted in repo_change_queued, got %+v", jobs)
	}
	found := false
	for _, queued := range reloaded.ListWorkItems() {
		if queued.ID == item.ID && queued.Queue == "sandbox" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected retried sandbox work item %s to persist", item.ID)
	}
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

	ingested, err := store.CreateIngestion(slackpkg.SlackEnvelope{
		BotRole:   slackpkg.BotArch,
		TeamID:    "T123",
		ChannelID: "D123",
		ThreadTS:  "171000001.000100",
		UserID:    "U123",
		Text:      "RSI is failing because Postgres persistence wedged the recursive loop.",
		TS:        "171000001.000100",
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CreateIngestion() error = %v", err)
	}
	if ingested.ID == "" {
		t.Fatal("expected ingestion to be created")
	}

	items := store.ListWorkItems()
	if len(items) == 0 {
		t.Fatal("expected workflow work item")
	}
	traceID := items[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}

	description := `subsystem=shared-store failure_mode=action_result_primary_key_collision provider=github action_intent_id=action-002 work_item_id=work-003 kind=tool_read sqlstate=23505 constraint=action_result_pkey table=action_result error="duplicate key value violates unique constraint \"action_result_pkey\""`
	if _, err := store.ApplyTraceUpdate(traceID, TraceUpdate{
		Status:         statusPtr(events.StatusNeedsHuman),
		WorkflowStatus: "needs-human",
		WorkflowError:  description,
		Events: []events.TraceEvent{
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
				StartedAt:      time.Now().UTC(),
				Description:    description,
			},
		},
	}); err != nil {
		t.Fatalf("ApplyTraceUpdate() error = %v", err)
	}

	run, judgments, err := store.EvaluateTrace(traceID, "integration")
	if err != nil {
		t.Fatalf("EvaluateTrace() error = %v", err)
	}

	result, err := store.RunProposalPromoter("integration-test")
	if err != nil {
		t.Fatalf("RunProposalPromoter() error = %v", err)
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
