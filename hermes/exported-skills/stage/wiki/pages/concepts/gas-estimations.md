---
title: "Gas Estimations for RSI Protocol Interactions"
type: "concept"
slug: "concepts/gas-estimations"
freshness: "2026-05-05T06:28:35Z"
tags:
  - "arbitrum"
  - "costs"
  - "ethereum"
  - "gas"
  - "optimism"
owners: []
source_revision_ids:
  - "srcrev_6198c1c7dcaaa4e572d20a9134f6680e"
conflict_state: "none"
---

# Gas Estimations for RSI Protocol Interactions

## Summary

Preliminary gas cost estimates for interacting with RSI contracts on Ethereum mainnet and Layer 2 networks (Optimism/Arbitrum).

## Claims

- Transaction cost formula: gas used (wei) * gas price (gwei) / 1e18 * ETH price. Example: 332,242 wei * 50 gwei / 1e18 * $2300 = $38. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`
- Average gas price on Ethereum is 27–50 gwei. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`
- registerIpOrg costs 262,410 wei on Ethereum, approximately $20–25. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`
- registerIPAsset costs 332,243 wei on Ethereum, approximately $30–35. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`
- configureIpOrgLicensing costs 132,008 wei on Ethereum, approximately $10–15. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`
- createLicense costs 482,120 wei on Ethereum, approximately $40–50. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`
- Etherscan estimated gas costs for common operations: Swap 356k wei, NFT Sale 602k wei, Bridging 114k wei, Borrowing 302k wei. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`
- Additional Ethereum gas references: ERC-721Enumerable mint 114k wei, USDC transfer 44k wei. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`
- On Optimism/Arbitrum, average gas price is 0.1 gwei, making RSI interactions much cheaper: registerIpOrg $0.06, registerIPAsset $0.08, configureIpOrgLicensing $0.03, createLicense $0.11 (at $2300/ETH). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a) `source_document_id=srcdoc_99e53819779178631445b9587947152a` `source_revision_id=srcrev_6198c1c7dcaaa4e572d20a9134f6680e` `chunk_id=srcchunk_162c1e70b2b46d154f016add36b4475e` `native_locator=https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a` `source_timestamp=2026-05-05T06:28:35Z`

## Sources

- `source_document_id`: `srcdoc_99e53819779178631445b9587947152a`
- `source_revision_id`: `srcrev_6198c1c7dcaaa4e572d20a9134f6680e`
- `source_url`: [Notion source](https://www.notion.so/Gas-Estimations-prelim-on-alpha-rc0-ca4efc315c104e0086519a280b3d902a)
