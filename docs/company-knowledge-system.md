# RSI Company Knowledge System

This document records the Phase 1A storage contract for RSI company memory.

## Authority Boundary

Honcho is RSI's canonical company memory and compiled knowledge substrate.
Hermes local memory remains runtime assistant state and is not company knowledge
source of truth.

## Slack Session Mapping

The Slack mirror keeps source-native session keys in metadata:

- Threaded message: `slack:{workspace}:{channel}:{thread_ts}`
- Unthreaded channel message: `slack:{workspace}:{channel}:channel`

Honcho resource names cannot use raw Slack keys because the Honcho API enforces
`^[a-zA-Z0-9_-]+$` and a 100 character maximum. The mirror therefore stores the
raw source key in metadata and uses a deterministic Honcho-compatible encoded
session name for API calls.

## Idempotency Mechanism

Slack message idempotency uses an RSI wrapper table, `source_mirror_record`, with
`(source_type, source_key)` as the primary key. The mirrored content itself is
still written through Honcho's supported HTTP API.

For Slack messages:

- `source_type = slack_message`
- `source_key = slack:{workspace}:{channel}:{slack_ts}`
- `source_revision = edited.ts` when present, otherwise a stable marker derived
  from Slack timestamp plus content and file metadata hash

Same `source_key` and same `source_revision` is a no-op. A changed revision
creates a new Honcho message and updates the wrapper record to point at the
latest Honcho message ID; old Honcho messages remain historical evidence.

This avoids search-before-write races and avoids direct SQL writes into Honcho.

For Slack attachment analyses:

- `source_type = slack_attachment_analysis`
- `source_key = slack_attachment_analysis:{workspace}:{channel}:{slack_ts}:{file_id}:{extraction_kind}`
- `source_revision = {extraction_kind}:sha256:{content_sha256}:status:{extraction_status}:model:{model}`

Attachment bytes are lazily cached under `RSI_ATTACHMENT_CACHE_ROOT`; Honcho
stores extracted text plus provenance, not the blob itself. Retried extraction
uses the same `source_mirror_record` claim/complete path as Slack messages, so
the same analyzed attachment revision creates at most one Honcho message.

## Mirror Runtime

`control-plane --mode slack-mirror` performs resumable backfill over
`RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST` and writes checkpoints under
`RSI_SOURCE_MIRROR_CHECKPOINT_ROOT`.

When `RSI_SLACK_MIRROR_ENABLED=true` on the Slack surface, normal Slack events
from allowlisted channels are mirrored asynchronously after receipt. Mentioned
thread response is not blocked by mirror write failures, but enabling the mirror
requires Honcho and the source-mirror idempotency store to be configured at
startup.

## Read Contract

The local Slack MCP server exposes compiled-corpus reads:

- `conversation_get(channel_id, thread_ts, limit, page)`
- `messages_read(channel_id, thread_ts, oldest_ts, latest_ts, limit, page)`

Channel-wide `messages_read` requires a time window and pagination. Unbounded
channel history reads are refused. All Slack read tools require the channel to
be present in `RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST`; an unset allowlist is a
configuration error, not an implicit grant.

Live Slack exact-source tools (`slack_read_thread`, `slack_read_permalink`) stay
available for freshness and ambiguity resolution. Slack `search.messages` and
Slack user tokens are intentionally not used.

## Attachment Extraction

`attachments_fetch` is metadata-only by default. With `include_content=true`, it
downloads Slack files visible to `SLACK_BOT_TOKEN`, writes an atomic lazy cache,
and persists extracted evidence through the control-plane source mirror API.
`RSI_CONTROL_PLANE_BASE_URL` must be configured explicitly; generated
Kubernetes service environment variable names are not part of this contract.

Supported text files are decoded as UTF-8 with replacement for invalid bytes.
Image files require `analyze_images=true` and use OpenRouter vision with
`RSI_ATTACHMENT_VISION_MODEL`, defaulting to `qwen/qwen3.6-flash`. Unsupported
binary files record an explicit unsupported status and must not be summarized
from metadata alone.
