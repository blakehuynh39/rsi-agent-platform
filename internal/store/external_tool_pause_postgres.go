package store

import (
	"database/sql"
	"strings"
	"time"
)

const externalToolPauseColumns = `id, idempotency_key, conversation_id, workflow_id, trace_id, operation_id, execution_id, hermes_session_id, canonical_tool_name, transport_tool_name, tool_call_id, args_hash, db_read_request_id, sql_sha256, approval_status, tool_outcome, resume_status, approval_ref, result_ref, expires_at, pending_assistant_message, transcript_snapshot, resume_payload, error_message, metadata, created_at, updated_at`

func (p *PostgresStore) ListExternalToolPauses() []ExternalToolPause {
	rows, err := p.db.Query(`select ` + externalToolPauseColumns + ` from external_tool_pause order by created_at desc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []ExternalToolPause{}
	for rows.Next() {
		item, err := scanExternalToolPause(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) GetExternalToolPause(id string) (ExternalToolPause, bool) {
	row := p.db.QueryRow(`select `+externalToolPauseColumns+` from external_tool_pause where id = $1`, strings.TrimSpace(id))
	item, err := scanExternalToolPause(row)
	return item, err == nil
}

func (p *PostgresStore) GetExternalToolPauseByDBReadRequestID(requestID string) (ExternalToolPause, bool) {
	row := p.db.QueryRow(`select `+externalToolPauseColumns+` from external_tool_pause where db_read_request_id = $1 order by created_at desc limit 1`, strings.TrimSpace(requestID))
	item, err := scanExternalToolPause(row)
	return item, err == nil
}

func (p *PostgresStore) UpsertExternalToolPause(input ExternalToolPauseCreateInput, now time.Time) (ExternalToolPause, bool, error) {
	item, err := NewExternalToolPause(input, now)
	if err != nil {
		return ExternalToolPause{}, false, err
	}
	var created ExternalToolPause
	var wasCreated bool
	err = p.withTx(func(tx *sql.Tx) error {
		if err := advisoryLock(tx, "external-tool-pause:"+item.IdempotencyKey); err != nil {
			return err
		}
		existing, scanErr := scanExternalToolPause(tx.QueryRow(`select `+externalToolPauseColumns+` from external_tool_pause where idempotency_key = $1`, item.IdempotencyKey))
		if scanErr == nil {
			created = existing
			wasCreated = false
			return nil
		}
		if scanErr != sql.ErrNoRows {
			return scanErr
		}
		if err := insertExternalToolPauseTx(tx, item); err != nil {
			return err
		}
		created = item
		wasCreated = true
		return nil
	})
	if err != nil {
		return ExternalToolPause{}, false, err
	}
	return cloneExternalToolPause(created), wasCreated, nil
}

func (p *PostgresStore) UpdateExternalToolPause(id string, mutate func(*ExternalToolPause) error) (ExternalToolPause, error) {
	var updated ExternalToolPause
	err := p.withTx(func(tx *sql.Tx) error {
		item, err := scanExternalToolPause(tx.QueryRow(`select `+externalToolPauseColumns+` from external_tool_pause where id = $1 for update`, strings.TrimSpace(id)))
		if err != nil {
			return err
		}
		if mutate != nil {
			if err := mutate(&item); err != nil {
				return err
			}
		}
		item.UpdatedAt = time.Now().UTC()
		if err := updateExternalToolPauseTx(tx, item); err != nil {
			return err
		}
		updated = item
		return nil
	})
	if err != nil {
		return ExternalToolPause{}, err
	}
	return cloneExternalToolPause(updated), nil
}

func insertExternalToolPauseTx(tx *sql.Tx, item ExternalToolPause) error {
	_, err := tx.Exec(`insert into external_tool_pause (`+externalToolPauseColumns+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21::jsonb,$22::jsonb,$23::jsonb,$24,$25::jsonb,$26,$27)`,
		item.ID, item.IdempotencyKey, nullString(item.ConversationID), item.WorkflowID, nullString(item.TraceID), nullString(item.OperationID), nullString(item.ExecutionID),
		item.HermesSessionID, nullString(item.CanonicalToolName), item.TransportToolName, item.ToolCallID, nullString(item.ArgsHash), nullString(item.DBReadRequestID), nullString(item.SQLSHA256),
		string(item.ApprovalStatus), string(item.ToolOutcome), string(item.ResumeStatus), nullString(item.ApprovalRef), nullString(item.ResultRef), nullTime(&item.ExpiresAt),
		jsonString(item.PendingAssistantMessage), jsonString(item.TranscriptSnapshot), jsonString(item.ResumePayload), nullString(item.ErrorMessage), jsonString(item.Metadata),
		item.CreatedAt, item.UpdatedAt,
	)
	return err
}

func updateExternalToolPauseTx(tx *sql.Tx, item ExternalToolPause) error {
	_, err := tx.Exec(`update external_tool_pause set conversation_id=$2, workflow_id=$3, trace_id=$4, operation_id=$5, execution_id=$6, hermes_session_id=$7, canonical_tool_name=$8, transport_tool_name=$9, tool_call_id=$10, args_hash=$11, db_read_request_id=$12, sql_sha256=$13, approval_status=$14, tool_outcome=$15, resume_status=$16, approval_ref=$17, result_ref=$18, expires_at=$19, pending_assistant_message=$20::jsonb, transcript_snapshot=$21::jsonb, resume_payload=$22::jsonb, error_message=$23, metadata=$24::jsonb, updated_at=$25 where id=$1`,
		item.ID, nullString(item.ConversationID), item.WorkflowID, nullString(item.TraceID), nullString(item.OperationID), nullString(item.ExecutionID),
		item.HermesSessionID, nullString(item.CanonicalToolName), item.TransportToolName, item.ToolCallID, nullString(item.ArgsHash), nullString(item.DBReadRequestID), nullString(item.SQLSHA256),
		string(item.ApprovalStatus), string(item.ToolOutcome), string(item.ResumeStatus), nullString(item.ApprovalRef), nullString(item.ResultRef), nullTime(&item.ExpiresAt),
		jsonString(item.PendingAssistantMessage), jsonString(item.TranscriptSnapshot), jsonString(item.ResumePayload), nullString(item.ErrorMessage), jsonString(item.Metadata), item.UpdatedAt,
	)
	return err
}

type externalToolPauseScanner interface{ Scan(dest ...any) error }

func scanExternalToolPause(row externalToolPauseScanner) (ExternalToolPause, error) {
	var item ExternalToolPause
	var conversationID, traceID, operationID, executionID, canonicalToolName, argsHash, dbReadRequestID, sqlSHA256, approvalRef, resultRef, errorMessage sql.NullString
	var expiresAt sql.NullTime
	var pendingRaw, transcriptRaw, resumeRaw, metadataRaw []byte
	if err := row.Scan(
		&item.ID, &item.IdempotencyKey, &conversationID, &item.WorkflowID, &traceID, &operationID, &executionID, &item.HermesSessionID,
		&canonicalToolName, &item.TransportToolName, &item.ToolCallID, &argsHash, &dbReadRequestID, &sqlSHA256,
		&item.ApprovalStatus, &item.ToolOutcome, &item.ResumeStatus, &approvalRef, &resultRef, &expiresAt,
		&pendingRaw, &transcriptRaw, &resumeRaw, &errorMessage, &metadataRaw, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return ExternalToolPause{}, err
	}
	item.ConversationID = conversationID.String
	item.TraceID = traceID.String
	item.OperationID = operationID.String
	item.ExecutionID = executionID.String
	item.CanonicalToolName = canonicalToolName.String
	item.ArgsHash = argsHash.String
	item.DBReadRequestID = dbReadRequestID.String
	item.SQLSHA256 = sqlSHA256.String
	item.ApprovalRef = approvalRef.String
	item.ResultRef = resultRef.String
	if expiresAt.Valid {
		item.ExpiresAt = expiresAt.Time
	}
	item.PendingAssistantMessage = decodeJSON(pendingRaw, map[string]any{})
	item.TranscriptSnapshot = decodeJSON(transcriptRaw, []map[string]any{})
	item.ResumePayload = decodeJSON(resumeRaw, map[string]any{})
	item.ErrorMessage = errorMessage.String
	item.Metadata = decodeJSON(metadataRaw, map[string]any{})
	return cloneExternalToolPause(item), nil
}
