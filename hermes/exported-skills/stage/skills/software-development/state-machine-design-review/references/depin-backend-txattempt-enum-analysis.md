# depin-backend TxAttemptState Enum Analysis

**Date:** 2026-05-07
**Repo:** piplabs/depin-backend (staging branch)
**Files:** `apps/ip-registration/src/domain.rs` (enum), `apps/ip-registration/src/reducer.rs` (transitions), `docs/architecture/ip-registration.md` (design doc), `docs/plans/active/2026-04-22-ip-registration-send-pending-state.md` (history), `docs/plans/active/2026-04-28-ip-registration-nonce-race-and-revert-cascade.md` (bug fix plan)

## The Enum

```rust
workflow_enum!(TxAttemptState, "tx-attempt state", {
    SendPending => "send_pending",       // nonce reserved, NOT advanced, broadcast not attempted
    Sending => "sending",                 // LEGACY — deprecated, kept for back-compat
    Submitted => "submitted",             // broadcast accepted, hash known, awaiting receipt
    Bumping => "bumping",                 // replacement scheduled for stale pending tx
    AwaitingVerify => "awaiting_verify",  // broadcast outcome uncertain, confirmer must reconcile
    Confirmed => "confirmed",            // terminal: receipt success
    Failed => "failed",                  // terminal: receipt reverted or unrecoverable
    TimedOut => "timed_out",             // terminal: nonce not consumed on chain
    Skipped => "skipped",                // terminal: superseded by replacement
    Discarded => "discarded",            // terminal: pre-broadcast revert (Patch 3b)
});
```

## Identified Orthogonal Dimensions

The 10 variants conflate 4 independent dimensions:

1. **Nonce bookkeeping** (2 values):
   - `Reserved` — `next_nonce` NOT advanced in DB (`SendPending`, `Sending`, `AwaitingVerify`, `Discarded`)
   - `Committed` — `next_nonce` advanced in DB (`Submitted`, `Bumping`, `Confirmed`, `Failed`, `TimedOut`, `Skipped`)

2. **Broadcast outcome** (3 values):
   - `NotAttempted` — nothing sent to chain (`SendPending`, `Discarded`)
   - `Accepted` — RPC returned Ok with tx hash (`Submitted`, `Bumping`, `Confirmed`, `Failed`, `Skipped`)
   - `Uncertain` — RPC failed, outcome unknown (`AwaitingVerify`)

3. **Receipt outcome** (3 values):
   - `Pending` — waiting for receipt (`Submitted`, `Bumping`)
   - `Confirmed` — receipt success (`Confirmed`)
   - `Failed` — receipt reverted (`Failed`)

4. **Replacement mode** (3 values):
   - `Normal` — no replacement in progress (most states)
   - `Bumping` — replacement scheduled (`Bumping`)
   - `Superseded` — replaced by newer attempt (`Skipped`)

## Bug Evidence Mapping

### Bug A: Nonce-drift alert (ongoing)
- **Symptom:** `chain.pending_nonce > wallets.next_nonce` on active wallet → Sentry alert
- **Root cause:** `AwaitingVerify` conflates "nonce NOT consumed" (should retry same nonce) with "nonce WAS consumed" (should recover hash). The confirmer at `confirmer.rs:516-552` must call `chain.pending_nonce()` at runtime to re-derive which scenario applies.
- **Fix:** None yet — runtime chain query is the workaround. Would be type-level impossible with separate `NonceState::Reserved|Committed` + `BroadcastOutcome::Uncertain`.

### Bug B: Revert cascade (2026-04-28)
- **Symptom:** 11 jobs failed in 18s window. `eth_estimateGas` returned `execution reverted` → routed through `abort_send_pending_attempt` → `BroadcastUncertain` → `AwaitingVerify` → burned `attempt_count` on phantom rows → confirmer classified `NonceUnused` → terminal fail.
- **Root cause:** `AwaitingVerify` was used for both "RPC error during broadcast (tx may have landed)" and "pre-broadcast simulation reverted (nothing was broadcast)."
- **Fix:** Patch 3b — added `Discarded` state + `discard_send_pending_attempt` repository path + `can_retry` excludes `Discarded` from budget.
- **Lesson:** This was an ad-hoc dimension split — the fix added one flat variant instead of decomposing the enum. Correct approach for the state count.

### Bug C: Ordering gap (this thread)
- **Symptom:** Cannot reverse the submitter ordering from "broadcast → commit nonce" to "commit nonce → broadcast" because no state represents "nonce committed in DB, broadcast not yet attempted."
- **Root cause:** `SendPending` specifically means "nonce reserved, `next_nonce` NOT advanced." There's no counterpart for "nonce committed, `next_nonce` advanced, broadcast pending."
- **Fix:** Needs new `NonceCommitted` (or `PendingBroadcast`) state.

## What a Product-Type Refactor Would Look Like

```rust
enum NonceState { Reserved, Committed }
enum BroadcastOutcome { NotAttempted, Accepted, Uncertain }
enum ReceiptOutcome { Pending, Confirmed, Failed }
enum ReplacementMode { Normal, Bumping, Superseded }

struct TxAttemptState {
    nonce: NonceState,
    broadcast: BroadcastOutcome,
    receipt: ReceiptOutcome,
    replacement: ReplacementMode,
}
```

Invalid combinations become type-level impossible:
- `NonceState::Reserved + ReceiptOutcome::Confirmed` — can't confirm a tx whose nonce isn't committed
- `BroadcastOutcome::NotAttempted + ReplacementMode::Superseded` — can't supersede what wasn't broadcast

Current flat variants map to product:
- `SendPending` = `(Reserved, NotAttempted, Pending, Normal)`
- `Submitted` = `(Committed, Accepted, Pending, Normal)`
- `AwaitingVerify` = `(Reserved, Uncertain, Pending, Normal)`
- `Discarded` = `(Reserved, NotAttempted, Pending, Normal)` — differentiated only by terminal flag

## Recommendation

**Immediate:** Add `NonceCommitted` state to close the ordering gap. This is one variant + three reducer arms. Gets 80% of the safety benefit for 5% of the effort.

**Long-term:** Revisit product-type decomposition when:
- Flat enum reaches ~15 active variants (currently 9 active + 1 legacy)
- A 5th orthogonal dimension is needed (e.g., `FeeBumpStrategy`)
- Another bug fires that traces to dimension conflation

**Not recommended:** Full product-type refactor now. The 10-state flat enum is well-tested, the reducer is correct, and the bugs were in callers routing wrong events — not in the reducer itself.
