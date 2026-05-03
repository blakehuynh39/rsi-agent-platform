package control

import (
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestRunnerExecutionLifecycleReconcilesFinalizingWithoutHeartbeatByCreatedAt(t *testing.T) {
	now := time.Now().UTC()
	createdAt := now.Add(-5 * time.Minute)
	store := storepkg.NewMemoryStore()
	if _, err := store.RecordRunnerExecution(storepkg.RunnerExecution{
		ExecutionID: "hexec-finalizing-no-heartbeat",
		Status:      "finalizing",
		Holder:      "hermes-executor:hexec-finalizing-no-heartbeat",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}); err != nil {
		t.Fatalf("RecordRunnerExecution() error = %v", err)
	}

	lifecycle := newRunnerExecutionLifecycle(config.Config{HermesExecutionHeartbeatTimeout: time.Minute}, store)
	lifecycle.now = func() time.Time { return now }
	remaining := lifecycle.reconcileStaleActiveExecutions()
	if len(remaining) != 0 {
		t.Fatalf("stale finalizing execution should be reconciled, remaining=%+v", remaining)
	}
	record, ok := store.GetRunnerExecution("hexec-finalizing-no-heartbeat")
	if !ok {
		t.Fatal("expected runner execution record")
	}
	if record.Status != "failed" || record.FailureClass != "plugin_execution_envelope_missing" {
		t.Fatalf("stale finalizing execution should fail closed as missing envelope, got %+v", record)
	}
	if record.CompletedAt == nil || record.HeartbeatAt == nil {
		t.Fatalf("terminal stale finalizing record should have completed and heartbeat timestamps, got %+v", record)
	}
}
