---
title: "Aeneid Faucet"
type: "system"
slug: "systems/aeneid-faucet"
freshness: "2026-02-14T01:55:56Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
  - "testnet"
owners:
  - "S07ASS1JQNP"
  - "U063MQ9ADCM"
  - "U0643B9PPC1"
source_revision_ids:
  - "srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7"
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_43c0946f3f4587e464c535b2f02e284a"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_50b60006f4eeb38f39acd80cb65b8cff"
conflict_state: "none"
---

# Aeneid Faucet

## Summary

The Aeneid testnet faucet distributes testnet tokens to users. It is currently hosted on a GCP instance and accessed via IP. We are migrating it to AWS and will switch to a domain name.

## Claims

- The faucet backend is hosted on the GCP instance 'use1-aeneid-fuacet' in project story-aeneid, with IP address 35.207.0.103. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_50b60006f4eeb38f39acd80cb65b8cff` `chunk_id=srcchunk_bf85328875d49cd44ea72f1a0ed48af2` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971356.987109` `source_timestamp=2026-02-13T08:29:16Z`
- The frontend queries the faucet backend at http://35.207.0.103:23313/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_13c0ccf5c4b68c23b101ff5abf0ad3c7` `chunk_id=srcchunk_b80ef615f7867327a434458e8ef1bd61` `native_locator=slack:C0547N89JUB:1770083249.859729:1770106765.126079` `source_timestamp=2026-02-03T08:19:25Z`
- The backend validates the frontend request, stores wallet addresses in a database table, and batch-sends testnet tokens chronologically. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_43c0946f3f4587e464c535b2f02e284a` `chunk_id=srcchunk_42b6f69cb1c483b952a70161fab8aedb` `native_locator=slack:C0547N89JUB:1770083249.859729:1770969770.716529` `source_timestamp=2026-02-13T08:03:17Z`
- We plan to migrate the faucet server from GCP to AWS and switch the frontend to use a domain name instead of the IP address. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- GCP access to project story-aeneid has been granted to Haodi (U0643B9PPC1) for migration. AWS access to account story-testnet is being set up. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`

## Open Questions

- Are there any database migration steps needed?
- Is the AWS instance fully provisioned and configured?
- What domain name will be used for the faucet backend?
- What is the timeline for completing the migration?
- Who will update the frontend code to use the new domain?

## Related Pages

- `story-aeneid`
- `story-testnet`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_7093544ae96b0aedcf62b57610d01155`
