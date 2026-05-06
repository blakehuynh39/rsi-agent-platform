# Cross-Repo Merge-Order Check Trace

Concrete instances of merge-order violations found during reviews.
Sharpens future reviews — patterns repeat.

## 2026-05-06: depin-backend#419 ↔ numo-monorepo#216

- **FE PR #216** (`feat/professional-multimodal-contract`) was **already merged** to `develop`
- **BE PR #419** (`feat/multimodal-submission-uploads`) was **still OPEN** (approved but not merged)
- Violation of contract: BE must merge to `staging` (and deploy) before FE merges to `develop`
- Flagged in review as non-blocking process note

## Review checklist enhancement

When a cross-repo pair is detected, always check:
- [ ] If FE is already merged, verify BE merged first. If not, flag it.
- [ ] If BE is already merged, verify FE is still open or was merged after BE.
