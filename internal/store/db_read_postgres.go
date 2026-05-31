package store

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const dbReadRequestColumns = `id, idempotency_key, target, purpose, sql_text, sql_sha256, execution_scope_key, requester, conversation_id, workflow_id, trace_id, channel_id, thread_ts, state, current_validation_attempt_id, approved_by_slack_user_id, approved_at, expires_at, caps, redaction, slack_message_channel_id, slack_message_ts, lease_holder, lease_token, lease_generation, lease_expires_at, result_artifact_ref, result_sample, row_count, truncated, error_message, metadata, created_at, updated_at`

func (p *PostgresStore) ListDBReadRequests() []DBReadRequest {
	rows, err := p.db.Query(`select ` + dbReadRequestColumns + ` from db_read_request order by created_at desc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []DBReadRequest{}
	for rows.Next() {
		item, err := scanDBReadRequest(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) ListDBReadRequestsByScope(conversationID string, workflowID string, traceID string, channelID string, threadTS string, notBefore time.Time) []DBReadRequest {
	conversationID = strings.TrimSpace(conversationID)
	workflowID = strings.TrimSpace(workflowID)
	traceID = strings.TrimSpace(traceID)
	channelID = strings.TrimSpace(channelID)
	threadTS = strings.TrimSpace(threadTS)
	var conditions []string
	var args []any
	if workflowID != "" {
		args = append(args, workflowID)
		conditions = append(conditions, "workflow_id = $"+strconv.Itoa(len(args)))
	}
	if traceID != "" {
		args = append(args, traceID)
		conditions = append(conditions, "trace_id = $"+strconv.Itoa(len(args)))
	}
	if channelID != "" && threadTS != "" {
		args = append(args, channelID, threadTS)
		conditions = append(conditions, "(channel_id = $"+strconv.Itoa(len(args)-1)+" and thread_ts = $"+strconv.Itoa(len(args))+")")
	}
	if conversationID != "" {
		args = append(args, conversationID)
		conditions = append(conditions, "conversation_id = $"+strconv.Itoa(len(args)))
	}
	if len(conditions) == 0 {
		return []DBReadRequest{}
	}
	query := `select ` + dbReadRequestColumns + ` from db_read_request where (`
	query += strings.Join(conditions, " or ")
	query += `)`
	if !notBefore.IsZero() {
		args = append(args, notBefore)
		query += ` and created_at >= $` + strconv.Itoa(len(args))
	}
	query += ` order by created_at desc`
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []DBReadRequest{}
	for rows.Next() {
		item, err := scanDBReadRequest(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) GetDBReadRequest(requestID string) (DBReadRequest, bool) {
	row := p.db.QueryRow(`select `+dbReadRequestColumns+` from db_read_request where id = $1`, requestID)
	item, err := scanDBReadRequest(row)
	return item, err == nil
}

func (p *PostgresStore) GetDBReadRequestByIdempotencyKey(key string) (DBReadRequest, bool) {
	row := p.db.QueryRow(`select `+dbReadRequestColumns+` from db_read_request where idempotency_key = $1`, key)
	item, err := scanDBReadRequest(row)
	return item, err == nil
}

func (p *PostgresStore) UpsertDBReadRequest(input DBReadCreateInput, now time.Time) (DBReadRequest, bool, error) {
	if existing, ok := p.GetDBReadRequestByIdempotencyKey(input.IdempotencyKey); ok {
		return existing, false, nil
	}
	item, err := NewDBReadRequest(input, now)
	if err != nil {
		return DBReadRequest{}, false, err
	}
	_, err = p.db.Exec(`insert into db_read_request (`+dbReadRequestColumns+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19::jsonb,$20::jsonb,$21,$22,$23,$24,$25,$26,$27,$28::jsonb,$29,$30,$31,$32::jsonb,$33,$34)`,
		item.ID, item.IdempotencyKey, item.Target, item.Purpose, item.SQL, item.SQLSHA256, nullString(item.ExecutionScopeKey), item.Requester,
		nullString(item.ConversationID), nullString(item.WorkflowID), nullString(item.TraceID), nullString(item.ChannelID), nullString(item.ThreadTS),
		string(item.State), nullString(item.CurrentValidationAttemptID), nullString(item.ApprovedBySlackUserID), nullTime(item.ApprovedAt), item.ExpiresAt,
		jsonString(item.Caps), jsonString(item.Redaction), nullString(item.SlackMessageChannelID), nullString(item.SlackMessageTS),
		nullString(item.LeaseHolder), nullString(item.LeaseToken), item.LeaseGeneration, nullTime(item.LeaseExpiresAt),
		nullString(item.ResultArtifactRef), jsonString(item.ResultSample), item.RowCount, item.Truncated, nullString(item.ErrorMessage), jsonString(item.Metadata),
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		if existing, ok := p.GetDBReadRequestByIdempotencyKey(input.IdempotencyKey); ok {
			return existing, false, nil
		}
		return DBReadRequest{}, false, err
	}
	return item, true, nil
}

func (p *PostgresStore) AppendDBReadValidationAttempt(attempt DBReadValidationAttempt) (DBReadValidationAttempt, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return DBReadValidationAttempt{}, err
	}
	defer tx.Rollback()
	request, err := scanDBReadRequest(tx.QueryRow(`select `+dbReadRequestColumns+` from db_read_request where id = $1 for update`, attempt.RequestID))
	if err != nil {
		return DBReadValidationAttempt{}, err
	}
	if attempt.ID == "" {
		attempt.ID = "dbreadval_" + uuid.NewString()
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = time.Now().UTC()
	}
	if _, err := tx.Exec(`insert into db_read_validation_attempt (id, request_id, target, sql_sha256, status, stage, error_code, error_message, details, validated_at, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11)`,
		attempt.ID, attempt.RequestID, attempt.Target, attempt.SQLSHA256, string(attempt.Status), attempt.Stage, nullString(attempt.ErrorCode), nullString(attempt.ErrorMessage), jsonString(attempt.Details), nullTime(attempt.ValidatedAt), attempt.CreatedAt,
	); err != nil {
		return DBReadValidationAttempt{}, err
	}
	nextState := DBReadStateValidationFailed
	if attempt.Status == DBReadValidationStatusSucceeded {
		nextState = DBReadStatePendingApproval
	}
	if err := ValidateDBReadStateTransition(request.State, nextState); err != nil {
		return DBReadValidationAttempt{}, err
	}
	if _, err := tx.Exec(`update db_read_request set state = $2, current_validation_attempt_id = $3, error_message = $4, lease_holder = null, lease_token = null, lease_expires_at = null, updated_at = $5 where id = $1`,
		request.ID, string(nextState), attempt.ID, nullString(attempt.ErrorMessage), attempt.CreatedAt,
	); err != nil {
		return DBReadValidationAttempt{}, err
	}
	if err := tx.Commit(); err != nil {
		return DBReadValidationAttempt{}, err
	}
	return cloneDBReadValidationAttempt(attempt), nil
}

func (p *PostgresStore) ListDBReadValidationAttempts(requestID string) []DBReadValidationAttempt {
	rows, err := p.db.Query(`select id, request_id, target, sql_sha256, status, stage, error_code, error_message, details, validated_at, created_at from db_read_validation_attempt where request_id = $1 order by created_at asc`, requestID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []DBReadValidationAttempt{}
	for rows.Next() {
		item, err := scanDBReadValidationAttempt(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) TransitionDBReadRequest(requestID string, from DBReadState, to DBReadState, mutate func(*DBReadRequest) error) (DBReadRequest, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return DBReadRequest{}, err
	}
	defer tx.Rollback()
	item, err := scanDBReadRequest(tx.QueryRow(`select `+dbReadRequestColumns+` from db_read_request where id = $1 for update`, requestID))
	if err != nil {
		return DBReadRequest{}, err
	}
	if item.State != from {
		return DBReadRequest{}, errors.New("db read request state mismatch")
	}
	if err := ValidateDBReadStateTransition(from, to); err != nil {
		return DBReadRequest{}, err
	}
	if mutate != nil {
		if err := mutate(&item); err != nil {
			return DBReadRequest{}, err
		}
	}
	item.State = to
	item.UpdatedAt = time.Now().UTC()
	if err := updateDBReadRequestTx(tx, item); err != nil {
		return DBReadRequest{}, err
	}
	if err := tx.Commit(); err != nil {
		return DBReadRequest{}, err
	}
	return cloneDBReadRequest(item), nil
}

func (p *PostgresStore) ClaimNextDBReadValidationRequest(holder string, lease time.Duration, now time.Time, targets []string) (DBReadLease, bool, error) {
	return p.claimNextDBReadRequestWithState(holder, lease, now, targets, DBReadStateValidating)
}

func (p *PostgresStore) ClaimNextDBReadRequest(holder string, lease time.Duration, now time.Time, targets []string) (DBReadLease, bool, error) {
	return p.claimNextDBReadRequestWithState(holder, lease, now, targets, DBReadStateApproved)
}

func (p *PostgresStore) claimNextDBReadRequestWithState(holder string, lease time.Duration, now time.Time, targets []string, state DBReadState) (DBReadLease, bool, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tx, err := p.db.Begin()
	if err != nil {
		return DBReadLease{}, false, err
	}
	defer tx.Rollback()
	args := []any{string(state), now}
	query := `select ` + dbReadRequestColumns + ` from db_read_request where state = $1 and (lease_expires_at is null or lease_expires_at <= $2)`
	targets = compactDBReadTargets(targets)
	if len(targets) > 0 {
		placeholders := make([]string, 0, len(targets))
		for _, target := range targets {
			args = append(args, target)
			placeholders = append(placeholders, "$"+strconv.Itoa(len(args)))
		}
		query += ` and target in (` + strings.Join(placeholders, ",") + `)`
	}
	query += ` order by approved_at asc nulls last, created_at asc for update skip locked limit 1`
	row := tx.QueryRow(query, args...)
	item, err := scanDBReadRequest(row)
	if err == sql.ErrNoRows {
		return DBReadLease{}, false, nil
	}
	if err != nil {
		return DBReadLease{}, false, err
	}
	token := "dbreadlease_" + uuid.NewString()
	expires := now.Add(lease)
	item.LeaseHolder = holder
	item.LeaseToken = token
	item.LeaseGeneration++
	item.LeaseExpiresAt = &expires
	item.UpdatedAt = now
	if err := updateDBReadRequestTx(tx, item); err != nil {
		return DBReadLease{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return DBReadLease{}, false, err
	}
	return DBReadLease{Request: cloneDBReadRequest(item), Token: token}, true, nil
}

func (p *PostgresStore) AppendDBReadExecutionResult(result DBReadExecutionResult) (DBReadExecutionResult, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return DBReadExecutionResult{}, err
	}
	defer tx.Rollback()
	request, err := scanDBReadRequest(tx.QueryRow(`select `+dbReadRequestColumns+` from db_read_request where id = $1 for update`, result.RequestID))
	if err != nil {
		return DBReadExecutionResult{}, err
	}
	if request.LeaseToken != "" && result.LeaseToken != "" && request.LeaseToken != result.LeaseToken {
		return DBReadExecutionResult{}, errors.New("db read lease token mismatch")
	}
	if result.ID == "" {
		result.ID = "dbreadexec_" + uuid.NewString()
	}
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now().UTC()
	}
	if _, err := tx.Exec(`insert into db_read_execution_result (id, request_id, lease_token, status, row_count, truncated, sample, artifact_ref, error_code, error_message, created_at) values ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9,$10,$11)`,
		result.ID, result.RequestID, nullString(result.LeaseToken), string(result.Status), result.RowCount, result.Truncated, jsonString(result.Sample), nullString(result.ArtifactRef), nullString(result.ErrorCode), nullString(result.ErrorMessage), result.CreatedAt,
	); err != nil {
		return DBReadExecutionResult{}, err
	}
	nextState := DBReadStateFailed
	if result.Status == DBReadExecutionStatusSucceeded {
		nextState = DBReadStateSucceeded
	}
	if request.State == DBReadStateApproved {
		if err := ValidateDBReadStateTransition(request.State, DBReadStateExecuting); err != nil {
			return DBReadExecutionResult{}, err
		}
		request.State = DBReadStateExecuting
	}
	if err := ValidateDBReadStateTransition(request.State, nextState); err != nil {
		return DBReadExecutionResult{}, err
	}
	request.State = nextState
	request.RowCount = result.RowCount
	request.Truncated = result.Truncated
	request.ResultSample = cloneDBReadSample(result.Sample)
	request.ResultArtifactRef = result.ArtifactRef
	request.ErrorMessage = result.ErrorMessage
	request.LeaseHolder = ""
	request.LeaseToken = ""
	request.LeaseExpiresAt = nil
	request.UpdatedAt = result.CreatedAt
	if err := updateDBReadRequestTx(tx, request); err != nil {
		return DBReadExecutionResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return DBReadExecutionResult{}, err
	}
	return cloneDBReadExecutionResult(result), nil
}

func (p *PostgresStore) ListDBReadExecutionResults(requestID string) []DBReadExecutionResult {
	rows, err := p.db.Query(`select id, request_id, lease_token, status, row_count, truncated, sample, artifact_ref, error_code, error_message, created_at from db_read_execution_result where request_id = $1 order by created_at asc`, requestID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []DBReadExecutionResult{}
	for rows.Next() {
		item, err := scanDBReadExecutionResult(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) ExpirePendingDBReadRequests(now time.Time) ([]DBReadRequest, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	rows, err := p.db.Query(`update db_read_request set state = $1, updated_at = $2, lease_holder = null, lease_token = null, lease_expires_at = null where state in ($3,$4,$5,$6) and expires_at <= $2 and (lease_expires_at is null or lease_expires_at <= $2) returning `+dbReadRequestColumns,
		string(DBReadStateExpired), now, string(DBReadStateValidating), string(DBReadStateValidationFailed), string(DBReadStatePendingApproval), string(DBReadStateApproved))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DBReadRequest{}
	for rows.Next() {
		item, err := scanDBReadRequest(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out, nil
}

func updateDBReadRequestTx(tx *sql.Tx, item DBReadRequest) error {
	_, err := tx.Exec(`update db_read_request set target=$2, purpose=$3, sql_text=$4, sql_sha256=$5, execution_scope_key=$6, requester=$7, conversation_id=$8, workflow_id=$9, trace_id=$10, channel_id=$11, thread_ts=$12, state=$13, current_validation_attempt_id=$14, approved_by_slack_user_id=$15, approved_at=$16, expires_at=$17, caps=$18::jsonb, redaction=$19::jsonb, slack_message_channel_id=$20, slack_message_ts=$21, lease_holder=$22, lease_token=$23, lease_generation=$24, lease_expires_at=$25, result_artifact_ref=$26, result_sample=$27::jsonb, row_count=$28, truncated=$29, error_message=$30, metadata=$31::jsonb, updated_at=$32 where id=$1`,
		item.ID, item.Target, item.Purpose, item.SQL, item.SQLSHA256, nullString(item.ExecutionScopeKey), item.Requester,
		nullString(item.ConversationID), nullString(item.WorkflowID), nullString(item.TraceID), nullString(item.ChannelID), nullString(item.ThreadTS),
		string(item.State), nullString(item.CurrentValidationAttemptID), nullString(item.ApprovedBySlackUserID), nullTime(item.ApprovedAt), item.ExpiresAt,
		jsonString(item.Caps), jsonString(item.Redaction), nullString(item.SlackMessageChannelID), nullString(item.SlackMessageTS),
		nullString(item.LeaseHolder), nullString(item.LeaseToken), item.LeaseGeneration, nullTime(item.LeaseExpiresAt),
		nullString(item.ResultArtifactRef), jsonString(item.ResultSample), item.RowCount, item.Truncated, nullString(item.ErrorMessage), jsonString(item.Metadata), item.UpdatedAt)
	return err
}

type dbReadScanner interface{ Scan(dest ...any) error }

func scanDBReadRequest(row dbReadScanner) (DBReadRequest, error) {
	var item DBReadRequest
	var executionScopeKey, conversationID, workflowID, traceID, channelID, threadTS, validationAttemptID, approvedBy, slackChannel, slackTS, leaseHolder, leaseToken, artifactRef, errorMessage sql.NullString
	var approvedAt, leaseExpiresAt sql.NullTime
	var capsRaw, redactionRaw, sampleRaw, metadataRaw []byte
	if err := row.Scan(&item.ID, &item.IdempotencyKey, &item.Target, &item.Purpose, &item.SQL, &item.SQLSHA256, &executionScopeKey, &item.Requester, &conversationID, &workflowID, &traceID, &channelID, &threadTS, &item.State, &validationAttemptID, &approvedBy, &approvedAt, &item.ExpiresAt, &capsRaw, &redactionRaw, &slackChannel, &slackTS, &leaseHolder, &leaseToken, &item.LeaseGeneration, &leaseExpiresAt, &artifactRef, &sampleRaw, &item.RowCount, &item.Truncated, &errorMessage, &metadataRaw, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return DBReadRequest{}, err
	}
	item.ExecutionScopeKey = executionScopeKey.String
	item.ConversationID = conversationID.String
	item.WorkflowID = workflowID.String
	item.TraceID = traceID.String
	item.ChannelID = channelID.String
	item.ThreadTS = threadTS.String
	item.CurrentValidationAttemptID = validationAttemptID.String
	item.ApprovedBySlackUserID = approvedBy.String
	if approvedAt.Valid {
		item.ApprovedAt = &approvedAt.Time
	}
	item.Caps = decodeJSON(capsRaw, DBReadCaps{})
	item.Redaction = decodeJSON(redactionRaw, DBReadRedactionPolicy{})
	item.SlackMessageChannelID = slackChannel.String
	item.SlackMessageTS = slackTS.String
	item.LeaseHolder = leaseHolder.String
	item.LeaseToken = leaseToken.String
	if leaseExpiresAt.Valid {
		item.LeaseExpiresAt = &leaseExpiresAt.Time
	}
	item.ResultArtifactRef = artifactRef.String
	item.ResultSample = decodeJSON(sampleRaw, []map[string]string{})
	item.ErrorMessage = errorMessage.String
	item.Metadata = decodeJSON(metadataRaw, map[string]any{})
	return cloneDBReadRequest(item), nil
}

func scanDBReadValidationAttempt(row dbReadScanner) (DBReadValidationAttempt, error) {
	var item DBReadValidationAttempt
	var errorCode, errorMessage sql.NullString
	var detailsRaw []byte
	var validatedAt sql.NullTime
	if err := row.Scan(&item.ID, &item.RequestID, &item.Target, &item.SQLSHA256, &item.Status, &item.Stage, &errorCode, &errorMessage, &detailsRaw, &validatedAt, &item.CreatedAt); err != nil {
		return DBReadValidationAttempt{}, err
	}
	item.ErrorCode = errorCode.String
	item.ErrorMessage = errorMessage.String
	item.Details = decodeJSON(detailsRaw, map[string]any{})
	if validatedAt.Valid {
		item.ValidatedAt = &validatedAt.Time
	}
	return cloneDBReadValidationAttempt(item), nil
}

func scanDBReadExecutionResult(row dbReadScanner) (DBReadExecutionResult, error) {
	var item DBReadExecutionResult
	var leaseToken, artifactRef, errorCode, errorMessage sql.NullString
	var sampleRaw []byte
	if err := row.Scan(&item.ID, &item.RequestID, &leaseToken, &item.Status, &item.RowCount, &item.Truncated, &sampleRaw, &artifactRef, &errorCode, &errorMessage, &item.CreatedAt); err != nil {
		return DBReadExecutionResult{}, err
	}
	item.LeaseToken = leaseToken.String
	item.Sample = decodeJSON(sampleRaw, []map[string]string{})
	item.ArtifactRef = artifactRef.String
	item.ErrorCode = errorCode.String
	item.ErrorMessage = errorMessage.String
	return cloneDBReadExecutionResult(item), nil
}

func compactDBReadTargets(targets []string) []string {
	out := make([]string, 0, len(targets))
	seen := map[string]struct{}{}
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		out = append(out, target)
	}
	return out
}
