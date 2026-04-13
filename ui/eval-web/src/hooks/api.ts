import type {
  ActionResult,
  KnowledgeEntry,
  KnowledgeSegment,
  NullableList,
  PRAttempt,
  RepoChangeJob,
  TabKey,
  ViewState
} from "@/types";

const TAB_KEYS: TabKey[] = ["conversations", "cases", "proposals", "knowledge"];

export function readViewState(): ViewState {
  const params = new URLSearchParams(window.location.search);
  const tabRaw = params.get("tab");
  const tab: TabKey = tabRaw && TAB_KEYS.includes(tabRaw as TabKey) ? (tabRaw as TabKey) : "conversations";
  return {
    tab,
    conversation: params.get("conversation") ?? undefined,
    case: params.get("case") ?? undefined,
    trace: params.get("trace") ?? undefined,
    proposal: params.get("proposal") ?? undefined,
    knowledge: params.get("knowledge") ?? undefined
  };
}

export function writeViewState(state: ViewState): void {
  const params = new URLSearchParams();
  params.set("tab", state.tab);
  if (state.conversation) params.set("conversation", state.conversation);
  if (state.case) params.set("case", state.case);
  if (state.trace) params.set("trace", state.trace);
  if (state.proposal) params.set("proposal", state.proposal);
  if (state.knowledge) params.set("knowledge", state.knowledge);
  const qs = params.toString();
  const url = qs ? `${window.location.pathname}?${qs}` : window.location.pathname;
  window.history.pushState(null, "", url);
}

export function getJSON<T>(url: string): Promise<T> {
  return fetch(url).then(async (response) => {
    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }
    return response.json();
  });
}

export function postJSON<T>(url: string, body: Record<string, unknown>): Promise<T> {
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

export function listOrEmpty<T>(items: NullableList<T> | undefined): T[] {
  return items ?? [];
}

export function recordOrEmpty<T>(items: Record<string, T> | undefined): Record<string, T> {
  return items ?? {};
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

export function actionResultsForIntent(intentId: string, results: NullableList<ActionResult> | undefined) {
  return listOrEmpty(results).filter((item) => item.action_intent_id === intentId);
}

export function latestActionResult(intentId: string, results: NullableList<ActionResult> | undefined) {
  return actionResultsForIntent(intentId, results)[0];
}

export function knowledgeEntriesForSegment(entries: KnowledgeEntry[], segment: KnowledgeSegment): KnowledgeEntry[] {
  return entries.filter((entry) => {
    switch (segment) {
      case "working":
        return entry.tier === "working";
      case "review":
        return entry.status === "review_pending";
      case "canonical":
        return entry.tier === "canonical" || entry.status === "canonical";
      case "stale":
        return entry.status === "stale";
      default:
        return true;
    }
  });
}

export function proposalJobState(proposalId: string, jobs: NullableList<RepoChangeJob> | undefined): RepoChangeJob | undefined {
  return listOrEmpty(jobs).find((job) => job.proposal_id === proposalId);
}

export function proposalPRState(proposalId: string, attempts: NullableList<PRAttempt> | undefined): PRAttempt | undefined {
  const rows = listOrEmpty(attempts).filter((attempt) => attempt.proposal_id === proposalId);
  if (rows.length === 0) return undefined;
  return rows.reduce((latest, attempt) => (attempt.created_at > latest.created_at ? attempt : latest));
}
