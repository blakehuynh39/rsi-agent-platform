package platform

import (
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

func TestMemoryStoreRatingAndReplay(t *testing.T) {
	store := NewMemoryStore()

	if _, err := store.AddRating("trace-oncall-001", review.HumanRating{
		Score:      4,
		Verdict:    "partial",
		Labels:     []string{"needs-human"},
		Notes:      "Useful investigation, incomplete mitigation.",
		ReviewerID: "alice",
	}); err != nil {
		t.Fatalf("AddRating() error = %v", err)
	}

	trace, ok := store.GetTrace("trace-oncall-001")
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.LastVerdict != "partial" {
		t.Fatalf("expected last verdict to be updated, got %q", trace.Summary.LastVerdict)
	}

	item, err := store.ScheduleReplay("trace-oncall-001", "alice")
	if err != nil {
		t.Fatalf("ScheduleReplay() error = %v", err)
	}
	if item.TraceID != "trace-oncall-001" {
		t.Fatalf("unexpected trace id: %s", item.TraceID)
	}
}

func TestMemoryStoreSetThreadState(t *testing.T) {
	store := NewMemoryStore()

	item, err := store.SetThreadState("slack:CENG:171000001.000100", policy.ThreadStateMuted, "")
	if err != nil {
		t.Fatalf("SetThreadState() error = %v", err)
	}
	if !item.Muted {
		t.Fatal("expected muted flag to be set")
	}
}

