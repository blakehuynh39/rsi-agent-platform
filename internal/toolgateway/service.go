package toolgateway

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/cluster"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/debuglog"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/githubapp"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/workflowplan"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const slackMentionsOnlySentinel = "MENTIONS_ONLY"

type Service struct {
	cfg         config.Config
	store       storepkg.Repository
	httpClient  *http.Client
	slackClient *slackapi.Client
	slackAPIURL string
	kubeClient  kubernetes.Interface
	launcher    sandbox.Launcher
}

func (s *Service) RecordRuntimeObservation(input map[string]interface{}) (int, map[string]interface{}) {
	item := harness.ExecutionObservation{
		ID:              strings.TrimSpace(stringValue(input["id"])),
		ExecutionID:     strings.TrimSpace(stringValue(input["execution_id"])),
		OperationID:     strings.TrimSpace(stringValue(input["operation_id"])),
		TraceID:         strings.TrimSpace(stringValue(input["trace_id"])),
		WorkflowID:      strings.TrimSpace(stringValue(input["workflow_id"])),
		HermesSessionID: strings.TrimSpace(stringValue(input["hermes_session_id"])),
		Role:            strings.TrimSpace(stringValue(input["role"])),
		Phase:           strings.TrimSpace(stringValue(input["phase"])),
		EventType:       strings.TrimSpace(stringValue(input["event_type"])),
		Status:          strings.TrimSpace(stringValue(input["status"])),
		Seq:             intValue(input["seq"]),
		Payload:         mapValue(input["payload"]),
	}
	recordedAt := strings.TrimSpace(stringValue(input["recorded_at"]))
	if recordedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, recordedAt); err == nil {
			item.RecordedAt = parsed.UTC()
		}
	}
	if item.ExecutionID == "" || item.Phase == "" || item.EventType == "" {
		return http.StatusBadRequest, map[string]interface{}{
			"error":        "missing required observation fields",
			"execution_id": item.ExecutionID,
			"phase":        item.Phase,
			"event_type":   item.EventType,
		}
	}
	recorded, err := s.store.RecordHarnessExecutionObservation(item)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		}
	}
	return http.StatusOK, map[string]interface{}{
		"status":      "ok",
		"observation": recorded,
	}
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
		slackAPIURL: "https://slack.com/api/",
		kubeClient:  kubeClient,
		launcher:    launcher,
	}
}

func (s *Service) Execute(name string, input map[string]interface{}) storepkg.ToolResult {
	if s.cfg.VerboseTraceLogging {
		log.Printf("tool-gateway execute_start tool=%s input=%s", strings.TrimSpace(name), debuglog.JSON(input, s.cfg.VerboseTraceLogLimit))
	}
	var result storepkg.ToolResult
	switch name {
	case "repo.context":
		result = s.repoContext(input)
	case "repo.read_file":
		result = s.repoReadFile(input)
	case "repo.search":
		result = s.repoSearch(input)
	case "knowledge.context":
		result = s.knowledgeContext(input)
	case "sentry.lookup":
		result = s.sentryLookup(input)
	case "kubernetes.inspect":
		result = s.kubernetesInspect(input)
	case "kubernetes.logs":
		result = s.kubernetesLogs(input)
	case "kubernetes.events":
		result = s.kubernetesEvents(input)
	case "slack.reply":
		result = s.slackReply(input)
	case "slack.upload_file":
		result = s.slackUploadFile(input)
	case "slack.history":
		result = s.slackHistory(input)
	case "slack.search":
		result = s.slackSearch(input)
	case "github.create_pr":
		result = s.githubCreatePR(input)
	case "github.repo_context":
		result = s.githubRepoContext(input)
	case "github.repo_activity":
		result = s.githubRepoActivity(input)
	case "cloudflare.inspect":
		result = s.cloudflareInspect(input)
	case "rsi.trace_context":
		result = s.rsiTraceContext(input)
	case "rsi.workflow_context":
		result = s.rsiWorkflowContext(input)
	case "rsi.action_chain":
		result = s.rsiActionChain(input)
	case "rsi.runner_execution":
		result = s.rsiRunnerExecution(input)
	case "rsi.runtime_config":
		result = s.rsiRuntimeConfig(input)
	case "rsi.runtime_health":
		result = s.rsiRuntimeHealth(input)
	case "rsi.runtime_deployment_facts":
		result = s.rsiRuntimeDeploymentFacts(input)
	case "rsi.proposal_memory":
		result = s.rsiProposalMemory(input)
	case "rsi.candidate_context":
		result = s.rsiCandidateContext(input)
	case "rsi.attempt_context":
		result = s.rsiAttemptContext(input)
	case "workspace.list_files":
		result = s.workspaceListFiles(input)
	case "workspace.read_file":
		result = s.workspaceReadFile(input)
	case "workspace.search":
		result = s.workspaceSearch(input)
	case "workspace.git_history":
		result = s.workspaceGitHistory(input)
	case "workspace.git_show":
		result = s.workspaceGitShow(input)
	case "workspace.git_search":
		result = s.workspaceGitSearch(input)
	case "workspace.write_file":
		result = s.workspaceWriteFile(input)
	case "workspace.apply_patch":
		result = s.workspaceApplyPatch(input)
	case "workspace.git_status":
		result = s.workspaceGitStatus(input)
	case "workspace.git_diff":
		result = s.workspaceGitDiff(input)
	case "workspace.run_validation":
		result = s.workspaceRunValidation(input)
	default:
		result = s.unavailableResult(name, input, "tool-gateway", fmt.Sprintf("Tool %s is not registered in the governed tool gateway.", strings.TrimSpace(name)), map[string]interface{}{
			"tool_name": strings.TrimSpace(name),
			"error":     "unknown_tool",
		})
	}
	if s.cfg.VerboseTraceLogging {
		log.Printf(
			"tool-gateway execute_end tool=%s status=%s available=%t summary=%q output=%s",
			strings.TrimSpace(name),
			strings.TrimSpace(result.Status),
			result.Available,
			strings.TrimSpace(result.Summary),
			debuglog.JSON(result.Output, s.cfg.VerboseTraceLogLimit),
		)
	}
	return result
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

func (s *Service) repoReadFile(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	filePath := strings.TrimPrefix(path.Clean("/"+stringValue(input["path"])), "/")
	if filePath == "" || filePath == "." {
		return s.failedResult("repo.read_file", input, "github", "Repository file read requires path.", map[string]interface{}{
			"repo":  s.cfg.GitHubRepoName(repo),
			"owner": s.cfg.GitHubRepoOwner(repo),
			"error": "missing path",
		})
	}
	owner := s.cfg.GitHubRepoOwner(repo)
	repoName := s.cfg.GitHubRepoName(repo)
	token, err := s.githubInstallationToken(repo)
	if err != nil {
		return s.unavailableResult("repo.read_file", input, "github", fmt.Sprintf("Repository file read unavailable for %s/%s: missing app authentication.", owner, repoName), map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"path":  filePath,
			"error": err.Error(),
		})
	}
	ref := strings.TrimSpace(stringValue(input["ref"]))
	if ref == "" {
		if meta, metaErr := s.githubRepoMetadata(owner, repoName, token); metaErr == nil {
			ref = stringValue(meta["default_branch"])
		}
	}
	content, err := s.githubRepoFileContents(owner, repoName, filePath, ref, token)
	if err != nil {
		return s.failedResult("repo.read_file", input, "github", fmt.Sprintf("Repository file read failed: %v", err), map[string]interface{}{
			"repo":  repoName,
			"owner": owner,
			"path":  filePath,
			"ref":   ref,
			"error": err.Error(),
		})
	}
	summary := fmt.Sprintf("Repository file %s loaded from %s/%s at %s.", filePath, owner, repoName, firstNonEmpty(ref, "default branch"))
	return s.result("repo.read_file", input, summary, map[string]interface{}{
		"repo":           repoName,
		"owner":          owner,
		"path":           filePath,
		"ref":            ref,
		"content":        truncate(content, 12000),
		"content_length": len(content),
		"truncated":      len(content) > len(truncate(content, 12000)),
	}, nil)
}

func (s *Service) repoSearch(input map[string]interface{}) storepkg.ToolResult {
	repo := firstNonEmpty(stringValue(input["repo"]), s.cfg.DefaultRepo)
	pattern := strings.TrimSpace(stringValue(input["pattern"]))
	pathPrefix := strings.TrimPrefix(path.Clean("/"+stringValue(input["path"])), "/")
	if pattern == "" {
		return s.failedResult("repo.search", input, "github", "Repository search requires pattern.", map[string]interface{}{
			"repo":  s.cfg.GitHubRepoName(repo),
			"owner": s.cfg.GitHubRepoOwner(repo),
			"error": "missing pattern",
		})
	}
	owner := s.cfg.GitHubRepoOwner(repo)
	repoName := s.cfg.GitHubRepoName(repo)
	token, err := s.githubInstallationToken(repo)
	if err != nil {
		return s.unavailableResult("repo.search", input, "github", fmt.Sprintf("Repository search unavailable for %s/%s: missing app authentication.", owner, repoName), map[string]interface{}{
			"repo":    repoName,
			"owner":   owner,
			"pattern": pattern,
			"path":    pathPrefix,
			"error":   err.Error(),
		})
	}
	matches, err := s.githubRepoSearch(owner, repoName, token, pattern, pathPrefix)
	if err != nil {
		return s.failedResult("repo.search", input, "github", fmt.Sprintf("Repository search failed: %v", err), map[string]interface{}{
			"repo":    repoName,
			"owner":   owner,
			"pattern": pattern,
			"path":    pathPrefix,
			"error":   err.Error(),
		})
	}
	summary := fmt.Sprintf("Repository search found %d match(es) in %s/%s for %q.", len(matches), owner, repoName, pattern)
	return s.result("repo.search", input, summary, map[string]interface{}{
		"repo":    repoName,
		"owner":   owner,
		"pattern": pattern,
		"path":    pathPrefix,
		"matches": matches,
	}, nil)
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

func (s *Service) githubRepoSearch(owner string, repo string, token string, pattern string, pathPrefix string) ([]map[string]interface{}, error) {
	values := url.Values{}
	queryParts := []string{fmt.Sprintf("repo:%s/%s", owner, repo), pattern}
	if trimmedPath := strings.TrimSpace(pathPrefix); trimmedPath != "" && trimmedPath != "." {
		queryParts = append(queryParts, fmt.Sprintf("path:%s", trimmedPath))
	}
	values.Set("q", strings.Join(queryParts, " "))
	values.Set("per_page", "10")
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
	matches := make([]map[string]interface{}, 0, minInt(len(payload.Items), 8))
	for _, item := range payload.Items {
		if len(matches) == 8 {
			break
		}
		snippet := ""
		for _, textMatch := range item.TextMatches {
			if fragment := truncate(strings.TrimSpace(textMatch.Fragment), 800); fragment != "" {
				snippet = fragment
				break
			}
		}
		matches = append(matches, map[string]interface{}{
			"name":     item.Name,
			"path":     item.Path,
			"html_url": item.HTMLURL,
			"snippet":  snippet,
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
	var workflowLine interface{}
	if workflow.CaseID != "" {
		if item, ok := s.store.GetWorkflowLine(workflow.CaseID); ok {
			workflowLine = item
		}
	}
	workflowAttempts := make([]interface{}, 0)
	for _, item := range s.store.ListWorkflows() {
		if workflow.CaseID != "" && item.CaseID == workflow.CaseID {
			workflowAttempts = append(workflowAttempts, item)
			continue
		}
		if item.ID == workflow.ID || item.TraceID == workflow.TraceID {
			workflowAttempts = append(workflowAttempts, item)
		}
	}
	summary := fmt.Sprintf("RSI workflow context loaded for workflow %s.", workflow.ID)
	return s.result("rsi.workflow_context", input, summary, map[string]interface{}{
		"workflow":                    workflow,
		"workflow_line":               workflowLine,
		"workflow_attempts":           workflowAttempts,
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
	observationsByExecution := map[string][]interface{}{}
	selectedExecutionIDs := map[string]struct{}{}
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
		selectedExecutionIDs[item.ID] = struct{}{}
	}
	for _, item := range s.store.ListHarnessExecutionObservations() {
		if traceID != "" && item.TraceID != traceID {
			continue
		}
		if proposalID != "" {
			if _, ok := selectedExecutionIDs[item.ExecutionID]; !ok {
				continue
			}
		}
		if role != "" && item.Role != role {
			continue
		}
		observationsByExecution[item.ExecutionID] = append(observationsByExecution[item.ExecutionID], item)
	}
	summary := fmt.Sprintf("RSI runner execution lookup returned %d harness execution(s) and %d observation group(s).", len(executions), len(observationsByExecution))
	return s.result("rsi.runner_execution", input, summary, map[string]interface{}{
		"trace_id":                       traceID,
		"proposal_id":                    proposalID,
		"role":                           role,
		"harness_executions":             executions,
		"harness_execution_observations": observationsByExecution,
	}, nil)
}

func (s *Service) rsiRuntimeConfig(input map[string]interface{}) storepkg.ToolResult {
	output := map[string]interface{}{
		"environment":                  s.cfg.Environment,
		"default_repo":                 s.cfg.DefaultRepo,
		"allowed_target_repos":         append([]string(nil), s.cfg.AllowedTargetRepos...),
		"default_reasoning_verbosity":  s.cfg.DefaultReasoningVerbosity,
		"active_proposal_cap":          s.cfg.DefaultProposalCap,
		"runner_urls":                  s.cfg.RunnerURLs(),
		"runner_timeouts_seconds":      runtimeTimeoutsSeconds(s.cfg),
		"runner_task_timeouts_seconds": runtimeTaskTimeoutsSeconds(s.cfg),
		"tool_gateway_base_url":        s.cfg.ToolGatewayBaseURL,
		"honcho_runtime_base_url":      s.cfg.HonchoRuntimeBaseURL,
		"public_base_url":              s.cfg.PublicBaseURL,
		"slack_search_auth_mode":       s.slackSearchAuthMode(),
		"slack_user_token_configured":  strings.TrimSpace(s.cfg.SlackUserToken) != "",
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

func (s *Service) rsiRuntimeDeploymentFacts(input map[string]interface{}) storepkg.ToolResult {
	namespace := firstNonEmpty(stringValue(input["namespace"]), s.cfg.SandboxNamespace)
	targets := runtimeDeploymentTargets(input)
	output := map[string]interface{}{
		"environment":                          s.cfg.Environment,
		"namespace":                            namespace,
		"service_targets":                      targets,
		"runner_urls":                          s.cfg.RunnerURLs(),
		"runner_timeouts_seconds":              runtimeTimeoutsSeconds(s.cfg),
		"runner_task_timeouts_seconds":         runtimeTaskTimeoutsSeconds(s.cfg),
		"tool_gateway_base_url":                s.cfg.ToolGatewayBaseURL,
		"honcho_runtime_base_url":              s.cfg.HonchoRuntimeBaseURL,
		"public_base_url":                      s.cfg.PublicBaseURL,
		"slack_app_identity":                   s.cfg.SlackAppIdentity,
		"slack_socket_mode_enabled":            s.cfg.SlackSocketModeEnabled,
		"slack_bot_configured":                 strings.TrimSpace(s.cfg.SlackBotToken) != "",
		"slack_user_token_configured":          strings.TrimSpace(s.cfg.SlackUserToken) != "",
		"slack_search_auth_mode":               s.slackSearchAuthMode(),
		"allowed_slack_channel_ids":            append([]string(nil), s.cfg.AllowedSlackChannelIDs...),
		"hermes_native_governed_tools_enabled": s.cfg.HermesNativeGovernedToolsEnabled,
		"kubernetes_available":                 s.kubeClient != nil,
		"deployments":                          []map[string]interface{}{},
	}
	if s.kubeClient == nil {
		output["kubernetes_error"] = "kubernetes client unavailable"
		return s.result("rsi.runtime_deployment_facts", input, "RSI runtime deployment facts loaded from config only; Kubernetes client unavailable.", output, nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	deployments, err := s.kubeClient.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		output["kubernetes_error"] = err.Error()
		return s.result("rsi.runtime_deployment_facts", input, fmt.Sprintf("RSI runtime deployment facts loaded from config, but Kubernetes deployment lookup failed: %v", err), output, nil)
	}

	items := make([]map[string]interface{}, 0)
	for _, deployment := range deployments.Items {
		matchedTargets := matchingDeploymentTargets(deployment, targets)
		if len(matchedTargets) == 0 {
			continue
		}
		entry := map[string]interface{}{
			"name":                deployment.Name,
			"namespace":           deployment.Namespace,
			"matched_targets":     matchedTargets,
			"generation":          deployment.Generation,
			"observed_generation": deployment.Status.ObservedGeneration,
			"replicas":            deployment.Status.Replicas,
			"ready_replicas":      deployment.Status.ReadyReplicas,
			"updated_replicas":    deployment.Status.UpdatedReplicas,
			"available_replicas":  deployment.Status.AvailableReplicas,
			"images":              deploymentImages(deployment),
		}
		if !deployment.CreationTimestamp.IsZero() {
			entry["created_at"] = deployment.CreationTimestamp.UTC().Format(time.RFC3339)
		}
		if condition, ok := deploymentCondition(deployment, appsv1.DeploymentProgressing); ok {
			entry["progressing_status"] = string(condition.Status)
			entry["progressing_reason"] = condition.Reason
			entry["progressing_message"] = truncate(condition.Message, 400)
			if !condition.LastUpdateTime.IsZero() {
				entry["progressing_updated_at"] = condition.LastUpdateTime.UTC().Format(time.RFC3339)
			}
			if !condition.LastTransitionTime.IsZero() {
				entry["progressing_transitioned_at"] = condition.LastTransitionTime.UTC().Format(time.RFC3339)
			}
		}
		if condition, ok := deploymentCondition(deployment, appsv1.DeploymentAvailable); ok {
			entry["available_status"] = string(condition.Status)
			entry["available_reason"] = condition.Reason
		}
		items = append(items, entry)
	}
	sort.Slice(items, func(i, j int) bool {
		return stringValue(items[i]["name"]) < stringValue(items[j]["name"])
	})
	output["deployments"] = items
	summary := fmt.Sprintf("RSI runtime deployment facts loaded for %s with %d matching deployment(s).", namespace, len(items))
	if len(items) == 0 {
		summary = fmt.Sprintf("RSI runtime deployment facts loaded for %s with no matching deployments.", namespace)
	}
	return s.result("rsi.runtime_deployment_facts", input, summary, output, nil)
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
	candidateKeys := make(map[string]struct{})
	for _, candidate := range s.store.ListCandidates() {
		if candidate.LatestTraceID == traceID || containsString(candidate.EvidenceTraceIDs, traceID) {
			candidateKeys[candidate.CandidateKey] = struct{}{}
		}
	}
	var workflowLine interface{}
	if caseID := strings.TrimSpace(trace.Summary.CaseID); caseID != "" {
		if item, ok := s.store.GetWorkflowLine(caseID); ok {
			workflowLine = item
		}
	}
	workflowAttempts := make([]interface{}, 0)
	for _, workflow := range s.store.ListWorkflows() {
		if workflow.CaseID != "" && workflow.CaseID == trace.Summary.CaseID {
			workflowAttempts = append(workflowAttempts, workflow)
			continue
		}
		if workflow.TraceID == traceID || workflow.ID == trace.Summary.WorkflowID {
			workflowAttempts = append(workflowAttempts, workflow)
		}
	}
	harnessExecutions := make([]interface{}, 0)
	for _, execution := range s.store.ListHarnessExecutions() {
		if execution.TraceID == traceID {
			harnessExecutions = append(harnessExecutions, execution)
		}
	}
	harnessExecutionObservations := make([]interface{}, 0)
	for _, observation := range s.store.ListHarnessExecutionObservations() {
		if observation.TraceID == traceID {
			harnessExecutionObservations = append(harnessExecutionObservations, observation)
		}
	}
	runtimeDiagnoses := make([]interface{}, 0)
	for _, diagnosis := range s.store.ListRuntimeDiagnoses() {
		if diagnosis.LatestTraceID == traceID || (trace.Summary.CaseID != "" && diagnosis.CaseID == trace.Summary.CaseID) {
			runtimeDiagnoses = append(runtimeDiagnoses, diagnosis)
			continue
		}
		if _, ok := candidateKeys[diagnosis.CandidateKey]; ok {
			runtimeDiagnoses = append(runtimeDiagnoses, diagnosis)
		}
	}
	summary := fmt.Sprintf("RSI trace context loaded for %s with %d events, %d workflow attempt(s), %d runtime diagnosis record(s), %d harness observation(s), and %d linked proposals.", traceID, len(trace.Events), len(workflowAttempts), len(runtimeDiagnoses), len(harnessExecutionObservations), len(linkedProposals))
	return s.result("rsi.trace_context", input, summary, map[string]interface{}{
		"trace":                          trace.Summary,
		"workflow_line":                  workflowLine,
		"workflow_attempts":              workflowAttempts,
		"runtime_diagnoses":              runtimeDiagnoses,
		"harness_executions":             harnessExecutions,
		"harness_execution_observations": harnessExecutionObservations,
		"events":                         trace.Events,
		"artifacts":                      trace.Artifacts,
		"reasoning":                      trace.Reasoning,
		"tool_calls":                     trace.ToolCalls,
		"slack_actions":                  trace.SlackActions,
		"eval_runs":                      evalRuns,
		"eval_judgments":                 evalJudgments,
		"linked_proposals":               linkedProposals,
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
	runtimeDiagnoses := make([]interface{}, 0)
	for _, item := range s.store.ListRuntimeDiagnoses() {
		if item.CandidateKey == candidateKey {
			runtimeDiagnoses = append(runtimeDiagnoses, item)
		}
	}
	summary := fmt.Sprintf("RSI candidate context loaded for %s with %d proposal(s), %d memory item(s), and %d runtime diagnosis record(s).", candidateKey, len(proposals), len(memories), len(runtimeDiagnoses))
	return s.result("rsi.candidate_context", input, summary, map[string]interface{}{
		"candidate_key":     candidateKey,
		"candidate":         candidate,
		"proposals":         proposals,
		"proposal_memory":   memories,
		"runtime_diagnoses": runtimeDiagnoses,
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

func (s *Service) slackUploadFile(input map[string]interface{}) storepkg.ToolResult {
	channelID := strings.TrimSpace(stringValue(input["channel_id"]))
	threadTS := strings.TrimSpace(stringValue(input["thread_ts"]))
	filename := strings.TrimSpace(stringValue(input["filename"]))
	title := strings.TrimSpace(stringValue(input["title"]))
	content := stringValue(input["content"])
	contentBase64 := strings.TrimSpace(stringValue(input["content_base64"]))
	initialComment := stringValue(input["initial_comment"])
	altTxt := strings.TrimSpace(stringValue(input["alt_txt"]))
	snippetType := strings.TrimSpace(stringValue(input["snippet_type"]))
	dryRun := boolValue(input["dry_run"])
	output := map[string]interface{}{
		"channel_id": channelID,
		"thread_ts":  threadTS,
		"filename":   filename,
		"title":      title,
		"uploaded":   false,
	}
	if channelID == "" || filename == "" {
		return s.failedResult("slack.upload_file", input, "slack", "Slack file upload missing channel or filename.", output)
	}
	var data []byte
	switch {
	case contentBase64 != "":
		decoded, err := base64.StdEncoding.DecodeString(contentBase64)
		if err != nil {
			output["error"] = err.Error()
			return s.failedResult("slack.upload_file", input, "slack", fmt.Sprintf("Slack file upload invalid base64 payload: %v", err), output)
		}
		data = decoded
	case content != "":
		data = []byte(content)
	default:
		return s.failedResult("slack.upload_file", input, "slack", "Slack file upload missing file content.", output)
	}
	if len(data) == 0 {
		return s.failedResult("slack.upload_file", input, "slack", "Slack file upload content is empty.", output)
	}
	output["size_bytes"] = len(data)
	if !dryRun && strings.TrimSpace(s.cfg.SlackBotToken) == "" {
		output["error"] = "missing RSI_SLACK_BOT_TOKEN"
		return s.unavailableResult("slack.upload_file", input, "slack", "Slack file upload unavailable: bot token is not configured.", output)
	}
	if dryRun {
		return s.result("slack.upload_file", input, "Slack file upload dry-run generated.", output, nil)
	}
	client := slackapi.New(
		s.cfg.SlackBotToken,
		slackapi.OptionAPIURL(s.slackAPIURL),
		slackapi.OptionHTTPClient(s.httpClient),
	)
	uploaded, err := client.UploadFileContext(context.Background(), slackapi.UploadFileParameters{
		Reader:          bytes.NewReader(data),
		FileSize:        len(data),
		Filename:        filename,
		Title:           title,
		InitialComment:  initialComment,
		Channel:         channelID,
		ThreadTimestamp: threadTS,
		AltTxt:          altTxt,
		SnippetType:     snippetType,
	})
	if err != nil {
		output["error"] = err.Error()
		return s.failedResult("slack.upload_file", input, "slack", fmt.Sprintf("Slack file upload failed: %v", err), output)
	}
	output["uploaded"] = true
	output["file_id"] = uploaded.ID
	if uploaded.Title != "" {
		output["title"] = uploaded.Title
	}
	refs := []string{}
	if uploaded.ID != "" {
		refs = append(refs, "slack-file://"+uploaded.ID)
		if info, _, _, infoErr := client.GetFileInfoContext(context.Background(), uploaded.ID, 1, 1); infoErr == nil {
			if info.Permalink != "" {
				output["permalink"] = info.Permalink
				refs = append(refs, info.Permalink)
			}
			if info.Mimetype != "" {
				output["mimetype"] = info.Mimetype
			}
			if info.Filetype != "" {
				output["filetype"] = info.Filetype
			}
			if info.Size > 0 {
				output["size_bytes"] = info.Size
			}
		}
	}
	return s.result("slack.upload_file", input, fmt.Sprintf("Slack file %s uploaded to %s.", filename, channelID), output, refs)
}

func (s *Service) slackHistory(input map[string]interface{}) storepkg.ToolResult {
	traceID := strings.TrimSpace(stringValue(input["trace_id"]))
	bound := s.boundSlackContext(traceID)
	question := firstNonEmpty(strings.TrimSpace(stringValue(input["question"])), strings.TrimSpace(bound.Prompt))
	promptContext := slackPromptContext(bound.Prompt, question)
	contextualChannelIDs := slackMentionChannelIDs(promptContext, bound.ChannelID, bound.ThreadTS, bound.EntityRefs, s.cfg.AllowedSlackChannelIDs)
	channelID, threadTS := resolveSlackHistoryTarget(input, promptContext, bound)
	scope := slackHistoryScope(input, question, threadTS)
	limit := slackHistoryLimit(input)
	oldest, latest, err := slackHistoryWindow(input)
	output := map[string]interface{}{
		"channel_id": channelID,
		"thread_ts":  threadTS,
		"scope":      scope,
		"limit":      limit,
		"oldest":     oldest,
		"latest":     latest,
		"messages":   []map[string]interface{}{},
	}
	if err != nil {
		output["error"] = err.Error()
		return s.failedResult("slack.history", input, "slack", fmt.Sprintf("Slack history request invalid: %v", err), output)
	}
	if channelID == "" {
		output["error"] = "missing channel_id"
		return s.failedResult("slack.history", input, "slack", "Slack history requires channel_id or a trace bound to a Slack ingestion.", output)
	}
	if !slackChannelAllowed(channelID, bound.ChannelID, contextualChannelIDs, s.cfg.AllowedSlackChannelIDs) {
		output["error"] = "channel_not_allowed"
		return s.failedResult("slack.history", input, "slack", fmt.Sprintf("Slack history not permitted for channel %s.", channelID), output)
	}
	if s.slackClient == nil {
		output["error"] = "missing RSI_SLACK_BOT_TOKEN"
		return s.unavailableResult("slack.history", input, "slack", "Slack history unavailable: bot token is not configured.", output)
	}
	if scope == "thread" && threadTS != "" {
		messages, hasMore, nextCursor, err := s.slackClient.GetConversationReplies(&slackapi.GetConversationRepliesParameters{
			ChannelID: channelID,
			Timestamp: threadTS,
			Oldest:    oldest,
			Latest:    latest,
			Limit:     limit,
			Inclusive: true,
		})
		if err != nil {
			output["error"] = err.Error()
			return s.failedResult("slack.history", input, "slack", fmt.Sprintf("Slack thread history failed: %v", err), output)
		}
		output["messages"] = slackMessagesPayload(messages)
		output["has_more"] = hasMore
		output["next_cursor"] = nextCursor
		summary := fmt.Sprintf("Slack thread history loaded from %s with %d message(s).", channelID, len(messages))
		return s.result("slack.history", input, summary, output, nil)
	}
	history, err := s.slackClient.GetConversationHistory(&slackapi.GetConversationHistoryParameters{
		ChannelID: channelID,
		Oldest:    oldest,
		Latest:    latest,
		Limit:     limit,
		Inclusive: true,
	})
	if err != nil {
		output["error"] = err.Error()
		return s.failedResult("slack.history", input, "slack", fmt.Sprintf("Slack channel history failed: %v", err), output)
	}
	output["messages"] = slackMessagesPayload(history.Messages)
	output["has_more"] = history.HasMore
	output["next_cursor"] = history.ResponseMetaData.NextCursor
	summary := fmt.Sprintf("Slack channel history loaded from %s with %d message(s).", channelID, len(history.Messages))
	return s.result("slack.history", input, summary, output, nil)
}

func resolveSlackHistoryTarget(input map[string]interface{}, question string, bound boundSlackContext) (string, string) {
	requestedChannelID := strings.TrimSpace(stringValue(input["channel_id"]))
	requestedThreadTS := strings.TrimSpace(stringValue(input["thread_ts"]))
	boundChannelID := strings.TrimSpace(bound.ChannelID)
	boundThreadTS := strings.TrimSpace(bound.ThreadTS)
	derivedThreadByChannel, derivedChannelByThread := derivedSlackThreadBindings(question, boundChannelID, boundThreadTS, bound.EntityRefs)

	switch {
	case requestedChannelID != "" && requestedThreadTS != "":
		if requestedChannelID != boundChannelID && requestedThreadTS == boundThreadTS {
			if derivedThread, ok := derivedThreadByChannel[requestedChannelID]; !ok || derivedThread != requestedThreadTS {
				return requestedChannelID, ""
			}
		}
		return requestedChannelID, requestedThreadTS
	case requestedChannelID != "":
		if requestedChannelID == boundChannelID {
			return requestedChannelID, boundThreadTS
		}
		if derivedThread, ok := derivedThreadByChannel[requestedChannelID]; ok {
			return requestedChannelID, derivedThread
		}
		return requestedChannelID, ""
	case requestedThreadTS != "":
		if requestedThreadTS == boundThreadTS {
			return boundChannelID, requestedThreadTS
		}
		if derivedChannel, ok := derivedChannelByThread[requestedThreadTS]; ok {
			return derivedChannel, requestedThreadTS
		}
		return "", requestedThreadTS
	default:
		return boundChannelID, boundThreadTS
	}
}

func slackPromptContext(values ...string) string {
	parts := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		parts = append(parts, value)
	}
	return strings.Join(parts, "\n")
}

func derivedSlackThreadBindings(question string, ingressChannelID string, ingressThreadTS string, entityRefs []slackpkg.EntityRef) (map[string]string, map[string]string) {
	threadByChannel := map[string]string{}
	channelByThread := map[string]string{}
	ambiguousChannels := map[string]struct{}{}
	ambiguousThreads := map[string]struct{}{}
	for _, item := range workflowplan.CandidateReadSurfacesForContext(workflowplan.RequestContext{
		Question:   question,
		ChannelID:  ingressChannelID,
		ThreadTS:   ingressThreadTS,
		EntityRefs: entityRefs,
	}) {
		channelID := strings.TrimSpace(item.ChannelID)
		threadTS := strings.TrimSpace(item.ThreadTS)
		if channelID == "" || threadTS == "" {
			continue
		}
		if _, blocked := ambiguousChannels[channelID]; !blocked {
			if existing, ok := threadByChannel[channelID]; ok && existing != threadTS {
				delete(threadByChannel, channelID)
				ambiguousChannels[channelID] = struct{}{}
			} else {
				threadByChannel[channelID] = threadTS
			}
		}
		if _, blocked := ambiguousThreads[threadTS]; !blocked {
			if existing, ok := channelByThread[threadTS]; ok && existing != channelID {
				delete(channelByThread, threadTS)
				ambiguousThreads[threadTS] = struct{}{}
			} else {
				channelByThread[threadTS] = channelID
			}
		}
	}
	return threadByChannel, channelByThread
}

type boundSlackContext struct {
	ChannelID   string
	ThreadTS    string
	Prompt      string
	ActionToken string
	EntityRefs  []slackpkg.EntityRef
}

func (s *Service) slackSearch(input map[string]interface{}) storepkg.ToolResult {
	query := firstNonEmpty(strings.TrimSpace(stringValue(input["query"])), strings.TrimSpace(stringValue(input["question"])))
	traceID := strings.TrimSpace(stringValue(input["trace_id"]))
	bound := s.boundSlackContext(traceID)
	contextualChannelIDs := slackMentionChannelIDs(bound.Prompt, bound.ChannelID, bound.ThreadTS, bound.EntityRefs, s.cfg.AllowedSlackChannelIDs)
	channelIDs := slackSearchChannelIDs(input, bound.ChannelID, s.cfg.AllowedSlackChannelIDs, contextualChannelIDs)
	limit := slackSearchLimit(input)
	since, until, err := parseActivityWindow(input)
	output := map[string]interface{}{
		"query":                 query,
		"channel_ids":           channelIDs,
		"limit":                 limit,
		"messages":              []map[string]interface{}{},
		"search_api":            "assistant.search.context",
		"action_token_present":  strings.TrimSpace(bound.ActionToken) != "",
		"action_token_required": false,
		"search_auth_mode":      "",
	}
	if strings.TrimSpace(query) == "" {
		output["error"] = "missing_query"
		return s.failedResult("slack.search", input, "slack", "Slack search requires query or question.", output)
	}
	if len(channelIDs) == 0 {
		output["error"] = "missing_channel_ids"
		return s.failedResult("slack.search", input, "slack", "Slack search requires an explicit, bound, or referenced channel.", output)
	}
	for _, channelID := range channelIDs {
		if !slackChannelAllowed(channelID, bound.ChannelID, contextualChannelIDs, s.cfg.AllowedSlackChannelIDs) {
			output["error"] = "channel_not_allowed"
			return s.failedResult("slack.search", input, "slack", fmt.Sprintf("Slack search not permitted for channel %s.", channelID), output)
		}
	}
	if err != nil {
		output["error"] = err.Error()
		return s.failedResult("slack.search", input, "slack", fmt.Sprintf("Slack search request invalid: %v", err), output)
	}
	output["since"] = since.UTC().Format(time.RFC3339)
	output["until"] = until.UTC().Format(time.RFC3339)
	searchToken := ""
	actionToken := ""
	switch {
	case strings.TrimSpace(s.cfg.SlackUserToken) != "":
		searchToken = strings.TrimSpace(s.cfg.SlackUserToken)
		output["search_auth_mode"] = "user"
	case strings.TrimSpace(s.cfg.SlackBotToken) != "":
		searchToken = strings.TrimSpace(s.cfg.SlackBotToken)
		output["search_auth_mode"] = "bot"
		output["action_token_required"] = true
		actionToken = strings.TrimSpace(bound.ActionToken)
	default:
		output["error"] = "missing_slack_bot_token"
		return s.unavailableResult("slack.search", input, "slack", "Slack search unavailable: bot token is not configured.", output)
	}
	if output["search_auth_mode"] == "bot" && actionToken == "" {
		output["error"] = "missing_action_token"
		return s.unavailableResult("slack.search", input, "slack", "Slack search unavailable: the bound Slack event did not include an action token.", output)
	}

	messages, nextCursor, err := s.slackAssistantSearchMessages(query, searchToken, actionToken, channelIDs, since, until, limit)
	if err != nil {
		output["error"] = err.Error()
		return s.failedResult("slack.search", input, "slack", fmt.Sprintf("Slack search failed: %v", err), output)
	}
	output["messages"] = messages
	output["next_cursor"] = nextCursor
	output["search_total"] = len(messages)
	summary := fmt.Sprintf("Slack search found %d matching message(s) for %q across %d governed channel(s).", len(messages), query, len(channelIDs))
	return s.result("slack.search", input, summary, output, nil)
}

func slackChannelAllowed(channelID string, boundChannelID string, contextualChannelIDs []string, configured []string) bool {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return false
	}
	if channelID == strings.TrimSpace(boundChannelID) {
		return true
	}
	for _, allowed := range contextualChannelIDs {
		if channelID == strings.TrimSpace(allowed) {
			return true
		}
	}
	for _, allowed := range configured {
		if channelID == strings.TrimSpace(allowed) {
			return true
		}
	}
	return false
}

func (s *Service) boundSlackContext(traceID string) boundSlackContext {
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return boundSlackContext{}
	}
	trace, ok := s.store.GetTrace(traceID)
	if !ok {
		return boundSlackContext{}
	}
	ingestionID := strings.TrimSpace(trace.Summary.IngestionID)
	if ingestionID == "" {
		return boundSlackContext{}
	}
	ctx := boundSlackContext{}
	eventID := ""
	for _, item := range s.store.ListIngestions() {
		if strings.TrimSpace(item.ID) != ingestionID {
			continue
		}
		ctx = boundSlackContext{
			ChannelID:  strings.TrimSpace(item.ChannelID),
			ThreadTS:   strings.TrimSpace(item.ThreadTS),
			Prompt:     strings.TrimSpace(item.Text),
			EntityRefs: append([]slackpkg.EntityRef(nil), item.EntityRefs...),
		}
		eventID = strings.TrimSpace(item.EventID)
		break
	}
	if eventID == "" {
		return ctx
	}
	for _, event := range s.store.ListEvents() {
		if strings.TrimSpace(event.ID) != eventID {
			continue
		}
		ctx.ActionToken = strings.TrimSpace(stringValue(event.Metadata["action_token"]))
		break
	}
	return ctx
}

func (s *Service) slackSearchAuthMode() string {
	if strings.TrimSpace(s.cfg.SlackUserToken) != "" {
		return "user"
	}
	if strings.TrimSpace(s.cfg.SlackBotToken) != "" {
		return "bot"
	}
	return "unconfigured"
}

func (s *Service) slackAssistantSearchMessages(query string, authToken string, actionToken string, channelIDs []string, since time.Time, until time.Time, limit int) ([]map[string]interface{}, string, error) {
	limit = maxInt(1, minInt(limit, 20))
	deduped := map[string]map[string]interface{}{}
	nextCursor := ""
	for _, channelID := range uniqueStrings(channelIDs) {
		messages, cursor, err := s.slackAssistantSearchChannel(query, authToken, actionToken, channelID, since, until, limit)
		if err != nil {
			return nil, "", err
		}
		if nextCursor == "" {
			nextCursor = cursor
		}
		for _, item := range messages {
			key := strings.TrimSpace(stringValue(item["channel_id"])) + ":" + strings.TrimSpace(stringValue(item["ts"]))
			if key == ":" {
				continue
			}
			deduped[key] = item
		}
	}
	matched := make([]map[string]interface{}, 0, len(deduped))
	for _, item := range deduped {
		matched = append(matched, item)
	}
	sort.Slice(matched, func(i, j int) bool {
		left, leftErr := parseFlexibleTimestamp(stringValue(matched[i]["ts"]))
		right, rightErr := parseFlexibleTimestamp(stringValue(matched[j]["ts"]))
		if leftErr != nil && rightErr != nil {
			return stringValue(matched[i]["ts"]) > stringValue(matched[j]["ts"])
		}
		if leftErr != nil {
			return false
		}
		if rightErr != nil {
			return true
		}
		return left.After(right)
	})
	if len(matched) > limit {
		matched = matched[:limit]
	}
	return matched, nextCursor, nil
}

func (s *Service) slackAssistantSearchChannel(query string, authToken string, actionToken string, channelID string, since time.Time, until time.Time, limit int) ([]map[string]interface{}, string, error) {
	payload := map[string]interface{}{
		"query":                    query,
		"context_channel_id":       channelID,
		"channel_types":            []string{"public_channel", "private_channel", "mpim", "im"},
		"content_types":            []string{"messages"},
		"include_context_messages": true,
		"include_bots":             true,
		"sort":                     "timestamp",
		"sort_dir":                 "desc",
		"after":                    since.Unix(),
		"before":                   until.Unix(),
		"limit":                    maxInt(1, minInt(limit, 20)),
	}
	if strings.TrimSpace(actionToken) != "" {
		payload["action_token"] = actionToken
	}
	var response map[string]interface{}
	if err := s.apiJSON("POST", s.slackMethodURL("assistant.search.context"), payload, map[string]string{
		"Authorization": "Bearer " + authToken,
		"Content-Type":  "application/json",
	}, &response); err != nil {
		return nil, "", err
	}
	if !boolValue(response["ok"]) {
		return nil, "", fmt.Errorf("assistant.search.context: %s", firstNonEmpty(strings.TrimSpace(stringValue(response["error"])), "request failed"))
	}
	messages := assistantSearchMessagesPayload(response)
	filtered := make([]map[string]interface{}, 0, len(messages))
	for _, item := range messages {
		if strings.TrimSpace(channelID) != "" && strings.TrimSpace(stringValue(item["channel_id"])) != strings.TrimSpace(channelID) {
			continue
		}
		filtered = append(filtered, item)
	}
	nextCursor := firstNonEmpty(
		strings.TrimSpace(stringValue(response["next_cursor"])),
		strings.TrimSpace(stringValueFromMap(mapValue(response["response_metadata"]), "next_cursor")),
	)
	return filtered, nextCursor, nil
}

func assistantSearchMessagesPayload(response map[string]interface{}) []map[string]interface{} {
	results := mapValue(response["results"])
	rawMessages, ok := results["messages"].([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}
	messages := make([]map[string]interface{}, 0, len(rawMessages))
	for _, raw := range rawMessages {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		channelID := firstNonEmpty(
			strings.TrimSpace(stringValue(item["channel_id"])),
			strings.TrimSpace(stringValueFromMap(mapValue(item["channel"]), "id")),
		)
		ts := firstNonEmpty(
			strings.TrimSpace(stringValue(item["message_timestamp"])),
			strings.TrimSpace(stringValue(item["ts"])),
			strings.TrimSpace(stringValue(item["timestamp"])),
		)
		if !isLikelySlackChannelID(channelID) || ts == "" {
			continue
		}
		message := map[string]interface{}{
			"channel_id":     channelID,
			"author_user_id": strings.TrimSpace(stringValue(item["author_user_id"])),
			"author_name":    strings.TrimSpace(stringValue(item["author_name"])),
			"ts":             ts,
			"text":           truncate(firstNonEmpty(strings.TrimSpace(stringValue(item["text"])), strings.TrimSpace(stringValue(item["message_text"]))), 2000),
			"permalink":      strings.TrimSpace(stringValue(item["permalink"])),
			"is_author_bot":  boolValue(item["is_author_bot"]),
		}
		messages = append(messages, message)
	}
	return messages
}

func (s *Service) slackMethodURL(method string) string {
	return strings.TrimRight(strings.TrimSpace(s.slackAPIURL), "/") + "/" + strings.TrimLeft(strings.TrimSpace(method), "/")
}

func slackHistoryScope(input map[string]interface{}, question string, threadTS string) string {
	explicit := strings.ToLower(strings.TrimSpace(stringValue(input["scope"])))
	if explicit == "thread" && strings.TrimSpace(threadTS) != "" {
		return "thread"
	}
	if explicit == "channel" {
		return "channel"
	}
	if strings.TrimSpace(threadTS) == "" {
		return "channel"
	}
	_ = question
	return "thread"
}

func slackHistoryLimit(input map[string]interface{}) int {
	switch raw := input["limit"].(type) {
	case int:
		return maxInt(1, minInt(raw, 100))
	case int32:
		return maxInt(1, minInt(int(raw), 100))
	case int64:
		return maxInt(1, minInt(int(raw), 100))
	case float64:
		return maxInt(1, minInt(int(raw), 100))
	case json.Number:
		if value, err := raw.Int64(); err == nil {
			return maxInt(1, minInt(int(value), 100))
		}
	case string:
		if value, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
			return maxInt(1, minInt(value, 100))
		}
	}
	return 25
}

func slackHistoryWindow(input map[string]interface{}) (string, string, error) {
	oldestRaw := firstNonEmpty(strings.TrimSpace(stringValue(input["oldest"])), strings.TrimSpace(stringValue(input["since"])))
	latestRaw := firstNonEmpty(strings.TrimSpace(stringValue(input["latest"])), strings.TrimSpace(stringValue(input["until"])))
	if oldestRaw == "" {
		oldestRaw = time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	}
	if latestRaw == "" {
		latestRaw = time.Now().UTC().Format(time.RFC3339)
	}
	oldest, err := parseFlexibleTimestamp(oldestRaw)
	if err != nil {
		return "", "", fmt.Errorf("invalid oldest timestamp %q", oldestRaw)
	}
	latest, err := parseFlexibleTimestamp(latestRaw)
	if err != nil {
		return "", "", fmt.Errorf("invalid latest timestamp %q", latestRaw)
	}
	if latest.Before(oldest) {
		return "", "", fmt.Errorf("latest must be after oldest")
	}
	return slackTimestamp(oldest), slackTimestamp(latest), nil
}

func parseFlexibleTimestamp(raw string) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, fmt.Errorf("timestamp is empty")
	}
	if parsed, err := time.Parse(time.RFC3339, text); err == nil {
		return parsed.UTC(), nil
	}
	if sec, err := strconv.ParseInt(text, 10, 64); err == nil {
		return time.Unix(sec, 0).UTC(), nil
	}
	parts := strings.SplitN(text, ".", 2)
	if len(parts) == 2 {
		sec, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		fraction := parts[1]
		if fraction == "" {
			return time.Unix(sec, 0).UTC(), nil
		}
		if len(fraction) > 9 {
			fraction = fraction[:9]
		}
		for len(fraction) < 9 {
			fraction += "0"
		}
		nsec, err := strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(sec, nsec).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp format")
}

func slackTimestamp(value time.Time) string {
	value = value.UTC()
	return fmt.Sprintf("%d.%06d", value.Unix(), value.Nanosecond()/1_000)
}

func slackMessagesPayload(messages []slackapi.Message) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(messages))
	for _, message := range messages {
		out = append(out, map[string]interface{}{
			"ts":          strings.TrimSpace(message.Timestamp),
			"thread_ts":   strings.TrimSpace(message.ThreadTimestamp),
			"user":        strings.TrimSpace(message.User),
			"username":    strings.TrimSpace(message.Username),
			"bot_id":      strings.TrimSpace(message.BotID),
			"subtype":     strings.TrimSpace(message.SubType),
			"text":        truncate(strings.TrimSpace(message.Text), 2000),
			"reply_count": message.ReplyCount,
		})
	}
	return out
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

func matchingDeploymentTargets(deployment appsv1.Deployment, targets []string) []string {
	matched := make([]string, 0, len(targets))
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		if matchesDeploymentTarget(deployment, target) {
			matched = append(matched, target)
		}
	}
	return uniqueStrings(matched)
}

func matchesDeploymentTarget(deployment appsv1.Deployment, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return true
	}
	if strings.Contains(strings.ToLower(deployment.Name), target) {
		return true
	}
	for key, value := range deployment.Labels {
		if strings.Contains(strings.ToLower(key), target) || strings.Contains(strings.ToLower(value), target) {
			return true
		}
	}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if strings.Contains(strings.ToLower(container.Name), target) || strings.Contains(strings.ToLower(container.Image), target) {
			return true
		}
	}
	return false
}

func deploymentImages(deployment appsv1.Deployment) []string {
	images := make([]string, 0, len(deployment.Spec.Template.Spec.Containers))
	for _, container := range deployment.Spec.Template.Spec.Containers {
		images = append(images, strings.TrimSpace(container.Image))
	}
	sort.Strings(images)
	return uniqueStrings(images)
}

func deploymentCondition(deployment appsv1.Deployment, targetType appsv1.DeploymentConditionType) (appsv1.DeploymentCondition, bool) {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == targetType {
			return condition, true
		}
	}
	return appsv1.DeploymentCondition{}, false
}

func runtimeDeploymentTargets(input map[string]interface{}) []string {
	targets := make([]string, 0, 8)
	targets = append(targets, stringSliceValue(input["services"])...)
	if service := strings.TrimSpace(stringValue(input["service"])); service != "" {
		targets = append(targets, service)
	}
	if len(targets) == 0 {
		targets = append(targets,
			"control-plane",
			"tool-gateway",
			"runner-prod",
			"runner-proactive",
			"runner-eval",
			"runner-proposal",
			"honcho",
		)
	}
	normalized := make([]string, 0, len(targets))
	for _, target := range targets {
		target = strings.ToLower(strings.TrimSpace(target))
		target = strings.NewReplacer("_", "-", " ", "-").Replace(target)
		target = strings.Trim(target, "-")
		if target == "" {
			continue
		}
		normalized = append(normalized, target)
	}
	return uniqueStrings(normalized)
}

func runtimeTimeoutsSeconds(cfg config.Config) map[string]int {
	return map[string]int{
		"prod":      int(cfg.RunnerTimeoutForRole("prod").Seconds()),
		"proactive": int(cfg.RunnerTimeoutForRole("proactive").Seconds()),
		"eval":      int(cfg.RunnerTimeoutForRole("eval").Seconds()),
		"proposal":  int(cfg.RunnerTimeoutForRole("proposal").Seconds()),
	}
}

func runtimeTaskTimeoutsSeconds(cfg config.Config) map[string]int {
	return map[string]int{
		"prod":      int(cfg.RunnerTaskTimeoutForRole("prod").Seconds()),
		"proactive": int(cfg.RunnerTaskTimeoutForRole("proactive").Seconds()),
		"eval":      int(cfg.RunnerTaskTimeoutForRole("eval").Seconds()),
		"proposal":  int(cfg.RunnerTaskTimeoutForRole("proposal").Seconds()),
	}
}

func slackSearchChannelIDs(input map[string]interface{}, boundChannelID string, allowed []string, contextual []string) []string {
	channelIDs := make([]string, 0, len(allowed)+1)
	channelIDs = append(channelIDs, stringSliceValue(input["channel_ids"])...)
	if channelID := strings.TrimSpace(stringValue(input["channel_id"])); channelID != "" {
		channelIDs = append(channelIDs, channelID)
	}
	if len(channelIDs) == 0 {
		if channelID := strings.TrimSpace(boundChannelID); channelID != "" {
			channelIDs = append(channelIDs, channelID)
		}
		channelIDs = append(channelIDs, contextual...)
	}
	filtered := make([]string, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		channelID = strings.ToUpper(strings.TrimSpace(channelID))
		if !isLikelySlackChannelID(channelID) {
			continue
		}
		filtered = append(filtered, channelID)
	}
	return uniqueStrings(filtered)
}

func slackMentionChannelIDs(question string, ingressChannelID string, ingressThreadTS string, entityRefs []slackpkg.EntityRef, allowed []string) []string {
	if !containsString(allowed, slackMentionsOnlySentinel) {
		return nil
	}
	hints := workflowplan.CandidateReadSurfacesForContext(workflowplan.RequestContext{
		Question:   question,
		ChannelID:  ingressChannelID,
		ThreadTS:   ingressThreadTS,
		EntityRefs: entityRefs,
	})
	out := make([]string, 0, len(hints))
	for _, item := range hints {
		channelID := strings.TrimSpace(item.ChannelID)
		if channelID == "" || channelID == strings.TrimSpace(ingressChannelID) {
			continue
		}
		out = append(out, channelID)
	}
	return uniqueStrings(out)
}

func slackSearchLimit(input map[string]interface{}) int {
	switch raw := input["limit"].(type) {
	case int:
		return maxInt(1, minInt(raw, 50))
	case int32:
		return maxInt(1, minInt(int(raw), 50))
	case int64:
		return maxInt(1, minInt(int(raw), 50))
	case float64:
		return maxInt(1, minInt(int(raw), 50))
	case json.Number:
		if value, err := raw.Int64(); err == nil {
			return maxInt(1, minInt(int(value), 50))
		}
	case string:
		if value, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
			return maxInt(1, minInt(value, 50))
		}
	}
	return 10
}

func stringSliceValue(value interface{}) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []interface{}:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := strings.TrimSpace(stringValue(item)); text != "" {
				out = append(out, text)
			}
		}
		return out
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return nil
		}
		if !strings.Contains(text, ",") {
			return []string{text}
		}
		parts := strings.Split(text, ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			if item := strings.TrimSpace(part); item != "" {
				out = append(out, item)
			}
		}
		return out
	default:
		return nil
	}
}

func isLikelySlackChannelID(value string) bool {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) < 2 {
		return false
	}
	switch value[0] {
	case 'C', 'D', 'G':
		for _, r := range value[1:] {
			if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
				return false
			}
		}
		return true
	default:
		return false
	}
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

func mapValue(value interface{}) map[string]interface{} {
	if typed, ok := value.(map[string]interface{}); ok {
		return typed
	}
	return map[string]interface{}{}
}

func intValue(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		item, _ := typed.Int64()
		return int(item)
	case string:
		item, _ := strconv.Atoi(strings.TrimSpace(typed))
		return item
	default:
		return 0
	}
}

func stringValueFromMap(values map[string]interface{}, key string) string {
	if values == nil {
		return ""
	}
	return stringValue(values[key])
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

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
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
	case "slack.upload_file":
		return firstNonEmpty(stringValue(output["file_id"]), stringValue(output["permalink"]), stringValue(output["thread_ts"]))
	case "slack.history":
		return firstNonEmpty(stringValue(output["thread_ts"]), stringValue(output["channel_id"]))
	case "slack.search":
		return stringValue(output["query"])
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
	case "workspace.git_history":
		return stringValue(output["workspace_id"])
	case "workspace.git_show":
		return stringValue(output["workspace_id"])
	case "workspace.git_search":
		return stringValue(output["workspace_id"])
	case "rsi.runtime_deployment_facts":
		return stringValue(output["namespace"])
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
