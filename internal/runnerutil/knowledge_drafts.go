package runnerutil

import (
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
)

func NormalizeKnowledgeDraft(item KnowledgeDraft, defaultScopeType knowledge.ScopeType, defaultScopeID string) (KnowledgeDraft, bool) {
	title := strings.TrimSpace(item.Title)
	summary := strings.TrimSpace(item.Summary)
	body := strings.TrimSpace(item.Body)
	if title == "" || (summary == "" && body == "") {
		return KnowledgeDraft{}, false
	}

	refs := normalizeKnowledgeEvidenceRefs(item.EvidenceRefs)
	if len(refs) == 0 {
		return KnowledgeDraft{}, false
	}

	kind := knowledge.Kind(strings.TrimSpace(item.Kind))
	if kind == "" {
		kind = knowledge.KindFact
	}
	if !knowledge.IsValidKind(kind) {
		return KnowledgeDraft{}, false
	}

	scopeType := knowledge.ScopeType(strings.TrimSpace(item.ScopeType))
	if scopeType == "" {
		scopeType = defaultScopeType
	}
	if scopeType == "" {
		scopeType = knowledge.ScopeCase
	}
	if !knowledge.IsValidScopeType(scopeType) {
		return KnowledgeDraft{}, false
	}

	scopeID := strings.TrimSpace(item.ScopeID)
	if scopeType == knowledge.ScopeGlobal {
		scopeID = ""
	} else if scopeID == "" {
		scopeID = strings.TrimSpace(defaultScopeID)
	}
	if scopeType != knowledge.ScopeGlobal && scopeID == "" {
		return KnowledgeDraft{}, false
	}

	return KnowledgeDraft{
		Kind:         string(kind),
		ScopeType:    string(scopeType),
		ScopeID:      scopeID,
		Title:        title,
		Summary:      summary,
		Body:         body,
		Confidence:   item.Confidence,
		FreshUntil:   strings.TrimSpace(item.FreshUntil),
		EvidenceRefs: refs,
	}, true
}

func normalizeKnowledgeEvidenceRefs(items []events.EvidenceRef) []events.EvidenceRef {
	out := make([]events.EvidenceRef, 0, len(items))
	for _, item := range items {
		kind := strings.TrimSpace(item.Kind)
		ref := strings.TrimSpace(item.Ref)
		if kind == "" || ref == "" {
			continue
		}
		out = append(out, events.EvidenceRef{
			Kind:    kind,
			Ref:     ref,
			Summary: strings.TrimSpace(item.Summary),
		})
	}
	return out
}
