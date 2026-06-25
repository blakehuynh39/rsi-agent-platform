---
title: "Consensus Errors on Aeneid Validator Nodes - 2026-02-02"
type: "concept"
slug: "concepts/incident-consensus-errors-aeneid-2026-02-02"
freshness: "2026-02-02T07:19:33Z"
tags:
  - "aeneid"
  - "consensus"
  - "validator"
owners: []
source_revision_ids:
  - "srcrev_e2a0bf66133695de4eb43b5201e3facb"
  - "srcrev_fa4c1d44e69e0258baa27997589f5c94"
conflict_state: "none"
---

# Consensus Errors on Aeneid Validator Nodes - 2026-02-02

## Summary

On February 2, 2026, Aeneid validator nodes experienced consensus errors including conflicting votes and invalid proposal signatures, possibly due to an unsafe_reset operation.

## Claims

- Validator1 logged conflicting votes from validator 0FC41199CE588948861A8DA86D725A5A073AE91A at height 13965696, round 4, at 07:14:25 UTC. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41ff534936febc6826bafb024edcd6de` `source_revision_id=srcrev_e2a0bf66133695de4eb43b5201e3facb` `chunk_id=srcchunk_7bd85cfa92037615d45d02a12595db30` `native_locator=slack:C0547N89JUB:1770016517.299769:1770016517.299769` `source_timestamp=2026-02-02T07:19:33Z`
- Validator1 later logged a conflicting vote from itself and asked 'did you unsafe_reset a validator?' at height 13965696, round 8, at 07:14:41 UTC. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41ff534936febc6826bafb024edcd6de` `source_revision_id=srcrev_e2a0bf66133695de4eb43b5201e3facb` `chunk_id=srcchunk_7bd85cfa92037615d45d02a12595db30` `native_locator=slack:C0547N89JUB:1770016517.299769:1770016517.299769` `source_timestamp=2026-02-02T07:19:33Z`
- Validator3 logged multiple 'invalid proposal signature' errors from various peers at height 13965721, round 0, at 07:17:45 UTC. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41ff534936febc6826bafb024edcd6de` `source_revision_id=srcrev_fa4c1d44e69e0258baa27997589f5c94` `chunk_id=srcchunk_a6d0e7df18ff853986160f4a2dd5defe` `native_locator=slack:C0547N89JUB:1770016517.299769:1770016689.624069` `source_timestamp=2026-02-02T07:18:09Z`

## Open Questions

- Are the invalid proposal signatures related to the same validator issue?
- Did an unsafe_reset trigger the conflicting votes?
- What caused the conflicting votes?
- What is the validator identity 0FC41199CE588948861A8DA86D725A5A073AE91A?

## Sources

- `source_document_id`: `srcdoc_41ff534936febc6826bafb024edcd6de`
- `source_revision_id`: `srcrev_e2a0bf66133695de4eb43b5201e3facb`
