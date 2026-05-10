---
title: "IPAccount Design Decision: Generic AA Wallet vs Restricted Module-based Wallet"
type: "decision"
slug: "decisions/ipaccount-generic-aa-vs-restricted-module"
freshness: "2024-03-14T02:06:00Z"
tags:
  - "access-control"
  - "architecture"
  - "erc-6551"
  - "ipaccount"
owners:
  - "protocol engineers"
source_revision_ids:
  - "srcrev_2a72933ec1b21f8e624d43f793a290a2"
conflict_state: "none"
---

# IPAccount Design Decision: Generic AA Wallet vs Restricted Module-based Wallet

## Summary

Design decision between keeping IPAccount as a restricted module-based wallet (D.1) or making it a generic AA wallet (D.2) to improve composability with external protocols.

## Claims

- In Beta, the access check in _execute restricts either the signer or to address to be a registered module on Story. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-1) `source_document_id=srcdoc_ea81afd536a9386cd2f5713bc54e52f5` `source_revision_id=srcrev_2a72933ec1b21f8e624d43f793a290a2` `chunk_id=srcchunk_2b03b111ef94743638b8140832bfb5cd` `native_locator=https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-1` `source_timestamp=2024-03-14T02:06:00Z`
- IPAccounts could not transfer Royalty NFTs to other accounts, leading to the creation of TokenWithdrawModule as a temporary fix. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-1) `source_document_id=srcdoc_ea81afd536a9386cd2f5713bc54e52f5` `source_revision_id=srcrev_2a72933ec1b21f8e624d43f793a290a2` `chunk_id=srcchunk_2b03b111ef94743638b8140832bfb5cd` `native_locator=https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-1` `source_timestamp=2024-03-14T02:06:00Z`
- The access control restriction forces the creation of unique modules for each external protocol interaction (e.g., UniswapModule, OpenSeaModule), increasing audit burden and limiting composability. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-1) `source_document_id=srcdoc_ea81afd536a9386cd2f5713bc54e52f5` `source_revision_id=srcrev_2a72933ec1b21f8e624d43f793a290a2` `chunk_id=srcchunk_2b03b111ef94743638b8140832bfb5cd` `native_locator=https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-1` `source_timestamp=2024-03-14T02:06:00Z`
- Option D.1 keeps IPAccount as a restricted module-based wallet with minimal changes, requiring modules for external interactions. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-2) `source_document_id=srcdoc_ea81afd536a9386cd2f5713bc54e52f5` `source_revision_id=srcrev_2a72933ec1b21f8e624d43f793a290a2` `chunk_id=srcchunk_7c0687873213f63ec79f76c9e32a08ea` `native_locator=https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-2` `source_timestamp=2024-03-14T02:06:00Z`
- Option D.2 makes IPAccount a generic AA wallet by removing the access check, enabling direct integration with external protocols but requiring new security assumptions and possibly a diamond pattern. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-2) `source_document_id=srcdoc_ea81afd536a9386cd2f5713bc54e52f5` `source_revision_id=srcrev_2a72933ec1b21f8e624d43f793a290a2` `chunk_id=srcchunk_7c0687873213f63ec79f76c9e32a08ea` `native_locator=https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-2` `source_timestamp=2024-03-14T02:06:00Z`
- D.2 benefits: IPAccount can be marketed as a normal wallet, external protocols can integrate directly, and standard ERC-6551 execution can be used. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-2) `source_document_id=srcdoc_ea81afd536a9386cd2f5713bc54e52f5` `source_revision_id=srcrev_2a72933ec1b21f8e624d43f793a290a2` `chunk_id=srcchunk_7c0687873213f63ec79f76c9e32a08ea` `native_locator=https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-2` `source_timestamp=2024-03-14T02:06:00Z`
- D.2 downsides: hard to define scope of interactions, potential security issues, requires diamond pattern or similar architecture, and is a new paradigm requiring more design. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-2) `source_document_id=srcdoc_ea81afd536a9386cd2f5713bc54e52f5` `source_revision_id=srcrev_2a72933ec1b21f8e624d43f793a290a2` `chunk_id=srcchunk_7c0687873213f63ec79f76c9e32a08ea` `native_locator=https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb#chunk-2` `source_timestamp=2024-03-14T02:06:00Z`

## Open Questions

- How to ensure security if access check is removed?
- Should IPAccount be a generic AA wallet (D.2) or remain a restricted module-based wallet (D.1)?
- What architecture (e.g., diamond pattern, ERC-7579) would be needed for D.2?

## Related Pages

- `access-controller`
- `erc-6551`
- `ipaccount`
- `token-withdraw-module`

## Sources

- `source_document_id`: `srcdoc_ea81afd536a9386cd2f5713bc54e52f5`
- `source_revision_id`: `srcrev_2a72933ec1b21f8e624d43f793a290a2`
- `source_url`: [Notion source](https://www.notion.so/2-IPAccount-as-a-Generic-AA-wallet-or-Restricted-Module-based-Wallet-9dc7b789eb8446b6acec6a9562bfd0eb)
