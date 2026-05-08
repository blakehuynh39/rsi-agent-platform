package control

import (
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func handleExternalToolPendingRunnerResult(cfg config.Config, store storepkg.Store, ctx workflowContext, effect transition.EffectExecution, runnerResp clients.RunnerResponse, runnerStarted time.Time) error {
	pauseID := externalToolPauseIDFromRunnerRaw(runnerResp.Raw)
	if pauseID == "" {
		return &workflowFailureError{failure: workflowFailureFromRunnerPostProcessing(runnerResp, "external_tool_pending_missing_pause_id", fmt.Errorf("external tool pending result did not include pause id"))}
	}
	pause, ok := store.GetExternalToolPause(pauseID)
	if !ok {
		return &workflowFailureError{failure: workflowFailureFromRunnerPostProcessing(runnerResp, "external_tool_pending_unknown_pause", fmt.Errorf("external tool pause %s not found", pauseID))}
	}
	messages := runnerMessagesFromRaw(runnerResp.Raw)
	pendingAssistant := pendingAssistantMessageForToolCall(messages, pause.ToolCallID)
	updatedPause, err := store.UpdateExternalToolPause(pause.ID, func(item *storepkg.ExternalToolPause) error {
		item.ExecutionID = firstNonEmpty(item.ExecutionID, stringValue(runnerResp.Raw["execution_id"]))
		item.PendingAssistantMessage = pendingAssistant
		item.TranscriptSnapshot = messages
		if item.Metadata == nil {
			item.Metadata = map[string]any{}
		}
		item.Metadata["runner_started_at"] = runnerStarted.Format(time.RFC3339Nano)
		item.Metadata["runner_provider"] = runnerResp.Provider
		return nil
	})
	if err != nil {
		return &workflowFailureError{failure: workflowFailureFromRunnerPostProcessing(runnerResp, "external_tool_pause_update", err)}
	}
	if storepkg.ExternalToolPauseTerminalOutcome(updatedPause.ToolOutcome) && strings.TrimSpace(updatedPause.DBReadRequestID) != "" {
		if request, found := store.GetDBReadRequest(updatedPause.DBReadRequestID); found {
			refreshedPayload := buildDBReadExternalToolResumePayload(updatedPause, request, updatedPause.ToolOutcome, updatedPause.ErrorMessage)
			updatedPause, _ = store.UpdateExternalToolPause(updatedPause.ID, func(item *storepkg.ExternalToolPause) error {
				item.ResumePayload = refreshedPayload
				return nil
			})
		}
	}
	now := time.Now().UTC()
	payload := map[string]any{
		"external_tool_pause_id": pause.ID,
		"termination_reason":     "external_tool_pending",
		"runner_diagnostics":     mergeWorkflowRunnerDiagnostics(cloneStringAnyMap(mapValue(runnerResp.Raw["runner_diagnostics"])), runnerResp.Raw),
		"trace_events": []events.TraceEvent{{
			TraceID:     ctx.trace.Summary.TraceID,
			IngestionID: ctx.trace.Summary.IngestionID,
			WorkflowID:  ctx.workflow.ID,
			Plane:       "execution",
			Service:     "runner",
			Actor:       ctx.workflow.AssignedBot,
			EventType:   "external_tool.pending",
			Status:      events.StatusRunning,
			StartedAt:   runnerStarted,
			EndedAt:     &now,
			Description: "Hermes paused at an external tool call awaiting approval or result.",
		}},
		"reasoning_steps": []events.ReasoningStep{{
			ID:         fmt.Sprintf("reason-external-tool-pending-%d", now.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.workflow.ID,
			StepType:   "external_tool_pending",
			Summary:    "Hermes paused before appending a tool result so RSI can approve and execute the external tool.",
			Confidence: 1.0,
			Decision:   pause.ID,
			CreatedAt:  now,
		}},
	}
	command := transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: ctx.workflow.ID,
		CommandKind: string(transition.CommandWorkflowWaitingExternalTool),
		CommandID:   fmt.Sprintf("cmd-workflow:%s:%s:%s", ctx.workflow.ID, transition.CommandWorkflowWaitingExternalTool, pause.ID),
		Actor:       firstNonEmpty(cfg.ServiceName, "control-plane"),
		OccurredAt:  now,
		Payload:     payload,
	}
	if _, err := store.SubmitCommand(command); err != nil {
		return &workflowFailureError{failure: workflowFailureFromRunnerPostProcessing(runnerResp, "submit_external_tool_pending", err)}
	}
	if _, _, err := store.ReconcileWorkflowTrace(ctx.workflow.ID); err != nil {
		return err
	}
	tryQueueExternalToolResume(cfg, store, pause.ID)
	return completeClaimedEffect(store, effect, fmt.Sprintf("trace:%s:external-tool-pending:%s", ctx.trace.Summary.TraceID, pause.ID))
}

func markDBReadExternalToolOutcome(cfg config.Config, store storepkg.Store, request storepkg.DBReadRequest, outcome storepkg.ExternalToolOutcome, message string) {
	if !cfg.ExternalToolResumeEnabled {
		return
	}
	pause, ok := store.GetExternalToolPauseByDBReadRequestID(request.ID)
	if !ok {
		return
	}
	resumePayload := buildDBReadExternalToolResumePayload(pause, request, outcome, message)
	_, err := store.UpdateExternalToolPause(pause.ID, func(item *storepkg.ExternalToolPause) error {
		switch outcome {
		case storepkg.ExternalToolOutcomeSucceeded:
			item.ApprovalStatus = storepkg.ExternalToolApprovalApproved
		case storepkg.ExternalToolOutcomeDenied:
			item.ApprovalStatus = storepkg.ExternalToolApprovalDenied
		case storepkg.ExternalToolOutcomeExpired:
			item.ApprovalStatus = storepkg.ExternalToolApprovalExpired
		case storepkg.ExternalToolOutcomeFailed:
			if item.ApprovalStatus == "" {
				item.ApprovalStatus = storepkg.ExternalToolApprovalPending
			}
		}
		item.ToolOutcome = outcome
		item.ResultRef = firstNonEmpty(request.ResultArtifactRef, request.ID)
		item.ResumePayload = resumePayload
		item.ErrorMessage = strings.TrimSpace(message)
		if item.Metadata == nil {
			item.Metadata = map[string]any{}
		}
		item.Metadata["db_read_state"] = string(request.State)
		item.Metadata["row_count"] = request.RowCount
		item.Metadata["truncated"] = request.Truncated
		return nil
	})
	if err != nil {
		return
	}
	tryQueueExternalToolResume(cfg, store, pause.ID)
}

func buildDBReadExternalToolResumePayload(pause storepkg.ExternalToolPause, request storepkg.DBReadRequest, outcome storepkg.ExternalToolOutcome, message string) map[string]any {
	status := "ok"
	if outcome != storepkg.ExternalToolOutcomeSucceeded {
		status = string(outcome)
	}
	content := map[string]any{
		"kind":          "db_read_result",
		"status":        status,
		"request_id":    request.ID,
		"target":        request.Target,
		"sql_sha256":    request.SQLSHA256,
		"row_count":     request.RowCount,
		"truncated":     request.Truncated,
		"result_ref":    firstNonEmpty(request.ResultArtifactRef, request.ID),
		"sample":        request.ResultSample,
		"error_message": strings.TrimSpace(firstNonEmpty(message, request.ErrorMessage)),
	}
	payload := map[string]any{
		"kind":         "external_tool_result",
		"session_id":   pause.HermesSessionID,
		"tool_call_id": pause.ToolCallID,
		"tool_name":    pause.TransportToolName,
		"status":       status,
		"content":      content,
		"metadata": map[string]any{
			"external_tool_pause_id": pause.ID,
			"canonical_tool_name":    pause.CanonicalToolName,
			"db_read_request_id":     request.ID,
			"approval_status":        string(pause.ApprovalStatus),
			"tool_outcome":           string(outcome),
		},
	}
	if len(pause.TranscriptSnapshot) > 0 {
		payload["transcript_snapshot"] = pause.TranscriptSnapshot
	}
	return payload
}

func tryQueueExternalToolResume(cfg config.Config, store storepkg.Store, pauseID string) {
	if !cfg.ExternalToolResumeEnabled {
		return
	}
	pause, ok := store.GetExternalToolPause(pauseID)
	if !ok || !storepkg.ExternalToolPauseTerminalOutcome(pause.ToolOutcome) || pause.ResumeStatus != storepkg.ExternalToolResumeNotReady {
		return
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), pause.WorkflowID)
	if !ok || strings.TrimSpace(workflow.Status) != string(transition.WorkflowStateWaitingExternalTool) {
		return
	}
	if len(pause.ResumePayload) == 0 {
		_, _ = store.UpdateExternalToolPause(pause.ID, func(item *storepkg.ExternalToolPause) error {
			item.ResumeStatus = storepkg.ExternalToolResumeFailed
			item.ErrorMessage = firstNonEmpty(item.ErrorMessage, "external tool result was missing resume payload")
			return nil
		})
		return
	}
	now := time.Now().UTC()
	command := transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: pause.WorkflowID,
		CommandKind: string(transition.CommandExternalToolResultReady),
		CommandID:   fmt.Sprintf("cmd-workflow:%s:%s:%s", pause.WorkflowID, transition.CommandExternalToolResultReady, pause.ID),
		Actor:       firstNonEmpty(cfg.ServiceName, "control-plane"),
		OccurredAt:  now,
		Payload: map[string]any{
			"external_tool_pause_id": pause.ID,
			"external_tool_resume":   pause.ResumePayload,
			"trace_events": []events.TraceEvent{{
				TraceID:     pause.TraceID,
				WorkflowID:  pause.WorkflowID,
				Plane:       "control",
				Service:     firstNonEmpty(cfg.ServiceName, "control-plane"),
				Actor:       "db-read-worker",
				EventType:   "external_tool.result_ready",
				Status:      events.StatusRunning,
				StartedAt:   now,
				Description: "External tool result is ready; queueing Hermes resume.",
			}},
		},
	}
	if _, err := store.SubmitCommand(command); err != nil {
		_, _ = store.UpdateExternalToolPause(pause.ID, func(item *storepkg.ExternalToolPause) error {
			item.ErrorMessage = err.Error()
			return nil
		})
		return
	}
	_, _ = store.UpdateExternalToolPause(pause.ID, func(item *storepkg.ExternalToolPause) error {
		if item.ResumeStatus == storepkg.ExternalToolResumeNotReady {
			item.ResumeStatus = storepkg.ExternalToolResumeQueued
		}
		return nil
	})
}

func externalToolPauseIDFromRunnerRaw(raw map[string]interface{}) string {
	if value := strings.TrimSpace(stringValue(raw["external_tool_pause_id"])); value != "" {
		return value
	}
	pending := mapValue(raw["external_tool_pending"])
	return firstNonEmpty(
		stringValueFromMap(pending, "external_tool_pause_id"),
		stringValueFromMap(pending, "external_action_id"),
		stringValueFromMap(pending, "pause_id"),
	)
}

func runnerMessagesFromRaw(raw map[string]interface{}) []map[string]interface{} {
	result := mapValue(raw["result"])
	value, ok := result["messages"].([]interface{})
	if !ok {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(value))
	for _, item := range value {
		if mapped, ok := item.(map[string]interface{}); ok {
			out = append(out, cloneStringAnyMap(mapped))
		}
	}
	return out
}

func pendingAssistantMessageForToolCall(messages []map[string]interface{}, toolCallID string) map[string]interface{} {
	toolCallID = strings.TrimSpace(toolCallID)
	if toolCallID == "" {
		return nil
	}
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if strings.TrimSpace(stringValue(msg["role"])) != "assistant" {
			continue
		}
		for _, toolCall := range interfaceSlice(msg["tool_calls"]) {
			mapped, ok := toolCall.(map[string]interface{})
			if !ok {
				continue
			}
			if strings.TrimSpace(stringValue(mapped["id"])) == toolCallID {
				return cloneStringAnyMap(msg)
			}
		}
	}
	return nil
}

func interfaceSlice(value interface{}) []interface{} {
	if items, ok := value.([]interface{}); ok {
		return items
	}
	return nil
}
