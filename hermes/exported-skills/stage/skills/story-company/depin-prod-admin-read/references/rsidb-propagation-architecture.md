# RSI Platform DB Read Result Propagation Architecture

Traced from `piplabs/rsi-agent-platform` source on 2026-05-08 during a deep investigation into why `db_read.query` results fail to reach Hermes.

## Three-Layer Async Resume Mechanism

```
┌──────────┐    db_read.query     ┌──────────────┐    Slack approval    ┌───────────────┐
│  Hermes  │ ──────────────────▶  │ Control Plane│ ─────────────────▶   │ DB Read Worker│
│          │                      │              │                      │               │
│          │ ◀── resume_with_     │              │ ◀── execution result │               │
│          │     tool_result()    │              │                      │               │
└──────────┘                      └──────────────┘                      └───────────────┘
```

## Key Files and Line Numbers

| Component | File | Lines | Purpose |
|-----------|------|-------|---------|
| Tool definition | `runner/rsi_runner/rsi_tools.py` | 400 | `db_read.query` — submits to control plane, returns pause |
| Pause detection | `runner/rsi_runner/hermes_runtime.py` | 4857-4880 | Detects `external_tool_pending` termination reason |
| Pause creation | `internal/control/external_tool_pause.go` | 15-95 | `handleExternalToolPendingRunnerResult` — creates pause, submits `CommandWorkflowWaitingExternalTool` |
| Workflow state | `internal/transition/workflow.go` | 233-253 | `reduceWorkflowWaitingExternalTool` — advances to `WorkflowStateWaitingExternalTool` |
| DB execution | `internal/control/db_read_worker.go` | 275-306 | Executes SQL, creates `DBReadExecutionResult`, calls `markDBReadExternalToolOutcome` |
| Sample population | `internal/store/db_read_postgres.go` | 259-312 | `AppendDBReadExecutionResult` — updates `ResultSample` from execution result rows |
| Resume payload | `internal/control/external_tool_pause.go` | 97-131, 134-170 | `markDBReadExternalToolOutcome` + `buildDBReadExternalToolResumePayload` |
| Resume trigger | `internal/control/external_tool_pause.go` | 172-213 | `tryQueueExternalToolResume` — checks pause is terminal, submits `CommandExternalToolResultReady` |
| Workflow resume | `internal/transition/workflow.go` | 255-285 | `reduceExternalToolResultReady` — creates `EffectInvokeRunner` |
| Runner dispatch | `internal/control/worker.go` | 333-334 | Maps `external_tool_resume` to `runnerTask.ExternalToolResume` |
| Hermes resume | `runner/rsi_runner/hermes_agent_adapter.py` | 547-561 | Calls `agent.resume_with_tool_result(session_id, tool_call_id, content)` |

## Resume Payload Structure

Built by `buildDBReadExternalToolResumePayload` (external_tool_pause.go:134-170):

```json
{
  "kind": "external_tool_result",
  "session_id": "<hermes_session_id>",
  "tool_call_id": "<paused_tool_call_id>",
  "tool_name": "db_read_query",
  "status": "ok",
  "content": {
    "kind": "db_read_result",
    "status": "ok",
    "request_id": "<db_read_request_id>",
    "target": "depin-prod",
    "sql_sha256": "<hash>",
    "row_count": 10,
    "truncated": false,
    "sample": [{"col1": "val1"}, {"col1": "val2"}]
  },
  "transcript_snapshot": [...]
}
```

## Known Failure Modes

### 1. `db_read.query` returns "did not create an external tool pause"
**False negative.** The control plane POSTs the approval card to Slack asynchronously while Hermes receives a synchronous error. Each call creates a new approval card. **Never loop-retry** — check the Slack thread instead.

### 2. `db_read.status` returns 404
**Token mismatch.** Caused by `DBReadClientToken` being unset or mismatched between the Hermes executor environment and the control plane config. The auth check at `db_read_api.go:277-284` verifies an HMAC-signed token; if the secret doesn't match, access is denied. Fall back to the Slack mirror for request status.

### 3. `[Result unavailable — see context summary above]`
**Result not propagated to Hermes context.** The query was approved and executed (confirmed by Slack card: `succeeded; rows=N`), but the sanitized row data was not delivered to Hermes. The `content` field in the resume payload contains the data, but Hermes may not match it to the paused tool call or may not parse the `db_read_result` content type correctly. Check Slack mirror for `*Result:*` or `*Sample:*` blocks in approval cards.

### 4. `tryQueueExternalToolResume` silent skip
At `external_tool_pause.go:174`, the function checks `ExternalToolPauseTerminalOutcome(pause.ToolOutcome)`. If the outcome is still "pending" or "executing", the function returns silently. This is by design — the resume is only triggered after `markDBReadExternalToolOutcome` sets ToolOutcome to "succeeded"/"failed"/"denied"/"expired".

## Grafana Access

Grafana at `grafana.ops.storyprotocol.net` is **blocked by Cloudflare** (Error 1010 — browser signature banned). The `rsi_observability.*` tools return 403 with a CF block. The `GRAFANA_TOKEN` is present in the executor environment but Cloudflare's WAF rejects the company-computer user-agent. This is a Cloudflare WAF configuration issue, not a token problem.

## Control Plane Logging

The control plane pods (both `control-plane-*` replicas and workers) have **very sparse debug-level logging**. The `db-read-worker` logs only startup messages. The `control-plane` pods don't log individual external tool pause/resume events at default log levels. When debugging result propagation issues, rely on:
- Slack approval cards (authoritative for approval and execution status)
- `session_search` for prior session summaries
- Direct code inspection of `piplabs/rsi-agent-platform` (above file references)

## DB Read Worker

```bash
kubectl get pods -n rsi-platform | grep db-read
# use1-stage-rsi-agent-platform-control-plane-db-read-workerjk9pl   1/1   Running
```

The worker starts with targets `[depin-stage depin-prod rsi-platform-stage]`. Stage queries route through the worker pod directly (5s timeout). Prod queries route through AWS Lambda relay (20s timeout, requires Slack approval).