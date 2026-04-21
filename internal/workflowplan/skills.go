package workflowplan

import (
	"regexp"
	"strings"
)

var slashSkillTokenPattern = regexp.MustCompile(`(?i)(?:^|[\s(])/(?:[a-z0-9][a-z0-9_-]*)`)

func RequestedSkillsForPrompt(userRequest string, prompt any) []string {
	return ExplicitSkillMentionsForPrompt(userRequest, prompt)
}

func ExplicitSkillMentionsForPrompt(userRequest string, prompt any) []string {
	text := ArtifactRequestText(userRequest, prompt)
	if strings.TrimSpace(text) == "" {
		return nil
	}
	matches := slashSkillTokenPattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		token := strings.TrimSpace(match)
		token = strings.TrimLeft(token, "(")
		token = strings.TrimPrefix(token, "/")
		if token == "" {
			continue
		}
		out = append(out, normalizeSkillName(token))
	}
	return dedupeSkillNames(out)
}

func dedupeSkillNames(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		normalized := normalizeSkillName(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeSkillName(value string) string {
	text := strings.TrimSpace(value)
	text = strings.TrimPrefix(text, "/")
	text = strings.ReplaceAll(text, "_", "-")
	return strings.ToLower(text)
}
