package transition

import "time"

type MachineKind string

const (
	MachineWorkflow     MachineKind = "workflow"
	MachineProposalLine MachineKind = "proposal_line"
	MachineAttempt      MachineKind = "attempt"
	MachineAction       MachineKind = "action_execution"
	MachineKnowledge    MachineKind = "knowledge_promotion"
	MachineHarness      MachineKind = "harness"
)

type DecisionKind string

const (
	DecisionAdvance DecisionKind = "advance"
	DecisionNoop    DecisionKind = "noop"
	DecisionReject  DecisionKind = "reject"
)

type EffectKind string

const (
	EffectQueueAttemptPhase EffectKind = "queue_attempt_phase"
	EffectOpenWorkspace     EffectKind = "open_workspace"
	EffectInvokeRunner      EffectKind = "invoke_runner"
	EffectWorkspaceValidate EffectKind = "workspace_validate"
	EffectOpenDraftPR       EffectKind = "open_draft_pr"
	EffectScheduleRetry     EffectKind = "schedule_retry"
	EffectRefreshProjection EffectKind = "refresh_projection"
)

type EffectStatus string

const (
	EffectQueued     EffectStatus = "queued"
	EffectRunning    EffectStatus = "running"
	EffectCompleted  EffectStatus = "completed"
	EffectFailed     EffectStatus = "failed"
	EffectCanceled   EffectStatus = "canceled"
	EffectSuperseded EffectStatus = "superseded"
)

type CommandEnvelope struct {
	MachineKind     MachineKind    `json:"machine_kind"`
	AggregateID     string         `json:"aggregate_id"`
	CommandKind     string         `json:"command_kind"`
	CommandID       string         `json:"command_id"`
	CausationID     string         `json:"causation_id,omitempty"`
	Actor           string         `json:"actor,omitempty"`
	OccurredAt      time.Time      `json:"occurred_at"`
	Payload         map[string]any `json:"payload,omitempty"`
	ExpectedVersion int64          `json:"expected_version,omitempty"`
}

type DomainEventDescriptor struct {
	Kind    string         `json:"kind"`
	Payload map[string]any `json:"payload,omitempty"`
}

type EffectRequest struct {
	Kind           EffectKind     `json:"kind"`
	Status         EffectStatus   `json:"status"`
	IdempotencyKey string         `json:"idempotency_key"`
	Payload        map[string]any `json:"payload,omitempty"`
}

type TransitionDecision struct {
	DecisionKind DecisionKind            `json:"decision_kind"`
	Reason       string                  `json:"reason,omitempty"`
	Events       []DomainEventDescriptor `json:"events,omitempty"`
	Effects      []EffectRequest         `json:"effects,omitempty"`
}

type DomainEvent struct {
	ID               string         `json:"id"`
	MachineKind      MachineKind    `json:"machine_kind"`
	AggregateID      string         `json:"aggregate_id"`
	AggregateVersion int64          `json:"aggregate_version"`
	EventKind        string         `json:"event_kind"`
	CommandID        string         `json:"command_id,omitempty"`
	CausationID      string         `json:"causation_id,omitempty"`
	Payload          map[string]any `json:"payload,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

type EffectExecution struct {
	ID             string         `json:"id"`
	MachineKind    MachineKind    `json:"machine_kind"`
	AggregateID    string         `json:"aggregate_id"`
	AttemptID      string         `json:"attempt_id,omitempty"`
	EffectKind     EffectKind     `json:"effect_kind"`
	Status         EffectStatus   `json:"status"`
	IdempotencyKey string         `json:"idempotency_key"`
	Payload        map[string]any `json:"payload,omitempty"`
	ResultRef      string         `json:"result_ref,omitempty"`
	LastError      string         `json:"last_error,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	StartedAt      *time.Time     `json:"started_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
}
