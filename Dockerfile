ARG GO_VERSION=1.26.2

FROM node:22-bookworm AS ui-builder

WORKDIR /src/ui/eval-web
COPY ui/eval-web/package.json ui/eval-web/pnpm-lock.yaml ui/eval-web/index.html ui/eval-web/tsconfig.json ui/eval-web/tsconfig.app.json ui/eval-web/tsconfig.node.json ui/eval-web/vite.config.ts ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY ui/eval-web/public ./public
COPY ui/eval-web/src ./src
RUN pnpm build

FROM golang:${GO_VERSION}-bookworm AS builder

ARG SERVICE=improvement-plane
ARG CGO_ENABLED=0
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
COPY --from=ui-builder /src/internal/reviewui/dist ./internal/reviewui/dist
RUN CGO_ENABLED=${CGO_ENABLED} go build -o /out/service ./cmd/${SERVICE}

FROM debian:bookworm-slim AS sentry-cli
ARG SENTRY_CLI_VERSION=0.33.0
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl gzip \
    && rm -rf /var/lib/apt/lists/*
RUN SENTRY_CLI_NO_TELEMETRY=1 SENTRY_INSTALL_DIR=/usr/local/bin SENTRY_VERSION=${SENTRY_CLI_VERSION} \
    bash -c 'curl -fsSL https://cli.sentry.dev/install | bash -s -- --no-modify-path --no-completions --no-agent-skills'

FROM gcr.io/distroless/base-debian12
ARG SERVICE=improvement-plane
ENV RSI_SERVICE_NAME=${SERVICE}
COPY --from=builder /out/service /service
COPY --from=sentry-cli /usr/local/bin/sentry /usr/local/bin/sentry
EXPOSE 8080
ENTRYPOINT ["/service"]
