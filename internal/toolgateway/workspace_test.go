package toolgateway

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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
	resolvedPod  string
	resolveErr   error
	execResult   sandbox.ExecResult
	execResults  []sandbox.ExecResult
	execErr      error
	execPodName  string
	execCommands [][]string
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

func (f *workspaceLauncherStub) Exec(_ context.Context, _, podName string, command []string) (sandbox.ExecResult, error) {
	f.execPodName = podName
	f.execCommands = append(f.execCommands, append([]string(nil), command...))
	if f.execErr != nil {
		return sandbox.ExecResult{}, f.execErr
	}
	if len(f.execResults) > 0 {
		result := f.execResults[0]
		f.execResults = f.execResults[1:]
		return result, nil
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

func TestWorkspaceGitHistoryUsesBoundedLogCommandAndParsesEntries(t *testing.T) {
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
	_, _, workspace := prepareWorkspaceForToolTest(t, store, proposal, "workspace-history", "codex/workspace-history", "rsi-platform", "workspace-job-history", "workspace-pod-history")

	launcher := &workspaceLauncherStub{
		execResult: sandbox.ExecResult{
			Stdout: "0123456789abcdef\t2026-04-17T01:02:03Z\tblake\tAdd workspace history tool\nfedcba9876543210\t2026-04-16T03:04:05Z\talex\tTighten tests\n",
		},
	}
	service := NewService(config.Config{ServiceName: "tool-gateway"}, store)
	service.launcher = launcher

	result := service.Execute("workspace.git_history", map[string]interface{}{
		"workspace_id": workspace.ID,
		"ref":          "HEAD~5",
		"path":         "internal/store/commands.go",
		"limit":        5,
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok result, got %+v", result)
	}
	if len(launcher.execCommands) != 1 {
		t.Fatalf("expected one exec command, got %d", len(launcher.execCommands))
	}
	wantCommand := []string{
		"git", "-C", workspaceRepoDir, "log",
		"--date=iso-strict",
		"--no-color",
		"--max-count=5",
		"--format=%H%x09%ad%x09%an%x09%s",
		"--follow",
		"HEAD~5",
		"--",
		"internal/store/commands.go",
	}
	if !reflect.DeepEqual(launcher.execCommands[0], wantCommand) {
		t.Fatalf("exec command = %#v, want %#v", launcher.execCommands[0], wantCommand)
	}
	entries, ok := result.Output["entries"].([]map[string]interface{})
	if !ok || len(entries) != 2 {
		t.Fatalf("expected two parsed entries, got %#v", result.Output["entries"])
	}
	if got := entries[0]["short_sha"]; got != "0123456789ab" {
		t.Fatalf("first short sha = %#v, want 0123456789ab", got)
	}
	if got := entries[1]["subject"]; got != "Tighten tests" {
		t.Fatalf("second subject = %#v, want Tighten tests", got)
	}
}

func TestWorkspaceGitShowLoadsHistoricalFileContent(t *testing.T) {
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
	_, _, workspace := prepareWorkspaceForToolTest(t, store, proposal, "workspace-show", "codex/workspace-show", "rsi-platform", "workspace-job-show", "workspace-pod-show")

	launcher := &workspaceLauncherStub{
		execResult: sandbox.ExecResult{
			Stdout: "package store\n\nconst commandTopic = \"workspace\"\n",
		},
	}
	service := NewService(config.Config{ServiceName: "tool-gateway"}, store)
	service.launcher = launcher

	result := service.Execute("workspace.git_show", map[string]interface{}{
		"workspace_id": workspace.ID,
		"ref":          "HEAD~1",
		"path":         "internal/store/commands.go",
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok result, got %+v", result)
	}
	wantCommand := []string{"git", "-C", workspaceRepoDir, "show", "--no-color", "HEAD~1:internal/store/commands.go"}
	if len(launcher.execCommands) != 1 || !reflect.DeepEqual(launcher.execCommands[0], wantCommand) {
		t.Fatalf("exec command = %#v, want %#v", launcher.execCommands, wantCommand)
	}
	if got := result.Output["mode"]; got != "file" {
		t.Fatalf("mode = %#v, want file", got)
	}
	if got := result.Output["content"]; got != "package store\n\nconst commandTopic = \"workspace\"" {
		t.Fatalf("content = %#v", got)
	}
}

func TestWorkspaceGitSearchUsesPickaxeCommand(t *testing.T) {
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
	_, _, workspace := prepareWorkspaceForToolTest(t, store, proposal, "workspace-search-history", "codex/workspace-search-history", "rsi-platform", "workspace-job-search-history", "workspace-pod-search-history")

	launcher := &workspaceLauncherStub{
		execResult: sandbox.ExecResult{
			Stdout: "abcdef1234567890\t2026-04-15T12:30:00Z\tblake\tTrack workspace_id through attempt metadata\n",
		},
	}
	service := NewService(config.Config{ServiceName: "tool-gateway"}, store)
	service.launcher = launcher

	result := service.Execute("workspace.git_search", map[string]interface{}{
		"workspace_id": workspace.ID,
		"pattern":      "workspace_id",
		"search_type":  "diff",
		"limit":        3,
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok result, got %+v", result)
	}
	wantCommand := []string{
		"git", "-C", workspaceRepoDir, "log",
		"--date=iso-strict",
		"--no-color",
		"--max-count=3",
		"--format=%H%x09%ad%x09%an%x09%s",
		"-S",
		"workspace_id",
		"HEAD",
		"--",
		"internal",
	}
	if len(launcher.execCommands) != 1 || !reflect.DeepEqual(launcher.execCommands[0], wantCommand) {
		t.Fatalf("exec command = %#v, want %#v", launcher.execCommands, wantCommand)
	}
	if got := result.Output["search_type"]; got != "changes" {
		t.Fatalf("search_type = %#v, want changes", got)
	}
}

func TestWorkspaceSearchDefaultsToAllowedPathScopes(t *testing.T) {
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
	_, _, workspace := prepareWorkspaceForToolTest(t, store, proposal, "workspace-scoped-search", "codex/workspace-scoped-search", "rsi-platform", "workspace-job-scoped-search", "workspace-pod-scoped-search")

	launcher := &workspaceLauncherStub{
		execResult: sandbox.ExecResult{Stdout: "internal/store/commands.go:12:workspace_id\n"},
	}
	service := NewService(config.Config{ServiceName: "tool-gateway"}, store)
	service.launcher = launcher

	result := service.Execute("workspace.search", map[string]interface{}{
		"workspace_id": workspace.ID,
		"pattern":      "workspace_id",
	})
	if result.Status != "ok" {
		t.Fatalf("expected ok result, got %+v", result)
	}
	wantCommand := []string{"bash", "-lc", "cd '/workspace/repo' && rg -n --hidden --glob '!.git' 'workspace_id' 'internal' | head -200"}
	if len(launcher.execCommands) != 1 || !reflect.DeepEqual(launcher.execCommands[0], wantCommand) {
		t.Fatalf("exec command = %#v, want %#v", launcher.execCommands, wantCommand)
	}
}

func TestWorkspaceGitShowRejectsUnsafeRef(t *testing.T) {
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
	_, _, workspace := prepareWorkspaceForToolTest(t, store, proposal, "workspace-show-invalid", "codex/workspace-show-invalid", "rsi-platform", "workspace-job-show-invalid", "workspace-pod-show-invalid")

	launcher := &workspaceLauncherStub{}
	service := NewService(config.Config{ServiceName: "tool-gateway"}, store)
	service.launcher = launcher

	result := service.Execute("workspace.git_show", map[string]interface{}{
		"workspace_id": workspace.ID,
		"ref":          "--format=raw",
	})
	if result.Status != "failed" {
		t.Fatalf("expected failed result, got %+v", result)
	}
	if got := result.Output["error"]; got != "invalid_ref" {
		t.Fatalf("error = %#v, want invalid_ref", got)
	}
	if len(launcher.execCommands) != 0 {
		t.Fatalf("expected no exec command for invalid ref, got %#v", launcher.execCommands)
	}
}

func TestWorkspaceGitShowRequiresPathForRestrictedScope(t *testing.T) {
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
	_, _, workspace := prepareWorkspaceForToolTest(t, store, proposal, "workspace-show-restricted", "codex/workspace-show-restricted", "rsi-platform", "workspace-job-show-restricted", "workspace-pod-show-restricted")

	launcher := &workspaceLauncherStub{}
	service := NewService(config.Config{ServiceName: "tool-gateway"}, store)
	service.launcher = launcher

	result := service.Execute("workspace.git_show", map[string]interface{}{
		"workspace_id": workspace.ID,
		"ref":          "HEAD~1",
	})
	if result.Status != "failed" {
		t.Fatalf("expected failed result, got %+v", result)
	}
	if got := result.Output["error"]; got != "path_required_for_restricted_scope" {
		t.Fatalf("error = %#v, want path_required_for_restricted_scope", got)
	}
	if len(launcher.execCommands) != 0 {
		t.Fatalf("expected no exec command for restricted commit show, got %#v", launcher.execCommands)
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
