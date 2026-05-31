---
name: destructive-migration-safety
description: "Two-phase deploy rule for destructive DB migrations (DROP COLUMN/TABLE/CONSTRAINT). Code removal first, column drop second. Includes CI workflow pattern and agent rule template."
version: 1.0.0
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [Database, Migration, DevOps, Safety, CI/CD, PostgreSQL]
    related_skills: [github-code-review]
---

# Destructive Migration Safety

Prevent production `column "X" does not exist` errors caused by rolling-deploy race conditions between destructive database migrations and old application code still referencing dropped columns.

## The Core Rule

**Destructive migrations (DROP COLUMN, DROP TABLE, DROP CONSTRAINT) must follow a two-phase deploy:**

1. **Phase 1 (PR A):** Deploy application code that **stops reading/writing** the targeted column/table. Merge, build, deploy to production. Verify zero errors.
2. **Phase 2 (PR B):** After Phase 1 is confirmed healthy in production, deploy the migration that **drops** the column/table/constraint.

The two PRs must be **sequential** — never bundled in the same PR or same deploy.

## Why This Matters

Applications with auto-migration on startup (e.g., `sqlx migrate run`) create a rolling-deploy race:

- New pod starts → runs destructive migration → drops column
- Old pods still serve traffic with code referencing that column → `column "X" does not exist` (HTTP 500)
- Error persists until old pods drain (15-60 seconds typical)

By separating code removal (Phase 1) from column removal (Phase 2), no running pod ever references a column that doesn't exist.

## Patterns to Flag

Destructive (require two-phase):
- `DROP COLUMN [IF EXISTS]` — column removal
- `DROP TABLE [IF EXISTS]` — table removal
- `ALTER TABLE ... DROP CONSTRAINT` — constraint removal code depends on
- `DROP INDEX [IF EXISTS]` — index removal (lower risk, flag as warning)

Safe to bundle (additive only):
- `ADD COLUMN` / `ADD CONSTRAINT` / `CREATE TABLE`
- `INSERT` / `UPDATE` / `DELETE` data migrations
- `CREATE INDEX` / `CREATE INDEX CONCURRENTLY`

## Implementation: Check Layers

### 1. Agent Rule (alwaysApply)

Create `.agents/rules/023-destructive-migration-safety.md`:

```markdown
---
description: Destructive database migrations (DROP COLUMN / DROP TABLE) must follow a two-phase deploy sequence
alwaysApply: true
---

# Destructive Migration Safety

## Rule

**Destructive database migrations** — any migration that drops a column, table, constraint, or index that application code may still reference — MUST follow a **two-phase deploy**:

1. **Phase 1 (PR A)**: Deploy application code that **stops reading/writing** the targeted column/table. Merge, build, and deploy to production. Verify no errors.

2. **Phase 2 (PR B)**: After Phase 1 is confirmed healthy in production, deploy the migration that **drops** the column/table/constraint.

The two PRs must be **sequential**, not bundled in the same PR or same deploy.
```

Symlink to `.cursor/rules/` and `.claude/rules/` so all agents see it.

### 2. CI Workflow (non-blocking warning)

Create `.github/workflows/migration-safety-check.yml`:

```yaml
name: Migration Safety Check

on:
  pull_request:
    paths:
      - 'apps/api/migrations/*.sql'

permissions:
  contents: read

jobs:
  check-destructive:
    name: Check for destructive migrations
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Scan migration files for destructive operations
        id: scan
        shell: bash
        run: |
          set -euo pipefail

          DESTRUCTIVE_PATTERNS=(
            "DROP[[:space:]]+COLUMN"
            "DROP[[:space:]]+TABLE"
            "ALTER[[:space:]]+TABLE.*DROP[[:space:]]+CONSTRAINT"
          )

          ISSUES=""
          for sql_file in apps/api/migrations/*.sql; do
            [ -f "$sql_file" ] || continue

            diff_sql="$(git diff "origin/$GITHUB_BASE_REF" -- "$sql_file" 2>/dev/null || true)"
            if [ -z "$diff_sql" ]; then
              if git diff --name-only "origin/$GITHUB_BASE_REF" --diff-filter=A -- "$sql_file" 2>/dev/null | grep -q .; then
                diff_sql="$(cat "$sql_file")"
              else
                continue
              fi
            fi

            for pattern in "${DESTRUCTIVE_PATTERNS[@]}"; do
              if echo "$diff_sql" | grep -qiE "$pattern"; then
                matches="$(echo "$diff_sql" | grep -niE "$pattern" | head -5)"
                filename="$(basename "$sql_file")"
                ISSUES="${ISSUES}"$'\n'"- **${filename}** matches \`${pattern}\`"$'\n'
                ISSUES="${ISSUES}"$'\n''```sql'$'\n'"${matches}"$'\n''```'$'\n'
              fi
            done
          done

          if [ -z "$ISSUES" ]; then
            echo "issues_found=false" >> "$GITHUB_OUTPUT"
            exit 0
          fi

          {
            echo "## Destructive Migration Detected"
            echo ""
            echo "This PR contains database migrations that **drop** columns, tables, or constraints."
            echo ""
            echo "### Two-Phase Deploy Required"
            echo ""
            echo "> **Phase 1**: Deploy application code that stops reading/writing the targeted column/table."
            echo "> **Phase 2**: After Phase 1 is healthy, deploy this migration to drop the column/table."
            echo ""
            echo "**Do not merge this PR unless**:"
            echo "1. The associated application code is already deployed to production, OR"
            echo "2. This PR is split into two: code removal first, then this migration."
            echo ""
            echo "---"
            echo "${ISSUES}"
          } >> "$GITHUB_STEP_SUMMARY"

          echo "issues_found=true" >> "$GITHUB_OUTPUT"

      - name: Warn on destructive migration
        if: steps.scan.outputs.issues_found == 'true'
        shell: bash
        run: |
          echo "::warning title=Destructive Migration Detected::This PR drops columns/tables/constraints. Verify two-phase deploy sequencing."
```

Key design choices:
- **Non-blocking**: Does not prevent merge — reviewers must manually verify
- **PR paths filter**: Only runs when migration SQL files change
- **Base ref diff**: Compares against the PR target branch, not main
- **Step summary**: Writes findings to GitHub Actions job page for visibility

### 3. Conventions Doc

Add to `docs/conventions.md`:

```markdown
## Migration safety

- **Destructive migrations (DROP COLUMN, DROP TABLE) require two-phase deploy**: 
  Deploy code that stops reading the column/table first (Phase 1), verify in production, 
  then deploy the migration in a separate PR (Phase 2).
- Additive migrations (ADD COLUMN, CREATE TABLE) are safe to bundle with code changes.
```

## Real-World Example

**Incident (2026-05-14):** `depin-backend` production outage at `api.numolabs.ai`.

- PR #479 merged staging→main, including commit `92eb40c` which added migration `0085_drop_user_voice_phrase.sql`
- Migration dropped `voice_phrase` column from `users` table
- Old pods (RS `8bfd94fcf`, running commit `11b169b`) still referenced `voice_phrase` in `USER_SELECT_COLUMNS`
- Every SELECT from `users` on old pods returned 500: `column "voice_phrase" does not exist`
- Duration: ~15 seconds (21:55:54–21:56:09 UTC), self-resolved when old pods drained

**Root cause:** Code removal and column drop were bundled in the same commit (`92eb40c`), deployed as a single image. The two-phase rule would have prevented this.

## PR Review Integration

When reviewing a PR via `github-code-review` skill, add this to the checklist:

- [ ] Does this PR change migration SQL files?
- [ ] If yes: does any migration contain DROP COLUMN/TABLE/CONSTRAINT?
- [ ] If destructive: is the associated code removal already deployed to production?
- [ ] If not: flag as blocking — split into two PRs
