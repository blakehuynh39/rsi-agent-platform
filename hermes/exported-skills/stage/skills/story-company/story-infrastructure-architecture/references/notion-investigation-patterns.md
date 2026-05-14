# Notion Investigation Patterns for Story Infrastructure

These patterns are useful when investigating Story protocol parameters, gas costs, contract documentation, and architectural decisions stored in Notion.

## Block Traversal Pattern

Notion's API returns pages and blocks separately. Always follow this sequence:

```
rsi_notion_search(query="...")          → find candidate pages
rsi_notion_page_get(page_id=...)         → get metadata (title, URL, created/edited)
rsi_notion_blocks_children(block_id=...) → get actual content (text, lists, tables)
```

**PITFALL:** `rsi_notion_page_get` only returns metadata (object, properties, parent). It does NOT return the page body. You must always call `rsi_notion_blocks_children` with the same page ID to read the content.

## Table Traversal

Notion tables are blocks with `type: "table"` and `has_children: true`. The table block itself has no cell data — the rows are children:

```
rsi_notion_blocks_children(block_id=<page_id>)     → returns top-level blocks including table block
rsi_notion_blocks_children(block_id=<table_block_id>) → returns table_row blocks with cell data
```

Table rows have `type: "table_row"` and `type_payload.cells` — a 2D array of rich text arrays. Each cell is an array of text annotations.

## Content Access Patterns

| Goal | Tool Sequence |
|---|---|
| Find pages by topic | `rsi_notion_search(query=...)` |
| Get page metadata + URL | `rsi_notion_page_get(page_id=...)` |
| Read page content | `rsi_notion_blocks_children(block_id=<page_id>)` |
| Expand toggle blocks | `rsi_notion_blocks_children(block_id=<toggle_block_id>)` |
| Read table rows | Two-level: page blocks → table block → table rows |
| Read child pages | `rsi_notion_blocks_children(block_id=<child_page_block_id>)` |

## Performance Note

`rsi_notion_search` responses can be extremely large (400K-500K chars) and may be truncated in tool output. Prefer `rsi_knowledge_search` with `source_types: ["notion"]` for broad topical searches — it returns a compact manifest with page titles, URLs, and edit dates. Use `rsi_knowledge_search` for discovery, then `rsi_notion_page_get` + `rsi_notion_blocks_children` for deep reading of specific pages.

## Key Notion Pages for Story Infrastructure

| Topic | Notion Page | Page ID (last 4 segments) |
|---|---|---|
| L1 chain config benchmarks | "Other Mainnet Configs" | `1690...6ca3` |
| Gas cost table (rc.2, Sepolia) | "Gas Cost on rc.2" | `6a07...d219` |
| Gas estimates (alpha rc0) | "Gas Estimations (prelim, on alpha rc0)" | `ca4e...902a` |
| Mainnet protocol design | "Mainnet Protocol Design Discussions" | `91a3...5063` |
| Infrastructure cost planning | "L1 cost estimate" | `ebb8...09a9` |
| Public testnet guide | "Story Public Testnet Guide" | multiple versions |
| Poseidon L2 design | "Poseidon Subnet as a Rollup-based Layer2 HLD" | `2390...3e3d` |
| Subnet contract spec | "Subnet Contract Spec V1" | `2690...5d7d` |

Use these page IDs directly with `rsi_notion_page_get` + `rsi_notion_blocks_children` when you need the full content.
