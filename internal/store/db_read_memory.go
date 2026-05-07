package store

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *MemoryStore) ListDBReadRequests() []DBReadRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]DBReadRequest, 0, len(s.dbReadRequests))
	for _, item := range s.dbReadRequests {
		out = append(out, cloneDBReadRequest(item))
	}
	SortDBReadRequests(out)
	return out
}

func (s *MemoryStore) ListDBReadRequestsByScope(conversationID string, workflowID string, traceID string, channelID string, threadTS string, notBefore time.Time) []DBReadRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conversationID = strings.TrimSpace(conversationID)
	workflowID = strings.TrimSpace(workflowID)
	traceID = strings.TrimSpace(traceID)
	channelID = strings.TrimSpace(channelID)
	threadTS = strings.TrimSpace(threadTS)
	out := make([]DBReadRequest, 0)
	for _, item := range s.dbReadRequests {
		if !notBefore.IsZero() && item.CreatedAt.Before(notBefore) {
			continue
		}
		matched := false
		if workflowID != "" && strings.TrimSpace(item.WorkflowID) == workflowID {
			matched = true
		} else if traceID != "" && strings.TrimSpace(item.TraceID) == traceID {
			matched = true
		} else if channelID != "" && threadTS != "" && strings.TrimSpace(item.ChannelID) == channelID && strings.TrimSpace(item.ThreadTS) == threadTS {
			matched = true
		} else if conversationID != "" && strings.TrimSpace(item.ConversationID) == conversationID {
			matched = true
		}
		if matched {
			out = append(out, cloneDBReadRequest(item))
		}
	}
	SortDBReadRequests(out)
	return out
}

func (s *MemoryStore) GetDBReadRequest(requestID string) (DBReadRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.dbReadRequests[requestID]
	return cloneDBReadRequest(item), ok
}

func (s *MemoryStore) GetDBReadRequestByIdempotencyKey(key string) (DBReadRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.dbReadRequestByIdempotencyKey[key]
	if !ok {
		return DBReadRequest{}, false
	}
	item, ok := s.dbReadRequests[id]
	return cloneDBReadRequest(item), ok
}

func (s *MemoryStore) UpsertDBReadRequest(input DBReadCreateInput, now time.Time) (DBReadRequest, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.dbReadRequestByIdempotencyKey[input.IdempotencyKey]; ok {
		item, ok := s.dbReadRequests[id]
		if ok {
			return cloneDBReadRequest(item), false, nil
		}
	}
	if existing, ok := s.findDBReadRequestByExecutionScopeLocked(input.Target, input.ExecutionScopeKey); ok {
		return cloneDBReadRequest(existing), false, nil
	}
	item, err := NewDBReadRequest(input, now)
	if err != nil {
		return DBReadRequest{}, false, err
	}
	s.dbReadRequests[item.ID] = item
	s.dbReadRequestByIdempotencyKey[item.IdempotencyKey] = item.ID
	return cloneDBReadRequest(item), true, nil
}

func (s *MemoryStore) findDBReadRequestByExecutionScopeLocked(target string, scopeKey string) (DBReadRequest, bool) {
	target = strings.TrimSpace(target)
	scopeKey = strings.TrimSpace(scopeKey)
	if target == "" || scopeKey == "" {
		return DBReadRequest{}, false
	}
	var found DBReadRequest
	ok := false
	for _, item := range s.dbReadRequests {
		if item.Target != target || item.ExecutionScopeKey != scopeKey {
			continue
		}
		if !DBReadRequestBlocksNewScopedRequest(item.State) {
			continue
		}
		if !ok || item.CreatedAt.After(found.CreatedAt) {
			found = item
			ok = true
		}
	}
	return found, ok
}

func (s *MemoryStore) AppendDBReadValidationAttempt(attempt DBReadValidationAttempt) (DBReadValidationAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request, ok := s.dbReadRequests[attempt.RequestID]
	if !ok {
		return DBReadValidationAttempt{}, errors.New("db read request not found")
	}
	if attempt.ID == "" {
		attempt.ID = "dbreadval_" + uuid.NewString()
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = time.Now().UTC()
	}
	now := attempt.CreatedAt
	nextState := DBReadStateValidationFailed
	if attempt.Status == DBReadValidationStatusSucceeded {
		nextState = DBReadStatePendingApproval
	}
	if err := ValidateDBReadStateTransition(request.State, nextState); err != nil {
		return DBReadValidationAttempt{}, err
	}
	s.dbReadValidationAttempts[attempt.RequestID] = append(s.dbReadValidationAttempts[attempt.RequestID], cloneDBReadValidationAttempt(attempt))
	request.State = nextState
	request.CurrentValidationAttemptID = attempt.ID
	request.ErrorMessage = attempt.ErrorMessage
	request.LeaseHolder = ""
	request.LeaseToken = ""
	request.LeaseExpiresAt = nil
	request.UpdatedAt = now
	s.dbReadRequests[request.ID] = request
	return cloneDBReadValidationAttempt(attempt), nil
}

func (s *MemoryStore) ListDBReadValidationAttempts(requestID string) []DBReadValidationAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.dbReadValidationAttempts[requestID]
	out := make([]DBReadValidationAttempt, 0, len(items))
	for _, item := range items {
		out = append(out, cloneDBReadValidationAttempt(item))
	}
	return out
}

func (s *MemoryStore) TransitionDBReadRequest(requestID string, from DBReadState, to DBReadState, mutate func(*DBReadRequest) error) (DBReadRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.dbReadRequests[requestID]
	if !ok {
		return DBReadRequest{}, errors.New("db read request not found")
	}
	if item.State != from {
		return DBReadRequest{}, errors.New("db read request state mismatch")
	}
	if err := ValidateDBReadStateTransition(from, to); err != nil {
		return DBReadRequest{}, err
	}
	if mutate != nil {
		if err := mutate(&item); err != nil {
			return DBReadRequest{}, err
		}
	}
	item.State = to
	item.UpdatedAt = time.Now().UTC()
	s.dbReadRequests[item.ID] = item
	return cloneDBReadRequest(item), nil
}

func (s *MemoryStore) ClaimNextDBReadValidationRequest(holder string, lease time.Duration, now time.Time, targets []string) (DBReadLease, bool, error) {
	return s.claimNextDBReadRequestWithState(holder, lease, now, targets, DBReadStateValidating)
}

func (s *MemoryStore) ClaimNextDBReadRequest(holder string, lease time.Duration, now time.Time, targets []string) (DBReadLease, bool, error) {
	return s.claimNextDBReadRequestWithState(holder, lease, now, targets, DBReadStateApproved)
}

func (s *MemoryStore) claimNextDBReadRequestWithState(holder string, lease time.Duration, now time.Time, targets []string, state DBReadState) (DBReadLease, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	allowedTargets := dbReadTargetSet(targets)
	candidates := make([]DBReadRequest, 0, len(s.dbReadRequests))
	for _, item := range s.dbReadRequests {
		if item.State != state {
			continue
		}
		if len(allowedTargets) > 0 && !allowedTargets[item.Target] {
			continue
		}
		if item.LeaseExpiresAt != nil && item.LeaseExpiresAt.After(now) {
			continue
		}
		candidates = append(candidates, item)
	}
	if len(candidates) == 0 {
		return DBReadLease{}, false, nil
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		aApproved := a.ApprovedAt != nil
		bApproved := b.ApprovedAt != nil
		if aApproved != bApproved {
			return aApproved
		}
		if aApproved && bApproved {
			if !a.ApprovedAt.Equal(*b.ApprovedAt) {
				return a.ApprovedAt.Before(*b.ApprovedAt)
			}
		}
		return a.CreatedAt.Before(b.CreatedAt)
	})
	item := candidates[0]
	token := "dbreadlease_" + uuid.NewString()
	expires := now.Add(lease)
	item.LeaseHolder = holder
	item.LeaseToken = token
	item.LeaseGeneration++
	item.LeaseExpiresAt = &expires
	item.UpdatedAt = now
	s.dbReadRequests[item.ID] = item
	return DBReadLease{Request: cloneDBReadRequest(item), Token: token}, true, nil
}

func (s *MemoryStore) AppendDBReadExecutionResult(result DBReadExecutionResult) (DBReadExecutionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request, ok := s.dbReadRequests[result.RequestID]
	if !ok {
		return DBReadExecutionResult{}, errors.New("db read request not found")
	}
	if request.LeaseToken != "" && result.LeaseToken != "" && request.LeaseToken != result.LeaseToken {
		return DBReadExecutionResult{}, errors.New("db read lease token mismatch")
	}
	if result.ID == "" {
		result.ID = "dbreadexec_" + uuid.NewString()
	}
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now().UTC()
	}
	nextState := DBReadStateFailed
	if result.Status == DBReadExecutionStatusSucceeded {
		nextState = DBReadStateSucceeded
	}
	if request.State != DBReadStateExecuting && request.State != DBReadStateApproved {
		return DBReadExecutionResult{}, errors.New("db read request is not executable")
	}
	if request.State == DBReadStateApproved {
		if err := ValidateDBReadStateTransition(request.State, DBReadStateExecuting); err != nil {
			return DBReadExecutionResult{}, err
		}
		request.State = DBReadStateExecuting
	}
	if err := ValidateDBReadStateTransition(request.State, nextState); err != nil {
		return DBReadExecutionResult{}, err
	}
	s.dbReadExecutionResults[result.RequestID] = append(s.dbReadExecutionResults[result.RequestID], cloneDBReadExecutionResult(result))
	request.State = nextState
	request.RowCount = result.RowCount
	request.Truncated = result.Truncated
	request.ResultSample = cloneDBReadSample(result.Sample)
	request.ResultArtifactRef = result.ArtifactRef
	request.ErrorMessage = result.ErrorMessage
	request.LeaseHolder = ""
	request.LeaseToken = ""
	request.LeaseExpiresAt = nil
	request.UpdatedAt = result.CreatedAt
	s.dbReadRequests[request.ID] = request
	return cloneDBReadExecutionResult(result), nil
}

func (s *MemoryStore) ListDBReadExecutionResults(requestID string) []DBReadExecutionResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.dbReadExecutionResults[requestID]
	out := make([]DBReadExecutionResult, 0, len(items))
	for _, item := range items {
		out = append(out, cloneDBReadExecutionResult(item))
	}
	return out
}

func (s *MemoryStore) ExpirePendingDBReadRequests(now time.Time) ([]DBReadRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	out := []DBReadRequest{}
	for _, item := range s.dbReadRequests {
		if item.State != DBReadStateValidating &&
			item.State != DBReadStateValidationFailed &&
			item.State != DBReadStatePendingApproval &&
			item.State != DBReadStateApproved {
			continue
		}
		if item.ExpiresAt.After(now) {
			continue
		}
		if item.LeaseExpiresAt != nil && item.LeaseExpiresAt.After(now) {
			continue
		}
		if err := ValidateDBReadStateTransition(item.State, DBReadStateExpired); err != nil {
			return nil, err
		}
		item.State = DBReadStateExpired
		item.LeaseHolder = ""
		item.LeaseToken = ""
		item.LeaseExpiresAt = nil
		item.UpdatedAt = now
		s.dbReadRequests[item.ID] = item
		out = append(out, cloneDBReadRequest(item))
	}
	return out, nil
}

func dbReadTargetSet(targets []string) map[string]bool {
	out := map[string]bool{}
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target != "" {
			out[target] = true
		}
	}
	return out
}
