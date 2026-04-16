package toolgateway

import (
	"context"
	"errors"
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
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposal.ID,
		CommandKind: string(transition.CommandProposalMarkRepoChangeRunning),
		CommandID:   "cmd-workspace-read-running",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SubmitCommand(proposal_mark_repo_change_running) error = %v", err)
	}
	now := time.Now().UTC()
	attempt, err := store.UpsertChangeAttempt(improvement.ChangeAttempt{
		ID:            "attempt-workspace-read",
		ProposalID:    proposal.ID,
		CandidateKey:  proposal.CandidateKey,
		AttemptNumber: 1,
		TargetLayer:   proposal.TargetLayer,
		TargetKind:    proposal.TargetKind,
		TargetRef:     proposal.TargetRef,
		BranchName:    "codex/workspace-read",
		State:         improvement.AttemptStatePatchPlan,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	workspace, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:         "workspace-read",
		AttemptID:  attempt.ID,
		ProposalID: proposal.ID,
		Repo:       "rsi-agent-platform",
		BaseRef:    "main",
		BranchName: attempt.BranchName,
		Namespace:  "rsi-platform",
		JobName:    "workspace-job-read",
		Status:     improvement.WorkspaceReady,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}

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
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposal.ID,
		CommandKind: string(transition.CommandProposalMarkRepoChangeRunning),
		CommandID:   "cmd-workspace-validate-running",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SubmitCommand(proposal_mark_repo_change_running) error = %v", err)
	}
	now := time.Now().UTC()
	attempt, err := store.UpsertChangeAttempt(improvement.ChangeAttempt{
		ID:            "attempt-workspace-validate",
		ProposalID:    proposal.ID,
		CandidateKey:  proposal.CandidateKey,
		AttemptNumber: 1,
		TargetLayer:   proposal.TargetLayer,
		TargetKind:    proposal.TargetKind,
		TargetRef:     proposal.TargetRef,
		BranchName:    "codex/workspace-validate",
		State:         improvement.AttemptStateValidationRunning,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	workspace, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:         "workspace-validate",
		AttemptID:  attempt.ID,
		ProposalID: proposal.ID,
		Repo:       "rsi-agent-platform",
		BaseRef:    "main",
		BranchName: attempt.BranchName,
		Namespace:  "rsi-platform",
		JobName:    "workspace-job-validate",
		PodName:    "workspace-pod-validate",
		Status:     improvement.WorkspaceReady,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}

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
