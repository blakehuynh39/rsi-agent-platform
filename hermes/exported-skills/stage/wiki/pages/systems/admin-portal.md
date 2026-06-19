---
title: "Admin Portal"
type: "system"
slug: "systems/admin-portal"
freshness: "2026-04-30T23:37:22Z"
tags:
  - "access-control"
  - "admin"
  - "monorepo"
  - "numo"
owners: []
source_revision_ids:
  - "srcrev_130a1b71de3dd181ac701e9e595738b0"
  - "srcrev_82a4276e31b5f0d625d808dc8f822296"
  - "srcrev_c6cc71563e538364ffbac84e7f069104"
  - "srcrev_cbc387150e4f13b2820c4f00a7ab443a"
  - "srcrev_f3c99e68d9f8ed5055f0262534661836"
conflict_state: "none"
---

# Admin Portal

## Summary

The admin portal is a gated web application within the Numo monorepo used for viewing internal data, with access controlled by an email allowlist implemented in the depin-backend.

## Claims

- The admin portal code resides under /apps/admin in the Numo monorepo. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e03f038f91c599ffbc4e838ea40403cc` `source_revision_id=srcrev_cbc387150e4f13b2820c4f00a7ab443a` `chunk_id=srcchunk_045e8faf0513672968f63b0f4df0a1d1` `native_locator=slack:C0AL7EKNHDF:1777578999.960999:1777580298.014139` `source_timestamp=2026-04-30T20:18:18Z`
- The admin portal is part of the Numo monorepo, similar to the web application. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e03f038f91c599ffbc4e838ea40403cc` `source_revision_id=srcrev_c6cc71563e538364ffbac84e7f069104` `chunk_id=srcchunk_65b4883337b0a9a4f9de69eb3766db50` `native_locator=slack:C0AL7EKNHDF:1777578999.960999:1777580287.095069` `source_timestamp=2026-04-30T20:18:07Z`
- Access to the admin portal is controlled by an email allowlist. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e03f038f91c599ffbc4e838ea40403cc` `source_revision_id=srcrev_f3c99e68d9f8ed5055f0262534661836` `chunk_id=srcchunk_bb91f5883309a678804c90eaf4e1fc9d` `native_locator=slack:C0AL7EKNHDF:1777578999.960999:1777581059.906979` `source_timestamp=2026-04-30T20:30:59Z`
- The allowlist logic is implemented in the depin-backend repository at apps/api/src/http/extractors.rs (staging branch). `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e03f038f91c599ffbc4e838ea40403cc` `source_revision_id=srcrev_130a1b71de3dd181ac701e9e595738b0` `chunk_id=srcchunk_8ef575f3b0c97c250ebeba9dd8096647` `native_locator=slack:C0AL7EKNHDF:1777578999.960999:1777591914.946719` `source_timestamp=2026-04-30T23:31:54Z`
- As of the conversation, Vinod Tiwari's email (vinod.tiwari@piplabs.xyz) was already in the allowlist, granting access. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e03f038f91c599ffbc4e838ea40403cc` `source_revision_id=srcrev_82a4276e31b5f0d625d808dc8f822296` `chunk_id=srcchunk_c0646296895741386c9209cdab0fb81e` `native_locator=slack:C0AL7EKNHDF:1777578999.960999:1777592242.697579` `source_timestamp=2026-04-30T23:37:22Z`

## Open Questions

- How are new emails added to the allowlist?
- Who is responsible for maintaining the admin portal allowlist?

## Sources

- `source_document_id`: `srcdoc_e03f038f91c599ffbc4e838ea40403cc`
- `source_revision_id`: `srcrev_82a4276e31b5f0d625d808dc8f822296`
