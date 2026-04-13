GO ?= go
PNPM ?= corepack pnpm
PYTHON ?= python3

.PHONY: ci test build ui-install ui-build ui-test runner-test runner-smoke improvement-cron-once db-migrate refresh-schema-snapshot test-postgres

ci: ui-test ui-build
	$(GO) test ./...
	@if [ -n "$$RSI_TEST_POSTGRES_URL" ]; then \
		$(MAKE) test-postgres; \
	else \
		echo "Skipping Postgres integration tests; set RSI_TEST_POSTGRES_URL to enable them."; \
	fi
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

db-migrate:
	$(GO) run ./cmd/improvement-plane --mode migrate

refresh-schema-snapshot:
	chmod +x ./scripts/refresh_schema_snapshot.sh
	./scripts/refresh_schema_snapshot.sh

test-postgres: ui-build
	RSI_TEST_POSTGRES_URL=$${RSI_TEST_POSTGRES_URL:?RSI_TEST_POSTGRES_URL is required} $(GO) test ./internal/db ./internal/store -run 'Test(Postgres|Migration)'
