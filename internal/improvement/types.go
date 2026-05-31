package improvement

import (
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/harness"
)

type CandidateStatus string

const (
	CandidateQueued        CandidateStatus = "queued_for_promotion"
	CandidateNeedsEvidence CandidateStatus = "needs_evidence"
	CandidatePromoted      CandidateStatus = "promoted"
	CandidateDormant       CandidateStatus = "dormant"
)

type LineStatus string

const (
	LineQueuedForPromotion LineStatus = "queued_for_promotion"
	LineActive             LineStatus = "active_line"
	LineNeedsRework        LineStatus = "needs_rework"
	LineNeedsEvidence      LineStatus = "needs_evidence"
	LineDormant            LineStatus = "dormant"
	LineClosed             LineStatus = "closed"
)

type RiskTier string

const (
	RiskLow    RiskTier = "low"
	RiskMedium RiskTier = "medium"
	RiskHigh   RiskTier = "high"
)

type Candidate struct {
	ID                            string              `json:"id"`
	CandidateKey                  string              `json:"candidate_key"`
	ConversationID                string              `json:"conversation_id,omitempty"`
	CaseID                        string              `json:"case_id,omitempty"`
	OriginTraceID                 string              `json:"origin_trace_id,omitempty"`
	EvidenceTraceIDs              []string            `json:"evidence_trace_ids,omitempty"`
	Subsystem                     string              `json:"subsystem"`
	FailureMode                   string              `json:"failure_mode"`
	InterventionType              string              `json:"intervention_type"`
	TargetLayer                   harness.TargetLayer `json:"target_layer"`
	TargetKind                    string              `json:"target_kind,omitempty"`
	TargetRef                     string              `json:"target_ref,omitempty"`
	Status                        CandidateStatus     `json:"status"`
	Severity                      string              `json:"severity"`
	RecurrenceCount               int                 `json:"recurrence_count"`
	ExpectedImpact                float64             `json:"expected_impact"`
	NoveltyScore                  float64             `json:"novelty_score"`
	ConfidenceScore               float64             `json:"confidence_score"`
	FreshnessScore                float64             `json:"freshness_score"`
	PriorityScore                 float64             `json:"priority_score"`
	RiskTier                      RiskTier            `json:"risk_tier"`
	Hypothesis                    string              `json:"hypothesis"`
	ProposedScope                 string              `json:"proposed_scope"`
	LatestTraceID                 string              `json:"latest_trace_id,omitempty"`
	SourceEvalIDs                 []string            `json:"source_eval_ids"`
	EvidenceArtifactIDs           []string            `json:"evidence_artifact_ids"`
	PriorSimilarProposalIDs       []string            `json:"prior_similar_proposal_ids"`
	NewEvidenceSinceLastRejection bool                `json:"new_evidence_since_last_rejection"`
	LineStatus                    LineStatus          `json:"line_status,omitempty"`
	RetryableFailureClass         string              `json:"retryable_failure_class,omitempty"`
	LastAttemptID                 string              `json:"last_attempt_id,omitempty"`
	AttemptCount                  int                 `json:"attempt_count,omitempty"`
	AutoRetryBudgetRemaining      int                 `json:"auto_retry_budget_remaining,omitempty"`
	CurrentTargetLayer            harness.TargetLayer `json:"current_target_layer,omitempty"`
	LastEvaluatedAt               time.Time           `json:"last_evaluated_at"`
	CreatedAt                     time.Time           `json:"created_at"`
	UpdatedAt                     time.Time           `json:"updated_at"`
}

type RuntimeDiagnosisStatus string

const (
	RuntimeDiagnosisQueued        RuntimeDiagnosisStatus = "queued"
	RuntimeDiagnosisInvestigating RuntimeDiagnosisStatus = "investigating"
	RuntimeDiagnosisGrounded      RuntimeDiagnosisStatus = "grounded"
	RuntimeDiagnosisNeedsEvidence RuntimeDiagnosisStatus = "needs_evidence"
	RuntimeDiagnosisPromoted      RuntimeDiagnosisStatus = "promoted"
	RuntimeDiagnosisClosed        RuntimeDiagnosisStatus = "closed"
)

type RuntimeDiagnosis struct {
	ID               string                 `json:"id"`
	CandidateKey     string                 `json:"candidate_key"`
	Repo             string                 `json:"repo"`
	ConversationID   string                 `json:"conversation_id,omitempty"`
	CaseID           string                 `json:"case_id,omitempty"`
	LatestTraceID    string                 `json:"latest_trace_id,omitempty"`
	Status           RuntimeDiagnosisStatus `json:"status"`
	Subsystem        string                 `json:"subsystem,omitempty"`
	FailureMode      string                 `json:"failure_mode,omitempty"`
	Summary          string                 `json:"summary,omitempty"`
	EvidenceRefs     []string               `json:"evidence_refs,omitempty"`
	MissingEvidence  []string               `json:"missing_evidence,omitempty"`
	RecommendedFix   string                 `json:"recommended_fix,omitempty"`
	TargetSurface    string                 `json:"target_surface,omitempty"`
	ValidationPlan   string                 `json:"validation_plan,omitempty"`
	SessionScopeKind string                 `json:"session_scope_kind,omitempty"`
	SessionScopeID   string                 `json:"session_scope_id,omitempty"`
	LastResult       map[string]any         `json:"last_result,omitempty"`
	LastError        string                 `json:"last_error,omitempty"`
	LastAttemptedAt  *time.Time             `json:"last_attempted_at,omitempty"`
	PromotedAt       *time.Time             `json:"promoted_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type ChangeAttemptTrigger string

const (
	AttemptTriggerProposalApproved ChangeAttemptTrigger = "proposal_approved"
	AttemptTriggerSandboxFailed    ChangeAttemptTrigger = "sandbox_failed"
	AttemptTriggerPRClosed         ChangeAttemptTrigger = "pr_closed"
	AttemptTriggerCIFailed         ChangeAttemptTrigger = "ci_failed"
	AttemptTriggerPostMergeRegress ChangeAttemptTrigger = "post_merge_regression"
	AttemptTriggerOperatorRetry    ChangeAttemptTrigger = "operator_retry"
)

type ChangeAttemptState string

const (
	AttemptStatePlanned                ChangeAttemptState = "planned"
	AttemptStateWorkspaceRequired      ChangeAttemptState = "workspace_required"
	AttemptStateImplementationRequired ChangeAttemptState = "implementation_required"
	AttemptStateValidationRequired     ChangeAttemptState = "validation_required"
	AttemptStatePRRequired             ChangeAttemptState = "pr_required"
	AttemptStateObservingCI            ChangeAttemptState = "observing_ci"
	AttemptStatePatchPlan           ChangeAttemptState = "patch_plan"
	AttemptStateInvestigateComplete ChangeAttemptState = "investigate_complete"
	AttemptStatePatchGenerated      ChangeAttemptState = "patch_generated"
	AttemptStateValidationRunning   ChangeAttemptState = "validation_running"
	AttemptStateCIObserving         ChangeAttemptState = "ci_observing"
	AttemptStateRetryDeciding       ChangeAttemptState = "retry_deciding"
	AttemptStateOverlayPlan         ChangeAttemptState = "overlay_plan"
	AttemptStateOverlayGenerated    ChangeAttemptState = "overlay_generated"
	AttemptStateOverlayValidating   ChangeAttemptState = "overlay_validating"
	AttemptStateOverlayActive       ChangeAttemptState = "overlay_active"
	AttemptStateSandboxFailed       ChangeAttemptState = "sandbox_failed"
	AttemptStatePROpen              ChangeAttemptState = "pr_open"
	AttemptStateCIFailed            ChangeAttemptState = "ci_failed"
	AttemptStateClosedUnmerged      ChangeAttemptState = "closed_unmerged"
	AttemptStateMerged              ChangeAttemptState = "merged"
	AttemptStateNeedsReview         ChangeAttemptState = "needs_review"
	AttemptStateAbandoned           ChangeAttemptState = "abandoned"
	AttemptStateSuperseded          ChangeAttemptState = "superseded"
)

type ChangeAttempt struct {
	ID                       string               `json:"id"`
	Version                  int64                `json:"version,omitempty"`
	ProposalID               string               `json:"proposal_id"`
	CandidateKey             string               `json:"candidate_key"`
	AttemptNumber            int                  `json:"attempt_number"`
	TargetLayer              harness.TargetLayer  `json:"target_layer"`
	TargetKind               string               `json:"target_kind,omitempty"`
	TargetRef                string               `json:"target_ref,omitempty"`
	Trigger                  ChangeAttemptTrigger `json:"trigger"`
	State                    ChangeAttemptState   `json:"state"`
	AttemptTraceID           string               `json:"attempt_trace_id,omitempty"`
	ParentAttemptID          string               `json:"parent_attempt_id,omitempty"`
	BranchName               string               `json:"branch_name,omitempty"`
	PRURL                    string               `json:"pr_url,omitempty"`
	HeadSHA                  string               `json:"head_sha,omitempty"`
	FailureClass             string               `json:"failure_class,omitempty"`
	FailureSummary           string               `json:"failure_summary,omitempty"`
	RetryDecision            string               `json:"retry_decision,omitempty"`
	RetryAfter               *time.Time           `json:"retry_after,omitempty"`
	MaterialHypothesisChange bool                 `json:"material_hypothesis_change,omitempty"`
	DiffSummary              string               `json:"diff_summary,omitempty"`
	ChangedFiles             []string             `json:"changed_files,omitempty"`
	ValidationSummary        string               `json:"validation_summary,omitempty"`
	ChangePlan               string               `json:"change_plan,omitempty"`
	RepoPatch                string               `json:"repo_patch,omitempty"`
	ValidationPlan           string               `json:"validation_plan,omitempty"`
	RetryAssessment          string               `json:"retry_assessment,omitempty"`
	HypothesisDelta          string               `json:"hypothesis_delta,omitempty"`
	OverlayPayload           map[string]any       `json:"overlay_payload,omitempty"`
	CreatedAt                time.Time            `json:"created_at"`
	UpdatedAt                time.Time            `json:"updated_at"`
}

type AttemptWorkspaceStatus string

const (
	WorkspaceQueued     AttemptWorkspaceStatus = "queued"
	WorkspaceReady      AttemptWorkspaceStatus = "ready"
	WorkspaceExecuting  AttemptWorkspaceStatus = "executing"
	WorkspaceValidating AttemptWorkspaceStatus = "validating"
	WorkspaceCompleted  AttemptWorkspaceStatus = "completed"
	WorkspaceFailed     AttemptWorkspaceStatus = "failed"
	WorkspaceClosed     AttemptWorkspaceStatus = "closed"
)

type AttemptWorkspace struct {
	ID               string                 `json:"id"`
	AttemptID        string                 `json:"attempt_id"`
	ProposalID       string                 `json:"proposal_id"`
	OperationID      string                 `json:"operation_id,omitempty"`
	Generation       int                    `json:"generation,omitempty"`
	Repo             string                 `json:"repo"`
	BaseRef          string                 `json:"base_ref"`
	BranchName       string                 `json:"branch_name"`
	Namespace        string                 `json:"namespace,omitempty"`
	JobName          string                 `json:"job_name,omitempty"`
	PodName          string                 `json:"pod_name,omitempty"`
	Status           AttemptWorkspaceStatus `json:"status"`
	LastError        string                 `json:"last_error,omitempty"`
	Repairable       bool                   `json:"repairable,omitempty"`
	AllowedPathGlobs []string               `json:"allowed_path_globs,omitempty"`
	HeadSHA          string                 `json:"head_sha,omitempty"`
	DiffSummary      string                 `json:"diff_summary,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
}

type ValidationRunStatus string

const (
	ValidationRunRequested ValidationRunStatus = "requested"
	ValidationRunRunning   ValidationRunStatus = "running"
	ValidationRunPassed    ValidationRunStatus = "passed"
	ValidationRunFailed    ValidationRunStatus = "failed"
)

type ValidationRun struct {
	ID               string              `json:"id"`
	ProposalID       string              `json:"proposal_id"`
	AttemptID        string              `json:"attempt_id,omitempty"`
	ConversationID   string              `json:"conversation_id,omitempty"`
	CaseID           string              `json:"case_id,omitempty"`
	OriginTraceID    string              `json:"origin_trace_id,omitempty"`
	WorkspaceID      string              `json:"workspace_id,omitempty"`
	OperationID      string              `json:"operation_id,omitempty"`
	Generation       int                 `json:"generation,omitempty"`
	Repo             string              `json:"repo"`
	BranchName       string              `json:"branch_name"`
	Command          string              `json:"command,omitempty"`
	Status           ValidationRunStatus `json:"status"`
	SandboxNamespace string              `json:"sandbox_namespace,omitempty"`
	SandboxJobName   string              `json:"sandbox_job_name,omitempty"`
	SandboxPodName   string              `json:"sandbox_pod_name,omitempty"`
	ValidationRef    string              `json:"validation_ref,omitempty"`
	ErrorMessage     string              `json:"error_message,omitempty"`
	LogArtifactID    string              `json:"log_artifact_id,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type RepoChangeJob struct {
	ID               string    `json:"id"`
	ProposalID       string    `json:"proposal_id"`
	AttemptID        string    `json:"attempt_id,omitempty"`
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
	AttemptID        string    `json:"attempt_id,omitempty"`
	ConversationID   string    `json:"conversation_id,omitempty"`
	CaseID           string    `json:"case_id,omitempty"`
	OriginTraceID    string    `json:"origin_trace_id,omitempty"`
	OperationID      string    `json:"operation_id,omitempty"`
	Generation       int       `json:"generation,omitempty"`
	Repo             string    `json:"repo"`
	BranchName       string    `json:"branch_name"`
	PRURL            string    `json:"pr_url,omitempty"`
	HeadSHA          string    `json:"head_sha,omitempty"`
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

func PublicAttemptState(state ChangeAttemptState) ChangeAttemptState {
	switch state {
	case AttemptStatePatchPlan, AttemptStateInvestigateComplete:
		return AttemptStateWorkspaceRequired
	case AttemptStatePatchGenerated:
		return AttemptStateValidationRequired
	case AttemptStateValidationRunning:
		return AttemptStatePRRequired
	case AttemptStateCIObserving, AttemptStatePROpen:
		return AttemptStateObservingCI
	case AttemptStateSandboxFailed:
		return AttemptStateNeedsReview
	default:
		return state
	}
}
