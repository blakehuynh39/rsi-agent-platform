package clients

import (
	"net/http"
	"strings"
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
	Namespaces                       []string       `json:"namespaces,omitempty"`
}

type RunnerRequestedArtifact struct {
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
}

const RunnerExecutionContractVersion = "execution-envelope/v1"

type RunnerCapabilityLease struct {
	LeaseID     string         `json:"lease_id,omitempty"`
	Capability  string         `json:"capability"`
	Scope       map[string]any `json:"scope,omitempty"`
	Constraints map[string]any `json:"constraints,omitempty"`
	Granted     bool           `json:"granted"`
}

type RunnerDeliveryPolicy struct {
	BoundChannelID     string `json:"bound_channel_id,omitempty"`
	BoundThreadTS      string `json:"bound_thread_ts,omitempty"`
	TargetSurface      string `json:"target_surface,omitempty"`
	DirectSendAllowed  bool   `json:"direct_send_allowed"`
	UploadAllowed      bool   `json:"upload_allowed"`
	IdempotencyKeyBase string `json:"idempotency_key_base,omitempty"`
}

type RunnerWorkspacePolicy struct {
	ComputerRoot     string   `json:"computer_root,omitempty"`
	RunRoot          string   `json:"run_root,omitempty"`
	ArtifactRoot     string   `json:"artifact_root,omitempty"`
	AllowedPathRoots []string `json:"allowed_path_roots,omitempty"`
}

type RunnerApprovalPolicy struct {
	DirectSlackAllowed               bool     `json:"direct_slack_allowed"`
	RequiresApproval                 []string `json:"requires_approval,omitempty"`
	PlatformMutationsExecuteDirectly bool     `json:"platform_mutations_execute_directly"`
}

func NewRunnerCapabilityLease(capability string, scope map[string]any) RunnerCapabilityLease {
	capability = strings.TrimSpace(capability)
	return RunnerCapabilityLease{
		LeaseID:    "lease-" + strings.ReplaceAll(capability, "_", "-"),
		Capability: capability,
		Scope:      scope,
		Granted:    true,
	}
}

func RunnerCapabilityLeases(capabilities ...string) []RunnerCapabilityLease {
	out := make([]RunnerCapabilityLease, 0, len(capabilities))
	seen := map[string]struct{}{}
	for _, capability := range capabilities {
		capability = strings.TrimSpace(capability)
		if capability == "" {
			continue
		}
		if _, ok := seen[capability]; ok {
			continue
		}
		seen[capability] = struct{}{}
		out = append(out, NewRunnerCapabilityLease(capability, nil))
	}
	return out
}

func AttachKubernetesReadNamespacesToLeases(leases []RunnerCapabilityLease, kubernetesReadNamespaces []string) []RunnerCapabilityLease {
	if len(kubernetesReadNamespaces) == 0 {
		return leases
	}
	for index := range leases {
		if leases[index].Capability != "read_context" {
			continue
		}
		leases[index].Scope = map[string]any{
			"kubernetes_read_namespaces": append([]string(nil), kubernetesReadNamespaces...),
		}
		break
	}
	return leases
}

func NewRunnerDeliveryPolicy(channelID string, threadTS string, replyDeliveryMode string, idempotencyKeyBase string) *RunnerDeliveryPolicy {
	mode := NormalizeRunnerReplyDeliveryMode(replyDeliveryMode)
	targetSurface := "channel"
	if strings.TrimSpace(threadTS) != "" {
		targetSurface = "thread"
	}
	if strings.HasPrefix(strings.TrimSpace(channelID), "D") && strings.TrimSpace(threadTS) == "" {
		targetSurface = "direct_message"
	}
	return &RunnerDeliveryPolicy{
		BoundChannelID:     strings.TrimSpace(channelID),
		BoundThreadTS:      strings.TrimSpace(threadTS),
		TargetSurface:      targetSurface,
		DirectSendAllowed:  mode == "direct",
		UploadAllowed:      mode == "direct" || mode == "mediated",
		IdempotencyKeyBase: strings.TrimSpace(idempotencyKeyBase),
	}
}

func NewRunnerWorkspacePolicy(computerRoot string, runRoot string, artifactRoot string) *RunnerWorkspacePolicy {
	roots := compactRunnerStrings(computerRoot, runRoot, artifactRoot)
	return &RunnerWorkspacePolicy{
		ComputerRoot:     strings.TrimSpace(computerRoot),
		RunRoot:          strings.TrimSpace(runRoot),
		ArtifactRoot:     strings.TrimSpace(artifactRoot),
		AllowedPathRoots: roots,
	}
}

func NewRunnerApprovalPolicy(directSlackAllowed bool) *RunnerApprovalPolicy {
	return &RunnerApprovalPolicy{
		DirectSlackAllowed: directSlackAllowed,
		RequiresApproval: []string{
			"repo_merge",
			"platform_config",
			"deployment",
			"k8s_mutation",
			"aws_iac",
			"destructive_action",
			"harness_platform_behavior_change",
		},
		PlatformMutationsExecuteDirectly: false,
	}
}

func NormalizeRunnerReplyDeliveryMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "direct", "mediated", "none":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "mediated"
	}
}

func compactRunnerStrings(values ...string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

type RunnerTask struct {
	TaskType                  string                          `json:"task_type"`
	Repo                      string                          `json:"repo"`
	RepoRef                   string                          `json:"repo_ref,omitempty"`
	Prompt                    string                          `json:"prompt"`
	SystemMessage             string                          `json:"system_message,omitempty"`
	RequestedSkills           []string                        `json:"requested_skills,omitempty"`
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
	OperationID               string                          `json:"operation_id,omitempty"`
	ExecutionID               string                          `json:"execution_id,omitempty"`
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
	KubernetesReadNamespaces  []string                        `json:"kubernetes_read_namespaces,omitempty"`
	ContractVersion           string                          `json:"contract_version,omitempty"`
	ExecutionIntent           map[string]any                  `json:"execution_intent,omitempty"`
	CapabilityLeases          []RunnerCapabilityLease         `json:"capability_leases,omitempty"`
	DeliveryPolicy            *RunnerDeliveryPolicy           `json:"delivery_policy,omitempty"`
	WorkspacePolicy           *RunnerWorkspacePolicy          `json:"workspace_policy,omitempty"`
	ApprovalPolicy            *RunnerApprovalPolicy           `json:"approval_policy,omitempty"`
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

type HermesExecutionRequest struct {
	Task  RunnerTask `json:"task"`
	Async bool       `json:"async,omitempty"`
}

type HermesExecutionResult = RunnerResponse

type HermesExecutionStatus struct {
	ExecutionID       string          `json:"execution_id,omitempty"`
	OperationID       string          `json:"operation_id,omitempty"`
	TraceID           string          `json:"trace_id,omitempty"`
	WorkflowID        string          `json:"workflow_id,omitempty"`
	Phase             string          `json:"phase,omitempty"`
	Status            string          `json:"status,omitempty"`
	WorkspaceRoot     string          `json:"workspace_root,omitempty"`
	SessionID         string          `json:"session_id,omitempty"`
	TerminationReason string          `json:"termination_reason,omitempty"`
	CompletionVerdict string          `json:"completion_verdict,omitempty"`
	Message           string          `json:"message,omitempty"`
	Result            *RunnerResponse `json:"result,omitempty"`
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
	if err := doJSON(c.httpClientForTask(task), http.MethodPost, c.baseURL+"/execute", map[string]RunnerTask{"task": task}, &out, "runner"); err != nil {
		return RunnerResponse{}, err
	}
	return out, nil
}

func (c *RunnerClient) ExecuteHermesExecution(task RunnerTask) (HermesExecutionResult, error) {
	var out HermesExecutionResult
	if err := doJSON(c.httpClientForTask(task), http.MethodPost, c.baseURL+"/internal/hermes-executions", HermesExecutionRequest{Task: task}, &out, "hermes executor"); err != nil {
		return HermesExecutionResult{}, err
	}
	return out, nil
}

func (c *RunnerClient) StartHermesExecution(task RunnerTask) (HermesExecutionStatus, error) {
	var out HermesExecutionStatus
	if err := doJSON(c.httpClientForTask(task), http.MethodPost, c.baseURL+"/internal/hermes-executions", HermesExecutionRequest{Task: task, Async: true}, &out, "hermes executor start"); err != nil {
		return HermesExecutionStatus{}, err
	}
	return out, nil
}

func (c *RunnerClient) httpClientForTask(task RunnerTask) *http.Client {
	if task.TimeoutSeconds <= 0 {
		return c.httpClient
	}
	required := time.Duration(task.TimeoutSeconds+30) * time.Second
	if required <= c.httpClient.Timeout {
		return c.httpClient
	}
	return newHTTPClient(required)
}

func (c *RunnerClient) HermesExecutionStatus(executionID string) (HermesExecutionStatus, error) {
	var out HermesExecutionStatus
	if err := doJSON(c.httpClient, http.MethodGet, c.baseURL+"/internal/hermes-executions/"+executionID, nil, &out, "hermes executor status"); err != nil {
		return HermesExecutionStatus{}, err
	}
	return out, nil
}

func (c *RunnerClient) CancelHermesExecution(executionID string) (HermesExecutionStatus, error) {
	var out HermesExecutionStatus
	if err := doJSON(c.httpClient, http.MethodPost, c.baseURL+"/internal/hermes-executions/"+executionID+"/cancel", map[string]any{}, &out, "hermes executor cancel"); err != nil {
		return HermesExecutionStatus{}, err
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
