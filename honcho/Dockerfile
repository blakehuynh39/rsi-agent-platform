# syntax=docker/dockerfile:1.7
FROM python:3.13-slim-bookworm

ARG HONCHO_REF=v3.0.6

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

RUN git clone --depth 1 --branch "${HONCHO_REF}" https://github.com/plastic-labs/honcho.git /tmp/honcho-src && \
    mv /tmp/honcho-src /app

WORKDIR /app

COPY honcho/apply_rsi_overlay.py /tmp/apply_rsi_overlay.py
COPY honcho/rsi-entrypoint.sh /app/docker/rsi-entrypoint.sh

RUN python /tmp/apply_rsi_overlay.py /app && rm -f /tmp/apply_rsi_overlay.py

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
