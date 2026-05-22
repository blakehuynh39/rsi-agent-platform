package harness

import "time"

const (
	DefaultMemoryBackend = "honcho"
	DefaultModel         = "deepseek/deepseek-v4-pro"
	LegacyDefaultModel   = "openai/gpt-5.4"
)

func DefaultProfileID(role string) string {
	return "harness-profile-" + role
}

func SeedProfiles(now time.Time) []Profile {
	return []Profile{
		{
			ID:                  DefaultProfileID("prod"),
			Role:                "prod",
			Name:                "Production Operator",
			Description:         "Live conversation and incident workflow agent with durable memory and explicit evidence-first reasoning.",
			Model:               DefaultModel,
			ReasoningEffort:     "xhigh",
			PromptFragments:     []string{"Ground answers in explicit evidence. Prefer concrete repo, Slack, and tool context over generic advice."},
			ToolPreferenceOrder: []string{"repo.context", "knowledge.context", "github.repo_activity", "sentry.lookup", "kubernetes.logs"},
			RetrievalBias:       "canonical_then_working_then_session",
			ReasoningVerbosity:  "verbose",
			MemoryReadEnabled:   true,
			MemoryWriteEnabled:  true,
			RepoRef:             "main",
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  DefaultProfileID("proactive"),
			Role:                "proactive",
			Name:                "Proactive Thread Agent",
			Description:         "Monitors and joins conversations when evidence justifies intervention.",
			Model:               DefaultModel,
			ReasoningEffort:     "xhigh",
			PromptFragments:     []string{"Intervene only when the evidence supports a useful reply or workflow launch."},
			ToolPreferenceOrder: []string{"knowledge.context", "repo.context", "github.repo_activity"},
			RetrievalBias:       "canonical_then_session",
			ReasoningVerbosity:  "verbose",
			MemoryReadEnabled:   true,
			MemoryWriteEnabled:  true,
			RepoRef:             "main",
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  DefaultProfileID("eval"),
			Role:                "eval",
			Name:                "Eval Analyst",
			Description:         "Summarizes failures, compares traces, and improves recurring eval lines without hiding uncertainty.",
			Model:               DefaultModel,
			ReasoningEffort:     "xhigh",
			PromptFragments:     []string{"Focus on observable evidence, failure patterns, and novelty relative to prior rejected proposals."},
			ToolPreferenceOrder: []string{"knowledge.context"},
			RetrievalBias:       "canonical_then_session",
			ReasoningVerbosity:  "verbose",
			MemoryReadEnabled:   true,
			MemoryWriteEnabled:  true,
			RepoRef:             "main",
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  DefaultProfileID("proposal"),
			Role:                "proposal",
			Name:                "Proposal Materializer",
			Description:         "Turns approved candidate lines into governed repo-change or overlay-ready reasoning with prior memory context.",
			Model:               DefaultModel,
			ReasoningEffort:     "xhigh",
			PromptFragments:     []string{"Respect proposal memory, review rationale, and rollback expectations before materializing work."},
			ToolPreferenceOrder: []string{"knowledge.context", "repo.context"},
			RetrievalBias:       "canonical_then_working_then_session",
			ReasoningVerbosity:  "verbose",
			MemoryReadEnabled:   true,
			MemoryWriteEnabled:  true,
			RepoRef:             "main",
			CreatedAt:           now,
			UpdatedAt:           now,
		},
	}
}
