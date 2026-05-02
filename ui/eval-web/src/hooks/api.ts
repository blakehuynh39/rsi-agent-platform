import type { ActionResult, JsonObject, KnowledgeEntry, KnowledgeSegment, NullableList, TabKey, ViewState } from "@/types";

type CommandRequest = {
  command_kind: string;
  command_id?: string;
  causation_id?: string;
  actor?: string;
  occurred_at?: string;
  expected_version?: number;
  payload?: JsonObject;
};

export function listOrEmpty<T>(items: NullableList<T> | undefined): T[] {
  return items ?? [];
}

export function formatTime(value?: string) {
  if (!value) return "Unknown";
  const date = new Date(value);
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit"
  }).format(date);
}

export function scoreBadge(score?: number) {
  if (typeof score !== "number") return "n/a";
  return score.toFixed(2);
}

export function getJSON<T>(url: string): Promise<T> {
  return fetch(url).then(async (response) => {
    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }
    return response.json();
  });
}

export function postJSON<T>(url: string, body: JsonObject): Promise<T> {
  return fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  }).then(async (response) => {
    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }
    return response.json();
  });
}

export function postCommand<T>(url: string, body: CommandRequest): Promise<T> {
  return postJSON(url, body as JsonObject);
}

export function readViewState(): ViewState {
  const params = new URLSearchParams(window.location.search);
  const tabValue = params.get("tab");
  const tab: TabKey =
    tabValue === "cases" ? "cases" :
    tabValue === "proposals" ? "proposals" :
    tabValue === "knowledge" ? "knowledge" :
    tabValue === "harness" ? "harness" :
    "conversations";
  return {
    tab,
    conversation: params.get("conversation") || undefined,
    case: params.get("case") || undefined,
    trace: params.get("trace") || undefined,
    proposal: params.get("proposal") || undefined,
    knowledge: params.get("knowledge") || undefined,
    role: params.get("role") || undefined
  };
}

export function writeViewState(next: ViewState) {
  const params = new URLSearchParams();
  params.set("tab", next.tab);
  if (next.conversation) params.set("conversation", next.conversation);
  if (next.case) params.set("case", next.case);
  if (next.trace) params.set("trace", next.trace);
  if (next.proposal) params.set("proposal", next.proposal);
  if (next.knowledge) params.set("knowledge", next.knowledge);
  if (next.role) params.set("role", next.role);
  const query = params.toString();
  const target = `${window.location.pathname}${query ? `?${query}` : ""}`;
  window.history.replaceState({}, "", target);
}

export function knowledgeEntriesForSegment(entries: KnowledgeEntry[], segment: KnowledgeSegment) {
  switch (segment) {
    case "working":
      return entries.filter((item) => item.tier === "working" && item.status === "draft");
    case "review":
      return entries.filter((item) => item.status === "review_pending");
    case "canonical":
      return entries.filter((item) => item.status === "canonical" || item.tier === "canonical");
    case "stale":
      return entries.filter((item) => ["stale", "contradicted", "archived"].includes(item.status));
    default:
      return entries;
  }
}

export function latestActionResult(intentId: string, results: NullableList<ActionResult> | undefined) {
  return listOrEmpty(results).filter((item) => item.action_intent_id === intentId)[0];
}

export function pageCount(total: number, pageSize: number) {
  return Math.max(1, Math.ceil(total / pageSize));
}

export function clampPage(page: number, total: number, pageSize: number) {
  return Math.min(Math.max(page, 1), pageCount(total, pageSize));
}
