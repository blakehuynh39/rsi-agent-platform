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

**Critical**: `GRAFANA_TOKEN` alone returns 401 on `/api/datasources/proxy/uid/thanos/...`. You MUST include all three headers:
```
Authorization: Bearer ${GRAFANA_TOKEN}
CF-Access-Client-Id: ${RSI_GRAFANA_CF_ACCESS_CLIENT_ID}
CF-Access-Client-Secret: ${RSI_GRAFANA_CF_ACCESS_CLIENT_SECRET}
```

## Metric Labels Reference

### Stage
- `job`: `use1-stage-depin-backend`
- `namespace`: `story`
- `cluster`: `use1-stage`
- `environment`: `stage`
- `prometheus_replica`: `prometheus-prometheus-kube-prometheus-prometheus-0`

### Prod
- `job`: `use1-prod-depin-backend`
- `namespace`: (prod namespace — not `story`)
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
  -H "CF-Access-Client-Id: ${RSI_GRAFANA_CF_ACCESS_CLIENT_ID}" \
  -H "CF-Access-Client-Secret: ${RSI_GRAFANA_CF_ACCESS_CLIENT_SECRET}" \
  "${GRAFANA_SERVER}/api/datasources/proxy/uid/thanos/api/v1/query?query=${ENCODED}"
```

Parse results with: `python3 -c "import json,sys; d=json.load(sys.stdin); ..."`

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

### 2. CF-Access Headers Required for Proxy
`GRAFANA_TOKEN` returns 200 on dashboard/health/search endpoints but 401 on datasource proxy endpoints. The proxy requires Cloudflare Access tunnel authentication.

### 3. Prod Pods Not in `story` Namespace
Prod pod queries via Thanos use the prod namespace (not `story`). `kubectl get pods -n story` will only show stage pods. For prod pod restarts/status, you must use Thanos metrics.

### 4. Memory Anomaly (jemalloc Arena Retention)
A prod pod may show 14× memory vs siblings (e.g., 570 MB vs 40 MB) without being a leak. Characteristics of jemalloc arena retention:
- Stable over hours (not growing)
- Low CPU usage (no active processing)
- No restarts (pod uptime matches siblings)
- Only one pod affected (not all)

This is normal behavior after heavy background job processing (multiplier sweep, hot path cache rebuild). If memory is growing linearly over time AND CPU is elevated, that's a real leak.

### 5. IP Registration Jobs Lack HTTP Metrics
The confirmer/poller/submitter pods don't expose `http_requests_total` or `http_request_duration_seconds_bucket`. Their health is only visible through container CPU/memory metrics and pod status.

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
