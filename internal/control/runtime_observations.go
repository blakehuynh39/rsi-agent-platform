package control

import (
	"net/http"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func recordRuntimeObservation(store storepkg.Repository, input map[string]any) (int, map[string]any) {
	item := harness.ExecutionObservation{
		ID:              strings.TrimSpace(stringValue(input["id"])),
		ExecutionID:     strings.TrimSpace(stringValue(input["execution_id"])),
		OperationID:     strings.TrimSpace(stringValue(input["operation_id"])),
		TraceID:         strings.TrimSpace(stringValue(input["trace_id"])),
		WorkflowID:      strings.TrimSpace(stringValue(input["workflow_id"])),
		HermesSessionID: strings.TrimSpace(stringValue(input["hermes_session_id"])),
		Role:            strings.TrimSpace(stringValue(input["role"])),
		Phase:           strings.TrimSpace(stringValue(input["phase"])),
		EventType:       strings.TrimSpace(stringValue(input["event_type"])),
		Status:          strings.TrimSpace(stringValue(input["status"])),
		Seq:             intValue(input["seq"]),
		Payload:         mapValue(input["payload"]),
	}
	if item.Payload == nil {
		item.Payload = map[string]any{}
	}
	recordedAt := strings.TrimSpace(stringValue(input["recorded_at"]))
	if recordedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, recordedAt); err == nil {
			item.RecordedAt = parsed.UTC()
		}
	}
	if item.ExecutionID == "" || item.Phase == "" || item.EventType == "" {
		return http.StatusBadRequest, map[string]any{
			"error":        "missing required observation fields",
			"execution_id": item.ExecutionID,
			"phase":        item.Phase,
			"event_type":   item.EventType,
		}
	}
	if item.Seq <= 0 {
		return http.StatusBadRequest, map[string]any{
			"error": "observation seq must be greater than zero",
			"seq":   item.Seq,
		}
	}
	recorded, err := store.RecordHarnessExecutionObservation(item)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}
	ledgerPayload := cloneStringAnyMap(recorded.Payload)
	if ledgerPayload == nil {
		ledgerPayload = map[string]any{}
	}
	ledgerPayload["observation_id"] = recorded.ID
	if recorded.Role != "" {
		ledgerPayload["role"] = recorded.Role
	}
	if recorded.HermesSessionID != "" {
		ledgerPayload["hermes_session_id"] = recorded.HermesSessionID
	}
	ledgerEvent := events.ExecutionLedgerEvent{
		ExecutionID:    recorded.ExecutionID,
		OperationID:    recorded.OperationID,
		TraceID:        recorded.TraceID,
		WorkflowID:     recorded.WorkflowID,
		PhaseID:        recorded.Phase,
		Kind:           recorded.EventType,
		Status:         recorded.Status,
		Seq:            recorded.Seq,
		IdempotencyKey: strings.TrimSpace(stringValue(ledgerPayload["idempotency_key"])),
		Payload:        ledgerPayload,
		RecordedAt:     recorded.RecordedAt,
	}
	if err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{ledgerEvent}); err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}
	return http.StatusOK, map[string]any{
		"status":       "ok",
		"observation":  recorded,
		"ledger_event": ledgerEvent,
	}
}

func intValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}
