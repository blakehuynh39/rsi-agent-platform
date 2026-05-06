# Grafana/Thanos Metrics Reference

Query reference and pitfalls for Grafana → Thanos (Prometheus) metrics used in depin-backend investigations.

## Access Setup

| Variable | Value | Notes |
|---|---|---|
| `GRAFANA_SERVER` | `https://grafana.ops.storyprotocol.net` | Grafana instance (v12.3.1) |
| `GRAFANA_TOKEN` | Service account token (prefix `glsa_`) | Works for `/api/health`, `/api/search`, `/api/dashboards/uid/*` |
| `RSI_GRAFANA_CF_ACCESS_CLIENT_ID` | Cloudflare Access client ID | Required for datasource proxy |
| `RSI_GRAFANA_CF_ACCESS_CLIENT_SECRET` | Cloudflare Access client secret | Required for datasource proxy |
| Datasource UID | `thanos` | Type: prometheus, proxy access, backend `http://thanos-query:9090` |

**Critical**: `GRAFANA_TOKEN` alone works for ALL API endpoints including the datasource proxy `/api/datasources/proxy/uid/thanos/...`. The CF-Access headers are NOT required (verified 2026-05-06). If the proxy returns 401, the token itself is expired/insufficient — check permissions, not CF Access.

## Metric Labels Reference

### Stage
- `job`: `use1-stage-depin-backend`
- `namespace`: `story`
- `cluster`: `use1-stage`
- `environment`: `stage`
- `prometheus_replica`: `prometheus-prometheus-kube-prometheus-prometheus-0`

### Prod
- `job`: `use1-prod-depin-backend`
- `namespace`: `story` (in Thanos kube-state-metrics — `container_*` metrics for prod pods also appear under `namespace="story"` despite prod pods NOT being in the K8s `story` namespace. This is a quirk of how Thanos federates metrics. **Always confirm with pod name prefix** — `use1-prod-*` vs `use1-stage-*`.)
- `cluster`: `use1-prod`
- `environment`: `prod`
- Note: Prod pods are NOT accessible via `kubectl` from the stage cluster. Use Thanos metrics exclusively for prod pod data.

### IP Registration Jobs (stage only)
- `job`: `use1-stage-depin-ip-registration-confirmer`, `use1-stage-depin-ip-registration-poller`, `use1-stage-depin-ip-registration-submitter`
- These jobs do NOT appear to expose HTTP metrics (`http_requests_total`). Use `container_cpu_usage_seconds_total` and `container_memory_working_set_bytes` for resource monitoring.

## Canonical Query Pattern

```bash
ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('<promql>'))")
curl -s \
  -H "Authorization: Bearer ${GRAFANA_TOKEN}" \
  "${GRAFANA_SERVER}/api/datasources/proxy/uid/thanos/api/v1/query?query=${ENCODED}"
```

Parse results with: `python3 -c "import json,sys; d=json.load(sys.stdin); ..."`

### Range Queries

```bash
NOW=$(date +%s)
DAY_AGO=$((NOW - 86400))
curl -s \
  -H "Authorization: Bearer ${GRAFANA_TOKEN}" \
  "${GRAFANA_SERVER}/api/datasources/proxy/uid/thanos/api/v1/query_range?query=${ENCODED}&start=${DAY_AGO}&end=${NOW}&step=1800"
```

## Essential PromQL Queries

### Throughput & Errors

```promql
# Request rate by status code (stage)
sum by (status) (rate(http_requests_total{job="use1-stage-depin-backend"}[1h]))

# Request rate by status code (prod)
sum by (status) (rate(http_requests_total{job="use1-prod-depin-backend"}[1h]))

# Total requests in 24h
sum(increase(http_requests_total{job="use1-stage-depin-backend"}[24h]))
sum(increase(http_requests_total{job="use1-prod-depin-backend"}[24h]))

# Error breakdown by path and status (4xx/5xx only)
sum by (status, path) (rate(http_requests_total{job="use1-stage-depin-backend",status=~"4..|5.."}[1h]))

# Top 5 healthy endpoints (200 only, prod)
topk(5, sum by (path, method) (rate(http_requests_total{job="use1-prod-depin-backend",status="200"}[15m])))

# Active endpoints with non-zero traffic
sum by (path, method) (rate(http_requests_total{job="use1-stage-depin-backend"}[15m])) > 0
```

### Latency

```promql
# Overall latency percentiles (stage)
histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket{job="use1-stage-depin-backend"}[15m])))
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{job="use1-stage-depin-backend"}[15m])))
histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket{job="use1-stage-depin-backend"}[15m])))

# Latency by path (p95, stage)
histogram_quantile(0.95, sum by (le, path) (rate(http_request_duration_seconds_bucket{job="use1-stage-depin-backend"}[15m])))

# Prod latency
histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket{job="use1-prod-depin-backend"}[15m])))
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{job="use1-prod-depin-backend"}[15m])))
histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket{job="use1-prod-depin-backend"}[15m])))
```

### Resource Usage (CPU / Memory)

```promql
# CPU (millicores) by pod — works for both stage and prod
sum by (pod) (rate(container_cpu_usage_seconds_total{namespace="story",container="depin-backend"}[5m])) * 1000

# Memory working set (MB) by pod
sum by (pod) (container_memory_working_set_bytes{namespace="story",container="depin-backend"}) / 1024 / 1024

# Memory anomaly check — average over time
avg_over_time(container_memory_working_set_bytes{namespace="story",pod="use1-prod-depin-backend-7f6566fbd8-vzp5r",container="depin-backend"}[6h]) / 1024 / 1024

# CPU percentage for a specific pod
rate(container_cpu_usage_seconds_total{namespace="story",pod="use1-prod-depin-backend-7f6566fbd8-vzp5r",container="depin-backend"}[15m]) * 100
```

### IP Registration Pipeline Throughput

```promql
# Confirmed + reverted rate (the definitive throughput metric)
sum by (status) (rate(ip_registration_total{environment="prod",status=~"confirmed|reverted"}[5m]))

# Broadcast outcomes
sum by (status) (rate(ip_registration_total{environment="prod",status=~"broadcasted|broadcast_error"}[5m]))

# Open jobs by (action, state) — queue depth
sum by (state, action) (ip_registration_open_jobs_by_state{environment="prod"})

# In-flight tx attempts
sum by (state) (ip_registration_current_tx_attempts_by_state{environment="prod"})

# Stuck broadcasted jobs (age > STUCK_THRESHOLD, default 1800s)
sum by (action) (ip_registration_stuck_jobs{environment="prod",state="BROADCASTED"})

# Oldest BROADCASTED job age (seconds)
max by (action) (ip_registration_oldest_non_terminal_job_age_seconds{environment="prod",state="BROADCASTED"})

# Active wallets
ip_registration_wallets_active{environment="prod"}

# Mempool inclusion latency (broadcast → confirm) p50/p95
histogram_quantile(0.50, sum by (le) (rate(ip_registration_attempt_broadcast_to_confirm_seconds_bucket{environment="prod"}[15m])))
histogram_quantile(0.95, sum by (le) (rate(ip_registration_attempt_broadcast_to_confirm_seconds_bucket{environment="prod"}[15m])))

# Submit tx + confirm receipt duration
histogram_quantile(0.50, sum by (le) (rate(ip_registration_duration_seconds_bucket{environment="prod",stage="submit_tx"}[15m])))
histogram_quantile(0.50, sum by (le) (rate(ip_registration_duration_seconds_bucket{environment="prod",stage="confirm_receipt"}[15m])))

# Full attempt lifecycle (created → terminal)
histogram_quantile(0.50, sum by (le) (rate(ip_registration_attempt_lifecycle_seconds_bucket{environment="prod",terminal_state="confirmed"}[15m])))
```

**Key labels**: `environment="prod"` (NOT `"production"`), `action` (CREATE_COLLECTION/REGISTER_CAMPAIGN_IP/REGISTER_FILE_IP), `state` (READY_TO_SUBMIT/BROADCASTED), `status` (broadcasted/broadcast_error/confirmed/reverted).

**Important**: A single job increments multiple `ip_registration_total` counters (broadcasted → confirmed). The `confirmed` rate is the definitive throughput metric. The 24h `increase()` on this counter will be roughly 2× the actual number of unique jobs confirmed.

### Dashboard Discovery

```bash
# List depin dashboards
curl -s -H "Authorization: Bearer ${GRAFANA_TOKEN}" \
  "${GRAFANA_SERVER}/api/search?query=depin&type=dash-db"

# Get dashboard definition (to inspect panels)
curl -s -H "Authorization: Bearer ${GRAFANA_TOKEN}" \
  "${GRAFANA_SERVER}/api/dashboards/uid/depin-backend-api"

# Health check
curl -s -H "Authorization: Bearer ${GRAFANA_TOKEN}" \
  "${GRAFANA_SERVER}/api/health"
```

## Pitfalls

### 1. 404 Noise Masquerading as Errors
Stage typically shows ~89% "error rate" but 88.9% is `404 unmatched` — crawlers, health probes, and scanners hitting non-existent paths like `/wp-admin`, `/.env`, `/robots.txt`. The `path="unmatched"` label means the request didn't match any registered route.

**Fix**: Always exclude `path="unmatched"` or filter to `5xx` only when reporting "real errors":
```promql
# Real error rate (excluding unmatched)
sum(rate(http_requests_total{job="use1-stage-depin-backend",status=~"4..|5..",path!="unmatched"}[1h]))
  /
sum(rate(http_requests_total{job="use1-stage-depin-backend",path!="unmatched"}[1h]))
```

### 2. CF-Access Headers NOT Required (verified 2026-05-06)
Contrary to earlier documentation, `GRAFANA_TOKEN` works for ALL endpoints including the datasource proxy. CF-Access headers are not needed. If the proxy returns 401, the token permissions are insufficient — request a token with datasource query access.

### 3. Prod Pods Share `namespace="story"` in Thanos Container Metrics

Prod pods (`use1-prod-depin-backend-*`) are NOT in the K8s `story` namespace and cannot be accessed via `kubectl -n story`. However, in Thanos kube-state-metrics, `container_*` queries with `namespace="story"` return **both stage and prod** depin-backend pods. Always filter by pod name prefix to separate environments:

```promql
# Stage pods
sum by (pod) (container_memory_working_set_bytes{namespace="story",container="depin-backend",pod=~"use1-stage.*"})

# Prod pods — same namespace filter works, but use pod prefix
sum by (pod) (container_memory_working_set_bytes{namespace="story",container="depin-backend",pod=~"use1-prod.*"})
```

`kubectl get pods -n story` shows **stage pods only**. For prod pod restarts/status, you must use Thanos metrics (container restarts via `kube_pod_container_status_restarts_total`).

### 4. Memory Anomaly (jemalloc Arena Retention)
A prod pod may show 14× memory vs siblings (e.g., 570 MB vs 40 MB) without being a leak. Characteristics of jemalloc arena retention:
- Stable over hours (not growing)
- Low CPU usage (no active processing)
- No restarts (pod uptime matches siblings)
- Only one pod affected (not all)

This is normal behavior after heavy background job processing (multiplier sweep, hot path cache rebuild). If memory is growing linearly over time AND CPU is elevated, that's a real leak.

### 5. IP Registration Jobs Lack HTTP Metrics
The confirmer/poller/submitter pods don't expose `http_requests_total` or `http_request_duration_seconds_bucket`. Their health is only visible through container CPU/memory metrics and pod status.

### 6. Range Queries Require Epoch Timestamps
The `query_range` endpoint rejects human-readable time strings (`-6h`, `now`). You must use numeric Unix epoch timestamps.

```bash
# Correct
NOW=$(date +%s)
SIX_H_AGO=$((NOW - 21600))
step=900  # 15 minutes
curl -s ... "?query=...&start=${SIX_H_AGO}&end=${NOW}&step=${step}"

# WRONG — will return 400
curl -s ... "?query=...&start=-6h&end=now&step=15m"
```

Error you'll see: `{"status":"error","errorType":"bad_data","error":"cannot parse \"-6h\" to a valid timestamp"}`

### 7. NaN p95 Values on Low/Zero-Traffic Endpoints

When `histogram_quantile(0.95, sum by (le, path) (...))` returns `NaN` for specific paths, it means the endpoint had **no requests in the 15m window** (zero bucket counts). This is normal for admin-only or infrequently-called endpoints. Don't interpret as errors.

```promql
# Filter out NaN paths when reporting
histogram_quantile(0.95, sum by (le, path) (rate(http_request_duration_seconds_bucket{job="use1-prod-depin-backend"}[15m]))) != NaN
```

### 8. Histogram Bucket Floor Artifact (4.75ms)

Many low-traffic endpoints show exactly **4.75ms p95** — this is the smallest histogram bucket boundary, not real end-to-end latency. When an endpoint receives only 1-2 requests in a 15m window that all land in the smallest bucket (≤5ms), the quantile calculation returns the bucket floor. These endpoints have negligible real latency but the number is misleading. Only interpret p95 values as accurate for endpoints with sustained traffic.

**How to distinguish**: If the endpoint also shows 4.75ms at p50 and p99, it's the bucket floor. Real latency shows distinct values across percentiles.

### 9. Multiple ReplicaSets in `kubectl top` During Rolling Deploys

During a rolling deploy, `kubectl top pods` may show pods from both the old and new ReplicaSet simultaneously (e.g., 4 submitter pods when deployment says 3/3). Old pods are terminating — cross-reference with `kubectl get pods | grep` to see actual pod states (`Running` vs `Terminating`). Only count Running pods when assessing health.

### 10. `job` Label Unreliable for Prod `container_*` Metrics

Querying `container_memory_working_set_bytes{job="use1-prod-depin-backend",...}` returns **empty results** in Thanos. The `job` label is populated by the Prometheus scrape config and may not match the container metric source. **Always use `namespace="story"` for `container_*` queries** for both stage and prod pods, and filter by pod name prefix (`pod=~"use1-prod.*"` or `pod=~"use1-stage.*"`) to separate environments.

```promql
# ✅ Correct — prod pod memory
sum by (pod) (container_memory_working_set_bytes{namespace="story",container="depin-backend",pod=~"use1-prod.*"}) / 1024 / 1024

# ❌ Wrong — returns empty
sum by (pod) (container_memory_working_set_bytes{job="use1-prod-depin-backend",container="depin-backend"}) / 1024 / 1024
```

## Cross-Referencing with kubectl

After querying Thanos, always cross-reference with `kubectl` for pod-level details:

```bash
# Stage pod status
kubectl get pods -n story | grep depin

# Resource usage snapshot
kubectl top pods -n story | grep depin

# Deployment health
kubectl get deployments -n story | grep depin
```

**Expected healthy baseline** (stage, 2026-05-06):
| Service | Replicas | CPU/pod | Memory/pod |
|---|---|---|---|
| depin-backend | 2/2 | ~0.5 mCPU | ~14 MB |
| ip-registration-confirmer | 3/3 | 1–142 mCPU | ~14 MB |
| ip-registration-poller | 2/2 | ~1 mCPU | ~4 MB |
| ip-registration-submitter | 3/3 | 731–919 mCPU | ~18 MB |

## Advanced Pod Investigation

When a pod shows anomalous memory/CPU, go beyond the basic queries. This methodology was proven on 2026-05-06 when investigating a pod with 14× sibling memory (570 MB vs 40 MB).

### Memory Composition Breakdown

High RSS with negligible cache → the app is actually using the memory. High cache with low RSS → file I/O, not a concern.

```promql
# Memory breakdown per pod — run for all pods to compare
container_memory_rss{namespace="story",container="depin-backend"} / 1024 / 1024
container_memory_cache{namespace="story",container="depin-backend"} / 1024 / 1024
container_memory_usage_bytes{namespace="story",container="depin-backend"} / 1024 / 1024
container_memory_working_set_bytes{namespace="story",container="depin-backend"} / 1024 / 1024
container_memory_max_usage_bytes{namespace="story",container="depin-backend"} / 1024 / 1024
```

**How to read the results:**
- `RSS` ≈ `Working Set` and `Cache` is tiny → actual application memory, not file buffers. Investigate further.
- `MaxEver` is much higher than current `RSS` → the pod spiked and came back down. Check when.
- `RSS` differs dramatically across pods (14×) but CPU is low → jemalloc arena retention after a one-time large allocation.

### Process-Level Parity Check

If one pod has a memory anomaly, rule out process/thread/socket count differences:

```promql
container_processes{namespace="story",container="depin-backend"}
container_threads{namespace="story",container="depin-backend"}
container_sockets{namespace="story",container="depin-backend"}
```

All pods should have identical counts. If the anomalous pod has more threads/sockets, it's doing extra work.

### Disk I/O Comparison

Rule out I/O-bound anomalies:

```promql
container_fs_reads_bytes_total{namespace="story",container="depin-backend"} / 1024 / 1024
container_fs_writes_bytes_total{namespace="story",container="depin-backend"} / 1024 / 1024
```

### OOM / Failure Counts

Check for near-miss OOM events:

```promql
container_oom_events_total{namespace="story",container="depin-backend"}
container_memory_failcnt{namespace="story",container="depin-backend"}
kube_pod_container_resource_requests{namespace="story",resource="memory"} / 1024 / 1024
```

### Correlating Deploy Events with Metric Spikes

When a pod has anomalous memory, correlate the spike time with the deployment timeline:

1. **Find pod creation time** via Thanos: `kube_pod_created{pod=~"use1-prod-depin-backend.*"}`
2. **Find the deploy PR** in GitHub: `gh pr list --state merged --search "merged:>=YYYY-MM-DD"` and look for `staging > main` merges
3. **Map the spike** with `query_range`: use 15m step, numeric epoch timestamps (NOT "-6h" strings — see Pitfall #6)
4. **Inspect the deploy PR's code changes** to find memory-heavy features (cluster graphs, safety signal refreshes, large migrations)
5. **Confirm with per-pod request rate** during the spike window to rule out traffic anomalies

### Range Query Timestamps

The `query_range` endpoint requires numeric Unix epoch timestamps. Strings like `-6h` or `now` are rejected with `bad_data: cannot parse "-6h" to a valid timestamp`.

```bash
# Correct range query
NOW=$(date +%s)
SIX_H_AGO=$((NOW - 21600))
curl -s ... "${GRAFANA_SERVER}/api/datasources/proxy/uid/thanos/api/v1/query_range?query=${ENCODED}&start=${SIX_H_AGO}&end=${NOW}&step=900"

# For Python-based range queries
start = int(datetime(2026, 5, 5, 19, 45, 0, tzinfo=timezone.utc).timestamp())
end   = int(datetime(2026, 5, 5, 21, 15, 0, tzinfo=timezone.utc).timestamp())
```

### Decision Framework

| Observation | Likely Cause | Action |
|---|---|---|
| Stable high RSS, low CPU, no restarts, one pod only | jemalloc arena retention | Monitor; restart pod if cosmetic |
| Growing RSS linearly over hours, CPU normal | Slow memory leak | Investigate code; check for unbounded collections |
| High RSS + high CPU + high latency | Active heavy processing | Check background job timing; consider rate limiting |
| All pods affected equally | Normal workload | Baseline; not anomalous |
| High cache, low RSS | File-system I/O | Normal; not a memory concern |
| MaxEver >> Current RSS | One-time spike | Check what happened at spike time; correlate with deploys |
