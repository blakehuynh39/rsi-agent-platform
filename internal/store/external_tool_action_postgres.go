package store

import (
	"database/sql"
	"time"
)

const externalToolActionColumns = `id, surface, operation, target_ref, idempotency_key, request_hash, state, actor, reason, destructive, execution_id, operation_id, trace_id, workflow_id, conversation_id, response_summary, error_message, source_ref, wiki_audit_id, mirror_effect, created_at, updated_at, completed_at`

func (p *PostgresStore) ListExternalToolActions() []ExternalToolAction {
	rows, err := p.db.Query(`select ` + externalToolActionColumns + ` from external_tool_action order by created_at desc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []ExternalToolAction{}
	for rows.Next() {
		item, err := scanExternalToolAction(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) GetExternalToolAction(actionID string) (ExternalToolAction, bool) {
	row := p.db.QueryRow(`select `+externalToolActionColumns+` from external_tool_action where id = $1`, actionID)
	item, err := scanExternalToolAction(row)
	return item, err == nil
}

func (p *PostgresStore) GetExternalToolActionByIdempotency(surface string, operation string, idempotencyKey string) (ExternalToolAction, bool) {
	row := p.db.QueryRow(`select `+externalToolActionColumns+` from external_tool_action where surface = $1 and operation = $2 and idempotency_key = $3`, surface, operation, idempotencyKey)
	item, err := scanExternalToolAction(row)
	return item, err == nil
}

func (p *PostgresStore) UpsertExternalToolAction(input ExternalToolActionCreateInput, now time.Time) (ExternalToolAction, ExternalToolActionUpsertStatus, error) {
	if existing, ok := p.GetExternalToolActionByIdempotency(input.Surface, input.Operation, input.IdempotencyKey); ok {
		status := ExternalToolActionUpsertReplay
		if existing.RequestHash != input.RequestHash {
			status = ExternalToolActionUpsertConflict
		}
		return existing, status, nil
	}
	item, err := NewExternalToolAction(input, now)
	if err != nil {
		return ExternalToolAction{}, "", err
	}
	_, err = p.db.Exec(`insert into external_tool_action (`+externalToolActionColumns+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20::jsonb,$21,$22,$23)`,
		item.ID, item.Surface, item.Operation, nullString(item.TargetRef), item.IdempotencyKey, item.RequestHash, string(item.State), item.Actor, nullString(item.Reason), item.Destructive,
		nullString(item.ExecutionID), nullString(item.OperationID), nullString(item.TraceID), nullString(item.WorkflowID), nullString(item.ConversationID),
		nullString(item.ResponseSummary), nullString(item.ErrorMessage), nullString(item.SourceRef), nullString(item.WikiAuditID), jsonString(item.MirrorEffect),
		item.CreatedAt, item.UpdatedAt, nullTime(item.CompletedAt),
	)
	if err != nil {
		if existing, ok := p.GetExternalToolActionByIdempotency(input.Surface, input.Operation, input.IdempotencyKey); ok {
			status := ExternalToolActionUpsertReplay
			if existing.RequestHash != input.RequestHash {
				status = ExternalToolActionUpsertConflict
			}
			return existing, status, nil
		}
		return ExternalToolAction{}, "", err
	}
	return item, ExternalToolActionUpsertCreated, nil
}

func (p *PostgresStore) UpdateExternalToolActionResult(actionID string, update ExternalToolActionResultUpdate, now time.Time) (ExternalToolAction, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	completedAt := (*time.Time)(nil)
	if update.State == ExternalToolActionStateSucceeded || update.State == ExternalToolActionStateFailed {
		completedAt = &now
	}
	state := update.State
	if state == "" {
		state = ExternalToolActionStateRequested
	}
	row := p.db.QueryRow(`update external_tool_action set state=$2, response_summary=$3, error_message=$4, source_ref=$5, wiki_audit_id=$6, mirror_effect=$7::jsonb, updated_at=$8, completed_at=$9 where id=$1 returning `+externalToolActionColumns,
		actionID, string(state), nullString(update.ResponseSummary), nullString(update.ErrorMessage), nullString(update.SourceRef), nullString(update.WikiAuditID), jsonString(update.MirrorEffect), now, nullTime(completedAt),
	)
	return scanExternalToolAction(row)
}

type externalToolActionScanner interface{ Scan(dest ...any) error }

func scanExternalToolAction(row externalToolActionScanner) (ExternalToolAction, error) {
	var item ExternalToolAction
	var targetRef, reason, executionID, operationID, traceID, workflowID, conversationID, responseSummary, errorMessage, sourceRef, wikiAuditID sql.NullString
	var mirrorEffectRaw []byte
	var completedAt sql.NullTime
	if err := row.Scan(&item.ID, &item.Surface, &item.Operation, &targetRef, &item.IdempotencyKey, &item.RequestHash, &item.State, &item.Actor, &reason, &item.Destructive, &executionID, &operationID, &traceID, &workflowID, &conversationID, &responseSummary, &errorMessage, &sourceRef, &wikiAuditID, &mirrorEffectRaw, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
		return ExternalToolAction{}, err
	}
	item.TargetRef = targetRef.String
	item.Reason = reason.String
	item.ExecutionID = executionID.String
	item.OperationID = operationID.String
	item.TraceID = traceID.String
	item.WorkflowID = workflowID.String
	item.ConversationID = conversationID.String
	item.ResponseSummary = responseSummary.String
	item.ErrorMessage = errorMessage.String
	item.SourceRef = sourceRef.String
	item.WikiAuditID = wikiAuditID.String
	item.MirrorEffect = decodeJSON(mirrorEffectRaw, map[string]any{})
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}
	return cloneExternalToolAction(item), nil
}
