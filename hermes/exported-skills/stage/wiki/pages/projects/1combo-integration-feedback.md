---
title: "1Combo Integration Feedback"
type: "project"
slug: "projects/1combo-integration-feedback"
freshness: "2026-05-05T04:31:24Z"
tags:
  - "1combo"
  - "defi"
  - "integration"
  - "story-protocol"
owners:
  - "Kingter Wang"
source_revision_ids:
  - "srcrev_0689834e55113acf62fdafb77b25ee87"
conflict_state: "none"
---

# 1Combo Integration Feedback

## Summary

Feedback and questions from 1Combo regarding integration with Story Protocol, covering mint fees, royalty module, delegation, hook modules, license tokens, revenue distribution, derivative graph, CryptoPunk IP accounts, and signless registration.

## Claims

- The current mint fee receiver is licensorIpId, but 1Combo needs it to be an address. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- Story Protocol suggests using the RoyaltyModule to claim revenue directly to an EOA by delegating an IP and retaining all royalty tokens. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- The IP owner can revoke delegation at any time, but revenue claims are tied to the address holding RoyaltyTokens at the time of snapshot. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- There is no public method to burn license tokens; they can be sent to a burn address. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- Revenue from license sales currently can only be directly transferred to IP accounts. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- 1Combo's revenue model involves a bonding curve pool with taxes deducted per transaction for distribution. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- 1Combo asks if it's possible to direct license sale revenue to a contract for distribution. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- 1Combo asks how to record the association between derivatives and original creations in the global graph. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- CryptoPunk V1 is not a standard ERC721, and 1Combo asks if an IP account can be created for it. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- 1Combo wants to create IP accounts on behalf of NFT holders without requiring signatures, to simplify user operations for blue-chip NFT holders. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`
- 1Combo suggests a whitelisted admin address to register IPAccount/IPRoyaltyVault and approve Distribution Contract in one transaction without holder signature. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98) `source_document_id=srcdoc_b34fe2649ef35a6cf031fb4487804fea` `source_revision_id=srcrev_0689834e55113acf62fdafb77b25ee87` `chunk_id=srcchunk_3f72f3931bbfce8bda34852116a64008` `native_locator=https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98` `source_timestamp=2026-05-05T04:31:24Z`

## Open Questions

- Can we create IP accounts on behalf of NFT holders without requiring signatures, and if so, how?
- Can you include msg.sender in IHookModule.verify(...) to check if the license minter is from 1Combo?
- How can we record the association between derivatives and original creations in the global graph?
- Is it possible to create an IP account for CryptoPunk V1?
- Is it possible to direct the revenue from license sales to a contract for distribution?

## Sources

- `source_document_id`: `srcdoc_b34fe2649ef35a6cf031fb4487804fea`
- `source_revision_id`: `srcrev_0689834e55113acf62fdafb77b25ee87`
- `source_url`: [Notion source](https://www.notion.so/DeFi-Feedback-fad11d7816f142c8b80a01b2f42baa98)
