import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";

import type {
  TabKey,
  ProposalSegment,
  KnowledgeSegment,
  TraceInspectorTab,
  ConversationListItem,
  CaseSummary,
  ConversationDetailResponse,
  CaseDetailResponse,
  TraceDetailResponse,
  EvalJudgment,
  KnowledgeEntry,
  NullableList,
  ActionResult,
  ProposalResponse,
  ProposalDetailResponse,
  KnowledgeListResponse,
  KnowledgeDetailResponse,
  RuntimeResponse,
  HarnessResponse,
  ViewState,
} from "@/types";

import { formatTime, listOrEmpty, scoreBadge } from "@/hooks/api";

import { EmptyDetail } from "@/components/detail/empty-detail";
import { ConversationDetail } from "@/components/detail/conversation-detail";
import { CaseDetail } from "@/components/detail/case-detail";
import { ProposalDetail } from "@/components/detail/proposal-detail";
import { KnowledgeDetail } from "@/components/detail/knowledge-detail";
import { HarnessDetail } from "@/components/detail/harness-detail";

const ACTIVE_PROPOSAL_STATES = new Set([
  "pending_review",
  "approved",
  "repo_change_queued",
  "repo_change_running",
  "validation_pending",
  "pr_open"
]);

function recordOrEmpty<T>(items: Record<string, T> | undefined): Record<string, T> {
  return items ?? {};
}

function getJSON<T>(url: string): Promise<T> {
  return fetch(url).then(async (response) => {
    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }
    return response.json();
  });
}

function postJSON<T>(url: string, body: Record<string, unknown>): Promise<T> {
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

function readViewState(): ViewState {
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

function writeViewState(next: ViewState) {
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

function knowledgeEntriesForSegment(entries: KnowledgeEntry[], segment: KnowledgeSegment) {
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

export function App() {
  const queryClient = useQueryClient();
  const [viewState, setViewState] = useState<ViewState>(() => readViewState());
  const [proposalSegment, setProposalSegment] = useState<ProposalSegment>("active");
  const [knowledgeSegment, setKnowledgeSegment] = useState<KnowledgeSegment>("working");
  const [traceInspectorTab, setTraceInspectorTab] = useState<TraceInspectorTab>("overview");
  const [proposalCapInput, setProposalCapInput] = useState("2");
  const [feedbackTargetType, setFeedbackTargetType] = useState("trace");
  const [feedbackTargetID, setFeedbackTargetID] = useState("");
  const [feedbackScore, setFeedbackScore] = useState("3");
  const [feedbackVerdict, setFeedbackVerdict] = useState("useful");
  const [feedbackNotes, setFeedbackNotes] = useState("");
  const [proposalRationale, setProposalRationale] = useState("");
  const [knowledgeReviewRationale, setKnowledgeReviewRationale] = useState("");

  useEffect(() => {
    const handlePopState = () => setViewState(readViewState());
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  const navigate = (next: ViewState) => {
    writeViewState(next);
    setViewState(next);
  };

  const conversationsQuery = useQuery({
    queryKey: ["conversations"],
    queryFn: () => getJSON<{ conversations: ConversationListItem[] }>("/api/conversations")
  });

  const casesQuery = useQuery({
    queryKey: ["cases"],
    queryFn: () => getJSON<{ cases: CaseSummary[] }>("/api/cases")
  });

  const proposalsQuery = useQuery({
    queryKey: ["proposals"],
    queryFn: () => getJSON<ProposalResponse>("/api/proposals")
  });

  const knowledgeQuery = useQuery({
    queryKey: ["knowledge"],
    queryFn: () => getJSON<KnowledgeListResponse>("/api/knowledge")
  });

  const runtimeQuery = useQuery({
    queryKey: ["runtime"],
    queryFn: () => getJSON<RuntimeResponse>("/api/runtime")
  });

  const harnessQuery = useQuery({
    queryKey: ["harness"],
    queryFn: () => getJSON<HarnessResponse>("/api/harness")
  });

  const conversationDetailQuery = useQuery({
    queryKey: ["conversation", viewState.conversation],
    queryFn: () => getJSON<ConversationDetailResponse>(`/api/conversations/${viewState.conversation}`),
    enabled: Boolean(viewState.tab === "conversations" && viewState.conversation)
  });

  const caseDetailQuery = useQuery({
    queryKey: ["case", viewState.case],
    queryFn: () => getJSON<CaseDetailResponse>(`/api/cases/${viewState.case}`),
    enabled: Boolean(viewState.tab === "cases" && viewState.case)
  });

  const traceDetailQuery = useQuery({
    queryKey: ["trace", viewState.trace],
    queryFn: () => getJSON<TraceDetailResponse>(`/api/traces/${viewState.trace}`),
    enabled: Boolean(viewState.trace)
  });

  const proposalDetailQuery = useQuery({
    queryKey: ["proposal", viewState.proposal],
    queryFn: () => getJSON<ProposalDetailResponse>(`/api/proposals/${viewState.proposal}`),
    enabled: Boolean(viewState.tab === "proposals" && viewState.proposal)
  });

  const knowledgeDetailQuery = useQuery({
    queryKey: ["knowledge", viewState.knowledge],
    queryFn: () => getJSON<KnowledgeDetailResponse>(`/api/knowledge/${viewState.knowledge}`),
    enabled: Boolean(viewState.tab === "knowledge" && viewState.knowledge)
  });

  const conversations = listOrEmpty(conversationsQuery.data?.conversations);
  const cases = listOrEmpty(casesQuery.data?.cases);
  const proposals = listOrEmpty(proposalsQuery.data?.proposals);
  const knowledgeEntries = listOrEmpty(knowledgeQuery.data?.knowledge_entries);
  const candidates = listOrEmpty(proposalsQuery.data?.candidates);
  const runtimeRoles = listOrEmpty(runtimeQuery.data?.roles);
  const harnessRoles = listOrEmpty(harnessQuery.data?.roles);
  const proposalSlotState = proposalsQuery.data?.proposal_slots;

  useEffect(() => {
    const settingValue = proposalsQuery.data?.settings?.active_proposal_cap;
    if (typeof settingValue === "number") {
      setProposalCapInput(String(settingValue));
    }
  }, [proposalsQuery.data?.settings?.active_proposal_cap]);

  useEffect(() => {
    if (viewState.conversation && !conversations.some((item) => item.conversation_id === viewState.conversation)) {
      navigate({ tab: "conversations" });
    }
  }, [viewState.conversation, conversations]);

  useEffect(() => {
    if (viewState.case && !cases.some((item) => item.case_id === viewState.case)) {
      navigate({ tab: "cases" });
    }
  }, [viewState.case, cases]);

  useEffect(() => {
    if (viewState.proposal && !proposals.some((item) => item.id === viewState.proposal)) {
      navigate({ tab: "proposals" });
    }
  }, [viewState.proposal, proposals]);

  useEffect(() => {
    if (viewState.knowledge && !knowledgeEntries.some((item) => item.id === viewState.knowledge)) {
      navigate({ tab: "knowledge" });
    }
  }, [viewState.knowledge, knowledgeEntries]);

  useEffect(() => {
    if (viewState.role && !harnessRoles.some((item) => item.role === viewState.role)) {
      navigate({ tab: "harness" });
    }
  }, [viewState.role, harnessRoles]);

  const refreshEverything = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["conversations"] }),
      queryClient.invalidateQueries({ queryKey: ["conversation"] }),
      queryClient.invalidateQueries({ queryKey: ["cases"] }),
      queryClient.invalidateQueries({ queryKey: ["case"] }),
      queryClient.invalidateQueries({ queryKey: ["trace"] }),
      queryClient.invalidateQueries({ queryKey: ["proposals"] }),
      queryClient.invalidateQueries({ queryKey: ["proposal"] }),
      queryClient.invalidateQueries({ queryKey: ["knowledge"] }),
      queryClient.invalidateQueries({ queryKey: ["runtime"] }),
      queryClient.invalidateQueries({ queryKey: ["harness"] })
    ]);
  };

  const evaluateMutation = useMutation({
    mutationFn: () => postJSON(`/api/traces/${viewState.trace}/evaluate`, {}),
    onSuccess: refreshEverything
  });

  const replayMutation = useMutation({
    mutationFn: () => postJSON(`/api/traces/${viewState.trace}/replay`, { requested_by: "ui-operator" }),
    onSuccess: refreshEverything
  });

  const feedbackMutation = useMutation({
    mutationFn: () =>
      postJSON(`/api/feedback`, {
        target_type: feedbackTargetType,
        target_id: feedbackTargetID,
        score: Number(feedbackScore),
        verdict: feedbackVerdict,
        labels: ["operator-ui"],
        notes: feedbackNotes,
        reviewer_id: "ui-operator"
      }),
    onSuccess: async () => {
      setFeedbackNotes("");
      await refreshEverything();
    }
  });

  const knowledgeReviewMutation = useMutation({
    mutationFn: (decision: string) =>
      postJSON(`/api/knowledge/${viewState.knowledge}/review`, {
        decision,
        rationale: knowledgeReviewRationale || `UI operator recorded ${decision}.`,
        reviewer_id: "ui-operator"
      }),
    onSuccess: async () => {
      setKnowledgeReviewRationale("");
      await refreshEverything();
    }
  });

  const promoteMutation = useMutation({
    mutationFn: () => postJSON(`/api/proposals/promote`, { requested_by: "ui-operator" }),
    onSuccess: refreshEverything
  });

  const settingsMutation = useMutation({
    mutationFn: () => postJSON(`/api/settings`, { active_proposal_cap: Number(proposalCapInput) }),
    onSuccess: refreshEverything
  });

  const proposalDecisionMutation = useMutation({
    mutationFn: (decision: string) =>
      postJSON(`/api/proposals/${viewState.proposal}/decision`, {
        decision,
        rationale: proposalRationale || `UI operator recorded ${decision}.`,
        reviewer_id: "ui-operator",
        failure_class: decision === "rejected" ? "insufficient_evidence" : ""
      }),
    onSuccess: async () => {
      setProposalRationale("");
      await refreshEverything();
    }
  });

  const proposalRetryMutation = useMutation({
    mutationFn: () =>
      postJSON(`/api/proposals/${viewState.proposal}/retry`, {
        requested_by: "ui-operator"
      }),
    onSuccess: refreshEverything
  });

  const proposalStopMutation = useMutation({
    mutationFn: () =>
      postJSON(`/api/proposals/${viewState.proposal}/stop`, {
        requested_by: "ui-operator",
        rationale: proposalRationale || "UI operator stopped the remediation line."
      }),
    onSuccess: async () => {
      setProposalRationale("");
      await refreshEverything();
    }
  });

  const activeProposals = useMemo(
    () => proposals.filter((proposal) => ACTIVE_PROPOSAL_STATES.has(proposal.status)),
    [proposals]
  );
  const historyProposals = useMemo(
    () => proposals.filter((proposal) => !ACTIVE_PROPOSAL_STATES.has(proposal.status)),
    [proposals]
  );
  const proposalRows = proposalSegment === "active" ? activeProposals : proposalSegment === "history" ? historyProposals : [];
  const knowledgeRows = useMemo(
    () => knowledgeEntriesForSegment(knowledgeEntries, knowledgeSegment),
    [knowledgeEntries, knowledgeSegment]
  );

  const traceDetail = traceDetailQuery.data;
  const feedbackTargets = useMemo(() => {
    if (!traceDetail) return [];
    const targets = [{ label: `Trace ${traceDetail.trace.summary.trace_id}`, value: traceDetail.trace.summary.trace_id, type: "trace" }];
    for (const step of listOrEmpty(traceDetail.trace.reasoning)) {
      targets.push({ label: `Reasoning: ${step.step_type}`, value: step.id, type: "reasoning_step" });
    }
    for (const call of listOrEmpty(traceDetail.trace.tool_calls)) {
      targets.push({ label: `Tool: ${call.tool_name}`, value: call.id, type: "tool_call" });
    }
    for (const intent of listOrEmpty(traceDetail.action_intents)) {
      targets.push({ label: `Action: ${intent.kind}`, value: intent.id, type: "action_intent" });
    }
    for (const action of listOrEmpty(traceDetail.trace.slack_actions)) {
      targets.push({ label: `Slack action ${formatTime(action.created_at)}`, value: action.id, type: "slack_action" });
    }
    return targets;
  }, [traceDetail]);

  useEffect(() => {
    if (feedbackTargets.length > 0) {
      setFeedbackTargetType(feedbackTargets[0].type);
      setFeedbackTargetID(feedbackTargets[0].value);
    }
  }, [traceDetail?.trace.summary.trace_id]);

  const traceJudgments = recordOrEmpty(traceDetail?.judgments_by_eval_run);

  return (
    <div className="app-shell">
      <aside className="nav-rail">
        <div className="brand-block">
          <p className="eyebrow">Improvement Plane</p>
          <h1>Evidence-first operator workspace</h1>
          <p className="muted">
            Start from conversations, move into cases, and inspect the exact trace evidence that produced a proposal.
          </p>
        </div>

        <nav className="tab-nav" aria-label="Sections">
          <button className={viewState.tab === "conversations" ? "tab-button active" : "tab-button"} onClick={() => navigate({ tab: "conversations" })}>
            <span>Conversations</span>
            <strong>{conversations.length}</strong>
          </button>
          <button className={viewState.tab === "cases" ? "tab-button active" : "tab-button"} onClick={() => navigate({ tab: "cases" })}>
            <span>Cases</span>
            <strong>{cases.length}</strong>
          </button>
          <button className={viewState.tab === "proposals" ? "tab-button active" : "tab-button"} onClick={() => navigate({ tab: "proposals" })}>
            <span>Proposals</span>
            <strong>{proposals.length}</strong>
          </button>
          <button className={viewState.tab === "knowledge" ? "tab-button active" : "tab-button"} onClick={() => navigate({ tab: "knowledge" })}>
            <span>Knowledge</span>
            <strong>{knowledgeEntries.length}</strong>
          </button>
          <button className={viewState.tab === "harness" ? "tab-button active" : "tab-button"} onClick={() => navigate({ tab: "harness" })}>
            <span>Harness</span>
            <strong>{harnessRoles.length}</strong>
          </button>
        </nav>

        <section className="operations-card">
          <div className="section-header">
            <div>
              <p className="eyebrow">Operations</p>
              <h2>Proposal cap</h2>
            </div>
            <span className="status-chip">{proposalSlotState?.active ?? 0}/{proposalSlotState?.cap ?? 0}</span>
          </div>
          <dl className="slot-grid">
            <div><dt>Active</dt><dd>{proposalSlotState?.active ?? 0}</dd></div>
            <div><dt>Available</dt><dd>{proposalSlotState?.available ?? 0}</dd></div>
            <div><dt>Stale</dt><dd>{listOrEmpty(proposalSlotState?.stale_proposal_ids).length}</dd></div>
            <div><dt>Candidates</dt><dd>{candidates.length}</dd></div>
          </dl>
          <label className="field">
            Active proposal cap
            <input type="number" min={1} value={proposalCapInput} onChange={(event) => setProposalCapInput(event.target.value)} />
          </label>
          <div className="button-row">
            <button onClick={() => settingsMutation.mutate()} disabled={settingsMutation.isPending}>Save cap</button>
            <button className="secondary" onClick={() => promoteMutation.mutate()} disabled={promoteMutation.isPending || (proposalSlotState?.available ?? 0) === 0}>
              Run promoter
            </button>
          </div>
        </section>

        <section className="operations-card">
          <div className="section-header">
            <div>
              <p className="eyebrow">Runtime</p>
              <h2>Model runtime</h2>
            </div>
          </div>
          <div className="runtime-list">
            {runtimeRoles.map((role) => (
              <div key={role.role} className="runtime-row">
                <div>
                  <strong>{role.role}</strong>
                  <p>{role.backend || "unreachable"} · {role.provider || "n/a"}</p>
                </div>
                <div className="runtime-meta">
                  <span className={role.healthy ? "status-dot ok" : "status-dot"} />
                  <small>{role.model}</small>
                  <small>{role.reasoning_effort}</small>
                </div>
              </div>
            ))}
          </div>
        </section>
      </aside>

      <section className="list-pane">
        {viewState.tab === "conversations" ? (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Conversations</p>
                <h2>Slack threads, DMs, and incident rooms</h2>
              </div>
            </header>
            <div className="list-stack">
              {conversations.map((item) => (
                <button
                  key={item.conversation_id}
                  className={item.conversation_id === viewState.conversation ? "list-card selected" : "list-card"}
                  onClick={() => navigate({ tab: "conversations", conversation: item.conversation_id, trace: item.active_case?.latest_trace_id })}
                >
                  <div className="list-card-header">
                    <div>
                      <strong>{item.title}</strong>
                      <p>{item.source} · {item.status}</p>
                    </div>
                    {item.latest_trace_verdict ? <span className="status-chip eval">{item.latest_trace_verdict}</span> : null}
                  </div>
                  <p className="trace-thread">{item.external_key}</p>
                  <dl className="mini-metrics">
                    <div><dt>Active case</dt><dd>{item.active_case?.title || "none"}</dd></div>
                    <div><dt>Latest</dt><dd>{formatTime(item.latest_message_at)}</dd></div>
                    <div><dt>Open traces</dt><dd>{item.open_trace_count}</dd></div>
                    <div><dt>Proposals</dt><dd>{item.proposal_count}</dd></div>
                  </dl>
                </button>
              ))}
            </div>
          </>
        ) : viewState.tab === "cases" ? (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Cases</p>
                <h2>Cross-conversation objectives</h2>
              </div>
            </header>
            <div className="list-stack">
              {cases.map((item) => (
                <button
                  key={item.case_id}
                  className={item.case_id === viewState.case ? "list-card selected" : "list-card"}
                  onClick={() => navigate({ tab: "cases", case: item.case_id, trace: item.latest_trace_id })}
                >
                  <div className="list-card-header">
                    <div>
                      <strong>{item.title}</strong>
                      <p>{item.kind} · {item.status}</p>
                    </div>
                    {item.latest_trace_verdict ? <span className="status-chip eval">{item.latest_trace_verdict}</span> : null}
                  </div>
                  <p className="trace-thread">{item.summary}</p>
                  <dl className="mini-metrics">
                    <div><dt>Conversation</dt><dd>{item.conversation_id}</dd></div>
                    <div><dt>Bot</dt><dd>{item.assigned_bot}</dd></div>
                    <div><dt>Recurrence</dt><dd>{item.recurrence}</dd></div>
                    <div><dt>Proposals</dt><dd>{listOrEmpty(item.linked_proposal_ids).length}</dd></div>
                  </dl>
                </button>
              ))}
            </div>
          </>
        ) : viewState.tab === "proposals" ? (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Proposals</p>
                <h2>Review path and PR readiness</h2>
              </div>
              <div className="segment-row">
                {(["active", "candidates", "history"] as ProposalSegment[]).map((segment) => (
                  <button key={segment} className={proposalSegment === segment ? "segment-button active" : "segment-button"} onClick={() => setProposalSegment(segment)}>
                    {segment}
                  </button>
                ))}
              </div>
            </header>
            {proposalSegment === "candidates" ? (
              <div className="list-stack">
                {candidates.map((candidate) => (
                  <div key={candidate.id} className="list-card static">
                    <div className="list-card-header">
                      <div>
                        <strong>{candidate.subsystem}</strong>
                        <p>{candidate.failure_mode}</p>
                      </div>
                      <span className="status-chip">{scoreBadge(candidate.priority_score)}</span>
                    </div>
                    <dl className="mini-metrics">
                      <div><dt>Status</dt><dd>{candidate.status}</dd></div>
                      <div><dt>Severity</dt><dd>{candidate.severity}</dd></div>
                      <div><dt>Recurrence</dt><dd>{candidate.recurrence_count}</dd></div>
                      <div><dt>Latest trace</dt><dd>{candidate.latest_trace_id || "none"}</dd></div>
                    </dl>
                  </div>
                ))}
              </div>
            ) : (
              <div className="list-stack">
                {proposalRows.map((proposal) => {
                  return (
                    <button
                      key={proposal.id}
                      className={proposal.id === viewState.proposal ? "list-card selected" : "list-card"}
                      onClick={() => navigate({ tab: "proposals", proposal: proposal.id })}
                    >
                      <div className="list-card-header">
                        <div>
                          <strong>{proposal.title}</strong>
                          <p>{proposal.status} · {proposal.candidate_key}</p>
                        </div>
                        <span className="status-chip">{proposal.pr_status || proposal.repo_change_status || proposal.status}</span>
                      </div>
                      <p className="trace-thread">{proposal.summary}</p>
                    <dl className="mini-metrics">
                      <div><dt>Risk</dt><dd>{proposal.risk_tier || "n/a"}</dd></div>
                      <div><dt>Target</dt><dd>{proposal.target_layer || "repo_change"}</dd></div>
                      <div><dt>Slot</dt><dd>{proposal.active_slot_consuming ? "occupied" : "free"}</dd></div>
                      <div><dt>PR</dt><dd>{proposal.pr_url ? "linked" : "none"}</dd></div>
                    </dl>
                    </button>
                  );
                })}
              </div>
            )}
          </>
        ) : viewState.tab === "harness" ? (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Harness</p>
                <h2>Persistent Hermes role agents</h2>
              </div>
            </header>
            <div className="list-stack">
              {harnessRoles.map((role) => {
                const bindings = listOrEmpty(harnessQuery.data?.session_bindings).filter((item) => item.role === role.role);
                const overlays = listOrEmpty(harnessQuery.data?.overlays).filter((item) => item.role === role.role && item.status === "active");
                const executions = listOrEmpty(harnessQuery.data?.executions).filter((item) => item.role === role.role);
                return (
                  <button
                    key={role.role}
                    className={role.role === viewState.role ? "list-card selected" : "list-card"}
                    onClick={() => navigate({ tab: "harness", role: role.role })}
                  >
                    <div className="list-card-header">
                      <div>
                        <strong>{role.role}</strong>
                        <p>{role.provider || "n/a"} · {role.status}</p>
                      </div>
                      <span className="status-chip">{role.honcho_available ? "honcho" : "memory off"}</span>
                    </div>
                    <p className="trace-thread">{role.model} · {role.reasoning_effort}</p>
                    <dl className="mini-metrics">
                      <div><dt>Sessions</dt><dd>{bindings.length}</dd></div>
                      <div><dt>Runs</dt><dd>{executions.length}</dd></div>
                      <div><dt>Overlay</dt><dd>{overlays[0]?.version || role.active_overlay_version || "baseline"}</dd></div>
                      <div><dt>Persistence</dt><dd>{role.persistence_enabled ? "on" : "off"}</dd></div>
                    </dl>
                  </button>
                );
              })}
            </div>
          </>
        ) : (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Knowledge</p>
                <h2>Working drafts and canonical memory</h2>
              </div>
              <div className="segment-row">
                {(["working", "review", "canonical", "stale"] as KnowledgeSegment[]).map((segment) => (
                  <button key={segment} className={knowledgeSegment === segment ? "segment-button active" : "segment-button"} onClick={() => setKnowledgeSegment(segment)}>
                    {segment}
                  </button>
                ))}
              </div>
            </header>
            <div className="list-stack">
              {knowledgeRows.map((entry) => (
                <button
                  key={entry.id}
                  className={entry.id === viewState.knowledge ? "list-card selected" : "list-card"}
                  onClick={() => navigate({ tab: "knowledge", knowledge: entry.id })}
                >
                  <div className="list-card-header">
                    <div>
                      <strong>{entry.title}</strong>
                      <p>{entry.kind} · {entry.scope_type}</p>
                    </div>
                    <span className="status-chip">{entry.status}</span>
                  </div>
                  <p className="trace-thread">{entry.summary || entry.body || "No summary."}</p>
                  <dl className="mini-metrics">
                    <div><dt>Tier</dt><dd>{entry.tier}</dd></div>
                    <div><dt>Confidence</dt><dd>{scoreBadge(entry.confidence)}</dd></div>
                    <div><dt>Source</dt><dd>{entry.source_type}</dd></div>
                    <div><dt>Updated</dt><dd>{formatTime(entry.updated_at)}</dd></div>
                  </dl>
                </button>
              ))}
            </div>
          </>
        )}
      </section>

      <section className="detail-pane">
        {viewState.tab === "conversations" ? (
          !viewState.conversation ? (
            <EmptyDetail title="Select a conversation" body="Start from the conversation list. Once selected, you'll see transcript context, case continuity, trace attempts, and the evidence behind the latest run." />
          ) : conversationDetailQuery.isLoading ? (
            <EmptyDetail title="Loading conversation" body="Fetching transcript, cases, and trace attempts." />
          ) : conversationDetailQuery.data ? (
            <ConversationDetail
              detail={conversationDetailQuery.data}
              selectedTraceId={viewState.trace}
              traceDetail={traceDetail}
              traceInspectorTab={traceInspectorTab}
              setTraceInspectorTab={setTraceInspectorTab}
              onSelectTrace={(traceId) => navigate({ tab: "conversations", conversation: viewState.conversation, trace: traceId })}
              onRunEval={() => evaluateMutation.mutate()}
              onReplay={() => replayMutation.mutate()}
              traceJudgments={traceJudgments}
              feedbackTargets={feedbackTargets}
              feedbackTargetType={feedbackTargetType}
              setFeedbackTargetType={setFeedbackTargetType}
              feedbackTargetID={feedbackTargetID}
              setFeedbackTargetID={setFeedbackTargetID}
              feedbackScore={feedbackScore}
              setFeedbackScore={setFeedbackScore}
              feedbackVerdict={feedbackVerdict}
              setFeedbackVerdict={setFeedbackVerdict}
              feedbackNotes={feedbackNotes}
              setFeedbackNotes={setFeedbackNotes}
              onSubmitFeedback={() => feedbackMutation.mutate()}
            />
          ) : (
            <EmptyDetail title="Conversation not found" body="The selected conversation no longer exists." />
          )
        ) : viewState.tab === "cases" ? (
          !viewState.case ? (
            <EmptyDetail title="Select a case" body="Cases are the active objectives inside conversations. Pick one to inspect its attempts and current evidence." />
          ) : caseDetailQuery.isLoading ? (
            <EmptyDetail title="Loading case" body="Fetching case summary and associated traces." />
          ) : caseDetailQuery.data ? (
            <CaseDetail
              detail={caseDetailQuery.data}
              selectedTraceId={viewState.trace}
              traceDetail={traceDetail}
              traceInspectorTab={traceInspectorTab}
              setTraceInspectorTab={setTraceInspectorTab}
              onSelectTrace={(traceId) => navigate({ tab: "cases", case: viewState.case, trace: traceId })}
              onRunEval={() => evaluateMutation.mutate()}
              onReplay={() => replayMutation.mutate()}
              traceJudgments={traceJudgments}
              feedbackTargets={feedbackTargets}
              feedbackTargetType={feedbackTargetType}
              setFeedbackTargetType={setFeedbackTargetType}
              feedbackTargetID={feedbackTargetID}
              setFeedbackTargetID={setFeedbackTargetID}
              feedbackScore={feedbackScore}
              setFeedbackScore={setFeedbackScore}
              feedbackVerdict={feedbackVerdict}
              setFeedbackVerdict={setFeedbackVerdict}
              feedbackNotes={feedbackNotes}
              setFeedbackNotes={setFeedbackNotes}
              onSubmitFeedback={() => feedbackMutation.mutate()}
            />
          ) : (
            <EmptyDetail title="Case not found" body="The selected case no longer exists." />
          )
        ) : viewState.tab === "proposals" ? (!viewState.proposal ? (
          <EmptyDetail title="Select a proposal" body="Proposals remain the review and PR-path surface. Select one to inspect reasoning, memory, evidence traces, and linked PR state." />
        ) : proposalDetailQuery.isLoading ? (
          <EmptyDetail title="Loading proposal" body="Fetching proposal reviews, memory, repo-change state, and linked traces." />
        ) : proposalDetailQuery.data ? (
          <ProposalDetail
            detail={proposalDetailQuery.data}
            proposalRationale={proposalRationale}
            setProposalRationale={setProposalRationale}
            onDecision={(decision) => proposalDecisionMutation.mutate(decision)}
            onRetry={() => proposalRetryMutation.mutate()}
            onStop={() => proposalStopMutation.mutate()}
            canRetry={["approved", "repo_change_queued", "failed_validation", "validation_pending"].includes(proposalDetailQuery.data.proposal.status)}
            canStop={ACTIVE_PROPOSAL_STATES.has(proposalDetailQuery.data.proposal.status)}
          />
        ) : (
          <EmptyDetail title="Proposal not found" body="The selected proposal no longer exists." />
        )) : viewState.tab === "harness" ? (
          !viewState.role ? (
            <EmptyDetail title="Select a harness role" body="Choose a role agent to inspect Hermes session continuity, overlays, experiments, and memory activity." />
          ) : harnessQuery.isLoading ? (
            <EmptyDetail title="Loading harness state" body="Fetching role agent sessions, overlays, and experiments." />
          ) : (
            <HarnessDetail detail={harnessQuery.data} selectedRole={viewState.role} />
          )
        ) : !viewState.knowledge ? (
          <EmptyDetail title="Select a knowledge entry" body="Knowledge tracks working drafts, canonical guidance, contradictions, and the evidence behind each entry." />
        ) : knowledgeDetailQuery.isLoading ? (
          <EmptyDetail title="Loading knowledge" body="Fetching provenance links and review history." />
        ) : knowledgeDetailQuery.data ? (
          <KnowledgeDetail
            detail={knowledgeDetailQuery.data}
            reviewRationale={knowledgeReviewRationale}
            setReviewRationale={setKnowledgeReviewRationale}
            onDecision={(decision) => knowledgeReviewMutation.mutate(decision)}
          />
        ) : (
          <EmptyDetail title="Knowledge entry not found" body="The selected knowledge entry no longer exists." />
        )}
      </section>
    </div>
  );
}
