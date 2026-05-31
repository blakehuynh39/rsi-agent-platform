# Story Aeneid Testnet Observability

## Dashboard

**Name:** Go Ethereum By Instance
**UID:** `go-ethereum-by-instance`
**URL:** `/d/go-ethereum-by-instance/go-ethereum-by-instance`
**Folder:** L1 Team

The dashboard covers both mainnet and Aeneid. Switch via the `network` variable:
- `mainnet` — Story mainnet (chain 1514)
- `aeneid` — Story testnet (chain 1315)

The `chain` variable shows `story` for mainnet (via chain-head-exporter) and no-chain-label for Aeneid (via story-geth job). For Aeneid metrics, filter on `network="aeneid"` rather than `chain`.

## Aeneid Topology (19 instances)

| Region | Instance | Role |
|--------|----------|------|
| us-east-1a | use1-aeneid-validator1 | validator |
| us-east-1a | use1-aeneid-validator3 | validator |
| us-east-1a | use1-aeneid-bootnode1 | bootnode |
| us-east-1a | use1-aeneid-internal-rpc-archive1 | internal-rpc-archive |
| us-east-1a | use1-aeneid-internal-rpc-full1 | internal-rpc-full |
| us-east-1a | use1-aeneid-public-rpc1 | public-rpc |
| us-east-1a | use1-aeneid-snapshot-archive1 | snapshot-archive |
| us-east-1a | use1-aeneid-snapshot-full1 | snapshot-full |
| us-east-1b | use1-aeneid-validator2 | validator |
| us-east-1b | use1-aeneid-validator4 | validator |
| us-east-1b | use1-aeneid-bootnode2 | bootnode |
| us-east-1b | use1-aeneid-internal-rpc-archive2 | internal-rpc-archive |
| us-east-1b | use1-aeneid-internal-rpc-full2 | internal-rpc-full |
| us-east-1b | use1-aeneid-public-rpc2 | public-rpc |
| japaneast | jpe-aeneid-validator1..5 | validator (5 instances) |

Validators: 9 total (4 use1 + 5 jpe). The `job` label differs by region: `story-geth` for use1, `cdr-aeneid-story-geth` for jpe.

## Datasource

The dashboard panels use the `thanos` Prometheus datasource (uid: `thanos`), not the default `prometheus` datasource. Always specify `datasource=thanos` when querying these metrics via `rsi_observability_metrics_query`.

## Key Bottleneck Metrics

When analyzing Aeneid performance under load, query these metrics:

### Transaction Pool State
```promql
# Executable (pending) transactions — spikes indicate load bursts
txpool_pending{network="aeneid", role="validator"}

# Gapped (queued) transactions — non-zero means nonce gaps
txpool_queued{network="aeneid", role="validator"}
```

### Transaction Acceptance Rate
```promql
# Rate of accepted transactions (per-validator)
rate(txpool_valid{network="aeneid", role="validator"}[1m])

# Rejected transactions (sum these to check for waste)
rate(txpool_invalid{network="aeneid", role="validator"}[1m]) +
rate(txpool_underpriced{network="aeneid", role="validator"}[1m])
```

### Block Execution Time (THE bottleneck indicator)
```promql
# p50 block execution time in nanoseconds
chain_execution{network="aeneid", role="validator", quantile="0.5"}

# Also available: chain_validation, chain_write, chain_account_reads,
# chain_account_updates, chain_storage_reads, chain_storage_updates,
# chain_storage_commits, chain_snapshot_commits
```

### System Resources
```promql
# CPU: geth process load
system_cpu_procload{network="aeneid"}

# Memory
system_memory_used{network="aeneid"}

# Disk I/O
rate(system_disk_readbytes{network="aeneid"}[1m])
rate(system_disk_writebytes{network="aeneid"}[1m])
```

## Bottleneck Analysis Recipe

1. **Check txpool_pending** — if it spikes and stays high, the bottleneck is upstream of block production
2. **Check txpool_invalid + txpool_underpriced** — if non-zero, transactions are malformed or gas too low
3. **Check chain_execution** — compare idle baseline (~50,000-70,000 ns) vs load-test peak. A 1000x+ increase means EVM execution is the bottleneck per block
4. **Check system resources** — CPU saturation or disk I/O spikes can indicate hardware limits

**Idle baseline (Aeneid validators, May 2026):**
- `txpool_pending`: ~310-330 (ambient network transactions)
- `txpool_valid rate`: ~0 tx/s (no load)
- `chain_execution` p50: ~50,000-70,000 ns (0.05-0.07 ms)

**Load-test peaks observed (May 21, 2026):**
- `txpool_pending`: 4,700-6,100
- `txpool_valid rate`: 624-712 tx/s per validator
- `chain_execution` p50: 71,000,000-99,000,000 ns (71-99 ms)
- All transactions accepted (0 invalid + underpriced)
- Mempool drained completely after each burst

## Grafana Dashboard Panel Map

| Panel | Title | Key Metrics |
|-------|-------|-------------|
| 5 | Executable transactions | `txpool_pending` (stat) |
| 6 | Transaction pool | `txpool_pending`, `txpool_queued`, `txpool_local` (timeseries) |
| 12 | Block processing | `chain_execution`, `chain_validation`, `chain_write` + storage/account metrics |
| 13 | Transaction propagation | `txpool_valid`, `txpool_invalid`, `txpool_underpriced` rates |
| 15 | Success RPC QPS | `rpc_success` by method |
| 20 | CPU | `system_cpu_procload` (geth), `system_cpu_sysload` |
| 21 | Memory | `system_memory_used`, `system_memory_held` |
