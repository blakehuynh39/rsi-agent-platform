---
title: "Multi-parent IP Derivation"
type: "decision"
slug: "decisions/multi-parent-ip-derivation"
freshness: "2024-05-07T19:28:00Z"
tags:
  - "composite-ipa"
  - "derivatives"
  - "ip-graph"
  - "multi-parent"
  - "royalties"
owners: []
source_revision_ids:
  - "srcrev_61be16f0c579aa1a481729d3e2d0b563"
conflict_state: "none"
---

# Multi-parent IP Derivation

## Summary

Discussion and proposals for handling derivative IP assets that have multiple parents, including the introduction of composite IPAs (cIPA), reciprocal derivative flags, royalty waterfall models, and term selection rules.

## Claims

- The issue of IPs with multiple parents was known and raised again during the audit of v1 contracts. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-1) `source_document_id=srcdoc_0a5ae3d99bf800e0569bf47ba842ff10` `source_revision_id=srcrev_61be16f0c579aa1a481729d3e2d0b563` `chunk_id=srcchunk_7822be0cf2fe4d36580829202cfff4af` `native_locator=https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-1` `source_timestamp=2024-05-07T19:28:00Z`
- Two derivative scenarios exist based on the 'derivativesReciprocal' parameter: if false, only one level of derivation is allowed; if true, reciprocal derivation is permitted. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-1) `source_document_id=srcdoc_0a5ae3d99bf800e0569bf47ba842ff10` `source_revision_id=srcrev_61be16f0c579aa1a481729d3e2d0b563` `chunk_id=srcchunk_7822be0cf2fe4d36580829202cfff4af` `native_locator=https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-1` `source_timestamp=2024-05-07T19:28:00Z`
- Ben proposed that P3 with multiple parents should be a new composite IPA (cIPA) that can set its own economic rules for added value, possibly disallowing further derivatives, and that the royalty module could support a waterfall/pool model where P3 participates. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-2) `source_document_id=srcdoc_0a5ae3d99bf800e0569bf47ba842ff10` `source_revision_id=srcrev_61be16f0c579aa1a481729d3e2d0b563` `chunk_id=srcchunk_239783b19796ac428bf8771e8e5ca110` `native_locator=https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-2` `source_timestamp=2024-05-07T19:28:00Z`
- In the off-chain world, composite works usually do not add new material, providing low incentive to create derivatives. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-2) `source_document_id=srcdoc_0a5ae3d99bf800e0569bf47ba842ff10` `source_revision_id=srcrev_61be16f0c579aa1a481729d3e2d0b563` `chunk_id=srcchunk_239783b19796ac428bf8771e8e5ca110` `native_locator=https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-2` `source_timestamp=2024-05-07T19:28:00Z`
- JZ proposed a royalty range model for share-alike terms, allowing children to set royalty splits within a specified range while maintaining share-alike compliance. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-2) `source_document_id=srcdoc_0a5ae3d99bf800e0569bf47ba842ff10` `source_revision_id=srcrev_61be16f0c579aa1a481729d3e2d0b563` `chunk_id=srcchunk_239783b19796ac428bf8771e8e5ca110` `native_locator=https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-2` `source_timestamp=2024-05-07T19:28:00Z`
- Raul proposed that when a composite IPA derives from multiple parents, it can either select one parent’s terms for its own derivatives, define new terms with different revenue share/mint fees, or create composite terms using a rule like 'most expensive term wins'. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-2) `source_document_id=srcdoc_0a5ae3d99bf800e0569bf47ba842ff10` `source_revision_id=srcrev_61be16f0c579aa1a481729d3e2d0b563` `chunk_id=srcchunk_239783b19796ac428bf8771e8e5ca110` `native_locator=https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160#chunk-2` `source_timestamp=2024-05-07T19:28:00Z`

## Open Questions

- How will the royalty module implement a waterfall/pool model for multi-parent IPs?
- Should composite IPAs (cIPAs) be a distinct asset class with special rules?
- What is the final mechanism for handling royalty splits and derivative permissions when an IPA has multiple parents?
- Which royalty range bounds or term selection rules will be adopted?

## Related Pages

- `derivative-parameters`
- `ipa-classification`
- `royalty-module`

## Sources

- `source_document_id`: `srcdoc_0a5ae3d99bf800e0569bf47ba842ff10`
- `source_revision_id`: `srcrev_61be16f0c579aa1a481729d3e2d0b563`
- `source_url`: [Notion source](https://www.notion.so/Growing-the-IP-Graph-for-IPs-with-multiple-parents-d3722c41e72c442baa32fc85ba34b160)
