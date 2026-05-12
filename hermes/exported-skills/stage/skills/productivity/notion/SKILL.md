---
name: notion
description: "Story RSI Notion access through native RSI Notion tools."
version: 2.1.0
author: RSI
license: MIT
metadata:
  hermes:
    tags: [Notion, Productivity, Notes, Database, API, RSI]
    homepage: https://developers.notion.com
prerequisites:
  native_tools: [rsi_notion]
---

# Notion

Use RSI native Notion tools for all Notion reads and writes. Do not call the
Notion HTTP API directly from the shell.

## Canonical Access Path

The canonical path is the RSI platform native Notion gateway:

- `rsi_notion.search` / transport name `rsi_notion_search`
- `rsi_notion.page_get`
- `rsi_notion.blocks_children`
- `rsi_notion.database_get`
- `rsi_notion.data_source_get`
- `rsi_notion.data_source_query`
- `rsi_notion.page_create`
- `rsi_notion.page_update`
- `rsi_notion.page_archive`
- `rsi_notion.blocks_append`
- `rsi_notion.block_update`
- `rsi_notion.block_delete`
- `rsi_notion.comment_create`

The Notion token is held server-side by the RSI control plane. Hermes should not
expect `NOTION_API_KEY`, `NOTION_TOKEN`, or any Notion secret in the shell
environment.

## Hard Rules

- Do not use `curl https://api.notion.com/...`.
- Do not ask the user to put `NOTION_API_KEY` in `~/.hermes/.env`.
- Do not print, inspect, or request Notion tokens.
- Do not conclude that Notion writes are unavailable just because
  `NOTION_API_KEY` is missing from the executor shell.
- If a native Notion tool returns a setup error such as `NOTION_TOKEN is
  required`, report it as RSI control-plane configuration drift.
- Pages and databases still must be shared with the RSI Notion integration in
  Notion, or Notion may return not found/forbidden through the native tool.

## Choosing Tools

Use `rsi_knowledge_search` for compiled company knowledge when the task is only
asking for background context from mirrored Notion/wiki content.

Use `rsi_notion.*` when the task needs live Notion objects, exact page/block
content, database/data-source queries, comments, or any Notion write/update.

## Read Workflow

1. Search by title or keywords with `rsi_notion.search`.
2. Retrieve pages with `rsi_notion.page_get`.
3. Retrieve page content with `rsi_notion.blocks_children`.
4. For databases, use `rsi_notion.database_get` to inspect metadata and
   `rsi_notion.data_source_get` / `rsi_notion.data_source_query` for rows.
5. **For tables:** `blocks_children` on the page returns a `type: "table"` block
   with metadata only (column count, header flags). The actual rows live as
   children of that table block. Make a **second** `blocks_children` call using
   the table block's ID to get `table_row` blocks with the cell data.
   See `references/block-types.md` for the full two-level traversal pattern.

## Write Workflow

1. Resolve the target page, database, or data source with native reads first.
2. Confirm the intended parent, block, or property names from live metadata.
3. Use the appropriate native write tool:
   - `rsi_notion.page_create` for new pages.
   - `rsi_notion.page_update` for property or page metadata changes.
   - `rsi_notion.blocks_append` for adding page content.
   - `rsi_notion.block_update` for editing one block.
   - `rsi_notion.block_delete` for deleting one block.
   - `rsi_notion.comment_create` for page/block comments.
4. Include a concise reason and a stable idempotency key when the tool schema
   asks for action metadata.

Prefer small, targeted writes. For larger documents, create or update a page in
sections so failures are easy to diagnose and retries are idempotent.

## Pitfalls

- **Tables need two `blocks_children` calls.** The page-level call returns a
  `table` block with metadata only. You must call `blocks_children` again on the
  table block's ID to get `table_row` blocks with actual cell data. Budget two
  sequential calls when you expect a table. See `references/block-types.md`.

- **`page_archive` returning `400 validation_error: body.archived should be not
  present`:** known RSI gateway bug (2026-05-11). Use `block_delete` as a partial
  workaround, or manually archive in the Notion UI.

## Common Diagnostics

- `object_not_found`, `not_found`, or `forbidden`: the page/database may not be
  shared with the RSI Notion integration.
- `NOTION_TOKEN is required for native Notion tools`: the control-plane pod is
  missing its server-side token configuration; this is an operator issue, not a
  Hermes executor issue.
- Rate limits or transient 5xx responses: retry later with the same idempotency
  key for writes when available.

## Deprecated Paths

Legacy Notion shell/curl instructions are retired in RSI. The Hermes executor is
not supposed to have a direct Notion token, and direct Notion API calls bypass
RSI policy, audit, idempotency, and redaction.
