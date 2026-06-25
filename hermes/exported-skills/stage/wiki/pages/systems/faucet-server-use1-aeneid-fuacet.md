---
title: "Faucet Server (use1-aeneid-fuacet)"
type: "system"
slug: "systems/faucet-server-use1-aeneid-fuacet"
freshness: "2026-02-14T03:37:31Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners:
  - "U063MQ9ADCM"
  - "U0643B9PPC1"
  - "U07A7AUGL5V"
  - "U07TNT9N4JC"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_2729f1cdf2797dbe69c29554151a9905"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_5ac5d10a5bfbe6b45a289d398ba89e67"
  - "srcrev_b1862fd1ab8f88e3da964b7d58b94d39"
  - "srcrev_c2f9042cf09f729cfa67f2387c5d853a"
conflict_state: "none"
---

# Faucet Server (use1-aeneid-fuacet)

## Summary

The GCP compute instance use1-aeneid-fuacet (IP 35.207.0.103) hosts the backend for the Aeneid testnet faucet at https://aeneid.faucet.story.foundation/. The server validates frontend requests, stores wallet addresses, and batch sends testnet tokens. It is planned for migration to AWS with a domain-based frontend.

## Claims

- The server instance name is use1-aeneid-fuacet, a GCP compute instance in zone us-east1-c with IP 35.207.0.103. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The faucet frontend queries the backend at http://35.207.0.103:23313/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
- The backend validates the frontend request, adds wallet addresses to a database table, and chronologically batch sends testnet tokens to the user wallet. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- The server is hosted in the GCP project story-aeneid. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_2729f1cdf2797dbe69c29554151a9905` `chunk_id=srcchunk_2a3b7fa2db993f52fdcf4a3068729d22` `native_locator=slack:C0547N89JUB:1770083249.859729:1771024705.562899` `source_timestamp=2026-02-13T23:18:25Z`
- There is a plan to migrate the server to AWS (story-testnet account) and to assign a domain to the faucet frontend instead of a hardcoded IP. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- GCP access has been granted to the user @U0643B9PPC1 (Haodi) using the email Haodi@piplabs.xyz, while AWS access is still being processed. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_c2f9042cf09f729cfa67f2387c5d853a` `chunk_id=srcchunk_a16b7e5dac6bfab93b978f4ef7c8d090` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039440.489309` `source_timestamp=2026-02-14T03:24:00Z`
- The user @U0643B9PPC1 mentioned that their piplabs email is suspended, potentially blocking access; a poseidon email might be needed. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_5ac5d10a5bfbe6b45a289d398ba89e67` `chunk_id=srcchunk_78a933dfbae53f7515d2f8b7e8b63c02` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039649.305579` `source_timestamp=2026-02-14T03:27:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_b1862fd1ab8f88e3da964b7d58b94d39` `chunk_id=srcchunk_3f9f2b20e4bbf469f7cb397b114ddee9` `native_locator=slack:C0547N89JUB:1770083249.859729:1771040251.317959` `source_timestamp=2026-02-14T03:37:31Z`

## Open Questions

- Has the faucet frontend been updated to use a domain name instead of the IP?
- Has the faucet server been successfully migrated to AWS?
- Will the email issue (piplabs suspension) be resolved, or will a poseidon email be used for access?

## Related Pages

- `aeneid-faucet-frontend`
- `aws-story-testnet`
- `gcp-story-aeneid`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_ba2250fabe6a4d7b63d6a0abe1711a2a`
