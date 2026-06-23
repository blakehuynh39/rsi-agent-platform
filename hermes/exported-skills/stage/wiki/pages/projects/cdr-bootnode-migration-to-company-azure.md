---
title: "CDR Bootnode Migration to Company Azure"
type: "project"
slug: "projects/cdr-bootnode-migration-to-company-azure"
freshness: "2026-01-21T00:28:23Z"
tags:
  - "aws"
  - "azure"
  - "bootnode"
  - "cdr"
  - "infrastructure"
owners: []
source_revision_ids:
  - "srcrev_07a191138ceca2aea3e7a173a8a199e7"
  - "srcrev_2896b27d572c72e037fe7a85b367e198"
  - "srcrev_4625faf5add9f365f1abd30b13a4663b"
  - "srcrev_4c13d34f8a82c90092077549b23c8ab8"
  - "srcrev_5abeeff78e095b18bc26e8748e103f29"
  - "srcrev_5f0b8e5c43f79f57753f0838a16d9cee"
  - "srcrev_6792157d8f44a2c1416d2b571a2be033"
  - "srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0"
  - "srcrev_a68b1287dadba8b820575f4d39379aa8"
  - "srcrev_cfd931a15d500789656c173e0645efe6"
  - "srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02"
  - "srcrev_fb369332858d64661f0921a2e5970b01"
conflict_state: "none"
---

# CDR Bootnode Migration to Company Azure

## Summary

Migrate the CDR bootnode from an AWS staging account to a company Azure subscription. The project involves recovering a closed AWS account where the bootnode was originally created, identifying its location, and reconfiguring the bootnode IP and potentially TEE setup on the new Azure environment.

## Claims

- The CDR bootnode was initially inaccessible because the previous AWS account was removed. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02` `chunk_id=srcchunk_baeadf16c43c8be70c8f8f4ecf26e3aa` `native_locator=slack:C0547N89JUB:1768898200.116689:1768898200.116689` `source_timestamp=2026-01-20T08:36:40Z`
- The bootnode was thought to be in the story-prod AWS account, but later identified as story-services-prod. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_cfd931a15d500789656c173e0645efe6` `chunk_id=srcchunk_de3e4a496737c915025ea80f03a2c38f` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912531.110089` `source_timestamp=2026-01-20T12:35:31Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5f0b8e5c43f79f57753f0838a16d9cee` `chunk_id=srcchunk_ea42c9c08496543d3d71a2c656780df8` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945728.509969` `source_timestamp=2026-01-20T21:48:48Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4c13d34f8a82c90092077549b23c8ab8` `chunk_id=srcchunk_48cf249c84184a95ab960b2e2e3905ce` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945746.724709` `source_timestamp=2026-01-20T21:49:06Z`
- The closed AWS account was restored, requiring SMS 2FA code from @U08D32C1EF3. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_2896b27d572c72e037fe7a85b367e198` `chunk_id=srcchunk_61b899213fccc7ca5ff0fec6e295ec90` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945789.712699` `source_timestamp=2026-01-20T21:49:49Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4625faf5add9f365f1abd30b13a4663b` `chunk_id=srcchunk_fc6b4b9a9a1dbf0b3bd8d9f41c996d55` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946049.112209` `source_timestamp=2026-01-20T21:54:09Z`
- After restoration, no bootnode instances were found in us-east-1 of story-services-prod, leading to search in other accounts. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_6792157d8f44a2c1416d2b571a2be033` `chunk_id=srcchunk_0fad24a4ce91aea715fee53b2baf1dd2` `native_locator=slack:C0547N89JUB:1768898200.116689:1768951206.135269` `source_timestamp=2026-01-20T23:20:06Z`
- The CDR bootnode was located in the story-services-staging AWS account. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0` `chunk_id=srcchunk_2ba98043fe8a681baf343529bb54d19b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955303.708259` `source_timestamp=2026-01-21T00:28:23Z`
- After locating the bootnode, its IP address was changed to 54.153.69.163. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a68b1287dadba8b820575f4d39379aa8` `chunk_id=srcchunk_1ba73dd4cad40a349168deb910d30a45` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955118.522269` `source_timestamp=2026-01-21T00:25:18Z`
- Migration to a company Azure subscription is planned, and an Azure subscription will be created with the company card. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_07a191138ceca2aea3e7a173a8a199e7` `chunk_id=srcchunk_f7df99a90dcfddff8ff9bc872c83e7fc` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912588.205689` `source_timestamp=2026-01-20T12:36:28Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5abeeff78e095b18bc26e8748e103f29` `chunk_id=srcchunk_1be807cd0340f11193388886dd219973` `native_locator=slack:C0547N89JUB:1768898200.116689:1768927250.985599` `source_timestamp=2026-01-20T16:40:50Z`
- The bootnode configuration needs updating on the client side, and there is a question about setting up SGX/TEE on the new machine. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_fb369332858d64661f0921a2e5970b01` `chunk_id=srcchunk_fc450f672b5723723b3d8eff07a935b1` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955270.510379` `source_timestamp=2026-01-21T00:27:50Z`

## Open Questions

- Does the SGX/TEE configuration need to be re-setup on the new Azure machine?

## Sources

- `source_document_id`: `srcdoc_4420812f05fe9ccafb6aef71709ac54c`
- `source_revision_id`: `srcrev_1e33f666797775a86cd17ec2e662948b`
