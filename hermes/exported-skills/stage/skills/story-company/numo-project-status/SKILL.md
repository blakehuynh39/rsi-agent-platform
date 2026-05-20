---
name: numo-project-status
description: "Multi-source investigation of Numo/depin project progress — weekly check-ins, follow-up item tracking, and cross-repo status reports."
version: 1.1.0
metadata:
  hermes:
    tags: [numo, depin, investigation, project-status, weekly, cross-repo, slack, github, kubernetes]
    related_skills: [depin-prod-admin-read, rsi-platform-investigation, github-pr-workflow]
---

# Numo Project Status Investigation

Use this skill when a Story request asks "how is the Numo/depin project coming along this week?", "what follow-up items need attention?", or any general project check-in across repos. Also trigger when the user asks for high-priority items that are *not* already tracked in PRs or GitHub issues ("gap analysis" — see `references/gap-analysis.md`).

## Investigation Workflow

Run these steps in parallel where possible. The goal is to gather evidence from **all six sources** before synthesizing.

**⚠️ SCALE THE RESPONSE TO THE QUESTION.** Not every request needs a full 6-source investigation:
- **"Latest number of submissions"** / simple stat → just curl the admin API (step 0) and answer concisely
- **"How's the project coming along this week?"** / "latest stats for numo" → full 6-source investigation with 3-section report
- **"What items are NOT in GitHub?"** → gap analysis mode (cross-reference all sources)
- **"Review PR #N"** → use step 4a (PR review mode)

### 0. Query the production admin API (for live stats)

Skip the full investigation when the question is a simple stat query. The production API is at `https://depin.storyprotocol.net` and requires three env vars:

```bash
# These are the actual env var names (NOT INTERNAL_ADMIN_READ_API_KEY!)
DEPIN_ADMIN_READ_API_KEY          # the credential value
DEPIN_ADMIN_READ_API_KEY_HEADER   # the HTTP header name (e.g. "x-admin-read-key")
DEPIN_ADMIN_BASE_URL              # base URL (e.g. "https://depin.storyprotocol.net")
```

**Key endpoints** (see `references/admin-api-endpoints.md` for full schema):

```bash
# Cumulative overview
curl -s -H "${DEPIN_ADMIN_READ_API_KEY_HEADER}: ${DEPIN_ADMIN_READ_API_KEY}" \
  "${DEPIN_ADMIN_BASE_URL}/v1/admin/overview"

# Daily submission counts
curl -s -H "${DEPIN_ADMIN_READ_API_KEY_HEADER}: ${DEPIN_ADMIN_READ_API_KEY}" \
  "${DEPIN_ADMIN_BASE_URL}/v1/admin/stats/submissions"

# Daily new user counts
curl -s -H "${DEPIN_ADMIN_READ_API_KEY_HEADER}: ${DEPIN_ADMIN_READ_API_KEY}" \
  "${DEPIN_ADMIN_BASE_URL}/v1/admin/stats/user-growth"

# Rewards/payout summary
curl -s -H "${DEPIN_ADMIN_READ_API_KEY_HEADER}: ${DEPIN_ADMIN_READ_API_KEY}" \
  "${DEPIN_ADMIN_BASE_URL}/v1/admin/rewards/summary"
```

**PITFALL — env var name mismatch:** The admin key env var was historically documented as `INTERNAL_ADMIN_READ_API_KEY`. The actual name is `DEPIN_ADMIN_READ_API_KEY`. Check both if one returns empty length.

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

### 3. Search the wiki and Notion for structured knowledge

Wiki pages contain compiled claims and Notion-backed structured data.

```
wiki_search(query="numo")
wiki_page_get(page_ref)  # for any relevant pages found
```

Look for: product backlog pages, project integration pages, Notion database manifests.

**Notion database query via MCP:** Use `mcp_*_notion_*_API_post_search(query="numo <topic>")` to search across the Notion corpus. Be aware that results return **page/data-source manifests** (title, URL, schema properties, status), NOT individual database rows with their populated field values. To see actual row data (e.g., which items have no owner), you'll need to cross-reference the manifest fields against GitHub and Slack findings — the gap-analysis methodology in `references/gap-analysis.md` covers this.

**PITFALL:** Notion database results often include stale entries from multiple snapshot revisions (e.g., same database appearing 3+ times with different `last_edited_time` values). Deduplicate by `database_id` and use the most recent revision.

### 4. Clone and inspect BOTH repos

**PITFALL:** Clones under `/tmp/` are **transient across Hermes sessions**. The `/tmp/depin-backend` and `/tmp/numo-monorepo` directories from a previous session will NOT exist in a new session. Always verify the directory exists before `cd`-ing into it. If missing, re-clone. For simple stat queries where the admin API can answer directly, skip the clone entirely — the `depin-prod-admin-read` skill covers that path without repo access.

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
# Stage
kubectl --context hermes-company-computer get deployments -n story | grep -E "depin|NAME"
kubectl --context hermes-company-computer get pods -n story | grep -E "depin|NAME"

# Prod
kubectl --context use1-prod get deployments -n story | grep -E "depin|NAME"
kubectl --context use1-prod get pods -n story | grep -E "depin|NAME"
```

Look for: replica counts, recent restarts (indicating deploys), age. The stage `story` namespace hosts `use1-stage-depin-backend`, `use1-stage-depin-ip-registration-confirmer`, `use1-stage-depin-ip-registration-poller`, and `use1-stage-depin-ip-registration-submitter`. The prod `story` namespace hosts the corresponding `use1-prod-*` deployments.

**PITFALL:** The `rsi-platform` namespace does NOT host depin services. Always use the `story` namespace, and always pass `--context use1-prod` for production checks. A plain `kubectl -n story` command reads the stage context inside the executor.

### 6. Check Grafana/Thanos metrics for runtime health

Grafana is hosted at `https://grafana.ops.storyprotocol.net` with the Thanos (Prometheus) datasource. Three depin dashboards exist: `depin-backend-api` (API Overview), `depin-opening-event` (War Room), and `a6feef84` (Poseidon Depin API). Query live metrics from the Thanos datasource proxy for both stage and prod — see `references/grafana-metrics.md` for the full query reference and pitfalls.

See also `references/observability-investigation.md` for Loki log query patterns, Sentry error investigation, and how to cross-reference logs with source code when investigating depin-backend errors and alerts.

**Credential setup (env vars):**
- `GRAFANA_SERVER=https://grafana.ops.storyprotocol.net`
- `GRAFANA_TOKEN` — service account token for dashboard/health API endpoints
- `RSI_GRAFANA_CF_ACCESS_CLIENT_ID` + `RSI_GRAFANA_CF_ACCESS_CLIENT_SECRET` — Cloudflare Access headers, **required for the datasource proxy** (`/api/datasources/proxy/...`). The `GRAFANA_TOKEN` alone does NOT authenticate the proxy endpoint.

**PITFALL:** `GRAFANA_TOKEN` works for `/api/health`, `/api/search`, and `/api/dashboards/uid/*` but returns 401 on `/api/datasources/proxy/uid/thanos/...` unless you ALSO pass `CF-Access-Client-Id` and `CF-Access-Client-Secret` headers. Always include all three headers when querying Thanos metrics.

**Query pattern (curl through Grafana proxy):**
```bash
ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('<promql>'))")
curl -s \
  -H "Authorization: Bearer ${GRAFANA_TOKEN}" \
  -H "CF-Access-Client-Id: ${RSI_GRAFANA_CF_ACCESS_CLIENT_ID}" \
  -H "CF-Access-Client-Secret: ${RSI_GRAFANA_CF_ACCESS_CLIENT_SECRET}" \
  "${GRAFANA_SERVER}/api/datasources/proxy/uid/thanos/api/v1/query?query=${ENCODED}"
```

**Key metrics to query** (see `references/grafana-metrics.md` for full PromQL):
- Request rate by status, error breakdown (4xx/5xx), latency percentiles (p50/p95/p99), 24h total requests
- Pod CPU (millicores) and memory (MB) from `container_cpu_usage_seconds_total` and `container_memory_working_set_bytes`
- Stage: job label `use1-stage-depin-backend`, namespace `story`
- Prod: job label `use1-prod-depin-backend`, namespace `story`. Use Thanos for time-series metrics and `kubectl --context use1-prod -n story` for read-only pod/deployment/log facts.

**PITFALL — 404 noise masquerading as errors:** Stage often shows an ~89% "error rate" but 88.9% of that is `404 unmatched` — crawlers, health probes, and scanners hitting non-existent paths. To get the real error rate, exclude `path="unmatched"` or focus on `5xx` only. Prod has ~6% 404 from unmatched paths but the noise is much lower relative to real traffic.

**PITFALL — memory anomaly pattern:** Prod sometimes has one pod using 14× more memory than siblings (e.g., `vzp5r` at 570 MB vs 40 MB siblings). Check `avg_over_time(container_memory_working_set_bytes{...}[6h])` — if CPU is low and no restarts, it's likely jemalloc arena retention, not a leak.

### 7. Check for follow-up items

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

The report shape depends on the question type:

### Weekly check-in (default)

Synthesize findings into three sections:

1. **🟢 This Week's Progress** — Merged PRs, shipped features, commit counts, K8s health, Grafana metrics (throughput, latency, errors, resource usage). Group by repo (depin-backend, numo-monorepo). Highlight cross-repo pairings.

2. **🔴 Follow-Up Items** — Blocked items, unfiled issues, permission gaps, items from prior check-ins that haven't moved. Use a table format with priority and status columns.

3. **🟡 Watch-Items / Risks** — Open issues requiring attention, unresolved contract questions, recent customer reports, large PRs in flight.

### Gap analysis ("items NOT in GitHub")

When the user asks for items that are *not* captured in PRs or issues, cross-reference all six sources and surface:

1. **🔴 Blocked items not in GitHub** — Items discussed in Slack/Notion that have no corresponding GH issue or PR. Include the blocker reason (e.g., bot write permissions).

2. **🟡 Notion backlog without GitHub counterparts** — Backlog items with priority set but no linked GH issue. Highlight missing Owner/Assignee fields.

3. **🟠 Filed but untriaged** — Issues that exist but lack owners, priority labels, or assignment. These are "in the system" but effectively orphaned.

For each item, include a "Why it's not tracked" column explaining the gap. See `references/gap-analysis.md` for the full cross-referencing methodology.

**Evidence standard:** Cite specific PR numbers, issue numbers, thread timestamps, and deployment names. Prefer the `<repo> #<number>` format (e.g., `depin-backend #404`).

See `references/report-template.md` for a concrete example of the three-section report format. See `references/admin-api-endpoints.md` for the production admin API endpoint reference and curl patterns — use this for simple stat queries (submission count, user count, payout status). See `references/ci-pipeline-checks.md` for the full CI surface of both repos (Rust Checks, Wiz scanners, Vercel deploys, Cursor Bugbot). See `references/grafana-metrics.md` for Thanos/Prometheus query patterns, pitfalls, and metric interpretation including memory anomaly investigations. See `references/observability-investigation.md` for Loki log query patterns, Sentry error investigation, Prometheus metrics, and how to cross-reference logs with depin-backend source code when investigating errors and alerts. See `references/code-fix-workflow.md` for the investigate → fix → PR workflow when a production observation needs a code change. See `references/gap-analysis.md` for the cross-referencing methodology used when surfacing items not tracked in GitHub. See `references/depin-codebase-investigation-paths.md` for code-level investigation paths when answering product/logic questions about how depin features work. See `references/admin-auth-architecture.md` for the current admin auth/ACL architecture across both repos — JWT extractors, email allowlist, route structure, DB tables, and FE auth flow. See `references/story-l1-chain-parameters.md` for Story Protocol L1 blockchain parameters (block time, gas limit, consensus, throughput analysis) — reference when investigating chain-side throughput, gas costs, or bottleneck analysis.

### 8. Code-logic investigation (product questions about how features work)

When the question is about *how* a specific feature works (e.g. "what multipliers do top contributors get?", "why do different languages have different payouts?") — not about project progress — investigate the source code directly. See `references/depin-codebase-investigation-paths.md` for the complete file reference by question type.

Quick path: read `AGENTS.md` → `search_files` for key terms → read service files → cross-ref design specs in `docs/superpowers/specs/` → check migrations for schema defaults.

**PITFALL — clone method:** Prefer HTTPS clone (`git clone https://github.com/piplabs/depin-backend.git`) over SSH. The SSH key is not always available in the agent environment and will fail with `error: cannot run ssh: No such file or directory`.

## Fallback: when gh isn't authenticated

If `gh` commands fail with auth errors, use `git log --oneline` for the local clone (which was cloned via HTTPS and may still work) and note that PR/issue API data is unavailable.
