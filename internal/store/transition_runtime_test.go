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

	item, err = store.FailEffectExecution("eff-1", "runner timeout")
	if err != nil {
		t.Fatalf("FailEffectExecution() error = %v", err)
	}
	if item.Status != transition.EffectFailed || item.RetryCount != 1 {
		t.Fatalf("expected failed effect execution with retry count, got %+v", item)
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
			"command_kind": string(transition.CommandWorkflowFailed),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, claimed, err := store.ClaimEffectExecution("eff-queue-eval", "worker-1", 30*time.Second); err != nil {
		t.Fatalf("ClaimEffectExecution() error = %v", err)
	} else if !claimed {
		t.Fatal("expected effect claim to succeed")
	}
	if _, err := store.CompleteEffectExecution("eff-queue-eval", "result-1"); err != nil {
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

	effect, err = store.CompleteEffectExecution("eff-pg-1", "result-1")
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
				"command_kind": string(transition.CommandWorkflowFailed),
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
	if _, err := store.CompleteEffectExecution("eff-pg-queue-eval", "result-1"); err != nil {
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
