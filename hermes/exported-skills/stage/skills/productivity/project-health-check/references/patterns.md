# Search Query Library

## Proven queries for Slack corpus

### Unanswered questions / open items
```
open question OR unanswered OR "how to" OR "should we" OR "what about" OR "can you" OR TODO OR pending
```

### Blockers / stuck / waiting
```
blocked OR stuck OR waiting OR help needed OR "any update" OR "what's the status"
```

### Bugs / broken things
```
not working OR bug OR broken OR "doesn't work" OR issue OR error OR fix needed
```

### Post-launch / fast follow
```
fast follow OR post-launch OR "after launch" OR "still open" OR "likely next"
```

### Specific domain terms (combine as needed)
```
"seed phrase" OR castle OR beehiiv OR referral OR payout OR "admin api key" OR intercom OR paypal OR account deletion
```

## Proven queries for Notion corpus

### Open decisions
```
open question OR unanswered OR misalignment OR "need to decide" OR "should we" OR pending decision
```

### Launch readiness
```
launch checklist OR complete OR done OR status OR not started
```

## Cross-reference patterns

### Reading full threads
After finding signals via search, always read the full thread:
```
mcp_rsi_task_trace_*_slack_*_messages_read(channel_id, thread_ts)
```

### GitHub cross-reference
```
gh issue list --repo piplabs/<repo> --limit 20 --state open --json number,title,state,createdAt,labels
```

### Notion document retrieval
```
mcp_rsi_task_trace_*_slack_*_document_get(document_id)
```

## Misalignment detection patterns

| Pattern | What to look for |
|---|---|
| Notion checkbox vs reality | Item marked "[x]" but Slack shows it broken/partial |
| Cross-repo drift | OpenAPI vs FE API reference docs out of sync |
| Conceptual divergence | Same feature described differently in FE, BE, Notion |
| Config mismatch | App CORS allows origin X, WAF blocks origin X |
| Past-due decisions | Explicit due date passed with no resolution |
| Parent/child gap | Top-level unchecked, all children checked |
| Schema quality | Duplicate status values in Notion databases |
| Disconnected tracks | FE and BE working on same feature with no integration plan |

## Numo-specific knowledge (example from 2026-05-04 session)

### Key channels
- C0AKH5SNGKH (#proj-numo-depin-app) — main project channel
- C0ASQ9K5V50 (#team-tiger) — RSI/eng coordination

### Key repos
- piplabs/depin-backend — backend API + IP registration worker
- piplabs/numo-monorepo — frontend (web + RN + admin)

### Key Notion pages
- Numo Todo Checklists: 33a051299a54805785aee1b3796da675
- Numo v1 Launch Checklist DB: 044051299a5482cca0f9010354f35971
- Numo Execution Timeline: 334051299a5480e3acc3d7963ba07ffa
- Numo Product Backlog DB: 34f051299a5480649b98c7337f8c8084

### Known misalignments found
- Castle.io: marked done in Notion but broken until 4/30 (PR #378)
- Seed phrase: three divergent concepts across FE (BIP-39), BE (verification phrase), Notion (voice-based)
- Cloudflare WAF: blocks dev origins that app CORS allows on numolabs.ai domain
- Payment provider: decision due 4/17, still undecided as of 5/4
- Load testing: gates not passed but app launched 4/29
