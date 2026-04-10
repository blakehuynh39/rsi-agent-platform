package events

import "time"

type Status string

const (
	StatusQueued      Status = "queued"
	StatusRunning     Status = "running"
	StatusCompleted   Status = "completed"
	StatusFailed      Status = "failed"
	StatusNeedsHuman  Status = "needs-human"
	StatusSuppressed  Status = "suppressed"
	StatusReplayed    Status = "replayed"
	StatusInReview    Status = "in-review"
	StatusDraft       Status = "draft"
)

type TraceEvent struct {
	TraceID      string     `json:"trace_id"`
	IngestionID  string     `json:"ingestion_id"`
	WorkflowID   string     `json:"workflow_id"`
	ParentEvent  string     `json:"parent_event_id,omitempty"`
	Plane        string     `json:"plane"`
	Service      string     `json:"service"`
	Actor        string     `json:"actor"`
	EventType    string     `json:"event_type"`
	Status       Status     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	PayloadRef   string     `json:"payload_ref,omitempty"`
	ArtifactRef  string     `json:"artifact_ref,omitempty"`
	CostTokens   int        `json:"cost_tokens,omitempty"`
	LatencyMs    int64      `json:"latency_ms,omitempty"`
	Description  string     `json:"description,omitempty"`
}

type Artifact struct {
	ID          string `json:"id"`
	TraceID     string `json:"trace_id"`
	Kind        string `json:"kind"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	SizeBytes   int64  `json:"size_bytes"`
	Source      string `json:"source"`
}

type TraceSummary struct {
	TraceID       string    `json:"trace_id"`
	IngestionID   string    `json:"ingestion_id"`
	WorkflowID    string    `json:"workflow_id"`
	ThreadKey     string    `json:"thread_key"`
	WorkflowKind  string    `json:"workflow_kind"`
	Status        Status    `json:"status"`
	LastVerdict   string    `json:"last_verdict,omitempty"`
	StartedAt     time.Time `json:"started_at"`
	EndedAt       time.Time `json:"ended_at"`
	EventCount    int       `json:"event_count"`
	ArtifactCount int       `json:"artifact_count"`
}

type Trace struct {
	Summary   TraceSummary `json:"summary"`
	Events    []TraceEvent `json:"events"`
	Artifacts []Artifact   `json:"artifacts"`
}

