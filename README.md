# RSI Agent Platform

Go-first platform control stack for the RSI agent factory, with Hermes retained as
the Python execution runtime for live, proactive, eval, proposal, and repo-change workloads.

## Layout

- `cmd/control-plane`: Slack ingress, workflow/session APIs, routing, policy, approval, orchestration
- `cmd/tool-gateway`: typed integration facade
- `cmd/improvement-plane`: trace/review APIs, eval/proposal cron mode, and embedded eval UI
- `internal/control`: control-plane HTTP APIs plus Slack socket-mode surface
- `internal/*`: shared contracts, storage, registries, and review/event logic
- `runner/`: Python Hermes runner wrapper
- `ui/eval-web`: React + Vite review UI
- `sandbox/`: sandbox runtime image definition

## Quick start

This repo is pinned to Go `1.26.2`.

```bash
make ci
go run ./cmd/improvement-plane
```

Use the shared Postgres-backed store in production:

```bash
export RSI_ENV=production
export RSI_STORE_BACKEND=postgres
export RSI_POSTGRES_URL=postgres://localhost:5432/rsi_agent_platform
go run ./cmd/control-plane
go run ./cmd/control-plane --mode slack-surface
go run ./cmd/improvement-plane
go run ./cmd/improvement-plane --mode cron --once
```

Build the UI when frontend changes land:

```bash
make ui-build
```

Run the Hermes runner wrapper:

```bash
cd runner
python3 -m rsi_runner.main
```

The runner accepts the legacy prompt contract and the new structured task contract for repo-scoped eval and patch-generation work. Set `RSI_RUNNER_ROLE` to `prod`, `proactive`, `eval`, or `proposal` to gate allowed task types.

`control-plane --mode slack-surface` uses the legacy Slack env contract:
`RSI_SLACK_APP_IDENTITY`, `RSI_SLACK_SOCKET_MODE_ENABLED`, `RSI_SLACK_APP_TOKEN`, and `RSI_SLACK_BOT_TOKEN`.

## CI/CD

GitHub Actions is split into:

- PR/push CI in `.github/workflows/ci.yml`
- stage image build + deploy-repo bump in `.github/workflows/cd.yml`

The CD workflow builds and pushes five stage images on `main`:

- `rsi-agent-platform:control-plane-<sha>`
- `rsi-agent-platform:tool-gateway-<sha>`
- `rsi-agent-platform:improvement-plane-<sha>`
- `rsi-agent-platform-runner:runner-<sha>`
- `rsi-agent-platform-sandbox:sandbox-<sha>`

After pushing images, CD updates
`storyprotocol/story-deployments:rsi-platform/rsi-agent-platform/use1-stage.yaml`
using the same token-driven pattern already used by `depin-backend`.

Repository secrets required for CD:

- `AWS_ECR_PUSH_ROLE_ARN`
- `STORY_DEPLOYMENTS_PUSH_TOKEN`
