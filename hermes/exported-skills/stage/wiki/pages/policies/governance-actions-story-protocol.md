---
title: "Governance Actions for Story Protocol"
type: "policy"
slug: "policies/governance-actions-story-protocol"
freshness: "2025-05-08T20:38:00Z"
tags:
  - "decentralization"
  - "governance"
  - "smart-contracts"
owners: []
source_revision_ids:
  - "srcrev_2e7695a83903f0afdd9176ea6cab8629"
conflict_state: "none"
---

# Governance Actions for Story Protocol

## Summary

List of admin actions requiring governance for Story Protocol smart contracts, with the goal of progressive decentralization.

## Claims

- The governance process for on-chain actions involves scheduling, executing, and cancelling, with a minimum delay typically 5-7 days. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d) `source_document_id=srcdoc_4c424c6b92ffd0f4e53ae0a720da852a` `source_revision_id=srcrev_2e7695a83903f0afdd9176ea6cab8629` `chunk_id=srcchunk_9a1ba50a56d6bea5774a4a3876814a75` `native_locator=https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d` `source_timestamp=2025-05-08T20:38:00Z`
- The Security Multisig can cancel scheduled actions during the delay period. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d) `source_document_id=srcdoc_4c424c6b92ffd0f4e53ae0a720da852a` `source_revision_id=srcrev_2e7695a83903f0afdd9176ea6cab8629` `chunk_id=srcchunk_9a1ba50a56d6bea5774a4a3876814a75` `native_locator=https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d` `source_timestamp=2025-05-08T20:38:00Z`
- The TimelockController's updateDelay action modifies the minimum delay duration and must itself be scheduled. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d) `source_document_id=srcdoc_4c424c6b92ffd0f4e53ae0a720da852a` `source_revision_id=srcrev_2e7695a83903f0afdd9176ea6cab8629` `chunk_id=srcchunk_9a1ba50a56d6bea5774a4a3876814a75` `native_locator=https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d` `source_timestamp=2025-05-08T20:38:00Z`
- Ownable contracts have admin actions transferOwnership and renounceOwnership. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d) `source_document_id=srcdoc_4c424c6b92ffd0f4e53ae0a720da852a` `source_revision_id=srcrev_2e7695a83903f0afdd9176ea6cab8629` `chunk_id=srcchunk_9a1ba50a56d6bea5774a4a3876814a75` `native_locator=https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d` `source_timestamp=2025-05-08T20:38:00Z`
- ProxyAdmins have an upgradeTo action to change the implementation address of upgradeable contracts. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d) `source_document_id=srcdoc_4c424c6b92ffd0f4e53ae0a720da852a` `source_revision_id=srcrev_2e7695a83903f0afdd9176ea6cab8629` `chunk_id=srcchunk_9a1ba50a56d6bea5774a4a3876814a75` `native_locator=https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d` `source_timestamp=2025-05-08T20:38:00Z`
- IPTokenStaking has admin actions setMinStakeAmount, setFee, and setMinCommissionRate. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d) `source_document_id=srcdoc_4c424c6b92ffd0f4e53ae0a720da852a` `source_revision_id=srcrev_2e7695a83903f0afdd9176ea6cab8629` `chunk_id=srcchunk_9a1ba50a56d6bea5774a4a3876814a75` `native_locator=https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d` `source_timestamp=2025-05-08T20:38:00Z`

## Sources

- `source_document_id`: `srcdoc_4c424c6b92ffd0f4e53ae0a720da852a`
- `source_revision_id`: `srcrev_2e7695a83903f0afdd9176ea6cab8629`
- `source_url`: [Notion source](https://www.notion.so/Decision-points-for-governance-at-Story-1c7051299a54802ea6dee2345917333d)
