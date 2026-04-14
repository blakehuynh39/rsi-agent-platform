package harness

import (
	"encoding/json"
	"sort"
	"strings"
)

type EffectiveConfig struct {
	Profile                 Profile
	Overlay                 *Overlay
	PromptFragments         []string
	FewShotSnippets         []string
	ToolPreferenceOrder     []string
	RetrievalBias           string
	ReasoningVerbosity      string
	MemoryReadEnabled       bool
	MemoryWriteEnabled      bool
	EffectiveOverlayID      string
	EffectiveOverlayVersion string
}

type ExecutionMetadata struct {
	HermesSessionID         string
	ParentSessionID         string
	HarnessProfileID        string
	EffectiveOverlayID      string
	EffectiveOverlayVersion string
	MemoryBackend           string
	AssistantPeerID         string
	UserPeerID              string
	SessionScopeKind        string
	SessionScopeID          string
	ParentScopeKind         string
	ParentScopeID           string
	MemoryReads             []MemoryArtifact
	MemoryWrites            []MemoryArtifact
}

type ConfigResolver interface {
	GetHarnessProfile(profileID string) (Profile, bool)
	GetActiveHarnessOverlay(role string) (Overlay, bool)
}

func ResolveEffectiveConfig(store ConfigResolver, role string, fallbackReasoningVerbosity string) EffectiveConfig {
	profile, ok := store.GetHarnessProfile(DefaultProfileID(role))
	if !ok {
		profile = Profile{
			ID:                 DefaultProfileID(role),
			Role:               role,
			Name:               role,
			Model:              "openai/gpt-5.4",
			ReasoningEffort:    "xhigh",
			ReasoningVerbosity: fallbackReasoningVerbosity,
			MemoryReadEnabled:  true,
			MemoryWriteEnabled: true,
		}
	}
	out := EffectiveConfig{
		Profile:             profile,
		PromptFragments:     append([]string(nil), profile.PromptFragments...),
		FewShotSnippets:     append([]string(nil), profile.FewShotSnippets...),
		ToolPreferenceOrder: append([]string(nil), profile.ToolPreferenceOrder...),
		RetrievalBias:       profile.RetrievalBias,
		ReasoningVerbosity:  firstNonEmpty(profile.ReasoningVerbosity, fallbackReasoningVerbosity),
		MemoryReadEnabled:   profile.MemoryReadEnabled,
		MemoryWriteEnabled:  profile.MemoryWriteEnabled,
	}
	if overlay, ok := store.GetActiveHarnessOverlay(role); ok {
		out.Overlay = &overlay
		out.EffectiveOverlayID = overlay.ID
		out.EffectiveOverlayVersion = overlay.Version
		out.PromptFragments = append(out.PromptFragments, overlay.PromptFragments...)
		out.FewShotSnippets = append(out.FewShotSnippets, overlay.FewShotSnippets...)
		if len(overlay.ToolPreferenceOrder) > 0 {
			out.ToolPreferenceOrder = append([]string(nil), overlay.ToolPreferenceOrder...)
		}
		out.RetrievalBias = firstNonEmpty(overlay.RetrievalBias, out.RetrievalBias)
		out.ReasoningVerbosity = firstNonEmpty(overlay.ReasoningVerbosity, out.ReasoningVerbosity)
		if overlay.MemoryReadEnabled != nil {
			out.MemoryReadEnabled = *overlay.MemoryReadEnabled
		}
		if overlay.MemoryWriteEnabled != nil {
			out.MemoryWriteEnabled = *overlay.MemoryWriteEnabled
		}
	}
	return out
}

func ComposeSystemMessage(base string, effective EffectiveConfig) string {
	parts := []string{}
	if trimmed := strings.TrimSpace(base); trimmed != "" {
		parts = append(parts, trimmed)
	}
	if len(effective.PromptFragments) > 0 {
		parts = append(parts, "Harness prompt fragments:\n- "+strings.Join(compactStrings(effective.PromptFragments), "\n- "))
	}
	if len(effective.FewShotSnippets) > 0 {
		parts = append(parts, "Harness few-shot snippets:\n- "+strings.Join(compactStrings(effective.FewShotSnippets), "\n- "))
	}
	if effective.RetrievalBias != "" {
		parts = append(parts, "Harness retrieval bias: "+effective.RetrievalBias)
	}
	if effective.ReasoningVerbosity != "" {
		parts = append(parts, "Harness reasoning verbosity: "+effective.ReasoningVerbosity)
	}
	parts = append(parts, "Hermes memory writes and recalled context must be explicit, reviewable, and never include hidden provider chain-of-thought.")
	return strings.Join(compactStrings(parts), "\n\n")
}

func ApplyToolPreference(allowed []string, preference []string) []string {
	if len(allowed) == 0 {
		return []string{}
	}
	allowed = append([]string(nil), allowed...)
	if len(preference) == 0 {
		return allowed
	}
	rank := map[string]int{}
	for idx, name := range preference {
		rank[strings.TrimSpace(name)] = idx + 1
	}
	sort.SliceStable(allowed, func(i, j int) bool {
		left := rank[strings.TrimSpace(allowed[i])]
		right := rank[strings.TrimSpace(allowed[j])]
		switch {
		case left == 0 && right == 0:
			return allowed[i] < allowed[j]
		case left == 0:
			return false
		case right == 0:
			return true
		default:
			return left < right
		}
	})
	return allowed
}

func DecodeExecutionMetadata(raw map[string]any) ExecutionMetadata {
	if raw == nil {
		return ExecutionMetadata{}
	}
	out := ExecutionMetadata{
		HermesSessionID:         stringFromMap(raw, "hermes_session_id"),
		ParentSessionID:         stringFromMap(raw, "parent_session_id"),
		HarnessProfileID:        stringFromMap(raw, "harness_profile_id"),
		EffectiveOverlayID:      stringFromMap(raw, "effective_overlay_id"),
		EffectiveOverlayVersion: stringFromMap(raw, "effective_overlay_version"),
		MemoryBackend:           stringFromMap(raw, "memory_backend"),
		AssistantPeerID:         stringFromMap(raw, "assistant_peer_id"),
		UserPeerID:              stringFromMap(raw, "user_peer_id"),
		SessionScopeKind:        stringFromMap(raw, "session_scope_kind"),
		SessionScopeID:          stringFromMap(raw, "session_scope_id"),
		ParentScopeKind:         stringFromMap(raw, "parent_session_scope_kind"),
		ParentScopeID:           stringFromMap(raw, "parent_session_scope_id"),
	}
	out.MemoryReads = decodeMemoryArtifacts(raw["memory_reads"])
	out.MemoryWrites = decodeMemoryArtifacts(raw["memory_writes"])
	return out
}

func compactStrings(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func stringFromMap(values map[string]any, key string) string {
	raw, ok := values[key]
	if !ok || raw == nil {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(strings.TrimSpace(toJSONString(value)))
	}
}

func decodeMemoryArtifacts(raw any) []MemoryArtifact {
	if raw == nil {
		return []MemoryArtifact{}
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return []MemoryArtifact{}
	}
	var out []MemoryArtifact
	if err := json.Unmarshal(data, &out); err != nil {
		return []MemoryArtifact{}
	}
	if out == nil {
		return []MemoryArtifact{}
	}
	return out
}

func toJSONString(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}
