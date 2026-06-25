---
title: "Decision to Stop Self-Hosted Blockscout Explorers"
type: "decision"
slug: "decisions/stop-self-hosted-blockscout-explorers"
freshness: "2026-02-03T03:07:47Z"
tags:
  - "blockscout"
  - "cost-optimization"
  - "explorer"
  - "infrastructure"
owners: []
source_revision_ids:
  - "srcrev_1e77f5fa854b12b45d73bbd301d79d9d"
  - "srcrev_1fe2d0a2ff388b301fa563f077769293"
  - "srcrev_2d3bbcb67b4230d277b27e579d8f7db7"
  - "srcrev_3d3ba1690e3455c6eed464fdfe84a634"
  - "srcrev_9284bb3965f3417e73ac6b1eb24dbe65"
  - "srcrev_949041360640422ba433f6a8592d4e56"
conflict_state: "none"
---

# Decision to Stop Self-Hosted Blockscout Explorers

## Summary

On 2026-01-22, the team decided to stop running self-hosted Blockscout instances on Mainnet and Aeneid, relying on the Blockscout team's hosted service and Dune, with manual backup only if needed.

## Claims

- Blockscout team already hosts the explorer service on Mainnet and Aeneid, and we have Dune. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_2d3bbcb67b4230d277b27e579d8f7db7` `chunk_id=srcchunk_126771e4169623acacd9804c2862cf4b` `native_locator=slack:C0547N89JUB:1770083176.152289:1770083176.152289` `source_timestamp=2026-02-03T01:46:16Z`
- We have our own Blockscout as backup. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_1e77f5fa854b12b45d73bbd301d79d9d` `chunk_id=srcchunk_36999eb627a5bf66c7503fd317f062b0` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087507.631609` `source_timestamp=2026-02-03T02:58:27Z`
- Spinning up our own Blockscout from scratch takes several days and requires terabytes of sync on Mainnet, making it expensive. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_3d3ba1690e3455c6eed464fdfe84a634` `chunk_id=srcchunk_c39f9a27d7da2240b0858779d7917fc7` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087669.132479` `source_timestamp=2026-02-03T03:01:09Z`
- Even if we use a snapshot, if we don't keep running it will take time to sync with the latest state. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_949041360640422ba433f6a8592d4e56` `chunk_id=srcchunk_febd2f4cf939d936227e57d37a24d863` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087966.039759` `source_timestamp=2026-02-03T03:06:06Z`
- The feasibility of a quick spin-up depends on snapshot freshness, and restoring Blockscout team's instances is likely faster. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_9284bb3965f3417e73ac6b1eb24dbe65` `chunk_id=srcchunk_cde0de2e8db7571227126495301ed5dd` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087998.521559` `source_timestamp=2026-02-03T03:06:38Z`
- The decision was made to stop running our own Blockscout instances. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_1fe2d0a2ff388b301fa563f077769293` `chunk_id=srcchunk_2e5dad33740ec594bae176f69290c049` `native_locator=slack:C0547N89JUB:1770083176.152289:1770088067.184179` `source_timestamp=2026-02-03T03:07:47Z`

## Open Questions

- How to quickly spin up a backup Blockscout instance if the hosted service fails? (Snapshot availability, time, cost)

## Sources

- `source_document_id`: `srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce`
- `source_revision_id`: `srcrev_9284bb3965f3417e73ac6b1eb24dbe65`
