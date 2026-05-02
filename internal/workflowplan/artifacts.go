package workflowplan

import (
	"strings"

	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
)

type RequestedArtifact struct {
	Kind        string
	Description string
}

func ArtifactRequestText(userRequest string, prompt slackpkg.SlackPromptEnvelope) string {
	if rendered := strings.TrimSpace(prompt.RenderedText); rendered != "" {
		return rendered
	}
	if raw := strings.TrimSpace(prompt.RawText); raw != "" {
		return raw
	}
	return strings.TrimSpace(userRequest)
}

func RequestedArtifactsForPrompt(userRequest string, prompt slackpkg.SlackPromptEnvelope) []RequestedArtifact {
	lower := strings.ToLower(ArtifactRequestText(userRequest, prompt))
	if lower == "" {
		return nil
	}
	requested := []RequestedArtifact{}
	if strings.Contains(lower, "/architecture-diagram") ||
		(strings.Contains(lower, "diagram") && (strings.Contains(lower, "architecture") || strings.Contains(lower, "architectural") || strings.Contains(lower, "system"))) ||
		(strings.Contains(lower, "draw") && strings.Contains(lower, "diagram")) {
		requested = append(requested, RequestedArtifact{
			Kind:        "diagram",
			Description: "Render the requested system or architecture diagram as a first-class artifact.",
		})
	}
	if strings.Contains(lower, "attachment") ||
		strings.Contains(lower, "attached") ||
		strings.Contains(lower, "rendered output") ||
		(strings.Contains(lower, "render") &&
			(strings.Contains(lower, "artifact") ||
				strings.Contains(lower, "attachment") ||
				strings.Contains(lower, "attached") ||
				strings.Contains(lower, "image") ||
				strings.Contains(lower, "file"))) {
		requested = append(requested, RequestedArtifact{
			Kind:        "rendered_output",
			Description: "Provide a rendered output artifact when the request calls for one.",
		})
	}
	return dedupeRequestedArtifacts(requested)
}

func RequestedArtifactsForUserRequest(userRequest string) []RequestedArtifact {
	return RequestedArtifactsForPrompt(userRequest, slackpkg.SlackPromptEnvelope{})
}

func dedupeRequestedArtifacts(items []RequestedArtifact) []RequestedArtifact {
	if len(items) == 0 {
		return nil
	}
	out := make([]RequestedArtifact, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := strings.TrimSpace(item.Kind)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}
