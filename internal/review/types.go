package review

import (
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/harness"
)

type ProposalStatus string

const (
	ProposalQueuedForPromotion ProposalStatus = "queued_for_promotion"
	ProposalPendingReview      ProposalStatus = "pending_review"
	ProposalApproved           ProposalStatus = "approved"
	ProposalRepoChangeQueued   ProposalStatus = "repo_change_queued"
	ProposalRepoChangeRunning  ProposalStatus = "repo_change_running"
	ProposalValidationPending  ProposalStatus = "validation_pending"
	ProposalPROpen             ProposalStatus = "pr_open"
	ProposalDismissed          ProposalStatus = "dismissed"
	ProposalRejected           ProposalStatus = "rejected"
	ProposalSuperseded         ProposalStatus = "superseded"
	ProposalMerged             ProposalStatus = "merged"
	ProposalFailedValidation   ProposalStatus = "failed_validation"
	ProposalCanceled           ProposalStatus = "canceled"
)

func ConsumesActiveProposalSlot(status ProposalStatus) bool {
	switch status {
	case ProposalPendingReview, ProposalApproved, ProposalRepoChangeQueued, ProposalRepoChangeRunning, ProposalValidationPending, ProposalPROpen:
		return true
	default:
		return false
	}
}

type HumanRating struct {
	TraceID    string    `json:"trace_id"`
	Score      int       `json:"score"`
	Verdict    string    `json:"verdict"`
	Labels     []string  `json:"labels"`
	Notes      string    `json:"notes"`
	ReviewerID string    `json:"reviewer_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type ImprovementNote struct {
	TraceID        string    `json:"trace_id"`
	Category       string    `json:"category"`
	Note           string    `json:"note"`
	SuggestedOwner string    `json:"suggested_owner"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
}

type FeedbackTargetType string

const (
	FeedbackTargetConversation FeedbackTargetType = "conversation"
	FeedbackTargetCase         FeedbackTargetType = "case"
	FeedbackTargetTrace        FeedbackTargetType = "trace"
	FeedbackTargetReasoning    FeedbackTargetType = "reasoning_step"
	FeedbackTargetToolCall     FeedbackTargetType = "tool_call"
	FeedbackTargetSlackAction  FeedbackTargetType = "slack_action"
	FeedbackTargetProposal     FeedbackTargetType = "proposal"
)

type FeedbackRecord struct {
	ID             string             `json:"id"`
	ConversationID string             `json:"conversation_id,omitempty"`
	CaseID         string             `json:"case_id,omitempty"`
	TraceID        string             `json:"trace_id,omitempty"`
	TargetType     FeedbackTargetType `json:"target_type"`
	TargetID       string             `json:"target_id"`
	Score          int                `json:"score,omitempty"`
	Verdict        string             `json:"verdict,omitempty"`
	Labels         []string           `json:"labels,omitempty"`
	Notes          string             `json:"notes,omitempty"`
	ReviewerID     string             `json:"reviewer_id"`
	CreatedAt      time.Time          `json:"created_at"`
}

type Proposal struct {
	ID                            string              `json:"id"`
	TraceID                       string              `json:"trace_id"`
	ConversationID                string              `json:"conversation_id,omitempty"`
	CaseID                        string              `json:"case_id,omitempty"`
	OriginTraceID                 string              `json:"origin_trace_id,omitempty"`
	EvidenceTraceIDs              []string            `json:"evidence_trace_ids,omitempty"`
	Title                         string              `json:"title"`
	Category                      string              `json:"category"`
	Summary                       string              `json:"summary"`
	Status                        ProposalStatus      `json:"status"`
	Reviewer                      string              `json:"reviewer,omitempty"`
	CandidateKey                  string              `json:"candidate_key"`
	TargetLayer                   harness.TargetLayer `json:"target_layer"`
	TargetKind                    string              `json:"target_kind,omitempty"`
	TargetRef                     string              `json:"target_ref,omitempty"`
	SourceEvalIDs                 []string            `json:"source_eval_ids,omitempty"`
	RiskTier                      string              `json:"risk_tier,omitempty"`
	ProposedScope                 string              `json:"proposed_scope,omitempty"`
	EvidenceArtifactIDs           []string            `json:"evidence_artifact_ids,omitempty"`
	ActiveSlotConsuming           bool                `json:"active_slot_consuming"`
	ReviewDeadline                time.Time           `json:"review_deadline,omitempty"`
	PriorSimilarProposalIDs       []string            `json:"prior_similar_proposal_ids,omitempty"`
	NewEvidenceSinceLastRejection bool                `json:"new_evidence_since_last_rejection"`
	CreatedAt                     time.Time           `json:"created_at"`
	Reviews                       []ProposalReview    `json:"reviews,omitempty"`
}

type ProposalReview struct {
	ID             int64     `json:"id"`
	ProposalID     string    `json:"proposal_id"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	Decision       string    `json:"decision"`
	Rationale      string    `json:"rationale"`
	ReviewerID     string    `json:"reviewer_id"`
	FailureClass   string    `json:"failure_class,omitempty"`
	FailureClasses []string  `json:"failure_classes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type ProposalMemory struct {
	ID                string         `json:"id"`
	ReviewID          int64          `json:"review_id,omitempty"`
	ProposalID        string         `json:"proposal_id"`
	CandidateKey      string         `json:"candidate_key"`
	ConversationID    string         `json:"conversation_id,omitempty"`
	CaseID            string         `json:"case_id,omitempty"`
	OriginTraceID     string         `json:"origin_trace_id,omitempty"`
	EvidenceTraceIDs  []string       `json:"evidence_trace_ids,omitempty"`
	Hypothesis        string         `json:"hypothesis"`
	DiffSummary       string         `json:"diff_summary"`
	ReviewRationale   string         `json:"review_rationale"`
	Disposition       ProposalStatus `json:"disposition"`
	DispositionReason string         `json:"disposition_reason,omitempty"`
	FailureClass      string         `json:"failure_class,omitempty"`
	FailureClasses    []string       `json:"failure_classes,omitempty"`
	SourceEvalIDs     []string       `json:"source_eval_ids,omitempty"`
	LinkedArtifactIDs []string       `json:"linked_artifact_ids,omitempty"`
	LinkedProposalIDs []string       `json:"linked_proposal_ids,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
}
