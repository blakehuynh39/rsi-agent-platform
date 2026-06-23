---
title: "Cursor Bugbot Setup for subnet-worker-client"
type: "decision"
slug: "decisions/cursor-bugbot-setup-subnet-worker-client"
freshness: "2026-01-16T21:58:21Z"
tags:
  - "bugbot"
  - "cursor"
  - "github"
  - "pr-checks"
  - "subnet-worker-client"
owners:
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_1064fc54c3592b91aac2cbb86b410ac8"
  - "srcrev_1e9a120e857019a5225f13789a6c87e2"
  - "srcrev_86f476303e618481cd70772efd680712"
  - "srcrev_8ec46b66c3f0df4d7d9512917bf4ea0d"
  - "srcrev_92ba20a4b3e49e860b63f296c4dc51bc"
  - "srcrev_ed06fb3a5021d8a96a321abbd80a1a9e"
conflict_state: "none"
---

# Cursor Bugbot Setup for subnet-worker-client

## Summary

Decision and log of setting up Cursor Bugbot on the subnet-worker-client GitHub repository to run on pull requests. Required an org account, which was not available, so a GitHub admin used their personal account to configure it.

## Claims

- There was a request to set up Cursor Bugbot on the subnet-worker-client GitHub repository to run on pull requests. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_ed06fb3a5021d8a96a321abbd80a1a9e` `chunk_id=srcchunk_7d7d2d46bbed144b9efcf937f83ffd76` `native_locator=slack:C0547N89JUB:1768599963.404579:1768599963.404579` `source_timestamp=2026-01-16T21:46:03Z`
- Setting up Cursor Bugbot requires a Cursor organization account and GitHub admin permissions. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_1e9a120e857019a5225f13789a6c87e2` `chunk_id=srcchunk_9ef2443dff2e50e12c96f9b862f27432` `native_locator=slack:C0547N89JUB:1768599963.404579:1768599997.220969` `source_timestamp=2026-01-16T21:46:37Z`
- The team does not have a Cursor organization account. The only known way to obtain one is to close existing personal Cursor accounts and be added to a new org account. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_8ec46b66c3f0df4d7d9512917bf4ea0d` `chunk_id=srcchunk_22eb73f6494d33d39fff7ceac8f7b9e4` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600216.842919` `source_timestamp=2026-01-16T21:50:16Z`
- A GitHub org admin can run Cursor Bugbot through their own Cursor account because of their admin permissions, and it should work for all existing PRs. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_86f476303e618481cd70772efd680712` `chunk_id=srcchunk_6c90f6764d62699f475ec0cb9cb51b98` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600380.562429` `source_timestamp=2026-01-16T21:53:00Z`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_92ba20a4b3e49e860b63f296c4dc51bc` `chunk_id=srcchunk_8111fdc403905ae3bd7f09c87bdf6700` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600388.232739` `source_timestamp=2026-01-16T21:53:08Z`
- The setup was completed successfully and bugbot is ready for testing on existing PRs. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aabef130e41d5252b1e445f9054d3c9` `source_revision_id=srcrev_1064fc54c3592b91aac2cbb86b410ac8` `chunk_id=srcchunk_2b722b9e275e8885478a96994ebdd032` `native_locator=slack:C0547N89JUB:1768599963.404579:1768600701.752019` `source_timestamp=2026-01-16T21:58:21Z`

## Sources

- `source_document_id`: `srcdoc_8aabef130e41d5252b1e445f9054d3c9`
- `source_revision_id`: `srcrev_09862f619cd6faf25058d13062250e86`
