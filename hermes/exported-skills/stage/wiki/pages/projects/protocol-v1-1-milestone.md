---
title: "Protocol v1.1 Milestone"
type: "project"
slug: "projects/protocol-v1-1-milestone"
freshness: "2024-04-17T08:40:00Z"
tags:
  - "milestone"
  - "protocol"
  - "v1.1"
owners: []
source_revision_ids:
  - "srcrev_53c410f86cf886abdd7164a234871659"
conflict_state: "none"
---

# Protocol v1.1 Milestone

## Summary

Tasks and considerations for the v1.1 protocol milestone, covering core repo changes and periphery repo guidelines.

## Claims

- The v1.1 milestone includes renaming RoyaltyTokenAddedToVault to RevenueTokenAddedToVault for clarity. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Default License Terms should remain immutable for old IPs. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Use Solady's ERC6551 implementation. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Separate setPermission from setting signers. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Refactor Minting Fee and Receiver Check Hooks into a Unified LicensingHook. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- License Token should have no expiry. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Optimize LicenseTokenMetadata struct in ILicenseToken. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Optimize PILTerms struct. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Resolve cyclical dependencies in contract deployment using CREATE3. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Consider implementing Multicall in IPAccount Storage. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Consider moving License Template compatibility check to individual LicenseTemplates. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Consider mismatch between revert requirements in IPAccount Implementation. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Consider allowing detaching unused License Terms. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Consider that IPAccountRegistry exposes registerIpAccount as public method. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`
- Periphery contracts should hold minimal states, working as wrapper functions that combine multiple core interactions. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9) `source_document_id=srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5` `source_revision_id=srcrev_53c410f86cf886abdd7164a234871659` `chunk_id=srcchunk_80692ab43c0ec20fd2778e995b4804f8` `native_locator=https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9` `source_timestamp=2024-04-17T08:40:00Z`

## Sources

- `source_document_id`: `srcdoc_e8eb9b48026dffb956ae64a7dc0b03d5`
- `source_revision_id`: `srcrev_53c410f86cf886abdd7164a234871659`
- `source_url`: [Notion source](https://www.notion.so/Protocol-Milestones-27f5ff955b294c088ad8fd8e4bd073e9)
