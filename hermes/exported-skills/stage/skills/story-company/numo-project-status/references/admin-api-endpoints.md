# Admin API Endpoints Reference

Production API endpoints for live Numo/depin stats. All require admin read auth.

## Auth

| Env Var | Purpose | Notes |
|---|---|---|
| `DEPIN_ADMIN_READ_API_KEY` | Credential value | NOT `INTERNAL_ADMIN_READ_API_KEY` |
| `DEPIN_ADMIN_READ_API_KEY_HEADER` | HTTP header name | e.g. `x-admin-read-key` |
| `DEPIN_ADMIN_BASE_URL` | Base URL | e.g. `https://depin.storyprotocol.net` |

Standard curl pattern:
```bash
curl -s \
  -H "${DEPIN_ADMIN_READ_API_KEY_HEADER}: ${DEPIN_ADMIN_READ_API_KEY}" \
  "${DEPIN_ADMIN_BASE_URL}/v1/admin/<endpoint>"
```

## Endpoints

### `/v1/admin/overview` — Cumulative stats

```json
{
  "total_users": 18480,
  "active_users_7d": 16359,
  "active_campaigns": 4,
  "featured_campaigns": 0,
  "submissions_24h": 46446,
  "submissions_7d": 220835,
  "submissions_7d_delta_pct": 181363.8,
  "pending_review_count": 210651,
  "recent_activity": [...]
}
```

**PITFALL**: `submissions_7d_delta_pct` can show absurd values (181K%) when growing from near-zero prior period. Don't report this number directly — compute the delta yourself or describe the trend in words.

### `/v1/admin/stats/submissions` — Daily submission counts

```json
[
  {"date": "2026-05-05", "count": 38786},
  {"date": "2026-05-06", "count": 40776},
  {"date": "2026-05-07", "count": 30638}
]
```

Returns the last several days. The most recent day is incomplete (in-progress).

### `/v1/admin/stats/user-growth` — Daily new user counts

```json
[
  {"date": "2026-05-05", "count": 962},
  {"date": "2026-05-06", "count": 512},
  {"date": "2026-05-07", "count": 204}
]
```

Same shape as submissions. Most recent day is incomplete.

### `/v1/admin/rewards/summary` — Payout state

```json
{
  "advances_paid_usd": "3110.5800",
  "remainders_paid_usd": "0.6120",
  "bonuses_paid_usd": "0.0180",
  "advances_forfeited_usd": "0.1620",
  "pending_advances_usd": "3107.3940",
  "pending_remainders_usd": "17608.5660",
  "pending_bonuses_estimated_usd": "492.4860",
  "pending_users_avg_multiplier": "1.0235",
  "counts": {
    "pending_review": 172633,
    "accepted": 6,
    "rejected": 9
  },
  "as_of": "2026-05-06T16:46:57.319015Z"
}
```

Monetary values are strings (decimals). `pending_*` fields represent money NOT yet paid out — blocked on review.

## Quick extraction with Python

```bash
# Get just the numbers you need
curl -s -H "${DEPIN_ADMIN_READ_API_KEY_HEADER}: ${DEPIN_ADMIN_READ_API_KEY}" \
  "${DEPIN_ADMIN_BASE_URL}/v1/admin/overview" | \
  python3 -c "import json,sys; d=json.load(sys.stdin); print(json.dumps({k:d[k] for k in ['total_users','submissions_24h','submissions_7d','pending_review_count']}, indent=2))"
```

## When to use

| Question type | Use |
|---|---|
| "What's the latest number of submissions?" | Just curl overview + stats/submissions |
| "How many users do we have?" | Just curl overview + stats/user-growth |
| "What's the payout status?" | Just curl rewards/summary |
| "How is the project coming along this week?" | Full 6-source investigation — admin API is only one source |
| "Show me the growth trend" | Curl stats/submissions + stats/user-growth, analyze the daily series |
