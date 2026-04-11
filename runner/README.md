# RSI Runner

Thin Python wrapper around Hermes runtime usage for the RSI agent platform.

The wrapper is intentionally small:

- health/readiness HTTP endpoints
- role-aware execution mode
- optional direct import of `AIAgent` from Hermes
- graceful fallback when Hermes is not installed in the environment
- structured task execution requests with repo scope, allowed tools/commands,
  expected outputs, artifact destinations, and rejected-proposal context
- explicit runner roles via `RSI_RUNNER_ROLE`: `prod`, `proactive`, `eval`, `proposal`
