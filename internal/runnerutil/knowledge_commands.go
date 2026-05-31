package runnerutil

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func PersistKnowledgeDraft(store storepkg.Repository, entry knowledge.Entry, links []knowledge.EvidenceLink, actor string, correlationKey string, ordinal int, occurredAt time.Time) (knowledge.Entry, error) {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = occurredAt
	}
	if entry.UpdatedAt.IsZero() {
		entry.UpdatedAt = occurredAt
	}
	if strings.TrimSpace(entry.ID) == "" {
		entry.ID = knowledgeDraftAggregateID(entry, correlationKey, ordinal, occurredAt)
	}
	commandID := fmt.Sprintf("cmd-knowledge:%s:record_draft", entry.ID)
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineKnowledge,
		AggregateID: entry.ID,
		CommandKind: string(transition.CommandKnowledgeRecordDraft),
		CommandID:   commandID,
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload: map[string]any{
			"tier":                     string(entry.Tier),
			"kind":                     string(entry.Kind),
			"scope_type":               string(entry.ScopeType),
			"scope_id":                 entry.ScopeID,
			"title":                    entry.Title,
			"summary":                  entry.Summary,
			"body":                     entry.Body,
			"structured_facts":         entry.StructuredFacts,
			"status":                   string(entry.Status),
			"confidence":               entry.Confidence,
			"fresh_until":              entry.FreshUntil,
			"source_type":              string(entry.SourceType),
			"supersedes_entry_id":      entry.SupersedesEntryID,
			"contradicted_by_entry_id": entry.ContradictedByEntryID,
			"created_at":               entry.CreatedAt,
			"updated_at":               entry.UpdatedAt,
			"evidence_links":           links,
		},
	})
	if err != nil {
		return knowledge.Entry{}, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return knowledge.Entry{}, errors.New(receipt.Reason)
	}
	stored, ok := store.GetKnowledgeEntry(entry.ID)
	if !ok {
		return knowledge.Entry{}, fmt.Errorf("knowledge entry %s not found after draft command", entry.ID)
	}
	return stored, nil
}

func knowledgeDraftAggregateID(entry knowledge.Entry, correlationKey string, ordinal int, occurredAt time.Time) string {
	seed := strings.Join([]string{
		strings.TrimSpace(correlationKey),
		string(entry.Kind),
		string(entry.ScopeType),
		strings.TrimSpace(entry.ScopeID),
		fmt.Sprintf("%d", ordinal),
	}, "|")
	sum := sha1.Sum([]byte(seed))
	return fmt.Sprintf("knowledge-%x", sum[:8])
}
