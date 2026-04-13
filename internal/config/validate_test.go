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
		"RSI_GITHUB_TOKEN is required",
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
