---
title: "Story Kernel Intel TDX Testing"
type: "project"
slug: "projects/story-kernel-tdx-testing"
freshness: "2026-05-08T06:27:41Z"
tags:
  - "azure"
  - "intel-tdx"
  - "story-kernel"
owners:
  - "Story Kernel Team"
source_revision_ids:
  - "srcrev_1385b97f488db402fea18a9d550d0d21"
  - "srcrev_2c5187283a66fc86fe41ddbf5abdf3cb"
  - "srcrev_60f85788641179d68c65519f5f08a2ed"
  - "srcrev_682754212573235c17474a6c7da7143f"
  - "srcrev_b52c078e649c051edcc9150897bebed3"
  - "srcrev_bb9dab66f53bfc53e537c5e65b570685"
  - "srcrev_c5f624dbf33db00da0ac9e73365f3dcd"
  - "srcrev_d7dfd3e0b5925f036b7ac85ef0027cb5"
conflict_state: "none"
---

# Story Kernel Intel TDX Testing

## Summary

Testing Intel TDX support for story-kernel using Azure VMs; not part of the 05/13 upgrade.

## Claims

- Intel TDX support is being added to story-kernel. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_d7dfd3e0b5925f036b7ac85ef0027cb5` `chunk_id=srcchunk_1a36056464016932a6c3ac7c4896ba03` `native_locator=slack:C0547N89JUB:1778215885.899199:1778215885.899199` `source_timestamp=2026-05-08T04:51:25Z`
- The test environment includes one SGX VM of type Standard_DC4s_v3. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_d7dfd3e0b5925f036b7ac85ef0027cb5` `chunk_id=srcchunk_1a36056464016932a6c3ac7c4896ba03` `native_locator=slack:C0547N89JUB:1778215885.899199:1778215885.899199` `source_timestamp=2026-05-08T04:51:25Z`
- The TDX VM type was changed from Standard_DC4es_v5 to Standard_DC4es_v6 (Intel Xeon 5) due to quicker availability. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_1385b97f488db402fea18a9d550d0d21` `chunk_id=srcchunk_af953032f32c34f3f16284f43623d087` `native_locator=slack:C0547N89JUB:1778215885.899199:1778221013.548679` `source_timestamp=2026-05-08T06:16:53Z`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_b52c078e649c051edcc9150897bebed3` `chunk_id=srcchunk_b5927e1ea6a7420d779a49cd43708e5d` `native_locator=slack:C0547N89JUB:1778215885.899199:1778221661.390719` `source_timestamp=2026-05-08T06:27:41Z`
- This TDX testing is not part of the upcoming 05/13 upgrade and will be applied later. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_2c5187283a66fc86fe41ddbf5abdf3cb` `chunk_id=srcchunk_668df3d4a82d76de0b4c03aaa64740f0` `native_locator=slack:C0547N89JUB:1778215885.899199:1778217133.717319` `source_timestamp=2026-05-08T05:12:20Z`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_682754212573235c17474a6c7da7143f` `chunk_id=srcchunk_15dc36073bbde354a3ceb24c5c665aa6` `native_locator=slack:C0547N89JUB:1778215885.899199:1778217188.650819` `source_timestamp=2026-05-08T05:13:08Z`
- The VMs will be used to set up a private devnet for testing, not connected to Aeneid. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_60f85788641179d68c65519f5f08a2ed` `chunk_id=srcchunk_f37bb7fbc6c159e3e68595305cb51305` `native_locator=slack:C0547N89JUB:1778215885.899199:1778216283.157599` `source_timestamp=2026-05-08T04:58:03Z`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_c5f624dbf33db00da0ac9e73365f3dcd` `chunk_id=srcchunk_51b25a47d63035c75562d112c9251abe` `native_locator=slack:C0547N89JUB:1778215885.899199:1778216305.452559` `source_timestamp=2026-05-08T04:58:25Z`
  - citation: `source_document_id=srcdoc_e86514d56923931bec8d24bb7974c129` `source_revision_id=srcrev_bb9dab66f53bfc53e537c5e65b570685` `chunk_id=srcchunk_b47ea96c94d42ca3ce75fa210ca9e57e` `native_locator=slack:C0547N89JUB:1778215885.899199:1778216318.298999` `source_timestamp=2026-05-08T04:58:38Z`

## Related Pages

- `aeneid`
- `story-kernel`

## Sources

- `source_document_id`: `srcdoc_e86514d56923931bec8d24bb7974c129`
- `source_revision_id`: `srcrev_b52c078e649c051edcc9150897bebed3`
