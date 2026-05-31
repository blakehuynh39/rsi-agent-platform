package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type CommandSubmitter interface {
	SubmitCommand(command transition.CommandEnvelope) (transition.CommandReceipt, error)
}

type CommandRequest struct {
	CommandKind     string         `json:"command_kind"`
	CommandID       string         `json:"command_id,omitempty"`
	CausationID     string         `json:"causation_id,omitempty"`
	Actor           string         `json:"actor,omitempty"`
	OccurredAt      time.Time      `json:"occurred_at,omitempty"`
	ExpectedVersion int64          `json:"expected_version,omitempty"`
	Payload         map[string]any `json:"payload,omitempty"`
}

func SubmitMachineCommand(w http.ResponseWriter, r *http.Request, store CommandSubmitter, machine transition.MachineKind, aggregateID string, defaultActor string) (transition.CommandReceipt, bool) {
	command, err := decodeCommandRequest(r, machine, aggregateID, defaultActor)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return transition.CommandReceipt{}, false
	}
	receipt, err := store.SubmitCommand(command)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return transition.CommandReceipt{}, false
	}
	if receipt.DecisionKind == transition.DecisionReject {
		WriteError(w, http.StatusConflict, errors.New(receipt.Reason))
		return transition.CommandReceipt{}, false
	}
	return receipt, true
}

func decodeCommandRequest(r *http.Request, machine transition.MachineKind, aggregateID string, defaultActor string) (transition.CommandEnvelope, error) {
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return transition.CommandEnvelope{}, err
	}
	commandKind := strings.TrimSpace(req.CommandKind)
	if commandKind == "" {
		return transition.CommandEnvelope{}, errors.New("command_kind is required")
	}
	commandID := strings.TrimSpace(req.CommandID)
	if commandID == "" {
		commandID = "cmd-" + uuid.NewString()
	}
	occurredAt := req.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	actor := strings.TrimSpace(req.Actor)
	if actor == "" {
		actor = strings.TrimSpace(defaultActor)
	}
	return transition.CommandEnvelope{
		MachineKind:     machine,
		AggregateID:     strings.TrimSpace(aggregateID),
		CommandKind:     commandKind,
		CommandID:       commandID,
		CausationID:     strings.TrimSpace(req.CausationID),
		Actor:           actor,
		OccurredAt:      occurredAt,
		ExpectedVersion: req.ExpectedVersion,
		Payload:         req.Payload,
	}, nil
}
