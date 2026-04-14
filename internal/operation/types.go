package operation

import (
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/queue"
)

type ScopeKind string

const (
	ScopeTrace       ScopeKind = "trace"
	ScopeProposal    ScopeKind = "proposal"
	ScopeAttempt     ScopeKind = "attempt"
	ScopeAction      ScopeKind = "action_intent"
	ScopeEvent       ScopeKind = "external_event"
	ScopeHarnessRun  ScopeKind = "harness_run"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusRunning    Status = "running"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCanceled   Status = "canceled"
	StatusSuperseded Status = "superseded"
)

type Execution struct {
	ID           string         `json:"id"`
	ScopeKind    ScopeKind      `json:"scope_kind"`
	ScopeID      string         `json:"scope_id"`
	OperationKind string        `json:"operation_kind"`
	OperationKey string         `json:"operation_key"`
	Status       Status         `json:"status"`
	Queue        queue.QueueName `json:"queue"`
	RequestedBy  string         `json:"requested_by,omitempty"`
	Holder       string         `json:"holder,omitempty"`
	TraceID      string         `json:"trace_id,omitempty"`
	ProposalID   string         `json:"proposal_id,omitempty"`
	AttemptID    string         `json:"attempt_id,omitempty"`
	PayloadHash  string         `json:"payload_hash,omitempty"`
	ResultRef    string         `json:"result_ref,omitempty"`
	LastError    string         `json:"last_error,omitempty"`
	RetryCount   int            `json:"retry_count,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
}
