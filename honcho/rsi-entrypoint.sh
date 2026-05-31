#!/bin/sh
set -eu

normalize_db_uri() {
  case "$1" in
    postgresql+psycopg://*)
      printf '%s' "$1"
      ;;
    postgres://*)
      printf 'postgresql+psycopg://%s' "${1#postgres://}"
      ;;
    postgresql://*)
      printf 'postgresql+psycopg://%s' "${1#postgresql://}"
      ;;
    *)
      printf '%s' "$1"
      ;;
  esac
}

raw_db_uri="${DB_CONNECTION_URI:-${RSI_POSTGRES_URL:-}}"
if [ -n "$raw_db_uri" ]; then
  export DB_CONNECTION_URI="$(normalize_db_uri "$raw_db_uri")"
fi

if [ -z "${DB_SCHEMA:-}" ]; then
  export DB_SCHEMA="honcho"
fi

mode="${1:-api}"
case "$mode" in
  api)
    exec /app/.venv/bin/fastapi run --host 0.0.0.0 src/main.py
    ;;
  deriver)
    exec /app/.venv/bin/python -m src.deriver
    ;;
  *)
    exec "$@"
    ;;
esac
