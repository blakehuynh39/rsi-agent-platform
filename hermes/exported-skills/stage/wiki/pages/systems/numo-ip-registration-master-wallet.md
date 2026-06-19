---
title: "Numo IP Registration Master Wallet"
type: "system"
slug: "systems/numo-ip-registration-master-wallet"
freshness: "2026-04-23T02:47:13Z"
tags:
  - "funding"
  - "ip-registration"
  - "mainnet"
  - "numo"
  - "wallet"
owners:
  - "Numo team"
source_revision_ids:
  - "srcrev_2c08ce313a75256a77a2f161b006a9d3"
  - "srcrev_4888388af1b3a04f29c5da829b4e7a27"
  - "srcrev_89cad84a7d4930c2b8515fb4b0fe9936"
  - "srcrev_90b3989ca10bb645bb368d8fec6f2ec5"
  - "srcrev_95feeba24834ea2716606dd7d480bc69"
  - "srcrev_a2ef78cdadd58c549fc226706198cfab"
  - "srcrev_b0e8f9164613f10170c70b25c25eab55"
  - "srcrev_d00eb8f2da0c64ca1d05927bba02f253"
conflict_state: "none"
---

# Numo IP Registration Master Wallet

## Summary

Master wallet address used for funding IP registration testing on mainnet. The private key is stored in AWS Vault and injected at runtime. A funding of 250 IP tokens was requested and fulfilled on 2026-04-23.

## Claims

- The master wallet address is 0xB6f315F1072781deBE4Af09B24D2CC7f796790de. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_4888388af1b3a04f29c5da829b4e7a27` `chunk_id=srcchunk_3ea786468eaed2ffab1f4e4cd84e7f77` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776903653.539259` `source_timestamp=2026-04-23T00:21:12Z`
- The wallet is used for funding IP registration testing on mainnet. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_4888388af1b3a04f29c5da829b4e7a27` `chunk_id=srcchunk_3ea786468eaed2ffab1f4e4cd84e7f77` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776903653.539259` `source_timestamp=2026-04-23T00:21:12Z`
- On 2026-04-23, the Numo team requested 250 IP tokens to be sent to this address to unblock IP registration testing. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_b0e8f9164613f10170c70b25c25eab55` `chunk_id=srcchunk_cd6b0b9081f18f74b65c7d508716d819` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776900599.639489` `source_timestamp=2026-04-22T23:29:59Z`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_4888388af1b3a04f29c5da829b4e7a27` `chunk_id=srcchunk_3ea786468eaed2ffab1f4e4cd84e7f77` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776903653.539259` `source_timestamp=2026-04-23T00:21:12Z`
- The private key for the wallet is stored in AWS Vault and injected into the pod at runtime. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_d00eb8f2da0c64ca1d05927bba02f253` `chunk_id=srcchunk_e909197a65646e54f4988d7ca32418a0` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776904398.714239` `source_timestamp=2026-04-23T00:33:18Z`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_90b3989ca10bb645bb368d8fec6f2ec5` `chunk_id=srcchunk_cd2d701769b431c5902b6db4cedb793c` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776904559.103349` `source_timestamp=2026-04-23T00:35:59Z`
- The current setup prioritizes velocity over maximum security; a more secure approach would use a dedicated TEE signing environment or third-party service. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_90b3989ca10bb645bb368d8fec6f2ec5` `chunk_id=srcchunk_cd2d701769b431c5902b6db4cedb793c` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776904559.103349` `source_timestamp=2026-04-23T00:35:59Z`
- The blast radius is limited to the current money distributed among all parallel signing wallets. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_90b3989ca10bb645bb368d8fec6f2ec5` `chunk_id=srcchunk_cd2d701769b431c5902b6db4cedb793c` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776904559.103349` `source_timestamp=2026-04-23T00:35:59Z`
- It was suggested to verify that logs around the pod with the injected private key are not leaking sensitive information. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_95feeba24834ea2716606dd7d480bc69` `chunk_id=srcchunk_d5fddc74b4ab004103ec14262ad26a5c` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776904595.332149` `source_timestamp=2026-04-23T00:36:44Z`
- Another team member can look into making a more secure setup if needed. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_89cad84a7d4930c2b8515fb4b0fe9936` `chunk_id=srcchunk_8e3b0ccba38b1b0f602cb2ecee74bb56` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776904570.254829` `source_timestamp=2026-04-23T00:36:10Z`
- The requested 250 IP tokens were transferred to the wallet and confirmed received. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_2c08ce313a75256a77a2f161b006a9d3` `chunk_id=srcchunk_88c36b6fd9963f2ae22e8a6c05ce12c6` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776906169.437569` `source_timestamp=2026-04-23T01:02:49Z`
  - citation: `source_document_id=srcdoc_f306905b25dfaa2356f0ee42ad0a9677` `source_revision_id=srcrev_a2ef78cdadd58c549fc226706198cfab` `chunk_id=srcchunk_634068235d748365989b72949b024f71` `native_locator=slack:C0AL7EKNHDF:1776900599.639489:1776912433.121809` `source_timestamp=2026-04-23T02:47:13Z`

## Open Questions

- Are pod logs containing the injected private key leaking?
- Will a dedicated TEE signing environment be implemented to enhance security?

## Sources

- `source_document_id`: `srcdoc_f306905b25dfaa2356f0ee42ad0a9677`
- `source_revision_id`: `srcrev_a2ef78cdadd58c549fc226706198cfab`
