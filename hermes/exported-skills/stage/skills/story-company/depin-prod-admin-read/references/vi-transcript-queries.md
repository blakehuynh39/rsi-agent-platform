# Vietnamese Transcript Distribution Queries

Session: 2026-05-08 — per-script unique user count distribution for Vietnamese transcripts.

## Schema (Joins Required)

Per-script unique user counts require a three-table LEFT JOIN because submissions link through `script_assignments`, not directly to scripts:

```
scripts (id, language_code, is_active)
  └── script_assignments (script_id, user_id)
        └── submissions (script_assignment_id, user_id)
```

## Query 1: Per-Script Unique User Counts

```sql
SELECT s.id, COUNT(DISTINCT sub.user_id) AS unique_users
FROM scripts s
LEFT JOIN script_assignments sa ON sa.script_id = s.id
LEFT JOIN submissions sub ON sub.script_assignment_id = sa.id
WHERE s.language_code = 'vi' AND s.is_active = true
GROUP BY s.id
ORDER BY unique_users DESC
LIMIT 100
```

**Note**: LIMIT 100 is the prod cap. For the full distribution, use a histogram query (Query 3).

## Query 2: Count of Active Scripts

```sql
SELECT COUNT(*) AS vi_scripts
FROM scripts
WHERE language_code = 'vi' AND is_active = true
```

## Query 3: Histogram (scripts per unique_user bucket)

```sql
SELECT unique_users, COUNT(*) AS script_count
FROM (
  SELECT s.id, COUNT(DISTINCT sub.user_id) AS unique_users
  FROM scripts s
  LEFT JOIN script_assignments sa ON sa.script_id = s.id
  LEFT JOIN submissions sub ON sub.script_assignment_id = sa.id
  WHERE s.language_code = 'vi' AND s.is_active = true
  GROUP BY s.id
) AS per_script
GROUP BY unique_users
ORDER BY unique_users DESC
```

This returns the full distribution in a few rows (one per distinct unique_user count) and fits well within the 100-row cap.

## Query 4: Per-Campaign Histogram (All Active Campaigns)

Generalizes Query 3 to all active campaigns by adding `campaign_id` to the inner grouping and joining to `campaigns` in the outer query:

```sql
SELECT 
    c.id AS campaign_id,
    c.campaign_name,
    c.campaign_type,
    unique_users,
    COUNT(*) AS script_count
FROM (
    SELECT 
        s.campaign_id,
        s.id AS script_id,
        COUNT(DISTINCT sub.user_id) AS unique_users
    FROM scripts s
    LEFT JOIN script_assignments sa ON sa.script_id = s.id
    LEFT JOIN submissions sub ON sub.script_assignment_id = sa.id
    WHERE s.is_active = true
    GROUP BY s.campaign_id, s.id
) AS per_script
JOIN campaigns c ON c.id = per_script.campaign_id
WHERE c.is_active = true
GROUP BY c.id, c.campaign_name, c.campaign_type, unique_users
ORDER BY c.campaign_name, unique_users DESC
```

**Pitfall**: The campaigns table uses `campaign_name` and `campaign_type`, not `name` and `type`. Always verify with `information_schema.columns` before joining.

## Findings (2026-05-08)

### Stage (depin-stage)
- 73 active Vietnamese scripts
- 64 scripts (87.7%): 0 unique submitters
- 9 scripts (12.3%): 1 unique submitter each

### Prod (depin-prod — 2026-05-08 08:54 UTC)
- Distribution query returned **10 rows** (histogram buckets), truncated=false
- REST API: 887 users with `primary_language = 'vi'`, 7,744 submissions, avg 8.73/user
- Campaign: 8,981 completed / 36,000 target tasks

### Prod (depin-prod — 2026-05-08 ~09:00 UTC, second session)
- **2,000** active Vietnamese scripts (up from ~1,000 in earlier session)
- Distribution spans **0–12 unique users** (prior cap of 9 appears raised or batch-refreshed)
- **903 scripts (45.1%)** in the 3–4 user bucket — dominant cluster
- **519 scripts (25.9%)** at 10+ users — near/at new cap of 12
- **480 scripts (24.0%)** in the 6–9 user range
- Only **98 scripts (4.9%)** underutilized (0–2 users)
- 12,815 total unique user→script submission pairs
- Distribution has diversified significantly compared to the earlier May 8 session

### Cross-Validation with REST API

Admin REST `/v1/admin/cohorts/languages?range=all`:
- 849 users with `primary_language = 'vi'`
- 7,078 total submissions
- 849 × 9 = 7,641 ≈ 7,078 — validates the 9-user cap hypothesis

Campaign: "Vietnamese Voice Data Collection" (`eed9e514`):
- 1,225 participants, 8,099 completed tasks

### Key Insight

The per-script cap of 9 unique users, combined with ~1,000 scripts and 849 users, explains the submission volume almost exactly. Each user can submit to at most 9 different scripts (or each script accepts at most 9 users), producing 7,078 submissions from a theoretical max of ~7,641.

### All-Campaign Distribution (2026-05-08 ~21:00 UTC, depin-prod)

Query 4 returned 25 rows covering all 5 active campaigns — all `campaign_type = 'voice'`:

| Campaign | Active Scripts | Scripts Used (≥1) | Utilization | Total User-Script Pairs |
|---|---|---|---|---|
| Vietnamese | 2,000 | 1,999 | 99.95% | 12,815 |
| Hindi | 350,000 | 103,061 | 29.4% | 103,084 |
| Bengali | 350,000 | 85,569 | 24.4% | 85,583 |
| Telugu | 350,000 | 39,636 | 11.3% | 39,649 |
| Tamil | 350,000 | 34,474 | 9.8% | 34,487 |

**Vietnamese** (2,000 scripts): Rich distribution centered at 3–4 unique users (mode: 4, 27.5%), range 0–12. Nearly saturated — only 1 script with 0 users. 12,815 total user-script pairs.

**Hindi / Bengali / Telugu / Tamil** (350,000 scripts each): Massively front-loaded. 70.6–90.2% of scripts have 0 unique users. Nearly all used scripts have exactly 1 unique user. Only 13–23 scripts per campaign have reached 2 users. These pools were bulk-loaded for future scale.

Key takeaways:
- Vietnamese is a mature, saturated campaign — users are cycling through the per-script cap
- The other four campaigns have enormous headroom — their 350K script pools are barely touched
- The `campaigns` table uses `campaign_name` and `campaign_type`, not `name`/`type` as documentation suggests
