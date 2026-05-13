---
name: story-infrastructure-architecture
description: "Investigate Story Protocol's technical infrastructure: L1 blockchain parameters, on-chain contracts (EAS, attestation, protocol core/periphery), off-chain services (indexer, orchestration, API), throughput analysis, and architectural patterns."
version: 1.1.0
metadata:
  hermes:
    tags: [story-protocol, blockchain, architecture, infrastructure, indexing, attestation, l1, evm, cometbft, eas, throughput, precompile, gas-optimization]
    related_skills: [numo-project-status, depin-prod-admin-read, rsi-platform-investigation, cloudflare-management]
---

# Story Infrastructure Architecture

Use this skill when a Story request asks about blockchain parameters (block time, gas limits, TPS), on-chain contract infrastructure (EAS, attestation, protocol core), off-chain services (indexer, orchestration, API proxy), throughput analysis, or infrastructure architecture questions. Also trigger when the user asks about "how does the Story chain work" or "what infrastructure do we have" at any level.

## Investigation Workflow

### 1. Read the ingress Slack thread

Always read the full thread first — prior messages may contain relevant calculations, parameters, or links.

```
mcp slack_read_thread(channel_id, thread_ts)
```

### 2. Determine what layer the question targets

Story infrastructure has five layers — narrow your search to the right one:

| Layer | Questions about | Primary Sources |
|---|---|---|
| **L1 Chain** | Block time, gas limit, consensus, throughput, finality | docs.story.foundation, Blockscout, piplabs/story (genesis.json), Thanos metrics |
| **On-chain Contracts** | EAS, attestation, protocol core, protocol periphery, gas costs | storyprotocol repos (attestation-contracts, protocol-core-v1, protocol-periphery-v1) |
| **Off-chain Services** | Indexer, orchestration, API proxy, Temporal workers, databases | piplabs repos (story-indexer, story-orchestration-service), K8s story namespace |
| **Infrastructure-as-Code** | AWS resources (EKS, VPC, RDS, IAM), Terraform state, account topology, k8s deployment manifests | storyprotocol/story-infra-aws (high-level AWS infra + permissions), story-deployments (Helm charts, ArgoCD, per-service k8s topology) |
| **DePIN/Numo** | IP registration pipeline, metadata flow, wallet fleet | piplabs/depin-backend, piplabs/numo-monorepo, story-deployments |

**IaC Repo Split:**
- `storyprotocol/story-infra-aws` — **high-level AWS infrastructure**: Terraform for EKS clusters, VPC networking, RDS databases, IAM roles/policies, Karpenter, monitoring stack (Prometheus/Grafana/Loki/Tempo), across 4 accounts (dev/stage/prod/ops). This repo provisions the *platform* that everything runs on.
- `story-deployments` — **per-service Kubernetes deployment topology**: Helm charts, ArgoCD Application manifests, per-service k8s YAML. This is where you find *what* runs where — not the AWS infra underneath.

### 3. Gather chain parameters (L1 Layer)

Pull live data from Blockscout and cross-reference with docs:

```bash
# Live block gas limit (all blocks use same value)
curl -s "https://www.storyscan.io/api/v2/blocks?limit=1" | jq '.items[0].gas_limit'

# Block time (compute from recent blocks)
curl -s "https://www.storyscan.io/api/v2/stats" | jq '.average_block_time'

# Genesis config for consensus parameters
curl -s "https://raw.githubusercontent.com/piplabs/story/main/lib/netconf/story/genesis.json" | jq '.consensus_params'
```

**Known parameters (verified May 2026):**
- Block gas limit: 36,000,000 (36M)
- Block time: ~2.0–2.2s (live), ~3.0s (design target)
- Consensus: CometBFT (Tendermint fork), 64 active validators, instant finality
- Execution: Geth fork, 100% EVM equivalent, Pectra upgrade support
- Cosmos `max_gas`: `-1` (unlimited at consensus layer)
- Current network utilization: ~0.16%
- Chain ID: 1514 (mainnet), 1315 (Aeneid testnet)

**PITFALL:** The design block time (3.0s from docs.story.foundation) differs from the live block time (~2.0-2.2s from Blockscout). Always cite both and note which one you're using for calculations.

### 4. Investigate on-chain contracts

EAS is deployed on mainnet. Start with these addresses:

| Contract | Mainnet (Chain 1514) | Aeneid (Chain 1315) |
|---|---|---|
| EAS | `0x5bF79CECE7D1C9DA45a9F0dE480589ecCE1B48c8` | `0xDcd40C896274E7e9776A48deB0fA34999935Ee55` |
| Schema Registry | `0x5F983ab12EE78535C9067dE1CDFc7C511320fB7d` | `0x2a3565551548abfcdeB9983230D9CAcBb8c6c16c` |
| Attestator | storyprotocol/attestation-contracts | UUPS proxy |

```bash
# Clone attestation contracts
cd /tmp && git clone https://github.com/storyprotocol/attestation-contracts

# Clone protocol core for IP registration flow analysis
cd /tmp && git clone https://github.com/storyprotocol/protocol-core-v1

# Clone protocol periphery for workflow/multicall analysis
cd /tmp && git clone https://github.com/storyprotocol/protocol-periphery-v1
```

**Key contract facts:**
- `Attestator.sol` — UUPS proxy with `multiAttest(request[])`, gated to `approvedCallers`
- `MulticallUpgradeable` on all workflow contracts — `batchMintAndRegisterIp()` in SDK
- `IPAccount.executeBatch(Call[])` — ERC-6551 smart wallet batch execution
- Story Attestation Service (SAS) is **entirely off-chain** — no on-chain publication yet
- No Merkle proof verification contracts exist anywhere in the protocol repos

**PITFALL:** The SDK `batchMintAndRegisterIp()` uses multicall at the workflow (periphery) level, not a native batch on `IPAssetRegistry`. Each IP registration is discrete at the core protocol level. Gas savings from multicall are from combining txs, not from amortizing per-IP overhead.

### 5. Investigate off-chain services (indexing pipeline)

The indexing pipeline is a two-layer architecture:

```
story-indexer (Layer 1) → PostgreSQL (raw events)
        ↓
story-orchestration-service (Layer 2) → PostgreSQL (enriched data)
        ↓
API + Temporal Workflows
```

**story-indexer** (piplabs/story-indexer):
- Generic, config-driven event indexer using `eth_getLogs`
- Dynamic table creation per event in PostgreSQL (`story_indexer` DB)
- Configs via REST API: `POST /config/add`, `GET /config/list`
- Key source: `src/indexer/event_indexer.go`, `src/indexer/event_store.go`
- 2 replicas on K8s, fetches in batches of 1,000 blocks

**story-orchestration-service** (piplabs/story-orchestration-service):
- Temporal-based ETL with multiple worker pools
- Processes: Raw Events → Aggregation → Content Moderation → Metadata Indexing
- API server (Gin): `/workflows/v1`, `/graph/v2`, `/wallet-manager/v1`, `/content-moderation/v1`, `/stats/v1`
- Already consumes `MetadataURISet` events for content moderation, NFT metadata indexing, embeddings
- Two Postgres DBs: Blockchain DB (raw) + Royalty Graph DB (enriched)
- Deployed as 6+ worker deployments on K8s (story namespace)

```bash
# Check deployment status
kubectl get deployments -n story | grep -E "indexer|orchestration|worker"

# List indexer configs
kubectl exec -n story deploy/story-indexer -- curl -s localhost:8080/config/list
```

**PITFALL:** The `story-api` and `story-api-v2` in K8s are **nginx placeholders**, not active services. The `story-orchestration-service` is the actual API for enriched data. Don't confuse the placeholder with the running service.

### 6. Analyze gas costs and throughput

| Transaction Type | Est. Gas | Max Tx/Block (36M) | TPS @ 2.0s | TPS @ 2.2s |
|---|---|---|---|---|
| Native $IP transfer | 21,000 | 1,714 | 857 | 779 |
| `MetadataURISet` (warm) | ~75,000 | 480 | 240 | 218 |
| `MetadataURISet` (cold) | ~120,000 | 300 | 150 | 136 |
| Full `mintAndRegisterIp` | ~1,000,000 | 36 | 18 | 16 |
| EAS `multiAttest` (batch) | ~100,000 | 360 | 180 | 164 |
| EAS attest per metadata (2K/batch) | ~50 amortized | N/A | N/A | N/A |

**PITFALL:** Chain gas limit gives ~18 IP registrations/block (~777K/day), but the current 40-wallet fleet limits to ~4-20 TPS. When capacity questions arise, distinguish the **chain gas ceiling** from the **wallet nonce bottleneck**.

### 6b. Gas limit vs EVM execution speed analysis

When the question is about whether raising the gas limit helps or EVM execution speed is the bottleneck:

1. **Check live utilization**: `curl -s "https://www.storyscan.io/api/v2/stats" | jq '.network_utilization_percentage'` — if <5%, neither limit is under pressure.

2. **Compute required MGas/s**: `gas_limit / block_time_s`. At 36M gas / 2.2s = ~16.4 MGas/s.

3. **Compare against client benchmarks**: See `references/evm-performance-benchmarks.md`. Modern EVM clients process 50-200+ MGas/s. Story's Geth fork has 3-12× headroom at 36M gas.

4. **Determine bottleneck hierarchy**:
   - Consensus (CometBFT, max_gas=-1): not the bottleneck at any realistic scale
   - EVM execution (Geth): ~50 MGas/s conservative ceiling → would bottleneck at ~110M gas limit
   - Gas limit (36M): first chain-side bottleneck if blocks filled → caps MetadataURISet at ~218 TPS
   - Application (wallets, demand): current actual bottleneck at ~4-20 TPS

The real constraints on gas limit increases are **state growth** (SSTORE bloat), **block propagation** latency across 64 validators, and **validator hardware requirements** — not raw EVM execution speed.

### 6c. Analyze precompile optimization potential

When the question involves precompiles, precompile gas models, or IP registration / metadata gas optimization:

1. **Load the ipgraph precompile reference**: Story's existing ipgraph precompile (0x0101 in story-geth) provides the calibration. Reads: 10-40 gas, writes: ~1,000 gas. Source: docs.story.foundation precompile docs, `references/precompile-gas-analysis.md`.

2. **Break down current Solidity gas**: Analyze the Solidity implementation (e.g., CoreMetadataModule.sol) to identify: external calls, SSTOREs, modifier checks, EVM interpreter overhead. The EVM interpreter typically accounts for ~50% of execution cost (Paradigm Research).

3. **Estimate precompile gas at ipgraph rates**: Assign ipgraph-level gas to each operation: state reads at ~40 gas, state writes at ~1,000 gas. Precompiles bypass the EVM interpreter and charge only native computation cost.

4. **Compute throughput impact**: `tx_per_block = 36M / precompile_gas`, `tps = tx_per_block / block_time_s`.

5. **Identify remaining bottlenecks**: After precompile optimization, the bottleneck shifts to state root calculation (75% of block time per Paradigm), block propagation, and transaction signature validation.

See `references/precompile-gas-analysis.md` for full breakdown with gas-per-operation tables and throughput projections.

**PITFALL:** Precompile gas costs are set by the protocol, not the EVM gas schedule. The ipgraph write rate (~1,000 gas) may not perfectly transfer to MetadataURISet writes (which write variable-length strings vs fixed mapping entries). Use ipgraph as a conservative calibration point and note the uncertainty.

### 7. Search Honcho/Slack for prior discussions

```bash
mcp conversations_search(query="block time gas limit TPS throughput EAS attest")
mcp documents_search(query="Story protocol infrastructure architecture")
```

### 8. Check K8s deployment state (story namespace only)

```bash
kubectl get deployments -n story
kubectl get pods -n story --sort-by=.metadata.creationTimestamp
```

**PITFALL:** Use only `story` namespace for Story infra. The `rsi-platform` namespace is for the RSI agent platform, not Story services.

## Key Architecture Facts (Reference)

See `references/chain-parameters.md` for the full reference sheet of Story L1 parameters, on-chain contract addresses, off-chain service architecture, and gas/throughput tables. This is the go-to quick reference.

Additional support files:
- `references/evm-performance-benchmarks.md` — Paradigm and Nethermind EVM client benchmarks (MGas/s data)
- `references/precompile-gas-analysis.md` — Precompile optimization analysis using ipgraph calibration

## Report Structure

For infrastructure questions, structure the response around:
1. **Parameters** — concrete numbers with sources (Blockscout, genesis, docs)
2. **Architecture** — how components connect (diagram when useful)
3. **Bottlenecks** — what limits throughput at each layer (consensus → EVM execution → gas limit → application)
4. **Gaps** — what doesn't exist yet
5. **Concrete evidence** — cite specific URLs, block numbers, contract addresses

For throughput / gas-limit questions, additionally:
6. **EVM performance context** — compare Story's MGas/s requirement against client benchmarks (see `references/evm-performance-benchmarks.md`)

For precompile optimization questions, additionally:
7. **ipgraph calibration** — ground estimates in Story's existing precompile gas model (see `references/precompile-gas-analysis.md`)
