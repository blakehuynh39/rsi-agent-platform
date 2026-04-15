package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
)

func (p *PostgresStore) ListOperations() []operation.Execution {
	rows, err := p.db.Query(`select ` + operationSelectColumns() + ` from operation_execution order by updated_at desc, id asc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []operation.Execution{}
	for rows.Next() {
		item, err := scanOperation(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) GetOperation(operationID string) (operation.Execution, bool) {
	row := p.db.QueryRow(`select `+operationSelectColumns()+` from operation_execution where id = $1`, strings.TrimSpace(operationID))
	item, err := scanOperation(row)
	if err != nil {
		return operation.Execution{}, false
	}
	return item, true
}

func (p *PostgresStore) ListOperationsByScope(scopeKind operation.ScopeKind, scopeID string) []operation.Execution {
	rows, err := p.db.Query(`select `+operationSelectColumns()+` from operation_execution where scope_kind = $1 and scope_id = $2 order by updated_at desc, id asc`, string(scopeKind), strings.TrimSpace(scopeID))
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []operation.Execution{}
	for rows.Next() {
		item, err := scanOperation(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) GetOrCreateOperation(item operation.Execution) (created operation.Execution, wasCreated bool, err error) {
	item, err = normalizeOperationExecution(item)
	if err != nil {
		return operation.Execution{}, false, err
	}
	err = p.withTx(func(tx *sql.Tx) error {
		created, wasCreated, err = getOrCreateOperationTx(tx, item)
		return err
	})
	return
}

func (p *PostgresStore) ClaimOperation(operationID string, holder string) (item operation.Execution, claimed bool, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		item, claimed, err = p.claimOperationTx(tx, operationID, holder)
		return err
	})
	return
}

func (p *PostgresStore) CompleteOperation(operationID string, resultRef string) (item operation.Execution, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		item, err = completeOperationTx(tx, operationID, resultRef)
		return err
	})
	return
}

func (p *PostgresStore) FailOperation(operationID string, lastError string) (item operation.Execution, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		item, err = failOperationTx(tx, operationID, lastError)
		return err
	})
	return
}

func loadOperations(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select ` + operationSelectColumns() + ` from operation_execution order by updated_at desc, id asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		item, err := scanOperation(rows)
		if err != nil {
			return err
		}
		store.operations[item.ID] = item
	}
	return rows.Err()
}

func persistOperations(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.operations)
	for _, key := range keys {
		item := store.operations[key]
		if _, err := tx.Exec(`
			insert into operation_execution (
				id, scope_kind, scope_id, operation_kind, operation_key, status, queue, requested_by, holder, trace_id, proposal_id, attempt_id, payload_hash, result_ref, last_error, retry_count, created_at, updated_at, started_at, completed_at
			) values (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
			)
			on conflict (id) do update set
				scope_kind = excluded.scope_kind,
				scope_id = excluded.scope_id,
				operation_kind = excluded.operation_kind,
				operation_key = excluded.operation_key,
				status = excluded.status,
				queue = excluded.queue,
				requested_by = excluded.requested_by,
				holder = excluded.holder,
				trace_id = excluded.trace_id,
				proposal_id = excluded.proposal_id,
				attempt_id = excluded.attempt_id,
				payload_hash = excluded.payload_hash,
				result_ref = excluded.result_ref,
				last_error = excluded.last_error,
				retry_count = excluded.retry_count,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at,
				started_at = excluded.started_at,
				completed_at = excluded.completed_at
		`,
			item.ID,
			string(item.ScopeKind),
			item.ScopeID,
			item.OperationKind,
			item.OperationKey,
			string(item.Status),
			string(item.Queue),
			item.RequestedBy,
			item.Holder,
			item.TraceID,
			item.ProposalID,
			item.AttemptID,
			item.PayloadHash,
			item.ResultRef,
			item.LastError,
			item.RetryCount,
			item.CreatedAt,
			item.UpdatedAt,
			nullTime(item.StartedAt),
			nullTime(item.CompletedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func selectOperationByNaturalKeyTx(tx *sql.Tx, scopeKind operation.ScopeKind, scopeID string, operationKind string, operationKey string) (operation.Execution, error) {
	row := tx.QueryRow(`select `+operationSelectColumns()+` from operation_execution where scope_kind = $1 and scope_id = $2 and operation_kind = $3 and operation_key = $4`, string(scopeKind), strings.TrimSpace(scopeID), strings.TrimSpace(operationKind), strings.TrimSpace(operationKey))
	return scanOperation(row)
}

func selectOperationByIDTx(tx *sql.Tx, operationID string) (operation.Execution, error) {
	row := tx.QueryRow(`select `+operationSelectColumns()+` from operation_execution where id = $1`, strings.TrimSpace(operationID))
	return scanOperation(row)
}

func getOrCreateOperationTx(tx *sql.Tx, item operation.Execution) (created operation.Execution, wasCreated bool, err error) {
	if err := advisoryLock(tx, fmt.Sprintf("operation:%s:%s:%s:%s", item.ScopeKind, item.ScopeID, item.OperationKind, item.OperationKey)); err != nil {
		return operation.Execution{}, false, err
	}
	existing, err := selectOperationByNaturalKeyTx(tx, item.ScopeKind, item.ScopeID, item.OperationKind, item.OperationKey)
	if err == nil {
		return existing, false, nil
	}
	if err != sql.ErrNoRows {
		return operation.Execution{}, false, err
	}
	row := tx.QueryRow(`
		insert into operation_execution (
			id, scope_kind, scope_id, operation_kind, operation_key, status, queue, requested_by, holder, trace_id, proposal_id, attempt_id, payload_hash, result_ref, last_error, retry_count, created_at, updated_at, started_at, completed_at
		) values (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
		)
		returning `+operationSelectColumns(),
		item.ID,
		string(item.ScopeKind),
		item.ScopeID,
		item.OperationKind,
		item.OperationKey,
		string(item.Status),
		string(item.Queue),
		item.RequestedBy,
		item.Holder,
		item.TraceID,
		item.ProposalID,
		item.AttemptID,
		item.PayloadHash,
		item.ResultRef,
		item.LastError,
		item.RetryCount,
		item.CreatedAt,
		item.UpdatedAt,
		nullTime(item.StartedAt),
		nullTime(item.CompletedAt),
	)
	created, err = scanOperation(row)
	if err != nil {
		return operation.Execution{}, false, fmt.Errorf("insert operation %s/%s/%s id=%s: %w", item.ScopeKind, item.ScopeID, item.OperationKind, item.ID, err)
	}
	return created, true, nil
}

func (p *PostgresStore) claimOperationTx(tx *sql.Tx, operationID string, holder string) (item operation.Execution, claimed bool, err error) {
	now := time.Now().UTC()
	row := tx.QueryRow(`
		update operation_execution
		set status = $2,
			holder = $3,
			started_at = coalesce(started_at, $4),
			updated_at = $4,
			completed_at = null
		where id = $1
		  and status in ($5, $6)
		returning `+operationSelectColumns(),
		strings.TrimSpace(operationID),
		string(operation.StatusRunning),
		strings.TrimSpace(holder),
		now,
		string(operation.StatusQueued),
		string(operation.StatusFailed),
	)
	item, err = scanOperation(row)
	if err == sql.ErrNoRows {
		item, err = selectOperationByIDTx(tx, operationID)
		return item, false, err
	}
	if err != nil {
		return operation.Execution{}, false, err
	}
	return item, true, nil
}

func completeOperationTx(tx *sql.Tx, operationID string, resultRef string) (operation.Execution, error) {
	now := time.Now().UTC()
	row := tx.QueryRow(`
		update operation_execution
		set status = $2,
			result_ref = $3,
			last_error = '',
			updated_at = $4,
			completed_at = $4
		where id = $1
		  and status not in ($5, $6, $7)
		returning `+operationSelectColumns(),
		strings.TrimSpace(operationID),
		string(operation.StatusCompleted),
		strings.TrimSpace(resultRef),
		now,
		string(operation.StatusCompleted),
		string(operation.StatusCanceled),
		string(operation.StatusSuperseded),
	)
	item, err := scanOperation(row)
	if err == sql.ErrNoRows {
		return selectOperationByIDTx(tx, operationID)
	}
	return item, err
}

func failOperationTx(tx *sql.Tx, operationID string, lastError string) (operation.Execution, error) {
	now := time.Now().UTC()
	row := tx.QueryRow(`
		update operation_execution
		set status = $2,
			last_error = $3,
			retry_count = retry_count + 1,
			updated_at = $4,
			completed_at = $4
		where id = $1
		  and status not in ($5, $6, $7)
		returning `+operationSelectColumns(),
		strings.TrimSpace(operationID),
		string(operation.StatusFailed),
		strings.TrimSpace(lastError),
		now,
		string(operation.StatusCompleted),
		string(operation.StatusCanceled),
		string(operation.StatusSuperseded),
	)
	item, err := scanOperation(row)
	if err == sql.ErrNoRows {
		return selectOperationByIDTx(tx, operationID)
	}
	return item, err
}

func requeueOperationTx(tx *sql.Tx, operationID string, lastError string) (operation.Execution, error) {
	now := time.Now().UTC()
	row := tx.QueryRow(`
		update operation_execution
		set status = $2,
			holder = '',
			last_error = $3,
			updated_at = $4,
			completed_at = null
		where id = $1
		  and status not in ($5, $6, $7)
		returning `+operationSelectColumns(),
		strings.TrimSpace(operationID),
		string(operation.StatusQueued),
		strings.TrimSpace(lastError),
		now,
		string(operation.StatusCompleted),
		string(operation.StatusCanceled),
		string(operation.StatusSuperseded),
	)
	item, err := scanOperation(row)
	if err == sql.ErrNoRows {
		return selectOperationByIDTx(tx, operationID)
	}
	return item, err
}

func operationSelectColumns() string {
	return `id, scope_kind, scope_id, operation_kind, operation_key, status, queue, requested_by, holder, trace_id, proposal_id, attempt_id, payload_hash, result_ref, last_error, retry_count, created_at, updated_at, started_at, completed_at`
}

type operationScanner interface {
	Scan(dest ...any) error
}

func scanOperation(scanner operationScanner) (operation.Execution, error) {
	var (
		item                   operation.Execution
		queueName, statusValue string
		startedAt              sql.NullTime
		completedAt            sql.NullTime
	)
	err := scanner.Scan(
		&item.ID,
		&item.ScopeKind,
		&item.ScopeID,
		&item.OperationKind,
		&item.OperationKey,
		&statusValue,
		&queueName,
		&item.RequestedBy,
		&item.Holder,
		&item.TraceID,
		&item.ProposalID,
		&item.AttemptID,
		&item.PayloadHash,
		&item.ResultRef,
		&item.LastError,
		&item.RetryCount,
		&item.CreatedAt,
		&item.UpdatedAt,
		&startedAt,
		&completedAt,
	)
	if err != nil {
		return operation.Execution{}, err
	}
	item.Queue = queue.QueueName(queueName)
	item.Status = operation.Status(statusValue)
	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}
	return normalizeOperationExecution(item)
}
