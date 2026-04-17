package improvementplane

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const attemptAutoRetryLimit = 3

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func latestAttemptForProposal(store storepkg.Store, proposalID string) (improvement.ChangeAttempt, bool) {
	items := attemptsForProposal(store, proposalID)
	if len(items) == 0 {
		return improvement.ChangeAttempt{}, false
	}
	return items[0], true
}

func latestCandidateForTrace(store storepkg.Store, traceID string) (improvement.Candidate, bool) {
	var (
		best   improvement.Candidate
		found  bool
		bestTS time.Time
	)
	for _, item := range store.ListCandidates() {
		matches := item.LatestTraceID == traceID
		if !matches {
			for _, evidence := range item.EvidenceTraceIDs {
				if evidence == traceID {
					matches = true
					break
				}
			}
		}
		if !matches {
			continue
		}
		if !found || item.UpdatedAt.After(bestTS) {
			best = item
			bestTS = item.UpdatedAt
			found = true
		}
	}
	return best, found
}

func latestApprovedProposalForCandidate(store storepkg.Store, candidateKey string) (review.Proposal, bool) {
	var (
		best   review.Proposal
		found  bool
		bestTS time.Time
	)
	for _, item := range store.ListProposals() {
		if item.CandidateKey != candidateKey || item.Status != review.ProposalApproved {
			continue
		}
		if !found || item.CreatedAt.After(bestTS) {
			best = item
			bestTS = item.CreatedAt
			found = true
		}
	}
	return best, found
}

func hasLiveAttemptExecutionEffect(store storepkg.Store, proposal review.Proposal) bool {
	attemptID := strings.TrimSpace(proposal.CurrentAttemptID)
	if attemptID == "" {
		return false
	}
	for _, effect := range store.ListEffectExecutionsByAggregate(transition.MachineAttempt, attemptID) {
		switch effect.Status {
		case transition.EffectQueued, transition.EffectRunning:
			return true
		}
	}
	return false
}

func hasInFlightRepoChange(store storepkg.Store, proposalID string) bool {
	for _, item := range store.ListRepoChangeJobs() {
		if item.ProposalID != proposalID {
			continue
		}
		switch item.Status {
		case string(review.ProposalRepoChangeQueued), string(review.ProposalRepoChangeRunning), string(review.ProposalValidationPending), string(review.ProposalPROpen):
			return true
		}
	}
	for _, item := range store.ListPRAttempts() {
		if item.ProposalID == proposalID && item.Status == string(review.ProposalPROpen) {
			return true
		}
	}
	return false
}

func ensureApprovedProposalWork(store storepkg.Store, trace events.Trace, requestedBy string) error {
	candidate, ok := latestCandidateForTrace(store, trace.Summary.TraceID)
	if !ok || candidate.LineStatus != improvement.LineActive {
		return nil
	}
	proposal, ok := latestApprovedProposalForCandidate(store, candidate.CandidateKey)
	if !ok {
		return nil
	}
	if !review.ProposalExecutableIntervention(proposal.RecommendedInterventionKind) {
		return nil
	}
	if hasLiveAttemptExecutionEffect(store, proposal) || hasInFlightRepoChange(store, proposal.ID) {
		return nil
	}
	commandID := fmt.Sprintf("cmd-proposal-resume:%s:%s", proposal.ID, trace.Summary.TraceID)
	_, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposal.ID,
		CommandKind: string(transition.CommandProposalResumeExecution),
		CommandID:   commandID,
		Actor:       requestedBy,
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"candidate_key": candidate.CandidateKey,
			"risk_tier":     proposal.RiskTier,
			"trigger":       string(improvement.AttemptTriggerOperatorRetry),
			"source_trace":  trace.Summary.TraceID,
		},
	})
	return err
}

func submitAttemptCommand(store storepkg.Store, attempt improvement.ChangeAttempt, kind transition.AttemptPhaseCommandKind, actor string, occurredAt time.Time, payload map[string]any) error {
	attemptID := strings.TrimSpace(attempt.ID)
	if attemptID == "" {
		return nil
	}
	commandID := fmt.Sprintf("cmd-attempt:%s:%s", attemptID, string(kind))
	if operationID := strings.TrimSpace(stringValue(payload["operation_id"])); operationID != "" {
		commandID += ":" + operationID
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAttempt,
		AggregateID: attemptID,
		CommandKind: string(kind),
		CommandID:   commandID,
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		return err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return errors.New(receipt.Reason)
	}
	return nil
}

func submitProposalCommand(store storepkg.Store, proposal review.Proposal, kind transition.ProposalLineCommandKind, actor string, occurredAt time.Time, commandID string, rationale string) error {
	proposalID := strings.TrimSpace(proposal.ID)
	if proposalID == "" {
		return nil
	}
	if strings.TrimSpace(commandID) == "" {
		commandID = fmt.Sprintf("cmd-proposal:%s:%s", proposalID, string(kind))
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposalID,
		CommandKind: string(kind),
		CommandID:   commandID,
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload: map[string]any{
			"rationale":       rationale,
			"reviewer_id":     actor,
			"idempotency_key": commandID,
			"scope":           string(review.FeedbackScopeLine),
		},
	})
	if err != nil {
		return err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return errors.New(receipt.Reason)
	}
	return nil
}

func attemptsForProposal(store storepkg.Store, proposalID string) []improvement.ChangeAttempt {
	out := make([]improvement.ChangeAttempt, 0)
	for _, item := range store.ListChangeAttempts() {
		if item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].AttemptNumber == out[j].AttemptNumber {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].AttemptNumber > out[j].AttemptNumber
	})
	return out
}

func buildAttemptBranchName(proposalID string, attemptNumber int) string {
	return fmt.Sprintf("codex/%s/attempt-%02d", proposalID, attemptNumber)
}

func patchChangedFiles(patch string) []string {
	files := []string{}
	seen := map[string]struct{}{}
	for _, line := range strings.Split(patch, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "+++ b/") {
			path := strings.TrimPrefix(line, "+++ b/")
			if path == "/dev/null" || path == "" {
				continue
			}
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			files = append(files, path)
		}
	}
	return files
}

func validateRepoPatch(output string, changedFiles []string) (string, string, []string) {
	patch := strings.TrimSpace(output)
	if patch == "" {
		return "no_op_diff", "Proposal runner returned an empty repo patch.", []string{}
	}
	if len(changedFiles) == 0 {
		changedFiles = patchChangedFiles(patch)
	}
	if len(changedFiles) == 0 {
		return "no_op_diff", "Generated patch does not touch any files.", []string{}
	}
	allMeta := true
	for _, file := range changedFiles {
		clean := filepath.Clean(strings.TrimSpace(file))
		if !strings.HasPrefix(clean, ".rsi/") {
			allMeta = false
		}
		if !isAllowedRepoFile(clean) {
			return "wrong_fix_surface", fmt.Sprintf("Generated patch touches disallowed file %s.", clean), changedFiles
		}
	}
	if allMeta {
		return "bad_patch", "Generated patch only changes .rsi/** metadata files.", changedFiles
	}
	return "", "", changedFiles
}

func isAllowedRepoFile(path string) bool {
	switch {
	case strings.HasPrefix(path, "cmd/"),
		strings.HasPrefix(path, "internal/"),
		strings.HasPrefix(path, "runner/"),
		strings.HasPrefix(path, "ui/"):
		return true
	case path == "README.md", path == "Makefile":
		return true
	default:
		return false
	}
}

func isAttemptTerminal(state improvement.ChangeAttemptState) bool {
	switch state {
	case improvement.AttemptStateSandboxFailed,
		improvement.AttemptStateCIFailed,
		improvement.AttemptStateClosedUnmerged,
		improvement.AttemptStateOverlayActive,
		improvement.AttemptStateMerged,
		improvement.AttemptStateNeedsReview,
		improvement.AttemptStateAbandoned,
		improvement.AttemptStateSuperseded:
		return true
	default:
		return false
	}
}

func shouldAutoRetryAttempt(attempt improvement.ChangeAttempt, failureClass string, materialChange bool) bool {
	if materialChange {
		return false
	}
	switch failureClass {
	case "bad_patch", "ci_regression", "no_op_diff", "wrong_fix_surface", "sandbox_failure", "stale_branch", "merge_conflict":
	default:
		return false
	}
	return attempt.AttemptNumber < attemptAutoRetryLimit
}

func latestRelevantAttemptEffect(store storepkg.Store, attempt improvement.ChangeAttempt) (transition.EffectExecution, bool) {
	var (
		current   transition.EffectExecution
		found     bool
		updatedAt time.Time
	)
	for _, item := range store.ListEffectExecutionsByAggregate(transition.MachineAttempt, attempt.ID) {
		if item.Status != transition.EffectQueued && item.Status != transition.EffectRunning {
			continue
		}
		if !attemptEffectMatchesState(item.EffectKind, attempt.State) {
			continue
		}
		if !found || item.UpdatedAt.After(updatedAt) {
			current = item
			updatedAt = item.UpdatedAt
			found = true
		}
	}
	return current, found
}

func attemptEffectMatchesState(kind transition.EffectKind, state improvement.ChangeAttemptState) bool {
	switch state {
	case improvement.AttemptStatePatchGenerated, improvement.AttemptStateOverlayGenerated:
		return kind == transition.EffectWorkspaceValidate
	case improvement.AttemptStateValidationRunning, improvement.AttemptStateOverlayValidating:
		return kind == transition.EffectWorkspaceValidate || kind == transition.EffectObserveWorkspaceValidation || kind == transition.EffectOpenDraftPR
	case improvement.AttemptStatePROpen, improvement.AttemptStateCIObserving:
		return false
	case improvement.AttemptStateSandboxFailed,
		improvement.AttemptStateCIFailed,
		improvement.AttemptStateClosedUnmerged,
		improvement.AttemptStateMerged,
		improvement.AttemptStateNeedsReview,
		improvement.AttemptStateAbandoned,
		improvement.AttemptStateSuperseded,
		improvement.AttemptStateOverlayActive:
		return false
	default:
		return kind == transition.EffectOpenWorkspace || kind == transition.EffectInvokeRunner
	}
}

func failureAttemptCommand(store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, failureClass string, retryable bool) (transition.AttemptPhaseCommandKind, string) {
	if currentAttempt, ok := store.GetChangeAttempt(attempt.ID); ok {
		attempt = currentAttempt
	}
	currentEffect, ok := latestRelevantAttemptEffect(store, attempt)
	currentEffectKind := transition.EffectKind("")
	currentOperationID := ""
	if ok {
		currentEffectKind = currentEffect.EffectKind
		currentOperationID = strings.TrimSpace(stringValue(currentEffect.Payload["operation_id"]))
	}
	switch strings.TrimSpace(failureClass) {
	case "ci_regression":
		return transition.CommandAttemptCIFailed, currentOperationID
	case "closed_unmerged":
		return transition.CommandAttemptClosedUnmerged, currentOperationID
	case "stale_branch":
		if retryable {
			return transition.CommandPROpenFailedRetryable, currentOperationID
		}
		return transition.CommandPROpenFailedReview, currentOperationID
	}
	switch currentEffectKind {
	case transition.EffectOpenWorkspace:
		if retryable {
			return transition.CommandWorkspaceFailedRetryable, currentOperationID
		}
		return transition.CommandWorkspaceFailedReview, currentOperationID
	case transition.EffectWorkspaceValidate, transition.EffectObserveWorkspaceValidation:
		if retryable {
			return transition.CommandValidationFailedRetryable, currentOperationID
		}
		return transition.CommandValidationFailedReview, currentOperationID
	case transition.EffectOpenDraftPR:
		if retryable {
			return transition.CommandPROpenFailedRetryable, currentOperationID
		}
		return transition.CommandPROpenFailedReview, currentOperationID
	case transition.EffectInvokeRunner:
		if retryable {
			return transition.CommandImplementationFailedRetryable, currentOperationID
		}
		return transition.CommandImplementationFailedReview, currentOperationID
	}
	switch attempt.State {
	case improvement.AttemptStateValidationRunning, improvement.AttemptStateOverlayValidating:
		if retryable {
			return transition.CommandValidationFailedRetryable, currentOperationID
		}
		return transition.CommandValidationFailedReview, currentOperationID
	case improvement.AttemptStatePROpen, improvement.AttemptStateCIObserving:
		if retryable {
			return transition.CommandPROpenFailedRetryable, currentOperationID
		}
		return transition.CommandPROpenFailedReview, currentOperationID
	}
	switch proposal.Status {
	case review.ProposalRepoChangeQueued:
		if retryable {
			return transition.CommandWorkspaceFailedRetryable, currentOperationID
		}
		return transition.CommandWorkspaceFailedReview, currentOperationID
	case review.ProposalValidationPending:
		if retryable {
			return transition.CommandValidationFailedRetryable, currentOperationID
		}
		return transition.CommandValidationFailedReview, currentOperationID
	}
	if retryable {
		return transition.CommandImplementationFailedRetryable, currentOperationID
	}
	return transition.CommandImplementationFailedReview, currentOperationID
}

func attemptFailureTraceStatus(failureClass string, retryable bool) string {
	switch strings.TrimSpace(failureClass) {
	case "stale_branch":
		if !retryable {
			return string(events.StatusNeedsHuman)
		}
		return string(events.StatusFailed)
	default:
		return string(events.StatusFailed)
	}
}

func attemptFailureTraceEvents(cfg config.Config, trace events.Trace, failureClass string, failureSummary string, retryable bool, now time.Time) []events.TraceEvent {
	items := make([]events.TraceEvent, 0, 2)
	if strings.TrimSpace(failureClass) == "stale_branch" {
		status := events.StatusFailed
		if !retryable {
			status = events.StatusNeedsHuman
		}
		items = append(items, events.TraceEvent{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "improvement",
			Service:     cfg.ServiceName,
			Actor:       "worker",
			EventType:   "github.pr.blocked",
			Status:      status,
			StartedAt:   now,
			EndedAt:     ptrTime(now),
			Description: failureSummary,
		})
	}
	items = append(items, events.TraceEvent{
		TraceID:     trace.Summary.TraceID,
		IngestionID: trace.Summary.IngestionID,
		WorkflowID:  trace.Summary.WorkflowID,
		Plane:       "improvement",
		Service:     cfg.ServiceName,
		Actor:       "attempt-supervisor",
		EventType:   "change_attempt.failed",
		Status:      events.StatusFailed,
		StartedAt:   now,
		EndedAt:     ptrTime(now),
		Description: failureSummary,
	})
	return items
}

type attemptFailureTraceExtras struct {
	Events    []events.TraceEvent
	Artifacts []events.Artifact
	Payload   map[string]any
}

func recordAttemptFailure(cfg config.Config, store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, trace events.Trace, failureClass string, failureSummary string, materialChange bool, trigger improvement.ChangeAttemptTrigger, extras ...attemptFailureTraceExtras) error {
	_ = trigger
	if attempt.AttemptTraceID != "" {
		if attemptTrace, ok := store.GetTrace(attempt.AttemptTraceID); ok {
			trace = attemptTrace
		}
	}
	now := time.Now().UTC()
	retryDecision := "needs_review"
	proposalCommand := transition.CommandProposalNeedsReview
	retryable := shouldAutoRetryAttempt(attempt, failureClass, materialChange)
	if retryable {
		retryDecision = "auto_retry"
		proposalCommand = transition.CommandProposalRetryableFailure
	}
	attemptCommand, operationID := failureAttemptCommand(store, proposal, attempt, failureClass, retryable)
	var extra attemptFailureTraceExtras
	if len(extras) > 0 {
		extra = extras[0]
	}
	traceEvents := append([]events.TraceEvent(nil), extra.Events...)
	traceEvents = append(traceEvents, attemptFailureTraceEvents(cfg, trace, failureClass, failureSummary, retryable, now)...)
	retryAt := now.Add(time.Minute)
	payload := map[string]any{
		"failure_class":              failureClass,
		"failure_summary":            failureSummary,
		"validation_summary":         failureSummary,
		"retry_decision":             retryDecision,
		"retry_after":                retryAt.Format(time.RFC3339),
		"trace_status":               attemptFailureTraceStatus(failureClass, retryable),
		"material_hypothesis_change": materialChange,
		"trace_events":               traceEvents,
		"trace_artifacts":            append([]events.Artifact(nil), extra.Artifacts...),
		"reasoning_steps": []events.ReasoningStep{
			{
				ID:         fmt.Sprintf("reason-attempt-failed-%d", now.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "retry_decision",
				Summary:    failureSummary,
				Confidence: 0.88,
				Decision:   retryDecision,
				CreatedAt:  now,
			},
		},
	}
	for key, value := range extra.Payload {
		payload[key] = value
	}
	if operationID != "" {
		payload["operation_id"] = operationID
	}
	if err := submitAttemptCommand(store, attempt, attemptCommand, cfg.ServiceName, now, payload); err != nil {
		return err
	}
	commandID := fmt.Sprintf("cmd-proposal-attempt-failure:%s:%s", attempt.ID, failureClass)
	return submitProposalCommand(store, proposal, proposalCommand, cfg.ServiceName, now, commandID, failureSummary)
}

func payloadString(raw any) string {
	return strings.TrimSpace(stringValue(raw))
}
