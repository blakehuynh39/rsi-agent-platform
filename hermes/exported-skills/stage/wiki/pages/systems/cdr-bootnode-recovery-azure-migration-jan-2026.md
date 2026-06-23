---
title: "CDR Bootnode Recovery and Azure Migration (2026-01-20)"
type: "system"
slug: "systems/cdr-bootnode-recovery-azure-migration-jan-2026"
freshness: "2026-01-21T00:28:23Z"
tags:
  - "aws"
  - "azure"
  - "bootnode"
  - "cdr"
  - "migration"
  - "recovery"
owners: []
source_revision_ids:
  - "srcrev_07a191138ceca2aea3e7a173a8a199e7"
  - "srcrev_10d5913d78928b9a1bfc247e1b5674c1"
  - "srcrev_4625faf5add9f365f1abd30b13a4663b"
  - "srcrev_5f0b8e5c43f79f57753f0838a16d9cee"
  - "srcrev_6792157d8f44a2c1416d2b571a2be033"
  - "srcrev_86706658b4a5efdec62ee819b2ab5a33"
  - "srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0"
  - "srcrev_a68b1287dadba8b820575f4d39379aa8"
  - "srcrev_cfd931a15d500789656c173e0645efe6"
  - "srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02"
  - "srcrev_fb369332858d64661f0921a2e5970b01"
conflict_state: "none"
---

# CDR Bootnode Recovery and Azure Migration (2026-01-20)

## Summary

On 2026-01-20, the CDR bootnode became inaccessible due to an unintended AWS account closure. The account was restored and the bootnode was found in the story-services-staging account. The team decided to migrate the bootnode to a company Azure account, and updated the bootnode IP to 54.153.69.163. Open questions remain regarding SGX configuration on the new machine.

## Claims

- A user reported being unable to access the CDR bootnode and suspected the previous AWS account had been removed. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02` `chunk_id=srcchunk_baeadf16c43c8be70c8f8f4ecf26e3aa` `native_locator=slack:C0547N89JUB:1768898200.116689:1768898200.116689` `source_timestamp=2026-01-20T08:36:40Z`
- The missing bootnode was initially believed to be in the story-prod or story-services-prod AWS account. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_cfd931a15d500789656c173e0645efe6` `chunk_id=srcchunk_de3e4a496737c915025ea80f03a2c38f` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912531.110089` `source_timestamp=2026-01-20T12:35:31Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5f0b8e5c43f79f57753f0838a16d9cee` `chunk_id=srcchunk_ea42c9c08496543d3d71a2c656780df8` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945728.509969` `source_timestamp=2026-01-20T21:48:48Z`
- A request was made to create an Azure subscription on the company card for migrating the bootnode to Story Azure. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_07a191138ceca2aea3e7a173a8a199e7` `chunk_id=srcchunk_f7df99a90dcfddff8ff9bc872c83e7fc` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912588.205689` `source_timestamp=2026-01-20T12:36:28Z`
- The AWS account was restored by a team member, requiring SMS 2FA verification. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4625faf5add9f365f1abd30b13a4663b` `chunk_id=srcchunk_fc6b4b9a9a1dbf0b3bd8d9f41c996d55` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946049.112209` `source_timestamp=2026-01-20T21:54:09Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_10d5913d78928b9a1bfc247e1b5674c1` `chunk_id=srcchunk_5b326265221430247237293aec04819b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768948086.089599` `source_timestamp=2026-01-20T22:28:06Z`
- After restoration, a check in us-east-1 region found no bootnode instances; additional accounts were restored. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_6792157d8f44a2c1416d2b571a2be033` `chunk_id=srcchunk_0fad24a4ce91aea715fee53b2baf1dd2` `native_locator=slack:C0547N89JUB:1768898200.116689:1768951206.135269` `source_timestamp=2026-01-20T23:20:06Z`
- The CDR bootnode was eventually located in the story-services-staging AWS account. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0` `chunk_id=srcchunk_2ba98043fe8a681baf343529bb54d19b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955303.708259` `source_timestamp=2026-01-21T00:28:23Z`
- The bootnode IP address was changed to 54.153.69.163, requiring only a configuration update for clients. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a68b1287dadba8b820575f4d39379aa8` `chunk_id=srcchunk_1ba73dd4cad40a349168deb910d30a45` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955118.522269` `source_timestamp=2026-01-21T00:25:18Z`
- A decision was made to migrate the CDR bootnode to a company Azure subscription. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_86706658b4a5efdec62ee819b2ab5a33` `chunk_id=srcchunk_e3c1d4786874fce1043b88b377ef2ab6` `native_locator=slack:C0547N89JUB:1768898200.116689:1768949438.340279` `source_timestamp=2026-01-20T22:50:38Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a68b1287dadba8b820575f4d39379aa8` `chunk_id=srcchunk_1ba73dd4cad40a349168deb910d30a45` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955118.522269` `source_timestamp=2026-01-21T00:25:18Z`
- An open question was raised about the need to re-setup SGX (TEE) configuration on the new machine. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_fb369332858d64661f0921a2e5970b01` `chunk_id=srcchunk_fc450f672b5723723b3d8eff07a935b1` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955270.510379` `source_timestamp=2026-01-21T00:27:50Z`

## Open Questions

- Should the SGX related configuration be re-setup on the new Azure machine?

## Sources

- `source_document_id`: `srcdoc_4420812f05fe9ccafb6aef71709ac54c`
- `source_revision_id`: `srcrev_a3b2c91a4d5ee0e334bdaaebb4eda36e`
