---
title: "Block Explorer Vendor Evaluation"
type: "decision"
slug: "decisions/block-explorer-vendor-evaluation"
freshness: "2024-02-20T05:56:00Z"
tags:
  - "block-explorer"
  - "lore"
  - "nexandria"
  - "socialscan"
  - "vendor-evaluation"
owners: []
source_revision_ids:
  - "srcrev_160fd3d85d4b63446cccee3141e1f0f8"
conflict_state: "none"
---

# Block Explorer Vendor Evaluation

## Summary

An evaluation of three block explorer solutions—Nexandria, SocialScan, and Lore—to inform a selection decision. Nexandria offers a powerful Dev Mode and stats but lacks core features and has weak search. SocialScan provides the cleanest and fastest Etherscan-like experience with strong dev tooling but weaker ENS. Lore feels like a hybrid social wallet with AI, overbloated but feature-rich. SocialScan and Lore are the top contenders.

## Claims

- Nexandria pricing is $5-7k/year. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-1) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_6ab999c16a5b63a8edff254f02df831a` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-1` `source_timestamp=2024-02-20T05:56:00Z`
- Nexandria native ENS support is nice, but its search feature is weak (e.g., cannot find a coin called Capybara) and historical data/portfolio tracking is lacking. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-1) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_6ab999c16a5b63a8edff254f02df831a` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-1` `source_timestamp=2024-02-20T05:56:00Z`
- Nexandria's filtering by NFT/Tokens/DEX is useful but buggy and inaccurate. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-2) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_e8c3dfcfd19a5040b18a94822859292c` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-2` `source_timestamp=2024-02-20T05:56:00Z`
- Nexandria Dev Mode for transactions is very powerful (like built-in ethtx.info), and overall stats overview is helpful. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-2) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_e8c3dfcfd19a5040b18a94822859292c` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-2` `source_timestamp=2024-02-20T05:56:00Z`
- Nexandria is very open to building custom views and making any IPA attributes first-class citizens; custom dashboards may be supported. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_10aafb8893662aa2c643e4f559e9221c` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3` `source_timestamp=2024-02-20T05:56:00Z`
- SocialScan pricing is $10k/year without RPC (e.g., Caldera provides), or $20k/year with RPC provided. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_10aafb8893662aa2c643e4f559e9221c` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3` `source_timestamp=2024-02-20T05:56:00Z`
- SocialScan has the cleanest UI, is very fast, and offers built-in messaging support and solid token searching. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_10aafb8893662aa2c643e4f559e9221c` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3` `source_timestamp=2024-02-20T05:56:00Z`
- SocialScan ENS integration is janky—only supports lookup by currently registered name. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_10aafb8893662aa2c643e4f559e9221c` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3` `source_timestamp=2024-02-20T05:56:00Z`
- SocialScan explorer UX closely mimics familiar Etherscan but is cleaner, and they excel at L2-native data presentation. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_10aafb8893662aa2c643e4f559e9221c` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-3` `source_timestamp=2024-02-20T05:56:00Z`
- SocialScan supports contract verification, contract source code integration with scrolling & compilation settings, and DA fee tracking page. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-4) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_93661ca65cec449336176631ed7e8a0d` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-4` `source_timestamp=2024-02-20T05:56:00Z`
- SocialScan TX logs debugging works but UI is not amazing. Overall explorer experience is developer-friendly. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-4) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_93661ca65cec449336176631ed7e8a0d` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-4` `source_timestamp=2024-02-20T05:56:00Z`
- Lore feels like a hybrid between a block explorer app and an AI-powered social wallet, whereas SocialScan feels more like a pure, browser-optimized block explorer. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_b585b82d750eca78ad177fb452312f3a` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5` `source_timestamp=2024-02-20T05:56:00Z`
- Lore can create monitoring alerts for custom pipelines, bookmarks for wallets/tokens, user-stories/feeds, bubble maps for inflow/outflow, AI assistant for transaction reasoning, and a debug mode like ethtx.info. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_b585b82d750eca78ad177fb452312f3a` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5` `source_timestamp=2024-02-20T05:56:00Z`
- Lore combines Zerion, Nansen, and Etherscan into one tool, excelling for power-user speculators/analysts and day-to-day consumers who value social/AI features, but it feels overbloated. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_b585b82d750eca78ad177fb452312f3a` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5` `source_timestamp=2024-02-20T05:56:00Z`
- Nexandria is missing too many core dev features that SocialScan and Lore already provide out-of-the-box, making them a safer bet for feature completeness at launch. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_b585b82d750eca78ad177fb452312f3a` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5` `source_timestamp=2024-02-20T05:56:00Z`
- Lore pricing is very high (exact amount unknown), and its UX is considered unoptimized for a pure explorer; given the team is building their own FPA, they should not start with Lore's social-app approach. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-6) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_5ce77d80e6b589cea61e0737d12b0fc5` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-6` `source_timestamp=2024-02-20T05:56:00Z`
- Next steps: follow up with SocialScan on bookmarking/alerts, ethtx.info-style debugging, bubble maps, and custom branding extent. Also follow up with Lore to understand future direction. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-6) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_5ce77d80e6b589cea61e0737d12b0fc5` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-6` `source_timestamp=2024-02-20T05:56:00Z`
- Grading summary: SocialScan gets mostly 'A' across Search, Contract Verification, Readiness for SP-native branding, Data Comprehensiveness, Visualization, Misc Dev Tooling, but B- on Crypto-native Familiarity. Nexandria gets C on Search, N/A on Contract Verification, B on SP-native readiness, C on Data and Visualization, C on Dev Tooling, and B+ on Familiarity. Lore gets A- on Search, A on Contract Verification, B on SP-native readiness, A on Data, B on Visualization, A on Dev Tooling, and A+ on Familiarity. `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5) `source_document_id=srcdoc_9fd08eb17e8e695d80d229704822c01e` `source_revision_id=srcrev_160fd3d85d4b63446cccee3141e1f0f8` `chunk_id=srcchunk_b585b82d750eca78ad177fb452312f3a` `native_locator=https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a#chunk-5` `source_timestamp=2024-02-20T05:56:00Z`

## Open Questions

- Can SocialScan provide bookmarking/alerts, ethtx.info-style debugging, and bubble maps features similar to Lore?
- To what extent can SocialScan add custom branding?
- What is the exact pricing for Lore?

## Sources

- `source_document_id`: `srcdoc_9fd08eb17e8e695d80d229704822c01e`
- `source_revision_id`: `srcrev_160fd3d85d4b63446cccee3141e1f0f8`
- `source_url`: [Notion source](https://www.notion.so/Block-Explorer-Evaluation-b62fc2b9806546e89715af43535ff82a)
