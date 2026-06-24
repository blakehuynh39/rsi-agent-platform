---
title: "GCP Mainnet Server Decommissioning (2026-01-28)"
type: "decision"
slug: "decisions/gcp-mainnet-server-decommissioning-2026-01-28"
freshness: "2026-01-29T12:02:40Z"
tags:
  - "decommissioning"
  - "gcp"
  - "infrastructure"
  - "mainnet"
owners: []
source_revision_ids:
  - "srcrev_1d36684274566a7b56e91a49e23c6a13"
  - "srcrev_24d5beed1ba17976d51e1c03cf15a60f"
  - "srcrev_359e1ecae71de54ea1b84dc97a8c3931"
  - "srcrev_3796c8db48dde3f39729cf3e5412b348"
  - "srcrev_d10d91aee35de53a9916592e764c5a72"
  - "srcrev_d4b68110301b21636c0d55e3ac0dd077"
conflict_state: "none"
---

# GCP Mainnet Server Decommissioning (2026-01-28)

## Summary

Decommissioning of multiple GCP mainnet servers on 2026-01-28 due to lack of traffic, retaining tmp-indexer-hans, use1-mainnet-bootnode1, use1-mainnet-bootnode2, and use1-mainnet-monitoring.

## Claims

- A decision was made to remove mainnet servers in GCP due to no traffic, with execution planned for 19:00 PT. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_1d36684274566a7b56e91a49e23c6a13` `chunk_id=srcchunk_c90f2a99ec12a99e1e02ea094877ee02` `native_locator=slack:C0547N89JUB:1769644852.834339:1769644852.834339` `source_timestamp=2026-01-29T00:47:10Z`
- The initial list of servers scheduled for deletion included 21 servers: yao-test-mainnet-2, yao-test-mainnet-20251224-171555, use1-mainnet-bootnode-private1, use1-mainnet-internal-archive-rpc1, use1-mainnet-internal-full-rpc1, use1-mainnet-monitoring, use1-mainnet-partner-rpc1, use1-mainnet-public-rpc1, use1-mainnet-snapshot-archive1, use1-mainnet-snapshot-archive2, use1-mainnet-snapshot-full1, use1-mainnet-validator1, use1-mainnet-blockscout, use1-mainnet-internal-archive-rpc2, use1-mainnet-partner-rpc2, use1-mainnet-public-rpc2, use1-mainnet-validator2, yao-test-mainnet-beefier, asias1-mainnet-internal-full-rpc1, asias1-mainnet-public-rpc1, asias1-mainnet-public-rpc2, euw3-mainnet-internal-full-rpc1, euw3-mainnet-public-rpc1, euw3-mainnet-public-rpc2. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_1d36684274566a7b56e91a49e23c6a13` `chunk_id=srcchunk_c90f2a99ec12a99e1e02ea094877ee02` `native_locator=slack:C0547N89JUB:1769644852.834339:1769644852.834339` `source_timestamp=2026-01-29T00:47:10Z`
- Initially, the servers to be retained were tmp-indexer-hans, use1-mainnet-bootnode1, and use1-mainnet-bootnode2. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_1d36684274566a7b56e91a49e23c6a13` `chunk_id=srcchunk_c90f2a99ec12a99e1e02ea094877ee02` `native_locator=slack:C0547N89JUB:1769644852.834339:1769644852.834339` `source_timestamp=2026-01-29T00:47:10Z`
- A team member confirmed the deletion plan, stating they were good with it and that mainnet was running without issues. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_24d5beed1ba17976d51e1c03cf15a60f` `chunk_id=srcchunk_acb43016e5a8b44a90e912f21993506c` `native_locator=slack:C0547N89JUB:1769644852.834339:1769687438.962409` `source_timestamp=2026-01-29T11:50:38Z`
- The deletion was executed after confirmation, approximately around the time indicated. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_359e1ecae71de54ea1b84dc97a8c3931` `chunk_id=srcchunk_e68fb2887e9908a88c554689bc13a866` `native_locator=slack:C0547N89JUB:1769644852.834339:1769687504.695589` `source_timestamp=2026-01-29T11:51:44Z`
- Another server owner gave permission to remove their servers. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_3796c8db48dde3f39729cf3e5412b348` `chunk_id=srcchunk_fda7dcf8ad08429c9360df4b648514bb` `native_locator=slack:C0547N89JUB:1769644852.834339:1769687531.016289` `source_timestamp=2026-01-29T11:52:11Z`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_d4b68110301b21636c0d55e3ac0dd077` `chunk_id=srcchunk_87fc1100f30e91ab7996e4606acddb0b` `native_locator=slack:C0547N89JUB:1769644852.834339:1769687584.392489` `source_timestamp=2026-01-29T11:53:04Z`
- Final retained servers: tmp-indexer-hans, use1-mainnet-bootnode1, use1-mainnet-bootnode2, and use1-mainnet-monitoring. All other listed servers were deleted. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_d10d91aee35de53a9916592e764c5a72` `chunk_id=srcchunk_dd32a502abde59f5e74636b184f79b11` `native_locator=slack:C0547N89JUB:1769644852.834339:1769688160.399129` `source_timestamp=2026-01-29T12:02:40Z`
- The server use1-mainnet-monitoring was initially slated for deletion but was ultimately retained. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_1d36684274566a7b56e91a49e23c6a13` `chunk_id=srcchunk_c90f2a99ec12a99e1e02ea094877ee02` `native_locator=slack:C0547N89JUB:1769644852.834339:1769644852.834339` `source_timestamp=2026-01-29T00:47:10Z`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_d10d91aee35de53a9916592e764c5a72` `chunk_id=srcchunk_dd32a502abde59f5e74636b184f79b11` `native_locator=slack:C0547N89JUB:1769644852.834339:1769688160.399129` `source_timestamp=2026-01-29T12:02:40Z`

## Sources

- `source_document_id`: `srcdoc_559e0554acbfffce7dc12e0e89ac200f`
- `source_revision_id`: `srcrev_3796c8db48dde3f39729cf3e5412b348`
