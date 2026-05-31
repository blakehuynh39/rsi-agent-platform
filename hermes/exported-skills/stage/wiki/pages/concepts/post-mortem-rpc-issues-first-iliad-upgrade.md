---
title: "Post-Mortem: RPC issues in the first iliad upgrade"
type: "concept"
slug: "concepts/post-mortem-rpc-issues-first-iliad-upgrade"
freshness: "2024-09-10T22:08:00Z"
tags:
  - "iliad-testnet"
  - "post-mortem"
  - "rpc"
  - "soft-fork"
  - "upgrade"
owners: []
source_revision_ids:
  - "srcrev_23a3f91ae3686c4ec845f32a1481b949"
conflict_state: "none"
---

# Post-Mortem: RPC issues in the first iliad upgrade

## Summary

Post-mortem analysis of RPC issues encountered during the first soft fork upgrade on the Iliad testnet on 9/9/2024. The upgrade aimed to reduce block time by removing unnecessary logs, but RPC and boot nodes running old versions fell behind validators, causing RPC test failures and transfer issues.

## Claims

- Iliad testnet was experiencing long block time (average 4.8s) before the upgrade. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
- The first soft fork upgrade on Iliad testnet aimed to remove unnecessary logs and reduce block time. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
- Genesis validators were upgraded one by one using a single node upgrade workflow between 9:03 PM and 9:49 PM PT on 9/9/2024. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
- RPC and boot node upgrades were deferred to the next day because they were not tested in devnet. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
- At 10:39 PM PT, the test team reported that RPC tests failed. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
- At 10:41 PM PT, the application team reported that IP cannot be transferred and Faucet stopped working. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
- At 10:51 PM PT, the team reviewed the Grafana dashboard and observed that RPC and boot nodes were far behind the genesis validators. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
- The root cause was that RPC and boot nodes were using old versions of the story binary, taking more time to generate logs and falling behind in block processing. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_e252f3084314b20ff6e72ac3b55b89a1` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2` `source_timestamp=2024-09-10T22:08:00Z`
- Single node upgrades on all RPC and boot nodes were performed between 10:57 PM and 11:37 PM PT, after which block heights caught up and token transfers resumed. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_a53ba48fc67bc548493a1ba14460cf51` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-1` `source_timestamp=2024-09-10T22:08:00Z`
- All story-hosted nodes should be upgraded at a similar time even for non-consensus-breaking changes, as version discrepancies can lead to usability issues over time. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_e252f3084314b20ff6e72ac3b55b89a1` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2` `source_timestamp=2024-09-10T22:08:00Z`
- Sufficient external and internal communications are needed for both soft fork and hard fork upgrades to reduce surprises and facilitate issue resolution. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_e252f3084314b20ff6e72ac3b55b89a1` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2` `source_timestamp=2024-09-10T22:08:00Z`
- Monitoring network degradation over time is necessary to detect issues that accumulate, rather than only monitoring for a few minutes after critical operations. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_e252f3084314b20ff6e72ac3b55b89a1` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2` `source_timestamp=2024-09-10T22:08:00Z`
- Runbooks for soft fork and hard fork upgrades should be created to reduce manual errors and help stakeholders anticipate next steps. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_e252f3084314b20ff6e72ac3b55b89a1` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2` `source_timestamp=2024-09-10T22:08:00Z`
- Devnet should be made more similar to testnet to detect and fix issues earlier. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2) `source_document_id=srcdoc_dd31c923337dbda48162c6da600ccd7f` `source_revision_id=srcrev_23a3f91ae3686c4ec845f32a1481b949` `chunk_id=srcchunk_e252f3084314b20ff6e72ac3b55b89a1` `native_locator=https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f#chunk-2` `source_timestamp=2024-09-10T22:08:00Z`

## Sources

- `source_document_id`: `srcdoc_dd31c923337dbda48162c6da600ccd7f`
- `source_revision_id`: `srcrev_23a3f91ae3686c4ec845f32a1481b949`
- `source_url`: [Notion source](https://www.notion.so/Post-Mortem-RPC-issues-in-the-first-iliad-upgrade-bea89953f6ef4a30ad20f12fb38f221f)
