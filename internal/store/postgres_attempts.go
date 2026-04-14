package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
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

func (p *PostgresStore) UpsertChangeAttempt(item improvement.ChangeAttempt) (improvement.ChangeAttempt, error) {
	item = normalizeChangeAttempt(item)
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextID("attempt", 0)
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	err := p.withTx(func(tx *sql.Tx) error {
		temp := newSubsetStore()
		temp.changeAttempts[item.ID] = item
		if err := persistChangeAttempts(tx, temp); err != nil {
			return err
		}
		if strings.TrimSpace(item.ProposalID) != "" {
			proposal, err := selectProposalTx(tx, item.ProposalID, true)
			if err != nil {
				return err
			}
			budget := defaultProposalRetryBudget - item.AttemptNumber
			if budget < 0 {
				budget = 0
			}
			proposal.CurrentAttemptID = item.ID
			proposal.AttemptCount = maxInt(proposal.AttemptCount, item.AttemptNumber)
			proposal.AutoRetryBudgetRemaining = budget
			proposal.LastFailureClass = item.FailureClass
			proposal.NextRetryAction = item.RetryDecision
			if err := updateProposalOperationalStateTx(tx, proposal); err != nil {
				return err
			}
		}
		if strings.TrimSpace(item.CandidateKey) != "" {
			budget := defaultProposalRetryBudget - item.AttemptNumber
			if budget < 0 {
				budget = 0
			}
			if _, err := tx.Exec(`update improvement_candidate set line_status = $2, retryable_failure_class = $3, last_attempt_id = $4, attempt_count = greatest(coalesce(attempt_count, 0), $5), auto_retry_budget_remaining = $6, current_target_layer = $7, updated_at = $8 where candidate_key = $1`,
				item.CandidateKey,
				string(improvement.LineActive),
				firstNonEmpty(item.FailureClass),
				firstNonEmpty(item.ID),
				item.AttemptNumber,
				budget,
				firstNonEmpty(string(item.TargetLayer), string(harness.TargetLayerRepoChange)),
				item.UpdatedAt,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return improvement.ChangeAttempt{}, err
	}
	return item, nil
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

func (p *PostgresStore) CreateDerivedTrace(req DerivedTraceRequest) (events.Trace, Workflow, error) {
	var (
		trace    events.Trace
		workflow Workflow
	)
	err := p.withTx(func(tx *sql.Tx) error {
		now := req.CreatedAt.UTC()
		if now.IsZero() {
			now = time.Now().UTC()
		}
		var sourceIngestionID, sourceConversationID, sourceCaseID, sourceTriggerEventID, sourceThreadKey, sourceWorkflowKind sql.NullString
		if strings.TrimSpace(req.SourceTraceID) != "" {
			if err := tx.QueryRow(`select ingestion_id, conversation_id, case_id, trigger_event_id, thread_key, workflow_kind from trace_summary where trace_id = $1`, req.SourceTraceID).
				Scan(&sourceIngestionID, &sourceConversationID, &sourceCaseID, &sourceTriggerEventID, &sourceThreadKey, &sourceWorkflowKind); err != nil && err != sql.ErrNoRows {
				return err
			}
		}
		traceID := nextID("trace", 0)
		workflowID := nextID("wf", 0)
		workflow = Workflow{
			ID:             workflowID,
			IngestionID:    firstNonEmpty(req.IngestionID, sourceIngestionID.String, "derived:"+traceID),
			TraceID:        traceID,
			ConversationID: firstNonEmpty(req.ConversationID, sourceConversationID.String),
			CaseID:         firstNonEmpty(req.CaseID, sourceCaseID.String),
			ThreadKey:      firstNonEmpty(req.ThreadKey, sourceThreadKey.String, "proposal:"+req.ProposalID),
			Kind:           firstNonEmpty(req.WorkflowKind, sourceWorkflowKind.String, "proposal_attempt"),
			Intent:         firstNonEmpty(req.WorkflowKind, sourceWorkflowKind.String, "proposal_attempt"),
			AssignedBot:    "proposal",
			ApprovalMode:   "human_review",
			ResponseMode:   "analysis",
			Status:         "running",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		trace = events.Trace{
			Summary: events.TraceSummary{
				TraceID:           traceID,
				IngestionID:       workflow.IngestionID,
				WorkflowID:        workflowID,
				ConversationID:    workflow.ConversationID,
				CaseID:            workflow.CaseID,
				TriggerEventID:    firstNonEmpty(req.TriggerEventID, sourceTriggerEventID.String),
				SupersedesTraceID: strings.TrimSpace(req.SourceTraceID),
				ThreadKey:         workflow.ThreadKey,
				WorkflowKind:      workflow.Kind,
				Status:            events.StatusQueued,
				StartedAt:         now,
				EndedAt:           now,
			},
			Events: []events.TraceEvent{
				{
					TraceID:        traceID,
					IngestionID:    workflow.IngestionID,
					WorkflowID:     workflowID,
					ConversationID: workflow.ConversationID,
					CaseID:         workflow.CaseID,
					TriggerEventID: firstNonEmpty(req.TriggerEventID, sourceTriggerEventID.String),
					Plane:          "improvement",
					Service:        "improvement-plane",
					Actor:          "attempt-supervisor",
					EventType:      "change_attempt.queued",
					Status:         events.StatusQueued,
					StartedAt:      now,
					Description:    firstNonEmpty(req.Description, fmt.Sprintf("Queued remediation attempt %s for proposal %s.", req.AttemptID, req.ProposalID)),
				},
			},
			Reasoning: []events.ReasoningStep{
				{
					ID:             nextID("reason", 0),
					TraceID:        traceID,
					WorkflowID:     workflowID,
					ConversationID: workflow.ConversationID,
					CaseID:         workflow.CaseID,
					StepType:       "attempt_bootstrap",
					Summary:        firstNonEmpty(req.Description, fmt.Sprintf("Start remediation attempt %s under proposal %s.", req.AttemptID, req.ProposalID)),
					Confidence:     0.9,
					Decision:       req.AttemptID,
					CreatedAt:      now,
				},
			},
		}
		recomputeTraceSummary(&trace)
		temp := newSubsetStore()
		temp.upsertWorkflowLocked(workflow)
		temp.traces[traceID] = trace
		if err := persistWorkflows(tx, temp); err != nil {
			return err
		}
		return persistTraces(tx, temp)
	})
	if err != nil {
		return events.Trace{}, Workflow{}, err
	}
	return trace, workflow, nil
}

func selectChangeAttemptTx(tx *sql.Tx, attemptID string, forUpdate bool) (improvement.ChangeAttempt, error) {
	query := `select id, proposal_id, candidate_key, attempt_number, target_layer, target_kind, target_ref, trigger, state, attempt_trace_id, parent_attempt_id, branch_name, pr_url, head_sha, failure_class, failure_summary, retry_decision, retry_after, material_hypothesis_change, diff_summary, changed_files, validation_summary, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta, overlay_payload, created_at, updated_at from change_attempt where id = $1`
	if forUpdate {
		query += ` for update`
	}
	var item improvement.ChangeAttempt
	var targetLayer, trigger, state string
	var retryAfter sql.NullTime
	var targetKind, targetRef, attemptTraceID, parentAttemptID, branchName, prURL, headSHA, failureClass, failureSummary, retryDecision, diffSummary, validationSummary, changePlan, repoPatch, validationPlan, retryAssessment, hypothesisDelta sql.NullString
	var changedFiles, overlayPayload []byte
	if err := tx.QueryRow(query, strings.TrimSpace(attemptID)).
		Scan(&item.ID, &item.ProposalID, &item.CandidateKey, &item.AttemptNumber, &targetLayer, &targetKind, &targetRef, &trigger, &state, &attemptTraceID, &parentAttemptID, &branchName, &prURL, &headSHA, &failureClass, &failureSummary, &retryDecision, &retryAfter, &item.MaterialHypothesisChange, &diffSummary, &changedFiles, &validationSummary, &changePlan, &repoPatch, &validationPlan, &retryAssessment, &hypothesisDelta, &overlayPayload, &item.CreatedAt, &item.UpdatedAt); err != nil {
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
