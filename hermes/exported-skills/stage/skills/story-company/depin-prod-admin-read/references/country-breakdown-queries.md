# Country Breakdown Queries (depin-prod)

Reusable SQL patterns for country-by-campaign analytics. All queries target `depin-prod`.

---

## Query 1: Self-Reported Country by Campaign

```sql
WITH campaign_totals AS (
  SELECT campaign_id, count(*) AS total
  FROM submissions
  GROUP BY campaign_id
),
country_counts AS (
  SELECT s.campaign_id, u.nationality AS country, count(*) AS submission_count
  FROM submissions s
  JOIN users u ON s.user_id = u.id
  WHERE u.nationality IS NOT NULL
  GROUP BY s.campaign_id, u.nationality
)
SELECT cc.campaign_id, c.campaign_name, cc.country,
       cc.submission_count, ct.total AS campaign_total,
       round(100.0 * cc.submission_count / ct.total, 2) AS pct
FROM country_counts cc
JOIN campaigns c ON cc.campaign_id = c.id
JOIN campaign_totals ct ON cc.campaign_id = ct.campaign_id
ORDER BY cc.campaign_id, cc.submission_count DESC
LIMIT 100;  -- paginate: OFFSET 100 for remaining rows
```

**Coverage**: All 298K+ submissions (every user has nationality set — 0 NULLs observed).
**Trust**: Low. Users can claim any nationality.

---

## Query 2: Castle IP Country by Campaign (Preferred)

```sql
WITH user_castle_country AS (
  SELECT DISTINCT ON (user_id) user_id, ip_country_code AS castle_country
  FROM castle_risk_events
  WHERE ip_country_code IS NOT NULL AND ip_country_code != ''
  ORDER BY user_id, created_at DESC
),
submission_country AS (
  SELECT s.id AS submission_id, s.campaign_id, s.user_id,
    COALESCE(uc.castle_country, u.nationality, 'UNKNOWN') AS effective_country,
    CASE WHEN uc.castle_country IS NOT NULL THEN 'castle'
         WHEN u.nationality IS NOT NULL THEN 'self_reported'
         ELSE 'unknown' END AS country_source
  FROM submissions s
  JOIN users u ON s.user_id = u.id
  LEFT JOIN user_castle_country uc ON u.id = uc.user_id
),
campaign_totals AS (
  SELECT campaign_id, count(*) AS total FROM submission_country GROUP BY campaign_id
),
country_counts AS (
  SELECT campaign_id, effective_country, count(*) AS submission_count
  FROM submission_country
  GROUP BY campaign_id, effective_country
)
SELECT cc.campaign_id, c.campaign_name, cc.effective_country AS country,
       cc.submission_count, ct.total AS campaign_total,
       round(100.0 * cc.submission_count / ct.total, 2) AS pct
FROM country_counts cc
JOIN campaigns c ON cc.campaign_id = c.id
JOIN campaign_totals ct ON cc.campaign_id = ct.campaign_id
ORDER BY cc.campaign_id, cc.submission_count DESC
LIMIT 100;
```

**Coverage**: ~9,400 users have Castle data; remainder fall back to self-reported.
**Trust**: Higher for Castle-covered users; self-reported fallback has same trust issues.
**Caveat**: Biased toward users who triggered Castle risk events (potentially more suspicious).

---

## Query 3: Unique Users vs Submission Count

```sql
SELECT u.nationality,
       count(DISTINCT s.user_id) AS unique_users,
       count(s.id) AS submission_count,
       round(count(s.id)::numeric / count(DISTINCT s.user_id), 1) AS subs_per_user
FROM submissions s
JOIN users u ON s.user_id = u.id
WHERE u.nationality IS NOT NULL
GROUP BY u.nationality
ORDER BY unique_users DESC;
```

**Interpretation**:
- `subs_per_user` < 5: likely organic / broad-but-shallow (e.g., Nigeria ~3.4)
- `subs_per_user` > 50: likely farming / power users (e.g., Poland ~103.6, Germany ~101.8)
- Big divergence between rank-by-subs and rank-by-users signals misrepresentation

---

## Query 4: Castle vs Self-Reported Mismatch Analysis

```sql
WITH user_castle_country AS (
  SELECT DISTINCT ON (user_id) user_id, ip_country_code AS castle_country
  FROM castle_risk_events
  WHERE ip_country_code IS NOT NULL AND ip_country_code != ''
  ORDER BY user_id, created_at DESC
)
SELECT uc.castle_country, u.nationality AS self_reported, count(*) AS user_count
FROM users u
JOIN user_castle_country uc ON u.id = uc.user_id
WHERE uc.castle_country != u.nationality AND u.nationality IS NOT NULL
GROUP BY uc.castle_country, u.nationality
ORDER BY user_count DESC
LIMIT 30;
```

---

## Query 5: Multi-Campaign Breadth (countries in ≥N campaigns)

```sql
SELECT u.nationality,
       count(DISTINCT s.campaign_id) AS campaign_count,
       count(DISTINCT s.user_id) AS unique_users,
       count(s.id) AS submission_count,
       round(count(s.id)::numeric / count(DISTINCT s.user_id), 1) AS subs_per_user
FROM submissions s
JOIN users u ON s.user_id = u.id
WHERE u.nationality IS NOT NULL
GROUP BY u.nationality
HAVING count(DISTINCT s.campaign_id) >= 3
ORDER BY campaign_count DESC, unique_users DESC
LIMIT 20;
```

---

## Query 6: Table Discovery & Size Estimation

Before writing complex queries, check counts and column names:

```sql
-- Discover all public tables
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY table_name;

-- Inspect columns for key tables
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name IN ('submissions', 'users', 'campaigns', 'castle_risk_events')
ORDER BY table_name, ordinal_position;

-- Size estimation
SELECT
  (SELECT count(*) FROM submissions) AS total_submissions,
  (SELECT count(*) FROM campaigns) AS total_campaigns,
  (SELECT count(DISTINCT nationality) FROM users WHERE nationality IS NOT NULL) AS distinct_nationalities;
```

---

## Case Study: Nigeria (2026-05-08)

**Self-reported**: 755 submissions, 219 unique users — rank #16 by volume, #4 by unique users.
**Castle IP**: 819 unique users with NG IPs — ~600 users with Nigerian IPs did NOT self-report as Nigerian.
**Subs per user**: 3.4 (lowest among top-10 countries by unique users).

Nigeria appears in all 5 active campaigns — one of only ~20 countries with this breadth. The high unique-user count with low submission volume suggests real-but-casual participants, not farming.

Compare: Poland has 2,589 submissions but only 25 unique users (103.6 subs/user) — concentrated farming pattern. These are fundamentally different signals that submission count alone obscures.

---

## Campaign Reference (2026-05-08)

| Campaign ID | Name | Active | Language | Submissions |
|-------------|------|--------|----------|-------------|
| 0a7932b7 | Hindi Voice Data Collection | yes | hi | 111,375 |
| b206aea5 | Bengali Voice Data Collection | yes | bn | 93,399 |
| 419046ee | Telugu Voice Data Collection | yes | te | 42,188 |
| d1d020d4 | Tamil Voice Data Collection | yes | ta | 38,134 |
| eed9e514 | Vietnamese Voice Data Collection | yes | vi | 13,782 |
| 18225d1d | Quartz Wave Society | no | fil | 0 |
| 66fcd877 | English Voice Data Collection | no | en | 0 |
| b1a81829 | Filipino Voice Data Collection | no | fil | 0 |
| fcd8f483 | Korean Voice Data Collection | no | ko | 0 |
