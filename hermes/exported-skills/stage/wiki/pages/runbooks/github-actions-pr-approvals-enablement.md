---
title: "Enabling GitHub Actions Create and Approve Pull Requests"
type: "runbook"
slug: "runbooks/github-actions-pr-approvals-enablement"
freshness: "2026-05-14T07:02:06Z"
tags:
  - "cdr-sdk"
  - "enterprise-policy"
  - "github-actions"
  - "org-policy"
  - "pull-requests"
owners:
  - "@U07KLPN0JN6"
  - "@U07TNT9N4JC"
  - "@U0AKJV8710S"
source_revision_ids:
  - "srcrev_28267caaf1eb7d11d76a048c4a06214b"
  - "srcrev_81fbea510d7908f14d534e661968184e"
  - "srcrev_9b1a7bd248c5cdea581dbd2f59a67de9"
conflict_state: "none"
---

# Enabling GitHub Actions Create and Approve Pull Requests

## Summary

Steps to enable GitHub Actions to create and approve pull requests, required for automated release pipelines like changesets/action. Involves org-level and enterprise-level settings.

## Claims

- The org-level setting "Allow GitHub Actions to create and approve pull requests" must be enabled for GitHub Actions workflows to open pull requests in the piplabs org. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99fbf5a5c5a239c8638522c1cd592819` `source_revision_id=srcrev_81fbea510d7908f14d534e661968184e` `chunk_id=srcchunk_be425a0562a51100580f4bbef9249ff6` `native_locator=slack:C0547N89JUB:1778741500.186739:1778741500.186739` `source_timestamp=2026-05-14T06:51:49Z`
- Even with the org-level setting enabled, the enterprise policy (one tier above org) may still block GitHub Actions PR creation if not configured correctly. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99fbf5a5c5a239c8638522c1cd592819` `source_revision_id=srcrev_9b1a7bd248c5cdea581dbd2f59a67de9` `chunk_id=srcchunk_9545d217e5fde6bc03b7dd8dc5eb5ce1` `native_locator=slack:C0547N89JUB:1778741500.186739:1778741755.732529` `source_timestamp=2026-05-14T06:55:55Z`
- When enterprise policy blocks PR creation, the API returns a 409 Conflict: 'The enterprise does not allow GitHub Actions to create or approve pull requests' and the repo-level state stays false. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99fbf5a5c5a239c8638522c1cd592819` `source_revision_id=srcrev_9b1a7bd248c5cdea581dbd2f59a67de9` `chunk_id=srcchunk_9545d217e5fde6bc03b7dd8dc5eb5ce1` `native_locator=slack:C0547N89JUB:1778741500.186739:1778741755.732529` `source_timestamp=2026-05-14T06:55:55Z`
- After enterprise admin enabled the setting, the changesets pipeline worked successfully. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99fbf5a5c5a239c8638522c1cd592819` `source_revision_id=srcrev_28267caaf1eb7d11d76a048c4a06214b` `chunk_id=srcchunk_333f81f49abc6faa90a8bf0d0b5a14ae` `native_locator=slack:C0547N89JUB:1778741500.186739:1778742126.451039` `source_timestamp=2026-05-14T07:02:06Z`

## Sources

- `source_document_id`: `srcdoc_99fbf5a5c5a239c8638522c1cd592819`
- `source_revision_id`: `srcrev_28267caaf1eb7d11d76a048c4a06214b`
