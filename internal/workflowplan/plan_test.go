package workflowplan

import "testing"

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
