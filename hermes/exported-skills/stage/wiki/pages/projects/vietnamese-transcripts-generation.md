---
title: "Vietnamese Transcripts Generation and Integration"
type: "project"
slug: "projects/vietnamese-transcripts-generation"
freshness: "2026-05-07T21:53:05Z"
tags:
  - "data-generation"
  - "poseidon"
  - "transcripts"
  - "vietnamese"
owners:
  - "U067QP5PD6J"
  - "U0772SH7BRA"
  - "U0A2D9U625V"
source_revision_ids:
  - "srcrev_058073e554263fcac16c1a9894d13f62"
  - "srcrev_124f6cf65612f788a9f3ed98a594a3b6"
  - "srcrev_32a13a1a74874fe04c16051199969e78"
  - "srcrev_3b563309766ba300cfe8e4275f3af423"
  - "srcrev_5f1467d9869ae1ed9c24dfdc39fe0785"
  - "srcrev_7f7b77637e96db0e124915aef12d5b0f"
  - "srcrev_813befc52a49c84446dd534063ec9a20"
  - "srcrev_8caa9d89d17571d48084c1f9d85ca6f5"
  - "srcrev_9ad82182fd9710421156f461cfecf204"
  - "srcrev_a702e1c26a390b4c9dbd3f2db6a77f44"
  - "srcrev_ac7ff4c38cb883eb8ce1d9b206deb417"
  - "srcrev_bdef89a7e928babe6f5bef3d5a0fffa9"
  - "srcrev_f38b1bccd512d4cb8283864f04ea90c8"
  - "srcrev_fd30163c05717fc696261ff96e9dd238"
conflict_state: "none"
---

# Vietnamese Transcripts Generation and Integration

## Summary

Effort to generate a large volume of diverse Vietnamese transcripts to support user submissions and platform training.

## Claims

- An additional 1000 Vietnamese transcripts were generated with slightly better diversity than before. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_9ad82182fd9710421156f461cfecf204` `chunk_id=srcchunk_b560b47a90f97b893cf0b528b36d165a` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778184868.941859` `source_timestamp=2026-05-07T20:14:28Z`
- The diversity metric discussed is for the newly generated 1000 transcripts only, not combined with previous ones. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_f38b1bccd512d4cb8283864f04ea90c8` `chunk_id=srcchunk_1fff63cf4b76dac762aab8ce790762db` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778185126.241359` `source_timestamp=2026-05-07T20:18:46Z`
- The goal is to evaluate diversity on the combined set of 2000 transcripts to ensure quality for the final 36k target. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_bdef89a7e928babe6f5bef3d5a0fffa9` `chunk_id=srcchunk_e92f6958219361946611a2d33a70ac35` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778185169.311289` `source_timestamp=2026-05-07T20:19:29Z`
- Diversity plots for the combined 2000 transcripts were shared. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_5f1467d9869ae1ed9c24dfdc39fe0785` `chunk_id=srcchunk_c0b836a3a67f7caf706a57143518d041` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778185927.411879` `source_timestamp=2026-05-07T20:32:07Z`
- User submissions for Vietnamese transcripts are at 6k, exceeding the available 1k transcripts, causing overlap and possible duplicate submissions. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_058073e554263fcac16c1a9894d13f62` `chunk_id=srcchunk_dbf8ccea5cae8ee99dbff02c46f25fc3` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778186883.198189` `source_timestamp=2026-05-07T20:48:08Z`
- U0A2D9U625V was asked to generate 40k transcripts to overshoot the 36k target, given the 6x submission-to-transcript ratio. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_fd30163c05717fc696261ff96e9dd238` `chunk_id=srcchunk_9a5487c7522f824265aea13e6775b973` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778186968.984849` `source_timestamp=2026-05-07T20:50:01Z`
- U0A2D9U625V started generating the 40k transcripts. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_a702e1c26a390b4c9dbd3f2db6a77f44` `chunk_id=srcchunk_1d2499dde0900d6f288db2e2ed85411c` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778187073.112349` `source_timestamp=2026-05-07T20:51:13Z`
- The format for transcript upload can be JSON, as previously used in the Cloudflare R2 bucket. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_124f6cf65612f788a9f3ed98a594a3b6` `chunk_id=srcchunk_fd12d5e12d2b90377b61bdc72673755b` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778187652.336289` `source_timestamp=2026-05-07T21:00:52Z`
- Transcripts were added manually, circumventing the non-working batch upload. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_32a13a1a74874fe04c16051199969e78` `chunk_id=srcchunk_f582b0f8101e4a9eb169e3fce5d3553b` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778189091.671419` `source_timestamp=2026-05-07T21:24:51Z`
- Duplicate submissions are expected and can be filtered later using usage counts per transcript. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_813befc52a49c84446dd534063ec9a20` `chunk_id=srcchunk_28a946004d10aa41b4a2dc9c7dbae253` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778189464.576829` `source_timestamp=2026-05-07T21:31:04Z`
- There is a lesson learned to vet transcripts beforehand and be ready at rollout to avoid quality issues. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_3b563309766ba300cfe8e4275f3af423` `chunk_id=srcchunk_c99c58ebd1c6f49f3bfa0e1fbd891b1c` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778189506.694319` `source_timestamp=2026-05-07T21:31:46Z`
- The admin batch upload feature is currently broken and needs fixing, including field validation. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_ac7ff4c38cb883eb8ce1d9b206deb417` `chunk_id=srcchunk_9079f8b92f88014b81fdda747f9d9519` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778189528.456329` `source_timestamp=2026-05-07T21:32:08Z`
- Suggestion to integrate transcript generation as a self-serve pipeline on the Poseidon web app. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_8caa9d89d17571d48084c1f9d85ca6f5` `chunk_id=srcchunk_4273cb5bdf64ee9fa9953c242369c541` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778189566.919979` `source_timestamp=2026-05-07T21:32:56Z`
- Agreement that transcript generation should be self-serve through the Poseidon console, leveraging existing pipeline, with integration needed. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_12dac6726c7aee2db5802127de71cbed` `source_revision_id=srcrev_7f7b77637e96db0e124915aef12d5b0f` `chunk_id=srcchunk_c623b90eaba4a39d2fddee112e07dad1` `native_locator=slack:C0AL7EKNHDF:1778184195.122739:1778190785.386249` `source_timestamp=2026-05-07T21:53:05Z`

## Open Questions

- How will duplicate submissions be filtered and what is the threshold?
- Is the batch upload fix prioritized and assigned to U067QP5PD6J?
- When will the self-serve transcript generation pipeline be integrated into the Poseidon console?

## Sources

- `source_document_id`: `srcdoc_12dac6726c7aee2db5802127de71cbed`
- `source_revision_id`: `srcrev_7f7b77637e96db0e124915aef12d5b0f`
