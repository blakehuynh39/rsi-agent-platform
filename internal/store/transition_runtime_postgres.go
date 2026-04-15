package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func (p *PostgresStore) ListDomainEvents() []transition.DomainEvent {
	rows, err := p.db.Query(`select id, machine_kind, aggregate_id, aggregate_version, event_kind, command_id, causation_id, payload, created_at from domain_event order by created_at desc, id asc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []transition.DomainEvent{}
	for rows.Next() {
		item, err := scanDomainEvent(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) ListEffectExecutions() []transition.EffectExecution {
	rows, err := p.db.Query(`select ` + effectExecutionSelectColumns() + ` from effect_execution order by updated_at desc, id asc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []transition.EffectExecution{}
	for rows.Next() {
		item, err := scanEffectExecution(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) ListEffectExecutionsByAggregate(machineKind transition.MachineKind, aggregateID string) []transition.EffectExecution {
	rows, err := p.db.Query(`select `+effectExecutionSelectColumns()+` from effect_execution where machine_kind = $1 and aggregate_id = $2 order by updated_at desc, id asc`, string(machineKind), strings.TrimSpace(aggregateID))
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []transition.EffectExecution{}
	for rows.Next() {
		item, err := scanEffectExecution(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) GetCommandReceipt(commandID string) (transition.CommandReceipt, bool) {
	row := p.db.QueryRow(`select command_id, machine_kind, aggregate_id, command_kind, causation_id, actor, decision_kind, reason, aggregate_version, result_ref, created_at, updated_at from command_receipt where command_id = $1`, strings.TrimSpace(commandID))
	item, err := scanCommandReceipt(row)
	if err != nil {
		return transition.CommandReceipt{}, false
	}
	return item, true
}

func (p *PostgresStore) RecordCommandReceipt(item transition.CommandReceipt) (created transition.CommandReceipt, wasCreated bool, err error) {
	item.CommandID = strings.TrimSpace(item.CommandID)
	if item.CommandID == "" {
		return transition.CommandReceipt{}, false, fmt.Errorf("command_id is required")
	}
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = item.CreatedAt
	}
	err = p.withTx(func(tx *sql.Tx) error {
		if err := advisoryLock(tx, "command-receipt:"+item.CommandID); err != nil {
			return err
		}
		existing, scanErr := scanCommandReceipt(tx.QueryRow(`select command_id, machine_kind, aggregate_id, command_kind, causation_id, actor, decision_kind, reason, aggregate_version, result_ref, created_at, updated_at from command_receipt where command_id = $1`, item.CommandID))
		if scanErr == nil {
			created = existing
			wasCreated = false
			return nil
		}
		if scanErr != sql.ErrNoRows {
			return scanErr
		}
		row := tx.QueryRow(`insert into command_receipt (command_id, machine_kind, aggregate_id, command_kind, causation_id, actor, decision_kind, reason, aggregate_version, result_ref, created_at, updated_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			returning command_id, machine_kind, aggregate_id, command_kind, causation_id, actor, decision_kind, reason, aggregate_version, result_ref, created_at, updated_at`,
			item.CommandID,
			string(item.MachineKind),
			item.AggregateID,
			item.CommandKind,
			firstNonEmpty(item.CausationID),
			firstNonEmpty(item.Actor),
			string(item.DecisionKind),
			firstNonEmpty(item.Reason),
			item.AggregateVersion,
			firstNonEmpty(item.ResultRef),
			item.CreatedAt,
			item.UpdatedAt,
		)
		var insertErr error
		created, insertErr = scanCommandReceipt(row)
		wasCreated = insertErr == nil
		return insertErr
	})
	return
}

func (p *PostgresStore) ClaimEffectExecution(effectID string, holder string, lease time.Duration) (item transition.EffectExecution, claimed bool, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		var leaseExpires any
		if lease > 0 {
			expires := now.Add(lease)
			leaseExpires = expires
		}
		row := tx.QueryRow(`
			update effect_execution
			set status = $2,
				holder = $3,
				updated_at = $4,
				started_at = coalesce(started_at, $4),
				lease_expires_at = $5,
				completed_at = null
			where id = $1
			  and (
				status in ($6, $7)
				or (status = $8 and (lease_expires_at is null or lease_expires_at <= $4))
			  )
			returning `+effectExecutionSelectColumns(),
			strings.TrimSpace(effectID),
			string(transition.EffectRunning),
			strings.TrimSpace(holder),
			now,
			leaseExpires,
			string(transition.EffectQueued),
			string(transition.EffectFailed),
			string(transition.EffectRunning),
		)
		item, err = scanEffectExecution(row)
		if err == sql.ErrNoRows {
			row = tx.QueryRow(`select `+effectExecutionSelectColumns()+` from effect_execution where id = $1`, strings.TrimSpace(effectID))
			item, err = scanEffectExecution(row)
			claimed = false
			return err
		}
		claimed = err == nil
		return err
	})
	return
}

func (p *PostgresStore) CompleteEffectExecution(effectID string, resultRef string) (item transition.EffectExecution, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		row := tx.QueryRow(`
			update effect_execution
			set status = $2,
				holder = '',
				result_ref = $3,
				last_error = '',
				updated_at = $4,
				lease_expires_at = null,
				completed_at = $4
			where id = $1
			  and status not in ($5, $6, $7)
			returning `+effectExecutionSelectColumns(),
			strings.TrimSpace(effectID),
			string(transition.EffectCompleted),
			strings.TrimSpace(resultRef),
			now,
			string(transition.EffectCompleted),
			string(transition.EffectCanceled),
			string(transition.EffectSuperseded),
		)
		item, err = scanEffectExecution(row)
		if err == sql.ErrNoRows {
			row = tx.QueryRow(`select `+effectExecutionSelectColumns()+` from effect_execution where id = $1`, strings.TrimSpace(effectID))
			item, err = scanEffectExecution(row)
		}
		return err
	})
	return
}

func (p *PostgresStore) FailEffectExecution(effectID string, lastError string) (item transition.EffectExecution, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		row := tx.QueryRow(`
			update effect_execution
			set status = $2,
				holder = '',
				last_error = $3,
				retry_count = retry_count + 1,
				updated_at = $4,
				lease_expires_at = null,
				completed_at = $4
			where id = $1
			  and status not in ($5, $6, $7)
			returning `+effectExecutionSelectColumns(),
			strings.TrimSpace(effectID),
			string(transition.EffectFailed),
			strings.TrimSpace(lastError),
			now,
			string(transition.EffectCompleted),
			string(transition.EffectCanceled),
			string(transition.EffectSuperseded),
		)
		item, err = scanEffectExecution(row)
		if err == sql.ErrNoRows {
			row = tx.QueryRow(`select `+effectExecutionSelectColumns()+` from effect_execution where id = $1`, strings.TrimSpace(effectID))
			item, err = scanEffectExecution(row)
		}
		return err
	})
	return
}

func effectExecutionSelectColumns() string {
	return `id, machine_kind, aggregate_id, attempt_id, effect_kind, status, holder, idempotency_key, payload, result_ref, last_error, retry_count, created_at, updated_at, started_at, lease_expires_at, completed_at`
}

func scanDomainEvent(scanner rowScanner) (transition.DomainEvent, error) {
	var item transition.DomainEvent
	var machineKind string
	var payload []byte
	if err := scanner.Scan(&item.ID, &machineKind, &item.AggregateID, &item.AggregateVersion, &item.EventKind, &item.CommandID, &item.CausationID, &payload, &item.CreatedAt); err != nil {
		return transition.DomainEvent{}, err
	}
	item.MachineKind = transition.MachineKind(machineKind)
	item.Payload = decodeJSON(payload, map[string]any{})
	return item, nil
}

func scanCommandReceipt(scanner rowScanner) (transition.CommandReceipt, error) {
	var item transition.CommandReceipt
	var machineKind, decisionKind string
	if err := scanner.Scan(&item.CommandID, &machineKind, &item.AggregateID, &item.CommandKind, &item.CausationID, &item.Actor, &decisionKind, &item.Reason, &item.AggregateVersion, &item.ResultRef, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return transition.CommandReceipt{}, err
	}
	item.MachineKind = transition.MachineKind(machineKind)
	item.DecisionKind = transition.DecisionKind(decisionKind)
	return item, nil
}

func scanEffectExecution(scanner rowScanner) (transition.EffectExecution, error) {
	var item transition.EffectExecution
	var machineKind, effectKind, status string
	var payload []byte
	var startedAt, leaseExpiresAt, completedAt sql.NullTime
	if err := scanner.Scan(&item.ID, &machineKind, &item.AggregateID, &item.AttemptID, &effectKind, &status, &item.Holder, &item.IdempotencyKey, &payload, &item.ResultRef, &item.LastError, &item.RetryCount, &item.CreatedAt, &item.UpdatedAt, &startedAt, &leaseExpiresAt, &completedAt); err != nil {
		return transition.EffectExecution{}, err
	}
	item.MachineKind = transition.MachineKind(machineKind)
	item.EffectKind = transition.EffectKind(effectKind)
	item.Status = transition.EffectStatus(status)
	item.Payload = decodeJSON(payload, map[string]any{})
	if startedAt.Valid {
		t := startedAt.Time
		item.StartedAt = &t
	}
	if leaseExpiresAt.Valid {
		t := leaseExpiresAt.Time
		item.LeaseExpiresAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		item.CompletedAt = &t
	}
	return item, nil
}
