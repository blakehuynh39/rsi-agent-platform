GO ?= go
PNPM ?= corepack pnpm
PYTHON ?= python3

.PHONY: ci test build ui-install ui-build ui-test runner-test runner-smoke improvement-cron-once

ci: ui-test ui-build
	$(GO) test ./...
	$(GO) build ./cmd/...
	PYTHONPATH=runner $(PYTHON) -m unittest discover -s runner/tests

test: ui-build
	$(GO) test ./...

build: ui-build
	$(GO) build ./cmd/...

ui-install:
	cd ui/eval-web && $(PNPM) install --frozen-lockfile

ui-build: ui-install
	cd ui/eval-web && $(PNPM) build

ui-test: ui-install
	cd ui/eval-web && $(PNPM) test

runner-test:
	PYTHONPATH=runner $(PYTHON) -m unittest discover -s runner/tests

runner-smoke:
	cd runner && $(PYTHON) -m rsi_runner.main --once

improvement-cron-once:
	$(GO) run ./cmd/improvement-plane --mode cron --once
