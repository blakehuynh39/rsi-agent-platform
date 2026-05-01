package improvementplane

import (
	"sort"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type actionDetailResponse struct {
	ActionIntent  action.Intent   `json:"action_intent"`
	ActionResults []action.Result `json:"action_results"`
}

type knowledgeDetailResponse struct {
	KnowledgeEntry knowledge.Entry          `json:"knowledge_entry"`
	EvidenceLinks  []knowledge.EvidenceLink `json:"evidence_links"`
	Reviews        []knowledge.Review       `json:"reviews"`
}

type actionFilters struct {
	ConversationID string
	CaseID         string
	TraceID        string
	ProposalID     string
}

type knowledgeFilters struct {
	Tier      string
	Status    string
	ScopeType string
	ScopeID   string
}

func listActionIntents(store storepkg.Repository, filters actionFilters) []action.Intent {
	items := storeActionIntents(store, filters)
	out := make([]action.Intent, 0)
	for _, item := range items {
		if filters.ConversationID != "" && item.ConversationID != filters.ConversationID {
			continue
		}
		if filters.CaseID != "" && item.CaseID != filters.CaseID {
			continue
		}
		if filters.TraceID != "" && item.TraceID != filters.TraceID {
			continue
		}
		if filters.ProposalID != "" && item.ProposalID != filters.ProposalID {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func buildActionDetail(store storepkg.Repository, actionID string) (actionDetailResponse, bool) {
	intent, ok := store.GetActionIntent(actionID)
	if !ok {
		return actionDetailResponse{}, false
	}
	return actionDetailResponse{
		ActionIntent:  intent,
		ActionResults: sliceOrEmpty(store.ListActionResults(actionID)),
	}, true
}

func flattenActionResults(store storepkg.Repository, intents []action.Intent) []action.Result {
	out := make([]action.Result, 0)
	for _, item := range intents {
		out = append(out, sliceOrEmpty(store.ListActionResults(item.ID))...)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CompletedAt.After(out[j].CompletedAt)
	})
	return out
}

func listOutcomes(store storepkg.Repository, conversationID string, caseID string, traceID string, proposalID string) []outcome.Record {
	items := storeOutcomes(store, conversationID, caseID, traceID, proposalID)
	out := make([]outcome.Record, 0)
	for _, item := range items {
		if conversationID != "" && item.ConversationID != conversationID {
			continue
		}
		if caseID != "" && item.CaseID != caseID {
			continue
		}
		if traceID != "" && item.TraceID != traceID {
			continue
		}
		if proposalID != "" && item.ProposalID != proposalID {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].RecordedAt.After(out[j].RecordedAt)
	})
	return out
}

func listKnowledgeEntries(store storepkg.Repository, filters knowledgeFilters) []knowledge.Entry {
	items := store.ListKnowledgeEntries()
	out := make([]knowledge.Entry, 0)
	for _, item := range items {
		if filters.Tier != "" && string(item.Tier) != filters.Tier {
			continue
		}
		if filters.Status != "" && string(item.Status) != filters.Status {
			continue
		}
		if filters.ScopeType != "" && string(item.ScopeType) != filters.ScopeType {
			continue
		}
		if filters.ScopeID != "" && item.ScopeID != filters.ScopeID {
			continue
		}
		if !knowledge.IsDisplayableEntry(item) {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func buildKnowledgeDetail(store storepkg.Repository, knowledgeID string) (knowledgeDetailResponse, bool) {
	item, ok := store.GetKnowledgeEntry(knowledgeID)
	if !ok {
		return knowledgeDetailResponse{}, false
	}
	return knowledgeDetailResponse{
		KnowledgeEntry: item,
		EvidenceLinks:  sliceOrEmpty(store.ListKnowledgeEvidenceLinks(knowledgeID)),
		Reviews:        sliceOrEmpty(store.ListKnowledgeReviews(knowledgeID)),
	}, true
}

func relatedKnowledgeEntries(store storepkg.Repository, conversationID string, caseID string, traceID string, proposalID string, extraEvidenceIDs ...string) []knowledge.Entry {
	evidenceIDs := map[string]struct{}{}
	for _, id := range extraEvidenceIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			evidenceIDs[id] = struct{}{}
		}
	}
	if traceID != "" {
		evidenceIDs[traceID] = struct{}{}
	}
	if proposalID != "" {
		evidenceIDs[proposalID] = struct{}{}
	}
	out := make([]knowledge.Entry, 0)
	seen := map[string]struct{}{}
	for _, item := range store.ListKnowledgeEntries() {
		if !knowledge.IsDisplayableEntry(item) {
			continue
		}
		if conversationID != "" && item.ScopeType == knowledge.ScopeConversation && item.ScopeID == conversationID {
			if _, ok := seen[item.ID]; !ok {
				seen[item.ID] = struct{}{}
				out = append(out, item)
			}
			continue
		}
		if caseID != "" && item.ScopeType == knowledge.ScopeCase && item.ScopeID == caseID {
			if _, ok := seen[item.ID]; !ok {
				seen[item.ID] = struct{}{}
				out = append(out, item)
			}
			continue
		}
		links := store.ListKnowledgeEvidenceLinks(item.ID)
		for _, link := range links {
			if _, ok := evidenceIDs[link.EvidenceID]; !ok {
				if _, ok := evidenceIDs[link.EvidenceRef.Ref]; !ok {
					continue
				}
			}
			if _, ok := seen[item.ID]; ok {
				break
			}
			seen[item.ID] = struct{}{}
			out = append(out, item)
			break
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}
