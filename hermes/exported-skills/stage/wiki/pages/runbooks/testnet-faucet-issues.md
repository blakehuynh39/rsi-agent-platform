---
title: "Testnet Faucet Issues"
type: "runbook"
slug: "runbooks/testnet-faucet-issues"
freshness: "2026-04-13T20:41:58Z"
tags:
  - "faucet"
  - "ip-token"
  - "testnet"
  - "wip"
owners:
  - "U04L0DD6B6F"
  - "U07A7AUGL5V"
  - "U07KLPN0JN6"
source_revision_ids:
  - "srcrev_2944f5a509b5441dd14f905a84e9a253"
  - "srcrev_623e8a04e1e6e2db28e9ef34d0ab8748"
  - "srcrev_7e3af0270f6bd11877606640c6d0256f"
  - "srcrev_802b236027f62252e877bf36470a0e20"
  - "srcrev_8d239a9e28b207421c55cf104c920df6"
  - "srcrev_99701306077a053546d42075d8a34b34"
  - "srcrev_ddd85f329245799e16b7f807b3dcb119"
conflict_state: "none"
---

# Testnet Faucet Issues

## Summary

The testnet faucet (faucet.story.foundation) provides native IP tokens, but was observed to be non-functional around April 2025. WIP tokens must be obtained via direct contract interaction.

## Claims

- The Story testnet faucet is located at https://faucet.story.foundation/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_7e3af0270f6bd11877606640c6d0256f` `chunk_id=srcchunk_ebf59bd8fad4d2f06a5835ad04250d71` `native_locator=slack:C04T5307FNU:1771451418.598879:1771451418.598879` `source_timestamp=2026-02-18T21:50:18Z`
- The faucet only distributes native IP tokens, not WIP tokens. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- WIP tokens on testnet can be obtained by directly interacting with the token contract. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_802b236027f62252e877bf36470a0e20` `chunk_id=srcchunk_0c44425bac1be371afb9b3aefa95fbdd` `native_locator=slack:C04T5307FNU:1771492328.823799:1771492328.823799` `source_timestamp=2026-02-19T09:12:08Z`
- On or around 2025-04-22, the faucet was not working, displaying error 'Something went wrong, please try again.' `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_623e8a04e1e6e2db28e9ef34d0ab8748` `chunk_id=srcchunk_3c134e4f448fe4c1dbf3a3be98582785` `native_locator=slack:C04T5307FNU:1771502968.212109:1771502968.212109` `source_timestamp=2026-02-19T12:09:28Z`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_2944f5a509b5441dd14f905a84e9a253` `chunk_id=srcchunk_31b56db02f480700b21f08a6bb9b25e6` `native_locator=slack:C04T5307FNU:1771520975.894629:1771520975.894629` `source_timestamp=2026-02-19T17:09:35Z`
- The faucet outage may be related to AWS migration in progress. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_99701306077a053546d42075d8a34b34` `chunk_id=srcchunk_b82cd1cc13a4f7f052bd1de0d0595685` `native_locator=slack:C04T5307FNU:1771519893.118039:1771519893.118039` `source_timestamp=2026-02-19T16:52:07Z`
- @U04L0DD6B6F and @U07A7AUGL5V were requested to investigate the faucet issue. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_ddd85f329245799e16b7f807b3dcb119` `chunk_id=srcchunk_6e6f662a2a371c193786e5573c4300df` `native_locator=slack:C04T5307FNU:1771519401.982239:1771519401.982239` `source_timestamp=2026-02-19T16:43:21Z`
- @U07KLPN0JN6 was asked to help. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0708298813917544f78c01e06c230684` `source_revision_id=srcrev_8d239a9e28b207421c55cf104c920df6` `chunk_id=srcchunk_f3439631f84656d17d046aaec6b60912` `native_locator=slack:C04T5307FNU:1776112918.671319:1776112918.671319` `source_timestamp=2026-04-13T20:41:58Z`

## Open Questions

- What caused the faucet to fail (AWS migration or other)?

## Sources

- `source_document_id`: `srcdoc_0708298813917544f78c01e06c230684`
- `source_revision_id`: `srcrev_74d2e561f21f27a4de4225811188f1e8`
