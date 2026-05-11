---
title: "Mainnet Protocol Design Discussions"
type: "project"
slug: "projects/mainnet-protocol-design-discussions"
freshness: "2024-03-30T18:28:00Z"
tags:
  - "decisions"
  - "mainnet"
  - "protocol-design"
owners: []
source_revision_ids:
  - "srcrev_4c45f47ff4b0ffb5e554c6911eb2516b"
conflict_state: "none"
---

# Mainnet Protocol Design Discussions

## Summary

A tracking document for pending and completed protocol design decisions for the RSI mainnet, organized by downstream effect and status.

## Claims

- Decision #2 (IPAccount as a Generic AA wallet or Restricted Module-based Wallet) is a pending decision with HIGH downstream effects. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #6 (Access Control Complexity) is a pending decision with HIGH downstream effects. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #8 (Relationship between License Registry and License NFT contracts) is a pending decision with MEDIUM downstream effects. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #10 (PIL Licensing Terms) is a pending decision with MEDIUM downstream effects. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #11 (License System) is a pending decision with MEDIUM downstream effects. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #12 (License Token 721 vs 1155) is a pending decision with MEDIUM downstream effects. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #5 (Protocol Upgradeability) is a pending decision with LOW downstream effects. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #9 (Meta-Tx on IPAccount-level or Protocol-level) is a pending decision with LOW downstream effects. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #1 (IPAccount: Open Data Access or Module-based Data System) is a completed decision. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #3 (Dynamic vs. Static Resolving Module Addresses for Function Calls) is a completed decision. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #4 (Core vs. Periphery Divisions) is a completed decision. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Decision #7 (Use of IPAssetRenderer) is a completed decision. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`
- Permissionless IP registration was completed without a proper context doc. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063) `source_document_id=srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce` `source_revision_id=srcrev_4c45f47ff4b0ffb5e554c6911eb2516b` `chunk_id=srcchunk_2ff5d339f1daacbb82f15c687aba0caf` `native_locator=https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063` `source_timestamp=2024-03-30T18:28:00Z`

## Open Questions

- What are the details of the unstructured/not ready yet items?
- What is the final resolution for pending HIGH downstream effect decisions?
- What is the status of the Royalty Module and Dispute Module?

## Related Pages

- `access-control-complexity`
- `core-vs-periphery-divisions`
- `dynamic-vs-static-module-addresses`
- `ipaccount-generic-aa-vs-restricted-module`
- `ipaccount-open-data-vs-module-based`
- `ipasset-renderer`
- `license-registry-vs-license-nft`
- `license-system`
- `license-token-721-vs-1155`
- `meta-tx-ipaccount-vs-protocol`
- `permissionless-ip-registration`
- `pil-licensing-terms`
- `protocol-upgradeability`

## Sources

- `source_document_id`: `srcdoc_30a412b7d605c04a4a6d5ef57d9c84ce`
- `source_revision_id`: `srcrev_4c45f47ff4b0ffb5e554c6911eb2516b`
- `source_url`: [Notion source](https://www.notion.so/Mainnet-Protocol-Design-Discussions-91a38c5a43c6437c97856211c6f05063)
