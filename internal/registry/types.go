package registry

type OwnershipRecord struct {
	Domain          string `json:"domain"`
	OwnerTeam       string `json:"owner_team"`
	EscalationSlack string `json:"escalation_slack"`
}

type CapabilityRecord struct {
	Name           string   `json:"name"`
	Kind           string   `json:"kind"`
	AllowedBots    []string `json:"allowed_bots"`
	ApprovalNeeded bool     `json:"approval_needed"`
}

type WorkflowTemplate struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

type ExperimentRecord struct {
	Name       string `json:"name"`
	Candidate  string `json:"candidate"`
	Baseline   string `json:"baseline"`
	State      string `json:"state"`
	ReviewedBy string `json:"reviewed_by,omitempty"`
}

