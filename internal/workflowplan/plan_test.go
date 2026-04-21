package workflowplan

import (
	"testing"

	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
)

func TestToolPlanAddsSlackSearchAndRuntimeDeploymentFactsForSlackDiscoveryQuestion(t *testing.T) {
	plan := ToolPlan(
		"question",
		"Search Slack for where we decided to bump the control plane to 5 minutes.",
		"rsi-agent-platform",
		"C123",
		"171000001.000100",
	)

	if !containsTool(plan, "slack.search") {
		t.Fatalf("expected slack.search in plan, got %#v", plan)
	}
	if !containsTool(plan, "rsi.runtime_deployment_facts") {
		t.Fatalf("expected rsi.runtime_deployment_facts in plan, got %#v", plan)
	}
}

func TestToolPlanAddsRuntimeDeploymentFactsWithoutSlackSearchWhenNoChannelBinding(t *testing.T) {
	plan := ToolPlan(
		"question",
		"What image is running on tool gateway right now?",
		"rsi-agent-platform",
		"",
		"",
	)

	if !containsTool(plan, "rsi.runtime_deployment_facts") {
		t.Fatalf("expected rsi.runtime_deployment_facts in plan, got %#v", plan)
	}
	if containsTool(plan, "slack.search") {
		t.Fatalf("did not expect slack.search in plan without a bound channel, got %#v", plan)
	}
}

func TestCandidateReadSurfacesKeepsRawThreadRefsUnbound(t *testing.T) {
	surfaces := CandidateReadSurfaces(
		"Check <#COTHER> thread_ts=1776483985.407559 for the rollout note.",
		"CINGRESS",
		"1776483000.000100",
	)

	if !containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "CINGRESS",
		ThreadTS:  "1776483000.000100",
		Source:    "ingress_thread",
	}) {
		t.Fatalf("expected ingress thread surface, got %#v", surfaces)
	}
	if !containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "COTHER",
		ThreadTS:  "",
		Source:    "channel_mention",
	}) {
		t.Fatalf("expected mentioned channel surface, got %#v", surfaces)
	}
	if !containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "",
		ThreadTS:  "1776483985.407559",
		Source:    "explicit_thread_ref",
	}) {
		t.Fatalf("expected unbound thread surface, got %#v", surfaces)
	}
	if containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "CINGRESS",
		ThreadTS:  "1776483985.407559",
		Source:    "explicit_thread_ref",
	}) {
		t.Fatalf("did not expect explicit thread ref to inherit ingress channel, got %#v", surfaces)
	}
}

func TestCandidateReadSurfacesParsesPlainSlackChannelIDs(t *testing.T) {
	surfaces := CandidateReadSurfaces(
		"Please review #C0AKH5SNGKH and #C0AL7EKNHDF for the latest NUMO discussion.",
		"CINGRESS",
		"1776483000.000100",
	)

	if !containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "CINGRESS",
		ThreadTS:  "1776483000.000100",
		Source:    "ingress_thread",
	}) {
		t.Fatalf("expected ingress thread surface, got %#v", surfaces)
	}
	if !containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "C0AKH5SNGKH",
		Source:    "channel_mention",
	}) {
		t.Fatalf("expected first plain channel mention surface, got %#v", surfaces)
	}
	if !containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "C0AL7EKNHDF",
		Source:    "channel_mention",
	}) {
		t.Fatalf("expected second plain channel mention surface, got %#v", surfaces)
	}
}

func TestCandidateReadSurfacesForContextPrefersStructuredEntityRefs(t *testing.T) {
	surfaces := CandidateReadSurfacesForContext(RequestContext{
		Question:  "Please use the latest discussions from Slack for this summary.",
		ChannelID: "CINGRESS",
		ThreadTS:  "1776483000.000100",
		EntityRefs: []slackpkg.EntityRef{
			{Kind: slackpkg.EntityChannel, ID: "C0AKH5SNGKH", Source: "mrkdwn"},
			{Kind: slackpkg.EntityChannel, ID: "C0AL7EKNHDF", Source: "mrkdwn"},
		},
	})

	if !containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "C0AKH5SNGKH",
		Source:    "entity_ref",
	}) {
		t.Fatalf("expected first structured channel surface, got %#v", surfaces)
	}
	if !containsSurface(surfaces, SlackSurfaceHint{
		ChannelID: "C0AL7EKNHDF",
		Source:    "entity_ref",
	}) {
		t.Fatalf("expected second structured channel surface, got %#v", surfaces)
	}
}

func TestRequestedArtifactsForUserRequestDetectsDiagramAndRenderedOutput(t *testing.T) {
	items := RequestedArtifactsForUserRequest("Please draw an architecture diagram and attach the rendered output.")
	if len(items) != 2 {
		t.Fatalf("expected two requested artifacts, got %#v", items)
	}
	if items[0].Kind != "diagram" {
		t.Fatalf("expected diagram artifact first, got %#v", items)
	}
	if items[1].Kind != "rendered_output" {
		t.Fatalf("expected rendered_output artifact second, got %#v", items)
	}
}

func TestRequestedArtifactsForUserRequestDedupesRepeatedSignals(t *testing.T) {
	items := RequestedArtifactsForUserRequest("Use /architecture-diagram and draw a system diagram, then render an attachment attached to the reply.")
	if len(items) != 2 {
		t.Fatalf("expected deduped artifacts, got %#v", items)
	}
}

func TestRequestedArtifactsForPromptUsesRenderedPromptText(t *testing.T) {
	items := RequestedArtifactsForPrompt("Summarize the architecture work.", slackpkg.SlackPromptEnvelope{
		RawText:      "@RSI can you draw an architecture diagram using /architecture-diagram",
		RenderedText: "@RSI can you draw an architecture diagram using /architecture-diagram",
	})
	if len(items) != 1 || items[0].Kind != "diagram" {
		t.Fatalf("expected diagram artifact from prompt envelope, got %#v", items)
	}
}

func TestRequestedArtifactsForUserRequestDoesNotTriggerRenderedOutputOnBareRender(t *testing.T) {
	items := RequestedArtifactsForUserRequest("Please render a summary of the rollout status.")
	if len(items) != 0 {
		t.Fatalf("expected no requested artifacts for bare render phrasing, got %#v", items)
	}
}

func TestRequestedSkillsForPromptIncludesExplicitMention(t *testing.T) {
	items := RequestedSkillsForPrompt(
		"Please use /architecture-diagram for this system diagram.",
		nil,
	)
	if len(items) != 1 || items[0] != "architecture-diagram" {
		t.Fatalf("expected one deduped architecture skill, got %#v", items)
	}
}

func TestRequestedSkillsForPromptDetectsArchitectureDiagramFromEnvelope(t *testing.T) {
	items := RequestedSkillsForPrompt(
		"Summarize the architecture work.",
		slackpkg.SlackPromptEnvelope{
			RawText:      "@RSI can you draw an architecture diagram of depin-backend? Use /architecture-diagram skill",
			RenderedText: "@RSI can you draw an architecture diagram of depin-backend? Use /architecture-diagram skill",
		},
	)
	if len(items) != 1 || items[0] != "architecture-diagram" {
		t.Fatalf("expected architecture-diagram from prompt envelope, got %#v", items)
	}
}

func TestRequestedSkillsForPromptDoesNotAddAutomaticArchitectureHint(t *testing.T) {
	items := RequestedSkillsForPrompt(
		"Please draw the architecture of depin-backend.",
		nil,
	)
	if len(items) != 0 {
		t.Fatalf("expected no automatic architecture-diagram hint, got %#v", items)
	}
}

func TestRequestedSkillsForPromptDoesNotTriggerOnBareSystemLanguage(t *testing.T) {
	items := RequestedSkillsForPrompt(
		"What system calls does this make?",
		nil,
	)
	if len(items) != 0 {
		t.Fatalf("expected no architecture-diagram hint for bare system language, got %#v", items)
	}
}

func containsTool(plan []string, tool string) bool {
	for _, item := range plan {
		if item == tool {
			return true
		}
	}
	return false
}

func containsSurface(surfaces []SlackSurfaceHint, target SlackSurfaceHint) bool {
	for _, item := range surfaces {
		if item.ChannelID == target.ChannelID && item.ThreadTS == target.ThreadTS && item.Source == target.Source {
			return true
		}
	}
	return false
}
