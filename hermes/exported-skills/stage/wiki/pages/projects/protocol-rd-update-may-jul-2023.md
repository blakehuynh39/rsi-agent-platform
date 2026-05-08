---
title: "Protocol R\u0026D Update (May–July 2023)"
type: "project"
slug: "projects/protocol-rd-update-may-jul-2023"
freshness: "2024-12-03T05:20:00Z"
tags: []
owners: []
source_revision_ids:
  - "srcrev_24efa49f1987840c4072cb1e4a5b4db9"
  - "srcrev_3c42c4d52451b5683bf003b3b6c81b24"
conflict_state: "none"
---

# Protocol R&D Update (May–July 2023)

## Summary

Summary of protocol research, development milestones, module discussions, repository activities, and Beta release design ideas from May to July 2023.

## Claims

- A demo was scheduled for July 3, 2023, focusing on creating characters and stories and linking them. `claim:jun2023-demo` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46) `source_document_id=srcdoc_8ea92cc7e4133b726e7ac0f9714063b0` `source_revision_id=srcrev_3c42c4d52451b5683bf003b3b6c81b24` `chunk_id=srcchunk_dfc5207ec567dc6579a9e3b049d09112` `native_locator=https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46` `source_timestamp=2024-12-03T05:20:00Z`
- In June 2023, protocol module discussions covered IP Account (EIP-6551), Data Access Module, and Storage Module. `claim:jun2023-module-discussions` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46) `source_document_id=srcdoc_8ea92cc7e4133b726e7ac0f9714063b0` `source_revision_id=srcrev_3c42c4d52451b5683bf003b3b6c81b24` `chunk_id=srcchunk_dfc5207ec567dc6579a9e3b049d09112` `native_locator=https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46` `source_timestamp=2024-12-03T05:20:00Z`
- In June 2023, protocol repository contributions included Contributor Attribution (PR #16), Story Block (PR #9), and Franchise Registry (PR #7). `claim:jun2023-repo-contributions` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46) `source_document_id=srcdoc_8ea92cc7e4133b726e7ac0f9714063b0` `source_revision_id=srcrev_3c42c4d52451b5683bf003b3b6c81b24` `chunk_id=srcchunk_dfc5207ec567dc6579a9e3b049d09112` `native_locator=https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46` `source_timestamp=2024-12-03T05:20:00Z`
- As of June 2023, the protocol roadmap included: PoC release end-August, Alpha release mid-October, Beta release mid-December. `claim:jun2023-roadmap` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46) `source_document_id=srcdoc_8ea92cc7e4133b726e7ac0f9714063b0` `source_revision_id=srcrev_3c42c4d52451b5683bf003b3b6c81b24` `chunk_id=srcchunk_dfc5207ec567dc6579a9e3b049d09112` `native_locator=https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46` `source_timestamp=2024-12-03T05:20:00Z`
- In May 2023, protocol data model exploration and design discussions included open questions, registry design, on-chain vs off-chain considerations, and an ENS resolver. `claim:may2023-exploration` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46) `source_document_id=srcdoc_8ea92cc7e4133b726e7ac0f9714063b0` `source_revision_id=srcrev_3c42c4d52451b5683bf003b3b6c81b24` `chunk_id=srcchunk_dfc5207ec567dc6579a9e3b049d09112` `native_locator=https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46` `source_timestamp=2024-12-03T05:20:00Z`
- In May 2023, the roadmap planned Alpha release end-July and Beta release mid-October. `claim:may2023-roadmap-superseded` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46) `source_document_id=srcdoc_8ea92cc7e4133b726e7ac0f9714063b0` `source_revision_id=srcrev_3c42c4d52451b5683bf003b3b6c81b24` `chunk_id=srcchunk_dfc5207ec567dc6579a9e3b049d09112` `native_locator=https://www.notion.so/Protocol-R-D-Update-42a06b67c39e4d7f81ff950e1c6b3d46` `source_timestamp=2024-12-03T05:20:00Z`
- During IPA registration, the IPA data is first written to the IPA registry, then an IPA NFT is minted. `claim:beta-design-ipa-registration` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91) `source_document_id=srcdoc_0fddaeecc02fc71b9fc66371fd61438a` `source_revision_id=srcrev_24efa49f1987840c4072cb1e4a5b4db9` `chunk_id=srcchunk_da7c25655881c5641be88ccb05595b87` `native_locator=https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91` `source_timestamp=2024-02-01T22:36:00Z`
- Minting IPA NFT is a pluggable module; the protocol provides a default implementation, but it can be customized via a minting hook (IMintHook) to add IPOrg-specific data. `claim:beta-design-minting-pluggable` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91) `source_document_id=srcdoc_0fddaeecc02fc71b9fc66371fd61438a` `source_revision_id=srcrev_24efa49f1987840c4072cb1e4a5b4db9` `chunk_id=srcchunk_da7c25655881c5641be88ccb05595b87` `native_locator=https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91` `source_timestamp=2024-02-01T22:36:00Z`
- The IPA record is the ground truth; the IPA NFT is just one view of the original IPA. `claim:beta-design-ipa-record-truth` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91) `source_document_id=srcdoc_0fddaeecc02fc71b9fc66371fd61438a` `source_revision_id=srcrev_24efa49f1987840c4072cb1e4a5b4db9` `chunk_id=srcchunk_da7c25655881c5641be88ccb05595b87` `native_locator=https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91` `source_timestamp=2024-02-01T22:36:00Z`
- An IPA owner can mint multiple IPA NFTs in different media or presentations, akin to posting a video to multiple platforms. `claim:beta-design-multiple-nfts` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91) `source_document_id=srcdoc_0fddaeecc02fc71b9fc66371fd61438a` `source_revision_id=srcrev_24efa49f1987840c4072cb1e4a5b4db9` `chunk_id=srcchunk_da7c25655881c5641be88ccb05595b87` `native_locator=https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91` `source_timestamp=2024-02-01T22:36:00Z`
- Minting an IPA NFT on an IPOrg is like licensing out the IPA to that platform/application, subject to its license agreement. `claim:beta-design-licensing-analogy` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91) `source_document_id=srcdoc_0fddaeecc02fc71b9fc66371fd61438a` `source_revision_id=srcrev_24efa49f1987840c4072cb1e4a5b4db9` `chunk_id=srcchunk_da7c25655881c5641be88ccb05595b87` `native_locator=https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91` `source_timestamp=2024-02-01T22:36:00Z`
- Royalty payments go to two parties: the IPA owner and the IPOrg pool; distribution within the pool is handled by a separate module or protocol. `claim:beta-design-royalty-split` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91) `source_document_id=srcdoc_0fddaeecc02fc71b9fc66371fd61438a` `source_revision_id=srcrev_24efa49f1987840c4072cb1e4a5b4db9` `chunk_id=srcchunk_da7c25655881c5641be88ccb05595b87` `native_locator=https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91` `source_timestamp=2024-02-01T22:36:00Z`
- In the EU use case, the EU owner registered a world bible as the first IPA and set license terms requiring an LNFT to register new IPAs; creators must obtain an LNFT to contribute. `claim:beta-design-eu-usecase` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91) `source_document_id=srcdoc_0fddaeecc02fc71b9fc66371fd61438a` `source_revision_id=srcrev_24efa49f1987840c4072cb1e4a5b4db9` `chunk_id=srcchunk_da7c25655881c5641be88ccb05595b87` `native_locator=https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91` `source_timestamp=2024-02-01T22:36:00Z`
- In the TOMO use case, the TOMO team created an IPOrg, set license terms, and will mint IPAs for creators uploading short videos; remixing could automatically mint LNFTs and register new IPAs, but this may be expensive and require an L2. `claim:beta-design-tomo-usecase` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91) `source_document_id=srcdoc_0fddaeecc02fc71b9fc66371fd61438a` `source_revision_id=srcrev_24efa49f1987840c4072cb1e4a5b4db9` `chunk_id=srcchunk_da7c25655881c5641be88ccb05595b87` `native_locator=https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91` `source_timestamp=2024-02-01T22:36:00Z`

## Sources

- `source_document_id`: `srcdoc_0fddaeecc02fc71b9fc66371fd61438a`
- `source_revision_id`: `srcrev_24efa49f1987840c4072cb1e4a5b4db9`
- `source_url`: [Notion source](https://www.notion.so/Beta-Release-7ceb51f7eebc4183a7df83b0941d0b91)
