---
title: "Blockscout Explorer Hosting Decision"
type: "decision"
slug: "decisions/blockscout-explorer-hosting-decision"
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
  - "srcrev_ba8705404d21f8bf699ff656835b2652"
conflict_state: "none"
---

# Blockscout Explorer Hosting Decision

## Summary

Decision to stop self-hosted Blockscout explorers for Mainnet and Aeneid, relying on Blockscout team's hosted service and Dune, with own instances kept as backup only for emergencies.

## Claims

- Blockscout team hosts a Blockscout explorer service for both Mainnet and Aeneid, and Dune is also available. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_2d3bbcb67b4230d277b27e579d8f7db7` `chunk_id=srcchunk_126771e4169623acacd9804c2862cf4b` `native_locator=slack:C0547N89JUB:1770083176.152289:1770083176.152289` `source_timestamp=2026-02-03T01:46:16Z`
- Our own Blockscout instances are maintained as a backup and should be stopped, only to be spun up if the Blockscout team's explorers become unavailable. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_1e77f5fa854b12b45d73bbd301d79d9d` `chunk_id=srcchunk_36999eb627a5bf66c7503fd317f062b0` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087507.631609` `source_timestamp=2026-02-03T02:58:27Z`
- Starting a Blockscout instance from scratch takes several days and requires terabytes of storage to sync on Mainnet, making it expensive when run on GCP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_3d3ba1690e3455c6eed464fdfe84a634` `chunk_id=srcchunk_c39f9a27d7da2240b0858779d7917fc7` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087669.132479` `source_timestamp=2026-02-03T03:01:09Z`
- Even with a database snapshot, significant catch-up time is required to sync latest blocks, and in an outage, recovery by the Blockscout team may be faster. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_949041360640422ba433f6a8592d4e56` `chunk_id=srcchunk_febd2f4cf939d936227e57d37a24d863` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087966.039759` `source_timestamp=2026-02-03T03:06:06Z`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_9284bb3965f3417e73ac6b1eb24dbe65` `chunk_id=srcchunk_cde0de2e8db7571227126495301ed5dd` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087998.521559` `source_timestamp=2026-02-03T03:06:38Z`
- The decision was made to stop running our own Blockscout instances. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_ba8705404d21f8bf699ff656835b2652` `chunk_id=srcchunk_9a43f0a904da7de89dc3cf346bdd7313` `native_locator=slack:C0547N89JUB:1770083176.152289:1770088033.344579` `source_timestamp=2026-02-03T03:07:13Z`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_1fe2d0a2ff388b301fa563f077769293` `chunk_id=srcchunk_2e5dad33740ec594bae176f69290c049` `native_locator=slack:C0547N89JUB:1770083176.152289:1770088067.184179` `source_timestamp=2026-02-03T03:07:47Z`

## Open Questions

- Is there a process to maintain recent snapshots to reduce spin-up time?
- What is the defined procedure and expected recovery time for spinning up the backup Blockscout instance when needed?

## Related Pages

- `backup-strategy`
- `blockscout`
- `dune-analytics`
- `infrastructure-cost-management`

## Sources

- `source_document_id`: `srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce`
- `source_revision_id`: `srcrev_1fe2d0a2ff388b301fa563f077769293`
