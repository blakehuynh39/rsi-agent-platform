---
title: "Aria Fractional Ownership Wrapper Proposal"
type: "concept"
slug: "concepts/fractional-ownership-wrapper-proposal"
freshness: "2026-01-09T22:24:26Z"
tags:
  - "aria"
  - "fractional-ownership"
  - "ip-asset"
  - "royalty-tokens"
  - "wrapper"
owners: []
source_revision_ids:
  - "srcrev_1fc2dd8ba221abb71aa96e519c65d6bc"
  - "srcrev_3ec0741c75c58e56d0f96b707d010a4b"
  - "srcrev_52535b13ca1e880d88128bba387eb4b5"
  - "srcrev_58818aff2e1df415c78563d1925a10e1"
  - "srcrev_7c215d7e386017818081e17270f1505e"
  - "srcrev_ec1aca0e2bdfdbae27c11468df9e4f8c"
conflict_state: "none"
---

# Aria Fractional Ownership Wrapper Proposal

## Summary

Design discussion on using a custom ERC-721 wrapper to enable fractional ownership and revenue distribution for IP Assets without protocol changes.

## Claims

- The ability to restrict royalty token transfer may require wrapped tokens, and hooks are being explored as a possible alternative. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_3ec0741c75c58e56d0f96b707d010a4b` `chunk_id=srcchunk_1a9c9c0b9ac1e4257ca26015b0c8bfb1` `native_locator=slack:C04T5307FNU:1767821445.274189:1767821445.274189` `source_timestamp=2026-01-07T21:30:45Z`
- Fractionalised tokens were initially assumed to be the same as royalty tokens, but later it was believed they are not using royalty tokens for this case. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_7c215d7e386017818081e17270f1505e` `chunk_id=srcchunk_dcfd4f2efae82a6a57295c7c882df3d5` `native_locator=slack:C04T5307FNU:1767821445.274189:1767822197.455179` `source_timestamp=2026-01-07T21:43:17Z`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_52535b13ca1e880d88128bba387eb4b5` `chunk_id=srcchunk_a88e359481c6f3a06687d7fe53786ae5` `native_locator=slack:C04T5307FNU:1767821445.274189:1767822376.598429` `source_timestamp=2026-01-07T21:46:16Z`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_ec1aca0e2bdfdbae27c11468df9e4f8c` `chunk_id=srcchunk_0bd5b60831fef3a203e768ab55fa8baf` `native_locator=slack:C04T5307FNU:1767821445.274189:1767841971.990529` `source_timestamp=2026-01-08T03:12:51Z`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_1fc2dd8ba221abb71aa96e519c65d6bc` `chunk_id=srcchunk_7dcb1ed07131a9db765b0fb37954b0e4` `native_locator=slack:C04T5307FNU:1767821445.274189:1767842028.388089` `source_timestamp=2026-01-08T03:13:48Z`
- A custom ERC-721 wrapper is proposed to handle fractional ownership and revenue distribution: the wrapper becomes the protocol-level IP owner, holds all royalty tokens, acts as the royalty vault, and includes payout logic based on fractional ERC-20 holdings. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3f446bd56f93223231001924baac66e6` `source_revision_id=srcrev_58818aff2e1df415c78563d1925a10e1` `chunk_id=srcchunk_2fd3e6bf75f88b936563564c27a46a3c` `native_locator=slack:C04T5307FNU:1767821445.274189:1767997466.942419` `source_timestamp=2026-01-09T22:24:26Z`

## Sources

- `source_document_id`: `srcdoc_3f446bd56f93223231001924baac66e6`
- `source_revision_id`: `srcrev_3ec0741c75c58e56d0f96b707d010a4b`
