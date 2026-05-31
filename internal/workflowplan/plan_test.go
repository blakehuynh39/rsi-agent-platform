package workflowplan

import (
	"testing"
	"time"

	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
)

func TestBuildLiveHintsKeepsDeterministicToolHintsOutOfContext(t *testing.T) {
	hints := BuildLiveHints(RuntimeConfig{
		DefaultRepo: "depin-backend",
	}, RequestContext{
		Question: "Draw the depin-backend architecture.",
	}, time.Unix(0, 0).UTC())

	if hints.Repo != "depin-backend" {
		t.Fatalf("repo = %q, want depin-backend", hints.Repo)
	}
	if len(hints.CandidateReadSurfaces) != 0 {
		t.Fatalf("candidate read surfaces = %#v, want none", hints.CandidateReadSurfaces)
	}
}

func TestBuildToolRequestPayloadAddsRepoActivityWindowOnlyForActivityRequests(t *testing.T) {
	now := time.Date(2026, 5, 7, 17, 15, 39, 0, time.UTC)
	cfg := RuntimeConfig{
		DefaultRepo:  "rsi-agent-platform",
		AllowedRepos: []string{"depin-backend", "rsi-agent-platform"},
	}

	generic := BuildToolRequestPayload(cfg, RequestContext{
		Question: "Draw the depin-backend architecture.",
	}, now)
	if _, ok := generic["since"]; ok {
		t.Fatalf("generic request should not get since hint, got %#v", generic["since"])
	}
	if _, ok := generic["until"]; ok {
		t.Fatalf("generic request should not get until hint, got %#v", generic["until"])
	}

	activity := BuildToolRequestPayload(cfg, RequestContext{
		Question: "Summarize depin-backend PR activity from last week.",
	}, now)
	since, ok := activity["since"].(string)
	if !ok || since == "" {
		t.Fatalf("activity request should get since hint, got %#v", activity["since"])
	}
	until, ok := activity["until"].(string)
	if !ok || until == "" {
		t.Fatalf("activity request should get until hint, got %#v", activity["until"])
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

func containsSurface(surfaces []SlackSurfaceHint, target SlackSurfaceHint) bool {
	for _, item := range surfaces {
		if item.ChannelID == target.ChannelID && item.ThreadTS == target.ThreadTS && item.Source == target.Source {
			return true
		}
	}
	return false
}
