import {
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
  useCallback,
  useRef,
} from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  AlertTriangle,
  CheckCircle2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Circle,
  Database,
  FileText,
  MessageSquare,
  Search,
  Clock,
  Terminal,
  Globe,
  MessageCircle,
  Hash,
  XCircle,
  X,
  Play,
  GitBranch,
} from "lucide-react";
import { api } from "@/lib/api";
import type {
  SessionInfo,
  SessionMessage,
  SessionSearchResult,
  TraceActivityItem,
  TraceActivitySnapshot,
  TraceLedgerResponse,
} from "@/lib/api";
import { cn, timeAgo } from "@/lib/utils";
import { Markdown } from "@/components/Markdown";
import { Button } from "@nous-research/ui/ui/components/button";
import { ListItem } from "@nous-research/ui/ui/components/list-item";
import { Spinner } from "@nous-research/ui/ui/components/spinner";
import { Badge } from "@nous-research/ui/ui/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useSystemActions } from "@/contexts/useSystemActions";
import { useI18n } from "@/i18n";
import { usePageHeader } from "@/contexts/usePageHeader";
import { PluginSlot } from "@/plugins";
import { isDashboardEmbeddedChatEnabled } from "@/lib/dashboard-flags";

const SOURCE_CONFIG: Record<string, { icon: typeof Terminal; color: string }> =
  {
    cli: { icon: Terminal, color: "text-primary" },
    telegram: { icon: MessageCircle, color: "text-[oklch(0.65_0.15_250)]" },
    discord: { icon: Hash, color: "text-[oklch(0.65_0.15_280)]" },
    slack: { icon: MessageSquare, color: "text-[oklch(0.7_0.15_155)]" },
    whatsapp: { icon: Globe, color: "text-success" },
    cron: { icon: Clock, color: "text-warning" },
    trace: { icon: GitBranch, color: "text-primary" },
  };

const PAGE_SIZE = 20;
const SESSION_REFRESH_MS = 10_000;

function isTraceSession(session: SessionInfo) {
  const title = session.title?.toLowerCase() ?? "";
  return (
    session.type === "trace" ||
    Boolean(session.trace_id) ||
    session.source === "trace" ||
    session.id.startsWith("trace-") ||
    /\btrace\s+trace-/.test(title)
  );
}

function shortID(value: string | null | undefined) {
  if (!value) return "";
  return value.length > 18 ? `${value.slice(0, 18)}...` : value;
}

function displaySource(source: string | null | undefined) {
  if (!source) return "local";
  return source === "slack" ? "slack thread" : source;
}

function displayTraceKind(session: SessionInfo) {
  return (
    session.workflow_kind ??
    session.model?.split("/").pop() ??
    session.source ??
    "trace"
  );
}

function sessionMatches(session: SessionInfo, matchedIds: Set<string> | null) {
  return matchedIds === null || matchedIds.has(session.id);
}

type SessionListItem =
  | {
      kind: "group";
      id: string;
      conversation: SessionInfo;
      traces: SessionInfo[];
      lastActive: number;
    }
  | {
      kind: "session";
      id: string;
      session: SessionInfo;
      lastActive: number;
    };

function buildSessionListItems(
  sessions: SessionInfo[],
  matchedIds: Set<string> | null,
  targetTraceId: string | null,
  targetConversationId: string | null,
  targetSessionId: string | null,
): SessionListItem[] {
  const conversations = new Map<string, SessionInfo>();
  const traces: SessionInfo[] = [];

  for (const session of sessions) {
    if (isTraceSession(session)) {
      traces.push(session);
      continue;
    }
    conversations.set(session.id, session);
  }

  const groups = new Map<
    string,
    { conversation: SessionInfo; traces: SessionInfo[] }
  >();
  for (const conversation of conversations.values()) {
    groups.set(conversation.id, { conversation, traces: [] });
  }

  const orphanTraces: SessionInfo[] = [];
  for (const trace of traces) {
    const parentID = trace.parent_session_id ?? trace.conversation_id ?? null;
    const group = parentID ? groups.get(parentID) : undefined;
    if (group) {
      group.traces.push(trace);
    } else {
      orphanTraces.push(trace);
    }
  }

  const items: SessionListItem[] = [];
  for (const [id, group] of groups) {
    const parentMatched = sessionMatches(group.conversation, matchedIds);
    const matchingTraces = group.traces.filter((trace) =>
      sessionMatches(trace, matchedIds),
    );
    if (matchedIds !== null && !parentMatched && matchingTraces.length === 0) {
      continue;
    }
    const visibleTraces =
      matchedIds === null || parentMatched ? group.traces : matchingTraces;
    visibleTraces.sort((a, b) => b.last_active - a.last_active);
    const lastActive = Math.max(
      group.conversation.last_active,
      ...visibleTraces.map((trace) => trace.last_active),
    );
    if (visibleTraces.length > 0) {
      items.push({
        kind: "group",
        id,
        conversation: group.conversation,
        traces: visibleTraces,
        lastActive,
      });
    } else {
      items.push({
        kind: "session",
        id,
        session: group.conversation,
        lastActive: group.conversation.last_active,
      });
    }
  }

  for (const trace of orphanTraces) {
    if (!sessionMatches(trace, matchedIds)) continue;
    items.push({
      kind: "session",
      id: trace.id,
      session: trace,
      lastActive: trace.last_active,
    });
  }

  const targetScore = (item: SessionListItem) => {
    if (item.kind === "group") {
      if (targetConversationId && item.conversation.id === targetConversationId) {
        return 0;
      }
      if (targetSessionId && item.conversation.id === targetSessionId) {
        return 0;
      }
      if (
        targetTraceId &&
        item.traces.some((trace) => trace.id === targetTraceId)
      ) {
        return 0;
      }
      return 1;
    }
    if (targetTraceId && item.session.id === targetTraceId) return 0;
    if (targetConversationId && item.session.id === targetConversationId) {
      return 0;
    }
    if (targetSessionId && item.session.id === targetSessionId) {
      return 0;
    }
    return 1;
  };

  items.sort((a, b) => {
    const targetDelta = targetScore(a) - targetScore(b);
    if (targetDelta !== 0) return targetDelta;
    if (a.lastActive !== b.lastActive) return b.lastActive - a.lastActive;
    return a.id.localeCompare(b.id);
  });
  return items;
}

/** Render an FTS5 snippet with highlighted matches.
 *  The backend wraps matches in >>> and <<< delimiters. */
function SnippetHighlight({ snippet }: { snippet: string }) {
  const parts: React.ReactNode[] = [];
  const regex = />>>(.*?)<<</g;
  let last = 0;
  let match: RegExpExecArray | null;
  let i = 0;
  while ((match = regex.exec(snippet)) !== null) {
    if (match.index > last) {
      parts.push(snippet.slice(last, match.index));
    }
    parts.push(
      <mark key={i++} className="bg-warning/30 text-warning px-0.5">
        {match[1]}
      </mark>,
    );
    last = regex.lastIndex;
  }
  if (last < snippet.length) {
    parts.push(snippet.slice(last));
  }
  return (
    <p className="text-xs text-muted-foreground/80 truncate max-w-lg mt-0.5">
      {parts}
    </p>
  );
}

function ToolCallBlock({
  toolCall,
}: {
  toolCall: { id: string; function: { name: string; arguments: string } };
}) {
  const [open, setOpen] = useState(false);
  const { t } = useI18n();

  let args = toolCall.function.arguments;
  try {
    args = JSON.stringify(JSON.parse(args), null, 2);
  } catch {
    // keep as-is
  }

  return (
    <div className="mt-2 border border-warning/20 bg-warning/5">
      <ListItem
        onClick={() => setOpen(!open)}
        aria-label={`${open ? t.common.collapse : t.common.expand} tool call ${toolCall.function.name}`}
        aria-expanded={open}
        className="px-3 py-2 text-xs text-warning hover:bg-warning/10 hover:text-warning"
      >
        {open ? (
          <ChevronDown className="h-3 w-3" />
        ) : (
          <ChevronRight className="h-3 w-3" />
        )}
        <span className="font-mono-ui font-medium">
          {toolCall.function.name}
        </span>
        <span className="text-warning/50 ml-auto">{toolCall.id}</span>
      </ListItem>
      {open && (
        <pre className="border-t border-warning/20 px-3 py-2 text-xs text-warning/80 overflow-x-auto whitespace-pre-wrap font-mono">
          {args}
        </pre>
      )}
    </div>
  );
}

function MessageBubble({
  msg,
  highlight,
}: {
  msg: SessionMessage;
  highlight?: string;
}) {
  const { t } = useI18n();

  const ROLE_STYLES: Record<
    string,
    { bg: string; text: string; label: string }
  > = {
    user: {
      bg: "bg-primary/10",
      text: "text-primary",
      label: t.sessions.roles.user,
    },
    assistant: {
      bg: "bg-success/10",
      text: "text-success",
      label: t.sessions.roles.assistant,
    },
    system: {
      bg: "bg-muted",
      text: "text-muted-foreground",
      label: t.sessions.roles.system,
    },
    tool: {
      bg: "bg-warning/10",
      text: "text-warning",
      label: t.sessions.roles.tool,
    },
  };

  const style = ROLE_STYLES[msg.role] ?? ROLE_STYLES.system;
  const label = msg.tool_name
    ? `${t.sessions.roles.tool}: ${msg.tool_name}`
    : style.label;

  // Check if any search term appears as a prefix of any word in content
  const isHit = (() => {
    if (!highlight || !msg.content) return false;
    const content = msg.content.toLowerCase();
    const terms = highlight.toLowerCase().split(/\s+/).filter(Boolean);
    return terms.some((term) => content.includes(term));
  })();

  // Split search query into terms for inline highlighting
  const highlightTerms =
    isHit && highlight ? highlight.split(/\s+/).filter(Boolean) : undefined;

  return (
    <div
      className={`${style.bg} p-3 font-mono-ui normal-case tracking-normal ${isHit ? "ring-1 ring-warning/40" : ""}`}
      data-search-hit={isHit || undefined}
    >
      <div className="flex items-center gap-2 mb-1">
        <span className={`text-[0.7rem] font-semibold ${style.text}`}>
          {label}
        </span>
        {isHit && (
          <Badge tone="warning" className="text-[9px] py-0 px-1.5">
            {t.common.match}
          </Badge>
        )}
        {msg.timestamp && (
          <span className="text-[0.65rem] text-muted-foreground">
            {timeAgo(msg.timestamp)}
          </span>
        )}
      </div>
      {msg.content &&
        (msg.role === "system" ? (
          <div className="text-sm text-foreground whitespace-pre-wrap leading-relaxed">
            {msg.content}
          </div>
        ) : (
          <Markdown content={msg.content} highlightTerms={highlightTerms} />
        ))}
      {msg.tool_calls && msg.tool_calls.length > 0 && (
        <div className="mt-1">
          {msg.tool_calls.map((tc) => (
            <ToolCallBlock key={tc.id} toolCall={tc} />
          ))}
        </div>
      )}
    </div>
  );
}

/** Message list with auto-scroll to first search hit. */
function MessageList({
  messages,
  highlight,
  autoScrollToLatest = false,
}: {
  messages: SessionMessage[];
  highlight?: string;
  autoScrollToLatest?: boolean;
}) {
  const containerRef = useRef<HTMLDivElement>(null);
  const latestMessageKey =
    messages.length > 0
      ? [
          messages.length,
          messages[messages.length - 1]?.timestamp ?? "",
          messages[messages.length - 1]?.role ?? "",
          messages[messages.length - 1]?.content?.length ?? 0,
        ].join(":")
      : "0";

  useEffect(() => {
    if (!highlight || !containerRef.current) return;
    // Scroll to first hit after render
    const timer = setTimeout(() => {
      const hit = containerRef.current?.querySelector("[data-search-hit]");
      if (hit) {
        hit.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 50);
    return () => clearTimeout(timer);
  }, [messages, highlight]);

  useEffect(() => {
    if (highlight || !autoScrollToLatest || !containerRef.current) return;
    const timer = window.setTimeout(() => {
      const el = containerRef.current;
      if (!el) return;
      el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
    }, 50);
    return () => window.clearTimeout(timer);
  }, [autoScrollToLatest, highlight, latestMessageKey]);

  return (
    <div
      ref={containerRef}
      className="flex flex-col gap-3 max-h-[600px] overflow-y-auto pr-2"
    >
      {messages.map((msg, i) => (
        <MessageBubble key={i} msg={msg} highlight={highlight} />
      ))}
    </div>
  );
}

type TraceViewMode = "clean" | "detailed" | "raw";

function TraceActivityPanel({
  traceId,
  streamEvents,
  autoScrollToLatest = true,
}: {
  traceId: string;
  streamEvents?: boolean;
  autoScrollToLatest?: boolean;
}) {
  const [mode, setMode] = useState<TraceViewMode>("clean");
  const [snapshot, setSnapshot] = useState<TraceActivitySnapshot | null>(null);
  const [ledger, setLedger] = useState<TraceLedgerResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [rawPageLoading, setRawPageLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const foregroundLoadRef = useRef(0);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const load = useCallback(
    async (showLoading = false, signal?: AbortSignal) => {
      const foregroundLoadID = showLoading
        ? foregroundLoadRef.current + 1
        : foregroundLoadRef.current;
      if (showLoading) {
        foregroundLoadRef.current = foregroundLoadID;
        setLoading(true);
        setError(null);
      }
      try {
        if (mode === "raw") {
          const raw = await api.getTraceLedger(traceId, { limit: 250 }, signal);
          if (signal?.aborted) return;
          setLedger(raw);
          setSnapshot(null);
        } else {
          const next = await api.getTraceActivity(
            traceId,
            { mode, limit: 250 },
            signal,
          );
          if (signal?.aborted) return;
          setSnapshot(next);
          setLedger(null);
        }
        setError(null);
      } catch (err) {
        if (signal?.aborted || (err as { name?: string })?.name === "AbortError") {
          return;
        }
        setError(String(err));
      } finally {
        if (
          showLoading &&
          mountedRef.current &&
          foregroundLoadRef.current === foregroundLoadID
        ) {
          setLoading(false);
        }
      }
    },
    [mode, traceId],
  );

  const loadOlderRawLedger = useCallback(async () => {
    const before = ledger?.paging.next_before;
    if (!before || rawPageLoading) return;
    setRawPageLoading(true);
    try {
      const older = await api.getTraceLedger(traceId, { limit: 250, before });
      if (!mountedRef.current) return;
      setLedger((current) => {
        if (!current) return older;
        const seen = new Set(
          current.events
            .map(traceLedgerEventID)
            .filter((id): id is string => Boolean(id)),
        );
        const olderEvents = older.events.filter((event) => {
          const id = traceLedgerEventID(event);
          return !id || !seen.has(id);
        });
        return {
          ...current,
          events: [...olderEvents, ...current.events],
          paging: older.paging,
        };
      });
      setError(null);
    } catch (err) {
      if (mountedRef.current) {
        setError(String(err));
      }
    } finally {
      if (mountedRef.current) {
        setRawPageLoading(false);
      }
    }
  }, [ledger?.paging.next_before, rawPageLoading, traceId]);

  useEffect(() => {
    const abortController = new AbortController();
    void load(true, abortController.signal);
    return () => abortController.abort();
  }, [load]);

  useEffect(() => {
    if (!streamEvents) return;
    if (mode === "raw") {
      let closed = false;
      let refreshTimer: number | undefined;
      const abortController = new AbortController();
      const source = new EventSource(api.getTraceLedgerStreamURL(traceId));
      const scheduleRefresh = () => {
        if (closed || refreshTimer !== undefined) return;
        refreshTimer = window.setTimeout(() => {
          refreshTimer = undefined;
          void load(false, abortController.signal);
        }, 200);
      };
      source.addEventListener("ledger", scheduleRefresh);
      source.onerror = scheduleRefresh;
      return () => {
        closed = true;
        abortController.abort();
        if (refreshTimer !== undefined) window.clearTimeout(refreshTimer);
        source.close();
      };
    }

    const source = new EventSource(api.getTraceActivityStreamURL(traceId, { mode }));
    source.addEventListener("activity.snapshot", (event) => {
      try {
        const next = JSON.parse((event as MessageEvent).data) as TraceActivitySnapshot;
        setSnapshot(next);
        setError(null);
      } catch {
        // ignore malformed stream chunks; the next poll/snapshot will repair it
      }
    });
    source.addEventListener("activity.upsert", (event) => {
      try {
        const item = JSON.parse((event as MessageEvent).data) as TraceActivityItem;
        setSnapshot((current) => {
          if (!current) return current;
          const items = current.items.filter((existing) => existing.id !== item.id);
          items.push(item);
          items.sort(compareActivityItems);
          return {
            ...current,
            items,
            metrics: {
              ...current.metrics,
              item_count: items.length,
            },
          };
        });
        setError(null);
      } catch {
        // ignore malformed stream chunks; the next snapshot will repair it
      }
    });
    source.addEventListener("activity.remove", (event) => {
      try {
        const payload = JSON.parse((event as MessageEvent).data) as { id?: string };
        if (!payload.id) return;
        setSnapshot((current) => {
          if (!current) return current;
          const items = current.items.filter((existing) => existing.id !== payload.id);
          if (items.length === current.items.length) return current;
          return {
            ...current,
            items,
            metrics: {
              ...current.metrics,
              item_count: items.length,
            },
          };
        });
      } catch {
        // ignore malformed stream chunks; the next snapshot will repair it
      }
    });
    source.onerror = () => {
      // Keep the last good snapshot visible; EventSource will retry.
    };
    return () => source.close();
  }, [load, mode, streamEvents, traceId]);

  useEffect(() => {
    if (!autoScrollToLatest || !containerRef.current) return;
    const timer = window.setTimeout(() => {
      const el = containerRef.current;
      if (!el) return;
      el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
    }, 50);
    return () => window.clearTimeout(timer);
  }, [autoScrollToLatest, mode, snapshot?.items.length, ledger?.events.length]);

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-1">
          {(["clean", "detailed", "raw"] as TraceViewMode[]).map((candidate) => (
            <Button
              key={candidate}
              size="sm"
              outlined={mode !== candidate}
              onClick={() => setMode(candidate)}
              className="capitalize"
            >
              {candidate}
            </Button>
          ))}
        </div>
        {snapshot && (
          <span className="font-mono-ui text-[0.68rem] normal-case tracking-normal text-muted-foreground">
            {snapshot.metrics.item_count} items · {snapshot.metrics.ledger_event_count} ledger events · {snapshot.metrics.projection_ms}ms
          </span>
        )}
      </div>

      {loading && (
        <div className="flex items-center justify-center py-8">
          <Spinner className="text-xl text-primary" />
        </div>
      )}
      {error && !loading && (
        <p className="text-sm text-destructive py-4 text-center">{error}</p>
      )}
      {!loading && mode === "raw" && ledger && (
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between gap-2">
            <span className="font-mono-ui text-[0.68rem] normal-case tracking-normal text-muted-foreground">
              {ledger.events.length} ledger events loaded
            </span>
            {ledger.paging.has_more && ledger.paging.next_before && (
              <Button
                size="sm"
                outlined
                onClick={() => void loadOlderRawLedger()}
                disabled={rawPageLoading}
              >
                {rawPageLoading ? "Loading..." : "Load older"}
              </Button>
            )}
          </div>
          <pre
            className="max-h-[640px] overflow-auto border border-border bg-secondary/30 px-3 py-2 font-mono text-[0.68rem] leading-relaxed whitespace-pre-wrap"
          >
            {JSON.stringify(ledger, null, 2)}
          </pre>
        </div>
      )}
      {!loading && mode !== "raw" && snapshot && (
        <div
          ref={containerRef}
          className="flex max-h-[640px] flex-col gap-2 overflow-y-auto pr-2"
        >
          {snapshot.items.length === 0 ? (
            <p className="py-4 text-center text-sm text-muted-foreground">
              no activity yet
            </p>
          ) : (
            snapshot.items.map((item) => (
              <TraceActivityCard key={`${item.id}:${item.revision}`} item={item} mode={mode} />
            ))
          )}
        </div>
      )}
    </div>
  );
}

function TraceActivityCard({
  item,
  mode,
}: {
  item: TraceActivityItem;
  mode: "clean" | "detailed";
}) {
  const [open, setOpen] = useState(item.status === "failed");
  if (item.kind === "todo") {
    return <TraceTodoCard item={item} mode={mode} />;
  }
  const statusTone = activityStatusTone(item.status);
  const Icon = activityIcon(item);
  const hasDetails = mode === "detailed" && item.details && Object.keys(item.details).length > 0;
  return (
    <div className={cn("overflow-hidden border", statusTone.border, statusTone.bg)}>
      <button
        type="button"
        className="flex w-full items-start gap-2 px-3 py-2 text-left hover:bg-secondary/20"
        onClick={() => hasDetails && setOpen(!open)}
        aria-expanded={open}
      >
        {hasDetails ? (
          open ? <ChevronDown className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" /> : <ChevronRight className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        ) : (
          <span className="w-3.5 shrink-0" />
        )}
        <Icon className={cn("mt-0.5 h-3.5 w-3.5 shrink-0", statusTone.icon)} />
        <div className="min-w-0 flex-1 font-mono-ui normal-case tracking-normal">
          <div className="flex min-w-0 items-center gap-2">
            <span className="truncate text-[0.78rem] font-medium text-foreground">
              {item.title}
            </span>
            {item.duration_ms ? (
              <span className="shrink-0 text-[0.64rem] text-muted-foreground">
                {formatDuration(item.duration_ms)}
              </span>
            ) : null}
          </div>
          {item.summary && (
            <div className="mt-0.5 line-clamp-2 text-[0.72rem] text-muted-foreground">
              {item.summary}
            </div>
          )}
        </div>
        <Badge tone={activityBadgeTone(item.status)} className="shrink-0 text-[9px]">
          {displayActivityStatus(item.status)}
        </Badge>
      </button>
      {open && hasDetails && (
        <div className="border-t border-border/60 px-3 py-2">
          <TraceActivityDetails details={item.details ?? {}} />
        </div>
      )}
    </div>
  );
}

function TraceTodoCard({
  item,
  mode,
}: {
  item: TraceActivityItem;
  mode: "clean" | "detailed";
}) {
  const todos = traceTodoItems(item.details);
  return (
    <div className="border border-primary/25 bg-primary/[0.035] px-3 py-2">
      <div className="mb-2 flex items-center justify-between gap-2">
        <div className="font-mono-ui text-[0.78rem] font-medium normal-case tracking-normal">
          {item.title}
        </div>
        {item.summary && (
          <span className="text-[0.68rem] text-muted-foreground">{item.summary}</span>
        )}
      </div>
      <div className="flex flex-col gap-1.5">
        {todos.map((todo, index) => (
          <div key={todo.id || index} className="flex items-start gap-2 text-sm">
            <TodoStatusIcon status={todo.status} />
            <span
              className={cn(
                "min-w-0 flex-1 leading-snug",
                todo.status === "completed" && "text-muted-foreground line-through",
                todo.status === "cancelled" && "text-muted-foreground",
              )}
            >
              {todo.content || `Task ${index + 1}`}
            </span>
          </div>
        ))}
      </div>
      {mode === "detailed" && item.raw_event_ids.length > 0 && (
        <div className="mt-2 text-[0.65rem] text-muted-foreground">
          source: {item.raw_event_ids.join(", ")}
        </div>
      )}
    </div>
  );
}

function TraceActivityDetails({ details }: { details: Record<string, unknown> }) {
  const entries = Object.entries(details).filter(([key]) => key !== "type" && key !== "version");
  if (entries.length === 0) return null;
  return (
    <div className="grid gap-1.5 font-mono text-[0.68rem]">
      {entries.map(([key, value]) => (
        <div key={key} className="grid grid-cols-[9rem_minmax(0,1fr)] gap-2">
          <span className="text-muted-foreground">{key}</span>
          <pre className="min-w-0 whitespace-pre-wrap break-words text-foreground/85">
            {typeof value === "string" ? value : JSON.stringify(value, null, 2)}
          </pre>
        </div>
      ))}
    </div>
  );
}

function TodoStatusIcon({ status }: { status: string }) {
  if (status === "in_progress") {
    return <Spinner className="mt-0.5 h-3.5 w-3.5 shrink-0 text-primary" />;
  }
  if (status === "completed") {
    return <CheckCircle2 className="mt-0.5 h-3.5 w-3.5 shrink-0 text-success" />;
  }
  if (status === "cancelled") {
    return <XCircle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />;
  }
  return <Circle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />;
}

function traceTodoItems(details: Record<string, unknown> | undefined) {
  const raw = details?.todos;
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => {
    const record = isRecord(item) ? item : {};
    return {
      id: String(record.id ?? ""),
      content: String(record.content ?? ""),
      status: String(record.status ?? "pending"),
    };
  });
}

function traceLedgerEventID(event: Record<string, unknown>) {
  return typeof event.id === "string" ? event.id : "";
}

function compareActivityItems(a: TraceActivityItem, b: TraceActivityItem) {
  const left = activityTime(a);
  const right = activityTime(b);
  if (left !== right) return left - right;
  return a.id.localeCompare(b.id);
}

function activityTime(item: TraceActivityItem) {
  return Date.parse(item.started_at ?? item.completed_at ?? "") || 0;
}

function activityIcon(item: TraceActivityItem) {
  if (item.kind === "final_response") return MessageSquare;
  if (item.kind === "tool_generation") return Clock;
  const tool = item.tool_name ?? "";
  if (tool === "terminal") return Terminal;
  if (tool.startsWith("rsi_observability.")) return Database;
  if (tool.startsWith("rsi_sentry.")) return AlertTriangle;
  if (tool.startsWith("rsi_notion.") || tool.startsWith("rsi_knowledge.")) return FileText;
  if (tool.startsWith("rsi_slack.")) return MessageSquare;
  if (tool.startsWith("db_read.")) return Database;
  return GitBranch;
}

function activityStatusTone(status: string) {
  const normalized = displayActivityStatus(status);
  if (normalized === "failed") {
    return {
      border: "border-destructive/40",
      bg: "bg-destructive/[0.04]",
      icon: "text-destructive",
    };
  }
  if (normalized === "running") {
    return {
      border: "border-primary/35",
      bg: "bg-primary/[0.035]",
      icon: "text-primary",
    };
  }
  if (normalized === "cancelled") {
    return {
      border: "border-muted",
      bg: "bg-muted/20",
      icon: "text-muted-foreground",
    };
  }
  return {
    border: "border-border",
    bg: "bg-background/35",
    icon: "text-success",
  };
}

function activityBadgeTone(status: string): "success" | "warning" | "destructive" | "outline" {
  const normalized = displayActivityStatus(status);
  if (normalized === "failed") return "destructive";
  if (normalized === "running") return "warning";
  if (normalized === "completed") return "success";
  return "outline";
}

function displayActivityStatus(status: string) {
  const normalized = status.trim().toLowerCase();
  if (["ok", "success", "succeeded", "posted", "delivered", "complete"].includes(normalized)) {
    return "completed";
  }
  if (["blocked", "error", "failed", "failure"].includes(normalized)) return "failed";
  if (["in_progress", "streaming", "started", "pending"].includes(normalized)) return "running";
  if (["canceled", "cancelled", "denied", "expired"].includes(normalized)) return "cancelled";
  return normalized || "unknown";
}

function formatDuration(ms: number) {
  if (ms < 1000) return `${ms}ms`;
  const seconds = ms / 1000;
  if (seconds < 10) return `${seconds.toFixed(1)}s`;
  if (seconds < 60) return `${Math.round(seconds)}s`;
  return `${Math.floor(seconds / 60)}m ${Math.round(seconds % 60)}s`;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

function AnimatedSessionTitle({
  title,
  tooltip,
  className,
}: {
  title: string;
  tooltip?: string;
  className?: string;
}) {
  const stageRef = useRef<HTMLSpanElement | null>(null);
  const currentRef = useRef<HTMLSpanElement | null>(null);
  const previousTitleRef = useRef(title);

  useLayoutEffect(() => {
    const previousTitle = previousTitleRef.current;
    if (previousTitle === title) return;
    previousTitleRef.current = title;

    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      return;
    }

    const stage = stageRef.current;
    const current = currentRef.current;
    if (!stage || !current) return;

    const outgoing = document.createElement("span");
    outgoing.textContent = previousTitle;
    outgoing.setAttribute("aria-hidden", "true");
    outgoing.className =
      "session-title-exit pointer-events-none absolute inset-x-0 top-0 min-w-0 truncate";
    stage.appendChild(outgoing);

    current.classList.remove("session-title-enter");
    void current.offsetWidth;
    current.classList.add("session-title-enter");

    const timer = window.setTimeout(() => {
      outgoing.remove();
      current.classList.remove("session-title-enter");
    }, 1200);

    return () => {
      window.clearTimeout(timer);
      outgoing.remove();
      current.classList.remove("session-title-enter");
    };
  }, [title]);

  return (
    <span
      ref={stageRef}
      className={cn(
        "session-title-stage relative grid min-w-0 overflow-hidden",
        className,
      )}
      title={tooltip ?? title}
    >
      <span ref={currentRef} className="col-start-1 row-start-1 min-w-0 truncate">
        {title}
      </span>
    </span>
  );
}

function SessionTranscript({
  sessionId,
  searchQuery,
  streamEvents = false,
  autoScrollToLatest = true,
}: {
  sessionId: string;
  searchQuery?: string;
  streamEvents?: boolean;
  autoScrollToLatest?: boolean;
}) {
  const [messages, setMessages] = useState<SessionMessage[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { t } = useI18n();

  const loadMessages = useCallback(
    async (showLoading = false, signal?: AbortSignal) => {
      if (showLoading) {
        setLoading(true);
        setError(null);
      }
      try {
        const resp = await api.getSessionMessages(sessionId, signal);
        if (signal?.aborted) return;
        setMessages(resp.messages);
        setError(null);
      } catch (err) {
        if (signal?.aborted || (err as { name?: string })?.name === "AbortError") {
          return;
        }
        if (showLoading) {
          setError(String(err));
        }
      } finally {
        if (showLoading && !signal?.aborted) {
          setLoading(false);
        }
      }
    },
    [sessionId],
  );

  useEffect(() => {
    const abortController = new AbortController();
    const initialLoadId = window.setTimeout(() => {
      void loadMessages(true, abortController.signal);
    }, 0);
    return () => {
      window.clearTimeout(initialLoadId);
      abortController.abort();
    };
  }, [loadMessages]);

  useEffect(() => {
    if (!streamEvents) return;
    let closed = false;
    let refreshTimer: number | undefined;
    const abortController = new AbortController();
    const source = new EventSource(api.getSessionStreamURL(sessionId));
    const scheduleRefresh = () => {
      if (closed || refreshTimer !== undefined) return;
      refreshTimer = window.setTimeout(() => {
        refreshTimer = undefined;
        void loadMessages(false, abortController.signal);
      }, 150);
    };
    const streamEventNames = [
      "tool.start",
      "tool.progress",
      "tool.complete",
      "reasoning.delta",
      "thinking.delta",
      "message.delta",
      "status.update",
    ];
    for (const eventName of streamEventNames) {
      source.addEventListener(eventName, scheduleRefresh);
    }
    source.onmessage = scheduleRefresh;
    source.onerror = scheduleRefresh;

    return () => {
      closed = true;
      abortController.abort();
      if (refreshTimer !== undefined) {
        window.clearTimeout(refreshTimer);
      }
      for (const eventName of streamEventNames) {
        source.removeEventListener(eventName, scheduleRefresh);
      }
      source.close();
    };
  }, [loadMessages, sessionId, streamEvents]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Spinner className="text-xl text-primary" />
      </div>
    );
  }

  if (error) {
    return <p className="text-sm text-destructive py-4 text-center">{error}</p>;
  }

  if (messages && messages.length === 0) {
    return (
      <p className="text-sm text-muted-foreground py-4 text-center">
        {t.sessions.noMessages}
      </p>
    );
  }

  if (messages && messages.length > 0) {
    return (
      <MessageList
        messages={messages}
        highlight={searchQuery}
        autoScrollToLatest={autoScrollToLatest}
      />
    );
  }

  return null;
}

function SessionRow({
  session,
  snippet,
  searchQuery,
  isExpanded,
  onToggle,
  resumeInChatEnabled,
}: {
  session: SessionInfo;
  snippet?: string;
  searchQuery?: string;
  isExpanded: boolean;
  onToggle: () => void;
  resumeInChatEnabled: boolean;
}) {
  const { t } = useI18n();
  const navigate = useNavigate();

  const sourceInfo = (session.source
    ? SOURCE_CONFIG[session.source]
    : null) ?? { icon: Globe, color: "text-muted-foreground" };
  const SourceIcon = sourceInfo.icon;
  const hasTitle = session.title && session.title !== "Untitled";
  const displayTitle = hasTitle
    ? session.title!
    : session.preview
      ? session.preview.slice(0, 60)
      : t.sessions.untitledSession;

  return (
    <div
      data-session-id={session.id}
      data-expanded-session={isExpanded || undefined}
      className={`border overflow-hidden transition-colors ${
        session.is_active
          ? "border-success/30 bg-success/[0.03]"
          : "border-border"
      }`}
    >
      <div
        className="flex items-center justify-between p-3 cursor-pointer hover:bg-secondary/30 transition-colors"
        onClick={onToggle}
      >
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <div className={`shrink-0 ${sourceInfo.color}`}>
            <SourceIcon className="h-4 w-4" />
          </div>
          <div className="flex flex-col gap-0.5 min-w-0 font-mono-ui normal-case tracking-normal">
            <div className="flex items-center gap-2">
              <AnimatedSessionTitle
                title={displayTitle}
                tooltip={
                  session.title_is_summary && session.original_title
                    ? session.original_title
                    : displayTitle
                }
                className={`text-[0.8rem] truncate pr-2 ${hasTitle ? "font-medium" : "text-muted-foreground italic"}`}
              />
              {session.is_active && (
                <Badge tone="success" className="text-[10px] shrink-0">
                  <span className="mr-1 inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-current" />
                  {t.common.live}
                </Badge>
              )}
            </div>
            <div className="flex items-center gap-1.5 text-[0.72rem] text-muted-foreground">
              <span className="truncate max-w-[120px] sm:max-w-[180px]">
                {(session.model ?? t.common.unknown).split("/").pop()}
              </span>
              <span className="text-border">&#183;</span>
              <span>
                {session.message_count} {t.common.msgs}
              </span>
              {session.tool_call_count > 0 && (
                <>
                  <span className="text-border">&#183;</span>
                  <span>
                    {session.tool_call_count} {t.common.tools}
                  </span>
                </>
              )}
              <span className="text-border">&#183;</span>
              <span>{timeAgo(session.last_active)}</span>
            </div>
            {snippet && <SnippetHighlight snippet={snippet} />}
          </div>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <Badge tone="outline" className="text-[10px]">
            {session.source ?? "local"}
          </Badge>
          {resumeInChatEnabled && (
            <Button
              ghost
              size="icon"
              className="text-muted-foreground hover:text-success"
              aria-label={t.sessions.resumeInChat}
              title={t.sessions.resumeInChat}
              onClick={(e) => {
                e.stopPropagation();
                navigate(`/chat?resume=${encodeURIComponent(session.id)}`);
              }}
            >
              <Play />
            </Button>
          )}
        </div>
      </div>

      {isExpanded && (
        <div className="border-t border-border bg-background/50 p-4">
          <SessionTranscript
            sessionId={session.id}
            searchQuery={searchQuery}
            streamEvents={isTraceSession(session)}
            autoScrollToLatest
          />
        </div>
      )}
    </div>
  );
}

function TraceAttemptRow({
  trace,
  snippet,
  isExpanded,
  onToggle,
}: {
  trace: SessionInfo;
  snippet?: string;
  isExpanded: boolean;
  onToggle: () => void;
}) {
  const { t } = useI18n();
  const traceKind = displayTraceKind(trace);
  const status = trace.status ?? trace.preview ?? "trace";
  const traceTitle = trace.title?.trim() || traceKind;

  return (
    <div
      data-session-id={trace.id}
      data-expanded-session={isExpanded || undefined}
      className={`border transition-colors ${
        isExpanded
          ? "border-primary/35 bg-primary/[0.04]"
          : "border-border/80 bg-background/35"
      }`}
    >
      <button
        type="button"
        className="flex w-full items-center justify-between gap-3 px-3 py-2 text-left hover:bg-secondary/25 transition-colors"
        onClick={onToggle}
        aria-expanded={isExpanded}
      >
        <div className="flex min-w-0 items-center gap-2.5">
          <GitBranch className="h-3.5 w-3.5 shrink-0 text-primary" />
          <div className="min-w-0 font-mono-ui normal-case tracking-normal">
            <div className="flex min-w-0 items-center gap-2">
              <AnimatedSessionTitle
                title={traceTitle}
                tooltip={
                  trace.title_is_summary && trace.original_title
                    ? trace.original_title
                    : traceTitle
                }
                className="text-[0.76rem] font-medium text-foreground"
              />
              {trace.is_active && (
                <Badge tone="success" className="text-[9px] shrink-0">
                  <span className="mr-1 inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-current" />
                  {t.common.live}
                </Badge>
              )}
            </div>
            <div className="mt-0.5 flex min-w-0 items-center gap-1.5 text-[0.68rem] text-muted-foreground">
              <span className="truncate max-w-[8rem]">{traceKind}</span>
              <span className="text-border">&#183;</span>
              <span className="truncate max-w-[12rem]">
                {shortID(trace.trace_id ?? trace.id)}
              </span>
              <span className="text-border">&#183;</span>
              <span>{status}</span>
              <span className="text-border">&#183;</span>
              <span>
                {trace.message_count} {t.common.msgs}
              </span>
              {trace.tool_call_count > 0 && (
                <>
                  <span className="text-border">&#183;</span>
                  <span>
                    {trace.tool_call_count} {t.common.tools}
                  </span>
                </>
              )}
              <span className="text-border">&#183;</span>
              <span>{timeAgo(trace.last_active)}</span>
            </div>
            {snippet && <SnippetHighlight snippet={snippet} />}
          </div>
        </div>
        <Badge tone="outline" className="text-[10px] shrink-0">
          trace
        </Badge>
      </button>
      {isExpanded && (
        <div className="border-t border-border bg-background/45 p-3">
          <TraceActivityPanel
            traceId={trace.trace_id ?? trace.id}
            streamEvents
            autoScrollToLatest
          />
        </div>
      )}
    </div>
  );
}

function SessionGroupRow({
  conversation,
  traces,
  snippet,
  childSnippets,
  searchQuery,
  expandedId,
  onToggle,
}: {
  conversation: SessionInfo;
  traces: SessionInfo[];
  snippet?: string;
  childSnippets: Map<string, string>;
  searchQuery?: string;
  expandedId: string | null;
  onToggle: (sessionId: string) => void;
}) {
  const { t } = useI18n();
  const sourceInfo = (conversation.source
    ? SOURCE_CONFIG[conversation.source]
    : null) ?? { icon: Globe, color: "text-muted-foreground" };
  const SourceIcon = sourceInfo.icon;
  const hasTitle = conversation.title && conversation.title !== "Untitled";
  const displayTitle = hasTitle
    ? conversation.title!
    : conversation.preview
      ? conversation.preview.slice(0, 60)
      : t.sessions.untitledSession;
  const isGroupExpanded =
    expandedId === conversation.id ||
    traces.some((trace) => trace.id === expandedId);
  const groupActive =
    conversation.is_active || traces.some((trace) => trace.is_active);
  const visibleTraceCount = Math.max(conversation.trace_count ?? 0, traces.length);
  const openTraceCount = conversation.open_trace_count ?? 0;

  return (
    <div
      data-session-id={conversation.id}
      data-expanded-session={expandedId === conversation.id || undefined}
      className={`border overflow-hidden transition-colors ${
        groupActive
          ? "border-success/30 bg-success/[0.03]"
          : "border-border"
      } ${isGroupExpanded ? "ring-1 ring-primary/20" : ""}`}
    >
      <div
        className="flex cursor-pointer items-center justify-between gap-3 p-3 hover:bg-secondary/30 transition-colors"
        onClick={() => onToggle(conversation.id)}
      >
        <div className="flex min-w-0 flex-1 items-center gap-3">
          <div className={`shrink-0 ${sourceInfo.color}`}>
            <SourceIcon className="h-4 w-4" />
          </div>
          <div className="flex min-w-0 flex-col gap-0.5 font-mono-ui normal-case tracking-normal">
            <div className="flex min-w-0 items-center gap-2">
              <AnimatedSessionTitle
                title={displayTitle}
                tooltip={
                  conversation.title_is_summary && conversation.original_title
                    ? conversation.original_title
                    : displayTitle
                }
                className={`truncate pr-2 text-[0.8rem] ${hasTitle ? "font-medium" : "text-muted-foreground italic"}`}
              />
              {groupActive && (
                <Badge tone="success" className="text-[10px] shrink-0">
                  <span className="mr-1 inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-current" />
                  {t.common.live}
                </Badge>
              )}
            </div>
            <div className="flex min-w-0 items-center gap-1.5 text-[0.72rem] text-muted-foreground">
              <span className="truncate max-w-[120px] sm:max-w-[180px]">
                {displaySource(conversation.source)}
              </span>
              <span className="text-border">&#183;</span>
              <span>
                {conversation.message_count} {t.common.msgs}
              </span>
              {visibleTraceCount > 0 && (
                <>
                  <span className="text-border">&#183;</span>
                  <span>
                    {visibleTraceCount}{" "}
                    {visibleTraceCount === 1 ? "trace" : "traces"}
                  </span>
                </>
              )}
              {openTraceCount > 0 && (
                <>
                  <span className="text-border">&#183;</span>
                  <span>{openTraceCount} open</span>
                </>
              )}
              <span className="text-border">&#183;</span>
              <span>{timeAgo(conversation.last_active)}</span>
            </div>
            {snippet && <SnippetHighlight snippet={snippet} />}
          </div>
        </div>

        <Badge tone="outline" className="text-[10px] shrink-0">
          {displaySource(conversation.source)}
        </Badge>
      </div>

      {expandedId === conversation.id && (
        <div className="border-t border-border bg-background/50 p-4">
          <SessionTranscript
            sessionId={conversation.id}
            searchQuery={searchQuery}
            autoScrollToLatest
          />
        </div>
      )}

      {traces.length > 0 && (
        <div className="border-t border-border/80 bg-background-base/35 px-3 py-2">
          <div className="mb-2 flex items-center gap-2 font-mono-ui text-[0.66rem] uppercase tracking-normal text-muted-foreground">
            <GitBranch className="h-3 w-3" />
            <span>Trace attempts</span>
          </div>
          <div className="flex flex-col gap-1.5">
            {traces.map((trace) => (
              <TraceAttemptRow
                key={trace.id}
                trace={trace}
                snippet={childSnippets.get(trace.id)}
                isExpanded={expandedId === trace.id}
                onToggle={() => onToggle(trace.id)}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function SessionListContent({
  sessionItems,
  snippetMap,
  search,
  expandedId,
  onToggle,
  resumeInChatEnabled,
  hasSearchResults,
  total,
  page,
  pageSize,
  onPageChange,
}: {
  sessionItems: SessionListItem[];
  snippetMap: Map<string, string>;
  search: string;
  expandedId: string | null;
  onToggle: (sessionId: string) => void;
  resumeInChatEnabled: boolean;
  hasSearchResults: boolean;
  total: number;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}) {
  const { t } = useI18n();

  if (sessionItems.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <Clock className="h-8 w-8 mb-3 opacity-40" />
        <p className="text-sm font-medium">
          {search ? t.sessions.noMatch : t.sessions.noSessions}
        </p>
        {!search && (
          <p className="text-xs mt-1 text-muted-foreground/60">
            {t.sessions.startConversation}
          </p>
        )}
      </div>
    );
  }

  return (
    <>
      <div className="flex flex-col gap-1.5">
        {sessionItems.map((item) =>
          item.kind === "group" ? (
            <SessionGroupRow
              key={item.id}
              conversation={item.conversation}
              traces={item.traces}
              snippet={snippetMap.get(item.conversation.id)}
              childSnippets={snippetMap}
              searchQuery={search || undefined}
              expandedId={expandedId}
              onToggle={onToggle}
            />
          ) : isTraceSession(item.session) ? (
            <TraceAttemptRow
              key={item.id}
              trace={item.session}
              snippet={snippetMap.get(item.session.id)}
              isExpanded={expandedId === item.session.id}
              onToggle={() => onToggle(item.session.id)}
            />
          ) : (
            <SessionRow
              key={item.id}
              session={item.session}
              snippet={snippetMap.get(item.session.id)}
              searchQuery={search || undefined}
              isExpanded={expandedId === item.session.id}
              onToggle={() => onToggle(item.session.id)}
              resumeInChatEnabled={resumeInChatEnabled}
            />
          ),
        )}
      </div>

      {!hasSearchResults && total > pageSize && (
        <div className="flex items-center justify-between pt-2">
          <span className="text-xs text-muted-foreground">
            {page * pageSize + 1}-{Math.min((page + 1) * pageSize, total)}{" "}
            {t.common.of} {total}
          </span>
          <div className="flex items-center gap-1">
            <Button
              outlined
              size="icon"
              disabled={page === 0}
              onClick={() => onPageChange(page - 1)}
              aria-label={t.sessions.previousPage}
            >
              <ChevronLeft />
            </Button>
            <span className="text-xs text-muted-foreground px-2">
              {t.common.page} {page + 1} {t.common.of}{" "}
              {Math.ceil(total / pageSize)}
            </span>
            <Button
              outlined
              size="icon"
              disabled={(page + 1) * pageSize >= total}
              onClick={() => onPageChange(page + 1)}
              aria-label={t.sessions.nextPage}
            >
              <ChevronRight />
            </Button>
          </div>
        </div>
      )}
    </>
  );
}

export default function SessionsPage() {
  const [sessions, setSessions] = useState<SessionInfo[]>([]);
  const [targetSession, setTargetSession] = useState<SessionInfo | null>(null);
  const [targetParentSession, setTargetParentSession] =
    useState<SessionInfo | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [searchResults, setSearchResults] = useState<
    SessionSearchResult[] | null
  >(null);
  const [searching, setSearching] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(null);
  const sessionsInFlightCountRef = useRef(0);
  const pendingSessionsLoadRef = useRef<{
    page: number;
    showLoading: boolean;
    signal?: AbortSignal;
  } | null>(null);
  const loadSessionsRef = useRef<
    ((p: number, showLoading?: boolean, signal?: AbortSignal) => void) | null
  >(null);
  const logScrollRef = useRef<HTMLPreElement | null>(null);
  const pageRef = useRef<HTMLDivElement | null>(null);
  const appliedDeepLinkScrollRef = useRef<string | null>(null);
  const { t } = useI18n();
  const { setAfterTitle, setEnd } = usePageHeader();
  const { activeAction, actionStatus, dismissLog } = useSystemActions();
  const resumeInChatEnabled = isDashboardEmbeddedChatEnabled();
  const [searchParams, setSearchParams] = useSearchParams();

  const targetTraceId = useMemo(
    () => searchParams.get("trace")?.trim() || null,
    [searchParams],
  );
  const targetConversationId = useMemo(() => {
    for (const key of ["conversation", "conversation_id"]) {
      const value = searchParams.get(key)?.trim();
      if (value) return value;
    }
    return null;
  }, [searchParams]);
  const targetSessionId = useMemo(() => {
    const explicitSession =
      searchParams.get("session")?.trim() ||
      searchParams.get("session_id")?.trim() ||
      null;
    return targetTraceId ?? explicitSession ?? targetConversationId;
  }, [searchParams, targetConversationId, targetTraceId]);

  useLayoutEffect(() => {
    if (loading) {
      setAfterTitle(null);
      setEnd(null);
      return;
    }
    setAfterTitle(
      <Badge tone="secondary" className="text-xs tabular-nums">
        {total}
      </Badge>,
    );
    setEnd(
      <div className="relative w-full min-w-0 sm:max-w-xs">
        {searching ? (
          <Spinner className="absolute left-2.5 top-1/2 -translate-y-1/2 text-[0.875rem] text-primary" />
        ) : (
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
        )}
        <Input
          placeholder={t.sessions.searchPlaceholder}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="h-8 pr-7 pl-8 text-xs"
        />
        {search && (
          <Button
            ghost
            size="xs"
            className="absolute right-1.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            onClick={() => setSearch("")}
            aria-label={t.common.clear}
          >
            <X />
          </Button>
        )}
      </div>,
    );
    return () => {
      setAfterTitle(null);
      setEnd(null);
    };
  }, [
    loading,
    search,
    searching,
    setAfterTitle,
    setEnd,
    t.common.clear,
    t.sessions.searchPlaceholder,
    total,
  ]);

  const flushPendingSessionsLoad = useCallback(() => {
    if (sessionsInFlightCountRef.current > 0) {
      return;
    }
    const pending = pendingSessionsLoadRef.current;
    if (!pending) {
      return;
    }
    pendingSessionsLoadRef.current = null;
    if (pending.signal?.aborted) {
      return;
    }
    loadSessionsRef.current?.(
      pending.page,
      pending.showLoading,
      pending.signal,
    );
  }, []);

  const loadSessions = useCallback(
    (p: number, showLoading = true, signal?: AbortSignal) => {
      if (sessionsInFlightCountRef.current > 0) {
        if (showLoading) {
          pendingSessionsLoadRef.current = { page: p, showLoading, signal };
          setLoading(true);
        }
        return;
      }
      sessionsInFlightCountRef.current += 1;
      if (showLoading) {
        setLoading(true);
      }
      api
        .getSessions(PAGE_SIZE, p * PAGE_SIZE, signal)
        .then((resp) => {
          if (signal?.aborted) return;
          setSessions(resp.sessions);
          setTotal(resp.total);
        })
        .catch((err) => {
          if (err?.name !== "AbortError") {
            // Silently ignore non-abort errors
          }
        })
        .finally(() => {
          sessionsInFlightCountRef.current = Math.max(
            0,
            sessionsInFlightCountRef.current - 1,
          );
          if (!signal?.aborted && showLoading) {
            setLoading(false);
          }
          flushPendingSessionsLoad();
        });
    },
    [flushPendingSessionsLoad],
  );

  useEffect(() => {
    loadSessionsRef.current = loadSessions;
  }, [loadSessions]);

  useEffect(() => {
    const abortController = new AbortController();
    loadSessions(page, true, abortController.signal);
    const refreshSessions = () => {
      if (document.visibilityState !== "visible") {
        return;
      }
      loadSessions(page, false, abortController.signal);
    };
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        loadSessions(page, false, abortController.signal);
      }
    };
    const intervalId = window.setInterval(refreshSessions, SESSION_REFRESH_MS);
    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () => {
      abortController.abort();
      window.clearInterval(intervalId);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [loadSessions, page]);

  useEffect(() => {
    if (!targetSessionId) {
      setTargetSession(null);
      return;
    }
    const existing = sessions.find((session) => session.id === targetSessionId);
    if (existing) {
      setTargetSession(existing);
    }
  }, [sessions, targetSessionId]);

  useEffect(() => {
    if (!targetSessionId) {
      return;
    }
    const expandTimer = window.setTimeout(() => {
      setExpandedId(targetSessionId);
    }, 0);
    return () => window.clearTimeout(expandTimer);
  }, [targetSessionId]);

  useEffect(() => {
    if (!targetSessionId) {
      return;
    }
    const existing = sessions.find((session) => session.id === targetSessionId);
    if (existing || targetSession?.id === targetSessionId) {
      return;
    }
    let cancelled = false;
    api
      .getSession(targetSessionId)
      .then((session) => {
        if (!cancelled) setTargetSession(session);
      })
      .catch(() => {
        if (!cancelled) setTargetSession(null);
      });
    return () => {
      cancelled = true;
    };
  }, [targetSessionId]);

  const relatedParentSessionId = useMemo(() => {
    if (targetConversationId) return targetConversationId;
    if (targetSession && isTraceSession(targetSession)) {
      return targetSession.parent_session_id ?? targetSession.conversation_id ?? null;
    }
    return null;
  }, [targetConversationId, targetSession]);

  useEffect(() => {
    if (!relatedParentSessionId || relatedParentSessionId === targetSessionId) {
      setTargetParentSession(null);
      return;
    }
    const existing = sessions.find(
      (session) => session.id === relatedParentSessionId,
    );
    if (existing) {
      setTargetParentSession(existing);
    }
  }, [relatedParentSessionId, sessions, targetSessionId]);

  useEffect(() => {
    if (!relatedParentSessionId || relatedParentSessionId === targetSessionId) {
      return;
    }
    const existing = sessions.find(
      (session) => session.id === relatedParentSessionId,
    );
    if (existing || targetParentSession?.id === relatedParentSessionId) {
      return;
    }
    let cancelled = false;
    api
      .getSession(relatedParentSessionId)
      .then((session) => {
        if (!cancelled) setTargetParentSession(session);
      })
      .catch(() => {
        if (!cancelled) setTargetParentSession(null);
      });
    return () => {
      cancelled = true;
    };
  }, [relatedParentSessionId, targetSessionId]);

  useEffect(() => {
    const el = logScrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [actionStatus?.lines]);

  // Debounced FTS search
  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    const query = search.trim();
    let cancelled = false;

    debounceRef.current = setTimeout(() => {
      if (!query) {
        setSearchResults(null);
        setSearching(false);
        return;
      }
      setSearching(true);
      api
        .searchSessions(query)
        .then((resp) => {
          if (!cancelled) setSearchResults(resp.results);
        })
        .catch(() => {
          if (!cancelled) setSearchResults(null);
        })
        .finally(() => {
          if (!cancelled) setSearching(false);
        });
    }, query ? 300 : 0);

    return () => {
      cancelled = true;
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [search]);

  // Build snippet map from search results (session_id -> snippet)
  const snippetMap = useMemo(() => {
    const out = new Map<string, string>();
    if (searchResults) {
      for (const result of searchResults) {
        out.set(result.session_id, result.snippet);
      }
    }
    return out;
  }, [searchResults]);

  const orderedSessions = useMemo(() => {
    const byID = new Map<string, SessionInfo>();
    for (const session of sessions) {
      byID.set(session.id, session);
    }
    if (
      targetSession &&
      targetSession.id === targetSessionId &&
      !byID.has(targetSession.id)
    ) {
      byID.set(targetSession.id, targetSession);
    }
    if (
      targetParentSession &&
      targetParentSession.id === relatedParentSessionId &&
      !byID.has(targetParentSession.id)
    ) {
      byID.set(targetParentSession.id, targetParentSession);
    }
    return Array.from(byID.values());
  }, [
    relatedParentSessionId,
    sessions,
    targetParentSession,
    targetSession,
    targetSessionId,
  ]);

  const sessionById = useMemo(() => {
    const out = new Map<string, SessionInfo>();
    for (const session of orderedSessions) {
      out.set(session.id, session);
    }
    return out;
  }, [orderedSessions]);

  const writeSelectedSessionToURL = useCallback(
    (session: SessionInfo | null) => {
      setSearchParams(
        (current) => {
          const next = new URLSearchParams(current);
          next.delete("conversation");
          next.delete("conversation_id");
          next.delete("trace");
          next.delete("session");
          next.delete("session_id");

          if (!session) {
            return next;
          }

          if (isTraceSession(session)) {
            const conversationId =
              session.parent_session_id?.trim() ||
              session.conversation_id?.trim() ||
              null;
            if (conversationId) {
              next.set("conversation", conversationId);
            }
            next.set("trace", session.trace_id?.trim() || session.id);
            return next;
          }

          if (session.type === "conversation" || session.conversation_id) {
            next.set(
              "conversation",
              session.conversation_id?.trim() || session.id,
            );
            return next;
          }

          next.set("session", session.id);
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  const handleToggleSession = useCallback(
    (sessionId: string) => {
      const nextExpandedId = expandedId === sessionId ? null : sessionId;
      setExpandedId(nextExpandedId);
      writeSelectedSessionToURL(
        nextExpandedId ? (sessionById.get(nextExpandedId) ?? null) : null,
      );
    },
    [expandedId, sessionById, writeSelectedSessionToURL],
  );

  const sessionItems = useMemo(
    () =>
      buildSessionListItems(
        orderedSessions,
        searchResults ? new Set(snippetMap.keys()) : null,
        targetTraceId,
        targetConversationId,
        targetSessionId,
      ),
    [
      orderedSessions,
      searchResults,
      snippetMap,
      targetConversationId,
      targetTraceId,
      targetSessionId,
    ],
  );

  useEffect(() => {
    if (!targetSessionId) {
      appliedDeepLinkScrollRef.current = null;
      return;
    }
    if (loading || !expandedId || expandedId !== targetSessionId) return;
    const scrollKey = `${targetSessionId}:${expandedId}`;
    if (appliedDeepLinkScrollRef.current === scrollKey) return;

    const timer = window.setTimeout(() => {
      const target = pageRef.current?.querySelector<HTMLElement>(
        '[data-expanded-session="true"]',
      );
      if (!target) return;
      target.scrollIntoView({ behavior: "smooth", block: "start" });
      appliedDeepLinkScrollRef.current = scrollKey;
    }, 100);

    return () => window.clearTimeout(timer);
  }, [expandedId, loading, sessionItems, targetSessionId]);

  const recentSessions =
    page === 0 && !searchResults
      ? sessions
          .filter((s) => !isTraceSession(s))
          .filter((s) => !s.is_active)
          .slice(0, 5)
      : [];

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Spinner className="text-2xl text-primary" />
      </div>
    );
  }

  return (
    <div ref={pageRef} className="flex flex-col gap-4">
      <PluginSlot name="sessions:top" />

      {activeAction && (
        <div className="border border-border bg-background-base/50">
          <div className="flex items-center justify-between gap-2 border-b border-border px-3 py-2">
            <div className="flex items-center gap-2 min-w-0">
              {actionStatus?.running ? (
                <Spinner className="shrink-0 text-[0.875rem] text-warning" />
              ) : actionStatus?.exit_code === 0 ? (
                <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-success" />
              ) : actionStatus !== null ? (
                <AlertTriangle className="h-3.5 w-3.5 shrink-0 text-destructive" />
              ) : (
                <Spinner className="shrink-0 text-[0.875rem] text-muted-foreground" />
              )}

              <span className="text-xs font-mono-ui tracking-normal normal-case truncate">
                {activeAction === "restart"
                  ? t.status.restartGateway
                  : t.status.updateHermes}
              </span>

              <Badge
                tone={
                  actionStatus?.running
                    ? "warning"
                    : actionStatus?.exit_code === 0
                      ? "success"
                      : actionStatus
                        ? "destructive"
                        : "outline"
                }
                className="text-[10px] shrink-0"
              >
                {actionStatus?.running
                  ? t.status.running
                  : actionStatus?.exit_code === 0
                    ? t.status.actionFinished
                    : actionStatus
                      ? `${t.status.actionFailed} (${actionStatus.exit_code ?? "?"})`
                      : t.common.loading}
              </Badge>
            </div>

            <Button
              ghost
              size="icon"
              onClick={dismissLog}
              className="shrink-0 opacity-60 hover:opacity-100"
              aria-label={t.common.close}
            >
              <X />
            </Button>
          </div>

          <pre
            ref={logScrollRef}
            className="max-h-72 overflow-auto px-3 py-2 font-mono-ui text-[11px] leading-relaxed whitespace-pre-wrap break-all"
          >
            {actionStatus?.lines && actionStatus.lines.length > 0
              ? actionStatus.lines.join("\n")
              : t.status.waitingForOutput}
          </pre>
        </div>
      )}

      <SessionListContent
        sessionItems={sessionItems}
        snippetMap={snippetMap}
        search={search}
        expandedId={expandedId}
        onToggle={handleToggleSession}
        resumeInChatEnabled={resumeInChatEnabled}
        hasSearchResults={Boolean(searchResults)}
        total={total}
        page={page}
        pageSize={PAGE_SIZE}
        onPageChange={setPage}
      />

      {page === 0 && recentSessions.length > 0 && (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Clock className="h-5 w-5 text-muted-foreground" />
              <CardTitle className="text-base">
                {t.status.recentSessions}
              </CardTitle>
            </div>
          </CardHeader>

          <CardContent className="grid gap-3">
            {recentSessions.map((s) => (
              <div
                key={s.id}
                className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 border border-border p-3 w-full"
              >
                <div className="flex flex-col gap-1 min-w-0 w-full font-mono-ui normal-case tracking-normal">
                  <AnimatedSessionTitle
                    title={s.title ?? t.common.untitled}
                    tooltip={
                      s.title_is_summary && s.original_title
                        ? s.original_title
                        : (s.title ?? t.common.untitled)
                    }
                    className="font-medium text-[0.8rem]"
                  />

                  <span className="text-[0.72rem] text-muted-foreground truncate">
                    <span>{displaySource(s.source)}</span> · {s.message_count}{" "}
                    {t.common.msgs}
                    {(s.trace_count ?? 0) > 0 && (
                      <>
                        {" "}
                        · {s.trace_count}{" "}
                        {s.trace_count === 1 ? "trace" : "traces"}
                      </>
                    )}{" "}
                    · {timeAgo(s.last_active)}
                  </span>

                  {s.preview && (
                    <span className="text-[0.72rem] text-muted-foreground/70 truncate">
                      {s.preview}
                    </span>
                  )}
                </div>

                <Badge
                  tone="outline"
                  className="text-[10px] shrink-0 self-start sm:self-center"
                >
                  <Database className="mr-1 h-3 w-3" />
                  {displaySource(s.source)}
                </Badge>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
      <PluginSlot name="sessions:bottom" />
    </div>
  );
}
