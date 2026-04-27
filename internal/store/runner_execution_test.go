package store

import (
	"errors"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
)

func TestMemoryStoreRecordRunnerExecutionHeartbeatOnlyPreservesStatus(t *testing.T) {
	store := NewMemoryStore()
	started := time.Now().Add(-time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-running",
		Status:      "running",
		HeartbeatAt: &started,
		CreatedAt:   started,
		UpdatedAt:   started,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(initial) error = %v", err)
	}

	heartbeat := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-running",
		HeartbeatAt: &heartbeat,
		UpdatedAt:   heartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(heartbeat) error = %v", err)
	}

	record, ok := store.GetRunnerExecution("hexec-running")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "running" {
		t.Fatalf("status = %q, want running", record.Status)
	}
	if record.HeartbeatAt == nil || !record.HeartbeatAt.Equal(heartbeat) {
		t.Fatalf("heartbeat not updated: %+v", record)
	}
}

func TestMemoryStoreRecordRunnerExecutionDoesNotDowngradeActiveStatusToQueued(t *testing.T) {
	store := NewMemoryStore()
	started := time.Now().Add(-time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-active",
		Status:      "running",
		HeartbeatAt: &started,
		CreatedAt:   started,
		UpdatedAt:   started,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(initial) error = %v", err)
	}

	updateTime := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-active",
		Status:      "queued",
		UpdatedAt:   updateTime,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(queued downgrade) error = %v", err)
	}

	record, ok := store.GetRunnerExecution("hexec-active")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "running" {
		t.Fatalf("status = %q, want running", record.Status)
	}
	if !record.UpdatedAt.After(started) {
		t.Fatalf("non-status update should still be recorded, got %+v", record)
	}
}

func TestMemoryStoreRecordRunnerExecutionReflectsCancelRequestedIntent(t *testing.T) {
	store := NewMemoryStore()
	started := time.Now().Add(-time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID:     "hexec-cancel-intent",
		Status:          "running",
		CancelRequested: true,
		HeartbeatAt:     &started,
		CreatedAt:       started,
		UpdatedAt:       started,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(initial) error = %v", err)
	}

	heartbeat := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-cancel-intent",
		Status:      "running",
		HeartbeatAt: &heartbeat,
		UpdatedAt:   heartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(heartbeat) error = %v", err)
	}

	record, ok := store.GetRunnerExecution("hexec-cancel-intent")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "cancel_requested" || !record.CancelRequested {
		t.Fatalf("cancel intent should be reflected in status, got %+v", record)
	}
	if record.HeartbeatAt == nil || !record.HeartbeatAt.Equal(heartbeat) {
		t.Fatalf("heartbeat not updated: %+v", record)
	}
}

func TestMemoryStoreRunnerExecutionNormalizationDefaultsBlankStatusToQueued(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.runnerExecutions["hexec-blank"] = RunnerExecution{
		ExecutionID: "hexec-blank",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	record, ok := store.GetRunnerExecution("hexec-blank")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "queued" {
		t.Fatalf("blank persisted status should normalize to queued, got %q", record.Status)
	}
	if active := store.ListActiveRunnerExecutions(); len(active) != 1 || active[0].ExecutionID != "hexec-blank" {
		t.Fatalf("blank status should be listed as queued active execution: %+v", active)
	}
}

func TestMemoryStoreRecordRunnerExecutionDefaultsNewBlankStatusToQueued(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	record, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-new",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	if record.Status != "queued" {
		t.Fatalf("new record status = %q, want queued", record.Status)
	}
}

func TestMemoryStoreRecordRunnerExecutionNormalizesStatusCase(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	record, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-case",
		Status:      "Running",
		HeartbeatAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}
	if record.Status != "running" {
		t.Fatalf("status = %q, want running", record.Status)
	}
	active := store.ListActiveRunnerExecutions()
	if len(active) != 1 || active[0].ExecutionID != "hexec-case" {
		t.Fatalf("case-normalized running status should remain active: %+v", active)
	}
}

func TestMemoryStoreRecordRunnerExecutionTerminalUpdateSetsHeartbeatWhenMissing(t *testing.T) {
	store := NewMemoryStore()
	staleHeartbeat := time.Now().Add(-5 * time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-terminal-heartbeat",
		Status:      "running",
		HeartbeatAt: &staleHeartbeat,
		CreatedAt:   staleHeartbeat,
		UpdatedAt:   staleHeartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(initial) error = %v", err)
	}

	completedAt := time.Now().UTC()
	record, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-terminal-heartbeat",
		Status:      "failed",
		CompletedAt: &completedAt,
		UpdatedAt:   completedAt,
	})
	if err != nil {
		t.Fatalf("RecordRunnerExecution(failed) error = %v", err)
	}
	if record.HeartbeatAt == nil || !record.HeartbeatAt.Equal(completedAt) {
		t.Fatalf("terminal update should advance heartbeat to completed_at, got %+v", record)
	}
	if record.Status != "failed" || record.CompletedAt == nil {
		t.Fatalf("expected failed terminal record, got %+v", record)
	}
}

func TestMemoryStoreRunnerExecutionOrphanedIsTerminal(t *testing.T) {
	store := NewMemoryStore()
	completedAt := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-orphaned",
		Status:      "orphaned",
		CompletedAt: &completedAt,
		CreatedAt:   completedAt,
		UpdatedAt:   completedAt,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(orphaned) error = %v", err)
	}
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-orphaned",
		Status:      "running",
		UpdatedAt:   completedAt.Add(time.Second),
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(regression) error = %v", err)
	}
	record, ok := store.GetRunnerExecution("hexec-orphaned")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "orphaned" || record.HeartbeatAt == nil || !record.HeartbeatAt.Equal(completedAt) {
		t.Fatalf("orphaned should be immutable terminal status with terminal heartbeat, got %+v", record)
	}
	for _, active := range store.ListActiveRunnerExecutions() {
		if active.ExecutionID == "hexec-orphaned" {
			t.Fatalf("orphaned execution must not be active: %+v", active)
		}
	}
}

func TestMemoryStoreRunnerExecutionHolderCASRequiresExistingRecord(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()

	_, err := store.RecordRunnerExecutionWithHolderCAS(RunnerExecution{
		ExecutionID: "hexec-missing",
		Status:      "running",
		Holder:      "hermes-executor:hexec-missing",
		HeartbeatAt: &now,
		UpdatedAt:   now,
	}, "previous-holder", nil)
	if !errors.Is(err, ErrHolderCASMismatch) {
		t.Fatalf("RecordRunnerExecutionWithHolderCAS() error = %v, want ErrHolderCASMismatch", err)
	}
	if _, ok := store.GetRunnerExecution("hexec-missing"); ok {
		t.Fatal("CAS against missing row must not create a runner execution")
	}
}

func TestMemoryStoreRunnerExecutionCASUsesHeartbeatWhenHolderWasBlank(t *testing.T) {
	store := NewMemoryStore()
	initialHeartbeat := time.Now().Add(-time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-unowned",
		Status:      "running",
		HeartbeatAt: &initialHeartbeat,
		CreatedAt:   initialHeartbeat,
		UpdatedAt:   initialHeartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(initial) error = %v", err)
	}

	firstHeartbeat := time.Now().UTC()
	if _, err := store.RecordRunnerExecutionWithHolderCAS(RunnerExecution{
		ExecutionID: "hexec-unowned",
		Status:      "running",
		Holder:      "hermes-executor:hexec-unowned",
		HeartbeatAt: &firstHeartbeat,
		UpdatedAt:   firstHeartbeat,
	}, "", &initialHeartbeat); err != nil {
		t.Fatalf("RecordRunnerExecutionWithHolderCAS(first heartbeat) error = %v", err)
	}

	staleHeartbeat := firstHeartbeat.Add(time.Second)
	_, err := store.RecordRunnerExecutionWithHolderCAS(RunnerExecution{
		ExecutionID: "hexec-unowned",
		Status:      "completed",
		Holder:      "hermes-executor:hexec-unowned",
		HeartbeatAt: &staleHeartbeat,
		UpdatedAt:   staleHeartbeat,
	}, "", &initialHeartbeat)
	if !errors.Is(err, ErrHolderCASMismatch) {
		t.Fatalf("stale heartbeat CAS error = %v, want ErrHolderCASMismatch", err)
	}
	record, _ := store.GetRunnerExecution("hexec-unowned")
	if record.Status != "running" || record.CompletedAt != nil {
		t.Fatalf("stale heartbeat CAS mutated record: %+v", record)
	}
}

func TestMemoryStoreRunnerExecutionCASRejectsTerminalRecord(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-terminal-cas",
		Status:      "completed",
		Holder:      "hermes-executor:hexec-terminal-cas",
		HeartbeatAt: &now,
		CompletedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(initial) error = %v", err)
	}

	_, err := store.RecordRunnerExecutionWithHolderCAS(RunnerExecution{
		ExecutionID: "hexec-terminal-cas",
		Status:      "running",
		Holder:      "hermes-executor:hexec-terminal-cas",
		UpdatedAt:   now.Add(time.Second),
	}, "hermes-executor:hexec-terminal-cas", &now)
	if !errors.Is(err, ErrHolderCASMismatch) {
		t.Fatalf("terminal CAS error = %v, want ErrHolderCASMismatch", err)
	}
	record, _ := store.GetRunnerExecution("hexec-terminal-cas")
	if record.Status != "completed" {
		t.Fatalf("terminal CAS mutated record: %+v", record)
	}
}

func TestMemoryStoreCancelRunnerExecutionsForCasePreservesCancelling(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID:     "hexec-cancelling",
		CaseID:          "case-1",
		TraceID:         "trace-old",
		Status:          "cancelling",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	cancelled := store.CancelRunnerExecutionsForCase("case-1", "trace-current", "trace_superseded")
	if len(cancelled) != 1 {
		t.Fatalf("cancelled count = %d, want 1", len(cancelled))
	}
	if cancelled[0].Status != "cancelling" || !cancelled[0].CancelRequested {
		t.Fatalf("cancelling execution should not be demoted: %+v", cancelled[0])
	}
}

func TestMemoryStoreRunnerExecutionStaleTakeoverOnEmptyHolderRequiresCAS(t *testing.T) {
	store := NewMemoryStore()
	initialTime := time.Now().Add(-time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-stale-empty",
		Status:      "running",
		HeartbeatAt: &initialTime,
		CreatedAt:   initialTime,
		UpdatedAt:   initialTime,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(initial) error = %v", err)
	}

	firstTakeover := time.Now().UTC()
	sentinel := HolderCASExpectEmpty()
	if _, err := store.RecordRunnerExecutionWithHolderCAS(RunnerExecution{
		ExecutionID: "hexec-stale-empty",
		Status:      "running",
		Holder:      "holder-1",
		HeartbeatAt: &firstTakeover,
		UpdatedAt:   firstTakeover,
	}, sentinel, nil); err != nil {
		t.Fatalf("RecordRunnerExecutionWithHolderCAS(first takeover) error = %v", err)
	}

	secondTakeover := firstTakeover.Add(time.Second)
	_, err := store.RecordRunnerExecutionWithHolderCAS(RunnerExecution{
		ExecutionID: "hexec-stale-empty",
		Status:      "running",
		Holder:      "holder-2",
		HeartbeatAt: &secondTakeover,
		UpdatedAt:   secondTakeover,
	}, sentinel, nil)
	if !errors.Is(err, ErrHolderCASMismatch) {
		t.Fatalf("concurrent stale takeover should fail with CAS mismatch, got error = %v", err)
	}

	record, _ := store.GetRunnerExecution("hexec-stale-empty")
	if record.Holder != "holder-1" {
		t.Fatalf("holder should remain holder-1, got %q", record.Holder)
	}
}

func TestPostgresRunnerExecutionUsesSharedMergeAndTerminalHeartbeat(t *testing.T) {
	postgresURL, cleanup := openTempPostgresURL(t)
	defer cleanup()
	store := openMigratedPostgresStoreForRunnerExecutionTest(t, postgresURL)
	defer store.db.Close()

	started := time.Now().Add(-5 * time.Minute).UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID:     "hexec-pg-shared-merge",
		Status:          "cancelling",
		CancelRequested: true,
		HeartbeatAt:     &started,
		CreatedAt:       started,
		UpdatedAt:       started,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(initial) error = %v", err)
	}
	blankStatusHeartbeat := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-pg-shared-merge",
		HeartbeatAt: &blankStatusHeartbeat,
		UpdatedAt:   blankStatusHeartbeat,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(blank status heartbeat) error = %v", err)
	}
	record, ok := store.GetRunnerExecution("hexec-pg-shared-merge")
	if !ok {
		t.Fatal("expected runner execution after heartbeat")
	}
	if record.Status != "cancelling" || !timeClose(record.HeartbeatAt, blankStatusHeartbeat, time.Millisecond) {
		t.Fatalf("blank status update should preserve status and update heartbeat, got %+v", record)
	}
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-pg-shared-merge",
		Status:      "running",
		UpdatedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(regression) error = %v", err)
	}
	record, ok = store.GetRunnerExecution("hexec-pg-shared-merge")
	if !ok {
		t.Fatal("expected runner execution")
	}
	if record.Status != "cancelling" {
		t.Fatalf("Postgres should preserve Go merge status ordering, got %+v", record)
	}

	completedAt := time.Now().UTC()
	if _, err := store.RecordRunnerExecution(RunnerExecution{
		ExecutionID: "hexec-pg-shared-merge",
		Status:      "failed",
		CompletedAt: &completedAt,
		UpdatedAt:   completedAt,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution(failed) error = %v", err)
	}
	record, ok = store.GetRunnerExecution("hexec-pg-shared-merge")
	if !ok {
		t.Fatal("expected terminal runner execution")
	}
	if !timeClose(record.HeartbeatAt, completedAt, time.Millisecond) {
		t.Fatalf("Postgres terminal update should advance heartbeat, got %+v", record)
	}
}

func timeClose(actual *time.Time, expected time.Time, tolerance time.Duration) bool {
	if actual == nil {
		return false
	}
	delta := actual.Sub(expected)
	if delta < 0 {
		delta = -delta
	}
	return delta <= tolerance
}

func openMigratedPostgresStoreForRunnerExecutionTest(t *testing.T, postgresURL string) *PostgresStore {
	t.Helper()
	db, err := platformdb.OpenPostgres(postgresURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	if _, err := platformdb.ApplyMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("apply migrations: %v", err)
	}
	store, err := NewPostgresStore(config.Config{
		StoreBackend: "postgres",
		PostgresURL:  postgresURL,
	})
	if err != nil {
		_ = db.Close()
		t.Fatalf("NewPostgresStore() error = %v", err)
	}
	_ = db.Close()
	return store
}
