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
