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

- `RSI_RUNNER_MODEL=openai/gpt-5.4`
- `RSI_RUNNER_REASONING_EFFORT=xhigh`
- Hermes `api_mode=codex_responses` for `openai/*` models

`/runtimez` reports the effective role, backend, provider, model, API mode, and reasoning effort so improvement-plane can surface the live runtime configuration in the operator UI.

## Hermes Fork Pin

The executor image installs `hermes-agent` from the RSI infra fork:

- source repo: `https://github.com/blakehuynh39/hermes-agent.git`
- current pin: `4819f068a5e3b6db33a712374761c638e4d8c44e`

Keep that fork's `main` branch containing the pinned commit before merging an
RSI platform pin bump. Do not pin the executor image to upstream
`NousResearch/hermes-agent` for infra-specific patches; upstream can lag or
decline changes that are only needed for the Story company-computer runtime.
