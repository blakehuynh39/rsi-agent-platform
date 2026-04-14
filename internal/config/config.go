package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName               string
	ServiceKind               string
	RuntimeMode               string
	ConfigValidated           bool
	SchemaVersionCurrent      int64
	SchemaVersionExpected     int64
	SchemaCompatibility       string
	Environment               string
	HTTPPort                  int
	StoreBackend              string
	PostgresURL               string
	RedisAddr                 string
	S3Bucket                  string
	PublicBaseURL             string
	WorkflowQueueURL          string
	ProactiveQueueURL         string
	EvalQueueURL              string
	ProposalQueueURL          string
	SandboxQueueURL           string
	RunnerBaseURL             string
	ProdRunnerBaseURL         string
	ProactiveRunnerBaseURL    string
	EvalRunnerBaseURL         string
	ProposalRunnerBaseURL     string
	ToolGatewayBaseURL        string
	HonchoRuntimeBaseURL      string
	ProdRunnerTimeout         time.Duration
	ProactiveRunnerTimeout    time.Duration
	EvalRunnerTimeout         time.Duration
	ProposalRunnerTimeout     time.Duration
	WorkerPollInterval        time.Duration
	WorkItemLeaseDuration     time.Duration
	SandboxPollInterval       time.Duration
	SlackAppIdentity          string
	SlackSocketModeEnabled    bool
	SlackAppToken             string
	SlackBotToken             string
	GitHubWebhookSecret       string
	GitHubOwner               string
	GitHubAPIBaseURL          string
	GitHubAppID               string
	GitHubAppInstallationID   string
	GitHubAppInstallationIDs  map[string]string
	GitHubAppPrivateKey       string
	GitHubRepoOwners          map[string]string
	GitHubCommitUser          string
	GitHubCommitEmail         string
	SentryAuthToken           string
	SentryOrganization        string
	SentryAPIBaseURL          string
	CloudflareAPIToken        string
	CloudflareAccountID       string
	CloudflareZoneID          string
	CloudflareAPIBaseURL      string
	KubeconfigPath            string
	KubernetesContext         string
	SandboxNamespace          string
	SandboxImage              string
	SandboxServiceAccount     string
	SandboxJobTTLSeconds      int
	SandboxDeadlineSeconds    int
	AllowedSlackChannelIDs    []string
	AllowedTargetRepos        []string
	DefaultOperatorDomain     string
	DefaultRepo               string
	DefaultKnowledgeBaseURL   string
	DefaultReasoningVerbosity string
	DefaultProposalCap        int
	ProposalPromoterInterval  time.Duration
}

func Load(serviceName string) Config {
	environment := stringEnv("RSI_ENV", "")
	runnerBaseURL := stringEnv("RSI_RUNNER_BASE_URL", "")
	return Config{
		ServiceName:               stringEnv("RSI_SERVICE_NAME", serviceName),
		ServiceKind:               serviceName,
		Environment:               environment,
		HTTPPort:                  intEnv("RSI_HTTP_PORT", 0),
		StoreBackend:              stringEnv("RSI_STORE_BACKEND", ""),
		PostgresURL:               stringEnv("RSI_POSTGRES_URL", ""),
		RedisAddr:                 stringEnv("RSI_REDIS_ADDR", ""),
		S3Bucket:                  stringEnv("RSI_S3_BUCKET", ""),
		PublicBaseURL:             stringEnv("RSI_PUBLIC_BASE_URL", ""),
		WorkflowQueueURL:          stringEnv("RSI_WORKFLOW_QUEUE_URL", ""),
		ProactiveQueueURL:         stringEnv("RSI_PROACTIVE_QUEUE_URL", ""),
		EvalQueueURL:              stringEnv("RSI_EVAL_QUEUE_URL", ""),
		ProposalQueueURL:          stringEnv("RSI_PROPOSAL_QUEUE_URL", ""),
		SandboxQueueURL:           stringEnv("RSI_SANDBOX_QUEUE_URL", ""),
		RunnerBaseURL:             runnerBaseURL,
		ProdRunnerBaseURL:         stringEnv("RSI_RUNNER_PROD_BASE_URL", ""),
		ProactiveRunnerBaseURL:    stringEnv("RSI_RUNNER_PROACTIVE_BASE_URL", ""),
		EvalRunnerBaseURL:         stringEnv("RSI_RUNNER_EVAL_BASE_URL", ""),
		ProposalRunnerBaseURL:     stringEnv("RSI_RUNNER_PROPOSAL_BASE_URL", ""),
		ToolGatewayBaseURL:        stringEnv("RSI_TOOL_GATEWAY_BASE_URL", ""),
		HonchoRuntimeBaseURL:      stringEnv("RSI_HONCHO_RUNTIME_BASE_URL", ""),
		ProdRunnerTimeout:         durationEnv("RSI_RUNNER_PROD_TIMEOUT", 60*time.Second),
		ProactiveRunnerTimeout:    durationEnv("RSI_RUNNER_PROACTIVE_TIMEOUT", 60*time.Second),
		EvalRunnerTimeout:         durationEnv("RSI_RUNNER_EVAL_TIMEOUT", 120*time.Second),
		ProposalRunnerTimeout:     durationEnv("RSI_RUNNER_PROPOSAL_TIMEOUT", 180*time.Second),
		WorkerPollInterval:        durationEnv("RSI_WORKER_POLL_INTERVAL", 5*time.Second),
		WorkItemLeaseDuration:     durationEnv("RSI_WORK_ITEM_LEASE_DURATION", 30*time.Second),
		SandboxPollInterval:       durationEnv("RSI_SANDBOX_POLL_INTERVAL", 10*time.Second),
		SlackAppIdentity:          stringEnv("RSI_SLACK_APP_IDENTITY", ""),
		SlackSocketModeEnabled:    boolEnv("RSI_SLACK_SOCKET_MODE_ENABLED", false),
		SlackAppToken:             stringEnv("RSI_SLACK_APP_TOKEN", ""),
		SlackBotToken:             stringEnv("RSI_SLACK_BOT_TOKEN", ""),
		GitHubWebhookSecret:       stringEnv("RSI_GITHUB_WEBHOOK_SECRET", ""),
		GitHubOwner:               stringEnv("RSI_GITHUB_OWNER", ""),
		GitHubAPIBaseURL:          stringEnv("RSI_GITHUB_API_BASE_URL", "https://api.github.com"),
		GitHubAppID:               stringEnv("RSI_GITHUB_APP_ID", ""),
		GitHubAppInstallationID:   stringEnv("RSI_GITHUB_APP_INSTALLATION_ID", ""),
		GitHubAppInstallationIDs:  mapEnv("RSI_GITHUB_APP_INSTALLATION_IDS", nil),
		GitHubAppPrivateKey:       stringEnv("RSI_GITHUB_APP_PRIVATE_KEY", ""),
		GitHubRepoOwners:          mapEnv("RSI_GITHUB_REPO_OWNERS", nil),
		GitHubCommitUser:          stringEnv("RSI_GITHUB_COMMIT_USER", ""),
		GitHubCommitEmail:         stringEnv("RSI_GITHUB_COMMIT_EMAIL", ""),
		SentryAuthToken:           stringEnv("RSI_SENTRY_AUTH_TOKEN", ""),
		SentryOrganization:        stringEnv("RSI_SENTRY_ORGANIZATION", ""),
		SentryAPIBaseURL:          stringEnv("RSI_SENTRY_API_BASE_URL", ""),
		CloudflareAPIToken:        stringEnv("RSI_CLOUDFLARE_API_TOKEN", ""),
		CloudflareAccountID:       stringEnv("RSI_CLOUDFLARE_ACCOUNT_ID", ""),
		CloudflareZoneID:          stringEnv("RSI_CLOUDFLARE_ZONE_ID", ""),
		CloudflareAPIBaseURL:      stringEnv("RSI_CLOUDFLARE_API_BASE_URL", ""),
		KubeconfigPath:            stringEnv("RSI_KUBECONFIG", ""),
		KubernetesContext:         stringEnv("RSI_KUBERNETES_CONTEXT", ""),
		SandboxNamespace:          stringEnv("RSI_SANDBOX_NAMESPACE", ""),
		SandboxImage:              stringEnv("RSI_SANDBOX_IMAGE", ""),
		SandboxServiceAccount:     stringEnv("RSI_SANDBOX_SERVICE_ACCOUNT_NAME", ""),
		SandboxJobTTLSeconds:      intEnv("RSI_SANDBOX_JOB_TTL_SECONDS", 0),
		SandboxDeadlineSeconds:    intEnv("RSI_SANDBOX_ACTIVE_DEADLINE_SECONDS", 0),
		AllowedSlackChannelIDs:    listEnv("RSI_ALLOWED_SLACK_CHANNEL_IDS", nil),
		AllowedTargetRepos:        listEnv("RSI_ALLOWED_TARGET_REPOS", nil),
		DefaultOperatorDomain:     stringEnv("RSI_OPERATOR_EMAIL_DOMAIN", ""),
		DefaultRepo:               stringEnv("RSI_DEFAULT_REPO", ""),
		DefaultKnowledgeBaseURL:   stringEnv("RSI_KNOWLEDGE_BASE_URL", ""),
		DefaultReasoningVerbosity: stringEnv("RSI_REASONING_VERBOSITY", ""),
		DefaultProposalCap:        intEnv("RSI_ACTIVE_PROPOSAL_CAP", 0),
		ProposalPromoterInterval:  durationEnv("RSI_PROPOSAL_PROMOTER_INTERVAL", 0),
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

func (c Config) RunnerTimeoutForRole(role string) time.Duration {
	switch strings.TrimSpace(role) {
	case "prod":
		return c.ProdRunnerTimeout
	case "proactive":
		return c.ProactiveRunnerTimeout
	case "eval":
		return c.EvalRunnerTimeout
	case "proposal":
		return c.ProposalRunnerTimeout
	default:
		return c.ProdRunnerTimeout
	}
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
		return fallback
	}
	return value
}

func listEnv(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func mapEnv(key string, fallback map[string]string) map[string]string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	out := make(map[string]string)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		keyPart, valuePart, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		keyPart = strings.TrimSpace(keyPart)
		valuePart = strings.TrimSpace(valuePart)
		if keyPart == "" || valuePart == "" {
			continue
		}
		out[keyPart] = valuePart
	}
	if len(out) == 0 {
		return fallback
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
		return fallback
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
		return fallback
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
