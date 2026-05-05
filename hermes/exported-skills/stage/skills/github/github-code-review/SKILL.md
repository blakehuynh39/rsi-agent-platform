---
name: github-code-review
description: "Review PRs: diffs, inline comments via gh or REST."
version: 1.2.1
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [GitHub, Code-Review, Pull-Requests, Git, Quality]
    related_skills: [github-auth, github-pr-workflow]
---

# GitHub Code Review

Perform code reviews on local changes before pushing, or review open PRs on GitHub. Most of this skill uses plain `git` — the `gh`/`curl` split only matters for PR-level interactions.

## Prerequisites

- Authenticated with GitHub (see `github-auth` skill)
- Inside a git repository

### Setup (for PR interactions)

```bash
if command -v gh &>/dev/null && gh auth status &>/dev/null; then
  AUTH="gh"
else
  AUTH="git"
  if [ -z "$GITHUB_TOKEN" ]; then
    if [ -f ~/.hermes/.env ] && grep -q "^GITHUB_TOKEN=" ~/.hermes/.env; then
      GITHUB_TOKEN=$(grep "^GITHUB_TOKEN=" ~/.hermes/.env | head -1 | cut -d= -f2 | tr -d '\n\r')
    elif grep -q "github.com" ~/.git-credentials 2>/dev/null; then
      GITHUB_TOKEN=$(grep "github.com" ~/.git-credentials 2>/dev/null | head -1 | sed 's|https://[^:]*:\([^@]*\)@.*|\1|')
    fi
  fi
fi

REMOTE_URL=$(git remote get-url origin)
OWNER_REPO=$(echo "$REMOTE_URL" | sed -E 's|.*github\.com[:/]||; s|\.git$||')
OWNER=$(echo "$OWNER_REPO" | cut -d/ -f1)
REPO=$(echo "$OWNER_REPO" | cut -d/ -f2)
```

---

## 1. Reviewing Local Changes (Pre-Push)

This is pure `git` — works everywhere, no API needed.

### Get the Diff

```bash
# Staged changes (what would be committed)
git diff --staged

# All changes vs main (what a PR would contain)
git diff main...HEAD

# File names only
git diff main...HEAD --name-only

# Stat summary (insertions/deletions per file)
git diff main...HEAD --stat
```

### Review Strategy

1. **Get the big picture first:**

```bash
git diff main...HEAD --stat
git log main..HEAD --oneline
```

2. **Review file by file** — use `read_file` on changed files for full context, and the diff to see what changed:

```bash
git diff main...HEAD -- src/auth/login.py
```

3. **Check for common issues:**

```bash
# Debug statements, TODOs, console.logs left behind
git diff main...HEAD | grep -n "print(\|console\.log\|TODO\|FIXME\|HACK\|XXX\|debugger"

# Large files accidentally staged
git diff main...HEAD --stat | sort -t'|' -k2 -rn | head -10

# Secrets or credential patterns
git diff main...HEAD | grep -in "password\|secret\|api_key\|token.*=\|private_key"

# Merge conflict markers
git diff main...HEAD | grep -n "<<<<<<\|>>>>>>\|======="
```

4. **Present structured feedback** to the user.

### Review Output Format

When reviewing local changes, present findings in this structure:

```
## Code Review Summary

### Critical
- **src/auth.py:45** — SQL injection: user input passed directly to query.
  Suggestion: Use parameterized queries.

### Warnings
- **src/models/user.py:23** — Password stored in plaintext. Use bcrypt or argon2.
- **src/api/routes.py:112** — No rate limiting on login endpoint.

### Suggestions
- **src/utils/helpers.py:8** — Duplicates logic in `src/core/utils.py:34`. Consolidate.
- **tests/test_auth.py** — Missing edge case: expired token test.

### Looks Good
- Clean separation of concerns in the middleware layer
- Good test coverage for the happy path
```

---

## 2. Reviewing a Pull Request on GitHub

### View PR Details

**With gh:**

```bash
gh pr view 123
gh pr diff 123
gh pr diff 123 --name-only
```

**With git + curl:**

```bash
PR_NUMBER=123

# Get PR details
curl -s \
  -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$OWNER/$REPO/pulls/$PR_NUMBER \
  | python3 -c "
import sys, json
pr = json.load(sys.stdin)
print(f\"Title: {pr['title']}\")
print(f\"Author: {pr['user']['login']}\")
print(f\"Branch: {pr['head']['ref']} -> {pr['base']['ref']}\")
print(f\"State: {pr['state']}\")
print(f\"Body:\n{pr['body']}\")"

# List changed files
curl -s \
  -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$OWNER/$REPO/pulls/$PR_NUMBER/files \
  | python3 -c "
import sys, json
for f in json.load(sys.stdin):
    print(f\"{f['status']:10} +{f['additions']:-4} -{f['deletions']:-4}  {f['filename']}\")"
```

### Check Out PR Locally for Full Review

This works with plain `git` — no `gh` needed:

```bash
# Fetch the PR branch and check it out
git fetch origin pull/123/head:pr-123
git checkout pr-123

# Now you can use read_file, search_files, run tests, etc.

# View diff against the base branch
git diff main...pr-123
```

**With gh (shortcut):**

```bash
gh pr checkout 123
```

### Leave Comments on a PR

**General PR comment — with gh:**

```bash
gh pr comment 123 --body "Overall looks good, a few suggestions below."
```

**General PR comment — with curl:**

```bash
curl -s -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$OWNER/$REPO/issues/$PR_NUMBER/comments \
  -d '{"body": "Overall looks good, a few suggestions below."}'
```

### Leave Inline Review Comments

**Single inline comment — with gh (via API):**

```bash
HEAD_SHA=$(gh pr view 123 --json headRefOid --jq '.headRefOid')

gh api repos/$OWNER/$REPO/pulls/123/comments \
  --method POST \
  -f body="This could be simplified with a list comprehension." \
  -f path="src/auth/login.py" \
  -f commit_id="$HEAD_SHA" \
  -f line=45 \
  -f side="RIGHT"
```

**Single inline comment — with curl:**

```bash
# Get the head commit SHA
HEAD_SHA=$(curl -s \
  -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$OWNER/$REPO/pulls/$PR_NUMBER \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['head']['sha'])")

curl -s -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$OWNER/$REPO/pulls/$PR_NUMBER/comments \
  -d "{
    \"body\": \"This could be simplified with a list comprehension.\",
    \"path\": \"src/auth/login.py\",
    \"commit_id\": \"$HEAD_SHA\",
    \"line\": 45,
    \"side\": \"RIGHT\"
  }"
```

### Submit a Formal Review (Approve / Request Changes)

**With gh:**

```bash
gh pr review 123 --approve --body "LGTM!"
gh pr review 123 --request-changes --body "See inline comments."
gh pr review 123 --comment --body "Some suggestions, nothing blocking."
```

**With curl — multi-comment review submitted atomically:**

```bash
HEAD_SHA=$(curl -s \
  -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$OWNER/$REPO/pulls/$PR_NUMBER \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['head']['sha'])")

curl -s -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$OWNER/$REPO/pulls/$PR_NUMBER/reviews \
  -d "{
    \"commit_id\": \"$HEAD_SHA\",
    \"event\": \"COMMENT\",
    \"body\": \"Code review from Hermes Agent\",
    \"comments\": [
      {\"path\": \"src/auth.py\", \"line\": 45, \"body\": \"Use parameterized queries to prevent SQL injection.\"},
      {\"path\": \"src/models/user.py\", \"line\": 23, \"body\": \"Hash passwords with bcrypt before storing.\"},
      {\"path\": \"tests/test_auth.py\", \"line\": 1, \"body\": \"Add test for expired token edge case.\"}
    ]
  }"
```

Event values: `"APPROVE"`, `"REQUEST_CHANGES"`, `"COMMENT"`

The `line` field refers to the line number in the *new* version of the file. For deleted lines, use `"side": "LEFT"`.

---

## 2b. Cross-Repo Paired PR Review

When a feature spans two repos (e.g., backend API + frontend app), the user often posts both PRs in a single review request. Review them together to catch cross-repo misalignment.

### Detection

- Two PR links in the same Slack message / review request
- PR descriptions reference each other (e.g., "Pairs with org/repo#NNN")
- Branch naming convention matches (e.g., `feat/castle-sub-scores` in both repos)

### Workflow

1. **Clone both repos** in parallel to `/tmp/<repo>-review/`
2. **Fetch both PRs** and check [branch/CI/mergeable] status on each
3. **Check cross-repo alignment:**
   - **Route naming:** Backend route path matches what the FE API client calls. If BE uses `/admin/users-safety-distributions`, FE must call exactly that path.
   - **Field names:** New fields on BE response schemas must appear on FE types/interfaces with matching names.
   - **Null handling:** If BE sends `null` for new fields on historical rows, the FE must render a fallback (e.g., `\u2014` em-dash) — verify with `!= null` not falsy `||`.
4. **Review each PR independently** following Steps 1-7, then check alignment
5. **Submit approvals together** — post the BE review first, then the FE, then post a summary in the original request thread

### Cross-Repo Checklist

- [ ] Route paths match between BE route definition and FE API client
- [ ] New response fields present on both sides with identical naming
- [ ] FE handles `null`/missing new fields for historical rows
- [ ] Both PRs have matching base branches (BE: `staging`, FE: `develop`)
- [ ] CI green on both repos
- [ ] BE PR description links to FE PR and vice versa

### Cross-Repo Contract: depin-backend ↔ numo-monorepo

When reviewing a `depin-backend` PR (base: `staging`), the accompanying FE repo is **always** `piplabs/numo-monorepo` (base: `develop`). Any BE change that affects the FE flow **must** be accompanied by a matching FE PR, or the review must flag the gap.

**Changes that REQUIRE a linked FE PR:**
- New or renamed API routes/paths → FE API client must add the matching call
- New fields on response types (`AdminUserSummary`, etc.) → FE types must be updated
- Changed field semantics (e.g., `state` values, enum variants) → FE rendering/switching must handle the new values
- New query parameters on existing endpoints → FE filters/sorting must wire them
- Changed auth/error behavior → FE error handling must surface the new cases

**Changes that do NOT require an FE PR:**
- Internal refactors, performance changes, logging
- Migration-only changes with no API surface impact
- Changes behind a feature flag with no default activation
- Bug fixes that don't alter the contract shape

**Detection during review:**
1. Check the BE PR description for a link to `numo-monorepo` (e.g., "Pairs with piplabs/numo-monorepo#NNN" or "FE: https://github.com/piplabs/numo-monorepo/pull/NNN")
2. Check commits for `references/docs/numo-admin-api.md` or `numo-api-reference.md` changes (spec-first pattern)
3. If the BE PR changes schemas, routes, or response shapes and NO FE PR is linked → **flag as missing cross-repo pair**
4. Search for an open PR on `numo-monorepo` with a matching branch prefix that may have been opened separately

**PITFALL:** Treating a BE PR as self-contained when it adds a new admin endpoint. The admin dashboard lives in `numo-monorepo/apps/admin` — if operators can't reach the new endpoint, the feature is incomplete. Always verify the FE half exists or flag the gap.

### Common Failure Modes

- **Route mismatch**: BE adds a depth-2 hyphenated route (`/admin/users-safety-distributions`) to avoid `{user_id}` routing conflicts, but FE uses a slashed path (`/v1/admin/users/safety-distributions`) — results in 404.
- **Stale types**: BE adds fields to `AdminUserSummary` but FE types aren't updated, causing silent `undefined` values that pass falsy checks and render as `0`.
- **Merged out of order**: FE merges to `develop` but BE hasn't merged to `staging` yet — staging previews break because the new endpoint doesn't exist.
- **Multi-pod cost amplification**: Background job added to BE but reviewer didn't check K8s replica count — the job's DB load is multiplied by N pods. Always verify deployment topology during cross-repo reviews (see Section 3b and `references/depin-backend-deployment-topology.md`).

---

## 3. Review Checklist

When performing a code review (local or PR), systematically check:

### Correctness
- Does the code do what it claims?
- Edge cases handled (empty inputs, nulls, large data, concurrent access)?
- Error paths handled gracefully?

### Security
- No hardcoded secrets, credentials, or API keys
- Input validation on user-facing inputs
- No SQL injection, XSS, or path traversal
- Auth/authz checks where needed

### Code Quality
- Clear naming (variables, functions, classes)
- No unnecessary complexity or premature abstraction
- DRY — no duplicated logic that should be extracted
- Functions are focused (single responsibility)

### Testing
- New code paths tested?
- Happy path and error cases covered?
- Tests readable and maintainable?

### Performance
- No N+1 queries or unnecessary loops
- Appropriate caching where beneficial
- No blocking operations in async code paths

### Documentation
- Public APIs documented
- Non-obvious logic has comments explaining "why"
- README updated if behavior changed

### Multi-Pod / Distributed Deployment (for services with >1 replica)

When reviewing PRs for services deployed with multiple replicas (e.g., Kubernetes Deployments with `replicas > 1` backed by a single shared database), every change must be evaluated through the lens of **N pods sharing one durable source of truth**. This is NOT just a single-repo code review — the deployment topology is part of the review surface.

**Always check:**
1. **How many replicas run this service?** — Verify with `kubectl get deployments -n <ns>` before starting the review. Note the replica count and whether it's static or HPA-driven.
2. **Background jobs / cron-like work:** If the PR introduces a background task (tokio spawn, timer loop, cronjob), does every pod run it independently? If so:
   - Is the task **idempotent**? (e.g., `INSERT ... ON CONFLICT DO UPDATE`, `UPDATE ... WHERE ... AND state != 'done'`)
   - Is the task **cheap enough** to run N times? (a `DELETE` of 3 stale rows is fine; a full-table scan with window functions is not)
   - Should only **one pod** execute it? (use `pg_try_advisory_lock`, a leader-election table, or a separate singleton CronJob deployment)
3. **In-memory state:** Caches, rate-limit buckets, leader-election tokens — these live per-pod and are invisible to peers. If the PR touches them, verify they don't create split-brain scenarios.
4. **Migrations:** `CREATE INDEX CONCURRENTLY` is critical in multi-pod deploys (a blocking index build locks writes for ALL pods). `IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS` prevents crash-loops on pod restarts.
5. **Startup coordination:** If all N pods restart simultaneously, do they all fire expensive init work at once? (e.g., cache warm-up, table scans). Consider `pg_try_advisory_lock` or jittered startup delays.

**For `depin-backend` specifically** (see `references/depin-backend-deployment-topology.md`):
- 2+ replicas in `story` namespace, all sharing one Postgres instance
- Existing background jobs (`multiplier_sweep`, `hot_path_cache`, `idempotency_cleanup`) follow the "every pod runs independently" pattern with idempotent SQL
- New background jobs should match this convention or document why they deviate
- Always check K8s deployment state during review — replica count may have changed since last review

**PITFALL:** Reviewing code as if it runs in a single process. A `refresh_all()` that scans the full users table is fine in local dev but runs N× concurrently in production. Always multiply the cost by the replica count when assessing DB load.

---

## 4. Pre-Push Review Workflow

When the user asks you to "review the code" or "check before pushing":

1. `git diff main...HEAD --stat` — see scope of changes
2. `git diff main...HEAD` — read the full diff
3. For each changed file, use `read_file` if you need more context
4. Apply the checklist above
5. Present findings in the structured format (Critical / Warnings / Suggestions / Looks Good)
6. If critical issues found, offer to fix them before the user pushes

---

## 5. PR Review Workflow (End-to-End)

When the user asks you to "review PR #N", "look at this PR", or gives you a PR URL, follow this recipe:

### Step 1: Set up environment

```bash
source "${HERMES_HOME:-$HOME/.hermes}/skills/github/github-auth/scripts/gh-env.sh"
# Or run the inline setup block from the top of this skill
```

### Step 2: Gather PR context

Get the PR metadata, description, and list of changed files to understand scope before diving into code.

**With gh:**
```bash
gh pr view 123
gh pr diff 123 --name-only
gh pr checks 123
```

**With curl:**
```bash
PR_NUMBER=123
```bash
# PR details (title, author, description, branch)
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$GH_OWNER/$GH_REPO/pulls/$PR_NUMBER

# Changed files with line counts
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$GH_OWNER/$GH_REPO/pulls/$PR_NUMBER/files
```

**PITFALL:** `gh pr diff --stat` is NOT a valid flag. `gh pr diff` supports `--name-only`, `--patch`, `--color`, and `--web`. For a stat summary (lines changed per file), use the REST API `/pulls/$PR_NUMBER/files` endpoint (see curl example above), or `git diff --stat` after checking out the PR branch locally.

### Step 3: Check out the PR locally

This gives you full access to `read_file`, `search_files`, and the ability to run tests.

```bash
git fetch origin pull/$PR_NUMBER/head:pr-$PR_NUMBER
git checkout pr-$PR_NUMBER
```

### Step 4: Read the diff and understand changes

```bash
# Full diff against the base branch
git diff main...HEAD

# Or file-by-file for large PRs
git diff main...HEAD --name-only
# Then for each file:
git diff main...HEAD -- path/to/file.py
```

For each changed file, use `read_file` to see full context around the changes — diffs alone can miss issues visible only with surrounding code.

### Step 5: Run automated checks locally (if applicable)

```bash
# Run tests if there's a test suite
python -m pytest 2>&1 | tail -20
# or: npm test, cargo test, go test ./..., etc.

# Run linter if configured
ruff check . 2>&1 | head -30
# or: eslint, clippy, etc.
```

### Step 6: Apply the review checklist (Section 3)

Go through each category: Correctness, Security, Code Quality, Testing, Performance, Documentation.

### Step 7: Post the review to GitHub

Collect your findings and submit them as a formal review with inline comments.

**With gh:**
```bash
gh pr review $PR_NUMBER --approve --body "Reviewed by Hermes Agent. Code looks clean — good test coverage, no security concerns."
gh pr review $PR_NUMBER --request-changes --body "Found a few issues — see inline comments."
gh pr review $PR_NUMBER --comment --body "Some suggestions, nothing blocking."
```

**PITFALL:** When review bodies contain backticks, markdown tables, shell-significant characters (`$`, `!`, `"`), or multi-line content, `--body "..."` can silently fail or produce garbled output. Use `--body-file` instead — write the content to a temp file first, then:

```bash
# Write review body to a temp file to avoid shell quoting issues
cat > /tmp/review-body.md << 'REVIEW_EOF'
## Code Review

| Area | Result |
|---|---|
| CI | pass |
REVIEW_EOF

gh pr review $PR_NUMBER --approve --body-file /tmp/review-body.md
```

**With curl — atomic review with multiple inline comments:**
```bash
HEAD_SHA=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$GH_OWNER/$GH_REPO/pulls/$PR_NUMBER \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['head']['sha'])")

# Build the review JSON — event is APPROVE, REQUEST_CHANGES, or COMMENT
curl -s -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$GH_OWNER/$GH_REPO/pulls/$PR_NUMBER/reviews \
  -d "{
    \"commit_id\": \"$HEAD_SHA\",
    \"event\": \"REQUEST_CHANGES\",
    \"body\": \"## Hermes Agent Review\n\nFound 2 issues, 1 suggestion. See inline comments.\",
    \"comments\": [
      {\"path\": \"src/auth.py\", \"line\": 45, \"body\": \"🔴 **Critical:** User input passed directly to SQL query — use parameterized queries.\"},
      {\"path\": \"src/models.py\", \"line\": 23, \"body\": \"⚠️ **Warning:** Password stored without hashing.\"},
      {\"path\": \"src/utils.py\", \"line\": 8, \"body\": \"💡 **Suggestion:** This duplicates logic in core/utils.py:34.\"}
    ]
  }"
```

### Step 8: Also post a summary comment

In addition to inline comments, leave a top-level summary so the PR author gets the full picture at a glance. Use the review output format from `references/review-output-template.md`.

**With gh:**
```bash
gh pr comment $PR_NUMBER --body "$(cat <<'EOF'
## Code Review Summary

**Verdict: Changes Requested** (2 issues, 1 suggestion)

### 🔴 Critical
- **src/auth.py:45** — SQL injection vulnerability

### ⚠️ Warnings
- **src/models.py:23** — Plaintext password storage

### 💡 Suggestions
- **src/utils.py:8** — Duplicated logic, consider consolidating

### ✅ Looks Good
- Clean API design
- Good error handling in the middleware layer

---
*Reviewed by Hermes Agent*
EOF
)"
```

### Step 9: Clean up

```bash
git checkout main
git branch -D pr-$PR_NUMBER
```

### Decision: Approve vs Request Changes vs Comment

- **Approve** — no critical or warning-level issues, only minor suggestions or all clear
- **Request Changes** — any critical or warning-level issue that should be fixed before merge
- **Comment** — observations and suggestions, but nothing blocking (use when you're unsure or the PR is a draft)
