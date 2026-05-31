package improvementplane

import (
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestBuildHarnessOverlayFromRunner(t *testing.T) {
	mem := store.NewMemoryStore()
	proposal := review.Proposal{
		ID:           "proposal-overlay-001",
		CandidateKey: "prod:behavioral-regression",
		TargetLayer:  harness.TargetLayerHarnessOverlay,
		TargetKind:   "runner_role",
		TargetRef:    "prod",
		Summary:      "Tighten the conversational role prompt and retrieval bias.",
	}
	output := runnerutil.StructuredOutput{
		ProposedActions: []runnerutil.ProposedAction{
			{
				Kind: string(action.KindHarnessOverlay),
				RequestPayload: map[string]any{
					"version":               "proposal-overlay-001",
					"prompt_fragments":      []any{"State uncertainty before tool calls.", "Prefer canonical RSI knowledge before Honcho recall."},
					"tool_preference_order": []any{"knowledge.context", "repo.context", "github.repo_activity"},
					"retrieval_bias":        "canonical_first",
					"reasoning_verbosity":   "verbose",
					"memory_read_enabled":   true,
					"memory_write_enabled":  true,
				},
			},
		},
	}

	overlay, err := buildHarnessOverlayFromRunner(mem, proposal, output)
	if err != nil {
		t.Fatalf("buildHarnessOverlayFromRunner() error = %v", err)
	}
	if overlay.Role != "prod" {
		t.Fatalf("expected prod role, got %s", overlay.Role)
	}
	if overlay.ProfileID != harness.DefaultProfileID("prod") {
		t.Fatalf("expected default prod profile, got %s", overlay.ProfileID)
	}
	if overlay.Status != harness.OverlayStatusActive {
		t.Fatalf("expected active overlay, got %s", overlay.Status)
	}
	if len(overlay.PromptFragments) != 2 {
		t.Fatalf("expected prompt fragments to persist, got %+v", overlay.PromptFragments)
	}
	if overlay.MemoryReadEnabled == nil || !*overlay.MemoryReadEnabled {
		t.Fatal("expected memory_read_enabled=true")
	}
	if overlay.MemoryWriteEnabled == nil || !*overlay.MemoryWriteEnabled {
		t.Fatal("expected memory_write_enabled=true")
	}
}

func TestBuildHarnessOverlayFromRunnerRequiresOverlayAction(t *testing.T) {
	mem := store.NewMemoryStore()
	proposal := review.Proposal{
		ID:           "proposal-overlay-002",
		CandidateKey: "prod:behavioral-regression",
		TargetLayer:  harness.TargetLayerHarnessOverlay,
		TargetKind:   "runner_role",
		TargetRef:    "prod",
	}

	_, err := buildHarnessOverlayFromRunner(mem, proposal, runnerutil.StructuredOutput{})
	if err == nil {
		t.Fatal("expected error when overlay action is missing")
	}
}
