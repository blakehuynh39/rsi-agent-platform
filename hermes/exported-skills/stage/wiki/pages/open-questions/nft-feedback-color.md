---
title: "NFT Feedback (Color)"
type: "open_question"
slug: "open-questions/nft-feedback-color"
freshness: "2025-04-03T20:14:00Z"
tags:
  - "license-tokens"
  - "marketplace"
  - "nft"
  - "story-protocol"
  - "token-bound-accounts"
owners: []
source_revision_ids:
  - "srcrev_65195dda5174bdc6972d944c7aebf54a"
conflict_state: "none"
---

# NFT Feedback (Color)

## Summary

Open questions and feedback regarding NFT integration with Story Protocol, license tokens, and marketplace functionality for the Color project.

## Claims

- All License Tokens (LTs) are currently in a single collection. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84) `source_document_id=srcdoc_b8274bf4302d18121fcbec22d8b1e58b` `source_revision_id=srcrev_65195dda5174bdc6972d944c7aebf54a` `chunk_id=srcchunk_8bf6efc65f96b882e2dc7db32af4c4f4` `native_locator=https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84` `source_timestamp=2025-04-03T20:14:00Z`
- All License Tokens have a single default image. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84) `source_document_id=srcdoc_b8274bf4302d18121fcbec22d8b1e58b` `source_revision_id=srcrev_65195dda5174bdc6972d944c7aebf54a` `chunk_id=srcchunk_8bf6efc65f96b882e2dc7db32af4c4f4` `native_locator=https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84` `source_timestamp=2025-04-03T20:14:00Z`
- EIP-6551 includes a fraud prevention segment. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84) `source_document_id=srcdoc_b8274bf4302d18121fcbec22d8b1e58b` `source_revision_id=srcrev_65195dda5174bdc6972d944c7aebf54a` `chunk_id=srcchunk_8bf6efc65f96b882e2dc7db32af4c4f4` `native_locator=https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84` `source_timestamp=2025-04-03T20:14:00Z`
- A user can list an NFT with a token bound account that holds value, then front-run the purchaseListing call to empty the token bound account of value. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84) `source_document_id=srcdoc_b8274bf4302d18121fcbec22d8b1e58b` `source_revision_id=srcrev_65195dda5174bdc6972d944c7aebf54a` `chunk_id=srcchunk_8bf6efc65f96b882e2dc7db32af4c4f4` `native_locator=https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84` `source_timestamp=2025-04-03T20:14:00Z`
- If the marketplace takes ownership of the NFT while listed, IP modules may be locked automatically. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84) `source_document_id=srcdoc_b8274bf4302d18121fcbec22d8b1e58b` `source_revision_id=srcrev_65195dda5174bdc6972d944c7aebf54a` `chunk_id=srcchunk_8bf6efc65f96b882e2dc7db32af4c4f4` `native_locator=https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84` `source_timestamp=2025-04-03T20:14:00Z`
- Royalties from IP modules might be sent to the NFT's token bound account. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84) `source_document_id=srcdoc_b8274bf4302d18121fcbec22d8b1e58b` `source_revision_id=srcrev_65195dda5174bdc6972d944c7aebf54a` `chunk_id=srcchunk_8bf6efc65f96b882e2dc7db32af4c4f4` `native_locator=https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84` `source_timestamp=2025-04-03T20:14:00Z`

## Open Questions

- Are the IP token bound accounts able to hold ETH (or story equivalent IP module ownership) as per EIP-6551 fraud prevention?
- Are we able to lock all story IP modules when listed on Color?
- Do license NFTs have covers?
- Do we need to track listed NFT token bound accounts for value change for updating on frontend?
- If a user lists an NFT on Color, does this mean none of the IP actions work for the original owner? Does this affect where the royalties are sent?
- Is there an API or functionality provided by Story Protocol to determine all the associated children, root, parents of an IP asset for the purpose of rendering an IP graph?
- Should we allow people to customize the LT image or show a default image with stats?
- Should we be considering the token bound accounts (IP) holding value in any way (described to user / collated in total price)?
- Should we have an external contract that checks these commitments before validating transfer of NFTs?
- Should we list 'asset commitments' on the marketplace, and if changed the purchase is voided?
- What if I want to view all the LTs for one single IPA? or all “pudgy penguns” related LTs?

## Sources

- `source_document_id`: `srcdoc_b8274bf4302d18121fcbec22d8b1e58b`
- `source_revision_id`: `srcrev_65195dda5174bdc6972d944c7aebf54a`
- `source_url`: [Notion source](https://www.notion.so/NFT-Feedback-Color-bf68673b532a4e6d81820299e1905a84)
