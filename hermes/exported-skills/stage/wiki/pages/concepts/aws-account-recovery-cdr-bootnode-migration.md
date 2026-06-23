---
title: "AWS Account Recovery and CDR Bootnode Migration to Azure"
type: "concept"
slug: "concepts/aws-account-recovery-cdr-bootnode-migration"
freshness: "2026-01-21T00:28:23Z"
tags:
  - "aws"
  - "azure"
  - "bootnode"
  - "cdr"
  - "postmortem"
  - "recovery"
owners: []
source_revision_ids:
  - "srcrev_07a191138ceca2aea3e7a173a8a199e7"
  - "srcrev_10d5913d78928b9a1bfc247e1b5674c1"
  - "srcrev_4c13d34f8a82c90092077549b23c8ab8"
  - "srcrev_5abeeff78e095b18bc26e8748e103f29"
  - "srcrev_5f0b8e5c43f79f57753f0838a16d9cee"
  - "srcrev_86706658b4a5efdec62ee819b2ab5a33"
  - "srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0"
  - "srcrev_a68b1287dadba8b820575f4d39379aa8"
  - "srcrev_cfd931a15d500789656c173e0645efe6"
  - "srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02"
  - "srcrev_fb369332858d64661f0921a2e5970b01"
conflict_state: "none"
---

# AWS Account Recovery and CDR Bootnode Migration to Azure

## Summary

In January 2026, a closed AWS account caused loss of access to the CDR bootnode. The account was restored, the bootnode was located in story-services-staging, and plans were made to migrate the bootnode to a company Azure account.

## Claims

- The previous AWS account was removed, causing inability to access the CDR bootnode. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02` `chunk_id=srcchunk_baeadf16c43c8be70c8f8f4ecf26e3aa` `native_locator=slack:C0547N89JUB:1768898200.116689:1768898200.116689` `source_timestamp=2026-01-20T08:36:40Z`
- The missing bootnode was thought to be in the story-prod account. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_cfd931a15d500789656c173e0645efe6` `chunk_id=srcchunk_de3e4a496737c915025ea80f03a2c38f` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912531.110089` `source_timestamp=2026-01-20T12:35:31Z`
- It was clarified that the intended account was story-services-prod. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5f0b8e5c43f79f57753f0838a16d9cee` `chunk_id=srcchunk_ea42c9c08496543d3d71a2c656780df8` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945728.509969` `source_timestamp=2026-01-20T21:48:48Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4c13d34f8a82c90092077549b23c8ab8` `chunk_id=srcchunk_48cf249c84184a95ab960b2e2e3905ce` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945746.724709` `source_timestamp=2026-01-20T21:49:06Z`
- The AWS account was restored. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_10d5913d78928b9a1bfc247e1b5674c1` `chunk_id=srcchunk_5b326265221430247237293aec04819b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768948086.089599` `source_timestamp=2026-01-20T22:28:06Z`
- The bootnode was found in the story-services-staging account. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0` `chunk_id=srcchunk_2ba98043fe8a681baf343529bb54d19b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955303.708259` `source_timestamp=2026-01-21T00:28:23Z`
- The bootnode IP changed to 54.153.69.163. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a68b1287dadba8b820575f4d39379aa8` `chunk_id=srcchunk_1ba73dd4cad40a349168deb910d30a45` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955118.522269` `source_timestamp=2026-01-21T00:25:18Z`
- Migration of the CDR bootnode to a company Azure account is planned. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_07a191138ceca2aea3e7a173a8a199e7` `chunk_id=srcchunk_f7df99a90dcfddff8ff9bc872c83e7fc` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912588.205689` `source_timestamp=2026-01-20T12:36:28Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5abeeff78e095b18bc26e8748e103f29` `chunk_id=srcchunk_1be807cd0340f11193388886dd219973` `native_locator=slack:C0547N89JUB:1768898200.116689:1768927250.985599` `source_timestamp=2026-01-20T16:40:50Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_86706658b4a5efdec62ee819b2ab5a33` `chunk_id=srcchunk_e3c1d4786874fce1043b88b377ef2ab6` `native_locator=slack:C0547N89JUB:1768898200.116689:1768949438.340279` `source_timestamp=2026-01-20T22:50:38Z`
- SGX related configuration may need to be re-setup on the new machine. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_fb369332858d64661f0921a2e5970b01` `chunk_id=srcchunk_fc450f672b5723723b3d8eff07a935b1` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955270.510379` `source_timestamp=2026-01-21T00:27:50Z`

## Open Questions

- Was the Azure subscription successfully created?
- Was the bootnode migrated to the Azure account?
- Was the SGX configuration re-established on the Azure machine?
- What is the current status of the CDR bootnode?

## Sources

- `source_document_id`: `srcdoc_4420812f05fe9ccafb6aef71709ac54c`
- `source_revision_id`: `srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0`
