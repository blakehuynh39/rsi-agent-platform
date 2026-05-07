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

### Config and environment

| What to find | Where to look |
|---|---|
| Config hierarchy | `apps/api/config/{base,local,staging,production}.toml` |
| Config resolution order | `AGENTS.md` "Config model" section |
| Reward config table schema | `apps/api/migrations/0021_reward_system.sql` |

## Investigation flow

1. **Search broad** → `search_files(pattern="multiplier|reward|season|language")` to see all files that mention the concept
2. **Read the map** → `read_file(AGENTS.md)` to understand where things live
3. **Drill into services** → read the relevant service files in `apps/api/src/services/`
4. **Cross-ref design specs** → check `docs/superpowers/specs/` and `docs/rewards/` for the intended design
5. **Check migrations** → `apps/api/migrations/` to see the actual DB schema and seed data
6. **Check config** → `apps/api/config/` for environment-specific overrides

## Repo access

The repo lives at `https://github.com/piplabs/depin-backend`. Prefer HTTPS clone (`git clone https://...`) over SSH — the SSH key is not always available in the agent environment. Base branch is `staging`, not `main`.
