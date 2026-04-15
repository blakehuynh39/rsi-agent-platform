package harness

type RuntimeComponent struct {
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
}

type DialecticLevel struct {
	Provider             string `json:"provider"`
	Model                string `json:"model"`
	ReasoningEffort      string `json:"reasoning_effort"`
	ThinkingBudgetTokens int    `json:"thinking_budget_tokens"`
}

type RuntimeResponse struct {
	Status                   string   `json:"status"`
	Role                     string   `json:"role"`
	Backend                  string   `json:"backend"`
	Provider                 string   `json:"provider"`
	Model                    string   `json:"model"`
	ProviderModel            string   `json:"provider_model"`
	APIMode                  string   `json:"api_mode"`
	ReasoningEffort          string   `json:"reasoning_effort"`
	HermesVersion            string   `json:"hermes_version"`
	HermesPin                string   `json:"hermes_pin"`
	MaxIterations            int      `json:"max_iterations"`
	TaskTimeoutSeconds       int      `json:"task_timeout_seconds"`
	InactivityTimeoutSeconds int      `json:"inactivity_timeout_seconds"`
	TransportTimeoutSeconds  int      `json:"transport_timeout_seconds"`
	ToolPolicyMode           string   `json:"tool_policy_mode"`
	ToolAllowlistEffective   []string `json:"tool_allowlist_effective"`
	BlockedToolNames         []string `json:"blocked_tool_names"`
	Available                bool     `json:"available"`
	HermesAvailable          bool     `json:"hermes_available"`
	OpenAIConfigured         bool     `json:"openai_configured"`
	PersistenceEnabled       bool     `json:"persistence_enabled"`
	SessionContinuityStatus  string   `json:"session_continuity_status"`
	HermesHome               string   `json:"hermes_home,omitempty"`
	SessionDBPath            string   `json:"session_db_path,omitempty"`
	ContextEngineMode        string   `json:"context_engine_mode,omitempty"`
	ContextEngineStatus      string   `json:"context_engine_status,omitempty"`
	LifecycleHookStatus      string   `json:"lifecycle_hook_status,omitempty"`
	MemoryBackend            string   `json:"memory_backend,omitempty"`
	HonchoConfigured         bool     `json:"honcho_configured"`
	HonchoAvailable          bool     `json:"honcho_available"`
	HonchoBaseURL            string   `json:"honcho_base_url,omitempty"`
	HonchoWorkspace          string   `json:"honcho_workspace,omitempty"`
	HonchoEnvironment        string   `json:"honcho_environment,omitempty"`
	HonchoRecallMode         string   `json:"honcho_recall_mode,omitempty"`
	HonchoWriteFrequency     string   `json:"honcho_write_frequency,omitempty"`
	HonchoSessionStrategy    string   `json:"honcho_session_strategy,omitempty"`
	HonchoAIPeer             string   `json:"honcho_ai_peer,omitempty"`
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
