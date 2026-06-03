package store

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPostgresInsertStatementsHaveMatchingColumnsAndValues(t *testing.T) {
	body, err := readPostgresSource(t)
	if err != nil {
		t.Fatalf("read postgres.go: %v", err)
	}
	matches := extractInsertStatements(string(body))
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

func TestPostgresProposalInsertKeepsTouchedFilesOnJSONColumn(t *testing.T) {
	body, err := readPostgresSource(t)
	if err != nil {
		t.Fatalf("read postgres.go: %v", err)
	}
	for _, match := range extractInsertStatements(string(body)) {
		if match[1] != "proposal" {
			continue
		}
		columns := splitSQLList(match[2])
		values := splitSQLList(match[3])
		columnIndex := indexOfSQLToken(columns, "touched_files")
		if columnIndex == -1 {
			t.Fatal("proposal insert missing touched_files column")
		}
		if columnIndex >= len(values) {
			t.Fatalf("proposal insert value missing for touched_files at index %d", columnIndex)
		}
		if !strings.Contains(strings.ToLower(values[columnIndex]), "::jsonb") {
			t.Fatalf("proposal.touched_files must bind through jsonb cast, got %q", values[columnIndex])
		}
		validationIndex := indexOfSQLToken(columns, "validation_plan")
		if validationIndex == -1 {
			t.Fatal("proposal insert missing validation_plan column")
		}
		if validationIndex >= len(values) {
			t.Fatalf("proposal insert value missing for validation_plan at index %d", validationIndex)
		}
		if strings.Contains(strings.ToLower(values[validationIndex]), "::jsonb") {
			t.Fatalf("proposal.validation_plan must not bind through jsonb cast, got %q", values[validationIndex])
		}
		return
	}
	t.Fatal("proposal insert statement not found")
}

func TestPostgresLoaderDoesNotInvokeLegacyBackfills(t *testing.T) {
	body, err := readPostgresSource(t)
	if err != nil {
		t.Fatalf("read postgres.go: %v", err)
	}
	source := string(body)
	for _, pattern := range []string{
		"backfillConversationCaseV2(",
		"backfillActionOutcomeKnowledgeV3(",
	} {
		if strings.Contains(source, pattern) {
			t.Fatalf("expected postgres loader to avoid legacy bootstrap backfills, found %q", pattern)
		}
	}
}

func TestHermesSessionsPageAvoidsGlobalLatestLedgerCTE(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("hermes_sessions.go"))
	if err != nil {
		t.Fatalf("read hermes_sessions.go: %v", err)
	}
	source := string(body)
	if strings.Contains(source, "latest_trace_ledger") {
		t.Fatal("sessions page query must not rebuild a global latest_trace_ledger CTE")
	}
	for _, pattern := range []string{
		"recentHermesLedgerTraceCandidates(window)",
		"jsonb_to_recordset($4::jsonb)",
		"base_trace_candidates",
	} {
		if !strings.Contains(source, pattern) {
			t.Fatalf("sessions page query missing bounded candidate pattern %q", pattern)
		}
	}
}

func readPostgresSource(t *testing.T) ([]byte, error) {
	t.Helper()
	return os.ReadFile(filepath.Join("postgres.go"))
}

func extractInsertStatements(body string) [][]string {
	re := regexp.MustCompile(`(?is)insert into\s+([a-z_]+)\s*\((.*?)\)\s*values\s*\((.*?)\)`)
	return re.FindAllStringSubmatch(body, -1)
}

func indexOfSQLToken(items []string, target string) int {
	for i, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), target) {
			return i
		}
	}
	return -1
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
