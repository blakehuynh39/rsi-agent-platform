---
title: "$IP Faucet Testnet Outage (April 2, 2025)"
type: "runbook"
slug: "runbooks/ip-faucet-testnet-outage-2025-04-02"
freshness: "2026-02-19T17:09:35Z"
tags:
  - "faucet"
  - "ip-token"
  - "outage"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_2944f5a509b5441dd14f905a84e9a253"
  - "srcrev_623e8a04e1e6e2db28e9ef34d0ab8748"
  - "srcrev_7e3af0270f6bd11877606640c6d0256f"
  - "srcrev_802b236027f62252e877bf36470a0e20"
  - "srcrev_99701306077a053546d42075d8a34b34"
  - "srcrev_ddd85f329245799e16b7f807b3dcb119"
conflict_state: "none"
---

# $IP Faucet Testnet Outage (April 2, 2025)

## Summary

On April 2, 2025, the Story Protocol testnet faucet at faucet.story.foundation returned errors, preventing users from claiming $IP tokens. The outage coincided with an AWS migration.

## Claims

- The $IP faucet at faucet.story.foundation was reported as not working, displaying 'Something went wrong, please try again.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_623e8a04e1e6e2db28e9ef34d0ab8748` `chunk_id=srcchunk_3c134e4f448fe4c1dbf3a3be98582785` `native_locator=slack:C04T5307FNU:1771502968.212109:1771502968.212109` `source_timestamp=2026-02-19T12:09:28Z`
- The faucet URL is https://faucet.story.foundation/ `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_7e3af0270f6bd11877606640c6d0256f` `chunk_id=srcchunk_ebf59bd8fad4d2f06a5835ad04250d71` `native_locator=slack:C04T5307FNU:1771451418.598879:1771451418.598879` `source_timestamp=2026-02-18T21:50:18Z`
- The faucet is only for native $IP token; WIP must be obtained via direct contract interaction. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- An AWS migration was ongoing at the time of the outage, and the faucet backend was expected to run on both sides during migration. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_99701306077a053546d42075d8a34b34` `chunk_id=srcchunk_b82cd1cc13a4f7f052bd1de0d0595685` `native_locator=slack:C04T5307FNU:1771519893.118039:1771519893.118039` `source_timestamp=2026-02-19T16:52:07Z`
- A team member was asked to investigate and expected to check after reaching their computer. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_ddd85f329245799e16b7f807b3dcb119` `chunk_id=srcchunk_6e6f662a2a371c193786e5573c4300df` `native_locator=slack:C04T5307FNU:1771519401.982239:1771519401.982239` `source_timestamp=2026-02-19T16:43:21Z`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_99701306077a053546d42075d8a34b34` `chunk_id=srcchunk_b82cd1cc13a4f7f052bd1de0d0595685` `native_locator=slack:C04T5307FNU:1771519893.118039:1771519893.118039` `source_timestamp=2026-02-19T16:52:07Z`
- Another user reported the faucet was not working, but they had tokens and needed WIP. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_2944f5a509b5441dd14f905a84e9a253` `chunk_id=srcchunk_31b56db02f480700b21f08a6bb9b25e6` `native_locator=slack:C04T5307FNU:1771520975.894629:1771520975.894629` `source_timestamp=2026-02-19T17:09:35Z`

## Open Questions

- Root cause of the faucet outage?
- Was the outage resolved? If so, what was the fix?

## Sources

- `source_document_id`: `srcdoc_0708298813917544f78c01e06c230684`
- `source_revision_id`: `srcrev_391babc60fc61846362540e66d5d76a5`
