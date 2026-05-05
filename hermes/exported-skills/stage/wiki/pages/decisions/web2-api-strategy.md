---
title: "Web2 API Strategy"
type: "decision"
slug: "decisions/web2-api-strategy"
freshness: "2026-05-05T06:32:49Z"
tags:
  - "api"
  - "custodian"
  - "magma"
  - "offchain"
  - "partners"
  - "royalties"
owners: []
source_revision_ids:
  - "srcrev_e4e5925f1a38fbee72b6ae5e9be723b6"
conflict_state: "none"
---

# Web2 API Strategy

## Summary

Strategy to build a web2 API for Magma and repurpose it for other partners, treating non-Network partners as 'offchain'. Addresses ownership transfers, payments, royalties, cross-chain integration, migration, UX, and custodian concerns.

## Claims

- Build web2 API for Magma, repurpose for other partners. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e) `source_document_id=srcdoc_ff94d5c72966449de767741397523d87` `source_revision_id=srcrev_e4e5925f1a38fbee72b6ae5e9be723b6` `chunk_id=srcchunk_ec049cb0422dee5da4bb326c981d00c7` `native_locator=https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e` `source_timestamp=2026-05-05T06:32:49Z`
- Treat all other partners not building on Network as 'offchain'. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e) `source_document_id=srcdoc_ff94d5c72966449de767741397523d87` `source_revision_id=srcrev_e4e5925f1a38fbee72b6ae5e9be723b6` `chunk_id=srcchunk_ec049cb0422dee5da4bb326c981d00c7` `native_locator=https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e` `source_timestamp=2026-05-05T06:32:49Z`
- Offchain ownership transfers require authentication via signatures/ID offchain. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e) `source_document_id=srcdoc_ff94d5c72966449de767741397523d87` `source_revision_id=srcrev_e4e5925f1a38fbee72b6ae5e9be723b6` `chunk_id=srcchunk_ec049cb0422dee5da4bb326c981d00c7` `native_locator=https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e` `source_timestamp=2026-05-05T06:32:49Z`
- Payments can use APIs supporting multiple forms like credit cards, avoiding bridging. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e) `source_document_id=srcdoc_ff94d5c72966449de767741397523d87` `source_revision_id=srcrev_e4e5925f1a38fbee72b6ae5e9be723b6` `chunk_id=srcchunk_ec049cb0422dee5da4bb326c981d00c7` `native_locator=https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e` `source_timestamp=2026-05-05T06:32:49Z`
- Royalties must be sent on Network only. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e) `source_document_id=srcdoc_ff94d5c72966449de767741397523d87` `source_revision_id=srcrev_e4e5925f1a38fbee72b6ae5e9be723b6` `chunk_id=srcchunk_ec049cb0422dee5da4bb326c981d00c7` `native_locator=https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e` `source_timestamp=2026-05-05T06:32:49Z`
- API will force more consolidation of state in our L2. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e) `source_document_id=srcdoc_ff94d5c72966449de767741397523d87` `source_revision_id=srcrev_e4e5925f1a38fbee72b6ae5e9be723b6` `chunk_id=srcchunk_ec049cb0422dee5da4bb326c981d00c7` `native_locator=https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e` `source_timestamp=2026-05-05T06:32:49Z`
- Backend can hold private keys or create a custodian solution for partners. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e) `source_document_id=srcdoc_ff94d5c72966449de767741397523d87` `source_revision_id=srcrev_e4e5925f1a38fbee72b6ae5e9be723b6` `chunk_id=srcchunk_ec049cb0422dee5da4bb326c981d00c7` `native_locator=https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e` `source_timestamp=2026-05-05T06:32:49Z`
- Investigate AA solution on backend using web2 authentication to unlock web3 account. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e) `source_document_id=srcdoc_ff94d5c72966449de767741397523d87` `source_revision_id=srcrev_e4e5925f1a38fbee72b6ae5e9be723b6` `chunk_id=srcchunk_ec049cb0422dee5da4bb326c981d00c7` `native_locator=https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e` `source_timestamp=2026-05-05T06:32:49Z`

## Open Questions

- How to authenticate new owner for offchain IPA transfers?
- How to handle royalty distribution and bridging for offchain partners?
- How to handle upgrade/migration to universal data layer?
- How to integrate IP created offchain with protocols on native chain?
- How to reduce double signing for cross-chain registration?
- Where does payment for license happen?

## Sources

- `source_document_id`: `srcdoc_ff94d5c72966449de767741397523d87`
- `source_revision_id`: `srcrev_e4e5925f1a38fbee72b6ae5e9be723b6`
- `source_url`: [Notion source](https://www.notion.so/Web2-API-Strategy-5d4c38c4037c4d18977fdfa03ab4890e)
