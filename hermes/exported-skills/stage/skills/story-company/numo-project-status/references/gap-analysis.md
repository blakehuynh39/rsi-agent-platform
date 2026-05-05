# Gap Analysis: Finding Numo Items Not in GitHub

How to cross-reference Notion, Slack, and GitHub to surface high-priority items that have no PR or issue tracking.

## When to Use

Triggered when the question is "what high-priority items aren't already in PRs or issues?" — as opposed to the standard "how's the project coming along?" which covers everything. The user is explicitly asking for tracking gaps.

## Methodology

### 1. Gather all five data sources (standard workflow)

Follow the investigation workflow steps 1–5 to collect:
- Slack thread + Honcho corpus (prior check-ins, decisions, blockers)
- Notion backlog database (manifest + schema)
- Both repos: open PRs, open issues, recently merged PRs
- K8s deployment status (for context)

### 2. Build the "tracked items" exclusion set

From the GitHub data, compile what's already captured:
- **All open PRs** (both repos) — item is being actively worked
- **All open issues** — item has tracking, even if unowned
- **Recently merged PRs** — item is shipped

These form your exclusion filter. Anything in this set is NOT a gap.

### 3. Cross-reference Slack discussions against GitHub

For each high-priority item mentioned in Slack threads:
- Does it have a corresponding GH issue? → Already tracked
- Is it in a PR? → Already being worked
- Neither? → **This is a gap**

Common gap patterns from historical sessions:
- Items "drafted" but never filed due to bot write-permission gaps
- P0/P1 items discussed in team channels but never converted to issues
- Cross-repo items where one side has a PR/issue but the other doesn't

### 4. Cross-reference Notion backlog against GitHub

For each item in the Notion "Numo Product Backlog" database:
- Check if there's a matching GitHub issue or PR (by title keywords, linked URLs)
- If the Notion item has a Priority set (P0/P0.5/P1) but no GH counterpart → gap
- If the Notion item has no Owner AND no GH issue → double gap (unowned + untracked)

**Notion API limitation:** The MCP `API_post_search` returns page/data-source manifests with schema properties (Priority, Status, Owner, etc.) but NOT individual database rows with their populated values. To see actual row data you'd need a direct Notion database query. What you CAN determine from the manifest:
- The database schema (which fields exist: Priority, Owner, Status, Person, etc.)
- Individual pages attached to the data source (their title, status, assigned people)
- Whether key fields like Owner are populated or empty

Deduplicate by `database_id` — the same database often appears in multiple snapshot revisions.

### 5. Check for "filed but orphaned" items

Issues that exist in GitHub but are effectively untracked:
- No assignee
- No priority label (P0/P1/P2)
- Filed recently (last 7 days) with no activity

These technically exist but have no owner driving them. Flag as a tertiary category.

## Known Recurring Blockers

### rsi-platform-bot write permission gap

As of May 2026, `rsi-platform-bot` has zero write/triage permissions on:
- `piplabs/depin-backend`
- `piplabs/numo-monorepo`

This means any follow-up items identified by RSI cannot be filed as GitHub issues by the bot. The fix (granting write access) has been discussed since at least Apr 29 but not yet propagated.

**Impact:** The 7 follow-up items from the May 1 thread (seed phrase P0, account deletion P0, PayPal P1, Beehiiv P1, admin API keys P1, monitoring dashboard P1, Intercom P2) are all blocked by this gap. They exist in Slack and Notion but NOT in GitHub.

## Report Format

For gap-analysis reports, use four sections with explicit "why not tracked" columns:

1. **🔴 Blocked items not in GitHub** — table with: Item, Priority, Why Not in GitHub, Blocker
2. **🟡 Notion backlog without GitHub counterparts** — table with: Item, Priority, Status, Owner (or "None"), Evidence
3. **🟠 Filed but untriaged** — list of issue references with missing owner/priority
4. **🟢 Context** — what IS tracked (to show the exclusion set is correct)

## Example Cross-Reference (from May 5, 2026 session)

| Source | Item | In GitHub? | Gap? |
|--------|------|-----------|------|
| May 1 Slack thread | Seed phrase backup P0 | ❌ (related #388/#191 exist but no single tracking issue) | ✅ Gap |
| May 1 Slack thread | Account deletion P0 | ❌ (only partially mitigated by #369) | ✅ Gap |
| Notion Backlog | Poseidon integration P0.5 | ❌ (no GH issue in either repo) | ✅ Gap |
| Notion Backlog | Serve least-used transcript | ❌ (stub page, no fields set) | ✅ Gap |
| numo-monorepo | #228 India payout support | ✅ Filed as issue | ❌ But no assignee |
| depin-backend | #422 trust-safety v2 | ✅ Open PR | ❌ Excluded |
