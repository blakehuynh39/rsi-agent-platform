---
title: "use1-aeneid-fuacet Faucet Server"
type: "system"
slug: "systems/use1-aeneid-fuacet-server"
freshness: "2026-02-14T01:55:56Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners:
  - "S07ASS1JQNP"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_2729f1cdf2797dbe69c29554151a9905"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_7093544ae96b0aedcf62b57610d01155"
  - "srcrev_7a75154558963d681535c20db1ffa060"
  - "srcrev_95c59ecd8381343566eb68b4867da4e2"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# use1-aeneid-fuacet Faucet Server

## Summary

GCP compute instance serving as backend for the Aeneid testnet faucet. It validates requests and batch-sends testnet tokens. Plans to migrate to AWS.

## Claims

- The faucet backend server is a GCP compute instance named 'use1-aeneid-fuacet' in zone us-east1-c, project 'story-aeneid'. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_7093544ae96b0aedcf62b57610d01155` `chunk_id=srcchunk_453f872fbaf84c6caa00f9f5e0e28d0e` `native_locator=slack:C0547N89JUB:1770083249.859729:1770083249.859729` `source_timestamp=2026-02-03T01:47:29Z`
- The frontend (aeneid.faucet.story.foundation) queries the backend at IP 35.207.0.103 on port 23313. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The backend validates frontend requests, adds wallet addresses to a database, and chronologically batch-sends testnet tokens. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- The development team for the faucet is subteam S07ASS1JQNP. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_7a75154558963d681535c20db1ffa060` `chunk_id=srcchunk_c1e3992cb7fb8a91cdd0b72b11718a94` `native_locator=slack:C0547N89JUB:1770083249.859729:1770965444.096269` `source_timestamp=2026-02-13T06:50:44Z`
- There is a plan to migrate the server from GCP to AWS and change the frontend to use a domain instead of an IP address. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- The AWS account intended for migration is 'story-testnet'. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_2729f1cdf2797dbe69c29554151a9905` `chunk_id=srcchunk_2a3b7fa2db993f52fdcf4a3068729d22` `native_locator=slack:C0547N89JUB:1770083249.859729:1771024705.562899` `source_timestamp=2026-02-13T23:18:25Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_95c59ecd8381343566eb68b4867da4e2` `chunk_id=srcchunk_5280ec09f50a147393aa122e7ffc5977` `native_locator=slack:C0547N89JUB:1770083249.859729:1770979821.418289` `source_timestamp=2026-02-13T10:50:21Z`
- GCP access was granted to U0643B9PPC1 for migration purposes. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`

## Open Questions

- Has the frontend been updated to use a domain for the backend?
- What is the current status of the migration to AWS?

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_718b7a43d21f7645fb54207e81146679`
