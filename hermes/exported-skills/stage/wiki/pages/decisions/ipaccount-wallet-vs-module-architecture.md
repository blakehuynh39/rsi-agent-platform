---
title: "Decision: IPAccount as a Generic Wallet vs. Restricted Module-based Wallet"
type: "decision"
slug: "decisions/ipaccount-wallet-vs-module-architecture"
freshness: "2024-03-12T21:19:00Z"
tags:
  - "architecture"
  - "erc-7579"
  - "ipaccount"
  - "module-system"
  - "security"
  - "wallet"
owners:
  - "JZ"
  - "Leo"
  - "Protocol Team"
source_revision_ids:
  - "srcrev_73d33af147c2ed975c29724c17e77194"
conflict_state: "none"
---

# Decision: IPAccount as a Generic Wallet vs. Restricted Module-based Wallet

## Summary

Meeting on 03/11/24 to decide whether IPAccount should be a generic AA wallet capable of interacting with external protocols, or a restricted module-based system (similar to ERC-7579) focused on intra-protocol logic. The discussion weighed the necessity of a wallet for Story Protocol, the core primitive of an IP vault, and the security/UX trade-offs of generic execution.

## Claims

- The meeting's purpose was to discuss the degree of autonomy of IPAccount within and outside Story Protocol, and to decide whether IPAccount should be a generic wallet or a restricted module-based system. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52) `source_document_id=srcdoc_16f7fe3cf1bad770c60f1c715867bdbb` `source_revision_id=srcrev_73d33af147c2ed975c29724c17e77194` `chunk_id=srcchunk_1bbef1a74f83b5fd6f16b23cde10e71b` `native_locator=https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52` `source_timestamp=2024-03-12T21:19:00Z`
- Leeren stated that the only absolute requirement for a generic external execute function is transferring Royalty NFTs or tokens from IPAccount based on the current code design. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52) `source_document_id=srcdoc_16f7fe3cf1bad770c60f1c715867bdbb` `source_revision_id=srcrev_73d33af147c2ed975c29724c17e77194` `chunk_id=srcchunk_1bbef1a74f83b5fd6f16b23cde10e71b` `native_locator=https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52` `source_timestamp=2024-03-12T21:19:00Z`
- IP should be able to hold Royalty NFTs, functioning as an IP vault. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52) `source_document_id=srcdoc_16f7fe3cf1bad770c60f1c715867bdbb` `source_revision_id=srcrev_73d33af147c2ed975c29724c17e77194` `chunk_id=srcchunk_1bbef1a74f83b5fd6f16b23cde10e71b` `native_locator=https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52` `source_timestamp=2024-03-12T21:19:00Z`
- Leeren expressed concern that building a permission system around interactions with both the protocol and external contracts in the same logic conflates generic functions with Story's core functionalities, posing security risks for users who do not require an IP wallet. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52) `source_document_id=srcdoc_16f7fe3cf1bad770c60f1c715867bdbb` `source_revision_id=srcrev_73d33af147c2ed975c29724c17e77194` `chunk_id=srcchunk_1bbef1a74f83b5fd6f16b23cde10e71b` `native_locator=https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52` `source_timestamp=2024-03-12T21:19:00Z`
- Using IPAccount generic execution for well-defined intra-protocol module calls is strictly worse UX than a dedicated module system, and there is a need to clearly distinguish between intra-protocol interactions and external protocols. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52) `source_document_id=srcdoc_16f7fe3cf1bad770c60f1c715867bdbb` `source_revision_id=srcrev_73d33af147c2ed975c29724c17e77194` `chunk_id=srcchunk_1bbef1a74f83b5fd6f16b23cde10e71b` `native_locator=https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52` `source_timestamp=2024-03-12T21:19:00Z`

## Open Questions

- Can IP vault be the core primitive, with IP wallet as a second-layer primitive?
- Do we need every IP to have a generic IPAccount, or can some be just records while others have more expressive features?
- Why is a wallet useful outside Story Protocol, given that external protocols must understand specific token structures (e.g., Uniswap LP tokens) to be composable?

## Related Pages

- `concept/ip-vault`
- `concept/ipaccount`
- `concept/royalty-nft`
- `decision/erc-7579-module-system`

## Sources

- `source_document_id`: `srcdoc_16f7fe3cf1bad770c60f1c715867bdbb`
- `source_revision_id`: `srcrev_73d33af147c2ed975c29724c17e77194`
- `source_url`: [Notion source](https://www.notion.so/Meeting-IPAccount-as-a-Generic-AA-Wallet-or-Restricted-Module-based-Wallet-1-c7f28d757ace41c09d6406bdcdce0f52)
