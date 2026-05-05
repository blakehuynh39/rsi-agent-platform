---
name: numo-project-status
description: "Multi-source investigation of Numo/depin project progress — weekly check-ins, follow-up item tracking, and cross-repo status reports."
version: 1.0.0
metadata:
  hermes:
    tags: [numo, depin, investigation, project-status, weekly, cross-repo, slack, github, kubernetes]
    related_skills: [depin-prod-admin-read, rsi-platform-investigation, github-pr-workflow]
---

# Numo Project Status Investigation

Use this skill when a Story request asks "how is the Numo/depin project coming along this week?", "what follow-up items need attention?", or any general project check-in across repos.

## Investigation Workflow

Run these steps in parallel where possible. The goal is to gather evidence from **all five sources** before synthesizing.

### 1. Read the ingress Slack thread

Always read the full thread first — it defines what's being asked and may contain prior context, links, or specific questions.

```
mcp slack_read_thread(channel_id, thread_ts)
```

### 2. Search Honcho for historical context

Search the Slack corpus for the project name and recent activity. This surfaces prior check-ins, decisions, and blockers from past threads.

```
mcp conversations_search(query="numo")
mcp documents_search(query="numo")
```

**PITFALL:** Honcho corpus may include private channel results marked `mirror_denied: true`. These are still useful context — just note that the full thread isn't accessible.

### 3. Search the wiki for structured knowledge

Wiki pages contain compiled claims and Notion-backed structured data.

```
wiki_search(query="numo")
wiki_page_get(page_ref)  # for any relevant pages found
```

Look for: product backlog pages, project integration pages, Notion database manifests.

### 4. Clone and inspect BOTH repos

The Numo project spans **two repos** that must be checked for cross-repo alignment:

```bash
# depin-backend (Rust API + IP registration worker)
cd /tmp && gh repo clone piplabs/depin-backend
cd depin-backend
git log --oneline --since="<last_week>" --until="<today>" --format="%h %s (%cr)"
gh pr list --state merged --limit 15 --search "merged:>=<last_week>"
gh pr list --state open --limit 5
gh issue list --state open --limit 20

# numo-monorepo (web, admin, mobile, landing)
cd /tmp && gh repo clone piplabs/numo-monorepo
cd numo-monorepo
git log --oneline --since="<last_week>" --until="<today>" --format="%h %s (%cr)"
gh pr list --state merged --limit 15 --search "merged:>=<last_week>"
gh issue list --state open --limit 15
```

**CRITICAL:** Always check **both repos** for cross-repo alignment. A backend PR often has a frontend counterpart (e.g., `depin-backend #385` rewired to `numo-monorepo #202`). Report these pairings when found.

**Base branches:**
- `depin-backend` PRs target **`staging`** (NOT `main`).
- `numo-monorepo` PRs target **`develop`** (changed from `main` as of 2026-05).
- Cross-repo PRs use matching branch prefixes (e.g., `feat/trust-safety-cluster-and-cadence`).
- Backend merges + deploys to staging **before** the FE PR is merged.

**Date ranges:** Use ISO dates for `--since`/`--until` in git log. For `gh pr list --search`, use `"merged:>=YYYY-MM-DD"`.

### 4a. Reviewing a specific cross-repo PR pair

When the ingress request is a PR review (not a weekly check-in), use `gh pr view`/`gh pr diff` directly rather than cloning:

```bash
# BE PR — Rust Checks, Image Builds, Validate migrations, Wiz (Data/IaC/SAST/Secret/Vuln)
gh pr view <N> --repo piplabs/depin-backend --json title,body,author,state,baseRefName,headRefName,createdAt,additions,deletions,changedFiles
gh pr diff <N> --repo piplabs/depin-backend --name-only
gh pr checks <N> --repo piplabs/depin-backend  # confirm Rust Checks, Image Builds, Wiz scanners all pass

# FE PR — Vercel deploys (admin/web/landing), Wiz scanners
gh pr view <N> --repo piplabs/numo-monorepo --json title,body,author,state,baseRefName,headRefName,createdAt,additions,deletions,changedFiles
gh pr diff <N> --repo piplabs/numo-monorepo --name-only
gh pr checks <N> --repo piplabs/numo-monorepo  # confirm Vercel + Wiz all pass
```

**Key review checklist for depin/numo PRs:**

- **Cross-repo contract:** Verify matching branch prefixes, BE targets `staging`, FE targets `develop`, both PRs link each other in descriptions.
- **Migration safety (BE):** CONCURRENTLY for indexes, IF NOT EXISTS for idempotency, no destructive ALTER/DDL, new columns have DEFAULTs or are nullable.
- **Cursor Bugbot reviews:** Both repos use Cursor Bugbot. Check `gh pr view --comments --json comments,reviews` for autofix findings — they often catch stale-audit-timestamp, NULL-handling, and navigation bugs. Note whether autofixes have been applied.
- **Enforcement points (BE):** Soft restrictions (`submissions_paused`, `withdrawals_blocked`) must use `FOR UPDATE` in transaction paths to prevent TOCTOU.
- **Per-row query cost (BE):** New LATERAL JOINs in `list_users`/`list_at_risk_users` add per-row subquery cost — note any window functions (LAG, PERCENTILE_CONT) or multi-UNION subqueries that repeat for every result row.
- **Bundle impact (FE):** New dependencies in `package.json` — flag heavy additions (d3, canvas libraries) and any `as any` TypeScript casts that indicate fragile type boundaries.

### 5. Check Kubernetes deployment status

```bash
kubectl get deployments -n story | grep -E "depin|NAME"
kubectl get pods -n story | grep -E "depin|NAME"
```

Look for: replica counts, recent restarts (indicating deploys), age. The `story` namespace hosts `use1-stage-depin-backend`, `use1-stage-depin-ip-registration-confirmer`, `use1-stage-depin-ip-registration-poller`, and `use1-stage-depin-ip-registration-submitter`.

**PITFALL:** The `rsi-platform` namespace does NOT host depin services. Always use the `story` namespace.

### 6. Check for follow-up items

If prior threads asked for follow-up work (e.g., GitHub issues to be created), verify whether those were completed:

```bash
cd /tmp/depin-backend
gh issue list --state open --limit 20 --json number,title,state,createdAt,labels

cd /tmp/numo-monorepo
gh issue list --state open --limit 15 --json number,title,state,createdAt,labels
```

Look specifically for:
- Issues that were "drafted" but never filed (blocked by permissions, etc.)
- P0/P1 items from prior check-ins
- Customer-reported issues opened recently

## Report Structure

Synthesize findings into three sections:

1. **🟢 This Week's Progress** — Merged PRs, shipped features, commit counts, K8s health. Group by repo (depin-backend, numo-monorepo). Highlight cross-repo pairings.

2. **🔴 Follow-Up Items** — Blocked items, unfiled issues, permission gaps, items from prior check-ins that haven't moved. Use a table format with priority and status columns.

3. **🟡 Watch-Items / Risks** — Open issues requiring attention, unresolved contract questions, recent customer reports, large PRs in flight.

**Evidence standard:** Cite specific PR numbers, issue numbers, thread timestamps, and deployment names. Prefer the `<repo> #<number>` format (e.g., `depin-backend #404`).

See `references/report-template.md` for a concrete example of the three-section report format. See `references/ci-pipeline-checks.md` for the full CI surface of both repos (Rust Checks, Wiz scanners, Vercel deploys, Cursor Bugbot).

## Fallback: when gh isn't authenticated

If `gh` commands fail with auth errors, use `git log --oneline` for the local clone (which was cloned via HTTPS and may still work) and note that PR/issue API data is unavailable.
