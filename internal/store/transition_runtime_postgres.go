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

func (p *PostgresStore) QueueEffectExecution(effect transition.EffectExecution) (created transition.EffectExecution, wasCreated bool, err error) {
	effect.ID = strings.TrimSpace(effect.ID)
	effect.AggregateID = strings.TrimSpace(effect.AggregateID)
	effect.AttemptID = strings.TrimSpace(effect.AttemptID)
	effect.IdempotencyKey = strings.TrimSpace(effect.IdempotencyKey)
	if effect.ID == "" {
		return transition.EffectExecution{}, false, fmt.Errorf("effect execution id is required")
	}
	if effect.MachineKind == "" {
		return transition.EffectExecution{}, false, fmt.Errorf("machine kind is required")
	}
	if effect.AggregateID == "" {
		return transition.EffectExecution{}, false, fmt.Errorf("aggregate id is required")
	}
	if effect.EffectKind == "" {
		return transition.EffectExecution{}, false, fmt.Errorf("effect kind is required")
	}
	if effect.IdempotencyKey == "" {
		return transition.EffectExecution{}, false, fmt.Errorf("idempotency key is required")
	}
	now := time.Now().UTC()
	if effect.Status == "" {
		effect.Status = transition.EffectQueued
	}
	if effect.Payload == nil {
		effect.Payload = map[string]any{}
	}
	if effect.CreatedAt.IsZero() {
		effect.CreatedAt = now
	}
	if effect.UpdatedAt.IsZero() || effect.UpdatedAt.Before(effect.CreatedAt) {
		effect.UpdatedAt = effect.CreatedAt
	}
	effect = normalizeEffectScheduling(effect)
	err = p.withTx(func(tx *sql.Tx) error {
		if err := advisoryLock(tx, "effect-execution:"+effect.IdempotencyKey); err != nil {
			return err
		}
		row := tx.QueryRow(`select `+effectExecutionSelectColumns()+` from effect_execution where idempotency_key = $1`, effect.IdempotencyKey)
		existing, scanErr := scanEffectExecution(row)
		if scanErr == nil {
			created = existing
			wasCreated = false
			return nil
		}
		if scanErr != sql.ErrNoRows {
			return scanErr
		}
		if err := persistEffectExecutions(tx, []transition.EffectExecution{effect}); err != nil {
			return err
		}
		row = tx.QueryRow(`select `+effectExecutionSelectColumns()+` from effect_execution where idempotency_key = $1`, effect.IdempotencyKey)
		created, scanErr = scanEffectExecution(row)
		wasCreated = scanErr == nil
		return scanErr
	})
	return
}

func (p *PostgresStore) ClaimEffectExecution(effectID string, holder string, lease time.Duration) (item transition.EffectExecution, claimed bool, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		var leaseExpires *time.Time
		if lease > 0 {
			expiresAt := now.Add(lease)
			leaseExpires = &expiresAt
		}
		row := tx.QueryRow(`
			update effect_execution
			set status = $2,
				holder = $3,
				updated_at = $4,
				started_at = coalesce(started_at, $4),
				lease_expires_at = $5::timestamptz,
				not_before = null,
				completed_at = null
			where id = $1
			  and (not_before is null or not_before <= $4)
				  and (
					status in ($6, $7)
					or (status = $8 and (
						lease_expires_at is not null and lease_expires_at <= $4
					))
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

func (p *PostgresStore) ClaimNextEffectExecution(holder string, lease time.Duration, queueNames []string, maxPerScope int) (item transition.EffectExecution, claimed bool, err error) {
	return p.ClaimNextEffectExecutionForKinds(holder, lease, queueNames, maxPerScope, nil)
}

func (p *PostgresStore) ClaimNextEffectExecutionForKinds(holder string, lease time.Duration, queueNames []string, maxPerScope int, selectors []EffectClaimSelector) (item transition.EffectExecution, claimed bool, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		var leaseExpires *time.Time
		if lease > 0 {
			expiresAt := now.Add(lease)
			leaseExpires = &expiresAt
		}
		args := []any{
			string(transition.EffectRunning),
			strings.TrimSpace(holder),
			now,
			leaseExpires,
			string(transition.EffectQueued),
			string(transition.EffectFailed),
			maxPerScope,
		}
		queueFilter := ""
		hasNonEmptyQueue := false
		for _, queueName := range queueNames {
			queueName = strings.TrimSpace(queueName)
			if queueName == "" {
				continue
			}
			hasNonEmptyQueue = true
			args = append(args, queueName)
			if queueFilter == "" {
				queueFilter = fmt.Sprintf(" and e.queue_name in ($%d", len(args))
			} else {
				queueFilter += fmt.Sprintf(", $%d", len(args))
			}
		}
		if !hasNonEmptyQueue {
			return nil
		}
		if queueFilter != "" {
			queueFilter += ")"
		}
		selectorGroups := []string{}
		for _, selector := range selectors {
			parts := []string{}
			if selector.MachineKind != "" {
				args = append(args, string(selector.MachineKind))
				parts = append(parts, fmt.Sprintf("e.machine_kind = $%d", len(args)))
			}
			if selector.EffectKind != "" {
				args = append(args, string(selector.EffectKind))
				parts = append(parts, fmt.Sprintf("e.effect_kind = $%d", len(args)))
			}
			for key, expected := range selector.PayloadEquals {
				key = strings.TrimSpace(key)
				if key == "" {
					continue
				}
				args = append(args, key, strings.TrimSpace(expected))
				parts = append(parts, fmt.Sprintf("lower(trim(coalesce(e.payload ->> $%d, ''))) = lower($%d)", len(args)-1, len(args)))
			}
			if len(parts) > 0 {
				selectorGroups = append(selectorGroups, "("+strings.Join(parts, " and ")+")")
			}
		}
		selectorFilter := ""
		if len(selectorGroups) > 0 {
			selectorFilter = " and (" + strings.Join(selectorGroups, " or ") + ")"
		}
		queueFilterForActive := strings.ReplaceAll(queueFilter, "e.", "active.")
		selectorFilterForActive := strings.ReplaceAll(selectorFilter, "e.", "active.")
		query := `
			with eligible as (
				select
					e.id,
					coalesce(nullif(e.scope_key, ''), e.id) as lock_scope,
					e.priority,
					e.created_at,
					row_number() over (
						partition by coalesce(nullif(e.scope_key, ''), e.id)
						order by e.priority desc, e.created_at asc, e.id asc
					) as scope_rank
					from effect_execution e
					where (
							e.status in ($5, $6)
							or (e.status = $1 and (
								e.lease_expires_at is not null and e.lease_expires_at <= $3
							))
						)
				  and (e.not_before is null or e.not_before <= $3)
				  and (
					$7 <= 0
					or coalesce((
						select count(*)
						from effect_execution active
						where coalesce(nullif(active.scope_key, ''), active.id) = coalesce(nullif(e.scope_key, ''), e.id)
								  and active.status = $1
								  and (
									active.lease_expires_at is null
									or active.lease_expires_at > $3
									or (active.not_before is not null and active.not_before > $3)
								  )` + queueFilterForActive + selectorFilterForActive + `
							), 0) < $7
				  )` + queueFilter + selectorFilter + `
			),
			candidates as (
				select e.id, eligible.lock_scope, e.priority, e.created_at
				from effect_execution e
				join eligible on eligible.id = e.id
				where eligible.scope_rank = 1
				order by e.priority desc, e.created_at asc, e.id asc
				for update skip locked
			),
			locked_candidate as (
				select c.id
				from candidates c
				where $7 <= 0 or pg_try_advisory_xact_lock(hashtext('rsi-effect-scope'), hashtext(c.lock_scope))
				order by c.priority desc, c.created_at asc, c.id asc
				limit 1
			)
			update effect_execution
			set status = $1,
				holder = $2,
				updated_at = $3,
				started_at = coalesce(started_at, $3),
				lease_expires_at = $4::timestamptz,
				not_before = null,
				completed_at = null
			where id = (select id from locked_candidate)
				  and (
					    status in ($5, $6)
					    or (status = $1 and (
							lease_expires_at is not null and lease_expires_at <= $3
						))
					  )
			  and (not_before is null or not_before <= $3)
			returning ` + effectExecutionSelectColumns()
		row := tx.QueryRow(query, args...)
		item, err = scanEffectExecution(row)
		if err == sql.ErrNoRows {
			claimed = false
			return nil
		}
		claimed = err == nil
		return err
	})
	return
}

func (p *PostgresStore) CompleteEffectExecution(effectID string, holder string, resultRef string) (item transition.EffectExecution, err error) {
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
				not_before = null,
				completed_at = $4
			where id = $1
			  and status = $5
			  and holder = $6
			  and (lease_expires_at is null or lease_expires_at > $4)
			returning `+effectExecutionSelectColumns(),
			strings.TrimSpace(effectID),
			string(transition.EffectCompleted),
			strings.TrimSpace(resultRef),
			now,
			string(transition.EffectRunning),
			strings.TrimSpace(holder),
		)
		item, err = scanEffectExecution(row)
		if err == sql.ErrNoRows {
			row = tx.QueryRow(`select `+effectExecutionSelectColumns()+` from effect_execution where id = $1`, strings.TrimSpace(effectID))
			item, err = scanEffectExecution(row)
		}
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (p *PostgresStore) DeferEffectExecution(effectID string, holder string, lease time.Duration, reason string) (item transition.EffectExecution, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		expires := now
		if lease > 0 {
			expires = now.Add(lease)
		}
		row := tx.QueryRow(`
			update effect_execution
			set holder = '',
				last_error = $2,
				updated_at = $3,
				lease_expires_at = $4,
				not_before = $4,
				completed_at = null
			where id = $1
			  and status = $5
			  and holder = $6
			  and (lease_expires_at is null or lease_expires_at > $3)
			returning `+effectExecutionSelectColumns(),
			strings.TrimSpace(effectID),
			strings.TrimSpace(reason),
			now,
			expires,
			string(transition.EffectRunning),
			strings.TrimSpace(holder),
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

func (p *PostgresStore) FailEffectExecution(effectID string, holder string, lastError string) (item transition.EffectExecution, err error) {
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
				not_before = null,
				completed_at = $4
			where id = $1
			  and status = $5
			  and holder = $6
			  and (lease_expires_at is null or lease_expires_at > $4)
			returning `+effectExecutionSelectColumns(),
			strings.TrimSpace(effectID),
			string(transition.EffectFailed),
			strings.TrimSpace(lastError),
			now,
			string(transition.EffectRunning),
			strings.TrimSpace(holder),
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
	return `id, machine_kind, aggregate_id, attempt_id, effect_kind, status, holder, idempotency_key, queue_name, scope_key, task_class, priority, not_before, payload, result_ref, last_error, retry_count, created_at, updated_at, started_at, lease_expires_at, completed_at`
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
	var startedAt, leaseExpiresAt, completedAt, notBefore sql.NullTime
	if err := scanner.Scan(&item.ID, &machineKind, &item.AggregateID, &item.AttemptID, &effectKind, &status, &item.Holder, &item.IdempotencyKey, &item.QueueName, &item.ScopeKey, &item.TaskClass, &item.Priority, &notBefore, &payload, &item.ResultRef, &item.LastError, &item.RetryCount, &item.CreatedAt, &item.UpdatedAt, &startedAt, &leaseExpiresAt, &completedAt); err != nil {
		return transition.EffectExecution{}, err
	}
	item.MachineKind = transition.MachineKind(machineKind)
	item.EffectKind = transition.EffectKind(effectKind)
	item.Status = transition.EffectStatus(status)
	item.Payload = decodeJSON(payload, map[string]any{})
	if notBefore.Valid {
		t := notBefore.Time
		item.NotBefore = &t
	}
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
	return normalizeEffectScheduling(item), nil
}
