#!/usr/bin/env bash
set -euo pipefail

changed_file_list="${1:-changed-files.txt}"

if [ ! -s "$changed_file_list" ]; then
  exit 0
fi

blocked_paths="$(grep -E '^hermes/skills/story-company/' "$changed_file_list" || true)"
if [ -z "$blocked_paths" ]; then
  exit 0
fi

cat >&2 <<'MSG'
Direct changes under hermes/skills/story-company are blocked.

That path is only the reviewed bootstrap seed. The live Story company skill tree
is /var/lib/hermes/skills/story-company on the Hermes executor PVC, and exported
repo snapshots under hermes/exported-skills/stage are for visibility only.

Manually seed or update the live PVC first, then let the exporter publish
visibility snapshots instead of hand-authoring runtime skill changes in this
repo path.
MSG

printf '\nBlocked path(s):\n%s\n' "$blocked_paths" >&2
exit 1
