package harness

import "time"

type TargetLayer string

const (
	TargetLayerPlatformRuntime TargetLayer = "platform_runtime"
	TargetLayerHarnessOverlay  TargetLayer = "harness_overlay"
	TargetLayerRepoChange      TargetLayer = "repo_change"
)

type OverlayStatus string

const (
	OverlayStatusDraft      OverlayStatus = "draft"
	OverlayStatusApproved   OverlayStatus = "approved"
	OverlayStatusActive     OverlayStatus = "active"
	OverlayStatusSuperseded OverlayStatus = "superseded"
	OverlayStatusRolledBack OverlayStatus = "rolled_back"
)

type ExperimentStatus string

const (
	ExperimentStatusQueued     ExperimentStatus = "queued"
	ExperimentStatusRunning    ExperimentStatus = "running"
	ExperimentStatusSucceeded  ExperimentStatus = "succeeded"
	ExperimentStatusFailed     ExperimentStatus = "failed"
	ExperimentStatusRolledBack ExperimentStatus = "rolled_back"
)

type MemoryArtifact struct {
	Kind      string    `json:"kind"`
	Summary   string    `json:"summary"`
	Ref       string    `json:"ref,omitempty"`
	Source    string    `json:"source,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Profile struct {
	ID                  string    `json:"id"`
	Role                string    `json:"role"`
	Name                string    `json:"name"`
	Description         string    `json:"description,omitempty"`
	Model               string    `json:"model,omitempty"`
	ReasoningEffort     string    `json:"reasoning_effort,omitempty"`
	PromptFragments     []string  `json:"prompt_fragments,omitempty"`
	FewShotSnippets     []string  `json:"few_shot_snippets,omitempty"`
	ToolPreferenceOrder []string  `json:"tool_preference_order,omitempty"`
	RetrievalBias       string    `json:"retrieval_bias,omitempty"`
	ReasoningVerbosity  string    `json:"reasoning_verbosity,omitempty"`
	MemoryReadEnabled   bool      `json:"memory_read_enabled"`
	MemoryWriteEnabled  bool      `json:"memory_write_enabled"`
	RepoRef             string    `json:"repo_ref,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Overlay struct {
	ID                  string        `json:"id"`
	ProfileID           string        `json:"profile_id"`
	Role                string        `json:"role"`
	Version             string        `json:"version"`
	Status              OverlayStatus `json:"status"`
	TargetKind          string        `json:"target_kind,omitempty"`
	TargetRef           string        `json:"target_ref,omitempty"`
	ProposalID          string        `json:"proposal_id,omitempty"`
	PromptFragments     []string      `json:"prompt_fragments,omitempty"`
	FewShotSnippets     []string      `json:"few_shot_snippets,omitempty"`
	ToolPreferenceOrder []string      `json:"tool_preference_order,omitempty"`
	RetrievalBias       string        `json:"retrieval_bias,omitempty"`
	ReasoningVerbosity  string        `json:"reasoning_verbosity,omitempty"`
	MemoryReadEnabled   *bool         `json:"memory_read_enabled,omitempty"`
	MemoryWriteEnabled  *bool         `json:"memory_write_enabled,omitempty"`
	CreatedBy           string        `json:"created_by,omitempty"`
	ApprovedBy          string        `json:"approved_by,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	ActivatedAt         *time.Time    `json:"activated_at,omitempty"`
}

type Experiment struct {
	ID         string           `json:"id"`
	ProfileID  string           `json:"profile_id"`
	OverlayID  string           `json:"overlay_id,omitempty"`
	ProposalID string           `json:"proposal_id,omitempty"`
	AttemptID  string           `json:"attempt_id,omitempty"`
	Role       string           `json:"role"`
	Status     ExperimentStatus `json:"status"`
	Summary    string           `json:"summary,omitempty"`
	Metrics    map[string]any   `json:"metrics,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type SessionBinding struct {
	Role                    string    `json:"role"`
	ScopeKind               string    `json:"scope_kind"`
	ScopeID                 string    `json:"scope_id"`
	ParentScopeKind         string    `json:"parent_scope_kind,omitempty"`
	ParentScopeID           string    `json:"parent_scope_id,omitempty"`
	HermesSessionID         string    `json:"hermes_session_id"`
	ParentSessionID         string    `json:"parent_session_id,omitempty"`
	MemoryBackend           string    `json:"memory_backend"`
	AssistantPeerID         string    `json:"assistant_peer_id,omitempty"`
	UserPeerID              string    `json:"user_peer_id,omitempty"`
	HarnessProfileID        string    `json:"harness_profile_id,omitempty"`
	EffectiveOverlayID      string    `json:"effective_overlay_id,omitempty"`
	EffectiveOverlayVersion string    `json:"effective_overlay_version,omitempty"`
	LastUsedAt              time.Time `json:"last_used_at"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type Execution struct {
	ID                      string           `json:"id"`
	OperationID             string           `json:"operation_id,omitempty"`
	TraceID                 string           `json:"trace_id,omitempty"`
	ProposalID              string           `json:"proposal_id,omitempty"`
	Role                    string           `json:"role"`
	SessionScopeKind        string           `json:"session_scope_kind"`
	SessionScopeID          string           `json:"session_scope_id"`
	HermesSessionID         string           `json:"hermes_session_id"`
	ParentSessionID         string           `json:"parent_session_id,omitempty"`
	HarnessProfileID        string           `json:"harness_profile_id,omitempty"`
	EffectiveOverlayID      string           `json:"effective_overlay_id,omitempty"`
	EffectiveOverlayVersion string           `json:"effective_overlay_version,omitempty"`
	MemoryBackend           string           `json:"memory_backend,omitempty"`
	MemoryReads             []MemoryArtifact `json:"memory_reads,omitempty"`
	MemoryWrites            []MemoryArtifact `json:"memory_writes,omitempty"`
	CreatedAt               time.Time        `json:"created_at"`
}

type ExecutionObservation struct {
	ID              string         `json:"id"`
	ExecutionID     string         `json:"execution_id"`
	OperationID     string         `json:"operation_id,omitempty"`
	TraceID         string         `json:"trace_id,omitempty"`
	WorkflowID      string         `json:"workflow_id,omitempty"`
	HermesSessionID string         `json:"hermes_session_id,omitempty"`
	Role            string         `json:"role,omitempty"`
	Phase           string         `json:"phase"`
	EventType       string         `json:"event_type"`
	Status          string         `json:"status,omitempty"`
	Seq             int            `json:"seq"`
	Payload         map[string]any `json:"payload,omitempty"`
	RecordedAt      time.Time      `json:"recorded_at"`
}
