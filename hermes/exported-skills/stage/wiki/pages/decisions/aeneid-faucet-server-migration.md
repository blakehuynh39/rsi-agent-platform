---
title: "Aeneid Faucet Server Migration"
type: "decision"
slug: "decisions/aeneid-faucet-server-migration"
freshness: "2026-02-14T03:37:31Z"
tags:
  - "aeneid"
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners:
  - "S07ASS1JQNP"
  - "U0643B9PPC1"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_2729f1cdf2797dbe69c29554151a9905"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_5ac5d10a5bfbe6b45a289d398ba89e67"
  - "srcrev_7093544ae96b0aedcf62b57610d01155"
  - "srcrev_82fd93b3502f740baf241e47f0ce88d5"
  - "srcrev_b1862fd1ab8f88e3da964b7d58b94d39"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
  - "srcrev_e8de05b1a44b7de06d1dddaf2dc8140f"
conflict_state: "none"
---

# Aeneid Faucet Server Migration

## Summary

Migrate the Aeneid testnet faucet backend from a GCP instance (use1-aeneid-fuacet) to AWS, enabling domain-based routing instead of a hardcoded IP.

## Claims

- The faucet server is a GCP instance named use1-aeneid-fuacet. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_7093544ae96b0aedcf62b57610d01155` `chunk_id=srcchunk_453f872fbaf84c6caa00f9f5e0e28d0e` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-03T01:47:29Z`
- The faucet frontend is hosted at aeneid.faucet.story.foundation. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-03T02:22:50Z`
- The frontend queries the backend at http://35.207.0.103:23313/. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-03T08:19:25Z`
- The IP 35.207.0.103 belongs to the use1-aeneid-fuacet instance. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-13T08:29:16Z`
- No domain is attached to the faucet backend; the frontend uses the IP directly. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-13T08:03:17Z`
- The backend validates frontend requests, stores wallet addresses, and batch sends testnet tokens. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-13T08:03:17Z`
- The team plans to migrate the server from GCP to AWS to allow the frontend to use a domain instead of an IP. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-13T08:32:14Z`
- Haodi (U0643B9PPC1) needs access to both the old GCP instance and the new AWS instance to perform the migration. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_e8de05b1a44b7de06d1dddaf2dc8140f` `chunk_id=srcchunk_bd24a221873fb1e952e42e728631fb9f` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-13T09:31:54Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_82fd93b3502f740baf241e47f0ce88d5` `chunk_id=srcchunk_86cbb175607d8bad9b3a03bbe9422c5d` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-13T10:05:36Z`
- GCP access has been granted to Haodi; AWS access was still in progress. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-14T01:55:56Z`
- The GCP project is story-aeneid and the AWS account is story-testnet. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_2729f1cdf2797dbe69c29554151a9905` `chunk_id=srcchunk_2a3b7fa2db993f52fdcf4a3068729d22` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-13T23:18:25Z`
- Haodi’s piplabs.xyz email (Haodi@piplabs.xyz) is suspended, requiring an alternative contact method. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_5ac5d10a5bfbe6b45a289d398ba89e67` `chunk_id=srcchunk_78a933dfbae53f7515d2f8b7e8b63c02` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-14T03:27:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_b1862fd1ab8f88e3da964b7d58b94d39` `chunk_id=srcchunk_3f9f2b20e4bbf469f7cb397b114ddee9` `native_locator=slack:C0547N89JUB:1770083249.859729` `source_timestamp=2026-02-14T03:37:31Z`

## Open Questions

- Has the AWS instance been provisioned?
- What domain will the faucet backend use once migrated?
- What is Haodi’s active email for AWS/GCP access?
- When is the planned cutover date?

## Related Pages

- `aeneid-testnet-overview`
- `faucet-architecture`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_c5120943f11a1d4f582631497f49ea41`
