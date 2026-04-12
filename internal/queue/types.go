package queue

import "time"

type QueueName string

const (
	WorkflowQueue  QueueName = "workflow"
	ProactiveQueue QueueName = "proactive"
	EvalQueue      QueueName = "eval"
	ProposalQueue  QueueName = "proposal"
	SandboxQueue   QueueName = "sandbox"
)

type WorkItemStatus string

const (
	WorkQueued    WorkItemStatus = "queued"
	WorkLeased    WorkItemStatus = "leased"
	WorkCompleted WorkItemStatus = "completed"
	WorkFailed    WorkItemStatus = "failed"
	WorkCanceled  WorkItemStatus = "canceled"
)

type WorkItem struct {
	ID             string                 `json:"id"`
	Queue          QueueName              `json:"queue"`
	Kind           string                 `json:"kind"`
	Status         WorkItemStatus         `json:"status"`
	TraceID        string                 `json:"trace_id,omitempty"`
	WorkflowID     string                 `json:"workflow_id,omitempty"`
	IngestionID    string                 `json:"ingestion_id,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	CaseID         string                 `json:"case_id,omitempty"`
	TriggerEventID string                 `json:"trigger_event_id,omitempty"`
	ProposalID     string                 `json:"proposal_id,omitempty"`
	ThreadKey      string                 `json:"thread_key,omitempty"`
	Intent         string                 `json:"intent,omitempty"`
	RepoScope      string                 `json:"repo_scope,omitempty"`
	RequestedBy    string                 `json:"requested_by,omitempty"`
	ApprovalMode   string                 `json:"approval_mode,omitempty"`
	ResponseMode   string                 `json:"response_mode,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
	Attempts       int                    `json:"attempts"`
	LeaseOwner     string                 `json:"lease_owner,omitempty"`
	LeaseExpiresAt *time.Time             `json:"lease_expires_at,omitempty"`
	LastError      string                 `json:"last_error,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
}
