package toolgateway

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
	"unicode"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/cluster"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/githubapp"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
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
	launcher    sandbox.Launcher
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
	var launcher sandbox.Launcher
	if item, err := sandbox.NewLauncher(cfg); err == nil {
		launcher = item
	}
	return &Service{
		cfg:         cfg,
		store:       store,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		slackClient: slackClient,
		kubeClient:  kubeClient,
		launcher:    launcher,
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
	case "kubernetes.logs":
		return s.kubernetesLogs(input)
	case "kubernetes.events":
		return s.kubernetesEvents(input)
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
	case "rsi.trace_context":
		return s.rsiTraceContext(input)
	case "rsi.workflow_context":
		return s.rsiWorkflowContext(input)
	case "rsi.action_chain":
		return s.rsiActionChain(input)
	case "rsi.runner_execution":
		return s.rsiRunnerExecution(input)
	case "rsi.runtime_config":
		return s.rsiRuntimeConfig(input)
	case "rsi.runtime_health":
		return s.rsiRuntimeHealth(input)
	case "rsi.proposal_memory":
		return s.rsiProposalMemory(input)
	case "rsi.candidate_context":
		return s.rsiCandidateContext(input)
	case "rsi.attempt_context":
		return s.rsiAttemptContext(input)
	case "workspace.list_files":
		return s.workspaceListFiles(input)
	case "workspace.read_file":
		return s.workspaceReadFile(input)
	case "workspace.search":
		return s.workspaceSearch(input)
	case "workspace.write_file":
		return s.workspaceWriteFile(input)
	case "workspace.apply_patch":
		return s.workspaceApplyPatch(input)
	case "workspace.git_status":
		return s.workspaceGitStatus(input)
	case "workspace.git_diff":
		return s.workspaceGitDiff(input)
	case "workspace.run_validation":
		return s.workspaceRunValidation(input)
	default:
		return s.unavailableResult(name, input, "tool-gateway", fmt.Sprintf("Tool %s is not registered in the governed tool gateway.", strings.TrimSpace(name)), map[string]interface{}{
			"tool_name": strings.TrimSpace(name),
			"error":     "unknown_tool",
		})
	}
}

func (s *Service) repoContext(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	question := stringValue(input["question"])
	repoName := s.cfg.GitHubRepoName(repo)
	owner := s.cfg.GitHubRepoOwner(repo)
	summary := fmt.Sprintf("Repo context for %s prepared.", repoName)
	output := map[string]interface{}{
		"repo":     repoName,
		"owner":    owner,
		"question": question,
	}
	token, err := s.githubInstallationToken(repo)
	if err == nil {
		repoMeta, metaErr := s.githubRepoMetadata(owner, repoName, token)
		if metaErr == nil {
			output["default_branch"] = stringValue(repoMeta["default_branch"])
			output["html_url"] = stringValue(repoMeta["html_url"])
			output["description"] = stringValue(repoMeta["description"])
		}
		searchTerms := repoContextSearchTerms(question)
		if len(searchTerms) == 0 {
			searchTerms = []string{repoName}
		}
		output["search_terms"] = searchTerms
		matches, searchErr := s.githubRepoSearchContext(owner, repoName, token, searchTerms, stringValue(output["default_branch"]))
		if searchErr == nil {
			output["matches"] = matches
			if len(matches) > 0 {
				summary = fmt.Sprintf("GitHub-backed repo context loaded for %s/%s with %d relevant code match(es).", owner, repoName, len(matches))
			} else {
				summary = fmt.Sprintf("GitHub-backed repo context loaded for %s/%s with no direct code matches for %q.", owner, repoName, strings.TrimSpace(question))
			}
			return s.result("repo.context", input, summary, output, nil)
		}
		output["search_error"] = searchErr.Error()
		if metaErr == nil {
			summary = fmt.Sprintf("GitHub repo metadata loaded for %s/%s, but code search failed: %v", owner, repoName, searchErr)
			return s.result("repo.context", input, summary, output, nil)
		}
		return s.failedResult("repo.context", input, "github", fmt.Sprintf("Repo context failed: %v", searchErr), output)
	}
	return s.unavailableResult("repo.context", input, "github", fmt.Sprintf("Repo context unavailable for %s/%s: missing app authentication.", owner, repoName), output)
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

func (s *Service) kubernetesLogs(input map[string]interface{}) storepkg.ToolResult {
	namespace := firstNonEmpty(stringValue(input["namespace"]), s.cfg.SandboxNamespace)
	target := firstNonEmpty(stringValue(input["target"]), stringValue(input["pod_name"]), stringValue(input["service"]))
	if s.kubeClient == nil {
		summary := fmt.Sprintf("Kubernetes logs unavailable for %s/%s.", namespace, firstNonEmpty(target, "unknown"))
		return s.unavailableResult("kubernetes.logs", input, "kubernetes", summary, map[string]interface{}{
			"namespace": namespace,
			"target":    target,
			"error":     "kubernetes client unavailable",
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pods, err := s.kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return s.failedResult("kubernetes.logs", input, "kubernetes", fmt.Sprintf("Kubernetes logs failed: %v", err), map[string]interface{}{
			"namespace": namespace,
			"target":    target,
			"error":     err.Error(),
		})
	}
	logs := make([]map[string]interface{}, 0)
	for _, pod := range pods.Items {
		if target != "" && !matchesKubernetesTarget(pod, target) && !strings.EqualFold(pod.Name, target) {
			continue
		}
		var container string
		if len(pod.Spec.Containers) > 0 {
			container = pod.Spec.Containers[0].Name
		}
		if explicit := strings.TrimSpace(stringValue(input["container"])); explicit != "" {
			container = explicit
		}
		tailLines := int64(80)
		req := s.kubeClient.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: container,
			TailLines: &tailLines,
		})
		stream, err := req.Stream(ctx)
		if err != nil {
			logs = append(logs, map[string]interface{}{
				"pod_name":  pod.Name,
				"container": container,
				"error":     err.Error(),
			})
			continue
		}
		data, readErr := io.ReadAll(stream)
		_ = stream.Close()
		if readErr != nil {
			logs = append(logs, map[string]interface{}{
				"pod_name":  pod.Name,
				"container": container,
				"error":     readErr.Error(),
			})
			continue
		}
		logs = append(logs, map[string]interface{}{
			"pod_name":  pod.Name,
			"container": container,
			"log_tail":  truncate(string(data), 4000),
		})
		if len(logs) == 3 {
			break
		}
	}
	summary := fmt.Sprintf("Kubernetes logs loaded for %d pod(s) in %s.", len(logs), namespace)
	return s.result("kubernetes.logs", input, summary, map[string]interface{}{
		"namespace": namespace,
		"target":    target,
		"logs":      logs,
	}, nil)
}

func (s *Service) kubernetesEvents(input map[string]interface{}) storepkg.ToolResult {
	namespace := firstNonEmpty(stringValue(input["namespace"]), s.cfg.SandboxNamespace)
	target := firstNonEmpty(stringValue(input["target"]), stringValue(input["service"]))
	if s.kubeClient == nil {
		summary := fmt.Sprintf("Kubernetes events unavailable for %s/%s.", namespace, firstNonEmpty(target, "unknown"))
		return s.unavailableResult("kubernetes.events", input, "kubernetes", summary, map[string]interface{}{
			"namespace": namespace,
			"target":    target,
			"error":     "kubernetes client unavailable",
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	eventsList, err := s.kubeClient.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return s.failedResult("kubernetes.events", input, "kubernetes", fmt.Sprintf("Kubernetes events failed: %v", err), map[string]interface{}{
			"namespace": namespace,
			"target":    target,
			"error":     err.Error(),
		})
	}
	items := make([]map[string]interface{}, 0)
	for _, event := range eventsList.Items {
		if target != "" && !strings.Contains(strings.ToLower(event.InvolvedObject.Name), strings.ToLower(target)) {
			continue
		}
		items = append(items, map[string]interface{}{
			"name":            event.Name,
			"involved_object": event.InvolvedObject.Name,
			"reason":          event.Reason,
			"type":            event.Type,
			"message":         event.Message,
		})
		if len(items) == 20 {
			break
		}
	}
	summary := fmt.Sprintf("Kubernetes events loaded for %d event(s) in %s.", len(items), namespace)
	return s.result("kubernetes.events", input, summary, map[string]interface{}{
		"namespace": namespace,
		"target":    target,
		"events":    items,
	}, nil)
}

func (s *Service) githubRepoContext(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	owner := s.cfg.GitHubRepoOwner(repo)
	repoName := s.cfg.GitHubRepoName(repo)
	token, err := s.githubInstallationToken(repo)
	if err != nil {
		summary := fmt.Sprintf("GitHub repo context unavailable for %s/%s: missing app authentication.", owner, repoName)
		return s.unavailableResult("github.repo_context", input, "github", summary, map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"error": err.Error(),
		})
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), owner, repoName)
	var payload map[string]interface{}
	if err := s.apiJSON(http.MethodGet, endpoint, nil, map[string]string{
		"Authorization": "Bearer " + token,
		"Accept":        "application/vnd.github+json",
	}, &payload); err != nil {
		return s.failedResult("github.repo_context", input, "github", fmt.Sprintf("GitHub repo context failed: %v", err), map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"error": err.Error(),
		})
	}
	summary := fmt.Sprintf("GitHub repo context loaded for %s/%s (default branch %s).", owner, repoName, stringValue(payload["default_branch"]))
	return s.result("github.repo_context", input, summary, payload, nil)
}

func (s *Service) githubRepoMetadata(owner string, repo string, token string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), owner, repo)
	payload := map[string]interface{}{}
	if err := s.apiJSON(http.MethodGet, endpoint, nil, map[string]string{
		"Authorization": "Bearer " + token,
		"Accept":        "application/vnd.github+json",
	}, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *Service) githubRepoSearchContext(owner string, repo string, token string, terms []string, defaultBranch string) ([]map[string]interface{}, error) {
	values := url.Values{}
	queryParts := make([]string, 0, len(terms)+1)
	queryParts = append(queryParts, fmt.Sprintf("repo:%s/%s", owner, repo))
	queryParts = append(queryParts, terms...)
	values.Set("q", strings.Join(queryParts, " "))
	values.Set("per_page", "5")
	endpoint := fmt.Sprintf("%s/search/code?%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), values.Encode())
	var payload struct {
		Items []struct {
			Name        string `json:"name"`
			Path        string `json:"path"`
			HTMLURL     string `json:"html_url"`
			TextMatches []struct {
				Fragment string `json:"fragment"`
			} `json:"text_matches"`
		} `json:"items"`
	}
	if err := s.apiJSON(http.MethodGet, endpoint, nil, map[string]string{
		"Authorization": "Bearer " + token,
		"Accept":        "application/vnd.github.text-match+json",
	}, &payload); err != nil {
		return nil, err
	}
	matches := make([]map[string]interface{}, 0, minInt(len(payload.Items), 3))
	for _, item := range payload.Items {
		if len(matches) == 3 {
			break
		}
		snippet := ""
		for _, textMatch := range item.TextMatches {
			if fragment := truncate(strings.TrimSpace(textMatch.Fragment), 800); fragment != "" {
				snippet = fragment
				break
			}
		}
		if snippet == "" {
			contents, err := s.githubRepoFileContents(owner, repo, item.Path, defaultBranch, token)
			if err == nil {
				snippet = extractRepoContextSnippet(contents, terms)
			}
		}
		matches = append(matches, map[string]interface{}{
			"name":     item.Name,
			"path":     item.Path,
			"html_url": item.HTMLURL,
			"snippet":  truncate(snippet, 800),
		})
	}
	return matches, nil
}

func (s *Service) githubRepoFileContents(owner string, repo string, filePath string, ref string, token string) (string, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), owner, repo, path.Clean(strings.TrimPrefix(filePath, "/")))
	if strings.TrimSpace(ref) != "" {
		endpoint += "?ref=" + url.QueryEscape(ref)
	}
	var payload struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := s.apiJSON(http.MethodGet, endpoint, nil, map[string]string{
		"Authorization": "Bearer " + token,
		"Accept":        "application/vnd.github+json",
	}, &payload); err != nil {
		return "", err
	}
	if !strings.EqualFold(strings.TrimSpace(payload.Encoding), "base64") {
		return "", fmt.Errorf("unsupported content encoding %q", payload.Encoding)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(payload.Content, "\n", ""))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func (s *Service) githubRepoActivity(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	owner := s.cfg.GitHubRepoOwner(repo)
	repoName := s.cfg.GitHubRepoName(repo)
	token, err := s.githubInstallationToken(repo)
	if err != nil {
		return s.unavailableResult("github.repo_activity", input, "github", fmt.Sprintf("GitHub repo activity unavailable for %s/%s: missing app authentication.", owner, repoName), map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"error": err.Error(),
		})
	}
	since, until, err := parseActivityWindow(input)
	if err != nil {
		return s.failedResult("github.repo_activity", input, "github", fmt.Sprintf("GitHub repo activity input invalid: %v", err), map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"error": err.Error(),
		})
	}
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Accept":        "application/vnd.github+json",
	}

	commitValues := url.Values{}
	commitValues.Set("since", since.Format(time.RFC3339))
	commitValues.Set("until", until.Format(time.RFC3339))
	commitValues.Set("per_page", "25")
	commitEndpoint := fmt.Sprintf("%s/repos/%s/%s/commits?%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), owner, repoName, commitValues.Encode())
	var commitPayload []map[string]interface{}
	if err := s.apiJSON(http.MethodGet, commitEndpoint, nil, headers, &commitPayload); err != nil {
		return s.failedResult("github.repo_activity", input, "github", fmt.Sprintf("GitHub repo activity failed to load commits: %v", err), map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"error": err.Error(),
		})
	}

	pullValues := url.Values{}
	pullValues.Set("state", "all")
	pullValues.Set("sort", "updated")
	pullValues.Set("direction", "desc")
	pullValues.Set("per_page", "50")
	pullEndpoint := fmt.Sprintf("%s/repos/%s/%s/pulls?%s", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), owner, repoName, pullValues.Encode())
	var pullPayload []map[string]interface{}
	if err := s.apiJSON(http.MethodGet, pullEndpoint, nil, headers, &pullPayload); err != nil {
		return s.failedResult("github.repo_activity", input, "github", fmt.Sprintf("GitHub repo activity failed to load pull requests: %v", err), map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
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
	summary := fmt.Sprintf("GitHub activity for %s/%s from %s to %s includes %d commits, %d merged PRs, and %d opened PRs.", owner, repoName, since.Format("2006-01-02"), until.Format("2006-01-02"), len(commits), len(mergedPRs), len(openedPRs))
	return s.result("github.repo_activity", input, summary, map[string]interface{}{
		"repo":                 repoName,
		"owner":                owner,
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
	owner := s.cfg.GitHubRepoOwner(repo)
	repoName := s.cfg.GitHubRepoName(repo)
	head := firstNonEmpty(stringValue(input["branch_name"]), stringValue(input["head"]))
	base := firstNonEmpty(stringValue(input["base_ref"]), "main")
	title := firstNonEmpty(stringValue(input["title"]), fmt.Sprintf("RSI proposal for %s", repoName))
	body := firstNonEmpty(stringValue(input["body"]), "Automated draft PR from RSI platform.")
	writeToken, err := s.githubInstallationToken(repo)
	if err != nil {
		return s.unavailableResult("github.create_pr", input, "github", "GitHub App authentication not configured; refusing draft PR execution.", map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"head":  head,
			"base":  base,
			"error": err.Error(),
		})
	}
	requestBody := map[string]interface{}{
		"title": title,
		"head":  head,
		"base":  base,
		"body":  body,
		"draft": true,
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s/pulls", strings.TrimRight(s.cfg.GitHubAPIBaseURL, "/"), owner, repoName)
	var response map[string]interface{}
	if err := s.apiJSON(http.MethodPost, endpoint, requestBody, map[string]string{
		"Authorization": "Bearer " + writeToken,
		"Accept":        "application/vnd.github+json",
	}, &response); err != nil {
		return s.failedResult("github.create_pr", input, "github", fmt.Sprintf("GitHub PR creation failed: %v", err), map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"head":  head,
			"base":  base,
			"error": err.Error(),
		})
	}
	summary := fmt.Sprintf("Draft PR opened for %s/%s:%s.", owner, repoName, head)
	return s.result("github.create_pr", input, summary, map[string]interface{}{
		"repo":     repoName,
		"owner":    owner,
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

func (s *Service) rsiWorkflowContext(input map[string]interface{}) storepkg.ToolResult {
	workflowID := strings.TrimSpace(stringValue(input["workflow_id"]))
	traceID := strings.TrimSpace(stringValue(input["trace_id"]))
	var workflow storepkg.Workflow
	var found bool
	if workflowID != "" {
		for _, item := range s.store.ListWorkflows() {
			if item.ID == workflowID {
				workflow = item
				found = true
				break
			}
		}
	}
	var trace events.Trace
	if traceID != "" {
		trace, _ = s.store.GetTrace(traceID)
		if !found && trace.Summary.WorkflowID != "" {
			workflowID = trace.Summary.WorkflowID
			for _, item := range s.store.ListWorkflows() {
				if item.ID == workflowID {
					workflow = item
					found = true
					break
				}
			}
		}
	}
	if !found {
		return s.failedResult("rsi.workflow_context", input, "internal", "Workflow context requires workflow_id or trace_id bound to a workflow.", map[string]interface{}{
			"workflow_id": workflowID,
			"trace_id":    traceID,
			"error":       "not_found",
		})
	}
	recentEntries := make([]interface{}, 0)
	for _, item := range s.store.ListConversationEntries(workflow.ConversationID) {
		recentEntries = append(recentEntries, item)
		if len(recentEntries) == 8 {
			break
		}
	}
	assignments := make([]interface{}, 0)
	for _, item := range s.store.ListAssignments() {
		if workflow.ConversationID != "" && item.ConversationID == workflow.ConversationID {
			assignments = append(assignments, item)
			continue
		}
		if workflow.CaseID != "" && item.CaseID == workflow.CaseID {
			assignments = append(assignments, item)
		}
	}
	summary := fmt.Sprintf("RSI workflow context loaded for workflow %s.", workflow.ID)
	return s.result("rsi.workflow_context", input, summary, map[string]interface{}{
		"workflow":                    workflow,
		"trace_summary":               trace.Summary,
		"recent_conversation_entries": recentEntries,
		"assignments":                 assignments,
	}, nil)
}

func (s *Service) rsiActionChain(input map[string]interface{}) storepkg.ToolResult {
	traceID := strings.TrimSpace(stringValue(input["trace_id"]))
	proposalID := strings.TrimSpace(stringValue(input["proposal_id"]))
	attemptID := strings.TrimSpace(stringValue(input["attempt_id"]))
	intents := make([]interface{}, 0)
	results := make([]interface{}, 0)
	outcomes := make([]interface{}, 0)
	for _, intent := range s.store.ListActionIntents() {
		if traceID != "" && intent.TraceID != traceID {
			continue
		}
		if proposalID != "" && intent.ProposalID != proposalID {
			continue
		}
		if attemptID != "" && intent.AttemptID != attemptID {
			continue
		}
		intents = append(intents, intent)
		for _, result := range s.store.ListActionResults(intent.ID) {
			results = append(results, result)
		}
	}
	for _, outcome := range s.store.ListOutcomes() {
		if traceID != "" && outcome.TraceID != traceID {
			continue
		}
		if proposalID != "" && outcome.ProposalID != proposalID {
			continue
		}
		if attemptID != "" && outcome.AttemptID != attemptID {
			continue
		}
		outcomes = append(outcomes, outcome)
	}
	summary := fmt.Sprintf("RSI action chain loaded with %d intent(s), %d result(s), and %d outcome(s).", len(intents), len(results), len(outcomes))
	return s.result("rsi.action_chain", input, summary, map[string]interface{}{
		"trace_id":       traceID,
		"proposal_id":    proposalID,
		"attempt_id":     attemptID,
		"action_intents": intents,
		"action_results": results,
		"outcomes":       outcomes,
	}, nil)
}

func (s *Service) rsiRunnerExecution(input map[string]interface{}) storepkg.ToolResult {
	traceID := strings.TrimSpace(stringValue(input["trace_id"]))
	proposalID := strings.TrimSpace(stringValue(input["proposal_id"]))
	role := strings.TrimSpace(stringValue(input["role"]))
	executions := make([]interface{}, 0)
	for _, item := range s.store.ListHarnessExecutions() {
		if traceID != "" && item.TraceID != traceID {
			continue
		}
		if proposalID != "" && item.ProposalID != proposalID {
			continue
		}
		if role != "" && item.Role != role {
			continue
		}
		executions = append(executions, item)
	}
	summary := fmt.Sprintf("RSI runner execution lookup returned %d harness execution(s).", len(executions))
	return s.result("rsi.runner_execution", input, summary, map[string]interface{}{
		"trace_id":           traceID,
		"proposal_id":        proposalID,
		"role":               role,
		"harness_executions": executions,
	}, nil)
}

func (s *Service) rsiRuntimeConfig(input map[string]interface{}) storepkg.ToolResult {
	output := map[string]interface{}{
		"environment":                 s.cfg.Environment,
		"default_repo":                s.cfg.DefaultRepo,
		"allowed_target_repos":        append([]string(nil), s.cfg.AllowedTargetRepos...),
		"default_reasoning_verbosity": s.cfg.DefaultReasoningVerbosity,
		"active_proposal_cap":         s.cfg.DefaultProposalCap,
		"runner_urls":                 s.cfg.RunnerURLs(),
		"runner_timeouts_seconds": map[string]int{
			"prod":      int(s.cfg.ProdRunnerTimeout.Seconds()),
			"proactive": int(s.cfg.ProactiveRunnerTimeout.Seconds()),
			"eval":      int(s.cfg.EvalRunnerTimeout.Seconds()),
			"proposal":  int(s.cfg.ProposalRunnerTimeout.Seconds()),
		},
		"tool_gateway_base_url":   s.cfg.ToolGatewayBaseURL,
		"honcho_runtime_base_url": s.cfg.HonchoRuntimeBaseURL,
		"public_base_url":         s.cfg.PublicBaseURL,
	}
	return s.result("rsi.runtime_config", input, "RSI runtime configuration summary loaded.", output, nil)
}

func (s *Service) rsiRuntimeHealth(input map[string]interface{}) storepkg.ToolResult {
	runners := map[string]interface{}{}
	for role, baseURL := range s.cfg.RunnerURLs() {
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			runners[role] = map[string]interface{}{"status": "disabled"}
			continue
		}
		resp, err := clients.NewRunnerClientWithTimeout(baseURL, 5*time.Second).Runtime()
		if err != nil {
			runners[role] = map[string]interface{}{
				"status": "unreachable",
				"error":  err.Error(),
			}
			continue
		}
		runners[role] = resp
	}
	honcho := map[string]interface{}{"status": "disabled"}
	if strings.TrimSpace(s.cfg.HonchoRuntimeBaseURL) != "" {
		resp, err := clients.NewHonchoClient(s.cfg.HonchoRuntimeBaseURL).Runtime()
		if err != nil {
			honcho = map[string]interface{}{
				"status": "unreachable",
				"error":  err.Error(),
			}
		} else {
			honcho = map[string]interface{}{
				"status":               firstNonEmpty(stringValue(resp.Status), "ok"),
				"namespace":            resp.Namespace,
				"db_schema":            resp.DBSchema,
				"cache_enabled":        resp.CacheEnabled,
				"cache_url_configured": resp.CacheURLConfigured,
				"deriver":              resp.Deriver,
				"summary":              resp.Summary,
				"dialectic_levels":     resp.DialecticLevels,
			}
		}
	}
	return s.result("rsi.runtime_health", input, "RSI runtime health summary loaded.", map[string]interface{}{
		"runners": runners,
		"honcho":  honcho,
	}, nil)
}

func (s *Service) rsiTraceContext(input map[string]interface{}) storepkg.ToolResult {
	traceID := stringValue(input["trace_id"])
	if traceID == "" {
		return s.failedResult("rsi.trace_context", input, "internal", "Trace context requires trace_id.", map[string]interface{}{
			"error": "missing trace_id",
		})
	}
	trace, ok := s.store.GetTrace(traceID)
	if !ok {
		return s.failedResult("rsi.trace_context", input, "internal", fmt.Sprintf("Trace %s not found.", traceID), map[string]interface{}{
			"trace_id": traceID,
			"error":    "not_found",
		})
	}
	evalRuns := make([]interface{}, 0)
	evalJudgments := map[string][]interface{}{}
	for _, run := range s.store.ListEvalRuns() {
		if run.TraceID != traceID {
			continue
		}
		evalRuns = append(evalRuns, run)
		judgments := make([]interface{}, 0)
		for _, judgment := range s.store.ListEvalJudgments(run.ID) {
			judgments = append(judgments, judgment)
		}
		evalJudgments[run.ID] = judgments
	}
	linkedProposals := make([]interface{}, 0)
	for _, proposal := range s.store.ListProposals() {
		if proposal.TraceID == traceID || proposal.OriginTraceID == traceID || containsString(proposal.EvidenceTraceIDs, traceID) {
			linkedProposals = append(linkedProposals, proposal)
		}
	}
	summary := fmt.Sprintf("RSI trace context loaded for %s with %d events and %d linked proposals.", traceID, len(trace.Events), len(linkedProposals))
	return s.result("rsi.trace_context", input, summary, map[string]interface{}{
		"trace":            trace.Summary,
		"events":           trace.Events,
		"artifacts":        trace.Artifacts,
		"reasoning":        trace.Reasoning,
		"tool_calls":       trace.ToolCalls,
		"slack_actions":    trace.SlackActions,
		"eval_runs":        evalRuns,
		"eval_judgments":   evalJudgments,
		"linked_proposals": linkedProposals,
	}, nil)
}

func (s *Service) rsiProposalMemory(input map[string]interface{}) storepkg.ToolResult {
	candidateKey := strings.TrimSpace(stringValue(input["candidate_key"]))
	proposalID := strings.TrimSpace(stringValue(input["proposal_id"]))
	items := make([]interface{}, 0)
	for _, memory := range s.store.ListProposalMemories() {
		if candidateKey != "" && memory.CandidateKey != candidateKey {
			continue
		}
		if proposalID != "" && memory.ProposalID != proposalID {
			continue
		}
		items = append(items, memory)
	}
	summary := fmt.Sprintf("RSI proposal memory returned %d record(s).", len(items))
	return s.result("rsi.proposal_memory", input, summary, map[string]interface{}{
		"candidate_key": candidateKey,
		"proposal_id":   proposalID,
		"items":         items,
	}, nil)
}

func (s *Service) rsiCandidateContext(input map[string]interface{}) storepkg.ToolResult {
	candidateKey := strings.TrimSpace(stringValue(input["candidate_key"]))
	if candidateKey == "" {
		return s.failedResult("rsi.candidate_context", input, "internal", "Candidate context requires candidate_key.", map[string]interface{}{
			"error": "missing candidate_key",
		})
	}
	var candidate interface{}
	for _, item := range s.store.ListCandidates() {
		if item.CandidateKey == candidateKey {
			candidate = item
			break
		}
	}
	proposals := make([]interface{}, 0)
	for _, item := range s.store.ListProposals() {
		if item.CandidateKey == candidateKey {
			proposals = append(proposals, item)
		}
	}
	memories := make([]interface{}, 0)
	for _, item := range s.store.ListProposalMemories() {
		if item.CandidateKey == candidateKey {
			memories = append(memories, item)
		}
	}
	summary := fmt.Sprintf("RSI candidate context loaded for %s with %d proposal(s) and %d memory item(s).", candidateKey, len(proposals), len(memories))
	return s.result("rsi.candidate_context", input, summary, map[string]interface{}{
		"candidate_key":   candidateKey,
		"candidate":       candidate,
		"proposals":       proposals,
		"proposal_memory": memories,
	}, nil)
}

func (s *Service) rsiAttemptContext(input map[string]interface{}) storepkg.ToolResult {
	attemptID := strings.TrimSpace(stringValue(input["attempt_id"]))
	if attemptID == "" {
		return s.failedResult("rsi.attempt_context", input, "internal", "Attempt context requires attempt_id.", map[string]interface{}{
			"error": "missing attempt_id",
		})
	}
	attempt, ok := s.store.GetChangeAttempt(attemptID)
	if !ok {
		return s.failedResult("rsi.attempt_context", input, "internal", fmt.Sprintf("Attempt %s not found.", attemptID), map[string]interface{}{
			"attempt_id": attemptID,
			"error":      "not_found",
		})
	}
	repoJobs := make([]interface{}, 0)
	for _, item := range s.store.ListRepoChangeJobs() {
		if item.AttemptID == attemptID {
			repoJobs = append(repoJobs, item)
		}
	}
	prAttempts := make([]interface{}, 0)
	for _, item := range s.store.ListPRAttempts() {
		if item.AttemptID == attemptID {
			prAttempts = append(prAttempts, item)
		}
	}
	actionIntents := make([]interface{}, 0)
	actionResults := make([]interface{}, 0)
	for _, intent := range s.store.ListActionIntents() {
		if intent.AttemptID != attemptID {
			continue
		}
		actionIntents = append(actionIntents, intent)
		for _, result := range s.store.ListActionResults(intent.ID) {
			actionResults = append(actionResults, result)
		}
	}
	outcomes := make([]interface{}, 0)
	for _, outcome := range s.store.ListOutcomes() {
		if outcome.AttemptID == attemptID {
			outcomes = append(outcomes, outcome)
		}
	}
	summary := fmt.Sprintf("RSI attempt context loaded for %s with %d repo jobs and %d PR attempt(s).", attemptID, len(repoJobs), len(prAttempts))
	return s.result("rsi.attempt_context", input, summary, map[string]interface{}{
		"attempt":          attempt,
		"repo_change_jobs": repoJobs,
		"pr_attempts":      prAttempts,
		"action_intents":   actionIntents,
		"action_results":   actionResults,
		"outcomes":         outcomes,
	}, nil)
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
	result := storepkg.ToolResult{
		Name:            name,
		ToolCallID:      newToolCallID(name),
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
	return result
}

func (s *Service) unavailableResult(name string, input map[string]interface{}, provider string, summary string, output map[string]interface{}) storepkg.ToolResult {
	result := storepkg.ToolResult{
		Name:          name,
		ToolCallID:    newToolCallID(name),
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
	return result
}

func (s *Service) failedResult(name string, input map[string]interface{}, provider string, summary string, output map[string]interface{}) storepkg.ToolResult {
	result := storepkg.ToolResult{
		Name:          name,
		ToolCallID:    newToolCallID(name),
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
	return result
}

func newToolCallID(name string) string {
	return fmt.Sprintf("%s-%d", sanitizeToolName(name), time.Now().UTC().UnixNano())
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

func (s *Service) githubInstallationToken(repo string) (string, error) {
	token, err := githubapp.NewClient(
		s.cfg.GitHubAppID,
		s.cfg.GitHubInstallationIDForRepo(repo),
		s.cfg.GitHubAppPrivateKey,
		s.cfg.GitHubAPIBaseURL,
		s.httpClient,
	).MintInstallationToken(context.Background(), []string{s.cfg.GitHubRepoName(repo)})
	if err != nil {
		return "", err
	}
	return token.Token, nil
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

func containsString(values []string, needle string) bool {
	needle = strings.TrimSpace(needle)
	for _, item := range values {
		if strings.TrimSpace(item) == needle {
			return true
		}
	}
	return false
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit]
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
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
	case strings.HasPrefix(name, "workspace."):
		return "sandbox"
	case strings.HasPrefix(name, "rsi."):
		return "rsi"
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
	case "workspace.git_diff":
		return stringValue(output["workspace_id"])
	case "workspace.run_validation":
		return stringValue(output["workspace_id"])
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

func repoContextSearchTerms(question string) []string {
	stopwords := map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "are": {}, "can": {}, "for": {}, "from": {}, "give": {}, "how": {}, "in": {}, "into": {}, "last": {}, "me": {}, "of": {}, "on": {}, "or": {}, "quick": {}, "rundown": {}, "the": {}, "this": {}, "to": {}, "week": {}, "with": {}, "you": {},
	}
	fields := strings.FieldsFunc(strings.ToLower(question), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' && r != '.'
	})
	out := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len(field) < 3 {
			continue
		}
		if _, blocked := stopwords[field]; blocked {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
		if len(out) == 6 {
			break
		}
	}
	return out
}

func extractRepoContextSnippet(contents string, terms []string) string {
	lines := strings.Split(contents, "\n")
	lowerTerms := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term != "" {
			lowerTerms = append(lowerTerms, term)
		}
	}
	matchIndex := -1
	for index, line := range lines {
		lower := strings.ToLower(line)
		for _, term := range lowerTerms {
			if strings.Contains(lower, term) {
				matchIndex = index
				break
			}
		}
		if matchIndex >= 0 {
			break
		}
	}
	start := 0
	end := minInt(len(lines), 20)
	if matchIndex >= 0 {
		start = maxInt(matchIndex-2, 0)
		end = minInt(matchIndex+3, len(lines))
	}
	snippetLines := make([]string, 0, end-start)
	for idx := start; idx < end; idx++ {
		snippetLines = append(snippetLines, fmt.Sprintf("%d: %s", idx+1, lines[idx]))
	}
	return strings.Join(snippetLines, "\n")
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
