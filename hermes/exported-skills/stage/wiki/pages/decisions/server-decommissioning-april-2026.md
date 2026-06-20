---
title: "Server Decommissioning Decision (April 2026)"
type: "decision"
slug: "decisions/server-decommissioning-april-2026"
freshness: "2026-04-09T07:24:02Z"
tags:
  - "decommissioning"
  - "infrastructure"
  - "story-protocol"
  - "testing-environments"
owners: []
source_revision_ids:
  - "srcrev_0a21fb0144dae9d5199944e6e4eca76d"
  - "srcrev_2eed1c13ec6936c7fdd891ea7b608bbe"
  - "srcrev_338db1fed06c7a4a0fd204fbe9c74e50"
  - "srcrev_642b1899ea9106e33314fe2890b6198e"
  - "srcrev_68fbe8f4dca709538dafe11965dc4811"
  - "srcrev_74e6eef13bb9964bde86cc491ae5e6a9"
  - "srcrev_7a5535b59eba22f7d0b5cb190c91404c"
  - "srcrev_f266be649df96aa8c61c0c230e01718d"
conflict_state: "none"
---

# Server Decommissioning Decision (April 2026)

## Summary

In April 2026, the team decided to decommission most development/test server groups, retaining only QA (JPE-QA-RG) and Yingyang (JPE-YINGYANG-RG) environments for ongoing testing. All other groups (CDR, Hans, Jdub, Steven, Raul, Aeneid) were slated for removal. QA provides a shared testing environment, though concerns about interference among testers were noted. Yingyang is necessary for testing `story-kernel` changes.

## Claims

- All existing servers across 8 resource groups (CDR, Hans, Jdub, Steven, Raul, QA, Aeneid, Yingyang) were listed, with VM names and IPs. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5` `source_revision_id=srcrev_f266be649df96aa8c61c0c230e01718d` `chunk_id=srcchunk_c0c5d9b1f0269132b8b418c0a6b2046b` `native_locator=slack:C0547N89JUB:1775699256.226589:1775699256.226589` `source_timestamp=2026-04-09T01:47:36Z`
- QA (JPE-QA-RG) environment was confirmed to be kept. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5` `source_revision_id=srcrev_642b1899ea9106e33314fe2890b6198e` `chunk_id=srcchunk_03abbb9809fbc7217ea7a4b718d707d7` `native_locator=slack:C0547N89JUB:1775699256.226589:1775699417.649779` `source_timestamp=2026-04-09T01:50:17Z`
- The Yingyang (JPE-YINGYANG-RG) validator is the only environment for testing story-kernel changes. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5` `source_revision_id=srcrev_0a21fb0144dae9d5199944e6e4eca76d` `chunk_id=srcchunk_5e3f6a394d9de8950e1e44fa34c58b60` `native_locator=slack:C0547N89JUB:1775699256.226589:1775699534.142149` `source_timestamp=2026-04-09T01:52:14Z`
- CDR (JPE-CDR-RG) was previously used for local testing but has been superseded by QA for automated tests. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5` `source_revision_id=srcrev_74e6eef13bb9964bde86cc491ae5e6a9` `chunk_id=srcchunk_1dfafce90201a68d3b977eacc8392edf` `native_locator=slack:C0547N89JUB:1775699256.226589:1775703846.843409` `source_timestamp=2026-04-09T03:04:58Z`
- Multiple testers sharing the QA environment interfered with each other's testing, reducing efficiency. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5` `source_revision_id=srcrev_2eed1c13ec6936c7fdd891ea7b608bbe` `chunk_id=srcchunk_717cda2569f1f9df625a2a9f1ad52b53` `native_locator=slack:C0547N89JUB:1775699256.226589:1775704633.854089` `source_timestamp=2026-04-09T03:17:13Z`
- The final decision was to keep QA (JPE-QA-RG) and Yingyang (JPE-YINGYANG-RG) environments, and to decommission all other environments: CDR, Hans, Jdub, Steven, Raul, and Aeneid. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5` `source_revision_id=srcrev_68fbe8f4dca709538dafe11965dc4811` `chunk_id=srcchunk_a786d43d508a10ce614fa94dfb6d336c` `native_locator=slack:C0547N89JUB:1775699256.226589:1775705521.811509` `source_timestamp=2026-04-09T03:32:01Z`
  - citation: `source_document_id=srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5` `source_revision_id=srcrev_7a5535b59eba22f7d0b5cb190c91404c` `chunk_id=srcchunk_3a84855754841067d0a736e28e24818d` `native_locator=slack:C0547N89JUB:1775699256.226589:1775705569.674859` `source_timestamp=2026-04-09T03:33:00Z`
  - citation: `source_document_id=srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5` `source_revision_id=srcrev_338db1fed06c7a4a0fd204fbe9c74e50` `chunk_id=srcchunk_e1115290f98219391c353b05fdf20d0a` `native_locator=slack:C0547N89JUB:1775699256.226589:1775719442.991559` `source_timestamp=2026-04-09T07:24:02Z`

## Open Questions

- How to manage QA environment to prevent interference between multiple testers?

## Sources

- `source_document_id`: `srcdoc_531c1ca7c8eeeae4d91b54b62f0ec0a5`
- `source_revision_id`: `srcrev_338db1fed06c7a4a0fd204fbe9c74e50`
