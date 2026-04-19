package control

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/questionrun"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
	"github.com/piplabs/rsi-agent-platform/internal/workflowplan"
)

var (
	projectKeyPattern = regexp.MustCompile(`(?i)\b([a-z0-9][a-z0-9._-]{1,63})\s+project\b`)
)

type questionRunContext struct {
	workflowContext
	questionRun storepkg.QuestionRun
}

func handleClaimedQuestionRunEffect(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, toolClient *clients.ToolGatewayClient, effect transition.EffectExecution) {
	ctx, queueName, err := loadQuestionRunContextForEffect(store, effect)
	if err != nil {
		_ = failClaimedEffect(store, effect, err.Error())
		return
	}
	if err := processQuestionRunEffect(cfg, store, runnerClients, toolClient, ctx, queueName, effect); err != nil {
		finalizeErr := finalizeQuestionRunFailure(cfg, store, ctx, err, time.Now().UTC())
		if finalizeErr != nil {
			_ = failClaimedEffect(store, effect, finalizeErr.Error())
			return
		}
		if completeErr := completeClaimedEffect(store, effect, fmt.Sprintf("question_run:%s:failed", ctx.questionRun.ID)); completeErr != nil {
			_ = failClaimedEffect(store, effect, completeErr.Error())
		}
		return
	}
	if err := completeClaimedEffect(store, effect, fmt.Sprintf("question_run:%s:%s", ctx.questionRun.ID, effect.EffectKind)); err != nil {
		_ = failClaimedEffect(store, effect, err.Error())
	}
}

func processQuestionRunEffect(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, toolClient *clients.ToolGatewayClient, ctx questionRunContext, queueName queue.QueueName, effect transition.EffectExecution) error {
	refreshed, err := refreshQuestionRunContextState(store, ctx)
	if err != nil {
		return err
	}
	ctx = refreshed
	if isTerminalWorkflowStatus(ctx.workflow.Status) || isTerminalTraceStatus(ctx.trace.Summary.Status) || isTerminalQuestionRunStatus(ctx.questionRun.Status) {
		return nil
	}
	switch effect.EffectKind {
	case transition.EffectCompileInvestigationSpec:
		return processCompileInvestigationSpec(cfg, store, ctx, queueName, time.Now().UTC())
	case transition.EffectGatherEvidence:
		return processGatherEvidence(cfg, store, runnerClients, ctx, queueName, time.Now().UTC())
	case transition.EffectReduceReply:
		return processReduceReply(cfg, store, runnerClients, ctx, queueName, time.Now().UTC())
	default:
		return fmt.Errorf("unsupported question_run effect kind %s", effect.EffectKind)
	}
}

func processCompileInvestigationSpec(cfg config.Config, store storepkg.Store, ctx questionRunContext, queueName queue.QueueName, occurredAt time.Time) error {
	spec := buildInvestigationSpec(cfg, ctx.workflow, ctx.ingestion, ctx.trace)
	_, err := submitQuestionRunCommand(store, ctx.questionRun.ID, transition.CommandInvestigationSpecBuilt, cfg.ServiceName, occurredAt, map[string]any{
		"workflow_id":        ctx.workflow.ID,
		"trace_id":           ctx.trace.Summary.TraceID,
		"conversation_id":    ctx.trace.Summary.ConversationID,
		"case_id":            ctx.trace.Summary.CaseID,
		"ingestion_id":       ctx.ingestion.ID,
		"role":               runnerRoleForQueue(queueName),
		"strategy":           "read_heavy_slack_qna",
		"investigation_spec": spec,
		"resume_queue":       string(queueName),
	})
	return err
}

func processGatherEvidence(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, ctx questionRunContext, queueName queue.QueueName, occurredAt time.Time) error {
	runnerClient := runnerClients[runnerRoleForQueue(queueName)]
	if runnerClient == nil {
		return fmt.Errorf("runner client unavailable for queue %s", queueName)
	}
	task := buildQuestionGatherTask(cfg, store, ctx, queueName)
	resp, err := runnerClient.Execute(task)
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", firstNonEmpty(resp.Message, "question_run gather failed"))
	}

	ledger := mergeRunnerEvidenceLedger(ctx.questionRun.EvidenceLedger, questionRunEvidenceLedgerFromRunnerRaw(resp.Raw))
	if delta, deltaErr := parseQuestionRunStructuredOutput[questionrun.EvidenceDelta](resp); deltaErr == nil {
		ledger = mergeEvidenceDelta(ledger, delta)
	}
	if runnerCalls := questionRunToolCallsFromRunnerRaw(resp.Raw); len(runnerCalls) > 0 {
		ledger.ToolCalls = dedupeQuestionRunToolCalls(append(ledger.ToolCalls, runnerCalls...))
	}
	if strings.TrimSpace(ledger.TerminationReason) == "" {
		ledger.TerminationReason = firstNonEmpty(strings.TrimSpace(stringValue(resp.Raw["termination_reason"])), "normal_completion")
	}
	if strings.TrimSpace(ledger.UserRequest) == "" {
		ledger.UserRequest = strings.TrimSpace(ctx.questionRun.InvestigationSpec.UserRequest)
	}
	if ledger.InvestigationSpec == nil {
		spec := ctx.questionRun.InvestigationSpec
		ledger.InvestigationSpec = &spec
	}
	if len(ledger.EvidenceItems) == 0 {
		ledger.MissingEvidence = uniqueStrings(append(ledger.MissingEvidence, "No grounded evidence was captured during the gather phase."))
	}
	runnerDiagnostics := mergeWorkflowRunnerDiagnostics(cloneAnyMap(ctx.questionRun.RunnerDiagnostics), resp.Raw)
	commandKind := transition.CommandEvidenceGathered
	if firstNonEmpty(strings.TrimSpace(stringValue(resp.Raw["completion_verdict"])), "complete") == "partial" || isQuestionRunBoundedStop(ledger.TerminationReason) {
		commandKind = transition.CommandEvidenceGatheredPartial
	}
	_, err = submitQuestionRunCommand(store, ctx.questionRun.ID, commandKind, cfg.ServiceName, occurredAt, map[string]any{
		"workflow_id":        ctx.workflow.ID,
		"trace_id":           ctx.trace.Summary.TraceID,
		"conversation_id":    ctx.trace.Summary.ConversationID,
		"case_id":            ctx.trace.Summary.CaseID,
		"ingestion_id":       ctx.ingestion.ID,
		"evidence_ledger":    ledger,
		"runner_diagnostics": runnerDiagnostics,
	})
	return err
}

func processRefreshAlignmentLedger(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, ctx questionRunContext, occurredAt time.Time) error {
	spec := ctx.questionRun.InvestigationSpec
	ledger := ctx.questionRun.EvidenceLedger
	projectKey := strings.TrimSpace(spec.ProjectKey)
	channelIDs := uniqueNonEmptyChannelIDs(spec.ReadSurfaces)
	toolCalls := append([]questionrun.ToolCall(nil), ledger.ToolCalls...)
	evidenceItems := append([]questionrun.EvidenceItem(nil), ledger.EvidenceItems...)
	sources := []string{}
	knowledgeResult, knowledgeItems, knowledgeCall := executeQuestionRunTool(toolClient, "knowledge.context", map[string]any{
		"topic":    projectKey,
		"scope_id": spec.Repo,
	})
	toolCalls = append(toolCalls, knowledgeCall)
	evidenceItems = append(evidenceItems, knowledgeItems...)
	if sourceRefs := evidenceSourceRefs(knowledgeItems); len(sourceRefs) > 0 {
		sources = append(sources, sourceRefs...)
	}
	if projectKey != "" && len(channelIDs) > 0 {
		searchResult, searchItems, searchCall := executeQuestionRunTool(toolClient, "slack.search", map[string]any{
			"query":       projectKey,
			"channel_ids": channelIDs,
			"trace_id":    ctx.trace.Summary.TraceID,
			"limit":       8,
		})
		toolCalls = append(toolCalls, searchCall)
		evidenceItems = append(evidenceItems, searchItems...)
		if sourceRefs := evidenceSourceRefs(searchItems); len(sourceRefs) > 0 {
			sources = append(sources, sourceRefs...)
		}
		_ = searchResult
	}
	evidenceItems = dedupeEvidenceItems(evidenceItems)
	toolCalls = dedupeQuestionRunToolCalls(toolCalls)
	alignmentLedger := questionrun.ProjectAlignmentLedger{
		ProjectKey:       projectKey,
		RequiredOutcomes: topEvidenceSummaries(evidenceItems, 4),
		Constraints:      []string{},
		OpenQuestions:    []string{},
		Sources:          uniqueStrings(sources),
		EvidenceItems:    evidenceItems[:minInt(len(evidenceItems), 8)],
	}
	if len(alignmentLedger.EvidenceItems) == 0 {
		alignmentLedger.Degraded = true
		alignmentLedger.DegradedReason = firstNonEmpty(
			toolResultSummary(knowledgeResult, nil),
			"no grounded alignment evidence was available",
		)
		alignmentLedger.OpenQuestions = []string{
			fmt.Sprintf("Need a canonical alignment source for %s.", firstNonEmpty(projectKey, "the referenced project")),
		}
		alignmentLedger.Summary = fmt.Sprintf("Alignment ledger for %s is degraded: %s.", firstNonEmpty(projectKey, "the referenced project"), alignmentLedger.DegradedReason)
	} else {
		alignmentLedger.Summary = fmt.Sprintf("Alignment ledger refreshed for %s from %d grounded evidence item(s).", firstNonEmpty(projectKey, "the referenced project"), len(alignmentLedger.EvidenceItems))
		if knowledgeID, err := persistAlignmentLedger(store, ctx, alignmentLedger, occurredAt); err == nil {
			alignmentLedger.KnowledgeEntryID = knowledgeID
		}
	}
	ledger.ToolCalls = toolCalls
	ledger.EvidenceItems = dedupeEvidenceItems(evidenceItems)
	ledger.AlignmentLedger = &alignmentLedger
	ledger.AlignmentRequired = true
	ledger.AlignmentDegraded = alignmentLedger.Degraded
	commandKind := transition.CommandAlignmentLedgerReady
	if alignmentLedger.Degraded {
		commandKind = transition.CommandAlignmentLedgerDegraded
	}
	_, err := submitQuestionRunCommand(store, ctx.questionRun.ID, commandKind, cfg.ServiceName, occurredAt, map[string]any{
		"workflow_id":      ctx.workflow.ID,
		"trace_id":         ctx.trace.Summary.TraceID,
		"conversation_id":  ctx.trace.Summary.ConversationID,
		"case_id":          ctx.trace.Summary.CaseID,
		"ingestion_id":     ctx.ingestion.ID,
		"alignment_ledger": alignmentLedger,
		"evidence_ledger":  ledger,
	})
	return err
}

func processCollectSeedEvidence(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, ctx questionRunContext, occurredAt time.Time) error {
	spec := ctx.questionRun.InvestigationSpec
	ledger := ctx.questionRun.EvidenceLedger
	toolCalls := append([]questionrun.ToolCall(nil), ledger.ToolCalls...)
	evidenceItems := append([]questionrun.EvidenceItem(nil), ledger.EvidenceItems...)
	if strings.TrimSpace(spec.Repo) != "" {
		_, items, call := executeQuestionRunTool(toolClient, "repo.context", map[string]any{
			"repo":     spec.Repo,
			"question": spec.UserRequest,
		})
		toolCalls = append(toolCalls, call)
		evidenceItems = append(evidenceItems, items...)
		_, activityItems, activityCall := executeQuestionRunTool(toolClient, "github.repo_activity", map[string]any{
			"repo":  spec.Repo,
			"since": spec.Since,
			"until": spec.Until,
		})
		toolCalls = append(toolCalls, activityCall)
		evidenceItems = append(evidenceItems, activityItems...)
	}
	searchQuery := slackSearchQueryForSpec(spec)
	for _, surface := range spec.ReadSurfaces {
		if surface.ChannelID == "" {
			continue
		}
		if surface.ThreadTS != "" {
			_, items, call := executeQuestionRunTool(toolClient, "slack.history", map[string]any{
				"channel_id": surface.ChannelID,
				"thread_ts":  surface.ThreadTS,
				"trace_id":   ctx.trace.Summary.TraceID,
				"limit":      12,
			})
			toolCalls = append(toolCalls, call)
			evidenceItems = append(evidenceItems, items...)
			continue
		}
		_, items, call := executeQuestionRunTool(toolClient, "slack.search", map[string]any{
			"query":       searchQuery,
			"channel_ids": []string{surface.ChannelID},
			"trace_id":    ctx.trace.Summary.TraceID,
			"limit":       8,
		})
		toolCalls = append(toolCalls, call)
		evidenceItems = append(evidenceItems, items...)
	}
	ledger.ToolCalls = dedupeQuestionRunToolCalls(toolCalls)
	ledger.EvidenceItems = dedupeEvidenceItems(evidenceItems)
	ledger.OpenQuestions = deriveOpenQuestions(spec, ledger)
	ledger.MissingEvidence = append([]string(nil), ledger.OpenQuestions...)
	shouldExpand := spec.AllowExpansion && len(ledger.OpenQuestions) > 0
	_, err := submitQuestionRunCommand(store, ctx.questionRun.ID, transition.CommandSeedEvidenceCollected, cfg.ServiceName, occurredAt, map[string]any{
		"workflow_id":     ctx.workflow.ID,
		"trace_id":        ctx.trace.Summary.TraceID,
		"conversation_id": ctx.trace.Summary.ConversationID,
		"case_id":         ctx.trace.Summary.CaseID,
		"ingestion_id":    ctx.ingestion.ID,
		"evidence_ledger": ledger,
		"should_expand":   shouldExpand,
	})
	return err
}

func processExpandEvidence(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, ctx questionRunContext, queueName queue.QueueName, occurredAt time.Time) error {
	runnerClient := runnerClients[runnerRoleForQueue(queueName)]
	if runnerClient == nil {
		return fmt.Errorf("runner client unavailable for queue %s", queueName)
	}
	task := buildQuestionExpandTask(cfg, store, ctx)
	resp, err := runnerClient.Execute(task)
	if err != nil {
		return err
	}
	if !resp.OK {
		terminationReason := strings.TrimSpace(stringValue(resp.Raw["termination_reason"]))
		if isQuestionRunBoundedStop(terminationReason) {
			ledger := mergeRunnerEvidenceLedger(ctx.questionRun.EvidenceLedger, questionRunEvidenceLedgerFromRunnerRaw(resp.Raw))
			ledger.MissingEvidence = uniqueStrings(append(ledger.MissingEvidence, fmt.Sprintf("Evidence expansion stopped early due to %s.", terminationReason)))
			ledger.OpenQuestions = uniqueStrings(append(ledger.OpenQuestions, fmt.Sprintf("Need best-effort reduction from the bounded evidence ledger because expansion stopped at %s.", terminationReason)))
			runnerDiagnostics := mergeWorkflowRunnerDiagnostics(cloneAnyMap(ctx.questionRun.RunnerDiagnostics), resp.Raw)
			runnerDiagnostics["completion_verdict"] = "partial"
			_, err = submitQuestionRunCommand(store, ctx.questionRun.ID, transition.CommandEvidenceExpanded, cfg.ServiceName, occurredAt, map[string]any{
				"workflow_id":        ctx.workflow.ID,
				"trace_id":           ctx.trace.Summary.TraceID,
				"conversation_id":    ctx.trace.Summary.ConversationID,
				"case_id":            ctx.trace.Summary.CaseID,
				"ingestion_id":       ctx.ingestion.ID,
				"evidence_ledger":    ledger,
				"runner_diagnostics": runnerDiagnostics,
			})
			return err
		}
		return fmt.Errorf("%s", firstNonEmpty(resp.Message, "question_run expansion failed"))
	}
	delta, err := parseQuestionRunStructuredOutput[questionrun.EvidenceDelta](resp)
	if err != nil {
		return err
	}
	ledger := mergeEvidenceDelta(ctx.questionRun.EvidenceLedger, delta)
	if runnerCalls := questionRunToolCallsFromRunnerRaw(resp.Raw); len(runnerCalls) > 0 {
		ledger.ToolCalls = dedupeQuestionRunToolCalls(append(ledger.ToolCalls, runnerCalls...))
	}
	runnerDiagnostics := mergeWorkflowRunnerDiagnostics(cloneAnyMap(ctx.questionRun.RunnerDiagnostics), resp.Raw)
	_, err = submitQuestionRunCommand(store, ctx.questionRun.ID, transition.CommandEvidenceExpanded, cfg.ServiceName, occurredAt, map[string]any{
		"workflow_id":        ctx.workflow.ID,
		"trace_id":           ctx.trace.Summary.TraceID,
		"conversation_id":    ctx.trace.Summary.ConversationID,
		"case_id":            ctx.trace.Summary.CaseID,
		"ingestion_id":       ctx.ingestion.ID,
		"evidence_ledger":    ledger,
		"runner_diagnostics": runnerDiagnostics,
	})
	return err
}

func processReduceReply(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, ctx questionRunContext, queueName queue.QueueName, occurredAt time.Time) error {
	runnerClient := runnerClients[runnerRoleForQueue(queueName)]
	if runnerClient == nil {
		return fmt.Errorf("runner client unavailable for queue %s", queueName)
	}
	task := buildQuestionReduceTask(cfg, store, ctx, queueName)
	resp, err := runnerClient.Execute(task)
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", firstNonEmpty(resp.Message, "question_run reducer failed"))
	}
	result, err := parseQuestionRunStructuredOutput[questionrun.Result](resp)
	if err != nil {
		return err
	}
	if strings.TrimSpace(result.CompletionVerdict) == "" {
		result.CompletionVerdict = firstNonEmpty(strings.TrimSpace(stringValue(resp.Raw["completion_verdict"])), "complete")
	}
	if strings.TrimSpace(result.TerminationReason) == "" {
		result.TerminationReason = firstNonEmpty(strings.TrimSpace(stringValue(resp.Raw["termination_reason"])), "normal_completion")
	}
	if result.CompletionVerdict == "partial" {
		result.ReplyMarkdown = standardizePartialReplyBody(result.ReplyMarkdown, result.TerminationReason)
	}
	if strings.TrimSpace(result.ReplyMarkdown) == "" {
		_, err := submitQuestionRunCommand(store, ctx.questionRun.ID, transition.CommandReplyBlocked, cfg.ServiceName, occurredAt, map[string]any{
			"workflow_id":        ctx.workflow.ID,
			"trace_id":           ctx.trace.Summary.TraceID,
			"conversation_id":    ctx.trace.Summary.ConversationID,
			"case_id":            ctx.trace.Summary.CaseID,
			"ingestion_id":       ctx.ingestion.ID,
			"result":             result,
			"last_error":         "question_run reducer returned no reply_markdown",
			"failure_class":      "reply_missing",
			"failure_summary":    "Question-run reducer returned successfully but did not produce a reply body.",
			"runner_diagnostics": mergeWorkflowRunnerDiagnostics(cloneAnyMap(ctx.questionRun.RunnerDiagnostics), resp.Raw),
			"reasoning_steps":    []events.ReasoningStep{},
			"tool_calls":         traceToolCallsFromQuestionRunLedger(ctx, ctx.questionRun.EvidenceLedger.ToolCalls, occurredAt),
		})
		return err
	}
	replyActionID := ""
	reasoningSteps := questionRunReasoningSteps(ctx, result, occurredAt)
	toolCalls := traceToolCallsFromQuestionRunLedger(ctx, dedupeQuestionRunToolCalls(append(ctx.questionRun.EvidenceLedger.ToolCalls, questionRunToolCallsFromRunnerRaw(resp.Raw)...)), occurredAt)
	allowed, policyVerdict := replyPolicy(store, ctx.workflow.Kind, ctx.trace.Summary.ThreadKey, ctx.ingestion.ChannelID)
	if !allowed {
		_, err := submitQuestionRunCommand(store, ctx.questionRun.ID, transition.CommandReplyBlocked, cfg.ServiceName, occurredAt, map[string]any{
			"workflow_id":        ctx.workflow.ID,
			"trace_id":           ctx.trace.Summary.TraceID,
			"conversation_id":    ctx.trace.Summary.ConversationID,
			"case_id":            ctx.trace.Summary.CaseID,
			"ingestion_id":       ctx.ingestion.ID,
			"result":             result,
			"last_error":         fmt.Sprintf("Slack reply blocked by policy: %s", policyVerdict),
			"failure_class":      "reply_policy_blocked",
			"failure_summary":    "Reply reducer produced a response, but Slack posting is blocked by policy.",
			"runner_diagnostics": mergeWorkflowRunnerDiagnostics(cloneAnyMap(ctx.questionRun.RunnerDiagnostics), resp.Raw),
			"reasoning_steps":    reasoningSteps,
			"tool_calls":         toolCalls,
		})
		return err
	}
	output := questionRunStructuredOutputForReply(ctx, result)
	intent, _, _, err := draftSlackPostAction(cfg, store, queueName, ctx.workflowContext, output, result.ReplyMarkdown, true, policyVerdict, result.CompletionVerdict, occurredAt)
	if err != nil {
		return err
	}
	replyActionID = intent.ID
	commandKind := transition.CommandReplyReduced
	if result.CompletionVerdict == "partial" {
		commandKind = transition.CommandReplyReducedPartial
	}
	_, err = submitQuestionRunCommand(store, ctx.questionRun.ID, commandKind, cfg.ServiceName, occurredAt, map[string]any{
		"workflow_id":        ctx.workflow.ID,
		"trace_id":           ctx.trace.Summary.TraceID,
		"conversation_id":    ctx.trace.Summary.ConversationID,
		"case_id":            ctx.trace.Summary.CaseID,
		"ingestion_id":       ctx.ingestion.ID,
		"result":             result,
		"reply_action_id":    replyActionID,
		"runner_diagnostics": mergeWorkflowRunnerDiagnostics(cloneAnyMap(ctx.questionRun.RunnerDiagnostics), resp.Raw),
		"reasoning_steps":    reasoningSteps,
		"tool_calls":         toolCalls,
	})
	return err
}

func loadQuestionRunContextForEffect(store storepkg.Store, effect transition.EffectExecution) (questionRunContext, queue.QueueName, error) {
	questionRunID := strings.TrimSpace(effect.AggregateID)
	if questionRunID == "" {
		return questionRunContext{}, "", fmt.Errorf("question_run effect %s missing aggregate id", effect.ID)
	}
	item, ok := store.GetQuestionRun(questionRunID)
	if !ok {
		return questionRunContext{}, "", fmt.Errorf("question_run %s not found", questionRunID)
	}
	workflowCtx, err := loadWorkflowContext(store, workflowLocator{
		traceID:     item.TraceID,
		workflowID:  item.WorkflowID,
		ingestionID: item.IngestionID,
	})
	if err != nil {
		return questionRunContext{}, "", err
	}
	queueName := queueNameFromString(firstNonEmpty(stringFromMap(effect.Payload, "resume_queue"), string(roleQueueName(item.Role))))
	return questionRunContext{workflowContext: workflowCtx, questionRun: item}, queueName, nil
}

func refreshQuestionRunContextState(store storepkg.Store, ctx questionRunContext) (questionRunContext, error) {
	workflowCtx, err := refreshWorkflowContextState(store, ctx.workflowContext)
	if err != nil {
		return questionRunContext{}, err
	}
	item, ok := store.GetQuestionRun(ctx.questionRun.ID)
	if !ok {
		return questionRunContext{}, fmt.Errorf("question_run %s not found", ctx.questionRun.ID)
	}
	return questionRunContext{workflowContext: workflowCtx, questionRun: item}, nil
}

func finalizeQuestionRunFailure(cfg config.Config, store storepkg.Store, ctx questionRunContext, procErr error, occurredAt time.Time) error {
	payload := map[string]any{
		"workflow_id":     ctx.workflow.ID,
		"trace_id":        ctx.trace.Summary.TraceID,
		"conversation_id": ctx.trace.Summary.ConversationID,
		"case_id":         ctx.trace.Summary.CaseID,
		"ingestion_id":    ctx.ingestion.ID,
		"last_error":      strings.TrimSpace(procErr.Error()),
		"failure_class":   "question_run_effect_failed",
		"failure_summary": "Question-run execution failed before producing a reducer result.",
	}
	_, err := submitQuestionRunCommand(store, ctx.questionRun.ID, transition.CommandQuestionRunFailed, cfg.ServiceName, occurredAt, payload)
	return err
}

func submitQuestionRunCommand(store storepkg.Store, questionRunID string, kind transition.QuestionRunCommandKind, actor string, occurredAt time.Time, payload map[string]any) (transition.CommandReceipt, error) {
	if strings.TrimSpace(questionRunID) == "" {
		return transition.CommandReceipt{}, nil
	}
	return store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineQuestionRun,
		AggregateID: questionRunID,
		CommandKind: string(kind),
		CommandID:   questionRunCommandID(questionRunID, kind),
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
}

func questionRunCommandID(questionRunID string, kind transition.QuestionRunCommandKind) string {
	return fmt.Sprintf("cmd-question-run:%s:%s", strings.TrimSpace(questionRunID), string(kind))
}

func buildInvestigationSpec(cfg config.Config, workflow storepkg.Workflow, ingestion slackpkg.Ingestion, trace events.Trace) questionrun.InvestigationSpec {
	now := time.Now().UTC()
	userRequest := firstNonEmpty(strings.TrimSpace(ingestion.Prompt.RenderedText), strings.TrimSpace(ingestion.Text))
	hints := workflowplan.BuildLiveHints(workflowplan.RuntimeConfig{
		DefaultRepo:      cfg.DefaultRepo,
		AllowedRepos:     append([]string(nil), cfg.AllowedTargetRepos...),
		KnowledgeBaseURL: cfg.DefaultKnowledgeBaseURL,
		SandboxNamespace: cfg.SandboxNamespace,
	}, workflowplan.RequestContext{
		Trace:          trace.Summary,
		WorkflowID:     workflow.ID,
		ConversationID: workflow.ConversationID,
		CaseID:         workflow.CaseID,
		WorkflowKind:   workflow.Kind,
		AssignedBot:    workflow.AssignedBot,
		Question:       userRequest,
		ChannelID:      ingestion.ChannelID,
		ThreadTS:       ingestion.ThreadTS,
		EntityRefs:     append([]slackpkg.EntityRef(nil), ingestion.EntityRefs...),
	}, now)
	surfaces := make([]questionrun.SlackSurface, 0, len(hints.CandidateReadSurfaces))
	for _, item := range hints.CandidateReadSurfaces {
		surfaces = append(surfaces, questionrun.SlackSurface{
			ChannelID: item.ChannelID,
			ThreadTS:  item.ThreadTS,
			Ref:       item.Ref,
			Source:    item.Source,
		})
	}
	return questionrun.InvestigationSpec{
		UserRequest:       userRequest,
		ReplyTarget:       questionrun.ReplyTarget{ChannelID: ingestion.ChannelID, ThreadTS: ingestion.ThreadTS},
		Prompt:            ingestion.Prompt,
		Repo:              hints.Repo,
		ProjectKey:        projectKeyFromQuestion(userRequest),
		Since:             hints.Since,
		Until:             hints.Until,
		ReadSurfaces:      surfaces,
		AlignmentRequired: questionRequiresAlignment(userRequest),
		RetrievalBudget:   20,
		WorkflowStrategy:  "read_heavy_slack_qna",
		GatherTaskType:    "question_gather",
		ReduceTaskType:    "question_reduce",
		ReductionTaskType: "question_reduce",
		ExpansionTaskType: "question_gather",
	}
}

func questionRequiresAlignment(question string) bool {
	lower := strings.ToLower(strings.TrimSpace(question))
	for _, token := range []string{"alignment", "aligned", "misaligned", "misalignment", "accordance with", "according to"} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func projectKeyFromQuestion(question string) string {
	matches := projectKeyPattern.FindStringSubmatch(strings.TrimSpace(question))
	if len(matches) < 2 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(matches[1]))
}

func uniqueNonEmptyChannelIDs(items []questionrun.SlackSurface) []string {
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		channelID := strings.TrimSpace(item.ChannelID)
		if channelID == "" {
			continue
		}
		if _, ok := seen[channelID]; ok {
			continue
		}
		seen[channelID] = struct{}{}
		out = append(out, channelID)
	}
	return out
}

func persistAlignmentLedger(store storepkg.Store, ctx questionRunContext, ledger questionrun.ProjectAlignmentLedger, occurredAt time.Time) (string, error) {
	entry, err := runnerutil.PersistKnowledgeDraft(store, knowledge.Entry{
		Tier:       knowledge.TierWorking,
		Kind:       knowledge.KindRepoNote,
		ScopeType:  knowledge.ScopeRepo,
		ScopeID:    ctx.questionRun.InvestigationSpec.Repo,
		Title:      fmt.Sprintf("Project alignment ledger: %s", firstNonEmpty(ledger.ProjectKey, "unnamed")),
		Summary:    ledger.Summary,
		Body:       strings.Join(topEvidenceSummaries(ledger.EvidenceItems, 8), "\n"),
		Status:     knowledge.StatusDraft,
		Confidence: 0.75,
		SourceType: knowledge.SourceAgent,
		CreatedAt:  occurredAt,
		UpdatedAt:  occurredAt,
	}, evidenceLinksFromQuestionRun(ledger.EvidenceItems), "control-plane", ctx.trace.Summary.TraceID, 0, occurredAt)
	if err != nil {
		return "", err
	}
	_, err = store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineKnowledge,
		AggregateID: entry.ID,
		CommandKind: string(transition.CommandKnowledgeApprove),
		CommandID:   fmt.Sprintf("cmd-knowledge:%s:approve", entry.ID),
		Actor:       "control-plane",
		OccurredAt:  occurredAt,
		Payload:     map[string]any{},
	})
	if err != nil {
		return "", err
	}
	return entry.ID, nil
}

func evidenceLinksFromQuestionRun(items []questionrun.EvidenceItem) []knowledge.EvidenceLink {
	out := make([]knowledge.EvidenceLink, 0, len(items))
	for idx, item := range items {
		ref := firstNonEmpty(item.SourceRef, item.Permalink, item.Path, fmt.Sprintf("evidence-%d", idx+1))
		out = append(out, knowledge.EvidenceLink{
			EvidenceType:     item.Kind,
			EvidenceID:       ref,
			RelevanceSummary: item.Summary,
			EvidenceRef: events.EvidenceRef{
				Kind:    item.Kind,
				Ref:     ref,
				Summary: item.Summary,
			},
		})
	}
	return out
}

func executeQuestionRunTool(toolClient *clients.ToolGatewayClient, name string, input map[string]any) (storepkg.ToolResult, []questionrun.EvidenceItem, questionrun.ToolCall) {
	result, err := toolClient.Execute(name, input)
	summary := toolResultSummary(result, err)
	call := questionrun.ToolCall{
		ToolName:        name,
		ToolCallID:      firstNonEmpty(result.ToolCallID, name),
		Request:         cloneAnyMap(input),
		Summary:         summary,
		Status:          firstNonEmpty(result.Status, "failed"),
		ProviderRef:     result.ProviderRef,
		RawArtifactRefs: append([]string(nil), result.RawArtifactRefs...),
	}
	if err != nil {
		call.Status = "failed"
		call.Summary = err.Error()
		return storepkg.ToolResult{}, nil, call
	}
	return result, extractQuestionRunEvidenceItems(name, input, result.Output, result.Summary), call
}

func extractQuestionRunEvidenceItems(toolName string, request map[string]any, output map[string]any, summary string) []questionrun.EvidenceItem {
	switch toolName {
	case "slack.history":
		return slackHistoryEvidenceItems(request, output)
	case "slack.search":
		return slackSearchEvidenceItems(output)
	case "repo.context":
		return repoContextEvidenceItems(request, output)
	case "github.repo_activity":
		return githubRepoActivityEvidenceItems(request, output)
	case "knowledge.context":
		return knowledgeContextEvidenceItems(request, output)
	default:
		if strings.TrimSpace(summary) == "" {
			return nil
		}
		return []questionrun.EvidenceItem{{
			Kind:      "tool_summary",
			Summary:   strings.TrimSpace(summary),
			SourceRef: firstNonEmpty(stringFromMap(request, "path"), stringFromMap(request, "channel_id")),
			ToolName:  toolName,
		}}
	}
}

func slackHistoryEvidenceItems(request map[string]any, output map[string]any) []questionrun.EvidenceItem {
	channelID := firstNonEmpty(stringFromMap(output, "channel_id"), stringFromMap(request, "channel_id"))
	threadTS := firstNonEmpty(stringFromMap(output, "thread_ts"), stringFromMap(request, "thread_ts"))
	raw, _ := output["messages"].([]any)
	items := make([]questionrun.EvidenceItem, 0, minInt(len(raw), 6))
	for _, entry := range raw {
		message, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		text := truncate(firstNonEmpty(stringFromMap(message, "text"), stringFromMap(message, "content"), stringFromMap(message, "body")), 500)
		if strings.TrimSpace(text) == "" {
			continue
		}
		ts := firstNonEmpty(stringFromMap(message, "ts"), stringFromMap(message, "message_timestamp"))
		items = append(items, questionrun.EvidenceItem{
			Kind:      "slack_message",
			Summary:   text,
			Snippet:   text,
			SourceRef: firstNonEmpty(stringFromMap(message, "permalink"), slackMessageRef(channelID, ts)),
			ToolName:  "slack.history",
			ChannelID: channelID,
			ThreadTS:  firstNonEmpty(stringFromMap(message, "thread_ts"), threadTS, ts),
			MessageTS: ts,
			Permalink: stringFromMap(message, "permalink"),
			Author:    firstNonEmpty(stringFromMap(message, "author_name"), stringFromMap(message, "user"), stringFromMap(message, "user_id")),
		})
	}
	return items
}

func slackSearchEvidenceItems(output map[string]any) []questionrun.EvidenceItem {
	raw, _ := output["messages"].([]any)
	items := make([]questionrun.EvidenceItem, 0, minInt(len(raw), 6))
	for _, entry := range raw {
		message, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		text := truncate(firstNonEmpty(stringFromMap(message, "text"), stringFromMap(message, "content")), 500)
		if strings.TrimSpace(text) == "" {
			continue
		}
		channelID := stringFromMap(message, "channel_id")
		ts := firstNonEmpty(stringFromMap(message, "ts"), stringFromMap(message, "message_timestamp"))
		items = append(items, questionrun.EvidenceItem{
			Kind:      "slack_search_match",
			Summary:   text,
			Snippet:   text,
			SourceRef: firstNonEmpty(stringFromMap(message, "permalink"), slackMessageRef(channelID, ts)),
			ToolName:  "slack.search",
			ChannelID: channelID,
			ThreadTS:  firstNonEmpty(stringFromMap(message, "thread_ts"), ts),
			MessageTS: ts,
			Permalink: stringFromMap(message, "permalink"),
			Author:    firstNonEmpty(stringFromMap(message, "author_name"), stringFromMap(message, "author_user_id")),
		})
	}
	return items
}

func repoContextEvidenceItems(request map[string]any, output map[string]any) []questionrun.EvidenceItem {
	repo := firstNonEmpty(stringFromMap(output, "repo"), stringFromMap(request, "repo"))
	items := []questionrun.EvidenceItem{}
	if description := strings.TrimSpace(stringFromMap(output, "description")); description != "" {
		items = append(items, questionrun.EvidenceItem{
			Kind:      "repo_context",
			Summary:   fmt.Sprintf("Repository context for %s: %s", repo, truncate(description, 400)),
			Snippet:   truncate(description, 400),
			SourceRef: firstNonEmpty(stringFromMap(output, "html_url"), repo),
			ToolName:  "repo.context",
			Repo:      repo,
		})
	}
	raw, _ := output["matches"].([]any)
	for _, entry := range raw {
		match, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		path := stringFromMap(match, "path")
		snippet := truncate(stringFromMap(match, "snippet"), 500)
		items = append(items, questionrun.EvidenceItem{
			Kind:      "repo_context_match",
			Summary:   firstNonEmpty(snippet, path),
			Snippet:   snippet,
			SourceRef: firstNonEmpty(stringFromMap(match, "html_url"), path),
			ToolName:  "repo.context",
			Path:      path,
			Repo:      repo,
		})
	}
	return items
}

func githubRepoActivityEvidenceItems(request map[string]any, output map[string]any) []questionrun.EvidenceItem {
	repo := firstNonEmpty(stringFromMap(output, "repo"), stringFromMap(request, "repo"))
	items := []questionrun.EvidenceItem{}
	for _, key := range []string{"commits", "merged_pull_requests", "opened_pull_requests"} {
		raw, _ := output[key].([]any)
		limit := 2
		if key == "opened_pull_requests" {
			limit = 1
		}
		for idx, entry := range raw {
			if idx >= limit {
				break
			}
			record, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			title := truncate(firstNonEmpty(stringFromMap(record, "message"), stringFromMap(record, "title"), stringFromMap(record, "sha")), 320)
			if strings.TrimSpace(title) == "" {
				continue
			}
			kind := "github_commit"
			summary := title
			if key != "commits" {
				kind = "github_pull_request"
				prefix := "Merged PR"
				if key == "opened_pull_requests" {
					prefix = "Opened PR"
				}
				summary = prefix + ": " + title
			}
			items = append(items, questionrun.EvidenceItem{
				Kind:      kind,
				Summary:   summary,
				SourceRef: firstNonEmpty(stringFromMap(record, "url"), repo),
				ToolName:  "github.repo_activity",
				Repo:      repo,
				Commit:    stringFromMap(record, "sha"),
				Author:    stringFromMap(record, "author"),
			})
		}
	}
	if len(items) > 0 {
		return items
	}
	if summary := strings.TrimSpace(stringFromMap(output, "summary")); summary != "" {
		return []questionrun.EvidenceItem{{
			Kind:      "github_activity_summary",
			Summary:   summary,
			SourceRef: repo,
			ToolName:  "github.repo_activity",
			Repo:      repo,
		}}
	}
	return nil
}

func knowledgeContextEvidenceItems(request map[string]any, output map[string]any) []questionrun.EvidenceItem {
	raw, _ := output["entries"].([]any)
	items := make([]questionrun.EvidenceItem, 0, minInt(len(raw), 6))
	for _, entry := range raw {
		record, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		summary := firstNonEmpty(stringFromMap(record, "summary"), stringFromMap(record, "body"), stringFromMap(record, "title"))
		if strings.TrimSpace(summary) == "" {
			continue
		}
		items = append(items, questionrun.EvidenceItem{
			Kind:      "knowledge_entry",
			Summary:   truncate(summary, 500),
			Snippet:   truncate(summary, 500),
			SourceRef: firstNonEmpty(stringFromMap(record, "id"), stringFromMap(request, "scope_id")),
			ToolName:  "knowledge.context",
		})
	}
	return items
}

func dedupeEvidenceItems(items []questionrun.EvidenceItem) []questionrun.EvidenceItem {
	out := make([]questionrun.EvidenceItem, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := strings.Join([]string{item.Kind, item.SourceRef, item.Summary, item.ToolName}, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func dedupeQuestionRunToolCalls(items []questionrun.ToolCall) []questionrun.ToolCall {
	out := make([]questionrun.ToolCall, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := firstNonEmpty(item.ToolCallID, item.ToolName+"|"+item.Summary)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func evidenceSourceRefs(items []questionrun.EvidenceItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if ref := strings.TrimSpace(firstNonEmpty(item.SourceRef, item.Permalink, item.Path)); ref != "" {
			out = append(out, ref)
		}
	}
	return uniqueStrings(out)
}

func topEvidenceSummaries(items []questionrun.EvidenceItem, limit int) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, minInt(len(items), limit))
	for _, item := range items {
		if strings.TrimSpace(item.Summary) == "" {
			continue
		}
		out = append(out, item.Summary)
		if len(out) == limit {
			break
		}
	}
	return out
}

func slackSearchQueryForSpec(spec questionrun.InvestigationSpec) string {
	parts := []string{}
	if strings.TrimSpace(spec.Repo) != "" {
		parts = append(parts, strings.TrimSpace(spec.Repo))
	}
	if strings.TrimSpace(spec.ProjectKey) != "" {
		parts = append(parts, strings.TrimSpace(spec.ProjectKey))
	}
	if len(parts) == 0 {
		return strings.TrimSpace(spec.UserRequest)
	}
	return strings.Join(parts, " ")
}

func deriveOpenQuestions(spec questionrun.InvestigationSpec, ledger questionrun.EvidenceLedger) []string {
	out := []string{}
	if len(ledger.EvidenceItems) == 0 {
		out = append(out, "No grounded evidence was collected yet.")
	}
	if spec.AlignmentRequired && (ledger.AlignmentLedger == nil || ledger.AlignmentDegraded) {
		out = append(out, fmt.Sprintf("Need a fresher alignment ledger for %s.", firstNonEmpty(spec.ProjectKey, "the referenced project")))
	}
	if !hasEvidenceKind(ledger.EvidenceItems, "slack_message", "slack_search_match") {
		out = append(out, "Need better Slack discussion evidence from the referenced channels or thread.")
	}
	if strings.TrimSpace(spec.Repo) != "" && !hasRepoEvidence(ledger.EvidenceItems) {
		out = append(out, fmt.Sprintf("Need repository activity evidence for %s before reducing the reply.", spec.Repo))
	}
	for _, surface := range referencedReadSurfaces(spec.ReadSurfaces) {
		if !hasSlackEvidenceForSurface(ledger.EvidenceItems, surface) {
			out = append(out, fmt.Sprintf("Need Slack discussion evidence from referenced channel %s before reducing the reply.", surface.ChannelID))
		}
	}
	return uniqueStrings(out)
}

func hasEvidenceKind(items []questionrun.EvidenceItem, kinds ...string) bool {
	allowed := map[string]struct{}{}
	for _, kind := range kinds {
		allowed[kind] = struct{}{}
	}
	for _, item := range items {
		if _, ok := allowed[item.Kind]; ok {
			return true
		}
	}
	return false
}

func hasRepoEvidence(items []questionrun.EvidenceItem) bool {
	return hasEvidenceKind(items, "github_commit", "github_pull_request", "github_activity_summary", "repo_context_match", "repo_context")
}

func referencedReadSurfaces(items []questionrun.SlackSurface) []questionrun.SlackSurface {
	out := make([]questionrun.SlackSurface, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		if strings.TrimSpace(item.ChannelID) == "" {
			continue
		}
		if item.Source == "ingress_thread" {
			continue
		}
		key := strings.Join([]string{item.ChannelID, item.ThreadTS, item.Source}, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func hasSlackEvidenceForSurface(items []questionrun.EvidenceItem, surface questionrun.SlackSurface) bool {
	for _, item := range items {
		if item.Kind != "slack_message" && item.Kind != "slack_search_match" {
			continue
		}
		if strings.TrimSpace(item.ChannelID) != strings.TrimSpace(surface.ChannelID) {
			continue
		}
		if strings.TrimSpace(surface.ThreadTS) != "" && strings.TrimSpace(item.ThreadTS) != strings.TrimSpace(surface.ThreadTS) {
			continue
		}
		return true
	}
	return false
}

func questionRunFinalizationReserveSeconds(total time.Duration) time.Duration {
	seconds := int(total / time.Second)
	if seconds <= 0 {
		return 30 * time.Second
	}
	reserve := seconds / 10
	if reserve < 10 {
		reserve = 10
	}
	if reserve > 30 {
		reserve = 30
	}
	return time.Duration(reserve) * time.Second
}

func buildQuestionGatherTask(cfg config.Config, store storepkg.Store, ctx questionRunContext, queueName queue.QueueName) clients.RunnerTask {
	sessionScopeKind, sessionScopeID, parentScopeKind, parentScopeID := workflowSessionScope(ctx.trace, ctx.workflow)
	role := runnerRoleForQueue(queueName)
	effectiveHarness := harness.ResolveEffectiveConfig(store, role, cfg.DefaultReasoningVerbosity)
	mcpServers := slackMCPServersForRead(ctx.ingestion.ChannelID, ctx.ingestion.ThreadTS)
	totalTimeout := cfg.RunnerTaskTimeoutForRole(role)
	timeoutSeconds := int(questionRunGatherTimeout(totalTimeout) / time.Second)
	if timeoutSeconds <= 0 {
		timeoutSeconds = int(totalTimeout / time.Second)
	}
	return clients.RunnerTask{
		TaskType:                  "question_gather",
		Repo:                      firstNonEmpty(ctx.questionRun.InvestigationSpec.Repo, cfg.DefaultRepo),
		RepoRef:                   "main",
		Prompt:                    questionGatherPrompt(ctx.questionRun.InvestigationSpec, ctx.questionRun.EvidenceLedger),
		SystemMessage:             questionGatherSystemMessage(ctx.questionRun.InvestigationSpec),
		MCPServers:                mcpServers,
		AllowedTools:              questionGatherAllowedTools(ctx.questionRun.InvestigationSpec),
		AllowedCommands:           []string{},
		TimeoutSeconds:            timeoutSeconds,
		ExpectedOutputs:           []string{"evidence_delta"},
		ArtifactDestination:       fmt.Sprintf("trace:%s", ctx.trace.Summary.TraceID),
		Intent:                    ctx.workflow.Intent,
		TraceID:                   ctx.trace.Summary.TraceID,
		WorkflowID:                ctx.workflow.ID,
		ConversationID:            ctx.trace.Summary.ConversationID,
		CaseID:                    ctx.trace.Summary.CaseID,
		ChannelID:                 ctx.ingestion.ChannelID,
		ThreadTS:                  ctx.ingestion.ThreadTS,
		ResponseMode:              ctx.workflow.ResponseMode,
		RecentConversationEntries: recentConversationEntries(store.ListConversationEntries(ctx.trace.Summary.ConversationID)),
		SessionScopeKind:          sessionScopeKind,
		SessionScopeID:            sessionScopeID,
		ParentSessionScopeKind:    parentScopeKind,
		ParentSessionScopeID:      parentScopeID,
		HarnessProfileID:          effectiveHarness.Profile.ID,
		HarnessOverlayVersion:     effectiveHarness.EffectiveOverlayVersion,
		MemoryBackend:             harness.DefaultMemoryBackend,
		AssistantPeerID:           fmt.Sprintf("rsi:%s:%s", cfg.Environment, role),
		UserPeerID:                workflowUserPeerID(store.ListConversationEntries(ctx.trace.Summary.ConversationID), sessionScopeKind, sessionScopeID),
	}
}

func buildQuestionExpandTask(cfg config.Config, store storepkg.Store, ctx questionRunContext) clients.RunnerTask {
	return buildQuestionGatherTask(cfg, store, ctx, roleQueueName(ctx.questionRun.Role))
}

func buildQuestionReduceTask(cfg config.Config, store storepkg.Store, ctx questionRunContext, queueName queue.QueueName) clients.RunnerTask {
	sessionScopeKind, sessionScopeID, parentScopeKind, parentScopeID := workflowSessionScope(ctx.trace, ctx.workflow)
	role := runnerRoleForQueue(queueName)
	effectiveHarness := harness.ResolveEffectiveConfig(store, role, cfg.DefaultReasoningVerbosity)
	totalTimeout := cfg.RunnerTaskTimeoutForRole(role)
	timeoutSeconds := int(questionRunFinalizationReserveSeconds(totalTimeout) / time.Second)
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	return clients.RunnerTask{
		TaskType:               "question_reduce",
		Repo:                   firstNonEmpty(ctx.questionRun.InvestigationSpec.Repo, cfg.DefaultRepo),
		RepoRef:                "main",
		Prompt:                 questionReducePrompt(ctx.questionRun.InvestigationSpec, ctx.questionRun.EvidenceLedger, ctx.questionRun.RunnerDiagnostics),
		SystemMessage:          "You are in the Slack Q&A reduce phase. Use only the supplied investigation spec, evidence ledger, and runner diagnostics. Do not call tools. Do not send Slack messages. Return JSON only with reply_markdown, confidence, completion_verdict, and termination_reason.",
		MCPServers:             []clients.RunnerMCPServer{},
		AllowedTools:           []string{},
		AllowedCommands:        []string{},
		TimeoutSeconds:         timeoutSeconds,
		ExpectedOutputs:        []string{"reply_markdown"},
		ArtifactDestination:    fmt.Sprintf("trace:%s", ctx.trace.Summary.TraceID),
		Intent:                 ctx.workflow.Intent,
		TraceID:                ctx.trace.Summary.TraceID,
		WorkflowID:             ctx.workflow.ID,
		ConversationID:         ctx.trace.Summary.ConversationID,
		CaseID:                 ctx.trace.Summary.CaseID,
		ChannelID:              ctx.ingestion.ChannelID,
		ThreadTS:               ctx.ingestion.ThreadTS,
		ResponseMode:           ctx.workflow.ResponseMode,
		SessionScopeKind:       sessionScopeKind,
		SessionScopeID:         sessionScopeID,
		ParentSessionScopeKind: parentScopeKind,
		ParentSessionScopeID:   parentScopeID,
		HarnessProfileID:       effectiveHarness.Profile.ID,
		HarnessOverlayVersion:  effectiveHarness.EffectiveOverlayVersion,
		MemoryBackend:          harness.DefaultMemoryBackend,
		AssistantPeerID:        fmt.Sprintf("rsi:%s:%s", cfg.Environment, role),
		UserPeerID:             workflowUserPeerID(store.ListConversationEntries(ctx.trace.Summary.ConversationID), sessionScopeKind, sessionScopeID),
	}
}

func questionGatherAllowedTools(spec questionrun.InvestigationSpec) []string {
	allowed := []string{}
	if strings.TrimSpace(spec.Repo) != "" {
		allowed = append(allowed, "repo.context", "repo.search", "repo.read_file", "github.repo_activity", "github.repo_context")
	}
	if spec.AlignmentRequired || strings.TrimSpace(spec.ProjectKey) != "" {
		allowed = append(allowed, "knowledge.context")
	}
	return uniqueStrings(allowed)
}

func questionExpandAllowedTools(spec questionrun.InvestigationSpec, _ bool) []string {
	return questionGatherAllowedTools(spec)
}

func questionGatherPrompt(spec questionrun.InvestigationSpec, ledger questionrun.EvidenceLedger) string {
	payload := map[string]any{
		"investigation_spec": spec,
		"evidence_ledger":    ledger,
		"gather_contract": map[string]any{
			"objective":        "Collect enough grounded evidence to answer the Slack question well without exhaustively searching every surface.",
			"retrieval_budget": maxInt(spec.RetrievalBudget, 1),
			"coverage_targets": questionGatherCoverageTargets(spec),
			"tooling_preferences": []string{
				"Use Slack MCP reads for Slack evidence.",
				"Use github.repo_activity and github.repo_context before broader repo search.",
				"Use repo.search or repo.read_file only when a specific file, subsystem, or claim needs verification.",
			},
			"stopping_rules": []string{
				"Stop once you have enough grounded evidence to answer the user request and cite the important surfaces.",
				"Do not repeatedly reread the same Slack channel or thread when existing evidence already covers it.",
				"If alignment evidence is thin, record the gap in open_questions or insufficiency_markers instead of broadening search indefinitely.",
				"Do not draft the final answer in this phase.",
			},
		},
	}
	body, _ := json.Marshal(payload)
	return string(body)
}

func questionRunGatherTimeout(total time.Duration) time.Duration {
	reasoningWindow := total - questionRunFinalizationReserveSeconds(total)
	if reasoningWindow <= 0 {
		return total
	}
	return reasoningWindow
}

func questionGatherSystemMessage(spec questionrun.InvestigationSpec) string {
	parts := []string{
		"You are in the Slack Q&A evidence-gather phase.",
		"Gather grounded evidence only; do not answer the user and do not send Slack messages.",
		"Prefer targeted Slack MCP reads plus governed repo and GitHub reads over broad exploratory loops.",
		"Use repo.search and repo.read_file only when a concrete file or subsystem needs verification.",
		"Stop once the evidence ledger covers the question, the bound thread, and the explicitly referenced Slack surfaces.",
		"Return JSON only with tool_calls, evidence_items, open_questions, draft_reply_candidates, insufficiency_markers, and confidence.",
	}
	if spec.AlignmentRequired {
		parts = append(parts, "If alignment evidence is incomplete, record that uncertainty explicitly instead of continuing wide searches.")
	}
	return strings.Join(parts, " ")
}

func questionGatherCoverageTargets(spec questionrun.InvestigationSpec) []string {
	targets := []string{
		fmt.Sprintf("Capture the main request from the bound reply target %s/%s.", firstNonEmpty(strings.TrimSpace(spec.ReplyTarget.ChannelID), "unknown-channel"), firstNonEmpty(strings.TrimSpace(spec.ReplyTarget.ThreadTS), "no-thread")),
	}
	if repo := strings.TrimSpace(spec.Repo); repo != "" {
		targets = append(targets, fmt.Sprintf("Capture recent repository progress for %s between %s and %s.", repo, firstNonEmpty(strings.TrimSpace(spec.Since), "the relevant start"), firstNonEmpty(strings.TrimSpace(spec.Until), "now")))
	}
	for _, surface := range spec.ReadSurfaces {
		channelID := strings.TrimSpace(surface.ChannelID)
		if channelID == "" {
			continue
		}
		if threadTS := strings.TrimSpace(surface.ThreadTS); threadTS != "" {
			targets = append(targets, fmt.Sprintf("Capture the salient evidence from Slack thread %s/%s.", channelID, threadTS))
			continue
		}
		targets = append(targets, fmt.Sprintf("Capture the salient evidence from Slack channel %s.", channelID))
	}
	if spec.AlignmentRequired || strings.TrimSpace(spec.ProjectKey) != "" {
		targets = append(targets, fmt.Sprintf("Capture the strongest available evidence about alignment with %s.", firstNonEmpty(strings.TrimSpace(spec.ProjectKey), "the referenced project")))
	}
	return uniqueStrings(targets)
}

func questionExpandPrompt(spec questionrun.InvestigationSpec, ledger questionrun.EvidenceLedger) string {
	return questionGatherPrompt(spec, ledger)
}

func questionReducePrompt(spec questionrun.InvestigationSpec, ledger questionrun.EvidenceLedger, diagnostics map[string]any) string {
	payload := map[string]any{
		"investigation_spec": spec,
		"evidence_ledger":    ledger,
		"runner_diagnostics": diagnostics,
	}
	body, _ := json.Marshal(payload)
	return string(body)
}

func parseQuestionRunStructuredOutput[T any](resp clients.RunnerResponse) (T, error) {
	var out T
	raw, ok := resp.Raw["structured_output"]
	if !ok {
		return out, fmt.Errorf("runner response missing structured_output")
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, err
	}
	return out, nil
}

func questionRunEvidenceLedgerFromRunnerRaw(raw map[string]any) questionrun.EvidenceLedger {
	var out questionrun.EvidenceLedger
	value, ok := raw["evidence_ledger"]
	if !ok || value == nil {
		return out
	}
	data, err := json.Marshal(value)
	if err != nil {
		return out
	}
	_ = json.Unmarshal(data, &out)
	return out
}

func mergeRunnerEvidenceLedger(base questionrun.EvidenceLedger, overlay questionrun.EvidenceLedger) questionrun.EvidenceLedger {
	out := base
	if overlay.InvestigationSpec != nil {
		spec := *overlay.InvestigationSpec
		out.InvestigationSpec = &spec
	}
	if strings.TrimSpace(overlay.UserRequest) != "" {
		out.UserRequest = overlay.UserRequest
	}
	if strings.TrimSpace(overlay.ReplyTarget.ChannelID) != "" {
		out.ReplyTarget.ChannelID = overlay.ReplyTarget.ChannelID
	}
	if strings.TrimSpace(overlay.ReplyTarget.ThreadTS) != "" {
		out.ReplyTarget.ThreadTS = overlay.ReplyTarget.ThreadTS
	}
	if strings.TrimSpace(overlay.Repo) != "" {
		out.Repo = overlay.Repo
	}
	if strings.TrimSpace(overlay.ProjectKey) != "" {
		out.ProjectKey = overlay.ProjectKey
	}
	if strings.TrimSpace(overlay.Since) != "" {
		out.Since = overlay.Since
	}
	if strings.TrimSpace(overlay.Until) != "" {
		out.Until = overlay.Until
	}
	if strings.TrimSpace(overlay.Prompt.RenderedText) != "" || strings.TrimSpace(overlay.Prompt.RawText) != "" {
		out.Prompt = overlay.Prompt
	}
	out.AlignmentRequired = out.AlignmentRequired || overlay.AlignmentRequired
	out.AlignmentDegraded = out.AlignmentDegraded || overlay.AlignmentDegraded
	if overlay.AlignmentLedger != nil {
		out.AlignmentLedger = overlay.AlignmentLedger
	}
	out.ToolCalls = dedupeQuestionRunToolCalls(append(out.ToolCalls, overlay.ToolCalls...))
	out.EvidenceItems = dedupeEvidenceItems(append(out.EvidenceItems, overlay.EvidenceItems...))
	out.OpenQuestions = uniqueStrings(append(out.OpenQuestions, overlay.OpenQuestions...))
	out.MissingEvidence = uniqueStrings(append(out.MissingEvidence, overlay.MissingEvidence...))
	out.DraftReplyCandidates = uniqueStrings(append(out.DraftReplyCandidates, overlay.DraftReplyCandidates...))
	out.TerminationReason = firstNonEmpty(overlay.TerminationReason, out.TerminationReason)
	return out
}

func mergeEvidenceDelta(ledger questionrun.EvidenceLedger, delta questionrun.EvidenceDelta) questionrun.EvidenceLedger {
	ledger.ToolCalls = dedupeQuestionRunToolCalls(append(ledger.ToolCalls, delta.ToolCalls...))
	ledger.EvidenceItems = dedupeEvidenceItems(append(ledger.EvidenceItems, delta.EvidenceItems...))
	ledger.OpenQuestions = uniqueStrings(append(delta.OpenQuestions, ledger.OpenQuestions...))
	ledger.MissingEvidence = uniqueStrings(append(delta.InsufficiencyMarks, ledger.MissingEvidence...))
	ledger.DraftReplyCandidates = uniqueStrings(append(delta.DraftReplyCandidates, ledger.DraftReplyCandidates...))
	return ledger
}

func isQuestionRunBoundedStop(terminationReason string) bool {
	switch strings.TrimSpace(terminationReason) {
	case "task_timeout", "iteration_budget_exhausted":
		return true
	default:
		return false
	}
}

func questionRunToolCallsFromRunnerRaw(raw map[string]any) []questionrun.ToolCall {
	records := toolCallRecordsFromRunnerRaw(raw)
	out := make([]questionrun.ToolCall, 0, len(records))
	for _, record := range records {
		item := questionrun.ToolCall{
			ToolName:        record.ToolName,
			ToolCallID:      record.ToolCallID,
			Request:         record.Request,
			Summary:         record.Summary,
			Status:          record.Status,
			RawArtifactRefs: append([]string(nil), record.RawArtifactRefs...),
		}
		out = append(out, item)
	}
	return out
}

func questionRunStructuredOutputForReply(ctx questionRunContext, result questionrun.Result) runnerutil.StructuredOutput {
	replyBody := strings.TrimSpace(result.ReplyMarkdown)
	return runnerutil.StructuredOutput{
		ContextSummary: "Read-heavy Slack Q&A reducer synthesized a reply from the evidence ledger.",
		ReplyDraft:     replyBody,
		FinalAnswer:    replyBody,
		Confidence:     result.Confidence,
		ProposedActions: []runnerutil.ProposedAction{{
			Kind:      string(action.KindSlackPost),
			TargetRef: firstNonEmpty(ctx.ingestion.ChannelID, ctx.ingestion.ThreadTS, ctx.trace.Summary.TraceID),
			RequestPayload: map[string]any{
				"channel_id": ctx.ingestion.ChannelID,
				"thread_ts":  ctx.ingestion.ThreadTS,
				"body":       replyBody,
				"draft_body": replyBody,
				"final_body": replyBody,
			},
			ApprovalMode:   "not_required",
			IdempotencyKey: fmt.Sprintf("%s:%s:%s:question_run", ctx.ingestion.ChannelID, ctx.ingestion.ThreadTS, ctx.trace.Summary.TraceID),
			Rationale:      "Post the reducer-authored reply back into the bound Slack thread.",
			EvidenceRefs: []events.EvidenceRef{{
				Kind:    "trace",
				Ref:     ctx.trace.Summary.TraceID,
				Summary: "question_run reducer result",
			}},
		}},
	}
}

func questionRunReasoningSteps(ctx questionRunContext, result questionrun.Result, createdAt time.Time) []events.ReasoningStep {
	summary := fmt.Sprintf("Reduced a reply from %d grounded evidence item(s).", len(ctx.questionRun.EvidenceLedger.EvidenceItems))
	if result.CompletionVerdict == "partial" {
		summary = partialCompletionReasoningSummary(result.TerminationReason)
	}
	steps := []events.ReasoningStep{{
		ID:         fmt.Sprintf("reason-question-run-reduce-%d", createdAt.UnixNano()),
		TraceID:    ctx.trace.Summary.TraceID,
		WorkflowID: ctx.workflow.ID,
		StepType:   "question_run_reduce",
		Summary:    summary,
		Confidence: result.Confidence,
		Decision:   result.ReplyMarkdown,
		CreatedAt:  createdAt,
	}}
	if result.AlignmentDegraded && strings.TrimSpace(result.AlignmentNotice) != "" {
		steps = append(steps, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-question-run-alignment-%d", createdAt.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.workflow.ID,
			StepType:   "alignment_degraded",
			Summary:    result.AlignmentNotice,
			Confidence: 1.0,
			Decision:   "alignment_degraded",
			CreatedAt:  createdAt,
		})
	}
	return steps
}

func traceToolCallsFromQuestionRunLedger(ctx questionRunContext, calls []questionrun.ToolCall, createdAt time.Time) []events.ToolCallRecord {
	out := make([]events.ToolCallRecord, 0, len(calls))
	for idx, call := range calls {
		out = append(out, events.ToolCallRecord{
			ID:              fmt.Sprintf("question-run-tool-%d-%d", createdAt.UnixNano(), idx),
			TraceID:         ctx.trace.Summary.TraceID,
			WorkflowID:      ctx.workflow.ID,
			ConversationID:  ctx.trace.Summary.ConversationID,
			CaseID:          ctx.trace.Summary.CaseID,
			ToolName:        call.ToolName,
			ToolCallID:      firstNonEmpty(call.ToolCallID, call.ToolName),
			Request:         cloneAnyMap(call.Request),
			Summary:         call.Summary,
			RawArtifactRefs: append([]string(nil), call.RawArtifactRefs...),
			Status:          call.Status,
			CreatedAt:       createdAt,
		})
	}
	return out
}

func isTerminalQuestionRunStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case string(transition.QuestionRunStateCompleted), string(transition.QuestionRunStateNeedsHuman), string(transition.QuestionRunStateFailed), string(transition.QuestionRunStateSuperseded):
		return true
	default:
		return false
	}
}

func roleQueueName(role string) queue.QueueName {
	if strings.EqualFold(strings.TrimSpace(role), "proactive") {
		return queue.ProactiveQueue
	}
	return queue.WorkflowQueue
}

func slackMessageRef(channelID string, ts string) string {
	if strings.TrimSpace(channelID) == "" || strings.TrimSpace(ts) == "" {
		return ""
	}
	return fmt.Sprintf("slack://%s/%s", channelID, ts)
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return strings.TrimSpace(value[:limit])
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
