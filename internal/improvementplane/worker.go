package improvementplane

import (
	"context"
	"crypto/sha1"
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
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

var (
	errDeferredWorkItem     = errors.New("work item deferred for retry")
	errDeferredEffect       = errors.New("effect deferred for retry")
	errProposalPhaseHandled = errors.New("proposal phase finalized by repository transition")
	errProposalPhaseFailed  = errors.New("proposal phase failed by repository transition")
)

func RunWorker(cfg config.Config, store storepkg.Store) error {
	workerID := fmt.Sprintf("%s-worker", cfg.ServiceName)
	runnerClients := map[string]runnerExecutor{
		"eval":     clients.NewRunnerClientWithTimeout(cfg.RunnerURLForRole("eval"), cfg.RunnerTimeoutForRole("eval")),
		"proposal": clients.NewRunnerClientWithTimeout(cfg.RunnerURLForRole("proposal"), cfg.RunnerTimeoutForRole("proposal")),
	}
	var toolClient toolExecutor = clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL)
	launcher, launcherErr := sandbox.NewLauncher(cfg)
	for {
		effect, ok, err := claimNextImprovementEffect(store, workerID, cfg.WorkItemLeaseDuration, cfg.SandboxPollInterval)
		if err != nil {
			return err
		}
		if ok {
			if err := processImprovementEffect(cfg, store, runnerClients["eval"], runnerClients["proposal"], toolClient, launcher, launcherErr, effect); err != nil {
				if errors.Is(err, errDeferredEffect) {
					continue
				}
				log.Printf("improvement-plane effect=%s kind=%s error=%v", effect.ID, effect.EffectKind, err)
			}
			continue
		}
		time.Sleep(cfg.WorkerPollInterval)
	}
}

func processImprovementEffect(cfg config.Config, store storepkg.Store, evalRunner runnerExecutor, proposalRunner runnerExecutor, toolClient toolExecutor, launcher sandbox.Launcher, launcherErr error, effect transition.EffectExecution) error {
	switch effect.MachineKind {
	case transition.MachineProblemLine:
		if effect.EffectKind == transition.EffectInvokeRunner {
			return processProblemLineEvalEffect(cfg, store, evalRunner, effect)
		}
	case transition.MachineAttempt:
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace:
			return processWorkspaceOpenEffect(cfg, store, launcher, launcherErr, effect)
		case transition.EffectInvokeRunner:
			return processImplementAttemptEffect(cfg, store, proposalRunner, toolClient, effect)
		case transition.EffectWorkspaceValidate:
			return processWorkspaceValidateEffect(cfg, store, toolClient, effect)
		case transition.EffectObserveWorkspaceValidation:
			return processWorkspaceValidationObservationEffect(cfg, store, launcher, launcherErr, effect)
		case transition.EffectOpenDraftPR:
			return processDraftPROpenEffect(cfg, store, toolClient, effect)
		}
	}
	return failClaimedImprovementEffect(store, effect, fmt.Sprintf("unsupported improvement effect %s/%s", effect.MachineKind, effect.EffectKind))
}

func loadAttemptEffectContext(store storepkg.Store, effect transition.EffectExecution) (review.Proposal, improvement.ChangeAttempt, events.Trace, error) {
	attemptID := firstNonEmpty(strings.TrimSpace(effect.AggregateID), strings.TrimSpace(stringValue(effect.Payload["attempt_id"])))
	if attemptID == "" {
		return review.Proposal{}, improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("%s effect %s missing attempt_id", effect.EffectKind, effect.ID)
	}
	attempt, ok := store.GetChangeAttempt(attemptID)
	if !ok {
		return review.Proposal{}, improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("attempt %s not found", attemptID)
	}
	proposal, ok := findProposal(store.ListProposals(), attempt.ProposalID)
	if !ok {
		return review.Proposal{}, improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("proposal %s not found", attempt.ProposalID)
	}
	traceID := firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["trace_id"])), attempt.AttemptTraceID, proposal.TraceID)
	if traceID == "" {
		return review.Proposal{}, improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("%s effect %s missing trace_id", effect.EffectKind, effect.ID)
	}
	trace, ok := store.GetTrace(traceID)
	if !ok {
		return review.Proposal{}, improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("trace %s not found", traceID)
	}
	return proposal, attempt, trace, nil
}

func effectStringValue(effect transition.EffectExecution, key string) string {
	if value := strings.TrimSpace(stringValue(effect.Payload[key])); value != "" {
		return value
	}
	payload := effectPayloadMap(effect.Payload["work_item_payload"])
	return strings.TrimSpace(stringValue(payload[key]))
}

func effectStringSliceValue(effect transition.EffectExecution, key string) []string {
	if values := stringSliceFromAny(effect.Payload[key]); len(values) > 0 {
		return values
	}
	payload := effectPayloadMap(effect.Payload["work_item_payload"])
	return stringSliceFromAny(payload[key])
}

func workspaceCommandPayload(workspace improvement.AttemptWorkspace, job improvement.RepoChangeJob) map[string]any {
	repo := firstNonEmpty(workspace.Repo, job.Repo)
	baseRef := firstNonEmpty(workspace.BaseRef, job.BaseRef, "main")
	branchName := firstNonEmpty(workspace.BranchName, job.BranchName)
	namespace := firstNonEmpty(workspace.Namespace, job.SandboxNamespace)
	jobName := firstNonEmpty(workspace.JobName, job.SandboxJobName)
	podName := firstNonEmpty(workspace.PodName, job.SandboxPodName)
	validationRef := firstNonEmpty(job.ValidationRef, fmt.Sprintf("%s/%s", namespace, firstNonEmpty(podName, jobName)))
	return map[string]any{
		"workspace_id":        workspace.ID,
		"workspace_namespace": workspace.Namespace,
		"workspace_job_name":  workspace.JobName,
		"workspace_pod_name":  workspace.PodName,
		"repo":                repo,
		"base_ref":            baseRef,
		"branch_name":         branchName,
		"allowed_path_globs":  append([]string(nil), workspace.AllowedPathGlobs...),
		"job_id":              job.ID,
		"sandbox_namespace":   namespace,
		"sandbox_job_name":    jobName,
		"sandbox_pod_name":    podName,
		"validation_ref":      validationRef,
	}
}

func attemptFailurePayload(workspace *improvement.AttemptWorkspace, job *improvement.RepoChangeJob, failureSummary string) map[string]any {
	payload := map[string]any{}
	if workspace != nil {
		payload["workspace_id"] = workspace.ID
		payload["workspace_namespace"] = workspace.Namespace
		payload["workspace_job_name"] = workspace.JobName
		payload["workspace_pod_name"] = workspace.PodName
		payload["repo"] = workspace.Repo
		payload["base_ref"] = workspace.BaseRef
		payload["branch_name"] = workspace.BranchName
		if len(workspace.AllowedPathGlobs) > 0 {
			payload["allowed_path_globs"] = append([]string(nil), workspace.AllowedPathGlobs...)
		}
	}
	if job != nil {
		payload["job_id"] = job.ID
		payload["sandbox_namespace"] = firstNonEmpty(job.SandboxNamespace, payloadString(payload["workspace_namespace"]))
		payload["sandbox_job_name"] = firstNonEmpty(job.SandboxJobName, payloadString(payload["workspace_job_name"]))
		payload["sandbox_pod_name"] = firstNonEmpty(job.SandboxPodName, payloadString(payload["workspace_pod_name"]))
		payload["validation_ref"] = job.ValidationRef
		payload["validation_error"] = failureSummary
		payload["repo"] = firstNonEmpty(job.Repo, payloadString(payload["repo"]))
		payload["base_ref"] = firstNonEmpty(job.BaseRef, payloadString(payload["base_ref"]))
		payload["branch_name"] = firstNonEmpty(job.BranchName, payloadString(payload["branch_name"]))
		if len(job.AllowedPathGlobs) > 0 {
			payload["allowed_path_globs"] = append([]string(nil), job.AllowedPathGlobs...)
		}
	}
	return payload
}

func processWorkspaceOpenEffect(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, effect transition.EffectExecution) error {
	proposal, attempt, trace, err := loadAttemptEffectContext(store, effect)
	if err != nil {
		return failClaimedImprovementEffect(store, effect, err.Error())
	}
	workspace, ready, err := ensureAttemptWorkspace(cfg, store, launcher, launcherErr, proposal, attempt, trace.Summary.TraceID)
	if err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	job, err := ensureWorkspaceRepoChangeJob(store, proposal, attempt, workspace)
	if err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	now := time.Now().UTC()
	payload := workspaceCommandPayload(workspace, job)
	if !ready {
		if proposal.Status != review.ProposalRepoChangeQueued {
			if err := submitProposalCommand(
				store,
				proposal,
				transition.CommandProposalMarkRepoChangeQueued,
				cfg.ServiceName,
				now,
				fmt.Sprintf("cmd-proposal-repo-change-queued:%s:%s", proposal.ID, attempt.ID),
				fmt.Sprintf("Workspace %s is still initializing for attempt %s.", workspace.ID, attempt.ID),
			); err != nil {
				_ = failClaimedImprovementEffect(store, effect, err.Error())
				return err
			}
		}
		if err := submitAttemptCommand(store, attempt, transition.CommandWorkspaceOpenDeferred, cfg.ServiceName, now, payload); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedImprovementEffect(store, effect, workspace.ID)
	}
	payload["trace_events"] = []events.TraceEvent{{
		TraceID:     trace.Summary.TraceID,
		IngestionID: trace.Summary.IngestionID,
		WorkflowID:  trace.Summary.WorkflowID,
		Plane:       "improvement",
		Service:     cfg.ServiceName,
		Actor:       "attempt-supervisor",
		EventType:   "workspace.ready",
		Status:      events.StatusQueued,
		StartedAt:   now,
		Description: fmt.Sprintf("Workspace %s is ready for attempt %s.", workspace.ID, attempt.ID),
	}}
	payload["reasoning_steps"] = []events.ReasoningStep{{
		ID:         fmt.Sprintf("reason-workspace-ready-%d", now.UnixNano()),
		TraceID:    trace.Summary.TraceID,
		WorkflowID: trace.Summary.WorkflowID,
		StepType:   "workspace_ready",
		Summary:    fmt.Sprintf("Workspace %s is ready and implementation can start.", workspace.ID),
		Confidence: 0.9,
		Decision:   workspace.ID,
		CreatedAt:  now,
	}}
	if proposal.Status != review.ProposalRepoChangeQueued && proposal.Status != review.ProposalRepoChangeRunning {
		if err := submitProposalCommand(
			store,
			proposal,
			transition.CommandProposalMarkRepoChangeQueued,
			cfg.ServiceName,
			now,
			fmt.Sprintf("cmd-proposal-repo-change-queued:%s:%s", proposal.ID, attempt.ID),
			fmt.Sprintf("Workspace %s opened for attempt %s.", workspace.ID, attempt.ID),
		); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
	}
	if err := submitAttemptCommand(store, attempt, transition.CommandWorkspaceReady, cfg.ServiceName, now, payload); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	if proposal.Status != review.ProposalRepoChangeRunning {
		if err := submitProposalCommand(
			store,
			proposal,
			transition.CommandProposalMarkRepoChangeRunning,
			cfg.ServiceName,
			now,
			fmt.Sprintf("cmd-proposal-repo-change-running:%s:%s", proposal.ID, attempt.ID),
			fmt.Sprintf("Workspace %s is ready for implementation.", workspace.ID),
		); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
	}
	return completeClaimedImprovementEffect(store, effect, workspace.ID)
}

func processImplementAttemptEffect(cfg config.Config, store storepkg.Store, runnerClient runnerExecutor, toolClient toolExecutor, effect transition.EffectExecution) error {
	proposal, attempt, attemptTrace, err := loadAttemptEffectContext(store, effect)
	if err != nil {
		return failClaimedImprovementEffect(store, effect, err.Error())
	}
	var workspace *improvement.AttemptWorkspace
	if proposal.RecommendedInterventionKind != review.InterventionHarnessOverlay && proposal.TargetLayer != harness.TargetLayerHarnessOverlay {
		workspaceID := effectStringValue(effect, "workspace_id")
		if workspaceID != "" {
			if itemWorkspace, ok := store.GetAttemptWorkspace(workspaceID); ok {
				workspace = &itemWorkspace
			}
		}
		if workspace == nil {
			if itemWorkspace, ok := store.GetAttemptWorkspaceByAttempt(attempt.ID); ok {
				workspace = &itemWorkspace
			}
		}
		if workspace == nil {
			return failClaimedImprovementEffect(store, effect, fmt.Sprintf("attempt %s missing workspace for implement phase", attempt.ID))
		}
	}
	memories := filterProposalMemory(store.ListProposalMemories(), proposal.CandidateKey)
	operationID := strings.TrimSpace(stringValue(effect.Payload["operation_id"]))
	runnerStarted := time.Now().UTC()
	runnerStartPayload := map[string]any{
		"operation_id": operationID,
		"trace_events": []events.TraceEvent{{
			TraceID:     attemptTrace.Summary.TraceID,
			IngestionID: attemptTrace.Summary.IngestionID,
			WorkflowID:  attemptTrace.Summary.WorkflowID,
			Plane:       "execution",
			Service:     "runner",
			Actor:       "proposal",
			EventType:   "runner.started",
			Status:      events.StatusRunning,
			StartedAt:   runnerStarted,
			Description: fmt.Sprintf("Change attempt %s dispatched to proposal runner.", attempt.ID),
		}},
	}
	if workspace != nil {
		runnerStartPayload["workspace_id"] = workspace.ID
	}
	if err := submitAttemptCommand(store, attempt, transition.CommandAttemptRunnerStarted, cfg.ServiceName, runnerStarted, runnerStartPayload); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	var (
		runnerResp   clients.RunnerResponse
		runnerOutput runnerutil.StructuredOutput
		runnerErr    error
	)
	if runnerClient != nil {
		runnerTask := buildProposalRunnerTask(cfg, store, attemptTrace, proposal, attempt, workspace, memories)
		runnerResp, runnerErr = runnerClient.Execute(runnerTask)
		if runnerErr == nil && !runnerResp.OK {
			runnerErr = fmt.Errorf("proposal runner returned non-ok result: %s", strings.TrimSpace(runnerResp.Message))
		}
		if runnerErr == nil {
			if err := runnerutil.PersistHarnessExecution(
				store,
				runnerResp,
				"proposal",
				operationID,
				attemptTrace.Summary.TraceID,
				proposal.ID,
				runnerTask.HarnessProfileID,
				runnerTask.HarnessOverlayVersion,
				runnerTask.SessionScopeKind,
				runnerTask.SessionScopeID,
				runnerTask.ParentSessionScopeKind,
				runnerTask.ParentSessionScopeID,
			); err != nil {
				_ = failClaimedImprovementEffect(store, effect, err.Error())
				return err
			}
			var parseErr error
			runnerOutput, parseErr = runnerutil.ParseStructuredOutput(runnerResp)
			if parseErr != nil {
				_ = failClaimedImprovementEffect(store, effect, parseErr.Error())
				return parseErr
			}
		}
	}
	if runnerErr != nil {
		retryAt := time.Now().UTC().Add(proposalRunnerBackoff(effect.RetryCount))
		if err := submitAttemptCommand(store, attempt, transition.CommandImplementationDeferred, cfg.ServiceName, time.Now().UTC(), map[string]any{
			"operation_id": operationID,
			"retry_after":  retryAt.Format(time.RFC3339),
			"trace_events": []events.TraceEvent{{
				TraceID:     attemptTrace.Summary.TraceID,
				IngestionID: attemptTrace.Summary.IngestionID,
				WorkflowID:  attemptTrace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       "proposal",
				EventType:   "runner.failed",
				Status:      events.StatusFailed,
				StartedAt:   runnerStarted,
				EndedAt:     ptrTime(time.Now().UTC()),
				Description: fmt.Sprintf("Proposal runner failed; effect will be retried after %s: %v", retryAt.Format(time.RFC3339), runnerErr),
			}},
		}); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedImprovementEffect(store, effect, attempt.ID)
	}
	if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		if err := processHarnessOverlayProposal(cfg, store, attemptTrace, proposal, attempt, runnerResp, runnerOutput, runnerStarted); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedImprovementEffect(store, effect, attempt.ID)
	}
	if workspace == nil {
		return failClaimedImprovementEffect(store, effect, fmt.Sprintf("proposal %s missing workspace for repo-change execution", proposal.ID))
	}
	if refreshedTrace, ok := store.GetTrace(attemptTrace.Summary.TraceID); ok {
		attemptTrace = refreshedTrace
	}
	job, jobErr := ensureWorkspaceRepoChangeJob(store, proposal, attempt, *workspace)
	if jobErr != nil {
		_ = failClaimedImprovementEffect(store, effect, jobErr.Error())
		return jobErr
	}
	if workspaceMutationCallCount(attemptTrace) == 0 {
		if err := recordAttemptFailure(cfg, store, proposal, attempt, attemptTrace, "no_op_diff", "Proposal implement run completed without any write-capable workspace tool calls.", false, improvement.AttemptTriggerProposalApproved, attemptFailureTraceExtras{
			Payload: attemptFailurePayload(workspace, &job, "Proposal implement run completed without any write-capable workspace tool calls."),
		}); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedImprovementEffect(store, effect, attempt.ID)
	}
	diffResult, execErr := toolClient.Execute("workspace.git_diff", map[string]any{
		"trace_id":     attemptTrace.Summary.TraceID,
		"workspace_id": workspace.ID,
		"attempt_id":   attempt.ID,
	})
	if execErr != nil || diffResult.Status != "ok" {
		summary := firstNonEmpty(improvementActionError(diffResult, execErr), "Workspace diff inspection failed.")
		if err := recordAttemptFailure(cfg, store, proposal, attempt, attemptTrace, "sandbox_failure", summary, false, improvement.AttemptTriggerSandboxFailed, attemptFailureTraceExtras{
			Payload: attemptFailurePayload(workspace, &job, summary),
		}); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedImprovementEffect(store, effect, attempt.ID)
	}
	changedFiles := stringSliceFromAny(diffResult.Output["changed_files"])
	patch := stringValue(diffResult.Output["patch"])
	diffSummary := firstNonEmpty(stringValue(diffResult.Output["diff_summary"]), strings.TrimSpace(runnerOutput.ChangePlan), strings.TrimSpace(runnerOutput.ContextSummary))
	failureClass, failureSummary, changedFiles := validateRepoPatch(patch, changedFiles)
	attempt.ChangePlan = strings.TrimSpace(runnerOutput.ChangePlan)
	attempt.RepoPatch = strings.TrimSpace(patch)
	attempt.ValidationPlan = strings.TrimSpace(runnerOutput.ValidationPlan)
	attempt.HypothesisDelta = strings.TrimSpace(runnerOutput.HypothesisDelta)
	attempt.DiffSummary = diffSummary
	attempt.ValidationSummary = strings.TrimSpace(runnerOutput.ValidationPlan)
	attempt.ChangedFiles = changedFiles
	attempt.FailureClass = firstNonEmpty(runnerOutput.RetryAssessment.FailureClass, failureClass)
	attempt.FailureSummary = firstNonEmpty(runnerOutput.RetryAssessment.FailureSummary, failureSummary)
	attempt.RetryDecision = strings.TrimSpace(runnerOutput.RetryAssessment.RetryDecision)
	attempt.MaterialHypothesisChange = runnerOutput.RetryAssessment.MaterialHypothesisChange
	attempt.HeadSHA = firstNonEmpty(stringValue(diffResult.Output["head_sha"]), attempt.HeadSHA)
	attempt.UpdatedAt = time.Now().UTC()
	if failureClass != "" {
		if err := recordAttemptFailure(cfg, store, proposal, attempt, attemptTrace, failureClass, failureSummary, runnerOutput.RetryAssessment.MaterialHypothesisChange, improvement.AttemptTriggerProposalApproved, attemptFailureTraceExtras{
			Payload: attemptFailurePayload(workspace, &job, failureSummary),
		}); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedImprovementEffect(store, effect, attempt.ID)
	}
	now := time.Now().UTC()
	reasoning := []events.ReasoningStep{{
		ID:         fmt.Sprintf("reason-proposal-%d", now.UnixNano()),
		TraceID:    attemptTrace.Summary.TraceID,
		WorkflowID: attemptTrace.Summary.WorkflowID,
		StepType:   "workspace_implemented",
		Summary:    firstNonEmpty(runnerOutput.ChangePlan, runnerOutput.FinalAnswer, runnerOutput.ContextSummary, fmt.Sprintf("Approved proposal %s produced a workspace-backed implementation.", proposal.ID)),
		Confidence: confidenceOr(0.86, runnerOutput.Confidence),
		Decision:   workspace.BranchName,
		CreatedAt:  now,
	}}
	reasoning = append(reasoning, runnerutil.ToTraceReasoning(attemptTrace.Summary.TraceID, attemptTrace.Summary.WorkflowID, runnerOutput, now)...)
	reasoning = append(reasoning, improvementOutcomeHypothesisReasoning(attemptTrace, runnerOutput.OutcomeHypotheses, now)...)
	if err := persistImprovementKnowledgeDrafts(store, runnerOutput.KnowledgeDrafts, attemptTrace, proposal.ID, now); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	if strings.TrimSpace(runnerOutput.SelfCritique) != "" {
		reasoning = append(reasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-proposal-self-%d", now.UnixNano()),
			TraceID:    attemptTrace.Summary.TraceID,
			WorkflowID: attemptTrace.Summary.WorkflowID,
			StepType:   "self_critique",
			Summary:    runnerOutput.SelfCritique,
			Confidence: confidenceOr(0.86, runnerOutput.Confidence),
			CreatedAt:  now,
		})
	}
	if err := submitAttemptCommand(store, attempt, transition.CommandAttemptRunnerCompleted, cfg.ServiceName, now, map[string]any{
		"operation_id": operationID,
		"trace_events": []events.TraceEvent{{
			TraceID:     attemptTrace.Summary.TraceID,
			IngestionID: attemptTrace.Summary.IngestionID,
			WorkflowID:  attemptTrace.Summary.WorkflowID,
			Plane:       "execution",
			Service:     "runner",
			Actor:       "proposal",
			EventType:   "runner.completed",
			Status:      events.StatusCompleted,
			StartedAt:   runnerStarted,
			EndedAt:     ptrTime(now),
			Description: fmt.Sprintf("Proposal runner returned repo-change rationale using %s.", runnerRuntimeLabel(runnerResp)),
		}},
		"reasoning_steps": reasoning,
	}); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	prAction, ok := proposedActionByKind(runnerOutput.ProposedActions, action.KindDraftPROpen)
	completionPayload := workspaceCommandPayload(*workspace, job)
	completionPayload["operation_id"] = operationID
	completionPayload["change_plan"] = attempt.ChangePlan
	completionPayload["repo_patch"] = attempt.RepoPatch
	completionPayload["validation_plan"] = attempt.ValidationPlan
	completionPayload["hypothesis_delta"] = attempt.HypothesisDelta
	completionPayload["diff_summary"] = attempt.DiffSummary
	completionPayload["validation_summary"] = attempt.ValidationSummary
	completionPayload["changed_files"] = append([]string(nil), attempt.ChangedFiles...)
	completionPayload["head_sha"] = attempt.HeadSHA
	completionPayload["workspace_id"] = workspace.ID
	completionPayload["validation_command"] = workspaceValidationCommand(firstNonEmpty(runnerOutput.ValidationPlan, proposal.ValidationPlan))
	if ok {
		completionPayload["branch_name"] = firstNonEmpty(stringValue(prAction.RequestPayload["branch_name"]), workspace.BranchName, attempt.BranchName)
		completionPayload["base_ref"] = firstNonEmpty(stringValue(prAction.RequestPayload["base_ref"]), workspace.BaseRef, "main")
		completionPayload["title"] = firstNonEmpty(stringValue(prAction.RequestPayload["title"]), fmt.Sprintf("RSI proposal %s attempt %s for %s", proposal.ID, attempt.ID, proposalTargetRepo(cfg, proposal)))
		completionPayload["body"] = firstNonEmpty(stringValue(prAction.RequestPayload["body"]), fmt.Sprintf("Automated draft PR for proposal %s attempt %s after workspace validation.", proposal.ID, attempt.ID))
	}
	if err := submitAttemptCommand(store, attempt, transition.CommandImplementationCompleted, cfg.ServiceName, now, completionPayload); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	return completeClaimedImprovementEffect(store, effect, attempt.ID)
}

func processWorkspaceValidateEffect(cfg config.Config, store storepkg.Store, toolClient toolExecutor, effect transition.EffectExecution) error {
	proposal, attempt, attemptTrace, err := loadAttemptEffectContext(store, effect)
	if err != nil {
		return failClaimedImprovementEffect(store, effect, err.Error())
	}
	workspace, ok := store.GetAttemptWorkspaceByAttempt(attempt.ID)
	if !ok {
		return failClaimedImprovementEffect(store, effect, fmt.Sprintf("attempt %s missing workspace for validation phase", attempt.ID))
	}
	job, err := ensureWorkspaceRepoChangeJob(store, proposal, attempt, workspace)
	if err != nil {
		return failClaimedImprovementEffect(store, effect, err.Error())
	}
	operationID := strings.TrimSpace(stringValue(effect.Payload["operation_id"]))
	validationCommand := firstNonEmpty(effectStringValue(effect, "validation_command"), workspaceValidationCommand(firstNonEmpty(attempt.ValidationPlan, proposal.ValidationPlan)))
	validationStarted := time.Now().UTC()
	if err := submitAttemptCommand(store, attempt, transition.CommandValidationStarted, cfg.ServiceName, validationStarted, map[string]any{
		"operation_id":       operationID,
		"workspace_id":       workspace.ID,
		"job_id":             job.ID,
		"sandbox_namespace":  firstNonEmpty(job.SandboxNamespace, workspace.Namespace),
		"sandbox_job_name":   firstNonEmpty(job.SandboxJobName, workspace.JobName),
		"sandbox_pod_name":   firstNonEmpty(job.SandboxPodName, workspace.PodName),
		"validation_ref":     firstNonEmpty(job.ValidationRef, fmt.Sprintf("%s/%s", workspace.Namespace, firstNonEmpty(workspace.PodName, workspace.JobName))),
		"validation_summary": fmt.Sprintf("Running workspace validation for attempt %s with %q.", attempt.ID, validationCommand),
		"trace_events": []events.TraceEvent{{
			TraceID:     attemptTrace.Summary.TraceID,
			IngestionID: attemptTrace.Summary.IngestionID,
			WorkflowID:  attemptTrace.Summary.WorkflowID,
			Plane:       "execution",
			Service:     cfg.ServiceName,
			Actor:       "attempt-supervisor",
			EventType:   "workspace.validation.started",
			Status:      events.StatusRunning,
			StartedAt:   validationStarted,
			Description: fmt.Sprintf("Started workspace validation for attempt %s using %q.", attempt.ID, validationCommand),
		}},
	}); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	validationResult, execErr := toolClient.Execute("workspace.run_validation", map[string]any{
		"trace_id":     attemptTrace.Summary.TraceID,
		"workspace_id": workspace.ID,
		"attempt_id":   attempt.ID,
		"command":      validationCommand,
	})
	if execErr != nil || validationResult.Status != "ok" {
		summary := firstNonEmpty(improvementActionError(validationResult, execErr), "Workspace validation failed.")
		if err := recordAttemptFailure(cfg, store, proposal, attempt, attemptTrace, "sandbox_failure", summary, false, improvement.AttemptTriggerSandboxFailed, attemptFailureTraceExtras{
			Payload: attemptFailurePayload(&workspace, &job, summary),
		}); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedImprovementEffect(store, effect, attempt.ID)
	}
	branchName := firstNonEmpty(effectStringValue(effect, "branch_name"), workspace.BranchName, attempt.BranchName)
	baseRef := firstNonEmpty(effectStringValue(effect, "base_ref"), workspace.BaseRef, "main")
	title := effectStringValue(effect, "title")
	body := effectStringValue(effect, "body")
	if title == "" || body == "" {
		if err := recordAttemptFailure(cfg, store, proposal, attempt, attemptTrace, "insufficient_evidence", "Proposal implement run completed without requesting a governed draft PR open.", true, improvement.AttemptTriggerProposalApproved, attemptFailureTraceExtras{
			Payload: attemptFailurePayload(&workspace, &job, "Proposal implement run completed without requesting a governed draft PR open."),
		}); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
		return completeClaimedImprovementEffect(store, effect, attempt.ID)
	}
	now := time.Now().UTC()
	payload := workspaceCommandPayload(workspace, job)
	payload["operation_id"] = operationID
	payload["workspace_id"] = workspace.ID
	payload["job_id"] = job.ID
	payload["validation_summary"] = firstNonEmpty(stringValue(validationResult.Output["stdout"]), validationCommand)
	payload["log_artifact_id"] = job.LogArtifactID
	payload["repo"] = firstNonEmpty(job.Repo, workspace.Repo)
	payload["branch_name"] = branchName
	payload["base_ref"] = baseRef
	payload["title"] = title
	payload["body"] = body
	payload["trace_events"] = []events.TraceEvent{{
		TraceID:     attemptTrace.Summary.TraceID,
		IngestionID: attemptTrace.Summary.IngestionID,
		WorkflowID:  attemptTrace.Summary.WorkflowID,
		Plane:       "improvement",
		Service:     cfg.ServiceName,
		Actor:       "worker",
		EventType:   "workspace.validation.completed",
		Status:      events.StatusQueued,
		StartedAt:   now,
		Description: fmt.Sprintf("Validated workspace %s and queued governed draft PR open for branch %s.", workspace.ID, branchName),
	}}
	payload["trace_artifacts"] = []events.Artifact{
		{
			ID:          fmt.Sprintf("artifact-patch-%d", now.UnixNano()),
			TraceID:     attemptTrace.Summary.TraceID,
			Kind:        "repo_patch",
			ContentType: "text/x-diff",
			URL:         fmt.Sprintf("memory://attempt/%s/repo.patch", attempt.ID),
			SizeBytes:   int64(len(attempt.RepoPatch)),
			Source:      "improvement-plane",
		},
		{
			ID:          fmt.Sprintf("artifact-workspace-%d", now.UnixNano()),
			TraceID:     attemptTrace.Summary.TraceID,
			Kind:        "workspace_diff_summary",
			ContentType: "text/plain",
			URL:         fmt.Sprintf("memory://workspace/%s/diff.txt", workspace.ID),
			SizeBytes:   int64(len(attempt.DiffSummary)),
			Source:      "improvement-plane",
		},
	}
	payload["reasoning_steps"] = []events.ReasoningStep{{
		ID:         fmt.Sprintf("reason-workspace-validate-%d", now.UnixNano()),
		TraceID:    attemptTrace.Summary.TraceID,
		WorkflowID: attemptTrace.Summary.WorkflowID,
		StepType:   "workspace_validate",
		Summary:    fmt.Sprintf("Validated workspace-backed diff for attempt %s using %q.", attempt.ID, validationCommand),
		Confidence: 0.87,
		Decision:   branchName,
		CreatedAt:  now,
	}}
	if err := submitAttemptCommand(store, attempt, transition.CommandValidationCompleted, cfg.ServiceName, now, payload); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	if proposal.Status != review.ProposalRepoChangeRunning && proposal.Status != review.ProposalValidationPending {
		if err := submitProposalCommand(
			store,
			proposal,
			transition.CommandProposalMarkRepoChangeRunning,
			cfg.ServiceName,
			now,
			fmt.Sprintf("cmd-proposal-repo-change-running:%s:%s", proposal.ID, attempt.ID),
			fmt.Sprintf("Workspace validation completed for attempt %s.", attempt.ID),
		); err != nil {
			_ = failClaimedImprovementEffect(store, effect, err.Error())
			return err
		}
	}
	if err := submitProposalCommand(
		store,
		proposal,
		transition.CommandProposalMarkValidationPending,
		cfg.ServiceName,
		now,
		fmt.Sprintf("cmd-proposal-validation-pending:%s:%s", proposal.ID, attempt.ID),
		fmt.Sprintf("Validated workspace %s for branch %s.", workspace.ID, branchName),
	); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	return completeClaimedImprovementEffect(store, effect, attempt.ID)
}

func processProblemLineEvalEffect(cfg config.Config, store storepkg.Store, runnerClient runnerExecutor, effect transition.EffectExecution) error {
	traceID := firstNonEmpty(strings.TrimSpace(effect.AggregateID), strings.TrimSpace(stringValue(effect.Payload["trace_id"])))
	if traceID == "" {
		return failClaimedImprovementEffect(store, effect, "problem-line eval effect missing trace_id")
	}
	trace, ok := store.GetTrace(traceID)
	if !ok {
		return failClaimedImprovementEffect(store, effect, fmt.Sprintf("trace %s not found", traceID))
	}
	runID := strings.TrimSpace(stringValue(effect.Payload["eval_run_id"]))
	if runID == "" {
		return failClaimedImprovementEffect(store, effect, "problem-line eval effect missing eval_run_id")
	}
	run, ok := loadEvalRunByID(store, runID)
	if !ok {
		return failClaimedImprovementEffect(store, effect, fmt.Sprintf("eval run %s not found", runID))
	}
	if err := processEvalRun(cfg, store, runnerClient, trace, run, store.ListEvalJudgments(run.ID), effect.ID, firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["trigger"])), run.Trigger), strings.TrimSpace(stringValue(effect.Payload["operation_id"]))); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	return completeClaimedImprovementEffect(store, effect, run.ID)
}

func processEvalRun(cfg config.Config, store storepkg.Store, runnerClient runnerExecutor, trace events.Trace, run evals.Run, judgments []evals.Judgment, evalID string, evalTrigger string, operationID string) error {
	started := time.Now().UTC()
	runnerStarted := time.Now().UTC()
	var (
		runnerResp   clients.RunnerResponse
		runnerOutput runnerutil.StructuredOutput
		runnerErr    error
	)
	if runnerClient != nil {
		runnerResp, runnerErr = runnerClient.Execute(buildEvalRunnerTask(cfg, store, trace, run, judgments, queue.WorkItem{
			ID:          evalID,
			Kind:        evalTrigger,
			TraceID:     trace.Summary.TraceID,
			OperationID: operationID,
		}))
		if runnerErr == nil && !runnerResp.OK {
			runnerErr = fmt.Errorf("eval runner returned non-ok result: %s", strings.TrimSpace(runnerResp.Message))
		}
		if runnerErr == nil {
			if err := runnerutil.PersistHarnessExecution(
				store,
				runnerResp,
				"eval",
				operationID,
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
			var parseErr error
			runnerOutput, parseErr = runnerutil.ParseStructuredOutput(runnerResp)
			if parseErr != nil {
				return parseErr
			}
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
	if _, err := submitProblemLineTraceProjection(
		store,
		trace.Summary.TraceID,
		cfg.ServiceName,
		completed,
		fmt.Sprintf("cmd-problem-line:trace:%s:%s", trace.Summary.TraceID, evalID),
		storepkg.TraceUpdate{
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
					Description: fmt.Sprintf("Started eval run kind=%s.", evalTrigger),
				},
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
		},
	); err != nil {
		return err
	}
	if err := ensureApprovedProposalWork(store, trace, cfg.ServiceName); err != nil {
		return err
	}
	return nil
}

func processProposalItem(cfg config.Config, store storepkg.Store, runnerClient runnerExecutor, toolClient toolExecutor, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	if item.ProposalID == "" {
		return fmt.Errorf("proposal work item %s missing proposal_id", item.ID)
	}
	operationKind := resolveProposalOperationKind(store, item)
	if strings.TrimSpace(operationKind) == "" {
		return fmt.Errorf("unsupported proposal operation for item %s", item.ID)
	}
	proposal, ok := findProposal(store.ListProposals(), item.ProposalID)
	if !ok {
		return fmt.Errorf("proposal %s not found", item.ProposalID)
	}
	if !proposalStatusAllowsPhaseExecution(proposal.Status, operationKind) {
		return nil
	}
	if !review.ProposalExecutableIntervention(proposal.RecommendedInterventionKind) {
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
	switch operationKind {
	case proposalOperationLineActivate:
		return requireProposalPhaseTerminal(processProposalLineActivate(cfg, store, proposal, trace, item), proposalOperationLineActivate)
	case proposalOperationAttemptPlan:
		return requireProposalPhaseTerminal(processProposalAttemptPlan(cfg, store, proposal, trace, item), proposalOperationAttemptPlan)
	case proposalOperationWorkspaceOpen:
		return requireProposalPhaseTerminal(processProposalWorkspaceOpen(cfg, store, launcher, launcherErr, proposal, trace, item), proposalOperationWorkspaceOpen)
	case proposalOperationImplementAttempt:
		return requireProposalPhaseTerminal(processProposalImplementAttempt(cfg, store, runnerClient, toolClient, proposal, trace, item), proposalOperationImplementAttempt)
	case proposalOperationWorkspaceValidate:
		return requireProposalPhaseTerminal(processProposalWorkspaceValidate(cfg, store, toolClient, proposal, trace, item), proposalOperationWorkspaceValidate)
	default:
		return fmt.Errorf("unsupported proposal operation for item %s", item.ID)
	}
}

func proposalStatusAllowsPhaseExecution(status review.ProposalStatus, operationKind string) bool {
	switch strings.TrimSpace(operationKind) {
	case proposalOperationLineActivate, proposalOperationAttemptPlan:
		return status == review.ProposalApproved
	case proposalOperationWorkspaceOpen, proposalOperationImplementAttempt, proposalOperationWorkspaceValidate:
		switch status {
		case review.ProposalApproved,
			review.ProposalRepoChangeQueued,
			review.ProposalRepoChangeRunning,
			review.ProposalValidationPending:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func requireProposalPhaseTerminal(err error, phaseKind string) error {
	if err == nil {
		return fmt.Errorf("proposal phase %s returned without explicit repository finalization", phaseKind)
	}
	return err
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

func processWorkspaceValidationObservationEffect(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, effect transition.EffectExecution) error {
	item, err := sandboxObservationWorkItemForEffect(store, effect)
	if err != nil {
		return failClaimedImprovementEffect(store, effect, err.Error())
	}
	err = observeSandboxJob(cfg, store, launcher, launcherErr, item, func() error {
		return errDeferredEffect
	})
	switch {
	case err == nil:
		resultRef := firstNonEmpty(stringValue(item.Payload["job_id"]), fmt.Sprintf("%s/%s", stringValue(item.Payload["namespace"]), stringValue(item.Payload["job_name"])))
		return completeClaimedImprovementEffect(store, effect, resultRef)
	case errors.Is(err, errDeferredEffect):
		return errDeferredEffect
	default:
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
}

func processHarnessOverlayProposal(cfg config.Config, store storepkg.Store, trace events.Trace, proposal review.Proposal, attempt improvement.ChangeAttempt, runnerResp clients.RunnerResponse, runnerOutput runnerutil.StructuredOutput, runnerStarted time.Time) error {
	overlay, err := buildHarnessOverlayFromRunner(store, proposal, runnerOutput)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	operationID := ""
	if currentOp, ok := latestActiveAttemptOperation(store, attempt.ID); ok {
		operationID = currentOp.ID
	}
	intentTemplate := improvementActionIntentBase(
		cfg.ServiceName,
		proposal,
		trace,
		attempt.ID,
		action.KindHarnessOverlay,
		overlay.Role,
		action.StatusQueued,
		firstNonEmpty(firstProposedActionRationale(runnerOutput.ProposedActions, action.KindHarnessOverlay), runnerOutput.FinalAnswer, "Activated runtime harness overlay after human approval."),
		fmt.Sprintf("harness-overlay:%s", proposal.ID),
		map[string]any{
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
		[]events.EvidenceRef{
			{Kind: "proposal", Ref: proposal.ID, Summary: proposal.Title},
			{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
		},
		now,
	)
	intent, err := ensureImprovementActionIntent(store, intentTemplate)
	if err != nil {
		return err
	}
	if _, err := submitImprovementActionCommand(store, intent.ID, transition.CommandActionStart, cfg.ServiceName, now, map[string]any{
		"operation_id": operationID,
		"attempt_id":   attempt.ID,
	}); err != nil {
		return err
	}
	changePlan := firstNonEmpty(strings.TrimSpace(runnerOutput.ChangePlan), strings.TrimSpace(runnerOutput.FinalAnswer), proposal.Summary)
	validationPlan := strings.TrimSpace(runnerOutput.ValidationPlan)
	overlayPayload := map[string]any{
		"overlay_id":            overlay.ID,
		"prompt_fragments":      overlay.PromptFragments,
		"few_shot_snippets":     overlay.FewShotSnippets,
		"tool_preference_order": overlay.ToolPreferenceOrder,
		"retrieval_bias":        overlay.RetrievalBias,
		"reasoning_verbosity":   overlay.ReasoningVerbosity,
	}
	experiment := harness.Experiment{
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
	}
	if _, err := submitHarnessActivationCommand(
		store,
		overlay,
		experiment,
		cfg.ServiceName,
		now,
		fmt.Sprintf("cmd-harness-overlay:%s:%s", proposal.ID, attempt.ID),
	); err != nil {
		_, _ = submitImprovementActionCommand(store, intent.ID, transition.CommandActionFail, cfg.ServiceName, now, map[string]any{
			"operation_id":  operationID,
			"attempt_id":    attempt.ID,
			"executor":      cfg.ServiceName,
			"provider":      "rsi-platform",
			"provider_ref":  overlay.ID,
			"error_code":    "overlay_activation_failed",
			"error_message": err.Error(),
			"started_at":    now,
			"completed_at":  now,
		})
		return err
	}
	if _, err := submitImprovementActionCommand(store, intent.ID, transition.CommandActionSucceed, cfg.ServiceName, now, map[string]any{
		"operation_id": operationID,
		"attempt_id":   attempt.ID,
		"executor":     cfg.ServiceName,
		"provider":     "rsi-platform",
		"provider_ref": overlay.ID,
		"started_at":   now,
		"completed_at": now,
	}); err != nil {
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
	payload := map[string]any{
		"change_plan":     changePlan,
		"validation_plan": validationPlan,
		"overlay_payload": overlayPayload,
		"trace_events": []events.TraceEvent{
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
		"reasoning_steps": reasoning,
	}
	if operationID != "" {
		payload["operation_id"] = operationID
	}
	if err := submitAttemptCommand(store, attempt, transition.CommandOverlayActivated, cfg.ServiceName, now, payload); err != nil {
		return err
	}
	return submitProposalCommand(
		store,
		proposal,
		transition.CommandProposalMarkMerged,
		cfg.ServiceName,
		now,
		fmt.Sprintf("cmd-proposal-overlay-merged:%s:%s", proposal.ID, attempt.ID),
		fmt.Sprintf("Activated harness overlay %s for role %s.", overlay.Version, overlay.Role),
	)
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
	proposal, ok := findProposal(store.ListProposals(), item.ProposalID)
	if !ok {
		proposal = review.Proposal{
			ID:             item.ProposalID,
			ConversationID: repoJob.ConversationID,
			CaseID:         repoJob.CaseID,
		}
	}
	intent, err := ensureImprovementActionIntent(store, improvementActionIntentBase(
		cfg.ServiceName,
		proposal,
		trace,
		attemptID,
		action.KindSandboxLaunch,
		fmt.Sprintf("%s/%s", cfg.SandboxNamespace, repoJob.ID),
		action.StatusQueued,
		"Launch the sandbox job to validate the approved repo change.",
		fmt.Sprintf("sandbox:%s", repoJob.ID),
		map[string]any{
			"job_id":      repoJob.ID,
			"attempt_id":  attemptID,
			"repo":        repoJob.Repo,
			"branch_name": repoJob.BranchName,
			"base_ref":    repoJob.BaseRef,
		},
		[]events.EvidenceRef{
			{Kind: "proposal", Ref: item.ProposalID, Summary: repoJob.CandidateKey},
			{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
		},
		time.Now().UTC(),
	))
	if err != nil {
		return err
	}
	started := time.Now().UTC()
	if _, err := submitImprovementActionCommand(store, intent.ID, transition.CommandActionStart, cfg.ServiceName, started, map[string]any{
		"operation_id": item.OperationID,
		"attempt_id":   attemptID,
	}); err != nil {
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
		if _, err := submitImprovementActionCommand(store, intent.ID, transition.CommandActionBlock, cfg.ServiceName, completed, map[string]any{
			"operation_id":   item.OperationID,
			"attempt_id":     attemptID,
			"executor":       "sandbox-runtime",
			"provider":       "kubernetes",
			"error_code":     "sandbox_unavailable",
			"error_message":  firstNonEmpty(errorString(launcherErr), "sandbox launcher not configured"),
			"started_at":     started,
			"completed_at":   completed,
			"policy_verdict": "sandbox_unavailable",
		}); err != nil {
			return err
		}
		if attempt.ID != "" && proposal.ID != "" {
			return recordAttemptFailure(cfg, store, proposal, attempt, trace, "sandbox_failure", firstNonEmpty(errorString(launcherErr), "Sandbox launcher unavailable."), false, improvement.AttemptTriggerSandboxFailed)
		}
		return nil
	}
	session, _, err := launcher.Launch(context.Background(), request)
	if err != nil {
		completed := time.Now().UTC()
		_, _ = submitImprovementActionCommand(store, intent.ID, transition.CommandActionFail, cfg.ServiceName, completed, map[string]any{
			"operation_id":  item.OperationID,
			"attempt_id":    attemptID,
			"executor":      "sandbox-runtime",
			"provider":      "kubernetes",
			"error_code":    "sandbox_launch_failed",
			"error_message": err.Error(),
			"started_at":    started,
			"completed_at":  completed,
		})
		if attempt.ID != "" && proposal.ID != "" {
			return recordAttemptFailure(cfg, store, proposal, attempt, trace, "sandbox_failure", err.Error(), false, improvement.AttemptTriggerSandboxFailed)
		}
		return err
	}
	return applySandboxLaunchSuccess(cfg, store, proposal, attempt, trace, repoJob, item, session)
}

func processSandboxWatch(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	return observeSandboxJob(cfg, store, launcher, launcherErr, item, func() error {
		return errDeferredWorkItem
	})
}

func observeSandboxJob(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem, onPending func() error) error {
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
		return onPending()
	}
	now := time.Now().UTC()
	statusArtifactID, logArtifactID, sandboxArtifacts := sandboxObservationArtifacts(trace.Summary.TraceID, observation, now)
	if observation.JobFailed {
		errorMessage := sandboxFailureMessage(observation)
		intent, _ := findAttemptActionIntent(store, item.ProposalID, firstNonEmpty(attemptID, attempt.ID), action.KindSandboxLaunch)
		if intent.ID != "" {
			if _, err := submitImprovementActionCommand(store, intent.ID, transition.CommandActionFail, cfg.ServiceName, now, map[string]any{
				"operation_id":         item.OperationID,
				"attempt_id":           firstNonEmpty(intent.AttemptID, attempt.ID),
				"executor":             "sandbox-runtime",
				"provider":             "kubernetes",
				"provider_ref":         fmt.Sprintf("%s/%s", observation.Namespace, observation.JobName),
				"response_artifact_id": statusArtifactID,
				"error_code":           "sandbox_job_failed",
				"error_message":        errorMessage,
				"started_at":           intent.UpdatedAt,
				"completed_at":         now,
			}); err != nil {
				return err
			}
		}
		if attempt.ID != "" && proposal.ID != "" {
			return recordAttemptFailure(cfg, store, proposal, attempt, trace, "sandbox_failure", errorMessage, false, improvement.AttemptTriggerSandboxFailed, attemptFailureTraceExtras{
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
				Payload: map[string]any{
					"job_id":            jobID,
					"sandbox_namespace": observation.Namespace,
					"sandbox_job_name":  observation.JobName,
					"sandbox_pod_name":  observation.PodName,
					"validation_ref":    fmt.Sprintf("%s/%s", observation.Namespace, observation.JobName),
					"log_artifact_id":   logArtifactID,
					"validation_error":  errorMessage,
				},
			})
		}
		if _, err := submitProblemLineTraceProjection(
			store,
			trace.Summary.TraceID,
			cfg.ServiceName,
			now,
			fmt.Sprintf("cmd-problem-line:trace:sandbox-failed:%s:%s", trace.Summary.TraceID, item.ID),
			storepkg.TraceUpdate{
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
			},
		); err != nil {
			return err
		}
		return nil
	}

	intent, _ := findAttemptActionIntent(store, item.ProposalID, firstNonEmpty(attemptID, attempt.ID), action.KindSandboxLaunch)
	if intent.ID != "" {
		if _, err := submitImprovementActionCommand(store, intent.ID, transition.CommandActionSucceed, cfg.ServiceName, now, map[string]any{
			"operation_id":         item.OperationID,
			"attempt_id":           firstNonEmpty(intent.AttemptID, attempt.ID),
			"executor":             "sandbox-runtime",
			"provider":             "kubernetes",
			"provider_ref":         fmt.Sprintf("%s/%s", observation.Namespace, observation.JobName),
			"response_artifact_id": statusArtifactID,
			"started_at":           intent.UpdatedAt,
			"completed_at":         now,
		}); err != nil {
			return err
		}
	}
	return applySandboxWatchSuccess(cfg, store, proposal, attempt, trace, item, repo, branchName, jobID, observation, sandboxArtifacts)
}

func sandboxObservationWorkItemForEffect(store storepkg.Store, effect transition.EffectExecution) (queue.WorkItem, error) {
	attemptID := firstNonEmpty(strings.TrimSpace(effect.AggregateID), strings.TrimSpace(stringValue(effect.Payload["attempt_id"])))
	if attemptID == "" {
		return queue.WorkItem{}, fmt.Errorf("observe_workspace_validation effect %s missing attempt_id", effect.ID)
	}
	attempt, ok := store.GetChangeAttempt(attemptID)
	if !ok {
		return queue.WorkItem{}, fmt.Errorf("attempt %s not found", attemptID)
	}
	proposal, ok := findProposal(store.ListProposals(), attempt.ProposalID)
	if !ok {
		return queue.WorkItem{}, fmt.Errorf("proposal %s not found", attempt.ProposalID)
	}
	jobID := strings.TrimSpace(stringValue(effect.Payload["job_id"]))
	job, jobOK := findRepoChangeJob(store.ListRepoChangeJobs(), jobID)
	if !jobOK {
		for _, candidate := range store.ListRepoChangeJobs() {
			if candidate.AttemptID == attempt.ID {
				job = candidate
				jobOK = true
				break
			}
		}
	}
	item := queue.WorkItem{
		ID:          effect.ID,
		OperationID: strings.TrimSpace(stringValue(effect.Payload["operation_id"])),
		Queue:       queue.SandboxQueue,
		Kind:        "watch_sandbox_job",
		Status:      queue.WorkQueued,
		TraceID: firstNonEmpty(
			strings.TrimSpace(stringValue(effect.Payload["trace_id"])),
			attempt.AttemptTraceID,
			proposal.TraceID,
		),
		ProposalID: proposal.ID,
		Payload: map[string]any{
			"attempt_id":  attempt.ID,
			"job_id":      firstNonEmpty(jobID, job.ID),
			"job_name":    firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["sandbox_job_name"])), job.SandboxJobName),
			"namespace":   firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["sandbox_namespace"])), job.SandboxNamespace),
			"repo":        firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["repo"])), job.Repo),
			"branch_name": firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["branch_name"])), job.BranchName),
			"base_ref":    firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["base_ref"])), job.BaseRef),
		},
	}
	if item.TraceID == "" {
		return queue.WorkItem{}, fmt.Errorf("observe_workspace_validation effect %s missing trace_id", effect.ID)
	}
	if strings.TrimSpace(stringValue(item.Payload["job_name"])) == "" || strings.TrimSpace(stringValue(item.Payload["namespace"])) == "" {
		return queue.WorkItem{}, fmt.Errorf("observe_workspace_validation effect %s missing sandbox job metadata", effect.ID)
	}
	return item, nil
}

func applySandboxLaunchSuccess(cfg config.Config, store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, trace events.Trace, repoJob improvement.RepoChangeJob, item queue.WorkItem, session sandbox.Session) error {
	now := time.Now().UTC()
	attemptTrace := attemptTraceForChange(store, trace, attempt)
	if attempt.ID != "" {
		if err := submitAttemptCommand(store, attempt, transition.CommandValidationStarted, cfg.ServiceName, now, map[string]any{
			"operation_id":       item.OperationID,
			"job_id":             repoJob.ID,
			"sandbox_namespace":  session.Namespace,
			"sandbox_job_name":   session.PodName,
			"sandbox_pod_name":   session.PodName,
			"validation_ref":     fmt.Sprintf("%s/%s", session.Namespace, session.PodName),
			"validation_summary": fmt.Sprintf("Sandbox validation started in %s/%s.", session.Namespace, session.PodName),
			"trace_events": []events.TraceEvent{
				{
					TraceID:     attemptTrace.Summary.TraceID,
					IngestionID: attemptTrace.Summary.IngestionID,
					WorkflowID:  attemptTrace.Summary.WorkflowID,
					Plane:       "execution",
					Service:     cfg.ServiceName,
					Actor:       "sandbox-launcher",
					EventType:   "sandbox.job.started",
					Status:      events.StatusRunning,
					StartedAt:   now,
					Description: fmt.Sprintf("Launched sandbox job %s in namespace %s.", session.PodName, session.Namespace),
				},
			},
			"trace_artifacts": []events.Artifact{
				{
					ID:          fmt.Sprintf("artifact-sandbox-launch-%d", now.UnixNano()),
					TraceID:     attemptTrace.Summary.TraceID,
					Kind:        "sandbox_job",
					ContentType: "text/plain",
					URL:         fmt.Sprintf("k8s://%s/jobs/%s", session.Namespace, session.PodName),
					SizeBytes:   0,
					Source:      "sandbox-runtime",
				},
			},
			"reasoning_steps": []events.ReasoningStep{
				{
					ID:         fmt.Sprintf("reason-sandbox-launch-%d", now.UnixNano()),
					TraceID:    attemptTrace.Summary.TraceID,
					WorkflowID: attemptTrace.Summary.WorkflowID,
					StepType:   "sandbox_launch",
					Summary:    fmt.Sprintf("Launched real sandbox job for repo %s branch %s.", repoJob.Repo, repoJob.BranchName),
					Confidence: 0.88,
					Decision:   session.PodName,
					CreatedAt:  now,
				},
			},
		}); err != nil {
			return err
		}
	}
	if proposal.ID != "" {
		if err := submitProposalCommand(
			store,
			proposal,
			transition.CommandProposalMarkRepoChangeRunning,
			cfg.ServiceName,
			now,
			fmt.Sprintf("cmd-proposal-repo-change-running:%s:%s", proposal.ID, attempt.ID),
			fmt.Sprintf("Sandbox validation started for branch %s.", repoJob.BranchName),
		); err != nil {
			return err
		}
	}
	return nil
}

func applySandboxWatchSuccess(cfg config.Config, store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, trace events.Trace, item queue.WorkItem, repo string, branchName string, jobID string, observation sandbox.JobObservation, sandboxArtifacts []events.Artifact) error {
	now := time.Now().UTC()
	attemptTrace := attemptTraceForChange(store, trace, attempt)
	baseRef := firstNonEmpty(stringValue(item.Payload["base_ref"]), "main")
	if attempt.ID != "" {
		for i := range sandboxArtifacts {
			sandboxArtifacts[i].TraceID = attemptTrace.Summary.TraceID
		}
		if err := submitAttemptCommand(store, attempt, transition.CommandValidationCompleted, cfg.ServiceName, now, map[string]any{
			"operation_id":       item.OperationID,
			"job_id":             jobID,
			"sandbox_namespace":  observation.Namespace,
			"sandbox_job_name":   observation.JobName,
			"sandbox_pod_name":   observation.PodName,
			"validation_ref":     fmt.Sprintf("%s/%s", observation.Namespace, observation.JobName),
			"log_artifact_id":    sandboxLogArtifactID(sandboxArtifacts),
			"validation_summary": fmt.Sprintf("Sandbox validation succeeded for %s.", observation.JobName),
			"repo":               repo,
			"branch_name":        branchName,
			"base_ref":           baseRef,
			"trace_events": []events.TraceEvent{
				{
					TraceID:     attemptTrace.Summary.TraceID,
					IngestionID: attemptTrace.Summary.IngestionID,
					WorkflowID:  attemptTrace.Summary.WorkflowID,
					Plane:       "improvement",
					Service:     cfg.ServiceName,
					Actor:       "worker",
					EventType:   "github.pr.queued",
					Status:      events.StatusQueued,
					StartedAt:   now,
					Description: fmt.Sprintf("Sandbox job %s succeeded; queued draft PR open for branch %s.", observation.JobName, branchName),
				},
			},
			"trace_artifacts": sandboxArtifacts,
			"reasoning_steps": []events.ReasoningStep{
				{
					ID:         fmt.Sprintf("reason-pr-open-%d", now.UnixNano()),
					TraceID:    attemptTrace.Summary.TraceID,
					WorkflowID: attemptTrace.Summary.WorkflowID,
					StepType:   "pr_queue",
					Summary:    fmt.Sprintf("Sandbox validation succeeded; queued draft PR open for branch %s.", branchName),
					Confidence: 0.9,
					Decision:   branchName,
					CreatedAt:  now,
				},
			},
		}); err != nil {
			return err
		}
	}
	if proposal.ID != "" {
		if err := submitProposalCommand(
			store,
			proposal,
			transition.CommandProposalMarkValidationPending,
			cfg.ServiceName,
			now,
			fmt.Sprintf("cmd-proposal-validation-pending:%s:%s", proposal.ID, attempt.ID),
			fmt.Sprintf("Sandbox validation succeeded for branch %s.", branchName),
		); err != nil {
			return err
		}
	}
	return nil
}

func processDraftPROpen(cfg config.Config, store storepkg.Store, toolClient toolExecutor, item queue.WorkItem) error {
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	repo := stringValue(item.Payload["repo"])
	branchName := stringValue(item.Payload["branch_name"])
	baseRef := firstNonEmpty(stringValue(item.Payload["base_ref"]), "main")
	jobID := stringValue(item.Payload["job_id"])
	attemptID := stringValue(item.Payload["attempt_id"])
	title := firstNonEmpty(stringValue(item.Payload["title"]), fmt.Sprintf("RSI proposal %s attempt %s for %s", item.ProposalID, attemptID, repo))
	body := firstNonEmpty(stringValue(item.Payload["body"]), fmt.Sprintf("Automated draft PR for proposal %s attempt %s after workspace validation.", item.ProposalID, attemptID))
	now := time.Now().UTC()
	proposal, _ := findProposal(store.ListProposals(), item.ProposalID)
	attempt, _ := store.GetChangeAttempt(firstNonEmpty(attemptID, proposal.CurrentAttemptID))
	intent, err := ensureImprovementActionIntent(store, improvementActionIntentBase(
		cfg.ServiceName,
		proposal,
		trace,
		attemptID,
		action.KindDraftPROpen,
		repo,
		action.StatusQueued,
		"Open a draft PR once sandbox validation succeeds.",
		fmt.Sprintf("pr:%s:%s", attemptID, branchName),
		map[string]any{
			"proposal_id": item.ProposalID,
			"attempt_id":  attemptID,
			"repo":        repo,
			"branch_name": branchName,
			"base_ref":    baseRef,
		},
		[]events.EvidenceRef{
			{Kind: "proposal", Ref: item.ProposalID, Summary: item.ProposalID},
			{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
		},
		now,
	))
	if err != nil {
		return err
	}
	if _, err := submitImprovementActionCommand(store, intent.ID, transition.CommandActionStart, cfg.ServiceName, now, map[string]any{
		"operation_id": item.OperationID,
		"attempt_id":   attemptID,
		"trace_events": []events.TraceEvent{
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
	}); err != nil {
		return err
	}
	prResult, execErr := toolClient.Execute("github.create_pr", map[string]any{
		"proposal_id": item.ProposalID,
		"attempt_id":  attemptID,
		"repo":        repo,
		"branch_name": branchName,
		"base_ref":    baseRef,
		"title":       title,
		"body":        body,
	})
	completed := time.Now().UTC()
	actionStatus := improvementActionStatus(prResult, execErr)
	commandKind, err := improvementActionCommandForStatus(actionStatus)
	if err != nil {
		return err
	}
	if _, err := submitImprovementActionCommand(store, intent.ID, commandKind, cfg.ServiceName, completed, map[string]any{
		"operation_id":  item.OperationID,
		"attempt_id":    attemptID,
		"executor":      "tool-gateway",
		"provider":      firstNonEmpty(prResult.Provider, "github"),
		"provider_ref":  prResult.ProviderRef,
		"error_code":    improvementActionErrorCode(actionStatus),
		"error_message": improvementActionError(prResult, execErr),
		"started_at":    now,
		"completed_at":  completed,
	}); err != nil {
		return err
	}
	if execErr != nil || actionStatus != action.StatusSucceeded {
		if attempt.ID != "" && proposal.ID != "" {
			return recordAttemptFailure(cfg, store, proposal, attempt, trace, "stale_branch", firstNonEmpty(improvementActionError(prResult, execErr), "Draft PR open blocked."), false, improvement.AttemptTriggerCIFailed)
		}
		return nil
	}
	prURL := stringValue(prResult.Output["pr_url"])
	headSHA := nestedString(prResult.Output, "response", "head", "sha")
	attemptTrace := trace
	if attempt.AttemptTraceID != "" {
		if itemTrace, found := store.GetTrace(attempt.AttemptTraceID); found {
			attemptTrace = itemTrace
		}
	}
	if attempt.ID != "" {
		if err := submitAttemptCommand(store, attempt, transition.CommandAttemptPROpened, cfg.ServiceName, completed, map[string]any{
			"operation_id": item.OperationID,
			"job_id":       jobID,
			"pr_url":       prURL,
			"head_sha":     headSHA,
			"repo":         repo,
			"branch_name":  branchName,
			"trace_events": []events.TraceEvent{
				{
					TraceID:     attemptTrace.Summary.TraceID,
					IngestionID: attemptTrace.Summary.IngestionID,
					WorkflowID:  attemptTrace.Summary.WorkflowID,
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
			"reasoning_steps": []events.ReasoningStep{
				{
					ID:         fmt.Sprintf("reason-pr-open-%d", completed.UnixNano()),
					TraceID:    attemptTrace.Summary.TraceID,
					WorkflowID: attemptTrace.Summary.WorkflowID,
					StepType:   "pr_opened",
					Summary:    fmt.Sprintf("Opened real draft PR for branch %s.", branchName),
					Confidence: 0.9,
					Decision:   prURL,
					CreatedAt:  completed,
				},
			},
		}); err != nil {
			return err
		}
	}
	if proposal.ID != "" {
		if err := submitProposalCommand(
			store,
			proposal,
			transition.CommandProposalMarkPROpen,
			cfg.ServiceName,
			completed,
			fmt.Sprintf("cmd-proposal-pr-open:%s:%s", proposal.ID, attemptID),
			fmt.Sprintf("Opened draft PR for %s on branch %s.", repo, branchName),
		); err != nil {
			return err
		}
	}
	return nil
}

func processDraftPROpenEffect(cfg config.Config, store storepkg.Store, toolClient toolExecutor, effect transition.EffectExecution) error {
	item, err := draftPROpenWorkItemForEffect(effect)
	if err != nil {
		return failClaimedImprovementEffect(store, effect, err.Error())
	}
	if err := processDraftPROpen(cfg, store, toolClient, item); err != nil {
		_ = failClaimedImprovementEffect(store, effect, err.Error())
		return err
	}
	if err := completeClaimedImprovementEffect(store, effect, item.ID); err != nil {
		return err
	}
	return nil
}

func draftPROpenWorkItemForEffect(effect transition.EffectExecution) (queue.WorkItem, error) {
	workItemID := strings.TrimSpace(stringValue(effect.Payload["work_item_id"]))
	payload := effectPayloadMap(effect.Payload["work_item_payload"])
	if len(payload) == 0 {
		payload = map[string]any{}
		for _, key := range []string{"attempt_id", "job_id", "job_name", "namespace", "repo", "branch_name", "base_ref", "title", "body"} {
			if value := strings.TrimSpace(stringValue(effect.Payload[key])); value != "" {
				payload[key] = value
			}
		}
	}
	item := queue.WorkItem{
		ID:             firstNonEmpty(workItemID, effect.ID),
		OperationID:    strings.TrimSpace(stringValue(effect.Payload["operation_id"])),
		Queue:          effectQueueName(stringValue(effect.Payload["queue"])),
		Kind:           firstNonEmpty(stringValue(effect.Payload["work_item_kind"]), "draft_pr_open"),
		Status:         queue.WorkQueued,
		TraceID:        strings.TrimSpace(stringValue(effect.Payload["trace_id"])),
		WorkflowID:     strings.TrimSpace(stringValue(effect.Payload["workflow_id"])),
		IngestionID:    strings.TrimSpace(stringValue(effect.Payload["ingestion_id"])),
		ConversationID: strings.TrimSpace(stringValue(effect.Payload["conversation_id"])),
		CaseID:         strings.TrimSpace(stringValue(effect.Payload["case_id"])),
		TriggerEventID: strings.TrimSpace(stringValue(effect.Payload["trigger_event_id"])),
		ProposalID:     firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["proposal_id"])), strings.TrimSpace(stringValue(payload["proposal_id"]))),
		ThreadKey:      strings.TrimSpace(stringValue(effect.Payload["thread_key"])),
		Intent:         strings.TrimSpace(stringValue(effect.Payload["intent"])),
		RepoScope:      strings.TrimSpace(stringValue(effect.Payload["repo_scope"])),
		RequestedBy:    strings.TrimSpace(stringValue(effect.Payload["requested_by"])),
		ApprovalMode:   strings.TrimSpace(stringValue(effect.Payload["approval_mode"])),
		ResponseMode:   strings.TrimSpace(stringValue(effect.Payload["response_mode"])),
		Payload:        payload,
	}
	if item.ProposalID == "" {
		return queue.WorkItem{}, fmt.Errorf("open_draft_pr effect %s missing proposal_id", effect.ID)
	}
	if item.TraceID == "" {
		return queue.WorkItem{}, fmt.Errorf("open_draft_pr effect %s missing trace_id", effect.ID)
	}
	if strings.TrimSpace(stringValue(item.Payload["attempt_id"])) == "" {
		return queue.WorkItem{}, fmt.Errorf("open_draft_pr effect %s missing attempt_id", effect.ID)
	}
	return item, nil
}

func proposalPhaseWorkItemForEffect(store storepkg.Store, effect transition.EffectExecution) (queue.WorkItem, bool, error) {
	workItemID := strings.TrimSpace(stringValue(effect.Payload["work_item_id"]))
	if workItemID != "" {
		if item, ok := workItemByID(store.ListWorkItems(), workItemID); ok {
			return item, true, nil
		}
	}
	operationID := strings.TrimSpace(stringValue(effect.Payload["operation_id"]))
	if operationID != "" {
		if item, ok := workItemByOperationID(store.ListWorkItems(), operationID); ok {
			return item, true, nil
		}
	}
	kind, err := proposalPhaseKindForEffect(effect)
	if err != nil {
		return queue.WorkItem{}, false, err
	}
	payload := effectPayloadMap(effect.Payload["work_item_payload"])
	item := queue.WorkItem{
		ID:             firstNonEmpty(workItemID, effect.ID),
		OperationID:    operationID,
		Queue:          queue.ProposalQueue,
		Kind:           kind,
		Status:         queue.WorkQueued,
		TraceID:        strings.TrimSpace(stringValue(effect.Payload["trace_id"])),
		WorkflowID:     strings.TrimSpace(stringValue(effect.Payload["workflow_id"])),
		IngestionID:    strings.TrimSpace(stringValue(effect.Payload["ingestion_id"])),
		ConversationID: strings.TrimSpace(stringValue(effect.Payload["conversation_id"])),
		CaseID:         strings.TrimSpace(stringValue(effect.Payload["case_id"])),
		TriggerEventID: strings.TrimSpace(stringValue(effect.Payload["trigger_event_id"])),
		ProposalID:     firstNonEmpty(strings.TrimSpace(stringValue(effect.Payload["proposal_id"])), strings.TrimSpace(stringValue(payload["proposal_id"]))),
		ThreadKey:      strings.TrimSpace(stringValue(effect.Payload["thread_key"])),
		Intent:         strings.TrimSpace(stringValue(effect.Payload["intent"])),
		RepoScope:      strings.TrimSpace(stringValue(effect.Payload["repo_scope"])),
		RequestedBy:    strings.TrimSpace(stringValue(effect.Payload["requested_by"])),
		ApprovalMode:   strings.TrimSpace(stringValue(effect.Payload["approval_mode"])),
		ResponseMode:   strings.TrimSpace(stringValue(effect.Payload["response_mode"])),
		Payload:        payload,
	}
	if item.Queue == "" {
		item.Queue = queue.ProposalQueue
	}
	if item.ProposalID == "" {
		return queue.WorkItem{}, false, fmt.Errorf("%s effect %s missing proposal_id", effect.EffectKind, effect.ID)
	}
	if item.TraceID == "" {
		return queue.WorkItem{}, false, fmt.Errorf("%s effect %s missing trace_id", effect.EffectKind, effect.ID)
	}
	if strings.TrimSpace(stringValue(item.Payload["attempt_id"])) == "" {
		return queue.WorkItem{}, false, fmt.Errorf("%s effect %s missing attempt_id", effect.EffectKind, effect.ID)
	}
	return item, false, nil
}

func proposalPhaseKindForEffect(effect transition.EffectExecution) (string, error) {
	if kind := strings.TrimSpace(stringValue(effect.Payload["work_item_kind"])); kind != "" {
		switch kind {
		case proposalOperationWorkspaceOpen, proposalOperationImplementAttempt, proposalOperationWorkspaceValidate:
			return kind, nil
		}
	}
	switch effect.EffectKind {
	case transition.EffectOpenWorkspace:
		return proposalOperationWorkspaceOpen, nil
	case transition.EffectInvokeRunner:
		return proposalOperationImplementAttempt, nil
	case transition.EffectWorkspaceValidate:
		return proposalOperationWorkspaceValidate, nil
	default:
		return "", fmt.Errorf("unsupported proposal phase effect %s", effect.EffectKind)
	}
}

func claimNextImprovementEffect(store storepkg.Store, holder string, lease time.Duration, sandboxObserveLease time.Duration) (transition.EffectExecution, bool, error) {
	for _, effect := range store.ListEffectExecutions() {
		switch effect.MachineKind {
		case transition.MachineProblemLine:
			if effect.EffectKind != transition.EffectInvokeRunner {
				continue
			}
		case transition.MachineAttempt:
			switch effect.EffectKind {
			case transition.EffectOpenWorkspace, transition.EffectInvokeRunner, transition.EffectWorkspaceValidate, transition.EffectObserveWorkspaceValidation, transition.EffectOpenDraftPR:
			default:
				continue
			}
		default:
			continue
		}
		effectLease := lease
		if effect.EffectKind == transition.EffectObserveWorkspaceValidation && sandboxObserveLease > 0 {
			effectLease = sandboxObserveLease
		}
		claimed, ok, err := store.ClaimEffectExecution(effect.ID, holder, effectLease)
		if err != nil {
			return transition.EffectExecution{}, false, err
		}
		if ok {
			return claimed, true, nil
		}
	}
	return transition.EffectExecution{}, false, nil
}

func completeClaimedImprovementEffect(store storepkg.Store, effect transition.EffectExecution, resultRef string) error {
	if strings.TrimSpace(effect.ID) == "" || effect.Status == transition.EffectCompleted {
		return nil
	}
	_, err := store.CompleteEffectExecution(effect.ID, resultRef)
	return err
}

func failClaimedImprovementEffect(store storepkg.Store, effect transition.EffectExecution, lastError string) error {
	if strings.TrimSpace(effect.ID) == "" || effect.Status == transition.EffectFailed {
		return nil
	}
	_, err := store.FailEffectExecution(effect.ID, lastError)
	return err
}

func attemptTraceForChange(store storepkg.Store, fallback events.Trace, attempt improvement.ChangeAttempt) events.Trace {
	if attempt.AttemptTraceID != "" {
		if trace, ok := store.GetTrace(attempt.AttemptTraceID); ok {
			return trace
		}
	}
	return fallback
}

func findRepoChangeJob(items []improvement.RepoChangeJob, jobID string) (improvement.RepoChangeJob, bool) {
	for _, item := range items {
		if item.ID == jobID {
			return item, true
		}
	}
	return improvement.RepoChangeJob{}, false
}

func workItemByID(items []queue.WorkItem, workItemID string) (queue.WorkItem, bool) {
	for _, item := range items {
		if item.ID == workItemID {
			return item, true
		}
	}
	return queue.WorkItem{}, false
}

func workItemByOperationID(items []queue.WorkItem, operationID string) (queue.WorkItem, bool) {
	for _, item := range items {
		if item.OperationID == operationID {
			return item, true
		}
	}
	return queue.WorkItem{}, false
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

func sandboxLogArtifactID(items []events.Artifact) string {
	for _, item := range items {
		if item.Kind == "sandbox_job_logs" {
			return item.ID
		}
	}
	return ""
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
	targetRepo := evalTargetRepo(cfg, store, trace)
	repoAllowlist := scopedImprovementRepoAllowlist(targetRepo)
	toolAllowlist := improvementReadOnlyTools(effectiveHarness)
	evalContextRefs := make([]clients.RunnerContextRef, 0, len(judgments)+1)
	evalContextRefs = append(evalContextRefs, clients.RunnerContextRef{
		Kind:     "eval_run",
		Ref:      run.ID,
		Summary:  run.OverallVerdict,
		TraceID:  trace.Summary.TraceID,
		ToolName: run.SuiteName,
		Decision: run.Trigger,
		Status:   string(item.Kind),
	})
	for _, judgment := range judgments {
		evalContextRefs = append(evalContextRefs, clients.RunnerContextRef{
			Kind:       "eval_judgment",
			Ref:        judgment.ID,
			Plane:      string(judgment.Layer),
			Service:    judgment.Category,
			Confidence: judgment.Score,
			Summary:    judgment.Rationale,
		})
	}
	evalContextRefs = append(evalContextRefs, improvementTraceEvidenceRefs(trace)...)
	evalContextRefs = append(evalContextRefs, improvementCandidateEvidenceRefs(store, trace, "")...)
	evalContextRefs = append(evalContextRefs, improvementProposalMemoryRefs(store, "")...)
	if targetRepo != "" {
		evalContextRefs = append(evalContextRefs, clients.RunnerContextRef{
			Kind:    "target_repo",
			Ref:     targetRepo,
			Summary: fmt.Sprintf("Authoritative target repository for this eval line is %s.", targetRepo),
		})
	}
	prompt := fmt.Sprintf(
		"Summarize the completed eval for trace %s. Workflow=%s status=%s thread=%s. Eval run=%s suite=%s verdict=%s score=%.2f. Judgments=%v. The authoritative target repository for this eval line is %s. Start from the supplied evidence pack, then use the read-only RSI tools when you need more trace, candidate, proposal-memory, or repo detail. If recalled memory conflicts with the target repository or trace evidence, prefer the target repository and explicit evidence. Explain what the eval found, what evidence mattered, whether improvement pressure should increase, and why.",
		trace.Summary.TraceID,
		trace.Summary.WorkflowKind,
		trace.Summary.Status,
		trace.Summary.ThreadKey,
		run.ID,
		run.SuiteName,
		run.OverallVerdict,
		run.OverallScore,
		judgmentDigest(judgments),
		firstNonEmpty(targetRepo, cfg.DefaultRepo),
	)
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
		}
	}
	sessionScopeKind, sessionScopeID := evalSessionScope(store, trace, run)
	return clients.RunnerTask{
		TaskType:            "eval",
		Repo:                firstNonEmpty(targetRepo, cfg.DefaultRepo),
		RepoRef:             "main",
		Prompt:              prompt,
		SystemMessage:       harness.ComposeSystemMessage("Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, and self_critique.", effectiveHarness),
		AllowedTools:        toolAllowlist,
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
		RepoAllowlist:             repoAllowlist,
		ToolAllowlist:             toolAllowlist,
		ResponseMode:              "analysis",
		ContextRefs:               evalContextRefs,
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

func buildProposalRunnerTask(cfg config.Config, store storepkg.Store, trace events.Trace, proposal review.Proposal, attempt improvement.ChangeAttempt, workspace *improvement.AttemptWorkspace, memories []review.ProposalMemory) clients.RunnerTask {
	effectiveHarness := harness.ResolveEffectiveConfig(store, "proposal", cfg.DefaultReasoningVerbosity)
	targetRepo := proposalTargetRepo(cfg, proposal)
	repoAllowlist := scopedImprovementRepoAllowlist(targetRepo)
	sessionScopeID := proposalSessionScopeID(proposal, targetRepo)
	executionMode := "investigate"
	toolAllowlist := improvementReadOnlyTools(effectiveHarness)
	if workspace != nil {
		executionMode = "implement"
		toolAllowlist = improvementImplementTools(effectiveHarness)
	}
	rejectedContext := make([]clients.RunnerRejectedProposalContext, 0, len(memories))
	for _, memory := range memories {
		rejectedContext = append(rejectedContext, clients.RunnerRejectedProposalContext{
			ProposalID:   memory.ProposalID,
			Disposition:  string(memory.Disposition),
			Rationale:    memory.ReviewRationale,
			FailureClass: firstNonEmpty(memory.FailureClass, strings.Join(memory.FailureClasses, ",")),
		})
	}
	contextRefs := []clients.RunnerContextRef{
		{
			Kind:                             "proposal",
			Ref:                              proposal.ID,
			Summary:                          proposal.Summary,
			CandidateKey:                     proposal.CandidateKey,
			TargetLayer:                      string(proposal.TargetLayer),
			TargetKind:                       proposal.TargetKind,
			TargetRef:                        proposal.TargetRef,
			RecommendedInterventionKind:      string(proposal.RecommendedInterventionKind),
			RecommendedInterventionRationale: proposal.RecommendedInterventionRationale,
			TargetSurface:                    proposal.TargetSurface,
			ValidationPlan:                   proposal.ValidationPlan,
			MaterialRiskSummary:              proposal.MaterialRiskSummary,
			RecommendedDisposition:           proposal.RecommendedDisposition,
		},
		{
			Kind:           "change_attempt",
			Ref:            attempt.ID,
			AttemptNumber:  attempt.AttemptNumber,
			AttemptID:      attempt.ID,
			BranchName:     attempt.BranchName,
			FailureClass:   attempt.FailureClass,
			FailureSummary: attempt.FailureSummary,
		},
	}
	if targetRepo != "" {
		contextRefs = append(contextRefs, clients.RunnerContextRef{
			Kind:    "target_repo",
			Ref:     targetRepo,
			Summary: fmt.Sprintf("Authoritative remediation repository is %s.", targetRepo),
		})
	}
	if workspace != nil {
		contextRefs = append(contextRefs, clients.RunnerContextRef{
			Kind:             "attempt_workspace",
			Ref:              workspace.ID,
			ProposalID:       workspace.ProposalID,
			AttemptID:        workspace.AttemptID,
			Repo:             workspace.Repo,
			BranchName:       workspace.BranchName,
			Status:           string(workspace.Status),
			AllowedPathGlobs: workspace.AllowedPathGlobs,
		})
	}
	contextRefs = append(contextRefs, improvementTraceEvidenceRefs(trace)...)
	contextRefs = append(contextRefs, improvementCandidateEvidenceRefs(store, trace, proposal.CandidateKey)...)
	contextRefs = append(contextRefs, improvementProposalMemoryRefs(store, proposal.CandidateKey)...)
	contextRefs = append(contextRefs, improvementAttemptHistoryRefs(store, proposal.ID, attempt.ID)...)
	prompt := fmt.Sprintf(
		"Execute approved intervention attempt %d for proposal %s. Candidate=%s risk=%s scope=%s summary=%s. The approved intervention kind is %s with rationale %q, target surface %q, and validation plan %q. The authoritative target repository is %s. Start from the dense evidence pack: latest failing trace evidence, root-cause metadata, prior rejected or dismissed proposal memory, and prior attempt failures. If recalled memory mentions a different repository, treat it as stale unless the supplied evidence pack explicitly supports it. Investigate the approved scope, ground the implementation in concrete files within %s, and return explicit visible reasoning, change_plan, validation_plan, retry_assessment, hypothesis_delta, and any governed proposed_actions needed for the next platform step.",
		attempt.AttemptNumber,
		proposal.ID,
		proposal.CandidateKey,
		proposal.RiskTier,
		proposal.ProposedScope,
		proposal.Summary,
		firstNonEmpty(string(proposal.RecommendedInterventionKind), string(review.InterventionRepoChange)),
		firstNonEmpty(proposal.RecommendedInterventionRationale),
		firstNonEmpty(proposal.TargetSurface),
		firstNonEmpty(proposal.ValidationPlan),
		firstNonEmpty(targetRepo, cfg.DefaultRepo),
		firstNonEmpty(targetRepo, cfg.DefaultRepo),
	)
	if workspace != nil {
		prompt = fmt.Sprintf(
			"Implement approved repo-change attempt %d for proposal %s inside governed workspace %s. Candidate=%s risk=%s scope=%s summary=%s. The approved intervention rationale is %q, target surface is %q, and validation plan is %q. The authoritative repository is %s on branch %s. Use workspace tools to inspect, edit, diff, and validate only inside the allowed path globs. You must make at least one justified in-scope workspace mutation when recommending a repo change; if no safe in-scope mutation is warranted, return retry_assessment with a concrete failure_class and do not request PR open. The authoritative patch is the workspace git diff, not repo_patch text. Validation must be grounded in the same workspace. After validation succeeds, decide whether opening a draft PR is warranted; if yes, include exactly one proposed action kind=%q with title, body, branch_name, and base_ref in request_payload. Do not mutate GitHub directly.",
			attempt.AttemptNumber,
			proposal.ID,
			workspace.ID,
			proposal.CandidateKey,
			proposal.RiskTier,
			proposal.ProposedScope,
			proposal.Summary,
			firstNonEmpty(proposal.RecommendedInterventionRationale),
			firstNonEmpty(proposal.TargetSurface),
			firstNonEmpty(proposal.ValidationPlan),
			firstNonEmpty(targetRepo, cfg.DefaultRepo),
			workspace.BranchName,
			action.KindDraftPROpen,
		)
	}
	if proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		targetRole := firstNonEmpty(strings.TrimSpace(proposal.TargetRef), "prod")
		prompt = fmt.Sprintf(
			"Design remediation attempt %d as an approved runtime harness overlay for role %s from proposal %s. Candidate=%s summary=%s. The approved intervention rationale is %q and the approved target surface is %q. Return explicit visible reasoning, a change_plan, validation_plan, retry_assessment, hypothesis_delta, and exactly one proposed action with kind=%q whose request_payload contains prompt_fragments, few_shot_snippets, tool_preference_order, retrieval_bias, reasoning_verbosity, memory_read_enabled, and memory_write_enabled. This is a runtime overlay activation, not a repo change.",
			attempt.AttemptNumber,
			targetRole,
			proposal.ID,
			proposal.CandidateKey,
			proposal.Summary,
			firstNonEmpty(proposal.RecommendedInterventionRationale),
			firstNonEmpty(proposal.TargetSurface),
			action.KindHarnessOverlay,
		)
	}
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
		}
	}
	return clients.RunnerTask{
		TaskType:                  "proposal",
		Repo:                      firstNonEmpty(targetRepo, cfg.DefaultRepo),
		RepoRef:                   firstNonEmpty(valueOrWorkspaceBaseRef(workspace), "main"),
		Prompt:                    prompt,
		SystemMessage:             harness.ComposeSystemMessage("Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta, proposed_actions, knowledge_drafts, and outcome_hypotheses. For repo-change implement mode, repo_patch is optional and the authoritative diff is the workspace git diff.", effectiveHarness),
		AllowedTools:              toolAllowlist,
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
		RepoAllowlist:             repoAllowlist,
		ToolAllowlist:             toolAllowlist,
		ResponseMode:              "analysis",
		ContextRefs:               contextRefs,
		ApprovalMode:              "human_review",
		ReasoningVerbosity:        effectiveHarness.ReasoningVerbosity,
		ExecutionMode:             executionMode,
		SessionScopeKind:          "proposal_candidate",
		SessionScopeID:            sessionScopeID,
		HarnessProfileID:          effectiveHarness.Profile.ID,
		HarnessOverlayVersion:     effectiveHarness.EffectiveOverlayVersion,
		MemoryBackend:             harness.DefaultMemoryBackend,
		AssistantPeerID:           fmt.Sprintf("rsi:%s:proposal", cfg.Environment),
		UserPeerID:                fmt.Sprintf("candidate:%s", sessionScopeID),
		AttemptID:                 attempt.ID,
		WorkspaceID:               workspaceIDValue(workspace),
		WorkspaceRepo:             workspaceRepoValue(workspace),
		WorkspaceBranch:           workspaceBranchValue(workspace),
		AllowedPathGlobs:          workspaceAllowedPathGlobsValue(workspace),
	}
}

func evalTargetRepo(cfg config.Config, store storepkg.Store, trace events.Trace) string {
	if candidate, ok := latestCandidateForTrace(store, trace.Summary.TraceID); ok {
		return improvementTargetRepo(cfg, candidate.TargetLayer, candidate.TargetKind, candidate.TargetRef)
	}
	return firstNonEmpty(cfg.DefaultRepo, "rsi-agent-platform")
}

func proposalTargetRepo(cfg config.Config, proposal review.Proposal) string {
	return improvementTargetRepo(cfg, proposal.TargetLayer, proposal.TargetKind, proposal.TargetRef)
}

func proposalSessionScopeID(proposal review.Proposal, targetRepo string) string {
	targetRepo = strings.TrimSpace(targetRepo)
	if targetRepo == "" {
		return proposal.CandidateKey
	}
	return fmt.Sprintf("%s|repo:%s|v2", proposal.CandidateKey, targetRepo)
}

func improvementTargetRepo(cfg config.Config, targetLayer harness.TargetLayer, targetKind string, targetRef string) string {
	if targetLayer == harness.TargetLayerHarnessOverlay {
		return firstNonEmpty(cfg.DefaultRepo, "rsi-agent-platform")
	}
	candidate := strings.TrimSpace(cfg.GitHubRepoName(targetRef))
	if candidate != "" {
		return candidate
	}
	if strings.EqualFold(strings.TrimSpace(targetKind), "repo") && strings.TrimSpace(targetRef) != "" {
		return strings.TrimSpace(targetRef)
	}
	return firstNonEmpty(cfg.DefaultRepo, "rsi-agent-platform")
}

func scopedImprovementRepoAllowlist(primary string) []string {
	primary = strings.TrimSpace(primary)
	if primary == "" {
		return nil
	}
	return []string{primary}
}

func improvementBaseToolNames() []string {
	return []string{
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
		"rsi.workflow_context",
		"rsi.action_chain",
		"rsi.runner_execution",
		"rsi.runtime_config",
		"rsi.runtime_health",
		"rsi.proposal_memory",
		"rsi.candidate_context",
		"rsi.attempt_context",
	}
}

func improvementWorkspaceToolNames() []string {
	return []string{
		"workspace.list_files",
		"workspace.read_file",
		"workspace.search",
		"workspace.write_file",
		"workspace.apply_patch",
		"workspace.git_status",
		"workspace.git_diff",
		"workspace.run_validation",
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
	return harness.ApplyToolPreference(improvementBaseToolNames(), effective.ToolPreferenceOrder)
}

func improvementImplementTools(effective harness.EffectiveConfig) []string {
	tools := append(improvementBaseToolNames(), improvementWorkspaceToolNames()...)
	return harness.ApplyToolPreference(tools, effective.ToolPreferenceOrder)
}

func improvementTraceEvidenceRefs(trace events.Trace) []clients.RunnerContextRef {
	eventRefs := make([]clients.RunnerContextRef, 0, minInt(len(trace.Events), 12))
	for _, item := range tailTraceEvents(trace.Events, 12) {
		eventRefs = append(eventRefs, clients.RunnerContextRef{
			Kind:        "trace_event",
			Ref:         item.EventType,
			Status:      string(item.Status),
			Plane:       item.Plane,
			Service:     item.Service,
			Description: item.Description,
		})
	}
	reasoningRefs := make([]clients.RunnerContextRef, 0, minInt(len(trace.Reasoning), 8))
	for _, item := range tailReasoning(trace.Reasoning, 8) {
		reasoningRefs = append(reasoningRefs, clients.RunnerContextRef{
			Kind:       "reasoning_step",
			Ref:        item.ID,
			StepType:   item.StepType,
			Summary:    item.Summary,
			Decision:   item.Decision,
			Confidence: item.Confidence,
		})
	}
	return append(eventRefs, reasoningRefs...)
}

func improvementCandidateEvidenceRefs(store storepkg.Store, trace events.Trace, candidateKey string) []clients.RunnerContextRef {
	refs := make([]clients.RunnerContextRef, 0)
	for _, item := range store.ListCandidates() {
		if candidateKey != "" && item.CandidateKey != candidateKey {
			continue
		}
		if candidateKey == "" && item.LatestTraceID != trace.Summary.TraceID && !containsString(item.EvidenceTraceIDs, trace.Summary.TraceID) {
			continue
		}
		refs = append(refs, clients.RunnerContextRef{
			Kind:                     "candidate",
			Ref:                      item.CandidateKey,
			Subsystem:                item.Subsystem,
			FailureMode:              item.FailureMode,
			TargetLayer:              string(item.TargetLayer),
			PriorityScore:            item.PriorityScore,
			RetryableFailureClass:    item.RetryableFailureClass,
			AttemptCount:             item.AttemptCount,
			AutoRetryBudgetRemaining: item.AutoRetryBudgetRemaining,
		})
	}
	return refs
}

func improvementProposalMemoryRefs(store storepkg.Store, candidateKey string) []clients.RunnerContextRef {
	refs := make([]clients.RunnerContextRef, 0)
	for _, item := range store.ListProposalMemories() {
		if candidateKey != "" && item.CandidateKey != candidateKey {
			continue
		}
		refs = append(refs, clients.RunnerContextRef{
			Kind:         "proposal_memory",
			Ref:          item.ID,
			ProposalID:   item.ProposalID,
			Disposition:  string(item.Disposition),
			FailureClass: firstNonEmpty(item.FailureClass, strings.Join(item.FailureClasses, ",")),
			Rationale:    item.ReviewRationale,
			Hypothesis:   item.Hypothesis,
			DiffSummary:  item.DiffSummary,
		})
		if len(refs) == 8 {
			break
		}
	}
	return refs
}

func improvementAttemptHistoryRefs(store storepkg.Store, proposalID string, currentAttemptID string) []clients.RunnerContextRef {
	refs := make([]clients.RunnerContextRef, 0)
	for _, item := range store.ListChangeAttempts() {
		if item.ProposalID != proposalID || item.ID == currentAttemptID {
			continue
		}
		refs = append(refs, clients.RunnerContextRef{
			Kind:                     "change_attempt_history",
			Ref:                      item.ID,
			AttemptNumber:            item.AttemptNumber,
			State:                    string(item.State),
			FailureClass:             item.FailureClass,
			FailureSummary:           item.FailureSummary,
			RetryDecision:            item.RetryDecision,
			MaterialHypothesisChange: item.MaterialHypothesisChange,
			ChangedFiles:             append([]string(nil), item.ChangedFiles...),
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

func valueOrWorkspaceBaseRef(workspace *improvement.AttemptWorkspace) string {
	if workspace == nil {
		return ""
	}
	return workspace.BaseRef
}

func workspaceIDValue(workspace *improvement.AttemptWorkspace) string {
	if workspace == nil {
		return ""
	}
	return workspace.ID
}

func workspaceRepoValue(workspace *improvement.AttemptWorkspace) string {
	if workspace == nil {
		return ""
	}
	return workspace.Repo
}

func workspaceBranchValue(workspace *improvement.AttemptWorkspace) string {
	if workspace == nil {
		return ""
	}
	return workspace.BranchName
}

func workspaceAllowedPathGlobsValue(workspace *improvement.AttemptWorkspace) []string {
	if workspace == nil {
		return nil
	}
	return append([]string(nil), workspace.AllowedPathGlobs...)
}

func improvementRecentConversationEntries(items []conversation.Entry) []clients.RunnerConversationEntry {
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

func improvementPriorTraceRefs(items []events.TraceSummary, caseID string, currentTraceID string) []clients.RunnerTraceRef {
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
	for idx, item := range drafts {
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
		if _, err := runnerutil.PersistKnowledgeDraft(store, knowledge.Entry{
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
		}, links, "improvement-plane", firstNonEmpty(proposalID, trace.Summary.TraceID), idx, createdAt); err != nil {
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

func improvementActionIntentBase(requestedBy string, proposal review.Proposal, trace events.Trace, attemptID string, kind action.Kind, targetRef string, status action.Status, rationale string, idempotencyKey string, requestPayload map[string]any, evidenceRefs []events.EvidenceRef, createdAt time.Time) action.Intent {
	return action.Intent{
		OwnerPlane:     "improvement",
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		TraceID:        trace.Summary.TraceID,
		ProposalID:     proposal.ID,
		AttemptID:      attemptID,
		Kind:           kind,
		TargetRef:      targetRef,
		ApprovalMode:   "approved",
		ApprovalState:  "approved",
		PolicyVerdict:  "approved_proposal",
		Status:         status,
		RequestedBy:    requestedBy,
		Rationale:      rationale,
		IdempotencyKey: idempotencyKey,
		RequestPayload: requestPayload,
		EvidenceRefs:   evidenceRefs,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}
}

func ensureImprovementActionIntent(store storepkg.Store, template action.Intent) (action.Intent, error) {
	if existing, ok := findMatchingActionIntent(store.ListActionIntents(), template); ok {
		return existing, nil
	}
	if strings.TrimSpace(template.IdempotencyKey) == "" {
		return action.Intent{}, errors.New("improvement action intent idempotency key is required")
	}
	template.ID = improvementActionIntentIDFromIdempotencyKey(template.IdempotencyKey)
	occurredAt := template.CreatedAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: template.ID,
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   improvementActionCommandID(template.ID, transition.CommandActionQueue, ""),
		Actor:       firstNonEmpty(template.RequestedBy, "improvement-plane"),
		OccurredAt:  occurredAt,
		Payload: map[string]any{
			"owner_plane":     template.OwnerPlane,
			"conversation_id": template.ConversationID,
			"case_id":         template.CaseID,
			"trace_id":        template.TraceID,
			"proposal_id":     template.ProposalID,
			"attempt_id":      template.AttemptID,
			"kind":            string(template.Kind),
			"phase_key":       template.PhaseKey,
			"target_ref":      template.TargetRef,
			"request_payload": clonePayload(template.RequestPayload),
			"idempotency_key": template.IdempotencyKey,
			"approval_mode":   template.ApprovalMode,
			"approval_state":  template.ApprovalState,
			"policy_verdict":  template.PolicyVerdict,
			"requested_by":    template.RequestedBy,
			"rationale":       template.Rationale,
			"evidence_refs":   normalizeImprovementEvidenceRefs(template.EvidenceRefs, template.TraceID, template.ProposalID),
		},
	})
	if err != nil {
		return action.Intent{}, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return action.Intent{}, errors.New(receipt.Reason)
	}
	intent, ok := store.GetActionIntent(template.ID)
	if !ok {
		return action.Intent{}, fmt.Errorf("action intent %s not found after queue command", template.ID)
	}
	return intent, nil
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

func findAttemptActionIntent(store storepkg.Store, proposalID string, attemptID string, kind action.Kind) (action.Intent, bool) {
	for _, item := range store.ListActionIntents() {
		if item.ProposalID == proposalID && item.AttemptID == attemptID && item.Kind == kind {
			return item, true
		}
	}
	return action.Intent{}, false
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
	if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		return fmt.Sprintf("Approved intervention %s for proposal %s is activating a runtime harness overlay on %s.", firstNonEmpty(string(proposal.RecommendedInterventionKind), string(review.InterventionHarnessOverlay)), proposal.ID, firstNonEmpty(proposal.TargetSurface, proposal.TargetRef, "prod"))
	}
	return fmt.Sprintf("Approved intervention %s for proposal %s is entering repo-change execution on %s.", firstNonEmpty(string(proposal.RecommendedInterventionKind), string(review.InterventionRepoChange)), proposal.ID, firstNonEmpty(proposal.TargetSurface, proposal.TargetRef, "rsi-agent-platform"))
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

func improvementActionIntentIDFromIdempotencyKey(key string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(key)))
	return fmt.Sprintf("action-%x", sum[:8])
}

func submitImprovementActionCommand(store storepkg.Store, actionID string, kind transition.ActionExecutionCommandKind, actor string, occurredAt time.Time, payload map[string]any) (transition.CommandReceipt, error) {
	if strings.TrimSpace(actionID) == "" {
		return transition.CommandReceipt{}, nil
	}
	return store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: actionID,
		CommandKind: string(kind),
		CommandID:   improvementActionCommandID(actionID, kind, stringValue(payload["operation_id"])),
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
}

func improvementActionCommandID(actionID string, kind transition.ActionExecutionCommandKind, operationID string) string {
	base := fmt.Sprintf("cmd-action:%s:%s", strings.TrimSpace(actionID), string(kind))
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return base
	}
	return base + ":" + operationID
}

func improvementActionCommandForStatus(status action.Status) (transition.ActionExecutionCommandKind, error) {
	switch status {
	case action.StatusSucceeded:
		return transition.CommandActionSucceed, nil
	case action.StatusBlocked:
		return transition.CommandActionBlock, nil
	case action.StatusFailed:
		return transition.CommandActionFail, nil
	default:
		return "", fmt.Errorf("unsupported terminal improvement action status %s", status)
	}
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

func effectPayloadMap(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return clonePayload(typed)
	default:
		return map[string]interface{}{}
	}
}

func effectQueueName(value string) queue.QueueName {
	switch strings.TrimSpace(value) {
	case string(queue.WorkflowQueue):
		return queue.WorkflowQueue
	case string(queue.ProactiveQueue):
		return queue.ProactiveQueue
	case string(queue.ProposalQueue):
		return queue.ProposalQueue
	case string(queue.SandboxQueue):
		return queue.SandboxQueue
	case string(queue.ImprovementActionQueue):
		return queue.ImprovementActionQueue
	case string(queue.KnowledgeMaintenanceQueue):
		return queue.KnowledgeMaintenanceQueue
	default:
		return queue.ImprovementActionQueue
	}
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
	case nil:
		return ""
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
