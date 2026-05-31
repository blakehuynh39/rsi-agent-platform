package control

import (
	"net/http"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const maxRuntimeObservationBatchSize = 256

type runtimeObservationRequest struct {
	ID              string         `json:"id,omitempty"`
	ExecutionID     string         `json:"execution_id"`
	OperationID     string         `json:"operation_id,omitempty"`
	TraceID         string         `json:"trace_id,omitempty"`
	WorkflowID      string         `json:"workflow_id,omitempty"`
	HermesSessionID string         `json:"hermes_session_id,omitempty"`
	Role            string         `json:"role,omitempty"`
	Phase           string         `json:"phase"`
	EventType       string         `json:"event_type"`
	Status          string         `json:"status,omitempty"`
	Seq             int            `json:"seq"`
	Payload         map[string]any `json:"payload,omitempty"`
	RecordedAt      string         `json:"recorded_at,omitempty"`
}

type runtimeObservationBatchRequest struct {
	Observations []runtimeObservationRequest `json:"observations"`
}

func recordRuntimeObservation(store storepkg.Repository, input runtimeObservationRequest) (int, map[string]any) {
	item := harness.ExecutionObservation{
		ID:              strings.TrimSpace(input.ID),
		ExecutionID:     strings.TrimSpace(input.ExecutionID),
		OperationID:     strings.TrimSpace(input.OperationID),
		TraceID:         strings.TrimSpace(input.TraceID),
		WorkflowID:      strings.TrimSpace(input.WorkflowID),
		HermesSessionID: strings.TrimSpace(input.HermesSessionID),
		Role:            strings.TrimSpace(input.Role),
		Phase:           strings.TrimSpace(input.Phase),
		EventType:       strings.TrimSpace(input.EventType),
		Status:          strings.TrimSpace(input.Status),
		Seq:             input.Seq,
		Payload:         cloneStringAnyMap(input.Payload),
	}
	if item.Payload == nil {
		item.Payload = map[string]any{}
	}
	recordedAt := strings.TrimSpace(input.RecordedAt)
	if recordedAt != "" {
		parsed, err := time.Parse(time.RFC3339, recordedAt)
		if err != nil {
			return http.StatusBadRequest, map[string]any{
				"error":       "recorded_at must be RFC3339",
				"recorded_at": recordedAt,
			}
		}
		item.RecordedAt = parsed.UTC()
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

func recordRuntimeObservationBatch(store storepkg.Repository, input runtimeObservationBatchRequest) (int, map[string]any) {
	if len(input.Observations) == 0 {
		return http.StatusBadRequest, map[string]any{"error": "observations must not be empty"}
	}
	if len(input.Observations) > maxRuntimeObservationBatchSize {
		return http.StatusBadRequest, map[string]any{
			"error": "observations batch is too large",
			"limit": maxRuntimeObservationBatchSize,
			"count": len(input.Observations),
		}
	}
	for index, observation := range input.Observations {
		status, out := recordRuntimeObservation(store, observation)
		if status >= http.StatusBadRequest {
			out["index"] = index
			return status, out
		}
	}
	return http.StatusOK, map[string]any{
		"status":   "ok",
		"recorded": len(input.Observations),
	}
}
