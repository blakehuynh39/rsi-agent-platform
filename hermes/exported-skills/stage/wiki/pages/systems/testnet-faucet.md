---
title: "Testnet Faucet"
type: "system"
slug: "systems/testnet-faucet"
freshness: "2026-02-19T16:52:07Z"
tags:
  - "faucet"
  - "ip-token"
  - "testnet"
  - "wip"
owners: []
source_revision_ids:
  - "srcrev_623e8a04e1e6e2db28e9ef34d0ab8748"
  - "srcrev_7e3af0270f6bd11877606640c6d0256f"
  - "srcrev_802b236027f62252e877bf36470a0e20"
  - "srcrev_99701306077a053546d42075d8a34b34"
conflict_state: "none"
---

# Testnet Faucet

## Summary

The Story Protocol testnet faucet provides native IP tokens but has had reliability issues. WIP tokens must be obtained via direct contract interaction.

## Claims

- The testnet faucet is located at https://faucet.story.foundation/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_7e3af0270f6bd11877606640c6d0256f` `chunk_id=srcchunk_ebf59bd8fad4d2f06a5835ad04250d71` `native_locator=slack:C04T5307FNU:1771451418.598879:1771451418.598879` `source_timestamp=2026-02-18T21:50:18Z`
- The faucet only dispenses native IP tokens, not WIP tokens. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- The $IP faucet was reported as not working with error 'Something went wrong, please try again.' `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_623e8a04e1e6e2db28e9ef34d0ab8748` `chunk_id=srcchunk_3c134e4f448fe4c1dbf3a3be98582785` `native_locator=slack:C04T5307FNU:1771502968.212109:1771502968.212109` `source_timestamp=2026-02-19T12:09:28Z`
- A possible cause for the faucet outage was an AWS migration. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_99701306077a053546d42075d8a34b34` `chunk_id=srcchunk_b82cd1cc13a4f7f052bd1de0d0595685` `native_locator=slack:C04T5307FNU:1771519893.118039:1771519893.118039` `source_timestamp=2026-02-19T16:52:07Z`

## Open Questions

- How can developers reliably obtain WIP tokens on testnet?

## Sources

- `source_document_id`: `srcdoc_0708298813917544f78c01e06c230684`
- `source_revision_id`: `srcrev_0fd5eef32db0608481dd1fbc064d18b9`
