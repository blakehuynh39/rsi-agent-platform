---
name: project-health-check
description: Cross-surface project audit — find unanswered questions, misalignments, and stale items across Slack, Notion, and GitHub. Use when asked to check project status, identify gaps, or find things that fell through the cracks.
---

# Project Health Check

Cross-reference mirrored Slack, Notion, and live GitHub to surface unanswered questions, misalignments, and stale/unresolved items on a project. Good for "what did we forget?", "any open questions?", "what's out of sync?" and post-launch audits.

## Triggers

- "unanswered questions", "open items", "what did we miss"
- "misalignments", "inconsistencies", "out of sync"
- "project health check", "audit the project"
- "what's still pending", "what hasn't been decided"
- Post-launch retrospectives and fast-follow identification

## Workflow

### Phase 1 — Cast a wide net (parallel)

Run these four searches simultaneously across both Slack and Notion corpora:

1. **Unanswered questions / decisions:** `open question OR unanswered OR "how to" OR "should we" OR "what about" OR "need to decide" OR pending decision`
2. **Blockers / stuck items:** `blocked OR stuck OR waiting OR help needed OR "any update" OR "what's the status"`
3. **Bugs / broken things:** `not working OR bug OR broken OR "doesn't work" OR issue OR error OR fix needed`
4. **Domain-specific terms:** project name, repo names, key features (e.g., `seed phrase OR castle OR beehiiv OR referral OR payout`)

Use `mcp_rsi_task_trace_*_slack_*_conversations_search` and `documents_search` for each query. Also try `session_search` for prior RSI session summaries.

### Phase 2 — Read promising threads

For each high-signal thread found, read the full conversation via `messages_read(channel_id, thread_ts)`. Look for:
- Questions asked with no follow-up response
- "waiting to hear back", "pending", "TBD"
- Items marked done in one place but broken in reality
- People tagged for action with no visible response

### Phase 3 — Cross-reference with live state

Pull open GitHub issues (`gh issue list --state open --json number,title,createdAt,labels`) and cross-reference:
- Notion items marked "[x] done" → does a GitHub fix exist? Was it merged?
- GitHub issues labeled P0/P1 → do they match Notion priority?
- FE vs BE issues on the same feature → are they connected or working on divergent assumptions?

### Phase 4 — Identify misalignments

Look for these specific patterns:
- **Notion checkbox vs reality:** Item marked complete in checklist but broken/partial in Slack discussions
- **Cross-repo drift:** Specs in one repo that don't match the other (e.g., OpenAPI vs FE API docs)
- **Conceptual divergence:** Same feature described differently across FE, BE, and Notion
- **Config mismatch:** App config allows X but infra/WAF blocks X (especially CORS/domains)
- **Past-due decisions:** Items with explicit due dates that have passed with no resolution
- **Parent/child status gap:** Top-level item unchecked while all children are checked
- **Schema quality:** Duplicate status values in Notion databases
- **Disconnected tracks:** FE and BE working on the same feature with no integration plan between them

### Phase 5 — Verify deployment (when applicable)

When the audit includes "is this deployed?" or "has X been deployed to staging/prod?":

1. **Pull latest git:** `git fetch --all` — local clones may be stale; the deployed commit may only exist on `origin/main`
2. **Find the target commit:** `git log origin/main --oneline --grep="<keyword>"` or `gh api repos/<org>/<repo>/commits`
3. **Check Kubernetes image tags:** `kubectl get deployments -n <ns> -o json` → extract commit hashes from ECR image tags
4. **Compare:** Does the deployed hash match the latest commit on origin/main? If yes, code is deployed.
5. **Verify API endpoints:** Curl the relevant API paths via ClusterIP (no auth needed internally). Check for 404 → not wired, 200 → functional.
6. **Check the data/content layer:** Even when code is deployed, content may not be populated yet (e.g., company wiki directory may be empty, database may need seeding, mirror/ingestion may not have run).

**Pitfall:** The `hermes-skill-exporter` pod may show a newer deployment time than others even when on the same commit hash — it's deployed as a separate pipeline step and may restart independently.

### Phase 6 — Deliver structured answer

Group findings into two sections:
1. **Unanswered questions** — organized by category (product decisions, engineering/QA, process/payment)
2. **Misalignments** — each with concrete evidence from at least two sources

End with actionable "would you like me to" options (create issues, draft cleanup PRD, verify contracts).
