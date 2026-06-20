---
title: "Indexing Infrastructure"
type: "system"
slug: "systems/indexing-infrastructure"
freshness: "2026-05-29T02:14:15Z"
tags:
  - "blockscout"
  - "goldsky"
  - "indexer"
  - "temporal"
owners:
  - "U0871SH0FNZ"
source_revision_ids:
  - "srcrev_1de58e005e5b2d70d7c3291013390e95"
  - "srcrev_39629e0dfa197e6083d3eea91589ea17"
  - "srcrev_4e1b8cbb33bd7ea6c15c75455d510473"
  - "srcrev_80dc9f71d0e5c7d1a974313a7a2c6d77"
  - "srcrev_92444925644989ffa21f72c0bbf30217"
  - "srcrev_db4842dcca6ea8976ec3e0f1ff21b4db"
  - "srcrev_e0c66ad87fbddce1e81bf2dfd67aca3a"
  - "srcrev_ee7390f53091756a88c4e119c178dc57"
  - "srcrev_fa7927f8ce05c1ac5dd499bd82b53fcb"
conflict_state: "none"
---

# Indexing Infrastructure

## Summary

Overview of indexing services used by RSI, including an internal Temporal‑based indexer, BlockScout, and the external Goldsky service.

## Claims

- RSI runs its own indexer using Temporal. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_80dc9f71d0e5c7d1a974313a7a2c6d77` `chunk_id=srcchunk_d147b49e9ac4f0b0f5c56d361a9f377e` `native_locator=slack:C0547N89JUB:1780011492.577489:1780011492.577489` `source_timestamp=2026-05-28T23:38:12Z`
- The internal indexer powers RSI's own API. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_ee7390f53091756a88c4e119c178dc57` `chunk_id=srcchunk_af3a1588e940fec729d64648fa5dbaf3` `native_locator=slack:C0547N89JUB:1780011492.577489:1780015553.601919` `source_timestamp=2026-05-29T00:45:53Z`
- BlockScout provides an indexer that is also used by RSI. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_92444925644989ffa21f72c0bbf30217` `chunk_id=srcchunk_ca63c231e75972732bee670cf947956c` `native_locator=slack:C0547N89JUB:1780011492.577489:1780011628.697779` `source_timestamp=2026-05-28T23:40:28Z`
- Goldsky is used by third‑party DEXes for indexing relevant events. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_fa7927f8ce05c1ac5dd499bd82b53fcb` `chunk_id=srcchunk_e67a9936b6d1b3245bab856ccbdf73b1` `native_locator=slack:C0547N89JUB:1780011492.577489:1780015542.246649` `source_timestamp=2026-05-29T00:45:42Z`
- Goldsky is a self‑serve service. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_4e1b8cbb33bd7ea6c15c75455d510473` `chunk_id=srcchunk_e767f52efa1c0e0d9266c3d19a334514` `native_locator=slack:C0547N89JUB:1780011492.577489:1780015565.765329` `source_timestamp=2026-05-29T00:46:05Z`
- The specific DEXes using Goldsky are not remembered; they were Story's partners, possibly including Piper X. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_1de58e005e5b2d70d7c3291013390e95` `chunk_id=srcchunk_7f4a33e6c8d842e7a978e35fb833afac` `native_locator=slack:C0547N89JUB:1780011492.577489:1780017504.860459` `source_timestamp=2026-05-29T01:18:24Z`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_db4842dcca6ea8976ec3e0f1ff21b4db` `chunk_id=srcchunk_709d0333b4070cd892dfb62d74ca0ee2` `native_locator=slack:C0547N89JUB:1780011492.577489:1780017512.134379` `source_timestamp=2026-05-29T01:18:32Z`
- There is no longer a developer relations person (like Tim) to manage external partnerships such as Goldsky. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_39629e0dfa197e6083d3eea91589ea17` `chunk_id=srcchunk_3650c91e2b134ff263fed7c75aa3f495` `native_locator=slack:C0547N89JUB:1780011492.577489:1780017586.775849` `source_timestamp=2026-05-29T01:19:57Z`
- User U0871SH0FNZ is likely the most knowledgeable about the indexing setup. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c5e774e7c95a39845c4039b506cdb9d9` `source_revision_id=srcrev_e0c66ad87fbddce1e81bf2dfd67aca3a` `chunk_id=srcchunk_f500295fdcc749817b6b8931be185efc` `native_locator=slack:C0547N89JUB:1780011492.577489:1780020855.024219` `source_timestamp=2026-05-29T02:14:15Z`

## Open Questions

- Can RSI safely discontinue Goldsky without impacting third‑party users?
- How does the BlockScout indexer integrate with the internal one?
- Which DEX projects are currently using Goldsky?
- Who is responsible for maintaining the internal Temporal indexer?

## Related Pages

- `goldsky`

## Sources

- `source_document_id`: `srcdoc_c5e774e7c95a39845c4039b506cdb9d9`
- `source_revision_id`: `srcrev_e0c66ad87fbddce1e81bf2dfd67aca3a`
