# RSI Agent Platform

Go-first platform control stack for the RSI agent factory, with Hermes retained as
the Python execution runtime for live, proactive, eval, proposal, and repo-change workloads.

## Layout

- `cmd/control-plane`: Slack ingress, workflow/session APIs, routing, policy, approval, orchestration
- `cmd/improvement-plane`: trace/review APIs, eval/proposal cron mode, and embedded eval UI
- `internal/control`: control-plane HTTP APIs plus Slack socket-mode surface
- `internal/*`: shared contracts, storage, registries, and review/event logic
- `honcho/`: pinned self-hosted Honcho image build for stage memory services, with the custom fork source documented in [`honcho/README.md`](./honcho/README.md)
- `runner/`: Python Hermes runner wrapper
- `ui/eval-web`: React + Vite review UI
- `sandbox/`: sandbox runtime image definition

## Quick start

This repo is pinned to Go `1.26.2`.

```bash
make ci
make db-migrate
go run ./cmd/improvement-plane
```

Use the shared Postgres-backed store in production:

```bash
export RSI_ENV=production
export RSI_STORE_BACKEND=postgres
export RSI_POSTGRES_URL=postgres://localhost:5432/rsi_agent_platform
go run ./cmd/improvement-plane --mode migrate
go run ./cmd/control-plane
go run ./cmd/control-plane --mode slack-surface
go run ./cmd/improvement-plane
go run ./cmd/improvement-plane --mode cron --once
```

Normal service startup no longer applies schema. Database changes are forward-only SQL
migrations under `internal/db/migrations`, and `improvement-plane --mode migrate` is the
only schema mutator in local, stage, and prod.

Build the UI when frontend changes land:

```bash
make ui-build
```

Run the Hermes runner wrapper:

```bash
cd runner
python3 -m rsi_runner.main
```

The runner uses the structured task contract for repo-scoped eval, proposal, and workspace work. Set `RSI_RUNNER_ROLE` to `prod`, `proactive`, `eval`, or `proposal` to gate allowed task types.

Runtime defaults:

- model: `openai/gpt-5.4`
- reasoning effort: `xhigh`
- Hermes harness: `AIAgent`
- OpenAI transport mode: `codex_responses`

For `openai/*` models, the runner uses Hermes directly and forwards the configured `reasoning_effort` into Hermes `reasoning_config`. The effective runtime is exposed at `/runtimez` on each runner and aggregated by `improvement-plane` at `/api/runtime`.

`control-plane --mode slack-surface` uses the Slack env contract:
`RSI_SLACK_APP_IDENTITY`, `RSI_SLACK_SOCKET_MODE_ENABLED`, `RSI_SLACK_APP_TOKEN`, and `SLACK_BOT_TOKEN`.
Hermes executor pods use RSI native tools for Slack/Notion/knowledge access and
delivery. Configure `RSI_NATIVE_TOOLS_CLIENT_TOKEN`; do not route RSI workflow
delivery through generic Hermes `send_message` or Slack/Notion MCP profiles.

## CI/CD

GitHub Actions is split into:

- PR/push CI in `.github/workflows/ci.yml`
- stage image build + deploy-repo bump in `.github/workflows/cd.yml`

CI also runs Postgres-backed migration and store integration tests against a
`pgvector`-enabled Postgres image. Set
`RSI_TEST_POSTGRES_URL` locally and run `make test-postgres` to exercise the same path.
The stage acceptance runbook for the persistence hardening rollout lives at
[`docs/persistence-hardening-stage-acceptance.md`](./docs/persistence-hardening-stage-acceptance.md).
The Honcho stage rollout and rollback runbook lives at
[`docs/honcho-stage-rollout.md`](./docs/honcho-stage-rollout.md).
The Slack-approved Postgres DB read gateway architecture and security-review
notes live at [`docs/db-read-gateway-architecture.md`](./docs/db-read-gateway-architecture.md).
The self-hosted Firecrawl web-search architecture lives at
[`docs/firecrawl-web-search-architecture.md`](./docs/firecrawl-web-search-architecture.md).

The CD workflow builds and pushes five stage images on `main`:

- `rsi-agent-platform:control-plane-<sha>`
- `rsi-agent-platform:improvement-plane-<sha>`
- `rsi-agent-platform-hermes-executor:hermes-executor-<sha>`
- `rsi-agent-platform-hermes-skill-exporter:hermes-skill-exporter-<sha>`
- `rsi-agent-platform-honcho:honcho-<sha>`

The Honcho image is currently built from a pinned commit in
`blakehuynh39/honcho`, not from a vanilla upstream image. See
[`honcho/README.md`](./honcho/README.md) for the exact fork commit currently in
use and the planned migration path to a dedicated `piplabs/honcho` fork.

After pushing images, CD updates
`storyprotocol/story-deployments:rsi-platform/rsi-agent-platform/use1-stage.yaml`
using the same token-driven pattern already used by `depin-backend`.

Repository secrets required for CD:

- `AWS_ECR_PUSH_ROLE_ARN`
- `STORY_DEPLOYMENTS_PUSH_TOKEN`
