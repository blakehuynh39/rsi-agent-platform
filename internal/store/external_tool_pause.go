package store

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ExternalToolApprovalStatus string

const (
	ExternalToolApprovalPending  ExternalToolApprovalStatus = "pending"
	ExternalToolApprovalApproved ExternalToolApprovalStatus = "approved"
	ExternalToolApprovalDenied   ExternalToolApprovalStatus = "denied"
	ExternalToolApprovalExpired  ExternalToolApprovalStatus = "expired"
)

type ExternalToolOutcome string

const (
	ExternalToolOutcomePending   ExternalToolOutcome = "pending"
	ExternalToolOutcomeSucceeded ExternalToolOutcome = "succeeded"
	ExternalToolOutcomeFailed    ExternalToolOutcome = "failed"
	ExternalToolOutcomeDenied    ExternalToolOutcome = "denied"
	ExternalToolOutcomeExpired   ExternalToolOutcome = "expired"
)

type ExternalToolResumeStatus string

const (
	ExternalToolResumeNotReady ExternalToolResumeStatus = "not_ready"
	ExternalToolResumeQueued   ExternalToolResumeStatus = "queued"
	ExternalToolResumeRunning  ExternalToolResumeStatus = "running"
	ExternalToolResumeResumed  ExternalToolResumeStatus = "resumed"
	ExternalToolResumeFailed   ExternalToolResumeStatus = "failed"
)

type ExternalToolPause struct {
	ID                      string                     `json:"id"`
	IdempotencyKey          string                     `json:"idempotency_key"`
	ConversationID          string                     `json:"conversation_id,omitempty"`
	WorkflowID              string                     `json:"workflow_id,omitempty"`
	TraceID                 string                     `json:"trace_id,omitempty"`
	OperationID             string                     `json:"operation_id,omitempty"`
	ExecutionID             string                     `json:"execution_id,omitempty"`
	HermesSessionID         string                     `json:"hermes_session_id,omitempty"`
	CanonicalToolName       string                     `json:"canonical_tool_name,omitempty"`
	TransportToolName       string                     `json:"transport_tool_name,omitempty"`
	ToolCallID              string                     `json:"tool_call_id,omitempty"`
	ArgsHash                string                     `json:"args_hash,omitempty"`
	DBReadRequestID         string                     `json:"db_read_request_id,omitempty"`
	SQLSHA256               string                     `json:"sql_sha256,omitempty"`
	ApprovalStatus          ExternalToolApprovalStatus `json:"approval_status"`
	ToolOutcome             ExternalToolOutcome        `json:"tool_outcome"`
	ResumeStatus            ExternalToolResumeStatus   `json:"resume_status"`
	ApprovalRef             string                     `json:"approval_ref,omitempty"`
	ResultRef               string                     `json:"result_ref,omitempty"`
	ExpiresAt               time.Time                  `json:"expires_at,omitempty"`
	PendingAssistantMessage map[string]interface{}     `json:"pending_assistant_message,omitempty"`
	TranscriptSnapshot      []map[string]interface{}   `json:"transcript_snapshot,omitempty"`
	ResumePayload           map[string]interface{}     `json:"resume_payload,omitempty"`
	ErrorMessage            string                     `json:"error_message,omitempty"`
	Metadata                map[string]interface{}     `json:"metadata,omitempty"`
	CreatedAt               time.Time                  `json:"created_at"`
	UpdatedAt               time.Time                  `json:"updated_at"`
}

type ExternalToolPauseCreateInput struct {
	IdempotencyKey          string
	ConversationID          string
	WorkflowID              string
	TraceID                 string
	OperationID             string
	ExecutionID             string
	HermesSessionID         string
	CanonicalToolName       string
	TransportToolName       string
	ToolCallID              string
	ArgsHash                string
	DBReadRequestID         string
	SQLSHA256               string
	ApprovalRef             string
	ResultRef               string
	ExpiresAt               time.Time
	PendingAssistantMessage map[string]interface{}
	TranscriptSnapshot      []map[string]interface{}
	ResumePayload           map[string]interface{}
	Metadata                map[string]interface{}
}

func NewExternalToolPause(input ExternalToolPauseCreateInput, now time.Time) (ExternalToolPause, error) {
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.WorkflowID = strings.TrimSpace(input.WorkflowID)
	input.HermesSessionID = strings.TrimSpace(input.HermesSessionID)
	input.TransportToolName = strings.TrimSpace(input.TransportToolName)
	input.ToolCallID = strings.TrimSpace(input.ToolCallID)
	if input.IdempotencyKey == "" {
		return ExternalToolPause{}, errors.New("external tool pause idempotency key is required")
	}
	if input.WorkflowID == "" {
		return ExternalToolPause{}, errors.New("external tool pause workflow_id is required")
	}
	if input.HermesSessionID == "" {
		return ExternalToolPause{}, errors.New("external tool pause hermes_session_id is required")
	}
	if input.TransportToolName == "" {
		return ExternalToolPause{}, errors.New("external tool pause transport_tool_name is required")
	}
	if input.ToolCallID == "" {
		return ExternalToolPause{}, errors.New("external tool pause tool_call_id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return ExternalToolPause{
		ID:                      "etpause_" + uuid.NewString(),
		IdempotencyKey:          input.IdempotencyKey,
		ConversationID:          strings.TrimSpace(input.ConversationID),
		WorkflowID:              input.WorkflowID,
		TraceID:                 strings.TrimSpace(input.TraceID),
		OperationID:             strings.TrimSpace(input.OperationID),
		ExecutionID:             strings.TrimSpace(input.ExecutionID),
		HermesSessionID:         input.HermesSessionID,
		CanonicalToolName:       strings.TrimSpace(input.CanonicalToolName),
		TransportToolName:       input.TransportToolName,
		ToolCallID:              input.ToolCallID,
		ArgsHash:                strings.TrimSpace(input.ArgsHash),
		DBReadRequestID:         strings.TrimSpace(input.DBReadRequestID),
		SQLSHA256:               strings.TrimSpace(input.SQLSHA256),
		ApprovalStatus:          ExternalToolApprovalPending,
		ToolOutcome:             ExternalToolOutcomePending,
		ResumeStatus:            ExternalToolResumeNotReady,
		ApprovalRef:             strings.TrimSpace(input.ApprovalRef),
		ResultRef:               strings.TrimSpace(input.ResultRef),
		ExpiresAt:               input.ExpiresAt,
		PendingAssistantMessage: cloneAnyMap(input.PendingAssistantMessage),
		TranscriptSnapshot:      cloneAnyMapSlice(input.TranscriptSnapshot),
		ResumePayload:           cloneAnyMap(input.ResumePayload),
		Metadata:                cloneAnyMap(input.Metadata),
		CreatedAt:               now,
		UpdatedAt:               now,
	}, nil
}

func ExternalToolPauseTerminalOutcome(outcome ExternalToolOutcome) bool {
	switch outcome {
	case ExternalToolOutcomeSucceeded, ExternalToolOutcomeFailed, ExternalToolOutcomeDenied, ExternalToolOutcomeExpired:
		return true
	default:
		return false
	}
}

func cloneExternalToolPause(item ExternalToolPause) ExternalToolPause {
	item.PendingAssistantMessage = cloneAnyMap(item.PendingAssistantMessage)
	item.TranscriptSnapshot = cloneAnyMapSlice(item.TranscriptSnapshot)
	item.ResumePayload = cloneAnyMap(item.ResumePayload)
	item.Metadata = cloneAnyMap(item.Metadata)
	return item
}

func cloneAnyMapSlice(in []map[string]interface{}) []map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(in))
	for _, item := range in {
		out = append(out, cloneAnyMap(item))
	}
	return out
}
