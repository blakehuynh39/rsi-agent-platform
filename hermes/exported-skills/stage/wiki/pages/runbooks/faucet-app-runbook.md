---
title: "Faucet App Runbook"
type: "runbook"
slug: "runbooks/faucet-app-runbook"
freshness: "2026-05-05T06:38:10Z"
tags:
  - "faucet"
  - "runbook"
  - "supabase"
  - "wallet"
owners: []
source_revision_ids:
  - "srcrev_740ec26a1828f01ba2b83d2581d9a761"
conflict_state: "none"
---

# Faucet App Runbook

## Summary

Runbook for the faucet application, covering wallet funding checks and Supabase issues.

## Claims

- If the wallet has no money, check the two primary wallets and add more funds. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79) `source_document_id=srcdoc_a8afea7c5307d37ca95c6dfe019a94fb` `source_revision_id=srcrev_740ec26a1828f01ba2b83d2581d9a761` `chunk_id=srcchunk_aa591eee8d456bbe60341fa2ef371c51` `native_locator=https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79` `source_timestamp=2026-05-05T06:38:10Z`
- The two primary wallets to check are 0x0cbcbBA6781B8085A8e65CE478F956b4a3923baD and 0xe0Ae4dD1d7FF8C8A5E8Ad91591dfc4563931b736. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79) `source_document_id=srcdoc_a8afea7c5307d37ca95c6dfe019a94fb` `source_revision_id=srcrev_740ec26a1828f01ba2b83d2581d9a761` `chunk_id=srcchunk_aa591eee8d456bbe60341fa2ef371c51` `native_locator=https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79` `source_timestamp=2026-05-05T06:38:10Z`
- There are 20 wallets used by the faucet in total. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79) `source_document_id=srcdoc_a8afea7c5307d37ca95c6dfe019a94fb` `source_revision_id=srcrev_740ec26a1828f01ba2b83d2581d9a761` `chunk_id=srcchunk_aa591eee8d456bbe60341fa2ef371c51` `native_locator=https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79` `source_timestamp=2026-05-05T06:38:10Z`
- The full list of 20 faucet wallets includes addresses such as 0xe0Ae4dD1d7FF8C8A5E8Ad91591dfc4563931b736, 0x0cbcbBA6781B8085A8e65CE478F956b4a3923baD, 0x90604F48b6F14e08a30cCa7bE87387605f86DE6C, and others. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79) `source_document_id=srcdoc_a8afea7c5307d37ca95c6dfe019a94fb` `source_revision_id=srcrev_740ec26a1828f01ba2b83d2581d9a761` `chunk_id=srcchunk_aa591eee8d456bbe60341fa2ef371c51` `native_locator=https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79` `source_timestamp=2026-05-05T06:38:10Z`
- There is a potential Supabase issue to be aware of. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79) `source_document_id=srcdoc_a8afea7c5307d37ca95c6dfe019a94fb` `source_revision_id=srcrev_740ec26a1828f01ba2b83d2581d9a761` `chunk_id=srcchunk_aa591eee8d456bbe60341fa2ef371c51` `native_locator=https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79` `source_timestamp=2026-05-05T06:38:10Z`

## Open Questions

- How are the 20 wallets funded and managed?
- What are the exact steps to resolve the Supabase issue?
- What specific Supabase issue is referenced?

## Sources

- `source_document_id`: `srcdoc_a8afea7c5307d37ca95c6dfe019a94fb`
- `source_revision_id`: `srcrev_740ec26a1828f01ba2b83d2581d9a761`
- `source_url`: [Notion source](https://www.notion.so/faucet-app-runbook-8c9ef7af16054154a42c887e63179b79)
