---
title: "Aeneid Faucet Server"
type: "system"
slug: "systems/aeneid-faucet-server"
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
  - "U063MQ9ADCM"
  - "U0643B9PPC1"
  - "U07TNT9N4JC"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_330da0aee08bde66fba660b900c252c5"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_7093544ae96b0aedcf62b57610d01155"
  - "srcrev_82fd93b3502f740baf241e47f0ce88d5"
  - "srcrev_8ffdb35cdf6746fb873cd6a2f7d14584"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# Aeneid Faucet Server

## Summary

Backend server for the Aeneid testnet faucet frontend. Currently hosted on GCP instance use1-aeneid-fuacet (IP 35.207.0.103) in project story-aeneid. Plans to migrate to AWS and assign a domain name.

## Claims

- The faucet frontend queries the backend server at http://35.207.0.103:23313/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The server is a GCP instance named use1-aeneid-fuacet in project story-aeneid, zone us-east1-c. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_7093544ae96b0aedcf62b57610d01155` `chunk_id=srcchunk_453f872fbaf84c6caa00f9f5e0e28d0e` `native_locator=slack:C0547N89JUB:1770083249.859729:1770083249.859729` `source_timestamp=2026-02-03T01:47:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The backend validates faucet requests, stores wallet addresses in a database, and batch-sends testnet tokens periodically. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- As of 2026-02-13, the team was planning to terminate GCP instances, implying deprecation of the current faucet server. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_330da0aee08bde66fba660b900c252c5` `chunk_id=srcchunk_0a4f8394446ed81066bffa793f077518` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969849.588979` `source_timestamp=2026-02-13T08:04:09Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_8ffdb35cdf6746fb873cd6a2f7d14584` `chunk_id=srcchunk_f1d2a4b3f82a9e0e4bb4d060060a5cad` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969923.789949` `source_timestamp=2026-02-13T08:05:23Z`
- Migration to AWS is planned; the frontend should be updated to use a domain name instead of a hardcoded IP to enable seamless migration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- GCP access has been granted to user U0643B9PPC1 for the migration; AWS access was pending. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_82fd93b3502f740baf241e47f0ce88d5` `chunk_id=srcchunk_86cbb175607d8bad9b3a03bbe9422c5d` `native_locator=slack:C0547N89JUB:1770083249.859729:1770977136.045799` `source_timestamp=2026-02-13T10:05:36Z`
- The faucet was developed by subteam S07ASS1JQNP. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`

## Open Questions

- Is the GCP instance still running or has it been terminated?
- What domain will be used for the backend?
- What is the status of the AWS migration?

## Related Pages

- `aeneid-faucet-frontend`
- `story-aeneid-gcp`
- `story-testnet-aws`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_590b62743a010bfec71ede358314ee55`
