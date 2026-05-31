---
title: "Story Partner Testnet L1 Readiness Plan"
type: "project"
slug: "projects/story-partner-testnet-l1-readiness-plan"
freshness: "2024-07-26T18:22:00Z"
tags:
  - "l1"
  - "partner-testnet"
  - "readiness"
  - "story"
owners: []
source_revision_ids:
  - "srcrev_cf6d7520bc332462165071b7806f8f80"
conflict_state: "none"
---

# Story Partner Testnet L1 Readiness Plan

## Summary

Outlines the documentation, release cycle, CI/CD, node setup, validator setup, and developer guide preparation for the Story partner testnet L1 launch.

## Claims

- Documentation must include an introduction, user guide, node setup, validator setup, and developer guide, with priorities assigned (p0 or p1). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-1) `source_document_id=srcdoc_c2b6a49dc99aa3cdce0e474fb712967c` `source_revision_id=srcrev_cf6d7520bc332462165071b7806f8f80` `chunk_id=srcchunk_aff3977e5849db8ad409036fa04a8611` `native_locator=https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-1` `source_timestamp=2024-07-26T18:22:00Z`
- The team proposes a release cycle with Alpha (now until partner testnet), Beta, Pre-release, and Stable stages using semantic versioning, but exact timelines and stage adoption are TBD. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-1) `source_document_id=srcdoc_c2b6a49dc99aa3cdce0e474fb712967c` `source_revision_id=srcrev_cf6d7520bc332462165071b7806f8f80` `chunk_id=srcchunk_aff3977e5849db8ad409036fa04a8611` `native_locator=https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-1` `source_timestamp=2024-07-26T18:22:00Z`
- CI/CD needs finalization of client distribution formats, a standardized release template, and an action plan to improve partner UX for node setup, possibly by unifying binaries and simplifying configuration. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-1) `source_document_id=srcdoc_c2b6a49dc99aa3cdce0e474fb712967c` `source_revision_id=srcrev_cf6d7520bc332462165071b7806f8f80` `chunk_id=srcchunk_aff3977e5849db8ad409036fa04a8611` `native_locator=https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-1` `source_timestamp=2024-07-26T18:22:00Z`
  - citation: [Notion source](https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-2) `source_document_id=srcdoc_c2b6a49dc99aa3cdce0e474fb712967c` `source_revision_id=srcrev_cf6d7520bc332462165071b7806f8f80` `chunk_id=srcchunk_6c6e8da10bb4255d3d7ce6d7d4c12f61` `native_locator=https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-2` `source_timestamp=2024-07-26T18:22:00Z`
- Detailed documentation is needed for Full nodes, Archive nodes, State Syncing nodes, and Validator nodes, including differences, service ports, configuration helpers, monitoring, RPC paths, caching rules, and troubleshooting. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-2) `source_document_id=srcdoc_c2b6a49dc99aa3cdce0e474fb712967c` `source_revision_id=srcrev_cf6d7520bc332462165071b7806f8f80` `chunk_id=srcchunk_6c6e8da10bb4255d3d7ce6d7d4c12f61` `native_locator=https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-2` `source_timestamp=2024-07-26T18:22:00Z`
- Validator setup requires documentation, CLI helper functions for provisioning and voting power checks, and ideally a programmatic faucet for ETH after user vetting. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-2) `source_document_id=srcdoc_c2b6a49dc99aa3cdce0e474fb712967c` `source_revision_id=srcrev_cf6d7520bc332462165071b7806f8f80` `chunk_id=srcchunk_6c6e8da10bb4255d3d7ce6d7d4c12f61` `native_locator=https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-2` `source_timestamp=2024-07-26T18:22:00Z`
- An internal runbook is needed for resetting the network after bugs on execution or consensus clients, or network fragmentation issues, and a public page for network uptime and genesis files. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-2) `source_document_id=srcdoc_c2b6a49dc99aa3cdce0e474fb712967c` `source_revision_id=srcrev_cf6d7520bc332462165071b7806f8f80` `chunk_id=srcchunk_6c6e8da10bb4255d3d7ce6d7d4c12f61` `native_locator=https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478#chunk-2` `source_timestamp=2024-07-26T18:22:00Z`

## Open Questions

- How should the binary handle one-time start and config initialization for different testnets?
- Should the Alpha release stage start now, and what timeline should it follow?
- Should we adopt the Beta stage as soon as the partner testnet starts?
- Should we introduce a Pre-release stage during the incentivized testnet?
- What is the exact format and timelines for distributing client versions for geth and iliad?

## Sources

- `source_document_id`: `srcdoc_c2b6a49dc99aa3cdce0e474fb712967c`
- `source_revision_id`: `srcrev_cf6d7520bc332462165071b7806f8f80`
- `source_url`: [Notion source](https://www.notion.so/Story-Partner-L1-Readiness-Doc-82243e0ec4384280b4b78211bfbab478)
