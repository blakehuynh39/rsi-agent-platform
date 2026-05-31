# CI Pipeline Checks — depin-backend & numo-monorepo

Quick reference for what to check in CI when reviewing PRs or doing project status.

## depin-backend (piplabs/depin-backend)

| Check | What it runs | Typical duration |
|-------|-------------|-----------------|
| Rust Checks (api) | `cargo check` / `cargo clippy` on the API crate | ~5m |
| Rust Checks (ip-registration) | Same for IP registration worker | ~2m |
| Rust Checks (wallet-management) | Same for wallet management worker | ~1m |
| PR Image Build (depin-backend-api) | Docker build for API image | ~50s |
| PR Image Build (ip-registration) | Docker build for IP registration | ~32s |
| Validate migrations | Checks migration numbering and idempotency | ~8s |
| Wiz Data Scanner | Data exposure scanning | ~18s |
| Wiz IaC Scanner | Infrastructure-as-code scanning | ~18s |
| Wiz SAST Scanner | Static application security testing | ~18s |
| Wiz Secret Scanner | Hardcoded secret detection | ~18s |
| Wiz Vulnerability Scanner | Dependency vulnerability scanning | ~18s |

**Red flags:**
- Any Wiz scanner failure (security)
- Validate migrations failure (broken DDL)
- Rust Checks failure on the api crate (most surface area)

## numo-monorepo (piplabs/numo-monorepo)

| Check | What it runs | Notes |
|-------|-------------|-------|
| Vercel – numo-monorepo-admin | Admin app preview deploy | Critical — the admin app |
| Vercel – numo-monorepo-web | Web app preview deploy | |
| Vercel – numo-landing | Landing page preview deploy | |
| Wiz Data Scanner | Data exposure scanning | ~25s |
| Wiz IaC Scanner | Infrastructure-as-code scanning | ~25s |
| Wiz SAST Scanner | Static analysis | ~25s |
| Wiz Secret Scanner | Secret detection | ~24s |
| Wiz Vulnerability Scanner | Dependency scanning | ~25s |

**Red flags:**
- Admin Vercel deploy failure (the main operator surface)
- Any Wiz scanner failure
- pnpm-lock.yaml merge conflicts (indicates dependency churn)

## Automated reviews

Both repos have **Cursor Bugbot** enabled — it posts a review comment on every commit. Check via:

```bash
gh pr view <N> --repo piplabs/<repo> --json reviews --jq '.reviews[] | select(.author.login=="cursor") | .body'
```

Bugbot typically finds:
- `useMemo` producing stale audit timestamps
- NULL vs 0 confusion in truthiness checks (`value ?` vs `value != null`)
- Navigation using `window.location.href` instead of router
- Unused exported functions
- Burst count / numeric edge cases with `NULLIF`
