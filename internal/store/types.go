package store

import (
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/questionrun"
)

type Workflow struct {
	ID                string         `json:"id"`
	IngestionID       string         `json:"ingestion_id,omitempty"`
	TraceID           string         `json:"trace_id,omitempty"`
	ConversationID    string         `json:"conversation_id,omitempty"`
	CaseID            string         `json:"case_id,omitempty"`
	ThreadKey         string         `json:"thread_key"`
	Kind              string         `json:"kind"`
	Intent            string         `json:"intent,omitempty"`
	AssignedBot       string         `json:"assigned_bot"`
	ApprovalMode      string         `json:"approval_mode,omitempty"`
	ResponseMode      string         `json:"response_mode,omitempty"`
	Status            string         `json:"status"`
	LastVerdict       string         `json:"last_verdict,omitempty"`
	LastError         string         `json:"last_error,omitempty"`
	AttemptNumber     int            `json:"attempt_number,omitempty"`
	ParentWorkflowID  string         `json:"parent_workflow_id,omitempty"`
	FailureClass      string         `json:"failure_class,omitempty"`
	FailureSummary    string         `json:"failure_summary,omitempty"`
	RetryDecision     string         `json:"retry_decision,omitempty"`
	RetryAfter        *time.Time     `json:"retry_after,omitempty"`
	RunnerDiagnostics map[string]any `json:"runner_diagnostics,omitempty"`
	RepairAttempted   bool           `json:"repair_attempted,omitempty"`
	RepairSucceeded   bool           `json:"repair_succeeded,omitempty"`
	Version           int64          `json:"version"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty"`
}

type QuestionRun struct {
	ID                string                        `json:"id"`
	WorkflowID        string                        `json:"workflow_id"`
	TraceID           string                        `json:"trace_id,omitempty"`
	ConversationID    string                        `json:"conversation_id,omitempty"`
	CaseID            string                        `json:"case_id,omitempty"`
	IngestionID       string                        `json:"ingestion_id,omitempty"`
	Role              string                        `json:"role,omitempty"`
	Strategy          string                        `json:"strategy,omitempty"`
	Status            string                        `json:"status"`
	InvestigationSpec questionrun.InvestigationSpec `json:"investigation_spec,omitempty"`
	EvidenceLedger    questionrun.EvidenceLedger    `json:"evidence_ledger,omitempty"`
	Result            questionrun.Result            `json:"result,omitempty"`
	FailureClass      string                        `json:"failure_class,omitempty"`
	FailureSummary    string                        `json:"failure_summary,omitempty"`
	LastError         string                        `json:"last_error,omitempty"`
	RunnerDiagnostics map[string]any                `json:"runner_diagnostics,omitempty"`
	Version           int64                         `json:"version"`
	CreatedAt         time.Time                     `json:"created_at"`
	UpdatedAt         time.Time                     `json:"updated_at"`
	CompletedAt       *time.Time                    `json:"completed_at,omitempty"`
}

type RunnerExecution struct {
	ExecutionID        string         `json:"execution_id"`
	OperationID        string         `json:"operation_id,omitempty"`
	WorkflowID         string         `json:"workflow_id,omitempty"`
	TraceID            string         `json:"trace_id,omitempty"`
	ConversationID     string         `json:"conversation_id,omitempty"`
	CaseID             string         `json:"case_id,omitempty"`
	Role               string         `json:"role,omitempty"`
	ExecutorInstanceID string         `json:"executor_instance_id,omitempty"`
	ExecutorBaseURL    string         `json:"executor_base_url,omitempty"`
	Status             string         `json:"status"`
	Task               map[string]any `json:"task,omitempty"`
	Result             map[string]any `json:"result,omitempty"`
	FailureClass       string         `json:"failure_class,omitempty"`
	Holder             string         `json:"holder,omitempty"`
	RetryCount         int            `json:"retry_count,omitempty"`
	CancelRequested    bool           `json:"cancel_requested,omitempty"`
	HeartbeatAt        *time.Time     `json:"heartbeat_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	StartedAt          *time.Time     `json:"started_at,omitempty"`
	CompletedAt        *time.Time     `json:"completed_at,omitempty"`
}

type WorkflowLine struct {
	CaseID                   string     `json:"case_id"`
	ConversationID           string     `json:"conversation_id"`
	Status                   string     `json:"status"`
	CurrentWorkflowID        string     `json:"current_workflow_id,omitempty"`
	LatestWorkflowID         string     `json:"latest_workflow_id,omitempty"`
	AttemptCount             int        `json:"attempt_count"`
	AutoRetryBudgetRemaining int        `json:"auto_retry_budget_remaining"`
	LastFailureClass         string     `json:"last_failure_class,omitempty"`
	NextRetryAction          string     `json:"next_retry_action,omitempty"`
	RetryAfter               *time.Time `json:"retry_after,omitempty"`
	LineStopReason           string     `json:"line_stop_reason,omitempty"`
	Version                  int64      `json:"version"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	CompletedAt              *time.Time `json:"completed_at,omitempty"`
}

type Assignment struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id,omitempty"`
	CaseID         string    `json:"case_id,omitempty"`
	ThreadKey      string    `json:"thread_key"`
	AssignedBot    string    `json:"assigned_bot"`
	Confidence     float64   `json:"confidence"`
	Rationale      string    `json:"rationale"`
	CreatedAt      time.Time `json:"created_at"`
}

type ToolResult struct {
	Name            string                 `json:"name"`
	ToolCallID      string                 `json:"tool_call_id"`
	Approved        bool                   `json:"approved"`
	ApprovalState   string                 `json:"approval_state,omitempty"`
	Status          string                 `json:"status,omitempty"`
	Available       bool                   `json:"available"`
	Provider        string                 `json:"provider,omitempty"`
	ProviderRef     string                 `json:"provider_ref,omitempty"`
	ExecutedAt      time.Time              `json:"executed_at"`
	Input           map[string]interface{} `json:"input"`
	Output          map[string]interface{} `json:"output"`
	Summary         string                 `json:"summary,omitempty"`
	RawArtifactRefs []string               `json:"raw_artifact_refs,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type ProposalSlotState struct {
	Cap               int      `json:"cap"`
	Active            int      `json:"active"`
	Available         int      `json:"available"`
	ActiveProposalIDs []string `json:"active_proposal_ids"`
	StaleProposalIDs  []string `json:"stale_proposal_ids"`
}

type PromotionResult struct {
	Promoted         int      `json:"promoted"`
	BlockedByCap     bool     `json:"blocked_by_cap"`
	PromotedIDs      []string `json:"promoted_ids"`
	StaleProposalIDs []string `json:"stale_proposal_ids"`
}

type AppDataResetResult struct {
	Backend         string    `json:"backend"`
	ResetAt         time.Time `json:"reset_at"`
	TruncatedTables []string  `json:"truncated_tables"`
	PreservedTables []string  `json:"preserved_tables"`
}

type DerivedTraceRequest struct {
	SourceTraceID  string    `json:"source_trace_id"`
	ProposalID     string    `json:"proposal_id,omitempty"`
	AttemptID      string    `json:"attempt_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	CaseID         string    `json:"case_id,omitempty"`
	ThreadKey      string    `json:"thread_key,omitempty"`
	WorkflowKind   string    `json:"workflow_kind"`
	RequestedBy    string    `json:"requested_by,omitempty"`
	Description    string    `json:"description,omitempty"`
	TriggerEventID string    `json:"trigger_event_id,omitempty"`
	IngestionID    string    `json:"ingestion_id,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}

type TraceUpdate struct {
	Status         *events.Status             `json:"status,omitempty"`
	LastVerdict    *string                    `json:"last_verdict,omitempty"`
	WorkflowStatus string                     `json:"workflow_status,omitempty"`
	WorkflowError  string                     `json:"workflow_error,omitempty"`
	Events         []events.TraceEvent        `json:"events,omitempty"`
	Artifacts      []events.Artifact          `json:"artifacts,omitempty"`
	Reasoning      []events.ReasoningStep     `json:"reasoning,omitempty"`
	ToolCalls      []events.ToolCallRecord    `json:"tool_calls,omitempty"`
	SlackActions   []events.SlackActionRecord `json:"slack_actions,omitempty"`
}
