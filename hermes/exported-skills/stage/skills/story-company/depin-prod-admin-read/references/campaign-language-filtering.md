# Campaign-Based Language Filtering

When the user asks "how many submissions are in language X" and `/v1/admin/cohorts/languages` doesn't suffice (or returns unexpected results because it uses `users.primary_language` instead of campaign language), fall back to campaign-based enumeration.

## When to use this

- `/v1/admin/cohorts/languages` uses `users.primary_language` — campaign-unaware
- The user wants campaign-scoped counts (e.g., "Vietnamese Voice Data Collection submissions")
- The user wants counts by a specific dimension not in cohorts (state, validation_decision)

## Step-by-step

### 1. Find the campaign by language

```bash
curl -s -H "${DEPIN_ADMIN_READ_API_KEY_HEADER}: ${DEPIN_ADMIN_READ_API_KEY}" \
  "${DEPIN_ADMIN_BASE_URL}/v1/admin/campaigns" \
  | python3 -c "
import json, sys
for c in json.load(sys.stdin).get('items', []):
    if 'vi' in c.get('supported_languages', []):
        print(c['id'], c['campaign_name'], c.get('participant_count'), c.get('completed_tasks'))
"
```

The campaign list already provides `participant_count` and `completed_tasks` — these are quick approximations that may satisfy the user without full pagination.

### 2. Full enumeration script

```bash
#!/bin/bash
# Count submissions for a campaign with state breakdown
CAMPAIGN_ID="eed9e514-efb7-4477-a81b-6af6d328e889"
count=0
cursor=""
pages=0

while [ $pages -lt 100 ]; do
  if [ -z "$cursor" ]; then
    url="${DEPIN_ADMIN_BASE_URL}/v1/admin/submissions?campaign_id=${CAMPAIGN_ID}&limit=200"
  else
    url="${DEPIN_ADMIN_BASE_URL}/v1/admin/submissions?campaign_id=${CAMPAIGN_ID}&limit=200&cursor=${cursor}"
  fi
  
  resp=$(curl -s --max-time 30 \
    -H "${DEPIN_ADMIN_READ_API_KEY_HEADER}: ${DEPIN_ADMIN_READ_API_KEY}" "$url")
  
  n=$(echo "$resp" | python3 -c "import json,sys; print(len(json.load(sys.stdin).get('items',[])))")
  cursor=$(echo "$resp" | python3 -c "import json,sys; print(json.load(sys.stdin).get('next_cursor',''))")
  
  count=$((count + n))
  pages=$((pages + 1))
  
  [ $((pages % 10)) -eq 0 ] && echo "Page $pages: +$n = $count total"
  [ -z "$cursor" ] || [ "$n" -eq 0 ] && break
done

echo "Total: $count submissions in $pages pages"
```

### 3. Performance notes

- API caps at ~200 items per page regardless of `limit` parameter
- 4,840 submissions = 26 pages ≈ 30-45 seconds
- For larger campaigns (10K+), expect 50+ pages and 2+ minutes
- Use `cohorts/languages` whenever possible — it's a single request

## Verified example (2026-05-07)

Vietnamese Voice Data Collection campaign (`eed9e514`):
- Campaign `completed_tasks`: 4,464
- Actual submissions (full pagination): **4,840**
- Difference: 376 — likely submissions in non-"created" states or retries
- 26 pages × ~200 items/page
