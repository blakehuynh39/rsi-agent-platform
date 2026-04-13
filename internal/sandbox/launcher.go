package sandbox

import (
	"context"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/cluster"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Launcher interface {
	Launch(ctx context.Context, req JobRequest) (Session, *batchv1.Job, error)
	GetJob(ctx context.Context, namespace string, name string) (*batchv1.Job, error)
	ObserveJob(ctx context.Context, namespace string, name string) (JobObservation, error)
}

type KubernetesLauncher struct {
	cfg       config.Config
	clientset kubernetes.Interface
}

func NewLauncher(cfg config.Config) (*KubernetesLauncher, error) {
	clientset, err := cluster.NewClientset(cfg)
	if err != nil {
		return nil, err
	}
	return &KubernetesLauncher{cfg: cfg, clientset: clientset}, nil
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

func terminatedContainerState(pod corev1.Pod) *corev1.ContainerStateTerminated {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Terminated != nil {
			return status.State.Terminated
		}
	}
	return nil
}
