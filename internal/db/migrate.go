package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationLockID int64 = 7_842_683_986_89

//go:embed migrations/*.sql
var migrationFS embed.FS

type Migration struct {
	Version int64
	Name    string
	SQL     string
}

type SchemaStatus struct {
	CurrentVersion  int64  `json:"current_version"`
	ExpectedVersion int64  `json:"expected_version"`
	State           string `json:"state"`
}

func OpenPostgres(postgresURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", postgresURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func LatestMigrationVersion() int64 {
	migrations, err := listMigrations()
	if err != nil || len(migrations) == 0 {
		return 0
	}
	return migrations[len(migrations)-1].Version
}

func VerifyCompatible(db *sql.DB) (SchemaStatus, error) {
	expected := LatestMigrationVersion()
	status := SchemaStatus{ExpectedVersion: expected, State: "unknown"}
	exists, err := migrationTableExists(db)
	if err != nil {
		return status, err
	}
	if !exists {
		status.State = "missing_version"
		return status, errors.New("rsi schema version missing; run improvement-plane --mode migrate")
	}
	current, err := currentMigrationVersion(db)
	if err != nil {
		return status, err
	}
	status.CurrentVersion = current
	switch {
	case current == expected:
		status.State = "compatible"
		return status, nil
	case current < expected:
		status.State = "behind"
		return status, fmt.Errorf("rsi schema behind binary: current=%d expected=%d; run improvement-plane --mode migrate", current, expected)
	default:
		status.State = "ahead"
		return status, fmt.Errorf("rsi schema ahead of binary: current=%d expected=%d", current, expected)
	}
}

func ApplyMigrations(db *sql.DB) (SchemaStatus, error) {
	migrations, err := listMigrations()
	if err != nil {
		return SchemaStatus{}, err
	}
	expected := int64(0)
	if len(migrations) > 0 {
		expected = migrations[len(migrations)-1].Version
	}
	if _, err := db.Exec(`select pg_advisory_lock($1)`, migrationLockID); err != nil {
		return SchemaStatus{ExpectedVersion: expected, State: "lock_failed"}, fmt.Errorf("acquire migration advisory lock: %w", err)
	}
	defer func() {
		_, _ = db.Exec(`select pg_advisory_unlock($1)`, migrationLockID)
	}()

	if err := ensureMigrationTable(db); err != nil {
		return SchemaStatus{ExpectedVersion: expected, State: "migration_table_error"}, err
	}

	applied, err := appliedMigrationVersions(db)
	if err != nil {
		return SchemaStatus{ExpectedVersion: expected, State: "read_failed"}, err
	}

	if len(applied) == 0 {
		empty, err := databaseLooksEmpty(db)
		if err != nil {
			return SchemaStatus{ExpectedVersion: expected, State: "inspect_failed"}, err
		}
		if !empty {
			ok, compatErr := baselineCompatible(db)
			if compatErr != nil {
				return SchemaStatus{ExpectedVersion: expected, State: "baseline_check_failed"}, compatErr
			}
			if !ok {
				return SchemaStatus{ExpectedVersion: expected, State: "incompatible"}, errors.New("existing RSI schema is incompatible with baseline migration; manual remediation required")
			}
			if len(migrations) == 0 {
				return SchemaStatus{ExpectedVersion: 0, State: "compatible"}, nil
			}
			if err := recordAppliedMigration(db, migrations[0]); err != nil {
				return SchemaStatus{ExpectedVersion: expected, State: "baseline_stamp_failed"}, err
			}
			applied[migrations[0].Version] = struct{}{}
		}
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; ok {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return SchemaStatus{ExpectedVersion: expected, State: "tx_begin_failed"}, err
		}
		if _, err := tx.Exec(migration.SQL); err != nil {
			_ = tx.Rollback()
			return SchemaStatus{ExpectedVersion: expected, State: "apply_failed"}, fmt.Errorf("apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}
		if _, err := tx.Exec(`insert into rsi_schema_migrations (version, name, applied_at) values ($1,$2,$3)`, migration.Version, migration.Name, time.Now().UTC()); err != nil {
			_ = tx.Rollback()
			return SchemaStatus{ExpectedVersion: expected, State: "record_failed"}, fmt.Errorf("record migration %d (%s): %w", migration.Version, migration.Name, err)
		}
		if err := tx.Commit(); err != nil {
			return SchemaStatus{ExpectedVersion: expected, State: "commit_failed"}, err
		}
	}

	return VerifyCompatible(db)
}

func RefreshSchemaSnapshot(destination string) error {
	migrations, err := listMigrations()
	if err != nil {
		return err
	}
	var builder strings.Builder
	for i, migration := range migrations {
		builder.WriteString(strings.TrimSpace(migration.SQL))
		builder.WriteString("\n")
		if i < len(migrations)-1 {
			builder.WriteString("\n")
		}
	}
	return os.WriteFile(destination, []byte(builder.String()), 0o644)
}

func listMigrations() ([]Migration, error) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	out := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		version, name, err := parseMigrationName(entry.Name())
		if err != nil {
			return nil, err
		}
		body, err := migrationFS.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, Migration{
			Version: version,
			Name:    name,
			SQL:     string(body),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

func parseMigrationName(name string) (int64, string, error) {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	parts := strings.SplitN(base, "_", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return 0, "", fmt.Errorf("invalid migration filename %q", name)
	}
	version, err := strconv.ParseInt(strings.TrimLeft(parts[0], "0"), 10, 64)
	if err != nil {
		if strings.Trim(parts[0], "0") == "" {
			version = 0
		} else {
			return 0, "", fmt.Errorf("parse migration version %q: %w", name, err)
		}
	}
	if version == 0 {
		version = 1
	}
	migrationName := base
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		migrationName = parts[1]
	}
	return version, migrationName, nil
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`
		create table if not exists rsi_schema_migrations (
			version bigint primary key,
			name text not null,
			applied_at timestamptz not null
		)
	`)
	return err
}

func migrationTableExists(db *sql.DB) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		select exists (
			select 1
			from information_schema.tables
			where table_schema = current_schema()
			  and table_name = 'rsi_schema_migrations'
		)
	`).Scan(&exists)
	return exists, err
}

func currentMigrationVersion(db *sql.DB) (int64, error) {
	var version sql.NullInt64
	if err := db.QueryRow(`select max(version) from rsi_schema_migrations`).Scan(&version); err != nil {
		return 0, err
	}
	if !version.Valid {
		return 0, nil
	}
	return version.Int64, nil
}

func appliedMigrationVersions(db *sql.DB) (map[int64]struct{}, error) {
	rows, err := db.Query(`select version from rsi_schema_migrations order by version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]struct{}{}
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		out[version] = struct{}{}
	}
	return out, rows.Err()
}

func recordAppliedMigration(db *sql.DB, migration Migration) error {
	_, err := db.Exec(`insert into rsi_schema_migrations (version, name, applied_at) values ($1,$2,$3)`, migration.Version, migration.Name, time.Now().UTC())
	return err
}

func databaseLooksEmpty(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow(`
		select count(*)
		from information_schema.tables
		where table_schema = current_schema()
		  and table_name <> 'rsi_schema_migrations'
	`).Scan(&count)
	return count == 0, err
}

func baselineCompatible(db *sql.DB) (bool, error) {
	required := map[string][]string{
		"event_envelope":        {"id", "source", "source_event_id", "dedupe_key", "workflow_hint"},
		"conversation":          {"id", "source", "external_key", "active_case_id", "latest_event_id"},
		"case_record":           {"id", "conversation_id", "resolution_state", "latest_outcome_id"},
		"trace_summary":         {"trace_id", "conversation_id", "case_id", "trigger_event_id", "supersedes_trace_id"},
		"trace_event":           {"trace_id", "conversation_id", "case_id", "trigger_event_id", "description"},
		"action_intent":         {"id", "kind", "phase_key", "status"},
		"action_result":         {"id", "action_intent_id", "status"},
		"outcome_record":        {"id", "proposal_id", "outcome_type", "verdict"},
		"knowledge_entry":       {"id", "tier", "status", "source_type"},
		"improvement_candidate": {"id", "candidate_key", "origin_trace_id", "evidence_trace_ids", "source_eval_ids"},
		"proposal":              {"id", "origin_trace_id", "evidence_trace_ids", "candidate_key"},
		"work_item":             {"id", "queue", "conversation_id", "case_id", "trigger_event_id"},
	}
	for table, columns := range required {
		ok, err := tableHasColumns(db, table, columns...)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func tableHasColumns(db *sql.DB, table string, columns ...string) (bool, error) {
	rows, err := db.Query(`
		select column_name
		from information_schema.columns
		where table_schema = current_schema()
		  and table_name = $1
	`, table)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	present := map[string]struct{}{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return false, err
		}
		present[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	if len(present) == 0 {
		return false, nil
	}
	for _, column := range columns {
		if _, ok := present[column]; !ok {
			return false, nil
		}
	}
	return true, nil
}
