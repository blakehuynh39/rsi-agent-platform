package toolcatalog

func GovernedReadOnlyToolNames() []string {
	return []string{
		"repo.context",
		"repo.read_file",
		"repo.search",
		"knowledge.context",
		"rsi.trace_context",
		"rsi.workflow_context",
		"rsi.action_chain",
		"rsi.runner_execution",
		"rsi.runtime_config",
		"rsi.runtime_health",
		"rsi.runtime_deployment_facts",
		"rsi.proposal_memory",
		"rsi.candidate_context",
		"rsi.attempt_context",
		"slack.history",
		"slack.search",
		"github.repo_activity",
		"github.repo_context",
		"sentry.lookup",
		"kubernetes.inspect",
		"kubernetes.logs",
		"kubernetes.events",
		"cloudflare.inspect",
	}
}

func GovernedWorkspaceToolNames() []string {
	return []string{
		"workspace.list_files",
		"workspace.read_file",
		"workspace.search",
		"workspace.git_history",
		"workspace.git_show",
		"workspace.git_search",
		"workspace.write_file",
		"workspace.apply_patch",
		"workspace.git_status",
		"workspace.git_diff",
		"workspace.run_validation",
	}
}
