package toolgateway

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const workspaceRepoDir = "/workspace/repo"

var allowedValidationCommands = map[string]struct{}{
	"make test":     {},
	"go test ./...": {},
	"pytest":        {},
	"pnpm test":     {},
	"npm test":      {},
	"yarn test":     {},
}

func (s *Service) workspaceListFiles(input map[string]interface{}) storepkg.ToolResult {
	workspace, _, errResult := s.resolveWorkspace(input, "workspace.list_files")
	if errResult != nil {
		return *errResult
	}
	relPath, result := s.workspacePathInput(input, "path", false, "workspace.list_files")
	if result != nil {
		return *result
	}
	command := fmt.Sprintf("cd %s && find %s -mindepth 1 -maxdepth 4 | sort | sed 's#^\\./##' | head -200", shellQuote(workspaceRepoDir), shellQuote(firstNonEmpty(relPath, ".")))
	execResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", command}, "workspace.list_files", input)
	if failed != nil {
		return *failed
	}
	items := compactLines(execResult.Stdout)
	summary := fmt.Sprintf("Workspace listed %d path(s) for attempt %s.", len(items), workspace.AttemptID)
	return s.result("workspace.list_files", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"path":         firstNonEmpty(relPath, "."),
		"items":        items,
	}, nil)
}

func (s *Service) workspaceReadFile(input map[string]interface{}) storepkg.ToolResult {
	workspace, _, errResult := s.resolveWorkspace(input, "workspace.read_file")
	if errResult != nil {
		return *errResult
	}
	relPath, result := s.workspacePathInput(input, "path", true, "workspace.read_file")
	if result != nil {
		return *result
	}
	command := fmt.Sprintf("cd %s && sed -n '1,260p' %s", shellQuote(workspaceRepoDir), shellQuote(relPath))
	execResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", command}, "workspace.read_file", input)
	if failed != nil {
		return *failed
	}
	content := truncate(execResult.Stdout, 12000)
	summary := fmt.Sprintf("Workspace loaded %s for attempt %s.", relPath, workspace.AttemptID)
	return s.result("workspace.read_file", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"path":         relPath,
		"content":      content,
	}, nil)
}

func (s *Service) workspaceSearch(input map[string]interface{}) storepkg.ToolResult {
	workspace, _, errResult := s.resolveWorkspace(input, "workspace.search")
	if errResult != nil {
		return *errResult
	}
	pattern := strings.TrimSpace(stringValue(input["pattern"]))
	if pattern == "" {
		return s.failedResult("workspace.search", input, "sandbox", "Workspace search requires pattern.", map[string]interface{}{"error": "missing pattern"})
	}
	relPath, result := s.workspacePathInput(input, "path", false, "workspace.search")
	if result != nil {
		return *result
	}
	scope := firstNonEmpty(relPath, ".")
	command := fmt.Sprintf("cd %s && rg -n --hidden --glob '!.git' %s %s | head -200", shellQuote(workspaceRepoDir), shellQuote(pattern), shellQuote(scope))
	execResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", command}, "workspace.search", input)
	if failed != nil {
		return *failed
	}
	matches := compactLines(execResult.Stdout)
	summary := fmt.Sprintf("Workspace search found %d match(es) in %s.", len(matches), scope)
	return s.result("workspace.search", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"path":         scope,
		"pattern":      pattern,
		"matches":      matches,
	}, nil)
}

func (s *Service) workspaceWriteFile(input map[string]interface{}) storepkg.ToolResult {
	workspace, _, errResult := s.resolveWorkspace(input, "workspace.write_file")
	if errResult != nil {
		return *errResult
	}
	relPath, result := s.workspacePathInput(input, "path", true, "workspace.write_file")
	if result != nil {
		return *result
	}
	content := stringValue(input["content"])
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	command := fmt.Sprintf("cd %s && mkdir -p %s && printf '%%s' %s | base64 -d > %s", shellQuote(workspaceRepoDir), shellQuote(filepath.Dir(relPath)), shellQuote(encoded), shellQuote(relPath))
	_, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", command}, "workspace.write_file", input)
	if failed != nil {
		return *failed
	}
	summary := fmt.Sprintf("Workspace wrote %s for attempt %s.", relPath, workspace.AttemptID)
	return s.result("workspace.write_file", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"path":         relPath,
		"bytes":        len(content),
	}, nil)
}

func (s *Service) workspaceApplyPatch(input map[string]interface{}) storepkg.ToolResult {
	workspace, _, errResult := s.resolveWorkspace(input, "workspace.apply_patch")
	if errResult != nil {
		return *errResult
	}
	patch := stringValue(input["patch"])
	if strings.TrimSpace(patch) == "" {
		return s.failedResult("workspace.apply_patch", input, "sandbox", "Workspace patch requires patch content.", map[string]interface{}{"error": "missing patch"})
	}
	changedFiles := patchChangedFiles(patch)
	if len(changedFiles) == 0 {
		return s.failedResult("workspace.apply_patch", input, "sandbox", "Workspace patch does not touch any files.", map[string]interface{}{"error": "no_changed_files"})
	}
	for _, file := range changedFiles {
		if !workspaceAllowsPath(workspace, file) {
			return s.failedResult("workspace.apply_patch", input, "sandbox", fmt.Sprintf("Workspace patch touches disallowed path %s.", file), map[string]interface{}{
				"error":         "disallowed_path",
				"changed_files": changedFiles,
			})
		}
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(patch))
	command := fmt.Sprintf("cd %s && printf '%%s' %s | base64 -d > /tmp/rsi-workspace.patch && git apply --whitespace=nowarn /tmp/rsi-workspace.patch", shellQuote(workspaceRepoDir), shellQuote(encoded))
	_, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", command}, "workspace.apply_patch", input)
	if failed != nil {
		return *failed
	}
	summary := fmt.Sprintf("Workspace applied patch touching %d file(s).", len(changedFiles))
	return s.result("workspace.apply_patch", input, summary, map[string]interface{}{
		"workspace_id":  workspace.ID,
		"attempt_id":    workspace.AttemptID,
		"repo":          workspace.Repo,
		"changed_files": changedFiles,
	}, nil)
}

func (s *Service) workspaceGitStatus(input map[string]interface{}) storepkg.ToolResult {
	workspace, _, errResult := s.resolveWorkspace(input, "workspace.git_status")
	if errResult != nil {
		return *errResult
	}
	execResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", fmt.Sprintf("cd %s && git status --short", shellQuote(workspaceRepoDir))}, "workspace.git_status", input)
	if failed != nil {
		return *failed
	}
	lines := compactLines(execResult.Stdout)
	summary := fmt.Sprintf("Workspace git status has %d changed path(s).", len(lines))
	return s.result("workspace.git_status", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"status":       lines,
	}, nil)
}

func (s *Service) workspaceGitDiff(input map[string]interface{}) storepkg.ToolResult {
	workspace, workspaceUpdated, errResult := s.resolveWorkspace(input, "workspace.git_diff")
	if errResult != nil {
		return *errResult
	}
	statusResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", fmt.Sprintf("cd %s && git status --short", shellQuote(workspaceRepoDir))}, "workspace.git_diff", input)
	if failed != nil {
		return *failed
	}
	diffResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", fmt.Sprintf("cd %s && git diff --binary", shellQuote(workspaceRepoDir))}, "workspace.git_diff", input)
	if failed != nil {
		return *failed
	}
	statResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", fmt.Sprintf("cd %s && git diff --stat", shellQuote(workspaceRepoDir))}, "workspace.git_diff", input)
	if failed != nil {
		return *failed
	}
	headResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", fmt.Sprintf("cd %s && git rev-parse HEAD", shellQuote(workspaceRepoDir))}, "workspace.git_diff", input)
	if failed != nil {
		return *failed
	}
	changedFiles := gitStatusChangedFiles(statusResult.Stdout)
	patch := truncate(diffResult.Stdout, 120000)
	diffSummary := truncate(strings.TrimSpace(statResult.Stdout), 4000)
	workspaceUpdated.HeadSHA = strings.TrimSpace(headResult.Stdout)
	workspaceUpdated.DiffSummary = diffSummary
	workspaceUpdated.UpdatedAt = time.Now().UTC()
	_, _ = s.store.UpsertAttemptWorkspace(workspaceUpdated)
	summary := fmt.Sprintf("Workspace diff for attempt %s touches %d file(s).", workspace.AttemptID, len(changedFiles))
	return s.result("workspace.git_diff", input, summary, map[string]interface{}{
		"workspace_id":  workspace.ID,
		"attempt_id":    workspace.AttemptID,
		"repo":          workspace.Repo,
		"head_sha":      workspaceUpdated.HeadSHA,
		"changed_files": changedFiles,
		"diff_summary":  diffSummary,
		"patch":         patch,
	}, nil)
}

func (s *Service) workspaceRunValidation(input map[string]interface{}) storepkg.ToolResult {
	workspace, workspaceUpdated, errResult := s.resolveWorkspace(input, "workspace.run_validation")
	if errResult != nil {
		return *errResult
	}
	command := firstNonEmpty(strings.TrimSpace(stringValue(input["command"])), "make test")
	if _, ok := allowedValidationCommands[command]; !ok {
		return s.failedResult("workspace.run_validation", input, "sandbox", fmt.Sprintf("Validation command %q is not allowed.", command), map[string]interface{}{
			"workspace_id": workspace.ID,
			"attempt_id":   workspace.AttemptID,
			"repo":         workspace.Repo,
			"command":      command,
			"error":        "command_not_allowed",
		})
	}
	workspaceUpdated.Status = improvement.WorkspaceValidating
	workspaceUpdated.UpdatedAt = time.Now().UTC()
	_, _ = s.store.UpsertAttemptWorkspace(workspaceUpdated)
	execResult, failed := s.execWorkspaceCommand(workspaceUpdated, []string{"bash", "-lc", fmt.Sprintf("cd %s && %s", shellQuote(workspaceRepoDir), command)}, "workspace.run_validation", input)
	completedAt := time.Now().UTC()
	if failed != nil {
		workspaceUpdated.Status = improvement.WorkspaceFailed
		workspaceUpdated.UpdatedAt = completedAt
		_, _ = s.store.UpsertAttemptWorkspace(workspaceUpdated)
		return *failed
	}
	workspaceUpdated.Status = improvement.WorkspaceCompleted
	workspaceUpdated.UpdatedAt = completedAt
	_, _ = s.store.UpsertAttemptWorkspace(workspaceUpdated)
	summary := fmt.Sprintf("Workspace validation succeeded for attempt %s using %q.", workspace.AttemptID, command)
	return s.result("workspace.run_validation", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"command":      command,
		"stdout":       truncate(execResult.Stdout, 12000),
		"stderr":       truncate(execResult.Stderr, 12000),
		"ok":           true,
	}, nil)
}

func (s *Service) resolveWorkspace(input map[string]interface{}, toolName string) (improvement.AttemptWorkspace, improvement.AttemptWorkspace, *storepkg.ToolResult) {
	if s.launcher == nil {
		result := s.unavailableResult(toolName, input, "sandbox", "Workspace tools unavailable: sandbox launcher not configured.", map[string]interface{}{"error": "sandbox_launcher_unavailable"})
		return improvement.AttemptWorkspace{}, improvement.AttemptWorkspace{}, &result
	}
	workspaceID := strings.TrimSpace(stringValue(input["workspace_id"]))
	attemptID := strings.TrimSpace(stringValue(input["attempt_id"]))
	var (
		workspace improvement.AttemptWorkspace
		ok        bool
	)
	if workspaceID != "" {
		workspace, ok = s.store.GetAttemptWorkspace(workspaceID)
	} else if attemptID != "" {
		workspace, ok = s.store.GetAttemptWorkspaceByAttempt(attemptID)
	}
	if !ok {
		result := s.failedResult(toolName, input, "sandbox", "Workspace not found.", map[string]interface{}{
			"workspace_id": workspaceID,
			"attempt_id":   attemptID,
			"error":        "workspace_not_found",
		})
		return improvement.AttemptWorkspace{}, improvement.AttemptWorkspace{}, &result
	}
	if workspace.Namespace == "" || workspace.JobName == "" {
		result := s.failedResult(toolName, input, "sandbox", "Workspace is missing Kubernetes identity.", map[string]interface{}{
			"workspace_id": workspace.ID,
			"attempt_id":   workspace.AttemptID,
			"error":        "workspace_identity_missing",
		})
		return improvement.AttemptWorkspace{}, improvement.AttemptWorkspace{}, &result
	}
	updated := workspace
	if strings.TrimSpace(updated.PodName) == "" {
		podName, err := s.launcher.ResolvePod(context.Background(), updated.Namespace, updated.JobName)
		if err != nil {
			result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("Workspace pod resolution failed: %v", err), map[string]interface{}{
				"workspace_id": workspace.ID,
				"attempt_id":   workspace.AttemptID,
				"job_name":     workspace.JobName,
			})
			return improvement.AttemptWorkspace{}, improvement.AttemptWorkspace{}, &result
		}
		updated.PodName = podName
		updated.UpdatedAt = time.Now().UTC()
		_, _ = s.store.UpsertAttemptWorkspace(updated)
	}
	return workspace, updated, nil
}

func (s *Service) execWorkspaceCommand(workspace improvement.AttemptWorkspace, command []string, toolName string, input map[string]interface{}) (sandbox.ExecResult, *storepkg.ToolResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := s.launcher.Exec(ctx, workspace.Namespace, workspace.PodName, command)
	if err != nil {
		failed := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("Workspace command failed: %v", err), map[string]interface{}{
			"workspace_id": workspace.ID,
			"attempt_id":   workspace.AttemptID,
			"repo":         workspace.Repo,
			"pod_name":     workspace.PodName,
			"command":      command,
			"stdout":       truncate(result.Stdout, 8000),
			"stderr":       truncate(result.Stderr, 8000),
			"error":        err.Error(),
		})
		return sandbox.ExecResult{}, &failed
	}
	return result, nil
}

func (s *Service) workspacePathInput(input map[string]interface{}, key string, required bool, toolName string) (string, *storepkg.ToolResult) {
	relPath := strings.TrimSpace(stringValue(input[key]))
	if relPath == "" {
		if required {
			result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("%s requires %s.", toolName, key), map[string]interface{}{"error": "missing_path"})
			return "", &result
		}
		return "", nil
	}
	relPath = filepath.Clean(strings.TrimPrefix(relPath, "/"))
	if relPath == "." || strings.HasPrefix(relPath, "..") {
		result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("%s path %q is outside the workspace.", toolName, relPath), map[string]interface{}{"error": "path_outside_workspace"})
		return "", &result
	}
	return relPath, nil
}

func workspaceAllowsPath(workspace improvement.AttemptWorkspace, relPath string) bool {
	relPath = filepath.ToSlash(filepath.Clean(strings.TrimSpace(relPath)))
	if relPath == "" || relPath == "." || strings.HasPrefix(relPath, "..") {
		return false
	}
	if len(workspace.AllowedPathGlobs) == 0 {
		return isDefaultAllowedWorkspacePath(relPath)
	}
	for _, pattern := range workspace.AllowedPathGlobs {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		switch {
		case strings.HasSuffix(pattern, "/**"):
			prefix := strings.TrimSuffix(pattern, "/**")
			if relPath == prefix || strings.HasPrefix(relPath, prefix+"/") {
				return true
			}
		case pattern == relPath:
			return true
		}
	}
	return false
}

func isDefaultAllowedWorkspacePath(relPath string) bool {
	switch {
	case strings.HasPrefix(relPath, "cmd/"),
		strings.HasPrefix(relPath, "internal/"),
		strings.HasPrefix(relPath, "runner/"),
		strings.HasPrefix(relPath, "ui/"):
		return true
	case relPath == "README.md", relPath == "Makefile":
		return true
	default:
		return false
	}
}

func patchChangedFiles(patch string) []string {
	seen := map[string]struct{}{}
	files := []string{}
	for _, line := range strings.Split(patch, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "+++ b/") {
			continue
		}
		path := strings.TrimSpace(strings.TrimPrefix(line, "+++ b/"))
		if path == "" || path == "/dev/null" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}

func gitStatusChangedFiles(status string) []string {
	seen := map[string]struct{}{}
	files := []string{}
	for _, line := range compactLines(status) {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		path := filepath.ToSlash(strings.TrimSpace(fields[len(fields)-1]))
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}

func compactLines(body string) []string {
	lines := strings.Split(body, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
