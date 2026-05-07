# DB Read Tool API Reference

Session: 2026-05-07 — first use of the db-read-worker for prod depin queries.

## Architecture

```
hermes executor → control-plane (pod IP:8080) → db-read-worker pod →
  ├─ depin-stage: direct PostgreSQL (Vault DSN)
  ├─ depin-prod:  AWS Lambda → prod PostgreSQL
  └─ rsi-platform-stage: direct PostgreSQL (Vault DSN)
```

## Infrastructure

- **Control plane pod**: `use1-stage-rsi-agent-platform-control-plane-*` (namespace: `rsi-platform`)
- **DB read worker pod**: `use1-stage-rsi-agent-platform-control-plane-db-read-worker-*` (namespace: `rsi-platform`)
- **Prod Lambda**: `arn:aws:lambda:us-east-1:625966732747:function:use1-prod-rsi-db-read-depin-prod`
- **Discovery**: ConfigMap `use1-stage-rsi-agent-platform-control-plane-db-read-worker` (namespace: `rsi-platform`)

## Reachability

- **kubectl port-forward DOES NOT WORK**. The hermes executor SA lacks `pods/portforward` in `rsi-platform`. Use direct pod IP.
- Find pod IP: `kubectl get pod -n rsi-platform -l app.kubernetes.io/component=control-plane -o jsonpath='{.items[0].status.podIP}'`
- The pod IP is reachable from the hermes executor pod.

## Authentication

- Token: `RSI_DB_READ_CLIENT_TOKEN` env var
- Header: `Authorization: Bearer $RSI_DB_READ_CLIENT_TOKEN`

## Endpoints

### GET /internal/db-read/sources
Returns available database targets with caps.

Response format:
```json
{
  "targets": [
    {
      "id": "depin-prod",
      "placement": "prod",
      "allowed_schemas": ["public"],
      "allowed_tables": ["*"],
      "caps": {
        "max_rows": 100,
        "max_bytes": 65536,
        "timeout_seconds": 20,
        "lock_timeout_ms": 250
      },
      "approval_ttl": "1h"
    }
  ]
}
```

### GET /internal/db-read/schema?target=depin-prod
Returns schema view for a target (available tables, columns, types).

### POST /internal/db-read/validate
Pre-validate a SQL query without executing it.

Request:
```json
{
  "target": "depin-prod",
  "sql": "SELECT COUNT(*) FROM scripts WHERE is_active = TRUE"
}
```

Response:
```json
{
  "ok": true,
  "sql_sha256": "sha256:...",
  "preview": "SELECT COUNT(*) FROM scripts WHERE is_active = TRUE"
}
```

### POST /internal/db-read/query
Submit a query for execution. For prod targets, triggers Slack approval flow.

Request:
```json
{
  "target": "depin-prod",
  "sql": "SELECT COUNT(*) AS available_transcripts FROM scripts WHERE is_active = TRUE",
  "purpose": "Count available transcripts on prod DB",
  "requester": "hermes",
  "conversation_id": "conv-...",
  "workflow_id": "wf-...",
  "trace_id": "trace-...",
  "channel_id": "C0ASQ9K5V50",
  "thread_ts": "1778137700.384029"
}
```

Initial response (state: "validating"):
```json
{
  "request": {
    "id": "dbread_...",
    "idempotency_key": "dbread:sha256:...",
    "target": "depin-prod",
    "state": "validating",
    "expires_at": "2026-05-07T08:22:41Z",
    "caps": { ... },
    "redaction": { ... }
  },
  "status": "validating",
  "validation": {
    "ok": true,
    "sql_sha256": "sha256:...",
    "preview": "..."
  }
}
```

Final result (posted to Slack after approval):
```
RSI DB read request `dbread_...`: succeeded; rows=1 truncated=false 
Target: `depin-prod` 
Sample: [{"available_transcripts": "1414532"}]
```

## Targets

| Target ID | Placement | Path | Timeout | Row Limit |
|-----------|-----------|------|---------|-----------|
| depin-prod | prod | Lambda | 20s | 100 |
| depin-stage | stage | Direct | 5s | 100 |
| rsi-platform-stage | stage | Direct | 5s | 50 |

## Prod Redaction Rules

These columns are silently stripped from prod query results:

`access_token`, `api_key`, `authorization`, `email`, `password`, `phone`, `private_key`, `refresh_token`, `secret`, `token`

## Depin-backend Schema (Key Tables)

### scripts
- `id` (UUID PK)
- `campaign_id` (UUID FK → campaigns)
- `language_code` (TEXT)
- `content` (TEXT — the transcript text)
- `romanized_content` (TEXT, nullable)
- `is_active` (BOOLEAN, default TRUE)
- `import_key` (TEXT, nullable)
- `topic` (TEXT, nullable)
- `scenario` (TEXT, nullable)
- `direction` (TEXT, nullable)
- `source_text_en` (TEXT, nullable)
- `created_at`, `updated_at` (TIMESTAMPTZ)

### campaigns
- `id` (UUID PK)
- `campaign_name` (TEXT)
- `is_active` (BOOLEAN)
- `is_featured` (BOOLEAN)
- `supported_languages` (JSONB — array of language codes)
- `tags` (JSONB)
- `daily_cap` (INT)
- `cooldown_seconds` (INT)
- `reward_amount_usd` (NUMERIC)
- etc.

### submissions
- `id` (UUID PK)
- `user_id` (UUID FK → users)
- `campaign_id` (UUID FK → campaigns)
- `artifact_id` (UUID FK → artifacts)
- `script_assignment_id` (UUID FK → script_assignments)
- `state` (TEXT: created|uploaded|pending_review|accepted|rejected|failed)
- `mock` (BOOLEAN)
- `created_at`, `updated_at`, `uploaded_at`, `reviewed_at`

### users
- `id` (UUID PK)
- `primary_language` (TEXT — "vi" for Vietnamese)
- `known_languages` (JSONB)
- `nationality` (TEXT — 2-letter ISO code)
- `email` (TEXT — REDACTED in prod db-read)
- `birth_year` (INT)
- `created_at`, `last_seen_at`

## Query History (Session 2026-05-07)

### Query 1: Available transcript count
```sql
SELECT COUNT(*) AS available_transcripts FROM scripts WHERE is_active = TRUE
```
Result: **1,414,532** — succeeded, rows=1, truncated=false.

### Query 2: Transcripts by language (active campaigns)
```sql
SELECT s.language_code, COUNT(*) AS transcript_count
FROM scripts s JOIN campaigns c ON c.id = s.campaign_id
WHERE s.is_active = TRUE AND c.is_active = TRUE
GROUP BY s.language_code ORDER BY transcript_count DESC LIMIT 20
```
Result: pending approval as of session end.
