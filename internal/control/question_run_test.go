package control

import (
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/questionrun"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func TestBuildInvestigationSpecIncludesPlainSlackChannelMentions(t *testing.T) {
	cfg := config.Config{
		DefaultRepo:        "depin-backend",
		AllowedTargetRepos: []string{"depin-backend", "rsi-agent-platform"},
	}
	workflow := storepkg.Workflow{
		ID:             "workflow-123",
		ConversationID: "conv-123",
		CaseID:         "case-123",
		Kind:           "slack",
		AssignedBot:    "arch",
	}
	ingestion := slackpkg.Ingestion{
		Text:      "Hello @U0ASDQKU3UL, can you give me a quick rundown of how depin-backend api progressed in this past week in accordance with the numo project, is there any misalignments ? you can find the channel in #C0AKH5SNGKH and #C0AL7EKNHDF for latest discussions",
		ChannelID: "CINGRESS",
		ThreadTS:  "171000001.000100",
	}
	trace := events.Trace{
		Summary: events.TraceSummary{
			TraceID:        "trace-123",
			WorkflowID:     workflow.ID,
			ConversationID: workflow.ConversationID,
			CaseID:         workflow.CaseID,
		},
	}

	spec := buildInvestigationSpec(cfg, workflow, ingestion, trace)

	if spec.Repo != "depin-backend" {
		t.Fatalf("repo = %q, want depin-backend", spec.Repo)
	}
	if !spec.AlignmentRequired {
		t.Fatalf("expected alignment_required=true for misalignment question")
	}
	if !containsQuestionRunSurface(spec.ReadSurfaces, questionrun.SlackSurface{
		ChannelID: "CINGRESS",
		ThreadTS:  "171000001.000100",
		Source:    "ingress_thread",
	}) {
		t.Fatalf("expected ingress thread surface, got %#v", spec.ReadSurfaces)
	}
	if !containsQuestionRunSurface(spec.ReadSurfaces, questionrun.SlackSurface{
		ChannelID: "C0AKH5SNGKH",
		Source:    "channel_mention",
	}) {
		t.Fatalf("expected first plain channel mention surface, got %#v", spec.ReadSurfaces)
	}
	if !containsQuestionRunSurface(spec.ReadSurfaces, questionrun.SlackSurface{
		ChannelID: "C0AL7EKNHDF",
		Source:    "channel_mention",
	}) {
		t.Fatalf("expected second plain channel mention surface, got %#v", spec.ReadSurfaces)
	}
}

func TestBuildInvestigationSpecIncludesStructuredEntityRefsFromIngestion(t *testing.T) {
	cfg := config.Config{
		DefaultRepo:        "depin-backend",
		AllowedTargetRepos: []string{"depin-backend", "rsi-agent-platform"},
	}
	workflow := storepkg.Workflow{
		ID:             "workflow-123",
		ConversationID: "conv-123",
		CaseID:         "case-123",
		Kind:           "slack",
		AssignedBot:    "arch",
	}
	ingestion := slackpkg.Ingestion{
		Text:      "Give me a quick rundown of how depin-backend api progressed this past week with respect to NUMO.",
		ChannelID: "CINGRESS",
		ThreadTS:  "171000001.000100",
		EntityRefs: []slackpkg.EntityRef{
			{Kind: slackpkg.EntityChannel, ID: "C0AKH5SNGKH", Source: "mrkdwn"},
			{Kind: slackpkg.EntityChannel, ID: "C0AL7EKNHDF", Source: "mrkdwn"},
		},
	}
	trace := events.Trace{
		Summary: events.TraceSummary{
			TraceID:        "trace-123",
			WorkflowID:     workflow.ID,
			ConversationID: workflow.ConversationID,
			CaseID:         workflow.CaseID,
		},
	}

	spec := buildInvestigationSpec(cfg, workflow, ingestion, trace)

	if !containsQuestionRunSurface(spec.ReadSurfaces, questionrun.SlackSurface{
		ChannelID: "C0AKH5SNGKH",
		Source:    "entity_ref",
	}) {
		t.Fatalf("expected first structured entity-ref surface, got %#v", spec.ReadSurfaces)
	}
	if !containsQuestionRunSurface(spec.ReadSurfaces, questionrun.SlackSurface{
		ChannelID: "C0AL7EKNHDF",
		Source:    "entity_ref",
	}) {
		t.Fatalf("expected second structured entity-ref surface, got %#v", spec.ReadSurfaces)
	}
}

func TestDeriveOpenQuestionsRequiresRepoAndReferencedChannelCoverage(t *testing.T) {
	spec := questionrun.InvestigationSpec{
		UserRequest: "How did depin-backend progress this week in accordance with the numo project?",
		ReplyTarget: questionrun.ReplyTarget{
			ChannelID: "CINGRESS",
			ThreadTS:  "171000001.000100",
		},
		Repo:              "depin-backend",
		ProjectKey:        "numo",
		AlignmentRequired: true,
		ReadSurfaces: []questionrun.SlackSurface{
			{ChannelID: "CINGRESS", ThreadTS: "171000001.000100", Source: "ingress_thread"},
			{ChannelID: "C0AKH5SNGKH", Source: "channel_mention"},
			{ChannelID: "C0AL7EKNHDF", Source: "channel_mention"},
		},
	}
	ledger := questionrun.EvidenceLedger{
		AlignmentRequired: true,
		AlignmentLedger: &questionrun.ProjectAlignmentLedger{
			ProjectKey: "numo",
		},
		EvidenceItems: []questionrun.EvidenceItem{
			{
				Kind:      "slack_search_match",
				Summary:   "Ingress thread mentions a weekly summary is needed.",
				ToolName:  "slack.search",
				ChannelID: "CINGRESS",
				ThreadTS:  "171000001.000100",
			},
		},
	}

	openQuestions := deriveOpenQuestions(spec, ledger)

	if !containsStringExact(openQuestions, "Need repository activity evidence for depin-backend before reducing the reply.") {
		t.Fatalf("expected repo evidence gap, got %#v", openQuestions)
	}
	if !containsStringExact(openQuestions, "Need Slack discussion evidence from referenced channel C0AKH5SNGKH before reducing the reply.") {
		t.Fatalf("expected first referenced channel gap, got %#v", openQuestions)
	}
	if !containsStringExact(openQuestions, "Need Slack discussion evidence from referenced channel C0AL7EKNHDF before reducing the reply.") {
		t.Fatalf("expected second referenced channel gap, got %#v", openQuestions)
	}
}

func TestBuildQuestionReduceTaskUsesSlackMCPReplyProfileWhenReplyAllowed(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}

	task := buildQuestionReduceTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		DefaultReasoningVerbosity: "verbose",
	}, store, questionRunContext{
		workflowContext: ctx,
		questionRun: storepkg.QuestionRun{
			Role: "prod",
			InvestigationSpec: questionrun.InvestigationSpec{
				UserRequest: "Summarize progress and reply in Slack.",
				Repo:        "depin-backend",
			},
			RunnerDiagnostics: map[string]any{},
		},
	})

	if len(task.MCPServers) != 1 {
		t.Fatalf("expected one MCP server, got %#v", task.MCPServers)
	}
	if task.MCPServers[0].Profile != "slack_mcp_reply" {
		t.Fatalf("expected slack_mcp_reply profile, got %#v", task.MCPServers)
	}
	if !containsStringExact(task.AllowedTools, "repo.context") {
		t.Fatalf("expected repo.context to remain available, got %#v", task.AllowedTools)
	}
	if containsStringExact(task.AllowedTools, "slack.history") || containsStringExact(task.AllowedTools, "slack.search") || containsStringExact(task.AllowedTools, "slack.reply") {
		t.Fatalf("expected legacy Slack tools to be absent, got %#v", task.AllowedTools)
	}
	if !strings.Contains(task.SystemMessage, "send exactly one reply") {
		t.Fatalf("expected reply-capable MCP system prompt, got %q", task.SystemMessage)
	}
	if task.MCPServers[0].Headers["X-RSI-Channel-ID"] != ctx.ingestion.ChannelID {
		t.Fatalf("expected bound channel header, got %#v", task.MCPServers[0].Headers)
	}
}

func TestBuildQuestionReduceTaskUsesReadOnlySlackMCPWhenReplyBlocked(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineThreadPolicy,
		AggregateID: ctx.trace.Summary.ThreadKey,
		CommandKind: string(transition.CommandThreadMute),
		CommandID:   "cmd-test-question-run-thread-mute",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"owner_bot": "tester",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(thread_mute) error = %v", err)
	}

	task := buildQuestionReduceTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		DefaultReasoningVerbosity: "verbose",
	}, store, questionRunContext{
		workflowContext: ctx,
		questionRun: storepkg.QuestionRun{
			Role: "prod",
			InvestigationSpec: questionrun.InvestigationSpec{
				UserRequest: "Summarize progress but do not post if policy blocks it.",
				Repo:        "depin-backend",
			},
			RunnerDiagnostics: map[string]any{},
		},
	})

	if len(task.MCPServers) != 1 {
		t.Fatalf("expected one MCP server, got %#v", task.MCPServers)
	}
	if task.MCPServers[0].Profile != "slack_mcp_read" {
		t.Fatalf("expected slack_mcp_read profile when blocked, got %#v", task.MCPServers)
	}
	if task.AllowedTools == nil || !containsStringExact(task.AllowedTools, "repo.context") {
		t.Fatalf("expected governed repo reads to remain available, got %#v", task.AllowedTools)
	}
	if !strings.Contains(task.SystemMessage, "Slack posting is blocked by policy") {
		t.Fatalf("expected blocked-posting MCP system prompt, got %q", task.SystemMessage)
	}
}

func containsQuestionRunSurface(surfaces []questionrun.SlackSurface, target questionrun.SlackSurface) bool {
	for _, item := range surfaces {
		if item.ChannelID == target.ChannelID && item.ThreadTS == target.ThreadTS && item.Source == target.Source {
			return true
		}
	}
	return false
}

func containsStringExact(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
