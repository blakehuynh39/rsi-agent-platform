package outcome

import "time"

type Type string

const (
	TypeAnswerQuality         Type = "answer_quality"
	TypeIncidentMitigation    Type = "incident_mitigation"
	TypeFeatureDelivery       Type = "feature_delivery"
	TypeProposalEffectiveness Type = "proposal_effectiveness"
)

type Verdict string

const (
	VerdictPositive   Verdict = "positive"
	VerdictNegative   Verdict = "negative"
	VerdictMixed      Verdict = "mixed"
	VerdictUnresolved Verdict = "unresolved"
)

type Record struct {
	ID             string    `json:"id"`
	Source         string    `json:"source"`
	SourceEventID  string    `json:"source_event_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	CaseID         string    `json:"case_id,omitempty"`
	TraceID        string    `json:"trace_id,omitempty"`
	ProposalID     string    `json:"proposal_id,omitempty"`
	AttemptID      string    `json:"attempt_id,omitempty"`
	OutcomeType    Type      `json:"outcome_type"`
	Verdict        Verdict   `json:"verdict"`
	Score          float64   `json:"score,omitempty"`
	Summary        string    `json:"summary,omitempty"`
	Details        string    `json:"details,omitempty"`
	ExternalRef    string    `json:"external_ref,omitempty"`
	RecordedBy     string    `json:"recorded_by,omitempty"`
	RecordedAt     time.Time `json:"recorded_at"`
}

type Hypothesis struct {
	OutcomeType         Type   `json:"outcome_type"`
	SuccessCondition    string `json:"success_condition"`
	MeasurementRef      string `json:"measurement_ref,omitempty"`
	ExpectedTimeHorizon string `json:"expected_time_horizon,omitempty"`
}
