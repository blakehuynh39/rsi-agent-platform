package sandbox

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

type JobRequest struct {
	TraceID      string            `json:"trace_id"`
	ProposalID   string            `json:"proposal_id,omitempty"`
	Repo         string            `json:"repo"`
	BaseRef      string            `json:"base_ref,omitempty"`
	Commands     []string          `json:"commands"`
	Env          map[string]string `json:"env,omitempty"`
	RequestedBy  string            `json:"requested_by,omitempty"`
	TimeoutSecs  int               `json:"timeout_seconds,omitempty"`
	ArtifactPath string            `json:"artifact_path,omitempty"`
}

type JobManifest struct {
	APIVersion string      `json:"apiVersion"`
	Kind       string      `json:"kind"`
	Metadata   Metadata    `json:"metadata"`
	Spec       JobSpec     `json:"spec"`
}

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type JobSpec struct {
	BackoffLimit            int          `json:"backoffLimit"`
	TTLSecondsAfterFinished int          `json:"ttlSecondsAfterFinished,omitempty"`
	ActiveDeadlineSeconds   int          `json:"activeDeadlineSeconds,omitempty"`
	Template                PodTemplate  `json:"template"`
}

type PodTemplate struct {
	Metadata Metadata `json:"metadata"`
	Spec     PodSpec  `json:"spec"`
}

type PodSpec struct {
	ServiceAccountName string      `json:"serviceAccountName"`
	RestartPolicy      string      `json:"restartPolicy"`
	Containers         []Container `json:"containers"`
	Volumes            []Volume    `json:"volumes,omitempty"`
}

type Container struct {
	Name         string          `json:"name"`
	Image        string          `json:"image"`
	Command      []string        `json:"command,omitempty"`
	Env          []EnvVar        `json:"env,omitempty"`
	VolumeMounts []VolumeMount   `json:"volumeMounts,omitempty"`
	Resources    map[string]any  `json:"resources,omitempty"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Volume struct {
	Name     string    `json:"name"`
	EmptyDir struct{}  `json:"emptyDir,omitempty"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9-]+`)

func BuildJob(cfg config.Config, req JobRequest) JobManifest {
	now := time.Now().UTC()
	name := buildName(req.TraceID, req.Repo)
	timeout := cfg.SandboxDeadlineSeconds
	if req.TimeoutSecs > 0 {
		timeout = req.TimeoutSecs
	}
	labels := map[string]string{
		"app.kubernetes.io/name":      "rsi-agent-platform",
		"app.kubernetes.io/component": "sandbox-runtime",
		"rsi.storyprotocol.net/repo":  sanitizeLabel(req.Repo),
	}
	if req.TraceID != "" {
		labels["rsi.storyprotocol.net/trace-id"] = sanitizeLabel(req.TraceID)
	}
	if req.ProposalID != "" {
		labels["rsi.storyprotocol.net/proposal-id"] = sanitizeLabel(req.ProposalID)
	}
	labels["rsi.storyprotocol.net/generated-at"] = now.Format("20060102-150405")

	env := []EnvVar{
		{Name: "RSI_TRACE_ID", Value: req.TraceID},
		{Name: "RSI_REPO", Value: req.Repo},
		{Name: "RSI_REQUESTED_BY", Value: req.RequestedBy},
		{Name: "RSI_ARTIFACT_PATH", Value: req.ArtifactPath},
		{Name: "RSI_BASE_REF", Value: req.BaseRef},
	}
	for key, value := range req.Env {
		if strings.TrimSpace(key) == "" {
			continue
		}
		env = append(env, EnvVar{Name: key, Value: value})
	}
	slices.SortFunc(env, func(a, b EnvVar) int {
		return strings.Compare(a.Name, b.Name)
	})

	commands := req.Commands
	if len(commands) == 0 {
		commands = []string{"bash", "-lc", "trap : TERM INT; sleep infinity & wait"}
	}

	return JobManifest{
		APIVersion: "batch/v1",
		Kind:       "Job",
		Metadata: Metadata{
			Name:      name,
			Namespace: cfg.SandboxNamespace,
			Labels:    labels,
		},
		Spec: JobSpec{
			BackoffLimit:            0,
			TTLSecondsAfterFinished: cfg.SandboxJobTTLSeconds,
			ActiveDeadlineSeconds:   timeout,
			Template: PodTemplate{
				Metadata: Metadata{Labels: labels},
				Spec: PodSpec{
					ServiceAccountName: cfg.SandboxServiceAccount,
					RestartPolicy:      "Never",
					Containers: []Container{
						{
							Name:    "sandbox-runtime",
							Image:   cfg.SandboxImage,
							Command: commands,
							Env:     env,
							VolumeMounts: []VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
							},
						},
					},
					Volumes: []Volume{
						{Name: "workspace"},
					},
				},
			},
		},
	}
}

func buildName(traceID string, repo string) string {
	base := sanitizeLabel(traceID)
	if base == "" {
		base = sanitizeLabel(repo)
	}
	if base == "" {
		base = "job"
	}
	return fmt.Sprintf("rsi-sandbox-%s", base)
}

func sanitizeLabel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, ".", "-")
	value = nonAlnum.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if len(value) > 48 {
		value = strings.Trim(value[:48], "-")
	}
	return value
}
