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

### ⚠️ Pitfall: 3-Table JOINs Time Out on depin-prod

Query 3 (scripts→script_assignments→submissions) may exceed the 20s timeout on `depin-prod` because it forces a triple-table hash join. When it times out, use the **two-path strategy**:

- **Path A (assignment-level)**: Query 5 — skips `submissions`, joins only `scripts`→`script_assignments`. Counts `COUNT(DISTINCT sa.user_id)`. Always fast (2-table join). Answers "how many unique users were assigned to each script?"
- **Path B (submission-level)**: Query 6 — joins `script_assignments`→`submissions` directly (no `scripts` table in the join, filter by `s.language_code` via `scripts` in WHERE subquery or use `sa.script_id` + `scripts` as separate filter). Answers "how many unique users actually submitted to each script?"

The two distributions tell different stories. The assignment distribution reveals the provisioning strategy (bimodal at 7 and 14-15); the submission distribution reveals actual user behavior (bell-shaped centered at 7-8). The gap between them — e.g., 3,629 assigned vs 1,890 submitted (52% conversion) — is a key diagnostic for campaign health.

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

## Query 5: Assignment-Only Histogram (No Submissions Table)

```sql
SELECT unique_users, COUNT(*) AS script_count
FROM (
  SELECT s.id, COUNT(DISTINCT sa.user_id) AS unique_users
  FROM scripts s
  LEFT JOIN script_assignments sa ON sa.script_id = s.id
  WHERE s.language_code = 'vi' AND s.is_active = true
  GROUP BY s.id
) dist
GROUP BY unique_users
ORDER BY unique_users
```

**Use when**: Query 3 times out. This joins only 2 tables (`scripts`→`script_assignments`) and counts unique assigned users per script. Returns the assignment-level distribution — consistently fast even on depin-prod at 2,000+ scripts.

## Query 6: Submission-Only Histogram (Filter by Script Language)

```sql
SELECT unique_users, COUNT(*) AS script_count
FROM (
  SELECT sa.script_id, COUNT(DISTINCT sub.user_id) AS unique_users
  FROM scripts s
  JOIN script_assignments sa ON sa.script_id = s.id
  JOIN submissions sub ON sub.script_assignment_id = sa.id
  WHERE s.language_code = 'vi' AND s.is_active = true AND sub.mock = false
  GROUP BY sa.script_id
) dist
GROUP BY unique_users
ORDER BY unique_users
```

**Alternative (no scripts in inner join)**: Join `script_assignments`→`submissions` directly and filter by script ID from a subquery on `scripts`:

```sql
SELECT unique_users, COUNT(*) AS script_count
FROM (
  SELECT sa.script_id, COUNT(DISTINCT sub.user_id) AS unique_users
  FROM script_assignments sa
  JOIN submissions sub ON sub.script_assignment_id = sa.id
  WHERE sa.script_id IN (SELECT id FROM scripts WHERE language_code = 'vi' AND is_active = true)
    AND sub.mock = false
  GROUP BY sa.script_id
) dist
GROUP BY unique_users
ORDER BY unique_users
```

**Use when**: You need submission-level data but Query 3 times out. This still uses 3 tables but the inner subquery on `scripts` may execute as a separate index scan. If this also times out, use Query 5 and separately query total submission counts.

## Gap Analysis Query

To check how many assigned users actually submitted (per-script gap):

```sql
SELECT 
  COUNT(DISTINCT sa.user_id) AS assigned_users,
  COUNT(DISTINCT sub.user_id) AS submitting_users,
  COUNT(DISTINCT sa.user_id) - COUNT(DISTINCT sub.user_id) AS gap
FROM scripts s
JOIN script_assignments sa ON sa.script_id = s.id
LEFT JOIN submissions sub ON sub.script_assignment_id = sa.id AND sub.mock = false
WHERE s.language_code = 'vi' AND s.is_active = true
GROUP BY s.id
LIMIT 20
```

For the aggregate conversion rate:

```sql
SELECT 
  COUNT(DISTINCT sa.user_id) AS total_assigned,
  COUNT(DISTINCT sub.user_id) AS total_submitted
FROM scripts s
JOIN script_assignments sa ON sa.script_id = s.id
LEFT JOIN submissions sub ON sub.script_assignment_id = sa.id AND sub.mock = false
WHERE s.language_code = 'vi' AND s.is_active = true
```

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

### Prod (depin-prod — 2026-05-10 08:19 UTC)

- 2,000 active VI scripts, distribution range 3–17 unique submitters
- Bell-shaped distribution centered at 7–8 (mode: 8, ~390 scripts; median ~9)
- 1,890 unique users who submitted, 21,817 total submissions
- 3,629 unique users assigned → 48% drop-off (1,739 assigned users never submitted)
- Submission states: 20,580 `pending_review`, 1,243 `created` — 94% awaiting review

### Prod (depin-prod — 2026-05-11 03:07 UTC)

- 2,000 active VI scripts, distribution range **4–17** unique submitters (14 histogram buckets)
- **Bimodal distribution** — major structural shift from prior sessions:
  - Primary mode: **8** (390 scripts, 19.5%)
  - Heavy secondary cluster: **13–17** (812 scripts, 40.6%)
  - The "floor" rose to 4 — no scripts below 4 submitters, pool is fully utilized
- median: 9, mean: 10.77, total user-script pairs: 21,544
- REST API cross-validation: 1,239 VI-language users, 22,271 submissions, avg 17.97/user
- DB read request: `dbread_ab06e84e`, 14 rows, truncated=false, approved and executed synchronously

**3-day evolution (May 8 → May 10 → May 11):**

| Date | Range | Mode | Shape | Key Change |
|------|-------|------|-------|-------------|
| May 8 | 0–12 | 3–4 | Unimodal, tight | 903 scripts (45%) at 3–4 |
| May 10 | 3–17 | 7–8 | Bell-shaped | Broader, moved right |
| May 11 | 4–17 | 8 + 13–15 cluster | **Bimodal** | 40.6% scripts at 13+; floor at 4 |

Interpretation: The bimodal pattern suggests two distinct script cohorts — one cycling at the ~8-user equilibrium from May 10, and a newer batch aggressively filled to 13–17 users close to the apparent ceiling. 40.6% of scripts in the Very High segment (13–17) indicates either a cap raise or campaign batch expansion. The disappearance of 0–3 user buckets means zero idle capacity — every active VI script has at least 4 submitters.

### Prod (depin-prod — 2026-05-10 08:28 UTC)

**Assignment-level distribution** (Query 5 — `script_assignments` only):
- 2,000 active Vietnamese scripts (1 campaign: "Vietnamese Voice Data Collection")
- 21,525 total script assignments, 3,629 unique assigned users
- Avg 10.76 unique users assigned per script
- **Bimodal**: clusters at 7 users (990 scripts, 49.5%) and 14-15 users (980 scripts, 49.0%)
- Range: 6–16 assigned users

**Submission-level distribution** (Query 6 — through `submissions`):
- 2,000 scripts with submissions
- 21,817 total submissions, 1,890 unique submitters
- **Bell-shaped**: centered at 7-8 users, range 3–17
- States: 20,580 `pending_review` (94.3%), 1,243 `created` (5.7%)

**Assignment→Submission gap**:
- 3,629 assigned → 1,890 submitted = **52.1% conversion**
- Scripts at 14-15 assigned tier have only 10-13 actual submitters (~20% drop-off)
- The bimodal assignment pattern collapses into a single bell curve at submission level

**vs May 8**:
- Scripts: 2,000 (unchanged)
- User→script pairs: 12,815 → 21,525 (+68%)
- Range: 0-12 → 6-16 (floor raised, cap raised)
- Dominant cluster: 3-4 users → 7 users (right-shifted)
- High-usage: 10-12 (26%) → 14-15 (49%) — doubled

**Key insight**: The 3-table JOIN (Query 3) timed out on depin-prod. All findings use the two-path strategy (Query 5 for assignments, Query 6 for submissions). The assignment bimodal pattern suggests two assignment waves; the submission bell curve reflects natural user behavior. The gap between them (48% non-submission) is a campaign health signal.
