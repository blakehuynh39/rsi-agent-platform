---
title: "Aeneid Faucet Server"
type: "system"
slug: "systems/aeneid-faucet-server"
freshness: "2026-02-14T01:55:56Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "infrastructure"
  - "testnet"
owners:
  - "@U063MQ9ADCM"
  - "@U0643B9PPC1"
source_revision_ids:
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
conflict_state: "none"
---

# Aeneid Faucet Server

## Summary

Backend server for the Aeneid testnet faucet, currently hosted on GCP instance use1-aeneid-fuacet, serving IP 35.207.0.103:23313. Planned migration to AWS.

## Claims

- The faucet frontend sends requests to http://35.207.0.103:23313/ (the IP of GCP instance use1-aeneid-fuacet). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The backend validates frontend requests, stores wallet addresses in a database table, and batch-sends testnet tokens to users chronologically. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- There is a plan to migrate the server from GCP to AWS, which will require the frontend to point to a domain instead of an IP address. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- GCP access was granted to @U0643B9PPC1, and AWS access was being set up to assist with migration or documentation. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
- The faucet server is currently active (in use). `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`

## Open Questions

- Has the migration to AWS been completed?
- What domain will the frontend point to after migration?

## Related Pages

- `aeneid-faucet-frontend`
- `story-aeneid-gcp-project`
- `story-testnet-aws-account`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_35436297c2b3747d63880c7842c5fdd0`
