package sandbox

import (
	"context"
	"fmt"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/cluster"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Launcher interface {
	Launch(ctx context.Context, req JobRequest) (Session, *batchv1.Job, error)
	GetJob(ctx context.Context, namespace string, name string) (*batchv1.Job, error)
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
