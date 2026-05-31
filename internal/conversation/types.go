package conversation

import (
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
)

type Status string

const (
	StatusActive     Status = "active"
	StatusMuted      Status = "muted"
	StatusClosed     Status = "closed"
	StatusSuperseded Status = "superseded"
)

type CaseStatus string

const (
	CaseActive     CaseStatus = "active"
	CaseResolved   CaseStatus = "resolved"
	CaseSuperseded CaseStatus = "superseded"
	CaseClosed     CaseStatus = "closed"
)

type ResolutionState string

const (
	ResolutionUnresolved ResolutionState = "unresolved"
	ResolutionMonitoring ResolutionState = "monitoring"
	ResolutionResolved   ResolutionState = "resolved"
	ResolutionRegressed  ResolutionState = "regressed"
)

type Conversation struct {
	ID                   string           `json:"id"`
	Source               ingestion.Source `json:"source"`
	ExternalKey          string           `json:"external_key"`
	ExternalConversation string           `json:"external_conversation"`
	Title                string           `json:"title"`
	Status               Status           `json:"status"`
	ParticipantIDs       []string         `json:"participant_ids,omitempty"`
	ActiveCaseID         string           `json:"active_case_id,omitempty"`
	LatestEventID        string           `json:"latest_event_id,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

type Entry struct {
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversation_id"`
	EventID        string                 `json:"event_id"`
	TraceID        string                 `json:"trace_id,omitempty"`
	Source         ingestion.Source       `json:"source"`
	SourceEventID  string                 `json:"source_event_id"`
	EntryType      string                 `json:"entry_type"`
	ActorID        string                 `json:"actor_id,omitempty"`
	ActorType      string                 `json:"actor_type,omitempty"`
	Body           string                 `json:"body"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

type Case struct {
	ID                 string          `json:"id"`
	ConversationID     string          `json:"conversation_id"`
	Kind               string          `json:"kind"`
	Intent             string          `json:"intent"`
	Title              string          `json:"title"`
	Summary            string          `json:"summary"`
	Status             CaseStatus      `json:"status"`
	ApprovalMode       string          `json:"approval_mode,omitempty"`
	ResponseMode       string          `json:"response_mode,omitempty"`
	AssignedBot        string          `json:"assigned_bot"`
	OpenedByEventID    string          `json:"opened_by_event_id,omitempty"`
	ClosedByEventID    string          `json:"closed_by_event_id,omitempty"`
	LatestTraceID      string          `json:"latest_trace_id,omitempty"`
	ResolutionState    ResolutionState `json:"resolution_state,omitempty"`
	ResolvedAt         *time.Time      `json:"resolved_at,omitempty"`
	LatestOutcomeID    string          `json:"latest_outcome_id,omitempty"`
	OutcomeScore       float64         `json:"outcome_score,omitempty"`
	SupersededByCaseID string          `json:"superseded_by_case_id,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	ClosedAt           *time.Time      `json:"closed_at,omitempty"`
}
