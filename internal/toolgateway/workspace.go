package toolgateway

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
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
	workspace, errResult := s.resolveWorkspace(input, "workspace.list_files")
	if errResult != nil {
		return *errResult
	}
	relPath, result := s.workspacePathInput(input, "path", false, "workspace.list_files")
	if result != nil {
		return *result
	}
	scopes, result := s.workspaceReadScopes(workspace, relPath, "workspace.list_files", input)
	if result != nil {
		return *result
	}
	command := fmt.Sprintf("cd %s && find %s -mindepth 1 -maxdepth 4 | sort | sed 's#^\\./##' | head -200", shellQuote(workspaceRepoDir), shellQuoteJoin(scopes))
	execResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", command}, "workspace.list_files", input)
	if failed != nil {
		return *failed
	}
	items := compactLines(execResult.Stdout)
	summary := fmt.Sprintf("Workspace listed %d path(s) for attempt %s.", len(items), workspace.AttemptID)
	scope := firstNonEmpty(relPath, strings.Join(scopes, ", "))
	return s.result("workspace.list_files", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"path":         scope,
		"items":        items,
	}, nil)
}

func (s *Service) workspaceReadFile(input map[string]interface{}) storepkg.ToolResult {
	workspace, errResult := s.resolveWorkspace(input, "workspace.read_file")
	if errResult != nil {
		return *errResult
	}
	relPath, result := s.workspacePathInput(input, "path", true, "workspace.read_file")
	if result != nil {
		return *result
	}
	if !workspaceAllowsPath(workspace, relPath) {
		return s.failedResult("workspace.read_file", input, "sandbox", fmt.Sprintf("Workspace read path %s is not allowed.", relPath), map[string]interface{}{
			"workspace_id": workspace.ID,
			"attempt_id":   workspace.AttemptID,
			"path":         relPath,
			"error":        "disallowed_path",
		})
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
	workspace, errResult := s.resolveWorkspace(input, "workspace.search")
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
	scopes, result := s.workspaceReadScopes(workspace, relPath, "workspace.search", input)
	if result != nil {
		return *result
	}
	scope := firstNonEmpty(relPath, strings.Join(scopes, ", "))
	command := fmt.Sprintf("cd %s && rg -n --hidden --glob '!.git' %s %s | head -200", shellQuote(workspaceRepoDir), shellQuote(pattern), shellQuoteJoin(scopes))
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

func (s *Service) workspaceGitHistory(input map[string]interface{}) storepkg.ToolResult {
	workspace, errResult := s.resolveWorkspace(input, "workspace.git_history")
	if errResult != nil {
		return *errResult
	}
	ref, result := s.workspaceGitRefInput(input, "ref", "HEAD", "workspace.git_history")
	if result != nil {
		return *result
	}
	relPath, result := s.workspacePathInput(input, "path", false, "workspace.git_history")
	if result != nil {
		return *result
	}
	scopes, result := s.workspaceReadScopes(workspace, relPath, "workspace.git_history", input)
	if result != nil {
		return *result
	}
	limit := workspaceLimitInput(input, "limit", 20, 1, 50)
	command := []string{
		"git", "-C", workspaceRepoDir, "log",
		"--date=iso-strict",
		"--no-color",
		fmt.Sprintf("--max-count=%d", limit),
		"--format=%H%x09%ad%x09%an%x09%s",
	}
	if relPath != "" {
		command = append(command, "--follow")
	}
	command = append(command, ref)
	command = append(command, "--")
	command = append(command, scopes...)
	execResult, failed := s.execWorkspaceCommand(workspace, command, "workspace.git_history", input)
	if failed != nil {
		return *failed
	}
	entries := parseWorkspaceGitLogEntries(execResult.Stdout)
	scope := firstNonEmpty(relPath, strings.Join(scopes, ", "), ref)
	summary := fmt.Sprintf("Workspace git history returned %d commit(s) for %s.", len(entries), scope)
	return s.result("workspace.git_history", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"ref":          ref,
		"path":         relPath,
		"limit":        limit,
		"entries":      entries,
	}, nil)
}

func (s *Service) workspaceGitShow(input map[string]interface{}) storepkg.ToolResult {
	workspace, errResult := s.resolveWorkspace(input, "workspace.git_show")
	if errResult != nil {
		return *errResult
	}
	ref, result := s.workspaceGitRefInput(input, "ref", "HEAD", "workspace.git_show")
	if result != nil {
		return *result
	}
	relPath, result := s.workspacePathInput(input, "path", false, "workspace.git_show")
	if result != nil {
		return *result
	}
	command := []string{"git", "-C", workspaceRepoDir}
	mode := "commit"
	contentLimit := 40000
	if relPath != "" {
		if !workspaceAllowsPath(workspace, relPath) {
			return s.failedResult("workspace.git_show", input, "sandbox", fmt.Sprintf("Workspace git show path %s is not allowed.", relPath), map[string]interface{}{
				"workspace_id": workspace.ID,
				"attempt_id":   workspace.AttemptID,
				"path":         relPath,
				"error":        "disallowed_path",
			})
		}
		mode = "file"
		contentLimit = 12000
		command = append(command, "show", "--no-color", fmt.Sprintf("%s:%s", ref, relPath))
	} else {
		if !workspaceAllowsRepoWideRead(workspace) {
			return s.failedResult("workspace.git_show", input, "sandbox", "Workspace git show requires path when read scope is restricted.", map[string]interface{}{
				"workspace_id": workspace.ID,
				"attempt_id":   workspace.AttemptID,
				"error":        "path_required_for_restricted_scope",
			})
		}
		command = append(command, "show", "--stat", "--format=fuller", "--patch", "--no-ext-diff", "--no-color", ref)
	}
	execResult, failed := s.execWorkspaceCommand(workspace, command, "workspace.git_show", input)
	if failed != nil {
		return *failed
	}
	fullContent := strings.TrimSpace(execResult.Stdout)
	content := truncate(fullContent, contentLimit)
	summary := fmt.Sprintf("Workspace git show loaded %s.", ref)
	if relPath != "" {
		summary = fmt.Sprintf("Workspace git show loaded %s at %s.", relPath, ref)
	}
	return s.result("workspace.git_show", input, summary, map[string]interface{}{
		"workspace_id":   workspace.ID,
		"attempt_id":     workspace.AttemptID,
		"repo":           workspace.Repo,
		"ref":            ref,
		"path":           relPath,
		"mode":           mode,
		"content":        content,
		"content_length": len(fullContent),
		"truncated":      len(fullContent) > len(content),
	}, nil)
}

func (s *Service) workspaceGitSearch(input map[string]interface{}) storepkg.ToolResult {
	workspace, errResult := s.resolveWorkspace(input, "workspace.git_search")
	if errResult != nil {
		return *errResult
	}
	pattern := strings.TrimSpace(stringValue(input["pattern"]))
	if pattern == "" {
		return s.failedResult("workspace.git_search", input, "sandbox", "Workspace git search requires pattern.", map[string]interface{}{"error": "missing_pattern"})
	}
	mode, result := s.workspaceGitSearchModeInput(input, "workspace.git_search")
	if result != nil {
		return *result
	}
	ref, result := s.workspaceGitRefInput(input, "ref", "HEAD", "workspace.git_search")
	if result != nil {
		return *result
	}
	relPath, result := s.workspacePathInput(input, "path", false, "workspace.git_search")
	if result != nil {
		return *result
	}
	scopes, result := s.workspaceReadScopes(workspace, relPath, "workspace.git_search", input)
	if result != nil {
		return *result
	}
	limit := workspaceLimitInput(input, "limit", 20, 1, 50)
	command := []string{
		"git", "-C", workspaceRepoDir, "log",
		"--date=iso-strict",
		"--no-color",
		fmt.Sprintf("--max-count=%d", limit),
		"--format=%H%x09%ad%x09%an%x09%s",
	}
	switch mode {
	case "message":
		command = append(command, "--fixed-strings", "--grep="+pattern)
	default:
		command = append(command, "-S", pattern)
	}
	command = append(command, ref)
	command = append(command, "--")
	command = append(command, scopes...)
	execResult, failed := s.execWorkspaceCommand(workspace, command, "workspace.git_search", input)
	if failed != nil {
		return *failed
	}
	entries := parseWorkspaceGitLogEntries(execResult.Stdout)
	summary := fmt.Sprintf("Workspace git search found %d commit(s) for %q.", len(entries), pattern)
	return s.result("workspace.git_search", input, summary, map[string]interface{}{
		"workspace_id": workspace.ID,
		"attempt_id":   workspace.AttemptID,
		"repo":         workspace.Repo,
		"ref":          ref,
		"path":         relPath,
		"limit":        limit,
		"pattern":      pattern,
		"search_type":  mode,
		"entries":      entries,
	}, nil)
}

func (s *Service) workspaceWriteFile(input map[string]interface{}) storepkg.ToolResult {
	workspace, errResult := s.resolveWorkspace(input, "workspace.write_file")
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
	workspace, errResult := s.resolveWorkspace(input, "workspace.apply_patch")
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
	workspace, errResult := s.resolveWorkspace(input, "workspace.git_status")
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
	workspace, errResult := s.resolveWorkspace(input, "workspace.git_diff")
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
	workspace.HeadSHA = strings.TrimSpace(headResult.Stdout)
	workspace.DiffSummary = diffSummary
	if failed := s.submitWorkspaceAttemptCommand(workspace, "workspace.git_diff", input, transition.CommandWorkspaceMetadataSynced, map[string]any{
		"workspace_namespace": workspace.Namespace,
		"workspace_job_name":  workspace.JobName,
		"workspace_pod_name":  workspace.PodName,
		"head_sha":            workspace.HeadSHA,
		"diff_summary":        workspace.DiffSummary,
	}); failed != nil {
		return *failed
	}
	summary := fmt.Sprintf("Workspace diff for attempt %s touches %d file(s).", workspace.AttemptID, len(changedFiles))
	return s.result("workspace.git_diff", input, summary, map[string]interface{}{
		"workspace_id":  workspace.ID,
		"attempt_id":    workspace.AttemptID,
		"repo":          workspace.Repo,
		"head_sha":      workspace.HeadSHA,
		"changed_files": changedFiles,
		"diff_summary":  diffSummary,
		"patch":         patch,
	}, nil)
}

func (s *Service) workspaceRunValidation(input map[string]interface{}) storepkg.ToolResult {
	workspace, errResult := s.resolveWorkspace(input, "workspace.run_validation")
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
	if failed := s.submitWorkspaceAttemptCommand(workspace, "workspace.run_validation", input, transition.CommandWorkspaceToolValidationStarted, map[string]any{
		"validation_command": command,
	}); failed != nil {
		return *failed
	}
	execResult, failed := s.execWorkspaceCommand(workspace, []string{"bash", "-lc", fmt.Sprintf("cd %s && %s", shellQuote(workspaceRepoDir), command)}, "workspace.run_validation", input)
	if failed != nil {
		if syncFailed := s.submitWorkspaceAttemptCommand(workspace, "workspace.run_validation", input, transition.CommandWorkspaceToolValidationFailed, map[string]any{
			"validation_command": command,
			"failure_summary":    firstNonEmpty(stringValue(failed.Output["error"]), failed.Summary),
		}); syncFailed != nil {
			return *syncFailed
		}
		return *failed
	}
	if failed := s.submitWorkspaceAttemptCommand(workspace, "workspace.run_validation", input, transition.CommandWorkspaceToolValidationCompleted, map[string]any{
		"validation_command": command,
	}); failed != nil {
		return *failed
	}
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

func (s *Service) resolveWorkspace(input map[string]interface{}, toolName string) (improvement.AttemptWorkspace, *storepkg.ToolResult) {
	if s.launcher == nil {
		result := s.unavailableResult(toolName, input, "sandbox", "Workspace tools unavailable: sandbox launcher not configured.", map[string]interface{}{"error": "sandbox_launcher_unavailable"})
		return improvement.AttemptWorkspace{}, &result
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
		return improvement.AttemptWorkspace{}, &result
	}
	if workspace.Namespace == "" || workspace.JobName == "" {
		result := s.failedResult(toolName, input, "sandbox", "Workspace is missing Kubernetes identity.", map[string]interface{}{
			"workspace_id": workspace.ID,
			"attempt_id":   workspace.AttemptID,
			"error":        "workspace_identity_missing",
		})
		return improvement.AttemptWorkspace{}, &result
	}
	if strings.TrimSpace(workspace.PodName) == "" {
		podName, err := s.launcher.ResolvePod(context.Background(), workspace.Namespace, workspace.JobName)
		if err != nil {
			result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("Workspace pod resolution failed: %v", err), map[string]interface{}{
				"workspace_id": workspace.ID,
				"attempt_id":   workspace.AttemptID,
				"job_name":     workspace.JobName,
			})
			return improvement.AttemptWorkspace{}, &result
		}
		workspace.PodName = podName
		if failed := s.submitWorkspaceAttemptCommand(workspace, toolName, input, transition.CommandWorkspaceMetadataSynced, map[string]any{
			"workspace_namespace": workspace.Namespace,
			"workspace_job_name":  workspace.JobName,
			"workspace_pod_name":  workspace.PodName,
		}); failed != nil {
			return improvement.AttemptWorkspace{}, failed
		}
	}
	return workspace, nil
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

func workspaceLimitInput(input map[string]interface{}, key string, defaultValue int, minValue int, maxValue int) int {
	switch raw := input[key].(type) {
	case int:
		return maxInt(minValue, minInt(raw, maxValue))
	case int32:
		return maxInt(minValue, minInt(int(raw), maxValue))
	case int64:
		return maxInt(minValue, minInt(int(raw), maxValue))
	case float64:
		return maxInt(minValue, minInt(int(raw), maxValue))
	case string:
		if value, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
			return maxInt(minValue, minInt(value, maxValue))
		}
	}
	return defaultValue
}

func (s *Service) workspaceGitRefInput(input map[string]interface{}, key string, defaultValue string, toolName string) (string, *storepkg.ToolResult) {
	ref := firstNonEmpty(strings.TrimSpace(stringValue(input[key])), strings.TrimSpace(defaultValue))
	if ref == "" {
		result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("%s requires %s.", toolName, key), map[string]interface{}{"error": "missing_ref"})
		return "", &result
	}
	if strings.HasPrefix(ref, "-") || strings.Contains(ref, ":") || len(ref) > 200 {
		result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("%s ref %q is not allowed.", toolName, ref), map[string]interface{}{"error": "invalid_ref"})
		return "", &result
	}
	for _, r := range ref {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("%s ref %q is not allowed.", toolName, ref), map[string]interface{}{"error": "invalid_ref"})
			return "", &result
		}
	}
	return ref, nil
}

func (s *Service) workspaceGitSearchModeInput(input map[string]interface{}, toolName string) (string, *storepkg.ToolResult) {
	mode := strings.ToLower(firstNonEmpty(strings.TrimSpace(stringValue(input["search_type"])), strings.TrimSpace(stringValue(input["mode"])), "changes"))
	switch mode {
	case "changes", "change", "diff":
		return "changes", nil
	case "message":
		return "message", nil
	default:
		result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("%s search_type %q is not supported.", toolName, mode), map[string]interface{}{"error": "invalid_search_type"})
		return "", &result
	}
}

func (s *Service) workspaceReadScopes(workspace improvement.AttemptWorkspace, relPath string, toolName string, input map[string]interface{}) ([]string, *storepkg.ToolResult) {
	if relPath != "" {
		if !workspaceAllowsPath(workspace, relPath) {
			result := s.failedResult(toolName, input, "sandbox", fmt.Sprintf("%s path %s is not allowed.", toolName, relPath), map[string]interface{}{
				"workspace_id": workspace.ID,
				"attempt_id":   workspace.AttemptID,
				"path":         relPath,
				"error":        "disallowed_path",
			})
			return nil, &result
		}
		return []string{relPath}, nil
	}
	return workspaceAllowedReadScopes(workspace), nil
}

func workspaceAllowedReadScopes(workspace improvement.AttemptWorkspace) []string {
	if len(workspace.AllowedPathGlobs) == 0 {
		return []string{"cmd", "internal", "runner", "ui", "README.md", "Makefile"}
	}
	seen := map[string]struct{}{}
	scopes := make([]string, 0, len(workspace.AllowedPathGlobs))
	for _, pattern := range workspace.AllowedPathGlobs {
		scope := workspaceAllowedReadScope(pattern)
		if scope == "." {
			return []string{"."}
		}
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	if len(scopes) == 0 {
		return []string{"."}
	}
	sort.Strings(scopes)
	return scopes
}

func workspaceAllowedReadScope(pattern string) string {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	if pattern == "" {
		return ""
	}
	wildcard := strings.IndexAny(pattern, "*?[")
	if wildcard == -1 {
		return strings.Trim(pattern, "/")
	}
	if wildcard == 0 {
		return "."
	}
	scope := strings.TrimSuffix(pattern[:wildcard], "/")
	if scope == "" {
		return "."
	}
	return scope
}

func workspaceAllowsRepoWideRead(workspace improvement.AttemptWorkspace) bool {
	for _, scope := range workspaceAllowedReadScopes(workspace) {
		if scope == "." {
			return true
		}
	}
	return false
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

func (s *Service) submitWorkspaceAttemptCommand(workspace improvement.AttemptWorkspace, toolName string, input map[string]interface{}, kind transition.AttemptPhaseCommandKind, payload map[string]any) *storepkg.ToolResult {
	attemptID := strings.TrimSpace(workspace.AttemptID)
	if attemptID == "" {
		result := s.failedResult(toolName, input, "transition", "Workspace is missing attempt identity for formal state sync.", map[string]interface{}{
			"workspace_id": workspace.ID,
			"error":        "workspace_attempt_missing",
		})
		return &result
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["workspace_id"] = workspace.ID
	occurredAt := time.Now().UTC()
	receipt, err := s.store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAttempt,
		AggregateID: attemptID,
		CommandKind: string(kind),
		CommandID:   workspaceAttemptCommandID(workspace, kind, input, occurredAt),
		Actor:       firstNonEmpty(strings.TrimSpace(s.cfg.ServiceName), "tool-gateway"),
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		result := s.failedResult(toolName, input, "transition", fmt.Sprintf("Workspace state sync failed: %v", err), map[string]interface{}{
			"workspace_id": workspace.ID,
			"attempt_id":   attemptID,
			"command_kind": string(kind),
			"error":        "workspace_state_sync_failed",
		})
		return &result
	}
	if receipt.DecisionKind == transition.DecisionReject {
		result := s.failedResult(toolName, input, "transition", fmt.Sprintf("Workspace state sync rejected: %s", receipt.Reason), map[string]interface{}{
			"workspace_id": workspace.ID,
			"attempt_id":   attemptID,
			"command_kind": string(kind),
			"error":        "workspace_state_sync_rejected",
		})
		return &result
	}
	return nil
}

func workspaceAttemptCommandID(workspace improvement.AttemptWorkspace, kind transition.AttemptPhaseCommandKind, input map[string]interface{}, occurredAt time.Time) string {
	suffix := firstNonEmpty(
		strings.TrimSpace(stringValue(input["tool_call_id"])),
		strings.TrimSpace(stringValue(input["provider_ref"])),
		strconv.FormatInt(occurredAt.UnixNano(), 10),
	)
	return fmt.Sprintf("cmd-attempt:%s:%s:%s", strings.TrimSpace(workspace.AttemptID), string(kind), suffix)
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

func parseWorkspaceGitLogEntries(body string) []map[string]interface{} {
	lines := compactLines(body)
	entries := make([]map[string]interface{}, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 4)
		entry := map[string]interface{}{
			"raw": line,
		}
		if len(parts) > 0 {
			sha := strings.TrimSpace(parts[0])
			entry["sha"] = sha
			entry["short_sha"] = sha[:minInt(len(sha), 12)]
		}
		if len(parts) > 1 {
			entry["date"] = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			entry["author"] = strings.TrimSpace(parts[2])
		}
		if len(parts) > 3 {
			entry["subject"] = strings.TrimSpace(parts[3])
		}
		entries = append(entries, entry)
	}
	return entries
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

func shellQuoteJoin(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		quoted = append(quoted, shellQuote(value))
	}
	return strings.Join(quoted, " ")
}
