---
title: "Etherscan Custom Error Handling Limitation"
type: "concept"
slug: "concepts/etherscan-custom-error-handling-limitation"
freshness: "2026-05-05T04:31:09Z"
tags:
  - "custom errors"
  - "debugging"
  - "etherscan"
  - "tenderly"
owners: []
source_revision_ids:
  - "srcrev_0f3cf6a9b645ca8742eae6ac46380c70"
conflict_state: "none"
---

# Etherscan Custom Error Handling Limitation

## Summary

Etherscan cannot handle custom errors well, a known issue for years. Use Tenderly transaction explorer to view custom errors in stack traces.

## Claims

- Etherscan cannot handle custom errors well; this is a known issue that developers have complained about for years. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0) `source_document_id=srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac` `source_revision_id=srcrev_0f3cf6a9b645ca8742eae6ac46380c70` `chunk_id=srcchunk_82d7505c2c79794f96136401202dbe7d` `native_locator=https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0` `source_timestamp=2026-05-05T04:31:09Z`
- Workaround: use Tenderly transaction explorer to show the custom error in the transaction stacktrace. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0) `source_document_id=srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac` `source_revision_id=srcrev_0f3cf6a9b645ca8742eae6ac46380c70` `chunk_id=srcchunk_82d7505c2c79794f96136401202dbe7d` `native_locator=https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0` `source_timestamp=2026-05-05T04:31:09Z`

## Related Pages

- `debugging-mint-and-register-ip-with-sig-revert`

## Sources

- `source_document_id`: `srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac`
- `source_revision_id`: `srcrev_0f3cf6a9b645ca8742eae6ac46380c70`
- `source_url`: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0)
