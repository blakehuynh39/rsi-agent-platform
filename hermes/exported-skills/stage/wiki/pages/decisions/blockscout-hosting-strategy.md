---
title: "Blockscout Hosting Strategy"
type: "decision"
slug: "decisions/blockscout-hosting-strategy"
freshness: "2026-02-03T03:07:47Z"
tags:
  - "aeneid"
  - "blockchain-explorer"
  - "blockscout"
  - "infrastructure"
  - "mainnet"
owners: []
source_revision_ids:
  - "srcrev_1e77f5fa854b12b45d73bbd301d79d9d"
  - "srcrev_1fe2d0a2ff388b301fa563f077769293"
  - "srcrev_2d3bbcb67b4230d277b27e579d8f7db7"
  - "srcrev_3d3ba1690e3455c6eed464fdfe84a634"
  - "srcrev_9284bb3965f3417e73ac6b1eb24dbe65"
  - "srcrev_949041360640422ba433f6a8592d4e56"
  - "srcrev_a7f33f0c26e9fd1aeffe202b6cd654f5"
  - "srcrev_ba8705404d21f8bf699ff656835b2652"
conflict_state: "none"
---

# Blockscout Hosting Strategy

## Summary

Decision to stop running our own Blockscout instances for Mainnet and Aeneid, relying on Blockscout team's hosted service and Dune. Our own instance will be kept as cold backup, but spinning it up could take days and be expensive.

## Claims

- Mainnet and Aeneid already use Blockscout team's hosted service and Dune, not our own Blockscout. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_2d3bbcb67b4230d277b27e579d8f7db7` `chunk_id=srcchunk_126771e4169623acacd9804c2862cf4b` `native_locator=slack:C0547N89JUB:1770083176.152289:1770083176.152289` `source_timestamp=2026-02-03T01:46:16Z`
- We have our own Blockscout as backup. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_1e77f5fa854b12b45d73bbd301d79d9d` `chunk_id=srcchunk_36999eb627a5bf66c7503fd317f062b0` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087507.631609` `source_timestamp=2026-02-03T02:58:27Z`
- Spinning up our own Blockscout would take several days. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_3d3ba1690e3455c6eed464fdfe84a634` `chunk_id=srcchunk_c39f9a27d7da2240b0858779d7917fc7` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087669.132479` `source_timestamp=2026-02-03T03:01:09Z`
- On Mainnet, syncing required terabytes and was expensive on GCP. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_3d3ba1690e3455c6eed464fdfe84a634` `chunk_id=srcchunk_c39f9a27d7da2240b0858779d7917fc7` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087669.132479` `source_timestamp=2026-02-03T03:01:09Z`
- Even with a snapshot, if not kept running, syncing to the latest state would take time. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_949041360640422ba433f6a8592d4e56` `chunk_id=srcchunk_febd2f4cf939d936227e57d37a24d863` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087966.039759` `source_timestamp=2026-02-03T03:06:06Z`
- Blockscout team recovering their own instance might be faster than us spinning up our backup. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_9284bb3965f3417e73ac6b1eb24dbe65` `chunk_id=srcchunk_cde0de2e8db7571227126495301ed5dd` `native_locator=slack:C0547N89JUB:1770083176.152289:1770087998.521559` `source_timestamp=2026-02-03T03:06:38Z`
- Decision made to stop running our own Blockscout instances. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_ba8705404d21f8bf699ff656835b2652` `chunk_id=srcchunk_9a43f0a904da7de89dc3cf346bdd7313` `native_locator=slack:C0547N89JUB:1770083176.152289:1770088033.344579` `source_timestamp=2026-02-03T03:07:13Z`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_a7f33f0c26e9fd1aeffe202b6cd654f5` `chunk_id=srcchunk_36136bc4d9b4a8d02dc9dff339fcf308` `native_locator=slack:C0547N89JUB:1770083176.152289:1770088056.255159` `source_timestamp=2026-02-03T03:07:36Z`
  - citation: `source_document_id=srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce` `source_revision_id=srcrev_1fe2d0a2ff388b301fa563f077769293` `chunk_id=srcchunk_2e5dad33740ec594bae176f69290c049` `native_locator=slack:C0547N89JUB:1770083176.152289:1770088067.184179` `source_timestamp=2026-02-03T03:07:47Z`

## Open Questions

- Is a snapshot available for the Blockscout database to accelerate spin-up?

## Sources

- `source_document_id`: `srcdoc_37a38b9a385fa8e3cdb8d8e1444948ce`
- `source_revision_id`: `srcrev_ba8705404d21f8bf699ff656835b2652`
