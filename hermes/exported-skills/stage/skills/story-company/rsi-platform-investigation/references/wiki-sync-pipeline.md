# Wiki Sync Pipeline Code Trace

Reproduction from 2026-05-04 investigation of "llm wiki commit deployed but wiki empty."

## The Pipeline

### Step 1: CronJob triggers mirror
```
slack-mirror (every 15min) → runs in mode: incremental or backfill
notion-mirror (every 30min) → same
```

### Step 2: mirrorSlackChannel (internal/control/slack_mirror.go:242)
```go
func mirrorSlackChannel(ctx, cfg, api, mirrorStore, mirror, workspaceID, channelID) {
    checkpoint := readSlackMirrorCheckpoint(...)        // line 243
    oldest, latest, mode := slackMirrorHistoryWindow(checkpoint) // line 248
    // mode is "incremental" if BackfillComplete=true and no new messages
    // mode is "backfill" if BackfillComplete=false
    
    for each API page of messages:
        wikiBatch := newSlackWikiPublishBatch(cfg, mirrorStore)  // line 266
        
        for each message:
            result := mirror.IngestMessage(ctx, input)            // line 275
            if shouldPublishSlackWikiSource(result) {              // line 280
                wikiBatch.record(ctx, input)                       // line 281
            }
            if msg.ReplyCount > 0:
                mirrorSlackThread(ctx, api, mirror, wikiBatch, ...) // line 290
        
        wikiBatch.publish(ctx)                                    // line 298
}
```

### Step 3: wikiBatch.record → RecordWikiSourceRevision (internal/companyknowledge/wiki.go:66,75)
```go
func RecordWikiSourceRevision(ctx, cfg, repo, input) {
    wikiStore := repo.(store.CompanyWikiStore)
    // Skips if:
    //   - store doesn't implement CompanyWikiStore → "store_not_company_wiki_capable"
    //   - CompanyWikiRoot is empty → "company_wiki_root_not_configured"
    
    source := wikiStore.UpsertCompanyWikiSourceRevision(input)
    if !source.Inserted:
        return Skipped(Reason: "revision_already_exists")
    return source
}
```

### Step 4: UpsertCompanyWikiSourceRevision (internal/store/company_wiki_postgres.go:12)
- Generates deterministic IDs: `CompanyWikiStableID("srcdoc", sourceType, documentSourceKey)`
- UPSERTs into `company_source_document` (on conflict: update metadata)
- INSERTS into `company_source_revision` with `ON CONFLICT DO NOTHING`
- If new revision: inserts chunks into `company_source_chunk`
- Returns `Inserted: true/false`

### Step 5: wikiBatch.publish → PublishWikiSourceDocument (internal/companyknowledge/wiki.go:94)
- Lists chunks for the document
- Builds compiled markdown with `BuildCompiledWikiMarkdown`
- Writes markdown file via `PublishMarkdownFile` (creates dirs on demand)
- Records publication in DB via `PublishCompanyWikiPage`
- Writes manifest, index, and log files

## The Bug: No Backfill

The wiki source creation is triggered ONLY by `IngestMessage` during mirror runs. Key insight:

```
Timeline:
T1: Mirrors backfill all historical Slack/Notion content (no wiki code exists)
T2: Wiki code deployed (commits 93819fb, 44960eb in e1d0154)
T3: Mirrors run in INCREMENTAL mode → 0 new messages → RecordWikiSourceRevision never called
T4: company_source_document table: 0 rows
T5: Wiki directory /workspace/company/wiki/ never created
```

The `BackfillComplete` field in the mirror checkpoint (`internal/control/slack_mirror.go:27`) is `true`, meaning the mirror has backfilled all historical content and is now in incremental-only mode.

## Key Files

| File | Purpose |
|---|---|
| `internal/companyknowledge/wiki.go` | Core wiki logic: RecordWikiSourceRevision, PublishWikiSourceDocument, PublishMarkdownFile, WriteIndexFile, AppendLogEntry |
| `internal/companyknowledge/slack.go` | SlackWikiSourceRevisionInput (line 316), should not be confused with shouldPublish function |
| `internal/companyknowledge/notion.go` | NotionWikiSourceRevisionInput (line 398) |
| `internal/control/slack_mirror.go` | Mirror entry points, shouldPublishSlackWikiSource (line 188), mirrorSlackChannel (line 242), mirrorSlackThread (line 337) |
| `internal/control/notion_mirror.go` | Notion mirror (same pattern) |
| `internal/control/company_wiki_api.go` | Wiki HTTP API handlers |
| `internal/control/router.go` | Route registration (wiki routes at /internal/company-wiki/*) |
| `internal/store/company_wiki.go` | Wiki data types and CompanyWikiStore interface |
| `internal/store/company_wiki_postgres.go` | Postgres implementation: UpsertCompanyWikiSourceRevision, PublishCompanyWikiPage, etc. |
| `internal/store/company_wiki_memory.go` | In-memory implementation (for testing) |
| `internal/store/source_mirror.go` | SourceMirrorWriteStore interface (does NOT embed CompanyWikiStore) |
| `internal/db/migrations/033_company_wiki.sql` | Wiki database schema (company_source_document, company_source_revision, company_source_chunk, wiki pages, citations, audits) |

## Component Route Ownership

Only the **control-plane** registers wiki routes (`internal/control/router.go`). The **improvement-plane** has its own router and returns HTML for the same paths. If you get HTML back from a wiki endpoint, you're hitting the wrong service.

## Deployment Verification (2026-05-04)

All components in staging namespace running commit `e1d01543410dfe214c8a4a6a26e700d388e37311`:
- control-plane, control-plane-action-worker, control-plane-slack-rsi, control-plane-worker
- hermes-skill-exporter, honcho-api, honcho-deriver
- improvement-plane
- notion-mcp (latest tag, not pinned)

Wiki commits included in this deployment:
- `93819fb` Add native company wiki vertical slice (#229) — 4,153 lines, 34 files
- `44960eb` Export company wiki with Hermes skills (#232) — 186 lines, 3 files
