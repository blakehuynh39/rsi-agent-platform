package dbread

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type DBResult struct {
	Rows      []map[string]string `json:"rows"`
	RowCount  int                 `json:"row_count"`
	Truncated bool                `json:"truncated"`
}

func ValidateAgainstTarget(ctx context.Context, target Target, sqlText string) SQLValidation {
	safety := ValidateSQLSafety(sqlText)
	if !safety.OK {
		return safety
	}
	if strings.TrimSpace(target.DSN) == "" {
		safety.OK = false
		safety.ErrorCode = "target_dsn_missing"
		safety.Message = "Target-side validation requires a configured DSN on the local worker/relay"
		return safety
	}
	db, err := sql.Open("pgx", target.DSN)
	if err != nil {
		safety.OK = false
		safety.ErrorCode = "db_open_failed"
		safety.Message = sanitizeError(err)
		return safety
	}
	defer db.Close()
	timeout := time.Duration(target.Caps.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		safety.OK = false
		safety.ErrorCode = "db_begin_failed"
		safety.Message = sanitizeError(err)
		return safety
	}
	defer tx.Rollback()
	if err := setReadOnlyGuards(ctx, tx, target.Caps); err != nil {
		safety.OK = false
		safety.ErrorCode = "db_guard_failed"
		safety.Message = sanitizeError(err)
		return safety
	}
	stmtName := "rsi_db_read_validate"
	if _, err := tx.ExecContext(ctx, "PREPARE "+stmtName+" AS "+sqlText); err != nil {
		safety.OK = false
		safety.ErrorCode = "prepare_failed"
		safety.Message = sanitizeError(err)
		return safety
	}
	_, _ = tx.ExecContext(ctx, "DEALLOCATE "+stmtName)
	safety.OK = true
	return safety
}

func ExecuteRead(ctx context.Context, target Target, request storepkg.DBReadRequest) (DBResult, error) {
	if strings.TrimSpace(target.DSN) == "" {
		return DBResult{}, fmt.Errorf("target DSN is not configured")
	}
	db, err := sql.Open("pgx", target.DSN)
	if err != nil {
		return DBResult{}, err
	}
	defer db.Close()
	timeout := time.Duration(request.Caps.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return DBResult{}, err
	}
	defer tx.Rollback()
	if err := setReadOnlyGuards(ctx, tx, request.Caps); err != nil {
		return DBResult{}, err
	}
	rows, err := tx.QueryContext(ctx, request.SQL)
	if err != nil {
		return DBResult{}, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return DBResult{}, err
	}
	maxRows := request.Caps.MaxRows
	if maxRows <= 0 {
		maxRows = 100
	}
	maxBytes := request.Caps.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 64 * 1024
	}
	deny := denyColumnSet(request.Redaction)
	out := []map[string]string{}
	usedBytes := 0
	truncated := false
	for rows.Next() {
		values := make([]sql.RawBytes, len(columns))
		ptrs := make([]any, len(values))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return DBResult{}, err
		}
		if len(out) >= maxRows || usedBytes >= maxBytes {
			truncated = true
			break
		}
		row := map[string]string{}
		for i, col := range columns {
			if deny[strings.ToLower(col)] {
				row[col] = "[redacted]"
				continue
			}
			value := string(values[i])
			usedBytes += len(col) + len(value)
			if usedBytes > maxBytes {
				truncated = true
				break
			}
			row[col] = value
		}
		if truncated {
			break
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return DBResult{}, err
	}
	return DBResult{Rows: out, RowCount: len(out), Truncated: truncated}, nil
}

func setReadOnlyGuards(ctx context.Context, tx *sql.Tx, caps storepkg.DBReadCaps) error {
	timeout := caps.TimeoutSeconds
	if timeout <= 0 {
		timeout = 5
	}
	lockTimeout := caps.LockTimeoutMS
	if lockTimeout <= 0 {
		lockTimeout = 250
	}
	commands := []string{
		"set local transaction read only",
		"set local default_transaction_read_only = on",
		"set local search_path = public",
		"set local statement_timeout = " + quoteSQLLiteral(fmt.Sprintf("%ds", timeout)),
		"set local lock_timeout = " + quoteSQLLiteral(fmt.Sprintf("%dms", lockTimeout)),
	}
	for _, command := range commands {
		if _, err := tx.ExecContext(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func quoteSQLLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func denyColumnSet(policy storepkg.DBReadRedactionPolicy) map[string]bool {
	out := map[string]bool{}
	for _, col := range []string{"password", "token", "secret", "api_key", "private_key"} {
		out[col] = true
	}
	for _, col := range policy.DenyColumns {
		col = strings.ToLower(strings.TrimSpace(col))
		if col != "" {
			out[col] = true
		}
	}
	return out
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 800 {
		msg = msg[:800] + "..."
	}
	return msg
}
