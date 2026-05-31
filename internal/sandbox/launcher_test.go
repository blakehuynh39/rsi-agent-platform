package sandbox

import (
	"context"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLauncherCreatesKubernetesJob(t *testing.T) {
	cfg := config.Config{
		SandboxNamespace:       "rsi-platform",
		SandboxImage:           "example.com/rsi-sandbox:latest",
		SandboxServiceAccount:  "rsi-sandbox",
		SandboxJobTTLSeconds:   3600,
		SandboxDeadlineSeconds: 900,
	}
	launcher := NewLauncherWithClient(cfg, fake.NewSimpleClientset())

	session, job, err := launcher.Launch(context.Background(), JobRequest{
		TraceID:    "trace-123",
		ProposalID: "proposal-1",
		Repo:       "rsi-agent-platform",
		BaseRef:    "main",
		Commands:   []string{"bash", "-lc", "echo hello"},
	})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}
	if session.PodName == "" {
		t.Fatal("expected created job name")
	}
	if job.Namespace != "rsi-platform" {
		t.Fatalf("unexpected job namespace %s", job.Namespace)
	}
	if _, err := launcher.GetJob(context.Background(), job.Namespace, job.Name); err != nil {
		t.Fatalf("GetJob() error = %v", err)
	}
}

func TestBuildBatchJobUsesSandboxContract(t *testing.T) {
	cfg := config.Config{
		SandboxNamespace:       "rsi-platform",
		SandboxImage:           "example.com/rsi-sandbox:latest",
		SandboxServiceAccount:  "rsi-sandbox",
		SandboxJobTTLSeconds:   3600,
		SandboxDeadlineSeconds: 900,
	}
	job := BuildBatchJob(cfg, JobRequest{
		TraceID:    "trace-123",
		ProposalID: "proposal-1",
		Repo:       "depin-backend",
		Commands:   []string{"bash", "-lc", "make test"},
	})
	if job.TypeMeta.Kind != "Job" {
		t.Fatalf("unexpected kind %s", job.TypeMeta.Kind)
	}
	if job.Spec.Template.Spec.Containers[0].Image != "example.com/rsi-sandbox:latest" {
		t.Fatalf("unexpected image %s", job.Spec.Template.Spec.Containers[0].Image)
	}
	if job.Spec.Template.Spec.Containers[0].Command[2] != "make test" {
		t.Fatalf("unexpected command %v", job.Spec.Template.Spec.Containers[0].Command)
	}
	if job.ObjectMeta.Labels["app.kubernetes.io/component"] != "sandbox-runtime" {
		t.Fatalf("unexpected labels %#v", job.ObjectMeta.Labels)
	}
	if _, err := fake.NewSimpleClientset(job).BatchV1().Jobs(job.Namespace).Get(context.Background(), job.Name, metav1.GetOptions{}); err != nil {
		t.Fatalf("fake job lookup error = %v", err)
	}
}
