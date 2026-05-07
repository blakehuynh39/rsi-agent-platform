package control

import (
	"context"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/dbread"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestDBReadWorkerExpiresApprovedRequestBeforeExecution(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	request, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey: "expired-approved",
		Target:         "depin-stage",
		Purpose:        "query",
		SQL:            "select 1",
		SQLSHA256:      "sha256:abc",
		Requester:      "hermes",
		ExpiresAt:      now.Add(-time.Minute),
		Caps:           storepkg.DBReadCaps{MaxRows: 10, MaxBytes: 1024, TimeoutSeconds: 5},
	}, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	attempt := storepkg.NewDBReadValidationAttempt(request, storepkg.DBReadValidationStatusSucceeded, "target_prepare", "", nil, now.Add(-30*time.Minute))
	if _, err := store.AppendDBReadValidationAttempt(attempt); err != nil {
		t.Fatal(err)
	}
	if _, err := store.TransitionDBReadRequest(request.ID, storepkg.DBReadStatePendingApproval, storepkg.DBReadStateApproved, func(item *storepkg.DBReadRequest) error {
		item.ApprovedBySlackUserID = "UADMIN"
		item.ApprovedAt = &now
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	lease, ok, err := store.ClaimNextDBReadRequest("worker", time.Minute, now, []string{"depin-stage"})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected expired approved request to be claimable for expiry handling")
	}
	registry, err := dbread.LoadRegistry(`{"targets":[{"id":"depin-stage","allowed_schemas":["public"],"allowed_tables":["*"]}]}`)
	if err != nil {
		t.Fatal(err)
	}

	handleDBReadLease(context.Background(), store, registry, nil, lease)

	updated, ok := store.GetDBReadRequest(request.ID)
	if !ok {
		t.Fatal("expected request")
	}
	if updated.State != storepkg.DBReadStateExpired {
		t.Fatalf("expected expired state, got %s", updated.State)
	}
	if updated.LeaseHolder != "" || updated.LeaseToken != "" || updated.LeaseExpiresAt != nil {
		t.Fatalf("expected expiration to clear lease fields")
	}
	if results := store.ListDBReadExecutionResults(request.ID); len(results) != 0 {
		t.Fatalf("expected no execution result for expiry transition, got %d", len(results))
	}
}
