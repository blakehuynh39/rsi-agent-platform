package toolgateway

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
	batchv1 "k8s.io/api/batch/v1"
)

func findProposalForTest(store *storepkg.MemoryStore, proposalID string) (review.Proposal, bool) {
	for _, proposal := range store.ListProposals() {
		if proposal.ID == proposalID {
			return proposal, true
		}
	}
	return review.Proposal{}, false
}

func prepareWorkspaceForToolTest(t *testing.T, store *storepkg.MemoryStore, proposal review.Proposal, workspaceID string, branchName string, namespace string, jobName string, podName string) (review.Proposal, improvement.ChangeAttempt, improvement.AttemptWorkspace) {
	t.Helper()

	refreshed, ok := findProposalForTest(store, proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s", proposal.ID)
	}
	proposal = refreshed
	attemptID := proposal.CurrentAttemptID
	if attemptID == "" {
		t.Fatalf("expected proposal %s to materialize a current attempt", proposal.ID)
	}
	attempt, ok := store.GetChangeAttempt(attemptID)
	if !ok {
		t.Fatalf("expected attempt %s", attemptID)
	}

	now := time.Now().UTC()
	for _, command := range []transition.CommandEnvelope{
		{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposal.ID,
			CommandKind: string(transition.CommandProposalMarkRepoChangeQueued),
			CommandID:   fmt.Sprintf("cmd-proposal-queued:%s", proposal.ID),
			Actor:       "tester",
			OccurredAt:  now,
		},
		{
			MachineKind: transition.MachineAttempt,
			AggregateID: attempt.ID,
			CommandKind: string(transition.CommandWorkspaceReady),
			CommandID:   fmt.Sprintf("cmd-workspace-ready:%s", attempt.ID),
			Actor:       "tester",
			OccurredAt:  now.Add(time.Millisecond),
			Payload: map[string]any{
				"workspace_id":        workspaceID,
				"repo":                "rsi-agent-platform",
				"base_ref":            "main",
				"branch_name":         branchName,
				"workspace_namespace": namespace,
				"workspace_job_name":  jobName,
				"workspace_pod_name":  podName,
				"allowed_path_globs":  []string{"internal/**"},
				"validation_ref":      fmt.Sprintf("workspace:%s", workspaceID),
				"sandbox_namespace":   namespace,
				"sandbox_job_name":    jobName,
				"sandbox_pod_name":    podName,
			},
		},
		{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposal.ID,
			CommandKind: string(transition.CommandProposalMarkRepoChangeRunning),
			CommandID:   fmt.Sprintf("cmd-proposal-running:%s", proposal.ID),
			Actor:       "tester",
			OccurredAt:  now.Add(2 * time.Millisecond),
		},
	} {
		receipt, err := store.SubmitCommand(command)
		if err != nil {
			t.Fatalf("SubmitCommand(%s) error = %v", command.CommandKind, err)
		}
		if receipt.DecisionKind == transition.DecisionReject {
			t.Fatalf("SubmitCommand(%s) rejected: %s", command.CommandKind, receipt.Reason)
		}
	}

	proposal, ok = findProposalForTest(store, proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after workspace setup", proposal.ID)
	}
	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s after workspace setup", attempt.ID)
	}
	workspace, ok := store.GetAttemptWorkspace(workspaceID)
	if !ok {
		t.Fatalf("expected workspace %s after workspace setup", workspaceID)
	}
	return proposal, attempt, workspace
}

type workspaceLauncherStub struct {
	resolvedPod string
	resolveErr  error
	execResult  sandbox.ExecResult
	execErr     error
	execPodName string
}

func (f *workspaceLauncherStub) Launch(context.Context, sandbox.JobRequest) (sandbox.Session, *batchv1.Job, error) {
	return sandbox.Session{}, nil, errors.New("unexpected launch call")
}

func (f *workspaceLauncherStub) GetJob(context.Context, string, string) (*batchv1.Job, error) {
	return nil, errors.New("unexpected get job call")
}

func (f *workspaceLauncherStub) ObserveJob(context.Context, string, string) (sandbox.JobObservation, error) {
	return sandbox.JobObservation{}, errors.New("unexpected observe job call")
}

func (f *workspaceLauncherStub) ResolvePod(context.Context, string, string) (string, error) {
	if f.resolveErr != nil {
		return "", f.resolveErr
	}
	return f.resolvedPod, nil
}

func (f *workspaceLauncherStub) Exec(_ context.Context, _, podName string, _ []string) (sandbox.ExecResult, error) {
	f.execPodName = podName
	if f.execErr != nil {
		return sandbox.ExecResult{}, f.execErr
	}
	return f.execResult, nil
}

func TestWorkspaceReadFileUsesResolvedPodAndFormalMetadataCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	base := store.ListProposals()[0]
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal, attempt, workspace := prepareWorkspaceForToolTest(t, store, proposal, "workspace-read", "codex/workspace-read", "rsi-platform", "workspace-job-read", "")

	launcher := &workspaceLauncherStub{
		resolvedPod: "workspace-pod-read",
		execResult: sandbox.ExecResult{
			Stdout: "package store\n",
		},
	}
	service := NewService(config.Config{ServiceName: "tool-gateway"}, store)
	service.launcher = launcher

	result := service.Execute("workspace.read_file", map[string]interface{}{
		"workspace_id": workspace.ID,
		"path":         "internal/store/commands.go",
		"tool_call_id": "tool-read-1",
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok result, got %+v", result)
	}
	if launcher.execPodName != "workspace-pod-read" {
		t.Fatalf("expected exec to use resolved pod, got %q", launcher.execPodName)
	}
	persisted, ok := store.GetAttemptWorkspace(workspace.ID)
	if !ok {
		t.Fatalf("expected workspace %s", workspace.ID)
	}
	if persisted.PodName != "workspace-pod-read" {
		t.Fatalf("workspace pod = %q, want workspace-pod-read", persisted.PodName)
	}
	receiptID := "cmd-attempt:" + attempt.ID + ":" + string(transition.CommandWorkspaceMetadataSynced) + ":tool-read-1"
	if receipt, ok := store.GetCommandReceipt(receiptID); !ok || receipt.MachineKind != transition.MachineAttempt {
		t.Fatalf("expected metadata sync command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
}

func TestWorkspaceRunValidationUsesFormalWorkspaceCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	base := store.ListProposals()[0]
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal, attempt, workspace := prepareWorkspaceForToolTest(t, store, proposal, "workspace-validate", "codex/workspace-validate", "rsi-platform", "workspace-job-validate", "workspace-pod-validate")

	launcher := &workspaceLauncherStub{
		execResult: sandbox.ExecResult{
			Stdout: "ok\n",
		},
	}
	service := NewService(config.Config{ServiceName: "tool-gateway"}, store)
	service.launcher = launcher

	result := service.Execute("workspace.run_validation", map[string]interface{}{
		"workspace_id": workspace.ID,
		"command":      "go test ./...",
		"tool_call_id": "tool-validate-1",
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok result, got %+v", result)
	}
	persisted, ok := store.GetAttemptWorkspace(workspace.ID)
	if !ok {
		t.Fatalf("expected workspace %s", workspace.ID)
	}
	if persisted.Status != improvement.WorkspaceCompleted {
		t.Fatalf("workspace status = %s, want %s", persisted.Status, improvement.WorkspaceCompleted)
	}
	startReceiptID := "cmd-attempt:" + attempt.ID + ":" + string(transition.CommandWorkspaceToolValidationStarted) + ":tool-validate-1"
	if receipt, ok := store.GetCommandReceipt(startReceiptID); !ok || receipt.MachineKind != transition.MachineAttempt {
		t.Fatalf("expected validation start command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	completeReceiptID := "cmd-attempt:" + attempt.ID + ":" + string(transition.CommandWorkspaceToolValidationCompleted) + ":tool-validate-1"
	if receipt, ok := store.GetCommandReceipt(completeReceiptID); !ok || receipt.MachineKind != transition.MachineAttempt {
		t.Fatalf("expected validation complete command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
}
