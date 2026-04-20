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

## Important constraint

Notion's hosted MCP requires user OAuth and is not a good fit for unattended,
platform-wide service auth by itself. For long-lived Hermes workloads, there are
two viable patterns:

1. Inject a user OAuth access token per task using `authorization` or
   `authorization_env_var`.
2. Self-host Notion's open-source MCP server with a Notion integration token,
   then point Hermes at that server instead.

The runner also supports unauthenticated remote MCP servers, so a self-hosted
Notion MCP proxy can be exposed behind your own network controls without forcing
the runner to send a bearer token.

## RSI platform envs

To have control-plane attach Notion MCP automatically to workflow and
question-gather runner tasks, set:

```bash
RSI_NOTION_MCP_ENABLED=true
RSI_NOTION_MCP_SERVER_URL=https://mcp.notion.com/mcp
RSI_NOTION_MCP_AUTHORIZATION_ENV_VAR=RSI_NOTION_MCP_AUTHORIZATION
```

Then provide the actual bearer value to runner processes via:

```bash
RSI_NOTION_MCP_AUTHORIZATION=...
```

For the hosted Notion MCP, that bearer value must be a valid user OAuth access
token for the target workspace. For the self-hosted server, it can be omitted
or replaced with whatever auth your deployment expects.
