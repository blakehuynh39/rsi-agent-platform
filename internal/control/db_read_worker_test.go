package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
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

	handleDBReadLease(context.Background(), config.Config{}, store, registry, nil, nil, lease)

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

func TestDBReadValidationExpiryTransitionFailureMarksExternalPauseFailed(t *testing.T) {
	base := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	request, _, err := base.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey: "expired-validation-transition-failure",
		Target:         "depin-stage",
		Purpose:        "query",
		SQL:            "select 1",
		SQLSHA256:      "sha256:abc",
		Requester:      "hermes",
		WorkflowID:     "workflow-1",
		TraceID:        "trace-1",
		ExpiresAt:      now.Add(-time.Minute),
		Caps:           storepkg.DBReadCaps{MaxRows: 10, MaxBytes: 1024, TimeoutSeconds: 5},
	}, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	pause, _, err := base.UpsertExternalToolPause(storepkg.ExternalToolPauseCreateInput{
		IdempotencyKey:    "pause-expired-validation-transition-failure",
		WorkflowID:        request.WorkflowID,
		TraceID:           request.TraceID,
		HermesSessionID:   "session-1",
		CanonicalToolName: "db_read.query",
		TransportToolName: "db_read_query",
		ToolCallID:        "call_db_1",
		DBReadRequestID:   request.ID,
		SQLSHA256:         request.SQLSHA256,
		ExpiresAt:         request.ExpiresAt,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := dbread.LoadRegistry(`{"targets":[{"id":"depin-stage","allowed_schemas":["public"],"allowed_tables":["*"]}]}`)
	if err != nil {
		t.Fatal(err)
	}
	store := failingDBReadTransitionStore{Store: base, requestID: request.ID, err: errors.New("forced transition failure")}
	handleDBReadValidationLease(context.Background(), config.Config{}, store, registry, nil, nil, storepkg.DBReadLease{Request: request})

	updated, ok := base.GetExternalToolPause(pause.ID)
	if !ok {
		t.Fatal("expected external pause")
	}
	if updated.ToolOutcome != storepkg.ExternalToolOutcomeFailed {
		t.Fatalf("tool outcome = %s, want failed", updated.ToolOutcome)
	}
	if updated.ErrorMessage == "" {
		t.Fatalf("expected error message to be recorded: %+v", updated)
	}
}

func TestDBReadWorkerDoesNotExecuteWithoutExternalToolPause(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	request, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey: "legacy-approved",
		Target:         "depin-stage",
		Purpose:        "query",
		SQL:            "select 1",
		SQLSHA256:      "sha256:abc",
		Requester:      "hermes",
		ExpiresAt:      now.Add(time.Hour),
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
		t.Fatal("expected approved request to be claimable")
	}
	registry, err := dbread.LoadRegistry(`{"targets":[{"id":"depin-stage","allowed_schemas":["public"],"allowed_tables":["*"]}]}`)
	if err != nil {
		t.Fatal(err)
	}

	handleDBReadLease(context.Background(), config.Config{}, store, registry, nil, nil, lease)

	updated, ok := store.GetDBReadRequest(request.ID)
	if !ok {
		t.Fatal("expected request")
	}
	if updated.State != storepkg.DBReadStateApproved {
		t.Fatalf("expected request to remain approved until external pause is ready, got %s", updated.State)
	}
	if results := store.ListDBReadExecutionResults(request.ID); len(results) != 0 {
		t.Fatalf("legacy request should not execute: %#v", results)
	}
}

type failingDBReadTransitionStore struct {
	storepkg.Store
	requestID string
	err       error
}

func (s failingDBReadTransitionStore) TransitionDBReadRequest(requestID string, from storepkg.DBReadState, to storepkg.DBReadState, mutate func(*storepkg.DBReadRequest) error) (storepkg.DBReadRequest, error) {
	if requestID == s.requestID {
		return storepkg.DBReadRequest{}, s.err
	}
	return s.Store.TransitionDBReadRequest(requestID, from, to, mutate)
}
