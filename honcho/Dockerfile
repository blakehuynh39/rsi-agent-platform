# syntax=docker/dockerfile:1.7
FROM python:3.13-slim-bookworm

ARG HONCHO_SOURCE_REPO=https://github.com/blakehuynh39/honcho.git
ARG HONCHO_SOURCE_BRANCH=main
ARG HONCHO_SOURCE_COMMIT=7ade732e583627864066093a6ac18952c1fe1239
ARG HONCHO_UPSTREAM_REPO=https://github.com/plastic-labs/honcho.git

LABEL org.opencontainers.image.title="rsi-agent-platform-honcho" \
      org.opencontainers.image.description="Pinned Honcho fork build used for RSI stage memory services." \
      org.opencontainers.image.source="https://github.com/piplabs/rsi-agent-platform" \
      io.storyprotocol.rsi.honcho.source_repo="${HONCHO_SOURCE_REPO}" \
      io.storyprotocol.rsi.honcho.source_branch="${HONCHO_SOURCE_BRANCH}" \
      io.storyprotocol.rsi.honcho.source_commit="${HONCHO_SOURCE_COMMIT}" \
      io.storyprotocol.rsi.honcho.upstream_repo="${HONCHO_UPSTREAM_REPO}" \
      io.storyprotocol.rsi.honcho.fork_reason="RSI currently pins a user fork until the maintained Honcho delta moves into a piplabs-owned fork."

RUN apt-get update && \
    apt-get install -y --no-install-recommends git ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=ghcr.io/astral-sh/uv:0.9.24 /uv /bin/uv

WORKDIR /tmp

ENV UV_COMPILE_BYTECODE=1
ENV UV_LINK_MODE=copy
ENV UV_CACHE_DIR=/tmp/uv-cache
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV HOME=/app

# Clone the pinned fork branch, then lock the build to the exact reviewed commit.
RUN git clone --branch "${HONCHO_SOURCE_BRANCH}" "${HONCHO_SOURCE_REPO}" /tmp/honcho-src && \
    cd /tmp/honcho-src && \
    git checkout "${HONCHO_SOURCE_COMMIT}" && \
    mv /tmp/honcho-src /app

WORKDIR /app

COPY honcho/rsi-entrypoint.sh /app/docker/rsi-entrypoint.sh

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-group dev && \
    chmod +x /app/docker/rsi-entrypoint.sh && \
    addgroup --system app && \
    adduser --system --group app && \
    mkdir -p /tmp/uv-cache && \
    chown -R app:app /app /tmp/uv-cache

ENV PATH="/app/.venv/bin:$PATH"

USER app

EXPOSE 8000

CMD ["/app/docker/rsi-entrypoint.sh", "api"]
