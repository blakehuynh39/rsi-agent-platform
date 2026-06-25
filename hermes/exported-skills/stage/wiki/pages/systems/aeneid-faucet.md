---
title: "Aeneid Faucet"
type: "system"
slug: "systems/aeneid-faucet"
freshness: "2026-02-13T08:29:16Z"
tags:
  - "aeneid"
  - "faucet"
  - "testnet"
owners:
  - "S07ASS1JQNP"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# Aeneid Faucet

## Summary

The Aeneid testnet faucet provides testnet tokens via a web interface at https://aeneid.faucet.story.foundation/. The backend runs on a GCP instance with IP 35.207.0.103, validates requests, and batch-transfers tokens.

## Claims

- The faucet web interface is at https://aeneid.faucet.story.foundation/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`
- The faucet frontend queries the backend at http://35.207.0.103:23313/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
- The backend server is a GCP instance named use1-aeneid-fuacet in project story-aeneid, with IP 35.207.0.103. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The backend validates frontend requests, adds wallet addresses to a database table, and batch sends testnet tokens chronologically. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- The backend currently uses the IP address directly without a domain name. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- The faucet was developed by subteam S07ASS1JQNP. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`

## Open Questions

- Is the backend server still actively used without a domain?
- Will the backend be migrated to AWS as planned?

## Related Pages

- `faucet-gcp-to-aws-migration`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_50b60006f4eeb38f39acd80cb65b8cff`
