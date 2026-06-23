---
title: "AWS Account Recovery for CDR Bootnode"
type: "runbook"
slug: "runbooks/aws-account-recovery-cdr-bootnode"
freshness: "2026-01-21T00:28:23Z"
tags:
  - "aws"
  - "bootnode"
  - "incident"
  - "recovery"
owners:
  - "U079ZJ48D62"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_10d5913d78928b9a1bfc247e1b5674c1"
  - "srcrev_1e33f666797775a86cd17ec2e662948b"
  - "srcrev_2896b27d572c72e037fe7a85b367e198"
  - "srcrev_4625faf5add9f365f1abd30b13a4663b"
  - "srcrev_4c13d34f8a82c90092077549b23c8ab8"
  - "srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0"
  - "srcrev_912b3eb6cb0432731042200191c7cf00"
  - "srcrev_a3b2c91a4d5ee0e334bdaaebb4eda36e"
  - "srcrev_a68b1287dadba8b820575f4d39379aa8"
  - "srcrev_cfd931a15d500789656c173e0645efe6"
  - "srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02"
  - "srcrev_fe142cd81e3357fb163f0c67831e8f4c"
conflict_state: "none"
---

# AWS Account Recovery for CDR Bootnode

## Summary

Recovery of a mistakenly closed AWS account that hosted the CDR bootnode. The bootnode was eventually found in story-services-staging.

## Claims

- The previous AWS account story-services-prod was closed, making CDR bootnode inaccessible. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_f2dc2fbe6e00a8d34cb08c9bba827a02` `chunk_id=srcchunk_baeadf16c43c8be70c8f8f4ecf26e3aa` `native_locator=slack:C0547N89JUB:1768898200.116689:1768898200.116689` `source_timestamp=2026-01-20T08:36:40Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_cfd931a15d500789656c173e0645efe6` `chunk_id=srcchunk_de3e4a496737c915025ea80f03a2c38f` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912531.110089` `source_timestamp=2026-01-20T12:35:31Z`
- Recovery of the account was requested and initiated. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_912b3eb6cb0432731042200191c7cf00` `chunk_id=srcchunk_508dc8620722e7309aab5fed0c172e36` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945660.248789` `source_timestamp=2026-01-20T21:47:40Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4c13d34f8a82c90092077549b23c8ab8` `chunk_id=srcchunk_48cf249c84184a95ab960b2e2e3905ce` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945746.724709` `source_timestamp=2026-01-20T21:49:06Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_2896b27d572c72e037fe7a85b367e198` `chunk_id=srcchunk_61b899213fccc7ca5ff0fec6e295ec90` `native_locator=slack:C0547N89JUB:1768898200.116689:1768945789.712699` `source_timestamp=2026-01-20T21:49:49Z`
- A 2FA code from U08D32C1EF3 was required and obtained to complete the recovery. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_4625faf5add9f365f1abd30b13a4663b` `chunk_id=srcchunk_fc6b4b9a9a1dbf0b3bd8d9f41c996d55` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946049.112209` `source_timestamp=2026-01-20T21:54:09Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_fe142cd81e3357fb163f0c67831e8f4c` `chunk_id=srcchunk_fd3ddfabaa94d37d17ea57d69623f9f0` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946065.730079` `source_timestamp=2026-01-20T21:54:25Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a3b2c91a4d5ee0e334bdaaebb4eda36e` `chunk_id=srcchunk_1a8e19dcfd803022b9cbcf9479011266` `native_locator=slack:C0547N89JUB:1768898200.116689:1768946782.054029` `source_timestamp=2026-01-20T22:06:22Z`
- The account was successfully restored, and all accounts were later restored. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_10d5913d78928b9a1bfc247e1b5674c1` `chunk_id=srcchunk_5b326265221430247237293aec04819b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768948086.089599` `source_timestamp=2026-01-20T22:28:06Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_1e33f666797775a86cd17ec2e662948b` `chunk_id=srcchunk_cc2cdb9011bae4171e103fa283b754d3` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955255.580949` `source_timestamp=2026-01-21T00:27:35Z`
- The CDR bootnode was found in story-services-staging, not story-services-prod. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_8dc1bbef56cb6a29b68e85edc5dd62b0` `chunk_id=srcchunk_2ba98043fe8a681baf343529bb54d19b` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955303.708259` `source_timestamp=2026-01-21T00:28:23Z`
- The bootnode IP changed to 54.153.69.163. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_a68b1287dadba8b820575f4d39379aa8` `chunk_id=srcchunk_1ba73dd4cad40a349168deb910d30a45` `native_locator=slack:C0547N89JUB:1768898200.116689:1768955118.522269` `source_timestamp=2026-01-21T00:25:18Z`

## Open Questions

- Should SGX/TEE related configuration be re-setup on the new machine?

## Related Pages

- `cdr-bootnode-migration-to-azure`

## Sources

- `source_document_id`: `srcdoc_4420812f05fe9ccafb6aef71709ac54c`
- `source_revision_id`: `srcrev_4625faf5add9f365f1abd30b13a4663b`
