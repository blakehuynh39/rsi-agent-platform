package improvementplane

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/githubapp"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func ensureAttemptWorkspace(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, proposal review.Proposal, attempt improvement.ChangeAttempt, traceID string) (improvement.AttemptWorkspace, bool, error) {
	if launcherErr != nil {
		return improvement.AttemptWorkspace{}, false, launcherErr
	}
	if launcher == nil {
		return improvement.AttemptWorkspace{}, false, fmt.Errorf("sandbox launcher not configured")
	}
	if workspace, ok := store.GetAttemptWorkspaceByAttempt(attempt.ID); ok {
		if workspace.PodName == "" && workspace.JobName != "" {
			if podName, err := launcher.ResolvePod(context.Background(), workspace.Namespace, workspace.JobName); err == nil {
				workspace.PodName = podName
				workspace.Status = improvement.WorkspaceReady
				workspace.UpdatedAt = time.Now().UTC()
				updated, err := store.UpsertAttemptWorkspace(workspace)
				if err != nil {
					return improvement.AttemptWorkspace{}, false, err
				}
				workspace = updated
			}
		}
		return workspace, workspace.PodName != "", nil
	}

	targetRepo := proposalTargetRepo(cfg, proposal)
	repoOwner := cfg.GitHubRepoOwner(targetRepo)
	repoName := cfg.GitHubRepoName(targetRepo)
	writeToken, err := githubapp.NewClient(
		cfg.GitHubAppID,
		cfg.GitHubInstallationIDForRepo(targetRepo),
		cfg.GitHubAppPrivateKey,
		cfg.GitHubAPIBaseURL,
		&http.Client{Timeout: 30 * time.Second},
	).MintInstallationToken(context.Background(), []string{repoName})
	if err != nil {
		return improvement.AttemptWorkspace{}, false, fmt.Errorf("mint github app installation token for workspace open: %w", err)
	}

	request := sandbox.JobRequest{
		TraceID:      traceID,
		ProposalID:   proposal.ID,
		Repo:         repoName,
		BaseRef:      "main",
		RequestedBy:  cfg.ServiceName,
		ArtifactPath: fmt.Sprintf("memory://workspace/%s", attempt.ID),
		Env: map[string]string{
			"GITHUB_TOKEN":        writeToken.Token,
			"GITHUB_OWNER":        repoOwner,
			"GITHUB_COMMIT_USER":  cfg.GitHubCommitUser,
			"GITHUB_COMMIT_EMAIL": cfg.GitHubCommitEmail,
			"RSI_BRANCH_NAME":     attempt.BranchName,
			"RSI_CONTEXT_SUMMARY": proposalRunnerContextSummary(proposal),
			"RSI_CHANGE_PLAN":     proposal.Summary,
			"RSI_VALIDATION_PLAN": proposal.ValidationPlan,
			"RSI_ATTEMPT_ID":      attempt.ID,
			"RSI_REPO":            repoName,
			"RSI_BASE_REF":        "main",
			"RSI_PROPOSAL_ID":     proposal.ID,
		},
		Commands:    workspaceOpenCommands(),
		TimeoutSecs: 60 * 60,
	}
	session, job, err := launcher.Launch(context.Background(), request)
	if err != nil {
		return improvement.AttemptWorkspace{}, false, err
	}
	workspace := improvement.AttemptWorkspace{
		ID:               fmt.Sprintf("workspace-%s", attempt.ID),
		AttemptID:        attempt.ID,
		ProposalID:       proposal.ID,
		Repo:             repoName,
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Namespace:        session.Namespace,
		JobName:          job.Name,
		Status:           improvement.WorkspaceQueued,
		AllowedPathGlobs: defaultWorkspaceAllowedPathGlobs(),
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	updated, err := store.UpsertAttemptWorkspace(workspace)
	if err != nil {
		return improvement.AttemptWorkspace{}, false, err
	}
	workspace = updated
	return workspace, false, nil
}

func ensureWorkspaceRepoChangeJob(store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, workspace improvement.AttemptWorkspace) (improvement.RepoChangeJob, error) {
	for _, item := range store.ListRepoChangeJobs() {
		if item.ProposalID == proposal.ID && item.AttemptID == attempt.ID {
			item.Repo = firstNonEmpty(item.Repo, workspace.Repo)
			item.BaseRef = firstNonEmpty(item.BaseRef, workspace.BaseRef)
			item.BranchName = firstNonEmpty(item.BranchName, workspace.BranchName)
			item.AllowedPathGlobs = append([]string(nil), workspace.AllowedPathGlobs...)
			item.SandboxNamespace = workspace.Namespace
			item.SandboxJobName = workspace.JobName
			item.SandboxPodName = workspace.PodName
			item.ValidationRef = firstNonEmpty(item.ValidationRef, fmt.Sprintf("%s/%s", workspace.Namespace, firstNonEmpty(workspace.PodName, workspace.JobName)))
			item.UpdatedAt = time.Now().UTC()
			return store.UpsertRepoChangeJob(item)
		}
	}
	now := time.Now().UTC()
	return store.UpsertRepoChangeJob(improvement.RepoChangeJob{
		ID:               fmt.Sprintf("job-%s", attempt.ID),
		ProposalID:       proposal.ID,
		AttemptID:        attempt.ID,
		ConversationID:   proposal.ConversationID,
		CaseID:           proposal.CaseID,
		OriginTraceID:    firstNonEmpty(attempt.AttemptTraceID, proposal.OriginTraceID, proposal.TraceID),
		CandidateKey:     proposal.CandidateKey,
		Status:           string(review.ProposalRepoChangeRunning),
		Repo:             workspace.Repo,
		BaseRef:          workspace.BaseRef,
		BranchName:       workspace.BranchName,
		AllowedPathGlobs: append([]string(nil), workspace.AllowedPathGlobs...),
		ContextSummary:   proposalRunnerContextSummary(proposal),
		SandboxNamespace: workspace.Namespace,
		SandboxJobName:   workspace.JobName,
		SandboxPodName:   workspace.PodName,
		ValidationRef:    fmt.Sprintf("%s/%s", workspace.Namespace, firstNonEmpty(workspace.PodName, workspace.JobName)),
		CreatedAt:        now,
		UpdatedAt:        now,
	})
}

func workspaceOpenCommands() []string {
	script := `
set -euo pipefail
mkdir -p /workspace
cd /workspace
rm -rf repo
git clone "https://x-access-token:${GITHUB_TOKEN}@github.com/${GITHUB_OWNER}/${RSI_REPO}.git" repo
cd repo
git checkout -B "${RSI_BRANCH_NAME}" "origin/${RSI_BASE_REF}"
git config user.name "${GITHUB_COMMIT_USER}"
git config user.email "${GITHUB_COMMIT_EMAIL}"
mkdir -p .rsi
printf "%s\n" "${RSI_CONTEXT_SUMMARY:-}" > .rsi/proposal-context.txt
printf "%s\n" "${RSI_CHANGE_PLAN:-}" > .rsi/change-plan.txt
printf "%s\n" "${RSI_VALIDATION_PLAN:-}" > .rsi/validation-plan.txt
touch /workspace/.workspace-ready
trap : TERM INT; sleep infinity & wait
`
	return []string{"bash", "-lc", script}
}

func defaultWorkspaceAllowedPathGlobs() []string {
	return []string{"cmd/**", "internal/**", "runner/**", "ui/**", "README.md", "Makefile"}
}

func workspaceValidationCommand(outputValidationPlan string) string {
	plan := strings.ToLower(strings.TrimSpace(outputValidationPlan))
	switch {
	case strings.Contains(plan, "go test ./..."):
		return "go test ./..."
	case strings.Contains(plan, "pytest"):
		return "pytest"
	case strings.Contains(plan, "pnpm test"):
		return "pnpm test"
	case strings.Contains(plan, "npm test"):
		return "npm test"
	case strings.Contains(plan, "yarn test"):
		return "yarn test"
	default:
		return "make test"
	}
}
