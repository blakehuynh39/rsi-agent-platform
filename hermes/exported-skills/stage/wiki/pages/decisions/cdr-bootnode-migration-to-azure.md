---
title: "CDR Bootnode Migration to Azure"
type: "decision"
slug: "decisions/cdr-bootnode-migration-to-azure"
freshness: "2026-01-20T22:50:38Z"
tags:
  - "azure"
  - "bootnode"
  - "cdr"
  - "migration"
owners:
  - "U079ZJ48D62"
  - "U07TNT9N4JC"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_07a191138ceca2aea3e7a173a8a199e7"
  - "srcrev_5abeeff78e095b18bc26e8748e103f29"
  - "srcrev_86706658b4a5efdec62ee819b2ab5a33"
conflict_state: "none"
---

# CDR Bootnode Migration to Azure

## Summary

Plan to migrate the CDR bootnode to a company Azure subscription, replacing the AWS setup.

## Claims

- There is a plan to migrate the CDR bootnode to a company Azure account. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_07a191138ceca2aea3e7a173a8a199e7` `chunk_id=srcchunk_f7df99a90dcfddff8ff9bc872c83e7fc` `native_locator=slack:C0547N89JUB:1768898200.116689:1768912588.205689` `source_timestamp=2026-01-20T12:36:28Z`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_5abeeff78e095b18bc26e8748e103f29` `chunk_id=srcchunk_1be807cd0340f11193388886dd219973` `native_locator=slack:C0547N89JUB:1768898200.116689:1768927250.985599` `source_timestamp=2026-01-20T16:40:50Z`
- U07TNT9N4JC will prepare the CDR on the company Azure account. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4420812f05fe9ccafb6aef71709ac54c` `source_revision_id=srcrev_86706658b4a5efdec62ee819b2ab5a33` `chunk_id=srcchunk_e3c1d4786874fce1043b88b377ef2ab6` `native_locator=slack:C0547N89JUB:1768898200.116689:1768949438.340279` `source_timestamp=2026-01-20T22:50:38Z`

## Open Questions

- Is it okay to migrate the current CDR devnet to the company Azure account?

## Related Pages

- `aws-account-recovery-cdr-bootnode`

## Sources

- `source_document_id`: `srcdoc_4420812f05fe9ccafb6aef71709ac54c`
- `source_revision_id`: `srcrev_4625faf5add9f365f1abd30b13a4663b`
