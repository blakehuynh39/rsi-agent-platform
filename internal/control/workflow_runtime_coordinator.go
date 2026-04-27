package control

import (
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type WorkflowRuntimeCoordinator struct {
	cfg   config.Config
	store storepkg.Store
	now   func() time.Time
}

func newWorkflowRuntimeCoordinator(cfg config.Config, store storepkg.Store) WorkflowRuntimeCoordinator {
	return WorkflowRuntimeCoordinator{
		cfg:   cfg,
		store: store,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (c WorkflowRuntimeCoordinator) snapshot(effect transition.EffectExecution, workflowStatus string, traceStatus events.Status, runner *storepkg.RunnerExecution) WorkflowRuntimeSnapshot {
	now := c.now()
	return WorkflowRuntimeSnapshot{
		WorkflowStatus:   workflowStatus,
		TraceStatus:      string(traceStatus),
		Effect:           effect,
		RunnerExecution:  runner,
		DrainStarted:     false,
		HeartbeatTimeout: c.cfg.HermesExecutionHeartbeatTimeout,
		Now:              now,
	}
}

func (c WorkflowRuntimeCoordinator) runnerSnapshot(executionID string) WorkflowRuntimeSnapshot {
	now := c.now()
	var runner *storepkg.RunnerExecution
	if existing, ok := c.store.GetRunnerExecution(executionID); ok {
		copy := existing
		runner = &copy
	}
	return WorkflowRuntimeSnapshot{
		RunnerExecution:  runner,
		HeartbeatTimeout: c.cfg.HermesExecutionHeartbeatTimeout,
		Now:              now,
	}
}

func (c WorkflowRuntimeCoordinator) recordRunnerExecution(update storepkg.RunnerExecution) (storepkg.RunnerExecution, error) {
	return c.recordRunnerExecutionWithHolderCAS(update, "", nil)
}

func (c WorkflowRuntimeCoordinator) recordRunnerExecutionWithHolderCAS(update storepkg.RunnerExecution, expectedHolder string, expectedHeartbeatAt *time.Time) (storepkg.RunnerExecution, error) {
	return c.recordRunnerExecutionEvent(WorkflowRuntimeEventRunnerRecord, update, expectedHolder, expectedHeartbeatAt)
}

func (c WorkflowRuntimeCoordinator) recordRunnerExecutionEvent(kind WorkflowRuntimeEventKind, update storepkg.RunnerExecution, expectedHolder string, expectedHeartbeatAt *time.Time) (storepkg.RunnerExecution, error) {
	snapshot := c.runnerSnapshot(update.ExecutionID)
	decision := ReduceWorkflowRuntime(snapshot, WorkflowRuntimeEvent{
		Kind:                kind,
		RunnerUpdate:        update,
		ExpectedHolder:      expectedHolder,
		ExpectedHeartbeatAt: expectedHeartbeatAt,
	})
	return c.applyRunnerDecision(snapshot, decision)
}

func (c WorkflowRuntimeCoordinator) applyRunnerDecision(snapshot WorkflowRuntimeSnapshot, decision WorkflowRuntimeDecision) (storepkg.RunnerExecution, error) {
	if decision.DecisionKind == "cas_conflict" {
		if existing, ok := workflowRuntimeExistingRunner(snapshot); ok {
			return existing, storepkg.ErrHolderCASMismatch
		}
		return storepkg.RunnerExecution{}, storepkg.ErrHolderCASMismatch
	}
	if decision.RunnerUpdate == nil {
		if existing, ok := workflowRuntimeExistingRunner(snapshot); ok {
			return existing, nil
		}
		return storepkg.RunnerExecution{}, nil
	}
	if decision.ExpectedHolder != "" || decision.ExpectedHeartbeatAt != nil {
		return c.store.RecordRunnerExecutionWithHolderCAS(*decision.RunnerUpdate, decision.ExpectedHolder, decision.ExpectedHeartbeatAt)
	}
	return c.store.RecordRunnerExecution(*decision.RunnerUpdate)
}

func (c WorkflowRuntimeCoordinator) completeClaimedEffect(effect transition.EffectExecution, resultRef string) error {
	decision := ReduceWorkflowRuntime(c.snapshot(effect, "", "", nil), WorkflowRuntimeEvent{
		Kind:      WorkflowRuntimeEventEffectComplete,
		ResultRef: resultRef,
		Reason:    "effect completed",
	})
	if decision.DecisionKind == "noop" {
		return nil
	}
	item, err := c.store.CompleteEffectExecution(effect.ID, effect.Holder, decision.EffectResultRef)
	if err != nil {
		return err
	}
	if item.Status != transition.EffectCompleted {
		return fmt.Errorf("effect %s completion was not applied; current status=%s holder=%s", effect.ID, item.Status, item.Holder)
	}
	return nil
}

func (c WorkflowRuntimeCoordinator) deferClaimedEffect(effect transition.EffectExecution, lease time.Duration, reason string) error {
	decision := ReduceWorkflowRuntime(c.snapshot(effect, "", "", nil), WorkflowRuntimeEvent{
		Kind:       WorkflowRuntimeEventEffectDefer,
		DeferLease: lease,
		Reason:     reason,
	})
	if decision.DecisionKind == "noop" {
		return nil
	}
	item, err := c.store.DeferEffectExecution(effect.ID, effect.Holder, decision.EffectDeferLease, decision.EffectError)
	if err != nil {
		return err
	}
	if item.Status != transition.EffectRunning || strings.TrimSpace(item.Holder) != "" || item.LeaseExpiresAt == nil {
		return fmt.Errorf("effect %s deferral was not applied; current status=%s holder=%s", effect.ID, item.Status, item.Holder)
	}
	return nil
}

func (c WorkflowRuntimeCoordinator) deferClaimedEffectForDrain(effect transition.EffectExecution) error {
	decision := ReduceWorkflowRuntime(WorkflowRuntimeSnapshot{
		Effect:       effect,
		DrainStarted: true,
		Now:          c.now(),
	}, WorkflowRuntimeEvent{Kind: WorkflowRuntimeEventDrainStarted})
	if decision.DecisionKind == "noop" {
		return nil
	}
	item, err := c.store.DeferEffectExecution(effect.ID, effect.Holder, decision.EffectDeferLease, decision.EffectError)
	if err != nil {
		return err
	}
	if item.Status != transition.EffectRunning || strings.TrimSpace(item.Holder) != "" || item.LeaseExpiresAt == nil {
		return fmt.Errorf("effect %s drain deferral was not applied; current status=%s holder=%s", effect.ID, item.Status, item.Holder)
	}
	return nil
}

func (c WorkflowRuntimeCoordinator) failClaimedEffect(effect transition.EffectExecution, lastError string) error {
	decision := ReduceWorkflowRuntime(c.snapshot(effect, "", "", nil), WorkflowRuntimeEvent{
		Kind:   WorkflowRuntimeEventEffectFail,
		Reason: lastError,
	})
	if decision.DecisionKind == "noop" {
		return nil
	}
	item, err := c.store.FailEffectExecution(effect.ID, effect.Holder, decision.EffectError)
	if err != nil {
		return err
	}
	if item.Status != transition.EffectFailed {
		return fmt.Errorf("effect %s failure was not applied; current status=%s holder=%s", effect.ID, item.Status, item.Holder)
	}
	return nil
}
