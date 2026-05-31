package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMigrationApplyToEmptyDatabase(t *testing.T) {
	db, cleanup := openTempDatabase(t)
	defer cleanup()

	status, err := ApplyMigrations(db)
	if err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}
	if status.State != "compatible" {
		t.Fatalf("expected compatible schema state, got %+v", status)
	}
	if status.CurrentVersion != LatestMigrationVersion() {
		t.Fatalf("expected current version %d, got %d", LatestMigrationVersion(), status.CurrentVersion)
	}
	assertHonchoSchemaArtifacts(t, db)
	assertQuestionRunSchemaRemoved(t, db)
}

func TestMigrationBaselineStampCompatibleDatabase(t *testing.T) {
	db, cleanup := openTempDatabase(t)
	defer cleanup()

	if _, err := db.Exec(SchemaSQL); err != nil {
		t.Fatalf("seed schema.sql: %v", err)
	}

	status, err := ApplyMigrations(db)
	if err != nil {
		t.Fatalf("ApplyMigrations() baseline stamp error = %v", err)
	}
	if status.State != "compatible" || status.CurrentVersion != LatestMigrationVersion() {
		t.Fatalf("unexpected schema status after baseline stamp: %+v", status)
	}
	migrations, err := listMigrations()
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	applied, err := appliedMigrationVersions(db)
	if err != nil {
		t.Fatalf("read applied migration versions: %v", err)
	}
	if len(applied) != len(migrations) {
		t.Fatalf("expected %d applied migrations after baseline stamp, got %d", len(migrations), len(applied))
	}
	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; !ok {
			t.Fatalf("expected migration %d to be stamped as applied", migration.Version)
		}
	}
	assertHonchoSchemaArtifacts(t, db)
	assertQuestionRunSchemaRemoved(t, db)
}

func TestMigrationUpgradeDropsQuestionRunCatalog(t *testing.T) {
	db, cleanup := openTempDatabase(t)
	defer cleanup()

	migrations, err := listMigrations()
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	if err := ensureMigrationTable(db); err != nil {
		t.Fatalf("ensure migration table: %v", err)
	}
	for _, migration := range migrations {
		if migration.Version >= 30 {
			break
		}
		if _, err := db.Exec(migration.SQL); err != nil {
			t.Fatalf("seed migration %d (%s): %v", migration.Version, migration.Name, err)
		}
		if _, err := db.Exec(`insert into rsi_schema_migrations (version, name, applied_at) values ($1,$2,$3)`, migration.Version, migration.Name, time.Now().UTC()); err != nil {
			t.Fatalf("record migration %d: %v", migration.Version, err)
		}
	}
	if _, err := db.Exec(`insert into question_run (id, workflow_id, status) values ('qr-prb', 'wf-prb', 'queued')`); err != nil {
		t.Fatalf("seed legacy question_run row: %v", err)
	}
	if _, err := db.Exec(`insert into effect_execution (id, machine_kind, aggregate_id, effect_kind, status, idempotency_key) values ('eff-prb', 'question_run', 'qr-prb', 'gather_evidence', 'queued', 'qr-prb:gather')`); err != nil {
		t.Fatalf("seed legacy question_run effect: %v", err)
	}
	if _, err := db.Exec(`insert into command_receipt (command_id, machine_kind, aggregate_id, command_kind, decision_kind) values ('cmd-prb', 'question_run', 'qr-prb', 'question_run_started', 'advance')`); err != nil {
		t.Fatalf("seed legacy question_run receipt: %v", err)
	}
	if _, err := db.Exec(`insert into domain_event (id, machine_kind, aggregate_id, aggregate_version, event_kind) values ('evt-prb', 'question_run', 'qr-prb', 1, 'question_run_started')`); err != nil {
		t.Fatalf("seed legacy question_run event: %v", err)
	}

	status, err := ApplyMigrations(db)
	if err != nil {
		t.Fatalf("ApplyMigrations() upgrade error = %v", err)
	}
	if status.State != "compatible" || status.CurrentVersion != LatestMigrationVersion() {
		t.Fatalf("unexpected schema status after upgrade: %+v", status)
	}
	assertQuestionRunSchemaRemoved(t, db)
	assertNoQuestionRunTransitionRows(t, db)
}

func TestVerifyAtLeastAllowsSchemaAheadOfBinary(t *testing.T) {
	db, cleanup := openTempDatabase(t)
	defer cleanup()

	if _, err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}
	aheadVersion := LatestMigrationVersion() + 1
	if _, err := db.Exec(`insert into rsi_schema_migrations (version, name, applied_at) values ($1,$2,$3)`, aheadVersion, "future", time.Now().UTC()); err != nil {
		t.Fatalf("record future migration: %v", err)
	}

	if _, err := VerifyCompatible(db); err == nil || !strings.Contains(err.Error(), "ahead of binary") {
		t.Fatalf("expected strict compatibility to reject ahead schema, got %v", err)
	}
	status, err := VerifyAtLeast(db, LatestMigrationVersion())
	if err != nil {
		t.Fatalf("VerifyAtLeast() error = %v", err)
	}
	if status.State != "ahead" || status.CurrentVersion != aheadVersion {
		t.Fatalf("expected ahead-but-allowed schema status, got %+v", status)
	}
}

func TestVerifyAtLeastRejectsSchemaBelowMinimum(t *testing.T) {
	db, cleanup := openTempDatabase(t)
	defer cleanup()

	if err := ensureMigrationTable(db); err != nil {
		t.Fatalf("ensure migration table: %v", err)
	}
	if _, err := db.Exec(`insert into rsi_schema_migrations (version, name, applied_at) values ($1,$2,$3)`, int64(28), "old", time.Now().UTC()); err != nil {
		t.Fatalf("record old migration: %v", err)
	}

	status, err := VerifyAtLeast(db, 29)
	if err == nil || !strings.Contains(err.Error(), "behind required minimum") {
		t.Fatalf("expected minimum schema error, got status=%+v err=%v", status, err)
	}
}

func TestMigrationRejectsIncompatibleExistingDatabase(t *testing.T) {
	db, cleanup := openTempDatabase(t)
	defer cleanup()

	if _, err := db.Exec(`create table event_envelope (id text primary key)`); err != nil {
		t.Fatalf("seed partial schema: %v", err)
	}
	if _, err := db.Exec(`create table rsi_schema_migrations (version bigint primary key, name text not null, applied_at timestamptz not null)`); err != nil {
		t.Fatalf("seed migration table: %v", err)
	}
	if _, err := db.Exec(`delete from rsi_schema_migrations`); err != nil {
		t.Fatalf("clear migration table: %v", err)
	}

	_, err := ApplyMigrations(db)
	if err == nil || !strings.Contains(err.Error(), "incompatible") {
		t.Fatalf("expected incompatible schema error, got %v", err)
	}
}

func TestMigrationRejectsBaselineWithoutHonchoArtifacts(t *testing.T) {
	db, cleanup := openTempDatabase(t)
	defer cleanup()

	migrations, err := listMigrations()
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	for _, migration := range migrations {
		if migration.Version >= 4 {
			break
		}
		if _, err := db.Exec(migration.SQL); err != nil {
			t.Fatalf("seed migration %d (%s): %v", migration.Version, migration.Name, err)
		}
	}

	_, err = ApplyMigrations(db)
	if err == nil || !strings.Contains(err.Error(), "incompatible") {
		t.Fatalf("expected incompatible schema error for baseline without honcho artifacts, got %v", err)
	}
}

func TestMigrationAdvisoryLockSerializesRunners(t *testing.T) {
	db, cleanup := openTempDatabase(t)
	defer cleanup()

	baseURL := strings.TrimSpace(os.Getenv("RSI_TEST_POSTGRES_URL"))
	var dbName string
	if err := db.QueryRow(`select current_database()`).Scan(&dbName); err != nil {
		t.Fatalf("read current database: %v", err)
	}
	lockURL, err := withDatabase(baseURL, dbName)
	if err != nil {
		t.Fatalf("build lock URL: %v", err)
	}
	locker, err := OpenPostgres(lockURL)
	if err != nil {
		t.Fatalf("open locker postgres: %v", err)
	}
	defer locker.Close()
	if _, err := locker.Exec(`select pg_advisory_lock($1)`, migrationLockID); err != nil {
		t.Fatalf("lock advisory id: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, migrateErr := ApplyMigrations(db)
		done <- migrateErr
	}()

	select {
	case err := <-done:
		t.Fatalf("ApplyMigrations returned before advisory unlock: %v", err)
	case <-time.After(500 * time.Millisecond):
	}

	if _, err := locker.Exec(`select pg_advisory_unlock($1)`, migrationLockID); err != nil {
		t.Fatalf("unlock advisory id: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ApplyMigrations after unlock error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ApplyMigrations did not finish after advisory unlock")
	}
}

func TestMigrationVersionsAreUniqueAndAscending(t *testing.T) {
	migrations, err := listMigrations()
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected embedded migrations")
	}
	last := int64(0)
	for i, migration := range migrations {
		if i > 0 && migration.Version <= last {
			t.Fatalf("expected strictly increasing migration versions, saw %d after %d (%s)", migration.Version, last, migration.Name)
		}
		last = migration.Version
	}
}

func openTempDatabase(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	baseURL := strings.TrimSpace(os.Getenv("RSI_TEST_POSTGRES_URL"))
	if baseURL == "" {
		t.Skip("RSI_TEST_POSTGRES_URL not set")
	}

	admin, err := OpenPostgres(baseURL)
	if err != nil {
		t.Fatalf("open admin postgres: %v", err)
	}

	dbName := fmt.Sprintf("rsi_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(`create database ` + dbName); err != nil {
		_ = admin.Close()
		t.Fatalf("create database %s: %v", dbName, err)
	}

	testURL, err := withDatabase(baseURL, dbName)
	if err != nil {
		_ = admin.Close()
		t.Fatalf("build database URL: %v", err)
	}
	db, err := OpenPostgres(testURL)
	if err != nil {
		_, _ = admin.Exec(`drop database if exists ` + dbName)
		_ = admin.Close()
		t.Fatalf("open test postgres: %v", err)
	}

	return db, func() {
		_ = db.Close()
		_, _ = admin.Exec(`drop database if exists ` + dbName + ` with (force)`)
		_ = admin.Close()
	}
}

func withDatabase(rawURL string, dbName string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.Path = "/" + dbName
	return parsed.String(), nil
}

func assertHonchoSchemaArtifacts(t *testing.T, db *sql.DB) {
	t.Helper()

	var extensionExists bool
	if err := db.QueryRow(`select exists (select 1 from pg_extension where extname = 'vector')`).Scan(&extensionExists); err != nil {
		t.Fatalf("check vector extension: %v", err)
	}
	if !extensionExists {
		t.Fatal("expected vector extension to be installed")
	}

	var schemaExists bool
	if err := db.QueryRow(`select exists (select 1 from information_schema.schemata where schema_name = 'honcho')`).Scan(&schemaExists); err != nil {
		t.Fatalf("check honcho schema: %v", err)
	}
	if !schemaExists {
		t.Fatal("expected honcho schema to be installed")
	}

	var tableExists bool
	if err := db.QueryRow(`
		select exists (
			select 1
			from information_schema.tables
			where table_schema = 'honcho'
			  and table_name = 'messages'
		)
	`).Scan(&tableExists); err != nil {
		t.Fatalf("check honcho messages table: %v", err)
	}
	if !tableExists {
		t.Fatal("expected honcho.messages table to exist")
	}
}

func assertQuestionRunSchemaRemoved(t *testing.T, db *sql.DB) {
	t.Helper()

	var tableExists bool
	if err := db.QueryRow(`
		select exists (
			select 1
			from information_schema.tables
			where table_schema = current_schema()
			  and table_name = 'question_run'
		)
	`).Scan(&tableExists); err != nil {
		t.Fatalf("check question_run table: %v", err)
	}
	if tableExists {
		t.Fatal("expected question_run table to be absent")
	}
}

func assertNoQuestionRunTransitionRows(t *testing.T, db *sql.DB) {
	t.Helper()

	for _, table := range []string{"effect_execution", "command_receipt", "domain_event"} {
		var count int
		if err := db.QueryRow(`select count(*) from ` + table + ` where machine_kind = 'question_run'`).Scan(&count); err != nil {
			t.Fatalf("count legacy question_run rows in %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected no legacy question_run rows in %s, got %d", table, count)
		}
	}
}
