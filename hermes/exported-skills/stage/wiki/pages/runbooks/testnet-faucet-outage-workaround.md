---
title: "Testnet Faucet Outage and Workaround"
type: "runbook"
slug: "runbooks/testnet-faucet-outage-workaround"
freshness: "2026-02-19T16:52:07Z"
tags:
  - "faucet"
  - "incident"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_623e8a04e1e6e2db28e9ef34d0ab8748"
  - "srcrev_802b236027f62252e877bf36470a0e20"
  - "srcrev_99701306077a053546d42075d8a34b34"
  - "srcrev_ddd85f329245799e16b7f807b3dcb119"
conflict_state: "none"
---

# Testnet Faucet Outage and Workaround

## Summary

Incident where Story Protocol testnet IP faucet was down; workaround to obtain WIP tokens via direct contract interaction.

## Claims

- The Story Protocol IP token faucet (https://faucet.story.foundation/) was not working, showing error 'Something went wrong, please try again.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_623e8a04e1e6e2db28e9ef34d0ab8748` `chunk_id=srcchunk_3c134e4f448fe4c1dbf3a3be98582785` `native_locator=slack:C04T5307FNU:1771502968.212109:1771502968.212109` `source_timestamp=2026-02-19T12:09:28Z`
- The faucet only provides native IP tokens; to get WIP tokens, one must interact directly with the contract. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- A developer mentioned that there is an ongoing AWS migration, but the faucet backend should be running on both sides during migration. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_99701306077a053546d42075d8a34b34` `chunk_id=srcchunk_b82cd1cc13a4f7f052bd1de0d0595685` `native_locator=slack:C04T5307FNU:1771519893.118039:1771519893.118039` `source_timestamp=2026-02-19T16:52:07Z`
- Team members @U04L0DD6B6F (Chris) and @U07A7AUGL5V were asked to investigate the faucet issue. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_ddd85f329245799e16b7f807b3dcb119` `chunk_id=srcchunk_6e6f662a2a371c193786e5573c4300df` `native_locator=slack:C04T5307FNU:1771519401.982239:1771519401.982239` `source_timestamp=2026-02-19T16:43:21Z`

## Sources

- `source_document_id`: `srcdoc_0708298813917544f78c01e06c230684`
- `source_revision_id`: `srcrev_deedcaf6a3f6dc95a990dec7a66135d3`
