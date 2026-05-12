# Story Protocol L1 Chain Parameters (Verified May 2026)

## Core Chain Parameters

| Parameter | Value | Source |
|---|---|---|
| Block gas limit | 36,000,000 (36M) | Blockscout live block data |
| Block time (live, measured) | ~2.0–2.2s (2.184s avg over 50 blocks) | [Blockscout Stats API](https://www.storyscan.io/api/v2/stats) |
| Block time (design target) | ~3.0s (10,368,000 blocks/year) | [Token Economy Docs](https://docs.story.foundation/network/learn/token-economy) |
| Chain ID | 1514 (mainnet), 1315 (Aeneid testnet) | docs.story.foundation |
| Current network utilization | ~0.16% | Blockscout |
| Total transactions (all-time) | ~104,981,080 | Blockscout |
| Avg TPS since genesis | ~2.8 TPS | Computed |

## Consensus Layer

| Component | Detail |
|---|---|
| Consensus engine | CometBFT (Tendermint fork) |
| Framework | Cosmos SDK |
| Interface | ABCI++ (ABCI 2.0) between CometBFT and state machine |
| Finality | Instant / one-shot finality |
| Active validators | Top 64 by stake weight |
| Cosmos `max_gas` | `-1` (unlimited at consensus layer — enforcement at execution layer) |

## Execution Layer

| Component | Detail |
|---|---|
| Execution client | Fork of Geth (go-ethereum) |
| EVM compatibility | 100% EVM equivalent |
| Ethereum upgrades | Pectra upgrade features supported |
| EL-CL communication | Standard Engine API |
| Custom precompiles | p256Verify (secp256r1), IP precompile (0x0000...0101) |
| EIP-1559 | Supported (base fee + priority fee, fees burned) |

## Tokenomics / Staking

| Parameter | Value |
|---|---|
| Genesis supply | 1 billion $IP |
| Annual inflation | 20M $IP |
| Min self-delegation | 1,024 IP |
| Min commission | 5% |
| Unbonding period | 14 days |
| Slashing (double sign) | 5% + permanent jailing |
| Slashing (downtime) | 0.02% |
| Jail cooldown | 10 minutes |

## On-Chain Contract Addresses

### EAS (Mainnet — Chain 1514)

| Contract | Address |
|---|---|
| EAS | `0x5bF79CECE7D1C9DA45a9F0dE480589ecCE1B48c8` |
| Schema Registry | `0x5F983ab12EE78535C9067dE1CDFc7C511320fB7d` |

### EAS (Aeneid Testnet — Chain 1315)

| Contract | Address |
|---|---|
| EAS | `0xDcd40C896274E7e9776A48deB0fA34999935Ee55` |
| Schema Registry | `0x2a3565551548abfcdeB9983230D9CAcBb8c6c16c` |

### Attestation Contracts

| Component | Detail |
|---|---|
| Repo | storyprotocol/attestation-contracts |
| Main contract | `Attestator.sol` — UUPS proxy |
| Key functions | `multiAttest(MultiAttestationRequest[])`, `multiRevoke(MultiRevocationRequest[])`, `registerSchema(string, resolver, revocable)` |
| Access control | `approvedCallers` mapping, set by owner |

## Off-Chain Services

### Indexing Pipeline (Two-Layer)

```
story-indexer (L1) → PostgreSQL (raw events)
        ↓
story-orchestration-service (L2) → PostgreSQL (enriched data)
        ↓
API (Gin HTTP) + Temporal Workflows
```

### story-indexer
- Repo: piplabs/story-indexer
- Language: Go
- Port: 8080
- DB: PostgreSQL (`story_indexer`)
- Replicas: 2 (staging)
- Config API: `POST /config/add`, `GET /config/list`, `POST /config/start`, `POST /config/stop`
- Key files: `src/indexer/event_indexer.go`, `src/indexer/event_store.go`

### story-orchestration-service
- Repo: piplabs/story-orchestration-service
- Language: Go (Gin HTTP + Temporal SDK)
- DBs: Blockchain DB (raw events) + Royalty Graph DB (enriched)
- Worker types: indexer-worker, story-temporal-worker, content-moderation-worker, royalty-graph-worker, royalty-graph-v2-worker
- API groups: `/workflows/v1`, `/graph/v2`, `/wallet-manager/v1`, `/content-moderation/v1`, `/stats/v1`, `/ownership/v1`
- Vercel app: Royalty Graph Explorer at `story-orchestration-service/web/` (Next.js 14)

### Note on story-api / story-api-v2
The `story-api` and `story-api-v2` deployments in K8s are **nginx placeholders**, not active API services. The `story-orchestration-service` is the actual enriched data API.

## Gas Reference

| Transaction Type | Est. Gas | Max Tx/Block (36M) | TPS @ 2.0s | TPS @ 2.2s |
|---|---|---|---|---|
| Native $IP transfer | 21,000 | 1,714 | 857 | 779 |
| ERC-20 transfer | ~65,000 | 554 | 277 | 252 |
| Uniswap V2 swap | ~150,000 | 240 | 120 | 109 |
| `MetadataURISet` (warm, typical) | ~75,000 | 480 | 240 | 218 |
| `MetadataURISet` (cold, first set) | ~120,000 | 300 | 150 | 136 |
| `MetadataURISet` (best case) | ~60,000 | 600 | 300 | 273 |
| Full `mintAndRegisterIp` | ~1,000,000 | 36 | 18 | 16 |
| EAS `multiAttest` (2000-item batch) | ~100,000 | 360 | 180 | 164 |
| EAS attest (amortized per metadata) | ~50 | N/A | N/A | N/A |

## Depin-Backend IP Registration Constants

| Constant | Value | Source |
|---|---|---|
| Hardcoded gas price | 0.1 Gwei (100,000,000 wei) | `infra/chain.rs:17` |
| Gas limit multiplier | 1.3× | `config.rs` |
| Block scan window | 32 blocks (~64s) | `config.rs` |
| Extended scan window | 3,600 blocks (~2h) | `config.rs` |
| Mainnet RPC | `https://internal-full.storyrpc.io` | `config.rs` |
| Aeneid RPC | `https://internal-full.aeneid.storyrpc.io` | `config.rs` |
| Pending nonce timeout | 10 seconds | `infra/chain.rs` |
| Active signer wallets | 40 | Production metrics |
| Current throughput | 0.44 tx/s (~38K/day) | Thanos (2026-05-06) |
| Peak 15-min throughput | 0.90 tx/s | Thanos (2026-05-06) |
| Conservative capacity (40 wallets) | ~4 tx/s (~345K/day) | Computed |
| Theoretical capacity (40 wallets) | ~20 tx/s (~1.7M/day) | Computed |

## Key URLs

| Resource | URL |
|---|---|
| Network Overview | https://docs.story.foundation/network/overview |
| Consensus Layer Docs | https://docs.story.foundation/network/learn/node-software/consensus_layer |
| Execution Layer Docs | https://docs.story.foundation/network/learn/node-software/execution_layer |
| Engine API Docs | https://docs.story.foundation/network/learn/node-software/engine_api |
| Token Economy / Staking | https://docs.story.foundation/network/learn/token-economy |
| Precompiled Contracts | https://docs.story.foundation/network/learn/node-software/precompiled-contracts |
| Mainnet Connection Info | https://docs.story.foundation/network/connect/mainnet |
| GitHub (consensus client) | https://github.com/piplabs/story |
| Genesis config | https://raw.githubusercontent.com/piplabs/story/main/lib/netconf/story/genesis.json |
| Blockscout Explorer | https://www.storyscan.io |
| Blockscout Stats API | https://www.storyscan.io/api/v2/stats |
| Grafana (depin) | https://grafana.ops.storyprotocol.net |

## Bottleneck Summary

| Layer | Limit | Approx Capacity |
|---|---|---|
| Consensus (CometBFT) | 1,000+ TPS | Not the bottleneck |
| EVM execution (Geth fork) | 50-200+ MGas/s (modern clients) | Story needs 16.4 MGas/s at 36M/2.2s — 3-12× headroom |
| Execution (Geth EVM, 36M gas) | Gas-limit governed | Depends on tx complexity |
| Full IP registration (1M gas) | ~18 tx/block | ~777K/day |
| Metadata-only update (75K gas) | ~480 tx/block | ~10.4M/day |
| Wallet nonce (40 wallets) | ~4-20 TPS | ~345K-1.7M/day |
| EAS batch attestation | ~360 batches/block | Millions/day |

### EVM Execution Speed Reference

Modern EVM execution clients process 50-200+ MGas/s on production hardware. Source: Paradigm Research (Apr 2024) reports Reth achieves 100-200 MGas/s during live sync; Nethermind Gas Benchmarking Framework (Nov 2025) shows clients handle hundreds of MGas/s for common opcodes. Geth (Story's fork) is in the same order of magnitude.

At Story's 36M gas / 2.2s block time: 16.4 MGas/s required — 3-12× below client capability. The EVM interpreter is NOT the bottleneck. Gas limit could be raised to ~110M (3×) before EVM execution approaches limits. The real constraints on gas limit increases are state growth (SSTORE bloat), block propagation latency, and validator hardware requirements — not raw execution speed.

### Precompile Optimization Potential

See `references/precompile-gas-analysis.md` for full analysis. Summary:

| Operation | Current (Solidity) | Precompile (est.) | Improvement | TPS Impact |
|---|---|---|---|---|
| MetadataURISet | ~75,000 gas | ~2,000-3,000 gas | 25-37× | 218 → ~5,400 TPS |
| mintAndRegisterIp | ~1,000,000 gas | ~30,000-50,000 gas | 20-33× | 16 → ~400 TPS |

Calibrated against Story's existing ipgraph precompile (0x0101): writes at ~1,000 gas, reads at 10-40 gas. Source: docs.story.foundation precompile docs, piplabs/story-geth contracts.go. External EVM benchmarks in `references/evm-performance-benchmarks.md`.
