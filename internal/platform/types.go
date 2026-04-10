package platform

import "time"

type Workflow struct {
	ID           string    `json:"id"`
	ThreadKey    string    `json:"thread_key"`
	Kind         string    `json:"kind"`
	AssignedBot  string    `json:"assigned_bot"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Assignment struct {
	ID          string    `json:"id"`
	ThreadKey   string    `json:"thread_key"`
	AssignedBot string    `json:"assigned_bot"`
	Confidence  float64   `json:"confidence"`
	Rationale   string    `json:"rationale"`
	CreatedAt   time.Time `json:"created_at"`
}

type ToolResult struct {
	Name        string                 `json:"name"`
	Approved    bool                   `json:"approved"`
	ExecutedAt  time.Time              `json:"executed_at"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
}

