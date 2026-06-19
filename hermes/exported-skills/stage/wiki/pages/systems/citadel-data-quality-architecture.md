---
title: "Citadel Data Quality Architecture"
type: "system"
slug: "systems/citadel-data-quality-architecture"
freshness: "2026-05-05T02:09:28Z"
tags:
  - "architecture"
  - "data-quality"
  - "incentive"
  - "processing"
  - "sentry"
owners:
  - "U04L0DD6B6F"
  - "U04L0DD71TM"
  - "U06A5AQ1VD3"
  - "U081RCLP2KB"
  - "U08AGDT08E7"
source_revision_ids:
  - "srcrev_650f6971f3c8add56ebe8bfc8041f9f6"
  - "srcrev_be92607bfb704be0641a5fc32cbe0097"
conflict_state: "none"
---

# Citadel Data Quality Architecture

## Summary

Proposed multi-layer architecture for improving data quality and minimizing costs, consisting of Sentry, Incentive, and Processing layers. Includes Castle location service, user metadata checks, boosts and golden label incentives, and a Numo-Poseidon API handover.

## Claims

- A Citadel Proposal was shared to improve data quality and minimize long-term costs by breaking responsibilities into smaller modules and creating feedback loops. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b55f38ffd282cd966e41a72635eb0514` `source_revision_id=srcrev_be92607bfb704be0641a5fc32cbe0097` `chunk_id=srcchunk_6e9fd2054dad1ddc18e135426633742c` `native_locator=slack:C0AL7EKNHDF:1777944330.894619:1777944330.894619` `source_timestamp=2026-05-05T01:25:30Z`
- The approach is based on past learnings from NFT airdrops, Amazon Mechanical Turk data collection, and similar products. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b55f38ffd282cd966e41a72635eb0514` `source_revision_id=srcrev_be92607bfb704be0641a5fc32cbe0097` `chunk_id=srcchunk_6e9fd2054dad1ddc18e135426633742c` `native_locator=slack:C0AL7EKNHDF:1777944330.894619:1777944330.894619` `source_timestamp=2026-05-05T01:25:30Z`
- The team already has a Sentry Layer consisting of Castle (location service) and user metadata cross-checked against IP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b55f38ffd282cd966e41a72635eb0514` `source_revision_id=srcrev_650f6971f3c8add56ebe8bfc8041f9f6` `chunk_id=srcchunk_54c710977e5a640e37be8d5725720c48` `native_locator=slack:C0AL7EKNHDF:1777944330.894619:1777946952.393849` `source_timestamp=2026-05-05T02:09:28Z`
- The Incentive Layer includes boosts with quests/multipliers and a golden label data quality experiment, with plans to create an open-source recaptcha-like experience. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b55f38ffd282cd966e41a72635eb0514` `source_revision_id=srcrev_650f6971f3c8add56ebe8bfc8041f9f6` `chunk_id=srcchunk_54c710977e5a640e37be8d5725720c48` `native_locator=slack:C0AL7EKNHDF:1777944330.894619:1777946952.393849` `source_timestamp=2026-05-05T02:09:28Z`
- The Processing Layer involves an API handover between Numo and Poseidon to propagate results and reflect on user balances. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b55f38ffd282cd966e41a72635eb0514` `source_revision_id=srcrev_650f6971f3c8add56ebe8bfc8041f9f6` `chunk_id=srcchunk_54c710977e5a640e37be8d5725720c48` `native_locator=slack:C0AL7EKNHDF:1777944330.894619:1777946952.393849` `source_timestamp=2026-05-05T02:09:28Z`

## Open Questions

- How will the open-source recaptcha-like golden label experiment be implemented? What components will be open-sourced?

## Sources

- `source_document_id`: `srcdoc_b55f38ffd282cd966e41a72635eb0514`
- `source_revision_id`: `srcrev_f2f31b32cd7216e47cc911f454265e61`
