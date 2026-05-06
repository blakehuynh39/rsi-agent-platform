package store

import "testing"

func TestCountProjectedLedgerToolCallsByTraceDedupesCanonicalAndProgress(t *testing.T) {
	canonical := map[string]hermesCanonicalToolIndex{
		"trace-1": {
			count: 1,
			ids: map[string]bool{
				"canonical-row":  true,
				"canonical-call": true,
			},
			names: map[string]bool{
				"search": true,
			},
		},
	}
	events := []hermesLedgerToolCountEvent{
		{traceID: "trace-1", id: "ledger-canonical-id", kind: "tool.call.completed", status: "completed", name: "search", stableID: "canonical-call"},
		{traceID: "trace-1", id: "ledger-canonical-name", kind: "tool.start", status: "running", name: "search"},
		{traceID: "trace-1", id: "ledger-write-start", kind: "tool.start", status: "running", name: "write", stableID: "ledger-write"},
		{traceID: "trace-1", id: "ledger-write-progress", kind: "tool.progress", status: "running", name: "write", stableID: "ledger-write"},
		{traceID: "trace-1", id: "ledger-write-complete", kind: "tool.complete", status: "completed", name: "write", stableID: "ledger-write"},
		{traceID: "trace-1", id: "ledger-plan-progress", kind: "tool.progress", status: "running", name: "plan"},
		{traceID: "trace-1", id: "ledger-plan-complete", kind: "tool.complete", status: "completed", name: "plan"},
		{traceID: "trace-2", id: "ledger-read-start", kind: "tool.start", status: "running", name: "read"},
		{traceID: "trace-2", id: "ledger-read-progress", kind: "tool.progress", status: "running", name: "read"},
	}

	counts := countProjectedLedgerToolCallsByTrace(events, canonical)

	if counts["trace-1"] != 2 {
		t.Fatalf("trace-1 projected count = %d, want 2", counts["trace-1"])
	}
	if counts["trace-2"] != 1 {
		t.Fatalf("trace-2 projected count = %d, want 1", counts["trace-2"])
	}
}
