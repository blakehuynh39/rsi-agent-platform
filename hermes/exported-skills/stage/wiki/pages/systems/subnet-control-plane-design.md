---
title: "Subnet Control Plane Design"
type: "system"
slug: "systems/subnet-control-plane-design"
freshness: "2025-08-26T03:05:00Z"
tags:
  - "control-plane"
  - "epochs"
  - "l2"
  - "poseidon"
  - "smart-contract"
  - "staking"
  - "subnet"
owners: []
source_revision_ids:
  - "srcrev_88d0f3058cd24cfbd8a6f022c432fb7f"
conflict_state: "none"
---

# Subnet Control Plane Design

## Summary

Design specification for the Subnet Control Plane smart contract on the Poseidon L2 OP Stack chain. Oversees worker registration, staking, epoch-based reward distribution, worker health monitoring via heartbeats, and jailing mechanics.

## Claims

- The Subnet Control Plane is a Layer 2 OP Stack chain governance and coordination contract for the Poseidon Subnet, managing worker registration, staking, epoch-based reward distribution, and worker health monitoring. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-1) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_81f60a303909036a19562fda07629506` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-1` `source_timestamp=2025-08-26T03:05:00Z`
- Architecture includes three core contracts: SubnetControlPlane for governance, SubnetTreasury for token custody, and PoseidonToken as a bridged ERC20 from L1. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-1) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_81f60a303909036a19562fda07629506` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-1` `source_timestamp=2025-08-26T03:05:00Z`
- WorkerInfo struct fields: workerAddress, stakedAmount, registeredAt, lastHeartbeat, isActive, isJailed, missedHeartbeats, unstakeRequestedAt, unstakeEffectiveEpoch. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-1) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_81f60a303909036a19562fda07629506` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-1` `source_timestamp=2025-08-26T03:05:00Z`
- Epoch struct fields: epochId, startTime, endTime, totalStaked, totalRewards, finalized, activeWorkersCount. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-1) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_81f60a303909036a19562fda07629506` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-1` `source_timestamp=2025-08-26T03:05:00Z`
- Workers must send heartbeats at configured intervals. Exceeding maxMissedHeartbeats results in jailing, which prevents reward claims and excludes the worker from the active set. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_e913f539788cbcffdeff39fd4dd1578c` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2` `source_timestamp=2025-08-26T03:05:00Z`
- Epochs are fixed 7-day periods; anyone can trigger advanceEpoch() when current epoch ends. It finalizes the epoch, selects active workers, distributes rewards equally among them, and creates a new epoch. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_e913f539788cbcffdeff39fd4dd1578c` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2` `source_timestamp=2025-08-26T03:05:00Z`
- Active worker selection criteria: must be active (registered, not unstaked), not jailed, no pending unstake for next epoch, and if exceeding max, earliest registered workers are selected first. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_e913f539788cbcffdeff39fd4dd1578c` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2` `source_timestamp=2025-08-26T03:05:00Z`
- Unstaking has a 2-epoch delay (T+2). Worker remains active and can earn rewards during the waiting period. Anyone can call withdrawStake after the effective epoch to receive tokens. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_e913f539788cbcffdeff39fd4dd1578c` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2` `source_timestamp=2025-08-26T03:05:00Z`
- Reward distribution is pull-based: rewards are calculated equally per active worker at epoch finalization and stored; workers call claimRewards(epochId) to withdraw, provided they are not jailed and haven't already claimed. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_e913f539788cbcffdeff39fd4dd1578c` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-2` `source_timestamp=2025-08-26T03:05:00Z`
- Default configuration: minimumStake = 10000 PSDN, rewardsPerEpoch = 1000 PSDN, epochInterval = 7 days, maxActiveWorkers = 100, heartbeatInterval = 1 hour, maxMissedHeartbeats = 5. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-3) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_261ded71c2282b3b40f5ea74e63b5a20` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-3` `source_timestamp=2025-08-26T03:05:00Z`
- Gas optimization strategies include batch operations during epoch advancement, pull-based rewards, storage packing for WorkerInfo, and minimal loops with early exit in worker selection. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-3) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_261ded71c2282b3b40f5ea74e63b5a20` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-3` `source_timestamp=2025-08-26T03:05:00Z`
- Workers integrate via registerWorker(stakeAmount), workerHeartbeat(activeTasks), requestUnstake(), withdrawStake(myAddress), claimRewards(epochId). Subnet systems can call advanceEpoch(), claimRewardsFor(), and query isWorkerActive/isWorkerJailed/getActiveWorkers. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-3) `source_document_id=srcdoc_63de5499415df2a8cbd2c3f5650fbc6d` `source_revision_id=srcrev_88d0f3058cd24cfbd8a6f022c432fb7f` `chunk_id=srcchunk_261ded71c2282b3b40f5ea74e63b5a20` `native_locator=https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17#chunk-3` `source_timestamp=2025-08-26T03:05:00Z`

## Sources

- `source_document_id`: `srcdoc_63de5499415df2a8cbd2c3f5650fbc6d`
- `source_revision_id`: `srcrev_88d0f3058cd24cfbd8a6f022c432fb7f`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Control-Plane-Design-Document-25b051299a548011a352d09d52feec17)
