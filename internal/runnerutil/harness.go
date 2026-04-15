package runnerutil

import (
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
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
	if _, err := store.UpsertHarnessSessionBinding(binding); err != nil {
		return err
	}
	_, err := store.RecordHarnessExecution(harness.Execution{
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
	})
	return err
}

func firstNonEmptyHarness(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
