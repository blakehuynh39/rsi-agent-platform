---
title: "Vote: Levels of IP Asset for Core-level Protocol Interactions"
type: "decision"
slug: "decisions/vote-ip-asset-interaction-levels"
freshness: "2026-05-05T06:31:41Z"
tags:
  - "core-protocol"
  - "interaction-levels"
  - "ip-asset"
  - "vote"
owners: []
source_revision_ids:
  - "srcrev_e1772e959badb06b385fc1b7dd2bb468"
conflict_state: "none"
---

# Vote: Levels of IP Asset for Core-level Protocol Interactions

## Summary

This page outlines two options for handling IP Asset interactions at the core protocol level: Option 1 treats IP as a simple identifier without accounts, while Option 2 gives each IP its own account (IPAccount).

## Claims

- Intra-protocol interactions are divided into core-level (IP Asset with Registries, Royalty & Dispute Module) and everything-else (periphery and external) interactions. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a#chunk-1) `source_document_id=srcdoc_239400b28ea47152f8a95785c4d9abe1` `source_revision_id=srcrev_e1772e959badb06b385fc1b7dd2bb468` `chunk_id=srcchunk_e48f459c90529ff4a8b76c791abeccf7` `native_locator=https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a#chunk-1` `source_timestamp=2026-05-05T06:31:41Z`
- Option 1: IP is a simple identifier without an account at the core level; core-level data is stored in the IPAssetRegistry, and assets like RNFTs are claimable by the owner of the IP's NFT. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a#chunk-1) `source_document_id=srcdoc_239400b28ea47152f8a95785c4d9abe1` `source_revision_id=srcrev_e1772e959badb06b385fc1b7dd2bb468` `chunk_id=srcchunk_e48f459c90529ff4a8b76c791abeccf7` `native_locator=https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a#chunk-1` `source_timestamp=2026-05-05T06:31:41Z`
- Option 2: Each IP has an IPAccount within Story Protocol, meaning all IPs are deployed accounts with a single entry point and a generic execute function to call core and community modules. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a#chunk-1) `source_document_id=srcdoc_239400b28ea47152f8a95785c4d9abe1` `source_revision_id=srcrev_e1772e959badb06b385fc1b7dd2bb468` `chunk_id=srcchunk_e48f459c90529ff4a8b76c791abeccf7` `native_locator=https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a#chunk-1` `source_timestamp=2026-05-05T06:31:41Z`
- Modules hold data for IPs; for instance, the Royalty Module holds balanceOf for Royalty NFTs ownership by IPs, regardless of whether IP is an account or a record. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a#chunk-2) `source_document_id=srcdoc_239400b28ea47152f8a95785c4d9abe1` `source_revision_id=srcrev_e1772e959badb06b385fc1b7dd2bb468` `chunk_id=srcchunk_c9b03e03bb1bded7aed0929abfab94e1` `native_locator=https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a#chunk-2` `source_timestamp=2026-05-05T06:31:41Z`

## Sources

- `source_document_id`: `srcdoc_239400b28ea47152f8a95785c4d9abe1`
- `source_revision_id`: `srcrev_e1772e959badb06b385fc1b7dd2bb468`
- `source_url`: [Notion source](https://www.notion.so/Vote-Levels-of-IP-Asset-for-Core-level-Protocol-Interactions-432b7099ee4e4f02a9a4deb80c94492a)
