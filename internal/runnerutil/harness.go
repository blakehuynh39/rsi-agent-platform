package runnerutil

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func PersistHarnessExecution(
	store storepkg.Repository,
	resp clients.RunnerResponse,
	role string,
	operationID string,
	traceID string,
	proposalID string,
	fallbackProfileID string,
	fallbackOverlayVersion string,
	requestedScopeKind string,
	requestedScopeID string,
	requestedParentScopeKind string,
	requestedParentScopeID string,
) error {
	meta := harness.DecodeExecutionMetadata(resp.Raw)
	if strings.TrimSpace(meta.HermesSessionID) == "" {
		return nil
	}
	now := time.Now().UTC()
	profileID := firstNonEmptyHarness(meta.HarnessProfileID, fallbackProfileID)
	binding := harness.SessionBinding{
		Role:                    role,
		ScopeKind:               firstNonEmptyHarness(meta.SessionScopeKind, requestedScopeKind),
		ScopeID:                 firstNonEmptyHarness(meta.SessionScopeID, requestedScopeID),
		ParentScopeKind:         firstNonEmptyHarness(meta.ParentScopeKind, requestedParentScopeKind),
		ParentScopeID:           firstNonEmptyHarness(meta.ParentScopeID, requestedParentScopeID),
		HermesSessionID:         meta.HermesSessionID,
		ParentSessionID:         meta.ParentSessionID,
		MemoryBackend:           meta.MemoryBackend,
		AssistantPeerID:         meta.AssistantPeerID,
		UserPeerID:              meta.UserPeerID,
		HarnessProfileID:        profileID,
		EffectiveOverlayID:      meta.EffectiveOverlayID,
		EffectiveOverlayVersion: firstNonEmptyHarness(meta.EffectiveOverlayVersion, fallbackOverlayVersion),
		LastUsedAt:              now,
		UpdatedAt:               now,
	}
	if err := submitHarnessSessionBinding(store, binding, role, operationID, now); err != nil {
		return err
	}
	return submitHarnessExecution(store, harness.Execution{
		OperationID:             operationID,
		TraceID:                 traceID,
		ProposalID:              proposalID,
		Role:                    role,
		SessionScopeKind:        binding.ScopeKind,
		SessionScopeID:          binding.ScopeID,
		HermesSessionID:         binding.HermesSessionID,
		ParentSessionID:         binding.ParentSessionID,
		HarnessProfileID:        profileID,
		EffectiveOverlayID:      meta.EffectiveOverlayID,
		EffectiveOverlayVersion: binding.EffectiveOverlayVersion,
		MemoryBackend:           meta.MemoryBackend,
		MemoryReads:             meta.MemoryReads,
		MemoryWrites:            meta.MemoryWrites,
		CreatedAt:               now,
	}, role, operationID, now)
}

func firstNonEmptyHarness(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func submitHarnessSessionBinding(store storepkg.Repository, binding harness.SessionBinding, actor string, operationID string, occurredAt time.Time) error {
	commandID := fmt.Sprintf("cmd-harness:bind:%s:%s", harnessBindingAggregateID(binding), firstNonEmptyHarness(operationID, binding.HermesSessionID))
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineHarness,
		AggregateID: harnessBindingAggregateID(binding),
		CommandKind: string(transition.CommandHarnessBindSession),
		CommandID:   commandID,
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload: map[string]any{
			"role":                      binding.Role,
			"scope_kind":                binding.ScopeKind,
			"scope_id":                  binding.ScopeID,
			"parent_scope_kind":         binding.ParentScopeKind,
			"parent_scope_id":           binding.ParentScopeID,
			"hermes_session_id":         binding.HermesSessionID,
			"parent_session_id":         binding.ParentSessionID,
			"memory_backend":            binding.MemoryBackend,
			"assistant_peer_id":         binding.AssistantPeerID,
			"user_peer_id":              binding.UserPeerID,
			"harness_profile_id":        binding.HarnessProfileID,
			"effective_overlay_id":      binding.EffectiveOverlayID,
			"effective_overlay_version": binding.EffectiveOverlayVersion,
			"last_used_at":              binding.LastUsedAt,
			"created_at":                binding.CreatedAt,
			"updated_at":                binding.UpdatedAt,
		},
	})
	if err != nil {
		return err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return errors.New(receipt.Reason)
	}
	return nil
}

func submitHarnessExecution(store storepkg.Repository, execution harness.Execution, actor string, operationID string, occurredAt time.Time) error {
	executionID := harnessExecutionAggregateID(execution, occurredAt)
	commandID := fmt.Sprintf("cmd-harness:execution:%s", executionID)
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineHarness,
		AggregateID: executionID,
		CommandKind: string(transition.CommandHarnessRecordExecution),
		CommandID:   commandID,
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload: map[string]any{
			"operation_id":              firstNonEmptyHarness(execution.OperationID, operationID),
			"trace_id":                  execution.TraceID,
			"proposal_id":               execution.ProposalID,
			"role":                      execution.Role,
			"session_scope_kind":        execution.SessionScopeKind,
			"session_scope_id":          execution.SessionScopeID,
			"hermes_session_id":         execution.HermesSessionID,
			"parent_session_id":         execution.ParentSessionID,
			"harness_profile_id":        execution.HarnessProfileID,
			"effective_overlay_id":      execution.EffectiveOverlayID,
			"effective_overlay_version": execution.EffectiveOverlayVersion,
			"memory_backend":            execution.MemoryBackend,
			"memory_reads":              execution.MemoryReads,
			"memory_writes":             execution.MemoryWrites,
			"created_at":                execution.CreatedAt,
		},
	})
	if err != nil {
		return err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return errors.New(receipt.Reason)
	}
	return nil
}

func harnessBindingAggregateID(binding harness.SessionBinding) string {
	return strings.TrimSpace(binding.Role) + "|" + strings.TrimSpace(binding.ScopeKind) + "|" + strings.TrimSpace(binding.ScopeID)
}

func harnessExecutionAggregateID(execution harness.Execution, occurredAt time.Time) string {
	if trimmed := strings.TrimSpace(execution.ID); trimmed != "" {
		return trimmed
	}
	seed := firstNonEmptyHarness(
		execution.OperationID,
		fmt.Sprintf("%s|%s|%s|%d", execution.Role, execution.SessionScopeKind, execution.SessionScopeID, occurredAt.UnixNano()),
	)
	sum := sha1.Sum([]byte(seed))
	return fmt.Sprintf("hexec-%x", sum[:8])
}
