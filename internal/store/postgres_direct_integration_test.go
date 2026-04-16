package store

import (
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
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

func TestPostgresEnqueueWorkItemReusesFailedOperationScopedItem(t *testing.T) {
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

	now := time.Now().UTC()
	op, _, err := store.GetOrCreateOperation(operation.Execution{
		ID:            "op-requeue",
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       "attempt-1",
		OperationKind: "sandbox_launch",
		OperationKey:  "sandbox_launch",
		Status:        operation.StatusQueued,
		Queue:         queue.SandboxQueue,
		RequestedBy:   "tester",
		TraceID:       "trace-1",
		ProposalID:    "proposal-1",
		AttemptID:     "attempt-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		t.Fatalf("GetOrCreateOperation() error = %v", err)
	}
	first, err := store.EnqueueWorkItem(queue.WorkItem{
		ID:          "work-op-requeue",
		OperationID: op.ID,
		Queue:       queue.SandboxQueue,
		Kind:        "repo_change_job",
		Status:      queue.WorkQueued,
		ProposalID:  "proposal-1",
		TraceID:     "trace-1",
		Payload:     map[string]any{"phase": 1},
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("EnqueueWorkItem() error = %v", err)
	}
	if _, err := store.FailWorkItem(first.ID, "temporary failure"); err != nil {
		t.Fatalf("FailWorkItem() error = %v", err)
	}
	second, err := store.EnqueueWorkItem(queue.WorkItem{
		ID:          "work-op-requeue-new",
		OperationID: op.ID,
		Queue:       queue.SandboxQueue,
		Kind:        "repo_change_job",
		Status:      queue.WorkQueued,
		ProposalID:  "proposal-1",
		TraceID:     "trace-1",
		Payload:     map[string]any{"phase": 2},
		CreatedAt:   now.Add(time.Second),
		UpdatedAt:   now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("second EnqueueWorkItem() error = %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected operation-backed work item reuse, got %s want %s", second.ID, first.ID)
	}
	reloadedOp, ok := store.GetOperation(op.ID)
	if !ok {
		t.Fatal("expected operation to remain present")
	}
	if reloadedOp.Status != operation.StatusQueued {
		t.Fatalf("expected requeued operation status, got %s", reloadedOp.Status)
	}
}

func TestPostgresUpsertAttemptWorkspacePersistsBlankDiffSummary(t *testing.T) {
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

	now := time.Now().UTC()
	item, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-1",
		AttemptID:        "attempt-1",
		ProposalID:       "proposal-1",
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       "codex/proposal-1/attempt-01",
		Namespace:        "rsi-platform",
		JobName:          "job-1",
		Status:           improvement.WorkspaceQueued,
		AllowedPathGlobs: []string{"internal/**"},
		DiffSummary:      "",
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	if item.DiffSummary != "" {
		t.Fatalf("expected blank diff summary, got %q", item.DiffSummary)
	}
	persisted, ok := store.GetAttemptWorkspaceByAttempt("attempt-1")
	if !ok {
		t.Fatal("expected workspace to be persisted")
	}
	if persisted.DiffSummary != "" {
		t.Fatalf("expected persisted blank diff summary, got %q", persisted.DiffSummary)
	}
}

func TestPostgresRescheduleWorkItemRequeuesOperation(t *testing.T) {
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

	now := time.Now().UTC()
	op, _, err := store.GetOrCreateOperation(operation.Execution{
		ID:            "op-1",
		ScopeKind:     operation.ScopeProposal,
		ScopeID:       "proposal-1",
		OperationKind: "line_activate",
		OperationKey:  "attempt-01",
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   "tester",
		TraceID:       "trace-1",
		ProposalID:    "proposal-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		t.Fatalf("GetOrCreateOperation() error = %v", err)
	}
	item, err := store.EnqueueWorkItem(queue.WorkItem{
		ID:          "work-1",
		OperationID: op.ID,
		Queue:       queue.ProposalQueue,
		Kind:        "approved_proposal",
		Status:      queue.WorkQueued,
		ProposalID:  "proposal-1",
		TraceID:     "trace-1",
		CreatedAt:   now,
		UpdatedAt:   now,
		RequestedBy: "tester",
	})
	if err != nil {
		t.Fatalf("EnqueueWorkItem() error = %v", err)
	}
	claimed, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.ProposalQueue}, "worker-a", 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("ClaimNextWorkItem() ok=%t err=%v", ok, err)
	}
	if claimed.ID != item.ID {
		t.Fatalf("claimed wrong item: %+v", claimed)
	}
	if _, err := store.RescheduleWorkItem(item.ID, map[string]interface{}{"workspace_id": "workspace-1"}, "workspace initializing", time.Time{}); err != nil {
		t.Fatalf("RescheduleWorkItem() error = %v", err)
	}
	rescheduledOp, ok := store.GetOperation(op.ID)
	if !ok {
		t.Fatal("expected operation to remain present")
	}
	if rescheduledOp.Status != operation.StatusQueued {
		t.Fatalf("expected operation to be requeued, got %s", rescheduledOp.Status)
	}
	reclaimed, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.ProposalQueue}, "worker-b", 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("second ClaimNextWorkItem() ok=%t err=%v", ok, err)
	}
	if reclaimed.ID != item.ID {
		t.Fatalf("expected reclaimed item %s, got %s", item.ID, reclaimed.ID)
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
	receipt := submitIngressCommandForTest(t, store, "github:github:delivery-123", transition.CommandIngressRecordEvent, "cmd-postgres-github-event", "tester", time.Now().UTC(), map[string]any{
		"source":          string(ingestion.SourceGitHub),
		"source_event_id": "delivery-123",
		"dedupe_key":      "github:delivery-123",
		"severity":        string(ingestion.SeverityWarning),
		"metadata": map[string]interface{}{
			"event_type":  "pull_request",
			"action":      "closed",
			"merged":      "true",
			"proposal_id": proposal.ID,
			"html_url":    "https://github.com/piplabs/rsi-agent-platform/pull/321",
		},
		"created_at": time.Now().UTC(),
	})
	if receipt.ResultRef == "" {
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
	foundEffect := false
	for _, effect := range reloaded.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != persisted.CurrentAttemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			foundEffect = true
		default:
			t.Fatalf("unexpected proposal approval bootstrap effect %s", effect.EffectKind)
		}
	}
	if !foundEffect {
		t.Fatal("expected one queued attempt bootstrap effect after idempotent proposal approval")
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
	submitProposalCommandForTest(t, storeA, proposal.ID, transition.CommandProposalMarkFailedValidation, "cmd-postgres-proposal-failed-validation", nil)

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

func TestPostgresRetryProposalRepoChangeReopensRunningWorkspaceOpenWithCanceledWorkItem(t *testing.T) {
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

	_, _, _, seededProposal := seedPromotableFailureProposal(t, store)
	if err := store.withProposalLockedStoreTx(seededProposal.ID, func(tx *sql.Tx, mem *MemoryStore) error {
		proposal := seedProposalAttemptPhaseReconcileFixture(t, mem, improvement.WorkspaceQueued)
		now := time.Now().UTC()

		op := mem.operations["op-workspace-open-1"]
		op.Status = operation.StatusRunning
		op.Holder = "worker-a"
		op.CompletedAt = nil
		op.UpdatedAt = now
		mem.operations[op.ID] = op

		item := mem.workItems["work-workspace-open-1"]
		item.Status = queue.WorkCanceled
		item.LeaseOwner = ""
		item.LeaseExpiresAt = nil
		item.LastError = "operation already terminal"
		item.CompletedAt = &now
		item.UpdatedAt = now
		mem.workItems[item.ID] = item

		if err := replaceProposalScope(tx, mem, proposal.ID); err != nil {
			return err
		}
		if err := replaceChangeAttemptScope(tx, mem.changeAttempts["attempt-reconcile-1"]); err != nil {
			return err
		}
		if err := replaceAttemptWorkspaceScope(tx, mem.attemptWorkspaces["workspace-reconcile-1"]); err != nil {
			return err
		}
		if err := replaceRepoChangeJobScope(tx, mem, proposal.ID); err != nil {
			return err
		}
		if err := replaceOperationScope(tx, mem.operations["op-workspace-open-1"]); err != nil {
			return err
		}
		if err := replaceWorkItemScope(tx, mem.workItems["work-workspace-open-1"]); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("seed running workspace_open state: %v", err)
	}

	item, err := store.RetryProposalRepoChange(seededProposal.ID, "tester")
	if err != nil {
		t.Fatalf("RetryProposalRepoChange() error = %v", err)
	}
	if item.Kind != "workspace_open" || item.Status != queue.WorkQueued {
		t.Fatalf("expected reopened workspace_open work item, got %+v", item)
	}

	reloaded, err := store.readStore()
	if err != nil {
		t.Fatalf("readStore() error = %v", err)
	}
	op, ok := reloaded.GetOperation("op-workspace-open-1")
	if !ok {
		t.Fatalf("expected operation op-workspace-open-1")
	}
	if op.Status != operation.StatusQueued || op.Holder != "" {
		t.Fatalf("expected workspace_open operation to be queued and unheld, got %+v", op)
	}

	var reopened queue.WorkItem
	found := false
	for _, candidate := range reloaded.ListWorkItems() {
		if candidate.ID == "work-workspace-open-1" {
			reopened = candidate
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected work item work-workspace-open-1")
	}
	if reopened.Status != queue.WorkQueued || reopened.LastError != "" {
		t.Fatalf("expected canceled work item to be reopened, got %+v", reopened)
	}
}

func TestPostgresRecordActionResultReusesCanonicalRowForDuplicateID(t *testing.T) {
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

	intent := queueActionIntentForTest(t, store, action.Intent{
		ID:         "intent-duplicate-action-result",
		OwnerPlane: "control",
		TraceID:    "trace-duplicate-action-result",
		Kind:       action.KindToolRead,
		CreatedAt:  time.Now().UTC(),
	}, "cmd-postgres-duplicate-action-queue")
	submitActionCommandForTest(t, store, intent.ID, transition.CommandActionStart, "cmd-postgres-duplicate-action-start", time.Now().UTC(), map[string]any{
		"operation_id": "op-duplicate-action-result",
	})

	now := time.Now().UTC()
	first := submitActionCommandForTest(t, store, intent.ID, transition.CommandActionSucceed, "cmd-postgres-duplicate-action-succeed", now, map[string]any{
		"operation_id": "op-duplicate-action-result",
		"executor":     "tester",
		"provider_ref": "action-result-001",
		"started_at":   now,
		"completed_at": now,
	})
	again := submitActionCommandForTest(t, store, intent.ID, transition.CommandActionSucceed, "cmd-postgres-duplicate-action-succeed", now, map[string]any{
		"operation_id": "op-duplicate-action-result",
		"executor":     "tester",
		"provider_ref": "action-result-001",
		"started_at":   now,
		"completed_at": now,
	})
	if again.ResultRef != first.ResultRef {
		t.Fatalf("expected duplicate action command to keep canonical result ref %s, got %+v", first.ResultRef, again)
	}
	results := store.ListActionResults(intent.ID)
	if len(results) != 1 {
		t.Fatalf("expected one persisted action result row, got %+v", results)
	}
}
