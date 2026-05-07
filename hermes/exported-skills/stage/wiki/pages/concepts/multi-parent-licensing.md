---
title: "Multi-Parent Licensing"
type: "concept"
slug: "concepts/multi-parent-licensing"
freshness: "2024-01-25T04:55:00Z"
tags:
  - "derivatives"
  - "licensing"
  - "multi-parent"
  - "policy"
owners: []
source_revision_ids:
  - "srcrev_17f5ce4aca3b66fe7fb1792b7208765e"
conflict_state: "none"
---

# Multi-Parent Licensing

## Summary

Describes how derivatives can be linked to multiple parent IPAs, including scenarios with a single common policy and the more complex case of multiple, potentially incompatible policies.

## Claims

- A derivative can be linked to a single parent by burning a License NFT (L1), which causes the derivative to inherit the parent's policy (P1). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-1) `source_document_id=srcdoc_9fb8b8b3fadb15514caf762c155b56c6` `source_revision_id=srcrev_17f5ce4aca3b66fe7fb1792b7208765e` `chunk_id=srcchunk_35d837622f0112f3adcf5a89834a5b91` `native_locator=https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-1` `source_timestamp=2024-01-25T04:55:00Z`
- For multiple parents agreeing to the same terms, all licensors delegate minting rights to a single address, which mints a license. Burning this license creates a derivative with multiple parents, useful for composite works. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-1) `source_document_id=srcdoc_9fb8b8b3fadb15514caf762c155b56c6` `source_revision_id=srcrev_17f5ce4aca3b66fe7fb1792b7208765e` `chunk_id=srcchunk_35d837622f0112f3adcf5a89834a5b91` `native_locator=https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-1` `source_timestamp=2024-01-25T04:55:00Z`
- When multiple parents have different, incompatible policies, linking them simultaneously may be blocked. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-2) `source_document_id=srcdoc_9fb8b8b3fadb15514caf762c155b56c6` `source_revision_id=srcrev_17f5ce4aca3b66fe7fb1792b7208765e` `chunk_id=srcchunk_a65a50309d3ac2ba274714f528676c09` `native_locator=https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-2` `source_timestamp=2024-01-25T04:55:00Z`
- If multiple parents have compatible policies, the output policy for the derivative could be the more restrictive one or a fusion of terms, keeping the more restrictive where they overlap. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-3) `source_document_id=srcdoc_9fb8b8b3fadb15514caf762c155b56c6` `source_revision_id=srcrev_17f5ce4aca3b66fe7fb1792b7208765e` `chunk_id=srcchunk_77422661f7755dd9f7c13827a2f2ccf6` `native_locator=https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-3` `source_timestamp=2024-01-25T04:55:00Z`
- Generalizing multi-parent policy resolution is considered a very complicated problem, especially for the Beta release. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-3) `source_document_id=srcdoc_9fb8b8b3fadb15514caf762c155b56c6` `source_revision_id=srcrev_17f5ce4aca3b66fe7fb1792b7208765e` `chunk_id=srcchunk_77422661f7755dd9f7c13827a2f2ccf6` `native_locator=https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-3` `source_timestamp=2024-01-25T04:55:00Z`
- A proposed compromise for Beta is to limit multiple policies only to Root IPAs and to introduce Limited Policy Flavors for UML. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-3) `source_document_id=srcdoc_9fb8b8b3fadb15514caf762c155b56c6` `source_revision_id=srcrev_17f5ce4aca3b66fe7fb1792b7208765e` `chunk_id=srcchunk_77422661f7755dd9f7c13827a2f2ccf6` `native_locator=https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1#chunk-3` `source_timestamp=2024-01-25T04:55:00Z`

## Open Questions

- How should incompatible policy terms be detected and handled on-chain?
- What is the exact algorithm for determining the output policy when multiple compatible parent policies are merged?

## Sources

- `source_document_id`: `srcdoc_9fb8b8b3fadb15514caf762c155b56c6`
- `source_revision_id`: `srcrev_17f5ce4aca3b66fe7fb1792b7208765e`
- `source_url`: [Notion source](https://www.notion.so/Licensing-flavors-multiple-parents-7f8685c27d3a499b9c13da0e3dd2b2a1)
