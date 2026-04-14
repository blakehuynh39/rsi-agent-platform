package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RunnerTask struct {
	TaskType                  string           `json:"task_type"`
	Repo                      string           `json:"repo"`
	RepoRef                   string           `json:"repo_ref,omitempty"`
	Prompt                    string           `json:"prompt"`
	SystemMessage             string           `json:"system_message,omitempty"`
	AllowedTools              []string         `json:"allowed_tools,omitempty"`
	AllowedCommands           []string         `json:"allowed_commands,omitempty"`
	TimeoutSeconds            int              `json:"timeout_seconds,omitempty"`
	ExpectedOutputs           []string         `json:"expected_outputs,omitempty"`
	ArtifactDestination       string           `json:"artifact_destination,omitempty"`
	ContextSummary            string           `json:"context_summary,omitempty"`
	RejectedProposalContext   []map[string]any `json:"rejected_proposal_context,omitempty"`
	Intent                    string           `json:"intent,omitempty"`
	TraceID                   string           `json:"trace_id,omitempty"`
	WorkflowID                string           `json:"workflow_id,omitempty"`
	ConversationID            string           `json:"conversation_id,omitempty"`
	CaseID                    string           `json:"case_id,omitempty"`
	TriggerEventID            string           `json:"trigger_event_id,omitempty"`
	RecentConversationEntries []map[string]any `json:"recent_conversation_entries,omitempty"`
	CaseSummary               map[string]any   `json:"case_summary,omitempty"`
	PriorTraceRefs            []map[string]any `json:"prior_trace_refs,omitempty"`
	RepoAllowlist             []string         `json:"repo_allowlist,omitempty"`
	ToolAllowlist             []string         `json:"tool_allowlist,omitempty"`
	ResponseMode              string           `json:"response_mode,omitempty"`
	ContextRefs               []map[string]any `json:"context_refs,omitempty"`
	ApprovalMode              string           `json:"approval_mode,omitempty"`
	ReasoningVerbosity        string           `json:"reasoning_verbosity,omitempty"`
	SessionScopeKind          string           `json:"session_scope_kind,omitempty"`
	SessionScopeID            string           `json:"session_scope_id,omitempty"`
	ParentSessionScopeKind    string           `json:"parent_session_scope_kind,omitempty"`
	ParentSessionScopeID      string           `json:"parent_session_scope_id,omitempty"`
	HarnessProfileID          string           `json:"harness_profile_id,omitempty"`
	HarnessOverlayVersion     string           `json:"harness_overlay_version,omitempty"`
	MemoryBackend             string           `json:"memory_backend,omitempty"`
	AssistantPeerID           string           `json:"assistant_peer_id,omitempty"`
	UserPeerID                string           `json:"user_peer_id,omitempty"`
}

type RunnerResponse struct {
	OK       bool                   `json:"ok"`
	Message  string                 `json:"message"`
	Provider string                 `json:"provider"`
	Raw      map[string]interface{} `json:"raw"`
}

type RuntimeResponse struct {
	Status                string `json:"status"`
	Role                  string `json:"role"`
	Backend               string `json:"backend"`
	Provider              string `json:"provider"`
	Model                 string `json:"model"`
	ProviderModel         string `json:"provider_model"`
	APIMode               string `json:"api_mode"`
	ReasoningEffort       string `json:"reasoning_effort"`
	Available             bool   `json:"available"`
	HermesAvailable       bool   `json:"hermes_available"`
	OpenAIConfigured      bool   `json:"openai_configured"`
	PersistenceEnabled    bool   `json:"persistence_enabled"`
	HermesHome            string `json:"hermes_home,omitempty"`
	SessionDBPath         string `json:"session_db_path,omitempty"`
	MemoryBackend         string `json:"memory_backend,omitempty"`
	HonchoConfigured      bool   `json:"honcho_configured"`
	HonchoAvailable       bool   `json:"honcho_available"`
	HonchoBaseURL         string `json:"honcho_base_url,omitempty"`
	HonchoWorkspace       string `json:"honcho_workspace,omitempty"`
	HonchoEnvironment     string `json:"honcho_environment,omitempty"`
	HonchoRecallMode      string `json:"honcho_recall_mode,omitempty"`
	HonchoWriteFrequency  string `json:"honcho_write_frequency,omitempty"`
	HonchoSessionStrategy string `json:"honcho_session_strategy,omitempty"`
	HonchoAIPeer          string `json:"honcho_ai_peer,omitempty"`
}

type RunnerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewRunnerClient(baseURL string) *RunnerClient {
	return NewRunnerClientWithTimeout(baseURL, 60*time.Second)
}

func NewRunnerClientWithTimeout(baseURL string, timeout time.Duration) *RunnerClient {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &RunnerClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *RunnerClient) Execute(task RunnerTask) (RunnerResponse, error) {
	body, err := json.Marshal(map[string]any{"task": task})
	if err != nil {
		return RunnerResponse{}, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/execute", bytes.NewReader(body))
	if err != nil {
		return RunnerResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return RunnerResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return RunnerResponse{}, fmt.Errorf("runner returned %d", resp.StatusCode)
	}
	var out RunnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return RunnerResponse{}, err
	}
	return out, nil
}

func (c *RunnerClient) Runtime() (RuntimeResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/runtimez", nil)
	if err != nil {
		return RuntimeResponse{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return RuntimeResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return RuntimeResponse{}, fmt.Errorf("runner returned %d", resp.StatusCode)
	}
	var out RuntimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return RuntimeResponse{}, err
	}
	return out, nil
}
