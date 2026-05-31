package harness

import (
	"encoding/json"
	"testing"
)

func TestDecodeExecutionMetadata(t *testing.T) {
	t.Parallel()

	metadata := DecodeExecutionMetadata(map[string]any{
		"hermes_session_id":         " sess-1 ",
		"parent_session_id":         " parent-1 ",
		"harness_profile_id":        " profile-1 ",
		"effective_overlay_id":      " overlay-1 ",
		"effective_overlay_version": " v1 ",
		"memory_backend":            " backend-1 ",
		"assistant_peer_id":         " assistant-1 ",
		"user_peer_id":              " user-1 ",
		"session_scope_kind":        " role ",
		"session_scope_id":          " scope-1 ",
		"parent_session_scope_kind": " parent-role ",
		"parent_session_scope_id":   " parent-scope-1 ",
		"memory_reads": []any{
			map[string]any{"kind": "trace", "summary": " one ", "ref": "ref-1"},
			map[string]any{},
		},
		"memory_writes": []any{
			map[string]any{"kind": "note", "summary": " write ", "ref": "ref-2"},
		},
	})

	if metadata.HermesSessionID != "sess-1" {
		t.Fatalf("expected trimmed hermes session id, got %q", metadata.HermesSessionID)
	}
	if metadata.ParentScopeKind != "parent-role" {
		t.Fatalf("expected trimmed parent scope kind, got %q", metadata.ParentScopeKind)
	}
	if len(metadata.MemoryReads) != 1 {
		t.Fatalf("expected blank memory read to be dropped, got %d entries", len(metadata.MemoryReads))
	}
	if metadata.MemoryReads[0].Summary != " one " {
		t.Fatalf("expected memory artifact contents to be preserved, got %+v", metadata.MemoryReads[0])
	}
	if len(metadata.MemoryWrites) != 1 {
		t.Fatalf("expected memory writes to be retained, got %d entries", len(metadata.MemoryWrites))
	}
}

func TestDialecticLevelJSONFlattensRuntimeComponent(t *testing.T) {
	t.Parallel()

	level := DialecticLevel{
		RuntimeComponent: RuntimeComponent{
			Provider:        "openai",
			Model:           "gpt-5.4",
			ReasoningEffort: "xhigh",
		},
		ThinkingBudgetTokens: 256,
	}

	data, err := json.Marshal(level)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, ok := got["runtime_component"]; ok {
		t.Fatalf("did not expect nested runtime_component key in JSON: %v", got)
	}
	if got["provider"] != "openai" || got["model"] != "gpt-5.4" || got["reasoning_effort"] != "xhigh" {
		t.Fatalf("unexpected flattened runtime component JSON: %v", got)
	}
	if got["thinking_budget_tokens"] != float64(256) {
		t.Fatalf("unexpected thinking budget tokens: %v", got["thinking_budget_tokens"])
	}
}
