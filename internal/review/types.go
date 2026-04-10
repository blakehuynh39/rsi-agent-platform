package review

import "time"

type HumanRating struct {
	TraceID     string    `json:"trace_id"`
	Score       int       `json:"score"`
	Verdict     string    `json:"verdict"`
	Labels      []string  `json:"labels"`
	Notes       string    `json:"notes"`
	ReviewerID  string    `json:"reviewer_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type ImprovementNote struct {
	TraceID        string    `json:"trace_id"`
	Category       string    `json:"category"`
	Note           string    `json:"note"`
	SuggestedOwner string    `json:"suggested_owner"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
}

type Proposal struct {
	ID          string          `json:"id"`
	TraceID     string          `json:"trace_id"`
	Title       string          `json:"title"`
	Category    string          `json:"category"`
	Summary     string          `json:"summary"`
	Status      string          `json:"status"`
	Reviewer    string          `json:"reviewer,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	Reviews     []ProposalReview `json:"reviews,omitempty"`
}

type ProposalReview struct {
	ProposalID string    `json:"proposal_id"`
	Decision   string    `json:"decision"`
	Rationale  string    `json:"rationale"`
	ReviewerID string    `json:"reviewer_id"`
	CreatedAt  time.Time `json:"created_at"`
}

