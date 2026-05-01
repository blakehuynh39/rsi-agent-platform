# Depin Prod Admin Read

Use this skill when a Story request asks for live Numo/depin user stats, submission stats, admin user lookup, or production depin API route context.

## Runtime Contract

- Use `DEPIN_ADMIN_BASE_URL` as the base URL. Stage RSI intentionally points this at the production public endpoint, `https://depin.storyprotocol.net`.
- Use `DEPIN_ADMIN_READ_API_KEY_HEADER` for the header name and `DEPIN_ADMIN_READ_API_KEY` for the header value.
- Never print, summarize, export, or store the credential value. Only report whether it is present.
- Treat this credential as read-only. Do not use it for write, mutation, delete, or admin management actions.

## Source Of Truth

- Runtime API contract: `GET ${DEPIN_ADMIN_BASE_URL}/openapi.json`.
- Backend semantics and route behavior: `piplabs/depin-backend`, especially `apps/api/src/http/routes/admin.rs`, `apps/api/src/services/admin_dashboard.rs`, and `docs/api-workflows.md`.
- Public DNS/WAF routing: `piplabs/cloudflare`, especially `src/zones/storyprotocol.net/records.ts` and `src/zones/storyprotocol.net/waf.ts`.
- Deployment and Vault wiring: `story-deployments`, `rsi-platform/rsi-agent-platform/use1-stage.yaml`, and `story/depin-backend/use1-prod.yaml`.

## Query Pattern

1. Fetch `/openapi.json` when route shape or parameters are uncertain.
2. For aggregate user stats, start with `/v1/admin/stats/user-growth`.
3. For aggregate submission stats, start with `/v1/admin/stats/submissions`.
4. For a specific user, resolve the user identifier from available context first, then use the documented `/v1/admin/users/**` read routes.
5. If the public endpoint returns a Cloudflare block before reaching depin, report it as a Cloudflare/WAF routing issue and check the Cloudflare SoT before guessing.
6. If depin returns `401` or `403`, report the auth failure without exposing the credential.

## Response Standard

- State whether the answer came from production live API data or from repository/OpenAPI context.
- Include endpoint paths and status codes when debugging.
- For stats, include the query time and any filters/parameters used.
- If live data is unavailable, explain the exact blocker and the next required infrastructure or API fix.
