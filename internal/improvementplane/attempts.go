package improvementplane

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const attemptAutoRetryLimit = 3

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

func hasLiveProposalQueueWork(store storepkg.Store, proposalID string) bool {
	for _, item := range store.ListWorkItems() {
		if item.Queue != queue.ProposalQueue || item.Kind != "approved_proposal" || item.ProposalID != proposalID {
			continue
		}
		if item.Status == queue.WorkQueued || item.Status == queue.WorkLeased {
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
	if hasLiveProposalQueueWork(store, proposal.ID) || hasInFlightRepoChange(store, proposal.ID) {
		return nil
	}
	if attempt, ok := latestAttemptForProposal(store, proposal.ID); ok && !isAttemptTerminal(attempt.State) {
		return nil
	}
	_, err := store.EnqueueWorkItem(queue.WorkItem{
		Queue:          queue.ProposalQueue,
		Kind:           "approved_proposal",
		Status:         queue.WorkQueued,
		TraceID:        proposal.TraceID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		TriggerEventID: proposal.OriginTraceID,
		ProposalID:     proposal.ID,
		RequestedBy:    requestedBy,
		ApprovalMode:   "human_review",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Payload: map[string]any{
			"candidate_key": candidate.CandidateKey,
			"risk_tier":     proposal.RiskTier,
			"dedupe_key":    fmt.Sprintf("proposal-runner:%s", proposal.ID),
			"trigger":       string(improvement.AttemptTriggerOperatorRetry),
		},
	})
	return err
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

func ensureProposalAttempt(cfg config.Config, store storepkg.Store, proposal review.Proposal, sourceTrace events.Trace, item queue.WorkItem) (improvement.ChangeAttempt, events.Trace, error) {
	if strings.TrimSpace(proposal.CurrentAttemptID) != "" {
		if existing, ok := store.GetChangeAttempt(proposal.CurrentAttemptID); ok && !isAttemptTerminal(existing.State) {
			trace, ok := store.GetTrace(existing.AttemptTraceID)
			if !ok {
				return improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("attempt trace %s not found", existing.AttemptTraceID)
			}
			return existing, trace, nil
		}
	}
	latest, hasLatest := latestAttemptForProposal(store, proposal.ID)
	nextNumber := 1
	parentAttemptID := stringValue(item.Payload["parent_attempt"])
	if hasLatest {
		nextNumber = latest.AttemptNumber + 1
		if parentAttemptID == "" {
			parentAttemptID = latest.ID
		}
	}
	trigger := improvement.AttemptTriggerProposalApproved
	if raw := strings.TrimSpace(stringValue(item.Payload["trigger"])); raw != "" {
		trigger = improvement.ChangeAttemptTrigger(raw)
	}
	attempt := improvement.ChangeAttempt{
		ID:              fmt.Sprintf("attempt-%s", strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(proposal.ID)), "/", "-"), " ", "-")),
		ProposalID:      proposal.ID,
		CandidateKey:    proposal.CandidateKey,
		AttemptNumber:   nextNumber,
		TargetLayer:     proposal.TargetLayer,
		TargetKind:      proposal.TargetKind,
		TargetRef:       proposal.TargetRef,
		Trigger:         trigger,
		State:           improvement.AttemptStatePlanning,
		ParentAttemptID: parentAttemptID,
		BranchName:      buildAttemptBranchName(proposal.ID, nextNumber),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	attempt.ID = fmt.Sprintf("attempt-%s-%02d", strings.ReplaceAll(strings.TrimSpace(proposal.ID), "/", "-"), nextNumber)
	description := fmt.Sprintf("Queued remediation attempt %d for proposal %s triggered by %s.", nextNumber, proposal.ID, trigger)
	attemptTrace, _, err := store.CreateDerivedTrace(storepkg.DerivedTraceRequest{
		SourceTraceID:  sourceTrace.Summary.TraceID,
		ProposalID:     proposal.ID,
		AttemptID:      attempt.ID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		ThreadKey:      sourceTrace.Summary.ThreadKey,
		WorkflowKind:   "proposal_attempt",
		RequestedBy:    cfg.ServiceName,
		Description:    description,
		TriggerEventID: sourceTrace.Summary.TriggerEventID,
		IngestionID:    sourceTrace.Summary.IngestionID,
		CreatedAt:      attempt.CreatedAt,
	})
	if err != nil {
		return improvement.ChangeAttempt{}, events.Trace{}, err
	}
	attempt.AttemptTraceID = attemptTrace.Summary.TraceID
	attempt.BranchName = buildAttemptBranchName(proposal.ID, nextNumber)
	attempt, err = store.UpsertChangeAttempt(attempt)
	if err != nil {
		return improvement.ChangeAttempt{}, events.Trace{}, err
	}
	return attempt, attemptTrace, nil
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

func queueProposalAttemptRetry(store storepkg.Store, proposal review.Proposal, failedAttempt improvement.ChangeAttempt, requestedBy string, trigger improvement.ChangeAttemptTrigger) error {
	_, err := store.EnqueueWorkItem(queue.WorkItem{
		Queue:          queue.ProposalQueue,
		Kind:           "approved_proposal",
		Status:         queue.WorkQueued,
		TraceID:        proposal.TraceID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		TriggerEventID: proposal.OriginTraceID,
		ProposalID:     proposal.ID,
		RequestedBy:    requestedBy,
		ApprovalMode:   "human_review",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Payload: map[string]any{
			"candidate_key":  proposal.CandidateKey,
			"risk_tier":      proposal.RiskTier,
			"trigger":        string(trigger),
			"parent_attempt": failedAttempt.ID,
			"dedupe_key":     fmt.Sprintf("proposal-runner:%s:attempt:%d", proposal.ID, failedAttempt.AttemptNumber+1),
		},
	})
	return err
}

func recordAttemptFailure(cfg config.Config, store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, trace events.Trace, failureClass string, failureSummary string, materialChange bool, trigger improvement.ChangeAttemptTrigger) error {
	now := time.Now().UTC()
	switch failureClass {
	case "sandbox_failure":
		attempt.State = improvement.AttemptStateSandboxFailed
	case "ci_regression":
		attempt.State = improvement.AttemptStateCIFailed
	case "closed_unmerged":
		attempt.State = improvement.AttemptStateClosedUnmerged
	default:
		attempt.State = improvement.AttemptStateNeedsReview
	}
	retryDecision := "needs_review"
	nextStatus := review.ProposalPendingReview
	if shouldAutoRetryAttempt(attempt, failureClass, materialChange) {
		retryDecision = "auto_retry"
		nextStatus = review.ProposalApproved
		attempt.RetryAfter = ptrTime(time.Now().UTC().Add(time.Minute))
	} else {
		attempt.State = improvement.AttemptStateNeedsReview
	}
	attempt.FailureClass = failureClass
	attempt.FailureSummary = failureSummary
	attempt.RetryDecision = retryDecision
	attempt.MaterialHypothesisChange = materialChange
	attempt.UpdatedAt = now
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		return err
	}
	if _, err := store.UpdateProposalStatus(proposal.ID, nextStatus); err != nil {
		return err
	}
	if shouldAutoRetryAttempt(attempt, failureClass, materialChange) {
		if err := queueProposalAttemptRetry(store, proposal, attempt, cfg.ServiceName, trigger); err != nil {
			return err
		}
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Status: ptrStatus(events.StatusFailed),
		Events: []events.TraceEvent{
			{
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
			},
		},
		Reasoning: []events.ReasoningStep{
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
	})
	return nil
}
