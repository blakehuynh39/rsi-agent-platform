---
name: github-code-review
description: "Review PRs: diffs, inline comments via gh or REST. Multi-angle thermo-nuclear review with idempotency, N+1, architecture, and deep correctness checks. Fresh subagent per review pass for anti-bias."
version: 2.3.2
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [GitHub, Code-Review, Pull-Requests, Git, Quality, Idempotency, N+1, Deep-Correctness, Architecture, Thermo-Nuclear, Anti-Bias]
    related_skills: [github-auth, github-pr-workflow, destructive-migration-safety]
---

# GitHub Code Review

## Mandatory Fresh-Subagent Rule

Every PR review and every PR re-review must be performed by a fresh subagent for that review pass.

The parent agent is an orchestrator only: gather raw inputs, call `delegate_task`, verify and format the result, then deliver it. The parent must not directly review a PR or re-review from its own accumulated context.

This applies to first reviews, "review this PR", "re-review", "check the fixes", "can we approve now?", and any Slack follow-up that asks for a new verdict on the PR. If fresh subagent delegation is unavailable, fails, or returns unusable output, report the review as blocked or partial instead of approving, requesting changes, or posting a complete review.

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

**PITFALL:** `gh pr diff N -- path/to/file.rs` does NOT work — `gh pr diff` does not accept `--` for file path filtering (unlike `git diff`). The error is `accepts at most 1 arg(s), received 2`. To inspect only specific files from a PR, either clone the repo and use `git diff main...HEAD -- path/to/file` or pipe `gh pr diff N | sed -n '/^diff --git a\/path\/to\/file/,/^diff /{p}'`. For targeted cross-repo checks on specific files, just clone and use `git diff` — it's cleaner.

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
- [ ] **Merge order**: BE merges to `staging` before FE merges to `develop`. If FE is already merged but BE is still open, flag it. See `references/cross-repo-check-trace.md` for examples.

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

**PITFALL:** When a PR description links to a deployment/infra counterpart PR (e.g., story-deployments) that the bot cannot access (repo 404, private repo, missing token scope), do NOT block the review on it. Review the accessible PRs independently, note the inaccessible PR as a deployment dependency, and proceed with approvals if the code is clean. The inaccessible PR may be a private repo that needs manual verification — flag it for the requester rather than stalling the review chain.

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

**Deep Correctness — go beyond surface-level checks:**

- **Race conditions:** In concurrent code (async Rust, Go goroutines, JS Promise chains, multi-threaded Java), check: shared mutable state protected by locks/channels/atomics? Lock ordering consistent to prevent deadlocks? `SELECT ... FOR UPDATE` used where read-then-write must be atomic?
- **Transaction boundaries:** Database operations that must be atomic — are they wrapped in a transaction? If multiple rows in different tables must be updated together, is there a single `BEGIN...COMMIT`? Watch for `COMMIT` inside loops (should be batched).
- **Error propagation:** Are errors from library calls or sub-functions actually checked and propagated? In Rust: no `unwrap()` or `expect()` in production code without justification. In Go: no bare `_` discarding errors. In TypeScript: `.catch()` blocks that just `console.error` without recovery.
- **Partial failure handling:** If step A succeeds and step B fails, what state is left behind? Is cleanup (rollback, compensating action) needed? Look for patterns like "update user, then call Stripe" — if Stripe fails, the user record is inconsistent.
- **Ordering dependencies:** Does the code assume a particular execution order that isn't guaranteed? (e.g., async tasks spawned but not awaited before reading results, event handler registration after event might have fired).
- **Off-by-one / boundary errors:** Check loop bounds: `<` vs `<=`, zero-index confusion, empty collections, single-element edge case, integer overflow at MAX/MIN values.
- **Time & timezone correctness:** `SystemTime::now()` vs monotonic clocks for durations. UTC storage vs local display. Daylight saving boundary handling. Leap seconds.

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

### Architecture Review

Pull back from line-level concerns and assess the *shape* of the change. This is the "thermo nuclear" layer — structural issues that cause death by a thousand cuts.

- **Boundary drift:** Where did the seam between layers move? Did UI code start reaching into the database? Did domain types start importing transport types? Check: `import` statements that cross established layer boundaries.
- **Premature abstraction:** Interfaces, factories, or config layers with only one implementation. These are debt — they make the code harder to understand without providing flexibility. Flag any abstraction that serves only one concrete case.
- **Coupling:** New utilities importing from feature modules. Shared mutable state being introduced. A change in module A forcing a change in module B when they should be independent. Look for bidirectional imports between modules.
- **Scalability:** If this code path goes 10x in volume, what breaks first? Are there unbounded collections? Missing pagination? Single-threaded bottlenecks?
- **Reversibility:** If this turns out wrong in a month, how hard is the rollback? One-way doors (database migrations, API contract changes, data format changes) should be called out explicitly.
- **Naming at the architecture level:** Types/functions named for the implementation (`UserManagerImplV2`) rather than the role (`UserDirectory`). Names should reflect what, not how.

### Testing
- New code paths tested?
- Happy path and error cases covered?
- Tests readable and maintainable?

**Deep Test Quality — a test that exists ≠ a test that catches regressions:**

- **Assertion strength:** Are assertions specific enough to catch bugs? `assert(result != null)` passes for wrong results. `assert(result.count === 3)` catches actual deviations. Flag tests that only check truthiness, existence, or "doesn't crash."
- **Mocking discipline:** Mocks that don't fail when the real interface changes are worse than no tests — they create false confidence. Check: does the mock verify the same contract the real implementation provides? Over-mocking (mocking everything except the function under test) produces tests that only test the mocks.
- **Determinism:** Date/time/random/network calls must be stubbed. A test that passes today and fails tomorrow (or passes on one machine but flakes on CI) is broken. Flag `new Date()`, `Math.random()`, `Date.now()`, `uuid()` in test code without seeding/stubbing.
- **Edge cases tested:** Empty input, null/undefined, boundary values (max int, zero, negative), errors thrown by dependencies. One happy-path test is not coverage.
- **Test names:** Names that describe behavior, not implementation. `it('returns 403 when user lacks admin role')` beats `it('test auth middleware #3')`. If you can't tell what failed from the test name alone, flag it.
- **Snapshot testing discipline:** Snapshots that are too large (entire page renders, full API responses) are brittle. Snapshots without human review (auto-accepted on update) are useless. Flag snapshot files over ~50 lines that weren't explicitly reviewed.

### Performance
- No N+1 queries or unnecessary loops
- Appropriate caching where beneficial
- No blocking operations in async code paths

**N+1 Query Detection — the silent performance killer:**

The N+1 problem: executing 1 query to fetch N records, then N additional queries to fetch related data for each record. Total: N+1 queries instead of 2.

**Patterns to flag (by language/ORM):**

- **Rust (SQLx / Diesel):** `.fetch()` or `query_as()` called inside a `for` loop that iterates over results of a prior query. Look for `for row in rows { ... sqlx::query_as!(...) ... }`.
- **Rust (SeaORM):** `find_related()` without `.find_with_related()` — loading related entities one at a time.
- **TypeScript/JS (Prisma):** `prisma.user.findMany()` without `include:` — then iterating and calling `prisma.post.findMany({ where: { authorId: user.id } })` per user.
- **TypeScript/JS (Drizzle):** Querying inside `.map()` or `for...of` without using Drizzle's relational queries or `inArray`.
- **TypeScript/JS (Knex/Kysely):** `for (const item of items) { await db.select().from('related').where('id', item.relId) }` — use `WHERE IN` with a single query instead.
- **Java (JPA/Hibernate):** `@OneToMany(fetch = FetchType.LAZY)` accessed in a loop without `JOIN FETCH` or `@EntityGraph`. Default lazy loading triggers N additional SELECTs.
- **Python (SQLAlchemy):** `relationship(lazy='select')` or accessing `.children` on each parent in a loop without `joinedload()` or `selectinload()`.
- **Go (sqlx / GORM):** `db.Select(&items, "SELECT * FROM parent")` followed by `for _, item := range items { db.Get(&child, "SELECT * FROM child WHERE parent_id = $1", item.ID) }`.

**Detection workflow during review:**
1. Scan for loops (`for`, `.map()`, `.forEach()`, `for...of`, `while`) that contain database calls.
2. Check if the loop iterates over results from a previous database query.
3. If yes → it's likely N+1. Verify: count the queries. If one query returns N rows and a second query is executed N times inside the loop, flag it.
4. Suggested fix: eager loading (`include`, `JOIN`, `prefetch_related`, `selectinload`), batch loading (`WHERE IN` with collected IDs), or DataLoader pattern for GraphQL.

**PITFALL:** N+1 queries often don't show in local dev with small datasets. A 10ms query run 100 times = 1 second — invisible locally. But 10,000 records = 100 seconds in production. Always flag N+1 patterns even if the dataset is currently small.

**PITFALL:** ORM "magic" hides queries. In JPA, `entity.getChildren().size()` triggers a SELECT you never wrote. In Django, accessing `parent.child_set.all()` in a template loop fires N queries. Look at the rendered SQL, not just the application code.

### Documentation
- Public APIs documented
- Non-obvious logic has comments explaining "why"
- README updated if behavior changed

### Idempotency & Safe Retries

Idempotency is the property that an operation can be applied multiple times without changing the result beyond the first application. This is critical in distributed systems where network failures, timeouts, and retries are inevitable.

**API Endpoints:**

- **GET, HEAD, PUT, DELETE (not PATCH):** Must be idempotent by HTTP spec. HTTP GET must never have side effects. HTTP PUT must produce the same result whether called once or N times.
- **POST (non-idempotent by default):** Check if the endpoint needs idempotency keys. For payment endpoints, order creation, or any mutation where a retry would duplicate the effect — require an `Idempotency-Key` header with server-side deduplication (store key → result mapping, return cached result on replay).
- **PATCH:** Generally not idempotent. Flag partial updates that don't use optimistic locking (e.g., `UPDATE ... WHERE version = $1`).

**Database Operations:**

- **INSERT:** `INSERT INTO ... ON CONFLICT DO NOTHING` or `ON CONFLICT DO UPDATE` for upserts. Plain `INSERT` without conflict handling will fail on retry or produce duplicates.
- **DELETE:** `DELETE FROM ... WHERE id = $1` is idempotent (deleting a deleted row is a no-op). But `DELETE FROM ... ORDER BY ... LIMIT N` is NOT idempotent — flag for review.
- **UPDATE:** `UPDATE ... SET counter = counter + 1` is NOT idempotent — each retry increments again. Use `UPDATE ... SET counter = $new_value WHERE version = $old_version` with optimistic locking, or `INSERT ... ON CONFLICT` for upserts.
- **Idempotency key pattern:** For multi-step operations, store the idempotency key in a transaction alongside the operation. If the key already exists (previous attempt), return the stored result without re-executing.

**Background Jobs:**

- **Every-pod jobs (depin-backend pattern):** Each pod runs the same job independently. Every operation in the job MUST be idempotent. Use `INSERT ... ON CONFLICT DO UPDATE`, `UPDATE ... WHERE state != 'done'`, or `pg_try_advisory_lock` to prevent duplicate work.
- **Scheduled tasks:** If a cron-like task can overlap with itself (slow run exceeds interval), ensure it uses a lock or checks for an existing run before starting.
- **Event-driven handlers:** If the handler might receive duplicate events (at-least-once delivery), it must be idempotent. Store processed event IDs and skip replays.

**External API Calls:**

- **Stripe / payment providers:** Use idempotency keys for payment creation. A network timeout doesn't mean the payment wasn't processed — retrying without a key can double-charge.
- **Email / notifications:** Flag code that sends emails inside loops or retry logic without deduplication. A retry storm can send hundreds of duplicate emails.
- **Webhook delivery:** Outgoing webhooks should include an idempotency key so the receiver can deduplicate.

**Detection workflow during review:**
1. Identify every mutation operation (INSERT, UPDATE, DELETE, POST, external API call).
2. For each: ask "what happens if this runs twice?" If the answer is "duplicate/bad state" → flag it.
3. Check: does the surrounding code have retry logic? If yes, idempotency is mandatory for every operation inside the retry loop.
4. For REST endpoints: check HTTP method semantics. POST that reads like PUT (e.g., "create-or-update") is a design smell — use PUT instead.
5. For depin-backend specifically: every background job must pass the "2-pod test" — if two pods run this simultaneously, the result must be correct.

**PITFALL:** "We'll never retry this" is always wrong in production. Network blips, deployment rollovers, and DB connection pool exhaustion all cause retries. Assume every operation will be retried at least once.

**PITFALL:** Idempotency keys stored without TTL/cleanup grow unbounded. Flag idempotency key tables that lack a `created_at` column and a cleanup job (see `idempotency_cleanup` in depin-backend for reference pattern).

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

## 3c. Severity Classification

Every finding must carry a severity label. This helps the author triage. **All severities must be addressed** — either fixed or pushback-commented (see decision rule below).

| Label | Meaning | Must Be Addressed? |
|---|---|---|
| 🔴 **CRITICAL** | Security vulnerability, data loss/corruption, crash, double-charge, PII leak | YES — fix or pushback comment |
| 🟠 **HIGH** | Core logic bug, data integrity risk, missing cross-repo pair, N+1 on hot path, non-idempotent mutation on retry path | YES — fix or pushback comment |
| 🟡 **MEDIUM** | Performance concern on non-hot path, missing error handling, incomplete test coverage, unclear naming that could cause bugs | YES — fix or pushback comment |
| 🔵 **LOW** | Style inconsistency, minor DRY opportunity, outdated comment, suggestion for cleaner approach | YES — fix or pushback comment |
| 💡 **SUGGESTION** | Optional improvement worth considering — alternative approach, future-proofing, educational note | YES — fix or pushback comment |

**Decision rule (HARD):**
- Any unresolved finding (any severity) → **REQUEST_CHANGES** — the author must either fix the issue or add a comment explaining why a pushback is necessary
- All findings resolved (fixed or pushback-commented) → **APPROVE**
- All clear with no findings → **APPROVE**

**Pushback comments** are valid resolutions: if the author believes a finding is not worth fixing (e.g., intentional design choice, out of scope, would introduce more risk), they must add a comment on the finding explaining their reasoning. Once pushback-commented, that finding is considered resolved for review purposes. Do not re-flag resolved findings on re-review unless the pushback rationale is demonstrably wrong.

## 3d. Collaborative Review Tone

Take the **question approach** — frame issues as inquiries rather than commands. This is borrowed from Cursor's and the Claude Code review skill's best practices:

```markdown
❌ "This will fail if the list is empty."
✅ "What happens if `items` is an empty array?"

❌ "You need error handling here."
✅ "How should this behave if the API call fails?"

❌ "Extract this into a function."
✅ "This logic appears in 3 places. Would it make sense to extract it?"

❌ "Use async/await here."
✅ "Suggestion: async/await might make this more readable. What do you think?"
```

**When to be direct (not collaborative):** Security vulnerabilities (SQL injection, hardcoded secrets, PII exposure). These are not suggestions — they are blocking. Use direct language: "This is a SQL injection vulnerability. User input is concatenated into the query string. Use parameterized queries."

## 3e. Output Discipline

Every finding must be **specific and actionable**. Vague feedback is noise.

**MANDATORY for every finding:**
- 📁 **File path** (e.g., `src/auth/login.py`)
- 📍 **Line number** (e.g., line 45)
- 🏷️ **Severity label** (🔴 CRITICAL / 🟠 HIGH / 🟡 MEDIUM / 🔵 LOW / 💡 SUGGESTION)
- 📝 **What's wrong** — one sentence
- ✅ **Suggested fix** — one sentence with code sketch if helpful

**Example:**
```
🔴 CRITICAL | src/auth/login.py:45
User input concatenated into SQL query string. SQL injection risk.
→ Use parameterized query: cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
```

**Anti-patterns to avoid:**
- ❌ "This looks risky" — what specifically is risky? Which line?
- ❌ "Consider refactoring" — what pattern should be used instead?
- ❌ "This could be better" — better how? By what metric?
- ❌ Findings without line citations — "somewhere in auth.ts" is not a finding

## 3f. Multi-Angle Review Approach

For non-trivial PRs (>10 files or >500 lines), perform **separate review passes** for each concern area. Don't try to check everything in one pass — context switching causes missed issues.

**Recommended passes (order matters):**
1. **Security pass** — auth/authz, input validation, injection, secrets, data exposure. Read ONLY with security eyes. Skip style and performance.
2. **Correctness pass** — logic, edge cases, error handling, race conditions, transaction boundaries. Read ONLY for bugs.
3. **Architecture pass** — boundary drift, coupling, naming, reversibility. Pull back from line-level.
4. **Performance pass** — N+1, allocations, async patterns, caching. Read ONLY for performance.
5. **Tests pass** — assertion strength, mocking, determinism, coverage gaps. Read ONLY test files.

This is the "thermo nuclear" approach adapted from Cursor's pr-review prompt: four focused angles, each with dedicated attention, producing a comprehensive review that leaves no stone unturned.

For small PRs (1-10 files, <500 lines), a single combined pass is fine — but still mentally separate security from style.

---

## 4. Pre-Push Review Workflow

When the user asks you to "review the code" or "check before pushing":

1. `git diff main...HEAD --stat` — see scope of changes
2. `git diff main...HEAD` — read the full diff
3. For each changed file, use `read_file` if you need more context
4. Apply the checklist above
5. Present findings in the structured format (Critical / Warnings / Suggestions)
6. If critical issues found, offer to fix them before the user pushes

---

## 5. PR Review Workflow (End-to-End)

When the user asks you to "review PR #N", "look at this PR", or gives you a PR URL, follow this recipe. **Choose the right path based on PR size:**

- **PRs under ~50 files / ~1000 lines**: Use the **Remote-Only fast path** (no local checkout). Most PRs fall here — `gh pr view --json` + `gh pr diff` + `gh pr checks` is enough.
- **PRs over ~50 files / ~10K lines**: Use the **Local Checkout path** (check out branch for `read_file` + `search_files` context), or delegate to parallel subagents (Section 5b).

### Remote-Only Fast Path (no checkout — use for most PRs)

**Step A1: Gather PR metadata in one structured call**

```bash
gh pr view N --json number,title,body,state,author,files,additions,deletions,baseRefName,headRefName,labels,reviews,mergeable
```

Key fields: `number`, `title`, `body`, `state`, `author.login`, `files` (array of `{path, additions, deletions}`), `additions`, `deletions`, `baseRefName`, `headRefName`, `mergeable` ("MERGEABLE"/"CONFLICTING"/"UNKNOWN"), `reviews[].state`.

**Step A2: Check CI status**

```bash
gh pr checks N
```

Look for failures — CI is the fastest signal. If any required check is failing, flag it before starting the code review.

**PITFALL:** `gh pr checks` does NOT support `--json` — it only outputs a human-readable table. When you need structured CI status (e.g., to filter by check name or programmatically check conclusions), use `gh pr view N --json statusCheckRollup --jq '...'` instead.

**Step A3: Read the full diff**

```bash
gh pr diff N
```

For large diffs, pipe through `head` or `grep` first to scan structure (e.g., `| grep "^diff \|^@@\|^+\|^-" | head -200`), then read the full diff.

**Step A4: Apply the review checklist (Section 3)**

Go through each category: Correctness, Security, Code Quality, Testing, Performance, Documentation, Multi-Pod Deployment. Use the diff + metadata to assess each area — no local checkout needed for PRs under ~50 files.

**PITFALL:** Don't checkout locally for small PRs when `gh pr diff` suffices. Checking out, running `git diff main...HEAD`, and running local tests adds friction with no additional signal when the PR is small and CI is already green. Remote-only is faster and equally thorough for most PRs.

For reading individual files from the PR branch without a full checkout, use `gh api .../contents/<path>?ref=<branch> | base64 -d` — see `references/remote-file-inspection.md` for patterns.

### Local Checkout Path (for large PRs needing deeper context)

### Step 1: Set up environment

```bash
source "${HERMES_HOME:-$HOME/.hermes}/skills/github/github-code-review/scripts/gh-env.sh"
# Or run the inline setup block from the top of this skill
```

### Step 2: Gather PR context

Get the PR metadata, description, and list of changed files to understand scope before diving into code.

**With gh:**

For human-readable output:
```bash
gh pr view 123
gh pr diff 123 --name-only
gh pr checks 123
```

**For structured/scriptable output**, use `--json` + `--jq` to get PR metadata, CI status, reviews, and body in one call:
```bash
gh pr view 123 --json title,author,state,isDraft,mergeable,baseRefName,headRefName,body,reviews,statusCheckRollup --jq '.'
```

Key fields: `title`, `author.login`, `state`, `isDraft`, `mergeable` (\"MERGEABLE\"/\"CONFLICTING\"/\"UNKNOWN\"), `baseRefName`, `headRefName`, `body`, `reviews[].state`, `statusCheckRollup[].{name,conclusion,status}`.

**With curl:**
```bash
PR_NUMBER=123
# PR details (title, author, description, branch)
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$GH_OWNER/$GH_REPO/pulls/$PR_NUMBER

# Changed files with line counts
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$GH_OWNER/$GH_REPO/pulls/$PR_NUMBER/files
```

**PITFALL:** `gh pr diff --stat` is NOT a valid flag. `gh pr diff` supports `--name-only`, `--patch`, `--color`, and `--web`. For a stat summary (lines changed per file), use the REST API `/pulls/$PR_NUMBER/files` endpoint (see curl example above), or `git diff --stat` after checking out the PR branch locally.

**PITFALL:** Shell escaping with `gh api` JSON bodies. When using `gh api` to submit reviews, the review body often contains backticks, parentheses `()`, markdown table pipes `|`, dollar signs `$`, and other shell-significant characters. Passing the body via `-f body='...'` or as an inline string will produce shell syntax errors (e.g., `syntax error near unexpected token '('`). Always write the JSON payload to a temp file and use `--input`:

```bash
# Write the full JSON payload (body + event) to a file first
cat > /tmp/review.json << 'REVIEW_EOF'
{"body": "**Approved** — clean change.\n\nAll CI green.", "event": "APPROVE"}
REVIEW_EOF

gh api repos/<owner>/<repo>/pulls/<N>/reviews --input /tmp/review.json
```

This applies to any `gh api` call where the JSON body contains markdown, code snippets, or special characters. The `--input` flag reads the raw file without the shell interpreting its contents.

**PITFALL:** This same escaping problem bites `execute_code` / `terminal` when the command being run contains `gh api` with JSON bodies. The shell that `terminal()` invokes still interprets backticks and parentheses in the command string. Always write the JSON to a file first, then reference the file with `--input`, rather than trying to pass it inline.

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

Go through each category: Correctness, Security, Code Quality, Testing, Performance, Documentation, Multi-Pod Deployment.

### Step 7 (shared): Post the review to GitHub

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

**PITFALL:** Non-pending reviews (COMMENTED, APPROVED, CHANGES_REQUESTED) cannot be deleted — GitHub's API only allows deleting PENDING reviews. If you post a test review comment to verify capabilities, it's permanent. Use a small, obvious "test review" message and be transparent with the PR author.

**PITFALL:** Emoji (e.g., 🔴, ⚠️, ✅) in review bodies may trigger security scanners like Tirith/Wiz that flag Unicode variation selectors. If the terminal blocks your review with "[MEDIUM] Variation selector characters detected", strip all emoji from the body and retry with plain-text markers (e.g., `[HIGH]`, `[MEDIUM]`, `[FIXED]`).

**PITFALL:** Even `$(cat /tmp/review.md)` is NOT a reliable alternative to `--body-file`. The terminal tool's security scanner may reject `&` or other shell-significant characters in the command string regardless of quoting. Always prefer `--body-file` — it passes the path (a safe string) and lets `gh` read the file directly, bypassing both shell interpretation and terminal-level content scanners.

**PITFALL:** When reviewing locale/i18n changes, do NOT flag a locale directory as "missed" without first checking the i18n configuration (e.g., `lingui.config.ts`, `next.config.js` i18n section) to confirm which locales are actually active. A locale directory may exist in the repo but be inactive — the config is the source of truth, not the filesystem.

**PITFALL:** Lingui `.po` diffs showing `#~`-prefixed lines (obsolete entries) are normal catalog hygiene — do NOT flag them as suspicious or erroneous. When `lingui extract` runs, it marks previously-extracted messages that are no longer referenced in source code as obsolete (`#~`) rather than deleting them, preserving translation history. This is standard Lingui behavior and is distinct from merge conflict markers (`<<<<<<<`). Obsolete entries do not affect runtime behavior — they are stripped during `lingui compile`.

**PITFALL:** `git push --force-with-lease` may be blocked by the terminal approval gate even when the intent is corrective (e.g., fixing a commit author email to satisfy Vercel). When this happens, fall back to the GitHub API commit-object + ref-update technique documented in [`references/force-push-via-api.md`](references/force-push-via-api.md). This creates a new commit object with the correct author metadata on GitHub and updates the branch ref directly — no git push required.

### Lingui `.po` Merge Conflicts

When merging branches in monorepos with Lingui i18n, locale `.po` files frequently conflict because `#:` source-reference comments drift between branches. See [`references/lingui-po-merge-conflicts.md`](references/lingui-po-merge-conflicts.md) for diagnosis and resolution patterns. In the common case (line-number-only conflicts with identical msgids), `git checkout --ours` for all `.po` files is safe — `pnpm lingui extract` regenerates references afterward.

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

See Section 3c for the full severity classification framework.

- **Approve** — all findings are resolved (fixed or pushback-commented). All clear with no findings is also fine.
- **Request Changes** — any unresolved finding exists (any severity). Findings are only resolved when the author has either fixed the code or added a comment explaining why a pushback is necessary.
- **Comment** — observations and suggestions, but nothing blocking (use when you're unsure or the PR is a draft)

**🚫 HARD RULE: NEVER approve a PR that has unresolved issues of ANY severity.** Every finding must be either fixed by the author or pushed back on with a comment explaining why the fix isn't necessary. Unresolved findings mean REQUEST_CHANGES regardless of severity label. This rule has no exceptions — not even for feature-gated code, POC branches, or "will fix in follow-up" promises. If the author argues the issues are acceptable, they can override the bot with a pushback comment — but RSI must never be the one to approve through unresolved findings.

Rationale: Every review finding, regardless of severity, represents a potential improvement identified by an independent reviewer. Skipping low-severity findings accumulates technical debt and erodes codebase quality over time. The pushback mechanism gives authors a lightweight way to disagree without blocking the review — but the burden is on the author to explicitly decline a fix, not on the reviewer to silently drop findings.

---

## 5b. Parallel Subagent Review for Large PRs (>50 files, >10K LOC)

When a PR is too large to review sequentially in reasonable time, delegate review to parallel subagents, each covering a different concern area. This is valuable for PRs that span multiple architectural layers (schema, business logic, API surface, infrastructure).

### When to use this pattern

- PR has >50 changed files or >10K lines of diff
- Changes span multiple domains (migrations + worker logic + API routes + config + CI)
- Review requires checking different security/performance properties per layer

### Parallel review decomposition

Spawn 3–4 `delegate_task` subagents simultaneously, each with a focused goal:

| Subagent | Scope | Toolsets |
|----------|-------|----------|
| **DB/Schema reviewer** | Migrations, indexes, constraints, query patterns (N+1, cursor pagination), repository layer, idempotency | terminal, file |
| **Security reviewer** | Admin routes (auth/authz), API schemas (data exposure, PII), config (secret handling, Debug derives), S3/storage (path traversal), SQL injection | terminal, file |
| **Core logic reviewer** | Workflow orchestration, activity implementation, decision logic, signal computation, provider integrations, error handling, race conditions, resource leaks | terminal, file |

Each subagent should:
1. Read the key files in its domain (not the full diff — use `git diff --stat` to identify targets)
2. Report findings with file paths and line numbers
3. Classify severity (HIGH/MEDIUM/LOW) for each finding

### Cross-repo FE PR check (parallel with subagents)

While subagents run, search for paired FE PRs using the GitHub Search API:

```bash
gh search prs --repo=piplabs/numo-monorepo --match title "submission quality"
# Also try matching branch prefix
gh search prs --repo=piplabs/numo-monorepo --head feat/fraud

**PITFALL:** `gh search prs` has a narrower `--json` field set than `gh pr list`/`gh pr view`. It does NOT support `headRefName`, `baseRefName`, `mergeable`, `reviews`, `statusCheckRollup`, or `changedFiles`. When you need those fields, use `gh pr list` with `--search` instead. For branch-prefix matching specifically, `gh search prs --head <prefix>` works fine as a filter (no `--json` needed).

**PITFALL:** A cross-repo partner PR (e.g., story-deployments) may return 404 from `gh pr view` or `gh api`. This is often because the GitHub App token isn't installed on that repo, not because the PR doesn't exist. Run `gh api /installation/repositories --jq '.repositories[].full_name'` to see the token's actual scope. If the target repo isn't listed, ask the repo admin to install the GitHub App on it. See `rsi-platform-investigation` skill, Section 7, for detailed diagnosis. The review should still proceed on the accessible PRs — just flag the inaccessible one rather than blocking.
```

When a BE PR claims a linked FE PR exists but provides no URL, use the thorough 4-step search procedure in `references/cross-repo-fe-pr-search.md`. A `gh search prs --head` alone is not sufficient to prove a PR doesn't exist — always run the full checklist before flagging `missing-cross-repo-pair`.

If the BE PR adds API routes, new schemas, or response fields and no FE PR exists, flag it as `missing-cross-repo-pair`.

### Merging findings

After all subagents complete:
1. Collect findings from each subagent
2. **Verify subagent findings before re-reporting** — subagents frequently produce false positives and false negatives, especially on cross-repo alignment checks. A subagent scanning a 699-line `api-client.ts` may miss `adminApi.withdrawals.*` and report an entire API surface as "uncovered" when it exists. Before reporting any CRITICAL or HIGH finding from a subagent, spot-check 2-3 key files yourself to confirm. If a subagent timed out, re-run a narrower version focused on the files that matter most.
3. Deduplicate (same issue found by multiple reviewers)
4. Classify into the standard output format: Critical → Warnings → Suggestions
5. Add a Verdict section at the bottom
6. Deliver the merged review as a single Slack message or GitHub comment

**PITFALL:** Subagents can't share context. Pass the repo path and branch name in the `context` parameter, and any repo-specific conventions (base branch, deployment topology, cross-repo contract rules) so each subagent has enough context to review independently.

**PITFALL:** Giving subagents too broad a scope causes false negatives. A subagent told to "check cross-repo alignment for all 70 files" will scan too broadly and miss things. Instead, give subagents explicit file paths to check. For a cross-repo alignment subagent, list the specific files where API client calls and route definitions live — e.g., `apps/api/src/http/routes/mod.rs`, `apps/admin/src/lib/api-client.ts`, `apps/admin/src/lib/types.ts`. Asking it to discover these by searching the repo is unreliable.

**PITFALL:** Subagents default to the repo's base branch (e.g., `staging` or `main`), NOT the PR branch. If you give a subagent `search_files` or `read_file` instructions without specifying the PR branch explicitly, it searches the base branch and reports \"findings\" that don't exist on the PR. This produces false positives — the subagent flags code that is already on the base branch, unrelated to the PR. Always pass `branch=<headRefName>` or explicitly check out the PR branch in the subagent's instructions. This is the #1 source of subagent false positives.

**PITFALL:** Even after instructing subagents to use the PR branch, they may silently fall back to base-branch file reads. Always spot-check 2-3 subagent findings against the actual PR diff (which you have from `gh pr diff`) before reporting them as issues.

---

## 5c. Re-Review After Fixes (Delta Review)

When the author addresses feedback and asks for a re-review, do NOT re-read the entire diff. Focus on verifying the specific fixes — this is a fraction of the cost of a full review.

**PITFALL (force-push amended commits):** When the author force-pushes an amended commit (common when fixing review feedback), the commit SHA from `gh pr view --json headRefOid` may not be reachable via `gh api /repos/.../commits/<sha>` — the API returns 422 "No commit found". The PR branch ref still points to the correct tree, so fall back to one of:
- Read individual files remotely: `gh api "repos/<owner>/<repo>/contents/<path>?ref=<branch>" --jq '.content' | base64 -d`
- Use `gh pr diff N` to see the full branch diff (works regardless of force-push history)
- Clone locally and use `git diff develop...HEAD`

### Workflow

1. **Pull latest commits** on both PRs (re-clone if `/tmp/` was cleaned between sessions):

```bash
cd /tmp/<repo>-review
git fetch origin pull/<N>/head:pr-<N> && git checkout pr-<N>
```

2. **Check CI first** — it's the fastest signal. If "Validate migrations" was failing before, it should pass now.

```bash
gh pr view <N> --json statusCheckRollup --jq '[.statusCheckRollup[] | {name, conclusion}]'
```

3. **Inspect the fix commit(s) without checkout** — use `gh api` to see exactly which files were changed in the round-2 commit (the one after your original review). This is faster than diffing the full branch:

```bash
# List commits since your review to find the fix commit(s)
gh pr view <N> --json commits --jq '.commits[-5:] | .[] | {oid: .oid[0:8], message: .messageHeadline}'

# See which files the fix commit touched
gh api "repos/<owner>/<repo>/commits/<fix-sha>" --jq '.files[] | "\(.filename): +\(.additions) -\(.deletions)"'

# Inspect a specific fix in detail
gh api "repos/<owner>/<repo>/commits/<fix-sha>" --jq '.files[] | select(.filename == "path/to/file.ts") | .patch'
```

**PITFALL:** Don't re-clone or re-diff the entire branch. The `gh api` commit-inspection pattern above tells you exactly what changed in the fix commit without a local checkout — use it to verify each flagged issue individually.

4. **Verify specific fixes** — check each previously-flagged issue against the current branch state using remote file reads (no checkout needed for small fixes). See `references/remote-file-inspection.md` for the full pattern set.

```bash
# Read a specific file from the PR branch without cloning
gh api "repos/<owner>/<repo>/contents/<path>?ref=<branch>" --jq '.content' | base64 -d | grep -A5 "<pattern>"
```

5. **Report only the delta** — the re-review message should focus on what changed. Structure:
   - "Previous blocking issues" table: each issue with FIXED / NOT FIXED / N/A status
   - CI status update
   - Remaining items (if any) with severity and whether they were downgraded
   - Verdict

**CRITICAL — Use a fresh subagent for re-reviews (see Section 5d).** The parent agent that performed the original review has anchoring bias. It "knows" auth.ts was clean and "knows" the migration was fine. A fresh subagent has no such knowledge — it reviews every line as if seeing it for the first time. See Section 5d for the full rationale and protocol.

---

## 5d. Fresh Subagent Per Review Pass (Anti-Bias Protocol)

**Core principle:** No agent should review the same PR twice. The agent that performed Review Round 1 has accumulated opinions, assumptions, and blind spots about the code. These pollute Review Round 2, Round 3, and beyond.

### Why This Matters

In a multi-turn review flow (`review → fix → re-review → fix → re-review`), the parent agent accumulates context that creates systematic bias:

| Bias | Example |
|---|---|
| **Anchoring bias** | "I flagged auth.ts:45 as the problem last time, so let me focus there." — misses new issues elsewhere. |
| **Confirmation bias** | "They said they fixed the N+1, and the ORM call looks different, so it's probably fine." — accepts incomplete fixes. |
| **Familiarity blindness** | "I already reviewed auth.ts in Round 1 and it was clean." — skips re-reading, misses new SQL injection introduced by the fix commit. |
| **Leniency drift** | After 3 rounds of back-and-forth, the reviewer wants to be "done" and lowers the bar. Fresh eyes don't have fatigue. |
| **Assumption propagation** | Round 1 found issue X but not issue Y. Round 2 subagent inherits the parent's blind spots about Y because the parent's context omitted it. |

All of these are eliminated by spinning off a **fresh subagent** that has never seen this PR before.

### Protocol

**For EVERY review pass (initial + all re-reviews), the parent agent does NOT review code directly.** Instead:

1. **Parent gathers raw inputs** — PR number, repo URL, branch names, `gh pr diff` output, repo conventions
2. **Parent spawns a fresh subagent** via `delegate_task` with ONLY these inputs
3. **Subagent produces findings** — file+line+severity per Section 3e
4. **Parent verifies subagent findings** — spot-check 2-3 critical findings against the diff
5. **Parent merges and delivers** the review to Slack/GitHub

### What to Pass to the Subagent

```
REQUIRED (pass every time):
- PR number and repo URL
- Branch names (head + base)
- Full git diff output (gh pr diff N)
- Review checklist from Section 3 (all categories)
- Repo-specific conventions (base branch, cross-repo contract, deployment topology)

FOR RE-REVIEWS ONLY (pass additionally):
- Previous review findings, including severity, file/line, expected fix, and current claimed status if known
- Fix commits or patch ranges being verified
- CI/check status before and after the fix commits
- Author responses, explicit pushback, and any items already accepted as non-issues
- A "do not re-flag" list for items that were fixed or intentionally withdrawn
```

### Delegate Task Shape

Use the local `delegate_task` tool or equivalent fresh-subagent primitive. The subagent prompt must be self-contained and must not include the parent agent's opinions about whether the PR is safe.

```
Goal: Independent PR review for <repo>#<number>.

Context:
- Review pass: initial | re-review
- Repo URL, PR number, base branch, head branch
- Full current PR diff
- Applicable review checklist from Section 3
- Repo-specific contracts and deployment topology
- For re-reviews only: previous findings, fix commits, CI deltas, and author responses

Output:
- Findings ordered by severity with file, line, evidence, impact, and concrete fix
- Previous-finding status table for re-reviews
- Explicit "no findings" statement only if no actionable issues remain
- Confidence notes and test gaps
```

### Hard Gate Before Posting

Do not call `gh pr review`, `gh pr comment`, Slack posting tools, `send_message`, or any other external delivery tool until a fresh subagent has completed for the current review pass.

If `delegate_task` is unavailable, fails, or returns unusable output, the review is blocked or partial. Say that clearly on the requested delivery surface. Do not approve, request changes, or present a parent-only review as complete.

When the review request originated in Slack, the Slack thread is the first-class delivery surface. The final harness response may summarize or point to the Slack answer, but the actual review verdict and findings should be delivered to Slack after the fresh-subagent gate is satisfied.

### Parent Verification Duties

The parent agent may verify and format subagent findings, but must not silently replace the subagent review with its own direct review. Verification means:

- Spot-check the highest-severity findings against the current diff
- Drop findings that are demonstrably false positives, explaining why if the drop affects the verdict
- Preserve unresolved findings even if the author claims they are fixed
- Include evidence that a current-pass subagent was used, such as the task label or a short "fresh subagent pass completed" note
