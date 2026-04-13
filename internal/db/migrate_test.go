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
