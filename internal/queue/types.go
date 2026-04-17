package queue

type QueueName string

const (
	WorkflowQueue             QueueName = "workflow"
	ProactiveQueue            QueueName = "proactive"
	EvalQueue                 QueueName = "eval"
	ProposalQueue             QueueName = "proposal"
	SandboxQueue              QueueName = "sandbox"
	ImprovementActionQueue    QueueName = "improvement_action"
	KnowledgeMaintenanceQueue QueueName = "knowledge_maintenance"
)
