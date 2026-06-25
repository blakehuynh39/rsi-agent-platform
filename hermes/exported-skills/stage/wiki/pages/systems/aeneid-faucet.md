---
title: "Aeneid Faucet"
type: "system"
slug: "systems/aeneid-faucet"
freshness: "2026-02-13T08:29:16Z"
tags:
  - "aeneid"
  - "faucet"
  - "gcp"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_7093544ae96b0aedcf62b57610d01155"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# Aeneid Faucet

## Summary

Faucet service for distributing testnet tokens on Story Aeneid.

## Claims

- The faucet distributes testnet tokens to user wallets by validating frontend requests, storing wallet addresses in a database, and batch sending tokens. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- The faucet frontend URL is https://aeneid.faucet.story.foundation/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`
- The faucet backend server listens on IP 35.207.0.103 port 23313. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
- The backend runs on a GCP Compute Engine instance named use1-aeneid-fuacet in zone us-east1-c, project story-aeneid. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_7093544ae96b0aedcf62b57610d01155` `chunk_id=srcchunk_453f872fbaf84c6caa00f9f5e0e28d0e` `native_locator=slack:C0547N89JUB:1770083249.859729:1770083249.859729` `source_timestamp=2026-02-03T01:47:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`

## Open Questions

- What domain name will replace the IP for backend access after migration?

## Related Pages

- `migrate-aeneid-faucet-to-aws`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_bf67d2e4f7df453b626f423673a760aa`
