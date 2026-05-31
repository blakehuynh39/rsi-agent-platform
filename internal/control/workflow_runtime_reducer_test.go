package control

import (
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func TestWorkflowRuntimeReducerRunnerTerminalImmutable(t *testing.T) {
	now := time.Now().UTC()
	existing := storepkg.RunnerExecution{
		ExecutionID: "hexec-terminal",
		Status:      "completed",
		HeartbeatAt: &now,
		CompletedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution: &existing,
		Now:             now,
	}, WorkflowRuntimeEvent{
		Kind:         WorkflowRuntimeEventRunnerRecord,
		RunnerUpdate: storepkg.RunnerExecution{ExecutionID: "hexec-terminal", Status: "running"},
	})
	if decision.DecisionKind != "noop" || decision.RunnerUpdate != nil {
		t.Fatalf("terminal runner mutation decision = %+v, want noop without update", decision)
	}
}

func TestWorkflowRuntimeReducerCancellationDominatesSuccessfulResult(t *testing.T) {
	now := time.Now().UTC()
	existing := storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	resp := clients.RunnerResponse{OK: true, Message: "late success", Provider: "hermes", Raw: map[string]any{}}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution: &existing,
		Now:             now,
	}, WorkflowRuntimeEvent{
		Kind: WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{
			ExecutionID: "hexec-cancel",
			Status:      "completed",
			Result:      &resp,
		},
	})
	if decision.RunnerUpdate == nil {
		t.Fatalf("expected runner audit update, got %+v", decision)
	}
	if decision.RunnerUpdate.Status != "cancelled" || decision.RunnerUpdate.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("late successful result update = %+v, want cancelled audit result", decision.RunnerUpdate)
	}
	if decision.WorkflowFailure == nil || decision.WorkflowFailure.Class != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("workflow failure = %+v, want cancellation failure", decision.WorkflowFailure)
	}
	if decision.RunnerResponse != nil {
		t.Fatalf("cancelled late result must not be deliverable: %+v", decision.RunnerResponse)
	}
}

func TestWorkflowRuntimeReducerCancellationOverridesFailedResultFailureClass(t *testing.T) {
	now := time.Now().UTC()
	existing := storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-failed",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	resp := clients.RunnerResponse{OK: false, Message: "late failure", Provider: "hermes", Raw: map[string]any{"failure_class": "worker_failed"}}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution: &existing,
		Now:             now,
	}, WorkflowRuntimeEvent{
		Kind: WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{
			ExecutionID: "hexec-cancel-failed",
			Status:      "failed",
			Result:      &resp,
		},
	})
	if decision.RunnerUpdate == nil {
		t.Fatalf("expected runner audit update, got %+v", decision)
	}
	if decision.RunnerUpdate.Status != "cancelled" || decision.RunnerUpdate.FailureClass != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("late failed result update = %+v, want cancellation to dominate failure class", decision.RunnerUpdate)
	}
	if decision.WorkflowFailure == nil || decision.WorkflowFailure.Class != workflowFailureRunnerExecutionCancelled {
		t.Fatalf("workflow failure = %+v, want cancellation failure", decision.WorkflowFailure)
	}
}

func TestWorkflowRuntimeReducerTerminalStatusCarriesExecutorTelemetry(t *testing.T) {
	now := time.Now().UTC()
	heartbeat := now.Add(-30 * time.Second)
	existing := storepkg.RunnerExecution{
		ExecutionID: "hexec-orphan",
		Status:      "running",
		HeartbeatAt: &heartbeat,
		CreatedAt:   heartbeat,
		UpdatedAt:   heartbeat,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution: &existing,
		Now:             now,
	}, WorkflowRuntimeEvent{
		Kind: WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{
			ExecutionID:                  "hexec-orphan",
			OperationID:                  "op-orphan",
			TraceID:                      "trace-orphan",
			WorkflowID:                   "wf-orphan",
			ExecutorInstanceID:           "executor-pod-old",
			CurrentExecutorInstanceID:    "executor-pod-new",
			ExecutorStartedAtUnix:        1779091470,
			CurrentExecutorStartedAtUnix: 1779128886,
			Status:                       "orphaned",
			Message:                      "Persisted executor status was running, but no local execution process is active.",
			Phase:                        "main",
			LastObservedStatus:           "running",
			LastObservedLedgerSeq:        "1223",
			LastObservedEventType:        "tool.call.progress",
			LastObservedEventStatus:      "running",
			LastObservedRecordedAt:       "2026-05-18T08:16:34Z",
			StatusFilePath:               "/workspace/company/.rsi/runs/_executions/hexec-orphan.json",
			StatusFileMtimeUnix:          1779092194,
			StatusFileSizeBytes:          640,
		},
	})
	if decision.RunnerResponse == nil {
		t.Fatalf("expected runner response, got %+v", decision)
	}
	diagnostics, _ := decision.RunnerResponse.Raw["runner_diagnostics"].(map[string]any)
	if diagnostics["executor_status"] != "orphaned" {
		t.Fatalf("executor_status diagnostic = %#v, want orphaned", diagnostics["executor_status"])
	}
	if diagnostics["last_observed_ledger_seq"] != "1223" || diagnostics["last_observed_event_type"] != "tool.call.progress" {
		t.Fatalf("missing last observation diagnostics: %#v", diagnostics)
	}
	if diagnostics["executor_instance_id"] != "executor-pod-old" || diagnostics["current_executor_instance_id"] != "executor-pod-new" {
		t.Fatalf("missing executor instance diagnostics: %#v", diagnostics)
	}
	if diagnostics["executor_started_at_unix"] != float64(1779091470) || diagnostics["current_executor_started_at_unix"] != float64(1779128886) {
		t.Fatalf("missing executor start diagnostics: %#v", diagnostics)
	}
	if diagnostics["status_file_size_bytes"] != int64(640) {
		t.Fatalf("status file diagnostics = %#v", diagnostics["status_file_size_bytes"])
	}
}

func TestWorkflowRuntimeReducerCancelRequestedRunningStatusRefreshesHeartbeat(t *testing.T) {
	heartbeat := time.Now().Add(-30 * time.Second).UTC()
	now := heartbeat.Add(10 * time.Second)
	existing := storepkg.RunnerExecution{
		ExecutionID:     "hexec-cancel-running",
		Status:          "cancel_requested",
		CancelRequested: true,
		HeartbeatAt:     &heartbeat,
		CreatedAt:       heartbeat,
		UpdatedAt:       heartbeat,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution:  &existing,
		HeartbeatTimeout: time.Minute,
		Now:              now,
	}, WorkflowRuntimeEvent{
		Kind:         WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{ExecutionID: "hexec-cancel-running", Status: "running"},
	})
	if decision.DecisionKind != "wait_runner" || decision.RunnerUpdate == nil {
		t.Fatalf("cancel-requested running decision = %+v, want wait runner update", decision)
	}
	if decision.RunnerUpdate.Status != "cancel_requested" || !decision.RunnerUpdate.CancelRequested {
		t.Fatalf("cancel-requested running update = %+v, want cancel_requested", decision.RunnerUpdate)
	}
	if decision.RunnerUpdate.HeartbeatAt == nil || !decision.RunnerUpdate.HeartbeatAt.Equal(now) {
		t.Fatalf("cancel-requested running status should refresh heartbeat from raw executor status, got %v want %v", decision.RunnerUpdate.HeartbeatAt, now)
	}
}

func TestWorkflowRuntimeReducerQueuedHeartbeatExpires(t *testing.T) {
	started := time.Now().Add(-5 * time.Minute).UTC()
	now := time.Now().UTC()
	existing := storepkg.RunnerExecution{
		ExecutionID: "hexec-queued",
		Status:      "queued",
		CreatedAt:   started,
		UpdatedAt:   started,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution:  &existing,
		HeartbeatTimeout: time.Minute,
		Now:              now,
	}, WorkflowRuntimeEvent{
		Kind:         WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{ExecutionID: "hexec-queued", Status: "queued"},
	})
	if decision.RunnerUpdate == nil || decision.RunnerUpdate.Status != "failed" {
		t.Fatalf("queued timeout decision = %+v, want failed runner update", decision)
	}
	if decision.RunnerResponse == nil || decision.RunnerResponse.OK {
		t.Fatalf("queued timeout response = %+v, want fail-closed response", decision.RunnerResponse)
	}
}

func TestWorkflowRuntimeReducerFinalizingHeartbeatExpiresAsMissingEnvelope(t *testing.T) {
	started := time.Now().Add(-5 * time.Minute).UTC()
	now := time.Now().UTC()
	existing := storepkg.RunnerExecution{
		ExecutionID: "hexec-finalizing-timeout",
		Status:      "finalizing",
		HeartbeatAt: &started,
		OperationID: "op-1",
		TraceID:     "trace-1",
		WorkflowID:  "wf-1",
		CreatedAt:   started,
		UpdatedAt:   started,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution:  &existing,
		HeartbeatTimeout: time.Minute,
		Now:              now,
	}, WorkflowRuntimeEvent{
		Kind:         WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{ExecutionID: "hexec-finalizing-timeout", Status: "finalizing"},
	})
	if decision.RunnerUpdate == nil || decision.RunnerUpdate.Status != "failed" {
		t.Fatalf("finalizing timeout decision = %+v, want failed runner update", decision)
	}
	if decision.RunnerUpdate.FailureClass != "plugin_execution_envelope_missing" {
		t.Fatalf("failure class = %q, want plugin_execution_envelope_missing", decision.RunnerUpdate.FailureClass)
	}
	if decision.RunnerResponse == nil || decision.RunnerResponse.OK {
		t.Fatalf("finalizing timeout response = %+v, want fail-closed response", decision.RunnerResponse)
	}
	if got := stringFromMap(decision.RunnerResponse.Raw, "failure_class"); got != "plugin_execution_envelope_missing" {
		t.Fatalf("runner response failure_class = %q", got)
	}
}

func TestWorkflowRuntimeReducerFinalizingTransitionRefreshesHeartbeatBeforeTimeout(t *testing.T) {
	heartbeat := time.Now().Add(-30 * time.Second).UTC()
	now := heartbeat.Add(45 * time.Second)
	existing := storepkg.RunnerExecution{
		ExecutionID: "hexec-finalizing-active",
		Status:      "running",
		HeartbeatAt: &heartbeat,
		CreatedAt:   heartbeat,
		UpdatedAt:   heartbeat,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution:  &existing,
		HeartbeatTimeout: time.Minute,
		Now:              now,
	}, WorkflowRuntimeEvent{
		Kind:         WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{ExecutionID: "hexec-finalizing-active", Status: "finalizing"},
	})
	if decision.DecisionKind != "wait_runner" || decision.RunnerUpdate == nil {
		t.Fatalf("finalizing active decision = %+v, want wait runner update", decision)
	}
	if decision.RunnerUpdate.Status != "finalizing" {
		t.Fatalf("finalizing active status = %q", decision.RunnerUpdate.Status)
	}
	if decision.RunnerUpdate.HeartbeatAt == nil || !decision.RunnerUpdate.HeartbeatAt.Equal(now) {
		t.Fatalf("finalizing transition must refresh heartbeat, got %v want %v", decision.RunnerUpdate.HeartbeatAt, now)
	}
}

func TestWorkflowRuntimeReducerFinalizingDoesNotRefreshHeartbeatBeforeTimeout(t *testing.T) {
	heartbeat := time.Now().Add(-30 * time.Second).UTC()
	now := heartbeat.Add(45 * time.Second)
	existing := storepkg.RunnerExecution{
		ExecutionID: "hexec-finalizing-active",
		Status:      "finalizing",
		HeartbeatAt: &heartbeat,
		CreatedAt:   heartbeat,
		UpdatedAt:   heartbeat,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution:  &existing,
		HeartbeatTimeout: time.Minute,
		Now:              now,
	}, WorkflowRuntimeEvent{
		Kind:         WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{ExecutionID: "hexec-finalizing-active", Status: "finalizing"},
	})
	if decision.DecisionKind != "wait_runner" || decision.RunnerUpdate == nil {
		t.Fatalf("finalizing active decision = %+v, want wait runner update", decision)
	}
	if decision.RunnerUpdate.Status != "finalizing" {
		t.Fatalf("finalizing active status = %q", decision.RunnerUpdate.Status)
	}
	if decision.RunnerUpdate.HeartbeatAt != nil {
		t.Fatalf("finalizing poll must not refresh heartbeat, got %v want nil update", decision.RunnerUpdate.HeartbeatAt)
	}
}

func TestWorkflowRuntimeReducerFinalizingInitializesMissingHeartbeat(t *testing.T) {
	started := time.Now().UTC()
	now := started.Add(10 * time.Second)
	existing := storepkg.RunnerExecution{
		ExecutionID: "hexec-finalizing-first-observed",
		Status:      "queued",
		CreatedAt:   started,
		UpdatedAt:   started,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution:  &existing,
		HeartbeatTimeout: time.Minute,
		Now:              now,
	}, WorkflowRuntimeEvent{
		Kind:         WorkflowRuntimeEventRunnerStatus,
		RunnerStatus: clients.HermesExecutionStatus{ExecutionID: "hexec-finalizing-first-observed", Status: "finalizing"},
	})
	if decision.DecisionKind != "wait_runner" || decision.RunnerUpdate == nil {
		t.Fatalf("first finalizing decision = %+v, want wait runner update", decision)
	}
	if decision.RunnerUpdate.Status != "finalizing" {
		t.Fatalf("first finalizing status = %q", decision.RunnerUpdate.Status)
	}
	if decision.RunnerUpdate.HeartbeatAt == nil || !decision.RunnerUpdate.HeartbeatAt.Equal(now) {
		t.Fatalf("first finalizing should establish heartbeat anchor at %v, got %+v", now, decision.RunnerUpdate)
	}
}

func TestWorkflowRuntimeReducerDrainDefersClaimedEffect(t *testing.T) {
	now := time.Now().UTC()
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		DrainStarted: true,
		Effect: transition.EffectExecution{
			ID:     "eff-drain",
			Status: transition.EffectRunning,
		},
		Now: now,
	}, WorkflowRuntimeEvent{Kind: WorkflowRuntimeEventEffectClaimed})
	if decision.DecisionKind != "defer_effect" || len(decision.SideEffects) != 1 {
		t.Fatalf("drain decision = %+v, want defer side effect", decision)
	}
	if decision.SideEffects[0].Kind != WorkflowRuntimeSideEffectDeferEffect {
		t.Fatalf("side effect = %+v, want defer", decision.SideEffects[0])
	}
}

func TestWorkflowRuntimeReducerSupersededRequestsSingleCancel(t *testing.T) {
	now := time.Now().UTC()
	existing := storepkg.RunnerExecution{
		ExecutionID: "hexec-old",
		Status:      "running",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution: &existing,
		Now:             now,
	}, WorkflowRuntimeEvent{Kind: WorkflowRuntimeEventSuperseded})
	if decision.RunnerUpdate == nil || decision.RunnerUpdate.Status != "cancel_requested" || !decision.RunnerUpdate.CancelRequested {
		t.Fatalf("supersession update = %+v, want cancel_requested", decision.RunnerUpdate)
	}
	if len(decision.SideEffects) != 1 || decision.SideEffects[0].Kind != WorkflowRuntimeSideEffectCancelRunner {
		t.Fatalf("supersession side effects = %+v, want one cancel", decision.SideEffects)
	}

	cancelling := existing
	cancelling.Status = "cancelling"
	decision = ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		RunnerExecution: &cancelling,
		Now:             now,
	}, WorkflowRuntimeEvent{Kind: WorkflowRuntimeEventSuperseded})
	if len(decision.SideEffects) != 0 {
		t.Fatalf("cancelling supersession must not redispatch cancel: %+v", decision.SideEffects)
	}
}
