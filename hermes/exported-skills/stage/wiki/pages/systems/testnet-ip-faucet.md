---
title: "Testnet IP Faucet"
type: "system"
slug: "systems/testnet-ip-faucet"
freshness: "2026-02-19T16:52:07Z"
tags:
  - "faucet"
  - "ip-token"
  - "testnet"
  - "wip"
owners: []
source_revision_ids:
  - "srcrev_0fd5eef32db0608481dd1fbc064d18b9"
  - "srcrev_623e8a04e1e6e2db28e9ef34d0ab8748"
  - "srcrev_7e3af0270f6bd11877606640c6d0256f"
  - "srcrev_802b236027f62252e877bf36470a0e20"
  - "srcrev_99701306077a053546d42075d8a34b34"
conflict_state: "none"
---

# Testnet IP Faucet

## Summary

Information about the Story Protocol testnet faucet for IP and WIP tokens, including known issues.

## Claims

- The Story Protocol testnet faucet is located at https://faucet.story.foundation/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_7e3af0270f6bd11877606640c6d0256f` `chunk_id=srcchunk_ebf59bd8fad4d2f06a5835ad04250d71` `native_locator=slack:C04T5307FNU:1771451418.598879:1771451418.598879` `source_timestamp=2026-02-18T21:50:18Z`
- The faucet is only for native IP tokens, not WIP tokens. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- To obtain WIP tokens on testnet, users must interact directly with the contract. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- As of 2026-02-19, the $IP faucet was reported not working, displaying error 'Something went wrong, please try again.' `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_623e8a04e1e6e2db28e9ef34d0ab8748` `chunk_id=srcchunk_3c134e4f448fe4c1dbf3a3be98582785` `native_locator=slack:C04T5307FNU:1771502968.212109:1771502968.212109` `source_timestamp=2026-02-19T12:09:28Z`
- An AWS migration was ongoing which could affect the faucet backend, though it should be running on both sides during migration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_99701306077a053546d42075d8a34b34` `chunk_id=srcchunk_b82cd1cc13a4f7f052bd1de0d0595685` `native_locator=slack:C04T5307FNU:1771519893.118039:1771519893.118039` `source_timestamp=2026-02-19T16:52:07Z`
- Developers needed WIP tokens for license flow testing during a buildathon. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_0fd5eef32db0608481dd1fbc064d18b9` `chunk_id=srcchunk_44ada0b99c936833a4bf2be2baa22002` `native_locator=slack:C04T5307FNU:1771436612.421279:1771436612.421279` `source_timestamp=2026-02-18T17:43:32Z`

## Open Questions

- Is the faucet now functional?
- What is the recommended contract interaction to obtain WIP?

## Sources

- `source_document_id`: `srcdoc_0708298813917544f78c01e06c230684`
- `source_revision_id`: `srcrev_9a523de58f35281175c64f7b2c1044d7`
