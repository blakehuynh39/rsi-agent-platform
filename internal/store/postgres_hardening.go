package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
)

func proposalDecisionIdempotencyKey(proposalID string, decision string, scope review.ProposalFeedbackScope) string {
	return fmt.Sprintf("%s:%s:%s", strings.TrimSpace(proposalID), strings.ToLower(strings.TrimSpace(decision)), strings.ToLower(firstNonEmpty(string(scope), string(review.FeedbackScopeLine))))
}

func (p *PostgresStore) upsertRepoChangeJobDirect(job improvement.RepoChangeJob) (item improvement.RepoChangeJob, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		if job.ID == "" {
			job.ID = nextID("job", 0)
		}
		if job.CreatedAt.IsZero() {
			job.CreatedAt = now
		}
		if job.UpdatedAt.IsZero() {
			job.UpdatedAt = now
		}
		temp := newSubsetStore()
		temp.repoChangeJobs[job.ID] = job
		if err := persistRepoChangeJobs(tx, temp); err != nil {
			return err
		}
		item = job
		return nil
	})
	return
}

func (p *PostgresStore) updateRepoChangeJobStatusDirect(jobID string, status string) (item improvement.RepoChangeJob, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(`update repo_change_job set status = $2, updated_at = $3 where id = $1 returning id, proposal_id, attempt_id, conversation_id, case_id, origin_trace_id, candidate_key, status, repo, base_ref, branch_name, allowed_path_globs, context_summary, sandbox_namespace, sandbox_job_name, sandbox_pod_name, validation_error, validation_ref, log_artifact_id, created_at, updated_at`,
			jobID,
			status,
			time.Now().UTC(),
		)
		var attemptID, conversationID, caseID, originTraceID, sandboxNamespace, sandboxJobName, sandboxPodName, validationError, validationRef, logArtifactID sql.NullString
		var allowed []byte
		if err := row.Scan(&item.ID, &item.ProposalID, &attemptID, &conversationID, &caseID, &originTraceID, &item.CandidateKey, &item.Status, &item.Repo, &item.BaseRef, &item.BranchName, &allowed, &item.ContextSummary, &sandboxNamespace, &sandboxJobName, &sandboxPodName, &validationError, &validationRef, &logArtifactID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		item.AttemptID = attemptID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.OriginTraceID = originTraceID.String
		item.AllowedPathGlobs = decodeJSON(allowed, []string{})
		item.SandboxNamespace = sandboxNamespace.String
		item.SandboxJobName = sandboxJobName.String
		item.SandboxPodName = sandboxPodName.String
		item.ValidationError = validationError.String
		item.ValidationRef = validationRef.String
		item.LogArtifactID = logArtifactID.String
		return nil
	})
	return
}

func (p *PostgresStore) recordPRAttemptDirect(attempt improvement.PRAttempt) (item improvement.PRAttempt, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		if attempt.ID == "" {
			attempt.ID = nextID("pr", 0)
		}
		if attempt.CreatedAt.IsZero() {
			attempt.CreatedAt = now
		}
		if attempt.ProposalID != "" {
			proposal, err := selectProposalTx(tx, attempt.ProposalID, true)
			if err != nil {
				return err
			}
			attempt.ConversationID = firstNonEmpty(attempt.ConversationID, proposal.ConversationID)
			attempt.CaseID = firstNonEmpty(attempt.CaseID, proposal.CaseID)
			attempt.OriginTraceID = firstNonEmpty(attempt.OriginTraceID, proposal.OriginTraceID)
			attempt.AttemptID = firstNonEmpty(attempt.AttemptID, proposal.CurrentAttemptID)
			if attempt.Status == string(review.ProposalPROpen) {
				proposal.Status = review.ProposalPROpen
				proposal.ActiveSlotConsuming = true
				proposal.CurrentAttemptID = firstNonEmpty(attempt.AttemptID)
				if err := updateProposalOperationalStateTx(tx, proposal); err != nil {
					return err
				}
			}
		}
		temp := newSubsetStore()
		temp.prAttempts[attempt.ID] = attempt
		if err := persistPRAttempts(tx, temp); err != nil {
			return err
		}
		item = attempt
		return nil
	})
	return
}

func selectProposalTx(tx *sql.Tx, proposalID string, forUpdate bool) (review.Proposal, error) {
	query := `select id, version, trace_id, conversation_id, case_id, origin_trace_id, evidence_trace_ids, title, category, summary, status, reviewer, candidate_key, target_layer, target_kind, target_ref, source_eval_ids, risk_tier, proposed_scope, evidence_artifact_ids, active_slot_consuming, review_deadline, prior_similar_proposal_ids, new_evidence_since_last_rejection, current_attempt_id, attempt_count, auto_retry_budget_remaining, last_failure_class, next_retry_action, line_stopped_by, line_stop_reason, line_stopped_at, recommended_intervention_kind, recommended_intervention_rationale, target_surface, touched_files, validation_plan, material_risk_summary, recommended_disposition, created_at from proposal where id = $1`
	if forUpdate {
		query += ` for update`
	}
	var item review.Proposal
	var status, targetLayer string
	var conversationID, caseID, originTraceID, reviewer, targetKind, targetRef, currentAttemptID, lastFailureClass, nextRetryAction, lineStoppedBy, lineStopReason, recommendedKind, recommendedRationale, targetSurface, validationPlan, materialRiskSummary, recommendedDisposition sql.NullString
	var evidenceTraceIDs, sourceEvalIDs, evidenceArtifactIDs, priorSimilarProposalIDs, touchedFiles []byte
	var reviewDeadline, lineStoppedAt sql.NullTime
	err := tx.QueryRow(query, proposalID).Scan(&item.ID, &item.Version, &item.TraceID, &conversationID, &caseID, &originTraceID, &evidenceTraceIDs, &item.Title, &item.Category, &item.Summary, &status, &reviewer, &item.CandidateKey, &targetLayer, &targetKind, &targetRef, &sourceEvalIDs, &item.RiskTier, &item.ProposedScope, &evidenceArtifactIDs, &item.ActiveSlotConsuming, &reviewDeadline, &priorSimilarProposalIDs, &item.NewEvidenceSinceLastRejection, &currentAttemptID, &item.AttemptCount, &item.AutoRetryBudgetRemaining, &lastFailureClass, &nextRetryAction, &lineStoppedBy, &lineStopReason, &lineStoppedAt, &recommendedKind, &recommendedRationale, &targetSurface, &touchedFiles, &validationPlan, &materialRiskSummary, &recommendedDisposition, &item.CreatedAt)
	if err != nil {
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

func updateProposalOperationalStateTx(tx *sql.Tx, item review.Proposal) error {
	item = normalizeProposalTargetFields(item)
	_, err := tx.Exec(`update proposal set reviewer = $2, status = $3, active_slot_consuming = $4, current_attempt_id = $5, attempt_count = $6, auto_retry_budget_remaining = $7, last_failure_class = $8, next_retry_action = $9, line_stopped_by = $10, line_stop_reason = $11, line_stopped_at = $12, recommended_intervention_kind = $13, recommended_intervention_rationale = $14, target_surface = $15, touched_files = $16::jsonb, validation_plan = $17, material_risk_summary = $18, recommended_disposition = $19, version = version + 1 where id = $1`,
		item.ID,
		nullString(item.Reviewer),
		string(item.Status),
		item.ActiveSlotConsuming,
		firstNonEmpty(item.CurrentAttemptID),
		item.AttemptCount,
		item.AutoRetryBudgetRemaining,
		firstNonEmpty(item.LastFailureClass),
		firstNonEmpty(item.NextRetryAction),
		firstNonEmpty(item.LineStoppedBy),
		firstNonEmpty(item.LineStopReason),
		nullTime(item.LineStoppedAt),
		string(item.RecommendedInterventionKind),
		firstNonEmpty(item.RecommendedInterventionRationale),
		firstNonEmpty(item.TargetSurface),
		jsonString(item.TouchedFiles),
		firstNonEmpty(item.ValidationPlan),
		firstNonEmpty(item.MaterialRiskSummary),
		firstNonEmpty(item.RecommendedDisposition),
	)
	return err
}

func selectRepoChangeJobByProposalTx(tx *sql.Tx, proposalID string, forUpdate bool) (improvement.RepoChangeJob, bool, error) {
	query := `select id, proposal_id, attempt_id, conversation_id, case_id, origin_trace_id, candidate_key, status, repo, base_ref, branch_name, allowed_path_globs, context_summary, sandbox_namespace, sandbox_job_name, sandbox_pod_name, validation_error, validation_ref, log_artifact_id, created_at, updated_at from repo_change_job where proposal_id = $1 order by created_at desc limit 1`
	if forUpdate {
		query += ` for update`
	}
	var item improvement.RepoChangeJob
	var attemptID, conversationID, caseID, originTraceID, sandboxNamespace, sandboxJobName, sandboxPodName, validationError, validationRef, logArtifactID sql.NullString
	var allowed []byte
	err := tx.QueryRow(query, proposalID).Scan(&item.ID, &item.ProposalID, &attemptID, &conversationID, &caseID, &originTraceID, &item.CandidateKey, &item.Status, &item.Repo, &item.BaseRef, &item.BranchName, &allowed, &item.ContextSummary, &sandboxNamespace, &sandboxJobName, &sandboxPodName, &validationError, &validationRef, &logArtifactID, &item.CreatedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return improvement.RepoChangeJob{}, false, nil
	}
	if err != nil {
		return improvement.RepoChangeJob{}, false, err
	}
	item.AttemptID = attemptID.String
	item.ConversationID = conversationID.String
	item.CaseID = caseID.String
	item.OriginTraceID = originTraceID.String
	item.AllowedPathGlobs = decodeJSON(allowed, []string{})
	item.SandboxNamespace = sandboxNamespace.String
	item.SandboxJobName = sandboxJobName.String
	item.SandboxPodName = sandboxPodName.String
	item.ValidationError = validationError.String
	item.ValidationRef = validationRef.String
	item.LogArtifactID = logArtifactID.String
	return item, true, nil
}

func selectLatestPRAttemptByProposalTx(tx *sql.Tx, proposalID string) (improvement.PRAttempt, bool, error) {
	var item improvement.PRAttempt
	var attemptID, conversationID, caseID, originTraceID, prURL, headSHA sql.NullString
	err := tx.QueryRow(`select id, proposal_id, attempt_id, conversation_id, case_id, origin_trace_id, repo, branch_name, pr_url, head_sha, status, validation_status, created_at from pr_attempt where proposal_id = $1 order by created_at desc limit 1`, proposalID).Scan(&item.ID, &item.ProposalID, &attemptID, &conversationID, &caseID, &originTraceID, &item.Repo, &item.BranchName, &prURL, &headSHA, &item.Status, &item.ValidationStatus, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return improvement.PRAttempt{}, false, nil
	}
	if err != nil {
		return improvement.PRAttempt{}, false, err
	}
	item.AttemptID = attemptID.String
	item.ConversationID = conversationID.String
	item.CaseID = caseID.String
	item.OriginTraceID = originTraceID.String
	item.PRURL = prURL.String
	item.HeadSHA = headSHA.String
	return item, true, nil
}

func latestEvalScoreForTraceTx(tx *sql.Tx, traceID string) (float64, error) {
	var score sql.NullFloat64
	err := tx.QueryRow(`select overall_score from eval_run where trace_id = $1 order by created_at desc limit 1`, traceID).Scan(&score)
	if err == sql.ErrNoRows || !score.Valid {
		return 0, nil
	}
	return score.Float64, err
}

func selectEventBySourceDedupeTx(tx *sql.Tx, source ingestion.Source, dedupeKey string) (ingestion.EventEnvelope, bool, error) {
	var item ingestion.EventEnvelope
	var sourceName, severity string
	var threadKey, incidentKey, ownershipHint, rawPayloadRef, workflowHint sql.NullString
	var raw []byte
	err := tx.QueryRow(`select id, source, source_event_id, thread_key, incident_key, dedupe_key, severity, normalized_problem_statement, ownership_hint, raw_payload_ref, workflow_hint, metadata, created_at from event_envelope where source = $1 and dedupe_key = $2`, string(source), dedupeKey).Scan(&item.ID, &sourceName, &item.SourceEventID, &threadKey, &incidentKey, &item.DedupeKey, &severity, &item.NormalizedProblemStatement, &ownershipHint, &rawPayloadRef, &workflowHint, &raw, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return ingestion.EventEnvelope{}, false, nil
	}
	if err != nil {
		return ingestion.EventEnvelope{}, false, err
	}
	item.Source = ingestion.Source(sourceName)
	item.ThreadKey = threadKey.String
	item.IncidentKey = incidentKey.String
	item.Severity = ingestion.Severity(severity)
	item.OwnershipHint = ownershipHint.String
	item.RawPayloadRef = rawPayloadRef.String
	item.WorkflowHint = workflowHint.String
	item.Metadata = decodeJSON(raw, map[string]interface{}{})
	return item, true, nil
}

func selectIngestionByEventIDTx(tx *sql.Tx, eventID string) (slack.Ingestion, bool, error) {
	var item slack.Ingestion
	var threadTS, intent, botRole sql.NullString
	err := tx.QueryRow(`select id, event_id, conversation_id, case_id, thread_key, thread_ts, workflow_hint, intent, bot_role, source, channel_id, user_id, text, created_at from ingestion where event_id = $1 order by created_at desc limit 1`, eventID).Scan(&item.ID, &item.EventID, &item.ConversationID, &item.CaseID, &item.ThreadKey, &threadTS, &item.WorkflowHint, &intent, &botRole, &item.Source, &item.ChannelID, &item.UserID, &item.Text, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return slack.Ingestion{}, false, nil
	}
	if err != nil {
		return slack.Ingestion{}, false, err
	}
	item.ThreadTS = threadTS.String
	item.Intent = intent.String
	item.BotRole = slack.BotRole(botRole.String)
	return item, true, nil
}

func buildSlackEventEnvelope(envelope slack.SlackEnvelope) ingestion.EventEnvelope {
	conversationKey := slackConversationKey(envelope)
	event := ingestion.EventEnvelope{
		ID:                         nextID("evt", 0),
		Source:                     ingestion.SourceSlack,
		SourceEventID:              envelope.TS,
		ThreadKey:                  conversationKey,
		DedupeKey:                  envelope.TS,
		Severity:                   severityFromText(envelope.Text),
		NormalizedProblemStatement: envelope.Text,
		OwnershipHint:              "platform",
		WorkflowHint:               deriveWorkflowHint(envelope.Text),
		Metadata: map[string]interface{}{
			"team_id":          envelope.TeamID,
			"channel_id":       envelope.ChannelID,
			"user_id":          envelope.UserID,
			"thread_ts":        envelope.ThreadTS,
			"conversation_key": conversationKey,
			"bot_role":         envelope.BotRole,
			"files":            envelope.Files,
		},
		CreatedAt: envelope.CreatedAt,
	}
	event.RawPayloadRef = fmt.Sprintf("memory://slack/%s/%s.json", envelope.ChannelID, strings.ReplaceAll(envelope.TS, ".", "-"))
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	return event
}

func selectConversationByExternalKeyTx(tx *sql.Tx, externalKey string) (conversation.Conversation, bool, error) {
	var item conversation.Conversation
	var participantIDs []byte
	var latestEventID, activeCaseID sql.NullString
	var sourceName string
	err := tx.QueryRow(`select id, source, external_key, external_conversation, title, status, participant_ids, active_case_id, latest_event_id, created_at, updated_at from conversation where external_key = $1`, externalKey).Scan(&item.ID, &sourceName, &item.ExternalKey, &item.ExternalConversation, &item.Title, &item.Status, &participantIDs, &activeCaseID, &latestEventID, &item.CreatedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return conversation.Conversation{}, false, nil
	}
	if err != nil {
		return conversation.Conversation{}, false, err
	}
	item.Source = ingestion.Source(sourceName)
	item.ParticipantIDs = decodeJSON(participantIDs, []string{})
	item.ActiveCaseID = activeCaseID.String
	item.LatestEventID = latestEventID.String
	return item, true, nil
}

func selectActiveCaseForConversationTx(tx *sql.Tx, conv conversation.Conversation) (conversation.Case, bool, error) {
	if strings.TrimSpace(conv.ActiveCaseID) != "" {
		item, ok, err := selectCaseByIDTx(tx, conv.ActiveCaseID, true)
		if err != nil {
			return conversation.Case{}, false, err
		}
		if ok && item.Status == conversation.CaseActive {
			return item, true, nil
		}
	}
	row := tx.QueryRow(`select id from case_record where conversation_id = $1 and status = $2 order by updated_at desc limit 1`, conv.ID, string(conversation.CaseActive))
	var caseID string
	if err := row.Scan(&caseID); err == sql.ErrNoRows {
		return conversation.Case{}, false, nil
	} else if err != nil {
		return conversation.Case{}, false, err
	}
	return selectCaseByIDTx(tx, caseID, true)
}

func selectCaseByIDTx(tx *sql.Tx, caseID string, forUpdate bool) (conversation.Case, bool, error) {
	query := `select id, conversation_id, kind, intent, title, summary, status, approval_mode, response_mode, assigned_bot, opened_by_event_id, closed_by_event_id, latest_trace_id, resolution_state, resolved_at, latest_outcome_id, outcome_score, superseded_by_case_id, created_at, updated_at, closed_at from case_record where id = $1`
	if forUpdate {
		query += ` for update`
	}
	var item conversation.Case
	var status, resolutionState string
	var approvalMode, responseMode, openedByEventID, closedByEventID, latestTraceID, latestOutcomeID, supersededByCaseID sql.NullString
	var resolvedAt, closedAt sql.NullTime
	err := tx.QueryRow(query, caseID).Scan(&item.ID, &item.ConversationID, &item.Kind, &item.Intent, &item.Title, &item.Summary, &status, &approvalMode, &responseMode, &item.AssignedBot, &openedByEventID, &closedByEventID, &latestTraceID, &resolutionState, &resolvedAt, &latestOutcomeID, &item.OutcomeScore, &supersededByCaseID, &item.CreatedAt, &item.UpdatedAt, &closedAt)
	if err == sql.ErrNoRows {
		return conversation.Case{}, false, nil
	}
	if err != nil {
		return conversation.Case{}, false, err
	}
	item.Status = conversation.CaseStatus(status)
	item.ApprovalMode = approvalMode.String
	item.ResponseMode = responseMode.String
	item.OpenedByEventID = openedByEventID.String
	item.ClosedByEventID = closedByEventID.String
	item.LatestTraceID = latestTraceID.String
	item.ResolutionState = conversation.ResolutionState(resolutionState)
	if resolvedAt.Valid {
		item.ResolvedAt = &resolvedAt.Time
	}
	item.LatestOutcomeID = latestOutcomeID.String
	item.SupersededByCaseID = supersededByCaseID.String
	if closedAt.Valid {
		item.ClosedAt = &closedAt.Time
	}
	return item, true, nil
}

func selectWorkflowByCaseTx(tx *sql.Tx, caseID string, forUpdate bool) (Workflow, bool, error) {
	query := `select id, ingestion_id, trace_id, conversation_id, case_id, thread_key, kind, intent, assigned_bot, approval_mode, response_mode, status, last_error, created_at, updated_at, completed_at from workflow where case_id = $1 order by created_at asc limit 1`
	if forUpdate {
		query += ` for update`
	}
	var item Workflow
	var ingestionID, traceID, conversationID, intent, approvalMode, responseMode, lastError sql.NullString
	var completedAt sql.NullTime
	err := tx.QueryRow(query, caseID).Scan(&item.ID, &ingestionID, &traceID, &conversationID, &item.CaseID, &item.ThreadKey, &item.Kind, &intent, &item.AssignedBot, &approvalMode, &responseMode, &item.Status, &lastError, &item.CreatedAt, &item.UpdatedAt, &completedAt)
	if err == sql.ErrNoRows {
		return Workflow{}, false, nil
	}
	if err != nil {
		return Workflow{}, false, err
	}
	item.IngestionID = ingestionID.String
	item.TraceID = traceID.String
	item.ConversationID = conversationID.String
	item.Intent = intent.String
	item.ApprovalMode = approvalMode.String
	item.ResponseMode = responseMode.String
	item.LastError = lastError.String
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}
	return item, true, nil
}

func (p *PostgresStore) createIngestionDirect(envelope slack.SlackEnvelope) (created slack.Ingestion, err error) {
	event := buildSlackEventEnvelope(envelope)
	err = p.withTx(func(tx *sql.Tx) error {
		existingEvent, ok, err := selectEventBySourceDedupeTx(tx, event.Source, event.DedupeKey)
		if err != nil {
			return err
		}
		if ok {
			found, foundOK, err := selectIngestionByEventIDTx(tx, existingEvent.ID)
			if err != nil {
				return err
			}
			if !foundOK {
				return fmt.Errorf("event %s already exists without ingestion row", existingEvent.ID)
			}
			created = found
			return nil
		}
		if _, err := tx.Exec(`insert into event_envelope (id, source, source_event_id, thread_key, incident_key, dedupe_key, severity, normalized_problem_statement, ownership_hint, raw_payload_ref, workflow_hint, metadata, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,$13)`,
			event.ID, string(event.Source), event.SourceEventID, nullString(event.ThreadKey), nullString(event.IncidentKey), event.DedupeKey, string(event.Severity), event.NormalizedProblemStatement, nullString(event.OwnershipHint), nullString(event.RawPayloadRef), nullString(event.WorkflowHint), jsonString(event.Metadata), event.CreatedAt,
		); err != nil {
			return err
		}
		createdAt := event.CreatedAt
		intent := intentForWorkflowHint(event.WorkflowHint)
		assignedBot := assignedBotFor(event.WorkflowHint)
		channelID := stringFromMetadata(event.Metadata, "channel_id")
		userID := stringFromMetadata(event.Metadata, "user_id")
		threadTS := stringFromMetadata(event.Metadata, "thread_ts")
		if raw := strings.TrimSpace(stringFromMetadata(event.Metadata, "bot_role")); raw != "" {
			assignedBot = raw
		}
		if channelID == "" {
			channelID = string(event.Source)
		}
		if userID == "" {
			userID = "system"
		}
		if threadTS == "" {
			threadTS = event.SourceEventID
		}
		externalKey := conversationKeyForEvent(event)
		conv, hasConv, err := selectConversationByExternalKeyTx(tx, externalKey)
		if err != nil {
			return err
		}
		if hasConv {
			conv.LatestEventID = event.ID
			conv.UpdatedAt = createdAt
			conv.Title = firstNonEmpty(conv.Title, conversation.NormalizeTitle(event.WorkflowHint, event.NormalizedProblemStatement))
			conv.ParticipantIDs = appendUnique(conv.ParticipantIDs, userID)
		} else {
			conv = conversation.Conversation{
				ID:                   nextID("conv", 0),
				Source:               event.Source,
				ExternalKey:          externalKey,
				ExternalConversation: externalKey,
				Title:                conversation.NormalizeTitle(event.WorkflowHint, event.NormalizedProblemStatement),
				Status:               conversation.StatusActive,
				ParticipantIDs:       compactStrings([]string{userID}),
				LatestEventID:        event.ID,
				CreatedAt:            createdAt,
				UpdatedAt:            createdAt,
			}
		}
		temp := newSubsetStore()
		temp.conversations[conv.ID] = conv
		if err := persistConversations(tx, temp); err != nil {
			return err
		}

		entry := externalConversationEntry(conv.ID, event, createdAt)
		temp = newSubsetStore()
		temp.conversationEntries = append(temp.conversationEntries, entry)
		if err := persistConversationEntries(tx, temp); err != nil {
			return err
		}

		caseRecord, hasCase, err := selectActiveCaseForConversationTx(tx, conv)
		if err != nil {
			return err
		}
		if hasCase && caseRecord.Kind == event.WorkflowHint && caseRecord.Status == conversation.CaseActive {
			caseRecord.Summary = event.NormalizedProblemStatement
			caseRecord.UpdatedAt = createdAt
			caseRecord.OpenedByEventID = firstNonEmpty(caseRecord.OpenedByEventID, event.ID)
		} else {
			if hasCase {
				caseRecord.Status = conversation.CaseSuperseded
				caseRecord.ClosedByEventID = event.ID
				caseRecord.UpdatedAt = createdAt
				caseRecord.ClosedAt = &createdAt
				temp = newSubsetStore()
				temp.cases[caseRecord.ID] = caseRecord
				if err := persistCases(tx, temp); err != nil {
					return err
				}
			}
			caseRecord = conversation.Case{
				ID:              nextID("case", 0),
				ConversationID:  conv.ID,
				Kind:            event.WorkflowHint,
				Intent:          intent,
				Title:           conversation.NormalizeTitle(event.WorkflowHint, event.NormalizedProblemStatement),
				Summary:         event.NormalizedProblemStatement,
				Status:          conversation.CaseActive,
				ApprovalMode:    approvalModeForIntent(intent),
				ResponseMode:    responseModeForIntent(intent),
				AssignedBot:     assignedBot,
				OpenedByEventID: event.ID,
				ResolutionState: conversation.ResolutionUnresolved,
				CreatedAt:       createdAt,
				UpdatedAt:       createdAt,
			}
		}
		temp = newSubsetStore()
		temp.cases[caseRecord.ID] = caseRecord
		if err := persistCases(tx, temp); err != nil {
			return err
		}

		workflow, hasWorkflow, err := selectWorkflowByCaseTx(tx, caseRecord.ID, true)
		if err != nil {
			return err
		}
		if hasWorkflow {
			workflow.ConversationID = caseRecord.ConversationID
			workflow.CaseID = caseRecord.ID
			workflow.ThreadKey = conversationKeyForCase(caseRecord, map[string]conversation.Conversation{conv.ID: conv})
			workflow.Kind = caseRecord.Kind
			workflow.Intent = caseRecord.Intent
			workflow.AssignedBot = caseRecord.AssignedBot
			workflow.ApprovalMode = caseRecord.ApprovalMode
			workflow.ResponseMode = caseRecord.ResponseMode
			workflow.Status = "queued"
			workflow.UpdatedAt = createdAt
		} else {
			workflow = Workflow{
				ID:             nextID("wf", 0),
				ConversationID: caseRecord.ConversationID,
				CaseID:         caseRecord.ID,
				ThreadKey:      conversationKeyForCase(caseRecord, map[string]conversation.Conversation{conv.ID: conv}),
				Kind:           caseRecord.Kind,
				Intent:         caseRecord.Intent,
				AssignedBot:    caseRecord.AssignedBot,
				ApprovalMode:   caseRecord.ApprovalMode,
				ResponseMode:   caseRecord.ResponseMode,
				Status:         "queued",
				CreatedAt:      createdAt,
				UpdatedAt:      createdAt,
			}
		}

		ingestionID := nextID("ing", 0)
		created = slack.Ingestion{
			ID:             ingestionID,
			EventID:        event.ID,
			ConversationID: conv.ID,
			CaseID:         caseRecord.ID,
			ThreadKey:      conv.ExternalKey,
			ThreadTS:       threadTS,
			WorkflowHint:   event.WorkflowHint,
			Intent:         intent,
			BotRole:        slack.BotRole(assignedBot),
			Source:         string(event.Source),
			ChannelID:      channelID,
			UserID:         userID,
			Text:           event.NormalizedProblemStatement,
			CreatedAt:      createdAt,
		}
		temp = newSubsetStore()
		temp.ingestions = append(temp.ingestions, created)
		if err := persistIngestions(tx, temp); err != nil {
			return err
		}

		temp = newSubsetStore()
		temp.assignments = append(temp.assignments, Assignment{
			ID:             nextID("as", 0),
			ConversationID: conv.ID,
			CaseID:         caseRecord.ID,
			ThreadKey:      conv.ExternalKey,
			AssignedBot:    assignedBot,
			Confidence:     routeConfidenceForEvent(event),
			Rationale:      routingRationale(event),
			CreatedAt:      createdAt,
		})
		if err := persistAssignments(tx, temp); err != nil {
			return err
		}
		if _, err := tx.Exec(`insert into thread_policy (thread_key, state, owner_bot, muted, close_reason, last_policy_version, updated_at) values ($1,$2,$3,$4,$5,$6,$7) on conflict (thread_key) do nothing`,
			conv.ExternalKey,
			string(policy.ThreadStateActive),
			assignedBot,
			false,
			nil,
			"conversation-v2",
			createdAt,
		); err != nil {
			return err
		}

		traceID := nextID("trace", 0)
		supersedesTraceID := ""
		rows, err := tx.Query(`select trace_id, workflow_id, ingestion_id from trace_summary where case_id = $1 and status in ($2, $3, $4) order by started_at asc`, caseRecord.ID, string(events.StatusQueued), string(events.StatusRunning), string(events.StatusReplayed))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var oldTraceID, oldWorkflowID, oldIngestionID string
			if err := rows.Scan(&oldTraceID, &oldWorkflowID, &oldIngestionID); err != nil {
				return err
			}
			if supersedesTraceID == "" {
				supersedesTraceID = oldTraceID
			}
			if _, err := tx.Exec(`update trace_summary set status = $2, ended_at = $3, event_count = event_count + 1 where trace_id = $1`, oldTraceID, string(events.StatusSuppressed), createdAt); err != nil {
				return err
			}
			if _, err := tx.Exec(`insert into trace_event (trace_id, ingestion_id, workflow_id, conversation_id, case_id, trigger_event_id, parent_event_id, plane, service, actor, event_type, status, started_at, ended_at, payload_ref, artifact_ref, cost_tokens, latency_ms, description) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
				oldTraceID, oldIngestionID, oldWorkflowID, conv.ID, caseRecord.ID, event.ID, nil, "control", "control-plane", "supersession", "trace.superseded", string(events.StatusSuppressed), createdAt, createdAt, nil, nil, 0, 0, fmt.Sprintf("Superseded by newer trace %s for case %s.", traceID, caseRecord.ID),
			); err != nil {
				return err
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		actionRows, err := tx.Query(`select id, coalesce((select count(*) from action_result ar where ar.action_intent_id = ai.id), 0) from action_intent ai where case_id = $1 and status not in ($2, $3, $4, $5, $6)`, caseRecord.ID, string(action.StatusSucceeded), string(action.StatusFailed), string(action.StatusBlocked), string(action.StatusCanceled), string(action.StatusSuperseded))
		if err != nil {
			return err
		}
		defer actionRows.Close()
		for actionRows.Next() {
			var actionIntentID string
			var attemptCount int
			if err := actionRows.Scan(&actionIntentID, &attemptCount); err != nil {
				return err
			}
			if _, err := tx.Exec(`update action_intent set status = $2, superseded_by_action_id = $3, updated_at = $4 where id = $1`,
				actionIntentID,
				string(action.StatusSuperseded),
				traceID,
				createdAt,
			); err != nil {
				return err
			}
			if err := insertActionResult(tx, action.Result{
				ID:             nextUUID("ares"),
				ActionIntentID: actionIntentID,
				AttemptNumber:  attemptCount + 1,
				Executor:       "supersession",
				Status:         action.StatusSuperseded,
				ErrorCode:      "trace_superseded",
				ErrorMessage:   fmt.Sprintf("Superseded by newer trace %s", traceID),
				StartedAt:      createdAt,
				CompletedAt:    createdAt,
			}); err != nil {
				return err
			}
		}
		if err := actionRows.Err(); err != nil {
			return err
		}

		workflow.IngestionID = ingestionID
		workflow.TraceID = traceID
		workflow.Status = "queued"
		workflow.UpdatedAt = createdAt
		temp = newSubsetStore()
		temp.workflows = append(temp.workflows, workflow)
		if err := persistWorkflows(tx, temp); err != nil {
			return err
		}

		caseRecord.LatestTraceID = traceID
		caseRecord.UpdatedAt = createdAt
		temp = newSubsetStore()
		temp.cases[caseRecord.ID] = caseRecord
		if err := persistCases(tx, temp); err != nil {
			return err
		}
		conv.ActiveCaseID = caseRecord.ID
		conv.LatestEventID = event.ID
		conv.UpdatedAt = createdAt
		if !hasConv && conv.Title == "" {
			conv.Title = conversation.NormalizeTitle(caseRecord.Kind, event.NormalizedProblemStatement)
		}
		temp = newSubsetStore()
		temp.conversations[conv.ID] = conv
		if err := persistConversations(tx, temp); err != nil {
			return err
		}

		trace := events.Trace{
			Summary: events.TraceSummary{
				TraceID:            traceID,
				IngestionID:        ingestionID,
				WorkflowID:         workflow.ID,
				ConversationID:     conv.ID,
				CaseID:             caseRecord.ID,
				TriggerEventID:     event.ID,
				SupersedesTraceID:  supersedesTraceID,
				ThreadKey:          conv.ExternalKey,
				WorkflowKind:       caseRecord.Kind,
				Status:             events.StatusQueued,
				StartedAt:          createdAt,
				EndedAt:            createdAt,
				EventCount:         2,
				ArtifactCount:      1,
				ReasoningStepCount: 1,
			},
			Events: []events.TraceEvent{
				{
					TraceID:        traceID,
					IngestionID:    ingestionID,
					WorkflowID:     workflow.ID,
					ConversationID: conv.ID,
					CaseID:         caseRecord.ID,
					TriggerEventID: event.ID,
					Plane:          "edge",
					Service:        "control-plane",
					Actor:          "event-ingestor",
					EventType:      "event.ingested",
					Status:         events.StatusCompleted,
					StartedAt:      createdAt,
					EndedAt:        &createdAt,
					PayloadRef:     event.RawPayloadRef,
					Description:    fmt.Sprintf("%s event normalized into conversation %s.", event.Source, conv.ExternalKey),
				},
				{
					TraceID:        traceID,
					IngestionID:    ingestionID,
					WorkflowID:     workflow.ID,
					ConversationID: conv.ID,
					CaseID:         caseRecord.ID,
					TriggerEventID: event.ID,
					Plane:          "control",
					Service:        "control-plane",
					Actor:          "router-policy",
					EventType:      "workflow.queued",
					Status:         events.StatusQueued,
					StartedAt:      createdAt,
					Description:    fmt.Sprintf("Queued %s trace for case %s and bot %s.", intent, caseRecord.ID, assignedBot),
				},
			},
			Artifacts: []events.Artifact{
				{
					ID:          nextID("artifact", 0),
					TraceID:     traceID,
					Kind:        "event_payload",
					ContentType: "application/json",
					URL:         event.RawPayloadRef,
					SizeBytes:   2048,
					Source:      "event-bus",
				},
			},
			Reasoning: []events.ReasoningStep{
				{
					ID:             nextID("reason", 0),
					TraceID:        traceID,
					WorkflowID:     workflow.ID,
					ConversationID: conv.ID,
					CaseID:         caseRecord.ID,
					StepType:       "goal_framing",
					Summary:        fmt.Sprintf("Treat event %s as %s within case %s.", event.ID, intent, caseRecord.ID),
					EvidenceRefs: []events.EvidenceRef{
						{Kind: "event", Ref: event.ID, Summary: event.NormalizedProblemStatement},
						{Kind: "conversation_entry", Ref: entry.ID, Summary: entry.Body},
					},
					Alternatives: []string{"incident", "feature_request", "question"},
					Confidence:   routeConfidenceForEvent(event),
					Decision:     fmt.Sprintf("route:%s assign:%s case:%s", intent, assignedBot, caseRecord.ID),
					CreatedAt:    createdAt,
				},
			},
		}
		temp = newSubsetStore()
		temp.traces[traceID] = trace
		if err := persistTraces(tx, temp); err != nil {
			return err
		}

		return nil
	})
	return
}
