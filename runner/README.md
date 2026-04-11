# RSI Runner

Thin Python wrapper around Hermes runtime usage for the RSI agent platform.

The wrapper is intentionally small:

- health/readiness HTTP endpoints
- role-aware execution mode
- Hermes `AIAgent` execution harness
- graceful fallback when Hermes is not installed in the environment
- structured task execution requests with repo scope, allowed tools/commands,
  expected outputs, artifact destinations, and rejected-proposal context
- explicit runner roles via `RSI_RUNNER_ROLE`: `prod`, `proactive`, `eval`, `proposal`

Default runtime:

- `RSI_RUNNER_MODEL=openai/gpt-5.4`
- `RSI_RUNNER_REASONING_EFFORT=xhigh`
- Hermes `api_mode=codex_responses` for `openai/*` models

`/runtimez` reports the effective role, backend, provider, model, API mode, and reasoning effort so improvement-plane can surface the live runtime configuration in the operator UI.
