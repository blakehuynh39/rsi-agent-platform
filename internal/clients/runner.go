package clients

import (
	"net/http"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/harness"
)

type RunnerRejectedProposalContext struct {
	ProposalID   string `json:"proposal_id"`
	Disposition  string `json:"disposition"`
	Rationale    string `json:"rationale"`
	FailureClass string `json:"failure_class,omitempty"`
}

type RunnerConversationEntry struct {
	ID            string    `json:"id"`
	EventID       string    `json:"event_id,omitempty"`
	TraceID       string    `json:"trace_id,omitempty"`
	Source        string    `json:"source"`
	SourceEventID string    `json:"source_event_id"`
	EntryType     string    `json:"entry_type"`
	ActorID       string    `json:"actor_id,omitempty"`
	ActorType     string    `json:"actor_type,omitempty"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"created_at"`
}

type RunnerCaseSummary struct {
	CaseID         string `json:"case_id"`
	ConversationID string `json:"conversation_id"`
	Kind           string `json:"kind"`
	Intent         string `json:"intent"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	Status         string `json:"status"`
	AssignedBot    string `json:"assigned_bot"`
	LatestTraceID  string `json:"latest_trace_id,omitempty"`
}

type RunnerTraceRef struct {
	TraceID        string    `json:"trace_id"`
	Status         string    `json:"status"`
	WorkflowKind   string    `json:"workflow_kind"`
	StartedAt      time.Time `json:"started_at"`
	TriggerEventID string    `json:"trigger_event_id,omitempty"`
}

type RunnerContextRef struct {
	Kind                             string         `json:"kind"`
	Ref                              string         `json:"ref,omitempty"`
	Summary                          string         `json:"summary,omitempty"`
	Source                           string         `json:"source,omitempty"`
	ToolCallID                       string         `json:"tool_call_id,omitempty"`
	ToolName                         string         `json:"tool_name,omitempty"`
	Status                           string         `json:"status,omitempty"`
	ChannelID                        string         `json:"channel_id,omitempty"`
	ThreadTS                         string         `json:"thread_ts,omitempty"`
	Since                            string         `json:"since,omitempty"`
	Until                            string         `json:"until,omitempty"`
	StepType                         string         `json:"step_type,omitempty"`
	Decision                         string         `json:"decision,omitempty"`
	Confidence                       float64        `json:"confidence,omitempty"`
	Plane                            string         `json:"plane,omitempty"`
	Service                          string         `json:"service,omitempty"`
	Description                      string         `json:"description,omitempty"`
	TraceID                          string         `json:"trace_id,omitempty"`
	Subsystem                        string         `json:"subsystem,omitempty"`
	FailureMode                      string         `json:"failure_mode,omitempty"`
	TargetLayer                      string         `json:"target_layer,omitempty"`
	PriorityScore                    float64        `json:"priority_score,omitempty"`
	RetryableFailureClass            string         `json:"retryable_failure_class,omitempty"`
	AttemptCount                     int            `json:"attempt_count,omitempty"`
	AutoRetryBudgetRemaining         int            `json:"auto_retry_budget_remaining,omitempty"`
	ProposalID                       string         `json:"proposal_id,omitempty"`
	CandidateKey                     string         `json:"candidate_key,omitempty"`
	Disposition                      string         `json:"disposition,omitempty"`
	Rationale                        string         `json:"rationale,omitempty"`
	Hypothesis                       string         `json:"hypothesis,omitempty"`
	DiffSummary                      string         `json:"diff_summary,omitempty"`
	AttemptNumber                    int            `json:"attempt_number,omitempty"`
	AttemptID                        string         `json:"attempt_id,omitempty"`
	State                            string         `json:"state,omitempty"`
	FailureClass                     string         `json:"failure_class,omitempty"`
	FailureSummary                   string         `json:"failure_summary,omitempty"`
	RetryDecision                    string         `json:"retry_decision,omitempty"`
	NextRetryAction                  string         `json:"next_retry_action,omitempty"`
	RetryAfter                       string         `json:"retry_after,omitempty"`
	LineStopReason                   string         `json:"line_stop_reason,omitempty"`
	RunnerDiagnostics                map[string]any `json:"runner_diagnostics,omitempty"`
	MaterialHypothesisChange         bool           `json:"material_hypothesis_change,omitempty"`
	ChangedFiles                     []string       `json:"changed_files,omitempty"`
	Repo                             string         `json:"repo,omitempty"`
	BranchName                       string         `json:"branch_name,omitempty"`
	TargetKind                       string         `json:"target_kind,omitempty"`
	TargetRef                        string         `json:"target_ref,omitempty"`
	RecommendedInterventionKind      string         `json:"recommended_intervention_kind,omitempty"`
	RecommendedInterventionRationale string         `json:"recommended_intervention_rationale,omitempty"`
	TargetSurface                    string         `json:"target_surface,omitempty"`
	ValidationPlan                   string         `json:"validation_plan,omitempty"`
	MaterialRiskSummary              string         `json:"material_risk_summary,omitempty"`
	RecommendedDisposition           string         `json:"recommended_disposition,omitempty"`
	AllowedPathGlobs                 []string       `json:"allowed_path_globs,omitempty"`
}

type RunnerRequestedArtifact struct {
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
}

type RunnerTask struct {
	TaskType                  string                          `json:"task_type"`
	Repo                      string                          `json:"repo"`
	RepoRef                   string                          `json:"repo_ref,omitempty"`
	Prompt                    string                          `json:"prompt"`
	SystemMessage             string                          `json:"system_message,omitempty"`
	MCPServers                []RunnerMCPServer               `json:"mcp_servers,omitempty"`
	AllowedTools              []string                        `json:"allowed_tools,omitempty"`
	AllowedCommands           []string                        `json:"allowed_commands,omitempty"`
	TimeoutSeconds            int                             `json:"timeout_seconds,omitempty"`
	ExpectedOutputs           []string                        `json:"expected_outputs,omitempty"`
	ArtifactDestination       string                          `json:"artifact_destination,omitempty"`
	RequestedArtifacts        []RunnerRequestedArtifact       `json:"requested_artifacts,omitempty"`
	ArtifactOptional          bool                            `json:"artifact_optional,omitempty"`
	ContextSummary            string                          `json:"context_summary,omitempty"`
	RejectedProposalContext   []RunnerRejectedProposalContext `json:"rejected_proposal_context,omitempty"`
	ExecutionMode             string                          `json:"execution_mode,omitempty"`
	Intent                    string                          `json:"intent,omitempty"`
	TraceID                   string                          `json:"trace_id,omitempty"`
	WorkflowID                string                          `json:"workflow_id,omitempty"`
	ConversationID            string                          `json:"conversation_id,omitempty"`
	CaseID                    string                          `json:"case_id,omitempty"`
	ChannelID                 string                          `json:"channel_id,omitempty"`
	ThreadTS                  string                          `json:"thread_ts,omitempty"`
	TriggerEventID            string                          `json:"trigger_event_id,omitempty"`
	RecentConversationEntries []RunnerConversationEntry       `json:"recent_conversation_entries,omitempty"`
	CaseSummary               *RunnerCaseSummary              `json:"case_summary,omitempty"`
	PriorTraceRefs            []RunnerTraceRef                `json:"prior_trace_refs,omitempty"`
	RepoAllowlist             []string                        `json:"repo_allowlist,omitempty"`
	ToolAllowlist             []string                        `json:"tool_allowlist,omitempty"`
	ResponseMode              string                          `json:"response_mode,omitempty"`
	ReplyDeliveryMode         string                          `json:"reply_delivery_mode,omitempty"`
	ContextRefs               []RunnerContextRef              `json:"context_refs,omitempty"`
	ApprovalMode              string                          `json:"approval_mode,omitempty"`
	ReasoningVerbosity        string                          `json:"reasoning_verbosity,omitempty"`
	SessionScopeKind          string                          `json:"session_scope_kind,omitempty"`
	SessionScopeID            string                          `json:"session_scope_id,omitempty"`
	ParentSessionScopeKind    string                          `json:"parent_session_scope_kind,omitempty"`
	ParentSessionScopeID      string                          `json:"parent_session_scope_id,omitempty"`
	HarnessProfileID          string                          `json:"harness_profile_id,omitempty"`
	HarnessOverlayVersion     string                          `json:"harness_overlay_version,omitempty"`
	MemoryBackend             string                          `json:"memory_backend,omitempty"`
	AssistantPeerID           string                          `json:"assistant_peer_id,omitempty"`
	UserPeerID                string                          `json:"user_peer_id,omitempty"`
	AttemptID                 string                          `json:"attempt_id,omitempty"`
	WorkspaceID               string                          `json:"workspace_id,omitempty"`
	WorkspaceRepo             string                          `json:"workspace_repo,omitempty"`
	WorkspaceBranch           string                          `json:"workspace_branch,omitempty"`
	AllowedPathGlobs          []string                        `json:"allowed_path_globs,omitempty"`
}

type RunnerMCPServer struct {
	ServerLabel         string            `json:"server_label,omitempty"`
	ServerURL           string            `json:"server_url,omitempty"`
	Authorization       string            `json:"authorization,omitempty"`
	AuthorizationEnvVar string            `json:"authorization_env_var,omitempty"`
	AllowedTools        map[string]any    `json:"allowed_tools,omitempty"`
	RequireApproval     any               `json:"require_approval,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	HeaderEnvVars       map[string]string `json:"header_env_vars,omitempty"`
	Profile             string            `json:"profile,omitempty"`
}

type RunnerResponse struct {
	OK       bool                   `json:"ok"`
	Message  string                 `json:"message"`
	Provider string                 `json:"provider"`
	Raw      map[string]interface{} `json:"raw"`
}

type RuntimeResponse = harness.RuntimeResponse

type RunnerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewRunnerClient(baseURL string) *RunnerClient {
	return NewRunnerClientWithTimeout(baseURL, 60*time.Second)
}

func NewRunnerClientWithTimeout(baseURL string, timeout time.Duration) *RunnerClient {
	return &RunnerClient{
		baseURL:    trimBaseURL(baseURL),
		httpClient: newHTTPClient(timeout),
	}
}

func (c *RunnerClient) Execute(task RunnerTask) (RunnerResponse, error) {
	var out RunnerResponse
	if err := doJSON(c.httpClient, http.MethodPost, c.baseURL+"/execute", map[string]RunnerTask{"task": task}, &out, "runner"); err != nil {
		return RunnerResponse{}, err
	}
	return out, nil
}

func (c *RunnerClient) Runtime() (RuntimeResponse, error) {
	var out RuntimeResponse
	if err := doJSON(c.httpClient, http.MethodGet, c.baseURL+"/runtimez", nil, &out, "runner"); err != nil {
		return RuntimeResponse{}, err
	}
	return out, nil
}
