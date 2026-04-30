package harness

type RuntimeComponent struct {
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
}

type DialecticLevel struct {
	RuntimeComponent
	ThinkingBudgetTokens int `json:"thinking_budget_tokens"`
}

type HermesContractStatus struct {
	OK                 bool              `json:"ok"`
	ExpectedPin        string            `json:"expected_pin,omitempty"`
	InstalledCommit    string            `json:"installed_commit,omitempty"`
	HermesVersion      string            `json:"hermes_version,omitempty"`
	APISignatureStatus string            `json:"api_signature_status,omitempty"`
	PinStatus          string            `json:"pin_status,omitempty"`
	PluginStatus       string            `json:"plugin_status,omitempty"`
	RequiredToolsets   []string          `json:"required_toolsets,omitempty"`
	ToolsetStatus      map[string]string `json:"toolset_status,omitempty"`
	SessionDBStatus    string            `json:"session_db_status,omitempty"`
	Errors             []string          `json:"errors,omitempty"`
	CheckedAtUnix      float64           `json:"checked_at_unix,omitempty"`
}

type RuntimeResponse struct {
	Status                      string                `json:"status"`
	Role                        string                `json:"role"`
	ExecutorInstanceID          string                `json:"executor_instance_id,omitempty"`
	DrainStatus                 string                `json:"drain_status,omitempty"`
	ActiveExecutionCount        int                   `json:"active_execution_count,omitempty"`
	Backend                     string                `json:"backend"`
	Provider                    string                `json:"provider"`
	Model                       string                `json:"model"`
	ProviderModel               string                `json:"provider_model"`
	APIMode                     string                `json:"api_mode"`
	ReasoningEffort             string                `json:"reasoning_effort"`
	HermesVersion               string                `json:"hermes_version"`
	HermesPin                   string                `json:"hermes_pin"`
	HermesContractStatus        *HermesContractStatus `json:"hermes_contract_status,omitempty"`
	ExecutionContractVersion    string                `json:"execution_contract_version,omitempty"`
	ExecutionEnvelopeV1Enabled  bool                  `json:"execution_envelope_v1_enabled,omitempty"`
	ExecutionLedgerFirstEnabled bool                  `json:"execution_ledger_first_projection_enabled,omitempty"`
	CompanyComputerRoot         string                `json:"company_computer_root,omitempty"`
	RunnerPlannerMode           string                `json:"runner_planner_mode,omitempty"`
	RequiredCapabilities        []string              `json:"required_capabilities,omitempty"`
	MaxIterations               int                   `json:"max_iterations"`
	TaskTimeoutSeconds          int                   `json:"task_timeout_seconds"`
	InactivityTimeoutSeconds    int                   `json:"inactivity_timeout_seconds"`
	TransportTimeoutSeconds     int                   `json:"transport_timeout_seconds"`
	ToolPolicyMode              string                `json:"tool_policy_mode"`
	ToolAllowlistEffective      []string              `json:"tool_allowlist_effective"`
	BlockedToolNames            []string              `json:"blocked_tool_names"`
	Available                   bool                  `json:"available"`
	HermesAvailable             bool                  `json:"hermes_available"`
	OpenAIConfigured            bool                  `json:"openai_configured"`
	SlackMCPEnabled             bool                  `json:"slack_mcp_enabled,omitempty"`
	SlackMCPConfigured          bool                  `json:"slack_mcp_configured,omitempty"`
	SlackMCPAvailable           bool                  `json:"slack_mcp_available,omitempty"`
	SlackMCPServerURL           string                `json:"slack_mcp_server_url,omitempty"`
	SlackMCPToolCount           int                   `json:"slack_mcp_tool_count,omitempty"`
	PersistenceEnabled          bool                  `json:"persistence_enabled"`
	SessionContinuityStatus     string                `json:"session_continuity_status"`
	HermesHome                  string                `json:"hermes_home,omitempty"`
	HermesExecutorWorkspaceRoot string                `json:"hermes_executor_workspace_root,omitempty"`
	HermesComputerRoot          string                `json:"hermes_computer_root,omitempty"`
	HermesRunRoot               string                `json:"hermes_run_root,omitempty"`
	HermesArtifactRoot          string                `json:"hermes_artifact_root,omitempty"`
	SessionDBPath               string                `json:"session_db_path,omitempty"`
	ContextEngineMode           string                `json:"context_engine_mode,omitempty"`
	ContextEngineStatus         string                `json:"context_engine_status,omitempty"`
	LifecycleHookStatus         string                `json:"lifecycle_hook_status,omitempty"`
	MemoryBackend               string                `json:"memory_backend,omitempty"`
	HonchoConfigured            bool                  `json:"honcho_configured"`
	HonchoAvailable             bool                  `json:"honcho_available"`
	HonchoRuntimeStatus         map[string]any        `json:"honcho_runtime_status,omitempty"`
	HonchoBaseURL               string                `json:"honcho_base_url,omitempty"`
	HonchoWorkspace             string                `json:"honcho_workspace,omitempty"`
	HonchoEnvironment           string                `json:"honcho_environment,omitempty"`
	HonchoEnvironmentEffective  string                `json:"honcho_environment_effective,omitempty"`
	HonchoRecallMode            string                `json:"honcho_recall_mode,omitempty"`
	HonchoWriteFrequency        string                `json:"honcho_write_frequency,omitempty"`
	HonchoSessionStrategy       string                `json:"honcho_session_strategy,omitempty"`
	HonchoAIPeer                string                `json:"honcho_ai_peer,omitempty"`
}

type HonchoRuntimeResponse struct {
	Status             string                    `json:"status"`
	Namespace          string                    `json:"namespace"`
	DBSchema           string                    `json:"db_schema"`
	CacheEnabled       bool                      `json:"cache_enabled"`
	CacheURLConfigured bool                      `json:"cache_url_configured"`
	Deriver            RuntimeComponent          `json:"deriver"`
	Summary            RuntimeComponent          `json:"summary"`
	DialecticLevels    map[string]DialecticLevel `json:"dialectic_levels"`
}
