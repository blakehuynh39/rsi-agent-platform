# Story Protocol L1 — Chain Parameters & Throughput Reference

> Last verified: 2026-05-06
> Sources: Blockscout (storyscan.io), docs.story.foundation, protocol-core-v1 GitHub, Thanos production metrics, depin-backend chain config

## Core Chain Parameters

| Parameter | Value | Source |
|---|---|---|
| **Block gas limit** | **36,000,000 (36M)** | Blockscout live block data |
| **Block time (live)** | **~2.0–2.2s** | Blockscout Stats API, Thanos metrics |
| **Block time (design)** | ~3.0s (10,368,000 blocks/year) | docs.story.foundation token economy |
| **Consensus engine** | CometBFT (Tendermint fork) with ABCI++ | docs.story.foundation consensus layer |
| **Framework** | Cosmos SDK | docs.story.foundation |
| **Execution client** | Fork of Geth (go-ethereum), 100% EVM equivalent | docs.story.foundation execution layer |
| **EIP-1559** | Supported (base fee + priority fee, fees burned) | Token economy docs |
| **Cosmos `max_gas`** | `-1` (unlimited at consensus layer) | genesis.json |
| **Active validators** | Top 64 by stake weight | Token economy docs |
| **Finality** | Instant/one-shot (CometBFT immediate finality) | docs.story.foundation |

## RPC Endpoints

| Network | URL |
|---|---|
| Mainnet | `https://internal-full.storyrpc.io` |
| Aeneid (testnet) | `https://internal-full.aeneid.storyrpc.io` |

## Gas Pricing (from depin-backend config)

| Constant | Value | Notes |
|---|---|---|
| `HARDCODED_FEE_WEI` | `100_000_000` (0.1 Gwei) | Both `max_fee_per_gas` and `max_priority_fee_per_gas` |
| `DEFAULT_GAS_LIMIT_MULTIPLIER` | `1.3×` | 30% headroom on `eth_estimateGas` |
| Max fee limit | Disabled (`None`) | Both Mainnet and Aeneid |

## Network Utilization (live as of 2026-05-06)

| Metric | Value |
|---|---|
| Total transactions (all-time) | ~105M |
| Average TPS since genesis | ~2.8 (very low) |
| Network utilization | ~1.23% |
| Per-block gas used | ~58K (0.16% of 36M limit) |

## Metadata Operations Gas Reference

### `setMetadataURI()` → `MetadataURISet` Event

From `CoreMetadataModule._setMetadataURI()` in `protocol-core-v1`:

```solidity
event MetadataURISet(address indexed ipId, string metadataURI, bytes32 metadataHash);
```

**Naming pitfall:** The event is `MetadataURISet`, NOT `MetaURIUpdate`. There is zero occurrence of "MetaURIUpdate" anywhere in the `protocol-core-v1` repository.

| Scenario | Gas |
|---|---|
| Warm storage (re-update) | ~60,000–85,000 |
| Cold storage (first set) | ~85,000–120,000 |
| **Typical (IPFS URI, warm controller)** | **~75,000** |

### Gas Breakdown

| Step | Operation | Est. Gas |
|---|---|---|
| 1 | Base tx + calldata (~167 bytes) | ~23,500 |
| 2 | `verifyPermission` → `IP_ASSET_REGISTRY.isIpAccount()` | ~2,600 |
| 3 | `verifyPermission` → `AccessController.checkPermission()` | ~12,000–22,000 |
| 4 | `onlyMutable` → `IPAccount.getBool("IMMUTABLE")` | ~2,600 |
| 5 | `LibString.escapeJSON(metadataURI)` | ~500–1,000 |
| 6 | `IPAccount.setString("METADATA_URI", ...)` (SSTORE) | ~8,000–25,000 |
| 7 | `IPAccount.setBytes32("METADATA_HASH", ...)` (SSTORE) | ~5,000–20,000 |
| 8 | `emit MetadataURISet` (LOG3, 1 indexed topic) | ~3,000 |
| 9 | Misc (dispatch, memory, stack) | ~2,000 |

## Max Theoretical TPS for MetadataURISet

Formula: `TPS = (BlockGasLimit / GasPerTx) × (1 / BlockTime)`

| Gas/Tx | Tx/Block | TPS @ 2.0s | TPS @ 2.2s | TPS @ 3.0s |
|---|---|---|---|---|
| 60,000 (best case) | 600 | 300 | 273 | 200 |
| **75,000 (typical)** | **480** | **240** | **218** | **160** |
| 100,000 | 360 | 180 | 164 | 120 |
| 120,000 (cold) | 300 | 150 | 136 | 100 |

## Bottleneck Analysis

| Layer | Capacity | Bottleneck? |
|---|---|---|
| Consensus (CometBFT) | >1,000 TPS | **No** |
| Execution (Geth EVM + 36M gas) | ~200+ TPS | Theoretical ceiling |
| Wallet nonce (40 wallets × 1 inflight) | ~4–20 TPS | **Practical bottleneck** |
| Mempool inclusion (p50) | 7.5s latency | Chain-side latency |

To exceed 20 TPS: add more signer wallets (~500 for 200 TPS) or use a batcher contract.

## Production Throughput (Numo depin-backend, as of 2026-05-06)

| Metric | Value |
|---|---|
| Current avg confirm rate | 0.44/s (~38K/day) |
| Peak 15-min rate | 0.90/s (~78K/day) |
| 40-wallet fleet conservative cap | ~4/s (~345K/day) |
| 40-wallet fleet theoretical cap | ~20/s (~1.7M/day) |
| Pipeline latency (p50, full lifecycle) | 10.0s |

## Key URLs

| Resource | URL |
|---|---|
| Blockscout Explorer | https://www.storyscan.io |
| Blockscout Stats API | https://www.storyscan.io/api/v2/stats |
| Network Overview | https://docs.story.foundation/network/overview |
| Consensus Layer | https://docs.story.foundation/network/learn/node-software/consensus_layer |
| Execution Layer | https://docs.story.foundation/network/learn/node-software/execution_layer |
| Token Economy | https://docs.story.foundation/network/learn/token-economy |
| GitHub (story client) | https://github.com/piplabs/story |
| protocol-core-v1 | https://github.com/storyprotocol/protocol-core-v1 |
| CoreMetadataModule.sol | https://github.com/storyprotocol/protocol-core-v1/blob/main/contracts/modules/metadata/CoreMetadataModule.sol |
