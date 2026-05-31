# RSI Runner

Thin Python wrapper around Hermes runtime usage for the RSI agent platform.

The wrapper is intentionally small:

- health/readiness HTTP endpoints
- role-aware execution mode
- Hermes `AIAgent` execution harness
- structured task execution requests with repo scope, allowed tools/commands,
  expected outputs, artifact destinations, and rejected-proposal context
- explicit runner roles via `RSI_RUNNER_ROLE`: `prod`, `proactive`, `eval`, `proposal`

Default runtime:

- `RSI_RUNNER_PROVIDER=deepseek`
- `RSI_RUNNER_MODEL=deepseek-v4-pro`
- `RSI_RUNNER_REASONING_EFFORT=xhigh`
- `RSI_RUNNER_THINKING=enabled`
- `RSI_SUMMARY_MODEL=deepseek-v4-flash`
- `RSI_SUMMARY_THINKING=disabled`

`/runtimez` reports the effective role, backend, provider, model, thinking mode, summary model, API mode, and reasoning effort so improvement-plane can surface the live runtime configuration in the operator UI.

## Hermes Fork Pin

The executor image installs `hermes-agent` from the RSI infra fork:

- source repo: `https://github.com/blakehuynh39/hermes-agent.git`
- current pin: `0aa2b52d52e6fdaf7992b9d6ac224573f3212f5d`

Keep that fork's `main` branch containing the pinned commit before merging an
RSI platform pin bump. Do not pin the executor image to upstream
`NousResearch/hermes-agent` for infra-specific patches; upstream can lag or
decline changes that are only needed for the Story company-computer runtime.
