package store

import (
	"database/sql"
)

func loadWorkflowLines(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select case_id, conversation_id, status, current_workflow_id, latest_workflow_id, attempt_count, auto_retry_budget_remaining, last_failure_class, next_retry_action, retry_after, line_stop_reason, version, created_at, updated_at, completed_at from workflow_line order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item WorkflowLine
		var currentWorkflowID, latestWorkflowID, lastFailureClass, nextRetryAction, lineStopReason sql.NullString
		var retryAfter, completedAt sql.NullTime
		if err := rows.Scan(&item.CaseID, &item.ConversationID, &item.Status, &currentWorkflowID, &latestWorkflowID, &item.AttemptCount, &item.AutoRetryBudgetRemaining, &lastFailureClass, &nextRetryAction, &retryAfter, &lineStopReason, &item.Version, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
			return err
		}
		item.CurrentWorkflowID = currentWorkflowID.String
		item.LatestWorkflowID = latestWorkflowID.String
		item.LastFailureClass = lastFailureClass.String
		item.NextRetryAction = nextRetryAction.String
		item.LineStopReason = lineStopReason.String
		if retryAfter.Valid {
			t := retryAfter.Time
			item.RetryAfter = &t
		}
		if completedAt.Valid {
			t := completedAt.Time
			item.CompletedAt = &t
		}
		store.workflowLines[item.CaseID] = normalizeWorkflowLine(item)
	}
	return rows.Err()
}

func persistWorkflowLines(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.workflowLines {
		if _, err := tx.Exec(`insert into workflow_line (case_id, conversation_id, status, current_workflow_id, latest_workflow_id, attempt_count, auto_retry_budget_remaining, last_failure_class, next_retry_action, retry_after, line_stop_reason, version, created_at, updated_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			on conflict (case_id) do update set
				conversation_id = excluded.conversation_id,
				status = excluded.status,
				current_workflow_id = excluded.current_workflow_id,
				latest_workflow_id = excluded.latest_workflow_id,
				attempt_count = excluded.attempt_count,
				auto_retry_budget_remaining = excluded.auto_retry_budget_remaining,
				last_failure_class = excluded.last_failure_class,
				next_retry_action = excluded.next_retry_action,
				retry_after = excluded.retry_after,
				line_stop_reason = excluded.line_stop_reason,
				version = excluded.version,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at,
				completed_at = excluded.completed_at`,
			item.CaseID,
			item.ConversationID,
			item.Status,
			nullString(item.CurrentWorkflowID),
			nullString(item.LatestWorkflowID),
			item.AttemptCount,
			item.AutoRetryBudgetRemaining,
			nullString(item.LastFailureClass),
			nullString(item.NextRetryAction),
			nullTime(item.RetryAfter),
			nullString(item.LineStopReason),
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

func replaceWorkflowLineScope(tx *sql.Tx, item WorkflowLine) error {
	temp := newSubsetStore()
	temp.workflowLines[item.CaseID] = item
	return persistWorkflowLines(tx, temp)
}

func (p *PostgresStore) ListWorkflowLines() []WorkflowLine {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListWorkflowLines()
}

func (p *PostgresStore) GetWorkflowLine(caseID string) (WorkflowLine, bool) {
	row := p.db.QueryRow(`select case_id, conversation_id, status, current_workflow_id, latest_workflow_id, attempt_count, auto_retry_budget_remaining, last_failure_class, next_retry_action, retry_after, line_stop_reason, version, created_at, updated_at, completed_at from workflow_line where case_id = $1`, caseID)
	var item WorkflowLine
	var currentWorkflowID, latestWorkflowID, lastFailureClass, nextRetryAction, lineStopReason sql.NullString
	var retryAfter, completedAt sql.NullTime
	if err := row.Scan(&item.CaseID, &item.ConversationID, &item.Status, &currentWorkflowID, &latestWorkflowID, &item.AttemptCount, &item.AutoRetryBudgetRemaining, &lastFailureClass, &nextRetryAction, &retryAfter, &lineStopReason, &item.Version, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
		return WorkflowLine{}, false
	}
	item.CurrentWorkflowID = currentWorkflowID.String
	item.LatestWorkflowID = latestWorkflowID.String
	item.LastFailureClass = lastFailureClass.String
	item.NextRetryAction = nextRetryAction.String
	item.LineStopReason = lineStopReason.String
	if retryAfter.Valid {
		t := retryAfter.Time
		item.RetryAfter = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		item.CompletedAt = &t
	}
	return normalizeWorkflowLine(item), true
}
