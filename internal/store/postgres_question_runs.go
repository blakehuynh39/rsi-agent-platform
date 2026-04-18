package store

import (
	"database/sql"
)

func loadQuestionRuns(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, workflow_id, trace_id, conversation_id, case_id, ingestion_id, role, strategy, status, investigation_spec, evidence_ledger, result, failure_class, failure_summary, last_error, runner_diagnostics, version, created_at, updated_at, completed_at from question_run order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item QuestionRun
		var traceID, conversationID, caseID, ingestionID, role, strategy, failureClass, failureSummary, lastError sql.NullString
		var investigationSpec, evidenceLedger, result, runnerDiagnostics []byte
		var completedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.WorkflowID,
			&traceID,
			&conversationID,
			&caseID,
			&ingestionID,
			&role,
			&strategy,
			&item.Status,
			&investigationSpec,
			&evidenceLedger,
			&result,
			&failureClass,
			&failureSummary,
			&lastError,
			&runnerDiagnostics,
			&item.Version,
			&item.CreatedAt,
			&item.UpdatedAt,
			&completedAt,
		); err != nil {
			return err
		}
		item.TraceID = traceID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.IngestionID = ingestionID.String
		item.Role = role.String
		item.Strategy = strategy.String
		item.InvestigationSpec = decodeJSON(investigationSpec, item.InvestigationSpec)
		item.EvidenceLedger = decodeJSON(evidenceLedger, item.EvidenceLedger)
		item.Result = decodeJSON(result, item.Result)
		item.FailureClass = failureClass.String
		item.FailureSummary = failureSummary.String
		item.LastError = lastError.String
		item.RunnerDiagnostics = decodeJSON(runnerDiagnostics, map[string]any{})
		if completedAt.Valid {
			t := completedAt.Time
			item.CompletedAt = &t
		}
		store.questionRuns[item.ID] = item
	}
	return rows.Err()
}

func persistQuestionRuns(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.questionRuns)
	for _, key := range keys {
		item := store.questionRuns[key]
		if _, err := tx.Exec(`insert into question_run (id, workflow_id, trace_id, conversation_id, case_id, ingestion_id, role, strategy, status, investigation_spec, evidence_ledger, result, failure_class, failure_summary, last_error, runner_diagnostics, version, created_at, updated_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11::jsonb,$12::jsonb,$13,$14,$15,$16::jsonb,$17,$18,$19,$20)
			on conflict (id) do update set
				workflow_id = excluded.workflow_id,
				trace_id = excluded.trace_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				ingestion_id = excluded.ingestion_id,
				role = excluded.role,
				strategy = excluded.strategy,
				status = excluded.status,
				investigation_spec = excluded.investigation_spec,
				evidence_ledger = excluded.evidence_ledger,
				result = excluded.result,
				failure_class = excluded.failure_class,
				failure_summary = excluded.failure_summary,
				last_error = excluded.last_error,
				runner_diagnostics = excluded.runner_diagnostics,
				version = excluded.version,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at,
				completed_at = excluded.completed_at`,
			item.ID,
			item.WorkflowID,
			nullString(item.TraceID),
			nullString(item.ConversationID),
			nullString(item.CaseID),
			nullString(item.IngestionID),
			nullString(item.Role),
			nullString(item.Strategy),
			item.Status,
			jsonString(item.InvestigationSpec),
			jsonString(item.EvidenceLedger),
			jsonString(item.Result),
			nullString(item.FailureClass),
			nullString(item.FailureSummary),
			nullString(item.LastError),
			jsonString(item.RunnerDiagnostics),
			item.Version,
			item.CreatedAt,
			item.UpdatedAt,
			nullTime(item.CompletedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStore) ListQuestionRuns() []QuestionRun {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListQuestionRuns()
}

func (p *PostgresStore) GetQuestionRun(questionRunID string) (QuestionRun, bool) {
	store, err := p.readStore()
	if err != nil {
		return QuestionRun{}, false
	}
	return store.GetQuestionRun(questionRunID)
}
