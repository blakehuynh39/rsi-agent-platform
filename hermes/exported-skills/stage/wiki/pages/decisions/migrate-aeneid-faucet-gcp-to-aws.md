---
title: "Migrate Aeneid Faucet from GCP to AWS"
type: "decision"
slug: "decisions/migrate-aeneid-faucet-gcp-to-aws"
freshness: "2026-02-14T03:37:31Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners:
  - "@U0643B9PPC1"
  - "@U07A7AUGL5V"
  - "@U07TNT9N4JC"
  - "@U08332YRB7W"
source_revision_ids:
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_2729f1cdf2797dbe69c29554151a9905"
  - "srcrev_330da0aee08bde66fba660b900c252c5"
  - "srcrev_4d66803ea8e09d84ceac3173f335a96a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_5ac5d10a5bfbe6b45a289d398ba89e67"
  - "srcrev_b1862fd1ab8f88e3da964b7d58b94d39"
  - "srcrev_e8de05b1a44b7de06d1dddaf2dc8140f"
conflict_state: "none"
---

# Migrate Aeneid Faucet from GCP to AWS

## Summary

Decision to migrate the Aeneid faucet backend from GCP (project story-aeneid) to AWS (account story-testnet). The migration aims to terminate GCP instances. Access to both environments has been requested for key personnel. The frontend will be updated to point to a domain instead of an IP address to facilitate the migration.

## Claims

- We are trying to terminate GCP side instances. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_330da0aee08bde66fba660b900c252c5` `chunk_id=srcchunk_0a4f8394446ed81066bffa793f077518` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969849.588979` `source_timestamp=2026-02-13T08:04:09Z`
- We want to migrate the server to AWS. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- The frontend should point to a domain instead of IP to allow migration. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- The GCP project is story-aeneid and the AWS account is story-testnet. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_2729f1cdf2797dbe69c29554151a9905` `chunk_id=srcchunk_2a3b7fa2db993f52fdcf4a3068729d22` `native_locator=slack:C0547N89JUB:1770083249.859729:1771024705.562899` `source_timestamp=2026-02-13T23:18:25Z`
- GCP access has been granted for @U0643B9PPC1; AWS access is pending. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
- The login email for GCP is Haodi@piplabs.xyz, but it may be suspended. `claim:claim_2_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_5ac5d10a5bfbe6b45a289d398ba89e67` `chunk_id=srcchunk_78a933dfbae53f7515d2f8b7e8b63c02` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039649.305579` `source_timestamp=2026-02-14T03:27:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_b1862fd1ab8f88e3da964b7d58b94d39` `chunk_id=srcchunk_3f9f2b20e4bbf469f7cb397b114ddee9` `native_locator=slack:C0547N89JUB:1770083249.859729:1771040251.317959` `source_timestamp=2026-02-14T03:37:31Z`
- @U07TNT9N4JC is asked to setup the AWS instance. `claim:claim_2_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- @U0643B9PPC1 needs server access to the new instance. `claim:claim_2_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4d66803ea8e09d84ceac3173f335a96a` `chunk_id=srcchunk_f456dad63becacd846c449d6407daf7e` `native_locator=slack:C0547N89JUB:1770083249.859729:1770974946.176129` `source_timestamp=2026-02-13T09:29:06Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_e8de05b1a44b7de06d1dddaf2dc8140f` `chunk_id=srcchunk_bd24a221873fb1e952e42e728631fb9f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770975114.879239` `source_timestamp=2026-02-13T09:31:54Z`

## Open Questions

- Has the AWS instance been set up?
- Has the migration been completed?
- What domain name will be used for the backend?

## Related Pages

- `aeneid-faucet`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_561d85e42049bebf64540cb5c5545848`
