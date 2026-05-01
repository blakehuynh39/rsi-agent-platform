package control

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/questionrun"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
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

func TestBuildQuestionReduceTaskDisablesMCPAndTools(t *testing.T) {
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
		ProdRunnerTaskTimeout:     5 * time.Minute,
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
	}, queue.WorkflowQueue)

	if len(task.MCPServers) != 0 {
		t.Fatalf("expected no MCP servers, got %#v", task.MCPServers)
	}
	if len(task.AllowedTools) != 0 {
		t.Fatalf("expected no tools in reduce phase, got %#v", task.AllowedTools)
	}
	if !strings.Contains(task.SystemMessage, "Do not call tools") {
		t.Fatalf("expected no-tools reducer prompt, got %q", task.SystemMessage)
	}
	if task.TimeoutSeconds != 30 {
		t.Fatalf("expected 30 second reducer reserve, got %d", task.TimeoutSeconds)
	}
}

func TestBuildQuestionReduceTaskKeepsReducerNoToolsWhenReplyBlocked(t *testing.T) {
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
		ProdRunnerTaskTimeout:     5 * time.Minute,
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
	}, queue.WorkflowQueue)

	if len(task.MCPServers) != 0 {
		t.Fatalf("expected no MCP servers, got %#v", task.MCPServers)
	}
	if len(task.AllowedTools) != 0 {
		t.Fatalf("expected no reducer tools when reply is blocked, got %#v", task.AllowedTools)
	}
	if !strings.Contains(task.SystemMessage, "Do not send Slack messages") {
		t.Fatalf("expected deterministic control-plane delivery prompt, got %q", task.SystemMessage)
	}
}

func TestIsQuestionRunBoundedStopIncludesOutputTokenBudgetExhaustion(t *testing.T) {
	if !isQuestionRunBoundedStop("output_token_budget_exhausted") {
		t.Fatal("expected output_token_budget_exhausted to be treated as a bounded stop")
	}
}

func TestBuildQuestionGatherTaskUsesFullReasoningWindowAndTrimsNoisyTools(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}

	task := buildQuestionGatherTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		DefaultReasoningVerbosity: "verbose",
		ProdRunnerTaskTimeout:     5 * time.Minute,
	}, store, questionRunContext{
		workflowContext: ctx,
		questionRun: storepkg.QuestionRun{
			Role: "prod",
			InvestigationSpec: questionrun.InvestigationSpec{
				UserRequest:       "Summarize depin-backend progress this week in accordance with numo.",
				ReplyTarget:       questionrun.ReplyTarget{ChannelID: "CINGRESS", ThreadTS: "171000001.000100"},
				Repo:              "depin-backend",
				ProjectKey:        "numo",
				Since:             "2026-04-11T00:00:00Z",
				Until:             "2026-04-18T00:00:00Z",
				AlignmentRequired: true,
				RetrievalBudget:   20,
				ReadSurfaces: []questionrun.SlackSurface{
					{ChannelID: "CINGRESS", ThreadTS: "171000001.000100", Source: "ingress_thread"},
					{ChannelID: "C0AKH5SNGKH", Source: "channel_mention"},
				},
			},
		},
	}, queue.WorkflowQueue)

	if task.TimeoutSeconds != 270 {
		t.Fatalf("expected gather timeout to use full reasoning window, got %d", task.TimeoutSeconds)
	}
	if containsStringExact(task.AllowedTools, "rsi.workflow_context") || containsStringExact(task.AllowedTools, "rsi.trace_context") {
		t.Fatalf("expected noisy RSI context tools to be removed, got %#v", task.AllowedTools)
	}
	for _, expected := range []string{"knowledge.context", "github.repo_activity", "github.repo_context", "repo.context"} {
		if !containsStringExact(task.AllowedTools, expected) {
			t.Fatalf("expected %s in allowed gather tools, got %#v", expected, task.AllowedTools)
		}
	}
	if !strings.Contains(task.SystemMessage, "Stop once the evidence ledger covers the question") {
		t.Fatalf("expected explicit stop condition in system prompt, got %q", task.SystemMessage)
	}
	if !strings.Contains(task.Prompt, "\"gather_contract\"") {
		t.Fatalf("expected structured gather contract in prompt payload, got %q", task.Prompt)
	}
}

func TestQuestionGatherAllowedToolsSkipsKnowledgeWithoutProjectContext(t *testing.T) {
	allowed := questionGatherAllowedTools(questionrun.InvestigationSpec{
		UserRequest: "Summarize recent repo activity.",
		Repo:        "depin-backend",
	})

	if containsStringExact(allowed, "knowledge.context") {
		t.Fatalf("expected no knowledge context without project or alignment need, got %#v", allowed)
	}
	if containsStringExact(allowed, "rsi.workflow_context") || containsStringExact(allowed, "rsi.trace_context") {
		t.Fatalf("expected no RSI self-inspection tools, got %#v", allowed)
	}
}

func TestBuildQuestionGatherTaskIncludesNotionMCPWhenEnabled(t *testing.T) {
	store := storepkg.NewMemoryStore()
	workflowItem := firstQueuedWorkflowItem(t, store, "slack:")
	ctx, err := loadWorkflowContext(store, workflowItem)
	if err != nil {
		t.Fatalf("loadWorkflowContext() error = %v", err)
	}
	task := buildQuestionGatherTask(config.Config{
		Environment:                  "stage",
		DefaultRepo:                  "rsi-agent-platform",
		DefaultReasoningVerbosity:    "verbose",
		NotionMCPEnabled:             true,
		NotionMCPServerURL:           "https://mcp.notion.com/mcp",
		NotionMCPHeaderEnvVars:       map[string]string{"CF-Access-Client-Secret": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_SECRET"},
		NotionMCPAuthorizationEnvVar: "RSI_NOTION_MCP_AUTHORIZATION",
	}, store, questionRunContext{
		workflowContext: ctx,
		questionRun: storepkg.QuestionRun{
			Role: "prod",
			InvestigationSpec: questionrun.InvestigationSpec{
				UserRequest: "Use the linked Notion plan and summarize depin-backend progress this week in accordance with numo.",
				ReplyTarget: questionrun.ReplyTarget{ChannelID: "CINGRESS", ThreadTS: "171000001.000100"},
				Repo:        "depin-backend",
				ProjectKey:  "numo",
				ReadSurfaces: []questionrun.SlackSurface{
					{ChannelID: "CINGRESS", ThreadTS: "171000001.000100", Source: "ingress_thread"},
				},
			},
		},
	}, queue.WorkflowQueue)

	if len(task.MCPServers) != 1 {
		t.Fatalf("expected only Notion MCP server, got %#v", task.MCPServers)
	}
	if task.MCPServers[0].ServerLabel != "notion" {
		t.Fatalf("expected notion MCP server, got %#v", task.MCPServers[0])
	}
	if !reflect.DeepEqual(task.MCPServers[0].HeaderEnvVars, map[string]string{"CF-Access-Client-Secret": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_SECRET"}) {
		t.Fatalf("unexpected notion header env vars %#v", task.MCPServers[0].HeaderEnvVars)
	}
	if !strings.Contains(task.SystemMessage, "Use Notion MCP search and fetch when the user request, pasted links, or gathered evidence point to Notion workspace content.") {
		t.Fatalf("expected notion-specific gather instruction, got %q", task.SystemMessage)
	}
	if !strings.Contains(task.Prompt, "Use Notion MCP search and fetch for Notion workspace evidence") {
		t.Fatalf("expected notion MCP preference in gather prompt, got %q", task.Prompt)
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
