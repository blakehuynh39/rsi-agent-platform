---
title: "Faucet Issues and Resolutions"
type: "runbook"
slug: "runbooks/faucet-issues"
freshness: "2026-02-19T16:52:07Z"
tags:
  - "faucet"
  - "ip-token"
  - "testnet"
  - "wip"
owners:
  - "U04L0DD6B6F"
  - "U07A7AUGL5V"
  - "U07C9478JUE"
source_revision_ids:
  - "srcrev_623e8a04e1e6e2db28e9ef34d0ab8748"
  - "srcrev_802b236027f62252e877bf36470a0e20"
  - "srcrev_99701306077a053546d42075d8a34b34"
conflict_state: "none"
---

# Faucet Issues and Resolutions

## Summary

Reports and troubleshooting of story.foundation faucet, including errors and WIP token acquisition.

## Claims

- The $IP faucet at https://faucet.story.foundation/ is not working, showing error 'Something went wrong, please try again.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_623e8a04e1e6e2db28e9ef34d0ab8748` `chunk_id=srcchunk_3c134e4f448fe4c1dbf3a3be98582785` `native_locator=slack:C04T5307FNU:1771502968.212109:1771502968.212109` `source_timestamp=2026-02-19T12:09:28Z`
- The faucet only provides native IP token; WIP tokens must be obtained via direct contract interaction. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- AWS migration may be impacting the faucet, but the backend should be running on both sides during migration. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_99701306077a053546d42075d8a34b34` `chunk_id=srcchunk_b82cd1cc13a4f7f052bd1de0d0595685` `native_locator=slack:C04T5307FNU:1771519893.118039:1771519893.118039` `source_timestamp=2026-02-19T16:52:07Z`

## Open Questions

- Is a WIP faucet planned?
- When will the faucet be fixed?

## Sources

- `source_document_id`: `srcdoc_0708298813917544f78c01e06c230684`
- `source_revision_id`: `srcrev_e72ddef389cad6ae7477c379b4c2e1b6`
