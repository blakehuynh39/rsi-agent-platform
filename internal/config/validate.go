package config

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
)

type ValidationError struct {
	Issues []string
}

func (e ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "configuration validation failed"
	}
	return "configuration validation failed: " + strings.Join(e.Issues, "; ")
}

func (c Config) ValidatedFor(serviceKind string, mode string) (Config, error) {
	cfg := c
	cfg.ServiceKind = firstNonEmpty(serviceKind, c.ServiceKind, c.ServiceName)
	cfg.RuntimeMode = strings.TrimSpace(mode)
	issues := cfg.validate()
	cfg.ConfigValidated = len(issues) == 0
	if len(issues) > 0 {
		return cfg, ValidationError{Issues: issues}
	}
	return cfg, nil
}

func (c Config) DependencyTargets() map[string]string {
	targets := map[string]string{
		"public_base_url": c.PublicBaseURL,
	}
	if c.HermesExecutorBaseURL != "" {
		targets["hermes_executor"] = c.HermesExecutorBaseURL
	}
	for index, target := range c.HermesExecutorPoolURLs {
		targets[fmt.Sprintf("hermes_executor_pool_%d", index)] = target
	}
	if c.HonchoRuntimeBaseURL != "" {
		targets["honcho_runtime"] = c.HonchoRuntimeBaseURL
	}
	if c.HonchoBaseURL != "" {
		targets["honcho"] = c.HonchoBaseURL
	}
	if c.ProdRunnerBaseURL != "" {
		targets["runner_prod"] = c.ProdRunnerBaseURL
	}
	if c.ProactiveRunnerBaseURL != "" {
		targets["runner_proactive"] = c.ProactiveRunnerBaseURL
	}
	if c.EvalRunnerBaseURL != "" {
		targets["runner_eval"] = c.EvalRunnerBaseURL
	}
	if c.ProposalRunnerBaseURL != "" {
		targets["runner_proposal"] = c.ProposalRunnerBaseURL
	}
	return targets
}

func (c Config) validate() []string {
	issues := make([]string, 0)
	addRequiredString(&issues, "RSI_ENV", c.Environment)
	if c.HTTPPort <= 0 {
		issues = append(issues, "RSI_HTTP_PORT must be set to a positive integer")
	}

	switch c.ServiceKind {
	case "control-plane":
		c.validateControlPlane(&issues)
	case "improvement-plane":
		c.validateImprovementPlane(&issues)
	}

	sort.Strings(issues)
	return issues
}

func (c Config) validateControlPlane(issues *[]string) {
	c.validateCommonPlaneConfig(issues)
	if c.RuntimeMode == "slack-mirror" {
		c.validateSlackMirror(issues)
		return
	}
	if c.RuntimeMode == "notion-mirror" {
		c.validateNotionMirror(issues)
		return
	}
	if c.RuntimeMode == "source-mirror-health" {
		c.validateSourceMirrorHealth(issues)
		return
	}
	if len(c.HermesExecutorURLs()) == 0 {
		addRequiredURL(issues, "RSI_RUNNER_PROD_BASE_URL", c.ProdRunnerBaseURL, c.nonLocalhostRequired())
		addRequiredURL(issues, "RSI_RUNNER_PROACTIVE_BASE_URL", c.ProactiveRunnerBaseURL, c.nonLocalhostRequired())
	}
	if strings.TrimSpace(c.HermesExecutorBaseURL) != "" {
		addRequiredURL(issues, "RSI_HERMES_EXECUTOR_BASE_URL", c.HermesExecutorBaseURL, c.nonLocalhostRequired())
	}
	for _, target := range c.HermesExecutorPoolURLs {
		addRequiredURL(issues, "RSI_HERMES_EXECUTOR_POOL_URLS", target, c.nonLocalhostRequired())
	}
	addRequiredString(issues, "RSI_DEFAULT_REPO", c.DefaultRepo)
	addRequiredString(issues, "RSI_KNOWLEDGE_BASE_URL", c.DefaultKnowledgeBaseURL)
	addRequiredList(issues, "RSI_ALLOWED_TARGET_REPOS", c.AllowedTargetRepos)
	addRequiredString(issues, "RSI_REASONING_VERBOSITY", c.DefaultReasoningVerbosity)
	if c.SlackMCPEnabled {
		addRequiredString(issues, "SLACK_BOT_TOKEN", c.SlackBotToken)
		addRequiredURL(issues, "RSI_SLACK_MCP_SERVER_URL", c.SlackMCPServerURL, false)
	}
	if c.RuntimeMode == "slack-surface" {
		addRequiredString(issues, "RSI_SLACK_APP_IDENTITY", c.SlackAppIdentity)
		if !c.SlackSocketModeEnabled {
			*issues = append(*issues, "RSI_SLACK_SOCKET_MODE_ENABLED must be true")
		}
		addRequiredString(issues, "RSI_SLACK_APP_TOKEN", c.SlackAppToken)
		addRequiredString(issues, "SLACK_BOT_TOKEN", c.SlackBotToken)
		addRequiredList(issues, "RSI_ALLOWED_SLACK_CHANNEL_IDS", c.AllowedSlackChannelIDs)
		if c.SlackMirrorEnabled {
			c.validateSlackMirror(issues)
		}
	}
}

func (c Config) validateSlackMirror(issues *[]string) {
	if !c.SlackMirrorEnabled {
		*issues = append(*issues, "RSI_SLACK_MIRROR_ENABLED must be true")
	}
	addRequiredString(issues, "SLACK_BOT_TOKEN", c.SlackBotToken)
	discovery := strings.ToLower(strings.TrimSpace(c.SlackMirrorChannelDiscovery))
	switch discovery {
	case "", "joined":
	case "explicit":
		addRequiredList(issues, "RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST", c.SlackMirrorChannelAllowlist)
	default:
		*issues = append(*issues, "RSI_SLACK_MIRROR_CHANNEL_DISCOVERY must be joined or explicit")
	}
	addRequiredURL(issues, "RSI_HONCHO_BASE_URL", c.HonchoBaseURL, c.nonLocalhostRequired())
	addRequiredString(issues, "RSI_HONCHO_WORKSPACE_ID", c.HonchoWorkspaceID)
	addRequiredString(issues, "RSI_SOURCE_MIRROR_CHECKPOINT_ROOT", c.SourceMirrorCheckpointRoot)
}

func (c Config) validateNotionMirror(issues *[]string) {
	if !c.NotionMirrorEnabled {
		*issues = append(*issues, "RSI_NOTION_MIRROR_ENABLED must be true")
	}
	addRequiredString(issues, "NOTION_TOKEN", c.NotionToken)
	addRequiredList(issues, "RSI_NOTION_MIRROR_ALLOWLIST", c.NotionMirrorAllowlist)
	addRequiredURL(issues, "RSI_NOTION_API_BASE_URL", c.NotionAPIBaseURL, false)
	addRequiredURL(issues, "RSI_HONCHO_BASE_URL", c.HonchoBaseURL, c.nonLocalhostRequired())
	addRequiredString(issues, "RSI_HONCHO_WORKSPACE_ID", c.HonchoWorkspaceID)
	addRequiredString(issues, "RSI_SOURCE_MIRROR_CHECKPOINT_ROOT", c.SourceMirrorCheckpointRoot)
	c.validateNotionMirrorCrawlerConfig(issues)
}

func (c Config) validateSourceMirrorHealth(issues *[]string) {
	if !c.SlackMirrorEnabled && !c.NotionMirrorEnabled {
		*issues = append(*issues, "at least one source mirror must be enabled for source-mirror-health")
	}
	addRequiredURL(issues, "RSI_HONCHO_BASE_URL", c.HonchoBaseURL, c.nonLocalhostRequired())
	addRequiredString(issues, "RSI_HONCHO_WORKSPACE_ID", c.HonchoWorkspaceID)
	addRequiredString(issues, "RSI_SOURCE_MIRROR_CHECKPOINT_ROOT", c.SourceMirrorCheckpointRoot)
	if c.SlackMirrorEnabled {
		addRequiredString(issues, "SLACK_BOT_TOKEN", c.SlackBotToken)
		discovery := strings.ToLower(strings.TrimSpace(c.SlackMirrorChannelDiscovery))
		switch discovery {
		case "", "joined":
		case "explicit":
			addRequiredList(issues, "RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST", c.SlackMirrorChannelAllowlist)
		default:
			*issues = append(*issues, "RSI_SLACK_MIRROR_CHANNEL_DISCOVERY must be joined or explicit")
		}
	}
	if c.NotionMirrorEnabled {
		addRequiredString(issues, "NOTION_TOKEN", c.NotionToken)
		addRequiredList(issues, "RSI_NOTION_MIRROR_ALLOWLIST", c.NotionMirrorAllowlist)
		addRequiredURL(issues, "RSI_NOTION_API_BASE_URL", c.NotionAPIBaseURL, false)
		c.validateNotionMirrorCrawlerConfig(issues)
	}
}

func (c Config) validateNotionMirrorCrawlerConfig(issues *[]string) {
	if c.NotionMirrorRequestsPerSecond < 0 {
		*issues = append(*issues, "RSI_NOTION_MIRROR_REQUESTS_PER_SECOND must be non-negative")
	}
	if c.NotionMirrorMaxRetries < 0 {
		*issues = append(*issues, "RSI_NOTION_MIRROR_MAX_RETRIES must be non-negative")
	}
	if c.NotionMirrorMaxRetries > 0 && c.NotionMirrorRetryBaseDelay < 0 {
		*issues = append(*issues, "RSI_NOTION_MIRROR_RETRY_BASE_DELAY must be non-negative")
	}
	if c.NotionMirrorMaxDatabasesPerRoot < 0 {
		*issues = append(*issues, "RSI_NOTION_MIRROR_MAX_DATABASES_PER_ROOT must be non-negative")
	}
	if c.NotionMirrorMaxBlocksPerPage < 0 {
		*issues = append(*issues, "RSI_NOTION_MIRROR_MAX_BLOCKS_PER_PAGE must be non-negative")
	}
	if c.NotionMirrorMaxDepth < 0 {
		*issues = append(*issues, "RSI_NOTION_MIRROR_MAX_DEPTH must be non-negative")
	}
	if c.NotionMirrorMaxDocumentBytes < 0 {
		*issues = append(*issues, "RSI_NOTION_MIRROR_MAX_DOCUMENT_BYTES must be non-negative")
	}
}

func (c Config) validateImprovementPlane(issues *[]string) {
	c.validateCommonPlaneConfig(issues)
	if c.RuntimeMode == "migrate" {
		return
	}
	if c.RuntimeMode == "worker" || c.RuntimeMode == "reconcile" || c.RuntimeMode == "cron" {
		addRequiredURL(issues, "RSI_RUNNER_EVAL_BASE_URL", c.EvalRunnerBaseURL, c.nonLocalhostRequired())
		addRequiredURL(issues, "RSI_RUNNER_PROPOSAL_BASE_URL", c.ProposalRunnerBaseURL, c.nonLocalhostRequired())
		if c.DefaultProposalCap <= 0 {
			*issues = append(*issues, "RSI_ACTIVE_PROPOSAL_CAP must be set to a positive integer")
		}
		if c.ProposalPromoterInterval <= 0 {
			*issues = append(*issues, "RSI_PROPOSAL_PROMOTER_INTERVAL must be set to a positive duration")
		}
	}
	addRequiredString(issues, "RSI_REASONING_VERBOSITY", c.DefaultReasoningVerbosity)
	if c.RuntimeMode == "serve" {
		addRequiredURL(issues, "RSI_HONCHO_RUNTIME_BASE_URL", c.HonchoRuntimeBaseURL, c.nonLocalhostRequired())
	}
	if c.RuntimeMode == "worker" || c.RuntimeMode == "reconcile" {
		addRequiredString(issues, "RSI_GITHUB_APP_ID", c.GitHubAppID)
		addRequiredString(issues, "RSI_GITHUB_APP_INSTALLATION_ID", c.GitHubAppInstallationID)
		addRequiredString(issues, "RSI_GITHUB_APP_PRIVATE_KEY", c.GitHubAppPrivateKey)
		addRequiredString(issues, "RSI_GITHUB_OWNER", c.GitHubOwner)
		addRequiredURL(issues, "RSI_GITHUB_API_BASE_URL", c.GitHubAPIBaseURL, false)
		addRequiredString(issues, "RSI_GITHUB_COMMIT_USER", c.GitHubCommitUser)
		addRequiredString(issues, "RSI_GITHUB_COMMIT_EMAIL", c.GitHubCommitEmail)
		addRequiredString(issues, "RSI_SANDBOX_NAMESPACE", c.SandboxNamespace)
		addRequiredString(issues, "RSI_SANDBOX_IMAGE", c.SandboxImage)
		addRequiredString(issues, "RSI_SANDBOX_SERVICE_ACCOUNT_NAME", c.SandboxServiceAccount)
		if c.SandboxJobTTLSeconds <= 0 {
			*issues = append(*issues, "RSI_SANDBOX_JOB_TTL_SECONDS must be set to a positive integer")
		}
		if c.SandboxDeadlineSeconds <= 0 {
			*issues = append(*issues, "RSI_SANDBOX_ACTIVE_DEADLINE_SECONDS must be set to a positive integer")
		}
		c.validateGitHubInstallations(issues)
	}
}

func (c Config) validateGitHubInstallations(issues *[]string) {
	defaultOwner := strings.TrimSpace(c.GitHubOwner)
	if defaultOwner == "" {
		return
	}
	for repo, owner := range c.GitHubRepoOwners {
		repo = strings.TrimSpace(repo)
		owner = strings.TrimSpace(owner)
		if repo == "" || owner == "" || strings.EqualFold(owner, defaultOwner) {
			continue
		}
		if _, ok := c.GitHubAppInstallationIDs[owner]; !ok {
			*issues = append(*issues, fmt.Sprintf("RSI_GITHUB_APP_INSTALLATION_IDS must include %s for repo %s", owner, repo))
		}
	}
}

func (c Config) validateCommonPlaneConfig(issues *[]string) {
	addRequiredString(issues, "RSI_STORE_BACKEND", c.StoreBackend)
	if strings.TrimSpace(c.StoreBackend) != "postgres" {
		*issues = append(*issues, "RSI_STORE_BACKEND must be set to postgres")
	}
	addRequiredString(issues, "RSI_POSTGRES_URL", c.PostgresURL)
	addRequiredURL(issues, "RSI_PUBLIC_BASE_URL", c.PublicBaseURL, c.nonLocalhostRequired())
}

func (c Config) nonLocalhostRequired() bool {
	env := strings.ToLower(strings.TrimSpace(c.Environment))
	return env == "stage" || env == "prod" || env == "production"
}

func addRequiredString(issues *[]string, name string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		*issues = append(*issues, name+" is required")
		return
	}
	if strings.HasPrefix(strings.ToLower(value), "vault:") {
		*issues = append(*issues, name+" must be resolved at runtime and may not start with vault:")
	}
}

func addRequiredList(issues *[]string, name string, values []string) {
	if len(values) == 0 {
		*issues = append(*issues, name+" is required")
		return
	}
	for _, value := range values {
		addRequiredString(issues, name, value)
	}
}

func addRequiredURL(issues *[]string, name string, value string, rejectLocalhost bool) {
	addRequiredString(issues, name, value)
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(strings.ToLower(value), "vault:") {
		return
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		*issues = append(*issues, fmt.Sprintf("%s must be a valid absolute URL", name))
		return
	}
	if !rejectLocalhost {
		return
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") || host == "127.0.0.1" || host == "::1" {
		*issues = append(*issues, fmt.Sprintf("%s may not point to localhost in stage/prod", name))
		return
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		*issues = append(*issues, fmt.Sprintf("%s may not point to loopback in stage/prod", name))
	}
}
