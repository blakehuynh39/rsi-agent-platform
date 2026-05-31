# Notion Block Types

Reference for creating and reading all common Notion block types via the API.

## Creating blocks

Use `PATCH /v1/blocks/{page_id}/children` with a `children` array. Each block follows this structure:

```json
{"object": "block", "type": "<type>", "<type>": { ... }}
```

### Paragraph

```json
{"type": "paragraph", "paragraph": {"rich_text": [{"text": {"content": "Hello world"}}]}}
```

### Headings

```json
{"type": "heading_1", "heading_1": {"rich_text": [{"text": {"content": "Title"}}]}}
{"type": "heading_2", "heading_2": {"rich_text": [{"text": {"content": "Section"}}]}}
{"type": "heading_3", "heading_3": {"rich_text": [{"text": {"content": "Subsection"}}]}}
```

### Bulleted list

```json
{"type": "bulleted_list_item", "bulleted_list_item": {"rich_text": [{"text": {"content": "Item"}}]}}
```

### Numbered list

```json
{"type": "numbered_list_item", "numbered_list_item": {"rich_text": [{"text": {"content": "Step 1"}}]}}
```

### To-do / checkbox

```json
{"type": "to_do", "to_do": {"rich_text": [{"text": {"content": "Task"}}], "checked": false}}
```

### Quote

```json
{"type": "quote", "quote": {"rich_text": [{"text": {"content": "Something wise"}}]}}
```

### Callout

```json
{"type": "callout", "callout": {"rich_text": [{"text": {"content": "Important note"}}], "icon": {"emoji": "💡"}}}
```

### Code

```json
{"type": "code", "code": {"rich_text": [{"text": {"content": "print('hello')"}}], "language": "python"}}
```

### Toggle

```json
{"type": "toggle", "toggle": {"rich_text": [{"text": {"content": "Click to expand"}}]}}
```

### Divider

```json
{"type": "divider", "divider": {}}
```

### Bookmark

```json
{"type": "bookmark", "bookmark": {"url": "https://example.com"}}
```

### Image (external URL)

```json
{"type": "image", "image": {"type": "external", "external": {"url": "https://example.com/photo.png"}}}
```

## Reading blocks

When reading blocks from `GET /v1/blocks/{page_id}/children`, each block has a `type` field. Extract readable text like this:

| Type | Text location | Extra fields |
|------|--------------|--------------|
| `paragraph` | `.paragraph.rich_text` | — |
| `heading_1/2/3` | `.heading_N.rich_text` | — |
| `bulleted_list_item` | `.bulleted_list_item.rich_text` | — |
| `numbered_list_item` | `.numbered_list_item.rich_text` | — |
| `to_do` | `.to_do.rich_text` | `.to_do.checked` (bool) |
| `toggle` | `.toggle.rich_text` | has children |
| `code` | `.code.rich_text` | `.code.language` |
| `quote` | `.quote.rich_text` | — |
| `callout` | `.callout.rich_text` | `.callout.icon.emoji` |
| `divider` | — | — |
| `image` | `.image.caption` | `.image.file.url` or `.image.external.url` |
| `bookmark` | `.bookmark.caption` | `.bookmark.url` |
| `child_page` | — | `.child_page.title` |
| `child_database` | — | `.child_database.title` |
| `table` | — | `.table.has_column_header`, `.table.has_row_header`, `.table.table_width`; **has children** — re-query to get rows |
| `table_row` | — | `.table_row.cells` — 2D array of rich-text arrays |

Rich text arrays contain objects with `.plain_text` — concatenate them for readable output.

### Table blocks — two-level traversal

Notion tables require **two `blocks_children` calls** to read fully:

1. **First call** on the page yields a `type: "table"` block with metadata only (`has_column_header`, `table_width`) — **no cell data**. The table block always has `has_children: true`.
2. **Second call** on the table block's ID yields `type: "table_row"` blocks. Each row has `.table_row.cells` — a 2D array of rich-text arrays. The first row is typically the header (bold text), subsequent rows are data.

Example workflow from a real session (reading "Gas Cost on rc.2"):

```
# Step 1 — get page blocks, find the table
rsi_notion_blocks_children(block_id=page_id)
→ block type="table", id="be71a5cf-...", has_children=true, table_width=5

# Step 2 — get the table's children (the actual rows)
rsi_notion_blocks_children(block_id="be71a5cf-...")
→ 14 table_row blocks, each with cells array:
  Row 0 (header): ["Action", "Gas Units", "Cost of Txn", "Cost on Base", "Txn Link"]
  Row 1: ["Register IP (without URI data)", "204,551", "$22", "$0.036", "tx link"]
  ...
```

**Pitfall:** The `rsi_notion.search` and `rsi_knowledge_search` results will include pages with tables, but you won't see the table content until you traverse both levels of children. Budget two sequential `blocks_children` calls when you suspect a page contains a table.

---

*Contributed by [@dogiladeveloper](https://github.com/dogiladeveloper)*
