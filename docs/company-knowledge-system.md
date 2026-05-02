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
