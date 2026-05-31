package evals

import "time"

type Layer string

const (
	LayerDeterministic Layer = "deterministic"
	LayerTaskQuality   Layer = "task-quality"
	LayerArchitecture  Layer = "architecture"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Suite struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	EventKinds  []string `json:"event_kinds"`
	Layers      []Layer  `json:"layers"`
}

type Run struct {
	ID             string    `json:"id"`
	TraceID        string    `json:"trace_id"`
	EventID        string    `json:"event_id,omitempty"`
	SuiteName      string    `json:"suite_name"`
	Status         Status    `json:"status"`
	Trigger        string    `json:"trigger"`
	OverallScore   float64   `json:"overall_score"`
	OverallVerdict string    `json:"overall_verdict"`
	CreatedAt      time.Time `json:"created_at"`
	CompletedAt    time.Time `json:"completed_at,omitempty"`
}

type Judgment struct {
	ID        string    `json:"id"`
	EvalRunID string    `json:"eval_run_id"`
	Layer     Layer     `json:"layer"`
	Category  string    `json:"category"`
	Score     float64   `json:"score"`
	Passed    bool      `json:"passed"`
	Rationale string    `json:"rationale"`
	CreatedAt time.Time `json:"created_at"`
}
