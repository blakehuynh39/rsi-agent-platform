---
title: "Test Wallet Policy for Mainnet IP"
type: "policy"
slug: "policies/test-wallet-policy-mainnet-ip"
freshness: "2026-06-10T00:56:49Z"
tags:
  - "funding"
  - "IP"
  - "mainnet"
  - "testing"
  - "token"
  - "wallet"
owners: []
source_revision_ids:
  - "srcrev_1ea27d51fbc6748fa1ac1b3f54e6c806"
  - "srcrev_a95d7392827e32fdb603d5336e755f5e"
  - "srcrev_b7777f6297702976263b20cc13745176"
  - "srcrev_e4973129a90c091ee1b9cb7ecdfbbf85"
conflict_state: "none"
---

# Test Wallet Policy for Mainnet IP

## Summary

Policy and process for testing with IP tokens on mainnet: use dedicated fresh wallets, fund only the minimum amount needed, never reuse infra/operator wallets or share private keys, and coordinate with custodians to obtain test funds.

## Claims

- There is no documented shared mainnet test wallet. `claim:no_documented_shared_test_wallet` `confidence:1.00`
  - citation: `source_document_id=srcdoc_341dbeb580ac6770b637d927065ba0b5` `source_revision_id=srcrev_1ea27d51fbc6748fa1ac1b3f54e6c806` `chunk_id=srcchunk_fa9a3163dbbb1f3d225efeedab4c4085` `native_locator=slack:C0547N89JUB:1780997154.068509:1780997193.656579` `source_timestamp=2026-06-09T09:26:33Z`
  - citation: `source_document_id=srcdoc_341dbeb580ac6770b637d927065ba0b5` `source_revision_id=srcrev_e4973129a90c091ee1b9cb7ecdfbbf85` `chunk_id=srcchunk_cf7be54ef77f972ea2204bda8072f054` `native_locator=slack:C0547N89JUB:1780997154.068509:1780998807.842599` `source_timestamp=2026-06-09T09:53:27Z`
- For mainnet IP testing, use a dedicated fresh wallet and fund it with the minimum IP needed. `claim:recommended_practice_dedicated_wallet_min_funding` `confidence:1.00`
  - citation: `source_document_id=srcdoc_341dbeb580ac6770b637d927065ba0b5` `source_revision_id=srcrev_1ea27d51fbc6748fa1ac1b3f54e6c806` `chunk_id=srcchunk_fa9a3163dbbb1f3d225efeedab4c4085` `native_locator=slack:C0547N89JUB:1780997154.068509:1780997193.656579` `source_timestamp=2026-06-09T09:26:33Z`
- Do not reuse infrastructure or operator wallets; do not share private keys in Slack. `claim:security_rules_no_reuse_no_share_keys` `confidence:1.00`
  - citation: `source_document_id=srcdoc_341dbeb580ac6770b637d927065ba0b5` `source_revision_id=srcrev_1ea27d51fbc6748fa1ac1b3f54e6c806` `chunk_id=srcchunk_fa9a3163dbbb1f3d225efeedab4c4085` `native_locator=slack:C0547N89JUB:1780997154.068509:1780997193.656579` `source_timestamp=2026-06-09T09:26:33Z`
- To fund a test wallet, provide the test address and amount needed to Woojin or the custodian; they will fund it. `claim:custodian_funding_process` `confidence:1.00`
  - citation: `source_document_id=srcdoc_341dbeb580ac6770b637d927065ba0b5` `source_revision_id=srcrev_1ea27d51fbc6748fa1ac1b3f54e6c806` `chunk_id=srcchunk_fa9a3163dbbb1f3d225efeedab4c4085` `native_locator=slack:C0547N89JUB:1780997154.068509:1780997193.656579` `source_timestamp=2026-06-09T09:26:33Z`
- A test wallet request was made for 2 IP and the address 0x9dd1C4d9Dc87dDbF4fa1721b94B7Af4F08D8A83C. `claim:example_request_2ip_to_address` `confidence:1.00`
  - citation: `source_document_id=srcdoc_341dbeb580ac6770b637d927065ba0b5` `source_revision_id=srcrev_a95d7392827e32fdb603d5336e755f5e` `chunk_id=srcchunk_043148112b656f086b478f81d8eaa6b2` `native_locator=slack:C0547N89JUB:1780997154.068509:1781053009.241579` `source_timestamp=2026-06-10T00:56:49Z`
  - citation: `source_document_id=srcdoc_341dbeb580ac6770b637d927065ba0b5` `source_revision_id=srcrev_b7777f6297702976263b20cc13745176` `chunk_id=srcchunk_261f02185b46b6f8598e9f50b016c0de` `native_locator=slack:C0547N89JUB:1780997154.068509:1781016511.323629` `source_timestamp=2026-06-09T14:48:31Z`

## Open Questions

- Should we maintain a shared mainnet test wallet for IP testing?
- Who is the designated custodian for mainnet IP test funds (besides Woojin)?

## Sources

- `source_document_id`: `srcdoc_341dbeb580ac6770b637d927065ba0b5`
- `source_revision_id`: `srcrev_a95d7392827e32fdb603d5336e755f5e`
