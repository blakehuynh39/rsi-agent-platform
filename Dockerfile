FROM node:22-bookworm AS ui-builder

WORKDIR /src/ui/eval-web
COPY ui/eval-web/package.json ui/eval-web/pnpm-lock.yaml ui/eval-web/index.html ui/eval-web/tsconfig.json ui/eval-web/vite.config.ts ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY ui/eval-web/src ./src
RUN pnpm build

FROM golang:1.24-bookworm AS builder

ARG SERVICE=improvement-plane
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
COPY --from=ui-builder /src/internal/reviewui/dist ./internal/reviewui/dist
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/base-debian12
ARG SERVICE=improvement-plane
ENV RSI_SERVICE_NAME=${SERVICE}
COPY --from=builder /out/service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
