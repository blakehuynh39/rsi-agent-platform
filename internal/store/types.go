package store

import "time"

type Workflow struct {
	ID          string    `json:"id"`
	ThreadKey   string    `json:"thread_key"`
	Kind        string    `json:"kind"`
	AssignedBot string    `json:"assigned_bot"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Assignment struct {
	ID          string    `json:"id"`
	ThreadKey   string    `json:"thread_key"`
	AssignedBot string    `json:"assigned_bot"`
	Confidence  float64   `json:"confidence"`
	Rationale   string    `json:"rationale"`
	CreatedAt   time.Time `json:"created_at"`
}

type ToolResult struct {
	Name       string                 `json:"name"`
	Approved   bool                   `json:"approved"`
	ExecutedAt time.Time              `json:"executed_at"`
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output"`
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
