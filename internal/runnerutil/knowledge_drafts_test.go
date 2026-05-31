package runnerutil

import (
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
)

func TestNormalizeKnowledgeDraftRejectsInvalidOrUngroundedDrafts(t *testing.T) {
	t.Parallel()

	tests := []KnowledgeDraft{
		{
			Kind:         "eval_summary",
			ScopeType:    "trace",
			Title:        "Architecture-review pass",
			Summary:      "Trace-local eval recap.",
			EvidenceRefs: []events.EvidenceRef{{Kind: "trace", Ref: "trace-1"}},
		},
		{
			Kind:      string(knowledge.KindFact),
			ScopeType: string(knowledge.ScopeCase),
		},
		{
			Kind:      string(knowledge.KindFact),
			ScopeType: string(knowledge.ScopeCase),
			Title:     "Ungrounded draft",
			Summary:   "No refs.",
		},
	}

	for _, item := range tests {
		if _, ok := NormalizeKnowledgeDraft(item, knowledge.ScopeCase, "case-1"); ok {
			t.Fatalf("expected draft %+v to be rejected", item)
		}
	}
}

func TestNormalizeKnowledgeDraftAppliesDefaultsAndTrimsFields(t *testing.T) {
	t.Parallel()

	item := KnowledgeDraft{
		Title:   "  Useful fact  ",
		Summary: "  Grounded summary. ",
		EvidenceRefs: []events.EvidenceRef{
			{Kind: "trace", Ref: " trace-1 ", Summary: " summary "},
			{Kind: "", Ref: "missing-kind"},
		},
	}

	got, ok := NormalizeKnowledgeDraft(item, knowledge.ScopeCase, "case-1")
	if !ok {
		t.Fatal("expected draft to normalize")
	}
	if got.Kind != string(knowledge.KindFact) {
		t.Fatalf("expected default fact kind, got %q", got.Kind)
	}
	if got.ScopeType != string(knowledge.ScopeCase) {
		t.Fatalf("expected default case scope, got %q", got.ScopeType)
	}
	if got.ScopeID != "case-1" {
		t.Fatalf("expected default scope id, got %q", got.ScopeID)
	}
	if got.Title != "Useful fact" || got.Summary != "Grounded summary." {
		t.Fatalf("expected trimmed content, got %+v", got)
	}
	if len(got.EvidenceRefs) != 1 || got.EvidenceRefs[0].Ref != "trace-1" {
		t.Fatalf("expected normalized evidence refs, got %+v", got.EvidenceRefs)
	}
}

func TestKnowledgeDraftAggregateIDIsStableAcrossRetries(t *testing.T) {
	t.Parallel()

	entry := knowledge.Entry{
		Kind:      knowledge.KindFact,
		ScopeType: knowledge.ScopeCase,
		ScopeID:   "case-1",
		Title:     "First title",
	}
	first := knowledgeDraftAggregateID(entry, "trace-1", 0, time.Unix(10, 0).UTC())

	entry.Title = "Updated title"
	second := knowledgeDraftAggregateID(entry, "trace-1", 0, time.Unix(20, 0).UTC())
	if first != second {
		t.Fatalf("expected stable knowledge draft id across retries, got %q and %q", first, second)
	}
}
