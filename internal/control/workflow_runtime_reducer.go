package control

import (
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type WorkflowRuntimeEventKind string

const (
	WorkflowRuntimeEventEffectClaimed       WorkflowRuntimeEventKind = "effect.claimed"
	WorkflowRuntimeEventEffectComplete      WorkflowRuntimeEventKind = "effect.complete"
	WorkflowRuntimeEventEffectDefer         WorkflowRuntimeEventKind = "effect.defer"
	WorkflowRuntimeEventEffectFail          WorkflowRuntimeEventKind = "effect.fail"
	WorkflowRuntimeEventRunnerRecord        WorkflowRuntimeEventKind = "runner.record"
	WorkflowRuntimeEventRunnerHeartbeat     WorkflowRuntimeEventKind = "runner.heartbeat"
	WorkflowRuntimeEventRunnerComplete      WorkflowRuntimeEventKind = "runner.complete"
	WorkflowRuntimeEventRunnerStatus        WorkflowRuntimeEventKind = "runner.status"
	WorkflowRuntimeEventRunnerCancel        WorkflowRuntimeEventKind = "runner.cancel"
	WorkflowRuntimeEventRunnerExecutorError WorkflowRuntimeEventKind = "runner.executor_error"
	WorkflowRuntimeEventSuperseded          WorkflowRuntimeEventKind = "workflow.superseded"
	WorkflowRuntimeEventDrainStarted        WorkflowRuntimeEventKind = "drain.started"
)

type WorkflowRuntimeSideEffectKind string

const (
	WorkflowRuntimeSideEffectStartRunner    WorkflowRuntimeSideEffectKind = "runner.start"
	WorkflowRuntimeSideEffectPollRunner     WorkflowRuntimeSideEffectKind = "runner.poll"
	WorkflowRuntimeSideEffectCancelRunner   WorkflowRuntimeSideEffectKind = "runner.cancel"
	WorkflowRuntimeSideEffectCompleteEffect WorkflowRuntimeSideEffectKind = "effect.complete"
	WorkflowRuntimeSideEffectDeferEffect    WorkflowRuntimeSideEffectKind = "effect.defer"
	WorkflowRuntimeSideEffectFailEffect     WorkflowRuntimeSideEffectKind = "effect.fail"
	WorkflowRuntimeSideEffectSlackDeliver   WorkflowRuntimeSideEffectKind = "slack.deliver"
)

type WorkflowRuntimeSnapshot struct {
	WorkflowStatus   string
	TraceStatus      string
	Effect           transition.EffectExecution
	RunnerExecution  *storepkg.RunnerExecution
	ActiveSameCase   []storepkg.RunnerExecution
	DrainStarted     bool
	HeartbeatTimeout time.Duration
	Now              time.Time
}

type WorkflowRuntimeEvent struct {
	Kind                WorkflowRuntimeEventKind
	RunnerUpdate        storepkg.RunnerExecution
	RunnerStatus        clients.HermesExecutionStatus
	RunnerResponse      *clients.RunnerResponse
	RunnerError         error
	ExpectedHolder      string
	ExpectedHeartbeatAt *time.Time
	ResultRef           string
	DeferLease          time.Duration
	Reason              string
}

type WorkflowRuntimeDecision struct {
	DecisionKind        string
	Reason              string
	RunnerUpdate        *storepkg.RunnerExecution
	ExpectedHolder      string
	ExpectedHeartbeatAt *time.Time
	RunnerResponse      *clients.RunnerResponse
	WorkflowFailure     *workflowFailure
	WaitForRunner       bool
	EffectResultRef     string
	EffectDeferLease    time.Duration
	EffectError         string
	SideEffects         []WorkflowRuntimeSideEffect
}

type WorkflowRuntimeSideEffect struct {
	Kind   WorkflowRuntimeSideEffectKind
	Target string
	Reason string
}

func ReduceWorkflowRuntime(snapshot WorkflowRuntimeSnapshot, event WorkflowRuntimeEvent) WorkflowRuntimeDecision {
	if snapshot.Now.IsZero() {
		snapshot.Now = time.Now().UTC()
	}
	switch event.Kind {
	case WorkflowRuntimeEventEffectClaimed:
		return reduceWorkflowRuntimeEffectClaimed(snapshot)
	case WorkflowRuntimeEventEffectComplete, WorkflowRuntimeEventEffectDefer, WorkflowRuntimeEventEffectFail:
		return reduceWorkflowRuntimeEffectDisposition(snapshot, event)
	case WorkflowRuntimeEventRunnerRecord, WorkflowRuntimeEventRunnerHeartbeat, WorkflowRuntimeEventRunnerComplete:
		return reduceWorkflowRuntimeRunnerRecord(snapshot, event)
	case WorkflowRuntimeEventRunnerStatus, WorkflowRuntimeEventRunnerCancel:
		return reduceWorkflowRuntimeRunnerStatus(snapshot, event)
	case WorkflowRuntimeEventRunnerExecutorError:
		return reduceWorkflowRuntimeRunnerExecutorError(snapshot, event)
	case WorkflowRuntimeEventSuperseded:
		return reduceWorkflowRuntimeSuperseded(snapshot, event)
	case WorkflowRuntimeEventDrainStarted:
		return reduceWorkflowRuntimeDrainStarted(snapshot)
	default:
		return WorkflowRuntimeDecision{DecisionKind: "reject", Reason: fmt.Sprintf("unsupported workflow runtime event %q", event.Kind)}
	}
}

func reduceWorkflowRuntimeEffectClaimed(snapshot WorkflowRuntimeSnapshot) WorkflowRuntimeDecision {
	if snapshot.DrainStarted {
		return WorkflowRuntimeDecision{
			DecisionKind:     "defer_effect",
			Reason:           "drain started; defer claimed effect",
			EffectDeferLease: time.Second,
			EffectError:      "deployment_shutdown",
			SideEffects: []WorkflowRuntimeSideEffect{{
				Kind:   WorkflowRuntimeSideEffectDeferEffect,
				Target: snapshot.Effect.ID,
				Reason: "deployment_shutdown",
			}},
		}
	}
	if isTerminalWorkflowStatus(snapshot.WorkflowStatus) || isTerminalTraceStatusString(snapshot.TraceStatus) {
		return WorkflowRuntimeDecision{
			DecisionKind:    "complete_effect",
			Reason:          "workflow or trace is terminal",
			EffectResultRef: strings.TrimSpace(snapshot.TraceStatus),
			SideEffects: []WorkflowRuntimeSideEffect{{
				Kind:   WorkflowRuntimeSideEffectCompleteEffect,
				Target: snapshot.Effect.ID,
				Reason: "terminal_workflow_or_trace",
			}},
		}
	}
	return WorkflowRuntimeDecision{DecisionKind: "advance", Reason: "effect is eligible for execution"}
}

func reduceWorkflowRuntimeEffectDisposition(snapshot WorkflowRuntimeSnapshot, event WorkflowRuntimeEvent) WorkflowRuntimeDecision {
	if effectTerminal(snapshot.Effect.Status) {
		return WorkflowRuntimeDecision{DecisionKind: "noop", Reason: "effect is already terminal"}
	}
	switch event.Kind {
	case WorkflowRuntimeEventEffectComplete:
		return WorkflowRuntimeDecision{
			DecisionKind:    "complete_effect",
			Reason:          firstNonEmpty(event.Reason, "effect completed"),
			EffectResultRef: event.ResultRef,
			SideEffects: []WorkflowRuntimeSideEffect{{
				Kind:   WorkflowRuntimeSideEffectCompleteEffect,
				Target: snapshot.Effect.ID,
				Reason: firstNonEmpty(event.Reason, "effect completed"),
			}},
		}
	case WorkflowRuntimeEventEffectDefer:
		return WorkflowRuntimeDecision{
			DecisionKind:     "defer_effect",
			Reason:           firstNonEmpty(event.Reason, "effect deferred"),
			EffectDeferLease: event.DeferLease,
			EffectError:      event.Reason,
			SideEffects: []WorkflowRuntimeSideEffect{{
				Kind:   WorkflowRuntimeSideEffectDeferEffect,
				Target: snapshot.Effect.ID,
				Reason: firstNonEmpty(event.Reason, "effect deferred"),
			}},
		}
	case WorkflowRuntimeEventEffectFail:
		return WorkflowRuntimeDecision{
			DecisionKind: "fail_effect",
			Reason:       firstNonEmpty(event.Reason, "effect failed"),
			EffectError:  event.Reason,
			SideEffects: []WorkflowRuntimeSideEffect{{
				Kind:   WorkflowRuntimeSideEffectFailEffect,
				Target: snapshot.Effect.ID,
				Reason: firstNonEmpty(event.Reason, "effect failed"),
			}},
		}
	default:
		return WorkflowRuntimeDecision{DecisionKind: "reject", Reason: "unsupported effect disposition"}
	}
}

func reduceWorkflowRuntimeRunnerRecord(snapshot WorkflowRuntimeSnapshot, event WorkflowRuntimeEvent) WorkflowRuntimeDecision {
	update := normalizeWorkflowRuntimeRunnerUpdate(snapshot, event.RunnerUpdate)
	if update.ExecutionID == "" {
		return WorkflowRuntimeDecision{DecisionKind: "reject", Reason: "runner execution id is required"}
	}
	existing, hasExisting := workflowRuntimeExistingRunner(snapshot)
	if hasExisting && storepkg.RunnerExecutionStatusTerminal(existing.Status) {
		if event.ExpectedHolder != "" || event.ExpectedHeartbeatAt != nil {
			return WorkflowRuntimeDecision{DecisionKind: "cas_conflict", Reason: "runner execution became terminal during CAS operation"}
		}
		return WorkflowRuntimeDecision{DecisionKind: "noop", Reason: "runner execution is already terminal"}
	}
	if hasExisting {
		update = applyWorkflowRuntimeRunnerInvariants(existing, update)
	}
	return WorkflowRuntimeDecision{
		DecisionKind:        "record_runner",
		Reason:              firstNonEmpty(event.Reason, string(event.Kind)),
		RunnerUpdate:        &update,
		ExpectedHolder:      event.ExpectedHolder,
		ExpectedHeartbeatAt: event.ExpectedHeartbeatAt,
	}
}

func reduceWorkflowRuntimeRunnerStatus(snapshot WorkflowRuntimeSnapshot, event WorkflowRuntimeEvent) WorkflowRuntimeDecision {
	existing, hasExisting := workflowRuntimeExistingRunner(snapshot)
	if hasExisting && storepkg.RunnerExecutionStatusTerminal(existing.Status) {
		if existing.CancelRequested || strings.EqualFold(existing.Status, "cancelled") {
			failure := workflowFailure{Class: workflowFailureRunnerExecutionCancelled, Summary: "Execution was cancelled or superseded as requested."}
			return WorkflowRuntimeDecision{DecisionKind: "workflow_failure", Reason: failure.Summary, WorkflowFailure: &failure}
		}
		if resp, ok := runnerResponseFromMap(existing.Result); ok {
			return WorkflowRuntimeDecision{DecisionKind: "runner_response", Reason: "stored terminal runner result", RunnerResponse: &resp}
		}
		failureResp := hermesExecutorRecoveryFailure(existing.ExecutionID, workflowFailureRunnerExecutorResultUnavailable, "Hermes executor reached a terminal state without a durable result.", existing.Status)
		return WorkflowRuntimeDecision{DecisionKind: "runner_response", Reason: "terminal runner result unavailable", RunnerResponse: &failureResp}
	}
	status := strings.ToLower(firstNonEmpty(event.RunnerStatus.Status, "running"))
	if event.RunnerStatus.Result != nil {
		updateStatus := "completed"
		if !event.RunnerStatus.Result.OK {
			updateStatus = "failed"
		}
		failureClass := stringValue(event.RunnerStatus.Result.Raw["failure_class"])
		if hasExisting && (existing.CancelRequested || runnerExecutionStatusCancellationPending(existing.Status)) {
			updateStatus = "cancelled"
			failureClass = workflowFailureRunnerExecutionCancelled
		}
		update := storepkg.RunnerExecution{
			ExecutionID:  firstNonEmpty(event.RunnerStatus.ExecutionID, existing.ExecutionID),
			Status:       updateStatus,
			Result:       runnerResponseMap(*event.RunnerStatus.Result),
			FailureClass: failureClass,
			HeartbeatAt:  &snapshot.Now,
			CompletedAt:  &snapshot.Now,
			UpdatedAt:    snapshot.Now,
		}
		decision := reduceWorkflowRuntimeRunnerRecord(snapshot, WorkflowRuntimeEvent{
			Kind:                WorkflowRuntimeEventRunnerRecord,
			RunnerUpdate:        update,
			ExpectedHolder:      event.ExpectedHolder,
			ExpectedHeartbeatAt: event.ExpectedHeartbeatAt,
			Reason:              string(event.Kind),
		})
		if hasExisting && (existing.CancelRequested || runnerExecutionStatusCancellationPending(existing.Status)) {
			failure := workflowFailure{Class: workflowFailureRunnerExecutionCancelled, Summary: "Execution completed after cancellation was requested and is not deliverable."}
			decision.DecisionKind = "workflow_failure"
			decision.WorkflowFailure = &failure
			decision.RunnerResponse = nil
			decision.WaitForRunner = false
		} else {
			decision.RunnerResponse = event.RunnerStatus.Result
		}
		return decision
	}
	switch status {
	case "queued":
		if hasExisting && workflowRuntimeHeartbeatExpired(snapshot, existing) {
			return workflowRuntimeFailureDecision(snapshot, existing.ExecutionID, workflowFailureRunnerExecutorStatusUnavailable, "Hermes executor remained queued past the heartbeat timeout; refusing to defer indefinitely.", "heartbeat_expired")
		}
		update := storepkg.RunnerExecution{ExecutionID: existing.ExecutionID, Status: "queued", UpdatedAt: snapshot.Now}
		return waitWithRunnerUpdate(snapshot, event, update, "runner remains queued")
	case "accepted", "starting", "running", "finalizing":
		enteringFinalizing := status == "finalizing" && hasExisting && runnerExecutionEnteringFinalizing(status, existing)
		if status == "finalizing" && hasExisting && !enteringFinalizing && workflowRuntimeHeartbeatExpired(snapshot, existing) {
			return workflowRuntimeFailureDecision(snapshot, existing.ExecutionID, "plugin_execution_envelope_missing", "Hermes executor heartbeat expired while finalizing the native execution envelope.", "heartbeat_expired")
		}
		updateStatus := status
		if hasExisting && existing.CancelRequested {
			updateStatus = "cancel_requested"
		}
		update := storepkg.RunnerExecution{ExecutionID: existing.ExecutionID, Status: updateStatus, CancelRequested: hasExisting && existing.CancelRequested, UpdatedAt: snapshot.Now}
		if runnerExecutionStatusRefreshesHeartbeat(status, existing) {
			update.HeartbeatAt = &snapshot.Now
		}
		return waitWithRunnerUpdate(snapshot, event, update, "runner is still active")
	case "cancel_requested", "cancelling":
		update := storepkg.RunnerExecution{ExecutionID: existing.ExecutionID, Status: "cancelling", CancelRequested: true, UpdatedAt: snapshot.Now}
		if status == "cancelling" {
			update.HeartbeatAt = &snapshot.Now
		}
		return waitWithRunnerUpdate(snapshot, event, update, "runner cancellation is still active")
	case "completed", "failed", "cancelled", "orphaned":
		if status == "cancelled" && hasExisting && existing.CancelRequested {
			failure := workflowFailure{Class: workflowFailureRunnerExecutionCancelled, Summary: "Execution was cancelled as requested."}
			update := storepkg.RunnerExecution{ExecutionID: existing.ExecutionID, Status: "cancelled", HeartbeatAt: &snapshot.Now, CompletedAt: &snapshot.Now, UpdatedAt: snapshot.Now}
			decision := reduceWorkflowRuntimeRunnerRecord(snapshot, WorkflowRuntimeEvent{Kind: WorkflowRuntimeEventRunnerRecord, RunnerUpdate: update, Reason: string(event.Kind)})
			decision.DecisionKind = "workflow_failure"
			decision.WorkflowFailure = &failure
			return decision
		}
		return workflowRuntimeFailureDecisionWithStatus(snapshot, existing.ExecutionID, workflowFailureRunnerExecutorResultUnavailable, "Hermes executor reached a terminal state without a durable result.", event.RunnerStatus)
	default:
		return workflowRuntimeFailureDecision(snapshot, existing.ExecutionID, workflowFailureRunnerExecutorStatusUnrecognized, fmt.Sprintf("Hermes executor returned unrecognized async status %q.", status), status)
	}
}

func reduceWorkflowRuntimeRunnerExecutorError(snapshot WorkflowRuntimeSnapshot, event WorkflowRuntimeEvent) WorkflowRuntimeDecision {
	existing, hasExisting := workflowRuntimeExistingRunner(snapshot)
	executionID := event.RunnerUpdate.ExecutionID
	if executionID == "" && hasExisting {
		executionID = existing.ExecutionID
	}
	if event.RunnerError != nil && strings.Contains(event.RunnerError.Error(), "returned 404") {
		return workflowRuntimeFailureDecision(snapshot, executionID, workflowFailureRunnerExecutorStatusUnavailable, "Hermes executor status was unavailable for a previously started execution; refusing to launch a duplicate run.", "")
	}
	if hasExisting && workflowRuntimeHeartbeatExpired(snapshot, existing) {
		failure := workflowFailureFromRunnerError(event.RunnerError)
		failureResp := clients.RunnerResponse{OK: false, Message: failure.Summary, Raw: map[string]any{"failure_class": failure.Class}}
		update := storepkg.RunnerExecution{
			ExecutionID:  executionID,
			Status:       "failed",
			Result:       runnerResponseMap(failureResp),
			FailureClass: failure.Class,
			HeartbeatAt:  &snapshot.Now,
			CompletedAt:  &snapshot.Now,
			UpdatedAt:    snapshot.Now,
		}
		decision := reduceWorkflowRuntimeRunnerRecord(snapshot, WorkflowRuntimeEvent{Kind: WorkflowRuntimeEventRunnerRecord, RunnerUpdate: update, Reason: string(event.Kind)})
		decision.DecisionKind = "workflow_failure"
		decision.WorkflowFailure = &failure
		return decision
	}
	return WorkflowRuntimeDecision{DecisionKind: "wait_runner", Reason: firstNonEmpty(event.Reason, "runner executor error remains retryable"), WaitForRunner: true}
}

func reduceWorkflowRuntimeSuperseded(snapshot WorkflowRuntimeSnapshot, event WorkflowRuntimeEvent) WorkflowRuntimeDecision {
	existing, hasExisting := workflowRuntimeExistingRunner(snapshot)
	if !hasExisting {
		return WorkflowRuntimeDecision{DecisionKind: "noop", Reason: "no runner execution to supersede"}
	}
	if storepkg.RunnerExecutionStatusTerminal(existing.Status) {
		return WorkflowRuntimeDecision{DecisionKind: "noop", Reason: "terminal runner execution cannot be superseded"}
	}
	status := "cancel_requested"
	if strings.EqualFold(existing.Status, "cancelling") {
		status = "cancelling"
	}
	update := storepkg.RunnerExecution{
		ExecutionID:     existing.ExecutionID,
		Status:          status,
		CancelRequested: true,
		FailureClass:    firstNonEmpty(existing.FailureClass, workflowFailureRunnerExecutionCancelled),
		UpdatedAt:       snapshot.Now,
	}
	decision := reduceWorkflowRuntimeRunnerRecord(snapshot, WorkflowRuntimeEvent{Kind: WorkflowRuntimeEventRunnerRecord, RunnerUpdate: update, Reason: firstNonEmpty(event.Reason, "workflow superseded")})
	if status == "cancel_requested" {
		decision.SideEffects = append(decision.SideEffects, WorkflowRuntimeSideEffect{Kind: WorkflowRuntimeSideEffectCancelRunner, Target: existing.ExecutionID, Reason: "workflow_superseded"})
	}
	return decision
}

func reduceWorkflowRuntimeDrainStarted(snapshot WorkflowRuntimeSnapshot) WorkflowRuntimeDecision {
	if snapshot.Effect.ID == "" {
		return WorkflowRuntimeDecision{DecisionKind: "noop", Reason: "drain started without a claimed effect"}
	}
	return WorkflowRuntimeDecision{
		DecisionKind:     "defer_effect",
		Reason:           "deployment drain started",
		EffectDeferLease: time.Second,
		EffectError:      "deployment_shutdown",
		SideEffects: []WorkflowRuntimeSideEffect{{
			Kind:   WorkflowRuntimeSideEffectDeferEffect,
			Target: snapshot.Effect.ID,
			Reason: "deployment_shutdown",
		}},
	}
}

func waitWithRunnerUpdate(snapshot WorkflowRuntimeSnapshot, event WorkflowRuntimeEvent, update storepkg.RunnerExecution, reason string) WorkflowRuntimeDecision {
	decision := reduceWorkflowRuntimeRunnerRecord(snapshot, WorkflowRuntimeEvent{
		Kind:                WorkflowRuntimeEventRunnerRecord,
		RunnerUpdate:        update,
		ExpectedHolder:      event.ExpectedHolder,
		ExpectedHeartbeatAt: event.ExpectedHeartbeatAt,
		Reason:              reason,
	})
	decision.DecisionKind = "wait_runner"
	decision.WaitForRunner = true
	return decision
}

func workflowRuntimeFailureDecision(snapshot WorkflowRuntimeSnapshot, executionID string, failureClass string, message string, status string) WorkflowRuntimeDecision {
	failureResp := hermesExecutorRecoveryFailure(executionID, failureClass, message, status)
	return workflowRuntimeFailureDecisionFromResponse(snapshot, executionID, failureClass, failureResp)
}

func workflowRuntimeFailureDecisionWithStatus(snapshot WorkflowRuntimeSnapshot, executionID string, failureClass string, message string, status clients.HermesExecutionStatus) WorkflowRuntimeDecision {
	failureResp := hermesExecutorRecoveryFailureWithStatus(executionID, failureClass, message, status)
	return workflowRuntimeFailureDecisionFromResponse(snapshot, executionID, failureClass, failureResp)
}

func workflowRuntimeFailureDecisionFromResponse(snapshot WorkflowRuntimeSnapshot, executionID string, failureClass string, failureResp clients.RunnerResponse) WorkflowRuntimeDecision {
	update := storepkg.RunnerExecution{
		ExecutionID:  executionID,
		Status:       "failed",
		Result:       runnerResponseMap(failureResp),
		FailureClass: failureClass,
		HeartbeatAt:  &snapshot.Now,
		CompletedAt:  &snapshot.Now,
		UpdatedAt:    snapshot.Now,
	}
	decision := reduceWorkflowRuntimeRunnerRecord(snapshot, WorkflowRuntimeEvent{Kind: WorkflowRuntimeEventRunnerRecord, RunnerUpdate: update, Reason: "fail closed"})
	decision.RunnerResponse = &failureResp
	return decision
}

func normalizeWorkflowRuntimeRunnerUpdate(snapshot WorkflowRuntimeSnapshot, update storepkg.RunnerExecution) storepkg.RunnerExecution {
	update.ExecutionID = strings.TrimSpace(update.ExecutionID)
	if update.ExecutionID == "" {
		if existing, ok := workflowRuntimeExistingRunner(snapshot); ok {
			update.ExecutionID = existing.ExecutionID
		}
	}
	update.Status = strings.ToLower(strings.TrimSpace(update.Status))
	if update.Status == "" {
		update.Status = "queued"
	}
	update.Holder = strings.TrimSpace(update.Holder)
	update.FailureClass = strings.TrimSpace(update.FailureClass)
	if update.UpdatedAt.IsZero() {
		update.UpdatedAt = snapshot.Now
	}
	if storepkg.RunnerExecutionStatusTerminal(update.Status) {
		if update.CompletedAt == nil {
			completedAt := snapshot.Now
			update.CompletedAt = &completedAt
		}
		if update.HeartbeatAt == nil {
			heartbeatAt := snapshot.Now
			update.HeartbeatAt = &heartbeatAt
		}
	}
	return update
}

func applyWorkflowRuntimeRunnerInvariants(existing storepkg.RunnerExecution, update storepkg.RunnerExecution) storepkg.RunnerExecution {
	existingStatus := strings.ToLower(strings.TrimSpace(existing.Status))
	updateStatus := strings.ToLower(strings.TrimSpace(update.Status))
	if existing.CancelRequested || runnerExecutionStatusCancellationPending(existingStatus) {
		update.CancelRequested = true
		switch {
		case updateStatus == "cancel_requested", updateStatus == "cancelling", storepkg.RunnerExecutionStatusTerminal(updateStatus):
		default:
			update.Status = firstNonEmpty(existingStatus, "cancel_requested")
			if update.Status != "cancelling" {
				update.Status = "cancel_requested"
			}
		}
		if storepkg.RunnerExecutionStatusTerminal(updateStatus) {
			update.Status = "cancelled"
			update.FailureClass = workflowFailureRunnerExecutionCancelled
		}
	}
	if storepkg.RunnerExecutionStatusBackward(existingStatus, update.Status) {
		update.Status = existingStatus
	}
	if strings.EqualFold(update.Status, "queued") && existingStatus != "queued" {
		update.Status = existingStatus
	}
	return update
}

func workflowRuntimeHeartbeatExpired(snapshot WorkflowRuntimeSnapshot, record storepkg.RunnerExecution) bool {
	if snapshot.HeartbeatTimeout <= 0 {
		return false
	}
	referenceTime := runnerExecutionHeartbeatReferenceTime(record)
	return !referenceTime.IsZero() && snapshot.Now.Sub(referenceTime) > snapshot.HeartbeatTimeout
}

func workflowRuntimeExistingRunner(snapshot WorkflowRuntimeSnapshot) (storepkg.RunnerExecution, bool) {
	if snapshot.RunnerExecution == nil || strings.TrimSpace(snapshot.RunnerExecution.ExecutionID) == "" {
		return storepkg.RunnerExecution{}, false
	}
	return *snapshot.RunnerExecution, true
}

func effectTerminal(status transition.EffectStatus) bool {
	switch status {
	case transition.EffectCompleted, transition.EffectCanceled, transition.EffectSuperseded:
		return true
	default:
		return false
	}
}

func isTerminalTraceStatusString(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "failed", "suppressed", "needs-human", "needs_human":
		return true
	default:
		return false
	}
}
