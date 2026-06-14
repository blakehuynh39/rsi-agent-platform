package dbread

import (
	"strings"
	"testing"
)

func TestTargetBuildsDSNFromParts(t *testing.T) {
	t.Setenv("DB_HOST", "example.rds.amazonaws.com")
	t.Setenv("DB_USER", "read user")
	t.Setenv("DB_PASSWORD", "p@ss/word")

	registry, err := LoadRegistry(`{"targets":[{"id":"story-api-stage","dsn_parts":{"host_env":"DB_HOST","database":"appdb","user_env":"DB_USER","password_env":"DB_PASSWORD","sslmode":"verify-full"}}]}`)
	if err != nil {
		t.Fatal(err)
	}
	target, ok := registry.Target("story-api-stage")
	if !ok {
		t.Fatal("target not found")
	}
	want := "postgres://read%20user:p%40ss%2Fword@example.rds.amazonaws.com:5432/appdb?sslmode=verify-full"
	if target.DSN != want {
		t.Fatalf("DSN = %q, want %q", target.DSN, want)
	}
}

func TestTargetBuildsDSNFromPartsWithPortEnv(t *testing.T) {
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_SSLMODE", "require")

	registry, err := LoadRegistry(`{"targets":[{"id":"story-api-stage","dsn_parts":{"host":"example.rds.amazonaws.com","port_env":"DB_PORT","database":"appdb","user":"readonly","password":"secret","sslmode_env":"DB_SSLMODE"}}]}`)
	if err != nil {
		t.Fatal(err)
	}
	target, ok := registry.Target("story-api-stage")
	if !ok {
		t.Fatal("target not found")
	}
	want := "postgres://readonly:secret@example.rds.amazonaws.com:6543/appdb?sslmode=require"
	if target.DSN != want {
		t.Fatalf("DSN = %q, want %q", target.DSN, want)
	}
}

func TestTargetDSNPartsMissingRequiredComponentReturnsEmptyDSN(t *testing.T) {
	registry, err := LoadRegistry(`{"targets":[{"id":"missing-password","dsn_parts":{"host":"example.rds.amazonaws.com","database":"appdb","user":"readonly"}}]}`)
	if err != nil {
		t.Fatal(err)
	}
	target, ok := registry.Target("missing-password")
	if !ok {
		t.Fatal("target not found")
	}
	if target.DSN != "" {
		t.Fatalf("DSN = %q, want empty", target.DSN)
	}
}

func TestPublicSourcesHideDSNParts(t *testing.T) {
	registry, err := LoadRegistry(`{"targets":[{"id":"stage","dsn_parts":{"host":"example","database":"db","user":"u","password":"p"}}]}`)
	if err != nil {
		t.Fatal(err)
	}
	sources := registry.PublicSources()
	if len(sources) != 1 {
		t.Fatalf("sources len = %d, want 1", len(sources))
	}
	for key := range sources[0] {
		if strings.Contains(key, "dsn") {
			t.Fatalf("public sources should not expose DSN config, got key %q in %#v", key, sources[0])
		}
	}
}
