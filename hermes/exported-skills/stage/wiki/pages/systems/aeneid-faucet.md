---
title: "Aeneid Faucet Service"
type: "system"
slug: "systems/aeneid-faucet"
freshness: "2026-02-14T01:55:56Z"
tags:
  - "aeneid"
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
  - "testnet"
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
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# Aeneid Faucet Service

## Summary

The Aeneid faucet is a web service that distributes testnet tokens. It has a frontend at https://aeneid.faucet.story.foundation/ and a backend API currently hosted on a GCP instance. A migration to AWS is planned.

## Claims

- The Aeneid faucet frontend is at https://aeneid.faucet.story.foundation/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`
- The faucet frontend queries backend API at http://35.207.0.103:23313/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
- The backend is hosted on GCP instance 'use1-aeneid-fuacet' in project 'story-aeneid', with IP 35.207.0.103. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- Backend validates frontend request, adds wallet addresses to a database, then batch sends testnet tokens. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- Migration to AWS is planned, requiring frontend to use a domain name instead of IP for backend. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- GCP access to project story-aeneid has been granted to Haodi (piplabs) for migration; AWS access to account story-testnet is being arranged. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_2729f1cdf2797dbe69c29554151a9905` `chunk_id=srcchunk_2a3b7fa2db993f52fdcf4a3068729d22` `native_locator=slack:C0547N89JUB:1770083249.859729:1771024705.562899` `source_timestamp=2026-02-13T23:18:25Z`

## Open Questions

- What is the target domain name for the migrated backend?
- What is the timeline for the AWS migration?
- Who will be responsible for ongoing maintenance after migration?

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_4f7762d903550900445eec55650e231b`
