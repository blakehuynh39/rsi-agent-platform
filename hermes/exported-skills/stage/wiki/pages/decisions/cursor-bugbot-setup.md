---
title: "Cursor Bugbot Setup for subnet-worker-client"
type: "decision"
slug: "decisions/cursor-bugbot-setup"
freshness: "2026-01-16T21:58:21Z"
tags:
  - "bugbot"
  - "ci"
  - "cursor"
  - "github"
  - "pull-requests"
owners: []
source_revision_ids:
  - "srcrev_1064fc54c3592b91aac2cbb86b410ac8"
  - "srcrev_1e9a120e857019a5225f13789a6c87e2"
  - "srcrev_86f476303e618481cd70772efd680712"
  - "srcrev_8ec46b66c3f0df4d7d9512917bf4ea0d"
  - "srcrev_92ba20a4b3e49e860b63f296c4dc51bc"
  - "srcrev_b0366eeccc9815a09de6eecd047352ff"
  - "srcrev_ed06fb3a5021d8a96a321abbd80a1a9e"
conflict_state: "disputed"
---

# Cursor Bugbot Setup for subnet-worker-client

## Summary

Process and decisions around enabling Cursor bugbot on GitHub PRs for the subnet-worker-client repo, addressing the requirement for a Cursor org account and workarounds via GitHub org admin permissions.

## Claims

- A request was made to set up cursor bugbot to run on PRs for the subnet-worker-client repo. `claim:request` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_ed06fb3a5021d8a96a321abbd80a1a9e` `chunk_id=srcchunk_7d7d2d46bbed144b9efcf937f83ffd76` `native_locator=slack:C0547N89JUB:1768599963.404579:1768599963.404579` `source_timestamp=2026-01-16T21:46:03Z`
- Setting up cursor bugbot requires a cursor org account. `claim:requirement_org_account` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_1e9a120e857019a5225f13789a6c87e2` `chunk_id=srcchunk_9ef2443dff2e50e12c96f9b862f27432` `native_locator=slack:C0547N89JUB:1768599963.404579:1768599997.220969` `source_timestamp=2026-01-16T21:46:37Z`
- The team does not have a cursor org account; the only way to get an org account is to close existing cursor accounts and join a new org account. `claim:no_org_account` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_8ec46b66c3f0df4d7d9512917bf4ea0d` `chunk_id=srcchunk_22eb73f6494d33d39fff7ceac8f7b9e4` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600216.842919` `source_timestamp=2026-01-16T21:50:16Z`
- A GitHub org admin may be able to run bugbot without an org account. `claim:admin_may_work` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_86f476303e618481cd70772efd680712` `chunk_id=srcchunk_6c90f6764d62699f475ec0cb9cb51b98` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600380.562429` `source_timestamp=2026-01-16T21:53:00Z`
- Bugbot should work for all existing PRs. `claim:works_existing_prs` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_92ba20a4b3e49e860b63f296c4dc51bc` `chunk_id=srcchunk_8111fdc403905ae3bd7f09c87bdf6700` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600388.232739` `source_timestamp=2026-01-16T21:53:08Z`
- A team member attempted the setup. `claim:attempt` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_b0366eeccc9815a09de6eecd047352ff` `chunk_id=srcchunk_8b5259084aa210b10b453f65f243d88e` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600465.085229` `source_timestamp=2026-01-16T21:54:25Z`
- The cursor bugbot setup was completed and indicated as done. `claim:setup_complete` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_1064fc54c3592b91aac2cbb86b410ac8` `chunk_id=srcchunk_2b722b9e275e8885478a96994ebdd032` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600701.752019` `source_timestamp=2026-01-16T21:58:21Z`

## Conflicts

- The claim that a cursor org account is required is disputed by later statements suggesting an admin can bypass it and by the successful completion of the setup. `claim:requirement_org_account`
  - conflict citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_1e9a120e857019a5225f13789a6c87e2` `chunk_id=srcchunk_9ef2443dff2e50e12c96f9b862f27432` `native_locator=slack:C0547N89JUB:1768599963.404579:1768599997.220969` `source_timestamp=2026-01-16T21:46:37Z`

## Open Questions

- Are there any limitations running bugbot without an org account?
- Does the bugbot actually function correctly on PRs?
- Is an org account ultimately necessary for full bugbot functionality?

## Sources

- `source_document_id`: `srcdoc_8aabef130e41d5252b1e445f9054d3c9`
- `source_revision_id`: `srcrev_1e9a120e857019a5225f13789a6c87e2`
