# RSI Agent Platform

Go-first platform control stack for the RSI agent factory, with Hermes retained as
the Python execution runtime for live, proactive, eval, and proposal workloads.

## Layout

- `cmd/workflow-api`: Slack ingestion and workflow/session APIs
- `cmd/control-plane`: routing, policy, approval, sandbox lifecycle
- `cmd/tool-gateway`: typed integration facade
- `cmd/improvement-plane`: trace/review APIs and embedded eval UI
- `internal/*`: shared contracts, storage, registries, and review/event logic
- `runner/`: Python Hermes runner wrapper
- `ui/eval-web`: React + Vite review UI
- `sandbox/`: sandbox runtime image definition

## Quick start

```bash
go test ./...
go run ./cmd/improvement-plane
```

Build the UI when frontend changes land:

```bash
cd ui/eval-web
pnpm install
pnpm build
```

Run the Hermes runner wrapper:

```bash
cd runner
python3 -m rsi_runner.main
```

