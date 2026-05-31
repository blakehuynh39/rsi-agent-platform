---
title: "IPA Expiration Design"
type: "concept"
slug: "concepts/ipa-expiration-design"
freshness: "2024-11-12T19:13:00Z"
tags:
  - "expiration"
  - "ipa"
  - "licensing"
  - "pil"
owners: []
source_revision_ids:
  - "srcrev_b7af3533e1e32eb1d960ae07c6e7b236"
conflict_state: "none"
---

# IPA Expiration Design

## Summary

High-level design for implementing expiration on-chain for IP Assets (IPAs), including two broad approaches: License Token expiration and IPA expiration. Also covers shared logic with dispute handling for child IPAs.

## Claims

- There are two broad approaches for expiration: License Token expiration and IPA expiration. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-1) `source_document_id=srcdoc_66f757bdc9ecfa07d6f8a314c33ec8a5` `source_revision_id=srcrev_b7af3533e1e32eb1d960ae07c6e7b236` `chunk_id=srcchunk_0e6010589c6723a70be5ce1332c6e852` `native_locator=https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-1` `source_timestamp=2024-11-12T19:13:00Z`
- License Token expiration: START is the time the License Token is minted. If unused after X period, the token no longer functions and cannot be burned to link to a parent IPA. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-1) `source_document_id=srcdoc_66f757bdc9ecfa07d6f8a314c33ec8a5` `source_revision_id=srcrev_b7af3533e1e32eb1d960ae07c6e7b236` `chunk_id=srcchunk_0e6010589c6723a70be5ce1332c6e852` `native_locator=https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-1` `source_timestamp=2024-11-12T19:13:00Z`
- IPA expiration has two options for START: time the License Token is minted, or time the derivative IPA is minted and linked to the parent IPA. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-1) `source_document_id=srcdoc_66f757bdc9ecfa07d6f8a314c33ec8a5` `source_revision_id=srcrev_b7af3533e1e32eb1d960ae07c6e7b236` `chunk_id=srcchunk_0e6010589c6723a70be5ce1332c6e852` `native_locator=https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-1` `source_timestamp=2024-11-12T19:13:00Z`
- After expiration, the derivative IPA should stop accepting royalties and no longer be able to mint new licenses. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-1) `source_document_id=srcdoc_66f757bdc9ecfa07d6f8a314c33ec8a5` `source_revision_id=srcrev_b7af3533e1e32eb1d960ae07c6e7b236` `chunk_id=srcchunk_0e6010589c6723a70be5ce1332c6e852` `native_locator=https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-1` `source_timestamp=2024-11-12T19:13:00Z`
- Shared logic between dispute and expiration for children of the affected IPA: remove staked $IP (slashed in dispute, returned in expiration), prevent royalties and license minting for both parent and children, and return staked $IP from children. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-2) `source_document_id=srcdoc_66f757bdc9ecfa07d6f8a314c33ec8a5` `source_revision_id=srcrev_b7af3533e1e32eb1d960ae07c6e7b236` `chunk_id=srcchunk_2c2fbf7bddeab21a8d3a2fdc004e496d` `native_locator=https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb#chunk-2` `source_timestamp=2024-11-12T19:13:00Z`

## Open Questions

- Should expiring assets be allowed to have children?
- What happens to children IPAs after the parent IPA expires?

## Sources

- `source_document_id`: `srcdoc_66f757bdc9ecfa07d6f8a314c33ec8a5`
- `source_revision_id`: `srcrev_b7af3533e1e32eb1d960ae07c6e7b236`
- `source_url`: [Notion source](https://www.notion.so/Expiration-Term-PIL-9e1c3b3945b74d10987dae6eafc245bb)
