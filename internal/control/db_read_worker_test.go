package control

import (
	"context"
	"errors"
	"net/url"
	"strings"
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

func TestDBReadWorkerAutoApprovesValidatedRequest(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	request, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey: "auto-approve",
		Target:         "depin-stage",
		Purpose:        "query",
		SQL:            "select 1",
		SQLSHA256:      "sha256:abc",
		Requester:      "user:U123",
		WorkflowID:     "workflow-1",
		TraceID:        "trace-1",
		ChannelID:      "C123",
		ThreadTS:       "171000001.000100",
		ExpiresAt:      now.Add(time.Hour),
		Caps:           storepkg.DBReadCaps{MaxRows: 10, MaxBytes: 1024, TimeoutSeconds: 5},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	pause, _, err := store.UpsertExternalToolPause(storepkg.ExternalToolPauseCreateInput{
		IdempotencyKey:    "pause-auto-approve",
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
	attempt, err := store.AppendDBReadValidationAttempt(storepkg.NewDBReadValidationAttempt(request, storepkg.DBReadValidationStatusSucceeded, "target_prepare", "", nil, now))
	if err != nil {
		t.Fatal(err)
	}
	pending, ok := store.GetDBReadRequest(request.ID)
	if !ok {
		t.Fatal("expected request")
	}
	poster := &fakeSlackPoster{}

	if err := autoApproveDBReadRequest(context.Background(), store, poster, pending, attempt, pending.SQL); err != nil {
		t.Fatal(err)
	}

	updated, ok := store.GetDBReadRequest(request.ID)
	if !ok {
		t.Fatal("expected request")
	}
	if updated.State != storepkg.DBReadStateApproved {
		t.Fatalf("expected approved state, got %s", updated.State)
	}
	if updated.ApprovedAt == nil {
		t.Fatal("expected ApprovedAt to be set")
	}
	if updated.ApprovedBySlackUserID != "" {
		t.Fatalf("auto-approval must not attribute a Slack approver, got %q", updated.ApprovedBySlackUserID)
	}
	if updated.SlackMessageChannelID == "" || updated.SlackMessageTS == "" {
		t.Fatal("expected audit card Slack coordinates to be recorded")
	}
	updatedPause, ok := store.GetExternalToolPause(pause.ID)
	if !ok {
		t.Fatal("expected external pause")
	}
	if updatedPause.ApprovalStatus != storepkg.ExternalToolApprovalApproved {
		t.Fatalf("pause approval status = %s, want approved", updatedPause.ApprovalStatus)
	}
	if updatedPause.ApprovalRef != "auto:read_only_validated" {
		t.Fatalf("pause approval ref = %q, want auto:read_only_validated", updatedPause.ApprovalRef)
	}
	if len(poster.calls) != 2 {
		t.Fatalf("expected audit card post and status update, got %d Slack calls", len(poster.calls))
	}
	card := poster.calls[0].values.Encode()
	if strings.Contains(card, dbReadSlackApproveAction) || strings.Contains(card, dbReadSlackDenyAction) {
		t.Fatalf("audit card must not contain approve/deny buttons: %s", card)
	}
	if !strings.Contains(card, "auto-approved") {
		t.Fatalf("audit card should state auto-approval: %s", card)
	}
	update := poster.calls[1].values.Encode()
	if !strings.Contains(update, url.QueryEscape("queued for execution")) {
		t.Fatalf("status update should mention queued execution: %s", update)
	}
}

func TestDBReadWorkerAutoApproveWithoutSlackChannel(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	request, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey: "auto-approve-no-channel",
		Target:         "depin-stage",
		Purpose:        "query",
		SQL:            "select 1",
		SQLSHA256:      "sha256:abc",
		Requester:      "hermes",
		ExpiresAt:      now.Add(time.Hour),
		Caps:           storepkg.DBReadCaps{MaxRows: 10, MaxBytes: 1024, TimeoutSeconds: 5},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := store.AppendDBReadValidationAttempt(storepkg.NewDBReadValidationAttempt(request, storepkg.DBReadValidationStatusSucceeded, "target_prepare", "", nil, now))
	if err != nil {
		t.Fatal(err)
	}
	pending, ok := store.GetDBReadRequest(request.ID)
	if !ok {
		t.Fatal("expected request")
	}

	if err := autoApproveDBReadRequest(context.Background(), store, nil, pending, attempt, pending.SQL); err != nil {
		t.Fatal(err)
	}

	updated, ok := store.GetDBReadRequest(request.ID)
	if !ok {
		t.Fatal("expected request")
	}
	if updated.State != storepkg.DBReadStateApproved {
		t.Fatalf("expected approved state, got %s", updated.State)
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
