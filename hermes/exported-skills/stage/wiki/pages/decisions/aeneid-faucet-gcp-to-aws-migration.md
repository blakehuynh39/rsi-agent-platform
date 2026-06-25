---
title: "Migrate Aeneid Faucet from GCP to AWS"
type: "decision"
slug: "decisions/aeneid-faucet-gcp-to-aws-migration"
freshness: "2026-02-14T03:37:31Z"
tags:
  - "faucet"
  - "infrastructure"
  - "migration"
owners:
  - "S07ASS1JQNP"
  - "U0643B9PPC1"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_2729f1cdf2797dbe69c29554151a9905"
  - "srcrev_330da0aee08bde66fba660b900c252c5"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_5ac5d10a5bfbe6b45a289d398ba89e67"
  - "srcrev_82fd93b3502f740baf241e47f0ce88d5"
  - "srcrev_95c59ecd8381343566eb68b4867da4e2"
  - "srcrev_b1862fd1ab8f88e3da964b7d58b94d39"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# Migrate Aeneid Faucet from GCP to AWS

## Summary

Decision to migrate the Aeneid faucet backend from GCP instance use1-aeneid-fuacet to AWS, and to update the faucet frontend from using an IP address to a domain name for easier future migrations.

## Claims

- The current Aeneid faucet backend runs on GCP instance use1-aeneid-fuacet with IP address 35.207.0.103. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
- The faucet frontend queries the backend at http://35.207.0.103:23313/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
- The faucet was developed by the subteam S07ASS1JQNP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`
- The backend validates frontend requests, adds wallet addresses to a database table, and batch-sends testnet tokens to users. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- There is a decision to migrate the faucet backend from GCP to AWS. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_330da0aee08bde66fba660b900c252c5` `chunk_id=srcchunk_0a4f8394446ed81066bffa793f077518` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969849.588979` `source_timestamp=2026-02-13T08:04:09Z`
- The frontend should be updated to point to a domain name instead of an IP address to facilitate migration. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- Access to GCP project story-aeneid and AWS account story-testnet is being granted to user U0643B9PPC1 (Haodi). `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_95c59ecd8381343566eb68b4867da4e2` `chunk_id=srcchunk_5280ec09f50a147393aa122e7ffc5977` `native_locator=slack:C0547N89JUB:1770083249.859729:1770979821.418289` `source_timestamp=2026-02-13T10:50:21Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_2729f1cdf2797dbe69c29554151a9905` `chunk_id=srcchunk_2a3b7fa2db993f52fdcf4a3068729d22` `native_locator=slack:C0547N89JUB:1770083249.859729:1771024705.562899` `source_timestamp=2026-02-13T23:18:25Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_82fd93b3502f740baf241e47f0ce88d5` `chunk_id=srcchunk_86cbb175607d8bad9b3a03bbe9422c5d` `native_locator=slack:C0547N89JUB:1770083249.859729:1770977136.045799` `source_timestamp=2026-02-13T10:05:36Z`
- The email haodi@piplabs.xyz is suspended, and an alternative poseidon email may be needed for AWS login. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_5ac5d10a5bfbe6b45a289d398ba89e67` `chunk_id=srcchunk_78a933dfbae53f7515d2f8b7e8b63c02` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039649.305579` `source_timestamp=2026-02-14T03:27:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_b1862fd1ab8f88e3da964b7d58b94d39` `chunk_id=srcchunk_3f9f2b20e4bbf469f7cb397b114ddee9` `native_locator=slack:C0547N89JUB:1770083249.859729:1771040251.317959` `source_timestamp=2026-02-14T03:37:31Z`

## Open Questions

- Has the new AWS instance been created?
- How will the frontend be updated to use the domain?
- What domain will the frontend use?
- When will the migration be completed?
- Who will set up the AWS instance (U07TNT9N4JC)?

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_15036f8cdcd7730d9bf3ef252d63cd08`
