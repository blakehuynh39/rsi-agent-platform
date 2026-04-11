GO ?= go
PNPM ?= pnpm
PYTHON ?= python3

.PHONY: test build ui-build runner-smoke improvement-cron-once

test:
	$(GO) test ./...

build:
	$(GO) build ./cmd/...

ui-build:
	cd ui/eval-web && $(PNPM) build

runner-smoke:
	cd runner && $(PYTHON) -m rsi_runner.main --once

improvement-cron-once:
	$(GO) run ./cmd/improvement-plane --mode cron --once
