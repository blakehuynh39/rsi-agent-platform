---
title: "Aeneid Faucet Server (use1-aeneid-fuacet)"
type: "system"
slug: "systems/aeneid-faucet-server"
freshness: "2026-02-14T03:37:31Z"
tags:
  - "faucet"
  - "gcp"
  - "infrastructure"
  - "migration"
owners:
  - "subteam^S07ASS1JQNP"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_5ac5d10a5bfbe6b45a289d398ba89e67"
  - "srcrev_7093544ae96b0aedcf62b57610d01155"
  - "srcrev_82fd93b3502f740baf241e47f0ce88d5"
  - "srcrev_95c59ecd8381343566eb68b4867da4e2"
  - "srcrev_b1862fd1ab8f88e3da964b7d58b94d39"
  - "srcrev_c2f9042cf09f729cfa67f2387c5d853a"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# Aeneid Faucet Server (use1-aeneid-fuacet)

## Summary

GCP server hosting backend for aeneid.faucet.story.foundation. Frontend queries 35.207.0.103:23313. Handles faucet request validation and batch sends testnet tokens. Planned migration to AWS.

## Claims

- The faucet frontend at aeneid.faucet.story.foundation queries the backend at http://35.207.0.103:23313/ `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The GCP server instance is named use1-aeneid-fuacet in project story-aeneid, zone us-east1-c, with IP 35.207.0.103. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_7093544ae96b0aedcf62b57610d01155` `chunk_id=srcchunk_453f872fbaf84c6caa00f9f5e0e28d0e` `native_locator=slack:C0547N89JUB:1770083249.859729:1770083249.859729` `source_timestamp=2026-02-03T01:47:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The backend validates requests, adds wallet addresses to a database table, and batch sends testnet tokens to users. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- The faucet was developed by subteam S07ASS1JQNP. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`
- There is a plan to migrate the server from GCP to AWS, requiring the frontend to point to a domain instead of an IP. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- Access to GCP project story-aeneid and AWS account story-testnet was requested for user haodi@piplabs.xyz (GCP granted, AWS pending, but email suspended). `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_82fd93b3502f740baf241e47f0ce88d5` `chunk_id=srcchunk_86cbb175607d8bad9b3a03bbe9422c5d` `native_locator=slack:C0547N89JUB:1770083249.859729:1770977136.045799` `source_timestamp=2026-02-13T10:05:36Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_95c59ecd8381343566eb68b4867da4e2` `chunk_id=srcchunk_5280ec09f50a147393aa122e7ffc5977` `native_locator=slack:C0547N89JUB:1770083249.859729:1770979821.418289` `source_timestamp=2026-02-13T10:50:21Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_c2f9042cf09f729cfa67f2387c5d853a` `chunk_id=srcchunk_a16b7e5dac6bfab93b978f4ef7c8d090` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039440.489309` `source_timestamp=2026-02-14T03:24:00Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_5ac5d10a5bfbe6b45a289d398ba89e67` `chunk_id=srcchunk_78a933dfbae53f7515d2f8b7e8b63c02` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039649.305579` `source_timestamp=2026-02-14T03:27:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_b1862fd1ab8f88e3da964b7d58b94d39` `chunk_id=srcchunk_3f9f2b20e4bbf469f7cb397b114ddee9` `native_locator=slack:C0547N89JUB:1770083249.859729:1771040251.317959` `source_timestamp=2026-02-14T03:37:31Z`

## Open Questions

- Has the frontend been updated to use a domain instead of IP?
- Is haodi@piplabs.xyz resolved for access (poseidon email)?
- What is the status of the migration to AWS?

## Related Pages

- `aws-infrastructure`
- `gcp-infrastructure`
- `story-faucet`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_221c76e6d295e5bfb1568848bf28ace8`
