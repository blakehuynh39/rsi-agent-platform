---
title: "IP Fractional Ownership Wrapper Pattern"
type: "concept"
slug: "concepts/ip-fractional-ownership-wrapper-pattern"
freshness: "2026-01-09T22:24:26Z"
tags:
  - "ERC-20"
  - "ERC-721"
  - "fractionalization"
  - "IPAsset"
  - "royalty"
  - "wrapper"
owners: []
source_revision_ids:
  - "srcrev_52535b13ca1e880d88128bba387eb4b5"
  - "srcrev_58818aff2e1df415c78563d1925a10e1"
  - "srcrev_7c215d7e386017818081e17270f1505e"
  - "srcrev_ec1aca0e2bdfdbae27c11468df9e4f8c"
conflict_state: "none"
---

# IP Fractional Ownership Wrapper Pattern

## Summary

A proposed pattern to enable fractional IP ownership and revenue distribution on Story Protocol while respecting the protocol's canonical IPAsset owner. It involves a custom ERC-721 wrapper that holds royalty tokens and distributes payouts to ERC-20 fractional holders.

## Claims

- Story Protocol recognizes only the IPAsset (its ERC-721 owner) as the canonical IP owner and does not natively recognize fractional ERC-20 holders created through TokenizerModule as owners. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_58818aff2e1df415c78563d1925a10e1` `chunk_id=srcchunk_2fd3e6bf75f88b936563564c27a46a3c` `native_locator=slack:C04T5307FNU:1767821445.274189:1767997466.942419` `source_timestamp=2026-01-09T22:24:26Z`
- The solution is a custom ERC-721 wrapper that becomes the protocol-level owner of the IP, holds all royalty tokens, acts as the royalty vault that collects all royalties on this IP, and includes payout logic based on the fractional ERC-20. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_58818aff2e1df415c78563d1925a10e1` `chunk_id=srcchunk_2fd3e6bf75f88b936563564c27a46a3c` `native_locator=slack:C04T5307FNU:1767821445.274189:1767997466.942419` `source_timestamp=2026-01-09T22:24:26Z`
- This custom IPA wrapper would be implemented by Aria and used when registering the IP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_58818aff2e1df415c78563d1925a10e1` `chunk_id=srcchunk_2fd3e6bf75f88b936563564c27a46a3c` `native_locator=slack:C04T5307FNU:1767821445.274189:1767997466.942419` `source_timestamp=2026-01-09T22:24:26Z`
- Aria is not using the royalty token for this case. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_ec1aca0e2bdfdbae27c11468df9e4f8c` `chunk_id=srcchunk_0bd5b60831fef3a203e768ab55fa8baf` `native_locator=slack:C04T5307FNU:1767821445.274189:1767841971.990529` `source_timestamp=2026-01-08T03:12:51Z`
- Fractionalised tokens are assumed to be royalty tokens, but this is unconfirmed. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_7c215d7e386017818081e17270f1505e` `chunk_id=srcchunk_dcfd4f2efae82a6a57295c7c882df3d5` `native_locator=slack:C04T5307FNU:1767821445.274189:1767822197.455179` `source_timestamp=2026-01-07T21:43:17Z`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_52535b13ca1e880d88128bba387eb4b5` `chunk_id=srcchunk_a88e359481c6f3a06687d7fe53786ae5` `native_locator=slack:C04T5307FNU:1767821445.274189:1767822376.598429` `source_timestamp=2026-01-07T21:46:16Z`

## Open Questions

- Are fractionalised tokens the same as royalty tokens in Aria's context, and if not, what exactly are they?

## Sources

- `source_document_id`: `srcdoc_3f446bd56f93223231001924baac66e6`
- `source_revision_id`: `srcrev_52535b13ca1e880d88128bba387eb4b5`
