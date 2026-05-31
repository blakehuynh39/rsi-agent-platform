package store

import (
	"database/sql"

	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type transitionPersistBundle struct {
	Events   []transition.DomainEvent
	Commands []transition.CommandEnvelope
	Effects  []transition.EffectExecution
}

func persistDomainEvents(tx *sql.Tx, items []transition.DomainEvent) error {
	for _, item := range items {
		if _, err := tx.Exec(`insert into domain_event (id, machine_kind, aggregate_id, aggregate_version, event_kind, command_id, causation_id, payload, created_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9)
			on conflict (id) do nothing`,
			item.ID,
			string(item.MachineKind),
			item.AggregateID,
			item.AggregateVersion,
			item.EventKind,
			firstNonEmpty(item.CommandID),
			firstNonEmpty(item.CausationID),
			jsonString(item.Payload),
			item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistEffectExecutions(tx *sql.Tx, items []transition.EffectExecution) error {
	for _, item := range items {
		item = normalizeEffectScheduling(item)
		if _, err := tx.Exec(`insert into effect_execution (id, machine_kind, aggregate_id, attempt_id, effect_kind, status, holder, idempotency_key, queue_name, scope_key, task_class, priority, not_before, payload, result_ref, last_error, retry_count, created_at, updated_at, started_at, lease_expires_at, completed_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::jsonb,$15,$16,$17,$18,$19,$20,$21,$22)
			on conflict (idempotency_key) do update set
				status = excluded.status,
				holder = excluded.holder,
				queue_name = excluded.queue_name,
				scope_key = excluded.scope_key,
				task_class = excluded.task_class,
				priority = excluded.priority,
				not_before = excluded.not_before,
				payload = excluded.payload,
				result_ref = excluded.result_ref,
				last_error = excluded.last_error,
				retry_count = excluded.retry_count,
				updated_at = excluded.updated_at,
				started_at = excluded.started_at,
				lease_expires_at = excluded.lease_expires_at,
				completed_at = excluded.completed_at`,
			item.ID,
			string(item.MachineKind),
			item.AggregateID,
			firstNonEmpty(item.AttemptID),
			string(item.EffectKind),
			string(item.Status),
			firstNonEmpty(item.Holder),
			item.IdempotencyKey,
			item.QueueName,
			item.ScopeKey,
			item.TaskClass,
			item.Priority,
			nullTime(item.NotBefore),
			jsonString(item.Payload),
			firstNonEmpty(item.ResultRef),
			firstNonEmpty(item.LastError),
			item.RetryCount,
			item.CreatedAt,
			item.UpdatedAt,
			nullTime(item.StartedAt),
			nullTime(item.LeaseExpiresAt),
			nullTime(item.CompletedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistCommandReceipts(tx *sql.Tx, items []transition.CommandReceipt) error {
	for _, item := range items {
		if _, err := tx.Exec(`insert into command_receipt (command_id, machine_kind, aggregate_id, command_kind, causation_id, actor, decision_kind, reason, aggregate_version, result_ref, created_at, updated_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			on conflict (command_id) do update set
				machine_kind = excluded.machine_kind,
				aggregate_id = excluded.aggregate_id,
				command_kind = excluded.command_kind,
				causation_id = excluded.causation_id,
				actor = excluded.actor,
				decision_kind = excluded.decision_kind,
				reason = excluded.reason,
				aggregate_version = excluded.aggregate_version,
				result_ref = excluded.result_ref,
				updated_at = excluded.updated_at`,
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
		); err != nil {
			return err
		}
	}
	return nil
}
