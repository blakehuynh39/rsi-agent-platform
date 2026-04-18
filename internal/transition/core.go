package transition

import "time"

type MachineKind string

const (
	MachineIngress          MachineKind = "ingress"
	MachineWorkflow         MachineKind = "workflow"
	MachineWorkflowLine     MachineKind = "workflow_line"
	MachineQuestionRun      MachineKind = "question_run"
	MachineProblemLine      MachineKind = "problem_line"
	MachineRuntimeDiagnosis MachineKind = "runtime_diagnosis"
	MachineProposalLine     MachineKind = "proposal_line"
	MachineAttempt          MachineKind = "attempt"
	MachineAction           MachineKind = "action_execution"
	MachineKnowledge        MachineKind = "knowledge_promotion"
	MachineHarness          MachineKind = "harness"
	MachineThreadPolicy     MachineKind = "thread_policy"
	MachineSettings         MachineKind = "platform_settings"
)

type DecisionKind string

const (
	DecisionAdvance DecisionKind = "advance"
	DecisionNoop    DecisionKind = "noop"
	DecisionReject  DecisionKind = "reject"
)

type EffectKind string

const (
	EffectInvokeAction               EffectKind = "invoke_action"
	EffectOpenWorkspace              EffectKind = "open_workspace"
	EffectInvokeRunner               EffectKind = "invoke_runner"
	EffectPostSlackReply             EffectKind = "post_slack_reply"
	EffectCompileInvestigationSpec   EffectKind = "compile_investigation_spec"
	EffectRefreshAlignmentLedger     EffectKind = "refresh_alignment_ledger"
	EffectCollectSeedEvidence        EffectKind = "collect_seed_evidence"
	EffectExpandEvidence             EffectKind = "expand_evidence"
	EffectReduceReply                EffectKind = "reduce_reply"
	EffectWorkspaceValidate          EffectKind = "workspace_validate"
	EffectObserveWorkspaceValidation EffectKind = "observe_workspace_validation"
	EffectOpenDraftPR                EffectKind = "open_draft_pr"
	EffectRecordOutcome              EffectKind = "record_outcome"
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

type CommandDescriptor struct {
	MachineKind MachineKind    `json:"machine_kind"`
	AggregateID string         `json:"aggregate_id"`
	CommandKind string         `json:"command_kind"`
	CommandID   string         `json:"command_id,omitempty"`
	Actor       string         `json:"actor,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
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
	Commands     []CommandDescriptor     `json:"commands,omitempty"`
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

type CommandReceipt struct {
	CommandID        string       `json:"command_id"`
	MachineKind      MachineKind  `json:"machine_kind"`
	AggregateID      string       `json:"aggregate_id"`
	CommandKind      string       `json:"command_kind"`
	CausationID      string       `json:"causation_id,omitempty"`
	Actor            string       `json:"actor,omitempty"`
	DecisionKind     DecisionKind `json:"decision_kind"`
	Reason           string       `json:"reason,omitempty"`
	AggregateVersion int64        `json:"aggregate_version,omitempty"`
	ResultRef        string       `json:"result_ref,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type EffectExecution struct {
	ID             string         `json:"id"`
	MachineKind    MachineKind    `json:"machine_kind"`
	AggregateID    string         `json:"aggregate_id"`
	AttemptID      string         `json:"attempt_id,omitempty"`
	EffectKind     EffectKind     `json:"effect_kind"`
	Status         EffectStatus   `json:"status"`
	Holder         string         `json:"holder,omitempty"`
	IdempotencyKey string         `json:"idempotency_key"`
	Payload        map[string]any `json:"payload,omitempty"`
	ResultRef      string         `json:"result_ref,omitempty"`
	LastError      string         `json:"last_error,omitempty"`
	RetryCount     int            `json:"retry_count,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	StartedAt      *time.Time     `json:"started_at,omitempty"`
	LeaseExpiresAt *time.Time     `json:"lease_expires_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
}
