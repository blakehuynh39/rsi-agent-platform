---
title: "Aeneid Faucet Migration to AWS"
type: "decision"
slug: "decisions/aeneid-faucet-migration-decision"
freshness: "2026-02-14T01:55:56Z"
tags:
  - "aeneid"
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners:
  - "U0643B9PPC1"
  - "U07A7AUGL5V"
  - "U07TNT9N4JC"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_e8de05b1a44b7de06d1dddaf2dc8140f"
conflict_state: "none"
---

# Aeneid Faucet Migration to AWS

## Summary

Decision to migrate the Aeneid faucet backend from GCP to AWS, requiring frontend to use a domain name instead of hardcoded IP. Access granted for GCP, pending for AWS as of Feb 2026.

## Claims

- Migration to AWS is desired; need frontend to point to domain instead of IP to enable migration `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- GCP access for @U0643B9PPC1 (haodi@piplabs.xyz) has been granted `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
- AWS access for @U0643B9PPC1 is pending as of thread `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
- @U0643B9PPC1 needs access to both old GCP and new AWS instances `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_e8de05b1a44b7de06d1dddaf2dc8140f` `chunk_id=srcchunk_bd24a221873fb1e952e42e728631fb9f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770975114.879239` `source_timestamp=2026-02-13T09:31:54Z`
- @U07TNT9N4JC may help set up the AWS instance `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`

## Open Questions

- Has the migration been completed? What is the new domain and AWS instance details? When will the GCP instance be terminated?

## Related Pages

- `aeneid-faucet`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_41aa38518a6479f8c26e3648c70793ab`
