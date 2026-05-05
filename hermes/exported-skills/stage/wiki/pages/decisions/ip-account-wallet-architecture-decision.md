---
title: "IP Account Wallet Architecture Decision"
type: "decision"
slug: "decisions/ip-account-wallet-architecture-decision"
freshness: "2026-05-05T06:31:43Z"
tags:
  - "AA"
  - "architecture"
  - "ERC-6900"
  - "ERC-7579"
  - "IP Account"
  - "modules"
  - "wallet"
owners: []
source_revision_ids:
  - "srcrev_d74fb8144445f3465ff895464161e05d"
conflict_state: "none"
---

# IP Account Wallet Architecture Decision

## Summary

Decision on whether IP Account should be a generic AA wallet or a module-based wallet (6900/7579-like), with options for separating core-level and external-level interactions.

## Claims

- IP Account will strictly deal with non-core-level interactions following a prior vote. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Core components are defined as modules, registries, ACL, and other components written by Story Protocol. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Core-level interactions are all interactions with core components. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- External-level interactions are all interactions that are not core-level interactions, such as calls to community modules or calls to external contracts. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Community Modules are modules not written by Story Protocol. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Executor Modules are modules that allow interactions with external contracts outside Story Protocol, considered part of community modules. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- External contracts are all contracts on the blockchain that are not part of Story Protocol (blockchain contracts minus core components and community modules). `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- An AA wallet is a smart contract account capable of executing all function calls to all external contracts. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- A module-based wallet is a smart contract account that can only call external contracts via executor modules. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Option 1 for core logic is to keep existing IPAccount but separate logic between external and internal calls (WALLET approach), with two separate functions for generic and protocol-specific executions. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Option 2 for core logic is to keep existing IPAccount design in its entirety (WALLET approach). `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Generic AA wallet for IP Account (4337-compatible) was considered as a research topic but is struck through, indicating it is not a current option. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Option 1 for IPAccount wallet type is a module-based wallet learning from 6900/7579, where all external contract calls must go through executor modules (e.g., TokenWithdrawalModule, OpenSeaModule). `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Option 2 for IPAccount wallet type allows IPAccount execute to call external contracts without restriction (ACL), but internal core-level interactions cannot happen through execute and must call modules directly. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`
- Option 3 for IPAccount wallet type is the current approach where IPAccount execute can call external or internal contracts, with all calls going through ACL and signer or to must be a registered module. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc) `source_document_id=srcdoc_e3aab56b404270fe1d1988fc0441af10` `source_revision_id=srcrev_d74fb8144445f3465ff895464161e05d` `chunk_id=srcchunk_81711251c91872470158b9a93548424b` `native_locator=https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc` `source_timestamp=2026-05-05T06:31:43Z`

## Open Questions

- Which of the three IPAccount wallet type options (module-based, external-only execute, current approach) will be selected?
- Which of the two core logic options (separate functions vs. keep entire design) will be selected?

## Sources

- `source_document_id`: `srcdoc_e3aab56b404270fe1d1988fc0441af10`
- `source_revision_id`: `srcrev_d74fb8144445f3465ff895464161e05d`
- `source_url`: [Notion source](https://www.notion.so/Vote-IP-Account-as-a-generic-wallet-or-module-based-6900-7579-like-wallet-b89e944dfb364f18a68fda382c236cbc)
