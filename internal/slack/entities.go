package slack

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	userMentionPattern         = regexp.MustCompile(`<@([A-Z0-9]+)(?:\|([^>]+))?>`)
	channelMentionPattern      = regexp.MustCompile(`<#([A-Z0-9]+)(?:\|([^>]+))?>`)
	plainUserMentionPattern    = regexp.MustCompile(`(^|[^A-Za-z0-9])@([A-Z0-9]{8,})($|[^A-Za-z0-9])`)
	plainChannelMentionPattern = regexp.MustCompile(`(^|[^A-Za-z0-9])#([A-Z0-9]{8,})($|[^A-Za-z0-9])`)
)

func ExtractEntityRefs(text string) []EntityRef {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	out := []EntityRef{}
	seen := map[string]struct{}{}
	appendRef := func(kind EntityKind, id string, label string, source string) {
		id = strings.ToUpper(strings.TrimSpace(id))
		if id == "" {
			return
		}
		key := string(kind) + ":" + id
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, EntityRef{
			Kind:   kind,
			ID:     id,
			Label:  strings.TrimSpace(label),
			Source: strings.TrimSpace(source),
		})
	}
	for _, match := range userMentionPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		appendRef(EntityUser, match[1], matchValue(match, 2), "mrkdwn")
	}
	for _, match := range channelMentionPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		appendRef(EntityChannel, match[1], matchValue(match, 2), "mrkdwn")
	}
	for _, match := range plainUserMentionPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 3 {
			continue
		}
		appendRef(EntityUser, match[2], "", "plain_text")
	}
	for _, match := range plainChannelMentionPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 3 {
			continue
		}
		appendRef(EntityChannel, match[2], "", "plain_text")
	}
	return out
}

func EntityRefsFromValue(value any) []EntityRef {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []EntityRef:
		return normalizeEntityRefs(append([]EntityRef(nil), typed...))
	case []map[string]any:
		data, err := json.Marshal(typed)
		if err != nil {
			return nil
		}
		var out []EntityRef
		if err := json.Unmarshal(data, &out); err != nil {
			return nil
		}
		return normalizeEntityRefs(out)
	case []any:
		data, err := json.Marshal(typed)
		if err != nil {
			return nil
		}
		var out []EntityRef
		if err := json.Unmarshal(data, &out); err != nil {
			return nil
		}
		return normalizeEntityRefs(out)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		var out []EntityRef
		if err := json.Unmarshal(data, &out); err != nil {
			return nil
		}
		return normalizeEntityRefs(out)
	}
}

func normalizeEntityRefs(items []EntityRef) []EntityRef {
	out := make([]EntityRef, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item.ID = strings.ToUpper(strings.TrimSpace(item.ID))
		item.Label = strings.TrimSpace(item.Label)
		item.Source = strings.TrimSpace(item.Source)
		if item.ID == "" || item.Kind == "" {
			continue
		}
		key := string(item.Kind) + ":" + item.ID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func matchValue(match []string, idx int) string {
	if idx < 0 || idx >= len(match) {
		return ""
	}
	return strings.TrimSpace(match[idx])
}
