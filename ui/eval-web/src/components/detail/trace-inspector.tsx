import { useEffect, useMemo, useRef, useState } from "react";

import type { TraceDetailResponse, TraceInspectorTab, NullableList, EvalJudgment, ExecutionLedgerEvent, JsonObject, JsonValue } from "@/types";
import { formatTime, getJSON, latestActionResult, listOrEmpty, scoreBadge } from "@/hooks/api";
import { EmptyDetail } from "./empty-detail";
import { FormattedMessage } from "@/components/formatted-message";

const LIVE_EVENT_PAGE_SIZE = 100;
const LIVE_EVENT_LIMIT = 1000;
const LIVE_EVENT_FAMILIES = ["all", "model", "tool", "terminal", "artifact", "slack", "notion", "mcp", "phase", "failure"];
const LIVE_DELTA_KINDS = new Set(["model.reasoning.delta", "model.output.delta", "terminal.output", "executor.subprocess.output"]);
const TOOL_LIFECYCLE_KINDS = new Set(["tool.call.started", "tool.call.progress", "tool.call.completed"]);
const COMPACT_METADATA_KEYS = new Set([
  "hermes_session_id",
  "observation_id",
  "recorded_at_unix",
  "role",
  "source"
]);

type LiveStreamItem = {
  id: string;
  kind: string;
  status?: string;
  phase_id?: string;
  seq: number;
  seq_end: number;
  recorded_at: string;
  count: number;
  event: ExecutionLedgerEvent;
  events: ExecutionLedgerEvent[];
  text?: string;
  toolCallKey?: string;
};

type LiveLedgerPageResponse = {
  events: ExecutionLedgerEvent[];
  paging: {
    has_more: boolean;
    next_before?: string;
  };
};

function eventFamily(kind: string) {
  if (kind.startsWith("tool_")) {
    return "tool";
  }
  if (kind === "file.written") {
    return "artifact";
  }
  const prefix = (kind || "").split(".")[0] || "event";
  if (prefix === "executor" || prefix === "command" || prefix === "terminal") {
    return "terminal";
  }
  return prefix;
}

function payloadText(payload: JsonObject | undefined, keys: string[]) {
  if (!payload) {
    return "";
  }
  for (const key of keys) {
    const value = payload[key];
    if (typeof value === "string" && value.trim()) {
      return value;
    }
  }
  return "";
}

function isToolLifecycleKind(kind: string) {
  return TOOL_LIFECYCLE_KINDS.has(kind);
}

function compactToolName(value: string) {
  let name = value.trim();
  if (!name) {
    return "";
  }
  const mcpTaskMatch = name.match(/^mcp_rsi_task_trace_[A-Za-z0-9-]+_\d+_(.+)$/);
  if (mcpTaskMatch?.[1]) {
    name = mcpTaskMatch[1];
  }
  name = name.replace(/^[a-z]+_[a-f0-9]{6,}_/, "");
  return name;
}

function toolCallKey(event: ExecutionLedgerEvent) {
  if (!isToolLifecycleKind(event.kind)) {
    return "";
  }
  const payload = event.payload || {};
  const callID = payloadText(payload, ["tool_call_id", "call_id", "id"]);
  const toolName = payloadText(payload, ["tool_name", "name"]);
  if (callID) {
    return `${event.execution_id}:${callID}`;
  }
  if (toolName) {
    return `${event.execution_id}:${event.phase_id || "main"}:${toolName}`;
  }
  return "";
}

function liveDeltaText(event: ExecutionLedgerEvent) {
  if (!LIVE_DELTA_KINDS.has(event.kind)) {
    return "";
  }
  const payload = event.payload;
  if (!payload) {
    return "";
  }
  for (const key of ["delta", "chunk_text", "text", "message"]) {
    const value = payload[key];
    if (typeof value === "string") {
      return value;
    }
  }
  return "";
}

function canMergeLiveDelta(current: LiveStreamItem | undefined, event: ExecutionLedgerEvent) {
  if (!current || !current.text || !LIVE_DELTA_KINDS.has(event.kind)) {
    return false;
  }
  return (
    current.kind === event.kind &&
    current.event.execution_id === event.execution_id &&
    (current.phase_id || "") === (event.phase_id || "") &&
    event.seq === current.seq_end + 1
  );
}

function buildLiveStreamItems(events: ExecutionLedgerEvent[]) {
  const groupedToolItems = new Map<string, LiveStreamItem>();
  return events.reduce<LiveStreamItem[]>((items, event) => {
    const toolKey = toolCallKey(event);
    if (toolKey) {
      const existing = groupedToolItems.get(toolKey);
      if (existing) {
        existing.kind = event.kind;
        existing.status = event.status;
        existing.seq_end = event.seq;
        existing.recorded_at = event.recorded_at;
        existing.count += 1;
        existing.event = event;
        existing.events.push(event);
        return items;
      }
      const item: LiveStreamItem = {
        id: `tool:${toolKey}`,
        kind: event.kind,
        status: event.status,
        phase_id: event.phase_id,
        seq: event.seq,
        seq_end: event.seq,
        recorded_at: event.recorded_at,
        count: 1,
        event,
        events: [event],
        toolCallKey: toolKey
      };
      groupedToolItems.set(toolKey, item);
      items.push(item);
      return items;
    }

    const delta = liveDeltaText(event);
    const previous = items[items.length - 1];
    if (delta && canMergeLiveDelta(previous, event)) {
      previous.text = `${previous.text || ""}${delta}`;
      previous.seq_end = event.seq;
      previous.recorded_at = event.recorded_at;
      previous.count += 1;
      previous.event = event;
      previous.events.push(event);
      return items;
    }

    items.push({
      id: delta ? `${event.kind}:${event.execution_id}:${event.phase_id || "main"}:${event.seq}` : event.id,
      kind: event.kind,
      status: event.status,
      phase_id: event.phase_id,
      seq: event.seq,
      seq_end: event.seq,
      recorded_at: event.recorded_at,
      count: 1,
      event,
      events: [event],
      text: delta || undefined
    });
    return items;
  }, []);
}

function isJsonObject(value: JsonValue): value is JsonObject {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function parseToolResultSummary(value: JsonValue | undefined) {
  if (typeof value !== "string" || !value.trim().startsWith("{")) {
    return "";
  }
  try {
    const parsed = JSON.parse(value) as JsonValue;
    if (!isJsonObject(parsed)) {
      return "";
    }
    return [
      typeof parsed.summary === "string" ? parsed.summary : "",
      typeof parsed.status === "string" ? `status: ${parsed.status}` : "",
      typeof parsed.provider_ref === "string" ? parsed.provider_ref : ""
    ].filter(Boolean).join(" · ");
  } catch {
    return "";
  }
}

function truncateText(value: string, limit = 900) {
  if (value.length <= limit) {
    return value;
  }
  return `${value.slice(0, limit).trimEnd()}...`;
}

function compactValue(value: JsonValue): JsonValue {
  if (typeof value === "string") {
    const parsedSummary = parseToolResultSummary(value);
    return parsedSummary || truncateText(value);
  }
  if (Array.isArray(value)) {
    return value.slice(0, 8).map(compactValue);
  }
  if (isJsonObject(value)) {
    const out: JsonObject = {};
    for (const [key, nested] of Object.entries(value)) {
      if (COMPACT_METADATA_KEYS.has(key)) {
        continue;
      }
      out[key] = compactValue(nested);
    }
    return out;
  }
  return value;
}

function compactEventDetails(item: LiveStreamItem) {
  const base: JsonObject = {
    kind: item.kind,
    status: item.status || "event",
    phase_id: item.phase_id || "main",
    seq: item.count > 1 ? `${item.seq}-${item.seq_end}` : item.seq,
    recorded_at: item.recorded_at,
    event_count: item.count
  };
  const payload = compactValue(item.event.payload || {});
  if (item.count > 1) {
    const firstEventID = item.events[0]?.id;
    const lastEventID = item.events[item.events.length - 1]?.id;
    if (firstEventID) {
      base.first_event_id = firstEventID;
    }
    if (lastEventID) {
      base.last_event_id = lastEventID;
    }
    base.characters = item.text?.length || 0;
  }
  return { ...base, payload };
}

function eventTitle(item: LiveStreamItem) {
  const payload = item.event.payload || {};
  const toolName = payloadText(payload, ["tool_name", "name"]);
  const shortToolName = compactToolName(toolName);
  switch (item.kind) {
    case "model.reasoning.delta":
      return "Reasoning stream";
    case "model.output.delta":
      return "Assistant output";
    case "model.thinking":
      return "Thinking";
    case "tool.call.started":
    case "tool.call.progress":
    case "tool.call.completed":
      return shortToolName || "Tool call";
    case "tool.generation.started":
      return "Tool generation";
    case "artifact.created":
    case "file.written":
      return "Artifact written";
    case "slack.message.sent":
    case "slack.upload.completed":
      return "Slack delivery";
    default:
      return item.kind;
  }
}

function eventPrimaryText(item: LiveStreamItem) {
  if (item.text) {
    return item.text;
  }
  for (let i = item.events.length - 1; i >= 0; i--) {
    const payload = item.events[i]?.payload || {};
    const parsedResult = parseToolResultSummary(payload.result);
    if (parsedResult) {
      return parsedResult;
    }
    const text = payloadText(payload, [
      "summary",
      "result_summary",
      "message",
      "text",
      "chunk_text",
      "preview",
      "error",
      "failure_reason",
      "reason"
    ]);
    if (text) {
      return text;
    }
  }
  return "";
}

function LiveTraceStream(props: { traceID: string }) {
  const [events, setEvents] = useState<ExecutionLedgerEvent[]>([]);
  const [status, setStatus] = useState("connecting");
  const [familyFilter, setFamilyFilter] = useState("all");
  const [autoscroll, setAutoscroll] = useState(true);
  const [hasOlderEvents, setHasOlderEvents] = useState(false);
  const [loadingOlderEvents, setLoadingOlderEvents] = useState(false);
  const viewportRef = useRef<HTMLDivElement | null>(null);
  const olderProbeKeyRef = useRef("");

  useEffect(() => {
    setEvents([]);
    setStatus("connecting");
    setAutoscroll(true);
    setHasOlderEvents(false);
    setLoadingOlderEvents(false);
    olderProbeKeyRef.current = "";
    if (typeof EventSource === "undefined") {
      setStatus("stream unavailable");
      return;
    }
    const source = new EventSource(`/api/traces/${props.traceID}/stream?scope=main&limit=${LIVE_EVENT_PAGE_SIZE}`);
    const onLedger = (event: MessageEvent) => {
      try {
        const parsed = JSON.parse(event.data) as ExecutionLedgerEvent;
        setEvents((current) => {
          if (!parsed.id || current.some((item) => item.id === parsed.id)) {
            return current;
          }
          return [...current, parsed].slice(-LIVE_EVENT_LIMIT);
        });
        setStatus("live");
      } catch {
        setStatus("parse error");
      }
    };
    source.addEventListener("ledger", onLedger as EventListener);
    source.onopen = () => setStatus("live");
    source.onerror = () => setStatus("reconnecting");
    return () => {
      source.removeEventListener("ledger", onLedger as EventListener);
      source.close();
    };
  }, [props.traceID]);

  const oldestEventID = events[0]?.id || "";

  useEffect(() => {
    if (hasOlderEvents || loadingOlderEvents || events.length < LIVE_EVENT_PAGE_SIZE || !oldestEventID) {
      return;
    }
    const probeKey = `${props.traceID}:${oldestEventID}`;
    if (olderProbeKeyRef.current === probeKey) {
      return;
    }
    olderProbeKeyRef.current = probeKey;
    let cancelled = false;
    getJSON<LiveLedgerPageResponse>(
      `/api/traces/${props.traceID}/ledger?scope=main&limit=1&before=${encodeURIComponent(oldestEventID)}`
    )
      .then((page) => {
        if (cancelled) {
          olderProbeKeyRef.current = "";
          return;
        }
        if (listOrEmpty(page.events).length > 0 || Boolean(page.paging?.has_more)) {
          setHasOlderEvents(true);
        }
      })
      .catch(() => {
        if (olderProbeKeyRef.current === probeKey) {
          olderProbeKeyRef.current = "";
        }
      });
    return () => {
      cancelled = true;
    };
  }, [props.traceID, events.length, oldestEventID, hasOlderEvents, loadingOlderEvents]);

  const streamItems = useMemo(() => buildLiveStreamItems(events), [events]);
  const visibleEvents = useMemo(
    () => streamItems.filter((item) => familyFilter === "all" || eventFamily(item.kind) === familyFilter),
    [streamItems, familyFilter]
  );

  useEffect(() => {
    if (!autoscroll || !viewportRef.current) {
      return;
    }
    viewportRef.current.scrollTop = viewportRef.current.scrollHeight;
  }, [visibleEvents, autoscroll]);

  const handleScroll = () => {
    const node = viewportRef.current;
    if (!node) {
      return;
    }
    const nearBottom = node.scrollHeight - node.scrollTop - node.clientHeight < 32;
    setAutoscroll(nearBottom);
  };

  const loadOlderEvents = async () => {
    const beforeID = events[0]?.id;
    if (!beforeID || loadingOlderEvents) {
      return;
    }
    setLoadingOlderEvents(true);
    setAutoscroll(false);
    try {
      const page = await getJSON<LiveLedgerPageResponse>(
        `/api/traces/${props.traceID}/ledger?scope=main&limit=${LIVE_EVENT_PAGE_SIZE}&before=${encodeURIComponent(beforeID)}`
      );
      setEvents((current) => {
        const seen = new Set<string>();
        const merged = [...listOrEmpty(page.events), ...current].filter((item) => {
          if (!item.id || seen.has(item.id)) {
            return false;
          }
          seen.add(item.id);
          return true;
        });
        return merged.slice(-LIVE_EVENT_LIMIT);
      });
      setHasOlderEvents(Boolean(page.paging?.has_more));
    } catch {
      setStatus("history unavailable");
    } finally {
      setLoadingOlderEvents(false);
    }
  };

  return (
    <div className="detail-section-body">
      <div className="live-toolbar">
        <div>
          <strong>Live execution stream</strong>
          <p className="muted">{status} · {events.length} events · {streamItems.length} rows</p>
        </div>
        <div className="button-row">
          {LIVE_EVENT_FAMILIES.map((family) => (
            <button key={family} className={familyFilter === family ? "segment-button active" : "segment-button"} onClick={() => setFamilyFilter(family)}>
              {family}
            </button>
          ))}
          <button className="secondary" onClick={() => setAutoscroll(true)}>{autoscroll ? "Auto" : "Resume"}</button>
        </div>
      </div>
      <div className="live-stream" ref={viewportRef} onScroll={handleScroll}>
        {hasOlderEvents ? (
          <button className="live-history-button secondary" onClick={loadOlderEvents} disabled={loadingOlderEvents}>
            {loadingOlderEvents ? "Loading older" : "Load older"}
          </button>
        ) : null}
        {visibleEvents.map((item) => (
          <LiveEventRow key={item.id} item={item} />
        ))}
        {!visibleEvents.length ? <div className="nested-card"><p className="detail-copy">Waiting for live runner events.</p></div> : null}
      </div>
    </div>
  );
}

function LiveEventRow(props: { item: LiveStreamItem }) {
  const item = props.item;
  const payload = item.event.payload || {};
  const family = eventFamily(item.kind);
  const primaryText = eventPrimaryText(item);
  const toolName = payloadText(payload, ["tool_name", "name"]);
  const shortToolName = compactToolName(toolName);
  const seqText = item.count > 1 ? `#${item.seq}-${item.seq_end}` : `#${item.seq}`;
  const kindText = isToolLifecycleKind(item.kind)
    ? `${item.count} ${item.count === 1 ? "update" : "updates"}`
    : `${item.kind}${item.count > 1 ? ` · ${item.count} chunks` : ""}`;
  return (
    <div className={`live-event ${family}`}>
      <div className="detail-row-header">
        <strong>{eventTitle(item)}</strong>
        <small>{item.status || "event"} · {item.phase_id || "main"} · {seqText} · {formatTime(item.recorded_at)}</small>
      </div>
      <p className="muted live-event-kind">{kindText}</p>
      {toolName && shortToolName && shortToolName !== toolName && !eventTitle(item).includes(shortToolName) ? <p className="muted">{shortToolName}</p> : null}
      {primaryText ? <pre className="detail-copy live-event-text">{primaryText}</pre> : null}
      <details>
        <summary>payload</summary>
        <pre className="detail-copy live-event-json">{JSON.stringify(compactEventDetails(item), null, 2)}</pre>
      </details>
    </div>
  );
}

export function TraceInspector(props: {
  selectedTraceId?: string;
  traceDetail?: TraceDetailResponse;
  tab: TraceInspectorTab;
  setTab: (tab: TraceInspectorTab) => void;
  onRunEval: () => void;
  onReplay: () => void;
  traceJudgments: Record<string, NullableList<EvalJudgment>>;
  feedbackTargets: { label: string; value: string; type: string }[];
  feedbackTargetType: string;
  setFeedbackTargetType: (value: string) => void;
  feedbackTargetID: string;
  setFeedbackTargetID: (value: string) => void;
  feedbackScore: string;
  setFeedbackScore: (value: string) => void;
  feedbackVerdict: string;
  setFeedbackVerdict: (value: string) => void;
  feedbackNotes: string;
  setFeedbackNotes: (value: string) => void;
  onSubmitFeedback: () => void;
}) {
  if (!props.selectedTraceId) {
    return <EmptyDetail title="Select a trace" body="Choose a trace attempt to inspect its timeline, visible reasoning, tools, Slack actions, evals, and operator feedback." />;
  }
  if (!props.traceDetail) {
    return <EmptyDetail title="Loading trace" body="Fetching the bounded evidence object for the selected trace." />;
  }

  const traceDetail = props.traceDetail;
  const trace = traceDetail.trace;
  const inspectorTabs: TraceInspectorTab[] = ["live", "overview", "timeline", "reasoning", "tools", "actions", "slack", "outcomes", "evals", "feedback", "proposals"];
  const runtimeSummary = traceDetail.runtime_summary;
  const executorObservations = listOrEmpty(traceDetail.harness_execution_observations);
  const recentExecutorOutput = executorObservations
    .filter((item) => item.event_type === "terminal.output" || item.event_type === "executor.subprocess.output")
    .slice(-20);

  return (
    <div className="detail-card">
      <div className="detail-header">
        <div>
          <p className="eyebrow">Trace inspector</p>
          <h2>{trace.summary.trace_id}</h2>
        </div>
        <div className="button-row">
          <button onClick={props.onRunEval}>Run eval</button>
          <button className="secondary" onClick={props.onReplay}>Queue replay</button>
        </div>
      </div>
      <div className="segment-row">
        {inspectorTabs.map((tab) => (
          <button key={tab} className={props.tab === tab ? "segment-button active" : "segment-button"} onClick={() => props.setTab(tab)}>
            {tab}
          </button>
        ))}
      </div>

      {props.tab === "live" ? <LiveTraceStream traceID={trace.summary.trace_id} /> : null}

      {props.tab === "overview" ? (
        <div className="detail-section-body">
          <dl className="overview-grid">
            <div><dt>Conversation</dt><dd>{props.traceDetail.conversation.title || props.traceDetail.conversation.external_key}</dd></div>
            <div><dt>Case</dt><dd>{props.traceDetail.case?.title || trace.summary.case_id}</dd></div>
            <div><dt>Status</dt><dd>{trace.summary.status}</dd></div>
            <div><dt>Thread key</dt><dd>{trace.summary.thread_key}</dd></div>
            <div><dt>Events</dt><dd>{trace.summary.event_count}</dd></div>
            <div><dt>Reasoning</dt><dd>{trace.summary.reasoning_step_count}</dd></div>
            <div><dt>Tools</dt><dd>{trace.summary.tool_call_count}</dd></div>
            <div><dt>Slack</dt><dd>{trace.summary.slack_action_count}</dd></div>
            <div><dt>Actions</dt><dd>{listOrEmpty(props.traceDetail.action_intents).length}</dd></div>
            <div><dt>Outcomes</dt><dd>{listOrEmpty(props.traceDetail.outcomes).length}</dd></div>
            <div><dt>Knowledge</dt><dd>{listOrEmpty(props.traceDetail.knowledge_entries).length}</dd></div>
            <div><dt>Harness runs</dt><dd>{listOrEmpty(props.traceDetail.harness_executions).length}</dd></div>
            <div><dt>Runtime source</dt><dd>{runtimeSummary?.runtime_source || "none"}</dd></div>
            <div><dt>Execution</dt><dd>{runtimeSummary?.execution_id || "none"}</dd></div>
            <div><dt>Phase</dt><dd>{runtimeSummary?.phase || "none"}</dd></div>
            <div><dt>Latest event</dt><dd>{runtimeSummary?.event_type || "none"}</dd></div>
            <div><dt>Runtime status</dt><dd>{runtimeSummary?.status || "none"}</dd></div>
            <div><dt>Engine</dt><dd>{runtimeSummary?.engine || "none"}</dd></div>
          </dl>
          {runtimeSummary ? (
            <div className="detail-card">
              <h3>Executor runtime</h3>
              <dl className="overview-grid">
                <div><dt>Recorded</dt><dd>{formatTime(runtimeSummary.recorded_at)}</dd></div>
                <div><dt>Workspace</dt><dd>{runtimeSummary.workspace_root || "none"}</dd></div>
              </dl>
            </div>
          ) : null}
          {props.traceDetail.workflow_line ? (
            <div className="detail-card">
              <h3>Workflow line</h3>
              <dl className="overview-grid">
                <div><dt>Status</dt><dd>{props.traceDetail.workflow_line.status}</dd></div>
                <div><dt>Current attempt</dt><dd>{props.traceDetail.workflow_line.current_workflow_id || "none"}</dd></div>
                <div><dt>Retry budget</dt><dd>{props.traceDetail.workflow_line.auto_retry_budget_remaining}</dd></div>
                <div><dt>Last failure</dt><dd>{props.traceDetail.workflow_line.last_failure_class || "none"}</dd></div>
                <div><dt>Retry at</dt><dd>{props.traceDetail.workflow_line.retry_after ? formatTime(props.traceDetail.workflow_line.retry_after) : "none"}</dd></div>
                <div><dt>Next action</dt><dd>{props.traceDetail.workflow_line.next_retry_action || "none"}</dd></div>
              </dl>
            </div>
          ) : null}
          {listOrEmpty(props.traceDetail.workflow_attempts).length ? (
            <div className="detail-card">
              <h3>Attempt lineage</h3>
              <div className="nested-list">
                {listOrEmpty(props.traceDetail.workflow_attempts).map((attempt) => (
                  <div key={attempt.workflow_id} className="nested-card">
                    <div className="detail-row-header">
                      <strong>{attempt.workflow_id}</strong>
                      <small>attempt {attempt.attempt_number} · {attempt.status}</small>
                    </div>
                    <p className="detail-copy">{attempt.trace_id || "No linked trace."}</p>
                    <p className="muted">
                      Parent: {attempt.parent_workflow_id || "none"} · Supersedes trace: {attempt.supersedes_trace_id || "none"}
                    </p>
                    <p className="muted">
                      Failure: {attempt.failure_class || "none"} · Retry: {attempt.retry_decision || "none"} · Repair: {attempt.repair_attempted ? (attempt.repair_succeeded ? "succeeded" : "failed") : "not needed"}
                    </p>
                  </div>
                ))}
              </div>
            </div>
          ) : null}
          {listOrEmpty(props.traceDetail.harness_executions).length ? (
            <div className="detail-card">
              <h3>Hermes session continuity</h3>
              <div className="nested-list">
                {listOrEmpty(props.traceDetail.harness_executions).map((item) => (
                  <div key={item.id} className="nested-card">
                    <div className="detail-row-header">
                      <strong>{item.role}</strong>
                      <small>{formatTime(item.created_at)}</small>
                    </div>
                    <p className="detail-copy">{item.hermes_session_id}</p>
                    <p className="muted">Scope: {item.session_scope_kind}:{item.session_scope_id}</p>
                    {item.parent_session_id ? <p className="muted">Parent session: {item.parent_session_id}</p> : null}
                    {listOrEmpty(item.memory_reads).length ? <p className="muted">Reads: {listOrEmpty(item.memory_reads).map((memory) => memory.summary).join(" • ")}</p> : null}
                    {listOrEmpty(item.memory_writes).length ? <p className="muted">Writes: {listOrEmpty(item.memory_writes).map((memory) => memory.summary).join(" • ")}</p> : null}
                  </div>
                ))}
              </div>
            </div>
          ) : null}
          {recentExecutorOutput.length ? (
            <div className="detail-card">
              <h3>Recent executor output</h3>
              <div className="nested-list">
                {recentExecutorOutput.map((item) => {
                  const payload = item.payload || {};
                  const chunkText = typeof payload.chunk_text === "string" ? payload.chunk_text : "";
                  const stream = typeof payload.stream === "string" ? payload.stream : "output";
                  const chunkIndex = typeof payload.chunk_index === "number" ? payload.chunk_index : item.seq;
                  return (
                    <div key={`${item.execution_id}-${item.seq}`} className="nested-card">
                      <div className="detail-row-header">
                        <strong>{stream}</strong>
                        <small>chunk {chunkIndex} · {formatTime(item.recorded_at)}</small>
                      </div>
                      <pre className="detail-copy" style={{ whiteSpace: "pre-wrap" }}>{chunkText || "[empty]"}</pre>
                    </div>
                  );
                })}
              </div>
            </div>
          ) : null}
          <div className="detail-card">
            <h3>Transcript slice used</h3>
            <div className="nested-list">
              {listOrEmpty(props.traceDetail.transcript_slice).map((entry) => (
                <div key={entry.id} className="nested-card">
                  <div className="detail-row-header">
                    <strong>{entry.actor_type || "actor"}</strong>
                    <small>{formatTime(entry.created_at)}</small>
                  </div>
                  <p className="detail-copy">
                    <FormattedMessage source={entry.source} text={entry.body} metadata={entry.metadata} />
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : null}

      {props.tab === "timeline" ? (
        <div className="detail-section-body">
          {listOrEmpty(trace.events).map((event) => (
            <div key={`${event.event_type}-${event.started_at}`} className="detail-row">
              <div className="detail-row-header">
                <strong>{event.event_type}</strong>
                <small>{event.status}</small>
              </div>
              <p className="detail-copy">{event.description || `${event.plane} · ${event.service} · ${event.actor}`}</p>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "reasoning" ? (
        <div className="detail-section-body">
          {listOrEmpty(trace.reasoning).map((step) => (
            <div key={step.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{step.step_type}</strong>
                <small>{step.confidence ? scoreBadge(step.confidence) : "n/a"}</small>
              </div>
              <p className="detail-copy">{step.summary}</p>
              {listOrEmpty(step.evidence_refs).length ? <p className="muted">Evidence: {listOrEmpty(step.evidence_refs).map((ref) => ref.summary || ref.ref).join(" • ")}</p> : null}
              {listOrEmpty(step.alternatives).length ? <p className="muted">Alternatives: {listOrEmpty(step.alternatives).join(" • ")}</p> : null}
              {step.decision ? <p className="muted">Decision: {step.decision}</p> : null}
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "tools" ? (
        <div className="detail-section-body">
          {listOrEmpty(trace.tool_calls).map((call) => (
            <div key={call.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{call.tool_name}</strong>
                <small>{call.approval_state || call.status}</small>
              </div>
              <p className="detail-copy">{call.summary || call.interpretation_summary}</p>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "actions" ? (
        <div className="detail-section-body">
          {listOrEmpty(traceDetail.action_intents).map((intent) => {
            const result = latestActionResult(intent.id, traceDetail.action_results);
            return (
              <div key={intent.id} className="detail-row">
                <div className="detail-row-header">
                  <strong>{intent.kind}</strong>
                  <small>{result?.status || intent.status}</small>
                </div>
                <p className="detail-copy">{intent.rationale || intent.target_ref || "No rationale recorded."}</p>
                <p className="muted">{intent.target_ref || intent.policy_verdict || "No target reference."}</p>
                {result?.provider ? <p className="muted">Provider: {result.provider}{result.provider_ref ? ` · ${result.provider_ref}` : ""}</p> : null}
                {result?.error_message ? <p className="muted">Error: {result.error_message}</p> : null}
              </div>
            );
          })}
        </div>
      ) : null}

      {props.tab === "slack" ? (
        <div className="detail-section-body">
          {listOrEmpty(trace.slack_actions).map((action) => (
            <div key={action.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{action.send_status || "draft"}</strong>
                <small>{formatTime(action.created_at)}</small>
              </div>
              <p className="detail-copy">
                <FormattedMessage source="slack" text={action.final_body || action.draft_body || ""} />
              </p>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "outcomes" ? (
        <div className="detail-section-body">
          {listOrEmpty(props.traceDetail.outcomes).map((item) => (
            <div key={item.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{item.outcome_type}</strong>
                <small>{item.verdict} · {scoreBadge(item.score)}</small>
              </div>
              <p className="detail-copy">{item.summary || item.details || "No summary."}</p>
              <p className="muted">{item.source} · {formatTime(item.recorded_at)}</p>
            </div>
          ))}
          {listOrEmpty(props.traceDetail.knowledge_entries).length ? (
            <div className="detail-card">
              <h3>Related knowledge</h3>
              <div className="nested-list">
                {listOrEmpty(props.traceDetail.knowledge_entries).map((item) => (
                  <div key={item.id} className="nested-card">
                    <div className="detail-row-header">
                      <strong>{item.title}</strong>
                      <small>{item.status} · {item.tier}</small>
                    </div>
                    <p className="detail-copy">{item.summary || item.body || "No summary."}</p>
                  </div>
                ))}
              </div>
            </div>
          ) : null}
        </div>
      ) : null}

      {props.tab === "evals" ? (
        <div className="detail-section-body">
          {listOrEmpty(props.traceDetail.linked_eval_runs).map((run) => (
            <div key={run.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{run.suite_name}</strong>
                <small>{run.overall_verdict} · {scoreBadge(run.overall_score)}</small>
              </div>
              <p className="detail-copy">Triggered by {run.trigger} at {formatTime(run.created_at)}</p>
              <div className="nested-list">
                {listOrEmpty(props.traceJudgments[run.id]).map((judgment) => (
                  <div key={judgment.id} className="nested-card">
                    <div className="detail-row-header">
                      <strong>{judgment.layer}/{judgment.category}</strong>
                      <small>{judgment.passed ? "pass" : "needs work"} · {scoreBadge(judgment.score)}</small>
                    </div>
                    <p className="detail-copy">{judgment.rationale}</p>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "feedback" ? (
        <div className="review-grid">
          <div className="detail-card">
            <h3>Recorded feedback</h3>
            <div className="nested-list">
              {listOrEmpty(props.traceDetail.feedback_records).map((item) => (
                <div key={item.id} className="nested-card">
                  <div className="detail-row-header">
                    <strong>{item.target_type}</strong>
                    <small>{item.verdict || "no verdict"} · {item.score || 0}</small>
                  </div>
                  <p className="detail-copy">{item.notes || "No notes."}</p>
                </div>
              ))}
            </div>
          </div>
          <div className="detail-card">
            <h3>Add feedback</h3>
            <label className="field">
              Target
              <select
                value={`${props.feedbackTargetType}:${props.feedbackTargetID}`}
                onChange={(event) => {
                  const [type, value] = event.target.value.split(":");
                  props.setFeedbackTargetType(type);
                  props.setFeedbackTargetID(value);
                }}
              >
                {props.feedbackTargets.map((target) => (
                  <option key={`${target.type}:${target.value}`} value={`${target.type}:${target.value}`}>{target.label}</option>
                ))}
              </select>
            </label>
            <label className="field">
              Score
              <input value={props.feedbackScore} onChange={(event) => props.setFeedbackScore(event.target.value)} />
            </label>
            <label className="field">
              Verdict
              <input value={props.feedbackVerdict} onChange={(event) => props.setFeedbackVerdict(event.target.value)} />
            </label>
            <label className="field">
              Notes
              <textarea value={props.feedbackNotes} onChange={(event) => props.setFeedbackNotes(event.target.value)} />
            </label>
            <button onClick={props.onSubmitFeedback}>Submit feedback</button>
          </div>
        </div>
      ) : null}

      {props.tab === "proposals" ? (
        <div className="detail-section-body">
          {listOrEmpty(props.traceDetail.linked_proposals).map((proposal) => (
            <div key={proposal.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{proposal.title}</strong>
                <small>{proposal.status}</small>
              </div>
              <p className="detail-copy">{proposal.summary}</p>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
