# Depin Prod Admin Read

Use this skill when a Story request asks for live Numo/depin user stats, submission stats, admin user lookup, or production depin API route context.

## Runtime Contract

- Use `DEPIN_ADMIN_BASE_URL` as the base URL. Stage RSI intentionally points this at the production public endpoint, `https://depin.storyprotocol.net`.
- Use `DEPIN_ADMIN_READ_API_KEY_HEADER` for the header name and `DEPIN_ADMIN_READ_API_KEY` for the header value.
- Use `User-Agent: rsi-hermes-company-computer/1.0` and `X-RSI-Source: hermes` on direct prod depin reads so Cloudflare can distinguish the company-computer path from generic scripted traffic. These headers are not credentials; origin authorization still depends on the admin read key.
- Never print, summarize, export, or store the credential value. Only report whether it is present.
- Treat this credential as read-only. Do not use it for write, mutation, delete, or admin management actions.

## Source Of Truth

- Public API discovery: the deployed production OpenAPI document is intentionally client-facing and must not be treated as authoritative for internal admin stats routes.
- Internal admin stats contract: `piplabs/depin-backend`, especially `apps/api/src/http/routes/admin.rs`, `apps/api/src/http/extractors.rs`, `apps/api/src/services/admin_dashboard.rs`, and `docs/api-workflows.md`.
- Public DNS/WAF routing: `piplabs/cloudflare`, especially `src/zones/storyprotocol.net/records.ts` and `src/zones/storyprotocol.net/waf.ts`.
- Deployment and Vault wiring: `story-deployments`, `rsi-platform/rsi-agent-platform/use1-stage.yaml`, and `story/depin-backend/use1-prod.yaml`.
- Runtime readiness: the `story-deployments` depin admin read validation hook curls `/v1/admin/stats/user-growth` from the stage RSI cluster with the mounted read key and must fail loudly on non-200 responses.

## Query Pattern

1. Confirm `DEPIN_ADMIN_READ_API_KEY` is present without printing its value.
2. For public/client-facing route context, prefer the checked-in `piplabs/depin-backend` OpenAPI/source generation over the production OpenAPI document. Do not expect production OpenAPI to advertise internal admin stats.
3. For internal admin stats route shape, inspect the deployed `piplabs/depin-backend` source code and `story-deployments` image pin before answering from memory.
4. For aggregate user stats, call `/v1/admin/stats/user-growth` directly with the configured read-key header.
5. For aggregate submission stats, call `/v1/admin/stats/submissions` directly with the configured read-key header.
6. For a specific user, resolve the user identifier from available context first, then use source-confirmed `/v1/admin/users/**` read routes.
7. If the public endpoint returns a Cloudflare block before reaching depin, report it as a Cloudflare/WAF routing issue and check the Cloudflare SoT before guessing.
8. If depin returns `401` or `403`, distinguish these cases without exposing the credential:
   - credential env var absent
   - credential env var mounted but rejected by prod
   - request blocked before reaching depin
9. Prefer `https://depin.storyprotocol.net` for production Numo/depin stats. Do not switch to staging APIs unless the user explicitly asks for staging data.

## Response Standard

- State whether the answer came from production live API data or from repository/OpenAPI context.
- Include endpoint paths and status codes when debugging.
- For stats, include the query time and any filters/parameters used.
- If live data is unavailable, explain the exact blocker and the next required infrastructure or API fix.
- Do not ask the user for a credential when the credential is already mounted but rejected; report that the Vault/config value needs to be fixed.
