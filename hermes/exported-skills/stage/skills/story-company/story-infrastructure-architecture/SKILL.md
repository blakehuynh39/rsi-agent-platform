---
name: story-infrastructure-architecture
description: "Investigate Story Protocol's technical infrastructure: L1 blockchain parameters, on-chain contracts (EAS, attestation, protocol core/periphery), off-chain services (indexer, orchestration, API), throughput analysis, and architectural patterns."
version: 1.0.0
metadata:
  hermes:
    tags: [story-protocol, blockchain, architecture, infrastructure, indexing, attestation, l1, evm, cometbft, eas, throughput]
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

Story infrastructure has four layers — narrow your search to the right one:

| Layer | Questions about | Primary Sources |
|---|---|---|
| **L1 Chain** | Block time, gas limit, consensus, throughput, finality | docs.story.foundation, Blockscout, piplabs/story (genesis.json), Thanos metrics |
| **On-chain Contracts** | EAS, attestation, protocol core, protocol periphery, gas costs | storyprotocol repos (attestation-contracts, protocol-core-v1, protocol-periphery-v1) |
| **Off-chain Services** | Indexer, orchestration, API proxy, Temporal workers, databases | piplabs repos (story-indexer, story-orchestration-service), K8s story namespace |
| **DePIN/Numo** | IP registration pipeline, metadata flow, wallet fleet | piplabs/depin-backend, piplabs/numo-monorepo, story-deployments |

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

## Report Structure

For infrastructure questions, structure the response around:
1. **Parameters** — concrete numbers with sources (Blockscout, genesis, docs)
2. **Architecture** — how components connect (diagram when useful)
3. **Bottlenecks** — what limits throughput at each layer
4. **Gaps** — what doesn't exist yet
5. **Concrete evidence** — cite specific URLs, block numbers, contract addresses
