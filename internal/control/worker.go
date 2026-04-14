package control

import (
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
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	controlPhaseCollectContext    = "collect_context"
	controlPhaseReplyPost         = "reply_post"
	controlWorkRunWorkflow        = "run_workflow"
	controlWorkResumeAfterContext = "resume_after_context"
	controlWorkResumeAfterReply   = "resume_after_reply"
)

type workflowContext struct {
	trace     events.Trace
	workflow  storepkg.Workflow
	ingestion slackpkg.Ingestion
}

func RunWorker(cfg config.Config, store storepkg.Store) error {
	runnerClients := map[string]*clients.RunnerClient{
		"prod":      clients.NewRunnerClientWithTimeout(cfg.RunnerURLForRole("prod"), cfg.RunnerTimeoutForRole("prod")),
		"proactive": clients.NewRunnerClientWithTimeout(cfg.RunnerURLForRole("proactive"), cfg.RunnerTimeoutForRole("proactive")),
	}
	workerID := fmt.Sprintf("%s-worker", cfg.ServiceName)
	for {
		item, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.WorkflowQueue, queue.ProactiveQueue}, workerID, cfg.WorkItemLeaseDuration)
		if err != nil {
			return err
		}
		if !ok {
			time.Sleep(cfg.WorkerPollInterval)
			continue
		}
		runnerClient := runnerClients[runnerRoleForQueue(item.Queue)]
		if err := processWorkflowItem(cfg, store, runnerClient, item); err != nil {
			log.Printf("control-plane worker item=%s error=%v", item.ID, err)
			_, _ = store.FailWorkItem(item.ID, err.Error())
			_, _ = store.ApplyTraceUpdate(item.TraceID, storepkg.TraceUpdate{
				Status: ptrStatus(events.StatusFailed),
				Events: []events.TraceEvent{
					{
						TraceID:     item.TraceID,
						IngestionID: item.IngestionID,
						WorkflowID:  item.WorkflowID,
						Plane:       "control",
						Service:     cfg.ServiceName,
						Actor:       "worker",
						EventType:   "workflow.failed",
						Status:      events.StatusFailed,
						StartedAt:   time.Now().UTC(),
						Description: err.Error(),
					},
				},
				WorkflowStatus: "failed",
				WorkflowError:  err.Error(),
			})
			continue
		}
		_, _ = store.CompleteWorkItem(item.ID)
	}
}

func RunActionWorker(cfg config.Config, store storepkg.Store) error {
	toolClient := clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL)
	workerID := fmt.Sprintf("%s-action-worker", cfg.ServiceName)
	for {
		item, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.ControlActionQueue}, workerID, cfg.WorkItemLeaseDuration)
		if err != nil {
			return err
		}
		if !ok {
			time.Sleep(cfg.WorkerPollInterval)
			continue
		}
		if err := processControlActionItem(cfg, store, toolClient, item); err != nil {
			log.Printf("control-plane action worker item=%s error=%v", item.ID, err)
			_, _ = store.FailWorkItem(item.ID, err.Error())
			continue
		}
		_, _ = store.CompleteWorkItem(item.ID)
	}
}

func processWorkflowItem(cfg config.Config, store storepkg.Store, runnerClient *clients.RunnerClient, item queue.WorkItem) error {
	switch strings.TrimSpace(item.Kind) {
	case "", "workflow", controlWorkRunWorkflow:
		return runWorkflowPhase(cfg, store, item)
	case controlWorkResumeAfterContext:
		return resumeAfterContextPhase(cfg, store, runnerClient, item)
	case controlWorkResumeAfterReply:
		return resumeAfterReplyPhase(cfg, store, item)
	default:
		return fmt.Errorf("unsupported control-plane work item kind %s", item.Kind)
	}
}

func runWorkflowPhase(cfg config.Config, store storepkg.Store, item queue.WorkItem) error {
	ctx, err := loadWorkflowContext(store, item)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	_, _ = store.UpdateWorkflowStatus(ctx.workflow.ID, "running", "")

	update := storepkg.TraceUpdate{
		Status:         ptrStatus(events.StatusRunning),
		WorkflowStatus: "running",
	}
	if !traceHasEventType(ctx.trace, "workflow.started") {
		update.Events = append(update.Events, events.TraceEvent{
			TraceID:     ctx.trace.Summary.TraceID,
			IngestionID: ctx.trace.Summary.IngestionID,
			WorkflowID:  ctx.trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     cfg.ServiceName,
			Actor:       "worker",
			EventType:   "workflow.started",
			Status:      events.StatusRunning,
			StartedAt:   now,
			Description: fmt.Sprintf("Started %s workflow worker loop.", ctx.workflow.Intent),
		})
		update.Reasoning = append(update.Reasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-start-%d", now.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "pre_action_summary",
			Summary:    fmt.Sprintf("Preparing %s response for conversation %s.", ctx.workflow.Intent, ctx.trace.Summary.ConversationID),
			Confidence: 0.9,
			Decision:   fmt.Sprintf("response_mode:%s", ctx.workflow.ResponseMode),
			CreatedAt:  now,
		})
	}

	toolNames := toolPlanForIntent(ctx.workflow.Intent, ingestionText(ctx.ingestion), resolveTargetRepo(cfg, ctx.ingestion.Text))
	for _, toolName := range toolNames {
		intent, created, err := ensureActionIntent(store, action.Intent{
			OwnerPlane:     "control",
			ConversationID: ctx.trace.Summary.ConversationID,
			CaseID:         ctx.trace.Summary.CaseID,
			TraceID:        ctx.trace.Summary.TraceID,
			Kind:           action.KindToolRead,
			PhaseKey:       controlPhaseCollectContext,
			TargetRef:      toolName,
			RequestPayload: toolInputForIntent(cfg, ctx.workflow, ctx.ingestion),
			IdempotencyKey: fmt.Sprintf("%s:%s:%s", ctx.trace.Summary.TraceID, toolName, ctx.trace.Summary.TriggerEventID),
			ApprovalMode:   "not_required",
			ApprovalState:  "not_required",
			Status:         action.StatusQueued,
			RequestedBy:    cfg.ServiceName,
			Rationale:      fmt.Sprintf("Collect context via %s.", toolName),
			EvidenceRefs: []events.EvidenceRef{
				{Kind: "trace", Ref: ctx.trace.Summary.TraceID, Summary: ctx.trace.Summary.WorkflowKind},
			},
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return err
		}
		if created {
			update.Events = append(update.Events, events.TraceEvent{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     "tool-gateway",
				Actor:       ctx.workflow.AssignedBot,
				EventType:   "tool.requested",
				Status:      events.StatusQueued,
				StartedAt:   now,
				Description: fmt.Sprintf("Requested %s.", toolName),
			})
		}
		if _, err := enqueueControlActionWork(store, item.Queue, ctx, intent, ctx.workflow); err != nil {
			return err
		}
	}
	if len(update.Events) > 0 || len(update.Reasoning) > 0 || update.Status != nil {
		if _, err := store.ApplyTraceUpdate(ctx.trace.Summary.TraceID, update); err != nil {
			return err
		}
	}
	if len(toolNames) == 0 || phaseActionsTerminal(store, ctx.trace.Summary.TraceID, controlPhaseCollectContext) {
		_, err := enqueueWorkflowResume(store, item.Queue, ctx, controlWorkResumeAfterContext, now)
		return err
	}
	return nil
}

func resumeAfterContextPhase(cfg config.Config, store storepkg.Store, runnerClient *clients.RunnerClient, item queue.WorkItem) error {
	ctx, err := loadWorkflowContext(store, item)
	if err != nil {
		return err
	}
	refreshedTrace, ok := store.GetTrace(ctx.trace.Summary.TraceID)
	if ok {
		ctx.trace = refreshedTrace
	}
	if isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return nil
	}
	contextSummary, contextRefs, toolNames := contextFromTrace(ctx.trace)
	runnerStarted := time.Now().UTC()
	if _, err := store.ApplyTraceUpdate(ctx.trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "context.collected",
				Status:      events.StatusCompleted,
				StartedAt:   runnerStarted,
				Description: firstNonEmpty(contextSummary, "Context collection phase completed."),
			},
			{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       ctx.workflow.AssignedBot,
				EventType:   "runner.started",
				Status:      events.StatusRunning,
				StartedAt:   runnerStarted,
				Description: "Runner task dispatched with verbose reasoning enabled.",
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:           fmt.Sprintf("reason-context-%d", runnerStarted.UnixNano()),
				TraceID:      ctx.trace.Summary.TraceID,
				WorkflowID:   ctx.trace.Summary.WorkflowID,
				StepType:     "context_collected",
				Summary:      firstNonEmpty(contextSummary, "No external context tools were required."),
				EvidenceRefs: evidenceRefsFromContext(contextRefs),
				Confidence:   0.82,
				Decision:     strings.Join(toolNames, ","),
				CreatedAt:    runnerStarted,
			},
		},
	}); err != nil {
		return err
	}

	role := runnerRoleForQueue(item.Queue)
	runnerTask := buildRunnerTask(cfg, store, role, ctx.trace, ctx.workflow, ctx.ingestion, contextSummary, contextRefs, toolNames)
	runnerResp, err := runnerClient.Execute(runnerTask)
	if err != nil {
		return err
	}
	if err := runnerutil.PersistHarnessExecution(
		store,
		runnerResp,
		role,
		ctx.trace.Summary.TraceID,
		"",
		runnerTask.HarnessProfileID,
		runnerTask.HarnessOverlayVersion,
		runnerTask.SessionScopeKind,
		runnerTask.SessionScopeID,
		runnerTask.ParentSessionScopeKind,
		runnerTask.ParentSessionScopeID,
	); err != nil {
		return err
	}
	runnerOutput := runnerutil.ParseStructuredOutput(runnerResp)
	runnerCompleted := time.Now().UTC()
	if err := persistKnowledgeDrafts(store, ctx.trace, runnerOutput.KnowledgeDrafts, runnerCompleted); err != nil {
		return err
	}

	allowed, policyVerdict := replyPolicy(store, ctx.workflow.Kind, ctx.trace.Summary.ThreadKey, ctx.ingestion.ChannelID)
	replyBody := firstNonEmpty(strings.TrimSpace(runnerOutput.FinalAnswer), strings.TrimSpace(runnerOutput.ReplyDraft), strings.TrimSpace(runnerResp.Message))
	replyAction, draftEvents, draftReasoning, err := draftSlackPostAction(cfg, store, item.Queue, ctx, runnerOutput, replyBody, allowed, policyVerdict, runnerCompleted)
	if err != nil {
		return err
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
	finalReasoning = append(finalReasoning, draftReasoning...)
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
	runnerDescription := "Runner returned visible reasoning and drafted reply action."
	if strings.TrimSpace(replyAction.ID) == "" {
		runnerDescription = "Runner returned visible reasoning."
	}
	_, err = store.ApplyTraceUpdate(ctx.trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: append([]events.TraceEvent{
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
				EndedAt:     ptrTime(runnerCompleted),
				Description: runnerDescription,
			},
		}, draftEvents...),
		Reasoning: finalReasoning,
	})
	return err
}

func resumeAfterReplyPhase(cfg config.Config, store storepkg.Store, item queue.WorkItem) error {
	ctx, err := loadWorkflowContext(store, item)
	if err != nil {
		return err
	}
	completedAt := time.Now().UTC()
	traceStatus, workflowStatus, workflowError := workflowOutcomeForTrace(store, ctx.trace.Summary.TraceID)
	if _, err := store.UpdateWorkflowStatus(ctx.workflow.ID, workflowStatus, workflowError); err != nil {
		return err
	}
	if _, err := queueEvalWork(store, cfg.ServiceName, ctx, completedAt); err != nil {
		return err
	}
	_, err = store.ApplyTraceUpdate(ctx.trace.Summary.TraceID, storepkg.TraceUpdate{
		Status:         ptrStatus(traceStatus),
		WorkflowStatus: workflowStatus,
		WorkflowError:  workflowError,
		Events: []events.TraceEvent{
			{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     "improvement-plane",
				Actor:       "eval-scheduler",
				EventType:   "eval_trace.queued",
				Status:      events.StatusQueued,
				StartedAt:   completedAt,
				Description: "Queued trace for post-completion evaluation.",
			},
			{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "workflow.completed",
				Status:      traceStatus,
				StartedAt:   completedAt,
				Description: fmt.Sprintf("Workflow ended with %s.", workflowStatus),
			},
		},
	})
	return err
}

func processControlActionItem(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, item queue.WorkItem) error {
	actionID := stringFromMap(item.Payload, "action_intent_id")
	if actionID == "" {
		return fmt.Errorf("control action item %s missing action_intent_id", item.ID)
	}
	intent, ok := store.GetActionIntent(actionID)
	if !ok {
		return fmt.Errorf("action intent %s not found", actionID)
	}
	if isTerminalActionStatus(intent.Status) {
		return maybeEnqueuePhaseResume(store, item, intent)
	}
	ctx, err := loadWorkflowContext(store, queue.WorkItem{
		TraceID:     intent.TraceID,
		WorkflowID:  item.WorkflowID,
		IngestionID: item.IngestionID,
	})
	if err != nil {
		return err
	}
	if isTerminalTraceStatus(ctx.trace.Summary.Status) && !isTerminalActionStatus(intent.Status) {
		intent.Status = action.StatusCanceled
		intent.PolicyVerdict = firstNonEmpty(intent.PolicyVerdict, fmt.Sprintf("trace_%s", ctx.trace.Summary.Status))
		intent.UpdatedAt = time.Now().UTC()
		_, err := store.UpsertActionIntent(intent)
		return err
	}

	intent.Status = action.StatusExecuting
	intent.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertActionIntent(intent); err != nil {
		return err
	}

	switch intent.Kind {
	case action.KindToolRead:
		if err := executeToolReadActionIntent(store, toolClient, ctx, intent); err != nil {
			if isPostgresActionPersistenceError(err) {
				if finalizeErr := finalizeControlActionPersistenceFailure(cfg, store, item, ctx, intent, err); finalizeErr != nil {
					return fmt.Errorf("finalize control action persistence failure: %w (original error: %v)", finalizeErr, err)
				}
			}
			return err
		}
	case action.KindSlackPost:
		if err := executeSlackPostActionIntent(cfg, store, toolClient, ctx, intent); err != nil {
			if isPostgresActionPersistenceError(err) {
				if finalizeErr := finalizeControlActionPersistenceFailure(cfg, store, item, ctx, intent, err); finalizeErr != nil {
					return fmt.Errorf("finalize control action persistence failure: %w (original error: %v)", finalizeErr, err)
				}
			}
			return err
		}
	default:
		return fmt.Errorf("unsupported control action kind %s", intent.Kind)
	}
	intent, _ = store.GetActionIntent(actionID)
	return maybeEnqueuePhaseResume(store, item, intent)
}

func executeToolReadActionIntent(store storepkg.Store, toolClient *clients.ToolGatewayClient, ctx workflowContext, intent action.Intent) error {
	started := time.Now().UTC()
	result, execErr := toolClient.Execute(intent.TargetRef, cloneAnyMap(intent.RequestPayload))
	completed := time.Now().UTC()
	actionStatus := actionStatusFromToolResult(result, execErr)
	summary := toolResultSummary(result, execErr)
	intent.Status = actionStatus
	intent.ApprovalState = firstNonEmpty(result.ApprovalState, intent.ApprovalState)
	intent.UpdatedAt = completed
	if _, err := store.UpsertActionIntent(intent); err != nil {
		return err
	}
	if _, err := store.RecordActionResult(action.Result{
		ActionIntentID: intent.ID,
		Executor:       "tool-gateway",
		Provider:       firstNonEmpty(result.Provider, providerForToolName(intent.TargetRef)),
		ProviderRef:    result.ProviderRef,
		Status:         actionStatus,
		ErrorCode:      actionErrorCode(actionStatus),
		ErrorMessage:   actionErrorMessage(result, execErr),
		StartedAt:      started,
		CompletedAt:    completed,
	}); err != nil {
		return err
	}

	eventStatus := events.StatusCompleted
	eventType := "tool.completed"
	switch actionStatus {
	case action.StatusBlocked:
		eventStatus = events.StatusNeedsHuman
		eventType = "tool.blocked"
	case action.StatusFailed:
		eventStatus = events.StatusNeedsHuman
		eventType = "tool.failed"
	}
	_, err := store.ApplyTraceUpdate(ctx.trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     "tool-gateway",
				Actor:       ctx.workflow.AssignedBot,
				EventType:   eventType,
				Status:      eventStatus,
				StartedAt:   started,
				EndedAt:     ptrTime(completed),
				Description: summary,
			},
		},
		ToolCalls: []events.ToolCallRecord{
			{
				ID:                    fmt.Sprintf("tool-record-%d", completed.UnixNano()),
				TraceID:               ctx.trace.Summary.TraceID,
				WorkflowID:            ctx.trace.Summary.WorkflowID,
				ConversationID:        ctx.trace.Summary.ConversationID,
				CaseID:                ctx.trace.Summary.CaseID,
				ToolName:              intent.TargetRef,
				ToolCallID:            firstNonEmpty(result.ToolCallID, intent.ID),
				Request:               firstNonEmptyMap(result.Input, intent.RequestPayload),
				Summary:               summary,
				RawArtifactRefs:       append([]string(nil), result.RawArtifactRefs...),
				ApprovalState:         firstNonEmpty(result.ApprovalState, intent.ApprovalState),
				InterpretationSummary: summary,
				Status:                string(actionStatus),
				CreatedAt:             completed,
			},
		},
	})
	return err
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

	intent.Status = actionStatus
	intent.ApprovalState = firstNonEmpty(result.ApprovalState, intent.ApprovalState)
	intent.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertActionIntent(intent); err != nil {
		return err
	}
	if _, err := store.RecordActionResult(action.Result{
		ActionIntentID: intent.ID,
		Executor:       firstNonEmpty(result.Provider, "tool-gateway"),
		Provider:       firstNonEmpty(result.Provider, "slack"),
		ProviderRef:    firstNonEmpty(result.ProviderRef, threadTS),
		Status:         actionStatus,
		ErrorCode:      actionErrorCode(actionStatus),
		ErrorMessage:   firstNonEmpty(actionErrorMessage(result, execErr), blockedReason),
		StartedAt:      started,
		CompletedAt:    time.Now().UTC(),
	}); err != nil {
		return err
	}

	eventType := "slack.reply.posted"
	eventStatus := events.StatusCompleted
	switch actionStatus {
	case action.StatusBlocked:
		eventType = "slack.reply.blocked"
		eventStatus = events.StatusNeedsHuman
	case action.StatusFailed:
		eventType = "slack.reply.failed"
		eventStatus = events.StatusNeedsHuman
	}
	if actionStatus == action.StatusSucceeded && strings.TrimSpace(baseRecord.SendStatus) == "" {
		baseRecord.SendStatus = "posted"
	}
	endedAt := time.Now().UTC()
	_, err := store.ApplyTraceUpdate(ctx.trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "edge",
				Service:     "tool-gateway",
				Actor:       ctx.workflow.AssignedBot,
				EventType:   eventType,
				Status:      eventStatus,
				StartedAt:   started,
				EndedAt:     ptrTime(endedAt),
				Description: firstNonEmpty(summary, blockedReason, "Slack reply action completed."),
			},
		},
		SlackActions: []events.SlackActionRecord{baseRecord},
	})
	return err
}

func maybeEnqueuePhaseResume(store storepkg.Store, item queue.WorkItem, intent action.Intent) error {
	if intent.TraceID == "" || !phaseActionsTerminal(store, intent.TraceID, intent.PhaseKey) {
		return nil
	}
	trace, ok := store.GetTrace(intent.TraceID)
	if !ok || isTerminalTraceStatus(trace.Summary.Status) {
		return nil
	}
	queueName := queueNameFromString(firstNonEmpty(stringFromMap(item.Payload, "resume_queue"), string(queue.WorkflowQueue)))
	ctx, err := loadWorkflowContext(store, queue.WorkItem{
		TraceID:     intent.TraceID,
		WorkflowID:  trace.Summary.WorkflowID,
		IngestionID: trace.Summary.IngestionID,
	})
	if err != nil {
		return err
	}
	switch intent.PhaseKey {
	case controlPhaseCollectContext:
		_, err = enqueueWorkflowResume(store, queueName, ctx, controlWorkResumeAfterContext, time.Now().UTC())
		return err
	case controlPhaseReplyPost:
		_, err = enqueueWorkflowResume(store, queueName, ctx, controlWorkResumeAfterReply, time.Now().UTC())
		return err
	default:
		return nil
	}
}

func buildRunnerTask(cfg config.Config, store storepkg.Store, role string, trace events.Trace, workflow storepkg.Workflow, ingestion slackpkg.Ingestion, contextSummary string, contextRefs []map[string]any, toolNames []string) clients.RunnerTask {
	effectiveHarness := harness.ResolveEffectiveConfig(store, role, cfg.DefaultReasoningVerbosity)
	caseSummary := map[string]any{}
	if caseRecord, ok := store.GetCase(trace.Summary.CaseID); ok {
		caseSummary = map[string]any{
			"case_id":         caseRecord.ID,
			"conversation_id": caseRecord.ConversationID,
			"kind":            caseRecord.Kind,
			"intent":          caseRecord.Intent,
			"title":           caseRecord.Title,
			"summary":         caseRecord.Summary,
			"status":          caseRecord.Status,
			"assigned_bot":    caseRecord.AssignedBot,
			"latest_trace_id": caseRecord.LatestTraceID,
		}
	}
	recentEntries := recentConversationEntries(store.ListConversationEntries(trace.Summary.ConversationID))
	priorTraceRefs := priorTraceRefsForCase(store.ListTraces(), trace.Summary.CaseID, trace.Summary.TraceID)
	sessionScopeKind, sessionScopeID, parentScopeKind, parentScopeID := workflowSessionScope(trace, workflow)
	systemMessage := harness.ComposeSystemMessage(
		"Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, and outcome_hypotheses. Do not assume actions from prose; emit them explicitly.",
		effectiveHarness,
	)
	prompt := fmt.Sprintf("User request: %s\n\nRespond in-thread for intent=%s. Use the gathered context and cite concrete evidence when possible. If unsure, say so explicitly.", ingestion.Text, workflow.Intent)
	return clients.RunnerTask{
		TaskType:                  "workflow",
		Repo:                      cfg.DefaultRepo,
		RepoRef:                   "main",
		Prompt:                    prompt,
		SystemMessage:             systemMessage,
		AllowedTools:              harness.ApplyToolPreference(toolNames, effectiveHarness.ToolPreferenceOrder),
		AllowedCommands:           []string{},
		TimeoutSeconds:            120,
		ExpectedOutputs:           []string{"visible_reasoning", "final_answer"},
		ArtifactDestination:       fmt.Sprintf("trace:%s", trace.Summary.TraceID),
		ContextSummary:            contextSummary,
		Intent:                    workflow.Intent,
		TraceID:                   trace.Summary.TraceID,
		WorkflowID:                trace.Summary.WorkflowID,
		ConversationID:            trace.Summary.ConversationID,
		CaseID:                    trace.Summary.CaseID,
		TriggerEventID:            trace.Summary.TriggerEventID,
		RecentConversationEntries: recentEntries,
		CaseSummary:               caseSummary,
		PriorTraceRefs:            priorTraceRefs,
		RepoAllowlist:             cfg.AllowedTargetRepos,
		ToolAllowlist:             harness.ApplyToolPreference(toolNames, effectiveHarness.ToolPreferenceOrder),
		ResponseMode:              workflow.ResponseMode,
		ContextRefs:               contextRefs,
		ApprovalMode:              workflow.ApprovalMode,
		ReasoningVerbosity:        effectiveHarness.ReasoningVerbosity,
		RejectedProposalContext:   []map[string]any{},
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

func draftSlackPostAction(cfg config.Config, store storepkg.Store, resumeQueue queue.QueueName, ctx workflowContext, output runnerutil.StructuredOutput, replyBody string, allowed bool, policyVerdict string, createdAt time.Time) (action.Intent, []events.TraceEvent, []events.ReasoningStep, error) {
	proposed := firstSlackPostAction(output.ProposedActions)
	idempotencyKey := firstNonEmpty(proposed.IdempotencyKey, fmt.Sprintf("%s:%s:%s", ctx.ingestion.ChannelID, ctx.ingestion.ThreadTS, ctx.trace.Summary.TraceID))
	requestPayload := map[string]any{
		"channel_id":     ctx.ingestion.ChannelID,
		"thread_ts":      ctx.ingestion.ThreadTS,
		"body":           replyBody,
		"draft_body":     firstNonEmpty(output.ReplyDraft, replyBody),
		"final_body":     replyBody,
		"policy_verdict": policyVerdict,
	}

	approvalState := "approved"
	resolvedPolicyVerdict := policyVerdict
	reasoning := []events.ReasoningStep{}
	if proposed.Kind == "" {
		approvalState = "missing_explicit_action"
		resolvedPolicyVerdict = "missing_explicit_action"
		requestPayload["blocked_reason"] = "missing_explicit_action"
		reasoning = append(reasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-slack-blocked-%d", createdAt.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "action_blocked",
			Summary:    "Blocked Slack posting because the runner did not emit an explicit slack_post action.",
			Confidence: 0.95,
			Decision:   "needs_human",
			CreatedAt:  createdAt,
		})
	} else if !allowed {
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
		TargetRef:      firstNonEmpty(proposed.TargetRef, ctx.ingestion.ChannelID),
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
	if _, err := enqueueControlActionWork(store, resumeQueue, ctx, intent, ctx.workflow); err != nil {
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
			Description: "Drafted explicit Slack reply action.",
		},
	}, reasoning, nil
}

func recentConversationEntries(items []conversation.Entry) []map[string]any {
	if len(items) > 8 {
		items = items[len(items)-8:]
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"id":              item.ID,
			"event_id":        item.EventID,
			"entry_type":      item.EntryType,
			"actor_id":        item.ActorID,
			"actor_type":      item.ActorType,
			"body":            item.Body,
			"created_at":      item.CreatedAt,
			"source":          item.Source,
			"source_event_id": item.SourceEventID,
		})
	}
	return out
}

func priorTraceRefsForCase(items []events.TraceSummary, caseID string, currentTraceID string) []map[string]any {
	out := make([]map[string]any, 0)
	for _, item := range items {
		if item.CaseID != caseID || item.TraceID == currentTraceID {
			continue
		}
		out = append(out, map[string]any{
			"trace_id":         item.TraceID,
			"status":           item.Status,
			"workflow_kind":    item.WorkflowKind,
			"started_at":       item.StartedAt,
			"trigger_event_id": item.TriggerEventID,
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

func toolPlanForIntent(intent string, question string, repo string) []string {
	switch intent {
	case "incident":
		return []string{"sentry.lookup", "kubernetes.inspect"}
	case "feature_request":
		return []string{"repo.context", "github.repo_context"}
	default:
		plan := []string{"repo.context", "knowledge.context"}
		if shouldUseGitHubRepoActivity(question, repo) {
			plan = append(plan, "github.repo_activity")
		}
		return plan
	}
}

func runnerRoleForQueue(name queue.QueueName) string {
	switch name {
	case queue.ProactiveQueue:
		return "proactive"
	default:
		return "prod"
	}
}

func replyPolicy(store storepkg.Store, workflowKind string, threadKey string, channelID string) (bool, string) {
	for _, item := range store.ListThreadPolicies() {
		if item.ThreadKey == threadKey && item.Muted {
			return false, "thread_muted"
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

func evidenceRefsFromContext(contextRefs []map[string]any) []events.EvidenceRef {
	out := make([]events.EvidenceRef, 0, len(contextRefs))
	for _, ref := range contextRefs {
		out = append(out, events.EvidenceRef{
			Kind:    stringFromMap(ref, "kind"),
			Ref:     firstNonEmpty(stringFromMap(ref, "ref"), stringFromMap(ref, "tool_call_id")),
			Summary: stringFromMap(ref, "summary"),
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
	for _, item := range drafts {
		freshUntil := parseTimeOrNil(item.FreshUntil)
		_, err := store.UpsertKnowledgeEntry(knowledge.Entry{
			Tier:       knowledge.TierWorking,
			Kind:       knowledge.Kind(firstNonEmpty(item.Kind, string(knowledge.KindFact))),
			ScopeType:  knowledge.ScopeType(firstNonEmpty(item.ScopeType, string(knowledge.ScopeCase))),
			ScopeID:    firstNonEmpty(item.ScopeID, trace.Summary.CaseID),
			Title:      item.Title,
			Summary:    item.Summary,
			Body:       item.Body,
			Status:     knowledge.StatusDraft,
			Confidence: item.Confidence,
			FreshUntil: freshUntil,
			SourceType: knowledge.SourceAgent,
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}, evidenceLinksFromDraft(item))
		if err != nil {
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

func ptrStatus(status events.Status) *events.Status {
	return &status
}

func ptrTime(value time.Time) *time.Time {
	return &value
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

func loadWorkflowContext(store storepkg.Store, item queue.WorkItem) (workflowContext, error) {
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return workflowContext{}, fmt.Errorf("trace %s not found", item.TraceID)
	}
	workflowID := firstNonEmpty(item.WorkflowID, trace.Summary.WorkflowID)
	workflow, ok := findWorkflow(store.ListWorkflows(), workflowID)
	if !ok {
		return workflowContext{}, fmt.Errorf("workflow %s not found", workflowID)
	}
	ingestionID := firstNonEmpty(item.IngestionID, trace.Summary.IngestionID)
	ingestion, ok := findIngestion(store.ListIngestions(), ingestionID)
	if !ok {
		return workflowContext{}, fmt.Errorf("ingestion %s not found", ingestionID)
	}
	return workflowContext{trace: trace, workflow: workflow, ingestion: ingestion}, nil
}

func ensureActionIntent(store storepkg.Store, template action.Intent) (action.Intent, bool, error) {
	if existing, ok := findActionIntentByIdempotencyKey(store.ListActionIntents(), template.IdempotencyKey); ok {
		return existing, false, nil
	}
	created, err := store.UpsertActionIntent(template)
	if err != nil {
		return action.Intent{}, false, err
	}
	return created, true, nil
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

func enqueueControlActionWork(store storepkg.Store, resumeQueue queue.QueueName, ctx workflowContext, intent action.Intent, workflow storepkg.Workflow) (queue.WorkItem, error) {
	return store.EnqueueWorkItem(queue.WorkItem{
		Queue:          queue.ControlActionQueue,
		Kind:           "execute_action",
		Status:         queue.WorkQueued,
		TraceID:        ctx.trace.Summary.TraceID,
		WorkflowID:     ctx.trace.Summary.WorkflowID,
		IngestionID:    ctx.trace.Summary.IngestionID,
		ConversationID: ctx.trace.Summary.ConversationID,
		CaseID:         ctx.trace.Summary.CaseID,
		TriggerEventID: ctx.trace.Summary.TriggerEventID,
		ThreadKey:      ctx.trace.Summary.ThreadKey,
		Intent:         workflow.Intent,
		RequestedBy:    "control_orchestrator",
		ApprovalMode:   workflow.ApprovalMode,
		ResponseMode:   workflow.ResponseMode,
		Payload: map[string]any{
			"action_intent_id": intent.ID,
			"resume_queue":     string(resumeQueue),
			"dedupe_key":       fmt.Sprintf("control_action:%s", intent.ID),
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
}

func enqueueWorkflowResume(store storepkg.Store, queueName queue.QueueName, ctx workflowContext, kind string, createdAt time.Time) (queue.WorkItem, error) {
	return store.EnqueueWorkItem(queue.WorkItem{
		Queue:          queueName,
		Kind:           kind,
		Status:         queue.WorkQueued,
		TraceID:        ctx.trace.Summary.TraceID,
		WorkflowID:     ctx.trace.Summary.WorkflowID,
		IngestionID:    ctx.trace.Summary.IngestionID,
		ConversationID: ctx.trace.Summary.ConversationID,
		CaseID:         ctx.trace.Summary.CaseID,
		TriggerEventID: ctx.trace.Summary.TriggerEventID,
		ThreadKey:      ctx.trace.Summary.ThreadKey,
		Intent:         ctx.workflow.Intent,
		RequestedBy:    "control_action_worker",
		ApprovalMode:   ctx.workflow.ApprovalMode,
		ResponseMode:   ctx.workflow.ResponseMode,
		Payload: map[string]any{
			"dedupe_key": fmt.Sprintf("%s:%s", ctx.trace.Summary.TraceID, kind),
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
}

func queueEvalWork(store storepkg.Store, requestedBy string, ctx workflowContext, createdAt time.Time) (queue.WorkItem, error) {
	return store.EnqueueWorkItem(queue.WorkItem{
		Queue:          queue.EvalQueue,
		Kind:           "evaluate_trace",
		Status:         queue.WorkQueued,
		TraceID:        ctx.trace.Summary.TraceID,
		WorkflowID:     ctx.trace.Summary.WorkflowID,
		IngestionID:    ctx.trace.Summary.IngestionID,
		ConversationID: ctx.trace.Summary.ConversationID,
		CaseID:         ctx.trace.Summary.CaseID,
		TriggerEventID: ctx.trace.Summary.TriggerEventID,
		ThreadKey:      ctx.trace.Summary.ThreadKey,
		RequestedBy:    requestedBy,
		ApprovalMode:   ctx.workflow.ApprovalMode,
		ResponseMode:   ctx.workflow.ResponseMode,
		Payload: map[string]any{
			"dedupe_key": fmt.Sprintf("%s:evaluate_trace", ctx.trace.Summary.TraceID),
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
}

func toolInputForIntent(cfg config.Config, workflow storepkg.Workflow, ingestion slackpkg.Ingestion) map[string]any {
	repo := resolveTargetRepo(cfg, ingestion.Text)
	since, until := repoActivityWindow(ingestion.Text, time.Now().UTC())
	return map[string]any{
		"repo":               repo,
		"question":           ingestion.Text,
		"topic":              ingestion.Text,
		"scope_id":           repo,
		"service":            workflow.AssignedBot,
		"alert":              ingestion.Text,
		"namespace":          cfg.SandboxNamespace,
		"target":             workflow.Kind,
		"knowledge_base_url": cfg.DefaultKnowledgeBaseURL,
		"channel_id":         ingestion.ChannelID,
		"thread_ts":          ingestion.ThreadTS,
		"since":              since,
		"until":              until,
	}
}

func contextFromTrace(trace events.Trace) (string, []map[string]any, []string) {
	contextRefs := make([]map[string]any, 0, len(trace.ToolCalls))
	summaries := make([]string, 0, len(trace.ToolCalls))
	toolNames := make([]string, 0, len(trace.ToolCalls))
	for _, call := range trace.ToolCalls {
		contextRefs = append(contextRefs, map[string]any{
			"kind":         "tool_call",
			"ref":          firstNonEmpty(call.ToolCallID, call.ID),
			"tool_call_id": firstNonEmpty(call.ToolCallID, call.ID),
			"summary":      call.Summary,
			"tool_name":    call.ToolName,
			"status":       call.Status,
		})
		summaries = append(summaries, call.Summary)
		if call.ToolName != "" {
			toolNames = append(toolNames, call.ToolName)
		}
	}
	if len(toolNames) == 0 {
		toolNames = toolPlanForIntent(trace.Summary.WorkflowKind, "", "")
	}
	return strings.Join(summaries, " "), contextRefs, uniqueStrings(toolNames)
}

func ingestionText(ingestion slackpkg.Ingestion) string {
	return strings.TrimSpace(ingestion.Text)
}

func resolveTargetRepo(cfg config.Config, question string) string {
	text := strings.ToLower(strings.TrimSpace(question))
	for _, repo := range cfg.AllowedTargetRepos {
		repo = strings.TrimSpace(repo)
		if repo == "" {
			continue
		}
		if strings.Contains(text, strings.ToLower(repo)) {
			return repo
		}
	}
	return cfg.DefaultRepo
}

func shouldUseGitHubRepoActivity(question string, repo string) bool {
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

func repoActivityWindow(question string, now time.Time) (string, string) {
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

func finalizeControlActionPersistenceFailure(cfg config.Config, store storepkg.Store, item queue.WorkItem, ctx workflowContext, intent action.Intent, execErr error) error {
	failure := classifyActionPersistenceFailure(intent, execErr)
	now := time.Now().UTC()
	intent.Status = action.StatusFailed
	intent.PolicyVerdict = firstNonEmpty(failure.FailureMode, "action_result_persistence_failure")
	intent.UpdatedAt = now
	if _, err := store.UpsertActionIntent(intent); err != nil {
		return err
	}

	description := actionPersistenceFailureDescription(intent, item, failure, execErr)
	if _, err := store.ApplyTraceUpdate(ctx.trace.Summary.TraceID, storepkg.TraceUpdate{
		Status:         ptrStatus(events.StatusNeedsHuman),
		WorkflowStatus: "needs-human",
		WorkflowError:  description,
		Events: []events.TraceEvent{
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

	resumeQueue := queueNameFromString(firstNonEmpty(stringFromMap(item.Payload, "resume_queue"), string(queue.WorkflowQueue)))
	_, err := enqueueWorkflowResume(store, resumeQueue, ctx, controlWorkResumeAfterReply, now)
	return err
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

func actionPersistenceFailureDescription(intent action.Intent, item queue.WorkItem, failure actionPersistenceFailure, execErr error) string {
	parts := []string{
		fmt.Sprintf("subsystem=%s", failure.Subsystem),
		fmt.Sprintf("failure_mode=%s", failure.FailureMode),
		fmt.Sprintf("provider=%s", firstNonEmpty(failure.Provider, "unknown")),
		fmt.Sprintf("action_intent_id=%s", intent.ID),
		fmt.Sprintf("work_item_id=%s", item.ID),
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

func isTerminalTraceStatus(status events.Status) bool {
	switch status {
	case events.StatusCompleted, events.StatusFailed, events.StatusNeedsHuman, events.StatusSuppressed:
		return true
	default:
		return false
	}
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

func traceHasEventType(trace events.Trace, eventType string) bool {
	for _, item := range trace.Events {
		if item.EventType == eventType {
			return true
		}
	}
	return false
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
