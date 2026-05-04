import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";

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
  AppDataResetResponse,
  RuntimeResponse,
  HarnessResponse,
  ViewState,
} from "@/types";

import { formatTime, getJSON, knowledgeEntriesForSegment, listOrEmpty, postCommand, postJSON, readViewState, scoreBadge, writeViewState, pageCount, clampPage } from "@/hooks/api";

import { EmptyDetail } from "@/components/detail/empty-detail";
import { ConversationDetail } from "@/components/detail/conversation-detail";
import { CaseDetail } from "@/components/detail/case-detail";
import { ProposalDetail } from "@/components/detail/proposal-detail";
import { KnowledgeDetail } from "@/components/detail/knowledge-detail";
import { HarnessDetail } from "@/components/detail/harness-detail";
import { Icon } from "@/components/icon";

const ACTIVE_PROPOSAL_STATES = new Set([
  "pending_review",
  "approved",
  "in_progress",
  "needs_review"
]);

const RUNNING_TRACE_STATES = new Set([
  "active",
  "executing",
  "in_progress",
  "pending",
  "processing",
  "queued",
  "running",
  "started"
]);

const RUNNING_LIST_STATES = new Set([
  "executing",
  "in_progress",
  "pending",
  "processing",
  "queued",
  "running",
  "started"
]);

const CONVERSATION_DETAIL_QUERY = "include=cases,traces,workflows,transcript,proposals,self_review&transcript_limit=50";
const CONVERSATION_PAGE_SIZE = 6;

function dateSortValue(value?: string) {
  if (!value) {
    return 0;
  }
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? 0 : parsed;
}

function normalizedStatus(value?: string) {
  return (value || "").toLowerCase().replace(/\s+/g, "_");
}

function isRunningListStatus(value?: string) {
  return RUNNING_LIST_STATES.has(normalizedStatus(value));
}

function hasTraceVerdict(value?: string) {
  return Boolean(value?.trim());
}

function isLiveConversation(item: ConversationListItem) {
  if (hasTraceVerdict(item.latest_trace_verdict) || hasTraceVerdict(item.active_case?.latest_trace_verdict)) {
    return false;
  }
  return item.open_trace_count > 0 || isRunningListStatus(item.status) || isRunningListStatus(item.active_case?.status);
}

function isLiveCase(item: CaseSummary) {
  if (hasTraceVerdict(item.latest_trace_verdict)) {
    return false;
  }
  return isRunningListStatus(item.status);
}

function isLiveProposal(status: string) {
  return ACTIVE_PROPOSAL_STATES.has(normalizedStatus(status));
}

function ActivityGlyph(props: { active: boolean; label: string }) {
  return (
    <span className={props.active ? "activity-glyph live" : "activity-glyph"} aria-label={props.label} title={props.label}>
      <span />
    </span>
  );
}

function countLabel(loaded: boolean, count: number) {
  return loaded ? String(count) : "...";
}

function defaultTraceInspectorTabForStatus(status?: string): TraceInspectorTab {
  return status && RUNNING_TRACE_STATES.has(normalizedStatus(status)) ? "raw" : "summary";
}

function canRetryProposal(detail: ProposalDetailResponse) {
  if (detail.proposal.status === "approved") {
    return true;
  }
  if (detail.proposal.status !== "needs_review") {
    return false;
  }
  return Boolean(detail.current_phase?.attempt_id || detail.proposal.current_attempt_id);
}

function knowledgeCommandKind(decision: string) {
  switch (decision) {
    case "approve":
      return "knowledge_approve";
    case "reject":
      return "knowledge_reject";
    case "mark_stale":
      return "knowledge_mark_stale";
    case "archive":
      return "knowledge_archive";
    default:
      throw new Error(`Unsupported knowledge decision: ${decision}`);
  }
}

function proposalCommandKind(decision: string) {
  switch (decision) {
    case "approved":
      return "proposal_approve_intervention";
    case "rejected":
      return "proposal_reject_line";
    case "dismissed":
      return "proposal_dismiss_line";
    case "merged":
      return "proposal_mark_merged";
    default:
      throw new Error(`Unsupported proposal decision: ${decision}`);
  }
}

export function App() {
  const queryClient = useQueryClient();
  const [viewState, setViewState] = useState<ViewState>(() => readViewState());
  const [proposalSegment, setProposalSegment] = useState<ProposalSegment>("active");
  const [knowledgeSegment, setKnowledgeSegment] = useState<KnowledgeSegment>("working");
  const [traceInspectorTab, setTraceInspectorTab] = useState<TraceInspectorTab>("summary");
  const defaultedTraceTabRef = useRef<string | undefined>();
  const [proposalCapInput, setProposalCapInput] = useState("2");
  const [feedbackTargetType, setFeedbackTargetType] = useState("trace");
  const [feedbackTargetID, setFeedbackTargetID] = useState("");
  const [feedbackScore, setFeedbackScore] = useState("3");
  const [feedbackVerdict, setFeedbackVerdict] = useState("useful");
  const [feedbackNotes, setFeedbackNotes] = useState("");
  const [proposalRationale, setProposalRationale] = useState("");
  const [knowledgeReviewRationale, setKnowledgeReviewRationale] = useState("");
  const [loadSecondaryData, setLoadSecondaryData] = useState(false);
  const [conversationPage, setConversationPage] = useState(1);
  const [planeOpen, setPlaneOpen] = useState(false);

  useEffect(() => {
    const handlePopState = () => setViewState(readViewState());
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  useEffect(() => {
    if (!planeOpen) {
      return;
    }
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setPlaneOpen(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [planeOpen]);

  const navigate = (next: ViewState) => {
    writeViewState(next);
    setViewState(next);
  };

  const wantsConversations = viewState.tab === "conversations" || Boolean(viewState.conversation) || loadSecondaryData;
  const wantsCases = viewState.tab === "cases" || Boolean(viewState.case) || loadSecondaryData;
  const wantsProposals = viewState.tab === "proposals" || Boolean(viewState.proposal) || loadSecondaryData;
  const wantsKnowledge = viewState.tab === "knowledge" || Boolean(viewState.knowledge) || loadSecondaryData;
  const wantsHarness = viewState.tab === "harness" || Boolean(viewState.role) || loadSecondaryData;

  const conversationsQuery = useQuery({
    queryKey: ["conversations"],
    queryFn: () => getJSON<{ conversations: ConversationListItem[] }>("/api/conversations"),
    enabled: wantsConversations
  });

  const casesQuery = useQuery({
    queryKey: ["cases"],
    queryFn: () => getJSON<{ cases: CaseSummary[] }>("/api/cases"),
    enabled: wantsCases
  });

  const proposalsQuery = useQuery({
    queryKey: ["proposals"],
    queryFn: () => getJSON<ProposalResponse>("/api/proposals"),
    enabled: wantsProposals
  });

  const knowledgeQuery = useQuery({
    queryKey: ["knowledge"],
    queryFn: () => getJSON<KnowledgeListResponse>("/api/knowledge"),
    enabled: wantsKnowledge
  });

  const runtimeQuery = useQuery({
    queryKey: ["runtime"],
    queryFn: () => getJSON<RuntimeResponse>("/api/runtime"),
    enabled: wantsHarness
  });

  const harnessQuery = useQuery({
    queryKey: ["harness"],
    queryFn: () => getJSON<HarnessResponse>("/api/harness"),
    enabled: wantsHarness
  });

  const conversationDetailQuery = useQuery({
    queryKey: ["conversation", viewState.conversation],
    queryFn: () => getJSON<ConversationDetailResponse>(`/api/conversations/${viewState.conversation}?${CONVERSATION_DETAIL_QUERY}`),
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

  useEffect(() => {
    if (loadSecondaryData) {
      return;
    }
    const waitingForConversation =
      viewState.tab === "conversations" &&
      Boolean(viewState.conversation) &&
      !conversationDetailQuery.isFetched &&
      conversationDetailQuery.fetchStatus !== "idle";
    const waitingForTrace = Boolean(viewState.trace) && !traceDetailQuery.isFetched && traceDetailQuery.fetchStatus !== "idle";
    if (waitingForConversation || waitingForTrace) {
      return;
    }
    const timeout = window.setTimeout(() => setLoadSecondaryData(true), 400);
    return () => window.clearTimeout(timeout);
  }, [
    loadSecondaryData,
    viewState.tab,
    viewState.conversation,
    viewState.trace,
    conversationDetailQuery.isFetched,
    conversationDetailQuery.fetchStatus,
    traceDetailQuery.isFetched,
    traceDetailQuery.fetchStatus
  ]);

  useEffect(() => {
    if (!viewState.trace) {
      defaultedTraceTabRef.current = undefined;
      setTraceInspectorTab("summary");
      return;
    }
    const traceSummary = traceDetailQuery.data?.trace.summary;
    if (!traceSummary || traceSummary.trace_id !== viewState.trace || defaultedTraceTabRef.current === viewState.trace) {
      return;
    }
    setTraceInspectorTab(defaultTraceInspectorTabForStatus(traceSummary.status));
    defaultedTraceTabRef.current = viewState.trace;
  }, [traceDetailQuery.data, viewState.trace]);

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
  const sortedConversations = useMemo(
    () => [...conversations].sort((left, right) => dateSortValue(right.latest_message_at) - dateSortValue(left.latest_message_at)),
    [conversations]
  );
  const conversationPageCount = pageCount(sortedConversations.length, CONVERSATION_PAGE_SIZE);
  const clampedConversationPage = Math.min(conversationPage, conversationPageCount);
  const conversationPageStart = sortedConversations.length === 0 ? 0 : (clampedConversationPage - 1) * CONVERSATION_PAGE_SIZE + 1;
  const conversationPageEnd = Math.min(sortedConversations.length, clampedConversationPage * CONVERSATION_PAGE_SIZE);
  const visibleConversations = useMemo(
    () => sortedConversations.slice((clampedConversationPage - 1) * CONVERSATION_PAGE_SIZE, clampedConversationPage * CONVERSATION_PAGE_SIZE),
    [clampedConversationPage, sortedConversations]
  );

  useEffect(() => {
    setConversationPage((current) => clampPage(current, sortedConversations.length, CONVERSATION_PAGE_SIZE));
  }, [sortedConversations.length]);

  useEffect(() => {
    if (!viewState.conversation) {
      return;
    }
    const selectedIndex = sortedConversations.findIndex((item) => item.conversation_id === viewState.conversation);
    if (selectedIndex === -1) {
      return;
    }
    setConversationPage(Math.floor(selectedIndex / CONVERSATION_PAGE_SIZE) + 1);
  }, [sortedConversations, viewState.conversation]);

  useEffect(() => {
    const settingValue = proposalsQuery.data?.settings?.active_proposal_cap;
    if (typeof settingValue === "number") {
      setProposalCapInput(String(settingValue));
    }
  }, [proposalsQuery.data?.settings?.active_proposal_cap]);

  useEffect(() => {
    if (viewState.conversation && conversationsQuery.isSuccess && !conversations.some((item) => item.conversation_id === viewState.conversation)) {
      navigate({ tab: "conversations" });
    }
  }, [viewState.conversation, conversations, conversationsQuery.isSuccess]);

  useEffect(() => {
    if (viewState.case && casesQuery.isSuccess && !cases.some((item) => item.case_id === viewState.case)) {
      navigate({ tab: "cases" });
    }
  }, [viewState.case, cases, casesQuery.isSuccess]);

  useEffect(() => {
    if (viewState.proposal && proposalsQuery.isSuccess && !proposals.some((item) => item.id === viewState.proposal)) {
      navigate({ tab: "proposals" });
    }
  }, [viewState.proposal, proposals, proposalsQuery.isSuccess]);

  useEffect(() => {
    if (viewState.knowledge && knowledgeQuery.isSuccess && !knowledgeEntries.some((item) => item.id === viewState.knowledge)) {
      navigate({ tab: "knowledge" });
    }
  }, [viewState.knowledge, knowledgeEntries, knowledgeQuery.isSuccess]);

  useEffect(() => {
    if (viewState.role && harnessQuery.isSuccess && !harnessRoles.some((item) => item.role === viewState.role)) {
      navigate({ tab: "harness" });
    }
  }, [viewState.role, harnessRoles, harnessQuery.isSuccess]);

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
    mutationFn: () =>
      postCommand(`/api/problem-lines/${viewState.trace}/commands`, {
        command_kind: "problem_line_evaluate_trace",
        actor: "ui-operator",
        payload: {
          trigger: "manual"
        }
      }),
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
      postCommand(`/api/knowledge/${viewState.knowledge}/commands`, {
        command_kind: knowledgeCommandKind(decision),
        actor: "ui-operator",
        payload: {
          rationale: knowledgeReviewRationale || `UI operator recorded ${decision}.`,
          reviewer_id: "ui-operator"
        }
      }),
    onSuccess: async () => {
      setKnowledgeReviewRationale("");
      await refreshEverything();
    }
  });

  const promoteMutation = useMutation({
    mutationFn: () =>
      postCommand(`/api/problem-lines/problem-lines/commands`, {
        command_kind: "problem_line_promote",
        actor: "ui-operator",
        payload: {
          requested_by: "ui-operator"
        }
      }),
    onSuccess: refreshEverything
  });

  const settingsMutation = useMutation({
    mutationFn: () =>
      postCommand(`/api/settings/commands`, {
        command_kind: "settings_update",
        actor: "ui-operator",
        payload: {
          active_proposal_cap: Number(proposalCapInput)
        }
      }),
    onSuccess: refreshEverything
  });

  const resetAppDataMutation = useMutation({
    mutationFn: () => postJSON<AppDataResetResponse>(`/api/app-data/reset`, {}),
    onSuccess: async () => {
      navigate({ tab: "conversations" });
      await refreshEverything();
    }
  });

  const proposalDecisionMutation = useMutation({
    mutationFn: (decision: string) =>
      postCommand(`/api/proposals/${viewState.proposal}/commands`, {
        command_kind: proposalCommandKind(decision),
        actor: "ui-operator",
        payload: {
          scope: "line",
          rationale: proposalRationale || `UI operator recorded ${decision}.`,
          reviewer_id: "ui-operator",
          failure_class: decision === "rejected" ? "insufficient_evidence" : ""
        }
      }),
    onSuccess: async () => {
      setProposalRationale("");
      await refreshEverything();
    }
  });

  const proposalRetryMutation = useMutation({
    mutationFn: () =>
      postCommand(`/api/proposals/${viewState.proposal}/commands`, {
        command_kind: "proposal_retry_attempt",
        actor: "ui-operator"
      }),
    onSuccess: refreshEverything
  });

  const proposalStopMutation = useMutation({
    mutationFn: () =>
      postCommand(`/api/proposals/${viewState.proposal}/commands`, {
        command_kind: "proposal_stop_line",
        actor: "ui-operator",
        payload: {
          rationale: proposalRationale || "UI operator stopped the remediation line."
        }
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

  const traceJudgments = traceDetail?.judgments_by_eval_run ?? {};
  const resetAppDataError = resetAppDataMutation.error instanceof Error ? resetAppDataMutation.error.message : "";

  const handleResetAppData = () => {
    if (!window.confirm("Reset all RSI and Honcho app data? Schema versions stay intact, but every conversation, trace, proposal, and memory row will be deleted.")) {
      return;
    }
    resetAppDataMutation.mutate();
  };

  return (
    <div className="app-shell">
      <header className="top-bar">
        <div className="top-brand">
          <Icon name="branch" />
          <span>RSI Router</span>
        </div>
        <label className="global-search">
          <Icon name="search" />
          <input type="search" placeholder="Search" aria-label="Search workspace" />
          <kbd>/</kbd>
        </label>
        <nav className="top-nav" aria-label="Workspace pages">
          <span>Home</span>
          <span>Fusion</span>
          <span>Models</span>
          <span className="active">Chat</span>
          <span>Rankings</span>
          <span>Apps</span>
          <span>Docs</span>
        </nav>
        <div className="profile-menu" aria-label="Profile">
          <span>B</span>
          <strong>Personal</strong>
          <Icon name="chevron" />
        </div>
      </header>

      <aside className="activity-rail" aria-label="Primary navigation">
        <div className="rail-section-title">Workspace</div>
        <nav className="rail-nav" aria-label="Sections">
          <button className={viewState.tab === "conversations" ? "rail-button active" : "rail-button"} onClick={() => navigate({ tab: "conversations" })} aria-label="Conversations">
            <Icon name="chat" />
            <span className="rail-label">Conversations</span>
            <span className="rail-count">{countLabel(conversationsQuery.isFetched, conversations.length)}</span>
          </button>
          <button className={viewState.tab === "cases" ? "rail-button active" : "rail-button"} onClick={() => navigate({ tab: "cases" })} aria-label="Projects">
            <Icon name="folder" />
            <span className="rail-label">Projects</span>
            <span className="rail-count">{countLabel(casesQuery.isFetched, cases.length)}</span>
          </button>
          <button className={viewState.tab === "proposals" ? "rail-button active" : "rail-button"} onClick={() => navigate({ tab: "proposals" })} aria-label="Proposals">
            <Icon name="spark" />
            <span className="rail-label">Proposals</span>
            <span className="rail-count">{countLabel(proposalsQuery.isFetched, proposals.length)}</span>
          </button>
          <button className={viewState.tab === "knowledge" ? "rail-button active" : "rail-button"} onClick={() => navigate({ tab: "knowledge" })} aria-label="Knowledge">
            <Icon name="database" />
            <span className="rail-label">Knowledge</span>
            <span className="rail-count">{countLabel(knowledgeQuery.isFetched, knowledgeEntries.length)}</span>
          </button>
          <button className={viewState.tab === "harness" ? "rail-button active" : "rail-button"} onClick={() => navigate({ tab: "harness" })} aria-label="Harness">
            <Icon name="terminal" />
            <span className="rail-label">Harness</span>
            <span className="rail-count">{countLabel(harnessQuery.isFetched, harnessRoles.length)}</span>
          </button>
        </nav>
        <button className="rail-button plane-trigger" onClick={() => setPlaneOpen(true)} aria-label="Open improvement plane">
          <Icon name="sliders" />
          <span className="rail-label">Improvement Plane</span>
        </button>
      </aside>

      <section className="list-pane">
        {viewState.tab === "conversations" ? (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Conversations</p>
                <h2>Threads and DMs</h2>
              </div>
              <div className="pane-tools">
                <span className="list-range">
                  {conversationsQuery.isFetched ? (
                    sortedConversations.length <= CONVERSATION_PAGE_SIZE
                      ? `${sortedConversations.length} total`
                      : `${conversationPageStart}-${conversationPageEnd} of ${sortedConversations.length}`
                  ) : "Loading"}
                </span>
                {sortedConversations.length > CONVERSATION_PAGE_SIZE ? (
                  <div className="pagination-row compact" aria-label="Conversation pages">
                    <button
                      className="pager-button"
                      aria-label="Previous conversations page"
                      onClick={() => setConversationPage((current) => clampPage(current - 1, sortedConversations.length, CONVERSATION_PAGE_SIZE))}
                      disabled={conversationPage <= 1}
                    >
                      Prev
                    </button>
                    <span>Page {clampedConversationPage} of {conversationPageCount}</span>
                    <button
                      className="pager-button"
                      aria-label="Next conversations page"
                      onClick={() => setConversationPage((current) => clampPage(current + 1, sortedConversations.length, CONVERSATION_PAGE_SIZE))}
                      disabled={conversationPage >= conversationPageCount}
                    >
                      Next
                    </button>
                  </div>
                ) : null}
              </div>
            </header>
            <div className="list-stack">
              {visibleConversations.map((item) => {
                const live = isLiveConversation(item);
                return (
                  <button
                    key={item.conversation_id}
                    className={item.conversation_id === viewState.conversation ? "list-card selected" : "list-card"}
                    onClick={() => navigate({ tab: "conversations", conversation: item.conversation_id, trace: item.active_case?.latest_trace_id })}
                  >
                    <div className="list-card-header">
                      <div className="card-title-block">
                        <span className="list-title-line">
                          <ActivityGlyph active={live} label={live ? "Live trace running" : "No live trace"} />
                          <strong>{item.title || item.external_key}</strong>
                        </span>
                        <p>{item.source} · {item.status}</p>
                      </div>
                      <div className="list-card-badges">
                        <span className="status-chip">{formatTime(item.latest_message_at)}</span>
                        {item.latest_trace_verdict ? <span className="status-chip eval">{item.latest_trace_verdict}</span> : null}
                      </div>
                    </div>
                    <p className="trace-thread">{item.external_key}</p>
                    <dl className="mini-metrics">
                      <div><dt>Active case</dt><dd>{item.active_case?.title || "none"}</dd></div>
                      <div><dt>Status</dt><dd>{item.status}</dd></div>
                      <div><dt>Open traces</dt><dd>{item.open_trace_count}</dd></div>
                      <div><dt>Proposals</dt><dd>{item.proposal_count}</dd></div>
                    </dl>
                  </button>
                );
              })}
              {conversationsQuery.isFetched && visibleConversations.length === 0 ? (
                <div className="empty-list">No conversations yet.</div>
              ) : null}
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
              {cases.map((item) => {
                const live = isLiveCase(item);
                return (
                  <button
                    key={item.case_id}
                    className={item.case_id === viewState.case ? "list-card selected" : "list-card"}
                    onClick={() => navigate({ tab: "cases", case: item.case_id, trace: item.latest_trace_id })}
                  >
                    <div className="list-card-header">
                      <div className="card-title-block">
                        <span className="list-title-line">
                          <ActivityGlyph active={live} label={live ? "Live trace running" : "No live trace"} />
                          <strong>{item.title}</strong>
                        </span>
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
                );
              })}
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
                  const live = isLiveProposal(proposal.status);
                  return (
                    <button
                      key={proposal.id}
                      className={proposal.id === viewState.proposal ? "list-card selected" : "list-card"}
                      onClick={() => navigate({ tab: "proposals", proposal: proposal.id })}
                    >
                      <div className="list-card-header">
                        <div className="card-title-block">
                          <span className="list-title-line">
                            <ActivityGlyph active={live} label={live ? "Proposal line running" : "Proposal line idle"} />
                            <strong>{proposal.title}</strong>
                          </span>
                          <p>{proposal.status} · {proposal.recommended_intervention_kind || "repo_change"} · {proposal.candidate_key}</p>
                        </div>
                        <span className="status-chip">{proposal.status}</span>
                      </div>
                      <p className="trace-thread">{proposal.summary}</p>
                    <dl className="mini-metrics">
                      <div><dt>Risk</dt><dd>{proposal.risk_tier || "n/a"}</dd></div>
                      <div><dt>Target</dt><dd>{proposal.target_surface || proposal.target_layer || "repo_change"}</dd></div>
                      <div><dt>Slot</dt><dd>{proposal.active_slot_consuming ? "occupied" : "free"}</dd></div>
                      <div><dt>Disposition</dt><dd>{proposal.recommended_disposition || "approve_intervention"}</dd></div>
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
            canRetry={canRetryProposal(proposalDetailQuery.data)}
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

      <button className="plane-edge-toggle" onClick={() => setPlaneOpen(true)} aria-label="Open improvement plane" title="Improvement Plane">
        <Icon name="sliders" />
        <span className="visually-hidden">Improvement Plane</span>
      </button>
      <div className={planeOpen ? "plane-backdrop visible" : "plane-backdrop"} onClick={() => setPlaneOpen(false)} />
      <aside className={planeOpen ? "improvement-drawer open" : "improvement-drawer"} aria-hidden={!planeOpen}>
        <div className="drawer-header">
          <div>
            <p className="eyebrow">Improvement Plane</p>
            <h1>Evidence-first operator workspace</h1>
          </div>
          <button className="secondary icon-button" onClick={() => setPlaneOpen(false)} aria-label="Close improvement plane">
            <Icon name="close" />
          </button>
        </div>
        <p className="muted">
          Start from conversations, move into projects, and inspect the exact trace evidence that produced a proposal.
        </p>

        <section className="operations-card">
          <div className="section-header">
            <div>
              <p className="eyebrow">Operations</p>
              <h2>Proposal cap</h2>
            </div>
            <span className="status-chip">{proposalSlotState ? `${proposalSlotState.active}/${proposalSlotState.cap}` : "..."}</span>
          </div>
          <dl className="slot-grid">
            <div><dt>Active</dt><dd>{proposalSlotState ? proposalSlotState.active : "..."}</dd></div>
            <div><dt>Available</dt><dd>{proposalSlotState ? proposalSlotState.available : "..."}</dd></div>
            <div><dt>Stale</dt><dd>{proposalSlotState ? listOrEmpty(proposalSlotState.stale_proposal_ids).length : "..."}</dd></div>
            <div><dt>Candidates</dt><dd>{proposalsQuery.isFetched ? candidates.length : "..."}</dd></div>
          </dl>
          <label className="field">
            Active proposal cap
            <input type="number" min={1} value={proposalCapInput} onChange={(event) => setProposalCapInput(event.target.value)} />
          </label>
          <div className="button-row">
            <button onClick={() => settingsMutation.mutate()} disabled={settingsMutation.isPending || !proposalSlotState}>Save cap</button>
            <button className="secondary" onClick={() => promoteMutation.mutate()} disabled={promoteMutation.isPending || !proposalSlotState || proposalSlotState.available === 0}>
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

        <section className="operations-card danger-zone">
          <div className="section-header">
            <div>
              <p className="eyebrow">Reset</p>
              <h2>App data</h2>
            </div>
          </div>
          <p className="danger-copy">
            Truncate all RSI and Honcho app tables while preserving schema versions so the next e2e run starts from a clean slate.
          </p>
          <div className="button-row">
            <button className="danger" onClick={handleResetAppData} disabled={resetAppDataMutation.isPending}>
              {resetAppDataMutation.isPending ? "Resetting..." : "Reset app data"}
            </button>
          </div>
          {resetAppDataMutation.isSuccess ? (
            <p className="inline-feedback success">
              Reset finished {formatTime(resetAppDataMutation.data?.reset_at)}.
            </p>
          ) : null}
          {resetAppDataError ? (
            <p className="inline-feedback error">{resetAppDataError}</p>
          ) : null}
        </section>
      </aside>
    </div>
  );
}
