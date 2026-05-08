package store

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DBReadState string

const (
	DBReadStateValidating       DBReadState = "validating"
	DBReadStateValidationFailed DBReadState = "validation_failed"
	DBReadStatePendingApproval  DBReadState = "pending_approval"
	DBReadStateApproved         DBReadState = "approved"
	DBReadStateDenied           DBReadState = "denied"
	DBReadStateExpired          DBReadState = "expired"
	DBReadStateExecuting        DBReadState = "executing"
	DBReadStateSucceeded        DBReadState = "succeeded"
	DBReadStateFailed           DBReadState = "failed"
)

type DBReadValidationStatus string

const (
	DBReadValidationStatusSucceeded DBReadValidationStatus = "succeeded"
	DBReadValidationStatusFailed    DBReadValidationStatus = "failed"
)

type DBReadExecutionStatus string

const (
	DBReadExecutionStatusSucceeded DBReadExecutionStatus = "succeeded"
	DBReadExecutionStatusFailed    DBReadExecutionStatus = "failed"
)

type DBReadCaps struct {
	MaxRows        int `json:"max_rows,omitempty"`
	MaxBytes       int `json:"max_bytes,omitempty"`
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
	LockTimeoutMS  int `json:"lock_timeout_ms,omitempty"`
}

type DBReadRedactionPolicy struct {
	DenyColumns []string `json:"deny_columns,omitempty"`
}

type DBReadRequest struct {
	ID                         string                 `json:"id"`
	IdempotencyKey             string                 `json:"idempotency_key"`
	Target                     string                 `json:"target"`
	Purpose                    string                 `json:"purpose"`
	SQL                        string                 `json:"sql"`
	SQLSHA256                  string                 `json:"sql_sha256"`
	ExecutionScopeKey          string                 `json:"execution_scope_key,omitempty"`
	Requester                  string                 `json:"requester"`
	ConversationID             string                 `json:"conversation_id,omitempty"`
	WorkflowID                 string                 `json:"workflow_id,omitempty"`
	TraceID                    string                 `json:"trace_id,omitempty"`
	ChannelID                  string                 `json:"channel_id,omitempty"`
	ThreadTS                   string                 `json:"thread_ts,omitempty"`
	State                      DBReadState            `json:"state"`
	CurrentValidationAttemptID string                 `json:"current_validation_attempt_id,omitempty"`
	ApprovedBySlackUserID      string                 `json:"approved_by_slack_user_id,omitempty"`
	ApprovedAt                 *time.Time             `json:"approved_at,omitempty"`
	ExpiresAt                  time.Time              `json:"expires_at"`
	Caps                       DBReadCaps             `json:"caps"`
	Redaction                  DBReadRedactionPolicy  `json:"redaction"`
	SlackMessageChannelID      string                 `json:"slack_message_channel_id,omitempty"`
	SlackMessageTS             string                 `json:"slack_message_ts,omitempty"`
	LeaseHolder                string                 `json:"lease_holder,omitempty"`
	LeaseToken                 string                 `json:"lease_token,omitempty"`
	LeaseGeneration            int                    `json:"lease_generation,omitempty"`
	LeaseExpiresAt             *time.Time             `json:"lease_expires_at,omitempty"`
	ResultArtifactRef          string                 `json:"result_artifact_ref,omitempty"`
	ResultSample               []map[string]string    `json:"result_sample,omitempty"`
	RowCount                   int                    `json:"row_count,omitempty"`
	Truncated                  bool                   `json:"truncated,omitempty"`
	ErrorMessage               string                 `json:"error_message,omitempty"`
	Metadata                   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt                  time.Time              `json:"created_at"`
	UpdatedAt                  time.Time              `json:"updated_at"`
}

type DBReadValidationAttempt struct {
	ID           string                 `json:"id"`
	RequestID    string                 `json:"request_id"`
	Target       string                 `json:"target"`
	SQLSHA256    string                 `json:"sql_sha256"`
	Status       DBReadValidationStatus `json:"status"`
	Stage        string                 `json:"stage"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	ValidatedAt  *time.Time             `json:"validated_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

type DBReadExecutionResult struct {
	ID           string                `json:"id"`
	RequestID    string                `json:"request_id"`
	LeaseToken   string                `json:"lease_token,omitempty"`
	Status       DBReadExecutionStatus `json:"status"`
	RowCount     int                   `json:"row_count,omitempty"`
	Truncated    bool                  `json:"truncated,omitempty"`
	Sample       []map[string]string   `json:"sample,omitempty"`
	ArtifactRef  string                `json:"artifact_ref,omitempty"`
	ErrorCode    string                `json:"error_code,omitempty"`
	ErrorMessage string                `json:"error_message,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
}

type DBReadCreateInput struct {
	IdempotencyKey    string
	Target            string
	Purpose           string
	SQL               string
	SQLSHA256         string
	ExecutionScopeKey string
	Requester         string
	ConversationID    string
	WorkflowID        string
	TraceID           string
	ChannelID         string
	ThreadTS          string
	ExpiresAt         time.Time
	Caps              DBReadCaps
	Redaction         DBReadRedactionPolicy
	Metadata          map[string]interface{}
}

type DBReadLease struct {
	Request DBReadRequest `json:"request"`
	Token   string        `json:"token"`
}

func NewDBReadRequest(input DBReadCreateInput, now time.Time) (DBReadRequest, error) {
	input.Target = strings.TrimSpace(input.Target)
	input.SQL = strings.TrimSpace(input.SQL)
	input.SQLSHA256 = strings.TrimSpace(input.SQLSHA256)
	input.ExecutionScopeKey = strings.TrimSpace(input.ExecutionScopeKey)
	input.Requester = strings.TrimSpace(input.Requester)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.IdempotencyKey == "" {
		return DBReadRequest{}, errors.New("db read idempotency key is required")
	}
	if input.Target == "" {
		return DBReadRequest{}, errors.New("db read target is required")
	}
	if input.SQL == "" || input.SQLSHA256 == "" {
		return DBReadRequest{}, errors.New("db read sql and sql hash are required")
	}
	if input.Requester == "" {
		return DBReadRequest{}, errors.New("db read requester is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if input.ExpiresAt.IsZero() {
		input.ExpiresAt = now.Add(time.Hour)
	}
	return DBReadRequest{
		ID:                "dbread_" + uuid.NewString(),
		IdempotencyKey:    input.IdempotencyKey,
		Target:            input.Target,
		Purpose:           firstNonEmpty(input.Purpose, "query"),
		SQL:               input.SQL,
		SQLSHA256:         input.SQLSHA256,
		ExecutionScopeKey: input.ExecutionScopeKey,
		Requester:         input.Requester,
		ConversationID:    input.ConversationID,
		WorkflowID:        input.WorkflowID,
		TraceID:           input.TraceID,
		ChannelID:         input.ChannelID,
		ThreadTS:          input.ThreadTS,
		State:             DBReadStateValidating,
		ExpiresAt:         input.ExpiresAt,
		Caps:              input.Caps,
		Redaction:         input.Redaction,
		Metadata:          cloneAnyMap(input.Metadata),
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func NewDBReadValidationAttempt(request DBReadRequest, status DBReadValidationStatus, stage string, message string, details map[string]interface{}, now time.Time) DBReadValidationAttempt {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	var validatedAt *time.Time
	if status == DBReadValidationStatusSucceeded {
		t := now
		validatedAt = &t
	}
	errorCode, _ := details["error_code"].(string)
	return DBReadValidationAttempt{
		ID:           "dbreadval_" + uuid.NewString(),
		RequestID:    request.ID,
		Target:       request.Target,
		SQLSHA256:    request.SQLSHA256,
		Status:       status,
		Stage:        strings.TrimSpace(stage),
		ErrorCode:    strings.TrimSpace(errorCode),
		ErrorMessage: strings.TrimSpace(message),
		Details:      cloneAnyMap(details),
		ValidatedAt:  validatedAt,
		CreatedAt:    now,
	}
}

func NewDBReadExecutionResult(request DBReadRequest, status DBReadExecutionStatus, sample []map[string]string, now time.Time) DBReadExecutionResult {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return DBReadExecutionResult{
		ID:         "dbreadexec_" + uuid.NewString(),
		RequestID:  request.ID,
		LeaseToken: request.LeaseToken,
		Status:     status,
		Sample:     cloneDBReadSample(sample),
		CreatedAt:  now,
	}
}

func DBReadStateTransitionAllowed(from DBReadState, to DBReadState) bool {
	if from == to {
		return true
	}
	allowed := map[DBReadState][]DBReadState{
		DBReadStateValidating:       {DBReadStateValidationFailed, DBReadStatePendingApproval, DBReadStateExpired},
		DBReadStateValidationFailed: {DBReadStateValidating, DBReadStateExpired},
		DBReadStatePendingApproval:  {DBReadStateApproved, DBReadStateDenied, DBReadStateExpired},
		DBReadStateApproved:         {DBReadStateExecuting, DBReadStateExpired},
		DBReadStateExecuting:        {DBReadStateSucceeded, DBReadStateFailed},
	}
	for _, candidate := range allowed[from] {
		if candidate == to {
			return true
		}
	}
	return false
}

func ValidateDBReadStateTransition(from DBReadState, to DBReadState) error {
	if !DBReadStateTransitionAllowed(from, to) {
		return fmt.Errorf("illegal db read state transition %q -> %q", from, to)
	}
	return nil
}

func SortDBReadRequests(items []DBReadRequest) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func cloneDBReadRequest(item DBReadRequest) DBReadRequest {
	item.Metadata = cloneAnyMap(item.Metadata)
	item.ResultSample = cloneDBReadSample(item.ResultSample)
	item.Caps = DBReadCaps{
		MaxRows:        item.Caps.MaxRows,
		MaxBytes:       item.Caps.MaxBytes,
		TimeoutSeconds: item.Caps.TimeoutSeconds,
		LockTimeoutMS:  item.Caps.LockTimeoutMS,
	}
	item.Redaction = DBReadRedactionPolicy{DenyColumns: append([]string(nil), item.Redaction.DenyColumns...)}
	return item
}

func cloneDBReadValidationAttempt(item DBReadValidationAttempt) DBReadValidationAttempt {
	item.Details = cloneAnyMap(item.Details)
	return item
}

func cloneDBReadExecutionResult(item DBReadExecutionResult) DBReadExecutionResult {
	item.Sample = cloneDBReadSample(item.Sample)
	return item
}

func cloneDBReadSample(in []map[string]string) []map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make([]map[string]string, 0, len(in))
	for _, row := range in {
		next := make(map[string]string, len(row))
		for key, value := range row {
			next[key] = value
		}
		out = append(out, next)
	}
	return out
}
