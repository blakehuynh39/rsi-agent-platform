package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func TestMemoryStoreRecordCommandReceiptIsIdempotent(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()

	first, created, err := store.RecordCommandReceipt(transition.CommandReceipt{
		CommandID:        "cmd-1",
		MachineKind:      transition.MachineWorkflow,
		AggregateID:      "wf-1",
		CommandKind:      "workflow_started",
		DecisionKind:     transition.DecisionAdvance,
		AggregateVersion: 1,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("RecordCommandReceipt(first) error = %v", err)
	}
	if !created {
		t.Fatal("expected first receipt insert to create a row")
	}

	second, created, err := store.RecordCommandReceipt(transition.CommandReceipt{
		CommandID:    "cmd-1",
		MachineKind:  transition.MachineWorkflow,
		AggregateID:  "wf-1",
		CommandKind:  "workflow_started",
		DecisionKind: transition.DecisionAdvance,
	})
	if err != nil {
		t.Fatalf("RecordCommandReceipt(second) error = %v", err)
	}
	if created {
		t.Fatal("expected duplicate command receipt to reuse the original row")
	}
	if second.CreatedAt != first.CreatedAt || second.CommandID != first.CommandID {
		t.Fatalf("expected duplicate receipt to return original row, got first=%+v second=%+v", first, second)
	}
}

func TestMemoryStoreEffectExecutionClaimAndFail(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-1"] = transition.EffectExecution{
		ID:             "eff-1",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-1",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		IdempotencyKey: "wf-1:invoke",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	item, claimed, err := store.ClaimEffectExecution("eff-1", "worker-1", 30*time.Second)
	if err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	}
	if !claimed || item.Status != transition.EffectRunning {
		t.Fatalf("expected running claim, got claimed=%v item=%+v", claimed, item)
	}
	if item.LeaseExpiresAt == nil {
		t.Fatal("expected claimed effect execution to have a lease")
	}

	item, err = store.FailEffectExecution("eff-1", "worker-1", "runner timeout")
	if err != nil {
		t.Fatalf("FailEffectExecution() error = %v", err)
	}
	if item.Status != transition.EffectFailed || item.RetryCount != 1 {
		t.Fatalf("expected failed effect execution with retry count, got %+v", item)
	}
}

func TestMemoryStoreClaimEffectExecutionZeroLeaseHasNoExpiry(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-zero-lease"] = transition.EffectExecution{
		ID:             "eff-zero-lease",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-zero-lease",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		IdempotencyKey: "wf-zero-lease:invoke",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	item, claimed, err := store.ClaimEffectExecution("eff-zero-lease", "worker-1", 0)
	if err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	}
	if !claimed || item.Status != transition.EffectRunning || item.Holder != "worker-1" {
		t.Fatalf("expected running zero-lease claim, got claimed=%v item=%+v", claimed, item)
	}
	if item.LeaseExpiresAt != nil {
		t.Fatalf("zero-lease claim should not create expiry, got %+v", item)
	}

	item, claimed, err = store.ClaimEffectExecution("eff-zero-lease", "worker-2", time.Second)
	if err != nil {
		t.Fatalf("ClaimEffectExecution(reclaim) error = %v", err)
	}
	if claimed {
		t.Fatalf("zero-lease running effect should not be automatically reclaimed, got %+v", item)
	}
}

func TestMemoryStoreClaimEffectExecutionDoesNotReclaimLeaselessRunningEffect(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-leaseless"] = transition.EffectExecution{
		ID:             "eff-leaseless",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-leaseless",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectRunning,
		Holder:         "old-worker",
		IdempotencyKey: "wf-leaseless:invoke",
		CreatedAt:      now.Add(-time.Minute),
		UpdatedAt:      now.Add(-time.Minute),
	}

	item, claimed, err := store.ClaimEffectExecution("eff-leaseless", "new-worker", 30*time.Second)
	if err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	}
	if claimed {
		t.Fatalf("leaseless running effect should not be recoverable without an explicit expired lease, got %+v", item)
	}
	if item.Holder != "old-worker" || item.LeaseExpiresAt != nil {
		t.Fatalf("leaseless running effect mutated: %+v", item)
	}
}

func TestMemoryStoreClaimEffectExecutionDoesNotReclaimFreshLeaselessRunningEffect(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-leaseless-fresh"] = transition.EffectExecution{
		ID:             "eff-leaseless-fresh",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-leaseless-fresh",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectRunning,
		Holder:         "old-worker",
		IdempotencyKey: "wf-leaseless-fresh:invoke",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	item, claimed, err := store.ClaimEffectExecution("eff-leaseless-fresh", "new-worker", 30*time.Second)
	if err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	}
	if claimed {
		t.Fatalf("fresh leaseless running effect should not be claimable, got %+v", item)
	}
	if item.Holder != "old-worker" || item.LeaseExpiresAt != nil {
		t.Fatalf("fresh leaseless running effect mutated: %+v", item)
	}
}

func TestMemoryStoreClaimNextEffectExecutionCountsLeaselessRunningAsActive(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-leaseless"] = transition.EffectExecution{
		ID:             "eff-leaseless",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-leaseless",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectRunning,
		Holder:         "old-worker",
		QueueName:      "leaseless-workflow",
		ScopeKey:       "conversation-1",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		IdempotencyKey: "wf-leaseless:invoke",
		CreatedAt:      now.Add(-time.Minute),
		UpdatedAt:      now.Add(-time.Minute),
	}

	item, claimed, err := store.ClaimNextEffectExecution("new-worker", 30*time.Second, []string{"leaseless-workflow"}, 1)
	if err != nil {
		t.Fatalf("ClaimNextEffectExecution() error = %v", err)
	}
	if claimed {
		t.Fatalf("leaseless running effect should block same-scope work, got %+v", item)
	}
}

func TestMemoryStoreClaimNextEffectExecutionCountsFreshLeaselessRunningAsActive(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-leaseless-fresh"] = transition.EffectExecution{
		ID:             "eff-leaseless-fresh",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-leaseless-fresh",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectRunning,
		Holder:         "old-worker",
		QueueName:      "fresh-workflow",
		ScopeKey:       "conversation-1",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		IdempotencyKey: "wf-leaseless-fresh:invoke",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	store.effectExecutions["eff-queued-same-scope"] = transition.EffectExecution{
		ID:             "eff-queued-same-scope",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-queued-same-scope",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		QueueName:      "fresh-workflow",
		ScopeKey:       "conversation-1",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		IdempotencyKey: "wf-queued-same-scope:invoke",
		CreatedAt:      now.Add(-time.Minute),
		UpdatedAt:      now.Add(-time.Minute),
	}

	item, claimed, err := store.ClaimNextEffectExecution("new-worker", 30*time.Second, []string{"fresh-workflow"}, 1)
	if err != nil {
		t.Fatalf("ClaimNextEffectExecution() error = %v", err)
	}
	if claimed {
		t.Fatalf("fresh leaseless running effect should block same-scope claim, got %+v", item)
	}
}

func TestMemoryStoreClaimNextEffectExecutionPrioritizesAndLimitsScope(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	leaseActive := now.Add(time.Minute)
	queueName := "test-workflow"
	store.effectExecutions["eff-running-scope-a"] = transition.EffectExecution{
		ID:             "eff-running-scope-a",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-running",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectRunning,
		Holder:         "worker-active",
		QueueName:      queueName,
		ScopeKey:       "conversation-a",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		IdempotencyKey: "wf-running:invoke",
		CreatedAt:      now.Add(-3 * time.Minute),
		UpdatedAt:      now.Add(-3 * time.Minute),
		LeaseExpiresAt: &leaseActive,
	}
	store.effectExecutions["eff-simple-scope-a"] = transition.EffectExecution{
		ID:             "eff-simple-scope-a",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-a",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		QueueName:      queueName,
		ScopeKey:       "conversation-a",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		IdempotencyKey: "wf-a:invoke",
		CreatedAt:      now.Add(-2 * time.Minute),
		UpdatedAt:      now.Add(-2 * time.Minute),
	}
	store.effectExecutions["eff-artifact-scope-b"] = transition.EffectExecution{
		ID:             "eff-artifact-scope-b",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-b",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		QueueName:      queueName,
		ScopeKey:       "conversation-b",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		IdempotencyKey: "wf-b:invoke",
		CreatedAt:      now.Add(-4 * time.Minute),
		UpdatedAt:      now.Add(-4 * time.Minute),
	}
	store.effectExecutions["eff-simple-scope-c"] = transition.EffectExecution{
		ID:             "eff-simple-scope-c",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-c",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		QueueName:      queueName,
		ScopeKey:       "conversation-c",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		IdempotencyKey: "wf-c:invoke",
		CreatedAt:      now.Add(-1 * time.Minute),
		UpdatedAt:      now.Add(-1 * time.Minute),
	}

	claimed, ok, err := store.ClaimNextEffectExecution("worker-1", 30*time.Second, []string{queueName}, 1)
	if err != nil {
		t.Fatalf("ClaimNextEffectExecution(first) error = %v", err)
	}
	if !ok {
		t.Fatal("expected first claim to succeed")
	}
	if claimed.ID != "eff-artifact-scope-b" {
		t.Fatalf("first claim = %s, want eff-artifact-scope-b; claimed=%+v", claimed.ID, claimed)
	}

	claimed, ok, err = store.ClaimNextEffectExecution("worker-2", 30*time.Second, []string{queueName}, 1)
	if err != nil {
		t.Fatalf("ClaimNextEffectExecution(second) error = %v", err)
	}
	if !ok {
		t.Fatal("expected second claim to succeed")
	}
	if claimed.ID != "eff-simple-scope-c" {
		t.Fatalf("second claim = %s, want eff-simple-scope-c; claimed=%+v", claimed.ID, claimed)
	}

	claimed, ok, err = store.ClaimNextEffectExecution("worker-3", 30*time.Second, []string{queueName}, 1)
	if err != nil {
		t.Fatalf("ClaimNextEffectExecution(third) error = %v", err)
	}
	if ok {
		t.Fatalf("expected scope-capped queue to have no claimable work, got %+v", claimed)
	}
}

func TestNormalizeEffectSchedulingHonorsExplicitImprovementTaskClass(t *testing.T) {
	item := normalizeEffectScheduling(transition.EffectExecution{
		ID:          "eff-improve",
		MachineKind: transition.MachineWorkflow,
		AggregateID: "wf-improve",
		EffectKind:  transition.EffectInvokeRunner,
		Status:      transition.EffectQueued,
		Payload: map[string]any{
			"task_class": "improvement",
		},
	})

	if item.TaskClass != effectTaskClassImprove {
		t.Fatalf("task_class = %q, want %q", item.TaskClass, effectTaskClassImprove)
	}
	if item.Priority != priorityForTaskClass(effectTaskClassImprove) {
		t.Fatalf("priority = %d, want %d", item.Priority, priorityForTaskClass(effectTaskClassImprove))
	}
}

func TestMemoryStoreClaimNextEffectExecutionForKindsSkipsWrongMachineKind(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-action-stale-workflow-queue"] = transition.EffectExecution{
		ID:             "eff-action-stale-workflow-queue",
		MachineKind:    transition.MachineAction,
		AggregateID:    "action-1",
		EffectKind:     transition.EffectInvokeAction,
		Status:         transition.EffectQueued,
		QueueName:      effectQueueWorkflow,
		ScopeKey:       "conversation-1",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple) + 10,
		IdempotencyKey: "action-1:invoke",
		CreatedAt:      now.Add(-2 * time.Minute),
		UpdatedAt:      now.Add(-2 * time.Minute),
	}
	store.effectExecutions["eff-workflow"] = transition.EffectExecution{
		ID:             "eff-workflow",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "workflow-1",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		QueueName:      effectQueueWorkflow,
		ScopeKey:       "conversation-2",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		IdempotencyKey: "workflow-1:invoke",
		CreatedAt:      now.Add(-1 * time.Minute),
		UpdatedAt:      now.Add(-1 * time.Minute),
	}

	claimed, ok, err := store.ClaimNextEffectExecutionForKinds(
		"worker-1",
		30*time.Second,
		[]string{effectQueueWorkflow},
		1,
		[]EffectClaimSelector{{MachineKind: transition.MachineWorkflow, EffectKind: transition.EffectInvokeRunner}},
	)
	if err != nil {
		t.Fatalf("ClaimNextEffectExecutionForKinds() error = %v", err)
	}
	if !ok {
		t.Fatal("expected workflow effect to be claimable")
	}
	if claimed.ID != "eff-workflow" {
		t.Fatalf("claimed %s, want eff-workflow", claimed.ID)
	}
	if action := store.effectExecutions["eff-action-stale-workflow-queue"]; action.Status != transition.EffectQueued {
		t.Fatalf("stale action effect should not be claimed by workflow worker: %+v", action)
	}
}

func TestMemoryStoreClaimNextEffectExecutionForKindsFiltersPayload(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-action-proposal"] = transition.EffectExecution{
		ID:             "eff-action-proposal",
		MachineKind:    transition.MachineAction,
		AggregateID:    "action-proposal",
		EffectKind:     transition.EffectInvokeAction,
		Status:         transition.EffectQueued,
		QueueName:      effectQueueAction,
		ScopeKey:       "proposal-scope",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple) + 10,
		Payload:        map[string]any{"owner_plane": "proposal"},
		IdempotencyKey: "action-proposal:invoke",
		CreatedAt:      now.Add(-2 * time.Minute),
		UpdatedAt:      now.Add(-2 * time.Minute),
	}
	store.effectExecutions["eff-action-control"] = transition.EffectExecution{
		ID:             "eff-action-control",
		MachineKind:    transition.MachineAction,
		AggregateID:    "action-control",
		EffectKind:     transition.EffectInvokeAction,
		Status:         transition.EffectQueued,
		QueueName:      effectQueueAction,
		ScopeKey:       "control-scope",
		TaskClass:      effectTaskClassSimple,
		Priority:       priorityForTaskClass(effectTaskClassSimple),
		Payload:        map[string]any{"owner_plane": "control"},
		IdempotencyKey: "action-control:invoke",
		CreatedAt:      now.Add(-1 * time.Minute),
		UpdatedAt:      now.Add(-1 * time.Minute),
	}

	claimed, ok, err := store.ClaimNextEffectExecutionForKinds(
		"worker-1",
		30*time.Second,
		[]string{effectQueueAction},
		1,
		[]EffectClaimSelector{{
			MachineKind:   transition.MachineAction,
			EffectKind:    transition.EffectInvokeAction,
			PayloadEquals: map[string]string{"owner_plane": "control"},
		}},
	)
	if err != nil {
		t.Fatalf("ClaimNextEffectExecutionForKinds() error = %v", err)
	}
	if !ok {
		t.Fatal("expected control-owned action to be claimable")
	}
	if claimed.ID != "eff-action-control" {
		t.Fatalf("claimed %s, want eff-action-control", claimed.ID)
	}
	if other := store.effectExecutions["eff-action-proposal"]; other.Status != transition.EffectQueued {
		t.Fatalf("proposal-owned action should remain queued: %+v", other)
	}
}

func TestMemoryStoreCompleteEffectExecutionIsLedgerOnly(t *testing.T) {
	store := NewMemoryStore()
	workflow := store.ListWorkflows()[0]
	trace, ok := store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	initialEvents := len(trace.Events)
	now := time.Now().UTC()
	store.effectExecutions["eff-queue-eval"] = transition.EffectExecution{
		ID:             "eff-queue-eval",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    workflow.ID,
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		IdempotencyKey: workflow.ID + ":invoke_runner",
		Payload: map[string]any{
			"command_kind": string(transition.CommandWorkflowExecutionFailed),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, claimed, err := store.ClaimEffectExecution("eff-queue-eval", "worker-1", 30*time.Second); err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	} else if !claimed {
		t.Fatal("expected effect claim to succeed")
	}
	if _, err := store.CompleteEffectExecution("eff-queue-eval", "worker-1", "result-1"); err != nil {
		t.Fatalf("CompleteEffectExecution() error = %v", err)
	}

	trace, ok = store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	if len(trace.Events) != initialEvents {
		t.Fatalf("expected effect completion to avoid direct trace mutation, got %d events want %d", len(trace.Events), initialEvents)
	}
}

func TestPostgresCommandReceiptAndEffectExecutionLifecycle(t *testing.T) {
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
	receipt, created, err := store.RecordCommandReceipt(transition.CommandReceipt{
		CommandID:        "cmd-pg-1",
		MachineKind:      transition.MachineWorkflow,
		AggregateID:      "wf-pg-1",
		CommandKind:      "workflow_started",
		DecisionKind:     transition.DecisionAdvance,
		AggregateVersion: 1,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("RecordCommandReceipt() error = %v", err)
	}
	if !created || receipt.CommandID != "cmd-pg-1" {
		t.Fatalf("unexpected created receipt %+v created=%v", receipt, created)
	}

	receipt, created, err = store.RecordCommandReceipt(transition.CommandReceipt{
		CommandID:    "cmd-pg-1",
		MachineKind:  transition.MachineWorkflow,
		AggregateID:  "wf-pg-1",
		CommandKind:  "workflow_started",
		DecisionKind: transition.DecisionAdvance,
	})
	if err != nil {
		t.Fatalf("RecordCommandReceipt(duplicate) error = %v", err)
	}
	if created {
		t.Fatal("expected duplicate receipt to reuse existing row")
	}

	err = store.withTx(func(tx *sql.Tx) error {
		return persistEffectExecutions(tx, []transition.EffectExecution{{
			ID:             "eff-pg-1",
			MachineKind:    transition.MachineWorkflow,
			AggregateID:    "wf-pg-1",
			EffectKind:     transition.EffectInvokeRunner,
			Status:         transition.EffectQueued,
			IdempotencyKey: "wf-pg-1:invoke",
			CreatedAt:      now,
			UpdatedAt:      now,
		}})
	})
	if err != nil {
		t.Fatalf("persistEffectExecutions() error = %v", err)
	}

	effect, claimed, err := store.ClaimEffectExecution("eff-pg-1", "worker-1", 45*time.Second)
	if err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	}
	if !claimed || effect.Status != transition.EffectRunning {
		t.Fatalf("expected running claim, got claimed=%v effect=%+v", claimed, effect)
	}

	effect, err = store.CompleteEffectExecution("eff-pg-1", "worker-1", "result-1")
	if err != nil {
		t.Fatalf("CompleteEffectExecution() error = %v", err)
	}
	if effect.Status != transition.EffectCompleted || effect.ResultRef != "result-1" {
		t.Fatalf("expected completed effect execution, got %+v", effect)
	}
}

func TestPostgresCompleteEffectExecutionIsLedgerOnly(t *testing.T) {
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

	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineIngress,
		AggregateID: "slack:171000001.000100",
		CommandKind: string(transition.CommandIngressRecordSlack),
		CommandID:   "cmd-transition-runtime-slack-ingress",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"bot_role":   string(slack.BotArch),
			"team_id":    "T123",
			"channel_id": "D123",
			"thread_ts":  "171000001.000100",
			"user_id":    "U123",
			"text":       "Queue an eval effect after a failed workflow.",
			"ts":         "171000001.000100",
			"created_at": time.Now().UTC(),
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(ingress_record_slack) error = %v", err)
	}
	workflow := store.ListWorkflows()[0]
	trace, ok := store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	initialEvents := len(trace.Events)
	now := time.Now().UTC()
	err = store.withTx(func(tx *sql.Tx) error {
		return persistEffectExecutions(tx, []transition.EffectExecution{{
			ID:             "eff-pg-queue-eval",
			MachineKind:    transition.MachineWorkflow,
			AggregateID:    workflow.ID,
			EffectKind:     transition.EffectInvokeRunner,
			Status:         transition.EffectQueued,
			IdempotencyKey: workflow.ID + ":invoke_runner",
			Payload: map[string]any{
				"command_kind": string(transition.CommandWorkflowExecutionFailed),
			},
			CreatedAt: now,
			UpdatedAt: now,
		}})
	})
	if err != nil {
		t.Fatalf("persistEffectExecutions() error = %v", err)
	}

	if _, claimed, err := store.ClaimEffectExecution("eff-pg-queue-eval", "worker-1", 30*time.Second); err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	} else if !claimed {
		t.Fatal("expected effect claim to succeed")
	}
	if _, err := store.CompleteEffectExecution("eff-pg-queue-eval", "worker-1", "result-1"); err != nil {
		t.Fatalf("CompleteEffectExecution() error = %v", err)
	}

	trace, ok = store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	if len(trace.Events) != initialEvents {
		t.Fatalf("expected effect completion to avoid direct trace mutation, got %d events want %d", len(trace.Events), initialEvents)
	}
}

func TestMemoryStoreStaleHolderCannotFinalizeEffectExecution(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	store.effectExecutions["eff-stale"] = transition.EffectExecution{
		ID:             "eff-stale",
		MachineKind:    transition.MachineWorkflow,
		AggregateID:    "wf-1",
		EffectKind:     transition.EffectInvokeRunner,
		Status:         transition.EffectQueued,
		IdempotencyKey: "wf-1:invoke",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	firstClaim, claimed, err := store.ClaimEffectExecution("eff-stale", "worker-1", time.Second)
	if err != nil || !claimed {
		t.Fatalf("ClaimEffectExecution(worker-1) claimed=%v err=%v", claimed, err)
	}
	expired := now.Add(-time.Second)
	firstClaim.LeaseExpiresAt = &expired
	store.effectExecutions["eff-stale"] = firstClaim

	secondClaim, claimed, err := store.ClaimEffectExecution("eff-stale", "worker-2", 30*time.Second)
	if err != nil || !claimed {
		t.Fatalf("ClaimEffectExecution(worker-2) claimed=%v err=%v", claimed, err)
	}

	item, err := store.CompleteEffectExecution("eff-stale", "worker-1", "result-1")
	if err != nil {
		t.Fatalf("CompleteEffectExecution(stale worker) error = %v", err)
	}
	if item.Status != transition.EffectRunning || item.Holder != "worker-2" {
		t.Fatalf("expected stale completion to be a no-op, got %+v", item)
	}

	item, err = store.FailEffectExecution("eff-stale", "worker-1", "late failure")
	if err != nil {
		t.Fatalf("FailEffectExecution(stale worker) error = %v", err)
	}
	if item.Status != transition.EffectRunning || item.Holder != "worker-2" {
		t.Fatalf("expected stale failure to be a no-op, got %+v", item)
	}

	if secondClaim.Holder != "worker-2" {
		t.Fatalf("expected reclaimed holder worker-2, got %+v", secondClaim)
	}
}

func TestPostgresStaleHolderCannotFinalizeEffectExecution(t *testing.T) {
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
	err = store.withTx(func(tx *sql.Tx) error {
		return persistEffectExecutions(tx, []transition.EffectExecution{{
			ID:             "eff-pg-stale",
			MachineKind:    transition.MachineWorkflow,
			AggregateID:    "wf-pg-stale",
			EffectKind:     transition.EffectInvokeRunner,
			Status:         transition.EffectQueued,
			IdempotencyKey: "wf-pg-stale:invoke",
			CreatedAt:      now,
			UpdatedAt:      now,
		}})
	})
	if err != nil {
		t.Fatalf("persistEffectExecutions() error = %v", err)
	}

	if _, claimed, err := store.ClaimEffectExecution("eff-pg-stale", "worker-1", 10*time.Millisecond); err != nil || !claimed {
		t.Fatalf("ClaimEffectExecution(worker-1) claimed=%v err=%v", claimed, err)
	}
	time.Sleep(25 * time.Millisecond)

	item, claimed, err := store.ClaimEffectExecution("eff-pg-stale", "worker-2", 30*time.Second)
	if err != nil || !claimed {
		t.Fatalf("ClaimEffectExecution(worker-2) claimed=%v err=%v", claimed, err)
	}
	if item.Holder != "worker-2" {
		t.Fatalf("expected reclaimed holder worker-2, got %+v", item)
	}

	item, err = store.FailEffectExecution("eff-pg-stale", "worker-1", "late failure")
	if err != nil {
		t.Fatalf("FailEffectExecution(stale worker) error = %v", err)
	}
	if item.Status != transition.EffectRunning || item.Holder != "worker-2" {
		t.Fatalf("expected stale failure to be a no-op, got %+v", item)
	}
}

func TestPostgresClaimEffectExecutionDoesNotReclaimLeaselessRunningEffect(t *testing.T) {
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
	err = store.withTx(func(tx *sql.Tx) error {
		return persistEffectExecutions(tx, []transition.EffectExecution{{
			ID:             "eff-pg-leaseless",
			MachineKind:    transition.MachineWorkflow,
			AggregateID:    "wf-pg-leaseless",
			EffectKind:     transition.EffectInvokeRunner,
			Status:         transition.EffectRunning,
			Holder:         "old-worker",
			IdempotencyKey: "wf-pg-leaseless:invoke",
			CreatedAt:      now.Add(-time.Minute),
			UpdatedAt:      now.Add(-time.Minute),
		}})
	})
	if err != nil {
		t.Fatalf("persistEffectExecutions() error = %v", err)
	}

	item, claimed, err := store.ClaimEffectExecution("eff-pg-leaseless", "new-worker", 30*time.Second)
	if err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	}
	if claimed {
		t.Fatalf("leaseless running effect should not be recoverable without an explicit expired lease, got %+v", item)
	}
	if item.Holder != "old-worker" || item.LeaseExpiresAt != nil {
		t.Fatalf("leaseless running effect mutated: %+v", item)
	}
}
