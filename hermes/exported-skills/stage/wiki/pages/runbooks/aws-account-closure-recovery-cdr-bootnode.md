---
title: "AWS Account Closure Recovery and CDR Bootnode Relocation"
type: "runbook"
slug: "runbooks/aws-account-closure-recovery-cdr-bootnode"
freshness: "2026-01-21T00:28:23Z"
tags:
  - "aws"
  - "azure"
  - "bootnode"
  - "cdr"
  - "incident-response"
  - "migration"
owners: []
source_revision_ids:
  - "srcrev_07a191138ceca2aea3e7a173a8a199e7"
  - "srcrev_1e33f666797775a86cd17ec2e662948b"
  - "srcrev_4625faf5add9f365f1abd30b13a4663b"
  - "srcrev_4c13d34f8a82c90092077549b23c8ab8"
  - "srcrev_5f0b8e5c43f79f57753f0838a16d9cee"
  - "srcrev_6792157d8f44a2c1416d2b571a2be033"
  - "srcrev_86706658b4a5efdec62ee819b2ab5a33"
  - "srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0"
  - "srcrev_a68b1287dadba8b820575f4d39379aa8"
  - "srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02"
conflict_state: "none"
---

# AWS Account Closure Recovery and CDR Bootnode Relocation

## Summary

Runbook for recovering a closed AWS account hosting a CDR bootnode, locating the bootnode in an alternative account, updating configuration with new IP, and planning migration to Azure.

## Claims

- CDR bootnode became inaccessible after a previous AWS account was accidentally closed/removed. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02` `chunk_id=srcchunk_baeadf16c43c8be70c8f8f4ecf26e3aa` `native_locator=slack:C0547N89JUB:1768898200.116689:1768898200.116689` `source_timestamp=2026-01-20T08:36:40Z`
- The closed account was identified as story-services-prod, later confirmed as the intended account for the bootnode. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5f0b8e5c43f79f57753f0838a16d9cee` `chunk_id=srcchunk_ea42c9c08496543d3d71a2c656780df8` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945728.509969` `source_timestamp=2026-01-20T21:48:48Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4c13d34f8a82c90092077549b23c8ab8` `chunk_id=srcchunk_48cf249c84184a95ab960b2e2e3905ce` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945746.724709` `source_timestamp=2026-01-20T21:49:06Z`
- Recovery of the closed AWS account required a 2FA SMS code sent to a specific user's phone number. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4625faf5add9f365f1abd30b13a4663b` `chunk_id=srcchunk_fc6b4b9a9a1dbf0b3bd8d9f41c996d55` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946049.112209` `source_timestamp=2026-01-20T21:54:09Z`
- After account restoration, no bootnode instances were found in the us-east-1 region of the story-services-prod account. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_6792157d8f44a2c1416d2b571a2be033` `chunk_id=srcchunk_0fad24a4ce91aea715fee53b2baf1dd2` `native_locator=slack:C0547N89JUB:1768898200.116689:1768951206.135269` `source_timestamp=2026-01-20T23:20:06Z`
- Additional AWS accounts were requested to be restored to locate the CDR bootnode. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_6792157d8f44a2c1416d2b571a2be033` `chunk_id=srcchunk_0fad24a4ce91aea715fee53b2baf1dd2` `native_locator=slack:C0547N89JUB:1768898200.116689:1768951206.135269` `source_timestamp=2026-01-20T23:20:06Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_1e33f666797775a86cd17ec2e662948b` `chunk_id=srcchunk_cc2cdb9011bae4171e103fa283b754d3` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955255.580949` `source_timestamp=2026-01-21T00:27:35Z`
- The CDR bootnode was eventually located in the story-services-staging AWS account. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0` `chunk_id=srcchunk_2ba98043fe8a681baf343529bb54d19b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955303.708259` `source_timestamp=2026-01-21T00:28:23Z`
- The bootnode's IP address changed to 54.153.69.163, requiring only a configuration update to point to the new IP. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a68b1287dadba8b820575f4d39379aa8` `chunk_id=srcchunk_1ba73dd4cad40a349168deb910d30a45` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955118.522269` `source_timestamp=2026-01-21T00:25:18Z`
- A plan was initiated to migrate the CDR bootnode to a company Azure subscription. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_07a191138ceca2aea3e7a173a8a199e7` `chunk_id=srcchunk_f7df99a90dcfddff8ff9bc872c83e7fc` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912588.205689` `source_timestamp=2026-01-20T12:36:28Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_86706658b4a5efdec62ee819b2ab5a33` `chunk_id=srcchunk_e3c1d4786874fce1043b88b377ef2ab6` `native_locator=slack:C0547N89JUB:1768898200.116689:1768949438.340279` `source_timestamp=2026-01-20T22:50:38Z`

## Open Questions

- Is SGX configuration required when migrating CDR bootnode to a new machine?

## Sources

- `source_document_id`: `srcdoc_4420812f05fe9ccafb6aef71709ac54c`
- `source_revision_id`: `srcrev_a68b1287dadba8b820575f4d39379aa8`
