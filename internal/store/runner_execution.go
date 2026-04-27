package store

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var ErrHolderCASMismatch = errors.New("holder CAS failed")

const holderCASExpectEmpty = "*empty*"

func HolderCASExpectEmpty() string {
	return holderCASExpectEmpty
}

var activeRunnerExecutionStatuses = map[string]struct{}{
	"queued":           {},
	"accepted":         {},
	"starting":         {},
	"running":          {},
	"cancelling":       {},
	"cancel_requested": {},
}

func (s *MemoryStore) ListRunnerExecutions() []RunnerExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listRunnerExecutionsLocked()
}

func (s *MemoryStore) listRunnerExecutionsLocked() []RunnerExecution {
	out := make([]RunnerExecution, 0, len(s.runnerExecutions))
	for _, item := range s.runnerExecutions {
		out = append(out, normalizeRunnerExecution(item))
	}
	sortRunnerExecutions(out)
	return out
}

func (s *MemoryStore) ListActiveRunnerExecutions() []RunnerExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.listRunnerExecutionsLocked()
	out := make([]RunnerExecution, 0, len(items))
	for _, item := range items {
		if runnerExecutionStatusActive(item.Status) {
			out = append(out, item)
		}
	}
	return out
}

func (s *MemoryStore) GetRunnerExecution(executionID string) (RunnerExecution, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.runnerExecutions[strings.TrimSpace(executionID)]
	if !ok {
		return RunnerExecution{}, false
	}
	return normalizeRunnerExecution(item), true
}

func (s *MemoryStore) RecordRunnerExecution(item RunnerExecution) (RunnerExecution, error) {
	return s.RecordRunnerExecutionWithHolderCAS(item, "", nil)
}

func (s *MemoryStore) RecordRunnerExecutionWithHolderCAS(item RunnerExecution, expectedOldHolder string, expectedHeartbeatAt *time.Time) (RunnerExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing := s.runnerExecutions[strings.TrimSpace(item.ExecutionID)]
	casRequired := expectedOldHolder != "" || expectedHeartbeatAt != nil
	if casRequired && existing.ExecutionID == "" {
		return RunnerExecution{}, fmt.Errorf("%w: execution %q does not exist", ErrHolderCASMismatch, strings.TrimSpace(item.ExecutionID))
	}
	if casRequired {
		existing = normalizeRunnerExecution(existing)
		if err := validateRunnerExecutionCAS(existing, expectedOldHolder, expectedHeartbeatAt); err != nil {
			return RunnerExecution{}, err
		}
		if RunnerExecutionStatusTerminal(existing.Status) {
			return RunnerExecution{}, fmt.Errorf("%w: execution %q is terminal", ErrHolderCASMismatch, strings.TrimSpace(item.ExecutionID))
		}
	}
	item = mergeRunnerExecution(existing, item)
	if item.ExecutionID == "" {
		return RunnerExecution{}, fmt.Errorf("execution_id is required")
	}
	s.runnerExecutions[item.ExecutionID] = item
	return item, nil
}

func (s *MemoryStore) CancelRunnerExecutionsForCase(caseID string, exceptTraceID string, reason string) []RunnerExecution {
	s.mu.Lock()
	defer s.mu.Unlock()
	caseID = strings.TrimSpace(caseID)
	exceptTraceID = strings.TrimSpace(exceptTraceID)
	if caseID == "" {
		return []RunnerExecution{}
	}
	now := time.Now().UTC()
	out := []RunnerExecution{}
	for id, item := range s.runnerExecutions {
		item = normalizeRunnerExecution(item)
		if strings.TrimSpace(item.CaseID) != caseID || strings.TrimSpace(item.TraceID) == exceptTraceID || !runnerExecutionStatusActive(item.Status) {
			continue
		}
		item.CancelRequested = true
		if !strings.EqualFold(strings.TrimSpace(item.Status), "cancelling") {
			item.Status = "cancel_requested"
		}
		item.FailureClass = firstNonEmpty(item.FailureClass, strings.TrimSpace(reason))
		item.UpdatedAt = now
		s.runnerExecutions[id] = item
		out = append(out, item)
	}
	sortRunnerExecutions(out)
	return out
}

func (p *PostgresStore) ListRunnerExecutions() []RunnerExecution {
	rows, err := p.db.Query(`select ` + runnerExecutionSelectColumns() + ` from runner_execution order by updated_at desc, execution_id asc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []RunnerExecution{}
	for rows.Next() {
		item, err := scanRunnerExecution(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) ListActiveRunnerExecutions() []RunnerExecution {
	rows, err := p.db.Query(`select ` + runnerExecutionSelectColumns() + ` from runner_execution where status in ('queued','accepted','starting','running','cancelling','cancel_requested') order by updated_at desc, execution_id asc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []RunnerExecution{}
	for rows.Next() {
		item, err := scanRunnerExecution(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) GetRunnerExecution(executionID string) (RunnerExecution, bool) {
	row := p.db.QueryRow(`select `+runnerExecutionSelectColumns()+` from runner_execution where execution_id = $1`, strings.TrimSpace(executionID))
	item, err := scanRunnerExecution(row)
	if err != nil {
		return RunnerExecution{}, false
	}
	return item, true
}

func (p *PostgresStore) RecordRunnerExecution(item RunnerExecution) (RunnerExecution, error) {
	return p.RecordRunnerExecutionWithHolderCAS(item, "", nil)
}

func (p *PostgresStore) RecordRunnerExecutionWithHolderCAS(item RunnerExecution, expectedOldHolder string, expectedHeartbeatAt *time.Time) (RunnerExecution, error) {
	rawItem := item
	item = normalizeRunnerExecution(item)
	rawItem.ExecutionID = item.ExecutionID
	if item.ExecutionID == "" {
		return RunnerExecution{}, fmt.Errorf("execution_id is required")
	}
	var out RunnerExecution
	err := p.withTx(func(tx *sql.Tx) error {
		if err := advisoryLock(tx, "runner_execution:"+item.ExecutionID); err != nil {
			return err
		}
		existing, exists, err := selectRunnerExecutionForUpdate(tx, item.ExecutionID)
		if err != nil {
			return err
		}
		if expectedOldHolder != "" || expectedHeartbeatAt != nil {
			if !exists {
				return fmt.Errorf("%w: execution %q does not exist", ErrHolderCASMismatch, item.ExecutionID)
			}
			if err := validateRunnerExecutionCAS(existing, expectedOldHolder, expectedHeartbeatAt); err != nil {
				return err
			}
			if RunnerExecutionStatusTerminal(existing.Status) {
				return fmt.Errorf("%w: execution %q is terminal", ErrHolderCASMismatch, item.ExecutionID)
			}
		}
		if exists {
			out, err = updateRunnerExecutionRow(tx, mergeRunnerExecution(existing, rawItem))
		} else {
			out, err = insertRunnerExecutionRow(tx, rawItem)
		}
		return err
	})
	return out, err
}

func (p *PostgresStore) CancelRunnerExecutionsForCase(caseID string, exceptTraceID string, reason string) []RunnerExecution {
	caseID = strings.TrimSpace(caseID)
	exceptTraceID = strings.TrimSpace(exceptTraceID)
	if caseID == "" {
		return []RunnerExecution{}
	}
	now := time.Now().UTC()
	rows, err := p.db.Query(`update runner_execution
		set status = case when lower(status) = 'cancelling' then status else 'cancel_requested' end,
			cancel_requested = true,
			failure_class = coalesce(nullif(failure_class, ''), $3),
			updated_at = $4
		where case_id = $1
		  and trace_id <> $2
		  and status in ('queued','accepted','starting','running','cancelling','cancel_requested')
		returning `+runnerExecutionSelectColumns(),
		caseID,
		exceptTraceID,
		strings.TrimSpace(reason),
		now,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []RunnerExecution{}
	for rows.Next() {
		item, err := scanRunnerExecution(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func normalizeRunnerExecution(item RunnerExecution) RunnerExecution {
	item.ExecutionID = strings.TrimSpace(item.ExecutionID)
	item.OperationID = strings.TrimSpace(item.OperationID)
	item.WorkflowID = strings.TrimSpace(item.WorkflowID)
	item.TraceID = strings.TrimSpace(item.TraceID)
	item.ConversationID = strings.TrimSpace(item.ConversationID)
	item.CaseID = strings.TrimSpace(item.CaseID)
	item.Role = strings.TrimSpace(item.Role)
	item.Status = strings.ToLower(strings.TrimSpace(item.Status))
	if item.Status == "" {
		item.Status = "queued"
	}
	item.FailureClass = strings.TrimSpace(item.FailureClass)
	item.Holder = strings.TrimSpace(item.Holder)
	if item.Task == nil {
		item.Task = map[string]any{}
	}
	if item.Result == nil {
		item.Result = map[string]any{}
	}
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = item.CreatedAt
	}
	if RunnerExecutionStatusTerminal(item.Status) && item.HeartbeatAt == nil {
		heartbeatAt := item.UpdatedAt
		if item.CompletedAt != nil {
			heartbeatAt = *item.CompletedAt
		}
		item.HeartbeatAt = &heartbeatAt
	}
	return item
}

func mergeRunnerExecution(existing RunnerExecution, next RunnerExecution) RunnerExecution {
	nextStatusProvided := strings.TrimSpace(next.Status) != ""
	next = normalizeRunnerExecution(next)
	if existing.ExecutionID == "" {
		if !nextStatusProvided {
			next.Status = "queued"
		}
		return next
	}
	existing = normalizeRunnerExecution(existing)
	if RunnerExecutionStatusTerminal(existing.Status) {
		existing.OperationID = firstNonEmpty(next.OperationID, existing.OperationID)
		existing.WorkflowID = firstNonEmpty(next.WorkflowID, existing.WorkflowID)
		existing.TraceID = firstNonEmpty(next.TraceID, existing.TraceID)
		existing.ConversationID = firstNonEmpty(next.ConversationID, existing.ConversationID)
		existing.CaseID = firstNonEmpty(next.CaseID, existing.CaseID)
		existing.Role = firstNonEmpty(next.Role, existing.Role)
		return existing
	}
	existing.OperationID = firstNonEmpty(next.OperationID, existing.OperationID)
	existing.WorkflowID = firstNonEmpty(next.WorkflowID, existing.WorkflowID)
	existing.TraceID = firstNonEmpty(next.TraceID, existing.TraceID)
	existing.ConversationID = firstNonEmpty(next.ConversationID, existing.ConversationID)
	existing.CaseID = firstNonEmpty(next.CaseID, existing.CaseID)
	existing.Role = firstNonEmpty(next.Role, existing.Role)
	existingStatusLower := strings.ToLower(strings.TrimSpace(existing.Status))
	if existing.CancelRequested && nextStatusProvided {
		nextStatusLower := strings.ToLower(strings.TrimSpace(next.Status))
		switch {
		case nextStatusLower == "cancel_requested" || nextStatusLower == "cancelling" || RunnerExecutionStatusTerminal(next.Status):
			if !RunnerExecutionStatusBackward(existingStatusLower, nextStatusLower) {
				existing.Status = next.Status
			}
		case existingStatusLower != "cancel_requested" && existingStatusLower != "cancelling":
			existing.Status = "cancel_requested"
		}
	} else if nextStatusProvided {
		nextStatusLower := strings.ToLower(strings.TrimSpace(next.Status))
		if (nextStatusLower != "queued" || existingStatusLower == "queued") && !RunnerExecutionStatusBackward(existingStatusLower, nextStatusLower) {
			existing.Status = next.Status
		}
	}
	if len(next.Task) > 0 {
		existing.Task = next.Task
	}
	if len(next.Result) > 0 {
		existing.Result = next.Result
	}
	existing.FailureClass = firstNonEmpty(next.FailureClass, existing.FailureClass)
	existing.Holder = firstNonEmpty(next.Holder, existing.Holder)
	if next.RetryCount > existing.RetryCount {
		existing.RetryCount = next.RetryCount
	}
	existing.CancelRequested = existing.CancelRequested || next.CancelRequested
	if next.HeartbeatAt != nil {
		if existing.HeartbeatAt == nil || next.HeartbeatAt.After(*existing.HeartbeatAt) {
			existing.HeartbeatAt = next.HeartbeatAt
		}
	}
	existing.UpdatedAt = next.UpdatedAt
	if next.StartedAt != nil {
		existing.StartedAt = next.StartedAt
	}
	if next.CompletedAt != nil {
		existing.CompletedAt = next.CompletedAt
	}
	return existing
}

func runnerExecutionStatusActive(status string) bool {
	_, ok := activeRunnerExecutionStatuses[strings.ToLower(strings.TrimSpace(status))]
	return ok
}

func RunnerExecutionStatusTerminal(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "completed", "failed", "cancelled", "orphaned":
		return true
	default:
		return false
	}
}

func RunnerExecutionStatusBackward(existingStatusLower, nextStatusLower string) bool {
	statusOrder := map[string]int{
		"queued":           1,
		"accepted":         2,
		"starting":         3,
		"running":          4,
		"cancel_requested": 5,
		"cancelling":       6,
		"completed":        7,
		"failed":           7,
		"cancelled":        7,
		"orphaned":         7,
	}
	existingOrder, existingKnown := statusOrder[existingStatusLower]
	nextOrder, nextKnown := statusOrder[nextStatusLower]
	if !existingKnown || !nextKnown {
		return false
	}
	return nextOrder < existingOrder
}

func validateRunnerExecutionCAS(existing RunnerExecution, expectedOldHolder string, expectedHeartbeatAt *time.Time) error {
	if expectedOldHolder != "" {
		actualHolder := strings.TrimSpace(existing.Holder)
		if expectedOldHolder == holderCASExpectEmpty {
			if actualHolder != "" {
				return fmt.Errorf("%w: expected empty holder, got %q", ErrHolderCASMismatch, actualHolder)
			}
		} else if actualHolder != expectedOldHolder {
			return fmt.Errorf("%w: expected %q, got %q", ErrHolderCASMismatch, expectedOldHolder, existing.Holder)
		}
	}
	if expectedHeartbeatAt != nil {
		existingHeartbeat := existing.HeartbeatAt
		if existingHeartbeat == nil {
			return fmt.Errorf("%w: expected heartbeat %v, got <nil>", ErrHolderCASMismatch, *expectedHeartbeatAt)
		}
		if !existingHeartbeat.Equal(*expectedHeartbeatAt) {
			return fmt.Errorf("%w: expected heartbeat %v, got %v", ErrHolderCASMismatch, *expectedHeartbeatAt, *existingHeartbeat)
		}
	}
	return nil
}

func selectRunnerExecutionForUpdate(tx *sql.Tx, executionID string) (RunnerExecution, bool, error) {
	row := tx.QueryRow(`select `+runnerExecutionSelectColumns()+` from runner_execution where execution_id = $1 for update`, executionID)
	item, err := scanRunnerExecution(row)
	if err == sql.ErrNoRows {
		return RunnerExecution{}, false, nil
	}
	if err != nil {
		return RunnerExecution{}, false, err
	}
	return item, true, nil
}

func insertRunnerExecutionRow(tx *sql.Tx, item RunnerExecution) (RunnerExecution, error) {
	item = normalizeRunnerExecution(item)
	row := tx.QueryRow(`insert into runner_execution (execution_id, operation_id, workflow_id, trace_id, conversation_id, case_id, role, status, task, result, failure_class, holder, retry_count, cancel_requested, heartbeat_at, created_at, updated_at, started_at, completed_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10::jsonb,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		returning `+runnerExecutionSelectColumns(),
		item.ExecutionID,
		item.OperationID,
		item.WorkflowID,
		item.TraceID,
		item.ConversationID,
		item.CaseID,
		item.Role,
		item.Status,
		jsonString(item.Task),
		jsonString(item.Result),
		item.FailureClass,
		item.Holder,
		item.RetryCount,
		item.CancelRequested,
		nullTime(item.HeartbeatAt),
		item.CreatedAt,
		item.UpdatedAt,
		nullTime(item.StartedAt),
		nullTime(item.CompletedAt),
	)
	return scanRunnerExecution(row)
}

func updateRunnerExecutionRow(tx *sql.Tx, item RunnerExecution) (RunnerExecution, error) {
	item = normalizeRunnerExecution(item)
	row := tx.QueryRow(`update runner_execution set
			operation_id = $2,
			workflow_id = $3,
			trace_id = $4,
			conversation_id = $5,
			case_id = $6,
			role = $7,
			status = $8,
			task = $9::jsonb,
			result = $10::jsonb,
			failure_class = $11,
			holder = $12,
			retry_count = $13,
			cancel_requested = $14,
			heartbeat_at = $15,
			created_at = $16,
			updated_at = $17,
			started_at = $18,
			completed_at = $19
		where execution_id = $1
		returning `+runnerExecutionSelectColumns(),
		item.ExecutionID,
		item.OperationID,
		item.WorkflowID,
		item.TraceID,
		item.ConversationID,
		item.CaseID,
		item.Role,
		item.Status,
		jsonString(item.Task),
		jsonString(item.Result),
		item.FailureClass,
		item.Holder,
		item.RetryCount,
		item.CancelRequested,
		nullTime(item.HeartbeatAt),
		item.CreatedAt,
		item.UpdatedAt,
		nullTime(item.StartedAt),
		nullTime(item.CompletedAt),
	)
	return scanRunnerExecution(row)
}

func sortRunnerExecutions(items []RunnerExecution) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].ExecutionID < items[j].ExecutionID
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
}

func runnerExecutionSelectColumns() string {
	return `execution_id, operation_id, workflow_id, trace_id, conversation_id, case_id, role, status, task, result, failure_class, holder, retry_count, cancel_requested, heartbeat_at, created_at, updated_at, started_at, completed_at`
}

func scanRunnerExecution(scanner rowScanner) (RunnerExecution, error) {
	var item RunnerExecution
	var task, result []byte
	var heartbeatAt, startedAt, completedAt sql.NullTime
	if err := scanner.Scan(&item.ExecutionID, &item.OperationID, &item.WorkflowID, &item.TraceID, &item.ConversationID, &item.CaseID, &item.Role, &item.Status, &task, &result, &item.FailureClass, &item.Holder, &item.RetryCount, &item.CancelRequested, &heartbeatAt, &item.CreatedAt, &item.UpdatedAt, &startedAt, &completedAt); err != nil {
		return RunnerExecution{}, err
	}
	item.Task = decodeJSON(task, map[string]any{})
	item.Result = decodeJSON(result, map[string]any{})
	if heartbeatAt.Valid {
		t := heartbeatAt.Time
		item.HeartbeatAt = &t
	}
	if startedAt.Valid {
		t := startedAt.Time
		item.StartedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		item.CompletedAt = &t
	}
	return normalizeRunnerExecution(item), nil
}
