package control

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const (
	workflowFailureRunnerMissingStructuredOutput  = "runner_missing_structured_output"
	workflowFailureRunnerInvalidRequest           = "runner_invalid_request"
	workflowFailureRunnerNonOK                    = "runner_non_ok"
	workflowFailureRunnerIterationBudgetExhausted = "runner_iteration_budget_exhausted"
	workflowFailureRunnerPostProcessing           = "runner_post_processing_failure"
	workflowFailureRunnerStateInvariant           = "runner_state_transition_invariant_failed"
	workflowFailureRunnerTransportTimeout         = "runner_transport_timeout"
	workflowFailureRunnerStructuredOutputParse    = "runner_structured_output_parse_failure"
	workflowFailureToolGatewayTimeout             = "tool_gateway_timeout"
	workflowFailureToolGatewayUnavailable         = "tool_gateway_unavailable"
)

type workflowFailure struct {
	Class             string
	Summary           string
	RunnerDiagnostics map[string]any
	RepairAttempted   bool
	RepairSucceeded   bool
	Retryable         bool
}

type workflowFailureError struct {
	failure workflowFailure
}

func (e *workflowFailureError) Error() string {
	return e.failure.Summary
}

func finalizeWorkflowFailureWithDetails(cfg config.Config, store storepkg.Store, workflow workflowLocator, failure workflowFailure) error {
	if strings.TrimSpace(workflow.traceID) == "" {
		return nil
	}
	ctx, err := loadWorkflowContext(store, workflow)
	if err != nil {
		return err
	}
	retryDecision := ""
	payload := map[string]any{
		"last_error":       firstNonEmpty(strings.TrimSpace(failure.Summary), strings.TrimSpace(failure.Class), "workflow failed"),
		"failure_class":    strings.TrimSpace(failure.Class),
		"failure_summary":  firstNonEmpty(strings.TrimSpace(failure.Summary), strings.TrimSpace(failure.Class), "workflow failed"),
		"repair_attempted": failure.RepairAttempted,
		"repair_succeeded": failure.RepairSucceeded,
	}
	if len(failure.RunnerDiagnostics) > 0 {
		payload["runner_diagnostics"] = cloneStringAnyMap(failure.RunnerDiagnostics)
	}
	if retryAt, ok := workflowRetryAt(cfg, store, ctx.workflow, failure); ok {
		retryDecision = "auto_retry"
		payload["retry_decision"] = retryDecision
		payload["retry_after"] = retryAt
		log.Printf("control-plane workflow retry_scheduled case=%s workflow=%s failure_class=%s retry_after=%s", ctx.workflow.CaseID, ctx.workflow.ID, failure.Class, retryAt.Format(time.RFC3339))
	} else {
		payload["retry_decision"] = "needs_human"
		log.Printf("control-plane workflow moved_to_needs_human case=%s workflow=%s failure_class=%s", ctx.workflow.CaseID, ctx.workflow.ID, failure.Class)
	}
	if _, submitErr := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandWorkflowFailed, cfg.ServiceName, time.Now().UTC(), payload); submitErr != nil {
		return submitErr
	}
	if _, _, err := store.ReconcileWorkflowTrace(ctx.workflow.ID); err != nil {
		return err
	}
	if retryDecision == "" {
		log.Printf("control-plane workflow retry_exhausted_or_blocked case=%s workflow=%s failure_class=%s", ctx.workflow.CaseID, ctx.workflow.ID, failure.Class)
	}
	return nil
}

func workflowRetryAt(cfg config.Config, store storepkg.Store, workflow storepkg.Workflow, failure workflowFailure) (time.Time, bool) {
	if !cfg.WorkflowAutoRetryEnabled || !failure.Retryable {
		return time.Time{}, false
	}
	if replyPostBegun(store, workflow.TraceID) {
		return time.Time{}, false
	}
	line, ok := store.GetWorkflowLine(workflow.CaseID)
	if !ok {
		return time.Time{}, false
	}
	maxAttempts := cfg.WorkflowAutoRetryMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if line.AttemptCount >= maxAttempts {
		return time.Time{}, false
	}
	backoffs := cfg.WorkflowAutoRetryBackoffSeconds
	if len(backoffs) == 0 {
		backoffs = []int{15, 60}
	}
	index := line.AttemptCount - 1
	if index < 0 {
		index = 0
	}
	if index >= len(backoffs) {
		index = len(backoffs) - 1
	}
	return time.Now().UTC().Add(time.Duration(backoffs[index]) * time.Second), true
}

func replyPostBegun(store storepkg.Store, traceID string) bool {
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return false
	}
	for _, intent := range store.ListActionIntents() {
		if strings.TrimSpace(intent.TraceID) != traceID || intent.OwnerPlane != "control" || intent.PhaseKey != controlPhaseReplyPost {
			continue
		}
		return true
	}
	return false
}

func workflowFailureFromRunnerError(err error) workflowFailure {
	summary := strings.TrimSpace(err.Error())
	return workflowFailure{
		Class:     workflowFailureRunnerTransportTimeout,
		Summary:   firstNonEmpty(summary, workflowFailureRunnerTransportTimeout),
		Retryable: true,
	}
}

func workflowFailureFromRunnerResponse(resp clients.RunnerResponse) workflowFailure {
	class := firstNonEmpty(strings.TrimSpace(stringValue(resp.Raw["failure_class"])), workflowFailureRunnerNonOK)
	runnerDiagnostics := mapValue(resp.Raw["runner_diagnostics"])
	if class == workflowFailureRunnerNonOK && strings.TrimSpace(stringValue(resp.Raw["structured_output_error"])) != "" {
		class = workflowFailureRunnerStructuredOutputParse
	}
	retryable := class == workflowFailureRunnerNonOK || class == workflowFailureRunnerStructuredOutputParse || class == workflowFailureRunnerTransportTimeout
	return workflowFailure{
		Class:             class,
		Summary:           firstNonEmpty(strings.TrimSpace(stringValueFromMap(runnerDiagnostics, "provider_error_message")), strings.TrimSpace(resp.Message), strings.TrimSpace(stringValue(resp.Raw["structured_output_error"])), class),
		RunnerDiagnostics: runnerDiagnostics,
		RepairAttempted:   boolValue(resp.Raw["repair_attempted"]),
		RepairSucceeded:   boolValue(resp.Raw["repair_succeeded"]),
		Retryable:         retryable,
	}
}

func workflowFailureFromStructuredOutputError(resp clients.RunnerResponse, err error) workflowFailure {
	class := workflowFailureRunnerStructuredOutputParse
	if strings.Contains(strings.ToLower(err.Error()), "missing structured_output") {
		class = workflowFailureRunnerMissingStructuredOutput
	}
	return workflowFailure{
		Class:             class,
		Summary:           strings.TrimSpace(err.Error()),
		RunnerDiagnostics: mapValue(resp.Raw["runner_diagnostics"]),
		RepairAttempted:   boolValue(resp.Raw["repair_attempted"]),
		RepairSucceeded:   boolValue(resp.Raw["repair_succeeded"]),
		Retryable:         true,
	}
}

func workflowFailureFromRunnerPostProcessing(resp clients.RunnerResponse, stage string, err error) workflowFailure {
	summary := strings.TrimSpace(err.Error())
	if stage = strings.TrimSpace(stage); stage != "" {
		summary = fmt.Sprintf("%s: %s", stage, firstNonEmpty(summary, workflowFailureRunnerPostProcessing))
	}
	return workflowFailure{
		Class:             workflowFailureRunnerPostProcessing,
		Summary:           firstNonEmpty(summary, workflowFailureRunnerPostProcessing),
		RunnerDiagnostics: mapValue(resp.Raw["runner_diagnostics"]),
		RepairAttempted:   boolValue(resp.Raw["repair_attempted"]),
		RepairSucceeded:   boolValue(resp.Raw["repair_succeeded"]),
		Retryable:         false,
	}
}

func workflowFailureFromRunnerStateInvariant(resp clients.RunnerResponse, workflowID string, expected transition.WorkflowStateKind, actual string) workflowFailure {
	expectedState := strings.TrimSpace(string(expected))
	actualState := strings.TrimSpace(actual)
	summary := fmt.Sprintf("workflow %s remained in invalid state after successful runner execution: expected %s, got %s", strings.TrimSpace(workflowID), expectedState, firstNonEmpty(actualState, "unknown"))
	return workflowFailure{
		Class:             workflowFailureRunnerStateInvariant,
		Summary:           summary,
		RunnerDiagnostics: mapValue(resp.Raw["runner_diagnostics"]),
		RepairAttempted:   boolValue(resp.Raw["repair_attempted"]),
		RepairSucceeded:   boolValue(resp.Raw["repair_succeeded"]),
		Retryable:         false,
	}
}

func workflowFailureFromContextAction(store storepkg.Store, intent action.Intent) workflowFailure {
	summary := firstNonEmpty(strings.TrimSpace(intent.PolicyVerdict), strings.TrimSpace(latestActionError(store, intent.ID)), string(intent.Status))
	class := ""
	lower := strings.ToLower(summary)
	switch {
	case strings.Contains(lower, "timeout"), strings.Contains(lower, "deadline exceeded"):
		class = workflowFailureToolGatewayTimeout
	case strings.Contains(lower, "unavailable"), strings.Contains(lower, "connection refused"), strings.Contains(lower, "service unavailable"):
		class = workflowFailureToolGatewayUnavailable
	}
	return workflowFailure{
		Class:           firstNonEmpty(class, "tool_gateway_failure"),
		Summary:         summary,
		RepairAttempted: false,
		RepairSucceeded: false,
		Retryable:       class == workflowFailureToolGatewayTimeout || class == workflowFailureToolGatewayUnavailable,
	}
}

func workflowPhaseFailure(store storepkg.Store, traceID string, phaseKey string) (workflowFailure, bool) {
	traceID = strings.TrimSpace(traceID)
	phaseKey = strings.TrimSpace(phaseKey)
	if traceID == "" || phaseKey == "" {
		return workflowFailure{}, false
	}
	for _, intent := range store.ListActionIntents() {
		if strings.TrimSpace(intent.TraceID) != traceID || intent.OwnerPlane != "control" || intent.PhaseKey != phaseKey {
			continue
		}
		switch intent.Status {
		case action.StatusBlocked, action.StatusCanceled, action.StatusFailed:
			return workflowFailureFromContextAction(store, intent), true
		}
	}
	return workflowFailure{}, false
}

func activateDueWorkflowLineRetries(cfg config.Config, store storepkg.Store, now time.Time) error {
	if !cfg.WorkflowAutoRetryEnabled {
		return nil
	}
	for _, line := range store.ListWorkflowLines() {
		if line.Status != string(transition.WorkflowLineStateRetryScheduled) || strings.TrimSpace(line.CurrentWorkflowID) == "" {
			continue
		}
		if line.RetryAfter != nil && now.Before(*line.RetryAfter) {
			continue
		}
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflowLine,
			AggregateID: line.CaseID,
			CommandKind: string(transition.CommandWorkflowLineActivateRetry),
			CommandID:   "cmd-workflow-line:" + line.CaseID + ":activate-retry",
			Actor:       cfg.ServiceName,
			OccurredAt:  now,
		}); err != nil {
			return err
		}
		if err := startWorkflowViaCommand(cfg, store, line.CurrentWorkflowID, now, queue.WorkflowQueue); err != nil {
			return err
		}
		log.Printf("control-plane workflow retry_activated case=%s workflow=%s", line.CaseID, line.CurrentWorkflowID)
	}
	return nil
}

func boolValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func mapValue(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneStringAnyMap(typed)
	default:
		return nil
	}
}

func stringValueFromMap(values map[string]any, key string) string {
	if len(values) == 0 {
		return ""
	}
	return stringValue(values[key])
}

func cloneStringAnyMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
