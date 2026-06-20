---
title: "Meng Agent Outage May 2026"
type: "runbook"
slug: "runbooks/meng-agent-outage-2026-05-08"
freshness: "2026-05-08T04:58:10Z"
tags:
  - "incident"
  - "meng-agent"
owners: []
source_revision_ids:
  - "srcrev_878594d6a41823097252a9c7982424a9"
  - "srcrev_a3df4db10a197d33a7a1dbc9cd5c2a27"
  - "srcrev_ebb2b2d14f2f4fd65d52c80a31c6b687"
conflict_state: "none"
---

# Meng Agent Outage May 2026

## Summary

On 2026-05-08, Meng agent was reported down with a network connection error. Investigation revealed the runtime host was stopped after migration. The issue was resolved by the reporting user.

## Claims

- User reported Meng agent down with error 'LLM request failed: network connection error.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_da11634ea7b7e1497d3659f78c5ee5f5` `source_revision_id=srcrev_878594d6a41823097252a9c7982424a9` `chunk_id=srcchunk_ea7657059c21ae745b60439c397e09c4` `native_locator=slack:C0547N89JUB:1778213410.573869:1778213410.573869` `source_timestamp=2026-05-08T04:10:10Z`
- Investigation found Meng agent's runtime host stopped since Apr 23 06:52 UTC after migration, target no longer exists. Internal LiteLLM endpoint reachable; network error from broken runtime path, not global outage. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_da11634ea7b7e1497d3659f78c5ee5f5` `source_revision_id=srcrev_ebb2b2d14f2f4fd65d52c80a31c6b687` `chunk_id=srcchunk_6d99417d8e6188beb2e23e715b430682` `native_locator=slack:C0547N89JUB:1778213410.573869:1778213557.656509` `source_timestamp=2026-05-08T04:12:37Z`
- The user resolved the issue themselves after confirmation. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_da11634ea7b7e1497d3659f78c5ee5f5` `source_revision_id=srcrev_a3df4db10a197d33a7a1dbc9cd5c2a27` `chunk_id=srcchunk_76f7945e1d0f88c512ad0eea936110bd` `native_locator=slack:C0547N89JUB:1778213410.573869:1778216290.618329` `source_timestamp=2026-05-08T04:58:10Z`

## Sources

- `source_document_id`: `srcdoc_da11634ea7b7e1497d3659f78c5ee5f5`
- `source_revision_id`: `srcrev_a3df4db10a197d33a7a1dbc9cd5c2a27`
