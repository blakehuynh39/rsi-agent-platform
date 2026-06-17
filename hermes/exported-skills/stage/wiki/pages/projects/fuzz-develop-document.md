---
title: "Fuzz Develop Document"
type: "project"
slug: "projects/fuzz-develop-document"
freshness: "2026-06-17T01:12:00Z"
tags:
  - "fuzzing"
  - "staking"
  - "story-protocol"
  - "testing"
owners: []
source_revision_ids:
  - "srcrev_257de84b32d68ea01c671f4fa301fd28"
  - "srcrev_617ec298cd5a27855c9d634504318d37"
conflict_state: "none"
---

# Fuzz Develop Document

## Summary

A comprehensive document detailing the fuzz testing system for the Story blockchain, including automation, configuration, code organization, engine modes, and extension points.

## Claims

- The fuzz code lives in the `fuzz/` directory of the `storyprotocol/story-devnet-aws` repository. `claim:claim_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_855194209ad780f55a117a1bd96f2f6d` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1` `source_timestamp=2026-06-17T01:12:00Z`
- Fuzz drives random staking traffic against any already-running chain and uses an independent oracle to assert on-chain reward-weight invariants; the oracle auto-selects based on the chain's position relative to the V190 upgrade height — linear before V190, cubic after, and when the run crosses V190 it additionally asserts the linear→cubic migration of every position. `claim:claim_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_855194209ad780f55a117a1bd96f2f6d` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1` `source_timestamp=2026-06-17T01:12:00Z`
- Fuzz runs via GitHub Actions workflow `.github/workflows/fuzz-any-chain.yml`, triggered manually by `workflow_dispatch`. `claim:claim_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_855194209ad780f55a117a1bd96f2f6d` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1` `source_timestamp=2026-06-17T01:12:00Z`
- The workflow_dispatch inputs include required fields `v190_height`, `stop_block`, `el_rpc_url`, `cl_rest_url`, and optional fields with defaults: `wallets` (12), `max_stake_ip` (2047), `freeze_delta` (60), `seed` ('' auto), `per_block` (true), `fail_fast` (true), `jail_target` ('' skip), `jail_svc` (''). `claim:claim_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_855194209ad780f55a117a1bd96f2f6d` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1` `source_timestamp=2026-06-17T01:12:00Z`
- The fuzz binary is configured entirely via environment variables prefixed with `FUZZ_`; all parameters except the V190 height can be probed from chain APIs. `claim:claim_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-2) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_21dd84115d66878f0229d8394e1ea8c8` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-2` `source_timestamp=2026-06-17T01:12:00Z`
- Each wallet has its own random decision stream seeded `seed ^ walletIdx`, making the fuzz reproducible in terms of decisions (though chain trajectory may still diverge). `claim:claim_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-2) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_21dd84115d66878f0229d8394e1ea8c8` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-2` `source_timestamp=2026-06-17T01:12:00Z`
- The main orchestration flow in `main.go` (≈669 lines) includes: load config, validate required params, connect & health check EL/CL, probe chain parameters, fund derived wallets balance-aware, decide run mode & freeze window, start boundary watcher if crossing V190, run engine, drain wallets, verify & render summary, exit with non-zero on failure. `claim:claim_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-2) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_21dd84115d66878f0229d8394e1ea8c8` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-2` `source_timestamp=2026-06-17T01:12:00Z`
- Code organization includes `main.go` for orchestration, `actions.go` for action registry, `driver_perblock.go` for concurrent engine, `driver.go` for sequential engine, `oracle.go` for reward-share assertions, `chain.go` for contract bindings, and other files for specific responsibilities. `claim:claim_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-2) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_21dd84115d66878f0229d8394e1ea8c8` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-2` `source_timestamp=2026-06-17T01:12:00Z`
- There are two engines: per-block concurrent (default, `FUZZ_PER_BLOCK=1`) and classic sequential (`FUZZ_PER_BLOCK=0`), both sharing the same `pbActions` registry. `claim:claim_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_9f37c1f2bbe229e9341f1f9f62268710` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3` `source_timestamp=2026-06-17T01:12:00Z`
- To add a new action, implement a send function in `driver_perblock.go` and append an `actionSpec` entry to `pbActions` in `actions.go` with a weight and optional `w0Only` flag; no other changes are needed. `claim:claim_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_9f37c1f2bbe229e9341f1f9f62268710` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3` `source_timestamp=2026-06-17T01:12:00Z`
- Run mode is automatically decided: if start height ≥ V190, pure cubic mode; if V190 falls within [start, stop_block], cross mode with freeze window; otherwise linear-only mode. `claim:claim_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_9f37c1f2bbe229e9341f1f9f62268710` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3` `source_timestamp=2026-06-17T01:12:00Z`
- The freeze window (`freeze_delta` blocks around V190) prevents submitting transactions near the migration to avoid nonce collisions. `claim:claim_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_855194209ad780f55a117a1bd96f2f6d` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-1` `source_timestamp=2026-06-17T01:12:00Z`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_9f37c1f2bbe229e9341f1f9f62268710` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3` `source_timestamp=2026-06-17T01:12:00Z`
- The `w0Only` flag ensures ghost/on-behalf actions only run on wallet 0 to avoid nonce races. `claim:claim_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_9f37c1f2bbe229e9341f1f9f62268710` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-3` `source_timestamp=2026-06-17T01:12:00Z`
- The CI runner needs slack in stop_block (spinup+build ~60-90 blocks), so stop_block should be set as current height + desired span + slack. `claim:claim_14` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-4) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_98ebd96c759a0961b7011f5f92bbd4d6` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-4` `source_timestamp=2026-06-17T01:12:00Z`
- Security measures include `shellSafe` validation for `FUZZ_STORY_BIN`, `FUZZ_SSH_USER`, `FUZZ_JAIL_SVC` to reject shell metacharacters. `claim:claim_15` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-4) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_617ec298cd5a27855c9d634504318d37` `chunk_id=srcchunk_98ebd96c759a0961b7011f5f92bbd4d6` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02#chunk-4` `source_timestamp=2026-06-17T01:12:00Z`
- The Fuzz Develop Document is empty. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02) `source_document_id=srcdoc_cd771799d96c1868906111fba78c24c7` `source_revision_id=srcrev_257de84b32d68ea01c671f4fa301fd28` `chunk_id=srcchunk_8a032549f4f6f936ff5dd76cca181e12` `native_locator=https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02` `source_timestamp=2026-06-16T15:15:00Z`

## Sources

- `source_document_id`: `srcdoc_cd771799d96c1868906111fba78c24c7`
- `source_revision_id`: `srcrev_617ec298cd5a27855c9d634504318d37`
- `source_url`: [source](https://app.notion.com/p/Fuzz-Develop-Document-381051299a548046b59fc3eeeb4ddf02)
