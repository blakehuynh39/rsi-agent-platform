package knowledge

import "strings"

func IsValidKind(kind Kind) bool {
	switch kind {
	case KindFact, KindPlaybook, KindArchitecturePattern, KindRepoNote, KindIncidentRunbook:
		return true
	default:
		return false
	}
}

func IsValidScopeType(scope ScopeType) bool {
	switch scope {
	case ScopeGlobal, ScopeRepo, ScopeService, ScopeConversation, ScopeCase:
		return true
	default:
		return false
	}
}

func IsDisplayableEntry(entry Entry) bool {
	if !IsValidKind(entry.Kind) || !IsValidScopeType(entry.ScopeType) {
		return false
	}
	if strings.TrimSpace(entry.Title) == "" {
		return false
	}
	if strings.TrimSpace(entry.Summary) == "" && strings.TrimSpace(entry.Body) == "" {
		return false
	}
	if entry.ScopeType != ScopeGlobal && strings.TrimSpace(entry.ScopeID) == "" {
		return false
	}
	return true
}
