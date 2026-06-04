---
name: github-code-review
description: "Review PRs and local changes with structured multi-angle analysis."
version: 3.0.0
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [GitHub, Code-Review, Pull-Requests, Git, Quality]
    related_skills: [github-auth, github-pr-workflow, destructive-migration-safety]
---

# GitHub Code Review

Review local changes or open PRs. Use `git` for local, `gh` for PR interactions.

## Prerequisites

```bash
# Auth detection (gh preferred, fallback to GITHUB_TOKEN)
if command -v gh &>/dev/null && gh auth status &>/dev/null; then AUTH="gh"; else AUTH="git"; fi
REMOTE_URL=$(git remote get-url origin)
OWNER_REPO=$(echo "$REMOTE_URL" | sed -E 's|.*github\.com[:/]||; s|\.git$||')
```

---

## 1. Local Review (Pre-Push)

```bash
git diff main...HEAD --stat            # scope
git diff main...HEAD                   # full diff
git diff main...HEAD | grep -nE "print\(|console\.log|TODO|HACK|debugger"  # cruft
git diff main...HEAD | grep -inE "password|secret|api_key|token.*=|private_key"  # secrets
git diff main...HEAD | grep -n "<<<<<<|>>>>>>|======="  # conflict markers
```

Present findings as: `{severity} | {file}:{line} — {issue} → {fix}`

---

## 2. PR Review

### Gather & Diff

```bash
gh pr view N --json number,title,body,state,author,files,additions,deletions,baseRefName,headRefName,labels,reviews,mergeable
gh pr checks N                       # CI status (fastest signal)
gh pr diff N                         # full diff
```

**PITFALL:** `gh pr diff N -- path/file` does NOT work. Use `gh pr diff N | sed -n '/^diff --git a\/path\/file/,/^diff /p'` or clone + `git diff`.

**PITFALL:** `gh pr checks` has no `--json`. Use `gh pr view N --json statusCheckRollup` for structured CI data.

### Post Review

```bash
# --body-file avoids shell escaping issues with markdown/special chars
cat > /tmp/review.md << 'EOF'
## Code Review

**Verdict: {APPROVE | REQUEST_CHANGES | COMMENT}**

### Findings
🔴 CRITICAL | src/auth.py:45 — SQL injection → parameterized query
🟠 HIGH | src/worker.rs:120 — non-idempotent INSERT → ON CONFLICT DO NOTHING
🟡 MEDIUM | tests/auth.test.ts — missing expired-token edge case
### Looks Good
- Clean error handling in middleware
EOF

gh pr review $N --{approve|request-changes|comment} --body-file /tmp/review.md
gh pr comment $N --body-file /tmp/review.md  # also post as top-level summary
```

**PITFALL:** Use `--body-file` not `--body "..."` — markdown backticks/pipes/dollars break shell escaping.
**PITFALL:** APPROVED/CHANGES_REQUESTED/COMMENTED reviews are permanent. No deletion via API.

### Check Out PR Locally (for deep review)

```bash
git fetch origin pull/N/head:pr-N && git checkout pr-N
git diff main...HEAD --name-only
# Use read_file on changed files for full context
```

---

## 2b. Cross-Repo Paired PR Review (depin-backend ↔ numo-monorepo)

BE repo: `piplabs/depin-backend` (base: `staging`). FE repo: `piplabs/numo-monorepo` (base: `develop`).

**Checklist:**
- [ ] Route paths match (BE route == FE API client call)
- [ ] New response fields on both sides, identical names
- [ ] FE handles `null` for new fields on historical rows (`!= null` not `||`)
- [ ] CI green on both
- [ ] BE merges to `staging` BEFORE FE merges to `develop`
- [ ] BE PR description links FE PR (and vice versa)

**Changes requiring linked FE PR:** new/renamed routes, new response fields, changed field semantics, new query params, changed auth behavior.
**No FE PR needed:** internal refactors, perf changes, logging, feature-flagged code, migration-only.

**Detection:** Check BE PR description for numo-monorepo link → check commits for api-reference doc changes → if schema/route change and no FE PR → flag `missing-cross-repo-pair`.

**PITFALL:** New admin endpoint without FE PR = operators can't reach it. Always verify.
**PITFALL:** Inaccessible deployment PRs (e.g., story-deployments 404) — note as dependency, don't block.

---

## 3. Review Checklist

Every finding must be: `{severity} | file:line — issue → fix`. See Section 3a for severity labels.

### Correctness
- Logic correct? Edge cases (empty, null, large, concurrent)?
- Race conditions: shared state protected? Lock ordering consistent? `SELECT FOR UPDATE`?
- Transaction boundaries: atomic ops in `BEGIN...COMMIT`? No `COMMIT` inside loops?
- Error propagation: no `unwrap()`/bare `_` in prod; `.catch()` with recovery?
- Partial failure: rollback/compensating action if step B fails after step A?
- Ordering deps: await before reading results? Event handler registered before event fires?
- Off-by-one (`<` vs `<=`), integer overflow, empty collection, single-element edge case.
- Time: monotonic clock for durations, UTC storage, DST boundaries.

### Security
- No hardcoded secrets/API keys. Input validation. No SQL injection/XSS/path traversal.
- Auth/authz on new endpoints. No PII in logs. No secrets in client-bundled code.
- Crypto: no `Math.random` for tokens, no MD5/SHA1, no custom crypto.
- IDOR: does the endpoint verify resource ownership?

### Code Quality & Architecture
- Clear naming. Single responsibility. No duplication.
- Boundary drift: UI reaching into DB? Domain types importing transport types?
- Premature abstraction: one-implementation interfaces are debt.
- Coupling: bidirectional imports? Shared mutable state introduced?
- Scalability: at 10x volume, what breaks? Missing pagination?
- Reversibility: one-way doors (migrations, API contract changes) flagged.

### Testing
- New paths tested with specific assertions (not just `!= null`).
- Mocks verify real contracts. Over-mocking = testing mocks, not code.
- Determinism: stub `Date`/`random`/`uuid`/network. No flakes allowed.
- Snapshot hygiene: >50 lines without review = brittle. Auto-accepted = useless.
- Test names describe behavior: `'returns 403 without admin role'`.

### Performance
- N+1: loop iterating query results, then DB call per iteration. Scan for `for... { await db... }`.
  - Rust: `for row in rows { sqlx::query_as!(...) }` — batch or JOIN.
  - TS: `prisma.x.findMany()` without `include:` then per-item `findMany`.
  - Java: `@OneToMany(LAZY)` accessed in loop without `JOIN FETCH`.
  - Python: `.children` in loop without `joinedload()`/`selectinload()`.
  - Go: `db.Select(...)` then `for ... { db.Get(...) }`.
- Hot-path allocations: new objects in loops, regex recompiled per call.
- Async: sequential awaits where `Promise.all` correct. Missing concurrency limits.
- Cache: keys missing variables, TTLs absent.

### Idempotency & Safe Retries
- **Every mutation must survive running twice.** Network, deploy rollovers, DB pool exhaustion all cause retries.
- API: POST needs `Idempotency-Key` header. PATCH needs optimistic locking (`WHERE version = $1`).
- DB: `INSERT ON CONFLICT` not plain `INSERT`. `UPDATE SET x=x+1` is NOT idempotent.
- Jobs: every-pod jobs must be idempotent (2-pod test). Scheduled tasks need overlap guard.
- External: Stripe idempotency keys. Emails need dedup. Webhooks need idempotency key.
- **Ask:** "what happens if this runs twice?" → duplicate/bad state → flag.
- Idempotency key tables need `created_at` + cleanup job (see depin-backend `idempotency_cleanup`).

### Documentation
- Public APIs documented. Non-obvious logic has "why" comments.

### Multi-Pod Deployment (services with >1 replica)
- Verify replica count (`kubectl get deployments -n <ns>`).
- Background jobs: idempotent? Cheap enough to run N×? Should one pod run it (`pg_try_advisory_lock`)?
- In-memory state: caches/rate-limiters per-pod = split-brain risk.
- Migrations: `CREATE INDEX CONCURRENTLY` (blocking index locks ALL pods).
- Startup: all N pods firing expensive init simultaneously?
- **Multiply DB load by replica count.** `refresh_all()` at 2 replicas = 2× scans.

---

## 3a. Severity & Decision

| Label | Meaning | Resolve By |
|---|---|---|
| 🔴 CRITICAL | Security, data loss, crash, double-charge, PII leak | Fix or pushback comment |
| 🟠 HIGH | Core logic bug, N+1 on hot path, non-idempotent mutation, missing cross-repo pair | Fix or pushback comment |
| 🟡 MEDIUM | Non-hot perf, missing error handling, test gaps | Fix or pushback comment |
| 🔵 LOW | Style, minor DRY, outdated comment | Fix or pushback comment |
| 💡 SUGGESTION | Optional improvement | Fix or pushback comment |

**HARD RULE (for PRs targeting `staging` or feature branches):** Any unresolved finding (any severity) → **REQUEST_CHANGES**. Author must fix OR add pushback comment explaining why not. All resolved → **APPROVE**.

**EXCEPTION — Staging→Main Promotion PRs:** When the `headRefName` is a staging branch (e.g., `staging-gcp`, `staging`) and `baseRefName` is `main`, the PR is a staging→main promotion. These should **always be APPROVED** — the code is already reviewed and vetted on staging. Document any findings as follow-up items for a staging fix, but do not block the promotion. The purpose is to keep prod in-line with staging.

---

## 3b. Review Tone

Question approach (non-security): *"What happens if `items` is empty?"* not *"This will fail."*

Direct for security: *"SQL injection: user input in query string. Use parameterized queries."*

Include 🌟 PRAISE for good work.

---

## 4. Multi-Angle Review

For PRs >10 files or >500 lines, do separate passes:

1. **Security** — auth, injection, secrets, data exposure. Skip style/perf.
2. **Correctness** — logic, edge cases, errors, races, transactions.
3. **Architecture** — boundary drift, coupling, naming, reversibility.
4. **Performance** — N+1, allocations, async, caching.
5. **Tests** — assertion strength, mocking, determinism.

For small PRs, single pass is fine. Mentally separate security from style.

---

## 5. Workflow: Small PRs (<50 files, <1000 lines)

```bash
gh pr view N --json number,title,body,state,author,files,additions,deletions,baseRefName,headRefName,labels,reviews,mergeable
gh pr checks N
gh pr diff N
# Apply Section 3 checklist. No local checkout needed.
```

For reading individual files remotely: `gh api "repos/<o>/<r>/contents/<path>?ref=<branch>" --jq '.content' | base64 -d`

---

## 5b. Workflow: Large PRs (>50 files, >10K LOC)

Spawn parallel subagents, each with one domain. Pass: repo path, branch name, explicit file paths. Never pass: opinions, conversation history, previous review text.

| Subagent | Scope |
|---|---|
| DB/Schema | Migrations, indexes, N+1, idempotency, query patterns |
| Security | Auth/authz, data exposure, secrets, SQL injection, path traversal |
| Core logic | Workflow, error handling, race conditions, resource leaks |

**PITFALL:** Subagents default to base branch. Always pass `branch=<headRefName>` explicitly.
**PITFALL:** Spot-check 2-3 findings against actual diff before reporting. Subagents produce false positives.
**PITFALL:** Timed-out subagent = review didn't happen. Re-spawn with narrower scope, don't fill from memory.

---

## 5c. Re-Review (Delta)

**CRITICAL: Spawn a fresh subagent for every re-review pass (even single-file).** The parent has anchoring bias — it "knows" file X was clean, "knows" the migration was fine. Fresh eyes see what the parent now can't.

**Parent only:** gathers raw inputs (PR diff, repo conventions, structured previous-issue checklist), spawns subagent, verifies 2-3 findings, delivers results.

**Pass to subagent:** PR diff, checklist rules, repo conventions, structured previous-issues list, "DO NOT RE-FLAG" list.
**NEVER pass:** parent opinions, previous review body text, conversation history, "I already checked X" claims.

For trivial fixes (≤3 lines, ≤1 file), parent may verify directly.

Delta report format: table of previous issues with FIXED/NOT FIXED status, CI update, remaining items, verdict.

---

## 5d. Fresh Subagent Per Pass (Anti-Bias)

Multi-turn review accumulates bias: anchoring, confirmation, familiarity blindness, leniency drift, assumption propagation.

**Protocol:** For every review pass, parent never reviews code directly. Parent gathers raw inputs → spawns fresh subagent → verifies findings → delivers. Subagent gets only: raw diff + checklist + structural metadata. Forms its own opinions.

**PITFALL:** `gh pr diff --stat` is NOT valid. Use REST API `/pulls/N/files` or `git diff --stat` after checkout.
**PITFALL:** Force-push amended commits break `gh api /commits/<sha>`. Fall back to `gh pr diff N` or remote file reads.
**PITFALL:** Emoji in review bodies trigger security scanners. Strip to plain-text markers `[HIGH]`, `[FIXED]`.
**PITFALL:** Lingui `.po` `#~`-prefixed lines are normal catalog hygiene, not merge conflicts.
