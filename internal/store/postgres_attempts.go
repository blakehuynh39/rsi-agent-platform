package store

import (
	"database/sql"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

func (p *PostgresStore) ListChangeAttempts() []improvement.ChangeAttempt {
	store, err := p.readStore()
	if err != nil {
		return []improvement.ChangeAttempt{}
	}
	return store.ListChangeAttempts()
}

func (p *PostgresStore) GetChangeAttempt(attemptID string) (improvement.ChangeAttempt, bool) {
	var item improvement.ChangeAttempt
	err := p.withTx(func(tx *sql.Tx) error {
		var err error
		item, err = selectChangeAttemptTx(tx, strings.TrimSpace(attemptID), false)
		return err
	})
	if err != nil {
		return improvement.ChangeAttempt{}, false
	}
	return normalizeChangeAttempt(item), true
}

func (p *PostgresStore) StopProposalLine(proposalID string, requestedBy string, rationale string) (review.Proposal, error) {
	var proposal review.Proposal
	err := p.withTx(func(tx *sql.Tx) error {
		current, err := selectProposalTx(tx, proposalID, true)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		current.Status = review.ProposalCanceled
		current.ActiveSlotConsuming = false
		current.NextRetryAction = ""
		current.LineStoppedBy = requestedBy
		current.LineStopReason = rationale
		current.LineStoppedAt = &now
		if err := updateProposalOperationalStateTx(tx, current); err != nil {
			return err
		}
		if current.CurrentAttemptID != "" {
			attempt, err := selectChangeAttemptTx(tx, current.CurrentAttemptID, true)
			if err == nil && !isTerminalAttemptState(attempt.State) {
				attempt.State = improvement.AttemptStateAbandoned
				attempt.FailureClass = firstNonEmpty(attempt.FailureClass, "stopped_by_operator")
				attempt.FailureSummary = firstNonEmpty(attempt.FailureSummary, rationale)
				attempt.RetryDecision = "stop_line"
				attempt.UpdatedAt = now
				temp := newSubsetStore()
				temp.changeAttempts[attempt.ID] = attempt
				if err := persistChangeAttempts(tx, temp); err != nil {
					return err
				}
			}
		}
		if _, err := tx.Exec(`update improvement_candidate set status = $2, line_status = $3, auto_retry_budget_remaining = 0, updated_at = $4 where candidate_key = $1`,
			current.CandidateKey,
			string(improvement.CandidateDormant),
			string(improvement.LineClosed),
			now,
		); err != nil {
			return err
		}
		temp := newSubsetStore()
		temp.proposalMemory = append(temp.proposalMemory, review.ProposalMemory{
			ID:                nextID("memory", 0),
			ProposalID:        current.ID,
			CandidateKey:      current.CandidateKey,
			ConversationID:    current.ConversationID,
			CaseID:            current.CaseID,
			OriginTraceID:     current.OriginTraceID,
			EvidenceTraceIDs:  append([]string(nil), current.EvidenceTraceIDs...),
			Hypothesis:        current.Summary,
			DiffSummary:       current.ProposedScope,
			ReviewRationale:   firstNonEmpty(rationale, "Line stopped by operator."),
			Disposition:       review.ProposalCanceled,
			DispositionReason: firstNonEmpty(rationale, "Line stopped by operator."),
			FailureClass:      "stopped_by_operator",
			LinkedProposalIDs: append([]string(nil), current.PriorSimilarProposalIDs...),
			CreatedAt:         now,
		})
		if err := persistProposalMemory(tx, temp); err != nil {
			return err
		}
		proposal, err = selectProposalTx(tx, proposalID, false)
		return err
	})
	if err != nil {
		return review.Proposal{}, err
	}
	return proposal, nil
}

func selectChangeAttemptTx(tx *sql.Tx, attemptID string, forUpdate bool) (improvement.ChangeAttempt, error) {
	query := `select id, version, proposal_id, candidate_key, attempt_number, target_layer, target_kind, target_ref, trigger, state, attempt_trace_id, parent_attempt_id, branch_name, pr_url, head_sha, failure_class, failure_summary, retry_decision, retry_after, material_hypothesis_change, diff_summary, changed_files, validation_summary, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta, overlay_payload, created_at, updated_at from change_attempt where id = $1`
	if forUpdate {
		query += ` for update`
	}
	var item improvement.ChangeAttempt
	var targetLayer, trigger, state string
	var retryAfter sql.NullTime
	var targetKind, targetRef, attemptTraceID, parentAttemptID, branchName, prURL, headSHA, failureClass, failureSummary, retryDecision, diffSummary, validationSummary, changePlan, repoPatch, validationPlan, retryAssessment, hypothesisDelta sql.NullString
	var changedFiles, overlayPayload []byte
	if err := tx.QueryRow(query, strings.TrimSpace(attemptID)).
		Scan(&item.ID, &item.Version, &item.ProposalID, &item.CandidateKey, &item.AttemptNumber, &targetLayer, &targetKind, &targetRef, &trigger, &state, &attemptTraceID, &parentAttemptID, &branchName, &prURL, &headSHA, &failureClass, &failureSummary, &retryDecision, &retryAfter, &item.MaterialHypothesisChange, &diffSummary, &changedFiles, &validationSummary, &changePlan, &repoPatch, &validationPlan, &retryAssessment, &hypothesisDelta, &overlayPayload, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return improvement.ChangeAttempt{}, err
	}
	item.TargetLayer = harness.TargetLayer(targetLayer)
	item.TargetKind = targetKind.String
	item.TargetRef = targetRef.String
	item.Trigger = improvement.ChangeAttemptTrigger(trigger)
	item.State = improvement.ChangeAttemptState(state)
	item.AttemptTraceID = attemptTraceID.String
	item.ParentAttemptID = parentAttemptID.String
	item.BranchName = branchName.String
	item.PRURL = prURL.String
	item.HeadSHA = headSHA.String
	item.FailureClass = failureClass.String
	item.FailureSummary = failureSummary.String
	item.RetryDecision = retryDecision.String
	if retryAfter.Valid {
		t := retryAfter.Time
		item.RetryAfter = &t
	}
	item.DiffSummary = diffSummary.String
	item.ChangedFiles = decodeJSON(changedFiles, []string{})
	item.ValidationSummary = validationSummary.String
	item.ChangePlan = changePlan.String
	item.RepoPatch = repoPatch.String
	item.ValidationPlan = validationPlan.String
	item.RetryAssessment = retryAssessment.String
	item.HypothesisDelta = hypothesisDelta.String
	item.OverlayPayload = decodeJSON(overlayPayload, map[string]any{})
	return normalizeChangeAttempt(item), nil
}

func selectLatestChangeAttemptByProposalTx(tx *sql.Tx, proposalID string, forUpdate bool) (improvement.ChangeAttempt, bool, error) {
	query := `select id from change_attempt where proposal_id = $1 order by attempt_number desc, created_at desc limit 1`
	if forUpdate {
		query += ` for update`
	}
	var attemptID string
	if err := tx.QueryRow(query, strings.TrimSpace(proposalID)).Scan(&attemptID); err != nil {
		if err == sql.ErrNoRows {
			return improvement.ChangeAttempt{}, false, nil
		}
		return improvement.ChangeAttempt{}, false, err
	}
	item, err := selectChangeAttemptTx(tx, attemptID, false)
	if err != nil {
		return improvement.ChangeAttempt{}, false, err
	}
	return item, true, nil
}
