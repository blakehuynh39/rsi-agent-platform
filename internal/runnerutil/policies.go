package runnerutil

import (
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
)

const (
	defaultHermesComputerRoot = "/workspace/company"
	defaultHermesRunRoot      = "/workspace/company/.rsi/runs"
	defaultHermesArtifactRoot = "/workspace/company/artifacts"
)

func WorkspacePolicyFromConfig(cfg config.Config) *clients.RunnerWorkspacePolicy {
	computerRoot := firstPolicyValue(cfg.HermesComputerRoot, defaultHermesComputerRoot)
	runRoot := firstPolicyValue(cfg.HermesRunRoot, defaultHermesRunRoot)
	artifactRoot := firstPolicyValue(cfg.HermesArtifactRoot, defaultHermesArtifactRoot)
	return clients.NewRunnerWorkspacePolicy(computerRoot, runRoot, artifactRoot)
}

func firstPolicyValue(value string, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return fallback
}
