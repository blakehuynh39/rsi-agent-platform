package sandbox

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       JobSpec  `json:"spec"`
}

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type JobSpec struct {
	BackoffLimit            int         `json:"backoffLimit"`
	TTLSecondsAfterFinished int         `json:"ttlSecondsAfterFinished,omitempty"`
	ActiveDeadlineSeconds   int         `json:"activeDeadlineSeconds,omitempty"`
	Template                PodTemplate `json:"template"`
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
	Name         string         `json:"name"`
	Image        string         `json:"image"`
	Command      []string       `json:"command,omitempty"`
	Env          []EnvVar       `json:"env,omitempty"`
	VolumeMounts []VolumeMount  `json:"volumeMounts,omitempty"`
	Resources    map[string]any `json:"resources,omitempty"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Volume struct {
	Name     string   `json:"name"`
	EmptyDir struct{} `json:"emptyDir,omitempty"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

type Session struct {
	ID              string     `json:"id"`
	TraceID         string     `json:"trace_id"`
	ProposalID      string     `json:"proposal_id,omitempty"`
	RepoChangeJobID string     `json:"repo_change_job_id,omitempty"`
	PodName         string     `json:"pod_name"`
	Namespace       string     `json:"namespace"`
	Status          string     `json:"status"`
	BranchName      string     `json:"branch_name,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

type JobObservation struct {
	Namespace          string   `json:"namespace"`
	JobName            string   `json:"job_name"`
	PodName            string   `json:"pod_name,omitempty"`
	JobSucceeded       bool     `json:"job_succeeded"`
	JobFailed          bool     `json:"job_failed"`
	JobConditions      []string `json:"job_conditions,omitempty"`
	PodPhase           string   `json:"pod_phase,omitempty"`
	ContainerExitCode  *int32   `json:"container_exit_code,omitempty"`
	TerminationReason  string   `json:"termination_reason,omitempty"`
	TerminationMessage string   `json:"termination_message,omitempty"`
	Logs               string   `json:"logs,omitempty"`
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

	envMap := map[string]string{
		"RSI_TRACE_ID":      req.TraceID,
		"RSI_REPO":          req.Repo,
		"RSI_REQUESTED_BY":  req.RequestedBy,
		"RSI_ARTIFACT_PATH": req.ArtifactPath,
		"RSI_BASE_REF":      req.BaseRef,
	}
	for key, value := range req.Env {
		if strings.TrimSpace(key) == "" {
			continue
		}
		envMap[key] = value
	}
	env := make([]EnvVar, 0, len(envMap))
	for key, value := range envMap {
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

func BuildBatchJob(cfg config.Config, req JobRequest) *batchv1.Job {
	manifest := BuildJob(cfg, req)
	envVars := make([]corev1.EnvVar, 0, len(manifest.Spec.Template.Spec.Containers[0].Env))
	for _, item := range manifest.Spec.Template.Spec.Containers[0].Env {
		envVars = append(envVars, corev1.EnvVar{Name: item.Name, Value: item.Value})
	}
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifest.APIVersion,
			Kind:       manifest.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      manifest.Metadata.Name,
			Namespace: manifest.Metadata.Namespace,
			Labels:    manifest.Metadata.Labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(int32(manifest.Spec.BackoffLimit)),
			TTLSecondsAfterFinished: int32Ptr(int32(manifest.Spec.TTLSecondsAfterFinished)),
			ActiveDeadlineSeconds:   int64Ptr(int64(manifest.Spec.ActiveDeadlineSeconds)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: manifest.Spec.Template.Metadata.Labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: manifest.Spec.Template.Spec.ServiceAccountName,
					RestartPolicy:      corev1.RestartPolicy(manifest.Spec.Template.Spec.RestartPolicy),
					Containers: []corev1.Container{
						{
							Name:         manifest.Spec.Template.Spec.Containers[0].Name,
							Image:        manifest.Spec.Template.Spec.Containers[0].Image,
							Command:      manifest.Spec.Template.Spec.Containers[0].Command,
							Env:          envVars,
							VolumeMounts: []corev1.VolumeMount{{Name: "workspace", MountPath: "/workspace"}},
						},
					},
					Volumes: []corev1.Volume{{Name: "workspace", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
				},
			},
		},
	}
}

func int32Ptr(value int32) *int32 {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
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
