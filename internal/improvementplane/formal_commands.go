package improvementplane

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func submitProblemLineCommand(store storepkg.Store, aggregateID string, kind transition.ProblemLineCommandKind, actor string, occurredAt time.Time, commandID string, payload map[string]any) (transition.CommandReceipt, error) {
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: strings.TrimSpace(aggregateID),
		CommandKind: string(kind),
		CommandID:   strings.TrimSpace(commandID),
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		return transition.CommandReceipt{}, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return transition.CommandReceipt{}, errors.New(receipt.Reason)
	}
	return receipt, nil
}

func loadEvalRunByID(store storepkg.Store, runID string) (evals.Run, bool) {
	runID = strings.TrimSpace(runID)
	for _, run := range store.ListEvalRuns() {
		if run.ID == runID {
			return run, true
		}
	}
	return evals.Run{}, false
}

func loadOutcomeFromReceipt(store storepkg.Store, receipt transition.CommandReceipt) (outcome.Record, error) {
	outcomeID := strings.TrimSpace(receipt.ResultRef)
	if outcomeID == "" {
		return outcome.Record{}, errors.New("missing outcome result ref")
	}
	for _, item := range store.ListOutcomes() {
		if item.ID == outcomeID {
			return item, nil
		}
	}
	return outcome.Record{}, fmt.Errorf("outcome %s not found", outcomeID)
}

func loadPromotionResultFromReceipt(store storepkg.Store, receipt transition.CommandReceipt) (storepkg.PromotionResult, error) {
	if receipt.DecisionKind == transition.DecisionNoop {
		slots := store.GetProposalSlots()
		return storepkg.PromotionResult{
			BlockedByCap:     slots.Available == 0,
			StaleProposalIDs: slots.StaleProposalIDs,
		}, nil
	}
	event, ok := findCommandDomainEvent(store, receipt.CommandID, "problem_line_promoted")
	if !ok {
		return storepkg.PromotionResult{}, errors.New("promotion result event not found")
	}
	return storepkg.PromotionResult{
		Promoted:         int(intFromAny(event.Payload["promoted"])),
		PromotedIDs:      stringSliceFromAny(event.Payload["promoted_ids"]),
		BlockedByCap:     boolFromAny(event.Payload["blocked_by_cap"]),
		StaleProposalIDs: stringSliceFromAny(event.Payload["stale_proposal_ids"]),
	}, nil
}

func submitProblemLineTraceProjection(store storepkg.Store, traceID string, actor string, occurredAt time.Time, commandID string, update storepkg.TraceUpdate) (transition.CommandReceipt, error) {
	payload := map[string]any{
		"trace_id":        traceID,
		"trace_events":    update.Events,
		"trace_artifacts": update.Artifacts,
		"reasoning_steps": update.Reasoning,
	}
	if update.Status != nil {
		payload["trace_status"] = string(*update.Status)
	}
	if update.LastVerdict != nil {
		payload["last_verdict"] = *update.LastVerdict
	}
	if strings.TrimSpace(update.WorkflowStatus) != "" {
		payload["workflow_status"] = update.WorkflowStatus
	}
	if strings.TrimSpace(update.WorkflowError) != "" {
		payload["workflow_error"] = update.WorkflowError
	}
	if len(update.ToolCalls) > 0 {
		payload["tool_calls"] = update.ToolCalls
	}
	if len(update.SlackActions) > 0 {
		payload["slack_actions"] = update.SlackActions
	}
	return submitProblemLineCommand(
		store,
		traceID,
		transition.CommandProblemLineProjectTrace,
		actor,
		occurredAt,
		commandID,
		payload,
	)
}

func submitHarnessActivationCommand(store storepkg.Store, overlay harness.Overlay, experiment harness.Experiment, actor string, occurredAt time.Time, commandID string) (transition.CommandReceipt, error) {
	payload := map[string]any{
		"profile_id":            overlay.ProfileID,
		"role":                  overlay.Role,
		"version":               overlay.Version,
		"status":                string(overlay.Status),
		"target_kind":           overlay.TargetKind,
		"target_ref":            overlay.TargetRef,
		"proposal_id":           overlay.ProposalID,
		"prompt_fragments":      overlay.PromptFragments,
		"few_shot_snippets":     overlay.FewShotSnippets,
		"tool_preference_order": overlay.ToolPreferenceOrder,
		"retrieval_bias":        overlay.RetrievalBias,
		"reasoning_verbosity":   overlay.ReasoningVerbosity,
		"created_by":            overlay.CreatedBy,
		"approved_by":           overlay.ApprovedBy,
		"experiment_id":         experiment.ID,
		"attempt_id":            experiment.AttemptID,
		"experiment_status":     string(experiment.Status),
		"experiment_summary":    experiment.Summary,
		"experiment_metrics":    experiment.Metrics,
	}
	if overlay.MemoryReadEnabled != nil {
		payload["memory_read_enabled"] = *overlay.MemoryReadEnabled
	}
	if overlay.MemoryWriteEnabled != nil {
		payload["memory_write_enabled"] = *overlay.MemoryWriteEnabled
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineHarness,
		AggregateID: overlay.ID,
		CommandKind: string(transition.CommandHarnessActivateOverlay),
		CommandID:   commandID,
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		return transition.CommandReceipt{}, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return transition.CommandReceipt{}, errors.New(receipt.Reason)
	}
	return receipt, nil
}

func findCommandDomainEvent(store storepkg.Store, commandID string, eventKind string) (transition.DomainEvent, bool) {
	commandID = strings.TrimSpace(commandID)
	eventKind = strings.TrimSpace(eventKind)
	for _, item := range store.ListDomainEvents() {
		if strings.TrimSpace(item.CommandID) != commandID {
			continue
		}
		if eventKind != "" && strings.TrimSpace(item.EventKind) != eventKind {
			continue
		}
		return item, true
	}
	return transition.DomainEvent{}, false
}

func traceEventSummary(events []events.TraceEvent) string {
	if len(events) == 0 {
		return ""
	}
	return strings.TrimSpace(events[len(events)-1].Description)
}

func boolFromAny(raw any) bool {
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(strings.TrimSpace(value), "true")
	default:
		return false
	}
}

func intFromAny(raw any) int64 {
	switch value := raw.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case int32:
		return int64(value)
	case float64:
		return int64(value)
	case float32:
		return int64(value)
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return 0
		}
		var out int64
		fmt.Sscanf(trimmed, "%d", &out)
		return out
	default:
		return 0
	}
}
