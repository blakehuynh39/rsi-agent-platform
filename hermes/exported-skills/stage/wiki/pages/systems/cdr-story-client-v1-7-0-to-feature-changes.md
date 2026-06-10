---
title: "CDR Story Client: v1.7.0 to Feature Changes"
type: "system"
slug: "systems/cdr-story-client-v1-7-0-to-feature-changes"
freshness: "2026-06-10T08:39:00Z"
tags:
  - "changelog"
  - "dkg"
  - "feature"
owners: []
source_revision_ids:
  - "srcrev_e268f384f9c3d4dfdc613992cedb2d75"
conflict_state: "none"
---

# CDR Story Client: v1.7.0 to Feature Changes

## Summary

Summary of changes merged or in progress from base v1.7.0 to a feature branch, including retry caps, decrypt queue draining fix, SGX/TDX validation hooks, IAVL bump, and DKG integrity checks.

## Claims

- Per-item retry cap for deals, responses, justifications, and decrypt requests: items that repeatedly fail kernel processing are dropped after a fixed cap instead of retried forever; only the kernel-reported failed indexes are requeued, successful items are left alone. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_e268f384f9c3d4dfdc613992cedb2d75` `chunk_id=srcchunk_0c9103d5b0391d36eec53f256a943f7a` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-10T08:39:00Z`
- Stops a successor session from draining the active round's decrypt queue: anchors the stale-drain heuristic on the latest finalized round instead of max(round), so a stuck or in-progress successor round can no longer drain a still-serving round's queue. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_e268f384f9c3d4dfdc613992cedb2d75` `chunk_id=srcchunk_0c9103d5b0391d36eec53f256a943f7a` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-10T08:39:00Z`
- Makes SGXValidationHook use Automata's standard TCB verification (parameterless verifyAndAttestOnChain), so Intel-defined TCB transitions (e.g. v18 to v19) take effect automatically with no owner action. Storage-layout/ABI compatible with the Aeneid proxy — in-place implementation swap, no state migration. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_e268f384f9c3d4dfdc613992cedb2d75` `chunk_id=srcchunk_0c9103d5b0391d36eec53f256a943f7a` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-10T08:39:00Z`
- Adds TDXValidationHook — on-chain validation of TDX V4/V5 attestation quotes used by the DKG flow. Mirrors SGXValidationHook (Ownable2Step + Pausable) with TDX-specific quote field offsets and an RTMR-bound identity model, enabling TDX TEE support. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_e268f384f9c3d4dfdc613992cedb2d75` `chunk_id=srcchunk_0c9103d5b0391d36eec53f256a943f7a` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-10T08:39:00Z`
- Bumps cosmos iavl v1.2.2 to v1.2.5 (backports the #832 fix onto dkg/dev), fixing the statesync IAVL Import race that could corrupt imported snapshots. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_e268f384f9c3d4dfdc613992cedb2d75` `chunk_id=srcchunk_0c9103d5b0391d36eec53f256a943f7a` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-10T08:39:00Z`
- At finalization, verifies each finalized validator's public key share against the consensus public coefficients (vss.VerifyPublicKeyShare) and invalidates off-polynomial shares, so a validator whose dealing-phase view diverged cannot produce partial decryptions inconsistent with the committee. Skips the round if no consensus coefficients were set. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_e268f384f9c3d4dfdc613992cedb2d75` `chunk_id=srcchunk_0c9103d5b0391d36eec53f256a943f7a` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-10T08:39:00Z`
- Records which dealers submitted a deal while processing vote extensions (ProcessDeals), then invalidates verified dealers with no recorded deal at BeginFinalization — covering the 'no deal at all' case the VSS complaint path misses. The dealt set is pruned once consumed. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_e268f384f9c3d4dfdc613992cedb2d75` `chunk_id=srcchunk_0c9103d5b0391d36eec53f256a943f7a` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-10T08:39:00Z`
- Adds a registration-status check to the partial-decryption handler, rejecting partials from validators whose registration is Invalidated for the round (which would otherwise break threshold decryption). `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be) `source_document_id=srcdoc_aeaa4036cd00e991702368cf4742be2f` `source_revision_id=srcrev_e268f384f9c3d4dfdc613992cedb2d75` `chunk_id=srcchunk_0c9103d5b0391d36eec53f256a943f7a` `native_locator=https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be` `source_timestamp=2026-06-10T08:39:00Z`

## Sources

- `source_document_id`: `srcdoc_aeaa4036cd00e991702368cf4742be2f`
- `source_revision_id`: `srcrev_e268f384f9c3d4dfdc613992cedb2d75`
- `source_url`: [source](https://app.notion.com/p/CDR-Story-Client-Change-Log-37b051299a548058956cc6fd4e2995be)
