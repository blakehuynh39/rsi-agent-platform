---
title: "Mutability Support"
type: "concept"
slug: "concepts/mutability-support"
freshness: "2026-05-05T06:37:31Z"
tags:
  - "ip-protocol"
  - "license-terms"
  - "mutability"
owners: []
source_revision_ids:
  - "srcrev_fbe808064347ff47bf4bfe270f2d4ac8"
conflict_state: "none"
---

# Mutability Support

## Summary

Discusses requirements for mutability of license terms and IP at the protocol level, including the ability to unattach license terms when no license tokens are in circulation and no derivatives exist.

## Claims

- IP owners want the ability to mutate, enable, or disable license terms and IP. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4) `source_document_id=srcdoc_007d7bdaca31b733c679fca6dfe3d18a` `source_revision_id=srcrev_fbe808064347ff47bf4bfe270f2d4ac8` `chunk_id=srcchunk_f0e0461329e23778fd7f3029e42437ce` `native_locator=https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4` `source_timestamp=2026-05-05T06:37:31Z`
- Minting a license token currently makes the IP 'forever a root', preventing unattachment of license terms. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4) `source_document_id=srcdoc_007d7bdaca31b733c679fca6dfe3d18a` `source_revision_id=srcrev_fbe808064347ff47bf4bfe270f2d4ac8` `chunk_id=srcchunk_f0e0461329e23778fd7f3029e42437ce` `native_locator=https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4` `source_timestamp=2026-05-05T06:37:31Z`
- License terms should be unattachable only when there are zero license tokens in circulation and no derivatives exist. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4) `source_document_id=srcdoc_007d7bdaca31b733c679fca6dfe3d18a` `source_revision_id=srcrev_fbe808064347ff47bf4bfe270f2d4ac8` `chunk_id=srcchunk_f0e0461329e23778fd7f3029e42437ce` `native_locator=https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4` `source_timestamp=2026-05-05T06:37:31Z`
- A mechanism to check the balance of license tokens in circulation is needed to support unattachment. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4) `source_document_id=srcdoc_007d7bdaca31b733c679fca6dfe3d18a` `source_revision_id=srcrev_fbe808064347ff47bf4bfe270f2d4ac8` `chunk_id=srcchunk_f0e0461329e23778fd7f3029e42437ce` `native_locator=https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4` `source_timestamp=2026-05-05T06:37:31Z`

## Open Questions

- How to implement balanceOf for license tokens in circulation to enable safe unattachment of license terms?

## Sources

- `source_document_id`: `srcdoc_007d7bdaca31b733c679fca6dfe3d18a`
- `source_revision_id`: `srcrev_fbe808064347ff47bf4bfe270f2d4ac8`
- `source_url`: [Notion source](https://www.notion.so/Mutability-Support-818f282cddd74c13976c50ea5f3e80f4)
