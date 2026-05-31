# Precompile Gas Optimization Analysis for Story Protocol

How much gas could be saved by implementing key Story operations (MetadataURISet, mintAndRegisterIp) as Geth-native precompiled contracts, following Story's existing ipgraph precompile model.

## Story's ipgraph Precompile (Calibration Baseline)

The ipgraph precompile at `0x0000000000000000000000000000000000000101` (registered as 0x0101 in piplabs/story-geth core/vm/contracts.go) is Story's existing custom precompile that manages IP relationships and royalties natively in Go.

Source: https://docs.story.foundation/network/learn/node-software/precompiled-contracts

**Gas model:**
- Internal reads: 10-40 gas (`getParentIpsCount`=10, `hasParentIp`=40, `getParentIps`=40)
- Internal writes: ~1,000 gas (`setRoyalty`=1000, `addParentIp`=>1100)
- External calls (calling back into EVM): 2,100-126,000 gas (pays full CALL overhead)
- Gas formulas are multiplicative: `hasParentIp = ipGraphReadGas * averageParentIpCount`

**Key insight**: A state write (SSTORE equivalent) costs ~1,000 gas in a precompile vs 20,000 (cold) or 2,900 (warm) in interpreted EVM bytecode. This is because precompiles:
1. Bypass the EVM interpreter entirely (~50% of execution cost per Paradigm Research)
2. Access the stateDB directly without per-opcode gas metering
3. Charge gas based on actual native computation cost, not the conservative EVM gas schedule
4. Eliminate Solidity-level overhead (function dispatch, proxy routing, modifier checks, CALL/RETURN stack)

**Implementation in story-geth**: Precompiles implement the Geth `PrecompiledContract` interface:
```go
type PrecompiledContract interface {
    RequiredGas(input []byte) uint64  // deterministic gas from input
    Run(evm *EVM, input []byte) ([]byte, error)  // native Go execution
    Name() string
}
```

The `RequiredGas()` method returns a fixed gas cost based purely on input — no runtime gas griefing. The `Run()` method executes native Go with direct stateDB access.

## MetadataURISet: Solidity vs Precompile

### Current Solidity Implementation (CoreMetadataModule.sol)

Source: https://github.com/storyprotocol/protocol-core-v1/blob/main/contracts/modules/metadata/CoreMetadataModule.sol

Operation breakdown:
1. `verifyPermission(ipId)` — external call to AccessController: ~5,000-20,000 gas
2. `onlyMutable(ipId)` modifier — SLOAD IMMUTABLE flag from IPAccount: ~2,100 gas
3. `IPAccount.setString("METADATA_URI", metadataURI)` — SSTORE (string): ~20,000 gas
4. `IPAccount.setBytes32("METADATA_HASH", metadataHash)` — SSTORE (bytes32): ~5,000 gas
5. `emit MetadataURISet(ipId, metadataURI, metadataHash)` — LOG opcodes: ~2,000 gas
6. EVM interpreter overhead (opcode dispatch, stack ops, CALL/RETURN): ~30,000-50,000 gas

**Total: ~75,000 gas (warm) / ~120,000 gas (cold)**

### Proposed Precompile Implementation

Operations (all native Go, following ipgraph gas rates):
1. Permission check: direct stateDB read → ~40 gas
2. Immutable check: direct stateDB read → ~10 gas
3. Metadata URI write: stateDB write → ~1,000 gas
4. Metadata hash write: stateDB write → ~1,000 gas
5. Event emission: Go native log → ~100 gas
6. Precompile call overhead (warm): ~100 gas

**Estimated: ~2,000-3,000 gas — 25-37× reduction**

### Throughput Impact (36M gas limit, 2.2s blocks)

| | Current (Solidity) | Precompile |
|---|---|---|
| Gas/tx | 75,000 | 3,000 |
| Tx/block | 480 | 12,000 |
| TPS | ~218 | ~5,455 |
| Daily capacity | ~18.9M | ~471M |

## mintAndRegisterIp: Solidity vs Precompile

### Current Solidity Implementation

Operation breakdown:
1. IPAccount deployment (ERC-6551, CREATE + code deposit): ~35,000 gas
2. Registration in IPAssetRegistry (multiple SSTOREs + events): ~200,000-400,000 gas
3. Permission setup (AccessController writes): ~50,000-100,000 gas
4. Metadata set (if included inline): ~75,000 gas
5. Proxy routing, checks, emits, overhead: ~300,000-500,000 gas

**Total: ~1,000,000 gas (typical)**

### Proposed Precompile Implementation

All operations natively in Go in a single precompile call:
1. IPAccount creation: direct stateDB → ~5,000-10,000 gas
2. Registration writes (5-10 SSTOREs at ipgraph rate): ~5,000-10,000 gas
3. Permission setup (state writes): ~2,000-5,000 gas
4. Metadata write: ~2,000-3,000 gas
5. Events and overhead: ~500-1,000 gas

**Estimated: ~30,000-50,000 gas — 20-33× reduction**

### Throughput Impact (36M gas limit, 2.2s blocks)

| | Current (Solidity) | Precompile |
|---|---|---|
| Gas/tx | 1,000,000 | 40,000 |
| Tx/block | 36 | 900 |
| TPS | ~16 | ~409 |
| Daily capacity | ~1.4M | ~35M |

## Remaining Bottlenecks After Precompile Optimization

Even with precompiles, these bottlenecks remain:

1. **State root calculation**: 75% of block time per Paradigm Research. With 12,000 writes/block, trie hashing dominates.
2. **Block propagation**: 12K tx × ~250 bytes ≈ 3MB/block, approaching Story's 4MB max_bytes.
3. **Transaction validation**: 12K signature verifications at ~0.1ms each ≈ 1.2s of block time.
4. **Gas limit (36M)**: Still caps at 12K tx/block for 3K-gas MetadataURISet. Raising to 100M → 33K tx/block = ~15,000 TPS.

## Key Sources

- Story Precompile Docs: https://docs.story.foundation/network/learn/node-software/precompiled-contracts
- piplabs/story-geth contracts.go: https://github.com/piplabs/story-geth/blob/main/core/vm/contracts.go
- CoreMetadataModule.sol: https://github.com/storyprotocol/protocol-core-v1/blob/main/contracts/modules/metadata/CoreMetadataModule.sol
- Paradigm Research: https://www.paradigm.xyz/2024/04/reth-perf
- Nethermind Gas Benchmarks: https://www.nethermind.io/blog/measuring-ethereums-execution-limits-the-gas-benchmarking-framework
- evm.codes precompile gas schedule: https://www.evm.codes/precompiled
