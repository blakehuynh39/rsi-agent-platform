---
title: "Reward backward compatibility patch"
type: "system"
slug: "systems/reward-backward-compatibility-patch"
freshness: "2025-07-14T18:08:00Z"
tags:
  - "backward-compatibility"
  - "patch"
  - "rewards"
owners: []
source_revision_ids:
  - "srcrev_f01f3c7323b513795e067fbd4640d883"
conflict_state: "none"
---

# Reward backward compatibility patch

## Summary

A patch to fix Cantina issue #67 by handling multiple reward requests from the same address caused state sync failures for old blocks. Only Kraken reported the issue and received a patch. No integrity or liveness impact. Learnings include improving backward compatibility testing and sharing context before public PRs.

## Claims

- The patch was prepared to fix Cantina issue #67 by updating the rewards functionality to handle two requests from the same address. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- The patch was not backward compatible, leading to different state when replaying old blocks with the new code. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- There was no integrity or liveness impact, and no chance of network state fork due to consensus requiring majority votes and finality and how upgrades are rolled out. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- The only known impact was that nodes performing state sync would fail the state root check when executing previous blocks. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- Only Kraken reported this issue and a patch was provided for them. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- No internal DEX database or external data sources were impacted or corrupted. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- Learning: The way of rolling out upgrades and the consensus mechanism prevents many possible network forks. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- Learning: Need to share context of patches to broader team before creating public pull requests. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- Learning: Need to update PR review checklists to ensure design is fully backward compatible. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`
- Learning: Need to ensure historic sync tests for new changes are complete before rollout, and find new ways like sample snapshots to do such testing. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361) `source_document_id=srcdoc_421d7855ff25083a3b6b049f8d0a894c` `source_revision_id=srcrev_f01f3c7323b513795e067fbd4640d883` `chunk_id=srcchunk_4b6c4cbcef71a506683791faba63044c` `native_locator=https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361` `source_timestamp=2025-07-14T18:08:00Z`

## Sources

- `source_document_id`: `srcdoc_421d7855ff25083a3b6b049f8d0a894c`
- `source_revision_id`: `srcrev_f01f3c7323b513795e067fbd4640d883`
- `source_url`: [Notion source](https://www.notion.so/Reward-backward-compatibility-patch-230051299a54809a9e5cf572bf535361)
