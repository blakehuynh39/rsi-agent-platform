---
title: "RSI Client Features (v1.7.0 Feature Branch)"
type: "system"
slug: "systems/client-features-v1-7-0-to-feature"
freshness: "2026-06-12T09:30:00Z"
tags:
  - "client"
  - "dkg"
  - "feature"
owners: []
source_revision_ids:
  - "srcrev_d75810da24d8c83ee308d894f82eb4df"
conflict_state: "none"
---

# RSI Client Features (v1.7.0 Feature Branch)

## Summary

Changes from base v1.7.0 to Feature branch as recorded in the CDR Story Client Change Log.

## Claims

- Per-item retry cap for deals, responses, justifications, and decrypt requests: items that repeatedly fail kernel processing are dropped after a fixed cap instead of retried forever; only the kernel-reported failed indexes are requeued, successful items are left alone. (Merged) `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_d75810da24d8c83ee308d894f82eb4df` `chunk_id=srcchunk_4577457043f4e38eb989f782a208192d` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-12T09:30:00Z`
- Stops a successor session from draining the active round's decrypt queue: anchors the stale-drain heuristic on the latest finalized round instead of max(round), so a stuck or in-progress successor round can no longer drain a still-serving round's queue. (Merged) `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_d75810da24d8c83ee308d894f82eb4df` `chunk_id=srcchunk_4577457043f4e38eb989f782a208192d` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-12T09:30:00Z`
- Makes SGXValidationHook use Automata's standard TCB verification (parameterless verifyAndAttestOnChain), so Intel-defined TCB transitions (e.g. v18 to v19) take effect automatically with no owner action. Storage-layout/ABI compatible with the Aeneid proxy — in-place implementation swap, no state migration. (Merged) `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_d75810da24d8c83ee308d894f82eb4df` `chunk_id=srcchunk_4577457043f4e38eb989f782a208192d` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-12T09:30:00Z`
- Adds TDXValidationHook — on-chain validation of TDX V4/V5 attestation quotes used by the DKG flow. Mirrors SGXValidationHook (Ownable2Step + Pausable) with TDX-specific quote field offsets and an RTMR-bound identity model, enabling TDX TEE support. (Merged, but not regression) `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_d75810da24d8c83ee308d894f82eb4df` `chunk_id=srcchunk_4577457043f4e38eb989f782a208192d` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-12T09:30:00Z`
- Bumps cosmos iavl v1.2.2 to v1.2.5 (backports the #832 fix onto dkg/dev), fixing the statesync IAVL Import race that could corrupt imported snapshots. (Merged) `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_d75810da24d8c83ee308d894f82eb4df` `chunk_id=srcchunk_4577457043f4e38eb989f782a208192d` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-12T09:30:00Z`
- At finalization, verifies each finalized validator's public key share against the consensus public coefficients (vss.VerifyPublicKeyShare) and invalidates off-polynomial shares, so a validator whose dealing-phase view diverged cannot produce partial decryptions inconsistent with the committee. Skips the round if no consensus coefficients were set. It will take effect only after upgrading to V190. (Regressing) `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_d75810da24d8c83ee308d894f82eb4df` `chunk_id=srcchunk_4577457043f4e38eb989f782a208192d` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-12T09:30:00Z`
- Records which dealers submitted a deal while processing vote extensions (ProcessDeals), then invalidates verified dealers with no recorded deal at BeginFinalization — covering the 'no deal at all' case the VSS complaint path misses. The dealt set is pruned once consumed. It will take effect only after upgrading to V190. (Regressing) `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_d75810da24d8c83ee308d894f82eb4df` `chunk_id=srcchunk_4577457043f4e38eb989f782a208192d` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-12T09:30:00Z`
- Adds a registration-status check to the partial-decryption handler, rejecting partials from validators whose registration is Invalidated for the round (which would otherwise break threshold decryption). It will take effect only after upgrading to V190. (Regressing) `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_d75810da24d8c83ee308d894f82eb4df` `chunk_id=srcchunk_4577457043f4e38eb989f782a208192d` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-12T09:30:00Z`

## Sources

- `source_document_id`: `srcdoc_aeaa4036cd00e991702368cf4742be2f`
- `source_revision_id`: `srcrev_d75810da24d8c83ee308d894f82eb4df`
- `source_url`: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be)
