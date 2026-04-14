package improvementplane

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/githubapp"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

var errDeferredWorkItem = errors.New("work item deferred for retry")

func RunWorker(cfg config.Config, store storepkg.Store) error {
	workerID := fmt.Sprintf("%s-worker", cfg.ServiceName)
	runnerClients := map[string]*clients.RunnerClient{
		"eval":     clients.NewRunnerClientWithTimeout(cfg.RunnerURLForRole("eval"), cfg.RunnerTimeoutForRole("eval")),
		"proposal": clients.NewRunnerClientWithTimeout(cfg.RunnerURLForRole("proposal"), cfg.RunnerTimeoutForRole("proposal")),
	}
	toolClient := clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL)
	launcher, launcherErr := sandbox.NewLauncher(cfg)
	for {
		item, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.EvalQueue, queue.ProposalQueue, queue.SandboxQueue, queue.ImprovementActionQueue}, workerID, cfg.WorkItemLeaseDuration)
		if err != nil {
			return err
		}
		if !ok {
			time.Sleep(cfg.WorkerPollInterval)
			continue
		}
		if err := processImprovementItem(cfg, store, runnerClients, toolClient, launcher, launcherErr, item); err != nil {
			if errors.Is(err, errDeferredWorkItem) {
				continue
			}
			log.Printf("improvement-plane worker item=%s error=%v", item.ID, err)
			_, _ = store.FailWorkItem(item.ID, err.Error())
			continue
		}
		_, _ = store.CompleteWorkItem(item.ID)
	}
}

func processImprovementItem(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, toolClient *clients.ToolGatewayClient, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	switch item.Queue {
	case queue.EvalQueue:
		return processEvalItem(cfg, store, runnerClients["eval"], item)
	case queue.ProposalQueue:
		return processProposalItem(cfg, store, runnerClients["proposal"], item)
	case queue.SandboxQueue:
		return processSandboxItem(cfg, store, launcher, launcherErr, item)
	case queue.ImprovementActionQueue:
		return processImprovementActionItem(cfg, store, toolClient, item)
	default:
		return fmt.Errorf("unsupported improvement work queue %s", item.Queue)
	}
}

func processEvalItem(cfg config.Config, store storepkg.Store, runnerClient *clients.RunnerClient, item queue.WorkItem) error {
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	started := time.Now().UTC()
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "eval.started",
				Status:      events.StatusRunning,
				StartedAt:   started,
				Description: fmt.Sprintf("Started eval item kind=%s.", item.Kind),
			},
		},
	})
	run, judgments, err := store.EvaluateTrace(item.TraceID, item.Kind)
	if err != nil {
		return err
	}
	runnerStarted := time.Now().UTC()
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       "eval",
				EventType:   "runner.started",
				Status:      events.StatusRunning,
				StartedAt:   runnerStarted,
				Description: "Eval summary dispatched to eval runner.",
			},
		},
	})
	var (
		runnerResp   clients.RunnerResponse
		runnerOutput runnerutil.StructuredOutput
		runnerErr    error
	)
	if runnerClient != nil {
		runnerResp, runnerErr = runnerClient.Execute(buildEvalRunnerTask(cfg, store, trace, run, judgments, item))
		if runnerErr == nil && !runnerResp.OK {
			runnerErr = fmt.Errorf("eval runner returned non-ok result: %s", strings.TrimSpace(runnerResp.Message))
		}
		if runnerErr == nil {
			if err := runnerutil.PersistHarnessExecution(
				store,
				runnerResp,
				"eval",
				item.OperationID,
				trace.Summary.TraceID,
				"",
				stringFromAny(runnerResp.Raw["harness_profile_id"]),
				stringFromAny(runnerResp.Raw["effective_overlay_version"]),
				stringFromAny(runnerResp.Raw["session_scope_kind"]),
				stringFromAny(runnerResp.Raw["session_scope_id"]),
				stringFromAny(runnerResp.Raw["parent_session_scope_kind"]),
				stringFromAny(runnerResp.Raw["parent_session_scope_id"]),
			); err != nil {
				return err
			}
			runnerOutput = runnerutil.ParseStructuredOutput(runnerResp)
		}
	}
	completed := time.Now().UTC()
	if runnerErr == nil {
		if err := persistImprovementKnowledgeDrafts(store, runnerOutput.KnowledgeDrafts, trace, "", completed); err != nil {
			return err
		}
	}
	reasoning := []events.ReasoningStep{
		{
			ID:         fmt.Sprintf("reason-eval-%d", completed.UnixNano()),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "eval_summary",
			Summary:    firstNonEmpty(runnerOutput.FinalAnswer, runnerOutput.ContextSummary, fmt.Sprintf("Recorded %d judgments with overall score %.2f.", len(judgments), run.OverallScore)),
			Confidence: confidenceOr(run.OverallScore, runnerOutput.Confidence),
			Decision:   run.OverallVerdict,
			CreatedAt:  completed,
		},
	}
	if runnerErr == nil {
		reasoning = append(runnerutil.ToTraceReasoning(trace.Summary.TraceID, trace.Summary.WorkflowID, runnerOutput, completed), reasoning...)
		reasoning = append(reasoning, improvementOutcomeHypothesisReasoning(trace, runnerOutput.OutcomeHypotheses, completed)...)
		if strings.TrimSpace(runnerOutput.SelfCritique) != "" {
			reasoning = append(reasoning, events.ReasoningStep{
				ID:         fmt.Sprintf("reason-eval-self-%d", completed.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "self_critique",
				Summary:    runnerOutput.SelfCritique,
				Confidence: confidenceOr(run.OverallScore, runnerOutput.Confidence),
				CreatedAt:  completed,
			})
		}
	}
	runnerEvent := events.TraceEvent{
		TraceID:     trace.Summary.TraceID,
		IngestionID: trace.Summary.IngestionID,
		WorkflowID:  trace.Summary.WorkflowID,
		Plane:       "execution",
		Service:     "runner",
		Actor:       "eval",
		EventType:   "runner.completed",
		Status:      events.StatusCompleted,
		StartedAt:   runnerStarted,
		EndedAt:     ptrTime(completed),
		Description: fmt.Sprintf("Eval runner returned summary using %s.", runnerRuntimeLabel(runnerResp)),
	}
	if runnerErr != nil {
		runnerEvent.EventType = "runner.failed"
		runnerEvent.Status = events.StatusFailed
		runnerEvent.Description = fmt.Sprintf("Eval runner unavailable; kept deterministic results only: %v", runnerErr)
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "eval.completed",
				Status:      events.StatusCompleted,
				StartedAt:   started,
				EndedAt:     ptrTime(completed),
				Description: fmt.Sprintf("Eval %s completed with verdict %s.", run.ID, run.OverallVerdict),
			},
			runnerEvent,
		},
		Reasoning: reasoning,
	})
	if err := ensureApprovedProposalWork(store, trace, cfg.ServiceName); err != nil {
		return err
	}
	return nil
}

func processProposalItem(cfg config.Config, store storepkg.Store, runnerClient *clients.RunnerClient, item queue.WorkItem) error {
	if item.ProposalID == "" {
		return fmt.Errorf("proposal work item %s missing proposal_id", item.ID)
	}
	proposal, ok := findProposal(store.ListProposals(), item.ProposalID)
	if !ok {
		return fmt.Errorf("proposal %s not found", item.ProposalID)
	}
	if proposal.Status != review.ProposalApproved {
		return nil
	}
	proposalTraceID := item.TraceID
	if proposalTraceID == "" {
		proposalTraceID = proposal.TraceID
	}
	trace, ok := store.GetTrace(proposalTraceID)
	if !ok {
		return nil
	}
	attempt, trace, err := ensureProposalAttempt(cfg, store, proposal, trace, item)
	if err != nil {
		return err
	}
	memories := filterProposalMemory(store.ListProposalMemories(), proposal.CandidateKey)
	runnerStarted := time.Now().UTC()
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       "proposal",
				EventType:   "runner.started",
				Status:      events.StatusRunning,
				StartedAt:   runnerStarted,
				Description: fmt.Sprintf("Change attempt %s dispatched to proposal runner.", attempt.ID),
			},
		},
	})
	var (
		runnerResp   clients.RunnerResponse
		runnerOutput runnerutil.StructuredOutput
		runnerErr    error
	)
	if runnerClient != nil {
		runnerTask := buildProposalRunnerTask(cfg, store, trace, proposal, attempt, memories)
		runnerResp, runnerErr = runnerClient.Execute(runnerTask)
		if runnerErr == nil && !runnerResp.OK {
			runnerErr = fmt.Errorf("proposal runner returned non-ok result: %s", strings.TrimSpace(runnerResp.Message))
		}
		if runnerErr == nil {
			if err := runnerutil.PersistHarnessExecution(
				store,
				runnerResp,
				"proposal",
				item.OperationID,
				trace.Summary.TraceID,
				proposal.ID,
				runnerTask.HarnessProfileID,
				runnerTask.HarnessOverlayVersion,
				runnerTask.SessionScopeKind,
				runnerTask.SessionScopeID,
				runnerTask.ParentSessionScopeKind,
				runnerTask.ParentSessionScopeID,
			); err != nil {
				return err
			}
			runnerOutput = runnerutil.ParseStructuredOutput(runnerResp)
		}
	}
	if runnerErr != nil {
		retryAt := time.Now().UTC().Add(proposalRunnerBackoff(item.Attempts))
		payload := clonePayload(item.Payload)
		if payload == nil {
			payload = map[string]any{}
		}
		payload["dedupe_key"] = fmt.Sprintf("proposal-runner:%s", item.ProposalID)
		if _, err := store.RescheduleWorkItem(item.ID, payload, runnerErr.Error(), retryAt); err != nil {
			return err
		}
		now := time.Now().UTC()
		_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
			Events: []events.TraceEvent{
				{
					TraceID:     trace.Summary.TraceID,
					IngestionID: trace.Summary.IngestionID,
					WorkflowID:  trace.Summary.WorkflowID,
					Plane:       "execution",
					Service:     "runner",
					Actor:       "proposal",
					EventType:   "runner.failed",
					Status:      events.StatusFailed,
					StartedAt:   runnerStarted,
					EndedAt:     ptrTime(now),
					Description: fmt.Sprintf("Proposal runner failed; proposal remains approved and will retry after %s: %v", retryAt.Format(time.RFC3339), runnerErr),
				},
			},
			Reasoning: []events.ReasoningStep{
				{
					ID:         fmt.Sprintf("reason-proposal-retry-%d", now.UnixNano()),
					TraceID:    trace.Summary.TraceID,
					WorkflowID: trace.Summary.WorkflowID,
					StepType:   "proposal_runner_retry",
					Summary:    "Proposal runner failed closed. Repo-change work was not materialized.",
					Confidence: 0.93,
					Decision:   retryAt.Format(time.RFC3339),
					CreatedAt:  now,
				},
			},
		})
		return errDeferredWorkItem
	}
	if proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		return processHarnessOverlayProposal(cfg, store, trace, proposal, attempt, runnerResp, runnerOutput, runnerStarted)
	}
	now := time.Now().UTC()
	failureClass, failureSummary, changedFiles := validateRepoPatch(runnerOutput.RepoPatch, runnerOutput.RetryAssessment.ChangedFiles)
	attempt.ChangePlan = strings.TrimSpace(runnerOutput.ChangePlan)
	attempt.RepoPatch = strings.TrimSpace(runnerOutput.RepoPatch)
	attempt.ValidationPlan = strings.TrimSpace(runnerOutput.ValidationPlan)
	attempt.HypothesisDelta = strings.TrimSpace(runnerOutput.HypothesisDelta)
	attempt.DiffSummary = firstNonEmpty(strings.TrimSpace(runnerOutput.ChangePlan), strings.TrimSpace(runnerOutput.ContextSummary))
	attempt.ValidationSummary = strings.TrimSpace(runnerOutput.ValidationPlan)
	attempt.ChangedFiles = changedFiles
	attempt.FailureClass = firstNonEmpty(runnerOutput.RetryAssessment.FailureClass, failureClass)
	attempt.FailureSummary = firstNonEmpty(runnerOutput.RetryAssessment.FailureSummary, failureSummary)
	attempt.RetryDecision = strings.TrimSpace(runnerOutput.RetryAssessment.RetryDecision)
	attempt.MaterialHypothesisChange = runnerOutput.RetryAssessment.MaterialHypothesisChange
	attempt.UpdatedAt = now
	if failureClass != "" {
		return recordAttemptFailure(cfg, store, proposal, attempt, trace, failureClass, failureSummary, runnerOutput.RetryAssessment.MaterialHypothesisChange, improvement.AttemptTriggerProposalApproved)
	}
	attempt.State = improvement.AttemptStatePatchGenerated
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		return err
	}
	job, err := store.MaterializeApprovedProposal(item.ProposalID, cfg.ServiceName)
	if err != nil {
		return err
	}
	manifest := sandbox.BuildJob(cfg, sandbox.JobRequest{
		TraceID:      trace.Summary.TraceID,
		ProposalID:   item.ProposalID,
		Repo:         job.Repo,
		BaseRef:      job.BaseRef,
		RequestedBy:  cfg.ServiceName,
		ArtifactPath: fmt.Sprintf("memory://sandbox/%s", job.ID),
		Env: map[string]string{
			"RSI_ATTEMPT_ID":      attempt.ID,
			"RSI_CHANGE_PLAN":     attempt.ChangePlan,
			"RSI_REPO_PATCH":      attempt.RepoPatch,
			"RSI_VALIDATION_PLAN": attempt.ValidationPlan,
		},
		Commands: repoChangeCommands(),
	})
	_, _ = store.UpdateRepoChangeJobStatus(job.ID, string(review.ProposalRepoChangeQueued))
	_, err = upsertImprovementActionIntent(store, action.Intent{
		OwnerPlane:     "improvement",
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		TraceID:        trace.Summary.TraceID,
		ProposalID:     proposal.ID,
		AttemptID:      attempt.ID,
		Kind:           action.KindSandboxLaunch,
		TargetRef:      fmt.Sprintf("%s/%s", cfg.SandboxNamespace, job.ID),
		RequestPayload: map[string]any{
			"job_id":      job.ID,
			"attempt_id":  attempt.ID,
			"repo":        job.Repo,
			"branch_name": job.BranchName,
			"base_ref":    job.BaseRef,
		},
		IdempotencyKey: fmt.Sprintf("sandbox:%s", job.ID),
		ApprovalMode:   "approved",
		ApprovalState:  "approved",
		PolicyVerdict:  "approved_proposal",
		Status:         action.StatusQueued,
		RequestedBy:    cfg.ServiceName,
		Rationale:      firstNonEmpty(firstProposedActionRationale(runnerOutput.ProposedActions, action.KindSandboxLaunch), "Launch the sandbox repo-change job after human approval."),
		EvidenceRefs: []events.EvidenceRef{
			{Kind: "proposal", Ref: proposal.ID, Summary: proposal.Title},
			{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
			{Kind: "change_attempt", Ref: attempt.ID, Summary: attempt.ChangePlan},
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return err
	}
	reasoning := []events.ReasoningStep{
		{
			ID:         fmt.Sprintf("reason-proposal-%d", now.UnixNano()),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "proposal_materialized",
			Summary:    firstNonEmpty(runnerOutput.ChangePlan, runnerOutput.FinalAnswer, runnerOutput.ContextSummary, fmt.Sprintf("Approved proposal %s moved into repo-change queue.", item.ProposalID)),
			Confidence: confidenceOr(0.86, runnerOutput.Confidence),
			Decision:   job.BranchName,
			CreatedAt:  now,
		},
	}
	reasoning = append(runnerutil.ToTraceReasoning(trace.Summary.TraceID, trace.Summary.WorkflowID, runnerOutput, now), reasoning...)
	reasoning = append(reasoning, improvementOutcomeHypothesisReasoning(trace, runnerOutput.OutcomeHypotheses, now)...)
	if err := persistImprovementKnowledgeDrafts(store, runnerOutput.KnowledgeDrafts, trace, proposal.ID, now); err != nil {
		return err
	}
	if strings.TrimSpace(runnerOutput.SelfCritique) != "" {
		reasoning = append(reasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-proposal-self-%d", now.UnixNano()),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "self_critique",
			Summary:    runnerOutput.SelfCritique,
			Confidence: confidenceOr(0.86, runnerOutput.Confidence),
			CreatedAt:  now,
		})
	}
	runnerEvent := events.TraceEvent{
		TraceID:     trace.Summary.TraceID,
		IngestionID: trace.Summary.IngestionID,
		WorkflowID:  trace.Summary.WorkflowID,
		Plane:       "execution",
		Service:     "runner",
		Actor:       "proposal",
		EventType:   "runner.completed",
		Status:      events.StatusCompleted,
		StartedAt:   runnerStarted,
		EndedAt:     ptrTime(now),
		Description: fmt.Sprintf("Proposal runner returned repo-change rationale using %s.", runnerRuntimeLabel(runnerResp)),
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			runnerEvent,
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "repo_change.queued",
				Status:      events.StatusQueued,
				StartedAt:   now,
				Description: fmt.Sprintf("Materialized repo change job %s.", job.ID),
			},
		},
		Artifacts: []events.Artifact{
			{
				ID:          fmt.Sprintf("artifact-patch-%d", now.UnixNano()),
				TraceID:     trace.Summary.TraceID,
				Kind:        "repo_patch",
				ContentType: "text/x-diff",
				URL:         fmt.Sprintf("memory://attempt/%s/repo.patch", attempt.ID),
				SizeBytes:   int64(len(attempt.RepoPatch)),
				Source:      "proposal-runner",
			},
			{
				ID:          fmt.Sprintf("artifact-sandbox-%d", now.UnixNano()),
				TraceID:     trace.Summary.TraceID,
				Kind:        "sandbox_job_manifest",
				ContentType: "application/json",
				URL:         fmt.Sprintf("memory://sandbox/%s/manifest.json", job.ID),
				SizeBytes:   int64(len(fmt.Sprintf("%v", manifest))),
				Source:      "improvement-plane",
			},
		},
		Reasoning: reasoning,
	})
	return nil
}

func processSandboxItem(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	switch item.Kind {
	case "repo_change_job":
		return processSandboxLaunch(cfg, store, launcher, launcherErr, item)
	case "watch_sandbox_job":
		return processSandboxWatch(cfg, store, launcher, launcherErr, item)
	default:
		return fmt.Errorf("unsupported sandbox item kind %s", item.Kind)
	}
}

func processHarnessOverlayProposal(cfg config.Config, store storepkg.Store, trace events.Trace, proposal review.Proposal, attempt improvement.ChangeAttempt, runnerResp clients.RunnerResponse, runnerOutput runnerutil.StructuredOutput, runnerStarted time.Time) error {
	overlay, err := buildHarnessOverlayFromRunner(store, proposal, runnerOutput)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	attempt.State = improvement.AttemptStateOverlayGenerated
	attempt.ChangePlan = firstNonEmpty(strings.TrimSpace(runnerOutput.ChangePlan), strings.TrimSpace(runnerOutput.FinalAnswer), proposal.Summary)
	attempt.ValidationPlan = strings.TrimSpace(runnerOutput.ValidationPlan)
	attempt.OverlayPayload = map[string]any{
		"overlay_id":            overlay.ID,
		"prompt_fragments":      overlay.PromptFragments,
		"few_shot_snippets":     overlay.FewShotSnippets,
		"tool_preference_order": overlay.ToolPreferenceOrder,
		"retrieval_bias":        overlay.RetrievalBias,
		"reasoning_verbosity":   overlay.ReasoningVerbosity,
	}
	attempt.UpdatedAt = now
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		return err
	}
	if _, err := store.UpsertHarnessOverlay(overlay); err != nil {
		return err
	}
	if _, err := store.RecordHarnessExperiment(harness.Experiment{
		ID:         fmt.Sprintf("hexp-%s", proposal.ID),
		ProfileID:  overlay.ProfileID,
		OverlayID:  overlay.ID,
		ProposalID: proposal.ID,
		AttemptID:  attempt.ID,
		Role:       overlay.Role,
		Status:     harness.ExperimentStatusSucceeded,
		Summary:    firstNonEmpty(runnerOutput.FinalAnswer, runnerOutput.ContextSummary, proposal.Summary),
		Metrics: map[string]any{
			"target_layer": proposal.TargetLayer,
			"target_kind":  proposal.TargetKind,
			"target_ref":   proposal.TargetRef,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return err
	}
	intent, err := upsertImprovementActionIntent(store, action.Intent{
		OwnerPlane:     "improvement",
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		TraceID:        trace.Summary.TraceID,
		ProposalID:     proposal.ID,
		AttemptID:      attempt.ID,
		Kind:           action.KindHarnessOverlay,
		TargetRef:      overlay.Role,
		RequestPayload: map[string]any{
			"overlay_id":                overlay.ID,
			"profile_id":                overlay.ProfileID,
			"version":                   overlay.Version,
			"prompt_fragments":          overlay.PromptFragments,
			"few_shot_snippets":         overlay.FewShotSnippets,
			"tool_preference_order":     overlay.ToolPreferenceOrder,
			"retrieval_bias":            overlay.RetrievalBias,
			"reasoning_verbosity":       overlay.ReasoningVerbosity,
			"memory_read_enabled":       boolPointerValue(overlay.MemoryReadEnabled),
			"memory_write_enabled":      boolPointerValue(overlay.MemoryWriteEnabled),
			"effective_overlay_version": overlay.Version,
		},
		IdempotencyKey: fmt.Sprintf("harness-overlay:%s", proposal.ID),
		ApprovalMode:   "approved",
		ApprovalState:  "approved",
		PolicyVerdict:  "approved_proposal",
		Status:         action.StatusSucceeded,
		RequestedBy:    cfg.ServiceName,
		Rationale:      firstNonEmpty(firstProposedActionRationale(runnerOutput.ProposedActions, action.KindHarnessOverlay), runnerOutput.FinalAnswer, "Activated runtime harness overlay after human approval."),
		EvidenceRefs: []events.EvidenceRef{
			{Kind: "proposal", Ref: proposal.ID, Summary: proposal.Title},
			{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return err
	}
		if _, err := store.RecordActionResult(action.Result{
			OperationID:    intent.OperationID,
			ActionIntentID: intent.ID,
			AttemptID:      attempt.ID,
			Executor:       cfg.ServiceName,
		Provider:       "rsi-platform",
		ProviderRef:    overlay.ID,
		Status:         action.StatusSucceeded,
		StartedAt:      runnerStarted,
		CompletedAt:    now,
	}); err != nil {
		return err
	}
	if _, err := store.UpdateProposalStatus(proposal.ID, review.ProposalMerged); err != nil {
		return err
	}
	attempt.State = improvement.AttemptStateMerged
	attempt.UpdatedAt = now
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		return err
	}
	reasoning := []events.ReasoningStep{
		{
			ID:         fmt.Sprintf("reason-overlay-%d", now.UnixNano()),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "harness_overlay_activation",
			Summary:    firstNonEmpty(runnerOutput.FinalAnswer, fmt.Sprintf("Activated overlay %s for role %s.", overlay.Version, overlay.Role)),
			Confidence: confidenceOr(0.87, runnerOutput.Confidence),
			Decision:   overlay.Version,
			CreatedAt:  now,
		},
	}
	reasoning = append(reasoning, runnerutil.ToTraceReasoning(trace.Summary.TraceID, trace.Summary.WorkflowID, runnerOutput, now)...)
	if err := persistImprovementKnowledgeDrafts(store, runnerOutput.KnowledgeDrafts, trace, proposal.ID, now); err != nil {
		return err
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       "proposal",
				EventType:   "runner.completed",
				Status:      events.StatusCompleted,
				StartedAt:   runnerStarted,
				EndedAt:     ptrTime(now),
				Description: fmt.Sprintf("Proposal runner returned harness overlay rationale using %s.", runnerRuntimeLabel(runnerResp)),
			},
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "harness.overlay.activated",
				Status:      events.StatusCompleted,
				StartedAt:   now,
				EndedAt:     ptrTime(now),
				Description: fmt.Sprintf("Activated harness overlay %s for role %s.", overlay.Version, overlay.Role),
			},
		},
		Reasoning: reasoning,
	})
	return nil
}

func processSandboxLaunch(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	jobID := stringValue(item.Payload["job_id"])
	if jobID == "" {
		return fmt.Errorf("sandbox work item missing job_id")
	}
	repoJob, ok := findRepoChangeJob(store.ListRepoChangeJobs(), jobID)
	if !ok {
		return fmt.Errorf("repo change job %s not found", jobID)
	}
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	attemptID := firstNonEmpty(stringValue(item.Payload["attempt_id"]), repoJob.AttemptID)
	attempt, _ := store.GetChangeAttempt(attemptID)
	proposal, _ := findProposal(store.ListProposals(), item.ProposalID)
	intent, err := upsertImprovementActionIntent(store, action.Intent{
		OwnerPlane:     "improvement",
		ConversationID: repoJob.ConversationID,
		CaseID:         repoJob.CaseID,
		TraceID:        trace.Summary.TraceID,
		ProposalID:     item.ProposalID,
		AttemptID:      attemptID,
		Kind:           action.KindSandboxLaunch,
		TargetRef:      fmt.Sprintf("%s/%s", cfg.SandboxNamespace, repoJob.ID),
		RequestPayload: map[string]any{
			"job_id":      repoJob.ID,
			"attempt_id":  attemptID,
			"repo":        repoJob.Repo,
			"branch_name": repoJob.BranchName,
			"base_ref":    repoJob.BaseRef,
		},
		IdempotencyKey: fmt.Sprintf("sandbox:%s", repoJob.ID),
		ApprovalMode:   "approved",
		ApprovalState:  "approved",
		PolicyVerdict:  "approved_proposal",
		Status:         action.StatusExecuting,
		RequestedBy:    cfg.ServiceName,
		Rationale:      "Launch the sandbox job to validate the approved repo change.",
		EvidenceRefs: []events.EvidenceRef{
			{Kind: "proposal", Ref: item.ProposalID, Summary: repoJob.CandidateKey},
			{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	repoOwner := cfg.GitHubRepoOwner(repoJob.Repo)
	repoName := cfg.GitHubRepoName(repoJob.Repo)
	writeToken, err := githubapp.NewClient(
		cfg.GitHubAppID,
		cfg.GitHubInstallationIDForRepo(repoJob.Repo),
		cfg.GitHubAppPrivateKey,
		cfg.GitHubAPIBaseURL,
		&http.Client{Timeout: 30 * time.Second},
	).MintInstallationToken(context.Background(), []string{repoName})
	if err != nil {
		return fmt.Errorf("mint github app installation token for sandbox launch: %w", err)
	}
	request := sandbox.JobRequest{
		TraceID:      trace.Summary.TraceID,
		ProposalID:   item.ProposalID,
		Repo:         repoName,
		BaseRef:      repoJob.BaseRef,
		RequestedBy:  cfg.ServiceName,
		ArtifactPath: fmt.Sprintf("memory://sandbox/%s", repoJob.ID),
		Env: map[string]string{
			"GITHUB_TOKEN":        writeToken.Token,
			"GITHUB_OWNER":        repoOwner,
			"GITHUB_COMMIT_USER":  cfg.GitHubCommitUser,
			"GITHUB_COMMIT_EMAIL": cfg.GitHubCommitEmail,
			"RSI_BRANCH_NAME":     repoJob.BranchName,
			"RSI_CONTEXT_SUMMARY": repoJob.ContextSummary,
			"RSI_CHANGE_PLAN":     attempt.ChangePlan,
			"RSI_REPO_PATCH":      attempt.RepoPatch,
			"RSI_VALIDATION_PLAN": attempt.ValidationPlan,
			"RSI_ATTEMPT_ID":      attemptID,
			"RSI_REPO":            repoName,
			"RSI_BASE_REF":        repoJob.BaseRef,
			"RSI_PROPOSAL_ID":     item.ProposalID,
		},
		Commands: repoChangeCommands(),
	}
	if launcherErr != nil || launcher == nil {
		completed := time.Now().UTC()
		if _, err := upsertImprovementActionIntent(store, action.Intent{
			ID:             intent.ID,
			OwnerPlane:     intent.OwnerPlane,
			ConversationID: intent.ConversationID,
			CaseID:         intent.CaseID,
			TraceID:        intent.TraceID,
			ProposalID:     intent.ProposalID,
			AttemptID:      intent.AttemptID,
			Kind:           intent.Kind,
			TargetRef:      intent.TargetRef,
			RequestPayload: intent.RequestPayload,
			IdempotencyKey: intent.IdempotencyKey,
			ApprovalMode:   intent.ApprovalMode,
			ApprovalState:  intent.ApprovalState,
			PolicyVerdict:  intent.PolicyVerdict,
			Status:         action.StatusBlocked,
			RequestedBy:    intent.RequestedBy,
			Rationale:      intent.Rationale,
			EvidenceRefs:   intent.EvidenceRefs,
			CreatedAt:      intent.CreatedAt,
			UpdatedAt:      completed,
		}); err != nil {
			return err
		}
		_, _ = store.RecordActionResult(action.Result{
			OperationID:    item.OperationID,
			ActionIntentID: intent.ID,
			AttemptID:      attemptID,
			Executor:       "sandbox-runtime",
			Provider:       "kubernetes",
			Status:         action.StatusBlocked,
			ErrorCode:      "sandbox_unavailable",
			ErrorMessage:   firstNonEmpty(errorString(launcherErr), "sandbox launcher not configured"),
			StartedAt:      completed,
			CompletedAt:    completed,
		})
		if attempt.ID != "" && proposal.ID != "" {
			return recordAttemptFailure(cfg, store, proposal, attempt, trace, "sandbox_failure", firstNonEmpty(errorString(launcherErr), "Sandbox launcher unavailable."), false, improvement.AttemptTriggerSandboxFailed)
		}
		return nil
	}
	session, _, err := launcher.Launch(context.Background(), request)
	if err != nil {
		completed := time.Now().UTC()
		_, _ = upsertImprovementActionIntent(store, action.Intent{
			ID:             intent.ID,
			OwnerPlane:     intent.OwnerPlane,
			ConversationID: intent.ConversationID,
			CaseID:         intent.CaseID,
			TraceID:        intent.TraceID,
			ProposalID:     intent.ProposalID,
			AttemptID:      intent.AttemptID,
			Kind:           intent.Kind,
			TargetRef:      intent.TargetRef,
			RequestPayload: intent.RequestPayload,
			IdempotencyKey: intent.IdempotencyKey,
			ApprovalMode:   intent.ApprovalMode,
			ApprovalState:  intent.ApprovalState,
			PolicyVerdict:  intent.PolicyVerdict,
			Status:         action.StatusFailed,
			RequestedBy:    intent.RequestedBy,
			Rationale:      intent.Rationale,
			EvidenceRefs:   intent.EvidenceRefs,
			CreatedAt:      intent.CreatedAt,
			UpdatedAt:      completed,
		})
		_, _ = store.RecordActionResult(action.Result{
			OperationID:    item.OperationID,
			ActionIntentID: intent.ID,
			AttemptID:      attemptID,
			Executor:       "sandbox-runtime",
			Provider:       "kubernetes",
			Status:         action.StatusFailed,
			ErrorCode:      "sandbox_launch_failed",
			ErrorMessage:   err.Error(),
			StartedAt:      intent.UpdatedAt,
			CompletedAt:    completed,
		})
		if attempt.ID != "" && proposal.ID != "" {
			return recordAttemptFailure(cfg, store, proposal, attempt, trace, "sandbox_failure", err.Error(), false, improvement.AttemptTriggerSandboxFailed)
		}
		return err
	}
	now := time.Now().UTC()
	if attempt.ID != "" {
		attempt.State = improvement.AttemptStateSandboxRunning
		attempt.UpdatedAt = now
		_, _ = store.UpsertChangeAttempt(attempt)
	}
	repoJob.Status = string(review.ProposalRepoChangeRunning)
	repoJob.SandboxNamespace = session.Namespace
	repoJob.SandboxJobName = session.PodName
	repoJob.SandboxPodName = session.PodName
	repoJob.ValidationRef = fmt.Sprintf("%s/%s", session.Namespace, session.PodName)
	repoJob.UpdatedAt = now
	_, _ = store.UpsertRepoChangeJob(repoJob)
	_, _ = store.UpdateProposalStatus(item.ProposalID, review.ProposalRepoChangeRunning)
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     cfg.ServiceName,
				Actor:       "sandbox-launcher",
				EventType:   "sandbox.job.started",
				Status:      events.StatusRunning,
				StartedAt:   now,
				Description: fmt.Sprintf("Launched sandbox job %s in namespace %s.", session.PodName, session.Namespace),
			},
		},
		Artifacts: []events.Artifact{
			{
				ID:          fmt.Sprintf("artifact-sandbox-launch-%d", now.UnixNano()),
				TraceID:     trace.Summary.TraceID,
				Kind:        "sandbox_job",
				ContentType: "text/plain",
				URL:         fmt.Sprintf("k8s://%s/jobs/%s", session.Namespace, session.PodName),
				SizeBytes:   0,
				Source:      "sandbox-runtime",
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:         fmt.Sprintf("reason-sandbox-launch-%d", now.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "sandbox_launch",
				Summary:    fmt.Sprintf("Launched real sandbox job for repo %s branch %s.", repoJob.Repo, repoJob.BranchName),
				Confidence: 0.88,
				Decision:   session.PodName,
				CreatedAt:  now,
			},
		},
	})
	_, err = enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attemptID,
		OperationKind: "sandbox_watch",
		OperationKey:  "sandbox_watch",
		Status:        operation.StatusQueued,
		Queue:         queue.SandboxQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       trace.Summary.TraceID,
		ProposalID:    item.ProposalID,
		AttemptID:     attemptID,
	}, queue.WorkItem{
		Queue:      queue.SandboxQueue,
		Kind:       "watch_sandbox_job",
		Status:     queue.WorkQueued,
		TraceID:    trace.Summary.TraceID,
		ProposalID: item.ProposalID,
		Payload: map[string]interface{}{
			"attempt_id":  attemptID,
			"job_name":    session.PodName,
			"namespace":   session.Namespace,
			"repo":        repoJob.Repo,
			"branch_name": repoJob.BranchName,
			"base_ref":    repoJob.BaseRef,
			"job_id":      repoJob.ID,
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	return err
}

func processSandboxWatch(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	if launcherErr != nil {
		return launcherErr
	}
	if launcher == nil {
		return fmt.Errorf("sandbox launcher not configured")
	}
	jobName := stringValue(item.Payload["job_name"])
	namespace := stringValue(item.Payload["namespace"])
	repo := stringValue(item.Payload["repo"])
	branchName := stringValue(item.Payload["branch_name"])
	jobID := stringValue(item.Payload["job_id"])
	attemptID := stringValue(item.Payload["attempt_id"])
	if jobName == "" || namespace == "" {
		return fmt.Errorf("sandbox watch item missing job metadata")
	}
	observation, err := launcher.ObserveJob(context.Background(), namespace, jobName)
	if err != nil {
		return err
	}
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	proposal, _ := findProposal(store.ListProposals(), item.ProposalID)
	attempt, _ := store.GetChangeAttempt(firstNonEmpty(attemptID, proposal.CurrentAttemptID))
	if !observation.JobSucceeded && !observation.JobFailed {
		retryAt := time.Now().UTC().Add(cfg.SandboxPollInterval)
		payload := clonePayload(item.Payload)
		if payload == nil {
			payload = map[string]any{}
		}
		if _, err := store.RescheduleWorkItem(item.ID, payload, "", retryAt); err != nil {
			return err
		}
		return errDeferredWorkItem
	}
	now := time.Now().UTC()
	statusArtifactID, logArtifactID, sandboxArtifacts := sandboxObservationArtifacts(trace.Summary.TraceID, observation, now)
	if observation.JobFailed {
		errorMessage := sandboxFailureMessage(observation)
		intent, _ := findProposalActionIntent(store, item.ProposalID, action.KindSandboxLaunch)
		if intent.ID != "" {
			_, _ = upsertImprovementActionIntent(store, withActionStatus(intent, action.StatusFailed, now))
			_, _ = store.RecordActionResult(action.Result{
				OperationID:        item.OperationID,
				ActionIntentID:     intent.ID,
				AttemptID:          firstNonEmpty(intent.AttemptID, attempt.ID),
				Executor:           "sandbox-runtime",
				Provider:           "kubernetes",
				ProviderRef:        fmt.Sprintf("%s/%s", observation.Namespace, observation.JobName),
				ResponseArtifactID: statusArtifactID,
				Status:             action.StatusFailed,
				ErrorCode:          "sandbox_job_failed",
				ErrorMessage:       errorMessage,
				StartedAt:          intent.UpdatedAt,
				CompletedAt:        now,
			})
		}
		repoJob, ok := findRepoChangeJob(store.ListRepoChangeJobs(), jobID)
		if ok {
			repoJob.Status = string(review.ProposalFailedValidation)
			repoJob.SandboxNamespace = observation.Namespace
			repoJob.SandboxJobName = observation.JobName
			repoJob.SandboxPodName = observation.PodName
			repoJob.ValidationError = errorMessage
			repoJob.ValidationRef = fmt.Sprintf("%s/%s", observation.Namespace, observation.JobName)
			repoJob.LogArtifactID = logArtifactID
			repoJob.UpdatedAt = now
			_, _ = store.UpsertRepoChangeJob(repoJob)
		} else {
			_, _ = store.UpdateRepoChangeJobStatus(jobID, string(review.ProposalFailedValidation))
		}
		_, _ = store.UpdateProposalStatus(item.ProposalID, review.ProposalFailedValidation)
		if attempt.ID != "" && proposal.ID != "" {
			attempt.State = improvement.AttemptStateSandboxFailed
			attempt.FailureClass = "sandbox_failure"
			attempt.FailureSummary = errorMessage
			attempt.ValidationSummary = errorMessage
			attempt.UpdatedAt = now
			_, _ = store.UpsertChangeAttempt(attempt)
		}
		_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
			Status: ptrStatus(events.StatusFailed),
			Events: []events.TraceEvent{
				{
					TraceID:     trace.Summary.TraceID,
					IngestionID: trace.Summary.IngestionID,
					WorkflowID:  trace.Summary.WorkflowID,
					Plane:       "execution",
					Service:     cfg.ServiceName,
					Actor:       "sandbox-launcher",
					EventType:   "sandbox.job.failed",
					Status:      events.StatusFailed,
					StartedAt:   now,
					Description: errorMessage,
				},
			},
			Artifacts: sandboxArtifacts,
		})
		if attempt.ID != "" && proposal.ID != "" {
			return recordAttemptFailure(cfg, store, proposal, attempt, trace, "sandbox_failure", errorMessage, false, improvement.AttemptTriggerSandboxFailed)
		}
		return nil
	}

	intent, _ := findProposalActionIntent(store, item.ProposalID, action.KindSandboxLaunch)
	if intent.ID != "" {
		_, _ = upsertImprovementActionIntent(store, withActionStatus(intent, action.StatusSucceeded, now))
		_, _ = store.RecordActionResult(action.Result{
			OperationID:        item.OperationID,
			ActionIntentID:     intent.ID,
			AttemptID:          firstNonEmpty(intent.AttemptID, attempt.ID),
			Executor:           "sandbox-runtime",
			Provider:           "kubernetes",
			ProviderRef:        fmt.Sprintf("%s/%s", observation.Namespace, observation.JobName),
			ResponseArtifactID: statusArtifactID,
			Status:             action.StatusSucceeded,
			StartedAt:          intent.UpdatedAt,
			CompletedAt:        now,
		})
	}
	repoJob, ok := findRepoChangeJob(store.ListRepoChangeJobs(), jobID)
	if ok {
		repoJob.Status = string(review.ProposalValidationPending)
		repoJob.SandboxNamespace = observation.Namespace
		repoJob.SandboxJobName = observation.JobName
		repoJob.SandboxPodName = observation.PodName
		repoJob.ValidationError = ""
		repoJob.ValidationRef = fmt.Sprintf("%s/%s", observation.Namespace, observation.JobName)
		repoJob.LogArtifactID = logArtifactID
		repoJob.UpdatedAt = now
		_, _ = store.UpsertRepoChangeJob(repoJob)
	} else {
		_, _ = store.UpdateRepoChangeJobStatus(jobID, string(review.ProposalValidationPending))
	}
	_, _ = store.UpdateProposalStatus(item.ProposalID, review.ProposalValidationPending)
	if attempt.ID != "" {
		attempt.State = improvement.AttemptStateValidationReady
		attempt.ValidationSummary = fmt.Sprintf("Sandbox validation succeeded for %s.", observation.JobName)
		attempt.UpdatedAt = now
		_, _ = store.UpsertChangeAttempt(attempt)
	}
	_, err = enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: "pr_open",
		OperationKey:  "pr_open",
		Status:        operation.StatusQueued,
		Queue:         queue.ImprovementActionQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       item.TraceID,
		ProposalID:    item.ProposalID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		Queue:      queue.ImprovementActionQueue,
		Kind:       "draft_pr_open",
		Status:     queue.WorkQueued,
		TraceID:    item.TraceID,
		ProposalID: item.ProposalID,
		CreatedAt:  now,
		UpdatedAt:  now,
		Payload: map[string]any{
			"attempt_id":  attempt.ID,
			"job_id":      jobID,
			"job_name":    observation.JobName,
			"namespace":   observation.Namespace,
			"repo":        repo,
			"branch_name": branchName,
			"base_ref":    firstNonEmpty(stringValue(item.Payload["base_ref"]), "main"),
		},
	})
	if err != nil {
		return err
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "github.pr.queued",
				Status:      events.StatusQueued,
				StartedAt:   now,
				Description: fmt.Sprintf("Sandbox job %s succeeded; queued draft PR open for branch %s.", observation.JobName, branchName),
			},
		},
		Artifacts: sandboxArtifacts,
		Reasoning: []events.ReasoningStep{
			{
				ID:         fmt.Sprintf("reason-pr-open-%d", now.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "pr_queue",
				Summary:    fmt.Sprintf("Sandbox validation succeeded; queued draft PR open for branch %s.", branchName),
				Confidence: 0.9,
				Decision:   branchName,
				CreatedAt:  now,
			},
		},
	})
	return nil
}

func processImprovementActionItem(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, item queue.WorkItem) error {
	switch item.Kind {
	case "draft_pr_open":
		return processDraftPROpen(cfg, store, toolClient, item)
	default:
		return fmt.Errorf("unsupported improvement action kind %s", item.Kind)
	}
}

func processDraftPROpen(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, item queue.WorkItem) error {
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	repo := stringValue(item.Payload["repo"])
	branchName := stringValue(item.Payload["branch_name"])
	baseRef := firstNonEmpty(stringValue(item.Payload["base_ref"]), "main")
	jobID := stringValue(item.Payload["job_id"])
	attemptID := stringValue(item.Payload["attempt_id"])
	now := time.Now().UTC()
	proposal, _ := findProposal(store.ListProposals(), item.ProposalID)
	attempt, _ := store.GetChangeAttempt(firstNonEmpty(attemptID, proposal.CurrentAttemptID))
	intent, err := upsertImprovementActionIntent(store, action.Intent{
		OwnerPlane:     "improvement",
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        trace.Summary.TraceID,
		ProposalID:     item.ProposalID,
		AttemptID:      attemptID,
		Kind:           action.KindDraftPROpen,
		TargetRef:      repo,
		RequestPayload: map[string]any{
			"proposal_id": item.ProposalID,
			"attempt_id":  attemptID,
			"repo":        repo,
			"branch_name": branchName,
			"base_ref":    baseRef,
		},
		IdempotencyKey: fmt.Sprintf("pr:%s:%s", attemptID, branchName),
		ApprovalMode:   "approved",
		ApprovalState:  "approved",
		PolicyVerdict:  "approved_proposal",
		Status:         action.StatusExecuting,
		RequestedBy:    cfg.ServiceName,
		Rationale:      "Open a draft PR once sandbox validation succeeds.",
		EvidenceRefs: []events.EvidenceRef{
			{Kind: "proposal", Ref: item.ProposalID, Summary: item.ProposalID},
			{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return err
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "github.pr.started",
				Status:      events.StatusRunning,
				StartedAt:   now,
				Description: fmt.Sprintf("Opening draft PR for %s on branch %s.", repo, branchName),
			},
		},
	})
	prResult, execErr := toolClient.Execute("github.create_pr", map[string]any{
		"proposal_id": item.ProposalID,
		"attempt_id":  attemptID,
		"repo":        repo,
		"branch_name": branchName,
		"base_ref":    baseRef,
		"title":       fmt.Sprintf("RSI proposal %s attempt %s for %s", item.ProposalID, attemptID, repo),
		"body":        fmt.Sprintf("Automated draft PR for proposal %s attempt %s after sandbox validation.", item.ProposalID, attemptID),
	})
	completed := time.Now().UTC()
	actionStatus := improvementActionStatus(prResult, execErr)
	_, _ = upsertImprovementActionIntent(store, withActionStatus(intent, actionStatus, completed))
	_, _ = store.RecordActionResult(action.Result{
		OperationID:    item.OperationID,
		ActionIntentID: intent.ID,
		AttemptID:      attemptID,
		Executor:       "tool-gateway",
		Provider:       firstNonEmpty(prResult.Provider, "github"),
		ProviderRef:    prResult.ProviderRef,
		Status:         actionStatus,
		ErrorCode:      improvementActionErrorCode(actionStatus),
		ErrorMessage:   improvementActionError(prResult, execErr),
		StartedAt:      intent.UpdatedAt,
		CompletedAt:    completed,
	})
	if execErr != nil || actionStatus != action.StatusSucceeded {
		_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
			Status: ptrStatus(events.StatusNeedsHuman),
			Events: []events.TraceEvent{
				{
					TraceID:     trace.Summary.TraceID,
					IngestionID: trace.Summary.IngestionID,
					WorkflowID:  trace.Summary.WorkflowID,
					Plane:       "improvement",
					Service:     cfg.ServiceName,
					Actor:       "worker",
					EventType:   "github.pr.blocked",
					Status:      events.StatusNeedsHuman,
					StartedAt:   now,
					EndedAt:     ptrTime(completed),
					Description: firstNonEmpty(improvementActionError(prResult, execErr), "Draft PR open blocked."),
				},
			},
		})
		if attempt.ID != "" && proposal.ID != "" {
			return recordAttemptFailure(cfg, store, proposal, attempt, trace, "stale_branch", firstNonEmpty(improvementActionError(prResult, execErr), "Draft PR open blocked."), false, improvement.AttemptTriggerCIFailed)
		}
		return nil
	}
	prURL := stringValue(prResult.Output["pr_url"])
	headSHA := nestedString(prResult.Output, "response", "head", "sha")
	if _, err := store.RecordPRAttempt(buildPRAttempt(item.ProposalID, attemptID, repo, branchName, prURL, headSHA)); err != nil {
		return err
	}
	if attempt.ID != "" {
		attempt.State = improvement.AttemptStatePROpen
		attempt.PRURL = prURL
		attempt.HeadSHA = headSHA
		attempt.UpdatedAt = completed
		_, _ = store.UpsertChangeAttempt(attempt)
	}
	_, _ = store.UpdateRepoChangeJobStatus(jobID, string(review.ProposalPROpen))
	_, _ = store.UpdateProposalStatus(item.ProposalID, review.ProposalPROpen)
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "github.pr.completed",
				Status:      events.StatusCompleted,
				StartedAt:   now,
				EndedAt:     ptrTime(completed),
				Description: fmt.Sprintf("Opened draft PR for %s on branch %s.", repo, branchName),
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:         fmt.Sprintf("reason-pr-open-%d", completed.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "pr_opened",
				Summary:    fmt.Sprintf("Opened real draft PR for branch %s.", branchName),
				Confidence: 0.9,
				Decision:   prURL,
				CreatedAt:  completed,
			},
		},
	})
	return nil
}

func buildPRAttempt(proposalID string, attemptID string, repo string, branchName string, prURL string, headSHA string) improvement.PRAttempt {
	return improvement.PRAttempt{
		ProposalID:       proposalID,
		AttemptID:        attemptID,
		Repo:             repo,
		BranchName:       branchName,
		PRURL:            prURL,
		HeadSHA:          headSHA,
		Status:           string(review.ProposalPROpen),
		ValidationStatus: "pending",
	}
}

func findRepoChangeJob(items []improvement.RepoChangeJob, jobID string) (improvement.RepoChangeJob, bool) {
	for _, item := range items {
		if item.ID == jobID {
			return item, true
		}
	}
	return improvement.RepoChangeJob{}, false
}

func sandboxObservationArtifacts(traceID string, observation sandbox.JobObservation, createdAt time.Time) (string, string, []events.Artifact) {
	statusArtifactID := fmt.Sprintf("artifact-sandbox-status-%d", createdAt.UnixNano())
	artifacts := []events.Artifact{
		{
			ID:          statusArtifactID,
			TraceID:     traceID,
			Kind:        "sandbox_job_status",
			ContentType: "application/json",
			URL:         fmt.Sprintf("k8s://%s/jobs/%s", observation.Namespace, observation.JobName),
			SizeBytes:   int64(len(strings.Join(observation.JobConditions, "\n"))),
			Source:      "sandbox-runtime",
		},
	}
	logArtifactID := ""
	if strings.TrimSpace(observation.Logs) != "" && strings.TrimSpace(observation.PodName) != "" {
		logArtifactID = fmt.Sprintf("artifact-sandbox-log-%d", createdAt.UnixNano())
		artifacts = append(artifacts, events.Artifact{
			ID:          logArtifactID,
			TraceID:     traceID,
			Kind:        "sandbox_job_logs",
			ContentType: "text/plain",
			URL:         fmt.Sprintf("k8s://%s/pods/%s/log", observation.Namespace, observation.PodName),
			SizeBytes:   int64(len(observation.Logs)),
			Source:      "sandbox-runtime",
		})
	}
	return statusArtifactID, logArtifactID, artifacts
}

func sandboxFailureMessage(observation sandbox.JobObservation) string {
	parts := []string{
		fmt.Sprintf("Sandbox job %s failed in namespace %s.", observation.JobName, observation.Namespace),
	}
	if strings.TrimSpace(observation.PodName) != "" {
		parts = append(parts, fmt.Sprintf("pod=%s", observation.PodName))
	}
	if strings.TrimSpace(observation.PodPhase) != "" {
		parts = append(parts, fmt.Sprintf("phase=%s", observation.PodPhase))
	}
	if observation.ContainerExitCode != nil {
		parts = append(parts, fmt.Sprintf("exit_code=%d", *observation.ContainerExitCode))
	}
	if strings.TrimSpace(observation.TerminationReason) != "" {
		parts = append(parts, fmt.Sprintf("reason=%s", observation.TerminationReason))
	}
	return strings.Join(parts, " ")
}

func findProposal(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}

func filterProposalMemory(items []review.ProposalMemory, candidateKey string) []review.ProposalMemory {
	out := make([]review.ProposalMemory, 0)
	for _, item := range items {
		if item.CandidateKey == candidateKey {
			out = append(out, item)
		}
	}
	return out
}

func buildEvalRunnerTask(cfg config.Config, store storepkg.Store, trace events.Trace, run evals.Run, judgments []evals.Judgment, item queue.WorkItem) clients.RunnerTask {
	effectiveHarness := harness.ResolveEffectiveConfig(store, "eval", cfg.DefaultReasoningVerbosity)
	contextRefs := make([]map[string]any, 0, len(judgments)+1)
	contextRefs = append(contextRefs, map[string]any{
		"kind":      "eval_run",
		"ref":       run.ID,
		"summary":   run.OverallVerdict,
		"trace_id":  trace.Summary.TraceID,
		"suite":     run.SuiteName,
		"trigger":   run.Trigger,
		"work_kind": item.Kind,
	})
	for _, judgment := range judgments {
		contextRefs = append(contextRefs, map[string]any{
			"kind":     "eval_judgment",
			"ref":      judgment.ID,
			"layer":    judgment.Layer,
			"category": judgment.Category,
			"score":    judgment.Score,
			"summary":  judgment.Rationale,
		})
	}
	contextRefs = append(contextRefs, improvementTraceEvidenceRefs(trace)...)
	contextRefs = append(contextRefs, improvementCandidateEvidenceRefs(store, trace, "")...)
	contextRefs = append(contextRefs, improvementProposalMemoryRefs(store, "")...)
	prompt := fmt.Sprintf(
		"Summarize the completed eval for trace %s. Workflow=%s status=%s thread=%s. Eval run=%s suite=%s verdict=%s score=%.2f. Judgments=%v. Start from the supplied evidence pack, then use the read-only RSI tools when you need more trace, candidate, or proposal-memory detail. Explain what the eval found, what evidence mattered, whether improvement pressure should increase, and why.",
		trace.Summary.TraceID,
		trace.Summary.WorkflowKind,
		trace.Summary.Status,
		trace.Summary.ThreadKey,
		run.ID,
		run.SuiteName,
		run.OverallVerdict,
		run.OverallScore,
		judgmentDigest(judgments),
	)
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
		}
	}
	sessionScopeKind, sessionScopeID := evalSessionScope(store, trace, run)
	return clients.RunnerTask{
		TaskType:            "eval",
		Repo:                cfg.DefaultRepo,
		RepoRef:             "main",
		Prompt:              prompt,
		SystemMessage:       harness.ComposeSystemMessage("Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, and self_critique.", effectiveHarness),
		AllowedTools:        improvementReadOnlyTools(effectiveHarness),
		TimeoutSeconds:      300,
		ExpectedOutputs:     []string{"visible_reasoning", "final_answer"},
		ArtifactDestination: fmt.Sprintf("trace:%s:eval:%s", trace.Summary.TraceID, run.ID),
		ContextSummary: fmt.Sprintf(
			"Eval %s completed with verdict=%s score=%.2f across %d judgments. Terminal trace evidence, candidate lineage, and proposal memory are available in the evidence pack and read-only RSI tools.",
			run.ID,
			run.OverallVerdict,
			run.OverallScore,
			len(judgments),
		),
		Intent:                    trace.Summary.WorkflowKind,
		TraceID:                   trace.Summary.TraceID,
		WorkflowID:                trace.Summary.WorkflowID,
		ConversationID:            trace.Summary.ConversationID,
		CaseID:                    trace.Summary.CaseID,
		TriggerEventID:            trace.Summary.TriggerEventID,
		RecentConversationEntries: improvementRecentConversationEntries(store.ListConversationEntries(trace.Summary.ConversationID)),
		CaseSummary:               caseSummary,
		PriorTraceRefs:            improvementPriorTraceRefs(store.ListTraces(), trace.Summary.CaseID, trace.Summary.TraceID),
		RepoAllowlist:             cfg.AllowedTargetRepos,
		ToolAllowlist:             improvementReadOnlyTools(effectiveHarness),
		ResponseMode:              "analysis",
		ContextRefs:               contextRefs,
		ApprovalMode:              "deterministic",
		ReasoningVerbosity:        effectiveHarness.ReasoningVerbosity,
		SessionScopeKind:          sessionScopeKind,
		SessionScopeID:            sessionScopeID,
		HarnessProfileID:          effectiveHarness.Profile.ID,
		HarnessOverlayVersion:     effectiveHarness.EffectiveOverlayVersion,
		MemoryBackend:             harness.DefaultMemoryBackend,
		AssistantPeerID:           fmt.Sprintf("rsi:%s:eval", cfg.Environment),
		UserPeerID:                fmt.Sprintf("line:%s", sessionScopeID),
	}
}

func buildProposalRunnerTask(cfg config.Config, store storepkg.Store, trace events.Trace, proposal review.Proposal, attempt improvement.ChangeAttempt, memories []review.ProposalMemory) clients.RunnerTask {
	effectiveHarness := harness.ResolveEffectiveConfig(store, "proposal", cfg.DefaultReasoningVerbosity)
	rejectedContext := make([]map[string]any, 0, len(memories))
	for _, memory := range memories {
		rejectedContext = append(rejectedContext, map[string]any{
			"proposal_id":   memory.ProposalID,
			"disposition":   memory.Disposition,
			"rationale":     memory.ReviewRationale,
			"failure_class": firstNonEmpty(memory.FailureClass, strings.Join(memory.FailureClasses, ",")),
		})
	}
	contextRefs := []map[string]any{
		{
			"kind":          "proposal",
			"ref":           proposal.ID,
			"summary":       proposal.Summary,
			"candidate_key": proposal.CandidateKey,
			"risk_tier":     proposal.RiskTier,
			"scope":         proposal.ProposedScope,
			"target_layer":  proposal.TargetLayer,
			"target_kind":   proposal.TargetKind,
			"target_ref":    proposal.TargetRef,
		},
		{
			"kind":            "change_attempt",
			"ref":             attempt.ID,
			"attempt_number":  attempt.AttemptNumber,
			"branch_name":     attempt.BranchName,
			"parent_attempt":  attempt.ParentAttemptID,
			"failure_class":   attempt.FailureClass,
			"failure_summary": attempt.FailureSummary,
		},
	}
	contextRefs = append(contextRefs, improvementTraceEvidenceRefs(trace)...)
	contextRefs = append(contextRefs, improvementCandidateEvidenceRefs(store, trace, proposal.CandidateKey)...)
	contextRefs = append(contextRefs, improvementProposalMemoryRefs(store, proposal.CandidateKey)...)
	contextRefs = append(contextRefs, improvementAttemptHistoryRefs(store, proposal.ID, attempt.ID)...)
	prompt := fmt.Sprintf(
		"Materialize remediation attempt %d for approved proposal %s. Candidate=%s risk=%s scope=%s summary=%s. Start from the dense evidence pack: latest failing trace evidence, root-cause metadata, prior rejected or dismissed proposal memory, and prior attempt failures. Use the read-only RSI tools when you need more repo, trace, candidate, attempt, or proposal-memory detail. Return explicit visible reasoning plus a concrete change_plan, a non-empty unified repo_patch touching allowed repo files, a validation_plan, retry_assessment, and hypothesis_delta. Do not return metadata-only or .rsi-only diffs.",
		attempt.AttemptNumber,
		proposal.ID,
		proposal.CandidateKey,
		proposal.RiskTier,
		proposal.ProposedScope,
		proposal.Summary,
	)
	if proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		targetRole := firstNonEmpty(strings.TrimSpace(proposal.TargetRef), "prod")
		prompt = fmt.Sprintf(
			"Design remediation attempt %d as an approved runtime harness overlay for role %s from proposal %s. Candidate=%s summary=%s. Return explicit visible reasoning, a change_plan, validation_plan, retry_assessment, hypothesis_delta, and exactly one proposed action with kind=%q whose request_payload contains prompt_fragments, few_shot_snippets, tool_preference_order, retrieval_bias, reasoning_verbosity, memory_read_enabled, and memory_write_enabled. This is a runtime overlay activation, not a repo change.",
			attempt.AttemptNumber,
			targetRole,
			proposal.ID,
			proposal.CandidateKey,
			proposal.Summary,
			action.KindHarnessOverlay,
		)
	}
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
		}
	}
	return clients.RunnerTask{
		TaskType:                  "proposal",
		Repo:                      cfg.DefaultRepo,
		RepoRef:                   "main",
		Prompt:                    prompt,
		SystemMessage:             harness.ComposeSystemMessage("Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta, proposed_actions, knowledge_drafts, and outcome_hypotheses.", effectiveHarness),
		AllowedTools:              improvementReadOnlyTools(effectiveHarness),
		TimeoutSeconds:            420,
		ExpectedOutputs:           []string{"visible_reasoning", "final_answer", "change_plan", "validation_plan", "retry_assessment"},
		ArtifactDestination:       fmt.Sprintf("trace:%s:proposal:%s:attempt:%s", trace.Summary.TraceID, proposal.ID, attempt.ID),
		ContextSummary:            proposalRunnerContextSummary(proposal) + " Latest trace failures, candidate lineage, prior attempt failures, and proposal memory are preloaded; additional read-only RSI tools are available for evidence lookup.",
		RejectedProposalContext:   rejectedContext,
		Intent:                    trace.Summary.WorkflowKind,
		TraceID:                   trace.Summary.TraceID,
		WorkflowID:                trace.Summary.WorkflowID,
		ConversationID:            trace.Summary.ConversationID,
		CaseID:                    trace.Summary.CaseID,
		TriggerEventID:            trace.Summary.TriggerEventID,
		RecentConversationEntries: improvementRecentConversationEntries(store.ListConversationEntries(trace.Summary.ConversationID)),
		CaseSummary:               caseSummary,
		PriorTraceRefs:            improvementPriorTraceRefs(store.ListTraces(), trace.Summary.CaseID, trace.Summary.TraceID),
		RepoAllowlist:             cfg.AllowedTargetRepos,
		ToolAllowlist:             improvementReadOnlyTools(effectiveHarness),
		ResponseMode:              "analysis",
		ContextRefs:               contextRefs,
		ApprovalMode:              "human_review",
		ReasoningVerbosity:        effectiveHarness.ReasoningVerbosity,
		SessionScopeKind:          "proposal_candidate",
		SessionScopeID:            proposal.CandidateKey,
		HarnessProfileID:          effectiveHarness.Profile.ID,
		HarnessOverlayVersion:     effectiveHarness.EffectiveOverlayVersion,
		MemoryBackend:             harness.DefaultMemoryBackend,
		AssistantPeerID:           fmt.Sprintf("rsi:%s:proposal", cfg.Environment),
		UserPeerID:                fmt.Sprintf("candidate:%s", proposal.CandidateKey),
	}
}

func evalSessionScope(store storepkg.Store, trace events.Trace, run evals.Run) (string, string) {
	for _, candidate := range store.ListCandidates() {
		for _, evalID := range candidate.SourceEvalIDs {
			if evalID == run.ID {
				return "eval_line", firstNonEmpty(candidate.Subsystem, "unknown") + ":" + firstNonEmpty(candidate.FailureMode, candidate.CandidateKey)
			}
		}
	}
	return "eval_line", "trace:" + trace.Summary.TraceID
}

func improvementReadOnlyTools(effective harness.EffectiveConfig) []string {
	return harness.ApplyToolPreference([]string{
		"repo.context",
		"knowledge.context",
		"github.repo_activity",
		"github.repo_context",
		"sentry.lookup",
		"kubernetes.inspect",
		"kubernetes.logs",
		"kubernetes.events",
		"cloudflare.inspect",
		"rsi.trace_context",
		"rsi.proposal_memory",
		"rsi.candidate_context",
		"rsi.attempt_context",
	}, effective.ToolPreferenceOrder)
}

func improvementTraceEvidenceRefs(trace events.Trace) []map[string]any {
	eventRefs := make([]map[string]any, 0, minInt(len(trace.Events), 12))
	for _, item := range tailTraceEvents(trace.Events, 12) {
		eventRefs = append(eventRefs, map[string]any{
			"kind":        "trace_event",
			"ref":         item.EventType,
			"status":      item.Status,
			"plane":       item.Plane,
			"service":     item.Service,
			"description": item.Description,
		})
	}
	reasoningRefs := make([]map[string]any, 0, minInt(len(trace.Reasoning), 8))
	for _, item := range tailReasoning(trace.Reasoning, 8) {
		reasoningRefs = append(reasoningRefs, map[string]any{
			"kind":       "reasoning_step",
			"ref":        item.ID,
			"step_type":  item.StepType,
			"summary":    item.Summary,
			"decision":   item.Decision,
			"confidence": item.Confidence,
		})
	}
	return append(eventRefs, reasoningRefs...)
}

func improvementCandidateEvidenceRefs(store storepkg.Store, trace events.Trace, candidateKey string) []map[string]any {
	refs := make([]map[string]any, 0)
	for _, item := range store.ListCandidates() {
		if candidateKey != "" && item.CandidateKey != candidateKey {
			continue
		}
		if candidateKey == "" && item.LatestTraceID != trace.Summary.TraceID && !containsString(item.EvidenceTraceIDs, trace.Summary.TraceID) {
			continue
		}
		refs = append(refs, map[string]any{
			"kind":                        "candidate",
			"ref":                         item.CandidateKey,
			"subsystem":                   item.Subsystem,
			"failure_mode":                item.FailureMode,
			"target_layer":                item.TargetLayer,
			"priority_score":              item.PriorityScore,
			"retryable_failure_class":     item.RetryableFailureClass,
			"attempt_count":               item.AttemptCount,
			"auto_retry_budget_remaining": item.AutoRetryBudgetRemaining,
		})
	}
	return refs
}

func improvementProposalMemoryRefs(store storepkg.Store, candidateKey string) []map[string]any {
	refs := make([]map[string]any, 0)
	for _, item := range store.ListProposalMemories() {
		if candidateKey != "" && item.CandidateKey != candidateKey {
			continue
		}
		refs = append(refs, map[string]any{
			"kind":          "proposal_memory",
			"ref":           item.ID,
			"proposal_id":   item.ProposalID,
			"disposition":   item.Disposition,
			"failure_class": firstNonEmpty(item.FailureClass, strings.Join(item.FailureClasses, ",")),
			"rationale":     item.ReviewRationale,
			"hypothesis":    item.Hypothesis,
			"diff_summary":  item.DiffSummary,
		})
		if len(refs) == 8 {
			break
		}
	}
	return refs
}

func improvementAttemptHistoryRefs(store storepkg.Store, proposalID string, currentAttemptID string) []map[string]any {
	refs := make([]map[string]any, 0)
	for _, item := range store.ListChangeAttempts() {
		if item.ProposalID != proposalID || item.ID == currentAttemptID {
			continue
		}
		refs = append(refs, map[string]any{
			"kind":                       "change_attempt_history",
			"ref":                        item.ID,
			"attempt_number":             item.AttemptNumber,
			"state":                      item.State,
			"failure_class":              item.FailureClass,
			"failure_summary":            item.FailureSummary,
			"retry_decision":             item.RetryDecision,
			"material_hypothesis_change": item.MaterialHypothesisChange,
			"changed_files":              item.ChangedFiles,
		})
	}
	return refs
}

func tailTraceEvents(values []events.TraceEvent, limit int) []events.TraceEvent {
	if len(values) <= limit {
		return append([]events.TraceEvent(nil), values...)
	}
	return append([]events.TraceEvent(nil), values[len(values)-limit:]...)
}

func tailReasoning(values []events.ReasoningStep, limit int) []events.ReasoningStep {
	if len(values) <= limit {
		return append([]events.ReasoningStep(nil), values...)
	}
	return append([]events.ReasoningStep(nil), values[len(values)-limit:]...)
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

func stringFromAny(raw any) string {
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func improvementRecentConversationEntries(items []conversation.Entry) []map[string]any {
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

func improvementPriorTraceRefs(items []events.TraceSummary, caseID string, currentTraceID string) []map[string]any {
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

func judgmentDigest(items []evals.Judgment) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("%s/%s=%.2f (%s)", item.Layer, item.Category, item.Score, item.Rationale))
	}
	return out
}

func repoChangeCommands() []string {
	script := `
set -euo pipefail
mkdir -p /workspace
cd /workspace
rm -rf repo
git clone "https://x-access-token:${GITHUB_TOKEN}@github.com/${GITHUB_OWNER}/${RSI_REPO}.git" repo
cd repo
git checkout -B "${RSI_BRANCH_NAME}" "origin/${RSI_BASE_REF}"
if [ -z "${RSI_REPO_PATCH:-}" ]; then
  echo "RSI_REPO_PATCH is required" >&2
  exit 1
fi
mkdir -p .rsi
printf "%s\n" "${RSI_CONTEXT_SUMMARY}" > .rsi/proposal-context.txt
printf "%s\n" "${RSI_CHANGE_PLAN:-}" > .rsi/change-plan.txt
printf "%s\n" "${RSI_VALIDATION_PLAN:-}" > .rsi/validation-plan.txt
printf "%s\n" "${RSI_REPO_PATCH}" > /tmp/rsi-change.patch
git apply --check /tmp/rsi-change.patch
git apply /tmp/rsi-change.patch
if [ -z "$(git status --short)" ]; then
  echo "Patch produced no working tree changes" >&2
  exit 1
fi
if ! git status --short | awk '{print $2}' | grep -qv '^\.rsi/'; then
  echo "Patch only changed .rsi metadata files" >&2
  exit 1
fi
git config user.name "${GITHUB_COMMIT_USER}"
git config user.email "${GITHUB_COMMIT_EMAIL}"
make test
git add -A
git commit -m "fix: RSI proposal ${RSI_PROPOSAL_ID} attempt ${RSI_ATTEMPT_ID}" || true
git push origin HEAD
`
	return []string{"bash", "-lc", script}
}

func persistImprovementKnowledgeDrafts(store storepkg.Store, drafts []runnerutil.KnowledgeDraft, trace events.Trace, proposalID string, createdAt time.Time) error {
	for _, item := range drafts {
		scopeType := knowledge.ScopeType(firstNonEmpty(item.ScopeType, string(knowledge.ScopeCase)))
		scopeID := firstNonEmpty(item.ScopeID, trace.Summary.CaseID)
		if scopeType == knowledge.ScopeConversation {
			scopeID = firstNonEmpty(item.ScopeID, trace.Summary.ConversationID)
		}
		if scopeType == knowledge.ScopeGlobal {
			scopeID = ""
		}
		links := improvementEvidenceLinksFromDraft(item)
		if proposalID != "" {
			links = append(links, knowledge.EvidenceLink{
				EvidenceType:     "proposal",
				EvidenceID:       proposalID,
				RelevanceSummary: proposalID,
				EvidenceRef:      events.EvidenceRef{Kind: "proposal", Ref: proposalID, Summary: proposalID},
			})
		}
		if _, err := store.UpsertKnowledgeEntry(knowledge.Entry{
			Tier:       knowledge.TierWorking,
			Kind:       knowledge.Kind(firstNonEmpty(item.Kind, string(knowledge.KindFact))),
			ScopeType:  scopeType,
			ScopeID:    scopeID,
			Title:      item.Title,
			Summary:    item.Summary,
			Body:       item.Body,
			Status:     knowledge.StatusDraft,
			Confidence: item.Confidence,
			FreshUntil: parseTimeOrNil(item.FreshUntil),
			SourceType: knowledge.SourceAgent,
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}, links); err != nil {
			return err
		}
	}
	return nil
}

func improvementEvidenceLinksFromDraft(item runnerutil.KnowledgeDraft) []knowledge.EvidenceLink {
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

func improvementOutcomeHypothesisReasoning(trace events.Trace, items []runnerutil.OutcomeHypothesis, createdAt time.Time) []events.ReasoningStep {
	out := make([]events.ReasoningStep, 0, len(items))
	for idx, item := range items {
		out = append(out, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-outcome-%d-%d", createdAt.UnixNano(), idx),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "outcome_hypothesis",
			Summary:    firstNonEmpty(item.SuccessCondition, string(item.OutcomeType)),
			Alternatives: []string{
				firstNonEmpty(item.MeasurementRef, "manual_review"),
				firstNonEmpty(item.ExpectedTimeHorizon, "unspecified"),
			},
			Confidence: 0.7,
			Decision:   string(item.OutcomeType),
			CreatedAt:  createdAt,
		})
	}
	return out
}

func upsertImprovementActionIntent(store storepkg.Store, intent action.Intent) (action.Intent, error) {
	if strings.TrimSpace(intent.ID) == "" {
		if existing, ok := findMatchingActionIntent(store.ListActionIntents(), intent); ok {
			intent.ID = existing.ID
			if intent.CreatedAt.IsZero() {
				intent.CreatedAt = existing.CreatedAt
			}
		}
	}
	now := time.Now().UTC()
	if intent.CreatedAt.IsZero() {
		intent.CreatedAt = now
	}
	if intent.UpdatedAt.IsZero() {
		intent.UpdatedAt = now
	}
	intent.EvidenceRefs = normalizeImprovementEvidenceRefs(intent.EvidenceRefs, intent.TraceID, intent.ProposalID)
	return store.UpsertActionIntent(intent)
}

func findMatchingActionIntent(items []action.Intent, candidate action.Intent) (action.Intent, bool) {
	for _, item := range items {
		if candidate.ID != "" && item.ID == candidate.ID {
			return item, true
		}
		if candidate.IdempotencyKey != "" && item.IdempotencyKey == candidate.IdempotencyKey && item.Kind == candidate.Kind {
			return item, true
		}
		if candidate.ProposalID != "" && item.ProposalID == candidate.ProposalID && item.Kind == candidate.Kind && item.TargetRef == candidate.TargetRef {
			return item, true
		}
	}
	return action.Intent{}, false
}

func findProposalActionIntent(store storepkg.Store, proposalID string, kind action.Kind) (action.Intent, bool) {
	for _, item := range store.ListActionIntents() {
		if item.ProposalID == proposalID && item.Kind == kind {
			return item, true
		}
	}
	return action.Intent{}, false
}

func enqueueImprovementOperationWork(store storepkg.Store, op operation.Execution, item queue.WorkItem) (queue.WorkItem, error) {
	created, _, err := store.GetOrCreateOperation(op)
	if err != nil {
		return queue.WorkItem{}, err
	}
	item.OperationID = created.ID
	return store.EnqueueWorkItem(item)
}

func withActionStatus(intent action.Intent, status action.Status, updatedAt time.Time) action.Intent {
	intent.Status = status
	intent.UpdatedAt = updatedAt
	return intent
}

func normalizeImprovementEvidenceRefs(items []events.EvidenceRef, traceID string, proposalID string) []events.EvidenceRef {
	if len(items) > 0 {
		return items
	}
	out := make([]events.EvidenceRef, 0, 2)
	if proposalID != "" {
		out = append(out, events.EvidenceRef{Kind: "proposal", Ref: proposalID, Summary: proposalID})
	}
	if traceID != "" {
		out = append(out, events.EvidenceRef{Kind: "trace", Ref: traceID, Summary: traceID})
	}
	return out
}

func firstProposedActionRationale(items []runnerutil.ProposedAction, kind action.Kind) string {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Kind), string(kind)) {
			return item.Rationale
		}
	}
	return ""
}

func proposedActionByKind(items []runnerutil.ProposedAction, kind action.Kind) (runnerutil.ProposedAction, bool) {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Kind), string(kind)) {
			return item, true
		}
	}
	return runnerutil.ProposedAction{}, false
}

func proposalRunnerContextSummary(proposal review.Proposal) string {
	if proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		return fmt.Sprintf("Approved proposal %s for candidate %s is activating a runtime harness overlay for %s.", proposal.ID, proposal.CandidateKey, firstNonEmpty(proposal.TargetRef, "prod"))
	}
	return fmt.Sprintf("Approved proposal %s for candidate %s is entering repo-change queue.", proposal.ID, proposal.CandidateKey)
}

func buildHarnessOverlayFromRunner(store storepkg.Store, proposal review.Proposal, output runnerutil.StructuredOutput) (harness.Overlay, error) {
	targetRole := firstNonEmpty(strings.TrimSpace(proposal.TargetRef), "prod")
	profile, ok := store.GetHarnessProfile(harness.DefaultProfileID(targetRole))
	if !ok {
		return harness.Overlay{}, fmt.Errorf("harness profile for role %s not found", targetRole)
	}
	actionSpec, ok := proposedActionByKind(output.ProposedActions, action.KindHarnessOverlay)
	if !ok {
		return harness.Overlay{}, fmt.Errorf("proposal runner did not return %s action for overlay target", action.KindHarnessOverlay)
	}
	payload := actionSpec.RequestPayload
	if payload == nil {
		return harness.Overlay{}, fmt.Errorf("proposal runner returned empty overlay payload")
	}
	overlay := harness.Overlay{
		ID:                  fmt.Sprintf("overlay-%s", proposal.ID),
		ProfileID:           profile.ID,
		Role:                targetRole,
		Version:             firstNonEmpty(stringFromAny(payload["version"]), proposal.ID),
		Status:              harness.OverlayStatusActive,
		TargetKind:          proposal.TargetKind,
		TargetRef:           firstNonEmpty(proposal.TargetRef, targetRole),
		ProposalID:          proposal.ID,
		PromptFragments:     stringSliceFromAny(payload["prompt_fragments"]),
		FewShotSnippets:     stringSliceFromAny(payload["few_shot_snippets"]),
		ToolPreferenceOrder: stringSliceFromAny(payload["tool_preference_order"]),
		RetrievalBias:       firstNonEmpty(stringFromAny(payload["retrieval_bias"]), profile.RetrievalBias),
		ReasoningVerbosity:  firstNonEmpty(stringFromAny(payload["reasoning_verbosity"]), profile.ReasoningVerbosity),
		MemoryReadEnabled:   boolPointerFromAny(payload["memory_read_enabled"]),
		MemoryWriteEnabled:  boolPointerFromAny(payload["memory_write_enabled"]),
		CreatedBy:           "proposal-runner",
		ApprovedBy:          "improvement-plane",
	}
	if len(overlay.PromptFragments) == 0 && len(overlay.FewShotSnippets) == 0 && len(overlay.ToolPreferenceOrder) == 0 && overlay.RetrievalBias == "" && overlay.ReasoningVerbosity == "" && overlay.MemoryReadEnabled == nil && overlay.MemoryWriteEnabled == nil {
		return harness.Overlay{}, fmt.Errorf("proposal runner returned overlay action without any runtime-safe fields")
	}
	return overlay, nil
}

func stringSliceFromAny(raw any) []string {
	switch value := raw.(type) {
	case []string:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if trimmed := strings.TrimSpace(fmt.Sprint(item)); trimmed != "" && trimmed != "<nil>" {
				out = append(out, trimmed)
			}
		}
		return out
	default:
		return []string{}
	}
}

func boolPointerFromAny(raw any) *bool {
	switch value := raw.(type) {
	case bool:
		return &value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true":
			parsed := true
			return &parsed
		case "false":
			parsed := false
			return &parsed
		}
	}
	return nil
}

func boolPointerValue(raw *bool) bool {
	if raw == nil {
		return false
	}
	return *raw
}

func improvementActionStatus(result storepkg.ToolResult, execErr error) action.Status {
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

func improvementActionErrorCode(status action.Status) string {
	switch status {
	case action.StatusBlocked:
		return "blocked"
	case action.StatusFailed:
		return "failed"
	default:
		return ""
	}
}

func improvementActionError(result storepkg.ToolResult, execErr error) string {
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

func proposalRunnerBackoff(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	backoff := time.Duration(attempts) * 30 * time.Second
	if backoff > 5*time.Minute {
		return 5 * time.Minute
	}
	return backoff
}

func clonePayload(payload map[string]interface{}) map[string]interface{} {
	if payload == nil {
		return nil
	}
	out := make(map[string]interface{}, len(payload))
	for key, value := range payload {
		out[key] = value
	}
	return out
}

func parseTimeOrNil(value string) *time.Time {
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

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func stringValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", value)
	}
}

func nestedString(value interface{}, path ...string) string {
	current := value
	for _, key := range path {
		next, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current, ok = next[key]
		if !ok {
			return ""
		}
	}
	return stringValue(current)
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

func confidenceOr(fallback float64, candidate float64) float64 {
	if candidate > 0 {
		return candidate
	}
	return fallback
}

func runnerRuntimeLabel(resp clients.RunnerResponse) string {
	if resp.Provider == "" {
		return "unknown runtime"
	}
	model := strings.TrimSpace(stringValue(resp.Raw["model"]))
	effort := strings.TrimSpace(stringValue(resp.Raw["reasoning_effort"]))
	if model == "" {
		return resp.Provider
	}
	if effort == "" {
		return fmt.Sprintf("%s %s", resp.Provider, model)
	}
	return fmt.Sprintf("%s %s effort=%s", resp.Provider, model, effort)
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func ptrStatus(status events.Status) *events.Status {
	return &status
}
