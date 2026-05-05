---
title: "Story Supernet"
type: "concept"
slug: "concepts/story-supernet"
freshness: "2026-05-05T06:32:31Z"
tags:
  - "cross-chain"
  - "design"
  - "ipa"
  - "story-supernet"
owners: []
source_revision_ids:
  - "srcrev_23c98da7dced924ec0f0607a8b8804d5"
conflict_state: "none"
---

# Story Supernet

## Summary

Design exploration for Story Supernet covering user actions, storage, API approach, and two deployment scenarios (single-chain IPA vs multichain).

## Claims

- Initial stage goals include probing early design choices, web2 integration compatibility, lock mechanisms to prevent changes after agreement, UX, and finality times. `claim:claim_ss_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67) `source_document_id=srcdoc_b9aca71621a9e442186876fb3d00b6ad` `source_revision_id=srcrev_23c98da7dced924ec0f0607a8b8804d5` `chunk_id=srcchunk_dd2729847044635bf25f9b8bebe0fbaf` `native_locator=https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67` `source_timestamp=2026-05-05T06:32:31Z`
- User actions include creating IPA with/without NFT, creating and adding policies, minting licenses, creating derivative IP, paying/claiming royalties, and raising disputes. `claim:claim_ss_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67) `source_document_id=srcdoc_b9aca71621a9e442186876fb3d00b6ad` `source_revision_id=srcrev_23c98da7dced924ec0f0607a8b8804d5` `chunk_id=srcchunk_dd2729847044635bf25f9b8bebe0fbaf` `native_locator=https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67` `source_timestamp=2026-05-05T06:32:31Z`
- Storage considerations: IPGraph should reside on one chain due to finality issues; royalty payment tree, templates/policies, and registries are also stored. `claim:claim_ss_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67) `source_document_id=srcdoc_b9aca71621a9e442186876fb3d00b6ad` `source_revision_id=srcrev_23c98da7dced924ec0f0607a8b8804d5` `chunk_id=srcchunk_dd2729847044635bf25f9b8bebe0fbaf` `native_locator=https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67` `source_timestamp=2026-05-05T06:32:31Z`
- Currently, web2 cross-chain NFT listing support is gated by an API; Story Network is otherwise inaccessible. `claim:claim_ss_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67) `source_document_id=srcdoc_b9aca71621a9e442186876fb3d00b6ad` `source_revision_id=srcrev_23c98da7dced924ec0f0607a8b8804d5` `chunk_id=srcchunk_dd2729847044635bf25f9b8bebe0fbaf` `native_locator=https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67` `source_timestamp=2026-05-05T06:32:31Z`
- Web2 API approach: users call an API to register an NFT as an IPA; the API creates an IP Account tied to the NFT and an AA account, checks NFT ownership before executing actions like adding policies, minting licenses, paying/claiming royalties, and linking IPAs. `claim:claim_ss_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67) `source_document_id=srcdoc_b9aca71621a9e442186876fb3d00b6ad` `source_revision_id=srcrev_23c98da7dced924ec0f0607a8b8804d5` `chunk_id=srcchunk_dd2729847044635bf25f9b8bebe0fbaf` `native_locator=https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67` `source_timestamp=2026-05-05T06:32:31Z`
- Scenario 1 (SN holds all IPA): all IPA logic resides on Story Network; considerations include liquidity of license/royalty tokens, user experience from source chains, and the need for oracles to verify underlying NFT state. Lifecycle involves verifying ownership, requiring a user wallet on SN, sending intent to mint licenses with fees, and performing all registrations and royalty operations on one chain to reduce bridging needs. `claim:claim_ss_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67) `source_document_id=srcdoc_b9aca71621a9e442186876fb3d00b6ad` `source_revision_id=srcrev_23c98da7dced924ec0f0607a8b8804d5` `chunk_id=srcchunk_dd2729847044635bf25f9b8bebe0fbaf` `native_locator=https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67` `source_timestamp=2026-05-05T06:32:31Z`
- Scenario 2 (multichain): SP deployed on each source chain; considerations include finality time, synchronization, maintaining a global IP Graph (onchain/offchain), a global IP registry to prevent double spends, supporting IPA transfer across chains, and handling cross-chain calls for verifying ownership, minting licenses, registering derivatives, and pulling royalties. `claim:claim_ss_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67) `source_document_id=srcdoc_b9aca71621a9e442186876fb3d00b6ad` `source_revision_id=srcrev_23c98da7dced924ec0f0607a8b8804d5` `chunk_id=srcchunk_dd2729847044635bf25f9b8bebe0fbaf` `native_locator=https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67` `source_timestamp=2026-05-05T06:32:31Z`

## Open Questions

- Bridge options
- Finality times

## Related Pages

- `projects/multichain-story-supernet`

## Sources

- `source_document_id`: `srcdoc_b9aca71621a9e442186876fb3d00b6ad`
- `source_revision_id`: `srcrev_23c98da7dced924ec0f0607a8b8804d5`
- `source_url`: [Notion source](https://www.notion.so/Story-Supernet-d9a158563a4a46cba3f8a583c7478b67)
