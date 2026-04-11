package sandbox

import (
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

func TestBuildJobUsesConfiguredSandboxContract(t *testing.T) {
	cfg := config.Config{
		SandboxNamespace:      "rsi-platform",
		SandboxImage:          "example.com/rsi-sandbox:latest",
		SandboxServiceAccount: "rsi-sandbox",
		SandboxJobTTLSeconds:  3600,
		SandboxDeadlineSeconds: 900,
	}
	job := BuildJob(cfg, JobRequest{
		TraceID:     "trace-123",
		ProposalID:  "proposal-1",
		Repo:        "depin-backend",
		RequestedBy: "alice",
		Commands:    []string{"bash", "-lc", "make test"},
		Env:         map[string]string{"CI": "true"},
	})

	if job.Metadata.Namespace != "rsi-platform" {
		t.Fatalf("unexpected namespace: %s", job.Metadata.Namespace)
	}
	if job.Spec.Template.Spec.ServiceAccountName != "rsi-sandbox" {
		t.Fatalf("unexpected service account: %s", job.Spec.Template.Spec.ServiceAccountName)
	}
	if got := job.Spec.Template.Spec.Containers[0].Image; got != "example.com/rsi-sandbox:latest" {
		t.Fatalf("unexpected image: %s", got)
	}
	if got := job.Spec.Template.Spec.Containers[0].Command[len(job.Spec.Template.Spec.Containers[0].Command)-1]; got != "make test" {
		t.Fatalf("unexpected command tail: %s", got)
	}
	if job.Spec.BackoffLimit != 0 {
		t.Fatalf("unexpected backoff limit: %d", job.Spec.BackoffLimit)
	}
}
