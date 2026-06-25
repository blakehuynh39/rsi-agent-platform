---
title: "Aeneid Faucet Server (use1-aeneid-fuacet)"
type: "system"
slug: "systems/aeneid-faucet-server"
freshness: "2026-02-14T03:37:31Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "infrastructure"
  - "migration"
owners:
  - "Faucet Dev Team (S07ASS1JQNP)"
  - "Haodi (U0643B9PPC1) (Migration)"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_b1862fd1ab8f88e3da964b7d58b94d39"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# Aeneid Faucet Server (use1-aeneid-fuacet)

## Summary

Infrastructure details and migration plans for the Aeneid faucet server, currently hosted on GCP (instance use1-aeneid-fuacet, IP 35.207.0.103), with frontend at aeneid.faucet.story.foundation. Migration to AWS is planned, requiring frontend to switch from IP to domain.

## Claims

- The faucet backend server is a GCP Compute Instance in us-east1-c named use1-aeneid-fuacet, with IP address 35.207.0.103. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The faucet frontend, hosted at https://aeneid.faucet.story.foundation/, sends requests to the backend at http://35.207.0.103:23313/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`
- The backend validates frontend requests, adds wallet addresses to a database table, and batch sends testnet tokens chronologically. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- The Aeneid faucet is currently hosted on GCP, and a migration to AWS is planned. The frontend needs to switch from using an IP address to a domain to facilitate migration. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- Access to GCP project story-aeneid and AWS account story-testnet has been requested for U0643B9PPC1 (Haodi) to perform the migration. GCP access was granted, AWS access was being arranged, but Haodi's piplabs.xyz email may be suspended. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_b1862fd1ab8f88e3da964b7d58b94d39` `chunk_id=srcchunk_3f9f2b20e4bbf469f7cb397b114ddee9` `native_locator=slack:C0547N89JUB:1770083249.859729:1771040251.317959` `source_timestamp=2026-02-14T03:37:31Z`

## Open Questions

- How to provide AWS access for Haodi given his piplabs.xyz email is suspended? (Possibly use poseidon email.)

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_746d49611a9cf0e72b071f5445df81cd`
