---
title: "Debugging mintAndRegisterIpWithSig Revert"
type: "runbook"
slug: "runbooks/debugging-mint-and-register-ip-with-sig-revert"
freshness: "2026-05-05T04:31:09Z"
tags:
  - "debugging"
  - "gas"
  - "mintAndRegisterIpWithSig"
  - "spg"
owners: []
source_revision_ids:
  - "srcrev_0f3cf6a9b645ca8742eae6ac46380c70"
conflict_state: "none"
---

# Debugging mintAndRegisterIpWithSig Revert

## Summary

When spg.mintAndRegisterIpWithSig() reverts, the root cause may be an out-of-gas error during contract deployment, not the spg function itself. Use Tenderly to debug and add gas limit checks.

## Claims

- Developer reported spg.mintAndRegisterIpWithSig() failed with no reason. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0) `source_document_id=srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac` `source_revision_id=srcrev_0f3cf6a9b645ca8742eae6ac46380c70` `chunk_id=srcchunk_82d7505c2c79794f96136401202dbe7d` `native_locator=https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0` `source_timestamp=2026-05-05T04:31:09Z`
- First debugging showed the error is not related to spg; it actually failed in user’s code during contract deployment due to out of gas. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0) `source_document_id=srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac` `source_revision_id=srcrev_0f3cf6a9b645ca8742eae6ac46380c70` `chunk_id=srcchunk_82d7505c2c79794f96136401202dbe7d` `native_locator=https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0` `source_timestamp=2026-05-05T04:31:09Z`
- User confirmed the failure was due to setting the gas limit when submitting the transaction. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0) `source_document_id=srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac` `source_revision_id=srcrev_0f3cf6a9b645ca8742eae6ac46380c70` `chunk_id=srcchunk_82d7505c2c79794f96136401202dbe7d` `native_locator=https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0` `source_timestamp=2026-05-05T04:31:09Z`
- After re-submitting a transaction, the issue reproduced and showed spg.mintAndRegisterIpWithSig() revert. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0) `source_document_id=srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac` `source_revision_id=srcrev_0f3cf6a9b645ca8742eae6ac46380c70` `chunk_id=srcchunk_82d7505c2c79794f96136401202dbe7d` `native_locator=https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0` `source_timestamp=2026-05-05T04:31:09Z`
- Root cause was pinpointed by debugging with Tenderly.co. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0) `source_document_id=srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac` `source_revision_id=srcrev_0f3cf6a9b645ca8742eae6ac46380c70` `chunk_id=srcchunk_82d7505c2c79794f96136401202dbe7d` `native_locator=https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0` `source_timestamp=2026-05-05T04:31:09Z`
- Adding code before calling spg resolved the issue (specific code not provided). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0) `source_document_id=srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac` `source_revision_id=srcrev_0f3cf6a9b645ca8742eae6ac46380c70` `chunk_id=srcchunk_82d7505c2c79794f96136401202dbe7d` `native_locator=https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0` `source_timestamp=2026-05-05T04:31:09Z`

## Open Questions

- What specific code was added before calling spg to resolve the issue?

## Related Pages

- `etherscan-custom-error-handling-limitation`

## Sources

- `source_document_id`: `srcdoc_6fe42aa8bfa199b8b7572b9bc5c087ac`
- `source_revision_id`: `srcrev_0f3cf6a9b645ca8742eae6ac46380c70`
- `source_url`: [Notion source](https://www.notion.so/Discord-question-Revert-when-calling-mintAndRegisterIpWithSig-method-d6c2163bf30547618b1bb1a9f09ff1e0)
