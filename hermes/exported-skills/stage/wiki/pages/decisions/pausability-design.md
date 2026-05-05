---
title: "Pausability Design Decision"
type: "decision"
slug: "decisions/pausability-design"
freshness: "2026-05-05T06:33:04Z"
tags:
  - "governance"
  - "pausability"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_09c15303adb9798842f69d53439ab7f3"
conflict_state: "none"
---

# Pausability Design Decision

## Summary

Decision to implement a simple protocol-wide pause initially, with the PAUSER role granted to governance multisig, developer multisig, and possibly security council services.

## Claims

- Current implementation uses a home-made Pausable state in Governance.sol, which is a global pause but only used in IPAccount setPermissions. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd) `source_document_id=srcdoc_a544edb4a9ea9b863d623940bbcb4f48` `source_revision_id=srcrev_09c15303adb9798842f69d53439ab7f3` `chunk_id=srcchunk_43b5b052ae629c9f133d68c701a80ab9` `native_locator=https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd` `source_timestamp=2026-05-05T06:33:04Z`
- The plan is to expand pausability in two avenues: use audited OZ Pausable instead of custom code, and define if we require a protocol wide pause, by module or both. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd) `source_document_id=srcdoc_a544edb4a9ea9b863d623940bbcb4f48` `source_revision_id=srcrev_09c15303adb9798842f69d53439ab7f3` `chunk_id=srcchunk_43b5b052ae629c9f133d68c701a80ab9` `native_locator=https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd` `source_timestamp=2026-05-05T06:33:04Z`
- The current direction is to implement a simple protocol wide pause as a first simple approach, and be more specific later on. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd) `source_document_id=srcdoc_a544edb4a9ea9b863d623940bbcb4f48` `source_revision_id=srcrev_09c15303adb9798842f69d53439ab7f3` `chunk_id=srcchunk_43b5b052ae629c9f133d68c701a80ab9` `native_locator=https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd` `source_timestamp=2026-05-05T06:33:04Z`
- The PAUSER role will be granted to governance multisig, developer multisig, and possibly security council detector services. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd) `source_document_id=srcdoc_a544edb4a9ea9b863d623940bbcb4f48` `source_revision_id=srcrev_09c15303adb9798842f69d53439ab7f3` `chunk_id=srcchunk_43b5b052ae629c9f133d68c701a80ab9` `native_locator=https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd` `source_timestamp=2026-05-05T06:33:04Z`
- Pausable methods per module include AccessController (setPermission, setGlobalPermission), DisputeModule (raiseDispute, setDisputeJudgment), and LicensingModule (mintLicenseTokens, registerDerivative, registerDerivative...). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd) `source_document_id=srcdoc_a544edb4a9ea9b863d623940bbcb4f48` `source_revision_id=srcrev_09c15303adb9798842f69d53439ab7f3` `chunk_id=srcchunk_43b5b052ae629c9f133d68c701a80ab9` `native_locator=https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd` `source_timestamp=2026-05-05T06:33:04Z`
- Localized pause has pros (less impactful, less economic/reputational risk) and cons (harder to contain damage, may delay identification). Protocol-wide pause has pros (no need to isolate, safer) and cons (more economic/reputational risk). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd) `source_document_id=srcdoc_a544edb4a9ea9b863d623940bbcb4f48` `source_revision_id=srcrev_09c15303adb9798842f69d53439ab7f3` `chunk_id=srcchunk_43b5b052ae629c9f133d68c701a80ab9` `native_locator=https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd` `source_timestamp=2026-05-05T06:33:04Z`

## Open Questions

- Should the setGlobalPermission method be removed? (Noted in the document)
- Should we pause the solving/cancelling of existing disputes?

## Sources

- `source_document_id`: `srcdoc_a544edb4a9ea9b863d623940bbcb4f48`
- `source_revision_id`: `srcrev_09c15303adb9798842f69d53439ab7f3`
- `source_url`: [Notion source](https://www.notion.so/Pausability-2bfc8ee78ab547ffa33c602aeffbb2bd)
