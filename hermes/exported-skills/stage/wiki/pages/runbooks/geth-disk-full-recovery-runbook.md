---
title: "Geth Disk Full Recovery on steven-vm"
type: "runbook"
slug: "runbooks/geth-disk-full-recovery-runbook"
freshness: "2026-06-16T02:39:12Z"
tags:
  - "disk"
  - "ebs"
  - "ec2"
  - "geth"
  - "recovery"
owners: []
source_revision_ids:
  - "srcrev_1c39a06930b6863405233dc1a984c996"
  - "srcrev_1fd718439a0e37a8d66f687017036ca6"
  - "srcrev_f19af41c7731afeddfed8e2499d4a38f"
conflict_state: "none"
---

# Geth Disk Full Recovery on steven-vm

## Summary

Steps to recover geth node on steven-vm when root disk is full, by increasing EBS volume and resizing filesystem.

## Claims

- The geth client on steven-vm (IP 3.226.240.21) shut down because the root disk was 100% full (1.2 TiB, ~2 GiB free). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_733fb7d3ff3796afee06f5dde787bf05` `source_revision_id=srcrev_f19af41c7731afeddfed8e2499d4a38f` `chunk_id=srcchunk_af2e906217852f4aacf7cc88025625ac` `native_locator=slack:C0547N89JUB:1781577344.125209:1781577344.125209` `source_timestamp=2026-06-16T02:35:44Z`
- The root volume is a gp3 EBS volume of 1200 GiB on /dev/sda1. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_733fb7d3ff3796afee06f5dde787bf05` `source_revision_id=srcrev_1c39a06930b6863405233dc1a984c996` `chunk_id=srcchunk_c2aefa2edb58d15732cc597d97484e3a` `native_locator=slack:C0547N89JUB:1781577344.125209:1781577441.627149` `source_timestamp=2026-06-16T02:37:21Z`
- The EBS volume was resized to 2000 GiB. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_733fb7d3ff3796afee06f5dde787bf05` `source_revision_id=srcrev_1fd718439a0e37a8d66f687017036ca6` `chunk_id=srcchunk_53fe9b8c737280a8e685fa3c5ee11c1e` `native_locator=slack:C0547N89JUB:1781577344.125209:1781577552.972529` `source_timestamp=2026-06-16T02:39:12Z`
- After volume growth, in-guest filesystem resize is needed; likely command: sudo resize2fs /dev/nvme0n1p1 or /dev/xvda1. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_733fb7d3ff3796afee06f5dde787bf05` `source_revision_id=srcrev_1fd718439a0e37a8d66f687017036ca6` `chunk_id=srcchunk_53fe9b8c737280a8e685fa3c5ee11c1e` `native_locator=slack:C0547N89JUB:1781577344.125209:1781577552.972529` `source_timestamp=2026-06-16T02:39:12Z`
- The geth node will not recover until there is free space on the root disk. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_733fb7d3ff3796afee06f5dde787bf05` `source_revision_id=srcrev_f19af41c7731afeddfed8e2499d4a38f` `chunk_id=srcchunk_af2e906217852f4aacf7cc88025625ac` `native_locator=slack:C0547N89JUB:1781577344.125209:1781577344.125209` `source_timestamp=2026-06-16T02:35:44Z`

## Sources

- `source_document_id`: `srcdoc_733fb7d3ff3796afee06f5dde787bf05`
- `source_revision_id`: `srcrev_1fd718439a0e37a8d66f687017036ca6`
