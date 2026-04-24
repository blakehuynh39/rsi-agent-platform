package store

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
)

func (s *MemoryStore) ListExecutionLedgerEvents() []events.ExecutionLedgerEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]events.ExecutionLedgerEvent(nil), s.executionLedgerEvents...)
	for i := range out {
		out[i] = normalizeExecutionLedgerEvent(out[i])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RecordedAt.Equal(out[j].RecordedAt) {
			if out[i].ExecutionID == out[j].ExecutionID {
				if out[i].Seq == out[j].Seq {
					return out[i].ID < out[j].ID
				}
				return out[i].Seq < out[j].Seq
			}
			return out[i].ExecutionID < out[j].ExecutionID
		}
		return out[i].RecordedAt.After(out[j].RecordedAt)
	})
	return out
}

func (s *MemoryStore) RecordExecutionLedgerEvents(items []events.ExecutionLedgerEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	index := map[string]int{}
	for i, existing := range s.executionLedgerEvents {
		key := executionLedgerEventKey(existing)
		if key != "" {
			index[key] = i
		}
	}
	for _, item := range items {
		item = normalizeExecutionLedgerEvent(item)
		if item.RecordedAt.IsZero() {
			item.RecordedAt = now
		}
		if item.ID == "" {
			item.ID = nextUUID("xled")
		}
		key := executionLedgerEventKey(item)
		if key == "" {
			continue
		}
		if idx, ok := index[key]; ok {
			item.ID = firstNonEmpty(s.executionLedgerEvents[idx].ID, item.ID)
			s.executionLedgerEvents[idx] = item
			continue
		}
		index[key] = len(s.executionLedgerEvents)
		s.executionLedgerEvents = append(s.executionLedgerEvents, item)
	}
	return nil
}

func (p *PostgresStore) ListExecutionLedgerEvents() []events.ExecutionLedgerEvent {
	rows, err := p.db.Query(`
		select id, execution_id, operation_id, trace_id, workflow_id, phase_id, kind, status, seq, idempotency_key, payload, recorded_at
		from execution_ledger_event
		order by recorded_at desc, execution_id asc, seq asc, id asc
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []events.ExecutionLedgerEvent{}
	for rows.Next() {
		item, err := scanExecutionLedgerEvent(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) RecordExecutionLedgerEvents(items []events.ExecutionLedgerEvent) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	now := time.Now().UTC()
	for _, item := range items {
		item = normalizeExecutionLedgerEvent(item)
		if item.RecordedAt.IsZero() {
			item.RecordedAt = now
		}
		if item.ID == "" {
			item.ID = nextUUID("xled")
		}
		if executionLedgerEventKey(item) == "" {
			continue
		}
		if _, err := tx.Exec(`
			insert into execution_ledger_event (
				id, execution_id, operation_id, trace_id, workflow_id, phase_id, kind, status, seq, idempotency_key, payload, recorded_at
			) values (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12
			)
			on conflict (execution_id, seq) do update set
				operation_id = excluded.operation_id,
				trace_id = excluded.trace_id,
				workflow_id = excluded.workflow_id,
				phase_id = excluded.phase_id,
				kind = excluded.kind,
				status = excluded.status,
				idempotency_key = excluded.idempotency_key,
				payload = excluded.payload,
				recorded_at = excluded.recorded_at
		`,
			item.ID,
			item.ExecutionID,
			item.OperationID,
			item.TraceID,
			item.WorkflowID,
			item.PhaseID,
			item.Kind,
			item.Status,
			item.Seq,
			item.IdempotencyKey,
			jsonString(item.Payload),
			item.RecordedAt,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type executionLedgerScanner interface {
	Scan(dest ...any) error
}

func scanExecutionLedgerEvent(scanner executionLedgerScanner) (events.ExecutionLedgerEvent, error) {
	var (
		item    events.ExecutionLedgerEvent
		payload []byte
	)
	err := scanner.Scan(
		&item.ID,
		&item.ExecutionID,
		&item.OperationID,
		&item.TraceID,
		&item.WorkflowID,
		&item.PhaseID,
		&item.Kind,
		&item.Status,
		&item.Seq,
		&item.IdempotencyKey,
		&payload,
		&item.RecordedAt,
	)
	if err != nil {
		return events.ExecutionLedgerEvent{}, err
	}
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &item.Payload)
	}
	return normalizeExecutionLedgerEvent(item), nil
}

func normalizeExecutionLedgerEvent(item events.ExecutionLedgerEvent) events.ExecutionLedgerEvent {
	item.ID = strings.TrimSpace(item.ID)
	item.ExecutionID = strings.TrimSpace(item.ExecutionID)
	item.OperationID = strings.TrimSpace(item.OperationID)
	item.TraceID = strings.TrimSpace(item.TraceID)
	item.WorkflowID = strings.TrimSpace(item.WorkflowID)
	item.PhaseID = strings.TrimSpace(item.PhaseID)
	item.Kind = strings.TrimSpace(item.Kind)
	item.Status = strings.TrimSpace(item.Status)
	item.IdempotencyKey = strings.TrimSpace(item.IdempotencyKey)
	if item.Payload == nil {
		item.Payload = map[string]any{}
	}
	return item
}

func executionLedgerEventKey(item events.ExecutionLedgerEvent) string {
	if strings.TrimSpace(item.ExecutionID) == "" || item.Seq <= 0 {
		return ""
	}
	return strings.TrimSpace(item.ExecutionID) + "|" + strconv.Itoa(item.Seq)
}
