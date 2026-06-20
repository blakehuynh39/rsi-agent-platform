---
title: "AWS Access Request Workflow"
type: "runbook"
slug: "runbooks/aws-access-request-workflow"
freshness: "2026-05-26T23:24:14Z"
tags:
  - "access-management"
  - "aws"
  - "self-service"
  - "ssm"
  - "terraform"
owners:
  - "IT Admin (Vinod)"
source_revision_ids:
  - "srcrev_1a2d6eccf132a14d5f1abb94abd0e8e0"
  - "srcrev_243a9f9953cf3802c744dc2d7e56ca67"
  - "srcrev_99e59dc8da5943f276775595c9cc5c4d"
  - "srcrev_d48e1f6f83cffba80fd7bad0fa48d8d6"
  - "srcrev_fd7857c6b9500bff1ea90746ef87af1d"
conflict_state: "none"
---

# AWS Access Request Workflow

## Summary

How to request AWS SSM access to EC2 instances, including self-service via SecurityBot and admin-level PRs in the aws-organization repo.

## Claims

- SecurityBot is a self-service tool; it cannot request permissions on behalf of someone else. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_fd7857c6b9500bff1ea90746ef87af1d` `chunk_id=srcchunk_139171127ff6fed70bdc6f73cd391de6` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813356.412119` `source_timestamp=2026-05-26T16:35:56Z`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_99e59dc8da5943f276775595c9cc5c4d` `chunk_id=srcchunk_8d695e01fa466f19737bc27cb8584fa8` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813365.387959` `source_timestamp=2026-05-26T16:36:05Z`
- Regular users cannot provision access for other employees; only admins can grant AWS access. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_243a9f9953cf3802c744dc2d7e56ca67` `chunk_id=srcchunk_c0df657873025044dcb8d89b045579c7` `native_locator=slack:C0547N89JUB:1779773836.978369:1779773845.954199` `source_timestamp=2026-05-26T05:37:25Z`
- Admins can open a PR in the aws-organization repo and comment `plan` and `apply` to provision AWS access via Terraform. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_fd7857c6b9500bff1ea90746ef87af1d` `chunk_id=srcchunk_139171127ff6fed70bdc6f73cd391de6` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813356.412119` `source_timestamp=2026-05-26T16:35:56Z`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_99e59dc8da5943f276775595c9cc5c4d` `chunk_id=srcchunk_8d695e01fa466f19737bc27cb8584fa8` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813365.387959` `source_timestamp=2026-05-26T16:36:05Z`
- To request AWS access, users should message SecurityBot directly with the specific instance IDs and account name. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_d48e1f6f83cffba80fd7bad0fa48d8d6` `chunk_id=srcchunk_1d5a442f5b3f59b1bc0b90565c26def8` `native_locator=slack:C0547N89JUB:1779773836.978369:1779837848.592949` `source_timestamp=2026-05-26T23:24:08Z`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_1a2d6eccf132a14d5f1abb94abd0e8e0` `chunk_id=srcchunk_6c25c8a54d50382fd1a8b80c6b68659c` `native_locator=slack:C0547N89JUB:1779773836.978369:1779837854.575679` `source_timestamp=2026-05-26T23:24:14Z`
- The IT admin is Vinod, and he is usually online starting 9 AM Los Angeles time. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_243a9f9953cf3802c744dc2d7e56ca67` `chunk_id=srcchunk_c0df657873025044dcb8d89b045579c7` `native_locator=slack:C0547N89JUB:1779773836.978369:1779773845.954199` `source_timestamp=2026-05-26T05:37:25Z`
  - citation: `source_document_id=srcdoc_db42607450fc76635597bd303a2f8d66` `source_revision_id=srcrev_fd7857c6b9500bff1ea90746ef87af1d` `chunk_id=srcchunk_139171127ff6fed70bdc6f73cd391de6` `native_locator=slack:C0547N89JUB:1779773836.978369:1779813356.412119` `source_timestamp=2026-05-26T16:35:56Z`

## Open Questions

- Can non-admin users request access for others through a delegated workflow?
- What is the process for obtaining admin-level permissions to grant AWS access to others?

## Related Pages

- `aws-organization-repo`
- `securitybot`

## Sources

- `source_document_id`: `srcdoc_db42607450fc76635597bd303a2f8d66`
- `source_revision_id`: `srcrev_3daf831d4e686825dcb46552751901a3`
