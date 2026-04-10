package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServiceName             string
	Environment             string
	HTTPPort                int
	PostgresURL             string
	RedisAddr               string
	S3Bucket                string
	PublicBaseURL           string
	WorkflowQueueURL        string
	ProactiveQueueURL       string
	EvalQueueURL            string
	ProposalQueueURL        string
	SandboxQueueURL         string
	AllowedSlackChannelIDs  []string
	DefaultOperatorDomain   string
	DefaultRepo             string
	DefaultKnowledgeBaseURL string
}

func Load(serviceName string) Config {
	return Config{
		ServiceName:             stringEnv("RSI_SERVICE_NAME", serviceName),
		Environment:             stringEnv("RSI_ENV", "development"),
		HTTPPort:                intEnv("RSI_HTTP_PORT", 8080),
		PostgresURL:             stringEnv("RSI_POSTGRES_URL", "postgres://localhost:5432/rsi_agent_platform"),
		RedisAddr:               stringEnv("RSI_REDIS_ADDR", "redis://redis.redis.svc.cluster.local:6379"),
		S3Bucket:                stringEnv("RSI_S3_BUCKET", "rsi-agent-platform-stage-artifacts"),
		PublicBaseURL:           stringEnv("RSI_PUBLIC_BASE_URL", "http://localhost:8080"),
		WorkflowQueueURL:        stringEnv("RSI_WORKFLOW_QUEUE_URL", "memory://workflow"),
		ProactiveQueueURL:       stringEnv("RSI_PROACTIVE_QUEUE_URL", "memory://proactive"),
		EvalQueueURL:            stringEnv("RSI_EVAL_QUEUE_URL", "memory://eval"),
		ProposalQueueURL:        stringEnv("RSI_PROPOSAL_QUEUE_URL", "memory://proposal"),
		SandboxQueueURL:         stringEnv("RSI_SANDBOX_QUEUE_URL", "memory://sandbox"),
		AllowedSlackChannelIDs:  listEnv("RSI_ALLOWED_SLACK_CHANNEL_IDS", []string{"C_AGENT_FACTORY"}),
		DefaultOperatorDomain:   stringEnv("RSI_OPERATOR_EMAIL_DOMAIN", "piplabs.xyz"),
		DefaultRepo:             stringEnv("RSI_DEFAULT_REPO", "depin-backend"),
		DefaultKnowledgeBaseURL: stringEnv("RSI_KNOWLEDGE_BASE_URL", "https://staging-depin.storyprotocol.net/openapi.json"),
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

