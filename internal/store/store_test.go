package store

import (
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

func TestMemoryStoreRatingAndReplay(t *testing.T) {
	store := NewMemoryStore()
	traces := store.ListTraces()
	if len(traces) == 0 {
		t.Fatal("expected seeded traces")
	}
	traceID := traces[0].TraceID

	if _, err := store.AddRating(traceID, review.HumanRating{
		Score:      4,
		Verdict:    "partial",
		Labels:     []string{"needs-human"},
		Notes:      "Useful investigation, incomplete mitigation.",
		ReviewerID: "alice",
	}); err != nil {
		t.Fatalf("AddRating() error = %v", err)
	}

	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.LastVerdict != "partial" {
		t.Fatalf("expected last verdict to be updated, got %q", trace.Summary.LastVerdict)
	}

	item, err := store.ScheduleReplay(traceID, "alice")
	if err != nil {
		t.Fatalf("ScheduleReplay() error = %v", err)
	}
	if item.TraceID != traceID {
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

func TestProposalCapEnforced(t *testing.T) {
	store := NewMemoryStore()

	slots := store.GetProposalSlots()
	if slots.Active < 2 {
		store.proposals["proposal-test-cap"] = review.Proposal{
			ID:                  "proposal-test-cap",
			TraceID:             store.ListTraces()[0].TraceID,
			Title:               "Synthetic cap filler",
			Category:            "policy_or_runtime_fix",
			Summary:             "Fill the second slot for cap enforcement.",
			Status:              review.ProposalPendingReview,
			CandidateKey:        "synthetic:incident:failed_workflow",
			ActiveSlotConsuming: true,
			CreatedAt:           time.Now().UTC(),
		}
		slots = store.GetProposalSlots()
	}
	if slots.Active != 2 {
		t.Fatalf("expected 2 active slots for cap test, got %d", slots.Active)
	}

	result, err := store.RunProposalPromoter("test-promoter")
	if err != nil {
		t.Fatalf("RunProposalPromoter() error = %v", err)
	}
	if !result.BlockedByCap {
		t.Fatal("expected promoter to be blocked by the slot cap")
	}
	if result.Promoted != 0 {
		t.Fatalf("expected no new proposals, got %d", result.Promoted)
	}
}

func TestProposalPromoterLease(t *testing.T) {
	store := NewMemoryStore()

	store.cronLeases["improvement-plane-cron"] = improvement.CronLease{
		Name:      "improvement-plane-cron",
		Holder:    "other-worker",
		ExpiresAt: time.Now().UTC().Add(time.Minute),
	}

	if _, err := store.RunProposalPromoter("test-worker"); err == nil {
		t.Fatal("expected promoter lease conflict")
	}
}

func TestRejectedProposalRequiresNewEvidence(t *testing.T) {
	store := NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposals")
	}
	proposal := proposals[0]

	if _, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalRejected),
		Rationale:  "Too similar to prior attempt.",
		ReviewerID: "alice",
	}); err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}

	candidate := store.candidates[proposal.CandidateKey]
	candidate.Status = improvement.CandidateQueued
	candidate.NewEvidenceSinceLastRejection = false
	store.candidates[proposal.CandidateKey] = candidate

	result, err := store.PromoteCandidates("alice", 2)
	if err != nil {
		t.Fatalf("PromoteCandidates() error = %v", err)
	}
	for _, promotedID := range result.PromotedIDs {
		if store.proposals[promotedID].CandidateKey == proposal.CandidateKey {
			t.Fatal("expected rejected candidate to stay blocked without new evidence")
		}
	}
}

func TestSettingsBackedProposalCap(t *testing.T) {
	store := NewMemoryStore()

	settings, err := store.UpdateSettings(improvement.Settings{ActiveProposalCap: 1})
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if settings.ActiveProposalCap != 1 {
		t.Fatalf("expected active proposal cap to be 1, got %d", settings.ActiveProposalCap)
	}

	slots := store.GetProposalSlots()
	if slots.Cap != 1 {
		t.Fatalf("expected slot cap to be 1, got %d", slots.Cap)
	}
}

func TestApproveProposalQueuesMaterializationWork(t *testing.T) {
	store := NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposals")
	}

	proposal, err := store.ReviewProposal(proposals[0].ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if proposal.Status != review.ProposalApproved {
		t.Fatalf("expected approved proposal, got %s", proposal.Status)
	}

	items := store.ListWorkItems()
	found := false
	for _, item := range items {
		if item.Queue == queue.ProposalQueue && item.ProposalID == proposal.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected proposal materialization work item to be queued")
	}
}
