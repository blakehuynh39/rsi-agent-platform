# State Machine Debugging

When a bug involves a state machine (job pipeline, transaction lifecycle, workflow FSM), audit three dimensions before proposing fixes. This checklist emerged from debugging the depin-backend IP registration nonce drift.

## Three-Dimension Audit

### 1. Missing States

Check whether the state enum covers every meaningful intermediate condition. Common gaps:
- **"Committed but not yet acted upon"**: DB has recorded intent, but the actual action (broadcast, send, execute) hasn't happened yet. Example: `send_pending` covers "reserved but nonce NOT advanced", but no state covers "nonce advanced in DB, broadcast pending."
- **"Action started but not confirmed"**: Action was initiated but we don't know the outcome. Example: `awaiting_verify` covers broadcast-uncertain, but only when nonce was NOT advanced.

**Litmus test**: Draw the happy path. For each transition, ask: "what if we crash right here?" If recovery can't distinguish two post-crash DB states without additional context, a state is likely missing.

### 2. Transition Function Gaps

For each combination of (current_state × event), verify:
- Every reachable combination has a defined transition
- The transition produces consistent side effects (nonce advancement, lease release, job state flip)
- The same-nonce retry path is explicit (not accidental)

**Common pitfall**: `NonceUnused` events firing on states where the nonce might actually be consumed but the tx hash just hasn't been recovered yet. Add grace windows for hashless attempts.

### 3. Event Ordering (DB Commit vs. External Action)

When a state machine bridges a database and an external system (blockchain, message queue, API call), the order of "commit DB" vs "perform external action" determines which crash scenarios you handle and which you can't.

**Pattern A: External Action → DB Commit** (external-first)
- Crash before action: nothing happened, safe
- Crash after action, before commit: action happened but DB doesn't know → recovery must rediscover
- Crash inside commit: worst case — partial DB write + action already happened

**Pattern B: DB Commit → External Action** (DB-first)
- Crash after commit, before action: DB says action happened, but it didn't → need recovery that can detect "committed but not done"
- Requires a state that distinguishes "committed, pending action" from "action completed"
- Idempotency: before re-acting, check if external system already saw the action (mempool scan, idempotency key, version check)

**Which is better?** DB-first generally preferred when:
- The DB is your source of truth for what "should" have happened
- You can query the external system for idempotency (e.g., "is this nonce already pending?")
- The recovery for DB-first (committed-but-not-done) is simpler than recovery for external-first (done-but-not-committed)

## Investigation Workflow

### Step 1: Map the State Enum
Read `domain.rs` (or equivalent) and list every state variant. Note what each one means about nonce/wallet/lease state.

### Step 2: Map the Transition Table
Read `reducer.rs` (or equivalent). Build a table: rows = current states, columns = events. Mark each cell as defined or missing.

### Step 3: Trace the Pipeline
Read the caller code (submitter, confirmer, etc.). Draw the sequence:
```
Reserve (tx1) → [external action] → Commit/Abort (tx2)
```
Mark where crashes can occur and what recovery each crash point triggers.

### Step 4: Audit the Recovery Paths
For each crash point, trace the handoff/recovery code. Verify:
- Recovery correctly identifies what happened (or didn't)
- No infinite loops (grace windows, max attempts, discard caps)
- No nonce conflicts (same-nonce retries are explicit and guarded)

## Concrete Example: depin-backend IP Registration

Originally debugged 2026-05-07. The nonce drift alert signaled `pending_nonce > next_nonce` on active wallets — meaning an on-chain tx landed without DB awareness.

**Root cause cluster:**
1. **Missing state**: No `NonceCommitted`/`PendingBroadcast` state for "next_nonce advanced, broadcast not yet done"
2. **Ordering**: Broadcast (external action) happens before DB commit, creating crash window where chain consumed nonce but DB didn't record it
3. **Recovery complexity**: The confirmer's `BroadcastUncertain` → `find_mined_transaction_by_sender_and_nonce` recovery path works but has latency window where tx is in mempool but not yet mined

**Proposed fix**: Add `NonceCommitted` state, reorder to commit-before-broadcast, add mempool idempotency check.

See `apps/ip-registration/src/domain.rs` (TxAttemptState), `reducer.rs` (reduce_tx_attempt), `submitter.rs` (process_job pipeline), and `repository.rs` (commit_submit_attempt, abort_send_pending_attempt) for the code.
