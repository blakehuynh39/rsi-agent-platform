# API Latency Investigation via Observability

## When to Use

Use this reference when investigating why a specific API endpoint has high latency.
It extends Phase 1 (Root Cause Investigation) of the systematic-debugging skill
for the specific case of observable API performance problems.

## Pre-requisites

- Prometheus/Thanos (or equivalent) with `http_request_duration_seconds` histograms
- Access to the service's source code (GitHub API, local clone)
- Access to the service's database schema (if DB queries are suspected)
- External call metrics (`ext_request_duration_seconds`) if the service instruments them

## Investigation Workflow

### Step 1: Quantify the problem

```promql
# Overall endpoint latency distribution (24h)
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path="/v1/endpoint"}[24h])) by (le))

# Break down by environment
histogram_quantile(0.95, sum by (le, job) (rate(http_request_duration_seconds_bucket{path="/v1/endpoint"}[24h])))

# Compare p50 / p95 / p99 to see tail behavior
histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{path="/v1/endpoint"}[1h])) by (le))
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{path="/v1/endpoint"}[1h])) by (le))
```

**Interpretation:**
- p50 ≈ p95 ≈ p99 → consistent latency, check for blocking external calls or DB scans
- p50 low, p95/p99 high → long tail, bursty external calls or cache misses
- Large difference between 1h and 24h p95 → intermittent issue, check deployment windows

### Step 2: Establish a baseline

Compare the slow endpoint against similar endpoints:

```promql
# All endpoints ranked by p95
histogram_quantile(0.95, sum by (le, path) (rate(http_request_duration_seconds_bucket[24h])))

# Request volume
sum by (path) (increase(http_requests_total[24h]))
```

A healthy endpoint serving similar data should have p95 within 2-3x of the fastest
endpoints. If the endpoint is 10x+ slower, there's likely an external call, a missing
index, or an N+1 query problem.

### Step 3: Check external call metrics

```promql
# List all external call metrics to see what's instrumented
ext_request_duration_seconds_count

# Get p95 latency per external service/endpoint
histogram_quantile(0.95, sum by (le, exported_service, exported_endpoint) (rate(ext_request_duration_seconds_bucket[24h])))

# Check error rates
sum by (exported_service, exported_endpoint, status) (increase(ext_request_duration_seconds_count[24h]))
```

**Critical label quirk:** The Prometheus `prometheus-net`/K8s annotation layer renames
`service` → `exported_service` and `endpoint` → `exported_endpoint` in the scraped
metrics. Always query with `exported_service` and `exported_endpoint`, not the raw
label names from the code.

**Pitfall:** Not all external calls are instrumented. Always cross-reference with
the source code (Step 4) to check if every HTTP client call has a matching
`track_external_request()` invocation. Missing instrumentation means that vendor's
latency contribution is invisible — you'll need to infer it from total endpoint
latency minus known components, or add instrumentation.

### Step 4: Trace the code path

Use the GitHub API (or local clone) to read the handler and trace every function call:

```
# Find the route registration
GET /repos/{owner}/{repo}/contents/apps/api/src/http/routes/{module}.rs

# Read the handler function body
# Trace into the service layer
GET /repos/{owner}/{repo}/contents/apps/api/src/services/{module}.rs

# Check integrations for external API clients
GET /repos/{owner}/{repo}/contents/apps/api/src/integrations/{vendor}.rs
```

Map each step in the handler to a latency source:
- **Local operations** (JWT verification, serialization): <1ms typically
- **Database queries**: 1-10ms with proper indexes, 100ms+ if scanning
- **External API calls**: variable, check `ext_request_duration_seconds`
- **Cache operations**: <1ms normally, variable on miss

### Step 5: Verify database indexes

If the code path includes database queries, verify each one hits an index:

```sql
-- List indexes on the relevant table
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = '{table}' ORDER BY indexname;

-- Check for sequential scans on the table (if you have pg_stat access)
SELECT seq_scan, seq_tup_read, idx_scan, idx_tup_fetch
FROM pg_stat_user_tables WHERE relname = '{table}';
```

Match each query in the code path against the available indexes:
- `WHERE id = $1` → needs PK or unique index on `id`
- `WHERE dynamic_user_id = $1` → needs index on `dynamic_user_id`
- `WHERE LOWER(email) = $1` → needs functional index on `lower(email)`
- `INSERT ... ON CONFLICT` → needs the unique constraint/index referenced in ON CONFLICT

### Step 6: Formulate the answer

The answer to "is this external vendors or missing indexes?" should be quantitative:

- "External call X takes Yms and is called on Z% of requests"
- "Query Q uses index I confirmed present"
- "External call E has zero instrumentation — latency contribution unknown"
- "No sequential scans on tables touched by this endpoint"

### Common Failure Patterns

| Pattern | Symptom | Fix |
|---------|---------|-----|
| External API returning 4xx/5xx | High p95 even on "fast" external calls (error responses are often slow) | Fix credentials or handle errors gracefully |
| Missing `track_external_request` | External call latency invisible, p95 higher than sum of known components | Add instrumentation |
| Cache miss on first request after deploy | High 24h p95 but low 1h p95 (cache is warm now) | Pre-warm cache or accept cold-start latency |
| N+1 in loop | p95 grows with data volume | Batch query or add eager loading |
| Sequential scan on large table | High p50 (consistent slowness, not just tail) | Add index matching the WHERE clause |
