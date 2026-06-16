# Cross-Repo Merge-Order Check Trace

Concrete instances of merge-order violations found during reviews.
Sharpens future reviews — patterns repeat.

## 2026-05-06: depin-backend#419 ↔ numo-monorepo#216

- **FE PR #216** (`feat/professional-multimodal-contract`) was **already merged** to `develop`
- **BE PR #419** (`feat/multimodal-submission-uploads`) was **still OPEN** (approved but not merged)
- Violation of contract: BE must merge to `staging` (and deploy) before FE merges to `develop`
- Flagged in review as non-blocking process note

## 2026-05-14: depin-backend#483 ↔ numo-monorepo#297

- **BE PR #483** (`feat/stripe-v2-global-payouts`) changed `AdminApproveWithdrawalResponse.stripe_transfer_id` from `String` to `Option<String>` — breaking API contract
- **No FE PR existed** at review time → flagged as MEDIUM: missing cross-repo pair
- **FE PR #297** (`feat/admin-stripe-v2-fields`) was opened hours later as companion, making TypeScript types optional
- Both approved in correct merge order (BE first, then FE). The gap caught early prevented a staging deploy with broken admin dashboard.
- **Lesson:** When BE makes a response field optional (was `String`, becomes `Option<String>`), the FE MUST make the corresponding TypeScript field `?: string | null`. Flag immediately if FE PR doesn't exist — this is not something that can be deferred to a follow-up.

## Review checklist enhancement

When a cross-repo pair is detected, always check:
- [ ] If FE is already merged, verify BE merged first. If not, flag it.
- [ ] If BE is already merged, verify FE is still open or was merged after BE.

## 2026-06-16: depin-backend#555 ↔ numo-monorepo#408

- **BE PR #555** (`fix/tax-status-gating`) split `payout_block_reason_for_record`'s mapping: `"opened"` moved from `"pending_verification"` to `"incomplete"`. New helper `is_tax_form_pending_verification()` created. Dropin-session now returns 409 for `submitted`/`awaiting_tin_match` forms.
- **FE PR #408** added matching `isTaxFormPendingVerification()` and `isTaxFormIncompleteInProgress()` helpers, plus `pending_verification` to `TaxBlockReason` type union. FE catches the 409 conflict and shows different UI per state.
- **Lesson 1 — block_reasons as implicit cross-repo contract:** The `block_reasons` array tokens (`pending_verification`, `incomplete`) are an implicit contract. The BE constructs `"tax_setup_incomplete:{block_reason}"` and the FE's `parseTaxSetupReason()` strips the prefix. When BE adds a new token, FE's `reasonCopyByReason` and `TaxBlockReason` type must both gain the new variant. The FE also reads `tax.status` (raw form status) independently from `block_reasons` (processed gating reason) — they serve different UI decisions.
- **Lesson 2 — shared function return-semantics change breaks distant tests:** The BE PR correctly updated unit tests in `mod.rs` but missed the integration test at `tax_form_resubmission.rs:809` that asserted `block_reason == "pending_verification"` after a dropin session creates an `opened` form. The test was 300 lines away from the function definition and in a different file. When changing what a shared function returns, search the entire repo for all callers and assertions — don't stop at the module's own tests.
- **CI masking:** The broken test wasn't caught because `cargo fmt --check` failed first (import ordering). If the fmt check stops the CI job before tests, a test regression can ship undetected.

## Review checklist enhancement

When a cross-repo pair is detected, always check:
- [ ] If FE is already merged, verify BE merged first. If not, flag it.
- [ ] If BE is already merged, verify FE is still open or was merged after BE.
- [ ] When BE changes `block_reasons` tokens or `payout_block_reason_for_record` return values, verify the FE's `parseTaxSetupReason` prefix strip + `reasonCopyByReason` dict + `TaxBlockReason` type union all cover the new/changed token.
- [ ] When BE changes a shared function's return semantics, use `search_files` repo-wide to find integration tests and distant callers that assert on the old return value.
