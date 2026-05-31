package sandbox

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/cluster"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Launcher interface {
	Launch(ctx context.Context, req JobRequest) (Session, *batchv1.Job, error)
	GetJob(ctx context.Context, namespace string, name string) (*batchv1.Job, error)
	ObserveJob(ctx context.Context, namespace string, name string) (JobObservation, error)
	ResolvePod(ctx context.Context, namespace string, jobName string) (string, error)
	Exec(ctx context.Context, namespace string, podName string, command []string) (ExecResult, error)
}

type KubernetesLauncher struct {
	cfg       config.Config
	clientset kubernetes.Interface
	restCfg   *rest.Config
}

func NewLauncher(cfg config.Config) (*KubernetesLauncher, error) {
	restCfg, err := cluster.NewConfig(cfg)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}
	return &KubernetesLauncher{cfg: cfg, clientset: clientset, restCfg: restCfg}, nil
}

func NewLauncherWithClient(cfg config.Config, clientset kubernetes.Interface) *KubernetesLauncher {
	return &KubernetesLauncher{cfg: cfg, clientset: clientset}
}

func (l *KubernetesLauncher) Launch(ctx context.Context, req JobRequest) (Session, *batchv1.Job, error) {
	job := BuildBatchJob(l.cfg, req)
	created, err := l.clientset.BatchV1().Jobs(job.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return Session{}, nil, fmt.Errorf("create sandbox job: %w", err)
	}
	now := time.Now().UTC()
	session := Session{
		ID:         created.Name,
		TraceID:    req.TraceID,
		ProposalID: req.ProposalID,
		PodName:    created.Name,
		Namespace:  created.Namespace,
		Status:     "running",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return session, created, nil
}

func (l *KubernetesLauncher) GetJob(ctx context.Context, namespace string, name string) (*batchv1.Job, error) {
	job, err := l.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get sandbox job: %w", err)
	}
	return job, nil
}

func (l *KubernetesLauncher) ObserveJob(ctx context.Context, namespace string, name string) (JobObservation, error) {
	job, err := l.GetJob(ctx, namespace, name)
	if err != nil {
		return JobObservation{}, err
	}
	observation := JobObservation{
		Namespace:    namespace,
		JobName:      name,
		JobSucceeded: job.Status.Succeeded > 0,
		JobFailed:    job.Status.Failed > 0,
	}
	for _, condition := range job.Status.Conditions {
		observation.JobConditions = append(observation.JobConditions, fmt.Sprintf("%s=%s:%s", condition.Type, condition.Status, condition.Reason))
	}
	pods, err := l.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", name),
	})
	if err != nil {
		return JobObservation{}, fmt.Errorf("list sandbox pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return observation, nil
	}
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].CreationTimestamp.Time.After(pods.Items[j].CreationTimestamp.Time)
	})
	pod := pods.Items[0]
	observation.PodName = pod.Name
	observation.PodPhase = string(pod.Status.Phase)
	if terminated := terminatedContainerState(pod); terminated != nil {
		exitCode := terminated.ExitCode
		observation.ContainerExitCode = &exitCode
		observation.TerminationReason = terminated.Reason
		observation.TerminationMessage = terminated.Message
	}
	if observation.JobSucceeded || observation.JobFailed || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		stream, err := l.clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: "sandbox-runtime"}).Stream(ctx)
		if err == nil {
			defer stream.Close()
			if body, readErr := io.ReadAll(stream); readErr == nil {
				observation.Logs = string(body)
			}
		}
	}
	return observation, nil
}

func (l *KubernetesLauncher) ResolvePod(ctx context.Context, namespace string, jobName string) (string, error) {
	pods, err := l.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return "", fmt.Errorf("list sandbox pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no sandbox pod found for job %s", jobName)
	}
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].CreationTimestamp.Time.After(pods.Items[j].CreationTimestamp.Time)
	})
	return pods.Items[0].Name, nil
}

func (l *KubernetesLauncher) Exec(ctx context.Context, namespace string, podName string, command []string) (ExecResult, error) {
	if l.restCfg == nil {
		return ExecResult{}, fmt.Errorf("sandbox exec unavailable: rest config not configured")
	}
	if len(command) == 0 {
		return ExecResult{}, fmt.Errorf("sandbox exec requires a command")
	}
	req := l.clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Container: "sandbox-runtime",
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)
	execURL, err := url.Parse(req.URL().String())
	if err != nil {
		return ExecResult{}, fmt.Errorf("sandbox exec url: %w", err)
	}
	executor, err := remotecommand.NewSPDYExecutor(l.restCfg, "POST", execURL)
	if err != nil {
		return ExecResult{}, fmt.Errorf("sandbox exec setup: %w", err)
	}
	var stdout, stderr strings.Builder
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	return ExecResult{Stdout: stdout.String(), Stderr: stderr.String()}, err
}

func terminatedContainerState(pod corev1.Pod) *corev1.ContainerStateTerminated {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Terminated != nil {
			return status.State.Terminated
		}
	}
	return nil
}
