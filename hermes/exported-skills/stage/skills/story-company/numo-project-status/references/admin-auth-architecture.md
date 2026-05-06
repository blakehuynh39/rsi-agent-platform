# Admin Auth & Access Control Architecture

Current state as of 2026-05-06. Documented during investigation of adding owner/admin/member role tiers to the Numo admin portal.

## depin-backend (Rust API)

### Auth Mechanism
- **No traditional middleware** — uses Axum extractors (`FromRequestParts` implementations) per-route
- Extractors live in `apps/api/src/http/extractors.rs`
- Two tiers:
  - `AdminAccess` — JWT + email allowlist + 1hr freshness + revocation check. Used on all mutating admin endpoints (POST/PATCH/DELETE)
  - `AdminReadOnlyAccess` — either `AdminAccess` (JWT path) **or** `X-Internal-Admin-Read-Api-Key` header (shared secret for service-to-service bots). Used on read-only GET endpoints
- `InternalAdminReadApiKeySession` — constant-time SHA-256 comparison against `settings.internal_admin_read_api_key`. Returns 503 when not configured

### JWT System (`domain/security.rs`)
- HS256 symmetric signing
- All users (regular + admin) share same `TokenRole::User` — **no admin role in JWT claims**
- Admin status determined entirely by checking JWT's `email` claim against hardcoded Rust constant
- `verify_current_admin_access()` also checks: token `iat` within 1hr (`ADMIN_TOKEN_MAX_AGE_SECONDS = 3600`), user not banned in DB, email not revoked via `admin_token_revocations`

### Admin Allowlist
- **Hardcoded Rust constant** `ADMIN_ALLOWLIST_EMAILS` (~13 specific `@piplabs.xyz` emails) in `extractors.rs` lines 21-35
- Adding/removing an admin requires **code deployment**
- Binary gate: every allowlisted person has identical, full admin access

### Database Tables (migrations)
| Migration | Table | Purpose |
|-----------|-------|---------|
| 0026 | `admin_audit_log` | Audit trail: `actor_email`, `action`, `resource_type`, `resource_id`, `note`, `created_at` |
| 0031 | `users.admin_notes` | Single TEXT column on `users` table for admin notes |
| 0058 | `admin_token_revocations` | One row per email: `tokens_valid_after` timestamp invalidates all older tokens |

**NOTABLY ABSENT**: No `admin_roles`, `admin_user_roles`, `admin_permissions`, or any role-related tables exist.

### Admin Route Structure
All under `/v1/admin/`:
- `/admin/campaigns` — CRUD + metrics/funnel/image-upload
- `/admin/scripts` — CRUD
- `/admin/ip-registration/*` — wallet seed/wipe/sync, tx-attempts, job detail
- `/admin/submissions` — list, detail, review (PATCH)
- `/admin/users` — list, bulk-update, detail, ban, multipliers, poseidon, referrals, castle events/profile
- `/admin/overview` — dashboard KPIs
- `/admin/stats/*` — user growth, submissions
- `/admin/cohorts/*` — retention, languages, demographics
- `/admin/referrals` + commissions
- `/admin/castle/*` — proxy for Castle.io
- `/admin/multiplier-tasks` — catalog CRUD
- `/admin/reward-config`, rewards metrics, balances reconciliation

### Auth Flow
1. Admin authenticates same as regular user: `POST /v1/auth/exchange` (Dynamic) or `/v1/auth/world/exchange` (World Protocol)
2. Backend JWT issued with email claim
3. Admin status inferred from email on each request
4. No separate admin login endpoint

## numo-monorepo (Frontend Admin App)

### Login / Authentication
- **Provider**: Dynamic Labs SDK (`@dynamic-labs/sdk-react-core`) — email OTP only
- `VITE_DYNAMIC_ENVIRONMENT_ID` env var controls Dynamic widget
- Token exchange: Dynamic JWT → `POST /v1/auth/exchange` → backend JWT
- Token stored in `localStorage` under `numo-admin:backend-token` + expiry key
- Dev modes: `VITE_DEV_SKIP_AUTH=true` (bypass), `VITE_DEV_MOCK_API=true` (mock JWT)

### Session Management
- `apps/admin/src/components/session-watcher.tsx` — schedules timeout at expiry, watches visibility/focus/storage events, handles silent refresh
- `apps/admin/src/lib/api-client.ts` — attaches `Authorization: Bearer <token>` on every request, fires `AUTH_EXPIRED_EVENT` on 401

### RBAC State — **NONE EXISTS**
- No role types, permission types, or access-level concepts anywhere in the admin app or monorepo
- The word "role" only appears in unrelated contexts (voice agent message roles, plan document metadata)
- No utilities for role checking, permission gating, or policy evaluation

### Email Allowlist (binary gate — UNWIRED)
- `VITE_ADMIN_ALLOWLIST` env var exists with comma-separated emails + `*@domain.xyz` wildcards
- `helpers/allowlist.ts` — `parseAllowlist()` and `isEmailAllowed()` are defined and tested
- **NOT wired into any route, component, or layout** — no code imports or calls these functions
- The `/forbidden` route exists but is never redirected to
- Server-side enforcement only (backend `AdminAccess` extractor)

### Route Guard
- `apps/admin/src/routes/_authenticated.tsx` — only checks `isAuthenticated()` (token not expired)
- No role or permission checks
- Natural place to add role-based redirects

### Nav Configuration
- `apps/admin/src/lib/nav-config.ts` — flat list, all items visible to all authenticated users
- No visibility filtering based on roles

### User Types (no role fields)
- `AdminUserSummary` / `AdminUserDetail` / `AdminUpdateUserRequest` in `apps/admin/src/lib/types.ts`
- `QualityLabel`: `"suspicious" | "watch" | "neutral" | "trusted"` — user quality tier, not admin role
- Shared `api-types` package has minimal user types

## Notion / Backlog
- **APP-17 "Admin - Access Control"** (2024-03-11): Status "Not started", Milestone "Public Testnet", Category "Admin tools" — skeleton card, no content
- **"Admin API Security Architecture"** (2023-02-22): Describes KMS-based encrypted message approach (pre-dates current JWT+Dynamic auth). Notes JWT as "alternative" for future if admin API goes public
- No other current admin access control docs found

## Gap Summary for Implementing Role Tiers

To add owner/admin/member roles, changes needed beyond Rust code:
1. **DB migration**: New `admin_roles` + `admin_user_roles` tables (or `role` column on admin users table)
2. **FE changes**: Route gating, nav filtering, conditional UI, member management page, wire allowlist
3. **Design decision**: Keep hardcoded allowlist as "who can admin" gate + DB roles on top, OR move fully to DB
4. **K8s/Infra**: No new services needed (existing depin-backend serves admin endpoints)
5. **Auth provider**: Dynamic stays, no changes needed
