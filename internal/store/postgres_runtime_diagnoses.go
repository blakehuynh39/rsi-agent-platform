package store

import (
	"database/sql"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
)

func loadRuntimeDiagnoses(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, candidate_key, repo, conversation_id, case_id, latest_trace_id, status, subsystem, failure_mode, summary, evidence_refs, missing_evidence, recommended_fix, target_surface, validation_plan, session_scope_kind, session_scope_id, last_result, last_error, last_attempted_at, promoted_at, created_at, updated_at from runtime_diagnosis order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.RuntimeDiagnosis
		var conversationID, caseID, latestTraceID, subsystem, failureMode, summary, recommendedFix, targetSurface, validationPlan, sessionScopeKind, sessionScopeID, lastError sql.NullString
		var evidenceRefs, missingEvidence, lastResult []byte
		var lastAttemptedAt, promotedAt sql.NullTime
		var status string
		if err := rows.Scan(&item.ID, &item.CandidateKey, &item.Repo, &conversationID, &caseID, &latestTraceID, &status, &subsystem, &failureMode, &summary, &evidenceRefs, &missingEvidence, &recommendedFix, &targetSurface, &validationPlan, &sessionScopeKind, &sessionScopeID, &lastResult, &lastError, &lastAttemptedAt, &promotedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.LatestTraceID = latestTraceID.String
		item.Status = improvement.RuntimeDiagnosisStatus(status)
		item.Subsystem = subsystem.String
		item.FailureMode = failureMode.String
		item.Summary = summary.String
		item.EvidenceRefs = decodeJSON(evidenceRefs, []string{})
		item.MissingEvidence = decodeJSON(missingEvidence, []string{})
		item.RecommendedFix = recommendedFix.String
		item.TargetSurface = targetSurface.String
		item.ValidationPlan = validationPlan.String
		item.SessionScopeKind = sessionScopeKind.String
		item.SessionScopeID = sessionScopeID.String
		item.LastResult = decodeJSON(lastResult, map[string]any{})
		item.LastError = lastError.String
		if lastAttemptedAt.Valid {
			t := lastAttemptedAt.Time
			item.LastAttemptedAt = &t
		}
		if promotedAt.Valid {
			t := promotedAt.Time
			item.PromotedAt = &t
		}
		store.runtimeDiagnoses[item.ID] = item
	}
	return rows.Err()
}

func persistRuntimeDiagnoses(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.runtimeDiagnoses)
	for _, key := range keys {
		item := store.runtimeDiagnoses[key]
		if _, err := tx.Exec(`insert into runtime_diagnosis (id, candidate_key, repo, conversation_id, case_id, latest_trace_id, status, subsystem, failure_mode, summary, evidence_refs, missing_evidence, recommended_fix, target_surface, validation_plan, session_scope_kind, session_scope_id, last_result, last_error, last_attempted_at, promoted_at, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12::jsonb,$13,$14,$15,$16,$17,$18::jsonb,$19,$20,$21,$22,$23)
			on conflict (id) do update set
				candidate_key = excluded.candidate_key,
				repo = excluded.repo,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				latest_trace_id = excluded.latest_trace_id,
				status = excluded.status,
				subsystem = excluded.subsystem,
				failure_mode = excluded.failure_mode,
				summary = excluded.summary,
				evidence_refs = excluded.evidence_refs,
				missing_evidence = excluded.missing_evidence,
				recommended_fix = excluded.recommended_fix,
				target_surface = excluded.target_surface,
				validation_plan = excluded.validation_plan,
				session_scope_kind = excluded.session_scope_kind,
				session_scope_id = excluded.session_scope_id,
				last_result = excluded.last_result,
				last_error = excluded.last_error,
				last_attempted_at = excluded.last_attempted_at,
				promoted_at = excluded.promoted_at,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at`,
			item.ID,
			item.CandidateKey,
			item.Repo,
			nullString(item.ConversationID),
			nullString(item.CaseID),
			nullString(item.LatestTraceID),
			string(item.Status),
			nullString(item.Subsystem),
			nullString(item.FailureMode),
			nullString(item.Summary),
			jsonString(item.EvidenceRefs),
			jsonString(item.MissingEvidence),
			nullString(item.RecommendedFix),
			nullString(item.TargetSurface),
			nullString(item.ValidationPlan),
			nullString(item.SessionScopeKind),
			nullString(item.SessionScopeID),
			jsonString(item.LastResult),
			nullString(item.LastError),
			nullTime(item.LastAttemptedAt),
			nullTime(item.PromotedAt),
			item.CreatedAt,
			item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStore) ListRuntimeDiagnoses() []improvement.RuntimeDiagnosis {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListRuntimeDiagnoses()
}

func (p *PostgresStore) GetRuntimeDiagnosis(diagnosisID string) (improvement.RuntimeDiagnosis, bool) {
	store, err := p.readStore()
	if err != nil {
		return improvement.RuntimeDiagnosis{}, false
	}
	return store.GetRuntimeDiagnosis(diagnosisID)
}
