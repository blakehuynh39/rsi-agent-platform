---
title: "Aeneid Faucet System"
type: "system"
slug: "systems/aeneid-faucet-system"
freshness: "2026-02-13T10:05:36Z"
tags:
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
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
  - "srcrev_7093544ae96b0aedcf62b57610d01155"
  - "srcrev_82fd93b3502f740baf241e47f0ce88d5"
  - "srcrev_dc87e72b13ffb646b259a7fc91c395d8"
conflict_state: "none"
---

# Aeneid Faucet System

## Summary

The Aeneid faucet is a testnet token distribution system hosted on a GCP VM instance (use1-aeneid-fuacet) with IP 35.207.0.103. The frontend (https://aeneid.faucet.story.foundation/) queries the backend via that IP. A migration to AWS is planned, requiring a domain-based address to replace the hardcoded IP.

## Claims

- The GCP instance name is use1-aeneid-fuacet, located in zone us-east1-c, project story-aeneid. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_7093544ae96b0aedcf62b57610d01155` `chunk_id=srcchunk_453f872fbaf84c6caa00f9f5e0e28d0e` `native_locator=slack:C0547N89JUB:1770083249.859729:1770083249.859729` `source_timestamp=2026-02-03T01:47:29Z`
- The faucet frontend is accessible at https://aeneid.faucet.story.foundation/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_dc87e72b13ffb646b259a7fc91c395d8` `chunk_id=srcchunk_3f7de3056380c3b0a128e380c3c3f79f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770085370.453109` `source_timestamp=2026-02-03T02:22:50Z`
- The frontend queries the backend at http://35.207.0.103:23313/. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
- The backend validates frontend requests, adds wallet addresses to a database table, and chronologically batch-sends testnet tokens to users. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- The IP 35.207.0.103 is confirmed to belong to the GCP instance use1-aeneid-fuacet, and the frontend actively queries it. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The team intends to migrate the server from GCP to AWS and change the frontend to use a domain name instead of the hardcoded IP. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- Access was requested for Haodi (U0643B9PPC1) to GCP project story-aeneid and AWS account story-testnet for the migration. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_82fd93b3502f740baf241e47f0ce88d5` `chunk_id=srcchunk_86cbb175607d8bad9b3a03bbe9422c5d` `native_locator=slack:C0547N89JUB:1770083249.859729:1770977136.045799` `source_timestamp=2026-02-13T10:05:36Z`

## Open Questions

- Has the migration been completed?
- Is the Piplabs email suspension resolved for Haodi's access?
- What domain will replace the IP?

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_75cb47489ae39a7dae7ff85bc62c8f91`
