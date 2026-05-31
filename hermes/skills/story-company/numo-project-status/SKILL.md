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

**Date ranges:** Use ISO dates for `--since`/`--until` in git log. For `gh pr list --search`, use `"merged:>=YYYY-MM-DD"`.

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

See `references/report-template.md` for a concrete example of the three-section report format.

## Fallback: when gh isn't authenticated

If `gh` commands fail with auth errors, use `git log --oneline` for the local clone (which was cloned via HTTPS and may still work) and note that PR/issue API data is unavailable.
