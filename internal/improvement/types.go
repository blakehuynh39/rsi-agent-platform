package improvement

import "time"

type CandidateStatus string

const (
	CandidateQueued        CandidateStatus = "queued_for_promotion"
	CandidateNeedsEvidence CandidateStatus = "needs_evidence"
	CandidatePromoted      CandidateStatus = "promoted"
	CandidateDormant       CandidateStatus = "dormant"
)

type RiskTier string

const (
	RiskLow    RiskTier = "low"
	RiskMedium RiskTier = "medium"
	RiskHigh   RiskTier = "high"
)

type Candidate struct {
	ID                            string          `json:"id"`
	CandidateKey                  string          `json:"candidate_key"`
	ConversationID                string          `json:"conversation_id,omitempty"`
	CaseID                        string          `json:"case_id,omitempty"`
	OriginTraceID                 string          `json:"origin_trace_id,omitempty"`
	EvidenceTraceIDs              []string        `json:"evidence_trace_ids,omitempty"`
	Subsystem                     string          `json:"subsystem"`
	FailureMode                   string          `json:"failure_mode"`
	InterventionType              string          `json:"intervention_type"`
	Status                        CandidateStatus `json:"status"`
	Severity                      string          `json:"severity"`
	RecurrenceCount               int             `json:"recurrence_count"`
	ExpectedImpact                float64         `json:"expected_impact"`
	NoveltyScore                  float64         `json:"novelty_score"`
	ConfidenceScore               float64         `json:"confidence_score"`
	FreshnessScore                float64         `json:"freshness_score"`
	PriorityScore                 float64         `json:"priority_score"`
	RiskTier                      RiskTier        `json:"risk_tier"`
	Hypothesis                    string          `json:"hypothesis"`
	ProposedScope                 string          `json:"proposed_scope"`
	LatestTraceID                 string          `json:"latest_trace_id,omitempty"`
	SourceEvalIDs                 []string        `json:"source_eval_ids"`
	EvidenceArtifactIDs           []string        `json:"evidence_artifact_ids"`
	PriorSimilarProposalIDs       []string        `json:"prior_similar_proposal_ids"`
	NewEvidenceSinceLastRejection bool            `json:"new_evidence_since_last_rejection"`
	LastEvaluatedAt               time.Time       `json:"last_evaluated_at"`
	CreatedAt                     time.Time       `json:"created_at"`
	UpdatedAt                     time.Time       `json:"updated_at"`
}

type RepoChangeJob struct {
	ID               string    `json:"id"`
	ProposalID       string    `json:"proposal_id"`
	ConversationID   string    `json:"conversation_id,omitempty"`
	CaseID           string    `json:"case_id,omitempty"`
	OriginTraceID    string    `json:"origin_trace_id,omitempty"`
	CandidateKey     string    `json:"candidate_key"`
	Status           string    `json:"status"`
	Repo             string    `json:"repo"`
	BaseRef          string    `json:"base_ref"`
	BranchName       string    `json:"branch_name"`
	AllowedPathGlobs []string  `json:"allowed_path_globs"`
	ContextSummary   string    `json:"context_summary"`
	SandboxNamespace string    `json:"sandbox_namespace,omitempty"`
	SandboxJobName   string    `json:"sandbox_job_name,omitempty"`
	SandboxPodName   string    `json:"sandbox_pod_name,omitempty"`
	ValidationError  string    `json:"validation_error,omitempty"`
	ValidationRef    string    `json:"validation_ref,omitempty"`
	LogArtifactID    string    `json:"log_artifact_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PRAttempt struct {
	ID               string    `json:"id"`
	ProposalID       string    `json:"proposal_id"`
	ConversationID   string    `json:"conversation_id,omitempty"`
	CaseID           string    `json:"case_id,omitempty"`
	OriginTraceID    string    `json:"origin_trace_id,omitempty"`
	Repo             string    `json:"repo"`
	BranchName       string    `json:"branch_name"`
	PRURL            string    `json:"pr_url,omitempty"`
	Status           string    `json:"status"`
	ValidationStatus string    `json:"validation_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type PostMergeReplay struct {
	ID             string    `json:"id"`
	ProposalID     string    `json:"proposal_id"`
	TraceID        string    `json:"trace_id"`
	ConversationID string    `json:"conversation_id,omitempty"`
	CaseID         string    `json:"case_id,omitempty"`
	BaselineScore  float64   `json:"baseline_score"`
	CandidateScore float64   `json:"candidate_score"`
	Improved       bool      `json:"improved"`
	CreatedAt      time.Time `json:"created_at"`
}

type CronLease struct {
	Name      string    `json:"name"`
	Holder    string    `json:"holder"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Settings struct {
	ActiveProposalCap int       `json:"active_proposal_cap"`
	UpdatedAt         time.Time `json:"updated_at"`
}
