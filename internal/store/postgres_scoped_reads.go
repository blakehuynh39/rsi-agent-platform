package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

const (
	caseColumns           = `id, conversation_id, kind, intent, title, summary, status, approval_mode, response_mode, assigned_bot, opened_by_event_id, closed_by_event_id, latest_trace_id, resolution_state, resolved_at, latest_outcome_id, outcome_score, superseded_by_case_id, created_at, updated_at, closed_at`
	workflowSelectColumns = `id, version, ingestion_id, trace_id, conversation_id, case_id, thread_key, kind, intent, assigned_bot, approval_mode, response_mode, status, last_verdict, last_error, attempt_number, parent_workflow_id, failure_class, failure_summary, retry_decision, retry_after, runner_diagnostics, repair_attempted, repair_succeeded, created_at, updated_at, completed_at`
	actionIntentColumns   = `id, operation_id, owner_plane, conversation_id, case_id, trace_id, proposal_id, attempt_id, kind, phase_key, target_ref, request_payload, idempotency_key, approval_mode, approval_state, policy_verdict, status, superseded_by_action_id, requested_by, rationale, evidence_refs, created_at, updated_at`
	outcomeColumns        = `id, operation_id, source, source_event_id, conversation_id, case_id, trace_id, proposal_id, attempt_id, outcome_type, verdict, score, summary, details, external_ref, recorded_by, recorded_at`
	proposalColumns       = `id, version, trace_id, conversation_id, case_id, origin_trace_id, evidence_trace_ids, title, category, summary, status, reviewer, candidate_key, target_layer, target_kind, target_ref, source_eval_ids, risk_tier, proposed_scope, evidence_artifact_ids, active_slot_consuming, review_deadline, prior_similar_proposal_ids, new_evidence_since_last_rejection, current_attempt_id, attempt_count, auto_retry_budget_remaining, last_failure_class, next_retry_action, line_stopped_by, line_stop_reason, line_stopped_at, recommended_intervention_kind, recommended_intervention_rationale, target_surface, touched_files, validation_plan, material_risk_summary, recommended_disposition, created_at`
	candidateColumns      = `id, candidate_key, conversation_id, case_id, origin_trace_id, evidence_trace_ids, subsystem, failure_mode, intervention_type, target_layer, target_kind, target_ref, status, severity, recurrence_count, expected_impact, novelty_score, confidence_score, freshness_score, priority_score, risk_tier, hypothesis, proposed_scope, latest_trace_id, source_eval_ids, evidence_artifact_ids, prior_similar_proposal_ids, new_evidence_since_last_rejection, line_status, retryable_failure_class, last_attempt_id, attempt_count, auto_retry_budget_remaining, current_target_layer, last_evaluated_at, created_at, updated_at`
)

func (p *PostgresStore) ListTracesByConversation(conversationID string) []events.TraceSummary {
	return p.listTraceSummariesWhere(`conversation_id = $1 order by started_at desc`, conversationID)
}

func (p *PostgresStore) ListTracesByCase(caseID string) []events.TraceSummary {
	return p.listTraceSummariesWhere(`case_id = $1 order by started_at desc`, caseID)
}

func (p *PostgresStore) listTraceSummariesWhere(where string, args ...any) []events.TraceSummary {
	rows, err := p.db.Query(`select trace_id, ingestion_id, workflow_id, conversation_id, case_id, trigger_event_id, supersedes_trace_id, thread_key, workflow_kind, status, last_verdict, started_at, ended_at, event_count, artifact_count, reasoning_step_count, tool_call_count, slack_action_count from trace_summary where `+where, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []events.TraceSummary{}
	for rows.Next() {
		summary, err := scanTraceSummary(rows)
		if err != nil {
			return nil
		}
		out = append(out, summary)
	}
	return out
}

func (p *PostgresStore) ListCasesByConversation(conversationID string) []conversation.Case {
	rows, err := p.db.Query(`select `+caseColumns+` from case_record where conversation_id = $1 order by updated_at desc`, conversationID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []conversation.Case{}
	for rows.Next() {
		item, err := scanCaseRecord(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func scanCaseRecord(scanner interface{ Scan(dest ...any) error }) (conversation.Case, error) {
	var item conversation.Case
	var status string
	var approvalMode, responseMode, openedByEventID, closedByEventID, latestTraceID, resolutionState, latestOutcomeID, supersededByCaseID sql.NullString
	var resolvedAt, closedAt sql.NullTime
	if err := scanner.Scan(&item.ID, &item.ConversationID, &item.Kind, &item.Intent, &item.Title, &item.Summary, &status, &approvalMode, &responseMode, &item.AssignedBot, &openedByEventID, &closedByEventID, &latestTraceID, &resolutionState, &resolvedAt, &latestOutcomeID, &item.OutcomeScore, &supersededByCaseID, &item.CreatedAt, &item.UpdatedAt, &closedAt); err != nil {
		return conversation.Case{}, err
	}
	item.Status = conversation.CaseStatus(status)
	item.ApprovalMode = approvalMode.String
	item.ResponseMode = responseMode.String
	item.OpenedByEventID = openedByEventID.String
	item.ClosedByEventID = closedByEventID.String
	item.LatestTraceID = latestTraceID.String
	item.ResolutionState = conversation.ResolutionState(resolutionState.String)
	item.LatestOutcomeID = latestOutcomeID.String
	item.SupersededByCaseID = supersededByCaseID.String
	if resolvedAt.Valid {
		t := resolvedAt.Time
		item.ResolvedAt = &t
	}
	if closedAt.Valid {
		t := closedAt.Time
		item.ClosedAt = &t
	}
	return item, nil
}

func (p *PostgresStore) ListWorkflowsByConversation(conversationID string) []Workflow {
	return p.listWorkflowsWhere(`conversation_id = $1 order by created_at desc`, conversationID)
}

func (p *PostgresStore) ListWorkflowsByCase(caseID string) []Workflow {
	return p.listWorkflowsWhere(`case_id = $1 order by created_at desc`, caseID)
}

func (p *PostgresStore) listWorkflowsWhere(where string, args ...any) []Workflow {
	rows, err := p.db.Query(`select `+workflowSelectColumns+` from workflow where `+where, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []Workflow{}
	for rows.Next() {
		item, err := scanWorkflowRecord(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func scanWorkflowRecord(scanner interface{ Scan(dest ...any) error }) (Workflow, error) {
	var item Workflow
	var ingestionID, traceID, conversationID, caseID, intent, approvalMode, responseMode, lastVerdict, lastError, parentWorkflowID, failureClass, failureSummary, retryDecision sql.NullString
	var retryAfter, completedAt sql.NullTime
	var runnerDiagnostics []byte
	if err := scanner.Scan(&item.ID, &item.Version, &ingestionID, &traceID, &conversationID, &caseID, &item.ThreadKey, &item.Kind, &intent, &item.AssignedBot, &approvalMode, &responseMode, &item.Status, &lastVerdict, &lastError, &item.AttemptNumber, &parentWorkflowID, &failureClass, &failureSummary, &retryDecision, &retryAfter, &runnerDiagnostics, &item.RepairAttempted, &item.RepairSucceeded, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
		return Workflow{}, err
	}
	item.IngestionID = ingestionID.String
	item.TraceID = traceID.String
	item.ConversationID = conversationID.String
	item.CaseID = caseID.String
	item.Intent = intent.String
	item.ApprovalMode = approvalMode.String
	item.ResponseMode = responseMode.String
	item.LastVerdict = lastVerdict.String
	item.LastError = lastError.String
	item.ParentWorkflowID = parentWorkflowID.String
	item.FailureClass = failureClass.String
	item.FailureSummary = failureSummary.String
	item.RetryDecision = retryDecision.String
	if retryAfter.Valid {
		t := retryAfter.Time
		item.RetryAfter = &t
	}
	item.RunnerDiagnostics = decodeJSON(runnerDiagnostics, map[string]any{})
	if completedAt.Valid {
		t := completedAt.Time
		item.CompletedAt = &t
	}
	return item, nil
}

func (p *PostgresStore) ListEvalRunsByTrace(traceID string) []evals.Run {
	return p.ListEvalRunsByTraceIDs([]string{traceID})
}

func (p *PostgresStore) ListEvalRunsByTraceIDs(traceIDs []string) []evals.Run {
	traceIDs = compactStrings(traceIDs)
	if len(traceIDs) == 0 {
		return nil
	}
	query := `select id, trace_id, event_id, suite_name, status, trigger, overall_score, overall_verdict, created_at, completed_at from eval_run where trace_id in (` + sqlPlaceholders(len(traceIDs), 1) + `) order by created_at desc`
	rows, err := p.db.Query(query, stringsToAny(traceIDs)...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []evals.Run{}
	for rows.Next() {
		item, err := scanEvalRunRecord(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func scanEvalRunRecord(scanner interface{ Scan(dest ...any) error }) (evals.Run, error) {
	var item evals.Run
	var eventID sql.NullString
	var status string
	var completedAt sql.NullTime
	if err := scanner.Scan(&item.ID, &item.TraceID, &eventID, &item.SuiteName, &status, &item.Trigger, &item.OverallScore, &item.OverallVerdict, &item.CreatedAt, &completedAt); err != nil {
		return evals.Run{}, err
	}
	item.EventID = eventID.String
	item.Status = evals.Status(status)
	if completedAt.Valid {
		item.CompletedAt = completedAt.Time
	}
	return item, nil
}

func (p *PostgresStore) ListProposalsByConversation(conversationID string) []review.Proposal {
	return p.listProposalsWhere(`conversation_id = $1 order by created_at desc`, conversationID)
}

func (p *PostgresStore) ListProposalsByCase(caseID string) []review.Proposal {
	return p.listProposalsWhere(`case_id = $1 order by created_at desc`, caseID)
}

func (p *PostgresStore) ListProposalsByTrace(traceID string) []review.Proposal {
	return p.listProposalsWhere(`trace_id = $1 or origin_trace_id = $1 or evidence_trace_ids ? $1 order by created_at desc`, traceID)
}

func (p *PostgresStore) listProposalsWhere(where string, args ...any) []review.Proposal {
	rows, err := p.db.Query(`select `+proposalColumns+` from proposal where `+where, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []review.Proposal{}
	for rows.Next() {
		item, err := scanProposalRecord(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func scanProposalRecord(scanner interface{ Scan(dest ...any) error }) (review.Proposal, error) {
	var item review.Proposal
	var status, targetLayer string
	var conversationID, caseID, originTraceID, reviewer, targetKind, targetRef, currentAttemptID, lastFailureClass, nextRetryAction, lineStoppedBy, lineStopReason, recommendedKind, recommendedRationale, targetSurface, validationPlan, materialRiskSummary, recommendedDisposition sql.NullString
	var evidenceTraceIDs, sourceEvalIDs, evidenceArtifactIDs, priorSimilarProposalIDs, touchedFiles []byte
	var reviewDeadline, lineStoppedAt sql.NullTime
	if err := scanner.Scan(&item.ID, &item.Version, &item.TraceID, &conversationID, &caseID, &originTraceID, &evidenceTraceIDs, &item.Title, &item.Category, &item.Summary, &status, &reviewer, &item.CandidateKey, &targetLayer, &targetKind, &targetRef, &sourceEvalIDs, &item.RiskTier, &item.ProposedScope, &evidenceArtifactIDs, &item.ActiveSlotConsuming, &reviewDeadline, &priorSimilarProposalIDs, &item.NewEvidenceSinceLastRejection, &currentAttemptID, &item.AttemptCount, &item.AutoRetryBudgetRemaining, &lastFailureClass, &nextRetryAction, &lineStoppedBy, &lineStopReason, &lineStoppedAt, &recommendedKind, &recommendedRationale, &targetSurface, &touchedFiles, &validationPlan, &materialRiskSummary, &recommendedDisposition, &item.CreatedAt); err != nil {
		return review.Proposal{}, err
	}
	item.ConversationID = conversationID.String
	item.CaseID = caseID.String
	item.OriginTraceID = originTraceID.String
	item.EvidenceTraceIDs = decodeJSON(evidenceTraceIDs, []string{})
	item.Status = review.ProposalStatus(status)
	item.Reviewer = reviewer.String
	item.TargetLayer = harness.TargetLayer(targetLayer)
	item.TargetKind = targetKind.String
	item.TargetRef = targetRef.String
	item.SourceEvalIDs = decodeJSON(sourceEvalIDs, []string{})
	item.EvidenceArtifactIDs = decodeJSON(evidenceArtifactIDs, []string{})
	item.PriorSimilarProposalIDs = decodeJSON(priorSimilarProposalIDs, []string{})
	item.CurrentAttemptID = currentAttemptID.String
	item.LastFailureClass = lastFailureClass.String
	item.NextRetryAction = nextRetryAction.String
	item.LineStoppedBy = lineStoppedBy.String
	item.LineStopReason = lineStopReason.String
	item.RecommendedInterventionKind = review.ProposalInterventionKind(recommendedKind.String)
	item.RecommendedInterventionRationale = recommendedRationale.String
	item.TargetSurface = targetSurface.String
	item.TouchedFiles = decodeJSON(touchedFiles, []string{})
	item.ValidationPlan = validationPlan.String
	item.MaterialRiskSummary = materialRiskSummary.String
	item.RecommendedDisposition = recommendedDisposition.String
	if reviewDeadline.Valid {
		item.ReviewDeadline = reviewDeadline.Time
	}
	if lineStoppedAt.Valid {
		t := lineStoppedAt.Time
		item.LineStoppedAt = &t
	}
	return normalizeProposalTargetFields(item), nil
}

func (p *PostgresStore) ListActionIntentsByConversation(conversationID string) []action.Intent {
	return p.listActionIntentsWhere(`conversation_id = $1 order by created_at desc`, conversationID)
}

func (p *PostgresStore) ListActionIntentsByCase(caseID string) []action.Intent {
	return p.listActionIntentsWhere(`case_id = $1 order by created_at desc`, caseID)
}

func (p *PostgresStore) ListActionIntentsByTrace(traceID string) []action.Intent {
	return p.listActionIntentsWhere(`trace_id = $1 order by created_at desc`, traceID)
}

func (p *PostgresStore) ListActionIntentsByProposal(proposalID string) []action.Intent {
	return p.listActionIntentsWhere(`proposal_id = $1 order by created_at desc`, proposalID)
}

func (p *PostgresStore) listActionIntentsWhere(where string, args ...any) []action.Intent {
	rows, err := p.db.Query(`select `+actionIntentColumns+` from action_intent where `+where, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []action.Intent{}
	for rows.Next() {
		item, err := scanActionIntentRecord(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func scanActionIntentRecord(scanner interface{ Scan(dest ...any) error }) (action.Intent, error) {
	var item action.Intent
	var operationID sql.NullString
	var conversationID, caseID, traceID, proposalID, attemptID, phaseKey, targetRef, idempotencyKey, approvalMode, approvalState, policyVerdict, supersededBy, requestedBy, rationale sql.NullString
	var requestPayload, evidenceRefs []byte
	var kind, status string
	if err := scanner.Scan(&item.ID, &operationID, &item.OwnerPlane, &conversationID, &caseID, &traceID, &proposalID, &attemptID, &kind, &phaseKey, &targetRef, &requestPayload, &idempotencyKey, &approvalMode, &approvalState, &policyVerdict, &status, &supersededBy, &requestedBy, &rationale, &evidenceRefs, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return action.Intent{}, err
	}
	item.OperationID = operationID.String
	item.ConversationID = conversationID.String
	item.CaseID = caseID.String
	item.TraceID = traceID.String
	item.ProposalID = proposalID.String
	item.AttemptID = attemptID.String
	item.Kind = action.Kind(kind)
	item.PhaseKey = phaseKey.String
	item.TargetRef = targetRef.String
	item.RequestPayload = decodeJSON(requestPayload, map[string]any{})
	item.IdempotencyKey = idempotencyKey.String
	item.ApprovalMode = approvalMode.String
	item.ApprovalState = approvalState.String
	item.PolicyVerdict = policyVerdict.String
	item.Status = action.Status(status)
	item.SupersededByActionID = supersededBy.String
	item.RequestedBy = requestedBy.String
	item.Rationale = rationale.String
	item.EvidenceRefs = decodeJSON(evidenceRefs, []events.EvidenceRef{})
	return item, nil
}

func (p *PostgresStore) ListOutcomesByConversation(conversationID string) []outcome.Record {
	return p.listOutcomesWhere(`conversation_id = $1 order by recorded_at desc`, conversationID)
}

func (p *PostgresStore) ListOutcomesByCase(caseID string) []outcome.Record {
	return p.listOutcomesWhere(`case_id = $1 order by recorded_at desc`, caseID)
}

func (p *PostgresStore) ListOutcomesByTrace(traceID string) []outcome.Record {
	return p.listOutcomesWhere(`trace_id = $1 order by recorded_at desc`, traceID)
}

func (p *PostgresStore) ListOutcomesByProposal(proposalID string) []outcome.Record {
	return p.listOutcomesWhere(`proposal_id = $1 order by recorded_at desc`, proposalID)
}

func (p *PostgresStore) listOutcomesWhere(where string, args ...any) []outcome.Record {
	rows, err := p.db.Query(`select `+outcomeColumns+` from outcome_record where `+where, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []outcome.Record{}
	for rows.Next() {
		item, err := scanOutcomeRecord(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func scanOutcomeRecord(scanner interface{ Scan(dest ...any) error }) (outcome.Record, error) {
	var item outcome.Record
	var operationID, sourceEventID, conversationID, caseID, traceID, proposalID, attemptID, summary, details, externalRef, recordedBy sql.NullString
	var outcomeType, verdict string
	if err := scanner.Scan(&item.ID, &operationID, &item.Source, &sourceEventID, &conversationID, &caseID, &traceID, &proposalID, &attemptID, &outcomeType, &verdict, &item.Score, &summary, &details, &externalRef, &recordedBy, &item.RecordedAt); err != nil {
		return outcome.Record{}, err
	}
	item.OperationID = operationID.String
	item.SourceEventID = sourceEventID.String
	item.ConversationID = conversationID.String
	item.CaseID = caseID.String
	item.TraceID = traceID.String
	item.ProposalID = proposalID.String
	item.AttemptID = attemptID.String
	item.OutcomeType = outcome.Type(outcomeType)
	item.Verdict = outcome.Verdict(verdict)
	item.Summary = summary.String
	item.Details = details.String
	item.ExternalRef = externalRef.String
	item.RecordedBy = recordedBy.String
	return item, nil
}

func (p *PostgresStore) ListHarnessExecutionsByTraceIDs(traceIDs []string) []harness.Execution {
	traceIDs = compactStrings(traceIDs)
	if len(traceIDs) == 0 {
		return nil
	}
	query := `select id, operation_id, trace_id, proposal_id, role, session_scope_kind, session_scope_id, hermes_session_id, parent_session_id, harness_profile_id, effective_overlay_id, effective_overlay_version, memory_backend, memory_reads, memory_writes, created_at from harness_execution where trace_id in (` + sqlPlaceholders(len(traceIDs), 1) + `) order by created_at desc, id asc`
	rows, err := p.db.Query(query, stringsToAny(traceIDs)...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []harness.Execution{}
	for rows.Next() {
		item, err := scanHarnessExecution(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) ListHarnessExecutionObservationsByTraceIDs(traceIDs []string) []harness.ExecutionObservation {
	traceIDs = compactStrings(traceIDs)
	if len(traceIDs) == 0 {
		return nil
	}
	query := `select id, execution_id, operation_id, trace_id, workflow_id, hermes_session_id, role, phase, event_type, status, seq, payload, recorded_at from harness_execution_observation where trace_id in (` + sqlPlaceholders(len(traceIDs), 1) + `) order by recorded_at desc, execution_id asc, seq asc, id asc`
	rows, err := p.db.Query(query, stringsToAny(traceIDs)...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []harness.ExecutionObservation{}
	for rows.Next() {
		item, err := scanHarnessExecutionObservation(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) LatestCandidateForTrace(traceID string) (improvement.Candidate, bool) {
	row := p.db.QueryRow(`select `+candidateColumns+` from improvement_candidate where latest_trace_id = $1 or evidence_trace_ids ? $1 order by updated_at desc limit 1`, traceID)
	item, err := scanCandidateRecord(row)
	if err != nil {
		return improvement.Candidate{}, false
	}
	return item, true
}

func scanCandidateRecord(scanner interface{ Scan(dest ...any) error }) (improvement.Candidate, error) {
	var item improvement.Candidate
	var status, riskTier, targetLayer, lineStatus, currentTargetLayer string
	var conversationID, caseID, originTraceID, latestTraceID, targetKind, targetRef, retryableFailureClass, lastAttemptID sql.NullString
	var evidenceTraceIDs, sourceEvalIDs, evidenceArtifactIDs, priorSimilarProposalIDs []byte
	var lastEvaluatedAt sql.NullTime
	if err := scanner.Scan(&item.ID, &item.CandidateKey, &conversationID, &caseID, &originTraceID, &evidenceTraceIDs, &item.Subsystem, &item.FailureMode, &item.InterventionType, &targetLayer, &targetKind, &targetRef, &status, &item.Severity, &item.RecurrenceCount, &item.ExpectedImpact, &item.NoveltyScore, &item.ConfidenceScore, &item.FreshnessScore, &item.PriorityScore, &riskTier, &item.Hypothesis, &item.ProposedScope, &latestTraceID, &sourceEvalIDs, &evidenceArtifactIDs, &priorSimilarProposalIDs, &item.NewEvidenceSinceLastRejection, &lineStatus, &retryableFailureClass, &lastAttemptID, &item.AttemptCount, &item.AutoRetryBudgetRemaining, &currentTargetLayer, &lastEvaluatedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return improvement.Candidate{}, err
	}
	item.ConversationID = conversationID.String
	item.CaseID = caseID.String
	item.OriginTraceID = originTraceID.String
	item.EvidenceTraceIDs = decodeJSON(evidenceTraceIDs, []string{})
	item.TargetLayer = harness.TargetLayer(targetLayer)
	item.TargetKind = targetKind.String
	item.TargetRef = targetRef.String
	item.Status = improvement.CandidateStatus(status)
	item.RiskTier = improvement.RiskTier(riskTier)
	item.LatestTraceID = latestTraceID.String
	item.SourceEvalIDs = decodeJSON(sourceEvalIDs, []string{})
	item.EvidenceArtifactIDs = decodeJSON(evidenceArtifactIDs, []string{})
	item.PriorSimilarProposalIDs = decodeJSON(priorSimilarProposalIDs, []string{})
	item.LineStatus = improvement.LineStatus(lineStatus)
	item.RetryableFailureClass = retryableFailureClass.String
	item.LastAttemptID = lastAttemptID.String
	item.CurrentTargetLayer = harness.TargetLayer(currentTargetLayer)
	if lastEvaluatedAt.Valid {
		item.LastEvaluatedAt = lastEvaluatedAt.Time
	}
	return normalizeCandidateTargetFields(item), nil
}

func stringsToAny(items []string) []any {
	out := make([]any, len(items))
	for i, item := range items {
		out[i] = item
	}
	return out
}

func sqlPlaceholders(count int, start int) string {
	parts := make([]string, count)
	for i := range parts {
		parts[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(parts, ",")
}
