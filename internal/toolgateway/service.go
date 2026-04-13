package toolgateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/cluster"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Service struct {
	cfg         config.Config
	store       storepkg.Repository
	httpClient  *http.Client
	slackClient *slackapi.Client
	kubeClient  kubernetes.Interface
}

func NewService(cfg config.Config, store storepkg.Repository) *Service {
	var slackClient *slackapi.Client
	if strings.TrimSpace(cfg.SlackBotToken) != "" {
		slackClient = slackapi.New(cfg.SlackBotToken)
	}
	var kubeClient kubernetes.Interface
	if clientset, err := cluster.NewClientset(cfg); err == nil {
		kubeClient = clientset
	}
	return &Service{
		cfg:         cfg,
		store:       store,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		slackClient: slackClient,
		kubeClient:  kubeClient,
	}
}

func (s *Service) Execute(name string, input map[string]interface{}) storepkg.ToolResult {
	switch name {
	case "repo.context":
		return s.repoContext(input)
	case "knowledge.context":
		return s.knowledgeContext(input)
	case "sentry.lookup":
		return s.sentryLookup(input)
	case "kubernetes.inspect":
		return s.kubernetesInspect(input)
	case "slack.reply":
		return s.slackReply(input)
	case "github.create_pr":
		return s.githubCreatePR(input)
	case "github.repo_context":
		return s.githubRepoContext(input)
	case "github.repo_activity":
		return s.githubRepoActivity(input)
	case "cloudflare.inspect":
		return s.cloudflareInspect(input)
	default:
		return s.store.ExecuteTool(name, input)
	}
}

func (s *Service) repoContext(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	question := stringValue(input["question"])
	summary := fmt.Sprintf("Repo context for %s prepared.", repo)
	excerpt := ""
	if repo == "rsi-agent-platform" {
		if data, err := os.ReadFile("README.md"); err == nil {
			excerpt = truncate(string(data), 500)
			if excerpt != "" {
				summary = fmt.Sprintf("Loaded README context for %s.", repo)
			}
		}
	}
	output := map[string]interface{}{
		"repo":     repo,
		"question": question,
		"excerpt":  excerpt,
	}
	return s.result("repo.context", input, summary, output, nil)
}

func (s *Service) knowledgeContext(input map[string]interface{}) storepkg.ToolResult {
	topic := strings.ToLower(strings.TrimSpace(firstNonEmpty(stringValue(input["topic"]), stringValue(input["question"]))))
	scopeID := firstNonEmpty(stringValue(input["scope_id"]), stringValue(input["repo"]))
	entries := s.store.ListKnowledgeEntries()
	sort.Slice(entries, func(i, j int) bool {
		left := entries[i]
		right := entries[j]
		leftRank := 1
		if left.Status == knowledge.StatusCanonical || left.Tier == knowledge.TierCanonical {
			leftRank = 0
		}
		rightRank := 1
		if right.Status == knowledge.StatusCanonical || right.Tier == knowledge.TierCanonical {
			rightRank = 0
		}
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if left.Confidence != right.Confidence {
			return left.Confidence > right.Confidence
		}
		return left.UpdatedAt.After(right.UpdatedAt)
	})
	matches := make([]knowledge.Entry, 0)
	links := make([][]knowledge.EvidenceLink, 0)
	for _, item := range entries {
		if item.Status != knowledge.StatusCanonical && item.Tier != knowledge.TierWorking {
			continue
		}
		if scopeID != "" && item.ScopeID != "" && item.ScopeID != scopeID {
			continue
		}
		haystack := strings.ToLower(strings.Join([]string{item.Title, item.Summary, item.Body, string(item.Kind), string(item.ScopeType)}, " "))
		if topic != "" && !strings.Contains(haystack, topic) {
			continue
		}
		matches = append(matches, item)
		links = append(links, s.store.ListKnowledgeEvidenceLinks(item.ID))
		if len(matches) == 8 {
			break
		}
	}
	summary := fmt.Sprintf("Retrieved %d structured knowledge entries.", len(matches))
	return s.result("knowledge.context", input, summary, map[string]interface{}{
		"topic":    topic,
		"scope_id": scopeID,
		"entries":  matches,
		"links":    links,
	}, nil)
}

func (s *Service) sentryLookup(input map[string]interface{}) storepkg.ToolResult {
	service := firstNonEmpty(stringValue(input["service"]), "unknown-service")
	alert := stringValue(input["alert"])
	if strings.TrimSpace(s.cfg.SentryAuthToken) == "" || strings.TrimSpace(s.cfg.SentryOrganization) == "" {
		summary := fmt.Sprintf("Sentry lookup unavailable for %s: missing credentials.", service)
		return s.unavailableResult("sentry.lookup", input, "sentry", summary, map[string]interface{}{
			"service": service,
			"alert":   alert,
			"error":   "missing RSI_SENTRY_AUTH_TOKEN or RSI_SENTRY_ORGANIZATION",
		})
	}

	query := firstNonEmpty(stringValue(input["query"]), alert, service)
	values := url.Values{}
	values.Set("query", query)
	endpoint := fmt.Sprintf("%s/organizations/%s/issues/?%s", strings.TrimRight(s.cfg.SentryAPIBaseURL, "/"), s.cfg.SentryOrganization, values.Encode())
	var issues []map[string]interface{}
	if err := s.apiJSON(http.MethodGet, endpoint, nil, map[string]string{
		"Authorization": "Bearer " + s.cfg.SentryAuthToken,
	}, &issues); err != nil {
		return s.failedResult("sentry.lookup", input, "sentry", fmt.Sprintf("Sentry lookup failed: %v", err), map[string]interface{}{
			"service": service,
			"alert":   alert,
			"error":   err.Error(),
		})
	}
	summary := fmt.Sprintf("Sentry returned %d issues for query %q.", len(issues), query)
	return s.result("sentry.lookup", input, summary, map[string]interface{}{
		"service": service,
		"alert":   alert,
		"issues":  issues,
		"query":   query,
	}, nil)
}

func (s *Service) kubernetesInspect(input map[string]interface{}) storepkg.ToolResult {
	namespace := firstNonEmpty(stringValue(input["namespace"]), s.cfg.SandboxNamespace)
	target := firstNonEmpty(stringValue(input["target"]), stringValue(input["service"]))
	if s.kubeClient == nil {
		summary := fmt.Sprintf("Kubernetes inspection unavailable for %s/%s.", namespace, firstNonEmpty(target, "unknown"))
		return s.unavailableResult("kubernetes.inspect", input, "kubernetes", summary, map[string]interface{}{
			"namespace": namespace,
			"target":    target,
			"error":     "kubernetes client unavailable",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pods, err := s.kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return s.failedResult("kubernetes.inspect", input, "kubernetes", fmt.Sprintf("Kubernetes inspection failed: %v", err), map[string]interface{}{
			"namespace": namespace,
			"target":    target,
			"error":     err.Error(),
		})
	}
	eventsList, _ := s.kubeClient.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	matchedPods := make([]map[string]interface{}, 0)
	for _, pod := range pods.Items {
		if target != "" && !matchesKubernetesTarget(pod, target) {
			continue
		}
		matchedPods = append(matchedPods, map[string]interface{}{
			"name":   pod.Name,
			"phase":  string(pod.Status.Phase),
			"node":   pod.Spec.NodeName,
			"reason": pod.Status.Reason,
		})
	}
	matchedEvents := make([]map[string]interface{}, 0)
	for _, event := range eventsList.Items {
		if target != "" && !strings.Contains(strings.ToLower(event.InvolvedObject.Name), strings.ToLower(target)) {
			continue
		}
		matchedEvents = append(matchedEvents, map[string]interface{}{
			"name":    event.Name,
			"reason":  event.Reason,
			"message": event.Message,
			"type":    event.Type,
		})
	}
	summary := fmt.Sprintf("Kubernetes inspection found %d pods and %d events in %s for target %s.", len(matchedPods), len(matchedEvents), namespace, firstNonEmpty(target, "all"))
	return s.result("kubernetes.inspect", input, summary, map[string]interface{}{
		"namespace": namespace,
		"target":    target,
		"pods":      matchedPods,
		"events":    matchedEvents,
	}, nil)
}

func (s *Service) githubRepoContext(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	if strings.TrimSpace(s.cfg.GitHubToken) == "" {
		summary := fmt.Sprintf("GitHub repo context unavailable for %s: missing token.", repo)
		return s.unavailableResult("github.repo_context", input, "github", summary, map[string]interface{}{
			"repo":  repo,
			"error": "missing RSI_GITHUB_TOKEN",
		})
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), s.cfg.GitHubOwner, repo)
	var payload map[string]interface{}
	if err := s.apiJSON(http.MethodGet, endpoint, nil, map[string]string{
		"Authorization": "Bearer " + s.cfg.GitHubToken,
		"Accept":        "application/vnd.github+json",
	}, &payload); err != nil {
		return s.failedResult("github.repo_context", input, "github", fmt.Sprintf("GitHub repo context failed: %v", err), map[string]interface{}{
			"repo":  repo,
			"error": err.Error(),
		})
	}
	summary := fmt.Sprintf("GitHub repo context loaded for %s (default branch %s).", repo, stringValue(payload["default_branch"]))
	return s.result("github.repo_context", input, summary, payload, nil)
}

func (s *Service) githubRepoActivity(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	if strings.TrimSpace(s.cfg.GitHubToken) == "" {
		return s.unavailableResult("github.repo_activity", input, "github", fmt.Sprintf("GitHub repo activity unavailable for %s: missing token.", repo), map[string]interface{}{
			"repo":  repo,
			"error": "missing RSI_GITHUB_TOKEN",
		})
	}
	since, until, err := parseActivityWindow(input)
	if err != nil {
		return s.failedResult("github.repo_activity", input, "github", fmt.Sprintf("GitHub repo activity input invalid: %v", err), map[string]interface{}{
			"repo":  repo,
			"error": err.Error(),
		})
	}
	headers := map[string]string{
		"Authorization": "Bearer " + s.cfg.GitHubToken,
		"Accept":        "application/vnd.github+json",
	}

	commitValues := url.Values{}
	commitValues.Set("since", since.Format(time.RFC3339))
	commitValues.Set("until", until.Format(time.RFC3339))
	commitValues.Set("per_page", "25")
	commitEndpoint := fmt.Sprintf("%s/repos/%s/%s/commits?%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), s.cfg.GitHubOwner, repo, commitValues.Encode())
	var commitPayload []map[string]interface{}
	if err := s.apiJSON(http.MethodGet, commitEndpoint, nil, headers, &commitPayload); err != nil {
		return s.failedResult("github.repo_activity", input, "github", fmt.Sprintf("GitHub repo activity failed to load commits: %v", err), map[string]interface{}{
			"repo":  repo,
			"error": err.Error(),
		})
	}

	pullValues := url.Values{}
	pullValues.Set("state", "all")
	pullValues.Set("sort", "updated")
	pullValues.Set("direction", "desc")
	pullValues.Set("per_page", "50")
	pullEndpoint := fmt.Sprintf("%s/repos/%s/%s/pulls?%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), s.cfg.GitHubOwner, repo, pullValues.Encode())
	var pullPayload []map[string]interface{}
	if err := s.apiJSON(http.MethodGet, pullEndpoint, nil, headers, &pullPayload); err != nil {
		return s.failedResult("github.repo_activity", input, "github", fmt.Sprintf("GitHub repo activity failed to load pull requests: %v", err), map[string]interface{}{
			"repo":  repo,
			"error": err.Error(),
		})
	}

	commits := make([]map[string]interface{}, 0, len(commitPayload))
	for _, item := range commitPayload {
		commits = append(commits, mapGitHubCommit(item))
	}
	mergedPRs := make([]map[string]interface{}, 0)
	openedPRs := make([]map[string]interface{}, 0)
	for _, item := range pullPayload {
		mapped := mapGitHubPull(item)
		if mergedAt := parseGitHubTimestamp(stringValue(item["merged_at"])); mergedAt != nil && !mergedAt.Before(since) && !mergedAt.After(until) {
			mergedPRs = append(mergedPRs, mapped)
		}
		if createdAt := parseGitHubTimestamp(stringValue(item["created_at"])); createdAt != nil && !createdAt.Before(since) && !createdAt.After(until) {
			openedPRs = append(openedPRs, mapped)
		}
	}
	summary := fmt.Sprintf("GitHub activity for %s from %s to %s includes %d commits, %d merged PRs, and %d opened PRs.", repo, since.Format("2006-01-02"), until.Format("2006-01-02"), len(commits), len(mergedPRs), len(openedPRs))
	return s.result("github.repo_activity", input, summary, map[string]interface{}{
		"repo":                 repo,
		"since":                since.Format(time.RFC3339),
		"until":                until.Format(time.RFC3339),
		"commits":              commits,
		"merged_pull_requests": mergedPRs,
		"opened_pull_requests": openedPRs,
		"summary":              summary,
	}, nil)
}

func (s *Service) githubCreatePR(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	head := firstNonEmpty(stringValue(input["branch_name"]), stringValue(input["head"]))
	base := firstNonEmpty(stringValue(input["base_ref"]), "main")
	title := firstNonEmpty(stringValue(input["title"]), fmt.Sprintf("RSI proposal for %s", repo))
	body := firstNonEmpty(stringValue(input["body"]), "Automated draft PR from RSI platform.")
	if strings.TrimSpace(s.cfg.GitHubToken) == "" {
		return s.unavailableResult("github.create_pr", input, "github", "GitHub token not configured; refusing draft PR execution.", map[string]interface{}{
			"repo":  repo,
			"head":  head,
			"base":  base,
			"error": "missing RSI_GITHUB_TOKEN",
		})
	}
	requestBody := map[string]interface{}{
		"title": title,
		"head":  head,
		"base":  base,
		"body":  body,
		"draft": true,
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s/pulls", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), s.cfg.GitHubOwner, repo)
	var response map[string]interface{}
	if err := s.apiJSON(http.MethodPost, endpoint, requestBody, map[string]string{
		"Authorization": "Bearer " + s.cfg.GitHubToken,
		"Accept":        "application/vnd.github+json",
	}, &response); err != nil {
		return s.failedResult("github.create_pr", input, "github", fmt.Sprintf("GitHub PR creation failed: %v", err), map[string]interface{}{
			"repo":  repo,
			"head":  head,
			"base":  base,
			"error": err.Error(),
		})
	}
	summary := fmt.Sprintf("Draft PR opened for %s:%s.", repo, head)
	return s.result("github.create_pr", input, summary, map[string]interface{}{
		"repo":     repo,
		"head":     head,
		"base":     base,
		"pr_url":   stringValue(response["html_url"]),
		"number":   response["number"],
		"response": response,
	}, nil)
}

func (s *Service) cloudflareInspect(input map[string]interface{}) storepkg.ToolResult {
	resource := firstNonEmpty(stringValue(input["resource"]), "zones")
	if strings.TrimSpace(s.cfg.CloudflareAPIToken) == "" {
		summary := fmt.Sprintf("Cloudflare inspection unavailable for %s: missing API token.", resource)
		return s.unavailableResult("cloudflare.inspect", input, "cloudflare", summary, map[string]interface{}{
			"resource": resource,
			"error":    "missing RSI_CLOUDFLARE_API_TOKEN",
		})
	}

	var endpoint string
	switch resource {
	case "zone", "zones":
		if strings.TrimSpace(s.cfg.CloudflareZoneID) != "" {
			endpoint = fmt.Sprintf("%s/zones/%s", strings.TrimRight(s.cfg.CloudflareAPIBaseURL, "/"), s.cfg.CloudflareZoneID)
		} else {
			endpoint = fmt.Sprintf("%s/zones", strings.TrimRight(s.cfg.CloudflareAPIBaseURL, "/"))
		}
	default:
		if strings.TrimSpace(s.cfg.CloudflareAccountID) == "" {
			return s.unavailableResult("cloudflare.inspect", input, "cloudflare", "Cloudflare account id required for non-zone inspection.", map[string]interface{}{
				"resource": resource,
				"error":    "missing RSI_CLOUDFLARE_ACCOUNT_ID",
			})
		}
		endpoint = fmt.Sprintf("%s/accounts/%s/%s", strings.TrimRight(s.cfg.CloudflareAPIBaseURL, "/"), s.cfg.CloudflareAccountID, path.Clean("/"+resource))
	}
	var response map[string]interface{}
	if err := s.apiJSON(http.MethodGet, endpoint, nil, map[string]string{
		"Authorization": "Bearer " + s.cfg.CloudflareAPIToken,
		"Content-Type":  "application/json",
	}, &response); err != nil {
		return s.failedResult("cloudflare.inspect", input, "cloudflare", fmt.Sprintf("Cloudflare inspection failed: %v", err), map[string]interface{}{
			"resource": resource,
			"error":    err.Error(),
		})
	}
	summary := fmt.Sprintf("Cloudflare inspection loaded for %s.", resource)
	return s.result("cloudflare.inspect", input, summary, response, nil)
}

func (s *Service) slackReply(input map[string]interface{}) storepkg.ToolResult {
	channelID := stringValue(input["channel_id"])
	threadTS := stringValue(input["thread_ts"])
	body := stringValue(input["body"])
	dryRun := boolValue(input["dry_run"])
	output := map[string]interface{}{
		"channel_id": channelID,
		"thread_ts":  threadTS,
		"posted":     false,
	}
	summary := "Slack reply drafted."
	if channelID == "" || body == "" {
		return s.failedResult("slack.reply", input, "slack", "Slack reply missing channel or body.", map[string]interface{}{
			"posted": false,
		})
	}
	if !dryRun && s.slackClient == nil {
		summary = "Slack reply unavailable: bot token is not configured."
		return s.unavailableResult("slack.reply", input, "slack", summary, map[string]interface{}{
			"posted": false,
			"error":  "missing RSI_SLACK_BOT_TOKEN",
		})
	}
	if !dryRun && s.slackClient != nil {
		params := slackapi.PostMessageParameters{ThreadTimestamp: threadTS}
		_, ts, err := s.slackClient.PostMessage(channelID, slackapi.MsgOptionText(body, false), slackapi.MsgOptionPostMessageParameters(params))
		if err == nil {
			output["posted"] = true
			output["ts"] = ts
			summary = fmt.Sprintf("Slack reply posted to %s.", channelID)
			return s.result("slack.reply", input, summary, output, nil)
		}
		output["error"] = err.Error()
		summary = fmt.Sprintf("Slack reply failed: %v", err)
		return s.failedResult("slack.reply", input, "slack", summary, output)
	}
	if dryRun {
		summary = "Slack reply dry-run generated."
	}
	return s.result("slack.reply", input, summary, output, nil)
}

func (s *Service) result(name string, input map[string]interface{}, summary string, output map[string]interface{}, refs []string) storepkg.ToolResult {
	callID := fmt.Sprintf("%s-%d", sanitizeToolName(name), time.Now().UTC().UnixNano())
	return storepkg.ToolResult{
		Name:            name,
		ToolCallID:      callID,
		Approved:        true,
		ApprovalState:   "not_required",
		Status:          "ok",
		Available:       true,
		Provider:        providerForToolName(name),
		ProviderRef:     providerRefForTool(name, output),
		ExecutedAt:      time.Now().UTC(),
		Input:           input,
		Output:          output,
		Summary:         summary,
		RawArtifactRefs: refs,
		Metadata: map[string]interface{}{
			"tool_name": name,
		},
	}
}

func (s *Service) unavailableResult(name string, input map[string]interface{}, provider string, summary string, output map[string]interface{}) storepkg.ToolResult {
	callID := fmt.Sprintf("%s-%d", sanitizeToolName(name), time.Now().UTC().UnixNano())
	return storepkg.ToolResult{
		Name:          name,
		ToolCallID:    callID,
		Approved:      false,
		ApprovalState: "provider_unavailable",
		Status:        "blocked",
		Available:     false,
		Provider:      provider,
		ExecutedAt:    time.Now().UTC(),
		Input:         input,
		Output:        output,
		Summary:       summary,
		Metadata: map[string]interface{}{
			"tool_name": name,
			"provider":  provider,
		},
	}
}

func (s *Service) failedResult(name string, input map[string]interface{}, provider string, summary string, output map[string]interface{}) storepkg.ToolResult {
	callID := fmt.Sprintf("%s-%d", sanitizeToolName(name), time.Now().UTC().UnixNano())
	return storepkg.ToolResult{
		Name:          name,
		ToolCallID:    callID,
		Approved:      true,
		ApprovalState: "not_required",
		Status:        "failed",
		Available:     true,
		Provider:      provider,
		ProviderRef:   providerRefForTool(name, output),
		ExecutedAt:    time.Now().UTC(),
		Input:         input,
		Output:        output,
		Summary:       summary,
		Metadata: map[string]interface{}{
			"tool_name": name,
			"provider":  provider,
		},
	}
}

func (s *Service) apiJSON(method string, endpoint string, body interface{}, headers map[string]string, out interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		return err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}

func matchesKubernetesTarget(pod corev1.Pod, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return true
	}
	if strings.Contains(strings.ToLower(pod.Name), target) {
		return true
	}
	for key, value := range pod.Labels {
		if strings.Contains(strings.ToLower(key), target) || strings.Contains(strings.ToLower(value), target) {
			return true
		}
	}
	return false
}

func sanitizeToolName(name string) string {
	return strings.NewReplacer(".", "-", "_", "-").Replace(strings.TrimSpace(name))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func boolValue(value interface{}) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(typed, "true")
	default:
		return false
	}
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit]
}

func providerForToolName(name string) string {
	switch {
	case strings.HasPrefix(name, "slack."):
		return "slack"
	case strings.HasPrefix(name, "github."):
		return "github"
	case strings.HasPrefix(name, "sentry."):
		return "sentry"
	case strings.HasPrefix(name, "kubernetes."):
		return "kubernetes"
	case strings.HasPrefix(name, "cloudflare."):
		return "cloudflare"
	case strings.HasPrefix(name, "knowledge."):
		return "knowledge"
	default:
		return "internal"
	}
}

func providerRefForTool(name string, output map[string]interface{}) string {
	switch name {
	case "slack.reply":
		return firstNonEmpty(stringValue(output["ts"]), stringValue(output["thread_ts"]))
	case "github.create_pr":
		return stringValue(output["pr_url"])
	case "github.repo_context":
		return stringValue(output["html_url"])
	case "github.repo_activity":
		return stringValue(output["repo"])
	case "sentry.lookup":
		return stringValue(output["query"])
	case "kubernetes.inspect":
		return stringValue(output["target"])
	case "cloudflare.inspect":
		return stringValue(output["resource"])
	case "knowledge.context":
		return stringValue(output["topic"])
	default:
		return stringValue(output["repo"])
	}
}

func parseActivityWindow(input map[string]interface{}) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	since := strings.TrimSpace(stringValue(input["since"]))
	until := strings.TrimSpace(stringValue(input["until"]))
	if since == "" {
		since = now.Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	}
	if until == "" {
		until = now.Format(time.RFC3339)
	}
	start, err := time.Parse(time.RFC3339, since)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid since timestamp %q", since)
	}
	end, err := time.Parse(time.RFC3339, until)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid until timestamp %q", until)
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("until must be after since")
	}
	return start, end, nil
}

func mapGitHubCommit(item map[string]interface{}) map[string]interface{} {
	commit, _ := item["commit"].(map[string]interface{})
	author, _ := commit["author"].(map[string]interface{})
	return map[string]interface{}{
		"sha":          stringValue(item["sha"]),
		"message":      stringValue(commit["message"]),
		"author":       firstNonEmpty(stringValue(author["name"]), stringValue(item["author"])),
		"committed_at": stringValue(author["date"]),
		"url":          stringValue(item["html_url"]),
	}
}

func mapGitHubPull(item map[string]interface{}) map[string]interface{} {
	user, _ := item["user"].(map[string]interface{})
	return map[string]interface{}{
		"number":     item["number"],
		"title":      stringValue(item["title"]),
		"author":     firstNonEmpty(stringValue(user["login"]), stringValue(user["name"])),
		"state":      stringValue(item["state"]),
		"created_at": stringValue(item["created_at"]),
		"merged_at":  stringValue(item["merged_at"]),
		"url":        stringValue(item["html_url"]),
	}
}

func parseGitHubTimestamp(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}
