package queue

import "time"

type QueueName string

const (
	WorkflowQueue QueueName = "workflow"
	ProactiveQueue QueueName = "proactive"
	EvalQueue      QueueName = "eval"
	ProposalQueue  QueueName = "proposal"
	SandboxQueue   QueueName = "sandbox"
)

type WorkItem struct {
	ID           string    `json:"id"`
	Queue        QueueName `json:"queue"`
	Kind         string    `json:"kind"`
	TraceID      string    `json:"trace_id,omitempty"`
	ThreadKey    string    `json:"thread_key,omitempty"`
	RepoScope    string    `json:"repo_scope,omitempty"`
	RequestedBy  string    `json:"requested_by,omitempty"`
	ApprovalMode string    `json:"approval_mode,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

