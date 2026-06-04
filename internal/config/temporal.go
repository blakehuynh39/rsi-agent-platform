package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type TemporalTarget struct {
	Environment               string   `json:"environment"`
	Name                      string   `json:"name"`
	HostPort                  string   `json:"host_port"`
	Namespace                 string   `json:"namespace"`
	CertPath                  string   `json:"cert_path,omitempty"`
	KeyPath                   string   `json:"key_path,omitempty"`
	CertPEMEnv                string   `json:"cert_pem_env,omitempty"`
	KeyPEMEnv                 string   `json:"key_pem_env,omitempty"`
	AllowedWorkflowTypes      []string `json:"allowed_workflow_types,omitempty"`
	AllowedTaskQueues         []string `json:"allowed_task_queues,omitempty"`
	AllowedScheduleIDs        []string `json:"allowed_schedule_ids,omitempty"`
	AllowedSchedulePrefixes   []string `json:"allowed_schedule_prefixes,omitempty"`
	AllowedWorkflowIDs        []string `json:"allowed_workflow_ids,omitempty"`
	AllowedWorkflowIDPrefixes []string `json:"allowed_workflow_id_prefixes,omitempty"`
}

type temporalTargetsEnvelope struct {
	Targets []TemporalTarget `json:"targets"`
}

func temporalTargetsEnv(name string) []TemporalTarget {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil
	}
	var envelope temporalTargetsEnvelope
	if err := json.Unmarshal([]byte(raw), &envelope); err == nil && len(envelope.Targets) > 0 {
		return normalizeTemporalTargets(envelope.Targets)
	}
	var targets []TemporalTarget
	if err := json.Unmarshal([]byte(raw), &targets); err == nil {
		return normalizeTemporalTargets(targets)
	}
	return nil
}

func normalizeTemporalTargets(targets []TemporalTarget) []TemporalTarget {
	out := make([]TemporalTarget, 0, len(targets))
	for _, target := range targets {
		target.Environment = normalizeTemporalEnvironment(target.Environment)
		target.Name = strings.ToLower(strings.TrimSpace(target.Name))
		target.HostPort = strings.TrimSpace(target.HostPort)
		target.Namespace = strings.TrimSpace(target.Namespace)
		target.CertPath = strings.TrimSpace(target.CertPath)
		target.KeyPath = strings.TrimSpace(target.KeyPath)
		target.CertPEMEnv = strings.TrimSpace(target.CertPEMEnv)
		target.KeyPEMEnv = strings.TrimSpace(target.KeyPEMEnv)
		target.AllowedWorkflowTypes = CompactUniqueStrings(target.AllowedWorkflowTypes)
		target.AllowedTaskQueues = CompactUniqueStrings(target.AllowedTaskQueues)
		target.AllowedScheduleIDs = CompactUniqueStrings(target.AllowedScheduleIDs)
		target.AllowedSchedulePrefixes = CompactUniqueStrings(target.AllowedSchedulePrefixes)
		target.AllowedWorkflowIDs = CompactUniqueStrings(target.AllowedWorkflowIDs)
		target.AllowedWorkflowIDPrefixes = CompactUniqueStrings(target.AllowedWorkflowIDPrefixes)
		out = append(out, target)
	}
	return out
}

func normalizeTemporalEnvironment(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "stg", "stage", "staging":
		return "stage"
	case "prod", "production":
		return "prod"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func (c Config) validateTemporalTargets(issues *[]string) {
	if !c.TemporalControlEnabled {
		return
	}
	if len(c.TemporalTargets) == 0 {
		*issues = append(*issues, "RSI_TEMPORAL_TARGETS_JSON must include at least one target when RSI_TEMPORAL_CONTROL_ENABLED is true")
		return
	}
	seen := map[string]struct{}{}
	for index, target := range c.TemporalTargets {
		prefix := fmt.Sprintf("RSI_TEMPORAL_TARGETS_JSON target %d", index)
		if target.Environment == "" {
			*issues = append(*issues, prefix+" environment is required")
		}
		if target.Name == "" {
			*issues = append(*issues, prefix+" name is required")
		}
		if target.HostPort == "" {
			*issues = append(*issues, prefix+" host_port is required")
		}
		if target.Namespace == "" {
			*issues = append(*issues, prefix+" namespace is required")
		}
		if (target.CertPath == "") != (target.KeyPath == "") {
			*issues = append(*issues, prefix+" cert_path and key_path must be set together")
		}
		if (target.CertPEMEnv == "") != (target.KeyPEMEnv == "") {
			*issues = append(*issues, prefix+" cert_pem_env and key_pem_env must be set together")
		}
		if target.CertPath == "" && target.CertPEMEnv == "" {
			*issues = append(*issues, prefix+" must set cert_path/key_path or cert_pem_env/key_pem_env")
		}
		key := target.Environment + "/" + target.Name
		if _, ok := seen[key]; ok && key != "/" {
			*issues = append(*issues, prefix+" duplicates target "+key)
		}
		seen[key] = struct{}{}
	}
}
