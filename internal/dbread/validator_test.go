package dbread

import (
	"strings"
	"testing"
)

func TestValidateSQLSafetyAllowsReadOnlySelect(t *testing.T) {
	cases := []string{
		"select id, status from orders where created_at > now() - interval '1 hour'",
		"with recent as (select id from orders) select count(*) from recent",
		"select 'insert update delete pg_sleep(1) pg_authid' as literal_only",
	}
	for _, sql := range cases {
		got := ValidateSQLSafety(sql)
		if !got.OK {
			t.Fatalf("expected SQL to be allowed: %s: %s %s", sql, got.ErrorCode, got.Message)
		}
		if got.SQLSHA256 == "" {
			t.Fatalf("expected hash")
		}
	}
}

func TestTruncateSQLHonorsMax(t *testing.T) {
	got := truncateSQL("select "+strings.Repeat("x", 100), 40)
	if len(got) != 40 {
		t.Fatalf("expected truncated SQL length 40, got %d", len(got))
	}
	if !strings.HasSuffix(got, "\n-- truncated --") {
		t.Fatalf("expected truncation suffix")
	}
}

func TestValidateSQLSafetyRejectsUnsafeSQL(t *testing.T) {
	cases := map[string]string{
		"insert into users(id) values (1)":                                       "not_select",
		"select * into temp t from users":                                        "select_into",
		"select * from users for update":                                         "row_lock",
		"with deleted as (delete from users returning id) select * from deleted": "unsafe_ast",
		"select pg_sleep(10)":                                                    "unsafe_function",
		"select dblink_exec('dbname=postgres', 'delete from users')":             "unsafe_function",
		"select public.dblink_connect('dbname=postgres')":                        "unsafe_function",
		"select * from pg_authid":                                                "unsafe_catalog",
		"select 1; select 2":                                                     "multi_statement",
	}
	for sql, code := range cases {
		got := ValidateSQLSafety(sql)
		if got.OK {
			t.Fatalf("expected SQL to be rejected: %s", sql)
		}
		if got.ErrorCode != code {
			t.Fatalf("expected %s for %q, got %s (%s)", code, sql, got.ErrorCode, got.Message)
		}
	}
}
