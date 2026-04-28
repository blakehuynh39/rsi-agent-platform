package events

import (
	"strings"
	"time"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusRunning    Status = "running"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusNeedsHuman Status = "needs-human"
	StatusSuppressed Status = "suppressed"
	StatusReplayed   Status = "replayed"
	StatusInReview   Status = "in-review"
	StatusDraft      Status = "draft"
)

type TraceEvent struct {
	TraceID        string     `json:"trace_id"`
	IngestionID    string     `json:"ingestion_id"`
	WorkflowID     string     `json:"workflow_id"`
	ConversationID string     `json:"conversation_id,omitempty"`
	CaseID         string     `json:"case_id,omitempty"`
	TriggerEventID string     `json:"trigger_event_id,omitempty"`
	ParentEvent    string     `json:"parent_event_id,omitempty"`
	Plane          string     `json:"plane"`
	Service        string     `json:"service"`
	Actor          string     `json:"actor"`
	EventType      string     `json:"event_type"`
	Status         Status     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
	PayloadRef     string     `json:"payload_ref,omitempty"`
	ArtifactRef    string     `json:"artifact_ref,omitempty"`
	CostTokens     int        `json:"cost_tokens,omitempty"`
	LatencyMs      int64      `json:"latency_ms,omitempty"`
	Description    string     `json:"description,omitempty"`
}

type Artifact struct {
	ID          string `json:"id"`
	TraceID     string `json:"trace_id"`
	Kind        string `json:"kind"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	SizeBytes   int64  `json:"size_bytes"`
	Source      string `json:"source"`
}

type EvidenceRef struct {
	Kind    string `json:"kind"`
	Ref     string `json:"ref"`
	Summary string `json:"summary,omitempty"`
}

type ReasoningStep struct {
	ID             string        `json:"id"`
	TraceID        string        `json:"trace_id"`
	WorkflowID     string        `json:"workflow_id,omitempty"`
	ConversationID string        `json:"conversation_id,omitempty"`
	CaseID         string        `json:"case_id,omitempty"`
	StepType       string        `json:"step_type"`
	Summary        string        `json:"summary"`
	EvidenceRefs   []EvidenceRef `json:"evidence_refs,omitempty"`
	Alternatives   []string      `json:"alternatives,omitempty"`
	Confidence     float64       `json:"confidence,omitempty"`
	Decision       string        `json:"decision,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

type ToolCallRecord struct {
	ID                    string                 `json:"id"`
	TraceID               string                 `json:"trace_id"`
	WorkflowID            string                 `json:"workflow_id,omitempty"`
	ConversationID        string                 `json:"conversation_id,omitempty"`
	CaseID                string                 `json:"case_id,omitempty"`
	ToolName              string                 `json:"tool_name"`
	ToolCallID            string                 `json:"tool_call_id"`
	Request               map[string]interface{} `json:"request,omitempty"`
	Summary               string                 `json:"summary,omitempty"`
	RawArtifactRefs       []string               `json:"raw_artifact_refs,omitempty"`
	ApprovalState         string                 `json:"approval_state,omitempty"`
	InterpretationSummary string                 `json:"interpretation_summary,omitempty"`
	Status                string                 `json:"status,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
}

type SlackActionRecord struct {
	ID             string    `json:"id"`
	TraceID        string    `json:"trace_id"`
	WorkflowID     string    `json:"workflow_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	CaseID         string    `json:"case_id,omitempty"`
	ChannelID      string    `json:"channel_id,omitempty"`
	ThreadTS       string    `json:"thread_ts,omitempty"`
	IdempotencyKey string    `json:"idempotency_key"`
	DraftBody      string    `json:"draft_body,omitempty"`
	FinalBody      string    `json:"final_body,omitempty"`
	PolicyVerdict  string    `json:"policy_verdict,omitempty"`
	SendStatus     string    `json:"send_status,omitempty"`
	ArtifactRefs   []string  `json:"artifact_refs,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type ExecutionLedgerEvent struct {
	ID             string         `json:"id"`
	ExecutionID    string         `json:"execution_id"`
	OperationID    string         `json:"operation_id,omitempty"`
	TraceID        string         `json:"trace_id,omitempty"`
	WorkflowID     string         `json:"workflow_id,omitempty"`
	PhaseID        string         `json:"phase_id,omitempty"`
	Kind           string         `json:"kind"`
	Status         string         `json:"status,omitempty"`
	Seq            int            `json:"seq"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
	RecordedAt     time.Time      `json:"recorded_at"`
}

type TraceSummary struct {
	TraceID            string    `json:"trace_id"`
	IngestionID        string    `json:"ingestion_id"`
	WorkflowID         string    `json:"workflow_id"`
	ConversationID     string    `json:"conversation_id,omitempty"`
	CaseID             string    `json:"case_id,omitempty"`
	TriggerEventID     string    `json:"trigger_event_id,omitempty"`
	SupersedesTraceID  string    `json:"supersedes_trace_id,omitempty"`
	ThreadKey          string    `json:"thread_key"`
	WorkflowKind       string    `json:"workflow_kind"`
	Status             Status    `json:"status"`
	LastVerdict        string    `json:"last_verdict,omitempty"`
	StartedAt          time.Time `json:"started_at"`
	EndedAt            time.Time `json:"ended_at"`
	EventCount         int       `json:"event_count"`
	ArtifactCount      int       `json:"artifact_count"`
	ReasoningStepCount int       `json:"reasoning_step_count"`
	ToolCallCount      int       `json:"tool_call_count"`
	SlackActionCount   int       `json:"slack_action_count"`
}

type Trace struct {
	Summary      TraceSummary        `json:"summary"`
	Events       []TraceEvent        `json:"events"`
	Artifacts    []Artifact          `json:"artifacts"`
	Reasoning    []ReasoningStep     `json:"reasoning"`
	ToolCalls    []ToolCallRecord    `json:"tool_calls"`
	SlackActions []SlackActionRecord `json:"slack_actions"`
}

// SlackDeliveryStatusSucceeded returns true if the status indicates a successful delivery.
func SlackDeliveryStatusSucceeded(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "posted", "sent", "uploaded", "completed", "ok", "success", "shared":
		return true
	default:
		return false
	}
}
