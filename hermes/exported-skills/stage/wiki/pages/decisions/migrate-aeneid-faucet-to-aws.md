---
title: "Migrate Aeneid Faucet Server from GCP to AWS"
type: "decision"
slug: "decisions/migrate-aeneid-faucet-to-aws"
freshness: "2026-02-14T03:37:31Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners:
  - "U0643B9PPC1"
  - "U07A7AUGL5V"
  - "U07TNT9N4JC"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_15036f8cdcd7730d9bf3ef252d63cd08"
  - "srcrev_4f7762d903550900445eec55650e231b"
  - "srcrev_5ac5d10a5bfbe6b45a289d398ba89e67"
  - "srcrev_95c59ecd8381343566eb68b4867da4e2"
  - "srcrev_b1862fd1ab8f88e3da964b7d58b94d39"
  - "srcrev_c2f9042cf09f729cfa67f2387c5d853a"
  - "srcrev_e8de05b1a44b7de06d1dddaf2dc8140f"
conflict_state: "none"
---

# Migrate Aeneid Faucet Server from GCP to AWS

## Summary

Decision to migrate the faucet backend server from GCP instance 'use1-aeneid-fuacet' to AWS, and to have the frontend point to a domain instead of an IP address to facilitate migration.

## Claims

- We want to migrate the server from GCP to AWS and have the frontend point to a domain instead of an IP address. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_15036f8cdcd7730d9bf3ef252d63cd08` `chunk_id=srcchunk_b8eb04c30ffaafd6bc68a3ae08061d6b` `native_locator=slack:C0547N89JUB:1770083249.859729:1770971534.533319` `source_timestamp=2026-02-13T08:32:14Z`
- Haodi (U0643B9PPC1) needs access to both the old GCP instance and the new AWS instance for migration. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_e8de05b1a44b7de06d1dddaf2dc8140f` `chunk_id=srcchunk_bd24a221873fb1e952e42e728631fb9f` `native_locator=slack:C0547N89JUB:1770083249.859729:1770975114.879239` `source_timestamp=2026-02-13T09:31:54Z`
- GCP access was granted to Haodi for project story-aeneid. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
- AWS access for Haodi is being worked on, requesting SSM access in story-testnet AWS account. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_95c59ecd8381343566eb68b4867da4e2` `chunk_id=srcchunk_5280ec09f50a147393aa122e7ffc5977` `native_locator=slack:C0547N89JUB:1770083249.859729:1770979821.418289` `source_timestamp=2026-02-13T10:50:21Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_4f7762d903550900445eec55650e231b` `chunk_id=srcchunk_52a56ac87d0fdd8ad3e0a4f9f3eea3dc` `native_locator=slack:C0547N89JUB:1770083249.859729:1771034156.878539` `source_timestamp=2026-02-14T01:55:56Z`
- Haodi should use email Haodi@piplabs.xyz for login, but that Google account seems banned; Slack account still active. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_c2f9042cf09f729cfa67f2387c5d853a` `chunk_id=srcchunk_a16b7e5dac6bfab93b978f4ef7c8d090` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039440.489309` `source_timestamp=2026-02-14T03:24:00Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_5ac5d10a5bfbe6b45a289d398ba89e67` `chunk_id=srcchunk_78a933dfbae53f7515d2f8b7e8b63c02` `native_locator=slack:C0547N89JUB:1770083249.859729:1771039649.305579` `source_timestamp=2026-02-14T03:27:29Z`
  - citation: `source_document_id=srcdoc_39d343ffa196af216bd4ea35c17da039` `source_revision_id=srcrev_b1862fd1ab8f88e3da964b7d58b94d39` `chunk_id=srcchunk_3f9f2b20e4bbf469f7cb397b114ddee9` `native_locator=slack:C0547N89JUB:1770083249.859729:1771040251.317959` `source_timestamp=2026-02-14T03:37:31Z`

## Open Questions

- Has the domain been set up for the frontend?
- Has the migration been completed?
- Is the AWS instance set up?
- Was the email issue resolved?

## Related Pages

- `aeneid-faucet-system`

## Sources

- `source_document_id`: `srcdoc_39d343ffa196af216bd4ea35c17da039`
- `source_revision_id`: `srcrev_38336933aedd8a0792bca9c5a79629fa`
