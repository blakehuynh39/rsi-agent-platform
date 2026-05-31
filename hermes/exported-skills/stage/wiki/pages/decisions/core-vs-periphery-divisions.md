---
title: "Core vs. Periphery Divisions"
type: "decision"
slug: "decisions/core-vs-periphery-divisions"
freshness: "2024-03-18T08:10:00Z"
tags:
  - "architecture"
  - "contracts"
  - "repository-structure"
owners:
  - "user://ce2077ef-5025-403e-b8e2-2d9d1c2c7bbd"
source_revision_ids:
  - "srcrev_138b13f915ee48b03b1d808c99f3ca3f"
conflict_state: "none"
---

# Core vs. Periphery Divisions

## Summary

Decision on dividing smart contracts into Core and Periphery repositories, based on Uniswap's pattern, with a deadline of March 14, 2024.

## Claims

- The decision to adopt core and periphery division requires agreement from stakeholders. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03) `source_document_id=srcdoc_27f0312a5a70894c863c425bcf2153b4` `source_revision_id=srcrev_138b13f915ee48b03b1d808c99f3ca3f` `chunk_id=srcchunk_3c1592bbf6ce94612ad517b8a2aeff30` `native_locator=https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03` `source_timestamp=2024-03-18T08:10:00Z`
- The decision deadline for Core vs. Periphery divisions is March 14, 2024. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03) `source_document_id=srcdoc_27f0312a5a70894c863c425bcf2153b4` `source_revision_id=srcrev_138b13f915ee48b03b1d808c99f3ca3f` `chunk_id=srcchunk_3c1592bbf6ce94612ad517b8a2aeff30` `native_locator=https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03` `source_timestamp=2024-03-18T08:10:00Z`
- Core contracts provide fundamental safety guarantees and define the logic of pool generation and interactions. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03) `source_document_id=srcdoc_27f0312a5a70894c863c425bcf2153b4` `source_revision_id=srcrev_138b13f915ee48b03b1d808c99f3ca3f` `chunk_id=srcchunk_3c1592bbf6ce94612ad517b8a2aeff30` `native_locator=https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03` `source_timestamp=2024-03-18T08:10:00Z`
- Periphery contracts interact with core contracts but are not part of the core, supporting domain-specific interactions. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03) `source_document_id=srcdoc_27f0312a5a70894c863c425bcf2153b4` `source_revision_id=srcrev_138b13f915ee48b03b1d808c99f3ca3f` `chunk_id=srcchunk_3c1592bbf6ce94612ad517b8a2aeff30` `native_locator=https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03` `source_timestamp=2024-03-18T08:10:00Z`

## Open Questions

- Which specific contract categories will be assigned to Core vs. Periphery?

## Sources

- `source_document_id`: `srcdoc_27f0312a5a70894c863c425bcf2153b4`
- `source_revision_id`: `srcrev_138b13f915ee48b03b1d808c99f3ca3f`
- `source_url`: [Notion source](https://www.notion.so/4-Core-vs-Periphery-Divisions-0e72a30b5dd44769a7127b0e6ae0ea03)
