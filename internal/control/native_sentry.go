package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

const sentryCLITimeout = 90 * time.Second

type sentryCommandRunnerFunc func(ctx context.Context, args []string, env []string) ([]byte, []byte, error)

var sentryCommandRunner sentryCommandRunnerFunc = runSentryCLICommand

func executeSentryNativeToolAction(ctx context.Context, cfg config.Config, input nativeToolActionRequest) (any, string, string, string, map[string]any, int, error) {
	if strings.TrimSpace(cfg.SentryAuthToken) == "" {
		return nil, "", "", "", nativeToolNoopEffect("missing_sentry_auth_token"), http.StatusFailedDependency, errors.New("RSI_SENTRY_AUTH_TOKEN or SENTRY_AUTH_TOKEN is required for native Sentry tools")
	}
	args, targetRef, err := sentryCLIArgs(cfg, input)
	if err != nil {
		return nil, "", "", "", nativeToolNoopEffect("invalid_sentry_arguments"), http.StatusBadRequest, err
	}
	env, cleanup, err := sentryCLIEnv(cfg)
	if err != nil {
		return nil, "", "", "", nativeToolNoopEffect("sentry_env_failed"), http.StatusInternalServerError, err
	}
	defer cleanup()

	cmdCtx, cancel := context.WithTimeout(ctx, sentryCLITimeout)
	defer cancel()
	stdout, stderr, runErr := sentryCommandRunner(cmdCtx, args, env)
	if cmdCtx.Err() == context.DeadlineExceeded {
		return nil, "", targetRef, "", nativeToolNoopEffect("sentry_cli_timeout"), http.StatusGatewayTimeout, errors.New("sentry CLI command timed out")
	}
	output := sentryParseOutput(stdout, cfg.SentryAuthToken)
	if runErr != nil {
		redactedErr := sentryRedact(strings.TrimSpace(stderrExcerpt(stderr, cfg.SentryAuthToken)), cfg.SentryAuthToken)
		if redactedErr == "" {
			redactedErr = sentryRedact(runErr.Error(), cfg.SentryAuthToken)
		}
		payload := map[string]any{
			"command": args,
			"stderr":  redactedErr,
			"output":  output,
		}
		if isSentryCLIMissing(runErr) {
			return payload, "", targetRef, "", nativeToolNoopEffect("sentry_cli_missing"), http.StatusFailedDependency, errors.New("sentry CLI binary is not installed in the control-plane image")
		}
		return payload, "", targetRef, "", nativeToolNoopEffect("sentry_cli_failed"), http.StatusBadGateway, fmt.Errorf("sentry CLI command failed: %s", redactedErr)
	}
	response := map[string]any{
		"command": args,
		"output":  output,
	}
	return response, sentryResponseSummary(input.Operation, output), targetRef, "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
}

func sentryCLIArgs(cfg config.Config, input nativeToolActionRequest) ([]string, string, error) {
	args := make([]string, 0, 12)
	targetRef := input.TargetRef
	org := firstNonEmpty(stringArg(input.Arguments, "org"), cfg.SentryOrganization)
	switch input.Operation {
	case "orgs_list":
		args = append(args, "org", "list")
		if limit := sentryLimitArg(input.Arguments, "limit", 25, 100); limit > 0 {
			args = append(args, "--limit", fmt.Sprintf("%d", limit))
		}
		targetRef = firstNonEmpty(targetRef, "sentry:orgs")
	case "projects_list":
		selector := sentryProjectSelector(cfg, input.Arguments, true)
		if selector == "" {
			return nil, "", errors.New("rsi_sentry.projects_list requires org, project_ref, or RSI_SENTRY_ORGANIZATION")
		}
		args = append(args, "project", "list", selector)
		if platform := stringArg(input.Arguments, "platform"); platform != "" {
			args = append(args, "--platform", platform)
		}
		if limit := sentryLimitArg(input.Arguments, "limit", 25, 100); limit > 0 {
			args = append(args, "--limit", fmt.Sprintf("%d", limit))
		}
		if cursor := stringArg(input.Arguments, "cursor"); cursor != "" {
			args = append(args, "--cursor", cursor)
		}
		targetRef = firstNonEmpty(targetRef, "sentry:projects:"+selector)
	case "issues_list":
		selector := sentryProjectSelector(cfg, input.Arguments, true)
		if selector == "" {
			return nil, "", errors.New("rsi_sentry.issues_list requires project_ref, org, or RSI_SENTRY_ORGANIZATION")
		}
		args = append(args, "issue", "list", selector)
		if query := stringArg(input.Arguments, "query"); query != "" {
			args = append(args, "--query", query)
		}
		if limit := sentryLimitArg(input.Arguments, "limit", 25, 100); limit > 0 {
			args = append(args, "--limit", fmt.Sprintf("%d", limit))
		}
		if sort := sentrySortArg(input.Arguments); sort != "" {
			args = append(args, "--sort", sort)
		}
		if period := stringArg(input.Arguments, "period"); period != "" {
			args = append(args, "--period", period)
		}
		if cursor := stringArg(input.Arguments, "cursor"); cursor != "" {
			args = append(args, "--cursor", cursor)
		}
		targetRef = firstNonEmpty(targetRef, "sentry:issues:"+selector)
	case "issue_view":
		issue := sentryIssueArg(input)
		if issue == "" {
			return nil, "", errors.New("rsi_sentry.issue_view requires issue")
		}
		args = append(args, "issue", "view", issue)
		if spans := stringArg(input.Arguments, "spans"); spans != "" {
			args = append(args, "--spans", spans)
		}
		targetRef = firstNonEmpty(targetRef, "sentry:issue:"+issue)
	case "issue_events":
		issue := sentryIssueArg(input)
		if issue == "" {
			return nil, "", errors.New("rsi_sentry.issue_events requires issue")
		}
		args = append(args, "issue", "events", issue)
		if limit := sentryLimitArg(input.Arguments, "limit", 25, 100); limit > 0 {
			args = append(args, "--limit", fmt.Sprintf("%d", limit))
		}
		if query := stringArg(input.Arguments, "query"); query != "" {
			args = append(args, "--query", query)
		}
		if period := stringArg(input.Arguments, "period"); period != "" {
			args = append(args, "--period", period)
		}
		if cursor := stringArg(input.Arguments, "cursor"); cursor != "" {
			args = append(args, "--cursor", cursor)
		}
		if boolArg(input.Arguments, "full", false) {
			args = append(args, "--full")
		}
		targetRef = firstNonEmpty(targetRef, "sentry:issue:"+issue+":events")
	case "issue_explain":
		issue := sentryIssueArg(input)
		if issue == "" {
			return nil, "", errors.New("rsi_sentry.issue_explain requires issue")
		}
		args = append(args, "issue", "explain", issue)
		if boolArg(input.Arguments, "force", false) {
			args = append(args, "--force")
		}
		targetRef = firstNonEmpty(targetRef, "sentry:issue:"+issue+":explain")
	case "issue_plan":
		issue := sentryIssueArg(input)
		if issue == "" {
			return nil, "", errors.New("rsi_sentry.issue_plan requires issue")
		}
		args = append(args, "issue", "plan", issue)
		if cause := stringArg(input.Arguments, "cause"); cause != "" {
			args = append(args, "--cause", cause)
		}
		if boolArg(input.Arguments, "force", false) {
			args = append(args, "--force")
		}
		targetRef = firstNonEmpty(targetRef, "sentry:issue:"+issue+":plan")
	case "releases_list":
		selector := sentryProjectSelector(cfg, input.Arguments, true)
		if selector == "" && org != "" {
			selector = org + "/"
		}
		if selector == "" {
			return nil, "", errors.New("rsi_sentry.releases_list requires project_ref, org, or RSI_SENTRY_ORGANIZATION")
		}
		args = append(args, "release", "list", selector)
		if limit := sentryLimitArg(input.Arguments, "limit", 25, 100); limit > 0 {
			args = append(args, "--limit", fmt.Sprintf("%d", limit))
		}
		if cursor := stringArg(input.Arguments, "cursor"); cursor != "" {
			args = append(args, "--cursor", cursor)
		}
		targetRef = firstNonEmpty(targetRef, "sentry:releases:"+selector)
	default:
		return nil, "", fmt.Errorf("native Sentry operation %s is registered but not implemented", input.Operation)
	}
	if boolArg(input.Arguments, "fresh", false) {
		args = append(args, "--fresh")
	}
	args = append(args, "--json")
	return args, targetRef, nil
}

func sentryCLIEnv(cfg config.Config) ([]string, func(), error) {
	configDir, err := os.MkdirTemp("", "rsi-sentry-cli-*")
	if err != nil {
		return nil, func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(configDir) }
	env := append([]string{}, os.Environ()...)
	env = append(env,
		"SENTRY_AUTH_TOKEN="+strings.TrimSpace(cfg.SentryAuthToken),
		"SENTRY_FORCE_ENV_TOKEN=1",
		"SENTRY_OUTPUT_FORMAT=json",
		"SENTRY_PLAIN_OUTPUT=1",
		"SENTRY_CLI_NO_TELEMETRY=1",
		"SENTRY_CLI_NO_UPDATE_CHECK=1",
		"SENTRY_NO_CACHE=1",
		"SENTRY_CONFIG_DIR="+configDir,
	)
	if org := strings.TrimSpace(cfg.SentryOrganization); org != "" {
		env = append(env, "SENTRY_ORG="+org)
	}
	if baseURL := strings.TrimSpace(cfg.SentryAPIBaseURL); baseURL != "" {
		baseURL = strings.TrimRight(baseURL, "/")
		baseURL = strings.TrimSuffix(baseURL, "/api/0")
		if baseURL != "" && baseURL != "https://sentry.io" {
			env = append(env, "SENTRY_HOST="+baseURL)
		}
	}
	return env, cleanup, nil
}

func runSentryCLICommand(ctx context.Context, args []string, env []string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, "sentry", args...)
	cmd.Env = env
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func sentryParseOutput(stdout []byte, token string) any {
	trimmed := bytes.TrimSpace(stdout)
	if len(trimmed) == 0 {
		return map[string]any{}
	}
	var parsed any
	if err := json.Unmarshal(trimmed, &parsed); err == nil {
		return sentryRedactJSON(parsed, token)
	}
	return map[string]any{"raw": sentryRedact(string(trimmed), token)}
}

func sentryRedactJSON(value any, token string) any {
	data, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var redacted any
	if err := json.Unmarshal([]byte(sentryRedact(string(data), token)), &redacted); err != nil {
		return value
	}
	return redacted
}

func sentryRedact(value string, token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return value
	}
	return strings.ReplaceAll(value, token, "[REDACTED_SENTRY_TOKEN]")
}

func stderrExcerpt(stderr []byte, token string) string {
	text := sentryRedact(string(stderr), token)
	const max = 4096
	if len(text) <= max {
		return text
	}
	return text[:max] + "...[truncated]"
}

func sentryProjectSelector(cfg config.Config, args map[string]any, allowOrgWide bool) string {
	ref := firstNonEmpty(stringArg(args, "project_ref"), stringArg(args, "project"), stringArg(args, "project_slug"))
	if ref != "" {
		if strings.Contains(ref, "/") {
			return ref
		}
		if org := firstNonEmpty(stringArg(args, "org"), cfg.SentryOrganization); org != "" {
			return org + "/" + ref
		}
		return ref
	}
	org := firstNonEmpty(stringArg(args, "org"), cfg.SentryOrganization)
	if org != "" && allowOrgWide {
		return org + "/"
	}
	return ""
}

func sentryIssueArg(input nativeToolActionRequest) string {
	return firstNonEmpty(
		stringArg(input.Arguments, "issue"),
		stringArg(input.Arguments, "issue_ref"),
		stringArg(input.Arguments, "short_id"),
		input.TargetRef,
	)
}

func sentryLimitArg(args map[string]any, key string, fallback int, max int) int {
	limit := intArg(args, key, fallback)
	if limit <= 0 {
		return fallback
	}
	if limit > max {
		return max
	}
	return limit
}

func sentrySortArg(args map[string]any) string {
	sort := strings.TrimSpace(stringArg(args, "sort"))
	switch sort {
	case "", "date", "new", "freq", "user":
		return sort
	default:
		return ""
	}
}

func sentryResponseSummary(operation string, output any) string {
	count := sentryOutputCount(output)
	switch operation {
	case "orgs_list":
		return fmt.Sprintf("loaded %d Sentry organization(s)", count)
	case "projects_list":
		return fmt.Sprintf("loaded %d Sentry project(s)", count)
	case "issues_list":
		return fmt.Sprintf("loaded %d Sentry issue(s)", count)
	case "issue_events":
		return fmt.Sprintf("loaded %d Sentry issue event(s)", count)
	case "issue_view":
		return "loaded Sentry issue"
	case "issue_explain":
		return "loaded Sentry issue root-cause explanation"
	case "issue_plan":
		return "loaded Sentry issue remediation plan"
	case "releases_list":
		return fmt.Sprintf("loaded %d Sentry release(s)", count)
	default:
		return "loaded Sentry data"
	}
}

func sentryOutputCount(output any) int {
	switch typed := output.(type) {
	case []any:
		return len(typed)
	case map[string]any:
		if items, ok := typed["items"].([]any); ok {
			return len(items)
		}
		if results, ok := typed["results"].([]any); ok {
			return len(results)
		}
		if data, ok := typed["data"].([]any); ok {
			return len(data)
		}
	}
	return 0
}

func nativeToolNoopEffect(reason string) map[string]any {
	return map[string]any{"status": "not_applicable", "reason": reason}
}

func isSentryCLIMissing(err error) bool {
	var pathErr *exec.Error
	return errors.As(err, &pathErr) && errors.Is(pathErr.Err, exec.ErrNotFound)
}
