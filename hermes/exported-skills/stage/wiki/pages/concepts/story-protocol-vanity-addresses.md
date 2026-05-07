---
title: "Story Protocol Vanity Addresses"
type: "concept"
slug: "concepts/story-protocol-vanity-addresses"
freshness: "2024-04-02T20:56:00Z"
tags:
  - "ip-accounts"
  - "l2-infra"
  - "precompiles"
  - "story-network"
  - "vanity-addresses"
owners: []
source_revision_ids:
  - "srcrev_7db83da9a27bc6f4d0cf12f4de199310"
conflict_state: "none"
---

# Story Protocol Vanity Addresses

## Summary

Proposal for using vanity addresses to evangelize Story Network, including standard precompiles, L2 infrastructure contracts, popular tokens, and Story Protocol ecosystem addresses with a 0x19 prefix resembling 'IP'.

## Claims

- The proposal aims to help evangelize aspects of the protocol using precompiled addresses or proxy addresses that redirect to vanity addresses. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c) `source_document_id=srcdoc_0ba185c59d8dcc9fa4a34d0bf5833448` `source_revision_id=srcrev_7db83da9a27bc6f4d0cf12f4de199310` `chunk_id=srcchunk_6e737b1999993f961efe46aa6295ac8e` `native_locator=https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c` `source_timestamp=2024-04-02T20:56:00Z`
- Standard precompiles (range 0x00 to 0xFF) must be deployed on all chains, including ECRecover (0x01), SHA-256 (0x02), RIPEMD-160 (0x03), Identity (0x04). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c) `source_document_id=srcdoc_0ba185c59d8dcc9fa4a34d0bf5833448` `source_revision_id=srcrev_7db83da9a27bc6f4d0cf12f4de199310` `chunk_id=srcchunk_6e737b1999993f961efe46aa6295ac8e` `native_locator=https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c` `source_timestamp=2024-04-02T20:56:00Z`
- L2 Infra-Specific addresses (range 0x99..00 to 0x99..FF) include L2ToL1MessagePasser (0x990000…00), AddressManager (0x990000…01), L1CrossDomainMessenger (0x990000…02). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c) `source_document_id=srcdoc_0ba185c59d8dcc9fa4a34d0bf5833448` `source_revision_id=srcrev_7db83da9a27bc6f4d0cf12f4de199310` `chunk_id=srcchunk_6e737b1999993f961efe46aa6295ac8e` `native_locator=https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c` `source_timestamp=2024-04-02T20:56:00Z`
- Popular Tokens addresses (prefix 0x88) include WETH (0x88…01) and USDC (0x88…02). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c) `source_document_id=srcdoc_0ba185c59d8dcc9fa4a34d0bf5833448` `source_revision_id=srcrev_7db83da9a27bc6f4d0cf12f4de199310` `chunk_id=srcchunk_6e737b1999993f961efe46aa6295ac8e` `native_locator=https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c` `source_timestamp=2024-04-02T20:56:00Z`
- Story Protocol Ecosystem Addresses (range 0x19..00 to 0x19..FF) use prefix 0x19 resembling 'IP', including IP Asset Registry (0x19…00), License Registry (0x19…01), Governance Contract (0x19…03), Access Controller (0x19…04), Dispute Module (0x19…05), Licensing Module (0x19...06), Royalty Module (0x19…07). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c) `source_document_id=srcdoc_0ba185c59d8dcc9fa4a34d0bf5833448` `source_revision_id=srcrev_7db83da9a27bc6f4d0cf12f4de199310` `chunk_id=srcchunk_6e737b1999993f961efe46aa6295ac8e` `native_locator=https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c` `source_timestamp=2024-04-02T20:56:00Z`
- IP Account Addresses (range 0x19..100 to 0x19..1FF) are proposed for IP Accounts, which are IP Data/ACL contracts created and bound to every IP (any NFT). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c) `source_document_id=srcdoc_0ba185c59d8dcc9fa4a34d0bf5833448` `source_revision_id=srcrev_7db83da9a27bc6f4d0cf12f4de199310` `chunk_id=srcchunk_6e737b1999993f961efe46aa6295ac8e` `native_locator=https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c` `source_timestamp=2024-04-02T20:56:00Z`

## Sources

- `source_document_id`: `srcdoc_0ba185c59d8dcc9fa4a34d0bf5833448`
- `source_revision_id`: `srcrev_7db83da9a27bc6f4d0cf12f4de199310`
- `source_url`: [Notion source](https://www.notion.so/Story-Protocol-Vanity-Addresses-81ec46166e814d5890d493b3e1cff38c)
