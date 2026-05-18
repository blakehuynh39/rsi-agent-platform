package control

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvementplane"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/timeutil"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
	"github.com/piplabs/rsi-agent-platform/internal/workflowplan"
)

const (
	controlPhaseCollectContext             = "collect_context"
	controlPhaseReplyPost                  = "reply_post"
	partialCompletionNoticeGeneric         = "Partial answer: I had to stop before I could finish a deeper pass. This is the best grounded answer so far."
	partialCompletionNoticeIterationBudget = "Partial answer: I hit my iteration budget before I could finish a deeper pass. This is the best grounded answer so far."
	partialCompletionNoticeTaskTimeout     = "Partial answer: I hit the workflow time limit before I could finish a deeper pass. This is the best grounded answer so far."
	partialCompletionNoticeOutputBudget    = "Partial answer: I hit my response output budget before I could finish a deeper pass. This is the best grounded answer so far."
	hermesExecutionRecoveryPollDelay       = 15 * time.Second
)

var errHermesExecutionStillRunning = errors.New("existing Hermes execution is still running")

type workflowContext struct {
	trace     events.Trace
	workflow  storepkg.Workflow
	ingestion slackpkg.Ingestion
}

type resolvedWorkflowIntent struct {
	UserRequest      string
	PlanningQuestion string
}

type workflowLocator struct {
	traceID     string
	workflowID  string
	ingestionID string
}

func RunWorker(cfg config.Config, store storepkg.Store) error {
	if cfg.DrainEnabled {
		app.InstallSignalDrain()
	}
	runnerClients := map[string]*clients.RunnerClient{}
	for _, role := range []string{"prod", "proactive"} {
		if baseURL := strings.TrimSpace(cfg.RunnerURLForRole(role)); baseURL != "" {
			runnerClients[role] = clients.NewRunnerClientWithTimeout(baseURL, cfg.RunnerTimeoutForRole(role))
		}
	}
	workerID := fmt.Sprintf("%s-worker", cfg.ServiceName)
	runnerEffectLease := cfg.EffectLeaseDuration(cfg.WorkItemLeaseDuration, "prod", "proactive")
	for {
		if app.IsDraining() {
			log.Printf("control-plane worker draining; stopped claiming new effects")
			return nil
		}
		effect, claimed, err := claimNextExecutionEffect(cfg, store, workerID, runnerEffectLease)
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
		handleClaimedExecutionEffect(cfg, store, runnerClients, effect)
	}
}

func RunActionWorker(cfg config.Config, store storepkg.Store) error {
	if cfg.DrainEnabled {
		app.InstallSignalDrain()
	}
	workerID := fmt.Sprintf("%s-action-worker", cfg.ServiceName)
	for {
		if app.IsDraining() {
			log.Printf("control-plane action worker draining; stopped claiming new effects")
			return nil
		}
		effect, ok, err := claimNextActionEffect(cfg, store, "control", workerID, cfg.WorkItemLeaseDuration)
		if err != nil {
			return err
		}
		if !ok {
			time.Sleep(cfg.WorkerPollInterval)
			continue
		}
		if app.IsDraining() {
			runtime := newWorkflowRuntimeCoordinator(config.Config{}, store)
			if err := runtime.deferClaimedEffectForDrain(effect); err != nil {
				_ = failClaimedEffect(store, effect, err.Error())
				log.Printf("control-plane action worker effect=%s aggregate=%s drain_defer_error=%v", effect.ID, effect.AggregateID, err)
			} else {
				log.Printf("control-plane action worker effect=%s aggregate=%s deferred_for_deployment_drain", effect.ID, effect.AggregateID)
			}
			continue
		}
		if err := processControlActionEffect(cfg, store, effect); err != nil {
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
		if errors.Is(err, errHermesExecutionStillRunning) {
			reason := "Hermes executor status is still running; deferred effect for recovery polling."
			if deferErr := deferClaimedEffect(store, effect, hermesExecutionRecoveryPollDelay, reason); deferErr != nil {
				_ = failClaimedEffect(store, effect, deferErr.Error())
				log.Printf("control-plane runner effect=%s aggregate=%s defer_existing_hermes_execution_error=%v", effect.ID, effect.AggregateID, deferErr)
			} else {
				log.Printf("control-plane runner effect=%s aggregate=%s deferred_existing_hermes_execution", effect.ID, effect.AggregateID)
			}
			return
		}
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

func handleClaimedExecutionEffect(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, effect transition.EffectExecution) {
	if app.IsDraining() {
		runtime := newWorkflowRuntimeCoordinator(config.Config{}, store)
		if err := runtime.deferClaimedEffectForDrain(effect); err != nil {
			_ = failClaimedEffect(store, effect, err.Error())
			log.Printf("control-plane effect=%s aggregate=%s drain_defer_error=%v", effect.ID, effect.AggregateID, err)
		} else {
			log.Printf("control-plane effect=%s aggregate=%s deferred_for_deployment_drain", effect.ID, effect.AggregateID)
		}
		return
	}
	switch {
	case effect.MachineKind == transition.MachineWorkflow && effect.EffectKind == transition.EffectInvokeRunner:
		handleClaimedWorkflowRunnerEffect(cfg, store, runnerClients, effect)
	case effect.MachineKind == transition.MachineWorkflow && effect.EffectKind == transition.EffectSummarizeSessionTitle:
		handleClaimedSessionTitleEffect(cfg, store, runnerClients, effect)
	default:
		_ = failClaimedEffect(store, effect, fmt.Sprintf("unsupported execution effect %s/%s", effect.MachineKind, effect.EffectKind))
	}
}

func handleClaimedSessionTitleEffect(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, effect transition.EffectExecution) {
	if err := processSessionTitleEffect(cfg, store, runnerClients, effect); err != nil {
		if failErr := failClaimedEffect(store, effect, err.Error()); failErr != nil {
			log.Printf("control-plane title effect=%s aggregate=%s fail_error=%v", effect.ID, effect.AggregateID, failErr)
		}
		log.Printf("control-plane title effect=%s aggregate=%s error=%v", effect.ID, effect.AggregateID, err)
	}
}

func processSessionTitleEffect(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, effect transition.EffectExecution) error {
	ctx, queueName, err := loadWorkflowContextForEffect(store, effect)
	if err != nil {
		return err
	}
	if isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return completeClaimedEffect(store, effect, ctx.trace.Summary.TraceID)
	}
	runnerRole := runnerRoleForQueue(queueName)
	runnerClient := runnerClients[runnerRole]
	hermesExecutorURLs := cfg.HermesExecutorURLs()
	hermesExecutorBaseURL := ""
	if len(hermesExecutorURLs) > 0 {
		hermesExecutorBaseURL = strings.TrimSpace(hermesExecutorURLs[0])
	}
	useHermesExecutor := hermesExecutorBaseURL != ""
	if runnerClient == nil && !useHermesExecutor {
		return fmt.Errorf("runner client unavailable for session title queue %s", queueName)
	}
	started := time.Now().UTC()
	runnerTask := buildSessionTitleRunnerTask(cfg, store, runnerRole, ctx, stringFromMap(effect.Payload, "raw_title"))
	runnerTask.OperationID = effect.ID
	runnerTask.ExecutionID = workflowExecutionID(effect.ID+":session-title", started)
	executorClient := runnerClient
	if useHermesExecutor {
		executorClient = clients.NewRunnerClientWithTimeout(hermesExecutorBaseURL, cfg.RunnerTimeoutForRole(runnerRole))
	}
	var runnerResp clients.RunnerResponse
	if useHermesExecutor {
		runnerResp, err = executorClient.ExecuteHermesExecution(runnerTask)
	} else {
		runnerResp, err = executorClient.Execute(runnerTask)
	}
	if err != nil {
		return err
	}
	if !runnerResp.OK {
		return fmt.Errorf("session title runner failed: %s", firstNonEmpty(strings.TrimSpace(runnerResp.Message), "runner returned ok=false"))
	}
	output, err := runnerutil.ParseStructuredOutput(runnerResp)
	if err != nil {
		return err
	}
	title := normalizeGeneratedSessionTitle(output.SessionTitle)
	if title == "" {
		return errors.New("session title runner returned empty session_title")
	}
	completed := time.Now().UTC()
	update := storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     ctx.trace.Summary.TraceID,
				IngestionID: ctx.trace.Summary.IngestionID,
				WorkflowID:  ctx.trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     cfg.ServiceName,
				Actor:       "session-title-summarizer",
				EventType:   "session.title.summarized",
				Status:      events.StatusCompleted,
				StartedAt:   started,
				EndedAt:     &completed,
				Description: "Generated a short session title from the Slack request.",
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:             fmt.Sprintf("reason-session-title-%d", completed.UnixNano()),
				TraceID:        ctx.trace.Summary.TraceID,
				WorkflowID:     ctx.trace.Summary.WorkflowID,
				ConversationID: ctx.trace.Summary.ConversationID,
				CaseID:         ctx.trace.Summary.CaseID,
				StepType:       "session_title",
				Summary:        title,
				Confidence:     firstNonZeroFloat(output.Confidence, 0.9),
				Decision:       "set_session_title",
				CreatedAt:      completed,
			},
		},
	}
	if _, err := improvementplane.SubmitProblemLineTraceProjection(store, ctx.trace.Summary.TraceID, cfg.ServiceName, completed, fmt.Sprintf("cmd-problem-line:%s:session-title:%s", ctx.trace.Summary.TraceID, effect.ID), update); err != nil {
		return err
	}
	return completeClaimedEffect(store, effect, fmt.Sprintf("trace:%s:session-title", ctx.trace.Summary.TraceID))
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
	runnerRole := runnerRoleForQueue(queueName)
	runnerClient := runnerClients[runnerRole]
	hermesExecutorURLs := cfg.HermesExecutorURLs()
	hermesExecutorBaseURL := ""
	if len(hermesExecutorURLs) > 0 {
		hermesExecutorBaseURL = strings.TrimSpace(hermesExecutorURLs[0])
	}
	useHermesExecutor := hermesExecutorBaseURL != ""
	if runnerClient == nil && !useHermesExecutor {
		return fmt.Errorf("runner client unavailable for queue %s", queueName)
	}
	effectiveWorkflow := ctx.workflow
	contextSummary, contextRefs := contextFromTrace(ctx.trace)
	if extraSummary, extraRefs := prefetchBoundSlackThreadContext(cfg, ctx.trace.Summary, effectiveWorkflow, ctx.ingestion); extraSummary != "" || len(extraRefs) > 0 {
		contextSummary = joinContextSummary(contextSummary, extraSummary)
		contextRefs = append(contextRefs, extraRefs...)
	}
	runnerStarted := time.Now().UTC()
	runnerTask := buildRunnerTask(cfg, store, runnerRole, ctx.trace, effectiveWorkflow, ctx.ingestion, contextSummary, contextRefs)
	runnerTask.OperationID = effect.ID
	runnerTask.ExecutionID = workflowExecutionID(effect.ID, runnerStarted)
	if resumePayload := mapValue(effect.Payload["external_tool_resume"]); len(resumePayload) > 0 {
		runnerTask.ExternalToolResume = resumePayload
	}
	resumePauseID := strings.TrimSpace(stringValueFromMap(effect.Payload, "external_tool_pause_id"))
	if resumePauseID != "" && len(runnerTask.ExternalToolResume) > 0 {
		_, _ = store.UpdateExternalToolPause(resumePauseID, func(item *storepkg.ExternalToolPause) error {
			item.ResumeStatus = storepkg.ExternalToolResumeRunning
			return nil
		})
	}
	executorClient := runnerClient
	if useHermesExecutor {
		executorClient = clients.NewRunnerClientWithTimeout(hermesExecutorBaseURL, cfg.RunnerTimeoutForRole(runnerRole))
		cancelSupersededHermesExecutions(cfg, store, executorClient, ctx.workflow.CaseID, ctx.trace.Summary.TraceID)
	}
	var runnerResp clients.RunnerResponse
	if useHermesExecutor && cfg.AsyncHermesExecutionEnabled {
		var waitForExisting bool
		runnerResp, waitForExisting, err = executeOrPollAsyncHermesExecution(cfg, store, executorClient, runnerTask, effect, runnerRole, ctx, runnerStarted)
		if err != nil {
			return err
		}
		if waitForExisting {
			return errHermesExecutionStillRunning
		}
	} else if useHermesExecutor {
		if shouldRecoverHermesExecution(effect) {
			recovered, waitForExisting, recoverErr := recoverHermesExecutionResult(executorClient, runnerTask.ExecutionID)
			if recoverErr != nil {
				return recoverErr
			}
			if waitForExisting {
				return errHermesExecutionStillRunning
			}
			runnerResp = recovered
		} else {
			runnerResp, err = executorClient.ExecuteHermesExecution(runnerTask)
		}
	} else {
		runnerResp, err = executorClient.Execute(runnerTask)
	}
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
	ledgerEvents := runnerutil.ExecutionLedgerEventsFromRunnerRaw(runnerResp.Raw, runnerStarted)
	useLedgerFirst := cfg.ExecutionLedgerFirstProjection && len(ledgerEvents) > 0
	if len(ledgerEvents) > 0 {
		if err := store.RecordExecutionLedgerEvents(ledgerEvents); err != nil {
			return runnerPostProcessingFailure("persist_execution_ledger", err)
		}
	}
	if ctx, err = refreshWorkflowContextState(store, ctx); err != nil {
		return runnerPostProcessingFailure("refresh_workflow_context_after_harness_execution", err)
	}
	if isTerminalWorkflowStatus(ctx.workflow.Status) || isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return completeClaimedEffect(store, effect, ctx.trace.Summary.TraceID)
	}
	if workflowTerminationReason(runnerResp.Raw) == "external_tool_pending" {
		markExternalToolResumeAttemptFinished(store, resumePauseID, runnerResp)
		return handleExternalToolPendingRunnerResult(cfg, store, ctx, effect, runnerResp, runnerStarted)
	}
	markExternalToolResumeAttemptFinished(store, resumePauseID, runnerResp)
	if !runnerResp.OK {
		if strings.TrimSpace(stringValue(runnerResp.Raw["failure_class"])) == "runner_reply_delivery_uncertain" {
			runnerCompleted := time.Now().UTC()
			runnerDiagnostics := cloneStringAnyMap(mapValue(runnerResp.Raw["runner_diagnostics"]))
			runnerDiagnostics = mergeWorkflowRunnerDiagnostics(runnerDiagnostics, runnerResp.Raw)
			runnerToolCalls := bindRunnerToolCallRecords(toolCallRecordsFromRunnerRaw(runnerResp.Raw), ctx.trace, ctx.workflow)
			if useLedgerFirst {
				runnerToolCalls = bindRunnerToolCallRecords(toolCallRecordsFromExecutionLedger(ledgerEvents, runnerCompleted), ctx.trace, ctx.workflow)
			}
			payload := map[string]any{
				"last_error":         firstNonEmpty(strings.TrimSpace(runnerResp.Message), "runner reply delivery became uncertain after a native Slack send attempt"),
				"failure_class":      "runner_reply_delivery_uncertain",
				"failure_summary":    "Runner attempted a native Slack send but did not complete with a durable delivery contract, so the workflow requires human review to avoid duplicate posting.",
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
						Description: "Runner attempted a native Slack reply but delivery could not be finalized safely.",
					},
				},
				"reasoning_steps": []events.ReasoningStep{
					{
						ID:         fmt.Sprintf("reason-reply-delivery-uncertain-%d", runnerCompleted.UnixNano()),
						TraceID:    ctx.trace.Summary.TraceID,
						WorkflowID: ctx.trace.Summary.WorkflowID,
						StepType:   "reply_delivery_uncertain",
						Summary:    "Runner attempted a native Slack send but the final delivery contract was not durable enough to safely retry automatically.",
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
		failure := workflowFailureFromRunnerResponse(runnerResp)
		if !nativeStrictEnvelopeFailure(runnerResp.Raw) {
			if runnerOutput, parseErr := runnerutil.ParseStructuredOutput(runnerResp); parseErr == nil {
				failure.TraceArtifacts = traceArtifactsFromProducedArtifacts(ctx.trace.Summary.TraceID, runnerOutput.ProducedArtifacts)
				if useLedgerFirst {
					if ledgerArtifacts := traceArtifactsFromExecutionLedger(ctx.trace.Summary.TraceID, ledgerEvents); len(ledgerArtifacts) > 0 {
						failure.TraceArtifacts = mergeTraceArtifacts(failure.TraceArtifacts, ledgerArtifacts)
					}
				}
				if len(failure.TraceArtifacts) > 0 {
					completedAt := time.Now().UTC()
					failure.ReasoningSteps = append(failure.ReasoningSteps, events.ReasoningStep{
						ID:         fmt.Sprintf("reason-runner-failed-artifacts-%d", completedAt.UnixNano()),
						TraceID:    ctx.trace.Summary.TraceID,
						WorkflowID: ctx.trace.Summary.WorkflowID,
						StepType:   "artifact_delivery",
						Summary:    fmt.Sprintf("Runner failed after producing %d artifact(s); preserving the workspace artifact records.", len(failure.TraceArtifacts)),
						Confidence: 1.0,
						Decision:   "artifacts_preserved_after_runner_failure",
						CreatedAt:  completedAt,
					})
				}
			}
		}
		return &workflowFailureError{failure: failure}
	}
	runnerOutput, err := workflowRunnerOutput(runnerResp)
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
	traceArtifacts := traceArtifactsFromProducedArtifacts(ctx.trace.Summary.TraceID, runnerOutput.ProducedArtifacts)
	if useLedgerFirst {
		if ledgerArtifacts := traceArtifactsFromExecutionLedger(ctx.trace.Summary.TraceID, ledgerEvents); len(ledgerArtifacts) > 0 {
			traceArtifacts = mergeTraceArtifacts(traceArtifacts, ledgerArtifacts)
		}
	}
	artifactProduced := len(traceArtifacts) > 0
	artifactFailureReason := strings.TrimSpace(runnerOutput.ArtifactFailureReason)
	artifactRequested := artifactProduced || artifactFailureReason != ""

	proposedReplyAction := firstSlackReplyAction(runnerOutput.ProposedActions)
	replyBody := firstNonEmpty(
		strings.TrimSpace(runnerOutput.FinalAnswer),
		strings.TrimSpace(runnerOutput.ReplyDraft),
		strings.TrimSpace(slackReportSummaryFromPayload(proposedReplyAction.RequestPayload)),
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
	_, policyVerdict := replyPolicy(store, ctx.workflow.Kind, replyThreadKey, replyChannelID)
	replyDeliveryMode := clients.NormalizeRunnerReplyDeliveryMode(runnerTask.ReplyDeliveryMode)
	if runnerRawReplyDeliveryMode := clients.NormalizeRunnerReplyDeliveryMode(stringValueFromMap(runnerResp.Raw, "reply_delivery_mode")); runnerRawReplyDeliveryMode == "direct" {
		replyDeliveryMode = runnerRawReplyDeliveryMode
	}
	replyDelivery, hasReplyDelivery := workflowReplyDeliveryProjection(runnerResp.Raw, ledgerEvents, useLedgerFirst, replyChannelID, replyThreadTS, runnerCompleted)
	replyDeliverySucceeded := hasReplyDelivery && events.SlackDeliveryStatusSucceeded(replyDelivery.SendStatus)
	nativeSlackDeliverySucceeded := replyDeliverySucceeded && isRSINativeSlackDelivery(replyDelivery)
	runnerToolCalls := bindRunnerToolCallRecords(
		toolCallRecordsFromRunnerRaw(runnerResp.Raw),
		ctx.trace,
		ctx.workflow,
	)
	if useLedgerFirst {
		runnerToolCalls = bindRunnerToolCallRecords(
			toolCallRecordsFromExecutionLedger(ledgerEvents, runnerCompleted),
			ctx.trace,
			ctx.workflow,
		)
	}
	if !nativeSlackDeliverySucceeded {
		if toolDelivery, ok := workflowReplyDeliveryFromNativeSlackToolCalls(runnerToolCalls, replyChannelID, replyThreadTS, runnerCompleted); ok {
			replyDelivery = toolDelivery
			hasReplyDelivery = true
			replyDeliverySucceeded = events.SlackDeliveryStatusSucceeded(replyDelivery.SendStatus)
			nativeSlackDeliverySucceeded = isRSINativeSlackDelivery(replyDelivery)
		}
	}
	if replyDeliveryMode == "direct" && !nativeSlackDeliverySucceeded {
		replyDeliveryMode = "mediated"
	}
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
	if artifactRequested {
		summary := fmt.Sprintf("Artifact request resulted in %d produced artifact(s).", len(traceArtifacts))
		decision := "artifact_produced"
		if !artifactProduced {
			summary = firstNonEmpty(artifactFailureReason, "Artifact production did not complete.")
			decision = "artifact_missing"
		}
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-artifact-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "artifact_delivery",
			Summary:    summary,
			Confidence: runnerOutput.Confidence,
			Decision:   decision,
			CreatedAt:  runnerCompleted,
		})
	}
	runnerDiagnostics := cloneStringAnyMap(mapValue(runnerResp.Raw["runner_diagnostics"]))
	runnerDiagnostics = mergeWorkflowRunnerDiagnostics(runnerDiagnostics, runnerResp.Raw)
	if artifactFailureReason != "" {
		runnerDiagnostics["artifact_failure_reason"] = artifactFailureReason
	}
	runnerSlackActions := []events.SlackActionRecord{}
	nativeSlackDeliveryRecorded := nativeSlackDeliverySucceeded
	runnerDeliveryRecorded := nativeSlackDeliveryRecorded
	if runnerDeliveryRecorded {
		replyDelivery.TraceID = ctx.trace.Summary.TraceID
		replyDelivery.WorkflowID = ctx.workflow.ID
		replyDelivery.ConversationID = ctx.trace.Summary.ConversationID
		replyDelivery.CaseID = ctx.trace.Summary.CaseID
		replyDelivery.PolicyVerdict = firstNonEmpty(replyDelivery.PolicyVerdict, policyVerdict)
		replyDelivery.ArtifactRefs = uniqueStrings(append(replyDelivery.ArtifactRefs, producedArtifactRefs(runnerOutput.ProducedArtifacts)...))
		replyDelivery.CreatedAt = runnerCompleted
		runnerSlackActions = append(runnerSlackActions, replyDelivery)
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-reply-delivery-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "reply_delivery",
			Summary:    runnerReplyDeliveryReasoningSummary(nativeSlackDeliveryRecorded),
			Confidence: 1.0,
			Decision:   "native_slack_delivery_recorded",
			CreatedAt:  runnerCompleted,
		})
	}
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
	if useLedgerFirst {
		runnerEvents = append(runnerEvents, traceEventsFromExecutionLedger(ctx.trace, ctx.workflow, ledgerEvents, runnerStarted, runnerCompleted)...)
	}
	var (
		replyAction    action.Intent
		draftEvents    []events.TraceEvent
		draftReasoning []events.ReasoningStep
	)
	if strings.TrimSpace(replyBody) != "" && strings.TrimSpace(proposedReplyAction.Kind) == "" && replyDeliveryMode == "mediated" && !nativeSlackDeliveryRecorded {
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-action-contract-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "action_contract_blocked",
			Summary:    "Runner produced a reply draft but did not deliver it through an RSI native Slack delivery tool.",
			Confidence: 0.95,
			Decision:   "needs_human",
			CreatedAt:  runnerCompleted,
		})
		runnerEvents[0].Description = "Runner returned visible reasoning and a reply draft but omitted RSI native Slack delivery."
		if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandWorkflowExecutionNeedsHuman, cfg.ServiceName, runnerCompleted, map[string]any{
			"last_error":         "runner produced a reply without rsi_slack.message_post or rsi_slack.report_post delivery",
			"failure_class":      "missing_explicit_action_contract",
			"failure_summary":    "Runner produced a reply draft but did not deliver it through an RSI native Slack delivery tool.",
			"runner_diagnostics": runnerDiagnostics,
			"repair_attempted":   boolValue(runnerResp.Raw["repair_attempted"]),
			"repair_succeeded":   boolValue(runnerResp.Raw["repair_succeeded"]),
			"trace_events":       runnerEvents,
			"trace_artifacts":    traceArtifacts,
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
	if strings.TrimSpace(proposedReplyAction.Kind) != "" && !runnerDeliveryRecorded {
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-action-contract-disabled-%d", runnerCompleted.UnixNano()),
			TraceID:    ctx.trace.Summary.TraceID,
			WorkflowID: ctx.trace.Summary.WorkflowID,
			StepType:   "action_contract_blocked",
			Summary:    "Runner emitted a legacy proposed Slack action; RSI workflow delivery must go through rsi_slack.message_post or rsi_slack.report_post.",
			Confidence: 0.95,
			Decision:   "needs_human",
			CreatedAt:  runnerCompleted,
		})
		runnerEvents[0].Description = "Runner returned a legacy proposed Slack action; native Slack tool delivery is required."
		if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandWorkflowExecutionNeedsHuman, cfg.ServiceName, runnerCompleted, map[string]any{
			"last_error":         "legacy proposed Slack actions are disabled; use rsi_slack.message_post or rsi_slack.report_post",
			"failure_class":      "legacy_proposed_slack_action_disabled",
			"failure_summary":    "Runner emitted a proposed Slack action instead of delivering through an RSI native Slack tool.",
			"runner_diagnostics": runnerDiagnostics,
			"repair_attempted":   boolValue(runnerResp.Raw["repair_attempted"]),
			"repair_succeeded":   boolValue(runnerResp.Raw["repair_succeeded"]),
			"trace_events":       runnerEvents,
			"trace_artifacts":    traceArtifacts,
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
	finalReasoning = append(finalReasoning, draftReasoning...)
	runnerDescription := "Runner returned visible reasoning."
	hasReplyAction := strings.TrimSpace(replyAction.ID) != ""
	runnerCommand := transition.WorkflowExecutionCompletionCommand(completionVerdict, hasReplyAction)
	if hasReplyAction {
		runnerDescription = "Runner returned visible reasoning and an explicit Slack reply action."
	} else if nativeSlackDeliveryRecorded {
		runnerDescription = "Runner returned visible reasoning and delivered the Slack reply through an RSI native Slack tool."
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
	completionPayload := map[string]any{
		"resume_queue":       string(queueName),
		"reply_action_id":    replyAction.ID,
		"repair_attempted":   boolValue(runnerResp.Raw["repair_attempted"]),
		"repair_succeeded":   boolValue(runnerResp.Raw["repair_succeeded"]),
		"runner_diagnostics": runnerDiagnostics,
		"trace_events":       append(runnerEvents, draftEvents...),
		"trace_artifacts":    traceArtifacts,
		"reasoning_steps":    finalReasoning,
		"tool_calls":         runnerToolCalls,
	}
	if len(runnerSlackActions) > 0 {
		completionPayload["slack_actions"] = runnerSlackActions
	}
	if _, err := submitWorkflowCommand(store, ctx.workflow.ID, runnerCommand, cfg.ServiceName, runnerCompleted, completionPayload); err != nil {
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

func markExternalToolResumeAttemptFinished(store storepkg.Store, pauseID string, runnerResp clients.RunnerResponse) {
	pauseID = strings.TrimSpace(pauseID)
	if pauseID == "" {
		return
	}
	nextStatus := storepkg.ExternalToolResumeResumed
	if !runnerResp.OK {
		nextStatus = storepkg.ExternalToolResumeFailed
	}
	_, _ = store.UpdateExternalToolPause(pauseID, func(item *storepkg.ExternalToolPause) error {
		item.ResumeStatus = nextStatus
		if !runnerResp.OK {
			item.ErrorMessage = firstNonEmpty(strings.TrimSpace(runnerResp.Message), item.ErrorMessage)
		}
		return nil
	})
}

func workflowExecutionID(operationID string, startedAt time.Time) string {
	seed := strings.TrimSpace(operationID)
	if seed == "" {
		seed = startedAt.UTC().Format(time.RFC3339Nano)
	}
	sum := sha1.Sum([]byte(seed))
	return fmt.Sprintf("hexec-%x", sum[:8])
}

func shouldRecoverHermesExecution(effect transition.EffectExecution) bool {
	if effect.StartedAt == nil || effect.UpdatedAt.IsZero() {
		return false
	}
	return !effect.StartedAt.Equal(effect.UpdatedAt)
}

func submitWorkflowContextCompleted(cfg config.Config, store storepkg.Store, ctx workflowContext, resumeQueue queue.QueueName, occurredAt time.Time) error {
	refreshedTrace, ok := store.GetTrace(ctx.trace.Summary.TraceID)
	if ok {
		ctx.trace = refreshedTrace
	}
	if isTerminalTraceStatus(ctx.trace.Summary.Status) {
		return nil
	}
	executionRole := runnerRoleForQueue(resumeQueue)
	executionStartedType := "runner.started"
	executionStartedService := "runner"
	executionStartedDescription := "Runner task dispatched with verbose reasoning enabled."
	contextSummary, contextRefs := contextFromTrace(ctx.trace)
	if len(contextRefs) > 0 {
		if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandContextCompleted, cfg.ServiceName, occurredAt, map[string]any{
			"tool_count":     len(contextRefs),
			"resume_queue":   string(resumeQueue),
			"execution_role": executionRole,
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
					Decision:     fmt.Sprintf("persisted_context_refs:%d", len(contextRefs)),
					CreatedAt:    occurredAt,
				},
			},
		}); err != nil {
			return err
		}
		return nil
	}
	if _, err := submitWorkflowCommand(store, ctx.workflow.ID, transition.CommandContextSkipped, cfg.ServiceName, occurredAt, map[string]any{
		"tool_count":     0,
		"resume_queue":   string(resumeQueue),
		"execution_role": executionRole,
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

func processControlActionEffect(cfg config.Config, store storepkg.Store, effect transition.EffectExecution) error {
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
		if intent.Kind == action.KindSlackPost || intent.Kind == action.KindSlackReport {
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
		if err := executeRemovedToolActionIntent(store, intent, "tool_read actions are not supported by the control plane; route this work through Hermes-native tools instead"); err != nil {
			return handleControlActionExecutionError(cfg, store, effect, ctx, intent, err)
		}
	case action.KindSlackPost:
		if err := executeSlackPostActionIntent(cfg, store, ctx, intent); err != nil {
			return handleControlActionExecutionError(cfg, store, effect, ctx, intent, err)
		}
	case action.KindSlackReport:
		if err := executeSlackReportActionIntent(cfg, store, ctx, intent); err != nil {
			return handleControlActionExecutionError(cfg, store, effect, ctx, intent, err)
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

func handleControlActionExecutionError(cfg config.Config, store storepkg.Store, effect transition.EffectExecution, ctx workflowContext, intent action.Intent, err error) error {
	if err == nil {
		return nil
	}
	if isPostgresActionPersistenceError(err) {
		if finalizeErr := finalizeControlActionPersistenceFailure(cfg, store, effect, ctx, intent, err); finalizeErr != nil {
			_ = failClaimedEffect(store, effect, finalizeErr.Error())
			return fmt.Errorf("finalize control action persistence failure: %w (original error: %v)", finalizeErr, err)
		}
	} else if updatedIntent, ok := store.GetActionIntent(intent.ID); ok {
		_ = maybeAdvanceWorkflowPhaseFromAction(cfg, store, updatedIntent)
	}
	_ = failClaimedEffect(store, effect, err.Error())
	return err
}

func executeRemovedToolActionIntent(store storepkg.Store, intent action.Intent, message string) error {
	started := time.Now().UTC()
	completed := time.Now().UTC()
	commandKind, err := actionCommandForStatus(action.StatusFailed)
	if err != nil {
		return err
	}
	if _, err := submitActionCommand(store, intent.ID, commandKind, "control-plane", completed, map[string]any{
		"operation_id":    intent.OperationID,
		"approval_state":  intent.ApprovalState,
		"executor":        "native-hermes-required",
		"provider":        providerForToolName(intent.TargetRef),
		"error_code":      actionErrorCode(action.StatusFailed),
		"error_message":   message,
		"started_at":      started,
		"completed_at":    completed,
		"summary":         message,
		"tool_call_id":    intent.ID,
		"request_payload": cloneAnyMap(intent.RequestPayload),
	}); err != nil {
		return err
	}
	return fmt.Errorf("%s", message)
}

type slackDeliveryActionOptions struct {
	RecordIDPrefix string
	DefaultSummary string
	Timeout        time.Duration
	Body           func(action.Intent) (draftBody string, finalBody string, deliveryBody string)
	Execute        slackDeliveryNativeExecutor
}

type slackDeliveryNativeExecutor func(context.Context, config.Config, storepkg.Store, workflowContext, action.Intent, time.Time, string, string, string, string) (storepkg.ToolResult, map[string]interface{}, error)

func executeSlackPostActionIntent(cfg config.Config, store storepkg.Store, ctx workflowContext, intent action.Intent) error {
	return executeSlackDeliveryActionIntent(cfg, store, ctx, intent, slackDeliveryActionOptions{
		RecordIDPrefix: "slack-action",
		DefaultSummary: "Slack reply action completed.",
		Timeout:        30 * time.Second,
		Body: func(intent action.Intent) (string, string, string) {
			draftBody := stringFromMap(intent.RequestPayload, "draft_body")
			finalBody := stringFromMap(intent.RequestPayload, "final_body")
			deliveryBody := firstNonEmpty(stringFromMap(intent.RequestPayload, "body"), finalBody, draftBody)
			return firstNonEmpty(draftBody, deliveryBody), firstNonEmpty(finalBody, deliveryBody), deliveryBody
		},
		Execute: executeSlackPostNativeDelivery,
	})
}

func executeSlackReportActionIntent(cfg config.Config, store storepkg.Store, ctx workflowContext, intent action.Intent) error {
	return executeSlackDeliveryActionIntent(cfg, store, ctx, intent, slackDeliveryActionOptions{
		RecordIDPrefix: "slack-report-action",
		DefaultSummary: "Slack report action completed.",
		Timeout:        45 * time.Second,
		Body: func(intent action.Intent) (string, string, string) {
			summaryBody := firstNonEmpty(
				stringFromMap(intent.RequestPayload, "summary"),
				stringFromMap(intent.RequestPayload, "final_body"),
				stringFromMap(intent.RequestPayload, "body"),
			)
			return firstNonEmpty(stringFromMap(intent.RequestPayload, "draft_body"), summaryBody), firstNonEmpty(stringFromMap(intent.RequestPayload, "final_body"), summaryBody), summaryBody
		},
		Execute: executeSlackReportNativeDelivery,
	})
}

func executeSlackDeliveryActionIntent(cfg config.Config, store storepkg.Store, ctx workflowContext, intent action.Intent, opts slackDeliveryActionOptions) error {
	started := time.Now().UTC()
	draftBody, finalBody, deliveryBody := opts.Body(intent)
	channelID := stringFromMap(intent.RequestPayload, "channel_id")
	threadTS := stringFromMap(intent.RequestPayload, "thread_ts")
	blockedReason := firstNonEmpty(stringFromMap(intent.RequestPayload, "blocked_reason"), blockedReasonFromIntent(intent))
	nativeDisabledReason := ""
	if !cfg.NativeToolsEnabled && blockedReason == "" {
		nativeDisabledReason = "native Slack delivery requires native tools to be enabled"
	}
	deliveryIdempotencyKey := slackPostDeliveryIdempotencyKey(ctx, intent, channelID, threadTS, deliveryBody)
	baseRecord := events.SlackActionRecord{
		ID:             fmt.Sprintf("%s-%d", opts.RecordIDPrefix, started.UnixNano()),
		TraceID:        ctx.trace.Summary.TraceID,
		WorkflowID:     ctx.trace.Summary.WorkflowID,
		ConversationID: ctx.trace.Summary.ConversationID,
		CaseID:         ctx.trace.Summary.CaseID,
		ChannelID:      channelID,
		ThreadTS:       threadTS,
		IdempotencyKey: deliveryIdempotencyKey,
		DraftBody:      draftBody,
		FinalBody:      finalBody,
		PolicyVerdict:  firstNonEmpty(intent.PolicyVerdict, blockedReason),
		SendStatus:     "draft_only",
		ArtifactRefs:   append([]string(nil), stringSliceFromMap(intent.RequestPayload, "artifact_refs")...),
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
		result         storepkg.ToolResult
		execErr        error
		actionStatus   action.Status
		summary        string
		renderManifest map[string]interface{}
	)
	if blockedReason != "" {
		actionStatus = action.StatusBlocked
		summary = blockedReason
		baseRecord.SendStatus = blockedReason
	} else if nativeDisabledReason != "" {
		actionStatus = action.StatusFailed
		summary = nativeDisabledReason
		execErr = errors.New(nativeDisabledReason)
		baseRecord.SendStatus = "native_tools_disabled"
	} else {
		timeout := opts.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		nativeCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		result, renderManifest, execErr = opts.Execute(nativeCtx, cfg, store, ctx, intent, started, channelID, threadTS, deliveryIdempotencyKey, deliveryBody)
		actionStatus = actionStatusFromToolResult(result, execErr)
		summary = toolResultSummary(result, execErr)
		baseRecord.ArtifactRefs = uniqueStrings(append(baseRecord.ArtifactRefs, result.RawArtifactRefs...))
		baseRecord.SendStatus = slackSendStatus(actionStatus, result)
	}

	commandKind, err := actionCommandForStatus(actionStatus)
	if err != nil {
		return err
	}
	completedAt := time.Now().UTC()
	payload := map[string]any{
		"operation_id":    intent.OperationID,
		"approval_state":  firstNonEmpty(result.ApprovalState, intent.ApprovalState),
		"policy_verdict":  firstNonEmpty(intent.PolicyVerdict, blockedReason),
		"executor":        firstNonEmpty(result.Provider, "native-hermes-required"),
		"provider":        firstNonEmpty(result.Provider, "slack"),
		"provider_ref":    firstNonEmpty(result.ProviderRef, threadTS),
		"error_code":      actionErrorCode(actionStatus),
		"error_message":   firstNonEmpty(actionErrorMessage(result, execErr), blockedReason),
		"started_at":      started,
		"completed_at":    completedAt,
		"summary":         firstNonEmpty(summary, blockedReason, opts.DefaultSummary),
		"channel_id":      channelID,
		"thread_ts":       threadTS,
		"idempotency_key": deliveryIdempotencyKey,
		"draft_body":      baseRecord.DraftBody,
		"final_body":      baseRecord.FinalBody,
		"send_status":     baseRecord.SendStatus,
		"artifact_refs":   append([]string(nil), baseRecord.ArtifactRefs...),
	}
	if len(renderManifest) > 0 {
		payload["render_manifest"] = renderManifest
	}
	if _, err := submitActionCommand(store, intent.ID, commandKind, cfg.ServiceName, completedAt, payload); err != nil {
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
	if execErr != nil {
		return execErr
	}
	return nil
}

func executeSlackPostNativeDelivery(nativeCtx context.Context, cfg config.Config, store storepkg.Store, ctx workflowContext, intent action.Intent, started time.Time, channelID string, threadTS string, deliveryIdempotencyKey string, deliveryBody string) (storepkg.ToolResult, map[string]interface{}, error) {
	nativeResp, _, nativeErr := handleNativeToolAction(nativeCtx, cfg, store, workflowNativeToolClaims(cfg, ctx, intent, started), nativeToolActionRequest{
		Surface:        "slack",
		Operation:      "message_post",
		TargetRef:      channelID,
		IdempotencyKey: deliveryIdempotencyKey,
		Reason:         firstNonEmpty(intent.Rationale, "workflow final reply"),
		Arguments: map[string]any{
			"channel_id": channelID,
			"thread_ts":  threadTS,
			"text":       deliveryBody,
		},
	})
	execErr := nativeDeliveryError(nativeResp, nativeErr, "native Slack reply failed")
	result := storepkg.ToolResult{
		Name:       "rsi_slack.message_post",
		ToolCallID: nativeResp.Action.ID,
		Approved:   true,
		Status:     nativeToolResultStatus(nativeResp.OK),
		Available:  nativeErr == nil,
		Provider:   "rsi-native-tools",
		ProviderRef: firstNonEmpty(
			nativeResp.Action.SourceRef,
			stringValueFromMap(mapValue(nativeResp.Output), "provider_ref"),
			threadTS,
		),
		ExecutedAt: time.Now().UTC(),
		Input: map[string]interface{}{
			"channel_id": channelID,
			"thread_ts":  threadTS,
		},
		Output: map[string]interface{}{
			"posted":    nativeResp.OK,
			"action_id": nativeResp.Action.ID,
			"source":    nativeResp.Action.SourceRef,
		},
		Summary: firstNonEmpty(nativeResp.Action.ResponseSummary, nativeResp.Error),
		Metadata: map[string]interface{}{
			"external_tool_action_id": nativeResp.Action.ID,
			"mirror_effect":           nativeResp.Action.MirrorEffect,
		},
	}
	return result, nil, execErr
}

func executeSlackReportNativeDelivery(nativeCtx context.Context, cfg config.Config, store storepkg.Store, ctx workflowContext, intent action.Intent, started time.Time, channelID string, threadTS string, deliveryIdempotencyKey string, _ string) (storepkg.ToolResult, map[string]interface{}, error) {
	args := storepkg.CloneJSONMap(intent.RequestPayload)
	if args == nil {
		args = map[string]any{}
	}
	args["channel_id"] = channelID
	args["thread_ts"] = threadTS
	nativeResp, _, nativeErr := handleNativeToolAction(nativeCtx, cfg, store, workflowNativeToolClaims(cfg, ctx, intent, started), nativeToolActionRequest{
		Surface:        "slack",
		Operation:      "report_post",
		TargetRef:      channelID,
		IdempotencyKey: deliveryIdempotencyKey,
		Reason:         firstNonEmpty(intent.Rationale, "workflow final report"),
		Arguments:      args,
	})
	execErr := nativeDeliveryError(nativeResp, nativeErr, "native Slack report failed")
	outputMap := mapValue(nativeResp.Output)
	renderManifest := mapValue(outputMap["render_manifest"])
	artifactRefs := append([]string(nil), stringSliceFromMap(intent.RequestPayload, "artifact_refs")...)
	if nativeResp.Action.ID != "" {
		artifactRefs = append(artifactRefs, "render_manifest:"+nativeResp.Action.ID)
	}
	for _, item := range arrayArg(outputMap, "uploaded_files") {
		if ref := strings.TrimSpace(stringValueFromMap(mapValue(item), "source_ref")); ref != "" {
			artifactRefs = append(artifactRefs, ref)
		}
	}
	result := storepkg.ToolResult{
		Name:       "rsi_slack.report_post",
		ToolCallID: nativeResp.Action.ID,
		Approved:   true,
		Status:     nativeToolResultStatus(nativeResp.OK),
		Available:  nativeErr == nil,
		Provider:   "rsi-native-tools",
		ProviderRef: firstNonEmpty(
			nativeResp.Action.SourceRef,
			stringValueFromMap(outputMap, "provider_ref"),
			threadTS,
		),
		RawArtifactRefs: uniqueStrings(artifactRefs),
		ExecutedAt:      time.Now().UTC(),
		Input: map[string]interface{}{
			"channel_id": channelID,
			"thread_ts":  threadTS,
		},
		Output: map[string]interface{}{
			"posted":          nativeResp.OK,
			"action_id":       nativeResp.Action.ID,
			"source":          nativeResp.Action.SourceRef,
			"render_manifest": renderManifest,
		},
		Summary: firstNonEmpty(nativeResp.Action.ResponseSummary, nativeResp.Error),
		Metadata: map[string]interface{}{
			"external_tool_action_id": nativeResp.Action.ID,
			"mirror_effect":           nativeResp.Action.MirrorEffect,
			"render_manifest":         renderManifest,
		},
	}
	return result, renderManifest, execErr
}

func workflowNativeToolClaims(cfg config.Config, ctx workflowContext, intent action.Intent, started time.Time) nativeToolClaims {
	slackChannelID := strings.TrimSpace(ctx.ingestion.ChannelID)
	slackThreadTS := strings.TrimSpace(ctx.ingestion.ThreadTS)
	slackScope := ""
	if slackChannelID != "" {
		slackScope = "bound_thread"
	}
	return nativeToolClaims{
		Audience:       nativeToolsAudience,
		IssuedAt:       started.Unix(),
		ExpiresAt:      started.Add(cfg.ProdRunnerTaskTimeout + time.Minute).Unix(),
		ExecutionID:    firstNonEmpty(ctx.trace.Summary.TraceID, ctx.workflow.ID),
		OperationID:    firstNonEmpty(intent.OperationID, intent.ID),
		TraceID:        ctx.trace.Summary.TraceID,
		WorkflowID:     ctx.workflow.ID,
		ConversationID: ctx.trace.Summary.ConversationID,
		Actor:          cfg.ServiceName,
		Surfaces:       []string{"slack"},
		SlackChannelID: slackChannelID,
		SlackThreadTS:  slackThreadTS,
		SlackScope:     slackScope,
	}
}

func nativeToolResultStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "failed"
}

func nativeDeliveryError(nativeResp nativeToolActionResponse, nativeErr error, fallback string) error {
	if nativeErr != nil {
		return nativeErr
	}
	if !nativeResp.OK {
		return errors.New(firstNonEmpty(nativeResp.Error, nativeResp.Action.ErrorMessage, fallback))
	}
	return nil
}

func slackPostDeliveryIdempotencyKey(ctx workflowContext, intent action.Intent, channelID string, threadTS string, body string) string {
	if key := firstNonEmpty(intent.IdempotencyKey, intent.ID); key != "" {
		return key
	}
	sum := sha1.Sum([]byte(strings.Join([]string{
		ctx.workflow.ID,
		ctx.trace.Summary.TraceID,
		ctx.trace.Summary.ConversationID,
		channelID,
		threadTS,
		body,
	}, "\x00")))
	return fmt.Sprintf("slack_post:%x", sum)
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
	resolvedIntent := resolveWorkflowIntent(ingestion, contextRefs, recentEntries)
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
		Question:       resolvedIntent.PlanningQuestion,
		ChannelID:      ingestion.ChannelID,
		ThreadTS:       ingestion.ThreadTS,
		EntityRefs:     append([]slackpkg.EntityRef(nil), ingestion.EntityRefs...),
	}, time.Now().UTC())
	hintRefs := liveHintContextRefs(liveHints)
	combinedContextRefs := append(append([]clients.RunnerContextRef{}, contextRefs...), hintRefs...)
	allowed, _ := replyPolicy(store, workflow.Kind, trace.Summary.ThreadKey, ingestion.ChannelID)
	mcpServers := workflowMCPServers(cfg)
	replyDeliveryMode := workflowReplyDeliveryMode(allowed)
	userRequest := resolvedIntent.UserRequest
	systemMessage := harness.ComposeSystemMessage(
		workflowRunnerSystemMessage(replyDeliveryMode),
		effectiveHarness,
	)
	promptParts := []string{
		fmt.Sprintf("User request: %s", userRequest),
	}
	if requesterContext := runnerRequesterContextLine(ingestion); requesterContext != "" {
		promptParts = append(promptParts, requesterContext)
	}
	if hasAttachedBoundSlackThreadContext(contextRefs, ingestion) {
		promptParts = append(promptParts, "Bound Slack thread context is attached in the task evidence. Recover the main request from that thread context before answering, and treat the latest inbound message as a follow-up within that thread.")
	}
	promptParts = append(promptParts, "Investigate with native Hermes tools and the company-computer terminal. Start with the attached persisted evidence and context, then expand with tools as needed. Cite concrete evidence when possible. Default any Slack reply to the ingress thread.")
	prompt := strings.Join(promptParts, "\n\n")
	repo := firstNonEmpty(liveHints.Repo, cfg.DefaultRepo)
	expectedOutputs := []string{"session_title", "visible_reasoning", "final_answer", "produced_artifacts", "artifact_failure_reason"}
	return clients.RunnerTask{
		TaskType:                  "workflow",
		Repo:                      repo,
		RepoRef:                   "main",
		Prompt:                    prompt,
		SystemMessage:             systemMessage,
		MCPServers:                mcpServers,
		AllowedCommands:           []string{},
		ExpectedOutputs:           expectedOutputs,
		ArtifactDestination:       fmt.Sprintf("trace:%s", trace.Summary.TraceID),
		ContextSummary:            contextSummary,
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
		ResponseMode:              workflow.ResponseMode,
		ReplyDeliveryMode:         replyDeliveryMode,
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
		ContractVersion:           clients.RunnerExecutionContractVersion,
		ExecutionIntent: map[string]any{
			"kind":                workflow.Kind,
			"intent":              workflow.Intent,
			"user_request":        userRequest,
			"runner_planner_mode": firstNonEmpty(cfg.RunnerPlannerMode, "runner_first"),
		},
		DeliveryPolicy:  runnerDeliveryPolicy(ingestion.ChannelID, ingestion.ThreadTS, replyDeliveryMode, trace.Summary.TraceID),
		WorkspacePolicy: runnerutil.WorkspacePolicyFromConfig(cfg),
		ApprovalPolicy:  runnerApprovalPolicy(false),
	}
}

func buildSessionTitleRunnerTask(cfg config.Config, store storepkg.Store, role string, ctx workflowContext, rawTitle string) clients.RunnerTask {
	effectiveHarness := harness.ResolveEffectiveConfig(store, role, cfg.DefaultReasoningVerbosity)
	conversationEntries := store.ListConversationEntries(ctx.trace.Summary.ConversationID)
	recentEntries := recentConversationEntries(conversationEntries)
	userRequest := firstNonEmpty(strings.TrimSpace(rawTitle), runnerUserRequest(ctx.ingestion))
	if userRequest == "" {
		userRequest = firstNonEmpty(ctx.workflow.Kind, ctx.workflow.Intent, ctx.trace.Summary.TraceID)
	}
	promptParts := []string{
		"Rewrite the latest Slack request into a concise session title.",
		"Use 3-8 words.",
		"Omit @mentions, channel names, filler phrases, and implementation details that are not part of the ask.",
		"Do not answer the request.",
		"Return only a JSON object. Fill session_title and leave all other structured-output fields empty.",
		fmt.Sprintf("Latest Slack request:\n%s", userRequest),
	}
	if len(recentEntries) > 1 {
		promptParts = append(promptParts, "Recent thread entries are attached only to resolve references such as \"this\" or \"that PR\"; title the latest request, not the whole thread.")
	}
	systemMessage := harness.ComposeSystemMessage(
		"Write compact product-style conversation titles. Return JSON only with session_title as a concise 3-8 word rewrite of the latest user ask. Do not use tools. Do not include @mentions or Slack boilerplate.",
		effectiveHarness,
	)
	return clients.RunnerTask{
		TaskType:                  "general",
		Repo:                      cfg.DefaultRepo,
		RepoRef:                   "main",
		Prompt:                    strings.Join(promptParts, "\n\n"),
		SystemMessage:             systemMessage,
		AllowedCommands:           []string{},
		TimeoutSeconds:            60,
		ExpectedOutputs:           []string{"session_title"},
		ContextSummary:            "Generate a short display title for the latest Slack-triggered RSI session.",
		Intent:                    "session_title",
		TraceID:                   ctx.trace.Summary.TraceID,
		WorkflowID:                ctx.trace.Summary.WorkflowID,
		ConversationID:            ctx.trace.Summary.ConversationID,
		CaseID:                    ctx.trace.Summary.CaseID,
		ChannelID:                 ctx.ingestion.ChannelID,
		ThreadTS:                  ctx.ingestion.ThreadTS,
		TriggerEventID:            ctx.trace.Summary.TriggerEventID,
		RecentConversationEntries: recentEntries,
		RepoAllowlist:             cfg.AllowedTargetRepos,
		ReplyDeliveryMode:         "none",
		ReasoningVerbosity:        effectiveHarness.ReasoningVerbosity,
		SessionScopeKind:          "trace",
		SessionScopeID:            ctx.trace.Summary.TraceID,
		ParentSessionScopeKind:    "conversation",
		ParentSessionScopeID:      ctx.trace.Summary.ConversationID,
		HarnessProfileID:          effectiveHarness.Profile.ID,
		HarnessOverlayVersion:     effectiveHarness.EffectiveOverlayVersion,
		MemoryBackend:             harness.DefaultMemoryBackend,
		AssistantPeerID:           fmt.Sprintf("rsi:%s:%s:title", cfg.Environment, role),
		UserPeerID:                workflowUserPeerID(conversationEntries, "trace", ctx.trace.Summary.TraceID),
		ContractVersion:           clients.RunnerExecutionContractVersion,
		ExecutionIntent: map[string]any{
			"kind":         "session_title",
			"intent":       ctx.workflow.Intent,
			"user_request": userRequest,
			"task_type":    "session_title",
		},
		DeliveryPolicy:  clients.NewRunnerDeliveryPolicy(ctx.ingestion.ChannelID, ctx.ingestion.ThreadTS, "none", strings.Join(nonEmptyStrings(ctx.ingestion.ChannelID, ctx.ingestion.ThreadTS, ctx.trace.Summary.TraceID, "title"), ":")),
		WorkspacePolicy: runnerutil.WorkspacePolicyFromConfig(cfg),
		ApprovalPolicy:  runnerApprovalPolicy(false),
	}
}

func normalizeGeneratedSessionTitle(value string) string {
	value = strings.Trim(strings.TrimSpace(value), "\"'`")
	if value == "" {
		return ""
	}
	return conversation.NormalizeTitle("", strings.Join(strings.Fields(value), " "))
}

func runnerDeliveryPolicy(channelID string, threadTS string, replyDeliveryMode string, traceID string) *clients.RunnerDeliveryPolicy {
	return clients.NewRunnerDeliveryPolicy(channelID, threadTS, replyDeliveryMode, strings.Join(nonEmptyStrings(channelID, threadTS, traceID), ":"))
}

func runnerApprovalPolicy(directSlackAllowed bool) *clients.RunnerApprovalPolicy {
	return clients.NewRunnerApprovalPolicy(directSlackAllowed)
}

func workflowRunnerSystemMessage(replyDeliveryMode string) string {
	parts := []string{
		"Return explicit visible reasoning only. Do not include hidden chain-of-thought.",
		"Produce a JSON object with session_title, visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, reply_delivery, knowledge_drafts, outcome_hypotheses, produced_artifacts, and artifact_failure_reason.",
		"Set session_title to a concise 3-8 word rewrite of the user's Slack question; preserve intent, omit @mentions and filler, and do not summarize your answer.",
		"Treat DeliveryPolicy, WorkspacePolicy, and ApprovalPolicy as execution metadata; infrastructure permissions come from the runner environment.",
		"You may use Hermes-native skills when they materially help satisfy the request.",
	}
	parts = append(parts, "If an artifact materially helps satisfy the request, produce it with Hermes artifact tools and include it in produced_artifacts. If an artifact cannot be produced, set artifact_failure_reason and still provide the best grounded reply.")
	parts = append(parts, "Use Hermes-native tools and terminal CLIs for evidence gathering.", "When native terminal GitHub credentials are available, use the gh CLI for explicitly requested GitHub issue, PR, comment, or review work; do not use it to merge code unless approval is granted.")
	if replyDeliveryMode == "mediated" {
		parts = append(parts, "For the final Slack reply, call exactly one RSI native Slack delivery tool in the bound thread. Use rsi_slack.message_post for simple prose. Use rsi_slack.report_post with report_schema_version=1, summary, sections, and structured tables/files/images for rich reports or tabular output. Do not put Markdown pipe tables in message_post; use report_post tables instead. Do not call non-RSI Slack delivery tools or proposed Slack actions for RSI workflow delivery. Leave proposed_actions empty.")
		parts = append(parts, "Leave reply_delivery empty unless the control plane reports delivery.")
		return strings.Join(parts, " ")
	}
	parts = append(parts, "Slack posting is blocked by policy for this workflow, so do not send any Slack messages.", "Leave reply_delivery empty when no Slack reply was delivered.")
	return strings.Join(parts, " ")
}

func workflowReplyDeliveryMode(replyAllowed bool) string {
	switch {
	case replyAllowed:
		return "mediated"
	default:
		return "none"
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

func workflowCompletionVerdict(raw map[string]any) string {
	if completion := mapValue(mapValue(raw["execution_envelope"])["completion"]); len(completion) > 0 {
		verdict := strings.TrimSpace(stringValueFromMap(completion, "completion_verdict"))
		if verdict != "" {
			return verdict
		}
	}
	verdict := strings.TrimSpace(stringValue(raw["completion_verdict"]))
	if verdict == "" {
		return "complete"
	}
	return verdict
}

func workflowTerminationReason(raw map[string]any) string {
	if completion := mapValue(mapValue(raw["execution_envelope"])["completion"]); len(completion) > 0 {
		reason := strings.TrimSpace(stringValueFromMap(completion, "termination_reason"))
		if reason != "" {
			return reason
		}
	}
	return strings.TrimSpace(stringValue(raw["termination_reason"]))
}

func workflowReplyDelivery(raw map[string]any, fallbackChannelID string, fallbackThreadTS string) (events.SlackActionRecord, bool) {
	var value any
	var ok bool
	if len(mapValue(raw["execution_envelope"])) > 0 {
		deliveries, _ := mapValue(raw["execution_envelope"])["deliveries"].([]any)
		if len(deliveries) > 0 {
			value = deliveries[0]
			ok = true
		}
	}
	if !ok || value == nil {
		value, ok = raw["reply_delivery"]
	}
	if !ok || value == nil {
		value, ok = mapValue(raw["structured_output"])["reply_delivery"]
	}
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
	if len(item) == 0 {
		return events.SlackActionRecord{}, false
	}
	body := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(item, "body")),
		strings.TrimSpace(stringValueFromMap(item, "body_excerpt")),
	)
	if body == "" &&
		strings.TrimSpace(stringValueFromMap(item, "tool_call_id")) == "" &&
		strings.TrimSpace(stringValueFromMap(item, "provider_ref")) == "" &&
		strings.TrimSpace(stringValueFromMap(item, "message_link")) == "" {
		return events.SlackActionRecord{}, false
	}
	record := slackActionRecordFromDeliveryMap(item, fallbackChannelID, fallbackThreadTS, time.Now().UTC())
	if strings.TrimSpace(record.FinalBody) == "" {
		record.DraftBody = body
		record.FinalBody = body
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "tool_call_id")), fmt.Sprintf("runner-reply-%x", sha1.Sum([]byte(body))))
	}
	if strings.TrimSpace(record.SendStatus) == "" {
		if len(mapValue(raw["execution_envelope"])) == 0 {
			record.SendStatus = "posted"
		} else {
			return events.SlackActionRecord{}, false
		}
	}
	if len(mapValue(raw["execution_envelope"])) > 0 && !events.SlackDeliveryStatusSucceeded(record.SendStatus) {
		return record, true
	}
	return record, true
}

func isRSINativeSlackDelivery(record events.SlackActionRecord) bool {
	return isRSINativeSlackToolName(record.ToolName) && events.SlackDeliveryStatusSucceeded(record.SendStatus)
}

func isRSINativeSlackToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case "rsi_slack.message_post", "rsi_slack_message_post", "rsi_slack.report_post", "rsi_slack_report_post":
		return true
	default:
		return false
	}
}

func canonicalRSINativeSlackToolName(name string) string {
	switch strings.TrimSpace(name) {
	case "rsi_slack.message_post", "rsi_slack_message_post":
		return "rsi_slack.message_post"
	case "rsi_slack.report_post", "rsi_slack_report_post":
		return "rsi_slack.report_post"
	default:
		return strings.TrimSpace(name)
	}
}

func runnerReplyDeliveryReasoningSummary(nativeSlack bool) string {
	if nativeSlack {
		return "Runner delivered the final Slack reply through an RSI native Slack delivery tool; preserving it as trace metadata without retrying delivery."
	}
	return "Runner reported Slack delivery through a non-RSI-native path; ignoring it because RSI native Slack delivery is required."
}

func nativeStrictEnvelopeFailure(raw map[string]any) bool {
	if !boolValue(raw["native_strict"]) {
		return false
	}
	if len(mapValue(raw["execution_envelope"])) > 0 {
		return false
	}
	switch strings.TrimSpace(stringValueFromMap(raw, "failure_class")) {
	case "native_workflow_preflight_failed",
		"native_envelope_plugin_unavailable",
		"plugin_execution_envelope_missing",
		"plugin_execution_envelope_mismatch",
		"plugin_execution_envelope_invalid":
		return true
	default:
		return false
	}
}

func workflowRunnerOutput(resp clients.RunnerResponse) (runnerutil.StructuredOutput, error) {
	if nativeStrictEnvelopeSuccess(resp) {
		envelope, ok, err := runnerutil.ParseExecutionEnvelope(resp)
		if err != nil {
			return runnerutil.StructuredOutput{}, err
		}
		if ok {
			return runnerutil.StructuredOutputFromEnvelope(envelope, resp)
		}
	}
	return runnerutil.ParseStructuredOutput(resp)
}

func nativeStrictEnvelopeSuccess(resp clients.RunnerResponse) bool {
	if !resp.OK || !boolValue(resp.Raw["native_strict"]) {
		return false
	}
	if len(mapValue(resp.Raw["execution_envelope"])) == 0 {
		return false
	}
	return strings.TrimSpace(stringValueFromMap(resp.Raw, "failure_class")) == ""
}

func workflowReplyDeliveryProjection(raw map[string]any, ledgerEvents []events.ExecutionLedgerEvent, useLedgerFirst bool, fallbackChannelID string, fallbackThreadTS string, createdAt time.Time) (events.SlackActionRecord, bool) {
	rawDelivery, hasRawDelivery := workflowReplyDelivery(raw, fallbackChannelID, fallbackThreadTS)
	if !useLedgerFirst {
		return rawDelivery, hasRawDelivery
	}
	if ledgerDelivery, ok := workflowReplyDeliveryFromExecutionLedger(ledgerEvents, fallbackChannelID, fallbackThreadTS, createdAt); ok {
		if hasRawDelivery && !slackDeliverySameAttempt(rawDelivery, ledgerDelivery) {
			return rawDelivery, true
		}
		return ledgerDelivery, true
	}
	return rawDelivery, hasRawDelivery
}

func workflowReplyDeliveryFromNativeSlackToolCalls(items []events.ToolCallRecord, fallbackChannelID string, fallbackThreadTS string, createdAt time.Time) (events.SlackActionRecord, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if !isRSINativeSlackToolName(item.ToolName) {
			continue
		}
		record, ok := slackActionRecordFromNativeSlackToolCall(item, fallbackChannelID, fallbackThreadTS, createdAt)
		if ok && isRSINativeSlackDelivery(record) {
			return record, true
		}
	}
	return events.SlackActionRecord{}, false
}

func slackActionRecordFromNativeSlackToolCall(item events.ToolCallRecord, fallbackChannelID string, fallbackThreadTS string, createdAt time.Time) (events.SlackActionRecord, bool) {
	request := mapValue(item.Request)
	if len(request) == 0 {
		return events.SlackActionRecord{}, false
	}
	args := mapValue(request["args"])
	if len(args) == 0 {
		args = mapValue(request["arguments"])
	}
	if len(args) == 0 {
		args = mapValue(request["input"])
	}
	if len(args) == 0 {
		args = request
	}
	result := toolCallResultPayload(request)
	if len(result) == 0 {
		result = mapValue(request["result"])
	}
	response := mapValue(result["output"])
	actionPayload := mapValue(response["action"])
	if len(actionPayload) == 0 {
		actionPayload = mapValue(result["action"])
	}
	toolOutput := mapValue(response["output"])
	if len(toolOutput) == 0 {
		toolOutput = response
	}
	manifest := mapValue(toolOutput["render_manifest"])
	mainMessage := mapValue(manifest["main_message"])

	if !nativeSlackToolCallSucceeded(item, result, response, actionPayload) {
		return events.SlackActionRecord{}, false
	}

	sourceRef := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(actionPayload, "source_ref")),
		strings.TrimSpace(stringValueFromMap(toolOutput, "source_ref")),
		strings.TrimSpace(stringValueFromMap(mainMessage, "source_ref")),
	)
	channelID := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(args, "channel_id")),
		strings.TrimSpace(stringValueFromMap(toolOutput, "channel_id")),
		strings.TrimSpace(stringValueFromMap(mainMessage, "channel_id")),
		fallbackChannelID,
	)
	threadTS := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(args, "thread_ts")),
		strings.TrimSpace(stringValueFromMap(toolOutput, "thread_ts")),
		strings.TrimSpace(stringValueFromMap(mainMessage, "thread_ts")),
		fallbackThreadTS,
	)
	body := firstNonEmpty(
		strings.TrimSpace(slackReportSummaryFromPayload(args)),
		strings.TrimSpace(stringValueFromMap(args, "text")),
		strings.TrimSpace(stringValueFromMap(args, "body")),
		strings.TrimSpace(stringValueFromMap(actionPayload, "response_summary")),
		strings.TrimSpace(stringValueFromMap(result, "summary")),
		strings.TrimSpace(item.Summary),
	)
	toolName := canonicalRSINativeSlackToolName(firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(result, "tool_name")),
		strings.TrimSpace(stringValueFromMap(response, "tool_name")),
		strings.TrimSpace(item.ToolName),
	))
	artifactRefs := append([]string(nil), item.RawArtifactRefs...)
	artifactRefs = append(artifactRefs, stringSliceFromMap(args, "artifact_refs")...)
	artifactRefs = append(artifactRefs, stringSliceFromMap(result, "artifact_refs")...)
	artifactRefs = append(artifactRefs, stringSliceFromMap(response, "artifact_refs")...)
	artifactRefs = append(artifactRefs, stringSliceFromMap(toolOutput, "artifact_refs")...)
	if actionID := strings.TrimSpace(stringValueFromMap(actionPayload, "id")); actionID != "" {
		artifactRefs = append(artifactRefs, "external_tool_action:"+actionID)
	}
	if sourceRef != "" {
		artifactRefs = append(artifactRefs, sourceRef)
	}
	artifactRefs = nativeSlackUploadedFileRefs(artifactRefs, toolOutput["uploaded_files"])
	record := events.SlackActionRecord{
		ID: firstNonEmpty(
			strings.TrimSpace(item.ToolCallID),
			strings.TrimSpace(item.ID),
			strings.TrimSpace(stringValueFromMap(actionPayload, "id")),
		),
		ToolName:  toolName,
		ChannelID: channelID,
		ThreadTS:  threadTS,
		IdempotencyKey: firstNonEmpty(
			strings.TrimSpace(stringValueFromMap(args, "idempotency_key")),
			strings.TrimSpace(stringValueFromMap(actionPayload, "idempotency_key")),
			strings.TrimSpace(item.ToolCallID),
			strings.TrimSpace(item.ID),
		),
		DraftBody:    body,
		FinalBody:    body,
		SendStatus:   "posted",
		ArtifactRefs: uniqueStrings(artifactRefs),
		CreatedAt:    createdAt,
	}
	return record, true
}

func nativeSlackToolCallSucceeded(item events.ToolCallRecord, result map[string]any, response map[string]any, actionPayload map[string]any) bool {
	for _, value := range []string{
		strings.TrimSpace(item.Status),
		strings.TrimSpace(stringValueFromMap(result, "status")),
		strings.TrimSpace(stringValueFromMap(response, "status")),
		strings.TrimSpace(stringValueFromMap(actionPayload, "state")),
	} {
		switch strings.ToLower(value) {
		case "failed", "failure", "error", "blocked":
			return false
		}
	}
	if boolValue(response["ok"]) {
		return true
	}
	for _, value := range []string{
		strings.TrimSpace(item.Status),
		strings.TrimSpace(stringValueFromMap(result, "status")),
		strings.TrimSpace(stringValueFromMap(response, "status")),
		strings.TrimSpace(stringValueFromMap(actionPayload, "state")),
	} {
		switch strings.ToLower(value) {
		case "completed", "posted", "sent", "uploaded", "ok", "success", "succeeded":
			return true
		}
	}
	return false
}

func nativeSlackUploadedFileRefs(refs []string, value any) []string {
	items, ok := value.([]any)
	if !ok {
		return refs
	}
	for _, item := range items {
		upload := mapValue(item)
		if len(upload) == 0 {
			continue
		}
		if sourceRef := strings.TrimSpace(stringValueFromMap(upload, "source_ref")); sourceRef != "" {
			refs = append(refs, sourceRef)
			continue
		}
		if fileID := strings.TrimSpace(stringValueFromMap(upload, "slack_file_id")); fileID != "" {
			refs = append(refs, "slack_file:"+fileID)
		}
	}
	return refs
}

func workflowReplyDeliveryFromExecutionLedger(items []events.ExecutionLedgerEvent, fallbackChannelID string, fallbackThreadTS string, createdAt time.Time) (events.SlackActionRecord, bool) {
	var latestAttempt events.SlackActionRecord
	hasLatestAttempt := false
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if !strings.HasPrefix(strings.TrimSpace(item.Kind), "slack.") {
			continue
		}
		status := firstNonEmpty(strings.TrimSpace(stringValueFromMap(item.Payload, "send_status")), strings.TrimSpace(stringValueFromMap(item.Payload, "status")), strings.TrimSpace(item.Status))
		record := slackActionRecordFromDeliveryMap(item.Payload, fallbackChannelID, fallbackThreadTS, createdAt)
		record.ID = firstNonEmpty(record.ID, item.ID)
		record.IdempotencyKey = firstNonEmpty(record.IdempotencyKey, item.IdempotencyKey, item.ID)
		record.SendStatus = status
		if strings.TrimSpace(record.FinalBody) == "" &&
			strings.TrimSpace(record.DraftBody) == "" &&
			len(record.ArtifactRefs) == 0 &&
			strings.TrimSpace(stringValueFromMap(item.Payload, "provider_ref")) == "" &&
			strings.TrimSpace(stringValueFromMap(item.Payload, "message_link")) == "" {
			continue
		}
		if !events.SlackDeliveryStatusSucceeded(status) {
			if !hasLatestAttempt {
				latestAttempt = record
				hasLatestAttempt = true
			}
			continue
		}
		return record, true
	}
	if hasLatestAttempt {
		return latestAttempt, true
	}
	return events.SlackActionRecord{}, false
}

func slackDeliverySameAttempt(left events.SlackActionRecord, right events.SlackActionRecord) bool {
	if strings.TrimSpace(left.FinalBody) != "" && strings.TrimSpace(right.FinalBody) != "" {
		return strings.TrimSpace(left.FinalBody) == strings.TrimSpace(right.FinalBody)
	}
	if strings.TrimSpace(left.DraftBody) != "" && strings.TrimSpace(right.DraftBody) != "" {
		return strings.TrimSpace(left.DraftBody) == strings.TrimSpace(right.DraftBody)
	}
	if strings.TrimSpace(left.ID) != "" && strings.TrimSpace(right.ID) != "" {
		return strings.TrimSpace(left.ID) == strings.TrimSpace(right.ID)
	}
	if strings.TrimSpace(left.IdempotencyKey) != "" && strings.TrimSpace(right.IdempotencyKey) != "" {
		return strings.TrimSpace(left.IdempotencyKey) == strings.TrimSpace(right.IdempotencyKey)
	}
	return false
}

func slackActionRecordFromDeliveryMap(item map[string]any, fallbackChannelID string, fallbackThreadTS string, createdAt time.Time) events.SlackActionRecord {
	body := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(item, "body")),
		strings.TrimSpace(stringValueFromMap(item, "body_excerpt")),
	)
	status := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(item, "send_status")),
		strings.TrimSpace(stringValueFromMap(item, "status")),
	)
	return events.SlackActionRecord{
		ID:             firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "tool_call_id")), strings.TrimSpace(stringValueFromMap(item, "id")), strings.TrimSpace(stringValueFromMap(item, "delivery_id"))),
		ToolName:       firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "tool_name")), strings.TrimSpace(stringValueFromMap(item, "name"))),
		ChannelID:      firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "channel_id")), fallbackChannelID),
		ThreadTS:       firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "thread_ts")), fallbackThreadTS),
		IdempotencyKey: firstNonEmpty(strings.TrimSpace(stringValueFromMap(item, "idempotency_key")), strings.TrimSpace(stringValueFromMap(item, "body_sha1")), strings.TrimSpace(stringValueFromMap(item, "tool_call_id")), strings.TrimSpace(stringValueFromMap(item, "id"))),
		DraftBody:      body,
		FinalBody:      body,
		SendStatus:     status,
		ArtifactRefs:   append([]string(nil), stringSliceFromMap(item, "artifact_refs")...),
		CreatedAt:      createdAt,
	}
}

func partialCompletionNoticeForTerminationReason(terminationReason string) string {
	switch strings.TrimSpace(terminationReason) {
	case "iteration_budget_exhausted":
		return partialCompletionNoticeIterationBudget
	case "task_timeout":
		return partialCompletionNoticeTaskTimeout
	case "inactivity_timeout":
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
	case "inactivity_timeout":
		return "Runner hit the workflow inactivity limit and returned a best-effort partial response."
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
	case "inactivity_timeout":
		if hasReplyAction {
			return "Runner hit the workflow inactivity limit and returned a partial Slack reply."
		}
		return "Runner hit the workflow inactivity limit and returned a partial completion without a reply side effect."
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
			ID:               item.ID,
			EventID:          item.EventID,
			TraceID:          item.TraceID,
			Source:           string(item.Source),
			SourceEventID:    item.SourceEventID,
			EntryType:        item.EntryType,
			ActorID:          item.ActorID,
			ActorType:        item.ActorType,
			ActorDisplayName: conversationEntryActorDisplayName(item),
			ChannelID:        conversationEntryChannelID(item),
			ThreadTS:         conversationEntryThreadTS(item),
			MessageTS:        conversationEntryMessageTS(item),
			Body:             item.Body,
			CreatedAt:        item.CreatedAt,
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

func resolveWorkflowIntent(ingestion slackpkg.Ingestion, contextRefs []clients.RunnerContextRef, recentEntries []clients.RunnerConversationEntry) resolvedWorkflowIntent {
	userRequest := runnerUserRequest(ingestion)
	contextText := workflowIntentContextText(contextRefs, recentEntries)
	planningQuestion := strings.TrimSpace(userRequest)
	if strings.TrimSpace(contextText) != "" {
		planningQuestion = strings.TrimSpace(strings.Join(nonEmptyStrings(planningQuestion, contextText), "\n\n"))
	}
	return resolvedWorkflowIntent{
		UserRequest:      userRequest,
		PlanningQuestion: firstNonEmpty(planningQuestion, userRequest),
	}
}

func workflowIntentContextText(contextRefs []clients.RunnerContextRef, recentEntries []clients.RunnerConversationEntry) string {
	parts := make([]string, 0, len(contextRefs)+len(recentEntries))
	for _, ref := range contextRefs {
		if ref.Source == "prefetched_slack_thread" || (ref.Kind == "tool_call" && ref.ToolName == "slack.history") {
			parts = append(parts, strings.TrimSpace(ref.Summary))
		}
	}
	for _, entry := range recentEntries {
		if strings.TrimSpace(entry.Body) == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", runnerConversationEntryActorLabel(entry), strings.TrimSpace(entry.Body)))
	}
	return strings.Join(nonEmptyStrings(parts...), "\n")
}

func runnerUserRequest(ingestion slackpkg.Ingestion) string {
	return firstNonEmpty(strings.TrimSpace(ingestion.Prompt.RenderedText), strings.TrimSpace(ingestion.Text))
}

func runnerRequesterContextLine(ingestion slackpkg.Ingestion) string {
	displayName := strings.TrimSpace(ingestion.Prompt.SenderDisplayName)
	userID := strings.TrimSpace(firstNonEmpty(ingestion.Prompt.SenderUserID, ingestion.UserID))
	switch {
	case displayName != "" && userID != "":
		return fmt.Sprintf("Slack requester: %s (%s). If you address the requester by name, use exactly this display name; do not infer or invent a different personal name.", displayName, userID)
	case displayName != "":
		return fmt.Sprintf("Slack requester: %s. If you address the requester by name, use exactly this display name; do not infer or invent a different personal name.", displayName)
	case userID != "":
		return fmt.Sprintf("Slack requester user ID: %s. If you do not know their display name, do not invent one.", userID)
	default:
		return ""
	}
}

func runnerConversationEntryActorLabel(entry clients.RunnerConversationEntry) string {
	displayName := strings.TrimSpace(entry.ActorDisplayName)
	actorID := strings.TrimSpace(entry.ActorID)
	actorType := strings.TrimSpace(entry.ActorType)
	switch {
	case displayName != "" && actorID != "":
		return fmt.Sprintf("%s (%s)", displayName, actorID)
	case displayName != "":
		return displayName
	case actorID != "" && actorType != "":
		return fmt.Sprintf("%s:%s", actorType, actorID)
	case actorType != "":
		return actorType
	case actorID != "":
		return actorID
	default:
		return "participant"
	}
}

func conversationEntryActorDisplayName(item conversation.Entry) string {
	metadata := mapValue(item.Metadata)
	prompt := slackpkg.PromptEnvelopeFromValue(metadata["prompt_envelope"])
	if prompt.SenderDisplayName != "" {
		return prompt.SenderDisplayName
	}
	if displayName := stringValueFromMap(metadata, "sender_display_name"); displayName != "" {
		return displayName
	}
	actorID := strings.ToUpper(strings.TrimSpace(firstNonEmpty(item.ActorID, stringValueFromMap(metadata, "user_id"), prompt.SenderUserID)))
	if actorID == "" {
		return ""
	}
	names := stringMapValue(metadata["slack_user_names"])
	if displayName := strings.TrimSpace(names[actorID]); displayName != "" {
		return displayName
	}
	if displayName := strings.TrimSpace(names[strings.ToUpper(strings.TrimSpace(item.ActorID))]); displayName != "" {
		return displayName
	}
	return ""
}

func conversationEntryChannelID(item conversation.Entry) string {
	metadata := mapValue(item.Metadata)
	prompt := slackpkg.PromptEnvelopeFromValue(metadata["prompt_envelope"])
	return firstNonEmpty(stringValueFromMap(metadata, "channel_id"), prompt.ChannelID)
}

func conversationEntryThreadTS(item conversation.Entry) string {
	metadata := mapValue(item.Metadata)
	prompt := slackpkg.PromptEnvelopeFromValue(metadata["prompt_envelope"])
	return firstNonEmpty(stringValueFromMap(metadata, "thread_ts"), prompt.ThreadTS)
}

func conversationEntryMessageTS(item conversation.Entry) string {
	metadata := mapValue(item.Metadata)
	for _, candidate := range []string{
		stringValueFromMap(metadata, "message_ts"),
		stringValueFromMap(metadata, "ts"),
		stringValueFromMap(metadata, "event_ts"),
		item.SourceEventID,
	} {
		if looksLikeSlackTimestamp(candidate) {
			return strings.TrimSpace(candidate)
		}
	}
	return ""
}

func stringMapValue(value any) map[string]string {
	out := map[string]string{}
	switch typed := value.(type) {
	case map[string]string:
		for key, val := range typed {
			if strings.TrimSpace(key) != "" && strings.TrimSpace(val) != "" {
				out[strings.ToUpper(strings.TrimSpace(key))] = strings.TrimSpace(val)
			}
		}
	case map[string]any:
		for key, val := range typed {
			if text := stringValue(val); strings.TrimSpace(key) != "" && text != "" {
				out[strings.ToUpper(strings.TrimSpace(key))] = text
			}
		}
	}
	return out
}

func looksLikeSlackTimestamp(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, ".") {
		return false
	}
	for _, ch := range value {
		if ch == '.' {
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func shouldAttachBoundSlackThreadContext(ingestion slackpkg.Ingestion) bool {
	return strings.TrimSpace(ingestion.ChannelID) != "" && strings.TrimSpace(ingestion.ThreadTS) != ""
}

func hasAttachedBoundSlackThreadContext(contextRefs []clients.RunnerContextRef, ingestion slackpkg.Ingestion) bool {
	if !shouldAttachBoundSlackThreadContext(ingestion) {
		return false
	}
	for _, ref := range contextRefs {
		if ref.ToolName == "slack.history" && ref.ChannelID == ingestion.ChannelID && ref.ThreadTS == ingestion.ThreadTS {
			return true
		}
	}
	return false
}

func prefetchBoundSlackThreadContext(cfg config.Config, trace events.TraceSummary, workflow storepkg.Workflow, ingestion slackpkg.Ingestion) (string, []clients.RunnerContextRef) {
	return "", nil
}

func compactWhitespace(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func producedArtifactRefs(items []runnerutil.ProducedArtifact) []string {
	out := []string{}
	for _, item := range items {
		out = append(out, item.ArtifactRefs...)
		if strings.TrimSpace(item.FileRef) != "" {
			out = append(out, strings.TrimSpace(item.FileRef))
		}
	}
	return uniqueStrings(out)
}

func traceArtifactsFromProducedArtifacts(traceID string, items []runnerutil.ProducedArtifact) []events.Artifact {
	if strings.TrimSpace(traceID) == "" || len(items) == 0 {
		return nil
	}
	out := make([]events.Artifact, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		refs := append([]string{}, item.ArtifactRefs...)
		if strings.TrimSpace(item.FileRef) != "" {
			refs = append(refs, item.FileRef)
		}
		for _, ref := range refs {
			ref = strings.TrimSpace(ref)
			if ref == "" {
				continue
			}
			key := fmt.Sprintf("%s:%s", strings.TrimSpace(item.Kind), ref)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			sum := sha1.Sum([]byte(strings.Join([]string{traceID, strings.TrimSpace(item.Kind), ref}, "|")))
			out = append(out, events.Artifact{
				ID:          fmt.Sprintf("artifact-workflow-%x", sum),
				TraceID:     traceID,
				Kind:        firstNonEmpty(strings.TrimSpace(item.Kind), "workflow_artifact"),
				ContentType: "",
				URL:         ref,
				SizeBytes:   item.SizeBytes,
				Source:      "runner",
			})
		}
	}
	return out
}

func mergeTraceArtifacts(groups ...[]events.Artifact) []events.Artifact {
	out := make([]events.Artifact, 0)
	seen := map[string]struct{}{}
	for _, group := range groups {
		for _, item := range group {
			key := storepkg.TraceArtifactDedupKey(item)
			if key == "" {
				out = append(out, item)
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}

func workflowMCPServers(cfg config.Config) []clients.RunnerMCPServer {
	_ = cfg
	return nil
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

func firstSlackReportAction(items []runnerutil.ProposedAction) runnerutil.ProposedAction {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Kind), string(action.KindSlackReport)) {
			return item
		}
	}
	return runnerutil.ProposedAction{}
}

func firstSlackReplyAction(items []runnerutil.ProposedAction) runnerutil.ProposedAction {
	if report := firstSlackReportAction(items); strings.TrimSpace(report.Kind) != "" {
		return report
	}
	return firstSlackPostAction(items)
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

func nonEmptyStrings(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
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

func int64Value(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case uint:
		return int64(typed)
	case uint8:
		return int64(typed)
	case uint16:
		return int64(typed)
	case uint32:
		return int64(typed)
	case uint64:
		if typed > uint64(^uint64(0)>>1) {
			return 0
		}
		return int64(typed)
	case float32:
		return int64(typed)
	case float64:
		return int64(typed)
	default:
		return 0
	}
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

func firstNonZeroFloat(items ...float64) float64 {
	for _, item := range items {
		if item != 0 {
			return item
		}
	}
	return 0
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

func claimNextExecutionEffect(cfg config.Config, store storepkg.Store, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	if cfg.EffectFairClaimEnabled {
		return store.ClaimNextEffectExecutionForKinds(
			holder,
			lease,
			[]string{string(queue.WorkflowQueue), string(queue.ProactiveQueue)},
			cfg.EffectMaxConcurrentPerScope,
			[]storepkg.EffectClaimSelector{
				{MachineKind: transition.MachineWorkflow, EffectKind: transition.EffectInvokeRunner},
				{MachineKind: transition.MachineWorkflow, EffectKind: transition.EffectSummarizeSessionTitle},
			},
		)
	}
	for _, effect := range store.ListEffectExecutions() {
		switch {
		case effect.MachineKind == transition.MachineWorkflow && effect.EffectKind == transition.EffectInvokeRunner:
		case effect.MachineKind == transition.MachineWorkflow && effect.EffectKind == transition.EffectSummarizeSessionTitle:
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

func claimNextActionEffect(cfg config.Config, store storepkg.Store, ownerPlane string, holder string, lease time.Duration) (transition.EffectExecution, bool, error) {
	ownerPlane = strings.TrimSpace(ownerPlane)
	if cfg.EffectFairClaimEnabled {
		selector := storepkg.EffectClaimSelector{MachineKind: transition.MachineAction, EffectKind: transition.EffectInvokeAction}
		if ownerPlane != "" {
			selector.PayloadEquals = map[string]string{"owner_plane": ownerPlane}
		}
		claimed, ok, err := store.ClaimNextEffectExecutionForKinds(
			holder,
			lease,
			[]string{"action"},
			cfg.EffectMaxConcurrentPerScope,
			[]storepkg.EffectClaimSelector{selector},
		)
		if err != nil || !ok {
			return claimed, ok, err
		}
		return claimed, true, nil
	}
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
	runtime := newWorkflowRuntimeCoordinator(config.Config{}, store)
	return runtime.completeClaimedEffect(effect, resultRef)
}

func deferClaimedEffect(store storepkg.Store, effect transition.EffectExecution, lease time.Duration, reason string) error {
	if strings.TrimSpace(effect.ID) == "" {
		return nil
	}
	runtime := newWorkflowRuntimeCoordinator(config.Config{}, store)
	return runtime.deferClaimedEffect(effect, lease, reason)
}

func failClaimedEffect(store storepkg.Store, effect transition.EffectExecution, lastError string) error {
	if strings.TrimSpace(effect.ID) == "" || effect.Status == transition.EffectFailed {
		return nil
	}
	runtime := newWorkflowRuntimeCoordinator(config.Config{}, store)
	return runtime.failClaimedEffect(effect, lastError)
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
	if intent.Kind == action.KindSlackPost || intent.Kind == action.KindSlackReport {
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
	cloned := storepkg.CloneJSONMap(input)
	if cloned == nil {
		return map[string]any{}
	}
	return cloned
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
	refs := make([]clients.RunnerContextRef, 0, len(hints.CandidateReadSurfaces))
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
		"action_contract_repair_attempts",
		"action_contract_repair_error",
		"action_contract_repair_errors",
		"action_contract_repair_response",
		"action_contract_repair_responses",
	} {
		if value, ok := raw[key]; ok && value != nil {
			diagnostics[key] = value
		}
	}
	return diagnostics
}

func toolCallRecordsFromRunnerRaw(raw map[string]any) []events.ToolCallRecord {
	if records := toolCallRecordsFromExecutionEnvelope(raw); len(records) > 0 {
		return records
	}
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

func traceEventsFromExecutionLedger(trace events.Trace, workflow storepkg.Workflow, items []events.ExecutionLedgerEvent, startedAt time.Time, completedAt time.Time) []events.TraceEvent {
	out := make([]events.TraceEvent, 0)
	for _, item := range items {
		kind := strings.TrimSpace(item.Kind)
		if kind == "" {
			continue
		}
		if !strings.HasPrefix(kind, "phase.") && !strings.HasPrefix(kind, "failure.") && !strings.HasPrefix(kind, "slack.") {
			continue
		}
		status := events.StatusCompleted
		switch strings.ToLower(strings.TrimSpace(item.Status)) {
		case "failed", "error":
			status = events.StatusFailed
		case "needs-human", "needs_human", "uncertain":
			status = events.StatusNeedsHuman
		case "running":
			status = events.StatusRunning
		case "skipped":
			status = events.StatusSuppressed
		case "planned":
			status = events.StatusQueued
		}
		recordedAt := item.RecordedAt
		if recordedAt.IsZero() {
			recordedAt = completedAt
		}
		eventStarted := startedAt
		if item.RecordedAt.After(startedAt) {
			eventStarted = item.RecordedAt
		}
		if recordedAt.Before(eventStarted) {
			recordedAt = eventStarted
		}
		out = append(out, events.TraceEvent{
			TraceID:        trace.Summary.TraceID,
			IngestionID:    trace.Summary.IngestionID,
			WorkflowID:     trace.Summary.WorkflowID,
			ConversationID: trace.Summary.ConversationID,
			CaseID:         trace.Summary.CaseID,
			Plane:          "execution",
			Service:        "runner-ledger",
			Actor:          workflow.AssignedBot,
			EventType:      "ledger." + kind,
			Status:         status,
			StartedAt:      eventStarted,
			EndedAt:        timeutil.PtrTime(recordedAt),
			Description:    ledgerEventDescription(item),
		})
	}
	return out
}

func ledgerEventDescription(item events.ExecutionLedgerEvent) string {
	if summary := strings.TrimSpace(stringValueFromMap(item.Payload, "summary")); summary != "" {
		return summary
	}
	if reason := strings.TrimSpace(stringValueFromMap(item.Payload, "failure_reason")); reason != "" {
		return reason
	}
	if reason := strings.TrimSpace(stringValueFromMap(item.Payload, "reason")); reason != "" {
		return reason
	}
	if phaseType := strings.TrimSpace(stringValueFromMap(item.Payload, "phase_type")); phaseType != "" {
		return fmt.Sprintf("%s %s.", phaseType, strings.TrimSpace(item.Status))
	}
	return strings.TrimSpace(item.Kind)
}

func traceArtifactsFromExecutionLedger(traceID string, items []events.ExecutionLedgerEvent) []events.Artifact {
	if strings.TrimSpace(traceID) == "" || len(items) == 0 {
		return nil
	}
	produced := make([]runnerutil.ProducedArtifact, 0)
	for _, item := range items {
		if item.Kind != "artifact.created" && item.Kind != "artifact.manifest" && item.Kind != "artifact.rendered" && item.Kind != "artifact.file.written" && item.Kind != "file.written" {
			continue
		}
		payload := item.Payload
		refs := stringSliceFromMap(payload, "artifact_refs")
		fileRef := strings.TrimSpace(stringValueFromMap(payload, "file_ref"))
		workspacePath := strings.TrimSpace(stringValueFromMap(payload, "workspace_path"))
		sourcePath := strings.TrimSpace(stringValueFromMap(payload, "source_path"))
		renderedPath := strings.TrimSpace(stringValueFromMap(payload, "rendered_path"))
		previewPath := strings.TrimSpace(stringValueFromMap(payload, "preview_path"))
		url := strings.TrimSpace(stringValueFromMap(payload, "url"))
		if len(refs) == 0 && fileRef != "" {
			refs = []string{fileRef}
		}
		for _, path := range []string{renderedPath, previewPath, sourcePath, url} {
			if strings.TrimSpace(path) == "" {
				continue
			}
			refs = append(refs, path)
		}
		if len(refs) == 0 && workspacePath != "" {
			refs = []string{"file://" + workspacePath}
		}
		if len(refs) == 0 {
			continue
		}
		produced = append(produced, runnerutil.ProducedArtifact{
			Kind:          firstNonEmpty(strings.TrimSpace(stringValueFromMap(payload, "kind")), "workflow_artifact"),
			ArtifactRefs:  refs,
			FileRef:       fileRef,
			SizeBytes:     int64Value(payload["size_bytes"]),
			SHA256:        strings.TrimSpace(stringValueFromMap(payload, "sha256")),
			ShareStatus:   strings.TrimSpace(stringValueFromMap(payload, "share_status")),
			WorkspacePath: firstNonEmpty(workspacePath, renderedPath, previewPath, sourcePath),
		})
	}
	return traceArtifactsFromProducedArtifacts(traceID, produced)
}

func toolCallRecordsFromExecutionLedger(items []events.ExecutionLedgerEvent, createdAt time.Time) []events.ToolCallRecord {
	out := make([]events.ToolCallRecord, 0)
	seen := map[string]int{}
	for index, item := range items {
		recordedAt := item.RecordedAt
		if recordedAt.IsZero() {
			recordedAt = createdAt
		}
		record, ok := toolCallRecordFromLedgerFields(
			strings.TrimSpace(item.ID),
			strings.TrimSpace(item.Kind),
			strings.TrimSpace(item.Status),
			item.Payload,
			recordedAt,
			index,
		)
		if !ok {
			continue
		}
		out = appendOrReplaceToolCallRecord(out, seen, record)
	}
	return out
}

func toolCallRecordsFromExecutionEnvelope(raw map[string]any) []events.ToolCallRecord {
	envelope := mapValue(raw["execution_envelope"])
	if len(envelope) == 0 {
		return nil
	}
	items, _ := envelope["ledger_events"].([]any)
	if len(items) == 0 {
		return nil
	}
	out := make([]events.ToolCallRecord, 0)
	seen := map[string]int{}
	for index, value := range items {
		item := mapValue(value)
		kind := strings.TrimSpace(stringValueFromMap(item, "kind"))
		payload := mapValue(item["payload"])
		record, ok := toolCallRecordFromLedgerFields(
			strings.TrimSpace(stringValueFromMap(item, "event_id")),
			kind,
			strings.TrimSpace(stringValueFromMap(item, "status")),
			payload,
			time.Now().UTC(),
			index,
		)
		if !ok {
			continue
		}
		out = appendOrReplaceToolCallRecord(out, seen, record)
	}
	return out
}

func toolCallRecordFromLedgerFields(id string, kind string, status string, payload map[string]any, createdAt time.Time, index int) (events.ToolCallRecord, bool) {
	if !isProjectedToolCallLedgerKind(kind) {
		return events.ToolCallRecord{}, false
	}
	result := toolCallResultPayload(payload)
	toolName := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(payload, "tool_name")),
		strings.TrimSpace(stringValueFromMap(payload, "name")),
		strings.TrimSpace(stringValueFromMap(result, "transport_tool_name")),
		strings.TrimSpace(stringValueFromMap(result, "tool_name")),
	)
	toolCallID := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(payload, "tool_call_id")),
		strings.TrimSpace(stringValueFromMap(payload, "call_id")),
		strings.TrimSpace(id),
	)
	if toolName == "" && toolCallID == "" {
		return events.ToolCallRecord{}, false
	}
	summary := firstNonEmpty(
		strings.TrimSpace(stringValueFromMap(result, "summary")),
		strings.TrimSpace(stringValueFromMap(payload, "summary")),
		strings.TrimSpace(kind),
	)
	rawArtifactRefs := stringSliceFromMap(result, "raw_artifact_refs")
	if len(rawArtifactRefs) == 0 {
		rawArtifactRefs = stringSliceFromMap(payload, "raw_artifact_refs")
	}
	if len(rawArtifactRefs) == 0 {
		rawArtifactRefs = stringSliceFromMap(payload, "artifact_refs")
	}
	return events.ToolCallRecord{
		ID:                    firstNonEmpty(strings.TrimSpace(id), fmt.Sprintf("runner-ledger-tool-%d", index)),
		ToolName:              toolName,
		ToolCallID:            toolCallID,
		Request:               payload,
		Summary:               summary,
		RawArtifactRefs:       rawArtifactRefs,
		ApprovalState:         firstNonEmpty(strings.TrimSpace(stringValueFromMap(result, "approval_state")), strings.TrimSpace(stringValueFromMap(payload, "approval_state"))),
		InterpretationSummary: summary,
		Status:                projectedToolCallStatus(status, strings.TrimSpace(stringValueFromMap(result, "status"))),
		CreatedAt:             createdAt,
	}, true
}

func isProjectedToolCallLedgerKind(kind string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "tool.call.completed", "tool.call.failed", "tool.call.error", "tool.tool_call_completed", "tool_call_completed":
		return true
	default:
		return false
	}
}

func toolCallResultPayload(payload map[string]any) map[string]any {
	if len(payload) == 0 {
		return nil
	}
	switch result := payload["result"].(type) {
	case map[string]any:
		return cloneStringAnyMap(result)
	case string:
		var parsed map[string]any
		if err := json.Unmarshal([]byte(result), &parsed); err == nil {
			return parsed
		}
	}
	return nil
}

func projectedToolCallStatus(eventStatus string, resultStatus string) string {
	switch strings.ToLower(strings.TrimSpace(resultStatus)) {
	case "failed", "failure", "error", "blocked":
		return "failed"
	case "ok", "success", "succeeded", "completed":
		return "completed"
	}
	switch strings.ToLower(strings.TrimSpace(eventStatus)) {
	case "ok", "success", "succeeded", "completed":
		return "completed"
	case "failed", "failure", "error", "blocked":
		return "failed"
	default:
		return strings.TrimSpace(eventStatus)
	}
}

func appendOrReplaceToolCallRecord(records []events.ToolCallRecord, seen map[string]int, record events.ToolCallRecord) []events.ToolCallRecord {
	key := firstNonEmpty(record.ToolCallID, record.ID)
	if key == "" {
		return append(records, record)
	}
	if existing, ok := seen[key]; ok {
		records[existing] = record
		return records
	}
	seen[key] = len(records)
	return append(records, record)
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
