---
title: "Meta-Transaction Architecture Decision"
type: "decision"
slug: "decisions/meta-transaction-architecture-decision"
freshness: "2024-03-25T18:14:00Z"
tags:
  - "architecture"
  - "EIP-1271"
  - "EIP-712"
  - "IPAccount"
  - "meta-transactions"
owners: []
source_revision_ids:
  - "srcrev_66d9d044fa15540d2fb8ffdce5afb612"
conflict_state: "none"
---

# Meta-Transaction Architecture Decision

## Summary

Analysis of two approaches for implementing meta-transactions in Story Protocol: IPAccount-level (current) vs Protocol-level. The current approach uses a single entry point via IPAccount.executeWithSig, while the alternative would add signature versions to individual protocol functions.

## Claims

- Currently, signature verification is enforced in IPAccount.executeWithSig, and once verified, the call is routed through the protocol like a normal call to IPAccount.execute. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd) `source_document_id=srcdoc_91077e5c486f3db5daba4fb3f5b1ddad` `source_revision_id=srcrev_66d9d044fa15540d2fb8ffdce5afb612` `chunk_id=srcchunk_32f344ea482538c43d6f47398eb74b72` `native_locator=https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd` `source_timestamp=2024-03-25T18:14:00Z`
- The current approach treats both execute and executeWithSig calls equally at the protocol level, providing a single entry point for meta-transaction verification and execution. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd) `source_document_id=srcdoc_91077e5c486f3db5daba4fb3f5b1ddad` `source_revision_id=srcrev_66d9d044fa15540d2fb8ffdce5afb612` `chunk_id=srcchunk_32f344ea482538c43d6f47398eb74b72` `native_locator=https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd` `source_timestamp=2024-03-25T18:14:00Z`
- The IPAccount-level approach exposes ALL functions in the protocol for signature-based calls, resulting in zero granular control over direct vs. sig-based calls. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd) `source_document_id=srcdoc_91077e5c486f3db5daba4fb3f5b1ddad` `source_revision_id=srcrev_66d9d044fa15540d2fb8ffdce5afb612` `chunk_id=srcchunk_32f344ea482538c43d6f47398eb74b72` `native_locator=https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd` `source_timestamp=2024-03-25T18:14:00Z`
- The IPAccount-level approach has no duplication of functions and their signature versions, which lowers attack surface and contract size. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd) `source_document_id=srcdoc_91077e5c486f3db5daba4fb3f5b1ddad` `source_revision_id=srcrev_66d9d044fa15540d2fb8ffdce5afb612` `chunk_id=srcchunk_32f344ea482538c43d6f47398eb74b72` `native_locator=https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd` `source_timestamp=2024-03-25T18:14:00Z`
- The alternative Protocol-level approach would implement signature versions of functions (e.g., mintLicenseWithSig) that accept a signature as an extra field, which IPAccount can call via execute like a normal call. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd) `source_document_id=srcdoc_91077e5c486f3db5daba4fb3f5b1ddad` `source_revision_id=srcrev_66d9d044fa15540d2fb8ffdce5afb612` `chunk_id=srcchunk_32f344ea482538c43d6f47398eb74b72` `native_locator=https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd` `source_timestamp=2024-03-25T18:14:00Z`

## Open Questions

- How to handle modules or registries that might not want signature-based calls?
- Should the protocol adopt IPAccount-level or Protocol-level meta-transactions?

## Related Pages

- `referenced-page`

## Sources

- `source_document_id`: `srcdoc_91077e5c486f3db5daba4fb3f5b1ddad`
- `source_revision_id`: `srcrev_66d9d044fa15540d2fb8ffdce5afb612`
- `source_url`: [Notion source](https://www.notion.so/9-Meta-Tx-sig-on-IPAccount-level-or-Protocol-level-f778a91f65b84050b21bab58dbfaa8bd)
