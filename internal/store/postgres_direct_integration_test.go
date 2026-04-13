package store

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

func TestPostgresWorkItemDirectPersistence(t *testing.T) {
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
		StoreBackend: "postgres",
		PostgresURL:  postgresURL,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer store.db.Close()

	item, err := store.EnqueueWorkItem(queue.WorkItem{
		ID:          "work-custom",
		Queue:       queue.EvalQueue,
		Kind:        "manual_eval",
		Status:      queue.WorkQueued,
		TraceID:     "trace-custom",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		RequestedBy: "tester",
	})
	if err != nil {
		t.Fatalf("EnqueueWorkItem() error = %v", err)
	}
	claimed, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.EvalQueue}, "tester", 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("ClaimNextWorkItem() ok=%t err=%v", ok, err)
	}
	if claimed.ID != item.ID {
		t.Fatalf("claimed wrong item: %+v", claimed)
	}
	completed, err := store.CompleteWorkItem(item.ID)
	if err != nil {
		t.Fatalf("CompleteWorkItem() error = %v", err)
	}
	if completed.Status != queue.WorkCompleted {
		t.Fatalf("expected completed work item, got %+v", completed)
	}
}

func TestPostgresMaterializeApprovedProposalPersistsRepoChangeJob(t *testing.T) {
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
		DefaultProposalCap: 3,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer store.db.Close()

	_, _, _, proposal := seedPromotableFailureProposal(t, store)
	reviewed, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		ProposalID: proposal.ID,
		Decision:   string(review.ProposalApproved),
		Rationale:  "Allow repo-change materialization.",
		ReviewerID: "alice",
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if reviewed.Status != review.ProposalApproved {
		t.Fatalf("expected approved proposal, got %+v", reviewed)
	}

	job, err := store.MaterializeApprovedProposal(proposal.ID, "alice")
	if err != nil {
		t.Fatalf("MaterializeApprovedProposal() error = %v", err)
	}
	if job.ID == "" || job.ProposalID != proposal.ID {
		t.Fatalf("expected repo change job for proposal %s, got %+v", proposal.ID, job)
	}

	jobs := store.ListRepoChangeJobs()
	if len(jobs) != 1 || jobs[0].ID != job.ID {
		t.Fatalf("expected persisted repo change job, got %+v", jobs)
	}

	var persisted review.Proposal
	for _, item := range store.ListProposals() {
		if item.ID == proposal.ID {
			persisted = item
			break
		}
	}
	if persisted.Status != review.ProposalRepoChangeQueued {
		t.Fatalf("expected proposal to advance to repo_change_queued, got %+v", persisted)
	}

	foundSandboxWork := false
	for _, item := range store.ListWorkItems() {
		if item.Queue == queue.SandboxQueue && item.ProposalID == proposal.ID {
			foundSandboxWork = true
			break
		}
	}
	if !foundSandboxWork {
		t.Fatalf("expected sandbox work item for proposal %s", proposal.ID)
	}
}

func TestPostgresGitHubEventPersistsProposalOutcome(t *testing.T) {
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

	trace, _, _, proposal := seedPromotableFailureProposal(t, store)
	created, err := store.CreateEvent(ingestion.EventEnvelope{
		Source:        ingestion.SourceGitHub,
		SourceEventID: "delivery-123",
		DedupeKey:     "github:delivery-123",
		Severity:      ingestion.SeverityWarning,
		Metadata: map[string]interface{}{
			"event_type":  "pull_request",
			"action":      "closed",
			"merged":      "true",
			"proposal_id": proposal.ID,
			"html_url":    "https://github.com/piplabs/rsi-agent-platform/pull/321",
		},
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected github event to be persisted")
	}

	foundOutcome := false
	for _, item := range store.ListOutcomes() {
		if item.SourceEventID != "delivery-123" {
			continue
		}
		foundOutcome = true
		if item.ProposalID != proposal.ID || item.CaseID != proposal.CaseID || item.TraceID != trace.Summary.TraceID {
			t.Fatalf("unexpected linked outcome: %+v", item)
		}
		if item.OutcomeType != outcome.TypeProposalEffectiveness || item.Verdict != outcome.VerdictPositive {
			t.Fatalf("unexpected github outcome mapping: %+v", item)
		}
	}
	if !foundOutcome {
		t.Fatal("expected github outcome record")
	}

	caseLinked := false
	for _, item := range store.ListCases() {
		if item.ID != proposal.CaseID {
			continue
		}
		caseLinked = true
		if item.LatestOutcomeID == "" || item.ResolutionState == "" {
			t.Fatalf("expected case outcome linkage, got %+v", item)
		}
	}
	if !caseLinked {
		t.Fatalf("expected case %s to remain linked", proposal.CaseID)
	}

	proposalUpdated := false
	for _, item := range store.ListProposals() {
		if item.ID != proposal.ID {
			continue
		}
		proposalUpdated = true
		if !item.NewEvidenceSinceLastRejection {
			t.Fatalf("expected proposal %s to record new evidence after merge outcome", proposal.ID)
		}
	}
	if !proposalUpdated {
		t.Fatalf("expected proposal %s to remain queryable", proposal.ID)
	}
}

func TestPostgresExecuteToolFallbackPersistsPRAttempt(t *testing.T) {
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
		ProposalID: proposal.ID,
		Decision:   string(review.ProposalApproved),
		Rationale:  "Allow draft PR fallback execution.",
		ReviewerID: "alice",
		CreatedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}

	result := store.ExecuteTool("github.create_pr", map[string]interface{}{"proposal_id": proposal.ID})
	if !result.Approved {
		t.Fatalf("expected fallback tool execution to succeed, got %+v", result)
	}

	attempts := store.ListPRAttempts()
	if len(attempts) != 1 || attempts[0].ProposalID != proposal.ID {
		t.Fatalf("expected persisted PR attempt for proposal %s, got %+v", proposal.ID, attempts)
	}
}

func TestPostgresClaimNextWorkItemIsAtomic(t *testing.T) {
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

	storeA, err := NewPostgresStore(config.Config{
		StoreBackend: "postgres",
		PostgresURL:  postgresURL,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer storeA.db.Close()

	storeB, err := NewPostgresStore(config.Config{
		StoreBackend: "postgres",
		PostgresURL:  postgresURL,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer storeB.db.Close()

	item, err := storeA.EnqueueWorkItem(queue.WorkItem{
		Queue:       queue.EvalQueue,
		Kind:        "manual_eval",
		Status:      queue.WorkQueued,
		TraceID:     "trace-claim-atomic",
		RequestedBy: "tester",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("EnqueueWorkItem() error = %v", err)
	}

	type result struct {
		item queue.WorkItem
		ok   bool
		err  error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	claim := func(store *PostgresStore, holder string) {
		<-start
		claimed, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.EvalQueue}, holder, 30*time.Second)
		results <- result{item: claimed, ok: ok, err: err}
	}
	go claim(storeA, "worker-a")
	go claim(storeB, "worker-b")
	close(start)

	first := <-results
	second := <-results
	for _, item := range []result{first, second} {
		if item.err != nil {
			t.Fatalf("ClaimNextWorkItem() error = %v", item.err)
		}
	}

	successes := 0
	for _, claim := range []result{first, second} {
		if claim.ok {
			successes++
			if claim.item.ID != item.ID {
				t.Fatalf("claimed wrong work item: %+v", claim.item)
			}
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one successful claim, got %d", successes)
	}
}

func TestPostgresReviewProposalIsIdempotent(t *testing.T) {
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

	storeA, err := NewPostgresStore(config.Config{
		StoreBackend:       "postgres",
		PostgresURL:        postgresURL,
		DefaultProposalCap: 2,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer storeA.db.Close()

	storeB, err := NewPostgresStore(config.Config{
		StoreBackend:       "postgres",
		PostgresURL:        postgresURL,
		DefaultProposalCap: 2,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer storeB.db.Close()

	_, _, _, proposal := seedPromotableFailureProposal(t, storeA)
	decision := review.ProposalReview{
		ProposalID: proposal.ID,
		Decision:   string(review.ProposalApproved),
		Rationale:  "Allow repo-change materialization.",
		ReviewerID: "alice",
		CreatedAt:  time.Now().UTC(),
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := storeA.ReviewProposal(proposal.ID, decision)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := storeB.ReviewProposal(proposal.ID, decision)
		errs <- err
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("ReviewProposal() error = %v", err)
		}
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

	var persisted review.Proposal
	for _, item := range reloaded.ListProposals() {
		if item.ID == proposal.ID {
			persisted = item
			break
		}
	}
	if len(persisted.Reviews) != 1 {
		t.Fatalf("expected one persisted review, got %+v", persisted.Reviews)
	}
	memories := 0
	for _, item := range reloaded.ListProposalMemories() {
		if item.ProposalID == proposal.ID {
			memories++
		}
	}
	if memories != 1 {
		t.Fatalf("expected one proposal memory entry, got %d", memories)
	}
	workItems := 0
	for _, item := range reloaded.ListWorkItems() {
		if item.ProposalID == proposal.ID && item.Queue == queue.ProposalQueue {
			workItems++
		}
	}
	if workItems != 1 {
		t.Fatalf("expected one proposal materialization work item, got %d", workItems)
	}
}

func TestPostgresRetryProposalRepoChangeIsIdempotent(t *testing.T) {
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

	storeA, err := NewPostgresStore(config.Config{
		StoreBackend:       "postgres",
		PostgresURL:        postgresURL,
		DefaultProposalCap: 2,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer storeA.db.Close()

	storeB, err := NewPostgresStore(config.Config{
		StoreBackend:       "postgres",
		PostgresURL:        postgresURL,
		DefaultProposalCap: 2,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer storeB.db.Close()

	_, _, _, proposal := seedPromotableFailureProposal(t, storeA)
	if _, err := storeA.ReviewProposal(proposal.ID, review.ProposalReview{
		ProposalID: proposal.ID,
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "alice",
		CreatedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	job, err := storeA.MaterializeApprovedProposal(proposal.ID, "alice")
	if err != nil {
		t.Fatalf("materialize proposal: %v", err)
	}
	if _, err := storeA.UpdateRepoChangeJobStatus(job.ID, string(review.ProposalFailedValidation)); err != nil {
		t.Fatalf("update job status: %v", err)
	}
	if _, err := storeA.UpdateProposalStatus(proposal.ID, review.ProposalFailedValidation); err != nil {
		t.Fatalf("update proposal status: %v", err)
	}

	var wg sync.WaitGroup
	type retryResult struct {
		item queue.WorkItem
		err  error
	}
	results := make(chan retryResult, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		item, err := storeA.RetryProposalRepoChange(proposal.ID, "alice")
		results <- retryResult{item: item, err: err}
	}()
	go func() {
		defer wg.Done()
		item, err := storeB.RetryProposalRepoChange(proposal.ID, "alice")
		results <- retryResult{item: item, err: err}
	}()
	wg.Wait()
	close(results)

	var ids []string
	for result := range results {
		if result.err != nil {
			t.Fatalf("RetryProposalRepoChange() error = %v", result.err)
		}
		ids = append(ids, result.item.ID)
	}
	if len(ids) != 2 || ids[0] != ids[1] {
		t.Fatalf("expected both retry calls to converge on the same sandbox work item, got %v", ids)
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

	sandboxItems := 0
	for _, item := range reloaded.ListWorkItems() {
		if item.ProposalID == proposal.ID && item.Queue == queue.SandboxQueue {
			sandboxItems++
		}
	}
	if sandboxItems != 1 {
		t.Fatalf("expected one sandbox work item after retry dedupe, got %d", sandboxItems)
	}
}

func TestPostgresRecordActionResultStillSurfacesDuplicateKey(t *testing.T) {
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
		StoreBackend: "postgres",
		PostgresURL:  postgresURL,
	})
	if err != nil {
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	defer store.db.Close()

	intent, err := store.UpsertActionIntent(action.Intent{
		ID:         "intent-duplicate-action-result",
		OwnerPlane: "control",
		TraceID:    "trace-duplicate-action-result",
		Kind:       action.KindToolRead,
		Status:     action.StatusExecuting,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("UpsertActionIntent() error = %v", err)
	}

	result := action.Result{
		ID:             "action-result-001",
		ActionIntentID: intent.ID,
		AttemptNumber:  1,
		Executor:       "tester",
		Status:         action.StatusSucceeded,
		StartedAt:      time.Now().UTC(),
		CompletedAt:    time.Now().UTC(),
	}
	if _, err := store.RecordActionResult(result); err != nil {
		t.Fatalf("first RecordActionResult() error = %v", err)
	}
	if _, err := store.RecordActionResult(result); err == nil {
		t.Fatal("expected duplicate action_result primary key error")
	} else if !strings.Contains(err.Error(), "action_result_pkey") {
		t.Fatalf("expected action_result_pkey error, got %v", err)
	}
}
