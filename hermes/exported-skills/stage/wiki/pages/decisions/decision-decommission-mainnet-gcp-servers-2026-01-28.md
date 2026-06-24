---
title: "Decommission Mainnet GCP Servers (January 2026)"
type: "decision"
slug: "decisions/decision-decommission-mainnet-gcp-servers-2026-01-28"
freshness: "2026-01-29T12:02:40Z"
tags:
  - "decommissioning"
  - "devops"
  - "gcp"
  - "mainnet"
  - "servers"
owners:
  - "U07A7AUGL5V"
  - "U08332YRB7W"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_1d36684274566a7b56e91a49e23c6a13"
  - "srcrev_24d5beed1ba17976d51e1c03cf15a60f"
  - "srcrev_359e1ecae71de54ea1b84dc97a8c3931"
  - "srcrev_d10d91aee35de53a9916592e764c5a72"
conflict_state: "none"
---

# Decommission Mainnet GCP Servers (January 2026)

## Summary

Decision to remove mainnet GCP servers due to no traffic, retaining only four instances (tmp-indexer-hans, use1-mainnet-bootnode1, use1-mainnet-bootnode2, use1-mainnet-monitoring). Removal executed on 2026-01-28 after team confirmation.

## Claims

- Mainnet had no traffic in GCP as of January 28, 2026. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_1d36684274566a7b56e91a49e23c6a13` `chunk_id=srcchunk_c90f2a99ec12a99e1e02ea094877ee02` `native_locator=slack:C0547N89JUB:1769644852.834339:1769644852.834339` `source_timestamp=2026-01-29T00:47:10Z`
- Scheduled removal of mainnet servers at 19:00 PT on 2026-01-28. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_1d36684274566a7b56e91a49e23c6a13` `chunk_id=srcchunk_c90f2a99ec12a99e1e02ea094877ee02` `native_locator=slack:C0547N89JUB:1769644852.834339:1769644852.834339` `source_timestamp=2026-01-29T00:47:10Z`
- Servers to be retained (not deleted): tmp-indexer-hans, use1-mainnet-bootnode1, use1-mainnet-bootnode2. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_1d36684274566a7b56e91a49e23c6a13` `chunk_id=srcchunk_c90f2a99ec12a99e1e02ea094877ee02` `native_locator=slack:C0547N89JUB:1769644852.834339:1769644852.834339` `source_timestamp=2026-01-29T00:47:10Z`
- List of servers targeted for deletion (20+ servers: yao-test-mainnet-2, use1-mainnet-*, euw3-mainnet-*, asias1-mainnet-*, etc.). `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_1d36684274566a7b56e91a49e23c6a13` `chunk_id=srcchunk_c90f2a99ec12a99e1e02ea094877ee02` `native_locator=slack:C0547N89JUB:1769644852.834339:1769644852.834339` `source_timestamp=2026-01-29T00:47:10Z`
- Team member confirmed the deletion, stating mainnet runs without issues. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_24d5beed1ba17976d51e1c03cf15a60f` `chunk_id=srcchunk_acb43016e5a8b44a90e912f21993506c` `native_locator=slack:C0547N89JUB:1769644852.834339:1769687438.962409` `source_timestamp=2026-01-29T11:50:38Z`
- Removal executed after confirmation from team members. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_359e1ecae71de54ea1b84dc97a8c3931` `chunk_id=srcchunk_e68fb2887e9908a88c554689bc13a866` `native_locator=slack:C0547N89JUB:1769644852.834339:1769687504.695589` `source_timestamp=2026-01-29T11:51:44Z`
- All servers except tmp-indexer-hans, use1-mainnet-bootnode1, use1-mainnet-bootnode2, and use1-mainnet-monitoring were deleted. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_559e0554acbfffce7dc12e0e89ac200f` `source_revision_id=srcrev_d10d91aee35de53a9916592e764c5a72` `chunk_id=srcchunk_dd32a502abde59f5e74636b184f79b11` `native_locator=slack:C0547N89JUB:1769644852.834339:1769688160.399129` `source_timestamp=2026-01-29T12:02:40Z`

## Sources

- `source_document_id`: `srcdoc_559e0554acbfffce7dc12e0e89ac200f`
- `source_revision_id`: `srcrev_b98c601d1bfdc461967f6b8964d49688`
