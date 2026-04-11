package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName              string
	Environment              string
	HTTPPort                 int
	StoreBackend             string
	PostgresURL              string
	RedisAddr                string
	S3Bucket                 string
	PublicBaseURL            string
	WorkflowQueueURL         string
	ProactiveQueueURL        string
	EvalQueueURL             string
	ProposalQueueURL         string
	SandboxQueueURL          string
	SlackAppIdentity         string
	SlackSocketModeEnabled   bool
	SlackAppToken            string
	SlackBotToken            string
	SandboxNamespace         string
	SandboxImage             string
	SandboxServiceAccount    string
	SandboxJobTTLSeconds     int
	SandboxDeadlineSeconds   int
	AllowedSlackChannelIDs   []string
	DefaultOperatorDomain    string
	DefaultRepo              string
	DefaultKnowledgeBaseURL  string
	ProposalPromoterInterval time.Duration
}

func Load(serviceName string) Config {
	environment := stringEnv("RSI_ENV", "development")
	return Config{
		ServiceName:              stringEnv("RSI_SERVICE_NAME", serviceName),
		Environment:              environment,
		HTTPPort:                 intEnv("RSI_HTTP_PORT", 8080),
		StoreBackend:             stringEnv("RSI_STORE_BACKEND", defaultStoreBackend(environment)),
		PostgresURL:              stringEnv("RSI_POSTGRES_URL", "postgres://localhost:5432/rsi_agent_platform"),
		RedisAddr:                stringEnv("RSI_REDIS_ADDR", "redis://redis.redis.svc.cluster.local:6379"),
		S3Bucket:                 stringEnv("RSI_S3_BUCKET", "rsi-agent-platform-stage-artifacts"),
		PublicBaseURL:            stringEnv("RSI_PUBLIC_BASE_URL", "http://localhost:8080"),
		WorkflowQueueURL:         stringEnv("RSI_WORKFLOW_QUEUE_URL", "memory://workflow"),
		ProactiveQueueURL:        stringEnv("RSI_PROACTIVE_QUEUE_URL", "memory://proactive"),
		EvalQueueURL:             stringEnv("RSI_EVAL_QUEUE_URL", "memory://eval"),
		ProposalQueueURL:         stringEnv("RSI_PROPOSAL_QUEUE_URL", "memory://proposal"),
		SandboxQueueURL:          stringEnv("RSI_SANDBOX_QUEUE_URL", "memory://sandbox"),
		SlackAppIdentity:         stringEnv("RSI_SLACK_APP_IDENTITY", ""),
		SlackSocketModeEnabled:   boolEnv("RSI_SLACK_SOCKET_MODE_ENABLED", false),
		SlackAppToken:            stringEnv("RSI_SLACK_APP_TOKEN", ""),
		SlackBotToken:            stringEnv("RSI_SLACK_BOT_TOKEN", ""),
		SandboxNamespace:         stringEnv("RSI_SANDBOX_NAMESPACE", "rsi-platform"),
		SandboxImage:             stringEnv("RSI_SANDBOX_IMAGE", "rsi-agent-platform-sandbox:latest"),
		SandboxServiceAccount:    stringEnv("RSI_SANDBOX_SERVICE_ACCOUNT_NAME", "rsi-sandbox"),
		SandboxJobTTLSeconds:     intEnv("RSI_SANDBOX_JOB_TTL_SECONDS", 3600),
		SandboxDeadlineSeconds:   intEnv("RSI_SANDBOX_ACTIVE_DEADLINE_SECONDS", 1800),
		AllowedSlackChannelIDs:   listEnv("RSI_ALLOWED_SLACK_CHANNEL_IDS", []string{"C_AGENT_FACTORY"}),
		DefaultOperatorDomain:    stringEnv("RSI_OPERATOR_EMAIL_DOMAIN", "piplabs.xyz"),
		DefaultRepo:              stringEnv("RSI_DEFAULT_REPO", "depin-backend"),
		DefaultKnowledgeBaseURL:  stringEnv("RSI_KNOWLEDGE_BASE_URL", "https://staging-depin.storyprotocol.net/openapi.json"),
		ProposalPromoterInterval: durationEnv("RSI_PROPOSAL_PROMOTER_INTERVAL", 15*time.Minute),
	}
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

func defaultStoreBackend(environment string) string {
	if strings.EqualFold(environment, "production") {
		return "postgres"
	}
	return "memory"
}
