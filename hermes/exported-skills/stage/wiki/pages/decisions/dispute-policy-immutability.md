---
title: "Dispute Policy Immutability"
type: "decision"
slug: "decisions/dispute-policy-immutability"
freshness: "2026-05-05T06:36:36Z"
tags:
  - "dispute-policy"
  - "immutability"
  - "p1"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_aba2bccd4890de44c51dfab17b2c3ca1"
conflict_state: "none"
---

# Dispute Policy Immutability

## Summary

Dispute policies should be made immutable to prevent changes that could rug derivatives, especially with IPA staking where staked capital can be directly slashed.

## Claims

- Dispute policies should be immutable to prevent rugging derivatives. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980) `source_document_id=srcdoc_e6330768515dc61898e883dd66fdd63f` `source_revision_id=srcrev_aba2bccd4890de44c51dfab17b2c3ca1` `chunk_id=srcchunk_2f37f255cbff962b81c266b55d826f95` `native_locator=https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980` `source_timestamp=2026-05-05T06:36:36Z`
- If a parent IP is tagged, its derivatives are also tagged, so an immutable dispute policy reduces risk. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980) `source_document_id=srcdoc_e6330768515dc61898e883dd66fdd63f` `source_revision_id=srcrev_aba2bccd4890de44c51dfab17b2c3ca1` `chunk_id=srcchunk_2f37f255cbff962b81c266b55d826f95` `native_locator=https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980` `source_timestamp=2026-05-05T06:36:36Z`
- Immutability is especially important when IPA staking is added because there is direct slashing of staked capital. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980) `source_document_id=srcdoc_e6330768515dc61898e883dd66fdd63f` `source_revision_id=srcrev_aba2bccd4890de44c51dfab17b2c3ca1` `chunk_id=srcchunk_2f37f255cbff962b81c266b55d826f95` `native_locator=https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980` `source_timestamp=2026-05-05T06:36:36Z`
- This is a P1 Small priority item. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980) `source_document_id=srcdoc_e6330768515dc61898e883dd66fdd63f` `source_revision_id=srcrev_aba2bccd4890de44c51dfab17b2c3ca1` `chunk_id=srcchunk_2f37f255cbff962b81c266b55d826f95` `native_locator=https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980` `source_timestamp=2026-05-05T06:36:36Z`

## Related Pages

- `dispute-module-ipa-staking-integration`
- `wrapped-ip-erc20-permit`

## Sources

- `source_document_id`: `srcdoc_e6330768515dc61898e883dd66fdd63f`
- `source_revision_id`: `srcrev_aba2bccd4890de44c51dfab17b2c3ca1`
- `source_url`: [Notion source](https://www.notion.so/Dispute-and-IPA-Staking-ca9eb8468cdc4846879fe9173e0e4980)
