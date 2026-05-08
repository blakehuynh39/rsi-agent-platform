package store

import (
	"testing"
	"time"
)

func TestDBReadStateTransitions(t *testing.T) {
	legal := [][2]DBReadState{
		{DBReadStateValidating, DBReadStatePendingApproval},
		{DBReadStateValidating, DBReadStateValidationFailed},
		{DBReadStateValidating, DBReadStateExpired},
		{DBReadStateValidationFailed, DBReadStateValidating},
		{DBReadStateValidationFailed, DBReadStateExpired},
		{DBReadStatePendingApproval, DBReadStateApproved},
		{DBReadStatePendingApproval, DBReadStateDenied},
		{DBReadStatePendingApproval, DBReadStateExpired},
		{DBReadStateApproved, DBReadStateExpired},
		{DBReadStateApproved, DBReadStateExecuting},
		{DBReadStateExecuting, DBReadStateSucceeded},
		{DBReadStateExecuting, DBReadStateFailed},
	}
	for _, pair := range legal {
		if err := ValidateDBReadStateTransition(pair[0], pair[1]); err != nil {
			t.Fatalf("expected %s -> %s to be legal: %v", pair[0], pair[1], err)
		}
	}
	illegal := [][2]DBReadState{
		{DBReadStatePendingApproval, DBReadStateExecuting},
		{DBReadStateDenied, DBReadStateApproved},
		{DBReadStateSucceeded, DBReadStateExecuting},
	}
	for _, pair := range illegal {
		if err := ValidateDBReadStateTransition(pair[0], pair[1]); err == nil {
			t.Fatalf("expected %s -> %s to be illegal", pair[0], pair[1])
		}
	}
}

func TestMemoryDBReadLifecycleAndIdempotency(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	input := DBReadCreateInput{
		IdempotencyKey: "conversation:thread:depin-stage:sha256:abc:hermes:query",
		Target:         "depin-stage",
		Purpose:        "query",
		SQL:            "select 1",
		SQLSHA256:      "sha256:abc",
		Requester:      "hermes",
		ConversationID: "conversation",
		ThreadTS:       "thread",
		ExpiresAt:      now.Add(time.Hour),
		Caps:           DBReadCaps{MaxRows: 10, MaxBytes: 1024, TimeoutSeconds: 5},
	}
	request, created, err := store.UpsertDBReadRequest(input, now)
	if err != nil {
		t.Fatal(err)
	}
	if !created || request.State != DBReadStateValidating {
		t.Fatalf("unexpected create result created=%t state=%s", created, request.State)
	}
	again, created, err := store.UpsertDBReadRequest(input, now)
	if err != nil {
		t.Fatal(err)
	}
	if created || again.ID != request.ID {
		t.Fatalf("expected idempotent upsert to return existing request")
	}
	attempt := NewDBReadValidationAttempt(request, DBReadValidationStatusSucceeded, "target_prepare", "", nil, now)
	if _, err := store.AppendDBReadValidationAttempt(attempt); err != nil {
		t.Fatal(err)
	}
	request, _ = store.GetDBReadRequest(request.ID)
	if request.State != DBReadStatePendingApproval {
		t.Fatalf("expected pending approval, got %s", request.State)
	}
	request, err = store.TransitionDBReadRequest(request.ID, DBReadStatePendingApproval, DBReadStateApproved, func(item *DBReadRequest) error {
		item.ApprovedBySlackUserID = "UADMIN"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	lease, ok, err := store.ClaimNextDBReadRequest("worker", time.Minute, now, []string{"depin-stage"})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || lease.Request.ID != request.ID || lease.Token == "" {
		t.Fatalf("expected lease for request")
	}
	result := NewDBReadExecutionResult(lease.Request, DBReadExecutionStatusSucceeded, []map[string]string{{"?column?": "1"}}, now)
	result.LeaseToken = lease.Token
	result.RowCount = 1
	if _, err := store.AppendDBReadExecutionResult(result); err != nil {
		t.Fatal(err)
	}
	request, _ = store.GetDBReadRequest(request.ID)
	if request.State != DBReadStateSucceeded || request.RowCount != 1 {
		t.Fatalf("unexpected terminal state=%s rows=%d", request.State, request.RowCount)
	}
}

func TestMemoryDBReadExecutionScopeAllowsDistinctToolRequests(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	first, created, err := store.UpsertDBReadRequest(DBReadCreateInput{
		IdempotencyKey:    "first",
		Target:            "depin-prod",
		Purpose:           "query",
		SQL:               "select count(*) from scripts",
		SQLSHA256:         "sha256:first",
		ExecutionScopeKey: "workflow:wf-1",
		Requester:         "user:U123",
		ExpiresAt:         now.Add(time.Hour),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatalf("expected first request to be created")
	}

	second, created, err := store.UpsertDBReadRequest(DBReadCreateInput{
		IdempotencyKey:    "second",
		Target:            "depin-prod",
		Purpose:           "query",
		SQL:               "select language_code, count(*) from scripts group by language_code",
		SQLSHA256:         "sha256:second",
		ExecutionScopeKey: "workflow:wf-1",
		Requester:         "user:U123",
		ExpiresAt:         now.Add(time.Hour),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if !created || second.ID == first.ID {
		t.Fatalf("expected distinct scoped request to be created, created=%t first=%s second=%s", created, first.ID, second.ID)
	}
}

func TestMemoryDBReadValidationFailureDoesNotBlockRepairRequest(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	first, created, err := store.UpsertDBReadRequest(DBReadCreateInput{
		IdempotencyKey:    "first-invalid",
		Target:            "depin-prod",
		Purpose:           "query",
		SQL:               "select * from missing_table",
		SQLSHA256:         "sha256:invalid",
		ExecutionScopeKey: "workflow:wf-1",
		Requester:         "user:U123",
		ExpiresAt:         now.Add(time.Hour),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatalf("expected first request to be created")
	}
	_, err = store.AppendDBReadValidationAttempt(NewDBReadValidationAttempt(first, DBReadValidationStatusFailed, "offline_parse", "missing table", map[string]interface{}{"error_code": "missing_table"}, now))
	if err != nil {
		t.Fatal(err)
	}

	second, created, err := store.UpsertDBReadRequest(DBReadCreateInput{
		IdempotencyKey:    "second-valid",
		Target:            "depin-prod",
		Purpose:           "query",
		SQL:               "select count(*) from scripts",
		SQLSHA256:         "sha256:valid",
		ExecutionScopeKey: "workflow:wf-1",
		Requester:         "user:U123",
		ExpiresAt:         now.Add(time.Hour),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if !created || second.ID == first.ID {
		t.Fatalf("expected repaired SQL to create a new request, created=%t first=%s second=%s", created, first.ID, second.ID)
	}
}

func TestMemoryExpireDBReadRequestsCoversValidationStates(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	expiredAt := now.Add(-time.Minute)
	states := []DBReadState{
		DBReadStateValidating,
		DBReadStateValidationFailed,
		DBReadStatePendingApproval,
		DBReadStateApproved,
	}
	for i, state := range states {
		request, _, err := store.UpsertDBReadRequest(DBReadCreateInput{
			IdempotencyKey: "expiry-test-" + string(rune('a'+i)),
			Target:         "depin-stage",
			Purpose:        "query",
			SQL:            "select 1",
			SQLSHA256:      "sha256:abc",
			Requester:      "hermes",
			ExpiresAt:      expiredAt,
		}, expiredAt.Add(-time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		request.State = state
		request.LeaseHolder = "worker"
		request.LeaseToken = "lease"
		request.LeaseExpiresAt = &now
		store.dbReadRequests[request.ID] = request
	}

	expired, err := store.ExpirePendingDBReadRequests(now)
	if err != nil {
		t.Fatal(err)
	}
	if len(expired) != len(states) {
		t.Fatalf("expected %d expired requests, got %d", len(states), len(expired))
	}
	for _, request := range expired {
		if request.State != DBReadStateExpired {
			t.Fatalf("expected expired state, got %s", request.State)
		}
		if request.LeaseHolder != "" || request.LeaseToken != "" || request.LeaseExpiresAt != nil {
			t.Fatalf("expected expiration to clear lease fields")
		}
	}
}
