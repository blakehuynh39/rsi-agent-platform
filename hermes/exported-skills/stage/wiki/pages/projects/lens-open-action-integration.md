---
title: "Lens Open Action Integration"
type: "project"
slug: "projects/lens-open-action-integration"
freshness: "2026-05-05T06:28:54Z"
tags:
  - "integration"
  - "lens"
  - "open-action"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_5a336353243defb4e6d12bc2a708da38"
conflict_state: "none"
---

# Lens Open Action Integration

## Summary

Integration of Lens Protocol open actions with Story Protocol for IP asset registration, including requirements, feedback, and open questions.

## Claims

- Reading materials for the integration include 'Create an open action' and 'Create smart post guide'. `claim:claim_3_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a) `source_document_id=srcdoc_1d23805ec756e17b4a435a141416d1dc` `source_revision_id=srcrev_5a336353243defb4e6d12bc2a708da38` `chunk_id=srcchunk_6c28bf4afee43ba8a700bd34e867bd31` `native_locator=https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a` `source_timestamp=2026-05-05T06:28:54Z`
- A demo of the integration is available at https://story-lens.vercel.app/. `claim:claim_3_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a) `source_document_id=srcdoc_1d23805ec756e17b4a435a141416d1dc` `source_revision_id=srcrev_5a336353243defb4e6d12bc2a708da38` `chunk_id=srcchunk_6c28bf4afee43ba8a700bd34e867bd31` `native_locator=https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a` `source_timestamp=2026-05-05T06:28:54Z`
- Feedback on the current implementation includes: why not directly registerIPAsset for the owner; no hardcode IPorg and mediaUrl, use a real one; call SPG instead of IPRegistry. `claim:claim_3_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a) `source_document_id=srcdoc_1d23805ec756e17b4a435a141416d1dc` `source_revision_id=srcrev_5a336353243defb4e6d12bc2a708da38` `chunk_id=srcchunk_6c28bf4afee43ba8a700bd34e867bd31` `native_locator=https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a` `source_timestamp=2026-05-05T06:28:54Z`
- Open Action requirements: simplify the Open Action interface for Orb to call; when a publication is init, call SP Open Action to create IPOrg, configure IPOrg license, create IPA, create license NFT; return IPOrg ID, IPA global ID and local ID; on FE, show IPA ID and link to Opensea. `claim:claim_3_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a) `source_document_id=srcdoc_1d23805ec756e17b4a435a141416d1dc` `source_revision_id=srcrev_5a336353243defb4e6d12bc2a708da38` `chunk_id=srcchunk_6c28bf4afee43ba8a700bd34e867bd31` `native_locator=https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a` `source_timestamp=2026-05-05T06:28:54Z`
- Questions to Lens: any way to store the license verification status in lens; open action repo? `claim:claim_3_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a) `source_document_id=srcdoc_1d23805ec756e17b4a435a141416d1dc` `source_revision_id=srcrev_5a336353243defb4e6d12bc2a708da38` `chunk_id=srcchunk_6c28bf4afee43ba8a700bd34e867bd31` `native_locator=https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a` `source_timestamp=2026-05-05T06:28:54Z`
- Question to us: Lens publication is not NFT, but our IPA is NFT. What happens if the IPA owner transfers the IPA? `claim:claim_3_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a) `source_document_id=srcdoc_1d23805ec756e17b4a435a141416d1dc` `source_revision_id=srcrev_5a336353243defb4e6d12bc2a708da38` `chunk_id=srcchunk_6c28bf4afee43ba8a700bd34e867bd31` `native_locator=https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a` `source_timestamp=2026-05-05T06:28:54Z`

## Open Questions

- Any way to store the license verification status in lens?
- Lens publication is not NFT, but our IPA is NFT. What happens if the IPA owner transfers the IPA?
- Open action repo?

## Related Pages

- `projects/lens-protocol-deep-dive`
- `projects/lens-v2`

## Sources

- `source_document_id`: `srcdoc_1d23805ec756e17b4a435a141416d1dc`
- `source_revision_id`: `srcrev_5a336353243defb4e6d12bc2a708da38`
- `source_url`: [Notion source](https://www.notion.so/Lens-0c9f5a1014564dadad4f578ebfa92d0a)
