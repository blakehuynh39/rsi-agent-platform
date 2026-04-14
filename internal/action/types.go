package action

import (
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
)

type Kind string

const (
	KindSlackPost        Kind = "slack_post"
	KindToolRead         Kind = "tool_read"
	KindToolMutation     Kind = "tool_mutation"
	KindSandboxLaunch    Kind = "sandbox_launch"
	KindDraftPROpen      Kind = "draft_pr_open"
	KindHarnessOverlay   Kind = "harness_overlay_activate"
	KindKnowledgePromote Kind = "knowledge_promote"
)

type Status string

const (
	StatusDrafted    Status = "drafted"
	StatusApproved   Status = "approved"
	StatusBlocked    Status = "blocked"
	StatusQueued     Status = "queued"
	StatusExecuting  Status = "executing"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusCanceled   Status = "canceled"
	StatusSuperseded Status = "superseded"
)

type Intent struct {
	ID                   string               `json:"id"`
	OperationID          string               `json:"operation_id,omitempty"`
	OwnerPlane           string               `json:"owner_plane"`
	ConversationID       string               `json:"conversation_id,omitempty"`
	CaseID               string               `json:"case_id,omitempty"`
	TraceID              string               `json:"trace_id,omitempty"`
	ProposalID           string               `json:"proposal_id,omitempty"`
	AttemptID            string               `json:"attempt_id,omitempty"`
	Kind                 Kind                 `json:"kind"`
	PhaseKey             string               `json:"phase_key,omitempty"`
	TargetRef            string               `json:"target_ref,omitempty"`
	RequestPayload       map[string]any       `json:"request_payload,omitempty"`
	IdempotencyKey       string               `json:"idempotency_key,omitempty"`
	ApprovalMode         string               `json:"approval_mode,omitempty"`
	ApprovalState        string               `json:"approval_state,omitempty"`
	PolicyVerdict        string               `json:"policy_verdict,omitempty"`
	Status               Status               `json:"status"`
	SupersededByActionID string               `json:"superseded_by_action_id,omitempty"`
	RequestedBy          string               `json:"requested_by,omitempty"`
	Rationale            string               `json:"rationale,omitempty"`
	EvidenceRefs         []events.EvidenceRef `json:"evidence_refs,omitempty"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

type Result struct {
	ID                 string    `json:"id"`
	OperationID        string    `json:"operation_id,omitempty"`
	ActionIntentID     string    `json:"action_intent_id"`
	AttemptID          string    `json:"attempt_id,omitempty"`
	AttemptNumber      int       `json:"attempt_number"`
	Executor           string    `json:"executor"`
	Provider           string    `json:"provider,omitempty"`
	ProviderRef        string    `json:"provider_ref,omitempty"`
	RequestArtifactID  string    `json:"request_artifact_id,omitempty"`
	ResponseArtifactID string    `json:"response_artifact_id,omitempty"`
	Status             Status    `json:"status"`
	ErrorCode          string    `json:"error_code,omitempty"`
	ErrorMessage       string    `json:"error_message,omitempty"`
	StartedAt          time.Time `json:"started_at"`
	CompletedAt        time.Time `json:"completed_at"`
}
