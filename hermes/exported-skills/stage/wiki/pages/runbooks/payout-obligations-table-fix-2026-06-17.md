---
title: "Payout Obligations Table Export Fix (2026-06-17)"
type: "runbook"
slug: "runbooks/payout-obligations-table-fix-2026-06-17"
freshness: "2026-06-17T19:22:02Z"
tags:
  - "admin dashboard"
  - "export"
  - "fix"
  - "payout obligations"
owners:
  - "U08951K4SRY"
  - "U0ASDQKU3UL"
source_revision_ids:
  - "srcrev_0438af46758aaa2ad4e2576c7f905ce5"
  - "srcrev_2b4ef026b8f225ab814467d3e921a96e"
  - "srcrev_44606dfa9dc07f468f7a0f656f744ee5"
  - "srcrev_8e87d144c679c5068eda759be4693cc0"
  - "srcrev_98e49213906e110d3017d6f10c04d5d9"
conflict_state: "none"
---

# Payout Obligations Table Export Fix (2026-06-17)

## Summary

Fixes for masked email export and UI flickering in the admin dashboard payout obligations table, including workaround instructions and PR details.

## Claims

- The export from the payout obligations table was returning masked emails instead of real emails. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2260623eeb0bba6c00f3086b35a55cb7` `source_revision_id=srcrev_8e87d144c679c5068eda759be4693cc0` `chunk_id=srcchunk_a41ccd27d7c0c086aeaea2aa046591f7` `native_locator=slack:C0AL7EKNHDF:1781662768.497329:1781682036.402269` `source_timestamp=2026-06-17T07:40:36Z`
- The UI table started to flicker non-stop when scrolling down. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2260623eeb0bba6c00f3086b35a55cb7` `source_revision_id=srcrev_8e87d144c679c5068eda759be4693cc0` `chunk_id=srcchunk_a41ccd27d7c0c086aeaea2aa046591f7` `native_locator=slack:C0AL7EKNHDF:1781662768.497329:1781682036.402269` `source_timestamp=2026-06-17T07:40:36Z`
- Feature requests included ability to add custom tags on users for searching/filtering and multi-select fields in filters. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2260623eeb0bba6c00f3086b35a55cb7` `source_revision_id=srcrev_8e87d144c679c5068eda759be4693cc0` `chunk_id=srcchunk_a41ccd27d7c0c086aeaea2aa046591f7` `native_locator=slack:C0AL7EKNHDF:1781662768.497329:1781682036.402269` `source_timestamp=2026-06-17T07:40:36Z`
- Two PRs were created to address the real email export, scroll flickering, and multi-select filters. Custom user tags require a separate planned effort. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2260623eeb0bba6c00f3086b35a55cb7` `source_revision_id=srcrev_98e49213906e110d3017d6f10c04d5d9` `chunk_id=srcchunk_d51666a79bfd82cb5b0b2062fc621e55` `native_locator=slack:C0AL7EKNHDF:1781662768.497329:1781722691.709449` `source_timestamp=2026-06-17T18:58:11Z`
- The backend PR (depin-backend#562) had a Rust compilation error (borrow-after-move) in the AdminUserSummary From impl. Fixed by computing email_masked before moving email. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2260623eeb0bba6c00f3086b35a55cb7` `source_revision_id=srcrev_2b4ef026b8f225ab814467d3e921a96e` `chunk_id=srcchunk_55d76ce959a4a65665af956ed60796b8` `native_locator=slack:C0AL7EKNHDF:1781662768.497329:1781723775.913739` `source_timestamp=2026-06-17T19:16:15Z`
- A workaround was provided to extract user data via browser DevTools network tab and console script, retrieving all pages of eligible users. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2260623eeb0bba6c00f3086b35a55cb7` `source_revision_id=srcrev_0438af46758aaa2ad4e2576c7f905ce5` `chunk_id=srcchunk_84c944385d81ec22658158b9e54b624a` `native_locator=slack:C0AL7EKNHDF:1781662768.497329:1781722922.286019` `source_timestamp=2026-06-17T19:02:02Z`
- Release PRs were opened for depin-backend (staging→main) and numo-monorepo (develop→main) with 2 commits each. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2260623eeb0bba6c00f3086b35a55cb7` `source_revision_id=srcrev_44606dfa9dc07f468f7a0f656f744ee5` `chunk_id=srcchunk_7e4b8f477758cbc04d506ee9cb744b8e` `native_locator=slack:C0AL7EKNHDF:1781662768.497329:1781724122.467519` `source_timestamp=2026-06-17T19:22:02Z`

## Sources

- `source_document_id`: `srcdoc_2260623eeb0bba6c00f3086b35a55cb7`
- `source_revision_id`: `srcrev_bf1e12a56bf90db178196bf2f06fff02`
