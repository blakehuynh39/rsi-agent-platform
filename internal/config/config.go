package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName                          string
	ServiceKind                          string
	RuntimeMode                          string
	ConfigValidated                      bool
	SchemaVersionCurrent                 int64
	SchemaVersionExpected                int64
	SchemaCompatibility                  string
	Environment                          string
	HTTPPort                             int
	StoreBackend                         string
	PostgresURL                          string
	RedisAddr                            string
	S3Bucket                             string
	PublicBaseURL                        string
	WorkflowQueueURL                     string
	ProactiveQueueURL                    string
	EvalQueueURL                         string
	ProposalQueueURL                     string
	SandboxQueueURL                      string
	RunnerBaseURL                        string
	HermesExecutorBaseURL                string
	HermesExecutorPoolURLs               []string
	HonchoBaseURL                        string
	HonchoAPIKey                         string
	HonchoWorkspaceID                    string
	ProdRunnerBaseURL                    string
	ProactiveRunnerBaseURL               string
	EvalRunnerBaseURL                    string
	ProposalRunnerBaseURL                string
	HonchoRuntimeBaseURL                 string
	ProdRunnerTimeout                    time.Duration
	ProactiveRunnerTimeout               time.Duration
	EvalRunnerTimeout                    time.Duration
	ProposalRunnerTimeout                time.Duration
	ProdRunnerTaskTimeout                time.Duration
	ProactiveRunnerTaskTimeout           time.Duration
	EvalRunnerTaskTimeout                time.Duration
	ProposalRunnerTaskTimeout            time.Duration
	WorkerPollInterval                   time.Duration
	WorkItemLeaseDuration                time.Duration
	SandboxPollInterval                  time.Duration
	SlackAppIdentity                     string
	SlackSocketModeEnabled               bool
	SlackAppToken                        string
	SlackBotToken                        string
	SlackWorkspaceID                     string
	SlackMirrorEnabled                   bool
	SlackMirrorChannelDiscovery          string
	SlackMirrorChannelAllowlist          []string
	SlackMirrorChannelDenylist           []string
	SourceMirrorCheckpointRoot           string
	AttachmentCacheRoot                  string
	CompanyWikiRoot                      string
	CompanyWikiSynthesisEnabled          bool
	CompanyWikiSourcePageMode            string
	CompanyWikiCompilerModel             string
	CompanyWikiCompilerBatchLimit        int
	CompanyWikiCompilerChunkLimit        int
	CompanyWikiCompilerTimeout           time.Duration
	CompanyWikiCompilerRunTimeout        time.Duration
	CompanyWikiCompilerShutdownGrace     time.Duration
	CompanyWikiCompilerOpenRouterBaseURL string
	CompanyWikiCompilerOpenRouterAPIKey  string
	NativeToolsEnabled                   bool
	NativeToolsClientToken               string
	NativeToolsSurfaces                  []string
	DBReadEnabled                        bool
	DBReadTargetsJSON                    string
	DBReadClientToken                    string
	DBReadApproverSlackUserIDs           []string
	DBReadWorkerTargets                  []string
	NotionToken                          string
	NotionMirrorEnabled                  bool
	NotionMirrorAllowlist                []string
	NotionMirrorRequestsPerSecond        float64
	NotionMirrorMaxRetries               int
	NotionMirrorRetryBaseDelay           time.Duration
	NotionMirrorMaxDatabasesPerRoot      int
	NotionMirrorMaxBlocksPerPage         int
	NotionMirrorMaxDepth                 int
	NotionMirrorMaxDocumentBytes         int
	NotionMirrorDeltaEnabled             bool
	NotionMirrorDeltaLookback            time.Duration
	NotionMirrorFullScanInterval         time.Duration
	NotionAPIBaseURL                     string
	NotionAPIVersion                     string
	GitHubWebhookSecret                  string
	GitHubOwner                          string
	GitHubAPIBaseURL                     string
	GitHubAppID                          string
	GitHubAppInstallationID              string
	GitHubAppInstallationIDs             map[string]string
	GitHubAppPrivateKey                  string
	GitHubRepoOwners                     map[string]string
	GitHubCommitUser                     string
	GitHubCommitEmail                    string
	SentryAuthToken                      string
	SentryOrganization                   string
	SentryAPIBaseURL                     string
	CloudflareAPIToken                   string
	CloudflareAccountID                  string
	CloudflareZoneID                     string
	CloudflareAPIBaseURL                 string
	KubeconfigPath                       string
	KubernetesContext                    string
	SandboxNamespace                     string
	SandboxImage                         string
	SandboxServiceAccount                string
	SandboxJobTTLSeconds                 int
	SandboxDeadlineSeconds               int
	SlackIngressAllowedChannelIDs        []string
	AllowedTargetRepos                   []string
	DefaultOperatorDomain                string
	DefaultRepo                          string
	DefaultKnowledgeBaseURL              string
	DefaultReasoningVerbosity            string
	DefaultProposalCap                   int
	ProposalPromoterInterval             time.Duration
	VerboseTraceLogging                  bool
	VerboseTraceLogLimit                 int
	WorkflowRunnerRepairAttempts         int
	WorkflowAutoRetryEnabled             bool
	WorkflowAutoRetryMaxAttempts         int
	WorkflowAutoRetryBackoffSeconds      []int
	SlackAckAfterDurableIngress          bool
	SlackDurableIngressAckTimeout        time.Duration
	EffectFairClaimEnabled               bool
	EffectMaxConcurrentPerScope          int
	AsyncHermesExecutionEnabled          bool
	HermesExecutionHeartbeatTimeout      time.Duration
	DrainEnabled                         bool
	DeploymentActiveExecutionPolicy      string
	HermesComputerRoot                   string
	HermesRunRoot                        string
	HermesArtifactRoot                   string
	ExecutionEnvelopeV1Enabled           bool
	ExecutionLedgerFirstProjection       bool
	RunnerPlannerMode                    string
	RuntimeDiagnosisEnabled              bool
	RuntimeDiagnosisLogFallbackEnabled   bool
}

const (
	defaultProdRunnerTimeout          = 1830 * time.Second
	defaultProactiveRunnerTimeout     = 330 * time.Second
	defaultEvalRunnerTimeout          = 330 * time.Second
	defaultProposalRunnerTimeout      = 450 * time.Second
	defaultProdRunnerTaskTimeout      = 1800 * time.Second
	defaultProactiveRunnerTaskTimeout = 300 * time.Second
	defaultEvalRunnerTaskTimeout      = 300 * time.Second
	defaultProposalRunnerTaskTimeout  = 420 * time.Second
)

type slackUserAllowlistEntry struct {
	ID   string
	Name string
}

// defaultDBReadApproverSlackUsers is the source-controlled DB-read approval allowlist.
// Slack user IDs are not secrets; update this list by PR when DB-read approvers change.
var defaultDBReadApproverSlackUsers = []slackUserAllowlistEntry{
	{ID: "U0772SH7BRA", Name: "Blake"},
	{ID: "U04L0DD6B6F", Name: "Allen"},
	{ID: "U08V4SFU7LZ", Name: "Romain Magne"},
	{ID: "U06A5AQ1VD3", Name: "Andrea Muttoni"},
	{ID: "U067QP5PD6J", Name: "Jongwon Park"},
	{ID: "U083MMT1771", Name: "Aiwei"},
}

func defaultDBReadApproverSlackUserIDs() []string {
	out := make([]string, 0, len(defaultDBReadApproverSlackUsers))
	for _, user := range defaultDBReadApproverSlackUsers {
		out = append(out, user.ID)
	}
	return out
}

func Load(serviceName string) Config {
	environment := stringEnv("RSI_ENV", "")
	runnerBaseURL := stringEnv("RSI_RUNNER_BASE_URL", "")
	dbReadApprovers := CompactUniqueStrings(append(defaultDBReadApproverSlackUserIDs(), listEnv("RSI_DB_READ_APPROVER_SLACK_USER_IDS")...))
	return Config{
		ServiceName:                          stringEnv("RSI_SERVICE_NAME", serviceName),
		ServiceKind:                          serviceName,
		Environment:                          environment,
		HTTPPort:                             intEnv("RSI_HTTP_PORT", 0),
		StoreBackend:                         stringEnv("RSI_STORE_BACKEND", ""),
		PostgresURL:                          stringEnv("RSI_POSTGRES_URL", ""),
		RedisAddr:                            stringEnv("RSI_REDIS_ADDR", ""),
		S3Bucket:                             stringEnv("RSI_S3_BUCKET", ""),
		PublicBaseURL:                        stringEnv("RSI_PUBLIC_BASE_URL", ""),
		WorkflowQueueURL:                     stringEnv("RSI_WORKFLOW_QUEUE_URL", ""),
		ProactiveQueueURL:                    stringEnv("RSI_PROACTIVE_QUEUE_URL", ""),
		EvalQueueURL:                         stringEnv("RSI_EVAL_QUEUE_URL", ""),
		ProposalQueueURL:                     stringEnv("RSI_PROPOSAL_QUEUE_URL", ""),
		SandboxQueueURL:                      stringEnv("RSI_SANDBOX_QUEUE_URL", ""),
		RunnerBaseURL:                        runnerBaseURL,
		HermesExecutorBaseURL:                stringEnv("RSI_HERMES_EXECUTOR_BASE_URL", ""),
		HermesExecutorPoolURLs:               listEnv("RSI_HERMES_EXECUTOR_POOL_URLS"),
		HonchoBaseURL:                        stringEnv("RSI_HONCHO_BASE_URL", ""),
		HonchoAPIKey:                         stringEnv("HONCHO_API_KEY", ""),
		HonchoWorkspaceID:                    stringEnv("RSI_HONCHO_WORKSPACE_ID", "rsi_company_knowledge"),
		ProdRunnerBaseURL:                    stringEnv("RSI_RUNNER_PROD_BASE_URL", ""),
		ProactiveRunnerBaseURL:               stringEnv("RSI_RUNNER_PROACTIVE_BASE_URL", ""),
		EvalRunnerBaseURL:                    stringEnv("RSI_RUNNER_EVAL_BASE_URL", ""),
		ProposalRunnerBaseURL:                stringEnv("RSI_RUNNER_PROPOSAL_BASE_URL", ""),
		HonchoRuntimeBaseURL:                 stringEnv("RSI_HONCHO_RUNTIME_BASE_URL", ""),
		ProdRunnerTimeout:                    durationEnv("RSI_RUNNER_PROD_TIMEOUT", defaultProdRunnerTimeout),
		ProactiveRunnerTimeout:               durationEnv("RSI_RUNNER_PROACTIVE_TIMEOUT", defaultProactiveRunnerTimeout),
		EvalRunnerTimeout:                    durationEnv("RSI_RUNNER_EVAL_TIMEOUT", defaultEvalRunnerTimeout),
		ProposalRunnerTimeout:                durationEnv("RSI_RUNNER_PROPOSAL_TIMEOUT", defaultProposalRunnerTimeout),
		ProdRunnerTaskTimeout:                durationEnv("RSI_RUNNER_PROD_TASK_TIMEOUT", defaultProdRunnerTaskTimeout),
		ProactiveRunnerTaskTimeout:           durationEnv("RSI_RUNNER_PROACTIVE_TASK_TIMEOUT", defaultProactiveRunnerTaskTimeout),
		EvalRunnerTaskTimeout:                durationEnv("RSI_RUNNER_EVAL_TASK_TIMEOUT", defaultEvalRunnerTaskTimeout),
		ProposalRunnerTaskTimeout:            durationEnv("RSI_RUNNER_PROPOSAL_TASK_TIMEOUT", defaultProposalRunnerTaskTimeout),
		WorkerPollInterval:                   durationEnv("RSI_WORKER_POLL_INTERVAL", 5*time.Second),
		WorkItemLeaseDuration:                durationEnv("RSI_WORK_ITEM_LEASE_DURATION", 30*time.Second),
		SandboxPollInterval:                  durationEnv("RSI_SANDBOX_POLL_INTERVAL", 10*time.Second),
		SlackAppIdentity:                     stringEnv("RSI_SLACK_APP_IDENTITY", ""),
		SlackSocketModeEnabled:               boolEnv("RSI_SLACK_SOCKET_MODE_ENABLED", false),
		SlackAppToken:                        stringEnv("RSI_SLACK_APP_TOKEN", ""),
		SlackBotToken:                        stringEnv("SLACK_BOT_TOKEN", ""),
		SlackWorkspaceID:                     stringEnv("RSI_SLACK_WORKSPACE_ID", ""),
		SlackMirrorEnabled:                   boolEnv("RSI_SLACK_MIRROR_ENABLED", false),
		SlackMirrorChannelDiscovery:          stringEnv("RSI_SLACK_MIRROR_CHANNEL_DISCOVERY", "joined"),
		SlackMirrorChannelAllowlist:          listEnv("RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST"),
		SlackMirrorChannelDenylist:           listEnv("RSI_SLACK_MIRROR_CHANNEL_DENYLIST"),
		SourceMirrorCheckpointRoot:           stringEnv("RSI_SOURCE_MIRROR_CHECKPOINT_ROOT", "/var/lib/hermes/source-mirror"),
		AttachmentCacheRoot:                  stringEnv("RSI_ATTACHMENT_CACHE_ROOT", "/var/lib/hermes/attachments"),
		CompanyWikiRoot:                      stringEnv("RSI_COMPANY_WIKI_ROOT", "/workspace/company/wiki"),
		CompanyWikiSynthesisEnabled:          boolEnv("RSI_COMPANY_WIKI_SYNTHESIS_ENABLED", false),
		CompanyWikiSourcePageMode:            stringEnv("RSI_COMPANY_WIKI_SOURCE_PAGE_MODE", "evidence"),
		CompanyWikiCompilerModel:             stringEnv("RSI_COMPANY_WIKI_COMPILER_MODEL", ""),
		CompanyWikiCompilerBatchLimit:        intEnv("RSI_COMPANY_WIKI_COMPILER_BATCH_LIMIT", 10),
		CompanyWikiCompilerChunkLimit:        intEnv("RSI_COMPANY_WIKI_COMPILER_CHUNK_LIMIT", 24),
		CompanyWikiCompilerTimeout:           durationEnv("RSI_COMPANY_WIKI_COMPILER_TIMEOUT", 120*time.Second),
		CompanyWikiCompilerRunTimeout:        durationEnv("RSI_COMPANY_WIKI_COMPILER_RUN_TIMEOUT", 25*time.Minute),
		CompanyWikiCompilerShutdownGrace:     durationEnv("RSI_COMPANY_WIKI_COMPILER_SHUTDOWN_GRACE", 30*time.Second),
		CompanyWikiCompilerOpenRouterBaseURL: stringEnv("RSI_COMPANY_WIKI_COMPILER_OPENROUTER_BASE_URL", stringEnv("RSI_OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1")),
		CompanyWikiCompilerOpenRouterAPIKey:  firstNonEmpty(stringEnv("RSI_COMPANY_WIKI_COMPILER_OPENROUTER_API_KEY", ""), stringEnv("RSI_OPENROUTER_API_KEY", ""), stringEnv("OPENROUTER_API_KEY", "")),
		NativeToolsEnabled:                   boolEnv("RSI_NATIVE_TOOLS_ENABLED", true),
		NativeToolsClientToken:               stringEnv("RSI_NATIVE_TOOLS_CLIENT_TOKEN", ""),
		NativeToolsSurfaces:                  listEnvWithDefault("RSI_NATIVE_TOOLS_SURFACES", []string{"slack", "notion", "knowledge"}),
		DBReadEnabled:                        boolEnv("RSI_DB_READ_ENABLED", false),
		DBReadTargetsJSON:                    stringEnv("RSI_DB_READ_TARGETS_JSON", ""),
		DBReadClientToken:                    stringEnv("RSI_DB_READ_CLIENT_TOKEN", ""),
		DBReadApproverSlackUserIDs:           dbReadApprovers,
		DBReadWorkerTargets:                  listEnv("RSI_DB_READ_WORKER_TARGETS"),
		NotionToken:                          stringEnv("NOTION_TOKEN", ""),
		NotionMirrorEnabled:                  boolEnv("RSI_NOTION_MIRROR_ENABLED", false),
		NotionMirrorAllowlist:                listEnv("RSI_NOTION_MIRROR_ALLOWLIST"),
		NotionMirrorRequestsPerSecond:        floatEnv("RSI_NOTION_MIRROR_REQUESTS_PER_SECOND", 3),
		NotionMirrorMaxRetries:               intEnv("RSI_NOTION_MIRROR_MAX_RETRIES", 3),
		NotionMirrorRetryBaseDelay:           durationEnv("RSI_NOTION_MIRROR_RETRY_BASE_DELAY", 500*time.Millisecond),
		NotionMirrorMaxDatabasesPerRoot:      intEnv("RSI_NOTION_MIRROR_MAX_DATABASES_PER_ROOT", 50),
		NotionMirrorMaxBlocksPerPage:         intEnv("RSI_NOTION_MIRROR_MAX_BLOCKS_PER_PAGE", 1000),
		NotionMirrorMaxDepth:                 intEnv("RSI_NOTION_MIRROR_MAX_DEPTH", 4),
		NotionMirrorMaxDocumentBytes:         intEnv("RSI_NOTION_MIRROR_MAX_DOCUMENT_BYTES", 256000),
		NotionMirrorDeltaEnabled:             boolEnv("RSI_NOTION_MIRROR_DELTA_ENABLED", true),
		NotionMirrorDeltaLookback:            durationEnv("RSI_NOTION_MIRROR_DELTA_LOOKBACK", 10*time.Minute),
		NotionMirrorFullScanInterval:         durationEnv("RSI_NOTION_MIRROR_FULL_SCAN_INTERVAL", 24*time.Hour),
		NotionAPIBaseURL:                     stringEnv("RSI_NOTION_API_BASE_URL", "https://api.notion.com"),
		NotionAPIVersion:                     stringEnv("RSI_NOTION_API_VERSION", "2026-03-11"),
		GitHubWebhookSecret:                  stringEnv("RSI_GITHUB_WEBHOOK_SECRET", ""),
		GitHubOwner:                          stringEnv("RSI_GITHUB_OWNER", ""),
		GitHubAPIBaseURL:                     stringEnv("RSI_GITHUB_API_BASE_URL", "https://api.github.com"),
		GitHubAppID:                          stringEnv("RSI_GITHUB_APP_ID", ""),
		GitHubAppInstallationID:              stringEnv("RSI_GITHUB_APP_INSTALLATION_ID", ""),
		GitHubAppInstallationIDs:             mapEnv("RSI_GITHUB_APP_INSTALLATION_IDS"),
		GitHubAppPrivateKey:                  stringEnv("RSI_GITHUB_APP_PRIVATE_KEY", ""),
		GitHubRepoOwners:                     mapEnv("RSI_GITHUB_REPO_OWNERS"),
		GitHubCommitUser:                     stringEnv("RSI_GITHUB_COMMIT_USER", ""),
		GitHubCommitEmail:                    stringEnv("RSI_GITHUB_COMMIT_EMAIL", ""),
		SentryAuthToken:                      stringEnv("RSI_SENTRY_AUTH_TOKEN", ""),
		SentryOrganization:                   stringEnv("RSI_SENTRY_ORGANIZATION", ""),
		SentryAPIBaseURL:                     stringEnv("RSI_SENTRY_API_BASE_URL", ""),
		CloudflareAPIToken:                   stringEnv("RSI_CLOUDFLARE_API_TOKEN", ""),
		CloudflareAccountID:                  stringEnv("RSI_CLOUDFLARE_ACCOUNT_ID", ""),
		CloudflareZoneID:                     stringEnv("RSI_CLOUDFLARE_ZONE_ID", ""),
		CloudflareAPIBaseURL:                 stringEnv("RSI_CLOUDFLARE_API_BASE_URL", ""),
		KubeconfigPath:                       stringEnv("RSI_KUBECONFIG", ""),
		KubernetesContext:                    stringEnv("RSI_KUBERNETES_CONTEXT", ""),
		SandboxNamespace:                     stringEnv("RSI_SANDBOX_NAMESPACE", ""),
		SandboxImage:                         stringEnv("RSI_SANDBOX_IMAGE", ""),
		SandboxServiceAccount:                stringEnv("RSI_SANDBOX_SERVICE_ACCOUNT_NAME", ""),
		SandboxJobTTLSeconds:                 intEnv("RSI_SANDBOX_JOB_TTL_SECONDS", 0),
		SandboxDeadlineSeconds:               intEnv("RSI_SANDBOX_ACTIVE_DEADLINE_SECONDS", 0),
		SlackIngressAllowedChannelIDs:        listEnv("RSI_SLACK_INGRESS_ALLOWED_CHANNEL_IDS"),
		AllowedTargetRepos:                   listEnv("RSI_ALLOWED_TARGET_REPOS"),
		DefaultOperatorDomain:                stringEnv("RSI_OPERATOR_EMAIL_DOMAIN", ""),
		DefaultRepo:                          stringEnv("RSI_DEFAULT_REPO", ""),
		DefaultKnowledgeBaseURL:              stringEnv("RSI_KNOWLEDGE_BASE_URL", ""),
		DefaultReasoningVerbosity:            stringEnv("RSI_REASONING_VERBOSITY", ""),
		DefaultProposalCap:                   intEnv("RSI_ACTIVE_PROPOSAL_CAP", 0),
		ProposalPromoterInterval:             durationEnv("RSI_PROPOSAL_PROMOTER_INTERVAL", 0),
		VerboseTraceLogging:                  boolEnv("RSI_VERBOSE_TRACE_LOGGING", false),
		VerboseTraceLogLimit:                 intEnv("RSI_VERBOSE_TRACE_LOG_LIMIT", 100000),
		WorkflowRunnerRepairAttempts:         intEnv("RSI_WORKFLOW_RUNNER_REPAIR_ATTEMPTS", 2),
		WorkflowAutoRetryEnabled:             boolEnv("RSI_WORKFLOW_AUTO_RETRY_ENABLED", false),
		WorkflowAutoRetryMaxAttempts:         intEnv("RSI_WORKFLOW_AUTO_RETRY_MAX_ATTEMPTS", 3),
		WorkflowAutoRetryBackoffSeconds:      intListEnv("RSI_WORKFLOW_AUTO_RETRY_BACKOFF_SECONDS", []int{15, 60}),
		SlackAckAfterDurableIngress:          boolEnv("RSI_SLACK_ACK_AFTER_DURABLE_INGRESS", false),
		SlackDurableIngressAckTimeout:        durationEnv("RSI_SLACK_DURABLE_INGRESS_ACK_TIMEOUT", 2*time.Second),
		EffectFairClaimEnabled:               boolEnv("RSI_EFFECT_FAIR_CLAIM_ENABLED", false),
		EffectMaxConcurrentPerScope:          intEnv("RSI_EFFECT_MAX_CONCURRENT_PER_SCOPE", 1),
		AsyncHermesExecutionEnabled:          boolEnv("RSI_ASYNC_HERMES_EXECUTION_ENABLED", false),
		HermesExecutionHeartbeatTimeout:      durationEnv("RSI_HERMES_EXECUTION_HEARTBEAT_TIMEOUT", 120*time.Second),
		DrainEnabled:                         boolEnv("RSI_DRAIN_ENABLED", false),
		DeploymentActiveExecutionPolicy:      stringEnv("RSI_DEPLOYMENT_ACTIVE_EXECUTION_POLICY", ""),
		HermesComputerRoot:                   stringEnv("RSI_HERMES_COMPUTER_ROOT", "/workspace/company"),
		HermesRunRoot:                        stringEnv("RSI_HERMES_RUN_ROOT", "/workspace/company/.rsi/runs"),
		HermesArtifactRoot:                   stringEnv("RSI_HERMES_ARTIFACT_ROOT", "/workspace/company/artifacts"),
		ExecutionEnvelopeV1Enabled:           boolEnv("RSI_EXECUTION_ENVELOPE_V1_ENABLED", true),
		ExecutionLedgerFirstProjection:       boolEnv("RSI_EXECUTION_LEDGER_FIRST_PROJECTION_ENABLED", false),
		RunnerPlannerMode:                    stringEnv("RSI_RUNNER_PLANNER_MODE", "runner_first"),
		RuntimeDiagnosisEnabled:              boolEnv("RSI_RUNTIME_DIAGNOSIS_ENABLED", false),
		RuntimeDiagnosisLogFallbackEnabled:   boolEnv("RSI_RUNTIME_DIAGNOSIS_LOG_FALLBACK_ENABLED", false),
	}
}

func (c Config) RunnerURLForRole(role string) string {
	switch strings.TrimSpace(role) {
	case "prod":
		return c.ProdRunnerBaseURL
	case "proactive":
		return c.ProactiveRunnerBaseURL
	case "eval":
		return c.EvalRunnerBaseURL
	case "proposal":
		return c.ProposalRunnerBaseURL
	default:
		return c.RunnerBaseURL
	}
}

func (c Config) RunnerURLs() map[string]string {
	return map[string]string{
		"prod":      c.RunnerURLForRole("prod"),
		"proactive": c.RunnerURLForRole("proactive"),
		"eval":      c.RunnerURLForRole("eval"),
		"proposal":  c.RunnerURLForRole("proposal"),
	}
}

func (c Config) HermesExecutorURLs() []string {
	values := append([]string(nil), c.HermesExecutorPoolURLs...)
	if len(values) == 0 && strings.TrimSpace(c.HermesExecutorBaseURL) != "" {
		values = append(values, c.HermesExecutorBaseURL)
	}
	return CompactUniqueStrings(values)
}

func (c Config) RunnerTimeoutForRole(role string) time.Duration {
	switch strings.TrimSpace(role) {
	case "prod":
		if c.ProdRunnerTimeout > 0 {
			return c.ProdRunnerTimeout
		}
		return defaultProdRunnerTimeout
	case "proactive":
		if c.ProactiveRunnerTimeout > 0 {
			return c.ProactiveRunnerTimeout
		}
		return defaultProactiveRunnerTimeout
	case "eval":
		if c.EvalRunnerTimeout > 0 {
			return c.EvalRunnerTimeout
		}
		return defaultEvalRunnerTimeout
	case "proposal":
		if c.ProposalRunnerTimeout > 0 {
			return c.ProposalRunnerTimeout
		}
		return defaultProposalRunnerTimeout
	default:
		if c.ProdRunnerTimeout > 0 {
			return c.ProdRunnerTimeout
		}
		return defaultProdRunnerTimeout
	}
}

func (c Config) RunnerTaskTimeoutForRole(role string) time.Duration {
	switch strings.TrimSpace(role) {
	case "prod":
		if c.ProdRunnerTaskTimeout > 0 {
			return c.ProdRunnerTaskTimeout
		}
		return defaultProdRunnerTaskTimeout
	case "proactive":
		if c.ProactiveRunnerTaskTimeout > 0 {
			return c.ProactiveRunnerTaskTimeout
		}
		return defaultProactiveRunnerTaskTimeout
	case "eval":
		if c.EvalRunnerTaskTimeout > 0 {
			return c.EvalRunnerTaskTimeout
		}
		return defaultEvalRunnerTaskTimeout
	case "proposal":
		if c.ProposalRunnerTaskTimeout > 0 {
			return c.ProposalRunnerTaskTimeout
		}
		return defaultProposalRunnerTaskTimeout
	default:
		if c.ProdRunnerTaskTimeout > 0 {
			return c.ProdRunnerTaskTimeout
		}
		return defaultProdRunnerTaskTimeout
	}
}

func (c Config) EffectLeaseDuration(base time.Duration, roles ...string) time.Duration {
	lease := base
	for _, role := range roles {
		timeout := c.RunnerTimeoutForRole(role)
		if timeout <= 0 {
			continue
		}
		if candidate := timeout + 30*time.Second; candidate > lease {
			lease = candidate
		}
	}
	return lease
}

func (c Config) GitHubRepoOwner(repo string) string {
	repo = strings.TrimSpace(repo)
	if owner, _, ok := splitGitHubRepo(repo); ok {
		return owner
	}
	if owner, ok := c.GitHubRepoOwners[repo]; ok && strings.TrimSpace(owner) != "" {
		return strings.TrimSpace(owner)
	}
	return strings.TrimSpace(c.GitHubOwner)
}

func (c Config) GitHubRepoName(repo string) string {
	repo = strings.TrimSpace(repo)
	if _, name, ok := splitGitHubRepo(repo); ok {
		return name
	}
	return repo
}

func (c Config) GitHubInstallationIDForRepo(repo string) string {
	owner := c.GitHubRepoOwner(repo)
	if owner == "" {
		return ""
	}
	if strings.EqualFold(owner, strings.TrimSpace(c.GitHubOwner)) {
		return strings.TrimSpace(c.GitHubAppInstallationID)
	}
	if id, ok := c.GitHubAppInstallationIDs[owner]; ok {
		return strings.TrimSpace(id)
	}
	return ""
}

func stringEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		panic(fmt.Errorf("%s must be a valid integer: %q: %w", key, raw, err))
	}
	return value
}

func floatEnv(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		panic(fmt.Errorf("%s must be a valid number: %q: %w", key, raw, err))
	}
	return value
}

func listEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			panic(fmt.Errorf("%s must not contain empty list entries: %q", key, raw))
		}
		out = append(out, part)
	}
	return out
}

func listEnvWithDefault(key string, fallback []string) []string {
	values := listEnv(key)
	if len(values) == 0 {
		return append([]string(nil), fallback...)
	}
	return values
}

func CompactUniqueStrings(values []string) []string {
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

func intListEnv(key string, fallback []int) []int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return append([]int(nil), fallback...)
	}
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			panic(fmt.Errorf("%s must not contain empty list entries: %q", key, raw))
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			panic(fmt.Errorf("%s must contain integers: %q: %w", key, raw, err))
		}
		if value <= 0 {
			panic(fmt.Errorf("%s values must be positive integers: %q", key, raw))
		}
		out = append(out, value)
	}
	return out
}

func mapEnv(key string) map[string]string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	out := make(map[string]string)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			panic(fmt.Errorf("%s must not contain empty map entries: %q", key, raw))
		}
		keyPart, valuePart, ok := strings.Cut(part, "=")
		if !ok {
			panic(fmt.Errorf("%s must use key=value pairs: %q", key, raw))
		}
		keyPart = strings.TrimSpace(keyPart)
		valuePart = strings.TrimSpace(valuePart)
		if keyPart == "" || valuePart == "" {
			panic(fmt.Errorf("%s must use non-empty key=value pairs: %q", key, raw))
		}
		out[keyPart] = valuePart
	}
	return out
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		panic(fmt.Errorf("%s must be a valid duration: %q: %w", key, raw, err))
	}
	return value
}

func boolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		panic(fmt.Errorf("%s must be a valid boolean: %q: %w", key, raw, err))
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func splitGitHubRepo(repo string) (string, string, bool) {
	owner, name, ok := strings.Cut(strings.TrimSpace(repo), "/")
	if !ok {
		return "", "", false
	}
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)
	if owner == "" || name == "" {
		return "", "", false
	}
	return owner, name, true
}
