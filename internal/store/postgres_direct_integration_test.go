package store

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

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
	item, err := SeedAttemptWorkspaceForTesting(store, improvement.AttemptWorkspace{
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
	reviewed, err := ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
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

	if strings.TrimSpace(reviewed.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt after approval, got %+v", reviewed)
	}
	foundEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != reviewed.CurrentAttemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			foundEffect = true
		default:
			t.Fatalf("unexpected approval bootstrap effect %s", effect.EffectKind)
		}
	}
	if !foundEffect {
		t.Fatalf("expected queued attempt bootstrap effect for proposal %s", proposal.ID)
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
		_, err := ReviewProposalForTesting(storeA, proposal.ID, decision)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := ReviewProposalForTesting(storeB, proposal.ID, decision)
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
	if _, err := ReviewProposalForTesting(storeA, proposal.ID, review.ProposalReview{
		ProposalID: proposal.ID,
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "alice",
		CreatedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("approve proposal: %v", err)
	}
	approved, ok := findProposalByID(storeA.ListProposals(), proposal.ID)
	if !ok || strings.TrimSpace(approved.CurrentAttemptID) == "" {
		t.Fatalf("expected current attempt after approval, got %+v", approved)
	}
	now := time.Now().UTC()
	if _, err := SeedRepoChangeJobForTesting(storeA, improvement.RepoChangeJob{
		ID:               "job-postgres-direct-retry-1",
		ProposalID:       proposal.ID,
		AttemptID:        approved.CurrentAttemptID,
		ConversationID:   approved.ConversationID,
		CaseID:           approved.CaseID,
		OriginTraceID:    firstNonEmpty(approved.OriginTraceID, approved.TraceID),
		CandidateKey:     approved.CandidateKey,
		Status:           string(review.ProposalFailedValidation),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       "codex/postgres-direct-retry",
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		t.Fatalf("upsert repo change job: %v", err)
	}
	if _, _, err := AdvanceProposalToFailedValidationForTesting(storeA, proposal.ID, now); err != nil {
		t.Fatalf("AdvanceProposalToFailedValidationForTesting() error = %v", err)
	}

	var wg sync.WaitGroup
	type retryResult struct {
		receipt transition.CommandReceipt
		err     error
	}
	results := make(chan retryResult, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		receipt, err := storeA.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposal.ID,
			CommandKind: string(transition.CommandProposalRetryAttempt),
			CommandID:   "cmd-postgres-retry-a",
			Actor:       "alice",
			OccurredAt:  now.Add(time.Second),
			Payload: map[string]any{
				"reviewer_id": "alice",
				"scope":       string(review.FeedbackScopeLine),
			},
		})
		results <- retryResult{receipt: receipt, err: err}
	}()
	go func() {
		defer wg.Done()
		receipt, err := storeB.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposal.ID,
			CommandKind: string(transition.CommandProposalRetryAttempt),
			CommandID:   "cmd-postgres-retry-b",
			Actor:       "alice",
			OccurredAt:  now.Add(2 * time.Second),
			Payload: map[string]any{
				"reviewer_id": "alice",
				"scope":       string(review.FeedbackScopeLine),
			},
		})
		results <- retryResult{receipt: receipt, err: err}
	}()
	wg.Wait()
	close(results)

	var receipts []transition.CommandReceipt
	for result := range results {
		if result.err != nil {
			t.Fatalf("SubmitCommand(proposal_retry_attempt) error = %v", result.err)
		}
		if result.receipt.DecisionKind == transition.DecisionReject {
			t.Fatalf("expected retry command accepted, got %+v", result.receipt)
		}
		receipts = append(receipts, result.receipt)
	}
	if len(receipts) != 2 {
		t.Fatalf("expected two retry receipts, got %d", len(receipts))
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

	reloadedProposal, ok := findProposalByID(reloaded.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after retry", proposal.ID)
	}
	if strings.TrimSpace(reloadedProposal.CurrentAttemptID) == "" || reloadedProposal.CurrentAttemptID == approved.CurrentAttemptID {
		t.Fatalf("expected exactly one new current attempt after retry, got %+v", reloadedProposal)
	}
	newAttempts := 0
	for _, attempt := range reloaded.ListChangeAttempts() {
		if attempt.ProposalID == proposal.ID && attempt.ID != approved.CurrentAttemptID {
			newAttempts++
		}
	}
	if newAttempts != 1 {
		t.Fatalf("expected one new retry attempt, got %d", newAttempts)
	}
	bootstrapEffects := 0
	for _, effect := range reloaded.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != reloadedProposal.CurrentAttemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			bootstrapEffects++
		}
	}
	if bootstrapEffects != 1 {
		t.Fatalf("expected one queued retry bootstrap effect, got %d", bootstrapEffects)
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
