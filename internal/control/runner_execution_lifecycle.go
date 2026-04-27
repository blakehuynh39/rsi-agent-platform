package control

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type runnerExecutionLifecycle struct {
	cfg   config.Config
	store storepkg.Store
	now   func() time.Time
}

func newRunnerExecutionLifecycle(cfg config.Config, store storepkg.Store) runnerExecutionLifecycle {
	return runnerExecutionLifecycle{
		cfg:   cfg,
		store: store,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (l runnerExecutionLifecycle) heartbeat(executionID string, payload map[string]any) (storepkg.RunnerExecution, int, error) {
	executionID = strings.TrimSpace(executionID)
	runtime := newWorkflowRuntimeCoordinator(l.cfg, l.store)
	existing, ok := l.store.GetRunnerExecution(executionID)
	if !ok {
		return storepkg.RunnerExecution{}, http.StatusNotFound, fmt.Errorf("runner execution %s not found", executionID)
	}
	if storepkg.RunnerExecutionStatusTerminal(existing.Status) {
		return existing, http.StatusConflict, fmt.Errorf("runner execution %s is terminal", executionID)
	}
	requestHolder := stringValue(payload["holder"])
	holderCAS, err := validateRunnerExecutionHolder(l.cfg, existing, requestHolder)
	if err != nil {
		if errors.Is(err, ErrRunnerExecutionHolderRequired) {
			return storepkg.RunnerExecution{}, http.StatusBadRequest, err
		}
		return storepkg.RunnerExecution{}, http.StatusForbidden, err
	}
	status := strings.ToLower(firstNonEmpty(stringValue(payload["status"]), "running"))
	if !runnerExecutionHeartbeatStatusAllowed(status) {
		return storepkg.RunnerExecution{}, http.StatusBadRequest, fmt.Errorf("invalid heartbeat status %q", status)
	}
	cancelRequestedFromStatus := runnerExecutionStatusCancellationPending(status)
	if existing.CancelRequested {
		statusLower := strings.ToLower(strings.TrimSpace(status))
		if statusLower != "cancel_requested" && statusLower != "cancelling" {
			return existing, http.StatusConflict, fmt.Errorf("runner execution %s has cancellation requested", executionID)
		}
	}
	now := l.now()
	expectedHeartbeat := existing.HeartbeatAt
	update := storepkg.RunnerExecution{
		ExecutionID:     executionID,
		Status:          status,
		Holder:          requestHolder,
		HeartbeatAt:     &now,
		UpdatedAt:       now,
		CancelRequested: existing.CancelRequested || runnerExecutionStatusCancellationPending(existing.Status) || cancelRequestedFromStatus,
	}
	item, err := runtime.recordRunnerExecutionEvent(WorkflowRuntimeEventRunnerHeartbeat, update, holderCAS.ExpectedHolder, expectedHeartbeat)
	if err != nil {
		if errors.Is(err, storepkg.ErrHolderCASMismatch) {
			return storepkg.RunnerExecution{}, http.StatusConflict, err
		}
		return storepkg.RunnerExecution{}, http.StatusInternalServerError, err
	}
	return item, http.StatusOK, nil
}

func (l runnerExecutionLifecycle) complete(executionID string, payload map[string]any) (storepkg.RunnerExecution, int, error) {
	executionID = strings.TrimSpace(executionID)
	runtime := newWorkflowRuntimeCoordinator(l.cfg, l.store)
	existing, ok := l.store.GetRunnerExecution(executionID)
	if !ok {
		return storepkg.RunnerExecution{}, http.StatusNotFound, fmt.Errorf("runner execution %s not found", executionID)
	}
	if storepkg.RunnerExecutionStatusTerminal(existing.Status) {
		return existing, http.StatusOK, nil
	}
	requestHolder := stringValue(payload["holder"])
	holderCAS, err := validateRunnerExecutionHolder(l.cfg, existing, requestHolder)
	if err != nil {
		if errors.Is(err, ErrRunnerExecutionHolderRequired) {
			return storepkg.RunnerExecution{}, http.StatusBadRequest, err
		}
		return storepkg.RunnerExecution{}, http.StatusForbidden, err
	}
	status := strings.ToLower(firstNonEmpty(stringValue(payload["status"]), "completed"))
	if !runnerExecutionCompleteStatusAllowed(status) {
		return storepkg.RunnerExecution{}, http.StatusBadRequest, fmt.Errorf("invalid completion status %q", status)
	}
	now := l.now()
	expectedHeartbeat := existing.HeartbeatAt
	update := storepkg.RunnerExecution{
		ExecutionID:  executionID,
		Status:       status,
		Result:       mapValue(payload["result"]),
		FailureClass: stringValue(payload["failure_class"]),
		Holder:       requestHolder,
		HeartbeatAt:  &now,
		CompletedAt:  &now,
		UpdatedAt:    now,
	}
	if existing.CancelRequested || runnerExecutionStatusCancellationPending(existing.Status) {
		update.CancelRequested = true
		if storepkg.RunnerExecutionStatusTerminal(status) {
			update.Status = "cancelled"
			update.FailureClass = workflowFailureRunnerExecutionCancelled
		}
	}
	item, err := runtime.recordRunnerExecutionEvent(WorkflowRuntimeEventRunnerComplete, update, holderCAS.ExpectedHolder, expectedHeartbeat)
	if err != nil {
		if errors.Is(err, storepkg.ErrHolderCASMismatch) {
			return storepkg.RunnerExecution{}, http.StatusConflict, err
		}
		return storepkg.RunnerExecution{}, http.StatusInternalServerError, err
	}
	return item, http.StatusOK, nil
}

func (l runnerExecutionLifecycle) reconcileStaleActiveExecutions() []storepkg.RunnerExecution {
	active := l.store.ListActiveRunnerExecutions()
	if l.cfg.HermesExecutionHeartbeatTimeout <= 0 {
		return active
	}
	now := l.now()
	runtime := newWorkflowRuntimeCoordinator(l.cfg, l.store)
	runtime.now = l.now
	out := make([]storepkg.RunnerExecution, 0, len(active))
	for _, item := range active {
		if !runnerExecutionLifecycleStatusReconcilesWhenStale(item.Status) {
			out = append(out, item)
			continue
		}
		referenceTime := runnerExecutionHeartbeatReferenceTime(item)
		itemStatus := strings.ToLower(strings.TrimSpace(item.Status))
		if referenceTime.IsZero() && runnerExecutionStatusCancellationPending(itemStatus) {
			referenceTime = item.UpdatedAt.UTC()
		}
		if referenceTime.IsZero() {
			out = append(out, item)
			continue
		}
		if itemStatus == "queued" && item.HeartbeatAt == nil {
			out = append(out, item)
			continue
		}
		if now.Sub(referenceTime) <= l.cfg.HermesExecutionHeartbeatTimeout {
			out = append(out, item)
			continue
		}
		failureClass := workflowFailureRunnerExecutorStatusUnavailable
		if runnerExecutionStatusCancellationPending(itemStatus) {
			failureClass = workflowFailureRunnerExecutionCancelled
		}
		failure := hermesExecutorRecoveryFailure(
			item.ExecutionID,
			failureClass,
			"Hermes executor heartbeat expired while reconciling active executions.",
			"heartbeat_expired",
		)
		expectedHolder := item.Holder
		if expectedHolder == "" {
			expectedHolder = storepkg.HolderCASExpectEmpty()
		}
		updated, err := runtime.recordRunnerExecutionWithHolderCAS(storepkg.RunnerExecution{
			ExecutionID:  item.ExecutionID,
			Status:       "failed",
			Result:       runnerResponseMap(failure),
			FailureClass: failureClass,
			HeartbeatAt:  &now,
			CompletedAt:  &now,
			UpdatedAt:    now,
		}, expectedHolder, item.HeartbeatAt)
		if err != nil || !storepkg.RunnerExecutionStatusTerminal(updated.Status) {
			out = append(out, item)
		}
	}
	return out
}

func runnerExecutionLifecycleStatusReconcilesWhenStale(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "cancel_requested", "cancelling", "running", "starting", "accepted":
		return true
	default:
		return false
	}
}

func runnerExecutionHeartbeatReferenceTime(record storepkg.RunnerExecution) time.Time {
	if record.HeartbeatAt != nil {
		return record.HeartbeatAt.UTC()
	}
	if strings.EqualFold(strings.TrimSpace(record.Status), "queued") {
		return record.CreatedAt.UTC()
	}
	return time.Time{}
}

func runnerExecutionStatusCancellationPending(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "cancel_requested", "cancelling":
		return true
	default:
		return false
	}
}
