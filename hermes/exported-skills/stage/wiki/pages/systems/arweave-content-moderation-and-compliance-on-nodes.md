---
title: "Arweave Content Moderation and Compliance on Nodes"
type: "system"
slug: "systems/arweave-content-moderation-and-compliance-on-nodes"
freshness: "2025-04-24T16:32:00Z"
tags:
  - "arweave"
  - "compliance"
  - "content-moderation"
  - "content-policy"
  - "node-operations"
owners: []
source_revision_ids:
  - "srcrev_0c840507bd345f4ef093fd16001863c2"
conflict_state: "none"
---

# Arweave Content Moderation and Compliance on Nodes

## Summary

Arweave allows node operators to define content policies to filter transactions before storage, enabling compliance with local laws without forcing all nodes to store all data. The network uses a voting phase and temporary forks to resolve disagreements, and provides tools like Shepherd and ANS-106 for content management.

## Claims

- Each Arweave node operator can define what content they will accept and store, with no protocol-mandated obligation to store everything. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_f6434b7b0a75486b9c2a55a64d928d34` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1` `source_timestamp=2025-04-24T16:32:00Z`
- Content policies are rules that a node applies to incoming data, which can be any computation or filter such as substring checks, hash matching, or image analysis. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_f6434b7b0a75486b9c2a55a64d928d34` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1` `source_timestamp=2025-04-24T16:32:00Z`
- The reference Arweave implementation supports content policies via simple mechanisms like blocking transactions containing specific data substrings or matching hashes on a blacklist. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_f6434b7b0a75486b9c2a55a64d928d34` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1` `source_timestamp=2025-04-24T16:32:00Z`
- Before accepting a transaction into its local pool, each node scans it against its content policy; if it violates the policy, the node rejects it and does not forward it, effectively voting against the content. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_f6434b7b0a75486b9c2a55a64d928d34` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1` `source_timestamp=2025-04-24T16:32:00Z`
- A transaction being included in a mined block does not guarantee network-wide acceptance; Arweave’s consensus allows one extra block period as a confirmation buffer to account for content votes. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_f6434b7b0a75486b9c2a55a64d928d34` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1` `source_timestamp=2025-04-24T16:32:00Z`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_9897f8c1db83d5d4232a10229c1da2b7` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2` `source_timestamp=2025-04-24T16:32:00Z`
- If nodes disagree on content acceptability, the network can temporarily fork to resolve the conflict. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_f6434b7b0a75486b9c2a55a64d928d34` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-1` `source_timestamp=2025-04-24T16:32:00Z`
- Arweave provides a tool called Shepherd to help node operators manage content policies, and the mining guide recommends using a content policy to protect against illegal material. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_9897f8c1db83d5d4232a10229c1da2b7` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2` `source_timestamp=2025-04-24T16:32:00Z`
- Miners can add a transaction_blacklist_url pointing to public_shepherd.arweave.net to automatically fetch an updated list of NSFW or disallowed content to filter out. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_9897f8c1db83d5d4232a10229c1da2b7` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2` `source_timestamp=2025-04-24T16:32:00Z`
- Content filtering is opt-in; node operators can choose to use no filter or plug in community-maintained blacklists, and the filtering occurs before data is written to disk. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_9897f8c1db83d5d4232a10229c1da2b7` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2` `source_timestamp=2025-04-24T16:32:00Z`
- New nodes syncing the network apply their content policies, scanning and downloading only transactions that adhere to their policies, avoiding those they prefer not to store. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_9897f8c1db83d5d4232a10229c1da2b7` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-2` `source_timestamp=2025-04-24T16:32:00Z`
- Game-theoretic pressure incentivizes miners to converge on similar content standards to ensure their blocks are accepted by others. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-3) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_ee8fef1d559a55a02a20a1462a37f2af` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-3` `source_timestamp=2025-04-24T16:32:00Z`
- The content policy mechanism serves compliance and moderation, allowing node operators to avoid storing or serving data that would put them in legal jeopardy, such as illegal pornography or hate speech. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-3) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_ee8fef1d559a55a02a20a1462a37f2af` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-3` `source_timestamp=2025-04-24T16:32:00Z`
- A node’s content policy only affects its own storage and propagation; it does not purge data from the entire network, and data may still reside with nodes that choose to store it. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-3) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_ee8fef1d559a55a02a20a1462a37f2af` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-3` `source_timestamp=2025-04-24T16:32:00Z`
- Arweave has proposed ANS-106, a standard for voluntary post-publication removal requests, allowing someone to broadcast a request that miners purge or avoid storing given content, which node operators can independently decide to accept. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-3) `source_document_id=srcdoc_50a46597d757797dc96dda00db0ae0f7` `source_revision_id=srcrev_0c840507bd345f4ef093fd16001863c2` `chunk_id=srcchunk_ee8fef1d559a55a02a20a1462a37f2af` `native_locator=https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9#chunk-3` `source_timestamp=2025-04-24T16:32:00Z`

## Sources

- `source_document_id`: `srcdoc_50a46597d757797dc96dda00db0ae0f7`
- `source_revision_id`: `srcrev_0c840507bd345f4ef093fd16001863c2`
- `source_url`: [Notion source](https://www.notion.so/Arweave-Content-Moderation-and-Compliance-on-Nodes-1df051299a548078a2c1c7d31da5efc9)
