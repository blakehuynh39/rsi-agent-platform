---
title: "GCP to AWS VM Migration"
type: "project"
slug: "projects/gcp-to-aws-vm-migration"
freshness: "2026-04-19T23:57:31Z"
tags:
  - "aws"
  - "gcp"
  - "migration"
  - "vms"
owners:
  - "slack:U04L0DD7UM9"
source_revision_ids:
  - "srcrev_024fd38cc417640399a3feddfc0a0976"
  - "srcrev_20609b29a08db84e934cf4461c925f6d"
  - "srcrev_5910b6a4643aeddf784508814e88d94d"
  - "srcrev_cbc13255241d9f95544bbf6b2c80da78"
  - "srcrev_e20c2b37349978e92cf8eb762c16942e"
conflict_state: "none"
---

# GCP to AWS VM Migration

## Summary

Migration of VMs from GCP to AWS, covering instance mappings, verification steps, and subsequent adjustments to termination plans. Some VMs were retained for testing after initial deprecation schedule.

## Claims

- GCP to AWS VM migration has been completed. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_024fd38cc417640399a3feddfc0a0976` `chunk_id=srcchunk_0f86c42bd2bdc2341d03e92bbd0df274` `native_locator=slack:C0547N89JUB:1776232959.799959:1776232959.799959` `source_timestamp=2026-04-15T06:02:39Z`
- Kingter-vm was migrated to AWS with public IP 34.195.7.171. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_024fd38cc417640399a3feddfc0a0976` `chunk_id=srcchunk_0f86c42bd2bdc2341d03e92bbd0df274` `native_locator=slack:C0547N89JUB:1776232959.799959:1776232959.799959` `source_timestamp=2026-04-15T06:02:39Z`
- Kingter-vm-backend was migrated to AWS with public IP 52.71.75.39. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_024fd38cc417640399a3feddfc0a0976` `chunk_id=srcchunk_0f86c42bd2bdc2341d03e92bbd0df274` `native_locator=slack:C0547N89JUB:1776232959.799959:1776232959.799959` `source_timestamp=2026-04-15T06:02:39Z`
- Jdub-archive was migrated to AWS with public IP 54.225.196.246. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_024fd38cc417640399a3feddfc0a0976` `chunk_id=srcchunk_0f86c42bd2bdc2341d03e92bbd0df274` `native_locator=slack:C0547N89JUB:1776232959.799959:1776232959.799959` `source_timestamp=2026-04-15T06:02:39Z`
- Steven-vm was migrated to AWS with public IP 3.226.240.21. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_024fd38cc417640399a3feddfc0a0976` `chunk_id=srcchunk_0f86c42bd2bdc2341d03e92bbd0df274` `native_locator=slack:C0547N89JUB:1776232959.799959:1776232959.799959` `source_timestamp=2026-04-15T06:02:39Z`
- Steven-vm was confirmed working by its owner. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_5910b6a4643aeddf784508814e88d94d` `chunk_id=srcchunk_a3a00f92f3daf2cc8bc32ee819dc737e` `native_locator=slack:C0547N89JUB:1776232959.799959:1776233205.591349` `source_timestamp=2026-04-15T06:06:45Z`
- Termination of the original GCP servers was planned within a week of the migration notice. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_cbc13255241d9f95544bbf6b2c80da78` `chunk_id=srcchunk_46451449435f24b58b1d18353fd34111` `native_locator=slack:C0547N89JUB:1776232959.799959:1776301276.078889` `source_timestamp=2026-04-16T01:01:16Z`
- A follow-up confirmation request was sent before proceeding with termination. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_e20c2b37349978e92cf8eb762c16942e` `chunk_id=srcchunk_e2182260c6aa91afcc7cf971e4bffb99` `native_locator=slack:C0547N89JUB:1776232959.799959:1776389304.360949` `source_timestamp=2026-04-17T01:28:24Z`
- The remaining servers belong to user slack:U0AKJF57GG2. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_20609b29a08db84e934cf4461c925f6d` `chunk_id=srcchunk_bdd002b6468d367523e4fb2dc2787d11` `native_locator=slack:C0547N89JUB:1776232959.799959:1776642524.362899` `source_timestamp=2026-04-19T23:57:31Z`
- After discussion with Kingter, the decision was made to keep those servers for testing purposes rather than proceed with immediate termination. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_57aca3204b70d49c766c68b1377718ac` `source_revision_id=srcrev_20609b29a08db84e934cf4461c925f6d` `chunk_id=srcchunk_bdd002b6468d367523e4fb2dc2787d11` `native_locator=slack:C0547N89JUB:1776232959.799959:1776642524.362899` `source_timestamp=2026-04-19T23:57:31Z`

## Open Questions

- What is the Posidon side AWS and why was migration to it originally proposed?
- What is the timeline for full decommission of the remaining GCP servers?

## Related Pages

- `jdub-archive`
- `kingter-vm`
- `steven-vm`

## Sources

- `source_document_id`: `srcdoc_57aca3204b70d49c766c68b1377718ac`
- `source_revision_id`: `srcrev_20609b29a08db84e934cf4461c925f6d`
