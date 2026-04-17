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

func containsTool(plan []string, tool string) bool {
	for _, item := range plan {
		if item == tool {
			return true
		}
	}
	return false
}
