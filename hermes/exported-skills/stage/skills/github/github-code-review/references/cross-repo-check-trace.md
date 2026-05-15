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
