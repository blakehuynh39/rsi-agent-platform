package workflowplan

import (
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
)

type RuntimeConfig struct {
	DefaultRepo      string
	AllowedRepos     []string
	KnowledgeBaseURL string
	SandboxNamespace string
}

type RequestContext struct {
	Trace          events.TraceSummary
	WorkflowID     string
	ConversationID string
	CaseID         string
	WorkflowKind   string
	AssignedBot    string
	Question       string
	ChannelID      string
	ThreadTS       string
}

func ToolPlan(intent string, question string, repo string, channelID string, threadTS string) []string {
	var plan []string
	switch strings.TrimSpace(intent) {
	case "incident":
		plan = []string{"sentry.lookup", "kubernetes.inspect", "rsi.workflow_context", "rsi.action_chain", "rsi.runtime_health"}
	case "feature_request":
		plan = []string{"repo.context", "github.repo_context", "rsi.workflow_context", "rsi.action_chain"}
	default:
		plan = []string{"repo.context", "knowledge.context", "rsi.workflow_context", "rsi.action_chain"}
		if ShouldUseGitHubRepoActivity(question, repo) {
			plan = append(plan, "github.repo_activity")
		}
	}
	if ShouldUseSlackSearch(question, channelID) {
		plan = append(plan, "slack.search")
	}
	if ShouldUseSlackHistory(question, repo, channelID, threadTS) {
		plan = append(plan, "slack.history")
	}
	if ShouldUseRuntimeDeploymentFacts(question) {
		plan = append(plan, "rsi.runtime_deployment_facts")
	}
	return plan
}

func BuildToolRequestPayload(cfg RuntimeConfig, ctx RequestContext, now time.Time) map[string]any {
	repo := ResolveTargetRepo(cfg, ctx.Question)
	since, until := RepoActivityWindow(ctx.Question, now)
	workflowID := ctx.Trace.WorkflowID
	if workflowID == "" {
		workflowID = ctx.WorkflowID
	}
	conversationID := ctx.Trace.ConversationID
	if conversationID == "" {
		conversationID = ctx.ConversationID
	}
	caseID := ctx.Trace.CaseID
	if caseID == "" {
		caseID = ctx.CaseID
	}
	return map[string]any{
		"repo":               repo,
		"question":           ctx.Question,
		"topic":              ctx.Question,
		"scope_id":           repo,
		"service":            ctx.AssignedBot,
		"alert":              ctx.Question,
		"namespace":          cfg.SandboxNamespace,
		"target":             ctx.WorkflowKind,
		"knowledge_base_url": cfg.KnowledgeBaseURL,
		"channel_id":         ctx.ChannelID,
		"thread_ts":          ctx.ThreadTS,
		"trace_id":           ctx.Trace.TraceID,
		"workflow_id":        workflowID,
		"conversation_id":    conversationID,
		"case_id":            caseID,
		"since":              since,
		"until":              until,
	}
}

func ResolveTargetRepo(cfg RuntimeConfig, question string) string {
	text := strings.ToLower(strings.TrimSpace(question))
	for _, repo := range cfg.AllowedRepos {
		repo = strings.TrimSpace(repo)
		if repo == "" {
			continue
		}
		if strings.Contains(text, strings.ToLower(repo)) {
			return repo
		}
	}
	return strings.TrimSpace(cfg.DefaultRepo)
}

func ShouldUseGitHubRepoActivity(question string, repo string) bool {
	if strings.TrimSpace(repo) == "" || strings.EqualFold(strings.TrimSpace(repo), "cloudflare") {
		return false
	}
	text := strings.ToLower(strings.TrimSpace(question))
	if text == "" {
		return false
	}
	indicators := []string{
		"progress",
		"activity",
		"recent",
		"last week",
		"past week",
		"this week",
		"today",
		"yesterday",
		"commits",
		"prs",
		"pull requests",
		"merged",
		"opened",
	}
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

func ShouldUseSlackHistory(question string, repo string, channelID string, threadTS string) bool {
	if strings.TrimSpace(channelID) == "" {
		return false
	}
	text := strings.ToLower(strings.TrimSpace(question))
	if text == "" {
		return strings.TrimSpace(threadTS) != ""
	}
	indicators := []string{
		"slack",
		"channel",
		"thread",
		"conversation",
		"convo",
		"discuss",
		"discussion",
		"talked",
		"said",
		"mention",
		"mentioned",
	}
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return ShouldUseGitHubRepoActivity(question, repo)
}

func ShouldUseSlackSearch(question string, channelID string) bool {
	if strings.TrimSpace(channelID) == "" {
		return false
	}
	text := strings.ToLower(strings.TrimSpace(question))
	if text == "" {
		return false
	}
	explicitPhrases := []string{
		"search slack",
		"did we discuss",
		"have we discussed",
		"have we talked",
		"where did we decide",
		"where was this decided",
		"find the thread",
		"find the conversation",
		"mentioned in slack",
		"discussed in slack",
	}
	for _, phrase := range explicitPhrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return strings.Contains(text, "search") && (strings.Contains(text, "channel") || strings.Contains(text, "thread") || strings.Contains(text, "slack"))
}

func ShouldUseRuntimeDeploymentFacts(question string) bool {
	text := strings.ToLower(strings.TrimSpace(question))
	if text == "" {
		return false
	}
	indicators := []string{
		"deployment",
		"deployments",
		"rollout",
		"image",
		"images",
		"tag",
		"timeout",
		"time out",
		"5 minute",
		"300s",
		"configured",
		"config",
		"control plane",
		"tool gateway",
		"runner",
		"honcho",
		"slack app",
		"allowed channel",
		"channel ids",
	}
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

func RepoActivityWindow(question string, now time.Time) (string, string) {
	text := strings.ToLower(strings.TrimSpace(question))
	start := now.Add(-7 * 24 * time.Hour)
	switch {
	case strings.Contains(text, "today"):
		start = now.Add(-24 * time.Hour)
	case strings.Contains(text, "yesterday"):
		start = now.Add(-48 * time.Hour)
	case strings.Contains(text, "last 24 hours"):
		start = now.Add(-24 * time.Hour)
	case strings.Contains(text, "last week"), strings.Contains(text, "past week"), strings.Contains(text, "this week"), strings.Contains(text, "recent"):
		start = now.Add(-7 * 24 * time.Hour)
	}
	return start.UTC().Format(time.RFC3339), now.UTC().Format(time.RFC3339)
}
