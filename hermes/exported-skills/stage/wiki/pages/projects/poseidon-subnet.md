---
title: "Poseidon Subnet"
type: "project"
slug: "projects/poseidon-subnet"
freshness: "2025-10-20T23:36:00Z"
tags:
  - "devnet"
  - "l2"
  - "poseidon"
  - "subnet"
  - "video-processing"
owners:
  - "Jdub"
  - "Ramtin"
  - "Ze"
source_revision_ids:
  - "srcrev_4f25de37cc1e5aa0437d23f87f742da6"
conflict_state: "none"
---

# Poseidon Subnet

## Summary

Status and roadmap of the Poseidon Subnet project as of the internal devnet phase. Achievements include a functional L2, on-chain control plane, token bridge, worker, and API. Remaining work covers production hardening, research, and user-facing tooling. Key open questions involve business model, product simplicity, and technical innovation.

## Claims

- The subnet team has been developing for about 1.5 months since the announcement. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- August was a very productive month, validating core use cases with an OP Stack L2 as control plane, and bootstrapping infrastructure, monitoring, design, documentation. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- A fully functional L2 with a block explorer and Grafana monitoring was achieved. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- A fully on-chain control plane handles task scheduling, worker registration, epoch and reward distribution. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- A L1 POS token contract and token deposit/withdrawal bridge were implemented. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- A worker runs the Poseidon video processing pipeline. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- A Subnet API service handles authentication and file download/upload. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- A SoT design spec keeps track of functional specs and design discussions. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- The design of the L2-centric subnet solution was validated with an end-to-end video processing workflow run. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- Still need to deliver a production system accounting for API, worker scalability, worker authentication, testing infrastructure, and release cadence. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- Design and research decisions were postponed; active research is still required. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- User-facing product needs documentation and tooling for API users, subnet operators, and worker operators. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- Need a reliable Keeper system. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- Need versioned and reproducible infrastructure. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- Need full test suites and release process. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- Need internal security audits. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- Research items include: scheduling algorithm, L2 Gas and MEV, support reward and slashing proof and proof challenge, validator consensus for data quality, commit reveal to prevent validator copying, data privacy integration, governance and tokenomics. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- Research is owned by Jdub, Ramtin, Ze. `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- Business level open question: PMF - who will be subnet operators and what tasks they will provision; ideal case is organic use cases. `claim:claim_1_19` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- Product level open question: how to make subnet operator's job easier, abstract away L2 complexity and technical barriers. `claim:claim_1_20` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- Technical level open question: how to leverage existing L2 stacks like OP Stack and Reth while still innovating. `claim:claim_1_21` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_dc29808856a2de5f3262a4bd1a9bfd7e` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-1` `source_timestamp=2025-10-20T23:36:00Z`
- Question: How much to pay per work? `claim:claim_1_22` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- Validation is subjective. `claim:claim_1_23` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- Question: Partner vs audio. `claim:claim_1_24` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`
- The timeline is tracked in the Poseidon Partner Devnet planning page. `claim:claim_1_25` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2) `source_document_id=srcdoc_032e37d2992f750a85022f43ac0c9319` `source_revision_id=srcrev_4f25de37cc1e5aa0437d23f87f742da6` `chunk_id=srcchunk_0a4cfea848a41f682a582833f29c4171` `native_locator=https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98#chunk-2` `source_timestamp=2025-10-20T23:36:00Z`

## Open Questions

- How much to pay per work?
- How to leverage existing L2 stacks (OP Stack, Reth) while still innovating?
- How to make the subnet operator’s job easier and abstract away L2 complexity?
- Partner vs audio use case prioritization?
- Validation is subjective – how to handle?
- Who will be the subnet operators and what tasks will they provision? (PMF)

## Related Pages

- `poseidon-partner-devnet-planning`
- `superbridge-ui-for-bridging`

## Sources

- `source_document_id`: `srcdoc_032e37d2992f750a85022f43ac0c9319`
- `source_revision_id`: `srcrev_4f25de37cc1e5aa0437d23f87f742da6`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-Subnet-The-Road-to-Partner-Devnet-25e051299a548004adb3dbca7add6f98)
