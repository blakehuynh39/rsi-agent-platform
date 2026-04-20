package toolcatalog

import "testing"

func TestImprovementReadOnlyToolNamesIncludesSlackTranscriptTools(t *testing.T) {
	tools := ImprovementReadOnlyToolNames()
	if !containsTool(tools, "slack.history") {
		t.Fatalf("expected slack.history in improvement read-only tools, got %#v", tools)
	}
	if !containsTool(tools, "slack.search") {
		t.Fatalf("expected slack.search in improvement read-only tools, got %#v", tools)
	}
}

func containsTool(tools []string, wanted string) bool {
	for _, tool := range tools {
		if tool == wanted {
			return true
		}
	}
	return false
}
