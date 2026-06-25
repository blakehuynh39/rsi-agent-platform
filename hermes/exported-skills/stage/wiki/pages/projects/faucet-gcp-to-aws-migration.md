---
title: "Faucet Migration from GCP to AWS"
type: "project"
slug: "projects/faucet-gcp-to-aws-migration"
freshness: "2026-02-14T03:37:31Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "infrastructure"
  - "migration"
owners: []
source_revision_ids:
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_2729f1cdf2797dbe69c29554151a9905"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_82fd93b3502f740baf241e47f0ce88d5"
  - "srcrev_95c59ecd8381343566eb68b4867da4e2"
  - "srcrev_b1862fd1ab8f88e3da964b7d58b94d39"
  - "srcrev_c2f9042cf09f729cfa67f2387c5d853a"
  - "srcrev_e8de05b1a44b7de06d1dddaf2dc8140f"
conflict_state: "none"
---

# Faucet Migration from GCP to AWS

## Summary

Decision to migrate the Aeneid faucet backend from GCP instance use1-aeneid-fuacet (project story-aeneid) to AWS (account story-testnet). The frontend will be updated to use a domain instead of IP. Access coordination is ongoing.

## Claims

- Decision was made to migrate the faucet backend from GCP to AWS. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- The frontend currently points to the IP; it is requested to update to a domain to facilitate the migration. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- The GCP instance to migrate is use1-aeneid-fuacet in project story-aeneid. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_2729f1cdf2797dbe69c29554151a9905` `chunk_id=srcchunk_2a3b7fa2db993f52fdcf4a3068729d22` `native_locator=slack:C0547N89JUB:1770083249.859729:1771024705.562899` `source_timestamp=2026-02-13T23:18:25Z`
- The target AWS account is story-testnet. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_2729f1cdf2797dbe69c29554151a9905` `chunk_id=srcchunk_2a3b7fa2db993f52fdcf4a3068729d22` `native_locator=slack:C0547N89JUB:1770083249.859729:1771024705.562899` `source_timestamp=2026-02-13T23:18:25Z`
- User U0643B9PPC1 (Haodi) needs access to both GCP and AWS for migration or documentation. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_82fd93b3502f740baf241e47f0ce88d5` `chunk_id=srcchunk_86cbb175607d8bad9b3a03bbe9422c5d` `native_locator=slack:C0547N89JUB:1770083249.859729:1770977136.045799` `source_timestamp=2026-02-13T10:05:36Z`
- GCP access to story-aeneid was granted to U0643B9PPC1. `claim:claim_2_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
- AWS access (including SSM) to story-testnet was requested for U0643B9PPC1 and was still in progress. `claim:claim_2_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_95c59ecd8381343566eb68b4867da4e2` `chunk_id=srcchunk_5280ec09f50a147393aa122e7ffc5977` `native_locator=slack:C0547N89JUB:1770083249.859729:1770979821.418289` `source_timestamp=2026-02-13T10:50:21Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
- The user U0643B9PPC1's piplabs email (haodi@piplabs.xyz) is suspended, potentially requiring an alternative login for AWS. `claim:claim_2_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_b1862fd1ab8f88e3da964b7d58b94d39` `chunk_id=srcchunk_3f9f2b20e4bbf469f7cb397b114ddee9` `native_locator=slack:C0547N89JUB:1770083249.859729:1771040251.317959` `source_timestamp=2026-02-14T03:37:31Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_c2f9042cf09f729cfa67f2387c5d853a` `chunk_id=srcchunk_a16b7e5dac6bfab93b978f4ef7c8d090` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039440.489309` `source_timestamp=2026-02-14T03:24:00Z`
- Setting up the AWS instance may require assistance from U07TNT9N4JC. `claim:claim_2_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- U0643B9PPC1 indicated need for access to both old (GCP) and new (AWS) servers to perform the migration setup. `claim:claim_2_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_e8de05b1a44b7de06d1dddaf2dc8140f` `chunk_id=srcchunk_bd24a221873fb1e952e42e728631fb9f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770975114.879239` `source_timestamp=2026-02-13T09:31:54Z`

## Open Questions

- Has the AWS instance been fully set up?
- Has the frontend been updated to use a domain?
- Is U0643B9PPC1 able to access AWS despite the suspended piplabs email?
- What is the timeline for completing the migration?

## Related Pages

- `aeneid-faucet`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_50b60006f4eeb38f39acd80cb65b8cff`
