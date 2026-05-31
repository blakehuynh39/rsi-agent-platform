#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="${ROOT_DIR}/internal/db/migrations"
SNAPSHOT="${ROOT_DIR}/internal/db/schema.sql"

tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

find "${MIGRATIONS_DIR}" -maxdepth 1 -type f -name '*.sql' | sort | while read -r file; do
  cat "${file}" >> "${tmp}"
  printf '\n\n' >> "${tmp}"
done

mv "${tmp}" "${SNAPSHOT}"
