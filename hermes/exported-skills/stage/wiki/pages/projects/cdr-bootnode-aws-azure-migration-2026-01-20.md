---
title: "CDR Bootnode AWS to Azure Migration (2026-01-20)"
type: "project"
slug: "projects/cdr-bootnode-aws-azure-migration-2026-01-20"
freshness: "2026-01-21T00:28:23Z"
tags:
  - "AWS"
  - "Azure"
  - "bootnode"
  - "CDR"
  - "incident"
  - "migration"
owners: []
source_revision_ids:
  - "srcrev_07a191138ceca2aea3e7a173a8a199e7"
  - "srcrev_10d5913d78928b9a1bfc247e1b5674c1"
  - "srcrev_19286702bdbf2dd7ec13662fecad6f73"
  - "srcrev_1e33f666797775a86cd17ec2e662948b"
  - "srcrev_2896b27d572c72e037fe7a85b367e198"
  - "srcrev_4625faf5add9f365f1abd30b13a4663b"
  - "srcrev_4c13d34f8a82c90092077549b23c8ab8"
  - "srcrev_5abeeff78e095b18bc26e8748e103f29"
  - "srcrev_5f0b8e5c43f79f57753f0838a16d9cee"
  - "srcrev_608723a566d204b5f7f34a798ee733e1"
  - "srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0"
  - "srcrev_912b3eb6cb0432731042200191c7cf00"
  - "srcrev_9f13ba58d1623f4d68a4a76eb9537cd7"
  - "srcrev_a68b1287dadba8b820575f4d39379aa8"
  - "srcrev_cfd931a15d500789656c173e0645efe6"
  - "srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02"
  - "srcrev_fb369332858d64661f0921a2e5970b01"
  - "srcrev_fe142cd81e3357fb163f0c67831e8f4c"
conflict_state: "none"
---

# CDR Bootnode AWS to Azure Migration (2026-01-20)

## Summary

The CDR bootnode in an AWS account (story-services-staging) was lost after the account was prematurely closed. The account was restored, the bootnode IP changed to 54.153.69.163, and a decision was made to migrate the CDR devnet to a company Azure account.

## Claims

- U079ZJ48D62 reported inability to access the CDR bootnode, suspecting that U08332YRB7W had removed the previous AWS account. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02` `chunk_id=srcchunk_baeadf16c43c8be70c8f8f4ecf26e3aa` `native_locator=slack:C0547N89JUB:1768898200.116689:1768898200.116689` `source_timestamp=2026-01-20T08:36:40Z`
- U08332YRB7W believed the bootnode would be in a new account, not the old one. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_608723a566d204b5f7f34a798ee733e1` `chunk_id=srcchunk_c3cf4418965397ba7e6116c902141a56` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912467.817399` `source_timestamp=2026-01-20T12:34:27Z`
- The lost bootnode was initially thought to be in the story-prod account, then story-services-prod, and finally confirmed as story-services-staging. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_cfd931a15d500789656c173e0645efe6` `chunk_id=srcchunk_de3e4a496737c915025ea80f03a2c38f` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912531.110089` `source_timestamp=2026-01-20T12:35:31Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5f0b8e5c43f79f57753f0838a16d9cee` `chunk_id=srcchunk_ea42c9c08496543d3d71a2c656780df8` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945728.509969` `source_timestamp=2026-01-20T21:48:48Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4c13d34f8a82c90092077549b23c8ab8` `chunk_id=srcchunk_48cf249c84184a95ab960b2e2e3905ce` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945746.724709` `source_timestamp=2026-01-20T21:49:06Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0` `chunk_id=srcchunk_2ba98043fe8a681baf343529bb54d19b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955303.708259` `source_timestamp=2026-01-21T00:28:23Z`
- U08332YRB7W attempted to migrate to a new Azure subscription and asked for a company card to create an Azure subscription. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_07a191138ceca2aea3e7a173a8a199e7` `chunk_id=srcchunk_f7df99a90dcfddff8ff9bc872c83e7fc` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912588.205689` `source_timestamp=2026-01-20T12:36:28Z`
- U079ZJ48D62 agreed to create the Azure subscription. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5abeeff78e095b18bc26e8748e103f29` `chunk_id=srcchunk_1be807cd0340f11193388886dd219973` `native_locator=slack:C0547N89JUB:1768898200.116689:1768927250.985599` `source_timestamp=2026-01-20T16:40:50Z`
- U079ZJ48D62 requested recovery of the server in the closed account. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_912b3eb6cb0432731042200191c7cf00` `chunk_id=srcchunk_508dc8620722e7309aab5fed0c172e36` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945660.248789` `source_timestamp=2026-01-20T21:47:40Z`
- U08332YRB7W began working on recovering the AWS account. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_2896b27d572c72e037fe7a85b367e198` `chunk_id=srcchunk_61b899213fccc7ca5ff0fec6e295ec90` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945789.712699` `source_timestamp=2026-01-20T21:49:49Z`
- To recover the closed AWS account, U08332YRB7W needed an SMS 2FA code sent to U08D32C1EF3's phone. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4625faf5add9f365f1abd30b13a4663b` `chunk_id=srcchunk_fc6b4b9a9a1dbf0b3bd8d9f41c996d55` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946049.112209` `source_timestamp=2026-01-20T21:54:09Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_fe142cd81e3357fb163f0c67831e8f4c` `chunk_id=srcchunk_fd3ddfabaa94d37d17ea57d69623f9f0` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946065.730079` `source_timestamp=2026-01-20T21:54:25Z`
- The 2FA code initially expired and had to be resent. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_9f13ba58d1623f4d68a4a76eb9537cd7` `chunk_id=srcchunk_0796ccf194d82673e26a05cdbcd58568` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946747.697219` `source_timestamp=2026-01-20T22:05:47Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_19286702bdbf2dd7ec13662fecad6f73` `chunk_id=srcchunk_4f89ae16e446a451d0165c390b724e20` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946750.552909` `source_timestamp=2026-01-20T22:05:50Z`
- The AWS account was successfully restored. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_10d5913d78928b9a1bfc247e1b5674c1` `chunk_id=srcchunk_5b326265221430247237293aec04819b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768948086.089599` `source_timestamp=2026-01-20T22:28:06Z`
- After restoration, the bootnode IP address changed to 54.153.69.163. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a68b1287dadba8b820575f4d39379aa8` `chunk_id=srcchunk_1ba73dd4cad40a349168deb910d30a45` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955118.522269` `source_timestamp=2026-01-21T00:25:18Z`
- U079ZJ48D62 accepted that the configuration only needed bootnode IP change and requested to migrate CDR devnet to company Azure. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a68b1287dadba8b820575f4d39379aa8` `chunk_id=srcchunk_1ba73dd4cad40a349168deb910d30a45` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955118.522269` `source_timestamp=2026-01-21T00:25:18Z`
- All AWS accounts were restored. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_1e33f666797775a86cd17ec2e662948b` `chunk_id=srcchunk_cc2cdb9011bae4171e103fa283b754d3` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955255.580949` `source_timestamp=2026-01-21T00:27:35Z`
- U079ZJ48D62 expressed concern about needing to re-setup SGX (TEE) configuration on the new Azure machine. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_fb369332858d64661f0921a2e5970b01` `chunk_id=srcchunk_fc450f672b5723723b3d8eff07a935b1` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955270.510379` `source_timestamp=2026-01-21T00:27:50Z`
- U07TNT9N4JC asked for confirmation whether server-side TEE setup is needed. `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0` `chunk_id=srcchunk_2ba98043fe8a681baf343529bb54d19b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955303.708259` `source_timestamp=2026-01-21T00:28:23Z`

## Open Questions

- Does the Azure migration require re-setup of SGX/TEE configuration?
- Has the migration to Azure been completed? No follow-up in the thread.

## Sources

- `source_document_id`: `srcdoc_4420812f05fe9ccafb6aef71709ac54c`
- `source_revision_id`: `srcrev_25eaa5c06fe3761c8d52495875700cac`
