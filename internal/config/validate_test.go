package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDoesNotInferStoreBackend(t *testing.T) {
	t.Setenv("RSI_ENV", "stage")
	t.Setenv("RSI_HTTP_PORT", "8080")
	t.Setenv("RSI_POSTGRES_URL", "postgres://user:pass@db.example/rsi")

	cfg := Load("control-plane")
	if cfg.StoreBackend != "" {
		t.Fatalf("expected empty store backend when RSI_STORE_BACKEND is unset, got %q", cfg.StoreBackend)
	}
}

func TestControlPlaneValidationRejectsLocalhostAndVaultRuntimeValues(t *testing.T) {
	cfg := validControlPlaneConfig()
	cfg.PublicBaseURL = "http://localhost:8080"
	cfg.ToolGatewayBaseURL = "vault:secret/data/use1-stage/rsi-agent-platform#TOOL_GATEWAY"

	_, err := cfg.ValidatedFor("control-plane", "serve")
	if err == nil {
		t.Fatal("expected validation error")
	}
	message := err.Error()
	if !strings.Contains(message, "RSI_PUBLIC_BASE_URL may not point to localhost in stage/prod") {
		t.Fatalf("expected localhost validation message, got %s", message)
	}
	if !strings.Contains(message, "RSI_TOOL_GATEWAY_BASE_URL must be resolved at runtime and may not start with vault:") {
		t.Fatalf("expected vault validation message, got %s", message)
	}
}

func TestSlackSurfaceValidationRequiresSlackContract(t *testing.T) {
	cfg := validControlPlaneConfig()

	_, err := cfg.ValidatedFor("control-plane", "slack-surface")
	if err == nil {
		t.Fatal("expected slack-surface validation error")
	}
	message := err.Error()
	for _, required := range []string{
		"RSI_SLACK_APP_IDENTITY is required",
		"RSI_SLACK_SOCKET_MODE_ENABLED must be true",
		"RSI_SLACK_APP_TOKEN is required",
		"RSI_SLACK_BOT_TOKEN is required",
		"RSI_ALLOWED_SLACK_CHANNEL_IDS is required",
	} {
		if !strings.Contains(message, required) {
			t.Fatalf("expected %q in validation message, got %s", required, message)
		}
	}
}

func TestImprovementPlaneValidationRequiresExplicitPromoterInterval(t *testing.T) {
	cfg := Config{
		ServiceName:               "improvement-plane",
		ServiceKind:               "improvement-plane",
		Environment:               "stage",
		HTTPPort:                  8080,
		StoreBackend:              "postgres",
		PostgresURL:               "postgres://user:pass@db.example/rsi",
		PublicBaseURL:             "https://staging-rsi-platform.storyprotocol.net",
		ToolGatewayBaseURL:        "http://use1-stage-rsi-agent-platform-tool-gateway:8080",
		EvalRunnerBaseURL:         "http://use1-stage-rsi-agent-platform-runner-eval:8090",
		ProposalRunnerBaseURL:     "http://use1-stage-rsi-agent-platform-runner-proposal:8090",
		DefaultProposalCap:        2,
		DefaultReasoningVerbosity: "verbose",
		ProposalPromoterInterval:  0,
	}

	_, err := cfg.ValidatedFor("improvement-plane", "serve")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "RSI_PROPOSAL_PROMOTER_INTERVAL must be set to a positive duration") {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestImprovementPlaneWorkerValidationRequiresSandboxAndGitIdentity(t *testing.T) {
	cfg := Config{
		ServiceName:               "improvement-plane",
		ServiceKind:               "improvement-plane",
		Environment:               "stage",
		HTTPPort:                  8080,
		StoreBackend:              "postgres",
		PostgresURL:               "postgres://user:pass@db.example/rsi",
		PublicBaseURL:             "https://staging-rsi-platform.storyprotocol.net",
		ToolGatewayBaseURL:        "http://use1-stage-rsi-agent-platform-tool-gateway:8080",
		EvalRunnerBaseURL:         "http://use1-stage-rsi-agent-platform-runner-eval:8090",
		ProposalRunnerBaseURL:     "http://use1-stage-rsi-agent-platform-runner-proposal:8090",
		DefaultProposalCap:        2,
		DefaultReasoningVerbosity: "verbose",
		ProposalPromoterInterval:  15 * time.Minute,
	}

	_, err := cfg.ValidatedFor("improvement-plane", "worker")
	if err == nil {
		t.Fatal("expected worker validation error")
	}
	message := err.Error()
	for _, required := range []string{
		"RSI_GITHUB_API_BASE_URL is required",
		"RSI_GITHUB_APP_ID is required",
		"RSI_GITHUB_APP_INSTALLATION_ID is required",
		"RSI_GITHUB_APP_PRIVATE_KEY is required",
		"RSI_GITHUB_OWNER is required",
		"RSI_GITHUB_COMMIT_USER is required",
		"RSI_GITHUB_COMMIT_EMAIL is required",
		"RSI_SANDBOX_NAMESPACE is required",
		"RSI_SANDBOX_IMAGE is required",
		"RSI_SANDBOX_SERVICE_ACCOUNT_NAME is required",
		"RSI_SANDBOX_JOB_TTL_SECONDS must be set to a positive integer",
		"RSI_SANDBOX_ACTIVE_DEADLINE_SECONDS must be set to a positive integer",
	} {
		if !strings.Contains(message, required) {
			t.Fatalf("expected %q in validation message, got %s", required, message)
		}
	}
}

func TestImprovementPlaneWorkerValidationAcceptsGitHubAppContract(t *testing.T) {
	cfg := Config{
		ServiceName:               "improvement-plane",
		ServiceKind:               "improvement-plane",
		Environment:               "stage",
		HTTPPort:                  8080,
		StoreBackend:              "postgres",
		PostgresURL:               "postgres://user:pass@db.example/rsi",
		PublicBaseURL:             "https://staging-rsi-platform.storyprotocol.net",
		ToolGatewayBaseURL:        "http://use1-stage-rsi-agent-platform-tool-gateway:8080",
		EvalRunnerBaseURL:         "http://use1-stage-rsi-agent-platform-runner-eval:8090",
		ProposalRunnerBaseURL:     "http://use1-stage-rsi-agent-platform-runner-proposal:8090",
		DefaultProposalCap:        2,
		DefaultReasoningVerbosity: "verbose",
		ProposalPromoterInterval:  15 * time.Minute,
		GitHubAPIBaseURL:          "https://api.github.com",
		GitHubAppID:               "123",
		GitHubAppInstallationID:   "456",
		GitHubAppInstallationIDs:  map[string]string{"storyprotocol": "789"},
		GitHubAppPrivateKey:       "-----BEGIN RSA PRIVATE KEY-----\nkey\n-----END RSA PRIVATE KEY-----",
		GitHubOwner:               "piplabs",
		GitHubRepoOwners:          map[string]string{"story-api": "storyprotocol"},
		GitHubCommitUser:          "rsi-agent-platform-bot",
		GitHubCommitEmail:         "rsi-agent-platform-bot@users.noreply.github.com",
		SandboxNamespace:          "rsi-platform",
		SandboxImage:              "sandbox-image",
		SandboxServiceAccount:     "rsi-sandbox",
		SandboxJobTTLSeconds:      3600,
		SandboxDeadlineSeconds:    1800,
	}

	if _, err := cfg.ValidatedFor("improvement-plane", "worker"); err != nil {
		t.Fatalf("expected github app worker config to validate, got %v", err)
	}
}

func TestToolGatewayValidationRequiresGitHubApp(t *testing.T) {
	cfg := Config{
		ServiceName:               "tool-gateway",
		ServiceKind:               "tool-gateway",
		Environment:               "stage",
		HTTPPort:                  8080,
		StoreBackend:              "postgres",
		PostgresURL:               "postgres://user:pass@db.example/rsi",
		PublicBaseURL:             "https://staging-rsi-platform.storyprotocol.net",
		SlackBotToken:             "xoxb-token",
		GitHubOwner:               "piplabs",
		GitHubAPIBaseURL:          "https://api.github.com",
		SentryAuthToken:           "sentry-token",
		SentryOrganization:        "story-protocol",
		SentryAPIBaseURL:          "https://sentry.io/api/0",
		DefaultKnowledgeBaseURL:   "https://staging-depin.storyprotocol.net/openapi.json",
		DefaultReasoningVerbosity: "verbose",
	}

	_, err := cfg.ValidatedFor("tool-gateway", "serve")
	if err == nil {
		t.Fatal("expected tool-gateway validation error")
	}
	for _, required := range []string{
		"RSI_GITHUB_APP_ID is required",
		"RSI_GITHUB_APP_INSTALLATION_ID is required",
		"RSI_GITHUB_APP_PRIVATE_KEY is required",
	} {
		if !strings.Contains(err.Error(), required) {
			t.Fatalf("unexpected validation error: %v", err)
		}
	}
}

func TestToolGatewayValidationAcceptsGitHubAppContract(t *testing.T) {
	cfg := Config{
		ServiceName:               "tool-gateway",
		ServiceKind:               "tool-gateway",
		Environment:               "stage",
		HTTPPort:                  8080,
		StoreBackend:              "postgres",
		PostgresURL:               "postgres://user:pass@db.example/rsi",
		PublicBaseURL:             "https://staging-rsi-platform.storyprotocol.net",
		SlackBotToken:             "xoxb-token",
		GitHubAppID:               "123",
		GitHubAppInstallationID:   "456",
		GitHubAppInstallationIDs:  map[string]string{"storyprotocol": "789"},
		GitHubAppPrivateKey:       "-----BEGIN RSA PRIVATE KEY-----\nkey\n-----END RSA PRIVATE KEY-----",
		GitHubOwner:               "piplabs",
		GitHubRepoOwners:          map[string]string{"story-api": "storyprotocol"},
		GitHubAPIBaseURL:          "https://api.github.com",
		SentryAuthToken:           "sentry-token",
		SentryOrganization:        "story-protocol",
		SentryAPIBaseURL:          "https://sentry.io/api/0",
		DefaultKnowledgeBaseURL:   "https://staging-depin.storyprotocol.net/openapi.json",
		DefaultReasoningVerbosity: "verbose",
	}

	if _, err := cfg.ValidatedFor("tool-gateway", "serve"); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidationRequiresInstallationForMappedRepoOwner(t *testing.T) {
	cfg := Config{
		ServiceName:               "tool-gateway",
		ServiceKind:               "tool-gateway",
		Environment:               "stage",
		HTTPPort:                  8080,
		StoreBackend:              "postgres",
		PostgresURL:               "postgres://user:pass@db.example/rsi",
		PublicBaseURL:             "https://staging-rsi-platform.storyprotocol.net",
		SlackBotToken:             "xoxb-token",
		GitHubAppID:               "123",
		GitHubAppInstallationID:   "456",
		GitHubAppPrivateKey:       "-----BEGIN RSA PRIVATE KEY-----\nkey\n-----END RSA PRIVATE KEY-----",
		GitHubOwner:               "piplabs",
		GitHubRepoOwners:          map[string]string{"story-api": "storyprotocol"},
		GitHubAPIBaseURL:          "https://api.github.com",
		SentryAuthToken:           "sentry-token",
		SentryOrganization:        "story-protocol",
		SentryAPIBaseURL:          "https://sentry.io/api/0",
		DefaultKnowledgeBaseURL:   "https://staging-depin.storyprotocol.net/openapi.json",
		DefaultReasoningVerbosity: "verbose",
	}

	_, err := cfg.ValidatedFor("tool-gateway", "serve")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "RSI_GITHUB_APP_INSTALLATION_IDS must include storyprotocol for repo story-api") {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestGitHubRepoResolutionUsesOverrides(t *testing.T) {
	cfg := Config{
		GitHubOwner:             "piplabs",
		GitHubAppInstallationID: "456",
		GitHubAppInstallationIDs: map[string]string{
			"storyprotocol": "789",
		},
		GitHubRepoOwners: map[string]string{
			"story-api": "storyprotocol",
		},
	}

	if owner := cfg.GitHubRepoOwner("story-api"); owner != "storyprotocol" {
		t.Fatalf("GitHubRepoOwner() = %q, want storyprotocol", owner)
	}
	if name := cfg.GitHubRepoName("storyprotocol/story-api"); name != "story-api" {
		t.Fatalf("GitHubRepoName() = %q, want story-api", name)
	}
	if id := cfg.GitHubInstallationIDForRepo("story-api"); id != "789" {
		t.Fatalf("GitHubInstallationIDForRepo() = %q, want 789", id)
	}
	if id := cfg.GitHubInstallationIDForRepo("rsi-agent-platform"); id != "456" {
		t.Fatalf("GitHubInstallationIDForRepo() default = %q, want 456", id)
	}
}

func TestImprovementPlaneMigrateModeOnlyRequiresSharedDatabaseContract(t *testing.T) {
	cfg := Config{
		ServiceName:   "improvement-plane",
		ServiceKind:   "improvement-plane",
		Environment:   "stage",
		HTTPPort:      8080,
		StoreBackend:  "postgres",
		PostgresURL:   "postgres://user:pass@db.example/rsi",
		PublicBaseURL: "https://staging-rsi-platform.storyprotocol.net",
	}

	if _, err := cfg.ValidatedFor("improvement-plane", "migrate"); err != nil {
		t.Fatalf("expected migrate mode to validate with shared database contract, got %v", err)
	}
}

func validControlPlaneConfig() Config {
	return Config{
		ServiceName:               "control-plane",
		ServiceKind:               "control-plane",
		Environment:               "stage",
		HTTPPort:                  8080,
		StoreBackend:              "postgres",
		PostgresURL:               "postgres://user:pass@db.example/rsi",
		PublicBaseURL:             "https://staging-rsi-platform.storyprotocol.net",
		ToolGatewayBaseURL:        "http://use1-stage-rsi-agent-platform-tool-gateway:8080",
		ProdRunnerBaseURL:         "http://use1-stage-rsi-agent-platform-runner-prod:8090",
		ProactiveRunnerBaseURL:    "http://use1-stage-rsi-agent-platform-runner-proactive:8090",
		DefaultRepo:               "depin-backend",
		DefaultKnowledgeBaseURL:   "https://staging-depin.storyprotocol.net/openapi.json",
		AllowedTargetRepos:        []string{"depin-backend", "rsi-agent-platform"},
		DefaultReasoningVerbosity: "verbose",
		ProposalPromoterInterval:  15 * time.Minute,
	}
}
