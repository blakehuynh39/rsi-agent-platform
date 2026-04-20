package control

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
	"github.com/piplabs/rsi-agent-platform/internal/workflowplan"
)

const (
	controlPhaseCollectContext             = "collect_context"
	controlPhaseReplyPost                  = "reply_post"
	defaultSlackMCPServerURL               = "https://mcp.slack.com/mcp"
	defaultNotionMCPServerURL              = "https://mcp.notion.com/mcp"
	partialCompletionNoticeGeneric         = "Partial answer: I had to stop before I could finish a deeper pass. This is the best grounded answer so far."
	partialCompletionNoticeIterationBudget = "Partial answer: I hit my iteration budget before I could finish a deeper pass. This is the best grounded answer so far."
	partialCompletionNoticeTaskTimeout     = "Partial answer: I hit the workflow time limit before I could finish a deeper pass. This is the best grounded answer so far."
	partialCompletionNoticeOutputBudget    = "Partial answer: I hit my response output budget before I could finish a deeper pass. This is the best grounded answer so far."
)

type workflowContext struct {
	trace     events.Trace
	workflow  storepkg.Workflow
	ingestion slackpkg.Ingestion
}

type workflowLocator struct {
	traceID     string
	workflowID  string
	ingestionID string
}

func RunWorker(cfg config.Config, store storepkg.Store) error {
	runnerClients := map[string]*clients.RunnerClient{
		"prod":      clients.NewRunnerClientWithTimeout(cfg.RunnerURLForRole("prod"), cfg.RunnerTimeoutForRole("prod")),
		"proactive": clients.NewRunnerClientWithTimeout(cfg.RunnerURLForRole("proactive"), cfg.RunnerTimeoutForRole("proactive")),
	}
	toolClient := clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL)
	workerID := fmt.Sprintf("%s-worker", cfg.ServiceName)
	runnerEffectLease := cfg.EffectLeaseDuration(cfg.WorkItemLeaseDuration, "prod", "proactive")
	for {
		effect, claimed, err := claimNextExecutionEffect(store, workerID, runnerEffectLease)
		if err != nil {
			return err
		}
		if !claimed {
			if err := activateDueWorkflowLineRetries(cfg, store, time.Now().UTC()); err != nil {
				return err
			}
			time.Sleep(cfg.WorkerPollInterval)
			continue
		}
		handleClaimedExecutionEffect(cfg, store, runnerClients, toolClient, effect)
	}
}

func RunActionWorker(cfg config.Config, store storepkg.Store) error {
	toolClient := clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL)
	workerID := fmt.Sprintf("%s-action-worker", cfg.ServiceName)
	for {
		effect, ok, err := claimNextActionEffect(store, "control", workerID, cfg.WorkItemLeaseDuration)
		if err != nil {
			return err
		}
		if !ok {
			time.Sleep(cfg.WorkerPollInterval)
			continue
		}
		if err := processControlActionEffect(cfg, store, toolClient, effect); err != nil {
			log.Printf("control-plane action worker effect=%s aggregate=%s error=%v", effect.ID, effect.AggregateID, err)
			continue
		}
	}
}

func startWorkflowViaCommand(cfg config.Config, store storepkg.Store, workflowID string, occurredAt time.Time, resumeQueue queue.QueueName) error {
	_, err := submitWorkflowCommand(store, workflowID, transition.CommandWorkflowStarted, cfg.ServiceName, occurredAt, map[string]any{
		"default_repo":         cfg.DefaultRepo,
		"allowed_target_repos": append([]string(nil), cfg.AllowedTargetRepos...),
		"knowledge_base_url":   cfg.DefaultKnowledgeBaseURL,
		"sandbox_namespace":    cfg.SandboxNamespace,
		"resume_queue":         string(resumeQueue),
	})
	return err
}

func finalizeWorkflowFailure(cfg config.Config, store storepkg.Store, workflow workflowLocator, procErr error) error {
	failure := workflowFailure{
		Summary: strings.TrimSpace(procErr.Error()),
	}
	var detailed *workflowFailureError
	if errors.As(procErr, &detailed) && detailed != nil {
		failure = detailed.failure
	}
	return finalizeWorkflowFailureWithDetails(cfg, store, workflow, failure)
}

func handleClaimedWorkflowRunnerEffect(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, effect transition.EffectExecution) {
	workflow, _ := workflowLocatorForEffect(store, effect)
	if err := processWorkflowRunnerEffect(cfg, store, runnerClients, effect); err != nil {
		var detailed *workflowFailureError
		if errors.As(err, &detailed) && detailed != nil {
			if finalizeErr := finalizeWorkflowFailureWithDetails(cfg, store, workflow, detailed.failure); finalizeErr != nil {
				_ = failClaimedEffect(store, effect, finalizeErr.Error())
				log.Printf("control-plane runner effect=%s aggregate=%s finalize_error=%v", effect.ID, effect.AggregateID, finalizeErr)
			} else {
				if _, _, reconcileErr := store.ReconcileWorkflowTrace(workflow.workflowID); reconcileErr != nil {
					_ = failClaimedEffect(store, effect, reconcileErr.Error())
					log.Printf("control-plane runner effect=%s aggregate=%s reconcile_error=%v", effect.ID, effect.AggregateID, reconcileErr)
				} else if completeErr := completeClaimedEffect(store, effect, fmt.Sprintf("trace:%s:workflow-failed", workflow.traceID)); completeErr != nil {
					log.Printf("control-plane runner effect=%s aggregate=%s finalize_complete_error=%v", effect.ID, effect.AggregateID, completeErr)
				}
			}
		} else if failErr := failClaimedEffect(store, effect, err.Error()); failErr != nil {
			log.Printf("control-plane runner effect=%s aggregate=%s fail_error=%v", effect.ID, effect.AggregateID, failErr)
		}
		log.Printf("control-plane runner effect=%s aggregate=%s error=%v", effect.ID, effect.AggregateID, err)
	}
}

func handleClaimedExecutionEffect(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, toolClient *clients.ToolGatewayClient, effect transition.EffectExecution) {
	switch {
	case effect.MachineKind == transition.MachineWorkflow && effect.EffectKind == transition.EffectInvokeRunner:
		handleClaimedWorkflowRunnerEffect(cfg, store, runnerClients, effect)
	case effect.MachineKind == transition.MachineQuestionRun:
		handleClaimedQuestionRunEffect(cfg, store, runnerClients, toolClient, effect)
	default:
		_ = failClaimedEffect(store, effect, fmt.Sprintf("unsupported execution effect %s/%s", effect.MachineKind, effect.EffectKind))
	}
}

func processWorkflowRunnerEffect(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, effect transition.EffectExecution) error {
	ctx, queueName, err := loadWorkflowContextForEffect(store, effect)
	if err != nil {
		_ = failClaimedEffect(store, effect, err.Error())
		return err
	}
	if ctx, err = refreshWorkflowContextState(store, ctx); err != nil {
		return err
	}
	if isTerminalWorkflowStatus(ctx.workflow.Status) || isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return completeClaimedEffect(store, effect, ctx.trace.Summary.TraceID)
	}
	runnerClient := runnerClients[runnerRoleForQueue(queueName)]
	if runnerClient == nil {
		return fmt.Errorf("runner client unavailable for queue %s", queueName)
	}
	contextSummary, contextRefs := contextFromTrace(ctx.trace)
	runnerStarted := time.Now().UTC()
	runnerTask := buildRunnerTask(cfg, store, runnerRoleForQueue(queueName), ctx.trace, ctx.workflow, ctx.ingestion, contextSummary, contextRefs)
	runnerResp, err := runnerClient.Execute(runnerTask)
	if err != nil {
		return &workflowFailureError{failure: workflowFailureFromRunnerError(err)}
	}
	runnerPostProcessingFailure := func(stage string, err error) error {
		if err == nil {
			return nil
		}
		return &workflowFailureError{failure: workflowFailureFromRunnerPostProcessing(runnerResp, stage, err)}
	}
	if err := runnerutil.PersistHarnessExecution(
		store,
		runnerResp,
		runnerRoleForQueue(queueName),
		effect.ID,
		ctx.trace.Summary.TraceID,
		"",
		runnerTask.HarnessProfileID,
		runnerTask.HarnessOverlayVersion,
		runnerTask.SessionScopeKind,
		runnerTask.SessionScopeID,
		runnerTask.ParentSessionScopeKind,
		runnerTask.ParentSessionScopeID,
	); err != nil {
		return runnerPostProcessingFailure("persist_harness_execution", err)
	}
	if ctx, err = refreshWorkflowContextState(store, ctx); err != nil {
		return runnerPostProcessingFailure("refresh_workflow_context_after_harness_execution", err)
	}
	if isTerminalWorkflowStatus(ctx.workflow.Status) || isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return completeClaimedEffect(store, effect, ctx.trace.Summary.TraceID)
	}
	if !runnerResp.OK {
		if strings.TrimSpace(stringValue(runnerResp.Raw["failure_class"])) == "runner_reply_delivery_uncertain" {
			runnerCompleted := time.Now().UTC()
			runnerDiagnostics := cloneStringAnyMap(mapValue(runnerResp.Raw["runner_diagnostics"]))
			runnerDiagnostics = mergeWorkflowRunnerDiagnostics(runnerDiagnostics, runnerResp.Raw)
			runnerToolCalls := bindRunnerToolCallRecords(toolCallRecordsFromRunnerRaw(runnerResp.Raw), ctx.trace, ctx.workflow)
			payload := map[string]any{
				"last_error":         firstNonEmpty(strings.TrimSpace(runnerResp.Message), "runner reply delivery became uncertain after a Slack MCP write attempt"),
				"failure_class":      "runner_reply_delivery_uncertain",
				"failure_summary":    "Runner attempted a Slack MCP write but did not complete with a durable delivery contract, so the workflow requires human review to avoid duplicate posting.",
				"runner_diagnostics": runnerDiagnostics,
				"tool_calls":         runnerToolCalls,
				"trace_events": []events.TraceEvent{
					{
						TraceID:     ctx.trace.Summary.TraceID,
						IngestionID: ctx.trace.Summary.IngestionID,
						WorkflowID:  ctx.trace.Summary.WorkflowID,
						Plane:       "execution",
						Service:     "runner",
						Actor:       ctx.workflow.AssignedBot,
						EventType:   "runner.reply_delivery_uncertain",
						Status:      events.StatusNeedsHuman,
						StartedAt:   runnerStarted,
						EndedAt:     &runnerCompleted,
						Description: "Runner attempted a Slack MCP reply but delivery could not be finalized safely.",
					},
				},
				"reasoning_steps": []events.ReasoningStep{
					{
						ID:         fmt.Sprintf("reason-reply-delivery-uncertain-%d", runnerCompleted.UnixNano()),
						TraceID:    ctx.trace.Summary.TraceID,
						WorkflowID: ctx.trace.Summary.WorkflowID,
						StepType:   "reply_delivery_uncertain",
						Summary:    "Runner attempted a Slack MCP write but the final delivery contract was not durable enough to safely retry automatically.",
						Confidence: 1.0,
						Decision:   "needs_human",
						CreatedAt:  runnerCompleted,
					},
				},
			}
			if replyDelivery, ok := workflowReplyDelivery(runnerResp.Raw, ctx.ingestion.ChannelID, ctx.ingestion.ThreadTS); ok {
				replyDelivery.TraceID = ctx.trace.Summary.TraceID
				replyDelivery.WorkflowID = ctx.workflow.ID
				replyDelivery.ConversationID = ctx.trace.Summary.ConversationID
				replyDelivery.CaseID = ctx.trace.Summary.CaseID
				replyDelivery.CreatedAt = runnerCompleted
				payload["slack_actions"] = []events.SlackActionRecord{replyDelivery}
			}
			if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandWorkflowExecutionNeedsHuman, cfg.ServiceName, runnerCompleted, payload); err != nil {
				return runnerPostProcessingFailure("submit_runner_reply_delivery_uncertain", err)
			}
			if _, _, err := store.ReconcileWorkflowTrace(ctx.workflow.ID); err != nil {
				return err
			}
			return completeClaimedEffect(store, effect, fmt.Sprintf("trace:%s:runner-reply-delivery-uncertain", ctx.trace.Summary.TraceID))
		}
		return &workflowFailureError{failure: workflowFailureFromRunnerResponse(runnerResp)}
	}
	runnerOutput, err := runnerutil.ParseStructuredOutput(runnerResp)
	if err != nil {
		return &workflowFailureError{failure: workflowFailureFromStructuredOutputError(runnerResp, err)}
	}
	completionVerdict := workflowCompletionVerdict(runnerResp.Raw)
	terminationReason := workflowTerminationReason(runnerResp.Raw)
	if completionVerdict == "partial" {
		runnerOutput = standardizePartialWorkflowReply(runnerOutput, terminationReason)
	}
	if ctx, err = refreshWorkflowContextState(store, ctx); err != nil {
		return runnerPostProcessingFailure("refresh_workflow_context_after_structured_output", err)
	}
	if isTerminalWorkflowStatus(ctx.workflow.Status) || isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return completeClaimedEffect(store, effect, ctx.trace.Summary.TraceID)
	}
	runnerCompleted := time.Now().UTC()
	if err := persistKnowledgeDrafts(store, ctx.trace, runnerOutput.KnowledgeDrafts, runnerCompleted); err != nil {
		return runnerPostProcessingFailure("persist_knowledge_drafts", err)
	}

	proposedReplyAction := firstSlackPostAction(runnerOutput.ProposedActions)
	replyBody := firstNonEmpty(
		strings.TrimSpace(runnerOutput.FinalAnswer),
		strings.TrimSpace(runnerOutput.ReplyDraft),
		strings.TrimSpace(stringValueFromMap(proposedReplyAction.RequestPayload, "body")),
		strings.TrimSpace(stringValueFromMap(proposedReplyAction.RequestPayload, "final_body")),
		strings.TrimSpace(stringValueFromMap(proposedReplyAction.RequestPayload, "draft_body")),
		strings.TrimSpace(runnerResp.Message),
	)
	replyChannelID := firstNonEmpty(strings.TrimSpace(stringValueFromMap(proposedReplyAction.RequestPayload, "channel_id")), ctx.ingestion.ChannelID)
	replyThreadTS := firstNonEmpty(strings.TrimSpace(stringValueFromMap(proposedReplyAction.RequestPayload, "thread_ts")), ctx.ingestion.ThreadTS)
	replyThreadKey := ctx.trace.Summary.ThreadKey
	if replyChannelID != ctx.ingestion.ChannelID || replyThreadTS != ctx.ingestion.ThreadTS {
		replyThreadKey = ""
	}
	allowed, policyVerdict := replyPolicy(store, ctx.workflow.Kind, replyThreadKey, replyChannelID)
	nativeMCPEnabled := workflowNativeMCPEnabled(runnerResp.Raw)
	replyDelivery, hasReplyDelivery := workflowReplyDelivery(runnerResp.Raw, replyChannelID, replyThreadTS)
	if hasReplyDelivery && strings.TrimSpace(replyBody) == "" {
		replyBody = strings.TrimSpace(replyDelivery.FinalBody)
	}

	finalReasoning := runnerutil.ToTraceReasoning(ctx.trace.Summary.TraceID, ctx.trace.Summary.WorkflowID, runnerOutput, runnerCompleted)
	if runnerOutput.SelfCritique != "" {
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-self-critique-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "self_critique",
			Summary:    runnerOutput.SelfCritique,
			Confidence: runnerOutput.Confidence,
			CreatedAt:  runnerCompleted,
		})
	}
	finalReasoning = append(finalReasoning, outcomeHypothesisReasoning(ctx.trace, ctx.workflow, runnerOutput.OutcomeHypotheses, runnerCompleted)...)
	if completionVerdict == "partial" {
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-partial-completion-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "completion_contract",
			Summary:    partialCompletionReasoningSummary(terminationReason),
			Confidence: 1.0,
			Decision:   "partial_completion",
			CreatedAt:  runnerCompleted,
		})
	}
	finalReasoning = append(finalReasoning, events.ReasoningStep{
		ID:         fmt.Sprintf("reason-final-%d", runnerCompleted.UnixNano()),
		TraceID:    ctx.trace.Summary.TraceID,
		WorkflowID: ctx.trace.Summary.WorkflowID,
		StepType:   "final_answer_rationale",
		Summary:    firstNonEmpty(runnerOutput.ContextSummary, "Prepared final response from collected context."),
		Confidence: runnerOutput.Confidence,
		Decision:   replyBody,
		CreatedAt:  runnerCompleted,
	})
	runnerDiagnostics := cloneStringAnyMap(mapValue(runnerResp.Raw["runner_diagnostics"]))
	runnerDiagnostics = mergeWorkflowRunnerDiagnostics(runnerDiagnostics, runnerResp.Raw)
	runnerToolCalls := bindRunnerToolCallRecords(
		toolCallRecordsFromRunnerRaw(runnerResp.Raw),
		ctx.trace,
		ctx.workflow,
	)
	runnerEvents := []events.TraceEvent{
		{
			TraceID:     ctx.trace.Summary.TraceID,
			IngestionID: ctx.trace.Summary.IngestionID,
			WorkflowID:  ctx.trace.Summary.WorkflowID,
			Plane:       "execution",
			Service:     "runner",
			Actor:       ctx.workflow.AssignedBot,
			EventType:   "runner.completed",
			Status:      events.StatusCompleted,
			StartedAt:   runnerStarted,
			EndedAt:     &runnerCompleted,
			Description: "Runner returned visible reasoning.",
		},
	}
	var (
		replyAction    action.Intent
		draftEvents    []events.TraceEvent
		draftReasoning []events.ReasoningStep
		slackActions   []events.SlackActionRecord
	)
	if strings.TrimSpace(replyBody) != "" && strings.TrimSpace(proposedReplyAction.Kind) == "" && !nativeMCPEnabled {
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-action-contract-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "action_contract_blocked",
			Summary:    "Runner produced a reply draft but omitted the required explicit slack_post action contract.",
			Confidence: 0.95,
			Decision:   "needs_human",
			CreatedAt:  runnerCompleted,
		})
		runnerEvents[0].Description = "Runner returned visible reasoning and a reply draft but omitted the explicit slack_post action."
		if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandWorkflowExecutionNeedsHuman, cfg.ServiceName, runnerCompleted, map[string]any{
			"last_error":         "runner produced a reply without the required explicit slack_post action",
			"failure_class":      "missing_explicit_action_contract",
			"failure_summary":    "Runner produced a reply draft but omitted the required explicit slack_post action.",
			"runner_diagnostics": runnerDiagnostics,
			"repair_attempted":   boolValue(runnerResp.Raw["repair_attempted"]),
			"repair_succeeded":   boolValue(runnerResp.Raw["repair_succeeded"]),
			"trace_events":       runnerEvents,
			"reasoning_steps":    finalReasoning,
			"tool_calls":         runnerToolCalls,
		}); err != nil {
			return runnerPostProcessingFailure("submit_workflow_execution_needs_human", err)
		}
		if ctx, err = refreshWorkflowContextState(store, ctx); err != nil {
			return err
		}
		if ctx.workflow.Status != string(transition.WorkflowStateNeedsHuman) {
			return &workflowFailureError{failure: workflowFailureFromRunnerStateInvariant(runnerResp, ctx.workflow.ID, transition.WorkflowStateNeedsHuman, ctx.workflow.Status)}
		}
		if _, _, err := store.ReconcileWorkflowTrace(ctx.workflow.ID); err != nil {
			return err
		}
		return completeClaimedEffect(store, effect, fmt.Sprintf("trace:%s:runner-blocked", ctx.trace.Summary.TraceID))
	}
	if nativeMCPEnabled && strings.TrimSpace(replyBody) != "" && !hasReplyDelivery && strings.TrimSpace(proposedReplyAction.Kind) == "" {
		summary := "Runner produced a grounded reply but did not report a durable Slack MCP reply_delivery."
		if !allowed {
			summary = fmt.Sprintf("Runner produced a grounded reply but Slack posting is blocked by policy: %s.", policyVerdict)
		}
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-mcp-reply-missing-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "reply_delivery_missing",
			Summary:    summary,
			Confidence: 0.95,
			Decision:   "needs_human",
			CreatedAt:  runnerCompleted,
		})
		if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandWorkflowExecutionNeedsHuman, cfg.ServiceName, runnerCompleted, map[string]any{
			"last_error":         summary,
			"failure_class":      "missing_reply_delivery",
			"failure_summary":    summary,
			"runner_diagnostics": runnerDiagnostics,
			"trace_events":       runnerEvents,
			"reasoning_steps":    finalReasoning,
			"tool_calls":         runnerToolCalls,
		}); err != nil {
			return runnerPostProcessingFailure("submit_workflow_execution_needs_human_missing_reply_delivery", err)
		}
		if _, _, err := store.ReconcileWorkflowTrace(ctx.workflow.ID); err != nil {
			return err
		}
		return completeClaimedEffect(store, effect, fmt.Sprintf("trace:%s:runner-mcp-needs-human", ctx.trace.Summary.TraceID))
	}
	if hasReplyDelivery {
		replyDelivery.TraceID = ctx.trace.Summary.TraceID
		replyDelivery.WorkflowID = ctx.workflow.ID
		replyDelivery.ConversationID = ctx.trace.Summary.ConversationID
		replyDelivery.CaseID = ctx.trace.Summary.CaseID
		replyDelivery.PolicyVerdict = policyVerdict
		replyDelivery.CreatedAt = runnerCompleted
		slackActions = []events.SlackActionRecord{replyDelivery}
		runnerEvents = append(runnerEvents, events.TraceEvent{
			TraceID:     ctx.trace.Summary.TraceID,
			IngestionID: ctx.trace.Summary.IngestionID,
			WorkflowID:  ctx.trace.Summary.WorkflowID,
			Plane:       "execution",
			Service:     "runner",
			Actor:       ctx.workflow.AssignedBot,
			EventType:   "slack.reply.posted",
			Status:      events.StatusCompleted,
			StartedAt:   runnerStarted,
			EndedAt:     &runnerCompleted,
			Description: "Runner delivered the final Slack reply through Slack MCP.",
		})
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-reply-delivered-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "reply_delivery",
			Summary:    "Runner delivered the final Slack reply through Slack MCP.",
			Confidence: 1.0,
			Decision:   "reply_delivered",
			CreatedAt:  runnerCompleted,
		})
	}
	if !hasReplyDelivery && strings.TrimSpace(proposedReplyAction.Kind) != "" {
		replyAction, draftEvents, draftReasoning, err = draftSlackPostAction(cfg, store, queueName, ctx, runnerOutput, replyBody, allowed, policyVerdict, completionVerdict, runnerCompleted)
		if err != nil {
			return runnerPostProcessingFailure("draft_slack_post_action", err)
		}
	}
	finalReasoning = append(finalReasoning, draftReasoning...)
	runnerDescription := "Runner returned visible reasoning."
	hasReplyAction := strings.TrimSpace(replyAction.ID) != ""
	runnerCommand := transition.WorkflowExecutionCompletionCommand(completionVerdict, hasReplyAction)
	if hasReplyAction {
		runnerDescription = "Runner returned visible reasoning and an explicit Slack reply action."
	} else if hasReplyDelivery {
		runnerDescription = "Runner returned visible reasoning and delivered a Slack reply through Slack MCP."
	}
	if completionVerdict == "partial" {
		runnerDescription = partialCompletionRunnerDescription(terminationReason, hasReplyAction)
	}
	expectedWorkflowState := transition.WorkflowStateCompleted
	if hasReplyAction {
		expectedWorkflowState = transition.WorkflowStateReplyPending
	}
	runnerEvents[0].Description = runnerDescription
	if ctx, err = refreshWorkflowContextState(store, ctx); err != nil {
		return runnerPostProcessingFailure("refresh_workflow_context_before_runner_transition", err)
	}
	if isTerminalWorkflowStatus(ctx.workflow.Status) || isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return completeClaimedEffect(store, effect, ctx.trace.Summary.TraceID)
	}
	if _, err := submitWorkflowCommand(store, ctx.workflow.ID, runnerCommand, cfg.ServiceName, runnerCompleted, map[string]any{
		"resume_queue":       string(queueName),
		"reply_action_id":    replyAction.ID,
		"repair_attempted":   boolValue(runnerResp.Raw["repair_attempted"]),
		"repair_succeeded":   boolValue(runnerResp.Raw["repair_succeeded"]),
		"runner_diagnostics": runnerDiagnostics,
		"trace_events":       append(runnerEvents, draftEvents...),
		"reasoning_steps":    finalReasoning,
		"tool_calls":         runnerToolCalls,
		"slack_actions":      slackActions,
	}); err != nil {
		return runnerPostProcessingFailure("submit_runner_completion", err)
	}
	if ctx, err = refreshWorkflowContextState(store, ctx); err != nil {
		return err
	}
	if ctx.workflow.Status != string(expectedWorkflowState) {
		return &workflowFailureError{failure: workflowFailureFromRunnerStateInvariant(runnerResp, ctx.workflow.ID, expectedWorkflowState, ctx.workflow.Status)}
	}
	if !hasReplyAction {
		if _, _, err := store.ReconcileWorkflowTrace(ctx.workflow.ID); err != nil {
			return err
		}
	}
	return completeClaimedEffect(store, effect, fmt.Sprintf("trace:%s:runner", ctx.trace.Summary.TraceID))
}

func submitWorkflowContextCompleted(cfg config.Config, store storepkg.Store, ctx workflowContext, resumeQueue queue.QueueName, occurredAt time.Time) error {
	refreshedTrace, ok := store.GetTrace(ctx.trace.Summary.TraceID)
	if ok {
		ctx.trace = refreshedTrace
	}
	if isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return nil
	}
	executionStrategy := workflowExecutionStrategy(ctx.workflow)
	executionRole := runnerRoleForQueue(resumeQueue)
	executionStartedType := "runner.started"
	executionStartedService := "runner"
	executionStartedDescription := "Runner task dispatched with verbose reasoning enabled."
	if executionStrategy == "read_heavy_slack_qna" {
		executionStartedType = "question_run.started"
		executionStartedService = cfg.ServiceName
		executionStartedDescription = "Question-run child machine dispatched with deterministic compiler inputs."
	}
	contextSummary, contextRefs := contextFromTrace(ctx.trace)
	if len(contextRefs) > 0 {
		if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandContextCompleted, cfg.ServiceName, occurredAt, map[string]any{
			"tool_count":         len(contextRefs),
			"resume_queue":       string(resumeQueue),
			"execution_strategy": executionStrategy,
			"execution_role":     executionRole,
			"trace_events": []events.TraceEvent{
				{
					TraceID:     ctx.trace.Summary.TraceID,
					IngestionID: ctx.trace.Summary.IngestionID,
					WorkflowID:  ctx.trace.Summary.WorkflowID,
					Plane:       "control",
					Service:     cfg.ServiceName,
					Actor:       "worker",
					EventType:   "context.collected",
					Status:      events.StatusCompleted,
					StartedAt:   occurredAt,
					Description: firstNonEmpty(contextSummary, "Context collection phase completed."),
				},
				{
					TraceID:     ctx.trace.Summary.TraceID,
					IngestionID: ctx.trace.Summary.IngestionID,
					WorkflowID:  ctx.trace.Summary.WorkflowID,
					Plane:       "execution",
					Service:     executionStartedService,
					Actor:       ctx.workflow.AssignedBot,
					EventType:   executionStartedType,
					Status:      events.StatusRunning,
					StartedAt:   occurredAt,
					Description: executionStartedDescription,
				},
			},
			"reasoning_steps": []events.ReasoningStep{
				{
					ID:           fmt.Sprintf("reason-context-%d", occurredAt.UnixNano()),
					TraceID:      ctx.trace.Summary.TraceID,
					WorkflowID:   ctx.trace.Summary.WorkflowID,
					StepType:     "context_collected",
					Summary:      firstNonEmpty(contextSummary, "No external context tools were required."),
					EvidenceRefs: evidenceRefsFromContext(contextRefs),
					Confidence:   0.82,
					Decision:     firstNonEmpty(executionStrategy, fmt.Sprintf("persisted_context_refs:%d", len(contextRefs))),
					CreatedAt:    occurredAt,
				},
			},
		}); err != nil {
			return err
		}
		return nil
	}
	if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandContextSkipped, cfg.ServiceName, occurredAt, map[string]any{
		"tool_count":         0,
		"resume_queue":       string(resumeQueue),
		"execution_strategy": executionStrategy,
		"execution_role":     executionRole,
		"trace_events": []events.TraceEvent{
			{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     executionStartedService,
				Actor:       ctx.workflow.AssignedBot,
				EventType:   executionStartedType,
				Status:      events.StatusRunning,
				StartedAt:   occurredAt,
				Description: executionStartedDescription,
			},
		},
	}); err != nil {
		return err
	}
	return nil
}

func processControlActionEffect(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, effect transition.EffectExecution) error {
	actionID := strings.TrimSpace(effect.AggregateID)
	if actionID == "" {
		return failClaimedEffect(store, effect, "invoke_action effect missing aggregate id")
	}
	intent, ok := store.GetActionIntent(actionID)
	if !ok {
		return failClaimedEffect(store, effect, fmt.Sprintf("action intent %s not found", actionID))
	}
	if isTerminalActionStatus(intent.Status) {
		if err := maybeAdvanceWorkflowPhaseFromAction(cfg, store, intent); err != nil {
			_ = failClaimedEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedEffect(store, effect, intent.ID)
	}
	ctx, err := loadWorkflowContext(store, workflowLocator{traceID: intent.TraceID})
	if err != nil {
		_ = failClaimedEffect(store, effect, err.Error())
		return err
	}
	if ctx, err = refreshWorkflowContextState(store, ctx); err != nil {
		_ = failClaimedEffect(store, effect, err.Error())
		return err
	}
	if isTerminalTraceStatus(ctx.trace.Summary.Status) && !isTerminalActionStatus(intent.Status) {
		_, err := submitActionCommand(store, intent.ID, transition.CommandActionCancel, cfg.ServiceName, time.Now().UTC(), map[string]any{
			"operation_id":   firstNonEmpty(intent.OperationID, effect.ID),
			"policy_verdict": firstNonEmpty(intent.PolicyVerdict, fmt.Sprintf("trace_%s", ctx.trace.Summary.Status)),
		})
		if err != nil {
			_ = failClaimedEffect(store, effect, err.Error())
			return err
		}
		if intent.Kind == action.KindSlackPost {
			log.Printf("control-plane duplicate_reply_prevented trace=%s action=%s status=%s", ctx.trace.Summary.TraceID, intent.ID, ctx.trace.Summary.Status)
		}
		return completeClaimedEffect(store, effect, intent.ID)
	}

	if _, err := submitActionCommand(store, intent.ID, transition.CommandActionStart, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"operation_id":   firstNonEmpty(intent.OperationID, effect.ID),
		"approval_state": intent.ApprovalState,
		"policy_verdict": intent.PolicyVerdict,
	}); err != nil {
		_ = failClaimedEffect(store, effect, err.Error())
		return err
	}
	intent, _ = store.GetActionIntent(actionID)

	switch intent.Kind {
	case action.KindToolRead:
		if err := executeToolReadActionIntent(store, toolClient, intent); err != nil {
			if isPostgresActionPersistenceError(err) {
				if finalizeErr := finalizeControlActionPersistenceFailure(cfg, store, effect, ctx, intent, err); finalizeErr != nil {
					_ = failClaimedEffect(store, effect, finalizeErr.Error())
					return fmt.Errorf("finalize control action persistence failure: %w (original error: %v)", finalizeErr, err)
				}
			} else if updatedIntent, ok := store.GetActionIntent(actionID); ok {
				_ = maybeAdvanceWorkflowPhaseFromAction(cfg, store, updatedIntent)
			}
			_ = failClaimedEffect(store, effect, err.Error())
			return err
		}
	case action.KindSlackPost:
		if err := executeSlackPostActionIntent(cfg, store, toolClient, ctx, intent); err != nil {
			if isPostgresActionPersistenceError(err) {
				if finalizeErr := finalizeControlActionPersistenceFailure(cfg, store, effect, ctx, intent, err); finalizeErr != nil {
					_ = failClaimedEffect(store, effect, finalizeErr.Error())
					return fmt.Errorf("finalize control action persistence failure: %w (original error: %v)", finalizeErr, err)
				}
			}
			_ = failClaimedEffect(store, effect, err.Error())
			return err
		}
	default:
		err := fmt.Errorf("unsupported control action kind %s", intent.Kind)
		_ = failClaimedEffect(store, effect, err.Error())
		return err
	}
	intent, _ = store.GetActionIntent(actionID)
	if err := maybeAdvanceWorkflowPhaseFromAction(cfg, store, intent); err != nil {
		_ = failClaimedEffect(store, effect, err.Error())
		return err
	}
	return completeClaimedEffect(store, effect, intent.ID)
}

func executeToolReadActionIntent(store storepkg.Store, toolClient *clients.ToolGatewayClient, intent action.Intent) error {
	started := time.Now().UTC()
	requestPayload := cloneAnyMap(intent.RequestPayload)
	delete(requestPayload, "resume_queue")
	result, execErr := toolClient.Execute(intent.TargetRef, requestPayload)
	completed := time.Now().UTC()
	actionStatus := actionStatusFromToolResult(result, execErr)
	summary := toolResultSummary(result, execErr)
	commandKind, err := actionCommandForStatus(actionStatus)
	if err != nil {
		return err
	}
	if _, err := submitActionCommand(store, intent.ID, commandKind, "tool-gateway", completed, map[string]any{
		"operation_id":   intent.OperationID,
		"approval_state": firstNonEmpty(result.ApprovalState, intent.ApprovalState),
		"executor":       "tool-gateway",
		"provider":       firstNonEmpty(result.Provider, providerForToolName(intent.TargetRef)),
		"provider_ref":   result.ProviderRef,
		"error_code":     actionErrorCode(actionStatus),
		"error_message":  actionErrorMessage(result, execErr),
		"started_at":     started,
		"completed_at":   completed,
		"summary":        summary,
		"tool_call_id":   firstNonEmpty(result.ToolCallID, intent.ID),
		"request_payload": firstNonEmptyMap(
			result.Input,
			intent.RequestPayload,
		),
		"raw_artifact_refs": append([]string(nil), result.RawArtifactRefs...),
	}); err != nil {
		return err
	}
	return nil
}

func executeSlackPostActionIntent(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, ctx workflowContext, intent action.Intent) error {
	started := time.Now().UTC()
	draftBody := stringFromMap(intent.RequestPayload, "draft_body")
	finalBody := stringFromMap(intent.RequestPayload, "final_body")
	body := firstNonEmpty(stringFromMap(intent.RequestPayload, "body"), finalBody, draftBody)
	channelID := stringFromMap(intent.RequestPayload, "channel_id")
	threadTS := stringFromMap(intent.RequestPayload, "thread_ts")
	blockedReason := firstNonEmpty(stringFromMap(intent.RequestPayload, "blocked_reason"), blockedReasonFromIntent(intent))

	baseRecord := events.SlackActionRecord{
		ID:             fmt.Sprintf("slack-action-%d", started.UnixNano()),
		TraceID:        ctx.trace.Summary.TraceID,
		WorkflowID:     ctx.trace.Summary.WorkflowID,
		ConversationID: ctx.trace.Summary.ConversationID,
		CaseID:         ctx.trace.Summary.CaseID,
		ChannelID:      channelID,
		ThreadTS:       threadTS,
		IdempotencyKey: intent.IdempotencyKey,
		DraftBody:      firstNonEmpty(draftBody, body),
		FinalBody:      firstNonEmpty(finalBody, body),
		PolicyVerdict:  firstNonEmpty(intent.PolicyVerdict, blockedReason),
		SendStatus:     "draft_only",
		CreatedAt:      started,
	}
	replyEffect, _, err := claimWorkflowEffectByPayload(
		store,
		ctx.workflow.ID,
		transition.EffectPostSlackReply,
		"reply_action_id",
		intent.ID,
		cfg.ServiceName,
		cfg.WorkItemLeaseDuration,
	)
	if err != nil {
		return err
	}

	var (
		result       storepkg.ToolResult
		execErr      error
		actionStatus action.Status
		summary      string
	)
	if blockedReason != "" {
		actionStatus = action.StatusBlocked
		summary = blockedReason
	} else {
		result, execErr = toolClient.Execute("slack.reply", map[string]any{
			"channel_id": channelID,
			"thread_ts":  threadTS,
			"body":       body,
		})
		actionStatus = actionStatusFromToolResult(result, execErr)
		summary = toolResultSummary(result, execErr)
		baseRecord.ArtifactRefs = append([]string(nil), result.RawArtifactRefs...)
		baseRecord.SendStatus = slackSendStatus(actionStatus, result)
	}
	if blockedReason != "" {
		baseRecord.SendStatus = blockedReason
	}

	commandKind, err := actionCommandForStatus(actionStatus)
	if err != nil {
		return err
	}
	completedAt := time.Now().UTC()
	if _, err := submitActionCommand(store, intent.ID, commandKind, cfg.ServiceName, completedAt, map[string]any{
		"operation_id":   intent.OperationID,
		"approval_state": firstNonEmpty(result.ApprovalState, intent.ApprovalState),
		"policy_verdict": firstNonEmpty(intent.PolicyVerdict, blockedReason),
		"executor":       firstNonEmpty(result.Provider, "tool-gateway"),
		"provider":       firstNonEmpty(result.Provider, "slack"),
		"provider_ref":   firstNonEmpty(result.ProviderRef, threadTS),
		"error_code":     actionErrorCode(actionStatus),
		"error_message":  firstNonEmpty(actionErrorMessage(result, execErr), blockedReason),
		"started_at":     started,
		"completed_at":   completedAt,
		"summary":        firstNonEmpty(summary, blockedReason, "Slack reply action completed."),
		"channel_id":     channelID,
		"thread_ts":      threadTS,
		"draft_body":     baseRecord.DraftBody,
		"final_body":     baseRecord.FinalBody,
		"send_status":    baseRecord.SendStatus,
		"artifact_refs":  append([]string(nil), baseRecord.ArtifactRefs...),
	}); err != nil {
		_ = failClaimedEffect(store, replyEffect, err.Error())
		return err
	}
	switch actionStatus {
	case action.StatusSucceeded:
		if err := completeClaimedEffect(store, replyEffect, firstNonEmpty(result.ProviderRef, intent.ID)); err != nil {
			return err
		}
	case action.StatusBlocked, action.StatusFailed:
		if err := failClaimedEffect(store, replyEffect, firstNonEmpty(actionErrorMessage(result, execErr), blockedReason, string(actionStatus))); err != nil {
			return err
		}
	}

	if actionStatus == action.StatusSucceeded && strings.TrimSpace(baseRecord.SendStatus) == "" {
		baseRecord.SendStatus = "posted"
	}
	return nil
}

func maybeAdvanceWorkflowPhaseFromAction(cfg config.Config, store storepkg.Store, intent action.Intent) error {
	if intent.TraceID == "" || !phaseActionsTerminal(store, intent.TraceID, intent.PhaseKey) {
		return nil
	}
	trace, ok := store.GetTrace(intent.TraceID)
	if !ok || isTerminalTraceStatus(trace.Summary.Status) {
		return nil
	}
	ctx, err := loadWorkflowContext(store, workflowLocator{
		traceID:     intent.TraceID,
		workflowID:  trace.Summary.WorkflowID,
		ingestionID: trace.Summary.IngestionID,
	})
	if err != nil {
		return err
	}
	queueName := queueNameFromString(firstNonEmpty(stringFromMap(intent.RequestPayload, "resume_queue"), string(queue.WorkflowQueue)))
	switch intent.PhaseKey {
	case controlPhaseCollectContext:
		if failure, failed := workflowPhaseFailure(store, intent.TraceID, controlPhaseCollectContext); failed {
			return finalizeWorkflowFailureWithDetails(cfg, store, workflowLocator{
				traceID:     ctx.trace.Summary.TraceID,
				workflowID:  ctx.workflow.ID,
				ingestionID: ctx.ingestion.ID,
			}, failure)
		}
		return submitWorkflowContextCompleted(cfg, store, ctx, queueName, time.Now().UTC())
	case controlPhaseReplyPost:
		completedAt := time.Now().UTC()
		_, workflowStatus, workflowError := workflowOutcomeForTrace(store, ctx.trace.Summary.TraceID)
		switch workflowStatus {
		case "needs-human":
			_, err = submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandWorkflowExecutionNeedsHuman, cfg.ServiceName, completedAt, map[string]any{
				"last_error": workflowError,
			})
		default:
			payload := map[string]any{
				"resume_queue": string(queueName),
			}
			replyCommand := workflowReplyPostedCommandFromPayload(intent.RequestPayload)
			_, err = submitWorkflowCommand(store, ctx.workflow.ID, replyCommand, cfg.ServiceName, completedAt, payload)
		}
		if err == nil {
			_, _, err = store.ReconcileWorkflowTrace(ctx.workflow.ID)
		}
		return err
	default:
		return nil
	}
}

func buildRunnerTask(cfg config.Config, store storepkg.Store, role string, trace events.Trace, workflow storepkg.Workflow, ingestion slackpkg.Ingestion, contextSummary string, contextRefs []clients.RunnerContextRef) clients.RunnerTask {
	effectiveHarness := harness.ResolveEffectiveConfig(store, role, cfg.DefaultReasoningVerbosity)
	var caseSummary *clients.RunnerCaseSummary
	if caseRecord, ok := store.GetCase(trace.Summary.CaseID); ok {
		caseSummary = &clients.RunnerCaseSummary{
			CaseID:         caseRecord.ID,
			ConversationID: caseRecord.ConversationID,
			Kind:           caseRecord.Kind,
			Intent:         caseRecord.Intent,
			Title:          caseRecord.Title,
			Summary:        caseRecord.Summary,
			Status:         string(caseRecord.Status),
			AssignedBot:    caseRecord.AssignedBot,
			LatestTraceID:  caseRecord.LatestTraceID,
		}
	}
	recentEntries := recentConversationEntries(store.ListConversationEntries(trace.Summary.ConversationID))
	priorTraceRefs := priorTraceRefsForCase(store.ListTraces(), trace.Summary.CaseID, trace.Summary.TraceID)
	sessionScopeKind, sessionScopeID, parentScopeKind, parentScopeID := workflowSessionScope(trace, workflow)
	liveHints := workflowplan.BuildLiveHints(workflowplan.RuntimeConfig{
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
		Question:       ingestion.Text,
		ChannelID:      ingestion.ChannelID,
		ThreadTS:       ingestion.ThreadTS,
		EntityRefs:     append([]slackpkg.EntityRef(nil), ingestion.EntityRefs...),
	}, time.Now().UTC())
	hintRefs := liveHintContextRefs(liveHints)
	combinedContextRefs := append(append([]clients.RunnerContextRef{}, contextRefs...), hintRefs...)
	combinedContextSummary := joinContextSummary(contextSummary, liveHintSummary(liveHints))
	allowed, _ := replyPolicy(store, workflow.Kind, trace.Summary.ThreadKey, ingestion.ChannelID)
	mcpServers := workflowMCPServers(cfg, allowed, ingestion.ChannelID, ingestion.ThreadTS)
	allowedTools := workflowRunnerAllowedTools(liveHints, hasSlackMCPServer(mcpServers))
	systemMessage := harness.ComposeSystemMessage(
		workflowRunnerSystemMessage(hasSlackMCPServer(mcpServers), hasNotionMCPServer(mcpServers), allowed),
		effectiveHarness,
	)
	prompt := fmt.Sprintf("User request: %s\n\nInvestigate within the governed tool boundary. Start with the attached persisted evidence and context, then expand with tools as needed. Cite concrete evidence when possible. Default any Slack reply to the ingress thread.", ingestion.Text)
	repo := firstNonEmpty(liveHints.Repo, cfg.DefaultRepo)
	return clients.RunnerTask{
		TaskType:                  "workflow",
		Repo:                      repo,
		RepoRef:                   "main",
		Prompt:                    prompt,
		SystemMessage:             systemMessage,
		MCPServers:                mcpServers,
		AllowedTools:              allowedTools,
		AllowedCommands:           []string{},
		TimeoutSeconds:            0,
		ExpectedOutputs:           []string{"visible_reasoning", "final_answer"},
		ArtifactDestination:       fmt.Sprintf("trace:%s", trace.Summary.TraceID),
		ContextSummary:            combinedContextSummary,
		Intent:                    workflow.Intent,
		TraceID:                   trace.Summary.TraceID,
		WorkflowID:                trace.Summary.WorkflowID,
		ConversationID:            trace.Summary.ConversationID,
		CaseID:                    trace.Summary.CaseID,
		ChannelID:                 ingestion.ChannelID,
		ThreadTS:                  ingestion.ThreadTS,
		TriggerEventID:            trace.Summary.TriggerEventID,
		RecentConversationEntries: recentEntries,
		CaseSummary:               caseSummary,
		PriorTraceRefs:            priorTraceRefs,
		RepoAllowlist:             cfg.AllowedTargetRepos,
		ToolAllowlist:             nil,
		ResponseMode:              workflow.ResponseMode,
		ContextRefs:               combinedContextRefs,
		ApprovalMode:              workflow.ApprovalMode,
		ReasoningVerbosity:        effectiveHarness.ReasoningVerbosity,
		RejectedProposalContext:   []clients.RunnerRejectedProposalContext{},
		SessionScopeKind:          sessionScopeKind,
		SessionScopeID:            sessionScopeID,
		ParentSessionScopeKind:    parentScopeKind,
		ParentSessionScopeID:      parentScopeID,
		HarnessProfileID:          effectiveHarness.Profile.ID,
		HarnessOverlayVersion:     effectiveHarness.EffectiveOverlayVersion,
		MemoryBackend:             harness.DefaultMemoryBackend,
		AssistantPeerID:           fmt.Sprintf("rsi:%s:%s", cfg.Environment, role),
		UserPeerID:                workflowUserPeerID(store.ListConversationEntries(trace.Summary.ConversationID), sessionScopeKind, sessionScopeID),
	}
}

func workflowRunnerSystemMessage(useSlackMCP bool, useNotionMCP bool, replyAllowed bool) string {
	if !useSlackMCP {
		parts := []string{
			"Return explicit visible reasoning only. Do not include hidden chain-of-thought.",
			"Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, and outcome_hypotheses.",
			"Start from persisted evidence, then choose governed tools and tool order yourself.",
		}
		if useNotionMCP {
			parts = append(parts, "Use Notion MCP for Notion workspace search and page fetches when relevant.")
		}
		parts = append(parts, "If you intend to reply in Slack, you must emit an explicit proposed action with kind=slack_post; prose alone is not enough.")
		return strings.Join(parts, " ")
	}
	if replyAllowed {
		parts := []string{
			"Return explicit visible reasoning only. Do not include hidden chain-of-thought.",
			"Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, and outcome_hypotheses.",
			"Use Slack MCP for Slack investigation.",
		}
		if useNotionMCP {
			parts = append(parts, "Use Notion MCP for Notion workspace search and page fetches when relevant.")
		}
		parts = append(parts, "Use governed repo, GitHub, knowledge, RSI, and workspace tools for non-Slack evidence.", "If you have a grounded final answer, send exactly one Slack reply to the bound ingress thread using Slack MCP, then return the JSON object.", "Do not emit a slack_post action contract.")
		return strings.Join(parts, " ")
	}
	parts := []string{
		"Return explicit visible reasoning only. Do not include hidden chain-of-thought.",
		"Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, and outcome_hypotheses.",
		"Use Slack MCP for Slack investigation.",
	}
	if useNotionMCP {
		parts = append(parts, "Use Notion MCP for Notion workspace search and page fetches when relevant.")
	}
	parts = append(parts, "Use governed repo, GitHub, knowledge, RSI, and workspace tools for non-Slack evidence.", "Slack posting is blocked by policy for this workflow, so do not send any Slack messages.")
	return strings.Join(parts, " ")
}

func workflowSessionScope(trace events.Trace, workflow storepkg.Workflow) (string, string, string, string) {
	if strings.EqualFold(strings.TrimSpace(workflow.Intent), "incident") || strings.EqualFold(strings.TrimSpace(workflow.Kind), "incident") {
		return "case", trace.Summary.CaseID, "conversation", trace.Summary.ConversationID
	}
	return "conversation", trace.Summary.ConversationID, "", ""
}

func workflowUserPeerID(entries []conversation.Entry, scopeKind string, scopeID string) string {
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if strings.TrimSpace(entry.ActorID) == "" {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(entry.ActorType), "user") || strings.EqualFold(strings.TrimSpace(entry.ActorType), "operator") {
			return fmt.Sprintf("%s:%s", strings.TrimSpace(entry.ActorType), strings.TrimSpace(entry.ActorID))
		}
	}
	return fmt.Sprintf("session:%s:%s", scopeKind, scopeID)
}

func draftSlackPostAction(cfg config.Config, store storepkg.Store, resumeQueue queue.QueueName, ctx workflowContext, output runnerutil.StructuredOutput, replyBody string, allowed bool, policyVerdict string, completionVerdict string, createdAt time.Time) (action.Intent, []events.TraceEvent, []events.ReasoningStep, error) {
	proposed := firstSlackPostAction(output.ProposedActions)
	if strings.TrimSpace(proposed.Kind) == "" {
		return action.Intent{}, nil, nil, errors.New("explicit slack_post action is required")
	}
	channelID := firstNonEmpty(strings.TrimSpace(stringValueFromMap(proposed.RequestPayload, "channel_id")), ctx.ingestion.ChannelID)
	threadTS := firstNonEmpty(strings.TrimSpace(stringValueFromMap(proposed.RequestPayload, "thread_ts")), ctx.ingestion.ThreadTS)
	draftBody := firstNonEmpty(strings.TrimSpace(stringValueFromMap(proposed.RequestPayload, "draft_body")), output.ReplyDraft, replyBody)
	finalBody := firstNonEmpty(strings.TrimSpace(stringValueFromMap(proposed.RequestPayload, "final_body")), strings.TrimSpace(stringValueFromMap(proposed.RequestPayload, "body")), replyBody)
	idempotencyKey := firstNonEmpty(proposed.IdempotencyKey, fmt.Sprintf("%s:%s:%s", ctx.ingestion.ChannelID, ctx.ingestion.ThreadTS, ctx.trace.Summary.TraceID))
	requestPayload := map[string]any{
		"channel_id":             channelID,
		"thread_ts":              threadTS,
		"body":                   firstNonEmpty(finalBody, replyBody),
		"draft_body":             draftBody,
		"final_body":             firstNonEmpty(finalBody, replyBody),
		"policy_verdict":         policyVerdict,
		"resume_queue":           string(resumeQueue),
		"workflow_reply_command": string(transition.WorkflowReplyPostedCommand(completionVerdict)),
	}

	approvalState := "approved"
	resolvedPolicyVerdict := policyVerdict
	reasoning := []events.ReasoningStep{}
	if !allowed {
		approvalState = "policy_blocked"
		requestPayload["blocked_reason"] = policyVerdict
		reasoning = append(reasoning, events.ReasoningStep{
			ID:           fmt.Sprintf("reason-slack-policy-%d", createdAt.UnixNano()),
			TraceID:      ctx.trace.Summary.TraceID,
			WorkflowID:   ctx.trace.Summary.WorkflowID,
			StepType:     "pre_action_decision",
			Summary:      fmt.Sprintf("Slack reply blocked by policy: %s.", policyVerdict),
			EvidenceRefs: normalizeActionEvidenceRefs(proposed.EvidenceRefs, ctx.trace.Summary.TraceID),
			Confidence:   output.Confidence,
			Decision:     "blocked_by_policy",
			CreatedAt:    createdAt,
		})
	} else {
		reasoning = append(reasoning, events.ReasoningStep{
			ID:           fmt.Sprintf("reason-slack-action-%d", createdAt.UnixNano()),
			TraceID:      ctx.trace.Summary.TraceID,
			WorkflowID:   ctx.trace.Summary.WorkflowID,
			StepType:     "pre_action_decision",
			Summary:      firstNonEmpty(proposed.Rationale, "Post the final answer back into Slack."),
			EvidenceRefs: normalizeActionEvidenceRefs(proposed.EvidenceRefs, ctx.trace.Summary.TraceID),
			Confidence:   output.Confidence,
			Decision:     "queued_for_post",
			CreatedAt:    createdAt,
		})
	}

	intent, _, err := ensureActionIntent(store, action.Intent{
		OwnerPlane:     "control",
		ConversationID: ctx.trace.Summary.ConversationID,
		CaseID:         ctx.trace.Summary.CaseID,
		TraceID:        ctx.trace.Summary.TraceID,
		Kind:           action.KindSlackPost,
		PhaseKey:       controlPhaseReplyPost,
		TargetRef:      firstNonEmpty(proposed.TargetRef, channelID),
		RequestPayload: requestPayload,
		IdempotencyKey: idempotencyKey,
		ApprovalMode:   ctx.workflow.ApprovalMode,
		ApprovalState:  approvalState,
		PolicyVerdict:  resolvedPolicyVerdict,
		Status:         action.StatusQueued,
		RequestedBy:    cfg.ServiceName,
		Rationale:      firstNonEmpty(proposed.Rationale, "Post the runner-authored reply into Slack."),
		EvidenceRefs:   normalizeActionEvidenceRefs(proposed.EvidenceRefs, ctx.trace.Summary.TraceID),
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	})
	if err != nil {
		return action.Intent{}, nil, nil, err
	}
	return intent, []events.TraceEvent{
		{
			TraceID:     ctx.trace.Summary.TraceID,
			IngestionID: ctx.trace.Summary.IngestionID,
			WorkflowID:  ctx.trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     cfg.ServiceName,
			Actor:       ctx.workflow.AssignedBot,
			EventType:   "slack.reply.drafted",
			Status:      events.StatusQueued,
			StartedAt:   createdAt,
			Description: "Drafted explicit Slack reply action from runner-authored side-effect contract.",
		},
	}, reasoning, nil
}

func workflowCompletionVerdict(raw map[string]any) string {
	verdict := strings.TrimSpace(stringValue(raw["completion_verdict"]))
	if verdict == "" {
		return "complete"
	}
	return verdict
}

func workflowTerminationReason(raw map[string]any) string {
	return strings.TrimSpace(stringValue(raw["termination_reason"]))
}

func workflowNativeMCPEnabled(raw map[string]any) bool {
	return boolValue(raw["native_mcp_enabled"])
}

func workflowReplyDelivery(raw map[string]any, fallbackChannelID string, fallbackThreadTS string) (events.SlackActionRecord, bool) {
	value, ok := raw["reply_delivery"]
	if !ok || value == nil {
		return events.SlackActionRecord{}, false
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return events.SlackActionRecord{}, false
	}
	var item map[string]any
	if err := json.Unmarshal(payload, &item); err != nil {
		return events.SlackActionRecord{}, false
	}
	body := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(item, "body")),
		strings.TrimSpace(stringValueFromMap(item, "body_excerpt")),
	)
	record := events.SlackActionRecord{
		ID:             firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "tool_call_id")), fmt.Sprintf("runner-reply-%x", sha1.Sum([]byte(body)))),
		ChannelID:      firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "channel_id")), fallbackChannelID),
		ThreadTS:       firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "thread_ts")), fallbackThreadTS),
		IdempotencyKey: firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "body_sha1")), strings.TrimSpace(stringValueFromMap(item, "tool_call_id"))),
		DraftBody:      body,
		FinalBody:      body,
		SendStatus:     firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "send_status")), "posted"),
	}
	return record, true
}

func partialCompletionNoticeForTerminationReason(terminationReason string) string {
	switch strings.TrimSpace(terminationReason) {
	case "iteration_budget_exhausted":
		return partialCompletionNoticeIterationBudget
	case "task_timeout":
		return partialCompletionNoticeTaskTimeout
	case "output_token_budget_exhausted":
		return partialCompletionNoticeOutputBudget
	default:
		return partialCompletionNoticeGeneric
	}
}

func partialCompletionReasoningSummary(terminationReason string) string {
	switch strings.TrimSpace(terminationReason) {
	case "iteration_budget_exhausted":
		return "Runner exhausted its iteration budget and returned a best-effort partial response."
	case "task_timeout":
		return "Runner hit the workflow time limit and returned a best-effort partial response."
	case "output_token_budget_exhausted":
		return "Runner exhausted its response output budget and returned a best-effort partial response."
	default:
		return "Runner stopped early and returned a best-effort partial response."
	}
}

func partialCompletionRunnerDescription(terminationReason string, hasReplyAction bool) string {
	switch strings.TrimSpace(terminationReason) {
	case "iteration_budget_exhausted":
		if hasReplyAction {
			return "Runner exhausted its iteration budget and returned a partial Slack reply."
		}
		return "Runner exhausted its iteration budget and returned a partial completion without a reply side effect."
	case "task_timeout":
		if hasReplyAction {
			return "Runner hit the workflow time limit and returned a partial Slack reply."
		}
		return "Runner hit the workflow time limit and returned a partial completion without a reply side effect."
	case "output_token_budget_exhausted":
		if hasReplyAction {
			return "Runner exhausted its response output budget and returned a partial Slack reply."
		}
		return "Runner exhausted its response output budget and returned a partial completion without a reply side effect."
	default:
		if hasReplyAction {
			return "Runner returned a partial Slack reply."
		}
		return "Runner returned a partial completion without a reply side effect."
	}
}

func standardizePartialWorkflowReply(output runnerutil.StructuredOutput, terminationReason string) runnerutil.StructuredOutput {
	output.ReplyDraft = standardizePartialReplyBody(output.ReplyDraft, terminationReason)
	output.FinalAnswer = standardizePartialReplyBody(output.FinalAnswer, terminationReason)
	for idx := range output.ProposedActions {
		if !strings.EqualFold(strings.TrimSpace(output.ProposedActions[idx].Kind), string(action.KindSlackPost)) {
			continue
		}
		if output.ProposedActions[idx].RequestPayload == nil {
			output.ProposedActions[idx].RequestPayload = map[string]any{}
		}
		for _, key := range []string{"body", "draft_body", "final_body"} {
			value := strings.TrimSpace(stringValueFromMap(output.ProposedActions[idx].RequestPayload, key))
			if value == "" {
				continue
			}
			output.ProposedActions[idx].RequestPayload[key] = standardizePartialReplyBody(value, terminationReason)
		}
	}
	return output
}

func standardizePartialReplyBody(body string, terminationReason string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	notice := partialCompletionNoticeForTerminationReason(terminationReason)
	if strings.HasPrefix(trimmed, notice) {
		return trimmed
	}
	return notice + "\n\n" + trimmed
}

func workflowReplyPostedCommandFromPayload(payload map[string]any) transition.WorkflowCommandKind {
	switch transition.WorkflowCommandKind(strings.TrimSpace(stringFromMap(payload, "workflow_reply_command"))) {
	case transition.CommandReplyPostedPartial:
		return transition.CommandReplyPostedPartial
	default:
		return transition.CommandReplyPosted
	}
}

func recentConversationEntries(items []conversation.Entry) []clients.RunnerConversationEntry {
	if len(items) > 8 {
		items = items[len(items)-8:]
	}
	out := make([]clients.RunnerConversationEntry, 0, len(items))
	for _, item := range items {
		out = append(out, clients.RunnerConversationEntry{
			ID:            item.ID,
			EventID:       item.EventID,
			TraceID:       item.TraceID,
			Source:        string(item.Source),
			SourceEventID: item.SourceEventID,
			EntryType:     item.EntryType,
			ActorID:       item.ActorID,
			ActorType:     item.ActorType,
			Body:          item.Body,
			CreatedAt:     item.CreatedAt,
		})
	}
	return out
}

func priorTraceRefsForCase(items []events.TraceSummary, caseID string, currentTraceID string) []clients.RunnerTraceRef {
	out := make([]clients.RunnerTraceRef, 0)
	for _, item := range items {
		if item.CaseID != caseID || item.TraceID == currentTraceID {
			continue
		}
		out = append(out, clients.RunnerTraceRef{
			TraceID:        item.TraceID,
			Status:         string(item.Status),
			WorkflowKind:   item.WorkflowKind,
			StartedAt:      item.StartedAt,
			TriggerEventID: item.TriggerEventID,
		})
		if len(out) == 6 {
			break
		}
	}
	return out
}

func findWorkflow(items []storepkg.Workflow, workflowID string) (storepkg.Workflow, bool) {
	for _, item := range items {
		if item.ID == workflowID {
			return item, true
		}
	}
	return storepkg.Workflow{}, false
}

func findIngestion(items []slackpkg.Ingestion, ingestionID string) (slackpkg.Ingestion, bool) {
	for _, item := range items {
		if item.ID == ingestionID {
			return item, true
		}
	}
	return slackpkg.Ingestion{}, false
}

func runnerRoleForQueue(name queue.QueueName) string {
	switch name {
	case queue.ProactiveQueue:
		return "proactive"
	default:
		return "prod"
	}
}

func workflowExecutionStrategy(workflow storepkg.Workflow) string {
	if workflowplan.UseReadHeavySlackQnAStrategy(workflow.Intent, workflow.ResponseMode) {
		return "read_heavy_slack_qna"
	}
	return ""
}

func workflowRunnerAllowedTools(hints workflowplan.LiveHintSet, useSlackMCP bool) []string {
	allowed := make([]string, 0, len(hints.PreferredTools))
	for _, toolName := range uniqueStrings(hints.PreferredTools) {
		trimmed := strings.TrimSpace(toolName)
		if trimmed == "" {
			continue
		}
		if useSlackMCP {
			switch trimmed {
			case "slack.history", "slack.search", "slack.reply":
				continue
			}
		}
		allowed = append(allowed, trimmed)
	}
	return uniqueStrings(allowed)
}

func appendMCPServers(groups ...[]clients.RunnerMCPServer) []clients.RunnerMCPServer {
	out := []clients.RunnerMCPServer{}
	for _, group := range groups {
		out = append(out, group...)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func workflowMCPServers(cfg config.Config, replyAllowed bool, channelID string, threadTS string) []clients.RunnerMCPServer {
	return appendMCPServers(
		slackMCPServersForReply(replyAllowed, channelID, threadTS),
		notionMCPServersForRead(cfg),
	)
}

func questionGatherMCPServers(cfg config.Config, channelID string, threadTS string) []clients.RunnerMCPServer {
	return appendMCPServers(
		slackMCPServersForRead(channelID, threadTS),
		notionMCPServersForRead(cfg),
	)
}

func slackMCPServersForRead(channelID string, threadTS string) []clients.RunnerMCPServer {
	if strings.TrimSpace(channelID) == "" {
		return nil
	}
	return []clients.RunnerMCPServer{
		{
			ServerLabel: "slack",
			ServerURL:   defaultSlackMCPServerURL,
			Profile:     "slack_mcp_read",
			Headers: map[string]string{
				"X-RSI-Channel-ID": channelID,
				"X-RSI-Thread-TS":  threadTS,
			},
		},
	}
}

func slackMCPServersForReply(allowed bool, channelID string, threadTS string) []clients.RunnerMCPServer {
	servers := slackMCPServersForRead(channelID, threadTS)
	if len(servers) == 0 {
		return nil
	}
	if allowed {
		servers[0].Profile = "slack_mcp_reply"
	}
	return servers
}

func notionMCPServersForRead(cfg config.Config) []clients.RunnerMCPServer {
	if !cfg.NotionMCPEnabled {
		return nil
	}
	serverURL := strings.TrimSpace(cfg.NotionMCPServerURL)
	if serverURL == "" {
		serverURL = defaultNotionMCPServerURL
	}
	server := clients.RunnerMCPServer{
		ServerLabel:  "notion",
		ServerURL:    serverURL,
		Profile:      "notion_mcp_read",
		AllowedTools: map[string]any{"read_only": true},
	}
	if authEnvVar := strings.TrimSpace(cfg.NotionMCPAuthorizationEnvVar); authEnvVar != "" {
		server.AuthorizationEnvVar = authEnvVar
	}
	return []clients.RunnerMCPServer{server}
}

func hasSlackMCPServer(servers []clients.RunnerMCPServer) bool {
	for _, server := range servers {
		profile := strings.TrimSpace(server.Profile)
		if profile == "slack_mcp_read" || profile == "slack_mcp_reply" {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(server.ServerLabel), "slack") {
			return true
		}
	}
	return false
}

func hasNotionMCPServer(servers []clients.RunnerMCPServer) bool {
	for _, server := range servers {
		if strings.EqualFold(strings.TrimSpace(server.Profile), "notion_mcp_read") {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(server.ServerLabel), "notion") {
			return true
		}
	}
	return false
}

func replyPolicy(store storepkg.Store, workflowKind string, threadKey string, channelID string) (bool, string) {
	for _, item := range store.ListThreadPolicies() {
		if item.ThreadKey != threadKey {
			continue
		}
		switch item.State {
		case policy.ThreadStateMuted:
			return false, "thread_muted"
		case policy.ThreadStateMuteUntilMention:
			return false, "thread_muted_until_mention"
		case policy.ThreadStateClosed:
			return false, "thread_closed"
		case policy.ThreadStateObserveOnly:
			return false, "thread_observe_only"
		case policy.ThreadStateActive:
			if strings.TrimSpace(threadKey) != "" {
				return true, "thread_allowed"
			}
		default:
			if item.Muted {
				return false, "thread_muted"
			}
		}
	}
	if strings.HasPrefix(channelID, "D") {
		return true, "direct_message"
	}
	for _, item := range store.ListChannelPolicies() {
		if item.ChannelID != channelID {
			continue
		}
		if !item.AutoPostAllowed {
			return false, "channel_autopost_disabled"
		}
		for _, allowed := range item.AllowedWorkflowKinds {
			if allowed == workflowKind {
				return true, "allowed"
			}
		}
		return false, "workflow_kind_not_allowed"
	}
	return false, "channel_policy_missing"
}

func evidenceRefsFromContext(contextRefs []clients.RunnerContextRef) []events.EvidenceRef {
	out := make([]events.EvidenceRef, 0, len(contextRefs))
	for _, ref := range contextRefs {
		out = append(out, events.EvidenceRef{
			Kind:    ref.Kind,
			Ref:     firstNonEmpty(ref.Ref, ref.ToolCallID),
			Summary: ref.Summary,
		})
	}
	return out
}

func stringFromMap(item map[string]any, key string) string {
	if item == nil {
		return ""
	}
	value, _ := item[key].(string)
	return value
}

func persistKnowledgeDrafts(store storepkg.Store, trace events.Trace, drafts []runnerutil.KnowledgeDraft, createdAt time.Time) error {
	for idx, item := range drafts {
		normalized, ok := runnerutil.NormalizeKnowledgeDraft(item, knowledge.ScopeCase, trace.Summary.CaseID)
		if !ok {
			continue
		}
		freshUntil := parseTimeOrNil(normalized.FreshUntil)
		if _, err := runnerutil.PersistKnowledgeDraft(store, knowledge.Entry{
			Tier:       knowledge.TierWorking,
			Kind:       knowledge.Kind(normalized.Kind),
			ScopeType:  knowledge.ScopeType(normalized.ScopeType),
			ScopeID:    normalized.ScopeID,
			Title:      normalized.Title,
			Summary:    normalized.Summary,
			Body:       normalized.Body,
			Status:     knowledge.StatusDraft,
			Confidence: normalized.Confidence,
			FreshUntil: freshUntil,
			SourceType: knowledge.SourceAgent,
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}, evidenceLinksFromDraft(normalized), "control-plane", trace.Summary.TraceID, idx, createdAt); err != nil {
			return err
		}
	}
	return nil
}

func evidenceLinksFromDraft(item runnerutil.KnowledgeDraft) []knowledge.EvidenceLink {
	out := make([]knowledge.EvidenceLink, 0, len(item.EvidenceRefs))
	for _, ref := range item.EvidenceRefs {
		out = append(out, knowledge.EvidenceLink{
			EvidenceType:     ref.Kind,
			EvidenceID:       ref.Ref,
			RelevanceSummary: ref.Summary,
			EvidenceRef:      ref,
		})
	}
	return out
}

func outcomeHypothesisReasoning(trace events.Trace, workflow storepkg.Workflow, items []runnerutil.OutcomeHypothesis, createdAt time.Time) []events.ReasoningStep {
	out := make([]events.ReasoningStep, 0, len(items))
	for idx, item := range items {
		out = append(out, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-outcome-hypothesis-%d-%d", createdAt.UnixNano(), idx),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: workflow.ID,
			StepType:   "outcome_hypothesis",
			Summary:    firstNonEmpty(item.SuccessCondition, item.OutcomeType),
			Confidence: 0.7,
			Decision:   firstNonEmpty(item.MeasurementRef, item.ExpectedTimeHorizon),
			CreatedAt:  createdAt,
		})
	}
	return out
}

func firstSlackPostAction(items []runnerutil.ProposedAction) runnerutil.ProposedAction {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Kind), string(action.KindSlackPost)) {
			return item
		}
	}
	return runnerutil.ProposedAction{}
}

func normalizeActionEvidenceRefs(items []events.EvidenceRef, traceID string) []events.EvidenceRef {
	if len(items) == 0 {
		return []events.EvidenceRef{{Kind: "trace", Ref: traceID, Summary: traceID}}
	}
	return items
}

func actionStatusFromToolResult(result storepkg.ToolResult, execErr error) action.Status {
	if execErr != nil {
		return action.StatusFailed
	}
	status := strings.ToLower(strings.TrimSpace(result.Status))
	switch status {
	case "blocked":
		return action.StatusBlocked
	case "failed", "error":
		return action.StatusFailed
	case "", "ok", "success", "completed":
		if !result.Available && strings.TrimSpace(result.Status) != "" {
			return action.StatusBlocked
		}
		return action.StatusSucceeded
	default:
		if !result.Available {
			return action.StatusBlocked
		}
		return action.StatusSucceeded
	}
}

func actionErrorCode(status action.Status) string {
	switch status {
	case action.StatusBlocked:
		return "blocked"
	case action.StatusFailed:
		return "failed"
	default:
		return ""
	}
}

func actionErrorMessage(result storepkg.ToolResult, execErr error) string {
	if execErr != nil {
		return execErr.Error()
	}
	status := strings.ToLower(strings.TrimSpace(result.Status))
	if status == "blocked" || status == "failed" || status == "error" {
		return result.Summary
	}
	if !result.Available && status != "" {
		return result.Summary
	}
	return ""
}

func slackSendStatus(status action.Status, result storepkg.ToolResult) string {
	switch status {
	case action.StatusSucceeded:
		return statusString(result.Output["posted"], "posted")
	case action.StatusBlocked:
		if !result.Available {
			return "provider_unavailable"
		}
		return "blocked"
	default:
		return "failed"
	}
}

func providerForToolName(toolName string) string {
	switch {
	case strings.HasPrefix(toolName, "slack."):
		return "slack"
	case strings.HasPrefix(toolName, "github."):
		return "github"
	case strings.HasPrefix(toolName, "sentry."):
		return "sentry"
	case strings.HasPrefix(toolName, "kubernetes."):
		return "kubernetes"
	case strings.HasPrefix(toolName, "cloudflare."):
		return "cloudflare"
	default:
		return "internal"
	}
}

func parseTimeOrNil(raw string) *time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
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

func statusString(value any, fallback string) string {
	if posted, ok := value.(bool); ok {
		if posted {
			return "posted"
		}
		return fallback
	}
	return fallback
}

func loadWorkflowContext(store storepkg.Store, workflow workflowLocator) (workflowContext, error) {
	trace, ok := store.GetTrace(workflow.traceID)
	if !ok {
		return workflowContext{}, fmt.Errorf("trace %s not found", workflow.traceID)
	}
	workflowID := firstNonEmpty(workflow.workflowID, trace.Summary.WorkflowID)
	record, ok := findWorkflow(store.ListWorkflows(), workflowID)
	if !ok {
		return workflowContext{}, fmt.Errorf("workflow %s not found", workflowID)
	}
	ingestionID := firstNonEmpty(workflow.ingestionID, trace.Summary.IngestionID)
	ingestion, ok := findIngestion(store.ListIngestions(), ingestionID)
	if !ok {
		return workflowContext{}, fmt.Errorf("ingestion %s not found", ingestionID)
	}
	return workflowContext{trace: trace, workflow: record, ingestion: ingestion}, nil
}

func ensureActionIntent(store storepkg.Store, template action.Intent) (action.Intent, bool, error) {
	if existing, ok := findActionIntentByIdempotencyKey(store.ListActionIntents(), template.IdempotencyKey); ok {
		return existing, false, nil
	}
	if strings.TrimSpace(template.IdempotencyKey) == "" {
		return action.Intent{}, false, errors.New("action intent idempotency key is required")
	}
	template.ID = actionIntentIDFromIdempotencyKey(template.IdempotencyKey)
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: template.ID,
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   actionCommandID(template.ID, transition.CommandActionQueue, ""),
		Actor:       firstNonEmpty(template.RequestedBy, "control-plane"),
		OccurredAt:  firstNonZeroTime(template.CreatedAt, time.Now().UTC()),
		Payload: map[string]any{
			"owner_plane":     template.OwnerPlane,
			"conversation_id": template.ConversationID,
			"case_id":         template.CaseID,
			"trace_id":        template.TraceID,
			"attempt_id":      template.AttemptID,
			"kind":            string(template.Kind),
			"phase_key":       template.PhaseKey,
			"target_ref":      template.TargetRef,
			"request_payload": cloneAnyMap(template.RequestPayload),
			"idempotency_key": template.IdempotencyKey,
			"approval_mode":   template.ApprovalMode,
			"approval_state":  template.ApprovalState,
			"policy_verdict":  template.PolicyVerdict,
			"requested_by":    template.RequestedBy,
			"rationale":       template.Rationale,
			"evidence_refs":   normalizeActionEvidenceRefs(template.EvidenceRefs, template.TraceID),
		},
	})
	if err != nil {
		return action.Intent{}, false, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return action.Intent{}, false, errors.New(receipt.Reason)
	}
	created, ok := store.GetActionIntent(template.ID)
	if !ok {
		return action.Intent{}, false, fmt.Errorf("action intent %s not found after queue command", template.ID)
	}
	return created, receipt.DecisionKind == transition.DecisionAdvance, nil
}

func findActionIntentByIdempotencyKey(items []action.Intent, key string) (action.Intent, bool) {
	key = strings.TrimSpace(key)
	if key == "" {
		return action.Intent{}, false
	}
	for _, item := range items {
		if strings.TrimSpace(item.IdempotencyKey) == key {
			return item, true
		}
	}
	return action.Intent{}, false
}

func actionIntentIDFromIdempotencyKey(key string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(key)))
	return fmt.Sprintf("action-%x", sum[:8])
}

func firstNonZeroTime(items ...time.Time) time.Time {
	for _, item := range items {
		if !item.IsZero() {
			return item
		}
	}
	return time.Time{}
}

func loadWorkflowContextForEffect(store storepkg.Store, effect transition.EffectExecution) (workflowContext, queue.QueueName, error) {
	workflow, err := workflowLocatorForEffect(store, effect)
	if err != nil {
		return workflowContext{}, "", err
	}
	ctx, err := loadWorkflowContext(store, workflow)
	if err != nil {
		return workflowContext{}, "", err
	}
	queueName := queueNameFromString(firstNonEmpty(stringFromMap(effect.Payload, "resume_queue"), string(queue.WorkflowQueue)))
	return ctx, queueName, nil
}

func workflowLocatorForEffect(store storepkg.Store, effect transition.EffectExecution) (workflowLocator, error) {
	workflowID := strings.TrimSpace(effect.AggregateID)
	if workflowID == "" {
		return workflowLocator{}, fmt.Errorf("workflow effect %s missing aggregate id", effect.ID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowID)
	if !ok {
		return workflowLocator{}, fmt.Errorf("workflow %s not found", workflowID)
	}
	return workflowLocator{
		traceID:     workflow.TraceID,
		workflowID:  workflow.ID,
		ingestionID: workflow.IngestionID,
	}, nil
}

func toolInputForIntent(cfg config.Config, trace events.TraceSummary, workflow storepkg.Workflow, ingestion slackpkg.Ingestion) map[string]any {
	return workflowplan.BuildToolRequestPayload(workflowplan.RuntimeConfig{
		DefaultRepo:      cfg.DefaultRepo,
		AllowedRepos:     append([]string(nil), cfg.AllowedTargetRepos...),
		KnowledgeBaseURL: cfg.DefaultKnowledgeBaseURL,
		SandboxNamespace: cfg.SandboxNamespace,
	}, workflowplan.RequestContext{
		Trace:          trace,
		WorkflowID:     workflow.ID,
		ConversationID: workflow.ConversationID,
		CaseID:         workflow.CaseID,
		WorkflowKind:   workflow.Kind,
		AssignedBot:    workflow.AssignedBot,
		Question:       ingestion.Text,
		ChannelID:      ingestion.ChannelID,
		ThreadTS:       ingestion.ThreadTS,
		EntityRefs:     append([]slackpkg.EntityRef(nil), ingestion.EntityRefs...),
	}, time.Now().UTC())
}

func contextFromTrace(trace events.Trace) (string, []clients.RunnerContextRef) {
	contextRefs := make([]clients.RunnerContextRef, 0, len(trace.ToolCalls)+len(trace.Reasoning)+len(trace.Events)+len(trace.SlackActions))
	summaries := make([]string, 0, len(trace.ToolCalls)+len(trace.Reasoning)+len(trace.Events)+len(trace.SlackActions))
	for _, call := range tailToolCalls(trace.ToolCalls, 8) {
		contextRefs = append(contextRefs, clients.RunnerContextRef{
			Kind:       "tool_call",
			Ref:        firstNonEmpty(call.ToolCallID, call.ID),
			ToolCallID: firstNonEmpty(call.ToolCallID, call.ID),
			Summary:    call.Summary,
			ToolName:   call.ToolName,
			Status:     string(call.Status),
		})
		summaries = append(summaries, call.Summary)
	}
	for _, step := range tailReasoning(trace.Reasoning, 6) {
		contextRefs = append(contextRefs, clients.RunnerContextRef{
			Kind:       "reasoning_step",
			Ref:        step.ID,
			Summary:    step.Summary,
			StepType:   step.StepType,
			Decision:   step.Decision,
			Confidence: step.Confidence,
			TraceID:    step.TraceID,
		})
		summaries = append(summaries, step.Summary)
	}
	for _, event := range tailTraceEvents(trace.Events, 6) {
		contextRefs = append(contextRefs, clients.RunnerContextRef{
			Kind:        "trace_event",
			Ref:         event.EventType,
			Summary:     event.Description,
			Status:      string(event.Status),
			Plane:       event.Plane,
			Service:     event.Service,
			Description: event.Description,
			TraceID:     event.TraceID,
		})
		summaries = append(summaries, event.Description)
	}
	for _, slackAction := range tailSlackActions(trace.SlackActions, 4) {
		contextRefs = append(contextRefs, clients.RunnerContextRef{
			Kind:      "slack_action",
			Ref:       firstNonEmpty(slackAction.IdempotencyKey, slackAction.ID),
			Summary:   firstNonEmpty(slackAction.FinalBody, slackAction.DraftBody),
			Status:    slackAction.SendStatus,
			ChannelID: slackAction.ChannelID,
			ThreadTS:  slackAction.ThreadTS,
			TraceID:   slackAction.TraceID,
		})
		summaries = append(summaries, firstNonEmpty(slackAction.FinalBody, slackAction.DraftBody))
	}
	return strings.Join(uniqueStrings(summaries), " "), contextRefs
}

func workflowOutcomeForTrace(store storepkg.Store, traceID string) (events.Status, string, string) {
	lastError := ""
	needsHuman := false
	for _, intent := range store.ListActionIntents() {
		if intent.TraceID != traceID || intent.OwnerPlane != "control" {
			continue
		}
		switch intent.Status {
		case action.StatusBlocked, action.StatusFailed, action.StatusCanceled:
			needsHuman = true
			if lastError == "" {
				lastError = firstNonEmpty(intent.PolicyVerdict, latestActionError(store, intent.ID), string(intent.Status))
			}
		}
	}
	if needsHuman {
		return events.StatusNeedsHuman, "needs-human", lastError
	}
	return events.StatusCompleted, "completed", ""
}

func submitWorkflowCommand(store storepkg.Store, workflowID string, kind transition.WorkflowCommandKind, actor string, occurredAt time.Time, payload map[string]any) (transition.CommandReceipt, error) {
	if strings.TrimSpace(workflowID) == "" {
		return transition.CommandReceipt{}, nil
	}
	return store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflowID,
		CommandKind: string(kind),
		CommandID:   workflowCommandID(workflowID, kind),
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
}

func workflowCommandID(workflowID string, kind transition.WorkflowCommandKind) string {
	return fmt.Sprintf("cmd-workflow:%s:%s", strings.TrimSpace(workflowID), string(kind))
}

func submitActionCommand(store storepkg.Store, actionID string, kind transition.ActionExecutionCommandKind, actor string, occurredAt time.Time, payload map[string]any) (transition.CommandReceipt, error) {
	if strings.TrimSpace(actionID) == "" {
		return transition.CommandReceipt{}, nil
	}
	return store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: actionID,
		CommandKind: string(kind),
		CommandID:   actionCommandID(actionID, kind, stringFromMap(payload, "operation_id")),
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
}

func actionCommandID(actionID string, kind transition.ActionExecutionCommandKind, operationID string) string {
	base := fmt.Sprintf("cmd-action:%s:%s", strings.TrimSpace(actionID), string(kind))
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return base
	}
	return base + ":" + operationID
}

func actionCommandForStatus(status action.Status) (transition.ActionExecutionCommandKind, error) {
	switch status {
	case action.StatusSucceeded:
		return transition.CommandActionSucceed, nil
	case action.StatusBlocked:
		return transition.CommandActionBlock, nil
	case action.StatusFailed:
		return transition.CommandActionFail, nil
	default:
		return "", fmt.Errorf("unsupported terminal action status %s", status)
	}
}

func claimLatestWorkflowEffect(store storepkg.Store, workflowID string, kind transition.EffectKind, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	effect, ok := workflowEffectByPayload(store, workflowID, kind, "", "")
	if !ok {
		return transition.EffectExecution{}, false, nil
	}
	return store.ClaimEffectExecution(effect.ID, holder, lease)
}

func claimNextExecutionEffect(store storepkg.Store, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	for _, effect := range store.ListEffectExecutions() {
		switch {
		case effect.MachineKind == transition.MachineWorkflow && effect.EffectKind == transition.EffectInvokeRunner:
		case effect.MachineKind == transition.MachineQuestionRun:
		default:
			continue
		}
		claimed, ok, err := store.ClaimEffectExecution(effect.ID, holder, lease)
		if err != nil {
			return transition.EffectExecution{}, false, err
		}
		if ok {
			return claimed, true, nil
		}
	}
	return transition.EffectExecution{}, false, nil
}

func claimNextWorkflowRunnerEffect(store storepkg.Store, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	return claimNextExecutionEffect(store, holder, lease)
}

func claimNextActionEffect(store storepkg.Store, ownerPlane string, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	ownerPlane = strings.TrimSpace(ownerPlane)
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAction || effect.EffectKind != transition.EffectInvokeAction {
			continue
		}
		if ownerPlane != "" && !strings.EqualFold(strings.TrimSpace(stringFromMap(effect.Payload, "owner_plane")), ownerPlane) {
			continue
		}
		claimed, ok, err := store.ClaimEffectExecution(effect.ID, holder, lease)
		if err != nil {
			return transition.EffectExecution{}, false, err
		}
		if ok {
			return claimed, true, nil
		}
	}
	return transition.EffectExecution{}, false, nil
}

func claimWorkflowEffectByPayload(store storepkg.Store, workflowID string, kind transition.EffectKind, payloadKey string, payloadValue string, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	effect, ok := workflowEffectByPayload(store, workflowID, kind, payloadKey, payloadValue)
	if !ok {
		return transition.EffectExecution{}, false, nil
	}
	return store.ClaimEffectExecution(effect.ID, holder, lease)
}

func completeClaimedEffect(store storepkg.Store, effect transition.EffectExecution, resultRef string) error {
	if strings.TrimSpace(effect.ID) == "" || effect.Status == transition.EffectCompleted {
		return nil
	}
	item, err := store.CompleteEffectExecution(effect.ID, effect.Holder, resultRef)
	if err != nil {
		return err
	}
	if item.Status != transition.EffectCompleted {
		return fmt.Errorf("effect %s completion was not applied; current status=%s holder=%s", effect.ID, item.Status, item.Holder)
	}
	return nil
}

func failClaimedEffect(store storepkg.Store, effect transition.EffectExecution, lastError string) error {
	if strings.TrimSpace(effect.ID) == "" || effect.Status == transition.EffectFailed {
		return nil
	}
	item, err := store.FailEffectExecution(effect.ID, effect.Holder, lastError)
	if err != nil {
		return err
	}
	if item.Status != transition.EffectFailed {
		return fmt.Errorf("effect %s failure was not applied; current status=%s holder=%s", effect.ID, item.Status, item.Holder)
	}
	return nil
}

func latestWorkflowEffect(store storepkg.Store, workflowID string, kind transition.EffectKind) (transition.EffectExecution, bool) {
	return workflowEffectByPayload(store, workflowID, kind, "", "")
}

func workflowEffectByPayload(store storepkg.Store, workflowID string, kind transition.EffectKind, payloadKey string, payloadValue string) (transition.EffectExecution, bool) {
	payloadKey = strings.TrimSpace(payloadKey)
	payloadValue = strings.TrimSpace(payloadValue)
	for _, effect := range store.ListEffectExecutionsByAggregate(transition.MachineWorkflow, workflowID) {
		if effect.EffectKind != kind {
			continue
		}
		if payloadKey != "" && stringFromMap(effect.Payload, payloadKey) != payloadValue {
			continue
		}
		return effect, true
	}
	return transition.EffectExecution{}, false
}

func latestActionError(store storepkg.Store, actionID string) string {
	results := store.ListActionResults(actionID)
	if len(results) == 0 {
		return ""
	}
	last := results[len(results)-1]
	return firstNonEmpty(last.ErrorMessage, last.ErrorCode)
}

func blockedReasonFromIntent(intent action.Intent) string {
	switch intent.ApprovalState {
	case "policy_blocked":
		return firstNonEmpty(intent.PolicyVerdict, "policy_blocked")
	case "missing_explicit_action":
		return "missing_explicit_action"
	default:
		return ""
	}
}

type actionPersistenceFailure struct {
	Subsystem   string
	FailureMode string
	Provider    string
	SQLState    string
	Constraint  string
	Table       string
}

func finalizeControlActionPersistenceFailure(cfg config.Config, store storepkg.Store, effect transition.EffectExecution, ctx workflowContext, intent action.Intent, execErr error) error {
	failure := classifyActionPersistenceFailure(intent, execErr)
	now := time.Now().UTC()
	if _, err := submitActionCommand(store, intent.ID, transition.CommandActionFail, cfg.ServiceName, now, map[string]any{
		"operation_id":   firstNonEmpty(intent.OperationID, effect.ID),
		"policy_verdict": firstNonEmpty(failure.FailureMode, "action_result_persistence_failure"),
		"error_code":     "failed",
		"error_message":  execErr.Error(),
		"record_result":  false,
	}); err != nil {
		return err
	}

	description := actionPersistenceFailureDescription(intent, effect, failure, execErr)
	if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandWorkflowExecutionNeedsHuman, cfg.ServiceName, now, map[string]any{
		"last_error": description,
		"trace_events": []events.TraceEvent{
			{
				TraceID:        ctx.trace.Summary.TraceID,
				IngestionID:    ctx.trace.Summary.IngestionID,
				WorkflowID:     ctx.trace.Summary.WorkflowID,
				ConversationID: ctx.trace.Summary.ConversationID,
				CaseID:         ctx.trace.Summary.CaseID,
				TriggerEventID: ctx.trace.Summary.TriggerEventID,
				Plane:          "control",
				Service:        cfg.ServiceName,
				Actor:          "action-worker",
				EventType:      "action.persistence_failed",
				Status:         events.StatusNeedsHuman,
				StartedAt:      now,
				Description:    description,
			},
		},
	}); err != nil {
		return err
	}

	return nil
}

func isPostgresActionPersistenceError(err error) bool {
	_, ok := postgresFailureDetailsFromError(err)
	return ok
}

func classifyActionPersistenceFailure(intent action.Intent, err error) actionPersistenceFailure {
	failure := actionPersistenceFailure{
		Subsystem:   "control-plane",
		FailureMode: "action_result_persistence_failure",
		Provider:    providerForToolName(intent.TargetRef),
	}
	if intent.Kind == action.KindSlackPost {
		failure.Provider = "slack"
	}
	if details, ok := postgresFailureDetailsFromError(err); ok {
		failure.Subsystem = "shared-store"
		failure.SQLState = details.SQLState
		failure.Constraint = details.Constraint
		failure.Table = details.Table
		failure.FailureMode = "postgres_persistence_failure"
		if isActionResultPrimaryKeyCollision(details, err) {
			failure.FailureMode = "action_result_primary_key_collision"
		} else if details.SQLState == "23505" {
			failure.FailureMode = "postgres_unique_constraint_violation"
		}
	}
	return failure
}

func actionPersistenceFailureDescription(intent action.Intent, effect transition.EffectExecution, failure actionPersistenceFailure, execErr error) string {
	parts := []string{
		fmt.Sprintf("subsystem=%s", failure.Subsystem),
		fmt.Sprintf("failure_mode=%s", failure.FailureMode),
		fmt.Sprintf("provider=%s", firstNonEmpty(failure.Provider, "unknown")),
		fmt.Sprintf("action_intent_id=%s", intent.ID),
		fmt.Sprintf("effect_execution_id=%s", effect.ID),
		fmt.Sprintf("kind=%s", intent.Kind),
	}
	if failure.SQLState != "" {
		parts = append(parts, fmt.Sprintf("sqlstate=%s", failure.SQLState))
	}
	if failure.Constraint != "" {
		parts = append(parts, fmt.Sprintf("constraint=%s", failure.Constraint))
	}
	if failure.Table != "" {
		parts = append(parts, fmt.Sprintf("table=%s", failure.Table))
	}
	parts = append(parts, fmt.Sprintf("error=%q", execErr.Error()))
	return strings.Join(parts, " ")
}

type postgresFailureDetails struct {
	SQLState   string
	Constraint string
	Table      string
}

func postgresFailureDetailsFromError(err error) (postgresFailureDetails, bool) {
	var pgErr *pgconn.PgError
	details := postgresFailureDetails{}
	if errors.As(err, &pgErr) && pgErr != nil {
		details.SQLState = strings.TrimSpace(pgErr.Code)
		details.Constraint = strings.TrimSpace(pgErr.ConstraintName)
		details.Table = strings.TrimSpace(pgErr.TableName)
	}
	msg := err.Error()
	if details.SQLState == "" {
		if idx := strings.Index(msg, "SQLSTATE "); idx >= 0 {
			code := strings.TrimSpace(msg[idx+len("SQLSTATE "):])
			if end := strings.IndexAny(code, " )]"); end >= 0 {
				code = code[:end]
			}
			details.SQLState = strings.TrimSpace(code)
		}
	}
	lower := strings.ToLower(msg)
	if details.Constraint == "" && strings.Contains(lower, "action_result_pkey") {
		details.Constraint = "action_result_pkey"
	}
	if details.Table == "" && strings.Contains(lower, "action_result") {
		details.Table = "action_result"
	}
	return details, details.SQLState != "" || details.Constraint != "" || details.Table != ""
}

func isActionResultPrimaryKeyCollision(details postgresFailureDetails, err error) bool {
	if strings.EqualFold(strings.TrimSpace(details.Constraint), "action_result_pkey") {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "action_result_pkey") && strings.Contains(lower, "duplicate key")
}

func toolResultSummary(result storepkg.ToolResult, execErr error) string {
	if execErr != nil {
		return execErr.Error()
	}
	return firstNonEmpty(result.Summary, result.Status, "tool action completed")
}

func firstNonEmptyMap(primary map[string]any, fallback map[string]any) map[string]any {
	if len(primary) > 0 {
		return primary
	}
	return fallback
}

func cloneAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func phaseActionsTerminal(store storepkg.Store, traceID string, phaseKey string) bool {
	if strings.TrimSpace(phaseKey) == "" {
		return false
	}
	found := false
	for _, intent := range store.ListActionIntents() {
		if intent.TraceID != traceID || intent.OwnerPlane != "control" || intent.PhaseKey != phaseKey {
			continue
		}
		found = true
		if !isTerminalActionStatus(intent.Status) {
			return false
		}
	}
	return found
}

func isTerminalActionStatus(status action.Status) bool {
	switch status {
	case action.StatusSucceeded, action.StatusBlocked, action.StatusFailed, action.StatusCanceled, action.StatusSuperseded:
		return true
	default:
		return false
	}
}

func isTerminalWorkflowStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case string(transition.WorkflowStateCompleted), string(transition.WorkflowStateFailed), string(transition.WorkflowStateNeedsHuman), string(transition.WorkflowStateSuperseded):
		return true
	default:
		return false
	}
}

func isTerminalTraceStatus(status events.Status) bool {
	switch status {
	case events.StatusCompleted, events.StatusFailed, events.StatusNeedsHuman, events.StatusSuppressed:
		return true
	default:
		return false
	}
}

func refreshWorkflowContextState(store storepkg.Store, ctx workflowContext) (workflowContext, error) {
	refreshed, err := loadWorkflowContext(store, workflowLocator{
		traceID:     ctx.trace.Summary.TraceID,
		workflowID:  ctx.workflow.ID,
		ingestionID: ctx.ingestion.ID,
	})
	if err != nil {
		return workflowContext{}, err
	}
	if workflowTraceStatusMismatch(refreshed.workflow.Status, refreshed.trace.Summary.Status) {
		if _, _, err := store.ReconcileWorkflowTrace(refreshed.workflow.ID); err != nil {
			return workflowContext{}, err
		}
		return loadWorkflowContext(store, workflowLocator{
			traceID:     refreshed.trace.Summary.TraceID,
			workflowID:  refreshed.workflow.ID,
			ingestionID: refreshed.ingestion.ID,
		})
	}
	return refreshed, nil
}

func workflowTraceStatusMismatch(workflowStatus string, traceStatus events.Status) bool {
	switch strings.TrimSpace(workflowStatus) {
	case string(transition.WorkflowStateFailed):
		return traceStatus != events.StatusFailed
	case string(transition.WorkflowStateCompleted):
		return traceStatus != events.StatusCompleted
	case string(transition.WorkflowStateNeedsHuman):
		return traceStatus != events.StatusNeedsHuman
	default:
		return false
	}
}

func liveHintContextRefs(hints workflowplan.LiveHintSet) []clients.RunnerContextRef {
	refs := make([]clients.RunnerContextRef, 0, 2+len(hints.PreferredTools)+len(hints.CandidateReadSurfaces))
	if repo := strings.TrimSpace(hints.Repo); repo != "" {
		refs = append(refs, clients.RunnerContextRef{
			Kind:    "repo_target",
			Ref:     repo,
			Repo:    repo,
			Summary: fmt.Sprintf("Target repository for live investigation: %s.", repo),
		})
	}
	if hints.Since != "" && hints.Until != "" {
		refs = append(refs, clients.RunnerContextRef{
			Kind:    "repo_activity_window",
			Ref:     fmt.Sprintf("%s..%s", hints.Since, hints.Until),
			Summary: fmt.Sprintf("Suggested repository activity window from %s to %s.", hints.Since, hints.Until),
			Since:   hints.Since,
			Until:   hints.Until,
		})
	}
	for _, toolName := range uniqueStrings(hints.PreferredTools) {
		refs = append(refs, clients.RunnerContextRef{
			Kind:     "preferred_tool_hint",
			Ref:      toolName,
			ToolName: toolName,
			Summary:  fmt.Sprintf("Suggested starting tool: %s.", toolName),
		})
	}
	for _, surface := range hints.CandidateReadSurfaces {
		refs = append(refs, clients.RunnerContextRef{
			Kind:      "candidate_read_surface",
			Ref:       firstNonEmpty(surface.Ref, fmt.Sprintf("%s:%s", surface.ChannelID, surface.ThreadTS)),
			Summary:   slackSurfaceHintSummary(surface),
			Source:    surface.Source,
			ChannelID: surface.ChannelID,
			ThreadTS:  surface.ThreadTS,
		})
	}
	return refs
}

func liveHintSummary(hints workflowplan.LiveHintSet) string {
	parts := make([]string, 0, 4)
	if repo := strings.TrimSpace(hints.Repo); repo != "" {
		parts = append(parts, fmt.Sprintf("Target repo: %s.", repo))
	}
	if len(hints.CandidateReadSurfaces) > 0 {
		parts = append(parts, fmt.Sprintf("Candidate Slack read surfaces: %d.", len(hints.CandidateReadSurfaces)))
	}
	if len(hints.PreferredTools) > 0 {
		parts = append(parts, fmt.Sprintf("Preferred starting tools: %s.", strings.Join(uniqueStrings(hints.PreferredTools), ", ")))
	}
	if hints.Since != "" && hints.Until != "" {
		parts = append(parts, fmt.Sprintf("Suggested repo-activity window: %s to %s.", hints.Since, hints.Until))
	}
	return strings.Join(parts, " ")
}

func slackSurfaceHintSummary(surface workflowplan.SlackSurfaceHint) string {
	switch {
	case strings.TrimSpace(surface.ChannelID) != "" && strings.TrimSpace(surface.ThreadTS) != "":
		return fmt.Sprintf("Candidate Slack surface from %s: channel %s thread %s.", firstNonEmpty(surface.Source, "hint"), surface.ChannelID, surface.ThreadTS)
	case strings.TrimSpace(surface.ChannelID) != "":
		return fmt.Sprintf("Candidate Slack surface from %s: channel %s.", firstNonEmpty(surface.Source, "hint"), surface.ChannelID)
	case strings.TrimSpace(surface.Ref) != "":
		return fmt.Sprintf("Candidate Slack surface from %s: %s.", firstNonEmpty(surface.Source, "hint"), surface.Ref)
	default:
		return "Candidate Slack surface."
	}
}

func joinContextSummary(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, "\n\n")
}

func mergeWorkflowRunnerDiagnostics(base map[string]any, raw map[string]any) map[string]any {
	diagnostics := cloneStringAnyMap(base)
	if diagnostics == nil {
		diagnostics = map[string]any{}
	}
	for _, key := range []string{
		"completion_verdict",
		"termination_reason",
		"max_iterations_reached",
		"candidate_read_surfaces",
		"selected_context_surfaces",
		"memory_warnings",
		"action_contract_repair_attempted",
		"action_contract_repair_succeeded",
		"action_contract_repair_error",
		"action_contract_repair_response",
	} {
		if value, ok := raw[key]; ok && value != nil {
			diagnostics[key] = value
		}
	}
	return diagnostics
}

func toolCallRecordsFromRunnerRaw(raw map[string]any) []events.ToolCallRecord {
	value, ok := raw["tool_calls"]
	if !ok || value == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var out []events.ToolCallRecord
	if err := json.Unmarshal(payload, &out); err == nil {
		return out
	}
	var single events.ToolCallRecord
	if err := json.Unmarshal(payload, &single); err == nil && strings.TrimSpace(single.ToolCallID) != "" {
		return []events.ToolCallRecord{single}
	}
	return nil
}

func bindRunnerToolCallRecords(records []events.ToolCallRecord, trace events.Trace, workflow storepkg.Workflow) []events.ToolCallRecord {
	if len(records) == 0 {
		return nil
	}
	bound := make([]events.ToolCallRecord, 0, len(records))
	for idx, record := range records {
		item := record
		if strings.TrimSpace(item.ID) == "" {
			key := firstNonEmpty(strings.TrimSpace(item.ToolCallID), strings.TrimSpace(item.ToolName), fmt.Sprintf("%d", idx+1))
			item.ID = fmt.Sprintf("runner-tool-record-%s", key)
		}
		if strings.TrimSpace(item.TraceID) == "" {
			item.TraceID = trace.Summary.TraceID
		}
		if strings.TrimSpace(item.WorkflowID) == "" {
			item.WorkflowID = workflow.ID
		}
		if strings.TrimSpace(item.ConversationID) == "" {
			item.ConversationID = trace.Summary.ConversationID
		}
		if strings.TrimSpace(item.CaseID) == "" {
			item.CaseID = trace.Summary.CaseID
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now().UTC()
		}
		bound = append(bound, item)
	}
	return bound
}

func queueNameFromString(raw string) queue.QueueName {
	switch strings.TrimSpace(raw) {
	case string(queue.ProactiveQueue):
		return queue.ProactiveQueue
	case string(queue.WorkflowQueue):
		return queue.WorkflowQueue
	default:
		return queue.WorkflowQueue
	}
}

func tailToolCalls(items []events.ToolCallRecord, limit int) []events.ToolCallRecord {
	if len(items) <= limit {
		return items
	}
	return items[len(items)-limit:]
}

func tailReasoning(items []events.ReasoningStep, limit int) []events.ReasoningStep {
	if len(items) <= limit {
		return items
	}
	return items[len(items)-limit:]
}

func tailTraceEvents(items []events.TraceEvent, limit int) []events.TraceEvent {
	if len(items) <= limit {
		return items
	}
	return items[len(items)-limit:]
}

func tailSlackActions(items []events.SlackActionRecord, limit int) []events.SlackActionRecord {
	if len(items) <= limit {
		return items
	}
	return items[len(items)-limit:]
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
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
