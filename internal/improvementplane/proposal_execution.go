package improvementplane

import (
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	proposalOperationLineActivate      = "line_activate"
	proposalOperationAttemptPlan       = "attempt_plan"
	proposalOperationWorkspaceOpen     = "workspace_open"
	proposalOperationImplementAttempt  = "implement_attempt"
	proposalOperationWorkspaceValidate = "workspace_validate"
)

type runnerExecutor interface {
	Execute(clients.RunnerTask) (clients.RunnerResponse, error)
}

type toolExecutor interface {
	Execute(string, map[string]any) (storepkg.ToolResult, error)
}

func resolveProposalOperationKind(store storepkg.Store, item queue.WorkItem) string {
	if item.OperationID != "" {
		if op, ok := store.GetOperation(item.OperationID); ok {
			switch strings.TrimSpace(op.OperationKind) {
			case proposalOperationLineActivate,
				proposalOperationAttemptPlan,
				proposalOperationWorkspaceOpen,
				proposalOperationImplementAttempt,
				proposalOperationWorkspaceValidate:
				return strings.TrimSpace(op.OperationKind)
			}
		}
	}
	switch strings.TrimSpace(item.Kind) {
	case proposalOperationAttemptPlan,
		proposalOperationWorkspaceOpen,
		proposalOperationImplementAttempt,
		proposalOperationWorkspaceValidate:
		return strings.TrimSpace(item.Kind)
	default:
		return proposalOperationLineActivate
	}
}

func queueProposalAttemptPhase(cfg config.Config, store storepkg.Store, proposal review.Proposal, trace events.Trace, attempt improvement.ChangeAttempt, operationKind string, payload map[string]any) (queue.WorkItem, error) {
	op, item := proposalAttemptPhaseWork(cfg, proposal, trace, attempt, operationKind, payload)
	return enqueueImprovementOperationWork(store, op, item)
}

func proposalAttemptPhaseWork(cfg config.Config, proposal review.Proposal, trace events.Trace, attempt improvement.ChangeAttempt, operationKind string, payload map[string]any) (operation.Execution, queue.WorkItem) {
	now := time.Now().UTC()
	if payload == nil {
		payload = map[string]any{}
	}
	payload["attempt_id"] = attempt.ID
	return operation.Execution{
			ScopeKind:     operation.ScopeAttempt,
			ScopeID:       attempt.ID,
			OperationKind: operationKind,
			OperationKey:  operationKind,
			Status:        operation.StatusQueued,
			Queue:         queue.ProposalQueue,
			RequestedBy:   cfg.ServiceName,
			TraceID:       trace.Summary.TraceID,
			ProposalID:    proposal.ID,
			AttemptID:     attempt.ID,
		}, queue.WorkItem{
			Queue:          queue.ProposalQueue,
			Kind:           operationKind,
			Status:         queue.WorkQueued,
			TraceID:        trace.Summary.TraceID,
			ConversationID: proposal.ConversationID,
			CaseID:         proposal.CaseID,
			TriggerEventID: proposal.OriginTraceID,
			ProposalID:     proposal.ID,
			RequestedBy:    cfg.ServiceName,
			ApprovalMode:   "human_review",
			CreatedAt:      now,
			UpdatedAt:      now,
			Payload:        payload,
		}
}

func processProposalLineActivate(cfg config.Config, store storepkg.Store, proposal review.Proposal, trace events.Trace, item queue.WorkItem) error {
	attempt, attemptTrace, err := ensureProposalAttempt(cfg, store, proposal, trace, item)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	nextOp, nextItem := proposalAttemptPhaseWork(cfg, proposal, attemptTrace, attempt, proposalOperationAttemptPlan, clonePayload(item.Payload))
	if err := store.AdvanceProposalAttemptPhase(storepkg.ProposalAttemptPhaseAdvance{
		ProposalID:  proposal.ID,
		WorkItemID:  item.ID,
		OperationID: item.OperationID,
		Attempt:     &attempt,
		TraceID:     attemptTrace.Summary.TraceID,
		TraceUpdate: &storepkg.TraceUpdate{
			Events: []events.TraceEvent{
				{
					TraceID:     attemptTrace.Summary.TraceID,
					IngestionID: attemptTrace.Summary.IngestionID,
					WorkflowID:  attemptTrace.Summary.WorkflowID,
					Plane:       "improvement",
					Service:     cfg.ServiceName,
					Actor:       "attempt-supervisor",
					EventType:   "change_attempt.phase_queued",
					Status:      events.StatusQueued,
					StartedAt:   now,
					Description: fmt.Sprintf("Activated proposal line %s and queued %s for attempt %s.", proposal.ID, proposalOperationAttemptPlan, attempt.ID),
				},
			},
		},
		NextOperation: &nextOp,
		NextWorkItem:  &nextItem,
	}); err != nil {
		return err
	}
	return errProposalPhaseHandled
}

func processProposalAttemptPlan(cfg config.Config, store storepkg.Store, proposal review.Proposal, trace events.Trace, item queue.WorkItem) error {
	attempt, attemptTrace, err := ensureProposalAttempt(cfg, store, proposal, trace, item)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		attempt.State = improvement.AttemptStateOverlayPlan
	} else {
		attempt.State = improvement.AttemptStatePatchPlan
	}
	attempt.UpdatedAt = now
	nextPhase := proposalOperationWorkspaceOpen
	if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		nextPhase = proposalOperationImplementAttempt
	}
	nextOp, nextItem := proposalAttemptPhaseWork(cfg, proposal, attemptTrace, attempt, nextPhase, clonePayload(item.Payload))
	if err := store.AdvanceProposalAttemptPhase(storepkg.ProposalAttemptPhaseAdvance{
		ProposalID:  proposal.ID,
		WorkItemID:  item.ID,
		OperationID: item.OperationID,
		Attempt:     &attempt,
		TraceID:     attemptTrace.Summary.TraceID,
		TraceUpdate: &storepkg.TraceUpdate{
			Events: []events.TraceEvent{
				{
					TraceID:     attemptTrace.Summary.TraceID,
					IngestionID: attemptTrace.Summary.IngestionID,
					WorkflowID:  attemptTrace.Summary.WorkflowID,
					Plane:       "improvement",
					Service:     cfg.ServiceName,
					Actor:       "attempt-supervisor",
					EventType:   "change_attempt.phase_queued",
					Status:      events.StatusQueued,
					StartedAt:   now,
					Description: fmt.Sprintf("Attempt %s entered planning and queued %s.", attempt.ID, nextPhase),
				},
			},
		},
		NextOperation: &nextOp,
		NextWorkItem:  &nextItem,
	}); err != nil {
		return err
	}
	return errProposalPhaseHandled
}

func processProposalWorkspaceOpen(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, proposal review.Proposal, trace events.Trace, item queue.WorkItem) error {
	attempt, attemptTrace, err := ensureProposalAttempt(cfg, store, proposal, trace, item)
	if err != nil {
		return err
	}
	workspace, ready, err := ensureAttemptWorkspace(cfg, store, launcher, launcherErr, proposal, attempt, attemptTrace.Summary.TraceID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	job, err := ensureWorkspaceRepoChangeJob(store, proposal, attempt, workspace)
	if err != nil {
		return err
	}
	if !ready {
		proposal.Status = review.ProposalRepoChangeQueued
		workspace.UpdatedAt = now
		job.Status = string(review.ProposalRepoChangeQueued)
		job.SandboxNamespace = workspace.Namespace
		job.SandboxJobName = workspace.JobName
		job.SandboxPodName = workspace.PodName
		job.ValidationRef = firstNonEmpty(job.ValidationRef, fmt.Sprintf("%s/%s", workspace.Namespace, firstNonEmpty(workspace.PodName, workspace.JobName)))
		job.UpdatedAt = now
		payload := clonePayload(item.Payload)
		if payload == nil {
			payload = map[string]any{}
		}
		payload["workspace_id"] = workspace.ID
		if err := store.DeferProposalAttemptPhase(storepkg.ProposalAttemptPhaseDefer{
			ProposalID:    proposal.ID,
			WorkItemID:    item.ID,
			OperationID:   item.OperationID,
			LastError:     "workspace initializing",
			AvailableAt:   time.Now().UTC().Add(5 * time.Second),
			Payload:       payload,
			Proposal:      &proposal,
			Workspace:     &workspace,
			RepoChangeJob: &job,
		}); err != nil {
			return err
		}
		return errDeferredWorkItem
	}
	if workspace.Status != improvement.WorkspaceReady {
		workspace.Status = improvement.WorkspaceReady
	}
	workspace.UpdatedAt = now
	job.Status = string(review.ProposalRepoChangeRunning)
	job.SandboxNamespace = workspace.Namespace
	job.SandboxJobName = workspace.JobName
	job.SandboxPodName = workspace.PodName
	job.ValidationRef = fmt.Sprintf("%s/%s", workspace.Namespace, firstNonEmpty(workspace.PodName, workspace.JobName))
	job.UpdatedAt = now
	proposal.Status = review.ProposalRepoChangeRunning
	nextOp, nextItem := proposalAttemptPhaseWork(cfg, proposal, attemptTrace, attempt, proposalOperationImplementAttempt, map[string]any{
		"workspace_id": workspace.ID,
	})
	if err := store.AdvanceProposalAttemptPhase(storepkg.ProposalAttemptPhaseAdvance{
		ProposalID:    proposal.ID,
		WorkItemID:    item.ID,
		OperationID:   item.OperationID,
		Proposal:      &proposal,
		Workspace:     &workspace,
		RepoChangeJob: &job,
		TraceID:       attemptTrace.Summary.TraceID,
		TraceUpdate: &storepkg.TraceUpdate{
			Events: []events.TraceEvent{
				{
					TraceID:     attemptTrace.Summary.TraceID,
					IngestionID: attemptTrace.Summary.IngestionID,
					WorkflowID:  attemptTrace.Summary.WorkflowID,
					Plane:       "improvement",
					Service:     cfg.ServiceName,
					Actor:       "attempt-supervisor",
					EventType:   "workspace.ready",
					Status:      events.StatusQueued,
					StartedAt:   now,
					Description: fmt.Sprintf("Workspace %s is ready; queued %s for attempt %s.", workspace.ID, proposalOperationImplementAttempt, attempt.ID),
				},
			},
		},
		NextOperation: &nextOp,
		NextWorkItem:  &nextItem,
	}); err != nil {
		return err
	}
	return errProposalPhaseHandled
}

func processProposalImplementAttempt(cfg config.Config, store storepkg.Store, runnerClient runnerExecutor, toolClient toolExecutor, proposal review.Proposal, trace events.Trace, item queue.WorkItem) error {
	attempt, attemptTrace, err := ensureProposalAttempt(cfg, store, proposal, trace, item)
	if err != nil {
		return err
	}
	var workspace *improvement.AttemptWorkspace
	if proposal.RecommendedInterventionKind != review.InterventionHarnessOverlay && proposal.TargetLayer != harness.TargetLayerHarnessOverlay {
		workspaceID := strings.TrimSpace(stringValue(item.Payload["workspace_id"]))
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
			return fmt.Errorf("attempt %s missing workspace for implement phase", attempt.ID)
		}
		workspace.Status = improvement.WorkspaceExecuting
		workspace.UpdatedAt = time.Now().UTC()
		if _, err := store.UpsertAttemptWorkspace(*workspace); err != nil {
			return err
		}
	}
	memories := filterProposalMemory(store.ListProposalMemories(), proposal.CandidateKey)
	runnerStarted := time.Now().UTC()
	_, _ = store.ApplyTraceUpdate(attemptTrace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
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
			},
		},
	})
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
				item.OperationID,
				attemptTrace.Summary.TraceID,
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
			var parseErr error
			runnerOutput, parseErr = runnerutil.ParseStructuredOutput(runnerResp)
			if parseErr != nil {
				return parseErr
			}
		}
	}
	if runnerErr != nil {
		retryAt := time.Now().UTC().Add(proposalRunnerBackoff(item.Attempts))
		payload := clonePayload(item.Payload)
		if payload == nil {
			payload = map[string]any{}
		}
		now := time.Now().UTC()
		if err := store.DeferProposalAttemptPhase(storepkg.ProposalAttemptPhaseDefer{
			ProposalID:  proposal.ID,
			WorkItemID:  item.ID,
			OperationID: item.OperationID,
			LastError:   runnerErr.Error(),
			AvailableAt: retryAt,
			Payload:     payload,
			TraceID:     attemptTrace.Summary.TraceID,
			TraceUpdate: &storepkg.TraceUpdate{
				Events: []events.TraceEvent{
					{
						TraceID:     attemptTrace.Summary.TraceID,
						IngestionID: attemptTrace.Summary.IngestionID,
						WorkflowID:  attemptTrace.Summary.WorkflowID,
						Plane:       "execution",
						Service:     "runner",
						Actor:       "proposal",
						EventType:   "runner.failed",
						Status:      events.StatusFailed,
						StartedAt:   runnerStarted,
						EndedAt:     ptrTime(now),
						Description: fmt.Sprintf("Proposal runner failed; attempt remains resumable and will retry after %s: %v", retryAt.Format(time.RFC3339), runnerErr),
					},
				},
				Reasoning: []events.ReasoningStep{
					{
						ID:         fmt.Sprintf("reason-proposal-retry-%d", now.UnixNano()),
						TraceID:    attemptTrace.Summary.TraceID,
						WorkflowID: attemptTrace.Summary.WorkflowID,
						StepType:   "proposal_runner_retry",
						Summary:    "Proposal runner failed closed. The attempt was deferred without materializing further side effects.",
						Confidence: 0.93,
						Decision:   retryAt.Format(time.RFC3339),
						CreatedAt:  now,
					},
				},
			},
		}); err != nil {
			return err
		}
		return errDeferredWorkItem
	}
	if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		return processHarnessOverlayProposal(cfg, store, attemptTrace, proposal, attempt, runnerResp, runnerOutput, runnerStarted)
	}
	if workspace == nil {
		return fmt.Errorf("proposal %s missing workspace for repo-change execution", proposal.ID)
	}
	refreshedTrace, ok := store.GetTrace(attemptTrace.Summary.TraceID)
	if ok {
		attemptTrace = refreshedTrace
	}
	if workspaceMutationCallCount(attemptTrace) == 0 {
		return recordProposalPhaseAttemptFailure(cfg, store, proposal, attempt, attemptTrace, item, "no_op_diff", "Proposal implement run completed without any write-capable workspace tool calls.", false, improvement.AttemptTriggerProposalApproved, workspace, nil)
	}
	diffResult, execErr := toolClient.Execute("workspace.git_diff", map[string]any{
		"trace_id":     attemptTrace.Summary.TraceID,
		"workspace_id": workspace.ID,
		"attempt_id":   attempt.ID,
	})
	if execErr != nil || diffResult.Status != "ok" {
		return recordProposalPhaseAttemptFailure(cfg, store, proposal, attempt, attemptTrace, item, "sandbox_failure", firstNonEmpty(improvementActionError(diffResult, execErr), "Workspace diff inspection failed."), false, improvement.AttemptTriggerSandboxFailed, workspace, nil)
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
		return recordProposalPhaseAttemptFailure(cfg, store, proposal, attempt, attemptTrace, item, failureClass, failureSummary, runnerOutput.RetryAssessment.MaterialHypothesisChange, improvement.AttemptTriggerProposalApproved, workspace, nil)
	}
	attempt.State = improvement.AttemptStatePatchGenerated
	workspace.DiffSummary = attempt.DiffSummary
	workspace.HeadSHA = attempt.HeadSHA
	workspace.UpdatedAt = attempt.UpdatedAt
	now := time.Now().UTC()
	reasoning := []events.ReasoningStep{
		{
			ID:         fmt.Sprintf("reason-proposal-%d", now.UnixNano()),
			TraceID:    attemptTrace.Summary.TraceID,
			WorkflowID: attemptTrace.Summary.WorkflowID,
			StepType:   "workspace_implemented",
			Summary:    firstNonEmpty(runnerOutput.ChangePlan, runnerOutput.FinalAnswer, runnerOutput.ContextSummary, fmt.Sprintf("Approved proposal %s produced a workspace-backed implementation.", proposal.ID)),
			Confidence: confidenceOr(0.86, runnerOutput.Confidence),
			Decision:   workspace.BranchName,
			CreatedAt:  now,
		},
	}
	reasoning = append(reasoning, runnerutil.ToTraceReasoning(attemptTrace.Summary.TraceID, attemptTrace.Summary.WorkflowID, runnerOutput, now)...)
	reasoning = append(reasoning, improvementOutcomeHypothesisReasoning(attemptTrace, runnerOutput.OutcomeHypotheses, now)...)
	if err := persistImprovementKnowledgeDrafts(store, runnerOutput.KnowledgeDrafts, attemptTrace, proposal.ID, now); err != nil {
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
	_, _ = store.ApplyTraceUpdate(attemptTrace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
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
			},
		},
		Reasoning: reasoning,
	})
	prAction, ok := proposedActionByKind(runnerOutput.ProposedActions, action.KindDraftPROpen)
	payload := map[string]any{
		"workspace_id":       workspace.ID,
		"validation_command": workspaceValidationCommand(firstNonEmpty(runnerOutput.ValidationPlan, proposal.ValidationPlan)),
	}
	if ok {
		payload["branch_name"] = firstNonEmpty(stringValue(prAction.RequestPayload["branch_name"]), workspace.BranchName, attempt.BranchName)
		payload["base_ref"] = firstNonEmpty(stringValue(prAction.RequestPayload["base_ref"]), workspace.BaseRef, "main")
		payload["title"] = firstNonEmpty(stringValue(prAction.RequestPayload["title"]), fmt.Sprintf("RSI proposal %s attempt %s for %s", proposal.ID, attempt.ID, proposalTargetRepo(cfg, proposal)))
		payload["body"] = firstNonEmpty(stringValue(prAction.RequestPayload["body"]), fmt.Sprintf("Automated draft PR for proposal %s attempt %s after workspace validation.", proposal.ID, attempt.ID))
	}
	nextOp, nextItem := proposalAttemptPhaseWork(cfg, proposal, attemptTrace, attempt, proposalOperationWorkspaceValidate, payload)
	if err := store.AdvanceProposalAttemptPhase(storepkg.ProposalAttemptPhaseAdvance{
		ProposalID:  proposal.ID,
		WorkItemID:  item.ID,
		OperationID: item.OperationID,
		Attempt:     &attempt,
		Workspace:   workspace,
		TraceID:     attemptTrace.Summary.TraceID,
		TraceUpdate: &storepkg.TraceUpdate{
			Events: []events.TraceEvent{
				{
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
				},
			},
			Reasoning: reasoning,
		},
		NextOperation: &nextOp,
		NextWorkItem:  &nextItem,
	}); err != nil {
		return err
	}
	return errProposalPhaseHandled
}

func processProposalWorkspaceValidate(cfg config.Config, store storepkg.Store, toolClient toolExecutor, proposal review.Proposal, trace events.Trace, item queue.WorkItem) error {
	attempt, attemptTrace, err := ensureProposalAttempt(cfg, store, proposal, trace, item)
	if err != nil {
		return err
	}
	workspace, ok := store.GetAttemptWorkspaceByAttempt(attempt.ID)
	if !ok {
		return fmt.Errorf("attempt %s missing workspace for validation phase", attempt.ID)
	}
	attempt.State = improvement.AttemptStateValidationRunning
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		return err
	}
	validationCommand := firstNonEmpty(strings.TrimSpace(stringValue(item.Payload["validation_command"])), workspaceValidationCommand(firstNonEmpty(attempt.ValidationPlan, proposal.ValidationPlan)))
	validationResult, execErr := toolClient.Execute("workspace.run_validation", map[string]any{
		"trace_id":     attemptTrace.Summary.TraceID,
		"workspace_id": workspace.ID,
		"attempt_id":   attempt.ID,
		"command":      validationCommand,
	})
	if execErr != nil || validationResult.Status != "ok" {
		return recordProposalPhaseAttemptFailure(cfg, store, proposal, attempt, attemptTrace, item, "sandbox_failure", firstNonEmpty(improvementActionError(validationResult, execErr), "Workspace validation failed."), false, improvement.AttemptTriggerSandboxFailed, &workspace, nil)
	}
	attempt.ValidationSummary = firstNonEmpty(stringValue(validationResult.Output["stdout"]), validationCommand)
	attempt.UpdatedAt = time.Now().UTC()
	job, err := ensureWorkspaceRepoChangeJob(store, proposal, attempt, workspace)
	if err != nil {
		return err
	}
	job.Status = string(review.ProposalValidationPending)
	job.ValidationError = ""
	job.ValidationRef = fmt.Sprintf("%s/%s", workspace.Namespace, firstNonEmpty(workspace.PodName, workspace.JobName))
	job.LogArtifactID = ""
	job.UpdatedAt = time.Now().UTC()
	proposal.Status = review.ProposalValidationPending
	branchName := firstNonEmpty(stringValue(item.Payload["branch_name"]), workspace.BranchName, attempt.BranchName)
	baseRef := firstNonEmpty(stringValue(item.Payload["base_ref"]), workspace.BaseRef, "main")
	title := strings.TrimSpace(stringValue(item.Payload["title"]))
	body := strings.TrimSpace(stringValue(item.Payload["body"]))
	if title == "" || body == "" {
		return recordProposalPhaseAttemptFailure(cfg, store, proposal, attempt, attemptTrace, item, "insufficient_evidence", "Proposal implement run completed without requesting a governed draft PR open.", true, improvement.AttemptTriggerProposalApproved, &workspace, &job)
	}
	now := time.Now().UTC()
	nextOp := operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: "pr_open",
		OperationKey:  "pr_open",
		Status:        operation.StatusQueued,
		Queue:         queue.ImprovementActionQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}
	nextItem := queue.WorkItem{
		Queue:          queue.ImprovementActionQueue,
		Kind:           "draft_pr_open",
		Status:         queue.WorkQueued,
		TraceID:        attemptTrace.Summary.TraceID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		TriggerEventID: proposal.OriginTraceID,
		ProposalID:     proposal.ID,
		RepoScope:      job.Repo,
		RequestedBy:    cfg.ServiceName,
		ApprovalMode:   "approved",
		CreatedAt:      now,
		UpdatedAt:      now,
		Payload: map[string]interface{}{
			"attempt_id":  attempt.ID,
			"job_id":      job.ID,
			"job_name":    job.SandboxJobName,
			"namespace":   job.SandboxNamespace,
			"repo":        job.Repo,
			"branch_name": branchName,
			"base_ref":    baseRef,
			"title":       title,
			"body":        body,
		},
	}
	if err := store.AdvanceProposalAttemptPhase(storepkg.ProposalAttemptPhaseAdvance{
		ProposalID:    proposal.ID,
		WorkItemID:    item.ID,
		OperationID:   item.OperationID,
		Proposal:      &proposal,
		Attempt:       &attempt,
		Workspace:     &workspace,
		RepoChangeJob: &job,
		TraceID:       attemptTrace.Summary.TraceID,
		TraceUpdate: &storepkg.TraceUpdate{
			Events: []events.TraceEvent{
				{
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
				},
			},
			Artifacts: []events.Artifact{
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
			},
			Reasoning: []events.ReasoningStep{
				{
					ID:         fmt.Sprintf("reason-workspace-validate-%d", now.UnixNano()),
					TraceID:    attemptTrace.Summary.TraceID,
					WorkflowID: attemptTrace.Summary.WorkflowID,
					StepType:   "workspace_validate",
					Summary:    fmt.Sprintf("Validated workspace-backed diff for attempt %s using %q.", attempt.ID, validationCommand),
					Confidence: 0.87,
					Decision:   branchName,
					CreatedAt:  now,
				},
			},
		},
		NextOperation: &nextOp,
		NextWorkItem:  &nextItem,
	}); err != nil {
		return err
	}
	return errProposalPhaseHandled
}

func workspaceMutationCallCount(trace events.Trace) int {
	count := 0
	for _, item := range trace.ToolCalls {
		switch strings.TrimSpace(item.ToolName) {
		case "workspace.write_file", "workspace.apply_patch":
			count++
		}
	}
	return count
}
