# Depin-Backend Codebase Investigation Paths

Use this reference when asked to investigate *how* a depin feature works from the code — product questions about multipliers, rewards, task logic, language differences, etc. The goal is to find the code that answers the question, not to assess project progress.

## Starting point

Always read `AGENTS.md` first — it's the codebase table of contents. It points to:
- `docs/rewards/` — reward system design + operator notes
- `docs/superpowers/specs/` — authoritative design specs
- `docs/architecture/` — service deep-dives (API, IP registration, database)

## Investigation by question type

### Rewards, multipliers, and payouts

| What to find | Where to look |
|---|---|
| Multiplier calculation logic | `apps/api/src/services/multipliers.rs` |
| Poseidon Season 1 tier mapping | `scripts/ingest_poseidon_season1.py` (tier math), `docs/superpowers/specs/2026-04-24-poseidon-season1-ingestion-design.md` (full design) |
| Cliff+decay curve | `apps/api/src/services/poseidon.rs` → `cliff_decay_value()` |
| Bonus formula (`α` dampening) | `docs/rewards/overview.md`, `docs/rewards/design-spec.md` |
| Multiplier task bars (quests) | `apps/api/src/services/multiplier_tasks/mod.rs` |
| Task bar evaluators (daily subs, streaks, verified total) | `apps/api/src/services/multiplier_tasks/evaluators/*.rs` |
| Reward config knobs (`α`, `max_effective_multiplier`, claim deadline) | `apps/api/src/services/payouts.rs` → `current_config()` |
| Payout finalization (advance + remainder + bonus) | `apps/api/src/services/payouts.rs`, `apps/api/src/services/submissions.rs` (reward snapshot context) |

### Tasks, campaigns, and limits

| What to find | Where to look |
|---|---|
| Campaign CRUD + access control | `apps/api/src/services/campaigns.rs` |
| Daily cap, cooldown, timeout logic | `apps/api/src/services/campaigns.rs` → `campaign_access_for_loaded_campaign()` |
| Submission lifecycle + overflow detection | `apps/api/src/services/submissions.rs` → `increment_campaign_completed_tasks()` |
| Public visibility rules (starts_at/ends_at/is_active) | `apps/api/src/services/campaigns.rs` → `is_campaign_publicly_visible()` |
| Script assignment + per-language content | `apps/api/src/services/campaigns.rs` → script sections |

### Language-specific behavior

| What to find | Where to look |
|---|---|
| Auto-scaling rate formula | `apps/api/src/services/language_targets.rs` → `compute_auto_rate()` |
| Modality rate ranges | `apps/api/src/services/language_targets.rs` → `modality_rate_range()` |
| Language target hours (defaults) | `apps/api/migrations/0021_reward_system.sql` (INSERT into `language_targets`) |
| Per-language campaign script availability | `scripts` table (has `language_code` column), `apps/api/src/services/campaigns.rs` script methods |

### Numo Data Validation (NDV) — internal validation pipeline

The NDV endpoints live under `/v1/internal/numo-data-validation/` and integrate with PSDN (the external validation service, also referred to internally as "Poseidon"). Two endpoints: source export (GET) and result ingestion (POST).

**Key files by concern:**

| What to find | Where to look |
|---|---|
| Design spec (authoritative) | `docs/superpowers/specs/2026-05-02-numo-data-validation-app-integration-design.md` |
| Implementation plan (task-level) | `docs/superpowers/plans/2026-05-02-numo-data-validation-app-integration.md` |
| HTTP handlers (routes, auth, body parsing) | `apps/api/src/http/routes/numo_validation.rs` |
| Auth extractor (bearer token + IP allowlist) | `apps/api/src/http/extractors_numo_validation.rs` |
| Export service (watermark + cursor) | `apps/api/src/services/numo_validation/export.rs` |
| Ingest service (batch insert + state binding) | `apps/api/src/services/numo_validation/ingest.rs` |
| State binding bridge (NDV decision → submission state) | `apps/api/src/services/numo_validation/state_binding.rs` |
| Request/response DTOs | `apps/api/src/services/numo_validation/schema.rs` |
| DB schema: validation_runs, validation_results, pointers | `apps/api/migrations/0062_*` through `0065_*`, `0061_campaign_auto_apply_validation.sql` |
| Config (enabled, CIDRs, token hashes) | `apps/api/src/bootstrap/config.rs` → `NumoValidationConfig`; `apps/api/config/base.toml` → `[numo_validation]` |
| Admin UI validation history | `apps/api/src/services/admin_dashboard.rs`, `apps/api/src/http/routes/admin.rs` |
| Runbook (token rotation, CIDR updates, triage) | `docs/runbooks/numo-validation.md` |
| Grafana dashboard | `grafana/numo-validation-dashboard.json` |

**Trace for "why do results immediately show as paid/failed in the activity tab?":**

1. `POST /v1/internal/numo-data-validation/validation-results` → `routes/numo_validation.rs::post_results_inner()` (line 265)
2. → `services/numo_validation/ingest.rs::ingest_batch()` (line 43)
3. Inside ingest_batch, after the DB transaction commits (line 135), a loop (lines 137-165) calls `state_binding::apply_decision()` for each pass/reject candidate
4. → `services/numo_validation/state_binding.rs::apply_decision()` (line 43) — gates on `campaigns.auto_apply_validation` and `submissions.state = 'pending_review'`
5. → `services/submissions.rs::review_submission_with_current_validation()` (line 267) — which calls `review_submission_with_actor_impl()` (line 284)
6. `review_submission_with_actor_impl()`: UPDATE submissions.state → `finalize_accepted()`/`finalize_rejected()` (payout) → `apply_review_side_effects()` (line 330)
7. → `apply_review_side_effects()` (line 474) → `activities::update_submission_activity_status()` (line 488) — flips activity card to "success" (paid) or "failed"

**Key config flags for buffering:**
- `campaigns.auto_apply_validation` (DB column, migration 0061) — per-campaign gate. When FALSE, `apply_decision()` returns `PointerOnly` (no state change, no payout, no activity update). Results are still stored in `validation_results` + `submission_current_validation` with `applied_to_state = FALSE`.
- `numo_validation.immediate_bind` (proposed, see plan `2026-05-11-ndv-ingestion-buffer.md`) — system-level gate that skips the `apply_decision()` loop entirely during ingestion. Defaults `true` for backward compat.

### Config and environment

| What to find | Where to look |
|---|---|
| Config hierarchy | `apps/api/config/{base,local,staging,production}.toml` |
| Config resolution order | `AGENTS.md` "Config model" section |
| Reward config table schema | `apps/api/migrations/0021_reward_system.sql` |

### Season 1 multiplier — two-layer clamp architecture

The Poseidon Season 1 handoff has a **two-layer clamp** that's easy to confuse. Understanding both layers prevents wrong answers about the effective range.

| Layer | What | Floor | Cap | Where |
|---|---|---|---|---|
| `m_szn_initial` | Season 1 multiplier value at compute time | 1.05 | 2.00 | `scripts/ingest_poseidon_season1.py:74` |
| `effective_multiplier` | Overall multiplier applied to earnings | 1.0 | 2.00 (`max_effective_multiplier` in `reward_config`) | `multipliers.rs:79` |

The `m_szn_initial` formula: `clamp(tier_value + lang_value, 1.05, 2.00)`.

**Edge cases demonstrating the clamp:**

| User | tier_value | lang_value | raw | m_szn_initial | Clamp direction |
|------|-----------|------------|-----|---------------|-----------------|
| T4 Hindi-only | 2.00 | +0.30 | 2.30 | **2.00** | Ceiling `min(2.00, 2.30)` |
| T1 English-only | 1.10 | −0.15 | 0.95 | **1.05** | Floor `max(1.05, 0.95)` |

**The decay curve** (`poseidon.rs::cliff_decay_value`):
- Weeks 0–4 (cliff): full `m_szn_initial` value
- Weeks 4–12 (decay): linear drop from `m_szn_initial` to 1.0
- Weeks 12+: returns to 1.0

**Key insight for public-facing answers:** Saying "Season 1 multipliers range from 1.05x to 2.00x" is correct at both bounds. The overall multiplier floor of 1.0 is the baseline (no bonus), so the bonus range is effectively 1.05–2.00.

## Investigation flow

1. **Search broad** → `search_files(pattern="multiplier|reward|season|language")` to see all files that mention the concept
2. **Read the map** → `read_file(AGENTS.md)` to understand where things live
3. **Drill into services** → read the relevant service files in `apps/api/src/services/`
4. **Cross-ref design specs** → check `docs/superpowers/specs/` and `docs/rewards/` for the intended design
5. **Check migrations** → `apps/api/migrations/` to see the actual DB schema and seed data
6. **Check config** → `apps/api/config/` for environment-specific overrides

## Repo access

The repo lives at `https://github.com/piplabs/depin-backend`. Prefer HTTPS clone (`git clone https://...`) over SSH — the SSH key is not always available in the agent environment. Base branch is `staging`, not `main`.
