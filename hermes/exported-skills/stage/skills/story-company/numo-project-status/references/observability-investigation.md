# Depin-Backend Observability Investigation

How to investigate depin-backend errors and alerts using Loki, Sentry, Prometheus, and source code. Use this when a Slack thread asks "dig into these errors" or "why is X happening."

## Investigation flow

1. **Read the Slack thread first.** The alert history in the channel (Grafana bot messages) often tells you exactly what fired. Look for FIRING/RESOLVED alert cards with titles, values, and labels — `env_target` tells you if it's stage or prod.

2. **Check active alerts.** `rsi_observability_active_alerts` confirms whether anything is firing right now.

3. **Query Sentry for error issues.** Start broad:
   ```
   rsi_sentry_issues_list(project_ref="depin-backend", query="is:unresolved environment:staging", sort="new")
   ```
   **PITFALL:** Sentry issue search matches against issue TITLES only. Terms like `eth_estimateGas` or `estimate_gas` that live in event breadcrumbs or tracing fields will return 0 results even if the data exists. Use broad queries and drill into individual issues with `rsi_sentry_issue_view`.

4. **Drill into a specific issue.** Use `rsi_sentry_issue_view(issue=<id>, spans=5)` to get the full event including breadcrumbs, Rust tracing fields (error message, wallet address, nonce, campaign_id, entity_id, etc.), tags, and release info. The `environment` tag confirms staging vs production.

5. **Query Loki for log context.** Use `rsi_observability_logs_query` with:
   - **IP registration workers**: `{app=~".*ip-registration.*", environment="stage"}`
   - **Broad namespace search**: `{namespace="story", environment="stage"} |= "error"`
   - Simple line-contains filters: `|= "estimate_gas"`, `|= "max_discard_attempts"`, `|= "revert"`
   - Narrow by component: add `component="submitter"` for submitter-only logs
   - **PITFALL:** `{app="depin-backend"}` often returns 0 results for API logs. Try `{namespace="story"}` with `|=` filters instead.

6. **Clone the repo and read the code.** The AGENTS.md file is the table of contents. For IP registration state machine questions:
   - `docs/architecture/ip-registration.md` — state machine diagram + submitter/confirmer flows
   - `apps/ip-registration/src/submitter.rs` — the discard/commit/abort logic
   - `apps/ip-registration/src/infra/chain.rs` — error classification (`classify_estimate` for eth_estimateGas reverts)
   - `docs/plans/active/` — design plans explaining WHY the current behavior exists
   - **PITFALL:** Prefer HTTPS clone (`git clone https://...`). SSH may fail with `error: cannot run ssh`.

7. **Cross-reference logs with code.** The log message format (e.g., `"ip-registration pre-broadcast revert at estimate_gas — discarding attempt without consuming retry budget"`) can be `grep`'d in the source to find the exact code path. `search_files(pattern="estimate_gas", target="content")` surfaces all related code.

8. **Check Prometheus for current rates.** `rsi_observability_metrics_query` with PromQL like:
   ```
   sum by (environment, status) (rate(ip_registration_total{environment="stage", status="reverted"}[1h]))
   ```
   This tells you if the issue is still active or has subsided.

## Key Loki label patterns for depin services

| Service | Loki app label | Component labels |
|---|---|---|
| IP Registration submitter | `use1-stage-depin-ip-registration` | `component="submitter"` |
| IP Registration confirmer | `use1-stage-depin-ip-registration` | `component="confirmer"` |
| IP Registration poller | `use1-stage-depin-ip-registration` | `component="poller"` |
| Depin Backend API | May NOT be in Loki under `depin-backend` | Try `{namespace="story"}` |

## Interpreting eth_estimateGas reverts

When you see `execution reverted, data: "0x"` at `estimate_gas`:

- **This IS expected state machine behavior.** The submitter's Discard path (step 3c) routes `ChainError::PreBroadcastRevert` to `discarded` without consuming nonce or retry budget.
- **A single job failing across all wallets** means a deterministic contract revert — likely a staging chain state mismatch (stale collection address, wiped deployment, test data inconsistency).
- **The safety cap** (`max_discard_attempts=20`) prevents infinite loops — after 20 consecutive discards, the job fails terminally (DEPIN-BACKEND-11).
- **Not a transient RPC issue** if the same job/campaign_id fails across 20+ different wallets with the identical `data: "0x"` revert.

## Interpreting Grafana alerts for depin

Alert notification channels for depin:
- `alert-numo-backend-warning` (channel C0AN6SRLWLB) — Tier1-stage and Tier2 warnings
- `alert-numo-backend-critical` — Tier1-prod criticals

Alert labels to check:
- `env_target`: `stage` vs `prod` — tells you which environment
- `severity`: `warning` vs `critical`
- `contact_point`: which channel the alert routes to
- `grafana_folder`: always `DePIN Backend` for these alerts
- `team`: always `depin-backend`
