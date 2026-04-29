# Notion MCP

This repo now supports generic remote MCP server auth on runner task payloads via
`authorization_env_var`.

## Codex

For local Codex usage, configure Notion MCP with:

```toml
[mcp_servers.notion]
url = "https://mcp.notion.com/mcp"
```

Then authenticate with:

```bash
codex mcp login notion
```

## Hermes runner

The Hermes runner can now resolve MCP auth from either:

- `authorization` on the task payload
- `authorization_env_var` on the task payload

Example task MCP server entry for Notion's hosted MCP:

```json
{
  "server_label": "notion",
  "server_url": "https://mcp.notion.com/mcp",
  "authorization_env_var": "RSI_NOTION_MCP_AUTHORIZATION",
  "allowed_tools": {
    "tool_names": ["search", "fetch"]
  }
}
```

Example task MCP server entry for a self-hosted `notion-mcp-server` protected by
`AUTH_TOKEN`:

```json
{
  "server_label": "notion",
  "server_url": "http://notion-mcp:3000/mcp",
  "authorization_env_var": "RSI_NOTION_MCP_AUTHORIZATION",
  "allowed_tools": {
    "read_only": true
  }
}
```

## Important constraint

Notion's hosted MCP requires user OAuth and is not a good fit for unattended,
platform-wide service auth by itself. For long-lived Hermes workloads, there are
two viable patterns:

1. Inject a user OAuth access token per task using `authorization` or
   `authorization_env_var`.
2. Self-host Notion's open-source MCP server with a Notion integration token,
   then point Hermes at that server instead.

The official self-hosted `notion-mcp-server` runtime contract is:

- `NOTION_TOKEN`: the Notion integration token used by the server to call the
  Notion API.
- `AUTH_TOKEN`: optional bearer token the server expects on inbound HTTP MCP
  requests.
- streamable HTTP transport on `/mcp`, with port `3000` by default.

The self-hosted `mcp/notion` server exposes Notion API tool names rather than
generic `search` / `fetch` tools. The runner's `notion_mcp_read` profile maps
read-only access to:

- `API-post-search`
- `API-retrieve-a-page`
- `API-get-block-children`

It also disables generic MCP resource and prompt utility tools for Notion,
because this server does not implement MCP resources/prompts.

The runner also supports unauthenticated remote MCP servers. If you expose a
self-hosted Notion MCP service only behind trusted network controls, you can
leave `RSI_NOTION_MCP_AUTHORIZATION_ENV_VAR` empty and RSI will omit the bearer
token entirely.

## RSI platform envs

To have control-plane attach Notion MCP automatically to workflow and
question-gather runner tasks, set:

```bash
RSI_NOTION_MCP_ENABLED=true
RSI_NOTION_MCP_SERVER_URL=http://notion-mcp:3000/mcp
RSI_NOTION_MCP_AUTHORIZATION_ENV_VAR=RSI_NOTION_MCP_AUTHORIZATION
```

For the recommended self-hosted deployment, configure the Notion MCP server with:

```bash
NOTION_TOKEN=ntn_...
AUTH_TOKEN=...
```

Then provide the same inbound bearer token to runner processes via:

```bash
RSI_NOTION_MCP_AUTHORIZATION=...
```

If the self-hosted service is intentionally unauthenticated, set
`RSI_NOTION_MCP_AUTHORIZATION_ENV_VAR=` and do not inject
`RSI_NOTION_MCP_AUTHORIZATION` into runner pods.

For the hosted Notion MCP, the runner bearer value must instead be a valid user
OAuth access token for the target workspace.
