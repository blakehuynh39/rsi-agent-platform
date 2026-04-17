package workflowplan

import (
	"fmt"
	"regexp"
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

type SlackSurfaceHint struct {
	ChannelID string `json:"channel_id,omitempty"`
	ThreadTS  string `json:"thread_ts,omitempty"`
	Ref       string `json:"ref,omitempty"`
	Source    string `json:"source,omitempty"`
}

type LiveHintSet struct {
	Repo                  string             `json:"repo,omitempty"`
	PreferredTools        []string           `json:"preferred_tools,omitempty"`
	CandidateReadSurfaces []SlackSurfaceHint `json:"candidate_read_surfaces,omitempty"`
	Since                 string             `json:"since,omitempty"`
	Until                 string             `json:"until,omitempty"`
}

var (
	slackPermalinkPattern  = regexp.MustCompile(`https://[^\s>]+/archives/([A-Z0-9]+)/p(\d{10,})`)
	slackChannelTagPattern = regexp.MustCompile(`<#([A-Z0-9]+)(?:\|[^>]+)?>`)
	slackThreadTSPattern   = regexp.MustCompile(`(?:thread(?:_ts)?|ts)\D+(\d{10}\.\d{6})`)
)

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

func BuildLiveHints(cfg RuntimeConfig, ctx RequestContext, now time.Time) LiveHintSet {
	repo := ResolveTargetRepo(cfg, ctx.Question)
	since, until := RepoActivityWindow(ctx.Question, now)
	return LiveHintSet{
		Repo:                  repo,
		PreferredTools:        ToolPlan(ctx.WorkflowKind, ctx.Question, repo, ctx.ChannelID, ctx.ThreadTS),
		CandidateReadSurfaces: CandidateReadSurfaces(ctx.Question, ctx.ChannelID, ctx.ThreadTS),
		Since:                 since,
		Until:                 until,
	}
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

func CandidateReadSurfaces(question string, ingressChannelID string, ingressThreadTS string) []SlackSurfaceHint {
	seen := map[string]struct{}{}
	out := make([]SlackSurfaceHint, 0, 4)
	appendSurface := func(item SlackSurfaceHint) {
		item.ChannelID = strings.TrimSpace(item.ChannelID)
		item.ThreadTS = strings.TrimSpace(item.ThreadTS)
		item.Ref = strings.TrimSpace(item.Ref)
		item.Source = strings.TrimSpace(item.Source)
		if item.ChannelID == "" && item.ThreadTS == "" && item.Ref == "" {
			return
		}
		key := strings.Join([]string{item.ChannelID, item.ThreadTS, item.Ref}, "|")
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}

	if strings.TrimSpace(ingressChannelID) != "" {
		appendSurface(SlackSurfaceHint{
			ChannelID: ingressChannelID,
			ThreadTS:  ingressThreadTS,
			Source:    "ingress_thread",
			Ref:       slackSurfaceRef(ingressChannelID, ingressThreadTS),
		})
	}

	text := strings.TrimSpace(question)
	if text == "" {
		return out
	}

	for _, match := range slackPermalinkPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 3 {
			continue
		}
		appendSurface(SlackSurfaceHint{
			ChannelID: match[1],
			ThreadTS:  slackPermalinkThreadTS(match[2]),
			Ref:       match[0],
			Source:    "slack_permalink",
		})
	}
	for _, match := range slackChannelTagPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		appendSurface(SlackSurfaceHint{
			ChannelID: match[1],
			Ref:       slackSurfaceRef(match[1], ""),
			Source:    "channel_mention",
		})
	}
	for _, match := range slackThreadTSPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		appendSurface(SlackSurfaceHint{
			ChannelID: ingressChannelID,
			ThreadTS:  match[1],
			Ref:       slackSurfaceRef(ingressChannelID, match[1]),
			Source:    "explicit_thread_ref",
		})
	}
	return out
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

func slackPermalinkThreadTS(raw string) string {
	digits := strings.TrimSpace(raw)
	if len(digits) <= 6 {
		return digits
	}
	return fmt.Sprintf("%s.%s", digits[:len(digits)-6], digits[len(digits)-6:])
}

func slackSurfaceRef(channelID string, threadTS string) string {
	channelID = strings.TrimSpace(channelID)
	threadTS = strings.TrimSpace(threadTS)
	switch {
	case channelID != "" && threadTS != "":
		return fmt.Sprintf("slack://%s/%s", channelID, threadTS)
	case channelID != "":
		return fmt.Sprintf("slack://%s", channelID)
	case threadTS != "":
		return fmt.Sprintf("slack://thread/%s", threadTS)
	default:
		return ""
	}
}
