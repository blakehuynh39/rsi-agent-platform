package ingestion

import "time"

type Source string

const (
	SourceSlack  Source = "slack"
	SourceSentry Source = "sentry"
	SourceGitHub Source = "github"
	SourceReplay Source = "replay"
	SourceSystem Source = "system"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

type EventEnvelope struct {
	ID                         string                 `json:"id"`
	Source                     Source                 `json:"source"`
	SourceEventID              string                 `json:"source_event_id"`
	ThreadKey                  string                 `json:"thread_key,omitempty"`
	IncidentKey                string                 `json:"incident_key,omitempty"`
	DedupeKey                  string                 `json:"dedupe_key"`
	Severity                   Severity               `json:"severity"`
	NormalizedProblemStatement string                 `json:"normalized_problem_statement"`
	OwnershipHint              string                 `json:"ownership_hint,omitempty"`
	RawPayloadRef              string                 `json:"raw_payload_ref,omitempty"`
	WorkflowHint               string                 `json:"workflow_hint,omitempty"`
	Metadata                   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt                  time.Time              `json:"created_at"`
}
