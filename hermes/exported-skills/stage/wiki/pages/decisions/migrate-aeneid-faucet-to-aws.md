---
title: "Migrate Aeneid Faucet from GCP to AWS"
type: "decision"
slug: "decisions/migrate-aeneid-faucet-to-aws"
freshness: "2026-02-14T01:55:56Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners: []
source_revision_ids:
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_82fd93b3502f740baf241e47f0ce88d5"
  - "srcrev_e8de05b1a44b7de06d1dddaf2dc8140f"
conflict_state: "none"
---

# Migrate Aeneid Faucet from GCP to AWS

## Summary

Decision to migrate the Aeneid faucet backend from GCP to AWS and update the frontend to use a domain name.

## Claims

- The team decided to migrate the faucet server from GCP to AWS. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- The frontend should be updated to point to a domain instead of an IP address to facilitate migration. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- Access to both the old GCP server and the new AWS server is required for migration. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_e8de05b1a44b7de06d1dddaf2dc8140f` `chunk_id=srcchunk_bd24a221873fb1e952e42e728631fb9f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770975114.879239` `source_timestamp=2026-02-13T09:31:54Z`
- Permissions were being arranged for GCP project story-aeneid and AWS account story-testnet to support the migration. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_82fd93b3502f740baf241e47f0ce88d5` `chunk_id=srcchunk_86cbb175607d8bad9b3a03bbe9422c5d` `native_locator=slack:C0547N89JUB:1770083249.859729:1770977136.045799` `source_timestamp=2026-02-13T10:05:36Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`

## Open Questions

- Is the migration complete?
- What domain name will be used for the new backend?

## Related Pages

- `aeneid-faucet`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_bf67d2e4f7df453b626f423673a760aa`
