package improvementplane

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const (
	reconcileStatusHealthy             = "healthy"
	reconcileStatusEffectInFlight      = "effect_in_flight"
	reconcileStatusReconciliationNeeded = "reconciliation_needed"
	reconcileStatusWaitingOnPrereq     = "waiting_on_prerequisite"
	reconcileStatusTerminal            = "terminal"
)

func RunReconciler(cfg config.Config, store storepkg.Store) error {
	for {
		if err := RunReconcilePass(cfg, store); err != nil {
			log.Printf("improvement-plane reconcile error: %v", err)
		}
		time.Sleep(cfg.WorkerPollInterval)
	}
}

func RunReconcilePass(cfg config.Config, store storepkg.Store) error {
	var firstErr error
	for _, item := range activeRepoChangeAttempts(store) {
		if err := reconcileRepoChangeAttempt(cfg, store, item.proposal, item.attempt); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			log.Printf("improvement-plane reconcile proposal=%s attempt=%s error=%v", item.proposal.ID, item.attempt.ID, err)
		}
	}
	return firstErr
}

type repoChangeAttemptRef struct {
	proposal review.Proposal
	attempt  improvement.ChangeAttempt
}

func activeRepoChangeAttempts(store storepkg.Store) []repoChangeAttemptRef {
	out := make([]repoChangeAttemptRef, 0)
	for _, proposal := range store.ListProposals() {
		if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == "harness_overlay" {
			continue
		}
		attemptID := strings.TrimSpace(proposal.CurrentAttemptID)
		attempt, ok := store.GetChangeAttempt(attemptID)
		if !ok && attemptID == "" {
			attempt, ok = latestAttemptForProposal(store, proposal.ID)
		}
		if !ok {
			continue
		}
		if isPublicAttemptTerminal(improvement.PublicAttemptState(attempt.State)) {
			continue
		}
		out = append(out, repoChangeAttemptRef{proposal: proposal, attempt: attempt})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].proposal.CreatedAt.Equal(out[j].proposal.CreatedAt) {
			return out[i].attempt.CreatedAt.Before(out[j].attempt.CreatedAt)
		}
		return out[i].proposal.CreatedAt.Before(out[j].proposal.CreatedAt)
	})
	return out
}

func reconcileRepoChangeAttempt(cfg config.Config, store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt) error {
	state := improvement.PublicAttemptState(attempt.State)
	workspaces := filterAttemptWorkspacesByProposal(store.ListAttemptWorkspaces(), proposal.ID)
	workspace, workspaceOK := findAttemptWorkspaceByAttempt(workspaces, attempt.ID)
	validationRuns := filterValidationRunsByAttempt(store.ListValidationRuns(), proposal.ID, attempt.ID)
	prAttempts := filterPRAttemptsByAttempt(store.ListPRAttempts(), proposal.ID, attempt.ID)
	effects := store.ListEffectExecutionsByAggregate(transition.MachineAttempt, attempt.ID)

	switch state {
	case improvement.AttemptStatePlanned, improvement.AttemptStateWorkspaceRequired:
		return ensureWorkspaceSession(cfg, store, proposal, attempt, workspace, workspaceOK, effects, true)
	case improvement.AttemptStateImplementationRequired:
		if !workspaceSessionReady(workspace) {
			return ensureWorkspaceSession(cfg, store, proposal, attempt, workspace, workspaceOK, effects, false)
		}
		if hasLiveAttemptEffectOfKind(effects, transition.EffectInvokeRunner) {
			return nil
		}
		_, err := queueAttemptEffect(store, proposal, attempt, transition.EffectInvokeRunner, nextAttemptOperationID(attempt.ID, "implement", 0), 0, map[string]any{
			"workspace_id": workspace.ID,
			"repo":         workspace.Repo,
			"base_ref":     workspace.BaseRef,
			"branch_name":  workspace.BranchName,
		})
		return err
	case improvement.AttemptStateValidationRequired:
		if !workspaceSessionReady(workspace) {
			return ensureWorkspaceSession(cfg, store, proposal, attempt, workspace, workspaceOK, effects, false)
		}
		if strings.TrimSpace(attempt.RepoPatch) == "" {
			return nil
		}
		run, ok := liveValidationRun(validationRuns)
		if !ok {
			generation := nextValidationGeneration(validationRuns)
			operationID := nextAttemptOperationID(attempt.ID, "validate", generation)
			run = improvement.ValidationRun{
				ID:               fmt.Sprintf("validation-%s-%03d", attempt.ID, generation),
				ProposalID:       proposal.ID,
				AttemptID:        attempt.ID,
				ConversationID:   proposal.ConversationID,
				CaseID:           proposal.CaseID,
				OriginTraceID:    firstNonEmpty(attempt.AttemptTraceID, proposal.OriginTraceID, proposal.TraceID),
				WorkspaceID:      workspace.ID,
				OperationID:      operationID,
				Generation:       generation,
				Repo:             workspace.Repo,
				BranchName:       workspace.BranchName,
				Command:          workspaceValidationCommand(firstNonEmpty(attempt.ValidationPlan, proposal.ValidationPlan)),
				Status:           improvement.ValidationRunRequested,
				SandboxNamespace: workspace.Namespace,
				SandboxJobName:   workspace.JobName,
				SandboxPodName:   workspace.PodName,
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
			}
			if _, err := store.RecordValidationRun(run); err != nil {
				return err
			}
		}
		if hasLiveAttemptEffectOfKind(effects, transition.EffectWorkspaceValidate) {
			return nil
		}
		_, err := queueAttemptEffect(store, proposal, attempt, transition.EffectWorkspaceValidate, run.OperationID, run.Generation, map[string]any{
			"workspace_id":       workspace.ID,
			"validation_run_id":  run.ID,
			"validation_command": firstNonEmpty(run.Command, workspaceValidationCommand(firstNonEmpty(attempt.ValidationPlan, proposal.ValidationPlan))),
			"repo":               workspace.Repo,
			"base_ref":           workspace.BaseRef,
			"branch_name":        workspace.BranchName,
		})
		return err
	case improvement.AttemptStatePRRequired:
		if !hasPassedValidation(validationRuns) {
			return nil
		}
		prAttempt, ok := livePRAttempt(prAttempts)
		if !ok {
			generation := nextPRAttemptGeneration(prAttempts)
			operationID := nextAttemptOperationID(attempt.ID, "pr", generation)
			prAttempt = improvement.PRAttempt{
				ID:               fmt.Sprintf("pr-%s", attempt.ID),
				ProposalID:       proposal.ID,
				AttemptID:        attempt.ID,
				ConversationID:   proposal.ConversationID,
				CaseID:           proposal.CaseID,
				OriginTraceID:    firstNonEmpty(attempt.AttemptTraceID, proposal.OriginTraceID, proposal.TraceID),
				OperationID:      operationID,
				Generation:       generation,
				Repo:             firstNonEmpty(workspace.Repo, proposalTargetRepo(cfg, proposal)),
				BranchName:       firstNonEmpty(workspace.BranchName, attempt.BranchName),
				Status:           "requested",
				ValidationStatus: "passed",
				CreatedAt:        time.Now().UTC(),
			}
			if _, err := store.RecordPRAttempt(prAttempt); err != nil {
				return err
			}
		}
		if hasLiveAttemptEffectOfKind(effects, transition.EffectOpenDraftPR) || strings.EqualFold(prAttempt.Status, "open") || prAttempt.Status == string(review.ProposalPROpen) {
			return nil
		}
		_, err := queueAttemptEffect(store, proposal, attempt, transition.EffectOpenDraftPR, prAttempt.OperationID, prAttempt.Generation, map[string]any{
			"pr_attempt_id": prAttempt.ID,
			"repo":          prAttempt.Repo,
			"branch_name":   prAttempt.BranchName,
			"base_ref":      firstNonEmpty(workspace.BaseRef, "main"),
		})
		return err
	default:
		return nil
	}
}

func ensureWorkspaceSession(cfg config.Config, store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, workspace improvement.AttemptWorkspace, workspaceOK bool, effects []transition.EffectExecution, replaceInvalid bool) error {
	if hasLiveAttemptEffectOfKind(effects, transition.EffectOpenWorkspace) {
		return nil
	}
	generation := 1
	operationID := ""
	needsReplacement := !workspaceOK || workspaceSessionRepairable(workspace)
	if workspaceOK {
		generation = maxInt(workspace.Generation, 0)
		if needsReplacement || replaceInvalid {
			generation++
		}
		if generation <= 0 {
			generation = 1
		}
		if !needsReplacement && workspace.OperationID != "" {
			operationID = workspace.OperationID
		}
	}
	if operationID == "" {
		operationID = nextAttemptOperationID(attempt.ID, "workspace", generation)
	}
	session := improvement.AttemptWorkspace{
		ID:          firstNonEmpty(workspace.ID, fmt.Sprintf("workspace-%s", attempt.ID)),
		AttemptID:   attempt.ID,
		ProposalID:  proposal.ID,
		OperationID: operationID,
		Generation:  generation,
		Repo:        firstNonEmpty(workspace.Repo, proposalTargetRepo(cfg, proposal)),
		BaseRef:     firstNonEmpty(workspace.BaseRef, "main"),
		BranchName:  firstNonEmpty(workspace.BranchName, attempt.BranchName, buildAttemptBranchName(proposal.ID, attempt.AttemptNumber)),
		Status:      improvement.WorkspaceQueued,
		LastError:   "",
		Repairable:  false,
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   time.Now().UTC(),
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = session.UpdatedAt
	}
	if len(workspace.AllowedPathGlobs) > 0 {
		session.AllowedPathGlobs = append([]string(nil), workspace.AllowedPathGlobs...)
	} else {
		session.AllowedPathGlobs = defaultWorkspaceAllowedPathGlobs()
	}
	if _, err := store.RecordAttemptWorkspace(session); err != nil {
		return err
	}
	_, err := queueAttemptEffect(store, proposal, attempt, transition.EffectOpenWorkspace, operationID, generation, map[string]any{
		"workspace_id": session.ID,
		"repo":         session.Repo,
		"base_ref":     session.BaseRef,
		"branch_name":  session.BranchName,
	})
	return err
}

func queueAttemptEffect(store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, kind transition.EffectKind, operationID string, generation int, payload map[string]any) (transition.EffectExecution, error) {
	now := time.Now().UTC()
	if payload == nil {
		payload = map[string]any{}
	}
	payload["proposal_id"] = proposal.ID
	payload["attempt_id"] = attempt.ID
	payload["trace_id"] = firstNonEmpty(attempt.AttemptTraceID, proposal.OriginTraceID, proposal.TraceID)
	payload["operation_id"] = operationID
	if generation > 0 {
		payload["operation_generation"] = generation
	}
	effect := transition.EffectExecution{
		ID:             fmt.Sprintf("effect-%s-%s-%d", attempt.ID, string(kind), now.UnixNano()),
		MachineKind:    transition.MachineAttempt,
		AggregateID:    attempt.ID,
		AttemptID:      attempt.ID,
		EffectKind:     kind,
		Status:         transition.EffectQueued,
		IdempotencyKey: fmt.Sprintf("attempt:%s:%s:%s", attempt.ID, kind, operationID),
		Payload:        payload,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	queued, _, err := store.QueueEffectExecution(effect)
	return queued, err
}

func nextAttemptOperationID(attemptID string, phase string, generation int) string {
	if generation > 0 {
		return fmt.Sprintf("%s:%s:%03d", attemptID, phase, generation)
	}
	return fmt.Sprintf("%s:%s:%d", attemptID, phase, time.Now().UTC().UnixNano())
}

func hasLiveAttemptEffectOfKind(items []transition.EffectExecution, kind transition.EffectKind) bool {
	for _, item := range items {
		if item.EffectKind != kind {
			continue
		}
		if item.Status == transition.EffectQueued || item.Status == transition.EffectRunning {
			return true
		}
	}
	return false
}

func workspaceSessionRepairable(item improvement.AttemptWorkspace) bool {
	if item.Repairable {
		return true
	}
	if workspaceSessionMissingProviderIdentity(item) {
		return true
	}
	switch item.Status {
	case improvement.WorkspaceFailed, improvement.WorkspaceClosed:
		return true
	default:
		return false
	}
}

func workspaceSessionMissingProviderIdentity(item improvement.AttemptWorkspace) bool {
	return strings.TrimSpace(item.Namespace) == "" || strings.TrimSpace(item.JobName) == ""
}

func workspaceSessionReady(item improvement.AttemptWorkspace) bool {
	if workspaceSessionMissingProviderIdentity(item) || item.Repairable {
		return false
	}
	switch item.Status {
	case improvement.WorkspaceReady, improvement.WorkspaceExecuting, improvement.WorkspaceValidating, improvement.WorkspaceCompleted:
		return true
	default:
		return false
	}
}

func liveValidationRun(items []improvement.ValidationRun) (improvement.ValidationRun, bool) {
	for _, item := range items {
		switch item.Status {
		case improvement.ValidationRunRequested, improvement.ValidationRunRunning:
			return item, true
		}
	}
	return improvement.ValidationRun{}, false
}

func hasPassedValidation(items []improvement.ValidationRun) bool {
	for _, item := range items {
		if item.Status == improvement.ValidationRunPassed {
			return true
		}
	}
	return false
}

func nextValidationGeneration(items []improvement.ValidationRun) int {
	maxGeneration := 0
	for _, item := range items {
		if item.Generation > maxGeneration {
			maxGeneration = item.Generation
		}
	}
	return maxGeneration + 1
}

func livePRAttempt(items []improvement.PRAttempt) (improvement.PRAttempt, bool) {
	for _, item := range items {
		switch strings.TrimSpace(item.Status) {
		case "requested", "opening":
			return item, true
		case "open", string(review.ProposalPROpen):
			return item, true
		}
	}
	return improvement.PRAttempt{}, false
}

func nextPRAttemptGeneration(items []improvement.PRAttempt) int {
	maxGeneration := 0
	for _, item := range items {
		if item.Generation > maxGeneration {
			maxGeneration = item.Generation
		}
	}
	return maxGeneration + 1
}

func isPublicAttemptTerminal(state improvement.ChangeAttemptState) bool {
	switch state {
	case improvement.AttemptStateMerged,
		improvement.AttemptStateClosedUnmerged,
		improvement.AttemptStateCIFailed,
		improvement.AttemptStateNeedsReview,
		improvement.AttemptStateAbandoned,
		improvement.AttemptStateSuperseded:
		return true
	default:
		return false
	}
}

func requiredResourceKindForAttemptState(state improvement.ChangeAttemptState) string {
	switch improvement.PublicAttemptState(state) {
	case improvement.AttemptStatePlanned, improvement.AttemptStateWorkspaceRequired:
		return "workspace_session"
	case improvement.AttemptStateImplementationRequired:
		return "implementation"
	case improvement.AttemptStateValidationRequired:
		return "validation_run"
	case improvement.AttemptStatePRRequired:
		return "pr_attempt"
	default:
		return ""
	}
}

func reconcileStatusForAttempt(state improvement.ChangeAttemptState, attemptID string, workspaces []improvement.AttemptWorkspace, validationRuns []improvement.ValidationRun, prAttempts []improvement.PRAttempt) string {
	state = improvement.PublicAttemptState(state)
	if isPublicAttemptTerminal(state) {
		return reconcileStatusTerminal
	}
	workspace, workspaceOK := findAttemptWorkspaceByAttempt(workspaces, attemptID)
	switch state {
	case improvement.AttemptStatePlanned, improvement.AttemptStateWorkspaceRequired:
		if workspaceOK && !workspaceSessionRepairable(workspace) {
			return reconcileStatusHealthy
		}
		return reconcileStatusReconciliationNeeded
	case improvement.AttemptStateImplementationRequired:
		if workspaceSessionReady(workspace) {
			return reconcileStatusHealthy
		}
		return reconcileStatusWaitingOnPrereq
	case improvement.AttemptStateValidationRequired:
		if !workspaceSessionReady(workspace) {
			return reconcileStatusWaitingOnPrereq
		}
		if _, ok := liveValidationRun(validationRuns); ok || hasPassedValidation(validationRuns) {
			return reconcileStatusHealthy
		}
		return reconcileStatusReconciliationNeeded
	case improvement.AttemptStatePRRequired:
		if !hasPassedValidation(validationRuns) {
			return reconcileStatusWaitingOnPrereq
		}
		if prAttempt, ok := livePRAttempt(prAttempts); ok && strings.TrimSpace(prAttempt.Status) != "" {
			return reconcileStatusHealthy
		}
		return reconcileStatusReconciliationNeeded
	default:
		return reconcileStatusHealthy
	}
}

func filterValidationRunsByAttempt(items []improvement.ValidationRun, proposalID string, attemptID string) []improvement.ValidationRun {
	out := make([]improvement.ValidationRun, 0)
	for _, item := range items {
		if item.ProposalID == proposalID && item.AttemptID == attemptID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}
