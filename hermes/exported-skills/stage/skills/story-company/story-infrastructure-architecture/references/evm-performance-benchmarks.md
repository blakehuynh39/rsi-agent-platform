# EVM Execution Client Performance Benchmarks

External benchmarks showing that modern EVM execution clients process 50-200+ MGas/s, making EVM execution speed NOT the bottleneck for chains operating at 36M gas / 2-3s block time (~12-18 MGas/s required).

## Paradigm Research — "Reth's Path to 1 Gigagas per Second" (Apr 2024)

Source: https://www.paradigm.xyz/2024/04/reth-perf

**Current performance**: Reth achieves **100-200 MGas/s** during live sync (including sender recovery, execution, and trie calculation).

**Key findings:**
- ~50% of execution time is in the EVM interpreter — JIT/AOT compilation could yield ~2× improvement
- Parallel EVM could yield up to 5× improvement (80% of Ethereum storage slots accessed independently)
- >75% of block sealing time is spent on state root calculation, NOT execution
- Short-term goal: 1 gigagas/s (10× improvement over current)

**Scaling strategies:**
- Vertical: JIT/AOT EVM, parallel EVM, pipelined state root
- Horizontal: Multi-rollup Reth, cloud-native Reth (autoscaling service stack)

**Relevance to Story**: Story's Geth fork at 36M gas / 2.2s block time needs ~16.4 MGas/s. Even a conservative 50 MGas/s Geth estimate gives 3× headroom before EVM execution becomes the bottleneck.

## Nethermind Gas Benchmarking Framework (Nov 2025)

Source: https://www.nethermind.io/blog/measuring-ethereums-execution-limits-the-gas-benchmarking-framework

**What it is**: A shared standard across all Ethereum execution clients (Nethermind, Geth, Reth, Besu, Erigon, Nimbus, Ethrex) that fills blocks with single opcodes or precompiles to push execution to computational limits, measuring throughput in MGas/s.

**Key findings:**
- Clients can process **hundreds of MGas/s** for common opcodes
- Only pathological precompiles (ModExp worst-case) pushed clients to their limits — leading to EIP-7883 repricing
- The Berlinterop event (June 2025) gave confidence to raise Ethereum's gas limit from 36M → 45M, with plans for 60M+ in the Fusaka fork

**Repository**: https://github.com/NethermindEth/gas-benchmarks
- 635+ commits, 115 branches, 10 releases
- Supports 7 clients: Nethermind, Geth, Reth, Besu, Erigon, Nimbus, Ethrex
- Integrated with Ethereum Execution Specs (EELS) for reproducible test definitions

## Gas-per-Second as a Standard Metric

Paradigm proposed gas-per-second (MGas/s) as a standard EVM performance metric alongside TPS. TPS is insufficient because it doesn't capture computational work (a 21K-gas transfer and a 1M-gas registration are both "1 transaction"). MGas/s enables direct comparison of:
- Client performance (how fast can each client execute?)
- Chain capacity (how much gas can the network handle per second?)
- DOS resistance (what's the worst-case gas/second an attacker can demand?)

## Implication for Gas Limit Decisions

| Gas Limit | MGas/s Needed (2.2s) | Verdict |
|-----------|---------------------|---------|
| 36M (current) | ~16.4 | Comfortable — 3-12× below client capability |
| 60M (Ethereum Fusaka target) | ~27.3 | Well within limits |
| 110M | ~50 | Still fine (conservative Geth ceiling) |
| 220M | ~100 | EVM can handle (Reth-like performance) |
| 440M | ~200 | Needs JIT/parallel EVM optimizations |

The real constraints on raising gas limits are NOT EVM execution speed — they are state growth, block propagation latency, validator hardware requirements, and DoS surface area.
