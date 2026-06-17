---
title: "Story Protocol Faucet and Testnet Token Acquisition"
type: "runbook"
slug: "runbooks/faucet-and-testnet-tokens"
freshness: "2026-02-19T16:52:07Z"
tags:
  - "faucet"
  - "ip-token"
  - "testnet"
  - "troubleshooting"
  - "wip"
owners: []
source_revision_ids:
  - "srcrev_623e8a04e1e6e2db28e9ef34d0ab8748"
  - "srcrev_802b236027f62252e877bf36470a0e20"
  - "srcrev_99701306077a053546d42075d8a34b34"
  - "srcrev_ddd85f329245799e16b7f807b3dcb119"
conflict_state: "none"
---

# Story Protocol Faucet and Testnet Token Acquisition

## Summary

Overview of the Story Protocol faucet issues, its functionality, and alternative methods to obtain WIP tokens on testnet.

## Claims

- The IP faucet is not working, displaying 'Something went wrong, please try again.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_623e8a04e1e6e2db28e9ef34d0ab8748` `chunk_id=srcchunk_3c134e4f448fe4c1dbf3a3be98582785` `native_locator=slack:C04T5307FNU:1771502968.212109:1771502968.212109` `source_timestamp=2026-02-19T12:09:28Z`
- The faucet only dispenses native IP token, not WIP. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- To obtain WIP token, direct contract interaction is required. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- The faucet issue was acknowledged and will be looked into by team members. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_ddd85f329245799e16b7f807b3dcb119` `chunk_id=srcchunk_6e6f662a2a371c193786e5573c4300df` `native_locator=slack:C04T5307FNU:1771519401.982239:1771519401.982239` `source_timestamp=2026-02-19T16:43:21Z`
- AWS migration may have affected the faucet, but the backend should be running on both sides during migration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_99701306077a053546d42075d8a34b34` `chunk_id=srcchunk_b82cd1cc13a4f7f052bd1de0d0595685` `native_locator=slack:C04T5307FNU:1771519893.118039:1771519893.118039` `source_timestamp=2026-02-19T16:52:07Z`

## Open Questions

- Is there a documented method for direct WIP token acquisition?
- What is the current status of the IP faucet?

## Sources

- `source_document_id`: `srcdoc_0708298813917544f78c01e06c230684`
- `source_revision_id`: `srcrev_8d239a9e28b207421c55cf104c920df6`
