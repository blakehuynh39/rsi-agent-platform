---
title: "Proteus 2.0 Planning"
type: "project"
slug: "projects/proteus-2-0-planning"
freshness: "2026-02-03T01:34:00Z"
tags:
  - "planning"
  - "proteus"
  - "roadmap"
  - "subnet"
owners: []
source_revision_ids:
  - "srcrev_bbdf50c34f0b8be0c05310d70c18eec0"
conflict_state: "none"
---

# Proteus 2.0 Planning

## Summary

Planning document for Proteus 2.0, covering high-level goals, issues found in Partner devnet, and next steps including optimistic validation, contract improvements, load testing, and a phased timeline.

## Claims

- High level goals for Proteus 2.0 are: 1) Build subnet 0 for audio data (public testnet Proteus), 2) Build a software stack for other subnet operators and workers (Subnet Stacks). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_19647b268a2bdcf60475d9fbfa1a6dd4` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1` `source_timestamp=2026-02-03T01:34:00Z`
- Through Partner devnet, Proteus works end-to-end with ability to handle decent loads, and Subnet Stacks allows subnet operators to run an entire subnet locally. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_19647b268a2bdcf60475d9fbfa1a6dd4` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1` `source_timestamp=2026-02-03T01:34:00Z`
- Issues found on Proteus include duplicate compute, non-happy path issues, L2 instability, reward fairness, load test infra setup, and low performance. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_19647b268a2bdcf60475d9fbfa1a6dd4` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1` `source_timestamp=2026-02-03T01:34:00Z`
- Issues found on Subnet Stacks include the need to improve documentation based on feedback from BHarvest and to add SMC and Worker Portal to the stack. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_19647b268a2bdcf60475d9fbfa1a6dd4` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1` `source_timestamp=2026-02-03T01:34:00Z`
- Next step: Move toward an optimistic validation system where validators only validate when someone challenges a miner's work, to reduce duplicate compute. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_19647b268a2bdcf60475d9fbfa1a6dd4` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1` `source_timestamp=2026-02-03T01:34:00Z`
- Contract improvements planned: give rewards per work (addresses reward fairness), retry/timeout handling, subnet owner registry, refactor functions, and delegation (P1). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_19647b268a2bdcf60475d9fbfa1a6dd4` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-1` `source_timestamp=2026-02-03T01:34:00Z`
- Timeline: Dec 15 - Jan 2 for planning/scoping, Jan 5 - Jan 30 for all engineering work, Feb 5 - Feb 14 for full testing and release. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-2) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_df32cab4954e136c16c726ba94c0abfc` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-2` `source_timestamp=2026-02-03T01:34:00Z`
- Engineering work includes integrating RaaS provider with Aeneid, adding optimistic validation, latest AI processing logic, reward distribution features, reliability code, performance improvements, and infra automation. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-2) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_df32cab4954e136c16c726ba94c0abfc` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-2` `source_timestamp=2026-02-03T01:34:00Z`
- Full testing and release phase includes optimistic validation, challenger system, infra automation (CICD, monitoring), performance/load test, and contract retry and reward distribution enhancement. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-2) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_df32cab4954e136c16c726ba94c0abfc` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-2` `source_timestamp=2026-02-03T01:34:00Z`
- Things to explore include decentralized Ray vision (resource based scheduling) and verifiability prototypes (tensor commits, COCOON - TEE). `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-2) `source_document_id=srcdoc_b116a497807ca0d392887662c4a7874f` `source_revision_id=srcrev_bbdf50c34f0b8be0c05310d70c18eec0` `chunk_id=srcchunk_df32cab4954e136c16c726ba94c0abfc` `native_locator=https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069#chunk-2` `source_timestamp=2026-02-03T01:34:00Z`

## Sources

- `source_document_id`: `srcdoc_b116a497807ca0d392887662c4a7874f`
- `source_revision_id`: `srcrev_bbdf50c34f0b8be0c05310d70c18eec0`
- `source_url`: [Notion source](https://www.notion.so/Subnet-planning-Proteus-2-0-2af051299a5480f2a6dff193454bb069)
