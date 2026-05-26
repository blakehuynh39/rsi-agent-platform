---
title: "Story → Data Rebrand: Migration Plan"
type: "project"
slug: "projects/story-to-data-rebrand-migration-plan"
freshness: "2026-05-26T21:52:00Z"
tags:
  - "branding"
  - "data"
  - "migration"
  - "rebrand"
  - "story"
  - "token"
owners: []
source_revision_ids:
  - "srcrev_e8f8f64ac77a70465ff9282517f19732"
conflict_state: "none"
---

# Story → Data Rebrand: Migration Plan

## Summary

Comprehensive migration plan to rebrand Story Foundation to Data Foundation and the gas token IP to DATA across all APIs, frontends, SDKs, and supporting infrastructure. Smart-contract and blockchain internals are out of scope.

## Claims

- The rebrand scope includes every API, indexer, website, frontend, backend, SDK, and supporting infra; changing Story Foundation to Data Foundation and the gas token IP to DATA. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_457de01e5c75fda4207c27ea7f757234` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1` `source_timestamp=2026-05-26T21:52:00Z`
- Smart-contract, blockchain, and consensus internals are out of scope; the on-chain symbol does not change, only its display. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_457de01e5c75fda4207c27ea7f757234` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1` `source_timestamp=2026-05-26T21:52:00Z`
- Five targeted changes: organization/brand to Data Foundation, native gas token symbol to DATA, wrapped token to WDATA, logos/wordmarks/favicons to Data marks, and public domains to new Data domains. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_457de01e5c75fda4207c27ea7f757234` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1` `source_timestamp=2026-05-26T21:52:00Z`
- Success criteria: no user-visible Story brand string or logo on any in-scope surface; token reads DATA/WDATA everywhere; SDK consumers have a clear path off old packages; domains resolve and redirect. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_457de01e5c75fda4207c27ea7f757234` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1` `source_timestamp=2026-05-26T21:52:00Z`
- Only the gas token named IP changes. The protocol’s intellectual-property concept 'IP' (e.g., IPGraph, IP Asset, ipId) remains unchanged. Blind find-and-replace must be avoided. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_457de01e5c75fda4207c27ea7f757234` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-1` `source_timestamp=2026-05-26T21:52:00Z`
- Workstream A (Marketing & Docs) must update story.foundation (marketing apex), docs.story.foundation, FAQ, and tokenomics pages; 4× “Story Foundation” instances, 7× story.foundation links need changing. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-2) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_f8dfa3a34124675c11bce7a7728b2211` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-2` `source_timestamp=2026-05-26T21:52:00Z`
- Workstream B (Explorers/Blockscout) requires config changes for mainnet 1514, aeneid 1315, devnet0 1511: flip NEXT_PUBLIC_NETWORK_CURRENCY_SYMBOL from IP to DATA and update token-detail labels from WIP/Wrapped IP to WDATA/Wrapped DATA. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-2) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_f8dfa3a34124675c11bce7a7728b2211` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-2` `source_timestamp=2026-05-26T21:52:00Z`
- Workstream C (Frontends/dApps) covers story-staking-dashboard, app-staking-tlv2, story-global-wallet, story-dex-swaps, jiffie-webapp, gif-monorepo, commemorative mints, badge portals, and faucet frontends; each needs brand, token symbol ($IP→$DATA), domain/RPC updates. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-2) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_f8dfa3a34124675c11bce7a7728b2211` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-2` `source_timestamp=2026-05-26T21:52:00Z`
- Workstream D (SDKs & Developer Tooling) includes rename of @story-protocol/core-sdk, story_protocol_python_sdk, ipkit, cdr-sdk, create-story-app, and examples; old packages must be deprecated with re-export and warnings. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-2) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_f8dfa3a34124675c11bce7a7728b2211` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-2` `source_timestamp=2026-05-26T21:52:00Z`
- Workstream E (AI agents, internal tools, docs/governance) includes agents like jinn-agent, numo-agent, blockchain-agent and others; internal dashboards (control-plane-dashboard, qa-metrics-api); and docs/SIPs – only brand strings need updating, not token UI. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-3) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_44ad1719b22b1640689fba543532e038` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-3` `source_timestamp=2026-05-26T21:52:00Z`
- Cross-cutting tasks: domain migration with 301 redirects, package release/deprecation, design asset production (Data logo set), partner coordination (Blockscout, Goldsky, exchanges), and discovery of missing repos (marketing apex, faucet frontends, Blockscout config deploy repo). `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-3) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_44ad1719b22b1640689fba543532e038` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-3` `source_timestamp=2026-05-26T21:52:00Z`
- Separate-brand repositories (poseidon-*, numo-*, cdr-*, trace-frontend, aura-app, qualify, scout-careers) require confirmation before inclusion; trace-frontend already ships “Data Foundation” copy and is a design reference. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-3) `source_document_id=srcdoc_968b131229778649c9ec4971acea517c` `source_revision_id=srcrev_e8f8f64ac77a70465ff9282517f19732` `chunk_id=srcchunk_44ad1719b22b1640689fba543532e038` `native_locator=https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846#chunk-3` `source_timestamp=2026-05-26T21:52:00Z`

## Open Questions

- What are the repos for the faucet frontends (faucet.*, aeneid.faucet.*, etc.)?
- What is the location of the marketing apex (story.foundation) repo?
- Where is the Blockscout config deploy repo that vendors common-frontend.env?
- Which repositories under separate brands (poseidon-*, numo-*, cdr-*, etc.) are in scope? Brand team must confirm.

## Sources

- `source_document_id`: `srcdoc_968b131229778649c9ec4971acea517c`
- `source_revision_id`: `srcrev_e8f8f64ac77a70465ff9282517f19732`
- `source_url`: [Notion source](https://www.notion.so/Story-Data-Rebrand-Migration-Plan-36c051299a5481f1895ac1408fe3d846)
