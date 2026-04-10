# RSI Runner

Thin Python wrapper around Hermes runtime usage for the RSI agent platform.

The wrapper is intentionally small:

- health/readiness HTTP endpoints
- role-aware execution mode
- optional direct import of `AIAgent` from Hermes
- graceful fallback when Hermes is not installed in the environment

