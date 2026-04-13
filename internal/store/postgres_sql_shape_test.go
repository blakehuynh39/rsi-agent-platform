package store

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPostgresInsertStatementsHaveMatchingColumnsAndValues(t *testing.T) {
	path := filepath.Join("postgres.go")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read postgres.go: %v", err)
	}
	re := regexp.MustCompile(`(?is)insert into\s+([a-z_]+)\s*\((.*?)\)\s*values\s*\((.*?)\)`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		t.Fatal("expected insert statements in postgres.go")
	}
	for _, match := range matches {
		table := match[1]
		columns := splitSQLList(match[2])
		values := splitSQLList(match[3])
		if len(columns) != len(values) {
			t.Fatalf("insert shape mismatch for %s: columns=%d values=%d", table, len(columns), len(values))
		}
	}
}

func splitSQLList(input string) []string {
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
