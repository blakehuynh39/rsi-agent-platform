package store

import (
	"errors"
	"sort"
	"strings"
	"time"

	deep "github.com/brunoga/deep/v5"
	"github.com/google/uuid"
)

type ExternalToolActionState string

const (
	ExternalToolActionStateRequested ExternalToolActionState = "requested"
	ExternalToolActionStateSucceeded ExternalToolActionState = "succeeded"
	ExternalToolActionStateFailed    ExternalToolActionState = "failed"
)

type ExternalToolActionUpsertStatus string

const (
	ExternalToolActionUpsertCreated  ExternalToolActionUpsertStatus = "created"
	ExternalToolActionUpsertReplay   ExternalToolActionUpsertStatus = "replay"
	ExternalToolActionUpsertConflict ExternalToolActionUpsertStatus = "conflict"
)

type ExternalToolAction struct {
	ID              string                  `json:"id"`
	Surface         string                  `json:"surface"`
	Operation       string                  `json:"operation"`
	TargetRef       string                  `json:"target_ref,omitempty"`
	IdempotencyKey  string                  `json:"idempotency_key"`
	RequestHash     string                  `json:"request_hash"`
	State           ExternalToolActionState `json:"state"`
	Actor           string                  `json:"actor"`
	Reason          string                  `json:"reason,omitempty"`
	Destructive     bool                    `json:"destructive"`
	ExecutionID     string                  `json:"execution_id,omitempty"`
	OperationID     string                  `json:"operation_id,omitempty"`
	TraceID         string                  `json:"trace_id,omitempty"`
	WorkflowID      string                  `json:"workflow_id,omitempty"`
	ConversationID  string                  `json:"conversation_id,omitempty"`
	ResponseSummary string                  `json:"response_summary,omitempty"`
	ErrorMessage    string                  `json:"error,omitempty"`
	SourceRef       string                  `json:"source_ref,omitempty"`
	WikiAuditID     string                  `json:"wiki_audit_id,omitempty"`
	ResultPayload   map[string]any          `json:"result_payload,omitempty"`
	MirrorEffect    map[string]any          `json:"mirror_effect,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	CompletedAt     *time.Time              `json:"completed_at,omitempty"`
}

type ExternalToolActionCreateInput struct {
	Surface        string
	Operation      string
	TargetRef      string
	IdempotencyKey string
	RequestHash    string
	Actor          string
	Reason         string
	Destructive    bool
	ExecutionID    string
	OperationID    string
	TraceID        string
	WorkflowID     string
	ConversationID string
}

type ExternalToolActionResultUpdate struct {
	State           ExternalToolActionState
	ResponseSummary string
	ErrorMessage    string
	SourceRef       string
	WikiAuditID     string
	ResultPayload   map[string]any
	MirrorEffect    map[string]any
}

type ExternalToolActionStore interface {
	ListExternalToolActions() []ExternalToolAction
	GetExternalToolAction(actionID string) (ExternalToolAction, bool)
	GetExternalToolActionByIdempotency(surface string, operation string, idempotencyKey string) (ExternalToolAction, bool)
	UpsertExternalToolAction(input ExternalToolActionCreateInput, now time.Time) (ExternalToolAction, ExternalToolActionUpsertStatus, error)
	UpdateExternalToolActionResult(actionID string, update ExternalToolActionResultUpdate, now time.Time) (ExternalToolAction, error)
}

func NewExternalToolAction(input ExternalToolActionCreateInput, now time.Time) (ExternalToolAction, error) {
	input.Surface = strings.TrimSpace(input.Surface)
	input.Operation = strings.TrimSpace(input.Operation)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.RequestHash = strings.TrimSpace(input.RequestHash)
	input.Actor = strings.TrimSpace(input.Actor)
	if input.Surface == "" {
		return ExternalToolAction{}, errors.New("external tool action surface is required")
	}
	if input.Operation == "" {
		return ExternalToolAction{}, errors.New("external tool action operation is required")
	}
	if input.IdempotencyKey == "" {
		return ExternalToolAction{}, errors.New("external tool action idempotency key is required")
	}
	if input.RequestHash == "" {
		return ExternalToolAction{}, errors.New("external tool action request hash is required")
	}
	if input.Actor == "" {
		return ExternalToolAction{}, errors.New("external tool action actor is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return ExternalToolAction{
		ID:             "extact_" + uuid.NewString(),
		Surface:        input.Surface,
		Operation:      input.Operation,
		TargetRef:      strings.TrimSpace(input.TargetRef),
		IdempotencyKey: input.IdempotencyKey,
		RequestHash:    input.RequestHash,
		State:          ExternalToolActionStateRequested,
		Actor:          input.Actor,
		Reason:         strings.TrimSpace(input.Reason),
		Destructive:    input.Destructive,
		ExecutionID:    strings.TrimSpace(input.ExecutionID),
		OperationID:    strings.TrimSpace(input.OperationID),
		TraceID:        strings.TrimSpace(input.TraceID),
		WorkflowID:     strings.TrimSpace(input.WorkflowID),
		ConversationID: strings.TrimSpace(input.ConversationID),
		ResultPayload:  map[string]any{},
		MirrorEffect:   map[string]any{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func ExternalToolActionIdempotencyKey(surface string, operation string, idempotencyKey string) string {
	return strings.Join([]string{
		strings.TrimSpace(surface),
		strings.TrimSpace(operation),
		strings.TrimSpace(idempotencyKey),
	}, "\x00")
}

func SortExternalToolActions(items []ExternalToolAction) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}
		return left.ID < right.ID
	})
}

func cloneExternalToolAction(item ExternalToolAction) ExternalToolAction {
	item.ResultPayload = CloneJSONMap(item.ResultPayload)
	item.MirrorEffect = CloneJSONMap(item.MirrorEffect)
	if item.CompletedAt != nil {
		completed := *item.CompletedAt
		item.CompletedAt = &completed
	}
	return item
}

func CloneJSONMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	return deep.Clone(input)
}
