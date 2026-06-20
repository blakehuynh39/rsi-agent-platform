---
title: "SecurityBot"
type: "system"
slug: "systems/securitybot"
freshness: "2026-05-26T23:24:14Z"
tags:
  - "access-management"
  - "aws"
  - "self-service"
  - "slack-bot"
owners: []
source_revision_ids:
  - "srcrev_1a2d6eccf132a14d5f1abb94abd0e8e0"
  - "srcrev_99e59dc8da5943f276775595c9cc5c4d"
  - "srcrev_fd7857c6b9500bff1ea90746ef87af1d"
conflict_state: "none"
---

# SecurityBot

## Summary

Self-service Slack bot for AWS access management. Users request access directly; it cannot process requests on behalf of others.

## Claims

- SecurityBot is a self-service tool. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_fd7857c6b9500bff1ea90746ef87af1d` `chunk_id=srcchunk_139171127ff6fed70bdc6f73cd391de6` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813356.412119` `source_timestamp=2026-05-26T16:35:56Z`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_99e59dc8da5943f276775595c9cc5c4d` `chunk_id=srcchunk_8d695e01fa466f19737bc27cb8584fa8` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813365.387959` `source_timestamp=2026-05-26T16:36:05Z`
- SecurityBot cannot request permissions on someone else's request. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_fd7857c6b9500bff1ea90746ef87af1d` `chunk_id=srcchunk_139171127ff6fed70bdc6f73cd391de6` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813356.412119` `source_timestamp=2026-05-26T16:35:56Z`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_99e59dc8da5943f276775595c9cc5c4d` `chunk_id=srcchunk_8d695e01fa466f19737bc27cb8584fa8` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813365.387959` `source_timestamp=2026-05-26T16:36:05Z`
- Users should message SecurityBot directly with their access request, including instance IDs and account. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_1a2d6eccf132a14d5f1abb94abd0e8e0` `chunk_id=srcchunk_6c25c8a54d50382fd1a8b80c6b68659c` `native_locator=slack:C0547N89JUB:1779773836.978369:1779837854.575679` `source_timestamp=2026-05-26T23:24:14Z`

## Open Questions

- What commands does SecurityBot support beyond granting SSM access?

## Related Pages

- `aws-access-request-workflow`
- `aws-organization-repo`

## Sources

- `source_document_id`: `srcdoc_db42607450fc76635597bd303a2f8d66`
- `source_revision_id`: `srcrev_3daf831d4e686825dcb46552751901a3`
